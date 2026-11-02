package guard

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
)

// The contract below moved here from the header of the deleted
// scripts/check-dispatch-paragraphs.sh shim (KAN-1058).
//
// flow-guard check-dispatch-paragraphs — the guard that keeps a required dispatch
// paragraph from silently disappearing.
//
// Why this exists: KAN-289's root cause was that the REPRODUCE, DON'T READ
// instruction lived in no template and depended on the dispatcher
// remembering to type it. Part 1 of that change fixed it by adding the
// paragraph, verbatim, at dispatch sites. Nothing then stops a LATER prose
// edit from trimming a required paragraph away one line at a time — this
// guard is what makes that loud instead of silent, the same role
// check-guard-symlinks.sh plays for its own templates. KAN-217 added a second required paragraph — the VERBATIM
// REPORT — THE FACT blockquote that tells the fix subagent its slot's
// report file outranks the dispatcher's summary — and generalized this
// guard from one hard-coded paragraph to a table of them so a third would
// cost a table row, not a second copy of the machinery. KAN-263 added a
// third required paragraph — FOREGROUND BUILDS, which forbids a dispatched
// agent from ending its turn with a build/test/long-running command still
// running in the background — at all four dispatch sites that can run one:
// the implementer dispatch, the conductor's own §4 instruction, the review
// panel's slot dispatch, and the panel-fix subagent dispatch. KAN-441 added a
// fourth required paragraph — TARGETED TESTS, which tells an implementer or
// panel-fix subagent to run only the task's own tests through the build
// tool's own selector rather than a whole module or repository suite mid-task
// — at the implementer dispatch and the panel-fix subagent dispatch. KAN-464
// added a fifth required paragraph — MUTATION PROOF, which hands the fix
// round's mutation-proof obligation to the panel-fix subagent itself rather
// than the conductor — at the panel-fix subagent dispatch alone. KAN-473
// added a sixth required paragraph — TOOLS, which tells every dispatched
// role to load a deferred tool by name in its first turn rather than a
// later wildcard ToolSearch, since a schema loaded mid-session changes the
// tools array ahead of every message and re-prices the whole context — at
// six sites across two new site files: the conductor dispatch and the
// implementer dispatch in implement.md, the panel slot dispatch and the
// panel-fix subagent dispatch in review-panel.md, the planner dispatch in
// brainstorm.md, and the verifier dispatch in verify-and-handoff.md.
// KAN-257 extended the MUTATION PROOF entry from three shared phrases to
// six, carrying the fix round's assert obligation: a mutation must confirm
// its edit landed before the tests run, and an edit that never applied is
// a refusal to redo — never a surviving mutant. KAN-472 task 10 added a
// seventh required paragraph — MODEL HANDSHAKE, the universal `Model:
// <model>` first-reply obligation stated once in implement.md's **Dispatch
// the conductor** section and cited from every other dispatch site — at the
// same six sites the TOOLS paragraph occupies: implement.md twice (conductor
// dispatch, implementer dispatch), review-panel.md twice (panel slot
// dispatch, panel-fix subagent dispatch), brainstorm.md once (planner
// dispatch), verify-and-handoff.md once (verifier dispatch). KAN-472
// task 20 added an eighth required paragraph — INDEPENDENT PASSES, which
// forbids one pass of a bundled review dispatch from citing, deferring to, or
// skipping a defect an earlier pass in the same bundle already raised — at
// one site, review-panel.md's bundle prompt, min 1 block. KAN-484 added a
// ninth required paragraph — NO DELEGATION, which tells every leaf agent the
// conductor dispatches to do the work itself and never call the Agent tool
// or spawn a subagent, closing the conductor's closed list one level down —
// at four sites: the implementer dispatch in implement.md, the panel slot
// dispatch and the panel-fix subagent dispatch in review-panel.md, and the
// verifier dispatch in verify-and-handoff.md. KAN-400 re-introduced the
// per-task reviewer as a gated dispatch — one reviewer per task commit
// whose review gate in skills/flow/implement.md fires — as implement.md's
// fifth dispatch site, raising that file's minimums by one for REPRODUCE,
// DON'T READ (the reviewer variant now required twice there), FOREGROUND
// BUILDS, TOOLS, MODEL HANDSHAKE and NO DELEGATION. This file pins only the
// paragraphs the gated reviewer dispatch carries, never the gate's own
// numbers. KAN-496
// added a tenth required paragraph — PIXEL PROBE, which requires any fix
// the panel-fix round lands to draw/geometry code to carry a probe
// assertion against the actual rendered pixels, and rejects a
// fix commit that lacks one at the fix step rather than discovering the
// regression next round — at the panel-fix subagent dispatch in
// review-panel.md alone, the same single site MUTATION PROOF occupies.
// KAN-328 added an eleventh required paragraph — PROVE THE GUARD BITES,
// which tells an implementer whose tests assert on configuration or file
// content that a passing run is not evidence: break the property the
// assertion protects, run the guard or the test against the broken state and capture
// its failure, restore, capture the pass — at the implementer dispatch in
// implement.md alone, beside the TDD and systematic-debugging REQUIRED
// SUB-SKILL paragraphs. KAN-405 added a twelfth required paragraph —
// REPORT, DON'T DECIDE, which tells an implementer that a question only the
// operator can settle is reported, never decided in code: carried to the
// REPORT FILE where the plan can proceed, a BLOCKED report where it cannot,
// because a decision the implementer makes silently is one nobody recorded —
// at the implementer dispatch in implement.md alone, the same single site
// PROVE THE GUARD BITES occupies. KAN-521 added a thirteenth required
// paragraph — ENTRY CONTEXT, which gives every panel slot dispatch a named
// entry context (the diff and the inlined touched-files list, the contracts
// among them named) so a reviewer's exploration starts scoped instead of from
// the whole tree, and has the reviewer state the coverage trade explicitly in
// its report — at the panel slot dispatch in review-panel.md alone, carried
// into every bundled dispatch by the shared-paragraphs list. KAN-564 added a
// fourteenth required paragraph — FINDINGS ARE INPUT, which tells the
// panel-fix subagent that a finding names a defect and suggests a route to
// fixing it — findings are input, not orders — that the subagent's obligation
// is to resolve the defect the finding names, and that a route deviating from
// the finding's literal suggestion is taken only when the report records the
// deviation and justifies it, so a justified deviation like kan-469's F2
// (shared helpers over the literal route that would have reordered a required
// clause ordering) stays legal while a silent one does not — at the
// panel-fix subagent dispatch in review-panel.md alone, the same single site
// MUTATION PROOF and PIXEL PROBE occupy. KAN-593 added a fifteenth required
// paragraph — CONTEXT BUNDLE FAILURE, which makes the review panel's
// pre-flight bundle rebuild fail loud instead of silent: a non-zero gather
// exit or an absent bundle file records the cause in the pass log and then
// stops the stage or takes an explicit operator override before any slot is
// dispatched, so the reviewers' context is never silently reduced — at the
// panel pre-flight in review-panel.md alone. KAN-635 added a sixteenth
// required paragraph — READ-ONLY REVIEW, which forbids the gated per-task
// reviewer from mutating the worktree it reviews: no writing git command, no
// commit, no index change, no edit outside its own report file, because a
// reviewer that mutates the tree (gymie KAN-635's `git checkout <sha> -- .`
// destroyed uncommitted planning artifacts) and self-reports a restore is a
// data-loss shape nobody checked — at the gated per-task reviewer dispatch
// in implement.md alone; the panel slots carry the same rule inside their
// prompt files' own Read-Only Review sections, which this table does not
// pin. KAN-1062 moved the implementer dispatch's paragraphs into
// skills/flow/implementer-brief.md and the gated per-task reviewer's into
// skills/flow/gated-reviewer-brief.md, each prompt now a stub naming its
// brief (decision briefs-on-disk): every implementer and gated reviewer
// site below names the brief that carries the paragraph, and implement.md
// keeps the parent's own REPRODUCE, DON'T READ (reviewer variant) and
// FOREGROUND BUILDS restatements and the MODEL HANDSHAKE line of each of its
// two stubs. Every "implementer dispatch" below is implementer-brief.md and
// every "gated per-task reviewer dispatch" gated-reviewer-brief.md.
//
// Argument-free and self-scoped, exactly like check-guard-symlinks.sh: the
// scan root is this repository's own root, resolved from this script's own
// location, so no call site can narrow it. CHECK_DISPATCH_PARAGRAPHS_ROOT is
// an explicit, opt-in override honored only when set (mirrors
// CHECK_GUARD_SYMLINKS_ROOT), so the Go tests
// (stats/internal/guard/check_dispatch_paragraphs_test.go) can point this
// guard at a sandboxed fixture tree under TMPDIR without touching this
// repository — never set it for a normal invocation.
//
// THE PARAGRAPH TABLE. Each entry is a label, its shared phrases (required
// of every block carrying that label, held as short literals rather than a
// copy of the whole blockquote, per design.md's `label-plus-phrases`), and
// its variants — each variant a name plus its own load-bearing phrase, or no
// variants at all when every block of that label is equivalent. Every entry
// lists the sites that require it, each site naming its minimum block count
// and which variants it requires.
//
//   The verifier dispatch moved from verify-and-handoff.md to visual-verify.md
//   when the visual-verification procedure split into its own on-demand file,
//   so its three sites below name the new file; the tooling-analyst dispatch
//   moved on to visual-verify-tooling-analysis.md (KAN-858), which loads only
//   on a fix run with a miss, so each file carries one block of each.
//
//   Paragraph                          Site                                       Min  Variants
//   **REPRODUCE, DON'T READ:**         skills/flow/review-panel.md                 1   reviewer
//   **REPRODUCE, DON'T READ:**         skills/flow/implement.md                    1   reviewer
//   **REPRODUCE, DON'T READ:**         skills/flow/implementer-brief.md            1   implementer
//   **REPRODUCE, DON'T READ:**         skills/flow/gated-reviewer-brief.md         1   reviewer
//   **VERBATIM REPORT — THE FACT:**    skills/flow/review-panel.md                 1   (none)
//   **REPRODUCER:**                    skills/flow/review-panel.md                 1   (none)
//   **FOREGROUND BUILDS:**             skills/flow/implement.md                    1   (none)
//   **FOREGROUND BUILDS:**             skills/flow/implementer-brief.md            1   (none)
//   **FOREGROUND BUILDS:**             skills/flow/gated-reviewer-brief.md         1   (none)
//   **FOREGROUND BUILDS:**             skills/flow/review-panel.md                 2   (none)
//   **TARGETED TESTS:**                skills/flow/implementer-brief.md            1   (none)
//   **TARGETED TESTS:**                skills/flow/review-panel.md                 1   (none)
//   **MUTATION PROOF:**                skills/flow/review-panel.md                 1   (none)
//   **PIXEL PROBE:**                   skills/flow/review-panel.md                 1   (none)
//   **TOOLS:**                         skills/flow/implementer-brief.md            1   (none)
//   **TOOLS:**                         skills/flow/gated-reviewer-brief.md         1   (none)
//   **TOOLS:**                         skills/flow/review-panel.md                 2   (none)
//   **TOOLS:**                         skills/flow/visual-verify.md                1   (none)
//   **TOOLS:**                         skills/flow/visual-verify-tooling-analysis.md 1 (none)
//   **MODEL HANDSHAKE:**               skills/flow/implement.md                    2   (none)
//   **MODEL HANDSHAKE:**               skills/flow/review-panel.md                 2   (none)
//   **MODEL HANDSHAKE:**               skills/flow/visual-verify.md                1   (none)
//   **MODEL HANDSHAKE:**               skills/flow/visual-verify-tooling-analysis.md 1 (none)
//   **INDEPENDENT PASSES:**            skills/flow/review-panel.md                 1   (none)
//   **NO DELEGATION:**                 skills/flow/implementer-brief.md            1   (none)
//   **NO DELEGATION:**                 skills/flow/gated-reviewer-brief.md         1   (none)
//   **NO DELEGATION:**                 skills/flow/review-panel.md                 2   (none)
//   **NO DELEGATION:**                 skills/flow/visual-verify.md                1   (none)
//   **NO DELEGATION:**                 skills/flow/visual-verify-tooling-analysis.md 1 (none)
//   **PROVE THE GUARD BITES:**         skills/flow/implementer-brief.md            1   (none)
//   **REPORT, DON'T DECIDE:**          skills/flow/implementer-brief.md            1   (none)
//   **ENTRY CONTEXT:**                 skills/flow/review-panel.md                 1   (none)
//   **FINDINGS ARE INPUT:**            skills/flow/review-panel.md                 1   (none)
//   **CONTEXT BUNDLE FAILURE:**        skills/flow/review-panel.md                 1   (none)
//   **OUTPUT BUDGET:**                 skills/flow/implementer-brief.md            1   (none)
//   **OUTPUT BUDGET:**                 skills/flow/review-panel.md                 1   (none)
//   **READ-ONLY REVIEW:**              skills/flow/gated-reviewer-brief.md         1   (none)
//   **FIX-ROUND SCOPE:**               skills/flow/review-panel-fix-round.md       1   (none)
//   **PROOF RUNS:**                    skills/flow/implementer-brief.md            1   (none)
//   "is allowed only under flow-guard run-capped" rules/fix-determinism-at-the-source.mdc 1 (none)
//   "is allowed only under flow-guard run-capped" skills/flow/verify-fix-loop.md       1   (none)
//   **REQUIRED SUB-SKILL:**            skills/flow/implementer-brief.md            2   tdd AND debugging
//   **PLANNING DELEGATION:**           skills/flow-self-review/SKILL.md            1   (none)
//
//   REPRODUCE, DON'T READ shared phrases: "crosses a boundary", "the store,
//   the filesystem, a guard, a real transcript", "exercise the real thing"
//     - reviewer variant's own phrase: "worth less than one you did"
//     - implementer variant's own phrase: "a fake or a hand-built value"
//
//   VERBATIM REPORT — THE FACT shared phrases (no variants — every block
//   carrying the label must carry all three): "the reviewer's own report",
//   "never a source of fact", "the report wins"
//
//   FOREGROUND BUILDS shared phrases (no variants — every block carrying the
//   label must carry all three): "still executing in the background", "Run
//   it in the foreground", "poll it to completion". Required once in each of
//   implementer-brief.md (implementer dispatch), gated-reviewer-brief.md
//   (gated per-task reviewer dispatch) and implement.md (the parent's own §4
//   restatement), and twice in review-panel.md (panel slot dispatch,
//   panel-fix subagent dispatch).
//
//   OUTPUT BUDGET shared phrases (no variants — every block carrying the
//   label must carry all four): "re-reads your", "by line range", "never
//   re-read a file already in your context", "Never print a generated
//   file". Required once in each of implementer-brief.md (implementer
//   dispatch) and review-panel.md (panel-fix subagent dispatch) — the two dispatches whose
//   tool output grows with the task: every turn re-reads the whole context,
//   so an uncapped read or log is paid for on every turn after it.
//
//   TARGETED TESTS shared phrases (no variants — every block carrying the
//   label must carry all three): "the build tool's own selector", "once for
//   RED, once for GREEN", "module or repository suite mid-task". Required
//   once in each of implementer-brief.md (implementer dispatch) and
//   review-panel.md (panel-fix subagent dispatch) — the panel slot dispatch
//   is not a site: reviewers do not run the task's tests.
//
//   MUTATION PROOF shared phrases (no variants — every block carrying the
//   label must carry the full set of seven): "mutation-proved before you end your
//   turn", "confirm an existing test fails, and restore", "a surviving
//   mutant", "confirm the edit landed", "a refusal, not a surviving
//   mutant", "never buys a test", and the line shape itself,
//   "`fix-mutation: <path> — <what was mutated> — <the test that failed>`" —
//   the fix subagent never receives the contract's fenced block, so its
//   dispatch states the shape (KAN-854). Required once, at the panel-fix subagent
//   dispatch in review-panel.md alone — the implementer dispatch and the
//   panel slot dispatch are not sites: the fix round is the only one this
//   obligation binds.
//
//   PIXEL PROBE shared phrases (no variants — every block carrying the
//   label must carry all three): "draw/geometry code", "actual rendered
//   pixels or geometry", "rejected at the fix step". Required once, at the
//   panel-fix subagent dispatch in review-panel.md alone — the implementer
//   dispatch and the panel slot dispatch are not sites: the fix round is
//   the only dispatch that lands a fix commit.
//
//   PROVE THE GUARD BITES shared phrases (no variants — every block
//   carrying the label must carry all three): "assert on configuration or
//   file content", "break the property the assertion protects", "against
//   the broken state". Required once, at the implementer dispatch in
//   implementer-brief.md alone — the gated per-task reviewer dispatch is not a
//   site: reviewers judge a diff, they do not run the task's guards.
//
//   REPORT, DON'T DECIDE shared phrases (no variants — every block
//   carrying the label must carry all four): "only the operator can
//   settle", "reported, never decided in code", "carry the question in your
//   REPORT FILE", "report BLOCKED". Required once, at the implementer
//   dispatch in implementer-brief.md alone — the gated per-task reviewer dispatch
//   is not a site: reviewers judge a diff, they do not settle operator
//   questions.
//
//   CONTEXT BUNDLE FAILURE shared phrases (no variants — every block
//   carrying the label must carry all five): "exited non-zero, or the
//   bundle file is still", "context bundle: build failed", "dispatch every
//   slot without the bundle", "context bundle: dispatched without it —
//   operator override", "outcome stopped". Required once, at the panel
//   pre-flight in review-panel.md alone — the fix subagent's rebuild takes
//   the same path through its existing "same as above", so it needs no
//   second block.
//
//   READ-ONLY REVIEW shared phrases (no variants — every block carrying the
//   label must carry all three): "never mutate it", "no file edit outside
//   your own report file", "a claim nobody can check". Required once, at the
//   gated per-task reviewer dispatch in gated-reviewer-brief.md alone — the panel slot
//   dispatch is not a site: its slots carry the rule inside their prompt
//   files' own Read-Only Review sections, not a shared paragraph this table
//   pins.
//
//   TOOLS shared phrases (no variants — every block carrying the label
//   must carry all three): "in your first turn", "never a wildcard
//   query", "re-prices your whole context". Required once in each of
//   implementer-brief.md (implementer dispatch) and gated-reviewer-brief.md
//   (gated per-task reviewer dispatch — the conductor's own copy, and the
//   planner's in brainstorm.md, were removed with those roles, kan-488), twice in
//   review-panel.md (panel slot dispatch, panel-fix subagent dispatch),
//   and once each in visual-verify.md (verifier dispatch) and
//   visual-verify-tooling-analysis.md (tooling-analyst dispatch).
//
//   MODEL HANDSHAKE shared phrases (no variants — every block carrying the
//   label must carry all three): "the first line of your first reply",
//   "and nothing else on that line", "before any tool call". Required twice
//   in implement.md (the implementer stub and the gated per-task reviewer
//   stub, each on one line — the conductor's own copy, and
//   the planner's in brainstorm.md, were removed with those roles,
//   kan-488), twice in review-panel.md (panel slot dispatch, panel-fix
//   subagent dispatch), and once each in visual-verify.md (verifier
//   dispatch) and visual-verify-tooling-analysis.md (tooling-analyst
//   dispatch) — the TOOLS sites, with the two stubs in place of the two
//   briefs (decision stub-carries-handshake).
//
//   INDEPENDENT PASSES shared phrases (no variants — every block carrying the
//   label must carry all three): "starts from `final-review.diff`", "raise it
//   again under this pass", "before beginning the next pass". Required once,
//   at review-panel.md's bundle prompt alone — the paragraph binds every pass
//   of a bundled dispatch to read only the diff and the code, never an
//   earlier pass's report.
//
//   NO DELEGATION shared phrases (no variants — every block carrying the
//   label must carry all three): "Never call the `Agent` tool", "never spawn
//   a subagent", "the leaf of this run". Required once in each of
//   implementer-brief.md (implementer dispatch) and gated-reviewer-brief.md
//   (gated per-task reviewer dispatch), twice in
//   review-panel.md (panel slot dispatch,
//   panel-fix subagent dispatch), once each in visual-verify.md (verifier
//   dispatch) and visual-verify-tooling-analysis.md (tooling-analyst
//   dispatch) — there is no conductor or planner dispatch left to be a site
//   (kan-488): the parent orchestrates directly and runs brainstorming
//   inline.
//
//   ENTRY CONTEXT shared phrases (no variants — every block carrying the
//   label must carry all three): "your named entry context", "not from a
//   whole-tree exploration", "the coverage trade". Required once, at the
//   panel slot dispatch in review-panel.md — the bundle prompt's
//   shared-paragraphs list carries it into every bundled dispatch, the same
//   way REPRODUCE, DON'T READ is carried.
//
//   FINDINGS ARE INPUT shared phrases (no variants — every block carrying
//   the label must carry all four): "input, not orders", "resolve the defect
//   the finding names", "records the deviation and justifies it", "A silent
//   deviation is an unfixed finding". Required once, at the panel-fix
//   subagent dispatch in review-panel.md alone — the implementer dispatch
//   and the panel slot dispatch are not sites: the fix round is the only
//   dispatch that acts on findings.
//
// A BLOCK is a line carrying a label, plus every immediately-following line
// that continues the same markdown blockquote (a line beginning with `>`) —
// i.e. the whole paragraph. A block counts toward a required variant only
// when it carries the label, every shared phrase, AND that variant's own
// phrase. Two blocks of the same variant do NOT satisfy a two-variant site —
// implementer-brief.md requires one TDD and one systematic-debugging REQUIRED
// SUB-SKILL block, not merely two blocks. An entry with no variants needs no variant match: every
// phrase it lists is simply required of the block.
//
// Every failure is reported as `file:line` naming the missing element — the
// label, the specific missing phrase, or the missing variant. Exit `0`
// clean, `1` a required site missing its block or one of its phrases, `2` it
// cannot answer at all: a scoped file missing, unreadable, a symlink, or a
// read failing for any reason. A scoped path that
// exists as neither a file nor a directory is folded into the same hard `2`
// ("not a regular file") — never a silent skip, which would be a vacuous
// "✓ clean".
//
// WHAT A GREEN RUN DOES NOT PROVE: only that each label and its phrases are
// present at each required site — never that a dispatcher actually wrote the
// report file described by VERBATIM REPORT — THE FACT, never that a fix
// agent read one, and never that a reviewer or an implementer actually
// obeyed REPRODUCE, DON'T READ. A paraphrase that drops one of the listed
// phrases fails loud; a paraphrase that keeps all of them while changing the
// surrounding prose passes clean.
//
// Ported to Go (KAN-841): the paragraph table is dpEntries and dpSites, and
// the body's reasoning for every step is in
// stats/internal/guard/dispatchparagraphs.go. The binary does not live in
// this checkout, so this shim exports FLOW_GUARD_REPO_ROOT — this script's
// parent directory, derived as the bash guard derived its ROOT — as the root
// scanned when CHECK_DISPATCH_PARAGRAPHS_ROOT is unset.

