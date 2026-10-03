// 声明HTTP 路由和请求处理包。
package handler

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入errors，用于错误类型匹配和构造。
	"errors"
	// 引入github.com/gin-gonic/gin，用于Gin 路由和请求上下文。
	"github.com/gin-gonic/gin"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/repository，用于仓储接口和统一错误。
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/service，用于业务逻辑。
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
	// 引入log，用于不含敏感数据的运行错误日志。
	"log"
	// 引入net/http，用于HTTP 请求、客户端和状态常量。
	"net/http"
	// 引入strconv，用于查询参数和主键转换。
	"strconv"
	// 引入strings，用于认证头及响应类型检查。
	"strings"
	// 结束当前声明或代码作用域。
)

// Handler 仅处理 HTTP 协议，并把业务委托给服务层。
type Handler struct{ Service *service.Service }

// 创建实例并注入所需依赖。
func New(s *service.Service) *Handler { return &Handler{Service: s} }

// 注册 HTTP 路由与通用中间件。
func (s *Handler) Router() *gin.Engine {
	// 创建不带默认中间件的 Gin 路由。
	r := gin.New()
	// 注册访问日志、异常恢复和跨域处理，避免请求异常终止服务。
	r.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, recovered any) { fail(c, 500, "服务器内部错误") }), cors())
	// 关闭代理信任，避免通过伪造代理头影响客户端地址识别。
	_ = r.SetTrustedProxies(nil)
	// 注册GET接口，并连接对应的处理函数。
	r.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Hello World"}) })
	// 注册GET接口，并连接对应的处理函数。
	r.GET("/health", s.health)
	// 为不存在的路径返回统一 404 响应。
	r.NoRoute(func(c *gin.Context) { fail(c, 404, "接口不存在") })
	// 为方法不匹配的请求返回统一 405 响应。
	r.NoMethod(func(c *gin.Context) { fail(c, 405, "请求方法不支持") })
	// 路径存在但请求方法不匹配时返回 405。
	r.HandleMethodNotAllowed = true
	// 创建新闻路由组，统一使用 /api/news 前缀。
	news := r.Group("/api/news")
	// 为当前路由组注册GET接口；需要认证的接口先执行认证中间件。
	news.GET("/categories", s.categories)
	// 为当前路由组注册GET接口；需要认证的接口先执行认证中间件。
	news.GET("/list", s.newsList)
	// 为当前路由组注册GET接口；需要认证的接口先执行认证中间件。
	news.GET("/detail", s.newsDetail)
	// 创建用户路由组，统一使用 /api/user 前缀。
	users := r.Group("/api/user")
	// 为当前路由组注册POST接口；需要认证的接口先执行认证中间件。
	users.POST("/register", s.register)
	// 为当前路由组注册POST接口；需要认证的接口先执行认证中间件。
	users.POST("/login", s.login)
	// 为当前路由组注册GET接口；需要认证的接口先执行认证中间件。
	users.GET("/info", s.auth(), s.userInfo)
	// 为当前路由组注册PUT接口；需要认证的接口先执行认证中间件。
	users.PUT("/update", s.auth(), s.updateUser)
	// 为当前路由组注册PUT接口；需要认证的接口先执行认证中间件。
	users.PUT("/password", s.auth(), s.changePassword)
	// 创建收藏路由组，为所有收藏接口启用令牌认证。
	favorites := r.Group("/api/favorite", s.auth())
	// 为当前路由组注册GET接口；需要认证的接口先执行认证中间件。
	favorites.GET("/check", s.checkFavorite)
	// 为当前路由组注册POST接口；需要认证的接口先执行认证中间件。
	favorites.POST("/add", s.addFavorite)
	// 为当前路由组注册DELETE接口；需要认证的接口先执行认证中间件。
	favorites.DELETE("/remove", s.removeFavorite)
	// 为当前路由组注册GET接口；需要认证的接口先执行认证中间件。
	favorites.GET("/list", s.favoriteList)
	// 为当前路由组注册DELETE接口；需要认证的接口先执行认证中间件。
	favorites.DELETE("/clear", s.clearFavorites)
	// 创建浏览历史路由组，为所有历史接口启用令牌认证。
	history := r.Group("/api/history", s.auth())
	// 为当前路由组注册POST接口；需要认证的接口先执行认证中间件。
	history.POST("/add", s.addHistory)
	// 为当前路由组注册GET接口；需要认证的接口先执行认证中间件。
	history.GET("/list", s.historyList)
	// 为当前路由组注册DELETE接口；需要认证的接口先执行认证中间件。
	history.DELETE("/delete/:history_id", s.deleteHistory)
	// 为当前路由组注册DELETE接口；需要认证的接口先执行认证中间件。
	history.DELETE("/clear", s.clearHistory)
	// 注册POST接口，并连接对应的处理函数。
	r.POST("/api/ai/chat", s.auth(), s.chat)
	// 返回已完成接口注册的 Gin 路由。
	return r
	// 结束当前作用域的代码块或结构体定义。
}

