## Context

- Today, three texts forbid dispatch during planning:
  - `skills/flow/brainstorm.md`: "**No dispatch of any kind from kickoff to the plan gate — a
    `fork` included.**" The recorded reason is about forks: a fork copies the whole conversation, so
    it saves no context and doubles the spend.
  - `skills/flow/brainstorm-planner.md` (opening paragraph): "no dispatched subagent, no relay".
  - `skills/flow-plan/SKILL.md`: "No subagent is dispatched" and "No `flow record dispatch` call".

  A fresh Explore or Plan agent is not a fork. It starts empty and returns only its conclusions, so
  it saves the planning session's context, which is the opposite of the failure the ban was written
  against.
- The self-review planner and the in-run fix planner run sections B–D as a `flow-medium` subagent
  (`skills/flow-self-review/SKILL.md`, the pairs table). `agents/flow-*.md` omit `Agent` from
  `tools:` on purpose (NO DELEGATION, `skills/flow/SKILL-rationale.md`), and the plan dispatch
  carries the NO DELEGATION paragraph verbatim.
- The reproducible-roll precedent is `scripts/plan-class.sh`'s `extra:` lines:
  `sha256("<name>extra") mod 100` decides `sonnet` or `opus` at `low`.
- Claude Code plan mode: `EnterPlanMode` asks the user to consent, then the harness keeps the
  session read-only. `ExitPlanMode` shows the plan file the harness named, and the user approves or
  rejects it with feedback. Only Claude Code has these tools. ZCode and Muse runs keep today's
  behaviour.

## 1. Subagents in planning

All three entry points (`/flow`'s creating run, `/flow-plan`, and the self-review / in-run fix
planners) may dispatch fresh subagents through the Agent tool while planning. A `fork` stays banned
for the recorded reason. The thinking itself stays in the planning session: questions, convergence,
the choice between designs, the artifacts, Decide and the gates. Agents investigate and propose,
and the session decides.

Every planning dispatch:

- picks its pair with `plan-agent-pair.sh` (§5);
- carries the two agent-baseline sentences and the brief-tersely line;
- states a narrow, explicit scope (what to look at, what is already known, what to return).

`/flow` records each one with `flow record dispatch` under stage `flow.brainstorm`: Explore agents
with role `explorer` (new), Plan agents with role `planner`. `/flow-plan` records none, because no
change exists before capture. The self-review planner's nested dispatches are not recorded either:
they carry no session token of their own.

## 2. Parallel Explore phase

Before the first design question (after the withdrawal reachability check, when that runs), the
session dispatches 1–3 Explore agents (subagent type `Explore`) in one message, so they run
concurrently. Each takes a different area: the code the change touches, existing patterns and
utilities to reuse, and the tests and guards around it. Use 1 when the change is confined to one
known area, and more when its scope is uncertain or spans areas. The session then reads the
critical files the agents name before asking anything. The agents' conclusions feed the questions;
they never replace asking.

## 3. Competing designs

For a change that is not trivial (more than one valid approach, a cross-cutting change, or an
architectural choice), the session dispatches 1–3 Plan agents (subagent type `Plan`) in one
message, each briefed with the Explore conclusions and its own angle: simplest, most robust, best
fit with existing patterns. The session compares their proposals and recommends one. Each
alternative it rejects becomes the `**Considered:**` line of the matching `## Decisions` entry,
with the tradeoff that ruled it out. A change with one obvious approach skips this step and records
why in `## Recommended approach`.

In `/flow` and `/flow-plan` the recommendation goes to the operator as part of the design (its
alternatives shown only when they help the operator decide). The self-review and in-run fix
planners take the recommendation without asking, per **Auto-resolution**
(`skills/flow-contracts/operator-prompts-auto-resolution.md`).

## 4. Plan sections in `design.md`

Section **C** writes four more `##` sections, each kept even when short:

- `## Recommended approach`: the one chosen approach, in prose, with a one-line reason for each
  rejected alternative pointing at its decision ID.
