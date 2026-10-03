// 声明数据库模型包。
package model

// 使用时间类型映射数据库日期字段。
import "time"

// 映射已有的用户令牌表。
type UserToken struct {
	// 保存 记录主键。GORM 将该字段识别为表主键。
	ID uint64 `gorm:"primaryKey"`
	// 保存 所属用户 ID。建立普通索引以加快筛选。
	UserID uint64 `gorm:"index"`
	// 保存 登录令牌。数据库唯一索引防止重复值。数据库字段长度上限为 255。
	Token string `gorm:"size:255;uniqueIndex:token_UNIQUE"`
	// 保存 令牌过期时间、创建时间。
	ExpiresAt, CreatedAt time.Time
	// 结束当前作用域的代码块或结构体定义。
}

// 显式指定数据库表名为 user_token，避免 GORM 默认复数表名。
func (UserToken) TableName() string { return "user_token" }
