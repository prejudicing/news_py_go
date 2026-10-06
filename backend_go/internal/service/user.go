package service

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/prejudicing/news_py_go/backend_go/internal/dto"
	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	"github.com/prejudicing/news_py_go/backend_go/internal/utils"
)

// 取得字符串指针，用于可空字段。
func str(value string) *string { return &value }

// newSession 同时签发短期 Access JWT 和高熵 Refresh Token。
func (s *UserService) newSession(uid uint64) (string, string, error) {
	now := s.now()
	accessToken, err := utils.IssueJWT(uid, s.JWTSecret, now)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := utils.NewRefreshToken()
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// saveSession 只保存刷新凭据摘要，并确保空闲期限不会超过会话族期限。
func (s *UserService) saveSession(ctx context.Context, tx repository.UserRepository, uid uint64, familyID, refreshToken string, familyExpiresAt time.Time) error {
	now := s.now()
	expiresAt := now.Add(utils.RefreshTokenTTL)
	if familyExpiresAt.Before(expiresAt) {
		expiresAt = familyExpiresAt
	}
	item := model.RefreshSession{
		UserID: uid, FamilyID: familyID, TokenHash: utils.HashRefreshToken(refreshToken),
		ExpiresAt: expiresAt, FamilyExpiresAt: familyExpiresAt, CreatedAt: now,
	}
	return tx.CreateRefreshSession(ctx, &item)
}

// issueSession 在事务中清理过期记录、锁定用户并创建一条新的设备会话族。
func (s *UserService) issueSession(ctx context.Context, tx repository.UserRepository, uid uint64) (string, string, error) {
	if err := tx.PruneExpiredRefreshSessions(ctx, s.now()); err != nil {
		return "", "", err
	}
	if _, err := tx.UserByID(ctx, uid, true); err != nil {
		return "", "", err
	}
	familyID, err := utils.NewRefreshFamilyID()
	if err != nil {
		return "", "", err
	}
	familyExpiresAt := s.now().Add(utils.RefreshFamilyTTL)
	accessToken, refreshToken, err := s.newSession(uid)
	if err != nil {
		return "", "", err
	}
	if err := s.saveSession(ctx, tx, uid, familyID, refreshToken, familyExpiresAt); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// 创建用户并在事务中签发登录令牌。
func (s *UserService) Register(ctx context.Context, input dto.Credentials) (dto.AuthResponse, error) {
	if len(input.Password) > 72 {
		return dto.AuthResponse{}, problem(ErrorInvalidArgument, "密码不能超过72字节")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AuthResponse{}, problem(ErrorInternal, "密码处理失败")
	}
	user := model.User{Username: input.Username, Password: string(hash), Avatar: str("https://fastly.jsdelivr.net/npm/@vant/assets/cat.jpeg"), Gender: str("unknown"), Bio: str("这个人很懒，什么都没留下"), CreatedAt: s.now(), UpdatedAt: s.now()}
	var accessToken, refreshToken string
	err = s.Repo.WithUserTransaction(ctx, func(tx repository.UserRepository) error {
		if err := tx.CreateUser(ctx, &user); err != nil {
			return err
		}
		var err error
		accessToken, refreshToken, err = s.issueSession(ctx, tx, user.ID)
		return err
	})
	return dto.AuthResponse{Token: accessToken, RefreshToken: refreshToken, UserInfo: PublicUser(user)}, err
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
		return dto.AuthResponse{}, problem(ErrorUnauthenticated, "用户名或密码错误")
	}
	if err != nil {
		return dto.AuthResponse{}, err
	}
	if !verifyPassword(input.Password, user.Password) {
		return dto.AuthResponse{}, problem(ErrorUnauthenticated, "用户名或密码错误")
	}
	var accessToken, refreshToken string
	err = s.Repo.WithUserTransaction(ctx, func(tx repository.UserRepository) error {
		var err error
		accessToken, refreshToken, err = s.issueSession(ctx, tx, user.ID)
		return err
	})
	return dto.AuthResponse{Token: accessToken, RefreshToken: refreshToken, UserInfo: PublicUser(user)}, err
}

// Refresh 轮换长期凭据并签发新的短期访问令牌。
func (s *UserService) Refresh(ctx context.Context, raw string) (dto.AuthResponse, error) {
	if raw == "" {
		return dto.AuthResponse{}, problem(ErrorUnauthenticated, "刷新凭据缺失或已失效")
	}
	var result dto.AuthResponse
	var rejected bool
	err := s.Repo.WithUserTransaction(ctx, func(tx repository.UserRepository) error {
		if err := tx.PruneExpiredRefreshSessions(ctx, s.now()); err != nil {
			return err
		}
		session, err := tx.RefreshSessionByHash(ctx, utils.HashRefreshToken(raw), true)
		if errors.Is(err, repository.ErrNotFound) {
			return problem(ErrorUnauthenticated, "刷新凭据缺失或已失效")
		}
		if err != nil {
			return err
		}
		now := s.now()
		if session.RevokedAt != nil || !session.ExpiresAt.After(now) || !session.FamilyExpiresAt.After(now) {
			if err := tx.RevokeRefreshFamily(ctx, session.FamilyID, now); err != nil {
				return err
			}
			rejected = true
			return nil
		}
		user, err := tx.UserByID(ctx, session.UserID, false)
		if err != nil {
			return err
		}
		accessToken, refreshToken, err := s.newSession(user.ID)
		if err != nil {
			return err
		}
		refreshHash := utils.HashRefreshToken(refreshToken)
		if err := tx.RotateRefreshSession(ctx, session.ID, now, refreshHash); errors.Is(err, repository.ErrNotFound) {
			rejected = true
			return tx.RevokeRefreshFamily(ctx, session.FamilyID, now)
		} else if err != nil {
			return err
		}
		if err := s.saveSession(ctx, tx, user.ID, session.FamilyID, refreshToken, session.FamilyExpiresAt); err != nil {
			return err
		}
		result = dto.AuthResponse{Token: accessToken, RefreshToken: refreshToken, UserInfo: PublicUser(user)}
		return nil
	})
	if err == nil && rejected {
		return dto.AuthResponse{}, problem(ErrorUnauthenticated, "刷新凭据已使用、撤销或过期，请重新登录")
	}
	return result, err
}

// Logout 撤销当前 Refresh Token。
func (s *UserService) Logout(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	return s.Repo.WithUserTransaction(ctx, func(tx repository.UserRepository) error {
		session, err := tx.RefreshSessionByHash(ctx, utils.HashRefreshToken(raw), true)
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		return tx.RevokeRefreshFamily(ctx, session.FamilyID, s.now())
	})
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
	}
	user, err := s.Repo.UserByID(ctx, user.ID, false)
	if err != nil {
		return dto.UserInfo{}, err
	}
	return PublicUser(user), nil
}

