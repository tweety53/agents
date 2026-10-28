# Self-review report for kan-1041-weekly-dashboard-model-effort-correlation

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits, files and store rows it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `protect-main-checkout.py` denied a subagent's two-heredoc report write: each heredoc's "rest" still held the other's body, so a stray backtick or a lone `.` failed closed. The sonnet mutation slot's reports were "lost" this way, not in the throwaway copy. The rest now drops the other bodies unless one names the target being judged — fixed: 54ffcd41
- **[flow-fix]** Panel-fix rows were keyed one round low (`panel-fix-5`/`-7` for findings rounds 5/7 raised), which defeated the start floor. `flow record dispatch` now refuses the round-low key, rows 8334/8344 were re-keyed to `panel-fix-6`/`-8`, and the archived ledger and panel record were re-rendered — fixed: 2faab919
- **[flow-fix]** An inline back-fill window (`task-13-implementer`) contained another inline row (`verify-fix-3`), leaving task 13 unmeasured. The derived start now floors at the change's other inline rows, and row 8340 was re-timed — fixed: ea7203d6
- **[flow-fix]** `review-panel.md` advised `PIPESTATUS`, which zsh leaves empty, so an empty status read as a pass — fixed: f8daa9eb
- **[flow-fix]** `document-fix.md` named the plan guards without their invocation, and implied the provenance guard ships among the skill's scripts — fixed: 314e9e5c
- **[flow-fix]** `measure-visual-properties --props matrix` paired 0 elements on every dark-theme image: the surface top landed at the image height — fixed: 26a36d40
- **[flow-fix]** The visual-verifier dispatch was denied twice because its description did not open with its row key; the hook works as designed — noted
- **[flow-fix]** `close-task.sh` did not export `FLOW_GUARD_SCRIPT_DIR` for the zip rebuild; already fixed on main by 33519490 — noted
- **[flow-fix]** `panel-4-mutation+failure-modes` ended at its re-written report's mtime, 6 s off the daemon's own close — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Four late-fix rounds each ran the full three-slot re-review over small diffs; designed behaviour, and round 5's F17 shows its value — noted
- **[flow-cost]** The largest spends were long inline edit/test loops on an opus parent (panel-fix-1, task 9, panel-fix-5), as the decision chose — noted
- **[flow-cost]** One citation send-back on round 1's primary, 19 s, guard-driven as intended — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The live check on a `pg_dump` copy behind a scratch flowd found two defects every test missed, without touching the dev stack — noted
- **[flow-improvement]** A race that did not reproduce under load was proven with a forced reproduction instead of being closed as a flake — noted
- **[flow-improvement]** The visual verifier separated this change's prototype departures from merge-base ones, each with its baseline cited — noted
- **[flow-improvement]** The parent's `test -s` after fold-back caught a slot's false "reports written" handback — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** The live check was hand-built over about 15 calls; `scripts/live-store-copy.sh up|down` now copies the dev store into a scratch database behind a scratch flowd and tears both down, never touching flowd on 4173 or the `flow` database — fixed: 437c849b

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Inline `verify` rows record zero or `{}` tokens; by the harvester's contract a window around one blocking command holds no parent turn — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** Round 0's mutation slot waits for the full suite and then ran 18m40s; a deliberate trade (kan-1039) — noted
- **[flow-speed]** Each in-run fix re-ran the whole verify list; those re-runs caught the `-race` failure — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1, Important, a cross-layer timeout (2 s CLI default, 30 s write deadline) on a long reprice; not offered for mining, per the operator's standing instruction — noted
  - severity: Important; location: `stats/cmd/flow/pricing.go:53`; slot primary, round 0
  - bad code: `git show 0ac44839:stats/cmd/flow/pricing.go`; fixed in b4090ed2
- **[flow-eval]** impl case — Task 1, price a dispatch by its served model with a deterministic tie-break; not offered for mining, per the operator's standing instruction — noted
  - code: `stats/internal/store/pricing.go:275` (`dispatchBucket.servedModel`)

## Instruction gaps that cost fix rounds — `flow-prompt-tuning`

_none — this angle produced no findings._
