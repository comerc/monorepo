package grpc

import (
	"context"
	"errors"
	"log/slog"

	alogger "github.com/pure-golang/adapters/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pure-golang/monorepo/backend/auth/internal/domain"
	authpb "github.com/pure-golang/monorepo/backend/auth/pkg/grpc"
)

type authService interface {
	ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error)
}

// Transport реализует gRPC API auth-сервиса.
type Transport struct {
	authpb.UnimplementedAuthServiceServer
	authService authService
}

// New создаёт gRPC-транспорт auth-сервиса.
func New(authService authService) *Transport {
	return &Transport{authService: authService}
}

// ValidateToken проверяет JWT-токен.
func (t *Transport) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}
	claims, err := t.authService.ValidateToken(ctx, req.GetToken())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrRevokedToken) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		alogger.FromContext(ctx).Error("Failed to validate token", slog.Any("err", err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authpb.ValidateTokenResponse{
		UserId:  claims.UserID,
		Email:   claims.Email,
		TokenId: claims.TokenID,
	}, nil
}
