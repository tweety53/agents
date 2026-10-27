package guard

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	icCellOpen  = regexp.MustCompile("^[ ]*`")
	icCellClose = regexp.MustCompile("`[ ]*$")
)

// workspaceCommand: the Runs cell of <verb>'s row in <main>'s
// `## workspace isolation` command table, or nothing.
func (c *ic) workspaceCommand(m, verb string) string {
	get := c.cmd("/", c.sibling("project-get"), m, "workspace isolation")
	get.Stderr = nil
	out, _ := c.capture(get)
	for _, l := range icLines(out) {
		f := strings.Split(l, "|")
		if len(f) < 2 {
			continue
		}
		if strings.NewReplacer(" ", "", "`", "").Replace(f[1]) != verb {
			continue
		}
		r := ""
		if len(f) > 2 {
			r = f[2]
		}
		return icCellClose.ReplaceAllString(icCellOpen.ReplaceAllString(r, ""), "")
	}
	return ""
}

// icWithID: <command> with the workspace id tokens substituted.
func icWithID(command, id string) string {
	command = strings.ReplaceAll(command, "<id_underscored>", strings.ReplaceAll(id, "-", "_"))
	return strings.ReplaceAll(command, "<id>", id)
}

// icWorkspaceID is the id alphabet; anything outside it never reaches a
// shell command.
var icWorkspaceID = regexp.MustCompile(`^[A-Za-z0-9_.-]*$`)

