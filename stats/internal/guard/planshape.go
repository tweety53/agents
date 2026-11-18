package guard

import (
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
	"unicode"
	"unicode/utf8"
)

// The contract below moved here from the header of the deleted
// scripts/check-plan-shape.sh shim (KAN-1058).
//
// This file exists so .flow/project.md's declared lint/test commands and an
// operator's muscle memory invoking this exact filename keep working. Scope
// is one file per call:
//
//   - no arguments: scan every live change's tasks.md under
//     <spec-root>/changes/*/tasks.md, aggregating exit codes — non-zero if ANY file has a violation,
//     printing each file's own violations;
//   - one argument: treat it as an explicit tasks.md path and scan only
//     that file.
//
// The no-argument scan's root is CHECK_PLAN_SHAPE_ROOT when set (a test
// harness's sandboxed fixture tree), else FLOW_GUARD_REPO_ROOT — exported by scripts/flow-guard — as the checkout whose changes are
// scanned.

// check-plan-shape checks that a tasks.md is SHAPED so the guard that later
// judges its commits (check-task-commit-fields) can actually read it. Thirty-two
// findings, F1-F34 (F1-F6's canonical definitions, decisions and rationale:
// spectre/changes/kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan/
// design.md -- not restated here; F7-F34's are the rows below).
//
// The guard reads the plan with the very grammar the commit guard reads it
// with -- tcfParseTask, tcfTaskIDs, tcfIterTasks, tcfSelectTask,
// tcfUnclosedFence and tcfSelectAfter in planparse.go -- never a copy:
// parser identity is structural (one function), never asserted.
//
// Arguments: `-rules` alone prints every live rule as one `<id>  <sentence>`
// line from psRules and exits 0 -- the list a planner loads before writing
// tasks.md (decision rules-flag); one tasks.md path checks that file; none
// scans every live
// change's <spec-root>/changes/*/tasks.md under the repository root --
// never the archive/ entry, since run 1 drops a change's directory rather
// than moving it. The root
// is CHECK_PLAN_SHAPE_ROOT when set (a harness's sandboxed fixture tree),
// else FLOW_GUARD_REPO_ROOT, which scripts/flow-guard exports from its own real location.
//
// Per file, the F1-F18 checks run, then plan-suite-steps
// (plansuitesteps.go: a task step running a module or repository suite),
// plan-script-paths (planscriptpaths.go: a step citing a scripts/ path the
// repository does not carry), plan-files-paths (planfilespaths.go: a
// **Files:** path whose directory exists only with a segment dropped) and
// plan-task-id-rule (plantaskidrule.go: a
// /flow plan stating task commits carry no Task-Id: trailer) and
// plan-step-files (planstepfiles.go: a path an edit step names that its
// task's **Files:**/**Allowed-collateral:** does not cover), in-process
// and in that order, each printing its own lines.
//
// Exit codes, per file and over the scan: 2 if any check cannot answer
// (wrong argument count, a file that cannot be read -- a plan that is not
// UTF-8 included, where check-plan-shape.py crashed with exit 1 and its
// header named exit 2; the header wins -- or an archive listing that
// fails), else 1 if any found a violation, else 0. Violations print to
// stdout one per line as `file:line: <id> message`, <id> the rule's
// -rules id -- an F-row, or a sub-guard's name (decision line-id-position)
// -- and the message naming the task id (F5 excepted: a file with no tasks
// has no task id to name). Each violation records its id as the run's
// reason (Reason).
//
// Findings -- every one reported as file:line, fences excluded from every
// scan (tcfFenceRE): a field-shaped line inside a worked example is not a
// declaration.
//
//	F1  A second Files, Tests, Commit, Baseline or Allowed-collateral line
//	    in one task body (psF1Fields) -- naming which occurrence wins, since
//	    the parser keeps the first.
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
//	F11 A file named in more than one task's Files with no path of GATING
//	    declared After edges between a co-owning pair, in either direction -- the
//	    shared-file conflict KAN-339's self-review recorded. The serial
//	    default notes nothing and `none` declares independence, not an
//	    order. Reported once per unordered pair at the LATER task's Files
//	    line; a task whose fence never closes is no owner. A file is its
//	    path under its task's `## <repository>` heading, one naming a repository the plan's project declares (pdgRepositories):
//	    one path under two headings is two repositories' files, never a
//	    shared one (kan-1112 self-review); a task under no heading is
//	    under "", one repository as before.
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
//	F17 A task whose Baseline before is below the after of the previous
//	    task whose Baseline carries the same measured: command, in plan
//	    order: chain-position absolutes chain, so a later task re-declaring
//	    a plan-time pair fails only at its commit (kan-1019). The later task
//	    of a Squash-with pair -- either naming the other -- shares its
//	    partner's commit and is exempt. `<` rather than `!=`: a task between
//	    the two may add tests to the same source set under another command
//	    (kan-916, kan-939). Reported at the task's Baseline line.
//	F18 An unticked task's absence sweep (a backticked `git grep ... --
//	    <paths>` followed by `prints nothing`) whose paths cover a later
//	    task's Files path the task does not own, while that file holds one
//	    of the sweep's `-e '<string>'` strings at the base F15 resolves --
//	    the sweep cannot print nothing at this task's commit (kan-1021's
//	    Tasks 5 and 6). A plan whose base does not resolve is not checked.
//	    Reported at the sweep's line, once per such file.
//	F20 A Baseline value (the joined field) carrying a `<!-- measured: … -->`
//	    comment tcfMeasuredRE cannot read -- no ` @ <ref>` before its `-->`:
//	    the parser takes no command from it, so the re-measurement is
//	    silently disabled (kan-1034's Task 2). Reported at the Baseline line.
//	F19 A tasks.md whose leading `>` header block carries
//	    `> **Design shape:** plan-mode` beside a design.md missing any of
//	    `## Recommended approach`, `## Critical files`, `## Existing code to
//	    reuse`, `## Verification`, `## Non-goals`, `## Risks` -- one `<design.md>:1: F19 missing section
//	    <heading> (Design shape: plan-mode)` line each; an absent design.md
//	    is one such line naming the file. A plan without the line is not
//	    checked, so changes planned before the sections existed keep
//	    passing (kan-1033 design.md §4, decision design-shape-line).
//	F21 Retired: an in-run visual-verify fix task's record is on disk only (the change directory is gitignored), so its **Files:** never names it.
//	F22 A plan whose sibling design.md cites a `.superpowers/sdd/*.html`
//	    prototype, in a repository whose `.flow/project.md` declares a
//	    `## visual verification` section with `screenshots` and no
//	    `regression checkout`, where some task's **Files:** path matches a
//	    `ui paths` glob and no task's **Files:** path lies under
//	    `screenshots`: the capture spec is planned with the UI change
//	    (skills/flow/brainstorm-planner.md; kan-1047). Reported at the first
//	    such task's title line. A Markdown path never counts — prose an app
//	    embeds, not a view a prototype draws (self-review-kan-1110 F4).
//	F23 A Files or Allowed-collateral entry under the planning paths
//	    (psPlanning): such a path rides a
//	    planning or `Flow-Key: flow.visual-verify` commit, never a task
//	    commit -- check-task-commit-planning-paths' exemption (kan-1048
//	    self-review, 1-E). F14 still names the all-planning case. Reported
//	    at the field's line, once per entry.
//	F24 A task titled `In-run fix <k> — …`, any stage, with no **Opener:**
//	    field, or whose **Opener:** value another such task before it names
//	    too, backticks ignored, unless the two share a **Files:** path -- a
//	    slice: one verifier report opens one round
//	    (skills/flow/verify-fix-loop.md step 1; decision
//	    one-report-one-round-opener). A ticked task is never
//	    reported -- it may predate the field -- but its opener still counts
//	    against a later task's. Reported at the task's title line -- for a
//	    shared opener, the second task's.
//	F25 A sibling design.md decision whose **Chosen:** text, wrapped lines
//	    joined, names the parent as actor (psParentActorRE; a possessive is
//	    not) and a run moment (psRunMomentRE: before the first task, at
//	    flow.<stage> -- without one it designs the parent's role; self-review
//	    review-1, R1) and whose entry carries no column-0 `- [ ] Parent:` or
//	    `- [x] Parent:` line: an action outside every task is tracked, so
//	    check-unfinished-work holds run 1 until it is ticked (decisions
//	    parent-actions-in-design, parent-action-detection; kan-1079
//	    self-review, F1). Reported at the decision's **ID:** line in
//	    design.md. An absent or unreadable design.md is not checked.
//	F26 A line, fenced or not -- a step's command sits in a fence --
//	    running npx with its yes flag (psNpxYesRE), or npx in a command's
//	    position -- a line's or code span's start, after ; | & ( or $
//	    (psNpxRE), so prose naming npx is not one -- without --no-install or
//	    --no (psNpxNoRE): with no TTY npm assumes --yes, so a bare npx
//	    fetches and runs a package the machine lacks just as -y does: the
//	    external code auto mode
//	    denies (decision npx-yes-refused; kan-1079 self-review, F3; its
//	    review-1, R2). A step runs a tool installed or declared in
//	    .flow/project.md. Reported at the line.
//	F27 A sibling design.md whose `## Open questions` section, fences
//	    excluded, holds an entry (a `###` heading) whose **Status:** does
//	    not open with `answered by`, a missing **Status:** included, or a
//	    list item: planning leaves nothing open, and a question the
//	    operator cannot answer is asked again, narrowed, never recorded
//	    open (skills/spectre-planning/SKILL.md, Convergence). Reported at
//	    the entry's heading or the item's line. An absent or unreadable
//	    design.md is not checked.
//	F28 An unticked task (`- [ ]`) whose **Files:** lists run-loaded prose
//	    -- a path vmScoped accepts (check-verbatim-moves' own predicate),
//	    not a -rationale.md -- while no verbatim-moves.txt sits beside
//	    tasks.md: the plan writes that file, `#` comment lines alone when
//	    it only moves text verbatim (.flow/project.md; kan-1103
//	    self-review, F9). Reported once per task at its task line.
//	F29 Any line, fenced or not, running `go test` (`go test` or
//	    `go -C <dir> test` followed by a flag or a `.`-path argument) with
//	    no `-count=` flag: the test cache never rechecks a file outside the
//	    module root, so a cached run passes an edited Markdown file the
//	    tests read (skills/flow/brainstorm-planner.md; self-review-kan-1110
//	    F6). Reported at the line.
//	F30 An unticked task's line -- its task line or any body line --
//	    presuming the machine "on an otherwise idle machine" (hyphenated
//	    or not): a shared dev host is never idle, so a timing gate waits
//	    for load at or under the core count (skills/flow/brainstorm-planner.md;
//	    kan-1124 self-review, F2). A ticked task is history, never
//	    reported. Reported at the line.
//	F31 An unticked task's line handing the parent an action on the
//	    sibling design file -- "the parent records/writes/adds/notes/
//	    commits" or "parent note", on a line naming design.md: the task's
//	    tick closes step prose whether the parent acted or not, so the
//	    action is a column-0 `- [ ] Parent:` line in design.md, which
//	    check-unfinished-work holds open (skills/flow/brainstorm-planner.md;
//	    kan-1124 self-review, F3). Measured over 1166 plans (decision
//	    f3-parent-step-actions): both conditions hit kan-1124's three lines
//	    alone, the actor phrase alone 9 lines of pipeline prose. A ticked
//	    task is never reported. Reported at the line.
//	F32 A task's **Size:** value -- its first column-0, unfenced
//	    **Size:** line -- is `small`, `medium` or `big`: an unknown value
//	    would be packed by plan-dispatch-groups as `medium`, and the plan
//	    gate owns field validation, so a typo fails here rather than first
//	    at dispatch (design decision size-checked-at-plan-gate). Reported
//	    at the size line.
//	F33 One fix round is one task (kan-1099: task 46 opened round 1 and
//	    five `Round 1 continued` tasks followed, each with its own 35-minute
//	    verify re-run). Refused, each at the later task's title line: an
//	    unticked `In-run fix <k> — …` task beside an earlier one that is
//	    unticked or carries the same `<k>`, unless the two share a
//	    **Files:** path -- a slice, cut only where both edit one file in
//	    conflicting ways; an unticked one whose **Opener:** carries no
//	    `-fix-<j>` segment with `<j>` at or above the highest earlier fix
//	    number below its own -- that report's departures were known when
//	    the earlier round was planned; and an unticked task after an in-run
//	    fix task that is neither titled so nor counted by
//	    `**Tasks appended:**`. A ticked task is never reported.
//	F34 A visual fix task cites every departure row of its opener
//	    (kan-1099 self-review, F22: a fix task dropped a departure its
//	    report named). An unticked `In-run fix <k> — …` task whose
//	    **Opener:** starts `visual-verify` cites `<frame> #<row>` for each
//	    `#<row>: departure` item of the report's `- <frame> over-tolerance:`
//	    lines, the report read from
//	    `<worktree>/.superpowers/sdd/verify-report-<opener>.md`, `<worktree>`
//	    the plan path above `spectre/changes/<name>/tasks.md`. An absent
//	    report owes nothing. Reported at the task's title line, one finding
//	    naming every uncited row.
//
// A duplicate task id is check-task-build-green's violation, not this
// guard's: tcfTaskIDs keeps the first occurrence, and every per-task read
// resolves to it.
// psSubGuards are the plan sub-guards psCheckOne runs after the F-checks,
// in order; each is also its own -rules row and reason id (decision
// sub-guard-ids).
var psSubGuards = []string{"plan-suite-steps", "plan-script-paths", "plan-files-paths", "plan-task-id-rule", "plan-step-files"}