- `## Critical files`: every file the change edits or depends on, by repository-relative path, one
  line on its role each.
- `## Existing code to reuse`: functions, scripts, guards and patterns the change builds on, by
  `path` (and symbol). `none` is a valid entry only with a one-line reason.
- `## Verification`: how to prove the change works end to end. Commands, the run or scenario to
  exercise, and what is observed. It goes beyond each task's own verify step.

The planner writes `> **Design shape:** plan-mode` into `tasks.md`'s header block. On a `tasks.md`
carrying that line, `check-plan-shape.sh` refuses a sibling `design.md` missing any of the four
headings, one line each. Without the line, the check skips it. This keeps the ~25 older live changes
under `spectre/changes/` passing, and every change planned after this lands carries the line,
because section C requires it.

## 5. The pair roll

`scripts/plan-agent-pair.sh <seed> <n>` prints `sonnet low` when `sha256("<seed>plan<n>") mod 100 <
40`, else `opus low`, using the same digest code as `plan-class.sh`. `<seed>` is the change name; in
`/flow-plan` before capture, it is the session token. `<n>` counts the session's planning
dispatches from 1. It is symlinked into `skills/flow/scripts/` and `skills/flow-plan/scripts/` per
**Guard resolution**. The dispatch passes the printed model as `model` and the effort as the Agent
tool's `effort`.

## 6. Real plan mode on Claude Code

When `<harness>` is `claude-code`:

- **`/flow`'s creating run**: after `flow.kickoff` closes and `flow.brainstorm` begins (`STARTED`
  and the worktree already exist), the session calls `EnterPlanMode` before the Explore phase.
  Investigation, questions and competing designs run inside plan mode. Once convergence closes, the
  session writes the design to the plan file the harness names and calls `ExitPlanMode`.
  - **Approve** is the design approval of the HARD GATE: the session then saves the design under
    `.superpowers/sdd/` and continues into section C.
  - **Reject with feedback** reopens convergence with the feedback as the next round's input.
  - On Claude Code the merged convergence confirm closes the checklist; `ExitPlanMode` grants the
    approval. Every other gate is unchanged, the plan review gate included.
- **`/flow-plan`**: the session calls `EnterPlanMode` at session start. It calls `ExitPlanMode`
  only once it reaches a capturable design, as the design approval before the capture question; the
  capture question is still asked after approval. A research session that ends without a capture
  never calls `ExitPlanMode`.
- **Planners**: never. They are subagents with no operator.

A declined `EnterPlanMode` consent is one line and planning continues without plan mode, exactly as
on another harness.

## 7. Planner agent definitions

`agents/flow-planner-low.md`, `agents/flow-planner-medium.md` and `agents/flow-planner-high.md` are copies of
`flow-low.md`, `flow-medium.md` and `flow-high.md` with `Agent` added to `tools:`. In the pairs table, the rows `plan`, `plan-<n>`
admit `flow-planner-medium` and `flow-planner-low` (medium by default; low when every finding the plan
carries is mechanical, with a one-line reason the session records), with `flow-planner-high` on a
recorded escalation. The plan dispatch
carries a **PLANNING DELEGATION** paragraph in place of NO DELEGATION. It may dispatch fresh
Explore and Plan agents per §1–§3 and §5, never a fork, and never an agent that edits files.
Fixers, reviewers and implementers keep NO DELEGATION and their definitions.
`check-dispatch-paragraphs.sh`'s table gains the new paragraph's row.
`check-loop-dispatch.sh` reads the table, so it admits the new types without a code change.

## Recommended approach

Extend each of the three entry points in place, in their own texts. §1–§3 are a new subsection of
**B. The checklist** in `brainstorm-planner.md`, which all three already run. `/flow` and
`/flow-plan` cite it for their own plan-mode wrapper (§6). The planners reach it through their
existing B–D instruction, plus §7's agent definitions. Rejected: a separate "plan mode" skill that
replaces brainstorming (decision `extend-not-replace`), and the parent session running the Explore
and Plan agents for the planners (decision `planner-carries-agent`).

