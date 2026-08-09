# myflow pipeline

The three-state pipeline itself: state definitions, the command→state transition table, git
boundaries and the handoff shape. The finish contract lives in
`skills/myflow-contracts/finish-contract.md`, which is canonical for it.

**Load this file when running any `/myflow-*` command.** It is split out of
`rules/myflow-manual-review.mdc` so the always-on rule layer carries only the trigger, not the
whole state machine — the same reason the other contract files beside it were split out.

This file is **canonical** for everything in it. Where a skill or command disagrees with it, this
file wins.

The reasoning behind this file lives in `skills/myflow-contracts/pipeline-rationale.md`;
**a `/myflow-*` run never loads it.**

## States

A change is always in exactly one of three states, recorded in its state file.

```text
/myflow-start  → STARTED      you: read the proposal artifact
/myflow-do     → IN_PROGRESS  you: review the staged diff and run the apps
/myflow-finish → FINISHED     terminal (second run — see the finish contract)
```

**Each command ends in the state named after it**, so the state vocabulary and the command
vocabulary are the same words.

**The human gate is a property of the state, not a separate stage.** `IN_PROGRESS` *means* a
staged diff is waiting for the human to review, alongside the run instructions the handoff
printed. Nothing records that the review or the testing happened; the operator running the next
command is what carries the change forward. This is why no `*-done` command exists — there would
be nothing for one to write.

| State | Means | Waiting on |
|-------|-------|-----------|
| `STARTED` | The proposal exists and is published | you — read the artifact |
| `IN_PROGRESS` | The implementation is staged and the handoff printed the run instructions | you — review the diff, run the apps |
| `FINISHED` | Archived, pushed, worktrees removed | — |

**Reviewing and testing are one gate.** `/myflow-do` produces both surfaces in the same run, so
the human does both at one sitting. There is no state between implementation and finishing —
integration is not a stage, it is the first half of finishing.

## Stage exit — never the command's own judgment

Within a single command's run, a stage that loops — most concretely `/myflow-start`'s brainstorm
stage, whose convergence test reopens after every planning-stage exchange that leaves a question the
command's inputs do not answer — never closes on the command's own judgment. It closes only on an
explicit operator answer: at a confirm, or by declining an offer, recording what is still open
rather than assuming it away. The one bounded exception is a session that cannot ask at all: it
records the confirm itself as an open question and ends the stage there, since no operator answer
could ever arrive through it. An operator who is present but silent is not that exception and still
gets another round.

The tuned threshold, the two prompts, and why their opposite recommendations are not to be
harmonised belong to the command itself — **Convergence** (`skills/myflow-start/SKILL.md`) — and are
not restated here.

## Command surface

Three pipeline commands, one composite command, plus one read-only one. `/myflow-fast` is the
composite command — it chains the other three's stage content across state transitions that carry
no human gate; see `skills/myflow-fast/SKILL.md` for what it does. **No command accepts a flag.**
The only argument is the optional change name — see **Change name resolution**.

An argument that is not a known change name is **reported**, not silently ignored — a silently
ignored word is indistinguishable from a flag that stopped working.

## State transitions

| Command | Accepts | Ends at |
|---------|---------|---------|
| `/myflow-start` | *(none — creates the change)* · `STARTED` | `STARTED` |
| `/myflow-do` | `STARTED` · `IN_PROGRESS` | `IN_PROGRESS` from `STARTED`; **otherwise unchanged** |
| `/myflow-finish` | `IN_PROGRESS` | `IN_PROGRESS` after run 1; `FINISHED` after run 2 |
| `/myflow-fast` | *(none — creates the change)* · `IN_PROGRESS` | `IN_PROGRESS` from a creating or fix run; from a bare invocation at `IN_PROGRESS`, `IN_PROGRESS` or `FINISHED` depending on the route chosen |
| `/myflow-status` | any — read-only, never block | unchanged |

**This table is authoritative.** Every command file — in **both** command trees (`commands/` and
`commands-claude/`) — must state exactly the states its row lists, and must agree with the skill it
delegates to. When a command and its skill disagree, whichever the agent reads first wins, which is
non-determinism in the one layer that must be deterministic.

### Every command is re-entrant

