package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// rbOutcome is one rebaseOntoTip answer. kind is "rebased"; "conflict" --
// the rebase stopped part-way and is left in progress for in-place
// resolution, the aside still stashed; or "refused" -- git would not start
// it, no rebase is in progress, and the aside is restored. sha is the tip the
// rebase ran onto; unmerged is a conflict's unmerged paths; aside is the
// stash this call set aside ("" when it set none), still stashed on a
// conflict.
type rbOutcome struct {
	kind     string
	sha      string
	unmerged []string
	aside    string
}

// rebaseOntoTip rebases wt onto ref's tip, resolved to a sha first so a fetch
// landing mid-run moves the ref and never the rebase. With aside, the
// planning paths are set aside for the rebase (aside-planning-artifacts, in
// process) and restored after it -- except on a conflict, where the aside
// stays until the resolution finishes. Every line git and the aside print
// goes to stderr; the caller owns stdout. An error is a question it cannot
// answer: ref does not resolve, the aside or its restore failed, or the tree
// cannot be read after a failed rebase.
func rebaseOntoTip(env Env, wt, ref string, aside bool, stderr io.Writer) (rbOutcome, error) {
	git := envGit(env)
	sha, ok := capture(git("-C", wt, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}"))
	if !ok {
		return rbOutcome{}, fmt.Errorf("base ref '%s' does not resolve in %s", ref, wt)
	}
	o := rbOutcome{sha: sha}
	// runAside runs one aside-planning-artifacts action with its verdict
	// line onto stderr, and reports whether it answered with verdict.
	if aside {
		var out bytes.Buffer
		code := asidePlanningArtifacts([]string{"aside", wt}, env, &out, stderr)
		fmt.Fprint(stderr, out.String())
		if code != 0 {
			return o, fmt.Errorf("the planning paths could not be set aside in %s", wt)
		}
		// PLANNING-ARTIFACTS-ASIDE: <worktree> — <stash sha>
		if line := strings.TrimSpace(out.String()); strings.HasPrefix(line, "PLANNING-ARTIFACTS-ASIDE:") {
			o.aside = line[strings.LastIndex(line, " ")+1:]
		}
	}
	// Only the stash this call pushed is restored, and a CLEAN aside pushed
	// none.
	restore := func() error {
		if o.aside == "" {
			return nil
		}
		return restoreAside(env, wt, o.aside, stderr)
	}

	cmd := git("-C", wt, "rebase", sha)
	cmd.Stdout, cmd.Stderr = stderr, stderr
	if cmd.Run() == nil {
		o.kind = "rebased"
		return o, restore()
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
	return o, restore()
}

// restoreAside pops the planning paths set aside as sha back into wt, and
// only while that stash is the top entry: the stash list is shared by every
// worktree of the repository, so the marker alone can name another
// worktree's aside.
func restoreAside(env Env, wt, sha string, stderr io.Writer) error {
	top, ok := capture(envGit(env)("-C", wt, "stash", "list", "-n", "1", "--format=%H"))
	if !ok {
		return fmt.Errorf("git stash list refused in %s", wt)
	}
	if !strings.HasPrefix(top, sha) {
		return fmt.Errorf("the planning paths set aside as %s are not the top stash in %s — git stash list holds them", sha, wt)
	}
	var out bytes.Buffer
	code := asidePlanningArtifacts([]string{"restore", wt}, env, &out, stderr)
	fmt.Fprint(stderr, out.String())
	if code != 0 || !strings.HasPrefix(out.String(), "PLANNING-ARTIFACTS-RESTORED:") {
		return fmt.Errorf("the planning paths could not be restored in %s — git stash list holds them", wt)
	}
	return nil
}

// asideRecord is the per-worktree file (under the worktree's own git dir) a
// call stopped on a conflict records its aside's stash sha, or none, in, so
// its resuming call restores exactly that stash and never another
// worktree's.
func asideRecord(env Env, wt, name string) (string, error) {
	p, ok := capture(envGit(env)("-C", wt, "rev-parse", "--git-path", name))
	if !ok {
		return "", fmt.Errorf("git rev-parse --git-path %s refused in %s", name, wt)
	}
	if !strings.HasPrefix(p, "/") {
		p = wt + "/" + p
	}
	return smcAbs(env, p), nil
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
