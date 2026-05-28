//go:build bdd

package bdd

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var chdirOnce sync.Once

func runEpic(t *testing.T, epic string) {
	t.Helper()

	for _, path := range featurePaths(epic) {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		t.Run(name, func(t *testing.T) {
			runFeature(t, path)
		})
	}
}

func runFeature(t *testing.T, path string) {
	t.Helper()

	stack, err := GetOrCreateStack(t)
	if err != nil {
		t.Fatalf("BDD stack init: %v", err)
	}

	testSuite := godog.TestSuite{
		Name: "bdd/" + strings.TrimSuffix(path, filepath.Ext(path)),
		Options: &godog.Options{
			Tags:     "@api",
			Format:   godogFormat(),
			Paths:    []string{path},
			Output:   colors.Colored(os.Stdout),
			TestingT: t,
			Strict:   true,
		},
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			s := &State{Stack: stack}
			ctx.Before(func(gctx context.Context, _ *godog.Scenario) (context.Context, error) {
				return gctx, s.Reset(t)
			})
			RegisterAllSteps(ctx, s)
		},
	}

	if testSuite.Run() != 0 {
		t.Fatal("bdd feature failed")
	}
}

func chdirProjectRoot() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to resolve backend/test/bdd/runner.go path via runtime.Caller")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "..")
	if err := os.Chdir(root); err != nil {
		panic("failed to chdir to project root: " + err.Error())
	}
}

func featurePaths(epic string) []string {
	if env := os.Getenv("BDD_PATHS"); env != "" {
		return strings.Split(env, ",")
	}
	matches, err := filepath.Glob(filepath.Join("features", epic, "*.feature"))
	if err != nil {
		panic("failed to list feature files: " + err.Error())
	}
	if len(matches) == 0 {
		panic("feature files not found for epic: " + epic)
	}
	return matches
}

func godogFormat() string {
	if format := os.Getenv("BDD_GODOG_FORMAT"); format != "" {
		return format
	}
	return "pretty"
}
