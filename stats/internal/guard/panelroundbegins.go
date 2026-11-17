package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/panel-round-begins.sh shim (KAN-1058).
//
// flow-guard panel-round-begins <worktree> <change> <round> <session-token> <slot>:<model>:<effort>[:<diff-base>]...
// — records every slot's begin of one review-panel round, each under the
// literal key panel-<round>-<slot>, in the one Bash call just before the
// round's launch message (on pass 1, every slot but a mutation-carrying extra
// bundle, which skills/flow/review-panel.md launches after the suite exits) — hooks/enforce-agent-baseline.py denies a launch
// whose row is not open yet (skills/flow/review-panel.md, "Every slot's
// dispatch is recorded"). <round> is the number itself, never a substitution.
//
// stdout: one `<key>: <flow's answer>` line per slot, in slot order; each
// begin writes that slot's `dispatched:` pass row itself (kan-1086
// self-review, F8).
//
// Before any begin it snapshots the plan tree to
// <worktree>/.superpowers/sdd/plan-snapshot-panel-<round>.txt
// (flow-guard check-plan-unchanged snapshot), overwriting a file an earlier
// invocation left; pass 1's later call, recording a mutation-carrying extra
// bundle alone, keeps the round's file (self-review-kan-1110 F16).
//
// Exit codes: 0 every begin recorded; 1 a begin failed (its output relayed,
// the later slots still recorded), or a round of 1 or later carries no
// mutation-carrying slot while no earlier round's completed one
// (`mutation-pass-missing`, stderr naming mutation+failure-modes, no begin
// recorded; a judged roster omitting mutation owes none -- kan-1099
// self-review, F24); 2 usage, the change's dispatch rows unreadable, no flow binary on PATH, or the
// plan-tree snapshot failed, no begin recorded; 3 a
// slot carries a role the fix run's judged roster omits -- in the round
// <worktree>/.superpowers/sdd/fix-run.json governs (flow-guard fix-run-floor,
// whose header is canonical for that file), stderr names the slot and role
// and no begin is recorded, so hooks/enforce-agent-baseline.py denies the
// launch.
//
// Go, in flow-guard (stats/internal/guard/panelroundbegins.go).

// panel-round-begins (flow-guard panel-round-begins) records every slot's
// begin of one review-panel round, each under the literal key
// panel-<round>-<slot> -- the round's begins block the parent hand-wrote
// before, where kan-1026's `-key panel-$R-$s` from `R=$(flow record
// next-round ...)` left the daemon nothing to read the key from
// (skills/flow/review-panel.md, "Every slot's dispatch is recorded").
// Each begin writes its own slot's `dispatched: panel-<round>-<slot> on
// <model>/<effort>[ over <diff-base>]` pass row (`flow record dispatch
// begin`), so this guard writes none (kan-1086 self-review, F8; kan-1033
// self-review moved the row out of hand-recording first).
// It first snapshots the plan tree to
// <worktree>/.superpowers/sdd/plan-snapshot-panel-<round>.txt, overwriting
// a stale one; a call recording only a mutation-carrying extra bundle keeps
// the round's file (self-review-kan-1110 F16).
// Exit 0 every begin recorded; 1 a begin failed (its output relayed, the
// rest still recorded) or a mutation pass is owed (no begin recorded); 2
// usage, unreadable dispatch rows, no flow binary on PATH, or a failed
// snapshot, no begin recorded; 3 a slot off the fix run's judged roster, no
// begin recorded.
func init() {
	Registry["panel-round-begins"] = panelRoundBegins
}

const prbUsage = "usage: flow-guard panel-round-begins <worktree> <change> <round> <session-token> <slot>:<model>:<effort>[:<diff-base>]...\n"

var prbRound = regexp.MustCompile(`^[0-9]+$`)

// prbPanelKey splits a panel slot key panel-<round>-<slot> into round and slot.
var prbPanelKey = regexp.MustCompile(`^panel-([0-9]+)-(.+)$`)

