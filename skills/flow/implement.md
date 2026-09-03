# Implement (SDD + TDD)

Loaded by `skills/flow/SKILL.md` once planning artifacts exist — continuing straight from
`skills/flow/brainstorm.md` on a creating run, or entered directly on a resumed `STARTED` run whose
plan is already ready, or on a fix run at `IN_PROGRESS`.

| Step | Skill | When |
|------|-------|------|
| **2** | **superpowers:using-git-worktrees** | Before the first code change, on a first run |
| **3** | **superpowers:writing-plans** | Validate the plan; repair `tasks.md` if it is not apply-ready |
| **4** | **superpowers:subagent-driven-development** | Execute the remaining tasks |
| **5** | **superpowers:test-driven-development** | Every implementer dispatch, every task |
| **6** | **superpowers:requesting-code-review** + the review panel | Per-task review, then the final whole-branch panel |
| **7** | **superpowers:systematic-debugging** | An unexpected test failure during implementation, or a review-panel finding confirmed as a real defect |
| **8** | **superpowers:verification-before-completion** | Evidence before claiming done |

**Never** invoke `finishing-a-development-branch`. Integration is `skills/flow/integrate.md`'s job.

## Dispatch the conductor

Sections **1**, **2** and **4** below, `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` are the **conductor's** work, not the parent's — one conductor
subagent runs `flow.load-context` through `flow.write-in-progress`, resumed between its returns so
its context carries from the plan to the handoff. Every "you" in those files addresses it. The
parent's own work on this branch is what this section states, plus — on a fix run — section **3**
below, which runs **before** the dispatch: the parent resolves the worktree from the state file's
`worktrees` map, runs section 3's planner dispatch and Jira sync, and only then dispatches the
conductor. A fix run's stage order is therefore document-fix → load-context → isolate (resume) →
sdd-tdd → …, so the appended plan is validated after the fix's edit.

**Resolve `DEFAULT_MODEL` and `REVIEWERS`** per **Model resolution** (`skills/flow/SKILL.md`),
and run the guard-presence check, before dispatching. Dispatch one subagent with the Agent tool's
`model` parameter set to `DEFAULT_MODEL` — or the run's plain-language session override, recorded
with the dispatch — and `subagent_type: general-purpose`. Its prompt carries, verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

and states: the change name `<name>`; the project root; `<changeRoot>`; this run's literal session
token; the harness; `DEFAULT_MODEL` and the resolved `REVIEWERS` list; the guard-presence result;
the run kind (creating or fix) and, on a fix run, the operator's fix instructions; and the
instruction to read this file's sections **1**, **2** and **4**, `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` and follow them **as the conductor**, running every stage mark
in that range itself with the token it was given.

**The relay contract**, stated in the same prompt. The conductor has no channel to the operator and
no task-list tool. It ends a turn only with one of three blocks, and never with a child subagent
still in flight — it waits for every implementer, reviewer, slot and fix subagent it launched
first:

- `## Question` — the question plus named options; the parent asks it verbatim through
  **AskUserQuestion** and resumes the conductor via **SendMessage** with the answer. Every operator
  prompt inside the covered stages goes this way: the over-cap choice, a second wall-clock breach,
  the non-converging-finding handback, BLOCKED, an empty resolved-worktree set, a plan-quality
  repair that needs the operator.
- `## Stage flow.<key>` plus one line of outcome, at every `flow stage end` it runs; the parent
  updates the harness task list — the stage is the granularity on this branch, per **Progress
  visibility** (`skills/flow-contracts/pipeline.md`) — and resumes it with `continue`.
- `## Handoff` carrying the `IN_PROGRESS` handoff block verbatim; the parent prints it unchanged.

The first line of its first reply is `Model: <the model named in its own system prompt>`.

**Record the dispatch immediately after the launch returns its identifier**, before anything else:

```bash
flow record dispatch begin -change <name> -role conductor -model <DEFAULT_MODEL> \
  -key conductor -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

**The handshake.** Compare the `Model:` line against `DEFAULT_MODEL` (or the override). A match
proceeds. A mismatch records the model that answered and **continues on the running agent — no
re-dispatch**: the planner's opus re-dispatch (**Dispatch the planner**,
`skills/flow/brainstorm.md`) exists because fable may be unavailable, which `DEFAULT_MODEL` does
not share, and a re-dispatch onto a costlier model would raise what this dispatch exists to cut:

```bash
flow record dispatch end -change <name> -key conductor -session-token mf-<literal-token> \
  -outcome fallback -ended-at <ts>
