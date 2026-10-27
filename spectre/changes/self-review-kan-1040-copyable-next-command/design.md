## Context

One spectre change for the eight non-`big` findings of the self-review pass of
`kan-1040-copyable-next-command`. The fixes ride the change's own branch
`spectre/kan-1040-copyable-next-command` in this worktree, from the HEAD it held when the pass
started (`9efd93be`). Every finding is carried by exactly one task; the brainstorm record is
`.superpowers/sdd/2026-10-27-self-review-kan-1040-copyable-next-command-design.md` in this worktree.

Facts settled at plan time:

- **Finding 1**: `docker inspect -f '{{.HostConfig.RestartPolicy.Name}}' flow-postgres` prints `no`;
  `stats/docker-compose.yml` declares no `restart:`.
- **Finding 2**: `runReproducer` (`stats/internal/guard/runreproducer.go`) compares `git status
  --porcelain` before and after the run and exits 5 on a difference; it never holds the bytes it
  would need to put an overwritten edit back.
- **Finding 3**: the review rounds are `panel-0`, `panel-1`, `panel-2`; `panelFixFloor`
  (`stats/cmd/flow/record.go`) and the panel-fix guard's chunk bound read `panel-fix-<r>` as the
  fix after review round r-1, while `skills/flow/review-panel-fix-round.md` never says which
  round `<round>` names. kan-1040 recorded `panel-fix-0` (row 8228, after `panel-0`) and
  `panel-fix-1` (row 8232, after `panel-1`); kan-1036-late-round-finding-prompt-tuning recorded
  `panel-fix-0` (row 7834, after its `panel-0`). Row 8228 started 17:48:40Z, before any `panel-0`
  reviewer began, because `panelFixFloor` has no floor for round 0.
- **Finding 4**: the three `panel-0-*` ends (rows 8223–8225) failed on the CLI's client timeout at
  18:07:44, 46 and 48Z (flowd log `/tmp/flow-live.log`), were journalled with the attempt instant,
  and were replayed by the reconciler (`reconcile: … applied=3` at 18:07:50Z), which calls
  `api.ApplyDispatchEnd` directly — `agentEndOf` lives only in the HTTP handler
  (`stats/internal/api/records.go`). Their agent transcripts last wrote at 17:51:11.494Z,
  17:50:57.540Z and 18:00:48.232Z.
- **Finding 5** (trace): no lock on `projects` exists to hold — its only write is `INSERT … ON
  CONFLICT (project_key) DO NOTHING` (`stats/internal/store/changes.go`), and no non-test code
  takes `FOR UPDATE` or updates or deletes it. The 18:09:46Z cancel is flow-guard's 500 ms
  telemetry timeout (`stats/cmd/flow-guard/telemetry.go`), a routine event. flowd never restarted;
  <!-- measured: grep -n telemetryTimeout stats/cmd/flow-guard/telemetry.go; docker logs flow-postgres; /tmp/flow-live.log 20:09:46.904 local @ branch spectre/kan-1040-copyable-next-command -->
  every store call slowed for about twelve seconds, cause unproven, because postgres runs with
  `log_lock_waits=off` and no slow-statement log. The live daemon logs to `/tmp/flow-live.log`
  (`LIVE_LOG`, `stats/Makefile`, written by `make restart`) while `stats/README.md` names only the
  launchd agent's `~/Library/Logs/flowd.log`, and that live log is swollen by journals stuck on
  `dispatch not found`, each warned on every ten-second reconcile pass.
- **Finding 6**: pass rows 2503 and 2504 of kan-1040 carry the identical round-0 note
  `dispatched: panel-0-mutation+failure-modes on sonnet/low`; `Store.RecordPass` is a plain
  append.
