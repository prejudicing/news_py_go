package model

import "time"

// 映射已有的新闻分类表。
type Category struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;uniqueIndex" json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// 显式指定数据库表名为 news_category，避免 GORM 默认复数表名。
func (Category) TableName() string { return "news_category" }
