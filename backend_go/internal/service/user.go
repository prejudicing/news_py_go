// 声明独立业务逻辑包。
package service

// 导入本文件使用的标准库、第三方依赖和内部包。
import (
	// 引入context，用于请求取消与超时上下文。
	"context"
	// 引入crypto/rand，用于安全的随机令牌字节。
	"crypto/rand"
	// 引入encoding/hex，用于随机令牌的十六进制编码。
	"encoding/hex"
	// 引入errors，用于错误类型匹配和构造。
	"errors"
	// 引入time，用于时间和有效期计算。
	"time"
	// 引入unicode/utf8，用于新密码的中文字符计数。
	"unicode/utf8"

	// 引入golang.org/x/crypto/bcrypt，用于bcrypt 密码哈希。
	"golang.org/x/crypto/bcrypt"

	// 引入github.com/prejudicing/news_py_go/backend_go/internal/dto，用于接口参数和响应结构。
	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/model，用于数据库实体。
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	// 引入github.com/prejudicing/news_py_go/backend_go/internal/repository，用于仓储接口和统一错误。
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	// 结束当前声明或代码作用域。
)

// 取得字符串指针，用于可空字段。
func str(value string) *string { return &value }

// issueToken 在调用方事务内锁定用户，串行替换旧令牌。
func (s *Service) issueToken(ctx context.Context, tx repository.Store, uid uint64) (string, error) {
	// 检查本次操作是否失败，必要时提前返回错误。
	if _, err := tx.UserByID(ctx, uid, true); err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return "", err
		// 结束当前声明或代码作用域。
	}
	// 分配 32 字节安全随机令牌缓冲。
	bytes := make([]byte, 32)
	// 检查本次操作是否失败，必要时提前返回错误。
	if _, err := rand.Read(bytes); err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return "", err
		// 结束当前声明或代码作用域。
	}
	// 将随机字节编码为十六进制令牌。
	token := hex.EncodeToString(bytes)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err := tx.DeleteTokens(ctx, uid); err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return "", err
		// 结束当前声明或代码作用域。
	}
	// 获取应用时区中的当前时间并保存操作结果。
	item := model.UserToken{UserID: uid, Token: token, ExpiresAt: s.now().Add(7 * 24 * time.Hour), CreatedAt: s.now()}
	// 写入新的登录令牌，把结果或错误返回给调用方。
	return token, tx.CreateToken(ctx, &item)
	// 结束当前声明或代码作用域。
}

// 创建用户并在事务中签发登录令牌。
func (s *Service) Register(ctx context.Context, input dto.Credentials) (dto.AuthResponse, error) {
	// 校验注册密码字节数，遵守 bcrypt 的长度限制。
	if len(input.Password) > 72 {
		// 返回校验失败对应的业务状态码和消息。
		return dto.AuthResponse{}, problem(422, "密码不能超过72字节")
		// 结束当前声明或代码作用域。
	}
	// 使用 bcrypt 默认工作因子计算新密码哈希。
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回校验失败对应的业务状态码和消息。
		return dto.AuthResponse{}, problem(500, "密码处理失败")
		// 结束当前声明或代码作用域。
	}
	// 获取应用时区中的当前时间并保存操作结果。
	user := model.User{Username: input.Username, Password: string(hash), Avatar: str("https://fastly.jsdelivr.net/npm/@vant/assets/cat.jpeg"), Gender: str("unknown"), Bio: str("这个人很懒，什么都没留下"), CreatedAt: s.now(), UpdatedAt: s.now()}
	// 声明新令牌。
	var token string
	// 执行事务回调，失败时回滚全部写入并保存操作结果。
	err = s.Repo.Transaction(ctx, func(tx repository.Store) error {
		// 检查本次操作是否失败，必要时提前返回错误。
		if err := tx.CreateUser(ctx, &user); err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return err
			// 结束当前声明或代码作用域。
		}
		// 声明操作错误。
		var err error
		// 锁定用户并在同一事务中替换旧令牌并保存操作结果。
		token, err = s.issueToken(ctx, tx, user.ID)
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return err
		// 结束当前回调或复合值构造。
	})
	// 返回当前结果与错误，由调用方决定响应或事务回滚。
	return dto.AuthResponse{Token: token, UserInfo: PublicUser(user)}, err
	// 结束当前声明或代码作用域。
}

