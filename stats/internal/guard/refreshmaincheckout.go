package guard

import (
	"fmt"
	"io"
)

// refreshMainCheckout is scripts/refresh-main-checkout.sh: that script's
// header comment is the contract -- 0 REFRESH-DONE or REFRESH-CURRENT, 1
// REFRESH-REFUSED with nothing touched, 2 usage or cannot answer, every
// verdict on stdout. A git read that fails outright -- an unborn HEAD -- is
// exit 2 with git's own message on stderr.
func init() {
	Registry["refresh-main-checkout"] = refreshMainCheckout
}

func refreshMainCheckout(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 || !isDir(pcAbs(env, args[0])) {
		fmt.Fprintln(stderr, "usage: refresh-main-checkout.sh <main-checkout> <base>")
		return 2
	}
	repo, base := args[0], args[1]
	git := envGit(env)
	// g is git -C <repo>, its stderr reaching the caller.
	g := func(a ...string) (string, bool) {
		cmd := git(append([]string{"-C", repo}, a...)...)
		cmd.Stderr = stderr
		return capture(cmd)
	}
	refuse := func(reason string) int {
		fmt.Fprintf(stdout, "REFRESH-REFUSED: %s %s — nothing touched\n", repo, reason)
		return 1
	}

	// The bounded, credential-free fetch resolve-base-branch.sh runs: no
	// credential prompt, chatter swallowed, and a failed fetch is not this
	// guard's failure -- a stale origin/<base> only means a smaller step.
	_ = git("-C", repo, "-c", "core.askpass=true", "fetch", "--quiet", "origin").Run()

	cur, _ := capture(git("-C", repo, "symbolic-ref", "-q", "--short", "HEAD"))
	if cur == "" {
		return refuse("is detached")
	}
	if cur != base {
		return refuse("is on " + cur + ", not " + base)
	}
	// Any tracked entry -- modified, staged or unmerged -- is work the
	// fast-forward could collide with; untracked files are git's own concern.
	status, ok := g("status", "--porcelain", "--untracked-files=no")
	if !ok {
		return 2
	}
	if status != "" {
		return refuse("has tracked changes")
	}
	tip, ok := g("rev-parse", "--short", "HEAD")
	if !ok {
		return 2
	}
	remote, _ := capture(git("-C", repo, "rev-parse", "-q", "--verify", "--short", "refs/remotes/origin/"+base))
	if remote == "" {
		return refuse("has no origin/" + base)
	}
	if tip == remote {
		fmt.Fprintf(stdout, "REFRESH-CURRENT: %s already at %s\n", repo, tip)
		return 0
	}
	if err := git("-C", repo, "merge-base", "--is-ancestor", "HEAD", "refs/remotes/origin/"+base).Run(); err != nil {
		return refuse("has " + base + " commits origin/" + base + " lacks")
	}
	if _, ok := g("merge", "--ff-only", "-q", "refs/remotes/origin/"+base); !ok {
		return 2
	}
	fmt.Fprintf(stdout, "REFRESH-DONE: %s fast-forwarded %s -> %s\n", repo, tip, remote)
	return 0
}
