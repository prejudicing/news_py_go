package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
)

// 绑定注册参数并调用业务服务。
func (s *Handler) register(c *gin.Context) {
	var input dto.Credentials
	if !bind(c, &input) {
		return
	}
	data, err := s.user.Register(c.Request.Context(), input)
	s.respond(c, "注册成功", data, err)
}

// 绑定登录参数并调用业务服务。
func (s *Handler) login(c *gin.Context) {
	var input dto.Credentials
	if !bind(c, &input) {
		return
	}
	data, err := s.user.Login(c.Request.Context(), input)
	s.respond(c, "登录成功", data, err)
}

// 返回认证用户的公开资料。
func (s *Handler) userInfo(c *gin.Context) {
	ok(c, "获取用户信息成功", service.PublicUser(currentUser(c)))
}

// 绑定资料更新参数并调用业务服务。
func (s *Handler) updateUser(c *gin.Context) {
	var input dto.ProfileRequest
	if !bind(c, &input) {
		return
	}
	data, err := s.user.UpdateUser(c.Request.Context(), currentUser(c), input)
	s.respond(c, "更新用户信息成功", data, err)
}

// 绑定密码参数并调用业务服务。
func (s *Handler) changePassword(c *gin.Context) {
	var input dto.PasswordRequest
	if !bind(c, &input) {
		return
	}
	err := s.user.ChangePassword(c.Request.Context(), currentUser(c), input)
	s.respond(c, "修改密码成功", nil, err)
}
