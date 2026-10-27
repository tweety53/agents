# kan-1038-stats-weekly-costs-one-chart — session narrative

## 2026-10-27 — creating run

- The base was red before any task: `scripts/test-prompt-tuning-line.sh` piped into `grep -q`, which
  `TestNoPipeIntoEarlyExitGrep` refuses; fixed in-run with a here-string.
- Verify found a crash the panel missed: a series with no measured week threw on
  `line.find(...)!.series`. In-run fix 1 returns no line for it.
- The first full visual verify found a stale full-app-suite baseline (the pre-change two-chart view)
  and two spec gaps — no pointer park before the toggled capture, no capture driving the hover
  tooltip. In-run fix 2 added both.
- The operator asked mid-run whether the chart would match the prototype exactly. The answer listed
  the deliberate departures (theme colours, ISO date labels, 1/2/2.5/5 y steps, no table, app panel
  chrome); verify-2 then read three tooltip departures (copy, width, shadow), closed by in-run fix 3.
  Verify-3 read one more — tooltip corner radius 3px against the prototype's 6px — left open for the
  operator, since panel chrome was named as coming from app styles.
- A mutation reproducer's restore trap used a relative path after `cd stats/web` and left its
  mutant in the tree; repaired to restore from an absolute backup. Lesson: a reproducer that changes
  directory restores by absolute path.
- Prototype width 150px still wrapped the tooltip copy under the app's padding; the prototype's own
  `6px 10px` content-box padding fixed it, which is why task 7 took both values.
- `prepare-workspace.sh` exports `FLOW_ADDR` for the worktree's own daemon; exporting it into a shell
  that then runs stage marks sends them to a daemon that is not up, and they fall to the journal.
  Stage marks run without that export.
- One `npm test` run hit a 5s timeout under load; its one allowed re-run passed.
