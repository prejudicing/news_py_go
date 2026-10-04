package model

import "time"

// 映射已有的用户令牌表。
type UserToken struct {
	ID                   uint64 `gorm:"primaryKey"`
	UserID               uint64 `gorm:"index"`
	Token                string `gorm:"size:255;uniqueIndex:token_UNIQUE"`
	ExpiresAt, CreatedAt time.Time
}

// 显式指定数据库表名为 user_token，避免 GORM 默认复数表名。
func (UserToken) TableName() string { return "user_token" }
