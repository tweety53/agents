package guard

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// planStepFilesCmd is `flow-guard plan-step-files <tasks.md>`, run by
// check-plan-shape.sh after plan-task-id-rule on every tasks.md it scans
// (decision plan-step-files-subguard, kan-1072 self-review, F4). It refuses
// a path an edit step names in an edit position that its task's
// `**Files:**`/`**Allowed-collateral:**` does not cover — the commit guard
// would refuse that task's commit at close, after the work is done.
//
//   - A step is `  - [ ] **Step N: <label>**<rest>` plus its continuation
//     lines, fences and lines opening `+`, `|` or `>` excluded, joined into
//     one string. Only a step whose label opens with an edit verb
//     (psfLabelVerbs) is read, and only in a task that declares **Files:**.
//   - A candidate is a backticked token with no whitespace ending
//     `.<ext>`, an optional `:N` or `:N-M` stripped. With a `/` it must
//     exist under the repository root; without one it must name a file in
//     exactly one directory among the task's **Files:** directories, and
//     resolves to that path.
//   - It is in an edit position directly after an edit verb, at the head
//     of a list after `in:`, or in a list that opens a clause and is
//     followed by `:`, `'s`, ` (` or a change verb (psfEditPos).
//   - It passes when tcfCheckFiles — the commit guard's own cover test —
//     accepts it.
//
// The repository is the one whose `<spec-root>/changes/<name>/tasks.md`
// this is; an archived plan, or a file outside a `changes/<name>/`
// directory, is not checked. Exit 0 clean, 1 one `file:line: task <id> step
// <n> edits …` line per finding, 2 a usage error or a file that cannot be
// read.
//
// ponytail: a fixture path written into a test's scratch repository in an
// edit step reads as an edit (the one false positive measured over the
// archived plans, self-review-kan-922 task 2) — upgrade path, a step-local
// exclusion when a second plan hits it.
func init() { Registry["plan-step-files"] = planStepFilesCmd }

var (
	psfStepRE   = regexp.MustCompile(`^\s+- \[[ xX]\] \*\*Step (\d+[a-z]?):\s*([^*]*)\*\*(.*)`)
	psfTokRE    = regexp.MustCompile("`([^`\\s]+)`")
	psfExtRE    = regexp.MustCompile(`\.[A-Za-z][A-Za-z0-9]{0,5}(?::\d+(?:-\d+)?)?$`)
	psfLineRE   = regexp.MustCompile(`:\d+(?:-\d+)?$`)
	psfOpenRE   = regexp.MustCompile(`(?i)(?:^|^\s*-\s+|[.;]\s+|\*\*\s*)(?:(?:` + psfVerbs + `|in)\s+(?:the\s+)?)?$`)
	psfVerbRE   = regexp.MustCompile(`(?i)\b(?:` + psfVerbs + `)\s+(?:the\s+)?$`)
	psfListRE   = regexp.MustCompile(`^(?:\s*\([^)]*\))?(?:\s*,\s*(?:and\s+)?|\s+and\s+)$`)
	psfInListRE = regexp.MustCompile(`(?i)\bin:\s*$`)
	psfAfterRE  = regexp.MustCompile(`^\s*(?::|'s\b|\(|\s(?:become|gets?|drops?|reads?|loses?|names?|lists?|follows?|keeps?|takes?|stops?|gains?))`)
)

// psfVerbs are the verbs a path directly after is edited by.
const psfVerbs = `edit|delete|create|write|rewrite|update|remove|rename|move|reword|replace in|extend|port|change|fix|add to`

// psfLabelVerbs are the step labels whose step edits files.
var psfLabelVerbs = func() map[string]bool {
	m := map[string]bool{}
	for _, v := range strings.Fields("implement edit write rewrite update add fix change create delete remove rename move reword wire replace extend port make drop split restore") {
		m[v] = true
	}
	return m
}()

type psfStep struct {
	n, label, text string
	offset         int
}