// 执行 cors 对应操作。
func cors() gin.HandlerFunc {
	// 返回对每个请求执行的 Gin 中间件函数。
	return func(c *gin.Context) {
		// 允许浏览器跨域访问本项目的 API。
		c.Header("Access-Control-Allow-Origin", "*")
		// 允许跨域请求携带令牌和内容类型头。
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		// 声明允许的跨域请求方法。
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// 浏览器发送 OPTIONS 预检时提前返回。
		if c.Request.Method == http.MethodOptions {
			// 用无响应体的成功状态结束跨域预检请求。
			c.AbortWithStatus(204)
			// 结束当前处理，错误或响应已在前面的语句中处理。
			return
			// 结束当前作用域的代码块或结构体定义。
		}
		// 继续执行后续中间件和业务处理器。
		c.Next()
		// 结束当前作用域的代码块或结构体定义。
	}
	// 结束当前作用域的代码块或结构体定义。
}

// 执行 ok 对应操作。
func ok(c *gin.Context, message string, data any) {
	// 以 JSON 输出成功状态、提示信息和业务数据。
	c.JSON(200, gin.H{"code": 200, "message": message, "data": data})
	// 结束当前作用域的代码块或结构体定义。
}

// 执行 fail 对应操作。
func fail(c *gin.Context, status int, message string) {
	// 以 JSON 输出错误状态和提示，并中止当前请求处理。
	c.AbortWithStatusJSON(status, gin.H{"code": status, "message": message, "data": nil})
	// 结束当前作用域的代码块或结构体定义。
}

// 执行 bind 对应操作。
func bind(c *gin.Context, value any) bool {
	// 把 JSON 请求体限制为一兆字节，避免超大请求。
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
	// 当前操作失败时进入错误处理分支。
	if err := c.ShouldBindJSON(value); err != nil {
		// 返回 422 错误响应，并阻止后续处理。
		fail(c, 422, "请求参数不合法")
		// 返回失败或未命中的标记。
		return false
		// 结束当前作用域的代码块或结构体定义。
	}
	// 返回成功的标记。
	return true
	// 结束当前作用域的代码块或结构体定义。
}

// 执行 number 对应操作。
func number(c *gin.Context, key string, fallback, min, max int) (int, bool) {
	// 先使用默认值，存在查询参数时再覆盖。
	value := fallback
	// 查询字符串包含该字段时覆盖默认值。
	if raw, present := c.GetQuery(key); present {
		// 声明变量，用于保存当前操作的错误。
		var err error
		// 将查询参数文本转换为整数。
		value, err = strconv.Atoi(raw)
		// 当前操作失败时进入错误处理分支。
		if err != nil {
			// 返回 422 错误响应，并阻止后续处理。
			fail(c, 422, key+"必须为整数")
			// 参数解析失败，返回零值和失败标记。
			return 0, false
			// 结束当前作用域的代码块或结构体定义。
		}
		// 结束当前作用域的代码块或结构体定义。
	}
	// 整数参数超出允许范围时返回参数错误。
	if value < min || value > max {
		// 返回 422 错误响应，并阻止后续处理。
		fail(c, 422, key+"超出允许范围")
		// 参数解析失败，返回零值和失败标记。
		return 0, false
		// 结束当前作用域的代码块或结构体定义。
	}
	// 返回已验证的整数参数及成功标记。
	return value, true
	// 结束当前作用域的代码块或结构体定义。
}

