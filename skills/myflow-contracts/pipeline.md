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
staged diff and a test guide are waiting for the human. Nothing records that the review or the
testing happened; the operator running the next command is what carries the change forward. This
is why no `*-done` command exists — there would be nothing for one to write.

| State | Means | Waiting on |
|-------|-------|-----------|
| `STARTED` | The proposal exists and is published | you — read the artifact |
| `IN_PROGRESS` | The implementation is staged and a test guide exists | you — review the diff, run the apps |
| `FINISHED` | Archived, pushed, worktrees removed | — |

**Reviewing and testing are one gate.** `/myflow-do` produces both surfaces in the same run, so
the human does both at one sitting. There is no state between implementation and finishing —
integration is not a stage, it is the first half of finishing.

## Pipeline flow

```mermaid
stateDiagram-v2
    [*] --> STARTED: /myflow-start
    STARTED --> STARTED: /myflow-start (revise the proposal)
    STARTED --> IN_PROGRESS: /myflow-do
    IN_PROGRESS --> IN_PROGRESS: /myflow-do (fix — never moves the state)
    IN_PROGRESS --> IN_PROGRESS: /myflow-finish (run 1 — integrate)
    IN_PROGRESS --> FINISHED: /myflow-finish (run 2 — after the merge)
    FINISHED --> [*]
```

### Level 1 — the stages of each command

**One row per command** — the five this pipeline has, three of them pipeline commands and two
read-only, exactly as **Command surface** (`pipeline.md`) below names them. A stage marked ▸ hides
substructure and is expanded at level 2 below. The gate column is the human gate that *follows* the
run — a property of the state the command ends in, never a stage of its own.

`/myflow-finish` is one command with two runs, so its row carries both, labelled: the preflight
verdict picks which one this invocation performs, and the run is never a command of its own.

| Command | Stages, in order | Gate after it |
|---------|------------------|---------------|
| `/myflow-start` | resolve the change → ask the planning effort and the three model choices *(creating run only)* → brainstorm ▸ → design approval → create the OpenSpec artifacts → writing-plans ▸ → publish the proposal artifact → write `STARTED` | you read the proposal artifact |
| `/myflow-do` | state gate → load context and validate the plan → isolate the workspace *(first run only)* → document the fix *(re-runs only)* → SDD + TDD per task ▸ → the review panel ▸ → write the manual test guide → validate the project's `## workspace isolation` section, then export what it declares → run the project's lint and test commands → stage, excluding the planning paths → write `IN_PROGRESS` | you review the staged diff **and** run the apps against the guide |
| `/myflow-finish` | the preflight verdict ▸, taken once per recorded worktree, decides which run follows — *run 1:* the unfinished-work gate ▸ → the landing question → preserve the session records → two commits, implementation first → the landing routes ▸ → move the issue to In Review → write `IN_PROGRESS`; *run 2:* verify the merge → sync delta specs and archive → commit and push the archive → cleanup ▸ → verify the cleanup → write `FINISHED` → self-review ▸ | after run 1, you wait for the branch to merge; after run 2, nothing — the state is terminal |
| `/myflow-status` | read-only — no stages, no state write; regenerates a handoff block when given a change name | — |
| `/myflow-info` | read-only — no stages, no state write; reads this file and explains the pipeline | — |

### Level 2 — the stages that hide substructure

#### Brainstorm — `/myflow-start`

superpowers:brainstorming runs its checklist in full and ends with the operator approving the
design, which is a **hard gate inside the command**: nothing is created under `openspec/changes/`
until that approval lands. The approved design is saved under `docs/superpowers/specs/` and is the
source for the change's OpenSpec design artifact — adapted, never duplicated into a conflicting
second design.

**The stage iterates rather than passing once.** After every planning-stage exchange — a round of
clarifying questions, the approval of a design section, the operator's review of the written spec —
one convergence test asks whether the command now holds a question its inputs do not answer, and
while it does, another round opens or is offered. The stage itself ends only on an explicit
operator answer — at the confirm, or by declining the offer, recording what is still open rather
than assuming it away — never on the command's own judgment, with one bounded exception: a session
that cannot ask at all records the confirm itself as an open question and ends the stage, since no
operator answer could ever arrive through it. An operator who is present but silent is not that
exception and still gets another round. One test after every exchange rather than a rule per gate is
the load-bearing part: a gate added later inherits the loop instead of escaping it by not being
enumerated.

The threshold, the two prompts, the bounded exception, and why their opposite recommendations are
both honest are **Convergence** (`skills/myflow-start/SKILL.md`), and are not restated here.

The planning level recorded on the creating run sizes the thinking *inside* this gate and never the
gate itself. The three levels and which of them is the default are owned by
**Planning effort** (`skills/myflow-contracts/state-file.md`) and are not restated here.

#### Writing-plans — `/myflow-start`

superpowers:writing-plans enriches `tasks.md` from a checkbox scaffold into a plan whose every item
carries exact paths, verification commands and no placeholders — the unit `/myflow-do`
dispatches one implementer against. Its self-review — spec coverage, placeholder scan, type
consistency — runs before the stage finishes.

