package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckFixRevert: a fix that inverts an earlier fix's lines is refused naming both
// commits; a fix that leaves them alone passes.
func TestCheckFixRevert(t *testing.T) {
	for _, tc := range []struct {
		name   string
		second string
		rc     int
		noBase bool
		key    string
	}{
		{"reverts", "a\nb\nc\nd\ne\n", 1, false, ""},
		{"keeps", "a\nB\nC\nD\nE\nf\n", 0, false, ""},
		{"reverts against the recorded flow base", "a\nb\nc\nd\ne\n", 1, true, ""},
		{"reverts a panel fix", "a\nb\nc\nd\ne\n", 1, false, "panel-fix-1"},
		{"reverts a task implementer fix", "a\nb\nc\nd\ne\n", 1, false, "task-3-implementer-fix-1"},
		{"reverts a full-suite fix", "a\nb\nc\nd\ne\n", 1, false, "full-suite-fix-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			gitRun(t, dir, "init", "-q", "-b", "main")
			f := filepath.Join(dir, "x.txt")
			commit := func(body, msg string) {
				writeFile(t, f, body)
				gitRun(t, dir, "add", "x.txt")
				gitRun(t, dir, "commit", "-q", "-m", msg)
			}
			commit("a\nb\nc\nd\ne\n", "base")
			gitRun(t, dir, "tag", "base")
			key := tc.key
			if key == "" {
				key = "fix-1-implementer"
			}
			commit("a\nB\nC\nD\nE\n", "fix: one\n\nFlow-Key: "+key)
			commit(tc.second, "fix: two\n\nFlow-Key: fix-2-implementer")
			env := Env{Dir: dir, Getenv: os.Getenv, LookupEnv: os.LookupEnv}
			args := []string{dir, "base"}
			if tc.noBase {
				gitRun(t, dir, "config", "branch.main.flowBase", "base")
				args = []string{dir}
			}
			r := runGuard("check-fix-revert", args, env)
			if r.rc != tc.rc {
				t.Fatalf("rc %d, want %d: %s", r.rc, tc.rc, r.out)
			}
			if tc.rc == 1 && !strings.Contains(r.out, "inverts 8 lines") {
				t.Fatalf("out: %s", r.out)
			}
		})
	}
}
