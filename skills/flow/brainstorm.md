# Brainstorm and plan

Superpowers Basic Workflow steps **#1** (brainstorming) and **#3** (writing-plans), intertwined
with spectre artifact creation, run here. Loaded by `skills/flow/SKILL.md` on a creating run (no state) or a
resuming run (`STARTED`).

## A. Resolve the change and write `STARTED`

**Load `skills/flow-contracts/jira-integration.md`** — this section resolves the linked issue and
derives the change name from it.

**Resolve the linked Jira issue first** — it decides the change name. Follow **Resolution (how
`jiraIssue` is decided)** in `skills/flow-contracts/jira-integration.md` exactly. This is the
only phase that resolves a key.

Then the change name:

- **With a linked issue**, the name is `<lowercased-key>-<slug>`, per **Change naming**
  (`skills/flow-contracts/jira-integration.md`). Derive the slug from the issue summary when only
  a key was given.
- **Without one**, the name is the descriptive slug alone.
- If a name or description was given, use it (derive kebab-case from the description if only a
  description was given).
- **If both are omitted:** enumerate the candidate set exactly as **Change name resolution**
  (`skills/flow-contracts/pipeline.md`) defines it, restricted to changes with incomplete planning
  artifacts. Exactly one match → resume it, announcing which; multiple → **AskUserQuestion** listing
  each (name, state, last modified); zero → ask what to build.

**Transition the issue to In Progress now**, per **Transitions** in Jira integration
(`skills/flow-contracts/jira-integration.md`) — before brainstorming, so the board is correct
while planning runs. A failure is one skipped-with-reason line and planning continues; nothing about
this call may delay or alter the run.

**Mark `flow.kickoff` now that the name is fixed, and write `STARTED` immediately — before
brainstorming begins.** Write it here, at the top of this phase, rather than at the
bottom of it.

```bash
flow stage begin -command '/flow' -stage flow.kickoff -harness <harness> -session-token mf-<literal-token> <name>
```

```json
{
  "state": "STARTED",
  "branch": null,
  "worktrees": {},
  "artifactUrl": null,
  "jiraIssue": "<resolved key, or null>",
  "planningEffort": null,
  "models": { "default": null },
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/flow"
}
```

`planningEffort` and `models.default` are written `null` and stay `null` for the life of the
change: `/flow` asks no planning-effort or model question on a creating run, and models are resolved per run from the settings store, not recorded per change. `artifactUrl` stays `null` —
`/flow` publishes no proposal artifact.

```bash
flow stage end -command '/flow' -stage flow.kickoff -outcome completed <name>
```

**No further command runs before this point on a creating run** — the state write above is the
first thing this invocation does once the name is fixed, ahead of even the design conversation. The
operator sees `STARTED` recorded the moment they invoke `/flow`, whether or not the run goes on to
finish brainstorming in the same sitting.

### Resuming at `STARTED`

A run finding `"state": "STARTED"` already recorded is resuming a creating run that stopped before
reaching `IN_PROGRESS` — an interrupted session, a context limit, an earlier stop. Skip **A** above
(the name and the `STARTED` write both already exist) and determine where the run actually left off
by reading, not by assuming:

- **Does a worktree already exist for this change** — `git worktree list` naming
  `<project>/.worktrees/<name>`, or the state file's `worktrees` map non-empty. **Worktree creation
  moves to the end of planning** (`design.md`), so a `STARTED` change interrupted anywhere in **B**,
  **C** or **D** has **no** worktree yet — resolve `<changeRoot>` against the main checkout's own
  `<project>/spectre/changes/<name>/` in that case, never a worktree path.
- `spectre list --json`'s entry for this change's `done`/`total`, read against whichever
  `<changeRoot>` the check above resolved — `total == 0` means no plan exists yet: resume at **B**
  in `skills/flow/brainstorm-planner.md`. <!-- refs-guard:allow -->
- The change root's own `tasks.md` — a scaffold with no enriched steps means writing-plans has not
  run: resume at **D** in `skills/flow/brainstorm-planner.md`, still in the main checkout (no <!-- refs-guard:allow -->
  worktree exists yet). A plan meeting writing-plans quality (exact paths, verification commands, no
  placeholders) with **no worktree yet** means **D** finished but the worktree step after it did
  not: resume at **The worktree is created at the end of planning, not here** above — create the
  worktree, move the planning output into it, print the `## Decision` block and record it, then
  continue. A plan meeting writing-plans quality **with a worktree already present** means planning
  is fully done: skip straight to `skills/flow/implement.md`.

This is a pragmatic re-entrancy rule, not an exhaustively-enumerated state machine — a run resuming
at `STARTED` reads what actually exists and continues from there, the same principle every other
`/flow` re-entry point already applies. State the resumption point plainly before continuing:
"resuming `<name>` at `<point>`."

## Run brainstorming and planning directly

Sections **B**, **C** and **D** of `skills/flow/brainstorm-planner.md` are the running session's own <!-- refs-guard:allow -->
work — no dispatched subagent runs them; the session that is already running `/flow` does the
checklist, the convergence loop, artifact creation, writing-plans and the Decide step itself, on
its own model, with no relay and no return in between. Every "you" in `brainstorm-planner.md`
addresses that session.

```bash
flow stage begin -command '/flow' -stage flow.brainstorm -harness <harness> -session-token mf-<literal-token> <name>
```

