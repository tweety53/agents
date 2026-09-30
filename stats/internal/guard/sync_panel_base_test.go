package guard

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// spbFx is rpbFx's origin clone with a five-line file on the recorded merge
// base, and a change that adds demo.txt and a planning file and edits
// five.txt's last line: a
// base editing five.txt's first line overlaps and still rebases cleanly.
func spbFx(t *testing.T) *bmFx {
	t.Helper()
	fx := &bmFx{dir: t.TempDir()}
	fx.withOrigin()
	fx.g.git(fx.repo, "remote", "set-head", "origin", "main")
	for _, kv := range [][2]string{{"user.name", "Test"}, {"user.email", "test@example.invalid"},
		{"commit.gpgsign", "false"}, {"rebase.autoStash", "false"}} {
		fx.g.git(fx.repo, "config", kv[0], kv[1])
	}
	fx.g.git(fx.repo, "checkout", "-q", "main")
	fx.g.write(fx.repo+"/five.txt", "1\n2\n3\n4\n5\n")
	fx.g.git(fx.repo, "add", "five.txt")
	fx.g.git(fx.repo, "commit", "-qm", "five-line file")
	fx.g.git(fx.repo, "push", "-q", "origin", "main")
	fx.recorded = fx.g.git(fx.repo, "rev-parse", "HEAD")
	fx.g.git(fx.repo, "checkout", "-q", "-B", "demo")
	fx.g.write(fx.repo+"/demo.txt", "demo")
	fx.g.write(fx.repo+"/five.txt", "1\n2\n3\n4\nchange\n")
	mkdirAll(&fx.g, fx.repo+"/spectre/changes/demo")
	fx.g.write(fx.repo+"/spectre/changes/demo/tasks.md", "plan")
	fx.g.git(fx.repo, "add", "demo.txt", "five.txt", "spectre")
	fx.g.git(fx.repo, "commit", "-qm", "the change")
	return fx
}

// spbNext commits a child of origin/main that edits five.txt on branch next,
// for spbMovingStub to move origin/main to once a rebase starts.
func spbNext(fx *bmFx) {
	fx.g.git(fx.repo, "checkout", "-q", "-b", "next", "origin/main")
	fx.g.write(fx.repo+"/five.txt", "base\n2\nnext\n4\n5\n")
	fx.g.git(fx.repo, "commit", "-qam", "next touches five")
	fx.g.git(fx.repo, "checkout", "-q", "demo")
}

// spbMovingStub writes a git that points origin/main at branch next the first
// time a rebase starts -- the base moving under the panel's rebase.
func spbMovingStub(t *testing.T, dir string) string {
	t.Helper()
	writeExec(t, dir+"/stub/git", fmt.Sprintf(`#!/usr/bin/env bash
REAL=%q D=%q
if [ "${3-}" = rebase ] && [ ! -e "$D/.stop" ]; then
  "$REAL" -C "$2" update-ref refs/remotes/origin/main next
  : > "$D/.stop"
fi
exec "$REAL" "$@"
`, fixtureGit, dir))
	return dir + "/stub"
}

