package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// check-plan-shape checks that a tasks.md is SHAPED so the guard that later
// judges its commits (check-task-commit-fields) can actually read it. Sixteen
// findings, F1-F16 (F1-F6's canonical definitions, decisions and rationale:
// spectre/changes/kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan/
// design.md -- not restated here; F7-F16's are the rows below).
//
// The guard reads the plan with the very grammar the commit guard reads it
// with -- tcfParseTask, tcfTaskIDs, tcfIterTasks, tcfSelectTask,
// tcfUnclosedFence and tcfSelectAfter in planparse.go -- never a copy:
// parser identity is structural (one function), never asserted.
//
// Arguments: one tasks.md path checks that file; none scans every live
// change's <spec-root>/changes/*/tasks.md under the repository root, plus
// that of every archived change under changes/archive/ the base does not
// carry yet (unlandedArchives) -- an archived change archived on this
// branch by integrate's run 1, then written into by a fix run, is still
// this branch's work; every archive the base carries stays out. The root
// is CHECK_PLAN_SHAPE_ROOT when set (a harness's sandboxed fixture tree),
// else FLOW_GUARD_REPO_ROOT, which scripts/check-plan-shape.sh exports from
// its own real location -- so a run through the shipped
// skills/flow/scripts/ symlink judges the repository, never skills/flow.
//
// Per file, the F1-F16 checks run, then plan-suite-steps
// (plansuitesteps.go: a task step running a module or repository suite),
// plan-script-paths (planscriptpaths.go: a step citing a scripts/ path the
// repository does not carry) and plan-task-id-rule (plantaskidrule.go: a
// /flow plan stating task commits carry no Task-Id: trailer), in-process
// and in that order, each printing its own lines.
//
// Exit codes, per file and over the scan: 2 if any check cannot answer
// (wrong argument count, a file that cannot be read -- a plan that is not
// UTF-8 included, where check-plan-shape.py crashed with exit 1 and its
// header named exit 2; the header wins -- or an archive listing that
// fails), else 1 if any found a violation, else 0. Violations print to
// stdout one per line as `file:line: message`, naming the task id (F5
// excepted: a file with no tasks has no task id to name).
//
// Findings -- every one reported as file:line, fences excluded from every
// scan (tcfFenceRE): a field-shaped line inside a worked example is not a
// declaration.
//
//	F1  A second Files, Tests, Commit, Baseline or Allowed-collateral line
//	    in one task body (psF1Fields) -- naming which occurrence wins, since
//	    the parser keeps the last.
//	F2  No Files line, or one whose backtick token list is empty -- read
//	    off tcfParseTask's own files, which is empty in exactly these two
//	    cases (a value with prose and no backticks falls back to that prose
//	    as one path, so F2 does not fire there; F16 catches that fallback's
//	    multi-path shape).
//	F3  a. a field-shaped line indented past column 0, invisible to
//	       tcfFieldRE (anchored at column 0) -- psIndentedFieldRE, derived
//	       from tcfFieldRE's own pattern, never a second name list;
//	    b. a task body whose fence never closes (tcfParseTask's
//	       unclosedFenceLine). When b fires, every other per-task check is
//	       skipped for that task: every field below the fence is code, and
//	       reporting those as empty or duplicated would name the consequence
//	       and hide the cause.
//	F4  A Tests value opening with `none` (tcfNoneOpenRE) on a task whose
//	    Files names a test-shaped path (psTestShaped) -- a contradiction
//	    between the opt-out and the declared files. The test-repair
//	    annotation (tcfRepairsOpenRE, `none — repairs existing tests:`) is
//	    exempt (KAN-794): the names after the colon are what the task
//	    repairs, not a declaration it abandons.
//	F5  A tasks.md the grammar finds zero tasks in -- a vacuous pass, not a
//	    clean file. Reported at line 1.
//	F6  A Tests value that is neither none-opening, nor carries a `Case N`
//	    label, nor a backtick token -- it declares nothing, so the commit
//	    guard verifies nothing. Read as tcfParseTask's tests being empty
//	    together with the same none check, so the two are never reported
//	    twice under two findings.
//	F7  A second GATING After line in one task body, after the first gating
//	    one -- tcfSelectAfter keeps the first; a non-gating line ahead of it
//	    is nobody's duplicate.
//	F8  A task whose first After candidate does not gate (neither
//	    `Task <ids>` nor `none`), reported with the offending value -- the
//	    shape plan-dispatch-bundles reports as its malformed-value finding,
//	    read from the same selector.
//	F9  An id in a task's GATING After set that names no task in the plan.
//	F10 A cycle in the RESOLVED After graph -- declared ids; every
//	    plan-order earlier id for a task carrying no field (the serial
//	    default, checked tasks included); nothing for `none` -- walked DFS
//	    from each task in plan order. The FIRST back-edge is reported, as
//	    `task <a> -> task <b> -> task <a>`, and the search stops: every
//	    further cycle names the same defect. A self-reference is the
//	    one-node cycle. A task whose fence never closes is skipped for
//	    F7-F11; its field reads as absent, the serial default, exactly as
//	    plan-dispatch-bundles resolves it.
//	F11 A file named in more than one task's Files with no GATING declared
//	    After edge between a co-owning pair, in either direction -- the
//	    shared-file conflict KAN-339's self-review recorded. The serial
//	    default notes nothing and `none` declares independence, not an
//	    order. Reported once per unordered pair at the LATER task's Files
//	    line; a task whose fence never closes is no owner.
//	F12 A plan whose tasks declare Baseline while no non-fenced line opens
//	    the blockquote `> **Baseline convention:**` (psBaselineConventionRE)
//	    -- the plan states no counting convention, so later tasks drift
//	    between conventions (KAN-678). Reported once, at the first declaring
//	    task's Baseline line; the guard checks that ONE is stated, never
//	    which (skills/flow/brainstorm-planner.md seeds the line).
//	F13 A non-fenced `<!-- measured: … -->` line whose command counts `@Test`
//	    with `grep -c` or `grep -o` and no `-w` -- the pattern also matches
//	    `@Testcontainers` and `@TestConfiguration` (KAN-936). The rule is the
//	    smallest one: a grep whose short-option clusters carry `c` or `o`
//	    (or that passes `--count`) and none carries `w` (nor
//	    `--word-regexp`), with a pattern opening `@Test` and no letter or
//	    backslash after it.
//	F14 A task whose Files all lie under the planning paths (psPlanning):
//	    check-task-commit-planning-paths refuses a task commit there, so the
//	    task lands as a planning commit and is ticked by hand (kan-941's
//	    Task 4). Reported at the task line.
//	F15 A backticked Tests name that a test-shaped file outside the task's
//	    Files (and its Allowed-collateral globs), in a directory one of
//	    those files lies in, already declares at the change's base: a
//	    func/fun/def of exactly that name, or one starting with it in the
//	    file whose subject the name is (`TestFoldFixup` and
//	    fold_fixup_test.go). The task's commit cannot touch that file, so
//	    check-task-commit-fields never finds the name (kan-918's Task 3).
//	    Another directory is never read -- Go allows one name per package,
//	    and TestMain is in many. The base is the merge base of HEAD with
//	    branch.<current>.flowBase (its origin/ copy preferred) or else with
//	    refs/remotes/origin/HEAD; a plan outside a git repository, or one
//	    whose base does not resolve, is not checked. Reported at the Tests
//	    line.
//	F16 A Files or Allowed-collateral value (the joined field) carrying no
//	    backticked token, not opening with `none`, and containing a comma or
//	    whitespace: the commit guard's prose fallback reads it as ONE path
//	    (kan-919's self-review). Reported at the field's line.
//
// A duplicate task id is check-task-build-green's violation, not this
// guard's: tcfTaskIDs keeps the first occurrence, and every per-task read
// resolves to it.
func init() { Registry["check-plan-shape"] = checkPlanShape }