Every fenced block and every numeric claim in a planning artifact carries a provenance tag,
`verified:<how>` or `unverified:`, and `scripts/check-plan-provenance.sh` is what makes that
mechanical rather than a habit. An unverifiable snippet is tagged and **kept**: a plan without the
snippet is worse than a plan carrying a labelled guess.

A revision round re-enters at this stage and republishes the proposal artifact to the **same** URL
rather than minting a second one.

#### SDD + TDD per task — `/myflow-do`

One implementer dispatch per checkbox in `tasks.md`, or per tightly coupled group, in plan order.
Every dispatch carries the same four required blocks — the no-commits boundary,
superpowers:test-driven-development as a required sub-skill, `engineering-principles.md` as required
reading, and the plan-provenance rule above — and names its model explicitly rather than inheriting
the parent's.

The task's diff is written to a file and the reviewer is given that path, never a commit range,
because nothing is committed at this stage. A checkbox is marked `[x]` only after its task passes
spec **and** quality review; a blocked task pauses and reports rather than guessing.

Which model a dispatch runs on, and the rule that every dispatch records it, are
**Model policy** (`pipeline.md`) below.

#### The review panel — `/myflow-do`

**Three required slots and four conditional ones.** Primary, Bugbot and Principles run on every
change; Security, Adversarial and the two extra principle lenses — B for simplicity and state, C for
robustness and ops — are selected from what the diff touches. Each selected slot is a **separate**
subagent with its own prompt, in every affected worktree; two slots are never merged into one.

**Every slot runs on the panel's model — Sonnet by default — except the two dispatched by
`subagent_type`.** Bugbot and Security Review carry their own agent definitions and take no model
override, which is also why their ledger entries read `unknown (agent-defined)`. There is no
parent-model inheritance and no economy tier: the panel's cost must not depend on the model the
operator happens to be running. That default, the change-level choice recorded against it, and the
operator override that may raise it are **Model policy** (`pipeline.md`) below, and are stated there
rather than a second time here.

**No handoff while any finding is open, at any severity.** A minor finding blocks exactly as a
critical one does. Every finding is recorded twice — as a row for the reader and as a marker line
for the guard — and `scripts/check-unfinished-work.sh` reads only the marker block.

**Re-runs are targeted by default and escalate to the full roster automatically**, without asking.
Escalation widens the panel's **breadth** — more lenses — and never its model, because the
implementers it would raise already sit at the ceiling. When a finding survives its last fix round
the run **hands back to the operator**, one finding at a time, with named options; only the
operator's answer may write a `withdrawn` marker, and only with the reason they give.

The tuned values are cited rather than copied. Which diff sizes and which touched areas select a
conditional slot is **Optional slot selection** in `skills/myflow-do/SKILL.md`; the conditions that
force a full re-run in place of a targeted one are **Panel re-runs** in `skills/myflow-do/SKILL.md`.

#### The preflight verdict — `/myflow-finish`

`scripts/check-finish-preflight.sh` decides which run happens, from three signals in a fixed order,
taken once per worktree recorded in the state file. It prints exactly one verdict line and exits 0
whenever it reached a verdict; **a missing verdict line is not a verdict**, and neither is a
worktree it cannot read. `RUN1` integrates, `RUN2` archives, and `REFUSE` **stops the run** and asks
the operator rather than guessing. Run 2 proceeds only when every recorded worktree returns `RUN2`.

The three signals and why their order is load-bearing are
**Finish contract** (`skills/myflow-contracts/finish-contract.md`).

#### The unfinished-work gate — `/myflow-finish` run 1

Runs **before** the landing question and before any git action, once per recorded worktree.
`scripts/check-unfinished-work.sh` returns `CLEAR` — go straight to the question, with no extra
prompt — or `OUTSTANDING`, which shows the breakdown and offers **exactly three** courses, with
**Stop** marked as the recommendation. There is no fourth course, and none that hands back to
`/myflow-do` inline.

The ordering is the point: an operator asked how to land a branch, and only then told it carries
unfinished work, has already answered a question about a branch they believed was complete. What
was integrated over is written into the planning commit's message and into the handoff, so the
record outlives the session.

Each course and what run 1 then does are
**Run 1 — the branch is not merged** (`skills/myflow-contracts/finish-contract.md`).

#### The landing routes — `/myflow-finish` run 1

The operator is asked once, before any git action, how the branch should land: open a pull request
*(default)*, merge and push, or handle it manually. The run then completes without asking again, and
the answer is never remembered between runs.

All three routes do the same two things first, in this order: preserve the session records out of
the gitignored worktree into the repository, then commit in **two** commits — implementation first,
planning artifacts second. The linked issue moves to In Review on every route, including the manual
one. Run 1 ends at `IN_PROGRESS` and names itself as the next command, because that is what the
operator runs once the branch is merged.

The route table is
**Run 1 — the branch is not merged** (`skills/myflow-contracts/finish-contract.md`); the guarded
two-commit chain every route uses is **Git boundaries** (`pipeline.md`) above.

#### Cleanup — `/myflow-finish` run 2

Every removal is *remove-or-move if present*, which is what makes run 2 re-entrant: a step whose
artifact is already gone is a success rather than an error, so a re-run after the operator clears a
leftover repeats the verification and nothing else.

