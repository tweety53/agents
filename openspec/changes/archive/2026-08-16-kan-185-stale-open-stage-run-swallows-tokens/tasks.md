> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** A dropped `stage end` mark costs its own stage's telemetry and nothing more. The
harvester stops handing an overlapping message to whichever window happened to come back first, and
the store stops leaving two windows of one session open at the same time.

## Global Constraints

- **Two independent parts, in two packages, in this order.** Task 1 is `internal/harvest` and
  touches no SQL; task 2 is `internal/store` and touches no attribution code. Neither depends on
  the other's behaviour, and both must land for the change to be what KAN-185 describes.
- **One migration, and it adds an index only.** `session_token` already exists
  (`0008_stage_run_session_token.sql`) and `outcome` is free text with no CHECK constraint, so
  tasks 1 and 2 need no schema change at all. Task 3 adds `0009`, a partial index serving the
  supersede UPDATE's own predicate — the operator's answer to panel finding F3, which measured that
  UPDATE filtering the whole open-run set. No task alters a column, a constraint or existing data.
- **No API, CLI or SPA change.** `superseded` reads on every existing view exactly as any other
  closed run's outcome does. No TypeScript is edited by this plan.
- **The store package's tests need the PostgreSQL stack, and they self-skip without it.** Bring it
  up with `cd stats && docker compose up -d` before task 2. A skipped test prints `ok` for the
  package — treat a run whose `-v` output shows `SKIP` for the new tests as *not run*, never as
  passed.
- **No task edits `openspec/` or `docs/superpowers/`.**

## Baseline

**Measured 2026-08-16 against `b85d4f5`:** `go test ./internal/harvest/... ./internal/store/...
-count=1` passes both packages (`ok` for each). `stats/internal/harvest/attribute_test.go` holds 16
top-level `Test` functions; `stats/internal/store/stageruns_test.go` holds 29.
<!-- measured: cd stats && go test ./internal/harvest/... ./internal/store/... -count=1, and grep -c '^func Test' on the two files @ b85d4f5 -->

**The defect's own fingerprint, for anyone reproducing it:** stage run 146 (`finish.verify-merge`,
`kan-175-more-ui-ux-fixes`) sat open from 10:11:01Z and absorbed 187.8M cache-read tokens that
belonged to nine stage runs of `kan-184-harden-the-release-and-deploy-path`.
<!-- measured: read from the myflow store's stage_runs table on 2026-08-16; the rows have since been closed by hand, so this is not re-runnable -->

---

