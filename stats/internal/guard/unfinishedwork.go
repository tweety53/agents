package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/gitops"
	"github.com/tweety53/agents/stats/internal/guard/plans"
	"github.com/tweety53/agents/stats/internal/records"
)

// The contract below moved here from the header of the deleted
// scripts/check-unfinished-work.sh shim (KAN-1058).
//
// flow-guard check-unfinished-work — report whether a change carries unfinished work,
// for /flow's integrate-run gate.
//
// Usage: flow-guard check-unfinished-work <worktree> <change-name> [canonical-worktree] [<worktree>=<merge-base>]...
//
// Prints ONE verdict line to stdout:
//   CLEAR: <reason>          nothing outstanding
//   OUTSTANDING: <breakdown> one or more signals fired
//
// Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at all —
// an unreadable worktree, a change name outside the allowlist, or a file that
// exists but cannot be read. The VERDICT carries the answer, not the exit
// status — see flow-guard check-finish-preflight's header for why this repository
// separates them. Finish runs this once per worktree, so each verdict names the
// worktree it judged; a bare CLEAR would be unattributable across several.
//
// A MISSING FILE COUNTS AS OUTSTANDING, NOT CLEAR. Treating an absent plan or
// an absent finding record as clearance is the failure this guard exists to
// prevent: a branch merged over unfinished work because nothing was written
// down.
//
// SIGNAL TWO: AN OPEN FINDING OF SEVERITY CRITICAL OR IMPORTANT IS
// UNFINISHED; an open Minor is not counted.
//
// SIGNAL TWO READS THE STORE, NOT A RENDERED FILE. A change's findings are
// rows in the store, keyed by change name, and `flow record findings -change
// <name> -C <worktree>` answers with the decoded JSON array — the same verb
// flow-guard check-panel-reproducers reads. There is no longer a rendered panel record
// for this guard to resolve a path to, and no missing-file case for it either:
// a change the store has never heard of, or one that genuinely raised no
// findings, answers `[]` at exit 0, which is CLEAR on this signal.
//
// A BARE `withdrawn` IS OPEN, not closed (KAN-791), the same line
// check-panel-findings-closed's duplicate of this predicate draws: the word
// with no reason after it is the reasonless drop the finding-status contract
// forbids, so only `withdrawn <reason>` counts as closed on this signal.
//
// A CALL PASSING NO RANGE GETS ITS OWN WORKTREE'S RANGE DERIVED (kan-989
// self-review): the base flow-guard resolve-base-branch names, then `git -C <worktree>
// merge-base HEAD origin/<base>`. A range that cannot be derived — no `origin`,
// no git — leaves the call exactly as one passing none.
//
// SIGNAL SEVEN: A CODE COMMIT NO COMPLETED INLINE VERIFY ROW COVERS IS
// UNFINISHED (kan-989 self-review). The newest commit in any range touching a
// path outside `<spec root>/changes/`, authored after the newest `startedAt`
// of the `verifier` rows with outcome `completed` and key `verify` or
// `verify-…` — or with no such row at all — names its short sha in the
// breakdown. A re-run after run 1's drop is judged from the newest drop
// commit on, so only a fix run's own code asks for a newer row. Three commits
// are excepted: run 1's reshaped `Flow-Key: flow.commit-two` commit, which
// reshapes verified code, the pipeline's `Flow-Key: flow.visual-verify`
// record commit, which carries a capture's records and screenshots, never
// code, and a commit whose `Flow-Key:` is the key of a `verifier` dispatch
// row keyed `visual-verify…` — a visual verifier's capture commit
// (skills/flow-contracts/git-boundaries.md, **Commits outside every task**;
// kan-1102 self-review, F5). A `visual-verify-fix-…` key belongs to an
// implementer or panel-fix row, so a fix round's code is still asked for a
// fresh verify row. Only the worktree holding the plan reads it.
//
// All signals are counted independently and reported together on the one
// line, so the operator sees the whole picture in one prompt rather than being
// sent back around the loop one signal at a time.
//
// THE STORE CALLS ANCHOR AT THE PLAN'S PROJECT (KAN-260). The change's
// records — findings and verdicts alike — live under the project of the repo
// that ran /flow, which is the repo where the plan lives, and the project key
// derives from -C's git common dir. So the findings query and both verdict
// calls pass -C anchored at STORE_ANCHOR: the resolved plan's directory when
// the plan is not local to the judged worktree, the worktree itself when it
// is or when nothing resolved. A satellite run anchored at its own worktree
// would read the SECOND repo's project — an empty findings array read as
// CLEAR on this very verdict line — and record the verdict where the
// canonical project's tools never look. -worktree keeps naming the worktree
// the guard was asked to judge, never the anchor.
//
// WHY THE CHECKLIST PATTERN IS ANCHORED TO COLUMN 0. `- [ ]` is matched only at
// the very start of a line — no leading whitespace at all. A task line sits at
// column 0 by this repository's task/step grammar (`build-green.md`); a step's
// checkbox is indented two columns beneath its task, and `implement.md` is
// explicit that "a step's checkbox tracks the step and gates nothing" — so a
// ticked or unticked step must never move this guard's count. Column-0
// anchoring also keeps the original reason for anchoring at all: unanchored,
// `- [ ]` also matches a plan that merely QUOTES a checklist inside a fenced
// example, which this repository's own plans do. A guard that fires on every
// plan that documents a checkbox is a guard the operator learns to click past.
// Fence tracking was rejected as the fuller fix: it is real complexity, and
// what it would still catch — a fenced example whose line BEGINS with `- [ ]`
// at column 0 — errs toward OUTSTANDING, which prompts the operator rather
// than clearing the gate silently.
//
// A SATELLITE'S MISSING LOCAL PLAN IS NOT THE MISSING-RECORD CASE ABOVE
// REJECTS. "A missing file counts as OUTSTANDING, not CLEAR" is about a
// change whose plan was simply never written — the absence itself is the
// only evidence there is, and treating it as clearance is the exact failure
// this guard exists to prevent. A satellite (KAN-363) carries no local
// `tasks.md` by design: `link.md`'s `## Part of` names a peer tree and a
// canonical change id, and the plan the operator actually needs to check is
// there, not here. Reading that absence the same way as a genuinely missing
// plan would report OUTSTANDING with nothing an operator can act on — the
// plan is not missing, it is elsewhere, and this guard's job is to reach it,
// per stats/internal/guard/plans/changeplan.go (task 7) and design.md's
// `guards-take-the-canonical-worktree-path`. So a satellite takes one of two
// outcomes and never the third: its canonical plan resolves and is counted
// exactly like a local one would be, or it does not resolve at all and the
// guard refuses outright (exit 2, "cannot determine anything") rather than
// reporting OUTSTANDING over a plan it never actually read. A change
// directory that carries a `link.md` with no `## Part of` — the canonical
// side's own copy, which names its `## Parts` instead — is not a satellite by
// this definition and falls through to the ordinary missing-plan case: this
// guard is never asked to judge a canonical change directory as if it were
// one of its own parts.
//
// HOW TO HAND-VERIFY AN OUTSTANDING VERDICT (KAN-446). A verdict this guard
// prints can be a structural false positive — the KAN-423 shape: a cross-repo
// change whose plan resolves only in the canonical worktree, so the worktree
// this verdict names holds neither signal. Verify from the primary records
// before acting on the verdict: resolve the plan the way this guard does
// (stats/internal/guard/plans/changeplan.go) and count column-0 `^- \[ \]` lines by hand;
// run `flow record findings -change <name> -C <worktree>` and expect `[]`.
// Both clean means the verdict was structural — record it (unfinished-work-gate.md's
// false-positive course) rather than trusting it or silently overriding it.
// Anything else means the verdict was right, and the courses it offers stand.

