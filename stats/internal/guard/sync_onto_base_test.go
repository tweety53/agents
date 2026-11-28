package guard

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// sobEnv runs the guard with root as FLOW_GUARD_REPO_ROOT,
// and stub's directory first on PATH when stub is not "".
func sobEnv(t *testing.T, dir, root, stub string) core.Env {
	t.Helper()
	// The moving stub execs FixtureGit itself; without it, FixtureGitPath.
	path := guardtest.FixtureGitPath(t)
	get := func(k string) string {
		if k == "PATH" {
			return path
		}
		return os.Getenv(k)
	}
	if stub != "" {
		get = guardtest.PathEnv(stub)
	}
	return core.Env{Dir: dir, Getenv: func(k string) string {
		if k == "FLOW_GUARD_REPO_ROOT" {
			return root
		}
		return get(k)
	}}
}

// sobOverlapFx is rbNewFx whose recorded merge base carries two five-line
// files, then a base that edits their first lines and a change that edits
// their last: both paths overlap and the rebase is still clean.
func sobOverlapFx(t *testing.T) *bmFx {
	t.Helper()
	fx := rbNewFx(t)
	five := "1\n2\n3\n4\n5\n"
	fx.g.Git(fx.repo, "checkout", "-q", "main")
	for _, f := range []string{"multi.txt", "plain.txt"} {
		if fx.g.Err == nil {
			fx.g.Err = os.WriteFile(fx.repo+"/"+f, []byte(five), 0o644)
		}
	}
	fx.g.Git(fx.repo, "add", "multi.txt", "plain.txt")
	fx.g.Git(fx.repo, "commit", "-qm", "five-line files")
	fx.recorded = fx.g.Git(fx.repo, "rev-parse", "HEAD")
	for _, f := range []string{"multi.txt", "plain.txt"} {
		if fx.g.Err == nil {
			fx.g.Err = os.WriteFile(fx.repo+"/"+f, []byte("base\n2\n3\n4\n5\n"), 0o644)
		}
	}
	fx.g.Git(fx.repo, "commit", "-qam", "base edits line 1")
	fx.g.Git(fx.repo, "checkout", "-q", "demo")
	fx.g.Git(fx.repo, "rebase", "-q", fx.recorded)
	for _, f := range []string{"multi.txt", "plain.txt"} {
		if fx.g.Err == nil {
			fx.g.Err = os.WriteFile(fx.repo+"/"+f, []byte("1\n2\n3\n4\nchange\n"), 0o644)
		}
	}
	fx.g.Git(fx.repo, "commit", "-qam", "change edits line 5")
	return fx
}

// sobMovingStub writes a git that moves main one empty commit forward each
// time a rebase starts -- a fetch landing mid-sync -- and logs each move to
// dir/.moves; with once, only the first rebase moves it. Every other call
// execs the real git at once, and dir is read from the stub's own path, so
// the stub's body is one of two and each is assessed once (ExecFixtures).
func sobMovingStub(t *testing.T, dir string, once bool) string {
	t.Helper()
	stop := ""
	if once {
		stop = `: > "$D/.stop"`
	}
	guardtest.WriteExec(t, dir+"/stub/git", fmt.Sprintf(`#!/bin/sh
[ "${3-}" = rebase ] || exec %[1]q "$@"
REAL=%[1]q D="${0%%/stub/git}"
if [ "${4-}" != --continue ] && [ ! -e "$D/.stop" ]; then
  c="$("$REAL" -C "$2" commit-tree "$("$REAL" -C "$2" rev-parse 'main^{tree}')" -p main -m more)"
  "$REAL" -C "$2" update-ref refs/heads/main "$c"
  echo moved >> "$D/.moves"
  %[2]s
fi
exec "$REAL" "$@"
`, guardtest.FixtureGit, stop))
	return dir + "/stub"
}

