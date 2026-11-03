package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// The contract below moved here from the header of the deleted
// scripts/remove-change-worktrees.sh shim (KAN-1058).
//
// flow-guard remove-change-worktrees — run /flow run 2's Worktree cleanup
// (skills/flow-contracts/finish-contract-run2.md) for one repository: every
// check on every worktree before anything is removed, then the removals, the
// local branch and the remote branch.
//
// Usage: flow-guard remove-change-worktrees <repo> <name> <merge-base|-> [--proceed]
//
// The worktrees are the ones `git -C <repo> worktree list --porcelain` lists on
// refs/heads/spectre/<name>, plus every detached wave-group copy of one
// (`<worktree>-wave-group-<g>`). A listed worktree whose directory is gone is
// already removed: success, its registration pruned. <merge-base> scopes the
// wave-group copies' disclosed log; `-` discloses their status alone.
//
// First the save gate: an apply worktree holding a file `flow change save`
// keeps — the plan, spectre/changes/<name>/tasks.md or
// .superpowers/sdd/<name>/tasks.md, or .superpowers/sdd/<name>/self-review-context.md,
// each gitignored and so the only copy — is refused, exit 1, with
// "flow change save <name> has not run" when flowd holds no saved copy of that
// file for that worktree's project (`flow change file`); flowd unreachable is
// exit 2. With the gate passed, check 4 counts spectre/changes/ and
// .superpowers/ as regeneratable.
//
// In order, never removing anything before every check has passed:
//   checks 1-4 on each apply worktree — no uncommitted tracked change; no
//     untracked, unignored file; no commit that exists only here (the base
//     resolved fresh for THIS worktree, in-process as flow-guard resolve-base-branch
//     does); and the ignored files `--force` will destroy, split by path into
//     regeneratable (counted) and unclassified (listed). Regeneratable takes
//     in a `.DS_Store` anywhere and, in a worktree whose root holds a
//     `playwright.config.*`, a `.png` at that root — a screenshot spec's
//     output; a `.png` in any subdirectory stays unclassified.
//     An unclassified entry at the worktree root that is a regular file
//     with a distinct regular file at the same path in the repository's
//     main checkout (the first entry of its worktree list) is preserved —
//     a per-worktree copy of a main-checkout local config, identical or
//     not — and so is a symlink there resolving to that regular file
//     (a link the project's dev tooling made; removal drops the link
//     alone); an entry with no such counterpart, or below the root,
//     stays unpreserved.
//   the disclosure stop — without --proceed, an unpreserved unclassified
//     entry or any wave-group copy stops here, before check 5 runs anything;
//     preserved entries alone are shown and the removal proceeds.
//   checks 5-6 on each copy, then each apply worktree — the fenced command
//     lines of <repo>/.flow/project.md's `## stop` body (the key's shape is
//     project-configuration.md's), run as one script with the
//     worktree as its directory under `bash -o pipefail -c`, bounded at 60
//     seconds (own process group, SIGTERM, a 2-second grace, SIGKILL); a file,
//     key or fence that is absent means no command is declared, the check is
//     skipped, and nothing outside the fence is ever run. Then the sibling
//     flow-guard check-worktree-processes.
//   `worktree remove --force` for each copy, then each apply worktree;
//   `worktree prune`; `branch -d spectre/<name>` (never -D) when it exists,
//     its upstream first set to origin/<base> (the base check 3's
//     resolve-base-branch printed; with every worktree already gone, the
//     base recorded as branch.spectre/<name>.flowBase, else origin/HEAD's)
//     so -d judges against where the change landed, not a pruned
//     origin/spectre/<name> or the not-yet-refreshed main checkout; a
//     refused -d puts the branch's previous upstream back, or unsets it
//     when it had none;
//   and `push origin --delete spectre/<name>`, not gated on the local steps —
//   a "remote ref does not exist" refusal is already gone, and prunes the
//   stale tracking ref with `fetch --prune`.
//
// Prints to stdout, one line each:
//   REFUSED: <worktree> — check <n>: <reason>   a check failed (gate)
//   HELD: <worktree> — <pid> <cwd>[; ...]        check 6 found a live process
//   UNCLASSIFIED: <worktree> — <path>            check 4, one per entry
//   UNCLASSIFIED: <worktree> — <path> — preserved: identical in <main-checkout>
//   UNCLASSIFIED: <worktree> — <path> — preserved: differs from <main-checkout>
//   UNCLASSIFIED: <worktree> — <path> — preserved: linked to <main-checkout>
//   REGENERATABLE: <worktree> — <count>          check 4
//   SKIPPED: check 5 — ## stop declares no fenced command   a declared `## stop`
//     whose body carries no fence; check 5 is skipped, never silently
//   DISCLOSE: <copy> — status: <line> / commit: <line>, and a final
//   DISCLOSE: stopped before removal — …          the disclosure stop
//   REMOVED: <worktree> | <worktree> — already gone | spectre/<name>
//   REFUSED: <worktree> — git worktree remove: <git's message>
//   REFUSED: spectre/<name> — git branch -d: <git's message>
//   REMOTE-DELETED: origin/spectre/<name>
//   REMOTE-GONE: origin/spectre/<name>
//   REMOTE-REFUSED: origin/spectre/<name> — <git's message>
//
// Exit 0 every check passed and every removal succeeded (whatever the REMOTE-*
// line says — it is reported, and flow-guard check-cleanup-complete verifies it);
// 1 the save gate or a check failed — nothing removed — or a removal failed — its entry stays
// in the state file's `worktrees`; 2 cannot answer, with NOTHING on stdout: a
// usage error, a name that is not a plain change name, a <repo> that is not a
// git repository, its worktree list unreadable, or flowd unreachable at the
// save gate; 3 the disclosure stop —
// nothing removed, nothing run: relay, judge each unpreserved entry, ask the
// one ask, then call again with --proceed.
//
// Run it from outside every worktree it removes: check 6 counts the caller's
// own shell.

