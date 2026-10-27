# self-review-kan-1040-copyable-next-command

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the eight non-`big` findings of the self-review pass of
`kan-1040-copyable-next-command`, and repair every store row they already wrote wrong.

**Architecture:** a compose setting plus a live `docker update`; a byte snapshot-and-restore in
the reproducer runner; one key shape for `panel-fix-<round>` across the contract, the CLI and the
guard; a journalled-end flag honoured by the daemon handler and the journal replay alike; a
once-per-entry stuck warning in the reconciler; a deduplicating pass write in the store; a
stale-end check in the harvester's settle pass; and a journalled suite write.

**Spec:** `design.md` beside this file; the brainstorm record is
`.superpowers/sdd/2026-10-27-self-review-kan-1040-copyable-next-command-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Every task works in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-1040-copyable-next-command`, branch `spectre/kan-1040-copyable-next-command`; `**Files:**` paths are relative to its root. Never touch a main checkout, and never create a worktree.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean.
- Never stop, restart or recreate the protected `flowd` on `127.0.0.1:4173` or the `flow-postgres` container, and never drop the `flow` database. A repair writes the dev store through `cd stats && FLOWD_DSN='postgres://flow:flow@localhost:5433/flow?sslmode=disable' go run ./cmd/flowd <command> …` and touches only the rows its step names; run each repair once, after the task's tests pass, and quote its output in the report.
- Every new test fails with its task's fix reverted: run it once against the reverted source before committing and state the failure in the report.

## Review Focus

- Task 2: a run that times out or leaves a survivor still restores the snapshot; a path deleted before the run and recreated by it is removed again; a clean tree snapshots nothing and keeps every existing verdict.
- Task 3: `full-suite-fix-<n>` keys and every well-formed `panel-fix-<r>[-<n>][-retry]` with `r ≥ 1` are admitted by begin and backfill exactly as before.
- Task 4: an end carrying a caller-supplied `-ended-at` (not journalled) keeps that instant; an inline row or a row with no agent id is untouched by the agent-end path.
- Task 5: a stuck entry whose error text changes warns again; a path that drains forgets its stuck entry.
- Task 7: a single-begin row still settles on its first sight of an end at or after its begin.

- [ ] 1. Restart flow-postgres with Docker
  - [ ] **Step 1: Edit compose.** In `stats/docker-compose.yml`, under `services.flow-postgres`, add `restart: unless-stopped` after `container_name: flow-postgres`, with a one-line comment: Docker Desktop's quit left the store down until a manual `docker start` (kan-1040).
  - [ ] **Step 2: Apply it live, without a stop.** Run `docker update --restart unless-stopped flow-postgres`, then `docker inspect -f '{{.HostConfig.RestartPolicy.Name}}' flow-postgres` — it prints `unless-stopped` (before the step it printed `no`).
  - [ ] **Step 3: Verify.** `docker compose -f stats/docker-compose.yml config --quiet` exits 0.
  - [ ] **Step 4: Commit.**

**Build:** green
**Files:** `stats/docker-compose.yml`
**Tests:** **none** — a compose setting; Step 2's inspect is its check.
**Regression:** none — no test is declared.
**Commit:** fix(stats): restart flow-postgres with the Docker daemon
**After:** none

