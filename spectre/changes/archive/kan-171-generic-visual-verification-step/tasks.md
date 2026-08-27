# Tasks — kan-171

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Three layers on one branch: the contract and its guard (tasks 1–3), the stage that reads it
(tasks 4–5), and the two projects that declare it (tasks 6–8). `design.md` is canonical for every
decision below.

**Baseline, measured before any edit:**

- The Go suite passes in every one of its seventeen packages.
  <!-- measured: cd stats && go test ./... -count=1 @ 1717d19 (before this change) -->
- The SPA suite passes: 9 files, 157 tests.
  <!-- measured: cd stats/web && npm test @ 1717d19 (before this change) -->
- `scripts/` carries 26 guards and 34 test harnesses.
  <!-- measured: ls scripts/ | grep -c '^check-'; ls scripts/ | grep -c '^test-' @ 1717d19 (before this change) -->

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` and reconciles
its `budgets()` row in the same commit.** That guard is in the project's `## lint` list, so a commit
leaving it red leaves lint red. **Reconcile means run the guard and answer what it reports** — raise
a row only when the guard actually fails on that file, and only to the size its own rule gives.
Raising a row the guard never complained about loosens the ratchet silently, and nothing catches it.

---

## Part 1 — the contract and its guard

- [x] 1. Guard and harness for the `## visual verification` section

  Write `scripts/test-check-visual-verification.sh` first, then
  `scripts/check-visual-verification.sh` against it — RED before GREEN.

  The guard takes one argument, the project root, and reads `<root>/.flow/project.md`. A project
  with no `## visual verification` section is **clean, not a failure** — the section is optional per
  `design.md`, and a guard that failed on its absence would make every project in the world declare
  one.

  Validate the two tables `design.md`'s **The contract** section specifies:

  ```bash unverified:confirm the exact heading text once task 2 has written the contract
  # settings table:  ui paths | screenshots | regression checkout | regression repo
  #                  push to default branch
  # commands table:  setup | verify | capture
  ```

  Findings, each reported as `file:line` with the row named:

  - `ui paths` or `screenshots` absent, or empty.
  - `verify` or `capture` absent, or empty.
  - `regression repo` present with no `regression checkout`, or the reverse.
  - `regression checkout` naming a path that is not an existing directory, or not a git checkout.
  - **`push to default branch` carrying any value but the literal `allowed`.** A near-miss
    (`yes`, `true`, `Allowed`) is a finding, never a silent grant.
  - **`push to default branch: allowed` where the `regression checkout`'s `origin` remote does not
    equal the declared `regression repo`.** Report **both** URLs. This is the security-relevant
    check and the reason the guard exists at all: `<project>/.flow/project.md` is tracked and
    editable in any pull request, and this setting ends in a `git push`. Without it a project could
    declare a push carve-out aimed at a repository it does not name.

  Exit `0` clean, `1` a finding, `2` it cannot answer — the project root missing, `.flow/project.md`
  unreadable, or `git remote get-url origin` failing for any reason other than "no such remote"
  (which is finding, not a `2`). **A `2` is never a silent skip**: a guard pointed at a moved file
  that reports `✓ clean` is the vacuous pass `scripts/check-vocabulary.sh`'s own header warns about.

  The harness builds every fixture under `TMPDIR`, including a real `git init` checkout with a
  planted `origin`, and **mutates the guard to prove each finding fails** — the requirement
  `scripts/test-check-reproduce-not-read.sh` and every other harness here already meets.

**Files:** `scripts/check-visual-verification.sh`, `scripts/test-check-visual-verification.sh`
**Tests:** `scripts/test-check-visual-verification.sh`
**Regression:** reverting this commit removes the only check that a declared push carve-out points
at the repository it names; a `regression repo` line could then disagree with the checkout's real
`origin` and the push would follow the checkout.
**Baseline:** before=34 after=35 harnesses in `scripts/`
<!-- predicted: ls scripts/ | grep -c '^test-' after task 1 -->
**Commit:** `feat(scripts): guard the visual verification section`
**Build:** green

- [x] 2. The `## visual verification` contract in `project-configuration.md`

  Add the section to `skills/flow-contracts/project-configuration.md`, which is canonical for
  `<project>/.flow/project.md`. Add its row to that file's own key table, worded like the
  `## workspace isolation` row: optional, **resolved and validated rather than read**, two tables and
  nothing else, prose beside them never resolved.

  State, in the section itself: both tables' columns; that `regression checkout` absent means the
  spec and baselines are committed to the change's own branch; that `push to default branch` absent
  means never push and only the literal `allowed` enables it; and that the permission is honoured
  only when `regression repo` matches the checkout's real `origin`, naming
  `check-visual-verification.sh` as what enforces it.

  Do not restate the guard's finding list — cite the guard. Do not restate the stage's procedure —
  task 4 writes it and it is canonical there.

  Run `scripts/check-contract-budget.sh` and reconcile this file's row in this commit.

