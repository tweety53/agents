# myflow-fast — design

**Change:** `kan-111-myflow-fast`
**Jira:** KAN-111 — "myflow-fast"
**Date:** 2026-08-09
**Planning effort:** default

## Problem

The `/myflow-*` pipeline has three human gates: reading the proposal artifact after `/myflow-start`,
reviewing the staged diff and running the apps after `/myflow-do`, and choosing how to land the
branch during `/myflow-finish` run 1. Between `/myflow-do` ending and `/myflow-finish` run 1
starting, and again between run 1 and run 2, the operator has to notice the state and type the next
command by hand even when nothing blocks moving forward.

For a change where the operator wants to stay hands-off except for the parts only a human can do —
answering brainstorming's design questions, reviewing the diff, testing the app, and choosing the
integration route — the three separate commands add friction with no compensating benefit.

## Approach

Add `/myflow-fast` as a driver over the same three-state machine (`STARTED` → `IN_PROGRESS` →
`FINISHED`), same state file, same Jira integration, same worktree/panel/finish machinery. It does
not duplicate pipeline logic — it chains the existing stage content and collapses the *inter-stage*
handoffs, while leaving every *human* gate exactly where it already is.

| Stage | What runs | Ends at |
|-------|-----------|---------|
| 1. start | `/myflow-start`'s brainstorming, **unchanged** — full interactive dialogue, one question at a time | `STARTED` |
| 2. do | `/myflow-do`'s SDD+TDD implementation and review panel, **unchanged**, fully unattended | `IN_PROGRESS` |
| *(human gate)* | operator reviews the staged diff, runs the app | — |
| 3. finish | `/myflow-finish`'s integrate + archive, chained together where nothing external blocks it | `IN_PROGRESS` (PR) or `FINISHED` |

Stage 1 is not a lighter or auto-answered version of `/myflow-start` — it is the same brainstorming
flow, verbatim. The only thing `/myflow-fast` changes about it is skipping the proposal artifact
publish step, because the operator was present for the conversation that produced the design.

## Invocation

`/myflow-fast <description>` or `/myflow-fast <JIRA-KEY>` — first call, no state file for the
change yet. `<JIRA-KEY>` pulls the issue's summary/description as the brainstorming seed, same as
`jira-integration.md` already resolves a linked issue elsewhere in the pipeline. A free-text
description seeds the brainstorming conversation directly.

Bare `/myflow-fast` (no argument) resolves the sole active change via **Change name resolution**
(`skills/myflow-contracts/pipeline.md`) — same rule every other `myflow-*` command already uses.

## Recorded choices

Made once, on the run that creates the change (stage 1), exactly as `/myflow-start` already does for
`planningEffort`, `models` and `reviewPanelRoster`:

- `models`: `{implementation: sonnet, reviewPanel: sonnet, panelFix: sonnet}` — recorded default for
  changes created via `/myflow-fast`, overridable the same way any recorded model choice already is.
- `reviewPanelRoster`: `light` — already the project-wide default (KAN-110); `/myflow-fast` does not
  change what `light` means, only which command reaches it without asking.
- `planningEffort`: same question `/myflow-start` already asks, defaulting to `low`.

None of these are new fields or new semantics — `/myflow-fast` just supplies recommended answers to
questions `/myflow-start` already asks, on the same recorded schema.

## Re-invoking at IN_PROGRESS

Two distinct intents share one state, so the argument disambiguates:

- **`/myflow-fast <fix instructions>`** — treated as a fix, resuming the worktree exactly as
  re-running `/myflow-do` at `IN_PROGRESS` does today, with the text as the fix's guidance. State is
  written back unchanged, per **A fix never moves the state** (`pipeline.md`).
- **`/myflow-fast` (bare)** — proceeds to `/myflow-finish` run 1's integrate question: open PR,
  merge+push, or manual.

## The integrate step

Runs the preflight (`RUN1`/`RUN2`/`REFUSE`) exactly as `/myflow-finish` does. On `RUN1`, after the
route question:

- **open PR** — opens the PR, prints its URL, stops at `IN_PROGRESS`. A merge happens outside this
  tool's control, so nothing can auto-continue past it. The next bare `/myflow-fast` call, once
  merged, runs `RUN2`.
- **merge+push** — merges and pushes, then *continues in the same invocation* into `RUN2`: archive,
  delta-spec sync, worktree/branch removal, Jira transition — ending at `FINISHED` with no further
  command needed.
- **manual** — hands off as `/myflow-finish` does today; the next bare `/myflow-fast` call, once
  merged, runs `RUN2`.

