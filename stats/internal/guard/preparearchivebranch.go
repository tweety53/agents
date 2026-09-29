package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// prepareArchiveBranch is scripts/prepare-archive-branch.sh: that script's
// header comment is the contract and the authority for its state machine
// (steps 1-8, cited by skills/flow-contracts/finish-contract-run2.md) — exit 0
// positioned, one stdout line; 1 a named refusal; 2 cannot answer or post-run
// drift; 3 the base cannot be reconciled with origin. Every git call runs in
// the bash's order. The reasoning for each step, moved here from the bash body
// it replaced (d71a2327), sits beside the code it explains. Since KAN-823 no
// step that stops the chain fails silently: a failing git call's own stderr is
// printed with the guard's prefix before the named line, and the landing
// directory is asserted to be a git worktree of its own before any further
// git call runs in it.
func init() {
	Registry["prepare-archive-branch"] = prepareArchiveBranch
}

func prepareArchiveBranch(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	landing, base, archive := arg(0), arg(1), arg(2)
	if landing == "" || base == "" || archive == "" {
		fmt.Fprintln(stderr, "usage: prepare-archive-branch.sh <landing-worktree> <base> <archive-branch>")
		return 2
	}
	say := func(rc int, format string, a ...any) int {
		fmt.Fprintf(stderr, "prepare-archive-branch: "+format+"\n", a...)
		return rc
	}

	// Step 2: the same branch-name shape resolve-base-branch.sh enforces —
	// first character one of [A-Za-z0-9._], every character one of
	// [A-Za-z0-9._/-] — so neither name ever reaches a git call as something
	// that could be read as an option. Byte ranges: the bash pinned LC_ALL=C
	// because a bracket range under a UTF-8 locale collates.
	for _, n := range []struct{ name, label string }{{base, "base branch"}, {archive, "archive branch"}} {
		if !pabBranchName(n.name) {
			return say(1, "%s '%s' is not a valid branch name", n.label, n.name)
		}
	}

	// Paths are taken as the bash took them: relative to the caller's cwd,
	// uncleaned so the kernel resolves any "..", and every git call runs
	// from it. The bash's `export LC_ALL=C` is not carried to git: it pinned
	// the bash's bracket ranges, which are byte comparisons here. What git
	// prints that is parsed is locale-independent (porcelain status, branch
	// names, stash entries); the git stderr passed through (show-ref, the
	// tree snapshot and recompute) is in the caller's locale, where the bash
	// printed it in C's -- visible only with a localized git.
	gitDir := liveDir(env.Dir)
	git := func(stdout, stderr io.Writer, a ...string) (string, int) {
		return gitExec(gitDir, stdout, stderr, a...)
	}
	fsPath := func(p string) string { return smcAbs(env, p) }
	// gitLoud is git with the stderr captured and printed with the guard's
	// prefix when the call fails: no step that stops the chain fails silently
	// (KAN-823 — a landing-chain step whose output went to /dev/null let every
	// later step act on the wrong tree). Steps whose failure is not this
	// chain's failure stay on plain git: the bounded fetch is documented
	// best-effort, and the dirty-file classification degrades to unclassified
	// by its own contract.
	gitLoud := func(a ...string) (string, int) {
		var errBuf bytes.Buffer
		out, rc := git(io.Discard, &errBuf, a...)
		if rc != 0 {
			for _, line := range strings.Split(strings.TrimRight(errBuf.String(), "\n"), "\n") {
				if line != "" {
					fmt.Fprintf(stderr, "prepare-archive-branch: git: %s\n", line)
				}
			}
		}
		return out, rc
	}

	// Step 2b: <landing-worktree> is created, from the main checkout, when it
	// does not already exist. `--force` is required: the main checkout is
	// ordinarily already on <base>, and git refuses a second checkout of a
	// branch without it (the header has the reasoning).
	if _, err := os.Stat(fsPath(landing)); err != nil {
		parent := gdcDirname(landing)
		if pabBasename(parent) != ".worktrees" {
			return say(1, "%s's parent is not a .worktrees directory — refusing to create it", landing)
		}
		top, rc := gitLoud("-C", gdcDirname(parent), "rev-parse", "--show-toplevel")
		if rc != 0 {
			return say(2, "cannot resolve the main checkout above %s", landing)
		}
		mainCheckout := strings.TrimRight(top, "\n")
		// `worktree add` runs from the main checkout, so a relative landing
		// is made absolute against the caller's directory first — the bash
		// handed it over as given, creating it under the main checkout.
		if _, rc := gitLoud("-C", mainCheckout, "worktree", "add", "--force", "--quiet", "--", fsPath(landing), base); rc != 0 {
			return say(2, "could not create the landing worktree %s from '%s' in %s", landing, base, mainCheckout)
		}
	}

	if !isDir(fsPath(landing)) {
		return say(2, "%s is not a directory", landing)
	}
	// KAN-823's post-positioning assertion: the landing must be a git worktree
	// of its own before any further `git -C <landing>` runs. A plain directory
	// merely sitting inside the parent repository — the incident: a
	// `_landing-<name>` dir holding a daemon's output, after `git worktree
	// remove` had deregistered the worktree — walks up to that parent instead,
	// and every later git call would act on the wrong tree. `--show-prefix` is
	// empty exactly at a toplevel.
	prefix, rc := gitLoud("-C", landing, "rev-parse", "--show-prefix")
	if rc != 0 {
		return say(2, "%s is not a git worktree", landing)
	}
	// Step 2c: that resolution can come from walking UP out of the landing
	// directory — a `_landing-<name>` a daemon recreated as a plain directory
	// inside the main checkout's tree resolves there, and every later
	// `git -C <landing>` would act on the main checkout, whose index the run
	// then found stale (KAN-823's incident, caught only by the downstream
	// dirty-tree refusal). `--show-prefix` is empty exactly at a toplevel, so
	// a resolution that walked up carries a non-empty prefix and is refused
	// here, naming where git resolved it.
	if p := strings.TrimRight(prefix, "\n"); p != "" {
		top, trc := gitLoud("-C", landing, "rev-parse", "--show-toplevel")
		if trc == 0 {
			return say(2, "%s is not itself a git worktree — git resolves it to %s", landing, strings.TrimRight(top, "\n"))
		}
		return say(2, "%s is not itself a git worktree — it resolves inside the repository above it", landing)
	}
	// A worktree root carries its own .git entry; anything the resolution
	// reached by walking up does not — and a bare repository's directory
	// carries none either.
	if _, err := os.Stat(filepath.Join(fsPath(landing), ".git")); err != nil {
		top, _ := gitLoud("-C", landing, "rev-parse", "--show-toplevel")
		return say(2, "%s is not itself a git worktree — git resolves it to %s", landing, strings.TrimRight(top, "\n"))
	}
	// The landing worktree must belong to the repository it was cut from —
	// the main checkout above its .worktrees parent, the same resolution
	// step 2b creates it from, compared through the common directory both
	// sides print. A landing that is some other repository's worktree is
	// refused here rather than acted on. Off the _landing-<name>
	// construction there is no repository to compare against, and none is
	// guessed.
	if pabBasename(gdcDirname(landing)) == ".worktrees" {
		mainTop, rc := gitLoud("-C", gdcDirname(gdcDirname(landing)), "rev-parse", "--show-toplevel")
		if rc != 0 {
			return say(2, "cannot resolve the main checkout above %s", landing)
		}
		mainCheckout := strings.TrimRight(mainTop, "\n")
		common, rc := gitLoud("-C", landing, "rev-parse", "--path-format=absolute", "--git-common-dir")
		if rc != 0 {
			return say(2, "cannot read the git common directory of %s", landing)
		}
		common = strings.TrimRight(common, "\n")
		mainCommon, rc := gitLoud("-C", mainCheckout, "rev-parse", "--path-format=absolute", "--git-common-dir")
		if rc != 0 {
			return say(2, "cannot read the git common directory of %s", mainCheckout)
		}
		if common != strings.TrimRight(mainCommon, "\n") {
			return say(2, "%s is a worktree of a different repository (%s, not %s) — refusing", landing, common, strings.TrimRight(mainCommon, "\n"))
		}
	}
	if _, rc := gitLoud("-C", landing, "remote", "get-url", "origin"); rc != 0 {
		return say(3, "no 'origin' remote configured in %s — cannot resolve origin/%s", landing, base)
	}

	// Step 3: the bounded, credential-free fetch (WHY THE FETCH IS WRAPPED,
	// in the header): core.askpass=true stops a credential prompt, its
	// chatter is swallowed, and a failed fetch is not this guard's failure —
	// a stale origin/<base> is still usable.
	git(stdout, io.Discard, "-C", landing, "-c", "core.askpass=true", "fetch", "--quiet", "origin")

	// Step 4: `branch --show-current` failing is a different fact from a
	// detached HEAD, where it succeeds and prints nothing.
	cur, rc := gitLoud("-C", landing, "branch", "--show-current")
	if rc != 0 {
		return say(2, "could not read the current branch in %s", landing)
	}
	if cur = strings.TrimRight(cur, "\n"); cur == "" {
		return say(1, "HEAD is detached in %s", landing)
	}

	// Step 5: a dirty tree is refused wherever it is — on <base> the changes
	// would otherwise ride onto the archive branch unremarked — naming every
	// dirty entry, classified against the change's own branch. A status call
	// that fails stops the chain (KAN-823): a failed read is never a clean
	// tree. `-z`, because porcelain v1 otherwise quotes a path with a control
	// character, or a rename's path with a space, and a quoted path never
	// matches the change's own names.
	dirty, rc := gitLoud("-C", landing, "-c", "core.quotePath=false", "status", "--porcelain", "-z", "--untracked-files=normal")
	if rc != 0 {
		return say(2, "cannot read the working tree state of %s", landing)
	}
	if dirty != "" {
		if cur == base {
			say(1, "%s has a dirty working tree on '%s' — refusing", landing, base)
		} else {
			say(1, "%s is on '%s' with uncommitted changes, not '%s' — refusing", landing, cur, base)
		}
		pabReportDirty(landing, base, dirty, git, fsPath, stderr)
		return 1
	}

	// Snapshot the tree the branch moves start from — status and stash list —
	// so the post-run check can tell residue from the state the run left. A
	// failing snapshot ends the run with its status, as `set -e` did.
	snapshot, rc := snapshotTreeState(fsPath(landing), stderr)
	if rc != 0 {
		return rc
	}

	if cur != base {
		if _, rc := gitLoud("-C", landing, "checkout", "-q", base); rc != 0 {
			return say(2, "could not check out '%s' in %s", base, landing)
		}
	}

	// Step 6: fast-forward <base> — a no-op when local already contains
	// origin's tip, a failure only on a genuine divergence.
	if _, rc := gitLoud("-C", landing, "rev-parse", "-q", "--verify", "refs/remotes/origin/"+base); rc != 0 {
		return say(3, "origin/%s does not exist — cannot fast-forward '%s'", base, base)
	}
	if _, rc := gitLoud("-C", landing, "merge", "--ff-only", "-q", "origin/"+base); rc != 0 {
		return say(3, "'%s' cannot be fast-forwarded to origin/%s — it has diverged", base, base)
	}

	// Step 7: reuse an existing <archive-branch> descended from
	// origin/<base>, never recreate or reset it; refuse one that is not;
	// otherwise create it from the fast-forwarded <base>.
	if _, rc := git(stdout, stderr, "-C", landing, "show-ref", "--verify", "--quiet", "refs/heads/"+archive); rc == 0 {
		if _, rc := gitLoud("-C", landing, "merge-base", "--is-ancestor", "origin/"+base, archive); rc != 0 {
			return say(1, "'%s' already exists and is not descended from origin/%s — refusing", archive, base)
		}
		if _, rc := gitLoud("-C", landing, "checkout", "-q", archive); rc != 0 {
			return say(2, "could not check out existing '%s' in %s", archive, landing)
		}
	} else if _, rc := gitLoud("-C", landing, "checkout", "-q", "-b", archive, base); rc != 0 {
		return say(2, "could not create '%s' from '%s' in %s", archive, base, landing)
	}

	// Post-run self-check: any new stash entry or unexpected status line
	// since the snapshot is residue this run left behind, named here rather
	// than left for a conductor to retry blind (KAN-423's incident). The
	// recompute itself failing is exit 2, never a silent no-drift.
	drift, rc := checkTreeRestored(fsPath(landing), snapshot, stderr)
	if rc == 2 {
		return say(2, "cannot read the working tree state of %s after the branch moves", landing)
	}
	if len(drift) > 0 {
		say(2, "post-run drift detected:")
		for _, d := range drift {
			fmt.Fprintln(stderr, d)
		}
		return 2
	}

	// Step 8: the one success line.
	fmt.Fprintf(stdout, "%s -> %s\n", cur, archive)
	return 0
}

