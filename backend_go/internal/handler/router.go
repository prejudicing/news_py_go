package handler

import "github.com/gin-gonic/gin"

// Router 注册中间件和所有 HTTP 路由。
func (h *Handler) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, _ any) { fail(c, 500, "服务器内部错误") }), h.cors())
	_ = r.SetTrustedProxies(nil)
	r.HandleMethodNotAllowed = true
	r.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "Hello World"}) })
	r.GET("/health", h.healthCheck)
	r.NoRoute(func(c *gin.Context) { fail(c, 404, "接口不存在") })
	r.NoMethod(func(c *gin.Context) { fail(c, 405, "请求方法不支持") })

	news := r.Group("/api/news")
	news.GET("/categories", h.categories)
	news.GET("/list", h.newsList)
	news.GET("/detail", h.newsDetail)

	users := r.Group("/api/user")
	users.POST("/register", h.register)
	users.POST("/login", h.login)
	users.POST("/refresh", h.refresh)
	users.POST("/logout", h.logout)
	users.GET("/info", h.auth(), h.userInfo)
	users.PUT("/update", h.auth(), h.updateUser)
	users.PUT("/password", h.auth(), h.changePassword)

	favorites := r.Group("/api/favorite", h.auth())
	favorites.GET("/check", h.checkFavorite)
	favorites.POST("/add", h.addFavorite)
	favorites.DELETE("/remove", h.removeFavorite)
	favorites.GET("/list", h.favoriteList)
	favorites.DELETE("/clear", h.clearFavorites)

	history := r.Group("/api/history", h.auth())
	history.POST("/add", h.addHistory)
	history.GET("/list", h.historyList)
	history.DELETE("/delete/:history_id", h.deleteHistory)
	history.DELETE("/clear", h.clearHistory)

	r.POST("/api/ai/chat", h.auth(), h.chat)
	return r
}
