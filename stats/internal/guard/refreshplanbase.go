package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// refreshPlanBase is scripts/refresh-plan-base.sh: that script's header
// comment is the contract -- after resolve-base-branch's fetch, one UNMOVED
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
		return refuse("usage: refresh-plan-base.sh <worktree> <merge-base> <tasks.md> [<spec-path>…]")
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