### 1 Attribute an overlapping message to the window that started last

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/attribute.go`
- Modify: `stats/internal/harvest/attribute_test.go`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: `Window.contains`, unchanged — which windows are candidates is not what this task
  changes.
- Produces: `bestWindow`'s selection rule, stated as "latest `StartedAt` wins, `Attempt` breaks a
  tie at the same instant" instead of "highest `Attempt` wins".

`bestWindow` today keeps the first containing window it sees and replaces it only on a strictly
higher `Attempt`. Two windows at attempt 1 — an orphan left open by a dropped end mark, and the
stage actually running — therefore resolve by iteration order, which is
`storeWindowSource.WindowsForSession`'s unsorted query, ordered by the primary key: oldest first,
forever.

- [x] **Step 1: The failing tests, written first**

Add to `stats/internal/harvest/attribute_test.go`, following the existing
`fakeWindowSource` / `harvest.NewAttributor` / `Attribute` shape and the both-orders loop
`TestReplayedBeginPrefersHighestAttempt` already uses:

- `TestOverlappingOpenWindowsPreferTheLatestStarted` — an orphan open from 15:00 at attempt 1 and a
  live window open from 15:10 at attempt 1, a record at 15:25, asserted in **both** returned
  orders: the delta belongs to the later-started run and the orphan has no entry at all. This is the
  incident, reduced.
- `TestSameInstantWindowsFallBackToHighestAttempt` — two windows sharing one `StartedAt`, attempts 1
  and 2: the higher attempt wins, so the tiebreaker survives with something to do.

Run them and watch the first fail before touching `attribute.go`:

```bash unverified:neither test exists yet — confirm the -run pattern matches both once step 1 has written them
cd stats && go test ./internal/harvest/ -run 'TestOverlappingOpenWindows|TestSameInstantWindows' -count=1 -v
```

- [x] **Step 2: The rule**

In `bestWindow`, replace the attempt-only comparison with: prefer a strictly later `StartedAt`;
where the starts are equal, prefer the strictly higher `Attempt`. Keep the first-match-wins
behaviour for a window that is neither later nor higher, so the function stays deterministic under
any input order.

- [x] **Step 3: The doc comment says why**

Rewrite `bestWindow`'s doc comment so it states the new rule and the reason: of two windows that
both contain a message, the one that started later is the one the session is actually in, and a
stale open window must not outrank it. Keep the existing paragraph explaining that this is distinct
from `Window.contains`' half-open boundary rule, and add that `Attempt` now decides only a
same-instant tie — naming the replayed-begin case it was written for, and that latest-start already
returns that same window because attempt 2 starts after attempt 1.

- [x] **Step 4: The whole package, including the tests that pinned the old rule**

```bash verified:the go test half was run at b85d4f5 and passed; vet and gofmt are .myflow/project.md's own lint commands
cd stats && go test ./internal/harvest/... -count=1 && go vet ./internal/harvest/... && gofmt -l internal/harvest
```

`TestReplayedBeginPrefersHighestAttempt` must pass **unchanged** — it is the case the old tie-break
existed for, and the new rule has to return the same window for it.

- [x] **Step 5: Correct the replay paragraph** *(added by fix round 2, from panel findings F1 and F4)*

`bestWindow`'s new doc comment says two things about the replayed-begin case that this change itself
makes false. It says the first attempt "stays open until the sweeper closes it" — task 2's supersede
now closes it at the replay's own start instant. And it says "a replayed attempt always starts after
the one it replays" — the journal replays the **original** request unchanged
(`stats/cmd/myflow/stage.go` captures `StartedAt` once before the RPC attempt;
`stats/internal/reconcile/reconcile.go` replays it as `StartedAt: req.StartedAt`), so the two
attempts start at the *same* instant and `Attempt` is what decides.

Rewrite that paragraph to say both correctly: a replay starts at the same instant as the attempt it
replays, which is why the same-instant `Attempt` tiebreak is the ordinary replay path rather than a
curiosity, and the earlier attempt is closed as `superseded` by the replay's own begin rather than
left for the sweeper. `design.md`'s `latest-start-wins` decision and the delta spec's replayed-begin
scenario carry the same correction, already made by the parent.

- [x] **Step 6: Say where the same-instant tie-break is actually reachable** *(added by fix round 4, from panel findings F7 and F11)*

Step 5's correction was right about the timestamps and wrong about the consequence. A replay does
carry the original's start instant — but task 2's supersede then closes the earlier attempt with
`ended_at` equal to that same instant, and `Window.contains` rejects an empty `[T, T)` interval. So
the ordinary replay path never presents two containing windows at all, and the same-instant
`Attempt` tie-break is **not** reachable through it.

Rewrite the paragraph a third time to say what is true: the tie-break is defensive, reachable for
windows the store no longer produces on the ordinary path — a run recorded before this change, a
run carrying no session token, or the moment before a supersede commits — and `Attempt` is kept for
those rather than for the replay case. Do not claim it is the ordinary path.

Also correct the doc comment on `TestReplayedBeginPrefersHighestAttempt`
(`stats/internal/harvest/attribute_test.go`), which frames its two windows — ten minutes apart — as
a replay. They are not: it is a latest-start case that also happens to carry a higher attempt.
Say so in the comment. Leave the test's assertions exactly as they are; what is wrong is the
narrative, not the expectation.

**Tests:** `TestOverlappingOpenWindowsPreferTheLatestStarted`,
`TestSameInstantWindowsFallBackToHighestAttempt`.

**Regression:** Reverting this task makes an orphaned open window win every message a later window
of the same session also contains — the two hours of kan-184's telemetry recorded against kan-175.

**Baseline:** before=16 after=18 top-level `Test` functions in
`stats/internal/harvest/attribute_test.go`.
<!-- predicted: grep -c '^func Test' stats/internal/harvest/attribute_test.go after step 1 -->

- [x] **Step 7: A test that pins the precedence itself** *(added by fix round 5, from panel finding F15)*

Three tests exercise `bestWindow`'s tie-break and not one of them discriminates its **order**.
`TestOverlappingOpenWindowsPreferTheLatestStarted` has both windows at attempt 1;
`TestSameInstantWindowsFallBackToHighestAttempt` has both starts equal; and
`TestReplayedBeginPrefersHighestAttempt`'s winner carries *both* the later start and the higher
attempt. Swap the comparator back to attempt-first and all three still pass — which is exactly how
the KAN-185 incident would return, with an orphan at a higher attempt number reabsorbing a live
stage's tokens and nothing failing.

Add `TestLatestStartOutranksAHigherAttempt`: window A with the **earlier** start and the **higher**
attempt, window B with the **later** start and the **lower** attempt, a record inside both,
asserted in both returned orders. B wins. Prove it discriminates by reversing the two clauses in
`bestWindow` in a scratch copy and watching only this test fail.

**Commit:** `fix(1): attribute an overlapping message to the window that started last`

---

### 2 Close the open runs a begin mark supersedes

**Build:** green

**Files:**
- Modify: `stats/internal/store/stageruns.go`
- Modify: `stats/internal/store/stageruns_test.go`
- Modify: `stats/internal/store/aggregate_test.go` *(added by fix round 1 — see step 6)*
- Modify: `stats/internal/api/stages.go` *(added by fix round 2 — comment only, see step 7)*
- Modify: `stats/internal/api/stages_test.go` *(added by fix round 2 — comment only, see step 7)*
- Modify: `stats/internal/harvest/watcher.go` *(added by fix round 3 — comment only, see step 8)*
- Modify: `stats/internal/harvest/watcher_test.go` *(added by fix round 3 — comment only, see step 8)*

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: `BeginStageInput.SessionToken` and `StartedAt`, both already present on every mark.
- Produces: `BeginStage` as one transaction — the insert it already performed, plus the closing of
  every still-open run of the same session token that started no later than the new run, with
  outcome `superseded`.
- Unchanged: `BeginStage`'s signature, its attempt-collision retry loop, and
  `ErrChangeNotFound`/`ErrTooManyAttemptCollisions`.

`insertStageRun` is a single `INSERT ... SELECT` today, and `BeginStage` retries it on the attempt
unique-violation. The supersede must land with the insert or not at all, so the pair moves into one
`pgx` transaction; the retry loop stays outside it, so a collision still rolls back cleanly and
tries again.

- [x] **Step 1: The failing tests, written first**

Add to `stats/internal/store/stageruns_test.go`, following the existing `newTestStore` /
`seedChange` / `baseBeginInput` shape:

- `TestBeginStageSupersedesAnEarlierOpenRunOfTheSameSession` — begin stage A, then begin stage B
  with the **same** session token and a later `StartedAt`; A comes back closed with
  `outcome = "superseded"` and `ended_at` equal to B's `started_at`, and B is open and untouched.
  Assert across two **different changes** sharing one token, because that is the incident: the
  orphan belonged to `kan-175` and the live stage to `kan-184`.
- `TestBeginStageLeavesAnOpenRunThatStartedLaterAlone` — an open run started at 11:00, then a begin
  carrying an older `StartedAt` of 10:00, as a journal replay does: the 11:00 run stays open.
- `TestBeginStageWithNoSessionTokenSupersedesNothing` — an open run recorded with a NULL token stays
  open when another run begins, and a begin with a NULL token closes nothing.

```bash unverified:none of the three tests exists yet — confirm the -run pattern reaches them, and that none of them reports SKIP
cd stats && docker compose up -d && go test ./internal/store/ -run TestBeginStage -count=1 -v
```

- [x] **Step 2: The transaction**

Rework `insertStageRun` to run inside a transaction obtained from the pool: keep the existing
`INSERT ... SELECT` exactly as it is (its `COALESCE` session-id resolution and attempt allocation
are unrelated to this change), then issue the supersede `UPDATE` against the returned row, then
commit. Defer a rollback the same way `CommitHarvestBatch` (`stats/internal/store/harvest.go`)
does — a rollback after a successful commit is a documented no-op in pgx.

The update, in shape:

```sql unverified:confirm the parameter numbering against the final insert statement
UPDATE stage_runs
SET ended_at = $2, outcome = 'superseded'
WHERE session_token = $1
  AND ended_at IS NULL
  AND id <> $3
  AND started_at <= $2