- [ ] 2. Restore an uncommitted edit a reproducer run overwrote
  - [ ] **Step 1: Write the failing test** `TestRunReproducerRestoresAnUncommittedEdit` in `stats/internal/guard/runreproducer_test.go`, shaped like `TestRunReproducerRefusesAChangedTree`: commit `f.txt` = `"x\n"`, then write `f.txt` = `"x\nlocal\n"` uncommitted; fixture script written by `rrFixture` at scripts/revert.sh inside the test's own temporary worktree, with body `git checkout -q -- f.txt; exit 1`. Assert: exit 5; output contains `run-reproducer: restored f.txt`; `f.txt` afterwards reads exactly `"x\nlocal\n"`.
  - [ ] **Step 2: Run it and see it fail** (`f.txt` reads `"x\n"`): `cd stats && go test ./internal/guard/ -run '^TestRunReproducerRestoresAnUncommittedEdit$' -count=1`.
  - [ ] **Step 3: Implement** in `stats/internal/guard/runreproducer.go` (decision `reproducer-restores-dirty-bytes`):
    - `func rrSnapshotDirty(worktree string) map[string]rrSnap` — the paths `git -C <worktree> diff --name-only -z HEAD --relative` prints; `rrSnap` holds the bytes, the file mode and whether the file was absent. A failed git call yields an empty map.
    - `func rrRestoreDirty(worktree string, snap map[string]rrSnap, stderr io.Writer) []string` — writes back every path whose current bytes (or presence) differ from its snapshot, removing one snapshotted absent that now exists, printing `run-reproducer: restored <path>` per path, and returns the restored paths sorted.
    - Take the snapshot beside `rrGitStatus` before `cmd.Start`; call `rrRestoreDirty` right after `rrSupervise` returns, before any exit decision. The exit-5 branch fires when the status listing differs **or** the restored list is non-empty; the exit-3 and exit-4 branches keep their codes.
  - [ ] **Step 4: Update the header.** In `scripts/run-reproducer.sh`, extend exit code 5's text: before the start the runner snapshots every tracked path that differs from `HEAD`, and after the run it writes back any the run changed, naming each on stderr (kan-1040: a reviewer's `git checkout -q -- skills` reverted an uncommitted edit).
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/guard/ && go vet ./internal/guard/ && go test ./internal/guard/ -run '^TestRunReproducer' -count=1` — green, `gofmt -l` silent.
  - [ ] **Step 6: Commit.**

**Build:** green
**Files:** `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`, `scripts/run-reproducer.sh`
**Tests:** `TestRunReproducerRestoresAnUncommittedEdit`
**Regression:** `TestRunReproducerRestoresAnUncommittedEdit` — reverting the commit leaves `f.txt` at its committed bytes.
**Baseline:** before=352 after=353
<!-- measured: cat stats/internal/guard/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1040-copyable-next-command -->
**Commit:** fix(guard): restore an uncommitted edit a reproducer run overwrote
**After:** none

- [ ] 3. Count panel-fix rounds from one
  - [ ] **Step 1: Write the failing test** `TestRecordDispatchRefusesPanelFixZero` in `stats/cmd/flow/record_test.go`: against a dead store address, `flow record dispatch begin` and `flow record dispatch backfill` with `-role panel-fix -key panel-fix-0` (and once with `panel-fix-01`) each exit 2 with stderr naming the key and `panel-fix-1`, and journal nothing; `-key panel-fix-1` and `-key full-suite-fix-1` are not refused by this check.
  - [ ] **Step 2: Run it and see it fail:** `cd stats && go test ./cmd/flow/ -run '^TestRecordDispatchRefusesPanelFixZero$' -count=1`.
  - [ ] **Step 3: Implement** (decision `panel-fix-counts-from-one`). In `stats/cmd/flow/record.go`, one `var panelFixKeyRE = regexp.MustCompile(`^panel-fix-[1-9][0-9]*(-[0-9]+)?(-retry)?$`)`; `runRecordDispatchBegin` and `runRecordDispatchBackfill`, before any store call, refuse exit 2 a `-key` with prefix `panel-fix-` that does not match it: `flow: -key %q: panel-fix-<round> counts fix rounds from 1 — the fix after panel-0 is panel-fix-1 (skills/flow/review-panel-fix-round.md)`. In `stats/internal/guard/panelfixsingledispatch.go`, `pfdKeyRE` becomes `^panel-fix-[1-9][0-9]*(-[0-9]+)?(-retry)?$` and its comment says the round counts from 1. In `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, the regexp table's `"panel-fix-01"` row expects `false` and a `"panel-fix-0": false` row is added; the kan-952 trailer test uses `panel-fix-1` in place of `panel-fix-0` throughout, its assertions otherwise unchanged. In `scripts/check-panel-fix-single-dispatch.sh`'s header, the key description says `<round>` counts from 1.
  - [ ] **Step 4: State it in the contract.** Apply this edit to `skills/flow/review-panel-fix-round.md` with `git apply`:

````diff verified:applied to this worktree's branch at plan time and reverted; check-verbatim-moves.sh then printed the added sentence as its one FAIL, which this change's verbatim-moves.txt acknowledges
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -440,5 +440,7 @@
 fragments one diff into competing fixups against the same worktree. Chunks run in order, each
 dispatch awaited in the foreground before the next chunk begins and before the round's reproducer
