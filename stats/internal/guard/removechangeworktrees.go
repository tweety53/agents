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

// The contract below moved here from the header of the deleted
// scripts/remove-change-worktrees.sh shim (KAN-1058).
//
// flow-guard remove-change-worktrees — run /flow run 2's Worktree cleanup
// (skills/flow-contracts/finish-contract-run2.md) for one repository: every
// check on every worktree before anything is removed, then the removals, the
// local branch and the remote branch.
//
// Usage: flow-guard remove-change-worktrees <repo> <name>
//
// The worktrees are the ones `git -C <repo> worktree list --porcelain` lists on
// refs/heads/spectre/<name>, plus every wave-group copy of one, detached or on
// its own flow-copy/ branch
// (`<worktree>-wave-group-<g>`). A listed worktree whose directory is gone is
// already removed: success, its registration pruned.
//
// First the save gate: an apply worktree holding a file `flow change save`
// keeps — the plan, spectre/changes/<name>/tasks.md or
// .superpowers/sdd/<name>/tasks.md, or .superpowers/sdd/<name>/self-review-context.md,
// each gitignored and so the only copy — is refused, exit 1, with
// "flow change save <name> has not run" when flowd holds no saved copy of that
// file for that worktree's project (`flow change file`); flowd unreachable is
// exit 2.
//
// In order, never removing anything before every check has passed:
//   checks 1-3 on each apply worktree — no uncommitted tracked change; no
//     untracked, unignored file; no commit that exists only here (the base
//     resolved fresh for THIS worktree, in-process as flow-guard resolve-base-branch
//     does). Ignored files go with `--force`, never asked about: everything
//     the run needs was committed earlier in it.
//   check 5 on each copy, then each apply worktree — the fenced command
//     lines of <repo>/.flow/project.md's `## stop` body (the key's shape is
//     project-configuration.md's), run as one script with the
//     worktree as its directory under `bash -o pipefail -c`, bounded at 60
//     seconds (own process group, SIGTERM, a 2-second grace, SIGKILL); a file,
//     key or fence that is absent means no command is declared, the check is
//     skipped, and nothing outside the fence is ever run.
//   `worktree remove --force` for each copy, then `branch -D
//     flow-copy/<its last path element>` when that branch exists, then each
//     apply worktree;
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
//   SKIPPED: check 5 — ## stop declares no fenced command   a declared `## stop`
//     whose body carries no fence; check 5 is skipped, never silently
//   REMOVED: <worktree> | <worktree> — already gone | spectre/<name>
//   REFUSED: <worktree> — git worktree remove: <git's message>
//   REFUSED: spectre/<name> — git branch -d: <git's message>
//   REMOTE-DELETED: origin/spectre/<name>
//   REMOTE-GONE: origin/spectre/<name>
//   REMOTE-REFUSED: origin/spectre/<name> — <git's message>
//
// Exit 0 every check passed and every removal succeeded (whatever the REMOTE-*
// line says — it is reported); 1 the save gate or a check failed — nothing
// removed — or a removal failed — its entry stays in the state file's
// `worktrees`; 2 cannot answer, with NOTHING on stdout: a usage error, a name
// that is not a plain change name, a <repo> that is not a git repository, its
// worktree list unreadable, or flowd unreachable at the save gate.

// removeChangeWorktrees is `flow-guard remove-change-worktrees`: the comment above is the contract. It runs Worktree cleanup
// (skills/flow-contracts/finish-contract-run2.md) for one repository: checks
// 1-3 and 5 on every apply worktree of spectre/<name> and check 5 on every
// wave-group copy, all before anything is removed; then the removals, the
// local branch and the remote branch. Check 5 reads the `## stop` key in
// project-configuration.md's shape, through kwFenced as `## worktree setup` is.
func init() { Registry["remove-change-worktrees"] = removeChangeWorktrees }

func removeChangeWorktrees(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(id, format string, a ...any) int {
		Reason(env, id)
		fmt.Fprintf(stderr, "remove-change-worktrees: "+format+"\n", a...)
		return 2
	}
	if len(args) != 2 {
		return refuse("usage", "usage: flow-guard remove-change-worktrees <repo> <name>")
	}
	repo, name := pcAbs(env, args[0]), args[1]
	if !plainChangeName(name) {
		return refuse("bad-change-name", "change name %q is not a plain change name", name)
	}
	// LC_ALL=C keeps the remote delete's "remote ref does not exist" in
	// English, the one git message this guard reads.
	git := envGit(env, "LC_ALL=C")
	if !isDir(repo) || git("-C", repo, "rev-parse", "--git-dir").Run() != nil {
		return refuse("not-a-git-repository", "%s is not a git repository", repo)
	}
	porcelain, err := git("-C", repo, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return refuse("cannot-list-the-worktrees", "cannot list the worktrees of %s", repo)
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
				Reason(env, "change-not-saved")
				return 1
			default:
				return refuse("change-save-unknown", "cannot tell whether flow change save %s has run: %s", name, rcwFlat(string(errOut)))
			}
		}
	}

	failed := false
	fail := func(wt string, check int, format string, a ...any) {
		failed = true
		fmt.Fprintf(stdout, "REFUSED: %s — check %d: %s\n", wt, check, rcwFlat(fmt.Sprintf(format, a...)))
	}
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
	}
	if failed {
		Reason(env, "removal-failed")
		return 1
	}

	// 5 runs on every worktree and every copy, and stays a gate.
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
	if failed {
		Reason(env, "removal-failed")
		return 1
	}

	// Then, and only then: copies before the worktree they were copied from.
	for _, wt := range append(slices.Clone(copies), wts...) {
		if out, err := git("-C", repo, "worktree", "remove", "--force", wt).CombinedOutput(); err != nil {
			failed = true
			fmt.Fprintf(stdout, "REFUSED: %s — git worktree remove: %s\n", wt, rcwFlat(string(out)))
		} else {
			fmt.Fprintf(stdout, "REMOVED: %s\n", wt)
			if err := twDropCopyBranch(env, repo, wt); err != nil {
				failed = true
				fmt.Fprintf(stdout, "REFUSED: %s — %v\n", wt, err)
			}
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
		// The forge deleted it on merge. Prune the ref it left behind.
		if git("-C", repo, "fetch", "--prune", "--quiet", "origin").Run() != nil {
			fmt.Fprintf(stderr, "remove-change-worktrees: fetch --prune failed in %s — the stale refs/remotes/origin/%s may survive\n", repo, branch)
		}
		fmt.Fprintf(stdout, "REMOTE-GONE: origin/%s\n", branch)
	default:
		fmt.Fprintf(stdout, "REMOTE-REFUSED: origin/%s — %s\n", branch, rcwFlat(string(out)))
	}
	if failed {
		Reason(env, "remote-removal-failed")
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

// rcwFlat keeps a multi-line message on its verdict line, its control bytes
// escaped (ccSanitize): git output and paths are not this guard's to trust.
func rcwFlat(s string) string {
	return strings.ReplaceAll(strings.TrimRight(ccSanitize(s), "\n"), "\n", "; ")
}
