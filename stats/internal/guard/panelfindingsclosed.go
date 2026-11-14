package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/tweety53/agents/stats/internal/records"
)

// The contract below moved here from the header of the deleted
// scripts/check-panel-findings-closed.sh shim (KAN-1058).
//
// flow-guard check-panel-findings-closed <worktree> <change-name>
//
// THE CHANGE NAME IS A SECOND ARGUMENT for the same reason it is one on
// flow-guard check-panel-reproducers: a change's findings are rows in the store,
// keyed by change name, not a file at a fixed path. `flow record findings
// -change <name> -C <worktree>` answers with the decoded JSON array of that
// change's findings, and the guard reads that array plus the change's
// dispatch rows the same way (`flow record dispatches -change <name> -C
// <worktree>`), the second read answering only the fixed-without-clean-rerun
// class below.
//
// This is the gate design.md's `gate-is-a-guard` decision chose: nothing
// checked, before this guard existed, that a review panel actually closed
// every finding it verified — `flow-guard check-unfinished-work` only ever caught
// the omission at /flow's integrate gate, after the work was done. This
// guard runs at the panel's own close, immediately before
// `flow stage end -command '/flow' -stage flow.review-panel`.
//
// THE OPEN-FINDING PREDICATE IS DUPLICATED, on purpose, from
// flow-guard check-unfinished-work's own copy — design.md's `duplicate-the-predicate`
// decision, following that guard's own precedent for why a one-line
// predicate gains nothing from being centralized and loses the property that
// both harnesses assert the same shape.
//
// A BARE `withdrawn` IS OPEN, not closed (KAN-791): the word with no reason
// after it is the silent drop the finding-status contract forbids — a
// finding leaves the board only into `fixed` or `withdrawn <reason>` — so
// both copies of the predicate require the reason, and a reasonless
// withdrawal reports exit 1 like any other open finding.
//
// THE GUARD NEVER CONSULTS THE JOURNAL — design.md's `no-journal-excuse`
// decision. `flow record status` never blocks, so a store outage journals a
// close instead of landing it, and a finding whose close only reached the
// journal still reads `open` here and reports exit 1. That is the honest
// verdict; the existing handback is where the operator resolves it.
//
// A GATED PER-TASK REVIEWER'S ROW (slot `gated-reviewer`) IS NO PANEL
// FINDING, and every check here skips it: a clean task re-review closes it,
// not a panel re-run (records.GatedReviewerSlot).
//
// Exit codes:
//   0  no finding is open or deferred, and every Critical or Important
//      recorded `fixed` has a clean re-run dispatch of its slot in a later
//      round whose panel-report-<round>-<id>.md is written — for a row whose
//      `supersedes` names a row of the same round (review-panel.md's
//      cross-role fold), the named row's slot covers it; a row of slot
//      `mutation` needs none, Mutation never re-running after a fix
//   1  a finding is open or `deferred`, or a Critical or Important is
//      recorded `fixed` with no clean re-run dispatch of its slot in any
//      later round — the ordering review-panel.md's **Recording findings**
//      requires, violated, the store's only witness of that re-run being
//      the slot's own dispatch row (KAN-770); each still-open ref and each
//      unverified ref with its slot is named on stderr; or a role's
//      panel-report-<round>-<id>[-<repo>].md files, summed, count on their
//      Critical: <n> and Important: <n> lines more findings than the store
//      holds for that round and role; or a report's `Unread:` line names a
//      section no written per-repository report
//      panel-report-<round>-<id>-<repo>.md of its round and role read
//      (kan-1096 self-review, F2); or a role's store rows of a round,
//      slot exactly the role and not joined, outnumber what its reports
//      count -- every severity when each report carries a Minor: <n> line,
//      Critical and Important otherwise (kan-1096 self-review, F3); or a re-run answered without writing its
//      panel-report-<round>-<id>[-<repo>].md (absent, empty, the parent's
//      `no verbatim report captured` placeholder, or an `Unread:` line
//      naming a section), named on stderr (KAN-915); a dispatch keyed
//      panel-<round>-<slot>-<repo> owes the -<repo> name;
//      or a completed reviewer dispatch left a role's
//      panel-report-<round>-<id>.md absent or empty (the placeholder
//      accepted); or a `fixed` finding's runnable reproducer resolves in no
//      worktree of the change (kan-1034); or a `fixed` Critical or Important
//      has no `regression-test:` line in any panel-fix-report-*.md, or names
//      a path untracked in every worktree of the change or under
//      .superpowers/ (kan-1019); or a round holding a `fixed` Critical or
//      Important finding has no `flow record mutations` row of that round or
//      the next — check-unfinished-work's predicate (kan-1039)
//   2  cannot answer at all — no worktree, no change name, a change name
//      outside the allowlist, a worktree that is not a directory, the store
//      unreachable, its answer unreadable (empty output included), or an
//      unreadable panel report, or the state record unreadable when a
//      `fixed` finding's reproducer token is absent from the worktree or its
//      regression test untracked there, or an unreadable panel-fix report,
//      or the fix-mutation rows unreadable
// Ported to Go: the body's reasoning for every branch is in
// stats/internal/guard/panelfindingsclosed.go.

