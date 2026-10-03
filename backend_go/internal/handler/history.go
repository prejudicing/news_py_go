// 声明HTTP 路由和请求处理包。
package handler

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入strconv，用于查询参数和主键转换。
	"strconv"

	// 引入github.com/gin-gonic/gin，用于Gin 路由和请求上下文。
	"github.com/gin-gonic/gin"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/service，用于业务逻辑。
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
	// 结束当前声明或代码作用域。
)

// 绑定新闻 ID 并调用历史写入服务。
func (s *Handler) addHistory(c *gin.Context) {
	// 声明请求参数。
	var input dto.NewsRequest
	// 绑定和校验请求参数，失败时终止处理。
	if !bind(c, &input) {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 在用户行锁保护下创建或更新浏览记录并保存操作结果。
	item, err := s.Service.AddHistory(c.Request.Context(), currentUser(c).ID, input.NewsID)
	// 格式化不带时区的时间字符串，兼容原前端并保存操作结果。
	s.respond(c, "添加成功", gin.H{"id": item.ID, "user_id": item.UserID, "news_id": item.NewsID, "view_time": service.Timestamp(item.ViewTime)}, err)
	// 结束当前声明或代码作用域。
}

// 校验页码并调用历史列表服务。
func (s *Handler) historyList(c *gin.Context) {
	// 构造或保存page, size, valid 操作结果。
	page, size, valid := pagination(c)
	// 查询参数不符合要求时终止处理。
	if !valid {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 查询当前用户的历史分页列表并保存操作结果。
	data, err := s.Service.Histories(c.Request.Context(), currentUser(c).ID, page, size)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "success", data, err)
	// 结束当前声明或代码作用域。
}

// 校验记录主键并调用历史删除服务。
func (s *Handler) deleteHistory(c *gin.Context) {
	// 把路由中的历史记录主键转换为无符号整数。
	id, err := strconv.ParseUint(c.Param("history_id"), 10, 64)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil || id == 0 {
		// 输出错误状态码和公开提示，并中止当前请求。
		fail(c, 422, "history_id必须为正整数")
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 按记录主键删除当前用户的历史并保存操作结果。
	err = s.Service.DeleteHistory(c.Request.Context(), currentUser(c).ID, id)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "删除成功", nil, err)
	// 结束当前声明或代码作用域。
}

// 调用历史清空服务。
func (s *Handler) clearHistory(c *gin.Context) {
	// 清空当前用户的浏览历史并保存操作结果。
	err := s.Service.ClearHistory(c.Request.Context(), currentUser(c).ID)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "清空成功", nil, err)
	// 结束当前声明或代码作用域。
}
