package gitops

import (
	"os"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// rcwFlowStub puts a `flow` on PATH whose `change file` exits rc, so the save
// gate reads flowd's answer from it.
func rcwFlowStub(t *testing.T, fx *rcwFx, rc string) func(*core.Env) {
	t.Helper()
	bin := fx.dir + "/bin"
	guardtest.WriteExec(t, bin+"/flow", "#!/usr/bin/env bash\n[ \"$1 $2\" = \"change file\" ] && exit "+rc+"\nexit 0\n")
	return func(env *core.Env) {
		inner := env.Getenv
		env.Getenv = func(k string) string {
			if k == "PATH" {
				return bin + ":" + os.Getenv("PATH")
			}
			return inner(k)
		}
	}
}

// TestRemoveChangeWorktreesRefusesBeforeTheSave: a worktree holding the
// change's plan, which flowd has not saved, is refused exit 1 with nothing
// removed — the plan's only copy would go with the worktree.
// smoke: core — no worktree is removed while it holds the plan's only copy
func TestRemoveChangeWorktreesRefusesBeforeTheSave(t *testing.T) {
	t.Parallel()
	fx := rcwNewFx(t)
	guardtest.WriteFile(t, fx.repo+"/.git/info/exclude", "spectre/changes/\n.superpowers/\n")
	guardtest.WriteFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
	code, out, errb := fx.run(t, rcwFlowStub(t, fx, "1"))
	if code != 1 || !strings.Contains(errb, "flow change save demo has not run") {
		t.Fatalf("exit %d, want 1 naming the missing save\nstdout:\n%s\nstderr:\n%s", code, out, errb)
	}
	if out != "" {
		t.Errorf("stdout %q, want nothing — the gate precedes every check", out)
	}
	fx.assertUntouched(t, fx.wt+"/spectre/changes/demo/tasks.md")
}

// TestRemoveChangeWorktreesRefusesWhenFlowdCannotAnswer: `flow change file`
// exiting 2 leaves the save unknown, so the gate refuses exit 2 with nothing
// removed.
func TestRemoveChangeWorktreesRefusesWhenFlowdCannotAnswer(t *testing.T) {
	t.Parallel()
	fx := rcwNewFx(t)
	guardtest.WriteFile(t, fx.repo+"/.git/info/exclude", "spectre/changes/\n.superpowers/\n")
	guardtest.WriteFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
	code, out, errb := fx.run(t, rcwFlowStub(t, fx, "2"))
	if code != 2 || !strings.Contains(errb, "cannot tell whether flow change save demo has run") {
		t.Fatalf("exit %d, want 2 naming the unknown save\nstdout:\n%s\nstderr:\n%s", code, out, errb)
	}
	fx.assertUntouched(t, fx.wt+"/spectre/changes/demo/tasks.md")
}

// TestRemoveChangeWorktreesRefusesAnUnsavedBundle: flowd holding the plan
// but not the self-review bundle the worktree holds -- a save that failed
// on its second write -- still refuses, naming the bundle.
func TestRemoveChangeWorktreesRefusesAnUnsavedBundle(t *testing.T) {
	t.Parallel()
	fx := rcwNewFx(t)
	guardtest.WriteFile(t, fx.repo+"/.git/info/exclude", "spectre/changes/\n.superpowers/\n")
	guardtest.WriteFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
	guardtest.WriteFile(t, fx.wt+"/.superpowers/sdd/demo/self-review-context.md", "bundle\n")
	bin := fx.dir + "/bin"
	guardtest.WriteExec(t, bin+"/flow", "#!/usr/bin/env bash\n[ \"$1 $2 $6\" = \"change file tasks.md\" ] && exit 0\n[ \"$1 $2\" = \"change file\" ] && exit 1\nexit 0\n")
	code, _, errb := fx.run(t, func(env *core.Env) {
		inner := env.Getenv
		env.Getenv = func(k string) string {
			if k == "PATH" {
				return bin + ":" + os.Getenv("PATH")
			}
			return inner(k)
		}
	})
	if code != 1 || !strings.Contains(errb, "self-review-context.md") {
		t.Fatalf("exit %d, want 1 naming the unsaved bundle\nstderr:\n%s", code, errb)
	}
	fx.assertUntouched(t, fx.wt+"/.superpowers/sdd/demo/self-review-context.md")
}

// TestRemoveChangeWorktreesTreatsTheIgnoredPlanAsExpected: once flowd holds
// the save, the ignored change directory and .superpowers/ go with the
// worktree.
func TestRemoveChangeWorktreesTreatsTheIgnoredPlanAsExpected(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := rcwNewFx(t)
	guardtest.WriteFile(t, fx.repo+"/.git/info/exclude", "spectre/changes/\n.superpowers/\n")
	guardtest.WriteFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
	guardtest.WriteFile(t, fx.wt+"/.superpowers/sdd/demo/self-review-context.md", "bundle\n")
	guardtest.WriteFile(t, fx.wt+"/.superpowers/brainstorm/notes.md", "notes\n")
	code, out, errb := fx.run(t, rcwFlowStub(t, fx, "0"))
	if code != 0 {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
	}
	if guardtest.Exists(fx.wt) {
		t.Error("the worktree survived")
	}
}
