# Implement (SDD + TDD)

Loaded by `skills/flow/SKILL.md` once planning artifacts exist — entered from
`skills/flow/resume.md` on a resumed `STARTED` run whose
plan is already ready, or on a fix run at `IN_PROGRESS`.

| Step | Skill | When |
|------|-------|------|
| **2** | resume the workspace created at kickoff, stated in **2. Isolate the workspace** below | Before the first code change, every run |
| **3** | **superpowers:writing-plans** | Validate the plan; repair `tasks.md` if it is not apply-ready |
| **4** | **superpowers:subagent-driven-development** | Execute the remaining tasks |
| **5** | **superpowers:test-driven-development** | Every implementer dispatch, every task |
| **6** | the review panel | The final whole-branch panel |
| **7** | **superpowers:systematic-debugging** | An unexpected test failure during implementation, or a review-panel finding confirmed as a real defect |
| **8** | **superpowers:verification-before-completion** | Evidence before claiming done |

**Never** invoke `finishing-a-development-branch`. Integration is `skills/flow/integrate.md`'s job.

## The parent orchestrates directly

Sections **1**, **2** and **4** below, `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` are **the parent's own work** — no resumed subagent runs
`flow.implement` through `flow.verify` on its behalf; the running session does it
itself. Every "you" in those files addresses the parent. On a fix run,
the parent resolves the worktree from the state file's `worktrees` map and continues at
`skills/flow/fix.md` (section **3** below) on the same session.

**The parent's shell discipline:** the parent's own Bash work above runs in the session's
interactive shell, and that shell is zsh — never assign `path`, the `PATH`-tied special
variable, and never lean on unquoted variable word-splitting: loop and list values go through
arrays or explicit lists. One git command per `git -C` call, and never
a compound command led by `cd` — the harness hook denies those by design — a compound chain
staying legitimate only when every member is `git -C`-led and pasted whole, as **The guarded commit chain** (`skills/flow-contracts/git-boundaries-commit-chain.md`) is. A background
call inherits no cwd, so every path it names is absolute or `-C`-given — the standing rule the
FOREGROUND BUILDS paragraph of section 4 already states for long-running commands
(`skills/flow/implement.md`).

Read the decision JSON
(`<abs-worktree>/.superpowers/sdd/decision.json`) for the decided `panel` — per **Plan check gate** (`skills/flow/brainstorm-planner.md`) — and for the recorded
`groups` field, which section **4** below
dispatches by.

**Never report a stage boundary or ask the operator with a child in flight** — wait for every
implementer, reviewer, slot or fix subagent launched first. Ending the turn while a child launched
in the background still runs is that wait (**Turn discipline**, below).

### Dispatch sites — the parent's closed list

These ten rows are **every** Agent-tool dispatch the parent may make, across sections **1**,
**2** and **4** below, `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md` and
`skills/flow-contracts/pipeline.md`'s **Pipeline defects found mid-run**:

| Site | Role | Role line | Key shape | Owning section |
|---|---|---|---|---|
| implementer, one per group; a fixer per fix | `implementer`; a fixer the row role its Role line records under (**Every fix is a fresh one-shot**) | `implementer`; a fixer the role whose pair **Every fix is a fresh one-shot** names | `task-<ids>-implementer`, ids `+`-joined in plan order; a fixer's `task-<ids>-implementer-fix-<k>`, `task-<n>-implementer-recommit-<k>`, `task-<ids>-implementer-pick-<k>` or `full-suite-fix-<n>` | section **4** below, **Every fix is a fresh one-shot** |
| gated reviewer bundle, one per batch of returned tasks of weight 5 | `reviewer` | `gated-reviewer`; a fix re-review `gated-re-reviewer` | `task-<n+n+n>-reviewer` | section **4** below, **The gated per-task reviewer** |
| panel bundle, one per decided `panel` bundle | `reviewer` | `panel-reviewer` on pass 1; `panel-rerun` for a re-run | `panel-<round>-<slot\|slot+slot>` | `skills/flow/review-panel.md`, **Bundled dispatch** |
| findings merger, one per round once its dispatches have ended | `merger` | `panel-merger` | `panel-<round>-merger` | `skills/flow/review-panel.md`, **The findings merger** |
| panel-fix, one per chunk of at most 10 findings | `panel-fix` | `panel-fix` | `panel-fix-<round>[-<chunk>]` | `skills/flow/review-panel.md`, the fix step |
| minor fix, one per round that raised only Minors | `panel-fix` | `minor-fix` | `panel-<round>-minor-fix` | `skills/flow/review-panel.md`, **Panel re-runs** |
| fix implementer, one per fix | `implementer` | `implementer` | `fix-<k>-implementer` | `skills/flow/fix.md`, step 2 |
| fix reviewer, one per fix | `reviewer` | `panel-reviewer` | `fix-<k>-review` | `skills/flow/fix.md`, step 4 |
| verifier, one per worktree, or one per shard of a sharded pass | `verifier` | `visual-verifier` | `visual-verify` (`-2`, `-retry`, `-rest`, `-rest-<n>`, `-fix-<k>`, `-s<i>of<n>`) | `skills/flow/visual-verify.md`, **The verifier dispatch** |
| in-run pipeline fix planner, fixer and review, one each per loop step | `planner` for the plan, `implementer` for every fixer, `reviewer` for the review | none — its `Key:` line names its role | `pipeline-fix-<k>-plan`, `pipeline-fix-<k>-fix` (`-fix-g<g>`, `-fix-<r>`, `-fix-v<v>`), `pipeline-fix-<k>-review-<r>` | `skills/flow-contracts/pipeline.md`, **Pipeline defects found mid-run** |

**Every dispatch's prompt carries its Role line** — `Role: <role id>`, a line of its own, the id
the **Role line** column names — and runs on a pair that role's rows of **Allowed pairs**
(`skills/flow-contracts/model-policy.md`) admit. A
no-progress escalation carries `Role: no-progress-fixer` (a pipeline-fix loop's, its loop's `Key:` line too).

