# Self-review context bundle for kan-891-flow-fix-state-the-verdict-capture-rule-a-guard

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-891-flow-fix-state-the-verdict-capture-rule-a-guard, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-891-flow-fix-state-the-verdict-capture-rule-a-guard.md (absent)
skipped: .superpowers/sdd/reviews/kan-891-flow-fix-state-the-verdict-capture-rule-a-guard-panel.md (absent)
skipped: spectre/changes/archive/kan-891-flow-fix-state-the-verdict-capture-rule-a-guard/tasks.md (absent)
skipped: spectre/changes/archive/kan-891-flow-fix-state-the-verdict-capture-rule-a-guard/design.md (absent)
skipped: spectre/changes/archive/kan-891-flow-fix-state-the-verdict-capture-rule-a-guard/narrative.md (absent)

## git log --stat

commit 31e17971a340d780e8de630891a6e073e5859430
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 01:44:07 2026 +0300

    docs(flow): state the verdict-capture rule for guard exit statuses
    
    A guard's exit status read through a pipe is the pipe's, not the guard's —
    kan-792's panel read the docs-only guard's exit 1 through a tail pipe, the
    pipe's 0 masked the verdict, and the docs-only reduction wrongly dropped
    principles from pass 1. Stated once in review-panel.md where the panel's
    guard steps are described: capture the output to a file and read $?, use
    set -o pipefail/PIPESTATUS, or run the guard bare, so a red verdict cannot
    be laundered into green by the reader command.

 skills/flow/review-panel.md | 6 ++++++
 1 file changed, 6 insertions(+)

## Session narrative

A one-paragraph docs change: KAN-891 asks the corpus to state, once, that a guard's exit status
read through a pipe is the pipe's, not the guard's — the kan-792 panel read the docs-only
guard's exit 1 through a `tail` pipe, and the reduction silently dispatched the wrong roster.
Brainstorm located the one home by following the corpus's own structure: the panel's guard steps
are described in `skills/flow/review-panel.md` between the base-movement check and the citation
pre-check, so the rule landed as one paragraph immediately after the "one Bash call" paragraph
that governs how every verdict below it is read; no second statement was added anywhere, and the
gradlew `pipefail` exception in the workspace-isolation contract was left alone because it is a
narrower, already-canonical rule. Implementation was a single six-line insertion, committed and
pushed as one commit. The one stumble was foreseeable rather than discovered: the new prose is
run-loaded, so `check-verbatim-moves.sh` failed it as a paraphrase (4 FAIL lines), resolved the
way that guard's own header prescribes for a deliberate addition — each FAIL line listed after
`::` in the change's `verbatim-moves.txt` — after which every lint command in `.flow/project.md`'s
`## lint` list ran green in the worktree. The `## test` list ran scoped to nothing: the diff
names only a corpus markdown file, which none of the three test suites (guard harnesses, stats
Go, stats SPA) covers — the empty scope is the stated reason, not a skipped sweep. The
self-review bundle's RECORDS LOSS note about missing panel dispatch rows is expected for a
flow-fast micro run: no panel ran (the decision's panel is the string `default`), the
review-panel stage was marked as an empty pair for the stats views, and no dispatch or finding
rows were ever owed to the store.