**Files:** `skills/flow-contracts/project-configuration.md`
**Tests:** none added — this task adds prose to a contract file; the guards that verify prose in
this repository are `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
`scripts/check-contract-budget.sh` and `scripts/check-markdown-integrity.py`
**Regression:** reverting this commit leaves `check-visual-verification.sh` validating a section
no contract defines, so a project author has nothing to write the section against.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `docs(flow-contracts): define the visual verification section`
**Build:** green

- [x] 3. Install the guard everywhere a guard is installed

  Add `check-visual-verification.sh` to `skills/flow/scripts/` and to `.flow/project.md`'s
  `## lint` list. Run `scripts/check-guard-symlinks.sh` and `scripts/test-setup.sh`.

  **`setup.sh` needs no edit, measured rather than assumed.** Its `install_skills()` symlinks each
  top-level skill directory as one unit, so a new file under `skills/flow/scripts/` installs with
  no per-guard list to update; `scripts/test-setup.sh` likewise derives its expected guard set from
  the repo tree (`for _gs in skills/*/scripts`) and picks a new guard up on its own.
  <!-- measured: SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global; readlink -f "$SANDBOX/.claude/skills/flow/scripts/check-visual-verification.sh" @ branch spectre/kan-171-generic-visual-verification-step -->

  Add the guard to the union `skills/flow/SKILL.md`'s **Check guard presence** paragraph enumerates,
  so a missing script prints the guard-presence block rather than being discovered at its call site.

**Files:** `skills/flow/SKILL.md`, `.flow/project.md`,
`skills/flow/scripts/check-visual-verification.sh`
**Tests:** `scripts/test-setup.sh`, `scripts/check-guard-symlinks.sh`
**Regression:** reverting this commit leaves the guard present in `scripts/` but absent from
`skills/flow/scripts/`, where **Guard resolution** (`skills/flow-contracts/pipeline.md`) resolves
it — so every run would fall through to the hand-run fallback.
**Baseline:** before=0 after=0 tests added
<!-- predicted: test-setup.sh gains assertions, not a new file -->
**Commit:** `chore(setup): install the visual verification guard`
**Build:** green

## Part 2 — the stage

- [x] 4. The `flow.visual-verify` stage

  Add the stage to `skills/flow/verify-and-handoff.md`, between `## Verify` and
  `## Stage, excluding the planning paths`, with its `flow stage begin`/`end` marks. Write the ten
  steps `design.md`'s **The stage** section specifies. Add the `Visual:` line to the handoff block,
  naming each view looked at and its screenshot's **absolute** path, per **Handoff output**
  (`skills/flow-contracts/pipeline.md`)'s every-path-is-absolute rule.

  State the blocking rule in full: a failed `setup`, `verify` or `capture`, a stack that would not
  start, an unreadable PNG, **and a defect the agent sees in a screenshot even when every assertion
  passed**. That last is the whole point — `design.md` records the three KAN-16 defects that passed
  a five-pass panel and both suites.

  State the stop rule in full: stop the stack **only if step 4 started it**. A stack the operator
  already had up is left alone.

  Add `flow.visual-verify` to `skills/flow/SKILL.md`'s **Stage keys** table.

  Run `scripts/check-stage-mark-calls.sh` — it rejects a hardcoded `-harness` literal and a
  substituted session token, and the new marks must pass it. Run `scripts/check-contract-budget.sh`
  and reconcile both files' rows in this commit.

**Files:** `skills/flow/verify-and-handoff.md`, `skills/flow/SKILL.md`,
`scripts/check-contract-budget.sh`
**Tests:** `scripts/check-stage-mark-calls.sh`, `scripts/check-references.sh`,
`scripts/check-reproduce-not-read.sh`, `scripts/check-contract-budget.sh`
**Regression:** reverting this commit removes the only place the pipeline runs a browser, leaving
the contract from task 2 declared by projects and read by nothing.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `feat(flow): add the visual verification stage`
**Build:** green

