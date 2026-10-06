package handler

import "github.com/gin-gonic/gin"

// healthCheck 汇总数据库和缓存依赖状态，供部署探针检查服务可用性。
func (h *Handler) healthCheck(c *gin.Context) {
	data, err := h.health.Health(c.Request.Context())
	h.respond(c, "success", data, err)
}
