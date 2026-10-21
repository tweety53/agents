package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
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
	// fast-forward could collide with, unless it is a landed copy (below);
	// untracked files are git's own concern.
	status, ok := g("status", "--porcelain", "--untracked-files=no")
	if !ok {
		return 2
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
		if status != "" {
			return refuse("has tracked changes")
		}
		fmt.Fprintf(stdout, "REFRESH-CURRENT: %s already at %s\n", repo, tip)
		return 0
	}
	if err := git("-C", repo, "merge-base", "--is-ancestor", "HEAD", "refs/remotes/origin/"+base).Run(); err != nil {
		return refuse("has " + base + " commits origin/" + base + " lacks")
	}
	if status != "" {
		paths := landedCopies(git, repo, "refs/remotes/origin/"+base)
		if paths == nil {
			return refuse("has tracked changes")
		}
		// Never discard ahead of a fast-forward git would then refuse: a
		// path origin adds that sits here untracked, outside the copies.
		added, _ := g("diff", "--name-only", "-z", "--no-renames", "--diff-filter=A", "HEAD", "refs/remotes/origin/"+base)
		for _, p := range strings.Split(added, "\x00") {
			if p != "" && !slices.Contains(paths, p) && rmcPresent(filepath.Join(pcAbs(env, repo), p)) {
				return refuse("has untracked " + p + " origin/" + base + " adds")
			}
		}
		if _, ok := g(append([]string{"--literal-pathspecs", "restore", "--source=HEAD", "--staged", "--worktree", "--"}, paths...)...); !ok {
			return 2
		}
		fmt.Fprintf(stdout, "REFRESH-DISCARDED: %s %d tracked paths identical to origin/%s\n", repo, len(paths), base)
	}
	if _, ok := g("merge", "--ff-only", "-q", "refs/remotes/origin/"+base); !ok {
		return 2
	}
	fmt.Fprintf(stdout, "REFRESH-DONE: %s fast-forwarded %s -> %s\n", repo, tip, remote)
	if restartFlowdIfStatsMoved(env, repo, tip, remote, stdout, stderr) != 0 {
		return 2
	}
	return 0
}

// landedCopies returns the tracked paths a dirty main checkout holds when
// every one is byte-identical to ref -- its worktree copy, and its index
// copy wherever one is staged -- the copies a change landed from another
// worktree leaves behind (KAN-926); nil when any path differs, any entry is
// unmerged, or a read fails. An unstaged path's index copy is HEAD's, already
// committed, so restoring exactly these paths to HEAD loses nothing the
// fast-forward does not write back.
func landedCopies(git func(args ...string) *exec.Cmd, repo, ref string) []string {
	if u, ok := capture(git("-C", repo, "ls-files", "--unmerged")); !ok || u != "" {
		return nil
	}
	names := func(cached ...string) ([]string, bool) {
		args := append([]string{"-C", repo, "diff", "--name-only", "-z", "--no-renames"}, cached...)
		out, ok := capture(git(append(args, "HEAD")...))
		var ps []string
		for _, p := range strings.Split(out, "\x00") {
			if p != "" {
				ps = append(ps, p)
			}
		}
		return ps, ok
	}
	// same is `git diff --quiet ref -- paths`: worktree, or with --cached
	// the index, byte-identical to ref at every path.
	same := func(paths []string, cached ...string) bool {
		args := append([]string{"-C", repo, "--literal-pathspecs", "diff", "--quiet"}, cached...)
		return len(paths) == 0 || git(append(append(args, ref, "--"), paths...)...).Run() == nil
	}
	worktree, ok := names()
	staged, ok2 := names("--cached")
	if !ok || !ok2 || !same(staged, "--cached") {
		return nil
	}
	paths := append([]string{}, worktree...)
	for _, p := range staged {
		if !slices.Contains(paths, p) {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 || !same(paths) {
		return nil
	}
	// A clean filter compares what it keeps, not the bytes on disk: a lossy
	// one (output stripping, secret scrubbing) would pass differing content
	// as identical, so any filtered path refuses.
	attrs, ok := capture(git(append([]string{"-C", repo, "check-attr", "filter", "-z", "--"}, paths...)...))
	if !ok {
		return nil
	}
	f := strings.Split(attrs, "\x00")
	for i := 2; i < len(f); i += 3 {
		if f[i] != "unspecified" && f[i] != "unset" {
			return nil
		}
	}
	return paths
}

func rmcPresent(p string) bool { _, err := os.Lstat(p); return err == nil }
