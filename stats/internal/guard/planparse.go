// planparse.go holds the tasks.md parsing helpers several guards share.

package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// tcfPlanLines is open(encoding="utf-8").read().splitlines(): strict UTF-8
// (a decode failure is an uncaught exception), universal newlines, then
// every line boundary str.splitlines() knows.
func tcfPlanLines(b []byte) ([]string, error) {
	s, err := tcfText(b)
	if err != nil {
		return nil, err
	}
	return tcfSplitLines(s), nil
}

// tcfParseTask is parse_task_fields; false is TaskNotFoundError.
func tcfParseTask(lines []string, taskID string) (tcfTask, bool) {
	found := tcfSelectTask(lines, taskID)
	if found == nil {
		return tcfTask{}, false
	}
	body := found.lines
	fields := tcfBodyFields(body)
	joined := func(name string) string { return strings.Join(fields[name], " ") }

	task := tcfTask{id: taskID, testsValue: joined("Tests")}
	if cp := fields["Commit"]; len(cp) > 0 {
		v := tcfStrip(strings.Join(cp, " "))
		if m := tcfBacktickAll.FindStringSubmatch(v); m != nil {
			v = m[1]
		}
		task.commit = &v
	}
	baselineValue := joined("Baseline")
	if m := tcfBaselineRE.FindStringSubmatch(baselineValue); m != nil {
		before, _ := new(big.Int).SetString(m[1], 10)
		after, _ := new(big.Int).SetString(m[2], 10)
		task.baseline = &[2]*big.Int{before, after}
	}
	for _, m := range tcfMeasuredRE.FindAllStringSubmatch(baselineValue, -1) {
		if c := tcfStrip(m[1]); c != "" && !slices.Contains(task.baselineMeasured, c) {
			task.baselineMeasured = append(task.baselineMeasured, c)
		}
	}
	task.build = tcfSelectBuildKind(body)
	task.squashValue, task.squashPartners, _ = tcfSelectSquashWith(body)
	if off := tcfUnclosedFence(body); off >= 0 {
		task.unclosedFenceLine = found.taskLine + off + 1
	}
	shorthand := tcfPathShorthand(lines)
	// `Files: none` declares no path, as `Tests: none` declares no test.
	if files := joined("Files"); !tcfNoneOpenRE.MatchString(files) {
		task.files = shorthand.expand(tcfBacktickTokens(files))
	}
	task.allowedCollateral = shorthand.expand(tcfBacktickTokens(joined("Allowed-collateral")))
	task.tests = tcfParseTestSpecs(task.testsValue)
	return task, true
}

// tcfBodyFields is parse_task_fields' field walk: each non-fenced field line
// opens a field, a blank or Squash-with line closes it, and every other line
// while one is open is a stripped continuation; a repeated field keeps its
// first occurrence, so a later prose line opening with a field never
// replaces the declaration. Joined with " ", a field's parts are the *_value strings
// parse_task_fields exposes (files_value, tests_value, collateral_value).
func tcfBodyFields(body []string) map[string][]string {
	fields := map[string][]string{}
	current, parts := "", []string(nil)
	commitField := func() {
		if _, seen := fields[current]; current != "" && !seen {
			fields[current] = parts
		}
	}
	tcfUnfenced(body, func(_ int, line string) bool {
		switch m := tcfFieldRE.FindStringSubmatch(line); {
		case tcfSquashFieldRE.MatchString(line):
			commitField()
			current, parts = "", nil
		case m != nil:
			commitField()
			current, parts = m[1], []string{m[2]}
		case tcfStrip(line) == "":
			commitField()
			current, parts = "", nil
		case current != "":
			parts = append(parts, tcfStrip(line))
		}
		return false
	})
	commitField()
	return fields
}

// tcfResolveTasksMD is the plan resolution close-task and
// check-review-gate share, so the two cannot drift: the tasks.md the named
// change, a satellite link or the glob-path resolves to. An empty path is a
// refusal -- detail is its could-not-judge text, or empty when stderr
// already carries the reason (an ambiguous state-record answer). label
// names the calling guard in that relayed reason. Once every spectre shape
// failed, a /flow-fast run's <worktree>/.superpowers/sdd/<plan>/tasks.md is
// tried (tcfFastPlan) -- never before, so a /flow plan is never shadowed.
func tcfResolveTasksMD(env core.Env, stderr io.Writer, label, worktree, name, canonical string) (string, string) {
	changes := worktree + "/" + core.SpecRootLeaf(worktree, stderr) + "/changes"

	if name != "" {
		if name == "." || name == ".." || strings.Trim(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-") != "" {
			return "", "invalid change name: " + name
		}
		tasksMD := ""
		if core.IsFile(changes + "/" + name + "/tasks.md") {
			// Scoped version of the highest-numbered-fix-sibling rule (see the
			// glob path below): look only at name's own -fix-N family, never at
			// any other directory under changes.
			var fixes []string
			for _, e := range tcfEntries(changes) {
				if strings.HasPrefix(e, name+"-fix-") && core.IsDir(changes+"/"+e) {
					fixes = append(fixes, e)
				}
			}
			tasksMD = changes + "/" + tcfHighestFix(changes, name, fixes) + "/tasks.md"
		} else {
			// One resolver for every shape without a local tasks.md: a
			// satellite's link.md, and — KAN-260 — NO change directory here at
			// all, where the plan lives only in the canonical worktree the
			// caller supplied. changePlanPath owns both branches and their
			// containment rules; every shape it cannot resolve still reaches the
			// refusal below. tcfResolvePlan adds one thing (KAN-267): an
			// ambiguous state-record answer is relayed and refused outright,
			// never read as an ordinary absence.
			p, rc := tcfResolvePlan(env, stderr, label, worktree, name, canonical)
			if rc == 3 {
				return "", ""
			}
			tasksMD = p
		}
		if tasksMD == "" || !core.IsFile(tasksMD) {
			if fast := tcfFastPlan(worktree, name); fast != "" {
				return fast, ""
			}
			return "", fmt.Sprintf("no tasks.md found for change '%s' under %s", name, changes)
		}
		return tasksMD, ""
	}

	var matches, names []string
	for _, e := range tcfEntries(changes) {
		if _, err := os.Stat(changes + "/" + e + "/tasks.md"); err == nil {
			matches = append(matches, changes+"/"+e+"/tasks.md")
			names = append(names, e)
		}
	}

	if len(matches) == 0 {
		// WHERE A LINK IS FOLLOWED (see header): find the link-only change
		// directories — link.md present, tasks.md absent — under changes.
		// Exactly one is a satellite worktree; anything else (none, or more
		// than one) cannot be resolved without guessing and falls through to
		// the same refusal a plain "no tasks.md at all" worktree always got.
		var satellites []string
		for _, e := range tcfEntries(changes) {
			d := changes + "/" + e
			if core.IsDir(d) && core.IsFile(d+"/link.md") && !core.IsFile(d+"/tasks.md") {
				satellites = append(satellites, e)
			}
		}
		tasksMD := ""
		if len(satellites) == 1 {
			p, rc := tcfResolvePlan(env, stderr, label, worktree, satellites[0], canonical)
			if rc == 3 {
				return "", ""
			}
			tasksMD = p
		}
		if tasksMD == "" || !core.IsFile(tasksMD) {
			if fast := tcfFastPlan(worktree, ""); fast != "" {
				return fast, ""
			}
			return "", "no tasks.md found under " + changes
		}
		return tasksMD, ""
	}

	// A <name>-fix-N SUB-CHANGE IS NOT AMBIGUITY. Under spectre a sub-change is
	// a FLAT SIBLING of its parent under spectre/changes/ -- `spectre new`
	// refuses an id that is not a single flat directory name -- so the scan
	// above matches the parent AND every fix sibling as soon as one exists.
	// Under OpenSpec a sub-change was nested and never matched, so "more than
	// one tasks.md" meant two unrelated changes and refusing was right.
	// Refusing now would take this guard out of service on exactly the runs it
	// was added for: every fix round after the first sub-change is created.
	var roots []string
	for _, n := range names {
		if !tcfIsFixSiblingOf(n, names) {
			roots = append(roots, n)
		}
	}
	// MORE THAN ONE ROOT IS STILL A REFUSAL, unchanged: two changes neither of
	// which is the other's fix sibling is the genuine ambiguity this check has
	// always existed to catch, and nothing here may guess between them.
	if len(roots) != 1 {
		return "", fmt.Sprintf("more than one tasks.md found under %s, cannot resolve which change: %s", changes, strings.Join(matches, " "))
	}
	// THE HIGHEST-NUMBERED FIX SIBLING WINS, and the root wins when there is
	// none. A fix round creates <name>-fix-N and implements THAT plan; an
	// earlier sub-change is finished, and the parent's own tasks were done
	// before any fix round opened. So the newest sub-change is the plan whose
	// tasks are being dispatched, which is the plan this guard has to read.
	//
	// TWO CAVEATS, both accepted. (1) N INCREMENTING IS CONVENTION, NOT
	// CONTRACT: nothing in skills/flow/review-panel.md specifies that a fix
	// round numbers its sub-change one higher than the last, so
	// "highest-numbered" reads an ordering nobody promised. (2) Reading the
	// wrong plan is normally LOUD rather than silent — the commit's files are
	// undeclared there and the guard exits 1 naming one — but it can pass
	// silently when the chosen plan's task N declares a SUPERSET of the
	// intended plan's files. Passing the change name as an argument would
	// remove both; that changes a call signature skills/flow/implement.md
	// documents, and was judged not worth it.
	return changes + "/" + tcfHighestFix(changes, roots[0], names) + "/tasks.md", ""
}

