package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// checkUnfinishedWork is scripts/check-unfinished-work.sh: that script's
// header comment is the contract -- one CLEAR/OUTSTANDING verdict line on
// stdout at exit 0, exit 2 when it cannot answer at all. Plan resolution is
// changePlanDir (changeplan.go), never a second copy; the findings read is
// Env.Findings when set, else `flow record findings`; the verdict write and
// the prior-false-positive read are always the `flow` CLI on Env's PATH. The
// reasoning for each branch, moved here from the bash body it replaced
// (3e48ecac), sits beside the code it explains.
func init() {
	Registry["check-unfinished-work"] = checkUnfinishedWork
}

const uwPrefix = "check-unfinished-work: "

func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktree, name, canonical := arg(0), arg(1), arg(2)
	if worktree == "" || name == "" {
		fmt.Fprintln(stderr, "usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]")
		return 2
	}

	// CONTAINMENT: the change name arrives from a pull-request-editable state
	// file and is concatenated into every path below, and passed to `flow
	// record findings -change`. Without this check `../../../planted/clear`
	// makes the gate read a plan outside the worktree entirely and report
	// CLEAR for a change that has no plan of its own.
	//
	// The rule is records.Destination's Protection 1
	// (stats/internal/records/render.go), character for character, and that
	// function's comment is canonical for why each hazard is in it. This is
	// plainChangeName (cleanupcomplete.go) itself now, where the bash guard
	// kept its own `case` copy: TestCheckUnfinishedWork's 8e cases and
	// TestCheckCleanupComplete/12 still assert the same rejected shapes.
	//
	// The bash guard exported LC_ALL=C because `case`'s ranges are
	// collating ranges -- `démo` was rejected under C and accepted under
	// en_US.UTF-8 on bash 3.2. plainChangeName compares bytes, so no locale
	// reaches it; the 8e-i cases pin that under both locales.
	//
	// It also closes the symlink question at these paths: isFile and isDir
	// follow symlinks, so a name that cannot leave the change's own directory
	// is what makes the paths this guard reads the ones it was asked about.
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", uwPrefix, name)
		return 2
	}
	if !isDir(worktree) {
		fmt.Fprintf(stderr, "%s%s is not a directory — cannot determine anything\n", uwPrefix, worktree)
		return 2
	}

	changes := worktree + "/" + specRootLeaf(worktree, stderr) + "/changes"
	var reasons []string

	// unreadable is a file that exists but cannot be read: neither signal but
	// an honest unknown, so this guard refuses rather than guessing. Reading
	// it as zero unticked boxes would fail toward the reassuring answer, the
	// one failure direction this whole guard exists to remove.
	unreadable := func(path string) int {
		fmt.Fprintf(stderr, "%scannot read %s — cannot determine anything\n", uwPrefix, path)
		return 2
	}

	// Signal one — the plan, including any fix sub-change.
	//
	// THE PRIMARY PLAN IS TRACKED SEPARATELY FROM THE REST, because their
	// absences mean opposite things. `<spec root>/changes/<name>/tasks.md` is
	// the change's plan and every flow change has one: missing, it is
	// outstanding like any other missing record, and reporting "every plan
	// item is checked" over a change with no plan at all is the same silent
	// clearance the header rejects. A fix sub-change's plan is genuinely
	// optional -- most changes have none -- so its absence is nothing.
	//
	// THE PRIMARY PLAN IS RESOLVED THROUGH changePlanDir, NOT COMPOSED
	// DIRECTLY, so a satellite's own `link.md` is followed to its canonical
	// `tasks.md` instead of reporting the local path's absence as a verdict
	// (the header's satellite paragraph). The directory, not only the path,
	// because THE FIX-SUB-CHANGE SWEEP BELOW MUST WALK THE SAME TREE THE PLAN
	// RESOLVED FROM: a satellite's `<canonical-id>-fix-N` sibling lives
	// beside the canonical change, and sweeping the satellite's own,
	// plan-less changes/ would report CLEAR over an unchecked fix-round item.
	//
	// changePlanDir answers 1 in two shapes told apart below: a plain change
	// with no plan and no `## Part of` link -- the ordinary missing-plan case
	// -- versus a genuine satellite whose canonical plan could not be
	// reached, which refuses outright. It answers 3 (KAN-267) when the state
	// record resolved the name to more than one project's plan: its lines
	// naming each match are printed and this guard refuses, since an
	// OUTSTANDING built on a guess between two projects' plans is exactly
	// the silent clearance this guard exists to prevent. Its stderr is shown
	// only then, as the bash guard showed its captured stderr.
	var planErr bytes.Buffer
	planDir, planRC := changePlanDir(env, &planErr, worktree, name, canonical)
	if planRC == 3 {
		_, _ = stderr.Write(planErr.Bytes())
		fmt.Fprintf(stderr, "%sthe state record resolves '%s' to more than one project's plan — cannot determine anything\n", uwPrefix, name)
		return 2
	}
	planOpen := 0
	var primary, sweepDir, sweepID string
	if planRC == 0 {
		primary = planDir + "/tasks.md"
		n, ok := uwCountUnticked(primary)
		if !ok {
			return unreadable(primary)
		}
		planOpen += n
		sweepDir, sweepID = gdcDirname(planDir), uwBasename(planDir)
	} else {
		link := changes + "/" + name + "/link.md"
		if uwHasLine(link, "## Part of") {
			if canonical != "" {
				fmt.Fprintf(stderr, "%s'%s' is a satellite change (link.md at %s); its canonical plan was not found under the supplied canonical worktree %s, and a canonical worktree that was supplied is never retried against peers — cannot determine anything\n", uwPrefix, name, link, canonical)
			} else {
				peers := worktree + "/" + specRootLeaf(worktree, stderr) + "/peers"
				fmt.Fprintf(stderr, "%s'%s' is a satellite change (link.md at %s); no canonical worktree was supplied and its canonical plan could not be resolved through %s — cannot determine anything\n", uwPrefix, name, link, peers)
			}
			return 2
		}
		primary = changes + "/" + name + "/tasks.md"
		reasons = append(reasons, "no plan at "+primary)
		sweepDir, sweepID = changes, name
	}

	// THE STORE ANCHOR (KAN-260). The change's records live under the project
	// of the repo that ran /flow -- the repo where the plan lives -- and the
	// project key derives from -C's git common dir
	// (stats/internal/fallback/statefile.go's ProjectKey), so a run in a
	// second repo's worktree reading -C <worktree> would query the SECOND
	// repo's project: a false `[]` on this verdict line, and a verdict
	// recorded where the canonical project's tools never look. Every store
	// call anchors at the resolved plan's directory when the plan is not
	// local, and at this worktree when it is or when nothing resolved -- the
	// no-plan fall-through cannot know another repo exists. -worktree keeps
	// naming the worktree the guard was asked to judge.
	anchor := worktree
	if planDir != "" && planDir != worktree && !strings.HasPrefix(planDir, worktree+"/") {
		anchor = planDir
	}

	// THE SWEEP. Every regular `tasks.md` at any depth under the change's own
	// directory, and under any sibling directory whose name starts
	// `<id>-fix-`, is read: the discovery does not depend on guessing the
	// layout a fix run wrote (nested or beside, a fix of a fix at any depth).
	// What remains undiscoverable is a sub-change whose name has no relation
	// to its parent's, which the contract does not permit. The walk is
	// `find <dir> -type f -name tasks.md`: symlinks are not followed -- the
	// root included, so a symlinked changes/ is swept as empty -- a directory
	// it cannot list is skipped silently, and each path is the root
	// concatenated with its entries, so the primary plan compares equal.
	if fi, err := os.Lstat(sweepDir); err == nil && fi.IsDir() {
		var bad string
		uwFindTasks(sweepDir, func(plan string) bool {
			if plan == primary {
				return true
			}
			if !strings.HasPrefix(plan, sweepDir+"/"+sweepID+"/") && !strings.HasPrefix(plan, sweepDir+"/"+sweepID+"-fix-") {
				return true
			}
			n, ok := uwCountUnticked(plan)
			if !ok {
				bad = plan
				return false
			}
			planOpen += n
			return true
		})
		if bad != "" {
			return unreadable(bad)
		}
	}
	if planOpen > 0 {
		reasons = append(reasons, strconv.Itoa(planOpen)+" unchecked plan item(s)")
	}

	// Signal two — findings whose recorded status is not closed.
	//
	// THIS SIGNAL DOES NOT PARSE A FINDINGS TABLE OR A MARKER BLOCK, and the
	// deletion is the point. Three review passes once hid an open Critical
	// from a hand-rolled GFM table parser six distinct ways, because the
	// guard was recovering one fact, "is any finding still open", by
	// re-implementing a document grammar defined in another file's prose. A
	// finding is a JSON object in the store, with no cells to split and no
	// table boundary to track. Write-time validation keeps every stored
	// status to the closed vocabulary the CLI's validator judges -- and,
	// since KAN-791, the store itself refuses the bare `withdrawn` shape
	// (the one shape an HTTP write or a replay could land below the CLI) --
	// so the guarantee no longer rides on the CLI alone. A bare `withdrawn`
	// that predates that rule, if one exists, is open here: the predicate's
	// own withdrawn test (below) requires the reason. There is no branch for
	// any other malformed status.
	//
	// THE STORE IS QUERIED ONCE, and a failed `flow record findings` is exit
	// 2, "cannot determine anything" -- never OUTSTANDING and never CLEAR. A
	// change the store has never heard of, or one that raised no findings,
	// prints `[]` at exit 0: the zero-findings case, not a refusal.
	//
	// STDOUT AND STDERR ARE READ SEPARATELY, and that separation is the whole
	// point. `flow` writes diagnostics to stderr -- most reliably the `flow:
	// using FLOW_ADDR=...` line it prints whenever the address is overridden,
	// as this repository's ui-test stack instructs -- while the JSON goes to
	// stdout. Folding the two together put that line at the head of the
	// payload and failed the parse on a run that had succeeded: it broke only
	// for the operator who followed the instructions.
	var findingsJSON []byte
	if env.Findings != nil {
		var err error
		if findingsJSON, err = env.Findings(name); err != nil {
			fmt.Fprintf(stderr, "%scannot read findings for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(findingsJSON), "\n"))
			return 2
		}
	} else {
		out, errOut, rc := pcFlow(env, false, "record", "findings", "-change", name, "-C", anchor)
		if rc != 0 {
			fmt.Fprintf(stderr, "%scannot read findings for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(errOut), "\n"))
			return 2
		}
		findingsJSON = out
	}
	// An open finding is any finding whose status is neither `fixed` nor a
	// `withdrawn <reason>` value -- a prefix test covers the family, reason
	// text included, without comparing the reason -- and, since KAN-791, a
	// bare `withdrawn` (the word with no reason after it) is open too: a
	// reasonless withdrawal is the silent drop the contract forbids, so it
	// can no longer read as a closed state. A `deferred <reason>` row is
	// open: nothing is deferred (KAN-862).
	open, ok := uwOpenFindings(findingsJSON)
	if !ok {
		fmt.Fprintln(stderr, uwPrefix+"jq failed — cannot determine anything")
		return 2
	}
	if open > 0 {
		reasons = append(reasons, strconv.Itoa(open)+" open finding(s) in the review panel record")
	}

	// THE VERDICT IS RECORDED, AND PRIOR FALSE POSITIVES ARE ADVISORY. Neither
	// `flow` call below can move the verdict line or the exit code already
	// committed to above: the write's result is ignored, and the
	// prior-false-positive read only ever adds a stderr hint, skipped on any
	// failure or an empty count. That read runs only on OUTSTANDING: on CLEAR
	// there is no breakdown for a prior false positive to be a hint about,
	// and a gate that cannot fire without a store must never block on a hint
	// (design.md's script-reads-are-advisory decision). The write's own
	// stdout and stderr are discarded, so a store outage stays silent.
	verdict := "CLEAR: " + worktree + " — every plan item is checked and no finding is open"
	if len(reasons) > 0 {
		verdict = "OUTSTANDING: " + worktree + " — " + strings.Join(reasons, "; ")
	}
	pcFlow(env, false, "record", "verdict", "-change", name, "-guard", "check-unfinished-work", "-worktree", worktree, "-verdict", verdict, "-C", anchor)
	if len(reasons) > 0 {
		if prior, _, rc := pcFlow(env, false, "record", "verdicts", "-guard", "check-unfinished-work", "-false-positive", "-C", anchor); rc == 0 {
			if n, last, ok := uwPriorFalsePositives(prior); ok && n > 0 {
				fmt.Fprintf(stderr, "%sprior false positives for this guard on this project: %d — last: %s\n", uwPrefix, n, last)
			}
		}
	}
	fmt.Fprintln(stdout, verdict)
	return 0
}

// uwCountUnticked is count_unticked: column-0 `- [ ]` lines (the script
// header says why column 0), false when the file cannot be read. The one
// pattern serves the primary plan and every fix sub-change's plan, so the
// copies cannot drift. The bash guard read with `grep -a` so a stray NUL
// byte could not put grep into binary mode and read a corrupted file as a
// clean one; a byte scan has no binary mode.
func uwCountUnticked(path string) (int, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	n := 0
	for _, l := range bytes.Split(b, []byte("\n")) {
		if bytes.HasPrefix(l, []byte("- [ ]")) {
			n++
		}
	}
	return n, true
}

// uwHasLine is `grep -qxF <line> <file>`: false for an unreadable file.
func uwHasLine(path, line string) bool {
	b, err := os.ReadFile(path)
	if err != nil || !isFile(path) {
		return false
	}
	for _, l := range strings.Split(string(b), "\n") {
		if l == line {
			return true
		}
	}
	return false
}

// uwFindTasks calls visit for every regular file named tasks.md under dir,
// as `find <dir> -type f -name tasks.md 2>/dev/null`; visit returning false
// stops the walk.
func uwFindTasks(dir string, visit func(string) bool) bool {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		p := dir + "/" + e.Name()
		switch {
		case e.Type().IsRegular() && e.Name() == "tasks.md":
			if !visit(p) {
				return false
			}
		case e.IsDir():
			if !uwFindTasks(p, visit) {
				return false
			}
		}
	}
	return true
}

