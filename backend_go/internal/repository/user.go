package repository

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

	"github.com/prejudicing/news_py_go/backend_go/internal/model"
)

// 按用户 ID 查询，按需锁定该用户行。
func (r *GORMStore) UserByID(ctx context.Context, id uint64, lock bool) (model.User, error) {
	var user model.User
	query := r.db.WithContext(ctx)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.First(&user, id).Error
	return user, translate(err)
}

// 按用户名查询登录用户。
func (r *GORMStore) UserByUsername(ctx context.Context, name string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", name).Take(&user).Error
	return user, translate(err)
}

func (r *GORMStore) RefreshSessionByHash(ctx context.Context, tokenHash string, lock bool) (model.RefreshSession, error) {
	var session model.RefreshSession
	query := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Take(&session).Error
	return session, translate(err)
}

// 写入新用户实体。
func (r *GORMStore) CreateUser(ctx context.Context, user *model.User) error {
	return translate(r.db.WithContext(ctx).Create(user).Error)
}

// 更新指定的用户资料字段。
func (r *GORMStore) UpdateUser(ctx context.Context, id uint64, fields map[string]any) error {
	return translate(r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(fields).Error)
}

func (r *GORMStore) CreateRefreshSession(ctx context.Context, session *model.RefreshSession) error {
	return translate(r.db.WithContext(ctx).Create(session).Error)
}

func (r *GORMStore) RotateRefreshSession(ctx context.Context, id uint64, now time.Time, replacementHash string) error {
	result := r.db.WithContext(ctx).Model(&model.RefreshSession{}).
		Where("id = ? AND revoked_at IS NULL AND expires_at > ? AND family_expires_at > ?", id, now, now).
		Updates(map[string]any{"revoked_at": now, "replaced_by_hash": replacementHash})
	if result.Error != nil {
		return translate(result.Error)
	}
	if result.RowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

func (r *GORMStore) RevokeRefreshFamily(ctx context.Context, familyID string, now time.Time) error {
	return translate(r.db.WithContext(ctx).Model(&model.RefreshSession{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).Update("revoked_at", now).Error)
}

func (r *GORMStore) RevokeAllRefreshSessions(ctx context.Context, userID uint64, now time.Time) error {
	return translate(r.db.WithContext(ctx).Model(&model.RefreshSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", now).Error)
}

func (r *GORMStore) PruneExpiredRefreshSessions(ctx context.Context, now time.Time) error {
	return translate(r.db.WithContext(ctx).Where("family_expires_at <= ?", now).Delete(&model.RefreshSession{}).Error)
}
