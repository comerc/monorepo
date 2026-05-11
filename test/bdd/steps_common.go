//go:build bdd

package bdd

import "github.com/cucumber/godog"

// RegisterAllSteps регистрирует шаги всех эпиков.
func RegisterAllSteps(ctx *godog.ScenarioContext, s *State) {
	RegisterIdentitySteps(ctx, s)
}
