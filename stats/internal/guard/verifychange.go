package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// flow-guard verify-change — the mechanical half of /flow's **Verify** stage
// (skills/flow/verify-and-handoff.md): the stage's begin mark, the worktree
// set, the workspace check, the base check, both verify lists per worktree
// with their dispatch rows, and the ledger render, sequenced through the
// existing guards so the parent does only what needs judgement — a failing
// list's triage and fix, the Live check, the visual stage.
//
// Usage: flow-guard verify-change <main-checkout> <name> --model <model> --effort <effort>
//
//	[--key <key>] [--cache-index-claimed] [--harness <harness>] [--session-token <literal token>]
//
// --model and --effort are the parent's own session pair, written into each
// list's `verifier` row. --key is the row's key, `verify` by default; a re-run
// passes `verify-<k>` or a fix round's `verify-fix-<k>` (skills/flow/fix.md).
// --harness and --session-token default as integrate-change's do (envHarness,
// envToken).
//
// In order, every step's lines printed as its guard prints them:
//
//  1. `flow stage begin` flow.verify (/flow). A warning on stderr when the
//     store refuses it, never a stop.
//  2. The worktree set: the state record's `worktrees` map, else the
//     porcelain scan (integrate-change's loadSet), less every path no longer
//     on disk; the canonical worktree is the one whose change directory holds
//     tasks.md. None → STOP: no-worktrees.
//  3. `prepare-workspace` per worktree: non-zero → STOP: prepare-workspace.
//     Its stderr naming a `(cache index)` row → STOP: cache-index, unless
//     --cache-index-claimed says the parent claimed it (**The cache index**,
//     skills/flow-contracts/workspace-isolation.md).
//  4. On the default key only — the run's first verify — `check-base-behind`
//     over the set: exit 1 runs `sync-base-behind` over it, whose non-zero
//     exit is STOP: base-sync; exit 2 prints `BASE-BEHIND-UNKNOWN: …` and goes on.
//  5. Per worktree: in a set of more than one, a worktree whose
//     `git diff --name-only <merge base>..HEAD` outside its own change
//     directory prints nothing is skipped, `SKIPPED: <worktree> — nothing
//     outside its change directory`; otherwise `run-verify-list <worktree>
//     lint` and `… test` run at once with the five recording arguments (plus
//     the canonical worktree for any other worktree), the key suffixed
//     `-<repository name>` in a set of more than one. Each list's output is
//     printed whole, lint first, then `VERIFY: <worktree> <list> — exit <n>`.
//     Exit 3 (no list declared) prints `AUTO-DETECT: <worktree> <list>` and
//     goes on. After every worktree, any other non-zero exit is STOP:
//     verify-failed.
//  6. `flow record render -change <name> -kind ledger -repo <canonical>`, its
//     outcome word relayed; a failed render prints `LEDGER-RENDER-FAILED: …`.
//  7. The closing NEXT names, in order and `; `-joined: `auto-detect <worktree>
//     <list>` per exit 3, `live-check` when the canonical design.md carries a
//     `## Live check` section whose first line opens with none of `None`,
//     `No ` or `Not applicable`, and `visual-verification` always.
//
// Ends on exactly one machine-readable line:
//
//	NEXT: <steps>                       exit 0
//	STOP: <kind> — <what the parent does>  exit 1
//
// Exit 0 every list ran green (or was skipped or undeclared); 1 a stop; 2
// usage, an unplain change name, or a state record that cannot be read. No
// exit ends flow.verify: on success the stage stays open through visual
// verification and the IN_PROGRESS write, and on a stop the fix procedure's
// step 1 (skills/flow/fix.md) closes it `stopped`.
//
// Go, in flow-guard (stats/internal/guard/verifychange.go).
func init() { Registry["verify-change"] = verifyChange }

func verifyChange(args []string, env Env, stdout, stderr io.Writer) int {
	c := &ic{env: env, stdout: stdout, stderr: stderr, prog: "verify-change", stopOutcome: "stopped",
		bases: map[int]string{}, rb: map[int]string{}, squashed: map[int][]string{}}
	c.childEnv = icChildEnv(env)
	return icCode(c.verify(args))
}

