package guard

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestCheckVerifyGreen(t *testing.T) {
	t.Parallel()
	row := func(key, token, outcome string) string {
		return `{"key":"` + key + `","role":"verifier","sessionToken":"` + token + `","outcome":"` + outcome + `"}`
	}
	for _, c := range []struct {
		name, rows string
		err        error
		want       int
		needle     string
	}{
		{"every inline verify row completed", "[" + row("verify-fe", "mf-tok", "completed") + "," + row("verify-be", "mf-tok", "completed") + "]", nil, 0, "VERIFY-GREEN"},
		{"a verify still running", "[" + row("verify", "mf-tok", "") + "]", nil, 1, "verify (running)"},
		{"a blocked verify", "[" + row("verify", "mf-tok", "blocked") + "]", nil, 1, "verify (blocked)"},
		{"a blocked verify re-run green by a fix round", "[" + row("verify", "mf-tok", "blocked") + "," + row("verify-fix-1", "mf-tok", "completed") + "]", nil, 0, "1 inline verify row(s) completed"},
		{"a fix round's verify judged per worktree", "[" + row("verify-fe", "mf-tok", "blocked") + "," + row("verify-be", "mf-tok", "completed") + "," + row("verify-fix-1-fe", "mf-tok", "completed") + "]", nil, 0, "2 inline verify row(s) completed"},
		{"a fix round's blocked verify after a green one", "[" + row("verify", "mf-tok", "completed") + "," + row("verify-fix-1", "mf-tok", "blocked") + "]", nil, 1, "verify-fix-1 (blocked)"},
		{"another run's green and a visual verifier do not count", "[" + row("verify", "mf-old", "completed") + "," + row("visual-verify", "mf-tok", "completed") + "]", nil, 1, "no inline verify"},
		{"an unreadable store cannot answer", "", errors.New("down"), 2, "cannot answer"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			env := Env{Dir: t.TempDir(), Dispatches: func(string) ([]byte, error) { return []byte(c.rows), c.err }}
			var out, errb bytes.Buffer
			got := Registry["check-verify-green"]([]string{env.Dir, "demo", "mf-tok"}, env, &out, &errb)
			if got != c.want || !strings.Contains(out.String()+errb.String(), c.needle) {
				t.Fatalf("exit %d, want %d naming %q\n%s%s", got, c.want, c.needle, out.String(), errb.String())
			}
		})
	}
}
