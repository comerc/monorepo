//go:build bdd

package bdd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cucumber/godog"
)

var fourDigitCode = regexp.MustCompile(`^\d{4}$`)

// RegisterIdentitySteps регистрирует шаги эпика identity.
func RegisterIdentitySteps(ctx *godog.ScenarioContext, s *State) {
	ctx.When(`^пользователь запрашивает код доступа для email "([^"]*)"$`, func(email string) error {
		s.Email = email
		after, err := s.Stack.RequestEmailCode(context.Background(), email)
		if err != nil {
			return err
		}
		s.CodeRequestedAfter = after
		return nil
	})

	ctx.Then(`^на email "([^"]*)" отправлен код доступа из 4 цифр$`, func(email string) error {
		code, err := s.Stack.WaitEmailCode(context.Background(), email, s.CodeRequestedAfter)
		if err != nil {
			return err
		}
		if !fourDigitCode.MatchString(code) {
			return fmt.Errorf("email code must contain 4 digits, got %q", code)
		}
		s.Email = email
		s.Code = code
		return nil
	})

	ctx.Given(`^пользователь получил код доступа для email "([^"]*)"$`, func(email string) error {
		code, err := s.Stack.RequestEmailCodeAndWait(context.Background(), email)
		if err != nil {
			return err
		}
		s.Email = email
		s.Code = code
		return nil
	})

	ctx.When(`^пользователь вводит полученный код доступа для email "([^"]*)"$`, func(email string) error {
		token, err := s.Stack.LoginWithEmailCode(context.Background(), email, s.Code)
		if err != nil {
			return err
		}
		s.Email = email
		s.Token = token
		return nil
	})

	ctx.Then(`^пользователь получает JWT-токен$`, func() error {
		return s.Stack.ExpectJWT(context.Background(), s.Token, s.Email)
	})

	ctx.Given(`^пользователь вошёл по email "([^"]*)"$`, func(email string) error {
		token, err := s.Stack.LoginByEmail(context.Background(), email)
		if err != nil {
			return err
		}
		s.Email = email
		s.Token = token
		return nil
	})

	ctx.When(`^пользователь выходит из системы$`, func() error {
		return s.Stack.Logout(context.Background(), s.Token)
	})

	ctx.Then(`^выданный JWT-токен больше не действует$`, func() error {
		return s.Stack.ExpectTokenRejected(context.Background(), s.Token)
	})

	ctx.When(`^пользователь указывает nickname "([^"]*)"$`, func(nickname string) error {
		s.LastNickname = nickname
		return s.Stack.SetNickname(context.Background(), s.Token, nickname)
	})

	ctx.Then(`^профиль пользователя содержит nickname "([^"]*)"$`, func(nickname string) error {
		got, err := s.Stack.MyNickname(context.Background(), s.Token)
		if err != nil {
			return err
		}
		if got != nickname {
			return fmt.Errorf("profile nickname mismatch: got %q, want %q", got, nickname)
		}
		return nil
	})
}