## Critical files

- `skills/flow/brainstorm.md`: the dispatch ban, and the plan-mode wrapper for the creating run.
- `skills/flow/brainstorm-planner.md`: the B checklist subsection, C's four sections and the
  design-shape line.
- `skills/flow-plan/SKILL.md`: the dispatch statements, and the plan-mode wrapper.
- `skills/flow-self-review/SKILL.md`: the pairs table and the planner dispatch paragraph.
- `skills/flow/SKILL-rationale.md`: the reason for the narrowed ban.
- `scripts/plan-agent-pair.sh`: new.
- `agents/flow-planner-low.md`, `agents/flow-planner-medium.md`, `agents/flow-planner-high.md`: new.
- `stats/internal/guard/planshape.go`: the design-section check.
- `stats/cmd/flow/record.go`: the `explorer` role.
- `scripts/check-dispatch-paragraphs.sh`: the PLANNING DELEGATION row.

## Existing code to reuse

- `scripts/plan-class.sh`: the `sha256 mod 100` roll and its digest code, for the pair roll.
- `stats/internal/guard/loop.go` `lpTable`: it already parses the pairs table, so no change is
  needed for new types.
- `agents/flow-low.md`, `agents/flow-medium.md`, `agents/flow-high.md`: copied for the planner definitions.
- `stats/internal/guard/planshape.go`: the existing header-block parsing of `tasks.md`.

## Verification

- `scripts/plan-agent-pair.sh demo 1` … `demo 20` prints only `sonnet low` / `opus low`, the same
  output on a repeat run, and roughly 40/60 over 100 seeds (the test pins exact values).
- `scripts/check-plan-shape.sh` passes over every live change, and fails on a fixture carrying
  the design-shape line without `## Verification`.
- `scripts/check-loop-dispatch.sh` admits a `self-review-x-plan` dispatch on `flow-planner-medium`
  and refuses one on `flow-medium`.
- `flow record dispatch begin -role explorer …` is accepted; `flow record roles` lists `explorer`.
- End to end: the next `/flow <KEY>` creating run on Claude Code asks for plan-mode consent,
  dispatches Explore agents in parallel, shows `ExitPlanMode`, and writes a `design.md` carrying
  the four sections. Its decision ledger shows the planning dispatches with role `explorer` /
  `planner` on `sonnet`/`opus` at `low`.

## Step-by-step breakdown

### Planning dispatch allowance

**What:** narrow the dispatch ban to forks in `/flow`, `/flow-plan` and the B–D preamble, keeping
the thinking in the planning session.
**Why:** the ban's recorded reason covers forks only; fresh agents save the context the ban was
protecting.
**Uses:** `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`,
`skills/flow-plan/SKILL.md`, `skills/flow/SKILL-rationale.md`.

### Explore phase and competing designs

**What:** a checklist subsection: 1–3 parallel Explore agents before the first question; 1–3 Plan
agents for a non-trivial change; one recommendation, with the rejected alternatives recorded.
**Why:** plan mode's two investigation and design practices flow planning lacks.
**Uses:** `skills/flow/brainstorm-planner.md`, the `Explore` and `Plan` subagent types.

### Pair roll

**What:** `plan-agent-pair.sh` prints `sonnet low` (40%) or `opus low` (60%), reproducibly per
seed and dispatch number.
**Why:** the operator's split, made reproducible like the extra-dispatch roll.
**Uses:** `scripts/plan-class.sh`'s digest.

### Plan sections and their check

**What:** four new `design.md` sections, the `Design shape` header line, and the
`check-plan-shape` rule.
**Why:** plan mode's plan always names its critical files, reuse and end-to-end verification.
**Uses:** `stats/internal/guard/planshape.go`.

### Plan mode wrapper

**What:** `EnterPlanMode` and `ExitPlanMode` around investigation and design in `/flow` and
`/flow-plan` on Claude Code.
**Why:** harness-enforced read-only, plus an approve-or-reject-with-feedback design gate.
**Uses:** the `EnterPlanMode`/`ExitPlanMode` tools, `<harness>`.

