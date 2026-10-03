package guard

import (
	"strings"
	"testing"
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
	t.Parallel()
	// The harness's repository: one empty commit. The parity rows': a.txt.
	empty, withA := t.TempDir()+"/empty", t.TempDir()+"/with-a"
	gitRun(t, "", "init", "-q", "-b", "main", "--template=", empty)
	gitRun(t, empty, "commit", "-q", "--allow-empty", "-m", "base")
	gitRun(t, "", "init", "-q", "-b", "main", "--template=", withA)
	writeFile(t, withA+"/a.txt", "a\n")
	gitRun(t, withA, "add", "a.txt")
	gitRun(t, withA, "commit", "-q", "-m", "init")

	const (
		sentinel = postMutationCheckSentinel + "\n"
		aDirty   = "1 .M N... 100644 100644 100644 78981922613b2afb6025042ff6bd878ac1994e85 78981922613b2afb6025042ff6bd878ac1994e85 a.txt"
	)
	for _, tc := range []struct {
		name          string
		base          string // "" is no repository at all
		missing       bool   // the worktree directory does not exist
		before, after func(t *testing.T, wt string)
		snap, check   libRes // check is run against snap's stdout as $(...) captured it
	}{
		{name: "clean tree matches its snapshot", base: empty,
			snap: libRes{sentinel, "", 0}, check: libRes{"", "", 0}},
		{name: "new stash entry is named and fails the check", base: empty,
			after: func(t *testing.T, wt string) {
				writeFile(t, wt+"/file.txt", "dirty\n")
				gitRun(t, wt, "stash", "push", "-q", "--include-untracked", "-m", "kan-448-test-residue")
			},
			snap:  libRes{sentinel, "", 0},
			check: libRes{"new stash entry: stash@{0}: On main: kan-448-test-residue\n", "", 1}},
		{name: "unexpected status line is named and fails the check", base: empty,
			after: func(t *testing.T, wt string) { writeFile(t, wt+"/stray.txt", "stray\n") },
			snap:  libRes{sentinel, "", 0},
			check: libRes{"unexpected status line: ? stray.txt\n", "", 1}},
		{name: "snapshot-recorded stash still present is not drift", base: empty,
			before: func(t *testing.T, wt string) {
				writeFile(t, wt+"/file.txt", "kept\n")
				gitRun(t, wt, "stash", "push", "-q", "--include-untracked", "-m", "kan-448-kept")
			},
			snap:  libRes{sentinel + "stash@{0}: On main: kan-448-kept\n", "", 0},
			check: libRes{"", "", 0}},
		{name: "clean tree", base: withA, snap: libRes{sentinel, "", 0}, check: libRes{"", "", 0}},
		{name: "dirty tracked file present at the snapshot", base: withA,
			before: func(t *testing.T, wt string) { writeFile(t, wt+"/a.txt", "dirty\n") },
			snap:   libRes{aDirty + "\n" + sentinel, "", 0}, check: libRes{"", "", 0}},
		{name: "tracked file dirtied after the snapshot", base: withA,
			after: func(t *testing.T, wt string) { writeFile(t, wt+"/a.txt", "dirty\n") },
			snap:  libRes{sentinel, "", 0}, check: libRes{"unexpected status line: " + aDirty + "\n", "", 1}},
		{name: "new untracked file", base: withA,
			after: func(t *testing.T, wt string) { writeFile(t, wt+"/new.txt", "x\n") },
			snap:  libRes{sentinel, "", 0}, check: libRes{"unexpected status line: ? new.txt\n", "", 1}},
		{name: "new stash entry, an older one renumbered", base: withA,
			before: func(t *testing.T, wt string) {
				writeFile(t, wt+"/a.txt", "one\n")
				gitRun(t, wt, "stash", "push", "-q", "-m", "old")
			},
			after: func(t *testing.T, wt string) {
				writeFile(t, wt+"/a.txt", "two\n")
				gitRun(t, wt, "stash", "push", "-q", "-m", "new")
			},
			snap: libRes{sentinel + "stash@{0}: On main: old\n", "", 0},
			check: libRes{"new stash entry: stash@{0}: On main: new\n" +
				"new stash entry: stash@{1}: On main: old\n", "", 1}},
		{name: "stash present at the snapshot and still present", base: withA,
			before: func(t *testing.T, wt string) {
				writeFile(t, wt+"/a.txt", "one\n")
				gitRun(t, wt, "stash", "push", "-q", "-m", "old")
			},
			snap: libRes{sentinel + "stash@{0}: On main: old\n", "", 0}, check: libRes{"", "", 0}},
		{name: "drift restored before the check", base: withA,
			before: func(t *testing.T, wt string) { writeFile(t, wt+"/a.txt", "dirty\n") },
			after:  func(t *testing.T, wt string) { writeFile(t, wt+"/a.txt", "a\n") },
			snap:   libRes{aDirty + "\n" + sentinel, "", 0}, check: libRes{"", "", 0}},
		{name: "stash and status drift together", base: withA,
			after: func(t *testing.T, wt string) {
				writeFile(t, wt+"/a.txt", "one\n")
				gitRun(t, wt, "stash", "push", "-q", "-m", "s")
				writeFile(t, wt+"/b.txt", "x\n")
			},
			snap: libRes{sentinel, "", 0},
			check: libRes{"new stash entry: stash@{0}: On main: s\n" +
				"unexpected status line: ? b.txt\n", "", 1}},
		{name: "missing worktree directory", missing: true,
			snap: libRes{sentinel, "fatal: cannot change to '<wt>': No such file or directory\n" +
				"fatal: cannot change to '<wt>': No such file or directory\n", 128},
			check: libRes{"", "fatal: cannot change to '<wt>': No such file or directory\n", 2}},
		{name: "not a git repository",
			snap: libRes{sentinel, "fatal: not a git repository (or any of the parent directories): .git\n" +
				"fatal: not a git repository (or any of the parent directories): .git\n", 128},
			check: libRes{"", "fatal: not a git repository (or any of the parent directories): .git\n", 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wt := t.TempDir() + "/work tree"
			switch {
			case tc.missing:
			case tc.base == "":
				mkdir(t, wt)
			default:
				mvCopyTree(t, tc.base, wt)
			}
			if tc.before != nil {
				tc.before(t, wt)
			}
			norm := strings.NewReplacer(wt, "<wt>")
			snap := goSnapshot(wt)
			if got := (libRes{snap.stdout, norm.Replace(snap.stderr), snap.status}); got != tc.snap {
				t.Errorf("snapshotTreeState\n got: %#v\nwant: %#v", got, tc.snap)
			}
			if tc.after != nil {
				tc.after(t, wt)
			}
			check := goCheck(wt, strings.TrimRight(snap.stdout, "\n"))
			if got := (libRes{check.stdout, norm.Replace(check.stderr), check.status}); got != tc.check {
				t.Errorf("checkTreeRestored\n got: %#v\nwant: %#v", got, tc.check)
			}
		})
	}
}

func goSnapshot(wt string) libRes {
	var errb strings.Builder
	out, status := snapshotTreeState(wt, &errb)
	return libRes{out, errb.String(), status}
}

func goCheck(wt, snapshot string) libRes {
	var errb strings.Builder
	findings, status := checkTreeRestored(wt, snapshot, &errb)
	var out string
	for _, f := range findings {
		out += f + "\n"
	}
	return libRes{out, errb.String(), status}
}
