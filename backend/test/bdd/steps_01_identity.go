//go:build bdd

package bdd

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

var fiveDigitCode = regexp.MustCompile(`^\d{5}$`)

type identitySteps struct {
	*State
}

// RegisterIdentitySteps регистрирует шаги эпика identity.
func RegisterIdentitySteps(ctx *godog.ScenarioContext, state *State) {
	steps := identitySteps{State: state}

	ctx.When(`^пользователь запрашивает код доступа для email "([^"]*)"$`, steps.requestEmailCode)
	ctx.Then(`^на email "([^"]*)" отправлен код доступа из 5 цифр$`, steps.emailReceivesFiveDigitCode)
	ctx.Then(`^пользователь видит ошибку email "([^"]*)"$`, steps.userSeesEmailError)
	ctx.Given(`^пользователь запросил код доступа для email "([^"]*)"$`, steps.userRequestedEmailCode)
	ctx.When(`^пользователь повторно запрашивает код доступа для email "([^"]*)"$`, steps.requestEmailCodeAgain)
	ctx.Then(`^следующий код доступа можно запросить через (\d+) секунд$`, steps.nextEmailCodeAvailableAfter)
	ctx.Given(`^пользователь исчерпал быстрые повторные запросы кода для email "([^"]*)"$`, steps.userExhaustedFastCodeRequests)
	ctx.Then(`^задержки повторной отправки кода для email "([^"]*)" равны 30, 60, 120 и 86400 секунд$`, steps.emailCodeRetryDelaysAre)
	ctx.Given(`^пользователь получил код доступа для email "([^"]*)"$`, steps.userHasEmailCode)
	ctx.When(`^пользователь вводит полученный код доступа для email "([^"]*)"$`, steps.loginWithReceivedEmailCode)
	ctx.When(`^пользователь вводит код доступа "([^"]*)" для email "([^"]*)"$`, steps.loginWithEmailCode)
	ctx.Then(`^пользователь видит ошибку входа "([^"]*)"$`, steps.userSeesLoginError)
	ctx.Given(`^код доступа для email "([^"]*)" истёк$`, steps.emailCodeExpired)
	ctx.Given(`^пользователь получил два кода доступа для email "([^"]*)"$`, steps.userHasTwoEmailCodes)
	ctx.When(`^пользователь вводит первый полученный код доступа для email "([^"]*)"$`, steps.loginWithFirstReceivedEmailCode)
	ctx.When(`^пользователь вводит второй полученный код доступа для email "([^"]*)"$`, steps.loginWithSecondReceivedEmailCode)
	ctx.Then(`^второй код доступа для email "([^"]*)" больше не действует$`, steps.secondEmailCodeIsRejected)
	ctx.Then(`^повторное использование этого кода для email "([^"]*)" отклоняется$`, steps.reusedEmailCodeIsRejected)
	ctx.Then(`^новый код доступа для email "([^"]*)" можно запросить сразу$`, steps.newEmailCodeCanBeRequestedImmediately)
	ctx.Then(`^пользователь получает JWT-токен$`, steps.userReceivesJWT)
	ctx.Given(`^пользователь вошёл по email "([^"]*)"$`, steps.userLoggedInByEmail)
	ctx.Given(`^пользователь вошёл по email "([^"]*)" с nickname "([^"]*)"$`, steps.userLoggedInByEmailWithNickname)
	ctx.Given(`^пользователь вошёл по email "([^"]*)" в двух сессиях$`, steps.userLoggedInByEmailInTwoSessions)
	ctx.When(`^пользователь выходит из системы$`, steps.logout)
	ctx.When(`^пользователь выходит из текущей сессии$`, steps.logout)
	ctx.When(`^пользователь выходит из системы на всех устройствах$`, steps.logoutEverywhere)
	ctx.Then(`^выданный JWT-токен больше не действует$`, steps.issuedTokenIsRejected)
	ctx.Then(`^текущая сессия больше не действует$`, steps.issuedTokenIsRejected)
	ctx.Then(`^другая сессия пользователя продолжает действовать$`, steps.otherSessionIsAccepted)
	ctx.Then(`^все сессии пользователя больше не действуют$`, steps.allSessionsAreRejected)
	ctx.When(`^пользователь указывает nickname "([^"]*)"$`, steps.setNickname)
	ctx.Given(`^nickname "([^"]*)" уже занят другим пользователем$`, steps.nicknameTakenByAnotherUser)
	ctx.Then(`^профиль пользователя содержит nickname "([^"]*)"$`, steps.profileHasNickname)
	ctx.Then(`^пользователь видит ошибку nickname "([^"]*)"$`, steps.userSeesNicknameError)
	ctx.When(`^пользователь открывает профиль без входа$`, steps.openProfileWithoutLogin)
	ctx.Then(`^пользователь видит ошибку профиля "([^"]*)"$`, steps.userSeesProfileError)
}

