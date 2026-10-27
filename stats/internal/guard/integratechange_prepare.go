package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// switchBack: a clean checkout on another branch goes back to its default
// branch (operator instruction, kan-922's integrate run) — never a detached
// HEAD, whose commits sit on no branch, and never a branch the change
// records as its base (KAN-911), which run 2 refreshes. The branch left
// behind keeps every commit; a refused switch changes nothing and leaves
// the drift for the question.
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
	for _, k := range c.mains {
		if base, ok := c.gitQuiet("-C", k, "config", "--get", "branch.spectre/"+c.name+".flowBase"); ok && base == b {
			return
		}
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

func (c *ic) prepare() {
	c.loadSet()
	if len(c.wt) == 0 {
		c.stop("preflight", "the resolved worktree set is empty — report it and ask the operator")
	}

	ask := false
	for _, m := range c.mains {
		c.run("check-foreign-staged", m)
		fs := c.rc == 0 && c.verdict("STAGED-CLEAN") != ""
		c.run("check-main-checkout-drift", m)
		if fs && c.rc == 0 && c.verdict("DRIFT-BRANCH") != "" && c.verdict("DRIFT-DIRTY") == "" {
			c.switchBack(m)
			c.run("check-main-checkout-drift", m)
		}
		if !fs || !(c.rc == 0 && c.verdict("DRIFT-CLEAN") != "") {
			ask = true
		}
	}
	if ask && !c.acceptForeign {
		c.stop("foreign-staged", "ask the operator the foreign-staged-work question; Continue re-runs prepare with --accept-foreign")
	}

	c.run("migrate-worktrees", c.mains...)
	if c.rc != 0 {
		c.stop("migrate", fmt.Sprintf("migrate-worktrees exit %d — report its lines", c.rc))
	}
	c.loadSet()

	run1 := false
	for i := range c.wt {
		if !isDir(c.wt[i]) {
			c.stop("preflight", c.wt[i]+" is not on disk — report it and ask the operator")
		}
		c.baseOf(i)
		c.bases[i] = c.base
		c.run("check-finish-preflight", c.wt[i], "origin/"+c.base, c.mb[i])
		switch v := c.verdict("RUN1", "RUN2", "REFUSE"); {
		case strings.HasPrefix(v, "RUN1:"):
			run1 = true
		case strings.HasPrefix(v, "RUN2:"):
		default:
			c.stop("preflight", "relay the verdict and its hand-verification procedure, then ask the operator")
		}
	}
	if !run1 {
		c.say("NEXT: run 2 — integrate-change.sh cleanup " + c.main + " " + c.name + " " + c.flags())
		panic(icExit(0))
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
					c.stop("base-mismatch", "creating origin/"+fb+" from "+c.bases[fj]+" and recording it as "+wj+"'s base failed — ask the operator")
				}
				c.say("BASE-CREATED: origin/" + fb + " from " + c.bases[fj] + " in " + wj + "'s repository; " + wj + "'s base recorded as it")
				c.bases[fj] = fb
				continue
			}
			c.stop("base-mismatch", wj+" resolves base "+c.bases[fj]+", but "+c.wt[fi]+" records "+fb+" — ask the operator; never land it on "+c.bases[fj])
		}
	}

	c.markBegin("flow.preflight")
	c.markEnd("flow.preflight", "completed")
	c.markBegin("flow.unfinished-work-gate")
	if !c.findCanonical() {
		c.stop("preflight", "no worktree in the set holds "+c.name+"'s tasks.md — ask the operator")
	}
	ci := c.idxOf(c.canon)
	live := c.canon + "/" + c.spec + "/changes/" + c.name
	arch := c.canon + "/" + c.spec + "/changes/archive/" + c.name
	if isDir(arch) {
		subjects, _ := c.gitQuiet("-C", c.canon, "log", "--format=%s", c.mb[ci]+"..HEAD")
		if !icHas(icLines(subjects), "chore(spectre): archive "+c.name) {
			if isDir(live) {
				c.stop("archive-undo", "both "+live+" and "+arch+" exist — ask the operator to check the live copy holds nothing the archived one lacks, remove it, and re-run")
			}
			fixes, _ := filepath.Glob(arch + "-fix-[0-9]*")
			for _, d := range append([]string{arch}, fixes...) {
				if !isDir(d) {
					continue
				}
				if !c.gitPass("-C", c.canon, "mv", strings.TrimPrefix(d, c.canon+"/"), c.spec+"/changes/"+filepath.Base(d)) {
					c.stop("archive-undo", "git mv of "+d+" failed")
				}
				c.say("UNARCHIVED: " + d)
			}
		}
	}
	archived := isDir(arch) && !isDir(live)
	newwork := false
	if archived {
		head, _ := c.git("-C", c.canon, "rev-parse", "HEAD")
		ab, _ := c.git("-C", c.canon, "rev-parse", c.archBase(c.canon, c.mb[ci]))
		newwork = head != ab
		for i := range c.wt {
			if st, _ := c.git("-C", c.wt[i], "status", "--porcelain", "--untracked-files=no"); st != "" {
				newwork = true
			}
		}
		if newwork {
			c.say("ARCHIVED: " + c.name + " — with new work")
		} else {
			c.say("ARCHIVED: " + c.name + " — without new work")
		}
	}

	if archived {
		c.liveCheckGate(arch)
	} else {
		c.liveCheckGate(live)
	}

	if archived && !newwork {
		// No new work still runs the visual gate: the archive is no verifier.
		c.visualGate()
		c.markEnd("flow.unfinished-work-gate", "completed")
	} else {
		ask = false
		// Every worktree's commit range, for signal four's Task-Id search. An
		// archived re-run starts at the archive: the guard exempts the tasks
		// already ticked there, whose commits run 1 squashed.
		rangeBase := func(i int) string {
			rb, _ := c.git("-C", c.wt[i], "merge-base", "HEAD", "origin/"+c.bases[i])
			if archived {
				rb = c.archBase(c.wt[i], rb)
			}
			return rb
		}
		var ranges []string
		for i := range c.wt {
			ranges = append(ranges, c.wt[i]+"="+rangeBase(i))
		}
		// A guard that reached no verdict is not OUTSTANDING: its own stop,
		// which --accept-outstanding never passes.
		for i := range c.wt {
			c.run("check-unfinished-work", append([]string{c.wt[i], c.name, c.canon}, ranges...)...)
			switch v := fmt.Sprintf("%d:%s", c.rc, c.verdict("CLEAR", "OUTSTANDING")); {
			case strings.HasPrefix(v, "0:CLEAR:"):
			case strings.HasPrefix(v, "0:OUTSTANDING:"):
				ask = true
			default:
				c.stop("no-verdict", fmt.Sprintf("check-unfinished-work reached no verdict for %s (exit %d) — report its output and ask the operator; --accept-outstanding does not pass this", c.wt[i], c.rc))
			}
			c.run("check-untasked-commits", c.wt[i], rangeBase(i))
			switch v := fmt.Sprintf("%d:%s", c.rc, c.verdict("UNTASKED-OK", "UNTASKED")); {
			case strings.HasPrefix(v, "0:UNTASKED-OK:"):
			case strings.HasPrefix(v, "1:UNTASKED:"):
				ask = true
			default:
				c.stop("no-verdict", fmt.Sprintf("check-untasked-commits reached no verdict for %s (exit %d) — report its output and ask the operator; --accept-outstanding does not pass this", c.wt[i], c.rc))
			}
		}
		// A UI no verifier saw at this HEAD is a hard stop (kan-939), never an
		// OUTSTANDING item: --accept-outstanding does not pass it, only an
		// operator's false-positive flag on that worktree's row for that HEAD.
		c.visualGate()
		if ask && !c.acceptOutstanding {
			c.stop("unfinished-work", "load skills/flow/unfinished-work-gate.md: finish the unfinished work, asking nothing, then re-run prepare")
		}
		c.markEnd("flow.unfinished-work-gate", "completed")
	}

	c.markBegin("flow.landing-question")
	var moved []int
	for i := range c.wt {
		c.run("check-base-moved", c.wt[i], "origin/"+c.bases[i], c.mb[i])
		v := c.verdict("CLEAR", "MOVED", "REFUSE")
		switch {
		case strings.HasPrefix(v, "CLEAR:") && strings.Contains(v[len("CLEAR:"):], "merge base now "):
			c.rb[i] = v[strings.LastIndex(v, "merge base now ")+len("merge base now "):]
		case strings.HasPrefix(v, "CLEAR:"):
			c.rb[i] = c.mb[i]
		case strings.HasPrefix(v, "MOVED:"):
			moved = append(moved, i)
		default:
			c.stop("base-moved", "report the verdict and ask the operator")
		}
	}
	for _, i := range moved {
		c.run("sync-onto-base", c.wt[i], c.bases[i], c.mb[i])
		switch c.rc {
		case 0:
		case 1:
			c.stop("sync-conflict", "load skills/flow/sync-onto-base.md and resolve in place, run '"+c.sibling("sync-onto-base")+"' --resume, the whole ## lint and ## test lists, then re-run prepare")
		default:
			c.stop("sync", fmt.Sprintf("sync-onto-base exit %d — report its output and ask the operator", c.rc))
		}
		vline := c.verdict("REBASED")
		if k := strings.LastIndex(vline, "onto "); k >= 0 {
			vline = vline[k+len("onto "):]
		}
		c.rb[i] = vline
		c.guardTests()
		c.rebaseLint(c.wt[i], "")
		// The rebase made a HEAD no verifier saw: the gate again.
		if !c.visualOK(c.wt[i], c.bases[i]) {
			c.stop("visual-verify", "rebased onto origin/"+c.bases[i]+" after the gate: "+visualStopText(c.wt[i], "prepare"))
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
	c.markEnd("flow.landing-question", "completed")

	reshape := false
	subjects := ""
	for i := range c.wt {
		f := c.reshapeFile(c.wt[i])
		os.Remove(f)
		if archived && !newwork {
			continue
		}
		rb := c.rb[i]
		if archived {
			rb = c.archBase(c.wt[i], rb)
		}
		if err := os.WriteFile(f, []byte(rb+"\n"), 0o644); err != nil {
			fmt.Fprintf(c.stderr, "integrate-change: %v\n", err)
		}
		reshape = true
		c.say("WORKTREE: " + c.wt[i] + " — base " + c.bases[i] + ", reshape base " + rb)
		sp := specRootLeaf(c.wt[i], c.stderr)
		c.gitPass("-C", c.wt[i], "diff", "--no-renames", "--stat", rb, "--", ".", ":(exclude)"+sp+"/changes/")
		others, _ := c.git("-C", c.wt[i], "ls-files", "--others", "--exclude-standard", "--", ".", ":(exclude)"+sp+"/changes/")
		for _, l := range icLines(others) {
			c.say(" new: " + l)
		}
		if len(c.implPaths(c.wt[i], rb)) > 0 && !c.linkOnly(i, rb) {
			subjects += " " + c.wt[i] + "=<impl subject>"
		}
	}
	narrative := live + "/narrative.md"
	if archived {
		narrative = arch + "/narrative.md"
	}
	if reshape {
		c.markBegin("flow.preserve-sessions")
		c.say("NEXT: append this run's narrative to " + narrative + ", then integrate-change.sh commit " + c.main + " " + c.name + subjects + " " + c.flags())
	} else {
		c.say("NEXT: integrate-change.sh commit " + c.main + " " + c.name + " " + c.flags())
	}
}

// defaultRoute is project-get.sh's `default landing route`, or nothing.
func (c *ic) defaultRoute() string {
	r, rc := c.capture(c.cmd("/", c.sibling("project-get"), c.main, "default landing route", "--enum", "pull request", "merge and push", "manual"))
	if rc != 0 {
		return ""
	}
	return r
}

// icLiveCheckNamed: design's `## Live check` section names something to
// exercise — its first non-blank line opens with none of None, No ,
// Not applicable. False when the file cannot be read or has no section.
func icLiveCheckNamed(design string) bool {
	b, err := os.ReadFile(design)
	if err != nil {
		return false
	}
	in := false
	for _, l := range strings.Split(string(b), "\n") {
		l = strings.TrimRight(l, " \t\r")
		if !in {
			in = l == "## Live check"
			continue
		}
		if strings.HasPrefix(l, "# ") || strings.HasPrefix(l, "## ") {
			return false
		}
		if l == "" {
			continue
		}
		return !strings.HasPrefix(l, "None") && !strings.HasPrefix(l, "No ") && !strings.HasPrefix(l, "Not applicable")
	}
	return false
}

// liveCheckGate: a design naming a live check needs live-verification.md
// beside it passing check-live-verification, else STOP: live-check, which
// --accept-outstanding never passes.
func (c *ic) liveCheckGate(dir string) {
	if !icLiveCheckNamed(dir + "/design.md") {
		return
	}
	rec := dir + "/live-verification.md"
	if !isFile(rec) {
		c.stop("live-check", dir+"/design.md names a live check and "+rec+" is missing — run Live check (skills/flow/verify-and-handoff.md), through the stand-in its section states wherever a step would write this change's integrate state, then re-run prepare; --accept-outstanding does not pass this")
	}
	c.run("check-live-verification", rec)
	if c.rc == 2 {
		c.stop("live-check", fmt.Sprintf("check-live-verification could not answer on %s (exit 2) — correct the input it names, then re-run prepare; --accept-outstanding does not pass this", rec))
	}
	if c.rc != 0 {
		c.stop("live-check", fmt.Sprintf("check-live-verification exit %d on %s — measure what it names, rewrite the record, then re-run prepare; --accept-outstanding does not pass this", c.rc, rec))
	}
}