Re-invoking a command is the supported way to revise its output. There is no separate `*-fix`
command:

- **`/myflow-start` at `STARTED`** revises the proposal and republishes the artifact to the
  **same** URL.
- **`/myflow-do` at `IN_PROGRESS`** resumes the existing worktree and applies a fix, documenting it
  in `proposal.md`/`tasks.md` or a nested `<name>-fix-N` sub-change first, and refreshing the test
  guide alongside the code so the two surfaces never drift apart.
- **`/myflow-finish` at `IN_PROGRESS`** integrates on its first run and archives on its second.

### A fix never moves the state

`/myflow-do` advances the state **only** from `STARTED` to `IN_PROGRESS`. From `IN_PROGRESS` it
writes the state back exactly as it found it.

No field records where a fix was raised. Whether the human re-reviews the diff or re-runs the apps
after a fix is their decision.

## Wrong state for this command

**On a mismatch, stop.** Report the actual state, the states the command expects, and the command
that should run instead, then **AskUserQuestion** for an explicit override with **"No — run the
suggested command instead"** as the default and recommended answer. Only proceed when the user
explicitly chooses to override. Never advance from a wrong starting state silently.

```
## Wrong state for this command

**Change:** <name>
**Current state:** <actual> (set by <updatedBy>, <updatedAt>)
**This command expects:** <expected>
**Suggested instead:** /myflow-<other> <name>
```

## Git boundaries

| Command | Condition | Allowed git actions |
|---------|-----------|---------------------|
| `/myflow-start` | — | **None — stages planning artifacts and never commits** |
| `/myflow-do` | from `STARTED` | Create branch/worktree + **commits each task** (fixups fold in) — no push, merge, or PR |
| `/myflow-do` | at `IN_PROGRESS`, no `prUrl` | Resume **existing** worktree + **commits fixups** the same way — no push, merge, or PR |
| `/myflow-do` | at `IN_PROGRESS`, `prUrl` recorded | **Commits twice and pushes** to the PR branch — implementation, then planning artifacts; the one exception |
| `/myflow-finish` | run 1 | **Commits twice** — implementation, then planning artifacts — and pushes; opens a PR or merges, by the operator's choice |
| `/myflow-finish` | run 2 | **Commits and pushes the archive**; removes worktrees and branches |
| `/myflow-status` | — | None — read-only |

**The planning paths** are the two that
**Handoff output** (`pipeline.md`) names below. `/myflow-do` clears them from the index and only
then stages with them excluded by pathspec — an exclusion governs what an
`add` adds and cannot retract what an earlier step staged, so the clearing pass is what makes the
rule hold rather than merely assert it. Its staging area therefore carries implementation only, and
`/myflow-finish` is what commits them.

`git add -A` respects `.gitignore`. Never force-add.

**Both commits are guarded, and an empty one is skipped rather than failed.** Each commit is
preceded by a staged-changes test, and the whole sequence is one `&&` chain, run as a single
command. See **Git boundaries** (`skills/myflow-contracts/pipeline-rationale.md`) for the ordinary
cases this guards against and why it is a chain rather than `set -e`.

```bash
git -C <abs-worktree> reset -q -- openspec/ docs/superpowers/ \
  && git -C <abs-worktree> add -A -- . ':(exclude)openspec/' ':(exclude)docs/superpowers/' \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "<type>(<name>): <what the implementation does>"; } \
  && git -C <abs-worktree> add -A \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "chore(<name>): plan and session records"; }
```

**A skipped commit is reported, and a FAILED commit — one a hook rejects — is a git failure: report
git's own output and stop.** See **Git boundaries** (`skills/myflow-contracts/pipeline-rationale.md`)
for what an unguarded sequence would do instead.

**A planning path that is a tracked symlink stops the run, and is never worked around.** When either
of the two is a symlink, `git add -A -- . ':(exclude)docs/superpowers/'` exits 128 with
`fatal: pathspec … is beyond a symbolic link` and stages **nothing at all**. Report that message,
name the path, and stop at `IN_PROGRESS`. The only way to stage past it is a bare `git add -A`,
which puts the planning artifacts into the implementation commit — the one outcome this split
exists to prevent — so the fix belongs in the repository, by making the path a real directory.

## Preserving the session records

