> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Give development and UI testing each a stack of their own. A `/myflow-do` worktree gets its
own database and port; ad-hoc UI testing from the main checkout gets a named second stack instead of
the live one.

## Global Constraints

- **Every `Default` in the isolation table is today's literal value.** That column is where the
  backwards-compatibility promise is kept — a wrong value there breaks the main checkout, not the
  worktree.
- **The daemon's own config resolution is not being redesigned.** `MYFLOWD_PORT`, `MYFLOWD_DSN`,
  `MYFLOW_TRANSCRIPTS_DIR` and `MYFLOW_STATE_DIR` are read today and are used as they stand.
- **The Go test suite's existing per-test-database isolation is not touched.**
- **No task edits `openspec/` or `docs/superpowers/`.**

## Baseline

**Measured 2026-08-15 against `33f77be`:** 338 top-level Go tests
(`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'`), 112 SPA tests
(`cd stats/web && npm test`).

<!-- measured: both suites run on 2026-08-15 against the myflow-postgres compose stack, with stats/internal/web/dist already built -->

---

### 1 `MYFLOW_ADDR` overrides the client's default daemon address

**Build:** green

**Files:**
- Modify: `stats/cmd/myflow/state.go`
- Modify: `stats/cmd/myflow/stage.go`
- Modify: `stats/cmd/myflow/state_test.go`

**Interfaces:**
- Consumes: the existing `defaultAddr` constant and the three `-addr` flag registrations.
- Produces: `resolveDefaultAddr()`, the value the three registrations take as their default.

- [x] **Step 1: A failing test for the precedence rule**

Write the table first and watch it fail: unset leaves the built-in default; `MYFLOW_ADDR` set
replaces it; an explicit `-addr` beats both. Use `t.Setenv` so the cases do not leak into each other.

- [x] **Step 2: The resolver**

`resolveDefaultAddr()` reads `MYFLOW_ADDR` and falls back to `defaultAddr` when it is unset or empty,
mirroring how `internal/config.FromEnv` already treats an empty variable as unset. The three flag
registrations — two in `state.go`, one in `stage.go` — take their default from it rather than from the
constant.

- [x] **Step 3: Say why the variable exists**

The resolver's doc comment records that the `-addr` flag already existed and was not the problem: it
was not passed, and a value that must be remembered once per session rather than once per command is
the actual change.

**Tests:** `TestResolveDefaultAddr` in `stats/cmd/myflow/state_test.go`, covering unset, set, empty,
and flag-beats-environment.

**Regression:** Reverting this leaves every client invocation defaulting to the live daemon, so a
worktree's commands write to the main checkout's database.

**Baseline:** before=338 after=339 Go top-level tests.
<!-- predicted: one new top-level test function, not yet written -->

**Commit:** `feat(1): resolve the client's default daemon address from MYFLOW_ADDR`

---

### 2 `scripts/workspace.sh` creates, removes and reports workspace databases

**Build:** green

**Files:**
- Add: `scripts/workspace.sh`
- Add: `scripts/test-workspace.sh`

**Interfaces:**
- Consumes: the `myflow-postgres` container and the workspace id passed as the single argument.
- Produces: the `create`, `remove` and `survivors` commands the isolation contract's command table
  names.

- [x] **Step 1: Failing round-trip test**

`scripts/test-workspace.sh`, in the style of the repository's existing `scripts/test-*.sh`. Assert,
in order: `create <id>` makes `myflow_<id>`; `survivors <id>` then prints that database and exits 0;
`remove <id>` drops it; `survivors <id>` then prints nothing **and still exits 0**. Skip the whole
file with a clear message when the container is not running, exactly as the Go suite does.

- [x] **Step 2: The three subcommands**

`create` makes the database when absent and is a no-op when present, so a re-run of `/myflow-do` in
an existing worktree is not an error. `remove` drops it if present. Both refuse an empty or
missing id argument rather than operating on `myflow` itself.

- [x] **Step 3: `survivors`, written against the two rules that decide its verdict**

Filter with `awk`, never `grep`: `grep` exits 1 when nothing matched, and the contract reads a
non-zero exit as *the check could not run*, so a `grep`-filtered report can never reach the one
result that verifies the cleanup. Run under `pipefail` so a failing `docker exec` is not reported as
an empty — and therefore verified — report. Do not end the pipeline with `head`.

- [x] **Step 4: A timeout inside the container**

`survivors` reaches its service through `docker exec`, and the guard's sixty-second bound terminates
a host process group that this work is not in. Bound it from the inside — a connect timeout and a
statement timeout on the client running in the container — so a hung query ends rather than
outliving the bound.

**Tests:** `scripts/test-workspace.sh`, added to `## test` in `.myflow/project.md` in task 3.

