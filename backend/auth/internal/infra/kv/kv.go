package kv

import (
	"context"
	"errors"
	"fmt"
	"time"

	akv "github.com/pure-golang/adapters/kv"
	aredis "github.com/pure-golang/adapters/kv/redis"

	"github.com/pure-golang/monorepo/backend/auth/internal/domain"
)

// Store хранит одноразовые email-коды и отозванные токены.
type Store struct {
	kv akv.Store
}

// New создаёт хранилище auth-данных поверх KV-адаптера.
func New(kv akv.Store) *Store {
	return &Store{kv: kv}
}

// SaveCode сохраняет код доступа с TTL.
func (s *Store) SaveCode(ctx context.Context, email string, code string, ttl time.Duration) error {
	return s.kv.Set(ctx, codeKey(email), code, ttl)
}

// GetCode возвращает сохранённый код доступа.
func (s *Store) GetCode(ctx context.Context, email string) (string, error) {
	code, err := s.kv.Get(ctx, codeKey(email))
	if err != nil {
		if errors.Is(err, aredis.ErrKeyNotFound) {
			return "", domain.ErrInvalidCode
		}
		return "", err
	}
	return code, nil
}

// DeleteCode удаляет использованный код доступа.
func (s *Store) DeleteCode(ctx context.Context, email string) error {
	return s.kv.Delete(ctx, codeKey(email))
}

// RevokeToken помечает JWT как отозванный.
func (s *Store) RevokeToken(ctx context.Context, tokenID string) error {
	return s.kv.Set(ctx, revokedKey(tokenID), "1", 0)
}

// IsTokenRevoked проверяет, был ли JWT отозван.
func (s *Store) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	exists, err := s.kv.Exists(ctx, revokedKey(tokenID))
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func codeKey(email string) string {
	return fmt.Sprintf("auth:code:%s", email)
}

func revokedKey(tokenID string) string {
	return fmt.Sprintf("auth:revoked:%s", tokenID)
}
