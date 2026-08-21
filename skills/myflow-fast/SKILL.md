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
records why they are printed rather than invoked:

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

## State gate

**Read the change's state before marking anything.** A mark that finds no change record creates one
via its own bootstrap path (`myflow stage begin`'s synthetic-change fallback), and that record must
never be the thing this gate then reads back to decide what state the run is in — a record whose
only author is a mark's own side effect does not satisfy a state gate expecting a state a pipeline
command wrote.

```bash
myflow state get <name-or-best-guess> -C <repo-root>
```

- **Exit 1** (`myflow: no state recorded for ...`) is **no state**: a creating run.
- **Exit 0** with `"synthetic": true` in the printed JSON is **also no state** — this is the record
  a stage mark's own bootstrap path would have produced had it run first; `<agents repo>/stats/cmd/myflow/state.go`
  (`markSyntheticIfNeeded`) sets this field whenever the record's `updatedBy` is that bootstrap
  path's own sentinel, so this is a field to test, never a string to compare by hand.
- **Exit 0** with no `"synthetic"` field reports the change's real `state` field.

**Check guard presence.** Per **Guard presence check** (`skills/myflow-contracts/pipeline.md`),
confirm every guard named in `/myflow-do`'s, `/myflow-finish`'s and `/myflow-status`'s own
presence checks — the union `skills/myflow-fast/scripts/` carries — is present there. A complete set prints nothing; any absence prints that section's
block once, and the run continues under each guard's own hand-run fallback.

**The read above tolerates a guess; the mark below, which writes, does not — its `<change>`
argument is `<name>`.** On a creating run `<name>` does not exist yet at this point, so defer the
`do.state-gate` `begin`/`end` pair below: do not fire it here. Instead, carry it with you into
**A. Resolve the change** (`skills/myflow-start/SKILL.md`) and fire it there yourself, the moment
that section fixes the change name — immediately *before* that section's own `start.resolve-change`
`begin` fires. **This `do.state-gate` mark is `/myflow-fast`'s own obligation, not something section
A's text states or fires for you**: section A knows nothing about it, and is cited here only for the
underlying rule of *how* a name gets resolved, per **The `<change>` argument is always a resolved
change name** (`skills/myflow-contracts/pipeline.md`). At `IN_PROGRESS` the name is already resolved
before this gate runs in every path, so both marks fire right here, in reading order, with no detour
into section A. Either way, a creating run's `do.state-gate` records a near-zero duration between
begin and end, since the name is fixed and marked in the same step — accepted, because the
alternative buys a longer duration by marking `<name-or-best-guess>` here, which bootstraps a change
row for a name nobody chose.

**Generate this run's session token once, right now, before the first mark below — a short, unique
literal string — and reuse that exact same value at every `stage begin` this run makes, including
every mark inside a cited section below.** Never generate a fresh token per mark, and never let a
cited section's own "generate once" instruction cause a second token part-way through this run: this
whole chained sequence, start through finish, is **one run** and carries **one token**
(design.md's "one token per session, not one per mark").

```bash
myflow stage begin -command '/myflow-fast' -stage do.state-gate -harness <harness> -session-token mf-<literal-token> <name>
```

Accepts **no state** (per the read above — literal absence, or a synthetic-only record) or
**`IN_PROGRESS`**. On any other state, emit the wrong-state handoff from
**Wrong state for this command** (`skills/myflow-contracts/pipeline.md`), naming the actual state,
the states this command expects, and the suggested command instead. Proceed only on an explicit
override.

```bash
myflow stage end -command '/myflow-fast' -stage do.state-gate -outcome completed <name>
```

**This command mints no stage key of its own — it marks the same granular Level 1 stages the
sections it cites already document**, per design.md's "`/myflow-fast` reuses the chained commands'
stage names" (kan-172): `/myflow-fast`'s allowed stage set is the **union** of `/myflow-start`'s,
`/myflow-do`'s and `/myflow-finish`'s, expressed in the Level 1 table's Commands column, not a
fourth set of names. A section cited here (`skills/myflow-start/SKILL.md`'s, `skills/myflow-do/SKILL.md`'s,
`skills/myflow-finish/SKILL.md`'s) runs exactly as written, **including its own stage marks** —
every `myflow stage begin`/`myflow stage end` in that section fires exactly where it is written,
with exactly the key and harness rules its own file states, and **two substitutions**: `-command`
reads `/myflow-fast` in every one of them, never the cited section's own command name, and
`-session-token` carries this run's own token throughout — the one generated once at the state gate
above, never a fresh one minted at the cited section's own first mark.
This is what makes a fast run's stages directly comparable, stage for stage, against the equivalent
`start`→`do`→`finish` sequence — a fast run records `do.review-panel`, not a `fast.*` key of its
own. A stage a cited section skips under the conditions stated in its own file (e.g.
`start.ask-options` on this skill's creating run, `do.document-fix` on a first run) is skipped here
too, and marks nothing. See **Stage marks** (`skills/myflow-contracts/pipeline.md`).

## No state file — brainstorm into implementation

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

Every stage mark for this branch is the cited section's own — `start.resolve-change`,
`start.brainstorm`, `start.design-approval`, `start.create-artifacts`, `start.writing-plans`,
`do.load-context`, `do.isolate-workspace`, `do.sdd-tdd`, `do.review-panel`, `do.run-instructions`,
`do.workspace-export`, `do.lint-and-test`, `do.stage-diff`, `do.write-in-progress`, in that order —
each fired exactly where its own `skills/myflow-start/SKILL.md` or `skills/myflow-do/SKILL.md`
section documents it, under `-command '/myflow-fast'`, per this file's own **State gate** section
above. `start.ask-options` and `start.publish-proposal` are skipped, per this branch's own overrides
above, and mark nothing.

## At `IN_PROGRESS`

**An argument present.** Treat it as fix instructions and resume the worktree, running
**3. Documenting a fix, before implementing it** (`skills/myflow-do/SKILL.md`) and the rest of
sections 1–7 exactly as `/myflow-do` runs them at `IN_PROGRESS`, using the argument text as the
fix's guidance. The state is written back unchanged, per
**A fix never moves the state** (`skills/myflow-contracts/pipeline.md`). Every stage mark is
`do.document-fix` followed by the same sections 1–7 keys listed above, again under
`-command '/myflow-fast'`.

**No argument (bare invocation).** Proceed to the integrate question, running
**Deciding which run this is** (`skills/myflow-finish/SKILL.md`) through
**1.3 Take the chosen route** (`skills/myflow-finish/SKILL.md`) exactly as `/myflow-finish` run 1
runs them: the unfinished-work gate, the landing question (open PR / merge and push / manual), the
two commits, and the chosen route. Every stage mark is the cited section's own —
`finish.preflight`, `finish.unfinished-work-gate`, `finish.landing-question`,
`finish.preserve-sessions`, `finish.commit-two`, `finish.landing-routes` — again under
`-command '/myflow-fast'`.

### After merge-and-push specifically

Continue, within the same invocation and without a further command from the operator, into
**Run 2 — archive and clean up** (`skills/myflow-finish/SKILL.md`) exactly as written: verify the
merge, position the checkout, sync delta specs, archive the change, commit the archive, remove the
worktrees/branches, verify the cleanup, write `FINISHED`, run self-review, and push the archive
branch and open its pull request. Nothing external blocks this route, so nothing pauses between
run 1 and run 2 here. Every stage mark is the cited section's own —
`finish.write-in-progress`, `finish.move-in-review`, `finish.verify-merge`, `finish.sync-archive`,
`finish.commit-archive`, `finish.cleanup`, `finish.verify-cleanup`, `finish.write-finished`,
`finish.self-review`, `finish.push-archive` — again under `-command '/myflow-fast'`.

### After open PR or manual specifically

Stop after the route completes, printing the same handoff **1.5 State and handoff**
(`skills/myflow-finish/SKILL.md`) prints for run 1. Each of these two routes needs an action outside
this command's control — an external merge, or the operator's own manual steps — before archiving
can happen, so nothing continues automatically. The next bare `/myflow-fast <name>` call, once the
branch is integrated, runs run 2 (archive). The route taken already closed `finish.landing-routes`
above — no further mark fires until that next call.

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
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)

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
- **Never** ask check 4's ignored-files confirmation before removing a worktree in run 2. **Report**
  what `--force` will destroy — how many ignored files, which are build output, and which are
  irreplaceable together with whether they were already preserved — and proceed. This is a scoped
  override of the disclosure ask in **Worktree cleanup**
  (`skills/myflow-contracts/finish-contract.md`), `/myflow-fast` only; `/myflow-finish` still asks.
  The reason it is safe here is that the records worth keeping are already out of the worktree by
  this point: `finish.preserve-sessions` copies the panel record and the SDD ledger into
  `<project>/docs/superpowers/`, and `finish.commit-two` commits them, so what check 4 lists is build output
  plus the per-review diffs the **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`)
  already declares worktree-lifetime. **Checks 1, 2, 3 and 5 remain gates** — a failure in any of
  them still stops cleanup with no worktree touched — and check 4 turning up something genuinely
  irreplaceable and *unpreserved* (a gitignored `.env`, a local override, a `.dev-state`) is not this
  override's case: stop and ask, exactly as the contract requires.
- **No flags.** The only argument is the optional change name or description on a creating run, or
  fix instructions at `IN_PROGRESS`; report anything else rather than ignoring it.