// uwBasename is basename(1) on a path with no trailing slash; dirname(1) is
// gdcDirname.
func uwBasename(p string) string { return p[strings.LastIndexByte(p, '/')+1:] }

// uwOpenFindings is the bash guard's jq count of open findings over one
// array of objects each carrying a string status. Anything else -- empty
// output, an object, a null element, two values -- is refused as jq failing,
// though jq itself read some of these as zero findings (CLEAR) or iterated an
// object's values: the header rules out clearance on input the guard cannot
// read (decided 2026-09-27, kan-778's task 4 Correction).
func uwOpenFindings(b []byte) (int, bool) {
	var all []map[string]json.RawMessage
	if json.Unmarshal(b, &all) != nil || all == nil {
		return 0, false
	}
	n := 0
	for _, f := range all {
		var status string
		if f == nil || !pcIsString(f["status"]) || json.Unmarshal(f["status"], &status) != nil {
			return 0, false
		}
		// Closed is `fixed` or a reasoned `withdrawn` (KAN-791): the bare
		// word -- no reason after it -- is open, the same line the store's
		// withdrawnWithoutReason and check-panel-findings-closed's own copy
		// draw. A `deferred <reason>` row is open: nothing is deferred
		// (KAN-862).
		if status != "fixed" {
			rest, withdrawn := strings.CutPrefix(status, "withdrawn")
			if !withdrawn || strings.TrimSpace(rest) == "" {
				n++
			}
		}
	}
	return n, true
}