- [x] 5. Document the stage key

  Add the `flow.visual-verify` row to **Level 1 — the stages of each command** (`README.md`) and the
  matching entry to `stats/internal/stages/names.go`, positioned between `flow.verify` and
  `flow.stage-diff` in both. `stats/internal/stages/names_test.go` parses the README table and
  compares it to the Go table, so the two must agree byte for byte on the key.

  Without this the CLI rejects every `flow stage begin -stage flow.visual-verify` call as an
  undocumented key — a caller mistake that reports rather than blocks, but one that would leave the
  stage unrecorded in the store for the life of the change.

**Files:** `README.md`, `stats/internal/stages/names.go`
**Tests:** `stats/internal/stages/names_test.go` — existing, extended by the table it parses
**Regression:** reverting this commit makes the CLI reject the stage's own marks, so no
`flow.visual-verify` run is ever recorded.
**Baseline:** before=17 after=17 Go packages passing
<!-- predicted: cd stats && go test ./... -count=1 after task 5 -->
**Commit:** `feat(stages): document the visual-verify stage key`
**Build:** green

## Part 3 — the projects that declare it

- [x] 6. A Playwright visual baseline for the stats SPA

  Add `@playwright/test` **1.62.1** to `stats/web`'s `devDependencies`.
  <!-- measured: npm view @playwright/test version @ 2026-08-27 -->

  Add `stats/web/playwright.config.ts` pinned to `http://127.0.0.1:4174` — the port
  `make ui-test-up` serves — with a fixed viewport and `deviceScaleFactor: 1`, and a small
  `maxDiffPixels` tolerance so font antialiasing does not flake the baseline while real drift still
  fails.

  Add `stats/web/tests/visual/baseline.spec.ts` covering the **dashboard** and **run-detail** views —
  the two KAN-16 broke — with `toHaveScreenshot()`, and commit the generated baseline PNGs.

  Add `test:visual` and `test:visual:update` scripts to `stats/web/package.json`.

  **The target is `make ui-test-up` / `make ui-test-down` on `127.0.0.1:4174`, never the dev
  workspace's daemon on 4173.** `CLAUDE.md` forbids this change stopping that service or its
  storage, and its data changes every run, so no baseline over it could be stable. The UI-test stack
  already drops and recreates `flow_uitest`, seeds a fixed fixture, and waits for the daemon to
  answer — which is exactly the deterministic seed a baseline needs.

**Files:** `stats/web/package.json`, `stats/web/package-lock.json`,
`stats/web/playwright.config.ts`, `stats/web/tests/visual/baseline.spec.ts`,
`stats/web/tests/visual/baseline.spec.ts-snapshots/**`
**Allowed-collateral:** `stats/web/node_modules/**`
**Tests:** `stats/web/tests/visual/baseline.spec.ts`
**Regression:** reverting this commit removes the only automated signal that the stats dashboard
changed appearance when nobody meant it to — the exact gap KAN-16 fell through.
**Baseline:** before=157 after=157 vitest tests; visual specs run under Playwright, not vitest
<!-- predicted: cd stats/web && npm test after task 6 -->
**Commit:** `test(web): add a visual baseline for the dashboard and run detail`
**Build:** green

- [x] 7. Declare `## visual verification` for `agents`

  Add the section to `.flow/project.md`: `ui paths` covering `stats/web/src/**`, `screenshots`
  pointing at the snapshots directory, `verify` and `capture` running the npm scripts task 6 added,
  and `setup` running `npm install && npx playwright install chromium`.

  **No `regression checkout` row** — `agents` commits its own baselines to its own change branch,
  per `design.md`'s `agents-has-no-regression-checkout`. **No `push to default branch` row** either,
  so nothing here can push anywhere.

  Run `scripts/check-visual-verification.sh .` and `scripts/check-contract-budget.sh`.

**Files:** `.flow/project.md`, `scripts/check-contract-budget.sh`,
`skills/flow-contracts/project-configuration.md`, `skills/flow/verify-and-handoff.md`
**Tests:** `scripts/check-visual-verification.sh`
**Regression:** reverting this commit leaves the stats SPA — the UI KAN-171 was filed about —
the one UI in this repository the new stage never runs on.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `chore(flow-config): declare visual verification for the stats SPA`
**Build:** green

  **This task's commit reaches two files task 2 and task 4 own, and that is recorded rather than
  silent.** Declaring the section against a real project is what exposed two contract defects: a
  fixed `screenshots` leaf path cannot serve both `verify` and a per-change `capture` spec, and
  `capture` exits non-zero on a first-run snapshot write, which would have blocked every run.
  Correcting the contract is what made the declaration expressible, so both corrections belong to
  this task's own commit.