// checkPanelFindingsClosed is `flow-guard check-panel-findings-closed`: the comment above is the contract -- exit 0 FINDINGS-CLOSED, 1 a
// violation (each named on stderr), 2 cannot answer. The findings read is
// Env.Findings when set, else `flow record findings`; the dispatches read is
// Env.Dispatches when set, else `flow record dispatches`.
func init() {
	Registry["check-panel-findings-closed"] = checkPanelFindingsClosed
}

const (
	cfcPrefix   = "check-panel-findings-closed: "
	cfcJQFailed = cfcPrefix + "jq failed — cannot determine anything"
	// cfcMutationSlot is the one role never re-dispatched after a fix.
	cfcMutationSlot = "mutation"
)

func checkPanelFindingsClosed(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktreeArg, name := arg(0), arg(1)
	if worktreeArg == "" || !isDir(pcAbs(env, worktreeArg)) {
		shown := worktreeArg
		if shown == "" {
			shown = "<missing>"
		}
		fmt.Fprintf(stderr, "%snot a directory: %s\n", cfcPrefix, shown)
		return 2
	}
	if name == "" {
		fmt.Fprintln(stderr, "usage: flow-guard check-panel-findings-closed <worktree> <change-name>")
		return 2
	}
	// CONTAINMENT: plainChangeName (changename.go) is canonical for why
	// -- the name arrives from a pull-request-editable state file and is
	// passed to `flow record findings -change`.
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", cfcPrefix, name)
		return 2
	}
	// Canonicalised before it is passed to `flow` as `-C`, and refused if it
	// vanished since the directory check -- panelreproducers.go's comment on
	// the same sequence is canonical for both hazards.
	worktree, err := filepath.EvalSymlinks(pcAbs(env, worktreeArg))
	if err != nil {
		fmt.Fprintf(stderr, "%sworktree vanished before it could be resolved: %s\n", cfcPrefix, worktreeArg)
		return 2
	}

	// STDOUT AND STDERR ARE CAPTURED SEPARATELY: `flow` prints diagnostics
	// such as `flow: using FLOW_ADDR=...` to stderr, and folding them into
	// stdout would put a non-JSON line at the head of what is parsed.
	var findingsJSON []byte
	if env.Findings != nil {
		findingsJSON, err = env.Findings(name)
		if err != nil {
			fmt.Fprintf(stderr, "%scannot read findings for '%s' from the store — cannot determine anything: %v\n", cfcPrefix, name, err)
			return 2
		}
	} else {
		var errOut []byte
		var rc int
		findingsJSON, errOut, rc = pcFlow(env, false, "record", "findings", "-change", name, "-C", worktree)
		if rc != 0 {
			fmt.Fprintf(stderr, "%scannot read findings for '%s' from the store — cannot determine anything: %s\n", cfcPrefix, name, strings.TrimRight(string(errOut), "\n"))
			return 2
		}
	}
	findings, ok := cfcParse(findingsJSON)
	if !ok {
		fmt.Fprintln(stderr, cfcJQFailed)
		return 2
	}
	// A gated per-task reviewer's row is no panel finding
	// (records.GatedReviewerSlot): every check below skips it.
	findings = slices.DeleteFunc(findings, func(f cfcFindingRow) bool { return f.slot == records.GatedReviewerSlot })

	// THE DISPATCHES READ CARRIES THE FINDINGS READ'S POSTURE, unchanged:
	// read unconditionally -- a store that cannot answer the correlation
	// below cannot pronounce FINDINGS-CLOSED, the same blindness rule the
	// findings read's exit 2 exists for -- and a failed or unreadable read
	// is exit 2, never "no rows". Env.Dispatches is the hook
	// check-panel-fix-single-dispatch shares, nil meaning the CLI on PATH.
	var dispatchesJSON []byte
	if env.Dispatches != nil {
		dispatchesJSON, err = env.Dispatches(name)
		if err != nil {
			fmt.Fprintf(stderr, "%scannot read dispatches for '%s' from the store — cannot determine anything: %v\n", cfcPrefix, name, err)
			return 2
		}
	} else {
		var errOut []byte
		var rc int
		dispatchesJSON, errOut, rc = pcFlow(env, false, "record", "dispatches", "-change", name, "-C", worktree)
		if rc != 0 {
			fmt.Fprintf(stderr, "%scannot read dispatches for '%s' from the store — cannot determine anything: %s\n", cfcPrefix, name, strings.TrimRight(string(errOut), "\n"))
			return 2
		}
	}
	dispatches, ok := cfcParseDispatches(dispatchesJSON)
	if !ok {
		fmt.Fprintln(stderr, cfcJQFailed)
		return 2
	}

	// An open finding is any finding whose status is neither `fixed` nor a
	// `withdrawn <reason>` value -- and, since KAN-791, a bare `withdrawn`
	// (the word with no reason after it) is open too: a reasonless
	// withdrawal is the silent drop the finding-status contract forbids, so
	// it can no longer read as a closed state here, the same line the
	// store's withdrawnWithoutReason and check-unfinished-work's own copy
	// draw. A `deferred <reason>` row is open the same way: nothing is
	// deferred (KAN-862).
	var open []string
	for _, f := range findings {
		rest, withdrawn := strings.CutPrefix(f.status, "withdrawn")
		if f.status != "fixed" && !(withdrawn && strings.TrimSpace(rest) != "") {
			open = append(open, f.ref)
		}
	}
	violated := false
	if len(open) > 0 {
		fmt.Fprintf(stderr, "%sfinding(s) still open: %s\n", cfcPrefix, strings.Join(open, " "))
		violated = true
	}

	// A FINDING IS RECORDED `fixed` ONLY AFTER THE RE-RUN THAT VERIFIES IT
	// (review-panel.md's **Recording findings**, KAN-770), and the store's
	// only witness of that re-run is the raising slot's own dispatch row at
	// a later round, ended `completed` or `clean`. A `fixed` finding whose slot has
	// none stands verified before its verification existed, which is the
	// defect this class exists to catch: the ordering was luck, not
	// discipline. A slot's re-run qualifies when its row carries role
	// `reviewer`, a key of the shape panel-<round>-... (a handshake retry's
	// trailing -retry tolerated; a panel-fix or task key never matches), a
	// round strictly greater than the finding's, an outcome exactly
	// `completed` or `clean` -- review-panel.md closes a slot `completed`,
	// while `flow record dispatch end` also takes a gated reviewer's
	// `clean`, and 50 panel rows in the store end that way (KAN-887's
	// self-review); a timed-out or never-ended dispatch is not a clean
	// re-run -- and a slot whose `+`-components cover every component of the
	// finding's: a bundled re-run covers its members, a solo re-run of one
	// member does not cover a joined finding. Withdrawn findings claim no
	// verification and check nothing here.
	//
	// A ROW FOLDED ONTO ANOTHER ROLE'S ROW is covered by that row's slot
	// instead: review-panel.md's **A defect raised by several roles is
	// recorded per role and fixed once** links it with -supersedes to a row
	// of the same round and re-runs only the first row's role. A
	// -supersedes naming a row of an earlier round is the same defect
	// re-raised by a later fix round, which still re-runs its own role, so
	// only a same-round target lends its slot.
	byRef := map[string]cfcFindingRow{}
	for _, f := range findings {
		byRef[f.ref] = f
	}
	// A RE-RUN COUNTS ONLY ONCE ITS REPORT FILE IS WRITTEN (KAN-915): a
	// dispatch provides a required component only when
	// panel-report-<round>-<component>.md is non-empty and not the parent's
	// `no verbatim report captured` placeholder. A re-run that would have
	// covered the finding but for a missing report names that file.
	covers := func(slot string, round int) (bool, []string) {
		required := strings.Split(slot, "+")
		var missing []string
		for _, d := range dispatches {
			if d.role != "reviewer" || (d.outcome != "completed" && d.outcome != "clean") || d.round <= round {
				continue
			}
			provides := strings.Split(d.slot, "+")
			all := true
			var lacking []string
			for _, r := range required {
				if !slices.Contains(provides, r) {
					all = false
					break
				}
				if f := d.report(r); !cfcReportWritten(worktree, f) {
					lacking = append(lacking, f)
				}
			}
			if !all {
				continue
			}
			if len(lacking) == 0 {
				return true, nil
			}
			missing = append(missing, lacking...)
		}
		return false, missing
	}
	var unverified, unreported []string
	for _, f := range findings {
		if f.status != "fixed" {
			continue
		}
		// A fixed Minor closes on its fix alone, and so does a fixed surviving
		// mutant: Mutation is never re-dispatched after a fix
		// (review-panel-fix-round.md's **Panel re-runs**). Its proof -- the
		// reproducer exiting non-zero on the fix -- is not in the store, so it
		// is admitted unchecked here; the fix-mutation-row and regression-test
		// checks below still hold it.
		if strings.EqualFold(f.severity, "minor") || f.slot == cfcMutationSlot {
			continue
		}
		covered, missing := covers(f.slot, f.round)
		if t, ok := byRef[f.supersedes]; !covered && ok && f.supersedes != "" && t.round == f.round {
			var more []string
			covered, more = covers(t.slot, f.round)
			missing = append(missing, more...)
		}
		if !covered {
			unverified = append(unverified, f.ref+" ("+f.slot+")")
			for _, m := range missing {
				if !slices.Contains(unreported, m) {
					unreported = append(unreported, m)
				}
			}
		}
	}
	if len(unreported) > 0 {
		fmt.Fprintf(stderr, "%sre-run(s) answered without writing their report file: %s\n", cfcPrefix, strings.Join(unreported, " "))
	}
	if len(unverified) > 0 {
		fmt.Fprintf(stderr, "%sfinding(s) recorded fixed with no clean re-run dispatch of their slot in any later round: %s\n", cfcPrefix, strings.Join(unverified, " "))
		violated = true
	}

	// EVERY COMPLETED REVIEWER DISPATCH LEAVES A REPORT PER ROLE, not only a
	// re-run: each `+`-component of a panel-<round>- reviewer dispatch ended
	// `completed` or `clean` needs a non-empty panel-report-<round>-<id>.md,
	// the parent's `no verbatim report captured` placeholder accepted, as
	// review-panel.md's **Recording findings** prescribes it. A file the
	// re-run line above already names is not named again. kan-1034: F7's
	// round-0 slot left no report and no guard noticed.
	var unwritten []string
	for _, d := range dispatches {
		if d.role != "reviewer" || (d.outcome != "completed" && d.outcome != "clean") || d.round < 0 {
			continue
		}
		for _, id := range strings.Split(d.slot, "+") {
			f := d.report(id)
			if st, err := os.Stat(filepath.Join(worktree, ".superpowers", "sdd", f)); err == nil && st.Size() > 0 {
				continue
			}
			if !slices.Contains(unreported, f) && !slices.Contains(unwritten, f) {
				unwritten = append(unwritten, f)
			}
		}
	}
	if len(unwritten) > 0 {
		fmt.Fprintf(stderr, "%scompleted reviewer dispatch(es) left no report file: %s\n", cfcPrefix, strings.Join(unwritten, " "))
		violated = true
	}

	// A FIXED FINDING'S RUNNABLE REPRODUCER STILL EXISTS: its path token,
	// derived as panelexitcontract.go derives it, resolves in the worktree,
	// else in one of the state record's worktrees (a cross-repo peer's
	// reproducer lives in its own tree, KAN-658) -- the record read only when
	// the worktree lacks a token. The `none` forms and an empty reproducer
	// claim nothing runnable; an open finding is the open line's. kan-1034:
	// F7's reproducer was lost with its throwaway copy.
	var lost []string
	var trees []string
	treesRead := false
	// recordedTrees reads the state record's worktrees once, on first need;
	// false means it printed why and the guard exits 2.
	recordedTrees := func(why string) bool {
		if treesRead {
			return true
		}
		treesRead = true
		out, errOut, rc := pcFlow(env, false, "state", "get", "-C", worktree, name)
		var state map[string]json.RawMessage
		if rc != 0 || json.Unmarshal(bytes.TrimRight(out, "\n"), &state) != nil || state == nil {
			fmt.Fprintf(stderr, "%scannot read the state record for '%s' while %s — cannot determine anything: %s\n", cfcPrefix, name, why, strings.TrimRight(string(errOut), "\n"))
			return false
		}
		recorded, ok := pcWorktreeKeys(state["worktrees"])
		if !ok {
			fmt.Fprintln(stderr, cfcJQFailed)
			return false
		}
		trees = resolvedRecordedTrees(env, recorded)
		return true
	}
	for _, f := range findings {
		r := f.reproducer
		if f.status != "fixed" || r == "" || isNoneReproducer(r) {
			continue
		}
		firstLine, _, _ := strings.Cut(r, "\n")
		token := reproducerToken(firstLine)
		if token == "" || pcExists(worktree+"/"+token) {
			continue
		}
		if !recordedTrees("a fixed finding's reproducer is absent from the worktree") {
			return 2
		}
		if !slices.ContainsFunc(trees, func(t string) bool { return pcExists(t + "/" + token) }) {
			lost = append(lost, f.ref+" ("+token+")")
		}
	}
	if len(lost) > 0 {
		fmt.Fprintf(stderr, "%sfixed finding(s) whose runnable reproducer resolves in no worktree of the change: %s\n", cfcPrefix, strings.Join(lost, " "))
		violated = true
	}

	// A FIXED CRITICAL OR IMPORTANT FINDING LEAVES A TRACKED SUITE TEST: a
	// `regression-test: F<n> — <path> — <test>` line in some
	// panel-fix-report-*.md, its path git-tracked in the worktree or a state-
	// record worktree and outside .superpowers/, or `regression-test: F<n> —
	// none — <reason>` (review-panel.md's REPORT FILE paragraph). A fixed
	// Minor needs none. kan-1019: three fixed Importants' only proof was
	// untracked reproducers, deleted with the worktree.
	regression := map[string][]string{}
	fixReports, _ := filepath.Glob(filepath.Join(worktree, ".superpowers", "sdd", "panel-fix-report-*.md"))
	for _, p := range fixReports {
		body, err := os.ReadFile(p)
		if err != nil {
			fmt.Fprintf(stderr, "%scannot read fix report %s — cannot determine anything: %v\n", cfcPrefix, p, err)
			return 2
		}
		for _, m := range cfcRegressionRE.FindAllStringSubmatch(string(body), -1) {
			path := strings.TrimSpace(m[2])
			if path == "none" && strings.TrimSpace(m[3]) == "" {
				path = "none (no reason)"
			}
			regression[m[1]] = append(regression[m[1]], path)
		}
	}
	// tracked is the path itself, never a pathspec: a directory, `.` or a
	// glob matching tracked files is not a test file.
	tracked := func(tree, path string) bool {
		out, ok := gitOut(env, "-C", tree, "ls-files", "-z", "--", ":(literal)"+path)
		return ok && out == path+"\x00"
	}
	var untested []string
	for _, f := range findings {
		if f.status != "fixed" || !(strings.EqualFold(f.severity, "critical") || strings.EqualFold(f.severity, "important")) {
			continue
		}
		lines := regression[f.ref]
		reason := "no regression-test: line"
		for _, path := range lines {
			if path == "none" {
				reason = ""
				break
			}
			if path == "none (no reason)" {
				reason = "none form with no reason"
				continue
			}
			if strings.HasPrefix(path, ".superpowers/") {
				reason = path + " is under .superpowers/"
				continue
			}
			if tracked(worktree, path) {
				reason = ""
				break
			}
			if !recordedTrees("a fixed finding's regression test is untracked in the worktree") {
				return 2
			}
			if slices.ContainsFunc(trees, func(t string) bool { return tracked(t, path) }) {
				reason = ""
				break
			}
			reason = path + " is not tracked"
		}
		if reason != "" {
			untested = append(untested, f.ref+" ("+reason+")")
		}
	}
	if len(untested) > 0 {
		fmt.Fprintf(stderr, "%sfixed Critical/Important finding(s) with no tracked regression test: %s\n", cfcPrefix, strings.Join(untested, " "))
		violated = true
	}

	// A SLOT'S REPORT COUNTS ITS OWN FINDINGS on one `Critical: <n>` and one
	// `Important: <n>` line (review-panel.md's REPORT FILE paragraph). When
	// the sum exceeds the store's Critical and Important rows of that round
	// whose slot's `+`-components include the report's id, any status, a
	// finding the slot raised never reached the store -- kan-1000's round 11.
	// A report missing either line predates the paragraph and is not compared.
	uncounted, undercounted, err := cfcUncounted(worktree, findings)
	if err != nil {
		fmt.Fprintf(stderr, "%s%v\n", cfcPrefix, err)
		return 2
	}
	if len(uncounted) > 0 {
		fmt.Fprintf(stderr, "%sfinding(s) a slot's report counts that the store does not hold: %s\n", cfcPrefix, strings.Join(uncounted, " "))
		violated = true
	}
	// THE STORE MAY NOT HOLD MORE THAN A ROLE'S REPORTS COUNT (kan-1096
	// self-review, F3): round-0 rows F23, F31, F34 and F35 stood with no
	// report on disk counting them, their four reports lost.
	if len(undercounted) > 0 {
		fmt.Fprintf(stderr, "%sstore row(s) a slot's reports do not count: %s\n", cfcPrefix, strings.Join(undercounted, " "))
		violated = true
	}
	// AN UNREAD SECTION NEVER COUNTS AS READ (kan-1096 self-review, F2): a
	// report whose `Unread:` line names a `# worktree:` section needs a
	// written per-repository report of that section, same round and role,
	// whose own `Unread:` is `none` or absent.
	unread, err := cfcUnreadSections(worktree, findings)
	if err != nil {
		fmt.Fprintf(stderr, "%s%v\n", cfcPrefix, err)
		return 2
	}
	if len(unread) > 0 {
		fmt.Fprintf(stderr, "%sreport(s) leaving a section unread with no per-repository report of it: %s\n", cfcPrefix, strings.Join(unread, " "))
		violated = true
	}
	// A fixed round with no fix-mutation row is check-unfinished-work's
	// predicate, applied here so the omission surfaces while the fix round
	// is still open (kan-1039: it surfaced only at the landing gate).
	rounds, cause := uwUnprovedFixRounds(env, name, worktree, findingsJSON)
	if cause != "" {
		fmt.Fprintln(stderr, cfcPrefix+cause)
		return 2
	}
	if len(rounds) > 0 {
		fmt.Fprintf(stderr, "%sfixed Critical/Important finding(s) in round(s) %s and no fix-mutation row for that fix round — record each fix's fix-mutation: line, or one -mutated none -test <reason> row for a fix that changed nothing executable (skills/flow/review-panel-fix-round.md)\n", cfcPrefix, uwJoinRounds(rounds))
		violated = true
	}
	if violated {
		return 1
	}
	fmt.Fprintln(stdout, "FINDINGS-CLOSED")
	return 0
}