flow record dispatch begin -change <name> -role conductor -model <the model the handshake named> \
  -key conductor-<that model, lowercased> -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

**A mark or a record never blocks** — proceed on the handshake's outcome regardless of whether any
`flow` call reached the store.

**The return.** Once `## Handoff` arrives, print the block unchanged and close the record under
whichever key is open:

```bash
flow record dispatch end -change <name> -key <the key currently open> -session-token mf-<literal-token> \
  -outcome completed -ended-at <ts>
```

The run is at `IN_PROGRESS`; nothing further runs in this invocation.

**A conductor that ends without one of the three blocks, or whose agent dies, is closed with
`-outcome aborted`, reported, and not retried**: print `/flow <name>` for the operator — a re-run
resumes from whatever the conductor left (checkbox state, the state file's worktrees, findings in
the store) through this file's own re-entry rules, and the operator should see the death rather
than have it hidden by a second dispatch.

**At depth, the Agent tool offers no `bugbot` or `security-review` type.** **An unspawnable id is
substituted, not skipped** (`skills/flow/review-panel.md`) applies unchanged; under the conductor
it is the norm, not the exception.

## 1. Load context and validate the plan

```bash
flow stage begin -command '/flow' -stage flow.load-context -harness <harness> -session-token mf-<literal-token> <name>
spectre validate "<name>"
spectre list --json
check-plan-shape.sh "<changeRoot>/tasks.md"
```

Exit `0` from `validate` is the only exit that proceeds. Exit `1` names findings in this change's
own artifacts — most often a step checkbox left at column 0 — and each is repaired here, before any
code is touched. Exit `2` is a usage or IO error, and `no such change "<name>"` is the one worth
naming: nothing has been proposed under that name, so stop and suggest `/flow <name>`.

Exit `0` from `check-plan-shape.sh` proceeds. Exit `1` names a shape defect in this plan's own
`**Files:**`, `**Tests:**` or other declared fields — repaired here, before any task is dispatched
and before any code is touched, exactly as an exit-1 from `spectre validate` already is. Exit `2`
stops the run.

**The change root is `<project>/spectre/changes/<name>/`, by construction.** Read:

- `<changeRoot>/proposal.md` — what and why
- `<changeRoot>/tasks.md` — the plan
- `<changeRoot>/design.md` — how, when the change carries one
- `<project>/spectre/specs/<capability>.md` for every capability the proposal names

**Whether there is anything left to implement is read off the task checkboxes**, from
`spectre list --json`'s `{"changes":[{"id","done","total"}]}` for this change:

- `total == 0` → no plan spectre can read: stop, resume at `skills/flow/brainstorm.md`.
- `total > 0` and `done == total` → every task is already checked: proceed to
  `skills/flow/integrate.md`.

Confirm `tasks.md` meets writing-plans quality; if it does not, invoke **superpowers:writing-plans**
to repair it before touching code.

Extract the **Global constraints** verbatim from the capability specs the proposal names and
`design.md` for the reviewers.

```bash
flow stage end -command '/flow' -stage flow.load-context -outcome completed <name>
```

## 2. Isolate the workspace (first run only)

**Load `skills/flow-contracts/artifacts-registry.md`** — the worktree and branch this step
creates are rows in it.

```bash
flow stage begin -command '/flow' -stage flow.isolate-workspace -harness <harness> -session-token mf-<literal-token> <name>
```

Invoke **superpowers:using-git-worktrees**. Branch `spectre/<name>`. Never implement on the default
branch without explicit consent. Record each worktree's merge base and absolute path in this run's
own working notes as soon as the worktree exists — the state file's `worktrees` map is written only
at the end of `skills/flow/verify-and-handoff.md`.

**First run only:** copy `<project>/spectre/changes/<name>/` from the main checkout into the same
relative path inside the new worktree, then remove the main checkout's own copy of that directory.
From this point on, the worktree's copy is the one that gets edited and eventually committed — the
main checkout's copy is stale and untracked, and nothing later in the pipeline removes it. Leaving
it in place is what made `prepare-archive-branch.sh` refuse on a dirty main checkout during KAN-271's
own run.

On a fix run, resume the existing worktree and make no such copy. **Never create a second one.**

**Load `skills/flow-contracts/worktree-resolution.md`** — it derives this run's resolved worktree
set.

