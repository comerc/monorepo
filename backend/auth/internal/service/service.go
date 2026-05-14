package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	netmail "net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	amail "github.com/pure-golang/adapters/mail"

	"github.com/pure-golang/monorepo/backend/auth/internal/domain"
)

type codeStore interface {
	SaveCode(ctx context.Context, email string, code string, ttl time.Duration) error
	GetCodes(ctx context.Context, email string) ([]string, error)
	DeleteCode(ctx context.Context, email string, code string) error
	DeleteCodes(ctx context.Context, email string) error
	GetEmailCodeCooldown(ctx context.Context, email string) (*domain.EmailCodeCooldown, error)
	SaveEmailCodeCooldown(ctx context.Context, email string, cooldown domain.EmailCodeCooldown, ttl time.Duration) error
	DeleteEmailCodeCooldown(ctx context.Context, email string) error
	SaveUserToken(ctx context.Context, userID string, tokenID string) error
	RevokeToken(ctx context.Context, tokenID string) error
	RevokeUserTokens(ctx context.Context, userID string) error
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)
}

type userClient interface {
	GetOrCreateByEmail(ctx context.Context, email string) (*domain.User, error)
}

type mailSender interface {
	Send(ctx context.Context, emails ...amail.Email) error
}

type codeGenerator = func() (string, error)
type clock = func() time.Time

var cooldownSteps = []time.Duration{30 * time.Second, 60 * time.Second, 120 * time.Second, 24 * time.Hour}

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
	clock         clock
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
		clock:         time.Now,
		config:        config,
		logger:        slog.Default().With("module", "service.auth"),
	}
}

// WithCodeGenerator подменяет генератор кода для тестов.
func (s *Service) WithCodeGenerator(generator codeGenerator) *Service {
	s.codeGenerator = generator
	return s
}

// WithClock подменяет источник времени для тестов.
func (s *Service) WithClock(clock clock) *Service {
	s.clock = clock
	return s
}

// RequestCode отправляет одноразовый код доступа на email.
func (s *Service) RequestCode(ctx context.Context, email string) (*domain.RequestCodeResult, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	now := s.clock().UTC()
	cooldown, err := s.codeStore.GetEmailCodeCooldown(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrInvalidCode) {
		return nil, err
	}
	if cooldown != nil && now.Unix() < cooldown.NextAllowedAt {
		nextAllowedAt := time.Unix(cooldown.NextAllowedAt, 0).UTC()
		return &domain.RequestCodeResult{
			Accepted:          false,
			RetryAfterSeconds: secondsUntil(nextAllowedAt, now),
			NextAllowedAt:     nextAllowedAt.Format(time.RFC3339),
		}, nil
	}

	code, err := s.codeGenerator()
	if err != nil {
		return nil, err
	}
	nextCooldown := nextEmailCodeCooldown(cooldown, now)
	if err := s.codeStore.SaveEmailCodeCooldown(ctx, email, nextCooldown, time.Unix(nextCooldown.ResetAt, 0).Sub(now)); err != nil {
		return nil, err
	}
	if err := s.codeStore.SaveCode(ctx, email, code, s.config.CodeTTL); err != nil {
		if cleanupErr := s.codeStore.DeleteEmailCodeCooldown(ctx, email); cleanupErr != nil {
			return nil, errors.Join(err, cleanupErr)
		}
		return nil, err
	}
	err = s.mailSender.Send(ctx, amail.Email{
		From:    amail.Address{Address: s.config.MailFrom},
		To:      []amail.Address{{Address: email}},
		Subject: "Your access code",
		Body:    fmt.Sprintf("Your access code is %s", code),
	})
	if err != nil {
		s.logger.Warn("Access code delivery failed", slog.String("email", email), slog.Any("err", err))
		cleanupErr := errors.Join(
			s.codeStore.DeleteCode(ctx, email, code),
			s.codeStore.DeleteEmailCodeCooldown(ctx, email),
		)
		if cleanupErr != nil {
			return nil, errors.Join(err, cleanupErr)
		}
		return nil, err
	}
	nextAllowedAt := time.Unix(nextCooldown.NextAllowedAt, 0).UTC()
	s.logger.Info("Access code accepted", slog.String("email", email))
	return &domain.RequestCodeResult{
		Accepted:          true,
		RetryAfterSeconds: secondsUntil(nextAllowedAt, now),
		NextAllowedAt:     nextAllowedAt.Format(time.RFC3339),
	}, nil
}

// LoginByCode проверяет код и выдаёт JWT-токен.
func (s *Service) LoginByCode(ctx context.Context, email string, code string) (*domain.Session, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	storedCodes, err := s.codeStore.GetCodes(ctx, email)
	if err != nil {
		return nil, err
	}
	if !containsString(storedCodes, strings.TrimSpace(code)) {
		return nil, domain.ErrInvalidCode
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
	if err := s.codeStore.SaveUserToken(ctx, user.ID, tokenID); err != nil {
		return nil, err
	}
	if err := s.codeStore.DeleteCodes(ctx, email); err != nil {
		return nil, err
	}
	if err := s.codeStore.DeleteEmailCodeCooldown(ctx, email); err != nil {
		s.logger.Warn("Email code cooldown reset failed", slog.String("email", email), slog.Any("err", err))
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

// LogoutEverywhere отзывает все активные JWT-токены пользователя.
func (s *Service) LogoutEverywhere(ctx context.Context, token string) error {
	claims, err := s.ValidateToken(ctx, token)
	if err != nil {
		return err
	}
	if err := s.codeStore.RevokeToken(ctx, claims.TokenID); err != nil {
		return err
	}
	return s.codeStore.RevokeUserTokens(ctx, claims.UserID)
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
	n, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%05d", n.Int64()), nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", errors.New("email is required")
	}
	address, err := netmail.ParseAddress(email)
	if err != nil || address.Address != email || !strings.Contains(email, "@") {
		return "", domain.ErrInvalidEmail
	}
	return email, nil
}

func nextEmailCodeCooldown(current *domain.EmailCodeCooldown, now time.Time) domain.EmailCodeCooldown {
	count := 0
	resetAt := now.Add(24 * time.Hour).Unix()
	if current != nil && now.Unix() < current.ResetAt {
		count = current.Count
		resetAt = current.ResetAt
	}
	stepIndex := count
	if stepIndex >= len(cooldownSteps) {
		stepIndex = len(cooldownSteps) - 1
	}
	nextAllowedAt := now.Add(cooldownSteps[stepIndex])
	if stepIndex == len(cooldownSteps)-1 {
		resetAt = nextAllowedAt.Unix()
	}
	return domain.EmailCodeCooldown{
		Count:         count + 1,
		NextAllowedAt: nextAllowedAt.Unix(),
		ResetAt:       resetAt,
	}
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func secondsUntil(nextAllowedAt time.Time, now time.Time) int {
	seconds := nextAllowedAt.Sub(now).Seconds()
	if seconds <= 0 {
		return 0
	}
	return int(math.Ceil(seconds))
}
