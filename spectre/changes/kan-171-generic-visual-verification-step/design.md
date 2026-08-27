# Generic visual verification, with Gymie bound to Playwright — design

**Jira:** [KAN-171](https://tweety53.atlassian.net/browse/KAN-171)
**Change:** `kan-171-generic-visual-verification-step`
**Date:** 2026-08-27

## Why

KAN-16 shipped three defects that were invisible in a diff and obvious the moment the page was
opened: the run-detail dashboard returned 400 and loaded no stage runs; the dashboard rendered
white because the dark palette sat behind `prefers-color-scheme`; every stat panel printed its
label twice. All three survived a five-pass review panel, 296 Go tests and 110 SPA tests, and took
roughly two minutes to find in a browser.

The instruction to look already exists — `rules/design-mockups-are-specs.mdc` says "run the app,
screenshot the page, compare it side by side" — and it was not followed. A rule to remember failed;
this change makes it a stage with an artifact.

## What this is not

A screenshot proves the page renders, not that it is right. This replaces no reviewer. Its value is
the defect class nothing else can see.

## The contract — `## visual verification`

A new optional section in `<project>/.flow/project.md`, canonical in
`skills/flow-contracts/project-configuration.md`.

Like `## workspace isolation`, its body is **resolved and validated rather than read**: two tables
and nothing else, because the section carries commands and filesystem paths.
Prose beside the tables is for the reader and is never resolved.

### The settings table

| Setting | Required | Meaning |
|---------|----------|---------|
| `ui paths` | yes | Comma-separated globs, relative to each app root in `## apps`. A run whose diff matches none of them skips this stage. |
| `screenshots` | yes | Directory the `capture` command writes PNGs into, relative to the regression checkout when one is declared, otherwise to the project root. |
| `regression checkout` | no | Absolute path to the repository the per-change spec and its PNGs are committed to. Absent means they are committed to the change's own branch. |
| `regression repo` | only with `regression checkout` | The remote URL that checkout must have as `origin`. |

### The commands table

| Command | Required | Meaning |
|---------|----------|---------|
| `setup` | no | Run once before `verify` when the toolchain is missing. |
| `verify` | yes | Runs the checked-in baseline suite. A non-zero exit blocks the handoff. |
| `capture` | yes | Runs the per-change spec. `<spec>` in the command line is substituted with the spec's path. |

### The literal on-disk shape

The two tables above document each row's **meaning**; what a project actually writes is two
two-column tables, headed exactly:

```markdown verified:parsed by scripts/check-visual-verification.sh at 92d050a, whose harness pins both headers
| Setting | Value |
|---------|-------|

| Command | Runs |
|---------|------|
```

`| Command | Runs |` is the convention `## workspace isolation` already establishes for a table of
project-declared commands; a second shape for the same job was rejected as gratuitous divergence.
Both vocabularies are **closed** — a `Setting` or `Command` name outside the rows above is reported
and its row dropped, never silently honoured.

## The stage — `flow.visual-verify`

Inserted into `skills/flow/verify-and-handoff.md` between `flow.verify` and `flow.stage-diff`. The
phase order becomes **verify → visual-verify → stage-diff → run-instructions → write-in-progress**.

1. Resolve the section. Absent → print `Visual: not configured` and continue.
2. Run `git diff --name-only <merge-base>..HEAD` in every worktree in the resolved set and match
   against `ui paths`. No match → print `Visual: no UI paths touched` and continue.
3. Run `setup`, if declared. A non-zero exit blocks, printing the command verbatim.
4. Probe the URL this worktree resolved. If nothing answers, start the stack from `## run` and
   record that this stage started it.
5. Run `verify`. A non-zero exit blocks.
6. Author the per-change spec and run `capture`. PNGs land in `screenshots`.
7. Read every captured PNG and state, per view, what was seen.
8. Write `<changeRoot>/visual-verification.md`.
9. Commit to the regression checkout, when one is declared. **Never push** — see
   `no-automatic-push`. The handoff prints the push command for the operator.
10. Stop the stack only if step 4 started it.

### Blocking

The stage blocks the `IN_PROGRESS` handoff on: a failed `setup`, a failed `verify`, a failed
`capture`, a stack that could not be started, and **a defect the agent sees in a captured
screenshot** — even when every assertion passed. That last one is the defect class KAN-16
demonstrates: 406 passing tests and a five-pass panel over three visibly broken views.

An unreadable PNG is reported and blocks; it is indistinguishable from a view that never rendered.

## The guard — `check-visual-verification.sh`

Validates the two tables' shape, and one thing beyond shape:

**`regression repo` records which repository the `regression checkout` is expected to be**, and a
mismatch against its real `origin` is reported. It is an **identity assertion, not an
authorisation** — nothing is granted by it, per `no-automatic-push`. Both fields live in a
pull-request-editable file, so a matching pair proves only that whoever wrote them was consistent;
that is why nothing may be authorised from here.

The guard gets a mutation test, as every guard in this repository does.

## Gymie

`gymie/.flow/project.md` gains the section, as an uncommitted working-tree edit in that repo:

- `ui paths` — `gymie-frontend/**`, `gymie-admin-frontend/**`, and gymie's own controller, route and
  DTO surface. The backend paths are there because one of KAN-16's three defects was a server-side
  allowlist mismatch that showed only as a broken page; frontend-only globs would have missed it.
- `regression checkout` — `/Users/tweety53/Projects/gymie-playwright`
- `regression repo` — `git@github.com:tweety53/gymie-playwright.git`
- `verify` — `npm run test:visual`; `capture` — `npx playwright test <spec>`;
  `setup` — `npm install && npx playwright install chromium`

Gymie's frontend keeps port 3000 in every worktree — `## workspace isolation` declares that
outright, because Google's OAuth client registers that origin by port — so `gymie-playwright`'s
fixed `baseURL: http://localhost:3000` is correct from an apply worktree without change.

## `agents`

`stats/web` gains `@playwright/test` 1.62.1 (current stable) as a devDependency, a visual spec and
baseline PNGs under `stats/web/tests/visual/`, covering the **dashboard** and **run-detail** views —
the two KAN-16 broke.

The target is the existing `make ui-test-up` stack on `127.0.0.1:4174`. It already builds the SPA,
drops and recreates `flow_uitest`, seeds a fixed fixture, and waits for the daemon to answer;
`make ui-test-down` tears it down. A fixed fixture is what makes a baseline stable, and reusing the
target means step 4 above is two existing commands rather than new machinery.

No `regression checkout` row: `agents` commits its own baselines to its own change branch.

## Decisions

### The push carve-out is declared per repository, never assumed

**ID:** `push-carve-out-is-declared`
**Status:** superseded by `no-automatic-push`
**Chosen:** `push to default branch: allowed`, valid only when `regression repo` matches the
checkout's real `origin` — the operator's explicit, repository-scoped override of
`no-direct-pushes-to-main` for `gymie-playwright`, which has no branch but `main` and no branch
protection.
**Considered:** push a branch and open a PR — obeys the global rule unchanged, rejected because it
costs a PR merge on every UI change; push a branch without a PR — same cost deferred, rejected for
the same reason; commit locally and let the operator push — rejected, it is the state the repository
is already in, with twenty uncommitted per-ticket specs and PNGs sitting in its working tree.

### The trigger is declared globs, not the agent's judgment

**ID:** `trigger-is-declared-globs`
**Status:** active
**Chosen:** the project declares `ui paths` and the run matches its own diff against them.
**Considered:** run on every change in a project that declares the section — rejected, it costs a
browser run on backend-only work; let the agent judge from the diff — rejected, that is exactly the
remembered-rule judgment KAN-171 records failing once already.

### Gymie's globs include the backend API surface

**ID:** `gymie-globs-include-backend`
**Status:** active
**Chosen:** the two frontend repos plus gymie's controller, route and DTO paths.
**Considered:** frontend repos only — rejected, it would have missed KAN-16's 400-error defect,
which lived in server code and showed only in the browser; any app with a URL in `## apps` —
rejected as effectively every Gymie change.

### The stage blocks, like lint and test

**ID:** `visual-blocks-the-handoff`
**Status:** active
**Chosen:** a non-zero exit, or a defect the agent sees, blocks the `IN_PROGRESS` handoff.
**Considered:** report-only — rejected, `/flow`'s integrate phase has no verification gate, so a
non-blocking visual stage would be a signal nothing acts on; block on baseline diff but report
infrastructure failure — rejected as a distinction the stage cannot reliably draw, since a browser
that will not start and a page that will not render look the same from outside.

### The agent reads the PNGs, and what it sees can block

**ID:** `agent-reads-the-screenshots`
**Status:** active
**Chosen:** the implementer reads every captured PNG and states per view what it saw; a defect it
sees blocks even when every assertion passed.
**Considered:** capture and commit without looking — rejected, it reduces the stage to a file-writing
step and reproduces KAN-16 exactly, where three broken views passed every automated check.

### The stage starts the stack, and stops only what it started

**ID:** `stage-starts-and-stops-the-stack`
**Status:** active
**Chosen:** probe first; start from `## run` only if nothing answers; stop at the end only if this
stage started it.
**Considered:** require the stack already up and fail loudly — rejected, it makes the operator
sequence a stack against a stage they cannot see coming; start and leave running — rejected, it
leaves a stack up that the run is not accountable for. Stopping only what was started is what keeps
an operator's own session untouched.

### `agents` verifies against the existing UI-test stack

**ID:** `agents-targets-ui-test-stack`
**Status:** active
**Chosen:** `make ui-test-up` / `make ui-test-down` on `127.0.0.1:4174`, against the fixed
`flow_uitest` fixture.
**Considered:** a new dedicated visual stack — rejected, the target already exists and already
guarantees the fixed seed a baseline needs; the dev workspace's own daemon on 4173 — rejected
outright, its data changes with every run, so no baseline over it could ever be stable, and
`CLAUDE.md` forbids this change touching that service at all.

### `agents` commits its baselines to itself

**ID:** `agents-has-no-regression-checkout`
**Status:** active
**Chosen:** `regression checkout` is optional; absent means the spec and baselines are committed to
the change's own branch.
**Considered:** requiring every project to name a second repository — rejected, it forces a repo
split on projects that have no reason for one. Gymie's split exists because a Node toolchain must
not land in a Gradle project, which is a Gymie fact, not a general one.

### The automatic push is dropped; the carve-out cannot be authorised from inside the repo

**ID:** `no-automatic-push`
**Status:** active
**Supersedes:** `push-carve-out-is-declared`, whose `**Status:**` is set to
`superseded by no-automatic-push` above.

**Chosen:** the stage commits the per-change spec and its PNGs to the regression checkout and
**stops there**. The handoff prints the push command for the operator. `push to default branch` is
removed from the contract, from both declarations, and from the stage;
`check-visual-push-gate.sh` and its harness are deleted.

**Why the earlier decision was wrong, demonstrated rather than argued:** the origin-equality check
compared two fields that both live in `<project>/.flow/project.md` — a repo-tracked,
pull-request-editable file. Both sides of the equality are therefore controlled by whoever edits
that file, and they agree by construction: point `regression checkout` at any git checkout already
on the operator's disk, set `regression repo` to that checkout's real `origin`, and the guard
returns `PUSH-GATE-ALLOWED`. A panel slot built the fixture and reproduced it. The guard's own
header claimed it existed to stop "a project declaring a push carve-out aimed at a repository it
does not name"; it did not stop that.

**The root cause is structural, not a bug in the check:** a file inside a repository cannot
authorise a push to another repository. Authorisation has to come from somewhere the pull request
cannot reach.

**Considered:** recording permitted regression repos in the harness settings store, outside the
repo — genuinely fixes the binding and keeps the push automatic, rejected by the operator;
confirming with the operator at push time — robust, rejected because it breaks the unattended
push; accepting and documenting the risk — rejected.

**This reverses an earlier operator decision, and both are recorded.** The operator first chose
"always allow direct pushes to main for gymie-playwright, always merge and push", and chose
`drop the automatic push` after being shown the reproduced attack. The first decision was sound
given what was known; the second supersedes it on evidence.

**`regression repo` survives as an identity assertion, not an authorisation.** It still records
which repository the checkout is expected to be, and `check-visual-verification.sh` still reports a
mismatch — but nothing is granted by it, so an attacker-chosen pair grants nothing.

## Open questions

None.
