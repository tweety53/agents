## Context

The six findings of `proposal.md` ride kan-920's own branch,
`spectre/kan-920-flow-improvement-port-the-markdown-and-lint`, in its worktree, as one
self-review change. Every fix was drafted, applied and checked in that worktree on a detached HEAD
at `6a023063`, one commit per finding in task order, then the worktree was returned to the branch;
`tasks.md` carries each as the diff to apply. Every planning-time figure below was read from the
dev store (`docker exec flow-postgres psql`, read-only, or inside a rolled-back transaction) or
from the creating run's transcripts on 2026-10-21; neither can be re-run at a ref.

**A1's numbering** (A1 dropped — see decision `unrecorded-fix-round`). Across the store, the `(change, round)` pairs holding a `fixed` finding split
between a `panel-fix-<round>` row (195), only a `panel-fix-<round+1>` row (186), and neither
(268): both numberings are in use (kan-952's `panel-fix-0` fixed round 0; the pass log counts fix
rounds from 1), so a round is covered by a row of either. The 268 are why the check reads only
findings this run's own dispatches raised — a finding carries `dispatchSeq`, and the raising row
its session token. kan-920's four findings were raised by dispatch 6 under `mf-k920b`; with the
fix, the guard exits 1 on this change, naming round 0 and F1–F4.

**A1's repair** (never made — A1 dropped). The inline fix ran from the panel's end (dispatch 7 ended 22:58:31 +03:00) to
the fix commit `121d9be1` (22:59:12 +03:00). That commit carries `Flow-Key: flow.review-panel`,
not the row's key, and was squashed into `c20a7ab2`, whose subject is no task's — the guard's
commit rule refuses either as `-commit`, so the back-filled row records none.

**A2's repair.** Task 4's row ended at 19:48:10.089Z; the parent's next call, at 19:48:16.246Z,
read the base guard (`git show 75261910:scripts/…`) for task 5's parity check, and its row began at
19:48:47Z. The CLI cannot move a recorded start (`RecordDispatch` replaces `started_at` only on a
re-begin of an aborted row), so the repair is one `UPDATE` on the dev store. The window's tokens
are not re-attributed: the harvester's offsets are already past them.

**A3's reach.** 165 stored decisions carry `parent.effort` `unknown`, 1 carries `default`.

**A4's evidence.** Across the 60 newest subagent transcripts, 1163 messages carry
`"stop_reason": null` on every line and 31 carry a stop reason on some line; the main-thread
transcripts carry `tool_use` or `end_turn` on every record. The reviewer row with output 184
(dispatch 6) wrote two reports and a reproducer; one tool_use of 5035 characters records
`output_tokens: 28`. The task notification carries only `subagent_tokens` (a total), so the true
output is in no file. kan-988 already took the per-message maximum (`ParseAssistantRecords`),
which cannot help when every line is the first chunk. Inside a rolled-back transaction, migration
0042 marks 3197 of 6878 dispatch rows and 1407 stage runs.

**S1's measurement.** Run one after the other, the two long `## test` commands took 6:25 and
6:58 wall in kan-920's run (the guard harnesses 207 s); run together on this 10-core machine,
307 s wall — the harnesses 277 s, `go test ./... -race` 307 s, each slowed by the other.
Two concurrent `make -s web-dist` with `internal/web/dist` absent ran `npm ci` once and both
exited 0, the second waiting on the lock.

## Live check

`flowd` runs the harvester, the pricing and the migrations, and an agent never restarts it
(`<project>/CLAUDE.md`), so the store-side half of this check runs once the operator has rebuilt
and restarted `flowd` from this branch.

- **Before the restart**, from this worktree:
  `flow record dispatches -change kan-920-flow-improvement-port-the-markdown-and-lint -C .` shows
  `task-5-implementer` starting 22:48:16 +03:00, no longer 22:48:47.
- **After the restart**, against the dev store:
  `select decision->'parent'->>'effort', count(*) from decisions group by 1` lists no `unknown`
  (before: 165); the dispatch rows of this change's two panel slots (seq 6 and 7) carry
  `tokens.sidechain.output_unmeasured`, and
  `flow record render -kind ledger -change kan-920-flow-improvement-port-the-markdown-and-lint`
  renders their tokens as `output not measured` (before: `output 184` and `output 76`); a
  dispatch begun after the restart whose agent ran as a subagent stores `output_unmeasured` above
  zero and an `output` that leaves out those messages.
