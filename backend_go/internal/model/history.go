// 声明数据库模型包。
package model

// 使用时间类型映射数据库日期字段。
import "time"

// 映射用户的浏览历史记录。
type History struct {
	// 保存 记录主键。GORM 将该字段识别为表主键。
	ID uint64 `gorm:"primaryKey" json:"id"`
	// 保存 所属用户 ID。建立普通索引以加快筛选。
	UserID uint64 `gorm:"index" json:"user_id"`
	// 保存 关联的新闻 ID。建立普通索引以加快筛选。
	NewsID uint64 `gorm:"index" json:"news_id"`
	// 保存 最近浏览时间。JSON 标签指定前端字段名。
	ViewTime time.Time `json:"view_time"`
	// 结束当前作用域的代码块或结构体定义。
}

// 显式指定数据库表名为 history，避免 GORM 默认复数表名。
func (History) TableName() string { return "history" }
