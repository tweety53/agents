# myflow-do — rationale

This file is the reasoning behind `skills/myflow-do/SKILL.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## State gate

## Superpowers Basic Workflow

## 1. Load context and validate the plan

## 2. Isolate the workspace (first run only)

## 3. Documenting a fix, before implementing it

Appending is recommended because most fixes are corrections within the change's existing scope, and
a sub-change per fix round buys a directory tree the operator has to read back. A nested sub-change
is never archived alone — it goes with its parent.

## 4. Execute (SDD + TDD)

## 5. The review panel

Free prose is not a record of a finding's state: a state that cannot be counted cannot be enforced.

### Optional slot selection

### Panel re-runs

## 6. Write the manual test guide

A guide written per plan task grows with the implementation rather than with the behaviour, which is
what made earlier guides long without making them more thorough: several entries could exercise one
behaviour while another went unlisted.

## 7. Verify, stage, and hand off

**This is the only place a project's declaration is validated, and that is why it happens here.**
The rules are mechanical, so re-deriving them by reading rows is the failure the guard exists to
remove; and the guard ships in the agents repository while the projects it judges do not, so a lint
list reaches one repository and this command reaches all of them. `/myflow-finish` does not repeat
it: run 2 reads the `survivors` row alone, and every input it cannot resolve is already a reported
skip under **Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`). Adding a
blocking validation there would strand an already-merged change over text nothing in that session
can correct — the trade
**Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`) rejects when it weighs
a change stranded short of its terminal state against stale storage.

**The `reset` is what enforces the rule; without it the `add` only assumes it** — the reason is
stated once under **Git boundaries** (`skills/myflow-contracts/pipeline.md`) and is not re-derived
here. What is specific to this command is *whose* staging it retracts (an implementer subagent's own
`git add`, or a worktree resumed with a dirty index) and why `git reset -- <paths>` is the tool:
it touches the index only, restores a tracked path to its `HEAD` entry instead of staging a deletion
the way `git rm --cached` would, and succeeds when a path is absent — which `docs/superpowers/` is
on every run that has not preserved records yet, and where `git restore --staged` would refuse the
whole command and unstage nothing.

The block below is **not** a second definition of the handoff. It is this command's rendering of the
`IN_PROGRESS`-after-`/myflow-do` template, which is defined once under
**The block each state renders** (`skills/myflow-contracts/pipeline.md`) and is canonical for the
labels, the field set and their order. What this block adds is the enumeration of the literal
alternatives `/myflow-do` writes. **Change the template first and bring this block with it** — a
field added here and not there is drift the moment `/myflow-status <name>` regenerates the same
state.

## Guardrails
