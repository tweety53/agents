# Tasks — kan-289

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Two concerns on one branch, per `design.md`'s `two-concerns-separate-commits`. Tasks 1-4 are
KAN-289 proper. Tasks 5-19 are the rename.

**The rename tasks are ordered by dependency, not by size.** The Go layer (5-10) is self-contained
and moves first; the directory renames (11-13) break every reference they touch and so each carries
its own reference sweep in the same commit; the prose and guard layers (14-17) come last because
they are what the earlier tasks' references point at. **Every task that grows or renames a file
budgeted by `scripts/check-contract-budget.sh` runs that guard and reconciles its table in the same
commit** — that guard is already in the project's `## lint` list, so a commit that leaves it red
leaves lint red.

**A rename moves a row's KEY and never its value; growth inside budget changes nothing at all.**
Task 13 renamed a budgeted file and deferred the key to a later task, which left
`check-contract-budget.sh` — and therefore the whole `## lint` list — red on the branch between two
tasks. The key was renamed in task 13's own commit instead, carrying the same number: a rename does
not change a file's size, so re-deriving the budget from the renamed file would silently re-baseline
a ratchet that had not moved.

**Reconcile means run the guard and answer what it actually reports — it does not mean raise a row.**
That guard is a ratchet: a row carries the size its file had when the change adding the row landed,
plus 25%, so an ordinary edit is expected to fit inside the existing headroom and change nothing.
**Raise a row only when the guard actually fails on that file, and only to the size the guard's own
rule gives.** A rename changes a row's *key* and must be reconciled; growth inside budget must not.
Raising a row the guard never complained about loosens the ratchet silently, which is the one thing
this guard exists to prevent — and it is not caught by any check, because the guard passes either
way. **Task 2 did exactly this and it was reverted**: `skills/flow/review-panel.md` grew 21958 → 22460
against a standing budget of 27420, and the row was raised to 28075 for an addition that was never
over.
<!-- measured: wc -c skills/flow/review-panel.md, before and after task 2's edit; and the budgets() row at 07acd2e @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->

**Baseline for the whole rename**, measured before any edit:

- 231 files under the live corpus carry the string, across 3297 occurrences.
  <!-- measured: grep -ril myflow skills scripts rules commands commands-claude stats .myflow README.md CLAUDE.md AGENTS.md setup.sh | wc -l; and the same with -roi | wc -l @ 07acd2e (before this change) -->
- The Go suite passes in every one of its sixteen packages.
  <!-- measured: cd stats && go test ./... -count=1 @ 07acd2e (before this change) -->
- The SPA suite passes: 9 files, 157 tests.
  <!-- measured: cd stats/web && npm test @ 07acd2e (before this change) -->

## Part 1 — the dispatch template

- [x] 1. Guard and harness for the REPRODUCE, DON'T READ block

  Write `scripts/test-check-reproduce-not-read.sh` first, then
  `scripts/check-reproduce-not-read.sh` against it — RED before GREEN.

  The guard is argument-free and self-scoped, resolving its scan set from its own location exactly as
  `scripts/check-guard-symlinks.sh` does, with an opt-in `CHECK_REPRODUCE_NOT_READ_ROOT` override
  honoured only when set so the harness can point it at a sandboxed fixture tree under `TMPDIR`.
  Never set that override for a normal invocation.

  The required-site table lives in the guard, keyed by path relative to the scan root:

  ```bash unverified:confirm the two skill paths still hold once tasks 2 and 3 have landed
  # site                              min blocks   required variant(s)
  # skills/flow/review-panel.md            1        reviewer
  # skills/flow/implement.md               2        reviewer AND implementer
  ```

  A block is a line carrying the literal label `**REPRODUCE, DON'T READ:**`. A variant is identified
  by its own load-bearing phrases, which the guard holds as short literals rather than as a copy of
  the whole paragraph, per `design.md`'s `label-plus-phrases`:

  - **shared, both variants:** `crosses a boundary`, `the store, the filesystem, a guard, a real
    transcript`, `exercise the real thing`
  - **reviewer only:** `worth less than one you did`
  - **implementer only:** `a fake or a hand-built value`

  Report every failure as `file:line` with the missing element named — the label, or which phrase.
  Exit `0` clean, `1` a required site missing its block or one of its phrases, `2` it cannot answer
  at all: a scoped file that is missing, unreadable, a symlink, or a `grep` that fails for any reason
  other than "no match". A path that exists as neither file nor directory is a hard `2`, never a
  silent skip — a guard pointed at a renamed file that reports `✓ clean` is the vacuous pass
  `check-vocabulary.sh`'s own header warns about.

  Give the guard a header in the house style: what it checks, why it exists (KAN-289's root cause —
  the instruction lived in no template and depended on the dispatcher remembering), and what a green
  run does *not* prove: that the paragraph is present and carries its phrases, never that a reviewer
  or implementer actually obeyed it.

  The harness builds a fixture tree under `TMPDIR` and covers, at minimum: both sites correct
  (exit 0); the label absent from `review-panel.md` (exit 1, names the file); the label present but a
  shared phrase missing (exit 1, names the phrase); `implement.md` carrying only one block (exit 1);
  `implement.md` carrying two blocks of the *same* variant (exit 1 — the reviewer/implementer pair is
  the requirement, not the count alone); a scoped file missing (exit 2); a scoped file that is a
  symlink (exit 2). It prints `all cases passed` on success, matching the harness convention the
  other `scripts/test-check-*.sh` files use.

  **Do not register the guard yet** — task 4 does that, once tasks 2 and 3 have made the real corpus
  satisfy it. Running it by hand against the real tree at this point is expected to exit 1.

  Verify: `scripts/test-check-reproduce-not-read.sh` prints `all cases passed`; the project's `##
  lint` list is clean (the guard is not in it yet, so this is unchanged from before the task).

**Build:** green
**Files:** `scripts/check-reproduce-not-read.sh`, `scripts/test-check-reproduce-not-read.sh`
**Tests:** `scripts/test-check-reproduce-not-read.sh` — seven cases, listed above.
**Regression:** reverting this task removes both files, so nothing asserts the block is present at
either site; the two skill files would satisfy no check at all and a later prose edit could delete
the paragraph silently. Each individual case is its own regression detector for the guard branch it
names — in particular the two-blocks-same-variant case, which is the only one that fails if the
guard degrades into counting labels instead of matching variants.
**Baseline:** before=0 after=7 cases in `scripts/test-check-reproduce-not-read.sh`; the file does
not exist before this task.
**Commit:** `feat(scripts): guard the REPRODUCE, DON'T READ dispatch block`

- [x] 2. Carry the reviewer variant into every review-panel slot dispatch

  In `skills/flow/review-panel.md`, add the reviewer variant to the cluster that already states what
  travels with every slot dispatch — immediately after the CONTEXT BUNDLE paragraph, so the three
  obligations that reach every slot sit together. Introduce it exactly as its neighbours are
  introduced: **Every slot's dispatch prompt also carries the REPRODUCE, DON'T READ paragraph**,
  followed by the blockquote reproduced verbatim from `design.md`.

  It reaches all five slots, Bugbot and Security included — the same reach the per-finding reproducer
  requirement already has. Do not add a per-slot exception and do not make it conditional on the
  slot's spawn shape: a slot dispatched by `subagent_type` carries its own agent definition, and this
  paragraph is dispatcher-supplied prompt text like the CONTEXT BUNDLE beside it.

  This grows the file, and `scripts/check-contract-budget.sh` is already in `## lint`. **Run that
  guard and do what it actually reports.** The standing row for `skills/flow/review-panel.md` has
  headroom this addition fits inside, so the expected outcome is that the guard passes and the table
  is left alone. Raise the row only if the guard fails on this file, and then only to the size its
  own rule gives. Do not narrow the guard's scope and do not delete the row.

  Verify: `scripts/check-contract-budget.sh` exits clean; `scripts/check-markdown-integrity.py` exits
  clean; `scripts/check-references.sh` exits clean; running `scripts/check-reproduce-not-read.sh` by
  hand now reports only the `implement.md` site as missing, and no longer `review-panel.md`.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** no new test — task 1's harness already covers the guard's behaviour against fixtures;
this task's own verification is the guard run against the real file, plus the two structural guards
named above.
**Regression:** reverting this task removes the paragraph from the panel's slot template, so no
panel slot is told to verify by running; `scripts/check-reproduce-not-read.sh` exits 1 naming
`skills/flow/review-panel.md`, which is precisely the check that would have been absent before
task 1.
**Baseline:** not applicable — no test suite exists for this file; verified by the four guard runs
named above.
**Commit:** `docs(flow): carry REPRODUCE, DON'T READ into every panel slot dispatch`

