package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/refresh-plan-base.sh shim (KAN-1058).
//
// flow-guard refresh-plan-base — report what the base branch moved under a plan
// before its first task runs (skills/flow/implement.md, "The plan is
// refreshed against the base before task 1 runs").
//
// Usage: flow-guard refresh-plan-base <worktree> <merge-base> <tasks.md> [<spec-path>…]
//
// Fetches through resolve-base-branch, then compares <merge-base> with
// origin/<base>. Prints to stdout:
//   UNMOVED: <worktree> — origin/<base> has not moved since <merge-base>
// or, on a moved base:
//   MOVED: <worktree> — <n> commits on origin/<base> since <merge-base>
//   CHANGED: <path>          one per path `git diff --name-only <merge-base>
//                            origin/<base>` names that a **Files:** field of
//                            <tasks.md> declares, exactly or as a leading
//                            directory; `CHANGED: none of the plan's
//                            **Files:** paths` when there is none
//   ----- BEGIN <spec-path> @ origin/<base> -----
//   <the spec's text at origin/<base>>
//   ----- END <spec-path> -----
//                            per <spec-path>, or instead
//   SPEC-ABSENT: <spec-path> — not at origin/<base>
//
// The intersection is done in Go, not as a git pathspec: a multi-repo plan's
// **Files:** names paths outside this repository, which git refuses.
//
// Exit 0 whenever it answered, UNMOVED or MOVED; exit 2 when it cannot
// answer — usage, a merge base that is not a commit, an unreadable
// <tasks.md>, or a base branch resolve-base-branch cannot resolve (its own
// reason precedes this guard's on stderr). It names, it never rebases: the
// change branch is rebased onto the base right after it, by flow-guard sync-panel-base
// --rebase (skills/flow/implement.md).

// refreshPlanBase is `flow-guard refresh-plan-base`: the comment above is the contract -- after resolve-base-branch's fetch, one UNMOVED
// line when origin/<base> gained nothing since the merge base; otherwise a
// MOVED line, one CHANGED line per path the base changed that the plan's
// **Files:** names (or `CHANGED: none of …`), and each spec argument printed
// at origin/<base> between BEGIN/END lines or reported SPEC-ABSENT. Exit 0
// whenever it answered, 2 when it cannot. It never rebases.
func init() {
	Registry["refresh-plan-base"] = refreshPlanBase
}

func refreshPlanBase(args []string, env Env, stdout, stderr io.Writer) int {
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "refresh-plan-base: "+format+"\n", a...)
		return 2
	}
	if len(args) < 3 || args[0] == "" || args[1] == "" || args[2] == "" {
		return refuse("usage: flow-guard refresh-plan-base <worktree> <merge-base> <tasks.md> [<spec-path>…]")
	}
	wt, mb, tasks, specs := args[0], args[1], args[2], args[3:]
	git := envGit(env)

	if _, ok := capture(git("-C", wt, "rev-parse", "--verify", "--quiet", "--end-of-options", mb+"^{commit}")); !ok {
		return refuse("%s is not a commit in %s", mb, wt)
	}
	plan, err := os.ReadFile(pcAbs(env, tasks))
	if err != nil {
		return refuse("cannot read %s: %v", tasks, err)
	}
	// resolve-base-branch does the fetch; its refusal is this guard's
	// cannot-answer, its reason passed through on stderr.
	var name bytes.Buffer
	if rc := resolveBaseBranch([]string{wt}, env, &name, stderr); rc != 0 {
		return refuse("could not resolve the base branch of %s (resolve-base-branch exit %d)", wt, rc)
	}
	ref := "origin/" + strings.TrimSpace(name.String())

	n, ok := capture(git("-C", wt, "rev-list", "--count", mb+".."+ref))
	if !ok {
		return refuse("cannot count %s..%s in %s", mb, ref, wt)
	}
	if n == "0" {
		fmt.Fprintf(stdout, "UNMOVED: %s — %s has not moved since %s\n", wt, ref, mb)
		return 0
	}
	changed, ok := capture(git("-C", wt, "diff", "--no-renames", "--name-only", mb, ref))
	if !ok {
		return refuse("cannot diff %s %s in %s", mb, ref, wt)
	}

	fmt.Fprintf(stdout, "MOVED: %s — %s commits on %s since %s\n", wt, n, ref, mb)
	// Intersected here rather than as a git pathspec: a multi-repo plan's
	// **Files:** names paths outside this repository, which git refuses.
	files := rpbFiles(plan)
	hits := 0
	for _, p := range lines([]byte(changed)) {
		if p != "" && rpbNamed(p, files) {
			fmt.Fprintf(stdout, "CHANGED: %s\n", p)
			hits++
		}
	}
	if hits == 0 {
		fmt.Fprintln(stdout, "CHANGED: none of the plan's **Files:** paths")
	}
	for _, s := range specs {
		body, ok := capture(git("-C", wt, "show", ref+":"+s))
		if !ok {
			fmt.Fprintf(stdout, "SPEC-ABSENT: %s — not at %s\n", s, ref)
			continue
		}
		fmt.Fprintf(stdout, "----- BEGIN %s @ %s -----\n%s\n----- END %s -----\n", s, ref, body, s)
	}
	return 0
}

// rpbFiles is every backticked path on a **Files:** field, the field running
// on across wrapped lines until a blank line or the next **field**.
func rpbFiles(plan []byte) []string {
	var files []string
	in := false
	for _, l := range lines(plan) {
		switch {
		case strings.HasPrefix(l, "**Files:**"):
			in = true
		case in && (strings.TrimSpace(l) == "" || strings.HasPrefix(l, "**")):
			in = false
		}
		if !in {
			continue
		}
		for _, m := range pcBacktick.FindAllString(l, -1) {
			files = append(files, strings.Trim(m, "`"))
		}
	}
	return files
}

// rpbNamed: p is a declared path, or lies under one as a leading directory.
func rpbNamed(p string, files []string) bool {
	for _, f := range files {
		if p == f || strings.HasPrefix(p, strings.TrimSuffix(f, "/")+"/") {
			return true
		}
	}
	return false
}
