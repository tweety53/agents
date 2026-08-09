---
name: myflow-fast
description: Chain brainstorming, implementation and the review panel into one command, then integrate and archive across the same three-state pipeline, pausing only at the human gates. Re-run to fix or to integrate. Use for /myflow-fast.
allowed-tools: Bash(openspec:*)
license: MIT
---

Drive the existing three-state pipeline (`STARTED` → `IN_PROGRESS` → `FINISHED`) end to end in one
command, chaining stages that carry no human gate between them. This skill introduces no new state,
no new state-file field, and no new git boundary — it reads and writes the exact same state file
every other `/myflow-*` command does, and every stage it runs is that stage's own existing skill,
cited rather than re-derived.

**Announce at start:** "Using myflow-fast for change `<name>`."

Immediately after that line, print these two commands for the operator to paste, per
**Handoff output** (`skills/myflow-contracts/pipeline.md`) — that section fixes the colour and
records why they are printed rather than invoked; do not restate its reasoning here:

```text
/rename <change-name>
/color cyan
```

**Load `skills/myflow-contracts/pipeline.md` first** — it is canonical for the states, the
command→state transition table, the wrong-state handoff, git boundaries, and the handoff output
shape.

**Then register this run's steps** with the harness's task-list mechanism, before any work begins,
and keep each entry's status current as the run proceeds, per
**Progress visibility** (`skills/myflow-contracts/pipeline.md`) — that section names which steps
this command registers and is the one to read.

## State gate

Accepts **no state** (creates a change) or **`IN_PROGRESS`**. On any other state, emit the
wrong-state handoff from **Wrong state for this command**
(`skills/myflow-contracts/pipeline.md`), naming the actual state, the states this command expects,
and the suggested command instead. Proceed only on an explicit override.

## No state file — brainstorm into implementation

This branch runs when no state file exists yet for the change — the same condition `/myflow-start`
treats as a creating run. Run these sections of `skills/myflow-start/SKILL.md` exactly as written,
by citation rather than by re-deriving their content:

- **A. Resolve the change** (`skills/myflow-start/SKILL.md`) — Jira resolution and naming
- **Ask the planning effort, the models, and the review panel roster — creating runs only**
  (`skills/myflow-start/SKILL.md`) — the same four questions, exactly as written, with the
  recommended answer overridden on two of them per **Recorded defaults favor speed** below
- **B. Basic Workflow #1 — Brainstorming** (`skills/myflow-start/SKILL.md`) — run in full, including
  its design-approval gate; brainstorming stays fully interactive here, exactly as under
  `/myflow-start`, with no auto-answering
- **C. Create the change and its artifacts** (`skills/myflow-start/SKILL.md`)
- **D. Basic Workflow #3 — Writing plans** (`skills/myflow-start/SKILL.md`)

Skip `/myflow-start`'s proposal-artifact publish step entirely — the operator who answers
brainstorming's questions is present in the same session that produces the design, so a separate
published summary of that same conversation adds no information. `artifactUrl` is written as `null`
in the state file (see **State write and handoff** below), and the handoff this branch ends with
does not offer an artifact link.

Once **D. Basic Workflow #3 — Writing plans** (`skills/myflow-start/SKILL.md`), cited above,
completes and the OpenSpec artifacts exist, continue — within the same invocation and without a
further command from the operator — into `skills/myflow-do/SKILL.md`'s implementation stage, run
exactly as `/myflow-do` runs it from `STARTED`:

- **1. Load context and validate the plan** (`skills/myflow-do/SKILL.md`)
- **2. Isolate the workspace (first run only)** (`skills/myflow-do/SKILL.md`)
- **4. Execute (SDD + TDD)** (`skills/myflow-do/SKILL.md`)
- **5. The review panel** (`skills/myflow-do/SKILL.md`) — dispatched at the roster recorded by the
  question round cited above
- **6. Resolve the run instructions** (`skills/myflow-do/SKILL.md`)
- **7. Verify, stage, and hand off** (`skills/myflow-do/SKILL.md`)

This branch ends at `IN_PROGRESS`, with the same staged-diff-plus-run-instructions handoff
`/myflow-do` produces (see **State write and handoff** below for the one shape that differs, because
this branch published no artifact).

This is the one point in the pipeline where two skills' stage content runs back-to-back inside a
single command invocation with no operator action between them, and that is deliberate: there is no
human gate between brainstorming converging and implementation starting in the existing pipeline
either, so nothing that used to pause now doesn't.

## At `IN_PROGRESS`

**An argument present.** Treat it as fix instructions and resume the worktree — run
**3. Documenting a fix, before implementing it** (`skills/myflow-do/SKILL.md`) and the rest of
sections 1–7 exactly as `/myflow-do` runs them at `IN_PROGRESS`, using the argument text as the
fix's guidance. The state is written back unchanged, per
**A fix never moves the state** (`skills/myflow-contracts/pipeline.md`).