// psRules is what `check-plan-shape -rules` prints: one row per live rule,
// its id and the planner-facing sentence -- what a plan must do. Each
// rule's rationale stays in this file's header (decisions rules-flag,
// rationale-stays-in-header). F21 is retired and has no row.
var psRules = []struct{ id, sentence string }{
	{"F1", `A task body declares each of **Files:**, **Tests:**, **Commit:**, **Baseline:**, **Allowed-collateral:** and **Size:** at most once.`},
	{"F2", `Every task declares a **Files:** line naming at least one file.`},
	{"F3a", "Every field line (`**Files:**`, `**Tests:**`, `**After:**`, …) starts at column 0, never indented."},
	{"F3b", `Every code fence a task body opens closes inside that body.`},
	{"F4", "A task whose **Files:** names a test-shaped path (`test-*`, `*_test.go`, `*.test.*`, `*.spec.ts`) never opens **Tests:** with `none`, save `none — repairs existing tests: <names>`."},
	{"F5", "The plan holds at least one `- [ ] <n>. <title>` task line."},
	{"F6", "**Tests:** opens with `none`, carries a `Case <N>` label, or names a test in backticks."},
	{"F7", `A task body carries at most one gating **After:** line.`},
	{"F8", "An **After:** value reads `Task <ids>` or `none`."},
	{"F9", `Every task an **After:** names is a task of this plan.`},
	{"F10", `The resolved **After:** graph has no cycle; a task with no **After:** comes after every earlier task.`},
	{"F11", `Two tasks whose **Files:** name one file are ordered by declared **After:** fields, directly or through other tasks; the absent-field default orders nothing.`},
	{"F12", "A plan whose tasks declare **Baseline:** carries the header line `> **Baseline convention:**`."},
	{"F13", "A `measured:` command counting `@Test` with `grep -c` or `grep -o` also passes `-w`."},
	{"F14", "A task's **Files:** names at least one path outside `spectre/changes/`, `openspec/changes/` and `docs/superpowers/`."},
	{"F15", `Every backticked **Tests:** name is new: no test-shaped file outside the task's **Files:**, in a directory those files lie in, declares it at the change's base.`},
	{"F16", `A **Files:** or **Allowed-collateral:** value naming more than one path backticks each path.`},
	{"F17", "A task's **Baseline:** before is never below the after of the previous task whose **Baseline:** carries the same `measured:` command, save the later task of a **Squash-with:** pair."},
	{"F18", `An absence sweep's paths never cover a later task's **Files:** path that still holds a swept string at the base.`},
	{"F19", "A tasks.md whose header carries `> **Design shape:** plan-mode` has a sibling design.md holding " + psDesignSectionList() + "."},
	{"F20", "A **Baseline:** `measured:` comment reads `<!-- measured: <command> @ <ref> -->`."},
	{"F22", "When design.md cites a `.superpowers/sdd/*.html` prototype and a task's non-Markdown **Files:** path matches a `ui paths` glob, some task's **Files:** names a capture under `screenshots`."},
	{"F23", "No **Files:** or **Allowed-collateral:** entry lies under `spectre/changes/`, `openspec/changes/` or `docs/superpowers/`."},
	{"F24", "Every unticked `In-run fix <k> — …` task carries an **Opener:**, one an earlier such task names only when the two share a **Files:** path."},
	{"F25", "A design.md decision whose **Chosen:** has the parent act before the first task or at `flow.<stage>` carries a column-0 `- [ ] Parent: <action>` line."},
	{"F26", "No line runs npx without `--no-install` or `--no`, nor with `-y` or `--yes`."},
	{"F27", "Every `## Open questions` entry in design.md reads `**Status:** answered by <decision-id>`, and the section holds no list item."},
	{"F28", `A plan whose unticked task edits run-loaded prose has a verbatim-moves.txt beside tasks.md.`},
	{"F29", "Every go test command passes `-count=`."},
	{"F30", "No unticked task's line presumes the machine otherwise idle; a timing gate waits for load at or under the core count."},
	{"F31", "No unticked task's line hands the parent an action on the sibling design file; that action is a column-0 Parent line there."},
	{"F32", "A task's **Size:** value is `small`, `medium` or `big`."},
	{"F33", "One fix round is one `In-run fix <k> — …` task: no second unticked one unless the two share a **Files:** path, none opened by a report older than the previous round's `-fix-<k>` re-run, and no task under another title after one outside **Tasks appended:**."},
	{"F34", "An unticked task whose **Opener:** is a `visual-verify…` key cites `<frame> #<row>` for every over-tolerance row its opener's report classifies `departure`."},
	{"plan-suite-steps", "No step runs a module or repository test suite: go test takes `-run`, Gradle `--tests`, npm test and vitest `-t`."},
	{"plan-script-paths", "Every `scripts/` path a task cites exists in the repository or is in some task's **Files:**."},
	{"plan-files-paths", `Every **Files:** path names the directory its file lands in, never one with a segment dropped.`},
	{"plan-task-id-rule", "A /flow plan never says its task commits go without a `Task-Id:` trailer; that rule is a fix loop's alone."},
	{"plan-step-files", `Every file an edit step names is covered by its task's **Files:** or **Allowed-collateral:**.`},
}

