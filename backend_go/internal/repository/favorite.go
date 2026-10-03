// 声明数据持久化和仓储接口包。
package repository

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"
	// 引入time，用于时间和有效期计算。
	"time"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 结束当前声明或代码作用域。
)

// 定义收藏与新闻关联查询结果。
type FavoriteRow struct {
	// 嵌入新闻实体以接收关联查询结果。
	model.News `gorm:"embedded"`
	// 保存收藏时间。
	FavoriteTime time.Time
	// 保存收藏记录主键。
	FavoriteID uint64
	// 结束当前声明或代码作用域。
}

// 统计该用户对指定新闻的收藏。
func (r *GORMStore) FavoriteExists(ctx context.Context, uid, news uint64) (bool, error) {
	// 声明count 变量。
	var count int64
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Model(&model.Favorite{}).Where("user_id = ? AND news_id = ?", uid, news).Count(&count).Error
	// 返回查询结果，并转换数据库专属错误。
	return count > 0, translate(err)
	// 结束当前声明或代码作用域。
}

// 写入收藏记录。
func (r *GORMStore) CreateFavorite(ctx context.Context, item *model.Favorite) error {
	// 返回查询结果，并转换数据库专属错误。
	return translate(r.db.WithContext(ctx).Create(item).Error)
	// 结束当前声明或代码作用域。
}

// 限定用户和新闻删除收藏，返回影响行数。
func (r *GORMStore) DeleteFavorite(ctx context.Context, uid, news uint64) (int64, error) {
	// 绑定请求上下文并执行仓储层数据库操作。
	result := r.db.WithContext(ctx).Where("user_id = ? AND news_id = ?", uid, news).Delete(&model.Favorite{})
	// 返回查询结果，并转换数据库专属错误。
	return result.RowsAffected, translate(result.Error)
	// 结束当前声明或代码作用域。
}

// 清空当前用户的收藏并返回数量。
func (r *GORMStore) ClearFavorites(ctx context.Context, uid uint64) (int64, error) {
	// 绑定请求上下文并执行仓储层数据库操作。
	result := r.db.WithContext(ctx).Where("user_id = ?", uid).Delete(&model.Favorite{})
	// 返回查询结果，并转换数据库专属错误。
	return result.RowsAffected, translate(result.Error)
	// 结束当前声明或代码作用域。
}

// 查询当前用户的收藏分页列表。
func (r *GORMStore) Favorites(ctx context.Context, uid uint64, offset, limit int) ([]FavoriteRow, int64, error) {
	// 声明总记录数。
	var total int64
	// 检查本次操作是否失败，必要时提前返回错误。
	if err := r.db.WithContext(ctx).Model(&model.Favorite{}).Where("user_id = ?", uid).Count(&total).Error; err != nil {
		// 返回查询结果，并转换数据库专属错误。
		return nil, 0, translate(err)
		// 结束当前声明或代码作用域。
	}
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	rows := make([]FavoriteRow, 0)
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Table("news").Select("news.*, favorite.created_at AS favorite_time, favorite.id AS favorite_id").
		// 关联新闻或用户表，并限定当前用户的数据范围。
		Joins("JOIN favorite ON favorite.news_id = news.id").Where("favorite.user_id = ?", uid).
		// 按业务要求排序，并限制返回记录数。
		Order("favorite.created_at DESC, favorite.id DESC").Offset(offset).Limit(limit).Scan(&rows).Error
	// 返回查询结果，并转换数据库专属错误。
	return rows, total, translate(err)
	// 结束当前声明或代码作用域。
}