- **Finding 7** (trace): row 8235 (`verify`, inline) and its transcript share session
  `a0cddeeb-2693-4c22-8a31-3ab1945f9fe7`, so the id-mismatch candidate is disproved. Two
  `run-verify-list.sh` calls used the key `verify` (lint, then test). `settleClosedRows`
  (`stats/internal/harvest/watcher.go`) settled both candidates while the row still held the lint
  run's end, wrote zeros over that empty window, and dropped them; the test run's end later moved
  `ended_at` to 18:39:47Z and nothing measured the row again.
- **Finding 8**: `runSuiteRecord` (`stats/cmd/flow/suite.go`) prints a not-recorded warning on
  any store failure and journals nothing; `api.ApplySuiteRunRecord` exists and
  `records.SuiteRun.RanAt` is honoured by the store (`COALESCE($6, now())`).

## Decisions

### Postgres restarts with Docker

**ID:** postgres-restart-policy
**Status:** active
**Chosen:** `restart: unless-stopped` on `flow-postgres` in `stats/docker-compose.yml`, and `docker
update --restart unless-stopped flow-postgres` on the live container, which changes the policy
without stopping it.
**Considered:** `always` — would also restart a container the operator stopped on purpose;
compose only — leaves the running container on `no` until someone recreates it, which is a stop.

### The reproducer runner restores what it changed

**ID:** reproducer-restores-dirty-bytes
**Status:** active
**Chosen:** before `cmd.Start`, `runReproducer` reads the bytes and mode of every path `git diff
--name-only -z HEAD --relative` names (absent files recorded as absent); after `rrSupervise`
returns, on every exit path, it writes back each snapshotted path whose bytes changed (removing
one recorded absent that now exists) and names each on stderr. A run that restored anything exits
5 like a changed status listing does, unless the run already answers 3 or 4.
**Considered:** refusing a reproducer that names `git checkout`/`git restore` — the destruction can
come from any command; running the reproducer in a scratch worktree — a mutation reproducer must
run against the tree under review, and a copy changes its paths.

### panel-fix rounds count from one

**ID:** panel-fix-counts-from-one
**Status:** active
**Chosen:** `<round>` in `panel-fix-<round>` is the fix round, one above the review round it
fixes — the reading `panelFixFloor` and the guard's chunk bound already use. One sentence in
`skills/flow/review-panel-fix-round.md` states it; `flow record dispatch begin` and `backfill`
refuse, exit 2 before any store call, a key that starts `panel-fix-` and does not match
`^panel-fix-[1-9][0-9]*(-[0-9]+)?(-retry)?$`; `pfdKeyRE` takes the same `[1-9][0-9]*`. The rows
are repaired with `flowd rekey-dispatch` (8232 → `panel-fix-2` first, then 8228 →
`panel-fix-1`; 7834 → `panel-fix-1`) and 8228's start with `flowd retime-dispatch` to
18:00:48.232Z, the latest `panel-0` reviewer end once Task 4's repair lands.
**Considered:** reading `<round>` as the review round — `panelFixFloor`, the chunk bound and every
other run's rows would all have to change instead of two runs' rows.

### A journalled end takes the agent's own end

**ID:** journalled-end-takes-agent-end
**Status:** active
**Chosen:** `records.DispatchEnd` gains `EndedAtJournalled bool` (`json:"endedAtJournalled,omitempty"`),
set by `runRecordDispatchEnd` exactly where it stamps the attempt instant. The handler's
apply-then-agent-end logic moves into `api.ApplyDispatchEndAt(ctx, rs, project, change, in,
agentEnd)`, whose stamped test is `in.EndedAt.IsZero() || in.EndedAtJournalled`; the handler and
the reconciler's `dispatch-end` replay both call it, the reconciler through a new
`Reconciler.AgentEnd` field that `cmd/flowd` sets from the same closure it passes to
`api.WithAgentEnds`. Rows 8223–8225 are repaired with `flowd retime-dispatch` to their
transcripts' last writes.
**Considered:** clearing `EndedAt` in the journal so the replay looks daemon-stamped — the replay
would then date the end to the replay; teaching only the handler — the kan-1040 ends never
reached it.

