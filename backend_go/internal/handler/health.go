package handler

import "github.com/gin-gonic/gin"

func (h *Handler) healthCheck(c *gin.Context) {
	data, err := h.health.Health(c.Request.Context())
	h.respond(c, "success", data, err)
}
