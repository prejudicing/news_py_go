// Package handler 负责 HTTP 协议适配，不直接访问数据库或缓存。
package handler

import "github.com/prejudicing/news_py_go/backend_go/internal/service"

// Handler 持有各领域服务，并负责 HTTP 中间件和响应配置。
type Handler struct {
	user           *service.UserService
	news           *service.NewsService
	favorite       *service.FavoriteService
	history        *service.HistoryService
	ai             *service.AIService
	health         *service.HealthService
	frontendOrigin string
	cookieSecure   bool
}

// New 将应用服务注入 HTTP 层，并设置跨域来源与 Refresh Cookie 安全属性。
func New(services *service.Services, frontendOrigin string, cookieSecure bool) *Handler {
	if frontendOrigin == "" {
		frontendOrigin = "http://127.0.0.1:5173"
	}
	return &Handler{
		user: services.User, news: services.News, favorite: services.Favorite,
		history: services.History, ai: services.AI, health: services.Health,
		frontendOrigin: frontendOrigin, cookieSecure: cookieSecure,
	}
}
