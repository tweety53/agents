package guard

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/unchanged.sh shim (KAN-1058).
//
// flow-guard unchanged — tell a `## test` line whether the branch left every path it reads untouched.
//
// Usage: flow-guard unchanged <pathspec>...
//
// Compares HEAD and the working tree (untracked files included) with its merge-base on
// origin/main. A suite line guards itself as `flow-guard unchanged <pathspecs> || { <suite>; }`,
// so the suite runs only when a path matching a <pathspec> (git pathspec syntax, so ':!docs/'
// excludes) changed. Under FLOW_VERIFY_FULL, which `flow-guard run-verify-list -full` exports,
// the guard answers 1 without looking, so the full suite runs every suite.
//
// Exit codes:
//   0  nothing matching changed: the suite is skipped, and one line says so
//   1  a matching path changed, or FLOW_VERIFY_FULL is set: the suite runs
//   2  usage: no pathspec, or a git call failed
//
// The logic is the Go port in stats/internal/guard/unchanged.go.

// unchanged is `flow-guard unchanged`: the comment above is the contract -- exit 0 with one skip line when no path matching the pathspecs
// changed since HEAD's merge-base on origin/main (committed, staged,
// unstaged or untracked), 1 when one did, 2 on a usage error or a git call
// that failed.
func init() { Registry["unchanged"] = unchanged }

func unchanged(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "unchanged"
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: flow-guard unchanged <pathspec>...")
		return 2
	}
	if env.Getenv("FLOW_VERIFY_FULL") != "" {
		return 1
	}
	base, paths, ok := changedSinceBase(env, prog, stderr, args...)
	if !ok {
		return 2
	}
	if len(paths) > 0 {
		return 1
	}
	fmt.Fprintf(stdout, "unchanged: nothing under %s changed since %s — suite skipped\n", strings.Join(args, " "), base)
	return 0
}

// changedSinceBase answers the merge base of HEAD and origin/main and the
// paths matching pathspecs (all paths when none) that changed since it --
// committed, staged, unstaged or untracked -- each repository-relative. A
// failed git call prints one `<prog>: git … failed: …` line and answers ok
// false.
func changedSinceBase(env Env, prog string, stderr io.Writer, pathspecs ...string) (base string, paths []string, ok bool) {
	gitBin, ok := panelResolveGit(env, prog, stderr)
	if !ok {
		return "", nil, false
	}
	git := func(a ...string) ([]byte, error) {
		cmd := exec.Command(gitBin, a...)
		cmd.Dir = env.Dir
		return cmd.Output()
	}
	out, err := git("merge-base", "HEAD", "origin/main")
	if err != nil {
		fmt.Fprintf(stderr, "%s: git merge-base HEAD origin/main failed: %v\n", prog, err)
		return "", nil, false
	}
	base = strings.TrimSpace(string(out))
	diff, err := git(append([]string{"diff", "-z", "--no-renames", "--name-only", base, "--"}, pathspecs...)...)
	if err != nil {
		fmt.Fprintf(stderr, "%s: git diff failed: %v\n", prog, err)
		return "", nil, false
	}
	untracked, err := git(append([]string{"ls-files", "-z", "--others", "--exclude-standard", "--full-name", "--"}, pathspecs...)...)
	if err != nil {
		fmt.Fprintf(stderr, "%s: git ls-files failed: %v\n", prog, err)
		return "", nil, false
	}
	nul := func(r rune) bool { return r == 0 }
	return base, append(strings.FieldsFunc(string(diff), nul), strings.FieldsFunc(string(untracked), nul)...), true
}
