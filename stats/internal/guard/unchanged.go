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
// excludes) changed. The full suite — the suite commands without the guard — runs only when the
// operator asks for it.
//
// Exit codes:
//   0  nothing matching changed: the suite is skipped, and one line says so
//   1  a matching path changed: the suite runs
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
	gitBin, ok := panelResolveGit(env, prog, stderr)
	if !ok {
		return 2
	}
	git := func(a ...string) *exec.Cmd {
		cmd := exec.Command(gitBin, a...)
		cmd.Dir = env.Dir
		return cmd
	}
	out, err := git("merge-base", "HEAD", "origin/main").Output()
	if err != nil {
		fmt.Fprintf(stderr, "%s: git merge-base HEAD origin/main failed: %v\n", prog, err)
		return 2
	}
	base := strings.TrimSpace(string(out))
	err = git(append([]string{"diff", "--quiet", base, "--"}, args...)...).Run()
	if err != nil {
		if exitCode(err) == 1 {
			return 1
		}
		fmt.Fprintf(stderr, "%s: git diff failed: %v\n", prog, err)
		return 2
	}
	untracked, err := git(append([]string{"ls-files", "--others", "--exclude-standard", "--"}, args...)...).Output()
	if err != nil {
		fmt.Fprintf(stderr, "%s: git ls-files failed: %v\n", prog, err)
		return 2
	}
	if len(untracked) > 0 {
		return 1
	}
	fmt.Fprintf(stdout, "unchanged: nothing under %s changed since %s — suite skipped\n", strings.Join(args, " "), base)
	return 0
}
