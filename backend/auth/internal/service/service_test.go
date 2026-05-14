package service

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/pure-golang/adapters/mail"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pure-golang/monorepo/backend/auth/internal/domain"
	"github.com/pure-golang/monorepo/backend/auth/internal/service/mocks"
)

func TestGenerateCodeReturnsFiveDigits(t *testing.T) {
	t.Parallel()

	// Arrange

	// Act
	code, err := generateCode()

	// Assert
	require.NoError(t, err)
	require.Regexp(t, regexp.MustCompile(`^\d{5}$`), code)
}

func TestServiceRequestCodeSendsEmail(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender).WithCodeGenerator(func() (string, error) {
		return "00427", nil
	})

	codeStore.EXPECT().
		GetEmailCodeCooldown(ctx, "user@example.test").
		Return(nil, domain.ErrInvalidCode)
	codeStore.EXPECT().
		SaveEmailCodeCooldown(ctx, "user@example.test", mock.Anything, mock.Anything).
		Return(nil)
	codeStore.EXPECT().
		SaveCode(ctx, "user@example.test", "00427", 5*time.Minute).
		Return(nil)
	mailSender.EXPECT().
		Send(ctx, mail.Email{
			From:    mail.Address{Address: "auth@example.test"},
			To:      []mail.Address{{Address: "user@example.test"}},
			Subject: "Your access code",
			Body:    "Your access code is 00427",
		}).
		Return(nil)

	// Act
	result, err := svc.RequestCode(ctx, " User@Example.Test ")

	// Assert
	require.NoError(t, err)
	require.True(t, result.Accepted)
}

func TestServiceRequestCodeReturnsDeliveryErrorWithoutCooldown(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender).WithCodeGenerator(func() (string, error) {
		return "00427", nil
	})
	deliveryErr := errors.New("delivery failed")

	codeStore.EXPECT().
		GetEmailCodeCooldown(ctx, "user@example.test").
		Return(nil, domain.ErrInvalidCode)
	codeStore.EXPECT().
		SaveEmailCodeCooldown(ctx, "user@example.test", mock.Anything, mock.Anything).
		Return(nil)
	codeStore.EXPECT().
		SaveCode(ctx, "user@example.test", "00427", 5*time.Minute).
		Return(nil)
	mailSender.EXPECT().
		Send(ctx, mail.Email{
			From:    mail.Address{Address: "auth@example.test"},
			To:      []mail.Address{{Address: "user@example.test"}},
			Subject: "Your access code",
			Body:    "Your access code is 00427",
		}).
		Return(deliveryErr)
	codeStore.EXPECT().
		DeleteCode(ctx, "user@example.test", "00427").
		Return(nil)
	codeStore.EXPECT().
		DeleteEmailCodeCooldown(ctx, "user@example.test").
		Return(nil)

	// Act
	result, err := svc.RequestCode(ctx, "user@example.test")

	// Assert
	require.ErrorIs(t, err, deliveryErr)
	require.Nil(t, result)
}

func TestServiceLoginByCodeIssuesValidToken(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)
	user := &domain.User{
		ID:    "018f3fbb-1f0f-7cc9-a4d3-7332535ef0b0",
		Email: "user@example.test",
	}

	codeStore.EXPECT().
		GetCodes(ctx, "user@example.test").
		Return([]string{"00427"}, nil)
	codeStore.EXPECT().
		DeleteCodes(ctx, "user@example.test").
		Return(nil)
	codeStore.EXPECT().
		DeleteEmailCodeCooldown(ctx, "user@example.test").
		Return(nil)
	userClient.EXPECT().
		GetOrCreateByEmail(ctx, "user@example.test").
		Return(user, nil)
	codeStore.EXPECT().
		SaveUserToken(ctx, user.ID, anyTokenID(t)).
		Return(nil)
	codeStore.EXPECT().
		IsTokenRevoked(ctx, anyTokenID(t)).
		RunAndReturn(func(_ context.Context, tokenID string) (bool, error) {
			require.NotEmpty(t, tokenID)
			return false, nil
		})

	// Act
	session, err := svc.LoginByCode(ctx, "user@example.test", "00427")

	// Assert
	require.NoError(t, err)
	require.Equal(t, user.ID, session.UserID)
	require.Equal(t, user.Email, session.Email)
	require.NotEmpty(t, session.Token)

	claims, err := svc.ValidateToken(ctx, session.Token)
	require.NoError(t, err)
	require.Equal(t, session.TokenID, claims.TokenID)
	require.Equal(t, user.ID, claims.UserID)
	require.Equal(t, user.Email, claims.Email)
}