The in-run pipeline fix row's dispatches are recorded by `hooks/enforce-loop-dispatch.py` under
the loop's own change, off each prompt's `Key:` line: the parent never begins or ends one under its
own change.

**Everything else in those sections is the parent's own Bash and Read work, never
delegated** — every `flow-guard check-*`, `flow-guard run-reproducer`,
`flow-guard prepare-workspace`, `## lint` and `## test`, every `flow record` and `flow stage` call,
worktree add and remove, every report read and every diff walk. Not to a "verify" reader, a
"re-verify" or "mutation re-verify" agent, a helper, a background task, or a subagent under any
other name. The findings merger's read of a round's reports (**The findings merger**,
`skills/flow/review-panel.md`) is the one exception.

**The self-check.** Before any Agent-tool call, the parent names which row above the call is. A
call that names no row is not made.

**No row is ever a fork.** No `/flow*` phase, and no loop a phase or the self-review runs,
dispatches `subagent_type: "fork"`: every flow dispatch is a fresh agent whose prompt carries what
it needs — paths to the run's artifacts, reports and records — never the parent's context, and
`hooks/enforce-agent-baseline.py` denies a fork in a flow session.

**Every dispatch's Agent-tool `description` names what it carries, in the operator's status-line
shape.** A dispatch carrying plan tasks — an implementer, a gated reviewer bundle — names them
by plan id over the plan's task count, then a few words in parentheses: `Task 3/22 (port guard)`
for one task, `Tasks 3+4+7/22 (review)` for a group, its ids in the order its key lists them. Every other row is its key and a few
words, with no parentheses: `panel-1-primary correctness review`. The `flow-task-list` mod
(`mods/flow-task-list/`) reads that prefix to number its row. The daemon stamps a launch onto the
row its description's first word names. The `Role:` line is the **Role line** column's id above.

**The hook writes the row's begin.** On a flow dispatch with no open row for its key,
`hooks/enforce-agent-baseline.py` runs `flow record dispatch begin` itself once its other checks
pass — the key from the prompt's `Key:` line, else the description's first word; the role from the
prompt's `Role:` line; the model and effort from the Agent call; the session token and change from
the session's latest stage run that is not a loop's own (`pipeline-fix-…`, `self-review-…`). The parent runs no `begin` of its own for such a row. A row
whose begin needs a flag the Agent call cannot carry — an
operator-asked row's `-request` — is still begun by its own
site, and the hook then finds it open and writes nothing. The end stays the parent's, riding the
Bash call it already makes on each return.

**Every dispatch is one-shot — a finished child is never resumed.**
The parent never sends a `SendMessage` to a child that has written its report. Whatever a
finished child's work still needs — a guard refusal, a pick conflict, a reviewer's `fix`, a
full-suite failure — goes to a fresh fixer (**Every fix is a fresh one-shot**, below); a re-review
is a fresh dispatch.

Every row's own prompt or brief carries the NO DELEGATION paragraph (`skills/flow/implementer-brief.md`,
`skills/flow/gated-reviewer-brief.md`, `skills/flow/review-panel.md`, `skills/flow/visual-verify.md`, `skills/flow-contracts/pipeline.md`) — a leaf never dispatches, so
nothing exists below these rows — save the in-run pipeline fix planner, which carries PLANNING
DELEGATION in its place on a `flow-planner-<effort>` definition and may dispatch fresh Explore and
Plan agents, never one that edits (**Allowed pairs**, `skills/flow-contracts/model-policy.md`).
**The `flow-<effort>` family (`agents/flow-low.md`, `agents/flow-medium.md`, `agents/flow-high.md`
— three definitions, one per effort, each carrying `effort:` and no
`model:`) carries a `tools:` allowlist that omits
`Agent`** — the NO DELEGATION paragraph is backed by a capability the dispatched agent
structurally does not have, not by prompt text alone. This covers the panel bundle,
panel-fix and gated per-task reviewer rows. The verifier
row dispatches on the `visual-verifier` role's pair (`skills/flow/visual-verify.md`),
a `flow-<effort>` definition, so it is structurally fork-free too. Every row takes the effort its role's pair names; the no-progress escalation of **Fewest operator actions** (`skills/flow-contracts/pipeline.md`) is the `no-progress-fixer` role's.

**The return.** Once a dispatch's completion notification arrives and its report file exists, read its verdict, print the change's own
handoff or continue to the next stage, and close the record under whichever key is open:

```bash
flow record dispatch end -change <name> -key <the key currently open> -session-token mf-<literal-token> \
  -outcome completed
```

**A dispatch whose agent dies is closed with `-outcome aborted`, reported, and not retried**: print
`/flow <name>` for the operator — a re-run resumes from whatever was left (checkbox state, the
state file's worktrees, findings in the store) through this file's own re-entry rules.

## 1. Load context and validate the plan

```bash
flow stage begin -command '/flow' -stage flow.implement -harness <harness> -session-token mf-<literal-token> <name>
spectre validate "<name>"
spectre list --json
```

Exit `0` from `validate` is the only exit that proceeds. Exit `1` names findings in this change's
own artifacts — most often a step checkbox left at column 0 — and each is repaired here, before any
code is touched. Exit `2` is a usage or IO error, and `no such change "<name>"` is the one worth
naming: nothing has been proposed under that name, so stop and suggest `/flow <name>`.

**The change root is the change's directory** (**A change's directory**,
`skills/flow-contracts/pipeline.md`). Read:

- `<changeRoot>/proposal.md` — what and why
- `<changeRoot>/tasks.md` — the plan
- `<changeRoot>/design.md` — how, when the change carries one
- `<project>/spectre/specs/<capability>.md` for every capability the proposal names

