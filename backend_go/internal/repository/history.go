package repository

import (
	"context"
	"time"

	"github.com/prejudicing/news_py_go/backend_go/internal/model"
)

// 定义历史与新闻关联查询结果。
type HistoryRow struct {
	model.News `gorm:"embedded"`
	ViewTime   time.Time
	HistoryID  uint64
}

// 查询当前用户对指定新闻的最新历史记录。
func (r *GORMStore) HistoryByNews(ctx context.Context, uid, news uint64) (model.History, error) {
	var item model.History
	err := r.db.WithContext(ctx).Where("user_id = ? AND news_id = ?", uid, news).Order("view_time DESC, id DESC").Take(&item).Error
	return item, translate(err)
}

// 写入新的浏览历史记录。
func (r *GORMStore) CreateHistory(ctx context.Context, item *model.History) error {
	return translate(r.db.WithContext(ctx).Create(item).Error)
}

// 按历史记录主键更新浏览时间。
func (r *GORMStore) UpdateHistory(ctx context.Context, id uint64, viewed time.Time) error {
	return translate(r.db.WithContext(ctx).Model(&model.History{}).Where("id = ?", id).Update("view_time", viewed).Error)
}

// 查询当前用户的历史分页列表。
func (r *GORMStore) Histories(ctx context.Context, uid uint64, offset, limit int) ([]HistoryRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.History{}).Where("user_id = ?", uid).Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	rows := make([]HistoryRow, 0)
	err := r.db.WithContext(ctx).Table("news").Select("news.*, history.view_time AS view_time, history.id AS history_id").
		Joins("JOIN history ON history.news_id = news.id").Where("history.user_id = ?", uid).
		Order("history.view_time DESC, history.id DESC").Offset(offset).Limit(limit).Scan(&rows).Error
	return rows, total, translate(err)
}

// 按记录主键删除当前用户的历史。
func (r *GORMStore) DeleteHistory(ctx context.Context, uid, id uint64) (int64, error) {
	result := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", uid, id).Delete(&model.History{})
	return result.RowsAffected, translate(result.Error)
}

// 清空当前用户的浏览历史。
func (r *GORMStore) ClearHistory(ctx context.Context, uid uint64) (int64, error) {
	result := r.db.WithContext(ctx).Where("user_id = ?", uid).Delete(&model.History{})
	return result.RowsAffected, translate(result.Error)
}
