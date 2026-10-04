package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
)

// Chat 返回经过检查的上游流，HTTP 层负责转发和关闭读取器。
func (s *AIService) Chat(ctx context.Context, input dto.ChatRequest) (io.ReadCloser, error) {
	if s.Config.AIKey == "" {
		return nil, problem(503, "AI服务尚未配置，请联系管理员")
	}
	body, err := json.Marshal(map[string]any{"model": s.Config.AIModel, "messages": input.Messages, "stream": true})
	if err != nil {
		return nil, problem(500, "请求编码失败")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.Config.AIEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, problem(502, "AI服务配置异常")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+s.Config.AIKey)
	request.Header.Set("X-DashScope-SSE", "enable")
	response, err := s.HTTP.Do(request)
	if err != nil {
		return nil, problem(502, "AI服务连接失败，请稍后重试")
	}
	if response.StatusCode != 200 || !strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		response.Body.Close()
		return nil, problem(502, "AI服务暂不可用，请检查配置或稍后重试")
	}
	return response.Body, nil
}
