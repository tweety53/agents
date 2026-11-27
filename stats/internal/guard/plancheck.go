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
	"time"
)

// flow-guard plan-check <worktree> <changeRoot> <out-decision.json> [-decide-only]
// — review an approved plan and decide its implementer groups, review panel
// bundles and visual verdict with one headless, read-only model call.
//
// The call is pjRun on the decider role's first base pair, run in <worktree>
// with tools Read, Grep and Glob under a 600 s deadline (Env.PairJudgeTimeout
// overrides it). Its system prompt is skills/flow/plan-check-prompt.md under
// FLOW_GUARD_REPO_ROOT followed by skills/flow/plan-check-decision.md, the
// decision rules both modes share; its stdin is whether <worktree>/.flow/project.md
// declares `## visual verification`, skills/flow/plan-review-rubric.md, then proposal.md,
// design.md and tasks.md from <changeRoot> (`none` for an absent file), cut at the digest
// cap so only the artifacts' tail is lost.
//
// The answer is the report, a line reading exactly `===DECISION===`, then the
// decision JSON. The report is written to <worktree>/.superpowers/sdd/plan-check.md
// and must end in `PLAN-CHECK-CLEAN` or `PLAN-CHECK-FINDINGS <n>`; the decision
// is checked by dgpvInvalid and
// written to <out-decision.json> in the same schema. -decide-only asks
// skills/flow/plan-check-decide-prompt.md, followed by the same rules, instead — no tools, a 120 s deadline,
// no rubric, an answer of the `===DECISION===` line and the JSON alone — writes
// the decision alone and leaves plan-check.md untouched.
//
// The full mode also derives <worktree>/.superpowers/sdd/confidence.json from
// the report (KAN-1130): `{"score": <n>, "gaps": [...]}`, the score of the
// `CONFIDENCE: <n>%` line (1-100) under the report's `### Execution
// confidence` heading and every non-blank line after it up to the next
// heading or the report's last line, verbatim. A report with no such line
// removes a confidence.json an earlier check left. -decide-only neither
// writes nor removes it.
//
// One plan-check makes at most two model calls (pcAsks): an answer owed a retry — a failed call, an
// unusable decision, a report with no terminal line — is asked once more with its reason appended to
// the system prompt, and a usable first decision is kept over an unusable retry.
//
// A report whose last line is no PLAN-CHECK-* line is not accepted at first: the judge is asked once
// more with the terminal line named (pcNoTerminal); a report still lacking it is closed from its body
// (pcDerive: `PLAN-CHECK-FINDINGS <n>` from its finding lines, `PLAN-CHECK-CLEAN` from a confidence
// score alone).
//
// A report holding neither a terminal line nor anything to derive one from falls back the report
// alone, to `PLAN-CHECK-NOT-RUN` and the words `no confidence estimate exists`, stdout reading
// `PLAN-CHECK-NOT-RUN — no confidence estimate`. The decision has no fallback: one failing its
// checks, a failed call (no decider pair, the command missing or failing, the deadline) or a missing
// separator is retried within those two calls and then BLOCKS — exit 1, one stderr line naming the cause, reason
// decision-unusable, nothing written — since a run on a made-up decision is an unplanned run.
// Otherwise stdout is one line, `plan-check (judge): <report> · <n>
// groups · panel <a+b> … · visual <verify>`, and in the full mode
// ` · confidence <n>%` or ` · confidence unknown` after it.
//
// Exit 0 on a written result. Exit 1 on no usable decision. Exit 2, nothing on
// stdout: usage (reason usage), a tasks.md that is unreadable or holds no
// task (plan-unreadable), or a decision, report or confidence file that cannot be written
// (cannot-write-result).
//
// Go, in flow-guard (stats/internal/guard/plancheck.go).
func init() { Registry["plan-check"] = planCheck }

const (
	pcLimit       = 600 * time.Second
	pcDecideLimit = 120 * time.Second
)