```

`$1` is the new run's session token, `$2` its `started_at`, `$3` its own id. A NULL token matches no
row through `=`, so no separate guard is needed — the same reasoning `insertStageRun`'s doc comment
already records for its session-id subquery.

- [x] **Step 3: Say why, where the next reader will be**

Extend `insertStageRun`'s doc comment (or `BeginStage`'s, whichever the update ends up in) with:
what a begin mark proves about the session's earlier open runs; why the match is on
`session_token` and never `session_id`, which is bound after the fact and routinely NULL at begin
time; why `ended_at` is the successor's `started_at` rather than `now()`; why the
`started_at <= ` guard exists, naming the journal replay it protects; and why the outcome is
`superseded` rather than `abandoned`, naming the rework-rate view that reads `abandoned` directly.

- [x] **Step 4: The whole package, and the tests that begin twice**

```bash verified:the go test half was run at b85d4f5 with the stack up and passed; vet and gofmt are .myflow/project.md's own lint commands
cd stats && go test ./internal/store/... -count=1 && go vet ./internal/store/... && gofmt -l internal/store
```

`TestBeginStageAllocatesAttempts` and `TestConcurrentBeginStageDoesNotCollide` both call
`BeginStage` repeatedly with one input, so they now exercise the supersede path too. Both must pass
unchanged; if either asserts on a run staying open, that assertion is the thing to read carefully
before touching it, not to relax reflexively.

- [x] **Step 5: The full suite**

```bash unverified:only ./internal/harvest/... and ./internal/store/... were run at b85d4f5 — confirm the full suite before the handoff
cd stats && go test ./... -count=1
```

- [x] **Step 6: Pin the rework-rate scenario** *(added by fix round 1, from the per-task review's F1)*

The delta spec's fourth scenario — "A superseded stage is not counted as rework" — was left to
`ReworkRate`'s `COUNT(*) FILTER (WHERE sr.outcome = 'abandoned')` being a literal string match that
cannot accidentally catch `'superseded'`. Structurally true, and pinned by nothing: no test seeds a
superseded run and asserts it stays out of `AbandonedCount`.

Add `TestReworkRateDoesNotCountASupersededRunAsAbandoned` to
`stats/internal/store/aggregate_test.go`, following that file's existing seeding shape: one run
closed `abandoned` and one closed `superseded` in the same period, asserting the row's
`AbandonedCount` counts the first and not the second. Prove it non-vacuous by widening the filter to
`sr.outcome IS NOT NULL` in a scratch copy and watching it fail.

```bash unverified:the test does not exist yet — confirm the -run pattern reaches it and that it does not report SKIP
cd stats && go test ./internal/store/ -run TestReworkRate -count=1 -v
```

- [x] **Step 7: The equal-instant boundary, and the name** *(added by fix round 2, from panel findings F2, F3 and F5)*

Three corrections the panel raised against this task, all in one fixup:

1. **Name the function for what it does.** `insertStageRun` now also closes every open run sharing
   the session token, and its name says only "insert" — a reader reaching it from `BeginStage`'s
   call site sees the name before the doc comment. Rename it to `insertStageRunAndSupersede`.
2. **Pin the `<=` boundary.** `TestBeginStageLeavesAnOpenRunThatStartedLaterAlone` covers a strictly
   *earlier* replay start; nothing covers an *equal* one — which, per step 5, is what a real journal
   replay actually carries. Add `TestBeginStageSupersedesAnOpenRunStartedAtTheSameInstant`: an open
   run and a begin sharing one session token and one `started_at`, asserting the earlier run closes
   as `superseded` with `ended_at` equal to that shared instant. Prove it non-vacuous by narrowing
   the guard to `started_at < $2` in a scratch copy and watching it fail.
3. **Correct two comments this change made false**, in files it did not otherwise touch:
   `stats/internal/api/stages.go`'s `findOpenStageRun` doc comment and
   `stats/internal/api/stages_test.go`'s `TestStageEndClosesHighestOpenAttempt` comment and
   assertion message both say a replayed begin's earlier attempt is left open for the sweeper. It
   is now closed as `superseded` by the replay's own begin. Comment text only — no behaviour, no
   assertion logic, and the test keeps passing because it runs against a fake store with two
   distinct session tokens.

```bash unverified:the test does not exist yet — confirm the -run pattern reaches it and that it does not report SKIP
cd stats && go test ./internal/store/ -run TestBeginStageSupersedes -count=1 -v
```

- [x] **Step 8: Finish the rename across the packages that cite it** *(added by fix round 3)*

Step 7's rename left four references to the old name in a package whose files it was not declared
to touch, so a reader grepping `insertStageRun` now finds prose about a function that does not
exist: `stats/internal/harvest/watcher.go` (two doc comments) and
`stats/internal/harvest/watcher_test.go` (two comments). Update the four to
`insertStageRunAndSupersede`. Comment text only — no behaviour, no assertion, no test logic.

- [x] **Step 9: The exported door, and a test that goes through it** *(added by fix round 4, from panel findings F6 and F12)*

Two gaps the rename left:

1. `BeginStage`'s own doc comment still says it records a stage run and allocates an attempt. It is
   the only entry point any caller or the `StageStore` interface sees, and it now also closes every
   open run sharing the session token. Add a sentence saying so, pointing at
   `insertStageRunAndSupersede`'s comment for the reasoning rather than duplicating it.
2. `TestReworkRateDoesNotCountASupersededRunAsAbandoned` seeds its superseded row by calling
   `EndStage` directly with the outcome string and no session token, so it never exercises the
   supersede write at all — it pins the rework filter and nothing else. Rework it to produce the
   superseded run the real way: a `BeginStage` carrying a session token, then a second `BeginStage`
   on the same token, so the `superseded` string under test is the one the store itself wrote.

**Tests:** `TestBeginStageSupersedesAnEarlierOpenRunOfTheSameSession`,
`TestBeginStageLeavesAnOpenRunThatStartedLaterAlone`,
`TestBeginStageWithNoSessionTokenSupersedesNothing`,
`TestReworkRateDoesNotCountASupersededRunAsAbandoned`,
`TestBeginStageSupersedesAnOpenRunStartedAtTheSameInstant`.

**Regression:** Reverting this task lets a session hold two open stage runs indefinitely — the
orphan keeps reading as "still running" on the dashboard until the 6-hour sweeper reaches it, and
every reader of the windows has to arbitrate an overlap that should not exist. Reverting step 6
alone leaves the rework rate free to start counting superseded runs as abandoned with no test
objecting.

**Baseline:** before=29 after=33 top-level `Test` functions in
`stats/internal/store/stageruns_test.go` (32 after step 1, plus step 7's boundary test); before=20
after=21 in `stats/internal/store/aggregate_test.go`.
<!-- predicted: grep -c '^func Test' on both files after steps 1, 6 and 7 -->

- [x] **Step 10: Say why `MergeMetrics` keeps no guard** *(added by fix round 5, from panel finding F13)*

Task 4 gave `EndStage` an `ended_at IS NULL` guard and explained it at the function.
`MergeMetrics`, its neighbour, still matches on `id` alone — deliberately, because its write is
additive telemetry about work that really happened and refusing it would discard measured usage.
That reasoning currently lives only in a commit message. Put one sentence in `MergeMetrics`'s own
doc comment, so a reader who has just read `EndStage`'s guard sees the asymmetry is chosen rather
than missed.

**Commit:** `fix(2): close the open runs a begin mark supersedes`

---

### 3 Index the supersede UPDATE's own predicate

**Build:** green

**Files:**
- Add: `stats/internal/store/migrations/0009_stage_run_open_session_token.sql`
- Modify: `stats/internal/store/stageruns_test.go`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: the supersede `UPDATE` task 2 added, whose predicate is `session_token = $1 AND
  ended_at IS NULL AND id <> $3 AND started_at <= $2`.
- Produces: a partial index on `stage_runs (session_token)` restricted to `ended_at IS NULL`, and
  nothing else. No column, constraint or row is touched.

*(Added by fix round 2, as the operator's answer to panel finding F3.)* The panel measured the
supersede UPDATE using `stage_runs_ended_at` and then filtering every open row: 0.47ms with 200
concurrent open runs against a 500,200-row table.
<!-- measured: EXPLAIN (ANALYZE, BUFFERS) against a seeded scratch database, reported by the panel's code-review slot on 2026-08-16; the scratch database was dropped afterwards, so this is not re-runnable as written -->
Negligible at this tool's real scale, where a session
holds one or two open runs — but it is the `stage begin` path, which is required never to block or
delay the stage it marks, and the cost grows with the system-wide open-run count rather than with
anything the marking session controls.

The existing `stage_runs_unresolved_session_token` index cannot serve this query: it is partial on
`session_id IS NULL`, a predicate the supersede does not carry.

- [x] **Step 1: The migration**

Add `stats/internal/store/migrations/0009_stage_run_open_session_token.sql`. Migrations are
discovered by `//go:embed migrations/*.sql` (`stats/internal/store/migrations.go`) and applied in
filename order, so adding the file is the whole registration step — there is no list to edit.

