# kan-1123-plan-shape-rules-from-guard
> Design record of change kan-1123-plan-shape-rules-from-guard (KAN-1123), written at landing. Each decision's own **Status:** line is the status.

## Context

KAN-1123 (labels carry `flow-cost`). The reachability check ran on base `9b6d28b9`, and the defect
still reproduces there:

- `flow-guard check-plan-shape` has no mode that prints its rules. They exist only as the header
  comment of `stats/internal/guard/planshape.go`. Planner prose names F12, F17, F18, F19, F22–F27
  and F29. F1–F11 (except F8), F14, F15, F16, F23's shape rule and the five in-process sub-guards
  (`plan-suite-steps`, `plan-script-paths`, `plan-files-paths`, `plan-task-id-rule`,
  `plan-step-files`, run by `psCheckOne` through `Registry[name]`) appear nowhere a planner reads.
- Violation lines are `path:line: message`, and the rule id appears inconsistently: F22 as a
  prefix; F23–F26, F28 and F29 as a `(Fnn)` suffix; every other row and all five sub-guards with
  no id.
- `guard_runs` has no reason column. `cmd/flow-guard/main.go`'s `runAndRecord` keeps only the
  first 256 bytes of stdout's first line, and that line is what `outcomeFor` classifies.
  `prove-reproducer` puts 14 exit-2 reasons on stderr, and its two exit-1 reasons on stdout
  around line 5, under a first line that always reads `== pre-fix leg …`.

This is one change because the ticket's two halves meet in `check-plan-shape`. Its rule ids are
the rows of the printed list and also the reasons its refusals store.

The guard package injects `Env` "so tests run in-process and in parallel" (`stats/internal/guard/guard.go`).
121 of its 143 test files call `t.Parallel()` (measured on `9b6d28b9`). The package registers
about 139 guards and helper commands in `guard.Registry`, and its non-test files hold about 780
`return <non-zero literal>` / `return CannotAnswer(…)` sites. About 697 of those sit in a function
that takes an `Env` parameter, and 29 sit in 17 int-returning helpers that do not (a `go/ast` count
over `stats/internal/guard` at `9b6d28b9`).

The approved design is round 1 and its answers (`.superpowers/sdd/planner/round-1.md`,
`answers-1.md`): Q1 in-process `guard.Reason`, Q2 `reasons TEXT[]`, Q3 every registered guard.
Nothing is visible to a user, so no prototype was built.

## Decisions

### The guard prints its own rule list, and the planner brief loads it

**ID:** rule-list-from-guard
**Status:** active
**Chosen:** `check-plan-shape` prints its rules, and the planner brief loads that output, so the rules have one source — the ticket's stated design ("rule list plus stored reasons").
**Considered:** keep the rules in planner prose and add the missing rows — two sources that drift, which is the defect itself.

### The list is a `-rules` argument of check-plan-shape

**ID:** rules-flag
**Status:** active
**Chosen:** `flow-guard check-plan-shape -rules` prints one `<id>  <sentence>` line per live rule and exits 0. The lines come from one table, `psRules` in `planshape.go`. F21 is retired and has no row.
**Considered:** a separate `plan-shape-rules` guard — a second registry entry for one table that lives in this guard; a `--help` text — help is usage, and its format promises no one-rule-per-line contract.

### Sub-guard rule ids are their guard names

**ID:** sub-guard-ids
**Status:** active
**Chosen:** the five sub-guards' rows and reason ids are their registry names (`plan-suite-steps`, `plan-script-paths`, `plan-files-paths`, `plan-task-id-rule`, `plan-step-files`).
**Considered:** invented short ids (F30–F34) — a second name for something the registry already names.

### Every check-plan-shape violation line names its id in one position

**ID:** line-id-position
**Status:** active
**Chosen:** every violation line from `check-plan-shape` and its five sub-guards reads `path:line: <id> <message>`, replacing today's F22 prefix and `(Fnn)` suffixes.
**Considered:** keep the mixed prefix and suffix and add ids where they are missing — a planner, or any reader of these lines, then has to search two positions.

### The header comment keeps the rationale

**ID:** rationale-stays-in-header
**Status:** active
**Chosen:** `psRules` holds only the planner-facing sentence, what a plan must do. The `planshape.go` header comment keeps each rule's rationale: why it exists, which kan, how it is detected.
**Considered:** print the rationale too — the planner reads a requirement, and the detection details and history would bury it.