// verifyPassword 兼容 Python bcrypt 对超过 72 字节旧密码的截断行为。
func verifyPassword(password, hash string) bool {
	// 构造或保存bytes 操作结果。
	bytes := []byte(password)
	// 兼容旧版 bcrypt 对长密码的截断处理。
	if len(bytes) > 72 {
		// 截断到 bcrypt 兼容的前 72 字节。
		bytes = bytes[:72]
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return bcrypt.CompareHashAndPassword([]byte(hash), bytes) == nil
	// 结束当前声明或代码作用域。
}

// 校验登录凭据并在事务中更新令牌。
func (s *Service) Login(ctx context.Context, input dto.Credentials) (dto.AuthResponse, error) {
	// 按用户名查询登录用户并保存操作结果。
	user, err := s.Repo.UserByUsername(ctx, input.Username)
	// 检查错误是否表示记录不存在，选择相应的业务分支。
	if errors.Is(err, repository.ErrNotFound) {
		// 返回校验失败对应的业务状态码和消息。
		return dto.AuthResponse{}, problem(401, "用户名或密码错误")
		// 结束当前声明或代码作用域。
	}
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return dto.AuthResponse{}, err
		// 结束当前声明或代码作用域。
	}
	// 校验密码是否匹配已保存的 bcrypt 哈希。
	if !verifyPassword(input.Password, user.Password) {
		// 返回校验失败对应的业务状态码和消息。
		return dto.AuthResponse{}, problem(401, "用户名或密码错误")
		// 结束当前声明或代码作用域。
	}
	// 声明新令牌。
	var token string
	// 执行事务回调，失败时回滚全部写入并保存操作结果。
	err = s.Repo.Transaction(ctx, func(tx repository.Store) error {
		// 声明操作错误。
		var err error
		// 锁定用户并在同一事务中替换旧令牌并保存操作结果。
		token, err = s.issueToken(ctx, tx, user.ID)
		// 返回当前结果与错误，由调用方决定响应或事务回滚。
		return err
		// 结束当前回调或复合值构造。
	})
	// 返回当前结果与错误，由调用方决定响应或事务回滚。
	return dto.AuthResponse{Token: token, UserInfo: PublicUser(user)}, err
	// 结束当前声明或代码作用域。
}

// 更新指定的用户资料字段。
func (s *Service) UpdateUser(ctx context.Context, user model.User, input dto.ProfileRequest) (dto.UserInfo, error) {
	// 创建用于收集待更新字段的映射。
	updates := make(map[string]any)
	// 遍历允许更新的资料字段及其可选值。
	for key, value := range map[string]*string{"nickname": input.Nickname, "avatar": input.Avatar, "gender": input.Gender, "bio": input.Bio, "phone": input.Phone} {
		// 只更新请求明确提供的字段，保留未提供的字段。
		if value != nil {
			// 把当前提供的资料字段加入更新集合。
			updates[key] = *value
			// 结束当前声明或代码作用域。
		}
		// 结束当前声明或代码作用域。
	}
	// 存在待更新字段时才执行写入和重新读取。
	if len(updates) > 0 {
		// 获取应用时区中的当前时间并保存操作结果。
		updates["updated_at"] = s.now()
		// 检查本次操作是否失败，必要时提前返回错误。
		if err := s.Repo.UpdateUser(ctx, user.ID, updates); err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return dto.UserInfo{}, err
			// 结束当前声明或代码作用域。
		}
		// 声明操作错误。
		var err error
		// 按用户 ID 查询，按需锁定该用户行并保存操作结果。
		user, err = s.Repo.UserByID(ctx, user.ID, false)
		// 检查本次操作是否失败，必要时提前返回错误。
		if err != nil {
			// 返回当前结果与错误，由调用方决定响应或事务回滚。
			return dto.UserInfo{}, err
			// 结束当前声明或代码作用域。
		}
		// 结束当前声明或代码作用域。
	}
	// 返回当前计算或构造的结果。
	return PublicUser(user), nil
	// 结束当前声明或代码作用域。
}

// 验证旧密码并保存新密码哈希。
func (s *Service) ChangePassword(ctx context.Context, user model.User, input dto.PasswordRequest) error {
	// 同时检查新密码字符数和字节数。
	if utf8.RuneCountInString(input.New) < 6 || len(input.New) > 72 {
		// 返回校验失败对应的业务状态码和消息。
		return problem(422, "新密码至少6个字符，且不能超过72字节")
		// 结束当前声明或代码作用域。
	}
	// 校验密码是否匹配已保存的 bcrypt 哈希。
	if !verifyPassword(input.Old, user.Password) {
		// 返回校验失败对应的业务状态码和消息。
		return problem(400, "旧密码错误")
		// 结束当前声明或代码作用域。
	}
	// 使用 bcrypt 默认工作因子计算新密码哈希。
	hash, err := bcrypt.GenerateFromPassword([]byte(input.New), bcrypt.DefaultCost)
	// 检查本次操作是否失败，必要时提前返回错误。
	if err != nil {
		// 返回校验失败对应的业务状态码和消息。
		return problem(500, "密码处理失败")
		// 结束当前声明或代码作用域。
	}
	// 获取应用时区中的当前时间，把结果或错误返回给调用方。
	return s.Repo.UpdateUser(ctx, user.ID, map[string]any{"password": string(hash), "updated_at": s.now()})
	// 结束当前声明或代码作用域。
}