- **Not working looks like:** `unknown` still counted 165; the slot rows still rendering `output
  184` and `output 76`; a new subagent row with `output_unmeasured` absent.

## Decisions

### A fixed round needs a panel-fix row of its round or the next

**ID:** unrecorded-fix-round
**Status:** retired
**Retired:** A1 dropped. **Disproved by review 1:** `skills/flow/review-panel.md` (lines 802-809 at this branch) has the parent fix a round that raised no Critical and no Important inline, with "no dispatch record"; kan-920's round 0 raised four Minors and nothing else, so the missing `panel-fix` row is the contract working, not a defect. The guard commit was dropped and the back-filled `panel-fix-0` row deleted from the dev store.
**Chosen:** `check-panel-fix-single-dispatch` refuses each round in which a finding raised by a dispatch of this session token is recorded `fixed` while no `panel-fix` row of the change — any token, any agent id, `full-suite-fix-<n>` never — is keyed to that round or the next; a finding with no `dispatchSeq` is not judged. Exit 1 with a violation line naming the round and refs, the remedy a back-fill.
**Considered:** any `panel-fix` row at all for a change with fixed findings — a later unrecorded round would pass behind an earlier recorded one; exactly `panel-fix-<round>` — the store uses both numberings (**Context**) and would refuse half the recorded rounds; every fixed finding of the change — 268 historical pairs with neither row would refuse a resumed change for an earlier run's records; the inline prompt's **Continue** remedy — a recorded row is what the round lacks, and continuing records nothing.

### Back-fill kan-920's panel-fix-0 with no commit

**ID:** backfill-panel-fix-0
**Status:** retired
**Retired:** A1 dropped. **Disproved by review 1:** `skills/flow/review-panel.md` (lines 802-809 at this branch) has the parent fix a round that raised no Critical and no Important inline, with "no dispatch record"; kan-920's round 0 raised four Minors and nothing else, so the missing `panel-fix` row is the contract working, not a defect. The guard commit was dropped and the back-filled `panel-fix-0` row deleted from the dev store.
**Chosen:** `flow record dispatch backfill` for `panel-fix-0` under `mf-k920b`, `-agent-id inline`, the parent's pair, 22:58:31 to 22:59:12 +03:00, outcome `completed`, no `-commit`.
**Considered:** `-commit 121d9be1` or `-commit c20a7ab2` — the guard's commit rule refuses both (**Context**); leaving the round unrecorded — the guard this change adds would refuse kan-920 for good.

### An inline begin refuses earlier work in the worktree

**ID:** inline-begin-clean-tree
**Status:** active
**Chosen:** `flow record dispatch begin -agent-id inline` without `-started-at` refuses, exit 2, while `git status` of its `-C` worktree shows a tracked modification, staged or not, other than `spectre/changes/<change>/tasks.md`, naming each path. A git call that fails refuses nothing. Untracked files are not judged.
**Considered:** a check at `close-task.sh` — after the fact, with the row already written; raising the three-second floor of `dispatch end` — a tolerance that a seven-second row still passes; allowing the whole change directory — task 5's own `design.md` edit is the defect; judging untracked files — a new file is created by the work too, but a planning or scratch file outside `.gitignore` would refuse rows for no defect.

### Repair the task-5 row's start in the store

**ID:** repair-task-5-start
**Status:** active
**Chosen:** one `UPDATE dispatches SET started_at = '2026-10-21T19:48:16.246Z'` for change `kan-920-flow-improvement-port-the-markdown-and-lint`, token `mf-k920b`, key `task-5-implementer`, through `docker exec flow-postgres psql`; its tokens stay as attributed.
**Considered:** a second `begin` — the store replaces `started_at` only on an aborted row; a new CLI verb — a one-row repair needs no permanent surface; re-harvesting the window — the harvester's offsets are past it and nothing re-reads behind them.

### An unstated parent effort is default