### Prose that only restates a printed rule is cut

**ID:** cut-restating-prose
**Status:** active
**Chosen:** in the briefs that load `-rules`, a passage that only restates a printed rule is cut in favour of the pointer. A passage that adds guidance beyond the rule (a worked shape, the reason to prefer one form) stays.
**Considered:** keep both copies — the two drift, which is this ticket's defect.

### Which briefs load the list

**ID:** briefs-load-rules
**Status:** active
**Chosen:** the planners that write a `tasks.md` run `flow-guard check-plan-shape -rules` before writing it: section D of `skills/flow/brainstorm-planner.md` (used by `/flow`, `/flow-plan` and the self-review and in-run fix planners), the plan step of `skills/flow-fast/SKILL.md`, and the in-run fix task append of `skills/flow/verify-fix-loop.md`. A repotest pins each.
**Considered:** `skills/spectre-planning/SKILL.md` as well — spectre vendors it byte for byte for planning without flow, where no `flow-guard` exists.

### Reason ids reach telemetry in-process

**ID:** reason-in-process
**Status:** active
**Chosen:** a guard records an id at each refusal and cannot-answer site by calling `guard.Reason`, and `cmd/flow-guard` reads the recorded ids after the run and posts them with it (operator, round 1 Q1). A guard that hands work to a child `flow-guard` process does not see the child's ids. The child records its own run, so those ids are attributed to the child guard, and the parent tags its own site.
**Considered:** tagged output (`[<id>]` on each refusal line) that the binary parses from a stdout and stderr tee — breaks on a coincidental bracketed token (a path, a quoted plan line), and misses any refusal site that prints nothing.

### The reason collector rides on Env

**ID:** reason-on-env
**Status:** active
**Chosen:** `guard.Reason(env Env, id string)` adds `id` to a collector that `Env` points at. The collector keeps distinct ids in first-seen order and is safe under concurrent calls. A nil collector, as in every test's `Env{…}` literal, makes the call a no-op. `cmd/flow-guard`'s `run` sets the collector, and `runAndRecord` reads it. In-process sub-guards get the same `env` and record into the same run.
**Considered:** a package-level collector behind `guard.Reason(id)` — with 121 test files running guards in parallel, no test could assert which ids its own run recorded, and `Env` exists to keep guard state per call (`guard.go`).

### The column holds every distinct id of a run

**ID:** reasons-array
**Status:** active
**Chosen:** `guard_runs.reasons TEXT[]` holds all distinct ids of the run in first-seen order (operator, round 1 Q2).
**Considered:** `reason TEXT`, the first id only — the F-checks print before the sub-guards, so a first-only column under-counts the later rules and the ranking comes out wrong.

### The column is nullable

**ID:** reason-nullable
**Status:** active
**Chosen:** migration `0064` adds `reasons TEXT[]` with no NOT NULL constraint. A clear run, or a post from an older cached binary that omits the field, stores NULL. The wire field is `omitempty`. The store refuses an empty-string id (`ErrInvalidGuardRun`), the same way it refuses an empty guard or worktree.
**Considered:** refuse a post with no reasons — that would drop every row an older cached binary sends.

### Every registered guard is tagged

**ID:** every-guard-tagged
**Status:** active
**Chosen:** each refusal and cannot-answer site of every guard in `guard.Registry` records an id, `records.GuardHelpers` included, since the pin covers the whole package (operator, round 1 Q3). `cmd/flow-guard`'s own cannot-answer exit, when it cannot read the working directory, records `cannot-read-working-dir`.
**Considered:** `check-plan-shape` and `prove-reproducer` only — leaves every other guard's refusals unrankable, against the ticket's "every guard run"; an automatic fallback id (`usage`, `exit-<code>`) from the binary — it restates the exit code and ranks nothing new.

### A pin test fails an untagged exit site

**ID:** refusal-site-pin
**Status:** active
**Chosen:** a `go/ast` test over `stats/internal/guard`'s non-test files fails on any `return` of a non-zero integer literal or of a `CannotAnswer(…)` call, inside a function that takes an `Env`, that is not preceded in its block by a `Reason(env, …)` call. A loop that records one id per violation also counts. Each of the 17 int-returning helpers with no `Env` either takes one or returns a value that is not an exit code (`psWorse`, `CannotAnswer` itself, `afterFieldLine`). The test fails on a helper it cannot classify.
**Considered:** a runtime check in `runAndRecord` — it sees only the paths a run happens to take, so an untagged site ships until production reaches it.