```sql unverified:confirm the index name is free and the partial predicate matches the UPDATE's own
CREATE INDEX stage_runs_open_session_token ON stage_runs (session_token)
  WHERE ended_at IS NULL;
```

Carry a header comment in the style of `0008`'s: what the index serves, why the existing
`stage_runs_unresolved_session_token` cannot serve it, and why the partial predicate is exactly the
row set the supersede ever scans — an already-closed run is never a supersede candidate again.

- [x] **Step 2: A test that the index exists**

Add `TestSupersedeIndexExists` to `stats/internal/store/stageruns_test.go`: after `newTestStore`
has migrated a fresh database, query `pg_indexes` for `stage_runs_open_session_token` and assert it
is present. An index is a performance property and a plan is not stable enough to assert on, so
what this pins is that the migration is embedded, applied, and names the index this change agreed
on — the part a later edit could silently drop.

```bash unverified:the test does not exist yet — confirm the -run pattern reaches it and that it does not report SKIP
cd stats && go test ./internal/store/ -run 'TestSupersedeIndexExists|TestMigrationsAreIdempotent' -count=1 -v
```

`TestMigrationsAreIdempotent` reads `EmbeddedMigrationCount()` rather than a hardcoded number, so it
absorbs the new file without an edit — run it to confirm that, rather than assuming it.

