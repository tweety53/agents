package guard

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// sobEnv runs the guard with root as the agents repository its shim exports,
// and stub's directory first on PATH when stub is not "".
func sobEnv(dir, root, stub string) Env {
	get := os.Getenv
	if stub != "" {
		get = pathEnv(stub)
	}
	return Env{Dir: dir, Getenv: func(k string) string {
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
	fx.g.git(fx.repo, "checkout", "-q", "main")
	for _, f := range []string{"multi.txt", "plain.txt"} {
		if fx.g.err == nil {
			fx.g.err = os.WriteFile(fx.repo+"/"+f, []byte(five), 0o644)
		}
	}
	fx.g.git(fx.repo, "add", "multi.txt", "plain.txt")
	fx.g.git(fx.repo, "commit", "-qm", "five-line files")
	fx.recorded = fx.g.git(fx.repo, "rev-parse", "HEAD")
	for _, f := range []string{"multi.txt", "plain.txt"} {
		if fx.g.err == nil {
			fx.g.err = os.WriteFile(fx.repo+"/"+f, []byte("base\n2\n3\n4\n5\n"), 0o644)
		}
	}
	fx.g.git(fx.repo, "commit", "-qam", "base edits line 1")
	fx.g.git(fx.repo, "checkout", "-q", "demo")
	fx.g.git(fx.repo, "rebase", "-q", fx.recorded)
	for _, f := range []string{"multi.txt", "plain.txt"} {
		if fx.g.err == nil {
			fx.g.err = os.WriteFile(fx.repo+"/"+f, []byte("1\n2\n3\n4\nchange\n"), 0o644)
		}
	}
	fx.g.git(fx.repo, "commit", "-qam", "change edits line 5")
	return fx
}

// sobMovingStub writes a git that moves main one empty commit forward each
// time a rebase starts -- a fetch landing mid-sync -- and logs each move to
// dir/.moves; with once, only the first rebase moves it.
func sobMovingStub(t *testing.T, dir string, once bool) string {
	t.Helper()
	stop := ""
	if once {
		stop = `: > "$D/.stop"`
	}
	writeExec(t, dir+"/stub/git", fmt.Sprintf(`#!/usr/bin/env bash
REAL=%q D=%q
if [ "${3-}" = rebase ] && [ "${4-}" != --continue ] && [ ! -e "$D/.stop" ]; then
  c="$("$REAL" -C "$2" commit-tree "$("$REAL" -C "$2" rev-parse 'main^{tree}')" -p main -m more)"
  "$REAL" -C "$2" update-ref refs/heads/main "$c"
  echo moved >> "$D/.moves"
  %s
fi
exec "$REAL" "$@"
`, fixtureGit, dir, stop))
	return dir + "/stub"
}

func TestSyncOntoBase(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeExec(t, root+"/scripts/test-multi.sh", "#!/usr/bin/env bash\n")
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
		{label: "MOVED is REBASED with a guard-test line per overlap path", fx: sobOverlapFx, code: 0,
			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: multi.txt, plain.txt\n" +
				"CLEAR: $WT — main has not moved since the recorded merge base\n" +
				"GUARD-TEST: multi.txt — $ROOT/scripts/test-multi.sh\nNO-GUARD-TEST: plain.txt\nREBASED: $WT — onto $TIP\n",
			verify: func(t *testing.T, fx *bmFx) {
				if got := rbRead(fx.repo + "/multi.txt"); got != "base\n2\n3\n4\nchange\n" {
					t.Errorf("multi.txt %q: want both sides kept", got)
				}
			}},
		{label: "a base that moves during the rebase is rebased onto again", fx: sobOverlapFx, stub: "once", code: 0, moves: 1,
			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: multi.txt, plain.txt\n" +
				"MOVED: $WT — 1 commits on main since the recorded merge base; no overlap with this change's paths\n" +
				"CLEAR: $WT — main has not moved since the recorded merge base\n" +
				"GUARD-TEST: multi.txt — $ROOT/scripts/test-multi.sh\nNO-GUARD-TEST: plain.txt\nREBASED: $WT — onto $TIP\n"},
		{label: "the re-check loop is capped at 3 rebases", fx: sobOverlapFx, stub: "always", code: 2, moves: 3,
			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: multi.txt, plain.txt\n" +
				strings.Repeat("MOVED: $WT — 1 commits on main since the recorded merge base; no overlap with this change's paths\n", 3)},
		{label: "a conflict is left mid-rebase with exit 1", fx: func(t *testing.T) *bmFx {
			fx := rbNewFx(t)
			fx.advanceBase("shared.txt")
			fx.g.appendLine(fx.repo+"/shared.txt", "demo")
			fx.g.git(fx.repo, "commit", "-qam", "demo touches shared")
			fx.g.write(fx.repo+"/spectre/changes/demo/tasks.md", "plan edited")
			return fx
		}, code: 1,
			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: shared.txt\nCONFLICT: $WT — onto $TIP; unmerged: shared.txt\n",
			verify: func(t *testing.T, fx *bmFx) {
				if _, err := os.Stat(fx.repo + "/.git/rebase-merge"); err != nil {
					t.Error("no rebase in progress: the conflict was not left for resolution")
				}
				if top := fx.g.git(fx.repo, "stash", "list", "--format=%gs"); !strings.Contains(top, apaMarker) {
					t.Errorf("stash list %q, want the aside kept", top)
				}
			}},
		{label: "REFUSE exits 2", fx: rbNewFx, args: func(fx *bmFx) []string { return []string{fx.repo, "main", "-"} }, code: 2,
			out: "REFUSE: no merge base recorded for $WT — cannot tell whether the base has moved\n"},
		{label: "a worktree that is not a directory exits 2", fx: rbNewFx, args: func(fx *bmFx) []string { return []string{fx.repo + "/missing", "main", fx.recorded} }, code: 2},
		{label: "a missing argument is a usage error", fx: rbNewFx, args: func(fx *bmFx) []string { return []string{fx.repo, "main"} }, code: 2},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			fx := c.fx(t)
			if fx.g.err != nil {
				t.Fatal(fx.g.err)
			}
			stub := ""
			if c.stub != "" {
				stub = sobMovingStub(t, fx.dir, c.stub == "once")
			}
			args := []string{fx.repo, "main", fx.recorded}
			if c.args != nil {
				args = c.args(fx)
			}
			r := runGuard("sync-onto-base", args, sobEnv(fx.dir, root, stub))
			want := strings.NewReplacer("$WT", fx.repo, "$ROOT", root, "$TIP", fx.g.git(fx.repo, "rev-parse", "main")).Replace(c.out)
			if r.rc != c.code || r.stdout != want {
				t.Fatalf("got exit %d stdout %q\nwant exit %d stdout %q\nstderr %s", r.rc, r.stdout, c.code, want, r.err)
			}
			if c.code == 2 && r.err == "" && !strings.HasPrefix(r.stdout, "REFUSE:") {
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
	t.Parallel()
	root := t.TempDir()
	writeExec(t, root+"/scripts/test-later.sh", "#!/usr/bin/env bash\n")
	// conflicted is a sync stopped on a conflict over shared.txt, the aside
	// held; it returns the tip the sync rebased onto.
	conflicted := func(t *testing.T) (*bmFx, string) {
		t.Helper()
		fx := rbNewFx(t)
		fx.advanceBase("shared.txt")
		fx.g.appendLine(fx.repo+"/shared.txt", "demo")
		fx.g.git(fx.repo, "commit", "-qam", "demo touches shared")
		fx.g.write(fx.repo+"/spectre/changes/demo/tasks.md", "plan edited")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		if r := runGuard("sync-onto-base", []string{fx.repo, "main", fx.recorded}, sobEnv(fx.dir, root, "")); r.rc != 1 {
			t.Fatalf("setup: sync exit %d, want the conflict's 1\n%s%s", r.rc, r.stdout, r.err)
		}
		return fx, fx.g.git(fx.repo, "rev-parse", "main")
	}

	t.Run("resolved: restores the aside and re-checks, with no guard-test lines", func(t *testing.T) {
		t.Parallel()
		fx, onto := conflicted(t)
		fx.g.write(fx.repo+"/shared.txt", "resolved")
		fx.g.git(fx.repo, "add", "shared.txt")
		fx.g.git(fx.repo, "-c", "core.editor=true", "rebase", "--continue")
		// The base moves again over a path this change also carries, so the
		// resumed re-check rebases once more with an overlap that has a
		// guard test -- and still prints no guard-test line.
		fx.g.write(fx.repo+"/later.txt", "moved")
		fx.g.git(fx.repo, "add", "later.txt")
		fx.g.git(fx.repo, "commit", "-qm", "change adds later.txt")
		fx.advanceBase("later.txt")
		tip := fx.g.git(fx.repo, "rev-parse", "main")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		r := runGuard("sync-onto-base", []string{"--resume", fx.repo, "main", onto}, sobEnv(fx.dir, root, ""))
		want := "MOVED: " + fx.repo + " — 1 commits on main since the recorded merge base; overlaps: later.txt\n" +
			"CLEAR: " + fx.repo + " — main has not moved since the recorded merge base\nREBASED: " + fx.repo + " — onto " + tip + "\n"
		if r.rc != 0 || r.stdout != want {
			t.Fatalf("got exit %d stdout %q\nwant 0 %q\nstderr %s", r.rc, r.stdout, want, r.err)
		}
		if got := rbRead(fx.repo + "/spectre/changes/demo/tasks.md"); got != "plan edited\n" {
			t.Errorf("planning file %q, want the aside restored", got)
		}
		if l := fx.g.git(fx.repo, "stash", "list"); l != "" {
			t.Errorf("stash list %q, want the aside popped", l)
		}
	})

	t.Run("refuses while the rebase is still in progress", func(t *testing.T) {
		t.Parallel()
		fx, onto := conflicted(t)
		r := runGuard("sync-onto-base", []string{"--resume", fx.repo, "main", onto}, sobEnv(fx.dir, root, ""))
		if r.rc != 2 || r.stdout != "" || !strings.Contains(r.err, "still in progress") {
			t.Fatalf("got exit %d stdout %q stderr %q, want a refusal naming the rebase in progress", r.rc, r.stdout, r.err)
		}
		if top := fx.g.git(fx.repo, "stash", "list", "--format=%gs"); !strings.Contains(top, apaMarker) {
			t.Errorf("stash list %q: the aside was restored mid-rebase", top)
		}
	})

	// foreign pushes a marker stash as another worktree's aside would: the
	// stash list is shared by every worktree of the repository.
	foreign := func(fx *bmFx) string {
		if fx.g.err == nil {
			fx.g.err = os.MkdirAll(fx.repo+"/spectre/changes/other", 0o755)
		}
		fx.g.write(fx.repo+"/spectre/changes/other/tasks.md", "other worktree's plan")
		fx.g.git(fx.repo, "stash", "push", "--include-untracked", "-m", apaMarker+": planning paths set aside for a rebase", "--", "spectre/changes/other")
		return fx.g.git(fx.repo, "rev-parse", "stash@{0}")
	}

	t.Run("a clean sync's resume never pops another worktree's aside", func(t *testing.T) {
		t.Parallel()
		fx := rbNewFx(t)
		fx.advanceBase("shared.txt")
		fx.g.appendLine(fx.repo+"/shared.txt", "demo")
		fx.g.git(fx.repo, "commit", "-qam", "demo touches shared")
		other := foreign(fx)
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		if r := runGuard("sync-onto-base", []string{fx.repo, "main", fx.recorded}, sobEnv(fx.dir, root, "")); r.rc != 1 {
			t.Fatalf("setup: sync exit %d, want the conflict's 1\n%s%s", r.rc, r.stdout, r.err)
		}
		onto := fx.g.git(fx.repo, "rev-parse", "main")
		fx.g.write(fx.repo+"/shared.txt", "resolved")
		fx.g.git(fx.repo, "add", "shared.txt")
		fx.g.git(fx.repo, "-c", "core.editor=true", "rebase", "--continue")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		r := runGuard("sync-onto-base", []string{"--resume", fx.repo, "main", onto}, sobEnv(fx.dir, root, ""))
		if r.rc != 0 {
			t.Fatalf("resume exit %d, want 0\nstdout %q\nstderr %s", r.rc, r.stdout, r.err)
		}
		if top := fx.g.git(fx.repo, "rev-parse", "stash@{0}"); top != other {
			t.Errorf("top stash %s, want the other worktree's %s left in place", top, other)
		}
		if _, err := os.Stat(fx.repo + "/spectre/changes/other/tasks.md"); err == nil {
			t.Errorf("the other worktree's planning file was popped into this one")
		}
	})

	t.Run("resume refuses when another aside sits on top of its own", func(t *testing.T) {
		t.Parallel()
		fx, onto := conflicted(t)
		fx.g.write(fx.repo+"/shared.txt", "resolved")
		fx.g.git(fx.repo, "add", "shared.txt")
		fx.g.git(fx.repo, "-c", "core.editor=true", "rebase", "--continue")
		other := foreign(fx)
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		r := runGuard("sync-onto-base", []string{"--resume", fx.repo, "main", onto}, sobEnv(fx.dir, root, ""))
		if r.rc != 2 || !strings.Contains(r.err, "not the top stash") {
			t.Fatalf("got exit %d stderr %q, want a refusal naming its own aside", r.rc, r.err)
		}
		if top := fx.g.git(fx.repo, "rev-parse", "stash@{0}"); top != other {
			t.Errorf("top stash %s, want the other worktree's %s untouched", top, other)
		}
	})

	t.Run("requires <onto> to be an ancestor of HEAD", func(t *testing.T) {
		t.Parallel()
		fx := rbNewFx(t)
		fx.advanceBase("unrelated1.txt")
		onto := fx.g.git(fx.repo, "rev-parse", "main")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		r := runGuard("sync-onto-base", []string{"--resume", fx.repo, "main", onto}, sobEnv(fx.dir, root, ""))
		if r.rc != 2 || r.stdout != "" || !strings.Contains(r.err, "ancestor") {
			t.Fatalf("got exit %d stdout %q stderr %q, want a refusal naming the ancestry", r.rc, r.stdout, r.err)
		}
	})
}