-re-runs begin, and no fix subagent is left in flight when the turn ends. The round's first
+re-runs begin, and no fix subagent is left in flight when the turn ends. `<round>` is the fix
+round, one above the review round whose findings it fixes: the fix after `panel-0` is
+`panel-fix-1`, and `flow record dispatch` refuses `panel-fix-0` (kan-1040). The round's first
 chunk's `-key` is exactly `panel-fix-<round>`; each further chunk appends `-<n>`
 (`panel-fix-<round>-2`, `-3`, …), contiguous from 2; the handshake retry suffixes `-retry` onto
````

  - [ ] **Step 5: Verify.** `cd stats && gofmt -l cmd/flow/ internal/guard/ && go vet ./cmd/flow/ ./internal/guard/ && go test ./cmd/flow/ -run '^TestRecordDispatch' -count=1 && go test ./internal/guard/ -run '^TestCheckPanelFixSingleDispatch' -count=1`, then `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — all green.
  - [ ] **Step 6: Repair the rows**, in this order (the run's key uniqueness needs 8232 moved first), with Task 4's repair already run:
    - `go run ./cmd/flowd rekey-dispatch 8232 panel-fix-2`
    - `go run ./cmd/flowd rekey-dispatch 8228 panel-fix-1`
    - `go run ./cmd/flowd rekey-dispatch 7834 panel-fix-1`
    - `go run ./cmd/flowd retime-dispatch 8228 2026-10-27T18:00:48.232Z 2026-10-27T18:15:03Z e3e1e20b424f17668baa04f377e7dea8a6aaed8e`
    <!-- measured: docker exec flow-postgres psql -U flow -d flow -c "select id, dispatch_key, started_at, ended_at, commit_sha from dispatches where id in (7834,8228,8232)" @ branch spectre/kan-1040-copyable-next-command -->
    Then the live check's dispatches query shows 7834 and 8228 keyed `panel-fix-1`, 8232 `panel-fix-2`, and 8228 started 18:00:48.232Z.
  - [ ] **Step 7: Commit.**

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/guard/panelfixsingledispatch.go`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `scripts/check-panel-fix-single-dispatch.sh`, `skills/flow/review-panel-fix-round.md`
**Tests:** `TestRecordDispatchRefusesPanelFixZero`
**Regression:** `TestRecordDispatchRefusesPanelFixZero` — reverting the commit admits `panel-fix-0` at begin and backfill.
**Baseline:** before=344 after=345
<!-- measured: cat stats/cmd/flow/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1040-copyable-next-command -->
**Commit:** fix(flow): count panel-fix rounds from one and refuse panel-fix-0
**After:** Task 4

