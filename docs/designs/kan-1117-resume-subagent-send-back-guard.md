# kan-1117-resume-subagent-send-back-guard
> Design record of change kan-1117-resume-subagent-send-back-guard (KAN-1117), written at landing. Each decision's own **Status:** line is the status.

## Context

KAN-1117. The approved design is
`.superpowers/sdd/planner/round-2.md` (approved in `answers-2.md`); the operator chose the flowd
dispatch row as the guard's source (round 1 Q1) and contract text plus a guard, with no hook, as
the enforcement (round 1 Q2). Nothing user-visible changes, so no prototype was built.

Facts the design rests on, read in code:

- **Age.** A row's `endedAt` is stamped by the daemon's clock when `flow record dispatch end`
  closes it (`stats/cmd/flow/record.go`, the "start and end instants are NOT caller inputs"
  paragraph). Both send-back sites close the row before they detect the send-back:
  `flow-guard close-task` runs the implementer's `dispatch end` before it refuses
  (`skills/flow/implement.md`; for a wave member, the `-end-key` pair closes its row at the pick,
  before the guard failure, `skills/flow/sdd-dispatch.md`), and `check-rerun-citations` runs as the re-run's `dispatch end`
  is recorded (`skills/flow/review-panel-fix-round.md`).
- **Context size.** `metrics.signals.sidechain.context_end` is a string leaf (last-write-wins
  under `jsonb_deep_add`) holding `input_tokens + cache_read_input_tokens +
  cache_creation_input_tokens` of the latest non-synthetic assistant record the row's window
  received (`stats/internal/harvest/signals.go`; dispatch patch in
  `stats/internal/harvest/attribute.go`). A resumed agent's records go to its latest row
  (`attributeAgentFileRecords`), so a second send-back on an already-resumed agent reads the
  latest context from the latest row.
- **Agent id.** `agentId` is filled by the daemon from the launch's tool result (`record.go`,
  `-agent-id` paragraph); it is empty, `none` or `inline` where the harness exposes none.
- **Harvest timing.** flowd's watcher reads transcripts every 5 s (`stats/cmd/flowd/main.go`
  `harvestInterval`), attributing `subagents/agent-<id>.jsonl` by agent id. At every site the
  guard runs after close-task's checks or the citation check, so normally well after one
  tick.
- **Read path.** Guards read rows through `Env.Dispatches`, else `flow record dispatches -change
  <name> -C <worktree>` (`stats/internal/guard/flowcli.go` `pfdRead`).

### The guard

`stats/internal/guard/resumeorfresh.go`, registered as `Registry["resume-or-fresh"]`, tested
beside it:

```text verified:run on the scratch copy kan-1117-resume-subagent-send-back-guard-scratch at 0419a52b against this change's live rows
flow-guard resume-or-fresh <worktree> <change-name> <key>
```

