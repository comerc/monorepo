package service

import (
	"context"
	"errors"
	"strings"

	"github.com/pure-golang/monorepo/backend/profile/internal/domain"
)

type profileRepo interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	UpsertNickname(ctx context.Context, userID string, nickname string) (*domain.Profile, error)
	IsNicknameTaken(ctx context.Context, nickname string) (bool, error)
}

// Service управляет профилями пользователей.
type Service struct {
	repo profileRepo
}

// New создаёт сервис профилей.
func New(repo profileRepo) *Service {
	return &Service{repo: repo}
}

// GetByUserID возвращает профиль пользователя.
func (s *Service) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	profile, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			return &domain.Profile{UserID: userID}, nil
		}
		return nil, err
	}
	return profile, nil
}

// SetNickname сохраняет уникальный nickname пользователя.
func (s *Service) SetNickname(ctx context.Context, userID string, nickname string) (*domain.Profile, error) {
	nickname = strings.TrimSpace(nickname)
	if len([]rune(nickname)) < 2 {
		return nil, domain.ErrNicknameTooShort
	}
	return s.repo.UpsertNickname(ctx, userID, nickname)
}

// IsNicknameAvailable проверяет доступность nickname.
func (s *Service) IsNicknameAvailable(ctx context.Context, nickname string) (bool, error) {
	nickname = strings.TrimSpace(nickname)
	if len([]rune(nickname)) < 2 {
		return false, domain.ErrNicknameTooShort
	}
	taken, err := s.repo.IsNicknameTaken(ctx, nickname)
	if err != nil {
		return false, err
	}
	return !taken, nil
}
