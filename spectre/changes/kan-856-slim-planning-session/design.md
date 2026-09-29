# Design — kan-856-slim-planning-session

Row ids are the audit's (`audit-planning.md`, `## Candidates`). The audit ran at `700e184`. Every
row was re-located by quoted text on `966d2573`.

## Decisions

### D1 — one file per load condition, siblings in `skills/flow/`

**ID:** split-files-by-load-condition
**Status:** active
**Chosen:** `resume.md`, `withdrawal.md` and `seeded-note.md`, each behind a `**Load …** only when`
directive at its offer point. They live in `skills/flow/`, so `check-guard-symlinks` rule 2 still
sees their guard citations. The reachability check shares `withdrawal.md` with the route it offers.
**Considered:**
- One "planning-conditional" file. Rejected: the resume condition is true on every implementation
  entry, so the withdrawal text would ride along into implementation.

### D2 — BP-20 (sdd-only Decide steps) stays in place

**ID:** bp20-deferred
**Status:** active
**Chosen:** keep steps 2 and 4 inline.
**Considered:**
- Moving them. Rejected: steps 2 and 4 also state the inline path's recorded values
  (`skipped — inline`, `groups … null`), which the decision JSON schema refers to as "above".
  Moving them would hide facts that every inline run records. The save is 2 KB for ~2–5% of runs.

### D3 — BP-08 (task-shape sentence) stays

**ID:** bp08-kept
**Status:** active
**Chosen:** keep it.
**Considered:**
- Cutting it. Rejected: writing-plans is invoked before `build-green.md` is loaded. Cutting the
  shape would change what writing-plans is told.

### D4 — `brainstorm.md` stops restating the planner's marks

**ID:** marks-live-in-planner
**Status:** active
**Chosen:**
- Cut BR-13/BR-14.
- Add `brainstorm-planner.md` to `smcCandidates`, so the only `flow.design-approval`,
  `flow.create-artifacts` and `flow.writing-plans` begin lines are still scanned.
- Split `stage-keys.md`'s row to match.

**Considered:**
- Keeping the copies. Rejected: two copies of a mark block drift.

### D5 — the Jira split is by session, not by topic

**ID:** jira-split-by-session
**Status:** active
**Chosen:** `jira-integration-finish.md` holds the In Review timing, the join confirmation, the
join echo, Labels and the follow-up pointer. `integrate.md` loads it beside the core. Labels
citers repoint: run 1, `jira-followups.md`, `/flow-plan` and `/flow-self-review`. The core keeps
"exactly two carve-outs" and cites the moved one.

### D6 — the carried `STARTED` block drops the `(run-only)` markers

**ID:** started-block-no-markers
**Status:** active
**Chosen:** the markers annotate `/flow-status`'s template and are never printed
(`handoff-blocks.md`). Every label, line and order is unchanged.

## Measured

`scripts/load-sets.sh`, before (`966d2573`) → after:
<!-- measured: scripts/load-sets.sh @ branch claude/brave-hamilton-533aad -->

| Session | Before | After |
|---|---|---|
| Planning, definite | 128,570 B | 108,290 B |
| Planning, incl. cited | 163,810 B | 127,576 B |
| Implementation, definite | 293,095 B | 278,358 B |
| Finish, definite | 183,510 B | 182,445 B |

Planning's conditional files (`withdrawal.md`, `seeded-note.md`, `resume.md`) total 10,289 B.

## Open questions
