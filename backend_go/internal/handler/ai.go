package handler

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
)

// 绑定对话参数并把上游 SSE 流转发给浏览器。
func (s *Handler) chat(c *gin.Context) {
	var input dto.ChatRequest
	if !bind(c, &input) {
		return
	}
	stream, err := s.ai.Chat(c.Request.Context(), input)
	if err != nil {
		s.respond(c, "", nil, err)
		return
	}
	defer stream.Close()
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")
	c.Status(200)
	c.Writer.Flush()
	buffer := make([]byte, 4096)
	for {
		n, err := stream.Read(buffer)
		if n > 0 {
			if _, writeErr := c.Writer.Write(buffer[:n]); writeErr != nil {
				return
			}
			c.Writer.Flush()
		}
		if err != nil {
			if err != io.EOF && c.Request.Context().Err() == nil {
				_, _ = c.Writer.WriteString("\nevent: error\ndata: {\"error\":{\"message\":\"AI stream interrupted\"}}\n\n")
				c.Writer.Flush()
			}
			return
		}
	}
}
