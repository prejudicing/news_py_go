package dto

// 定义注册和登录的请求参数。
type Credentials struct {
	Username string `json:"username" binding:"required,max=50"`
	Password string `json:"password" binding:"required"`
}

// 定义不含密码的用户响应。
type UserInfo struct {
	ID       uint64  `json:"id"`
	Username string  `json:"username"`
	Nickname *string `json:"nickname"`
	Avatar   *string `json:"avatar"`
	Gender   *string `json:"gender"`
	Bio      *string `json:"bio"`
}

// 定义登录令牌及用户信息响应。
type AuthResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"-"`
	UserInfo     UserInfo `json:"userInfo"`
}

// 定义可选的用户资料更新参数。
type ProfileRequest struct {
	Nickname *string `json:"nickname" binding:"omitempty,max=50"`
	Avatar   *string `json:"avatar" binding:"omitempty,max=255"`
	Gender   *string `json:"gender" binding:"omitempty,oneof=male female unknown"`
	Bio      *string `json:"bio" binding:"omitempty,max=500"`
	Phone    *string `json:"phone" binding:"omitempty,max=20"`
}

// 定义旧密码和新密码请求。
type PasswordRequest struct {
	Old string `json:"oldPassword" binding:"required"`
	New string `json:"newPassword" binding:"required"`
}