func (c *ic) cleanup() {
	rbase := map[int]string{}
	c.loadSet()
	c.markBegin("flow.verify-merge")
	for i := range c.wt {
		if !isDir(c.wt[i]) {
			continue
		}
		c.baseOf(i)
		rbase[i] = c.base
		v := ""
		if _, ok := lookPath(c.env, "gh"); c.prURLState != "" && ok {
			gh := c.cmd("/", "gh", "pr", "view", c.prURLState, "--json", "state", "-q", ".state")
			gh.Stderr = nil
			if out, rc := c.capture(gh); rc == 0 {
				v = out
			}
		}
		if v == "MERGED" {
			continue
		}
		if _, ok := c.git("-C", c.wt[i], "merge-base", "--is-ancestor", "HEAD", "origin/"+c.base); ok {
			continue
		}
		c.stopOutcome = "not-run-2"
		c.stop("not-merged", c.wt[i]+" is not merged into origin/"+c.base+" — this is not run 2; remove nothing and run prepare")
	}
	c.markEnd("flow.verify-merge", "completed")

	c.markBegin("flow.cleanup")
	for _, m := range c.mains {
		mb := "-"
		for i := range c.wt {
			if c.repo[i] == m {
				mb = c.mb[i]
				break
			}
		}
		args := []string{m, c.name, mb}
		if icHas(c.proceed, m) {
			args = append(args, "--proceed")
		}
		c.run("remove-change-worktrees", args...)
		switch c.rc {
		case 0:
		case 3:
			c.stop("disclose", "relay "+m+"'s UNCLASSIFIED and DISCLOSE lines, ask only what they ask, then re-run integrate-change.sh cleanup "+c.main+" "+c.name+" adding --proceed "+m+" to every --proceed already given "+c.flags())
		default:
			c.stop("worktree-cleanup", fmt.Sprintf("remove-change-worktrees exit %d for %s — report its lines; every worktree it refused is left alone", c.rc, m))
		}
	}
	id, rc := c.capture(c.cmd("/", "flow", "workspace-id", c.name))
	if rc != 0 {
		id = ""
	}
	for _, m := range c.mains {
		row := c.workspaceCommand(m, "remove")
		if row == "" || id == "" {
			c.say("WORKSPACE: " + m + " — skipped, no remove command declared")
			continue
		}
		if rc := c.bashIn(m, icWithID(row, id)); rc == 0 {
			c.say("WORKSPACE-REMOVED: " + m)
		} else {
			c.say(fmt.Sprintf("WORKSPACE-REMOVE-FAILED: %s — exit %d; the cleanup check decides", m, rc))
		}
		orow := c.workspaceCommand(m, "orphans")
		if orow == "" {
			continue
		}
		out, rc := c.capture(c.cmd(m, "bash", "-c", orow))
		if rc != 0 {
			c.say(fmt.Sprintf("WORKSPACE-ORPHANS-FAILED: %s — exit %d", m, rc))
			continue
		}
		for _, o := range strings.Split(out, "\n") {
			if o == "" {
				continue
			}
			// The id reaches a shell command; anything outside the id alphabet is refused.
			if !icWorkspaceID.MatchString(o) {
				c.say("WORKSPACE-ORPHAN-REMOVE-FAILED: " + o + " — refused, not an id")
				continue
			}
			if rc := c.bashIn(m, icWithID(row, o)); rc == 0 {
				c.say("WORKSPACE-ORPHAN-REMOVED: " + o)
			} else {
				c.say(fmt.Sprintf("WORKSPACE-ORPHAN-REMOVE-FAILED: %s — exit %d", o, rc))
			}
		}
	}
	sd, _ := c.capture(c.cmd("/", "flow", "state", "dir", "-C", c.main))
	if art := sd + "/" + c.name + "-proposal-artifact.html"; isFile(art) {
		os.Remove(art)
		c.say("PROPOSAL-ARTIFACT: removed")
	} else {
		c.say("PROPOSAL-ARTIFACT: none to remove")
	}
	c.markEnd("flow.cleanup", "completed")

	c.markBegin("flow.verify-cleanup")
	leftover := false
	for _, m := range c.mains {
		// flow state dir resolves the path and never creates it; a store-only
		// project has none on disk, which the guard would read as a typo.
		sd, rc := c.capture(c.cmd("/", "flow", "state", "dir", "-C", m))
		if rc == 0 {
			if err := os.MkdirAll(sd, 0o755); err != nil {
				fmt.Fprintf(c.stderr, "integrate-change: %v\n", err)
			}
		}
		c.run("check-cleanup-complete", m, c.name, sd)
		if c.rc != 0 || c.verdict("COMPLETE") == "" {
			leftover = true
		}
	}
	c.stopOutcome = "leftover"
	if leftover {
		c.stop("leftover", "name what remains; FINISHED is not written — re-run cleanup once the operator clears it")
	}
	c.stopOutcome = "stopped"
	c.markEnd("flow.verify-cleanup", "completed")

	c.markBegin("flow.write-finished")
	keep := []string{}
	for _, w := range c.wt {
		if isDir(w) {
			keep = append(keep, w)
		}
	}
	if !c.writeState("FINISHED", keep, "") {
		c.stop("state", "the FINISHED write failed — report it")
	}
	c.markEnd("flow.write-finished", "completed")
	if c.jiraKey != "" {
		c.say("JIRA: transition " + c.jiraKey + " to Done")
	}

	c.markBegin("flow.refresh-main-checkout")
	for _, m := range c.mains {
		base, _ := c.gitQuiet("-C", m, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
		base = strings.TrimPrefix(base, "origin/")
		// A main checkout on a --base change's base is refreshed onto it (KAN-911).
		cur, _ := c.gitQuiet("-C", m, "branch", "--show-current")
		hit := false
		for i, rb := range rbase {
			if c.repo[i] == m {
				hit = true
				if cur != "" && rb == cur {
					base = cur
				}
			}
		}
		// A re-run after the worktrees were removed has no rbase for m: fall back
		// to the checkout's own branch when origin carries it and it is not the
		// change's branch.
		if !hit && cur != "" && cur != "spectre/"+c.name {
			if _, ok := c.git("-C", m, "rev-parse", "--verify", "--quiet", "refs/remotes/origin/"+cur); ok {
				base = cur
			}
		}
		c.run("refresh-main-checkout", m, base)
	}
	c.markEnd("flow.refresh-main-checkout", "completed")
	c.say("NEXT: print the Finished handoff (skills/flow/cleanup.md) — its Learning: and Prompt tuning: lines included")
}

// bashIn is `(cd <dir> && bash -c <command>)`, both streams passed
// through: its exit code.
func (c *ic) bashIn(dir, command string) int {
	cmd := c.cmd(dir, "bash", "-c", command)
	cmd.Stdout = c.stdout
	return rcOf(cmd.Run())
}
