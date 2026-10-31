package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	icFixLine   = regexp.MustCompile(`^(- )?In-run pipeline fix:`)
	icFixFormat = regexp.MustCompile("In-run pipeline fix: `?([0-9a-f]{7,40}|deferred)`? — .+ \\(blast radius ([0-9]+) files?\\)\\.?[[:space:]]*$")
)

func (c *ic) commit() {
	c.loadSet()
	if !c.findCanonical() {
		c.die("no worktree in the set holds %s's tasks.md", c.name)
	}
	live := c.canon + "/" + c.spec + "/changes/" + c.name
	// A dropped re-run without new work has no change directory left: its
	// narrative was checked before the drop commit that removed it.
	if isDir(live) {
		c.checkNarrative(live + "/narrative.md")
	}
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
			c.pinPreReshape(c.wt[o])
			c.run("reshape-branch", c.wt[o], c.name, rb)
			if c.rc != 0 {
				c.stop("reshape", fmt.Sprintf("reshape-branch exit %d in %s — report its lines", c.rc, c.wt[o]))
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
				c.stop("commit-split", "no implementation subject given for "+c.wt[o]+" — pass "+c.wt[o]+"=<subject>")
			}
			c.run("commit-split", c.wt[o], c.name, subj, c.planMsg)
			if c.rc != 0 {
				c.stop("commit-split", fmt.Sprintf("commit-split exit %d in %s — report its lines", c.rc, c.wt[o]))
			}
			os.Remove(f)
		}
		c.markEnd("flow.commit-two", "completed")
	}

	// flow.sync-archive keeps its key (decision keep-guard-and-stage-names):
	// nothing is moved any more, so it marks the live directory's presence.
	if isDir(live) {
		c.markBegin("flow.sync-archive")
		c.markEnd("flow.sync-archive", "completed")
	}
	c.markBegin("flow.commit-archive")
	c.run("commit-archive", c.canon, c.name)
	if c.rc != 0 {
		c.stop("commit-archive", fmt.Sprintf("commit-archive exit %d — report its lines", c.rc))
	}
	// Every other worktree drops its own change directory -- a satellite's
	// link.md -- while it is still there, on a dropped re-run too, and
	// commits it on its own branch (KAN-815).
	for _, o := range c.order {
		w := c.wt[o]
		if w == c.canon || !isDir(w+"/"+specRootLeaf(w, io.Discard)+"/changes/"+c.name) {
			continue
		}
		c.run("commit-archive", w, c.name)
		if c.rc != 0 {
			c.stop("commit-archive", fmt.Sprintf("commit-archive exit %d in %s — report its lines", c.rc, w))
		}
	}
	// The rewritten branch reaches the remote now, not at land: a run that
	// stops before land leaves the next plain branch-backup push a fast-forward.
	for _, o := range c.order {
		c.baseOf(o)
		if !c.visualOK(c.wt[o], c.base) {
			c.stop("visual-verify", "spectre/"+c.name+" not pushed from "+c.wt[o]+". "+visualStopText(c.wt[o], "commit"))
		}
		c.pushBranch(c.wt[o])
		c.pushPin(c.wt[o])
	}
	c.markEnd("flow.commit-archive", "completed")
	c.markBegin("flow.self-review")
	if _, ok := c.gitQuiet("-C", c.canon, "cat-file", "-e", "HEAD:docs/self-review/"+c.name+"-self-review.md"); ok {
		c.say("NEXT: the self-review report is already committed — skip the pass; integrate-change.sh land " + c.main + " " + c.name + " " + c.flags())
	} else {
		c.say("NEXT: run the self-review pass over `flow self-review bundle -C " + c.canon + " -change " + c.name + "` and land its report in " + c.canon + ", then integrate-change.sh land " + c.main + " " + c.name + " " + c.flags())
	}
}

