# myflow pipeline — rationale

This file is the reasoning behind `skills/myflow-contracts/pipeline.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

The finish contract lives in `skills/myflow-contracts/finish-contract.md`, which is canonical for it.

## States

## Stage exit — never the command's own judgment

The tuned threshold, the two prompts, and why their opposite recommendations are not to be
harmonised belong to the command itself — **Convergence** (`skills/myflow-start/SKILL.md`) — and are
not restated here.

## Command surface

Behaviour a flag used to select is now either asked at invocation (the integration choice in
`/myflow-finish`), derived from the current state, or fixed at the single sensible default
(review-panel breadth, decided by its escalation triggers).

## State transitions

### Every command is re-entrant

### A fix never moves the state

## Wrong state for this command

## Git boundaries

**Why `/myflow-do` commits once a PR exists and not before.** A PR is a remote surface, so a
staged-only fix would be invisible on it. Before that the operator reviews a staged diff in the
IDE, and committing would take that away.

**Both commits are guarded, and an empty one is skipped rather than failed.** `git commit` exits
non-zero when nothing is staged, so an unguarded two-commit sequence dead-ends on three ordinary
cases: a fix touching only the two planning paths leaves the implementation commit empty — which
is exactly what run 1's unfinished-work **Stop** course invites — a fix touching only implementation
leaves the planning commit empty, and a re-run after a rejected push finds both commits already
made. Each commit is therefore preceded by a staged-changes test, and the whole sequence is one
`&&` chain:

**A skipped commit is reported, and a FAILED commit stops the sequence.** Those are different
outcomes: "nothing to commit" is normal and costs one line in the handoff, while a commit a hook
rejects is a git failure and gets the standard treatment — report git's own output and stop. The
chain is what enforces the second, and it is a chain rather than `set -e` deliberately: bash before
4.0 ignores `set -e` inside a subshell whose parent has errexit off, and this block runs through an
agent's shell whose state it does not control. Run it as one command. Without that, a first commit
a hook rejects falls through to the unconstrained second `add`, and the sole resulting commit —
titled `chore(...)` — carries the implementation, silently breaking the very split this section
exists to enforce.

## Preserving the session records

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
`<project>/docs/superpowers/` files on every ordinary staged-only run.
**Do not harmonise the two orderings for symmetry** — the asymmetry is what keeps preserved records
out of the staged-only path.

## Progress visibility

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

### The block each state renders

The block a state hands off is defined in **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`), which is canonical for the three per-state templates,
the run-only rule and the rendering-selection table.

- **The next command is the last line** — bare, copy-pasteable, with no prose after it. An agent
  cannot drive a harness's autocomplete; nothing lets a running session prefill the operator's
  input box. The last-line convention plus a four-command surface is the whole mechanism.

- **`/myflow-do` never stages `<project>/openspec/` or `<project>/docs/superpowers/` before
  finish.** The plan was read at `STARTED`; presenting it again as code to review hides the
  implementation diff it is mixed into. Leaving them unstaged, rather than filtering them out of one
  display command, is what makes them absent from *every* view of the staging area — a filtered
  display leaves them in the index, where `git status`, a graphical client and the IDE's
  staged-changes pane show them again. The list is fixed here rather than configured per project; the
  pipeline chooses these paths
  itself, so no project can differ. `/myflow-finish` run 1 stages them and commits them separately
  from the implementation, so nothing is lost.

### The tab commands, printed at the start of a run

They sit at the **start** of the run rather than in the block because labelling a tab is
useful before a long run rather than after it; the rules govern what a command prints when it
*ends*, and these lines are not part of a handoff. The colour is one fixed value for every command
and every change — `cyan`, chosen over `red`, `yellow` and `orange` because those already read as
error and warning states in this pipeline's output — and it signifies only that a pipeline command
owns the tab. `/myflow-status` prints neither line: a read-only report does not
own the tab.

**Both of those facts are Claude Code's, and the rule is stated against the mechanism rather than
against them** — for the reason **Progress visibility** (`pipeline.md`) gives, which answers
the identical question for the task list and is not restated here. What every harness can do is
**print two lines of text**, which is why printing is the rule and the measurement is only the
reason invoking is not. So:

Both commands are real, and both routes to calling them are closed. That was established by measurement,
and it is recorded here so the next reader neither repeats the investigation nor treats the printing
as an oversight to correct:

- the harness's `SlashCommand` tool exposes only commands of `type: "prompt"`, while `/rename` and
  `/color` are `type: "local"` or `"local-jsx"` — so the tool route is closed; and
- no writable `/dev/tty` is available to a command — so writing the terminal escape sequence
  directly is closed too.

## IntelliJ commands

## Finish contract

### Run 1 — the branch is not merged

