# kan-1064-flow-self-review-run-the-coordination-around-the

## Why

The `flow.self-review` stage's main session still carries the heavy content itself. In-run it
fetches the context bundle and holds it in memory (`skills/flow-contracts/finish-contract-run1.md`,
**Run the self-review pass**), then writes it again, in full, into the pass dispatch's prompt
(`skills/flow-self-review/SKILL.md` step 2). The kan-1060 bundle was about 33k tokens:
<!-- measured: wc -c of the kan-1060 in-run bundle, 131,213 bytes / 4 @ machine-local file read during kan-1064 planning round 1, cannot be re-run -->
main reads it once as a tool result, writes it once as output, and cache-reads both copies on
every later turn of the stage — 30 to 50 coordination turns.
<!-- predicted: the turn count of the next three flow.self-review stages after landing -->
The same holds for every other handoff: main writes each prompt-tuning case file itself
(step 2a), does angle 7's `git show` extractions, and quotes every finding, review report and
failing verify output into the plan, fix-r and fix-v prompts. Over the last 30 days the stage cost
$821 across 135 runs, about half of it in the main session.
<!-- measured: flowd stage cost split for flow.self-review over the 30 days to 2026-10-30 @ machine-local flowd store read during kan-1064 planning, cannot be re-run -->

The ticket also asks that the self-review fixers run on opus or sonnet at medium effort, where
**Allowed pairs** (`skills/flow-contracts/model-policy.md`) still has them on `opus/flow-low` with a
`flow-high` no-progress escalation row.

## What changes

- The self-review's main session passes file paths, never file contents. Every handoff is a file
  in `<sr-dir>`, `<state-dir>/self-review/<name>/`, outside every checkout.
- The in-run bundle is written to `<sr-dir>/bundle.md` and never printed; the pass reads it there.
- The pass writes `findings.md`, a `report-draft.md` with disposition placeholders, each
  prompt-tuning case's files and angle 7's extractions, and hands back a short summary.
- Step 2a becomes a copy, a render and a check. The plan, fix-r and fix-v prompts name the
  findings, review and verify files; each reviewer writes its report to `<sr-dir>` and hands back
  only its counts.
- Main fills the draft's placeholders with shas and Jira keys, records the store rows from it and
  lands it as the report. Every `big` / file / land judgement stays main's, and no subagent waits
  on another.
- `self-review-fix` runs on `opus/flow-medium` or `sonnet/flow-medium`; its `opus/flow-high`
  escalation row is deleted, and a no-progress fix is answered by the existing fresh re-plan.
- No review merger is added to the self-review loop.