// checkDispatchParagraphs is `flow-guard check-dispatch-paragraphs`: that
// script's header comment is the contract -- every required dispatch
// paragraph is present, with every shared phrase and every required variant,
// at every site the table below names; exit 0 clean, 1 a site missing a
// block or a phrase, 2 cannot answer. The reasoning for each step, moved here
// from the bash body it replaced (d71a2327), sits beside the code it explains.
func init() {
	Registry["check-dispatch-paragraphs"] = checkDispatchParagraphs
}

const dpName = "check-dispatch-paragraphs"

// dpEntry is one paragraph of the table: its label, its shared phrases
// (required of every block carrying the label, held as short literals rather
// than a copy of the whole blockquote), and its variants -- each variant's
// name mapped to its own load-bearing phrase; none means every block of that
// label is equivalent.
type dpEntry struct {
	label    string
	shared   []string
	variants map[string]string
}

// dpEntries is ENTRY_LABEL, ENTRY_SHARED_PHRASES, ENTRY_VARIANTS and
// VARIANT_PHRASE, keyed by the bash table's entry names.
var dpEntries = map[string]dpEntry{
	"reproduce": {"**REPRODUCE, DON'T READ:**",
		[]string{"crosses a boundary", "the store, the filesystem, a guard, a real transcript", "exercise the real thing"},
		map[string]string{"reviewer": "worth less than one you did", "implementer": "a fake or a hand-built value"}},
	"reproducer": {"**REPRODUCER:**",
		[]string{"exits non-zero when the defect is present", "`# demonstrates: <path>:<line>:<content>`", "never a prose description", "|| true", "go test -overlay"}, nil},
	"verbatim": {"**VERBATIM REPORT — THE FACT:**",
		[]string{"the reviewer's own report", "never a source of fact", "the report wins"}, nil},
	"foreground": {"**FOREGROUND BUILDS:**",
		[]string{"still executing in the background", "Run it in the foreground", "poll it to completion"}, nil},
	"targeted": {"**TARGETED TESTS:**",
		[]string{"the build tool's own selector", "once for RED, once for GREEN", "module or repository suite mid-task"}, nil},
	"mutation": {"**MUTATION PROOF:**",
		[]string{"mutation-proved before you end your turn", "confirm an existing test fails, and restore", "a surviving mutant",
			"confirm the edit landed", "a refusal, not a surviving mutant", "never buys a test",
			"`fix-mutation: <path> — <what was mutated> — <the test that failed>`"}, nil},
	"pixel": {"**PIXEL PROBE:**",
		[]string{"draw/geometry code", "actual rendered pixels, in the style", "rejected at the fix step"}, nil},
	"tools": {"**TOOLS:**",
		[]string{"in your first turn", "never a wildcard query", "re-prices your whole context"}, nil},
	"handshake": {"**MODEL HANDSHAKE:**",
		[]string{"the first line of your first reply", "and nothing else on that line", "before any tool call"}, nil},
	"independent": {"**INDEPENDENT PASSES:**",
		[]string{"starts from `final-review.diff`", "raise it again under this pass", "before beginning the next pass"}, nil},
	"delegation": {"**NO DELEGATION:**",
		[]string{"Never call the `Agent` tool", "never spawn a subagent", "the leaf of this run"}, nil},
	"prove": {"**PROVE THE GUARD BITES:**",
		[]string{"assert on configuration or file content", "break the property the assertion protects", "against the broken state"}, nil},
	"decide": {"**REPORT, DON'T DECIDE:**",
		[]string{"only the operator can settle", "reported, never decided in code", "carry the question in your REPORT FILE", "report BLOCKED"}, nil},
	"entry": {"**ENTRY CONTEXT:**",
		[]string{"your named entry context", "not from a whole-tree exploration", "the coverage trade"}, nil},
	"findings": {"**FINDINGS ARE INPUT:**",
		[]string{"input, not orders", "resolve the defect the finding names", "records the deviation and justifies it", "A silent deviation is an unfixed finding"}, nil},
	"contextbundle": {"**CONTEXT BUNDLE FAILURE:**",
		[]string{"exited non-zero, or the bundle file is still", "context bundle: build failed", "dispatch every slot without the bundle",
			"context bundle: dispatched without it — operator override", "outcome stopped"}, nil},
	"budget": {"**OUTPUT BUDGET:**",
		[]string{"re-reads your", "by line range", "never re-read a file already in your context", "Never print a generated file"}, nil},
	"readonly": {"**READ-ONLY REVIEW:**",
		[]string{"never mutate it", "no file edit outside your own report file", "a claim nobody can check"}, nil},
	"fixround": {"**FIX-ROUND SCOPE:**",
		[]string{"this is a fix-round re-review", "and nothing else on the branch", "never the module, repository or live-spec suite", "for a defect the fix itself introduces"}, nil},
	"subskill": {"**REQUIRED SUB-SKILL:**",
		[]string{"from its file", "on Claude Code the newest `~/.claude/plugins/cache/*/superpowers/*/skills/"},
		map[string]string{"tdd": "skills/test-driven-development/SKILL.md`", "debugging": "skills/systematic-debugging/SKILL.md`"}},
	"planning": {"**PLANNING DELEGATION:**",
		[]string{"fresh `Explore` and `Plan` subagents", "never a `fork`", "never an agent that edits files", "flow-guard plan-agent-pair"}, nil},
	"proofruns": {"**PROOF RUNS:**",
		[]string{"flow-guard run-capped", "bounded worker count", "at most 5 capped runs", "straight to code", "shows the check failing with the mechanism removed"}, nil},
	"cappedrepro": {"is allowed only under flow-guard run-capped", []string{"bounded worker count"}, nil},
}

