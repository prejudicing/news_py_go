// 声明数据库模型包。
package model

// 使用时间类型映射数据库日期字段。
import "time"

// 映射已有的新闻内容表。
type News struct {
	// 保存 记录主键。GORM 将该字段识别为表主键。
	ID uint64 `gorm:"primaryKey"`
	// 保存 新闻标题。数据库字段长度上限为 255。
	Title string `gorm:"size:255"`
	// 保存 新闻简介。数据库字段长度上限为 500。指针允许区分空值与实际内容。
	Description *string `gorm:"size:500"`
	// 保存 新闻正文。使用数据库 TEXT 类型存储长正文。
	Content string `gorm:"type:text"`
	// 保存 新闻封面地址。数据库字段长度上限为 255。指针允许区分空值与实际内容。
	Image *string `gorm:"size:255"`
	// 保存 新闻作者。数据库字段长度上限为 50。指针允许区分空值与实际内容。
	Author *string `gorm:"size:50"`
	// 保存 新闻所属分类 ID。建立普通索引以加快筛选。
	CategoryID uint64 `gorm:"index"`
	// 保存 新闻浏览量。
	Views int64
	// 保存 新闻发布时间。
	PublishTime time.Time
	// 保存 创建时间、最后更新时间。
	CreatedAt, UpdatedAt time.Time
	// 结束当前作用域的代码块或结构体定义。
}

// 显式指定数据库表名为 news，避免 GORM 默认复数表名。
func (News) TableName() string { return "news" }
