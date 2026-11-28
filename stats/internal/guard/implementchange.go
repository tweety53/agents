package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// flow-guard implement-change — the mechanical opening of /flow's implement
// phase (skills/flow/implement.md, **1. Load context and validate the plan**
// and the workspace id of **2. Isolate the workspace**), sequenced through the
// existing guards so the parent does only what needs judgement: reading the
// plan, repairing it, resolving `## apps`, and every dispatch of **4. Execute**.
//
// Usage: flow-guard implement-change prepare <main-checkout> <name>
//
//	[--harness <harness>] [--session-token <literal token>]
//
// --harness and --session-token default as integrate-change's do.
//
// prepare, in order, every step's lines printed as its guard prints them:
//
//  1. `flow stage begin` flow.implement (/flow).
//  2. The worktree set, as integrate-change's loadSet reads it, less every
//     path no longer on disk; the canonical worktree is the one whose change
//     directory holds tasks.md. None → STOP: no-worktrees.
//  3. `spectre validate <name>` in the canonical worktree: exit 1 → STOP:
//     validate (repair each finding, run prepare again); any other non-zero
//     exit → STOP: no-change.
//  4. `check-base-behind` over the set: exit 1 runs `sync-base-behind` over
//     it, whose non-zero exit is STOP: base-sync; exit 2 prints
//     `BASE-BEHIND-UNKNOWN: …` and goes on.
//  5. `flow workspace-id <name>`, printed as `WORKSPACE-ID: <id>`; a failed
//     call is a warning on stderr.
//  6. `spectre list --json`, this change's done and total, printed as
//     `TASKS: <done>/<total>`: total 0, or no entry → STOP: no-plan; done ==
//     total ends flow.implement `completed` and prints `NEXT: review-panel`;
//     otherwise `NEXT: execute`.
//
// Ends on exactly one machine-readable line:
//
//	NEXT: execute | review-panel          exit 0
//	STOP: <kind> — <what the parent does>  exit 1
//
// Exit 1 closes the open flow.implement mark `stopped`, its end carrying
// `-reason "STOP: <kind>"`; 2 usage, an unplain change name, or a state
// record that cannot be read.
//
// Go, in flow-guard (stats/internal/guard/implementchange.go).
func init() { Registry["implement-change"] = implementChange }

func implementChange(args []string, env core.Env, stdout, stderr io.Writer) int {
	c := &ic{env: env, stdout: stdout, stderr: stderr, prog: "implement-change", stopOutcome: "stopped",
		bases: map[int]string{}, rb: map[int]string{}, squashed: map[int][]string{}}
	c.childEnv = core.ChildEnv(env)
	return guardCode(c.implementPrepare(args))
}

func (c *ic) implementPrepare(args []string) error {
	if len(args) < 3 || args[0] != "prepare" {
		return c.die("usage", "usage: flow-guard implement-change prepare <main-checkout> <name> [--harness <h>] [--session-token <t>]")
	}
	if err := c.parse(args); err != nil {
		return err
	}
	if len(c.subjects) > 0 || c.route != "" || len(c.prArgs) > 0 {
		return c.die("usage", "usage: flow-guard implement-change prepare <main-checkout> <name> [--harness <h>] [--session-token <t>]")
	}
	if c.env.Dir != "/" {
		if err := os.Chdir("/"); err != nil {
			return c.die("cannot-cd", "cannot cd /: %v", err)
		}
	}
	c.markBegin("flow.implement")

	if err := c.loadSet(); err != nil {
		return err
	}
	var wts []string
	for _, w := range c.wt {
		if core.IsDir(w) {
			wts = append(wts, w)
		}
	}
	if len(wts) == 0 || !c.findCanonical() {
		return c.stop("no-worktrees", "no worktree of "+c.name+" on disk holds its tasks.md — report the empty resolved set (worktree-resolution.md) and do not proceed")
	}

	validate := c.cmd(c.canon, "spectre", "validate", c.name)
	validate.Stdout = c.stdout
	switch rc := rcOf(validate.Run()); rc {
	case 0:
	case 1:
		return c.stopAs("validate-failed", "validate", "spectre validate named findings in this change's own artifacts — repair each before any code is touched, then run prepare again")
	default:
		return c.stopAs("change-missing", "no-change", fmt.Sprintf("spectre validate exited %d — on `no such change`, nothing is proposed under %s: stop and suggest `/flow %s`", rc, c.name, c.name))
	}

	c.run("check-base-behind", wts...)
	switch c.rc {
	case 0:
	case 1:
		c.run("sync-base-behind", wts...)
		if c.rc != 0 {
			return c.stopAs("base-sync-failed", "base-sync", fmt.Sprintf("flow-guard sync-base-behind exited %d — name its stderr in the handoff and stop", c.rc))
		}
	default:
		c.say("BASE-BEHIND-UNKNOWN: check-base-behind could not answer — name its stderr in the handoff")
	}

	if id, rc := c.capture(c.cmd(c.canon, "flow", "workspace-id", c.name)); rc == 0 {
		c.say("WORKSPACE-ID: " + id)
	} else {
		fmt.Fprintf(c.stderr, "implement-change: warning: flow workspace-id %s exited %d\n", c.name, rc)
	}

	list, _ := c.capture(c.cmd(c.canon, "spectre", "list", "--json"))
	var parsed struct {
		Changes []struct {
			ID          string
			Done, Total int
		}
	}
	_ = json.Unmarshal([]byte(list), &parsed)
	done, total := 0, 0
	for _, ch := range parsed.Changes {
		if ch.ID == c.name {
			done, total = ch.Done, ch.Total
		}
	}
	c.say(fmt.Sprintf("TASKS: %d/%d", done, total))
	if total == 0 {
		return c.stopAs("plan-missing", "no-plan", "spectre reads no plan for "+c.name+" — resume at skills/flow/resume.md")
	}
	if done == total {
		c.markEnd("flow.implement", "completed")
		c.say("NEXT: review-panel")
		return nil
	}
	c.say("NEXT: execute")
	return nil
}