func checkPlanShape(args []string, env Env, stdout, stderr io.Writer) int {
	switch len(args) {
	case 1:
		return psCheckOne(args[0], env, "", stdout, stderr)
	case 0:
	default:
		fmt.Fprintln(stderr, "usage: check-plan-shape.sh [path-to-tasks.md]")
		return 2
	}
	root := env.Getenv("CHECK_PLAN_SHAPE_ROOT")
	if root == "" {
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintln(stderr, "check-plan-shape: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-plan-shape.sh, which sets it")
			return 2
		}
	}
	changes := root + "/" + specRootLeaf(psAbs(env, root), stderr) + "/changes"
	if !isDir(psAbs(env, changes)) {
		return 0
	}
	var archived bytes.Buffer
	if unlandedArchivesCmd([]string{root}, env, &archived, stderr) != 0 {
		return 2
	}
	var files []string
	for _, name := range tcfEntries(psAbs(env, changes)) {
		files = append(files, changes+"/"+name+"/tasks.md")
	}
	for _, name := range strings.Split(archived.String(), "\n") {
		if name != "" {
			files = append(files, changes+"/archive/"+name+"/tasks.md")
		}
	}
	status := 0
	for _, f := range files {
		if _, err := os.Stat(psAbs(env, f)); err != nil {
			continue
		}
		status = psWorse(status, psCheckOne(f, env, "", stdout, stderr))
	}
	return status
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