The removals are verified rather than assumed. `scripts/check-cleanup-complete.sh` runs once per
repository, **after** all of them: `COMPLETE:` allows the `FINISHED` write, `LEFTOVER:` names what
remains and leaves the change at `IN_PROGRESS`, and a non-zero exit carrying no verdict line is
treated exactly as `LEFTOVER` — an unverified cleanup is not a verified one. A `COMPLETE:` line may
carry a `SKIPPED:` clause naming a row that was not verified, and that clause is relayed rather than
dropped — the rule is stated once in step 6 of
**Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`).

What is removed, when, and on what condition is
**Temporary artifacts registry** (`pipeline.md`) below — the one place a cleanup rule is stated. The
procedure for the rows it removes is
**Worktree cleanup** (`skills/myflow-contracts/finish-contract.md`).

#### Self-review — `/myflow-finish` run 2

Self-review runs only after `FINISHED` is written — never before it, and never in a run that stops
earlier. It is skippable per run, with running it the default. It gathers its input by invoking
`scripts/gather-self-review-context.sh` rather than having the reasoning pass re-read files inline,
runs **one** combined reasoning pass covering all four angles — problems and fixes, cost, what went
well, and automation candidates — together with the operator's 1-5 rating, never as four separate
dispatches, and offers a per-finding Jira filing ask before committing its report to
`docs/self-review/<name>-self-review.md`.

**Which file to change first.** The normative requirement is
**Requirement: Self-review runs only after FINISHED is written** (`openspec/specs/myflow-self-review/spec.md`),
read alongside the sibling requirements in that same file for context gathering, the combined pass,
the per-finding filing ask, the rating, and the report path. Naming the requirement in full, rather
than giving the path alone, is what makes `scripts/check-references.sh` check this pointer — an
OpenSpec `### Requirement: …` heading is a heading like any other. The guard skips a path that does
not resolve, so this one is checked only once the capability lands in `openspec/specs/` at finish
run 2; until then it is a reference nobody verifies, which is said here rather than left to look
otherwise, exactly as **Planning effort** (`skills/myflow-contracts/state-file.md`) already states
for its own forward reference.

## Command surface

Three pipeline commands and two read-only ones. **No command accepts a flag.** The only argument
is the optional change name — see **Change name resolution**.

An argument that is not a known change name is **reported**, not silently ignored — a silently
ignored word is indistinguishable from a flag that stopped working.

## State transitions

| Command | Accepts | Ends at |
|---------|---------|---------|
| `/myflow-start` | *(none — creates the change)* · `STARTED` | `STARTED` |
| `/myflow-do` | `STARTED` · `IN_PROGRESS` | `IN_PROGRESS` from `STARTED`; **otherwise unchanged** |
| `/myflow-finish` | `IN_PROGRESS` | `IN_PROGRESS` after run 1; `FINISHED` after run 2 |
| `/myflow-status`, `/myflow-info` | any — read-only, never block | unchanged |

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
| `/myflow-do` | from `STARTED` | Create branch/worktree + **`git add` excluding the planning paths** — no commits, push, merge, or PR |
| `/myflow-do` | at `IN_PROGRESS`, no `prUrl` | Resume **existing** worktree + **`git add` excluding the planning paths** — no commits |
| `/myflow-do` | at `IN_PROGRESS`, `prUrl` recorded | **Commits twice and pushes** to the PR branch — implementation, then planning artifacts; the one exception |
| `/myflow-finish` | run 1 | **Commits twice** — implementation, then planning artifacts — and pushes; opens a PR or merges, by the operator's choice |
| `/myflow-finish` | run 2 | **Commits and pushes the archive**; removes worktrees and branches |
| `/myflow-status`, `/myflow-info` | — | None — read-only |

**The planning paths** are the three that
**Handoff output** (`pipeline.md`) names below. `/myflow-do` clears them from the index and only
then stages with them excluded by pathspec — an exclusion governs what an
`add` adds and cannot retract what an earlier step staged, so the clearing pass is what makes the
rule hold rather than merely assert it. Its staging area therefore carries implementation only, and
`/myflow-finish` is what commits them.

`git add -A` respects `.gitignore`. Never force-add.

**Both commits are guarded, and an empty one is skipped rather than failed.** `git commit` exits
non-zero when nothing is staged, so an unguarded two-commit sequence dead-ends on three ordinary
cases: a fix touching only the three planning paths leaves the implementation commit empty — which
is exactly what run 1's unfinished-work **Stop** course invites — a fix touching only implementation
leaves the planning commit empty, and a re-run after a rejected push finds both commits already
made. Each commit is therefore preceded by a staged-changes test, and the whole sequence is one
`&&` chain:

```bash
git -C <abs-worktree> reset -q -- openspec/ docs/manual-test/ docs/superpowers/ \
  && git -C <abs-worktree> add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/' \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "<type>(<name>): <what the implementation does>"; } \
  && git -C <abs-worktree> add -A \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "chore(<name>): plan, test guide and session records"; }
```