### Planner definitions

**What:** `flow-planner-low`/`-medium`/`-high` with `Agent`, the pairs-table rows, the PLANNING
DELEGATION paragraph.
**Why:** the self-review and in-run fix planners spawn their own Explore and Plan agents while
every leaf role keeps no-delegation.
**Uses:** `agents/`, `skills/flow-self-review/SKILL.md`, `scripts/check-dispatch-paragraphs.sh`,
`stats/internal/guard/loop.go`.

### Dispatch recording

**What:** the `explorer` role, and `/flow` recording its planning dispatches.
**Why:** planning spend shows in the ledger like any other dispatch.
**Uses:** `stats/cmd/flow/record.go`.

## Decisions

### extend-not-replace

**ID:** extend-not-replace
**Status:** active
**Chosen:** flow planning keeps every current practice and gate and adds the plan-mode practices
(operator's answer: "current ideas + plan-mode ideas").
**Considered:** replacing brainstorming with plan mode — rejected by the operator.

### all-four-practices

**ID:** all-four-practices
**Status:** active
**Chosen:** parallel Explore agents, competing designs, plan sections and real plan mode — all four
(operator's answer).
**Considered:** a subset — the operator chose all four.

### three-entry-points

**ID:** three-entry-points
**Status:** active
**Chosen:** `/flow`'s creating run, `/flow-plan`, and the self-review and in-run fix planners; the
planners auto-pick (operator's answers).
**Considered:** `/flow-plan` only — rejected by the operator.

### agent-tool-everywhere

**ID:** agent-tool-everywhere
**Status:** active
**Chosen:** every planning entry point may spawn subagents through the Agent tool; forks stay
banned (operator's instruction; the fork ban keeps its recorded reason).
**Considered:** keeping the ban for every dispatch — rejected because fresh agents save the context
the ban protected.

### low-effort-40-60

**ID:** low-effort-40-60
**Status:** active
**Chosen:** Explore and Plan agents run at `low`, `sonnet` 40% / `opus` 60%, rolled per dispatch
by `plan-agent-pair.sh` (operator's split; the reproducible roll is my call, after `plan-class.sh`).
**Considered:** a live random roll — rejected because a run could not be reproduced or audited.

### planner-carries-agent

**ID:** planner-carries-agent
**Status:** active
**Chosen:** new `flow-planner-low`/`-medium`/`-high` definitions with `Agent`, used by the `plan` rows only
(my call for `Agent`; the low/medium/high family is the operator's request, mirroring `flow-<effort>`).
**Considered:** adding `Agent` to `flow-medium` — rejected because it would remove no-delegation
from every medium leaf role; the parent dispatching on the planner's behalf — rejected because the
operator asked for subagents in the planner itself, and the parent cannot know what the planner
will want explored.

### design-shape-line

**ID:** design-shape-line
**Status:** active
**Chosen:** the four sections are enforced on changes whose `tasks.md` carries
`> **Design shape:** plan-mode` (my call).
**Considered:** enforcing on every live change — rejected because ~25 live changes planned before
this would fail `## lint`; no check — rejected because it would be an unenforced rule.

### exit-plan-mode-is-approval

**ID:** exit-plan-mode-is-approval
**Status:** active
**Chosen:** on Claude Code, `ExitPlanMode` is the design approval; the convergence confirm still
closes the checklist; the plan review gate stays (my call).
**Considered:** ExitPlanMode in addition to the merged confirm's approval — rejected as two
approvals of one design; ExitPlanMode replacing the plan review gate — rejected because that gate
reviews `tasks.md` and the decision, which do not exist yet at design time.

### flow-plan-records-nothing

**ID:** flow-plan-records-nothing
**Status:** active
**Chosen:** `/flow` records its planning dispatches (roles `explorer`, `planner`); `/flow-plan`
and the planners' nested dispatches record none (my call).
**Considered:** recording in `/flow-plan` — rejected because no change exists to record against
before capture.

## Open questions
