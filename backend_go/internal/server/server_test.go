package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
)

// 测试不依赖数据库的参数验证、认证和错误格式。
func TestValidationAndAuthWithoutDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New(nil, nil, config.Config{}).Router()
	cases := []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/", "", 200},
		{"OPTIONS", "/api/news/list", "", 204},
		{"GET", "/api/news/list", "", 422},
		{"GET", "/api/news/list?categoryId=1&pageSize=0", "", 422},
		{"GET", "/api/news/list?categoryId=1&page=-1", "", 422},
		{"GET", "/api/news/detail?id=invalid", "", 422},
		{"GET", "/api/news/categories?skip=-1", "", 422},
		{"GET", "/api/favorite/list", "", 401},
		{"POST", "/api/ai/chat", `{"messages":[]}`, 401},
		{"POST", "/api/user/register", `{"username":"test"}`, 422},
		{"POST", "/api/user/login", "invalid", 422},
		{"GET", "/missing", "", 404},
		{"POST", "/api/news/list", "", 405},
	}
	for _, tc := range cases {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", w.Code, tc.status, w.Body.String())
			}
			if w.Header().Get("Access-Control-Allow-Origin") != "*" {
				t.Fatal("CORS header missing")
			}
			if tc.status >= 400 {
				var body struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
					Data    any    `json:"data"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Code != tc.status || body.Message == "" || body.Data != nil {
					t.Fatal("invalid error envelope")
				}
			}
		})
	}
}
