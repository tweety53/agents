# kan-100 — myflow: get rid of staging, use commits

**Jira:** KAN-100 — "Myflow get rid of staging, use commits"
**Date:** 2026-08-09

## Why

`/myflow-do` today implements every task uncommitted (`NO-COMMITS`), reviews it via a
checkpoint/diff snapshot (`checkpoint`, `uncommitted-review-package`), and leaves everything
staged for the human gate. `/myflow-finish` run 1 makes the only two commits the branch ever gets:
implementation, then planning artifacts.

That loses information the KAN-71 change surfaced was needed: which files a task actually touched
vs. declared, whether tests it claimed to add exist, what regresses if it's reverted, and what the
real baseline test count was — all of it currently only visible in an agent's self-report, because
nothing mechanically checks it against real state. A `Build:` tag's `red — merges with Task N`
prose is likewise never checked against anything.

Real per-task commits give guards something to check *against* — a commit's actual diff, not an
agent's claim about it.

## What changes

### 1. Per-task commits during `/myflow-do`

The `NO-COMMITS` implementer clause is replaced by a `COMMIT-PER-TASK` clause: after finishing a
task (RED-GREEN-REFACTOR complete), the implementer commits it —

```
task(<n>): <short subject>

Task-Id: <n>
```

— **before** review, not after. Review then reads `git diff <task-base>..<task-sha>` directly.
The `checkpoint` and `uncommitted-review-package` scripts are retired; nothing needs a synthetic
snapshot once work is genuinely committed.

A panel finding in a fix round is committed as `git commit --fixup=<task-sha>` and immediately
autosquashed (`git rebase --autosquash`) into the task commit before the re-review reads the diff
— review always sees one commit per task, folded, never a fixup trailing behind it.

A `Build: red` task's commit is folded into its green partner's commit the same way, driven by the
`Squash-with:` field (below) rather than by the `Build:` tag's own text.

`/myflow-do` still never pushes, merges, or opens a PR — that boundary is unchanged, only the
*committing* boundary moves.

### 2. Mechanically-checked per-task fields

New tags in each task's body in `tasks.md`, written during `/myflow-start`'s writing-plans stage
(alongside the existing `Build:` and plan-provenance tags), and — unlike those, which only assert
the *plan text* is well-formed — checked by a **runtime** guard against the *real commit*, right
after `/myflow-do` commits each task and before it dispatches review:

| Tag | Declares | Guard checks |
|-----|----------|--------------|
| `**Files:**` | paths this task touches | `git diff --name-only <parent>..<task-sha>` ⊆ declared set, or covered by `**Allowed-collateral:**` glob |
| `**Tests:**` | test names this task adds | each name exists in the commit's diff |
| `**Regression:**` | per test, what fails on revert | revert the task commit, run the named test, confirm it fails, un-revert |
| `**Baseline:**` | `before=<N> after=<N>` test counts | run the project's `## test` command at parent and at task commit, compare to declared |
| `**Squash-with:** Task <N>` | which green task this red task's commit folds into | `build-green.md`'s guard resolves the partner from here now, not from `Build: red`'s own text |
| `**Commit:**` | the subject line to use | the actual commit's subject matches |

A guard failure sends the same implementer back to fix it — a fast mechanical gate, not a full
review round.

`**Regression:**` and `**Baseline:**` degrade to **skipped, not verified** (never failed) when the
project's `## test` command can't target a named test or doesn't report a parseable count — the
same skip-not-verified shape this pipeline already uses for workspace-isolation survivor checks.

`build-green.md` changes: `Build: red` becomes bare (no inline `— merges with Task N`); the partner
is `**Squash-with:**` instead. The guard's partner-resolution logic moves accordingly.

### 3. `/myflow-finish` reshapes the branch

Per the ticket's own phrase — "commit per task on the branch, then **reshape at finish**" — the
granular history `/myflow-do` built exists for review and mechanical-field checking, not as the
final shape. Run 1 gains one step, immediately before its existing two-commit sequence:

```bash
git -C <worktree> reset --soft <recorded-merge-base>
```

This collapses every task and fixup commit back into the working tree. Everything after that is
**unchanged**: stage excluding planning paths → commit implementation → stage planning paths →
commit those. The branch still ends with exactly two commits, same as today — the difference is
they're now built from real per-task commits instead of a staged worktree.

## Decisions

### Squash everything at finish, not bisectable-per-task history

**ID:** squash-everything-at-finish
**Status:** active
**Chosen:** finish squashes all task/fixup commits into one implementation commit — final branch
history is unchanged from today's shape.
**Considered:** keeping one commit per task on the merged branch for bisectability, which the
ticket's own `Squash-with:` rationale ("keeping history bisectable") suggests. Rejected — operator
confirmed twice, explicitly aware of the tension with the ticket text, that the final shape should
stay a single implementation commit. Per-task granularity is valuable during `/myflow-do` for
review and mechanical checking; it is not wanted in the merged history.

### Fix rounds commit as `git commit --fixup`, autosquashed immediately

**ID:** fixup-autosquash
**Status:** active
**Chosen:** every fix-round commit targets its task via `--fixup=<task-sha>` and is folded in
immediately, so review always reads one commit per task.
**Considered:** leaving fix commits as separate, unfolded commits. Rejected — the ticket names
`git commit --fixup` explicitly as the reason `Task-Id:` trailers exist, and unfolded fixups would
leave the mechanical-field guards (`Files:`, `Tests:`) checking a moving target across multiple
commits per task.

### Runtime guard, not a plan-time-only guard

**ID:** runtime-mechanical-guard
**Status:** active
**Chosen:** the new `Files:`/`Tests:`/`Regression:`/`Baseline:` guard runs during `/myflow-do`,
against the real commit, immediately after each task is committed.
**Considered:** a plan-time-only check like `plan-provenance`/`build-green`, which only assert the
plan text is well-formed. Rejected — the whole point of these fields (per KAN-71 evidence cited in
the ticket) is to catch a mismatch between what was *claimed* and what was *actually committed*,
which a plan-time check can never see.

### Scope: one change, not split

**ID:** one-change-not-split
**Status:** active
**Chosen:** the commit-per-task mechanism and the mechanical per-task fields ship together in this
one change.
**Considered:** shipping commit-per-task first as its own change, with the mechanical fields as a
follow-up once real commits exist to check against. Operator chose to do both together.

## Open questions

*(none)*

## Non-goals

- No change to `/myflow-status`'s reporting shape beyond reflecting the new commit-per-task state
  where it currently shows staged/committed.
- No change to the review panel's slot roster, model policy, or escalation ladder — only what diff
  it reads changes (commit range instead of a checkpoint snapshot).
- No change to `/myflow-finish` run 2 (archive/cleanup) beyond run 1 now starting from a
  squash-in-progress worktree state, which run 2 never sees (run 1 already committed by the time
  run 2 runs).
