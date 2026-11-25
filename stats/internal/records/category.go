package records

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

// CategoryRow is one row of the dispatch category table: a dispatch whose
// key matches Key -- and, when Opener is set, whose first task's
// **Opener:** starts with it -- is classed under ID and shown as Label.
type CategoryRow struct {
	ID, Label string
	Key       *regexp.Regexp
	Opener    string
}

// Categories is the one ordered table every reader classes a dispatch by,
// at read time; the first matching row wins and nothing is stored per row,
// so editing the table re-labels history. The key shapes are the ones
// describeDispatch (stats/web/src/lib/describeDispatch.ts) parses; the
// shared fixture testdata/dispatch-keys.json pins both sides.
var Categories = []CategoryRow{
	{ID: "gated-fixer", Label: "gated fixer", Key: regexp.MustCompile(`^tasks?-[\d+]+-implementer-fix-\d+$`)},
	{ID: "in-run-fix-visual-verify", Label: "in-run fix (visual verify)", Key: regexp.MustCompile(`^tasks?-[\d+]+-implementer`), Opener: "visual-verify"},
	{ID: "in-run-fix-verify", Label: "in-run fix (verify)", Key: regexp.MustCompile(`^tasks?-[\d+]+-implementer`), Opener: "verify"},
	{ID: "planned-implementer", Label: "planned implementer", Key: regexp.MustCompile(`^tasks?-[\d+]+-implementer`)},
	{ID: "gated-reviewer", Label: "gated reviewer", Key: regexp.MustCompile(`^tasks?-[\d+]+-reviewer`)},
	{ID: "panel-fixer", Label: "panel fixer", Key: regexp.MustCompile(`^panel-fix-`)},
	{ID: "merger", Label: "merger", Key: regexp.MustCompile(`^panel-\d+-merger`)},
	{ID: "minor-fixer", Label: "minor fixer", Key: regexp.MustCompile(`^panel-\d+-minor-fix$`)},
	{ID: "panel-reviewer", Label: "panel reviewer", Key: regexp.MustCompile(`^panel-\d+-`)},
	{ID: "runner", Label: "runner", Key: regexp.MustCompile(`^runner-\d+$`)},
	{ID: "full-suite-fixer", Label: "full-suite fixer", Key: regexp.MustCompile(`^full-suite-fix-\d+$`)},
	{ID: "planner", Label: "planner", Key: regexp.MustCompile(`^(plan|design)-\d+$`)},
	{ID: "planner-short", Label: "short planner", Key: regexp.MustCompile(`^plan-short-\d+$`)},
	{ID: "fix-implementer", Label: "fix implementer", Key: regexp.MustCompile(`^fix-\d+-implementer$`)},
	{ID: "fix-reviewer", Label: "fix reviewer", Key: regexp.MustCompile(`^fix-\d+-review$`)},
	{ID: "explorer", Label: "explorer", Key: regexp.MustCompile(`^explore-\d+$`)},
	{ID: "visual-verifier", Label: "visual verifier", Key: regexp.MustCompile(`^visual-verify`)},
	{ID: "verify-list", Label: "verify list", Key: regexp.MustCompile(`^verify`)},
	{ID: "self-review-planner", Label: "self-review planner", Key: regexp.MustCompile(`^self-review-.+-plan$`)},
	{ID: "self-review-reviewer", Label: "self-review reviewer", Key: regexp.MustCompile(`^self-review-.+-(pass|review-r?\d+)$`)},
	{ID: "self-review-fixer", Label: "self-review fixer", Key: regexp.MustCompile(`^self-review-.+-fix(-[gvr]?\d+)?$`)},
	{ID: "pipeline-fix-planner", Label: "pipeline-fix planner", Key: regexp.MustCompile(`^pipeline-fix-\d+-plan(-\d+)?$`)},
	{ID: "pipeline-fix-fixer", Label: "pipeline-fix fixer", Key: regexp.MustCompile(`^pipeline-fix-\d+-fix(-[gvr]?\d+)?$`)},
	{ID: "pipeline-fix-reviewer", Label: "pipeline-fix reviewer", Key: regexp.MustCompile(`^pipeline-fix-\d+-review(-r?\d+)?$`)},
	{ID: "operator-request", Label: "operator request", Key: regexp.MustCompile(`^operator-\d+$`)},
}

// Category classes one dispatch by the table: opener is the **Opener:** of
// the key's first task (FirstTaskID), empty when it has none. A key no row
// matches is "other:<role>".
func Category(key, role, opener string) (id, label string) {
	for _, row := range Categories {
		if row.Key.MatchString(key) && strings.HasPrefix(opener, row.Opener) {
			return row.ID, row.Label
		}
	}
	return "other:" + role, "other: " + role
}

var taskIDs = regexp.MustCompile(`^tasks?-([\d+]+)-`)

// TaskIDs is every task id a task key names -- 13, 14 for
// task-13+14-implementer -- or nil for a key naming no task.
func TaskIDs(key string) []string {
	if m := taskIDs.FindStringSubmatch(key); m != nil {
		return strings.Split(m[1], "+")
	}
	return nil
}

// FirstTaskID is the first of TaskIDs, or "" for a key naming no task.
func FirstTaskID(key string) string {
	if ids := TaskIDs(key); ids != nil {
		return ids[0]
	}
	return ""
}

var (
	taskLine   = regexp.MustCompile(`^- \[[ xX]\] (\d+)\. `)
	openerLine = regexp.MustCompile("^\\*\\*Opener:\\*\\*\\s*`?([^`\\s]+)`?")
)

// TaskOpeners maps each task id of a tasks.md to the value of its column-0
// **Opener:** line; a task with none is absent. Fenced lines are skipped.
func TaskOpeners(tasksMD []byte) map[string]string {
	out := map[string]string{}
	task, fenced := "", false
	sc := bufio.NewScanner(bytes.NewReader(tasksMD))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "```") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		if m := taskLine.FindStringSubmatch(line); m != nil {
			task = m[1]
		} else if m := openerLine.FindStringSubmatch(line); m != nil && task != "" {
			out[task] = m[1]
		}
	}
	return out
}