// psDesignSectionList is F19's headings as `a`, `b` and `c`.
func psDesignSectionList() string {
	q := make([]string, len(psDesignSections))
	for i, h := range psDesignSections {
		q[i] = "`" + h + "`"
	}
	return strings.Join(q[:len(q)-1], ", ") + " and " + q[len(q)-1]
}

func init() { Registry["check-plan-shape"] = checkPlanShape }

func checkPlanShape(args []string, env Env, stdout, stderr io.Writer) int {
	switch {
	case len(args) == 1 && args[0] == "-rules":
		for _, r := range psRules {
			fmt.Fprintf(stdout, "%s  %s\n", r.id, r.sentence)
		}
		return 0
	case len(args) == 1:
		return psCheckOne(args[0], env, "", stdout, stderr)
	case len(args) == 0:
	default:
		fmt.Fprintln(stderr, "usage: flow-guard check-plan-shape [-rules | path-to-tasks.md]")
		Reason(env, "usage")
		return 2
	}
	root := env.Getenv("CHECK_PLAN_SHAPE_ROOT")
	if root == "" {
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintln(stderr, "check-plan-shape: FLOW_GUARD_REPO_ROOT is unset — run flow-guard check-plan-shape, which sets it")
			Reason(env, "repo-root-unset")
			return 2
		}
	}
	files, ok := tcfScanPlans(env, root, stderr)
	if !ok {
		Reason(env, "scan-failed")
		return 2
	}
	status := 0
	for _, f := range files {
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

// psCheckOne checks one file: F1-F18, then the three plan sub-guards. off
// names one finding whose decision is disabled -- the harness's mutation
// proof, which sed-disabled that decision's `if` in a copy of the guard;
// production passes "".
func psCheckOne(path string, env Env, off string, stdout, stderr io.Writer) int {
	rc := psShape(path, env, off, stdout, stderr)
	for _, name := range psSubGuards {
		rc = psWorse(rc, Registry[name]([]string{path}, env, stdout, stderr))
	}
	return rc
}

func psShape(path string, env Env, off string, stdout, stderr io.Writer) int {
	lines, err := tcfReadPlanTyped(env, path)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", path, err)
		Reason(env, "cannot-read-file")
		return 2
	}
	p := &psPlan{path: path, lines: lines, off: off, env: env}
	violations := p.check()
	for _, f := range violations {
		Reason(env, f.id)
		fmt.Fprintln(stdout, f.text)
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

// psFinding is one violation: its rule id and its printed line.
type psFinding struct{ id, text string }

// at is a finding at line of the plan; atFile one at line of file. Both
// print `<file>:<line>: <id> <msg>`, the id a -rules id (decision
// line-id-position).
func (p *psPlan) at(line int, id, msg string) psFinding { return p.atFile(p.path, line, id, msg) }

func (p *psPlan) atFile(file string, line int, id, msg string) psFinding {
	return psFinding{id, fmt.Sprintf("%s:%d: %s %s", file, line, id, msg)}
}

// psF1Fields are the six names F1 counts; Regression and Build are not
// part of it -- Build's precedence is tcfSelectBuildKind's job.
var psF1Fields = []string{"Files", "Tests", "Commit", "Baseline", "Allowed-collateral", "Size"}

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
	// psSweepRE is F18's absence sweep: group 1 the git grep's options and
	// strings, group 2 its paths. psSweepPatRE is one `-e '<string>'`.
	psSweepRE    = regexp.MustCompile("`(git grep [^`]*?) -- ([^`]*)`[" + tcfWS + "]*prints nothing")
	psSweepPatRE = regexp.MustCompile(`-e[` + tcfWS + `]+'([^']*)'`)
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

func (p *psPlan) check() []psFinding {
	ids := tcfTaskIDs(p.lines)
	if len(ids) == 0 && p.on("F5") {
		return []psFinding{p.at(1, "F5", "this plan defines zero tasks — the grammar finds no `- [ ] <n>. ` checkbox line, so every check that follows would be a vacuous pass rather than a real one")}
	}
	var v []psFinding
	for _, id := range ids {
		v = append(v, p.checkTask(*tcfSelectTask(p.lines, id))...)
	}
	v = append(v, p.afterReferences()...)
	v = append(v, p.fileOwnership()...)
	v = append(v, p.baselineConvention()...)
	v = append(v, p.baselineChain()...)
	v = append(v, p.baseDeclared(ids)...)
	v = append(v, p.sweepScope()...)
	v = append(v, p.prototypeCapture()...)
	v = append(v, p.inRunBatches(ids)...)
	v = append(v, p.parentActions()...)
	v = append(v, p.npxFetch()...)
	v = append(v, p.goTestCache()...)
	v = append(v, p.idleTiming()...)
	v = append(v, p.parentSteps()...)
	v = append(v, p.openQuestions()...)
	v = append(v, p.verbatimMoves()...)
	return append(v, p.designSections()...)
}

var (
	psDesignShapeRE  = regexp.MustCompile(`^>[` + tcfWS + `]*\*\*Design shape:\*\*[` + tcfWS + `]*plan-mode[` + tcfWS + `]*$`)
	psDesignSections = []string{"## Recommended approach", "## Critical files", "## Existing code to reuse", "## Verification", "## Non-goals", "## Risks"}
)

// designSections is F19: the plan-mode sections of the sibling design.md,
// checked only when the header block -- the `>` lines before the first
// task -- declares the plan-mode design shape.
func (p *psPlan) designSections() []psFinding {
	shaped := false
	for _, l := range p.lines {
		if pcTaskLine.MatchString(l) {
			break
		}
		if psDesignShapeRE.MatchString(l) {
			shaped = true
			break
		}
	}
	if !shaped || !p.on("F19") {
		return nil
	}
	design := filepath.Join(filepath.Dir(p.path), "design.md")
	raw, err := os.ReadFile(psAbs(p.env, design))
	if errors.Is(err, fs.ErrNotExist) {
		return []psFinding{p.atFile(design, 1, "F19", "missing file design.md (Design shape: plan-mode)")}
	}
	if err != nil {
		return []psFinding{p.atFile(design, 1, "F19", "cannot read design.md: "+err.Error()+" (Design shape: plan-mode)")}
	}
	have := map[string]bool{}
	for _, l := range strings.Split(string(raw), "\n") {
		have[strings.TrimRight(l, " \t\r")] = true
	}
	var v []psFinding
	for _, h := range psDesignSections {
		if !have[h] {
			v = append(v, p.atFile(design, 1, "F19", "missing section "+h+" (Design shape: plan-mode)"))
		}
	}
	return v
}

// psPrototypeRE is F22's prototype citation in design.md.
var psPrototypeRE = regexp.MustCompile("\\.superpowers/sdd/[^\\s`)]+\\.html")

// prototypeCapture is F22: a prototype-specified UI change plans its capture.
func (p *psPlan) prototypeCapture() []psFinding {
	if !p.on("F22") {
		return nil
	}
	design, err := os.ReadFile(psAbs(p.env, filepath.Join(filepath.Dir(p.path), "design.md")))
	if err != nil || !psPrototypeRE.Match(design) {
		return nil
	}
	top, err := exec.Command("git", "-C", psAbs(p.env, filepath.Dir(p.path)), "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return nil
	}
	root := strings.TrimSpace(string(top))
	cfg, err := os.ReadFile(filepath.Join(root, ".flow", "project.md"))
	if err != nil {
		return nil
	}
	cells := vvCells(string(cfg))
	shots := strings.TrimSuffix(cells["screenshots"], "/")
	if shots == "" || cells["regression checkout"] != "" {
		return nil
	}
	var first tcfTaskBody
	uiPath := ""
	for _, id := range tcfTaskIDs(p.lines) {
		for _, f := range p.parse(id).files {
			if strings.HasPrefix(f, shots+"/") {
				return nil
			}
			if strings.HasSuffix(f, ".md") {
				continue
			}
			if uiPath == "" && visualTrigger(quietReasons(p.env), root, func() ([]string, error) { return []string{f}, nil }, io.Discard, io.Discard) == 0 {
				uiPath, first = f, *tcfSelectTask(p.lines, id)
			}
		}
	}
	if uiPath == "" {
		return nil
	}
	return []psFinding{p.at(first.taskLine, "F22", "design.md cites a prototype and this task touches `ui paths` (`"+uiPath+"`), but no task's **Files:** names a capture under `"+shots+"/` — plan the capture spec driving every interactive state the prototype draws (skills/flow/brainstorm-planner.md)")}
}

// psInRunFixRE is F24's and F33's title: an in-run fix task of any stage,
// group 1 its fix number. psFixSegRE is a dispatch key's `-fix-<j>` segment.
var (
	psInRunFixRE = regexp.MustCompile(`^\s*- \[.\] \S+\s+In-run fix ([0-9]+) — `)
	psFixSegRE   = regexp.MustCompile(`-fix-([0-9]+)`)
)

// inRunBatches is F24, F33 and F34: every in-run fix task names the report that
// opened its round, and one round is one task carrying every open item --
// a second task of a round is a slice, and shares a **Files:** path with
// the task it is cut from.
func (p *psPlan) inRunBatches(ids []string) []psFinding {
	type fix struct {
		id, opener string
		k          int
		files      []string
		ticked     bool
	}
	var v []psFinding
	var fixes []fix
	appended := uwAppendedIn(strings.Join(p.lines, "\n"), false)
	for _, id := range ids {
		found := tcfSelectTask(p.lines, id)
		title := p.lines[found.taskLine-1]
		ticked := strings.HasPrefix(strings.TrimSpace(title), "- [x]")
		m := psInRunFixRE.FindStringSubmatch(title)
		if m == nil {
			// kan-1099: five `Round 1 continued` tasks, each its own verify re-run.
			if len(fixes) > 0 && !ticked && !slices.Contains(appended, id) && p.on("F33") {
				v = append(v, p.at(found.taskLine, "F33", "task "+id+" follows in-run fix task "+fixes[len(fixes)-1].id+" under another title — a task the fix loop appends is titled `In-run fix <k> — …`, and a round's open items go in its one task (skills/flow/verify-fix-loop.md step 1)"))
			}
			continue
		}
		cur := fix{id: id, ticked: ticked, files: p.parse(id).files}
		cur.k, _ = strconv.Atoi(m[1])
		cur.opener = strings.Trim(strings.TrimSpace(strings.Join(tcfBodyFields(found.lines)["Opener"], " ")), "`")
		slice := func(e fix) bool {
			return slices.ContainsFunc(cur.files, func(f string) bool { return slices.Contains(e.files, f) })
		}
		if !ticked {
			prev, sliced := 0, false
			for _, e := range fixes {
				sameOpener := cur.opener != "" && e.opener == cur.opener
				sliced = sliced || sameOpener // F24's, or a slice: never stale as well
				switch {
				case slice(e):
					sliced = sliced || e.k == cur.k
				case sameOpener && p.on("F24"):
					v = append(v, p.at(found.taskLine, "F24", "task "+id+"'s **Opener:** "+cur.opener+" is task "+e.id+"'s too — one report opens one round, and a second task transcribes it only as a slice sharing a **Files:** path with the first (skills/flow/verify-fix-loop.md step 1)"))
				case (e.k == cur.k || !e.ticked) && !sameOpener && p.on("F33"):
					v = append(v, p.at(found.taskLine, "F33", "task "+id+" is a second fix round beside in-run fix task "+e.id+" — every departure, open reviewer finding and named cause known now goes in that one task; slice only where both tasks edit one file in conflicting ways, naming that file in both **Files:** (skills/flow/verify-fix-loop.md step 1)"))
				}
				if e.k < cur.k {
					prev = max(prev, e.k)
				}
			}
			newest := 0
			for _, seg := range psFixSegRE.FindAllStringSubmatch(cur.opener, -1) {
				j, _ := strconv.Atoi(seg[1])
				newest = max(newest, j)
			}
			switch {
			case cur.opener == "" && p.on("F24"):
				v = append(v, p.at(found.taskLine, "F24", "task "+id+" is an in-run fix with no **Opener:** field — name the verifier report that opened its round (skills/flow/verify-fix-loop.md step 1)"))
			case cur.opener != "" && !sliced && newest < prev && p.on("F33"):
				v = append(v, p.at(found.taskLine, "F33", fmt.Sprintf("task %s's **Opener:** %s predates In-run fix %d's re-run — every departure that report names was known when round %d was planned and belonged in its one task; open a round only from a report keyed `-fix-%d` or later (skills/flow/verify-fix-loop.md step 1)", id, cur.opener, prev, prev, prev)))
			}
			if missing := p.uncitedDepartures(cur.opener, strings.Join(found.lines, "\n")); len(missing) > 0 && p.on("F34") {
				v = append(v, p.at(found.taskLine, "F34", "task "+id+" drops departure rows of its opener "+cur.opener+"'s report: "+strings.Join(missing, ", ")+" — cite each as `<frame> #<row>` (skills/flow/verify-fix-loop.md step 1)"))
			}
		}
		fixes = append(fixes, cur)
	}
	return v
}

// psChangePlanRE is a plan path under `spectre/changes/<name>/tasks.md`,
// group 1 its worktree.
var psChangePlanRE = regexp.MustCompile(`^(.*)/spectre/changes/[^/]+/tasks\.md$`)

// uncitedDepartures is F34: the `<frame> #<row>` of every over-tolerance row
// a `visual-verify…` opener's report classifies `departure` that the task
// text does not cite. No report, or a plan outside a worktree, owes none.
func (p *psPlan) uncitedDepartures(opener, text string) []string {
	m := psChangePlanRE.FindStringSubmatch(filepath.ToSlash(p.path))
	if m == nil || !strings.HasPrefix(opener, "visual-verify") {
		return nil
	}
	report, err := os.ReadFile(filepath.Join(m[1], ".superpowers", "sdd", "verify-report-"+opener+".md"))
	if err != nil {
		return nil
	}
	var missing []string
	_, rows := reportDepartures(string(report))
	for _, cite := range rows {
		if !citesWhole(text, cite) {
			missing = append(missing, cite)
		}
	}
	return missing
}

// psDepartureItem is an over-tolerance item the verifier classified a
// departure: its cause alone, or followed by ` — <its row's figures>`.
var psDepartureItem = regexp.MustCompile(`^#(\d+):\s*departure\s*(?:—.*)?$`)

// reportDepartures reads a verifier report's `- <frame> over-tolerance:`
// lines: every frame they judge, in order, and the `<frame> #<row>` of every
// item classified `departure`. An item quotes its row's figures after its
// cause (`#1: departure — gap a=37 b=43; ink.width a=705 b=701`), so the
// cause is read before any ` — `, and a figure split off at a `;` is no item.
func reportDepartures(report string) (judged, rows []string) {
	for _, line := range strings.Split(report, "\n") {
		over := vrOverLine.FindStringSubmatch(line)
		if over == nil {
			continue
		}
		judged = append(judged, over[1])
		_, items, _ := strings.Cut(over[2], "—")
		for _, item := range strings.Split(items, ";") {
			if it := psDepartureItem.FindStringSubmatch(strings.TrimSpace(item)); it != nil {
				rows = append(rows, over[1]+" #"+it[1])
			}
		}
	}
	return judged, rows
}

// citesWhole reports whether text carries cite as a whole token: no word
// rune on either side, so `xframe #3` and `frame #30` do not cite `frame #3`.
func citesWhole(text, cite string) bool {
	word := func(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) }
	for i := 0; ; {
		j := strings.Index(text[i:], cite)
		if j < 0 {
			return false
		}
		start, end := i+j, i+j+len(cite)
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		after, _ := utf8.DecodeRuneInString(text[end:])
		if (start == 0 || !word(before)) && (end == len(text) || !word(after)) {
			return true
		}
		i = start + 1
	}
}

var (
	// psDecisionIDRE, psParentActorRE, psRunMomentRE and psParentLineRE are
	// F25's reads of design.md: a decision's id, the parent named as actor
	// (not as a possessive), the run moment that makes it an action, and the
	// tracked line.
	psDecisionIDRE  = regexp.MustCompile("^\\*\\*ID:\\*\\*[ \\t]*`?([^`\\s]+)")
	psParentActorRE = regexp.MustCompile(`(?i)\bthe parent(?:[^'’\w]|$)`)
	psRunMomentRE   = regexp.MustCompile("(?i)\\bbefore the first task\\b|\\bat `?flow\\.[a-z]")
	psParentLineRE  = regexp.MustCompile(`^- \[[ xX]\] Parent:`)
)

// parentActions is F25: a decision giving the parent an action outside every
// task carries the `- [ ] Parent:` line check-unfinished-work counts.
func (p *psPlan) parentActions() []psFinding {
	if !p.on("F25") {
		return nil
	}
	design := filepath.Join(filepath.Dir(p.path), "design.md")
	raw, err := os.ReadFile(psAbs(p.env, design))
	if err != nil {
		return nil
	}
	var v []psFinding
	id, idLine, chosen, inChosen, tracked := "", 0, "", false, false
	flush := func() {
		if id != "" && !tracked && psParentActorRE.MatchString(chosen) && psRunMomentRE.MatchString(chosen) {
			v = append(v, p.atFile(design, idLine, "F25", fmt.Sprintf("decision %s's **Chosen:** gives the parent an action, but the entry carries no `- [ ] Parent: <action>` line — add one beneath its **Chosen:**, so check-unfinished-work holds run 1 until it is ticked (skills/flow/brainstorm-planner.md)", id)))
		}
	}
	for i, l := range strings.Split(string(raw), "\n") {
		switch {
		case strings.HasPrefix(l, "#"):
			flush()
			id, chosen, inChosen, tracked = "", "", false, false
		case psDecisionIDRE.MatchString(l):
			flush()
			id, idLine, chosen, inChosen, tracked = psDecisionIDRE.FindStringSubmatch(l)[1], i+1, "", false, false
		case strings.HasPrefix(l, "**Chosen:**"):
			chosen, inChosen = l, true
		case psParentLineRE.MatchString(l):
			tracked, inChosen = true, false
		case strings.HasPrefix(l, "**") || strings.TrimSpace(l) == "":
			inChosen = false
		case inChosen:
			chosen += " " + l
		}
	}
	flush()
	return v
}

// psNpxYesRE is F26's npx told to fetch without asking, psNpxRE npx run as a
// command, and psNpxNoRE the flag after it that keeps npx from fetching.
var (
	psNpxYesRE = regexp.MustCompile(`\bnpx[ \t]+(?:-y|--yes)\b`)
	psNpxRE    = regexp.MustCompile("(?:^[ \\t]*|[`;|&($][ \\t]*)npx[ \\t]+[^\\s\"']")
	psNpxNoRE  = regexp.MustCompile(`(?:^|\s)--no(?:-install)?(?:\s|$)`)
)

// npxFetch is F26: a plan step runs only what is installed.
func (p *psPlan) npxFetch() []psFinding {
	if !p.on("F26") {
		return nil
	}
	var v []psFinding
	for i, l := range p.lines {
		loc := psNpxRE.FindStringIndex(l)
		if psNpxYesRE.MatchString(l) || loc != nil && !psNpxNoRE.MatchString(l[loc[1]-1:]) {
			v = append(v, p.at(i+1, "F26", "this line runs npx without --no-install, which fetches and runs a package not installed here — run a tool already installed or declared in .flow/project.md, a cached package as `npx --no-install <bin>` (skills/flow/brainstorm-planner.md)"))
		}
	}
	return v
}

// psGoTestRE is a go test command — its next token a flag, a relative or
// import path, or none at the line's end — psCountRE its -count flag (F29).
var (
	psGoTestRE = regexp.MustCompile(`\bgo(?:\s+-C\s+\S+)?\s+test(?:\s+(?:[-.]|[\w.-]+/)|\s*(?:$|&&|;|\|))`)
	psCountRE  = regexp.MustCompile(`(?:^|\s)-count[= ]?[0-9]`)
)

// goTestCache is F29: a planned go test command disables the test cache.
func (p *psPlan) goTestCache() []psFinding {
	if !p.on("F29") {
		return nil
	}
	var v []psFinding
	for i, l := range p.lines {
		if psGoTestRE.MatchString(l) && !psCountRE.MatchString(l) {
			v = append(v, p.at(i+1, "F29", "this line runs go test without -count=, so a cached result can pass a file the run never re-read — the test cache never rechecks a file outside the module root; pass -count=1 (skills/flow/brainstorm-planner.md)"))
		}
	}
	return v
}

// untickedLines calls fn with each unticked task's line and then every body
// line, each with its 1-based line number.
func (p *psPlan) untickedLines(fn func(n int, l string)) {
	for _, task := range tcfIterTasks(p.lines) {
		if !strings.HasPrefix(p.lines[task.taskLine-1], "- [ ]") {
			continue
		}
		fn(task.taskLine, p.lines[task.taskLine-1])
		for off, l := range task.lines {
			fn(task.taskLine+off+1, l)
		}
	}
}

var psIdleRE = regexp.MustCompile(`(?i)\bon an otherwise[- ]idle machine\b`)

// idleTiming is F30: no unticked task's line presumes an idle machine.
func (p *psPlan) idleTiming() []psFinding {
	if !p.on("F30") {
		return nil
	}
	var v []psFinding
	p.untickedLines(func(n int, l string) {
		if psIdleRE.MatchString(l) {
			v = append(v, p.at(n, "F30", "this line presumes an idle machine a shared dev host never provides — wait for load at or under the core count, record the load beside the figure and name what a figure taken under load does (skills/flow/brainstorm-planner.md)"))
		}
	})
	return v
}

var (
	psParentStepRE = regexp.MustCompile(`(?i)\b(?:the parent[ ](?:record|write|add|note|commit)s?|parent[ ]note)\b`)
	psDesignRefRE  = regexp.MustCompile(`design\.md`)
)

// parentSteps is F31: no unticked task's line hands the parent an action on
// the sibling design file.
func (p *psPlan) parentSteps() []psFinding {
	if !p.on("F31") {
		return nil
	}
	var v []psFinding
	p.untickedLines(func(n int, l string) {
		if psParentStepRE.MatchString(l) && psDesignRefRE.MatchString(l) {
			v = append(v, p.at(n, "F31", "this line hands the parent an action on the design record as step prose, which the task's tick closes whether the parent acted or not — write it as a column-0 `- [ ] Parent: after Task <n>, <the action>` line in the sibling design file, which check-unfinished-work holds open (skills/flow/brainstorm-planner.md)"))
		}
	})
	return v
}

func (p *psPlan) checkTask(found tcfTaskBody) []psFinding {
	id, body, start := found.id, found.lines, found.taskLine
	fields := p.parse(id)
	if fields.unclosedFenceLine != 0 && p.on("F3b") {
		return []psFinding{p.at(fields.unclosedFenceLine, "F3b", "task "+id+" opens a code fence here that is never closed in its body, so every field below it is inside the fence and was not read")}
	}
	var v []psFinding
	seen := map[string]int{}
	tcfUnfenced(body, func(off int, line string) bool {
		m := tcfFieldRE.FindStringSubmatch(line)
		if m == nil || !slices.Contains(psF1Fields, m[1]) {
			return false
		}
		if first, ok := seen[m[1]]; ok && p.on("F1") {
			v = append(v, p.at(start+off+1, "F1", fmt.Sprintf("task %s declares a second **%s:** line (first at line %d); the parser keeps the first occurrence and ignores this one", id, m[1], first)))
		} else {
			seen[m[1]] = start + off + 1
		}
		return false
	})
	tcfUnfenced(body, func(off int, line string) bool {
		if m := psIndentedFieldRE.FindStringSubmatch(line); m != nil && p.on("F3a") {
			v = append(v, p.at(start+off+1, "F3a", "task "+id+" carries a **"+m[1]+":** line indented past column 0, invisible to the field grammar (FIELD_RE anchors at column 0) and therefore never read as a declaration"))
		}
		return false
	})
	tcfUnfenced(body, func(off int, line string) bool {
		m := tcfFieldRE.FindStringSubmatch(line)
		if m == nil || m[1] != "Size" {
			return false
		}
		if sz := tcfStrip(m[2]); p.on("F32") && !slices.Contains(pdbSizes, sz) {
			v = append(v, p.at(start+off+1, "F32", "task "+id+"'s **Size:** value `"+sz+"` is not `small`, `medium` or `big` — plan-dispatch-groups cannot weigh it and plan-dispatch-bundles refuses it"))
		}
		return true
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
			v = append(v, p.at(start+off+1, "F7", fmt.Sprintf("task %s declares a second **After:** line (first at line %d); select_after keeps the first", id, firstGating)))
		} else {
			firstGating = start + off + 1
		}
		return false
	})
	tcfUnfenced(body, func(off int, line string) bool {
		if psCountsTestUnsafely(line) && p.on("F13") {
			v = append(v, p.at(start+off+1, "F13", "task "+id+"'s measured: command counts `@Test` with `grep -c` or `grep -o` and no `-w`, so `@Testcontainers` and `@TestConfiguration` lines count too — use a word-boundary-safe counter such as `grep -cw '@Test'` or `grep -ow '@Test' | wc -l` (the **Baseline convention:** line)"))
		}
		return false
	})

	// F8 -- read from tcfSelectAfter itself, never a second loop deciding
	// which line the field is; the refusal names the accepted values so
	// the first rewrite is informed (KAN-605).
	if a := tcfSelectAfter(body); a != nil && !a.gates && p.on("F8") {
		v = append(v, p.at(start+a.offset+1, "F8", "task "+id+" has a malformed **After:** value: "+a.value+" — an **After:** value is `Task <ids>` or `none`"))
	}
	if len(fields.files) == 0 && p.on("F2") {
		v = append(v, p.at(found.taskLine, "F2", "task "+id+" has no **Files:** line, or one naming no file at all"))
	}
	planning := func(f string) bool {
		return slices.ContainsFunc(psPlanning, func(pfx string) bool { return strings.HasPrefix(f, pfx) })
	}
	if len(fields.files) > 0 && !slices.ContainsFunc(fields.files, func(f string) bool { return !planning(f) }) && p.on("F14") {
		v = append(v, p.at(found.taskLine, "F14", "task "+id+"'s **Files:** all lie under the planning paths (spectre/changes/, openspec/changes/, docs/superpowers/), so its commit cannot carry a `Task-Id:` — name the file the task's work lands in, or make the record a planning commit with no task"))
	}
	joined := tcfBodyFields(body)
	for _, name := range []string{"Files", "Allowed-collateral"} {
		value := strings.Join(joined[name], " ")
		if value != "" && !tcfBacktickRE.MatchString(value) && !tcfNoneOpenRE.MatchString(value) && psAnyWSRE.MatchString(value) && p.on("F16") {
			line := psFieldLine(body, start, name)
			if line == 0 {
				line = found.taskLine
			}
			v = append(v, p.at(line, "F16", "task "+id+"'s **"+name+":** value names no backticked path but carries a comma or whitespace — the commit guard reads it as one path; backtick each path"))
		}
		for _, f := range tcfBacktickTokens(value) {
			if planning(f) && p.on("F23") {
				v = append(v, p.at(psFieldLine(body, start, name), "F23", "task "+id+"'s **"+name+":** names the planning path `"+f+"` — it rides a planning or `Flow-Key: flow.visual-verify` commit, never a task commit (check-task-commit-planning-paths); drop it from the task's fields"))
			}
		}
	}

	for _, c := range psMeasuredCommentRE.FindAllString(strings.Join(joined["Baseline"], " "), -1) {
		if !tcfMeasuredRE.MatchString(c) && p.on("F20") {
			v = append(v, p.at(psFieldLine(body, start, "Baseline"), "F20", "task "+id+"'s **Baseline:** carries a `<!-- measured: … -->` comment with no ` @ <ref>`, so no command is read from it and the re-measurement never runs — write `<!-- measured: <command> @ <ref> -->`"))
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
		v = append(v, p.at(testsLine, "F4", "task "+id+"'s **Tests:** field opens with `none` but **Files:** names a test-shaped path — a contradiction between the opt-out and the declared files; a task repairing existing tests and adding none says so with `Tests: none — repairs existing tests: <names>` (KAN-794)"))
	}
	if !noneOpen && len(fields.tests) == 0 && p.on("F6") {
		line := testsLine
		if line == 0 {
			line = found.taskLine
		}
		v = append(v, p.at(line, "F6", "task "+id+"'s **Tests:** field declares nothing — it opens with neither `none`, a `Case N` label, nor a backtick-quoted test name, so nothing is verified"))
	}
	return v
}

// psMeasuredCommentRE is any `<!-- measured: … -->` comment, read or not;
// F20 holds each one against tcfMeasuredRE.
var psMeasuredCommentRE = regexp.MustCompile(`<!--[` + tcfWS + `]*measured:.*?-->`)

// baselineConvention is F12: the declaring task is the first whose
// non-fenced body carries a Baseline field line; the stated convention is
// any non-fenced line of the file opening as psBaselineConventionRE.
func (p *psPlan) baselineConvention() []psFinding {
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
	return []psFinding{p.at(line, "F12", "task "+declaring+" declares **Baseline:** fields but the plan states no **Baseline convention:** header line — one stated convention is what keeps later tasks from drifting between counting methods; seed the `> **Baseline convention:**` line (skills/flow/brainstorm-planner.md)")}
}

// baselineChain is F17: per measured: command, each task's Baseline before
// against the after of the last task seen under that command.
func (p *psPlan) baselineChain() []psFinding {
	type seen struct {
		id      string
		after   *big.Int
		partner []string
	}
	last := map[string]seen{}
	var v []psFinding
	for _, task := range tcfIterTasks(p.lines) {
		if tcfUnclosedFence(task.lines) >= 0 {
			continue // F3b: this task's fields are unread
		}
		t := p.parse(task.id)
		if t.baseline == nil {
			continue
		}
		for _, cmd := range t.baselineMeasured {
			if prev, ok := last[cmd]; ok && t.baseline[0].Cmp(prev.after) < 0 &&
				!slices.Contains(t.squashPartners, prev.id) && !slices.Contains(prev.partner, task.id) && p.on("F17") {
				v = append(v, p.at(psFieldLine(task.lines, task.taskLine, "Baseline"), "F17", fmt.Sprintf("task %s's **Baseline:** before=%s is below task %s's after=%s under the same measured: command — chain-position absolutes chain in plan order; restate before from task %s's after", task.id, t.baseline[0], prev.id, prev.after, prev.id)))
			}
			last[cmd] = seen{task.id, t.baseline[1], t.squashPartners}
		}
	}
	return v
}

// fileOwnership is F11. Owners come from tcfParseTask's files, one task at a
// time in plan order, each id once, keyed by the task's repository heading
// and the path; an ordering is a gating declared After edge between the pair
// in either direction.
func (p *psPlan) fileOwnership() []psFinding {
	type file struct{ repo, name string }
	owners := map[file][]string{}
	var order []file
	repoAt := pdgRepositories(p.env, p.path, p.lines)
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
			f := file{repoAt[task.taskLine], name}
			if _, ok := owners[f]; !ok {
				order = append(order, f)
			}
			// A repeated token within one task's own list is one owner: a
			// pair needs two tasks.
			if !slices.Contains(owners[f], task.id) {
				owners[f] = append(owners[f], task.id)
			}
		}
		if a := tcfSelectAfter(task.lines); a != nil && a.gates {
			declared[task.id] = a.ids
		}
		if l := psFieldLine(task.lines, task.taskLine, "Files"); l != 0 {
			filesLines[task.id] = l
		}
	}
	var v []psFinding
	for _, f := range order {
		ids := owners[f]
		for i := 1; i < len(ids); i++ {
			later := ids[i]
			for _, earlier := range ids[:i] {
				unordered := !psReaches(declared, later, earlier) && !psReaches(declared, earlier, later)
				if unordered && p.on("F11") {
					line := filesLines[later]
					if line == 0 {
						line = 1
					}
					v = append(v, p.at(line, "F11", "task "+later+"'s **Files:** also names `"+f.name+"`, first named by task "+earlier+", and no **After:** ordering is declared between the two — two tasks editing one file is the shared-file conflict this check exists to make explicit"))
				}
			}
		}
	}
	return v
}

