package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-only-signing-secret-with-at-least-32-bytes"

// TestJWTSignAndValidate 验证签发令牌可被相同服务策略正常解析。
func TestJWTSignAndValidate(t *testing.T) {
	issued := time.Now().UTC()
	raw, err := IssueJWT(42, []byte(testSecret), issued)
	if err != nil {
		t.Fatal(err)
	}
	userID, err := ParseJWT(raw, []byte(testSecret))
	if err != nil || userID != 42 {
		t.Fatalf("signed token failed validation: user=%d err=%v", userID, err)
	}
	second, err := IssueJWT(42, []byte(testSecret), issued)
	if err != nil || raw == second {
		t.Fatal("tokens issued within one second must be distinct")
	}
}

// TestJWTRejectsTamperingAndExpiry 确认签名篡改及过期令牌均被拒绝。
func TestJWTRejectsTamperingAndExpiry(t *testing.T) {
	raw, err := IssueJWT(42, []byte(testSecret), time.Now().Add(-8*24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseJWT(raw, []byte(testSecret)); err == nil {
		t.Fatal("expired token accepted")
	}
	raw, err = IssueJWT(42, []byte(testSecret), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(raw, ".")
	parts[1] = parts[1] + "A"
	if _, err := ParseJWT(strings.Join(parts, "."), []byte(testSecret)); err == nil {
		t.Fatal("tampered token accepted")
	}
	if _, err := ParseJWT(raw, []byte("another-signing-secret-with-at-least-32-bytes")); err == nil {
		t.Fatal("token with wrong key accepted")
	}
	foreign := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.RegisteredClaims{
		Issuer: issuer, Subject: "42", ID: "test", IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	hs384, err := foreign.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseJWT(hs384, []byte(testSecret)); err == nil {
		t.Fatal("unexpected signing algorithm accepted")
	}
}
