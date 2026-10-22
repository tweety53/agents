# Self-review report for kan-923-flow-improvement-port-the-python-backed-guards

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits, files and logs it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** Three new shims hand-rolled `resolve-file.sh` + `SELF_REAL` root resolution that `flow_guard_root()` already provides through skill symlinks — fixed: f161defe
- **[flow-fix]** `check-plan-unchanged verify` false-fired on the parent's own `flow tasks tick` made while a reviewer ran; it now snapshots plan content and passes checkbox-only flips — fixed: f477ba07
- **[flow-fix]** `tcfBodyFields` was last-wins, so a Correction line opening with `**Files:**)` overwrote task 5's real Files field; the first occurrence now wins — fixed: 1e8705a6
- **[flow-fix]** Tracked `skills/*/scripts/<name>.py` symlinks were not foreseen by the plan; `check-guard-symlinks.sh` rule 1 caught each at the first commit — noted
- **[flow-fix]** The design.md premise on `check-task-records.sh`'s root was disproved: its Context already stated `${1:-$PWD}`, and the task-6 hedge was corrected — noted
- **[flow-fix]** The live check ran the worktree's `skills/flow/scripts/` symlinks because `~/.claude/skills/flow` points at the main checkout; same path shape, byte-identical parity — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The panel diff carried about 10.8k lines of deleted-file bodies that every pass-1 slot read; `write-panel-diff` now passes `--irreversible-delete` — fixed: caafd6ca
- **[flow-cost]** Implementer cache reads of 9.3M and 11.6M on the parity-heavy ports, with no redundant re-dispatch — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Tasks 1 and 7 ran in parallel and reviewers overlapped later implementers, leaving no idle gaps beyond the `**After:**` chain — noted
- **[flow-improvement]** Parity evidence: a byte `cmp` table for all 10 guard/input pairs and an old-vs-new sweep over 416 tracked `tasks.md` with 0 differences — noted
- **[flow-improvement]** Panel Minor F1 overrode byte parity for a deleted-file label, recorded in a dated Correction with its reason — noted
- **[flow-improvement]** No stale `.py` citation was left on the branch — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** "output not measured" on subagent dispatch rows is the deliberate kan-920 state (transcripts carry `stop_reason: null`) — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** A race in `views.test.tsx` (the region renders while loading) forced a full `npm test` re-run in verify; the five sites now await loaded content — fixed: b4071157

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case: task 4, Port plan-dispatch-groups (the operator chose not to mine it) — noted
  - Task: tasks.md task 4, "Port plan-dispatch-groups"
  - Code: `stats/internal/guard/plandispatchgroups.go:83` (`pdgComputeGroups`), commit 66a10378
  - Why: small, self-contained, a byte-exact oracle and a 14-case parity floor
