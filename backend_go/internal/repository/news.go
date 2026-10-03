// 声明数据持久化和仓储接口包。
package repository

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"

	// 引入gorm.io/gorm，用于GORM 查询与事务。
	"gorm.io/gorm"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 结束当前声明或代码作用域。
)

// 读取排序后的新闻分类。
func (r *GORMStore) Categories(ctx context.Context, skip, limit int) ([]model.Category, error) {
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	items := make([]model.Category, 0)
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Offset(skip).Limit(limit).Find(&items).Error
	// 返回查询结果，并转换数据库专属错误。
	return items, translate(err)
	// 结束当前声明或代码作用域。
}

// 统计指定分类的新闻总数。
func (r *GORMStore) CountNews(ctx context.Context, category int) (int64, error) {
	// 声明总记录数。
	var total int64
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Model(&model.News{}).Where("category_id = ?", category).Count(&total).Error
	// 返回查询结果，并转换数据库专属错误。
	return total, translate(err)
	// 结束当前声明或代码作用域。
}

// 按偏移和数量查询指定分类的新闻。
func (r *GORMStore) ListNews(ctx context.Context, category, offset, limit int) ([]model.News, error) {
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	rows := make([]model.News, 0)
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Where("category_id = ?", category).Order("id ASC").Offset(offset).Limit(limit).Find(&rows).Error
	// 返回查询结果，并转换数据库专属错误。
	return rows, translate(err)
	// 结束当前声明或代码作用域。
}

// 按主键查找新闻实体。
func (r *GORMStore) NewsByID(ctx context.Context, id uint64) (model.News, error) {
	// 声明当前实体。
	var item model.News
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).First(&item, id).Error
	// 返回查询结果，并转换数据库专属错误。
	return item, translate(err)
	// 结束当前声明或代码作用域。
}

// 在数据库事务中递增并读取最新浏览量。
func (r *GORMStore) IncreaseViews(ctx context.Context, id uint64) (int64, error) {
	// 声明最新浏览次数。
	var views int64
	// 执行事务回调，失败时回滚全部写入并保存操作结果。
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 构造或保存result 操作结果。
		result := tx.Model(&model.News{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + 1"))
		// 检查当前条件并进入对应的处理分支。
		if result.Error != nil {
			// 返回当前计算或构造的结果。
			return result.Error
			// 结束当前声明或代码作用域。
		}
		// 没有影响任何记录时返回记录不存在错误。
		if result.RowsAffected == 0 {
			// 返回当前计算或构造的结果。
			return ErrNotFound
			// 结束当前声明或代码作用域。
		}
		// 返回当前计算或构造的结果。
		return tx.Model(&model.News{}).Select("views").Where("id = ?", id).Scan(&views).Error
		// 结束当前回调或复合值构造。
	})
	// 返回查询结果，并转换数据库专属错误。
	return views, translate(err)
	// 结束当前声明或代码作用域。
}

// 查询同分类的其他热门新闻。
func (r *GORMStore) RelatedNews(ctx context.Context, id, category uint64) ([]model.News, error) {
	// 创建非空列表容器，使无数据时 JSON 返回空数组。
	rows := make([]model.News, 0)
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Where("category_id = ? AND id <> ?", category, id).
		// 按业务要求排序，并限制返回记录数。
		Order("views DESC, publish_time DESC, id DESC").Limit(5).Find(&rows).Error
	// 返回查询结果，并转换数据库专属错误。
	return rows, translate(err)
	// 结束当前声明或代码作用域。
}