func psAbs(env Env, p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(env.Dir, p)
}

// psCheckOne checks one file: F1-F16, then the three plan sub-guards. off
// names one finding whose decision is disabled -- the harness's mutation
// proof, which sed-disabled that decision's `if` in a copy of the guard;
// production passes "".
func psCheckOne(path string, env Env, off string, stdout, stderr io.Writer) int {
	rc := psShape(path, env, off, stdout, stderr)
	for _, name := range []string{"plan-suite-steps", "plan-script-paths", "plan-task-id-rule"} {
		rc = psWorse(rc, Registry[name]([]string{path}, env, stdout, stderr))
	}
	return rc
}

func psShape(path string, env Env, off string, stdout, stderr io.Writer) int {
	b, err := os.ReadFile(psAbs(env, path))
	var lines []string
	if err == nil {
		lines, err = tcfPlanLines(b)
	} else if pe := (*fs.PathError)(nil); errors.As(err, &pe) {
		pe.Path = path // str(OSError) names the path as given
		err = tcfOSError(path, pe)
	}
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", path, err)
		return 2
	}
	p := &psPlan{path: path, lines: lines, off: off, env: env}
	violations := p.check()
	for _, v := range violations {
		fmt.Fprintln(stdout, v)
	}
	if len(violations) > 0 {
		return 1
	}
	return 0
}

type psPlan struct {
	path  string
	lines []string
	off   string
	env   Env
}

// on is false only for the finding the mutation proof disabled.
func (p *psPlan) on(tag string) bool { return p.off != tag }

// parse is check-plan-shape.py's parse_task_fields: tcfParseTask with
// Python's own files list (tcfPythonFiles).
func (p *psPlan) parse(id string) tcfTask {
	t, _ := tcfParseTask(p.lines, id)
	t.files = tcfPythonFiles(p.lines, t)
	return t
}

func (p *psPlan) at(line int, msg string) string { return fmt.Sprintf("%s:%d: %s", p.path, line, msg) }

// psF1Fields are the five names F1 counts; Regression and Build are not
// part of it -- Build's precedence is tcfSelectBuildKind's job.
var psF1Fields = []string{"Files", "Tests", "Commit", "Baseline", "Allowed-collateral"}

// psTestShaped are F4's (and F15's) test-shaped file globs; psPlanning
// F14's planning paths -- check-task-commit-planning-paths' <leaf>/changes/
// for both spec leaves, and docs/superpowers/.
var (
	psTestShaped = []string{"test-*", "*_test.go", "*.test.*", "*.spec.ts"}
	psPlanning   = []string{"spectre/changes/", "openspec/changes/", "docs/superpowers/"}
)

