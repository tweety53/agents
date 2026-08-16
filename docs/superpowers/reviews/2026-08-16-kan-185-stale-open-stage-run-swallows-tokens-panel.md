# Final review panel — kan-185-stale-open-stage-run-swallows-tokens

Roster: `light`. Panel model: sonnet. Panel-fix model: sonnet.

Diff under review: `.superpowers/sdd/final-review.diff` — `git diff b85d4f5`, 400 changed lines
across 5 files.

`scripts/check-panel-diff-size.sh` measured **400** changed lines against a cap of **2000** — exit
0, under cap. No operator question was reached.

## Slots

| # | Slot | Status | Model |
|---|------|--------|-------|
| 0 | Primary — plan alignment + code quality | ran | sonnet |
| 2 | Principles — Merged lens | ran | sonnet |
| 3 | Code review (low) | ran | sonnet |
| 4 | Security | declined by the operator | — |
| 5 | Adversarial | declined by the operator | — |
| 6 | Lens B — simplicity & state | declined by the operator | — |
| 6 | Lens C — robustness & ops | declined by the operator | — |

Four optional triggers fired against the diff: Security (query construction — task 2 adds a SQL
UPDATE to the begin-mark path), Adversarial (behaviour change to code with existing tests, and >~300
changed lines), Lens B (>~200 changed lines) and Lens C (error handling — the new transaction's
begin/exec/commit/rollback failure paths). Under the `light` preset all four went to the operator as
one multi-select, and the operator selected none. All four are **declined**, distinct from a slot
whose trigger never fired.

## Pass 1 — findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Primary | Minor | `stats/internal/harvest/attribute.go:347` | the new doc comment says a replayed begin's first attempt stays open until the sweeper, which task 2's own supersede makes false |
| F2 | Principles | Minor | `stats/internal/store/stageruns.go:273` | `insertStageRun` also closes every open run sharing the session token, and its name says only "insert" |
| F3 | Code review (low) | Minor | `stats/internal/store/stageruns.go:244` | no index covers the supersede UPDATE's `session_token` predicate, so it filters the open-run set linearly — measured 0.47ms at 200 concurrent open runs |
| F4 | Code review (low) | Minor | `stats/internal/harvest/attribute.go:351` | "a replayed attempt always starts after the one it replays" is wrong: the journal replays the original request's own start instant, so the two are equal |
| F5 | Code review (low) | Minor | `stats/internal/store/stageruns_test.go` | no test covers the supersede guard's equal-`started_at` boundary, which per F4 is the real replay shape |

findings-total: 24
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed

reproducers-total: 24
finding-reproducer: F1 none — documentation accuracy, no runtime defect to demonstrate
finding-reproducer: F2 none — a naming judgment, not a runnable check
finding-reproducer: F3 none — measured by hand with EXPLAIN ANALYZE against a seeded scratch database; the slot's own command carries shell metacharacters and a container path, so it is not runnable under the reproducer shape rules
finding-reproducer: F4 none — documentation accuracy, no runtime defect to demonstrate
finding-reproducer: F5 none — a coverage gap; the missing test is the demonstration
finding-reproducer: F6 none — a documentation gap on an exported symbol, not a runnable check
finding-reproducer: F7 none — test-narrative accuracy; the test passes and asserts the right outcome
finding-reproducer: F8 none — requires interleaving two database transactions by hand to force the visibility gap; not a single runnable command
finding-reproducer: F9 none — requires a live race between an in-flight end mark and a replayed begin for the same session token
finding-reproducer: F10 none — demonstrating the lock needs two concurrent database sessions; the transactional runner itself is visible in migrations.go
finding-reproducer: F11 none — documentation accuracy, no runtime defect to demonstrate
finding-reproducer: F12 none — a coverage gap; the reworked test is the demonstration
finding-reproducer: F13 none — documentation of a deliberate asymmetry, not a runnable check
finding-reproducer: F14 git log -1 --format=%B 4bc5597
finding-reproducer: F15 none — a coverage gap; the missing discriminating test is the demonstration
finding-reproducer: F16 none — requires a direct store call passing a non-nil empty session token, which every shipped entry point rejects first
finding-reproducer: F17 none — requires a session token whose hash collides with the migration runner's fixed key, contending against a live migration
finding-reproducer: F18 none — unverifiable: the slot supplied a shell command line, which the shape guard refuses; it is recorded as given in the prose below rather than rewritten to pass
finding-reproducer: F19 grep -n ErrNoOpenStageRun stats/internal/api/stages.go
finding-reproducer: F20 none — documentation consistency, no runtime defect
finding-reproducer: F21 none — documentation consistency, no runtime defect
finding-reproducer: F22 git log -1 --format=%B 0c13f3c
finding-reproducer: F23 none — documentation consistency, no runtime defect
finding-reproducer: F24 none — a stale worked example in a doc comment, read against the writer count in the test it cites

