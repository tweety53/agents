package guard

import (
	"bytes"
	"errors"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCannotAnswerCodes(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]int{
		"run-reproducer":           4,
		"mutate-and-verify":        4,
		"break-and-prove":          4,
		"check-hand-notes-in-step": 1,
		"check-installed-rules":    1,
		"check-task-commit-fields": 2,
		"check-cleanup-complete":   2,
		"no-such-guard":            2,
		"":                         2,
	} {
		if got := CannotAnswer(name); got != want {
			t.Errorf("CannotAnswer(%q) = %d, want %d", name, got, want)
		}
	}
}

// TestRunSiblingSelfExec is not parallel: it clears siblingExe, which the
// package's TestMain points at the built flow-guard, while no other test runs.
func TestRunSiblingSelfExec(t *testing.T) {
	saved := siblingExe
	siblingExe = ""
	t.Cleanup(func() { siblingExe = saved })
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	cmd := runSibling(Env{Dir: dir}, "check-task-commit-fields", "a", "b")
	if cmd.Path != self || !slices.Equal(cmd.Args, []string{self, "check-task-commit-fields", "a", "b"}) || cmd.Dir != dir {
		t.Fatalf("path %q args %q dir %q; want %q running check-task-commit-fields a b in %q", cmd.Path, cmd.Args, cmd.Dir, self, dir)
	}
	// The real exec: this test binary dispatches a guard name as flow-guard
	// does, so the sibling answers with its own usage refusal.
	cmd = runSibling(Env{Dir: dir}, "check-task-commit-fields")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 2 || !strings.Contains(stderr.String(), "usage: flow-guard check-task-commit-fields") {
		t.Fatalf("the self-exec'd sibling: %v, stderr %q; want exit 2 and its usage", err, stderr.String())
	}
}

// cwdFixture is a repository whose main checkout carries scripts/flow-guard,
// a second worktree of it at that commit, and one at a commit without it.
func cwdFixture(t *testing.T) (main, wt, bare string) {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	main, wt, bare = dir+"/main", dir+"/wt", dir+"/bare"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", main)
	g.git(main, "commit", "-q", "--allow-empty", "-m", "base")
	writeExec(t, main+"/scripts/flow-guard", "#!/bin/sh\n")
	g.git(main, "add", "scripts/flow-guard")
	g.git(main, "commit", "-q", "-m", "bootstrap")
	g.git(main, "worktree", "add", "-q", "--detach", wt, "HEAD")
	g.git(main, "worktree", "add", "-q", "--detach", bare, "HEAD~1")
	if g.err != nil {
		t.Fatal(g.err)
	}
	return main, wt, bare
}

func cwdEnv(root, dir string) Env {
	return Env{Dir: dir, Getenv: func(k string) string {
		if k == "FLOW_GUARD_REPO_ROOT" {
			return root
		}
		return os.Getenv(k)
	}}
}

// TestCwdCheckoutReexec ports scripts/test-lib-flow-guard.sh's case 15: a
// self-judging guard called from another worktree of its checkout's
// repository answers from that worktree's own scripts/flow-guard, and from
// anywhere else it answers itself.
func TestCwdCheckoutReexec(t *testing.T) {
	t.Parallel()
	main, wt, bare := cwdFixture(t)
	if got, ok := CwdCheckout(cwdEnv(main, wt), "check-plan-shape"); !ok || got != wt+"/scripts/flow-guard" {
		t.Errorf("from another worktree: %q %v; want %s/scripts/flow-guard", got, ok, wt)
	}
	if got, ok := CwdCheckout(cwdEnv(main, wt+"/scripts"), "check-plan-shape"); !ok || got != wt+"/scripts/flow-guard" {
		t.Errorf("from inside another worktree: %q %v; want %s/scripts/flow-guard", got, ok, wt)
	}

	other, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var g fxGit
	g.git("", "init", "-q", "-b", "main", other)
	writeExec(t, other+"/scripts/flow-guard", "#!/bin/sh\n")
	if g.err != nil {
		t.Fatal(g.err)
	}
	for _, c := range []struct{ what, root, dir string }{
		{"its own checkout", main, main},
		{"a directory in no repository", main, t.TempDir()},
		{"a different repository carrying scripts/flow-guard", main, other},
		{"a worktree without scripts/flow-guard", main, bare},
		{"no FLOW_GUARD_REPO_ROOT", "", wt},
	} {
		if got, ok := CwdCheckout(cwdEnv(c.root, c.dir), "check-plan-shape"); ok {
			t.Errorf("from %s: re-exec'd %q; want no re-exec", c.what, got)
		}
	}
}

// TestCwdCheckoutGuardSet ports case 16: exactly the guards whose shim
// called flow_guard_cwd_checkout re-exec, and no other.
func TestCwdCheckoutGuardSet(t *testing.T) {
	t.Parallel()
	want := []string{
		"check-ask-silence", "check-dev-stack-fresh", "check-dispatch-paragraphs", "check-guard-calls",
		"check-installed-citations", "check-markdown-integrity", "check-mutation-reproducer-pin", "check-no-bash",
		"check-normative-inventory", "check-plan-provenance", "check-plan-shape", "check-python-suppressions",
		"check-references", "check-spec-reach", "check-stage-mark-calls", "check-task-build-green",
		"check-uitest-overrides", "check-verbatim-moves", "plan-dispatch-bundles",
	}
	if got := slices.Sorted(maps.Keys(selfJudging)); !slices.Equal(got, want) {
		t.Fatalf("selfJudging = %q, want %q", got, want)
	}
	main, wt, _ := cwdFixture(t)
	for _, name := range []string{"check-cleanup-complete", "run-reproducer", "integrate-change"} {
		if got, ok := CwdCheckout(cwdEnv(main, wt), name); ok {
			t.Errorf("%s is not self-judging but re-exec'd %q", name, got)
		}
	}
	for _, name := range want {
		if _, ok := CwdCheckout(cwdEnv(main, wt), name); !ok {
			t.Errorf("%s is self-judging but did not re-exec", name)
		}
	}
}