func TestServiceLogoutRevokesToken(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)
	user := &domain.User{
		ID:    "018f3fbb-1f0f-7cc9-a4d3-7332535ef0b0",
		Email: "user@example.test",
	}

	codeStore.EXPECT().GetCodes(ctx, "user@example.test").Return([]string{"00427"}, nil)
	codeStore.EXPECT().DeleteCodes(ctx, "user@example.test").Return(nil)
	codeStore.EXPECT().DeleteEmailCodeCooldown(ctx, "user@example.test").Return(nil)
	userClient.EXPECT().GetOrCreateByEmail(ctx, "user@example.test").Return(user, nil)
	codeStore.EXPECT().SaveUserToken(ctx, user.ID, anyTokenID(t)).Return(nil)
	codeStore.EXPECT().IsTokenRevoked(ctx, anyTokenID(t)).Return(false, nil)
	codeStore.EXPECT().
		RevokeToken(ctx, anyTokenID(t)).
		RunAndReturn(func(_ context.Context, tokenID string) error {
			require.NotEmpty(t, tokenID)
			return nil
		})

	session, err := svc.LoginByCode(ctx, "user@example.test", "00427")
	require.NoError(t, err)

	// Act
	err = svc.Logout(ctx, session.Token)

	// Assert
	require.NoError(t, err)
}

func TestServiceLogoutEverywhereRevokesCurrentToken(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)
	user := &domain.User{
		ID:    "018f3fbb-1f0f-7cc9-a4d3-7332535ef0b0",
		Email: "user@example.test",
	}

	codeStore.EXPECT().GetCodes(ctx, "user@example.test").Return([]string{"00427"}, nil)
	codeStore.EXPECT().DeleteCodes(ctx, "user@example.test").Return(nil)
	codeStore.EXPECT().DeleteEmailCodeCooldown(ctx, "user@example.test").Return(nil)
	userClient.EXPECT().GetOrCreateByEmail(ctx, "user@example.test").Return(user, nil)
	codeStore.EXPECT().SaveUserToken(ctx, user.ID, anyTokenID(t)).Return(nil)
	codeStore.EXPECT().IsTokenRevoked(ctx, anyTokenID(t)).Return(false, nil)
	codeStore.EXPECT().
		RevokeToken(ctx, anyTokenID(t)).
		RunAndReturn(func(_ context.Context, tokenID string) error {
			require.NotEmpty(t, tokenID)
			return nil
		})
	codeStore.EXPECT().RevokeUserTokens(ctx, user.ID).Return(nil)

	session, err := svc.LoginByCode(ctx, "user@example.test", "00427")
	require.NoError(t, err)

	// Act
	err = svc.LogoutEverywhere(ctx, session.Token)

	// Assert
	require.NoError(t, err)
}

func TestServiceLoginByCodeRejectsWrongCode(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)

	codeStore.EXPECT().
		GetCodes(ctx, "user@example.test").
		Return([]string{"00427"}, nil)

	// Act
	session, err := svc.LoginByCode(ctx, "user@example.test", "00000")

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidCode)
	require.Nil(t, session)
}