`RUN2` on its own (reached directly, e.g. a change already merged when `/myflow-fast` is next
invoked) behaves exactly as `/myflow-finish` run 2 does today — no chaining decision to make, since
there is nothing after it but `FINISHED`.

## What does not change

Every contract `/myflow-fast` touches is cited, not restated: git boundaries, the handoff shape, the
tab-naming commands, session-record preservation, the temporary-artifacts registry, the wrong-state
override prompt, and the finish contract's preflight/base-branch/cleanup procedures all apply
verbatim. `/myflow-fast` adds no new rule to any of them — it is a sequencing layer that decides
*when* to call into stage 1/2/3 content and *whether* to keep going past a stage boundary, never
*what* a stage does.

The review panel's handoff bar is unchanged: `/myflow-do`'s stage still hands off only at zero open
findings at any severity.

## Files touched

- `skills/myflow-fast/SKILL.md` — new skill, stages as above, citing into `myflow-start`,
  `myflow-do`, `myflow-finish` and the contracts rather than re-deriving their content.
- `commands-claude/myflow-fast.md` (+ Cursor/Codex command variants, per the existing per-harness
  pattern) — frontmatter `model: sonnet`.
- `CLAUDE.md` — command table gets a `/myflow-fast` row; skill index gets a `myflow-fast` entry.
- `skills/myflow-contracts/pipeline.md` — command surface table gets a `/myflow-fast` row.

## Decisions

### Brainstorming stays interactive, unchanged

**ID:** `brainstorming-untouched`
**Status:** active
**Chosen:** Stage 1 runs `/myflow-start`'s full interactive dialogue verbatim — the operator answers
every question, same as today. `/myflow-fast` only removes the proposal-artifact publish step and
supplies recommended defaults for the planning-effort/model/roster questions that stage already
asks.
**Considered:** Auto-answering the design questions from the model's own judgment — rejected after
the operator confirmed they want to answer brainstorming questions themselves, same as full myflow.

### Fix vs. integrate disambiguation

**ID:** `arg-present-means-fix`
**Status:** active
**Chosen:** At `IN_PROGRESS`, an argument means fix instructions (resumes `/myflow-do`-style);
bare invocation means proceed to the integrate question. No extra confirmation prompt.
**Considered:** Always asking "fix or integrate?" explicitly — unambiguous, but adds a manual step
on every re-invocation, working against the "one command, one answer" goal.

### Chaining merge+push straight to archive

**ID:** `merge-push-auto-archives`
**Status:** active
**Chosen:** Once the operator picks merge+push, nothing external blocks proceeding — so
`/myflow-fast` runs `RUN1` then `RUN2` in the same invocation, ending at `FINISHED`.
**Considered:** Stopping after merge+push like a normal `/myflow-finish` run 1 — mirrors the
existing two-run contract exactly, but leaves a manual second invocation with nothing actually
gating it, which is exactly the friction `/myflow-fast` exists to remove.

### PR route stops and hands off

**ID:** `pr-route-stops`
**Status:** active
**Chosen:** Opening a PR still requires an external human merge (on the forge, possibly by someone
else), so `/myflow-fast` cannot auto-continue past it. It prints the PR URL and stops.
**Considered:** Polling until merged, then auto-archiving — would tie up a long-running session or
need a background job for an event this pipeline has no mechanism to observe; rejected as
disproportionate to the problem.

### Model defaults: Sonnet everywhere

**ID:** `sonnet-everywhere`
**Status:** active
**Chosen:** `/myflow-fast`-created changes record `implementation`, `reviewPanel` and `panelFix` all
as Sonnet, deviating from the normal pipeline's Opus-for-implementer/panel-fix default.
**Considered:** Keeping the normal pipeline's Opus/Sonnet/Opus split — rejected in favor of lower
cost/latency per run, since the operator explicitly chose speed over the normal default here.

### Panel roster: light

**ID:** `light-roster-default`
**Status:** active
**Chosen:** `reviewPanelRoster: light`, matching the project-wide default already set by KAN-110.
**Considered:** `standard` — broader coverage, slower fix-cycles; rejected to keep `/myflow-fast`
at the fastest already-supported preset rather than introduce a fourth roster just for this command.

### No proposal artifact

**ID:** `no-proposal-artifact`
**Status:** active
**Chosen:** Stage 1 skips publishing the proposal HTML artifact. The operator participated in the
brainstorming dialogue directly, so a separate published summary of that same conversation adds a
step without adding information.
**Considered:** Publishing it anyway for parity with `/myflow-start` — rejected as pure overhead for
a command whose whole purpose is cutting steps that don't carry new information.

## Open questions

*(none — every question raised during planning was answered)*