**This run's resolved worktree set — the set `skills/flow/verify-and-handoff.md` iterates — is the
worktree just created or resumed above, plus any additional worktree this change affects.** Per
**Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`), non-empty by
construction on every ordinary run.

**After creating each worktree beyond the canonical one, run `spectre link <canonical-peer>:<name>`
in it**, where `<canonical-peer>` is the canonical repository's own name in that worktree's
`<project>/spectre/peers` file. Record what the command wrote — or that it refused — alongside that
worktree's merge base in this run's working notes. A failure is reported and the run continues: the
link is not a gate, and a change with one worktree runs nothing here.

**Then run `flow workspace-id <name>` for this worktree's workspace id**, once per run, on a fix
run exactly as on the first.

```bash
flow stage end -command '/flow' -stage flow.isolate-workspace -outcome completed <name>
```

## 3. Documenting a fix, before implementing it

**Parent work, run before the conductor is dispatched** — see **Dispatch the conductor** above.
Everything below is the parent's own; the conductor never sees this section.

**Fix runs only** — a first run creates the worktree instead, per **2** above, and marks nothing
here:

```bash
flow stage begin -command '/flow' -stage flow.document-fix -harness <harness> -session-token mf-<literal-token> <name>
flow record dispatch begin -change <name> -role planner -model <PLANNING_MODEL> \
  -key planner-fix-<n> -session-token mf-<literal-token> -started-at <ts>
```

Record what changed **before** writing code, so the proposal never goes stale. `<n>` is this fix
run's own ordinal — one more than the number of fix rounds already recorded in `proposal.md`/
`tasks.md` or as `<name>-fix-N` sub-changes, the same `N` the "where should it go" prompt's
sub-change option below names — so the first fix run's dispatch is `planner-fix-1`, the second
`planner-fix-2`, and so on. Dispatch this fix's planner the same way **Dispatch the planner**
(`skills/flow/brainstorm.md`) dispatches a creating run's — same handshake, same `opus` fallback
**and the same key-suffix rule that section states: the opus re-dispatch records under
`planner-fix-<n>-opus`, and a second mismatch's under `planner-fix-<n>-<model>`, never a repeat of
`planner-fix-<n>`** — same relay contract, same `Model:` first line — with the fix instructions in
place of the design checklist; that section is canonical for the mechanics and is not restated
here.

The planner opens with a `## Question` asking where the fix should go, relayed through the parent
exactly as any other, shape per Operator prompts (`skills/flow-contracts/operator-prompts.md`):

> **This fix has to be recorded before it is written — where should it go?**
> - **Append to `proposal.md` and `tasks.md`** *(default, recommended)* — nothing new is created
> - **Create a linked `<name>-fix-N` sub-change** — its own proposal and plan, for a fix that adds
>   scope the parent change does not describe

The planner writes the append, or the sub-change's own proposal and plan, and returns `## Plan`.
**The Jira description sync stays in the parent** — never the planner's job. **Load
`skills/flow-contracts/jira-integration.md`.** If the fix adds scope the linked Jira issue does not
describe, sync the issue **description** per **Description sync** in Jira integration
(`skills/flow-contracts/jira-integration.md`). Never transition the issue here.

```bash
flow stage end -command '/flow' -stage flow.document-fix -outcome completed <name>
flow record dispatch end -change <name> -key <the key currently open> -session-token mf-<literal-token> \
  -outcome completed -ended-at <ts> -agent-id <id>
```

`<the key currently open>` is `planner-fix-<n>` on a clean handshake, `planner-fix-<n>-opus` after
one mismatch, or `planner-fix-<n>-<model>` after a second — the same rule stated above.

## 4. Execute (SDD + TDD)

```bash
flow stage begin -command '/flow' -stage flow.sdd-tdd -harness <harness> -session-token mf-<literal-token> <name>
```

