package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// flow-guard check-plan-departures — refuse a plan that restyles a shipped
// view to a mockup frame without having measured the view against that
// frame first, or that leaves a measured departure out of every task
// (kan-1099: its plan restyled shipped screens with none measured, and the
// departures surfaced one verify round at a time — 20 rounds and 16
// appended tasks on a plan of 34).
//
// Usage: flow-guard check-plan-departures <worktree> <name> <tasks.md>
//
// Run once, when the plan is written (skills/flow/brainstorm-planner.md),
// never at a task's close: the manifest it reads is the planner's scratch.
//
// A SHIPPED VIEW is a capture spec an unticked task names by its sidecar — a
// backticked `<path>.mockups` token from the task's `- [ ] ` line to the
// next task's — for which
// flow-guard resolve-visual-screenshots <worktree> <spec basename> finds a
// capture (exit 0). A spec with no capture yet draws a view the change
// builds, and owes nothing; so does a ticked task, whose view the change
// itself captured — a plan revised mid-run is judged on what is left to do.
//
// THE PLAN-TIME MEASURE is
// <worktree>/.superpowers/sdd/visual-manifest-plan.json, what
// `flow-guard visual-verify-pass measure <worktree> <name> plan <spec path>`
// writes. Every shipped view needs a pair whose `spec` has its basename,
// and every `over` row of every pair is cited in <tasks.md> as
// `<frame> #<row>`, a whole token (citesWhole), by the task that closes it
// or names why it stays.
//
// Prints its verdict lines to stdout:
//
//	PLAN-MEASURE-MISSING: <spec basename>; …
//	PLAN-DEPARTURES-UNCITED: <frame> #<row>; …
//	PLAN-DEPARTURES-CITED: <n> row(s) of <m> measured spec(s)
//
// the last alone on exit 0; a MISSING line is followed on stderr by the
// measure call.
//
// Exit 0 nothing owed, 1 a view unmeasured or a row uncited, 2 cannot
// answer (usage, an unreadable <tasks.md>, a manifest that does not parse).
func init() { Registry["check-plan-departures"] = checkPlanDepartures }

var cpdSidecar = regexp.MustCompile("`([^`\\s]+)\\.mockups`")

func checkPlanDepartures(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-plan-departures"
	if len(args) != 3 {
		fmt.Fprintf(stderr, "%s: usage: flow-guard check-plan-departures <worktree> <name> <tasks.md>\n", prog)
		Reason(env, "usage")
		return 2
	}
	wt, name := args[0], args[1]
	plan, err := os.ReadFile(args[2])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[2], err)
		Reason(env, "cannot-read")
		return 2
	}
	var m struct {
		Pairs []struct {
			Spec, Frame string
			Over        []struct{ Row int }
		}
	}
	path := filepath.Join(wt, ".superpowers", "sdd", "visual-manifest-plan.json")
	if b, err := os.ReadFile(path); err == nil && json.Unmarshal(b, &m) != nil {
		fmt.Fprintf(stderr, "%s: %s does not parse — run the measure again\n", prog, path)
		Reason(env, "manifest-cannot-parse")
		return 2
	}
	var measured, missing, uncited []string
	rows := 0
	for _, p := range m.Pairs {
		if base := filepath.Base(p.Spec); !slices.Contains(measured, base) {
			measured = append(measured, base)
		}
		for _, o := range p.Over {
			rows++
			if cite := fmt.Sprintf("%s #%d", p.Frame, o.Row); !citesWhole(string(plan), cite) {
				uncited = append(uncited, cite)
			}
		}
	}
	open := false // inside an unticked task
	for _, line := range strings.Split(string(plan), "\n") {
		if strings.HasPrefix(line, "- [") {
			open = strings.HasPrefix(line, "- [ ] ")
		}
		if !open {
			continue
		}
		for _, s := range cpdSidecar.FindAllStringSubmatch(line, -1) {
			base := filepath.Base(s[1])
			if slices.Contains(measured, base) || slices.Contains(missing, base) {
				continue
			}
			var out bytes.Buffer
			if resolveVisualScreenshots([]string{wt, base}, quietReasons(env), &out, io.Discard) == 0 {
				missing = append(missing, base)
			}
		}
	}
	if len(missing) > 0 {
		fmt.Fprintf(stdout, "PLAN-MEASURE-MISSING: %s\n", strings.Join(missing, "; "))
		fmt.Fprintf(stderr, "%s: each already has captures, so its view ships: point its sidecar at the frames the plan draws, run `flow-guard visual-verify-pass measure %s %s plan <spec path>` on it, and write every `over` row, flagged band and seam into the task that restyles the view\n", prog, wt, name)
		Reason(env, "plan-measure-missing")
	}
	if len(uncited) > 0 {
		fmt.Fprintf(stdout, "PLAN-DEPARTURES-UNCITED: %s\n", strings.Join(uncited, "; "))
		Reason(env, "plan-departures-uncited")
	}
	if len(missing)+len(uncited) > 0 {
		return 1
	}
	fmt.Fprintf(stdout, "PLAN-DEPARTURES-CITED: %d row(s) of %d measured spec(s)\n", rows, len(measured))
	return 0
}