- [x] 3. Carry both variants into the implementer and per-task reviewer dispatches

  In `skills/flow/implement.md`, add the **implementer** variant to the run of blockquotes every
  implementer dispatch carries, immediately after the `PLAN PROVENANCE` blockquote — it belongs
  beside that one, which already tells the implementer to establish the real API rather than
  transcribe a guess, and this extends the same instinct from the API to the test.

  Then add the **reviewer** variant to the per-task review paragraph, so the combined reviewer that
  judges each task's commit diff is held to the same standard as a panel slot.

  Both blockquotes are reproduced verbatim from `design.md`. They share the label and differ in
  their bodies; do not merge them, do not cross-reference one from the other, and do not replace
  either with a pointer — a pointer cannot resolve inside a subagent prompt, which is the reasoning
  panel finding F33 on `kan-201` recorded when it withdrew the same objection against CONTEXT BUNDLE.

  Run `scripts/check-contract-budget.sh` and do what it actually reports, exactly as task 2 does.
  `skills/flow/implement.md` stands at 13775 bytes against a budget of 16559, so two blockquotes fit
  inside the existing headroom and the table is expected to need no edit at all. Raise the row only
  if the guard fails on this file.
  <!-- measured: wc -c skills/flow/implement.md; and grep '^skills/flow/implement.md' scripts/check-contract-budget.sh @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->

  Verify: `scripts/check-reproduce-not-read.sh` now exits 0 over the real corpus;
  `scripts/check-contract-budget.sh`, `scripts/check-markdown-integrity.py` and
  `scripts/check-references.sh` all exit clean.

**Build:** green
**Files:** `skills/flow/implement.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** no new test — task 1's harness covers the guard; this task's verification is the guard
reaching exit 0 over the real corpus for the first time.
**Regression:** reverting this task leaves the implementer untold (it writes the fake-backed test
KAN-289's fourth defect describes) and the per-task reviewer untold (it is the review that would
catch it before the panel). `scripts/check-reproduce-not-read.sh` exits 1 naming
`skills/flow/implement.md` and reporting which variant is absent.
**Baseline:** not applicable — no test suite exists for this file; verified by the four guard runs
named above.
**Commit:** `docs(flow): carry REPRODUCE, DON'T READ into implementer and per-task review dispatches`

- [x] 4. Register the guard in the project's lint and test lists

  Add `scripts/check-reproduce-not-read.sh` to `.myflow/project.md`'s `## lint` list and
  `scripts/test-check-reproduce-not-read.sh` to its `## test` list. Place each next to the guards it
  most resembles rather than at the end, and do not add a count of the list anywhere — that file
  states, deliberately, that it cites its guard list by count nowhere.

  Add a short paragraph under `## lint` in the same style as the ones already there, saying what this
  guard checks and what a green run does not prove.

  Note that `.myflow/project.md` is itself renamed by task 11; this task edits it at its current
  path, and task 11 carries the edit forward.

  Verify: the whole `## lint` list runs clean, split across more than one invocation per that file's
  own measured-runtime note; the whole `## test` list runs clean, likewise split.

**Build:** green
**Files:** `.myflow/project.md`
**Tests:** no new test — this task adds two existing scripts to two lists.
**Regression:** reverting this task leaves the guard present but unrun, so a later edit deleting
the paragraph from either skill file passes every check the project actually executes — the guard
would exist and prove nothing.
**Baseline:** not applicable — no test suite exists for this file; verified by running both
configured lists.
**Commit:** `chore(project-config): run the REPRODUCE, DON'T READ guard in lint and test`

## Part 2 — the rename

- [x] 5. Rename the Go command directories and the build targets

  `git mv stats/cmd/myflow stats/cmd/flow` and `git mv stats/cmd/myflowd stats/cmd/flowd`. Update
  every package clause, import path and in-package identifier that names the old directory. Update
  `stats/Makefile`: the two `-o` lines become `bin/flowd` and `bin/flow`, `LIVE_BIN` becomes
  `bin/flowd`, `LIVE_LOG` becomes `/tmp/flow-live.log`, and `LIVE_CLI` becomes
  `$(HOME)/.local/bin/flow`. Update `scripts/test-make-build.sh`, which asserts both binaries are
  real files, to the new names.

  Leave `stats/cmd/uitest-seed` alone. Leave the retired command-name constants in
  `stats/internal/stages/names.go` alone — `design.md`'s `retired-command-constants-are-data` states
  why, and a task that "helpfully" renames them breaks every lookup of an existing `stage_runs` row.

  Verify: `cd stats && make build` produces `bin/flow` and `bin/flowd`;
  `scripts/test-make-build.sh` passes; `cd stats && go vet ./... && gofmt -l .` is clean;
  `cd stats && go test ./... -race -count=1` passes in all sixteen packages.

**Build:** green
**Files:** `stats/cmd/flow/**`, `stats/cmd/flowd/**`, `stats/Makefile`,
`scripts/test-make-build.sh`
**Allowed-collateral:** `stats/**/*.go`, `scripts/check-panel-reproducers.sh`,
`scripts/check-stage-mark-calls.sh`

Those last two guard scripts each carry exactly one `stats/cmd/myflow*/*.go` path citation in a
comment, and the directory rename makes both stale. **No guard catches them**:
`check-references.sh` does not scan a `.sh` comment's citation of a `.go` path. They were added to
this field after `check-task-commit-fields.sh` reported the commit touching two files it had not
declared — the edits were right and the declaration was incomplete, which is the honest way round to
fix it.
**Tests:** no new test — `scripts/test-make-build.sh` and the existing per-package Go suites are
the surface; both binaries' own package tests move with their directories.
**Regression:** reverting this task restores `cmd/myflow`/`cmd/myflowd`, and
`scripts/test-make-build.sh` fails on the binary names it asserts.
**Baseline:** before=16 after=16 Go packages, all passing.
<!-- measured: cd stats && go test ./... -count=1 @ 07acd2e (before this change) -->
**Commit:** `refactor(stats): rename the myflow and myflowd commands to flow and flowd`

- [x] 6. Rename the environment variables

  `MYFLOWD_PORT`, `MYFLOWD_DSN`, `MYFLOWD_HOST` become `FLOWD_*`. `MYFLOW_ADDR`,
  `MYFLOW_STATE_DIR`, `MYFLOW_TRANSCRIPTS_DIR`, `MYFLOW_HARNESS`, `MYFLOW_STATS_DSN`,
  `MYFLOW_STATS_ADMIN_DSN`, `MYFLOW_STATS_SKIP_SLOW_TESTS` become `FLOW_*`.

  **Enumerate the real set before trusting that list.** An earlier draft of this task also named
  `MYFLOWD_PIDFILE` and `MYFLOW_STATE_FILE`; neither is a live variable, and the implementer found
  so by measuring rather than by following the list. `MYFLOWD_PIDFILE` occurs only in two frozen
  design documents, where it names an override that was **considered and rejected** — the pidfile
  path is derived from the resolved port by `internal/pidfile.Path` and was never read from the
  environment. `MYFLOW_STATE_FILE` occurs once, in a comment in `scripts/check-references.sh`
  describing an "idiom", and never as a variable any code reads. Ten variables are real; both of
  these were plan errors, and they are recorded here rather than quietly dropped.
  <!-- measured: grep -rhoE 'MYFLOWD?_[A-Z_]+' --include='*.go' --include='*.sh' --include='*.md' --include='Makefile' --include='*.yml' --include='*.plist' | sort -u, reconciled against this list @ 07acd2e (before this change) -->

  Sweep every reader and every writer together: `stats/internal/config`, the two command packages,
  `stats/Makefile`, `stats/docker-compose.yml`, `stats/launchd/*.plist`, `stats/README.md`, the guard
  scripts under `scripts/` that export or read them, and the skill and contract prose that names
  them. A variable renamed in the reader but not in the Makefile that sets it fails silently — the
  process simply sees an unset value and takes its default, which is exactly the shape that survives
  a reading review.

  There is **no** compatibility fallback reading the old name. `design.md`'s `dotmyflow-hard-cutover`
  reasoning applies to the environment identically: a dual read lets two shells sit in different
  states indefinitely with nothing to notice.

  Verify: `cd stats && go test ./... -race -count=1` passes; `scripts/check-workspace-isolation.sh`
  exits clean against this repository (it reads the `## workspace isolation` table, whose `Variable`
  column this task rewrites); `scripts/check-uitest-overrides.sh` exits clean; a fresh
  `cd stats && make ui-test-up` followed by `make ui-test-down` completes.