**A skipped commit is reported, and a FAILED commit stops the sequence.** Those are different
outcomes: "nothing to commit" is normal and costs one line in the handoff, while a commit a hook
rejects is a git failure and gets the standard treatment — report git's own output and stop. The
chain is what enforces the second, and it is a chain rather than `set -e` deliberately: bash before
4.0 ignores `set -e` inside a subshell whose parent has errexit off, and this block runs through an
agent's shell whose state it does not control. Run it as one command. Without that, a first commit
a hook rejects falls through to the unconstrained second `add`, and the sole resulting commit —
titled `chore(...)` — carries the implementation, silently breaking the very split this section
exists to enforce.

**A planning path that is a tracked symlink stops the run, and is never worked around.** When any of
the three is a symlink, `git add -A -- . ':(exclude)docs/superpowers/'` exits 128 with
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

**A non-zero exit is never silent and never a stop.** Those two rules pull in opposite directions
and both hold: a preservation step able to abandon an integration whose work is already committed
would be a worse failure than the missing record, *and* a refused write that passed unmentioned
would leave the operator believing a record exists when none does. So it is reported and the run
continues — and the handoff says which records were preserved and which were not. The remaining
sources are still attempted after any one failure, so a single bad path costs one record, not three.

**That ordering is deliberate, and deliberately differs from `/myflow-do`'s.** Here the preservation
call comes *before* any staging, because all three routes commit: the second, unconstrained `add`
covers the preserved files wherever in the run they were written, and placing the call first keeps it
out of the route branches, where it could be forgotten on one of them. `/myflow-do` runs the same
script *after* its unconditional staging and stages a second time, because it stages on every run
and commits only when a `prUrl` is already recorded; hoisting the call there would create
`docs/superpowers/` files on every ordinary staged-only run.
**Do not harmonise the two orderings for symmetry** — the asymmetry is what keeps preserved records
out of the staged-only path.

## Progress visibility

**Every pipeline command drives the harness's task-list mechanism.** `/myflow-start`, `/myflow-do`
and `/myflow-finish` register their steps with it at the start of a run and keep each entry's status
current — in progress when its step begins, completed when that step finishes — so the harness's
live progress view, a count line and one line per task, renders throughout the run rather than
arriving with the handoff. The count line then distinguishes done, in progress and open at every
point.

| Command | One entry per |
|---------|---------------|
| `/myflow-start` | its brainstorming checklist item and each artifact it produces |
| `/myflow-do` | each item in `tasks.md`, in plan order |
| `/myflow-finish` | each step of the run it is performing — run 1's steps or run 2's, never both |

`/myflow-status` and `/myflow-info` are read-only and **register nothing**. Registering steps for a
report would put entries on the operator's task list for work nobody is doing.

**The progress view is a view, never a record.** No command, guard or contract reads the harness's
task list back as evidence of what was done. `tasks.md` remains the single source of truth for a
plan's completion state, and `scripts/check-unfinished-work.sh` reads that file — a second source of
completion state would be one that guard cannot see.

**No third checkbox marker is added to `tasks.md`** to carry an in-progress state. A marker written
at dispatch and resolved at completion would survive a crashed run as a permanently in-progress
task, in a file two guards parse. The in-progress count comes from the harness's task list alone,
which no run persists.

**Stated against the mechanism, never against one harness's tool.** myflow runs in Claude Code,
Cursor and Codex, and a rule written against one harness's API is unimplementable in the other two.
Where a harness offers no task-list mechanism, the command prints the equivalent block in its output
instead: a count line naming how many steps are done, in progress and open, followed by one line per
step marked done or not done. The rule is satisfied by whichever mechanism the harness provides, and
no harness has to gain a task tool to satisfy it.

## Handoff output

Every command ends in the same shape, and prints **nothing** after it:

```
<1–3 lines: what actually happened>

<absolute paths to anything the operator needs to open>

Next:
/myflow-finish <name>
```

- **The next command is the last line** — bare, copy-pasteable, with no prose after it. An agent
  cannot drive a harness's autocomplete; nothing lets a running session prefill the operator's
  input box. The last-line convention plus a five-command surface is the whole mechanism.
- **`/myflow-finish` run 1 names itself** as the next command, because that is what the operator
  runs once the branch is merged. Only a run 2 that **completed** is terminal and names nothing — a
  run 2 that stopped on a cleanup leftover names itself too, for the same reason: the operator
  clears what remains and runs it again.
- **Only what the operator must act on.** Do not restate the plan, enumerate completed internal
  steps, or repeat content available at a path you just gave.
- **Link, never paste.** Manual test guides, diffs and plans are given as absolute paths.
- **Every path is absolute** — in handoffs, in generated guides, in IntelliJ commands, in run
  instructions. Never a relative path, never `../<other-app>`, and never a main-checkout path while
  an apply worktree holds the work. Resolve app roots from `git worktree list` or the state file's
  `worktrees` keys.
- **`/myflow-do` never stages `openspec/`, `docs/manual-test/` or `docs/superpowers/` before
  finish.** The plan was read at `STARTED`; presenting it again as code to review hides the
  implementation diff it is mixed into. Leaving them unstaged, rather than filtering them out of one
  display command, is what makes them absent from *every* view of the staging area — a filtered
  display leaves them in the index, where `git status`, a graphical client and the IDE's
  staged-changes pane show them again. The list is fixed here rather than configured per project; the
  pipeline chooses these paths
  itself, so no project can differ. `/myflow-finish` run 1 stages them and commits them separately
  from the implementation, so nothing is lost:

  ```bash
  git -C <abs-worktree> diff --cached --stat
  git -C <abs-worktree> diff --cached
  ```

