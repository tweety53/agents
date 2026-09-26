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
	"time"
	"unicode"
	"unicode/utf8"
)

// The port of scripts/check-task-commit-fields.sh at 0747740 -- its wrapper
// body, and the check-task-commit-fields.py it exec'd with the parts of
// lib/plan_grammar.py that module imports. The shim's header comment is
// canonical for the calling convention and the plan resolution; the Python
// module's docstring is canonical for the field grammar, the fold rules and
// every check. Neither is restated here: each function below names the
// Python or bash function it ports, and reproduces its output byte for byte.
// The Python module's verdict half -- the fold resolution, every check
// against git, check_task_commit and main -- no longer exists (panel round
// 0, F9): nothing ran it once the shim exec'd this port. Each deleted
// function's own docstring and comments moved here verbatim, beside its
// port, under a "Moved from" line; the names this file cites are those
// functions at 0747740.
// The wrapper body's own reasoning moved here with it (0747740), beside the
// resolution code it explains. The wrapper also checked that
// lib/plan_grammar.py, lib/spec-root.sh and lib/change-plan.sh were present
// before using them — a Python import and a `source` being invisible to
// check-guard-symlinks.sh's rule 2 — and those checks have no counterpart:
// the grammar and the resolvers are compiled in.
//
// KAN-676 adds the one check that reads a tag's PAYLOAD rather than a field:
// any provenance tag in the closing task's own record — a `verified:`/
// `unverified:` on a fence info string, a `measured:`/`predicted:` comment
// in its prose — must carry its evidence, the command, source URL or output
// that shows the check ran, or what confirms or would confirm the claim
// (`check_evidence_tags`, tcfCheckEvidenceTags). The evidence rule is
// canonical in skills/flow-contracts/plan-provenance.md; a tag naming nothing
// after the colon is the false-label shape that rule exists to prevent —
// worse than no tag, because the reader trusts it — so the close fails like
// any other field violation. The tag set is the plan-provenance guard's own,
// so a tag cannot pass one boundary and fail the other; each guard reports at
// its own boundary. Fence content is never scanned (a comment-shaped line
// inside a worked example is code), on lib/plan_grammar.py's own fence rule
// (tcfFenceRE here), the same toggle `parse_task_fields` uses.
//
// check-task-commit-fields.py keeps its parse half and check_files:
// check-task-records.py and check-plan-shape.py still load it as a module,
// and TestTaskFieldParseMatchesPython pins this parse against it.
//
// Python's str regexes are Unicode-aware; these are RE2. Whitespace (\s,
// str.strip, str.split) is Python's Unicode set throughout (tcfWS,
// tcfIsSpace); \d, \w and \b stay ASCII, so a non-ASCII digit or letter
// directly beside a task id, a Case label or a build keyword can parse
// differently. No plan in the corpus writes one.

func init() { Registry["check-task-commit-fields"] = checkTaskCommitFields }

// Moved from check-task-commit-fields.py's MEASURED_TIMEOUT_SECONDS (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// The wall-clock ceiling on one recorded-command run (KAN-409 fix round,
// panel F4): a recorded command that hangs must not hang the guard. Ten
// minutes covers a real measurement command — the plan author recorded it
// as runnable — and expiry is the skip path below, never a verdict.
//
// tcfMeasuredTimeout is MEASURED_TIMEOUT_SECONDS.
const tcfMeasuredTimeout = 600 * time.Second

// Moved from check-task-commit-fields.py's NOT_A_VERDICT (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// NOT_A_VERDICT — the exact opening every could-not-judge exit (2) prints.
// The Go port duplicates this literal (tcfNotAVerdict,
// stats/internal/guard/taskcommitfields.go — Go cannot import a Python
// constant); TestCheckTaskCommitFields/case_92-93 asserts the full
// opening the guard prints. Exit 1 is the
// verdict against the commit; exit 2 is a caller or environment mistake,
// and this prefix is what keeps the two apart at a glance in a scrolling
// transcript (KAN-330: a merge base mistyped by one character used to
// surface git's bare "ambiguous argument" message here, indistinguishable
// from the guard refusing the commit).
//
// tcfNotAVerdict is NOT_A_VERDICT, the opening of every exit-2 refusal.
const tcfNotAVerdict = "check-task-commit-fields: COULD NOT JUDGE — not a commit verdict:"

