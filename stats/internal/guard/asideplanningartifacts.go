package guard

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// asidePlanningArtifacts is scripts/aside-planning-artifacts.sh: that
// script's header comment is the contract -- `aside <worktree>` stashes the
// planning paths' uncommitted state under the marker, `restore <worktree>`
// pops it back only when the top stash carries the marker, one verdict line
// on stdout, exit 0 answered, 1 conflicted restore, 2 cannot answer with
// nothing on stdout. Ported byte for byte from the bash at ae805186.
func init() {
	Registry["aside-planning-artifacts"] = asidePlanningArtifacts
}

const (
	apaMarker = "aside-planning-artifacts"
	apaUsage  = "usage: aside-planning-artifacts.sh aside <worktree>\n" +
		"       aside-planning-artifacts.sh restore <worktree>\n"
)

func asidePlanningArtifacts(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(msg string) int {
		fmt.Fprintf(stderr, "aside-planning-artifacts.sh: %s\n", msg)
		return 2
	}
	if len(args) != 2 {
		fmt.Fprint(stderr, apaUsage)
		return 2
	}
	action, wt := args[0], args[1]
	if wt == "" || !isDir(smcAbs(env, wt)) {
		return refuse("not a readable directory: " + wt)
	}
	git := envGit(env)
	if git("-C", wt, "rev-parse", "--is-inside-work-tree").Run() != nil {
		return refuse("not a git repository: " + wt)
	}
	// gitOut is `x="$(git ...)"`: stdout captured with trailing newlines
	// stripped, git's own stderr passed through. gitLoud is `git ... 1>&2`:
	// both of git's streams on this guard's stderr, in the order git wrote
	// them (one writer, so exec shares one pipe).
	gitOut := func(a ...string) (string, bool) {
		cmd := git(a...)
		cmd.Stderr = stderr
		return capture(cmd)
	}
	gitLoud := func(a ...string) bool {
		cmd := git(a...)
		cmd.Stdout, cmd.Stderr = stderr, stderr
		return cmd.Run() == nil
	}

	// The planning pathspec: each path only when its directory exists, since
	// git refuses a stash whose pathspec names an absent directory. The leaf
	// probe takes wt as given, as the bash printed it in the both-roots note.
	// ponytail: specRootLeaf probes a relative wt against the process's own
	// directory, which is env.Dir under flow-guard; an in-process caller with
	// a relative wt and another env.Dir would probe elsewhere.
	leaf := specRootLeaf(wt, stderr)
	var paths []string
	for _, p := range []string{leaf + "/changes", "docs/superpowers"} {
		if isDir(smcAbs(env, wt) + "/" + p) {
			paths = append(paths, p)
		}
	}
	clean := fmt.Sprintf("PLANNING-ARTIFACTS-CLEAN: %s — nothing to set aside\n", wt)
	none := fmt.Sprintf("PLANNING-ARTIFACTS-NONE: %s — no aside stash on top\n", wt)
	if len(paths) == 0 && action == "aside" {
		fmt.Fprint(stdout, clean)
		return 0
	}
	pathspec := append([]string{"--"}, paths...)

	switch action {
	case "aside":
		dirt, ok := gitOut(append([]string{"-C", wt, "status", "--porcelain", "--untracked-files=normal"}, pathspec...)...)
		if !ok {
			return refuse("git status refused the planning-path check in: " + wt)
		}
		if dirt == "" {
			fmt.Fprint(stdout, clean)
			return 0
		}
		if !gitLoud(append([]string{"-C", wt, "stash", "push", "--include-untracked", "-m", apaMarker + ": planning paths set aside for a rebase"}, pathspec...)...) {
			return refuse("git stash push refused in: " + wt)
		}
		sha, ok := gitOut("-C", wt, "rev-parse", "--verify", "--short", "stash@{0}")
		if !ok {
			return refuse("the stash it just pushed does not resolve in: " + wt)
		}
		fmt.Fprintf(stdout, "PLANNING-ARTIFACTS-ASIDE: %s — %s\n", wt, apaCut(sha))
		return 0
	case "restore":
		// Never pop over an unresolved sequencer state; --git-path answers
		// relative to wt.
		for _, f := range []string{"rebase-merge", "rebase-apply", "MERGE_HEAD", "CHERRY_PICK_HEAD"} {
			p, ok := gitOut("-C", wt, "rev-parse", "--git-path", f)
			if !ok {
				return refuse("git rev-parse --git-path refused: " + f)
			}
			if !strings.HasPrefix(p, "/") {
				p = wt + "/" + p
			}
			if _, err := os.Stat(smcAbs(env, p)); err == nil {
				return refuse("restore refused — a rebase/merge/cherry-pick is still in progress in: " + wt)
			}
		}
		top, ok := gitOut("-C", wt, "stash", "list", "--format=%H %gs")
		if !ok {
			return refuse("git stash list refused in: " + wt)
		}
		top, _, _ = strings.Cut(top, "\n")
		if top == "" {
			fmt.Fprint(stdout, none)
			return 0
		}
		// `${top%% *}` and `${top#* }`: with no space, both are top whole.
		sha, subject, found := strings.Cut(top, " ")
		if !found {
			subject = top
		}
		if !strings.Contains(subject, apaMarker) {
			fmt.Fprint(stdout, none)
			return 0
		}
		if gitLoud("-C", wt, "stash", "pop") {
			fmt.Fprintf(stdout, "PLANNING-ARTIFACTS-RESTORED: %s — %s\n", wt, apaCut(sha))
			return 0
		}
		// git keeps the entry on a conflicted pop; git stash list is the
		// caller's recovery path.
		fmt.Fprintf(stdout, "PLANNING-ARTIFACTS-CONFLICT: %s — %s kept\n", wt, apaCut(sha))
		return 1
	}
	fmt.Fprint(stderr, apaUsage)
	return 2
}

// apaCut is `${sha:0:12}`.
func apaCut(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