### The block each state renders

The block a state hands off is defined **here and nowhere else**, as one template per rendering. Two
commands render it: the command that ends in that state, and `/myflow-status <name>`. **No command
stores the emitted text.** `/myflow-status <name>` regenerates the block from the state file and the
artifacts as they now stand; it never reads back a stored copy, because a stored copy reproduces the
original exactly and then goes wrong silently the moment anything it names moves — a worktree
removed, an artifact republished, a PR opened. Regeneration is the mechanism, not an implementation
detail of it.

**The template is the definition, and it carries what the commands print.** A field a command emits
and the template omits is drift the moment `/myflow-status` renders the same state: the two blocks
would differ in the one place this section exists to keep identical. So each template below carries
every field its command emits, under that command's own label and in its `**Label:**` style, and a
field added to a command is added here in the same change. Labels are the part that must match; the
`<…>` placeholders describe each value rather than reproducing the alternatives a command writes.

**"Here and nowhere else" is a duty on the producing skills, not a claim about them.** Each of the
three — `skills/myflow-start/SKILL.md`, `skills/myflow-do/SKILL.md` and
`skills/myflow-finish/SKILL.md` — carries the block it prints, and each **cites this section as the
definition** at that block. A block sitting in a skill with no citation is a second, independently
authored definition however faithfully it happens to match today, and it is exactly how the two
copies drift: nothing tells the next editor of the skill that this file exists. The citation is what
turns three copies into one definition and three renderings of it.

**What a skill's block may differ in, and what it may not.** The **label set and the field set are
identical** — same labels, same order, no field in one and not the other; that is the part
`/myflow-status` has to reproduce. What a skill *may* do is **enumerate**, in place of a
placeholder, the literal alternatives its own command writes: this file describes a value's space
across both renderers, while a skill states what that one command emits. An enumeration is
therefore a refinement of the placeholder beside it and must stay **inside** it — a skill that lists
fewer cases than the placeholder describes is narrowing the contract, and teaches a reader to drop a
case the rule above requires. The missing rule and the run-only rule below bind both files equally.

**A value the state file does not carry is reported as missing, not dropped.** A block whose
artifact URL reads *missing* is distinguishable from one whose URL was never printed at all; a
silently absent line is not.

**A run-only field is not a missing one, and the two are never collapsed.** *Missing* means the
block could have carried the value and did not — the state file has the key and it is `null`.
**Run-only** means no regenerated block can ever carry it, because nothing on disk holds it: it
exists only inside the run that emitted it. `(run-only)` immediately after a label in a template
below marks such a field — the command ending in the state prints it, and `/myflow-status` omits it
rather than reporting it missing. Reporting a run-only field as *missing* would tell the operator
something is wrong when nothing is, which is the opposite of what the rule above buys. The marker
annotates the template and is never text a command prints.

Only the form carrying a change name regenerates a block. `/myflow-status` with no argument stays
the table it is today. Regenerating performs no action named in the block it prints — the command is
read-only in both forms.

**`STARTED`** — printed by `/myflow-start`, regenerated by `/myflow-status <name>`

```text
## Proposal ready — review required

**Change:** <name>
**Artifact:** <artifactUrl, or "missing">
**Decisions recorded:** <count, or "none">
**Open questions:** <count, or "none">
**Jira:** (run-only) <issue key and the transition made, or "none linked", or a skipped-with-reason line>
**Jira description (pre-edit):** (run-only) <the text as it stood before the write, verbatim in a fenced block>
**Planning effort:** <the level in force, or "not recorded — planned at default">
**Models:** implementation <model, or "not recorded">, review panel <…>, panel fixes <…>

Open in IntelliJ:
open -na "IntelliJ IDEA" --args "<absolute main-checkout path>"

<what the operator does next>

Next:
/myflow-do <name>
```

**Why the open-questions line is not run-only, and carries no marker.** It is derived from an
artifact on disk — the entries under `## Open questions` in the change's design whose status is
still `open` — exactly as the decisions count above it is, so `/myflow-status <name>` regenerates it
rather than omitting it. The two lines sit next to the `Jira` line and are the opposite case to it:
what makes that one run-only is that nothing on disk holds it, and that test is about where the
value lives, not about how close it sits to a line that failed it. A count that has changed since
`/myflow-start` printed it — a revision round answered a question and moved the entry to
`answered by <decision-id>` — is this field working: the line reports what is open now, not what was
open then. It reads `none` when nothing is open, by the missing-rather-than-dropped rule. The
entry shape, the immutable ID and the never-delete rule the count reads through are stated once
under **Open questions** (`skills/myflow-start/SKILL.md`) and are not repeated here.

**`IN_PROGRESS`, after `/myflow-do`** — printed by `/myflow-do`, regenerated by `/myflow-status <name>`