## Pass 1 — what each slot verified

**Primary** reported the code itself clean: two commits matching `tasks.md`, no file touched outside
the two packages each task declares, no migration, no API/CLI/SPA/TypeScript change, and all seven
delta-spec scenarios mapped to a named test that runs rather than skips. It additionally ran
`EXPLAIN` against the supersede UPDATE and found it using the existing `stage_runs_ended_at` index
rather than a sequential scan.

**Principles** (Merged lens) found no Critical and no Important finding: the transactional merge of
insert and supersede is a deliberate atomicity requirement documented at the function, not an SRP
violation, and no hard invariant, suppression or lint-config weakening appears in the diff. The
project's standards files resolved to `CLAUDE.md` and `AGENTS.md`, which carry the lint-fix-priority
rule and an empty project-specific-standards placeholder.

**Code review (low)** ran the store and harvest suites under `-race` — 170 pass, 0 skip, 0 fail —
and traced the replay path through `stats/cmd/myflow/stage.go` and
`stats/internal/reconcile/reconcile.go` to establish F4, which is the finding that also justifies
F5 and reshapes the delta spec's replayed-begin scenario.

No pass bounced a finding.

## Fix round 2

Mode: one fix subagent on sonnet, handed F1, F2, F4 and F5 as the combined list. F3 was not
dispatched with them — its only real fix is a schema migration, which this change's own global
constraint excluded, so it went to the operator instead of to the fixer.

`scripts/check-panel-reproducers.sh` passed before the dispatch: five findings declared, every one
carrying the `none` exemption form with its reason, since all five are documentation, naming,
coverage or measured-performance findings with no runnable failing check between them. No finding
was bounced.

Fixed, folded into the two task commits by `git commit --fixup` and `git rebase --autosquash` —
task 1 became `3660a7e`, task 2 became `aec86df`, still two commits on the merge base:

- **F1, F4** — `bestWindow`'s replay paragraph rewritten: the earlier attempt is closed as
  `superseded` by the replay's own begin rather than left for the sweeper, and a replay starts at
  the *same* instant as the attempt it replays rather than after it. The fixer verified both claims
  at their source before writing, in `stats/cmd/myflow/stage.go` and
  `stats/internal/reconcile/reconcile.go`. The same correction was carried into
  `stats/internal/api/stages.go` and `stats/internal/api/stages_test.go`, which repeated the
  now-false sweeper claim.
- **F2** — `insertStageRun` renamed to `insertStageRunAndSupersede`.
- **F5** — `TestBeginStageSupersedesAnOpenRunStartedAtTheSameInstant` added, proven non-vacuous by
  narrowing the guard to a strict `<` and watching it fail.

The parent carried the same F4 correction into the planning artifacts, which no fix subagent may
touch: `design.md`'s `latest-start-wins` decision and the delta spec's replayed-begin scenario both
said a replay starts later, and both now say it starts at the same instant.

## Fix round 3 — F3, by the operator's decision

Put to the operator as a finding whose only fix contradicted the plan's own no-migration
constraint. The operator chose to **take another round and add the index** rather than withdraw it.

That answer made F3 real work rather than a note, so it became **task 3** in the plan — migration
`0009_stage_run_open_session_token.sql`, a partial index on `stage_runs (session_token)` restricted
to `ended_at IS NULL`, with a test that the migration is embedded, applied, and names that index.
`proposal.md`, `design.md`'s Migration Plan and the plan's global constraints were amended to say
this change ships one additive migration.

Fix round 3 also carried step 8: four references to the pre-rename `insertStageRun` left in
`stats/internal/harvest/watcher.go` and `watcher_test.go`, which fix round 2 correctly declined to
touch because they sat outside the files task 2 declared.

## Pass 2 — full re-run

Escalated to Full automatically: the fixes altered a delta spec and added a migration, and three
fix rounds had already run. All three required slots re-ran against a rewritten `final-review.diff`
(615 changed lines, 10 files, three commits). The four conditional slots stay **declined** — the
operator's pass-1 answer holds for the change, and no slot is re-offered.

`scripts/check-panel-diff-size.sh` measured **615** against the cap of **2000** — exit 0.

Every pass-1 finding was independently re-verified as fixed by the slot that raised it, including a
fresh `EXPLAIN (ANALYZE, BUFFERS)` for F3 against a seeded scratch database: the supersede UPDATE
now runs an index scan on `stage_runs_open_session_token` rather than reaching through
`stage_runs_ended_at` and filtering every open row.

