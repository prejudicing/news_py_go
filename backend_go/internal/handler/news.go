// 声明HTTP 路由和请求处理包。
package handler

// 引入github.com/gin-gonic/gin，用于Gin 路由和请求上下文。
import "github.com/gin-gonic/gin"

// 校验分类分页参数并调用业务服务。
func (s *Handler) categories(c *gin.Context) {
	// 构造或保存skip, valid 操作结果。
	skip, valid := number(c, "skip", 0, 0, 1000000)
	// 查询参数不符合要求时终止处理。
	if !valid {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 构造或保存limit, valid 操作结果。
	limit, valid := number(c, "limit", 100, 1, 1000)
	// 查询参数不符合要求时终止处理。
	if !valid {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 读取排序后的新闻分类并保存操作结果。
	data, err := s.Service.Categories(c.Request.Context(), skip, limit)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "获取新闻分类成功", data, err)
	// 结束当前声明或代码作用域。
}

// 校验新闻分类与页码并调用业务服务。
func (s *Handler) newsList(c *gin.Context) {
	// 构造或保存category, valid 操作结果。
	category, valid := number(c, "categoryId", -1, 1, 2147483647)
	// 查询参数不符合要求时终止处理。
	if !valid {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 构造或保存page, size, valid 操作结果。
	page, size, valid := pagination(c)
	// 查询参数不符合要求时终止处理。
	if !valid {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 读取指定分类的分页新闻与总数并保存操作结果。
	data, err := s.Service.NewsList(c.Request.Context(), category, page, size)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "获取新闻列表成功", data, err)
	// 结束当前声明或代码作用域。
}

// 校验新闻 ID 并调用详情服务。
func (s *Handler) newsDetail(c *gin.Context) {
	// 构造或保存id, valid 操作结果。
	id, valid := number(c, "id", -1, 1, 2147483647)
	// 查询参数不符合要求时终止处理。
	if !valid {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 读取新闻详情、累计浏览量并查询相关新闻并保存操作结果。
	data, err := s.Service.NewsDetail(c.Request.Context(), uint64(id))
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "success", data, err)
	// 结束当前声明或代码作用域。
}
