package guard

import (
	"fmt"
	"io"
	"strings"
)

// checkArchiveScope is scripts/check-archive-scope.sh: that script's header
// comment is the contract -- one OUT-OF-SCOPE line per staged path outside
// every <allowed-prefix>, then one SCOPE-OK/SCOPE-VIOLATION verdict line;
// exit 0 SCOPE-OK (an empty staged diff included), 1 SCOPE-VIOLATION, 2 with
// nothing on stdout when it cannot answer.
//
// WHY THIS GUARD EXISTS (KAN-472 follow-up, self-review finding). The finish
// contract's own step 4 stages the archive commit with `git -C
// <landing-worktree> add -A` — every path in the worktree, not only the
// archived change's own move. A landing worktree left stale by a skipped or
// failed fast-forward in step 2 stages and commits whatever else is sitting
// in that stale tree right alongside the archive move, silently. This is
// exactly what happened to kan-474's archive commit (`a574bf4`): it reverted
// skill-file content another change had shipped minutes earlier, because
// `add -A` swept it up along with the intended archive move. This guard
// makes that class of accident a refusal instead of a silent commit.
func init() { Registry["check-archive-scope"] = checkArchiveScope }

func checkArchiveScope(args []string, env Env, stdout, stderr io.Writer) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-archive-scope: "+format+"\n", a...)
		return 2
	}
	if len(args) < 2 {
		return die("usage: check-archive-scope.sh <worktree> <allowed-prefix> [<allowed-prefix> ...]")
	}
	root, ok := cdPhysical(env, args[0])
	if !ok {
		return die("%s is not a directory", args[0])
	}
	git := envGit(env)
	if git("-C", root, "rev-parse", "--is-inside-work-tree").Run() != nil {
		return die("%s is not a git worktree", root)
	}
	staged, ok := capture(git("-C", root, "diff", "--no-renames", "--cached", "--name-only"))
	if !ok {
		return die("cannot read the staged diff of %s", root)
	}

	var offenders []string
	for _, path := range strings.Split(staged, "\n") {
		if path != "" && !asInScope(path, args[1:]) {
			offenders = append(offenders, path)
		}
	}
	if len(offenders) > 0 {
		for _, path := range offenders {
			fmt.Fprintf(stdout, "OUT-OF-SCOPE: %s\n", path)
		}
		fmt.Fprintf(stdout, "SCOPE-VIOLATION: %s — %d\n", root, len(offenders))
		return 1
	}
	fmt.Fprintf(stdout, "SCOPE-OK: %s\n", root)
	return 0
}

// asInScope is whether path sits under one of prefixes. A PREFIX MATCH IS A
// PATH-COMPONENT MATCH, NOT A BARE STRING PREFIX: each prefix is normalized
// to end in exactly one `/` (the bash's `${prefix%/}/`, which strips one) before
// matching, so `spectre/changes/` never lets `spectre/changes-backup/x`
// through — the same distinction check-worktree-location.sh's own header
// draws for worktree roots.
func asInScope(path string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(path, strings.TrimSuffix(p, "/")+"/") {
			return true
		}
	}
	return false
}
