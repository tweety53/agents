package caller

import (
	"slices"
	"testing"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

// One agent's calls share a token, two agents of one session get two, and
// no harness session leaves the token to the caller.
func TestToken(t *testing.T) {
	main := Token(env(map[string]string{SessionEnv: "s1"}))
	sub := Token(env(map[string]string{SessionEnv: "s1", AgentEnv: "a9"}))
	if main != "cc-s1-main" || sub != "cc-s1-a9" {
		t.Errorf("tokens = %q, %q; want cc-s1-main, cc-s1-a9", main, sub)
	}
	if got := Token(env(map[string]string{AgentEnv: "a9"})); got != "" {
		t.Errorf("no session: token %q, want empty", got)
	}
	if got := For("s1", ""); got != main {
		t.Errorf("For(s1, \"\") = %q, want %q", got, main)
	}
}

func TestStrip(t *testing.T) {
	for _, c := range []struct {
		in, out []string
		agent   string
	}{
		{[]string{Flag, "a9", "stage", "begin"}, []string{"stage", "begin"}, "a9"},
		{[]string{"stage", "begin"}, []string{"stage", "begin"}, ""},
		{[]string{Flag}, []string{Flag}, ""},
		// The hook inserts the flag before one the caller already wrote: both go, the hook's id wins.
		{[]string{Flag, "hook", Flag, "own", "run-verify-list", "wt", "lint"}, []string{"run-verify-list", "wt", "lint"}, "hook"},
	} {
		out, agent := Strip(c.in)
		if !slices.Equal(out, c.out) || agent != c.agent {
			t.Errorf("Strip(%q) = %q, %q; want %q, %q", c.in, out, agent, c.out, c.agent)
		}
	}
}
