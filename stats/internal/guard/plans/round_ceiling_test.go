package plans

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

func TestRoundCeiling(t *testing.T) {
	t.Parallel()
	project := func(t *testing.T, body string) string {
		wt := t.TempDir()
		if body != "" {
			if err := os.MkdirAll(filepath.Join(wt, ".flow"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(wt, ".flow", "project.md"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return wt
	}
	cases := []struct {
		name, body, failed string
		rc                 int
		want               string
	}{
		{"first failure retries", "", "1", 0, "NEXT: retry"},
		{"third failure rethinks", "", "3", 0, "NEXT: rethink"},
		{"below the default ceiling rethinks", "", "5", 0, "NEXT: rethink"},
		{"the default ceiling defers", "", "6", 0, "NEXT: defer"},
		{"past the ceiling defers", "", "9", 0, "NEXT: defer"},
		{"a configured ceiling defers", "## fix round ceiling\n\n4 rounds, then defer\n", "4", 0, "NEXT: defer"},
		{"a configured ceiling rethinks below it", "## fix round ceiling\n4\n", "3", 0, "NEXT: rethink"},
		{"a ceiling under three defers before rethinking", "## fix round ceiling\n2\n", "2", 0, "NEXT: defer"},
		{"a malformed ceiling", "## fix round ceiling\nmany\n", "1", 2, "many"},
		{"a zero ceiling", "## fix round ceiling\n0\n", "1", 2, "fix round ceiling"},
		{"no failed rounds", "", "0", 2, "usage"},
		{"a non-number round", "", "x", 2, "usage"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := runGuard("round-ceiling", []string{project(t, c.body), c.failed}, core.Env{})
			if r.Rc != c.rc || !strings.Contains(r.Out, c.want) {
				t.Fatalf("rc=%d out=%q, want rc=%d containing %q", r.Rc, r.Out, c.rc, c.want)
			}
		})
	}
	if r := runGuard("round-ceiling", nil, core.Env{}); r.Rc != 2 {
		t.Fatalf("no args: rc=%d, want 2", r.Rc)
	}
}