**Regression:** Reverting this leaves the isolation section naming commands that do not exist, so
`/myflow-do` cannot create a workspace database and run 2 cannot remove one.

**Baseline:** Go and SPA counts unchanged; one new Bash test harness.
<!-- predicted: this repository's Bash guards are not counted by either suite's total -->

**Commit:** `feat(2): add the workspace create, remove and survivors commands`

---

### 3 This repository declares `## workspace isolation`

**Build:** green

**Files:**
- Modify: `.myflow/project.md`

**Interfaces:**
- Consumes: `scripts/workspace.sh` from task 2 and `MYFLOW_ADDR` from task 1.
- Produces: the declaration `/myflow-do` resolves and `scripts/check-workspace-isolation.sh`
  validates.

- [x] **Step 1: The two tables**

The resource table carries three rows — `database` on `MYFLOWD_DSN`, `port` on `MYFLOWD_PORT`, `url`
on `MYFLOW_ADDR` built from `<value:MYFLOWD_PORT>` — with each `Default` copied from the value the
code uses today, verified by reading `stats/internal/config/config.go` rather than from memory. The
command table names the three subcommands from task 2.

- [x] **Step 2: State what is not isolated, and why**

In prose beside the tables: `MYFLOW_STATE_DIR` and `MYFLOW_TRANSCRIPTS_DIR` have no rows because the
`Resource` vocabulary is closed and a directory path is none of its five words. Record the bounded
consequence — a worktree harvests the same transcripts into its own database, so the data is
duplicated rather than shared. The contract requires this prose: an absent row must read as a
decision, not an oversight.

- [x] **Step 3: Name the new harness**

Add `scripts/test-workspace.sh` to `## test`, and a `## run` note pointing at the UI-test stack
targets that task 6 adds.

- [x] **Step 4: The guard actually validates something now**

Run `scripts/check-workspace-isolation.sh` and fix every reported row. Until this task it passed by
declaring nothing; a green run now means the declaration is well formed.

**Tests:** `scripts/check-workspace-isolation.sh` against this repository, which must report
`ISOLATION-OK`.

**Regression:** Reverting this returns every apply worktree to the main checkout's database and port.

**Baseline:** Go and SPA counts unchanged.
<!-- predicted: a configuration change with no test-count effect -->

**Commit:** `feat(3): declare this repository's workspace isolation`

---

### 4 The contract files stop citing this repository as having nothing to isolate

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/project-configuration.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `scripts/prepare-workspace.sh`
- Modify: `scripts/check-workspace-isolation.sh`
- Modify: `scripts/test-check-workspace-isolation.sh`
- Modify: `.myflow/project.md`

**Allowed-collateral:** `skills/**/*.md`, `scripts/*.sh`

**Interfaces:**
- Consumes: the declaration added in task 3.
- Produces: contract text whose worked example is true.

**The file list is wider than the two contract files, and that is the task rather than a scope
creep.** The stale claim was repeated in a script header and in a comment, and a half-corrected
claim is worse than an uncorrected one — a reader who finds the surviving copy has no way to tell
which one is current. `skills/myflow-do/SKILL.md`'s copy carries a live consequence rather than a
cosmetic one: section 6 resolves the handoff's run instructions from it, so a run reading the stale
sentence would refuse this repository's own application the URL it demonstrably has.

`skills/myflow-contracts/workspace-isolation.md` was read and found to need no edit — its phrasing
was already generic and never named this repository — so it is deliberately absent from the list
above. It is recorded here rather than dropped silently, because "checked and correct" and "not
checked" are the same absence otherwise.

- [x] **Step 1: Correct both claims**

Both files use this repository as the example of a project that "declares no runnable application, so
it has nothing to isolate". `stats/` made that false, and task 3 makes it doubly so. Replace the
example with the general statement it was illustrating — a project with no runnable application has
nothing to isolate — without naming this repository as that project.

- [x] **Step 2: Keep the surrounding rules intact**

The point being made in both places is that an absent section is a supported state, not a
misconfiguration. That rule does not change; only the example does.

- [x] **Step 3: Run the reference and vocabulary guards**

`scripts/check-references.sh` and `scripts/check-vocabulary.sh` both read these files, and a
cross-reference broken by an edit here fails the build rather than being found later.