var (
	pcSeparator  = regexp.MustCompile(`(?m)^===DECISION===[ \t]*$`)
	pcConfidence = regexp.MustCompile(`(?m)^CONFIDENCE: ([0-9]{1,3})%[ \t]*$`)
	pcLastLine   = regexp.MustCompile(`^(PLAN-CHECK-CLEAN|PLAN-CHECK-FINDINGS [0-9]+)$`)
	pcNotRunText = "plan check not run: the model answer could not be used — no confidence estimate exists\nPLAN-CHECK-NOT-RUN\n"
	pcFinding    = regexp.MustCompile(`(?m)^\*\*F[0-9]+\*\* \[(?:finding|blocker)\]`)
	pcAnyFinding = regexp.MustCompile(`(?m)^\*\*F[0-9]+\*\* \[(?:finding|advisory|blocker)\]`)
)

// pcNoTerminal is the retry reason of a report with no terminal line.
const pcNoTerminal = "its report did not end with a terminal line. The report's last line, before the ===DECISION=== line, must be exactly PLAN-CHECK-CLEAN or PLAN-CHECK-FINDINGS <n>, <n> the number of finding lines above it."

// pcAsks bounds the model calls one plan-check makes: the first, and one retry naming why the
// first could not be used.
const pcAsks = 2

// pcHasTerminal is whether the answer's report part (all of it when no ===DECISION=== line is there)
// ends in a PLAN-CHECK-* line.
func pcHasTerminal(result string) bool {
	r := result
	if loc := pcSeparator.FindStringIndex(result); loc != nil {
		r = result[:loc[0]]
	}
	r = strings.TrimSpace(r)
	return pcLastLine.MatchString(r[strings.LastIndex(r, "\n")+1:])
}

// pcDerive closes a report from its body: the blocker and finding lines it holds make
// `PLAN-CHECK-FINDINGS <n>`; with none, an advisory line or a CONFIDENCE line makes
// `PLAN-CHECK-CLEAN` — an advisory is never owed a fix, so it never sends the plan to triage;
// "" for a body holding neither (nothing to derive from).
func pcDerive(body string) string {
	if n := len(pcFinding.FindAllString(body, -1)); n > 0 {
		return fmt.Sprintf("PLAN-CHECK-FINDINGS %d", n)
	}
	if score, _ := pcConfidenceOf(body); score > 0 || pcAnyFinding.MatchString(body) {
		return "PLAN-CHECK-CLEAN"
	}
	return ""
}

func planCheck(args []string, env Env, stdout, stderr io.Writer) int {
	decideOnly := len(args) == 4 && args[3] == "-decide-only"
	if len(args) != 3 && !decideOnly {
		fmt.Fprintln(stderr, "usage: flow-guard plan-check <worktree> <changeRoot> <out-decision.json> [-decide-only]")
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
		fmt.Fprintf(stderr, "plan-check: %v\n", err)
		Reason(env, "plan-unreadable")
		return 2
	}
	project, _ := os.ReadFile(filepath.Join(wt, ".flow", "project.md"))
	visual := dgpvVisualHeading.Match(project)

	var report, source, spend, cause, note string
	var d *dgpvDecision
	repoAt := pdgRepositories(env, planPath, lines)
	for range pcAsks {
		r, dd, src, sp, c, retry := pcAsk(env, wt, root, lines, tasks, repoAt, visual, decideOnly, note)
		spend += sp
		if dd != nil || d == nil { // a usable first answer is never traded for an unusable retry
			report, d, source, cause = r, dd, src, c
		}
		if retry == "" {
			break
		}
		if len(retry) > 2000 {
			retry = retry[:2000]
		}
		note = "\n\nYour previous answer could not be used: " + retry + "\n"
	}
	if d == nil {
		fmt.Fprintf(stderr, "plan-check: no usable decision, the run is blocked — %s\n", strings.TrimSpace(pjSpace.ReplaceAllString(cause, " ")))
		Reason(env, "decision-unusable")
		return 1
	}
	if cause != "" {
		fmt.Fprintf(stderr, "plan-check: judge answer unusable, using the fallback report — %s\n", strings.TrimSpace(pjSpace.ReplaceAllString(cause, " ")))
	}
	spAddSecurity(d, filepath.Join(wt, ".flow", "project.md"), tasks)
	d.Overrides = []json.RawMessage{}
	b, _ := json.MarshalIndent(d, "", "  ")
	write := func(path string, content []byte) error {
		err := os.MkdirAll(filepath.Dir(path), 0o755)
		if err == nil {
			err = os.WriteFile(path, content, 0o644)
		}
		return err
	}
	last, confidence := "unchanged", ""
	if !decideOnly {
		if report == "" {
			report = pcNotRunText
		}
		if err := write(filepath.Join(wt, ".superpowers", "sdd", "plan-check.md"), []byte(report)); err != nil {
			fmt.Fprintf(stderr, "plan-check: %v\n", err)
			Reason(env, "cannot-write-result")
			return 2
		}
		last = strings.TrimSpace(report[strings.LastIndex(strings.TrimSpace(report), "\n")+1:])
		if last == "PLAN-CHECK-NOT-RUN" {
			last += " — no confidence estimate"
		}
		confPath := filepath.Join(wt, ".superpowers", "sdd", "confidence.json")
		var err error
		if score, gaps := pcConfidenceOf(report); score > 0 {
			confidence = fmt.Sprintf(" · confidence %d%%", score)
			cb, _ := json.MarshalIndent(map[string]any{"score": score, "gaps": gaps}, "", "  ")
			err = write(confPath, append(cb, '\n'))
		} else {
			confidence = " · confidence unknown"
			if report == pcNotRunText {
				confidence = ""
			}
			if err = os.Remove(confPath); os.IsNotExist(err) {
				err = nil
			}
		}
		if err != nil {
			fmt.Fprintf(stderr, "plan-check: %v\n", err)
			Reason(env, "cannot-write-result")
			return 2
		}
	}
	if err := write(out, append(b, '\n')); err != nil {
		fmt.Fprintf(stderr, "plan-check: %v\n", err)
		Reason(env, "cannot-write-result")
		return 2
	}
	var bundles []string
	for _, p := range d.Panel {
		bundles = append(bundles, strings.Join(p.Slots, "+"))
	}
	fmt.Fprintf(stdout, "plan-check (%s): %s · %d groups · panel %s · visual %s%s%s\n", source, last, len(d.Groups), strings.Join(bundles, " "), d.Visual.Verify, confidence, spend)
	return 0
}