### A fire on exit 0 records a reason too

**ID:** outstanding-is-a-fire
**Status:** active
**Chosen:** `check-unfinished-work`, whose exit-0 `OUTSTANDING:` verdict `outcomeFor` classifies as `fired`, records an id for that verdict as well. Any run whose outcome is `fired` or `cannot-answer` carries a reason. The pin (`refusal-site-pin`) covers non-zero returns only, so this exit-0 tag is held by a named test in `unfinishedwork`'s tests asserting the `OUTSTANDING:` run records its id.
**Considered:** tag non-zero exits only — the one guard that fires on exit 0 would then store NULL for its fires.

### Reason ids are local kebab-case

**ID:** reason-id-shape
**Status:** active
**Chosen:** outside `check-plan-shape`, a reason id is kebab-case and local to its guard (`prove-reproducer`: `usage`, `not-a-worktree`, `pre-fix-not-demonstrated`, …). The guard column already scopes it.
**Considered:** a global registry of ids — a second table to keep in step with about 780 sites, for no reader that needs ids to be unique across guards.

### No view change

**ID:** no-view-change
**Status:** active
**Chosen:** this change stores the reasons. Ranking them is a SQL `GROUP BY` over `guard_runs`. `GuardHealth`, the `guard-health` view and `web/src/views/FlowHealth.tsx` stay unchanged.
**Considered:** extend `GuardHealth` and the Flow health SPA with a reasons column — the ticket asks for storage, and a view is a later change once ranking has a reader.

### The checker lands first, and each tagging task runs it over its own files

**ID:** pin-per-range
**Status:** active
**Chosen:** the pin's checker lands with the collector (Task 1). `check-plan-shape`'s task and each tagging task run it over their own files in a range test, and the package-wide pin (`refusal-site-pin`) replaces those range tests once the last tagging task lands. Each tagging task can then prove its own sites, and the tagging tasks run side by side on disjoint files.
**Considered:** one package-wide test only, after the last tagging task — no tagging task could check its own sites, and a missed site would surface only in the last task.

### The pin also sees exits by panic and methods on a struct holding Env

