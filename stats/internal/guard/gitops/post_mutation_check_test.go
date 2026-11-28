package gitops

import (
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// TestPostMutationCheck pins snapshotTreeState and checkTreeRestored against
// fixed expected output: the stdout, stderr and status
// scripts/lib/post-mutation-check.sh printed for the same fixture at
// 9cd35da8, before that library was deleted (delete-post-mutation-check-lib).
// The first four rows are scripts/test-lib-post-mutation-check.sh's cases,
// named after its ok: labels; the rest are the rows
// TestPostMutationCheckParity, deleted with the library, compared against
// it. <wt> stands for the row's worktree path, which carries a space.
func TestPostMutationCheck(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	// The harness's repository: one empty commit. The parity rows': a.txt.
	empty, withA := t.TempDir()+"/empty", t.TempDir()+"/with-a"
	guardtest.GitRun(t, "", "init", "-q", "-b", "main", "--template=", empty)
	guardtest.GitRun(t, empty, "commit", "-q", "--allow-empty", "-m", "base")
	guardtest.GitRun(t, "", "init", "-q", "-b", "main", "--template=", withA)
	guardtest.WriteFile(t, withA+"/a.txt", "a\n")
	guardtest.GitRun(t, withA, "add", "a.txt")
	guardtest.GitRun(t, withA, "commit", "-q", "-m", "init")

	const (
		sentinel = postMutationCheckSentinel + "\n"
		aDirty   = "1 .M N... 100644 100644 100644 78981922613b2afb6025042ff6bd878ac1994e85 78981922613b2afb6025042ff6bd878ac1994e85 a.txt"
	)
	for _, tc := range []struct {
		name          string
		base          string // "" is no repository at all
		missing       bool   // the worktree directory does not exist
		before, after func(t *testing.T, wt string)
		snap, check   guardtest.LibRes // check is run against snap's stdout as $(...) captured it
	}{
		{name: "clean tree matches its snapshot", base: empty,
			snap: guardtest.LibRes{Stdout: sentinel, Stderr: "", Status: 0}, check: guardtest.LibRes{Stdout: "", Stderr: "", Status: 0}},
		{name: "new stash entry is named and fails the check", base: empty,
			after: func(t *testing.T, wt string) {
				guardtest.WriteFile(t, wt+"/file.txt", "dirty\n")
				guardtest.GitRun(t, wt, "stash", "push", "-q", "--include-untracked", "-m", "kan-448-test-residue")
			},
			snap:  guardtest.LibRes{Stdout: sentinel, Stderr: "", Status: 0},
			check: guardtest.LibRes{Stdout: "new stash entry: stash@{0}: On main: kan-448-test-residue\n", Stderr: "", Status: 1}},
		{name: "unexpected status line is named and fails the check", base: empty,
			after: func(t *testing.T, wt string) { guardtest.WriteFile(t, wt+"/stray.txt", "stray\n") },
			snap:  guardtest.LibRes{Stdout: sentinel, Stderr: "", Status: 0},
			check: guardtest.LibRes{Stdout: "unexpected status line: ? stray.txt\n", Stderr: "", Status: 1}},
		{name: "snapshot-recorded stash still present is not drift", base: empty,
			before: func(t *testing.T, wt string) {
				guardtest.WriteFile(t, wt+"/file.txt", "kept\n")
				guardtest.GitRun(t, wt, "stash", "push", "-q", "--include-untracked", "-m", "kan-448-kept")
			},
			snap:  guardtest.LibRes{Stdout: sentinel + "stash@{0}: On main: kan-448-kept\n", Stderr: "", Status: 0},
			check: guardtest.LibRes{Stdout: "", Stderr: "", Status: 0}},
		{name: "clean tree", base: withA, snap: guardtest.LibRes{Stdout: sentinel, Stderr: "", Status: 0}, check: guardtest.LibRes{Stdout: "", Stderr: "", Status: 0}},
		{name: "dirty tracked file present at the snapshot", base: withA,
			before: func(t *testing.T, wt string) { guardtest.WriteFile(t, wt+"/a.txt", "dirty\n") },
			snap:   guardtest.LibRes{Stdout: aDirty + "\n" + sentinel, Stderr: "", Status: 0}, check: guardtest.LibRes{Stdout: "", Stderr: "", Status: 0}},
		{name: "tracked file dirtied after the snapshot", base: withA,
			after: func(t *testing.T, wt string) { guardtest.WriteFile(t, wt+"/a.txt", "dirty\n") },
			snap:  guardtest.LibRes{Stdout: sentinel, Stderr: "", Status: 0}, check: guardtest.LibRes{Stdout: "unexpected status line: " + aDirty + "\n", Stderr: "", Status: 1}},
		{name: "new untracked file", base: withA,
			after: func(t *testing.T, wt string) { guardtest.WriteFile(t, wt+"/new.txt", "x\n") },
			snap:  guardtest.LibRes{Stdout: sentinel, Stderr: "", Status: 0}, check: guardtest.LibRes{Stdout: "unexpected status line: ? new.txt\n", Stderr: "", Status: 1}},
		{name: "new stash entry, an older one renumbered", base: withA,
			before: func(t *testing.T, wt string) {
				guardtest.WriteFile(t, wt+"/a.txt", "one\n")
				guardtest.GitRun(t, wt, "stash", "push", "-q", "-m", "old")
			},
			after: func(t *testing.T, wt string) {
				guardtest.WriteFile(t, wt+"/a.txt", "two\n")
				guardtest.GitRun(t, wt, "stash", "push", "-q", "-m", "new")
			},
			snap: guardtest.LibRes{Stdout: sentinel + "stash@{0}: On main: old\n", Stderr: "", Status: 0},
			check: guardtest.LibRes{Stdout: "new stash entry: stash@{0}: On main: new\n" +
				"new stash entry: stash@{1}: On main: old\n", Stderr: "", Status: 1}},
		{name: "stash present at the snapshot and still present", base: withA,
			before: func(t *testing.T, wt string) {
				guardtest.WriteFile(t, wt+"/a.txt", "one\n")
				guardtest.GitRun(t, wt, "stash", "push", "-q", "-m", "old")
			},
			snap: guardtest.LibRes{Stdout: sentinel + "stash@{0}: On main: old\n", Stderr: "", Status: 0}, check: guardtest.LibRes{Stdout: "", Stderr: "", Status: 0}},
		{name: "drift restored before the check", base: withA,
			before: func(t *testing.T, wt string) { guardtest.WriteFile(t, wt+"/a.txt", "dirty\n") },
			after:  func(t *testing.T, wt string) { guardtest.WriteFile(t, wt+"/a.txt", "a\n") },
			snap:   guardtest.LibRes{Stdout: aDirty + "\n" + sentinel, Stderr: "", Status: 0}, check: guardtest.LibRes{Stdout: "", Stderr: "", Status: 0}},
		{name: "stash and status drift together", base: withA,
			after: func(t *testing.T, wt string) {
				guardtest.WriteFile(t, wt+"/a.txt", "one\n")
				guardtest.GitRun(t, wt, "stash", "push", "-q", "-m", "s")
				guardtest.WriteFile(t, wt+"/b.txt", "x\n")
			},
			snap: guardtest.LibRes{Stdout: sentinel, Stderr: "", Status: 0},
			check: guardtest.LibRes{Stdout: "new stash entry: stash@{0}: On main: s\n" +
				"unexpected status line: ? b.txt\n", Stderr: "", Status: 1}},
		{name: "missing worktree directory", missing: true,
			snap: guardtest.LibRes{Stdout: sentinel, Stderr: "fatal: cannot change to '<wt>': No such file or directory\n" +
				"fatal: cannot change to '<wt>': No such file or directory\n", Status: 128},
			check: guardtest.LibRes{Stdout: "", Stderr: "fatal: cannot change to '<wt>': No such file or directory\n", Status: 2}},
		{name: "not a git repository",
			snap: guardtest.LibRes{Stdout: sentinel, Stderr: "fatal: not a git repository (or any of the parent directories): .git\n" +
				"fatal: not a git repository (or any of the parent directories): .git\n", Status: 128},
			check: guardtest.LibRes{Stdout: "", Stderr: "fatal: not a git repository (or any of the parent directories): .git\n", Status: 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wt := t.TempDir() + "/work tree"
			switch {
			case tc.missing:
			case tc.base == "":
				guardtest.Mkdir(t, wt)
			default:
				guardtest.CopyTreeMode(t, tc.base, wt)
			}
			if tc.before != nil {
				tc.before(t, wt)
			}
			norm := strings.NewReplacer(wt, "<wt>")
			snap := goSnapshot(wt)
			if got := (guardtest.LibRes{Stdout: snap.Stdout, Stderr: norm.Replace(snap.Stderr), Status: snap.Status}); got != tc.snap {
				t.Errorf("snapshotTreeState\n got: %#v\nwant: %#v", got, tc.snap)
			}
			if tc.after != nil {
				tc.after(t, wt)
			}
			check := goCheck(wt, strings.TrimRight(snap.Stdout, "\n"))
			if got := (guardtest.LibRes{Stdout: check.Stdout, Stderr: norm.Replace(check.Stderr), Status: check.Status}); got != tc.check {
				t.Errorf("checkTreeRestored\n got: %#v\nwant: %#v", got, tc.check)
			}
		})
	}
}

func goSnapshot(wt string) guardtest.LibRes {
	var errb strings.Builder
	out, status := SnapshotTreeState(wt, &errb)
	return guardtest.LibRes{Stdout: out, Stderr: errb.String(), Status: status}
}

func goCheck(wt, snapshot string) guardtest.LibRes {
	var errb strings.Builder
	findings, status := CheckTreeRestored(wt, snapshot, &errb)
	var out string
	for _, f := range findings {
		out += f + "\n"
	}
	return guardtest.LibRes{Stdout: out, Stderr: errb.String(), Status: status}
}
