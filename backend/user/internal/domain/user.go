package domain

import "errors"

// ErrUserNotFound сообщает, что пользователь не найден.
var ErrUserNotFound = errors.New("user not found")

// User описывает пользователя, которым владеет user-сервис.
type User struct {
	ID    string
	Email string
}
