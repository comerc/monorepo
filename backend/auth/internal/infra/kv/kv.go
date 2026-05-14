package kv

import (
	"context"
	"encoding/json"
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
	if err := s.kv.Set(ctx, codeValueKey(email, code), code, ttl); err != nil {
		return err
	}
	if err := s.kv.SAdd(ctx, codesKey(email), code); err != nil {
		if cleanupErr := s.kv.Delete(ctx, codeValueKey(email, code)); cleanupErr != nil && !errors.Is(cleanupErr, aredis.ErrKeyNotFound) {
			return errors.Join(err, cleanupErr)
		}
		return err
	}
	if err := s.kv.Expire(ctx, codesKey(email), ttl); err != nil {
		if cleanupErr := s.DeleteCode(ctx, email, code); cleanupErr != nil {
			return errors.Join(err, cleanupErr)
		}
		return err
	}
	return nil
}

// GetCodes возвращает активные коды доступа.
func (s *Store) GetCodes(ctx context.Context, email string) ([]string, error) {
	codes, err := s.kv.SMembers(ctx, codesKey(email))
	if err != nil {
		return nil, err
	}
	activeCodes := make([]string, 0, len(codes))
	staleCodes := make([]any, 0)
	for _, code := range codes {
		exists, err := s.kv.Exists(ctx, codeValueKey(email, code))
		if err != nil {
			return nil, err
		}
		if exists == 0 {
			staleCodes = append(staleCodes, code)
			continue
		}
		activeCodes = append(activeCodes, code)
	}
	if len(staleCodes) > 0 {
		if err := s.kv.SRem(ctx, codesKey(email), staleCodes...); err != nil {
			return nil, err
		}
	}
	if len(activeCodes) == 0 {
		return nil, domain.ErrInvalidCode
	}
	return activeCodes, nil
}

// DeleteCode удаляет один активный код доступа.
func (s *Store) DeleteCode(ctx context.Context, email string, code string) error {
	if err := s.kv.Delete(ctx, codeValueKey(email, code)); err != nil && !errors.Is(err, aredis.ErrKeyNotFound) {
		return err
	}
	return s.kv.SRem(ctx, codesKey(email), code)
}

// DeleteCodes удаляет все активные коды доступа.
func (s *Store) DeleteCodes(ctx context.Context, email string) error {
	codes, err := s.kv.SMembers(ctx, codesKey(email))
	if err != nil {
		return err
	}
	keys := []string{codesKey(email)}
	for _, code := range codes {
		keys = append(keys, codeValueKey(email, code))
	}
	return s.kv.Delete(ctx, keys...)
}

// GetEmailCodeCooldown возвращает текущее окно повторной отправки кода.
func (s *Store) GetEmailCodeCooldown(ctx context.Context, email string) (*domain.EmailCodeCooldown, error) {
	raw, err := s.kv.Get(ctx, cooldownKey(email))
	if err != nil {
		if errors.Is(err, aredis.ErrKeyNotFound) {
			return nil, domain.ErrInvalidCode
		}
		return nil, err
	}
	var cooldown domain.EmailCodeCooldown
	if err := json.Unmarshal([]byte(raw), &cooldown); err != nil {
		return nil, err
	}
	return &cooldown, nil
}

// SaveEmailCodeCooldown сохраняет окно повторной отправки кода.
func (s *Store) SaveEmailCodeCooldown(ctx context.Context, email string, cooldown domain.EmailCodeCooldown, ttl time.Duration) error {
	raw, err := json.Marshal(cooldown)
	if err != nil {
		return err
	}
	return s.kv.Set(ctx, cooldownKey(email), raw, ttl)
}

// DeleteEmailCodeCooldown сбрасывает окно повторной отправки кода.
func (s *Store) DeleteEmailCodeCooldown(ctx context.Context, email string) error {
	if err := s.kv.Delete(ctx, cooldownKey(email)); err != nil && !errors.Is(err, aredis.ErrKeyNotFound) {
		return err
	}
	return nil
}

// SaveUserToken связывает токен с пользователем для logout everywhere.
func (s *Store) SaveUserToken(ctx context.Context, userID string, tokenID string) error {
	return s.kv.SAdd(ctx, userTokensKey(userID), tokenID)
}

// RevokeToken помечает JWT как отозванный.
func (s *Store) RevokeToken(ctx context.Context, tokenID string) error {
	return s.kv.Set(ctx, revokedKey(tokenID), "1", 0)
}

// RevokeUserTokens помечает все известные JWT пользователя как отозванные.
func (s *Store) RevokeUserTokens(ctx context.Context, userID string) error {
	tokenIDs, err := s.kv.SMembers(ctx, userTokensKey(userID))
	if err != nil {
		return err
	}
	for _, tokenID := range tokenIDs {
		if err := s.RevokeToken(ctx, tokenID); err != nil {
			return err
		}
	}
	return nil
}

// IsTokenRevoked проверяет, был ли JWT отозван.
func (s *Store) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	exists, err := s.kv.Exists(ctx, revokedKey(tokenID))
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func codesKey(email string) string {
	return fmt.Sprintf("auth:codes:%s", email)
}

func codeValueKey(email string, code string) string {
	return fmt.Sprintf("auth:code:%s:%s", email, code)
}

func cooldownKey(email string) string {
	return fmt.Sprintf("auth:code-cooldown:%s", email)
}

func revokedKey(tokenID string) string {
	return fmt.Sprintf("auth:revoked:%s", tokenID)
}

func userTokensKey(userID string) string {
	return fmt.Sprintf("auth:user-tokens:%s", userID)
}
