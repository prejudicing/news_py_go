// 声明独立业务逻辑包。
package service

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"
	// 引入errors，用于错误类型匹配和构造。
	"errors"
	// 引入net/http，用于HTTP 请求、客户端和状态常量。
	"net/http"
	// 引入time，用于时间和有效期计算。
	"time"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/cache，用于Redis 缓存封装。
	"github.com/prejudicing/news_py_go/backend_go/internal/cache"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/config，用于环境配置。
	"github.com/prejudicing/news_py_go/backend_go/internal/config"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/repository，用于仓储接口和统一错误。
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	// 结束当前声明或代码作用域。
)

// Service 通过仓储接口执行业务，不依赖 Gin 或 GORM。
type Service struct {
	// 保存业务使用的仓储接口。
	Repo repository.Store
	// 保存Redis 缓存封装。
	Cache *cache.Store
	// 保存应用配置。
	Config config.Config
	// 保存AI 上游 HTTP 客户端。
	HTTP *http.Client
	// 结束当前声明或代码作用域。
}

// 创建实例并注入所需依赖。
func New(repo repository.Store, cached *cache.Store, cfg config.Config) *Service {
	// 返回当前计算或构造的结果。
	return &Service{Repo: repo, Cache: cached, Config: cfg, HTTP: &http.Client{
		// 限制 AI 请求最多持续九十秒。
		Timeout: 90 * time.Second,
		// 禁止自动跟随重定向，避免把 AI 密钥发送到意外地址。
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		// 结束当前回调或复合值构造。
	}}
	// 结束当前声明或代码作用域。
}

// Error 表达业务失败，底层仓储错误由 HTTP 层统一转换。
type Error struct {
	// 保存HTTP 状态码。
	Status int
	// 保存公开错误消息。
	Message string
	// 结束当前声明或代码作用域。
}

// 返回业务错误的公开消息。
func (e *Error) Error() string { return e.Message }

// 构造包含状态码和消息的业务错误。
func problem(status int, message string) error { return &Error{Status: status, Message: message} }

// 获取应用时区中的当前时间。
func (s *Service) now() time.Time {
	// 有应用时区配置时使用指定时区。
	if s.Config.Location != nil {
		// 返回当前计算或构造的结果。
		return time.Now().In(s.Config.Location)
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return time.Now()
	// 结束当前声明或代码作用域。
}

// 格式化不带时区的时间字符串，兼容原前端。
func Timestamp(value time.Time) string { return value.Format("2006-01-02T15:04:05.999999") }

// 构造不含密码的公开用户资料。
func PublicUser(u model.User) dto.UserInfo {
	// 构造接口响应，返回数据和成功标记。
	return dto.UserInfo{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Avatar: u.Avatar, Gender: u.Gender, Bio: u.Bio}
	// 结束当前声明或代码作用域。
}

// 把新闻实体转换为前端使用的字段。
func newsDTO(n model.News) dto.NewsItem {
	// 构造接口响应，返回数据和成功标记。
	return dto.NewsItem{ID: n.ID, Title: n.Title, Description: n.Description, Image: n.Image, Author: n.Author, CategoryID: n.CategoryID, Views: n.Views, PublishTime: Timestamp(n.PublishTime)}
	// 结束当前声明或代码作用域。
}

// 检查 MySQL 与 Redis 的连接状态。
func (s *Service) Health(ctx context.Context) (map[string]any, error) {
	// 检查本次操作是否失败，必要时提前返回错误。
	if err := s.Repo.Ping(ctx); err != nil {
		// 返回校验失败对应的业务状态码和消息。
		return nil, problem(503, "MySQL连接异常")
		// 结束当前声明或代码作用域。
	}
	// 检查底层数据库连接并保存操作结果。
	redisOK := s.Cache != nil && s.Cache.Client != nil && s.Cache.Client.Ping(ctx).Err() == nil
	// 返回当前计算或构造的结果。
	return map[string]any{"mysql": "ok", "redis": redisOK, "backend": "go"}, nil
	// 结束当前声明或代码作用域。
}

// 校验令牌及有效期并返回所属用户。
func (s *Service) Authenticate(ctx context.Context, token string) (model.User, error) {
	// 没有令牌时要求用户先登录。
	if token == "" {
		// 返回校验失败对应的业务状态码和消息。
		return model.User{}, problem(401, "请先登录")
		// 结束当前声明或代码作用域。
	}
	// 获取应用时区中的当前时间并保存操作结果。
	user, err := s.Repo.UserByToken(ctx, token, s.now())
	// 检查错误是否表示记录不存在，选择相应的业务分支。
	if errors.Is(err, repository.ErrNotFound) {
		// 返回校验失败对应的业务状态码和消息。
		return user, problem(401, "无效的令牌或已经过期的令牌")
		// 结束当前声明或代码作用域。
	}
	// 返回当前结果与错误，由调用方决定响应或事务回滚。
	return user, err
	// 结束当前声明或代码作用域。
}
