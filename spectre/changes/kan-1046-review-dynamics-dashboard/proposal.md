# kan-1046-review-dynamics-dashboard

## Why

Nothing on the stats dashboard shows whether the review panel is getting better or worse over time. The Reviewers view totals findings per slot for one period, with no time axis, no model or effort split, no fix rounds, and no link to the prompt changes meant to move those numbers. The operator cannot tell from the dashboard whether to keep the panel process as it is or adjust it (KAN-1046). Prompt tunings and prompt edits are not in the stats store at all.

## What changes

- A new **Review dynamics** view (`#/review-dynamics`) charts trends per day (ranges of 2 weeks or less) or per week:
  - **Review panel:** findings per change by severity, findings per dispatch by slot, by reviewer model and by reviewer effort, fix rounds per change (mean and max), and changes reviewed.
  - **Implementer:** panel findings per implementer task by implementer model · effort, and the share of implementer dispatches that were blocked or aborted.
  - **Visual verify:** verify rounds per change (mean and max), and the share of changes that passed on the first run.
- Verdict tiles compare the range against the previous range of the same length and colour each change as better or worse.
- A heatmap shows findings per dispatch for each slot × model · effort, with the change against the previous range and the counts on hover.
- Every chart has a hover crosshair whose tooltip lists each series' value, what it is based on, and the prompt change at that point. Dashed markers show prompt changes for the chart's own area.
- The view has its own range presets: Last week, Last 2 weeks, Last 4 / 8 / 12 weeks, Since last prompt change, All time.
- The store gains a `prompt_changes` table:
  - `flow prompt-tuning record` writes one row when `/flow-prompt-tuning` applies a case.
  - `flow prompt-tuning import` loads the git history of every tunable prompt (reviewer, implementer, verifier, panel fixer) as prompt-edit rows.
- Severity spellings fold to Critical / Important / Minor in one SQL function.
- Model versions are never merged silently: every model chart, the implementer chart and the heatmap carry a **Family / Version** toggle — `model · effort` (versions merged) or `model-version · effort` (each version its own series).
  - The harvester stamps each dispatch's exact model ID (`claude-opus-5-5`) from its transcript's `message.model`; today ~7,300 of ~8,300 rows hold only the alias `opus`/`sonnet`.
  - `flowd backfill-models` re-reads the transcripts still on disk and stamps past dispatches; a row whose transcript is gone reads `<family> (version unrecorded)`.

**In-run fix 1 (flow.visual-verify, 2026-10-28):** the view now follows the approved prototype:
- The intro is 12px.
- The heatmap is packed into 92px cells with no row rules. Its rows run busiest slot first and its efforts run low → high, in both the heatmap and the effort chart.
- The tooltip sits on its own `--surface-tip` surface.
- The Version toggle, the default preset and the heatmap cell hover are captured.

The light palette's departure from the prototype predates this change (the app's theme tokens), so it is left as is.