func TestServiceLoginByCodeKeepsCodesWhenSessionCannotBeSaved(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)
	user := &domain.User{
		ID:    "018f3fbb-1f0f-7cc9-a4d3-7332535ef0b0",
		Email: "user@example.test",
	}
	saveErr := errors.New("save token failed")

	codeStore.EXPECT().
		GetCodes(ctx, "user@example.test").
		Return([]string{"00427"}, nil)
	userClient.EXPECT().
		GetOrCreateByEmail(ctx, "user@example.test").
		Return(user, nil)
	codeStore.EXPECT().
		SaveUserToken(ctx, user.ID, anyTokenID(t)).
		Return(saveErr)

	// Act
	session, err := svc.LoginByCode(ctx, "user@example.test", "00427")

	// Assert
	require.ErrorIs(t, err, saveErr)
	require.Nil(t, session)
}

func TestServiceRequestCodeRespectsCooldown(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	now := time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender).WithClock(func() time.Time {
		return now
	})

	codeStore.EXPECT().
		GetEmailCodeCooldown(ctx, "user@example.test").
		Return(&domain.EmailCodeCooldown{
			Count:         1,
			NextAllowedAt: now.Add(30 * time.Second).Unix(),
			ResetAt:       now.Add(24 * time.Hour).Unix(),
		}, nil)

	// Act
	result, err := svc.RequestCode(ctx, "user@example.test")

	// Assert
	require.NoError(t, err)
	require.False(t, result.Accepted)
	require.Equal(t, 30, result.RetryAfterSeconds)
}

func TestServiceLoginByCodeConsumesAllCodes(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)
	user := &domain.User{
		ID:    "018f3fbb-1f0f-7cc9-a4d3-7332535ef0b0",
		Email: "user@example.test",
	}

	codeStore.EXPECT().GetCodes(ctx, "user@example.test").Return([]string{"11111", "22222"}, nil)
	codeStore.EXPECT().DeleteCodes(ctx, "user@example.test").Return(nil)
	codeStore.EXPECT().DeleteEmailCodeCooldown(ctx, "user@example.test").Return(nil)
	userClient.EXPECT().GetOrCreateByEmail(ctx, "user@example.test").Return(user, nil)
	codeStore.EXPECT().SaveUserToken(ctx, user.ID, anyTokenID(t)).Return(nil)

	// Act
	session, err := svc.LoginByCode(ctx, "user@example.test", "11111")

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, session.Token)
}

func TestServiceLoginByCodeResetsEmailCodeCooldown(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)
	user := &domain.User{
		ID:    "018f3fbb-1f0f-7cc9-a4d3-7332535ef0b0",
		Email: "user@example.test",
	}

	codeStore.EXPECT().GetCodes(ctx, "user@example.test").Return([]string{"00427"}, nil)
	codeStore.EXPECT().DeleteCodes(ctx, "user@example.test").Return(nil)
	codeStore.EXPECT().DeleteEmailCodeCooldown(ctx, "user@example.test").Return(nil)
	userClient.EXPECT().GetOrCreateByEmail(ctx, "user@example.test").Return(user, nil)
	codeStore.EXPECT().SaveUserToken(ctx, user.ID, anyTokenID(t)).Return(nil)

	// Act
	session, err := svc.LoginByCode(ctx, "user@example.test", "00427")

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, session.Token)
}

func TestServiceRequestCodeRejectsDisplayNameEmail(t *testing.T) {
	t.Parallel()

	// Arrange
	svc := New(Config{
		CodeTTL:   5 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, mocks.NewCodeStore(t), mocks.NewUserClient(t), mocks.NewMailSender(t))

	// Act
	result, err := svc.RequestCode(context.Background(), "Alice <user@example.test>")

	// Assert
	require.ErrorIs(t, err, domain.ErrInvalidEmail)
	require.Nil(t, result)
}

func anyTokenID(t *testing.T) any {
	t.Helper()
	return mock.MatchedBy(func(tokenID string) bool {
		return tokenID != ""
	})
}