- [x] **Step 3: The whole package**

```bash verified:the same command was run at 29affa4 with the stack up and passed; vet and gofmt are .myflow/project.md's own lint commands
cd stats && go test ./internal/store/... -count=1 && go vet ./internal/store/... && gofmt -l internal/store
```

- [x] **Step 4: Say what the index build actually costs** *(added by fix round 4, from panel finding F10)*

The migration's header comment and `design.md`'s Migration Plan both call the migration safe to
apply while the daemon is running, without qualification. `applyMigration`
(`stats/internal/store/migrations.go`) wraps every migration in one transaction, and a plain
`CREATE INDEX` inside a transaction holds a write-blocking lock on `stage_runs` until it commits —
while `CREATE INDEX CONCURRENTLY` cannot run inside a transaction block at all, so the runner's own
design forecloses that escape.

Correct the header comment to say what is true: writes to `stage_runs` block for the length of the
index build, which is proportional to the table — trivial at this tool's scale (a few hundred rows
in the live store) and not trivial for a large one — and that `CONCURRENTLY` is unavailable inside
the transactional runner rather than merely unchosen. Do not change the migration's SQL: making the
runner non-transactional for one migration is a larger change than this finding justifies, and the
lock is measured in milliseconds here. `design.md` carries the same correction, made by the parent.