```text
## Implementation staged — review and test

**Change:** <name>
**Panel:** (run-only) <the required slots, and the optional ones selected or "none — no triggers fired">
**Progress:** <completed>/<total> tasks
**Git:** <staged and uncommitted, committed and pushed to the PR branch, or committed and pushed with no PR — run 1 merged it or handed it over>
**Jira description (pre-edit):** (run-only) <the text as it stood before the write, verbatim in a fenced block>

Worktree:   <absolute worktree path>
Test guide: <absolute path to the guide, or "missing">

Review the diff, then run the apps against the guide:
  <the review command that matches the Git line — see below>
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"

<what the operator does next>

Next:
/myflow-finish <name>
```

**The `Git` line has a third option, and the review command follows it.** `/myflow-do` itself only
ever emits the first two — it stages, or it commits and pushes to a PR branch. The third is reached
only when `/myflow-status` regenerates this rendering for a change whose run 1 took the *merge and
push* or *handle it manually* route: the work is committed and pushed with no `prUrl` to prove it.
Leaving the field at two options meant that block stated something untrue about every such change,
and the pair below is what the command prints instead:

| `Git` line | Review command |
|------------|----------------|
| staged and uncommitted | `git -C <absolute worktree path> diff --cached` |
| committed and pushed to the PR branch | `git -C <absolute worktree path> diff <merge base>..HEAD` |
| committed and pushed with no PR | `git -C <absolute worktree path> diff <merge base>..HEAD` |

**A committed branch has an empty staged diff, so `--cached` is not a sparse answer there — it is a
wrong one.** It exits 0 printing nothing, which reads to an operator as "there is nothing to
review". The merge base is the value already recorded against that worktree's key in the state
file's `worktrees` map, so no new field is needed and nothing has to be inferred; a worktree with no
recorded merge base falls back to the staged-diff line and says the range could not be resolved.

**`IN_PROGRESS`, after `/myflow-finish` run 1** — printed by run 1, regenerated by `/myflow-status <name>`

```text
## Branch integrated — <waiting on the merge, or merged and waiting on run 2>

**Change:** <name>
**Route:** (run-only) <pull request, merged and pushed, or manual>
**PR:** <prUrl, or why there is none on the route taken>
**Outstanding:** (run-only) <what the unfinished-work gate reported and the operator integrated over, or "none">

<what the operator must do before the next run>

Next:
/myflow-finish <name>
```

**Which rendering `/myflow-status` regenerates.** **Merge status decides it whenever the merge
status is known**, and the command already has that answer: it runs the merge-status test in its own
step 2 to fill the next-command column. A branch proven to have reached the base branch **is**
integrated, so the run-1 rendering is the correct one for it, whatever `prUrl` says. `prUrl` is the
tiebreaker only where the stronger signal is genuinely absent:

| Merge status | `prUrl` | Rendering |
|--------------|---------|-----------|
| **no commits of its own** — `HEAD` is still the recorded merge base | either | `/myflow-do` |
| merged (proven) | either | run 1 — and its heading reads *merged and waiting on run 2* |
| not merged (proven) | either | run 1 — *waiting on the merge* |
| inconclusive | recorded | run 1 — *waiting on the merge* |
| inconclusive | none | `/myflow-do` |

**The first row is a pre-check, not a special case, and it must be answered before the ancestor
test.** `git merge-base --is-ancestor HEAD <base>` returns **true** for a branch carrying no commits
of its own — every commit it has, the base branch already had — and that is the *ordinary*
`IN_PROGRESS` shape, because `/myflow-do` stages without committing. Without the pre-check every
change that has never been through `/myflow-finish` reads as merged and is shown *merged and waiting
on run 2*. The test is `HEAD` against the merge base **already recorded for that worktree** in the
state file's `worktrees` map: equal means the branch has no commits of its own and is therefore not
merged, whatever the ancestor test then says. `scripts/check-finish-preflight.sh` documents this
trap and guards it in exactly that order — its comment (b), on why the recorded-merge-base check
must run before the ancestor test — and the reasoning is not re-derived here.

**The recorded merge base has three conditions, not two, and two of them are `inconclusive`.**
Recorded and resolving is the ordinary case, and the pre-check above answers it. **Absent** is the
plain unknown: the pre-check cannot be performed, so the ancestor test alone cannot be trusted.
**Recorded but unresolvable** — `git rev-parse --verify` cannot turn the stored sha into a commit,
because history was rewritten, the clone is shallow, or the object was pruned — is the same unknown
wearing a value,
and it is the one that gets missed: compared as a *string* it is merely "not equal to `HEAD`", which
reads as "the branch has commits of its own" and falls straight through to the bare ancestor test,
reporting *merged* for a branch that has never been through `/myflow-finish`. Both unknowns put the
change on the two `inconclusive` rows, where `prUrl` is the tiebreaker. That is the same
refusal-to-infer `check-finish-preflight.sh` makes twice over — when it is handed `-` for the
recorded merge base, and when `rev-parse --verify` on a recorded one fails — and the pre-check is
therefore *resolve, then compare*, never compare alone.

