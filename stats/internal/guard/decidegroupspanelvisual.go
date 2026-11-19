package guard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// flow-guard decide-groups-panel-visual <worktree> <changeRoot> <out-decision.json>
// — decide an approved plan's implementer groups, review panel bundles and
// visual verdict with one headless model call.
//
// The call is pjRun on the decider role's first base pair read from `flow
// policy get`. Its system prompt is skills/flow/decider-prompt.md under
// FLOW_GUARD_REPO_ROOT, canonical for the rules and the answer's shape; its
// stdin is <changeRoot>/tasks.md and <changeRoot>/design.md in full (`none`
// for an absent design.md) and whether <worktree>/.flow/project.md declares
// `## visual verification`, cut at 200 KB as pair-judge cuts its digest.
//
// The answer is the one JSON object in the result text, and is valid when
// every plan task sits in exactly one group; a `**Size:** big` task is a
// group alone and any other group holds at most three; no group holds tasks
// of two `## <repository>` headings (pdgRepositories); every task a task's
// **After:** names sits in its group or an earlier one; every panel bundle is
// one or two distinct catalog slots, `mutation` alone, no slot twice and
// `primary` among them; and visual.verify is `required`, `skipped` or `not
// configured` -- `not configured` exactly when the project declares no
// `## visual verification`.
//
// Writes {groups, panel, visual, overrides: []} to <out-decision.json> and
// prints one line, `decided (judge|fallback): <n> groups · panel <a+b> <c> …
// · visual <verify>`.
//
// The fallback, on any failure -- an unreadable decider-prompt.md, no decider
// pair, the command missing or failing, the deadline, an answer that is not
// JSON, or one the checks above reject -- is one stderr line naming the
// cause and the decision: one group per task in plan order; panel
// primary+principles, test-sanity+call-efficiency, failure-modes, mutation;
// visual `required` when the project declares `## visual verification`, else
// `not configured`; every reason `fallback`.
//
// Exit 0 on a written decision, whatever the model did. Exit 2, nothing on
// stdout: usage (reason usage), a tasks.md that is unreadable or holds no
// task (plan-unreadable), or a decision file that cannot be written
// (cannot-write-decision).
//
// Go, in flow-guard (stats/internal/guard/decidegroupspanelvisual.go).
func init() { Registry["decide-groups-panel-visual"] = decideGroupsPanelVisual }

var dgpvVisualHeading = regexp.MustCompile(`(?m)^## visual verification[ \t]*$`)

type dgpvEntry struct {
	Tasks  []string `json:"tasks,omitempty"`
	Slots  []string `json:"slots,omitempty"`
	Reason string   `json:"reason"`
}

type dgpvDecision struct {
	Groups []dgpvEntry `json:"groups"`
	Panel  []dgpvEntry `json:"panel"`
	Visual struct {
		Verify string `json:"verify"`
		Reason string `json:"reason"`
	} `json:"visual"`
	Overrides []json.RawMessage `json:"overrides"`
}

func decideGroupsPanelVisual(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 {
		fmt.Fprintln(stderr, "usage: flow-guard decide-groups-panel-visual <worktree> <changeRoot> <out-decision.json>")
		Reason(env, "usage")
		return 2
	}
	wt, root, out := args[0], args[1], args[2]
	planPath := filepath.Join(root, "tasks.md")
	lines, err := tcfReadPlan(env, planPath)
	tasks := pdbParseTasks(lines)
	if err == nil && len(tasks) == 0 {
		err = fmt.Errorf("%s holds no task", planPath)
	}
	if err != nil {
		fmt.Fprintf(stderr, "decide-groups-panel-visual: %v\n", err)
		Reason(env, "plan-unreadable")
		return 2
	}
	project, _ := os.ReadFile(filepath.Join(wt, ".flow", "project.md"))
	visual := dgpvVisualHeading.Match(project)

	d, source, spend, cause := dgpvAsk(env, root, lines, tasks, pdgRepositories(env, planPath, lines), visual)
	if cause != "" {
		fmt.Fprintf(stderr, "decide-groups-panel-visual: fallback — %s\n", strings.TrimSpace(pjSpace.ReplaceAllString(cause, " ")))
		d = dgpvFallback(tasks, visual)
	}
	d.Overrides = []json.RawMessage{}
	b, _ := json.MarshalIndent(d, "", "  ")
	if err := os.WriteFile(out, append(b, '\n'), 0o644); err != nil {
		fmt.Fprintf(stderr, "decide-groups-panel-visual: %v\n", err)
		Reason(env, "cannot-write-decision")
		return 2
	}
	var bundles []string
	for _, p := range d.Panel {
		bundles = append(bundles, strings.Join(p.Slots, "+"))
	}
	fmt.Fprintf(stdout, "decided (%s): %d groups · panel %s · visual %s%s\n", source, len(d.Groups), strings.Join(bundles, " "), d.Visual.Verify, spend)
	return 0
}