- [x] 8. Declare `## visual verification` for Gymie

  **This task edits a different repository** — `/Users/tweety53/Projects/gymie` — and leaves the edit
  **uncommitted** in that repo's working tree for the operator to review. Nothing in this task is
  part of this change's own branch, and nothing here is pushed.

  Add the section to `/Users/tweety53/Projects/gymie/.flow/project.md`:

  ```markdown verified:written against the real tree and validated by scripts/check-visual-verification.sh /Users/tweety53/Projects/gymie
  | Setting | Value |
  |---------|-------|
  | `ui paths` | `gymie-frontend/**`, `gymie-admin-frontend/**`, `src/app/*/src/main/kotlin/**/infrastructure/presentation/web/**`, `src/app/common/src/main/kotlin/com/gymie/common/web/**` |
  | `regression checkout` | `/Users/tweety53/Projects/gymie-playwright` |
  | `regression repo` | `git@github.com:tweety53/gymie-playwright.git` |
  | `push to default branch` | `allowed` |
  | `screenshots` | `tests/visual-baseline.spec.ts-snapshots` |
  ```

  **`screenshots` is Playwright's own default snapshot path, not the repository root.** The value
  was guessed `.` when this plan was written; the real baseline PNGs sit under
  `tests/visual-baseline.spec.ts-snapshots/`, which is where `toHaveScreenshot()` puts them.

  **The `src/main/kotlin` segment already excludes `build/**`** — Gradle's compiled classes land
  under `build/classes/kotlin/...`, a path shape the glob cannot match — so no separate exclusion is
  needed. The `invite` module carries no `presentation/web` package; its HTTP surface goes through
  `admin`'s controllers, already covered.

  Commands: `setup` = `npm install && npx playwright install chromium`; `verify` =
  `npm run test:visual`; `capture` = `npx playwright test <spec>`.

  **Write the push permission with its reason beside it, in prose the resolver never reads.**
  `gymie-playwright` has no branch but `main` and no branch protection, so this is the operator's
  explicit, repository-scoped override of `no-direct-pushes-to-main` — recorded at the point it
  takes effect, scoped to that one repository, and to no other.

  The backend paths are in `ui paths` because one of KAN-16's three defects was a server-side
  allowlist mismatch that showed only as a broken page; frontend-only globs would have missed it.

  Gymie's frontend keeps port 3000 in every worktree — its `## workspace isolation` declares that
  outright, because Google's OAuth client registers that origin by port — so `gymie-playwright`'s
  fixed `baseURL` is correct from an apply worktree without change.

  Verify with `scripts/check-visual-verification.sh /Users/tweety53/Projects/gymie`, which is what
  proves the declared `regression repo` matches that checkout's real `origin`.

**Files:** `/Users/tweety53/Projects/gymie/.flow/project.md` — outside this repository, left
uncommitted
**Tests:** `scripts/check-visual-verification.sh /Users/tweety53/Projects/gymie`
**Regression:** reverting this task's edit leaves Gymie — the project the ticket names — without
the declaration, so the stage skips every Gymie change.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task; the guard is run against gymie's tree -->
**Commit:** none — this task commits nothing to this repository
**Build:** green

## Part 4 — make the stage testable, and test what is visible

Added after the first human gate. The stage's steps were prose no test executed; the operator
required everything visible tested before handoff, with e2e coverage added.

- [x] 9. Extract the trigger match into a guard

Write `scripts/test-check-visual-trigger.sh` first, then `scripts/check-visual-trigger.sh` — RED
before GREEN. The guard takes a project root and a list of changed paths on stdin, resolves
`ui paths` from `## visual verification`, and exits `0` when at least one path matches, `1` when
none does, `2` when it cannot answer. A project declaring no section is `2`, not `1` — "not
configured" and "configured and unmatched" are different answers and the stage acts differently on
each.

Then replace step 2's prose in `skills/flow/verify-and-handoff.md` with an invocation of it.

**The glob semantics are the whole risk** — `**` spanning directories, a leading `./`, a path
outside every app root, an absolute path, a glob with a space. Pin each in the harness, and
mutation-prove every finding.

**Files:** `scripts/check-visual-trigger.sh`, `scripts/test-check-visual-trigger.sh`,
`skills/flow/verify-and-handoff.md`, `skills/flow/scripts/check-visual-trigger.sh`, `.flow/project.md`
**Tests:** `scripts/test-check-visual-trigger.sh`
**Regression:** reverting this commit returns the trigger decision to prose, where two runs can read
the same globs differently and nothing catches it.
**Baseline:** before=35 after=36 harnesses in `scripts/`
<!-- predicted: ls scripts/ | grep -c '^test-' after task 9 -->
**Commit:** `feat(scripts): decide the visual trigger in a guard, not in prose`
**Build:** green

- [x] 10. Extract the screenshot resolution into a guard

Write `scripts/test-resolve-visual-screenshots.sh` first, then
`scripts/resolve-visual-screenshots.sh` — RED before GREEN. Given a project root and a capture
spec's basename, it prints one **absolute** PNG path per line, searching recursively beneath
`screenshots`, and exits `1` on zero matches — the case the stage must block on, per `design.md`.

Replace steps 7 and 8's resolution prose with it, keeping the requirement that the agent **reads**
each PNG, which no script can do.

**Pin the cases that bit already:** snapshots nested under `<spec>-snapshots/`, a `screenshots` root
that is the repository root, and a root containing `node_modules` whose own PNGs must not match a
spec basename.

**Files:** `scripts/resolve-visual-screenshots.sh`, `scripts/test-resolve-visual-screenshots.sh`,
`skills/flow/verify-and-handoff.md`, `skills/flow/scripts/resolve-visual-screenshots.sh`,
`.flow/project.md`
**Tests:** `scripts/test-resolve-visual-screenshots.sh`
**Regression:** reverting this commit returns PNG resolution to prose, where the zero-match case —
a capture that silently wrote nothing — can read as a pass.
**Baseline:** before=36 after=37 harnesses in `scripts/`
<!-- predicted: ls scripts/ | grep -c '^test-' after task 10 -->
**Commit:** `feat(scripts): resolve captured screenshots in a guard, not in prose`
**Build:** green

- [x] 11. Extract the push gate into a guard

Write `scripts/test-check-visual-push-gate.sh` first, then `scripts/check-visual-push-gate.sh` —
RED before GREEN. It exits `0` **only** when `push to default branch: allowed` is declared **and**
the `regression checkout`'s real `origin` equals the declared `regression repo`; `1` when a push is
not permitted; `2` when it cannot answer.

**This is the one untested path that performs an irreversible outward-facing action.** Step 9
currently asks an agent to check two conditions by reading prose; after this task it runs a guard
and pushes only on exit `0`.

Reuse `check-visual-verification.sh`'s own ambient-git neutralisation — `GIT_DIR` and its siblings
defeat `-C`, and this guard is the one where that matters most. Pin it in the harness with a
hostile `GIT_DIR`, as case 22 there already does.

**Files:** `scripts/check-visual-push-gate.sh`, `scripts/test-check-visual-push-gate.sh`,
`skills/flow/verify-and-handoff.md`, `skills/flow/scripts/check-visual-push-gate.sh`, `.flow/project.md`
**Tests:** `scripts/test-check-visual-push-gate.sh`
**Regression:** reverting this commit returns the push decision to prose — the guard authorising an
irreversible push would again be evaluated by reading rather than by running.
**Baseline:** before=37 after=38 harnesses in `scripts/`
<!-- predicted: ls scripts/ | grep -c '^test-' after task 11 -->
**Commit:** `feat(scripts): gate the visual push in a guard, not in prose`
**Build:** green

- [x] 12. Cover every control the SPA exposes, plus combinations

Expand `stats/web/tests/visual/` from three view-level specs to the whole visible surface, against
the `make ui-test-up` fixture stack on `127.0.0.1:4174`.

**The surface, enumerated from the source rather than guessed:**

```text verified:read from stats/web/src/App.tsx and stats/web/src/components/ at 75f302b
views:    state-board (default), stage-leaderboard, trend, cache-efficiency,
          run/<project>/<change>
controls: PeriodPicker 2 buttons + 3 inputs, SearchBox 1 input, ProjectFilter 1 input,
          FilterBar 1 select, DataTable 5 sort buttons, ChangeVariable 1 select,
          ModelVariable 1 select
```

Every view reachable, every control exercised, and **combinations** — at minimum: a filter plus a
search; a period change plus a sort; and the built-in interaction where `ModelVariable` is disabled
on `state-board` and enabled elsewhere, which is a real conditional in `App.tsx` and exactly the
kind of composition defect KAN-16 shipped.

**Assert on more than pixels.** A screenshot proves a page rendered, not that it is right — assert
the data actually loaded (no error banner, expected row counts) alongside `toHaveScreenshot()`, the
way the existing run-detail spec catches the 400.

**Files:** `stats/web/tests/visual/**`, `stats/web/package.json`
**Tests:** the specs this task adds, under `stats/web/tests/visual/`
**Regression:** reverting this commit leaves every control in the SPA unexercised, so a change that
breaks a button, a filter or a sort ships with the baseline still green.
**Baseline:** before=3 after=3+ visual specs; vitest unchanged at 9 files, 157 tests
<!-- predicted: cd stats/web && npm test, and npm run test:visual, after task 12 -->
**Commit:** `test(web): cover every SPA control and key combinations`
**Build:** green

- [x] 13. Reload the project's applications after a fix, before handing off

A fix run hands the operator a diff **and** run instructions, and the state means "a staged diff is
waiting for the human to review, alongside the run instructions the handoff printed." If the
applications those instructions name are still serving pre-fix code, the operator reviews one thing
and runs another.

**This is measured, not hypothetical.** `stats/internal/web/embed.go` bakes the built SPA into the
daemon binary with `//go:embed all:dist`, so a change to SPA source is invisible to an
already-running daemon until it is rebuilt; `stats/Makefile`'s `ui-test-up` depends on `web-build`
for exactly that reason. Twice in this change's own development an agent's edit ran against a stale
binary and looked like a pass — caught only by watching the served bundle hash change.
<!-- measured: stats/internal/web/embed.go's //go:embed all:dist, and stats/Makefile's `ui-test-up: web-build` prerequisite @ branch spectre/kan-171-generic-visual-verification-step -->

State the rule in `skills/flow/verify-and-handoff.md`, in the phase that runs before the handoff:

- **After a fix run, and before the handoff prints its run instructions, reload the applications
  those instructions name** — rebuild and restart them from the project's `## run` commands, so the
  operator's review and the operator's run see the same code.
- **Reload what the handoff names, and nothing else.** A project's applications only.
- **Never the flow dev stack.** `flowd` on `127.0.0.1:4173`, its `flow-postgres` container and the
  `flow` database are never stopped, restarted or dropped by any run — `<project>/CLAUDE.md` states
  that prohibition and this rule does not weaken it. A project whose own `## run` names that service
  is the one case where the rule yields to the prohibition rather than the other way round.
- **A reload that fails blocks the handoff**, and says which application and what the command
  printed. Handing over run instructions that cannot be followed is the failure this rule exists to
  prevent.
- **Where a project declares no runnable application**, there is nothing to reload and the rule is
  satisfied by saying so — not by silently skipping it.

**The rule is generic: it binds every project that uses `/flow`, not this repository.** Reload
resolves from the two keys projects already declare — the project's `## stop`, where it declares a
command, followed by its `## run` — restricted to the applications the handoff names. **No new
project-configuration key is added**, and none is needed.

**A `## stop` that deliberately declares no command means there is nothing to stop**, not that the
rule is skipped: the reload is then whatever `## run` does to bring the application up fresh.

Two worked examples, both real, and they resolve differently — which is why the rule is stated
against the keys rather than against a command:

```bash verified:read from /Users/tweety53/Projects/gymie/.flow/project.md and this repository's own .flow/project.md
# gymie — `## stop` declares a command, so reload is stop-then-run:
./gradlew devStop
docker compose up -d && ./gradlew devStart -PfrontendRoot=<abs> -PadminFrontendRoot=<abs>

# this repository — `## stop` declares no command by design; `make ui-test-up` rebuilds the
# embedded SPA and restarts the UI-test daemon in one target, and never touches 4173:
cd stats && make ui-test-up
```

Do not restate the stage's own start/stop rule — the visual stage stops only the stack it started,
and that is stated once, where it belongs. Cite it.

**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** none added — this task states a rule in a contract file; the prose guards in this
repository's `## lint` list are what verify it
**Regression:** reverting this commit lets a fix run hand the operator run instructions pointing at
applications still serving pre-fix code — the stale-binary failure this change hit twice.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `feat(flow): reload the project's applications after a fix, before handing off`
**Build:** green

## Part 5 — the review panel's findings

All four panel slots blocked. Every finding below was demonstrated, not argued.

- [x] 14. Drop the automatic push

`design.md`'s `no-automatic-push` is canonical. The stage commits to the regression checkout and
**stops**; the handoff prints the push command for the operator.

Remove `push to default branch` from `skills/flow-contracts/project-configuration.md`, from
`skills/flow/verify-and-handoff.md`'s step 9, and from Gymie's declaration
(`/Users/tweety53/Projects/gymie/.flow/project.md` — **uncommitted, that repo is not ours to
commit**). Delete `scripts/check-visual-push-gate.sh`, its harness, its symlink and its `## test`
entry.

**`regression repo` stays**, downgraded to an identity assertion: `check-visual-verification.sh`
still reports a mismatch against the checkout's real `origin`, but nothing is authorised by it.
Reword the guard's header so it no longer claims to gate anything — its current claim is what the
panel proved false.

**Task 11 was not wrong; the design changed under it.** Say so where the deletion is recorded,
rather than implying the guard was defective.

**Files:** `skills/flow-contracts/project-configuration.md`, `skills/flow/verify-and-handoff.md`,
`scripts/check-visual-verification.sh`, `scripts/test-check-visual-verification.sh`,
`scripts/check-visual-push-gate.sh` *(deleted)*, `scripts/test-check-visual-push-gate.sh`
*(deleted)*, `skills/flow/scripts/check-visual-push-gate.sh` *(deleted)*, `.flow/project.md`
**Tests:** `scripts/test-check-visual-verification.sh` — cases for the removed setting
**Regression:** reverting this commit restores a push authorised by two pull-request-editable fields
that need only agree with each other.
**Baseline:** before=38 after=37 harnesses in `scripts/`
<!-- predicted: ls scripts/ | grep -c '^test-' after task 14 -->
**Commit:** `feat(flow): drop the automatic push to a regression checkout`
**Build:** green

- [x] 15. Source the shared guard helpers instead of copying them

`split_cells`, `trimcell`, `foldcell`, `sanitize_display` and `git_clean` are byte-identical
across the new guards — `sanitize_display` is 18 lines and `git_clean` 7, and the parser trio is
larger than either.
<!-- measured: awk-extracted each function body from scripts/check-visual-verification.sh and piped it through wc -l @ branch spectre/kan-171-generic-visual-verification-step --> The repository's WET convention
(`scripts/check-workspace-isolation.sh:93-113`) carves out exactly this case: a guard reachable
**only through the symlink farm** may source a lib, because the lib "travels WITH the guard,
unconditionally — never a way to be present-but-unrunnable."

All the new guards meet that condition: they are symlinks in `skills/flow/scripts/`, and
`skills/flow/scripts/lib -> ../../../scripts/lib` already exists. Five guards already source from
it.

```text verified:ls -l skills/flow/scripts/ and shasum of the extracted functions, run in this worktree
sanitize_display  identical across all four new guards (67b22701)
git_clean         identical across the two that carry it (0ce381be)
```

Extract into `scripts/lib/` and source, following `resolve-file.sh`'s existing precedent. Update
each guard's header so it cites the farm exception rather than the single-file rule it no longer
follows.

**The copies have not diverged today** — the hazard is that the next fix to a shared helper must
reach every copy, and nothing checks that it did. This change's own history has two such fixes
already.

**Files:** `scripts/lib/**`, `scripts/check-visual-verification.sh`,
`scripts/check-visual-trigger.sh`, `scripts/resolve-visual-screenshots.sh`
**Tests:** the existing harnesses, which must pass unchanged
**Regression:** reverting this commit restores hand-copied helpers whose next divergence nothing
detects.
**Baseline:** before=37 after=37 harnesses; no new test file
<!-- predicted: the existing harnesses cover the extracted helpers unchanged -->
**Commit:** `refactor(scripts): source the shared visual-guard helpers from lib`
**Build:** green

- [x] 16. Three guard defects the panel demonstrated

**a. Substring collision** — `resolve-visual-screenshots.sh`'s PNG filter is a bare substring test,
so resolving `baseline.spec.ts` also returns `visual-baseline.spec.ts`'s PNGs. Not triggerable by
this repo's five spec names today; it is the exact path-handling class the guard's own header claims
to have closed. Anchor the match at a boundary.

**b. BOM misparse** — a `.flow/project.md` beginning with a UTF-8 BOM, with `## visual
verification` as its first line, reads as declaring no section at all. It fails **safe**, but it is
a silent misparse of a heading that is present.

**c. Validator gap** — `push to default branch: allowed` with neither `regression checkout` nor
`regression repo` passes the validator clean. **Task 14 removes that setting**, so confirm this is
moot and say so; if any equivalent shape survives, close it.

Pin each in the harness with a case that fails against the current code.

**Files:** `scripts/resolve-visual-screenshots.sh`, `scripts/test-resolve-visual-screenshots.sh`,
`scripts/check-visual-verification.sh`, `scripts/test-check-visual-verification.sh`
**Tests:** the two harnesses above
**Regression:** reverting this commit returns a resolver that hands one spec another spec's
screenshots, and a parser that silently ignores a declared section behind a BOM.
**Baseline:** before=37 after=37 harnesses; cases added to existing files
<!-- predicted: no new test file; cases added to two existing harnesses -->
**Commit:** `fix(scripts): anchor the screenshot match and parse a BOM-prefixed heading`
**Build:** green

- [x] 17. Say what the reload rule does for this repository

Task 13's rule reloads "the applications the run instructions name". This repository's `## apps`
names exactly one URL-bearing application — the **protected** daemon on `127.0.0.1:4173` — and the
rule's worked example quietly substitutes the 4174 UI-test stack, which `## apps` never names. So
for the flagship project the rule's own promise is undelivered and its required behaviour is
unspecified.

State what an agent does when the only application a project names is one it must not touch:
**reload nothing, and say so in the handoff, naming the application and why.** Silence is what makes
this a gap; an explicit "not reloaded, and here is why" is the answer.

Then reconcile the worked example with `## apps`, either by naming the UI-test stack there or by
being explicit that the example reloads a stack `## apps` does not list and why that is correct.

**Files:** `skills/flow/verify-and-handoff.md`, `.flow/project.md`
**Tests:** none added — a rule in a contract file; the prose guards verify it
**Regression:** reverting this commit leaves the rule silent on the one project it ships with,
where the only named application is the one it must never restart.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `fix(flow): say what the reload rule does when the only app is protected`
**Build:** green

- [x] 18. List the visual-verification harness in `## test`

`scripts/test-check-visual-verification.sh` exists and passes, but `.flow/project.md`'s `## test`
list names only its two siblings — so the harness for the **most security-relevant guard** is run
by nobody. It has passed throughout this change only because agents invoked it by hand.

Found by the panel fix agent while doing something else, and left rather than fixed, correctly:
it predates that dispatch and was outside its scope.

```text verified:grep -n for both harness names in .flow/project.md, compared against ls scripts/, run in this worktree
in `## test`:  test-check-visual-trigger.sh, test-resolve-visual-screenshots.sh
on disk:       those two, plus test-check-visual-verification.sh
```

Add it, and check the same way for any other harness in `scripts/` absent from the list — a
one-off omission and a systematic one need different answers, and only looking tells you which
this is.

**Files:** `.flow/project.md`
**Tests:** `scripts/test-check-visual-verification.sh` — now actually run by the project's own list
**Regression:** reverting this commit returns the guard's harness to being run by nobody, so a
regression in the guard that gates the visual stage would ship green.
**Baseline:** before=2 after=3 visual harnesses listed in `## test`
<!-- measured: grep -c 'test-check-visual\|test-resolve-visual' .flow/project.md, before and after @ branch spectre/kan-171-generic-visual-verification-step -->
**Commit:** `fix(flow-config): run the visual-verification guard's harness`
**Build:** green

- [x] 19. Share the BOM handling, do not copy it a second time

Task 16b added BOM stripping to `check-visual-verification.sh` alone. `check-visual-trigger.sh` and
`resolve-visual-screenshots.sh` share the identical heading-parse convention and did not get it.

**The consequence is a silent skip, which is the failure this whole change exists to prevent.** The
trigger guard decides whether `flow.visual-verify` runs; on a BOM-prefixed `.flow/project.md` it
reports "declares no section" and exits 2, which step 2 treats as *not configured* — so a project
that declares the section correctly gets no visual verification and no warning.

```text verified:run in this worktree against a BOM-prefixed fixture built with printf '\xEF\xBB\xBF'
validator  exit 0   (handles the BOM)
trigger    exit 2   (reads as NOT CONFIGURED -> silent skip)
resolver   exit 2
```

**Fix it by sharing, not by copying it twice more.** Task 15 extracted the parser helpers into
`scripts/lib/`; the BOM handling arrived afterwards and stayed in one guard. Put it where the other
shared helpers already live so a third guard cannot miss the next fix.

**This is the maintenance hazard the panel predicted, and it materialised inside its own change** —
which is the argument for the shared lib, not against it. Say so where the extraction is recorded.

Pin it with a harness case **per guard**: a BOM-prefixed declaration must be seen by all three.
Verify RED→GREEN for the two that are currently wrong.

**Files:** `scripts/lib/**`, `scripts/check-visual-verification.sh`,
`scripts/check-visual-trigger.sh`, `scripts/resolve-visual-screenshots.sh`,
`scripts/test-check-visual-trigger.sh`, `scripts/test-resolve-visual-screenshots.sh`
**Tests:** `scripts/test-check-visual-trigger.sh`, `scripts/test-resolve-visual-screenshots.sh`
**Regression:** reverting this commit returns two of the three guards to reading a BOM-prefixed
declaration as absent — the trigger guard among them, so the stage skips silently.
**Baseline:** before=37 after=37 harnesses; cases added to two existing files
<!-- predicted: no new test file; cases added to two existing harnesses -->
**Commit:** `fix(scripts): share the BOM handling across the visual guards`
**Build:** green

