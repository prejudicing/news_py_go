package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
)

// cors 只允许配置的前端来源携带 Cookie，其他来源不能读取凭据响应。
func (h *Handler) cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin == h.frontendOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// auth 从 Authorization 请求头验证短期 Access JWT，并将用户写入请求上下文。
func (h *Handler) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		user, err := h.user.Authenticate(c.Request.Context(), token)
		if err != nil {
			h.respond(c, "", nil, err)
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

// currentUser 读取 auth 中间件已验证并写入上下文的用户。
func currentUser(c *gin.Context) model.User { return c.MustGet("user").(model.User) }