`/myflow-do` reads this table on its `prUrl` commit path, and `/myflow-finish` reads it in
run 1; the invocation of `scripts/preserve-session-records.sh` itself is described by each
caller.

| Outcome | What it means | What you do |
|---------|---------------|-------------|
| `skipped: <src> (absent)`, exit 0 | The source does not exist. A change may legitimately have no panel record. | Nothing. Proceed. |
| `preserved: <dest>`, exit 0 | The record reached the repository at that path. | Nothing. Proceed. |
| A message on stderr, **exit non-zero** | A copy was attempted and refused or failed — an untrusted source or destination path, or a write that could not be made. | **Report it to the operator, with the script's own message.** Then proceed with the integration. |

**A non-zero exit is never silent and never a stop; the remaining sources are still attempted after
any one failure, and the handoff names which records were preserved and which were not.** See
**Preserving the session records** (`skills/myflow-contracts/pipeline-rationale.md`) for why.

**Do not harmonise the two orderings for symmetry** — here the preservation call runs before
staging; in `/myflow-do` it runs after. See **Preserving the session records**
(`skills/myflow-contracts/pipeline-rationale.md`) for why the asymmetry is what keeps preserved
records out of the staged-only path.

## Progress visibility

**Every pipeline command drives the harness's task-list mechanism.** `/myflow-start`, `/myflow-do`,
`/myflow-finish` and `/myflow-fast` register their steps with it at the start of a run and keep each entry's status
current — in progress when its step begins, completed when that step finishes — so the harness's
live progress view, a count line and one line per task, renders throughout the run rather than
arriving with the handoff. The count line then distinguishes done, in progress and open at every
point.

| Command | One entry per |
|---------|---------------|
| `/myflow-start` | its brainstorming checklist item and each artifact it produces |
| `/myflow-do` | each item in `tasks.md`, in plan order |
| `/myflow-finish` | each step of the run it is performing — run 1's steps or run 2's, never both |
| `/myflow-fast` | whichever cited stage is running at the time, at that stage's own granularity — brainstorming checklist items on the brainstorm+create branch, `tasks.md` items on the implementation branch, a finish run's steps on the integrate branch |

`/myflow-status` is read-only and **registers nothing**. Registering steps for a
report would put entries on the operator's task list for work nobody is doing.

**The progress view is a view, never a record.** No command, guard or contract reads the harness's
task list back as evidence of what was done. `tasks.md` remains the single source of truth for a
plan's completion state, and `scripts/check-unfinished-work.sh` reads that file — a second source of
completion state would be one that guard cannot see.

**No third checkbox marker is added to `tasks.md`** to carry an in-progress state; the in-progress
count comes from the harness's task list alone, which no run persists. See **Progress visibility**
(`skills/myflow-contracts/pipeline-rationale.md`) for why a marker would be unsafe.

**Stated against the mechanism, never against one harness's tool.** Where a harness offers no
task-list mechanism, the command prints the equivalent block instead — a count line naming how many
steps are done, in progress and open, followed by one line per step marked done or not done — and no
harness has to gain a task tool to satisfy the rule. See **Progress visibility**
(`skills/myflow-contracts/pipeline-rationale.md`) for why the rule is stated against the mechanism.

## Handoff output

Every command ends in the same shape, and prints **nothing** after it:

```
<1–3 lines: what actually happened>

<absolute paths to anything the operator needs to open>

Next:
/myflow-finish <name>
```

- **The next command is the last line** — bare, copy-pasteable, with no prose after it. See
  **Handoff output** (`skills/myflow-contracts/pipeline-rationale.md`) for why.
- **`/myflow-finish` run 1 names itself** as the next command, because that is what the operator
  runs once the branch is merged. Only a run 2 that **completed** is terminal and names nothing — a
  run 2 that stopped on a cleanup leftover names itself too, for the same reason: the operator
  clears what remains and runs it again.
- **Only what the operator must act on.** Do not restate the plan, enumerate completed internal
  steps, or repeat content available at a path you just gave.
- **Link, never paste.** Diffs and plans are given as absolute paths.
- **Every path is absolute** — in handoffs, in IntelliJ commands, in run instructions. Never a
  relative path, never `../<other-app>`, and never a main-checkout path while an apply worktree
  holds the work. Resolve app roots from `git worktree list` or the state file's `worktrees` keys.
