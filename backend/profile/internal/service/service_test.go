package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pure-golang/monorepo/backend/profile/internal/domain"
	"github.com/pure-golang/monorepo/backend/profile/internal/service/mocks"
)

func TestServiceSetNicknameTrimsOuterSpaces(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := mocks.NewProfileRepo(t)
	svc := New(repo)
	repo.EXPECT().
		UpsertNickname(ctx, "user-1", "aka").
		Return(&domain.Profile{UserID: "user-1", Nickname: stringPtr("aka")}, nil)

	// Act
	profile, err := svc.SetNickname(ctx, "user-1", " aka ")

	// Assert
	require.NoError(t, err)
	require.Equal(t, "aka", *profile.Nickname)
}

func TestServiceSetNicknameRejectsTooShortNickname(t *testing.T) {
	t.Parallel()

	// Arrange
	svc := New(mocks.NewProfileRepo(t))

	// Act
	profile, err := svc.SetNickname(context.Background(), "user-1", "a")

	// Assert
	require.ErrorIs(t, err, domain.ErrNicknameTooShort)
	require.Nil(t, profile)
}

func TestServiceIsNicknameAvailable(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := mocks.NewProfileRepo(t)
	svc := New(repo)
	repo.EXPECT().IsNicknameTaken(ctx, "aka").Return(false, nil)

	// Act
	available, err := svc.IsNicknameAvailable(ctx, "aka")

	// Assert
	require.NoError(t, err)
	require.True(t, available)
}

func stringPtr(value string) *string {
	return &value
}