var (
	// psIndentedFieldRE is tcfFieldRE indented past column 0 (F3a), built
	// from tcfFieldRE's own pattern so the two never drift on the names.
	psIndentedFieldRE = regexp.MustCompile(`^[ \t]+` + strings.TrimPrefix(tcfFieldRE.String(), "^"))
	// psBaselineConventionRE is F12's stated convention line.
	psBaselineConventionRE = regexp.MustCompile(`^>[` + tcfWS + `]*\*\*Baseline convention:\*\*`)
	// psGrepTestRE is F13's `grep [opts] '@Test'`: group 1 the option
	// tokens. Python's \b before grep and its lookahead after @Test (no
	// letter or backslash) are checked in psCountsTestUnsafely.
	psGrepTestRE = regexp.MustCompile(`grep((?:[` + tcfWS + `]+--?[A-Za-z][^` + tcfWS + `]*)*)[` + tcfWS + `]+['"]?@Test`)
	psShortRE    = regexp.MustCompile(`^[A-Za-z]*`)
	psAnyWSRE    = regexp.MustCompile(`[,` + tcfWS + `]`)
	// psDeclRE is F15's test declaration (Go func, Kotlin fun, Python def)
	// and its name; group 1 is the boundary Python's \b asserted.
	psDeclRE  = regexp.MustCompile(`(^|[^\p{L}\p{N}_])(?:func|fun|def)[` + tcfWS + `]+([A-Za-z_][\p{L}\p{N}_]*)`)
	psIdentRE = regexp.MustCompile(`^[A-Za-z_][\p{L}\p{N}_]*$`)
	psTestPfx = regexp.MustCompile(`(?i)^test_?`)
	psNonKey  = regexp.MustCompile(`[^a-z0-9]`)
)