### Stuck journal entries warn once; the stall becomes diagnosable

**ID:** stuck-journal-warns-once
**Status:** active
**Chosen:** the reconciler remembers, per journal path, the error text it last warned for a stuck
entry, and warns again only when it changes; a path whose replay ends without stopping forgets
it. `stats/README.md` names `/tmp/flow-live.log` as the log of the daemon `make restart` runs.
`flow-postgres` gets `log_lock_waits = on` and `log_min_duration_statement = 1000` through `ALTER
SYSTEM` plus `pg_reload_conf()`, which persists in the data volume and needs no restart.
**Considered:** retiring entries stuck on `dispatch not found` — a begin still queued can resolve
them (`isDefinitiveRecordOutcome`); lowering the warning to debug — hides a newly stuck entry;
setting the postgres flags in compose — they take effect only on a recreate, which is a stop.

### The store drops an identical pass note

**ID:** pass-note-dedup-in-store
**Status:** active
**Chosen:** `Store.RecordPass` inserts only when no row of the change carries the same round and
the same note, and returns the existing row otherwise — one place both the live write and the
journal replay route through. `flowd drop-pass <pass-id>` (`Store.DropPass`) deletes row 2504.
**Considered:** deduplicating in `panel-round-begins.sh` — misses a replayed write; a unique index
— would make the replay's duplicate an error the reconciler would have to classify.

### Settle waits out a stale end

**ID:** settle-waits-out-stale-end
**Status:** active
**Chosen:** `settleClosedRows` keeps a candidate whose row's `ended_at` is before the candidate's
own begin instant (`c.StartedAt`), the end of an earlier run under the same key. Row 8235 is
repaired with `flowd restamp-dispatch 8235 <parent transcript>`.
**Considered:** refusing a second begin under a completed key — `run-verify-list.sh`'s lint-then-test
pair relies on reusing `verify`; distinct keys per call — every project's verify contract would
change, and the harvester would still mis-settle any other reused key.

### Suite runs journal

**ID:** suite-run-journals
**Status:** active
**Chosen:** on a store failure other than a refusal (`client.ErrRecordRejected`,
`client.ErrNotFound`), `runSuiteRecord` sets `RanAt` to the run's start and journals the row as
kind `suite` under the record journal of the reserved name `_suite-runs`; the reconciler's
`applyRecordEntry` replays it through `api.ApplySuiteRunRecord`. The child's exit code still
passes through unchanged.
**Considered:** a separate suite journal file and replay loop — a second copy of what the record
journal already does.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — the reproducer runner writes files back into a worktree, and the
harvester, the reconciler and the store write paths change behaviour on every run.
**Considered:** `flow-low` — Task 1 alone would qualify.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** three groups, in series, each on `flow-low`/`opus` —
`self-review-kan-1040-copyable-next-command-fix-g1`: Tasks 1, 2 (the dev stack and the reproducer
runner); `-fix-g2`: Tasks 3, 4, 7 (dispatch keys, ends and harvest, with their row repairs);
`-fix-g3`: Tasks 5, 6, 8 (reconciler, store and suite journalling, with the pass repair).
**Considered:** one fixer for all eight — the store repairs, the harvester and the reproducer runner
in one context; one per task — eight dispatches where three file-coherent groups serve.

## Open questions

none

## Recommended approach

Fix each defect where it lives: the container's restart policy (`postgres-restart-policy`), the
runner's restore (`reproducer-restores-dirty-bytes`), the key's numbering (`panel-fix-counts-from-one`),
the journalled end (`journalled-end-takes-agent-end`), the reconciler's warning and postgres's
logging (`stuck-journal-warns-once`), the pass write (`pass-note-dedup-in-store`), the settle pass
(`settle-waits-out-stale-end`) and the suite write (`suite-run-journals`), repairing each row the
defect already wrote. Every rejected alternative is under its decision.

## Critical files

