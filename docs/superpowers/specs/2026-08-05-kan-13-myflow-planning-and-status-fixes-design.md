# KAN-13 — myflow planning-rule and status-enumeration fixes

**Date:** 2026-08-05
**Jira:** KAN-13 (agents-repo scope only — items 2 and 3 of that issue belong to `gymie` and are
out of scope here)

## Why

Two independent defects surfaced during the KAN-9 frontend run, both in myflow itself:

1. `/myflow-start` produced a plan whose task order left `:shared` uncompilable from Task 2 until
   Task 9 — tests couldn't run for seven consecutive tasks, two tasks had to be merged mid-flight
   to reach a green state, and the plan needed a human-adjudicated repair mid-run.
2. `/myflow-status` (and, by the same code path, every other `/myflow-*` command's change-name
   resolution) enumerates open changes from `openspec list --json` alone. That command only sees
   change directories present in the checkout it runs in, so a change whose `openspec/changes/`
   directory lives only in a worktree — staged but never committed to the main checkout — is
   silently invisible, even while it sits at the human gate with a fully staged diff.

## What

### 1. Build-green task ordering

- New contract file `skills/myflow-contracts/build-green.md`, canonical for the `**Build:**` tag
  format and the guard's rule and scope, added to the `myflow-contracts` index in that directory's
  own `SKILL.md`.
- `/myflow-start` section D (writing-plans) requires every task in `tasks.md` to carry:
  - `**Build:** green` — the project builds and this task's own verification command can run,
    given only the preceding tasks, or
  - `**Build:** red — merges with Task <N>` — this task alone does not leave a green build, and is
    dispatched together with the named task(s) as one unit, whose combined result must be green.
- New guard `scripts/check-task-build-green.py` (stdlib only, thin `scripts/check-task-build-green.sh`
  wrapper — following the `check-plan-provenance.py`/`.sh` split) parses `tasks.md` and fails on:
  - a task with no `**Build:**` tag;
  - a task tagged `red` with no named merge partner;
  - a named merge partner that doesn't exist in the plan, or isn't itself tagged `green`;
  - a `red` chain that never resolves to `green` before the plan's last task.
- Run at the same point `check-plan-provenance` runs today: before publishing the proposal
  artifact, "run the project's configured build-green guard if it declares one, and fix any hit" —
  same optional, auto-detected-by-presence pattern the provenance guard already uses (no
  `.myflow/project.md` declaration key; the script's presence in `scripts/` is the declaration).
- Registered in this repo's `.myflow/project.md`: `## lint` gains
  `scripts/check-task-build-green.sh`; `## test` gains `scripts/test-check-task-build-green.sh`.
- New test `scripts/test-check-task-build-green.sh`, fixture-driven like
  `test-check-plan-provenance.sh`, covering: missing tag, red without partner, red with a
  nonexistent partner, red with a red partner, a red chain that resolves to green, and an
  all-green plan.

**What this does not do.** The guard cannot verify the code actually compiles — no script can, for
an arbitrary project in an arbitrary language. It verifies that the plan *declares* a build state
for every task and never leaves a declared-red task unresolved, exactly as `plan-provenance.md`'s
guard verifies that a claim is *stated*, never that it is true. The obligation to tag honestly
falls on whoever writes the plan, same as the existing `verified:`/`unverified:` tags.

### 2. Change-name-resolution enumeration

- Fix lands in the **one shared** `## Change name resolution` section of
  `skills/myflow-contracts/pipeline.md` — used by `/myflow-start`, `/myflow-do`, `/myflow-finish`
  and `/myflow-status` — not duplicated per command.
- Step 1 becomes: union the names from `openspec list --json` with the basenames (minus `.json`)
  of every file in the project's state directory
  (`/Users/tweety53/Agents/myflow/state/<project-key>/*.json`, `<project-key>` per the existing
  formula in `state-file.md`), then drop any name whose `openspec/changes/<name>/` has reached
  `archive/`.
- A state-directory file that can't be read (corrupt/unreadable) is reported and skipped from the
  union — the change's own status ends up `⚠`-marked via self-heal as today, but it is never
  silently dropped from the *listing* the way it was invisible before this fix.
- `skills/myflow-status/SKILL.md` step 1 is updated to cite this shared section rather than
  running `openspec list --json` on its own.

**No new script.** This is prose the agent follows, not code — its verification is behavioral, via
the manual test guide `/myflow-do` produces (create a change whose openspec directory exists only
in a worktree, confirm `/myflow-status` now lists it), not a unit test.

## Decisions

### Bundle both agents-repo items into one change

**ID:** bundle-scope
**Status:** active
**Chosen:** one combined change (`kan-13-myflow-planning-and-status-fixes`) — both are myflow
pipeline fixes discovered together during the same KAN-9 run, and bundling avoids two proposals
sharing one Jira issue with near-identical context.
**Considered:** two separate changes (mirrors how KAN-9 split by repo, but these two items share a
repo and neither depends on the other, so splitting by fix rather than by repo added review
overhead for no isolation benefit); scoping this run to the planning rule alone and deferring the
status fix (rejected — the status fix is small and already fully specified by the Jira issue).

### Build-green check is a structural tag, not a semantic build check

**ID:** build-green-mechanism
**Status:** active
**Chosen:** a required per-task `**Build:**` tag with a merge-partner rule, checked structurally by
a script that parses `tasks.md` — mirrors the existing `verified:`/`unverified:` provenance-tag
pattern, generalizes to any reason a task is red (not just deletions), and costs one line per task.
**Considered:** a file-touch-graph heuristic (`Creates:`/`Modifies:`/`Deletes:` per task, flagging a
verification command whose target is deleted by a later task) — catches the literal KAN-9 shape
automatically, but only that shape, and needs a new required field on every task just for this one
check; a fully automated semantic/AST-based compileability check — rejected as infeasible in
general (myflow plans span arbitrary languages and projects) and as over-engineering for what the
postmortem actually needed.

### Fix the shared resolution section, not just `/myflow-status`

**ID:** enumeration-fix-scope
**Status:** active
**Chosen:** fix `pipeline.md`'s one shared `## Change name resolution` section — KAN-13 itself
flags that `/myflow-finish`'s resolution has the identical blind spot, and it uses the same code
path, so fixing the shared section fixes every command at once and prevents the two from drifting.
**Considered:** fixing only `/myflow-status/SKILL.md`'s step 1 — narrower, faster, but leaves
`/myflow-do` and `/myflow-finish` silently exposed to the exact bug this change exists to close.

## Open questions

None — the one loose end (whether a new test script was needed for the enumeration fix) resolved
during design: that resolution logic is prose the agent follows, not code, so there is no existing
or new unit test to extend; it is covered by the manual test guide instead.
