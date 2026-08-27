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

## 1. Load context and validate the plan

```bash
flow stage begin -command '/flow' -stage flow.load-context -harness <harness> -session-token mf-<literal-token> <name>
spectre validate "<name>"
spectre list --json
```

Exit `0` from `validate` is the only exit that proceeds. Exit `1` names findings in this change's
own artifacts — most often a step checkbox left at column 0 — and each is repaired here, before any
code is touched. Exit `2` is a usage or IO error, and `no such change "<name>"` is the one worth
naming: nothing has been proposed under that name, so stop and suggest `/flow <name>`.

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

**This run's resolved worktree set — the set `skills/flow/verify-and-handoff.md` iterates — is the
worktree just created or resumed above, plus any additional worktree this change affects.** Per
**Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`), non-empty by
construction on every ordinary run.

**Then compute this worktree's workspace id from the change name.** The derivation is stated once
under **The workspace id** (`skills/flow-contracts/workspace-isolation.md`) — do not re-derive it
by hand. Compute it once per run, on a fix run exactly as on the first.

```bash
flow stage end -command '/flow' -stage flow.isolate-workspace -outcome completed <name>
```

## 3. Documenting a fix, before implementing it

**Fix runs only** — a first run creates the worktree instead, per **2** above, and marks nothing
here:

```bash
flow stage begin -command '/flow' -stage flow.document-fix -harness <harness> -session-token mf-<literal-token> <name>
```

Record what changed **before** writing code, so the proposal never goes stale. Ask which of exactly
two, shape per Operator prompts (`skills/flow-contracts/operator-prompts.md`):

> **This fix has to be recorded before it is written — where should it go?**
> - **Append to `proposal.md` and `tasks.md`** *(default, recommended)* — nothing new is created
> - **Create a linked `<name>-fix-N` sub-change** — its own proposal and plan, for a fix that adds
>   scope the parent change does not describe

If the fix adds scope the linked Jira issue does not describe, sync the issue **description** per
**Description sync** in Jira integration (`skills/flow-contracts/jira-integration.md`). Never
transition the issue here.

```bash
flow stage end -command '/flow' -stage flow.document-fix -outcome completed <name>
```

## 4. Execute (SDD + TDD)

```bash
flow stage begin -command '/flow' -stage flow.sdd-tdd -harness <harness> -session-token mf-<literal-token> <name>
```

**Gather the dispatch context bundle before dispatching any implementer.**

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  > <worktree>/.superpowers/sdd/dispatch-context.md
```

where `<changeRoot>` is `<project>/spectre/changes/<name>/` resolved inside this worktree, and
`<principles-path>` is the **absolute** path of `engineering-principles.md` **beside this file** —
`skills/flow/`, always. A non-zero exit — including the guard being absent — is reported, and
dispatching proceeds with the prompt shape this stage used before this capability existed; the
bundle never gates a run. Confirm the bundle was actually written (`test -f
<worktree>/.superpowers/sdd/dispatch-context.md`) and report plainly if it is not.

**At most one implementer subagent may be in flight against a given worktree at any moment.** The
parent waits for the previous implementer's commit sha for that worktree before dispatching the
next implementer into it; dispatches into different worktrees remain free to run concurrently. This
explicitly overrides `superpowers:subagent-driven-development`'s parallel dispatch guidance and
`superpowers:dispatching-parallel-agents` for same-worktree tasks.

**The parent records each dispatch in two calls — one as it goes out, one as it comes back.**
Immediately before dispatching:

```bash
flow record dispatch begin -change <name> -task <n> -role implementer -model <m> \
  -key task-<n>-implementer -session-token mf-<literal-token> -started-at <ts>
```

and as soon as that dispatch reports back, before the next one goes out:

```bash
flow record dispatch end -change <name> -key task-<n>-implementer \
  -session-token mf-<literal-token> -commit <sha> -outcome completed -ended-at <ts> \
  -agent-id <id>
```

**Both calls are required, and `begin` must go out BEFORE the dispatch does — never delayed to
obtain an identifier.** `-key` is this dispatch's own literal label, unique within the run's
session token — `task-<n>-implementer`, reused identically in both calls. `-role` is one of
`implementer`, `reviewer`, `panel-fix` or `red-partner`; `-task` is the task's flat integer id,
omitted for a dispatch against no single task; `-started-at`/`-ended-at` are RFC 3339.
`-session-token` takes a literal, never a shell substitution. `-agent-id` goes on `end` here — a
serialized implementer dispatch reports its own identifier only once it comes back.

