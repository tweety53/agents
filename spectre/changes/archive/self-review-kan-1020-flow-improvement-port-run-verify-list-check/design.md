# self-review-kan-1020-flow-improvement-port-run-verify-list-check

## Context

The three findings of `proposal.md` are planned as one self-review change on kan-1020's own branch
`spectre/kan-1020-flow-improvement-port-run-verify-list-check`, in its worktree
`/Users/tweety53/Projects/agents-worktrees/kan-1020-flow-improvement-port-run-verify-list-check`,
fix base `243eca24`. The base `origin/main` is still `8c54d6d3`, kan-1020's merge base, so no path
a finding names moved upstream; no archived change records a rejection of any remedy below.

Evidence at `243eca24`:

- 1: in a scratch clone of this branch, rewording Task 3's archived Correction to carry a bare count
  makes `scripts/check-plan-provenance.sh` report that `tasks.md` line as "numeric claim with no
  measured:/predicted: provenance comment" and exit 1.
  <!-- measured: sed on the archived tasks.md in a scratch clone, then scripts/check-plan-provenance.sh @ 243eca24 -->
  The guard scans an archived change the base does not carry yet, so it already judges a
  Correction transcribed during the run; only the rule never says to run it then.
- 2: `wc -c` gives `pipeline.md` 41670 bytes and `finish-contract-run1.md` 33050.
  <!-- measured: wc -c skills/flow-contracts/pipeline.md skills/flow-contracts/finish-contract-run1.md @ 243eca24 -->
  The guards that read `skills/flow/SKILL.md` — `check-stage-mark-calls.sh`,
  `check-loop-dispatch.sh`, `check-self-review-report.sh`, `check-fast-route-record.sh` — parse stage
  marks and dispatch rows, never the load paragraph; `check-verbatim-moves.sh` judges any new
  sentence there.
- 3: kan-1020's archived `narrative.md` records Task 2's byte-identical case passing with the mtime
  pin removed, because the harness re-touched the PNG inside zip's DOS-time granularity; the shipped
  case (`stats/internal/guard/rebuild_full_app_suite_zip_test.go`, case 2) moves the mtime back an
  hour. Task 2's entry carries no Correction.

## Live check

None — the change edits skill prose and one archived task record; no service or store holds its
state.

## Decisions

### A Correction runs the provenance guard when transcribed

**ID:** correction-provenance-at-transcription
**Status:** active
**Chosen:** one sentence appended to `implement.md`'s Correction paragraph: a Correction is plan
prose `check-plan-provenance.sh` reads — run that guard when transcribing it, and give each numeric
claim its `measured:` comment or word it without a count. No new guard or test: the existing guard
already refuses the shape, and the defect is only when it runs; a prose rule has no behaviour a test
can pin.
**Considered:** running `check-plan-provenance.sh` inside `check-task-commit-fields.sh` at task
close — changes a guard verdict at every task close in every project for a cost of one lint re-run;
a new guard over Correction paragraphs — duplicates the provenance guard's numeric-claim classifier.

### Contract files load through the file-read tool

**ID:** contract-load-via-read
**Status:** active
**Chosen:** one sentence added to `skills/flow/SKILL.md` right after the `pipeline.md` load
paragraph: load every contract file with the harness's file-read tool, Read on Claude Code, in
offset chunks wherever one read cannot return it whole, never with `cat`. Added, not reworded, so
the load directive stays byte-identical. Named "the harness's file-read tool" because the skill also
runs on ZCode and Muse.
**Considered:** splitting `pipeline.md` below the cap — a relocation across many citers, `big`; the
same sentence in every skill that loads a contract — `SKILL.md` is the router every `/flow` run
loads first, the one place the first load happens.

### kan-1020's Task 2 Correction is written at plan time

**ID:** repair-task-2-correction-at-plan-time
**Status:** active
**Chosen:** the planner appends Task 2's `Correction (2026-10-22):` paragraph to kan-1020's archived
`tasks.md`, after its Step 6, worded without counts, and commits it with this plan; Task 1's verify
re-runs `check-plan-provenance.sh` over it, so finding 3 is carried by Task 1.
**Considered:** a fixer task of its own — a fixer never writes `spectre/changes/`, and a task whose
`**Files:**` lie only under the planning paths fails `check-plan-shape.sh` F14.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — two prose sentences and a record repair; no code, guard verdict or test moves.
**Considered:** flow-medium — fits a changed guard verdict or code path, which this change has none
of.

### Fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch, g1: Tasks 1 and 2 — both are one-sentence prose edits checked by the
same guards.
**Considered:** one dispatch per task — two cold starts for two sentences.

## Open questions
