package handler

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
)

// ok 使用项目统一的成功响应信封。
func ok(c *gin.Context, message string, data any) {
	c.JSON(200, gin.H{"code": 200, "message": message, "data": data})
}

// fail 使用统一错误结构，并中止后续 Gin handler 执行。
func fail(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"code": status, "message": message, "data": nil})
}

// respond 将业务错误和仓储错误映射为稳定的 HTTP 状态及对外消息。
func (h *Handler) respond(c *gin.Context, message string, data any, err error) {
	if err == nil {
		ok(c, message, data)
		return
	}
	var business *service.Error
	switch {
	case errors.As(err, &business):
		fail(c, business.Status, business.Message)
	case errors.Is(err, repository.ErrNotFound):
		fail(c, 404, "记录不存在")
	case errors.Is(err, repository.ErrDuplicate):
		fail(c, 400, "记录已存在")
	case errors.Is(err, repository.ErrRelation):
		fail(c, 400, "关联数据不存在")
	default:
		log.Printf("operation failed: %T", err)
		fail(c, 500, "数据库操作失败，请稍后重试")
	}
}
