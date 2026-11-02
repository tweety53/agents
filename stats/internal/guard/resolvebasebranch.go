package guard

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/resolve-base-branch.sh shim (KAN-1058).
//
// flow-guard resolve-base-branch — the single owner of base-branch resolution.
//
// Usage: flow-guard resolve-base-branch <dir>
//
// stdout carries the bare branch name and NOTHING ELSE on success, so a
// caller composes it directly: BASE="$(flow-guard resolve-base-branch "$WT")". Every
// refusal writes its reason to stderr and leaves stdout empty — an empty
// stdout must never be readable as a resolved base.
//
//   Exit 0   Resolved; the name is on stdout.
//   Exit 1   A named refusal — detached HEAD, no base resolved, base equal
//            to the current branch, a base name that fails validation, or a
//            recorded base origin no longer carries.
//            Validation requires the first character to be one of
//            [A-Za-z0-9._] (so a leading '-' or a leading '/' is refused)
//            and every character in the name to be one of [A-Za-z0-9._/-].
//   Exit 2   Cannot answer — the argument is missing, <dir> is absent,
//            unreadable, or not a git worktree, or HEAD's own ref cannot be
//            read (corrupt or permission-denied).
//   Exit 3   The repository has no 'origin' remote at all.
//
// A RECORDED BASE WINS. When `git config branch.<current>.flowBase` is set —
// flow-guard kickoff-worktree writes it when a change is created with `--base
// <branch>` — that name is the base, ahead of origin/HEAD, and faces the same
// equal-to-current and name-validation refusals. A recorded base origin no
// longer carries — merged and deleted — is refused by name (`ls-remote
// --exit-code` status 2); an unreachable remote never refuses.
//
// WHY THE FETCH IS WRAPPED. `git remote show origin` against an unreachable
// host blocks for roughly 75s on the default TCP timeout, which would turn a
// correct refusal into a two-minute hang. `-c core.askpass=true` stops it
// prompting for credentials, `2>/dev/null` swallows the chatter, and `|| true`
// means a failed fetch is not this script's failure — a stale origin/HEAD is
// still resolved, just possibly out of date, which the fetch on the next
// invocation corrects.
//
// WHY HEAD@{upstream} IS NEVER CONSULTED. Inside an apply worktree, HEAD *is*
// the change's own branch, spectre/<name>, by construction. Its upstream is
// origin/spectre/<name>, so HEAD@{upstream} would compare the branch with
// itself the moment it is pushed — always "true", never useful. The base
// branch is resolved from origin/HEAD (or the `git remote show origin`
// fallback), never from HEAD's own upstream.
//
// WHY EXIT 3 IS SEPARATE FROM EXIT 2. The finish contract requires a distinct
// no-remote message. A caller that could only tell the two apart by matching
// stderr text would stop being able to the first time the wording changes.
//
// The unreachable-remote hang is deliberately NOT covered by this script's
// tests (TestResolveBaseBranch): a wall-clock assertion on a TCP timeout is flaky by
// construction. The wrapping above is carried forward verbatim from the
// fenced block this script replaces, and this comment is where its reasoning
// now lives.

// resolveBaseBranch is scripts/resolve-base-branch.sh: that script's header
// comment is the contract -- the bare branch name alone on stdout at exit 0;
// 1 a named refusal, 2 cannot answer, 3 no `origin` remote, each with its
// reason on stderr and stdout empty. The reasoning for each branch, moved
// here from the bash body it replaced (c5379c0a), sits beside the code it
// explains.
func init() {
	Registry["resolve-base-branch"] = resolveBaseBranch
}

