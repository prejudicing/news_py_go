package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/prejudicing/news_py_go/backend_go/internal/model"
)

// 读取排序后的新闻分类。
func (r *GORMStore) Categories(ctx context.Context, skip, limit int) ([]model.Category, error) {
	items := make([]model.Category, 0)
	err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Offset(skip).Limit(limit).Find(&items).Error
	return items, translate(err)
}

// 统计指定分类的新闻总数。
func (r *GORMStore) CountNews(ctx context.Context, category int) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.News{}).Where("category_id = ?", category).Count(&total).Error
	return total, translate(err)
}

// 按偏移和数量查询指定分类的新闻。
func (r *GORMStore) ListNews(ctx context.Context, category, offset, limit int) ([]model.News, error) {
	rows := make([]model.News, 0)
	err := r.db.WithContext(ctx).Where("category_id = ?", category).Order("id ASC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, translate(err)
}

// 按主键查找新闻实体。
func (r *GORMStore) NewsByID(ctx context.Context, id uint64) (model.News, error) {
	var item model.News
	err := r.db.WithContext(ctx).First(&item, id).Error
	return item, translate(err)
}

// 在数据库事务中递增并读取最新浏览量。
func (r *GORMStore) IncreaseViews(ctx context.Context, id uint64) (int64, error) {
	var views int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.News{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + 1"))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return tx.Model(&model.News{}).Select("views").Where("id = ?", id).Scan(&views).Error
	})
	return views, translate(err)
}

// 查询同分类的其他热门新闻。
func (r *GORMStore) RelatedNews(ctx context.Context, id, category uint64) ([]model.News, error) {
	rows := make([]model.News, 0)
	err := r.db.WithContext(ctx).Where("category_id = ? AND id <> ?", category, id).
		Order("views DESC, publish_time DESC, id DESC").Limit(5).Find(&rows).Error
	return rows, translate(err)
}