**Gather the dispatch context bundle before dispatching any implementer.**

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context.md
```

where `<changeRoot>` is `<project>/spectre/changes/<name>/` resolved inside this worktree, and
`<principles-path>` is the **absolute** path of `engineering-principles.md` **beside this file** —
`skills/flow/`, always. A non-zero exit — including the guard being absent — is reported, and
dispatching proceeds with the prompt shape this stage used before this capability existed; the
bundle never gates a run. Confirm the bundle was actually written (`test -f
<worktree>/.superpowers/sdd/dispatch-context.md`) and report plainly if it is not. **Never read the
bundle back into this context** — `test -f` is the whole check; its content is the implementer's
input, not the dispatcher's. Report the script's stderr line for this stage (`bundle unchanged —
reusing …` or `bundle rebuilt — …`) as part of this stage's own reporting.

**At most one implementer subagent may be in flight against a given worktree at any moment** —
a reviewer is not one: it reads an immutable commit range, and any number of them may run beside
the one implementer. Dispatches into different worktrees remain free to run concurrently. This
explicitly overrides `superpowers:subagent-driven-development`'s parallel dispatch guidance and
`superpowers:dispatching-parallel-agents` for same-worktree tasks.

**The parent records each dispatch in two calls — one as it goes out, one as it comes back.**
Immediately before dispatching:

```bash
flow record dispatch begin -change <name> -task <n> -role implementer -model <m> \
  -key task-<n>-implementer -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

and as soon as that dispatch reports back, before the next one goes out:

```bash
flow record dispatch end -change <name> -key task-<n>-implementer \
  -session-token mf-<literal-token> -commit <sha> -outcome completed -ended-at <ts> \
  -agent-id <id>
```

**Both calls are required. Every launch is asynchronous and returns the agent's identifier at
launch, so `begin` carries `-agent-id <id>` and is recorded immediately after the launch returns,
before any other action; `end` may repeat the id.** `-key` is this dispatch's own literal label,
unique within the run's session token — `task-<n>-implementer`, reused identically in both calls.
`-role` is one of `implementer`, `reviewer`, `panel-fix` or `verifier` (**Verify**,
`skills/flow/verify-and-handoff.md`); `-task` is the task's
flat integer id, omitted for a dispatch against no single task; `-started-at`/`-ended-at` are
RFC 3339 — `-started-at` the launch time. `-session-token` takes a literal, never a shell
substitution. Two dispatches starting at one instant are told apart only by id, and a resumed
dispatch shares its id with the original — which is why the id is recorded at launch.

**`-model` is the model this dispatch was actually given — `DEFAULT_MODEL`** (`skills/flow/SKILL.md`'s
**Model resolution**), or the run's session-instruction override when one was given for the
implementer role. Name it explicitly — never by omission. A slot whose model the dispatcher cannot
read records the literal `unknown (agent-defined)` and never a guess.

**A record write never blocks.** An unreachable store journals the intent, prints one warning line,
and exits 0 — never branch on this command's exit code as a signal about the record.

Dispatch one implementer per bundle from:

```bash
plan-dispatch-bundles.sh <changeRoot>/tasks.md
```

Exit 0 proceeds. A non-zero exit is a plan defect: exit 1 names a task missing its `**Files:**`
field, repaired by `superpowers:writing-plans` before any dispatch happens; exit 2 stops the run.
Bundling does not change the commit-per-task model — an implementer handed a bundle still makes one
commit per task, carrying that task's own `Task-Id:` trailer — a red task and its partner make one
commit between them — and a `Build: red` task is bundled with, and commits with, the partner its
`**Squash-with:**` field names. Every implementer dispatch **must** carry:

> **FLOW — COMMIT-PER-TASK:** Do **not** run `git push`, merge, or open a PR. As soon as
> RED-GREEN-REFACTOR completes for this task — before the parent dispatches review for it — commit
> your work with `git commit`, carrying a `Task-Id: <n>` trailer. The trailer identifies the task;
> the subject is this task's declared `**Commit:**` field, reproduced exactly. **Never weaken or
> bypass a project's commit validation to fit** — no `--no-verify`. You **may** `git add`/`git
> commit` your own work, but never `<project>/spectre/changes/` or `<project>/docs/superpowers/`.
> **A capability spec under `<project>/spectre/specs/` is your work, not theirs**: when this task's
> `**Files:**` names one, edit it and commit it here, in this task's own commit.

**A `Build: red` task is dispatched with its `Squash-with:` partner, in one bundle.** The
implementer runs the red task first — writes its tests, runs them, and reports the failing
output — then the green partner, and makes **one** commit for the pair: the partner's declared
`**Commit:**` subject and `Task-Id: <partner>` trailer. The red task never has a commit of its
own; `check-task-commit-fields.sh` resolves the pair from either id against that commit.

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this
> task. Delete any code written before its test.

> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, invoke
> superpowers:systematic-debugging before writing a fix. An expected RED step needs no invocation.

> **REQUIRED READING:** the engineering principles section of the context bundle below — your
> implementation must satisfy these principles; the panel's principles reviewer checks the diff
> against them.

> **CONTEXT BUNDLE:** `<abs-worktree>/.superpowers/sdd/dispatch-context.md` carries this change's
> proposal, design, plan and engineering principles, gathered for you. You **must** still read the
> actual diff and the actual code — the bundle is shared *input*, never a substitute for the source.
> It also carries this project's `## lint`/`## test`/`## run` commands, already resolved — you do
> not need to open `<project>/.flow/project.md` yourself for them.

