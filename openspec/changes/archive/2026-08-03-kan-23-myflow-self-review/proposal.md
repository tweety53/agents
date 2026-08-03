## Why

Once a change reaches `FINISHED`, nothing captures what the run itself revealed: what problems came
up and how the pipeline could avoid them, what drove token/time cost and what would cut it without
losing quality, what went well worth repeating, and what could be scripted instead of reasoned
through by hand. KAN-23 asks for a retrospective step that surfaces these, offers to file
actionable findings as Jira issues, and lets the operator rate the run.

## What Changes

- `/myflow-finish` run 2 gains a new step immediately after the `FINISHED` write: a self-review
  pass, folded into the existing run rather than a new command or pipeline state.
- A skip prompt (default: run it) lets the operator opt out per invocation.
- A new deterministic script, `scripts/gather-self-review-context.sh`, mechanically collects the
  SDD ledger, the review-panel record, `tasks.md` and the relevant git log into one bundle — no LLM
  tokens spent scanning for it.
- One combined reasoning pass (not four separate dispatches) answers all four angles from KAN-23
  using that bundle plus the live session's own context, and produces a report.
- Each actionable finding gets its own Jira-issue filing ask (default: don't file), reusing the
  existing labelling and never-blocking rules.
- The operator is asked to rate the run 1-5; the rating is recorded only in the report, not in the
  state file.
- The report is written to `docs/self-review/<name>-self-review.md`, committed and pushed on the
  base branch.
- Run 2's terminal handoff gains one line naming the report path and rating, or `skipped`.

## Capabilities

### New Capabilities
- `myflow-self-review`: the self-review step itself — trigger and placement, the skip prompt,
  deterministic context gathering, the combined reasoning pass, per-finding Jira filing, the
  operator rating, the report file, and the handoff line.

### Modified Capabilities
- `myflow-finish-cleanup`: run 2's procedure gains a step, after the `FINISHED` write, that invokes
  self-review. Self-review never blocks or delays `FINISHED` and never re-opens a finished change.

## Impact

- `skills/myflow-finish/SKILL.md` — run 2 gains step 8.
- `skills/myflow-contracts/pipeline.md` — Finish contract's run-2 outline and terminal handoff
  template gain the self-review step and line.
- New script: `scripts/gather-self-review-context.sh` (+ its test, per this repo's `## test`
  convention in `.myflow/project.md`).
- New directory: `docs/self-review/`.
- `skills/myflow-contracts/jira-integration.md` — no rule changes; the new filing ask reuses
  **Labels on issues the pipeline creates** and **Never blocking** verbatim.
- No state-file schema change, no new pipeline state, no new command.