// tcfUnfenced calls fn for every body line outside a fence, with its offset;
// fn returning true stops the walk.
func tcfUnfenced(body []string, fn func(offset int, line string) bool) {
	inFence := false
	for i, line := range body {
		if tcfFenceRE.MatchString(line) {
			inFence = !inFence
			continue
		}
		if !inFence && fn(i, line) {
			return
		}
	}
}

// tcfPathShorthand is path_shorthand.
func tcfPathShorthand(lines []string) tcfShorthand {
	sh := tcfShorthand{to: map[string]string{}}
	tasks := tcfIterTasks(lines)
	if len(tasks) == 0 {
		return sh
	}
	var pending []string
	armed := false
	tcfUnfenced(lines[:tasks[0].taskLine-1], func(_ int, line string) bool {
		if tcfStrip(line) == "" {
			pending, armed = nil, false
			return false
		}
		for i, segment := range strings.Split(line, "`") {
			var words []string
			if i%2 == 1 {
				if segment != "" {
					words = []string{segment}
				}
			} else {
				for _, w := range strings.FieldsFunc(segment, tcfIsSpace) {
					if w = strings.Trim(w, ".,;:()[]"); w != "" {
						words = append(words, w)
					}
				}
			}
			for _, tok := range words {
				switch {
				case tok == "abbreviates" || tok == "abbreviating":
					armed = true
				case strings.Contains(tok, "/"):
					if armed {
						for _, a := range pending {
							if _, ok := sh.to[a]; !ok {
								sh.to[a] = tok
								sh.keys = append(sh.keys, a)
							}
						}
					}
					pending = nil
				case i%2 == 1:
					pending = append(pending, tok)
				}
			}
		}
		return false
	})
	return sh
}

// tcfSelectBuildKind is select_build_tag's kind: "green", "red", or "" for
// no tag or a first tag opening with neither.
func tcfSelectBuildKind(body []string) string {
	if tag := tcfSelectBuildTag(body); tag != nil {
		return tag.kind
	}
	return ""
}

// tcfBuildTag is BuildTag: offset is the 0-based body index of the line,
// kind "green", "red", or "" for a malformed tag, value the stripped text
// after **Build:**.
type tcfBuildTag struct {
	offset      int
	kind, value string
}

// tcfSelectBuildTag is select_build_tag: the first non-fenced **Build:**
// line, well-formed or not, or nil when the body carries none.
func tcfSelectBuildTag(body []string) *tcfBuildTag {
	var tag *tcfBuildTag
	tcfUnfenced(body, func(off int, line string) bool {
		m := tcfBuildLineRE.FindStringSubmatch(line)
		if m == nil {
			return false
		}
		tag = &tcfBuildTag{offset: off, value: tcfStrip(m[1])}
		if k := tcfBuildKindRE.FindStringSubmatch(m[1]); k != nil {
			tag.kind = k[1]
		}
		return true
	})
	return tag
}

// tcfBacktickTokens is _extract_backtick_tokens.
func tcfBacktickTokens(value string) []string {
	var tokens []string
	for _, m := range tcfBacktickRE.FindAllStringSubmatch(value, -1) {
		tokens = append(tokens, m[1])
	}
	if tokens == nil {
		if s := tcfStrip(value); s != "" {
			tokens = []string{s}
		}
	}
	return tokens
}

// tcfEntries is the names a "$dir"/*/... glob expands through: dotfiles
// excluded, none when dir cannot be read, and in the glob's order -- bash
// sorts the whole expanded path, so "a-b/" comes before "a/" ('-' < '/').
func tcfEntries(dir string) []string {
	entries, _ := os.ReadDir(dir)
	var names []string
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	sort.Slice(names, func(i, j int) bool { return names[i]+"/" < names[j]+"/" })
	return names
}

// tcfIsFixSiblingOf is is_fix_sibling_of_set: true when name is
// "<stem>-fix-<digits>" AND <stem> is itself one of the matched change names.
// The digit test is there because a change merely named like a neighbour
// (`demo-fix-the-parser`) is a change of its own, not this change's
// sub-change, and must still count as ambiguity.
func tcfIsFixSiblingOf(name string, names []string) bool {
	i, _, ok := tcfFixSuffix(name)
	return ok && slices.Contains(names, name[:i])
}

