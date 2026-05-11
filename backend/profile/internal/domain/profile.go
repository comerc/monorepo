package domain

import "errors"

// ErrProfileNotFound сообщает, что профиль не найден.
var ErrProfileNotFound = errors.New("profile not found")

// ErrNicknameTaken сообщает, что nickname уже занят.
var ErrNicknameTaken = errors.New("nickname is already taken")

// Profile описывает публичный профиль пользователя.
type Profile struct {
	UserID   string
	Nickname *string
}

// AuthUser описывает пользователя, проверенного auth-сервисом.
type AuthUser struct {
	UserID string
	Email  string
}
