package guard

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
	"time"
)

// removeChangeWorktrees is scripts/remove-change-worktrees.sh: that script's
// header is the contract. It runs Worktree cleanup
// (skills/flow-contracts/finish-contract-run2.md) for one repository: the
// six checks on every apply worktree of spectre/<name> and checks 5-6 on
// every wave-group copy, all before anything is removed; then the removals,
// the local branch and the remote branch. Check 5 reads the `## stop` key in
// project-configuration.md's shape, through kwFenced as `## worktree setup` is.
func init() { Registry["remove-change-worktrees"] = removeChangeWorktrees }

// rcwRegenDirs are the path components check 4 counts as regeneratable — an
// image under one included, since it is that directory's, not a loose capture.
var rcwRegenDirs = []string{"build", ".gradle", ".kotlin", "node_modules", "dist", ".next", "target", "out", "coverage", "test-results"}

func removeChangeWorktrees(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "remove-change-worktrees: "+format+"\n", a...)
		return 2
	}
	proceed := slices.Contains(args, "--proceed")
	args = slices.DeleteFunc(slices.Clone(args), func(a string) bool { return a == "--proceed" })
	if len(args) != 3 {
		return refuse("usage: remove-change-worktrees.sh <repo> <name> <merge-base|-> [--proceed]")
	}
	repo, name, mergeBase := pcAbs(env, args[0]), args[1], args[2]
	if !plainChangeName(name) {
		return refuse("change name %q is not a plain change name", name)
	}
	if mergeBase != "-" && strings.HasPrefix(mergeBase, "-") {
		return refuse("merge base %q is not a revision", mergeBase)
	}
	scriptDir, ok := guardSelfDir(env, stderr, "remove-change-worktrees: ", "remove-change-worktrees")
	if !ok {
		return 2
	}
	// LC_ALL=C keeps the remote delete's "remote ref does not exist" in
	// English, the one git message this guard reads.
	git := envGit(env, "LC_ALL=C")
	if !isDir(repo) || git("-C", repo, "rev-parse", "--git-dir").Run() != nil {
		return refuse("%s is not a git repository", repo)
	}
	porcelain, err := git("-C", repo, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return refuse("cannot list the worktrees of %s", repo)
	}
	listed, copies := changeWorktrees(porcelain, name)
	// An already-removed worktree is success: only its registration is left,
	// which the prune below clears.
	var wts, gone []string
	for _, wt := range listed {
		if isDir(wt) {
			wts = append(wts, wt)
		} else {
			gone = append(gone, wt)
		}
	}

	failed := false
	fail := func(wt string, check int, format string, a ...any) {
		failed = true
		fmt.Fprintf(stdout, "REFUSED: %s — check %d: %s\n", wt, check, rcwFlat(fmt.Sprintf(format, a...)))
	}
	unclassified := false
	upstream := "" // origin/<base>, as check 3 resolved it for the first live worktree
	for _, wt := range wts {
		// 1. no uncommitted tracked changes.
		if s, ok := capture(git("-C", wt, "status", "--porcelain", "--untracked-files=no")); !ok {
			fail(wt, 1, "git status failed")
		} else if s != "" {
			fail(wt, 1, "uncommitted tracked changes: %s", s)
		}
		// 2. no untracked files git does not already ignore.
		if s, ok := capture(git("-C", wt, "ls-files", "--others", "--exclude-standard")); !ok {
			fail(wt, 2, "git ls-files failed")
		} else if s != "" {
			fail(wt, 2, "untracked files: %s", s)
		}
		// 3. no commits that exist only here, against a base resolved fresh
		// for THIS worktree.
		base, msg := rcwOnlyHere(env, git, wt, stderr)
		if msg != "" {
			fail(wt, 3, "%s", msg)
		} else if upstream == "" {
			upstream = "origin/" + base
		}
		// 4. what --force will destroy, split by path.
		out, err := git("-C", wt, "ls-files", "-z", "--others", "--ignored", "--exclude-standard").Output()
		if err != nil {
			fail(wt, 4, "git ls-files failed")
			continue
		}
		regen := 0
		for _, p := range strings.Split(string(out), "\x00") {
			switch {
			case p == "":
			case rcwRegeneratable(p):
				regen++
			default:
				unclassified = true
				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s\n", wt, rcwFlat(p))
			}
		}
		fmt.Fprintf(stdout, "REGENERATABLE: %s — %d\n", wt, regen)
	}
	if failed {
		return 1
	}

	// The disclosure stop: the run relays, judges each unclassified entry,
	// asks its one ask, and calls again with --proceed.
	if !proceed && (unclassified || len(copies) > 0) {
		for _, c := range copies {
			fmt.Fprintf(stdout, "DISCLOSE: %s — wave-group copy, removed with --force on --proceed\n", c)
			rcwDisclose(stdout, c, "status", git("-C", c, "status", "--short"))
			if mergeBase != "-" {
				rcwDisclose(stdout, c, "commit", git("-C", c, "log", "--oneline", mergeBase+"..HEAD"))
			}
		}
		fmt.Fprintln(stdout, "DISCLOSE: stopped before removal — relay the entries above, then re-run with --proceed")
		return 3
	}

	// 5 and 6 run on every worktree and every copy, and stay gates.
	stopBody := projectSection(repo+"/.flow/project.md", "stop")
	stopCmd := strings.Join(kwFenced(stopBody), "\n")
	if stopBody != "" && stopCmd == "" {
		fmt.Fprintln(stdout, "SKIPPED: check 5 — ## stop declares no fenced command")
	}
	for _, wt := range append(slices.Clone(copies), wts...) {
		if stopCmd != "" {
			timeout, grace := ccDefaultTimeout, ccDefaultGrace
			if env.SurvivorsTimeout > 0 {
				timeout = env.SurvivorsTimeout
			}
			if env.SurvivorsKillGrace > 0 {
				grace = env.SurvivorsKillGrace
			}
			out, rc, timedOut, started := ccRunSurvivors(env, wt, stopCmd, timeout, grace, stderr)
			_, _ = stderr.Write(out)
			switch {
			case !started:
				fail(wt, 5, "the ## stop command could not be started")
			case timedOut:
				fail(wt, 5, "the ## stop command timed out after %s", timeout.Round(time.Second))
			case rc != 0:
				fail(wt, 5, "the ## stop command exited %d", rc)
			}
		}
		procs := exec.Command(scriptDir+"/check-worktree-processes.sh", wt)
		procs.Stderr = stderr
		verdict, err := procs.Output()
		switch v := string(verdict); {
		case err == nil && strings.HasPrefix(v, "CLEAR:"):
		case err == nil && strings.HasPrefix(v, "HELD:"):
			failed = true
			_, _ = io.WriteString(stdout, v)
		default:
			fail(wt, 6, "check-worktree-processes.sh could not answer")
		}
	}
	if failed {
		return 1
	}

	// Then, and only then: copies before the worktree they were copied from.
	for _, wt := range append(slices.Clone(copies), wts...) {
		if out, err := git("-C", repo, "worktree", "remove", "--force", wt).CombinedOutput(); err != nil {
			failed = true
			fmt.Fprintf(stdout, "REFUSED: %s — git worktree remove: %s\n", wt, rcwFlat(string(out)))
		} else {
			fmt.Fprintf(stdout, "REMOVED: %s\n", wt)
		}
	}
	for _, wt := range gone {
		fmt.Fprintf(stdout, "REMOVED: %s — already gone\n", wt)
	}
	// Pruned before the branch delete, so a gone worktree's registration
	// does not hold the branch as checked out.
	_ = git("-C", repo, "worktree", "prune").Run()
	branch := "spectre/" + name
	if git("-C", repo, "show-ref", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil {
		// -d judges against the branch's upstream, else the main checkout's
		// HEAD — which run 2 refreshes only after cleanup, and which a forge
		// merge never moved. Pointing the upstream at origin/<base> first
		// makes -d judge against where the change landed, even after the
		// forge deleted origin/spectre/<name> and a prune removed its ref.
		// With every worktree already gone there is no checkout of the branch
		// for resolve-base-branch to read: the base recorded on the branch,
		// else origin/HEAD's, stands in.
		if upstream == "" {
			upstream, _ = capture(git("-C", repo, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"))
			if base := recordedBase(git, repo, branch); base != "" {
				upstream = "origin/" + base
			}
		}
		remote, _ := capture(git("-C", repo, "config", "--get", "branch."+branch+".remote"))
		merge, _ := capture(git("-C", repo, "config", "--get", "branch."+branch+".merge"))
		if upstream != "" {
			_ = git("-C", repo, "branch", "-q", "--set-upstream-to="+upstream, branch).Run()
		}
		// -d, never -D: it must be free to refuse an unmerged branch.
		if out, err := git("-C", repo, "branch", "-d", branch).CombinedOutput(); err != nil {
			failed = true
			fmt.Fprintf(stdout, "REFUSED: %s — git branch -d: %s\n", branch, rcwFlat(string(out)))
			// A surviving branch keeps the upstream it had, never origin/<base>.
			rcwRestoreUpstream(git, repo, branch, remote, merge)
		} else {
			fmt.Fprintf(stdout, "REMOVED: %s\n", branch)
		}
	}

	// The remote delete is not gated on the local one succeeding.
	out, err := git("-C", repo, "push", "origin", "--delete", branch).CombinedOutput()
	switch {
	case err == nil:
		fmt.Fprintf(stdout, "REMOTE-DELETED: origin/%s\n", branch)
	case bytes.Contains(out, []byte("remote ref does not exist")):
		// The forge deleted it on merge. Prune the ref it left behind:
		// check-cleanup-complete reads a surviving tracking ref as a leftover.
		if git("-C", repo, "fetch", "--prune", "--quiet", "origin").Run() != nil {
			fmt.Fprintf(stderr, "remove-change-worktrees: fetch --prune failed in %s — the stale refs/remotes/origin/%s may survive\n", repo, branch)
		}
		fmt.Fprintf(stdout, "REMOTE-GONE: origin/%s\n", branch)
	default:
		fmt.Fprintf(stdout, "REMOTE-REFUSED: origin/%s — %s\n", branch, rcwFlat(string(out)))
	}
	if failed {
		return 1
	}
	return 0
}

// rcwOnlyHere is check 3: "" when HEAD is merged into origin/<base> or has
// nothing its upstream lacks, else why not, with the base it resolved. A
// failed lookup never passes.
func rcwOnlyHere(env Env, git func(...string) *exec.Cmd, wt string, stderr io.Writer) (string, string) {
	var buf bytes.Buffer
	if rc := resolveBaseBranch([]string{wt}, env, &buf, stderr); rc != 0 {
		return "", fmt.Sprintf("cannot resolve the base branch (resolve-base-branch exit %d) — stop and ask", rc)
	}
	base := strings.TrimSpace(buf.String())
	if git("-C", wt, "merge-base", "--is-ancestor", "HEAD", "origin/"+base).Run() == nil {
		return base, ""
	}
	up, ok := capture(git("-C", wt, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"))
	if !ok || up == "" {
		return base, "not merged, and no upstream — cannot prove these commits exist anywhere else"
	}
	log, ok := capture(git("-C", wt, "log", "--oneline", up+"..HEAD"))
	if !ok {
		return base, "cannot read " + up + "..HEAD"
	}
	if log != "" {
		return base, "commits not on " + up + ": " + log
	}
	return base, ""
}

// rcwRestoreUpstream puts back the branch.<branch>.remote/.merge pair read
// before the upstream was pointed at origin/<base>, or unsets the upstream
// when the branch had none.
func rcwRestoreUpstream(git func(...string) *exec.Cmd, repo, branch, remote, merge string) {
	if remote == "" || merge == "" {
		_ = git("-C", repo, "branch", "-q", "--unset-upstream", branch).Run()
		return
	}
	_ = git("-C", repo, "config", "branch."+branch+".remote", remote).Run()
	_ = git("-C", repo, "config", "branch."+branch+".merge", merge).Run()
}

// rcwRegeneratable is check 4's regeneratable bucket, decided by path alone
// (Worktree cleanup, finish-hand-fallbacks.md). A .png/.jpg capture counts
// only under a regeneratable location, test-results among them — a capture
// loose anywhere else matches no rule below; everything the list does not name
// stays unclassified.
func rcwRegeneratable(p string) bool {
	parts := strings.Split(p, "/")
	if strings.HasPrefix(p, ".superpowers/sdd/") || strings.HasPrefix(p, ".dev-stack/") || strings.HasSuffix(p, ".log") {
		return true
	}
	for _, c := range parts {
		if slices.Contains(rcwRegenDirs, c) {
			return true
		}
	}
	return false
}

// rcwDisclose prints one DISCLOSE line per line cmd prints.
func rcwDisclose(stdout io.Writer, wt, what string, cmd *exec.Cmd) {
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(stdout, "DISCLOSE: %s — %s unreadable\n", wt, what)
		return
	}
	for _, l := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if l != "" {
			fmt.Fprintf(stdout, "DISCLOSE: %s — %s: %s\n", wt, what, rcwFlat(l))
		}
	}
}

// rcwFlat keeps a multi-line message on its verdict line, its control bytes
// escaped (ccSanitize): git output and paths are not this guard's to trust.
func rcwFlat(s string) string {
	return strings.ReplaceAll(strings.TrimRight(ccSanitize(s), "\n"), "\n", "; ")
}
