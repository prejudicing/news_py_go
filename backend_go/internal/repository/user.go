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

// 关联令牌表查询尚未过期的用户。
func (r *GORMStore) UserByToken(ctx context.Context, token string, now time.Time) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Table("user").Select("user.*").
		Joins("JOIN user_token ON user_token.user_id = user.id").
		Where("user_token.token = ? AND user_token.expires_at > ?", token, now).Take(&user).Error
	return user, translate(err)
}

// 写入新用户实体。
func (r *GORMStore) CreateUser(ctx context.Context, user *model.User) error {
	return translate(r.db.WithContext(ctx).Create(user).Error)
}

// 更新指定的用户资料字段。
func (r *GORMStore) UpdateUser(ctx context.Context, id uint64, fields map[string]any) error {
	return translate(r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(fields).Error)
}

// 删除该用户的旧令牌。
func (r *GORMStore) DeleteTokens(ctx context.Context, id uint64) error {
	return translate(r.db.WithContext(ctx).Where("user_id = ?", id).Delete(&model.UserToken{}).Error)
}

// 写入新的登录令牌。
func (r *GORMStore) CreateToken(ctx context.Context, token *model.UserToken) error {
	return translate(r.db.WithContext(ctx).Create(token).Error)
}