**The gate precedes the question, and that ordering is the point.** An operator asked how to land a
branch, and only then told it carries unfinished work, has already answered a question about a
branch they believed was complete — and the cheapest of the three courses to take by mistake is the
one that integrates.

**Implementation first is the order, not an accident of it.** The newest commit is the one a forge
shows first, and that should be the code; and the second commit's message is where the outstanding
list from the gate above is written down.

**That script has three outcomes, and they are not interchangeable.** This is where they are
defined; the two call sites point here rather than each describing them.

### Run 2 — the branch is merged

   **The removal goes after the worktree half, and the order is load-bearing.** Worktree cleanup's
   check 5 runs the project's `## stop` command, and it is the **only** place run 2 stops the stack —
   which run 2 nearly always has to do, because `/myflow-do`'s handoff prints the run instructions
   and the operator runs the applications against them, so a stack still up when run 2 starts is the
   common case rather than a rare one. Dropping a database the project's own stack still holds open is a
   removal that fails on the ordinary path, so the removal waits until that stack is down.
   Nothing pulls the other way: these resources live in the project's shared data services rather
   than in the worktree, so taking the worktree down neither removes them nor puts them out of reach,
   and `remove` runs from the main checkout — stated with the command table, in
   **Project configuration** (`skills/myflow-contracts/project-configuration.md`) — so no worktree
   removal can destroy the directory it runs from. What has always constrained this step still does:
   the removal stays ahead of step 6, because a verification that runs before the thing it verifies
   can only ever fail.

   **Interleaving the removal into the worktree half was considered and rejected.** Slotting it
   between that half's checks and its destructive commands would put the stack down first just as
   well, but it would bind this step's order to another section's internal check numbering, and it
   would run a once-per-change removal inside a per-worktree loop — two couplings bought for a
   property the placement above already has.

   **The cost of this order is that a worktree half which stops early takes the removal with it.**
   Any failed check leaves every worktree alone —
   **Worktree cleanup** (`skills/myflow-contracts/finish-contract.md`) — and the removal behind it
   does not run, so a run blocked by something unrelated to the workspace, an
   uncommitted file in a worktree say, has its database and bucket named as leftovers at step 6 as
   well, with nothing wrong with either. That is the right cost to accept. It lands on a run that has
   already stopped and already needs the operator, it adds lines to a report rather than a failure,
   and run 2 is re-entrant, so the pass after the blocker is cleared does both halves. The check most
   likely to fail is check 5 — a stack that will not stop — and that is exactly the condition under
   which the removal would have failed anyway, which is also why it is not attempted regardless:
   running with the stack down is the guarantee this order buys, and removing after a failed check
   would spend it. The order this replaces put its cost on the ordinary path instead.

   **Both halves share one numbered step deliberately.** They are one act — undoing what this
   change's run created — with an order between them that has to hold, and giving the removal a
   number of its own would renumber steps 6, 7 and 8, which are cited *by number* from
   `<agents repo>/openspec/specs/` and from `skills/myflow-finish/SKILL.md`.

### Worktree cleanup

## Resolving a change's worktrees

**Resolving a change's worktrees** (`skills/myflow-contracts/finish-contract.md`) is canonical for
`/myflow-finish`'s own scan-and-resolve procedure and is the fullest example of applying this rule.

## Temporary artifacts registry

**This table is the one place a cleanup rule is stated.** Everything else that mentions a removal
points here rather than restating it, in this file and in every skill. **Worktree cleanup**
(`skills/myflow-contracts/finish-contract.md`) is the *procedure* for the rows removed there and
not a second statement of the rule: the table says what is removed and when, that section says how
— and a stale second copy of a rule governing `git worktree remove --force` is a copy that deletes
the wrong thing.

**An artifact no row accounts for is a defect in the registry**, corrected by adding the row. It is
never left unaccounted for on the grounds that something probably removes it — that assumption is
exactly how the remote branch went unremoved until it was given a row here.

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

**What that leaves behind is bounded, and where it stops being acceptable is named rather than
glossed.** What stays in the index is sessions and cache entries, which are disposable by
construction — the accepted cost under
**The cache index** (`skills/myflow-contracts/workspace-isolation.md`) is precisely that nothing
which must survive a restart may be kept there, so leaving them costs a login and never data. The
cost that is *not* free is slot exhaustion: a cache offers sixteen indices, one of which is the
empty-id default, so a probe that reads a non-empty index as taken finds fewer free slots as
finished changes accumulate, and a workspace that can claim none falls back to sharing — the failure
this contract exists to remove. The remedy is the operator flushing the cache, and it is safe for
exactly the reason the leftovers are: nothing durable is in there. A project may ship its own
command to list or flush its stale indices; that is the project's tooling, and this row does not
claim it — the `Removed by` cell stays `nothing in this pipeline` either way.

