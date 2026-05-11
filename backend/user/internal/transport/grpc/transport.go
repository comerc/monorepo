package grpc

import (
	"context"
	"errors"
	"log/slog"

	alogger "github.com/pure-golang/adapters/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pure-golang/monorepo/backend/user/internal/domain"
	userpb "github.com/pure-golang/monorepo/backend/user/pkg/grpc"
)

type userService interface {
	GetOrCreateByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

// Transport реализует gRPC API пользователей.
type Transport struct {
	userpb.UnimplementedUserServiceServer
	userService userService
}

// New создаёт gRPC-транспорт пользователей.
func New(userService userService) *Transport {
	return &Transport{userService: userService}
}

// GetOrCreateUser возвращает или создаёт пользователя по email.
func (t *Transport) GetOrCreateUser(ctx context.Context, req *userpb.GetOrCreateUserRequest) (*userpb.UserResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	user, err := t.userService.GetOrCreateByEmail(ctx, req.GetEmail())
	if err != nil {
		alogger.FromContext(ctx).Error("Failed to get or create user", slog.Any("err", err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return userToProto(user), nil
}

// GetUser возвращает пользователя по идентификатору.
func (t *Transport) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	user, err := t.userService.GetByID(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		alogger.FromContext(ctx).Error("Failed to get user", slog.Any("err", err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return userToProto(user), nil
}

func userToProto(user *domain.User) *userpb.UserResponse {
	if user == nil {
		return nil
	}
	return &userpb.UserResponse{
		UserId: user.ID,
		Email:  user.Email,
	}
}
