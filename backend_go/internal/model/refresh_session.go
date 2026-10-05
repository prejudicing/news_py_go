package model

import "time"

type RefreshSession struct {
	ID              uint64    `gorm:"primaryKey"`
	UserID          uint64    `gorm:"index:idx_refresh_sessions_user"`
	FamilyID        string    `gorm:"size:32;index:idx_refresh_sessions_family"`
	TokenHash       string    `gorm:"size:64;uniqueIndex:uidx_refresh_sessions_token_hash"`
	ExpiresAt       time.Time `gorm:"index:idx_refresh_sessions_expiry"`
	FamilyExpiresAt time.Time `gorm:"index:idx_refresh_sessions_family_expiry"`
	RevokedAt       *time.Time
	ReplacedByHash  *string `gorm:"size:64"`
	CreatedAt       time.Time
}

func (RefreshSession) TableName() string { return "refresh_sessions" }
