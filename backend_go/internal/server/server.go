package server

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	"github.com/prejudicing/news_py_go/backend_go/internal/handler"
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
)

// Server 是应用依赖的装配入口。
type Server struct {
	Services       *service.Services
	FrontendOrigin string
	CookieSecure   bool
}

// 创建实例并注入所需依赖。
func New(db *gorm.DB, cached *cache.Store, cfg config.Config) *Server {
	repo := repository.New(db)
	return &Server{Services: service.NewServices(repo, cached, cfg), FrontendOrigin: cfg.FrontendOrigin, CookieSecure: cfg.CookieSecure}
}

// 注册 HTTP 路由与通用中间件。
func (s *Server) Router() *gin.Engine {
	return handler.New(s.Services, s.FrontendOrigin, s.CookieSecure).Router()
}