**Each slot was told to number new findings from F6 so they could not collide with pass 1 — and
all three then started at F6, so they collided with each other.** The table below is the
reconciled namespace; the slot column says who raised each one. That is a defect in how the
dispatch was written, recorded here rather than quietly renumbered.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F6 | Principles | Important | `stats/internal/store/stageruns.go:107` | `BeginStage` — the only entry point callers and the `StageStore` interface see — still documents itself as recording a run and allocating an attempt, with no mention that it now closes what it supersedes |
| F7 | Primary | Minor | `stats/internal/harvest/attribute_test.go:474` | `TestReplayedBeginPrefersHighestAttempt`'s comment frames two windows ten minutes apart as a replay, which the corrected understanding says a replay cannot be |
| F8 | Primary | Minor | `stats/internal/store/stageruns.go:189` | under READ COMMITTED two concurrent begins for one session token each miss the other's uncommitted insert, so neither supersedes and both stay open |
| F9 | Code review (low) | Major | `stats/internal/store/stageruns.go:279` | `EndStage` matches on `id` alone, so a supersede landing between `ApplyEndStageMark`'s lookup and its write is silently overwritten and the closed run reopens |
| F10 | Code review (low) | Major | `stats/internal/store/migrations/0009_stage_run_open_session_token.sql` | a plain `CREATE INDEX` inside the runner's transaction holds a write-blocking lock, which the migration's own comment and design.md both denied by omission |
| F11 | Code review (low) | Minor | `stats/internal/harvest/attribute.go:356` | the same-instant tie-break is called the ordinary replay path, but the supersede sets `ended_at = started_at`, and `Window.contains` never matches an empty interval — so that path cannot reach the tie-break at all |
| F12 | Code review (low) | Minor | `stats/internal/store/aggregate_test.go:652` | the rework-rate test seeds its superseded row through `EndStage` with no session token, so it never exercises the supersede write it exists to cover |

Two of the slot's raw findings were dropped by the slot itself as already litigated in design.md,
and its own report says which — the session-token-only supersede scope, and the absence of a unique
constraint. One further pass-2 claim, that a delayed replay leaves a stray open row nothing
supersedes, was examined and folded into F8 rather than carried separately: it is the same
visibility gap seen from the other end, and the same serialisation closes both.

## Pass 3 — full re-run

All three required slots re-ran against the 982-line diff (four commits). Every pass-2 finding was
re-verified as fixed by the slot that raised it, each doing its own work rather than reading this
record: the advisory lock reverted in a scratch copy to watch the concurrency test fail, the
`EndStage` call sites grepped for unhandled errors, `Window.contains` checked against an empty
interval, and the rewritten sweeper-race assertion judged on its merits — all three slots
independently called it **stronger** than the one it replaced, which is also what the parent found
reading the diff directly.

Nine new findings. **The F-number collision repeated**: all three slots were again told to number
from the next free id and all three started at F13, so the table below is the reconciled namespace.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F13 | Principles | Minor | `stats/internal/store/stageruns.go:375` | `MergeMetrics`'s deliberate exemption from `EndStage`'s new guard is explained only in a commit message, in a diff that annotates every other asymmetric decision at the code |
| F14 | Primary | Minor | commit `4bc5597`'s message | still says a replay's second attempt starts after the first — the claim corrected in the code, the design and the spec, now permanently contradicting them from the history |
| F15 | Code review (low) | **Major** | `stats/internal/harvest/attribute.go:380` | no test discriminates start-time-before-attempt *precedence*: all three tie-break tests pass under an attempt-first comparator, so reordering the clauses would reopen the KAN-185 incident with nothing failing |
| F16 | Code review (low) | Minor | `stats/internal/store/stageruns.go:245` | the advisory lock is skipped for a non-nil empty session token while the supersede UPDATE's `=` still matches it against other empty-string rows |
| F17 | Code review (low) | Minor | `stats/internal/store/stageruns.go:246` | `hashtext` takes pseudo-random keys in the same single-bigint advisory-lock space as the migration runner's fixed key |
| F18 | Code review (low) | Minor | `stats/internal/store/stageruns_test.go:1684` | the index test asserts the index's name and nothing else — not its column, not the partial predicate that is the migration's whole point |
| F19 | Code review (low) | Minor | `stats/internal/api/stages.go:551` | the same five-argument `ErrNoOpenStageRun` wrap is now built twice in one function |
| F20 | Code review (low) | Minor | `design.md` `latest-start-wins` | still called the same-instant tie-break "the ordinary shape of a replay" — the parent's own round-2 text, which round 4 corrected in the code and not here |
| F21 | Code review (low) | Minor | delta spec, rework scenario | the scenario's title claimed more than its THEN asserts: the rework count is `attempt > 1` with no outcome filter, so a superseded second attempt does count as rework |

