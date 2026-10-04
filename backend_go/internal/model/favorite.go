package model

import "time"

// 映射用户与新闻之间的收藏关系。
type Favorite struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"uniqueIndex:user_news_unique" json:"user_id"`
	NewsID    uint64    `gorm:"uniqueIndex:user_news_unique" json:"news_id"`
	CreatedAt time.Time `json:"created_at"`
}

// 显式指定数据库表名为 favorite，避免 GORM 默认复数表名。
func (Favorite) TableName() string { return "favorite" }