// pcConfidenceOf reads the report's `### Execution confidence` section: the
// score of its `CONFIDENCE: <n>%` line, 1-100, and every non-blank line after
// that one up to the next heading or the report's last line, verbatim. A
// score of 0 means the section or a well-formed line is missing.
func pcConfidenceOf(report string) (score int, gaps []string) {
	gaps = []string{}
	_, section, found := strings.Cut(report, "### Execution confidence")
	if !found {
		return 0, gaps
	}
	m := pcConfidence.FindStringSubmatchIndex(section)
	if m == nil {
		return 0, gaps
	}
	score, _ = strconv.Atoi(section[m[2]:m[3]])
	if score < 1 || score > 100 {
		return 0, gaps
	}
	for _, l := range strings.Split(section[m[1]:], "\n") {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "#") || pcLastLine.MatchString(l) {
			break
		}
		if ids, ok := strings.CutPrefix(l, "Lowered by:"); ok {
			// F<k> alone means nothing outside the report: spell each cited finding out.
			for _, id := range strings.Split(ids, ",") {
				id = strings.TrimSpace(id)
				if id != "" {
					gaps = append(gaps, "Lowered by "+id+pcFindingText(report, id))
				}
			}
		} else if l != "" {
			gaps = append(gaps, l)
		}
	}
	return score, gaps
}

// pcFindingText is " — <where> — <what>" of the report's `**<id>** [<kind>] <where> — <what>`
// line, "" when the report has none.
func pcFindingText(report, id string) string {
	for _, l := range strings.Split(report, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(l), "**"+id+"** "); ok {
			if _, text, ok := strings.Cut(rest, "] "); ok {
				return " — " + text
			}
		}
	}
	return ""
}