**Build:** green
**Files:** `stats/internal/config/**`, `stats/cmd/flow/**`, `stats/cmd/flowd/**`,
`stats/Makefile`, `stats/docker-compose.yml`, `stats/launchd/*.plist`, `stats/README.md`,
`.myflow/project.md`, `scripts/*.sh`
**Allowed-collateral:** `stats/**/*.go`, `stats/*.go`, `skills/**/*.md`,
`scripts/check-references.sh`
**Tests:** no new test — the existing `stats/internal/config` suite already covers precedence and
defaults for each variable, and it moves to the new names with the code.
**Regression:** reverting this task restores the `MYFLOW*_` names; the config suite fails on the
names it looks up, and `check-uitest-overrides.sh` fails on the `UITEST_*` assignments whose
surrounding variables this task touches.
**Baseline:** before=16 after=16 Go packages, all passing.
<!-- measured: cd stats && go test ./... -count=1 @ 07acd2e (before this change) -->
**Commit:** `refactor(stats): rename the MYFLOW and MYFLOWD environment variables to FLOW and FLOWD`

- [x] 7. Rename the container, volume, database, role and DSN defaults

  In `stats/docker-compose.yml`: service and `container_name` become `flow-postgres`; `POSTGRES_USER`,
  `POSTGRES_PASSWORD` and `POSTGRES_DB` become `flow`; the named volume becomes `flow-pgdata`. Keep
  the host port and the `/var/lib/postgresql` mount path exactly as they are, and keep the comment
  explaining why the mount is not `.../data` — that comment records a real incompatibility, not a
  preference.

  Update the default DSN everywhere it is written out: `stats/internal/config`, `stats/Makefile`,
  `stats/README.md`, and `.myflow/project.md`'s `## workspace isolation` table, whose `Default` and
  `In a workspace` columns both carry it. The per-workspace derived database becomes `flow_<id>` and
  the UI-test database becomes `flow_uitest`; update `scripts/workspace.sh` and the `ui-test-*`
  Makefile targets accordingly.

  Rename `stats/launchd/com.tweety53.myflowd.plist` to `com.tweety53.flowd.plist` and update its
  `Label`, its program path and its log paths.

  **This task changes no running container and no live database.** It changes the declarations. The
  operator's own steps are written in task 19 and run by the operator.

  Verify: `cd stats && go test ./... -race -count=1` passes against the **existing** container — the
  suite reaches the store through the DSN the environment supplies, so a declaration change does not
  require the container to be renamed first; `scripts/test-workspace.sh` passes;
  `scripts/check-workspace-isolation.sh` exits clean; `docker compose -f stats/docker-compose.yml
  config` parses.

**Build:** green
**Files:** `stats/docker-compose.yml`, `stats/launchd/com.tweety53.flowd.plist`,
`stats/internal/config/**`, `stats/Makefile`, `stats/README.md`, `scripts/workspace.sh`,
`.myflow/project.md`
**Allowed-collateral:** `stats/internal/config/config.go`, `stats/cmd/flow/journal.go`,
`scripts/test-workspace.sh`
**Tests:** no new test — `scripts/test-workspace.sh` already covers the derived database and bucket
naming, and it moves to the `flow_` prefix with the code.
**Regression:** reverting this task restores the `myflow-postgres` declarations and the `myflow_`
database prefix; `scripts/test-workspace.sh` fails on the derived database name, and
`check-workspace-isolation.sh` fails on the `## workspace isolation` table it reads.
**Baseline:** before=16 after=16 Go packages, all passing.
<!-- measured: cd stats && go test ./... -count=1 @ 07acd2e (before this change) -->
**Commit:** `refactor(stats): rename the postgres container, volume, database and DSN defaults`

- [x] 8. Rename the on-disk state root

  `stats/internal/fallback`'s `DefaultStateRoot` becomes `/Users/tweety53/Agents/flow/state`. Update
  `stats/internal/fallback/journal_test.go`'s comment about the path it `mkdir -p`s, and
  `skills/flow-contracts/state-file.md`'s two path lines (that file is renamed by task 12; if task 12
  has not landed yet, edit it at its current path and let task 12 carry the edit forward).

  **Measured before the edit, and the number belongs in task 19's runbook:** the old root holds 56
  records and 2 journal files, both of them empty — so no unreplayed write is at risk today and the
  operator's move is a formality rather than a rescue. That is true of this machine at this moment,
  not of the mechanism: a journal is non-empty exactly when a write could not reach the store, so the
  operator re-measures rather than trusting this line.
  <!-- measured: ls ~/Agents/myflow/state/*/*.json | wc -l and ls ~/Agents/myflow/state/*/*.journal | wc -l, plus a size check on each journal @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->

  **This is the one rename that can strand data.** The fallback directory holds records and journal
  entries a failed write left behind, and the daemon replays the journal at startup. A rename with no
  move leaves any unreplayed entry at the old path, where nothing will ever look for it again. The
  move is an operator step because the daemon must be down for it; task 19 writes it out, and this
  task's own commit message says so.

  Verify: `cd stats && go test ./... -race -count=1` passes — in particular the `fallback` package,
  whose tests construct their own root under `TMPDIR` and so are unaffected by the constant; then
  confirm by hand that the constant's value is the only occurrence changed, with
  `grep -rn 'Agents/myflow' stats/` returning nothing.

**Build:** green
**Files:** `stats/internal/fallback/statefile.go`, `stats/internal/fallback/journal_test.go`,
`skills/myflow-contracts/state-file.md`, `skills/myflow-contracts/pipeline.md`,
`stats/cmd/flow/state_test.go`

The contract files are named at their **current** path — task 12 renames the directory and carries
these edits forward. `pipeline.md` and `state_test.go` were added to this field during
implementation: both carry a further live citation of the old root, and this task's own verification
step (`grep -rn 'Agents/myflow'` returning nothing) cannot pass while either survives, so they belong
to this task rather than to a later sweep.
**Tests:** no new test — the `fallback` suite already covers write, read and replay against a
temporary root; the constant is not what those tests exercise, which is why the stranding risk is
called out in prose rather than left to a test to catch.
**Regression:** reverting this task restores the old root; nothing fails, which is exactly why the
operator step in task 19 is the real safeguard and this note exists.
**Baseline:** before=16 after=16 Go packages, all passing.
<!-- measured: cd stats && go test ./... -count=1 @ 07acd2e (before this change) -->
**Commit:** `refactor(stats): move the fallback state root to Agents/flow`

- [x] 9. Rename the remaining Go identifiers and strings

  Sweep `stats/` for every remaining occurrence: package-level identifiers, type and function names,
  struct tags, log messages, error strings, HTTP route prose, test fixture names and test helper
  names.

  **Three exclusions, each of which a careless sweep will get wrong:**

  - Every `/myflow-*` **string literal**, wherever it occurs. They are values already in the live
    `stage_runs` table. They are **not** constants in `stats/internal/stages/names.go` — that file
    declares only `/flow` — but bare literals spread widely, production code included
    (`internal/api/stats.go`, `internal/fallback/statefile.go`, `internal/client/client.go`,
    `internal/records/render.go`, `internal/store/changes.go`, `cmd/flow/state.go`). **Classify each
    occurrence**, per `design.md`'s corrected `retired-command-constants-are-data`: stored-data
    fixtures stay, live filters selecting on a command name stay, and doc comments describing what a
    retired command did stay. None is renamed; the reason differs per occurrence, and a sweep that
    cannot tell them apart will rename the wrong one.
  - The `mf-` session-token prefix and every token literal carrying it, in both the CLI's
    `validateSessionToken` and the daemon's `validateSessionTokenShape`.
  - Any string that quotes a historical path under `openspec/` or `docs/superpowers/`, since those
    trees are not renamed.

  After the sweep, `grep -rn myflow stats/` should return only lines covered by one of those three
  exclusions. Read the remaining list rather than assuming it is empty, and record what it contains
  in the commit body.

  **This task broke the ui-test teardown, and the way it broke is the whole point of KAN-289.**
  Renaming `internal/pidfile.Path`'s output from `myflowd-<port>.pid` to `flowd-<port>.pid` moved the
  **writer** while `stats/Makefile`'s `UITEST_PIDFILE` — the **reader** — still named the old path, so
  `make ui-test-down` could not find the daemon it had just started and refused to drop
  `flow_uitest`. Both sides were green the entire time: `internal/pidfile`'s own Go test asserts the
  new name, and `scripts/test-make-build.sh` asserts the Makefile's old one. Each test agreed with the
  file it covered, and nothing compared the two. Found only because task 10 actually ran
  `make ui-test-up` / `make ui-test-down` against the real disposable stack rather than trusting a
  green suite.

  The fix — folded into this task's own commit — moves the reader and the harness's assertion to
  `flowd-<port>.pid`, and corrects a Makefile comment that described the old path as fact.

  **A second thing was folded in after the per-task review: a test that pins
  `stages.SyntheticChangeUpdatedBy` to its stored literal.** The review mutation-tested that constant
  and found a **surviving mutant** — renaming it to `flow stage begin (synthetic)` left `go vet`,
  `gofmt` and the whole sixteen-package race suite green, because every existing test refers to the
  constant *symbolically*, so producer and consumer agree with each other whatever it says while both
  disagree with the `changes.updated_by` value the live store holds. `synthetic_test.go` now asserts
  the literal and fails loudly, with the reason in its own doc comment. Verified by re-running the
  mutation against the new test. **Whenever
  this task renames a value that crosses a process boundary, find its reader and move both in the
  same commit**; a passing test on either side alone proves nothing about the pair.

  Verify: `cd stats && go vet ./... && gofmt -l .` clean; `cd stats && go test ./... -race -count=1`
  passes in all sixteen packages; `scripts/test-make-build.sh` passes; a real
  `make ui-test-up` / `make ui-test-down` cycle completes and the pidfile appears at the path the
  Makefile reads; the surviving-occurrence list matches the three exclusions and nothing else.