**Tests:** `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
`scripts/check-markdown-integrity.sh`.

**Regression:** Reverting this restores a worked example that contradicts the repository's own
configuration.

**Baseline:** Go and SPA counts unchanged.
<!-- predicted: a documentation change with no test-count effect -->

**Commit:** `docs(4): stop citing this repository as having nothing to isolate`

---

### 5 The UI-test fixture seeder, behind a `_uitest` name guard

**Build:** green

**Files:**
- Add: `stats/cmd/uitest-seed/main.go`
- Add: `stats/cmd/uitest-seed/guard.go`
- Add: `stats/cmd/uitest-seed/guard_test.go`
- Add: `stats/cmd/uitest-seed/seed.go`
- Add: `stats/cmd/uitest-seed/seed_test.go`

**Interfaces:**
- Consumes: `internal/store`'s exported write methods and `internal/config.FromEnv` for the DSN.
- Produces: a seeded fixture, and a guard that refuses any target not naming a `_uitest` database.

- [x] **Step 1: The guard, tested first**

Parse the database name out of the DSN and require it to end in `_uitest`. Write the refusal cases
first — the live DSN, a DSN naming `myflow`, a DSN whose name merely contains `uitest` in the middle
— and watch them fail. The check runs before any statement is issued; a guard that runs after the
first write is not a guard.

- [x] **Step 2: The fixture**

Two projects, changes covering `STARTED`, `IN_PROGRESS` and `FINISHED`, and stage runs carrying token
usage so the cost and statistics views have something to render. Written through the store's exported
types, so a migration that changes a column stops this compiling.

- [x] **Step 3: Test it against a throwaway database**

Use the existing per-test-database helper pattern, naming the test database with a `_uitest` suffix so
the guard admits it. The test must touch neither the live database nor the shared UI-test one.

**Tests:** `TestGuardRefusesNonUitestDSN`, `TestGuardAcceptsUitestDSN`,
`TestGuardRefusesUnparsableDSN`, `TestSeedPopulatesFixture`.

**The `_uitest` match is a suffix, not a substring, and the guard's own test states so.** A database
named `myflow_uitest_real` is refused: it merely contains the word, and is exactly as unprotected as
`myflow` itself. A substring match would admit it, which is the failure this guard exists to
prevent.

**Regression:** Reverting this leaves the test stack empty, and removes the refusal that stops a
misaimed seed from writing into the live database.

**Baseline:** before=339 after=342 Go top-level tests.
<!-- predicted: three new top-level test functions, none written yet; 339 is task 1's post-state -->

**Commit:** `feat(5): seed the UI-test stack behind a _uitest name guard`

---

### 6 `make ui-test-up` and `make ui-test-down`

**Build:** green

**Files:**
- Modify: `stats/Makefile`
- Modify: `stats/README.md`

**Interfaces:**
- Consumes: `stats/cmd/uitest-seed` from task 5 and the daemon's existing environment resolution.
- Produces: the two targets an operator runs to get and release a UI-test stack.

- [x] **Step 1: `ui-test-up`**

Drop and recreate `myflow_uitest`; start `myflowd` with `MYFLOWD_PORT=4174`, `MYFLOWD_DSN` naming
that database, and `MYFLOW_TRANSCRIPTS_DIR` and `MYFLOW_STATE_DIR` both pointed at temporary
directories; wait for it to answer rather than sleeping a fixed interval; then run the seeder.

Pointing the transcripts root at an empty directory is load-bearing, not tidiness: the daemon
harvests on a five-second interval, so a test database left on the real root refills itself with
genuine session data within seconds of starting.

- [x] **Step 2: `ui-test-down`**

Stop the daemon and drop the database, both idempotently, so a second `down` is not an error. The
drop refuses any name not ending in `_uitest`, the same rule task 5's seeder applies.

- [x] **Step 3: Document both**

`stats/README.md` gains a section covering what the stack is for, how to point a session at it with
`MYFLOW_ADDR`, and the fact that bring-up resets it. State plainly that the Go test suite does not
need this stack — it already creates per-test databases — so nobody reaches for it when running
`go test`.

**Tests:** Manual, and recorded as such: `make ui-test-up`, confirm the SPA renders the fixture at
`http://127.0.0.1:4174`, confirm `select count(*) from projects` in the live `myflow` database is
unchanged, then `make ui-test-down`.

**Regression:** Reverting this removes the only supported way to bring up a stack that is not the
live one.

**Baseline:** Go and SPA counts unchanged; the targets are exercised manually.
<!-- predicted: Makefile targets carry no automated test in this repository -->

**Two mechanisms guard the reset, and review established that neither is redundant.** `dropdb
--force` is what makes the drop succeed against a live connection, including one this Makefile did
not start. The kill block is what frees port 4174: measured during review, the daemon does **not**
exit when its database is force-dropped — it keeps running, holds the port, and answers 500s, so a
second `myflowd` fails to bind. An earlier fix round claimed the ordering alone was the fix; it was
not, and the comment now states what each mechanism actually does.
<!-- measured: both behaviours reproduced live against myflow-postgres on 2026-08-15 -->

**Commit:** `feat(6): add the ui-test-up and ui-test-down targets`