**ID:** parent-effort-default
**Status:** active
**Chosen:** the planner's `parent` sentence cites **Harness mapping** (`unknown (agent-defined)` and `default`); `flow decision render` refuses a `parent.effort` outside `low`, `medium`, `high`, `default`; migration `0041_decision_parent_effort_default.sql` rewrites every stored `unknown` to `default`.
**Considered:** teaching `-effort` to accept `unknown` — two words for one fact, and every reader of the column learns both; repairing only kan-920's decision — 164 other rows carry the same wrong word; the check in `flow record decision` — Decide always renders before it records, and the render is where the file is corrected.

### A stream-start message's output is not measured

**ID:** output-unmeasured
**Status:** active
**Chosen:** a message every line of which carries an explicit `"stop_reason": null` is `Usage.OutputUnmeasured`; `Bucket.add` counts it in `output_unmeasured` and adds none of its output or thinking; `chargeableTokens.cost` refuses a bucket with `output_unmeasured` above zero; the ledger's token line and the SPA's `readOutputTokens` render such output as not measured; a line with no `stop_reason` key stays measured.
**Considered:** keeping the per-message maximum — kan-988 did, and every line is the first chunk; estimating output from the notification's `subagent_tokens` — a total, not an output count, so any split is a guess; pricing the measured part — a partial sum that reads as complete, the failure Price's `cache_creation_unknown` rule already refuses; treating a missing `stop_reason` key as streaming — older transcript shapes and every fixture without the key would read unmeasured.
**Consequence:** a stage run or dispatch whose bucket carries subagent usage gets no `cost_usd` from now on, until the harness writes final usage.

### Mark the stored rows, keep the stored costs

**ID:** output-unmeasured-repair
**Status:** active
**Chosen:** migration `0042_dispatch_output_unmeasured.sql` sets `tokens.sidechain.output_unmeasured` to 1 on every dispatch row and stage run whose sidechain bucket carries output above zero and no such key; `cost_usd` figures already written stay.
**Considered:** subtracting the stored output — the first-chunk share is not recoverable per message once summed; deleting stored `cost_usd` — Price never deletes a figure, and a migration that did would be the only writer that does; marking the per-model buckets under `models` too — each sits under its own model key, and the figure Price derived from them is already written.

### The test list runs concurrently

**ID:** concurrent-test-list
**Status:** active
**Chosen:** `run-verify-list.sh` starts every `test` command that is not skipped in the background, each into its own file, waits for each in list order and reports it then; a `lint` list runs one command at a time as before. `stats/Makefile`'s SPA build takes a lock directory in the git dir, so concurrent `make -s web-dist` callers on a fresh worktree build once.
**Considered:** a concurrency cap or a per-line opt-in — configuration nobody asked for, and the list has three lines; running `lint` concurrently too — not in the finding, and its guards are short; a lock inside the tree — the guard suite and the green-sha check both refuse a tree that changes under them; serialising only the SPA build in `run-verify-list.sh` — project knowledge in a project-agnostic runner.

### Pass 1 reads while the suite runs

**ID:** panel-during-suite
**Status:** active
**Chosen:** on a run whose panel opens a pass 1, the parent closes `flow.sdd-tdd`, enters `review-panel.md`, and runs the full suite between the pass-1 begins and the round's wait; a full-suite fix commit is reviewed by every pass-1 role in the next round over its delta, a role that raised nothing included.
**Considered:** re-dispatching pass 1 whole — its keys are taken for this token, and the delta is all that changed; the suite in the background — FOREGROUND BUILDS forbids ending a turn on it, and the slots already supply the concurrency; dispatching the slots before `sync-panel-base.sh` — its rebase would move the tree under them.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 4 spans the harvester, pricing, two renderers and a data migration, and Tasks 1 and 2 add refusals to gates every run passes; each needs a full first look.
**Considered:** `flow-low` — every diff is already verified and the tests fail without each fix, but a refusal that misfires stops runs, which a low-effort read is likelier to miss.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-920-flow-improvement-port-the-markdown-and-lint-fix`, Tasks 1 to 6 in order: each applies a diff verified in that order, Tasks 1 and 6 both edit `skills/flow/review-panel.md`, Tasks 2 and 6 both edit `skills/flow/implement.md`, and Tasks 3 and 4 number their migrations 0041 and 0042.
**Considered:** one fixer per task — six dispatches for six `git apply` runs; Go and prose as two groups — the shared files would serialise them anyway.

## Open questions
