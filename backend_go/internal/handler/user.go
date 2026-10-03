// 声明HTTP 路由和请求处理包。
package handler

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入github.com/gin-gonic/gin，用于Gin 路由和请求上下文。
	"github.com/gin-gonic/gin"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/service，用于业务逻辑。
	"github.com/prejudicing/news_py_go/backend_go/internal/service"
	// 结束当前声明或代码作用域。
)

// 绑定注册参数并调用业务服务。
func (s *Handler) register(c *gin.Context) {
	// 声明请求参数。
	var input dto.Credentials
	// 绑定和校验请求参数，失败时终止处理。
	if !bind(c, &input) {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 创建用户并在事务中签发登录令牌并保存操作结果。
	data, err := s.Service.Register(c.Request.Context(), input)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "注册成功", data, err)
	// 结束当前声明或代码作用域。
}

// 绑定登录参数并调用业务服务。
func (s *Handler) login(c *gin.Context) {
	// 声明请求参数。
	var input dto.Credentials
	// 绑定和校验请求参数，失败时终止处理。
	if !bind(c, &input) {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 校验登录凭据并在事务中更新令牌并保存操作结果。
	data, err := s.Service.Login(c.Request.Context(), input)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "登录成功", data, err)
	// 结束当前声明或代码作用域。
}

// 返回认证用户的公开资料。
func (s *Handler) userInfo(c *gin.Context) {
	// 构造不含密码的公开用户资料并保存操作结果。
	ok(c, "获取用户信息成功", service.PublicUser(currentUser(c)))
	// 结束当前声明或代码作用域。
}

// 绑定资料更新参数并调用业务服务。
func (s *Handler) updateUser(c *gin.Context) {
	// 声明请求参数。
	var input dto.ProfileRequest
	// 绑定和校验请求参数，失败时终止处理。
	if !bind(c, &input) {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 更新指定的用户资料字段并保存操作结果。
	data, err := s.Service.UpdateUser(c.Request.Context(), currentUser(c), input)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "更新用户信息成功", data, err)
	// 结束当前声明或代码作用域。
}

// 绑定密码参数并调用业务服务。
func (s *Handler) changePassword(c *gin.Context) {
	// 声明请求参数。
	var input dto.PasswordRequest
	// 绑定和校验请求参数，失败时终止处理。
	if !bind(c, &input) {
		// 结束当前处理，避免继续执行后续操作。
		return
		// 结束当前声明或代码作用域。
	}
	// 验证旧密码并保存新密码哈希并保存操作结果。
	err := s.Service.ChangePassword(c.Request.Context(), currentUser(c), input)
	// 统一输出成功数据或转换后的错误响应并保存操作结果。
	s.respond(c, "修改密码成功", nil, err)
	// 结束当前声明或代码作用域。
}