func TestSyncPanelBase(t *testing.T) {
	t.Parallel()
	moved := "MOVED: $WT — 1 commits on origin/main since the recorded merge base; "
	clear := "CLEAR: $WT — origin/main has not moved since the recorded merge base\n"
	cases := []struct {
		label   string
		setup   func(fx *bmFx) // nil: origin/main does not move
		rebase  bool
		stub    bool
		args    func(fx *bmFx) []string // nil: {[--rebase], repo, recorded}
		code    int
		out     string // $WT the worktree, $MB HEAD's merge base with origin/main after the run
		err     string // a substring of stderr
		rebased bool   // HEAD moved
	}{
		{label: "an unmoved base is CLEAR and nothing else", code: 0, out: "BASE: main\n" + clear},
		{label: "MOVED without overlap rebases with no prompt", code: 0, rebased: true,
			setup: func(fx *bmFx) { rpbAdvance(fx, map[string]string{"unrelated.txt": "u"}) },
			out: "BASE: main\n" + moved + "no overlap with this change's paths\n" +
				"REBASED: $WT — merge base $MB\n" + clear},
		{label: "MOVED with overlap does not rebase without --rebase", code: 3,
			setup: func(fx *bmFx) { rpbAdvance(fx, map[string]string{"five.txt": "base\n2\n3\n4\n5\n"}) },
			out:   "BASE: main\n" + moved + "overlaps: five.txt\n"},
		{label: "MOVED with overlap rebases with --rebase", code: 0, rebase: true, rebased: true,
			setup: func(fx *bmFx) { rpbAdvance(fx, map[string]string{"five.txt": "base\n2\n3\n4\n5\n"}) },
			out: "BASE: main\n" + moved + "overlaps: five.txt\n" +
				"REBASED: $WT — merge base $MB\n" + clear},
		{label: "a fresh overlap on the re-check stops for the prompt", code: 3, stub: true, rebased: true,
			setup: func(fx *bmFx) {
				rpbAdvance(fx, map[string]string{"unrelated.txt": "u"})
				spbNext(fx)
			},
			out: "BASE: main\n" + moved + "no overlap with this change's paths\n" +
				"REBASED: $WT — merge base $MB\n" + moved + "overlaps: five.txt\n"},
		{label: "a conflict is left mid-rebase with exit 1", code: 1, rebase: true,
			setup: func(fx *bmFx) { rpbAdvance(fx, map[string]string{"five.txt": "base\n2\n3\n4\nbase\n"}) },
			out:   "BASE: main\n" + moved + "overlaps: five.txt\nCONFLICT: $WT — onto $TIP; unmerged: five.txt\n"},
		{label: "no aside: a dirty planning file makes git refuse, exit 2", code: 2,
			setup: func(fx *bmFx) {
				rpbAdvance(fx, map[string]string{"unrelated.txt": "u"})
				fx.g.write(fx.repo+"/spectre/changes/demo/tasks.md", "plan edited")
			},
			out: "BASE: main\n" + moved + "no overlap with this change's paths\n",
			err: "git refused to rebase"},
		{label: "REFUSE exits 2", code: 2, args: func(fx *bmFx) []string { return []string{fx.repo, "-"} },
			out: "BASE: main\nREFUSE: no merge base recorded for $WT — cannot tell whether the base has moved\n"},
		{label: "no origin remote exits 2", code: 2, err: "could not resolve the base branch",
			args: func(fx *bmFx) []string {
				fx.g.git(fx.repo, "remote", "remove", "origin")
				return []string{fx.repo, fx.recorded}
			}},
		{label: "a missing argument is a usage error", code: 2, err: "usage: sync-panel-base.sh",
			args: func(fx *bmFx) []string { return []string{fx.repo} }},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			fx := spbFx(t)
			if c.setup != nil {
				c.setup(fx)
			}
			head := fx.g.git(fx.repo, "rev-parse", "HEAD")
			if fx.g.err != nil {
				t.Fatal(fx.g.err)
			}
			env := Env{Dir: fx.dir, Getenv: os.Getenv}
			if c.stub {
				env.Getenv = pathEnv(spbMovingStub(t, fx.dir))
			}
			args := []string{fx.repo, fx.recorded}
			if c.rebase {
				args = append([]string{"--rebase"}, args...)
			}
			if c.args != nil {
				args = c.args(fx)
			}
			r := runGuard("sync-panel-base", args, env)
			want := strings.ReplaceAll(c.out, "$WT", fx.repo)
			if strings.Contains(want, "$TIP") || strings.Contains(want, "$MB") {
				want = strings.NewReplacer(
					"$TIP", fx.g.git(fx.repo, "rev-parse", "refs/remotes/origin/main"),
					"$MB", fx.g.git(fx.repo, "merge-base", "HEAD", "refs/remotes/origin/main")).Replace(want)
			}
			if r.rc != c.code || r.stdout != want {
				t.Fatalf("got exit %d stdout %q\nwant exit %d stdout %q\nstderr %s", r.rc, r.stdout, c.code, want, r.err)
			}
			if !strings.Contains(r.err, c.err) {
				t.Errorf("stderr %q lacks %q", r.err, c.err)
			}
			if got := fx.g.git(fx.repo, "rev-parse", "HEAD"); (got != head) != c.rebased && c.code != 1 {
				t.Errorf("HEAD %s -> %s, want moved=%v", head, got, c.rebased)
			}
			// The panel's rebase is bare: nothing is ever set aside.
			if s := fx.g.git(fx.repo, "stash", "list"); s != "" {
				t.Errorf("stash list %q: an aside was taken", s)
			}
			if c.code == 1 {
				if _, err := os.Stat(fx.repo + "/.git/rebase-merge"); err != nil {
					t.Error("no rebase in progress: the conflict was not left for the operator")
				}
			}
		})
	}
}