func resolveBaseBranch(args []string, env Env, stdout, stderr io.Writer) int {
	// NO REFUSAL BELOW MAY DEPEND ON THE OPERATOR'S LOCALE. The bash pinned
	// `export LC_ALL=C` because `A-Za-z0-9` in a `case` bracket expression is
	// a collating range: under en_US.UTF-8 on bash 3.2 a non-ASCII byte such
	// as `é` collated inside it and was admitted. The validation here
	// compares bytes, so no locale reaches it; git still runs under
	// LC_ALL=C, as it did as a child of the bash, so the `HEAD branch:` line
	// `remote show` prints is never translated.
	git := envGit(env, "LC_ALL=C")

	dir := ""
	if len(args) > 0 {
		dir = args[0]
	}
	if dir == "" {
		fmt.Fprintln(stderr, "usage: flow-guard resolve-base-branch <dir>")
		return 2
	}
	if !isDir(pcAbs(env, dir)) {
		fmt.Fprintf(stderr, "resolve-base-branch: %s is not a directory\n", dir)
		return 2
	}
	if git("-C", dir, "rev-parse", "--git-dir").Run() != nil {
		fmt.Fprintf(stderr, "resolve-base-branch: %s is not a git worktree\n", dir)
		return 2
	}
	// Exit 3 stays separate from exit 2: the finish contract requires a
	// distinct no-remote verdict a caller can tell apart without matching
	// stderr text.
	if git("-C", dir, "remote", "get-url", "origin").Run() != nil {
		fmt.Fprintf(stderr, "resolve-base-branch: no 'origin' remote configured in %s\n", dir)
		return 3
	}

	// A failed fetch is not this guard's failure (the header's WHY THE FETCH
	// IS WRAPPED): a stale origin/HEAD still resolves, and the next
	// invocation's fetch corrects it.
	fetched := git("-C", dir, "-c", "core.askpass=true", "fetch", "--quiet", "origin").Run() == nil

	// Each resolution attempt is allowed to fail -- a missing origin/HEAD or
	// an unreachable remote is an ordinary outcome here -- so only its output
	// is read, as the bash's `|| true` read it: each line through the bash's
	// sed, NUL bytes dropped and trailing newlines stripped as `$(...)` did.
	symref, _ := capture(git("-C", dir, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"))
	base := rbbSed(symref, func(l string) (string, bool) {
		return strings.TrimPrefix(l, "origin/"), true
	})
	if base == "" {
		show, _ := capture(envGit(env, "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")("-C", dir, "remote", "show", "origin"))
		base = rbbSed(show, func(l string) (string, bool) {
			return strings.CutPrefix(strings.TrimLeft(l, " "), "HEAD branch: ")
		})
	}

	// `branch --show-current` failing means HEAD's own ref cannot be read --
	// corrupt or permission-denied -- a DIFFERENT fact from detached HEAD,
	// where the same command succeeds and prints nothing. Reported as
	// "detached" it would misdiagnose a tree resolution cannot read, so it is
	// refused as its own case, exit 2, like the other cannot-read-this-tree
	// checks above; exit 1 detached HEAD is reserved for the empty-but-
	// successful case.
	cur, ok := capture(git("-C", dir, "branch", "--show-current"))
	if !ok {
		fmt.Fprintf(stderr, "resolve-base-branch: could not read the current branch in %s\n", dir)
		return 2
	}
	cur = strings.TrimRight(strings.ReplaceAll(cur, "\x00", ""), "\n")
	if cur == "" {
		fmt.Fprintf(stderr, "resolve-base-branch: HEAD is detached in %s\n", dir)
		return 1
	}
	// A base recorded on this branch when its worktree was created
	// (kickoff-worktree.sh <project> <name> <base>) wins over origin/HEAD:
	// a change cut from a feature branch lands back on it. It is read
	// after the fetch so the remote resolution above still warms
	// origin/<base>, and it faces the same refusals below.
	rec := recordedBase(git, dir, cur)
	if rec != "" {
		base = rec
	}
	if base == "" {
		fmt.Fprintf(stderr, "resolve-base-branch: could not resolve a base branch in %s\n", dir)
		return 1
	}
	if base == cur {
		fmt.Fprintf(stderr, "resolve-base-branch: base branch '%s' is the same as the current branch — refusing to compare a branch with itself\n", base)
		return 1
	}

	// The base arrived from the remote and flows into refs a caller builds
	// (origin/<base>) and into further git commands. Refuse -- without
	// echoing the untrusted value back -- unless the first byte is one of
	// [A-Za-z0-9._] (no leading `-` a downstream git call would read as an
	// option) and every byte is one of [A-Za-z0-9._/-] (no control
	// character).
	for i := 0; i < len(base); i++ {
		if !rbbNameByte(base[i], i > 0) {
			fmt.Fprintln(stderr, "resolve-base-branch: the resolved base branch name is invalid")
			return 1
		}
	}

	// A recorded base origin no longer carries -- merged and deleted -- is
	// refused by name (KAN-911). Only `ls-remote --exit-code`'s status 2
	// (reached, no such head) refuses: an unreachable remote resolves as
	// before, and an origin/HEAD-resolved base never reaches this call. A
	// fetch that just failed already answered "unreachable", so the call is
	// skipped rather than blocking a second time.
	if rec != "" && fetched {
		err := envGit(env, "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")("-C", dir, "-c", "core.askpass=true", "ls-remote", "--exit-code", "--heads", "origin", "refs/heads/"+base).Run()
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 2 {
			fmt.Fprintf(stderr, "resolve-base-branch: recorded base '%s' is not on origin — re-point the change (skills/flow/SKILL.md)\n", base)
			return 1
		}
	}

	fmt.Fprintln(stdout, base)
	return 0
}

// recordedBase is the base kickoff-worktree.sh recorded on branch, or ""
// when none was: `git config branch.<branch>.flowBase`.
func recordedBase(git func(...string) *exec.Cmd, dir, branch string) string {
	rec, _ := capture(git("-C", dir, "config", "--get", "branch."+branch+".flowBase"))
	return strings.TrimRight(strings.ReplaceAll(rec, "\x00", ""), "\n")
}

// rbbNameByte is the C locale's [A-Za-z0-9._], plus `/` and `-` when rest.
func rbbNameByte(c byte, rest bool) bool {
	switch {
	case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '.', c == '_':
		return true
	case rest:
		return c == '/' || c == '-'
	}
	return false
}

// rbbSed runs a line edit over captured output as `sed` did per line --
// keep reports whether the line is printed (`sed -n …p`) -- and returns the
// result as `$(...)` captured it.
func rbbSed(out string, edit func(string) (string, bool)) string {
	if out == "" {
		return ""
	}
	var kept []string
	for _, l := range strings.Split(out, "\n") {
		if e, keep := edit(l); keep {
			kept = append(kept, e)
		}
	}
	return strings.TrimRight(strings.ReplaceAll(strings.Join(kept, "\n"), "\x00", ""), "\n")
}
