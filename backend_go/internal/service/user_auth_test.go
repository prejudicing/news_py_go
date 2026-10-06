package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prejudicing/news_py_go/backend_go/internal/model"
	"github.com/prejudicing/news_py_go/backend_go/internal/repository"
	"github.com/prejudicing/news_py_go/backend_go/internal/utils"
)

// refreshMemoryStore 为认证单元测试提供可观察令牌轮换行为的内存仓储。
type refreshMemoryStore struct {
	repository.UserRepository
	user     model.User
	sessions map[string]model.RefreshSession
	nextID   uint64
}

func (r *refreshMemoryStore) WithUserTransaction(_ context.Context, fn func(repository.UserRepository) error) error {
	return fn(r)
}

func (r *refreshMemoryStore) UserByID(_ context.Context, id uint64, _ bool) (model.User, error) {
	if id != r.user.ID {
		return model.User{}, repository.ErrNotFound
	}
	return r.user, nil
}

func (r *refreshMemoryStore) RefreshSessionByHash(_ context.Context, hash string, _ bool) (model.RefreshSession, error) {
	session, ok := r.sessions[hash]
	if !ok {
		return model.RefreshSession{}, repository.ErrNotFound
	}
	return session, nil
}

func (r *refreshMemoryStore) CreateRefreshSession(_ context.Context, session *model.RefreshSession) error {
	r.nextID++
	session.ID = r.nextID
	r.sessions[session.TokenHash] = *session
	return nil
}

func (r *refreshMemoryStore) RotateRefreshSession(_ context.Context, id uint64, now time.Time, replacement string) error {
	for hash, session := range r.sessions {
		if session.ID != id {
			continue
		}
		if session.RevokedAt != nil || !session.ExpiresAt.After(now) || !session.FamilyExpiresAt.After(now) {
			return repository.ErrNotFound
		}
		session.RevokedAt = &now
		session.ReplacedByHash = &replacement
		r.sessions[hash] = session
		return nil
	}
	return repository.ErrNotFound
}

func (r *refreshMemoryStore) RevokeRefreshFamily(_ context.Context, family string, now time.Time) error {
	for hash, session := range r.sessions {
		if session.FamilyID == family && session.RevokedAt == nil {
			session.RevokedAt = &now
			r.sessions[hash] = session
		}
	}
	return nil
}

func (r *refreshMemoryStore) RevokeAllRefreshSessions(_ context.Context, userID uint64, now time.Time) error {
	for hash, session := range r.sessions {
		if session.UserID == userID && session.RevokedAt == nil {
			session.RevokedAt = &now
			r.sessions[hash] = session
		}
	}
	return nil
}

func (r *refreshMemoryStore) PruneExpiredRefreshSessions(_ context.Context, now time.Time) error {
	for hash, session := range r.sessions {
		if !session.FamilyExpiresAt.After(now) {
			delete(r.sessions, hash)
		}
	}
	return nil
}

// TestRefreshRotationDetectsReplayAndRevokesOnlyItsFamily 覆盖轮换、重放撤销和设备隔离。
func TestRefreshRotationDetectsReplayAndRevokesOnlyItsFamily(t *testing.T) {
	now := time.Now()
	oldRaw := "original-refresh-secret"
	oldHash := utils.HashRefreshToken(oldRaw)
	repo := &refreshMemoryStore{
		user:     model.User{ID: 42, Username: "session-user"},
		sessions: map[string]model.RefreshSession{},
		nextID:   11,
	}
	repo.sessions[oldHash] = model.RefreshSession{
		ID: 10, UserID: 42, FamilyID: "family-a", TokenHash: oldHash,
		ExpiresAt: now.Add(utils.RefreshTokenTTL), FamilyExpiresAt: now.Add(utils.RefreshFamilyTTL), CreatedAt: now,
	}
	otherHash := utils.HashRefreshToken("other-device-refresh")
	repo.sessions[otherHash] = model.RefreshSession{
		ID: 11, UserID: 42, FamilyID: "family-b", TokenHash: otherHash,
		ExpiresAt: now.Add(utils.RefreshTokenTTL), FamilyExpiresAt: now.Add(utils.RefreshFamilyTTL), CreatedAt: now,
	}
	service := &UserService{Repo: repo, JWTSecret: []byte("unit-test-signing-secret-at-least-32-bytes")}

	rotated, err := service.Refresh(context.Background(), oldRaw)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.RefreshToken == "" || rotated.RefreshToken == oldRaw {
		t.Fatal("refresh token was not rotated")
	}
	if userID, err := utils.ParseJWT(rotated.Token, service.JWTSecret); err != nil || userID != repo.user.ID {
		t.Fatalf("new access token is invalid: user=%d err=%v", userID, err)
	}
	rotatedHash := utils.HashRefreshToken(rotated.RefreshToken)
	oldSession := repo.sessions[oldHash]
	newSession := repo.sessions[rotatedHash]
	if oldSession.RevokedAt == nil || oldSession.ReplacedByHash == nil || *oldSession.ReplacedByHash != rotatedHash {
		t.Fatal("rotation history was not retained")
	}
	if newSession.FamilyID != oldSession.FamilyID || !newSession.FamilyExpiresAt.Equal(oldSession.FamilyExpiresAt) {
		t.Fatal("rotation changed the session family or its absolute lifetime")
	}

	_, err = service.Refresh(context.Background(), oldRaw)
	var business *Error
	if !errors.As(err, &business) || business.Status != 401 {
		t.Fatalf("replayed refresh token should be rejected with 401: %v", err)
	}
	if repo.sessions[rotatedHash].RevokedAt == nil {
		t.Fatal("replay did not revoke the active token in its family")
	}
	if repo.sessions[otherHash].RevokedAt != nil {
		t.Fatal("replay revoked a different device session")
	}
	if _, err := service.Refresh(context.Background(), rotated.RefreshToken); !errors.As(err, &business) || business.Status != 401 {
		t.Fatalf("revoked family token should not refresh: %v", err)
	}
}
