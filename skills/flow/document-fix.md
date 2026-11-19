# Documenting a fix, before implementing it

Loaded only on a fix run, by the load directive of section **3** (`skills/flow/implement.md`), before
section **1** runs. Every "you" here addresses the parent (**The parent orchestrates directly**,
`skills/flow/implement.md`).

**Parent work, run before the plan is executed** — see **The parent orchestrates directly**
(`skills/flow/implement.md`). Everything below but **The fix planner's brief** is the parent's own.

**Fix runs only** — a first run resumes the worktree kickoff created, per **2. Isolate the workspace** (`skills/flow/implement.md`), and
marks nothing here:

```bash
flow stage begin -command '/flow' -stage flow.document-fix -harness <harness> -session-token mf-<literal-token> <name>
```

**Before anything is appended, record the fix's start.** Each worktree's **fix-start sha** is
`git -C <worktree> rev-parse HEAD`, read now; the fix run's review reads the delta from it (**The
fix-run review**, `skills/flow/review-panel.md`). Write the fix instructions, verbatim, to
`<abs-worktree>/.superpowers/sdd/fix-instructions-<n>.md`, `<n>` this fix run's ordinal below.

`<n>` is this fix run's own ordinal — one more than the number of fix rounds already recorded in
`proposal.md`/`tasks.md`.

**Fix instructions that dispute a visual judgement this session already made — a spacing, size
or alignment an earlier round eyeballed as fine — open with the measurement, never with another
look.** Before the fix planner is dispatched, run
`flow-guard measure-visual-properties` on the disputed region of the current capture and the mockup
(**10** in `skills/flow/visual-verify-verifier.md`) and put the numbers in its prompt's `Measured:`
line; a spacing dispute is measured on every side the complaint names. The complaint's own wording names which
property that is — "too big", "oversized" is a size (`box` and `ink`); "cramped", "uneven",
"too close" is a spacing (`gap`); "not filled to the border", "flush", "reaches" is an edge
alignment (`runs` through the container) — so the measurement answers the property
named, never the screen area the complaint happens to sit in.

