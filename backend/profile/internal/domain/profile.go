package domain

import "errors"

// ErrProfileNotFound сообщает, что профиль не найден.
var ErrProfileNotFound = errors.New("profile not found")

// ErrNicknameTaken сообщает, что nickname уже занят.
var ErrNicknameTaken = errors.New("nickname is already taken")

// ErrNicknameTooShort сообщает, что nickname короче допустимого минимума.
var ErrNicknameTooShort = errors.New("nickname is too short")

// ErrUnauthenticated сообщает, что пользователь не аутентифицирован.
var ErrUnauthenticated = errors.New("unauthenticated")

// Profile описывает публичный профиль пользователя.
type Profile struct {
	UserID   string
	Nickname *string
}

// User описывает пользователя, полученного от user-сервиса.
type User struct {
	UserID string
	Email  string
}