// tcfParseTestSpecs is _parse_test_specs.
func tcfParseTestSpecs(value string) []tcfTestSpec {
	if tcfNoneOpenRE.MatchString(value) {
		return nil
	}
	var numbers []string
	for _, m := range tcfCaseLabelRE.FindAllStringSubmatch(value, -1) {
		if !slices.Contains(numbers, m[1]) {
			numbers = append(numbers, m[1])
		}
	}
	var specs []tcfTestSpec
	for _, n := range numbers {
		specs = append(specs, tcfTestSpec{"Case " + n, regexp.MustCompile(`(?i)case[` + tcfWS + `]*` + n + `\b`)})
	}
	if specs != nil {
		return specs
	}
	for _, m := range tcfBacktickRE.FindAllStringSubmatch(value, -1) {
		specs = append(specs, tcfTestSpec{m[1], tcfFoldedPattern(m[1])})
	}
	return specs
}

// tcfHighestFix is highest_fix_sibling: the highest-numbered
// "<root>-fix-N" candidate whose own tasks.md exists, root unchanged when none
// qualify. Shared by the named-change resolution and the ambiguity-scan path,
// so the fix-sibling-selection rule (same digit validation, same numeric
// comparison, same tasks.md existence requirement, same root fallback) lives
// in exactly one place.
func tcfHighestFix(changes, root string, candidates []string) string {
	chosen, chosenN := root, int64(-1)
	for _, c := range candidates {
		if !strings.HasPrefix(c, root+"-fix-") {
			continue
		}
		_, suffix, ok := tcfFixSuffix(c)
		if !ok || !core.IsFile(changes+"/"+c+"/tasks.md") {
			continue
		}
		if n, _ := strconv.ParseInt(suffix, 10, 64); n > chosenN {
			chosen, chosenN = c, n
		}
	}
	return chosen
}

// tcfSelectSquashWith is select_squash_with: the first candidate whose value
// gates, else the first candidate with no partners; nil value for none.
// gated is the body offset of a gating line, -1 when none gates -- Python's
// `partners is not None`, which an empty partners slice cannot tell apart.
func tcfSelectSquashWith(body []string) (value *string, partners []string, gated int) {
	gated = -1
	tcfUnfenced(body, func(off int, line string) bool {
		m := tcfSquashFieldRE.FindStringSubmatch(line)
		if m == nil {
			return false
		}
		v := tcfStrip(m[1])
		if g := tcfSquashValueRE.FindStringSubmatch(v); g != nil {
			value, partners, gated = &v, tcfDottedIDRE.FindAllString(g[1], -1), off
			return true
		}
		if value == nil {
			value = &v
		}
		return false
	})
	return value, partners, gated
}

// tcfUnclosedFence is unclosed_fence: the offset of a fence the body opens
// and never closes, or -1.
func tcfUnclosedFence(body []string) int {
	opened := -1
	for i, line := range body {
		if tcfFenceRE.MatchString(line) {
			if opened >= 0 {
				opened = -1
			} else {
				opened = i
			}
		}
	}
	return opened
}

// tcfResolvePlan is resolve_plan_or_ambiguity: change_plan_path, whose
// stderr is relayed only for an ambiguous store answer (3). When the store
// step returns 3 — the state record resolved the name to more than one
// project's plan — the per-match lines are relayed and the caller turns the 3
// into this guard's outright exit-2 refusal (KAN-267). Checking one task's
// fields against a plan the record names twice would make the verdict a coin
// flip, which is the one thing a commit-fields guard must never be.
func tcfResolvePlan(env core.Env, stderr io.Writer, label, worktree, name, canonical string) (string, int) {
	var captured bytes.Buffer
	p, rc := changePlanPath(env, &captured, worktree, name, canonical)
	if rc == 3 {
		stderr.Write(captured.Bytes())
		fmt.Fprintf(stderr, "%s: the state record resolves change '%s' to more than one project's plan — cannot determine which\n", label, name)
	}
	return p, rc
}

var (
	tcfFieldRE     = regexp.MustCompile(`^\*\*(Files|Tests|Regression|Baseline|Commit|Allowed-collateral|Build|After|Decision|Opener|Size):\*\*[` + tcfWS + `]*(.*)$`)
	tcfBacktickAll = regexp.MustCompile("^`([^`]+)`$")
	tcfBaselineRE  = regexp.MustCompile(`before=([0-9]+)[` + tcfWS + `]+after=([0-9]+)`)
	tcfMeasuredRE  = regexp.MustCompile(`<!--[` + tcfWS + `]*measured:[` + tcfWS + `]*(.+?)[` + tcfWS + `]+@[` + tcfWS + `]+(.+?)[` + tcfWS + `]*-->`)
)

// tcfFixSuffix is ${name##*-fix-} when that is a non-empty digit run.
func tcfFixSuffix(name string) (int, string, bool) {
	i := strings.LastIndex(name, "-fix-")
	if i < 0 {
		return 0, "", false
	}
	suffix := name[i+len("-fix-"):]
	if suffix == "" || strings.Trim(suffix, "0123456789") != "" {
		return 0, "", false
	}
	return i, suffix, true
}

// tcfFoldedPattern is _folded_ws_pattern.
func tcfFoldedPattern(token string) *regexp.Regexp {
	words := strings.FieldsFunc(token, tcfIsSpace)
	for i := range words {
		words[i] = regexp.QuoteMeta(words[i])
	}
	return regexp.MustCompile(strings.Join(words, `[`+tcfWS+`]+`))
}

var (
	tcfBuildLineRE   = regexp.MustCompile(`^\*\*Build:\*\*(.*)$`)
	tcfBuildKindRE   = regexp.MustCompile(`^[` + tcfWS + `]+(green|red)\b`)
	tcfSquashFieldRE = regexp.MustCompile(`^\*\*Squash-with:\*\*[` + tcfWS + `]*(.*)$`)
	tcfSquashValueRE = regexp.MustCompile(`^Task[` + tcfWS + `]+([0-9.,` + tcfWS + `]+)[` + tcfWS + `]*$`)
	tcfDottedIDRE    = regexp.MustCompile(`[0-9]+(?:\.[0-9]+)*`)
)

// AFTER_FIELD_RE is the field's presence on one physical line; AFTER_VALUE_RE
// the gate: a value is `Task <ids>` or the literal `none`, nothing else.
var (
	tcfAfterFieldRE = regexp.MustCompile(`^\*\*After:\*\*[` + tcfWS + `]*(.*)$`)
	tcfAfterValueRE = regexp.MustCompile(`^(?:Task[` + tcfWS + `]+[0-9.,` + tcfWS + `]+|none)[` + tcfWS + `]*$`)
)

// tcfAfterIDs is after_ids: the ids an After value names, empty for `none`,
// and false when the value does not gate -- a caller never falls back to
// extracting ids from a raw value. The serial default (an absent field means
// every plan-order earlier task) is applied where after-sets are resolved,
// never here.
func tcfAfterIDs(value string) ([]string, bool) {
	if !tcfAfterValueRE.MatchString(value) {
		return nil, false
	}
	if tcfStrip(value) == "none" {
		return []string{}, true
	}
	return tcfDottedIDRE.FindAllString(value, -1), true
}

// tcfAfter is AfterField: offset is the 0-based body index of the line,
// value its stripped remainder, gates whether ids is what after_ids made of it.
type tcfAfter struct {
	offset int
	value  string
	ids    []string
	gates  bool
}