// psReaches answers whether to is reachable from from along declared After
// edges, depth-first with a visited set.
func psReaches(declared map[string][]string, from, to string) bool {
	seen := map[string]bool{}
	stack := []string{from}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, next := range declared[n] {
			if next == to {
				return true
			}
			if !seen[next] {
				seen[next] = true
				stack = append(stack, next)
			}
		}
	}
	return false
}

// afterReferences is F9 and F10 over the After fields: F9 a gating id
// naming no task, F10 the first back-edge of the resolved graph. A task
// whose fence never closes is skipped for both.
func (p *psPlan) afterReferences() []psFinding {
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

	var v []psFinding
	for _, id := range all {
		d := declared[id]
		if d == nil || !d.gates {
			continue
		}
		for _, name := range d.ids {
			if !idSet[name] {
				v = append(v, p.at(p.afterFieldLine(id), "F9", "task "+id+"'s **After:** names task "+name+", which this plan does not define"))
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
					v = append(v, p.at(anchor, "F10", strings.Join(parts, " -> ")+": the resolved After graph has a cycle, so neither bundle can ever dispatch"))
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

// baseRev is the change's base: the merge base of HEAD with
// branch.<current>.flowBase (its origin/ copy preferred) or else with
// refs/remotes/origin/HEAD, and run, a git call in the plan's repository
// root. rev is empty when the plan sits outside a git repository or no base
// resolves.
func (p *psPlan) baseRev() (run func(args ...string) (string, bool), rev string) {
	git := envGit(p.env)
	inDir := func(dir string, args ...string) (string, bool) {
		out, err := git(append([]string{"-C", dir, "-c", "core.quotePath=false"}, args...)...).Output()
		return string(out), err == nil
	}
	root, ok := inDir(psDir(filepath.Clean(psAbs(p.env, p.path))), "rev-parse", "--show-toplevel")
	if !ok {
		return nil, ""
	}
	root = tcfStrip(root)
	run = func(args ...string) (string, bool) { return inDir(root, args...) }
	branch, _ := run("branch", "--show-current")
	recorded := ""
	if branch = tcfStrip(branch); branch != "" {
		recorded, _ = run("config", "--get", "branch."+branch+".flowBase")
		recorded = tcfStrip(recorded)
	}
	candidates := []string{"refs/remotes/origin/HEAD"}
	if recorded != "" {
		candidates = []string{"refs/remotes/origin/" + recorded, recorded}
	}
	for _, ref := range candidates {
		if out, ok := run("merge-base", "HEAD", ref); ok {
			return run, tcfStrip(out)
		}
	}
	return run, ""
}

// baseDeclarations is F15's (file, name) test declarations in the
// test-shaped files of the change's base tree; none when no base resolves.
func (p *psPlan) baseDeclarations() []psDecl {
	run, rev := p.baseRev()
	if rev == "" {
		return nil
	}
	args := []string{"grep", "-I", "-z", "-E", "(func|fun|def)[[:space:]]", rev, "--"}
	for _, g := range psTestShaped {
		args = append(args, ":(glob)**/"+g)
	}
	out, _ := run(args...)
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

// sweepScope is F18. A ticked task's sweep already ran, and a later commit
// may have emptied the file, so it is not read. A file a task at or before
// the sweeping task lists is skipped: that task may remove the string first.
// ponytail: one-line sweeps, single-quoted -e strings and plain or ./ paths
// only (a quoted or :(glob) path is never matched) -- the shape
// skills/flow/brainstorm-planner.md prescribes; "later" is document order,
// not **After:** order; and an earlier task listing the file without
// removing the string is a missed finding. Upgrade: read each earlier
// task's diff when a plan hits one.
func (p *psPlan) sweepScope() []psFinding {
	if !p.on("F18") {
		return nil
	}
	tasks := tcfIterTasks(p.lines)
	type sweep struct {
		task        tcfTaskBody
		line        int
		paths, pats []string
		own         []string
	}
	var sweeps []sweep
	files := map[string][]string{}
	for _, task := range tasks {
		if tcfUnclosedFence(task.lines) >= 0 {
			continue // F3b: this task's fields are unread
		}
		files[task.id] = p.parse(task.id).files
		if tcfTaskLineRE.FindStringSubmatch(p.lines[task.taskLine-1])[1] == "x" {
			continue
		}
		tcfUnfenced(task.lines, func(off int, l string) bool {
			for _, m := range psSweepRE.FindAllStringSubmatch(l, -1) {
				var pats []string
				for _, e := range psSweepPatRE.FindAllStringSubmatch(m[1], -1) {
					pats = append(pats, e[1])
				}
				if len(pats) > 0 {
					sweeps = append(sweeps, sweep{task, task.taskLine + off + 1, strings.FieldsFunc(m[2], tcfIsSpace), pats, files[task.id]})
				}
			}
			return false
		})
	}
	if len(sweeps) == 0 {
		return nil
	}
	run, rev := p.baseRev()
	if rev == "" {
		return nil
	}
	covers := func(paths []string, f string) bool {
		return slices.ContainsFunc(paths, func(s string) bool {
			s = strings.TrimSuffix(strings.TrimPrefix(s, "./"), "/")
			return s == "." || s == ":/" || f == s || strings.HasPrefix(f, s+"/")
		})
	}
	var v []psFinding
	for _, s := range sweeps {
		owner := map[string]string{}
		earlier := map[string]bool{}
		later := false
		for _, task := range tasks {
			if !later {
				for _, f := range files[task.id] {
					earlier[f] = true
				}
				later = task.id == s.task.id
				continue
			}
			for _, f := range files[task.id] {
				if _, ok := owner[f]; !ok && !earlier[f] && covers(s.paths, f) {
					owner[f] = task.id
				}
			}
		}
		if len(owner) == 0 {
			continue
		}
		args := []string{"grep", "-l", "-F"}
		for _, pat := range s.pats {
			args = append(args, "-e", pat)
		}
		args = append(args, rev, "--")
		for f := range owner {
			args = append(args, f)
		}
		out, _ := run(args...)
		var held []string
		for _, l := range tcfSplitLines(out) {
			if f := strings.TrimPrefix(l, rev+":"); owner[f] != "" {
				held = append(held, f)
			}
		}
		sort.Strings(held)
		for _, f := range held {
			v = append(v, p.at(s.line, "F18", "task "+s.task.id+"'s absence sweep searches "+f+", which task "+owner[f]+" owns and which still holds a swept string at the base — the sweep cannot print nothing at this task's commit; scope it to the paths this task touches"))
		}
	}
	return v
}

// baseDeclared is F15.
func (p *psPlan) baseDeclared(ids []string) []psFinding {
	decls := p.baseDeclarations()
	if len(decls) == 0 {
		return nil
	}
	var v []psFinding
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
			v = append(v, p.at(psFieldLine(found.lines, found.taskLine, "Tests"), "F15", "task "+id+"'s **Tests:** names `"+spec.label+"`, which the base already declares outside its **Files:** — "+strings.Join(where, "; ")+"; the task's commit cannot add it, so check-task-commit-fields refuses the commit — name the test this task adds, or list the file in **Files:**"))
		}
	}
	return v
}

// psStatusRE is an entry's **Status:** line; psListItemRE a list item;
// psSectionRE a heading that ends a `##` section.
var (
	psStatusRE   = regexp.MustCompile(`^\*\*Status:\*\*[` + tcfWS + `]*(.*)$`)
	psListItemRE = regexp.MustCompile(`^[` + tcfWS + `]*(?:[-*+]|[0-9]+\.)[` + tcfWS + `]+\S`)
	psSectionRE  = regexp.MustCompile(`^#{1,2}[` + tcfWS + `]`)
)

// verbatimMoves is F28: an unticked task editing run-loaded prose needs the
// change's verbatim-moves.txt, written with the plan.
func (p *psPlan) verbatimMoves() []psFinding {
	if !p.on("F28") {
		return nil
	}
	if _, err := os.Stat(psAbs(p.env, filepath.Join(filepath.Dir(p.path), "verbatim-moves.txt"))); err == nil {
		return nil
	}
	var v []psFinding
	for _, task := range tcfIterTasks(p.lines) {
		if !strings.HasPrefix(p.lines[task.taskLine-1], "- [ ]") {
			continue
		}
		t, _ := tcfParseTask(p.lines, task.id)
		for _, f := range t.files {
			if vmScoped(f) && !strings.HasSuffix(f, "-rationale.md") {
				v = append(v, p.at(task.taskLine, "F28", "task "+task.id+" edits run-loaded prose (`"+f+"`) and this change has no verbatim-moves.txt — write it with the plan, listing every sentence the tasks add or reword, or `#` comment lines alone when they only move text verbatim (.flow/project.md, check-verbatim-moves)"))
				break
			}
		}
	}
	return v
}

// openQuestions is F27: the sibling design.md records no question left open.
func (p *psPlan) openQuestions() []psFinding {
	if !p.on("F27") {
		return nil
	}
	design := filepath.Join(filepath.Dir(p.path), "design.md")
	raw, err := os.ReadFile(psAbs(p.env, design))
	if err != nil {
		return nil
	}
	const fix = "planning leaves nothing open — ask it again, narrowed, and record the answer under ## Decisions (skills/spectre-planning/SKILL.md)"
	var v []psFinding
	in, fenced, title, line, answered := false, false, "", 0, false
	flush := func() {
		if title != "" && !answered {
			v = append(v, p.atFile(design, line, "F27", fmt.Sprintf("open question %q is not answered by a decision — %s", title, fix)))
		}
		title = ""
	}
	for i, l := range strings.Split(string(raw), "\n") {
		if tcfFenceRE.MatchString(l) {
			fenced = !fenced
			continue
		}
		switch {
		case fenced:
		case psSectionRE.MatchString(l):
			flush()
			in = strings.TrimSpace(l) == "## Open questions"
		case !in:
		case strings.HasPrefix(l, "### "):
			flush()
			title, line, answered = strings.TrimSpace(l[4:]), i+1, false
		case title != "":
			if m := psStatusRE.FindStringSubmatch(l); m != nil {
				answered = strings.HasPrefix(m[1], "answered by")
			}
		case psListItemRE.MatchString(l):
			v = append(v, p.atFile(design, i+1, "F27", fmt.Sprintf("open question %q is recorded as a list item — %s", strings.TrimSpace(l), fix)))
		}
	}
	flush()
	return v
}
