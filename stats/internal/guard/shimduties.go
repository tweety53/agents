package guard

import (
	"os"
	"os/exec"
	"path/filepath"
)

// The duties every guard's scripts/<name>.sh shim carried beside its exec,
// now in the binary: the cannot-answer code by guard name, a sibling guard's
// command, and the KAN-902 re-exec into the working directory's worktree.

// CannotAnswer is guard name's "could not answer" exit code — the code
// scripts/flow-guard exits when it cannot derive the binary, and the code
// cmd/flow-guard exits for a failure of its own past the name: 4 for
// run-reproducer (whose 2 is "refused"), mutate-and-verify and
// break-and-prove; 1 for check-hand-notes-in-step and check-installed-rules;
// 2 for every other guard.
func CannotAnswer(name string) int {
	switch name {
	case "run-reproducer", "mutate-and-verify", "break-and-prove":
		return 4
	case "check-hand-notes-in-step", "check-installed-rules":
		return 1
	}
	return 2
}

// siblingExe, when set, is the executable sibling guards run as instead of
// this one: this package's tests point it at a flow-guard built from these
// sources, so a sibling runs as the production binary, never as the
// race-instrumented test binary.
var siblingExe string

// selfExe is the executable a sibling guard runs as: siblingExe, else this
// very executable.
func selfExe() (string, error) {
	if siblingExe != "" {
		return siblingExe, nil
	}
	return os.Executable()
}

// runSibling is the command that runs sibling guard name with args: this
// very executable, the sibling's name its first argument — the same process
// isolation and exit codes the sibling's own scripts/<name>.sh exec gave.
// It runs in env.Dir; a caller sets the rest. An executable that cannot be
// resolved surfaces as the command's start error.
func runSibling(env Env, name string, args ...string) *exec.Cmd {
	self, err := selfExe()
	cmd := exec.Command(self, append([]string{name}, args...)...)
	if err != nil {
		cmd.Err = err
	}
	cmd.Dir = env.Dir
	return cmd
}

// siblingCmd is runSibling, unless override — the sibling's
// FLOW_GUARD_<NAME> test override — names an executable to run instead.
func siblingCmd(env Env, override, name string, args ...string) *exec.Cmd {
	if p := env.Getenv(override); p != "" {
		cmd := exec.Command(p, args...)
		cmd.Dir = env.Dir
		return cmd
	}
	return runSibling(env, name, args...)
}

// selfJudging is every guard that judges its own checkout — the guards whose
// shim called flow_guard_cwd_checkout before deriving its root.
var selfJudging = map[string]bool{
	"check-ask-silence":             true,
	"check-dev-stack-fresh":         true,
	"check-installed-citations":     true,
	"check-guard-symlinks":          true,
	"check-dispatch-paragraphs":     true,
	"check-plan-shape":              true,
	"check-references":              true,
	"check-markdown-integrity":      true,
	"check-task-build-green":        true,
	"check-spec-reach":              true,
	"check-uitest-overrides":        true,
	"check-mutation-reproducer-pin": true,
	"check-normative-inventory":     true,
	"check-plan-provenance":         true,
	"check-python-suppressions":     true,
	"check-verbatim-moves":          true,
	"check-stage-mark-calls":        true,
	"plan-dispatch-bundles":         true,
	"run-guard-tests":               true,
}

// CwdCheckout is the KAN-902 re-exec: for a self-judging guard called from
// another worktree of the repository FLOW_GUARD_REPO_ROOT names, it returns
// that worktree's own scripts/flow-guard, so an absolute-path call from a
// worktree judges the worktree rather than reporting a false clean on the
// bootstrap's checkout. Not ok — carry on here — for any other guard, an
// unset root, a working directory in no repository or in this checkout, a
// different repository, or a worktree without scripts/flow-guard. The
// re-exec'd bootstrap exports its own root, so it never loops.
func CwdCheckout(env Env, name string) (string, bool) {
	root := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if !selfJudging[name] || root == "" {
		return "", false
	}
	top, ok := gitOut(env, "rev-parse", "--show-toplevel")
	if !ok {
		return "", false
	}
	top, ok = physicalDir(top)
	if !ok {
		return "", false
	}
	if r, ok := physicalDir(root); ok && r == top {
		return "", false
	}
	script := filepath.Join(top, "scripts", "flow-guard")
	if st, err := os.Stat(script); err != nil || !st.Mode().IsRegular() {
		return "", false
	}
	commonDir := []string{"rev-parse", "--path-format=absolute", "--git-common-dir"}
	common, ok := gitOut(Env{Dir: root}, commonDir...)
	if !ok || common == "" {
		return "", false
	}
	if other, ok := gitOut(Env{Dir: top}, commonDir...); !ok || other != common {
		return "", false
	}
	return script, true
}