// Cost: 2.00 s — ten cases of real git over their own repositories, two of
// them rebasing again onto a base the stub moves, up to the three-rebase cap
// the test proves.
// smoke: core — the rebase onto a moved base
func TestSyncOntoBase(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	type sobCase struct {
		label  string
		fx     func(t *testing.T) *bmFx
		args   func(fx *bmFx) []string // nil: {repo, main, recorded}
		stub   string                  // "", "always" or "once"
		code   int
		out    string // $WT the worktree, $TIP main's tip after the run, $ROOT the agents root
		moves  int    // rebases the stub saw
		verify func(t *testing.T, fx *bmFx)
	}
	cases := []sobCase{
		{label: "CLEAR is CLEAN", fx: rbNewFx, code: 0,
			out: "CLEAR: $WT — main has not moved since the recorded merge base\nCLEAN: $WT — nothing to rebase\n"},
		{label: "MOVED is REBASED", fx: sobOverlapFx, code: 0,
			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: multi.txt, plain.txt\n" +
				"CLEAR: $WT — main has not moved since the recorded merge base\n" +
				"REBASED: $WT — onto $TIP\n",
			verify: func(t *testing.T, fx *bmFx) {
				if got := rbRead(fx.repo + "/multi.txt"); got != "base\n2\n3\n4\nchange\n" {
					t.Errorf("multi.txt %q: want both sides kept", got)
				}
			}},
		{label: "a base that moves during the rebase is rebased onto again", fx: sobOverlapFx, stub: "once", code: 0, moves: 1,
			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: multi.txt, plain.txt\n" +
				"MOVED: $WT — 1 commits on main since the recorded merge base; no overlap with this change's paths\n" +
				"CLEAR: $WT — main has not moved since the recorded merge base\n" +
				"REBASED: $WT — onto $TIP\n"},
		{label: "the re-check loop is capped at 3 rebases", fx: sobOverlapFx, stub: "always", code: 2, moves: 3,
			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: multi.txt, plain.txt\n" +
				strings.Repeat("MOVED: $WT — 1 commits on main since the recorded merge base; no overlap with this change's paths\n", 3)},
		{label: "a conflict is left mid-rebase with exit 1", fx: func(t *testing.T) *bmFx {
			fx := rbNewFx(t)
			fx.advanceBase("shared.txt")
			fx.g.AppendLine(fx.repo+"/shared.txt", "demo")
			fx.g.Git(fx.repo, "commit", "-qam", "demo touches shared")
			fx.g.Write(fx.repo+"/spectre/changes/demo/tasks.md", "plan edited")
			return fx
		}, code: 1,
			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: shared.txt\nCONFLICT: $WT — onto $TIP; unmerged: shared.txt\n",
			verify: func(t *testing.T, fx *bmFx) {
				if _, err := os.Stat(fx.repo + "/.git/rebase-merge"); err != nil {
					t.Error("no rebase in progress: the conflict was not left for resolution")
				}
			}},
		{label: "a conflicted snapshot PNG or full app suite zip is named for recapture", fx: func(t *testing.T) *bmFx {
			fx := rbNewFx(t)
			paths := []string{"e2e/a.spec.ts-snapshots/x.png", "full-app-suite.zip", "logo.png", "shared.txt"}
			fx.g.Git(fx.repo, "checkout", "-q", "main")
			guardtest.Mkdir(t, fx.repo+"/e2e/a.spec.ts-snapshots")
			for _, f := range paths {
				fx.g.Write(fx.repo+"/"+f, "base side")
			}
			fx.g.Git(fx.repo, "add", "-A")
			fx.g.Git(fx.repo, "commit", "-qm", "base re-baselines")
			fx.g.Git(fx.repo, "checkout", "-q", "demo")
			guardtest.Mkdir(t, fx.repo+"/e2e/a.spec.ts-snapshots")
			for _, f := range paths {
				fx.g.Write(fx.repo+"/"+f, "change side")
			}
			fx.g.Git(fx.repo, "add", "-A")
			fx.g.Git(fx.repo, "commit", "-qm", "demo re-baselines")
			return fx
		}, code: 1,
			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: e2e/a.spec.ts-snapshots/x.png, full-app-suite.zip, logo.png, shared.txt\n" +
				"CONFLICT: $WT — onto $TIP; unmerged: e2e/a.spec.ts-snapshots/x.png, full-app-suite.zip, logo.png, shared.txt\n" +
				"RECAPTURE: e2e/a.spec.ts-snapshots/x.png\nRECAPTURE: full-app-suite.zip\n"},
		{label: "REFUSE exits 2", fx: rbNewFx, args: func(fx *bmFx) []string { return []string{fx.repo, "main", "-"} }, code: 2,
			out: "REFUSE: no merge base recorded for $WT — cannot tell whether the base has moved\n"},
		{label: "a worktree that is not a directory exits 2", fx: rbNewFx, args: func(fx *bmFx) []string { return []string{fx.repo + "/missing", "main", fx.recorded} }, code: 2},
		{label: "a missing argument is a usage error", fx: rbNewFx, args: func(fx *bmFx) []string { return []string{fx.repo, "main"} }, code: 2},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			fx := c.fx(t)
			if fx.g.Err != nil {
				t.Fatal(fx.g.Err)
			}
			stub := ""
			if c.stub != "" {
				stub = sobMovingStub(t, fx.dir, c.stub == "once")
			}
			args := []string{fx.repo, "main", fx.recorded}
			if c.args != nil {
				args = c.args(fx)
			}
			r := runGuard("sync-onto-base", args, sobEnv(t, fx.dir, root, stub))
			want := strings.NewReplacer("$WT", fx.repo, "$ROOT", root, "$TIP", fx.g.Git(fx.repo, "rev-parse", "main")).Replace(c.out)
			if r.Rc != c.code || r.Stdout != want {
				t.Fatalf("got exit %d stdout %q\nwant exit %d stdout %q\nstderr %s", r.Rc, r.Stdout, c.code, want, r.Err)
			}
			if c.code == 2 && r.Err == "" && !strings.HasPrefix(r.Stdout, "REFUSE:") {
				t.Error("exit 2 with no REFUSE line and nothing on stderr")
			}
			if moves := strings.Count(rbRead(fx.dir+"/.moves"), "moved"); moves != c.moves {
				t.Errorf("the stub saw %d rebases, want %d", moves, c.moves)
			}
			if c.verify != nil {
				c.verify(t, fx)
			}
		})
	}
}

