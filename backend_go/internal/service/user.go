package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
)

// 取得字符串指针，用于可空字段。
func str(value string) *string { return &value }

// issueToken 在调用方事务内锁定用户，串行替换旧令牌。
func (s *UserService) issueToken(ctx context.Context, tx repository.UserRepository, uid uint64) (string, error) {
	if _, err := tx.UserByID(ctx, uid, true); err != nil {
		return "", err
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	if err := tx.DeleteTokens(ctx, uid); err != nil {
		return "", err
	}
	item := model.UserToken{UserID: uid, Token: token, ExpiresAt: s.now().Add(7 * 24 * time.Hour), CreatedAt: s.now()}
	return token, tx.CreateToken(ctx, &item)
}

// 创建用户并在事务中签发登录令牌。
func (s *UserService) Register(ctx context.Context, input dto.Credentials) (dto.AuthResponse, error) {
	if len(input.Password) > 72 {
		return dto.AuthResponse{}, problem(422, "密码不能超过72字节")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AuthResponse{}, problem(500, "密码处理失败")
	}
	user := model.User{Username: input.Username, Password: string(hash), Avatar: str("https://fastly.jsdelivr.net/npm/@vant/assets/cat.jpeg"), Gender: str("unknown"), Bio: str("这个人很懒，什么都没留下"), CreatedAt: s.now(), UpdatedAt: s.now()}
	var token string
	err = s.Repo.WithUserTransaction(ctx, func(tx repository.UserRepository) error {
		if err := tx.CreateUser(ctx, &user); err != nil {
			return err
		}
		var err error
		token, err = s.issueToken(ctx, tx, user.ID)
		return err
	})
	return dto.AuthResponse{Token: token, UserInfo: PublicUser(user)}, err
}

// verifyPassword 兼容 Python bcrypt 对超过 72 字节旧密码的截断行为。
func verifyPassword(password, hash string) bool {
	bytes := []byte(password)
	if len(bytes) > 72 {
		bytes = bytes[:72]
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), bytes) == nil
}

// 校验登录凭据并在事务中更新令牌。
func (s *UserService) Login(ctx context.Context, input dto.Credentials) (dto.AuthResponse, error) {
	user, err := s.Repo.UserByUsername(ctx, input.Username)
	if errors.Is(err, repository.ErrNotFound) {
		return dto.AuthResponse{}, problem(401, "用户名或密码错误")
	}
	if err != nil {
		return dto.AuthResponse{}, err
	}
	if !verifyPassword(input.Password, user.Password) {
		return dto.AuthResponse{}, problem(401, "用户名或密码错误")
	}
	var token string
	err = s.Repo.WithUserTransaction(ctx, func(tx repository.UserRepository) error {
		var err error
		token, err = s.issueToken(ctx, tx, user.ID)
		return err
	})
	return dto.AuthResponse{Token: token, UserInfo: PublicUser(user)}, err
}

// 更新指定的用户资料字段。
func (s *UserService) UpdateUser(ctx context.Context, user model.User, input dto.ProfileRequest) (dto.UserInfo, error) {
	updates := make(map[string]any)
	for key, value := range map[string]*string{"nickname": input.Nickname, "avatar": input.Avatar, "gender": input.Gender, "bio": input.Bio, "phone": input.Phone} {
		if value != nil {
			updates[key] = *value
		}
	}
	if len(updates) > 0 {
		updates["updated_at"] = s.now()
		if err := s.Repo.UpdateUser(ctx, user.ID, updates); err != nil {
			return dto.UserInfo{}, err
		}
		var err error
		user, err = s.Repo.UserByID(ctx, user.ID, false)
		if err != nil {
			return dto.UserInfo{}, err
		}
	}
	return PublicUser(user), nil
}

// 验证旧密码并保存新密码哈希。
func (s *UserService) ChangePassword(ctx context.Context, user model.User, input dto.PasswordRequest) error {
	if utf8.RuneCountInString(input.New) < 6 || len(input.New) > 72 {
		return problem(422, "新密码至少6个字符，且不能超过72字节")
	}
	if !verifyPassword(input.Old, user.Password) {
		return problem(400, "旧密码错误")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.New), bcrypt.DefaultCost)
	if err != nil {
		return problem(500, "密码处理失败")
	}
	return s.Repo.UpdateUser(ctx, user.ID, map[string]any{"password": string(hash), "updated_at": s.now()})
}