func panelRoundBegins(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 5 || !prbRound.MatchString(args[2]) {
		fmt.Fprint(stderr, prbUsage)
		Reason(env, "usage")
		return 2
	}
	wt, change, round, token := args[0], args[1], args[2], args[3]
	type slot struct{ name, model, effort, diffBase string }
	var slots []slot
	for _, spec := range args[4:] {
		f := strings.Split(spec, ":")
		if len(f) < 3 || len(f) > 4 || f[0] == "" || f[1] == "" || f[2] == "" {
			fmt.Fprintf(stderr, "panel-round-begins: slot %q is not <slot>:<model>:<effort>[:<diff-base>]\n%s", spec, prbUsage)
			Reason(env, "bad-slot")
			return 2
		}
		s := slot{name: f[0], model: f[1], effort: f[2]}
		if len(f) == 4 {
			s.diffBase = f[3]
		}
		slots = append(slots, s)
	}
	// THE JUDGED ROSTER: in the round a fix run's verdict governs, a slot
	// carrying a role the verdict's roster omits is refused before any begin.
	v, judged := fixRunVerdict(env, smcAbs(env, wt))
	judged = judged && strconv.Itoa(v.Round) == round
	if judged {
		for _, s := range slots {
			for _, role := range strings.Split(s.name, "+") {
				if !slices.Contains(v.Roster, role) {
					fmt.Fprintf(stderr, "panel-round-begins: slot %s carries %s, which this fix run's judged roster (%s) omits — no begin recorded (flow-guard fix-run-floor)\n", s.name, role, strings.Join(v.Roster, ","))
					Reason(env, "slot-role-off-roster")
					return 3
				}
			}
		}
	}
	// THE OWED MUTATION PASS: a round of 1 or later begins only once a
	// mutation-carrying slot of an earlier round ended completed, or carries
	// one itself -- a pass that never ran no mutant is owed to the next round
	// (kan-1099 self-review, F24: rounds 0 to 3 ran none). A judged roster
	// omitting mutation owes none.
	if round != "0" && !(judged && !slices.Contains(v.Roster, "mutation")) &&
		!slices.ContainsFunc(slots, func(s slot) bool { return slices.Contains(strings.Split(s.name, "+"), "mutation") }) {
		out, _, rc := pcFlow(env, false, "record", "dispatches", "-C", wt, "-change", change)
		var rows []struct{ Key, Outcome string }
		if rc != 0 || json.Unmarshal(out, &rows) != nil {
			fmt.Fprintln(stderr, "panel-round-begins: the change's dispatch rows could not be read — no begin recorded")
			Reason(env, "dispatches-unreadable")
			return 2
		}
		n, _ := strconv.Atoi(round)
		ran := slices.ContainsFunc(rows, func(r struct{ Key, Outcome string }) bool {
			m := prbPanelKey.FindStringSubmatch(r.Key)
			if m == nil || r.Outcome != "completed" {
				return false
			}
			k, _ := strconv.Atoi(m[1])
			return k < n && slices.Contains(strings.Split(m[2], "+"), "mutation")
		})
		if !ran {
			fmt.Fprintf(stderr, "panel-round-begins: no mutation pass has completed before round %s — dispatch mutation+failure-modes in this round; no begin recorded\n", round)
			Reason(env, "mutation-pass-missing")
			return 1
		}
	}
	// THE ROUND'S PLAN-TREE SNAPSHOT is taken here by the round's first call,
	// overwriting any file an earlier invocation left under the same round
	// number. Pass 1's later call, recording a mutation-carrying extra bundle
	// alone -- roles within {mutation, failure-modes}, mutation among them, as
	// the slot triggers may shrink it -- keeps it, so a tree the first wave's
	// slots changed is never re-snapshotted (self-review-kan-1110 F16).
	snap := filepath.Join(smcAbs(env, wt), ".superpowers", "sdd", "plan-snapshot-panel-"+round+".txt")
	later := true
	for _, s := range slots {
		roles := strings.Split(s.name, "+")
		later = later && slices.Contains(roles, "mutation") && !slices.ContainsFunc(roles, func(r string) bool {
			return r != "mutation" && r != "failure-modes"
		})
	}
	if _, err := os.Stat(snap); err != nil || !later {
		if os.MkdirAll(filepath.Dir(snap), 0o755) != nil || checkPlanUnchanged([]string{"snapshot", wt, change, snap}, env, io.Discard, stderr) != 0 {
			fmt.Fprintln(stderr, "panel-round-begins: plan-tree snapshot failed — no begin recorded")
			Reason(env, "plan-tree-snapshot-failed")
			return 2
		}
	}
	flow, ok := lookPath(env, "flow")
	if !ok {
		fmt.Fprintln(stderr, "panel-round-begins: no flow binary on PATH")
		Reason(env, "no-flow-binary-on-path")
		return 2
	}
	code := 0
	for _, s := range slots {
		key := "panel-" + round + "-" + s.name
		argv := []string{"record", "dispatch", "begin", "-C", wt, "-change", change, "-role", "reviewer",
			"-slot", s.name, "-model", s.model, "-effort", s.effort}
		if s.diffBase != "" {
			argv = append(argv, "-diff-base", s.diffBase)
		}
		argv = append(argv, "-key", key, "-session-token", token)
		rc, out := kwRun(env, "", flow, argv...)
		fmt.Fprintf(stdout, "%s: %s\n", key, strings.TrimSpace(out))
		if rc != 0 {
			Reason(env, "dispatch-begin-failed")
			code = 1
		}
	}
	return code
}
