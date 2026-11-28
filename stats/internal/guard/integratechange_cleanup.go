package guard

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

var (
	icCellOpen  = regexp.MustCompile("^[ ]*`")
	icCellClose = regexp.MustCompile("`[ ]*$")
)

// workspaceCommand: the Runs cell of <verb>'s row in <main>'s
// `## workspace isolation` command table, or nothing.
func (c *ic) workspaceCommand(m, verb string) string {
	get := c.sibling("project-get", m, "workspace isolation")
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

// selfReviewPassRow: the change holds a completed self-review-<name>-pass
// dispatch row, the pass the flow.self-review stage mark stands for.
func (c *ic) selfReviewPassRow() (bool, error) {
	out, rc := c.capture(c.cmd("/", "flow", "record", "dispatches", "-change", c.name, "-C", c.main))
	if rc != 0 {
		return false, c.die("self-review-state-unknown", "flowd cannot list the dispatches of %s — nothing was removed", c.name)
	}
	var rows []struct{ Key, Outcome string }
	if json.Unmarshal([]byte(out), &rows) != nil {
		return false, nil
	}
	for _, r := range rows {
		if r.Key == "self-review-"+c.name+"-pass" && r.Outcome == "completed" {
			return true, nil
		}
	}
	return false, nil
}

// icWithID: <command> with the workspace id tokens substituted.
func icWithID(command, id string) string {
	command = strings.ReplaceAll(command, "<id_underscored>", strings.ReplaceAll(id, "-", "_"))
	return strings.ReplaceAll(command, "<id>", id)
}

// icWorkspaceID is the id alphabet; anything outside it never reaches a
// shell command.
var icWorkspaceID = regexp.MustCompile(`^[A-Za-z0-9_.-]*$`)

func (c *ic) cleanup() error {
	rbase := map[int]string{}
	if err := c.loadSet(); err != nil {
		return err
	}
	done, ok := c.selfReviewCompleted()
	if !ok {
		return c.die("self-review-state-unknown", "flowd cannot say whether flow.self-review completed for %s — nothing was removed", c.name)
	}
	if !done {
		return c.stopAs("self-review-not-done", "self-review-report", "flow.self-review has not completed for "+c.name+" — run the self-review pass and end its stage completed, then re-run cleanup; nothing was removed")
	}
	passed, err := c.selfReviewPassRow()
	if err != nil {
		return err
	}
	if !passed {
		return c.stopAs("self-review-pass-missing", "self-review-report", "flow.self-review completed for "+c.name+" with no completed self-review-"+c.name+"-pass dispatch row — run the self-review pass, then re-run cleanup; nothing was removed")
	}
	c.markBegin("flow.cleanup")
	// The save keeps tasks.md and self-review-context.md in flowd before
	// anything removes a worktree; a re-run past a removed canonical worktree
	// already saved, since remove-change-worktrees refuses an unsaved one.
	if c.findCanonical() {
		out, rc := c.capture(c.cmd("/", "flow", "change", "save", "-C", c.canon, c.name))
		if out != "" {
			fmt.Fprintln(c.stdout, out)
		}
		if rc != 0 {
			return c.stopAs("change-save-failed", "change-save", fmt.Sprintf("flow change save exit %d for %s — report its lines, restore each file it names, re-run cleanup; nothing was removed", rc, c.name))
		}
	}
	for i := range c.wt {
		if !core.IsDir(c.wt[i]) {
			continue
		}
		if err := c.baseOf(i); err != nil {
			return err
		}
		rbase[i] = c.base
	}
	for _, m := range c.mains {
		c.run("remove-change-worktrees", m, c.name)
		if c.rc != 0 {
			return c.stopAs("worktree-removal-failed", "worktree-cleanup", fmt.Sprintf("remove-change-worktrees exit %d for %s — report its lines; every worktree it refused is left alone", c.rc, m))
		}
	}
	id, rc := c.capture(c.cmd("/", "flow", "workspace-id", c.name))
	if rc != 0 {
		id = ""
	}
	for _, m := range c.mains {
		row, _ := c.removeWorkspace(m, id)
		if row == "" {
			continue
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
	if art := sd + "/" + c.name + "-proposal-artifact.html"; core.IsFile(art) {
		os.Remove(art)
		c.say("PROPOSAL-ARTIFACT: removed")
	} else {
		c.say("PROPOSAL-ARTIFACT: none to remove")
	}
	keep := []string{}
	for _, w := range c.wt {
		if core.IsDir(w) {
			keep = append(keep, w)
		}
	}
	if !c.writeState("FINISHED", keep, "") {
		return c.stopAs("state-write-failed", "state", "the FINISHED write failed — report it")
	}
	if c.jiraKey != "" {
		c.say("JIRA: transition " + c.jiraKey + " to Done")
	}

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
	c.markEnd("flow.cleanup", "completed")
	c.say("NEXT: print the Finished handoff (skills/flow/cleanup.md)")
	return nil
}

// removeWorkspace runs <m>'s `remove` row for workspace id, printing its
// verdict line: the row, "" when none is declared or id is empty, and
// whether the row ran and failed. Both integrate-change cleanup and
// workspace-remove tear a change's workspace down through it.
func (c *ic) removeWorkspace(m, id string) (row string, failed bool) {
	row = c.workspaceCommand(m, "remove")
	if row == "" || id == "" {
		c.say("WORKSPACE: " + m + " — skipped, no remove command declared")
		return "", false
	}
	if rc := c.bashIn(m, icWithID(row, id)); rc != 0 {
		c.say(fmt.Sprintf("WORKSPACE-REMOVE-FAILED: %s — exit %d", m, rc))
		return row, true
	}
	c.say("WORKSPACE-REMOVED: " + m)
	return row, false
}

// bashIn is `(cd <dir> && bash -c <command>)`, both streams passed
// through: its exit code.
func (c *ic) bashIn(dir, command string) int {
	cmd := c.cmd(dir, "bash", "-c", command)
	cmd.Stdout = c.stdout
	return rcOf(cmd.Run())
}
