// 声明数据库模型包。
package model

// 使用时间类型映射数据库日期字段。
import "time"

// 映射已有的用户表，包含资料、密码哈希和时间字段。
type User struct {
	// 保存 记录主键。GORM 将该字段识别为表主键。
	ID uint64 `gorm:"primaryKey"`
	// 保存 用户名。数据库唯一索引防止重复值。数据库字段长度上限为 50。
	Username string `gorm:"size:50;uniqueIndex:username_UNIQUE"`
	// 保存 数据库中的 bcrypt 密码哈希。数据库字段长度上限为 255。
	Password string `gorm:"size:255"`
	// 保存 用户昵称。数据库字段长度上限为 50。指针允许区分空值与实际内容。
	Nickname *string `gorm:"size:50"`
	// 保存 头像地址。数据库字段长度上限为 255。指针允许区分空值与实际内容。
	Avatar *string `gorm:"size:255"`
	// 保存 用户性别。数据库只接受 male、female、unknown 三种值。指针允许区分空值与实际内容。
	Gender *string `gorm:"type:enum('male','female','unknown')"`
	// 保存 个人简介。数据库字段长度上限为 500。指针允许区分空值与实际内容。
	Bio *string `gorm:"size:500"`
	// 保存 手机号。数据库唯一索引防止重复值。数据库字段长度上限为 20。指针允许区分空值与实际内容。
	Phone *string `gorm:"size:20;uniqueIndex:phone_UNIQUE"`
	// 保存 创建时间、最后更新时间。
	CreatedAt, UpdatedAt time.Time
	// 结束当前作用域的代码块或结构体定义。
}

// 显式指定数据库表名为 user，避免 GORM 默认复数表名。
func (User) TableName() string { return "user" }
