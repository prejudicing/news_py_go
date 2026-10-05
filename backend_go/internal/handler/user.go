package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
	"github.com/prejudicing/news_py_go/backend_go/internal/utils"
)

// 绑定注册参数并调用业务服务。
func (s *Handler) register(c *gin.Context) {
	var input dto.Credentials
	if !bind(c, &input) {
		return
	}
	data, err := s.user.Register(c.Request.Context(), input)
	if err == nil {
		s.setRefreshCookie(c, data.RefreshToken)
	}
	s.respond(c, "注册成功", data, err)
}

// 绑定登录参数并调用业务服务。
func (s *Handler) login(c *gin.Context) {
	var input dto.Credentials
	if !bind(c, &input) {
		return
	}
	data, err := s.user.Login(c.Request.Context(), input)
	if err == nil {
		s.setRefreshCookie(c, data.RefreshToken)
	}
	s.respond(c, "登录成功", data, err)
}

func (s *Handler) refresh(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		s.clearRefreshCookie(c)
		_, refreshErr := s.user.Refresh(c.Request.Context(), "")
		s.respond(c, "", nil, refreshErr)
		return
	}
	data, err := s.user.Refresh(c.Request.Context(), cookie)
	if err == nil {
		s.setRefreshCookie(c, data.RefreshToken)
	} else {
		var business *service.Error
		if errors.As(err, &business) && business.Status == http.StatusUnauthorized {
			s.clearRefreshCookie(c)
		}
	}
	s.respond(c, "令牌已刷新", data, err)
}

func (s *Handler) logout(c *gin.Context) {
	cookie, _ := c.Cookie("refresh_token")
	err := s.user.Logout(c.Request.Context(), cookie)
	s.clearRefreshCookie(c)
	s.respond(c, "已退出登录", nil, err)
}

func (s *Handler) setRefreshCookie(c *gin.Context, value string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", value, int(utils.RefreshTokenTTL.Seconds()), "/api/user", "", s.cookieSecure, true)
}

func (s *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", "", -1, "/api/user", "", s.cookieSecure, true)
}

// 返回认证用户的公开资料。
func (s *Handler) userInfo(c *gin.Context) {
	data, err := s.user.UserInfo(c.Request.Context(), currentUser(c).ID)
	s.respond(c, "获取用户信息成功", data, err)
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
