// Package handler 负责 HTTP 协议适配，不直接访问数据库或缓存。
package handler

import "github.com/prejudicing/news_py_go/backend_go/internal/service"

type Handler struct {
	user     *service.UserService
	news     *service.NewsService
	favorite *service.FavoriteService
	history  *service.HistoryService
	ai       *service.AIService
	health   *service.HealthService
}

func New(services *service.Services) *Handler {
	return &Handler{
		user: services.User, news: services.News, favorite: services.Favorite,
		history: services.History, ai: services.AI, health: services.Health,
	}
}
