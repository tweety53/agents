# Design: self-review after `/myflow-finish` run 2

**Change:** `kan-23-myflow-self-review`
**Jira:** KAN-23 — "Myflow self-review"
**Date:** 2026-08-03

## Why

KAN-23 asks for a retrospective step once a change finishes: what problems came up and how the
pipeline could avoid them next time, what could reduce token/time cost without quality loss, what
went well and how to reproduce it, and what could be scripted instead of reasoned through. Findings
should be offerable as Jira issues, and the operator should be able to rate the run.

## What

A new step folded into `/myflow-finish` run 2, immediately after the `FINISHED` write and before
the terminal handoff prints. No new pipeline state, no new command — self-review is purely
additive to run 2's existing procedure and never affects whether `FINISHED` gets written.

### 1. Trigger and placement

Runs only when run 2 reaches `FINISHED` normally (i.e. `check-cleanup-complete.sh` returned
`COMPLETE:` and the state write succeeded). A run that stops at a cleanup leftover never reaches
this step — self-review is never a reason `FINISHED` is delayed, and it never re-opens a finished
change.

### 2. Skip prompt

On entering the step:

> **Run self-review for this change?**
> - **Yes — run it** *(default, recommended)*
> - **No — skip**

Only an explicit **No** skips (one line: `Self-review: skipped`, and the existing run-2 handoff
prints unchanged). Silence, a stalled prompt, or a session that cannot ask **runs it** — the safe
default here is to run, not to skip, since nothing external is written until the later gated asks
in step 5.

### 3. Context gathering — `scripts/gather-self-review-context.sh`

```
scripts/gather-self-review-context.sh <archived-change-path> <name> <state-dir>
```

Deterministic, no LLM involvement, following the pattern already used by
`check-finish-preflight.sh` and `preserve-session-records.sh`. Collects and prints one formatted
bundle to stdout:

- the SDD ledger, `docs/superpowers/ledgers/<name>.md` (dispatch count, model per dispatch, fix
  rounds)
- the review-panel record, `docs/superpowers/reviews/<name>-panel.md` (slots run, findings count,
  re-run count)
- `tasks.md` from the archived change (planned vs. actual task count)
- `git log --stat` for the change's two finish-run-1 commits plus the archive commit

Each source that does not exist is reported `skipped: <src> (absent)` on stdout and gathering
continues — a change may legitimately have no panel record, exactly as
`preserve-session-records.sh` already treats a missing source. The script exits 2 on an invocation
error (a missing argument or an invalid change name); otherwise it always exits 0 — a missing source
is never a pass/fail determination. It makes no judgement, only assembles input.

### 4. The reasoning pass

One combined pass — not four separate dispatches — fed the gathered bundle plus the live session's
own context (what the agent that ran `/myflow-do` and `/myflow-finish` directly experienced),
answering all four angles from KAN-23 in one report:

1. problems encountered, and what pipeline change would avoid them
2. token/time cost, and what would reduce it without quality loss
3. what went well, and how to reproduce it
4. what could be automated or moved to a script

One pass rather than four keeps this consistent with what angle #2 itself is checking for — a
self-review mechanism that burns disproportionate tokens auditing cost efficiency would be its own
finding.

### 5. Per-finding Jira ask

For each **actionable** item surfaced above (one that names a concrete change to make, not a bare
observation), ask once:

> **File `<one-line finding>` as a Jira issue?**
> - **No — don't file** *(default, recommended)*
> - **Yes — file it**

Only an explicit **Yes** files. A filed issue carries every label on the change's linked issue plus
`AI-generated`, and links to that issue when one exists — reusing
**Labels on issues the pipeline creates** (`skills/myflow-contracts/jira-integration.md`) verbatim,
no new labelling rule. A filing failure degrades to the standard
`⚠ Jira: skipped — <reason>` line and self-review continues — **Never blocking**
(`jira-integration.md`) applies here exactly as it does to every other Jira write in this pipeline.

