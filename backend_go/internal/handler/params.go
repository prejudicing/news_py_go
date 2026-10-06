package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// bind 限制 JSON 请求体大小并执行 DTO 绑定校验；失败时直接写入 422 响应。
func bind(c *gin.Context, value any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
	if err := c.ShouldBindJSON(value); err != nil {
		fail(c, 422, "请求参数不合法")
		return false
	}
	return true
}

// number 解析有边界的整数查询参数，并在缺省时采用 fallback。
func number(c *gin.Context, key string, fallback, min, max int) (int, bool) {
	value := fallback
	if raw, present := c.GetQuery(key); present {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			fail(c, 422, key+"必须为整数")
			return 0, false
		}
		value = parsed
	}
	if value < min || value > max {
		fail(c, 422, key+"超出允许范围")
		return 0, false
	}
	return value, true
}

// pagination 统一解析分页参数，避免各列表接口采用不同默认值和上限。
func pagination(c *gin.Context) (int, int, bool) {
	page, valid := number(c, "page", 1, 1, 1000000)
	if !valid {
		return 0, 0, false
	}
	size, valid := number(c, "pageSize", 10, 1, 100)
	return page, size, valid
}