func checkTaskCommitFields(args []string, env Env, stdout, stderr io.Writer) int {
	// refuse is could_not_judge: the one printer for every exit-2 refusal, so
	// the EXIT CODES opening stays identical at every site. The detail is the
	// site's own specifics; this adds the standard opening in front of it.
	refuse := func(detail string) int {
		fmt.Fprintf(stderr, "%s %s\n", tcfNotAVerdict, detail)
		return 2
	}
	if len(args) < 3 || len(args) > 6 {
		return refuse(fmt.Sprintf("usage: check-task-commit-fields.sh <worktree> <task-id> <commit-sha> [parent-sha] [canonical-worktree] [change-name] — got %d argument(s)", len(args)))
	}
	a := append(slices.Clone(args), "", "", "")
	worktree, taskID, commit, parent, canonical, name := a[0], a[1], a[2], a[3], a[4], a[5]

	// A task id is ONE flat integer — plan_grammar.py's TASK_ID, spectre's own
	// task-line id. Anything else is a caller mistake, not a plan fact, and it
	// is refused here at the argument boundary rather than flowed down to the
	// grammar: the shape this refusal exists for is a verification loop
	// handing its whole joined task list to this one argument, which used to
	// surface as "task <joined list> not found" — every task reported missing,
	// read as a plan defect, and costing a rerun one call per task to diagnose
	// (KAN-528). A dotted id is no task either (plan_grammar's DOTTED_ID), so
	// the digit test refuses it from here too, where "task 1.2 not found"
	// would have read as the task missing rather than the argument malformed —
	// with its own remediation line, since the fix for a dotted id is a flat
	// id, not a per-task rerun (panel F4). A leading zero ("01") names no task
	// spectre ever wrote either, and gets the same refusal rather than a
	// "task 01 not found" that reads as the task missing (panel F1).
	switch {
	case taskID == "" || taskID[0] == '0' || strings.Trim(taskID, "0123456789.") != "":
		return refuse(fmt.Sprintf("task id argument is not a single flat-integer task id: '%s' — a joined multi-task list in one call is the loop-clobber shape; invoke one call per task", taskID))
	case strings.Contains(taskID, "."):
		return refuse(fmt.Sprintf("task id argument is a dotted id, and a dotted id is no task to spectre: '%s' — use the task's own flat integer id", taskID))
	}
	if !isDir(worktree) {
		return refuse("worktree not found: " + worktree)
	}
	changes := worktree + "/" + specRootLeaf(worktree, stderr) + "/changes"

	// judge is dispatch_python_guard: the check against tasksMD, parent sha
	// forwarded. Shared by all three resolution paths (named change, satellite
	// link, glob-path) so the dispatch lives in exactly one place.
	judge := func(tasksMD string) int {
		violations, err := tcfCheckTaskCommit(env, tasksMD, taskID, worktree, commit, parent)
		var crash tcfCrash
		switch {
		case errors.As(err, &crash):
			fmt.Fprintln(stderr, crash.Error())
			return 1
		case err != nil:
			return refuse(err.Error())
		}
		for _, v := range violations {
			fmt.Fprintln(stdout, v)
		}
		if len(violations) > 0 {
			return 1
		}
		return 0
	}

	if name != "" {
		if name == "." || name == ".." || strings.Trim(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-") != "" {
			return refuse("invalid change name: " + name)
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
			p, rc := tcfResolvePlan(env, stderr, worktree, name, canonical)
			if rc == 3 {
				return 2
			}
			tasksMD = p
		}
		if tasksMD == "" || !isFile(tasksMD) {
			return refuse(fmt.Sprintf("no tasks.md found for change '%s' under %s", name, changes))
		}
		return judge(tasksMD)
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
			p, rc := tcfResolvePlan(env, stderr, worktree, satellites[0], canonical)
			if rc == 3 {
				return 2
			}
			tasksMD = p
		}
		if tasksMD == "" || !isFile(tasksMD) {
			return refuse("no tasks.md found under " + changes)
		}
		return judge(tasksMD)
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
		return refuse(fmt.Sprintf("more than one tasks.md found under %s, cannot resolve which change: %s", changes, strings.Join(matches, " ")))
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
	return judge(changes + "/" + tcfHighestFix(changes, roots[0], names) + "/tasks.md")
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

// tcfResolvePlan is resolve_plan_or_ambiguity: change_plan_path, whose
// stderr is relayed only for an ambiguous store answer (3). When the store
// step returns 3 — the state record resolved the name to more than one
// project's plan — the per-match lines are relayed and the caller turns the 3
// into this guard's outright exit-2 refusal (KAN-267). Checking one task's
// fields against a plan the record names twice would make the verdict a coin
// flip, which is the one thing a commit-fields guard must never be.
func tcfResolvePlan(env Env, stderr io.Writer, worktree, name, canonical string) (string, int) {
	var captured bytes.Buffer
	p, rc := changePlanPath(env, &captured, worktree, name, canonical)
	if rc == 3 {
		stderr.Write(captured.Bytes())
		fmt.Fprintf(stderr, "check-task-commit-fields.sh: the state record resolves change '%s' to more than one project's plan — cannot determine which\n", name)
	}
	return p, rc
}

// tcfCrash is an exception check-task-commit-fields.py's main does not
// catch: Python printed a traceback and exited 1. Its message is the
// traceback's last line; the frames above it have no Go equivalent.
type tcfCrash struct{ msg string }

func (c tcfCrash) Error() string { return c.msg }

// ---- lib/plan_grammar.py ----

// tcfWS is Python's str \s: the characters str.isspace() accepts.
const tcfWS = `\t\n\v\f\r \x{1c}-\x{1f}\x{85}\p{Z}`

var (
	tcfFenceRE       = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
	tcfTaskLineRE    = regexp.MustCompile(`^- \[([ x])\] ([0-9]+)\. `)
	tcfBoundaryRE    = regexp.MustCompile(`^#{2,3}(?:[` + tcfWS + `]|$)`)
	tcfBuildLineRE   = regexp.MustCompile(`^\*\*Build:\*\*(.*)$`)
	tcfBuildKindRE   = regexp.MustCompile(`^[` + tcfWS + `]+(green|red)\b`)
	tcfSquashFieldRE = regexp.MustCompile(`^\*\*Squash-with:\*\*[` + tcfWS + `]*(.*)$`)
	tcfSquashValueRE = regexp.MustCompile(`^Task[` + tcfWS + `]+([0-9.,` + tcfWS + `]+)[` + tcfWS + `]*$`)
	tcfDottedIDRE    = regexp.MustCompile(`[0-9]+(?:\.[0-9]+)*`)
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

// tcfSelectBuildKind is select_build_tag's kind: "green", "red", or "" for
// no tag or a first tag opening with neither.
func tcfSelectBuildKind(body []string) string {
	kind := ""
	tcfUnfenced(body, func(_ int, line string) bool {
		m := tcfBuildLineRE.FindStringSubmatch(line)
		if m == nil {
			return false
		}
		if k := tcfBuildKindRE.FindStringSubmatch(m[1]); k != nil {
			kind = k[1]
		}
		return true
	})
	return kind
}

// tcfSelectSquashWith is select_squash_with: the first candidate whose value
// gates, else the first candidate with no partners; nil value for none.
func tcfSelectSquashWith(body []string) (value *string, partners []string) {
	tcfUnfenced(body, func(_ int, line string) bool {
		m := tcfSquashFieldRE.FindStringSubmatch(line)
		if m == nil {
			return false
		}
		v := tcfStrip(m[1])
		if g := tcfSquashValueRE.FindStringSubmatch(v); g != nil {
			value, partners = &v, tcfDottedIDRE.FindAllString(g[1], -1)
			return true
		}
		if value == nil {
			value = &v
		}
		return false
	})
	return value, partners
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

// ---- check-task-commit-fields.py: field grammar ----

var (
	tcfFieldRE     = regexp.MustCompile(`^\*\*(Files|Tests|Regression|Baseline|Commit|Allowed-collateral|Build|After|Decision):\*\*[` + tcfWS + `]*(.*)$`)
	tcfBacktickRE  = regexp.MustCompile("`([^`]+)`")
	tcfBacktickAll = regexp.MustCompile("^`([^`]+)`$")
	tcfCaseLabelRE = regexp.MustCompile(`(?i)\bCase[` + tcfWS + `]+([0-9]+)\b`)
	tcfNoneOpenRE  = regexp.MustCompile(`(?i)^[` + tcfWS + `*_]*none\b`)
	tcfBaselineRE  = regexp.MustCompile(`before=([0-9]+)[` + tcfWS + `]+after=([0-9]+)`)
	tcfMeasuredRE  = regexp.MustCompile(`<!--[` + tcfWS + `]*measured:[` + tcfWS + `]*(.+?)[` + tcfWS + `]+@[` + tcfWS + `]+(.+?)[` + tcfWS + `]*-->`)
	tcfCamelRE     = regexp.MustCompile(`\b[A-Za-z][a-z0-9]*(?:[A-Z][a-z0-9]*[a-z][a-z0-9]*)+\b`)
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

// tcfParseTask is parse_task_fields; false is TaskNotFoundError.
func tcfParseTask(lines []string, taskID string) (tcfTask, bool) {
	found := tcfSelectTask(lines, taskID)
	if found == nil {
		return tcfTask{}, false
	}
	body := found.lines

	fields := map[string][]string{}
	current, parts := "", []string(nil)
	commitField := func() {
		if current != "" {
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
	task.squashValue, task.squashPartners = tcfSelectSquashWith(body)
	if off := tcfUnclosedFence(body); off >= 0 {
		task.unclosedFenceLine = found.taskLine + off + 1
	}
	shorthand := tcfPathShorthand(lines)
	task.files = shorthand.expand(tcfBacktickTokens(joined("Files")))
	task.allowedCollateral = shorthand.expand(tcfBacktickTokens(joined("Allowed-collateral")))
	task.tests = tcfParseTestSpecs(task.testsValue)
	return task, true
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

// tcfShorthand is path_shorthand's answer, keys in insertion order.
type tcfShorthand struct {
	keys []string
	to   map[string]string
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

// tcfFoldedPattern is _folded_ws_pattern.
func tcfFoldedPattern(token string) *regexp.Regexp {
	words := strings.FieldsFunc(token, tcfIsSpace)
	for i := range words {
		words[i] = regexp.QuoteMeta(words[i])
	}
	return regexp.MustCompile(strings.Join(words, `[`+tcfWS+`]+`))
}

// tcfTreeNames is _extract_tree_names.
func tcfTreeNames(testsValue string) []string {
	if tcfNoneOpenRE.MatchString(testsValue) || tcfCaseLabelRE.MatchString(testsValue) {
		return nil
	}
	var names []string
	add := func(n string) {
		if !slices.Contains(names, n) {
			names = append(names, n)
		}
	}
	for _, m := range tcfBacktickRE.FindAllStringSubmatch(testsValue, -1) {
		add(m[1])
	}
	for _, n := range tcfCamelRE.FindAllString(testsValue, -1) {
		add(n)
	}
	return names
}

// ---- check-task-commit-fields.py: folded red tasks ----

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
func tcfRunGit(env Env, worktree string, args ...string) (string, error) {
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

// Moved from check-task-commit-fields.py's check_task_commit (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Returns the violations that fail the run (exit 1).
//
// Inside the function:
//
// An unclosed fence in this task's body is reported on its own, before
// anything else, and REPLACES every downstream check for it (fix round
// 11, F22). CommonMark runs an unclosed fence to the end of the block,
// so every field after it is code rather than a declaration: `Files:`
// arrives empty, `Commit:` absent, `Build:` and `Squash-with:` gone.
// Checking on against that field set does not merely add noise — it
// blames the commit for a defect in the plan, reporting the file the
// task really did declare as undeclared collateral. The precedent is
// the unresolvable fold below: a task whose declaration cannot be read
// is one no verdict can be reached for.
// A folded pair is resolved before anything is checked; an unresolvable
// one is a plan defect reported on its own, since checking a red task
// against a commit that is not its own would report every field wrong.
// --no-renames pins the changed-file list against the invoking machine's
// git config: rename detection (on by default) elides a rename's source
// path, so a commit renaming a declared path would judge differently —
// or not at all — on another machine (panel round 0, F1). With it a
// rename is what git without detection sees: the old path deleted, the
// new added, both listed, both "touched".
// The full diff text gets the same pin: with rename detection a renamed
// file contributes only its modified hunks, without it the whole delete
// and add blocks — a Tests: name on a renamed file's unchanged lines is
// found on one machine and missed on another.
// `folded` carries the union file set and the surviving commit's expected
// subject; `task` carries what this task itself declared, which is what
// Tests: and the declared-scope check are about either way.
//
// tcfCheckTaskCommit is check_task_commit.
func tcfCheckTaskCommit(env Env, tasksMD, taskID, worktree, commit, parent string) ([]string, error) {
	raw, err := os.ReadFile(tasksMD)
	if err != nil {
		return nil, tcfOSError(tasksMD, err)
	}
	lines, err := tcfPlanLines(raw)
	if err != nil {
		return nil, err
	}
	task, found := tcfParseTask(lines, taskID)
	if !found {
		return nil, fmt.Errorf("task %s not found", taskID)
	}
	tasksAbs, _ := filepath.Abs(tasksMD)
	changeName := filepath.Base(filepath.Dir(tasksAbs))

	if task.unclosedFenceLine != 0 {
		return []string{fmt.Sprintf("task %s: code fence opened at tasks.md line %d is never closed in this task's body, so every field below it is inside the fence and was not read", task.id, task.unclosedFenceLine)}, nil
	}
	folded, squashViolations := tcfResolveFolded(lines, task)
	if squashViolations != nil {
		return squashViolations, nil
	}

	if parent == "" {
		out, err := tcfRunGit(env, worktree, "rev-parse", commit+"^")
		if err != nil {
			return nil, err
		}
		parent = tcfStrip(out)
	}
	names, err := tcfRunGit(env, worktree, "diff", "--no-renames", "--name-only", parent+".."+commit)
	if err != nil {
		return nil, err
	}
	var changed []string
	for _, p := range tcfSplitLines(names) {
		if p != "" {
			changed = append(changed, p)
		}
	}
	diff, err := tcfRunGit(env, worktree, "diff", "--no-renames", parent+".."+commit)
	if err != nil {
		return nil, err
	}
	subject, err := tcfRunGit(env, worktree, "log", "-1", "--format=%s", commit)
	if err != nil {
		return nil, err
	}

	var v []string
	// The evidence rule (KAN-676) is a plan-text check like the unclosed
	// fence above, but a defective tag does not blind the field checks the
	// way an unclosed fence does — it joins them as one more violation of
	// the same close.
	v = append(v, tcfCheckEvidenceTags(*tcfSelectTask(lines, taskID))...)
	v = append(v, tcfCheckFiles(folded, changed)...)
	v = append(v, tcfCheckDeclaredFiles(folded, changed)...)
	v = append(v, tcfCheckTests(task, diff)...)
	counts, err := tcfCheckBaselineCounts(env, task, worktree, changed, parent, commit)
	if err != nil {
		return nil, err
	}
	v = append(v, counts...)
	measured, err := tcfCheckBaselineMeasured(env, task, worktree, parent, commit, tcfMeasuredTimeout)
	if err != nil {
		return nil, err
	}
	v = append(v, measured...)
	inTree, err := tcfCheckTestsInTree(env, task, worktree, commit, tasksAbs)
	if err != nil {
		return nil, err
	}
	v = append(v, inTree...)
	v = append(v, tcfCheckCommitSubject(folded, tcfStrip(subject))...)
	v = append(v, tcfCheckCommitScope(task, changeName)...)
	return v, nil
}

// tcfCheckFiles is check_files.
func tcfCheckFiles(task tcfTask, changed []string) []string {
	var v []string
	for _, path := range changed {
		if slices.Contains(task.files, path) || slices.ContainsFunc(task.allowedCollateral, func(g string) bool { return tcfFnmatch(path, g) }) {
			continue
		}
		v = append(v, fmt.Sprintf("task %s: file %s is not declared in Files: and not covered by Allowed-collateral:", task.id, path))
	}
	return v
}

// Moved from check-task-commit-fields.py's check_declared_files (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Every path the task's `Files:` declares must appear in the commit's
// diff — the mirror of check_files, which requires the reverse (KAN-540:
// kan-30's task 34 declared a baseline its commit never carried, and the
// record then named a surface the branch never had). Same changed-file
// list check_files reads, so a deletion of a declared path counts as
// touching it; `Allowed-collateral:` is an allowance, not a declaration,
// and names nothing that has to appear. The list is deduplicated because
// the fold union joins partner lists that may repeat a path, and one
// defect is one violation line, not one per copy.
//
// tcfCheckDeclaredFiles is check_declared_files.
func tcfCheckDeclaredFiles(task tcfTask, changed []string) []string {
	var v, seen []string
	for _, path := range task.files {
		if slices.Contains(seen, path) {
			continue
		}
		seen = append(seen, path)
		if !slices.Contains(changed, path) {
			v = append(v, fmt.Sprintf("task %s: file %s is declared in Files: but the commit does not touch it", task.id, path))
		}
	}
	return v
}

// Moved from check-task-commit-fields.py's TESTS_PARSE_RULE (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// TESTS_PARSE_RULE — the parse rule both sibling `Tests:` violation messages
// state, the diff check's and the tree check's, so one prose-misuse failure
// is diagnosed one way (panel round 0, F1/F3/F4). The rule is conditional
// because the parser is: a `Case <N>` label makes the field label-checked
// with its backticks parsed as nothing, so an unconditional "every
// backticked token is a declared test name" claim would be false on exactly
// the labelled shape. The bare-camelCase clause is the tree check's own
// extension (panel round 1, F5): `_extract_tree_names` also extracts a bare
// camelCase identifier, so a tree-check violation for an unbackticked token
// must not arrive under a rule under which it was never declared.
//
// tcfTestsParseRule is TESTS_PARSE_RULE.
const tcfTestsParseRule = "a **Tests:** field is parsed, not read: a `Case <N>` label is checked by that label alone and backticks beside it parse as nothing, any other backticked token is a declared test name, the tree check also treats a bare camelCase identifier as one, and a task adding no tests opens the field with `none`; coverage prose belongs outside the field"

// Moved from check-task-commit-fields.py's DIFF_LEADING_MARKER_RE (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// DIFF_LEADING_MARKER_RE — one leading `+` or `-` per diff line, stripped so
// the search below runs on rendered lines rather than diff syntax: the wrap
// this check has to bridge (KAN-562) arrives in the raw diff as a newline
// followed by the added line's `+`, which no whitespace fold can cross. One
// strip per line, so content that itself leads with `+`/`-` (a Markdown
// bullet, say) keeps it.
var tcfDiffMarkerRE = regexp.MustCompile(`(?m)^[+-]`)

// tcfCheckTests is check_tests.
func tcfCheckTests(task tcfTask, diff string) []string {
	rendered := tcfDiffMarkerRE.ReplaceAllString(diff, "")
	var v []string
	for _, spec := range task.tests {
		if !spec.pattern.MatchString(rendered) {
			v = append(v, fmt.Sprintf("task %s: declared test %s not found in the diff — %s", task.id, spec.label, tcfTestsParseRule))
		}
	}
	return v
}

// Moved from check-task-commit-fields.py's count_test_annotations (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Occurrences of `@Test` across <paths> at <revision> — `-o` so a file
// holding several matches on one line still counts each one, and `-w` so a
// lifecycle annotation like `@TestFactory` or `@TestInstance` is not
// counted as a test (panel S1): only whole-word `@Test` is.
//
// Moved from check-task-commit-fields.py's _git_grep (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Run `git grep` in <worktree> and return (stdout, stderr, exit code)
// without raising on exit 1 — git grep's no-match code, which every caller
// here reads as a zero count or an absent name, never as an error. The
// stderr rides along so a real failure's message reaches the raised
// RuntimeError instead of dying as a bare exit code (panel F3).
//
// tcfCountTests is count_test_annotations.
func tcfCountTests(env Env, worktree, revision string, paths []string) (int, error) {
	if len(paths) == 0 {
		return 0, nil
	}
	out, errOut, code, err := tcfExec(env.Dir, nil, "git", append([]string{"-C", worktree, "grep", "-o", "-w", "-F", "@Test", revision, "--"}, paths...)...)
	if err != nil {
		return 0, err
	}
	so, se, err := tcfTextOutput(out, errOut)
	if err != nil {
		return 0, err
	}
	switch code {
	case 0:
		return len(tcfSplitLines(so)), nil
	case 1:
		return 0, nil
	}
	return 0, fmt.Errorf("git grep @Test at %s failed with exit %d: %s", revision, code, tcfStrip(se))
}

// Moved from check-task-commit-fields.py's check_baseline_counts (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// A declared `**Baseline:** before=N after=M` must equal the `@Test`
// delta the commit's changed files measure, at the parent and at the
// commit. Skips — never fails — when neither revision carries a single
// `@Test` in the counted set: the field's unit is then prose-declared
// (harnesses, cases), not annotations, and no verdict about it can be
// reached from `@Test` counts (the kan-100 spec's skip-not-fail rule,
// which KAN-442's removal left standing).
//
// Inside the function:
//
// One distinct recorded command supersedes this static count (KAN-409
// fix round, panel F1): the command then DEFINES the declared unit, and
// check_baseline_measured verifies the declaration against what that
// command measures. Stacking the two let the @Test prose this guard's
// own source carries (docstrings included) fail a commit whose recorded
// measurement agreed with the declaration.
//
// tcfCheckBaselineCounts is check_baseline_counts.
func tcfCheckBaselineCounts(env Env, task tcfTask, worktree string, changed []string, parent, commit string) ([]string, error) {
	if task.baseline == nil {
		return nil, nil
	}
	before, err := tcfCountTests(env, worktree, parent, changed)
	if err != nil {
		return nil, err
	}
	after, err := tcfCountTests(env, worktree, commit, changed)
	if err != nil {
		return nil, err
	}
	if before == 0 && after == 0 || len(task.baselineMeasured) == 1 {
		return nil, nil
	}
	declared := new(big.Int).Sub(task.baseline[1], task.baseline[0])
	if declared.Cmp(big.NewInt(int64(after-before))) == 0 {
		return nil, nil
	}
	return []string{fmt.Sprintf("task %s: **Baseline:** declares before=%s after=%s, but the changed files count @Test before=%d after=%d", task.id, task.baseline[0], task.baseline[1], before, after)}, nil
}

var tcfDigitsRE = regexp.MustCompile(`^[0-9]+$`)

// Moved from check-task-commit-fields.py's _run_measured_at (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Run the recorded command at <sha> inside a throwaway detached
// worktree of <worktree>'s repository, and return the single integer it
// prints on stdout — or None when anything about the environment makes
// the measurement unavailable: a temp worktree that cannot be created or
// removed, a command that exits non-zero, or a stdout that is not one
// integer. None is the skip path, never a verdict (the kan-100 rule the
// module docstring states). The worktree is created with `--detach`, so
// no branch moves and no checked-out tree is touched — the checked
// worktree's own checkout, the one a running conductor and the next
// implementer stand in (KAN-423), never changes state — and it is removed
// again in every path below. One run is bounded by
// MEASURED_TIMEOUT_SECONDS (panel F4): expiry lands in the same skip.
//
// tcfRunMeasuredAt is _run_measured_at: the recorded command's single
// integer at sha, run in a throwaway detached worktree; false is the skip.
// err is set only where Python raised past the function's own except. The
// throwaway worktree goes under env's TMPDIR (else /tmp), as Python's
// tempfile read it from the guard's own environment.
func tcfRunMeasuredAt(env Env, worktree, sha, command string, timeout time.Duration) (*big.Int, bool, error) {
	tmpdir := env.Getenv("TMPDIR")
	if tmpdir == "" {
		tmpdir = "/tmp"
	}
	tmp, err := os.MkdirTemp(tmpdir, "ctcf-measured-")
	if err != nil {
		return nil, false, tcfOSError("", err)
	}
	defer os.RemoveAll(tmp)
	if _, err := tcfRunGit(Env{}, worktree, "worktree", "add", "--detach", "--quiet", tmp, sha); err != nil {
		var crash tcfCrash
		if errors.As(err, &crash) {
			return nil, false, err
		}
		return nil, false, nil
	}
	defer exec.Command("git", "-C", worktree, "worktree", "remove", "--force", tmp).Run()

	out, errOut, code, timedOut, err := tcfRunBounded(tmp, timeout, "bash", "-c", command)
	if err != nil || timedOut {
		return nil, false, nil
	}
	so, _, err := tcfTextOutput(out, errOut)
	if err != nil {
		return nil, false, err
	}
	if code != 0 {
		return nil, false, nil
	}
	so = tcfStrip(so)
	if !tcfDigitsRE.MatchString(so) {
		return nil, false, nil
	}
	n, _ := new(big.Int).SetString(so, 10)
	return n, true, nil
}

// tcfRunBounded is subprocess.run(capture_output=True, timeout=...): both
// pipes are read to EOF and the child waited for inside the deadline; on
// expiry the child is killed and waited for, its pipes are not.
func tcfRunBounded(dir string, timeout time.Duration, name string, args ...string) (stdout, stderr []byte, code int, timedOut bool, err error) {
	outR, outW, err := os.Pipe()
	if err != nil {
		return nil, nil, 0, false, err
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		outR.Close()
		outW.Close()
		return nil, nil, 0, false, err
	}
	cmd := exec.Command(name, args...)
	cmd.Dir, cmd.Stdin, cmd.Stdout, cmd.Stderr = dir, os.Stdin, outW, errW
	err = cmd.Start()
	outW.Close()
	errW.Close()
	if err != nil {
		outR.Close()
		errR.Close()
		return nil, nil, 0, false, err
	}
	readAll := func(r *os.File) <-chan []byte {
		c := make(chan []byte, 1)
		go func() {
			b, _ := io.ReadAll(r)
			r.Close()
			c <- b
		}()
		return c
	}
	outC, errC := readAll(outR), readAll(errR)
	waitC := make(chan struct{})
	go func() {
		cmd.Wait()
		close(waitC)
	}()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for outC != nil || errC != nil || waitC != nil {
		select {
		case b := <-outC:
			stdout, outC = b, nil
		case b := <-errC:
			stderr, errC = b, nil
		case <-waitC:
			waitC = nil
		case <-deadline.C:
			cmd.Process.Kill()
			if waitC != nil {
				<-waitC
			}
			return nil, nil, 0, true, nil
		}
	}
	return stdout, stderr, cmd.ProcessState.ExitCode(), false, nil
}

// Moved from check-task-commit-fields.py's check_baseline_measured (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// The Baseline field's one recorded measurement command, re-run at the
// commit's parent (`before`) and at the commit (`after`), must measure
// exactly the declared counts — the dynamic counterpart to
// check_baseline_counts' static @Test delta, and the mechanical detector
// for the wrong delta and wrong predictions that run's reviewers caught
// only by hand (KAN-409). Skips — never fails — whenever the measurement
// cannot be taken (see _run_measured_at), when the task declares no
// Baseline, and when the field records no command or more than one
// DISTINCT command: which command would be THE task's measurement is then
// unknowable, and guessing one would check the declaration against
// nothing.
//
// tcfCheckBaselineMeasured is check_baseline_measured.
func tcfCheckBaselineMeasured(env Env, task tcfTask, worktree, parent, commit string, timeout time.Duration) ([]string, error) {
	if task.baseline == nil || len(task.baselineMeasured) != 1 {
		return nil, nil
	}
	var measured [2]*big.Int
	for i, sha := range []string{parent, commit} {
		n, ok, err := tcfRunMeasuredAt(env, worktree, sha, task.baselineMeasured[0], timeout)
		if err != nil || !ok {
			return nil, err
		}
		measured[i] = n
	}
	if measured[0].Cmp(task.baseline[0]) == 0 && measured[1].Cmp(task.baseline[1]) == 0 {
		return nil, nil
	}
	return []string{fmt.Sprintf("task %s: **Baseline:** declares before=%s after=%s, but the recorded measurement command measures before=%s after=%s", task.id, task.baseline[0], task.baseline[1], measured[0], measured[1])}, nil
}

var tcfFoldWSRE = regexp.MustCompile(`[\t\n\v\f\r ]+`)

// Moved from check-task-commit-fields.py's _fold_ws_bytes (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// The one fold rule both `Tests:` match sites read — every whitespace
// run, line breaks included, becomes a single space (KAN-562) — stated
// once and shared by the folded tree text and the folded declared name.
//
// tcfFoldWS is _fold_ws_bytes.
func tcfFoldWS(b []byte) []byte { return tcfFoldWSRE.ReplaceAll(b, []byte(" ")) }

// Moved from check-task-commit-fields.py's _folded_tree_text (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// The tree's file CONTENT at <commit_sha>, minus <exclude_dir> and
// everything under it and minus <exclude_path> when given, as one
// whitespace-folded byte string. `git ls-tree -r -z` lists the blobs and
// one `git cat-file --batch` streams their contents — neither applies
// export attributes, so the search sees every blob the `git grep -F` it
// replaced saw (panel F1: `git archive` honors export-ignore and
// export-subst, which silently narrowed the read). The path filter lives
// here rather than in git pathspecs because `git ls-tree` supports no
// exclude magic. Blobs join on NUL, which no fold touches, so a name
// never matches across two files' boundary; every whitespace run, line
// breaks included, folds to a single space. An empty listing folds to an
// empty string, the same no-match answer `git grep` gave on such a
// tree.
//
// tcfFoldedTreeText is _folded_tree_text.
func tcfFoldedTreeText(env Env, worktree, commit, excludeDir, excludePath string) ([]byte, error) {
	listed, err := tcfRunGit(env, worktree, "ls-tree", "-r", "-z", commit)
	if err != nil {
		return nil, err
	}
	var shas []string
	for _, entry := range strings.Split(listed, "\x00") {
		meta, path, ok := strings.Cut(entry, "\t")
		if !ok {
			continue
		}
		parts := strings.FieldsFunc(meta, tcfIsSpace)
		if len(parts) != 3 || parts[1] != "blob" || path == excludeDir || strings.HasPrefix(path, excludeDir+"/") || excludePath != "" && path == excludePath {
			continue
		}
		shas = append(shas, parts[2])
	}
	var folded []byte
	if len(shas) > 0 {
		buf, errOut, code, err := tcfExec(env.Dir, []byte(strings.Join(shas, "\n")+"\n"), "git", "-C", worktree, "cat-file", "--batch")
		if err != nil {
			return nil, err
		}
		if code != 0 {
			return nil, fmt.Errorf("git cat-file --batch at %s failed with exit %d: %s", commit, code, tcfStrip(strings.ToValidUTF8(string(errOut), "�")))
		}
		for i := 0; i < len(buf); {
			nl := bytes.IndexByte(buf[i:], '\n')
			if nl < 0 {
				break
			}
			nl += i
			head := bytes.Fields(buf[i:nl])
			if len(head) != 3 { // "<sha> missing"
				i = nl + 1
				continue
			}
			size, _ := strconv.Atoi(string(head[2]))
			end := min(nl+1+size, len(buf))
			folded = append(append(folded, 0), buf[nl+1:end]...)
			i = nl + 1 + size + 1
		}
	}
	return tcfFoldWS(folded), nil
}

// Moved from check-task-commit-fields.py's check_tests_in_tree (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Every backticked or bare-camelCase name in `Tests:` must appear in
// the tree's CONTENT at the commit. The diff check passes a test the
// commit removes — the removal hunk carries the name — so the tree is
// what catches the stale declaration. The search is whitespace-folded on
// both sides (KAN-562): the declared name and the tree text each have
// their whitespace runs folded to one space before the fixed-string
// comparison, so a sentence the tree wraps across a source-line break
// matches the words it names. The plan file itself is excluded from the
// search: its own `**Tests:**` line declares the name, and without
// the exclusion every declared name would trivially match itself. A
// satellite's plan lives OUTSIDE the worktree searched (the change-plan
// resolution reads it from the canonical repository), and a path outside
// the worktree names nothing in its listing — there the plan cannot
// self-match anyway, so no exclusion is passed. Every other plan under
// `spectre/changes/` — archived changes included — is excluded too
// (panel F2): a stale name surviving in another change's `**Tests:**`
// line must not vouch for itself, or the guard passes the exact stale
// declaration it exists to catch.
//
// Inside the function:
//
// Content search missed; the name may still be a real committed
// PATH (panel F1) — a `Tests:` token naming a file the commit
// carries, whose string appears nowhere as file content.
//
// tcfCheckTestsInTree is check_tests_in_tree.
func tcfCheckTestsInTree(env Env, task tcfTask, worktree, commit, tasksAbs string) ([]string, error) {
	wtAbs, _ := filepath.Abs(worktree)
	planRel, _ := filepath.Rel(wtAbs, tasksAbs)
	if strings.HasPrefix(planRel, "../") {
		planRel = ""
	}
	tree, err := tcfFoldedTreeText(env, worktree, commit, "spectre/changes", planRel)
	if err != nil {
		return nil, err
	}
	var v []string
	for _, name := range tcfTreeNames(task.testsValue) {
		if bytes.Contains(tree, tcfFoldWS([]byte(name))) {
			continue
		}
		out, errOut, code, err := tcfExec(env.Dir, nil, "git", "-C", worktree, "cat-file", "-e", commit+":"+name)
		if err != nil {
			return nil, err
		}
		if _, _, err := tcfTextOutput(out, errOut); err != nil {
			return nil, err
		}
		if code == 0 {
			continue
		}
		v = append(v, fmt.Sprintf("task %s: declared test %s not found in the tree at %s — %s", task.id, name, commit, tcfTestsParseRule))
	}
	return v, nil
}

// tcfCheckCommitSubject is check_commit_subject.
func tcfCheckCommitSubject(task tcfTask, actual string) []string {
	if task.commit == nil || *task.commit == actual {
		return nil
	}
	return []string{fmt.Sprintf("task %s: commit subject %s does not match declared Commit: %s", task.id, tcfRepr(actual), tcfRepr(*task.commit))}
}

// Moved from check-task-commit-fields.py's SUBJECT_SCOPE_RE (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// check_commit_scope's grammar: a declared Commit: subject is
// `<type>(<scope>): <rest>` or the Conventional Commits breaking-change form
// `<type>(<scope>)!: <rest>`; a subject with no parenthesised scope before
// the first colon (or `!:`) has no scope at all (SUBJECT_SCOPE_RE then does
// not match). The optional `!` (pass 2, finding A) must sit between the
// closing paren and the colon, or a subject like
// `feat(kan-202-commit-split-and-module-scopes)!: add alpha` bypassed the
// scope check entirely — a real, valid Conventional Commits shape this
// guard has to see. TASK_ID_SCOPE_RE reuses DOTTED_ID rather than the
// narrower task-id grammar, deliberately: a scope written `1.2` names no
// task any more and is still exactly the mistake this refuses. A scope that is
// nothing but digits and dots is a task id, never a module.
//
// Moved from check-task-commit-fields.py's _LEADING_KEY_RE (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// _LEADING_KEY_RE: a Jira key is `<letters>-<digits>` — never digits alone —
// at the very start of the change name (per skills/flow-contracts/
// jira-integration.md's own `[A-Z]{2,10}-\d+` shape and this repository's
// "Change naming" convention, `<key>-<slug>`; matched case-insensitively
// here since check_commit_scope's own comparison is, per finding B).
var (
	tcfSubjectScopeRE = regexp.MustCompile(`^[A-Za-z]+\(([^)]*)\)!?:`)
	tcfTaskIDScopeRE  = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)*$`)
	tcfLeadingKeyRE   = regexp.MustCompile(`^[A-Za-z]+-[0-9]+`)
)

// Moved from check-task-commit-fields.py's _leading_jira_key (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// The change name's leading Jira key, e.g. `kan-202` from
// `kan-202-commit-split-and-module-scopes`. A change name that does not
// begin with `<letters>-<digits>` at all has no key and this returns
// `None` (pass 2, finding C — the prior implementation returned the name
// up through its first ALL-DIGIT hyphen segment, which both over-matched
// a plain `<word>-<year>`-shaped module name and, worse, could not tell
// that shape apart from a real key at all).
//
// A leading `<letters>-<digits>` match is itself ambiguous, not a key,
// when the text immediately following it is ANOTHER `<letters>-<digits>`
// pair — e.g. `release-2026-kan-450-cleanup`, where `release-2026` and
// `kan-450` are equally plausible candidates and nothing here can tell
// which one (if either) is the real Jira key; per the same "does not
// begin with one has no key" rule, this returns `None` rather than guess.
// A single unambiguous match, e.g. `kan-202` in
// `kan-202-commit-split-and-module-scopes` (followed by `commit-split`,
// which is not itself a `<letters>-<digits>` pair), is returned as-is.
//
// tcfLeadingJiraKey is _leading_jira_key; "" is None.
func tcfLeadingJiraKey(change string) string {
	key := tcfLeadingKeyRE.FindString(change)
	if rest := change[len(key):]; key == "" || strings.HasPrefix(rest, "-") && tcfLeadingKeyRE.MatchString(rest[1:]) {
		return ""
	}
	return key
}

// Moved from check-task-commit-fields.py's check_commit_scope (KAN-760 deleted the
// Python copy once this port replaced it; panel round 0, F9):
//
// Fail when the **declared** `Commit:` field's scope — parsed out of
// `task.commit`, never the real commit's subject, per flow-commit-
// scope's requirement — equals the change name, the change name's bare
// Jira key, or a dotted/numeric task id. A field with no scope, or any
// other scope, passes: there is no vocabulary of legal module names to
// maintain.
//
// The change-name and Jira-key comparisons are case-insensitive (pass 2,
// finding B): a declared scope of `feat(KAN-202): ...` against change
// `kan-202-...` is the shape a human is *most* likely to type, since Jira
// keys are conventionally written uppercase, and a case-sensitive compare
// let it straight through. The task-id comparison is deliberately left
// case-sensitive: `TASK_ID_SCOPE_RE` only ever matches a scope that is
// digits and dots (`^\d+(?:\.\d+)*$`), which carries no letters at all,
// so case does not apply to it.
//
// tcfCheckCommitScope is check_commit_scope.
func tcfCheckCommitScope(task tcfTask, change string) []string {
	if task.commit == nil {
		return nil
	}
	m := tcfSubjectScopeRE.FindStringSubmatch(*task.commit)
	if m == nil || m[1] == "" {
		return nil
	}
	scope, lower := m[1], strings.ToLower(m[1])
	switch key := tcfLeadingJiraKey(change); {
	case lower == strings.ToLower(change):
		return []string{fmt.Sprintf("task %s: declared Commit: scope %s names the change, not a module", task.id, tcfRepr(scope))}
	case key != "" && lower == strings.ToLower(key):
		return []string{fmt.Sprintf("task %s: declared Commit: scope %s names the change's Jira key, not a module", task.id, tcfRepr(scope))}
	case tcfTaskIDScopeRE.MatchString(scope):
		return []string{fmt.Sprintf("task %s: declared Commit: scope %s is a task id, not a module", task.id, tcfRepr(scope))}
	}
	return nil
}

// The evidence-rule tag shapes this guard refuses at task close — all four
// tags' payloads, the same set the plan-provenance guard tests, so a tag
// cannot pass one boundary and fail the other. The fence tag is anchored
// like check-plan-provenance.py's FENCE_TAG_RE — start or whitespace before
// the colon word, so `preverified:` cannot satisfy it as a mere substring —
// and is searched on the INFO STRING, the line after lib/plan_grammar.py's
// fence run is stripped, never on the raw line: a language-less fence
// (```verified:) would otherwise hide its tag behind the backtick run the
// (^|\s) anchor reads as neither. The comment regex is that guard's
// PROVENANCE_RE with both comment tags. Payloads are tested for evidence by
// the function below, never by these patterns.
var (
	tcfEvidenceFenceTagRE = regexp.MustCompile(`(^|[` + tcfWS + `])(un)?verified:`)
	tcfEvidenceMeasuredRE = regexp.MustCompile(`<!--[` + tcfWS + `]*(measured|predicted):`)
)

// tcfCheckEvidenceTags is check_evidence_tags.
//
// The evidence rule, enforced on the closing task's own record: any
// provenance tag in the task's text — a `verified:`/`unverified:` on a
// fence info string, a `measured:`/`predicted:` comment in its prose —
// must name what its claim rests on: the command, the source URL, or
// the output that shows the check ran, or what confirms or would
// confirm it. A tag naming nothing after the colon asserts a status
// while carrying no evidence, which
// skills/flow-contracts/plan-provenance.md ("The evidence is part of
// the tag") makes a violation rather than an omission: it is worse
// than the honest untagged shape, because the reader trusts it.
//
// The tag set is the plan-provenance guard's own, so a tag cannot pass
// one boundary and fail the other; each guard still reports at its own
// boundary, so one defective tag costs one finding from each. Fence
// CONTENT is never scanned — the toggle is lib/plan_grammar.py's fence
// rule (tcfFenceRE), the same one `parse_task_fields` uses, so a
// comment-shaped line inside a worked example stays code — and the
// fence's own OPENING line is where an info-string tag is looked for, a
// closing line carrying no info string to read.
func tcfCheckEvidenceTags(body tcfTaskBody) []string {
	var v []string
	inFence := false
	for offset, line := range body.lines {
		// body.taskLine is the 1-based line of the task line, so the file
		// line number of body.lines[offset] is taskLine + offset + 1.
		lineno := body.taskLine + offset + 1
		if fence := tcfFenceRE.FindStringIndex(line); fence != nil {
			if !inFence {
				info := line[fence[1]:]
				if m := tcfEvidenceFenceTagRE.FindStringIndex(info); m != nil && tcfStrip(info[m[1]:]) == "" {
					tag := tcfStrip(info[m[0]:m[1]])
					v = append(v, fmt.Sprintf("task %s: %s on the fence at tasks.md line %d carries no evidence after the colon — write the command, source URL or output that shows the check ran, or what to confirm before trusting the block", body.id, tag, lineno))
				}
			}
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if m := tcfEvidenceMeasuredRE.FindStringSubmatchIndex(line); m != nil {
			closer := strings.Index(line[m[1]:], "-->")
			if closer != -1 && tcfStrip(line[m[1]:m[1]+closer]) == "" {
				v = append(v, fmt.Sprintf("task %s: %s: comment at tasks.md line %d carries no evidence after the colon — write the command and ref that were run, or what would confirm the claim", body.id, line[m[2]:m[3]], lineno))
			}
		}
	}
	return v
}

// ---- Python builtins the checks lean on ----

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
// backslash is a literal.
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
			b.WriteString(".*")
			for i < n && p[i] == '*' {
				i++
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