func TestSyncOntoBaseResume(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root := t.TempDir()
	// conflicted is a sync stopped on a conflict over shared.txt; it returns the tip the sync rebased onto.
	conflicted := func(t *testing.T) (*bmFx, string) {
		t.Helper()
		fx := rbNewFx(t)
		fx.advanceBase("shared.txt")
		fx.g.AppendLine(fx.repo+"/shared.txt", "demo")
		fx.g.Git(fx.repo, "commit", "-qam", "demo touches shared")
		fx.g.Write(fx.repo+"/spectre/changes/demo/tasks.md", "plan edited")
		if fx.g.Err != nil {
			t.Fatal(fx.g.Err)
		}
		if r := runGuard("sync-onto-base", []string{fx.repo, "main", fx.recorded}, sobEnv(t, fx.dir, root, "")); r.Rc != 1 {
			t.Fatalf("setup: sync exit %d, want the conflict's 1\n%s%s", r.Rc, r.Stdout, r.Err)
		}
		return fx, fx.g.Git(fx.repo, "rev-parse", "main")
	}

	t.Run("resolved: re-checks, the plan edit untouched", func(t *testing.T) {
		t.Parallel()
		fx, onto := conflicted(t)
		fx.g.Write(fx.repo+"/shared.txt", "resolved")
		fx.g.Git(fx.repo, "add", "shared.txt")
		fx.g.Git(fx.repo, "-c", "core.editor=true", "rebase", "--continue")
		// The base moves again over a path this change also carries, so the
		// resumed re-check rebases once more.
		fx.g.Write(fx.repo+"/later.txt", "moved")
		fx.g.Git(fx.repo, "add", "later.txt")
		fx.g.Git(fx.repo, "commit", "-qm", "change adds later.txt")
		fx.advanceBase("later.txt")
		tip := fx.g.Git(fx.repo, "rev-parse", "main")
		if fx.g.Err != nil {
			t.Fatal(fx.g.Err)
		}
		r := runGuard("sync-onto-base", []string{"--resume", fx.repo, "main", onto}, sobEnv(t, fx.dir, root, ""))
		want := "MOVED: " + fx.repo + " — 1 commits on main since the recorded merge base; overlaps: later.txt\n" +
			"CLEAR: " + fx.repo + " — main has not moved since the recorded merge base\nREBASED: " + fx.repo + " — onto " + tip + "\n"
		if r.Rc != 0 || r.Stdout != want {
			t.Fatalf("got exit %d stdout %q\nwant 0 %q\nstderr %s", r.Rc, r.Stdout, want, r.Err)
		}
		if got := rbRead(fx.repo + "/spectre/changes/demo/tasks.md"); got != "plan edited\n" {
			t.Errorf("planning file %q, want the edit untouched", got)
		}
		if l := fx.g.Git(fx.repo, "stash", "list"); l != "" {
			t.Errorf("stash list %q, want nothing stashed", l)
		}
	})

	t.Run("refuses while the rebase is still in progress", func(t *testing.T) {
		t.Parallel()
		fx, onto := conflicted(t)
		r := runGuard("sync-onto-base", []string{"--resume", fx.repo, "main", onto}, sobEnv(t, fx.dir, root, ""))
		if r.Rc != 2 || r.Stdout != "" || !strings.Contains(r.Err, "still in progress") {
			t.Fatalf("got exit %d stdout %q stderr %q, want a refusal naming the rebase in progress", r.Rc, r.Stdout, r.Err)
		}
	})

	t.Run("requires <onto> to be an ancestor of HEAD", func(t *testing.T) {
		t.Parallel()
		fx := rbNewFx(t)
		fx.advanceBase("unrelated1.txt")
		onto := fx.g.Git(fx.repo, "rev-parse", "main")
		if fx.g.Err != nil {
			t.Fatal(fx.g.Err)
		}
		r := runGuard("sync-onto-base", []string{"--resume", fx.repo, "main", onto}, sobEnv(t, fx.dir, root, ""))
		if r.Rc != 2 || r.Stdout != "" || !strings.Contains(r.Err, "ancestor") {
			t.Fatalf("got exit %d stdout %q stderr %q, want a refusal naming the ancestry", r.Rc, r.Stdout, r.Err)
		}
	})
}
