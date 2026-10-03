// 声明接口请求和响应结构包。
package dto

// 定义注册和登录的请求参数。
type Credentials struct {
	// 保存登录用户名；JSON 标签定义接口字段名，binding 标签约束输入。
	Username string `json:"username" binding:"required,max=50"`
	// 保存登录密码或密码哈希；JSON 标签定义接口字段名，binding 标签约束输入。
	Password string `json:"password" binding:"required"`
	// 结束当前声明或代码作用域。
}

// 定义不含密码的用户响应。
type UserInfo struct {
	// 保存记录主键；JSON 标签定义接口字段名。
	ID uint64 `json:"id"`
	// 保存登录用户名；JSON 标签定义接口字段名。
	Username string `json:"username"`
	// 保存用户昵称；JSON 标签定义接口字段名。
	Nickname *string `json:"nickname"`
	// 保存用户头像地址；JSON 标签定义接口字段名。
	Avatar *string `json:"avatar"`
	// 保存用户性别；JSON 标签定义接口字段名。
	Gender *string `json:"gender"`
	// 保存用户简介；JSON 标签定义接口字段名。
	Bio *string `json:"bio"`
	// 结束当前声明或代码作用域。
}

// 定义登录令牌及用户信息响应。
type AuthResponse struct {
	// 保存登录令牌；JSON 标签定义接口字段名。
	Token string `json:"token"`
	// 保存公开用户资料；JSON 标签定义接口字段名。
	UserInfo UserInfo `json:"userInfo"`
	// 结束当前声明或代码作用域。
}

// 定义可选的用户资料更新参数。
type ProfileRequest struct {
	// 保存用户昵称；JSON 标签定义接口字段名，binding 标签约束输入。
	Nickname *string `json:"nickname" binding:"omitempty,max=50"`
	// 保存用户头像地址；JSON 标签定义接口字段名，binding 标签约束输入。
	Avatar *string `json:"avatar" binding:"omitempty,max=255"`
	// 保存用户性别；JSON 标签定义接口字段名，binding 标签约束输入。
	Gender *string `json:"gender" binding:"omitempty,oneof=male female unknown"`
	// 保存用户简介；JSON 标签定义接口字段名，binding 标签约束输入。
	Bio *string `json:"bio" binding:"omitempty,max=500"`
	// 保存用户手机号；JSON 标签定义接口字段名，binding 标签约束输入。
	Phone *string `json:"phone" binding:"omitempty,max=20"`
	// 结束当前声明或代码作用域。
}

// 定义旧密码和新密码请求。
type PasswordRequest struct {
	// 保存旧密码；JSON 标签定义接口字段名，binding 标签约束输入。
	Old string `json:"oldPassword" binding:"required"`
	// 保存新密码；JSON 标签定义接口字段名，binding 标签约束输入。
	New string `json:"newPassword" binding:"required"`
	// 结束当前声明或代码作用域。
}