// removeChangeWorktrees is `flow-guard remove-change-worktrees`: the comment above is the contract. It runs Worktree cleanup
// (skills/flow-contracts/finish-contract-run2.md) for one repository: the
// six checks on every apply worktree of spectre/<name> and checks 5-6 on
// every wave-group copy, all before anything is removed; then the removals,
// the local branch and the remote branch. Check 5 reads the `## stop` key in
// project-configuration.md's shape, through kwFenced as `## worktree setup` is.
func init() { Registry["remove-change-worktrees"] = removeChangeWorktrees }

// rcwRegenDirs are the path components check 4 counts as regeneratable — an
// image under one included, since it is that directory's, not a loose capture.
var rcwRegenDirs = []string{"build", ".gradle", ".kotlin", "node_modules", "dist", ".next", "target", "out", "coverage", "test-results", "__pycache__"}

func removeChangeWorktrees(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "remove-change-worktrees: "+format+"\n", a...)
		return 2
	}
	proceed := slices.Contains(args, "--proceed")
	args = slices.DeleteFunc(slices.Clone(args), func(a string) bool { return a == "--proceed" })
	if len(args) != 3 {
		return refuse("usage: flow-guard remove-change-worktrees <repo> <name> <merge-base|-> [--proceed]")
	}
	repo, name, mergeBase := pcAbs(env, args[0]), args[1], args[2]
	if !plainChangeName(name) {
		return refuse("change name %q is not a plain change name", name)
	}
	if mergeBase != "-" && strings.HasPrefix(mergeBase, "-") {
		return refuse("merge base %q is not a revision", mergeBase)
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
	// The main checkout is the porcelain's first entry, whatever <repo> named:
	// an unclassified entry with a counterpart at the same path there — a
	// per-worktree copy of a main-checkout local config, or a symlink to
	// it — is preserved.
	mainCheckout, _ := strings.CutPrefix(strings.SplitN(string(porcelain), "\n", 2)[0], "worktree ")
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

	// The save gate (KAN-1073): the change directory and .superpowers/ are
	// gitignored, so a worktree holding the plan (spectre/changes/<name>/ or
	// /flow-fast's .superpowers/sdd/<name>/tasks.md) or the self-review
	// bundle holds its only copy until `flow change save` keeps it in flowd
	// (cmd/flow/change.go's changeSavedFiles). Refused before anything else
	// when flowd holds none of a file the worktree holds.
	saved := []struct{ on, as string }{
		{"spectre/changes/" + name + "/tasks.md", "tasks.md"},
		{".superpowers/sdd/" + name + "/tasks.md", "tasks.md"},
		{".superpowers/sdd/" + name + "/self-review-context.md", "self-review-context.md"},
	}
	for _, wt := range wts {
		for _, f := range saved {
			if !isFile(wt + "/" + f.on) {
				continue
			}
			switch _, errOut, rc := pcFlow(env, false, "change", "file", "-C", wt, name, f.as); rc {
			case 0:
			case 1:
				fmt.Fprintf(stderr, "remove-change-worktrees: flow change save %s has not run — %s holds the only copy of its %s\n", name, wt, f.on)
				return 1
			default:
				return refuse("cannot tell whether flow change save %s has run: %s", name, rcwFlat(string(errOut)))
			}
		}
	}

	failed := false
	fail := func(wt string, check int, format string, a ...any) {
		failed = true
		fmt.Fprintf(stdout, "REFUSED: %s — check %d: %s\n", wt, check, rcwFlat(fmt.Sprintf(format, a...)))
	}
	unpreserved := false
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
		playwright := rcwPlaywrightCheckout(wt)
		for _, p := range strings.Split(string(out), "\x00") {
			switch {
			case p == "":
			case rcwRegeneratable(p), rcwCaptureOutput(playwright, p):
				regen++
			case wt != mainCheckout && !strings.Contains(p, "/") && rcwCounterpart(wt+"/"+p, mainCheckout+"/"+p):
				how := "differs from"
				if rcwSameBytes(wt+"/"+p, mainCheckout+"/"+p) {
					how = "identical in"
				}
				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s — preserved: %s %s\n", wt, rcwFlat(p), how, rcwFlat(mainCheckout))
			case wt != mainCheckout && !strings.Contains(p, "/") && rcwLinkToCounterpart(wt+"/"+p, mainCheckout+"/"+p):
				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s — preserved: linked to %s\n", wt, rcwFlat(p), rcwFlat(mainCheckout))
			default:
				unpreserved = true
				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s\n", wt, rcwFlat(p))
			}
		}
		fmt.Fprintf(stdout, "REGENERATABLE: %s — %d\n", wt, regen)
	}
	if failed {
		return 1
	}

	// The disclosure stop: the run relays, judges each unpreserved
	// unclassified entry, asks its one ask, and calls again with --proceed.
	// A preserved entry is shown and never stops the removal.
	if !proceed && (unpreserved || len(copies) > 0) {
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
	all := append(slices.Clone(copies), wts...)
	for _, wt := range all {
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
	}
	for _, wt := range all {
		procs := siblingCmd(env, "FLOW_GUARD_CHECK_WORKTREE_PROCESSES", "check-worktree-processes", wt)
		procs.Stderr = stderr
		verdict, err := procs.Output()
		switch v := string(verdict); {
		case err == nil && strings.HasPrefix(v, "CLEAR:"):
		case err == nil && strings.HasPrefix(v, "HELD:"):
			failed = true
			_, _ = io.WriteString(stdout, v)
		default:
			fail(wt, 6, "flow-guard check-worktree-processes could not answer")
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
// loose anywhere else matches no rule below. Python (__pycache__, *.pyc) and
// TypeScript (*.tsbuildinfo) compiler output is regeneratable wherever it
// sits, and so is a Finder .DS_Store. The ignored change directory
// (spectre/changes/) and .superpowers/ are expected content: the save gate
// above has already kept what of them outlives the worktree. Everything the
// list does not name stays unclassified.
func rcwRegeneratable(p string) bool {
	parts := strings.Split(p, "/")
	if parts[len(parts)-1] == ".DS_Store" || strings.HasPrefix(p, ".superpowers/") || strings.HasPrefix(p, "spectre/changes/") || strings.HasPrefix(p, ".dev-stack/") || strings.HasSuffix(p, ".log") || strings.HasSuffix(p, ".pyc") || strings.HasSuffix(p, ".tsbuildinfo") {
		return true
	}
	for _, c := range parts {
		if slices.Contains(rcwRegenDirs, c) {
			return true
		}
	}
	return false
}

// rcwCaptureOutput is check 4's one worktree-aware regeneratable rule: an
// ignored .png at the root of a Playwright checkout (a playwright.config.* at
// the worktree root) is a screenshot spec's output, which a re-run of the
// spec writes again; a tracked baseline never reaches check 4, and a .png in
// any subdirectory (a kept screenshots/run-* folder) stays unclassified.
func rcwCaptureOutput(playwright bool, p string) bool {
	return playwright && !strings.Contains(p, "/") && strings.EqualFold(filepath.Ext(p), ".png")
}

// rcwPlaywrightCheckout reports whether wt has a playwright.config.* at its
// root — looked up once per worktree, not once per entry.
func rcwPlaywrightCheckout(wt string) bool {
	configs, _ := filepath.Glob(wt + "/playwright.config.*")
	return len(configs) > 0
}

// rcwCounterpart reports whether a and b are two distinct regular files —
// check 4's "preserved", for an entry at the worktree root: the entry has a
// counterpart in the main checkout, a per-worktree copy of a main-checkout
// local config (kan-999: a worktree's .env with its own ports). One file reached by both paths (a symlinked
// directory in the main checkout pointing into the worktree) is not a
// counterpart: removing the worktree destroys it.
func rcwCounterpart(a, b string) bool {
	fa, errA := os.Lstat(a)
	fb, errB := os.Lstat(b)
	return errA == nil && errB == nil && fa.Mode().IsRegular() && fb.Mode().IsRegular() && !os.SameFile(fa, fb)
}

// rcwLinkToCounterpart reports whether a is a symlink resolving to b, a
// regular file — check 4's "preserved" for a main-checkout local config a
// project's dev tooling links into the worktree (kan-1026: gymie's
// linkEnvFile .env). Removing the worktree removes the link alone.
func rcwLinkToCounterpart(a, b string) bool {
	la, errL := os.Lstat(a)
	fa, errA := os.Stat(a)
	fb, errB := os.Lstat(b)
	return errL == nil && errA == nil && errB == nil && la.Mode()&os.ModeSymlink != 0 && fb.Mode().IsRegular() && os.SameFile(fa, fb)
}

// rcwSameBytes reports whether a and b hold the same bytes — which of the
// two preserved labels check 4 prints.
func rcwSameBytes(a, b string) bool {
	da, errA := os.ReadFile(a)
	db, errB := os.ReadFile(b)
	return errA == nil && errB == nil && bytes.Equal(da, db)
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