func (s identitySteps) requestEmailCode(email string) error {
	s.Email = email

	after, err := s.Stack.RequestEmailCode(context.Background(), email)
	if err != nil {
		s.LastError = err.Error()
		return nil
	}
	s.CodeRequestedAfter = after
	return nil
}

func (s identitySteps) emailReceivesFiveDigitCode(email string) error {
	code, err := s.Stack.WaitEmailCode(context.Background(), email, s.CodeRequestedAfter)
	if err != nil {
		return err
	}
	if !fiveDigitCode.MatchString(code) {
		return fmt.Errorf("email code must contain 5 digits, got %q", code)
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

func (s identitySteps) userRequestedEmailCode(email string) error {
	return s.requestEmailCode(email)
}

func (s identitySteps) requestEmailCodeAgain(email string) error {
	before := s.Stack.MailCount()
	payload, err := s.Stack.RequestEmailCodePayload(context.Background(), email)
	if err != nil {
		s.LastError = err.Error()
		return nil
	}
	s.LastAccepted = payload.Accepted
	s.LastRetryAfter = payload.RetryAfterSeconds
	s.LastMailCount = s.Stack.MailCount() - before
	return nil
}

func (s identitySteps) nextEmailCodeAvailableAfter(seconds int) error {
	if s.LastAccepted {
		return fmt.Errorf("requestEmailCode was accepted during cooldown")
	}
	if s.LastRetryAfter != seconds {
		return fmt.Errorf("retryAfterSeconds mismatch: got %d, want %d", s.LastRetryAfter, seconds)
	}
	if s.LastMailCount != 0 {
		return fmt.Errorf("cooldown request sent %d unexpected emails", s.LastMailCount)
	}
	return nil
}

func (s identitySteps) userExhaustedFastCodeRequests(email string) error {
	s.Email = email
	s.RetryDelays = nil
	for i := 0; i < 4; i++ {
		payload, err := s.Stack.RequestEmailCodePayload(context.Background(), email)
		if err != nil {
			return err
		}
		s.RetryDelays = append(s.RetryDelays, payload.RetryAfterSeconds)
		if err := s.Stack.ForceEmailCodeRequestAllowed(context.Background(), email); err != nil {
			return err
		}
	}
	return nil
}

func (s identitySteps) emailCodeRetryDelaysAre(email string) error {
	want := []int{30, 60, 120, 86400}
	if len(s.RetryDelays) != len(want) {
		return fmt.Errorf("retry delay count mismatch for %q: got %v, want %v", email, s.RetryDelays, want)
	}
	for i := range want {
		if s.RetryDelays[i] != want[i] {
			return fmt.Errorf("retry delay mismatch for %q: got %v, want %v", email, s.RetryDelays, want)
		}
	}
	return nil
}

func (s identitySteps) loginWithReceivedEmailCode(email string) error {
	token, err := s.Stack.LoginWithEmailCode(context.Background(), email, s.Code)
	if err != nil {
		s.LastError = err.Error()
		return nil
	}
	s.Email = email
	s.Token = token
	return nil
}

func (s identitySteps) loginWithEmailCode(code string, email string) error {
	if err := s.Stack.ExpectLoginWithEmailCodeRejected(context.Background(), email, code, "invalid code"); err != nil {
		return err
	}
	s.LastError = "invalid code"
	return nil
}

func (s identitySteps) userSeesLoginError(message string) error {
	if !strings.Contains(s.LastError, message) {
		return fmt.Errorf("expected login error %q, got %q", message, s.LastError)
	}
	return nil
}

func (s identitySteps) emailCodeExpired(email string) error {
	return s.Stack.ExpireEmailCodes(context.Background(), email)
}

func (s identitySteps) userHasTwoEmailCodes(email string) error {
	firstAfter, err := s.Stack.RequestEmailCode(context.Background(), email)
	if err != nil {
		return err
	}
	firstCode, err := s.Stack.WaitEmailCode(context.Background(), email, firstAfter)
	if err != nil {
		return err
	}
	if err := s.Stack.ForceEmailCodeRequestAllowed(context.Background(), email); err != nil {
		return err
	}
	secondAfter, err := s.Stack.RequestEmailCode(context.Background(), email)
	if err != nil {
		return err
	}
	secondCode, err := s.Stack.WaitEmailCode(context.Background(), email, secondAfter)
	if err != nil {
		return err
	}
	s.Email = email
	s.Code = firstCode
	s.Codes = []string{firstCode, secondCode}
	return nil
}

func (s identitySteps) loginWithFirstReceivedEmailCode(email string) error {
	token, err := s.Stack.LoginWithEmailCode(context.Background(), email, s.Codes[0])
	if err != nil {
		return err
	}
	s.Email = email
	s.Code = s.Codes[0]
	s.Token = token
	return nil
}

func (s identitySteps) loginWithSecondReceivedEmailCode(email string) error {
	token, err := s.Stack.LoginWithEmailCode(context.Background(), email, s.Codes[1])
	if err != nil {
		return err
	}
	s.Email = email
	s.Code = s.Codes[1]
	s.Token = token
	return nil
}

func (s identitySteps) secondEmailCodeIsRejected(email string) error {
	return s.Stack.ExpectLoginWithEmailCodeRejected(context.Background(), email, s.Codes[1], "invalid code")
}

func (s identitySteps) reusedEmailCodeIsRejected(email string) error {
	return s.Stack.ExpectLoginWithEmailCodeRejected(context.Background(), email, s.Code, "invalid code")
}

func (s identitySteps) newEmailCodeCanBeRequestedImmediately(email string) error {
	before := s.Stack.MailCount()
	payload, err := s.Stack.RequestEmailCodePayload(context.Background(), email)
	if err != nil {
		return err
	}
	if !payload.Accepted {
		return fmt.Errorf("requestEmailCode was not accepted after successful login")
	}
	if s.Stack.MailCount() <= before {
		return fmt.Errorf("requestEmailCode after successful login did not send email")
	}
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

func (s identitySteps) userLoggedInByEmailWithNickname(email string, nickname string) error {
	if err := s.userLoggedInByEmail(email); err != nil {
		return err
	}
	return s.Stack.SetNickname(context.Background(), s.Token, nickname)
}

func (s identitySteps) userLoggedInByEmailInTwoSessions(email string) error {
	token, err := s.Stack.LoginByEmail(context.Background(), email)
	if err != nil {
		return err
	}
	if err := s.Stack.ForceEmailCodeRequestAllowed(context.Background(), email); err != nil {
		return err
	}
	otherToken, err := s.Stack.LoginByEmail(context.Background(), email)
	if err != nil {
		return err
	}
	s.Email = email
	s.Token = token
	s.OtherToken = otherToken
	return nil
}

func (s identitySteps) logout() error {
	return s.Stack.Logout(context.Background(), s.Token)
}

func (s identitySteps) logoutEverywhere() error {
	return s.Stack.LogoutEverywhere(context.Background(), s.Token)
}

func (s identitySteps) issuedTokenIsRejected() error {
	return s.Stack.ExpectTokenRejected(context.Background(), s.Token)
}

func (s identitySteps) otherSessionIsAccepted() error {
	return s.Stack.ExpectJWT(context.Background(), s.OtherToken, s.Email)
}

func (s identitySteps) allSessionsAreRejected() error {
	if err := s.Stack.ExpectTokenRejected(context.Background(), s.Token); err != nil {
		return err
	}
	return s.Stack.ExpectTokenRejected(context.Background(), s.OtherToken)
}

func (s identitySteps) setNickname(nickname string) error {
	s.LastNickname = nickname
	err := s.Stack.SetNickname(context.Background(), s.Token, nickname)
	if err != nil {
		s.LastError = err.Error()
		return nil
	}
	return nil
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

func (s identitySteps) nicknameTakenByAnotherUser(nickname string) error {
	token, err := s.Stack.LoginByEmail(context.Background(), "other-"+nickname+"@example.com")
	if err != nil {
		return err
	}
	return s.Stack.SetNickname(context.Background(), token, nickname)
}

func (s identitySteps) userSeesNicknameError(message string) error {
	if !strings.Contains(s.LastError, message) {
		return fmt.Errorf("expected nickname error %q, got %q", message, s.LastError)
	}
	return nil
}

func (s identitySteps) openProfileWithoutLogin() error {
	err := s.Stack.ExpectMyProfileRejected(context.Background(), "", "unauthenticated")
	if err != nil {
		return err
	}
	s.LastError = "unauthenticated"
	return nil
}

func (s identitySteps) userSeesProfileError(message string) error {
	if !strings.Contains(s.LastError, message) {
		return fmt.Errorf("expected profile error %q, got %q", message, s.LastError)
	}
	return nil
}

func (s identitySteps) userSeesEmailError(message string) error {
	if !strings.Contains(s.LastError, message) {
		return fmt.Errorf("expected email error %q, got %q", message, s.LastError)
	}
	return nil
}