// cfcReport is one panel-report-<round>-<rest>.md under .superpowers/sdd,
// resolved to its role.
type cfcReport struct {
	name, role string
	round      int
	body       []byte
}

// cfcReports reads every panel report under <worktree>'s .superpowers/sdd,
// each resolved to its role per decision f2-report-role (kan-1096
// self-review, F2): <rest> itself when it is a known id, else the longest
// known id <k> with <rest> = <k>-<suffix>, else <rest>. The known ids are
// every store row's slot and each of its `+`-components.
func cfcReports(worktree string, findings []cfcFindingRow) ([]cfcReport, error) {
	var known []string
	for _, f := range findings {
		for _, k := range append(strings.Split(f.slot, "+"), f.slot) {
			if k != "" && !slices.Contains(known, k) {
				known = append(known, k)
			}
		}
	}
	slices.SortFunc(known, func(a, b string) int { return len(b) - len(a) })
	paths, _ := filepath.Glob(filepath.Join(worktree, ".superpowers", "sdd", "panel-report-*.md"))
	var out []cfcReport
	for _, p := range paths {
		m := cfcReportNameRE.FindStringSubmatch(filepath.Base(p))
		if m == nil {
			continue
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("cannot read panel report %s — cannot determine anything: %v", p, err)
		}
		round, _ := strconv.Atoi(m[1])
		role := m[2]
		if !slices.Contains(known, role) {
			for _, k := range known {
				if strings.HasPrefix(role, k+"-") {
					role = k
					break
				}
			}
		}
		out = append(out, cfcReport{name: filepath.Base(p), role: role, round: round, body: body})
	}
	return out, nil
}