**No argument (bare invocation).** Proceed to the integrate question — run
**Deciding which run this is** (`skills/myflow-finish/SKILL.md`) through
**1.3 Take the chosen route** (`skills/myflow-finish/SKILL.md`) exactly as `/myflow-finish` run 1
runs them: the unfinished-work gate, the landing question (open PR / merge and push / manual), the
two commits, and the chosen route.

### After merge-and-push specifically

Continue, within the same invocation and without a further command from the operator, into
**Run 2 — archive and clean up** (`skills/myflow-finish/SKILL.md`) exactly as written: verify the
merge, sync delta specs, archive the change, commit and push the archive, remove the
worktrees/branches, verify the cleanup, and write `FINISHED`. Nothing external blocks this route, so
nothing pauses between run 1 and run 2 here.

### After open PR or manual specifically

Stop after the route completes, printing the same handoff **1.5 State and handoff**
(`skills/myflow-finish/SKILL.md`) prints for run 1. Each of these two routes needs an action outside
this command's control — an external merge, or the operator's own manual steps — before archiving
can happen, so nothing continues automatically. The next bare `/myflow-fast <name>` call, once the
branch is integrated, runs run 2 (archive).

## Recorded defaults favor speed

The question round in **Ask the planning effort, the models, and the review panel roster — creating
runs only** (`skills/myflow-start/SKILL.md`), cited above, asks the same four questions this skill
runs, unchanged. Two of those four already default the way `/myflow-fast` wants and need no
override: `models.reviewPanel` already recommends `sonnet`, and `reviewPanelRoster` already
recommends `light`. The other two — `models.implementation` and `models.panelFix` — default to
`opus` there, since `/myflow-start` optimizes for a stronger implementer and fixer; this skill
overrides the recommended answer on those two questions only, to `sonnet`, favoring speed over the
stronger default. In all four `AskUserQuestion` prompts the marked recommendation is never a forced
value — an operator who answers differently has that answer recorded, exactly as under
`/myflow-start`.

## State write and handoff

Write the state file exactly per **State file** (`skills/myflow-contracts/state-file.md`) — same
shape, same fields, `updatedBy: "/myflow-fast"` — with `artifactUrl: null` on the branch that
created the change (above), and every other field written by whichever cited section actually ran:
`/myflow-do`'s **7. Verify, stage, and hand off** (`skills/myflow-do/SKILL.md`) write, when this run
ends at `IN_PROGRESS` from brainstorming and implementation or from a fix at `IN_PROGRESS`;
`/myflow-finish`'s **1.5 State and handoff** (`skills/myflow-finish/SKILL.md`) write, when this run
stops after run 1; or **Run 2 — archive and clean up** (`skills/myflow-finish/SKILL.md`)'s write,
when merge-and-push carries through to `FINISHED`.

Every handoff shape this skill prints is exactly the cited section's own block, with one exception:
the `IN_PROGRESS`-with-no-artifact case, the one shape none of the three cited skills already print,
since none of them skips the artifact.

```
## Implementation staged — review and test

**Change:** <name>
**Artifact:** none — myflow-fast does not publish one
**Recorded:** <N> decisions | none · <N> open questions | none · effort <level> · models
implementation <model>, review panel <model>, panel fixes <model> · roster <preset>
**Panel:** clean — roster: <light | standard | full>, required: <that roster's required slots>;
optional: <selected, or "none — no triggers fired">
**Staged:** N/N tasks · staged and uncommitted

Worktree:   <absolute worktree path>

Run it:
  <command>          # <app or check name>
  <command>

Review the diff, then run it:
  git -C <absolute worktree path> diff --cached
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"

Re-run this command to fix anything you find, or bare to move on to integrating it.

Next:
/myflow-fast <name>
```

Every other handoff shape — the bare-at-`IN_PROGRESS` integrate handoffs, and `FINISHED` — is
exactly the cited section's own block; no new shape is needed for either.

## Guardrails

Carry forward, by citation, every guardrail already stated under **Guardrails**
(`skills/myflow-start/SKILL.md`), **Guardrails** (`skills/myflow-do/SKILL.md`) and **Guardrails**
(`skills/myflow-finish/SKILL.md`) for the sections this skill cites above — they are not re-listed
here. Add exactly the guardrails specific to this skill:

- **Never** auto-answer a brainstorming question. The design-approval gate is interactive,
  unchanged.
- **Never** continue from brainstorming into implementation before the design is approved and the
  OpenSpec artifacts exist.
- **Never** continue past open PR or manual without stopping — only merge-and-push auto-continues to
  run 2.
- **Never** treat a bare invocation at `IN_PROGRESS` as a fix, and never treat an invocation carrying
  an argument as ready-to-integrate.
- **Never** publish a proposal artifact.
- **No flags.** The only argument is the optional change name or description on a creating run, or
  fix instructions at `IN_PROGRESS`; report anything else rather than ignoring it.