**Build:** green
**Files:** `stats/internal/store/store.go`, `stats/internal/api/server.go`,
`stats/internal/harvest/watcher.go`, `stats/health_test.go`, `stats/Makefile`,
`scripts/test-make-build.sh`
**Allowed-collateral:** `stats/*.go`

**A glob belongs in `Allowed-collateral:`, never in `Files:` — the two fields are matched
differently, and the difference is invisible until it bites.** `check-task-commit-fields.py` globs
`Allowed-collateral:` with `fnmatch` and compares `Files:` as literal strings, so a pattern written
in `Files:` matches no file at all. An earlier draft of this task declared `stats/**/*.go` in
`Files:` and the guard rejected every single file the commit touched.

Two further properties of that matching, both load-bearing here: `fnmatch`'s `*` **does** cross `/`,
so `stats/*.go` covers every Go file at any depth under `stats/`; but `stats/**/*.go` requires a
literal `/` after the `**`, so it does **not** cover `stats/health_test.go` sitting directly under
`stats/`. Tasks 6 and 7 each tripped that second property.
<!-- measured: check-task-commit-fields.py:829 (fnmatch over allowed_collateral only), and the guard re-run against this task's own commit @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->
**Tests:** no new test — the sixteen existing package suites are the surface, and a rename that
breaks a route, a struct tag or an error string fails one of them.
**Regression:** reverting this task restores the old identifiers across the daemon and CLI; the
API and store suites fail on the route and column names they assert.
**Baseline:** before=16 after=16 Go packages, all passing.
<!-- measured: cd stats && go test ./... -count=1 @ 07acd2e (before this change) -->
**Commit:** `refactor(stats): rename the remaining myflow identifiers to flow`

- [x] 10. Rename the SPA's TypeScript identifiers

  In `stats/web`: the package name `myflow-stats-web` becomes `flow-stats-web`, and every component,
  hook, type, fixture and test name carrying `myflow` is renamed. Leave any string that displays a
  retired command name — the dashboard renders `/myflow-do` for historical stage runs, and those
  rows are real.

  Verify: `cd stats/web && npx tsc -b` clean; `cd stats/web && npm test` passes 9 files and 157
  tests; `cd stats && make build` still embeds the SPA build output.
  <!-- measured: cd stats/web && npm test @ 07acd2e (before this change) -->

**Build:** green
**Files:** `stats/web/package.json`, `stats/web/src/**`
**Allowed-collateral:** `stats/web/package-lock.json`
**Tests:** no new test — the nine existing SPA test files are the surface.
**Regression:** reverting this task restores the old names; `npx tsc -b` fails on the identifiers
the tests import.
**Baseline:** before=157 after=157 tests in `stats/web`.
<!-- measured: cd stats/web && npm test @ 07acd2e (before this change) -->
**Commit:** `refactor(stats): rename the SPA's myflow identifiers to flow`

- [x] 11. Rename `.myflow/` to `.flow/` and give the resolver its actionable error

  `git mv .myflow .flow`. Update every reader: the Go configuration resolver, every guard under
  `scripts/` that reads `<project>/.myflow/project.md`, and every skill, rule, command and contract
  file that names the path.

  **Add the actionable error**, per `design.md`'s `dotmyflow-hard-cutover`. When the resolver finds
  `<project>/.myflow/` and no `<project>/.flow/`, it reports — naming the project root and the exact
  rename to perform — and exits non-zero. It never falls back to the old path, never reads both, and
  never silently reports "no project configuration". Two other projects on this machine carry
  `.myflow/`, so a bare not-found would send the operator debugging the wrong layer.

  Add a test for that branch: a fixture project with `.myflow/` and no `.flow/` produces the error,
  and one with both present reads `.flow/` without consulting the other.

  Verify: `cd stats && go test ./... -race -count=1` passes, including the new resolver cases; the
  whole `## lint` list runs clean; `scripts/check-workspace-isolation.sh` and
  `scripts/check-references.sh` exit clean.

**Build:** green
**Files:** `.flow/project.md`, `AGENTS.md`, `CLAUDE.md`, `README.md`, `rules/agent-baseline.md`,
`rules/kotlin-backend-development-standard.mdc`, `rules/lint-fix-priority.mdc`,
`rules/myflow-manual-review.mdc`, `scripts/check-cleanup-complete.sh`,
`scripts/check-contract-budget.sh`, `scripts/check-guard-symlinks.sh`,
`scripts/check-installed-citations.py`, `scripts/check-plan-provenance.sh`,
`scripts/check-references.sh`, `scripts/check-task-build-green.sh`,
`scripts/check-task-commit-fields.py`, `scripts/check-task-commit-fields.sh`,
`scripts/check-workspace-isolation.sh`, `scripts/lib/owned-corpus.sh`, `scripts/lib/plan_grammar.py`,
`scripts/prepare-workspace.sh`, `scripts/test-check-cleanup-complete.sh`,
`scripts/test-check-contract-budget.sh`, `scripts/test-check-guard-symlinks.sh`,
`scripts/test-check-installed-citations.sh`, `scripts/test-check-normative-inventory.sh`,
`scripts/test-check-references.sh`, `scripts/test-check-task-commit-fields.sh`,
`scripts/test-check-workspace-isolation.sh`, `scripts/test-prepare-workspace.sh`,
`scripts/test-setup.sh`, `scripts/workspace.sh`, `setup.sh`, `skills/flow-status/SKILL.md`,
`skills/flow/principles-reviewer-prompt.md`, `skills/flow/review-panel.md`,
`skills/flow/verify-and-handoff.md`, `skills/myflow-contracts/SKILL.md`,
`skills/myflow-contracts/finish-contract.md`, `skills/myflow-contracts/jira-followups.md`,
`skills/myflow-contracts/jira-integration.md`, `skills/myflow-contracts/pipeline.md`,
`skills/myflow-contracts/project-configuration.md`, `stats/README.md`

Every entry above is the literal touched path, not a glob — task 9's own lesson, restated by
necessity rather than repeated: `check-task-commit-fields.py` compares `Files:` as literal strings,
so a pattern written there matches no file at all. `commands/*.md` and `commands-claude/*.md` from
this task's original draft are dropped rather than kept as dead entries: neither directory carries a
`.myflow` occurrence, so this task touches neither.

**Allowed-collateral:** `stats/**/*.go`

Covers `stats/internal/config/projectconfig.go` and `stats/internal/config/projectconfig_test.go` —
both new files. Measured against the plan: this task's draft named "the Go configuration resolver"
as an existing reader to update. No Go code anywhere in this repository read `.myflow/project.md`
before this commit — `stats/internal/config` resolved only `flowd`'s environment-variable
configuration — so `ProjectConfigPath` is new code the task's own `stats/internal/config/**` `Files:`
entry called for, not an update to something that already existed.
**Tests:** two new resolver cases — `.myflow/` present with no `.flow/` reports the actionable
error and exits non-zero; both present reads `.flow/`. Mirrored as real-subprocess fixture cases in
`scripts/test-check-workspace-isolation.sh`, `scripts/test-check-cleanup-complete.sh` and
`scripts/test-setup.sh`, since those three guards each resolve an arbitrary project root's own
`.flow/project.md` independently of the Go resolver and are exactly the readers the task's own trap
warning names.
**Regression:** reverting this task restores `.myflow/`; the two new resolver cases fail, and every
guard reading the project configuration finds no file at the path it was renamed to.
**Baseline:** before=16 after=16 Go packages, all passing; the config package gains two cases.
<!-- measured: cd stats && go test ./... -count=1 @ 07acd2e (before this change) -->
**Commit:** `refactor(project-config): rename .myflow to .flow and refuse the stale directory loudly`