// tcfSelectAfter is select_after: the first non-fenced candidate whose value
// gates, else the first candidate with gates false; nil when the body carries
// none. Which line is the field is decided here and nowhere else, so a
// non-gating candidate ahead of the real field never shadows it.
func tcfSelectAfter(body []string) *tcfAfter {
	var ungated *tcfAfter
	var found *tcfAfter
	tcfUnfenced(body, func(off int, line string) bool {
		m := tcfAfterFieldRE.FindStringSubmatch(line)
		if m == nil {
			return false
		}
		v := tcfStrip(m[1])
		if ids, ok := tcfAfterIDs(v); ok {
			found = &tcfAfter{off, v, ids, true}
			return true
		}
		if ungated == nil {
			ungated = &tcfAfter{offset: off, value: v}
		}
		return false
	})
	if found != nil {
		return found
	}
	return ungated
}

// tcfPythonFiles is check-task-commit-fields.py's parse_task_fields files
// list, which -- unlike tcfParseTask's -- keeps a `none`-opening Files value
// as the prose path it falls back to (["none"] for `**Files:** none`): the
// Go commit guard learned that `Files: none` declares no path, the Python
// importers never did, and their ports keep the Python's verdicts -- F2
// stays silent on a measurement task's `**Files:** none`.
func tcfPythonFiles(lines []string, t tcfTask) []string {
	found := tcfSelectTask(lines, t.id)
	if t.files != nil || found == nil {
		return t.files
	}
	return tcfPathShorthand(lines).expand(tcfBacktickTokens(strings.Join(tcfBodyFields(found.lines)["Files"], " ")))
}

// tcfScanPlans is the no-argument scan's file list the tasks.md wrappers
// shared: every live change's tasks.md under root's spec leaf -- never an
// archive/ entry, which no change writes since run 1 drops the change
// directory -- each spelled root-relative as the wrapper's "$CHANGES_DIR"/...
// did and kept only while it exists. Empty with true when root has no
// changes directory. The bool stays for the callers' error path; nothing
// here fails today.
func tcfScanPlans(env core.Env, root string, stderr io.Writer) ([]string, bool) {
	changes := root + "/" + core.SpecRootLeaf(psAbs(env, root), stderr) + "/changes"
	if !core.IsDir(psAbs(env, changes)) {
		return nil, true
	}
	var files []string
	for _, name := range tcfEntries(psAbs(env, changes)) {
		if name != "archive" {
			files = append(files, changes+"/"+name+"/tasks.md")
		}
	}
	return slices.DeleteFunc(files, func(f string) bool {
		_, err := os.Stat(psAbs(env, f))
		return err != nil
	}), true
}

// pdbTask is plan-dispatch-bundles.py's Task: one task of a tasks.md, read
// for dispatch bundling. afterSet is the DECLARED ids when the task's
// **After:** value gates (afterDeclared; empty for `none` or `Task ,`) --
// with no field the serial default is applied where bundles are resolved,
// never here. afterMalformed holds the offending value when the field is
// present but does not gate. size is the stripped value of the task's first
// column-0, unfenced **Size:** line; sizeDeclared is false when it has none.
type pdbTask struct {
	id             string
	taskLine       int
	unchecked      bool
	filesPresent   bool
	files          []string
	partners       []string
	afterSet       []string
	afterDeclared  bool
	afterMalformed *string
	size           string
	sizeDeclared   bool
}

// pdbSizes is the one list of valid **Size:** values; a task without the
// field reads as `medium`.
var pdbSizes = []string{"small", "medium", "big"}

// pdbParseTasks is plan-dispatch-bundles.py's parse_tasks: every task of
// lines in document order (tcfIterTasks' task and body grammar), its open
// state read off its own task line, its gated **Squash-with:** partners and
// **After:** field, its **Files:** bullet entries, and its first column-0,
// unfenced **Size:** value.
//
// The first line opening with `**Files:**` opens the field; its entries are
// every immediately following `- ` bullet that is not a checkbox (`- [ ]`,
// `- [x]`, `- [X]`), read until the first line that is not such a bullet. A
// fence line toggles fence state without ending the run, and a fenced line
// is never read. The Files line's own inline value is read as
// tcfParseTask reads it, through tcfBacktickTokens, `none` declaring no path. An **Allowed-collateral:** entry never joins.
func pdbParseTasks(lines []string) []pdbTask {
	var tasks []pdbTask
	for _, b := range tcfIterTasks(lines) {
		t := pdbTask{id: b.id, taskLine: b.taskLine, unchecked: tcfTaskLineRE.FindStringSubmatch(lines[b.taskLine-1])[1] == " "}
		if _, partners, gated := tcfSelectSquashWith(b.lines); gated >= 0 {
			t.partners = partners
		}
		if a := tcfSelectAfter(b.lines); a != nil {
			if a.gates {
				t.afterSet, t.afterDeclared = a.ids, true
			} else {
				v := a.value
				t.afterMalformed = &v
			}
		}
		inFence, inFiles := false, false
		for _, line := range b.lines {
			if tcfFenceRE.MatchString(line) {
				inFence = !inFence
				continue
			}
			if inFence {
				continue
			}
			if inFiles {
				if entry, ok := pdbBullet(line); ok {
					t.files = append(t.files, pdbExtractPaths(entry)...)
					continue
				}
				inFiles = false
			}
			if v, ok := strings.CutPrefix(line, "**Size:**"); ok && !t.sizeDeclared {
				t.size, t.sizeDeclared = tcfStrip(v), true
			}
			if inline, ok := strings.CutPrefix(line, "**Files:**"); ok {
				t.filesPresent, inFiles = true, true
				if !tcfNoneOpenRE.MatchString(inline) {
					t.files = append(t.files, tcfBacktickTokens(inline)...)
				}
			}
		}
		tasks = append(tasks, t)
	}
	return tasks
}

// pdbBullet is BULLET_RE, `^- (?!\[[ xX]\])\s*(.*)$`: a Files entry line --
// never a checkbox, so neither a step nor the next task's own line is read
// as a path -- and its text after the dash and any whitespace.
func pdbBullet(line string) (string, bool) {
	rest, ok := strings.CutPrefix(line, "- ")
	if !ok || len(rest) >= 3 && rest[0] == '[' && strings.IndexByte(" xX", rest[1]) >= 0 && rest[2] == ']' {
		return "", false
	}
	return strings.TrimLeftFunc(rest, tcfIsSpace), true
}

var (
	pdbEntryPrefixRE = regexp.MustCompile(`^(?:Create|Modify|Delete):[` + tcfWS + `]*`)
	pdbBacktickRE    = regexp.MustCompile("`([^`]+)`")
)

// pdbExtractPaths is _extract_paths: one entry's every backtick-quoted path
// after a Create:/Modify:/Delete: prefix, or its remaining text when it
// quotes none. Paths compare as this literal text, never resolved.
func pdbExtractPaths(entry string) []string {
	stripped := tcfStrip(pdbEntryPrefixRE.ReplaceAllLiteralString(tcfStrip(entry), ""))
	var paths []string
	for _, m := range pdbBacktickRE.FindAllStringSubmatch(stripped, -1) {
		paths = append(paths, tcfStrip(m[1]))
	}
	if paths == nil {
		return []string{stripped}
	}
	return paths
}

