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
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
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

// tcfResolveTasksMD is the plan resolution check-task-commit-fields and
// check-review-gate share, so the two cannot drift: the tasks.md the named
// change, a satellite link or the glob-path resolves to. An empty path is a
// refusal -- detail is its could-not-judge text, or empty when stderr
// already carries the reason (an ambiguous state-record answer). label
// names the calling guard in that relayed reason.
func tcfResolveTasksMD(env Env, stderr io.Writer, label, worktree, name, canonical string) (string, string) {
	changes := worktree + "/" + specRootLeaf(worktree, stderr) + "/changes"

	if name != "" {
		if name == "." || name == ".." || strings.Trim(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-") != "" {
			return "", "invalid change name: " + name
		}
		tasksMD := ""
		if isFile(changes + "/" + name + "/tasks.md") {
			// Scoped version of the highest-numbered-fix-sibling rule (see the
			// glob path below): look only at name's own -fix-N family, never at
			// any other directory under changes.
			var fixes []string
			for _, e := range tcfEntries(changes) {
				if strings.HasPrefix(e, name+"-fix-") && isDir(changes+"/"+e) {
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
		if tasksMD == "" || !isFile(tasksMD) {
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
			if isDir(d) && isFile(d+"/link.md") && !isFile(d+"/tasks.md") {
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
		if tasksMD == "" || !isFile(tasksMD) {
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
// The digit test is the same one check-cleanup-complete's sub-change row uses,
// and for the same reason: a change merely named like a neighbour
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
		if !ok || !isFile(changes+"/"+c+"/tasks.md") {
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
func tcfResolvePlan(env Env, stderr io.Writer, label, worktree, name, canonical string) (string, int) {
	var captured bytes.Buffer
	p, rc := changePlanPath(env, &captured, worktree, name, canonical)
	if rc == 3 {
		stderr.Write(captured.Bytes())
		fmt.Fprintf(stderr, "%s: the state record resolves change '%s' to more than one project's plan — cannot determine which\n", label, name)
	}
	return p, rc
}

var (
	tcfFieldRE     = regexp.MustCompile(`^\*\*(Files|Tests|Regression|Baseline|Commit|Allowed-collateral|Build|After|Decision|Opener):\*\*[` + tcfWS + `]*(.*)$`)
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

// tcfRepairsOpenRE is REPAIRS_OPEN_RE, the test-repair annotation (KAN-794):
// a Tests value opening with `none` (bold `**none**` included), an optional em dash,
// en dash or `--`, then `repairs existing tests:` -- the phrase and its colon required, so
// "none added — repairs existing tests" is not the form. Its only effect is
// check-plan-shape's F4 exemption.
var tcfRepairsOpenRE = regexp.MustCompile(`(?i)^[` + tcfWS + `*_]*none\b[*_]*[` + tcfWS + `]*(?:[—–]|--)?[` + tcfWS + `]*repairs[` + tcfWS + `]+existing[` + tcfWS + `]+tests[` + tcfWS + `]*:`)

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
func tcfScanPlans(env Env, root string, stderr io.Writer) ([]string, bool) {
	changes := root + "/" + specRootLeaf(psAbs(env, root), stderr) + "/changes"
	if !isDir(psAbs(env, changes)) {
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
// present but does not gate.
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
}

// pdbParseTasks is plan-dispatch-bundles.py's parse_tasks: every task of
// lines in document order (tcfIterTasks' task and body grammar), its open
// state read off its own task line, its gated **Squash-with:** partners and
// **After:** field, and its **Files:** bullet entries.
//
// The first line opening with `**Files:**` opens the field; its entries are
// every immediately following `- ` bullet that is not a checkbox (`- [ ]`,
// `- [x]`, `- [X]`), read until the first line that is not such a bullet. A
// fence line toggles fence state without ending the run, and a fenced line
// is never read. The Files line's own inline value is not read -- the
// Python never did; generate-relocation-comparison reads Files through
// tcfPythonFiles instead. An **Allowed-collateral:** entry never joins.
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
			if strings.HasPrefix(line, "**Files:**") {
				t.filesPresent, inFiles = true, true
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
func tcfReadPlan(env Env, path string) ([]string, error) {
	lines, err := tcfReadPlanTyped(env, path)
	if err != nil {
		return nil, errors.New(strings.TrimPrefix(err.Error(), "UnicodeDecodeError: "))
	}
	return lines, nil
}

// tcfReadPlanTyped is tcfReadPlan keeping a decode failure's
// "UnicodeDecodeError: " prefix, which check-plan-shape and
// check-task-build-green print after "cannot read file: ".
func tcfReadPlanTyped(env Env, path string) ([]string, error) {
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