**`-model` is the model this dispatch was actually given — `DEFAULT_MODEL`** (`skills/flow/SKILL.md`'s
**Model resolution**), or the run's session-instruction override when one was given for the
implementer role. Name it explicitly — never by omission. A slot whose model the dispatcher cannot
read records the literal `unknown (agent-defined)` and never a guess.

**A record write never blocks.** An unreachable store journals the intent, prints one warning line,
and exits 0 — never branch on this command's exit code as a signal about the record.

Invoke **superpowers:subagent-driven-development**, dispatching one implementer per bundle from:

```bash
plan-dispatch-bundles.sh <changeRoot>/tasks.md
```

Exit 0 proceeds. A non-zero exit is a plan defect: exit 1 names a task missing its `**Files:**`
field, repaired by `superpowers:writing-plans` before any dispatch happens; exit 2 stops the run.
Bundling does not change the commit-per-task model — an implementer handed a bundle still makes one
commit per task, carrying that task's own `Task-Id:` trailer, and a `Build: red` task still folds
into the commit its `**Squash-with:**` field names. Every implementer dispatch **must** carry:

> **FLOW — COMMIT-PER-TASK:** Do **not** run `git push`, merge, or open a PR. As soon as
> RED-GREEN-REFACTOR completes for this task — before the parent dispatches review for it — commit
> your work with `git commit`, carrying a `Task-Id: <n>` trailer. The trailer identifies the task;
> the subject is this task's declared `**Commit:**` field, reproduced exactly. **Never weaken or
> bypass a project's commit validation to fit** — no `--no-verify`. You **may** `git add`/`git
> commit` your own work, but never `<project>/spectre/changes/` or `<project>/docs/superpowers/`.
> **A capability spec under `<project>/spectre/specs/` is your work, not theirs**: when this task's
> `**Files:**` names one, edit it and commit it here, in this task's own commit.

**A `Build: red` task's commit folds into its green partner.** Once the partner task has its own
commit, fold the red task's commit into it: `git commit --fixup=<partner-task-sha>` followed by
`git rebase --autosquash`.

**A `Build: red` task's own dispatch records `-role red-partner`, not `implementer`.** Record it as
a pair like any other, with `-task` its own id and the end call's `-commit` the green partner's sha
as it stands after the fold.

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this
> task. Delete any code written before its test.

> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, invoke
> superpowers:systematic-debugging before writing a fix. An expected RED step needs no invocation.

> **REQUIRED READING:** `engineering-principles.md` — your implementation must satisfy these
> principles; the panel's principles reviewer checks the diff against them.

> **CONTEXT BUNDLE:** `<abs-worktree>/.superpowers/sdd/dispatch-context.md` carries this change's
> proposal, design, plan and engineering principles, gathered for you. You **must** still read the
> actual diff and the actual code — the bundle is shared *input*, never a substitute for the source.

> **PLAN PROVENANCE:** a fenced block tagged `unverified:` is a hypothesis, not code to transcribe.
> Establish the real API before writing against it, and report what you found. When what you
> measure contradicts the plan, stop and report the measurement: see **When a measurement
> contradicts the plan** (`skills/flow-contracts/plan-provenance.md`).

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one test you write MUST exercise the real
> thing. A test backed by a fake or a hand-built value passes while the real integration is broken:
> the shape you construct by hand is not the shape the real producer emits. Build the value the way
> production builds it, or assert against the real boundary.

**Guard the commit before dispatching review.** As soon as the implementer reports the task's
commit sha back, and **before** the parent dispatches that task for review:

```bash
check-task-commit-fields.sh <worktree> <task-id> <task-sha> <task-base>
```

A nonzero exit is a guard failure, not a review finding — it does **not** consume a fix-round slot.
The parent sends the task back to the **same implementer** to correct it, then re-runs the guard.

**When the script cannot be located**, apply `flow-task-commit-fields`'s rules by hand: check the
commit's `Files:` against `git diff --name-only <task-base>..<task-sha>`, its `Tests:` against the
commit's diff, and its `Commit:` against the commit's actual subject line.

**Per-task review:** the parent gives the reviewer the commit-range diff `git diff
<task-base>..<task-sha>` — a real commit diff, never a snapshot of the uncommitted working tree.
**Record the reviewer's dispatch too** — `-role reviewer`, the same `-task <n>`, and `-model
DEFAULT_MODEL`. **The per-task review is a single combined reviewer**, covering spec compliance and
code quality together, dispatched on `DEFAULT_MODEL` — `/flow`'s panel carries no roster, so there
is no `full`-preset split into two per-task reviewers. Mark a **task's** checkbox `[x]` only after
that task passes spec **and** quality review; a step's checkbox tracks the step and gates nothing.

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
