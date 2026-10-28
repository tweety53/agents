## Context

One spectre change for the seven non-`big` findings of the self-review pass of
`kan-1041-weekly-dashboard-model-effort-correlation`
(`.superpowers/sdd/self-review-pass.md` in this worktree). The fixes ride the change's own branch
`spectre/kan-1041-weekly-dashboard-model-effort-correlation` in this worktree, from the HEAD it
held when the pass started (`52eee152`). Every finding is carried by exactly one task; the
brainstorm record is
`.superpowers/sdd/2026-10-28-self-review-kan-1041-weekly-dashboard-model-effort-correlation-design.md`.

Facts settled at plan time, each by a run at `52eee152`:

- **Finding 1** (blast radius 3): the pass's minimal reproduction fed to the hook prints
  `"permissionDecision": "deny"`, and with Task 1's diff applied it prints nothing (allow). The
  new test fails at `52eee152` with `wanted allow, got deny`. The archived narrative's misattributed
  line is `narrative.md` line 16, not 1081 as the pass report says; its correction rides this plan
  commit, since a fixer never writes `spectre/changes/`.
- **Finding 2** (blast radius about 6): the repair commands already exist — `flowd rekey-dispatch
  <id> <key>` and `flowd retime-dispatch <id> <started> <ended> <commit>`
  (`stats/cmd/flowd/main.go`) — so no new command is added. The ledger's Dispatch 35 and 42 are
  store ids 8334 (`panel-fix-5`, 02:45:00→02:48:46, commit `21b73d94…`) and 8344 (`panel-fix-7`,
  03:02:38→03:05:29, commit `3e7e8429…`); round 5's latest reviewer end is 02:47:03 and round 7's
  03:04:14; no row of the change holds `panel-fix-6` or `panel-fix-8`. Pass-log note 2535 reads
  `fix round re-run of mutation over panel-fix-5`.
  <!-- measured: docker exec flow-postgres psql -U flow -d flow -tAc "select id, dispatch_key, started_at, ended_at, commit_sha from dispatches …" and "select id, note from panel_passes …" @ branch spectre/kan-1041-weekly-dashboard-model-effort-correlation (live dev store, read-only) -->
- **Finding 3** (blast radius 2): Dispatch 38 is store id 8340 (`task-13-implementer`,
  02:48:46→03:02:22, commit `8808872e…`); `verify-fix-3` (8336) ends 03:02:17. The new test fails
  at `52eee152` with the begin's `startedAt` at the commit-derived 02:48:46.
- **Finding 6** (blast radius 2): the new dark-page test fails at `52eee152` with
  `surface_top 200, want 0`.
- **Finding 7** (blast radius 3): this machine has no host `psql`/`pg_dump`/`createdb` client —
  every repository script reaches the store through `docker exec flow-postgres`
  (`stats/internal/guard/workspace.go`); `flowd` writes its own pidfile at
  `$TMPDIR/flowd-<port>.pid`, pid on line 1 and executable on line 2
  (`stats/internal/pidfile/pidfile.go`); `scripts/run-guard-tests.sh` runs every
  `scripts/test-*.sh` by glob.
- **Findings 4 and 5** (blast radius 1 each): `zsh -c 'false | tail -1; echo "P=${PIPESTATUS[0]}"'`
  prints `P=`; `check-plan-provenance.sh` lives in `<agents repo>/scripts/`, not among
  `skills/flow/scripts/`. `scripts/check-verbatim-moves.sh 52eee152` prints the four reworded
  sentences as FAIL until this change's `verbatim-moves.txt` lists them, then `0 violation(s)`.

## Decisions

### Strip only bodies that run nothing

**ID:** rest-drops-data-bodies
**Status:** active
**Chosen:** `drop_data_bodies(lines)` removes from a body's rest every other heredoc body with a
quoted delimiter, or an unquoted one holding no `$(` or backtick, keeping its opener and delimiter
lines; `strip_data_heredocs` builds both of its rests (`cat`/`tee` and `flow` openers) through it.
**Considered:** stripping every other body — an unquoted body's `$(bash x.sh)` runs the first
heredoc's file, and its removal would read that file as data (the test's second case denies it);
judging each body against the original command only — the defect itself.