- `stats/docker-compose.yml` — the `flow-postgres` service.
- `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`, `scripts/run-reproducer.sh` — the runner, its tests, its exit-code header.
- `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go` — dispatch begin/backfill key check, the journalled end's flag, `panelFixFloor` (read).
- `stats/internal/guard/panelfixsingledispatch.go`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `scripts/check-panel-fix-single-dispatch.sh`, `skills/flow/review-panel-fix-round.md` — the key regexp, its tests, its header, its contract.
- `stats/internal/records/types.go` — `DispatchEnd`.
- `stats/internal/api/records.go` — `endDispatch`, `agentEndOf`, the new `ApplyDispatchEndAt`.
- `stats/internal/reconcile/reconcile.go`, `stats/internal/reconcile/record_test.go` — the `dispatch-end` and `suite` replays, the stuck-entry warning.
- `stats/cmd/flowd/main.go` — the reconciler's `AgentEnd`, `drop-pass`; `retime-dispatch`, `rekey-dispatch`, `restamp-dispatch` (read, unchanged).
- `stats/internal/store/records.go`, `stats/internal/store/records_test.go` — `RecordPass`, `DropPass`.
- `stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go` — `settleClosedRows`.
- `stats/cmd/flow/suite.go`, `stats/cmd/flow/suite_test.go` — `runSuiteRecord`.
- `stats/README.md` — the live log's location.
- `spectre/changes/self-review-kan-1040-copyable-next-command/verbatim-moves.txt` — the one acknowledged new sentence.

## Existing code to reuse

- `rrGitStatus` and the exit-5 branch (`stats/internal/guard/runreproducer.go`); `rrWorktree`, `rrFixture`, `rrRun`, `rrExpect` (`runreproducer_test.go`).
- `harvest.AgentTranscriptEnd` and `agentEndOf` (`stats/internal/api/records.go`); `api.WithAgentEnds`.
- `journalRecordWrite`, `classifyRecordWrite`, `recordJournalBody` (`stats/cmd/flow/record.go`); `applyRecordEntry` (`stats/internal/reconcile/reconcile.go`).
- `flowd retime-dispatch`, `rekey-dispatch`, `restamp-dispatch` and the `rekeyDispatch` command shape `drop-pass` copies (`stats/cmd/flowd/main.go`).

## Verification

Each task's own verify step, then `cd stats && make -s web-dist && go vet ./... && gofmt -l .` clean
and the stats Go suite green (`.flow/project.md` `## test`), then the live check below.

## Live check

Against the dev workspace's own `flow-postgres` and `flow` database, never stopping either, and
never restarting `flowd`:

- `docker inspect -f '{{.HostConfig.RestartPolicy.Name}}' flow-postgres` prints `unless-stopped`
  (before: `no`); still `no` means Task 1's live step did not run.
- `docker exec flow-postgres psql -U flow -d flow -Atc "show log_lock_waits; show log_min_duration_statement"`
  prints `on` and `1s` (before: `off`, `-1`).
- `docker exec flow-postgres psql -U flow -d flow -Atc "select id, dispatch_key, started_at, ended_at from dispatches where id in (7834,8223,8224,8225,8228,8232) order by id"`
  shows 7834 and 8228 keyed `panel-fix-1`, 8232 `panel-fix-2`, 8223–8225 ended at
  17:51:11.494Z, 17:50:57.540Z, 18:00:48.232Z, and 8228 started 18:00:48.232Z; any `panel-fix-0`
  key or a `panel-0-*` end at 18:07Z means the repair did not run.
- Row 8235's `metrics->'tokens'->'main'` carries a non-zero output figure (before: every field 0).
- `select count(*) from panel_passes where change_id = 2434 and round = 0 and note like 'dispatched: panel-0-mutation%'` is 1 (before: 2).
- The behaviour changes run against the stand-ins their tests build — a scratch git worktree for
  the reproducer runner, the test store and fake transcripts for the daemon paths — since the
  protected `flowd` is not rebuilt before landing.
