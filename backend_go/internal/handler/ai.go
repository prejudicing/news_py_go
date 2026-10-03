// 声明HTTP 路由和请求处理包。
package handler

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入io，用于SSE 读取流和结束标记。
	"io"

	// 引入github.com/gin-gonic/gin，用于Gin 路由和请求上下文。
	"github.com/gin-gonic/gin"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 结束当前声明或代码作用域。
)

// 绑定对话参数并把上游 SSE 流转发给浏览器。
func (s *Handler) chat(c *gin.Context) {
	// 声明请求参数。
	var input dto.ChatRequest
	// 绑定和校验请求参数，失败时终止处理。
	if !bind(c, &input) {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 建立经过校验的 AI 上游流式请求并保存操作结果。
	stream, err := s.Service.Chat(c.Request.Context(), input)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 统一输出成功数据或转换后的错误响应并保存操作结果。
		s.respond(c, "", nil, err)
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 保存defer 字段。
	defer stream.Close()
	// 设置 SSE 响应头，关闭缓存或代理缓冲。
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	// 设置 SSE 响应头，关闭缓存或代理缓冲。
	c.Header("Cache-Control", "no-cache")
	// 设置 SSE 响应头，关闭缓存或代理缓冲。
	c.Header("X-Accel-Buffering", "no")
	// 设置流式响应的成功状态。
	c.Status(200)
	// 立即刷新响应缓冲，让浏览器接收当前数据。
	c.Writer.Flush()
	// 分配 4096 字节的 SSE 读取缓冲。
	buffer := make([]byte, 4096)
	// 持续读取上游数据，直到流结束或请求取消。
	for {
		// 从 AI 上游流读取一块数据并保存读取状态。
		n, err := stream.Read(buffer)
		// 上游读取到数据时立即转发这段内容。
		if n > 0 {
			// 检查本次操作是否失败，必要时提前返回错误。
			if _, writeErr := c.Writer.Write(buffer[:n]); writeErr != nil {
				// 结束当前处理，避免继续执行后续操作。
				return
				// 结束当前声明或代码作用域。
			}
			// 立即刷新响应缓冲，让浏览器接收当前数据。
			c.Writer.Flush()
			// 结束当前声明或代码作用域。
		}
		// 检查本次操作是否失败，必要时提前返回错误。
		if err != nil {
			// 检查本次操作是否失败，必要时提前返回错误。
			if err != io.EOF && c.Request.Context().Err() == nil {
				// 上游流异常结束时发送不含内部信息的 SSE 错误。
				_, _ = c.Writer.WriteString("\nevent: error\ndata: {\"error\":{\"message\":\"AI stream interrupted\"}}\n\n")
				// 立即刷新响应缓冲，让浏览器接收当前数据。
				c.Writer.Flush()
				// 结束当前声明或代码作用域。
			}
			// 结束当前处理，避免继续执行后续操作。
			return
			// 结束当前声明或代码作用域。
		}
		// 结束当前声明或代码作用域。
	}
	// 结束当前声明或代码作用域。
}
