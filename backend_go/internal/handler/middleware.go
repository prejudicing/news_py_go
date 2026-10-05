package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
)

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

func currentUser(c *gin.Context) model.User { return c.MustGet("user").(model.User) }
