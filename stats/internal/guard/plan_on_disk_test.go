package guard

import (
	"path/filepath"
	"slices"
	"testing"
)

// TestFindCanonicalIgnoresADropCommit pins that integrate-change reads the
// change's plan from the worktree only: a branch whose log holds
// `chore(spectre): drop <name>`, with the plan in the drop commit's parent
// and none on disk, has no canonical worktree.
func TestFindCanonicalIgnoresADropCommit(t *testing.T) {
	t.Parallel()
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitRun(t, "", "init", "-q", "-b", "main", wt)
	writeFile(t, wt+"/spectre/changes/demo/tasks.md", "# Tasks\n\n- [x] 1. Do the work\n")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-q", "-m", "chore(spectre): plan demo")
	gitRun(t, wt, "rm", "-r", "-q", "spectre/changes/demo")
	gitRun(t, wt, "commit", "-q", "-m", "chore(spectre): drop demo")

	c := &ic{wt: []string{wt}, name: "demo"}
	if c.findCanonical() {
		t.Fatalf("findCanonical found %q from history; want nothing with no plan on disk", c.canon)
	}

	writeFile(t, wt+"/spectre/changes/demo/tasks.md", "# Tasks\n\n- [x] 1. Do the work\n")
	if !c.findCanonical() || c.canon != wt || c.spec != "spectre" {
		t.Fatalf("plan on disk: canon %q spec %q, want %q spectre", c.canon, c.spec, wt)
	}
}

// TestMergeOrderMissingLinkIsSingleRepo pins that a canonical worktree with no
// link.md on disk is the single-repository case: every worktree in the set's
// own order, with no history read standing in for the missing file.
func TestMergeOrderMissingLinkIsSingleRepo(t *testing.T) {
	t.Parallel()
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a, b := top+"/a", top+"/b"
	gitRun(t, "", "init", "-q", "-b", "main", a)
	writeFile(t, a+"/spectre/changes/demo/link.md", "## Merge order\n\n1. `b`\n2. `.`\n")
	writeFile(t, a+"/spectre/peers", "b ../b\n")
	gitRun(t, a, "add", "-A")
	gitRun(t, a, "commit", "-q", "-m", "chore(spectre): plan demo")
	gitRun(t, a, "rm", "-r", "-q", "spectre/changes/demo")
	gitRun(t, a, "commit", "-q", "-m", "chore(spectre): drop demo")

	c := &ic{wt: []string{a, b}, name: "demo", canon: a, spec: "spectre"}
	c.mergeOrder()
	if want := []int{0, 1}; !slices.Equal(c.order, want) {
		t.Fatalf("order %v, want %v: a missing link.md is the single-repository order", c.order, want)
	}
}
