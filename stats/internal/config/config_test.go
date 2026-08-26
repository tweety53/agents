package config_test

import (
	"errors"
	"testing"

	"github.com/tweety53/agents/stats/internal/config"
)

// TestFromEnvRejectsInvalidPort asserts a non-integer FLOWD_PORT is a
// startup error (config.ErrInvalidPort), not a silently-discarded parse
// failure that leaves the daemon listening on DefaultPort with no
// diagnostic. Mutation check performed by hand: reverting FromEnv to
// discard strconv.Atoi's error (`if p, err := strconv.Atoi(v); err == nil
// { cfg.Port = p }`) makes this test fail -- FromEnv returns (Config{Port:
// DefaultPort}, nil) instead of an error -- confirming the test actually
// catches the regression this finding was about.
func TestFromEnvRejectsInvalidPort(t *testing.T) {
	t.Setenv("FLOWD_PORT", "abc")

	_, err := config.FromEnv()
	if err == nil {
		t.Fatal("expected an error for a non-integer FLOWD_PORT, got nil")
	}
	if !errors.Is(err, config.ErrInvalidPort) {
		t.Fatalf("expected config.ErrInvalidPort, got %v", err)
	}
}

// TestFromEnvAcceptsValidPort confirms the above is not a blanket
// rejection: a genuine integer port still resolves, unchanged.
func TestFromEnvAcceptsValidPort(t *testing.T) {
	t.Setenv("FLOWD_PORT", "9999")

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9999 {
		t.Fatalf("got port %d, want 9999", cfg.Port)
	}
}

// TestFromEnvDefaultsPortWhenUnset confirms an unset FLOWD_PORT still
// resolves to DefaultPort, distinguishing "unset" from "invalid".
func TestFromEnvDefaultsPortWhenUnset(t *testing.T) {
	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != config.DefaultPort {
		t.Fatalf("got port %d, want default %d", cfg.Port, config.DefaultPort)
	}
}