// dpSite is one row of SITE_ENTRY/SITE_PATHS/SITE_MIN_BLOCKS/SITE_VARIANTS:
// the entry (paragraph) it requires, its path relative to the root, its
// minimum block count, and the variants it requires (none: nothing beyond
// the shared phrases).
type dpSite struct {
	entry    string
	path     string
	min      int
	variants []string
}

// dpSites is the site table -- the rows the port inherited, in the bash
// table's order, then the rows added since; a site's violations print in
// table order.
var dpSites = []dpSite{
	{"reproduce", "skills/flow/review-panel.md", 1, []string{"reviewer"}},
	{"reproduce", "skills/flow/implement.md", 1, []string{"reviewer"}},
	{"reproduce", "skills/flow/implementer-brief.md", 1, []string{"implementer"}},
	{"reproduce", "skills/flow/gated-reviewer-brief.md", 1, []string{"reviewer"}},
	{"verbatim", "skills/flow/review-panel.md", 1, nil},
	{"reproducer", "skills/flow/review-panel.md", 1, nil},
	{"foreground", "skills/flow/implement.md", 1, nil},
	{"foreground", "skills/flow/implementer-brief.md", 1, nil},
	{"foreground", "skills/flow/gated-reviewer-brief.md", 1, nil},
	{"foreground", "skills/flow/review-panel.md", 2, nil},
	{"targeted", "skills/flow/implementer-brief.md", 1, nil},
	{"targeted", "skills/flow/review-panel.md", 1, nil},
	{"mutation", "skills/flow/review-panel.md", 1, nil},
	{"pixel", "skills/flow/review-panel.md", 1, nil},
	{"tools", "skills/flow/implementer-brief.md", 1, nil},
	{"tools", "skills/flow/gated-reviewer-brief.md", 1, nil},
	{"tools", "skills/flow/review-panel.md", 2, nil},
	{"tools", "skills/flow/visual-verify.md", 1, nil},
	{"tools", "skills/flow/visual-verify-tooling-analysis.md", 1, nil},
	{"handshake", "skills/flow/implement.md", 2, nil},
	{"handshake", "skills/flow/review-panel.md", 2, nil},
	{"handshake", "skills/flow/visual-verify.md", 1, nil},
	{"handshake", "skills/flow/visual-verify-tooling-analysis.md", 1, nil},
	{"independent", "skills/flow/review-panel.md", 1, nil},
	{"delegation", "skills/flow/implementer-brief.md", 1, nil},
	{"delegation", "skills/flow/gated-reviewer-brief.md", 1, nil},
	{"delegation", "skills/flow/review-panel.md", 2, nil},
	{"delegation", "skills/flow/visual-verify.md", 1, nil},
	{"delegation", "skills/flow/visual-verify-tooling-analysis.md", 1, nil},
	{"prove", "skills/flow/implementer-brief.md", 1, nil},
	{"decide", "skills/flow/implementer-brief.md", 1, nil},
	{"entry", "skills/flow/review-panel.md", 1, nil},
	{"findings", "skills/flow/review-panel.md", 1, nil},
	{"contextbundle", "skills/flow/review-panel.md", 1, nil},
	{"budget", "skills/flow/implementer-brief.md", 1, nil},
	{"budget", "skills/flow/review-panel.md", 1, nil},
	{"readonly", "skills/flow/gated-reviewer-brief.md", 1, nil},
	{"fixround", "skills/flow/review-panel-fix-round.md", 1, nil},
	{"proofruns", "skills/flow/implementer-brief.md", 1, nil},
	{"cappedrepro", "rules/fix-determinism-at-the-source.mdc", 1, nil},
	{"cappedrepro", "skills/flow/verify-fix-loop.md", 1, nil},
	{"subskill", "skills/flow/implementer-brief.md", 2, []string{"tdd", "debugging"}},
	{"planning", "skills/flow-self-review/SKILL.md", 1, nil},
}

