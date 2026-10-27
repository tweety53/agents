# Self-review report for kan-1040-copyable-next-command

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits, store rows, transcripts and logs it names, and nothing beyond them. The operator waived the rating and the flow-eval mining asks.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `flow-postgres` had restart policy `no`, so quitting Docker Desktop left the flow store down until a manual `docker start` — fixed: df9afca1
- **[flow-fix]** A reviewer reproducer run in the canonical worktree reverted an uncommitted edit; `run-reproducer` only detected it afterwards — fixed: e25ef772
- **[flow-fix]** `panel-fix-<round>` was ambiguous and `panel-fix-0` was admitted, so an inline back-fill took no reviewer floor; three rows rekeyed and one retimed — fixed: b03e5a14
- **[flow-fix]** A journalled `dispatch end` froze the caller's clock and skipped the transcript end correction; three panel-0 rows retimed — fixed: 8c97d640
- **[flow-fix]** The 18:07Z store stall left no evidence: the live flowd log was undocumented and swollen by per-tick stuck-journal warnings, and postgres logged no lock waits; lock-holder hypothesis disproved — fixed: 07d4b6cb
- **[flow-fix]** Re-running `panel-round-begins.sh` after a journalled begin duplicated a pass-log note; duplicate row 2504 dropped — fixed: 60a30324
- **[flow-fix]** The inline verify row was harvested as 0 tokens: two verify groups shared key `verify` and the settle pass zeroed the row against the lint run's end; session-mismatch hypothesis disproved; row 8235 restamped — fixed: 948f2a75
- **[flow-fix]** The `internal/store` suite hit the 10-minute timeout during the outage; storetest rejects fixed deadlines by design and the package passed alone in 94.7 s — noted
- **[flow-fix]** The store refused the late inline `panel-fix-0` begin, correctly — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The `panel-fix-0` row's 5.76M cache-read tokens were a mis-stamped window, repaired with the panel-fix rekey above — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The mutation slot caught unguarded template fences on a docs-only change; signal 7 landed with tests and round 1 tightened it — noted
- **[flow-improvement]** The docs-only reduction kept round 0 to primary plus the extras, and round 2 re-ran mutation alone — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** `flow suite record` did not journal, so the outage lost suite records and prompted a 477 s guard-suite re-run — fixed: ebb95e80

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F2 Important, the outer template fence reverted to three backticks survived every guard (`skills/flow/integrate.md:225`); operator waived mining — noted
  - Commit: bad `907f6125`, fixed `e3e1e20b`
- **[flow-eval]** review case — F5 Important, dropping signal 7's opening-fence clause survived every test (`stats/internal/guard/markdownintegrity.go:681`); operator waived mining — noted
  - Commit: bad `e3e1e20b`, fixed `27629e97`

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

_none — this angle produced no findings._