// tcfReadPlan reads a tasks.md as the Python importers' open(path,
// encoding="utf-8") did, rendering a failure as str(OSError) naming the path
// as given, or as the UnicodeDecodeError's own message.
func tcfReadPlan(env core.Env, path string) ([]string, error) {
	lines, err := tcfReadPlanTyped(env, path)
	if err != nil {
		return nil, errors.New(strings.TrimPrefix(err.Error(), "UnicodeDecodeError: "))
	}
	return lines, nil
}

// tcfReadPlanTyped is tcfReadPlan keeping a decode failure's
// "UnicodeDecodeError: " prefix.
func tcfReadPlanTyped(env core.Env, path string) ([]string, error) {
	b, err := os.ReadFile(psAbs(env, path))
	if pe := (*fs.PathError)(nil); errors.As(err, &pe) {
		pe.Path = path
		return nil, tcfOSError(path, pe)
	}
	if err != nil {
		return nil, err
	}
	return tcfPlanLines(b)
}

// tcfFastPlan is a /flow-fast run's plan, <worktree>/.superpowers/sdd/<plan>/tasks.md,
// or "" when it is not a file. <plan> is name, else the worktree's current
// branch -- /flow-fast names its branch after the change; a detached HEAD
// names none.
func tcfFastPlan(worktree, name string) string {
	plan := name
	if plan == "" {
		out, err := exec.Command("git", "-C", worktree, "branch", "--show-current").Output()
		if err != nil {
			return ""
		}
		plan = strings.TrimSpace(string(out))
	}
	if plan == "" || plan == "." || plan == ".." || strings.Contains(plan, "/") {
		return ""
	}
	if p := filepath.Join(worktree, ".superpowers/sdd", plan, "tasks.md"); core.IsFile(p) {
		return p
	}
	return ""
}

// psWorse combines two exit codes as the wrapper's check_one did: 2 beats
// any other failure, which beats 0.
func psWorse(a, b int) int {
	switch {
	case a == 2 || b == 2:
		return 2
	case a != 0 || b != 0:
		return 1
	}
	return 0
}

func psAbs(env core.Env, p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(env.Dir, p)
}

// tcfRepoCommit is one "<worktree>=<sha>" pair of the third argument's
// per-repository commit map — or the single-commit form's one implicit pair.
type tcfRepoCommit struct{ worktree, commit string }

// tcfCrash is an exception check-task-commit-fields.py's main does not
// catch: Python printed a traceback and exited 1. Its message is the
// traceback's last line; the frames above it have no Go equivalent.
type tcfCrash struct{ msg string }

func (c tcfCrash) Error() string { return c.msg }

// ---- lib/plan_grammar.py ----

// tcfWS is Python's str \s: the characters str.isspace() accepts.
const tcfWS = `\t\n\v\f\r \x{1c}-\x{1f}\x{85}\p{Z}`

var (
	tcfFenceRE    = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
	tcfTaskLineRE = regexp.MustCompile(`^- \[([ x])\] ([0-9]+)\. `)
	tcfBoundaryRE = regexp.MustCompile(`^#{2,3}(?:[` + tcfWS + `]|$)`)
)

func tcfIsSpace(r rune) bool { return unicode.IsSpace(r) || r >= 0x1c && r <= 0x1f }

// tcfStrip is str.strip().
func tcfStrip(s string) string { return strings.TrimFunc(s, tcfIsSpace) }

type tcfTaskBody struct {
	id       string
	taskLine int // 1-based line of the task line; the body starts right after it
	lines    []string
}

// tcfIterTasks is iter_tasks.
func tcfIterTasks(lines []string) []tcfTaskBody {
	var tasks []tcfTaskBody
	open, openID, inFence := -1, "", false
	closeAt := func(end int) {
		if open >= 0 {
			tasks = append(tasks, tcfTaskBody{openID, open + 1, lines[open+1 : end]})
		}
	}
	for i, line := range lines {
		if tcfFenceRE.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if m := tcfTaskLineRE.FindStringSubmatch(line); m != nil {
			closeAt(i)
			open, openID = i, m[2]
			continue
		}
		if tcfBoundaryRE.MatchString(line) {
			closeAt(i)
			open = -1
		}
	}
	closeAt(len(lines))
	return tasks
}

// tcfSelectTask is select_task: the first task carrying taskID, or nil.
func tcfSelectTask(lines []string, taskID string) *tcfTaskBody {
	for _, t := range tcfIterTasks(lines) {
		if t.id == taskID {
			return &t
		}
	}
	return nil
}

// ---- check-task-commit-fields.py: field grammar ----

var (
	tcfBacktickRE  = regexp.MustCompile("`([^`]+)`")
	tcfCaseLabelRE = regexp.MustCompile(`(?i)\bCase[` + tcfWS + `]+([0-9]+[a-z]?)\b`)
	tcfNoneOpenRE  = regexp.MustCompile(`(?i)^[` + tcfWS + `*_]*none\b`)
)

type tcfTestSpec struct {
	label   string
	pattern *regexp.Regexp
}

// tcfTask is TaskFields; a nil pointer is Python's None.
type tcfTask struct {
	id                string
	files             []string
	tests             []tcfTestSpec
	testsValue        string
	allowedCollateral []string
	commit            *string
	baseline          *[2]*big.Int
	baselineMeasured  []string
	build             string
	squashValue       *string
	squashPartners    []string
	unclosedFenceLine int // 0: every fence the body opens is closed
}

// tcfText is Python's text-mode decode: strict UTF-8, then universal
// newlines.
func tcfText(b []byte) (string, error) {
	if !utf8.Valid(b) {
		return "", tcfDecodeError(b)
	}
	return strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\r", "\n"), nil
}

// tcfDecodeError is the UnicodeDecodeError line CPython's utf-8 codec
// raises for b's first invalid sequence.
func tcfDecodeError(b []byte) error {
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		i += size
	}
	lead, need, lo, hi := b[i], 0, byte(0x80), byte(0xBF)
	switch {
	case lead >= 0xC2 && lead <= 0xDF:
		need = 1
	case lead >= 0xE0 && lead <= 0xEF:
		need = 2
		if lead == 0xE0 {
			lo = 0xA0
		} else if lead == 0xED {
			hi = 0x9F
		}
	case lead >= 0xF0 && lead <= 0xF4:
		need = 3
		if lead == 0xF0 {
			lo = 0x90
		} else if lead == 0xF4 {
			hi = 0x8F
		}
	}
	reason, n := "invalid start byte", 1
	if need > 0 {
		reason = "invalid continuation byte"
		for k := 1; k <= need; k++ {
			if i+k >= len(b) {
				reason = "unexpected end of data"
				break
			}
			if c := b[i+k]; c < lo || c > hi {
				break
			}
			lo, hi = 0x80, 0xBF
			n++
		}
	}
	if n == 1 {
		return tcfCrash{fmt.Sprintf("UnicodeDecodeError: 'utf-8' codec can't decode byte 0x%02x in position %d: %s", lead, i, reason)}
	}
	return tcfCrash{fmt.Sprintf("UnicodeDecodeError: 'utf-8' codec can't decode bytes in position %d-%d: %s", i, i+n-1, reason)}
}

