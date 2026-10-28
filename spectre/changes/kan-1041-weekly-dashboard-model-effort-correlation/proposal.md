# kan-1041-weekly-dashboard-model-effort-correlation

## Why

The Weekly costs view shows where spend goes by pipeline phase, but not why a phase spikes. The model and effort mix behind each week is the likeliest cause, and nothing on the dashboard shows it. A dot's cost also hides behind a hover, so reading a week means hovering every point. And the store has never priced a single subagent dispatch: dispatches record the alias `opus`/`sonnet`, which no pricing row matches, so per-dispatch cost — the only place effort lives — is empty.

## What changes

- The phase chart gains a **Visual verify** line (`flow.visual-verify`) and an **Other** line: every cost outside the named phases, so the phase lines add up to Overall.
- A second chart, **Cost by model · effort**, sits under it on the same 8-week axis: one toggleable line per model·effort pair seen in the window — each dispatch's served model and effort — plus one `<model> · session` line per model for the parent session's own cost, whose effort is not recorded. The lines add up to Overall.
- Every dot on both charts permanently shows its USD cost beside its week-over-week change; the first week shows the cost alone.
- Dispatch pricing resolves its rate from the model the transcript says served the dispatch, so dispatches get a `cost_usd`.
- `flow pricing reprice -since <YYYY-MM-DD>` re-runs pricing over every stage run started since that date, so the existing 8 weeks fill in.

## In-run fix 1 — flow.verify

- The live check's reprice failed with a store error: `GetStageRun` cannot read an unattached plan session (NULL `change_id`). It reads it as change 0, as `StageRun` documents, and the model·effort series counts only runs attached to a change, as Overall does.


## In-run fix 2 — flow.verify

- The live check's per-week model·effort sums exceeded Overall (2026-09-30: $1551.51 against $939.78). Runs priced before a model with no rate in effect (`claude-sonnet-5-5`) joined them kept their earlier top-level `cost_usd`, now short of their buckets. Price drops that total whenever a bucket cannot be priced, and the model·effort series counts only runs carrying a total, as Overall does.

## In-run fix 3 — flow.visual-verify

- The visual verifier found the header copy and the pill width departing from the approved prototype. The header reads `Wednesday-to-Wednesday cost, last 8 weeks.`, and each pill is the prototype's 40 (cost alone) or 84 (with a delta) wide, widened only when its text is longer.

## In-run fix 4 — flow.verify

- The full guard suite failed under `-race` (`TestWorkspace` and the tests running beside it): `runGuard`'s merged buffer was written by both of a child process's stream-copy goroutines. It now takes a lock.