- `<key>` is the key of the row the parent just closed: the failed implementer group's
  `task-<ids>-implementer` (a wave member's own row), or the failed re-run's `panel-<round>-<…>`.
  When an earlier send-back already resumed the agent, `<key>` is that resume's row key.
- It reads the rows with `pfdRead(env, env.Dispatches, "dispatches", name, worktree)` and takes
  the highest-`seq` row whose `key` equals `<key>`, decoding only `agentId`, `endedAt` and
  `metrics.signals.sidechain.context_end` into a local struct.
- Exit 0, stdout `resume <agentId>`, when all hold: the row exists; `agentId` is a harness id
  (not empty, `none` or `inline`); `endedAt` is set; `now − endedAt < 240s`; `context_end` is
  present and `< 100000`.
- Exit 1, stdout `fresh: <reason>`, otherwise — the first failing check, in this order:
  1. `rows unreadable` — the store could not be reached;
  2. `no row <key>`;
  3. `no agent id` — a non-Claude harness or an inline row;
  4. `row open`;
  5. `age <n>s >= 240s`;
  6. `context unknown` — harvest lag, or a harness whose transcript is not harvested;
  7. `context <n> >= 100000`.
- Exit 2: a wrong argument count only.
- Both thresholds are constants in the file.

### Where it runs

Only for a send-back the parent finds at the agent's own hand-back, after that row's
`dispatch end` and before any other dispatch or wait:

| Site | Today | With the guard |
|---|---|---|
| `flow-guard close-task` exit 1 → `task-<n>-implementer-recommit-<k>` (`implement.md` **Every fix is a fresh one-shot**) | fresh fixer | guard on the group's row; `resume <id>` → `SendMessage` to `<id>` |
| wave pick conflict → `task-<ids>-implementer-pick-<k>` | fresh fixer | unchanged — fresh fixer (`pick-conflict-stays-fresh`) |
| `check-rerun-citations` exit 1 → `panel-<round>-<…>-cite` (`review-panel-fix-round.md`) | always resumed | guard on the failed re-run's row; `fresh: …` → fresh dispatch on the same key and pair, carrying the re-run's report path and every `UNCITED:` line |

### Recording

A resumed send-back keeps the row key the fresh dispatch would have used; its `begin` passes
`-agent-id <the id the guard printed>`, which replaces the `-cite` site's separate `flow record
dispatches` lookup. The resume's prompt carries only the send-back input — the refusal text or the
`UNCITED:` lines, plus the `FIX INPUT:` line — because the agent already holds its brief.

### Contract edits

- `skills/flow/implement.md`, "Every dispatch is one-shot" paragraph: gains the exception — an
  immediate send-back runs `flow-guard resume-or-fresh <worktree> <change> <closed row's key>`,
  resumes on exit 0, and goes fresh on any other exit.
- `skills/flow/implement.md`, **Every fix is a fresh one-shot** table: the recommit row says
  the same; the pick row is unchanged.
- `skills/flow/implement.md`, **Every fix is a fresh one-shot** opening: "never a resumed child"
  gains "save an immediate send-back the guard answered `resume` for".
- `skills/flow/implement.md`, "No child … resumes another": unchanged — the parent resumes, not a
  child.
- `skills/flow-contracts/pipeline.md`: the `SendMessage`-to-a-finished-child bans stay (they
  govern orchestrators and loops — `loop-contracts-stay-one-shot`).
- `skills/flow/review-panel-fix-round.md`: the `-cite` send-back goes behind the guard.
- `skills/flow/review-panel.md`: the `-agent-id` exception names the resumed recommit row beside
  `-cite`.
- `skills/flow/sdd-dispatch.md`, **As wave members return**: a guard failure's re-commit runs the
  guard on the member's row as the recommit row does; a pick conflict still goes to a fresh fixer.
- `stats/internal/repotest/cite_send_back_test.go`: updated to the new `-cite` wording.
- `.flow/project.md`: no change — the guard is a runtime decision, not a lint entry.

## Decisions

### Read age and context from the flowd dispatch row

**ID:** guard-source-dispatch-row
**Status:** active
**Chosen:** the flowd dispatch row (`endedAt`, `signals.sidechain.context_end`, `agentId`) — the operator's answer to round 1 Q1; the row already carries every fact at hand-back, so no store, harvest or CLI change is needed.
**Considered:** reading the agent's transcript directly (round 1's recommendation) — not chosen by the operator; it would duplicate harvest logic the daemon already runs.

### Enforce through contract text plus the guard

**ID:** enforcement-contract-plus-guard
**Status:** active
**Chosen:** contract text at the send-back sites plus the `resume-or-fresh` guard, no hook — the operator's answer to round 1 Q2.
**Considered:** a hook that intercepts `SendMessage` — more machinery for two call sites the contracts already name.

### Thresholds are 240 s and 100 000 tokens, constants

**ID:** thresholds-constants
**Status:** active
**Chosen:** 240 s last-turn age and 100 000 context tokens, constants in `resumeorfresh.go` — 60 s margin under the 300 s cache TTL; the ticket proposes 100k.
**Considered:** a configurable `.flow/project.md` setting — configurability nobody asked for.

### Context size is the last assistant record's input total

**ID:** context-measure-last-record
**Status:** superseded by context-from-row-signal
**Chosen:** the last assistant record's `input_tokens + cache_read_input_tokens + cache_creation_input_tokens`, read from the transcript.
**Considered:** summing every turn's tokens — counts every turn's re-read, not the final context.
**Superseded because:** the operator chose the dispatch row as the source (round 1 Q1); the row's existing `context_end` signal already holds this same measure.

### Two immediate send-back sites