func planStepFilesCmd(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] == "" {
		fmt.Fprintln(stderr, "usage: flow-guard plan-step-files <tasks.md>")
		return 2
	}
	file := args[0]
	if !filepath.IsAbs(file) {
		file = filepath.Join(env.Dir, file)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[0], err)
		return 2
	}
	lines, err := tcfPlanLines(b)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[0], err)
		return 2
	}
	changes := filepath.Dir(filepath.Dir(file))
	if filepath.Base(changes) != "changes" {
		return 0
	}
	root := filepath.Dir(filepath.Dir(changes))
	found := 0
	for _, body := range tcfIterTasks(lines) {
		task, ok := tcfParseTask(lines, body.id)
		if !ok || len(task.files) == 0 {
			continue
		}
		for _, s := range psfSteps(body.lines) {
			if !psfLabelVerbs[s.label] {
				continue
			}
			toks := psfTokRE.FindAllStringSubmatchIndex(s.text, -1)
			for k, m := range toks {
				tok := s.text[m[2]:m[3]]
				if !psfExtRE.MatchString(tok) {
					continue
				}
				p := psfResolve(root, psfLineRE.ReplaceAllString(tok, ""), task.files)
				if p == "" || len(tcfCheckFiles(task, []string{p})) == 0 || !psfEditPos(s.text, toks, k) {
					continue
				}
				fmt.Fprintf(stdout, "%s:%d: task %s step %s edits `%s`, which no **Files:** or **Allowed-collateral:** entry covers — add it to **Files:** when this task's commit changes it, else reword the step\n",
					args[0], body.taskLine+1+s.offset, task.id, s.n, p)
				found++
			}
		}
	}
	if found > 0 {
		return 1
	}
	return 0
}

// psfSteps is every step of a task body, fences excluded: the step line's
// text after its bold label, joined with each continuation line up to a
// line opening `**` or an unindented `- `; lines opening `+`, `|` or `>`
// are left out.
func psfSteps(body []string) []psfStep {
	var steps []psfStep
	cur := -1
	tcfUnfenced(body, func(i int, l string) bool {
		if m := psfStepRE.FindStringSubmatch(l); m != nil {
			label := ""
			if f := strings.Fields(strings.ToLower(m[2])); len(f) > 0 {
				label = strings.Trim(f[0], ".:")
			}
			steps = append(steps, psfStep{n: m[1], label: label, text: m[3], offset: i})
			cur = len(steps) - 1
			return false
		}
		if strings.HasPrefix(l, "**") || strings.HasPrefix(l, "- ") {
			cur = -1
			return false
		}
		if t := strings.TrimLeft(l, " \t"); cur >= 0 && !strings.HasPrefix(t, "+") && !strings.HasPrefix(t, "|") && !strings.HasPrefix(t, ">") {
			steps[cur].text += " " + strings.TrimSpace(l)
		}
		return false
	})
	return steps
}

// psfResolve is the repository path tok names, or "": with a `/`, tok
// itself when it exists under root; without one, the one **Files:**
// directory holding a file of that name.
func psfResolve(root, tok string, files []string) string {
	exists := func(p string) bool {
		_, err := os.Lstat(filepath.Join(root, p))
		return err == nil
	}
	if strings.Contains(tok, "/") {
		if exists(tok) {
			return tok
		}
		return ""
	}
	seen, hit := map[string]bool{}, []string{}
	for _, f := range files {
		d := path.Dir(f)
		if seen[d] {
			continue
		}
		seen[d] = true
		if p := path.Join(d, tok); exists(p) {
			hit = append(hit, p)
		}
	}
	if len(hit) == 1 {
		return hit[0]
	}
	return ""
}

// psfEditPos reports whether token k of text stands where an edit names
// its target: right after an edit verb, at the head of a list after `in:`,
// or in a list opening a clause and followed by `:`, `'s`, ` (` or a change
// verb. A list is tokens joined by `, `, ` and ` or `, and `, a
// parenthetical allowed between two.
func psfEditPos(text string, toks [][]int, k int) bool {
	gap := func(a, b int) string { return text[toks[a][1]:toks[b][0]] }
	j := k
	for j > 0 && psfListRE.MatchString(gap(j-1, j)) {
		j--
	}
	pre := text[:toks[j][0]]
	if psfVerbRE.MatchString(pre[max(0, len(pre)-30):]) || psfInListRE.MatchString(pre[max(0, len(pre)-10):]) {
		return true
	}
	if !psfOpenRE.MatchString(pre) {
		return false
	}
	e := k
	for e+1 < len(toks) && psfListRE.MatchString(gap(e, e+1)) {
		e++
	}
	return psfAfterRE.MatchString(text[toks[e][1]:])
}