### Refuse a round-low panel-fix key at its end

**ID:** panel-fix-round-refusal
**Status:** active
**Chosen:** `panelFixKeyedARoundLow` refuses `panel-fix-<r>[-<n>]` (exit 2, message naming
`panel-fix-<r+1>`) when a reviewer row of round `r` started before the fix's end — the parsed or
derived `ended` on `backfill`, now on a live `begin`; a back-filled `begin` (its `-started-at` set)
skips the live check, since its `backfill` already ran it. A failed record read passes. To run on
every back-fill, `backfill` resolves its connection flags and project key before the inline branch.
**Considered:** judging against the fix's start — `panel-fix-7`'s derived start came before round
7's reviewers began only because the low key had floored it at round 6, so the start cannot catch
the key that caused it; renumbering automatically — a silent rewrite of a key the caller chose.

### Repair through the existing commands

**ID:** repair-existing-commands
**Status:** active
**Chosen:** `flowd rekey-dispatch` and `flowd retime-dispatch` repair rows 8334, 8344 and 8340;
note 2535 is corrected by one `UPDATE … WHERE id = 2535 AND note = '<old note>'` through
`docker exec flow-postgres psql`; then `flow record render -kind all` re-renders the ledger and the
panel record from the store and both are copied over the archived `ledger.md` and `panel.md`.
**Considered:** a new combined re-key-and-retime command — both halves exist; a `flowd` command for
one pass-log note — a one-row, one-time correction the guarded statement states in full.

### Archived records ride the fix commit

**ID:** archived-records-in-fix
**Status:** active
**Chosen:** Task 3's commit carries the re-rendered `spectre/changes/archive/kan-1041-weekly-dashboard-model-effort-correlation/ledger.md`
and `panel.md` — the one exception to "never write `spectre/changes/`", which guards the live
change's ticks; the archived narrative's correction, known at plan time, rides the plan commit.
**Considered:** leaving the archived records stale — the forward-only fix the self-review contract
forbids; re-rendering in the plan commit — the store is not yet repaired then.

### Inline floor among the change's rows

**ID:** inline-floor
**Status:** active
**Chosen:** `inlineFloor` is the latest `EndedAt` not after the commit's time among the change's
other `agent_id = inline` rows (any key but the one being back-filled), applied after
`panelFixFloor`; a failed record read sets no floor.
**Considered:** flooring per session — the run record carries no session per row and every inline
row of a change is the same parent's main thread; refusing an overlapping back-fill — leaves the
row unrecorded.

### Dark page is the whole surface

**ID:** dark-page-surface
**Status:** active
**Chosen:** `t = 0` when the walk moved no row (`y == r.h-1`), alongside the existing `t < 60`.
**Considered:** a theme flag — the image already says it.

### The live copy is a bash tool

**ID:** live-copy-bash
**Status:** active
**Chosen:** `scripts/live-store-copy.sh up <name> [<port>] | down <name>`, bash: `pg_dump | psql`
inside `flow-postgres`, `flowd` built from the checkout and started with `TMPDIR=/tmp` so its own
pidfile is the state; port default 4175, 4173 and 4174 refused; `down` signals a pid only when its
pidfile's executable is this script's binary, and drops only `flow_<name>_live`; named in
`.flow/project.md`'s `## run`. The harness stubs `docker`, `make`, `go` and `curl`.
**Considered:** a `flow-guard` subcommand — the script judges nothing (the new-guards-in-Go rule
binds guard logic), the same kind of orchestration as `ui-test-up`; reusing `workspace.sh` —
its `remove --force` is limited to `*_uitest` databases; a derived per-worktree port — a port the
caller passes is simpler and the default stays clear of both protected ones.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — the main-checkout protection hook decides what a session may write, the
record CLI gains a refusal every panel-fix row passes, and the new script drops a database.
**Considered:** `flow-low` — every diff is pre-applied and test-pinned, but a hole in the hook or a
wrong drop target is a high-cost miss.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups, in series, each on `flow-low`/`opus` —
`self-review-kan-1041-weekly-dashboard-model-effort-correlation-fix-g1`: Tasks 1, 4, 5, 6, 7 (the
hook, the two skill edits, the matrix and the live-copy script — file-disjoint diffs);
`-fix-g2`: Tasks 2, 3 (the record CLI and the dev-store repairs with the archived re-render).
**Considered:** one fixer for all seven — the store repairs and render in the same context as five
unrelated edits; one per task — seven dispatches where two coherent groups serve.

