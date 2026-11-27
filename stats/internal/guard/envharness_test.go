package guard

import "testing"

// envHarness: $FLOW_HARNESS, else claude-code under CLAUDECODE=1, else "".
func TestEnvHarness(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ flow, cc, want string }{
		{"codex", "1", "codex"},
		{"", "1", "claude-code"},
		{"", "", ""},
	} {
		env := Env{Getenv: func(k string) string {
			return map[string]string{"FLOW_HARNESS": c.flow, "CLAUDECODE": c.cc}[k]
		}}
		if got := envHarness(env); got != c.want {
			t.Errorf("FLOW_HARNESS=%q CLAUDECODE=%q: envHarness = %q, want %q", c.flow, c.cc, got, c.want)
		}
	}
}
