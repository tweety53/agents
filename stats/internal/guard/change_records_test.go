package guard

import (
	"os"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// The change directory is gitignored (KAN-1073): the guards that once read a
// change from git — a committed directory, a drop or archive subject — read
// flowd's change records and the worktrees on disk instead, and cleanup
// refuses to remove a worktree holding the only copy of a plan flowd has not
// saved.

// TestPipelineFixHighestKReadsFlowdChanges pins pfHighestK's sources: a
// change flowd records (FINISHED included) and a directory under
// <main>-worktrees/ count, and an answer from flowd's fallback journal
// refuses exit 2.
func TestPipelineFixHighestKReadsFlowdChanges(t *testing.T) {
	pfIdentity(t)
	fx := pfNew(t, "")
	writeFile(t, fx.states, `{"source":"store","complete":true,"records":[{"name":"pipeline-fix-x-9","state":"FINISHED"}]}`)
	mkdir(t, fx.repo+"-worktrees/pipeline-fix-y-11")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "3"); rc != 1 || !strings.Contains(out, "<k> is 12") {
		t.Fatalf("rc %d, want the next key 12 — one past the worktree's 11, flowd's 9 below it: %s", rc, out)
	}
	if !strings.Contains(fx.calls(t), "state list -C "+fx.repo) {
		t.Fatalf("flowd's change records were never asked for: %s", fx.calls(t))
	}
	writeFile(t, fx.states, `{"source":"fallback","complete":false,"records":[]}`)
	if rc, out := fx.pf(t, "start", "kan-1-demo", "12"); rc != 2 || !strings.Contains(out, "could not be read in full") {
		t.Fatalf("rc %d, want 2 on a fallback answer: %s", rc, out)
	}
}

// rcwFlowStub puts a `flow` on PATH whose `change file` exits rc, so the save
// gate reads flowd's answer from it.
func rcwFlowStub(t *testing.T, fx *rcwFx, rc string) func(*Env) {
	t.Helper()
	bin := fx.dir + "/bin"
	writeExec(t, bin+"/flow", "#!/usr/bin/env bash\n[ \"$1 $2\" = \"change file\" ] && exit "+rc+"\nexit 0\n")
	return func(env *Env) {
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
	writeFile(t, fx.repo+"/.git/info/exclude", "spectre/changes/\n.superpowers/\n")
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
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
	writeFile(t, fx.repo+"/.git/info/exclude", "spectre/changes/\n.superpowers/\n")
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
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
	writeFile(t, fx.repo+"/.git/info/exclude", "spectre/changes/\n.superpowers/\n")
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
	writeFile(t, fx.wt+"/.superpowers/sdd/demo/self-review-context.md", "bundle\n")
	bin := fx.dir + "/bin"
	writeExec(t, bin+"/flow", "#!/usr/bin/env bash\n[ \"$1 $2 $6\" = \"change file tasks.md\" ] && exit 0\n[ \"$1 $2\" = \"change file\" ] && exit 1\nexit 0\n")
	code, _, errb := fx.run(t, func(env *Env) {
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
// the save, the ignored change directory and .superpowers/ are counted as
// regeneratable, never listed as unclassified, and the worktree is removed.
func TestRemoveChangeWorktreesTreatsTheIgnoredPlanAsExpected(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := rcwNewFx(t)
	writeFile(t, fx.repo+"/.git/info/exclude", "spectre/changes/\n.superpowers/\n")
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
	writeFile(t, fx.wt+"/.superpowers/sdd/demo/self-review-context.md", "bundle\n")
	writeFile(t, fx.wt+"/.superpowers/brainstorm/notes.md", "notes\n")
	code, out, errb := fx.run(t, rcwFlowStub(t, fx, "0"))
	if code != 0 {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
	}
	if l := rcwLines(out, "UNCLASSIFIED: "); len(l) != 0 {
		t.Errorf("UNCLASSIFIED lines %q, want none", l)
	}
	if l := rcwLines(out, "REGENERATABLE: "); len(l) != 1 || !strings.HasSuffix(l[0], " — 3") {
		t.Errorf("REGENERATABLE lines %q, want one counting 3", l)
	}
	if rcwExists(fx.wt) {
		t.Error("the worktree survived")
	}
}