// checkNarrative stops `narrative` when the change's narrative nf holds an
// in-run fix line off the format, defers a small fix, lacks a line a fixed
// flow-fix row needs (requireFixLines) or a section a run the store marked
// needs (check-narrative-runs).
func (c *ic) checkNarrative(nf string) {
	var named []string // the sha group of every in-format in-run fix line
	if b, err := os.ReadFile(nf); err == nil && isFile(nf) {
		var bad []string
		for n, l := range icLines(string(b)) {
			if !icFixLine.MatchString(l) {
				continue
			}
			if m := icFixFormat.FindStringSubmatch(l); m != nil {
				// Only a `big` fix -- 60 files or more -- is ever deferred
				// (Pipeline defects found mid-run, pipeline.md).
				if r, _ := strconv.Atoi(m[2]); m[1] == "deferred" && r < 60 {
					c.stop("narrative", fmt.Sprintf("%s:%d defers an in-run fix of blast radius %d — only a fix of 60 files or more is deferred; fix and land it in this run, then re-run commit", nf, n+1, r))
				}
				named = append(named, m[1])
			} else {
				bad = append(bad, fmt.Sprintf("%d:%s", n+1, l))
			}
		}
		if len(bad) > 0 {
			c.stop("narrative", nf+" holds in-run fix line(s) off the format 'In-run pipeline fix: <agents sha | deferred> — <the defect> (blast radius <N> files)': "+strings.Join(bad, "; ")+" — rewrite each to it, then re-run commit")
		}
	}
	c.requireFixLines(named, nf)
	// stderr too: its exit-2 message is the stop's cause.
	runs := c.cmd("/", c.sibling("check-narrative-runs"), c.canon, c.name, nf)
	var b bytes.Buffer
	runs.Stdout, runs.Stderr = &b, &b
	c.rc = rcOf(runs.Run())
	c.out = strings.TrimRight(b.String(), "\n")
	if c.rc != 0 {
		lines := strings.Split(c.out, "\n")
		c.stop("narrative", lines[len(lines)-1]+" — append each missing run's section (skills/flow/verify-and-handoff.md, Write IN_PROGRESS), then re-run commit")
	}
}

// requireFixLines appends to the narrative nf an in-run fix line for each
// `flow-fix` row the store lists as `fixed` whose ref no in-run fix line's
// sha (named) prefixes and the self-review report does not carry, derived
// from the row's ref, note and blast radius (kan-1039: a fix the planning
// session landed had no narrative to be recorded in). Such a row lacking a
// one-line note or a blast radius, or with no narrative file, stops
// `narrative` (kan-1032 self-review: a landed in-run fix was narrated
// without its line). An unreadable store skips the check.
func (c *ic) requireFixLines(named []string, nf string) {
	read := c.cmd("/", "flow", "self-review", "findings", "-C", c.main, "-change", c.name)
	read.Stderr = nil
	out, rc := c.capture(read)
	var rows []struct {
		Angle, Disposition, Ref, Note string
		BlastRadius                   *int `json:"blastRadius"`
	}
	if rc != 0 || json.Unmarshal([]byte(out), &rows) != nil {
		fmt.Fprintln(c.stderr, "integrate-change: warning: flow self-review findings unreadable — the in-run fix line check is skipped")
		return
	}
	report, _ := os.ReadFile(c.report())
	var missing []string
	for _, r := range rows {
		if r.Angle != "flow-fix" || r.Disposition != "fixed" || r.Ref == "" || bytes.Contains(report, []byte(r.Ref)) {
			continue
		}
		found := false
		for _, s := range named {
			if strings.HasPrefix(r.Ref, s) || strings.HasPrefix(s, r.Ref) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		short := r.Ref[:min(7, len(r.Ref))]
		if r.Note == "" || strings.ContainsAny(r.Note, "\r\n") || r.BlastRadius == nil || !isFile(nf) {
			missing = append(missing, short)
			continue
		}
		line := fmt.Sprintf("- In-run pipeline fix: %s — %s (blast radius %d files)", short, r.Note, *r.BlastRadius)
		b, err := os.ReadFile(nf)
		if err != nil {
			missing = append(missing, short)
			continue
		}
		add := line + "\n"
		if len(b) > 0 && b[len(b)-1] != '\n' {
			add = "\n" + add
		}
		f, err := os.OpenFile(nf, os.O_APPEND|os.O_WRONLY, 0)
		if err == nil {
			_, err = f.WriteString(add)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}
		if err != nil {
			missing = append(missing, short)
			continue
		}
		c.say("narrative: appended " + line)
	}
	if len(missing) > 0 {
		c.stop("narrative", "no in-run fix line names the fixed flow-fix row(s) "+strings.Join(missing, ", ")+", and the row carries no one-line note or blast radius to derive it from — append 'In-run pipeline fix: <agents sha> — <the defect> (blast radius <N> files)' for each, then re-run commit")
	}
}

// icReadTrimmed is `"$(cat <f>)"`.
func icReadTrimmed(f string) string {
	b, _ := os.ReadFile(f)
	return strings.TrimRight(string(b), "\n")
}