func (c *ic) verify(args []string) error {
	model, effort, key, claimed, err := c.parseVerify(args)
	if err != nil {
		return err
	}
	if c.env.Dir != "/" {
		if err := os.Chdir("/"); err != nil {
			return c.die("cannot-cd", "cannot cd /: %v", err)
		}
	}
	c.markBegin("flow.verify")
	c.openStage = "" // a stop leaves flow.verify open: fix.md step 1 closes it

	if err := c.loadSet(); err != nil {
		return err
	}
	var wts, mbs, repos []string
	for i, w := range c.wt {
		if isDir(w) {
			wts, mbs, repos = append(wts, w), append(mbs, c.mb[i]), append(repos, c.repo[i])
		}
	}
	if len(wts) == 0 || !c.findCanonical() {
		return c.stop("no-worktrees", "no worktree of "+c.name+" on disk holds its tasks.md — report the empty resolved set (worktree-resolution.md) and do not proceed")
	}

	for _, w := range wts {
		var errOut bytes.Buffer
		cmd := c.sibling("prepare-workspace", w)
		cmd.Stderr = io.MultiWriter(c.stderr, &errOut)
		out, rc := c.capture(cmd)
		if rc != 0 {
			fmt.Fprintln(c.stdout, out)
			return c.stopAs("prepare-workspace-failed", "prepare-workspace", fmt.Sprintf("flow-guard prepare-workspace %s exited %d — relay its lines and stop before any list runs, the state file unwritten", w, rc))
		}
		if strings.Contains(errOut.String(), "(cache index)") && !claimed {
			return c.stopAs("cache-index-unresolved", "cache-index", w+" declares a cache index row — claim it per **The cache index** (workspace-isolation.md), then run this driver again with --cache-index-claimed")
		}
	}

	if key == "verify" {
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
	}

	var next []string
	failed := false
	multi := len(wts) > 1
	for i, w := range wts {
		spec := specRootLeaf(w, io.Discard)
		if multi && mbs[i] != "-" {
			if d, _ := c.git("-C", w, "diff", "--no-renames", "--name-only", mbs[i]+"..HEAD", "--", ".", ":(exclude)"+spec+"/changes/"+c.name); d == "" {
				c.say("SKIPPED: " + w + " — nothing outside its change directory")
				continue
			}
		}
		k := key
		if multi {
			r := repos[i]
			if r == "" {
				r = w
			}
			k += "-" + filepath.Base(r)
		}
		rec := []string{c.name, k, c.token, model, effort}
		if w != c.canon {
			rec = append(rec, c.canon)
		}
		lists := []string{"lint", "test"}
		outs := make([]bytes.Buffer, len(lists))
		rcs := make([]int, len(lists))
		var wg sync.WaitGroup
		for j, l := range lists {
			cmd := siblingCmd(c.env, "FLOW_GUARD_RUN_VERIFY_LIST", "run-verify-list", append([]string{w, l}, rec...)...)
			cmd.Dir, cmd.Env = "/", c.childEnv
			cmd.Stdout, cmd.Stderr = &outs[j], &outs[j]
			wg.Add(1)
			go func() {
				defer wg.Done()
				rcs[j] = rcOf(cmd.Run())
			}()
		}
		wg.Wait()
		for j, l := range lists {
			if s := strings.TrimRight(outs[j].String(), "\n"); s != "" {
				c.say(s)
			}
			c.say(fmt.Sprintf("VERIFY: %s %s — exit %d", w, l, rcs[j]))
			switch rcs[j] {
			case 0:
			case 3:
				c.say("AUTO-DETECT: " + w + " " + l)
				next = append(next, "auto-detect "+w+" "+l)
			default:
				failed = true
			}
		}
	}
	if failed {
		return c.stop("verify-failed", "a list exited non-zero, its VERIFY: line above: 1 — read its VERDICT: block and act per **Inline verify — a failing command** (verify-and-handoff.md); 4 — the store refused the row's begin, clear its cause and run this driver again; 5 — the load never dropped, run it again; 2 — its stderr names the cause. flow.verify stays open")
	}

	render := c.cmd("/", "flow", "record", "render", "-C", c.main, "-change", c.name, "-kind", "ledger", "-repo", c.canon)
	render.Stdout = c.stdout
	if render.Run() != nil {
		c.say("LEDGER-RENDER-FAILED: a destination was refused or could not be written — report it")
	}

	if vcLiveCheck(c.canon + "/" + c.spec + "/changes/" + c.name + "/design.md") {
		next = append(next, "live-check")
	}
	c.say("NEXT: " + strings.Join(append(next, "visual-verification"), "; "))
	return nil
}

func (c *ic) parseVerify(args []string) (model, effort, key string, claimed bool, err error) {
	usage := "usage: flow-guard verify-change <main-checkout> <name> --model <model> --effort <effort> [--key <key>] [--cache-index-claimed] [--harness <h>] [--session-token <t>]"
	if len(args) < 2 {
		return "", "", "", false, c.die("usage", "%s", usage)
	}
	m, ok := icPhys(c.env.Dir, args[0])
	if !ok {
		return "", "", "", false, c.die("not-a-directory", "%s is not a directory", args[0])
	}
	c.main = m
	if !changeNamePlain(args[1]) {
		return "", "", "", false, c.die("change-name-not-plain", "change name '%s' is not a plain change name", args[1])
	}
	c.name = args[1]
	c.harness, c.token = envHarness(c.env), envToken(c.env)
	key = "verify"
	rest := args[2:]
	for i := 0; i < len(rest); i++ {
		if rest[i] == "--cache-index-claimed" {
			claimed = true
			continue
		}
		if i+1 >= len(rest) {
			return "", "", "", false, c.die("usage", "%s", usage)
		}
		v := rest[i+1]
		switch rest[i] {
		case "--model":
			model = v
		case "--effort":
			effort = v
		case "--key":
			key = v
		case "--harness":
			c.harness = v
		case "--session-token":
			c.token = v
		default:
			return "", "", "", false, c.die("unknown-argument", "unknown argument '%s'", rest[i])
		}
		i++
	}
	if model == "" || effort == "" {
		return "", "", "", false, c.die("usage", "%s", usage)
	}
	if c.harness == "" || c.token == "" {
		return "", "", "", false, c.die("harness-or-token-missing", "--harness and --session-token are required (or FLOW_HARNESS or CLAUDECODE=1 / FLOW_SESSION_TOKEN or a Claude Code session)")
	}
	return model, effort, key, claimed, nil
}

// vcLiveCheck is whether design.md carries a `## Live check` section naming
// something to exercise: present, and its first non-blank line opening with
// none of the no-runtime forms (**Live check**, verify-and-handoff.md).
func vcLiveCheck(design string) bool {
	b, err := os.ReadFile(design)
	if err != nil {
		return false
	}
	on := false
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "## ") {
			if on {
				return false
			}
			on = strings.TrimSpace(l) == "## Live check"
			continue
		}
		if !on || strings.TrimSpace(l) == "" {
			continue
		}
		l = strings.TrimSpace(l)
		return !strings.HasPrefix(l, "None") && !strings.HasPrefix(l, "No ") && !strings.HasPrefix(l, "Not applicable")
	}
	return false
}