// cfcUncounted answers, in the label `panel-report-<round>-<role>.md
// (report <n>, recorded <m>)`, first each round and role whose summed panel
// reports count more findings than the store's rows of that round whose
// slot's `+`-components include the role, then (decision f3-reverse-count,
// kan-1096 self-review F3) each whose store rows of that round with slot
// exactly the role -- any status, joined slots left out, since the
// joined-slot rule re-records them per role -- outnumber what its reports
// count. Only reports carrying both count lines are summed, and a group with
// none is not compared. When every summed report of a group carries a
// `Minor: <n>` line, both directions compare all-severity totals, so a
// severity the parent re-grades moves a row between counts without changing
// the comparison; otherwise both compare Critical plus Important.
func cfcUncounted(worktree string, findings []cfcFindingRow) (over, under []string, err error) {
	reports, err := cfcReports(worktree, findings)
	if err != nil {
		return nil, nil, err
	}
	type group struct {
		round int
		role  string
	}
	// ci counts Critical plus Important, all adds Minor; allMinor is false
	// once a report of the group carries no Minor: line.
	type sum struct {
		ci, all  int
		allMinor bool
	}
	sums := map[group]*sum{}
	var order []group
	for _, r := range reports {
		c, i := cfcCriticalRE.FindSubmatch(r.body), cfcImportantRE.FindSubmatch(r.body)
		if c == nil || i == nil {
			continue
		}
		nc, _ := strconv.Atoi(string(c[1]))
		ni, _ := strconv.Atoi(string(i[1]))
		g := group{r.round, r.role}
		if sums[g] == nil {
			sums[g] = &sum{allMinor: true}
			order = append(order, g)
		}
		sums[g].ci += nc + ni
		sums[g].all += nc + ni
		if m := cfcMinorRE.FindSubmatch(r.body); m != nil {
			nm, _ := strconv.Atoi(string(m[1]))
			sums[g].all += nm
		} else {
			sums[g].allMinor = false
		}
	}
	for _, g := range order {
		s := sums[g]
		n := s.ci
		if s.allMinor {
			n = s.all
		}
		counted := func(f cfcFindingRow) bool {
			return f.round == g.round && (s.allMinor || strings.EqualFold(f.severity, "critical") || strings.EqualFold(f.severity, "important"))
		}
		recorded, exact := 0, 0
		for _, f := range findings {
			if !counted(f) {
				continue
			}
			if slices.Contains(strings.Split(f.slot, "+"), g.role) {
				recorded++
			}
			if f.slot == g.role {
				exact++
			}
		}
		if n > recorded {
			over = append(over, fmt.Sprintf("panel-report-%d-%s.md (report %d, recorded %d)", g.round, g.role, n, recorded))
		}
		if exact > n {
			under = append(under, fmt.Sprintf("panel-report-%d-%s.md (report %d, recorded %d)", g.round, g.role, n, exact))
		}
	}
	return over, under, nil
}

