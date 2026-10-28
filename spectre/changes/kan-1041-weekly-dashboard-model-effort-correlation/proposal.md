# kan-1041-weekly-dashboard-model-effort-correlation

## Why

The Weekly costs view shows where spend goes by pipeline phase, but not why a phase spikes. The model and effort mix behind each week is the likeliest cause, and nothing on the dashboard shows it. A dot's cost also hides behind a hover, so reading a week means hovering every point. And the store has never priced a single subagent dispatch: dispatches record the alias `opus`/`sonnet`, which no pricing row matches, so per-dispatch cost — the only place effort lives — is empty.

## What changes

- The phase chart gains a **Visual verify** line (`flow.visual-verify`) and an **Other** line: every cost outside the named phases, so the phase lines add up to Overall.
- A second chart, **Cost by model · effort**, sits under it on the same 8-week axis: one toggleable line per model·effort pair seen in the window — each dispatch's served model and effort — plus one `<model> · session` line per model for the parent session's own cost, whose effort is not recorded. The lines add up to Overall.
- Every dot on both charts permanently shows its USD cost beside its week-over-week change; the first week shows the cost alone.
- Dispatch pricing resolves its rate from the model the transcript says served the dispatch, so dispatches get a `cost_usd`.
- `flow pricing reprice -since <YYYY-MM-DD>` re-runs pricing over every stage run started since that date, so the existing 8 weeks fill in.
