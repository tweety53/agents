package guard

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// unchanged is scripts/unchanged.sh: that script's header comment is the
// contract -- exit 0 with one skip line when no path matching the pathspecs
// changed since HEAD's merge-base on origin/main (committed, staged,
// unstaged or untracked), 1 when one did, 2 on a usage error or a git call
// that failed.
func init() { Registry["unchanged"] = unchanged }

func unchanged(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "unchanged"
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: unchanged.sh <pathspec>...")
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
