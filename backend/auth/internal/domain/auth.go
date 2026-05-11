package domain

import "errors"

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

// TokenClaims описывает проверенные данные JWT-токена.
type TokenClaims struct {
	TokenID string
	UserID  string
	Email   string
}