F20 and F21 are planning artifacts, which no fix subagent may touch; the parent fixed both directly
and they are recorded fixed above. The remaining seven went to fix round 5.

The slot dropped eight further candidates from the `code-review` skill's raw output and gave a
reason for each — self-refuted by the report's own conclusion, speculative, or a convention the
codebase already followed before this diff. That filtering is the slot doing its job, and is
recorded here so a later reader can see what was considered and rejected rather than missed.

**One reproducer was refused by `scripts/check-panel-reproducers.sh` and is recorded unverifiable
rather than rewritten to pass it.** F18's slot supplied `cd stats && go test ./internal/store/ -run
TestSupersedeIndexExists -v`, a shell command line; the guard accepts a bare path with plain
arguments and nothing else, precisely so a line that could carry an injection is never quietly
reshaped until it is accepted. The command is written here for a human to run, the finding was
dispatched to fix round 5 on its merits, and the operator is told it was refused — the disposition
the contract requires, not a silent substitution.

## Fix round 5

One fix subagent on sonnet took all seven code findings; the parent had already fixed F20 and F21
in the planning artifacts. Each fix folded into its own task's commit, rewriting the branch to
`c29600b`, `3038570`, `7eefb0e`, `0c13f3c` — still four commits, one per task.

- **F15**, the Major, is the one worth recording in full. The fixer added
  `TestLatestStartOutranksAHigherAttempt` — a window with the earlier start and the *higher* attempt
  against one with the later start and the *lower* attempt — then proved it discriminates by
  reversing `bestWindow`'s two clauses in a scratch copy: only the new test failed, and the three
  existing tie-break tests all passed under the reversed comparator, exactly as the slot predicted.
  That is the blind spot closed and demonstrated rather than asserted.
- **F17** moved the advisory lock to the two-argument `pg_advisory_xact_lock(namespace, hashtext)`
  form under a named constant, which Postgres keeps in a keyspace it never compares against the
  migration runner's single-bigint lock — the collision is now impossible rather than improbable —
  and the concurrency test was re-run five times under `-race` to confirm serialisation still holds.
- **F14** required rewriting the first of four commits' message. Done with a detached-HEAD amend and
  `git rebase --onto` rather than an interactive rebase, which this harness does not offer; the tree
  diff across the reword was empty, so only the message changed.
- **F13**, **F16**, **F18** and **F19** landed as described in their rows.

## Pass 4

Dispatched against the current four commits, at the operator's explicit choice: the handoff bar
requires a clean result that is not stale, and five fix rounds had passed since any slot last read
the diff.

### Pass 4 — result

**Principles: clean, no findings at any severity.** It verified F13 fixed at the code and judged fix
round 5's additions against the one architectural boundary this codebase actually states — `store`
owns all SQL — including the index test's direct `pg_indexes` query, which it accepted as a
documented, precedented test-only exception rather than a new violation.

**Primary** confirmed F14 and all six of pass 3's other fixes hold, each by its own check rather
than by reading this record, and established that the task-1 reword changed the message alone by
diffing the tree across it. It raised two.

**Code review (low)** confirmed F15 by repeating the comparator-swap experiment itself and raised
one. It also named a candidate it declined to raise, and why — the empty-token normalisation has no
exercising test, but no shipped path can reach it, so it is defensive code rather than a live gap.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F22 | Primary | Minor | commit `0c13f3c`'s message | still describes the single-argument advisory lock that fix round 5 replaced, and omits the empty-token normalisation from the same round — the F14 defect class, on task 4 |
| F23 | Primary | Minor | delta spec, replayed-begin scenario | still described the first attempt as *still open* at the replay's instant, which is exactly the state the supersede prevents; `design.md` got this correction in round 4 and the spec did not |
| F24 | Code review (low) | Minor | `stats/internal/store/stageruns.go:51` | `maxAttemptRetries`'s doc comment works its arithmetic against 20 writers, and task 4 raised that test to 30 |

**The F-number collision happened a third time**, across all four passes now: told to number from the
next free id, two slots again chose the same one. The reconciliation here is the parent's, and the
lesson belongs in the dispatch rather than in the reviewers — a prompt that hands each slot its own
reserved range would end it, and this record is the evidence for making that change.

F23 was the parent's to fix and is recorded fixed. F22 and F24 went to fix round 6, which changes no
code at all: a commit message and a stale worked example in a doc comment. **That is why pass 4's
clean verdict on the code still stands after it** — no slot re-reads the branch after round 6,
and nothing round 6 touches is code a slot examined.