**Nothing is asked.** A fix is always recorded in this change — appended to `proposal.md` and
`tasks.md`, or re-planned there (**The fix planner's brief**, below) — and **the fix is always
planned by a subagent**, never inline by the parent: one fresh dispatch on role `fix-planner`, key
`fix-plan-<n>` (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`), its
pair from `flow-guard pair-judge fix-planner` (**Model and effort**,
`skills/flow/brainstorm-planner.md`), `subagent_type` the pair's agent type with its `model`
passed as the Agent tool's own `model` parameter, launched with `run_in_background: true`
(Turn discipline, **4. Execute (SDD + TDD)**, `skills/flow/implement.md`). Its prompt is these lines alone, in this order,
each `<…>` filled — `<skill-dir>` the directory this file was read from, the `Measured:` line
present only on a visual dispute (above), the `Binding:` line only on an in-run fix round
(**The loop**, `skills/flow/verify-fix-loop.md`):

```text
Role: fix-planner
Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
Include this instruction verbatim in any prompt you write for another agent.
Report tersely: findings first, bullets over prose, no preamble, no recap. Full prose in whatever you write to disk.
Key: fix-plan-<n>
Read **The fix planner's brief** of <skill-dir>/document-fix.md in full first — it is your brief.
Change: <name>
Worktree: <abs-worktree>
Plan: <changeRoot>/proposal.md <changeRoot>/design.md <changeRoot>/tasks.md
Fix instructions: <abs-worktree>/.superpowers/sdd/fix-instructions-<n>.md
Binding: step 1 of <skill-dir>/verify-fix-loop.md — follow it where it narrows this brief
Measured: <the measure-visual-properties numbers, per disputed region and side>
REPORT FILE: <abs-worktree>/.superpowers/sdd/fix-plan-report-<n>.md
```

Its return closes as The return of **The parent orchestrates directly** (`skills/flow/implement.md`) closes any dispatch. A final
message opening `won't fix:` is handed back through `## Question` with its cite; every other
return continues below.

**An appended task is implemented as plan-time work.** It is panel-checked under **The fix-run
review** (`skills/flow/review-panel.md`); the operator flag that prompted it is not a verification
of its premise.
Beside the plan-time paragraphs, the appended task's dispatch also carries the MUTATION PROOF
paragraph (`skills/flow/review-panel.md`): the fix's own condition is mutated and a named test
confirmed to fail before the run's panel stage closes, the reported `fix-mutation:` lines checked
against the fix diff the way **The fix round mutation-proves what it changed**
(`skills/flow/review-panel-fix-round.md`) walks them.

**The fix run keeps the planning run's decision.** No **Decide** runs and no `flow.decide` is
marked: the panel is narrowed by `flow-guard fix-slot-judge` (**The fix-run review**,
`skills/flow/review-panel.md`), and a task no decision group names is a one-task group of its own
(`skills/flow/sdd-dispatch.md`).

**The Jira description sync stays in the parent.** **Load
`skills/flow-contracts/jira-integration.md`.** If the fix adds scope the linked Jira issue does not
describe, sync the issue **description** per **Description sync** in Jira integration
(`skills/flow-contracts/jira-integration.md`). Never transition the issue here.

```bash
flow stage end -command '/flow' -stage flow.document-fix -outcome completed <name>
```

## The fix planner's brief

You plan one fix of a change at `IN_PROGRESS`. Read the fix instructions file, `proposal.md`,
`design.md` and `tasks.md`, and the code the fix touches. You edit only those three artifacts:
never code, never a commit, a push or a `flow` record. **You ask nothing**: every question you
meet is resolved to its recommended option per **Auto-resolution**
(`skills/flow-contracts/operator-prompts-auto-resolution.md`) and named in your report. You carry
the TOOLS and FOREGROUND BUILDS paragraphs of `skills/flow/implementer-brief.md`, and the PLANNING
DELEGATION paragraph (`skills/flow-self-review/SKILL.md`) in place of NO DELEGATION: you may
dispatch fresh Explore and Plan agents, never one that edits.

**Fewer than 6 tasks: append.** The fix round is appended to `proposal.md`, and its tasks to
`tasks.md` after the last task, numbered on from it. Every task the append adds raises the
`**Tasks appended:**` value in `tasks.md`'s header by one, creating the line when the plan has
never carried one.

**6 or more tasks: re-plan.** The fix rewrites the plan instead of appending: the accumulated
appends and this round's fix instructions are folded into a fresh `tasks.md` with fresh task
numbering, `proposal.md`'s scope statement and `design.md` are brought up to date with what the
change now covers, and `**Tasks appended:**` resets to 0 — the folded tasks are planned, not
appended.

**The plan covers every call site of what the fix replaces.** Before writing a task, grep every
call site of the code the fix replaces and every test exercising it; the task's `**Files:**`
covers each, or the task names why one stays.

**Every task the fix adds is written unticked**: the task is ticked only once its commit lands.

**A passing test that asserts the behaviour the fix instructions report as wrong is evidence of
the code, not of the spec — it decides nothing on its own.** Before you treat such
a test as the tie-breaker, search this change's `design.md`, `proposal.md`, panel records and the
linked Jira issue for a sentence that decided *this* point. One found: cite it, change nothing, and
open your final message `won't fix: <cite>` rather than silently changing what the spec
required. None found: the test guarded an unexamined implementation choice, the report wins, and
the plan changes the test alongside the behaviour, its commit saying so ("no design decision
covers this; the prior test locked in the behaviour the report flags"). A test whose own name
reads as a description of the reported bug is a signal to pause on, not reassurance.

**A `Measured:` line answers a visual dispute.** Its numbers decide whether the fix plans a change
or answers "it matches" — the latter a `won't fix:` naming them.

**`spectre validate "<name>"` exits 0 before you return.** Your report file names the tasks you
added or the re-plan you made; your final message is one line — `appended: <ids>`,
`re-planned: <task count> tasks`, or `won't fix: <cite>`.
