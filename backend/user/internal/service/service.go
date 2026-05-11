package service

import (
	"context"
	"strings"

	"github.com/pure-golang/monorepo/backend/user/internal/domain"
)

type userRepo interface {
	GetOrCreateByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

// Service управляет пользователями.
type Service struct {
	repo userRepo
}

// New создаёт сервис пользователей.
func New(repo userRepo) *Service {
	return &Service{repo: repo}
}

// GetOrCreateByEmail возвращает пользователя по email или создаёт его.
func (s *Service) GetOrCreateByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetOrCreateByEmail(ctx, normalizeEmail(email))
}

// GetByID возвращает пользователя по идентификатору.
func (s *Service) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
