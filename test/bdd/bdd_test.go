//go:build bdd

package bdd

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	chdirOnce.Do(chdirProjectRoot)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	stackOnce.Do(func() {
		liveStack, liveStackErr = startStack(ctx)
	})
	cancel()

	if liveStackErr != nil {
		fmt.Fprintf(os.Stderr, "BDD stack init: %v\n", liveStackErr)
		os.Exit(1)
	}

	code := m.Run()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	liveStack.Close(shutdownCtx)
	shutdownCancel()

	os.Exit(code)
}

func TestIdentity(t *testing.T) {
	runEpic(t, "01_identity")
}

func TestIdentity2(t *testing.T) {
	// runEpic(t, "01_identity")
}
