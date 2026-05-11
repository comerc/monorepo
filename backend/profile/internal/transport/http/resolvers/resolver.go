package resolvers

import (
	"context"

	"github.com/pure-golang/monorepo/backend/profile/internal/domain"
)

type profileService interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	SetNickname(ctx context.Context, userID string, nickname string) (*domain.Profile, error)
}

type authService interface {
	ValidateToken(ctx context.Context, token string) (*domain.AuthUser, error)
}

// Resolver связывает GraphQL-схему с profile-сервисом.
type Resolver struct {
	profileService profileService
	authService    authService
}

// New создаёт resolver с внедрёнными зависимостями.
func New(profileService profileService, authService authService) *Resolver {
	return &Resolver{profileService: profileService, authService: authService}
}
