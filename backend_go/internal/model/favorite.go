// 声明数据库模型包。
package model

// 使用时间类型映射数据库日期字段。
import "time"

// 映射用户与新闻之间的收藏关系。
type Favorite struct {
	// 保存 记录主键。GORM 将该字段识别为表主键。
	ID uint64 `gorm:"primaryKey" json:"id"`
	// 保存 所属用户 ID。参与用户与新闻联合唯一索引，防止重复收藏。
	UserID uint64 `gorm:"uniqueIndex:user_news_unique" json:"user_id"`
	// 保存 关联的新闻 ID。参与用户与新闻联合唯一索引，防止重复收藏。
	NewsID uint64 `gorm:"uniqueIndex:user_news_unique" json:"news_id"`
	// 保存 创建时间。JSON 标签指定前端字段名。
	CreatedAt time.Time `json:"created_at"`
	// 结束当前作用域的代码块或结构体定义。
}

// 显式指定数据库表名为 favorite，避免 GORM 默认复数表名。
func (Favorite) TableName() string { return "favorite" }
