package model

import "time"

// 映射已有的用户表，包含资料、密码哈希和时间字段。
type User struct {
	ID                   uint64  `gorm:"primaryKey"`
	Username             string  `gorm:"size:50;uniqueIndex:username_UNIQUE"`
	Password             string  `gorm:"size:255"`
	Nickname             *string `gorm:"size:50"`
	Avatar               *string `gorm:"size:255"`
	Gender               *string `gorm:"type:enum('male','female','unknown')"`
	Bio                  *string `gorm:"size:500"`
	Phone                *string `gorm:"size:20;uniqueIndex:phone_UNIQUE"`
	CreatedAt, UpdatedAt time.Time
}

// 显式指定数据库表名为 user，避免 GORM 默认复数表名。
func (User) TableName() string { return "user" }
