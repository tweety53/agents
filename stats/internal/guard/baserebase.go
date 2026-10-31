package guard

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// rbOutcome is one rebaseOntoTip answer. kind is "rebased"; "conflict" --
// the rebase stopped part-way and is left in progress for in-place
// resolution; or "refused" -- git would not start it and no rebase is in
// progress. sha is the tip the rebase ran onto; unmerged is a conflict's
// unmerged paths.
type rbOutcome struct {
	kind     string
	sha      string
	unmerged []string
}

// rebaseOntoTip rebases wt onto ref's tip, resolved to a sha first so a fetch
// landing mid-run moves the ref and never the rebase. Every line git prints
// goes to stderr; the caller owns stdout. An error is a question it cannot
// answer: ref does not resolve, or the tree cannot be read after a failed
// rebase.
func rebaseOntoTip(env Env, wt, ref string, stderr io.Writer) (rbOutcome, error) {
	git := envGit(env)
	sha, ok := capture(git("-C", wt, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}"))
	if !ok {
		return rbOutcome{}, fmt.Errorf("base ref '%s' does not resolve in %s", ref, wt)
	}
	o := rbOutcome{sha: sha}
	cmd := git("-C", wt, "rebase", sha)
	cmd.Stdout, cmd.Stderr = stderr, stderr
	if cmd.Run() == nil {
		o.kind = "rebased"
		return o, nil
	}
	if rebasing, err := inProgress(env, wt, "rebase-merge", "rebase-apply"); err != nil {
		return o, err
	} else if rebasing {
		paths, ok := capture(git("-C", wt, "diff", "--no-renames", "--name-only", "--diff-filter=U"))
		if !ok {
			return o, fmt.Errorf("cannot list the unmerged paths in %s", wt)
		}
		o.kind = "conflict"
		for _, u := range strings.Split(paths, "\n") {
			if u != "" {
				o.unmerged = append(o.unmerged, u)
			}
		}
		return o, nil
	}
	o.kind = "refused"
	return o, nil
}

// inProgress reports whether any of markers (rebase-merge, MERGE_HEAD, …)
// exists at its --git-path in wt: a sequencer operation is in progress.
func inProgress(env Env, wt string, markers ...string) (bool, error) {
	for _, f := range markers {
		p, ok := capture(envGit(env)("-C", wt, "rev-parse", "--git-path", f))
		if !ok {
			return false, fmt.Errorf("git rev-parse --git-path %s refused in %s", f, wt)
		}
		if !strings.HasPrefix(p, "/") {
			p = wt + "/" + p
		}
		if _, err := os.Stat(smcAbs(env, p)); err == nil {
			return true, nil
		}
	}
	return false, nil
}
