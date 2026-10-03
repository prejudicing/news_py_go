// 声明独立业务逻辑包。
package service

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入bytes，用于AI 请求体字节读取器。
	"bytes"
	// 引入context，用于请求取消与超时上下文。
	"context"
	// 引入encoding/json，用于JSON 请求和数据编码。
	"encoding/json"
	// 引入io，用于SSE 读取流和结束标记。
	"io"
	// 引入net/http，用于HTTP 请求、客户端和状态常量。
	"net/http"
	// 引入strings，用于认证头及响应类型检查。
	"strings"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 结束当前声明或代码作用域。
)

// Chat 返回经过检查的上游流，HTTP 层负责转发和关闭读取器。
func (s *Service) Chat(ctx context.Context, input dto.ChatRequest) (io.ReadCloser, error) {
	// AI 密钥未配置时返回服务不可用错误。
	if s.Config.AIKey == "" {
		// 返回校验失败对应的业务状态码和消息。
		return nil, problem(503, "AI服务尚未配置，请联系管理员")
		// 结束当前声明或代码作用域。
	}
	// 把 AI 模型、对话消息和流式标记编码为 JSON。
	body, err := json.Marshal(map[string]any{"model": s.Config.AIModel, "messages": input.Messages, "stream": true})
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回校验失败对应的业务状态码和消息。
		return nil, problem(500, "请求编码失败")
		// 结束当前声明或代码作用域。
	}
	// 使用请求上下文创建 AI 上游请求，支持客户端取消。
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.Config.AIEndpoint, bytes.NewReader(body))
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回校验失败对应的业务状态码和消息。
		return nil, problem(502, "AI服务配置异常")
		// 结束当前声明或代码作用域。
	}
	// 设置 AI 上游所需的内容类型、认证或流式请求头。
	request.Header.Set("Content-Type", "application/json")
	// 设置 AI 上游所需的内容类型、认证或流式请求头。
	request.Header.Set("Authorization", "Bearer "+s.Config.AIKey)
	// 设置 AI 上游所需的内容类型、认证或流式请求头。
	request.Header.Set("X-DashScope-SSE", "enable")
	// 发送 AI 请求并保存上游响应。
	response, err := s.HTTP.Do(request)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回校验失败对应的业务状态码和消息。
		return nil, problem(502, "AI服务连接失败，请稍后重试")
		// 结束当前声明或代码作用域。
	}
	// 确认 AI 上游成功返回 SSE，失败时拒绝转发内容。
	if response.StatusCode != 200 || !strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		// 关闭不能转发的 AI 响应体，释放连接。
		response.Body.Close()
		// 返回校验失败对应的业务状态码和消息。
		return nil, problem(502, "AI服务暂不可用，请检查配置或稍后重试")
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return response.Body, nil
	// 结束当前声明或代码作用域。
}