**ID:** immediate-send-back-sites
**Status:** active
**Chosen:** the close-task recommit and the `-cite` send-back — the send-backs found at the agent's own hand-back.
**Considered:** the wave pick conflict — rejected in `pick-conflict-stays-fresh`; extending to reviewer `fix` verdicts or full-suite fixes — those are found after a wait, when the cache has lapsed (the ticket's Decided line).

### The wave pick conflict stays fresh

**ID:** pick-conflict-stays-fresh
**Status:** active
**Chosen:** drop the pick-conflict site; its fixer stays a fresh one-shot (plan-review-1 F4).
**Considered:** closing the member's row before the guard at that site — the row closes only at its pick (`skills/flow/sdd-dispatch.md`, "each member's row closes when it is picked, never at the handoff"), so the guard would always answer `fresh: row open`; closing it earlier changes the pick-time close contract for one rarely hit site, and the agent there did not cause the conflict, so its context buys little.

### The existing -cite resume goes behind the guard

**ID:** cite-behind-guard
**Status:** active
**Chosen:** the `-cite` send-back resumes only on the guard's `resume`, else goes fresh on the same key and pair.
**Considered:** leaving it unconditionally resumed — it then resumes a large or cache-lapsed agent, the cost the guard exists to stop.

### Loop contracts stay one-shot

**ID:** loop-contracts-stay-one-shot
**Status:** active
**Chosen:** the pipeline-fix loop, `flow-self-review` and `flow-prompt-tuning` stay strictly one-shot; `pipeline.md`'s `SendMessage` bans stand.
**Considered:** extending the exception to them — the ticket observed `implement.md` only, and their planners are typically over 100k anyway.

### Unknown facts answer fresh, never fail the run (transcript form)

**ID:** unknown-is-fresh-transcript
**Status:** superseded by unknown-is-fresh
**Chosen:** no transcript or a non-Claude harness → `fresh` exit 1, never exit 2.
**Considered:** failing the run — fresh is always safe.
**Superseded because:** the source moved from the transcript to the dispatch row (round 1 Q1), so the unknown cases are the row's.

### Unknown facts answer fresh; exit 2 is usage only

**ID:** unknown-is-fresh
**Status:** active
**Chosen:** every unknown fact → exit 1 `fresh: <reason>` — an unreachable store, no agent id, an open row, context not yet harvested; exit 2 is a wrong argument count only — fresh is always safe here.
**Considered:** exit 2 on a store failure, as `check-panel-fix-single-dispatch` does — stops a run that a fresh dispatch would carry on; waiting on the harvester — a sleep for a case fresh already covers.

### A resumed row keeps the fresh row's key

**ID:** resume-keeps-fresh-key
**Status:** active
**Chosen:** a resumed send-back keeps the key the fresh dispatch would use and begins with `-agent-id <resumed id>`, as `-cite` does today.
**Considered:** a new `-resume` key suffix — a second key shape for the same send-back, and every key consumer would need to learn it.

### Transcript decoded locally in the guard

**ID:** local-decode-transcript
**Status:** superseded by local-decode-row
**Chosen:** the guard decodes the transcript itself.
**Considered:** importing `internal/harvest` — grows `scripts/flow-guard`'s bootstrap package list.
**Superseded because:** the source moved from the transcript to the dispatch row (round 1 Q1).

### Row JSON decoded locally in the guard

**ID:** local-decode-row
**Status:** active
**Chosen:** the guard decodes `flow record dispatches`' JSON into a local struct of `agentId`, `endedAt` and `metrics.signals.sidechain.context_end`.
**Considered:** importing `internal/records` — grows `scripts/flow-guard`'s bootstrap package list.

### The 1-hour subagent cache tier is out of scope

**ID:** one-hour-tier-out-of-scope
**Status:** active
**Chosen:** not investigated in this change — the ticket's Decided line rules out keep-warm and cross-round resume.
**Considered:** checking whether a subagent's cache tier can be set to 1 hour — a different lever with its own cost trade, outside the ticket.

### The guard takes the closed row's key

**ID:** guard-takes-row-key
**Status:** active
**Chosen:** the argument is the closed row's key, and the guard prints `resume <agentId>` — the caller already holds the key.
**Considered:** taking an agent id — every caller would have to look it up first.

### Age is measured from the row's endedAt

**ID:** age-from-ended-at
**Status:** active
**Chosen:** `now − endedAt`.
**Considered:** adding a last-record timestamp to the harvested signals — a store and harvest change for a gap of seconds, which the 60 s margin absorbs.

### Context size is the row's existing context_end signal

**ID:** context-from-row-signal
**Status:** active
**Chosen:** `metrics.signals.sidechain.context_end`, the last non-synthetic assistant record's input total.
**Considered:** summing the row's token totals — counts every turn's re-read rather than the final context.

### Fixtures time endedAt from time.Now()

**ID:** fixtures-relative-clock
**Status:** active
**Chosen:** fixtures set `endedAt` relative to `time.Now()` at 10 s and 600 s against the 240 s bound — the margins make the test deterministic without a seam.
**Considered:** a clock seam on `Env` — an abstraction with one caller.

## Open questions

## Recommended approach

One Go guard, `resume-or-fresh`, reads the closed dispatch row through the existing `pfdRead` /
`Env.Dispatches` seam and answers `resume <agentId>` or `fresh: <reason>`; the two immediate
send-back sites in `implement.md`, `sdd-dispatch.md` and `review-panel-fix-round.md` run it and resume with
`SendMessage` on exit 0. Rejected: a transcript read (`guard-source-dispatch-row`), a hook
(`enforcement-contract-plus-guard`), configurable thresholds (`thresholds-constants`), a
`-resume` key (`resume-keeps-fresh-key`), importing `internal/records` (`local-decode-row`), a
clock seam (`fixtures-relative-clock`).

## Critical files

- `stats/internal/guard/resumeorfresh.go` — new: the guard and its `Registry` entry.
- `stats/internal/guard/resumeorfresh_test.go` — new: its tests, `Env.Dispatches` fed fixture JSON.
- `stats/internal/guard/flowcli.go` — `pfdRead`, the read path the guard reuses (unchanged).
- `stats/internal/guard/guard.go` — `Env.Dispatches` seam and `Registry` (unchanged).
- `skills/flow/implement.md` — the one-shot paragraph, the **Every fix is a fresh one-shot**
  opening and table.
- `skills/flow/review-panel-fix-round.md` — the `-cite` send-back.
- `skills/flow/review-panel.md` — the `-agent-id` exception.
- `skills/flow/sdd-dispatch.md` — the wave guard-failure re-commit.
- `stats/internal/repotest/cite_send_back_test.go` — pins the `-cite` wording; updated with it.
- `skills/flow-contracts/pipeline.md` — its `SendMessage` bans, read and left as they are.
- `README.md` — **Guardrails** › **Implement:** names the new guard.

## Existing code to reuse

- `pfdRead` (`stats/internal/guard/flowcli.go`) — reads `flow record dispatches` or the
  `Env.Dispatches` hook.
- `Env.Dispatches` (`stats/internal/guard/guard.go`) — the test seam other guards' tests already
  feed fixture JSON through (e.g. `check_task_reviewer_single_dispatch_test.go`).
- `Registry` registration in an `init()`, as `stats/internal/guard/reruncitations.go` does.
- The `-cite` send-back's existing `SendMessage` + `-agent-id` mechanics
  (`skills/flow/review-panel-fix-round.md`), generalised to the recommit site.

## Verification

- `go -C stats test -count=1 ./internal/guard/ -run 'TestResumeOrFresh'` — every `fresh:` reason, the
  `resume` case, highest-`seq` selection and exit 2 pass.
- `go -C stats test -count=1 ./internal/repotest/ -run 'TestSendBacksRunResumeOrFresh'` — the
  send-back wording pin passes.
- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean.
- `scripts/flow-guard check-references` and every other `## lint` guard in `.flow/project.md` pass
  over the edited contracts.
- The **Live check** below.

## Live check

The guard reads the running flowd's dispatch rows, so it is exercised against the dev store's real
rows of this change, through the worktree's own `<worktree>/scripts/flow-guard`, which builds the
guard from the worktree — no install and no flowd restart, since no store, harvest or CLI code
changes. At `flow.verify`, from the worktree, run
`scripts/flow-guard resume-or-fresh "$PWD" kan-1117-resume-subagent-send-back-guard <key>` for:

1. the most recently closed row of this run (`flow record dispatches -change
   kan-1117-resume-subagent-send-back-guard -C "$PWD"`, the highest-`seq` row with `endedAt` set)
   — record its `endedAt` age and `context_end` beside the guard's line: `resume <its agentId>` when
   both are under 240 s and 100 000, else the matching `fresh: age …` or `fresh: context …`;
2. `plan-1` — ended long before verify: `fresh: age <n>s >= 240s`, exit 1;
3. the verify stage's own open row, if one is open (the highest-`seq` row with no `endedAt`) —
   `fresh: row open`, exit 1;
4. `no-such-key` — `fresh: no row no-such-key`, exit 1.

Not working looks like: exit 2 or `unknown guard` (the guard is not registered); `resume` on
`plan-1` or on an open row (the age or open check is missing); `fresh: context unknown` on a closed
row whose `context_end` the store shows (the signal path is decoded wrong); or `fresh: rows
unreadable` while `flow record dispatches` answers (the row shape is decoded wrong). At plan time
the scratch copy printed `fresh: age 637s >= 240s` for `plan-1`, `fresh: row open` for the open
`plan-5`, `resume aeca4990e26c2bcc0` for the just-closed `plan-review-1` and `fresh: no row nope`.

## Non-goals

- Resume across fix rounds or after any wait, and keep-warm pings — ruled out by the ticket.
- The 1-hour subagent cache tier (`one-hour-tier-out-of-scope`).
- The wave pick-conflict fixer (`pick-conflict-stays-fresh`), reviewer `fix` verdicts, gated-pass fixes, full-suite fixes, panel-fix rounds and every loop
  contract stay one-shot (`immediate-send-back-sites`, `loop-contracts-stay-one-shot`).
- No store, harvest, CLI or `.flow/project.md` change.

## Risks

- `endedAt` is when the parent closed the row, not the agent's last turn, so the measured age
  falls short by the parent's reaction time — seconds at these sites, absorbed by the 60 s margin
  under the 300 s TTL.
- A harvest tick that split the agent's last turn from earlier ones leaves `context_end` one turn
  behind — a few thousand tokens against a 100 000 bound.
- A send-back found before the harvester's first tick answers `fresh: context unknown` — the
  saving is lost for that send-back, never correctness.