// dgpvAsk runs the decider and checks its answer; a non-empty cause is why
// it cannot be used.
func dgpvAsk(env Env, root string, lines []string, tasks []pdbTask, repoAt []string, visual bool) (d dgpvDecision, source, spend, cause string) {
	source = "fallback"
	prompt, err := os.ReadFile(filepath.Join(env.Getenv("FLOW_GUARD_REPO_ROOT"), "skills", "flow", "decider-prompt.md"))
	if err != nil {
		return d, source, "", "decider prompt unreadable: " + err.Error()
	}
	rows, err := policyTable(env)
	if err != nil {
		return d, source, "", err.Error()
	}
	design, err := os.ReadFile(filepath.Join(root, "design.md"))
	if err != nil {
		design = []byte("none")
	}
	declares := "no"
	if visual {
		declares = "yes"
	}
	digest := []byte("## tasks.md\n\n" + strings.Join(lines, "\n") + "\n\n## design.md\n\n" + string(design) +
		"\n\n## The project declares `## visual verification`: " + declares + "\n")
	if len(digest) > pjDigestCap {
		digest = append(digest[:pjDigestCap:pjDigestCap], pjTruncated...)
	}
	result, spend, cause := pjRun(env, rows, "decider", string(prompt), digest)
	if cause != "" {
		return d, source, spend, cause
	}
	i, j := strings.Index(result, "{"), strings.LastIndex(result, "}")
	if i < 0 || j < i || json.Unmarshal([]byte(result[i:j+1]), &d) != nil {
		return d, source, spend, "decider answer is not a JSON object: " + result + spend
	}
	catalog := map[string]bool{}
	for _, m := range fsjCatalog.FindAllStringSubmatch(string(prompt), -1) {
		catalog[m[1]] = true
	}
	if why := dgpvInvalid(d, tasks, repoAt, catalog, visual); why != "" {
		return d, source, spend, "decider answer rejected: " + why + spend
	}
	return d, "judge", spend, ""
}

// dgpvInvalid is why d breaks a rule the header names, "" when none.
func dgpvInvalid(d dgpvDecision, tasks []pdbTask, repoAt []string, catalog map[string]bool, visual bool) string {
	byID := map[string]pdbTask{}
	for _, t := range tasks {
		byID[t.id] = t
	}
	groupOf := map[string]int{}
	for gi, g := range d.Groups {
		if len(g.Tasks) == 0 {
			return fmt.Sprintf("group %d names no task", gi+1)
		}
		for _, id := range g.Tasks {
			t, ok := byID[id]
			if !ok {
				return fmt.Sprintf("group %d names task %s, which the plan lacks", gi+1, id)
			}
			if _, dup := groupOf[id]; dup {
				return fmt.Sprintf("task %s is in two groups", id)
			}
			groupOf[id] = gi
			if t.size == "big" && len(g.Tasks) > 1 {
				return fmt.Sprintf("big task %s shares group %d", id, gi+1)
			}
			if repoAt[t.taskLine] != repoAt[byID[g.Tasks[0]].taskLine] {
				return fmt.Sprintf("group %d spans two repositories", gi+1)
			}
		}
		if len(g.Tasks) > 3 {
			return fmt.Sprintf("group %d holds %d tasks", gi+1, len(g.Tasks))
		}
	}
	for _, t := range tasks {
		gi, ok := groupOf[t.id]
		if !ok {
			return fmt.Sprintf("task %s is in no group", t.id)
		}
		for _, a := range t.afterSet {
			if ag, ok := groupOf[a]; ok && ag > gi {
				return fmt.Sprintf("task %s comes before its After task %s", t.id, a)
			}
		}
	}
	var seen []string
	for bi, b := range d.Panel {
		if len(b.Slots) < 1 || len(b.Slots) > 2 {
			return fmt.Sprintf("bundle %d holds %d slots", bi+1, len(b.Slots))
		}
		for _, s := range b.Slots {
			switch {
			case !catalog[s]:
				return fmt.Sprintf("bundle %d names %s, which the catalog lacks", bi+1, s)
			case slices.Contains(seen, s):
				return fmt.Sprintf("slot %s is in two bundles", s)
			case s == "mutation" && len(b.Slots) > 1:
				return "mutation shares a bundle"
			}
			seen = append(seen, s)
		}
	}
	if !slices.Contains(seen, "primary") {
		return "no bundle carries primary"
	}
	if v := d.Visual.Verify; v != "required" && v != "skipped" && v != "not configured" {
		return fmt.Sprintf("visual.verify %q is not required, skipped or not configured", v)
	} else if (v == "not configured") == visual {
		return fmt.Sprintf("visual.verify %q disagrees with the project's `## visual verification`", v)
	}
	return ""
}

// dgpvFallback is the decision taken when the decider's answer cannot be used.
func dgpvFallback(tasks []pdbTask, visual bool) dgpvDecision {
	var d dgpvDecision
	for _, t := range tasks {
		d.Groups = append(d.Groups, dgpvEntry{Tasks: []string{t.id}, Reason: "fallback"})
	}
	for _, s := range [][]string{{"primary", "principles"}, {"test-sanity", "call-efficiency"}, {"failure-modes"}, {"mutation"}} {
		d.Panel = append(d.Panel, dgpvEntry{Slots: s, Reason: "fallback"})
	}
	d.Visual.Verify, d.Visual.Reason = "not configured", "fallback"
	if visual {
		d.Visual.Verify = "required"
	}
	return d
}
