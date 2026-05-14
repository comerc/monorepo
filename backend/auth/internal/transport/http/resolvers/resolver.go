package resolvers

import (
	"context"

	"github.com/pure-golang/monorepo/backend/auth/internal/domain"
)

type authService interface {
	RequestCode(ctx context.Context, email string) (*domain.RequestCodeResult, error)
	LoginByCode(ctx context.Context, email string, code string) (*domain.Session, error)
	Logout(ctx context.Context, token string) error
	LogoutEverywhere(ctx context.Context, token string) error
}

// Resolver связывает GraphQL-схему с auth-сервисом.
type Resolver struct {
	authService authService
}

// New создаёт resolver с внедрёнными зависимостями.
func New(authService authService) *Resolver {
	return &Resolver{authService: authService}
}