func psWord(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' }

// psCountsTestUnsafely is F13's rule on one line. A short cluster's flags
// end at its first non-letter (`-m1`): what follows is a value; a long
// option is matched whole.
func psCountsTestUnsafely(line string) bool {
	if !strings.Contains(line, "measured:") {
		return false
	}
	for start := 0; start < len(line); {
		loc := psGrepTestRE.FindStringSubmatchIndex(line[start:])
		if loc == nil {
			return false
		}
		ms, me := start+loc[0], start+loc[1]
		before, _ := utf8.DecodeLastRuneInString(line[:ms])
		after, _ := utf8.DecodeRuneInString(line[me:])
		if (ms > 0 && psWord(before)) || (me < len(line) && (after < 0x80 && unicode.IsLetter(after) || after == '\\')) {
			start = ms + 1
			continue
		}
		opts := strings.FieldsFunc(line[start+loc[2]:start+loc[3]], tcfIsSpace)
		short := ""
		for _, o := range opts {
			if !strings.HasPrefix(o, "--") {
				short += psShortRE.FindString(o[1:])
			}
		}
		counts := strings.ContainsAny(short, "co") || slices.Contains(opts, "--count")
		if counts && !strings.Contains(short, "w") && !slices.Contains(opts, "--word-regexp") {
			return true
		}
		start = me
	}
	return false
}

func psIsTestShaped(path string) bool {
	name := path[strings.LastIndex(path, "/")+1:]
	return slices.ContainsFunc(psTestShaped, func(g string) bool { return tcfFnmatch(name, g) })
}

// psFieldLine is the 1-based file line of the first non-fenced line in body
// whose tcfFieldRE match names name, or 0; start is the body's first index.
func psFieldLine(body []string, start int, name string) int {
	line := 0
	tcfUnfenced(body, func(off int, l string) bool {
		if m := tcfFieldRE.FindStringSubmatch(l); m != nil && m[1] == name {
			line = start + off + 1
			return true
		}
		return false
	})
	return line
}

func (p *psPlan) check() []string {
	ids := tcfTaskIDs(p.lines)
	if len(ids) == 0 && p.on("F5") {
		return []string{p.at(1, "this plan defines zero tasks — the grammar finds no `- [ ] <n>. ` checkbox line, so every check that follows would be a vacuous pass rather than a real one")}
	}
	var v []string
	for _, id := range ids {
		v = append(v, p.checkTask(*tcfSelectTask(p.lines, id))...)
	}
	v = append(v, p.afterReferences()...)
	v = append(v, p.fileOwnership()...)
	v = append(v, p.baselineConvention()...)
	return append(v, p.baseDeclared(ids)...)
}

func (p *psPlan) checkTask(found tcfTaskBody) []string {
	id, body, start := found.id, found.lines, found.taskLine
	fields := p.parse(id)
	if fields.unclosedFenceLine != 0 && p.on("F3b") {
		return []string{p.at(fields.unclosedFenceLine, "task "+id+" opens a code fence here that is never closed in its body, so every field below it is inside the fence and was not read")}
	}
	var v []string
	seen := map[string]int{}
	tcfUnfenced(body, func(off int, line string) bool {
		m := tcfFieldRE.FindStringSubmatch(line)
		if m == nil || !slices.Contains(psF1Fields, m[1]) {
			return false
		}
		if first, ok := seen[m[1]]; ok && p.on("F1") {
			v = append(v, p.at(start+off+1, fmt.Sprintf("task %s declares a second **%s:** line (first at line %d); the parser keeps the last occurrence and silently discards this one", id, m[1], first)))
		} else {
			seen[m[1]] = start + off + 1
		}
		return false
	})
	tcfUnfenced(body, func(off int, line string) bool {
		if m := psIndentedFieldRE.FindStringSubmatch(line); m != nil && p.on("F3a") {
			v = append(v, p.at(start+off+1, "task "+id+" carries a **"+m[1]+":** line indented past column 0, invisible to the field grammar (FIELD_RE anchors at column 0) and therefore never read as a declaration"))
		}
		return false
	})
	firstGating := 0
	tcfUnfenced(body, func(off int, line string) bool {
		m := tcfAfterFieldRE.FindStringSubmatch(line)
		if m == nil {
			return false
		}
		if _, ok := tcfAfterIDs(tcfStrip(m[1])); !ok {
			return false
		}
		if firstGating != 0 && p.on("F7") {
			v = append(v, p.at(start+off+1, fmt.Sprintf("task %s declares a second **After:** line (first at line %d); select_after keeps the first", id, firstGating)))
		} else {
			firstGating = start + off + 1
		}
		return false
	})
	tcfUnfenced(body, func(off int, line string) bool {
		if psCountsTestUnsafely(line) && p.on("F13") {
			v = append(v, p.at(start+off+1, "task "+id+"'s measured: command counts `@Test` with `grep -c` or `grep -o` and no `-w`, so `@Testcontainers` and `@TestConfiguration` lines count too — use a word-boundary-safe counter such as `grep -cw '@Test'` or `grep -ow '@Test' | wc -l` (the **Baseline convention:** line)"))
		}
		return false
	})

	// F8 -- read from tcfSelectAfter itself, never a second loop deciding
	// which line the field is; the refusal names the accepted values so
	// the first rewrite is informed (KAN-605).
	if a := tcfSelectAfter(body); a != nil && !a.gates && p.on("F8") {
		v = append(v, p.at(start+a.offset+1, "task "+id+" has a malformed **After:** value: "+a.value+" — an **After:** value is `Task <ids>` or `none`"))
	}
	if len(fields.files) == 0 && p.on("F2") {
		v = append(v, p.at(found.taskLine, "task "+id+" has no **Files:** line, or one naming no file at all"))
	}
	planning := func(f string) bool {
		return slices.ContainsFunc(psPlanning, func(pfx string) bool { return strings.HasPrefix(f, pfx) })
	}
	if len(fields.files) > 0 && !slices.ContainsFunc(fields.files, func(f string) bool { return !planning(f) }) && p.on("F14") {
		v = append(v, p.at(found.taskLine, "task "+id+"'s **Files:** all lie under the planning paths (spectre/changes/, openspec/changes/, docs/superpowers/), so its commit cannot carry a `Task-Id:` — name the file the task's work lands in, or make the record a planning commit with no task"))
	}
	joined := tcfBodyFields(body)
	for _, name := range []string{"Files", "Allowed-collateral"} {
		value := strings.Join(joined[name], " ")
		if value != "" && !tcfBacktickRE.MatchString(value) && !tcfNoneOpenRE.MatchString(value) && psAnyWSRE.MatchString(value) && p.on("F16") {
			line := psFieldLine(body, start, name)
			if line == 0 {
				line = found.taskLine
			}
			v = append(v, p.at(line, "task "+id+"'s **"+name+":** value names no backticked path but carries a comma or whitespace — the commit guard reads it as one path; backtick each path"))
		}
	}

	// F4 and F6 read the continuation-joined Tests value tcfParseTask
	// itself reads, never the field's first physical line: a `none` on
	// the line after **Tests:** is invisible to the latter (KAN-121
	// panel-fix round, finding 1; the repair form likewise, KAN-794). The
	// field's line only anchors the message.
	testsLine := psFieldLine(body, start, "Tests")
	noneOpen := tcfNoneOpenRE.MatchString(fields.testsValue)
	if noneOpen && !tcfRepairsOpenRE.MatchString(fields.testsValue) && slices.ContainsFunc(fields.files, psIsTestShaped) && p.on("F4") {
		v = append(v, p.at(testsLine, "task "+id+"'s **Tests:** field opens with `none` but **Files:** names a test-shaped path — a contradiction between the opt-out and the declared files; a task repairing existing tests and adding none says so with `Tests: none — repairs existing tests: <names>` (KAN-794)"))
	}
	if !noneOpen && len(fields.tests) == 0 && p.on("F6") {
		line := testsLine
		if line == 0 {
			line = found.taskLine
		}
		v = append(v, p.at(line, "task "+id+"'s **Tests:** field declares nothing — it opens with neither `none`, a `Case N` label, nor a backtick-quoted test name, so nothing is verified"))
	}
	return v
}

// baselineConvention is F12: the declaring task is the first whose
// non-fenced body carries a Baseline field line; the stated convention is
// any non-fenced line of the file opening as psBaselineConventionRE.
func (p *psPlan) baselineConvention() []string {
	declaring, line := "", 0
	for _, task := range tcfIterTasks(p.lines) {
		if tcfUnclosedFence(task.lines) >= 0 {
			continue // F3b: this task's fields are unread
		}
		if l := psFieldLine(task.lines, task.taskLine, "Baseline"); l != 0 {
			declaring, line = task.id, l
			break
		}
	}
	if declaring == "" {
		return nil
	}
	stated := false
	tcfUnfenced(p.lines, func(_ int, l string) bool {
		stated = psBaselineConventionRE.MatchString(l)
		return stated
	})
	if stated || !p.on("F12") {
		return nil
	}
	return []string{p.at(line, "task "+declaring+" declares **Baseline:** fields but the plan states no **Baseline convention:** header line — one stated convention is what keeps later tasks from drifting between counting methods; seed the `> **Baseline convention:**` line (skills/flow/brainstorm-planner.md)")}
}

// fileOwnership is F11. Owners come from tcfParseTask's files, one task at a
// time in plan order, each id once; an ordering is a gating declared After
// edge between the pair in either direction.
func (p *psPlan) fileOwnership() []string {
	owners := map[string][]string{}
	var order []string
	filesLines := map[string]int{}
	declared := map[string][]string{}
	seen := map[string]bool{}
	for _, task := range tcfIterTasks(p.lines) {
		if seen[task.id] {
			continue
		}
		seen[task.id] = true
		if tcfUnclosedFence(task.lines) >= 0 {
			continue // F3b: this task's fields are unread
		}
		fields := p.parse(task.id)
		for _, name := range fields.files {
			if _, ok := owners[name]; !ok {
				order = append(order, name)
			}
			// A repeated token within one task's own list is one owner: a
			// pair needs two tasks.
			if !slices.Contains(owners[name], task.id) {
				owners[name] = append(owners[name], task.id)
			}
		}
		if a := tcfSelectAfter(task.lines); a != nil && a.gates {
			declared[task.id] = a.ids
		}
		if l := psFieldLine(task.lines, task.taskLine, "Files"); l != 0 {
			filesLines[task.id] = l
		}
	}
	var v []string
	for _, name := range order {
		ids := owners[name]
		for i := 1; i < len(ids); i++ {
			later := ids[i]
			for _, earlier := range ids[:i] {
				unordered := !slices.Contains(declared[earlier], later) && !slices.Contains(declared[later], earlier)
				if unordered && p.on("F11") {
					line := filesLines[later]
					if line == 0 {
						line = 1
					}
					v = append(v, p.at(line, "task "+later+"'s **Files:** also names `"+name+"`, first named by task "+earlier+", and no **After:** ordering is declared between the two — two tasks editing one file is the shared-file conflict this check exists to make explicit"))
				}
			}
		}
	}
	return v
}

// afterReferences is F9 and F10 over the After fields: F9 a gating id
// naming no task, F10 the first back-edge of the resolved graph. A task
// whose fence never closes is skipped for both.
func (p *psPlan) afterReferences() []string {
	tasks := tcfIterTasks(p.lines)
	var all []string
	idSet := map[string]bool{}
	taskLines := map[string]int{}
	for _, t := range tasks {
		all = append(all, t.id)
		idSet[t.id] = true
		if _, ok := taskLines[t.id]; !ok {
			taskLines[t.id] = t.taskLine
		}
	}
	declared := map[string]*tcfAfter{}
	for _, t := range tasks {
		if _, ok := declared[t.id]; ok {
			continue
		}
		if tcfUnclosedFence(t.lines) >= 0 {
			continue // F3b: the field is unread for this task
		}
		if a := tcfSelectAfter(t.lines); a != nil {
			declared[t.id] = a
		}
	}

	var v []string
	for _, id := range all {
		d := declared[id]
		if d == nil || !d.gates {
			continue
		}
		for _, name := range d.ids {
			if !idSet[name] {
				v = append(v, p.at(p.afterFieldLine(id), "task "+id+"'s **After:** names task "+name+", which this plan does not define"))
			}
		}
	}

	edges := map[string][]string{}
	for i, id := range all {
		if d := declared[id]; d != nil && d.gates {
			var e []string
			for _, name := range d.ids {
				if idSet[name] {
					e = append(e, name)
				}
			}
			edges[id] = e
		} else {
			edges[id] = slices.Clone(all[:i])
		}
	}
	visited := map[string]bool{}
	reported := false
	for _, s := range all {
		if reported {
			break
		}
		if visited[s] {
			continue
		}
		var stack []string
		onStack := map[string]bool{}
		var dfs func(node string)
		dfs = func(node string) {
			if reported {
				return
			}
			visited[node] = true
			onStack[node] = true
			stack = append(stack, node)
			for _, next := range edges[node] {
				if reported {
					return
				}
				if onStack[next] {
					cycle := append(slices.Clone(stack[slices.Index(stack, next):]), next)
					parts := make([]string, len(cycle))
					for i, n := range cycle {
						parts[i] = "task " + n
					}
					anchor := taskLines[cycle[0]]
					if anchor == 0 {
						anchor = 1
					}
					v = append(v, p.at(anchor, strings.Join(parts, " -> ")+": the resolved After graph has a cycle, so neither bundle can ever dispatch"))
					reported = true
					return
				}
				if !visited[next] {
					dfs(next)
				}
			}
			delete(onStack, node)
			stack = stack[:len(stack)-1]
		}
		dfs(s)
	}
	return v
}

// afterFieldLine is the file line of task id's After field, or its task
// line when the body carries none.
func (p *psPlan) afterFieldLine(id string) int {
	for _, t := range tcfIterTasks(p.lines) {
		if t.id != id {
			continue
		}
		if a := tcfSelectAfter(t.lines); a != nil {
			return t.taskLine + a.offset + 1
		}
		return t.taskLine
	}
	return 1
}

// psDir is os.path.dirname.
func psDir(p string) string {
	head := p[:strings.LastIndex(p, "/")+1]
	if strings.Trim(head, "/") != "" {
		head = strings.TrimRight(head, "/")
	}
	return head
}

// psStem is a test-shaped file's subject: fold_fixup_test.go -> fold_fixup.
func psStem(path string) string {
	name := path[strings.LastIndex(path, "/")+1:]
	for _, suffix := range []string{"_test.go", ".spec.ts"} {
		name = strings.TrimSuffix(name, suffix)
	}
	name, _, _ = strings.Cut(name, ".test.")
	return strings.TrimPrefix(name, "test-")
}

// psKey folds a name for F15's stem comparison: no leading Test/test_,
// lower case, letters and digits only.
func psKey(name string) string {
	return psNonKey.ReplaceAllString(strings.ToLower(psTestPfx.ReplaceAllString(name, "")), "")
}

type psDecl struct{ file, name string }

// baseDeclarations is F15's (file, name) test declarations in the
// test-shaped files of the change's base tree; none when the plan sits
// outside a git repository or no base resolves.
func (p *psPlan) baseDeclarations() []psDecl {
	git := envGit(p.env)
	run := func(dir string, args ...string) (string, bool) {
		out, err := git(append([]string{"-C", dir, "-c", "core.quotePath=false"}, args...)...).Output()
		return string(out), err == nil
	}
	root, ok := run(psDir(filepath.Clean(psAbs(p.env, p.path))), "rev-parse", "--show-toplevel")
	if !ok {
		return nil
	}
	root = tcfStrip(root)
	branch, _ := run(root, "branch", "--show-current")
	recorded := ""
	if branch = tcfStrip(branch); branch != "" {
		recorded, _ = run(root, "config", "--get", "branch."+branch+".flowBase")
		recorded = tcfStrip(recorded)
	}
	candidates := []string{"refs/remotes/origin/HEAD"}
	if recorded != "" {
		candidates = []string{"refs/remotes/origin/" + recorded, recorded}
	}
	rev := ""
	for _, ref := range candidates {
		if out, ok := run(root, "merge-base", "HEAD", ref); ok {
			rev = tcfStrip(out)
			break
		}
	}
	if rev == "" {
		return nil
	}
	args := []string{"grep", "-I", "-z", "-E", "(func|fun|def)[[:space:]]", rev, "--"}
	for _, g := range psTestShaped {
		args = append(args, ":(glob)**/"+g)
	}
	out, _ := run(root, args...)
	var decls []psDecl
	for _, record := range tcfSplitLines(out) {
		where, text, _ := strings.Cut(record, "\x00")
		file := ""
		if len(where) > len(rev) {
			file = where[len(rev)+1:]
		}
		for _, m := range psDeclRE.FindAllStringSubmatch(text, -1) {
			decls = append(decls, psDecl{file, m[2]})
		}
	}
	return decls
}

// baseDeclared is F15.
func (p *psPlan) baseDeclared(ids []string) []string {
	decls := p.baseDeclarations()
	if len(decls) == 0 {
		return nil
	}
	var v []string
	for _, id := range ids {
		fields := p.parse(id)
		if fields.unclosedFenceLine != 0 {
			continue
		}
		dirs := map[string]bool{}
		for _, f := range fields.files {
			dirs[psDir(f)] = true
		}
		for _, spec := range fields.tests {
			if !psIdentRE.MatchString(spec.label) {
				continue
			}
			hits := map[string][]string{}
			for _, d := range decls {
				same := d.name == spec.label || strings.HasPrefix(d.name, spec.label) && psKey(psStem(d.file)) == psKey(spec.label)
				owned := slices.Contains(fields.files, d.file) ||
					slices.ContainsFunc(fields.allowedCollateral, func(g string) bool { return tcfFnmatch(d.file, g) })
				if same && !owned && dirs[psDir(d.file)] && p.on("F15") {
					hits[d.file] = append(hits[d.file], "`"+d.name+"`")
				}
			}
			if len(hits) == 0 {
				continue
			}
			found := tcfSelectTask(p.lines, id)
			files := make([]string, 0, len(hits))
			for f := range hits {
				files = append(files, f)
			}
			sort.Strings(files)
			where := make([]string, len(files))
			for i, f := range files {
				where[i] = f + " (" + strings.Join(hits[f], ", ") + ")"
			}
			v = append(v, p.at(psFieldLine(found.lines, found.taskLine, "Tests"), "task "+id+"'s **Tests:** names `"+spec.label+"`, which the base already declares outside its **Files:** — "+strings.Join(where, "; ")+"; the task's commit cannot add it, so check-task-commit-fields refuses the commit — name the test this task adds, or list the file in **Files:**"))
		}
	}
	return v
}