func checkDispatchParagraphs(_ []string, env Env, stdout, stderr io.Writer) int {
	// The root is this repository's own root by default -- the parent of
	// scripts/, derived as the bash guard derived it from its own location;
	// the binary lives in a cache, so the shim exports it as
	// FLOW_GUARD_REPO_ROOT. CHECK_DISPATCH_PARAGRAPHS_ROOT is an explicit,
	// opt-in override honoured only when set, so the tests can point this
	// guard at a sandboxed fixture tree -- never set it for a normal
	// invocation. Set but empty is a refusal, not the default.
	root, set := env.LookupEnv("CHECK_DISPATCH_PARAGRAPHS_ROOT")
	switch {
	case set && root == "":
		fmt.Fprintf(stderr, "%s: CHECK_DISPATCH_PARAGRAPHS_ROOT is set but empty\n", dpName)
		return 2
	case !set:
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprintf(stderr, "%s: FLOW_GUARD_REPO_ROOT is unset — run flow-guard check-dispatch-paragraphs, which sets it\n", dpName)
			return 2
		}
	}
	// Paths print as the bash built them ("$ROOT/<site>"); only the file
	// system calls see a relative root joined to the working directory,
	// uncleaned so the kernel resolves any "..", as bash's did.
	fsPath := func(p string) string { return smcAbs(env, p) }
	if !gsReadableDir(fsPath(root)) {
		fmt.Fprintf(stderr, "%s: %s is not a readable directory — cannot scan\n", dpName, root)
		return 2
	}

	var violations []string
	for _, site := range dpSites {
		full := root + "/" + site.path
		lines, refusal := dpReadSite(fsPath(full), dpEntries[site.entry].label)
		if refusal != "" {
			// A hard "cannot answer at all" refusal, in the same file:line
			// shape as every other finding: an access failure at one required
			// site means the whole run cannot answer, not that this one site
			// failed -- so the violations gathered so far are never printed.
			fmt.Fprintf(stderr, "%s:0: %s\n", full, refusal)
			return 2
		}
		violations = append(violations, dpCheckSite(site, full, lines)...)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(stdout, v)
		}
		fmt.Fprintf(stdout, "DISPATCH-PARAGRAPHS-INVALID: %s — %d violation(s)\n", root, len(violations))
		return 1
	}
	fmt.Fprintf(stdout, "DISPATCH-PARAGRAPHS-OK: %s — %d site(s) validated\n", root, len(dpSites))
	return 0
}