// UserInfo 查询用户资料并仅返回公开字段。
func (s *UserService) UserInfo(ctx context.Context, userID uint64) (dto.UserInfo, error) {
	user, err := s.Repo.UserByID(ctx, userID, false)
	if err != nil {
		return dto.UserInfo{}, err
	}
	return PublicUser(user), nil
}

// 验证旧密码并保存新密码哈希。
func (s *UserService) ChangePassword(ctx context.Context, user model.User, input dto.PasswordRequest) error {
	if utf8.RuneCountInString(input.New) < 6 || len(input.New) > 72 {
		return problem(ErrorInvalidArgument, "新密码至少6个字符，且不能超过72字节")
	}
	currentUser, err := s.Repo.UserByID(ctx, user.ID, false)
	if err != nil {
		return err
	}
	if !verifyPassword(input.Old, currentUser.Password) {
		return problem(ErrorInvalidArgument, "旧密码错误")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.New), bcrypt.DefaultCost)
	if err != nil {
		return problem(ErrorInternal, "密码处理失败")
	}
	return s.Repo.WithUserTransaction(ctx, func(tx repository.UserRepository) error {
		if err := tx.UpdateUser(ctx, user.ID, map[string]any{"password": string(hash), "updated_at": s.now()}); err != nil {
			return err
		}
		return tx.RevokeAllRefreshSessions(ctx, user.ID, s.now())
	})
}