// 执行 pagination 对应操作。
func pagination(c *gin.Context) (int, int, bool) {
	// 页码默认第一页，必须为允许范围内的正整数。
	page, valid := number(c, "page", 1, 1, 1000000)
	// 查询参数校验未通过时立即返回。
	if !valid {
		// 分页参数校验失败，返回两个零值和失败标记。
		return 0, 0, false
		// 结束当前作用域的代码块或结构体定义。
	}
	// 每页默认十条，允许范围为一到一百条。
	size, valid := number(c, "pageSize", 10, 1, 100)
	// 返回页码、每页数量以及校验结果。
	return page, size, valid
	// 结束当前作用域的代码块或结构体定义。
}

// 执行 currentUser 对应操作。
func currentUser(c *gin.Context) model.User { return c.MustGet("user").(model.User) }

// 统一输出成功数据或转换后的错误响应。
func (s *Handler) respond(c *gin.Context, message string, data any, err error) {
	// 操作成功时使用统一的成功响应。
	if err == nil {
		// 输出成功状态码、提示消息和业务数据。
		ok(c, message, data)
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 声明业务错误对象。
	var business *service.Error
	// 按错误类型选择公开的响应或仓储错误。
	switch {
	// 匹配当前错误类型并进入对应处理分支。
	case errors.As(err, &business):
		// 输出错误状态码和公开提示，并中止当前请求。
		fail(c, business.Status, business.Message)
	// 匹配当前错误类型并进入对应处理分支。
	case errors.Is(err, repository.ErrNotFound):
		// 输出错误状态码和公开提示，并中止当前请求。
		fail(c, 404, "记录不存在")
	// 匹配当前错误类型并进入对应处理分支。
	case errors.Is(err, repository.ErrDuplicate):
		// 输出错误状态码和公开提示，并中止当前请求。
		fail(c, 400, "记录已存在")
	// 匹配当前错误类型并进入对应处理分支。
	case errors.Is(err, repository.ErrRelation):
		// 输出错误状态码和公开提示，并中止当前请求。
		fail(c, 400, "关联数据不存在")
	// 其他错误使用通用提示，避免泄露底层细节。
	default:
		// 只记录错误类型，避免日志输出密码和密钥。
		log.Printf("operation failed: %T", err)
		// 输出错误状态码和公开提示，并中止当前请求。
		fail(c, 500, "数据库操作失败，请稍后重试")
		// 结束当前声明或代码作用域。
	}
	// 结束当前声明或代码作用域。
}

// 提取认证头并通过业务服务验证令牌。
func (s *Handler) auth() gin.HandlerFunc {
	// 返回当前计算或构造的结果。
	return func(c *gin.Context) {
		// 去掉认证头首尾空白并取得实际令牌。
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		// 识别 Bearer 前缀并允许原有裸令牌格式。
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			// 去掉认证头首尾空白并取得实际令牌。
			token = strings.TrimSpace(token[7:])
			// 结束当前声明或代码作用域。
		}
		// 校验令牌及有效期并返回所属用户并保存操作结果。
		user, err := s.Service.Authenticate(c.Request.Context(), token)
		// 检查本次操作是否失败，必要时提前返回错误。
		if err != nil {
			// 统一输出成功数据或转换后的错误响应并保存操作结果。
			s.respond(c, "", nil, err)
			// 结束当前处理，避免继续执行后续操作。
			return
			// 结束当前声明或代码作用域。
		}
		// 把认证用户存入请求上下文供后续接口使用。
		c.Set("user", user)
		// 继续执行后续中间件和业务处理器。
		c.Next()
		// 结束当前声明或代码作用域。
	}
	// 结束当前声明或代码作用域。
}

// 调用健康检查业务并输出结果。
func (s *Handler) health(c *gin.Context) {
	// 检查 MySQL 与 Redis 的连接状态并保存操作结果。
	data, err := s.Service.Health(c.Request.Context())
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "success", data, err)
	// 结束当前声明或代码作用域。
}
