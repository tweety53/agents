package guard

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	icFixedSha = regexp.MustCompile(`fixed: [0-9a-f]{7,40}`)
	icFixedRow = regexp.MustCompile(`^- \*\*\[([a-z-]+)\]\*\* (.*) — fixed: ([0-9a-f]+)`)
)

func (c *ic) report() string { return c.canon + "/docs/self-review/" + c.name + "-self-review.md" }

// rereadFixShas: before a landing push from worktree <i>, rewrite each
// report `fixed:` sha this worktree's repository holds and a rebase
// changed, found by its subject on this worktree's branch, and commit the
// rewrite in the canonical worktree, which carries the report (step 5 of
// /flow-self-review). A rewrite after the canonical push pushes the
// canonical branch again, and on merge and push its base, a fast-forward.
func (c *ic) rereadFixShas(i int) {
	w, rpt := c.wt[i], c.report()
	b, err := os.ReadFile(rpt)
	if err != nil || !isFile(rpt) {
		return
	}
	seen := map[string]bool{}
	var olds []string
	for _, m := range icFixedSha.FindAllString(string(b), -1) {
		if old := m[len("fixed: "):]; !seen[old] {
			seen[old] = true
			olds = append(olds, old)
		}
	}
	sort.Strings(olds)
	changed := false
	for _, old := range olds {
		if _, ok := c.gitQuiet("-C", w, "merge-base", "--is-ancestor", old, "HEAD"); ok {
			continue
		}
		subj, ok := c.gitQuiet("-C", w, "log", "-1", "--format=%s", old)
		if !ok {
			continue
		}
		log, _ := c.git("-C", w, "log", "--format=%h %s", "HEAD", "--not", "origin/"+c.base)
		nw := icFirstSubject(log, subj)
		if nw == "" {
			continue
		}
		cur, _ := os.ReadFile(rpt)
		if err := os.WriteFile(rpt, []byte(strings.ReplaceAll(string(cur), "fixed: "+old, "fixed: "+nw)), 0o644); err != nil {
			fmt.Fprintf(c.stderr, "integrate-change: %v\n", err)
		}
		c.say("FIXED-SHA: " + old + " -> " + nw)
		changed = true
	}
	if !changed {
		return
	}
	c.run("land-self-review-report", c.canon, "spectre/"+c.name, "docs(self-review): "+c.name+" self-review report", strings.TrimPrefix(rpt, c.canon+"/"))
	if c.rc != 0 {
		c.stop("self-review-report", fmt.Sprintf("land-self-review-report exit %d — report its line", c.rc))
	}
	if w == c.canon || c.canonBase == "" {
		return
	}
	if !c.visualOK(c.canon, c.canonBase) {
		c.stop("visual-verify", "the rewritten report was not pushed. "+visualStopText(c.canon, "land"))
	}
	c.pushBranch(c.canon)
	if c.route != "merge and push" {
		return
	}
	if !c.gitPass("-C", c.canon, "push", "origin", "HEAD:"+c.canonBase) {
		c.stop("push", "the re-push of "+c.canon+" to origin/"+c.canonBase+" was rejected — report git's output; the change stays IN_PROGRESS")
	}
}

// recordFixedRows: once the landing push succeeded, one `fixed` row per
// report line whose sha a repository of the worktree set carries and the
// store does not list.
func (c *ic) recordFixedRows() {
	b, err := os.ReadFile(c.report())
	if err != nil {
		return
	}
	known := c.cmd("/", "flow", "self-review", "findings", "-C", c.main, "-change", c.name)
	known.Stderr = nil
	knownOut, _ := c.capture(known)
	// `while IFS= read -r line`: nothing trimmed, an unterminated last line unread.
	lines := strings.Split(string(b), "\n")
	for _, line := range lines[:len(lines)-1] {
		m := icFixedRow.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		found := false
		for _, w := range c.wt {
			if _, ok := c.gitQuiet("-C", w, "cat-file", "-e", m[3]+"^{commit}"); ok {
				found = true
				break
			}
		}
		if !found || strings.Contains(knownOut, m[3]) {
			continue
		}
		rec := c.cmd("/", "flow", "self-review", "finding", "-C", c.main, "-change", c.name, "-angle", m[1],
			"-disposition", "fixed", "-ref", m[3], "-note", m[2])
		rec.Stdout = c.stdout
		if rec.Run() != nil {
			fmt.Fprintf(c.stderr, "integrate-change: warning: recording the fixed row for %s failed\n", m[3])
		}
	}
}

// prArg: the --pr-url answer given for worktree <i>, by its path or its
// main checkout.
func (c *ic) prArg(i int) (string, bool) {
	for _, s := range c.prArgs {
		k, v, _ := strings.Cut(s, "=")
		if k == c.wt[i] || k == c.repo[i] {
			return v, true
		}
	}
	return "", false
}

// ghPRCreate is `(cd <w> && gh pr create --fill …) 2>&1 | tail -1`: its
// last line, and whether the pipeline (pipefail) succeeded.
func (c *ic) ghPRCreate(w string) (string, bool) {
	cmd := c.cmd(w, "gh", "pr", "create", "--fill", "--base", c.base, "--head", "spectre/"+c.name)
	var b bytes.Buffer
	cmd.Stdout, cmd.Stderr = &b, &b
	err := cmd.Run()
	if rcOf(err) == 127 {
		b.WriteString("gh: command not found\n")
	}
	lines := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	return lines[len(lines)-1], err == nil
}

