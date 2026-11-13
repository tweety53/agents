package guard

import (
	"fmt"
	"io"
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
// round's launch message (on pass 1, every slot but mutation+failure-modes,
// which skills/flow/review-panel.md launches after the suite exits) — hooks/enforce-agent-baseline.py denies a launch
// whose row is not open yet (skills/flow/review-panel.md, "Every slot's
// dispatch is recorded"). <round> is the number itself, never a substitution.
//
// stdout: one `<key>: <flow's answer>` line per slot, in slot order; each
// begin writes that slot's `dispatched:` pass row itself (kan-1086
// self-review, F8).
//
// Exit codes: 0 every begin recorded; 1 a begin failed (its output relayed,
// the later slots still recorded); 2 usage, or no flow binary on PATH; 3 a
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
// Exit 0 every begin recorded; 1 a begin failed (its output relayed, the
// rest still recorded); 2 usage, or no flow binary on PATH; 3 a slot off the
// fix run's judged roster, no begin recorded.
func init() {
	Registry["panel-round-begins"] = panelRoundBegins
}

const prbUsage = "usage: flow-guard panel-round-begins <worktree> <change> <round> <session-token> <slot>:<model>:<effort>[:<diff-base>]...\n"

var prbRound = regexp.MustCompile(`^[0-9]+$`)

func panelRoundBegins(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) < 5 || !prbRound.MatchString(args[2]) {
		fmt.Fprint(stderr, prbUsage)
		return 2
	}
	wt, change, round, token := args[0], args[1], args[2], args[3]
	type slot struct{ name, model, effort, diffBase string }
	var slots []slot
	for _, spec := range args[4:] {
		f := strings.Split(spec, ":")
		if len(f) < 3 || len(f) > 4 || f[0] == "" || f[1] == "" || f[2] == "" {
			fmt.Fprintf(stderr, "panel-round-begins: slot %q is not <slot>:<model>:<effort>[:<diff-base>]\n%s", spec, prbUsage)
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
	if v, ok := fixRunVerdict(env, smcAbs(env, wt)); ok && strconv.Itoa(v.Round) == round {
		for _, s := range slots {
			for _, role := range strings.Split(s.name, "+") {
				if !slices.Contains(v.Roster, role) {
					fmt.Fprintf(stderr, "panel-round-begins: slot %s carries %s, which this fix run's judged roster (%s) omits — no begin recorded (flow-guard fix-run-floor)\n", s.name, role, strings.Join(v.Roster, ","))
					return 3
				}
			}
		}
	}
	flow, ok := lookPath(env, "flow")
	if !ok {
		fmt.Fprintln(stderr, "panel-round-begins: no flow binary on PATH")
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
			code = 1
		}
	}
	return code
}