**The workspace row belongs only to a project that declares isolation, and for every other project
it is a row about nothing — which is why it names no database, no bucket and no service.** A project
declares the commands that create these resources, that remove them, and that report which of them
survived; the section holding those declarations is the one
**Project configuration** (`skills/myflow-contracts/project-configuration.md`) is canonical for.

**Which rows run 2 verifies is read off this table, not listed again.** Every row whose lifetime
ends at run 2 is checked back by `<agents repo>/scripts/check-cleanup-complete.sh`, whose header explains which
rows that leaves it reading and why; step 6 of
**Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`) is where its
verdict is acted on.

**That derivation is declared, not left implicit.** The guard carries one marker line per row of
this table saying whether it checks that row or deliberately does not, with the reason; its harness
reads this table and those markers and fails when the two disagree in either direction. So a row
added here goes nowhere until someone records a decision about it — which is what stops a future
artifact from being confirmed clean by a guard that never looked for it.

The terminal
state file keeps `artifactUrl` indefinitely, so deleting the only source that could republish that
URL would leave it advertised and unrepublishable.

## State file

## Project configuration

## Jira integration

## Model policy

The field shape, and the rule that an absent key reads as *not recorded*, belong to
**State file** (`skills/myflow-contracts/state-file.md`) and are not restated here.

**This section is canonical for the model roles, their defaults and how an override applies.** One
location, named here rather than left to be worked out: every `/myflow-*` command is required to
load this file before acting, and none of them loads `<agents repo>/openspec/specs/`, so the file runtime actually
reads is the file the rule has to live in. **State file** (`skills/myflow-contracts/state-file.md`)
cites this section for the `models` field rather than defining the roles a second time, and
`<project>/CLAUDE.md` and `<project>/AGENTS.md` name this section for the same reason.

**Which file to change first.** The normative requirements behind this section belong to the
OpenSpec capability `myflow-model-policy`, whose **Requirement: Implementer subagents run on the strongest available model** (`<agents repo>/openspec/specs/myflow-model-policy/spec.md`) anchors the defaults. That capability is the requirement; this section is the **operational form the commands read**, and the two-layer split is the same one **Planning effort** (`skills/myflow-contracts/state-file.md`) already uses. Change the capability first and bring this section with it: a section that contradicts the requirement is this file's defect, not the spec's. A live spec is also behind by construction while a change is open — its delta lands in `<agents repo>/openspec/specs/` only at finish run 2 — which is the second reason runtime reads this section rather than that file.

That citation is a **checked** one, not a courtesy: the guard associates a bold token with the path
beside it and matches it against the target's headings, and an OpenSpec `### Requirement: …` heading
is a heading like any other, so naming the requirement in full is what makes
`<agents repo>/scripts/check-references.sh` fire when it moves. A bare backticked path with no bold token beside
it is **not** checked and rots silently — which is what this bullet's predecessor did.

The two rules point in opposite directions on purpose. A reviewer's job is to be many independent
readings of a finished diff, so its cost must not scale with the operator's session model. An
implementer's job is to get the diff right the first time, where capability compounds.

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

**The panel-fix default is the strongest available model, and deliberately not Sonnet.** The role
names the agent that *applies* a fix, which is an implementer — so the implementer rule above
already governs it. Fix rounds escalate the panel's breadth rather than its model precisely because
implementers sit at the ceiling from round 1, and a fix-wave default of Sonnet would contradict both
of those rules at once.

**Every subagent dispatch records the model it used** in the SDD ledger, alongside the task it ran.
A model policy that nothing records is a policy nothing can verify — and the absence of that record
is precisely how this rule came to be missing in the first place.

**This record outlives the change.** The ledger is authored under `<abs-worktree>/.superpowers/`, which is
gitignored, in a worktree `/myflow-finish` run 2 removes — but run 1 preserves it into the
repository first, under `<project>/docs/superpowers/ledgers/`, so it serves the operator and the panel
*during* the change and stays answerable afterwards. An after-the-fact audit of which model
implemented which task therefore reads the preserved ledger rather than a transcript nobody kept.
The preservation duty itself is stated once, under
**Run 1 — the branch is not merged** (`skills/myflow-contracts/finish-contract.md`); this section
depends on it rather than restating it.

Durability is a **stronger** reason to leave an unobserved entry unobserved, not a weaker one. A
persisting record makes an invented model slug permanent, so `unknown (agent-defined)` stays exactly
as written and no step fills it in on the way into the repository.

## Change name resolution (all `/myflow-*` commands)

This is why the second source matters: `openspec list --json` only
sees change directories present in the *current* git checkout, so a change staged in a worktree —
`<project>/openspec/changes/<name>/` created there but never committed to the main checkout — is invisible to
it alone, even while it sits at a human gate with a fully staged diff.
