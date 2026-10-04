package model

import "time"

// 映射已有的新闻内容表。
type News struct {
	ID                   uint64  `gorm:"primaryKey"`
	Title                string  `gorm:"size:255"`
	Description          *string `gorm:"size:500"`
	Content              string  `gorm:"type:text"`
	Image                *string `gorm:"size:255"`
	Author               *string `gorm:"size:50"`
	CategoryID           uint64  `gorm:"index"`
	Views                int64
	PublishTime          time.Time
	CreatedAt, UpdatedAt time.Time
}

// 显式指定数据库表名为 news，避免 GORM 默认复数表名。
func (News) TableName() string { return "news" }