// dpReadSite reads one required site's lines, or names why it cannot: a
// symlink, a missing path, a non-regular file and an unreadable one are each
// refused -- never a silent skip, which would be a vacuous "✓ clean". A
// path that exists as neither a file nor a directory folds into "not a
// regular file".
func dpReadSite(p, label string) ([]string, string) {
	if fi, err := os.Lstat(p); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return nil, "is a symlink — a required dispatch-paragraph site must be a real file, never a symlink"
	}
	fi, err := os.Stat(p)
	if err != nil {
		return nil, "does not exist — this is a required dispatch-paragraph site"
	}
	if !fi.Mode().IsRegular() {
		return nil, "exists but is not a regular file (neither file nor a symlink to check) — cannot scan"
	}
	if syscall.Access(p, pcReadOK) != nil {
		return nil, "is not readable — cannot scan"
	}
	data, err := os.ReadFile(p)
	if err != nil {
		// The bash reported a grep exit >= 2 here; the port reads the file
		// itself, so the refusal names the read error instead.
		return nil, fmt.Sprintf("could not be read while scanning for the label \"%s\" (%v) — a failure to look, not an absence", label, err)
	}
	lines := strings.Split(string(data), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, ""
}

// dpCheckSite returns one site's violations. A block counts toward a required
// variant only when it carries the label, every shared phrase, AND that
// variant's own phrase: two blocks of the same variant do NOT satisfy a
// two-variant site. Every failure is reported as `file:line` naming the
// missing element -- the label, the specific missing phrase, or the missing
// variant.
func dpCheckSite(site dpSite, full string, lines []string) []string {
	e := dpEntries[site.entry]
	var out []string
	blocks := 0
	hasVariant := map[string]bool{}
	for i, line := range lines {
		// The label is matched on the raw line, as the script's header states
		// and as GNU grep -aF and LC_ALL=C did; the text walked below is
		// mapfile's view of it, which ends at the first NUL byte. One
		// divergence from the bash at d71a2327: macOS BSD grep 2.6.0 in a
		// UTF-8 locale finds no non-ASCII label (the em-dash of **VERBATIM
		// REPORT — THE FACT:**) anywhere in a file holding a NUL byte, so the
		// bash reported such a label missing where this finds it.
		if !strings.Contains(line, e.label) {
			continue
		}
		blocks++
		text := dpBlockText(lines, i)
		missing := false
		for _, phrase := range e.shared {
			if !strings.Contains(text, phrase) {
				missing = true
				out = append(out, fmt.Sprintf("%s:%d: block carrying \"%s\" is missing the required phrase: \"%s\"", full, i+1, e.label, phrase))
			}
		}
		if !missing {
			for v, phrase := range e.variants {
				if strings.Contains(text, phrase) {
					hasVariant[v] = true
				}
			}
		}
	}
	if blocks < site.min {
		out = append(out, fmt.Sprintf("%s:0: requires at least %d block(s) carrying the label \"%s\", found %d", full, site.min, e.label, blocks))
	}
	for _, v := range site.variants {
		if !hasVariant[v] {
			out = append(out, fmt.Sprintf("%s:0: missing a block that satisfies the %s variant (the label \"%s\" plus every shared phrase plus \"%s\")", full, v, e.label, e.variants[v]))
		}
	}
	return out
}

