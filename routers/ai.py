import os
from typing import Literal

import httpx
from dotenv import load_dotenv
from fastapi import APIRouter, Depends, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field

from utils.auth import get_current_user

load_dotenv()
router = APIRouter(prefix="/api/ai", tags=["ai"])


class ChatMessage(BaseModel):
    role: Literal["user", "assistant"]
    content: str = Field(min_length=1, max_length=20000)


class ChatRequest(BaseModel):
    messages: list[ChatMessage] = Field(min_length=1, max_length=50)


@router.post("/chat")
async def chat(data: ChatRequest, user=Depends(get_current_user)):
    api_key = os.getenv("AI_API_KEY")
    if not api_key:
        raise HTTPException(status_code=503, detail="AI服务尚未配置，请联系管理员")

    client = httpx.AsyncClient(timeout=httpx.Timeout(60.0, connect=10.0))
    request = client.build_request(
        "POST",
        os.getenv("AI_API_ENDPOINT", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"),
        headers={"Authorization": f"Bearer {api_key}", "X-DashScope-SSE": "enable"},
        json={
            "model": os.getenv("AI_MODEL", "qwen3-max-preview"),
            "messages": [message.model_dump() for message in data.messages],
            "stream": True,
        },
    )
    try:
        response = await client.send(request, stream=True)
    except httpx.RequestError:
        await client.aclose()
        raise HTTPException(status_code=502, detail="AI服务连接失败，请稍后重试") from None

    if response.status_code != 200 or "text/event-stream" not in response.headers.get("content-type", ""):
        await response.aclose()
        await client.aclose()
        raise HTTPException(status_code=502, detail="AI服务暂不可用，请检查后端配置或稍后重试")

    async def stream():
        try:
            async for chunk in response.aiter_bytes():
                yield chunk
        except httpx.RequestError:
            yield b'event: error\ndata: {"error":{"message":"AI stream interrupted"}}\n\n'
        finally:
            await response.aclose()
            await client.aclose()

    return StreamingResponse(stream(), media_type="text/event-stream", headers={"Cache-Control": "no-cache"})