**Every recorded base is compared with the branch it tracks before task 1 runs.** Run
`flow-guard check-base-behind <worktree>...` over the run's resolved worktree set; its header is canonical
for its lines. Exit 0 → continue. Exit 1 → relay every `BASE-BEHIND:` block, run its two command
lines in order as **Verify** (`skills/flow/verify-and-handoff.md`) runs them, and run the check
again. Exit 2 → name its stderr in the handoff and continue.

**Whether there is anything left to implement is read off the task checkboxes**, from
`spectre list --json`'s `{"changes":[{"id","done","total"}]}` for this change.

- `total == 0` → no plan spectre can read: stop, resume at `skills/flow/resume.md`.
- `total > 0` and `done == total` → every task is already checked: close this stage and continue
  into `skills/flow/review-panel.md`, then `skills/flow/verify-and-handoff.md` — never straight to
  integrate.

Confirm `tasks.md` meets writing-plans quality; if it does not, invoke **superpowers:writing-plans**
to repair it before touching code.

Extract the **Global constraints** verbatim from the capability specs the proposal names and
`design.md` for the reviewers.

## 2. Isolate the workspace

Resume `<project>-worktrees/<name>`, created at kickoff. Never implement on the
default branch without explicit consent.

**Persist each worktree's merge base and absolute path to the state file as soon as the worktree
exists — never defer this to the end of the run.** A run that stops, is interrupted, or is
resumed after a context compaction anywhere between here and `flow.verify`
(`skills/flow/verify-and-handoff.md`) must not leave the state record looking like a creating run
with no worktrees, when real worktrees, branches and commits already exist. Record it with one
call, which keeps every peer's entry and leaves `state` exactly as read (a creating run stays
`STARTED`; `flow.verify` is still the only step entitled to flip it to `IN_PROGRESS`):

```bash
flow state add-worktree <name> <abs-path> <merge-base-sha>
```

It exits 2 on a relative path or a malformed sha, and 1 when no `STARTED` record exists yet,
writing nothing either way. Do this once per worktree, immediately after
`spectre link` succeeds for it (or immediately after resuming it, on a resumed run; a fix run
records only an entry the `worktrees` map lacks), not
batched at the end. The kickoff worktree's own entry never waits for this stage: it is recorded
where the worktree is created — `skills/flow/brainstorm.md` step 2 — so on a first creating run
this stage finds it already present, and the paragraph's rule above keeps governing each
additional worktree.

On a fix run whose state record's `worktrees` map is not empty, that map is this run's resolved
set — no `flow state add-worktree` call repeats an entry it holds, the worktree-resolution
contract below is not loaded, and `## apps` is read only when an appended task's `**Files:**` name
a path outside every recorded worktree's repository. `flow workspace-id` and both stage marks run
as on any run.

**Load `skills/flow-contracts/worktree-resolution.md`** — it derives this run's resolved worktree
set.

