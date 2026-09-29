# Documenting a fix, before implementing it

Loaded only on a fix run, by the load directive of section **3** (`skills/flow/implement.md`), before
section **1** runs. Every "you" here addresses the parent (**The parent orchestrates directly**,
`skills/flow/implement.md`).

**Parent work, run before the plan is executed** — see **The parent orchestrates directly**
(`skills/flow/implement.md`). Everything below is the parent's own.

**Fix runs only** — a first run resumes the worktree `flow.kickoff` created, per **2. Isolate the workspace** (`skills/flow/implement.md`), and
marks nothing here:

```bash
flow stage begin -command '/flow' -stage flow.document-fix -harness <harness> -session-token mf-<literal-token> <name>
```

**Before the planning pass, the appended-task budget is checked.** Read the `**Tasks appended:** <n>`
line from the header of this change's `tasks.md` — the count of tasks appended at the human gate
since the plan was first written; a plan that has never carried the line reads as 0. When the
count has reached **6**, the re-plan budget, this fix round is offered the planning pass before
anything is appended. Ask the
operator, the shape **The shape** (`skills/flow-contracts/operator-prompts.md`) fixes:

> **This change's plan has had <n> tasks appended at the human gate — at the re-plan budget of
> 6. Re-plan instead of appending?**
> - **Re-plan** *(default, recommended)* — this fix's planning pass rewrites the plan instead of
>   appending: the accumulated appends and this round's fix instructions are folded into a fresh
>   `tasks.md` with fresh task numbering, `proposal.md`'s scope statement is brought up to date
>   with what the change now covers, and `**Tasks appended:**` resets to 0 — the folded tasks are
>   planned, not appended
> - **Append anyway** — the fix is appended exactly as this section otherwise states, and the
>   count keeps growing

Silence takes the recommended re-plan, and the ⚠ line names it. Under the
`## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`) the ask is not made: **Re-plan** is taken and
recorded the way the mode records a taken default. Either answer continues into the
planning pass below — the answer names its brief: on **Append anyway** the pass runs as this
section states it, its own where-should-it-go question included; on **Re-plan** the rewrite is the
brief and that question does not arise.

Record what changed **before** writing code, so the proposal never goes stale. `<n>` is this fix
run's own ordinal — one more than the number of fix rounds already recorded in `proposal.md`/
`tasks.md` or as `<name>-fix-N` sub-changes, the same `N` the "where should it go" prompt's
sub-change option below names. **This planning pass is the parent's own work**, run inline on this
session's model with the fix instructions in place of the design checklist — no dispatch, no
handshake, no relay.

The planning pass opens by asking where the fix should go, asked directly by the parent, shape per
Operator prompts (`skills/flow-contracts/operator-prompts.md`):

> **This fix has to be recorded before it is written — where should it go?**
> - **Append to `proposal.md` and `tasks.md`** *(default, recommended)* — nothing new is created
> - **Create a linked `<name>-fix-N` sub-change** — its own proposal and plan, for a fix that adds
>   scope the parent change does not describe

