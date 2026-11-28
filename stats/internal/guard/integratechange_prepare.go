package guard

import (
	"fmt"
	"os"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// switchBack: a clean checkout on another branch goes back to its default
// branch (operator instruction, kan-922's integrate run) — never a detached
// HEAD, whose commits sit on no branch, and never a branch the change
// records as its base (KAN-911), which run 2 refreshes. The branch left
// behind keeps every commit; a refused switch changes nothing and leaves
// the drift for STOP: main-checkout.
func (c *ic) switchBack(m string) {
	b, ok := c.gitQuiet("-C", m, "branch", "--show-current")
	if !ok || b == "" {
		return
	}
	d, ok := c.gitQuiet("-C", m, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")
	if !ok {
		return
	}
	d = strings.TrimPrefix(d, "origin/")
	if c.onRecordedBase(b) {
		return
	}
	sha, _ := c.git("-C", m, "rev-parse", "--short", "HEAD")
	// git overwrites an ignored file the target tracks without asking.
	deleted, _ := c.git("-C", m, "diff", "--name-only", "--no-renames", "--diff-filter=D", d, "HEAD")
	for _, p := range icLines(deleted) {
		if _, err := os.Lstat(m + "/" + p); err == nil {
			c.say("DRIFT-SWITCH-REFUSED: " + m + " — " + p + " exists on disk and " + d + " tracks it; left on " + b)
			return
		}
	}
	if c.gitPass("-C", m, "switch", "-q", d) {
		n, _ := c.git("-C", m, "rev-list", "--count", d+".."+b)
		c.say(fmt.Sprintf("DRIFT-SWITCHED: %s — %s -> %s, %s kept at %s (%s commits ahead of %s)", m, b, d, b, sha, n, d))
	} else {
		c.say("DRIFT-SWITCH-REFUSED: " + m + " — git switch " + d + " failed; left on " + b)
	}
}

// onRecordedBase: branch b is the base the change records
// (branch.spectre/<name>.flowBase) in any of its main checkouts.
func (c *ic) onRecordedBase(b string) bool {
	for _, k := range c.mains {
		if base, ok := c.gitQuiet("-C", k, "config", "--get", "branch.spectre/"+c.name+".flowBase"); ok && base == b {
			return true
		}
	}
	return false
}

func (c *ic) prepare() error {
	if err := c.loadSet(); err != nil {
		return err
	}
	if len(c.wt) == 0 {
		return c.stopAs("worktree-set-empty", "preflight", "the resolved worktree set is empty — report it and ask the operator")
	}

	// A main checkout holds no staged or dirty work — everything is done in
	// worktrees — so anything but a clean checkout on its default branch, or on
	// the base the change records (KAN-911: run 2 refreshes it), stops.
	dirty := false
	for _, m := range c.mains {
		c.runIn(checkMainCheckoutDrift, m)
		if c.rc == 0 && c.verdict("DRIFT-BRANCH") != "" && c.verdict("DRIFT-DIRTY") == "" {
			c.switchBack(m)
			c.runIn(checkMainCheckoutDrift, m)
		}
		if !(c.rc == 0 && c.verdict("DRIFT-CLEAN") != "") {
			b, _ := c.gitQuiet("-C", m, "branch", "--show-current")
			onBase := c.rc == 0 && c.verdict("DRIFT-DIRTY") == "" && b != "" && c.onRecordedBase(b)
			dirty = dirty || !onBase
		}
	}
	if dirty {
		return c.stopAs("main-checkout-dirty", "main-checkout", "a main checkout holds staged or dirty work, or is off its default branch — report the lines above; the operator cleans it, then re-run prepare")
	}

	run1 := false
	bm := make([]string, len(c.wt)) // each RUN1 worktree's base-moved line, the preflight's second line
	for i := range c.wt {
		if !core.IsDir(c.wt[i]) {
			return c.stopAs("worktree-missing", "preflight", c.wt[i]+" is not on disk — report it and ask the operator")
		}
		if err := c.baseOf(i); err != nil {
			return err
		}
		c.bases[i] = c.base
		c.run("check-finish-preflight", c.wt[i], "origin/"+c.base, c.mb[i])
		switch v := c.verdict("RUN1", "RUN2", "REFUSE"); {
		case strings.HasPrefix(v, "RUN1:"):
			run1 = true
			bm[i] = c.verdict("CLEAR", "MOVED", "REFUSE")
		case strings.HasPrefix(v, "RUN2:"):
		default:
			return c.stopAs("verdict-needs-operator", "preflight", "relay the verdict and its hand-verification procedure, then ask the operator")
		}
	}
	if !run1 {
		c.say("NEXT: run 2 — flow-guard integrate-change cleanup " + c.main + " " + c.name + " " + c.flags())
		return nil
	}

	// A worktree off the base another worktree records would land the change
	// on two different branches. One on its default branch whose origin lacks
	// that base gets it: created from the default branch at the same commit, so
	// the recorded merge base stands and the base-moved sync below moves the
	// branch. Any other mismatch stops.
	flowBase := "branch.spectre/" + c.name + ".flowBase"
	for fi := range c.wt {
		fb, ok := c.git("-C", c.wt[fi], "config", "--get", flowBase)
		if !ok {
			continue
		}
		for fj := range c.wt {
			if c.bases[fj] == fb {
				continue
			}
			wj := c.wt[fj]
			c.gitQuiet("-C", wj, "fetch", "-q", "origin")
			_, hasRemote := c.git("-C", wj, "rev-parse", "-q", "--verify", "refs/remotes/origin/"+fb)
			_, hasBase := c.git("-C", wj, "config", "--get", flowBase)
			if !hasRemote && !hasBase && c.mb[fj] != "-" {
				created := c.gitPass("-C", wj, "push", "-q", "origin", "refs/remotes/origin/"+c.bases[fj]+":refs/heads/"+fb) &&
					c.gitPass("-C", wj, "fetch", "-q", "origin", fb) &&
					c.gitPass("-C", wj, "config", flowBase, fb)
				if created {
					add := c.cmd("/", "flow", "state", "add-worktree", "-C", c.main, c.name, wj, c.mb[fj])
					add.Stdout = c.stderr
					created = add.Run() == nil
				}
				if !created {
					return c.stop("base-mismatch", "creating origin/"+fb+" from "+c.bases[fj]+" and recording it as "+wj+"'s base failed — ask the operator")
				}
				c.say("BASE-CREATED: origin/" + fb + " from " + c.bases[fj] + " in " + wj + "'s repository; " + wj + "'s base recorded as it")
				c.bases[fj] = fb
				continue
			}
			return c.stop("base-mismatch", wj+" resolves base "+c.bases[fj]+", but "+c.wt[fi]+" records "+fb+" — ask the operator; never land it on "+c.bases[fj])
		}
	}

	c.markBegin("flow.land")
	if !c.findCanonical() {
		return c.stopAs("tasks-md-not-found", "preflight", "no worktree in the set holds "+c.name+"'s tasks.md — ask the operator")
	}
	// This prepare's own flow.land run supersedes a stopped earlier one; a stale
	// marker would make the next commit open a second run (c.canon is known only now).
	_ = os.Remove(c.landStoppedFile())

	ask := false
	// Every worktree's commit range, from the merge base, for signal seven's
	// newest code commit.
	var ranges []string
	for i := range c.wt {
		rb, _ := c.git("-C", c.wt[i], "merge-base", "HEAD", "origin/"+c.bases[i])
		ranges = append(ranges, c.wt[i]+"="+rb)
	}
	// A guard that reached no verdict is not OUTSTANDING: its own stop.
	for i := range c.wt {
		c.run("check-unfinished-work", append([]string{c.wt[i], c.name, c.canon}, ranges...)...)
		switch v := fmt.Sprintf("%d:%s", c.rc, c.verdict("CLEAR", "OUTSTANDING")); {
		case strings.HasPrefix(v, "0:CLEAR:"):
		case strings.HasPrefix(v, "0:OUTSTANDING:"):
			ask = true
		default:
			return c.stop("no-verdict", fmt.Sprintf("check-unfinished-work reached no verdict for %s (exit %d) — report its output and ask the operator", c.wt[i], c.rc))
		}
	}
	if ask {
		return c.stop("unfinished-work", "load skills/flow/unfinished-work-gate.md: finish the unfinished work, asking nothing, then re-run prepare")
	}
	var moved []int
	for i := range c.wt {
		v := bm[i]
		switch {
		case v == "": // a RUN2 worktree, already landed: nothing to sync
			continue
		case strings.HasPrefix(v, "CLEAR:") && strings.Contains(v[len("CLEAR:"):], "merge base now "):
			c.rb[i] = v[strings.LastIndex(v, "merge base now ")+len("merge base now "):]
		case strings.HasPrefix(v, "CLEAR:"):
			c.rb[i] = c.mb[i]
		case strings.HasPrefix(v, "MOVED:"):
			moved = append(moved, i)
		default:
			return c.stop("base-moved", "report the verdict and ask the operator")
		}
	}
	for _, i := range moved {
		c.run("sync-onto-base", c.wt[i], c.bases[i], c.mb[i])
		switch c.rc {
		case 0:
		case 1:
			return c.stop("sync-conflict", "load skills/flow/sync-onto-base.md and resolve in place, run 'flow-guard sync-onto-base --resume', the whole ## lint and ## test lists, then re-run prepare")
		default:
			return c.stopAs("sync-failed", "sync", fmt.Sprintf("sync-onto-base exit %d — report its output and ask the operator", c.rc))
		}
		vline := c.verdict("REBASED")
		if k := strings.LastIndex(vline, "onto "); k >= 0 {
			vline = vline[k+len("onto "):]
		}
		c.rb[i] = vline
		if err := c.rebaseLint(c.wt[i], ""); err != nil {
			return err
		}
	}

	if route := c.defaultRoute(); route != "" {
		c.say("ROUTE: " + route + " — from this project's configured default, not asked")
	} else {
		c.say("ROUTE: ask — no configured default; ask the landing question, then pass --route to land")
	}
	if c.prURLState != "" {
		c.say("PR: " + c.prURLState)
	}

	subjects := ""
	for i := range c.wt {
		f := c.reshapeFile(c.wt[i])
		os.Remove(f)
		rb := c.reshapeBase(c.wt[i], c.rb[i])
		if err := os.WriteFile(f, []byte(rb+"\n"), 0o644); err != nil {
			fmt.Fprintf(c.stderr, "integrate-change: %v\n", err)
		}
		c.say("WORKTREE: " + c.wt[i] + " — base " + c.bases[i] + ", reshape base " + rb)
		sp := core.SpecRootLeaf(c.wt[i], c.stderr)
		c.gitPass("-C", c.wt[i], "diff", "--no-renames", "--stat", rb, "--", ".", ":(exclude)"+sp+"/changes/")
		others, _ := c.git("-C", c.wt[i], "ls-files", "--others", "--exclude-standard", "--", ".", ":(exclude)"+sp+"/changes/")
		for _, l := range icLines(others) {
			c.say(" new: " + l)
		}
		if len(c.implPaths(c.wt[i], rb)) > 0 && !c.linkOnly(i, rb) {
			subjects += " " + c.wt[i] + "=<impl subject>"
		}
	}
	c.say("NEXT: flow-guard integrate-change commit " + c.main + " " + c.name + subjects + " " + c.flags())
	return nil
}

// defaultRoute is project-get.sh's `default landing route`, or nothing.
func (c *ic) defaultRoute() string {
	r, rc := c.capture(c.sibling("project-get", c.main, "default landing route", "--enum", "pull request", "merge and push", "manual"))
	if rc != 0 {
		return ""
	}
	return r
}
