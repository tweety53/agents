# Self-review report for kan-1062-flow-implement-one-shot-subagents

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** check-panel-fix-single-dispatch refused a chunk row whose fix commit carries the round's Flow-Key (in-run pipeline fix) — fixed: 52f3e77d
- **[flow-fix]** the REPRODUCER paragraph named reproducers `R-ID-N.sh`, so slots wrote off-shape `R-0-mutation-N.sh` names (kan-1047, kan-1048, kan-1062) — fixed: 8eb56e6b
- **[flow-fix]** a whole-package `go test ./cmd/flow` left a detached `sleep` alive (stage wrap killed only its `sh`), so run-reproducer exited 3 — fixed: 28d9c60b
- **[flow-fix]** the self-review bundle omitted the gated per-task reviewers' findings — fixed: 6fd1e12e
- **[flow-fix]** a scratch change on the live-store copy needed a hand-made change directory and STARTED record before `flow record` accepted it — fixed: b9c7fdd7
- **[flow-fix]** the plan used a bare `-reproducer none`, which the validator refuses; the run used `none — <reason>` — noted
- **[flow-fix]** the installed CLI predated Task 3, so ref-less findings were numbered by hand; the post-landing reinstall closes it — noted
- **[flow-fix]** the failure-modes slot gave Important F18 no reproducer; the rule exists, slot variance; the parent wrote one — noted
- **[flow-fix]** Task 6's gated contract defects (Model: line, early stage-end trigger, exit-3 step) were fixed inside the change — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** planning cost 51M cache reads (plan-2 at xhigh, 36 min); the two-patch contract design paid off with zero guard failures — noted
- **[flow-cost]** the parent read implement.md, review-panel.md, the fix-round file, verify-and-handoff.md and the self-review skill in full — the cost this change itself targets — noted
- **[flow-cost]** about 4.4 min of parent time went on reproducer repairs, removed by the 8eb56e6b and 28d9c60b fixes — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** measured plan premises caught Task 2's unreachable refusal; the two-patch contracts landed with zero verbatim or reference failures; parallel waves ran six tasks in 26 min; the in-run pipeline fix landed without stopping the run; all four live-check steps matched — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** Nothing beyond the reproducer and live-store gaps, fixed under flow-fix — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Nothing to move: the gated findings now reach the bundle from the store, fixed under flow-fix — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** the mutation+failure-modes slot began 5m45s after the reading slots because review-panel.md holds it until the full suite exits (kan-1039) — by design — noted
- **[flow-speed]** the 3h53m gap before implementation was the operator's plan gate — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case F18 (Important) — not mined — noted
  - location: stats/cmd/flow/record.go:2054 at f219c53c; fixed in 7d3de853 (`replaceFile`)
  - sentence: "partial write: WriteFile truncates then writes in place, so a kill mid-write leaves a truncated findings file with no copy"
- **[flow-eval]** review case F7 (Important) — not mined — noted
  - location: stats/internal/store/aggregate.go:261; fixed in ef85dd69
  - sentence: "full-suite fix rounds silently drop out of FixRounds"
- **[flow-eval]** impl case Task 3 (`numberRefLessFindings`) — not mined — noted
  - code: stats/cmd/flow/record.go:2025

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** A rewrite of a file whose old bytes are the only copy truncates it in place, so an interrupted write loses the data — implementer — docs/prompt-tuning/kan-1062-flow-implement-one-shot-subagents-1/ — noted