**`FINISHED`** has **no regenerated block**: the state is terminal and finished changes are omitted
from the report, so there is nothing left waiting on the operator to hand off. `/myflow-finish`
run 2 does print a terminal block — what it synced, archived, removed and verified — and every field
of it is run-only, because it reports what that run did rather than what the change now is. One
renderer means nothing to keep in step, which is why that block takes no template here. That block
carries one more field now: `**Self-review:** <path> (rating: <n>/5) | skipped`, immediately after
`**Cleanup:** verified`, naming step 8's outcome — a value only run 2 ever has, exactly like the
fields beside it. A run 2 that **stops** on a cleanup leftover is not this case: it leaves the change
at `IN_PROGRESS` and prints its own interrupted-run report, every field of which is likewise
run-only — what that run synced, archived and left behind, which the state file does not record.
That interrupted-run report carries no `Self-review` field, because it is printed only when run 2
stops **before** step 7 — step 8 never runs there, so there is nothing for the field to name; adding
it regardless would misstate a run that never reached self-review. `/myflow-status` regenerates one
of the two `IN_PROGRESS` renderings above for such a change, by the test just given.

Which path each `open` line names, and why `open -na` rather than the `idea` shim, are
**IntelliJ commands** (`pipeline.md`) below.

### The tab commands, printed at the start of a run

`/myflow-start`, `/myflow-do` and `/myflow-finish` each print, immediately after their announcement
line and before any work, two commands for the operator to paste:

```text
/rename <change-name>
/color cyan
```

They sit at the **start** of the run rather than in the block above because labelling a tab is
useful before a long run rather than after it; the rules above govern what a command prints when it
*ends*, and these lines are not part of a handoff. The colour is one fixed value for every command
and every change — `cyan`, chosen over `red`, `yellow` and `orange` because those already read as
error and warning states in this pipeline's output — and it signifies only that a pipeline command
owns the tab. `/myflow-status` and `/myflow-info` print neither line: a read-only report does not
own the tab.

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
| `IN_PROGRESS` | apply worktree root, plus the test guide's absolute path |

Paths are absolute, resolved from `git worktree list`. Never emit a relative path.

## Finish contract

**Finish contract** (`skills/myflow-contracts/finish-contract.md`) governs the preflight signals,
both runs' procedures, base-branch resolution and worktree cleanup, and `/myflow-finish` is the
only command that loads it.

## Resolving a change's worktrees

The scan that finds the worktrees carrying a change's branch. `/myflow-finish` uses it during
cleanup, and `skills/myflow-contracts/state-self-heal.md` uses it to rebuild a state file's
`worktrees` map — and no command that self-heals a state file loads
`skills/myflow-contracts/finish-contract.md`, which is why this lives here.

The set of worktrees is the **keys of the state file's `worktrees` map**. When that is absent, scan
each affected repository:

```bash
git -C "$REPO" worktree list --porcelain \
  | awk '/^worktree /{w=substr($0, 10)} /^branch /{if ($2=="refs/heads/openspec/<name>") print w}'
```

**Never guess a path.** Worktree layout differs per repository — this repo keeps its worktrees in a
sibling directory, not `.worktrees/`.

**The path is taken with `substr`, never `$2`.** `worktree list --porcelain` emits it raw, so a
field reference truncates any path containing a space at the first one: fed
`worktree /tmp/my worktree/x` it yields `/tmp/my`, and the run then `--force`-removes a path that is
not the worktree, or fails having named the wrong one. `10` is one past the length of the literal
`worktree ` prefix. The branch on the next line is a ref name and cannot contain a space, so `$2` is
right for it. `scripts/check-cleanup-complete.sh` parses the same stream the same way — the guard
and the snippet it verifies must not disagree, or the wrong one gets copied next.

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
points here rather than restating it, in this file and in every skill. **Worktree cleanup**
(`skills/myflow-contracts/finish-contract.md`) is the *procedure* for the rows removed there and
not a second statement of the rule: the table says what is removed and when, that section says how
— and a stale second copy of a rule governing `git worktree remove --force` is a copy that deletes
the wrong thing.

**An artifact no row accounts for is a defect in the registry**, corrected by adding the row. It is
never left unaccounted for on the grounds that something probably removes it — that assumption is
exactly how the remote branch went unremoved until it was given a row here.

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

**This is the one row whose removal is verified by asking rather than by looking**, and the reason
is that "ran the removal" is not "verified gone": a removal that reported success against a stale
connection leaves this row's promise broken with nothing having failed. So a survivor is established
from the project's own survivor report, never inferred from the removal's exit code — stated once
under **Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`), with the report's
output and exit-code contract under
**Project configuration** (`skills/myflow-contracts/project-configuration.md`). A report that could
not reach its service is skipped rather than failed, so this is the one row a stopped service leaves
unverified without stranding an already-merged change; that asymmetry is likewise
**Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`).

**Nothing removes the claimed cache index, and nothing in this pipeline can — which is why the row
says so instead of naming a remover it does not have.** The rule one paragraph above is that an
artifact no row accounts for is a defect in the registry; the row exists for that reason, and an
honest `Removed by` cell is the whole of what it buys. The index is the one workspace value that is
**not** a function of the change name — it is claimed by probing, for the reasons under
**The cache index** (`skills/myflow-contracts/workspace-isolation.md`) — and it is not written into
the state file either. So by the time run 2 runs, nothing on the machine can say which index this
change held: there is no derivation to repeat and no record to read, and a run that swept an index
it guessed would flush another workspace's. The project's `remove` command does not touch it for the
same reason, which
**Project configuration** (`skills/myflow-contracts/project-configuration.md`) states as a property
of the `cache index` resource word.