Read `skills/flow/brainstorm-planner.md`'s sections **B**, **C** and **D** and follow them <!-- refs-guard:allow -->
directly — the resolved `EXECUTION_MODE_TOGGLE`, `IMPLEMENTER_MODEL_TOGGLE` and
`REVIEW_PANEL_TOGGLE` (**Model resolution**, `skills/flow/SKILL.md`) and the resolved worktree
count (per **Resolving a change's worktrees**, `skills/flow-contracts/worktree-resolution.md`) are
already in scope from this run's own earlier resolution — nothing further needs passing to a
dispatch that does not happen.

**Questions are the session's own direct `AskUserQuestion` calls**, batched exactly as
`brainstorm-planner.md`'s checklist section already states: every pending question whose wording
does not depend on another pending answer in one call, up to four per call, a dependent question
waiting for the next turn. Section B's merged convergence-and-approval confirm and its third-round
offer are asked the same way — directly, with named options, exactly as B states them.

**Prose preceding a question is shown too, not dropped.** When the checklist carries a summary
before a question — most concretely the convergence confirm's "state what you believe settled"
paragraph — show that prose to the operator as ordinary text before the **AskUserQuestion** call,
not only the bare question and options. The operator approving or answering the question is
approving against the summary they were actually shown.

Mark `flow.brainstorm` end and `flow.design-approval` begin/end around the merged
convergence-and-approval confirm, exactly as today:

```bash
flow stage end   -command '/flow' -stage flow.brainstorm -outcome completed <name>
flow stage begin -command '/flow' -stage flow.design-approval -harness <harness> -session-token mf-<literal-token> <name>
# … the operator's approve-and-move-on answer — the HARD GATE …
flow stage end   -command '/flow' -stage flow.design-approval -outcome completed <name>
```

**On the fully-seeded bypass** (**Seed from a staged research note, if one exists**,
`skills/flow/brainstorm-planner.md`), the checklist and the confirm above never run — but the same
four marks still fire, back-to-back with no interactive gap between them, so stage bookkeeping
stays consistent with every other run:

```bash
flow stage end   -command '/flow' -stage flow.brainstorm -outcome completed <name>
flow stage begin -command '/flow' -stage flow.design-approval -harness <harness> -session-token mf-<literal-token> <name>
flow stage end   -command '/flow' -stage flow.design-approval -outcome completed <name>
```

**The worktree is created at the end of planning, not here.** After the `flow.design-approval` mark
above closes, mark `flow.create-artifacts` begin and continue directly into **C** — `spectre new`
and the three artifacts — **against the main checkout's own** `<project>/spectre/changes/<name>/`,
uncommitted and never staged there, per the existing git-boundaries rule. No worktree exists yet;
**C** and **D** both run in the main checkout. Mark `flow.create-artifacts` end once **C**'s
artifacts are written, then mark `flow.writing-plans` begin and run **D** — writing-plans
enrichment and the Decide step — also in the main checkout. **D**'s Decide step writes its
decision JSON to `<project>/spectre/changes/<name>/.superpowers-sdd-decision.json` in the main
checkout instead of the usual `<abs-worktree>/.superpowers/sdd/decision.json` path, since no
worktree exists yet to hold it. Mark `flow.writing-plans` end once **D**'s plan enrichment and
Decide step complete.

**Only once `flow.writing-plans` ends does the worktree get created:**

1. `check-worktree-location.sh <project>` — exit 1 or 2 stops the run with the guard's own lines.
2. `git check-ignore -q .worktrees` from the project root. Where it exits non-zero, append
   `<project>/.worktrees/` to `<project>/.git/info/exclude` — never a commit on any branch.
3. `git worktree add <project>/.worktrees/<name> -b spectre/<name> <default-branch>` — the default
   branch by name, never HEAD: the main checkout may be on any branch and is never moved.
4. `project-get.sh <worktree> "worktree setup"`. Exit 0: run every printed line from the worktree
   root, in order, in the foreground — the printed body can carry fence markers and trailing prose
   outside the fence (as `<project>/.flow/project.md`'s `## worktree setup` section does); run only
   the fenced command lines, not those. Exit 1: the project declares no `## worktree setup`; say so
   and continue. Exit 2: stop the run, relaying the script's own line. **A command's non-zero exit
   ends your turn** naming the command and its output — a worktree that cannot be set up fails
   `flow.verify` later anyway, and the operator should see it here. The key is canonical in
   **Project configuration** (`skills/flow-contracts/project-configuration.md`).

**Then move the planning output into the new worktree**, leaving nothing behind in the main
checkout:

```bash
mkdir -p <worktree>/.superpowers/sdd
mv <project>/spectre/changes/<name> <worktree>/spectre/changes/<name>
mv <project>/spectre/changes/<name>/.superpowers-sdd-decision.json <worktree>/.superpowers/sdd/decision.json
```

Once the Decide step finishes, print the `## Decision`
block verbatim — the shape **The `## Decision` block** (`design.md`) shows — then run the record
sequence that section states:

```bash
flow stage begin -command '/flow' -stage flow.decide -harness <harness> -session-token mf-<literal-token> <name>
flow record decision -change <name> -session-token mf-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json
flow stage end -command '/flow' -stage flow.decide -outcome completed <name>
```

`<abs-worktree>/.superpowers/sdd/decision.json` is the file the move above just placed there.
Continue into `skills/flow/implement.md` directly — no dispatch record to close, since nothing was
dispatched.

## Resume and fix runs

**Resume and fix runs** (`design.md`) is canonical for both cases: a run resumed at `STARTED` reads
`flow record decisions -change <name>` and follows the newest row rather than re-rolling, or
re-runs the Decide step alone when none exists yet; a fix run's `flow.document-fix`
(`skills/flow/implement.md`) hands the appended plan through the same Decide step, re-grouping the
review panel on the same name-derived `bundle_roll`, and recording a second row whose rolls — being
name-derived — stay identical to the first, so only `class`, `groups` (the appended plan's bundles)
and a free grouping's shape can change.