// tcfSplitLines is str.splitlines().
func tcfSplitLines(s string) []string {
	var out []string
	start := 0
	for i, r := range s {
		switch r {
		case '\n', '\r', '\v', '\f', 0x1c, 0x1d, 0x1e, 0x85, 0x2028, 0x2029:
			if r == '\n' && i > 0 && s[i-1] == '\r' {
				start = i + 1
				continue
			}
			out = append(out, s[start:i])
			start = i + utf8.RuneLen(r)
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// tcfTaskIDs is collect_task_ids.
func tcfTaskIDs(lines []string) []string {
	var ids []string
	for _, t := range tcfIterTasks(lines) {
		if !slices.Contains(ids, t.id) {
			ids = append(ids, t.id)
		}
	}
	return ids
}

// tcfShorthand is path_shorthand's answer, keys in insertion order.
type tcfShorthand struct {
	keys []string
	to   map[string]string
}

// expand is _expand_shorthand: longest abbreviation first.
func (sh tcfShorthand) expand(tokens []string) []string {
	if len(sh.keys) == 0 {
		return tokens
	}
	byLength := slices.Clone(sh.keys)
	sort.SliceStable(byLength, func(a, b int) bool {
		return utf8.RuneCountInString(byLength[a]) > utf8.RuneCountInString(byLength[b])
	})
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		for _, a := range byLength {
			if strings.HasPrefix(tok, a+"/") {
				tok = sh.to[a] + tok[len(a):]
				break
			}
		}
		out = append(out, tok)
	}
	return out
}

// Moved from check-task-commit-fields.py's _fold_group (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Resolve every partner a `Build: red` task's `Squash-with:` names —
// "one or more other tasks in the same plan", per **The build-green tag**
// (`skills/flow-contracts/build-green.md`). Returns `(partner files,
// partner allowed-collateral, the folded commit's expected subject,
// violations)`.
//
// The whole set is ONE unit folding into ONE commit ("whose commit this
// task's commit folds into" — singular, over a plural set). So every named
// partner has to declare the SAME `Commit:` subject: two partners naming
// different subjects cannot both describe the one surviving commit, and
// that disagreement is a plan defect reported here rather than a subject
// mismatch blamed on the commit. Taking the first-listed partner's subject
// and checking the commit against that would instead report a false
// mismatch whenever the real subject is a later-listed partner's.
//
// A partner declaring NO `Commit:` field contributes no subject (fix round
// 2, F5/F6). It is skipped rather than counted, so it cannot manufacture a
// disagreement with a partner that did declare one. Whether the fold ends
// up with no subject at all is NOT decided here (fix round 4, item 2):
// this red's partners declaring none is only half the question when
// another red joined to the same commit supplies one, so
// `resolve_folded_task` asks it of the combined group and this function
// simply returns `None` for the subject.
//
// Whether a partner is missing, itself red, or disagrees about the
// subject, the violation names the RED task — that is where the defective
// field is, whichever id the guard was invoked for.
//
// Scope is ONE red task's own `Squash-with:` list. Where two red tasks
// share a partner their folds are one combined unit, and it is
// `resolve_folded_task` that walks every red in that unit and reconciles
// what each of them returns here (fix round 3, F8).
//
// Inside the function:
//
// partner_error: this fold already has a defect in the partner SET
// itself — a partner that does not exist, or one that is itself red
// (fix round 3, F7). The disagreement check below is about the subjects
// the partners declared, which is only a meaningful question once the
// set resolved; and a fold with an unresolvable partner reporting a
// second violation would name one defect twice. The flag states that
// precondition instead of inferring it from `violations` being empty,
// which was true only because these are the sole violations that can
// precede it. Case 47 pins it: an unresolvable partner alongside two
// disagreeing resolvable ones reports the missing partner and nothing
// else.
//
// tcfFoldGroup is _fold_group.
func tcfFoldGroup(lines []string, red tcfTask) (files, collateral []string, subject *string, violations []string) {
	var subjects []string
	partnerError := false
	for _, id := range red.squashPartners {
		partner, found := tcfParseTask(lines, id)
		if !found {
			violations = append(violations, fmt.Sprintf("task %s: Squash-with: names Task %s, which does not exist in this plan", red.id, id))
			partnerError = true
			continue
		}
		if partner.build == "red" {
			violations = append(violations, fmt.Sprintf("task %s: Squash-with: names Task %s, which is itself red", red.id, id))
			partnerError = true
			continue
		}
		files = append(files, partner.files...)
		collateral = append(collateral, partner.allowedCollateral...)
		if partner.commit != nil && !slices.Contains(subjects, *partner.commit) {
			subjects = append(subjects, *partner.commit)
		}
	}
	if !partnerError && len(subjects) > 1 {
		violations = append(violations, fmt.Sprintf("task %s: Squash-with: names partners declaring different Commit: subjects %s — one folded unit is one commit, so one subject", red.id, tcfReprList(subjects)))
	}
	if len(subjects) == 1 {
		subject = &subjects[0]
	}
	return files, collateral, subject, violations
}

// Moved from check-task-commit-fields.py's _red_tasks (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Every `Build: red` task in the plan that carries a `Squash-with:`
// field at all, in document order — the reds whose folds and whose
// defective fields both have to be accounted for before any task in this
// plan can be checked.
//
// tcfRedTasks is _red_tasks.
func tcfRedTasks(lines []string) []tcfTask {
	var reds []tcfTask
	for _, id := range tcfTaskIDs(lines) {
		if t, _ := tcfParseTask(lines, id); t.build == "red" && t.squashValue != nil {
			reds = append(reds, t)
		}
	}
	return reds
}

// Moved from check-task-commit-fields.py's _joined_partners (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// The partner ids of `red` that RESOLVE — the named task exists in this
// plan and is not itself red — and therefore actually join `red`'s commit
// to theirs.
//
// An edge that does not resolve is not an edge (fix round 6, F15). It is
// a plan defect, reported by `_fold_group` against the red task carrying
// the field; but it names no commit, so it must not pull the named task —
// nor, transitively, the valid fold that task belongs to — into this red's
// membership set. Growing the group through such an edge made one red's
// broken reference fail a separate, entirely valid folded commit from
// every id in it.
//
// tcfJoinedPartners is _joined_partners.
func tcfJoinedPartners(lines []string, red tcfTask) []string {
	var joined []string
	for _, id := range red.squashPartners {
		if p, found := tcfParseTask(lines, id); found && p.build != "red" {
			joined = append(joined, id)
		}
	}
	return joined
}

// Moved from check-task-commit-fields.py's _fold_reds (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Every red task whose fold folds into the ONE commit `task` belongs to,
// in document order (fix round 3, F8). `reds` is `_red_tasks(lines)`,
// already known to name at least one partner each.
//
// `Squash-with:` is an edge, not a tree. Two red tasks may name the SAME
// green partner, and each of their commits then folds into that one
// partner's commit — so both folds, and every task in either, are one
// combined unit carried by one commit. Resolving only the first red found
// in document order gave the shared partner's id a narrower file union and
// a different subject from the red ids: two verdicts on one commit, which
// is exactly what "either id gives the same verdict" forbids.
//
// The group is therefore grown to a fixed point: a red belongs if it IS a
// member or NAMES one through a RESOLVING edge, and admitting it makes
// those partners members in turn, which may admit further reds. Membership
// is grown over `_joined_partners`, never over the raw `squash_partners`
// list (fix round 6, F15). A task that is neither a red with partners nor
// named by one yields the empty list and is left alone — the union widens
// a fold, never an ordinary task.
//
// tcfFoldReds is _fold_reds.
func tcfFoldReds(lines []string, task tcfTask, reds []tcfTask) []string {
	edges := map[string][]string{}
	for _, red := range reds {
		edges[red.id] = tcfJoinedPartners(lines, red)
	}
	members := map[string]bool{task.id: true}
	for growing := true; growing; {
		growing = false
		for _, red := range reds {
			joined := edges[red.id]
			if !members[red.id] && !slices.ContainsFunc(joined, func(p string) bool { return members[p] }) {
				continue
			}
			for _, id := range append([]string{red.id}, joined...) {
				if !members[id] {
					members[id], growing = true, true
				}
			}
		}
	}
	var out []string
	for _, red := range reds {
		if members[red.id] {
			out = append(out, red.id)
		}
	}
	return out
}

// Moved from check-task-commit-fields.py's _shared_partner (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// The first task named by more than one of `reds` — the join that made
// their folds one unit, and the task a disagreement between those folds is
// reported against. `None` when the reds share nobody, which a group of
// two or more reds cannot be (they are only in one group because they
// share a partner or because one names the other, and the latter is
// already a violation); the caller falls back rather than assume it.
//
// tcfSharedPartner is _shared_partner; "" is None.
func tcfSharedPartner(reds []tcfTask) string {
	var seen []string
	for _, red := range reds {
		for _, p := range red.squashPartners {
			if slices.Contains(seen, p) {
				return p
			}
			seen = append(seen, p)
		}
	}
	return ""
}

// Moved from check-task-commit-fields.py's resolve_folded_task (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Resolve a folded task against the whole fold its commit carries, per
// `flow-task-commit-fields`'s requirement **A folded red task is checked
// against its partner's commit**. Returns `(task to check, violations found
// resolving it)`; a non-empty violation list is a plan defect and the caller
// reports it instead of checking anything.
//
// After the fold a red task has no commit of its own: its declared
// `Commit:` subject exists nowhere, and one commit carries every folded
// task's files. So every task in the fold — the red tasks and their green
// partners alike — is checked against the fold's agreed `Commit:` subject
// with the UNION of the whole fold's `Files:`/`Allowed-collateral:` sets.
// Either id therefore reaches one verdict, invalid folds included: every
// red in the group is re-validated from every id, so a fold one id rejects
// cannot pass through another's.
//
// The fold is `_fold_reds`'s combined group, not one red's partner list
// (fix round 3, F8): where two reds name the same partner, both commits
// fold into that partner's one commit. Their partner sets must then agree
// about that commit's subject exactly as one red's partners must, and a
// disagreement between the two folds is the same plan defect — reported
// against the shared task that joined them, and naming the reds, rather
// than silently taking the first red in document order.
//
// The no-subject rule is asked of that same combined group (fix round 4,
// item 2): it fires only when NO red in the group has a partner declaring
// a `Commit:` subject. A red whose own partners declared none takes the
// subject a joined red's partner supplied, because there is one commit and
// therefore one subject.
//
// A task that neither declares a `Squash-with:` nor is named by a red task
// in the same plan is returned unchanged.
//
// Inside the function:
//
// A red whose `Squash-with:` value names no partner the grammar admits
// is reported from EVERY id in the plan, not only from its own (fix
// round 6, F16). The guard cannot know which commit that task folded
// into — that is the whole content of the defect — so it cannot know
// which folds the plan really has, and no verdict it reaches for any
// task in this plan is one it can vouch for. Reported only from the red
// itself, the defect was invisible from the task its free text names,
// which was instead told that the folded commit's other file was
// undeclared collateral: true of the commit, and about the wrong field.
//
// This is deliberately NOT how an unresolvable EDGE is treated. There
// the named partner and the reason the edge is illegal are both known,
// so the guard knows the fold does not exist and reports it against the
// red task alone (see `_joined_partners`). Here nothing is known.
// A red's own files, plus its partners' — including this task's own
// where it is one of them (a harmless duplicate), its SIBLING
// partners', and every task in a fold joined to this one.
// One combined fold is one commit, so it carries exactly one subject.
// Both failures — the folds disagreeing, and nobody declaring one at all
// — are asked of the COMBINED group (fix round 4, item 2 for the second
// of them): a red whose own partners declared no subject is no defect
// where a red joined to the same commit supplied one, since that is the
// fold's subject. Both are gated the same way `_fold_group` gates its own
// disagreement check: a group already carrying an unresolvable partner
// has its violation, and what its remaining partners happened to declare
// is not a second, separate defect. Both are anchored on the shared task
// that joined the folds, where there is one.
//
// Moved from check-task-commit-fields.py's _widen (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// The one widening step both of `resolve_folded_task`'s branches use:
// add the rest of the fold's declared paths to this task's own, and set
// the subject the surviving commit is expected to carry. Written once so
// the two branches cannot drift apart.
//
// tcfResolveFolded is resolve_folded_task.
func tcfResolveFolded(lines []string, task tcfTask) (tcfTask, []string) {
	reds := tcfRedTasks(lines)
	var unresolved []string
	for _, red := range reds {
		if len(red.squashPartners) == 0 {
			unresolved = append(unresolved, fmt.Sprintf("task %s: Squash-with: value %s is not `Task <id>[, <id>...]`, so no merge partner can be resolved from it", red.id, tcfRepr(*red.squashValue)))
		}
	}
	if unresolved != nil {
		return task, unresolved
	}
	foldReds := tcfFoldReds(lines, task, reds)
	if len(foldReds) == 0 {
		return task, nil
	}
	var extraFiles, extraCollateral, subjects, violations []string
	for _, id := range foldReds {
		red, _ := tcfParseTask(lines, id)
		files, collateral, subject, groupViolations := tcfFoldGroup(lines, red)
		violations = append(violations, groupViolations...)
		if id != task.id {
			extraFiles = append(extraFiles, red.files...)
			extraCollateral = append(extraCollateral, red.allowedCollateral...)
		}
		extraFiles = append(extraFiles, files...)
		extraCollateral = append(extraCollateral, collateral...)
		if subject != nil && !slices.Contains(subjects, *subject) {
			subjects = append(subjects, *subject)
		}
	}
	if violations == nil && len(subjects) != 1 {
		var group []tcfTask
		for _, id := range foldReds {
			red, _ := tcfParseTask(lines, id)
			group = append(group, red)
		}
		anchor := tcfSharedPartner(group)
		if anchor == "" {
			anchor = foldReds[0]
		}
		if len(subjects) > 0 {
			violations = append(violations, fmt.Sprintf("task %s: Task %s is folded by Tasks %s, whose partners declare different Commit: subjects %s — one folded unit is one commit, so one subject", anchor, anchor, strings.Join(foldReds, ", "), tcfReprList(subjects)))
		} else {
			violations = append(violations, fmt.Sprintf("task %s: no partner named by Squash-with: declares a Commit: subject in the fold carried by Tasks %s — the folded commit's subject would be checked against nothing", anchor, strings.Join(foldReds, ", ")))
		}
	}
	if violations != nil {
		return task, violations
	}
	task.files = append(slices.Clone(task.files), extraFiles...)
	task.allowedCollateral = append(slices.Clone(task.allowedCollateral), extraCollateral...)
	task.commit = &subjects[0]
	return task, nil
}

// ---- check-task-commit-fields.py: the checks against git ----

// tcfExec runs a command in dir and returns its raw output and exit code;
// err is set only when it could not run at all (Python's OSError).
func tcfExec(dir string, stdin []byte, name string, args ...string) (stdout, stderr []byte, code int, err error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	err = cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return o.Bytes(), e.Bytes(), exitErr.ExitCode(), nil
	}
	if err != nil {
		return nil, nil, 0, tcfOSError(name, err)
	}
	return o.Bytes(), e.Bytes(), 0, nil
}

// tcfOSError renders an error the way str(OSError) does.
func tcfOSError(name string, err error) error {
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("[Errno 2] No such file or directory: %s", tcfRepr(name))
	}
	var pathErr *fs.PathError
	var errno syscall.Errno
	if errors.As(err, &pathErr) && errors.As(pathErr.Err, &errno) {
		msg := errno.Error()
		return fmt.Errorf("[Errno %d] %s: %s", int(errno), strings.ToUpper(msg[:1])+msg[1:], tcfRepr(pathErr.Path))
	}
	return err
}

