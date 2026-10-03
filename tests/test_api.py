import os
import unittest
from datetime import datetime
from unittest.mock import patch, AsyncMock
import asyncio
from types import SimpleNamespace

import httpx
from fastapi.testclient import TestClient

from main import app
from models.news import News
from schemas.favorite import FavoriteNewsItemResponse
from schemas.history import HistoryNewsItemResponse
from utils.auth import get_current_user


class ApiTests(unittest.TestCase):
    def setUp(self):
        self.client = TestClient(app)

    def tearDown(self):
        app.dependency_overrides.clear()
        self.client.close()

    def test_news_list_serializes_frontend_fields(self):
        item = News(id=1, title="test", category_id=2, views=3, publish_time=datetime(2026, 10, 3))
        with patch("routers.news.news_cache.get_news_list", return_value=[item]), patch(
            "routers.news.news.get_news_count", return_value=1
        ):
            response = self.client.get("/api/news/list?categoryId=2")
        self.assertEqual(response.status_code, 200)
        payload = response.json()["data"]["list"][0]
        self.assertEqual(payload["categoryId"], 2)
        self.assertEqual(payload["publishTime"], "2026-10-03T00:00:00")
        self.assertNotIn("publish_time", payload)

    def test_history_delete_filters_record_id_and_current_user(self):
        from crud.history import delete_history
        db = SimpleNamespace(execute=AsyncMock(return_value=SimpleNamespace(rowcount=1)), commit=AsyncMock())
        self.assertTrue(asyncio.run(delete_history(db, 7, 3)))
        query = db.execute.call_args.args[0].compile()
        self.assertIn("history.id =", str(query))
        self.assertIn("history.user_id =", str(query))
        self.assertNotIn("history.news_id =", str(query))
        self.assertEqual(set(query.params.values()), {7, 3})

    def test_favorite_and_history_use_same_time_field(self):
        values = dict(id=1, title="test", category_id=2, views=3, publish_time=datetime(2026, 10, 3))
        favorite = FavoriteNewsItemResponse(**values, favorite_id=1, favorite_time=datetime(2026, 10, 3))
        history = HistoryNewsItemResponse(**values, history_id=1, view_time=datetime(2026, 10, 3))
        for item in (favorite, history):
            self.assertIn("publishTime", item.model_dump(by_alias=True))

    def test_ai_requires_authentication(self):
        with patch("routers.ai.httpx.AsyncClient") as upstream:
            response = self.client.post("/api/ai/chat", json={"messages": [{"role": "user", "content": "hello"}]})
        self.assertIn(response.status_code, (401, 422))
        upstream.assert_not_called()

    def test_ai_missing_config(self):
        app.dependency_overrides[get_current_user] = lambda: object()
        with patch.dict(os.environ, {"AI_API_KEY": ""}):
            response = self.client.post("/api/ai/chat", json={"messages": [{"role": "user", "content": "hello"}]})
        self.assertEqual(response.status_code, 503)

    def test_ai_forwards_sse_with_server_credentials(self):
        app.dependency_overrides[get_current_user] = lambda: object()
        frames = b'data: {"choices":[{"delta":{"content":"hello"}}]}\n\ndata: [DONE]\n\n'

        def upstream(request):
            self.assertEqual(request.headers["Authorization"], "Bearer server-only-test-key")
            self.assertIn(b'"stream":true', request.content)
            return httpx.Response(200, headers={"content-type": "text/event-stream"}, content=frames)

        upstream_client = httpx.AsyncClient(transport=httpx.MockTransport(upstream))
        with patch.dict(os.environ, {"AI_API_KEY": "server-only-test-key"}), patch(
            "routers.ai.httpx.AsyncClient", return_value=upstream_client
        ):
            response = self.client.post("/api/ai/chat", json={"messages": [{"role": "user", "content": "hello"}]})
        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.content, frames)
        self.assertNotIn("server-only-test-key", response.text)
        self.assertTrue(upstream_client.is_closed)

    def test_ai_upstream_error_is_sanitized(self):
        app.dependency_overrides[get_current_user] = lambda: object()
        upstream_client = httpx.AsyncClient(transport=httpx.MockTransport(
            lambda request: httpx.Response(401, json={"error": "server-only-test-key"})
        ))
        with patch.dict(os.environ, {"AI_API_KEY": "server-only-test-key"}), patch(
            "routers.ai.httpx.AsyncClient", return_value=upstream_client
        ):
            response = self.client.post("/api/ai/chat", json={"messages": [{"role": "user", "content": "hello"}]})
        self.assertEqual(response.status_code, 502)
        self.assertNotIn("server-only-test-key", response.text)
        self.assertTrue(upstream_client.is_closed)


if __name__ == "__main__":
    unittest.main()
