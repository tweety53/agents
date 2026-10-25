## Context

- Five findings from kan-1013's self-review pass, planned per item 1 of **3. Fix every finding that
  is not `big`** (`skills/flow-self-review/SKILL.md`) on `self-review-kan-1013-templates-step-3-be-review-states-admin`,
  based at `361d6649`. Each fixable finding is carried by exactly one task; the brainstorm design is
  `.superpowers/sdd/2026-10-25-self-review-kan-1013-templates-step-3-be-review-states-admin-design.md`
  in this worktree. The kan-1013 gymie worktree is
  `/Users/tweety53/Projects/gymie-worktrees/kan-1013-templates-step-3-be-review-states-admin`
  (branch `spectre/kan-1013-templates-step-3-be-review-states-admin`, state `IN_PROGRESS`).
- **Finding 1 (gateway on the shared 8080) — disproved, no task.** The traced start path is
  `## run`'s plain `./gradlew devStart` (`gymie/.flow/project.md`), the stack `/flow` hands off.
  In an isolated workspace `deterministicBlock(declared, offset, isolateOAuthPorts())`
  (`gradle/dev-lifecycle.gradle.kts`, `resolveWorkspaceBlock`) moves the backend and admin panel
  and keeps the OAuth-pinned frontend and gateway on their declared 3000/8080 unless
  `-PisolateOAuthPorts` is passed; `.flow/project.md` states it in two places ("`devStart` only
  moves them under `-PisolateOAuthPorts`", and "The operator's `## run` does not, so the stack
  `/flow` hands off runs on the registered 3000/8080"). Only the pipeline's own `start` row (visual
  verification) passes the flag. The record carried no gateway entry when the pass read it because
  `recordWorkspaceBlock` writes each service's entry once that service is up: re-read at
  2026-10-25 23:45 the worktree's `.dev-stack/port-block` lists `admin-frontend=4421`,
  `backend=9501`, `frontend=3000`, `gateway=8080` — the gateway was recorded, on the port the
  design gives it. Nothing omitted the gateway or the frontend from the block. Placement had there
  been a fix: gymie dev tooling, committed on kan-1013's own branch in its gymie worktree
  (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`).
- **Finding 2 (Task 1), blast radius 3.** `plan-script-paths` (`stats/internal/guard/planscriptpaths.go`)
  checks only `scripts/` citations; no check reads a `**Files:**` path against the tree.
  kan-1013's plan named `src/app/app/src/main/resources/db/migration/V48__template_review.sql`;
  the real directory is `src/app/src/main/resources/db/migration/`. Measured over history with the
  current trees: of 6,235 `**Files:**` paths in gymie's archived plans and 3,249 in this repository's,
  a rule "parent directory absent, and present with one non-final segment dropped, leaving at least
  two segments" fires on 0, and on kan-1013's wrong path it fires (dropping the second `app`).
  Without the two-segment floor it fires once in this repository (kan-363's `cmd/spectre/main.go`,
  a path in the separate spectre repository, which resolves to this repository's top-level
  `spectre/`).
- **Finding 3 (Task 2), blast radius 3.** Traced: kan-1013's recorded decision (`flow record
  decisions`, id 326, recorded 20:57:13Z) says `"execution": "inline"`, `"implementer": "skipped —
  inline"`; the ledger then carries `task-1-implementer` … `task-7-implementer` from 20:59:40Z and
  `panel-fix-1` (commit `ac49f43f`, `Flow-Key: panel-fix-1`), each with a subagent behind it. The
  decision was right and the dispatches broke it: `skills/flow/implement.md`'s inline table
  ("the parent's only permitted dispatches inline are the panel-bundle, verifier and
  tooling-analyst rows") has no mechanical check. Two records then wrote it wrong: `narrative.md`'s
  "Inline execution on opus" and the round-1 pass note "inline fix by parent: F1, F2 (Important)
  fixed in ac49f43f", rendered into `panel.md`. Pass notes are append-only (`runRecordPass`), so the
  repair is an appended correction note, then a re-render.
- **Finding 4 (Task 3), blast radius 6.** kan-1013's decision carries the cap only as one
  `overrides` entry applying to `panel.dispatches[0].effort`; `override` is `null`, and no field
  states a session-wide ceiling. The ledger shows implementers, `panel-fix-1` and `verify` at
  `effort=default`. The handshake (`skills/flow/implement.md`, **The handshake**) compares the model
  alone, since a model cannot report its effort. The records are truthful — those rows ran at
  `default` — so nothing needs repairing.
- **Finding 5 (Task 4), blast radius 5.** `explore-1` (id 7684) and `explore-2` were begun at
  20:54:30.507Z and ended in the same second — no back-fill: a begin and an end sent after both agents
  had returned; the harvester stamped their agents and tokens afterwards. `EndDispatch`
  (`stats/internal/store/records.go`) already refuses this shape for an inline row
  (`ErrInlineWindowTooShort`, `minInlineWindow` = 3 s) but not for a subagent row. `flowd
  retime-dispatch <id> <started> <ended> <commit>` exists for the repair.

## Decisions

### Finding 1 is disproved

**ID:** gateway-8080-by-design
**Status:** active
**Chosen:** no task — the handed-off stack's 3000/8080 is `## run`'s documented behaviour, and the
record lists the gateway (**Context**).
**Considered:** passing `-PisolateOAuthPorts` from `## run` — it moves the handed-off stack off the
ports Google's OAuth client has registered, breaking the operator's login, which is why the project
chose the split.

### Where the Files-path check lives

**ID:** files-path-check-own-command
**Status:** active
**Chosen:** a new `flow-guard plan-files-paths <tasks.md>` in `stats/internal/guard/planfilespaths.go`,
run by `check-plan-shape` beside `plan-script-paths`, with its repository root, archive skip and
placeholder skip taken the same way.
**Considered:** folding it into `plan-script-paths` — that command's contract and message are about
cited `scripts/` paths and its test table would mix two contracts; a new F20 inside `planshape.go`'s
`checkTask` — that file is already the largest guard in the package and the per-file in-process list is the existing seam for
tree-reading checks.

### What the Files-path check flags

**ID:** files-path-dropped-segment
**Status:** active
**Chosen:** a path is flagged when neither it nor its parent directory exists under the repository
root, and dropping one non-final directory segment names an existing directory of at least two
segments; the message names that directory.
**Considered:** flagging every path whose file and parent are both absent (the finding's literal
wording) — it fires on every new package and new module a plan creates (`core/services/warmup/`),
the "must not misfire on files a task creates" constraint; flagging when the nearest existing
ancestor is more than one level up — new modules hit it as well; the rule without the two-segment
floor — misfires on kan-363's cross-repository `cmd/spectre/main.go`.

### Refuse an inline breach at dispatch begin

**ID:** inline-breach-refused-at-begin
**Status:** active
**Chosen:** `flow record dispatch begin` reads the change's latest recorded decision and refuses,
exit 2, a `-role implementer` or `-role panel-fix` row whose `-agent-id` is not `inline` when that
decision's `execution` is `inline`; keys opening `self-review-` or `pipeline-fix-` are exempt, and
a read that fails skips the check, as the panel-round floor does.
**Considered:** correcting the narrative and records alone — the breach would recur unseen; a
`check-loop-dispatch`-style hook on the Agent call — no hook sees a /flow run's decision today, and
the begin row is the one place every dispatch is already recorded.

### Repair kan-1013's execution records

**ID:** k1013-inline-records-repaired
**Status:** active
**Chosen:** append one round-1 pass note naming `panel-fix-1` as the fixer, re-render kan-1013's
records, and correct `narrative.md`'s execution line, in kan-1013's gymie worktree.
**Considered:** leaving the records — a fix repairs what its defect wrote (`skills/flow-self-review/SKILL.md`,
step 3); a store command to rewrite a pass note — pass notes are append-only by design.

### Record the session cap in the decision

**ID:** session-cap-field
**Status:** active
**Chosen:** `decision.json` gains an optional `session_cap` — `{model, effort, instruction}` or
`null`, absent read as `null` — written at **Decide** from any session instruction or saved memory
stating a run-wide ceiling; `flow decision render` validates it; `dispatch begin` refuses a row
whose effort is `default` or ranks above the cap's (`low` < `medium` < `high`), or whose model ranks
above it (`sonnet` < `opus`).
**Considered:** extending `overrides` — each entry replaces one named pair and cannot bind a row
the decision never planned, like kan-1013's implementers; a settings-record cap (`flow settings`) —
harness-wide and outlives the session it was given in; making `session_cap` required — every
earlier decision would stop rendering.

### Refuse an after-the-fact subagent window

**ID:** dispatch-window-harvested
**Status:** active
**Chosen:** `EndDispatch` returns a new `ErrDispatchWindowTooShort` for a `completed` end with no
token report landing under `minInlineWindow` after its begin, on a session whose harness is in
`harvestedHarnesses`, whatever the agent; mapped where `ErrInlineWindowTooShort` is
(`internal/api/server.go`, `internal/reconcile/reconcile.go`).
**Considered:** refusing in `dispatch backfill` alone — kan-1013's rows came through begin and end,
not back-fill; retiming rows in the harvester from the agent transcript — a second writer of a
row's instants, more code than the refusal; every session regardless of harness — store tests and
sessions with a `-tokens` report legitimately close fast.

### Repair kan-1013's explorer rows

**ID:** k1013-explorer-retimed
**Status:** active
**Chosen:** `flowd retime-dispatch` each of the two rows to the Agent call's launch and its result
instants in the kan-1013 parent transcript.
**Considered:** leaving them 0 s — the defect's own output; deleting them — loses their tokens.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Go guard and store logic with refusal semantics a light review would
miss; one fixer dispatch (key `self-review-kan-1013-templates-step-3-be-review-states-admin-fix`)
carrying Tasks 1–4 in order, since Tasks 2, 3 and 4 share `stats/cmd/flow/record.go` or its tests
and run in series anyway.
**Considered:** `flow-low` — the session cap allows it, but the begin-time refusals can lock out
legitimate rows if misjudged; two groups (Task 1, Tasks 2–4) — the dispatches run in series either
way, so a second group buys nothing.

## Open questions

## Recommended approach

Four small refusals at the points that already record or check the thing they guard — the plan-shape
scan for **Files:** paths, `dispatch begin` for execution mode and the session cap, `EndDispatch` for
the dispatch window — each with one test that fails without it, plus the record repairs the defects
left in kan-1013. Rejected: correcting records alone (`inline-breach-refused-at-begin`), a broader
missing-path rule (`files-path-dropped-segment`), extending `overrides` (`session-cap-field`), a
harvester retime (`dispatch-window-harvested`), changing `## run` (`gateway-8080-by-design`).

## Critical files

- `stats/internal/guard/planfilespaths.go` — new `plan-files-paths` check.
- `stats/internal/guard/plan_files_paths_test.go` — its cases.
- `stats/internal/guard/planshape.go` — runs it in the per-file list; header names it.
- `stats/cmd/flow/record.go` — `runRecordDispatchBegin`'s inline and cap refusals.
- `stats/cmd/flow/record_test.go` — their cases.
- `stats/cmd/flow/decision.go`, `stats/cmd/flow/decision_test.go` — `session_cap` validation.
- `skills/flow/brainstorm-planner.md` — **Decide**'s field list gains `session_cap`.
- `skills/flow/implement.md` — **The handshake** names the begin-time cap refusal; the inline table
  names the begin-time refusal.
- `stats/internal/store/records.go`, `stats/internal/store/records_test.go` — `ErrDispatchWindowTooShort`.
- `stats/internal/api/server.go`, `stats/internal/reconcile/reconcile.go` — map the new error.
- kan-1013 gymie worktree: `spectre/changes/archive/kan-1013-templates-step-3-be-review-states-admin/narrative.md`
  and the re-rendered `panel.md`.

## Existing code to reuse

- `planScriptPathsCmd` (`stats/internal/guard/planscriptpaths.go`) — root derivation, archive skip,
  placeholder skip, `tcfIterTasks`/`tcfParseTask`.
- `highestRecordedRound` read in `runRecordDispatchBegin` — the skip-on-read-failure store read
  pattern; `client.ListDecisions`.
- `dMissing`/`dVisualVerify` validation in `stats/cmd/flow/decision.go`.
- `ErrInlineWindowTooShort`, `minInlineWindow`, `harvestedHarnesses` in `stats/internal/store/records.go`.
- `flowd retime-dispatch` (`stats/cmd/flowd/main.go`); `flow record pass`; `flow record render`.

## Verification

- `cd stats && go test ./internal/guard/ -run PlanFilesPaths`, `go test ./cmd/flow/ -run 'RecordDispatchBegin|Decision'`,
  `go test ./internal/store/ -run DispatchWindow` pass, and each new case fails with its fix reverted.
- `scripts/check-plan-shape.sh` over kan-1013's original plan text (the `src/app/app/...` path,
  restored into a scratch `changes/<name>/tasks.md` under a scratch root) exits 1 naming
  `src/app/src/main/resources/db/migration`; over every live change in this repository and gymie's it exits 0.
- `flow record render` for kan-1013 shows `explore-1` and `explore-2` with non-zero durations and
  round 1's correction note.

## Live check

Against a scratch daemon on its own database — `cd stats && make ui-test-up` (port 4174, database
`flow_uitest`), `FLOW_ADDR=http://127.0.0.1:4174`, the CLI built from this worktree
(`go build -o /tmp/sr1013-flow ./cmd/flow`) — never the live `flow` database:

1. Record a decision with `"execution": "inline"` and `"session_cap": {"model":"opus","effort":"low","instruction":"live check"}`
   on scratch change `sr1013-live`; `dispatch begin -role implementer -key task-1-implementer -effort low -model opus`
   exits 2 naming the inline decision; the same with `-agent-id inline` exits 0.
2. `dispatch begin -role reviewer -key panel-0-primary -slot primary -effort medium -model opus`
   exits 2 naming the cap; `-effort low` exits 0.
3. Begin a row on a `claude-code` session token and end it `completed` at once: exit 2 naming the
   window; end it after 4 s: exit 0.

Before the change, steps 1–3's refusals all exit 0 — that is what the change not working looks
like. Then `make ui-test-down`. The kan-1013 repairs are checked on the live store read-only with
`flow record render` (Verification above).
