package model

import "time"

// RefreshSession 保存刷新令牌的哈希及轮换链，不保存可直接使用的明文令牌。
type RefreshSession struct {
	// ID 是会话记录主键。
	ID uint64 `gorm:"primaryKey"`
	// UserID 和 FamilyID 分别用于查询用户的会话及撤销整条设备会话链。
	UserID   uint64 `gorm:"index:idx_refresh_sessions_user"`
	FamilyID string `gorm:"size:32;index:idx_refresh_sessions_family"`
	// TokenHash 唯一标识当前刷新令牌；数据库泄露时不能直接拿它续期。
	TokenHash string `gorm:"size:64;uniqueIndex:uidx_refresh_sessions_token_hash"`
	// ExpiresAt 实施空闲期限，FamilyExpiresAt 限制整条会话族的绝对寿命。
	ExpiresAt       time.Time `gorm:"index:idx_refresh_sessions_expiry"`
	FamilyExpiresAt time.Time `gorm:"index:idx_refresh_sessions_family_expiry"`
	// RevokedAt 标记已使用或已撤销令牌；ReplacedByHash 保留轮换关系以检测重放。
	RevokedAt      *time.Time
	ReplacedByHash *string `gorm:"size:64"`
	CreatedAt      time.Time
}

// TableName 固定映射到迁移脚本创建的会话表名。
func (RefreshSession) TableName() string { return "refresh_sessions" }
