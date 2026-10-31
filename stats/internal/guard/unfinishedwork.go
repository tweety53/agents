package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tweety53/agents/stats/internal/records"
)

// checkUnfinishedWork is scripts/check-unfinished-work.sh: that script's
// header comment is the contract -- one CLEAR/OUTSTANDING verdict line on
// stdout at exit 0, exit 2 when it cannot answer at all. Plan resolution is
// changePlanDir (changeplan.go), never a second copy; the findings read is
// Env.Findings when set, else `flow record findings`; the verdict write and
// the prior-false-positive read are always the `flow` CLI on Env's PATH. The
// reasoning for each branch, moved here from the bash body it replaced
// (3e48ecac), sits beside the code it explains.
//
// Signal eleven (kan-1044 self-review, F1-c) flags a range commit whose
// `Flow-Key: panel-fix-<r>[-<n>]` trailer no dispatch row carries by key or
// by commit, reading HEAD and the pre-reshape pin when it resolves.
func init() {
	Registry["check-unfinished-work"] = checkUnfinishedWork
}

const uwPrefix = "check-unfinished-work: "

// uwPanelFixKey is a panel-fix dispatch key, panel-fix-<r>[-<n>] (signal eleven).
var uwPanelFixKey = regexp.MustCompile(`^panel-fix-[0-9]+(-[0-9]+)?$`)

// unsplitListHint names the cause of an argument that cannot hold
// whitespace and does (kan-970 self-review): a list built in a variable and
// passed unquoted reaches a guard as one argument under zsh.
const unsplitListHint = "looks like a list passed as one argument — zsh does not word-split an unquoted $VAR; pass each value as its own argument"

