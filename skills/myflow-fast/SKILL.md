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

**Load `skills/myflow-contracts/pipeline.md` first.**

**Then register this run's steps** with the harness's task-list mechanism, before any work begins,
and keep each entry's status current as the run proceeds, per
**Progress visibility** (`skills/myflow-contracts/pipeline.md`) — that section names which steps
this command registers and is the one to read.

The reasoning behind this file lives in `skills/myflow-fast/SKILL-rationale.md`; **a
`/myflow-*` run never loads it.**

Every citation below is canonical at its target. Never restate its content here and never act on
a remembered version of it — read it fresh each time it is needed.

## State gate

```bash
myflow stage begin -command '/myflow-fast' -stage 'state gate' <name-or-best-guess>
```

Accepts **no state** (creates a change) or **`IN_PROGRESS`**. On any other state, emit the
wrong-state handoff from **Wrong state for this command**
(`skills/myflow-contracts/pipeline.md`), naming the actual state, the states this command expects,
and the suggested command instead. Proceed only on an explicit override.

```bash
myflow stage end -command '/myflow-fast' -stage 'state gate' -outcome completed <name-or-best-guess>
```

**This command marks its own four Level 1 stages only** — `/myflow-fast`'s own row in the table, not
the finer-grained `/myflow-start` and `/myflow-do` stages the sections below cite and run by
reference. A section cited here (`skills/myflow-start/SKILL.md`'s, `skills/myflow-do/SKILL.md`'s)
runs exactly as written **except** for its own stage marks: this command does not re-emit them
under `-command '/myflow-start'` or `-command '/myflow-do'`, since under `/myflow-fast` that content
is folded into one coarser stage of this command's own. See **Stage marks**
(`skills/myflow-contracts/pipeline.md`).

## No state file — brainstorm into implementation

```bash
myflow stage begin -command '/myflow-fast' \
  -stage '*(creating run)* brainstorming chained straight into implementation and the review panel, no gate in between — the full sequence, by cited section, is **No state file — brainstorm into implementation** (`skills/myflow-fast/SKILL.md`)' \
  <name>
```

This branch runs when no state file exists yet for the change — the same condition `/myflow-start`
treats as a creating run. Run these sections of `skills/myflow-start/SKILL.md` exactly as written,
by citation rather than by re-deriving their content:

- **A. Resolve the change** (`skills/myflow-start/SKILL.md`) — Jira resolution and naming
- **Ask the planning effort, the models, and the review panel roster — creating runs only**
  (`skills/myflow-start/SKILL.md`) — this skill does **not** run that question round on a creating
  run; it records the defaults directly instead, per **Recorded defaults favor speed** below.
  Sections A, B, C and D listed here still run exactly as written
- **B. Basic Workflow #1 — Brainstorming** (`skills/myflow-start/SKILL.md`) — the clarifying
  questions stay fully interactive here, exactly as under `/myflow-start`, with no auto-answering.
  The design presentation does **not** end a section, or the whole design, with a "does this look
  right?" question — present the section(s) and proceed directly, section to section and then into
  artifact creation, unless the operator raises an objection during or after that presentation. This
  is a scoped override of `superpowers:brainstorming`'s hard design-approval gate, `/myflow-fast`
  only; `/myflow-start` still asks after each section and at the separate post-design confirm,
  unaffected
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
- **5. The review panel** (`skills/myflow-do/SKILL.md`) — dispatched at the roster recorded per
  **Recorded defaults favor speed** below
- **6. Resolve the run instructions** (`skills/myflow-do/SKILL.md`)
- **7. Verify, stage, and hand off** (`skills/myflow-do/SKILL.md`)

This branch ends at `IN_PROGRESS`, with the same staged-diff-plus-run-instructions handoff
`/myflow-do` produces (see **State write and handoff** below for the one shape that differs, because
this branch published no artifact).

This is the one point in the pipeline where two skills' stage content runs back-to-back inside a
single command invocation with no operator action between them, and that is deliberate: there is no
human gate between brainstorming converging and implementation starting in the existing pipeline
either, so nothing that used to pause now doesn't.

```bash
myflow stage end -command '/myflow-fast' \
  -stage '*(creating run)* brainstorming chained straight into implementation and the review panel, no gate in between — the full sequence, by cited section, is **No state file — brainstorm into implementation** (`skills/myflow-fast/SKILL.md`)' \
  -outcome completed <name>
```