> **PLAN PROVENANCE:** a fenced block tagged `unverified:` is a hypothesis, not code to transcribe.
> Establish the real API before writing against it, and report what you found. When what you
> measure contradicts the plan, stop and report the measurement: see **When a measurement
> contradicts the plan** (`skills/flow-contracts/plan-provenance.md`).

Every implementer dispatch **must** also carry:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one test you write MUST exercise the real
> thing. A test backed by a fake or a hand-built value passes while the real integration is broken:
> the shape you construct by hand is not the shape the real producer emits. Build the value the way
> production builds it, or assert against the real boundary.

**Review overlaps the next implementer.** The unit is the bundle `plan-dispatch-bundles.sh` emits;
"bundle N's reviewers" is one reviewer per commit in it — a red task and its partner share one. At
each boundary, in this order:

1. **Bundle N+1's implementer commits** and reports its shas.
2. **A pending fix for bundle N folds in first.** If bundle N's review raised a fix, resume that
   bundle's implementer (`SendMessage`; record the resumption as its own pair under
   `task-<n>-implementer-fix-<k>`, `-agent-id` the implementer's own id) with the reviewer's
   report path; it commits `git commit --fixup=<task-sha>` and runs `git rebase --autosquash`. A
   conflict there is between two of the branch's own commits and the implementer resolves it —
   `skills/flow/integrate.md`'s never-auto-resolve rule concerns the operator's base branch.
3. **Guard every commit whose sha is new** — bundle N+1's tasks and every task the rebase rewrote —
   before dispatching review for it, passing the canonical worktree's absolute path (the worktree
   created or resumed in step **2** above) as the guard's fifth argument and this run's resolved
   `<name>` as its sixth:

   ```bash
   check-task-commit-fields.sh <worktree> <task-id> <task-sha> <task-base> <canonical-worktree> <name>
   ```

   The guard reads git objects and `tasks.md` only, so it is safe while the tree changes. A
   nonzero exit is a guard failure, not a review finding — it does **not** consume a fix-round
   slot; send the task back to the **same implementer**, then re-run the guard, before anything
   below.
4. **Dispatch together:** bundle N's re-reviewer if step 2 ran (the task's full range,
   `git diff <task-base>..<new-task-sha>`), bundle N+1's reviewers, and bundle N+2's implementer.

**When the script cannot be located**, apply `flow-task-commit-fields`'s rules by hand: check the
commit's `Files:` against `git diff --name-only <task-base>..<task-sha>`, its `Tests:` against the
commit's diff, and its `Commit:` against the commit's actual subject line.

**Per-task review:** the reviewer gets the commit-range diff `git diff <task-base>..<task-sha>` —
a real commit diff, never a snapshot of the working tree, which the next implementer is editing.
**Record the reviewer's dispatch too** — `-role reviewer`, the same `-task <n>`, `-model
DEFAULT_MODEL`, `-agent-id` on `begin` — and close it with **`-outcome clean` or `-outcome fix`**,
so per-task review yield is measurable. **The per-task review is a single combined reviewer**,
covering spec compliance and code quality together, dispatched on `DEFAULT_MODEL` — `/flow`'s
panel carries no roster, so there is no `full`-preset split into two per-task reviewers. Mark a
**task's** checkbox `[x]` (`flow tasks tick`) only after that task passes spec **and** quality
review — the guard has already passed by then; a step's checkbox tracks the step and gates nothing.
A red task's checkbox is ticked together with its partner's, when their one commit passes review.

**The last bundle's reviewers run alone.** Overlap them only with the review panel's pre-work —
its citation check, bundle rebuild, relocation comparison and diff-size check; `final-review.diff`
is written and the slots dispatched once the last review is clean and any fix folded.

**Never end a turn with a child in flight** — wait for every implementer and reviewer launched
before reporting a stage boundary or asking the operator anything.

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

On BLOCKED: pause and report. Never guess.

```bash
flow stage end -command '/flow' -stage flow.sdd-tdd -outcome completed <name>
```

Once this stage completes, continue into `skills/flow/review-panel.md`.