- [x] 12. Rename the contract skill directory and repoint every citation

  `git mv skills/myflow-contracts skills/flow-contracts`. Repoint every citation to it —
  `skills/myflow-contracts/pipeline.md`, `.../state-file.md`, `.../finish-contract.md` and the rest —
  across `skills/`, `rules/`, `commands/`, `commands-claude/`, `CLAUDE.md`, `AGENTS.md`,
  `README.md` and `stats/`, which carries at least one such path in a Go string.

  Update the skill's own `SKILL.md` name and description, `setup.sh`'s install list, and
  `scripts/check-contract-budget.sh`'s `budgets()` table, whose keys are paths relative to the
  repository root and every one of which changes here.

  Verify: `scripts/check-references.sh` exits clean — this is the guard that catches a citation
  pointing at a path that no longer resolves, and it is the whole safety net for this task;
  `scripts/check-installed-citations.sh` exits clean; `scripts/check-contract-budget.sh` exits clean;
  `scripts/check-guard-symlinks.sh` exits clean; a sandboxed `HOME="$(mktemp -d)" ./setup.sh global`
  installs the renamed directory.

**Build:** green
**Files:** `CLAUDE.md`, `AGENTS.md`, `README.md`, `scripts/check-contract-budget.sh`
**Allowed-collateral:** `skills/*`, `rules/*`, `commands/*`, `commands-claude/*`, `scripts/*`,
`stats/*`, `.flow/*`

**The globs live in `Allowed-collateral:` because that is the only field matched with `fnmatch`;
`Files:` is compared as literal strings.** This is the third time on this branch that a drafted
`Files:` field carried a glob and the guard rejected every path the commit touched — tasks 9 and 11
hit it before this one. `fnmatch`'s `*` crosses `/`, so `skills/*` covers the whole subtree and no
`**` form is needed or wanted.

`setup.sh` is deliberately **absent** from both fields: it needed no edit at all. It discovers skill
directories dynamically from `basename "$skill_dir"` and carries no hardcoded skill-name list, so the
rename flows through it untouched — the earlier draft of this task named "setup.sh's install list",
which does not exist.
<!-- measured: git show 6883fd5 --name-only, reconciled against check-task-commit-fields.sh's report, and grep for a hardcoded skill list in setup.sh @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->
**Tests:** no new test. This task is verified by running guards over the real corpus rather than by
adding a case: the references guard and the installed-citations guard both police exactly this, and a
sandboxed installer run into a throwaway HOME confirms the installed tree carries the renamed
directory with every symlink resolving when followed.

No guard or harness is named in backticks above, deliberately. This field's parser reads every
backtick-quoted token as a **declared test** and then requires it to appear in the commit's own diff;
naming a guard that correctly needed no edit makes the guard report a test that is missing. Prose,
not paths, is the right shape for a task whose verification is "run the existing checks".
**Regression:** reverting this task restores the old directory name while leaving citations
repointed, so `check-references.sh` reports every repointed citation as unresolvable.
**Baseline:** not applicable — no test suite exists for these files; verified by the five guard
runs and the sandboxed installer run named above.
**Commit:** `refactor(flow-contracts): rename the myflow-contracts skill directory to flow-contracts`

- [x] 13. Rename the always-on rule stub and its installed managed block

  `git mv rules/myflow-manual-review.mdc rules/flow-manual-review.mdc`. Update its own heading and
  body, `setup.sh`'s managed-block rendering and its rule list, and every citation to it — `CLAUDE.md`
  and `AGENTS.md` both name it.

  The managed block that `setup.sh global` writes into `~/.claude/CLAUDE.md` is derived from these
  `.mdc` files, and `scripts/check-installed-rules.sh` compares the two. Both sides move together in
  this commit or the guard reports the mismatch.

  **`scripts/check-installed-rules.sh` gives this task ZERO coverage when run inside a worktree**,
  so naming it as a verify step is misleading. In a linked worktree it short-circuits — printing that
  the install tracks the main checkout and there is nothing to check — and exits 0 before comparing
  anything. All implementation happens in a worktree, so every run of it here is a vacuous pass. Real
  coverage comes from pointing it at a sandboxed install (`CHECK_INSTALLED_RULES_HOME=<sandbox>`),
  which is what the task 13 review did to mutation-test the source-versus-install pairing: reverting
  the source rename while leaving the install in place made it report six violations and exit 1.
  Found by running the guard and reading its output rather than taking its clean exit as coverage.

  Verify: `scripts/check-installed-rules.sh` pointed at a sandboxed install exits clean — a bare run
  inside the worktree proves nothing; `scripts/test-setup.sh` passes;
  `scripts/check-references.sh` exits clean; a sandboxed `HOME="$(mktemp -d)" ./setup.sh global`
  writes the renamed rule and its managed block.

**Build:** green
**Files:** `rules/flow-manual-review.mdc`, `setup.sh`, `CLAUDE.md`, `AGENTS.md`, `scripts/check-contract-budget.sh`,
`skills/flow-contracts/SKILL.md`
**Tests:** no new test. Verified by running the existing installer harness and the installed-rules
comparison over the real corpus, plus a sandboxed installer run into a throwaway HOME whose written
managed block was read back and whose installed symlink was followed to its target.

No harness is named in backticks above: this field's parser reads every backtick-quoted token as a
declared test and then requires it to appear in this commit's diff, so naming a harness that
correctly needed no edit makes the guard report a missing test.
**Regression:** reverting this task restores the old rule filename while `setup.sh` still names the
new one, so the installer writes no rule and `check-installed-rules.sh` reports the absence.
**Baseline:** not applicable — no test suite exists for these files; verified by the four runs
named above.
**Commit:** `refactor(rules): rename the myflow-manual-review rule stub to flow-manual-review`

- [x] 14. Rename the guard-contract names across the contracts and guard scripts

  The contract-name vocabulary — `myflow-task-commit-fields`, `myflow-planning-effort`,
  `myflow-review-panel-economics` and the rest as they appear in **live** prose and in guard script
  headers — becomes the `flow-` prefixed form.

  **Do not rename the directories under `openspec/specs/`.** Those are frozen. Where live prose cites
  a frozen spec by path — `skills/flow-contracts/state-file.md` cites
  `openspec/specs/myflow-planning-effort/spec.md` by name so `check-references.sh` keeps the pointer
  checked — the path stays exactly as it is, because the file it names still has that name. Renaming
  the citation while the target keeps its name is how a checked pointer becomes a broken one.

  **A third class emerged during implementation, beyond vocabulary and frozen-tree paths: a string
  whose exact bytes are an INPUT to a documented computation.**
  `skills/flow-contracts/workspace-isolation.md`'s worked example derives a workspace id from the
  change name `kan-15-parallel-myflow-do-task-lanes` and records the result, `kan-15-55a6`, in a
  `verified:`-tagged fenced block. Renaming the string inside it changes the SHA-256 the id is built
  from — the same block would then claim an id that its own recipe no longer produces. Nothing checks
  a worked example's arithmetic, so this would have shipped as a canonical contract documenting a
  false result. Recomputed both ways to confirm: the recorded id reproduces from the original name,
  and the renamed form yields a different digest.
  <!-- measured: the contract's own recipe re-run over both spellings of the name @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->

  **Before renaming any string, ask whether something computes over it** — not only whether something
  reads it. An identifier, a stored value and a hash input fail in three different ways, and only the
  first is a rename at all.

  Verify: `scripts/check-references.sh` exits clean — including every citation into the frozen tree;
  `scripts/check-vocabulary.sh` exits clean; the whole `## lint` list runs clean; and any worked
  example whose inputs this task touched still reproduces its own recorded output.

**Build:** green
**Files:** `skills/flow-contracts/build-green.md`, `skills/flow/implement.md`,
`skills/flow/brainstorm.md`, `scripts/check-task-commit-fields.py`
**Tests:** no new test. Verified by running the references guard, the vocabulary guard and the whole
lint list over the real corpus, and by recomputing the workspace-id worked example from its own
recipe to confirm the strings this task left alone are still the ones it derives from.

No guard is named in backticks above: this field's parser reads every backtick-quoted token as a
declared test and requires it in the diff, and none of these guards needed editing.
**Regression:** reverting this task restores the old contract names in prose while guard headers
carry the new ones; `check-references.sh` reports the citations that no longer resolve.
**Baseline:** not applicable — no test suite exists for these files; verified by the three runs
named above.
**Commit:** `refactor(flow-contracts): rename the myflow- contract names to flow-`

