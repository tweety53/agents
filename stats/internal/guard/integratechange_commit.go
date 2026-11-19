package guard

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

func (c *ic) commit() {
	c.loadSet()
	if !c.findCanonical() {
		c.die("no-worktree-in-the-set", "no worktree in the set holds %s's tasks.md", c.name)
	}
	live := c.canon + "/" + c.spec + "/changes/" + c.name
	c.mergeOrder()
	reshape := false
	for i := range c.wt {
		reshape = reshape || isFile(c.reshapeFile(c.wt[i]))
	}
	if reshape {
		c.openStage = "flow.preserve-sessions"
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
		c.markEnd("flow.preserve-sessions", "completed")
		c.markBegin("flow.commit-two")
		for _, o := range c.order {
			f := c.reshapeFile(c.wt[o])
			if !isFile(f) {
				continue
			}
			rb := icReadTrimmed(f)
			if c.wt[o] == c.canon && isFile(live+"/design.md") {
				c.writeDesignRecord(live + "/design.md")
			}
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
		c.markEnd("flow.commit-two", "completed")
	}

	// The rewritten branch reaches the remote now, not at land: a run that
	// stops before land leaves the next plain branch-backup push a fast-forward.
	for _, o := range c.order {
		c.pushBranch(c.wt[o])
	}
	if done, _ := c.selfReviewCompleted(); done {
		c.say("NEXT: flow.self-review has already completed — skip the pass; flow-guard integrate-change land " + c.main + " " + c.name + " " + c.flags())
	} else {
		c.markBegin("flow.self-review")
		c.say("NEXT: run the self-review pass over `flow self-review bundle -C " + c.canon + " -change " + c.name + "` and end flow.self-review completed, then flow-guard integrate-change land " + c.main + " " + c.name + " " + c.flags())
	}
}

// icReadTrimmed is `"$(cat <f>)"`.
func icReadTrimmed(f string) string {
	b, _ := os.ReadFile(f)
	return strings.TrimRight(string(b), "\n")
}

// designRecord is `integrate-change design-record`: the design record alone,
// re-derived from the canonical change directory's design.md.
func (c *ic) designRecord() {
	c.loadSet()
	if !c.findCanonical() {
		c.die("no-worktree-in-the-set", "no worktree in the set holds %s's tasks.md", c.name)
	}
	design := c.canon + "/" + c.spec + "/changes/" + c.name + "/design.md"
	if !isFile(design) {
		c.die("no-design", "%s does not exist — no design record to write", design)
	}
	c.writeDesignRecord(design)
	c.say(c.canon + "/docs/designs/" + c.name + ".md")
}

// writeDesignRecord writes the change's design.md to docs/designs/<name>.md in
// the canonical worktree, under a fixed two-line header, for the
// implementation commit to carry (The design record,
// skills/flow-contracts/finish-contract-run1.md). commitTwo stages it.
func (c *ic) writeDesignRecord(design string) {
	body, err := os.ReadFile(design)
	if err != nil {
		c.stopAs("design-unreadable", "design-record", "cannot read "+design+": "+err.Error())
	}
	key := ""
	if c.jiraKey != "" {
		key = " (" + c.jiraKey + ")"
	}
	head := "# " + c.name + "\n> Design record of change " + c.name + key +
		", written at landing. Each decision's own **Status:** line is the status.\n\n"
	out := c.canon + "/docs/designs/" + c.name + ".md"
	if err := os.MkdirAll(c.canon+"/docs/designs", 0o755); err != nil {
		c.stopAs("record-not-writable", "design-record", "cannot create the directory of "+out+": "+err.Error())
	}
	if err := os.WriteFile(out, append([]byte(head), body...), 0o644); err != nil {
		c.stopAs("record-not-writable", "design-record", "cannot write "+out+": "+err.Error())
	}
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