// uwPriorFalsePositives is the advisory's two jq reads of `flow record
// verdicts`: the array's length, and the newest entry as
// "<falsePositiveReason> (<change>, <flaggedAt[:10]>)". ok is false where
// the length read fails; an entry the second read cannot render renders
// empty, as a failed command substitution did.
func uwPriorFalsePositives(b []byte) (int, string, bool) {
	var all []json.RawMessage
	if json.Unmarshal(bytes.TrimRight(b, "\n"), &all) != nil || all == nil {
		return 0, "", false
	}
	if len(all) == 0 {
		return 0, "", true
	}
	var first map[string]json.RawMessage
	if json.Unmarshal(all[0], &first) != nil || first == nil {
		return len(all), "", true
	}
	flagged := pcRaw(first["flaggedAt"])
	if pcIsString(first["flaggedAt"]) {
		for i := range flagged {
			if utf8.RuneCountInString(flagged[:i]) == 10 {
				flagged = flagged[:i]
				break
			}
		}
	} else if elems := []json.RawMessage(nil); json.Unmarshal(first["flaggedAt"], &elems) == nil && elems != nil {
		// jq slices an array too, and string interpolation renders the
		// slice as compact JSON.
		b, _ := json.Marshal(elems[:min(len(elems), 10)])
		flagged = string(b)
	} else if flagged != "null" {
		return len(all), "", true // jq cannot slice a number, object or bool
	}
	last := pcRaw(first["falsePositiveReason"]) + " (" + pcRaw(first["change"]) + ", " + flagged + ")"
	return len(all), strings.TrimRight(last, "\n"), true
}
