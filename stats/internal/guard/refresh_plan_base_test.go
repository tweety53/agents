package guard

import (
	"os"
	"strings"
	"testing"
)

// rpbFx is bmFx.withOrigin with origin/HEAD set -- resolve-base-branch reads
// it -- and a plan whose **Files:** field wraps onto a second line.
func rpbFx(t *testing.T) *bmFx {
	t.Helper()
	fx := &bmFx{dir: t.TempDir()}
	fx.withOrigin()
	fx.g.git(fx.repo, "remote", "set-head", "origin", "main")
	fx.g.write(fx.dir+"/tasks.md", "- [ ] 1. a task\n\n"+
		"**Files:** `stats/`, `other.txt`, `oth`,\n"+
		"`docs/x.md`, `../sibling-repo/other.txt`\n"+
		"**Tests:** `TestX`\n")
	return fx
}

// rpbAdvance commits files on main and pushes it, so origin/main moves.
func rpbAdvance(fx *bmFx, files map[string]string) {
	fx.g.git(fx.repo, "checkout", "-q", "main")
	for f, body := range files {
		mkdirAll(&fx.g, fx.repo+"/"+f[:strings.LastIndex(f, "/")+1])
		fx.g.write(fx.repo+"/"+f, body)
		fx.g.git(fx.repo, "add", f)
	}
	fx.g.git(fx.repo, "commit", "-qm", "base moves")
	fx.g.git(fx.repo, "push", "-q", "origin", "main")
	fx.g.git(fx.repo, "checkout", "-q", "demo")
}

func mkdirAll(g *fxGit, dir string) {
	if g.err == nil && dir != "" {
		g.err = os.MkdirAll(dir, 0o755)
	}
}

func TestRefreshPlanBase(t *testing.T) {
	t.Parallel()
	specs := []string{"spectre/specs/cap.md", "spectre/specs/gone.md"}
	cases := []struct {
		label string
		move  map[string]string // nil: origin/main does not move
		args  func(fx *bmFx) []string
		code  int
		out   string // $WT the worktree, $MB the merge base
		err   string // a substring of stderr
	}{
		{label: "an unmoved base is one line and prints no spec", code: 0,
			out: "UNMOVED: $WT — origin/main has not moved since $MB\n"},
		{label: "a moved base names the plan's changed paths and prints each spec at origin", code: 0,
			move: map[string]string{
				"stats/a.go": "package a", "other.txt": "o", "othello.txt": "not a Files match",
				"docs/x.md": "x", "unrelated.txt": "u", "spectre/specs/cap.md": "# cap at the moved base",
			},
			out: "MOVED: $WT — 1 commits on origin/main since $MB\n" +
				"CHANGED: docs/x.md\nCHANGED: other.txt\nCHANGED: stats/a.go\n" +
				"----- BEGIN spectre/specs/cap.md @ origin/main -----\n# cap at the moved base\n" +
				"----- END spectre/specs/cap.md -----\n" +
				"SPEC-ABSENT: spectre/specs/gone.md — not at origin/main\n"},
		{label: "a moved base touching none of the plan's paths says so", code: 0,
			move: map[string]string{"unrelated.txt": "u"},
			out: "MOVED: $WT — 1 commits on origin/main since $MB\nCHANGED: none of the plan's **Files:** paths\n" +
				"SPEC-ABSENT: spectre/specs/cap.md — not at origin/main\n" +
				"SPEC-ABSENT: spectre/specs/gone.md — not at origin/main\n"},
		{label: "too few arguments is usage, exit 2", code: 2,
			args: func(fx *bmFx) []string { return []string{fx.repo, fx.recorded} },
			err:  "usage: refresh-plan-base.sh <worktree> <merge-base> <tasks.md> [<spec-path>…]"},
		{label: "an unresolvable merge base is exit 2", code: 2,
			args: func(fx *bmFx) []string { return []string{fx.repo, "no-such-ref", fx.dir + "/tasks.md"} },
			err:  "no-such-ref is not a commit"},
		{label: "an unreadable tasks.md is exit 2", code: 2,
			args: func(fx *bmFx) []string { return []string{fx.repo, fx.recorded, fx.dir + "/missing.md"} },
			err:  "cannot read"},
		{label: "no origin remote is exit 2", code: 2,
			args: func(fx *bmFx) []string {
				fx.g.git(fx.repo, "remote", "remove", "origin")
				return []string{fx.repo, fx.recorded, fx.dir + "/tasks.md"}
			},
			err: "no 'origin' remote"},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			fx := rpbFx(t)
			if c.move != nil {
				rpbAdvance(fx, c.move)
			}
			head := fx.g.git(fx.repo, "rev-parse", "HEAD")
			if fx.g.err != nil {
				t.Fatal(fx.g.err)
			}
			args := append([]string{fx.repo, fx.recorded, fx.dir + "/tasks.md"}, specs...)
			if c.args != nil {
				args = c.args(fx)
			}
			r := runGuard("refresh-plan-base", args, Env{Dir: fx.dir, Getenv: os.Getenv})
			if r.rc != c.code {
				t.Fatalf("exit %d, want %d\n%s", r.rc, c.code, r.out)
			}
			want := strings.NewReplacer("$WT", fx.repo, "$MB", fx.recorded).Replace(c.out)
			if r.stdout != want {
				t.Errorf("stdout:\n%s\nwant:\n%s", r.stdout, want)
			}
			if !strings.Contains(r.err, c.err) {
				t.Errorf("stderr %q lacks %q", r.err, c.err)
			}
			// It names, it never rebases: HEAD and the branch stay put.
			if got := fx.g.git(fx.repo, "rev-parse", "HEAD"); c.code == 0 && got != head {
				t.Errorf("HEAD moved %s -> %s", head, got)
			}
		})
	}

	// A base that renames a declared path away still names the old path:
	// rename detection would list only the new one.
	t.Run("a declared path renamed away on the base", func(t *testing.T) {
		t.Parallel()
		fx := rpbFx(t)
		rpbAdvance(fx, map[string]string{"sub/other.txt": "x"})
		mb := fx.g.git(fx.repo, "rev-parse", "main")
		fx.g.write(fx.dir+"/tasks.md", "- [ ] 1. a task\n\n**Files:** `sub/other.txt`\n**Tests:** `TestX`\n")
		fx.g.git(fx.repo, "checkout", "-q", "main")
		fx.g.git(fx.repo, "mv", "sub/other.txt", "sub/renamed.txt")
		fx.g.git(fx.repo, "commit", "-qm", "base renames it")
		fx.g.git(fx.repo, "push", "-q", "origin", "main")
		fx.g.git(fx.repo, "checkout", "-q", "demo")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		r := runGuard("refresh-plan-base", []string{fx.repo, mb, fx.dir + "/tasks.md"}, Env{Dir: fx.dir, Getenv: os.Getenv})
		if r.rc != 0 || !strings.Contains(r.stdout, "CHANGED: sub/other.txt\n") {
			t.Fatalf("exit %d stdout %q, want exit 0 naming CHANGED: sub/other.txt\nstderr %s", r.rc, r.stdout, r.err)
		}
	})
}
