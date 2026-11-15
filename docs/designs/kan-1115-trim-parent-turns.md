# kan-1115-trim-parent-turns
> Design record of change kan-1115-trim-parent-turns (KAN-1115), written at landing. Each decision's own **Status:** line is the status.

## Context

KAN-1115 (epic KAN-1114) asks for three small, independent trims of the `/flow` parent session's
turns, with no new message schema, no new agent role and no mode switch. The kan-1087 run's parent
made 924 turns, 54 of them `flow record dispatch` calls, 28 `check-plan-unchanged` calls, and about
320 of its 838 tool calls file reads.

All three defects reproduce on the base:

1. **The parent spends a turn recording each dispatch.** `begin_row_reason` in
   `hooks/enforce-agent-baseline.py` denies a flow dispatch with no open ledger row, so every
   non-loop, non-panel dispatch needs a hand-run `flow record dispatch begin` first
   (`skills/flow/brainstorm.md`, `skills/flow/implement.md`, `skills/flow/review-panel.md`,
   `skills/flow/brainstorm-planner.md`), and a hand-run `end` after.
2. **The parent runs `check-plan-unchanged` itself.** The gated reviewer's `snapshot` is a separate
   parent call before every bundle (`skills/flow/implement.md`, **The plan tree survives every
   reviewer dispatch**); the panel round's `verify` is a separate call "once every report file
   exists" (`skills/flow/review-panel.md`). The panel `snapshot` already rides
   `flow-guard panel-round-begins`, and the gated `verify` already rides the bundle's return call.
3. **The parent reads report bodies.** The Read discipline in `skills/flow/implement.md` lets the
   parent read a report body "where a judgement needs it, once, narrowly"; the same carve-out sits in
   `skills/flow/review-panel.md` and `skills/flow/implementer-brief.md`, and
   `skills/flow/review-panel-fix-round.md` transcribes `fix-mutation:` lines "from the fix
   subagent's report".

Constraints the design rests on:

- `check-plan-unchanged verify` is a contractual ordering gate: it runs "before any verdict is read
  or acted on", and a reviewer's own clean-or-restored claim is not evidence
  (`stats/internal/guard/planunchanged.go`). It cannot simply leave the parent; it must ride a call
  the parent still makes before acting on the verdict.
- Stage end of brainstorm, self-review and sdd-tdd, and stage begin of decide, refuse while a
  dispatch row is open (`stats/cmd/flow/stage.go`), so every row still needs an `end`.
- The two `Agent|Task` PreToolUse hooks run side by side; a begin written by a second hook would
  race `enforce-agent-baseline.py`'s open-row check.
- The PreToolUse payload carries `tool_input.{prompt, description, subagent_type, model, effort}`,
  `session_id`, `cwd` and `transcript_path`. It does not carry the `-pair-pick` line
  `flow-guard pair-judge` printed, nor a panel row's `-diff-base`.
- No PostToolUse hook exists, and the SubagentStop hook records nothing and is not shown to name the
  stopping agent's row.

It is one change because all three trims serve one measured cost (parent turns) and the verification
counts them together on the next run against kan-1087.

## Decisions

### Where the plan-tree snapshot and verify ride