- [ ] 4. Close a journalled dispatch end at its agent's last write
  - [ ] **Step 1: Write the failing test** `TestReplayClosesAJournalledEndAtTheAgentsLastWrite` in `stats/internal/reconcile/record_test.go`, beside `TestReplayAppliesPendingRecordEntries`: an open row with an agent id begun at T0; a journalled `dispatch-end` entry with `EndedAt` T0+10m and `EndedAtJournalled: true`; `Reconciler.AgentEnd` answering T0+2m for that agent. After `Run`, the row's end is T0+2m. A second case, the same entry without `EndedAtJournalled`, keeps T0+10m.
  - [ ] **Step 2: Run it and see it fail** (it does not compile until the field exists — that is the expected failure): `cd stats && go test ./internal/reconcile/ -run '^TestReplayClosesAJournalledEndAtTheAgentsLastWrite$' -count=1`.
  - [ ] **Step 3: Implement** (decision `journalled-end-takes-agent-end`):
    - `stats/internal/records/types.go`: `DispatchEnd.EndedAtJournalled bool` with tag `json:"endedAtJournalled,omitempty"` and a comment: the CLI stamped `EndedAt` with its attempt instant because the store was unreachable, so the daemon treats it as its own stamp.
    - `stats/cmd/flow/record.go`, `runRecordDispatchEnd`: where `callErr != nil && in.EndedAt.IsZero()` sets `in.EndedAt = attemptAt`, also set `in.EndedAtJournalled = true`.
    - `stats/internal/api/records.go`: `func ApplyDispatchEndAt(ctx context.Context, rs RecordStore, project, change string, in records.DispatchEnd, agentEnd func(agentID string) (time.Time, bool)) (records.Dispatch, error)` carries `endDispatch`'s apply-then-agent-end sequence with `stamped := in.EndedAt.IsZero() || in.EndedAtJournalled`; `agentEndOf` takes the `agentEnd` func as a parameter instead of reading the handler; `endDispatch` calls `ApplyDispatchEndAt(…, h.agentEnd)`.
    - `stats/internal/reconcile/reconcile.go`: `Reconciler.AgentEnd func(agentID string) (time.Time, bool)` (nil: no agent-end lookup); the `dispatch-end` case calls `api.ApplyDispatchEndAt(…, r.AgentEnd)`.
    - `stats/cmd/flowd/main.go`: name the closure passed to `api.WithAgentEnds` once as a local and assign the same value to `reconciler.AgentEnd` after `reconcile.New`.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l cmd/ internal/ && go vet ./cmd/flow/ ./cmd/flowd/ ./internal/api/ ./internal/reconcile/ ./internal/records/ && go test ./internal/reconcile/ -run '^TestReplay' -count=1 && go test ./internal/api/ -run 'Dispatch' -count=1 && go test ./cmd/flow/ -run 'DispatchEnd' -count=1` — green.
  - [ ] **Step 5: Repair rows 8223–8225** to their transcripts' last writes (`harvest.AgentTranscriptEnd` over `~/.claude/projects`):
    - `go run ./cmd/flowd retime-dispatch 8223 2026-10-27T17:49:08.28503Z 2026-10-27T17:51:11.494Z ''`
    - `go run ./cmd/flowd retime-dispatch 8224 2026-10-27T17:49:08.362213Z 2026-10-27T17:50:57.540Z ''`
    - `go run ./cmd/flowd retime-dispatch 8225 2026-10-27T17:58:50.167094Z 2026-10-27T18:00:48.232Z ''`
    <!-- measured: tail -c 4000 ~/.claude/projects/-Users-tweety53-Projects-agents/a0cddeeb-2693-4c22-8a31-3ab1945f9fe7/subagents/agent-<id>.jsonl | grep -o '"timestamp":"[^"]*"' | tail -1, for ab1f86cf03ae9e041, ad0f7cb24db10e191, a23793a6fc47b3f7b @ branch spectre/kan-1040-copyable-next-command -->
  - [ ] **Step 6: Commit.**

**Build:** green
**Files:** `stats/internal/records/types.go`, `stats/cmd/flow/record.go`, `stats/internal/api/records.go`, `stats/internal/reconcile/reconcile.go`, `stats/internal/reconcile/record_test.go`, `stats/cmd/flowd/main.go`
**Tests:** `TestReplayClosesAJournalledEndAtTheAgentsLastWrite`
**Regression:** `TestReplayClosesAJournalledEndAtTheAgentsLastWrite` — reverting the commit closes the replayed row at the journalled attempt instant.
**Baseline:** before=36 after=37
<!-- measured: cat stats/internal/reconcile/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1040-copyable-next-command -->
**Commit:** fix(stats): close a journalled dispatch end at its agent's last write
**After:** none

- [ ] 5. Warn once per stuck journal entry and make a store stall diagnosable
  - [ ] **Step 1: Write the failing test** `TestReplayWarnsOnceForAStuckEntry` in `stats/internal/reconcile/record_test.go`: a record journal holding one `dispatch-end` entry whose row does not exist (`store.ErrDispatchNotFound`, not definitive), a `slog` logger writing to a buffer; `Run` three times; the buffer holds exactly one `journal replay stopped on an unresolved entry` line, and the entry is still in the journal.
  - [ ] **Step 2: Run it and see it fail** (three lines): `cd stats && go test ./internal/reconcile/ -run '^TestReplayWarnsOnceForAStuckEntry$' -count=1`.
  - [ ] **Step 3: Implement** (decision `stuck-journal-warns-once`) in `stats/internal/reconcile/reconcile.go`: `Reconciler.stuck map[string]string`, guarded by `r.mu` (held by `Run`), maps a journal path to the error text last warned for it. `replayJournalFile`'s stop branch warns only when `r.stuck[path]` differs from `applyErr.Error()`, then records it; a replay of that path that ends without stopping deletes its entry.
  - [ ] **Step 4: Name the live log.** In `stats/README.md`, after the paragraph ending "regardless of how it is launched.", add: `make restart` runs `flowd` outside launchd instead, appending its log to `/tmp/flow-live.log` (`LIVE_LOG`, `stats/Makefile`); while it does, `~/Library/Logs/flowd.log` is not written. Postgres logs lock waits and statements slower than one second (`ALTER SYSTEM`, kan-1040), so a store stall leaves evidence in `docker logs flow-postgres`.
  - [ ] **Step 5: Turn the logging on, live, without a restart:** `docker exec flow-postgres psql -U flow -d flow -c "ALTER SYSTEM SET log_lock_waits = on" -c "ALTER SYSTEM SET log_min_duration_statement = 1000" -c "SELECT pg_reload_conf()"`, then `docker exec flow-postgres psql -U flow -d flow -Atc "show log_lock_waits" -c "show log_min_duration_statement"` prints `on` and `1s`.
    <!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "show log_lock_waits" -c "show log_min_duration_statement" printed off and -1; usesuper is t for role flow @ branch spectre/kan-1040-copyable-next-command -->
  - [ ] **Step 6: Verify.** `cd stats && gofmt -l internal/reconcile/ && go vet ./internal/reconcile/ && go test ./internal/reconcile/ -run '^TestReplay' -count=1` — green.
  - [ ] **Step 7: Commit.**

**Build:** green
**Files:** `stats/internal/reconcile/reconcile.go`, `stats/internal/reconcile/record_test.go`, `stats/README.md`
**Tests:** `TestReplayWarnsOnceForAStuckEntry`
**Regression:** `TestReplayWarnsOnceForAStuckEntry` — reverting the commit warns on every pass.
**Baseline:** before=37 after=38
<!-- measured: cat stats/internal/reconcile/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1040-copyable-next-command (plus Task 4's one test) -->
**Commit:** fix(stats): warn once per stuck journal entry and log store stalls
**After:** Task 4

- [ ] 6. Drop an identical pass-log note
  - [ ] **Step 1: Write the failing test** `TestRecordPassDropsAnIdenticalNote` in `stats/internal/store/records_test.go`, beside `TestRecordPanelPassAndMutation`: record round 0 note `n` twice — the second call returns the first row's ID and the run record carries one pass; record round 0 note `m` and round 1 note `n` — both are added (three passes).
  - [ ] **Step 2: Run it and see it fail:** `cd stats && go test ./internal/store/ -run '^TestRecordPassDropsAnIdenticalNote$' -count=1`.
  - [ ] **Step 3: Implement** (decision `pass-note-dedup-in-store`) in `stats/internal/store/records.go`: `RecordPass` inserts only where no `panel_passes` row of the change has the same round and note, and otherwise returns that row (unknown change still `ErrChangeNotFound`); its doc comment says so, replacing the sentence that no dedup key exists. `func (s *Store) DropPass(ctx context.Context, id int64) error` deletes one row by id, an unknown id an error. In `stats/cmd/flowd/main.go`, `flowd drop-pass <pass-id>` calls it, shaped like `rekey-dispatch`.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/store/ cmd/flowd/ && go vet ./internal/store/ ./cmd/flowd/ && go test ./internal/store/ -run 'Pass' -count=1` — green.
  - [ ] **Step 5: Repair:** `go run ./cmd/flowd drop-pass 2504`; then `docker exec flow-postgres psql -U flow -d flow -Atc "select count(*) from panel_passes where change_id = 2434 and round = 0 and note like 'dispatched: panel-0-mutation%'"` prints `1`.
    <!-- measured: docker exec flow-postgres psql -U flow -d flow -c "select p.* from panel_passes p join changes c on c.id=p.change_id where c.name like 'kan-1040%'" showed rows 2503 and 2504 with the identical round-0 note @ branch spectre/kan-1040-copyable-next-command -->
  - [ ] **Step 6: Commit.**

