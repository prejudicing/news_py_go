package repository

import (
	"context"
	"time"

	"github.com/prejudicing/news_py_go/backend_go/internal/model"
)

// 定义收藏与新闻关联查询结果。
type FavoriteRow struct {
	model.News   `gorm:"embedded"`
	FavoriteTime time.Time
	FavoriteID   uint64
}

// 统计该用户对指定新闻的收藏。
func (r *GORMStore) FavoriteExists(ctx context.Context, uid, news uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Favorite{}).Where("user_id = ? AND news_id = ?", uid, news).Count(&count).Error
	return count > 0, translate(err)
}

// 写入收藏记录。
func (r *GORMStore) CreateFavorite(ctx context.Context, item *model.Favorite) error {
	return translate(r.db.WithContext(ctx).Create(item).Error)
}

// 限定用户和新闻删除收藏，返回影响行数。
func (r *GORMStore) DeleteFavorite(ctx context.Context, uid, news uint64) (int64, error) {
	result := r.db.WithContext(ctx).Where("user_id = ? AND news_id = ?", uid, news).Delete(&model.Favorite{})
	return result.RowsAffected, translate(result.Error)
}

// 清空当前用户的收藏并返回数量。
func (r *GORMStore) ClearFavorites(ctx context.Context, uid uint64) (int64, error) {
	result := r.db.WithContext(ctx).Where("user_id = ?", uid).Delete(&model.Favorite{})
	return result.RowsAffected, translate(result.Error)
}

// 查询当前用户的收藏分页列表。
func (r *GORMStore) Favorites(ctx context.Context, uid uint64, offset, limit int) ([]FavoriteRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Favorite{}).Where("user_id = ?", uid).Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	rows := make([]FavoriteRow, 0)
	err := r.db.WithContext(ctx).Table("news").Select("news.*, favorite.created_at AS favorite_time, favorite.id AS favorite_id").
		Joins("JOIN favorite ON favorite.news_id = news.id").Where("favorite.user_id = ?", uid).
		Order("favorite.created_at DESC, favorite.id DESC").Offset(offset).Limit(limit).Scan(&rows).Error
	return rows, total, translate(err)
}