// pabBranchName is validate_branch_name's shape check.
func pabBranchName(n string) bool {
	word := func(c byte) bool {
		return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_'
	}
	if !word(n[0]) {
		return false
	}
	for i := 0; i < len(n); i++ {
		if !word(n[i]) && n[i] != '/' && n[i] != '-' {
			return false
		}
	}
	return true
}

// pabReportDirty is report_dirty_files: one stderr line per dirty entry, its
// porcelain status and path, then whether the path looks like this change's
// output. A renamed entry is classified by its new path. The change is
// resolved from the landing path itself: by construction
// <project>/.worktrees/_landing-<name>, with the apply worktree
// <project>/.worktrees/<name> on branch <name> beside it. When the leaf is not
// `_landing-<name>`, that worktree does not exist, the branch does not
// resolve, or its merge-base or diff with <base> fails, the entries are named
// without classification rather than guessed.
func pabReportDirty(landing, base, dirty string, git func(io.Writer, io.Writer, ...string) (string, int),
	fsPath func(string) string, stderr io.Writer) {
	branch, changed := "", []string(nil)
	if name, ok := strings.CutPrefix(pabBasename(landing), "_landing-"); ok && name != "" {
		apply := gdcDirname(landing) + "/" + name
		if isDir(fsPath(apply)) {
			if _, rc := git(io.Discard, io.Discard, "-C", apply, "rev-parse", "--git-dir"); rc == 0 {
				if _, rc := git(io.Discard, io.Discard, "-C", apply, "rev-parse", "-q", "--verify", "refs/heads/"+name); rc == 0 {
					if mb, rc := git(nil, io.Discard, "-C", apply, "merge-base", name, base); rc == 0 {
						diff, rc := git(nil, io.Discard, "-C", apply, "diff", "-z", "--no-renames",
							"--name-only", strings.TrimRight(mb, "\n"), name)
						if rc == 0 {
							branch, changed = name, strings.Split(diff, "\x00")
						}
					}
				}
			}
		}
	}
	fmt.Fprintln(stderr, "prepare-archive-branch: dirty files:")
	if branch == "" {
		fmt.Fprintf(stderr, "prepare-archive-branch:   (cannot classify -- no change worktree with a branch beside %s)\n", landing)
	}
	// `status --porcelain -z`: each record is `XY <path>`, NUL-terminated, and
	// a rename or copy's record is followed by one more holding the source
	// path. The entry is printed as porcelain v1 prints it unquoted.
	records := strings.Split(dirty, "\x00")
	for i := 0; i < len(records); i++ {
		entry := records[i]
		if entry == "" {
			continue
		}
		path := ""
		if len(entry) > 3 {
			path = entry[3:]
		}
		if strings.ContainsAny(entry[:min(2, len(entry))], "RC") && i+1 < len(records) {
			i++
			entry = entry[:3] + records[i] + " -> " + path
		}
		switch {
		case branch != "" && pabPathChanged(path, changed):
			fmt.Fprintf(stderr, "prepare-archive-branch:   %s -- looks like this change's output (changed on '%s')\n", entry, branch)
		case branch != "":
			fmt.Fprintf(stderr, "prepare-archive-branch:   %s -- does not look like this change's output (not changed on '%s')\n", entry, branch)
		default:
			fmt.Fprintf(stderr, "prepare-archive-branch:   %s\n", entry)
		}
	}
}

// pabPathChanged is path_changed: an exact, fully literal match against a
// changed path; a porcelain directory entry, which ends in `/`, also matches
// any changed path under it. Neither side is ever a pattern for the other.
func pabPathChanged(p string, changed []string) bool {
	for _, f := range changed {
		if f != "" && (f == p || strings.HasSuffix(p, "/") && strings.HasPrefix(f, p)) {
			return true
		}
	}
	return false
}

// pabBasename is basename(1).
func pabBasename(p string) string {
	if p != "" && strings.Trim(p, "/") == "" {
		return "/"
	}
	p = strings.TrimRight(p, "/")
	return p[strings.LastIndex(p, "/")+1:]
}