// pcAsk runs the check once, note appended to the system prompt; a nil decision or empty report
// is the part to fall back. retry is why the answer is owed one more model call ("" none): a
// failed run, an unusable decision, or a report with no terminal line.
func pcAsk(env Env, wt, root string, lines []string, tasks []pdbTask, repoAt []string, visual, decideOnly bool, note string) (report string, d *dgpvDecision, source, spend, cause, retry string) {
	source = "fallback"
	repo := env.Getenv("FLOW_GUARD_REPO_ROOT")
	promptFile, tools, limit := "plan-check-prompt.md", "Read,Grep,Glob", pcLimit
	if decideOnly {
		promptFile, tools, limit = "plan-check-decide-prompt.md", "", pcDecideLimit
	}
	prompt, err := os.ReadFile(filepath.Join(repo, "skills", "flow", promptFile))
	if err != nil {
		return "", nil, source, "", "plan-check prompt unreadable: " + err.Error(), ""
	}
	rules, err := os.ReadFile(filepath.Join(repo, "skills", "flow", "plan-check-decision.md"))
	if err != nil {
		return "", nil, source, "", "plan-check prompt unreadable: " + err.Error(), ""
	}
	prompt = append(append(prompt, '\n'), rules...)
	rows, err := policyTable(env)
	if err != nil {
		return "", nil, source, "", err.Error(), ""
	}
	read := func(p string) string {
		if b, err := os.ReadFile(p); err == nil {
			return string(b)
		}
		return "none"
	}
	declares := "no"
	if visual {
		declares = "yes"
	}
	rubric := read(filepath.Join(repo, "skills", "flow", "plan-review-rubric.md"))
	if decideOnly {
		rubric = "not needed"
	}
	// The rubric and the visual declaration lead, so the cap cuts only the artifacts' tail.
	digest := []byte("## The project declares `## visual verification`: " + declares + "\n\n## plan-review-rubric.md\n\n" + rubric +
		"\n\n## proposal.md\n\n" + read(filepath.Join(root, "proposal.md")) + "\n\n## design.md\n\n" + read(filepath.Join(root, "design.md")) +
		"\n\n## tasks.md\n\n" + strings.Join(lines, "\n") + "\n")
	if len(digest) > pjDigestCap {
		digest = append(digest[:pjDigestCap:pjDigestCap], pjTruncated...)
	}
	env.Dir = wt
	result, spend, cause := pjRun(env, rows, "decider", string(prompt)+note, digest, tools, limit)
	if cause != "" {
		return "", nil, source, spend, cause, cause
	}
	if !decideOnly && !pcHasTerminal(result) {
		retry = pcNoTerminal
	}
	report, d, source, cause = pcParse(result, tasks, repoAt, visual, decideOnly, spend)
	if d == nil {
		retry = cause
	}
	return report, d, source, spend, cause, retry
}

// pcParse splits one answer into its report and its checked decision.
func pcParse(result string, tasks []pdbTask, repoAt []string, visual, decideOnly bool, spend string) (report string, d *dgpvDecision, source, cause string) {
	source = "fallback"
	loc := pcSeparator.FindStringIndex(result)
	if loc == nil {
		return "", nil, source, "plan-check answer has no ===DECISION=== line: " + result + spend
	}
	if r := strings.TrimSpace(result[:loc[0]]); decideOnly {
		// the decision-only call writes no report
	} else if pcLastLine.MatchString(r[strings.LastIndex(r, "\n")+1:]) {
		report = r + "\n"
		if body := r[:max(strings.LastIndex(r, "\n"), 0)]; pcAnyFinding.MatchString(body) {
			// The answer's own count is re-derived from its lines, so an advisory never counts.
			report = body + "\n" + pcDerive(body) + "\n"
		}
	} else if last := pcDerive(r); last != "" {
		report = r + "\n" + last + "\n"
	} else {
		cause = "plan-check report does not end in PLAN-CHECK-CLEAN or PLAN-CHECK-FINDINGS <n>" + spend
	}
	source = "judge"
	text := result[loc[1]:]
	var got dgpvDecision
	i, j := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if i < 0 || j < i || json.Unmarshal([]byte(text[i:j+1]), &got) != nil {
		return report, nil, source, "plan-check decision is not a JSON object: " + text + spend
	}
	if why := dgpvInvalid(got, tasks, repoAt, visual); why != "" {
		return report, nil, source, "plan-check decision rejected: " + why + spend
	}
	return report, &got, source, cause
}

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

// dgpvSlots is the slot catalog, in the order of the table skills/flow/plan-check-decision.md shows
// the model (TestDgpvCatalogMatchesTheDecisionTable holds the two equal).
var dgpvSlots = []string{"primary", "principles", "test-sanity", "call-efficiency", "failure-modes", "mutation"}

// dgpvInvalid is why d breaks a rule the header names, "" when none.
func dgpvInvalid(d dgpvDecision, tasks []pdbTask, repoAt []string, visual bool) string {
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
			case !slices.Contains(dgpvSlots, s):
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
