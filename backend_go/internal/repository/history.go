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

// 定义历史与新闻关联查询结果。
type HistoryRow struct {
	// 嵌入新闻实体以接收关联查询结果。
	model.News `gorm:"embedded"`
	// 保存浏览时间。
	ViewTime time.Time
	// 保存历史记录主键。
	HistoryID uint64
	// 结束当前声明或代码作用域。
}

// 查询当前用户对指定新闻的最新历史记录。
func (r *GORMStore) HistoryByNews(ctx context.Context, uid, news uint64) (model.History, error) {
	// 声明当前实体。
	var item model.History
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Where("user_id = ? AND news_id = ?", uid, news).Order("view_time DESC, id DESC").Take(&item).Error
	// 返回查询结果，并转换数据库专属错误。
	return item, translate(err)
	// 结束当前声明或代码作用域。
}

// 写入新的浏览历史记录。
func (r *GORMStore) CreateHistory(ctx context.Context, item *model.History) error {
	// 返回查询结果，并转换数据库专属错误。
	return translate(r.db.WithContext(ctx).Create(item).Error)
	// 结束当前声明或代码作用域。
}

// 按历史记录主键更新浏览时间。
func (r *GORMStore) UpdateHistory(ctx context.Context, id uint64, viewed time.Time) error {
	// 返回查询结果，并转换数据库专属错误。
	return translate(r.db.WithContext(ctx).Model(&model.History{}).Where("id = ?", id).Update("view_time", viewed).Error)
	// 结束当前声明或代码作用域。
}

// 查询当前用户的历史分页列表。
func (r *GORMStore) Histories(ctx context.Context, uid uint64, offset, limit int) ([]HistoryRow, int64, error) {
	// 声明总记录数。
	var total int64
	// 检查本次操作是否失败，必要时提前返回错误。
	if err := r.db.WithContext(ctx).Model(&model.History{}).Where("user_id = ?", uid).Count(&total).Error; err != nil {
		// 返回查询结果，并转换数据库专属错误。
		return nil, 0, translate(err)
		// 结束当前声明或代码作用域。
	}
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	rows := make([]HistoryRow, 0)
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Table("news").Select("news.*, history.view_time AS view_time, history.id AS history_id").
		// 关联新闻或用户表，并限定当前用户的数据范围。
		Joins("JOIN history ON history.news_id = news.id").Where("history.user_id = ?", uid).
		// 按业务要求排序，并限制返回记录数。
		Order("history.view_time DESC, history.id DESC").Offset(offset).Limit(limit).Scan(&rows).Error
	// 返回查询结果，并转换数据库专属错误。
	return rows, total, translate(err)
	// 结束当前声明或代码作用域。
}

// 按记录主键删除当前用户的历史。
func (r *GORMStore) DeleteHistory(ctx context.Context, uid, id uint64) (int64, error) {
	// 绑定请求上下文并执行仓储层数据库操作。
	result := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", uid, id).Delete(&model.History{})
	// 返回查询结果，并转换数据库专属错误。
	return result.RowsAffected, translate(result.Error)
	// 结束当前声明或代码作用域。
}

// 清空当前用户的浏览历史。
func (r *GORMStore) ClearHistory(ctx context.Context, uid uint64) (int64, error) {
	// 绑定请求上下文并执行仓储层数据库操作。
	result := r.db.WithContext(ctx).Where("user_id = ?", uid).Delete(&model.History{})
	// 返回查询结果，并转换数据库专属错误。
	return result.RowsAffected, translate(result.Error)
	// 结束当前声明或代码作用域。
}
