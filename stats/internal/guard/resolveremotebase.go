package guard

import (
	"os"
	"os/exec"
	"strings"
)

// resolveremotebase.go is the Go home of scripts/lib/resolve-remote-base.sh at d71a2327
// (KAN-88 half 1 hardening, design.md: preflight-resolves-remote-tracking),
// shared by check-base-moved and check-finish-preflight so the two guards can
// never disagree about which ref answers a question about the base, plus the
// plumbing both run git through and the usage message both print.

// baseRefUsage is the usage both guards print, verbatim from their heredocs.
func baseRefUsage(script string) string {
	return "usage: " + script + " <worktree> <base-ref> <recorded-merge-base|->\n" +
		"  <base-ref>  the base branch name, bare (main) or remote-tracking\n" +
		"              (origin/main). The guard prefers refs/remotes/origin/<base-ref>\n" +
		"              when it resolves, so a bare name is never tested against a\n" +
		"              stale local branch.\n"
}

// envGit returns a builder for `git <args>` run as the bash guards ran it:
// git resolved over env's PATH, in env.Dir (so a relative worktree argument
// resolves as it did), with extra appended to the process environment.
func envGit(env Env, extra ...string) func(args ...string) *exec.Cmd {
	git := "git"
	if p, ok := lookPath(env, "git"); ok {
		git = p
	}
	return func(args ...string) *exec.Cmd {
		cmd := exec.Command(git, args...)
		cmd.Dir = env.Dir
		if len(extra) > 0 {
			cmd.Env = append(os.Environ(), extra...)
		}
		return cmd
	}
}

// capture is `VAR="$(cmd 2>/dev/null)"`: stdout with trailing newlines
// stripped, and whether the command exited 0.
func capture(cmd *exec.Cmd) (string, bool) {
	out, err := cmd.Output()
	return strings.TrimRight(string(out), "\n"), err == nil
}

// resolveRemoteBase is resolve_remote_base <worktree> <base-ref>:
// `origin/<base-ref>` when `refs/remotes/origin/<base-ref>` resolves to a
// commit in <worktree>, and <base-ref> unchanged otherwise.
//
// THE LOOKUP IS A PREFERENCE, NOT A REWRITE (design.md:
// remote-lookup-is-a-preference-not-a-rewrite). The correct caller already
// composes `origin/$BASE` itself; feeding that back in must be a no-op.
// `origin/main` looks up `refs/remotes/origin/origin/main`, finds nothing,
// and passes through unchanged — so the contract's own call site, and every
// caller handed a repository with no `origin` remote, keep their behaviour.
//
// Never fails the caller: an unreadable worktree, or a ref that does not
// resolve either way, still yields a value. The caller's own next step
// already refuses a ref that does not resolve, by name — this function only
// ever narrows which name that check tests.
//
// `--end-of-options` on the `rev-parse` so a ref beginning with `-` is read
// as a ref rather than parsed as a git option, matching every other ref
// resolution in this repository's guards. The fixed `refs/remotes/origin/`
// prefix means no input can make the flag observable; a source test pins it.
func resolveRemoteBase(git func(...string) *exec.Cmd, worktree, baseRef string) string {
	if git("-C", worktree, "rev-parse", "--verify", "--end-of-options",
		"refs/remotes/origin/"+baseRef+"^{commit}").Run() == nil {
		return "origin/" + baseRef
	}
	return baseRef
}
