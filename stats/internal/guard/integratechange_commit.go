package guard

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

func (c *ic) commit() {
	c.openStage = "flow.land"
	c.loadSet()
	if !c.findCanonical() {
		c.die("no-worktree-in-the-set", "no worktree in the set holds %s's tasks.md", c.name)
	}
	c.reopenLand()
	c.mergeOrder()
	reshape := false
	for i := range c.wt {
		reshape = reshape || isFile(c.reshapeFile(c.wt[i]))
	}
	if reshape {
		for _, o := range c.order {
			f := c.reshapeFile(c.wt[o])
			if !isFile(f) {
				continue
			}
			rb := icReadTrimmed(f)
			ids, _ := c.git("-C", c.wt[o], "log", "--no-merges", "--format=%(trailers:key=Task-Id,valueonly)", rb+"..HEAD")
			for _, id := range icLines(ids) {
				if t := "Task-Id: " + strings.TrimSpace(id); strings.TrimSpace(id) != "" && !slices.Contains(c.squashed[o], t) {
					c.squashed[o] = append(c.squashed[o], t)
				}
			}
			// The reshape: every commit since the reshape base folded back into
			// the index; only HEAD moves.
			if !c.gitPass("-C", c.wt[o], "reset", "-q", "--soft", rb) {
				c.stopAs("reshape-failed", "reshape", "git reset --soft "+rb+" failed in "+c.wt[o]+" — report its output")
			}
		}
		render := c.cmd("/", "flow", "record", "render", "-C", c.main, "-change", c.name, "-kind", "ledger", "-repo", c.canon)
		render.Stdout = c.stdout
		if render.Run() != nil {
			c.say("LEDGER-RENDER-FAILED: a destination was refused or could not be written — report it")
		}
		for _, o := range c.order {
			f := c.reshapeFile(c.wt[o])
			if !isFile(f) {
				continue
			}
			rb := icReadTrimmed(f)
			subj := ""
			for _, s := range c.subjects {
				k, v, _ := strings.Cut(s, "=")
				if k == c.wt[o] || k == c.repo[o] {
					subj = v
				}
			}
			if subj == "" && c.linkOnly(o, rb) {
				subj = "chore(spectre): merge order"
			}
			if subj == "" && len(c.implPaths(c.wt[o], rb)) > 0 {
				c.stopAs("subject-not-given", "commit-two", "no implementation subject given for "+c.wt[o]+" — pass "+c.wt[o]+"=<subject>")
			}
			if !c.commitTwo(c.wt[o], subj, c.squashed[o]) {
				c.stopAs("commit-two-failed", "commit-two", "the implementation or plan commit failed in "+c.wt[o]+" — report its lines")
			}
			os.Remove(f)
		}
	}

	// The rewritten branch reaches the remote now, not at land: a run that
	// stops before land leaves the next plain branch-backup push a fast-forward.
	for _, o := range c.order {
		c.pushBranch(c.wt[o])
	}
	c.say("NEXT: flow-guard integrate-change land " + c.main + " " + c.name + " " + c.flags())
}

// icReadTrimmed is `"$(cat <f>)"`.
func icReadTrimmed(f string) string {
	b, _ := os.ReadFile(f)
	return strings.TrimRight(string(b), "\n")
}

// commitTwo is the guarded commit chain of skills/flow-contracts/git-boundaries.md
// in wt, refused (false, its verdict printed) where wt is a main checkout or
// not on spectre/<name>: it stages everything and commits it under subj — skipped, not
// failed, when nothing is staged — with `Flow-Key: flow.commit-two` and each
// trailer (a squashed commit's `Task-Id: <n>`) in its trailer block, then
// commits anything left under c.planMsg the same way. The change directory
// is in neither commit: `<spec root>/changes/` must be gitignored (false,
// named, otherwise), and `<spec root>/changes/<name>/` is first untracked,
// kept on disk, where an older tooling committed it. False on any git
// failure, its own lines passed through.
func (c *ic) commitTwo(wt, subj string, trailers []string) bool {
	if checkChangeWorktreeLocation([]string{wt, c.name}, c.env, c.stdout, c.stderr) != 0 {
		return false
	}
	changes := specRootLeaf(wt, c.stderr) + "/changes/"
	// A worktree with nothing to commit and no change file tracked needs no
	// ignore: neither commit can take the change directory.
	dirty, okStatus := c.gitQuiet("-C", wt, "status", "--porcelain")
	tracked, okTracked := c.gitQuiet("-C", wt, "ls-files", "--", changes+c.name)
	if okStatus && okTracked && dirty == "" && tracked == "" {
		return true
	}
	if _, ok := c.gitQuiet("-C", wt, "check-ignore", "-q", "--no-index", changes); !ok {
		fmt.Fprintf(c.stderr, "integrate-change: %s is not gitignored in %s — add it to .gitignore before integrating\n", changes, wt)
		return false
	}
	commit := func(msg ...string) bool {
		if _, staged := c.gitQuiet("-C", wt, "diff", "--cached", "--quiet"); staged {
			return true
		}
		return c.gitPass(append([]string{"-C", wt, "commit"}, msg...)...)
	}
	impl := []string{"-m", subj, "--trailer", "Flow-Key: flow.commit-two"}
	for _, t := range trailers {
		impl = append(impl, "--trailer", t)
	}
	return c.gitPass("-C", wt, "rm", "-r", "-q", "-f", "--cached", "--ignore-unmatch", "--", changes+c.name) &&
		c.gitPass("-C", wt, "add", "-A", "--", ".") && commit(impl...) &&
		c.gitPass("-C", wt, "add", "-A") && commit("-m", c.planMsg)
}
