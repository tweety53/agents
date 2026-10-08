package guard

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// lftFx is one worktree whose since-close sha is the commit carrying a
// tracked tasks.md, with a one-line change on top of it in the working tree.
type lftFx struct {
	wt, tasks, since string
	g                fxGit
}

func lftNewFx(t *testing.T, dir string) *lftFx {
	t.Helper()
	fx := &lftFx{wt: dir}
	fx.tasks = dir + "/spectre/changes/demo/tasks.md"
	fx.g.git("", "init", "-q", "-b", "work", dir)
	mkdir(t, dir+"/spectre/changes/demo")
	fx.g.write(fx.tasks, "- [x] 1. the first task")
	fx.g.write(dir+"/a.txt", "a")
	fx.g.git(dir, "add", ".")
	fx.g.git(dir, "commit", "-qm", "closed clean")
	fx.since = fx.g.git(dir, "rev-parse", "HEAD")
	fx.g.appendLine(dir+"/a.txt", "the late fix")
	return fx
}

// lines appends n lines to name.
func (fx *lftFx) lines(name string, n int) {
	for i := 0; i < n; i++ {
		fx.g.appendLine(fx.wt+"/"+name, fmt.Sprintf("line %d", i))
	}
}

func lftEnv(dir, findings string) Env {
	return Env{Dir: dir, Getenv: os.Getenv,
		Findings:   func(string) ([]byte, error) { return []byte(findings), nil },
		Dispatches: func(string) ([]byte, error) { return []byte("[]"), nil }}
}