// dpBlockText is a block's text: the label line plus every immediately
// following blockquote-continuation line, each stripped of its `>` and one
// space and joined with single spaces (matching how a wrapped markdown
// paragraph reads as prose), so a phrase split across two source lines by
// the wrap is still found as one substring. A label line that is not itself
// a blockquote line is a block of one line. A blockquote indented under a list
// item counts as one: leading spaces before its `>` are dropped (kan-1033).
//
// The continuation walk stops at the first line that ITSELF carries any known
// label, even though that line still begins with '>'. Without this, two
// required blocks glued together by a missing blank line (a future edit's
// copy-paste slip) merge into one block, and a paraphrase in the first that
// quietly drops a required phrase can pass by absorbing the second block's
// own text -- the exact failure class this guard exists to catch, defeated
// by its own continuation rule.
func dpBlockText(lines []string, start int) string {
	quoted := strings.HasPrefix(dpDedent(lines[start]), ">")
	var parts []string
	for j := start; ; {
		cur := dpDedent(lines[j])
		if strings.HasPrefix(cur, ">") {
			cur = strings.TrimPrefix(cur[1:], " ")
		}
		parts = append(parts, cur)
		j++
		if !quoted || j >= len(lines) {
			break
		}
		next := dpDedent(lines[j])
		if !strings.HasPrefix(next, ">") || dpCarriesLabel(next) {
			break
		}
	}
	return strings.Join(parts, " ")
}

// dpDedent is dpMapfile with leading spaces dropped.
func dpDedent(line string) string { return strings.TrimLeft(dpMapfile(line), " ") }

func dpCarriesLabel(line string) bool {
	for _, e := range dpEntries {
		if strings.Contains(line, e.label) {
			return true
		}
	}
	return false
}

// dpMapfile is a line as bash's mapfile held it: a shell variable cannot
// carry a NUL, so the line ends at the first one.
func dpMapfile(line string) string {
	if i := strings.IndexByte(line, 0); i >= 0 {
		return line[:i]
	}
	return line
}
