package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	icFixLine   = regexp.MustCompile(`^(- )?In-run pipeline fix:`)
	icFixFormat = regexp.MustCompile("In-run pipeline fix: `?([0-9a-f]{7,40}|deferred)`? — .+ \\(blast radius [0-9]+ files?\\)\\.?[[:space:]]*$")
)

func (c *ic) commit() {
	c.loadSet()
	if !c.findCanonical() {
		c.die("no worktree in the set holds %s's tasks.md", c.name)
	}
	live := c.canon + "/" + c.spec + "/changes/" + c.name
	arch := c.canon + "/" + c.spec + "/changes/archive/" + c.name
	var named []string // the sha group of every in-format in-run fix line
	for _, nf := range []string{live + "/narrative.md", arch + "/narrative.md"} {
		b, err := os.ReadFile(nf)
		if err != nil || !isFile(nf) {
			continue
		}
		var bad []string
		for n, l := range icLines(string(b)) {
			if !icFixLine.MatchString(l) {
				continue
			}
			if m := icFixFormat.FindStringSubmatch(l); m != nil {
				named = append(named, m[1])
			} else {
				bad = append(bad, fmt.Sprintf("%d:%s", n+1, l))
			}
		}
		if len(bad) > 0 {
			c.stop("narrative", nf+" holds in-run fix line(s) off the format 'In-run pipeline fix: <agents sha | deferred> — <the defect> (blast radius <N> files)': "+strings.Join(bad, "; ")+" — rewrite each to it, then re-run commit")
		}
	}
	c.requireFixLines(named)
	nf := live + "/narrative.md"
	if !isFile(nf) && isFile(arch+"/narrative.md") {
		nf = arch + "/narrative.md"
	}
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

	if !isDir(arch) || isDir(live) {
		c.markBegin("flow.sync-archive")
		fixes, _ := filepath.Glob(live + "-fix-[0-9]*")
		for _, d := range append([]string{live}, fixes...) {
			if !isDir(d) {
				continue
			}
			c.spectreArchive(c.canon, filepath.Base(d), "spectre archive "+filepath.Base(d)+" refused — report it; never --force")
		}
		c.markEnd("flow.sync-archive", "completed")
	}
	c.markBegin("flow.commit-archive")
	c.run("commit-archive", c.canon, c.name)
	if c.rc != 0 {
		c.stop("commit-archive", fmt.Sprintf("commit-archive exit %d — report its lines", c.rc))
	}
	// Every other worktree archives its own change directory -- a
	// satellite's link.md -- while it is still there, on an archived re-run
	// too, and commits it on its own branch (KAN-815).
	for _, o := range c.order {
		w := c.wt[o]
		if w == c.canon || !isDir(w+"/"+specRootLeaf(w, io.Discard)+"/changes/"+c.name) {
			continue
		}
		c.spectreArchive(w, c.name, "spectre archive "+c.name+" refused in "+w+" — report it; never --force")
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

// requireFixLines stops `narrative` on each `flow-fix` row the store lists
// as `fixed` whose ref no in-run fix line's sha (named) prefixes and the
// self-review report does not carry (kan-1032 self-review: a landed in-run
// fix was narrated without its line). An unreadable store skips the check.
func (c *ic) requireFixLines(named []string) {
	read := c.cmd("/", "flow", "self-review", "findings", "-C", c.main, "-change", c.name)
	read.Stderr = nil
	out, rc := c.capture(read)
	var rows []struct{ Angle, Disposition, Ref string }
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
		if !found {
			missing = append(missing, r.Ref[:min(7, len(r.Ref))])
		}
	}
	if len(missing) > 0 {
		c.stop("narrative", "no in-run fix line names the fixed flow-fix row(s) "+strings.Join(missing, ", ")+" — append 'In-run pipeline fix: <agents sha> — <the defect> (blast radius <N> files)' for each, then re-run commit")
	}
}

// spectreArchive is `(cd <dir> && spectre archive <change>)`, stopping
// `archive` with text when it is refused.
func (c *ic) spectreArchive(dir, change, text string) {
	cmd := c.cmd(dir, "spectre", "archive", change)
	cmd.Stdout = c.stdout
	if cmd.Run() != nil {
		c.stop("archive", text)
	}
}

// icReadTrimmed is `"$(cat <f>)"`.
func icReadTrimmed(f string) string {
	b, _ := os.ReadFile(f)
	return strings.TrimRight(string(b), "\n")
}
