// 声明数据持久化和仓储接口包。
package repository

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"
	// 引入errors，用于错误类型匹配和构造。
	"errors"
	// 引入time，用于时间和有效期计算。
	"time"

	// 引入github.com/go-sql-driver/mysql，用于MySQL 驱动错误类型。
	driver "github.com/go-sql-driver/mysql"
	// 引入gorm.io/gorm，用于GORM 查询与事务。
	"gorm.io/gorm"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 结束当前声明或代码作用域。
)

// 仓储错误屏蔽 GORM 和 MySQL 驱动类型，上层无需依赖数据库实现。
var (
	// 定义找不到记录时的仓储错误。
	ErrNotFound = errors.New("record not found")
	// 定义唯一约束冲突时的仓储错误。
	ErrDuplicate = errors.New("duplicate record")
	// 定义关联记录不存在时的仓储错误。
	ErrRelation = errors.New("related record missing")

// 结束当前声明或代码作用域。
)

// Store 定义业务层需要的数据操作，SQL、行锁和驱动错误都由仓储实现。
type Store interface {
	// 声明仓储方法：执行事务回调，失败时回滚全部写入。
	Transaction(context.Context, func(Store) error) error
	// 声明仓储方法：检查底层数据库连接。
	Ping(context.Context) error
	// 声明仓储方法：按用户 ID 查询，按需锁定该用户行。
	UserByID(context.Context, uint64, bool) (model.User, error)
	// 声明仓储方法：按用户名查询登录用户。
	UserByUsername(context.Context, string) (model.User, error)
	// 声明仓储方法：关联令牌表查询尚未过期的用户。
	UserByToken(context.Context, string, time.Time) (model.User, error)
	// 声明仓储方法：写入新用户实体。
	CreateUser(context.Context, *model.User) error
	// 声明仓储方法：更新指定的用户资料字段。
	UpdateUser(context.Context, uint64, map[string]any) error
	// 声明仓储方法：删除该用户的旧令牌。
	DeleteTokens(context.Context, uint64) error
	// 声明仓储方法：写入新的登录令牌。
	CreateToken(context.Context, *model.UserToken) error
	// 声明仓储方法：读取排序后的新闻分类。
	Categories(context.Context, int, int) ([]model.Category, error)
	// 声明仓储方法：统计指定分类的新闻总数。
	CountNews(context.Context, int) (int64, error)
	// 声明仓储方法：按偏移和数量查询指定分类的新闻。
	ListNews(context.Context, int, int, int) ([]model.News, error)
	// 声明仓储方法：按主键查找新闻实体。
	NewsByID(context.Context, uint64) (model.News, error)
	// 声明仓储方法：在数据库事务中递增并读取最新浏览量。
	IncreaseViews(context.Context, uint64) (int64, error)
	// 声明仓储方法：查询同分类的其他热门新闻。
	RelatedNews(context.Context, uint64, uint64) ([]model.News, error)
	// 声明仓储方法：统计该用户对指定新闻的收藏。
	FavoriteExists(context.Context, uint64, uint64) (bool, error)
	// 声明仓储方法：写入收藏记录。
	CreateFavorite(context.Context, *model.Favorite) error
	// 声明仓储方法：限定用户和新闻删除收藏，返回影响行数。
	DeleteFavorite(context.Context, uint64, uint64) (int64, error)
	// 声明仓储方法：清空当前用户的收藏并返回数量。
	ClearFavorites(context.Context, uint64) (int64, error)
	// 声明仓储方法：查询当前用户的收藏分页列表。
	Favorites(context.Context, uint64, int, int) ([]FavoriteRow, int64, error)
	// 声明仓储方法：查询当前用户对指定新闻的最新历史记录。
	HistoryByNews(context.Context, uint64, uint64) (model.History, error)
	// 声明仓储方法：写入新的浏览历史记录。
	CreateHistory(context.Context, *model.History) error
	// 声明仓储方法：按历史记录主键更新浏览时间。
	UpdateHistory(context.Context, uint64, time.Time) error
	// 声明仓储方法：查询当前用户的历史分页列表。
	Histories(context.Context, uint64, int, int) ([]HistoryRow, int64, error)
	// 声明仓储方法：按记录主键删除当前用户的历史。
	DeleteHistory(context.Context, uint64, uint64) (int64, error)
	// 声明仓储方法：清空当前用户的浏览历史。
	ClearHistory(context.Context, uint64) (int64, error)
	// 结束当前声明或代码作用域。
}

// GORMStore 只负责持久化，持有普通连接或当前事务连接。
type GORMStore struct{ db *gorm.DB }

// 创建实例并注入所需依赖。
func New(db *gorm.DB) Store { return &GORMStore{db: db} }

// 执行事务回调，失败时回滚全部写入。
func (r *GORMStore) Transaction(ctx context.Context, fn func(Store) error) error {
	// 返回查询结果，并转换数据库专属错误。
	return translate(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 返回当前计算或构造的结果。
		return fn(&GORMStore{db: tx})
		// 结束数据库事务回调，并转换底层事务错误。
	}))
	// 结束当前声明或代码作用域。
}

// 检查底层数据库连接。
func (r *GORMStore) Ping(ctx context.Context) error {
	// 取出 GORM 底层的数据库连接池。
	db, err := r.db.DB()
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return err
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return db.PingContext(ctx)
	// 结束当前声明或代码作用域。
}

// 把数据库专属错误转换为仓储层错误。
func translate(err error) error {
	// 声明MySQL 驱动错误。
	var mysqlErr *driver.MySQLError
	// 按错误类型选择公开的响应或仓储错误。
	switch {
	// 匹配当前错误类型并进入对应处理分支。
	case errors.Is(err, gorm.ErrRecordNotFound):
		// 返回当前计算或构造的结果。
		return ErrNotFound
	// 匹配当前错误类型并进入对应处理分支。
	case errors.As(err, &mysqlErr) && mysqlErr.Number == 1062:
		// 返回当前计算或构造的结果。
		return ErrDuplicate
	// 匹配当前错误类型并进入对应处理分支。
	case errors.As(err, &mysqlErr) && mysqlErr.Number == 1452:
		// 返回当前计算或构造的结果。
		return ErrRelation
	// 其他错误使用通用提示，避免泄露底层细节。
	default:
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return err
		// 结束当前声明或代码作用域。
	}
	// 结束当前声明或代码作用域。
}

// 在编译时确认 GORM 仓储完整实现接口。
var _ Store = (*GORMStore)(nil)