**Build:** green
**Files:** `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flowd/main.go`
**Tests:** `TestRecordPassDropsAnIdenticalNote`
**Regression:** `TestRecordPassDropsAnIdenticalNote` — reverting the commit records the identical note twice.
**Baseline:** before=323 after=324
<!-- measured: cat stats/internal/store/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1040-copyable-next-command -->
**Commit:** fix(stats): drop a pass-log note identical to one already recorded
**After:** Task 4

- [ ] 7. Wait out a stale end before settling an inline row
  - [ ] **Step 1: Write the failing test** `TestSettleClosedRowsWaitsOutAStaleEnd` in `stats/internal/harvest/watcher_test.go`, beside `TestSettleClosedRowsMeasuresARunVerifyListRow`: two `run-verify-list.sh … verify …` begins under one key, the second after the first's end; the row's `ended_at` first reads the first run's end (before the second begin), and a later pass reads an end after the parent turns that fall between the two begins. The row's final restamp carries those turns' output, never a zero patch.
  - [ ] **Step 2: Run it and see it fail** (zero restamp): `cd stats && go test ./internal/harvest/ -run '^TestSettleClosedRowsWaitsOutAStaleEnd$' -count=1`.
  - [ ] **Step 3: Implement** (decision `settle-waits-out-stale-end`) in `stats/internal/harvest/watcher.go`, `settleClosedRows`: a candidate is kept when `endedAt.Before(c.StartedAt)`, alongside the existing not-found, zero and after-newest conditions, with a comment naming kan-1040's `verify` row.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/harvest/ && go vet ./internal/harvest/ && go test ./internal/harvest/ -run 'Settle' -count=1` — green.
  - [ ] **Step 5: Repair row 8235:** `go run ./cmd/flowd restamp-dispatch 8235 /Users/tweety53/.claude/projects/-Users-tweety53-Projects-agents/a0cddeeb-2693-4c22-8a31-3ab1945f9fe7.jsonl`; then `docker exec flow-postgres psql -U flow -d flow -Atc "select metrics->'tokens'->'main' from dispatches where id = 8235"` shows a non-zero output figure.
  - [ ] **Step 6: Commit.**

