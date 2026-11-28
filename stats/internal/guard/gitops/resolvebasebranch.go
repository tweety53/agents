package gitops

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The contract below moved here from the header of the deleted
// scripts/resolve-base-branch.sh shim (KAN-1058).
//
// flow-guard resolve-base-branch — the single owner of base-branch resolution.
//
// Usage: flow-guard resolve-base-branch [--fetch] <dir>
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
// WHY THE FETCH IS WRAPPED. A call to origin — the `fetch`, the `remote show
// origin` fallback, the `ls-remote --exit-code` — against an unreachable or
// stalled host blocks for as long as the connection does: roughly 75s on the
// default TCP timeout, with no end at all over an ssh connection that stalls
// without closing. A live panel or verifier begin runs this guard for each of
// the change's worktrees (stats/cmd/flow/record.go), so such a stall held the
// begin with it (kan-1132 self-review, F2). Each of the three calls is
// therefore bounded at rbbRemoteTimeout (20s): its process group — git and its
// ssh child — is killed at the bound, and the call counts as failed, exactly
// as an unreachable remote does. `-c core.askpass=true` stops it prompting for
// credentials, and a failed fetch is not this guard's failure — a stale
// origin/HEAD is still resolved, just possibly out of date, which the fetch on
// the next invocation corrects.
//
// WHY THE FETCH IS SKIPPED. A full fetch of origin that left a non-empty
// FETCH_HEAD younger than originFreshFor (60s), in any of its worktrees,
// stands in for this call's own: origin's refs are then no older than that.
// A single-ref fetch or another remote's leaves the rest stale, so it never
// counts.
// A failed fetch truncates FETCH_HEAD, so it never counts, and the next call
// fetches as before. `--fetch` always fetches: a caller about to push or
// rebase onto origin/<base> (integrate-change's land and cleanup, the
// finish contract's sync) passes it, so a 60s-old base is never landed on.
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
// The stalled-remote case is covered (TestResolveBaseBranchBoundsAStalledRemote)
// by a GIT_SSH_COMMAND stub that sleeps, which stalls deterministically where
// a TCP timeout does not: the test shortens the bound and asserts the guard
// still resolves the recorded base, well inside the stub's stall.

// resolveBaseBranch is `flow-guard resolve-base-branch`: the comment above is the contract -- the bare branch name alone on stdout at exit 0;
// 1 a named refusal, 2 cannot answer, 3 no `origin` remote, each with its
// reason on stderr and stdout empty. The reasoning for each branch, moved
// here from the bash body it replaced (c5379c0a), sits beside the code it
// explains.
func init() {
	core.Registry["resolve-base-branch"] = ResolveBaseBranch
}

// rbbRemoteTimeout bounds each call resolve-base-branch makes to origin.
var rbbRemoteTimeout = 20 * time.Second

