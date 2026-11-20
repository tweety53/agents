package guard

import (
	"bytes"
	"fmt"
	"strings"
)

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

// recordLandBase stores the pre-push merge base for the post-land self-review range.
// A re-land keeps the first base: by then HEAD is its own merge base.
func (c *ic) recordLandBase(w string) {
	if v, _ := c.git("-C", w, "config", "--get", "branch.spectre/"+c.name+".flowLandBase"); v != "" {
		return
	}
	if mb, ok := c.git("-C", w, "merge-base", "HEAD", "origin/"+c.base); ok && mb != "" {
		c.gitPass("-C", w, "config", "branch.spectre/"+c.name+".flowLandBase", mb)
	}
}

func (c *ic) land() {
	c.canonBase = ""
	c.loadSet()
	if !c.findCanonical() {
		c.die("no-worktree-in-the-set", "no worktree in the set holds %s's tasks.md", c.name)
	}
	c.mergeOrder()
	if c.route == "" {
		c.route = c.defaultRoute()
	}
	switch c.route {
	case "pull request", "merge and push", "manual":
	default:
		c.stopAs("landing-route-not-given", "landing-route", "no route — ask the landing question, then flow-guard integrate-change land "+c.main+" "+c.name+" --route <route> "+c.flags())
	}
	c.openStage = "flow.land"
	c.reopenLand()
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
		c.pushBranch(w)
		if w == c.canon {
			c.canonBase = c.base
		}
		switch c.route {
		case "merge and push":
			c.recordLandBase(w)
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
					c.stopAs("sync-failed", "sync", fmt.Sprintf("sync-onto-base exit %d — report its output", c.rc))
				}
				c.rebaseLint(w, "already pushed: "+alreadyPushed()+"; ")
				c.pushBranch(w)
				c.recordLandBase(w)
				if !c.gitPass("-C", w, "push", "origin", "HEAD:"+c.base) {
					c.stopAs("push-rejected", "push", "the second push of "+w+" to origin/"+c.base+" was rejected — report git's output; the change stays IN_PROGRESS")
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
				c.stopAs("gh-unusable", "pr", "no usable gh for "+w+" ("+url+") — print the forge's create-PR URL, ask whether it was opened, then re-run land adding --pr-url "+w+"=<url|none> to every --pr-url already given")
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
		c.stopAs("state-write-failed", "state", "the IN_PROGRESS write failed — report it")
	}
	c.markEnd("flow.land", "completed")
	c.markBegin("flow.self-review")
	pass := "run the self-review pass over `flow self-review bundle -C " + c.canon + " -change " + c.name + "`"
	if c.route == "merge and push" {
		c.say("NEXT: " + pass + ", end flow.self-review completed, then flow-guard integrate-change cleanup " + c.main + " " + c.name + " " + c.flags())
	} else {
		c.say("NEXT: " + pass + ", then print the run 1 handoff — waiting on the merge")
	}
}
