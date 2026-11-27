package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// flow-guard check-panel-recorded <worktree> <round> — a round's panel reports
// count more findings (the sum of their Critical/Important/Minor header lines)
// than the change's store holds panel-slot rows for that round: the parent never ran
// `flow record finding`, so the rendered record would say `findings-total: 0`
// (kan-1158). Exit 0 recorded or nothing raised, 1 unrecorded (the report files
// named), 2 usage. The change is the worktree's directory name; rows of the
// `gated-reviewer` slot are task reviews, not the panel's.
func init() { Registry["check-panel-recorded"] = checkPanelRecorded }

var cprCount = regexp.MustCompile(`(?m)^(?:Critical|Important|Minor): (\d+)`)

func checkPanelRecorded(args []string, env Env, stdout, stderr io.Writer) int {
	round, err := 0, error(nil)
	if len(args) == 2 {
		round, err = strconv.Atoi(args[1])
	}
	if len(args) != 2 || err != nil {
		Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard check-panel-recorded <worktree> <round>")
		return 2
	}
	worktree := args[0]
	reports, _ := filepath.Glob(filepath.Join(worktree, ".superpowers/sdd", fmt.Sprintf("panel-report-%d-*.md", round)))
	var raised []string
	want := 0
	for _, r := range reports {
		body, _ := os.ReadFile(r)
		n := 0
		for _, m := range cprCount.FindAllSubmatch(body, -1) {
			k, _ := strconv.Atoi(string(m[1]))
			n += k
		}
		if n > 0 {
			raised = append(raised, r)
			want += n
		}
	}
	if want == 0 {
		return 0
	}
	out, _, rc := pcFlow(env, false, "record", "findings", "-change", filepath.Base(worktree), "-C", worktree)
	var rows []struct {
		Round int    `json:"round"`
		Slot  string `json:"slot"`
	}
	if rc != 0 || json.Unmarshal(out, &rows) != nil {
		Reason(env, "store-unreadable")
		fmt.Fprintf(stderr, "check-panel-recorded: flow record findings exit %d\n", rc)
		return 1
	}
	got := 0
	for _, r := range rows {
		if r.Round == round && r.Slot != "gated-reviewer" {
			got++
		}
	}
	if got >= want {
		return 0
	}
	Reason(env, "panel-unrecorded")
	fmt.Fprintf(stderr, "check-panel-recorded: round %d reports raise %d findings but the store holds %d panel rows for it — record them (flow record finding -from findings-%d.jsonl): %v\n", round, want, got, round, raised)
	return 1
}
