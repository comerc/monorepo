package service

import (
	"context"
	"testing"
	"time"

	"github.com/pure-golang/adapters/mail"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pure-golang/monorepo/backend/auth/internal/domain"
	"github.com/pure-golang/monorepo/backend/auth/internal/service/mocks"
)

func TestServiceRequestCodeSendsEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   10 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender).WithCodeGenerator(func() (string, error) {
		return "0427", nil
	})

	codeStore.EXPECT().
		SaveCode(ctx, "user@example.test", "0427", 10*time.Minute).
		Return(nil)
	mailSender.EXPECT().
		Send(ctx, mail.Email{
			From:    mail.Address{Address: "auth@example.test"},
			To:      []mail.Address{{Address: "user@example.test"}},
			Subject: "Your access code",
			Body:    "Your access code is 0427",
		}).
		Return(nil)

	err := svc.RequestCode(ctx, " User@Example.Test ")

	require.NoError(t, err)
}

func TestServiceLoginByCodeIssuesValidToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   10 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)
	user := &domain.User{
		ID:    "018f3fbb-1f0f-7cc9-a4d3-7332535ef0b0",
		Email: "user@example.test",
	}

	codeStore.EXPECT().
		GetCode(ctx, "user@example.test").
		Return("0427", nil)
	codeStore.EXPECT().
		DeleteCode(ctx, "user@example.test").
		Return(nil)
	userClient.EXPECT().
		GetOrCreateByEmail(ctx, "user@example.test").
		Return(user, nil)
	codeStore.EXPECT().
		IsTokenRevoked(ctx, anyTokenID(t)).
		RunAndReturn(func(_ context.Context, tokenID string) (bool, error) {
			require.NotEmpty(t, tokenID)
			return false, nil
		})

	session, err := svc.LoginByCode(ctx, "user@example.test", "0427")
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

	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   10 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)
	user := &domain.User{
		ID:    "018f3fbb-1f0f-7cc9-a4d3-7332535ef0b0",
		Email: "user@example.test",
	}

	codeStore.EXPECT().GetCode(ctx, "user@example.test").Return("0427", nil)
	codeStore.EXPECT().DeleteCode(ctx, "user@example.test").Return(nil)
	userClient.EXPECT().GetOrCreateByEmail(ctx, "user@example.test").Return(user, nil)
	codeStore.EXPECT().IsTokenRevoked(ctx, anyTokenID(t)).Return(false, nil)
	codeStore.EXPECT().
		RevokeToken(ctx, anyTokenID(t)).
		RunAndReturn(func(_ context.Context, tokenID string) error {
			require.NotEmpty(t, tokenID)
			return nil
		})

	session, err := svc.LoginByCode(ctx, "user@example.test", "0427")
	require.NoError(t, err)

	err = svc.Logout(ctx, session.Token)

	require.NoError(t, err)
}

func TestServiceLoginByCodeRejectsWrongCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	codeStore := mocks.NewCodeStore(t)
	userClient := mocks.NewUserClient(t)
	mailSender := mocks.NewMailSender(t)
	svc := New(Config{
		CodeTTL:   10 * time.Minute,
		JWTSecret: "secret",
		MailFrom:  "auth@example.test",
	}, codeStore, userClient, mailSender)

	codeStore.EXPECT().
		GetCode(ctx, "user@example.test").
		Return("0427", nil)

	session, err := svc.LoginByCode(ctx, "user@example.test", "0000")

	require.ErrorIs(t, err, domain.ErrInvalidCode)
	require.Nil(t, session)
}

func anyTokenID(t *testing.T) any {
	t.Helper()
	return mock.MatchedBy(func(tokenID string) bool {
		return tokenID != ""
	})
}
