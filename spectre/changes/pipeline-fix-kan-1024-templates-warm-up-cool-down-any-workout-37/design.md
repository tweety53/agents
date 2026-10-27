## Context

Evidence: gymie kan-1024, 2026-10-27 — all 7 tasks ticked, record `STARTED`, `reviewBeforeLanding`
true, the parent ended its turn on a report with the suite, panel, verify and
`flow.write-in-progress` ahead. The prose rule (`skills/flow-contracts/pipeline.md`, "A turn ends
only at a gate, a block, or the run's final handoff") has failed five times (kan-1015, kan-1017,
kan-1018, kan-1033, kan-1024).

Class of defect: a turn-boundary rule stated only in prose, with nothing at the harness that refuses
the turn's end.

Source of "this session's /flow change": `hooks/flow-active-change.py`'s `resolve_session_id` and
the store query `GET /api/v1/stage-runs?session_id=<id>&command=/flow&sort=-started_at&limit=1`;
each row carries `changeName`, `stage` and `endedAt` (`stats/internal/api/stages.go`). The change's
record (`GET /api/v1/changes/find?name=<name>`, `stats/internal/api/changes.go`) carries `state` and
`reviewBeforeLanding`; ticked/total task counts come from `flow tasks count` on the change's plan.

Blast radius: 5 files — `hooks/flow-stop-guard.py` (new), its test
<!-- predicted: 5 files, task 1's Files: list -->
`stats/internal/repotest/flow_stop_guard_hook_test.go` (new), `setup.sh`, `README.md` (hook list),
`skills/flow-contracts/pipeline.md`. `git grep -l "A turn ends only at a gate"` finds only
`pipeline.md`; no existing name is renamed. Under 60, so not `big`; fixed in this run. No saved
memory encodes the defect.
<!-- measured: git grep -l "A turn ends only at a gate" @ 027778a2 → skills/flow-contracts/pipeline.md -->

## Decisions

### Mechanism

**ID:** stop-hook-blocks-mid-phase
**Status:** active
**Chosen:** a Claude Code `Stop` hook, `hooks/flow-stop-guard.py`, installed by `setup.sh global`'s existing `install_hooks` loop and registered under `"Stop"` in `settings.json`; it prints `{"decision":"block","reason":"<next stage>"}` to refuse the stop.
**Considered:** more prose in `pipeline.md` — five recorded failures; a `flow` CLI check the parent runs before ending — the parent that forgets the rule forgets the check too.

### When it blocks

**ID:** block-on-open-stage-or-all-ticked
**Status:** active
**Chosen:** block when the session's last `/flow` stage row has `endedAt` null (an open `flow stage begin`), or the change's record is `STARTED` with ticked == total > 0 tasks (implementation done, `flow.write-in-progress` not yet written). The reason names the open stage, or "full suite, review panel, verify, then flow.write-in-progress" for the all-ticked case.
**Considered:** parsing the transcript for the stage sequence — the store already holds it, and `flow-active-change.py` avoids a transcript re-parse for the same reason.

### When it allows

**ID:** allow-at-gates-asks-and-explicit-stop
**Status:** active
**Chosen:** exit 0 with no output when any holds: the last assistant message's final tool use is `AskUserQuestion` (transcript tail); the last operator message matches `\bstop\b` case-insensitively; the record is `IN_PROGRESS` or `FINISHED`; the record is `STARTED` with total == 0 (planning) or with ticked == 0 and no open stage (plan-gate handoff); no `/flow` change for the session; any error (store down, bad input) — fail open like `flow-active-change.py`.
**Considered:** failing closed on a store error — a down flowd would trap every session's turn.

### No infinite loop

**ID:** stop-hook-active-needs-progress
**Status:** active
**Chosen:** on each block the hook writes a fingerprint (last stage-run id, its `endedAt`, ticked count) to `~/.claude/flow-stop-guard/<session_id>.json`; when `stop_hook_active` is true and the current fingerprint equals the stored one, it allows the stop.
**Considered:** allowing whenever `stop_hook_active` is true — one ignored block and the turn ends, the kan-1024 shape again; no state file — nothing else distinguishes "made progress" from "ignored the block".

### Code sharing

**ID:** import-flow-active-change-helpers
**Status:** active
**Chosen:** load `resolve_session_id` and the store address/timeout from `flow-active-change.py` beside it via `importlib.util.spec_from_file_location` (hyphenated name); the stage-run query is extended locally to read `stage`, `endedAt`, `stageRunId`.
**Considered:** copying the helpers — two owners of the session-id resolution; a mode flag on `flow-active-change.py` — one script answering two events with different output contracts.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — a new hook that can trap every session's turn if a branch is wrong.
**Considered:** `flow-low` — the diff is small, but the allow conditions are the risk.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

None.