// tcfTextOutput decodes both streams as subprocess.run(text=True) does.
func tcfTextOutput(stdout, stderr []byte) (string, string, error) {
	so, err := tcfText(stdout)
	if err != nil {
		return "", "", err
	}
	se, err := tcfText(stderr)
	return so, se, err
}

// tcfRunGit is run_git.
func tcfRunGit(env core.Env, worktree string, args ...string) (string, error) {
	out, errOut, code, err := tcfExec(env.Dir, nil, "git", append([]string{"-C", worktree}, args...)...)
	if err != nil {
		return "", err
	}
	so, se, err := tcfTextOutput(out, errOut)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("git -C %s %s failed: %s", worktree, strings.Join(args, " "), tcfStrip(se))
	}
	return so, nil
}

// tcfAppRoots is every absolute, backticked `Repo root` cell of an `## apps`
// table body.
func tcfAppRoots(apps string) []string {
	var roots []string
	for _, line := range strings.Split(apps, "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 4 {
			continue
		}
		cell := strings.Trim(strings.TrimSpace(cells[2]), "`")
		if strings.HasPrefix(cell, "/") {
			roots = append(roots, cell)
		}
	}
	return roots
}

// tcfRepr is repr() of a str.
func tcfRepr(s string) string {
	quote := '\''
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		quote = '"'
	}
	var b strings.Builder
	b.WriteRune(quote)
	for _, r := range s {
		switch {
		case r == quote || r == '\\':
			b.WriteRune('\\')
			b.WriteRune(r)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r < 0x20 || r == 0x7f || r >= 0x80 && r < 0x100 && !unicode.IsPrint(r):
			fmt.Fprintf(&b, `\x%02x`, r)
		case !unicode.IsPrint(r) && r < 0x10000:
			fmt.Fprintf(&b, `\u%04x`, r)
		case !unicode.IsPrint(r):
			fmt.Fprintf(&b, `\U%08x`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteRune(quote)
	return b.String()
}

// tcfReprList is ", ".join(repr(s) for s in list).
func tcfReprList(list []string) string {
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = tcfRepr(s)
	}
	return strings.Join(out, ", ")
}

// tcfFnmatch is fnmatch.fnmatch on POSIX: the pattern translated as
// Python 3.14's fnmatch.translate translates it -- `*` crosses `/`, and a
// backslash is a literal -- save that `**/` at a segment start matches zero
// or more directories, as git reads it (kan-1090 self-review, F2).
func tcfFnmatch(name, pattern string) bool {
	re, err := regexp.Compile(`^(?s:` + tcfTranslate(pattern) + `)\z`)
	return err == nil && re.MatchString(name)
}

func tcfTranslate(pattern string) string {
	p := []rune(pattern)
	n := len(p)
	var b strings.Builder
	for i := 0; i < n; {
		c := p[i]
		i++
		switch c {
		case '*':
			start := i - 1
			for i < n && p[i] == '*' {
				i++
			}
			if i-start >= 2 && (start == 0 || p[start-1] == '/') && i < n && p[i] == '/' {
				b.WriteString("(?:.*/)?")
				i++
			} else {
				b.WriteString(".*")
			}
		case '?':
			b.WriteString(".")
		case '[':
			j := i
			if j < n && p[j] == '!' {
				j++
			}
			if j < n && p[j] == ']' {
				j++
			}
			for j < n && p[j] != ']' {
				j++
			}
			if j >= n {
				b.WriteString(`\[`)
				continue
			}
			stuff := string(p[i:j])
			if strings.Contains(stuff, "-") {
				var chunks []string
				k := i + 1
				if p[i] == '!' {
					k = i + 2
				}
				for {
					dash := slices.Index(p[min(k, j):j], '-')
					if dash < 0 {
						break
					}
					dash += min(k, j)
					chunks = append(chunks, string(p[i:dash]))
					i, k = dash+1, dash+3
				}
				if chunk := string(p[i:j]); chunk != "" {
					chunks = append(chunks, chunk)
				} else if len(chunks) > 0 {
					chunks[len(chunks)-1] += "-"
				}
				for k := len(chunks) - 1; k > 0; k-- {
					prev, cur := []rune(chunks[k-1]), []rune(chunks[k])
					if prev[len(prev)-1] > cur[0] {
						chunks[k-1] = string(prev[:len(prev)-1]) + string(cur[1:])
						chunks = slices.Delete(chunks, k, k+1)
					}
				}
				for k, s := range chunks {
					chunks[k] = strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), "-", `\-`)
				}
				stuff = strings.Join(chunks, "-")
			} else {
				stuff = strings.ReplaceAll(stuff, `\`, `\\`)
			}
			i = j + 1
			switch {
			case stuff == "":
				b.WriteString(`[^\x00-\x{10FFFF}]`) // Python's (?!): never matches
			case stuff == "!":
				b.WriteString(".")
			default:
				// Python's re reads a `[` inside a class as a literal; Go's
				// reads `[:alpha:]` as a POSIX class. Escape it.
				stuff = strings.ReplaceAll(stuff, "[", `\[`)
				for _, op := range []string{"&", "~", "|"} {
					stuff = strings.ReplaceAll(stuff, op, `\`+op)
				}
				if stuff[0] == '!' {
					stuff = "^" + stuff[1:]
				} else if stuff[0] == '^' || stuff[0] == '[' {
					stuff = `\` + stuff
				}
				b.WriteString("[" + stuff + "]")
			}
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	return b.String()
}