## State file

The contract governing where a change's state file lives, its full JSON shape, monotonic state
writes, and carry-forward rules. **State file** (`skills/myflow-contracts/state-file.md`) — load it
before reading or writing a state file.

## State self-heal

The contract governing how a state file is validated against on-disk artifacts, and how a missing,
unparseable, or contradicted file is corrected.
**State self-heal** (`skills/myflow-contracts/state-self-heal.md`) — load it before self-healing a
state file.

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

**This section is canonical for the model roles, their defaults and how an override applies.** One
location, named here rather than left to be worked out: every `/myflow-*` command is required to
load this file before acting, and none of them loads `openspec/specs/`, so the file runtime actually
reads is the file the rule has to live in. **State file** (`skills/myflow-contracts/state-file.md`)
cites this section for the `models` field rather than defining the roles a second time, and
`CLAUDE.md` and `AGENTS.md` name this section for the same reason.

**Which file to change first.** The normative requirements behind this section belong to the
OpenSpec capability `myflow-model-policy`, whose **Requirement: Implementer subagents run on the strongest available model** (`openspec/specs/myflow-model-policy/spec.md`) anchors the defaults below. That capability is the requirement; this section is the **operational form the commands read**, and the two-layer split is the same one **Planning effort** (`skills/myflow-contracts/state-file.md`) already uses. Change the capability first and bring this section with it: a section that contradicts the requirement is this file's defect, not the spec's. A live spec is also behind by construction while a change is open — its delta lands in `openspec/specs/` only at finish run 2 — which is the second reason runtime reads this section rather than that file.

`/myflow-start` should run on **Opus** (or the harness's strongest available model) — brainstorming
and design benefit most from stronger reasoning. Every other `/myflow-*` command should run on
**Sonnet** (or the harness's standard default), and **every review-panel reviewer runs on the
panel's model — Sonnet by default** — regardless of the parent model. Sonnet is the default rather
than an absolute because a change may record its own panel model, per the three roles below; what
never varies is that the panel's model is *chosen*, not inherited from the parent session.

**Implementer subagents dispatched by `/myflow-do` run on Opus** (or the harness's strongest
available model). This **explicitly overrides** superpowers:subagent-driven-development's model
guidance — that skill says to pick "the least powerful model that can handle each role" and to use
the cheapest tier where the plan already contains the code to write. That guidance does not govern
this pipeline. The saving it offers is false here: an implementation defect is not avoided by the
review panel, it is *found* by the panel, at the cost of a fix wave and a re-run of every slot.
Buying a cheaper implementer with a more expensive review is the wrong trade.

**Two further instructions in that same upstream skill are also overridden, and are named here
rather than left to be discovered.** subagent-driven-development says to dispatch the *final
review* on the most capable model: myflow does not — it fixes every panel slot at the panel's
model, Sonnet by default, for the reason above, and escalates the panel's **breadth** instead (the conditional Security, Adversarial
and extra-principle-lens slots), which buys more independent readings rather than one stronger one.
It also says to *escalate the model in fix rounds 4-5*: myflow cannot, because its implementers
already sit at the ceiling from round 1. Fix rounds escalate the same way — more lenses, not a
bigger model — and round 5 hands back to the operator rather than pretending an escalation is
available.

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

**The panel-fix default is the strongest available model, and deliberately not Sonnet.** The role
names the agent that *applies* a fix, which is an implementer — so the implementer rule above
already governs it. Fix rounds escalate the panel's breadth rather than its model precisely because
implementers sit at the ceiling from round 1, and a fix-wave default of Sonnet would contradict both
of those rules at once.

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
A model policy that nothing records is a policy nothing can verify — and the absence of that record
is precisely how this rule came to be missing in the first place.

Where the dispatcher **cannot know** the model, the ledger records `unknown (agent-defined)` and
never a guess. Slots dispatched by `subagent_type` (Bugbot, Security Review) resolve their model
from their own agent definition, which the dispatcher does not read; writing a plausible slug for
them puts an unmeasured value into the audit trail.

**This record outlives the change.** The ledger is authored under `.superpowers/`, which is
gitignored, in a worktree `/myflow-finish` run 2 removes — but run 1 preserves it into the
repository first, under `docs/superpowers/ledgers/`, so it serves the operator and the panel
*during* the change and stays answerable afterwards. An after-the-fact audit of which model
implemented which task therefore reads the preserved ledger rather than a transcript nobody kept.
The preservation duty itself is stated once, under
**Run 1 — the branch is not merged** (`skills/myflow-contracts/finish-contract.md`); this section
depends on it rather than restating it.

Durability is a **stronger** reason to leave an unobserved entry unobserved, not a weaker one. A
persisting record makes an invented model slug permanent, so `unknown (agent-defined)` stays exactly
as written above and no step fills it in on the way into the repository.

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
`openspec/changes/archive/`. This is why the second source matters: `openspec list --json` only
sees change directories present in the *current* git checkout, so a change staged in a worktree —
`openspec/changes/<name>/` created there but never committed to the main checkout — is invisible to
it alone, even while it sits at a human gate with a fully staged diff. The state directory is
per-project rather than per-worktree, so a change's state file is reachable from the main checkout
regardless of which worktree created it.

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
