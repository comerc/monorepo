//go:build bdd

package bdd

import (
	"context"
	"testing"
)

// State хранит данные одного BDD-сценария.
type State struct {
	Stack *Stack

	Email      string
	Code       string
	Codes      []string
	Token      string
	OtherToken string

	CodeRequestedAfter int
	LastNickname       string
	LastError          string
	LastAccepted       bool
	LastRetryAfter     int
	LastMailCount      int
	RetryDelays        []int
}

// Reset очищает состояние сценария и переиспользует общий black-box stack.
func (s *State) Reset(t *testing.T) error {
	stack, err := GetOrCreateStack(t)
	if err != nil {
		return err
	}

	s.Stack = stack
	if err := stack.ResetData(context.Background()); err != nil {
		return err
	}
	s.Email = ""
	s.Code = ""
	s.Codes = nil
	s.Token = ""
	s.OtherToken = ""
	s.CodeRequestedAfter = 0
	s.LastNickname = ""
	s.LastError = ""
	s.LastAccepted = false
	s.LastRetryAfter = 0
	s.LastMailCount = 0
	s.RetryDelays = nil
	return nil
}