Under the mode the ask is not made: **Append to `proposal.md` and `tasks.md`** is taken and
recorded the way the mode records a taken default (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`).

The parent writes the append, or the sub-change's own proposal and plan. Whichever brief the
budget answer named, it keeps the counter true: every task its append adds raises the `**Tasks
appended:**` value by one, creating the line in `tasks.md`'s header when the plan has never
carried one.

**The appended task's verification tags are evidence-checked like a seeded note's.** An appended
task carries a `verified:`/`measured:` tag only when its evidence is in hand; an unverifiable one
is written `unverified:`/`predicted:` instead, never appended as a verification tag. The append
is where a fix round is most tempted to assert a check nobody made: `check-task-commit-fields.sh`
refuses the close of a task whose record carries the evidence-free shape, per **Plan
provenance**'s evidence rule (`skills/flow-contracts/plan-provenance.md`).

**An appended task is implemented and panel-checked exactly as plan-time work — the append never
narrows the panel.** The operator flag that prompted it is not a verification of its premise: its
work lands in the fix run's diff and takes the panel beside every other task's, and the narrow
late-fix path stays closed to an append (**The late-fix reduction**, `skills/flow/review-panel.md`).

**A passing test that asserts the behaviour the fix instructions report as wrong is evidence of
the code, not of the spec — it decides nothing on its own.** Before the planning pass treats such
a test as the tie-breaker, search this change's `design.md`, `proposal.md`, panel records and the
linked Jira issue for a sentence that decided *this* point. One found: cite it and hand the fix
back as "won't fix, per <cite>" through `## Question` rather than silently changing what the spec
required. None found: the test guarded an unexamined implementation choice, the report wins, and
the plan changes the test alongside the behaviour, its commit saying so ("no design decision
covers this; the prior test locked in the behaviour the report flags"). A test whose own name
reads as a description of the reported bug is a signal to pause on, not reassurance.

**Fix instructions that dispute a visual judgement this session already made — a spacing, size
or alignment an earlier round eyeballed as fine — open with the measurement, never with another
look.** Before the planning pass answers "it matches" or plans a fix, run
`measure-visual-properties.sh` on the disputed region of the current capture and the mockup
(**10** in `skills/flow/visual-verify-verifier.md`) and put the numbers in the plan or the
`## Question`; a spacing dispute is measured on every side the complaint names. The complaint's own wording names which
property that is — "too big", "oversized" is a size (`box` and `ink`); "cramped", "uneven",
"too close" is a spacing (`gap`); "not filled to the border", "flush", "reaches" is an edge
alignment (`runs` through the container) — so the measurement answers the property
named, never the screen area the complaint happens to sit in.

**The Jira description sync stays in the parent.** **Load
`skills/flow-contracts/jira-integration.md`.** If the fix adds scope the linked Jira issue does not
describe, sync the issue **description** per **Description sync** in Jira integration
(`skills/flow-contracts/jira-integration.md`). Never transition the issue here.

**The appended plan's growth is recorded.** After the planning pass writes its appends and bumps
`**Tasks appended:**`, the plan's new size is recorded as the next observation of the change's
plan-growth series:

```bash
flow tasks count -C <worktree> <name>
```

Then make the fix-run planning commit over the appended plan (**Planning commits**,
`skills/flow-contracts/git-boundaries.md`), before any implementer is dispatched.

```bash
flow stage end -command '/flow' -stage flow.document-fix -outcome completed <name>
```

**The appended plan is re-decided before load-context.** Appended tasks can move the plan's
class, and the run executes by the newest decision row, so on every fix run — **Re-plan** and
**Append anyway** alike — the decision follows the plan here. Read **Decide**
(`skills/flow/brainstorm-planner.md`) by its heading, that section alone, and run it from
`plan-class.sh` on over the appended `tasks.md`, with `<merge-base>` the resolved worktree's entry
in the state file's `worktrees` map. The rolls are name-derived, so they come out identical to the
first row's; everything else Decide outputs is recomputed — `class`, `execution`, the implementer
and fixer pairs, `groups`, the panel roster (a class move can add or drop slots, and `micro` makes it
the string `default`) and a free grouping's shape — and the run follows the new row, never the
first row's panel or pairs. Write the
decision JSON and print the `## Decision` block with its two preamble lines, then record the second
row below. The section's closing `flow.writing-plans` mark belongs to the planning run and never
runs here, and **Plan review gate** (`skills/flow/brainstorm-planner.md`) does not run on a fix run:

```bash
flow stage begin -command '/flow' -stage flow.decide -harness <harness> -session-token mf-<literal-token> <name>
flow record decision -change <name> -session-token mf-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json
flow stage end -command '/flow' -stage flow.decide -outcome completed <name>
```