- [ ] 15. Rename the remaining prose vocabulary across the live corpus

  Sweep what is left: `skills/`, `rules/`, `commands/`, `commands-claude/`, `README.md`,
  `CLAUDE.md`, `AGENTS.md`, `setup.sh`, the remaining prose in `scripts/` headers, **and
  `stats/README.md`'s general CLI/daemon vocabulary**. The pipeline is "flow", the CLI is `flow`, the
  daemon is `flowd`, the project configuration is `<project>/.flow/project.md`.

  **Repoint the citations to files this change renamed or that no longer exist**, which the task 13
  review enumerated: `rules/myflow-manual-review.mdc` (cited by `README.md`, `commands/flow.md`,
  `commands/flow-status.md`, `commands-claude/flow.md`, `commands-claude/flow-status.md`,
  `skills/README.md`, `skills/flow-contracts/pipeline.md`) and the `skills/myflow-do/SKILL.md`,
  `skills/myflow-fast/SKILL.md`, `skills/myflow-finish/SKILL.md`, `skills/myflow-start/SKILL.md`
  family, whose directories were consolidated into `skills/flow/` before this change and which are
  cited widely. **Those last ones are pre-existing staleness, not this change's doing** — repoint a
  citation that is genuinely pointing at today's file, and leave a sentence that is describing what a
  retired command's skill used to say, per this task's historical-artifact rule above.

  `stats/README.md` is named here because no earlier task covered it: task 6 claimed it for
  environment variables and task 7 for container and DSN values, each editing only its own lines, and
  neither owned the general prose ("the `myflow` CLI's default daemon address", "`myflow state` /
  `myflow stage`"). Task 7 found the gap and flagged it rather than widening its own scope.

  **Also fix `stats/Makefile`'s `ui-test-up`/`ui-test-down` comment block**, which still describes a
  `myflow_kan_999_fake` workspace database, "workspace.sh's own `myflow_` prefix check", and "an
  active `/myflow-do` worktree". Task 7 changed the derivation those lines describe to `flow_<id>`,
  which makes all three statements **false**, not merely stale. **Nothing guards them**:
  `check-vocabulary.sh`'s `DEFAULT_TARGETS` excludes `stats/` entirely, and before this task no
  later task declared `stats/Makefile` again — so without this paragraph they would survive the whole
  rename. Found by the tasks 6-7 review, which is also why the file is now in this task's `Files`.
  Keep the comment's reasoning about why `override` is needed; only the names it cites change.

  **Also rename `scripts/workspace.sh`'s `CONTAINER` and `SUPERUSER` values** to `flow-postgres` and
  `flow`. Task 7 deliberately left them, because at that point they addressed a real, running
  container it was forbidden to touch. They are renamed here so the branch ends internally
  consistent — with the ordering consequence `design.md`'s cutover step 4 states in full: from the
  merge until the operator renames the real container, `workspace.sh` names one that does not exist,
  so no new `/flow` run can isolate a workspace in that window.

  **Leave every reference to a historical artifact exactly as it is**: a path under `openspec/` or
  `docs/superpowers/`, a retired command name in a sentence describing what that command did, and the
  archived change directories under `spectre/changes/archive/`. A sentence like "the content
  `/myflow-start` carried" is a true statement about a command that had that name; rewriting it makes
  the sentence false. The retired-name exclusion is per-occurrence, exactly as task 9 states it.

  After the sweep, `grep -rn myflow skills scripts rules commands commands-claude README.md
  CLAUDE.md AGENTS.md setup.sh` should return only lines covered by that exclusion. Read the list,
  do not assume it, and record it in the commit body.

  Verify: the whole `## lint` list runs clean, split across more than one invocation;
  `scripts/check-markdown-integrity.py` exits clean; `scripts/check-references.sh` exits clean;
  `scripts/check-installed-citations.sh` exits clean.

**Build:** green
**Files:** `README.md`, `CLAUDE.md`, `setup.sh`, `stats/README.md`, `stats/Makefile`

**The globs a draft of this task carried in `Files:` are moved to `Allowed-collateral:` below,
the same fix tasks 9, 11 and 12 each made in their own commit** — `check-task-commit-fields.py`
compares `Files:` as literal strings and only `Allowed-collateral:` with `fnmatch`, so a glob
written in `Files:` matches no real file the commit touches. `AGENTS.md` is deliberately absent
from both fields: its only `myflow` occurrence was the managed-block delimiter, correctly left
unrenamed (see the finding below), so this task's commit does not touch it at all.
**Allowed-collateral:** `skills/*`, `rules/*`, `commands/*`, `commands-claude/*`, `scripts/*`

**Tests:** no new test — the configured lint list over the real corpus is the surface.
**Regression:** reverting this task restores `myflow` throughout the prose while the code, the
directories and the guard names carry `flow`; `check-references.sh` and
`check-installed-citations.sh` report the citations that no longer resolve.
**Baseline:** not applicable — no test suite exists for these files; verified by the four runs
named above.

**Found during implementation, load-bearing enough to record here:** `setup.sh`'s managed-block
delimiters (`<!-- myflow:begin -->` / `<!-- myflow:end -->`), its `.myflow.bak` backup suffix, and
`.myflow`-vs-`.flow` existence checks in `scripts/check-workspace-isolation.sh`,
`scripts/check-cleanup-complete.sh` and `setup.sh`'s `install_project_standards` are excluded from
this sweep and left exactly as they were. The delimiter and backup suffix are literal values
`setup.sh` writes into and `grep -Fx` compares against already-installed real files —
`~/.claude/CLAUDE.md` on this machine still carries `<!-- myflow:begin -->` and a real
`CLAUDE.md.myflow.bak` sits beside it, both confirmed by reading them — so renaming the delimiter
in source would make the next real `setup.sh global` run fail to find its own prior block and
silently append a second, duplicate one, exactly the failure `require_grep_ok`'s own comment
warns can never converge. The `.myflow`-vs-`.flow` checks are the hard-cutover detection logic
itself: it must keep testing for the literal old path to detect it, so renaming what it tests for
would make the guard stop detecting the exact condition it exists to catch. A first pass renamed
all of these blindly; found by running the sandboxed installer and the guards' own test suites
against a real fixture, not by reading the diff, and reverted by hand. No task in this plan's
Part 2 names this delimiter/backup mechanism, so it is recorded here as a gap: task 19's runbook
should either add a sixth step for it or state explicitly that it is never renamed.

Two further findings, same root cause (a blind substitution rather than a per-site read):
`scripts/check-installed-citations.sh` gained two duplicate `declare_if_present` calls for
`commands/flow-research.md` and `commands-claude/flow-research.md` (the pre-existing `myflow-*`
row and the already-current `flow-*` row collapsed onto the same string), which
`coverage_declare` rejects outright — deduplicated. `scripts/test-check-vocabulary.sh`'s case 6.3
fixture renamed `check-vocabulary.sh`'s own retired-literal `'myflow-full'` to `'flow-full'`,
so the guard no longer caught the fixture it was supposed to — reverted to the guard's real,
unrenamed literal.

**stats/Makefile's `UITEST_BIN`/`UITEST_TRANSCRIPTS_DIR`/`UITEST_STATE_DIR`/`UITEST_LOG` were
renamed to the `flow-uitest-*` prefix**, reversing this task's own first-pass decision to leave
them: `stats/internal/pidfile/pidfile.go`'s truncation-example comment and
`pidfile_test.go`'s `uitestExecutable` constant were already renamed to `/tmp/flow-uitest-flowd`
by task 9's own commit (its message calls it "a pre-existing staleness bug"), which only makes
sense if `UITEST_BIN` becomes that same value — task 9 could not touch the Makefile itself (Go
identifiers were its scope), so this task closes that loop. `scripts/test-make-build.sh`'s pidfile
fixture, which had been blindly swept to `/tmp/flow-uitest-flowd` and then reverted by hand
alongside the container/DSN cluster, was corrected a second time to match — it was right the
first time.

**Commit:** `docs(flow): rename the myflow vocabulary to flow across the live corpus`

- [x] 16. Switch the documented Jira label taxonomy to the flow- prefix

  In the finish contract's step 9 angle-to-label table — canonical for it — and in
  `skills/flow-contracts/jira-integration.md` where it is cited, the labels become `flow-fix`,
  `flow-cost`, `flow-improvement`, `flow-automation` and `flow`.

  **Record the consequence in the contract text itself**, per `design.md`'s `jira-labels-docs-only`:
  existing issues keep the labels they carry, so the board holds both taxonomies indefinitely and a
  label-based search has to match either form. This is a stated cost, not an omission — a reader who
  finds only `flow-improvement` issues and concludes the older ones do not exist has been misled by
  the contract.

  Do not relabel any existing issue. That is an outward-facing bulk write over a board this
  repository does not own.

  **The guard must read BOTH label eras, and that is the whole difficulty of this task.**
  `check-self-review-report.sh` matches the angle labels as literal backtick-quoted strings in the
  reports under `<project>/docs/self-review/`. Twelve of those reports are actively checked and every
  one was written when the labels carried the `myflow-` prefix. Renaming the guard's `ANGLE_LABELS`
  alone makes it report **166 violations across all twelve** — measured by doing exactly that and
  running it over the real corpus. Rewriting those reports to match would edit the record of what past
  runs found in order to satisfy a present-day spelling, so the guard gains a positionally-aligned
  `LEGACY_ANGLE_LABELS` set instead.

  **`cur` must hold the spelling the report actually used, not the canonical one.** The guard also
  compares each finding line's own `[label]` against its section, so canonicalising `cur` makes a
  legacy report contradict itself — `finding line label `myflow-fix` does not match its section
  `flow-fix`` — which fails three harness cases and one real historical report. The cost is that a
  message pairing `cur` with a neighbouring label drawn from the canonical array can name two eras in
  one sentence; that is accepted, because the alternative tells an author their report is malformed
  when it is not. Both spellings were tried and measured before choosing.

  Verify: `scripts/check-references.sh` exits clean; `scripts/check-self-review-report.sh` exits clean
  **over the real historical corpus**, not merely over the harness's fixtures; the harness passes; and
  deleting the legacy set fails 21 harness cases and the real corpus alike (mutation-tested, it does).

**Build:** green
**Files:** `skills/flow-contracts/finish-contract.md`, `scripts/check-self-review-report.sh`,
`scripts/test-check-self-review-report.sh`
**Tests:** no new test — `scripts/test-check-self-review-report.sh` already covers the guard's
angle-label handling and moves to the new labels with it.
**Regression:** reverting this task restores the `myflow-` labels in the contract while the guard
carries the new ones; `check-self-review-report.sh` reports every angle heading as unrecognised.
**Baseline:** not applicable — no test suite exists for these files; verified by the two runs
named above.
**Commit:** `docs(flow-contracts): switch the documented Jira label taxonomy to the flow prefix`

- [x] 17. Add the retired myflow literals to the vocabulary guard

  Add to `scripts/check-vocabulary.sh`'s retired-literal list exactly five spellings: `myflowd`,
  `myflow-postgres`, `myflow-contracts`, `MYFLOWD_` and `MYFLOW_`.

  **The plan originally listed `myflow` and `.myflow` too, and banning them would have been wrong.**
  The guard's own scan set carries the word on 852 lines.
  <!-- measured: grep -rn myflow over check-vocabulary.sh's DEFAULT_TARGETS @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->
  Almost every one
  is legitimate — 478 are `/myflow-*` command literals that are values in the live `stage_runs`
  table, 115 are the self-review angle labels whose legacy spelling
  `check-self-review-report.sh` must keep recognising, 57 are the `.myflow` that the hard-cutover
  detection exists to look for, and 15 are the managed-block delimiters `setup.sh` compares
  byte-for-byte against files already in the operator's home. A flat ban needs a suppression marker
  on roughly eight hundred lines, which this guard's own `myflow-fast` note calls teaching the guard
  to lie.
  <!-- measured: grep -rn myflow over the guard's DEFAULT_TARGETS, bucketed by why each hit is legitimate @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->

  **Each of the five was counted before being added and measured zero live occurrences**, so a hit is
  drift rather than history and needs no marker beside it. Counting them exposed three real
  stragglers that earlier tasks missed, all fixed here: `CLAUDE.md`'s protected-service paragraph
  still named `myflowd`, `myflow-postgres` and the `myflow` database; and
  `scripts/test-workspace.sh` still hardcoded `CONTAINER="myflow-postgres"` / `SUPERUSER="myflow"`
  while `workspace.sh` had moved to the renamed pair — a writer/reader split of exactly the shape
  this branch has now hit four times.

  **`__pycache__` is pruned from the scan, and that is a pre-existing scope bug this task exposed.**
  The guard enumerated compiled `.pyc` bytecode, which embeds its source's docstrings verbatim, so a
  stale one reported retired vocabulary from a version of the source that no longer exists — at a
  `file:line` pointing inside a binary nobody can edit. It is gitignored and regenerated; it was
  never part of the corpus. Scope the scan to the live corpus and
  exclude `openspec/`, `docs/superpowers/` and `spectre/changes/archive/` — this is the guard's
  `DEFAULT_TARGETS`, and that variable is stated in the script as the one place the invocation lives,
  so the exclusion goes there rather than into each call site.

  Extend the guard's header with the same honesty the existing one carries: a green run means these
  exact strings are absent from the scanned set, and nothing wider. It does not prove the rename is
  complete — a paraphrase, or a spelling this list does not carry, passes clean.

  Add harness cases to `scripts/test-check-vocabulary.sh` for each new literal, and one asserting a
  file under an excluded tree carrying `myflow` does **not** trip the guard — that exclusion is
  load-bearing and would otherwise fail the moment the guard ran over the archives.

  Verify: `scripts/test-check-vocabulary.sh` prints `all cases passed`;
  `scripts/check-vocabulary.sh` exits clean over the real corpus — which it can only do once tasks 5
  through 16 have all landed.

**Build:** green
**Files:** `scripts/check-vocabulary.sh`, `scripts/test-check-vocabulary.sh`,
`scripts/test-workspace.sh`, `CLAUDE.md`
**Tests:** eleven new cases — one per retired literal (five), one per live replacement spelling
(five) proving the ban is not broad enough to match the renamed corpus, and one proving a stale
`.pyc` under a pruned `__pycache__` is not scanned.
**Regression:** reverting this task removes the literals, so `myflow` can regrow anywhere in the
live corpus with every configured check green; the excluded-tree case is the one that fails loudly
if the scope is later widened over the archives.
**Baseline:** before and after counts are read from the harness's own `all cases passed` line;
the file gains eight cases.
**Commit:** `feat(scripts): retire the myflow vocabulary in check-vocabulary`

- [ ] 18. Reconcile the contract-budget table with the renamed paths

  Every renamed file changed its key in `scripts/check-contract-budget.sh`'s `budgets()` table, and
  task 12 updated the rows it renamed. Tasks 2 and 3 grew two files inside their existing headroom
  and correctly changed no row. This task audits the whole table: no key
  naming a path that no longer exists, and no owned file without a row.

  The guard resolves the owned corpus through `scripts/lib/owned-corpus.sh`, which
  `check-normative-inventory.sh` also calls, so a file this table misses is a file both guards
  disagree about. Confirm the two agree by running each and comparing the file sets they report.

  Where a row must be added for a file the rename created, set the budget to that file's current size
  plus the margin the table's own rule states. Do not narrow the guard's scope and do not delete a row
  to make it pass.

  Verify: `scripts/check-contract-budget.sh` exits clean; `scripts/test-check-contract-budget.sh`
  passes; `scripts/check-normative-inventory.sh` exits 0 and its file set matches the budget table's.

**Build:** green
**Files:** `scripts/check-contract-budget.sh`
**Tests:** no new test — `scripts/test-check-contract-budget.sh` already covers the table's
parsing and its violation shapes.
**Regression:** reverting this task leaves stale keys naming paths that no longer exist and renamed
files with no row at all; `check-contract-budget.sh` reports the unbudgeted files.
**Baseline:** not applicable — the harness's case count is unchanged; verified by the three runs
named above.
**Commit:** `chore(scripts): reconcile the contract budget table with the renamed paths`

- [ ] 19. Write the operator cutover runbook and record the rename's completeness evidence

  Add a cutover section to `README.md` and expand `stats/README.md`'s operations text with the five
  ordered steps `design.md`'s **The cutover is ordered** section states. Reproduce the order and the
  reason each step depends on the one before it; do not compress the sequence into a single command
  block, and do not present step 4 as something the pipeline performs.

  Write out, as copy-paste commands the **operator** runs:

  ```bash unverified:confirm each command against the renamed compose file and Makefile before publishing
  # 2. build and install the renamed CLI
  cd stats && make build && cp bin/flow "$HOME/.local/bin/flow"

  # 3. rename the project configuration directory, in EVERY consuming project
  git -C <project> mv .myflow .flow

  # 4. stop the daemon, then rename the database and container, then restart
  #    ALTER DATABASE requires zero live connections.
  # 5. reinstall the launchd agent under its new label, then remove the old binary
  ```

  Leave steps 4 and 5 as prose plus the exact commands, with the stop-first precondition stated on
  its own line. State plainly that bringing the daemon back up does not repair a run that already
  fell through to the on-disk journal, and that step 4 should run when no `/flow` run is in flight.
  Include the fallback-state-root move task 8 identified, as part of step 4.

  **Record the completeness evidence in this commit's body**, from two measurements taken over the
  finished branch:

  - `scripts/check-normative-inventory.sh` captured before the first rename edit and after the last.
    The two are **not** expected to be byte-identical — the rename changes the text of normative
    sentences. The assertion is stronger and mechanical: applying the same `myflow`→`flow`
    substitution to the *before* inventory must yield the *after* inventory exactly. Any line that
    does not is a sentence this change altered or lost beyond the rename, and it is restored rather
    than accepted.
  - The surviving-occurrence list from `grep -rn myflow` over the live corpus, which must contain
    only the historical references tasks 9 and 15 enumerated.
  - **A broken-citation sweep, because no guard in this repository performs one.** Enumerate every
    backticked repository-relative path in the live corpus and confirm each resolves to a real file,
    classifying anything that does not as a deliberate fixture example, a path relative to another
    directory, or a genuine stale citation. **`check-references.sh` does not do this**: it resolves
    *headings*, and a backticked path that fails to resolve to a file is silently treated as out of
    scope and skipped. A citation to a renamed or deleted file is therefore structurally invisible
    to it — which matters here, because this change renames files that other files cite by path, and
    several tasks were verified against that guard on the assumption it would catch exactly this.
    Found by the task 13 review, by reading the guard rather than trusting its clean exit.

  Verify: the whole `## lint` list runs clean; the whole `## test` list runs clean, both split across
  more than one invocation; both measurements above are recorded and hold.

**Build:** green
**Files:** `README.md`, `stats/README.md`, `scripts/check-contract-budget.sh`
**Tests:** no new test — this task's assertions are the two corpus-wide measurements named above,
which are recorded rather than automated; neither is a claim a script in this repository checks.
**Regression:** reverting this task removes the runbook, leaving five ordered operator steps —
including the only one that stops the dev workspace's service and storage — recorded nowhere the
operator will find them.
**Baseline:** not applicable — no test suite exists for these files; verified by both configured
lists and the two recorded measurements.
**Commit:** `docs(readme): add the myflow-to-flow cutover runbook`

- [ ] 20. Stop the dashboard suggesting retired commands

  `stats/internal/api/stats.go`'s `nextCommandFor` maps a change's pipeline state to "the command
  that runs next" and returns `/myflow-do` for `STARTED` and `/myflow-finish` for `IN_PROGRESS`.
  Its doc comment describes a three-command pipeline (`/myflow-start` → `/myflow-do` →
  `/myflow-finish`) and cites `/myflow-status`. **None of those commands exist.** The pipeline is one
  command, `/flow`, and `/flow-status` is the read-only report.

  Return `/flow` for both `STARTED` and `IN_PROGRESS`, and keep `""` for `FINISHED`, which is
  terminal. Rewrite the doc comment to describe the pipeline that actually exists, and update the
  citation to `/flow-status`. Update `stats/internal/api/stats_test.go`'s expectation to match.

  **This is a behaviour change, not a lexical rename, which is why it is its own task and its own
  commit.** It is also the one `/myflow-*` occurrence in the corpus that is neither stored data, nor
  a live selector matching stored rows, nor a doc comment describing history — the three classes
  `design.md`'s `retired-command-constants-are-data` excludes. It is *forward-looking guidance*
  computed from current state, telling a dashboard user to run something that cannot be run. Task 9
  found it, correctly declined to sweep it, and the per-task review argued it into scope: leaving it
  would ship a dashboard that recommends retired commands under a change whose entire purpose is that
  the pipeline is called `flow`.

  It is genuinely pre-existing — byte-identical at the merge base, untouched by every commit on this
  branch — so it is a defect this change surfaces rather than one it introduces.

  Verify: `cd stats && go test ./internal/api/... -count=1` passes; `go vet ./... && gofmt -l .`
  clean; then confirm against the **real** rendered dashboard rather than the unit test alone —
  `make ui-test-up`, fetch the state board from the API on port 4174, confirm the `nextCommand`
  values it returns are `/flow` and not a retired command, then `make ui-test-down`.

**Build:** green
**Files:** `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`
**Tests:** no new test — `stats/internal/api/stats_test.go`'s existing `nextCommand` expectation is
retargeted, and it is the regression surface.
**Regression:** reverting this task restores `/myflow-do` and `/myflow-finish` as the dashboard's
recommended next command; the retargeted expectation in `stats_test.go` fails, naming both values.
**Baseline:** before=16 after=16 Go packages, all passing.
<!-- measured: cd stats && go test ./... -count=1 @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->
**Commit:** `fix(stats): stop the dashboard recommending retired commands`

- [ ] 21. Rename the managed-block delimiters, with a migration

  `setup.sh`'s `CLAUDE_MD_BEGIN` / `CLAUDE_MD_END` are `<!-- myflow:begin -->` and
  `<!-- myflow:end -->`, and its backup suffix is `.myflow.bak`. Task 15 renamed them, discovered the
  consequence, and reverted — correctly, because **these are not vocabulary. They are values compared
  byte-for-byte against files that already exist in the operator's home directory.** This machine's
  own `~/.claude/CLAUDE.md` carries `<!-- myflow:begin -->` today, and a `CLAUDE.md.myflow.bak` sits
  beside it.

  Renaming the markers alone makes the installer fail to find the block it wrote last time, so it
  appends a second one. Measured in a sandboxed HOME: install once with the current markers, then
  install again with renamed markers, and the file ends up carrying **both** — an orphaned
  `myflow:`-delimited block and a new `flow:`-delimited one, at twice the length. The orphan is never
  updated again by any later install, so the operator silently accumulates a stale copy of every rule.
  <!-- measured: two sandboxed setup.sh global runs into one throwaway HOME, the second with the delimiters renamed, counting anchored marker lines and file length @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->

  So rename the markers **and** make the installer self-healing: before writing its block, remove a
  legacy `<!-- myflow:begin -->` … `<!-- myflow:end -->` block if one is present, then write the
  `flow:`-delimited block as usual. Recognise the legacy `.myflow.bak` suffix the same way — keep
  treating an existing one as the pre-install backup rather than overwriting it, while writing
  `.flow.bak` going forward.

  **A migration, not a refusal, and that asymmetry is deliberate.** `.myflow/` → `.flow/` refuses
  loudly because it is a directory in a repository the operator must rename themselves, in projects
  this change cannot reach. The managed block is different: it lives in the operator's home, the
  installer already owns and rewrites it on every run, and it can therefore fix itself with no
  operator step at all. Refusing there would demand a manual edit of a generated file.

  Verify: in a throwaway HOME, install with the **pre-change** installer to plant a legacy block, then
  run the new installer and confirm the result carries exactly one block, `flow:`-delimited, with the
  orphan gone and the file back to a single block's length. Then run it a third time and confirm it
  stays idempotent. Do this with real installs into a real sandbox HOME — the failure this task exists
  to prevent was invisible to every guard and only appeared when the installer was run twice.

**Build:** green
**Files:** `setup.sh`, `scripts/test-setup.sh`
**Tests:** two new harness cases — a legacy-marker block is migrated to one flow-delimited block, and
a second run stays idempotent.
**Regression:** reverting this task restores the myflow-delimited markers; the two new harness cases
fail, and an operator upgrading across the rename accumulates a permanently stale duplicate block.
**Baseline:** before=343 after=345 assertions in the setup harness.
<!-- predicted: scripts/test-setup.sh after this task -->
**Commit:** `feat(setup): migrate the legacy managed-block delimiters to flow`

- [x] 22. Keep the CLI's diagnostics out of the findings JSON

  `scripts/check-unfinished-work.sh` captured `flow record findings` with `2>&1`, folding the CLI's
  stderr into the JSON it hands to `jq`. `flow` writes diagnostics to stderr and the payload to
  stdout, so any diagnostic puts a non-JSON line at the head of the payload and `jq` fails with a
  parse error — the guard then reports "cannot determine anything" for a store it read perfectly.

  **The one diagnostic that fires reliably is `flow: using FLOW_ADDR=...`, printed whenever the
  address is overridden — which is exactly what this repository instructs an operator to do to point
  a session at the ui-test stack.** So the guard broke only for the operator who followed the
  documented instructions, and only in a way that looks like a store outage.

  Capture the two streams separately: JSON from stdout, diagnostics to a temporary file reported in
  the failure message. Add a harness case pinning it — a stub writing a diagnostic to stderr and
  valid JSON to stdout, asserting exit 0 and that nothing reached `jq`.

  **This is pre-existing, not introduced by the rename** — the retired CLI printed
  `myflow: using MYFLOW_ADDR=...` on stderr in exactly the same shape, confirmed by running it. It is
  fixed here rather than deferred because it is what stands between this branch and a verifiable
  green suite: the Go suite's one guard-shelling test cannot pass against a daemon on any address but
  the default until this is repaired, and the default is occupied by a daemon the operator must not
  be asked to stop mid-change.

  Verify: the harness passes; reverting the capture to `2>&1` makes the new case fail (mutation-tested,
  it does); and with the fix in place the full Go suite passes in all sixteen packages against a
  branch-built daemon on a non-default port.

**Build:** green
**Files:** `scripts/check-unfinished-work.sh`, `scripts/test-check-unfinished-work.sh`
**Tests:** one new harness case — a stderr diagnostic alongside valid stdout JSON leaves the guard at
exit 0 and never reaches jq.
**Regression:** reverting this task restores the combined capture; the new case fails with the real
`jq: parse error`, and the guard again reports an unreachable store whenever the address is
overridden.
**Baseline:** before=16 after=16 Go packages, all passing against a branch-built daemon.
<!-- measured: cd stats && PATH=stats/bin:$PATH FLOW_ADDR=http://127.0.0.1:7773 go test ./... -count=1 @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->
**Commit:** `fix(scripts): keep flow's diagnostics out of check-unfinished-work's JSON`