// checkUnfinishedWork is `flow-guard check-unfinished-work`: the comment above is the contract -- one CLEAR/OUTSTANDING verdict line on
// stdout at exit 0, exit 2 when it cannot answer at all. Plan resolution is
// PlanDir (changeplan.go), never a second copy; the findings read is
// Env.Findings when set, else `flow record findings`; the verdict write and
// the prior-false-positive read are always the `flow` CLI on Env's PATH. The
// reasoning for each branch, moved here from the bash body it replaced
// (3e48ecac), sits beside the code it explains.
func init() {
	Registry["check-unfinished-work"] = checkUnfinishedWork
}

const uwPrefix = "check-unfinished-work: "

// unsplitListHint names the cause of an argument that cannot hold
// whitespace and does (kan-970 self-review): a list built in a variable and
// passed unquoted reaches a guard as one argument under zsh.
const unsplitListHint = "looks like a list passed as one argument — zsh does not word-split an unquoted $VAR; pass each value as its own argument"

func checkUnfinishedWork(args []string, env core.Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	worktree, name, canonical := arg(0), arg(1), arg(2)
	if worktree == "" || name == "" {
		fmt.Fprintln(stderr, "usage: flow-guard check-unfinished-work <worktree> <change-name> [canonical-worktree] [<worktree>=<merge-base>]...")
		core.Reason(env, "usage")
		return 2
	}
	// The change's commit ranges, one per worktree of its resolved set:
	// signal seven walks them for the newest code commit.
	var ranges [][2]string
	for _, a := range args[min(3, len(args)):] {
		wt, base, ok := strings.Cut(a, "=")
		if !ok || wt == "" || base == "" {
			fmt.Fprintf(stderr, "%s'%s' is not <worktree>=<merge-base>\n", uwPrefix, a)
			core.Reason(env, "bad-range-arg")
			return 2
		}
		if strings.ContainsAny(base, " \t\n") {
			fmt.Fprintf(stderr, "%s'%s' is not <worktree>=<merge-base> — %s\n", uwPrefix, a, unsplitListHint)
			core.Reason(env, "bad-range-arg")
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
	// PlainChangeName (changename.go) itself now, where the bash guard
	// kept its own `case` copy: TestCheckUnfinishedWork's 8e cases still
	// assert the same rejected shapes.
	//
	// The bash guard exported LC_ALL=C because `case`'s ranges are
	// collating ranges -- `démo` was rejected under C and accepted under
	// en_US.UTF-8 on bash 3.2. PlainChangeName compares bytes, so no locale
	// reaches it; the 8e-i cases pin that under both locales.
	//
	// It also closes the symlink question at these paths: IsFile and IsDir
	// follow symlinks, so a name that cannot leave the change's own directory
	// is what makes the paths this guard reads the ones it was asked about.
	if !core.PlainChangeName(name) {
		fmt.Fprintf(stderr, "%schange name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", uwPrefix, name)
		core.Reason(env, "change-name-not-plain")
		return 2
	}
	if !core.IsDir(worktree) {
		fmt.Fprintf(stderr, "%s%s is not a directory — cannot determine anything\n", uwPrefix, worktree)
		core.Reason(env, "not-a-directory")
		return 2
	}

	changes := worktree + "/" + core.SpecRootLeaf(worktree, stderr) + "/changes"
	var reasons []string

	// unreadable is a file that exists but cannot be read: neither signal but
	// an honest unknown, so this guard refuses rather than guessing. Reading
	// it as zero unticked boxes would fail toward the reassuring answer, the
	// one failure direction this whole guard exists to remove.
	unreadable := func(path string) int {
		fmt.Fprintf(stderr, "%scannot read %s — cannot determine anything\n", uwPrefix, path)
		core.Reason(env, "cannot-read")
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
	// THE PRIMARY PLAN IS RESOLVED THROUGH PlanDir, NOT COMPOSED
	// DIRECTLY, so a satellite's own `link.md` is followed to its canonical
	// `tasks.md` instead of reporting the local path's absence as a verdict
	// (the header's satellite paragraph). The directory, not only the path,
	// because THE FIX-SUB-CHANGE SWEEP BELOW MUST WALK THE SAME TREE THE PLAN
	// RESOLVED FROM: a satellite's `<canonical-id>-fix-N` sibling lives
	// beside the canonical change, and sweeping the satellite's own,
	// plan-less changes/ would report CLEAR over an unchecked fix-round item.
	//
	// PlanDir answers 1 in two shapes told apart below: a plain change
	// with no plan and no `## Part of` link -- the ordinary missing-plan case
	// -- versus a genuine satellite whose canonical plan could not be
	// reached, which refuses outright. It answers 3 (KAN-267) when the state
	// record resolved the name to more than one project's plan: its lines
	// naming each match are printed and this guard refuses, since an
	// OUTSTANDING built on a guess between two projects' plans is exactly
	// the silent clearance this guard exists to prevent. Its stderr is shown
	// only then, as the bash guard showed its captured stderr.
	var planErr bytes.Buffer
	planDir, planRC := plans.PlanDir(env, &planErr, worktree, name, canonical)
	if planRC == 3 {
		_, _ = stderr.Write(planErr.Bytes())
		fmt.Fprintf(stderr, "%sthe state record resolves '%s' to more than one project's plan — cannot determine anything\n", uwPrefix, name)
		core.Reason(env, "plan-ambiguous")
		return 2
	}
	planOpen, parentOpen := 0, 0
	var primary, sweepDir, sweepID string
	if planRC == 0 {
		primary = planDir + "/tasks.md"
		n, ok := uwCountUnticked(primary)
		if !ok {
			return unreadable(primary)
		}
		planOpen += n
		design := planDir + "/design.md"
		if parentOpen, ok = uwCountParentActions(design); !ok {
			return unreadable(design)
		}
		sweepDir, sweepID = core.Dirname(planDir), uwBasename(planDir)
	} else {
		link := changes + "/" + name + "/link.md"
		if uwHasLine(link, "## Part of") {
			if canonical != "" {
				fmt.Fprintf(stderr, "%s'%s' is a satellite change (link.md at %s); its canonical plan was not found under the supplied canonical worktree %s, and a canonical worktree that was supplied is never retried against peers — cannot determine anything\n", uwPrefix, name, link, canonical)
			} else {
				peers := worktree + "/" + core.SpecRootLeaf(worktree, stderr) + "/peers"
				fmt.Fprintf(stderr, "%s'%s' is a satellite change (link.md at %s); no canonical worktree was supplied and its canonical plan could not be resolved through %s — cannot determine anything\n", uwPrefix, name, link, peers)
			}
			core.Reason(env, "canonical-plan-unresolved")
			return 2
		}
		primary = changes + "/" + name + "/tasks.md"
		core.Reason(env, "no-plan")
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
		core.Reason(env, "unchecked-plan-items")
		reasons = append(reasons, strconv.Itoa(planOpen)+" unchecked plan item(s)")
	}
	if parentOpen > 0 {
		core.Reason(env, "unticked-parent-actions")
		reasons = append(reasons, strconv.Itoa(parentOpen)+" unticked parent action(s) in "+planDir+"/design.md")
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
			core.Reason(env, "cannot-read-findings")
			return 2
		}
	} else {
		out, errOut, rc := core.Flow(env, false, "record", "findings", "-change", name, "-C", anchor)
		if rc != 0 {
			fmt.Fprintf(stderr, "%scannot read findings for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(errOut), "\n"))
			core.Reason(env, "cannot-read-findings")
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
	// open: nothing is deferred (KAN-862). A Minor is never counted.
	open, ok := uwOpenFindings(findingsJSON)
	if !ok {
		fmt.Fprintln(stderr, uwPrefix+"jq failed — cannot determine anything")
		core.Reason(env, "jq-failed")
		return 2
	}
	if open > 0 {
		core.Reason(env, "open-findings")
		reasons = append(reasons, strconv.Itoa(open)+" open Critical or Important finding(s) in the review panel record")
	}
	// Signal seven reads the change's dispatch rows, not a worktree's, and
	// the gate runs once per worktree: only the worktree holding the plan
	// (anchor == worktree) reads them. A failed or unreadable read is exit 2,
	// like signal two's.
	if anchor == worktree && len(ranges) > 0 {
		dispOut, dispErr, dispRC := core.Flow(env, false, "record", "dispatches", "-change", name, "-C", anchor)
		if dispRC != 0 {
			fmt.Fprintf(stderr, "%scannot read dispatches for '%s' from the store — cannot determine anything: %s\n", uwPrefix, name, strings.TrimRight(string(dispErr), "\n"))
			core.Reason(env, "cannot-read-dispatches")
			return 2
		}
		var dispatches []struct {
			Key       string    `json:"key"`
			Role      string    `json:"role"`
			Outcome   string    `json:"outcome"`
			StartedAt time.Time `json:"startedAt"`
		}
		if json.Unmarshal(dispOut, &dispatches) != nil || dispatches == nil {
			fmt.Fprintf(stderr, "%sdispatches for '%s' are not one JSON array — cannot determine anything\n", uwPrefix, name)
			core.Reason(env, "dispatches-unparseable")
			return 2
		}
		// SIGNAL SEVEN — a code commit no completed inline verify row covers
		// (decision uw-verify-row-after-code): the newest commit outside
		// `<spec root>/changes/` in any range, authored after the newest
		// completed `verify`/`verify-…` verifier row, or with none at all —
		// run 1's reshaped `Flow-Key: flow.commit-two` commit and the
		// pipeline's `Flow-Key: flow.visual-verify` record commit excepted, and
		// so is a visual verifier's capture commit, its Flow-Key a `verifier`
		// row's `visual-verify…` key (kan-1102 self-review, F5). The
		// range runs from the merge base. A `skipped` row counts as completed: it only ends a test
		// group whose every command was unchanged since a green clean run
		// (decision f1-skipped-is-verified, kan-1037).
		{
			var lastVerify int64 = -1
			for _, d := range dispatches {
				if records.Class(d.Role) == "verifier" && (d.Outcome == "completed" || d.Outcome == "skipped") && (d.Key == "verify" || strings.HasPrefix(d.Key, "verify-")) {
					lastVerify = max(lastVerify, d.StartedAt.Unix())
				}
			}
			verifierKeys := map[string]bool{}
			for _, d := range dispatches {
				if records.Class(d.Role) == "verifier" && strings.HasPrefix(d.Key, "visual-verify") {
					verifierKeys[d.Key] = true
				}
			}
			newest, newestAt := "", int64(-1)
			for _, r := range ranges {
				leaf := core.SpecRootLeaf(r[0], io.Discard)
				out, ok := core.Capture(core.EnvGit(env)("-C", r[0], "log", "--no-merges", "--format=%h %at %(trailers:key=Flow-Key,valueonly,separator=%x20)", r[1]+"..HEAD", "--", ".", ":(exclude)"+leaf+"/changes/"))
				if !ok {
					fmt.Fprintf(stderr, "%scannot read the commits of %s..HEAD in %s — cannot determine anything\n", uwPrefix, r[1], r[0])
					core.Reason(env, "cannot-read-the-commits")
					return 2
				}
				for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
					sha, rest, _ := strings.Cut(l, " ")
					at, flowKey, _ := strings.Cut(rest, " ")
					// Run 1's reshaped implementation commit carries a fresh author
					// date; it reshapes verified code and is not new code. The
					// pipeline's visual-verify record commit carries a capture's
					// records and screenshots, never code.
					if strings.Contains(flowKey, "flow.commit-two") || strings.Contains(flowKey, "flow.visual-verify") || verifierKeys[strings.TrimSpace(flowKey)] {
						continue
					}
					if n, err := strconv.ParseInt(at, 10, 64); err == nil && n > newestAt {
						newest, newestAt = sha, n
					}
				}
			}
			if newest != "" && newestAt > lastVerify {
				core.Reason(env, "no-inline-verify")
				reasons = append(reasons, "no completed inline verify row since "+newest+" — run Inline verify (skills/flow/verify-and-handoff.md) and record its row before landing — a re-run after a new commit takes a fresh `verify-<k>` key (the store keeps a reused key's first start)")
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
	core.Flow(env, false, "record", "verdict", "-change", name, "-guard", "check-unfinished-work", "-worktree", worktree, "-verdict", verdict, "-C", anchor)
	if len(reasons) > 0 {
		if prior, _, rc := core.Flow(env, false, "record", "verdicts", "-guard", "check-unfinished-work", "-false-positive", "-C", anchor); rc == 0 {
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
func uwDerivedRange(env core.Env, worktree string) ([2]string, bool) {
	var out bytes.Buffer
	// A failure here only means no derived range, never a refusal of this
	// run: the resolver's reason is not this run's.
	quiet := env
	quiet.Reasons = nil
	if gitops.ResolveBaseBranch([]string{worktree}, quiet, &out, io.Discard) != 0 {
		return [2]string{}, false
	}
	base, ok := core.Capture(core.EnvGit(env)("-C", worktree, "merge-base", "HEAD", "origin/"+strings.TrimSpace(out.String())))
	if !ok || strings.TrimSpace(base) == "" {
		return [2]string{}, false
	}
	return [2]string{worktree, strings.TrimSpace(base)}, true
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

// uwCountParentActions counts design.md's column-0 `- [ ] Parent:` lines: the
// actions a decision gives the orchestrating session outside every task,
// each held open until ticked. A missing design.md
// counts none — a plan predating the line has no such action to hold — and
// false is an existing file that cannot be read.
func uwCountParentActions(path string) (int, bool) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return 0, true
	}
	if err != nil {
		return 0, false
	}
	return bytes.Count(append([]byte("\n"), b...), []byte("\n- [ ] Parent:")), true
}

// uwHasLine is `grep -qxF <line> <file>`: false for an unreadable file.
func uwHasLine(path, line string) bool {
	b, err := os.ReadFile(path)
	if err != nil || !core.IsFile(path) {
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
// Dirname.
func uwBasename(p string) string { return p[strings.LastIndexByte(p, '/')+1:] }

// uwOpenFindings is the count of open Critical and Important findings over one
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
		if f == nil || !core.IsJSONString(f["status"]) || json.Unmarshal(f["status"], &status) != nil {
			return 0, false
		}
		// Closed is `fixed` or a reasoned `withdrawn` (KAN-791): the bare
		// word -- no reason after it -- is open, the same line the store's
		// withdrawnWithoutReason and check-panel-findings-closed's own copy
		// draw. A `deferred <reason>` row is open: nothing is deferred
		// (KAN-862). Only a Critical or Important finding counts.
		var severity string
		_ = json.Unmarshal(f["severity"], &severity)
		if sev := strings.ToLower(severity); sev != "critical" && sev != "important" {
			continue
		}
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
	flagged := core.JQRaw(first["flaggedAt"])
	if core.IsJSONString(first["flaggedAt"]) {
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
	last := core.JQRaw(first["falsePositiveReason"]) + " (" + core.JQRaw(first["change"]) + ", " + flagged + ")"
	return len(all), strings.TrimRight(last, "\n"), true
}