**ID:** pin-reach-extended
**Status:** active
**Chosen:** besides `refusal-site-pin`'s returns, the pin counts `panic(<T>(<non-zero integer literal>))` as an exit site. A method whose receiver struct holds an `Env` field counts as a function that takes one. A function whose integer is not its guard's exit code (a child process's status, a count, a position, a line number, or a combination of codes its callers already recorded) is named, with why, in an exempt map the pin reads, and its callers record at their own exits.
**Considered:** returns in functions with an `Env` parameter only — `integrate-change`, `live-store-copy` and `visual-verify-pass` exit through `panic` in `die`, `stop` and `usage` methods on a struct holding their `Env`, so their exits would ship untagged and unpinned.

### A clear run posts no reasons

**ID:** clear-run-drops-reasons
**Status:** active
**Chosen:** `cmd/flow-guard` posts the recorded ids only when the run's outcome is `fired` or `cannot-answer`. A run classified `clear` posts none and stores NULL (`reason-nullable`), even when some path recorded an id before the guard recovered.
**Considered:** post every recorded id — an in-process sub-guard a parent tolerates, or a helper's recovered failure, would store reasons on a clear row.

### check-unfinished-work records one id per OUTSTANDING signal

**ID:** unfinished-work-per-signal
**Status:** active
**Chosen:** each signal that adds a reason to `check-unfinished-work`'s `OUTSTANDING:` verdict records its own id (`unticked-parent-actions`, `unchecked-plan-items`, …), so a ranking shows which signal holds runs open (`outstanding-is-a-fire`).
**Considered:** one `outstanding` id for the verdict — it restates the `fired` outcome and ranks nothing.

## Open questions

## Recommended approach

1. **Rule table and `-rules`** (`rules-flag`, `sub-guard-ids`, `rationale-stays-in-header`).
   `planshape.go` gets `psRules`, an ordered table of rule id and planner-facing sentence. It has
   one row per live F-row (F3a and F3b separately, F21 omitted) and one per sub-guard.
   `checkPlanShape` answers a lone `-rules` argument by printing the table and exiting 0. A test
   pins table and code together: every `p.on` tag and every name in `psCheckOne`'s sub-guard list
   has a row, and every row is used.
2. **Uniform violation lines** (`line-id-position`). `psPlan.at` takes the id and writes
   `path:line: <id> <message>`. Each sub-guard's line gets its name in the same position. Existing
   tests that match today's line text move to the new shape.
3. **Reason collector** (`reason-in-process`, `reason-on-env`). `guard.go` gains the collector and
   `Reason(env, id)`. `cmd/flow-guard` builds it into the run's `Env`. `runAndRecord` passes the
   ids to `recordRun`, which adds `reasons` to the `guardRun` wire struct (`omitempty`) unless
   the outcome is `clear` (`clear-run-drops-reasons`).
4. **Storage** (`reasons-array`, `reason-nullable`). Migration `0064_guard_run_reasons.sql` adds
   `reasons TEXT[]`. `records.GuardRun` gains `Reasons []string` (`json:"reasons,omitempty"`), and
   `store.RecordGuardRun` inserts it through pgx's native `[]string` ↔ `text[]` mapping. The POST
   handler needs no code: it decodes `records.GuardRun`.
5. **Tag every guard** (`every-guard-tagged`, `reason-id-shape`, `outstanding-is-a-fire`,
   `unfinished-work-per-signal`).
   `check-plan-shape` records each violation's rule id, and each sub-guard records its own name.
   Every other guard records a local id at each refusal and cannot-answer site. The work is split
   into five tasks by filename range, each carrying about 190 sites, the calls routed through an
   exiting method included.
6. **Pin** (`refusal-site-pin`, `pin-per-range`, `pin-reach-extended`). The checker lands with
   the collector, and each tagging task runs it over its own files. The package-wide `go/ast`
   test lands after the last tagging task, when it first passes.
7. **Briefs** (`briefs-load-rules`, `cut-restating-prose`). The three briefs run `-rules` before
   writing `tasks.md`, and prose that restates a printed rule is cut. Each brief gets a repotest.
8. **Docs.** `stats/README.md`'s `guard_runs` paragraph names the `reasons` column. `README.md`'s
   **Plan:** guard line names `-rules`.

Competing Plan agents were skipped. The ticket fixes the shape, a guard-printed list plus stored
reasons. The remaining choices were round 1's three questions, each with one honest
recommendation and each answered by the operator.

## Critical files

- `stats/internal/guard/planshape.go` — rule table, `-rules`, line shape, rule-id reasons.
- `stats/internal/guard/plansuitesteps.go`, `planscriptpaths.go`, `planfilespaths.go`,
  `plantaskidrule.go`, `planstepfiles.go` — sub-guard line shape and reasons.
- `stats/internal/guard/guard.go` — `Env`'s collector and `Reason`.
- every other non-test `stats/internal/guard/*.go` file that registers a guard or holds an exit
  site — reason ids (among them `provereproducer.go` and `unfinishedwork.go`).
- `stats/internal/guard/*_test.go` beside each of those — tests asserting the ids, and line-text
  assertions moved to the new shape.
- `stats/cmd/flow-guard/main.go`, `telemetry.go`, `telemetry_test.go` — collector wiring and the
  wire field.
- `stats/internal/store/migrations/0064_guard_run_reasons.sql` — the column.
- `stats/internal/records/types.go` — `GuardRun.Reasons`.
- `stats/internal/store/health.go`, `health_test.go` — insert and read back.
- `stats/internal/api/health.go`, `health_test.go` — the POST carries the field (handler unchanged, test extended).
- `stats/internal/repotest/` — the brief pins.
- `skills/flow/brainstorm-planner.md`, `skills/flow-fast/SKILL.md`, `skills/flow/verify-fix-loop.md` — the briefs.
- `stats/README.md`, `README.md` — docs.
- depends on: `stats/cmd/uitest-seed/seed.go` (writes `GuardRun`s; NULL reasons are valid), `stats/internal/guard/planshape_prose_test.go` (`TestDesignSectionsNamedInProse`, which pins prose to guard data).

## Existing code to reuse

- `psPlan.at` and `psPlan.on` (`planshape.go`) — the one place a line is formatted and the one place a rule tag is named. The pin test reads the tags from `on`.
- `psCheckOne`'s sub-guard name list (`planshape.go`) — the source for the sub-guard rows.
- `guard.CannotAnswer` (`shimduties.go`) — the cannot-answer code the pin recognises.
- `firstLine` and `outcomeFor` (`cmd/flow-guard/telemetry.go`) — unchanged; the outcome is still classified from them.
- `decodeJSONBody` (`internal/api/stages.go`) — the strict decoder the handler already uses; the new field passes through it unchanged.
- `TestDesignSectionsNamedInProse` (`planshape_prose_test.go`) and `stats/internal/repotest/planner_reads_test.go` — the patterns for pinning prose to guard data and a skill's wording.
- `scripts/flow-guard live-store-copy` (`livestorecopy.go`) — the live check's stand-in store.

## Verification

- `cd stats && go test -count=1 ./internal/guard/ ./cmd/flow-guard/ ./internal/store/ ./internal/api/ ./internal/repotest/`, with the pin test passing over every tagged site.
- `scripts/flow-guard check-plan-shape -rules` prints one line per live rule, F3a, F3b and the five sub-guards included, and exits 0.
- `scripts/flow-guard check-plan-shape <fixture tasks.md>` on a fixture that breaks F2, F6 and `plan-step-files` prints `path:line: <id> <message>` lines whose ids each match a `-rules` line, and exits 1.
- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` is clean. The project's `## lint` list passes, the prose guards included.

## Live check

From the change's worktree, `scripts/flow-guard live-store-copy up kan-1123-live` builds `flowd`
from the worktree against a copy of the dev store, so migration `0064` runs on the copy. Export
the `FLOW_RECORDS_ADDR` it prints, then run, still from the worktree:

1. `scripts/flow-guard check-plan-shape <fixture>` on a tasks.md that breaks two known rules (expected exit 1).
2. `scripts/flow-guard prove-reproducer` with no arguments (expected cannot-answer, reason `usage`).
3. `scripts/flow-guard check-plan-shape -rules` (expected exit 0, a clear run).

Record the copy's `guard_runs` before and after: the count of rows with non-NULL `reasons` (0
before), and the last three rows' `guard`, `outcome`, `reasons`. Success: row 1 is `fired` with
both rule ids, row 2 is `cannot-answer` with `{usage}`, and row 3 is `clear` with NULL. Failure
looks like: no new rows (the post was refused — strict decoding, or an unknown project), `fired`
or `cannot-answer` rows with NULL `reasons`, or only the first rule id on row 1. Finally,
`scripts/flow-guard live-store-copy down kan-1123-live`.