// cfcUnreadSections answers each section a report's `Unread:` line names
// that no written per-repository report of the same round and role read,
// each as `<file> (<repo>)`.
func cfcUnreadSections(worktree string, findings []cfcFindingRow) ([]string, error) {
	reports, err := cfcReports(worktree, findings)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, r := range reports {
		for _, repo := range cfcUnread(r.body) {
			if !cfcReportWritten(worktree, fmt.Sprintf("panel-report-%d-%s-%s.md", r.round, r.role, repo)) {
				out = append(out, r.name+" ("+repo+")")
			}
		}
	}
	return out, nil
}

// cfcUnread is the sections a report's `Unread:` line names; none for
// `Unread: none` or no such line.
func cfcUnread(body []byte) []string {
	m := cfcUnreadRE.FindSubmatch(body)
	if m == nil || strings.TrimSpace(string(m[1])) == "none" {
		return nil
	}
	var out []string
	for _, s := range strings.Split(string(m[1]), ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// cfcReportWritten reports whether a dispatch wrote its report <name>: the
// file is non-empty, does not begin with the parent's placeholder, and its
// `Unread:` line, if any, names no section.
func cfcReportWritten(worktree, name string) bool {
	b, err := os.ReadFile(filepath.Join(worktree, ".superpowers", "sdd", name))
	return err == nil && len(b) > 0 && !bytes.HasPrefix(b, []byte("no verbatim report captured")) && len(cfcUnread(b)) == 0
}

type cfcFindingRow struct {
	ref, status, severity string
	round                 int
	slot, supersedes      string
	reproducer            string
}

// cfcDispatchRow is one row of `flow record dispatches`' array, reduced to
// what the fixed-without-clean-rerun correlation reads. round is parsed from
// the key -- the table carries no round column; the panel keys its slot
// dispatches panel-<round>-<slot> -- and -1 marks a key of another shape, so
// such a row can never satisfy a later-round requirement.
type cfcDispatchRow struct {
	key, role, slot, outcome string
	round                    int
}

// cfcRetryRE is a handshake retry's trailing -retry or -retry-<n>.
var cfcRetryRE = regexp.MustCompile(`-retry(-[0-9]+)?$`)

// report is the file a dispatch owes for component c:
// panel-report-<round>-<c>[-<repo>].md, <repo> the key's remainder after
// panel-<round>-<slot>, a trailing retry dropped (kan-1096 self-review, F2).
func (d cfcDispatchRow) report(c string) string {
	rest, ok := strings.CutPrefix(d.key, fmt.Sprintf("panel-%d-%s", d.round, d.slot))
	if !ok {
		rest = ""
	}
	rest = cfcRetryRE.ReplaceAllString(rest, "")
	return fmt.Sprintf("panel-report-%d-%s%s.md", d.round, c, rest)
}

// cfcPanelRoundRE is the dispatch-key shape the correlation reads. A
// handshake retry's key carries a trailing -retry after the slot, which this
// regexp's prefix match tolerates; panel-fix-<round> and task keys never
// match, panel- not being followed by digits in them.
var cfcPanelRoundRE = regexp.MustCompile(`^panel-([0-9]+)-`)

// cfcReportNameRE matches a slot's panel-report-<round>-<slot>.md file name,
// and cfcCriticalRE / cfcImportantRE read the Critical: / Important: count
// lines of that report, cfcUnreadRE its `Unread:` line.
var (
	cfcUnreadRE     = regexp.MustCompile(`(?m)^Unread: (.*)$`)
	cfcReportNameRE = regexp.MustCompile(`^panel-report-([0-9]+)-(.+)\.md$`)
	cfcCriticalRE   = regexp.MustCompile(`(?m)^Critical: ([0-9]+)\s*$`)
	cfcImportantRE  = regexp.MustCompile(`(?m)^Important: ([0-9]+)\s*$`)
	cfcMinorRE      = regexp.MustCompile(`(?m)^Minor: ([0-9]+)\s*$`)
)

// cfcRegressionRE reads a fix report's `regression-test: F<n> — <path> — …`
// line, bare or as a `- `/`* ` list item: the finding's ref, the path field (`none` in the none form) and the
// rest, which the none form needs non-empty as its reason.
var cfcRegressionRE = regexp.MustCompile(`(?m)^(?:[-*] )?regression-test: (\S+) — ([^—\n]+)(?:— ([^\n]*))?`)

// cfcParse decodes `flow record findings`' array. It refuses, where the
// bash guard's jq failed: a non-array, a non-object row, a non-string
// status. It also refuses empty output, which jq read as zero findings -- a
// store read that printed nothing is not an answer, the same rule
// pcParseFindings applies. A null or absent ref joins as the empty string,
// a scalar ref as its JSON text (jq's join); a non-string severity reads as
// none, and an absent or non-integer round as 0, the store's zero value.
func cfcParse(b []byte) ([]cfcFindingRow, bool) {
	var raw []map[string]json.RawMessage
	if json.Unmarshal(b, &raw) != nil || raw == nil {
		return nil, false
	}
	out := make([]cfcFindingRow, 0, len(raw))
	for _, o := range raw {
		if o == nil || !pcIsString(o["status"]) {
			return nil, false
		}
		var f cfcFindingRow
		_ = json.Unmarshal(o["status"], &f.status)
		switch ref := strings.TrimSpace(string(o["ref"])); {
		case ref == "" || ref == "null":
		case pcIsString(o["ref"]):
			_ = json.Unmarshal(o["ref"], &f.ref)
		case strings.HasPrefix(ref, "{") || strings.HasPrefix(ref, "["):
			return nil, false
		default:
			f.ref = ref
		}
		if pcIsString(o["severity"]) {
			_ = json.Unmarshal(o["severity"], &f.severity)
		}
		if n, err := strconv.Atoi(strings.TrimSpace(string(o["round"]))); err == nil {
			f.round = n
		}
		if pcIsString(o["slot"]) {
			_ = json.Unmarshal(o["slot"], &f.slot)
		}
		if pcIsString(o["supersedes"]) {
			_ = json.Unmarshal(o["supersedes"], &f.supersedes)
		}
		if pcIsString(o["reproducer"]) {
			_ = json.Unmarshal(o["reproducer"], &f.reproducer)
		}
		out = append(out, f)
	}
	return out, true
}

// cfcParseDispatches decodes `flow record dispatches`' array with the same
// posture cfcParse carries: a non-array or a non-object row is jq failing
// (exit 2, cannot answer), and absent or non-string fields read as the empty
// string, the shape a row legitimately carries when its verb's flag was
// omitted -- an empty role, slot or outcome simply never qualifies below.
func cfcParseDispatches(b []byte) ([]cfcDispatchRow, bool) {
	var raw []map[string]json.RawMessage
	if json.Unmarshal(b, &raw) != nil || raw == nil {
		return nil, false
	}
	out := make([]cfcDispatchRow, 0, len(raw))
	for _, o := range raw {
		if o == nil {
			return nil, false
		}
		var d cfcDispatchRow
		for _, pair := range []struct {
			field string
			dest  *string
		}{{"key", &d.key}, {"role", &d.role}, {"slot", &d.slot}, {"outcome", &d.outcome}} {
			if pcIsString(o[pair.field]) {
				_ = json.Unmarshal(o[pair.field], pair.dest)
			}
		}
		d.round = -1
		if m := cfcPanelRoundRE.FindStringSubmatch(d.key); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil {
				d.round = n
			}
		}
		out = append(out, d)
	}
	return out, true
}
