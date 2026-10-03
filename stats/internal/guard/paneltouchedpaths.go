package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// The Go twin of scripts/lib/panel-touched-paths.sh, which stays the source
// of truth for its bash callers and whose header carries the reasoning: this
// change's own touched paths, defined once for every guard asking "which
// paths did this change touch?". Each function takes the calling guard's own
// name as its error-message prefix, so each guard's stderr wording is the
// bash's. TestPanelTouchedPathsParity runs the bash library and these
// functions over the same inputs and fails on any difference.

// panelResolveGit is panel_resolve_git: the git binary's path, found by a
// real search of env's PATH (`type -P git`), so every later call runs that
// binary by path. False, with a message on stderr, when PATH has no git.
func panelResolveGit(env Env, prog string, stderr io.Writer) (string, bool) {
	p, ok := lookPath(env, "git")
	if !ok {
		fmt.Fprintf(stderr, "%s: no git binary found on PATH\n", prog)
	}
	return p, ok
}

// panelGit runs gitBin -C worktree args from env's directory, its stderr
// discarded, as the library's `2>/dev/null` calls do.
func panelGit(env Env, gitBin, worktree string, args ...string) *exec.Cmd {
	cmd := exec.Command(gitBin, append([]string{"-C", worktree}, args...)...)
	cmd.Dir = env.Dir
	// GIT_CONFIG_GLOBAL is forwarded from env, as verbatimmoves.go does: a
	// no-op in production, the tests' /dev/null under test.
	if env.LookupEnv != nil {
		if v, ok := env.LookupEnv("GIT_CONFIG_GLOBAL"); ok {
			cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+v)
		}
	}
	return cmd
}

// panelValidateWorktree is panel_validate_worktree: true when worktree is a
// directory, a git worktree, and mergebase resolves to a commit inside it;
// false, with a message on stderr naming the failed check, otherwise —
// including an empty worktree or mergebase.
func panelValidateWorktree(env Env, prog, worktree, mergebase, gitBin string, stderr io.Writer) bool {
	switch {
	case worktree == "" || mergebase == "":
		fmt.Fprintf(stderr, "%s: usage: %s.sh <worktree> <merge-base>\n", prog, prog)
	case !isDir(smcAbs(env, worktree)):
		fmt.Fprintf(stderr, "%s: %s is not a directory — cannot determine anything\n", prog, worktree)
	case panelGit(env, gitBin, worktree, "rev-parse", "--git-dir").Run() != nil:
		fmt.Fprintf(stderr, "%s: %s is not a git worktree — cannot determine anything\n", prog, worktree)
	// --end-of-options: a merge base shaped like an option is read as a
	// revision, fails to resolve, and is reported as such.
	case panelGit(env, gitBin, worktree, "rev-parse", "--verify", "--end-of-options", mergebase+"^{commit}").Run() != nil:
		fmt.Fprintf(stderr, "%s: merge base '%s' does not resolve in %s\n", prog, mergebase, worktree)
	default:
		return true
	}
	return false
}

// panelTouchedPaths is panel_touched_paths: the union of what HEAD carries
// since mergebase, what is staged and what is unstaged (design.md:
// touched-paths-include-index-and-worktree), blank lines dropped, sorted and
// de-duplicated by an exec'd `sort -u` so the order is the caller's
// collation (crSort). False, with a message on stderr naming the collection
// that failed, when a `git diff` fails.
func panelTouchedPaths(env Env, prog, worktree, mergebase, gitBin string, stderr io.Writer) ([]string, bool) {
	var lines []string
	for _, c := range []struct {
		what string
		args []string
	}{
		{"committed", []string{"diff", "--no-renames", "--name-only", "--end-of-options", mergebase + "..HEAD"}},
		{"staged", []string{"diff", "--no-renames", "--name-only", "--cached"}},
		{"unstaged", []string{"diff", "--no-renames", "--name-only"}},
	} {
		out, ok := capture(panelGit(env, gitBin, worktree, c.args...))
		if !ok {
			fmt.Fprintf(stderr, "%s: cannot list this change's %s paths in %s\n", prog, c.what, worktree)
			return nil, false
		}
		for _, l := range strings.Split(out, "\n") {
			// awk 'NF': a line of only blanks and tabs is dropped too.
			if strings.Trim(l, " \t") != "" {
				lines = append(lines, l)
			}
		}
	}
	sorted, err := crSort(env, lines, true)
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", prog, err)
		return nil, false
	}
	return sorted, true
}
