## Context

`skills/flow-self-review/SKILL.md` is canonical for the self-review pass, its fix loop and its
report; `skills/flow-contracts/finish-contract-run1.md` (**Run the self-review pass**),
`skills/flow/integrate.md` step 4 and `skills/flow-fast/SKILL.md` assemble the in-run bundle and
hand it to that file. The pass, planner, fixers and reviewers are already one-shot dispatches, no
subagent waits on another, and every loop dispatch is admitted and recorded with its model and
effort by `check-loop-dispatch.sh` (`stats/internal/guard/loop.go`). What is left is the content
main carries itself — the bundle, the case files, the extractions and every quoted report — and
the fixer pair the ticket names. Both live in the same three skill files and the one policy table,
so they are one change.

## Decisions

### Main passes file paths, never file contents

**ID:** paths-not-contents
**Status:** active
**Chosen:** every handoff between the self-review's steps is a file in `<sr-dir>`; the pass writes `findings.md`, `report-draft.md`, `cases/<case-id>/` and angle 7's extractions; reviewers write `review-<r>.md`; a red verify writes `verify-<v>.log`; main reads only hand-backs, the `## Offered` section and the filled draft — it is the one change that removes the bundle and every quoted report from main's context while main keeps every judgement.
**Considered:** a coordinator subagent running the loop — rejected, the ticket requires that only main waits; a Go command turning report lines into store rows — rejected for now, main records them in one Bash call, added if that proves expensive.

### Where the handoff files live

**ID:** handoff-dir-state-dir
**Status:** active
**Chosen:** `<state-dir>/self-review/<name>/` — outside every checkout, so nothing can be staged by accident, and beside angle 7's existing `<state-dir>/flow-eval/<name>/`.
**Considered:** `<worktree>/.superpowers/sdd/self-review/` — a standalone run with no non-`big` finding has no worktree; the session scratchpad — harness-specific and unreachable by a later standalone run.

### Self-review fixer pair

**ID:** self-review-fix-pair
**Status:** active
**Chosen:** `self-review-fix` → `opus/flow-medium`, `sonnet/flow-medium` (sonnet only for a simple, predictable fix, its one-line reason recorded with the dispatch); the `opus/flow-high` `No-progress escalation` row is deleted, and a no-progress fix is answered by the fresh re-plan step 3 already prescribes. The review rows and every `pipeline-fix-*` row stay as they are — the ticket names fixers only, and the first review's effort is already a recorded per-loop decision.
**Considered:** also moving both review rows to medium only, per the operator memory's "and reviewers" — rejected, it widens the change past the ticket and removes the `first-review-effort` low option the guard parses; keeping the table and updating the memory instead — rejected, it rejects the ticket's acceptance line.

### No review merger in the self-review loop

**ID:** no-review-merger
**Status:** active
**Chosen:** no merger — the loop runs one reviewer per round, so `panel-merger` would have nothing to de-duplicate and would add a dispatch and a hop of latency; the merger's benefit that matters here, main never reading the findings, comes from the reviewer writing its report to `<sr-dir>/review-<r>.md` and handing back only its counts. This answers the ticket's question.
**Considered:** a merger per review round — an extra dispatch with no input to merge; a merger only once the loop gains several reviewers per round — speculative, revisited then.

## Open questions

## Recommended approach

Rewrite the self-review's handoffs as files in `<sr-dir>` (decision `paths-not-contents`, in
`<state-dir>` per `handoff-dir-state-dir`): the in-run bundle sites write the bundle to
`<sr-dir>/bundle.md`; `skills/flow-self-review/SKILL.md` has the pass name that path and write its
findings, a report draft with placeholders, the case files and the extractions there, step 2a copy
and render the cases, the loop's prompts name files, and main fill and copy the draft. In the same
change, move `self-review-fix` to `flow-medium` with no escalation row (`self-review-fix-pair`), pinned
by a guard test that reads the real table. A coordinator subagent and a review merger are rejected
(`paths-not-contents`, `no-review-merger`).

## Critical files

- `skills/flow-self-review/SKILL.md` — canonical self-review procedure; carries `<sr-dir>` and every handoff.
- `skills/flow-contracts/finish-contract-run1.md` — **Run the self-review pass**, the in-run bundle's assembly.
- `skills/flow/integrate.md` — step 4, the `/flow` in-run bundle hand-off.
- `skills/flow-fast/SKILL.md` — `/flow-fast`'s in-run bundle assembly.
- `skills/flow-contracts/model-policy.md` — **Allowed pairs**, the `self-review-fix` row.
- `stats/internal/guard/check_loop_dispatch_test.go` — pins the new fixer pair against the real table.
- `stats/internal/guard/pipeline_fix_test.go` — its self-review `escalate` case now expects a refusal.
- `stats/internal/guard/loop.go` — `check-loop-dispatch`, reads the table; unchanged, depended on.
- `scripts/pipeline-fix.sh` — `end <key> <report-file>` and `escalate`; unchanged, depended on.

## Existing code to reuse

- `flow self-review bundle -change <name>` — assembles the bundle; redirected to a file instead of read.
- `flow state dir` — resolves `<state-dir>`, as angle 7 already does.
- `scripts/pipeline-fix.sh end <key> <report-file>` — already closes a review from its report file's `Critical:`/`Important:` lines.
- `flow prompt-tuning render` and `scripts/check-prompt-tuning-cases.sh` — step 2a's render and check.
- `land-self-review-report.sh` and `scripts/check-self-review-report.sh` — land and check the copied draft.
- `stats/internal/guard/loop.go` `policyAllows` / `lpEscalation` — already refuse a pair off the table and admit an escalation only on a marked row, so deleting the row needs no code change.
- The `lpNew` / `lpRowsJSON` fixtures in `stats/internal/guard/*_test.go`.

## Verification

- `cd stats && make -s web-dist && go test ./internal/modelpolicy/... ./internal/guard/ -count=1` — `TestCheckLoopDispatchSelfReviewFixPairs` admits `opus`/`sonnet` `flow-medium` fixes and refuses `flow-low` and an escalated `flow-high` with `PAIR —`; `TestPipelineFixOwnBranchInAnotherProject` sees `escalate` of a self-review fix refused.
- `scripts/check-verbatim-moves.sh` — every added or reworded sentence is the one `verbatim-moves.txt` lists.
- `scripts/run-verify-list.sh <worktree> lint` — every prose guard.
- After landing (not an in-run guard): the `flow.self-review` main-session vs dispatch cost split over the following runs, compared with the 30-day baseline of $821 across 135 runs, about half in main.
  <!-- measured: flowd stage cost split for flow.self-review over the 30 days to 2026-10-30 @ machine-local flowd store read during kan-1064 planning, cannot be re-run -->

## Live check

None — the change edits skill prose, one row of the policy table and guard tests; no daemon, store or scheduler behaves differently until a later self-review runs, and the table's effect on `check-loop-dispatch` is exercised against the real `model-policy.md` by the guard tests above. The cost figure is measured after landing.
