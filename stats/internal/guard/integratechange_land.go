package guard

import (
	"bytes"
	"fmt"
	"strings"
)

// selfReviewCompleted asks flowd whether the change holds a completed
// flow.self-review stage run -- the self-review pass's land marker, in
// place of a committed report (KAN-1073). ok false when flowd cannot answer.
func (c *ic) selfReviewCompleted() (done, ok bool) {
	cmd := c.cmd("/", "flow", "stage", "completed", "-C", c.main, "-stage", "flow.self-review", c.name)
	cmd.Stdout = c.stdout
	switch rcOf(cmd.Run()) {
	case 0:
		return true, true
	case 1:
		return false, true
	default:
		return false, false
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
	done, ok := c.selfReviewCompleted()
	if !ok {
		c.die("flowd cannot say whether flow.self-review completed for %s — nothing was pushed", c.name)
	}
	if !done {
		c.stop("self-review-report", "flow.self-review has not completed for "+c.name+" — run the self-review pass and end its stage completed, then re-run land; nothing was pushed")
	}
	if c.route == "" {
		c.route = c.defaultRoute()
	}
	switch c.route {
	case "pull request", "merge and push", "manual":
	default:
		c.stop("landing-route", "no route — ask the landing question, then integrate-change.sh land "+c.main+" "+c.name+" --route <route> "+c.flags())
	}
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
