# kan-1130-confidence-estimate
> Design record of change kan-1130-confidence-estimate (KAN-1130), written at landing. Each decision's own **Status:** line is the status.

## Context

The step the operator asked for:

> Estimate how confident you are in your ability to execute the plan (1-100%) and what it would
> take to move you to 100%.

Models report 85–95% on almost anything, so the number alone separates few plans from each
other; the half that carries the value is the list of what stands between the plan and 100%.
The first design (an ideas note, since deleted) had the plan check estimate once and its gaps
triaged like any other finding. The operator revised it on five points, which this record
follows:

1. Every response below 100% is kept in full in a table of its own in the flowd store, with a
   dashboard view laid out as the Self-review view is.
2. In planning, a response below 100% opens a dialog with the suggestions that would raise it,
   and the operator either has them applied or answers "later", which only files the response.
3. Every implementer estimates at the start of its dispatch, files the result and continues
   without stopping.
4. Tags are dynamic: the filing step lists the tags in the table, reuses one that fits and
   creates one otherwise.
5. The operator analyses the recorded data later and asks for changes then.

## Decisions

### The plan's estimate is a lens of the plan check

**ID:** confidence-is-a-plan-check-lens
**Status:** active
**Chosen:** `## Execution confidence` is the last lens of `skills/flow/plan-review-rubric.md`,
run by the `decider` in the one `flow-guard plan-check` call it already makes, full mode only.
Its section opens `CONFIDENCE: <n>%`, cites earlier findings on a `Lowered by:` line and raises
each new gap as a finding line whose text opens `closes with: plan|operator|running it —`.
**Considered:** a separate headless call, `flow-guard plan-confidence` (re-reads what the decider
read seconds earlier, and a `-decide-only`-shaped call has no tools); the final planner estimating
its own plan (the author's confidence in what it just wrote is the least informative reading);
a threshold below which the run stops (nothing to do below it that is not done above it, and it
invites reporting one point over).

### The guard derives the record from the report

**ID:** confidence-json-derived
**Status:** active
**Chosen:** `flow-guard plan-check` writes `<worktree>/.superpowers/sdd/confidence.json`,
`{"score": <n>, "gaps": [<line>, …]}`: the score of the `CONFIDENCE:` line under the report's
`### Execution confidence` heading, 1–100, and every non-blank line after it up to the next
heading or the report's last line, verbatim — the `Lowered by:` line included. With no
well-formed line it removes any earlier file. `-decide-only` neither writes nor removes it. The
stdout line gains ` · confidence <n>%` or ` · confidence unknown`.
**Considered:** a `confidence` object inside `decision.json` (`-decide-only` rewrites that file
and writes no report, so the estimate would be lost on a gate **No**, and the model would write
each gap twice); parsing each gap line into id, severity, location and closing kind (nothing
reads those fields: the planner reads the line, and the store keeps it verbatim).

### One table, every response below 100%

**ID:** confidence-estimates-table
**Status:** active
**Chosen:** `confidence_estimates` (migration 0097): project, change, `stage` (`plan` or
`implement`), `task` (an implementer's task ids, implement rows only), `score` 1–99, `gaps` and
`tags` as text arrays, `disposition` (`raised` or `later`, plan rows only), `recorded_at`, and
— migration 0098 — `review_status` (`to-review` on insert, then `approved` or `declined`) with an
optional `review_reason`, both set from the dashboard and never by `flow confidence record`. No
uniqueness: a re-dispatched implementer's second estimate is a second response. A 100% response
is not stored — `flow confidence record -score 100` exits 0 and writes nothing, so an implementer
runs the same command whatever its number.
**Considered:** rows in `self_review_findings` under a new angle (the operator asked for a
different table, and that one's angle, severity, effort and theme columns mean nothing here);
storing 100% responses too (the operator asked for those below 100%).

### Planning asks once: raise it, or later

**ID:** confidence-question
**Status:** active
**Chosen:** when `confidence.json`'s score is below 100, the planner triaging the report first
returns `QUESTIONS` with one question listing every gap beside the suggestion that closes it,
and two options, **Raise it** and **Later** *(recommended)*. On **Raise it** the next planner closes each gap by
its kind and files the estimate with `-disposition raised`; on **Later** every gap goes to
`## Risks` as accepted and the estimate is filed with `-disposition later`. Under
`## decisions: recommended` Auto-resolution takes **Later** and the question is not asked; off
that mode it is asked, the one exception to **Now or later is never asked**. The filed score is the cold reader's and is not
re-estimated after the gaps are closed, because the plan is never checked a second time.
**Considered:** a recommended **Raise it** (the recommended mode would auto-take it and every
plan would stop to close gaps unasked); no recommended option, so that the recommended mode asks
it too (the first build; the operator had **Later** auto-picked there instead); asking each `closes with: operator` gap
in the same round (a "later" answer would have made the operator read questions it then
defers); re-running the plan check to measure the raised score (a second ten-minute read per
plan, for a number nothing acts on).

### An implementer estimates first and never stops for it

**ID:** implementer-estimates-first
**Status:** active
**Chosen:** **CONFIDENCE FIRST** in `skills/flow/implementer-brief.md`: after reading its tasks
and before its first edit an implementer estimates, lists what would move it to 100%, and files
it with `-stage implement -task <its Key's ids>`. It then continues whatever the number — no
question, no `BLOCKED`; **REPORT, DON'T DECIDE** is unchanged. A store that cannot take the row
warns and exits 0. A fixer skips it.
**Considered:** the parent filing from the implementer's report (the estimate would be written
after the work, when it is no longer a prediction); a `BLOCKED` below a threshold (the operator
ruled it out: the implementer continues).

### Tags are whatever is already in the table

**ID:** confidence-tags-dynamic
**Status:** active
**Chosen:** `flow confidence tags` prints every tag in use with its row count, across all
projects; the filing step reuses a tag that fits and writes a new lowercase kebab-case one
otherwise. The store checks the shape and nothing else — no vocabulary table, no configuration.
**Considered:** a fixed vocabulary (the operator asked for none until the data shows which tags
earn their place); per-project tags (the same kind of gap recurs across projects, and one
vocabulary is what makes rows comparable).

### The dashboard view is the Self-review view's layout

**ID:** confidence-view-mirrors-self-review
**Status:** active
**Chosen:** `stats/web/src/views/Confidence.tsx` reuses the Self-review view's frame, chip bar,
paged list (25), selected-row pane and `sr-*` classes as they are: the **To review** (selected
on open), **Approved** and **Declined** chips with their counts, the selected estimate's gaps in
full in the pane, **Approve** and **Decline** with an optional reason on a row to review, and the
reason beside **Move back to review** on a decided one. A row's tags show on it and in the pane.
It is served by `GET /api/v1/stats/confidence-estimates?status[&limit][&offset]` and
`PATCH /api/v1/confidence/{project}/estimates/{id}/review`.
**Considered:** a table with a group header per tag (a new layout, which the operator ruled
out); a chip per tag filtering the list (the first build; the operator had the status chips take
their place).

## Not covered

`/flow-fast`, `/flow-interactive`, a fix run's re-plan and the self-review planner file no
estimate: none of them runs the full plan check. Nothing reads a stored estimate yet — no
threshold, no gate — by the operator's fifth point.