- **`/myflow-do` never stages `openspec/` or `docs/superpowers/` before
  finish**, and the list is fixed here rather than configured per project. `/myflow-finish` run 1
  stages them and commits them separately from the implementation, so nothing is lost. See
  **Handoff output** (`skills/myflow-contracts/pipeline-rationale.md`) for why leaving them unstaged
  — rather than filtering a display — is what keeps them out of every view of the staging area:

  ```bash
  git -C <abs-worktree> diff --cached --stat
  git -C <abs-worktree> diff --cached
  ```

### The block each state renders

The block a state hands off is defined in **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`), which is canonical for the three per-state templates,
the run-only rule and the rendering-selection table. `/myflow-status` loads it; a producing command
carries only the block it prints.

### The tab commands, printed at the start of a run

`/myflow-start`, `/myflow-do` and `/myflow-finish` each print, immediately after their announcement
line and before any work, two commands for the operator to paste:

```text
/rename <change-name>
/color cyan
```

**They sit at the start of the run, not in the handoff block, and the colour is one fixed value —
`cyan` — for every command and change, signifying only that a pipeline command owns the tab.**
`/myflow-status` prints neither line: a read-only report does not own the tab. See **The tab
commands, printed at the start of a run** (`skills/myflow-contracts/pipeline-rationale.md`) for why
cyan was chosen and why the lines are printed at the start rather than the end.

**They are printed rather than invoked because neither is reachable from inside a run.** Both
commands are real, and both routes to calling them are closed. That was established by measurement,
and it is recorded here so the next reader neither repeats the investigation nor treats the printing
as an oversight to correct:

- the harness's `SlashCommand` tool exposes only commands of `type: "prompt"`, while `/rename` and
  `/color` are `type: "local"` or `"local-jsx"` — so the tool route is closed; and
- no writable `/dev/tty` is available to a command — so writing the terminal escape sequence
  directly is closed too.

- **Where a harness offers a reachable way to set the tab's name and colour from inside a run**, the
  command may use it, and then prints nothing — the lines exist to be pasted, and there is nothing
  to paste once the thing is done.
- **Where it does not** — Claude Code today, for the two reasons measured above, and any harness
  with no tab concept at all — the command prints the two lines. In a harness with no such commands
  they are inert text the operator ignores, which costs two lines and leaves nothing broken.

The rule is satisfied by whichever mechanism the harness provides, and no harness has to gain a tab
API to satisfy it. What is **not** optional is that the naming happens at the start of the run: a
command that silently skips it because its harness offers no tool has dropped the requirement, not
adapted it.

## IntelliJ commands

Every state that waits on a human must print a copy-paste command in its handoff:

```bash
open -na "IntelliJ IDEA" --args "<absolute path>"
```

Use `open -na`, not the `idea` shim — that shim is not on this machine's PATH. `open` resolves the
app by name (bundle `com.jetbrains.intellij`), returns immediately instead of blocking the shell,
and reuses a running instance.

| State | Path to open |
|-------|--------------|
| `STARTED` | main checkout (artifacts live there; no worktree exists yet) |
| `IN_PROGRESS` | apply worktree root |

Paths are absolute, resolved from `git worktree list`. Never emit a relative path.

## Finish contract

**Finish contract** (`skills/myflow-contracts/finish-contract.md`) governs the preflight signals,
both runs' procedures, base-branch resolution and worktree cleanup, and `/myflow-finish` is the
only command that loads it.

## Resolving a change's worktrees

Any step in any command that needs "the worktrees" for a change — a preflight verdict, a gate that
runs once per worktree, a status report, or a removal — resolves the set first; it never loops over
the state file's `worktrees` map directly. A map with zero keys and a map that was never populated
look identical to a raw read, so a direct read cannot tell "nothing to do" from "unpopulated," and
a zero-iteration loop over either reads as a pass it never earned.

**A resolved set that comes back empty is never a vacuous pass.** Report it explicitly rather than
let a zero-iteration loop read as "every worktree passed," "every worktree is merged," or whatever
verdict the calling step would otherwise default to on no evidence at all. A gate stops and asks the
operator, exactly as it would on any other refusal; a read-only report says so in its own output
instead of silently omitting the change.

This binds every command that iterates a change's worktrees: `/myflow-do`'s workspace-isolation
gate, `/myflow-status`'s merge-status report, and `/myflow-finish`'s preflight verdict,
unfinished-work gate and run 2 removal alike — each resolves its own set through this rule rather
than restating it. How a command resolves the set beyond reading the state file's map — whether it
falls back to a filesystem scan, and what an inconclusive answer does next — is that command's own;
**Resolving a change's worktrees** (`skills/myflow-contracts/finish-contract.md`) is canonical for
`/myflow-finish`'s own scan-and-resolve procedure and is the fullest example of applying this rule.

## Temporary artifacts registry

Every artifact the pipeline creates, with what creates it, where it lives, and what removes it. It
stays in the core because two contracts `/myflow-do` loads — `project-configuration.md` and
`workspace-isolation.md` — cite into it, so a reader of either never meets a citation to a file
their command does not load.

| Artifact | Created by | Location | Removed by |
|----------|-----------|----------|-----------|
| Per-task and review diffs | `/myflow-do` | `.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Panel record | `/myflow-do` | `.superpowers/sdd/` | preserved at run 1; removed with the worktree |
| SDD ledger | `/myflow-do` | `.superpowers/sdd/` | preserved at run 1; removed with the worktree |
| Proposal artifact source | `/myflow-start` | the state directory | run 2, only if a preserved copy exists |
| Worktree | `/myflow-do` | per the `worktrees` keys | run 2, after its existing checks |
| Local branch | `/myflow-do` | the repository | run 2, `git branch -d` |
| Remote branch | finish run 1 | `origin` | run 2, without a further prompt |
| Change directory | `/myflow-start` | `openspec/changes/<name>/` | moved to the archive, never deleted |
| Workspace database and bucket | the project's `create` command, on first start in a worktree | inside the project's shared data services | run 2, the project's `remove` command |
| Claimed cache index | `/myflow-do`, by probing, when it exports the workspace's variables | one of the shared cache's fixed indices | nothing in this pipeline — see below |
| State file | every command | the state directory | never — it is the terminal record |

