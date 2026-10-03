// 声明数据库模型包。
package model

// 使用时间类型映射数据库日期字段。
import "time"

// 映射已有的新闻分类表。
type Category struct {
	// 保存 记录主键。GORM 将该字段识别为表主键。
	ID uint64 `gorm:"primaryKey" json:"id"`
	// 保存 分类名称。数据库唯一索引防止重复值。数据库字段长度上限为 50。
	Name string `gorm:"size:50;uniqueIndex" json:"name"`
	// 保存 分类排序值。JSON 标签指定前端字段名。
	SortOrder int `json:"sort_order"`
	// 保存 创建时间。JSON 标签指定前端字段名。
	CreatedAt time.Time `json:"-"`
	// 保存 最后更新时间。JSON 标签指定前端字段名。
	UpdatedAt time.Time `json:"-"`
	// 结束当前作用域的代码块或结构体定义。
}

// 显式指定数据库表名为 news_category，避免 GORM 默认复数表名。
func (Category) TableName() string { return "news_category" }