**ID:** plan-check-rides-row
**Status:** active
**Chosen:** the snapshot rides the row's begin, the verify rides the row's end. The hook that writes
a `task-*-reviewer*` row's begin first runs `flow-guard check-plan-unchanged snapshot` to a path
derived from the row's key, as `panel-round-begins` already does for panel rows. `flow record
dispatch end` of a gated reviewer or findings-merger row runs the verify in-process first (both live
in the `stats/` module), against a snapshot path it derives from the key alone:
`<worktree>/.superpowers/sdd/plan-snapshot-<key>.txt` for a `task-<ids>-reviewer[-<suffix>]` key, and
`<worktree>/.superpowers/sdd/plan-snapshot-panel-<r>.txt` for a `panel-<r>-merger[-<suffix>]` key,
`<r>` the round `mergerKeyRE` (`stats/cmd/flow/record.go`) captures — the file
`panel-round-begins` wrote for that round. A panel slot row (`panel-<r>-<slot>`) verifies nothing at
its own end: its round's verify rides the merger's end, which `flow record dispatch begin` admits
only once every slot row of the round has ended. On exit 1 or 2 the end refuses with the guard's
lines, and the existing `## Question` stop follows. **A missing snapshot fails closed:** the end
refuses, naming the path it derived, rather than skipping the verify. The one end that skips the
verify is `-outcome aborted`, the outcome for a row whose dispatch never ran. The parent's return call already orders the end before any
verdict is acted on, so the gate holds with no call of its own.
**Considered:** the stage's last worker (gated reviewer or panel merger) runs the verify and reports
its exit — ruled out by the guard's own premise that a dispatched agent's report of the guard is a
claim, not the gate. Contract-only folds (gated snapshot joins the group's boundary call, panel
verify joins the merger's begin call) — ruled out because once the hook writes begins those begin
calls disappear and the folds go with them.

### Which half of the dispatch record moves out of the parent

**ID:** begin-moves-end-stays
**Status:** active
**Chosen:** only the begin moves. The end rides the Bash call the parent already makes on each
return — the gated bundle's return call, `flow-guard close-task`, the panel round's record call —
never a call of its own; its `-outcome` and `-commit` stay the parent's judgement. A return with no
other Bash call (a planner design round, whose relay writes a file) keeps its one end call.
**Considered:** a SubagentStop hook records `-outcome completed` and the parent writes an explicit
end only for `fallback`, `blocked` or a `-commit` — ruled out because no SubagentStop payload is
shown to identify the row's key, an auto-written `completed` the parent must then correct writes an
end twice, and verify riding an end written outside the parent's ordering breaks
`plan-check-rides-row`.

### Where the hook reads the row's pair-pick

**ID:** pair-pick-file
**Status:** active
**Chosen:** `flow-guard pair-judge` saves the line it prints beside the prompt file it was given,
as `<dir of the prompt file>/<key>.pick` (`<key>` the prompt file's base name without `.md`), and
the hook reads `<worktree>/.superpowers/sdd/pair-judge/<key>.pick`, the worktree being
`<main-checkout>-worktrees/<change>`, the main checkout derived from the git common dir as
`project_key` already derives it. No prompt gains a line. A dispatch whose prompt file was not
written there (a `/flow-plan` session before capture, a guard redispatch on `plan-agent-pair`)
records no pick.
**Considered:** every flow dispatch prompt carries a `Pair-pick:` line the hook parses — ruled out
by the ticket's "no new message schema", and every dispatch template would gain a line. Drop
`-pair-pick` from hook-written rows — ruled out because it empties the pair-judge provenance the
stats read (`skills/flow/brainstorm-planner.md`, `docs/designs/kan-1119-pair-judge.md`).

### Which hook writes the begin

**ID:** hook-same-file
**Status:** active
**Chosen:** the begin lives in `hooks/enforce-agent-baseline.py`, replacing `begin_row_reason`'s
"no open row" denial.
**Considered:** a new PreToolUse hook — ruled out because the two `Agent|Task` hooks run side by side
and a separate begin would race the open-row check.

### Where the hook gets the session token and change

**ID:** hook-token-from-stage-run
**Status:** active
**Chosen:** from the session's latest stage run in flowd, which `session_change()` already reads and
whose store record carries the run's `sessionToken`; `GET /api/v1/stage-runs` gains the
`sessionToken` field its DTO drops today (`stageRunDTO`, `stats/internal/api/stats.go`).
**Considered:** a `Session-token:` prompt line — ruled out as a new message line.

### What a refused begin does

**ID:** hook-begin-refusal-denies
**Status:** active
**Chosen:** a begin `flow` refuses (any non-zero exit) denies the launch with `flow`'s stderr verbatim, so the
run gates (`session_cap`, the pair gate, open sdd-tdd, mutator/merger gates) hold as they do today. An
unreachable store behaves as today: `flow record` journals and the hook fails open.
**Considered:** failing open on a refusal — ruled out because it lets through a dispatch
`session_cap` or the pair gate refuses.

### Panel rows

**ID:** panel-round-begins-kept
**Status:** active
**Chosen:** panel rows stay begun by `flow-guard panel-round-begins`, one call per round carrying
`-diff-base`; the hook sees their open rows and does nothing.
**Considered:** moving them into the hook — ruled out because the payload has no diff-base.

### Roles whose begin needs a flag the payload lacks

**ID:** explicit-begin-fallback
**Status:** active
**Chosen:** such a role keeps its explicit begin; the hook sees the open row and does nothing.
**Considered:** adding prompt lines for those flags — ruled out as new message schema.

### Planner Explore and Plan agents

**ID:** planning-agents-hook-begun
**Status:** active
**Chosen:** a planner's Explore and Plan agents are begun by the hook too, under their
`explore-<n>|design-<n>` key; the hook's `PLANNING_AGENT_TYPES` early return no longer skips them.
**Considered:** leaving them hand-recorded — ruled out since they are part of the 54 calls.

### The Read discipline

**ID:** read-discipline-absolute
**Status:** active
**Chosen:** the Read discipline's report bullet becomes "the parent opens no report or evidence file
a dispatched agent wrote". The verdict is the final message: the implementer/fixer head, the gated
per-pass line, the verifier's `## Verdict`, the panel's merger summary. A judgement that needs a
report body goes to the agent that is passed the path (the fixer via `FIX INPUT:`, the merger, the
planner). The narrow-read carve-outs in `skills/flow/implement.md`, `skills/flow/review-panel.md` and
`skills/flow/implementer-brief.md` are deleted, and `skills/flow/review-panel-fix-round.md` reads
`fix-mutation:` from the final message head.
**Considered:** keeping the narrow-read carve-out — ruled out because the ticket asks to remove it.

### No new status line for panel slots

**ID:** no-new-status-lines
**Status:** active
**Chosen:** panel slots get no new final-message status line; their verdict is the merger summary.
**Considered:** adding one — ruled out by the ticket's "no new message schema".

### Reads that are not the parent's own

**ID:** guard-and-operator-reads-kept
**Status:** active
**Chosen:** a file a guard or `flow` parses (`-from findings-*.jsonl`, `check-rerun-citations`,
`check-verify-report`, `check-live-verification`) and output the operator must see (self-review
`## Offered`, a check's `tail`) stay as they are.
**Considered:** routing them through an agent — ruled out because they are not parent judgement
reads.

### live-before.md

**ID:** live-before-kept
**Status:** active
**Chosen:** `live-before.md` is the parent's own measurement, so the parent still compares its sha
and copies it.
**Considered:** handing it to an agent — ruled out because it is not a dispatched agent's report.

### How the saving is verified

**ID:** verification-by-next-run
**Status:** active
**Chosen:** the next real `/flow` run's parent turns, report reads, dispatch-record calls and
`check-plan-unchanged` calls are counted against kan-1087 with the existing transcript stats.
**Considered:** new measurement tooling — ruled out as unrequested.

## Open questions

## Recommended approach

Three trims, one per ticket item.

**Trim 1 — the hook writes the begin.** `begin_row_reason` in `hooks/enforce-agent-baseline.py`
stops denying "no open row" and runs `flow record dispatch begin` with the fields it derives: the
key is the prompt's own column-0 `Key: <key>` line where it carries one (the implementer and gated
reviewer stubs do, and only that line tells an implementer group from its reviewer bundle, whose
descriptions both read `Task <ids>/<n>`), else the description's first word, `-task <ids>` from a
`task-<ids>-…` key; a `Task <ids>/<n>` description with no `Key:` line derives no key and keeps
today's "no open row" denial; the role is the prompt's `Role:` line mapped back as
`launchRoleID` maps it (`stats/cmd/flow/record.go`); the model is `tool_input.model`; the effort is
the `flow-<effort>` subagent type, or `tool_input.effort` for Explore/Plan; the session token and
change come from the session's latest stage run (`hook-token-from-stage-run`); the pair-pick from
`pair-judge/<key>.pick` (`pair-pick-file`). It writes the begin only after its own checks — the
description, the loop key line and the pair check — have passed, so a launch it denies leaves no
row. It does nothing when an open row already exists for the
key (panel rows, explicit-begin roles), for a loop dispatch (`enforce-loop-dispatch.py` records
those), when the session's change is FINISHED, and for an Explore or Plan agent whose prompt does not name
the session change's worktree (`<main-checkout>-worktrees/<change>`) — a self-review or in-run fix
planner's agents name their loop's worktree and keep the planner's explicit begin under the loop's
change. A refused begin denies the launch
(`hook-begin-refusal-denies`). Every parent `flow record dispatch begin` block in the contracts is
deleted, except the explicit-begin roles' (`explicit-begin-fallback`). The end stays the parent's,
riding the Bash call it already makes (`begin-moves-end-stays`).

**Trim 2 — the plan check rides the row.** For a `task-*-reviewer*` key the hook runs
`flow-guard check-plan-unchanged snapshot` before writing the begin; `flow record dispatch end` of a
gated reviewer or merger row runs the verify in-process against the snapshot its key names, and
refuses on exit 1 or 2 or a missing snapshot (`plan-check-rides-row`). The contracts' separate `snapshot` and panel `verify` calls are
deleted; `check-tree-markers` is untouched and stays on the return call.

**Trim 3 — the parent reads no report.** The Read discipline in `skills/flow/implement.md` becomes
absolute and the carve-outs go (`read-discipline-absolute`), keeping guard and operator reads
(`guard-and-operator-reads-kept`) and `live-before.md` (`live-before-kept`).

Rejected alternatives: a SubagentStop end (`begin-moves-end-stays`), a worker-run verify or
contract-only folds (`plan-check-rides-row`), a `Pair-pick:` prompt line or no pick
(`pair-pick-file`), a second hook (`hook-same-file`).

## Critical files

- `hooks/enforce-agent-baseline.py` — `begin_row_reason` writes the begin, takes the reviewer
  snapshot and reads the `.pick` file; `session_change` also returns the session token.
- `stats/internal/repotest/enforce_agent_baseline_hook_test.go` — the hook's tests.
- `stats/internal/api/stats.go` — `stageRunDTO` carries `sessionToken`, with its test in
  `stats/internal/api/stats_test.go`.
- `stats/cmd/flow/record.go` — `flow record dispatch end` runs the in-process verify for a reviewer
  or merger row; `launchRoleID` is the role mapping the hook mirrors.
- `stats/cmd/flow/record_test.go` — the end's verify tests.
- `stats/internal/guard/planunchanged.go` — `checkPlanUnchanged`, the snapshot/verify the hook and
  the end call.
- `stats/internal/guard/panelroundbegins.go` — the existing panel snapshot path the end resolves.
- `stats/internal/guard/pairjudge.go`, `stats/internal/guard/pairjudge_test.go` — saving the
  `.pick` file.
- `skills/flow/implement.md` — dispatch begin blocks, the gated snapshot/verify rule, the Read
  discipline.
- `skills/flow/review-panel.md` — the panel verify call, the Read carve-out, dispatch begin blocks.
- `skills/flow/review-panel-fix-round.md` — `fix-mutation:` transcription, begin blocks.
- `skills/flow/implementer-brief.md` — the Read carve-out.
- `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md` — the planner, plan-review and
  planning-agent begin blocks. The begins in `skills/flow/verify-and-handoff.md` (`-agent-id
  inline`), `skills/flow-interactive/SKILL.md` (`-request`), `skills/flow-contracts/pipeline.md`
  (loop rows) and the visual-verify guards stay (`explicit-begin-fallback`); the mentions in
  `skills/flow/sdd-dispatch.md` and `skills/flow/verify-fix-loop.md` name begin refusals, which
  hold for a hook-written begin unchanged.
- `skills/flow/SKILL-rationale.md` — records the panel verify's reasoning, updated with the rule.
- `README.md` — names `check-plan-unchanged` around each subagent.

## Existing code to reuse

- `hooks/enforce-agent-baseline.py`: `session_change`, `project_key`, `change_finished`,
  `in_flow_run`, `TASK_DESCRIPTION`, `ROW_DESCRIPTION`, `LOOP_KEY_LINE`, `loop_word`, `bound`.
- `stats/internal/guard/planunchanged.go`: `checkPlanUnchanged` (registered as
  `Registry["check-plan-unchanged"]`), called in-process by `flow record dispatch end` — `record.go`
  already imports `internal/guard`.
- `stats/internal/guard/panelroundbegins.go`: the snapshot-before-begin pattern and the
  `plan-snapshot-panel-<round>.txt` path.
- `stats/cmd/flow/record.go`: `launchRoleID`, the existing begin refusals the hook surfaces.
- `scripts/flow-guard live-store-copy`: the live check's stand-in store.

## Verification

- Each task's own lint and targeted tests (writing-plans fills them in).
- `flow-guard run-verify-list <worktree> lint` over the contract edits.
- The `## Live check` below, at `flow.verify`.
- The saving itself, per `verification-by-next-run`: on the next real `/flow` run, count the
  parent's turns, its Read calls on report and evidence paths, its `flow record dispatch` calls and
  its `check-plan-unchanged` calls from its transcript with the existing transcript stats, against
  kan-1087's 924 turns, 54 dispatch-record calls and 28 `check-plan-unchanged` calls. Failure looks
  like dispatch-record or `check-plan-unchanged` parent calls still near those figures.

## Non-goals

- The dispatch's `end` stays a parent action (`begin-moves-end-stays`); no SubagentStop or
  PostToolUse hook is added.
- Panel rows stay begun by `flow-guard panel-round-begins` (`panel-round-begins-kept`).
- `check-tree-markers` is not moved.
- No new message lines, status lines or agent roles (`no-new-status-lines`).
- The parent's model (KAN-1074) and the implementer's cost (KAN-1129) are out of scope.

## Risks

- **A begin the hook writes but the dispatch never launches** — another PreToolUse hook denies the
  same call after the row is written. The row stays open and blocks the stage end. Limited because
  the hook writes the begin only after its own checks pass; a re-sent call finds the open row and
  writes nothing; the parent can end the stale row `aborted`, the one outcome that skips the
  plan-tree verify (`plan-check-rides-row`).
- **The stage-run lookup misses** (store unreachable, no stage run for the session) — the hook
  fails open with no row, and the stage end later refuses or the parent's end finds no row. Same
  behaviour as today's fail-open path.
- **A reviewer or merger row ended without a snapshot file** (snapshot failed, or a row begun
  before this change lands) — the end refuses (`plan-check-rides-row`), so the row stays open and
  holds its stage end until the parent ends it `aborted` or re-runs the dispatch. Limited by the hook
  denying the launch when the snapshot fails, so no gated reviewer row exists without its snapshot.
- **A stale `.pick` file** from an earlier dispatch under the same key — the row records the earlier
  pick. Limited because pair-judge overwrites the file on every judgement and keys are unique per
  dispatch.
- **A loop planner's Explore or Plan agent whose prompt also names the run's worktree** — the hook
  begins it under the run's change as well as the planner's own begin under the loop's change, a
  stray row the run's stage end then refuses until the parent ends it `aborted`.
- **Overlap with KAN-1110**, which edits the same contract files — merge conflicts at integrate.

## Live check

Run from this change's worktree against a copy of the dev store, never the store itself:
`scripts/flow-guard live-store-copy up kan1115-live` and export what it prints, so `flow` and
`FLOW_ADDR` point at the copy and a scratch change `kan1115-live` exists there. Then:

1. Mark a stage run for the scratch change with `flow stage begin -command '/flow' -stage
   flow.sdd-tdd -harness claude-code -session-token mf-kan1115-live -session kan1115-live-session
   kan1115-live`, write a pair-judge prompt file and its `.pick` line under the scratch change's
   `.superpowers/sdd/pair-judge/`, and feed `hooks/enforce-agent-baseline.py` a synthesized
   PreToolUse `Agent` payload whose `session_id` is `kan1115-live-session` for key `task-1-implementer` (`Role: implementer`, a `flow-low`
   subagent type, a model the role admits, a `transcript_path` naming a scratch transcript that holds
   one typed `/flow` user turn). Record `flow record dispatches -change kan1115-live`
   before and after.
   - Before (base hook): the call is denied "no open ledger row", 0 rows for the key.
   - After: the call is allowed and exactly 1 open row exists with that key, role, model, effort,
     session token and the `.pick` line as its pair-pick.
   - Not working looks like: an allow with 0 rows, or a row with an empty pair-pick.
2. Re-send the same payload: still 1 open row (no duplicate).
3. Feed a payload for `task-1-reviewer`: the plan-snapshot file for that key now exists, and 1 open
   reviewer row. Edit a non-checkbox line of the scratch change's plan tree, then run
   `flow record dispatch end -key task-1-reviewer -outcome completed …`: it exits non-zero with
   `check-plan-unchanged`'s lines and the row stays open. Restore the line and end again: exit 0,
   row closed.
   - Not working looks like: the first end exiting 0 over the edited plan tree.
   Then begin a `panel-9-merger` row by hand with no `plan-snapshot-panel-9.txt` present and end it
   `-outcome completed`: it exits non-zero naming that path and the row stays open; end it
   `-outcome aborted`: exit 0.
   - Not working looks like: the `completed` end exiting 0 with no snapshot file.
4. Feed a payload the run gates refuse (a model the role does not admit): the hook denies with
   `flow`'s stderr and no row is written.

`scripts/flow-guard live-store-copy down kan1115-live` afterwards.

<!-- measured: `scripts/flow-guard live-store-copy` is declared in .flow/project.md `## run` (line 60) and registered as Registry["live-store-copy"] in stats/internal/guard/livestorecopy.go:65 @ worktree HEAD -->
<!-- measured: `flow stage begin` takes `-session <id>` and stores it as the run's SessionID (stats/cmd/flow/stage.go:453-458), and the hook's session_change queries /api/v1/stage-runs?session_id=<payload session_id> (hooks/enforce-agent-baseline.py:228-235) @ worktree HEAD -->
<!-- measured: in_flow_run admits a payload whose transcript_path holds one JSONL line `{"type":"user","message":{"content":"<command-name>/flow</command-name>"}}`-shaped user turn (is_flow_command, hooks/enforce-agent-baseline.py:378-388), so the check writes such a scratch transcript beside the copy and names it as the payload's transcript_path @ worktree HEAD -->