func ResolveBaseBranch(args []string, env core.Env, stdout, stderr io.Writer) int {
	// NO REFUSAL BELOW MAY DEPEND ON THE OPERATOR'S LOCALE. The bash pinned
	// `export LC_ALL=C` because `A-Za-z0-9` in a `case` bracket expression is
	// a collating range: under en_US.UTF-8 on bash 3.2 a non-ASCII byte such
	// as `é` collated inside it and was admitted. The validation here
	// compares bytes, so no locale reaches it; git still runs under
	// LC_ALL=C, as it did as a child of the bash, so the `HEAD branch:` line
	// `remote show` prints is never translated.
	git := core.EnvGit(env, "LC_ALL=C")

	force := len(args) > 0 && args[0] == "--fetch"
	if force {
		args = args[1:]
	}
	dir := ""
	if len(args) > 0 {
		dir = args[0]
	}
	if dir == "" {
		fmt.Fprintln(stderr, "usage: flow-guard resolve-base-branch [--fetch] <dir>")
		core.Reason(env, "usage")
		return 2
	}
	if !core.IsDir(core.CleanAbs(env, dir)) {
		fmt.Fprintf(stderr, "resolve-base-branch: %s is not a directory\n", dir)
		core.Reason(env, "not-a-directory")
		return 2
	}
	if git("-C", dir, "rev-parse", "--git-dir").Run() != nil {
		fmt.Fprintf(stderr, "resolve-base-branch: %s is not a git worktree\n", dir)
		core.Reason(env, "not-a-worktree")
		return 2
	}
	// Exit 3 stays separate from exit 2: the finish contract requires a
	// distinct no-remote verdict a caller can tell apart without matching
	// stderr text.
	if git("-C", dir, "remote", "get-url", "origin").Run() != nil {
		fmt.Fprintf(stderr, "resolve-base-branch: no 'origin' remote configured in %s\n", dir)
		core.Reason(env, "no-origin-remote")
		return 3
	}

	// A failed fetch is not this guard's failure (the header's WHY THE FETCH
	// IS WRAPPED): a stale origin/HEAD still resolves, and the next
	// invocation's fetch corrects it. A recent fetch stands in for it (the
	// header's WHY THE FETCH IS SKIPPED) and, having succeeded, counts as
	// one.
	fetched := true
	if force || !fetchedRecently(git, dir) {
		_, fetched, _ = rbbRemote(git("-C", dir, "-c", "core.askpass=true", "fetch", "--quiet", "origin"))
	}

	// Each resolution attempt is allowed to fail -- a missing origin/HEAD or
	// an unreachable remote is an ordinary outcome here -- so only its output
	// is read, as the bash's `|| true` read it: each line through the bash's
	// sed, NUL bytes dropped and trailing newlines stripped as `$(...)` did.
	symref, _ := core.Capture(git("-C", dir, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"))
	base := rbbSed(symref, func(l string) (string, bool) {
		return strings.TrimPrefix(l, "origin/"), true
	})
	if base == "" {
		show, _, _ := rbbRemote(core.EnvGit(env, "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")("-C", dir, "remote", "show", "origin"))
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
	cur, ok := core.Capture(git("-C", dir, "branch", "--show-current"))
	if !ok {
		fmt.Fprintf(stderr, "resolve-base-branch: could not read the current branch in %s\n", dir)
		core.Reason(env, "cannot-read-current-branch")
		return 2
	}
	cur = strings.TrimRight(strings.ReplaceAll(cur, "\x00", ""), "\n")
	if cur == "" {
		fmt.Fprintf(stderr, "resolve-base-branch: HEAD is detached in %s\n", dir)
		core.Reason(env, "head-detached")
		return 1
	}
	// A base recorded on this branch when its worktree was created
	// (kickoff-worktree.sh <project> <name> <base>) wins over origin/HEAD:
	// a change cut from a feature branch lands back on it. It is read
	// after the fetch so the remote resolution above still warms
	// origin/<base>, and it faces the same refusals below.
	rec := RecordedBase(git, dir, cur)
	if rec != "" {
		base = rec
	}
	if base == "" {
		fmt.Fprintf(stderr, "resolve-base-branch: could not resolve a base branch in %s\n", dir)
		core.Reason(env, "could-not-resolve-a-base")
		return 1
	}
	if base == cur {
		fmt.Fprintf(stderr, "resolve-base-branch: base branch '%s' is the same as the current branch — refusing to compare a branch with itself\n", base)
		core.Reason(env, "base-is-current-branch")
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
			core.Reason(env, "base-name-invalid")
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
		if _, _, rc := rbbRemote(core.EnvGit(env, "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")("-C", dir, "-c", "core.askpass=true", "ls-remote", "--exit-code", "--heads", "origin", "refs/heads/"+base)); rc == 2 {
			fmt.Fprintf(stderr, "resolve-base-branch: recorded base '%s' is not on origin — re-point the change (skills/flow/SKILL.md)\n", base)
			core.Reason(env, "recorded-base-gone")
			return 1
		}
	}

	fmt.Fprintln(stdout, base)
	return 0
}

// originFreshFor is how long a fetch of origin stands in for the next one.
var originFreshFor = 60 * time.Second

// fetchedRecently reports whether a fetch in any worktree of dir's
// repository left a non-empty FETCH_HEAD younger than originFreshFor: the
// repository's own FETCH_HEAD or a linked worktree's, since each worktree
// writes its own while origin's refs are shared. Empty is a failed fetch.
// Only a full fetch of origin counts (fullOriginFetch): a single-ref
// `git fetch origin <ref>` or another remote's fetch leaves origin's other
// refs as old as they were.
func fetchedRecently(git func(...string) *exec.Cmd, dir string) bool {
	common, ok := core.Capture(git("-C", dir, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	if !ok {
		return false
	}
	heads, _ := filepath.Glob(filepath.Join(common, "worktrees", "*", "FETCH_HEAD"))
	var url string
	var branches []string
	for _, p := range append(heads, filepath.Join(common, "FETCH_HEAD")) {
		fi, err := os.Stat(p)
		if err != nil || fi.Size() == 0 {
			continue
		}
		if age := time.Since(fi.ModTime()); age < 0 || age >= originFreshFor {
			continue
		}
		body, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if url == "" {
			u, ok := core.Capture(git("-C", dir, "remote", "get-url", "origin"))
			refs, ok2 := core.Capture(git("-C", dir, "for-each-ref", "--format=%(refname:lstrip=3)", "refs/remotes/origin"))
			if !ok || !ok2 {
				return false
			}
			url = fetchURL(strings.TrimSpace(u))
			for _, b := range strings.Fields(refs) {
				if b != "HEAD" {
					branches = append(branches, b)
				}
			}
		}
		if fullOriginFetch(string(body), url, branches) {
			return true
		}
	}
	return false
}

// fullOriginFetch reports whether FETCH_HEAD body was written by a fetch of
// origin (every line's source is url) that listed each of origin's
// remote-tracking branches. A stale remote-tracking branch, deleted
// upstream and never pruned, makes it false: the guard then just fetches.
func fullOriginFetch(body, url string, branches []string) bool {
	listed := map[string]bool{}
	for _, line := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		f := strings.SplitN(line, "\t", 3)
		if len(f) != 3 {
			return false
		}
		desc, src, ok := strings.Cut(f[2], "' of ")
		if !ok || fetchURL(src) != url {
			return false
		}
		if b, ok := strings.CutPrefix(desc, "branch '"); ok {
			listed[b] = true
		}
	}
	for _, b := range branches {
		if !listed[b] {
			return false
		}
	}
	return len(listed) > 0
}

// fetchURL is url as git writes it into FETCH_HEAD: its user part dropped
// and a trailing "/" and ".git" stripped (git@host:x.git -> host:x).
func fetchURL(url string) string {
	if scheme, rest, ok := strings.Cut(url, "://"); ok {
		if at := strings.Index(rest, "@"); at >= 0 && at < strings.IndexAny(rest+"/", "/") {
			rest = rest[at+1:]
		}
		url = scheme + "://" + rest
	} else if at := strings.Index(url, "@"); at >= 0 && at < strings.IndexAny(url+":/", ":/") {
		url = url[at+1:]
	}
	url = strings.TrimRight(url, "/")
	return strings.TrimSuffix(url, ".git")
}

// recordedBase is the base kickoff-worktree.sh recorded on branch, or ""
// when none was: `git config branch.<branch>.flowBase`.
func RecordedBase(git func(...string) *exec.Cmd, dir, branch string) string {
	rec, _ := core.Capture(git("-C", dir, "config", "--get", "branch."+branch+".flowBase"))
	return strings.TrimRight(strings.ReplaceAll(rec, "\x00", ""), "\n")
}

// defaultBase is the merge base of HEAD with the change's base, resolved
// offline — the base recorded on the current branch
// (branch.<cur>.flowBase, resolve-base-branch.sh's first answer), else
// origin/HEAD, else origin/main, else main — so close-task never fetches. On
// the default branch itself the base is HEAD.
func defaultBase(git func(...string) *exec.Cmd, root string) (sha, label string, ok bool) {
	var candidates []string
	if cur, ok := core.Capture(git("-C", root, "branch", "--show-current")); ok && strings.TrimSpace(cur) != "" {
		if rec := RecordedBase(git, root, strings.TrimSpace(cur)); rec != "" {
			candidates = append(candidates, "origin/"+rec)
		}
	}
	if sym, ok := core.Capture(git("-C", root, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")); ok && sym != "" {
		candidates = append(candidates, sym)
	}
	candidates = append(candidates, "origin/main", "main")
	for _, c := range candidates {
		if mb, ok := core.Capture(git("-C", root, "merge-base", "HEAD", c)); ok && mb != "" {
			return mb, "merge-base(HEAD, " + c + ") " + mb[:min(12, len(mb))], true
		}
	}
	return "", "", false
}

// rbbRemote runs a git call that reaches origin, its process group killed
// once rbbRemoteTimeout elapses (rbbRunBounded), and returns its stdout as
// capture does, whether it succeeded, and its exit code. A call that timed
// out or never started reports failure and exit -1, as an unreachable
// remote does (kan-1132 self-review, F2).
func rbbRemote(cmd *exec.Cmd) (string, bool, int) {
	var out strings.Builder
	cmd.Stdout = &out
	if rbbRunBounded(cmd, rbbRemoteTimeout, nil) || cmd.ProcessState == nil {
		return "", false, -1
	}
	return strings.TrimRight(out.String(), "\n"), cmd.ProcessState.Success(), cmd.ProcessState.ExitCode()
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

// rbbRunBounded runs cmd, killing its whole process group (SIGTERM, then
// SIGKILL after 2s) once timeout elapses. The group kill is what reaches
// a test executor the harness forked; WaitDelay keeps an escaped
// descendant holding the capture pipe from hanging the wait.
func rbbRunBounded(cmd *exec.Cmd, timeout time.Duration, expire <-chan time.Time) (timedOut bool) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Start(); err != nil {
		return false
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	if expire == nil {
		bound := time.NewTimer(timeout)
		defer bound.Stop()
		expire = bound.C
	}
	select {
	case <-done:
		return false
	case <-expire:
	}
	pid := cmd.Process.Pid
	if syscall.Kill(-pid, syscall.SIGTERM) != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}
	select {
	case <-done:
		return true
	case <-time.After(2 * time.Second):
	}
	if syscall.Kill(-pid, syscall.SIGKILL) != nil {
		_ = cmd.Process.Kill()
	}
	<-done
	return true
}
