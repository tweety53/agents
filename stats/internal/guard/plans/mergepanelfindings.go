package plans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// flow-guard merge-panel-findings <abs-worktree> <round> <id>=<dispatch-seq>... —
// the findings merger of a panel round (**The findings merger**,
// skills/flow/review-panel.md). Each named dispatch role wrote
// .superpowers/sdd/findings-<round>-<id>.jsonl beside its report: one JSON
// object per finding carrying task, severity, location and note. The guard
// concatenates them, in argument order, into findings-<round>.jsonl, each line
// given round, slot (the panel role the id opens with, so an `-unread` or
// `-<repo>` suffix is dropped), dispatch-seq and
// status open — the file `flow record finding -from` records in one call — and
// prints one `finding:` line per finding and one `duplicate:` line per
// location cited by more than one role. It judges nothing. Exit 0 written,
// 1 a role's file missing or not JSON objects (named), 2 usage.
func init() { core.Registry["merge-panel-findings"] = mergePanelFindings }

func mergePanelFindings(args []string, env core.Env, stdout, stderr io.Writer) int {
	round, err := 0, error(nil)
	if len(args) >= 3 {
		round, err = strconv.Atoi(args[1])
	}
	if len(args) < 3 || err != nil {
		core.Reason(env, "usage")
		fmt.Fprintln(stderr, "usage: flow-guard merge-panel-findings <abs-worktree> <round> <id>=<dispatch-seq>...")
		return 2
	}
	sdd := filepath.Join(args[0], ".superpowers/sdd")
	var out bytes.Buffer
	var lines []string
	byLoc := map[string][]string{}
	var locs []string
	for _, a := range args[2:] {
		id, seq, ok := strings.Cut(a, "=")
		n, err := strconv.Atoi(seq)
		if !ok || id == "" || err != nil {
			core.Reason(env, "usage")
			fmt.Fprintf(stderr, "merge-panel-findings: %q is not <id>=<dispatch-seq>\n", a)
			return 2
		}
		path := filepath.Join(sdd, fmt.Sprintf("findings-%d-%s.jsonl", round, id))
		body, err := os.ReadFile(path)
		if err != nil {
			core.Reason(env, "findings-file-missing")
			fmt.Fprintf(stderr, "merge-panel-findings: %v — that role is not run\n", err)
			return 1
		}
		slot := panelSlotOf(id)
		k := 0
		for i, line := range strings.Split(string(body), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			var row map[string]any
			if json.Unmarshal([]byte(line), &row) != nil || row == nil {
				core.Reason(env, "findings-file-invalid")
				fmt.Fprintf(stderr, "merge-panel-findings: %s:%d is not a JSON object\n", path, i+1)
				return 1
			}
			k++
			row["round"], row["slot"], row["dispatch-seq"] = round, slot, n
			if row["status"] == nil {
				row["status"] = "open"
			}
			enc, _ := json.Marshal(row)
			out.Write(append(enc, '\n'))
			loc := fmt.Sprint(row["location"])
			task := "none"
			if row["task"] != nil {
				task = fmt.Sprint(row["task"])
			}
			ref := fmt.Sprintf("%s#%d", slot, k)
			lines = append(lines, fmt.Sprintf("finding: %s — %v — %s — %s — %v", ref, row["severity"], loc, task, row["note"]))
			if byLoc[loc] == nil {
				locs = append(locs, loc)
			}
			byLoc[loc] = append(byLoc[loc], ref)
		}
	}
	dst := filepath.Join(sdd, fmt.Sprintf("findings-%d.jsonl", round))
	if err := os.WriteFile(dst, out.Bytes(), 0o644); err != nil {
		core.Reason(env, "write-failed")
		fmt.Fprintf(stderr, "merge-panel-findings: %v\n", err)
		return 1
	}
	for _, l := range lines {
		fmt.Fprintln(stdout, l)
	}
	for _, loc := range locs {
		refs := byLoc[loc]
		roles := map[string]bool{}
		for _, r := range refs {
			roles[r[:strings.LastIndex(r, "#")]] = true
		}
		if len(roles) > 1 {
			fmt.Fprintf(stdout, "duplicate: %s — %s\n", loc, strings.Join(refs, ", "))
		}
	}
	fmt.Fprintln(stdout, dst)
	return 0
}

// panelSlots are the panel's role ids (**The panel**, skills/flow/review-panel.md).
var panelSlots = []string{"primary", "principles", "test-sanity", "performance", "failure-modes", "security", "mutation"}

// panelSlotOf is the panel role a findings file's id names: the role it equals
// or opens with followed by a dash; any other id is its own slot.
func panelSlotOf(id string) string {
	for _, s := range panelSlots {
		if id == s || strings.HasPrefix(id, s+"-") {
			return s
		}
	}
	return id
}