**This table is the one place a cleanup rule is stated.** Everything else that mentions a removal
points here rather than restating it. **Worktree cleanup**
(`skills/myflow-contracts/finish-contract.md`) is the *procedure* for the rows removed there, not a
second statement of the rule. See **Temporary artifacts registry**
(`skills/myflow-contracts/pipeline-rationale.md`) for why a stale second copy would be dangerous.

**An artifact no row accounts for is a defect in the registry**, corrected by adding the row —
never left unaccounted for on the grounds that something probably removes it. See **Temporary
artifacts registry** (`skills/myflow-contracts/pipeline-rationale.md`) for the incident that
established this.

**Where the proposal artifact source comes from, and why its row is conditional.** `/myflow-start`
writes `<state-dir>/<name>-proposal-artifact.html` so a revision round can republish to the same
URL, and the preserved copy its row requires lives under `docs/superpowers/artifacts/`. The terminal
state file keeps `artifactUrl` indefinitely, so deleting the only source that could republish that
URL would leave it advertised and unrepublishable. No preserved copy → leave the file and say so.
The deletion is disclosed the same way the worktree removal is.

**The workspace row belongs only to a project that declares isolation, and for every other project
it is a row about nothing — which is why it names no database, no bucket and no service.** A project
declares the commands that create these resources, that remove them, and that report which of them
survived; the section holding those declarations is the one
**Project configuration** (`skills/myflow-contracts/project-configuration.md`) is canonical for.
Which resources there are, and how each derived value is derived, is stated under
**What the id derives** (`skills/myflow-contracts/workspace-isolation.md`).