**Tests:** `TestSupersedeIndexExists`.

**Regression:** Reverting this task returns the supersede UPDATE to filtering the whole open-run set
on every `stage begin`, on a path required never to block or delay the stage it marks.

**Baseline:** before=33 after=34 top-level `Test` functions in
`stats/internal/store/stageruns_test.go`; embedded migrations before=8 after=9.
<!-- predicted: grep -c '^func Test' stats/internal/store/stageruns_test.go, and ls stats/internal/store/migrations/*.sql | wc -l, after steps 1 and 2 -->

- [x] **Step 5: Assert the index, not just its name** *(added by fix round 5, from panel finding F18)*

`TestSupersedeIndexExists` matches `pg_indexes.indexname` and nothing else, so an index rebuilt
under the same name over a different column, or without the `WHERE ended_at IS NULL` predicate that
is this migration's whole point, still passes. Assert the definition: that it indexes
`session_token` and carries the partial predicate. `pg_indexes.indexdef` carries both.

**Commit:** `perf(3): index the open runs a supersede scans`

---

### 4 Make the supersede hold under concurrency, and stop a late end mark undoing it

**Build:** green

**Files:**
- Modify: `stats/internal/store/stageruns.go`
- Modify: `stats/internal/store/stageruns_test.go`
- Modify: `stats/internal/api/stages.go`
- Modify: `stats/internal/api/stages_test.go`
- Modify: `stats/internal/sweep/sweep_test.go` — `EndStage`'s guard is store-wide, not scoped to the
  supersede path, so it also decides a live `stage end` racing the abandoned-run sweeper:
  previously the live call silently overwrote whatever the sweep had written, and now it reports
  `ErrStageRunAlreadyClosed`. `TestSweepRacingLiveEndDoesNotCorruptOutcome` pinned the old
  behaviour and is updated to assert the stored outcome matches whichever call reported winning —
  a **stronger** assertion than the "either outcome is acceptable" one it replaces, not a relaxed
  one.

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: `insertStageRunAndSupersede`'s transaction and `EndStage`'s single-row UPDATE, both as
  task 2 left them.
- Produces: a per-session serialisation point inside that transaction, and an `EndStage` that
  refuses to reopen a run something else already closed.

*(Added by fix round 4, from panel pass 2's findings F8 and F9.)* Task 2 made the supersede atomic
with the insert, which is not the same as making it correct against a *concurrent* begin. Two
findings, both demonstrated by hand against a real database rather than argued on paper:

- **Two begins for one session token can both stay open.** Under READ COMMITTED — the pool's
  default, which the transaction sets no level against — neither transaction sees the other's
  uncommitted insert, so neither supersedes it. The live path for this is the reconciler replaying
  a journalled begin while a live begin runs, which is exactly the situation the journal exists to
  produce.
- **A late end mark resurrects a superseded run.** `ApplyEndStageMark` resolves an open run and
  then closes it in two separate, non-transactional calls, and `EndStage`'s UPDATE matches on `id`
  alone. A supersede landing in that gap is silently overwritten, and the session is back to two
  overlapping windows — the state this whole change exists to prevent.

- [x] **Step 1: The failing tests, written first**

In `stats/internal/store/stageruns_test.go`:

- `TestConcurrentBeginStageForOneSessionLeavesOneOpenRun` — N goroutines calling `BeginStage` with
  the **same** session token and increasing `StartedAt`, asserting exactly one run is left open
  afterwards and every other is `superseded`. Follow `TestConcurrentBeginStageDoesNotCollide`'s own
  shape, which already drives concurrent begins — it does not exercise this path because it leaves
  `SessionToken` nil, and a NULL token matches no row.
- `TestEndStageRefusesARunAlreadyClosed` — close a run, then call `EndStage` on it again and assert
  the second call neither changes `ended_at`/`outcome` nor reports success.

In `stats/internal/api/stages_test.go`, against the package's existing fake store: an end mark whose
target was closed between the lookup and the write is reported as the definitive refusal
`ErrNoOpenStageRun`, not as a success.

- [x] **Step 2: Serialise the supersede per session**

Take `pg_advisory_xact_lock(hashtext($token))` as the transaction's first statement in
`insertStageRunAndSupersede`, and **only when the session token is non-empty** — a mark without one
supersedes nothing, so it has nothing to serialise against and must not queue behind an unrelated
session. The lock is transaction-scoped, so it releases on commit or rollback with no unlock path to
forget. `stats/internal/store/migrations.go` already uses a Postgres advisory lock for the migration
runner, so this is the store's existing mechanism rather than a new one.

- [x] **Step 3: Guard the close**

`EndStage`'s UPDATE gains `AND ended_at IS NULL`. A row that exists but is already closed is a
distinct outcome from a row that does not exist: return a new `ErrStageRunAlreadyClosed` for it, and
have `ApplyEndStageMark` translate that into the `ErrNoOpenStageRun` it already treats as a
definitive, loudly-logged refusal — the mark cannot attach to anything, which is exactly what that
error means and what the journal already knows how to retire.

**`MergeMetrics` is deliberately left matching on `id` alone.** Its write is additive telemetry
about work that really happened; landing it on a run that has since closed records the truth, where
refusing it would discard measured usage. Only `ended_at` and `outcome` — the fields that say
whether the window is open — need the guard.

- [x] **Step 4: The packages**

```bash unverified:the three new tests do not exist yet — confirm each is reached and that none reports SKIP
cd stats && go test ./internal/store/... ./internal/api/... -race -count=1
```

**Tests:** `TestConcurrentBeginStageForOneSessionLeavesOneOpenRun`,
`TestEndStageRefusesARunAlreadyClosed`, and the `stages_test.go` case above.

**Regression:** Reverting this task lets two open windows exist for one session again — by a
concurrent begin that supersedes nothing, or by an end mark that reopens what a supersede closed —
which is the precondition for the misattribution KAN-185 exists to fix.

**Baseline:** before=34 after=36 top-level `Test` functions in
`stats/internal/store/stageruns_test.go`.
<!-- predicted: grep -c '^func Test' stats/internal/store/stageruns_test.go after step 1 -->

- [x] **Step 5: Close three seams around the lock** *(added by fix round 5, from panel findings F16, F17 and F19)*

1. **An empty-string token is not a token.** The lock is taken only for a non-empty token, but the
   supersede `UPDATE` matches `session_token = $1`, and SQL `=` happily matches `''` against other
   empty-string rows — so a caller passing a non-nil empty string would supersede unserialised,
   which is the race the lock exists to close. No shipped path can do it (`validateSessionTokenShape`
   rejects `""` on both the HTTP and reconcile paths), but the store carries no such invariant of
   its own. Normalise an empty-string token to NULL at the top of `insertStageRunAndSupersede`, so
   the lock guard and the UPDATE cannot disagree about what counts as a token.
2. **Do not share the migration runner's lock keyspace.** `pg_advisory_xact_lock(hashtext($1))`
   takes a pseudo-random key in the same single-bigint space as `migrations.go`'s fixed
   `pg_advisory_lock(725016001)`, whose own comment states its only requirement is not colliding
   with another subsystem's key. Use the **two-argument** form — a fixed namespace int of this
   change's choosing plus `hashtext($1)` — which Postgres keeps in a separate keyspace from
   single-bigint locks entirely, so the collision cannot occur rather than being unlikely. Say that
   in the comment.
3. **Build the refusal error once.** `ApplyEndStageMark` now constructs the identical five-argument
   `ErrNoOpenStageRun` wrap twice. Extract it.

**Commit:** `fix(4): serialise the supersede and stop a late end mark reopening it`