## Non-goals

- No flow-health view, SPA column or `GuardHealth` change (`no-view-change`).
- No backfill of `reasons` for existing `guard_runs` rows. Their reasons were never captured.
- `skills/spectre-planning/SKILL.md` is unchanged. spectre vendors it for planning without `flow-guard` (`briefs-load-rules`).
- No ids for the `scripts/flow-guard` bootstrap's own cannot-answer exit, or for the KAN-902 re-exec failure in `main`. Both exit before any telemetry is recorded.
- No change to `guard_verdicts` or to any guard's exit codes or verdict lines, beyond the `check-plan-shape` line shape.

## Risks

- **Strict decoding during rollout.** `decodeJSONBody` refuses unknown fields. A worktree's new
  binary that posts `reasons` to a dev `flowd` built before this change gets a 400, and that run's
  row is dropped, until the post-landing `make -C stats restart`. Clear runs omit the field and
  still record. The loss is bounded to the refusing runs in that window, and telemetry is
  best-effort by contract.
- **Plan size.** About 820 literal exit sites, plus about 140 calls routed through an exiting
  method, across about 150 files.
  <!-- measured: a go/ast count of non-zero literal and CannotAnswer returns, and grep -cE '\b(c|v|l)\.(die|stop|usage)\(' per file, over stats/internal/guard's non-test files — a scratch tool outside the tree, not re-runnable from it @ 9b6d28b9 -->
  The pin test is what keeps the tagging complete. A missed site fails the pin, not production.
- **The pin's reach.** The pin recognises literal and `CannotAnswer` returns and
  `panic(<T>(<literal>))` only (`pin-reach-extended`). A code that one
  helper returns and a caller passes through as `return rc` is checked at the helper's own site,
  not at the caller. A helper with no `Env` is caught by the test's classify-or-fail rule. An
  exempt helper's code is checked nowhere: each caller that passes it on records by hand.
- **Line-shape churn.** Tests and any prose that match today's `(Fnn)` suffix or bare lines must
  move to the new shape in the same task, or that task's targeted tests go red.
- **Child processes.** A refusal a parent guard relays from a child `flow-guard` is stored under
  the child's guard name. The parent's own row carries only the parent's id (`reason-in-process`).