**This is the one row whose removal is verified by asking rather than by looking: a survivor is
established from the project's own survivor report, never inferred from the removal's exit code** —
stated once under **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`), with
the report's output and exit-code contract under **Project configuration**
(`skills/myflow-contracts/project-configuration.md`). A report that could not reach its service is
skipped rather than failed. See **Temporary artifacts registry**
(`skills/myflow-contracts/pipeline-rationale.md`) for why asking, not looking, is required here.

**Nothing removes the claimed cache index, and nothing in this pipeline can.** It is not written
into the state file, and the project's `remove` command does not touch it either — stated as a
property of the `cache index` resource word under **Project configuration**
(`skills/myflow-contracts/project-configuration.md`). See **Temporary artifacts registry**
(`skills/myflow-contracts/pipeline-rationale.md`) for why: guessing an index to sweep risks flushing
another workspace's.

## State file

The contract governing where a change's state file lives, its full JSON shape, monotonic state
writes, and carry-forward rules. **State file** (`skills/myflow-contracts/state-file.md`) — load it
before reading or writing a state file.

## Project configuration

The contract governing `.myflow/project.md` — its optional keys, how a `## standards` entry
resolves to a file, and the containment rules that keep resolution safe.
**Project configuration** (`skills/myflow-contracts/project-configuration.md`) — load it before
resolving project configuration.

## Jira integration

The contract governing how a change is linked to a Jira issue, transitioned through the pipeline,
and has its description synced — including that Jira is never a gate and never blocks a state
write. **Jira integration** (`skills/myflow-contracts/jira-integration.md`) — load it before any
Jira-related step.

## Model policy

**This section is canonical for the model roles, their defaults and how an override applies** — the
one file every `/myflow-*` command loads for them. See **Model policy**
(`skills/myflow-contracts/pipeline-rationale.md`) for why this is the one place the rule lives.

**Change the capability first and bring this section with it: a section that contradicts the
OpenSpec requirement is this file's defect, not the spec's.** See **Model policy**
(`skills/myflow-contracts/pipeline-rationale.md`) for which file governs which layer and why runtime
reads this section rather than the live spec.

