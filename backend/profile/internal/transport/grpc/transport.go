package grpc

import (
	"context"
	"errors"
	"log/slog"

	alogger "github.com/pure-golang/adapters/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pure-golang/monorepo/backend/profile/internal/domain"
	profilepb "github.com/pure-golang/monorepo/backend/profile/pkg/grpc"
)

type profileService interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	SetNickname(ctx context.Context, userID string, nickname string) (*domain.Profile, error)
}

// Transport реализует gRPC API профилей.
type Transport struct {
	profilepb.UnimplementedProfileServiceServer
	profileService profileService
}

// New создаёт gRPC-транспорт профилей.
func New(profileService profileService) *Transport {
	return &Transport{profileService: profileService}
}

// GetProfile возвращает профиль пользователя.
func (t *Transport) GetProfile(ctx context.Context, req *profilepb.GetProfileRequest) (*profilepb.ProfileResponse, error) {
	profile, err := t.profileService.GetByUserID(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		alogger.FromContext(ctx).Error("Failed to get profile", slog.Any("err", err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return profileToProto(profile), nil
}

// SetNickname сохраняет nickname пользователя.
func (t *Transport) SetNickname(ctx context.Context, req *profilepb.SetNicknameRequest) (*profilepb.ProfileResponse, error) {
	profile, err := t.profileService.SetNickname(ctx, req.GetUserId(), req.GetNickname())
	if err != nil {
		if errors.Is(err, domain.ErrNicknameTaken) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		alogger.FromContext(ctx).Error("Failed to set nickname", slog.Any("err", err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return profileToProto(profile), nil
}

func profileToProto(profile *domain.Profile) *profilepb.ProfileResponse {
	if profile == nil {
		return nil
	}
	nickname := ""
	if profile.Nickname != nil {
		nickname = *profile.Nickname
	}
	return &profilepb.ProfileResponse{UserId: profile.UserID, Nickname: nickname}
}
