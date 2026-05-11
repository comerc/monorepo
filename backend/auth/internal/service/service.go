package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pure-golang/adapters/mail"

	"github.com/pure-golang/monorepo/backend/auth/internal/domain"
)

type codeStore interface {
	SaveCode(ctx context.Context, email string, code string, ttl time.Duration) error
	GetCode(ctx context.Context, email string) (string, error)
	DeleteCode(ctx context.Context, email string) error
	RevokeToken(ctx context.Context, tokenID string) error
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)
}

type userClient interface {
	GetOrCreateByEmail(ctx context.Context, email string) (*domain.User, error)
}

type mailSender interface {
	Send(ctx context.Context, emails ...mail.Email) error
}

type codeGenerator = func() (string, error)

// Config описывает параметры сервиса аутентификации.
type Config struct {
	CodeTTL   time.Duration
	JWTSecret string
	MailFrom  string
}

// Service управляет email-кодами и JWT-сессиями.
type Service struct {
	codeStore     codeStore
	userClient    userClient
	mailSender    mailSender
	codeGenerator codeGenerator
	config        Config
	logger        *slog.Logger
}

// New создаёт сервис аутентификации.
func New(config Config, codeStore codeStore, userClient userClient, mailSender mailSender) *Service {
	return &Service{
		codeStore:     codeStore,
		userClient:    userClient,
		mailSender:    mailSender,
		codeGenerator: generateCode,
		config:        config,
		logger:        slog.Default().With("module", "service.auth"),
	}
}

// WithCodeGenerator подменяет генератор кода для тестов.
func (s *Service) WithCodeGenerator(generator codeGenerator) *Service {
	s.codeGenerator = generator
	return s
}

// RequestCode отправляет одноразовый код доступа на email.
func (s *Service) RequestCode(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	code, err := s.codeGenerator()
	if err != nil {
		return err
	}
	if err := s.codeStore.SaveCode(ctx, email, code, s.config.CodeTTL); err != nil {
		return err
	}
	err = s.mailSender.Send(ctx, mail.Email{
		From:    mail.Address{Address: s.config.MailFrom},
		To:      []mail.Address{{Address: email}},
		Subject: "Your access code",
		Body:    fmt.Sprintf("Your access code is %s", code),
	})
	if err != nil {
		return err
	}
	s.logger.Info("Access code sent", slog.String("email", email))
	return nil
}

// LoginByCode проверяет код и выдаёт JWT-токен.
func (s *Service) LoginByCode(ctx context.Context, email string, code string) (*domain.Session, error) {
	email = normalizeEmail(email)
	storedCode, err := s.codeStore.GetCode(ctx, email)
	if err != nil {
		return nil, err
	}
	if storedCode != code {
		return nil, domain.ErrInvalidCode
	}
	if err := s.codeStore.DeleteCode(ctx, email); err != nil {
		return nil, err
	}
	user, err := s.userClient.GetOrCreateByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	tokenUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	tokenID := tokenUUID.String()
	token, err := s.signToken(tokenID, user)
	if err != nil {
		return nil, err
	}
	return &domain.Session{
		Token:   token,
		TokenID: tokenID,
		UserID:  user.ID,
		Email:   user.Email,
	}, nil
}

// ValidateToken проверяет JWT и возвращает его claims.
func (s *Service) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	claims, err := s.parseToken(token)
	if err != nil {
		return nil, err
	}
	revoked, err := s.codeStore.IsTokenRevoked(ctx, claims.TokenID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, domain.ErrRevokedToken
	}
	return claims, nil
}

// Logout отзывает JWT-токен.
func (s *Service) Logout(ctx context.Context, token string) error {
	claims, err := s.ValidateToken(ctx, token)
	if err != nil {
		return err
	}
	return s.codeStore.RevokeToken(ctx, claims.TokenID)
}

func (s *Service) signToken(tokenID string, user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"jti":   tokenID,
		"sub":   user.ID,
		"email": user.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = "auth"
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) parseToken(rawToken string) (*domain.TokenClaims, error) {
	parsed, err := jwt.Parse(rawToken, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, domain.ErrInvalidToken
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, domain.ErrInvalidToken
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	tokenID, _ := claims["jti"].(string)
	userID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	if tokenID == "" || userID == "" || email == "" {
		return nil, domain.ErrInvalidToken
	}
	return &domain.TokenClaims{TokenID: tokenID, UserID: userID, Email: email}, nil
}

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
