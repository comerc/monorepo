package resolvers

import (
	"context"

	"github.com/pure-golang/monorepo/backend/profile/internal/domain"
)

type profileService interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	SetNickname(ctx context.Context, userID string, nickname string) (*domain.Profile, error)
}

type userService interface {
	GetUser(ctx context.Context, userID string) (*domain.User, error)
}

// Resolver связывает GraphQL-схему с profile-сервисом.
type Resolver struct {
	profileService profileService
	userService    userService
}

// New создаёт resolver с внедрёнными зависимостями.
func New(profileService profileService, userService userService) *Resolver {
	return &Resolver{profileService: profileService, userService: userService}
}
