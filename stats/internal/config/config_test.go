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
// resolves to DefaultPort, distinguishing "unset" from "invalid". It pins
// FLOWD_PORT to empty because the caller's shell may already export it (a
// /flow worktree does), and FromEnv treats an empty value as unset.
func TestFromEnvDefaultsPortWhenUnset(t *testing.T) {
	t.Setenv("FLOWD_PORT", "")

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != config.DefaultPort {
		t.Fatalf("got port %d, want default %d", cfg.Port, config.DefaultPort)
	}
}

// TestFromEnvJira pins the all-or-nothing rule of the FLOWD_JIRA_* block:
// unset everywhere is a normal unconfigured daemon; all three set resolve;
// any partial combination is the startup refusal ErrPartialJiraConfig; and
// a site that is not https:// is refused with ErrInvalidJiraSite. Each
// case pins FLOWD_* to empty first, because the caller's shell may already
// export one.
func TestFromEnvJira(t *testing.T) {
	setJira := func(t *testing.T, site, email, token string) {
		t.Helper()
		t.Setenv("FLOWD_JIRA_SITE", site)
		t.Setenv("FLOWD_JIRA_EMAIL", email)
		t.Setenv("FLOWD_JIRA_TOKEN", token)
	}

	t.Run("unset means unconfigured", func(t *testing.T) {
		t.Setenv("FLOWD_JIRA_SITE", "")
		t.Setenv("FLOWD_JIRA_EMAIL", "")
		t.Setenv("FLOWD_JIRA_TOKEN", "")
		cfg, err := config.FromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Jira.Configured() {
			t.Fatalf("got configured jira %+v, want unconfigured", cfg.Jira)
		}
	})

	t.Run("all three set resolve", func(t *testing.T) {
		setJira(t, "https://example.atlassian.net/", "ops@example.com", "token")
		cfg, err := config.FromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Jira.Site != "https://example.atlassian.net" || cfg.Jira.Email != "ops@example.com" || cfg.Jira.APIToken != "token" {
			t.Fatalf("got jira %+v, want the resolved site (trailing slash trimmed), email and token", cfg.Jira)
		}
	})

	t.Run("partial is a startup refusal", func(t *testing.T) {
		setJira(t, "https://example.atlassian.net", "ops@example.com", "")
		_, err := config.FromEnv()
		if !errors.Is(err, config.ErrPartialJiraConfig) {
			t.Fatalf("got %v, want config.ErrPartialJiraConfig", err)
		}
	})

	t.Run("http site is a startup refusal", func(t *testing.T) {
		setJira(t, "http://example.atlassian.net", "ops@example.com", "token")
		_, err := config.FromEnv()
		if !errors.Is(err, config.ErrInvalidJiraSite) {
			t.Fatalf("got %v, want config.ErrInvalidJiraSite", err)
		}
	})

	t.Run("degenerate bare scheme is a startup refusal", func(t *testing.T) {
		setJira(t, "https://", "ops@example.com", "token")
		_, err := config.FromEnv()
		if !errors.Is(err, config.ErrInvalidJiraSite) {
			t.Fatalf("got %v, want config.ErrInvalidJiraSite", err)
		}
	})
}

// TestJiraConfigConfigured pins the zero-value reading: a zero JiraConfig
// is unconfigured, and any non-empty field implies all three -- the
// invariant resolveJira's all-or-nothing output lets callers rely on.
func TestJiraConfigConfigured(t *testing.T) {
	if (config.JiraConfig{}).Configured() {
		t.Fatal("zero JiraConfig reported configured")
	}
	if !(config.JiraConfig{Site: "https://x", Email: "e", APIToken: "t"}).Configured() {
		t.Fatal("complete JiraConfig reported unconfigured")
	}
}