`/myflow-start` should run on **Opus** (or the harness's strongest available model) — brainstorming
and design benefit most from stronger reasoning. Every other `/myflow-*` command should run on
**Sonnet** (or the harness's standard default), and **every review-panel reviewer runs on the
panel's model — Sonnet by default** — regardless of the parent model. Sonnet is the default rather
than an absolute because a change may record its own panel model, per the three roles below; what
never varies is that the panel's model is *chosen*, not inherited from the parent session.

**Implementer subagents dispatched by `/myflow-do` run on Opus** (or the harness's strongest
available model), which **explicitly overrides** superpowers:subagent-driven-development's model
guidance. See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for why that
guidance's cost savings do not apply here.

**Two further instructions in that same upstream skill are also overridden: dispatching the final
review on the most capable model, and escalating the model in fix rounds 4-5.** myflow fixes every
panel slot at the panel's model instead and escalates breadth (the conditional Security, Adversarial
and extra-principle-lens slots) rather than the model. See **Model policy**
(`skills/myflow-contracts/pipeline-rationale.md`) for the reasoning.

**An explicit operator instruction overrides either default, in either direction** — raising the
panel to Opus for a change that warrants it, or lowering the implementer for genuinely mechanical
work. Record the instruction with the dispatch; an override nobody wrote down is indistinguishable
from a mistake.

**Three model roles are chosen once per change and recorded in its state file.** The run that
**creates** a change — `/myflow-start` finding no state file, exactly as the planning-effort
question determines it — asks three separate questions, one per role, each naming its default and
marking it as the recommendation. A revision round states the recorded values and does not ask
again, and every other command carries them forward verbatim, as it does the linked Jira issue.

| Role | Key under `models` | Default |
|------|--------------------|---------|
| The implementer subagents `/myflow-do` dispatches | `implementation` | Opus, or the harness's strongest available model |
| Every review-panel slot that takes a model override | `reviewPanel` | Sonnet |
| The subagents that repair panel findings | `panelFix` | Opus, or the harness's strongest available model |

The field shape, and the rule that an absent key reads as *not recorded*, belong to
**State file** (`skills/myflow-contracts/state-file.md`) and are not restated here.

**The panel-fix default is the strongest available model, and deliberately not Sonnet** — the role
applies fixes, which is implementer work, so the implementer rule above governs it too. See **Model
policy** (`skills/myflow-contracts/pipeline-rationale.md`) for why.

**A recorded choice is the operator override this section already permits, made durable.** It
applies to every run of the change without being restated, which is the point of recording it. A
session instruction is the narrower and later of the two: it governs the run in which it is given
and is recorded with its dispatch exactly as above.

**These fields record intent; the ledger records what happened.** A recorded value does **not**
replace the per-dispatch ledger line, which remains the only evidence of the model a dispatch
actually ran on. Slots dispatched by `subagent_type` take no override from this mechanism either —
no recorded panel model is passed to them, none is written for them in the ledger, and their entries
still read `unknown (agent-defined)`.

**Every subagent dispatch records the model it used** in the SDD ledger, alongside the task it ran.
See **Model policy** (`skills/myflow-contracts/pipeline-rationale.md`) for why, and for the history
behind this rule.

Where the dispatcher **cannot know** the model, the ledger records `unknown (agent-defined)` and
never a guess. Slots dispatched by `subagent_type` (Bugbot, Security Review) resolve their model
from their own agent definition, which the dispatcher does not read; writing a plausible slug for
them puts an unmeasured value into the audit trail.

**This record outlives the change: the ledger is preserved under `docs/superpowers/ledgers/` at run
1, before the worktree carrying `.superpowers/sdd/` is removed.** See **Model policy**
(`skills/myflow-contracts/pipeline-rationale.md`) for why, and **Run 1 — the branch is not merged**
(`skills/myflow-contracts/finish-contract.md`) for the preservation duty itself.

**A persisting record must not fill in `unknown (agent-defined)` on the way into the repository** —
no preservation step invents a model slug. See **Model policy**
(`skills/myflow-contracts/pipeline-rationale.md`) for why.

- **Claude Code**: the **session** model is enforced via `model: opus` / `model: sonnet` in each
  command's frontmatter (`commands-claude/*.md`) — no manual action needed *for the session*.
  Frontmatter cannot set a **subagent's** model, so the implementer rule above is not enforced by
  it: `/myflow-do` must name the model on each implementer dispatch, and the ledger line for that
  task is what records that it did. Slots dispatched by `subagent_type` (Bugbot, Security Review)
  carry their own agent definitions and take no override from either mechanism.
- **Cursor**: not enforceable yet (no per-command model frontmatter support as of this writing) —
  each `.cursor/commands/myflow-*.md` file carries an explicit note; switch models manually in the
  composer/chat picker.
- **Codex**: no per-command/skill model override mechanism either — model is a session or profile
  level setting; switch manually before starting a new proposal.

## Change name resolution (all `/myflow-*` commands)

`<name>` is **optional** on every `/myflow-*` command. When omitted, the candidate set is the
**union** of two sources, not `openspec list --json` alone:

- the non-archived names `openspec list --json` reports; and
- the basenames (minus `.json`) of every file directly under the project's state directory,
  `/Users/tweety53/Agents/myflow/state/<project-key>/*.json`, with `<project-key>` resolved exactly
  as **State file** (`skills/myflow-contracts/state-file.md`) already defines it — that file owns
  the formula and the bash that computes it, and neither is re-derived here.

From that union, drop any name whose `openspec/changes/<name>/` directory has already reached
`openspec/changes/archive/`. The state directory is per-project rather than per-worktree, so a
change's state file is reachable from the main checkout regardless of which worktree created it. See
**Change name resolution (all `/myflow-*` commands)** (`skills/myflow-contracts/pipeline-rationale.md`)
for why the second source (the state directory) is needed alongside `openspec list --json`.

**A state-directory file that cannot be parsed is reported and skipped from the union — never
silently dropped.** Name the unreadable file in the resolution's own output; do not fold it into a
"zero matches" or "no change" result as if it were never there.

Once the candidate set is built, resolution proceeds exactly as before:

- Exactly one match → use it automatically; announce which change was picked.
- Multiple matches → **AskUserQuestion** listing each (name, state, last modified) — never guess.
- Zero matches → fall back to that command's normal "no change" handling (e.g. `/myflow-start` asks
  what to build; others suggest the prior state's command).

This resolution is defined **once, here**, and every `/myflow-*` command's own enumeration step
cites this section rather than repeating or re-deriving the union — so no command's list of open
changes can drift from another's.

A change linked to a Jira issue is named `<lowercased-key>-<slug>` — see
**Change naming** in `skills/myflow-contracts/jira-integration.md`.
