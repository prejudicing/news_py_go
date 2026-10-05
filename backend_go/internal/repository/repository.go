// Package repository 隔离数据库实现，并向业务层暴露按领域划分的最小接口。
package repository

import (
	"context"
	"errors"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/prejudicing/news_py_go/backend_go/internal/model"
)

// 仓储错误屏蔽 GORM 和 MySQL 驱动类型，上层不需要依赖具体数据库实现。
var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("duplicate record")
	ErrRelation  = errors.New("related record missing")
)

// HealthRepository 提供健康检查所需的最小数据库能力。
type HealthRepository interface {
	Ping(context.Context) error
}

// UserRepository 只包含用户、登录令牌及其事务所需的数据操作。
type UserRepository interface {
	WithUserTransaction(context.Context, func(UserRepository) error) error
	UserByID(context.Context, uint64, bool) (model.User, error)
	UserByUsername(context.Context, string) (model.User, error)
	RefreshSessionByHash(context.Context, string, bool) (model.RefreshSession, error)
	CreateUser(context.Context, *model.User) error
	UpdateUser(context.Context, uint64, map[string]any) error
	CreateRefreshSession(context.Context, *model.RefreshSession) error
	RotateRefreshSession(context.Context, uint64, time.Time, string) error
	RevokeRefreshFamily(context.Context, string, time.Time) error
	RevokeAllRefreshSessions(context.Context, uint64, time.Time) error
	PruneExpiredRefreshSessions(context.Context, time.Time) error
}

// NewsRepository 提供新闻查询和浏览量更新能力。
type NewsRepository interface {
	Categories(context.Context, int, int) ([]model.Category, error)
	CountNews(context.Context, int) (int64, error)
	ListNews(context.Context, int, int, int) ([]model.News, error)
	NewsByID(context.Context, uint64) (model.News, error)
	IncreaseViews(context.Context, uint64) (int64, error)
	RelatedNews(context.Context, uint64, uint64) ([]model.News, error)
}

// FavoriteRepository 只暴露收藏业务实际使用的数据操作。
type FavoriteRepository interface {
	NewsByID(context.Context, uint64) (model.News, error)
	FavoriteExists(context.Context, uint64, uint64) (bool, error)
	CreateFavorite(context.Context, *model.Favorite) error
	DeleteFavorite(context.Context, uint64, uint64) (int64, error)
	ClearFavorites(context.Context, uint64) (int64, error)
	Favorites(context.Context, uint64, int, int) ([]FavoriteRow, int64, error)
}

// HistoryRepository 包含浏览历史更新事务需要的用户、新闻和历史操作。
type HistoryRepository interface {
	WithHistoryTransaction(context.Context, func(HistoryRepository) error) error
	UserByID(context.Context, uint64, bool) (model.User, error)
	NewsByID(context.Context, uint64) (model.News, error)
	HistoryByNews(context.Context, uint64, uint64) (model.History, error)
	CreateHistory(context.Context, *model.History) error
	UpdateHistory(context.Context, uint64, time.Time) error
	Histories(context.Context, uint64, int, int) ([]HistoryRow, int64, error)
	DeleteHistory(context.Context, uint64, uint64) (int64, error)
	ClearHistory(context.Context, uint64) (int64, error)
}

// GORMStore 是各领域仓储接口共用的 GORM 实现。
type GORMStore struct{ db *gorm.DB }

func New(db *gorm.DB) *GORMStore { return &GORMStore{db: db} }

func (r *GORMStore) WithUserTransaction(ctx context.Context, fn func(UserRepository) error) error {
	return r.withTransaction(ctx, func(tx *GORMStore) error { return fn(tx) })
}

func (r *GORMStore) WithHistoryTransaction(ctx context.Context, fn func(HistoryRepository) error) error {
	return r.withTransaction(ctx, func(tx *GORMStore) error { return fn(tx) })
}

func (r *GORMStore) withTransaction(ctx context.Context, fn func(*GORMStore) error) error {
	return translate(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&GORMStore{db: tx})
	}))
}

func (r *GORMStore) Ping(ctx context.Context) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

func translate(err error) error {
	var mysqlErr *driver.MySQLError
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.As(err, &mysqlErr) && mysqlErr.Number == 1062:
		return ErrDuplicate
	case errors.As(err, &mysqlErr) && mysqlErr.Number == 1452:
		return ErrRelation
	default:
		return err
	}
}

var (
	_ HealthRepository   = (*GORMStore)(nil)
	_ UserRepository     = (*GORMStore)(nil)
	_ NewsRepository     = (*GORMStore)(nil)
	_ FavoriteRepository = (*GORMStore)(nil)
	_ HistoryRepository  = (*GORMStore)(nil)
)