func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktree, name, canonical := arg(0), arg(1), arg(2)
	if worktree == "" || name == "" {
		fmt.Fprintln(stderr, "usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree] [<worktree>=<merge-base>]...")
		return 2
	}
	// The change's commit ranges, one per worktree of its resolved set:
	// signal four's commit half searches them for each ticked appended
	// task's `Task-Id:` trailer.
	var ranges [][2]string
	for _, a := range args[min(3, len(args)):] {
		wt, base, ok := strings.Cut(a, "=")
		if !ok || wt == "" || base == "" {
			fmt.Fprintf(stderr, "%s'%s' is not <worktree>=<merge-base>\n", uwPrefix, a)
			return 2
		}
		if strings.ContainsAny(base, " \t\n") {
			fmt.Fprintf(stderr, "%s'%s' is not <worktree>=<merge-base> — %s\n", uwPrefix, a, unsplitListHint)
			return 2
		}
		ranges = append(ranges, [2]string{wt, base})
	}
	// A call passing no range gets this worktree's own, derived as the finish
	// contract derives every range it passes (decision uw-derive-missing-range):
	// a range that cannot be derived leaves ranges empty, exactly as before.
	if len(ranges) == 0 {
		if r, ok := uwDerivedRange(env, worktree); ok {
			ranges = append(ranges, r)
		}
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
	// plainChangeName (changename.go) itself now, where the bash guard
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
	rounds, cause := uwUnprovedFixRounds(env, name, anchor, findingsJSON)
	if cause != "" {
		fmt.Fprintln(stderr, uwPrefix+cause)
		return 2
	}
	if len(rounds) > 0 {
		reasons = append(reasons, "fixed Critical/Important finding(s) in round(s) "+uwJoinRounds(rounds)+" and no fix-mutation row for that fix round — record each fix's fix-mutation: line, or one `-mutated none -test <reason>` row for a fix that changed nothing executable (skills/flow/review-panel-fix-round.md) before landing")
	}
	// Signal two's report half (kan-1015): a slot's report counting more
	// Critical and Important findings than the store holds is a finding that
	// never reached the store -- check-panel-findings-closed's count, read
	// once, by the worktree holding the plan.
	if anchor == worktree {
		rows, ok := cfcParse(findingsJSON)
		if !ok {
			fmt.Fprintln(stderr, uwPrefix+"jq failed — cannot determine anything")
			return 2
		}
		uncounted, err := cfcUncounted(anchor, rows)
		if err != nil {
			fmt.Fprintf(stderr, "%s%v\n", uwPrefix, err)
			return 2
		}
		if len(uncounted) > 0 {
			reasons = append(reasons, strconv.Itoa(len(uncounted))+" panel report(s) count findings the store does not hold ("+strings.Join(uncounted, " ")+") — record each with `flow record finding -change "+name+" -C "+anchor+"` before landing")
		}
	}

	// Signal three — dispatch rows never closed with an outcome (kan-929: a
	// `dispatch end` run without -C in a second repo's worktree journalled
	// against the wrong project, another dropped -outcome, a third was
	// never sent, and the ledger read "Outcome: not recorded" with nothing
	// stopping integrate). Read and refused exactly like signal two. The
	// rows are the change's, not a worktree's, and the gate runs once per
	// worktree: only the worktree holding the plan (anchor == worktree)
	// reads them, so a multi-repo change lists each open row once.
	if anchor == worktree {
		dispOut, dispErr, dispRC := pcFlow(env, false, "record", "dispatches", "-change", name, "-C", anchor)
		if dispRC != 0 {
			fmt.Fprintf(stderr, "%scannot read dispatches for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(dispErr), "\n"))
			return 2
		}
		var dispatches []struct {
			Seq          int       `json:"seq"`
			Key          string    `json:"key"`
			TaskID       string    `json:"taskId"`
			Role         string    `json:"role"`
			Outcome      string    `json:"outcome"`
			SessionToken string    `json:"sessionToken"`
			CommitSha    string    `json:"commitSha"`
			StartedAt    time.Time `json:"startedAt"`
		}
		if json.Unmarshal(dispOut, &dispatches) != nil || dispatches == nil {
			fmt.Fprintf(stderr, "%sdispatches for '%s' are not one JSON array — cannot determine anything\n", uwPrefix, name)
			return 2
		}
		// Each row is closed under the token of the run that BEGAN it, not
		// the closing run's own (kan-931: a token filled in by the integrate
		// run answered 404), so every command carries the row's own token.
		var unclosed, closeCmds []string
		for _, d := range dispatches {
			if strings.TrimSpace(d.Outcome) == "" {
				unclosed = append(unclosed, d.Key)
				// A legacy row with no key or no token cannot be matched by
				// dispatch end at all: say so rather than print a command
				// whose next flag the shell would read as the missing value.
				if d.Key == "" || d.SessionToken == "" {
					closeCmds = append(closeCmds, "row seq "+strconv.Itoa(d.Seq)+" has no key or session token and cannot be closed by `flow record dispatch end`")
					continue
				}
				cmd := "flow record dispatch end -change " + name + " -key " + d.Key + " -session-token " + d.SessionToken + " -outcome <outcome> -C " + anchor + " -ended-at <when its work ended>"
				if d.Role == "implementer" {
					cmd += " -commit <its task commit>"
				}
				closeCmds = append(closeCmds, "`"+cmd+"`")
			}
		}
		if len(unclosed) > 0 {
			reasons = append(reasons, strconv.Itoa(len(unclosed))+" dispatch row(s) with no outcome ("+strings.Join(unclosed, ", ")+") — close each with "+strings.Join(closeCmds, ", "))
		}
		// SIGNAL FIVE — a panel round dispatched with no pass row (kan-939
		// self-review: round 1 ran panel-1-primary+principles and recorded
		// nothing, and the gate cleared it). Every panel run and fix-round
		// pass records pass rows (review-panel.md), so a round some
		// `panel-<round>-…` reviewer row ran in and no pass row carries is
		// unfinished. Pass rows are read only when such a round exists, and
		// a failed read is refused like the dispatch read above.
		rounds := map[int][]string{}
		for _, d := range dispatches {
			if n, ok := records.PanelDispatchRound(records.Dispatch{Key: d.Key, Role: d.Role}); ok {
				rounds[n] = append(rounds[n], d.Key)
			}
		}
		if len(rounds) > 0 {
			passOut, passErr, passRC := pcFlow(env, false, "record", "passes", "-change", name, "-C", anchor)
			if passRC != 0 {
				fmt.Fprintf(stderr, "%scannot read pass rows for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(passErr), "\n"))
				return 2
			}
			var passes []struct {
				Round int `json:"round"`
			}
			if json.Unmarshal(passOut, &passes) != nil || passes == nil {
				fmt.Fprintf(stderr, "%spass rows for '%s' are not one JSON array — cannot determine anything\n", uwPrefix, name)
				return 2
			}
			for _, p := range passes {
				delete(rounds, p.Round)
			}
			if len(rounds) > 0 {
				var nums []int
				for n := range rounds {
					nums = append(nums, n)
				}
				slices.Sort(nums)
				var named []string
				for _, n := range nums {
					named = append(named, "round "+strconv.Itoa(n)+": "+strings.Join(rounds[n], ", "))
				}
				reasons = append(reasons, strconv.Itoa(len(nums))+" panel round(s) dispatched with no pass row ("+strings.Join(named, "; ")+") — record each round's pass rows with `flow record pass -change "+name+" -round <round>` before landing")
			}
		}
		// SIGNAL SIX — no change summary recorded (kan-957 self-review: 35 of
		// 36 gymie changes landed with no summary row although pipeline.md
		// prescribes one per run). The summary is read through the
		// self-review bundle, the one read of it the CLI has: its header —
		// the lines above its first `## ` source section — carries a `found:`
		// line and names the summary `skipped: change summary (absent)` when the
		// store holds none. A failed read, or a payload whose header carries
		// no `found:` line, is refused like the dispatch read above.
		bundleOut, bundleErr, bundleRC := pcFlow(env, false, "self-review", "bundle", "-change", name, "-C", anchor)
		if bundleRC != 0 {
			fmt.Fprintf(stderr, "%scannot read the self-review bundle for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(bundleErr), "\n"))
			return 2
		}
		var headerLines []string
		for _, l := range strings.Split(string(bundleOut), "\n") {
			if strings.HasPrefix(l, "## ") {
				break
			}
			headerLines = append(headerLines, l)
		}
		if !slices.ContainsFunc(headerLines, func(l string) bool { return strings.HasPrefix(l, "found: ") }) {
			fmt.Fprintf(stderr, "%sthe self-review bundle for '%s' carries no found: line — cannot determine anything\n", uwPrefix, name)
			return 2
		}
		if slices.Contains(headerLines, "skipped: change summary (absent)") {
			reasons = append(reasons, "no change summary recorded — record the run's summary with `flow record summary -change "+name+" -C "+anchor+" -file -` before landing")
		}
		// SIGNAL SEVEN — a code commit no completed inline verify row covers
		// (decision uw-verify-row-after-code): the newest commit outside
		// `<spec root>/changes/` in any range, authored after the newest
		// completed `verify`/`verify-…` verifier row, or with none at all —
		// run 1's reshaped `Flow-Key: flow.commit-two` commit excepted. A
		// re-run after the drop is judged from the drop commit on, so only
		// a fix run's own code asks for a newer verify row. A `skipped` row counts as completed: it only ends a test
		// group whose every command was unchanged since a green clean run
		// (decision f1-skipped-is-verified, kan-1037).
		if len(ranges) > 0 {
			var lastVerify int64 = -1
			for _, d := range dispatches {
				if d.Role == "verifier" && (d.Outcome == "completed" || d.Outcome == "skipped") && (d.Key == "verify" || strings.HasPrefix(d.Key, "verify-")) {
					lastVerify = max(lastVerify, d.StartedAt.Unix())
				}
			}
			newest, newestAt := "", int64(-1)
			for _, r := range ranges {
				leaf := specRootLeaf(r[0], io.Discard)
				out, ok := capture(envGit(env)("-C", r[0], "log", "--no-merges", "--format=%h %at %(trailers:key=Flow-Key,valueonly,separator=%x20)", r[1]+"..HEAD", "--", ".", ":(exclude)"+leaf+"/changes/"))
				if !ok {
					fmt.Fprintf(stderr, "%scannot read the commits of %s..HEAD in %s — cannot determine anything\n", uwPrefix, r[1], r[0])
					return 2
				}
				for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
					sha, rest, _ := strings.Cut(l, " ")
					at, flowKey, _ := strings.Cut(rest, " ")
					// Run 1's reshaped implementation commit carries a fresh author
					// date; it reshapes verified code and is not new code.
					if strings.Contains(flowKey, "flow.commit-two") {
						continue
					}
					if n, err := strconv.ParseInt(at, 10, 64); err == nil && n > newestAt {
						newest, newestAt = sha, n
					}
				}
			}
			if newest != "" && newestAt > lastVerify {
				reasons = append(reasons, "no completed inline verify row since "+newest+" — run Inline verify (skills/flow/verify-and-handoff.md) and record its row before landing — a re-run after a new commit takes a fresh `verify-<k>` key (the store keeps a reused key's first start)")
			}
			// SIGNAL EIGHT — a code commit and no review-panel pass at all
			// (kan-970 self-review: four implementer groups and the panel ran
			// unrecorded, and the gate cleared the change). Every panel run
			// records pass rows (review-panel.md), so a change with a commit
			// outside `<spec root>/changes/` and none was never reviewed as far
			// as the store can tell. A change with a panel reviewer row is
			// signal five's to judge; the pass rows are read only otherwise,
			// and a failed read is refused like signal five's.
			if newest != "" && len(rounds) == 0 {
				passOut, passErr, passRC := pcFlow(env, false, "record", "passes", "-change", name, "-C", anchor)
				if passRC != 0 {
					fmt.Fprintf(stderr, "%scannot read pass rows for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(passErr), "\n"))
					return 2
				}
				var passes []json.RawMessage
				if json.Unmarshal(passOut, &passes) != nil || passes == nil {
					fmt.Fprintf(stderr, "%spass rows for '%s' are not one JSON array — cannot determine anything\n", uwPrefix, name)
					return 2
				}
				if len(passes) == 0 {
					reasons = append(reasons, "no review-panel pass recorded for code commit "+newest+" — run the review panel (skills/flow/review-panel.md) and record its pass rows with `flow record pass -change "+name+" -round <round>` before landing")
				}
			}
		}
		// SIGNAL ELEVEN — a panel-fix commit with no dispatch row (kan-1044
		// self-review, F1-c: three panel fixes landed under trailers keyed a
		// round low, and no row carried them). Each `Flow-Key: panel-fix-<r>[-<n>]`
		// trailer in a range needs a row of the same key or one carrying its
		// commit -- a trailer one round low is carried by the row its repair
		// back-fills under the real round's key on that commit. HEAD and the
		// pre-reshape pin are both read, since after the reshape the change's
		// earlier commits are reachable from the pin alone.
		if len(ranges) > 0 {
			keys := map[string]bool{}
			for _, d := range dispatches {
				keys[d.Key] = true
				if d.CommitSha != "" {
					keys[d.CommitSha] = true
				}
			}
			var missing []string
			for _, r := range ranges {
				tips := []string{"HEAD"}
				if _, ok := capture(envGit(env)("-C", r[0], "rev-parse", "-q", "--verify", "refs/flow/pre-reshape/"+name)); ok {
					tips = append(tips, "refs/flow/pre-reshape/"+name)
				}
				out, ok := capture(envGit(env)(append([]string{"-C", r[0], "log", "--no-merges", "--format=%H %h %(trailers:key=Flow-Key,valueonly,separator=%x20)", "^" + r[1]}, tips...)...))
				if !ok {
					fmt.Fprintf(stderr, "%scannot read the commits of %s..%s in %s — cannot determine anything\n", uwPrefix, r[1], strings.Join(tips, " "), r[0])
					return 2
				}
				for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
					full, rest, _ := strings.Cut(l, " ")
					sha, rest, _ := strings.Cut(rest, " ")
					for _, k := range strings.Fields(rest) {
						if uwPanelFixKey.MatchString(k) && !keys[k] && !keys[full] {
							missing = append(missing, sha+" ("+k+")")
						}
					}
				}
			}
			if len(missing) > 0 {
				reasons = append(reasons, strconv.Itoa(len(missing))+" panel-fix commit(s) whose Flow-Key no dispatch row carries ("+strings.Join(missing, ", ")+") — record each with `flow record dispatch backfill -change "+name+" -role panel-fix -agent-id inline -commit <sha>` under its real round's key before landing")
			}
		}
		// SIGNAL NINE — a verify-report failure no later commit addressed
		// (decision verify-failure-signal, kan-994 self-review: a failed
		// capture a later green run hid reached landing untouched). Every
		// `<spec>.spec.ts:<line>` in a `.superpowers/sdd/verify-report-*.md`
		// `- verify: exit <n≠0>` item is a failure, its capture the first
		// `<name>.png` after it; it is addressed by a range commit at or after
		// the report's mtime touching that capture (or, naming none, the
		// spec), or by a narrative line citing a range commit. Archived plans
		// are read too.
		if len(ranges) > 0 {
			reports, _ := filepath.Glob(worktree + "/.superpowers/sdd/verify-report-*.md")
			type uwCommit struct {
				sha   string
				at    int64
				bases []string
			}
			var commits []uwCommit
			if len(reports) > 0 {
				for _, r := range ranges {
					out, ok := capture(envGit(env)("-C", r[0], "log", "--no-merges", "--format=%x00%H %ct", "--name-only", r[1]+"..HEAD"))
					if !ok {
						fmt.Fprintf(stderr, "%scannot read the commits of %s..HEAD in %s — cannot determine anything\n", uwPrefix, r[1], r[0])
						return 2
					}
					for _, chunk := range strings.Split(out, "\x00")[1:] {
						lines := strings.Split(strings.TrimSpace(chunk), "\n")
						sha, at, _ := strings.Cut(lines[0], " ")
						n, _ := strconv.ParseInt(at, 10, 64)
						c := uwCommit{sha: sha, at: n}
						for _, l := range lines[1:] {
							if l = strings.TrimSpace(l); l != "" {
								c.bases = append(c.bases, uwBasename(l))
							}
						}
						commits = append(commits, c)
					}
				}
			}
			narrative, _ := os.ReadFile(filepath.Dir(primary) + "/narrative.md")
			var flagged []string
			flaggedRead := false
			// The reports in mtime order: a report whose next one passed a
			// rerun of its failed specs is addressed (kan-1022's self-review).
			bodies, infos := make([]string, len(reports)), make([]os.FileInfo, len(reports))
			for i, rp := range reports {
				b, err := os.ReadFile(rp)
				if err != nil {
					return unreadable(rp)
				}
				st, err := os.Stat(rp)
				if err != nil {
					return unreadable(rp)
				}
				bodies[i], infos[i] = string(b), st
			}
			order := make([]int, len(reports))
			for i := range order {
				order[i] = i
			}
			slices.SortStableFunc(order, func(a, b int) int { return infos[a].ModTime().Compare(infos[b].ModTime()) })
			for k, i := range order {
				rp, st := reports[i], infos[i]
				rerunPassed := k+1 < len(order) && uwRerunPassed.MatchString(bodies[order[k+1]])
				for _, fl := range uwVerifyFailures(bodies[i]) {
					addressed := rerunPassed
					for _, c := range commits {
						if c.at < st.ModTime().Unix() {
							continue
						}
						for _, base := range c.bases {
							if (fl.png != "" && (base == fl.png+".png" || strings.HasPrefix(base, fl.png+"-") && strings.HasSuffix(base, ".png"))) ||
								(fl.png == "" && base == uwBasename(fl.spec)) {
								addressed = true
							}
						}
					}
					for _, l := range strings.Split(string(narrative), "\n") {
						_, cited, ok := strings.Cut(l, "Verify failure addressed: "+fl.spec+":"+fl.line+" — ")
						cited = strings.TrimSpace(strings.Trim(strings.TrimSpace(cited), "`"))
						if !ok || len(cited) < 7 {
							continue
						}
						for _, c := range commits {
							if strings.HasPrefix(c.sha, cited) {
								addressed = true
							}
						}
					}
					// A failure an earlier false-positive verdict of this
					// change names by its `<spec>:<line>` is cleared: the
					// flag lands on that run's newest row only, so without
					// this every later run re-flags it (kan-1010's
					// self-review: screenshots.spec.ts:52, three times).
					if !addressed {
						if !flaggedRead {
							flagged, flaggedRead = uwFlaggedReasons(env, name, anchor), true
						}
						for _, r := range flagged {
							if uwNamesFailure(r, fl.spec+":"+fl.line) {
								addressed = true
							}
						}
					}
					if !addressed {
						what := fl.spec + ":" + fl.line
						if fl.png != "" {
							what += " " + fl.png + ".png"
						}
						reasons = append(reasons, "verify failure not addressed: "+what+" ("+uwBasename(rp)+") — fix it in this run, or cite its fixing commit as a narrative line \"Verify failure addressed: "+fl.spec+":"+fl.line+" — <sha>\"")
					}
				}
			}
		}
		// SIGNAL TEN — a multi-repository change whose merge-order record
		// misses a repository (decision merge-order-signal, kan-995
		// self-review). Every range whose worktree's main checkout differs
		// from the plan worktree's must be named, in backticks, by a numbered
		// line of link.md's `## Merge order` section whose peer name the
		// spectre/peers file resolves to that main checkout. A single-range
		// or single-repository call reads nothing; an absent link.md holds
		// no entries.
		if planRC == 0 && len(ranges) > 1 {
			if own, ok := changePlanMainCheckout(env, worktree); ok {
				own = filepath.Clean(own)
				link := planDir + "/link.md"
				var named []string
				if b, err := os.ReadFile(link); err == nil {
					in := false
					for _, l := range strings.Split(string(b), "\n") {
						if strings.HasPrefix(l, "## ") {
							in = strings.TrimSpace(l) == "## Merge order"
							continue
						}
						if m := uwMergeOrderEntryRE.FindStringSubmatch(l); in && m != nil {
							if root, ok := changePlanPeerRoot(env, worktree, specRootLeaf(worktree, io.Discard), m[1]); ok {
								named = append(named, filepath.Clean(root))
							}
						}
					}
				} else if !os.IsNotExist(err) {
					return unreadable(link)
				}
				seen := map[string]bool{}
				for _, r := range ranges {
					mc, ok := changePlanMainCheckout(env, r[0])
					if !ok {
						continue
					}
					mc = filepath.Clean(mc)
					if mc == own || seen[mc] || slices.Contains(named, mc) {
						continue
					}
					seen[mc] = true
					reasons = append(reasons, "merge order does not name "+mc+": "+link+" — write the change's ## Merge order per skills/flow/cross-repo-worktrees.md (\"Then make the merge-order record cover the whole set\") before landing")
				}
			}
		}
		// THE VISUAL VERDICT IS FOLDED IN (decision uw-fold-visual-verdict):
		// check-visual-verify-dispatched over this guard's own worktree's
		// range, its MISSING line a reason here and its refusal this guard's.
		for _, r := range ranges {
			if r[0] != worktree {
				continue
			}
			var vOut, vErr bytes.Buffer
			switch checkVisualVerifyDispatched([]string{worktree, name, r[1], canonical}, env, &vOut, &vErr) {
			case 1:
				if line := strings.TrimSpace(vOut.String()); !uwVisualFlagged(env, name, worktree, anchor, line) {
					reasons = append(reasons, line)
				}
			case 2:
				fmt.Fprintf(stderr, "%s%s\n", uwPrefix, strings.TrimRight(vErr.String(), "\n"))
				return 2
			}
		}
		// SIGNAL FOUR — a ticked appended task no dispatch row names: a fix
		// round that recorded nothing (kan-979's fix-round Task 7 reached
		// integrate with no dispatch row at all). A row names a task by its
		// taskId or by the `+`-joined ids of a task-<ids>[-…] key. Only the tasks a
		// fix run appended — the last `**Tasks appended:** <n>` of the plan,
		// and every `In-run fix <k> — …` task (uwAppendedIn) — are judged: each is implemented alone, so it carries a row of its
		// own, while plan-time tasks run grouped under one task-<n> key or
		// inline with one row per bundle and most have none. It needs the
		// plan read above. An appended task also needs a `reviewer` row whose
		// seq is greater than its earliest non-reviewer row's — the fix run's
		// panel names no task, so any panel reviewer row (key `panel-…`) of the change counts, and
		// the earliest row is the anchor so a Minor-only round's inline
		// `panel-fix` row never demands a re-review.
		if planRC == 0 {
			first := map[string]int{}
			lastReview := -1
			for _, d := range dispatches {
				if d.Role == "reviewer" {
					if strings.HasPrefix(d.Key, "panel-") {
						lastReview = max(lastReview, d.Seq)
					}
					continue
				}
				for _, id := range uwRowTasks(d.Key, d.TaskID) {
					if s, seen := first[id]; id != "" && (!seen || d.Seq < s) {
						first[id] = d.Seq
					}
				}
			}
			appended, ok := uwAppended(primary, true)
			if !ok {
				return unreadable(primary)
			}
			var missing, unreviewed []string
			for _, id := range appended {
				if s, seen := first[id]; !seen {
					missing = append(missing, id)
				} else if lastReview <= s {
					unreviewed = append(unreviewed, id)
				}
			}
			if len(missing) > 0 {
				reasons = append(reasons, strconv.Itoa(len(missing))+" ticked appended task(s) with no dispatch row (Task "+strings.Join(missing, ", ")+") — record each task's dispatch, and its fix round's pass rows and findings, before landing")
			}
			if len(unreviewed) > 0 {
				reasons = append(reasons, strconv.Itoa(len(unreviewed))+" ticked appended task(s) with no panel reviewer row (key `panel-<round>-<slot>`) begun after its implementer row (Task "+strings.Join(unreviewed, ", ")+") — run its fix round's panel and record its pass rows and findings, before landing")
			}
			// A dispatch row is not a commit: a row back-filled after the
			// fact names a task whose work never landed. Each ticked
			// appended task also needs a commit carrying `Task-Id: <n>` in
			// one of the change's ranges -- all of them, since a cross-repo
			// change's fix task may commit in a peer repository.
			if len(appended) > 0 {
				if len(ranges) == 0 {
					fmt.Fprintf(stderr, "%sticked appended task(s) (Task %s) and no <worktree>=<merge-base> range was passed — cannot judge their commits\n", uwPrefix, strings.Join(appended, ", "))
					return 2
				}
				committed := map[string]bool{}
				for _, r := range ranges {
					out, ok := capture(envGit(env)("-C", r[0], "log", "--no-merges", "--format=%(trailers:key=Task-Id,valueonly)", r[1]+"..HEAD"))
					if !ok {
						fmt.Fprintf(stderr, "%scannot read the commits of %s..HEAD in %s — cannot determine anything\n", uwPrefix, r[1], r[0])
						return 2
					}
					for _, id := range ctTaskNumberRE.FindAllString(out, -1) {
						committed[id] = true
					}
					// A dropped re-run's range starts at the drop commit: run 1
					// squashed the earlier task commits, trailers and all, and
					// its gate already judged every task ticked in the plan
					// as it stood there.
					if rel, err := filepath.Rel(r[0], primary); err == nil && !strings.HasPrefix(rel, "..") {
						if old, ok := capture(envGit(env)("-C", r[0], "show", r[1]+":./"+filepath.ToSlash(rel))); ok {
							for _, id := range uwAppendedIn(old, true) {
								committed[id] = true
							}
						}
					}
				}
				var uncommitted []string
				for _, id := range appended {
					if !committed[id] {
						uncommitted = append(uncommitted, id)
					}
				}
				if len(uncommitted) > 0 {
					reasons = append(reasons, strconv.Itoa(len(uncommitted))+" ticked appended task(s) with no commit carrying its `Task-Id:` trailer on the branch (Task "+strings.Join(uncommitted, ", ")+") — commit each task's work with `Task-Id: <n>` before landing")
				}
			}
		}
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

// uwDerivedRange is `<worktree>=<merge-base>` for a call that passed no
// range: the base resolve-base-branch names, then `git merge-base HEAD
// origin/<base>`. False on any failure.
func uwDerivedRange(env Env, worktree string) ([2]string, bool) {
	var out bytes.Buffer
	if resolveBaseBranch([]string{worktree}, env, &out, io.Discard) != 0 {
		return [2]string{}, false
	}
	base, ok := capture(envGit(env)("-C", worktree, "merge-base", "HEAD", "origin/"+strings.TrimSpace(out.String())))
	if !ok || strings.TrimSpace(base) == "" {
		return [2]string{}, false
	}
	return [2]string{worktree, strings.TrimSpace(base)}, true
}

var (
	uwTask          = regexp.MustCompile(`^- \[([ xX])\] ([0-9]+)\. `)
	uwTasksAppended = regexp.MustCompile(`^(?:> )?\*\*Tasks appended:\*\* ([0-9]+)`)
	// uwInRunFix is verify-fix-loop.md's in-run fix task title, appended
	// without raising **Tasks appended:** (kan-941: tasks 12-16 unjudged).
	uwInRunFix = regexp.MustCompile(`^- \[[ xX]\] [0-9]+\. In-run fix [0-9]+ — `)
	uwTaskKey  = regexp.MustCompile(`^task-([0-9]+(?:\+[0-9]+)*)(?:-|$)`)
)

// uwRowTasks is the task ids a dispatch row names: its taskId, and the
// `+`-joined ids of a task-<ids>[-…] key. Signal four and close-task's
// close-time warning share it, so the two cannot drift.
func uwRowTasks(key, taskID string) []string {
	ids := []string{taskID}
	if m := uwTaskKey.FindStringSubmatch(key); m != nil {
		ids = append(ids, strings.Split(m[1], "+")...)
	}
	return ids
}

// uwAppended returns the ids of plan's appended tasks — the last n
// column-0 tasks, n from the `**Tasks appended:** <n>` header line
// (blockquoted or not), 0 when absent — in plan order, only the ticked ones
// when tickedOnly, whatever rows name them: the caller judges coverage.
// False when the plan cannot be read.
func uwAppended(plan string, tickedOnly bool) ([]string, bool) {
	b, err := os.ReadFile(plan)
	if err != nil {
		return nil, false
	}
	return uwAppendedIn(string(b), tickedOnly), true
}

// uwAppendedIn is uwAppended over a plan's text. A task titled
// `In-run fix <k> — …` (verify-fix-loop.md step 1) counts as appended
// wherever it sits, since that loop never raises **Tasks appended:**.
func uwAppendedIn(plan string, tickedOnly bool) []string {
	appended := 0
	var tasks [][]string
	var inRun []bool
	for l := range strings.Lines(plan) {
		if m := uwTasksAppended.FindStringSubmatch(l); m != nil {
			appended, _ = strconv.Atoi(m[1])
		} else if m := uwTask.FindStringSubmatch(l); m != nil {
			tasks = append(tasks, m)
			inRun = append(inRun, uwInRunFix.MatchString(l))
		}
	}
	var ids []string
	for i, m := range tasks {
		if (i >= len(tasks)-appended || inRun[i]) && (!tickedOnly || m[1] != " ") {
			ids = append(ids, m[2])
		}
	}
	return ids
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

// uwFlaggedReasons is the reason of every false-positive verdict the store
// holds for change under check-unfinished-work. A failed or unparseable
// read is none: the read can only clear an item, so a store outage leaves
// every item outstanding, never refused (script-reads-are-advisory).
func uwFlaggedReasons(env Env, change, anchor string) []string {
	out, _, rc := pcFlow(env, false, "record", "verdicts", "-guard", "check-unfinished-work", "-false-positive", "-C", anchor)
	var rows []struct {
		Change string `json:"change"`
		Reason string `json:"falsePositiveReason"`
	}
	if rc != 0 || json.Unmarshal(out, &rows) != nil {
		return nil
	}
	var reasons []string
	for _, r := range rows {
		if r.Change == change {
			reasons = append(reasons, r.Reason)
		}
	}
	return reasons
}

// uwNamesFailure reports whether reason names spec:line as a whole token:
// no path character before it and no digit after, so `a.spec.ts:5` names
// neither `ba.spec.ts:5` nor `a.spec.ts:52`.
func uwNamesFailure(reason, specLine string) bool {
	for i := 0; ; {
		j := strings.Index(reason[i:], specLine)
		if j < 0 {
			return false
		}
		start, end := i+j, i+j+len(specLine)
		before := start == 0 || !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_./-", rune(reason[start-1]))
		after := end == len(reason) || reason[end] < '0' || reason[end] > '9'
		if before && after {
			return true
		}
		i = start + 1
	}
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

// uwVisualFlagged reports whether an operator flagged the folded
// VISUAL-VERIFY-MISSING line a false positive: a check-visual-verify-dispatched
// verdict row of this change on this worktree whose verdict ends with the
// line's own " — HEAD <sha>" — integrate-change.sh's visual_false_positive, so a
// flag that passes integrate's visual gate does not stop this one. Any read
// failure is not flagged.
func uwVisualFlagged(env Env, name, worktree, anchor, line string) bool {
	i := strings.LastIndex(line, " — HEAD ")
	if i < 0 {
		return false
	}
	head := line[i:]
	out, _, rc := pcFlow(env, false, "record", "verdicts", "-guard", "check-visual-verify-dispatched", "-false-positive", "-C", anchor)
	if rc != 0 {
		return false
	}
	var rows []struct {
		Change, Worktree, Verdict string
		FalsePositive             bool
	}
	if json.Unmarshal(out, &rows) != nil {
		return false
	}
	wt := worktree
	if p, err := filepath.EvalSymlinks(worktree); err == nil {
		wt = p
	}
	for _, r := range rows {
		if r.Change == name && r.FalsePositive && r.Worktree == wt && strings.HasSuffix(r.Verdict, head) {
			return true
		}
	}
	return false
}

// uwMergeOrderEntryRE is one numbered `## Merge order` entry of link.md,
// its backticked peer name captured (signal ten).
var uwMergeOrderEntryRE = regexp.MustCompile("^\\s*\\d+\\.\\s+`([^`]+)`")

// uwVerifyFailure is one failure signal nine reads: a spec, its line, and
// the capture base name (no `.png`) it names, if any.
type uwVerifyFailure struct{ spec, line, png string }

var (
	// A report's `- verify:` item passing a rerun of the failed specs.
	uwRerunPassed = regexp.MustCompile(`(?m)^- verify: exit 0 — rerun of\b`)
	uwSpecRE      = regexp.MustCompile("([^\\s:(\\[`'\"]+\\.spec\\.ts):(\\d+)")
	uwPngRE       = regexp.MustCompile(`([A-Za-z0-9_.-]+)\.png`)
)

// uwVerifyFailures reads a verify report's failed `- verify: exit <n>` item
// (n not 0) and its whitespace-indented continuation lines, up to the next
// `- ` item: every `<spec>.spec.ts:<line>` there, with the first `<name>.png`
// after it and before the next spec.
func uwVerifyFailures(report string) []uwVerifyFailure {
	var out []uwVerifyFailure
	lines := strings.Split(report, "\n")
	for i, l := range lines {
		code, ok := strings.CutPrefix(strings.TrimSpace(l), "- verify: exit ")
		if !ok || !strings.HasPrefix(l, "- ") || strings.TrimSpace(code) == "0" {
			continue
		}
		var body []string
		for _, c := range lines[i+1:] {
			if c == "" || (c[0] != ' ' && c[0] != '\t') {
				break
			}
			body = append(body, c)
		}
		text := strings.Join(body, "\n")
		ms := uwSpecRE.FindAllStringSubmatchIndex(text, -1)
		for k, m := range ms {
			end := len(text)
			if k+1 < len(ms) {
				end = ms[k+1][0]
			}
			fl := uwVerifyFailure{spec: text[m[2]:m[3]], line: text[m[4]:m[5]]}
			if p := uwPngRE.FindStringSubmatch(text[m[1]:end]); p != nil {
				fl.png = p[1]
			}
			out = append(out, fl)
		}
	}
	return out
}

// uwUnprovedFixRounds is the rounds, sorted, holding a fixed Critical or
// Important finding no fix-mutation row proves. A fixed Critical or
// Important finding is proved by its fix round's fix-mutation: rows
// (skills/flow/review-panel-fix-round.md; a `none — <reason>` row counts,
// and a fix round that changed nothing executable still records one). A
// Minor-only round is exempt (review-panel.md). The store holds round N's
// rows under N or N+1, so either satisfies round N; a withdrawn finding
// needs none. cause is non-empty when the findings or the mutation rows
// cannot be read: the message, without its guard's prefix. Shared by
// check-unfinished-work and check-panel-findings-closed. A gated reviewer's
// row (records.GatedReviewerSlot) records no fix-mutation row and is skipped.
func uwUnprovedFixRounds(env Env, name, anchor string, findingsJSON []byte) (rounds []int, cause string) {
	var fixedRows []struct {
		Round    int    `json:"round"`
		Status   string `json:"status"`
		Severity string `json:"severity"`
		Slot     string `json:"slot"`
	}
	if json.Unmarshal(findingsJSON, &fixedRows) != nil {
		return nil, "jq failed — cannot determine anything"
	}
	fixedRounds := map[int]bool{}
	for _, r := range fixedRows {
		if sev := strings.ToLower(r.Severity); r.Slot != records.GatedReviewerSlot && r.Status == "fixed" && (sev == "critical" || sev == "important") {
			fixedRounds[r.Round] = true
		}
	}
	if len(fixedRounds) == 0 {
		return nil, ""
	}
	mutOut, mutErr, mutRC := pcFlow(env, false, "record", "mutations", "-change", name, "-C", anchor)
	if mutRC != 0 {
		return nil, fmt.Sprintf("cannot read fix-mutation rows for '%s' from the store — cannot determine anything: %s", name, strings.TrimRight(string(mutErr), "\n"))
	}
	var muts []struct {
		Round int `json:"round"`
	}
	if json.Unmarshal(mutOut, &muts) != nil || muts == nil {
		return nil, fmt.Sprintf("fix-mutation rows for '%s' are not one JSON array — cannot determine anything", name)
	}
	for _, m := range muts {
		delete(fixedRounds, m.Round)
		delete(fixedRounds, m.Round-1)
	}
	return slices.Sorted(maps.Keys(fixedRounds)), ""
}

func uwJoinRounds(rounds []int) string {
	var nums []string
	for _, n := range rounds {
		nums = append(nums, strconv.Itoa(n))
	}
	return strings.Join(nums, ", ")
}
