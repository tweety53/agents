# Self-review report for kan-1034-spectre-optional-bundled-planning-prompts-that

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

**Rating:** skipped — paused by the operator

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** In-run pipeline fix loop 31: check-task-records read a cross-repo peer's state from the workspace daemon under verify's exported FLOW_ADDR; the state read now follows FLOW_RECORDS_ADDR — fixed: 36eb356c
- **[flow-fix]** A completed slot that left no report file, and a fixed finding whose reproducer was missing, passed every close guard; check-panel-findings-closed now requires both — fixed: bfbe2d75
- **[flow-fix]** The bold `**none**` test-repair form was refused although the planner text names it; the repair pattern now accepts it — fixed: b940ea6f
- **[flow-fix]** A Baseline `measured:` comment without ` @ <ref>` silently disabled the re-measurement; plan-shape now refuses it — fixed: 0cb8e71e
- **[flow-fix]** The relocation comparison refused the whole change over one symlinked scoped path; it now skips the symlink with a note — fixed: 3c3a5c55
- **[flow-fix]** spectre declared no `## lint` or `## test`, so its verify ran by hand and integrate stopped once; spectre's project file now declares both — fixed: a23f18c
- **[flow-fix]** Planning classed flow-dependent paragraphs as generic for a destination another repository vendors; the planner now checks a relocation's citations against what its destination ships — fixed: 644df1f9
- **[flow-fix]** The plan had a committed spectre file cite a change-branch sha that integrate's reshape orphans; the planner now forbids such citations — fixed: 644df1f9

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Explorer rows recorded begin and end together after the agents returned, so they read 0 s; brainstorm.md now records the begin before the Agent call — fixed: ff9707a8
- **[flow-cost]** Task 1's implementer and the panel fix round dominate the spend (6.18M and 5.71M cache read); the fix round traces to the planning classification finding above — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The operator's mid-panel scope choice was recorded as a decision with its rejected alternative, and every Files widening carries a dated Correction — noted
- **[flow-improvement]** The mutation slot caught a 0-byte fixture the primary slot passed, and round 1 re-ran the mutant to confirm the kill — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** A missing final newline and doubled blank lines in skill Markdown were unguarded (F8, F10); check-markdown-integrity now flags both — fixed: 9486eae3
- **[flow-automation]** Nothing told an editor of `skills/spectre-planning/SKILL.md` to re-copy it into spectre; the skill-index row now names the duty — fixed: 35fb7e73

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** The slot that left no report, the spectre stop at prepare and the manual relocation audit each trace to an angle-1 fix above — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review candidate: F7, Important, `stats/internal/guard/loadsets.go:74` — not mined (mining paused by the operator) — noted
- **[flow-eval]** impl candidate: task 2, `TestPluginShipsSpectrePlanning` in spectre — not mined (mining paused by the operator) — noted

## Instruction gaps that cost review-panel fix rounds — `flow-prompt-tuning`

_none — this angle produced no findings._