func (c *ic) land() {
	c.canonBase = ""
	c.loadSet()
	if !c.findCanonical() {
		c.die("no worktree in the set holds %s's tasks.md", c.name)
	}
	c.mergeOrder()
	if _, ok := c.gitQuiet("-C", c.canon, "cat-file", "-e", "HEAD:docs/self-review/"+c.name+"-self-review.md"); !ok {
		c.stop("self-review-report", "docs/self-review/"+c.name+"-self-review.md is not committed in "+c.canon+" — run the self-review pass and land its report, then re-run land; nothing was pushed")
	}
	if c.route == "" {
		c.route = c.defaultRoute()
	}
	switch c.route {
	case "pull request", "merge and push", "manual":
	default:
		c.stop("landing-route", "no route — ask the landing question, then integrate-change.sh land "+c.main+" "+c.name+" --route <route> "+c.flags())
	}
	c.markEnd("flow.self-review", "completed")
	c.markBegin("flow.landing-routes")
	pr, pushed := "", ""
	alreadyPushed := func() string {
		if pushed == "" {
			return "none"
		}
		return pushed
	}
	for _, o := range c.order {
		w := c.wt[o]
		c.baseOf(o)
		c.rereadFixShas(o)
		if !c.visualOK(w, c.base) {
			c.stop("visual-verify", "spectre/"+c.name+" not pushed from "+w+"; already pushed: "+alreadyPushed()+". "+visualStopText(w, "land"))
		}
		c.pushBranch(w)
		if w == c.canon {
			c.canonBase = c.base
		}
		switch c.route {
		case "merge and push":
			// An own-branch self-review loop lands only once its rows admit it (NA passes).
			c.run("check-loop-landing", c.canon, "spectre/"+c.name)
			if c.rc != 0 {
				c.stop("loop-landing", fmt.Sprintf("check-loop-landing.sh exit %d — finish the self-review loop its LOOP-LANDING-REFUSED lines name, then re-run land; nothing landed from %s", c.rc, w))
			}
			if !c.gitPass("-C", w, "push", "origin", "HEAD:"+c.base) {
				c.say("PUSH-REJECTED: " + w + " — re-syncing onto origin/" + c.base + " once")
				c.baseOf(o)
				fork, _ := c.git("-C", w, "merge-base", "HEAD", "origin/"+c.base)
				c.run("sync-onto-base", w, c.base, fork)
				switch c.rc {
				case 0:
				case 1:
					c.stop("sync-conflict", "load skills/flow/sync-onto-base.md, resolve in place, --resume, run the whole ## lint and ## test lists, then re-run land — that push is the one retry")
				default:
					c.stop("sync", fmt.Sprintf("sync-onto-base exit %d — report its output", c.rc))
				}
				c.guardTests()
				c.rebaseLint(w, "already pushed: "+alreadyPushed()+"; ")
				c.rereadFixShas(o)
				if !c.visualOK(w, c.base) {
					c.stop("visual-verify", "re-synced onto origin/"+c.base+" after its push was rejected, so it was not pushed there; already pushed: "+alreadyPushed()+". "+visualStopText(w, "land"))
				}
				c.pushBranch(w)
				if !c.gitPass("-C", w, "push", "origin", "HEAD:"+c.base) {
					c.stop("push", "the second push of "+w+" to origin/"+c.base+" was rejected — report git's output; the change stays IN_PROGRESS")
				}
			}
			c.say("PUSHED: " + w + " — spectre/" + c.name + " onto origin/" + c.base)
			if pushed != "" {
				pushed += ", "
			}
			pushed += w
		case "pull request":
			if c.prURLState != "" {
				pr = c.prURLState
				c.say("PR: " + pr + " — updated by the push")
				continue
			}
			url, ok := c.prArg(o)
			if !ok {
				url, ok = c.ghPRCreate(w)
				ok = ok && strings.HasPrefix(url, "http")
			}
			if !ok {
				c.stop("pr", "no usable gh for "+w+" ("+url+") — print the forge's create-PR URL, ask whether it was opened, then re-run land adding --pr-url "+w+"=<url|none> to every --pr-url already given")
			}
			if url == "none" {
				continue
			}
			c.say("PR: " + url)
			if w == c.canon || pr == "" {
				pr = url
			}
		case "manual":
			c.say("PUSHED: " + w + " — spectre/" + c.name + "; the rest is the operator's")
		}
	}
	c.recordFixedRows()
	if !c.writeState("IN_PROGRESS", nil, pr) {
		c.stop("state", "the IN_PROGRESS write failed — report it")
	}
	c.markEnd("flow.landing-routes", "completed")
	if c.jiraKey != "" {
		c.say("JIRA: transition " + c.jiraKey + " to In Review")
	}
	if c.route == "merge and push" {
		c.cleanup()
	} else {
		c.say("NEXT: print the run 1 handoff — waiting on the merge")
	}
}