### 6. Operator rating

After the report and the filing asks:

> **Rate this myflow run, 1 (rough) to 5 (excellent):**

Recorded only in the report file (step 7) — explicitly **not** written to the state file, since the
state file is machine-local metadata and the change is already terminal by this point.

### 7. Report and commit

Write `docs/self-review/<name>-self-review.md` — the four-angle report, the rating, and which
findings were filed vs. declined. Commit it on the base branch in the main checkout, alongside (or
immediately after) the archive commit already made earlier in run 2, and push.

### 8. Handoff

The existing `## Finished` block gains one line, nothing else changes:

```
**Self-review:** docs/self-review/<name>-self-review.md (rating: <n>/5) | skipped
```

### 9. No schema change

No new pipeline state, no new command, no new `state.json` field. This is entirely inside
`/myflow-finish` run 2's existing procedure.

## Decisions

### Where self-review sits in the pipeline

**ID:** self-review-placement
**Status:** active
**Chosen:** inside `/myflow-finish` run 2, after the `FINISHED` write — never blocks or delays the
archive
**Considered:** before the `FINISHED` write (risks a self-review failure blocking the terminal
state); a new standalone `/myflow-self-review` command (collides with this repo's explicit
three-pipeline-command invariant and would require rewriting the command-surface table and every
cross-referencing doc for a feature that needs no new pipeline state)

### Review structure — one pass vs. four

**ID:** self-review-pass-count
**Status:** active
**Chosen:** one combined reasoning pass covering all four angles plus the rating
**Considered:** four distinct passes/dispatches, one per angle, mirroring the review panel's
separate-subagent-per-lens model — rejected because it directly works against angle #2 (reducing
token/time cost), which a self-review mechanism should not itself violate

### Context gathering — script vs. inline reading

**ID:** self-review-context-gathering
**Status:** active
**Chosen:** a new deterministic script, `scripts/gather-self-review-context.sh`, collects the
ledger/panel record/tasks.md/git log mechanically; the reasoning pass only judges the bundle
**Considered:** the skill's own prose instructing the agent to read the files itself — rejected
because every self-review run would then re-read and re-summarize the same files from scratch
in-context, working against angle #2 for the same reason four passes would

### What the self-review analyzes

**ID:** self-review-inputs
**Status:** active
**Chosen:** the preserved artifacts (ledger, panel record, OpenSpec artifacts, archived git
history) plus the live session's own context
**Considered:** artifacts only, for reproducibility if self-review were ever re-run standalone
later — rejected in favor of the richer signal live session context gives for "what went well"

### Task creation for findings

**ID:** self-review-task-creation
**Status:** active
**Chosen:** Jira issues, one ask per actionable finding, default No, reusing the existing labelling
and never-blocking rules
**Considered:** one batched ask covering all findings (less granular control per finding)

### Report storage

**ID:** self-review-report-storage
**Status:** active
**Chosen:** `docs/self-review/<name>-self-review.md`, committed
**Considered:** only in the state file (machine-local, not in git history — arguably not what the
state file is for); both a committed report and a rating mirrored into the state file (rejected
once the report-only option was chosen — no reason to duplicate the rating into a file that is
already terminal at this point)

### Skip prompt semantics

**ID:** self-review-skip-prompt
**Status:** active
**Chosen:** fires after `FINISHED` is written, never blocks; default answer to "skip?" is No, so
self-review runs unless the operator explicitly opts out
**Considered:** none — this matched the issue's own phrasing once clarified

### Rating source

**ID:** self-review-rating-source
**Status:** active
**Chosen:** the operator is asked interactively for the 1-5 rating
**Considered:** the agent self-assigning a rating as part of its own self-review — rejected as a
weaker signal than the human who watched the run

## Open questions

None — every question raised during brainstorming was answered and recorded as a decision above.