## Open questions

none

## Recommended approach

Apply each task's pre-verified diff in order, prove each new test fails without its fix, and run
the store repairs once Task 3's code has landed, then re-render the archived records. Rejected
alternatives are recorded under `rest-drops-data-bodies`, `panel-fix-round-refusal`,
`repair-existing-commands`, `archived-records-in-fix`, `inline-floor`, `dark-page-surface` and
`live-copy-bash`.

## Critical files

- `hooks/protect-main-checkout.py` — `strip_data_heredocs` and the new `drop_data_bodies`.
- `stats/internal/repotest/protect_main_checkout_test.go` — the two-heredoc test.
- `stats/cmd/flow/record.go` — `runRecordDispatchBegin`, `runRecordDispatchBackfill`, the new
  `panelFixKeyedARoundLow` and `inlineFloor`.
- `stats/cmd/flow/record_test.go` — the round-low and inline-floor tests.
- `skills/flow/review-panel.md`, `skills/flow/document-fix.md` — the two reworded sentences.
- `stats/internal/guard/measurevisualproperties.go`, `stats/internal/guard/measure_visual_properties_test.go`
  — `measureElements` and the dark-page test.
- `scripts/live-store-copy.sh`, `scripts/test-live-store-copy.sh`, `.flow/project.md` — the new
  tool, its harness and its `## run` paragraph.
- `spectre/changes/archive/kan-1041-weekly-dashboard-model-effort-correlation/{narrative,ledger,panel}.md`
  — the archived records the defects left wrong.

## Existing code to reuse

- `panelFixFloor`, `callRecord`, `records.PanelDispatchRound` (`stats/cmd/flow/record.go`) — the
  run-record fetch and round parse the two new helpers copy.
- `TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers` (`stats/cmd/flow/record_test.go`)
  — the `genuineDaemon` fake-store pattern both new tests copy.
- `flowd rekey-dispatch`, `flowd retime-dispatch` (`stats/cmd/flowd/main.go`) — the row repairs.
- `flow record render -kind all` — the ledger and panel render.
- `ui-test-up` / `ui-test-down` (`stats/Makefile`) — the build, pinned `TMPDIR`, readiness poll
  and bounded stop the live-copy script follows.
- `scripts/test-prompt-tuning-line.sh` — the bash harness shape.

## Verification

- `cd stats && go test ./cmd/flow ./internal/guard -count=1` and
  `go test ./internal/repotest -run TestProtectMainCheckout -count=1` pass; `gofmt -l .` and
  `go vet ./...` are clean.
- `scripts/test-live-store-copy.sh` prints `all cases passed`; `scripts/check-verbatim-moves.sh`,
  `scripts/check-references.sh` and `scripts/check-markdown-integrity.sh` exit 0.
- Feeding the pass report's minimal reproduction to `hooks/protect-main-checkout.py` prints no
  decision (allow).
- After the repairs, `docker exec flow-postgres psql -U flow -d flow -tAc "select id, dispatch_key,
  started_at from dispatches where id in (8334,8340,8344) order by id"` prints `panel-fix-6` at
  02:47:03, `task-13-implementer` at 03:02:17 and `panel-fix-8` at 03:04:14, and the archived
  `ledger.md` carries `Key: panel-fix-6` and `Key: panel-fix-8` and no `panel-fix-5` or
  `panel-fix-7`.

## Live check

The repairs in Tasks 2 and 3 are themselves run against the live dev store, and their selects are
the check: before, ids 8334/8344 read `panel-fix-5`/`panel-fix-7` starting 02:45:00/03:02:38 and
8340 starts 02:48:46; after, the Verification select above. Unchanged keys or starts after the
repair commands report success mean the commands wrote another database — `FLOWD_DSN` was not the
dev `flow` database.