## At `IN_PROGRESS`

**An argument present.** Treat it as fix instructions and resume the worktree — run

```bash
myflow stage begin -command '/myflow-fast' \
  -stage '*(at `IN_PROGRESS`, argument present)* a fix, chained the same way — **At `IN_PROGRESS`** (`skills/myflow-fast/SKILL.md`)' \
  <name>
```

**3. Documenting a fix, before implementing it** (`skills/myflow-do/SKILL.md`) and the rest of
sections 1–7 exactly as `/myflow-do` runs them at `IN_PROGRESS`, using the argument text as the
fix's guidance. The state is written back unchanged, per
**A fix never moves the state** (`skills/myflow-contracts/pipeline.md`).

```bash
myflow stage end -command '/myflow-fast' \
  -stage '*(at `IN_PROGRESS`, argument present)* a fix, chained the same way — **At `IN_PROGRESS`** (`skills/myflow-fast/SKILL.md`)' \
  -outcome completed <name>
```

**No argument (bare invocation).** Proceed to the integrate question — run

```bash
myflow stage begin -command '/myflow-fast' \
  -stage '*(bare at `IN_PROGRESS`)* the landing question, and merge-and-push alone continues into the archive sequence within the same invocation — same section, plus **After merge-and-push specifically** (`skills/myflow-fast/SKILL.md`)' \
  <name>
```

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

```bash
myflow stage end -command '/myflow-fast' \
  -stage '*(bare at `IN_PROGRESS`)* the landing question, and merge-and-push alone continues into the archive sequence within the same invocation — same section, plus **After merge-and-push specifically** (`skills/myflow-fast/SKILL.md`)' \
  -outcome completed <name>
```

### After open PR or manual specifically

Stop after the route completes, printing the same handoff **1.5 State and handoff**
(`skills/myflow-finish/SKILL.md`) prints for run 1. Each of these two routes needs an action outside
this command's control — an external merge, or the operator's own manual steps — before archiving
can happen, so nothing continues automatically. The next bare `/myflow-fast <name>` call, once the
branch is integrated, runs run 2 (archive).

```bash
myflow stage end -command '/myflow-fast' \
  -stage '*(bare at `IN_PROGRESS`)* the landing question, and merge-and-push alone continues into the archive sequence within the same invocation — same section, plus **After merge-and-push specifically** (`skills/myflow-fast/SKILL.md`)' \
  -outcome completed <name>
```

## Recorded defaults favor speed

On the run that creates a change, `/myflow-fast` does **not** ask the planning-effort, model, or
review-panel-roster questions **Ask the planning effort, the models, and the review panel roster —
creating runs only** (`skills/myflow-start/SKILL.md`) asks interactively. It records the
recommended defaults directly, with no `AskUserQuestion` prompt: `planningEffort: default`,
`models.implementation: sonnet`, `models.reviewPanel: sonnet`, `models.panelFix: sonnet`, and
`reviewPanelRoster: light`.

An explicit session instruction naming a different value for one of these fields still overrides
that field — the recorded value is the operator's, not the default — with the override recorded
alongside its dispatch exactly as **Model policy** (`skills/myflow-contracts/pipeline.md`) already
permits for any operator override.

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

- **Never** auto-answer a brainstorming clarifying question. Clarifying questions and the design
  presentation are interactive, unchanged.
- **Never** ask the planning-effort, model, or review-panel-roster questions on a creating run —
  record the defaults per **Recorded defaults favor speed** instead.
- **Never** ask a "does this look right?" question after a design section or after the whole design
  — present and proceed directly, section to section and then into artifact creation, unless the
  operator objects during or after that presentation. This skip is scoped to `/myflow-fast`;
  `/myflow-start` still asks after each section and at the separate post-design confirm.
- **Never** continue past open PR or manual without stopping — only merge-and-push auto-continues to
  run 2.
- **Never** treat a bare invocation at `IN_PROGRESS` as a fix, and never treat an invocation carrying
  an argument as ready-to-integrate.
- **Never** publish a proposal artifact.
- **No flags.** The only argument is the optional change name or description on a creating run, or
  fix instructions at `IN_PROGRESS`; report anything else rather than ignoring it.
