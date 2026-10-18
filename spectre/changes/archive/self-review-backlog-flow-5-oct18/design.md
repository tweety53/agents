## Context

- Five Jira tickets in status "Backlog Flow" planned as self-review findings per item 1 of **3. Fix
  every finding that is not `big`** (`skills/flow-self-review/SKILL.md`), on
  `self-review-backlog-flow-5-oct18` based at agents `52438f0d`. Selection: the backlog's two High
  tickets, then the oldest `flow-fix` tickets, each reproduced before it was taken; a ticket whose
  defect no longer holds was reported and the next candidate taken. The brainstorm design is
  `.superpowers/sdd/2026-10-18-self-review-backlog-flow-5-oct18-design.md` in this worktree. Three
  tasks reach outside this repository, each committed on `spectre/self-review-backlog-flow-5-oct18`
  in the change's worktree of that repository, added by `kickoff-worktree.sh --from`: gymie at
  `/Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-5-oct18` off `origin/develop`
  (`eb06468c`), gymie-playwright at
  `/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-backlog-flow-5-oct18` off
  `origin/main` (`eb0dbb6`). Both exist, clean.
  - KAN-913 (Task 1), blast radius 2: `stats/cmd/flowd/main.go` runs
    `reconcile.New(..., fallback.StateRoot(), ...)` unconditionally at startup and on reconnect.
    `fallback.StateRoot()` is `FLOW_STATE_DIR` or the shared default, and `.flow/project.md`'s
    `## workspace isolation` deliberately isolates no directory, so a workspace daemon on
    `flow_<id>` replays the shared journal. The radius is `main.go` and `wiring_test.go`; the
    helper's name is new and nothing cites it.
  - KAN-773 (Task 2), blast radius 3: `setup.sh` only echoes "Install from a stable main checkout"
    after linking. A plain refusal would break every installer test and sandbox script, which all
    run `setup.sh global` from a linked worktree into a sandboxed `$HOME`
    (`stats/internal/setuptest/*`, `installedcitations.go`, `scripts/installer-sandbox-diff.sh`,
    `scripts/test-setup-agents.sh`) — so the refusal applies only into the operator's own home. The
    radius is `setup.sh`, `harnesses_test.go` and `helpers_test.go`.
  - KAN-818 (Task 3), blast radius 1: gymie-playwright carries no `.flow/project.md`, so
    `kickoff-worktree.sh` prints `no ## worktree setup declared — skipped` for it (seen at this
    change's own kickoff) and the worktree has no `node_modules`. With the section declared,
    `project-get.sh <worktree> 'worktree setup'` exits 0 and `npm ci` makes
    `require.resolve('@playwright/test')` resolve inside the worktree.
  - KAN-780 (Task 4), blast radius 1: gymie's `.flow/project.md` `## lint` lists `ktlintFormat`
    then `verifyChange` with no word on why the format is separate.
  - KAN-799 (Task 5), blast radius 1: reproduced in the gymie worktree —
    `DEV_INVITE_EMAIL=dev@example.com ./gradlew :app:test --tests
    'com.gymie.dev.LocalDevSeederIntegrationTest'` fails `seedIsIdempotent`, `expected: 3L but was:
    21L`; `application-local.yml` maps `app.dev.invite-email` to `DEV_INVITE_EMAIL`, which the
    friend, invite and group-session seeders read.
  - Not planned, defect no longer holds:
    - KAN-731: the landing worktree it names is retired (no `_landing-` path in `scripts/`), and
      the main checkout is write-protected by hook since 2026-09-17.
    - KAN-626: `check-plan-shape.py` refuses a substring-unsafe `@Test` counter (KAN-936), and the
      Baseline convention mandates a word-boundary-safe counter.
    - KAN-699: `run-reproducer`'s bound reads `RUN_REPRODUCER_BOUND_SECONDS`, which
      `skills/flow/review-panel-fix-round.md` sets for a build or test reproducer.
    - KAN-717: fixed in gymie by KAN-796 (`devWorkspaceSurvivors` scrubs the daemon's `GIT_DIR`).
    - KAN-659: `check-visual-trigger.sh --repo <name>` matches another worktree's paths prefixed.
    - KAN-769: a `/flow-fast` default panel now runs the always-on extra dispatch, which records
      dispatch rows, so the loss note no longer fires on a mark-through.
    - KAN-779: `rules/kotlin-backend-development-standard.mdc` renders the lint commands from the
      project (`dd90af08`); gymie's managed blocks still carry the old text only because gymie was
      not re-stamped — `setup.sh claude-code /Users/tweety53/Projects/gymie`, an operator step.
- Spiked at plan time, every edit reverted: each task's diff applies in its worktree; Tasks 1 and 2's
  new tests pass with the fix and fail without it; `go vet` and `gofmt -l` are clean; Task 5's
  reproducer passes with the pin and `ktlintCheck` is clean.

## Decisions

### Replay the journal only into the flow database

**ID:** journal-replay-flow-db-only
**Status:** active
**Chosen:** `flowd` runs the reconciler, at startup and on reconnect, only when its DSN parses as a
URL whose path is `/flow`; otherwise it logs one skip line.
**Considered:** pointing a workspace daemon's journal at an isolated state root — the isolation
table's `Resource` vocabulary is closed and admits no directory (`.flow/project.md`); refusing to
replay an entry whose change the target database lacks — a journaled first write of a change has
no row anywhere yet, so the main daemon would refuse it too; a key-value DSN naming `dbname=flow`
— no configuration uses that form, and it reads as not the flow database, so it skips replay
rather than consuming the journal.

### Refuse a global install into the operator's home from a linked worktree

**ID:** setup-global-worktree-refusal
**Status:** active
**Chosen:** `setup.sh global` exits 1 naming `<main checkout>/setup.sh global` when `$HOME` is the
password database's home (`SETUP_REAL_HOME` in a test) and the checkout's git dir differs from its
common dir; outside git, or into any other `$HOME`, it installs as before.
**Considered:** refusing on the worktree alone — every installer test and sandbox script runs from
a worktree; an opt-out variable every sandbox caller sets — five callers changed to keep doing what
they do today.

### Install gymie-playwright's dependencies at worktree setup

**ID:** playwright-worktree-npm-ci
**Status:** active
**Chosen:** gymie-playwright gains a `.flow/project.md` declaring only `## worktree setup` with
`npm ci`.
**Considered:** a preflight asserting `@playwright/test` resolves inside the worktree — it detects
what setup can prevent, and a second check where none is needed once setup installs.

### State the read-only bundle rule in gymie's lint section

**ID:** gymie-lint-read-only-bundle
**Status:** active
**Chosen:** one paragraph under gymie's `## lint`, where bundle authors read the verify commands.
**Considered:** the agents repo's planning guidance — the constraint is Gradle's and gymie's, and no
other project bundles Gradle tasks.

### Pin the invite email in the seeder test

**ID:** seeder-test-pins-invite-email
**Status:** active
**Chosen:** `LocalDevSeederIntegrationTest`'s `@DynamicPropertySource` sets `app.dev.invite-email`
to empty, keeping every count assertion exact.
**Considered:** asserting the count the environment implies — the test would then pass two
different worlds and pin neither; pinning in `AppIntegrationTest` — the friend and group-session
seeder tests set the property themselves.

## Open questions

## Live check

Against a workspace `flowd` started from this branch on its own `flow_<id>` database with
`FLOW_STATE_DIR` at a scratch copy holding one replayable journal entry: before, its log shows
`flowd journal replay … applied=1` and the entry is gone from the copy; after, it logs `flowd
journal replay skipped` and the entry stays. The dev workspace's protected daemon is never
restarted for this; once the landing restarts it, its own log still shows `flowd journal replay`
at startup — a missing replay line there is the change failing. Tasks 2–5 hold no runtime state;
their checks are the tasks' own runs.