func TestCheckLateFixTrigger(t *testing.T) {
	t.Parallel()
	const open = `[{"ref":"F1","status":"open","severity":"Important","round":1,"slot":"primary"}]`
	cases := []struct {
		label    string
		setup    func(t *testing.T, fx, peer *lftFx)
		args     func(fx, peer *lftFx) []string // nil: {demo, tasks, wt, since, CLEAR}
		findings string                         // "": no findings
		code     int
		lines    int    // exit 1: stdout's line count, 0 meaning 1
		out      string // exit 0: the whole stdout; exit 1: the first line's prefix; exit 2: a stderr substring; exit 3: the one line's prefix
	}{
		{label: "every condition holds", code: 0, out: "late-fix reduction: 1 changed lines since $SINCE\n"},
		{label: "condition 1 — an open finding", findings: open, code: 1, out: "full path: condition 1 — "},
		{label: "condition 1 — no earlier clean close", code: 1, out: "full path: condition 1 — ",
			args: func(fx, _ *lftFx) []string { return []string{"demo", fx.tasks, fx.wt, "-", "CLEAR:"} }},
		{label: "condition 2 — the base moved", code: 1, out: "full path: condition 2 — ",
			args: func(fx, _ *lftFx) []string { return []string{"demo", fx.tasks, fx.wt, fx.since, "MOVED:"} }},
		{label: "append scope — over 40 lines", code: 3, out: "append scope: ",
			setup: func(t *testing.T, fx, _ *lftFx) { fx.lines("a.txt", 40) }},
		{label: "append scope — summed across worktrees", code: 3, out: "append scope: 41 changed lines since $SINCE\n",
			setup: func(t *testing.T, fx, peer *lftFx) { fx.lines("a.txt", 20); peer.lines("a.txt", 19) },
			args: func(fx, peer *lftFx) []string {
				return []string{"demo", fx.tasks, fx.wt, fx.since, "CLEAR:", peer.wt, peer.since, "CLEAR:"}
			}},
		{label: "append scope — a binary numstat entry", code: 3, out: "append scope: ",
			setup: func(t *testing.T, fx, _ *lftFx) {
				if err := os.WriteFile(fx.wt+"/blob.bin", []byte{0, 1, 2, 0}, 0o644); err != nil {
					t.Fatal(err)
				}
				fx.g.git(fx.wt, "add", "blob.bin")
			}},
		{label: "append scope — a new task line", code: 3, out: "append scope: ",
			setup: func(t *testing.T, fx, _ *lftFx) { fx.g.appendLine(fx.tasks, "- [ ] 2. a second task") }},
		{label: "full path — a new task line with a moved base", code: 1, lines: 2, out: "full path: condition 2 — ",
			setup: func(t *testing.T, fx, _ *lftFx) { fx.g.appendLine(fx.tasks, "- [ ] 2. a second task") },
			args:  func(fx, _ *lftFx) []string { return []string{"demo", fx.tasks, fx.wt, fx.since, "MOVED:"} }},
		{label: "append scope — an untracked tasks.md", code: 3, out: "append scope: ",
			setup: func(t *testing.T, fx, _ *lftFx) { fx.g.write(fx.wt+"/untracked-tasks.md", "- [ ] 1. x") },
			args: func(fx, _ *lftFx) []string {
				return []string{"demo", fx.wt + "/untracked-tasks.md", fx.wt, fx.since, "CLEAR:"}
			}},
		{label: "condition 5 — machinery renamed away", code: 1, out: "full path: condition 5 — ",
			setup: func(t *testing.T, fx, _ *lftFx) {
				mkdir(t, fx.wt+"/skills/flow")
				fx.g.write(fx.wt+"/skills/flow/review-panel.md", "the panel")
				fx.g.git(fx.wt, "add", ".")
				fx.g.git(fx.wt, "commit", "-qm", "machinery at the clean close")
				fx.since = fx.g.git(fx.wt, "rev-parse", "HEAD")
				fx.g.git(fx.wt, "mv", "skills/flow/review-panel.md", "skills/flow/notes.md")
			}},
		{label: "condition 5 — the panel's own machinery", code: 1, out: "full path: condition 5 — ",
			setup: func(t *testing.T, fx, _ *lftFx) {
				mkdir(t, fx.wt+"/skills/flow")
				fx.g.write(fx.wt+"/skills/flow/review-panel-extra.md", "x")
				fx.g.git(fx.wt, "add", "skills")
			}},
		{label: "cannot answer — usage", code: 2, out: "usage: check-late-fix-trigger.sh",
			args: func(fx, _ *lftFx) []string { return []string{"demo", fx.tasks, fx.wt, fx.since} }},
		{label: "cannot answer — a sha that does not resolve", code: 2, out: "does not resolve",
			args: func(fx, _ *lftFx) []string { return []string{"demo", fx.tasks, fx.wt, "deadbeef", "CLEAR:"} }},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			fx, peer := lftNewFx(t, dir+"/wt"), lftNewFx(t, dir+"/peer")
			if tc.setup != nil {
				tc.setup(t, fx, peer)
			}
			for _, f := range []*lftFx{fx, peer} {
				if f.g.err != nil {
					t.Fatal(f.g.err)
				}
			}
			args := []string{"demo", fx.tasks, fx.wt, fx.since, "CLEAR: " + fx.wt + " — main has not moved since the recorded merge base"}
			if tc.args != nil {
				args = tc.args(fx, peer)
			}
			findings := tc.findings
			if findings == "" {
				findings = "[]"
			}
			r := runGuard("check-late-fix-trigger", args, lftEnv(dir, findings))
			want := strings.ReplaceAll(tc.out, "$SINCE", fx.since)
			lines := max(tc.lines, 1)
			switch {
			case r.rc != tc.code:
				t.Errorf("exit %d, want %d\n%s", r.rc, tc.code, r.out)
			case tc.code == 0 && r.stdout != want:
				t.Errorf("stdout %q, want %q", r.stdout, want)
			case tc.code == 1 && (strings.Count(r.stdout, "\n") != lines || !strings.HasPrefix(r.stdout, want)):
				t.Errorf("stdout %q, want exactly %d line(s), the first starting %q", r.stdout, lines, want)
			case tc.code == 3 && (strings.Count(r.stdout, "\n") != 1 || !strings.HasPrefix(r.stdout, want)):
				t.Errorf("stdout %q, want exactly one line starting %q", r.stdout, want)
			case tc.code == 2 && (r.stdout != "" || !strings.Contains(r.err, want)):
				t.Errorf("stdout %q stderr %q, want nothing on stdout and stderr naming %q", r.stdout, r.err, want)
			}
		})
	}
}
