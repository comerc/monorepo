package domain

import "errors"

// ErrInvalidEmail сообщает, что email пустой или некорректный.
var ErrInvalidEmail = errors.New("email is invalid")

// ErrInvalidCode сообщает, что код доступа неверный или истёк.
var ErrInvalidCode = errors.New("invalid code")

// ErrInvalidToken сообщает, что JWT-токен некорректен.
var ErrInvalidToken = errors.New("invalid token")

// ErrRevokedToken сообщает, что JWT-токен отозван.
var ErrRevokedToken = errors.New("token is revoked")

// User описывает пользователя, полученного от user-сервиса.
type User struct {
	ID    string
	Email string
}

// Session описывает выданную пользовательскую сессию.
type Session struct {
	Token   string
	TokenID string
	UserID  string
	Email   string
}

// RequestCodeResult описывает результат запроса email-кода.
type RequestCodeResult struct {
	Accepted          bool
	RetryAfterSeconds int
	NextAllowedAt     string
}

// EmailCodeCooldown описывает окно повторной отправки email-кода.
type EmailCodeCooldown struct {
	Count         int   `json:"count"`
	NextAllowedAt int64 `json:"nextAllowedAt"`
	ResetAt       int64 `json:"resetAt"`
}

// TokenClaims описывает проверенные данные JWT-токена.
type TokenClaims struct {
	TokenID string
	UserID  string
	Email   string
}
