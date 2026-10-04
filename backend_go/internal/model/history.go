package model

import "time"

// 映射用户的浏览历史记录。
type History struct {
	ID       uint64    `gorm:"primaryKey" json:"id"`
	UserID   uint64    `gorm:"index" json:"user_id"`
	NewsID   uint64    `gorm:"index" json:"news_id"`
	ViewTime time.Time `json:"view_time"`
}

// 显式指定数据库表名为 history，避免 GORM 默认复数表名。
func (History) TableName() string { return "history" }
