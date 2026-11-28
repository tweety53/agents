package guard

import (
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// envHarness: $FLOW_HARNESS, else claude-code under CLAUDECODE=1, else "".
func TestEnvHarness(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ flow, cc, want string }{
		{"codex", "1", "codex"},
		{"", "1", "claude-code"},
		{"", "", ""},
	} {
		env := core.Env{Getenv: func(k string) string {
			return map[string]string{"FLOW_HARNESS": c.flow, "CLAUDECODE": c.cc}[k]
		}}
		if got := envHarness(env); got != c.want {
			t.Errorf("FLOW_HARNESS=%q CLAUDECODE=%q: envHarness = %q, want %q", c.flow, c.cc, got, c.want)
		}
	}
}

// envToken: $FLOW_SESSION_TOKEN, else the calling agent's token (internal/caller), else "".
func TestEnvToken(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ flow, session, agent, want string }{
		{"mf-x", "s", "a", "mf-x"},
		{"", "s", "a", "cc-s-a"},
		{"", "s", "", "cc-s-main"},
		{"", "", "a", ""},
	} {
		env := core.Env{Getenv: func(k string) string {
			return map[string]string{"FLOW_SESSION_TOKEN": c.flow, "CLAUDE_CODE_SESSION_ID": c.session, "FLOW_CALLER_AGENT": c.agent}[k]
		}}
		if got := envToken(env); got != c.want {
			t.Errorf("%+v: envToken = %q, want %q", c, got, c.want)
		}
	}
}
