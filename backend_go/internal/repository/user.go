// 声明数据持久化和仓储接口包。
package repository

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"
	// 引入time，用于时间和有效期计算。
	"time"

	// 引入gorm.io/gorm/clause，用于GORM 行锁。
	"gorm.io/gorm/clause"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 结束当前声明或代码作用域。
)

// 按用户 ID 查询，按需锁定该用户行。
func (r *GORMStore) UserByID(ctx context.Context, id uint64, lock bool) (model.User, error) {
	// 声明用户实体。
	var user model.User
	// 绑定请求上下文并执行仓储层数据库操作。
	query := r.db.WithContext(ctx)
	// 需要串行更新用户数据时增加行级写锁。
	if lock {
		// 通过 FOR UPDATE 锁住用户行，串行执行相关更新。
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		// 结束当前声明或代码作用域。
	}
	// 执行用户主键查询并保存查询错误。
	err := query.First(&user, id).Error
	// 返回查询结果，并转换数据库专属错误。
	return user, translate(err)
	// 结束当前声明或代码作用域。
}

// 按用户名查询登录用户。
func (r *GORMStore) UserByUsername(ctx context.Context, name string) (model.User, error) {
	// 声明用户实体。
	var user model.User
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Where("username = ?", name).Take(&user).Error
	// 返回查询结果，并转换数据库专属错误。
	return user, translate(err)
	// 结束当前声明或代码作用域。
}

// 关联令牌表查询尚未过期的用户。
func (r *GORMStore) UserByToken(ctx context.Context, token string, now time.Time) (model.User, error) {
	// 声明用户实体。
	var user model.User
	// 绑定请求上下文并执行仓储层数据库操作。
	err := r.db.WithContext(ctx).Table("user").Select("user.*").
		// 关联新闻或用户表，并限定当前用户的数据范围。
		Joins("JOIN user_token ON user_token.user_id = user.id").
		// 筛选指定且尚未过期的令牌，查询对应的用户实体。
		Where("user_token.token = ? AND user_token.expires_at > ?", token, now).Take(&user).Error
	// 返回查询结果，并转换数据库专属错误。
	return user, translate(err)
	// 结束当前声明或代码作用域。
}

// 写入新用户实体。
func (r *GORMStore) CreateUser(ctx context.Context, user *model.User) error {
	// 返回查询结果，并转换数据库专属错误。
	return translate(r.db.WithContext(ctx).Create(user).Error)
	// 结束当前声明或代码作用域。
}

// 更新指定的用户资料字段。
func (r *GORMStore) UpdateUser(ctx context.Context, id uint64, fields map[string]any) error {
	// 返回查询结果，并转换数据库专属错误。
	return translate(r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(fields).Error)
	// 结束当前声明或代码作用域。
}

// 删除该用户的旧令牌。
func (r *GORMStore) DeleteTokens(ctx context.Context, id uint64) error {
	// 返回查询结果，并转换数据库专属错误。
	return translate(r.db.WithContext(ctx).Where("user_id = ?", id).Delete(&model.UserToken{}).Error)
	// 结束当前声明或代码作用域。
}

// 写入新的登录令牌。
func (r *GORMStore) CreateToken(ctx context.Context, token *model.UserToken) error {
	// 返回查询结果，并转换数据库专属错误。
	return translate(r.db.WithContext(ctx).Create(token).Error)
	// 结束当前声明或代码作用域。
}
