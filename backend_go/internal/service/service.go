// Package service 实现与 HTTP、GORM 无关的业务规则。
package service

import (
	"context"
	"net/http"
	"time"

	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	"github.com/prejudicing/news_py_go/backend_go/internal/utils"
)

// Services 是依赖装配结果，不承载跨领域业务方法。
type Services struct {
	User     *UserService
	News     *NewsService
	Favorite *FavoriteService
	History  *HistoryService
	AI       *AIService
	Health   *HealthService
}

// UserService 处理注册、登录、认证和资料维护。
type UserService struct {
	Repo      repository.UserRepository
	JWTSecret []byte
	clock
}

// NewsService 处理新闻查询、缓存和浏览量更新。
type NewsService struct {
	Repo  repository.NewsRepository
	Cache *cache.Store
}

// FavoriteService 处理用户收藏。
type FavoriteService struct {
	Repo repository.FavoriteRepository
	clock
}

// HistoryService 处理用户浏览历史。
type HistoryService struct {
	Repo repository.HistoryRepository
	clock
}

// AIService 代理兼容 OpenAI 格式的流式问答。
type AIService struct {
	Config config.Config
	HTTP   *http.Client
}

// HealthService 检查数据库和 Redis 的可用性。
type HealthService struct {
	Repo  repository.HealthRepository
	Cache *cache.Store
}

type clock struct{ location *time.Location }

func (c clock) now() time.Time {
	if c.location != nil {
		return time.Now().In(c.location)
	}
	return time.Now()
}

// NewServices 构造按领域拆分的服务集合。
func NewServices(repo *repository.GORMStore, cached *cache.Store, cfg config.Config) *Services {
	return &Services{
		User:     NewUserService(repo, cfg.Location, cfg.JWTSecret),
		News:     NewNewsService(repo, cached),
		Favorite: NewFavoriteService(repo, cfg.Location),
		History:  NewHistoryService(repo, cfg.Location),
		AI:       NewAIService(cfg),
		Health:   NewHealthService(repo, cached),
	}
}

// NewUserService 注入用户仓储和业务时区。
func NewUserService(repo repository.UserRepository, location *time.Location, secret string) *UserService {
	return &UserService{Repo: repo, JWTSecret: []byte(secret), clock: clock{location: location}}
}

// NewNewsService 注入新闻仓储和缓存。
func NewNewsService(repo repository.NewsRepository, cached *cache.Store) *NewsService {
	return &NewsService{Repo: repo, Cache: cached}
}

// NewFavoriteService 注入收藏仓储和业务时区。
func NewFavoriteService(repo repository.FavoriteRepository, location *time.Location) *FavoriteService {
	return &FavoriteService{Repo: repo, clock: clock{location: location}}
}

// NewHistoryService 注入历史仓储和业务时区。
func NewHistoryService(repo repository.HistoryRepository, location *time.Location) *HistoryService {
	return &HistoryService{Repo: repo, clock: clock{location: location}}
}

// NewAIService 创建带超时和重定向保护的 AI 客户端。
func NewAIService(cfg config.Config) *AIService {
	return &AIService{Config: cfg, HTTP: &http.Client{
		Timeout:       90 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// NewHealthService 注入健康检查所需依赖。
func NewHealthService(repo repository.HealthRepository, cached *cache.Store) *HealthService {
	return &HealthService{Repo: repo, Cache: cached}
}

// Error 表达可安全返回给客户端的业务失败。
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

func problem(status int, message string) error { return &Error{Status: status, Message: message} }

func Timestamp(value time.Time) string { return value.Format("2006-01-02T15:04:05.999999") }

func PublicUser(u model.User) dto.UserInfo {
	return dto.UserInfo{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Avatar: u.Avatar, Gender: u.Gender, Bio: u.Bio}
}

func newsDTO(n model.News) dto.NewsItem {
	return dto.NewsItem{ID: n.ID, Title: n.Title, Description: n.Description, Image: n.Image, Author: n.Author, CategoryID: n.CategoryID, Views: n.Views, PublishTime: Timestamp(n.PublishTime)}
}

func (s *HealthService) Health(ctx context.Context) (map[string]any, error) {
	if err := s.Repo.Ping(ctx); err != nil {
		return nil, problem(503, "MySQL连接异常")
	}
	redisOK := s.Cache != nil && s.Cache.Client != nil && s.Cache.Client.Ping(ctx).Err() == nil
	return map[string]any{"mysql": "ok", "redis": redisOK, "backend": "go"}, nil
}

func (s *UserService) Authenticate(ctx context.Context, token string) (model.User, error) {
	if token == "" {
		return model.User{}, problem(401, "请先登录")
	}
	userID, err := utils.ParseJWT(token, s.JWTSecret)
	if err != nil {
		return model.User{}, problem(401, "无效的令牌或已经过期的令牌")
	}
	return model.User{ID: userID}, nil
}
