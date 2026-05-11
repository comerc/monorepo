//go:build bdd

package bdd

import "testing"

// State хранит данные одного BDD-сценария.
type State struct {
	Stack *Stack

	Email string
	Code  string
	Token string

	CodeRequestedAfter int
	LastNickname       string
}

// Reset очищает состояние сценария и переиспользует общий black-box stack.
func (s *State) Reset(t *testing.T) error {
	stack, err := GetOrCreateStack(t)
	if err != nil {
		return err
	}

	s.Stack = stack
	s.Email = ""
	s.Code = ""
	s.Token = ""
	s.CodeRequestedAfter = 0
	s.LastNickname = ""
	return nil
}