**This run's resolved worktree set — the set `skills/flow/verify-and-handoff.md` iterates — is the
worktree resumed above, plus any additional worktree this change affects.** Per
**Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`), non-empty by
construction on every ordinary run.

**Every `## apps` entry is resolved in this stage too, before any implementation runs.** A later
stage that finds it needs a worktree in a repository the set does not hold — a fix round's task,
`flow.verify`'s commit destination, a stack that must serve from a checkout of its own —
runs that entry's recipe of
`skills/flow/cross-repo-worktrees.md` (the `flow-guard kickoff-worktree --from` call, the regression
checkout's toolchain, the merge-order extension), never a bare `git worktree add`. A worktree
that will carry no commit is no exception: the recipe records it in the state record's
`worktrees` map with its HEAD as its merge base, so run 2's per-worktree checks and its removal
reach it like any other. Read
`<project>/.flow/project.md`'s `## apps` table and resolve each entry per the roots-are-main-checkouts rule of **Where the agents repository is**
(`skills/flow-contracts/project-configuration.md`): an entry whose
repository already holds a worktree for this change — the kickoff worktree's repository, a
peer's repository linked above, a second entry naming a repository already resolved by an
earlier entry, or a repository where an earlier run of this change created one — records that
worktree as its root and creates nothing.

**Load `skills/flow/cross-repo-worktrees.md` only when** this change links a peer, an `## apps`
entry names a repository holding no worktree for this change, `## visual verification` names a
`regression checkout`, or this run's resolved worktree set spans more than one repository — it carries each created worktree's `## worktree setup`, the `spectre link`
step, the rest of `## apps` resolution, the regression checkout's toolchain and the merge-order
record.

**Then run `flow workspace-id <name>` for this worktree's workspace id**, once per run, on a fix
run exactly as on the first.

## 3. Fixing

**Load `skills/flow/fix.md` only when** this run is a fix run — state `IN_PROGRESS` with
fix instructions, or a plain message at `IN_PROGRESS` (`skills/flow/SKILL.md`) — or a verify
stage reports a fixable defect; it carries the whole fix procedure, and a first run marks nothing
there.

## 4. Execute (SDD + TDD)

**At most one mutator — any Edit/Write-capable dispatch: an implementer, a panel-fix, a fix-round
agent of any name — may be in flight against a given worktree at any moment** — a reviewer is not
one: it reads an immutable commit range, and any number of them may run beside the one mutator.
Dispatches into different worktrees remain free to run concurrently. A visual verifier is a
mutator here too — it builds the stack from the tree and writes `<changeRoot>/visual-verification/`.
The shards of one sharded `flow.verify` pass (`skills/flow/visual-verify.md`)
are the one exception: up to four verifiers of that pass run at once in one worktree, each writing
only its own slice's files. This explicitly overrides
`superpowers:subagent-driven-development`'s parallel dispatch guidance and
`superpowers:dispatching-parallel-agents` for same-worktree tasks. The invariant is the working
tree, not the build tool, and it holds however file-disjoint two tasks look on paper. **A mutating dispatch's
report is complete only when its build's own success line is quoted and `git -C <worktree>
status --porcelain` is empty or every entry it prints is explained in the report** — "waiting for
the build" is never a finished report, and uncommitted WIP is a finding, never a state the next
dispatch inherits; a fix to something visual names the fresh capture taken after that build. A mutator that
finds unrelated uncommitted changes mid-task says so and stops rather than fixing around them.

**The hook begins each implementer's row as it goes out** (**The hook writes the row's begin**,
above), and the parent ends it as soon as that dispatch reports back, before the next one goes out:

```bash
flow record dispatch end -change <name> -key task-<ids>-implementer \
  -session-token mf-<literal-token> -commit <sha> -outcome completed
```

`-commit` takes the task commit's full sha; an abbreviated one is refused.

Both calls are required. `begin` is recorded immediately before the dispatch — when several
launches go out together in one message, in the one Bash call immediately before that message, one
call for all (step 2 of **A group boundary is one wake, one Bash call and one launch message**, below) — carrying no
`-agent-id`: the daemon captures the agent's identifier, and the harvester fills the row's empty
`agent_id` from the launch result afterwards. `-agent-id` is accepted on both
calls as recorded intent the daemon never overwrites, for a caller that knows the id. `-key` is this dispatch's own literal label,
unique within the run's session token — `task-<ids>-implementer`, reused identically in both calls.
Inside a `for` loop it may expand the loop variable — `$k`, `${k}`, or a `#`/`%` strip of it — the
only shapes the harvester reads; any other expansion leaves the row without its agent.
`-role` is one of `implementer`, `reviewer`, `merger`, `panel-fix` or `verifier` (**Verify**,
`skills/flow/verify-and-handoff.md`); `-task` is the task's
flat integer id, omitted for a dispatch against no single task. `-session-token` takes a literal,
never a shell substitution. The start and end instants are the daemon's own — never
caller inputs.
**A dispatch recorded only after it finished is a back-fill, and is one call** — `flow record dispatch backfill`, carrying `begin`'s flags and `end`'s together with the dispatch's real `-started-at` and `-ended-at`, both required — never a `begin` and `end` pair, whose instants would read the back-fill's own moment.

**`-model` is the model this dispatch was actually given — the `implementer` role's pair**
(**Model and effort**, `skills/flow/brainstorm-planner.md`, or the run's session-instruction
override when one was given for the implementer role), `-effort` its effort, and the dispatch's `subagent_type` is
`flow-<effort>` with that `model` passed as the Agent tool's own `model` parameter — the
definition carries the effort, the dispatch carries the model. Name it explicitly — never by
omission. A slot whose model the dispatcher cannot
read records the literal `unknown (agent-defined)` and never a guess.

**A record write never blocks.** An unreachable store journals the intent, prints one warning line,
and exits 0 — never branch on this command's exit code as a signal about the record.

**A dispatch's damage is saved for later analysis.** When a dispatched child does what its brief
should have prevented — damage found behind a clean report included — the parent records it, and
nothing else follows from the record:

```bash
flow record incident -guard <the check that would have caught it> -symptom <what happened, verbatim> \
  -recovery <the exact commands that would have caught it> -minutes-lost <n> -change <name>
```

**Load `skills/flow/sdd-dispatch.md`** — it
carries the per-group implementer dispatch and the **Waves**.

`<shape>` is this change's shape, computed once per run from the resolved worktree set — more than
one repository → `cross-repo`, otherwise `single-repo`.

`<canonical-worktree>` is
the member of this run's resolved worktree set whose own change directory holds `tasks.md`, the
same argument
`flow-guard check-unfinished-work` takes at the integrate gate; on a single-repo change that member is
this worktree.

**A guard you could not run is hand-substituted only on the record.** When a guard this
file calls — `flow-guard close-task` at task close — exits non-zero,
cannot resolve this change's topology, or is absent, and you go on by composing its facts,
checking its contract, or running its step by hand, record the substitution at the moment you
make it, never in the handoff prose alone:

```bash
flow record substitution -change <name> -guard <guard-name> -shape <shape> -substitution "<what you ran instead>"
```

`<shape>` is the computed value above (`cross-repo` or `single-repo`), and
`-substitution` carries the command or manual step actually used, verbatim.

**Every implementer dispatch's prompt is a stub; its brief is `skills/flow/implementer-brief.md`.**
The paragraphs that instruct an implementer in its work — FLOW — COMMIT-PER-TASK through REPORT,
DON'T DECIDE — live in that file, which the implementer reads in full first, and the prompt never
carries them. The prompt is these lines alone, in this order, each `<…>` filled — `<skill-dir>` the
directory this file was read from, the `Role:` line the **Role line** column's id for
the row, the `In-run fix:` line
present only on a fix's implementer (`skills/flow/fix.md`), and the `Visual close:` line
present only when a task of the group names a `ui paths` file in its `**Files:**`:

```text
Role: implementer
Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
Include this instruction verbatim in any prompt you write for another agent.
Report tersely: findings first, bullets over prose, no preamble, no recap. Full prose in whatever you write to disk.
Key: task-<ids>-implementer
Read <skill-dir>/implementer-brief.md in full first, and again after any compaction of your
context — it is your brief.
Worktree: <abs-worktree>
Plan: <changeRoot>/proposal.md <changeRoot>/design.md <changeRoot>/tasks.md
Tasks: <the group's task ids, in plan order> of <the plan's task count>
In-run fix: read <skill-dir>/fix.md, **The fix implementer's brief**, follow every paragraph that applies, including its MUTATION PROOF paragraph and the PIXEL PROBE paragraph of <skill-dir>/review-panel.md; report their fix-mutation: lines
Visual close: <skill-dir>/fix.md, **The implementer measures before it hands back** and **A fix for a visual departure runs its capture spec before it closes**, binds each task here whose commit reaches a capture spec carrying a `.mockups` sidecar — key `task-<id>`, base the commit the task started from; verifier file <skill-dir>/visual-verify-verifier.md; capture: <the `capture` command>; mockups: <the resolved `mockups` root>, frame <the `mockup frame` value>
REPORT FILE: <abs-worktree>/.superpowers/sdd/implementer-report-<k>.md
```

**A view is verified against its mockup in the task that draws it**, so a
departure is fixed by the implementer that drew it, and `flow.verify` confirms the change
instead of finding it.

A fixer's prompt is the same stub under its own `Key:`, its `Tasks:` line the tasks it fixes and a
`FIX INPUT:` line before its REPORT FILE line (**Every fix is a fresh one-shot**, below).

**The parent runs the test list — never an implementer.** Once the plan-last group — the group
holding the plan's last task — has passed the guard, by
direct commit or by its wave's final pick, **the parent itself** runs the resolved `## test` list
once per worktree in this run's resolved set, one worktree after another, in the foreground, as
`flow-guard run-verify-list <worktree> test` — skipping a worktree the skip rule of **Inline verify**
(`skills/flow/verify-and-handoff.md`) prints nothing for
— by hand, in the order `## test` lists it, only where that exits 3 — its output
through `tail`; the green sha each run records in its own worktree is what lets **Inline verify**
(`skills/flow/verify-and-handoff.md`) skip what no later commit touched. Every failure, wherever it lies, is fixed at its root
cause and the list re-run until it passes. A fourth `full-suite-fix-<n>` after three failed ones
is the fresh dispatch of **Three failed rounds** (`skills/flow-contracts/pipeline.md`). The fix is a fresh one-shot, the
`full-suite-fix-<n>` row of **Every fix is a fresh one-shot** below, its `FIX INPUT:` the failing
block, and a failing visual `verify` row in it is re-run narrowed as **Narrowed visual re-runs** states (`skills/flow/fix.md`).
The commit carries `-m "Flow-Key: full-suite-fix-<n>"` (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`).
A fix for a failure that does not recur on every run is a determinism fix, held to the PROOF RUNS paragraph (`skills/flow/implementer-brief.md`): the failure is shown returning with the fix reverted before the fix is committed; one that never returns is handled as `skills/flow/fix.md` states for an observed failure that does not reproduce.
`flow-guard run-verify-list`'s `test` list ends with that section's `verify` run on the specs its `specs`
row maps for the diff — the smoke set alone when it maps nothing or no `specs` row is declared, the
list failing when it exits non-zero, and the whole suite, `<specs>` empty, only when the operator asks for it
— when its own worktree's diff from its base branch touches a `ui paths` glob and
`decision.json` does not record `visual.verify` `skipped`, its header canonical for the rest, so a
capture a shared-component change moved fails here rather than at `flow.verify`. When only the
`--repo` run step 2 of **Visual verification** (`skills/flow/verify-and-handoff.md`) adds exits 0,
for another worktree of the resolved set, its merge base from this run's working notes, the parent
runs that `verify` itself after the list, against the canonical worktree's stack, under the stack
rules of **One stack for the verification tasks** below.
The section's `setup` row runs before the list when the toolchain is missing. A failure a product defect caused
is fixed in this step's fix round; a capture the change moved on purpose — one whose difference
the diff makes as design.md or the task's `**Files:**` intend — is re-recorded in it: its
committed PNGs are deleted and `capture` runs with `<spec>` substituted for its spec, writing them
as PNGs that do not yet exist, and they are committed by the regression checkout's rules, step 12
of `skills/flow/visual-verify.md`.

**The pass-1 panel reads while the suite runs.** On a run whose `skills/flow/review-panel.md`
opens a pass 1, the test-list run above waits for that pass: the parent closes `flow.implement`,
enters the panel, and runs the list there, between the pass-1 launch and the round's
wait, as that file's **Bundled dispatch** states. The reading slots are read-only, so their verdict and the
suite's arrive together rather than one after the other.
The mutation slot is not read-only: it runs the tests once per mutant on the same CPUs, so a `mutation` bundle is begun and launched only once the suite has exited. A full-suite fix is applied only after the round's wait ends, never
while the slots are reading the tree.

**A group boundary is one wake, one Bash call and one launch message.** The unit is the group — the decision's `groups`
entry. At each boundary, in this order:

1. **The returned group's implementer commits**, writes its report and ends; its completion notification
   wakes the parent with its summary head (`skills/flow/implementer-brief.md`).
2. **One Bash call closes the returned group and prints the ready set**, every verdict echoed:
   `flow-guard close-task`, then — only on its exit 0 — the rest.

   ```bash
   flow-guard close-task <canonical-worktree> <name> <merge-base> <task-id>[:<task-sha>] …
   rc=$?; echo "close-task: $rc"
   if [ "$rc" -eq 0 ]; then
     flow-guard ready-groups <changeRoot>/tasks.md <canonical-worktree> <name>; echo "ready-groups: $?"
   fi
   ```

   `flow-guard close-task` checks that each task's commit exists; only when it does,
   `flow tasks tick` per **The guard's pass ticks the task.** (below), and
   `git -C <worktree> push origin spectre/<name>` per **Branch backup**
   (`skills/flow-contracts/git-boundaries.md`): the parent pushes, never the implementer.

   `<canonical-worktree>` is the worktree created or resumed in **2. Isolate the workspace**
   above, `<name>` this run's resolved name. **On a change whose resolved worktree set spans more
   than one repository, a task's sha is the per-repository commit map instead** —
   `<task-id>:<worktree>=<sha>[,<worktree>=<sha>…]`, one pair per repository carrying any of the
   task's commits. A task that landed in one repository passes that repository's
   pair alone, which is the single-sha form; skipping the guard on a cross-repo change
   is the substitution shape **A guard you could not run is
   hand-substituted only on the record** exists to prevent, not a licence — the guard runs.
   A bare `<task-id>` resolves to the newest commit since `<merge-base>` in the canonical worktree whose `Task-Id:` trailer names it.
   The guard's header (`stats/internal/guard/closetask.go`) is
   canonical for its arguments and its order.

   The guard reads git objects and `tasks.md` only, so it is safe while the tree changes, and
   never stashes, reverts or resets. Every verdict is printed and read before anything
   launches: **exit 1** is a re-commit, with nothing ticked or pushed — the re-commit row of **Every fix is a
   fresh one-shot** (below) — and `flow-guard close-task` re-runs before anything below; **exit 2 — a guard's not-a-verdict close (an unreadable plan, a task it
   cannot resolve, a commit range git cannot resolve, a usage error), or a failed tick or push —
   stops the run**: re-committing cannot repair an inability, so it is never folded into
   the re-commit loop.

   **A guard call that times out is inspected before it is retried.** Run
   `git status --porcelain=v2 --branch` and `git stash list` in that worktree first. A
   reverting, rebasing or merging state on the `# branch` lines, a change the run did not
   make, or a stash entry the parent did not push means the tree is not the one the run
   left — end the turn with `## Question` carrying both outputs verbatim; never re-run the
   guard on top of it.
3. **A second Bash call readies each group that call printed** — a
   `throwaway` group's copy (`skills/flow/sdd-dispatch.md`) and, when a batch is due, its reviewer bundle's diffs (**The
   gated per-task reviewer**, below) — and **the next message launches every printed group's
   implementer and the due reviewer bundle**, each with `run_in_background: true`, and the turn ends — the wait
   of **Turn discipline** (below). A boundary whose ready set is empty launches the due reviewer
   bundle alone.

**The guard's pass ticks the task.** Mark a **task's** checkbox `[x]` (`flow tasks tick`)
once `flow-guard close-task`'s checks pass on its commit. A step's checkbox tracks
the step and gates nothing.

**A pivot reconciles the three artifacts together.** When implementation pivots — a reality
discovered mid-run (a route already taken on the base, a capability spec another change already
moved, a premise a measurement disproved) has remaining tasks redesigned rather than implemented
as written — the pivot is resolved **before the colliding code is written**: the run stops at the
discovery, and the redesign is a prompt resolved per **Auto-resolution**
(`skills/flow-contracts/operator-prompts.md`). Only then does the
parent edit `proposal.md`, `design.md` and `tasks.md` in the same pass, never
`tasks.md` alone: `proposal.md`'s `## What changes` is brought to the pivoted scope, the
capability spec the plan edits (`<project>/spectre/specs/<capability>.md`) is rewritten to the pivoted
scope in that same pass, and a
decision the pivot displaces is superseded by ID per **Decisions**
(`skills/flow/brainstorm-planner.md`) — the old entry's `**Status:**` set to `superseded by
<new-id>`, its reasoning retained and its `**Superseded because:**` trigger line appended, a
fresh entry appended, nothing deleted or rewritten — so the record carries the
superseded decision beside its replacement. The amended plan is
re-validated — `spectre validate` — before the next task dispatches.

**The gated per-task reviewer.** One combined review per task — spec compliance and
code quality together — but **one dispatch per batch of returned tasks, never one per task**,
the discipline **Bundled dispatch** (`skills/flow/review-panel.md`) applies to panel rounds.
**The bundle is a batch of weight 5**: returned tasks not yet reviewed accumulate, in plan order, and
at the boundary where their weight reaches 5 — a `**Size:** big` task weighing 3, every other task 1 —
or where no implementer of the repository is left in flight or queued, every one of them, of one repository, goes out in one reviewer Agent call beside the implementers the boundary launches, on the
`gated-reviewer` role's pair (**Model and effort**, `skills/flow/brainstorm-planner.md`), its prompt opening with the
line `Role: gated-reviewer`. A fix re-review bundle runs on its own row instead, its prompt opening
with `Role: gated-re-reviewer` (`skills/flow/gated-review-fix.md`). Each task inside the bundle keeps its own pass: its own
commit-range diff `git diff <task-sha>^..<task-sha>` — a real commit diff, never a snapshot of
the working tree, which the next implementer is editing — its own verdict and its own report
file. Record the bundle as one `dispatches` row (`-role reviewer`, `-key task-<n+n+n>-reviewer`
with the task ids `+`-joined in plan order, the same convention as `panel-<round>-<slot|slot+slot+slot>`;
`-task <n>` only on a one-task bundle, omitted otherwise) and close it with **`-outcome clean`
when every pass is clean, `-outcome fix` when any pass is `fix`**; each pass's own verdict is
its report file's `## Verdict`, so per-task review yield stays measurable.
The reviewer writes every finding its passes raise to the bundle's findings file,
`<abs-worktree>/.superpowers/sdd/findings-gated-<n+n+n>.jsonl`, one line per finding and no ref
(`skills/flow/gated-reviewer-brief.md`), and the parent records them all in one call,
`flow record finding -change <name> -slot gated-reviewer -round 0 -dispatch-seq <the bundle's seq>
-status open -from
<that file>`, which gives each line the change's next free ref, so the run's Findings count holds
what the gated reviewer caught; the panel's rendered record skips `gated-reviewer` rows.
**A returning bundle is one Bash call**, its verdicts read from its final message — one line per
pass — never its report bodies: the record's end, then the findings call above — riding the
boundary call above when the next group's implementer returns in the same wake.
**A pass whose findings are all Minor is `clean`** — no re-review; its Minors stay open findings for the stage-end fixer
(`skills/flow/gated-review-fix.md`). A Minor no change to the tree can resolve is named,
one clause, in the parent's output, and recorded `withdrawn <that clause>`. A `fix` pass
sends its Minors to the same fixer. **A mixed-verdict bundle is handled per task**: every `fix` task waits for the stage-end fix path
on its own sha, independently of its bundle-mates.

**Load `skills/flow/gated-review-fix.md` only when**, once the last group's guard has passed and
every gated bundle has returned, a gated pass closed `fix` or left a Minor open — it carries the stage-end fix path.

**Each pass is a primary pass, scoped to its task.** The bundle's stub names
`skills/flow/primary-reviewer-prompt.md` once, every placeholder but the two per-task ones
filled per the placeholder table of `skills/flow/review-panel.md`.
Each **PASS task-`<n>`** section gives that pass's values for the two, which are its Scope
fields:

- `[DIFF_PATH]` — `<abs-worktree>/.superpowers/sdd/task-<n>.diff`, the parent's
  `git diff <task-sha>^..<task-sha>` written there before the Agent call; on a fix re-review,
  `task-<n>-fix-<k>.diff`, of the range `skills/flow/gated-review-fix.md` names;
- `[ARTIFACT_PATHS]` — the absolute paths of the change's `proposal.md`, `design.md` and
  `tasks.md`, as the table states them.

**Every gated reviewer bundle's prompt is a stub too; its brief is
`skills/flow/gated-reviewer-brief.md`**, read in full first. The prompt is these lines, in this order,
then one `PASS task-<n>` block per task of the group in plan order, each carrying its three Scope
values and its own REPORT FILE line:

```text
Role: gated-reviewer
Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
Include this instruction verbatim in any prompt you write for another agent.
Report tersely: findings first, bullets over prose, no preamble, no recap. Full prose in whatever you write to disk.
Key: task-<n+n+n>-reviewer
Read <skill-dir>/gated-reviewer-brief.md in full first — it is your brief. Each pass below reviews per <skill-dir>/primary-reviewer-prompt.md, [CALIBRATION_PATH] being <skill-dir>/reviewer-calibration.md and its two Scope placeholders the values its PASS block names.
FINDINGS FILE: <abs-worktree>/.superpowers/sdd/findings-gated-<n+n+n>.jsonl
PASS task-<n>:
[DIFF_PATH] <abs-worktree>/.superpowers/sdd/task-<n>.diff
[ARTIFACT_PATHS] <changeRoot>/proposal.md <changeRoot>/design.md <changeRoot>/tasks.md
REPORT FILE: <abs-worktree>/.superpowers/sdd/reviewer-report-task-<n>.md
```

A fix re-review's `Role:` line is `gated-re-reviewer`, its `Key:` `task-<n+n+n>-reviewer-fix-<k>` and
its PASS values those `skills/flow/gated-review-fix.md` names.

**The last group's guard pass is the stage's last boundary.** `final-review.diff` is written and
the slots dispatched once it has passed, every gated reviewer bundle has returned, the stage-end
gated fix and its re-review have closed clean (`skills/flow/gated-review-fix.md`), and the parent's
test-list run has passed, after any fix — save a pass 1,
whose slots are dispatched before the suite runs (**The pass-1 panel reads while the suite runs**,
above); the review panel's pre-work may ride the last boundary's Bash call. No
panel verdict is taken, and no round after pass 1 is dispatched, on a red branch.

**A finished phase is never a stop.** Once a task, a fix round or a re-capture closes clean,
the same turn loads the next phase's file and dispatches its first step — the review panel after
the last tick, verify after the panel. A phase file still to be read, the size of the next phase,
or the context it will cost are never reasons to end the turn; the run ends only where
`skills/flow-contracts/pipeline.md` names a stop — blocked, a human gate, or a question only the
operator can settle. A turn that ends with a background child out is that child's wait
(**Turn discipline**, below), never a stop.

**Turn discipline.** A turn is spent only where an output must be read before the next action is
chosen. Calls that do not depend on one another share one Bash call — every verdict printed, each
read afterwards — and launches that do not depend on one another share one message. A stage's
`flow stage begin` rides its first command and its `flow stage end` its last, in the same Bash
call. **A wait on a child is the end of the turn, never a call:** every child goes out with
`run_in_background: true`, the parent ends its turn once the launch message is sent, and the
harness's completion notification wakes it, carrying the child's final message. While any child
is out and no shorter timer is armed — a slot's ceiling (**No forking, and a wall-clock ceiling on
every slot**, `skills/flow/review-panel.md`) is one — one background `sleep 3000`, launched with
`run_in_background: true` beside the launches, wakes the parent before its one-hour prompt cache
lapses; a wake that finds a child still out re-arms it and ends the turn again. Once the last child out
has returned, the parent stops that timer (`TaskStop` on its background task) before any gate or
turn end, so no stale wake reaches a later turn. `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` state their own batches under this paragraph and restate
none of it.

**A wait never outlives its child.** A child can end without writing its file — its turn closes on
a final message, or it dies — and its completion notification arrives either way, so no wait is
ever a loop on the file: never an `until test -s …` loop, foreground or backgrounded, which waits
forever on a file that will never come. On the notification, one `test
-s` settles it, and an absent file takes that dispatch's no-report path at once — a panel slot's
file written by the parent carrying `no verbatim report captured — <reason>` (**Recording
findings, and the record's format**, `skills/flow/review-panel.md`), its findings read from the child's final message; any
other child closed `-outcome aborted`, as **The return** above closes a dispatch whose agent dies.

**A hand-back is the end of a child; one that keeps running is stopped.** A child whose final
message has arrived (a `SubagentHandback` message, or a completion notification carrying the
report) and whose notification says it is `waiting on its own background work` — or that `ListAgents`
still lists `running` after its report file exists — is a zombie: it holds a live process and, on a
writer, the worktree. The parent `TaskStop`s it at once and takes the commit its report names. The
parent ends each dispatch row at the hand-back (`-outcome completed`, the report being the
record), never at the stage end, so the stage-end check of open rows finds none. A `flow-guard` run
after the report is written is a long-running command under FOREGROUND BUILDS: the child ends its
turn only once it has exited.

**Read discipline.** These five rules are the run's own size
control,
stated once here and cited — never restated — from `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` wherever they read a report or a diff:

- **Never open a report.** The parent opens no report or evidence file a dispatched agent wrote. A
  dispatch's verdict is its final message — an implementer's, fixer's or gated bundle's summary
  head (`skills/flow/implementer-brief.md`, `skills/flow/gated-reviewer-brief.md`), the verifier's
  `## Verdict`, the panel's merger summary — and a judgement that needs a report body goes to the
  agent passed that path: the fixer through `FIX INPUT:`, the merger, the planner. A file a guard or
  `flow` parses, and output the operator must see, are not the parent's reads.
- **Never read `final-review.diff` or a panel-fix diff whole.** A slot
  reads the diff it was dispatched against; the parent walks a fix's hunks through `git diff
  --stat` and the specific hunks a finding names, never the whole diff.
- **Test/lint output through `tail`.** A targeted test or lint run's output is piped through
  `tail` (`| tail -20`, the failing block reproduced from the log file on a failure) — already the
  rule for implementers (TARGETED TESTS); it binds the parent's own `## lint`/`## test` runs in
  `flow.verify` and the test-list run after the last group the same way.
- **Phase files read once per run.** `implement.md`, `review-panel.md`, `verify-and-handoff.md`
  — and `fix.md`, `cross-repo-worktrees.md`, `sdd-dispatch.md`, `gated-review-fix.md`,
  `implementer-brief.md`, `review-panel-fix-round.md`, `review-panel-optional-slots.md`,
  `visual-verify.md` when their load directive or stage loads them — are each read in full
  once, at the start of the stage that needs them; a later need is served by
  `grep -n` for the heading plus `sed -n` for that section, never a second full read.
- **Change artifacts read once**, `proposal.md`/`design.md`/`tasks.md` at `flow.implement`;
  `tasks.md` re-read only through `spectre list --json` and `flow tasks tick` output afterward.

The parent's own `## lint`/`## test` runs and boundary checks are bound by the same two paragraphs:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

> **SCRUB A REPRODUCED ARTIFACT:** A real artifact reproduced into a committed fixture — a transcript, a log, a database dump — is reduced to the lines the test needs and scrubbed of every identifier, path, prompt text and personal datum before it is staged; the raw copy stays under `<abs-worktree>/.superpowers/sdd/`, never in the tree (kan-1132 self-review, F6).

On BLOCKED: pause and report. Never guess. A question the report carries that takes one
recommended option is not a guess: **Auto-resolution** (`skills/flow-contracts/operator-prompts.md`)
takes it, and only a question with none pauses the run.

```bash
flow stage end -command '/flow' -stage flow.implement -outcome completed <name>
```

Once this stage completes, continue into `skills/flow/review-panel.md`.

### Every fix is a fresh one-shot

**The parent edits no code.** Whatever a task's close, a wave's pick, a gated pass or
the test list still needs changed in the tree goes to a fresh one-shot dispatch — never the
parent, never a resumed child (**Every dispatch is one-shot**, above) — on its row below. Its prompt is the implementer stub (section
**4** above) under the row's own `Key:`, its `Tasks:` line the tasks it fixes and a `FIX INPUT:`
line naming what to fix and, on the gated and full-suite rows, `MUTATION PROOF`; its work is the
FIX DISPATCH paragraph (`skills/flow/implementer-brief.md`).

| Fix | Key | Pair | When |
|---|---|---|---|
| every `fix` verdict and every Minor the gated passes left open | `task-<ids>-implementer-fix-<k>`, the ids every task it fixes, `+`-joined in plan order | the `panel-fix` role's pair | at the stage's end, one per group of at most weight 3, in series within a worktree (`skills/flow/gated-review-fix.md`) |
| a `flow-guard close-task` exit 1 a re-commit repairs | `task-<n>-implementer-recommit-<k>` | the `panel-fix` role's pair | at once, before the next group launches |
| a wave member's pick conflict | `task-<ids>-implementer-pick-<k>`, the member's task ids | the `implementer` role's pair | at once, the pick aborted first (`skills/flow/sdd-dispatch.md`) |
| a full-suite failure | `full-suite-fix-<n>` | the `panel-fix` role's pair | each time the list fails (**The parent runs the test list**, above) |

A full-suite fix takes its name from its key: the run it repairs is the test list, never the
full suite, which is `flow-guard run-verify-list -full` and runs only when the operator asks.
`full-suite-fix-<n>` keys the implement phase's full-suite fixes only; a fix at `flow.verify` is an
`fix-<k>-implementer` fix, its re-runs keyed `-fix-<k>` (`skills/flow/fix.md`).

Every row's prompt carries the `Role:` line of the role whose pair the table names, and its row records under that role — `-role panel-fix` on the `panel-fix` rows, `-role implementer` on the pick row — by `flow record
dispatch begin` and `end` as a group's implementer is — its `-model` and `-effort` the pair the
table names, the end's `-commit` its last commit — and launched with `run_in_background: true` like
every child. `<k>` counts that key's fixes from 1. A fixer is a mutator: it goes out only while no
implementer writes its worktree. 
**The parent keeps every judgement**, and delegates none: whether a finding stands, or is withdrawn with its one clause; a BLOCKED report or a question,
per **Auto-resolution** (`skills/flow-contracts/operator-prompts.md`); a pivot; every `flow record finding`, `status` and `incident` call; and whether a fixer or a
re-review goes out. No child dispatches, waits on or resumes another.

### One stack for the verification tasks

**Before the first task whose `**Files:**` only add end-to-end specs, fidelity captures or their
baselines goes out**, the
parent starts the canonical worktree's stack once, from the project's `## stop` when it declares a
command, then its `## run` — never the flow dev stack (`<project>/CLAUDE.md`); a start that fails or
is refused on a held port ends the turn, naming the command's output — then runs any build the plan's Global Constraints name
for those tasks. Every such task, and every wave copy it runs in, runs against that stack's worktree-resolved
URLs, which its prompt names, and starts, stops and rebuilds nothing. The stack stays up for
`flow.verify`.
Every such task's prompt, and every task whose `**Files:**` add or change a capture spec, carries the capture rules of step 8 of `skills/flow/visual-verify-verifier.md` by absolute path: a capture is taken in the state it names — every interaction producing that state first, each asserted, and the pointer parked clear of every control after the last click.
