// 声明HTTP 路由和请求处理包。
package handler

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入fmt，用于缓存键和提示消息格式化。
	"fmt"

	// 引入github.com/gin-gonic/gin，用于Gin 路由和请求上下文。
	"github.com/gin-gonic/gin"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/service，用于业务逻辑。
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
	// 结束当前声明或代码作用域。
)

// 校验新闻 ID 并返回收藏状态。
func (s *Handler) checkFavorite(c *gin.Context) {
	// 构造或保存id, valid 操作结果。
	id, valid := number(c, "newsId", -1, 1, 2147483647)
	// 查询参数不符合要求时终止处理。
	if !valid {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 检查当前用户是否收藏指定新闻并保存操作结果。
	found, err := s.Service.CheckFavorite(c.Request.Context(), currentUser(c).ID, uint64(id))
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "检查收藏状态成功", gin.H{"isFavorite": found}, err)
	// 结束当前声明或代码作用域。
}

// 绑定新闻 ID 并调用收藏服务。
func (s *Handler) addFavorite(c *gin.Context) {
	// 声明请求参数。
	var input dto.NewsRequest
	// 绑定和校验请求参数，失败时终止处理。
	if !bind(c, &input) {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 检查新闻存在后添加当前用户的收藏并保存操作结果。
	item, err := s.Service.AddFavorite(c.Request.Context(), currentUser(c).ID, input.NewsID)
	// 格式化不带时区的时间字符串，兼容原前端并保存操作结果。
	s.respond(c, "添加收藏成功", gin.H{"id": item.ID, "user_id": item.UserID, "news_id": item.NewsID, "created_at": service.Timestamp(item.CreatedAt)}, err)
	// 结束当前声明或代码作用域。
}

// 校验新闻 ID 并调用删除服务。
func (s *Handler) removeFavorite(c *gin.Context) {
	// 构造或保存id, valid 操作结果。
	id, valid := number(c, "newsId", -1, 1, 2147483647)
	// 查询参数不符合要求时终止处理。
	if !valid {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 删除当前用户的指定收藏并检查删除结果并保存操作结果。
	err := s.Service.RemoveFavorite(c.Request.Context(), currentUser(c).ID, uint64(id))
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "删除收藏成功", nil, err)
	// 结束当前声明或代码作用域。
}

// 校验页码并调用收藏列表服务。
func (s *Handler) favoriteList(c *gin.Context) {
	// 构造或保存page, size, valid 操作结果。
	page, size, valid := pagination(c)
	// 查询参数不符合要求时终止处理。
	if !valid {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 查询当前用户的收藏分页列表并保存操作结果。
	data, err := s.Service.Favorites(c.Request.Context(), currentUser(c).ID, page, size)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "获取收藏列表成功", data, err)
	// 结束当前声明或代码作用域。
}

// 调用收藏清空服务并返回删除数量。
func (s *Handler) clearFavorites(c *gin.Context) {
	// 清空当前用户的收藏并返回数量并保存操作结果。
	count, err := s.Service.ClearFavorites(c.Request.Context(), currentUser(c).ID)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, fmt.Sprintf("清空了%d条记录", count), nil, err)
	// 结束当前声明或代码作用域。
}
