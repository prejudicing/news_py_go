package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "news_py_go"
const audience = "news_py_go_api"

const AccessTokenTTL = 15 * time.Minute
const RefreshTokenTTL = 7 * 24 * time.Hour
const RefreshFamilyTTL = 30 * 24 * time.Hour

var ErrInvalidToken = errors.New("invalid JWT")

// IssueJWT 为用户签发短期 HS256 访问令牌。
func IssueJWT(userID uint64, secret []byte, now time.Time) (string, error) {
	if userID == 0 || len(secret) < 32 {
		return "", ErrInvalidToken
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Audience:  jwt.ClaimStrings{audience},
		Subject:   strconv.FormatUint(userID, 10),
		ID:        hex.EncodeToString(random),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// ParseJWT 验证签名、算法、签发者和有效期，再读取用户 ID。
func ParseJWT(raw string, secret []byte) (uint64, error) {
	if len(secret) < 32 {
		return 0, ErrInvalidToken
	}
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(issuer), jwt.WithAudience(audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid || claims.ID == "" || claims.IssuedAt == nil {
		return 0, ErrInvalidToken
	}
	userID, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || userID == 0 {
		return 0, ErrInvalidToken
	}
	return userID, nil
}

// NewRefreshToken 生成只用于续期的高熵不透明凭据。
func NewRefreshToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func NewRefreshFamilyID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

// HashRefreshToken 只保存刷新凭据的摘要，避免数据库泄露可直接使用的令牌。
func HashRefreshToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}
