#!/usr/bin/env bash
# check-dispatch-paragraphs.sh — the guard that keeps a required dispatch
# paragraph from silently disappearing.
#
# Why this exists: KAN-289's root cause was that the REPRODUCE, DON'T READ
# instruction lived in no template and depended on the dispatcher
# remembering to type it. Part 1 of that change fixed it by adding the
# paragraph, verbatim, at dispatch sites. Nothing then stops a LATER prose
# edit from trimming a required paragraph away one line at a time — this
# guard is what makes that loud instead of silent, the same role
# check-guard-symlinks.sh plays for its own templates. KAN-217 added a second required paragraph — the VERBATIM
# REPORT — THE FACT blockquote that tells the fix subagent its slot's
# report file outranks the dispatcher's summary — and generalized this
# guard from one hard-coded paragraph to a table of them so a third would
# cost a table row, not a second copy of the machinery. KAN-263 added a
# third required paragraph — FOREGROUND BUILDS, which forbids a dispatched
# agent from ending its turn with a build/test/long-running command still
# running in the background — at all four dispatch sites that can run one:
# the implementer dispatch, the conductor's own §4 instruction, the review
# panel's slot dispatch, and the panel-fix subagent dispatch. KAN-441 added a
# fourth required paragraph — TARGETED TESTS, which tells an implementer or
# panel-fix subagent to run only the task's own tests through the build
# tool's own selector rather than a whole module or repository suite mid-task
# — at the implementer dispatch and the panel-fix subagent dispatch. KAN-464
# added a fifth required paragraph — MUTATION PROOF, which hands the fix
# round's mutation-proof obligation to the panel-fix subagent itself rather
# than the conductor — at the panel-fix subagent dispatch alone. KAN-473
# added a sixth required paragraph — TOOLS, which tells every dispatched
# role to load a deferred tool by name in its first turn rather than a
# later wildcard ToolSearch, since a schema loaded mid-session changes the
# tools array ahead of every message and re-prices the whole context — at
# six sites across two new site files: the conductor dispatch and the
# implementer dispatch in implement.md, the panel slot dispatch and the
# panel-fix subagent dispatch in review-panel.md, the planner dispatch in
# brainstorm.md, and the verifier dispatch in verify-and-handoff.md.
# KAN-257 extended the MUTATION PROOF entry from three shared phrases to
# six, carrying the fix round's assert obligation: a mutation must confirm
# its edit landed before the tests run, and an edit that never applied is
# a refusal to redo — never a surviving mutant. KAN-472 task 10 added a
# seventh required paragraph — MODEL HANDSHAKE, the universal `Model:
# <model>` first-reply obligation stated once in implement.md's **Dispatch
# the conductor** section and cited from every other dispatch site — at the
# same six sites the TOOLS paragraph occupies: implement.md twice (conductor
# dispatch, implementer dispatch), review-panel.md twice (panel slot
# dispatch, panel-fix subagent dispatch), brainstorm.md once (planner
# dispatch), verify-and-handoff.md once (verifier dispatch). KAN-472
# task 20 added an eighth required paragraph — INDEPENDENT PASSES, which
# forbids one pass of a bundled review dispatch from citing, deferring to, or
# skipping a defect an earlier pass in the same bundle already raised — at
# one site, review-panel.md's bundle prompt, min 1 block. KAN-484 added a
# ninth required paragraph — NO DELEGATION, which tells every leaf agent the
# conductor dispatches to do the work itself and never call the Agent tool
# or spawn a subagent, closing the conductor's closed list one level down —
# at four sites: the implementer dispatch in implement.md, the panel slot
# dispatch and the panel-fix subagent dispatch in review-panel.md, and the
# verifier dispatch in verify-and-handoff.md. KAN-400 re-introduced the
# per-task reviewer as a gated dispatch — one reviewer per task commit
# whose review gate in skills/flow/implement.md fires — as implement.md's
# fifth dispatch site, raising that file's minimums by one for REPRODUCE,
# DON'T READ (the reviewer variant now required twice there), FOREGROUND
# BUILDS, TOOLS, MODEL HANDSHAKE and NO DELEGATION. implement.md's gate
# definition is the one statement of the gate's threshold and risk arm,
# and the one site to re-tune; this file pins only the paragraphs the
# gated reviewer dispatch carries, never the gate's own numbers. KAN-496
# added a tenth required paragraph — PIXEL PROBE, which requires any fix
# the panel-fix round lands to draw/geometry code to carry a probe
# assertion against the actual rendered pixels or geometry, and rejects a
# fix commit that lacks one at the fix step rather than discovering the
# regression next round — at the panel-fix subagent dispatch in
# review-panel.md alone, the same single site MUTATION PROOF occupies.
# KAN-328 added an eleventh required paragraph — PROVE THE GUARD BITES,
# which tells an implementer whose tests assert on configuration or file
# content that a passing run is not evidence: break the property the
# assertion protects, run the guard or the test against the broken state and capture
# its failure, restore, capture the pass — at the implementer dispatch in
# implement.md alone, beside the TDD and systematic-debugging REQUIRED
# SUB-SKILL paragraphs. KAN-405 added a twelfth required paragraph —
# REPORT, DON'T DECIDE, which tells an implementer that a question only the
# operator can settle is reported, never decided in code: carried to the
# REPORT FILE where the plan can proceed, a BLOCKED report where it cannot,
# because a decision the implementer makes silently is one nobody recorded —
# at the implementer dispatch in implement.md alone, the same single site
# PROVE THE GUARD BITES occupies. KAN-521 added a thirteenth required
# paragraph — ENTRY CONTEXT, which gives every panel slot dispatch a named
# entry context (the diff and the inlined touched-files list, the contracts
# among them named) so a reviewer's exploration starts scoped instead of from
# the whole tree, and has the reviewer state the coverage trade explicitly in
# its report — at the panel slot dispatch in review-panel.md alone, carried
# into every bundled dispatch by the shared-paragraphs list. KAN-564 added a
# fourteenth required paragraph — FINDINGS ARE INPUT, which tells the
# panel-fix subagent that a finding names a defect and suggests a route to
# fixing it — findings are input, not orders — that the subagent's obligation
# is to resolve the defect the finding names, and that a route deviating from
# the finding's literal suggestion is taken only when the report records the
# deviation and justifies it, so a justified deviation like kan-469's F2
# (shared helpers over the literal route that would have reordered a required
# clause ordering) stays legal while a silent one does not — at the
# panel-fix subagent dispatch in review-panel.md alone, the same single site
# MUTATION PROOF and PIXEL PROBE occupy. KAN-593 added a fifteenth required
# paragraph — CONTEXT BUNDLE FAILURE, which makes the review panel's
# pre-flight bundle rebuild fail loud instead of silent: a non-zero gather
# exit or an absent bundle file records the cause in the pass log and then
# stops the stage or takes an explicit operator override before any slot is
# dispatched, so the reviewers' context is never silently reduced — at the
# panel pre-flight in review-panel.md alone. KAN-635 added a sixteenth
# required paragraph — READ-ONLY REVIEW, which forbids the gated per-task
# reviewer from mutating the worktree it reviews: no writing git command, no
# commit, no index change, no edit outside its own report file, because a
# reviewer that mutates the tree (gymie KAN-635's `git checkout <sha> -- .`
# destroyed uncommitted planning artifacts) and self-reports a restore is a
# data-loss shape nobody checked — at the gated per-task reviewer dispatch
# in implement.md alone; the panel slots carry the same rule inside their
# prompt files' own Read-Only Review sections, which this table does not
# pin.
#
# Argument-free and self-scoped, exactly like check-guard-symlinks.sh: the
# scan root is this repository's own root, resolved from this script's own
# location, so no call site can narrow it. CHECK_DISPATCH_PARAGRAPHS_ROOT is
# an explicit, opt-in override honored only when set (mirrors
# CHECK_GUARD_SYMLINKS_ROOT), so the Go tests
# (stats/internal/guard/check_dispatch_paragraphs_test.go) can point this
# guard at a sandboxed fixture tree under TMPDIR without touching this
# repository — never set it for a normal invocation.
#
# THE PARAGRAPH TABLE. Each entry is a label, its shared phrases (required
# of every block carrying that label, held as short literals rather than a
# copy of the whole blockquote, per design.md's `label-plus-phrases`), and
# its variants — each variant a name plus its own load-bearing phrase, or no
# variants at all when every block of that label is equivalent. Every entry
# lists the sites that require it, each site naming its minimum block count
# and which variants it requires.
#
#   The verifier dispatch moved from verify-and-handoff.md to visual-verify.md
#   when the visual-verification procedure split into its own on-demand file,
#   so its three sites below name the new file; the tooling-analyst dispatch
#   moved on to visual-verify-tooling-analysis.md (KAN-858), which loads only
#   on a fix run with a miss, so each file carries one block of each.
#
#   Paragraph                          Site                        Min  Variants
#   **REPRODUCE, DON'T READ:**         skills/flow/review-panel.md  1   reviewer
#   **REPRODUCE, DON'T READ:**         skills/flow/implement.md     3   reviewer AND implementer
#   **VERBATIM REPORT — THE FACT:**    skills/flow/review-panel.md  1   (none)
#   **FOREGROUND BUILDS:**             skills/flow/implement.md     3   (none)
#   **FOREGROUND BUILDS:**             skills/flow/review-panel.md  2   (none)
#   **TARGETED TESTS:**                skills/flow/implement.md     1   (none)
#   **TARGETED TESTS:**                skills/flow/review-panel.md  1   (none)
#   **MUTATION PROOF:**                skills/flow/review-panel.md  1   (none)
#   **PIXEL PROBE:**                   skills/flow/review-panel.md  1   (none)
#   **TOOLS:**                         skills/flow/implement.md     2   (none)
#   **TOOLS:**                         skills/flow/review-panel.md  2   (none)
#   **TOOLS:**                         skills/flow/visual-verify.md 1   (none)
#   **TOOLS:**                         skills/flow/visual-verify-tooling-analysis.md 1 (none)
#   **MODEL HANDSHAKE:**               skills/flow/implement.md     2   (none)
#   **MODEL HANDSHAKE:**               skills/flow/review-panel.md  2   (none)
#   **MODEL HANDSHAKE:**               skills/flow/visual-verify.md 1   (none)
#   **MODEL HANDSHAKE:**               skills/flow/visual-verify-tooling-analysis.md 1 (none)
#   **INDEPENDENT PASSES:**            skills/flow/review-panel.md  1   (none)
#   **NO DELEGATION:**                 skills/flow/implement.md     2   (none)
#   **NO DELEGATION:**                 skills/flow/review-panel.md  2   (none)
#   **NO DELEGATION:**                 skills/flow/visual-verify.md 1   (none)
#   **NO DELEGATION:**                 skills/flow/visual-verify-tooling-analysis.md 1 (none)
#   **PROVE THE GUARD BITES:**         skills/flow/implement.md     1   (none)
#   **REPORT, DON'T DECIDE:**          skills/flow/implement.md     1   (none)
#   **ENTRY CONTEXT:**                 skills/flow/review-panel.md  1   (none)
#   **FINDINGS ARE INPUT:**            skills/flow/review-panel.md  1   (none)
#   **CONTEXT BUNDLE FAILURE:**        skills/flow/review-panel.md  1   (none)
#   **OUTPUT BUDGET:**                 skills/flow/implement.md     1   (none)
#   **OUTPUT BUDGET:**                 skills/flow/review-panel.md  1   (none)
#   **READ-ONLY REVIEW:**              skills/flow/implement.md     1   (none)
#
#   REPRODUCE, DON'T READ shared phrases: "crosses a boundary", "the store,
#   the filesystem, a guard, a real transcript", "exercise the real thing"
#     - reviewer variant's own phrase: "worth less than one you did"
#     - implementer variant's own phrase: "a fake or a hand-built value"
#
#   VERBATIM REPORT — THE FACT shared phrases (no variants — every block
#   carrying the label must carry all three): "the reviewer's own report",
#   "never a source of fact", "the report wins"
#
#   FOREGROUND BUILDS shared phrases (no variants — every block carrying the
#   label must carry all three): "still executing in the background", "Run
#   it in the foreground", "poll it to completion". Required three times in
#   implement.md (implementer dispatch, the gated per-task reviewer dispatch,
#   the parent's own §4 restatement) and twice in review-panel.md (panel slot
#   dispatch, panel-fix subagent dispatch).
#
#   OUTPUT BUDGET shared phrases (no variants — every block carrying the
#   label must carry all four): "re-reads your", "by line range", "never
#   re-read a file already in your context", "Never print a generated
#   file". Required once in each of implement.md (implementer dispatch) and
#   review-panel.md (panel-fix subagent dispatch) — the two dispatches whose
#   tool output grows with the task: every turn re-reads the whole context,
#   so an uncapped read or log is paid for on every turn after it.
#
#   TARGETED TESTS shared phrases (no variants — every block carrying the
#   label must carry all three): "the build tool's own selector", "once for
#   RED, once for GREEN", "module or repository suite mid-task". Required
#   once in each of implement.md (implementer dispatch) and
#   review-panel.md (panel-fix subagent dispatch) — the panel slot dispatch
#   is not a site: reviewers do not run the task's tests.
#
#   MUTATION PROOF shared phrases (no variants — every block carrying the
#   label must carry the full set of seven): "mutation-proved before you end your
#   turn", "confirm an existing test fails, and restore", "a surviving
#   mutant", "confirm the edit landed", "a refusal, not a surviving
#   mutant", "never buys a test", and the line shape itself,
#   "`fix-mutation: <path> — <what was mutated> — <the test that failed>`" —
#   the fix subagent never receives the contract's fenced block, so its
#   dispatch states the shape (KAN-854). Required once, at the panel-fix subagent
#   dispatch in review-panel.md alone — the implementer dispatch and the
#   panel slot dispatch are not sites: the fix round is the only one this
#   obligation binds.
#
#   PIXEL PROBE shared phrases (no variants — every block carrying the
#   label must carry all three): "draw/geometry code", "actual rendered
#   pixels or geometry", "rejected at the fix step". Required once, at the
#   panel-fix subagent dispatch in review-panel.md alone — the implementer
#   dispatch and the panel slot dispatch are not sites: the fix round is
#   the only dispatch that lands a fix commit.
#
#   PROVE THE GUARD BITES shared phrases (no variants — every block
#   carrying the label must carry all three): "assert on configuration or
#   file content", "break the property the assertion protects", "against
#   the broken state". Required once, at the implementer dispatch in
#   implement.md alone — the gated per-task reviewer dispatch is not a
#   site: reviewers judge a diff, they do not run the task's guards.
#
#   REPORT, DON'T DECIDE shared phrases (no variants — every block
#   carrying the label must carry all four): "only the operator can
#   settle", "reported, never decided in code", "carry the question in your
#   REPORT FILE", "report BLOCKED". Required once, at the implementer
#   dispatch in implement.md alone — the gated per-task reviewer dispatch
#   is not a site: reviewers judge a diff, they do not settle operator
#   questions.
#
#   CONTEXT BUNDLE FAILURE shared phrases (no variants — every block
#   carrying the label must carry all five): "exited non-zero, or the
#   bundle file is still", "context bundle: build failed", "dispatch every
#   slot without the bundle", "context bundle: dispatched without it —
#   operator override", "outcome stopped". Required once, at the panel
#   pre-flight in review-panel.md alone — the fix subagent's rebuild takes
#   the same path through its existing "same as above", so it needs no
#   second block.
#
#   READ-ONLY REVIEW shared phrases (no variants — every block carrying the
#   label must carry all three): "never mutate it", "no file edit outside
#   your own report file", "a claim nobody can check". Required once, at the
#   gated per-task reviewer dispatch in implement.md alone — the panel slot
#   dispatch is not a site: its slots carry the rule inside their prompt
#   files' own Read-Only Review sections, not a shared paragraph this table
#   pins.
#
#   TOOLS shared phrases (no variants — every block carrying the label
#   must carry all three): "in your first turn", "never a wildcard
#   query", "re-prices your whole context". Required twice in implement.md
#   (implementer dispatch, gated per-task reviewer dispatch — the
#   conductor's own copy, and the planner's in
#   brainstorm.md, were removed with those roles, kan-488), twice in
#   review-panel.md (panel slot dispatch, panel-fix subagent dispatch),
#   and once each in visual-verify.md (verifier dispatch) and
#   visual-verify-tooling-analysis.md (tooling-analyst dispatch).
#
#   MODEL HANDSHAKE shared phrases (no variants — every block carrying the
#   label must carry all three): "the first line of your first reply",
#   "and nothing else on that line", "before any tool call". Required twice
#   in implement.md (implementer dispatch, gated per-task reviewer
#   dispatch — the conductor's own copy, and
#   the planner's in brainstorm.md, were removed with those roles,
#   kan-488), twice in review-panel.md (panel slot dispatch, panel-fix
#   subagent dispatch), and once each in visual-verify.md (verifier
#   dispatch) and visual-verify-tooling-analysis.md (tooling-analyst
#   dispatch) — the same sites TOOLS occupies.
#
#   INDEPENDENT PASSES shared phrases (no variants — every block carrying the
#   label must carry all three): "starts from `final-review.diff`", "raise it
#   again under this pass", "before beginning the next pass". Required once,
#   at review-panel.md's bundle prompt alone — the paragraph binds every pass
#   of a bundled dispatch to read only the diff and the code, never an
#   earlier pass's report.
#
#   NO DELEGATION shared phrases (no variants — every block carrying the
#   label must carry all three): "Never call the `Agent` tool", "never spawn
#   a subagent", "the leaf of this run". Required twice in implement.md
#   (implementer dispatch, gated per-task reviewer dispatch), twice in
#   review-panel.md (panel slot dispatch,
#   panel-fix subagent dispatch), once each in visual-verify.md (verifier
#   dispatch) and visual-verify-tooling-analysis.md (tooling-analyst
#   dispatch) — there is no conductor or planner dispatch left to be a site
#   (kan-488): the parent orchestrates directly and runs brainstorming
#   inline.
#
#   ENTRY CONTEXT shared phrases (no variants — every block carrying the
#   label must carry all three): "your named entry context", "not from a
#   whole-tree exploration", "the coverage trade". Required once, at the
#   panel slot dispatch in review-panel.md — the bundle prompt's
#   shared-paragraphs list carries it into every bundled dispatch, the same
#   way REPRODUCE, DON'T READ is carried.
#
#   FINDINGS ARE INPUT shared phrases (no variants — every block carrying
#   the label must carry all four): "input, not orders", "resolve the defect
#   the finding names", "records the deviation and justifies it", "A silent
#   deviation is an unfixed finding". Required once, at the panel-fix
#   subagent dispatch in review-panel.md alone — the implementer dispatch
#   and the panel slot dispatch are not sites: the fix round is the only
#   dispatch that acts on findings.
#
# A BLOCK is a line carrying a label, plus every immediately-following line
# that continues the same markdown blockquote (a line beginning with `>`) —
# i.e. the whole paragraph. A block counts toward a required variant only
# when it carries the label, every shared phrase, AND that variant's own
# phrase. Two blocks of the same variant do NOT satisfy a two-variant site —
# implement.md requires one reviewer block and one implementer block, not
# merely two blocks. An entry with no variants needs no variant match: every
# phrase it lists is simply required of the block.
#
# Every failure is reported as `file:line` naming the missing element — the
# label, the specific missing phrase, or the missing variant. Exit `0`
# clean, `1` a required site missing its block or one of its phrases, `2` it
# cannot answer at all: a scoped file missing, unreadable, a symlink, or a
# read failing for any reason. A scoped path that
# exists as neither a file nor a directory is folded into the same hard `2`
# ("not a regular file") — never a silent skip, which would be a vacuous
# "✓ clean".
#
# WHAT A GREEN RUN DOES NOT PROVE: only that each label and its phrases are
# present at each required site — never that a dispatcher actually wrote the
# report file described by VERBATIM REPORT — THE FACT, never that a fix
# agent read one, and never that a reviewer or an implementer actually
# obeyed REPRODUCE, DON'T READ. A paraphrase that drops one of the listed
# phrases fails loud; a paraphrase that keeps all of them while changing the
# surrounding prose passes clean.
#
# Ported to Go (KAN-841): the paragraph table is dpEntries and dpSites, and
# the body's reasoning for every step is in
# stats/internal/guard/dispatchparagraphs.go. The binary does not live in
# this checkout, so this shim exports FLOW_GUARD_REPO_ROOT — this script's
# parent directory, derived as the bash guard derived its ROOT — as the root
# scanned when CHECK_DISPATCH_PARAGRAPHS_ROOT is unset.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-dispatch-paragraphs: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-dispatch-paragraphs 2 "check-dispatch-paragraphs:" "$@"
