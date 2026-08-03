## Context

`/myflow-finish` run 2 today ends at step 7 (write `FINISHED`), followed only by the Jira `Done`
transition and the terminal handoff. KAN-23 asks for a retrospective step at that point: what
problems came up and how the pipeline could avoid them, what drove token/time cost and what would
cut it without losing quality, what went well worth repeating, and what could be scripted instead
of reasoned through — plus an offer to file actionable findings as Jira issues and an operator
rating of the run.

The pipeline's contracts (`skills/myflow-contracts/pipeline.md`) fix the command surface at three
pipeline commands and two read-only ones, and fix the state machine at three states. Any design here
has to fit inside that surface rather than widen it.

## Goals / Non-Goals

**Goals:**
- Surface the four self-review angles from KAN-23 once a change reaches `FINISHED`, without adding
  a pipeline state or a new command.
- Keep the retrospective's own cost proportionate — a self-review mechanism should not itself become
  a token/time cost problem, since that is exactly what angle #2 checks for.
- Give the operator control: a per-run skip, a per-finding filing decision, and the rating itself.
- Leave a durable, committed record independent of the (machine-local, already-terminal) state file.

**Non-Goals:**
- No new pipeline state and no new `/myflow-*` command.
- No change to the state-file schema — the rating is not written there.
- No verification/quality gate — self-review never blocks or reopens `FINISHED`.
- No four-way subagent dispatch — this is deliberately one reasoning pass, not the review panel's
  separate-lens model.

## Decisions

### Where self-review sits in the pipeline

**ID:** self-review-placement
**Status:** active
**Chosen:** inside `/myflow-finish` run 2, as step 8, after the `FINISHED` write — never blocks or
delays the archive, and never reopens a finished change.
**Considered:** before the `FINISHED` write (a self-review failure could then block the terminal
state); a new standalone `/myflow-self-review` command (collides with the fixed three-command
pipeline surface and would require rewriting the command-surface and state-transition tables for a
feature that needs no new state).

### Review structure — one pass vs. four

**ID:** self-review-pass-count
**Status:** active
**Chosen:** one combined reasoning pass covering all four angles plus the rating.
**Considered:** four distinct passes/dispatches mirroring the review panel's separate-lens model —
rejected because it would directly work against angle #2 (cost reduction), which a self-review
mechanism should not itself violate.

### Context gathering — script vs. inline reading

**ID:** self-review-context-gathering
**Status:** active
**Chosen:** a new deterministic script, `scripts/gather-self-review-context.sh`, mechanically
collects the ledger, panel record, `tasks.md` and git log; the reasoning pass only judges the
bundle it returns.
**Considered:** the skill's own prose instructing the agent to read the files itself — rejected
because every self-review run would then re-read and re-summarize the same files from scratch
in-context, for the same reason four passes were rejected.

### What the self-review analyzes

**ID:** self-review-inputs
**Status:** active
**Chosen:** the preserved artifacts (ledger, panel record, OpenSpec artifacts, archived git
history) plus the live session's own context.
**Considered:** artifacts only, for reproducibility if self-review were ever re-run standalone
later — rejected in favor of the richer signal live session context gives for "what went well".

### Task creation for findings

**ID:** self-review-task-creation
**Status:** active
**Chosen:** Jira issues, one filing ask per actionable finding, default No, reusing **Labels on
issues the pipeline creates** and **Never blocking** (`jira-integration.md`) verbatim — no new
labelling rule and no change to the existing "two carve-outs from Never blocking" count in that
file, since that count is scoped to `/myflow-start`'s own guardrail about writes aimed at an issue
the pipeline did not choose, not to every interactive Jira question in the pipeline. Filing a new
issue on explicit operator confirmation is the same shape `/myflow-finish` run 1's follow-up filing
already uses.
**Considered:** one batched ask covering all findings — less granular control per finding.

### Report storage

**ID:** self-review-report-storage
**Status:** active
**Chosen:** `docs/self-review/<name>-self-review.md`, committed and pushed on the base branch.
**Considered:** only in the state file (machine-local, not in git history, and the change is
already terminal by this point); both a committed report and a rating mirrored into the state file
— rejected once report-only was chosen, since duplicating the rating into an already-terminal file
buys nothing.

### Skip prompt semantics

**ID:** self-review-skip-prompt
**Status:** active
**Chosen:** fires after `FINISHED` is written, never blocks; default answer to "run it?" is Yes, so
self-review runs unless the operator explicitly opts out. Silence or a session that cannot ask also
runs it, since nothing external is written until the later per-finding and rating asks.
**Considered:** none — this matched the issue's own phrasing once clarified during brainstorming.

### Rating source

**ID:** self-review-rating-source
**Status:** active
**Chosen:** the operator is asked interactively for the 1-5 rating.
**Considered:** the agent self-assigning a rating as part of its own self-review — rejected as a
weaker signal than the human who watched the run.

### The unread `<state-dir>` parameter

**ID:** self-review-state-dir-parity
**Status:** active
**Chosen:** `gather-self-review-context.sh` keeps a third positional `<state-dir>` argument in its
signature, unused by any of its four sources, for CLI-shape parity with
`preserve-session-records.sh`'s own `<worktree> <name> <state-dir>` shape — every script this
pipeline's finish procedure invokes takes the same three-argument pattern.
**Considered:** dropping the parameter to a two-argument
`<archived-change-path> <name>` signature — rejected because it breaks the established shape and
buys nothing: the caller already has the value in hand from resolving the other two scripts' calls.

## Risks / Trade-offs

- **Scope creep of "actionable finding".** A vague line for what counts as actionable could turn
  into a filing-ask per bullet point. Mitigation: `tasks.md` states the test explicitly (names a
  concrete pipeline/script change, not a bare observation) and the review panel checks it.
- **The new script has no worktree to run against once run 2 has removed it.** Mitigation: it reads
  from the **archived** change path (`openspec/changes/archive/<date>-<name>/`) and the
  `docs/superpowers/` records already committed by run 1 — both exist by the time step 8 runs,
  after step 4 (cleanup) has already removed the worktree.
- **A second commit after the archive commit adds one more push in run 2.** Mitigation: this is the
  same "guarded, skip-if-empty" chain pattern already used elsewhere in this pipeline; a failure
  here is reported and does not un-write `FINISHED`, since `FINISHED` was already written in step 7.

## Migration Plan

Not applicable — this is a purely additive change to `/myflow-finish` run 2 with no data migration
and no schema change. Existing `FINISHED` changes are unaffected; the new step only runs for changes
that reach `FINISHED` after this change lands.

## Open Questions

None — every question raised during brainstorming was answered and is recorded as a decision above.
