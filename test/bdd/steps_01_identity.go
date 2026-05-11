//go:build bdd

package bdd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cucumber/godog"
)

var fourDigitCode = regexp.MustCompile(`^\d{4}$`)

type identitySteps struct {
	*State
}

// RegisterIdentitySteps регистрирует шаги эпика identity.
func RegisterIdentitySteps(ctx *godog.ScenarioContext, state *State) {
	steps := identitySteps{State: state}

	ctx.When(`^пользователь запрашивает код доступа для email "([^"]*)"$`, steps.requestEmailCode)
	ctx.Then(`^на email "([^"]*)" отправлен код доступа из 4 цифр$`, steps.emailReceivesFourDigitCode)
	ctx.Given(`^пользователь получил код доступа для email "([^"]*)"$`, steps.userHasEmailCode)
	ctx.When(`^пользователь вводит полученный код доступа для email "([^"]*)"$`, steps.loginWithReceivedEmailCode)
	ctx.Then(`^пользователь получает JWT-токен$`, steps.userReceivesJWT)
	ctx.Given(`^пользователь вошёл по email "([^"]*)"$`, steps.userLoggedInByEmail)
	ctx.When(`^пользователь выходит из системы$`, steps.logout)
	ctx.Then(`^выданный JWT-токен больше не действует$`, steps.issuedTokenIsRejected)
	ctx.When(`^пользователь указывает nickname "([^"]*)"$`, steps.setNickname)
	ctx.Then(`^профиль пользователя содержит nickname "([^"]*)"$`, steps.profileHasNickname)
}

func (s identitySteps) requestEmailCode(email string) error {
	s.Email = email

	after, err := s.Stack.RequestEmailCode(context.Background(), email)
	if err != nil {
		return err
	}
	s.CodeRequestedAfter = after
	return nil
}

func (s identitySteps) emailReceivesFourDigitCode(email string) error {
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
}

func (s identitySteps) userHasEmailCode(email string) error {
	code, err := s.Stack.RequestEmailCodeAndWait(context.Background(), email)
	if err != nil {
		return err
	}
	s.Email = email
	s.Code = code
	return nil
}

func (s identitySteps) loginWithReceivedEmailCode(email string) error {
	token, err := s.Stack.LoginWithEmailCode(context.Background(), email, s.Code)
	if err != nil {
		return err
	}
	s.Email = email
	s.Token = token
	return nil
}

func (s identitySteps) userReceivesJWT() error {
	return s.Stack.ExpectJWT(context.Background(), s.Token, s.Email)
}

func (s identitySteps) userLoggedInByEmail(email string) error {
	token, err := s.Stack.LoginByEmail(context.Background(), email)
	if err != nil {
		return err
	}
	s.Email = email
	s.Token = token
	return nil
}

func (s identitySteps) logout() error {
	return s.Stack.Logout(context.Background(), s.Token)
}

func (s identitySteps) issuedTokenIsRejected() error {
	return s.Stack.ExpectTokenRejected(context.Background(), s.Token)
}

func (s identitySteps) setNickname(nickname string) error {
	s.LastNickname = nickname
	return s.Stack.SetNickname(context.Background(), s.Token, nickname)
}

func (s identitySteps) profileHasNickname(nickname string) error {
	got, err := s.Stack.MyNickname(context.Background(), s.Token)
	if err != nil {
		return err
	}
	if got != nickname {
		return fmt.Errorf("profile nickname mismatch: got %q, want %q", got, nickname)
	}
	return nil
}
