package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckScopeSibling pins tcfCheckScopeSibling over two repositories of
// one project: gymie-playwright, whose commit touches only
// fidelity/builder.spec.ts, and gymie, which carries a `group` module.
func TestCheckScopeSibling(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	own, sib := filepath.Join(root, "gymie-playwright"), filepath.Join(root, "gymie")
	g := &fxGit{}
	for _, r := range []string{own, sib} {
		g.git("", "init", "-q", "-b", "main", r)
	}
	for _, p := range []string{"src/group/Group.kt", "e2e/x.ts", "helpers/h.ts", "workout-ui/w.kt", "playwright/p.ts", "build/b.kts"} {
		writeFile(t, filepath.Join(sib, p), "x\n")
	}
	g.git(sib, "add", "-A")
	g.git(sib, "commit", "-qm", "base")
	writeFile(t, filepath.Join(own, ".flow", "project.md"),
		"# p\n\n## apps\n\n| App | Repo root | Kind |\n|-----|-----------|------|\n| Suite | `"+own+"` | Node |\n| Backend | `"+sib+"` | Kotlin |\n")
	for _, p := range []string{"docs/self-review/kan-30-group-workouts.md", "helpers/h.ts", "workout/a.ts", "ui/b.ts"} {
		writeFile(t, filepath.Join(own, p), "x\n")
	}
	g.git(own, "add", "-A")
	g.git(own, "commit", "-qm", "test(e2e): base")
	writeFile(t, filepath.Join(own, "fidelity", "builder.spec.ts"), "y\n")
	g.git(own, "add", "-A")
	g.git(own, "commit", "-qm", "work")
	if g.err != nil {
		t.Fatal(g.err)
	}
	diffs := []tcfRepoDiff{{worktree: own, commit: "HEAD", parent: "HEAD^", changed: []string{"fidelity/builder.spec.ts"}}}
	cases := []struct {
		name, subject, projectDir string
		refused                   bool
	}{
		{"kan-982 F16: a backend module on a playwright-only commit", "fix(group): the sheet closes", own, true},
		{"a touched path passes", "fix(fidelity): the sheet closes", own, false},
		{"a scope no repository carries passes", "test(visual): rebaseline", own, false},
		{"an earlier scope of this repository passes", "fix(e2e): settle", own, false},
		{"a directory of this repository passes", "fix(helpers): tidy", own, false},
		{"every part a directory of this repository passes", "fix(workout-ui): tidy", own, false},
		{"a word of this repository's name passes", "fix(playwright): config", own, false},
		{"a sibling directory this repository lacks is refused", "chore(build): bump", own, true},
		{"no apps table refuses nothing", "fix(group): the sheet closes", t.TempDir(), false},
		{"no scope passes", "fix: the sheet closes", own, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			subject := tc.subject
			v, err := tcfCheckScopeSibling(Env{Getenv: os.Getenv, Dir: own}, tcfTask{id: "21", commit: &subject}, diffs, tc.projectDir)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(v) > 0; got != tc.refused {
				t.Fatalf("refused = %v, want %v: %q", got, tc.refused, v)
			}
			if tc.refused && !strings.Contains(v[0], "names a module of gymie and nothing this commit touched") {
				t.Errorf("message %q does not name the sibling", v[0])
			}
		})
	}
}