**Build:** green
**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`
**Tests:** `TestSettleClosedRowsWaitsOutAStaleEnd`
**Regression:** `TestSettleClosedRowsWaitsOutAStaleEnd` — reverting the commit zeroes the row against the first run's end.
**Baseline:** before=176 after=177
<!-- measured: cat stats/internal/harvest/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1040-copyable-next-command -->
**Commit:** fix(stats): wait out a stale end before settling an inline row
**After:** none

- [ ] 8. Journal a suite run the store could not take
  - [ ] **Step 1: Write the failing test** `TestSuiteRecordJournalsWhenStoreDown` in `stats/cmd/flow/suite_test.go`: with `isolatedStateRoot` and a dead address, `flow suite record -suite probe -- /bin/sh -c 'exit 5'` exits 5, and the record journal `fallback.JournalFilePath(<project key>, "_suite-runs") + ".record"` holds one entry of kind `suite` whose request carries suite `probe`, exit code 5 and a non-zero `ranAt`. In `TestSuiteRecordWarnsWithoutAlteringExitWhenStoreDown`, the stderr needle becomes `wrote local journal`.
  - [ ] **Step 2: Run it and see it fail:** `cd stats && go test ./cmd/flow/ -run '^TestSuiteRecordJournalsWhenStoreDown$' -count=1`.
  - [ ] **Step 3: Implement** (decision `suite-run-journals`). `stats/cmd/flow/suite.go`: `const suiteJournalName = "_suite-runs"`; on `callErr` that is neither `client.ErrRecordRejected` nor `client.ErrNotFound`, set `in.RanAt = start` and call `journalRecordWrite(projectKey, suiteJournalName, "suite", in, stderr)`; a refusal keeps today's not-recorded warning; the child's exit code is returned either way. `stats/cmd/flow/record.go`: `recordJournalBody`'s kind list names `"suite"`. `stats/internal/reconcile/reconcile.go`: `applyRecordEntry` gains `case "suite"`, decoding a `records.SuiteRun` and calling `api.ApplySuiteRunRecord(ctx, r.recordStore, e.Project, in)`; extend `TestReplayAppliesVerdictAndIncidentEntries` in `stats/internal/reconcile/record_test.go` with one `suite` entry asserting the store received it.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l cmd/flow/ internal/reconcile/ && go vet ./cmd/flow/ ./internal/reconcile/ && go test ./cmd/flow/ -run '^TestSuiteRecord' -count=1 && go test ./internal/reconcile/ -run '^TestReplay' -count=1` — green.
  - [ ] **Step 5: Commit.**

**Build:** green
**Files:** `stats/cmd/flow/suite.go`, `stats/cmd/flow/suite_test.go`, `stats/cmd/flow/record.go`, `stats/internal/reconcile/reconcile.go`, `stats/internal/reconcile/record_test.go`
**Tests:** `TestSuiteRecordJournalsWhenStoreDown`
**Regression:** `TestSuiteRecordJournalsWhenStoreDown` — reverting the commit leaves no journal entry for the run.
**Baseline:** before=345 after=346
<!-- measured: cat stats/cmd/flow/*_test.go | grep -c -E '^func Test' @ branch spectre/kan-1040-copyable-next-command (plus Task 3's one test) -->
**Commit:** fix(flow): journal a suite run the store could not take
**After:** Task 3, 5
