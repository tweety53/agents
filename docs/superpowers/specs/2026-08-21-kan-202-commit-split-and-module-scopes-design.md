# Commit scopes name the module — design

**Change:** `kan-202-commit-split-and-module-scopes`
**Issues:** KAN-202 (linked), KAN-254 (also closed by this change)
**Date:** 2026-08-21

## The problem

This project's commit convention is `<type>(<scope>): <subject>` where `<scope>` is the **module or
area inside the repository** — `finish-contract`, `gather-self-review-context`, `stages`, `openspec`,
`scripts`. The scope exists so a reader scanning `git log` sees *what part of the codebase moved*.

The myflow pipeline instructs the opposite in three places, and the result reaches `main`:

```text
b0b9ae1 chore(kan-239-run-2-asserts-base-branch-and-archives-via-pr): plan and session records
981fc20 feat(kan-239-run-2-asserts-base-branch-and-archives-via-pr): assert run 2's base branch …
```

A 50-character scope that repeats what the branch and the ticket already say, and names no part of
the codebase at all.

### The three sites

| Site | What it says today | Scope it produces |
|------|--------------------|-------------------|
| `skills/myflow-do/SKILL.md:232` | "where that convention has a scope, `<n>` is the scope" | the dotted task id — `feat(3.2): …` |
| `skills/myflow-contracts/pipeline.md:157` | `commit -m "chore(<name>): plan and session records"` | the change name |
| `skills/myflow-start/SKILL.md:403` | "`**Commit:**` — the commit subject line this task's implementer must use" | unstated, so every plan written under it uses the change name |

`pipeline.md`'s literal is realised twice more: in `scripts/commit-split.sh`'s two call sites,
`skills/myflow-finish/SKILL.md:220` and `skills/myflow-do/SKILL.md:1041`, each of which passes
`"chore(<name>): plan and session records"` as the fourth argument. The implementation commit's
message at those same call sites is `"<type>(<name>): <what the implementation does>"` — the same
defect, on the commit that carries the actual work.

### Why it cannot be fixed at commit time

`scripts/check-task-commit-fields.py`'s `check_commit_subject` compares the **real commit subject**
against the task's declared `**Commit:**` field, byte for byte. An implementer who writes the
correct module scope against a plan declaring a change-name scope **fails the guard**. The plan and
the guard have to agree, so the module scope must be chosen and written at planning time, in
`/myflow-start`'s writing-plans stage.

This is what makes the contradiction actionable rather than cosmetic: today, following the stated
preference breaks a green build.

### What actually reaches `main`

`/myflow-finish` run 1 reshapes the branch — `git reset --soft <recorded-merge-base>` — collapsing
every per-task and fixup commit back into the working tree before the two-commit chain runs. So the
scopes a reader scanning `git log main` sees come from **finish's two messages**, not from the
`**Commit:**` fields. The per-task subjects survive only when a PR was opened first and later fix
rounds pushed on top of it, which is why `kan-102` has ten commits on `main` and `kan-239` has two.

Both sources are fixed here. The finish messages matter more than KAN-254's description implied.

## KAN-202: the symptom was stale

KAN-202 states that `/myflow-fast` cites `finish-contract.md` for run 1 and therefore hand-writes the
guarded `&&` chain instead of calling `commit-split.sh`. That is not what the file says, and was not
what it said on the day the ticket was filed: `skills/myflow-fast/SKILL.md`'s "No argument (bare
invocation)" section routes run 1 through **1.2** and **1.3 of `skills/myflow-finish/SKILL.md`**, and
that file calls the script at line 220. Verified against the file as it stood at `05bac7a`
(2026-08-18), the commit contemporaneous with the filing.

**No live path hand-writes the chain.** What remains true is the ticket's *proposed fix*:
`skills/myflow-contracts/finish-contract.md`'s two-commit section describes the chain — the clearing
pass, the pathspec exclusion, the second bare `add` — and cites `pipeline.md`'s Git boundaries for
the sequence, without ever naming the script that implements it. A reader treating the contract as
the authority still has nothing to call.

That two-line edit is folded into this change because these files are being edited anyway. It is
defensive documentation against a failure mode that currently has no route, and it is not the reason
this change exists.

## The design

### 1. Finish run 1's two messages

The implementation commit derives its scope from the reshaped diff — the module carrying the
change's substance, or a broader area when it genuinely spans several, never a list. The planning
commit becomes a **fixed literal**, since it always touches the same two trees:

```text
feat(finish-contract): name commit-split.sh at the two-commit step
chore(openspec): plan and session records
```

`chore(openspec)` is fixed rather than derived because there is nothing to derive: every planning
commit stages `openspec/` and `docs/superpowers/`, in every change, forever. A derived value would
compute a constant.

Changed at four sites — `pipeline.md`'s Git-boundaries chain, `finish-contract.md`'s two-commit
section, and the two `commit-split.sh` call sites in `myflow-finish/SKILL.md` and
`myflow-do/SKILL.md`.

### 2. The `**Commit:**` field spec

`skills/myflow-start/SKILL.md`'s field list gains the scope rule: the scope names the module or area,
derived from the paths in the task's own `**Files:**` field. Where a task spans modules, name the one
carrying the substance or a broader area — never a list, never the change name, never the task id.

`skills/myflow-do/SKILL.md`'s COMMIT-PER-TASK block drops "`<n>` is the scope". The `Task-Id: <n>`
trailer already identifies the task; the subject follows the plan's declared `**Commit:**` field.

### 3. The guard

`check-task-commit-fields.py` gains `check_commit_scope(task, change_name)`, called from
`check_task_commit` alongside the existing `check_commit_subject`. It parses the scope out of the
declared `**Commit:**` field and fails the task when that scope is:

* the change name (`kan-202-commit-split-and-module-scopes`);
* the change name's bare Jira key (`kan-202`); or
* a dotted or numeric task id (`3`, `3.2`, `12.4.1`).

Anything else passes. **A scope is optional** — a `**Commit:**` field declaring `fix: <subject>` with
no scope at all passes, matching what this repository already does (`e61603b fix: drop an invented
spec citation`). Only a scope that is *present and wrong* fails.

The change name is derived from `tasks_md_path`, which is already
`<worktree>/openspec/changes/<name>/tasks.md` — the guard's wrapper resolves exactly that path, so
no new argument is needed.

Per this repository's convention that every guard carries a mutation test, the check gets cases in
`scripts/test-check-task-commit-fields.sh`.

### 4. The always-on rule

`rules/commit-scope-is-the-module.mdc`, with `alwaysApply: true` and `<!-- core -->` / `<!-- /core -->`
markers, so `setup.sh` renders the core into every managed `CLAUDE.md` block and installs the full
text at `~/.claude/rules/commit-scope-is-the-module.md`. **`setup.sh` needs no edit** —
`always_on_rules()` discovers a rule from its frontmatter.

`rules/agent-baseline.md`'s table gains a matching row, so the convention reaches dispatched agents
at every depth.

## What is deliberately not done

* **The finish commits stay unguarded.** Their messages are written by the agent at integration time
  from a diff, not declared in a file the guard reads. The rule is stated in the contract; nothing
  checks it. Guarding a message derived at runtime would mean re-deriving the module from the diff
  inside a guard, which is the same judgment call the agent already made.
* **History is not rewritten**, and archived plans under `openspec/changes/archive/` are not
  repaired. The change governs plans written from here on.
* **No scope vocabulary is introduced.** There is no list of legal module names to maintain; the
  guard rejects three specific wrong shapes and accepts everything else.

## Testing

| Layer | How |
|-------|-----|
| `check_commit_scope` | New cases in `scripts/test-check-task-commit-fields.sh`: change-name scope rejected, bare-key scope rejected, dotted task id rejected, module scope accepted, absent scope accepted, and a mutation case proving the check fires. |
| The four message sites | Covered by this change's own plan: its tasks declare module scopes, so a green `check-task-commit-fields.sh` run on each task commit is the check. |
| The rule | `scripts/test-setup.sh` — a sandboxed `setup.sh global` must render the new rule's core into the managed block and install its full text. |
| Repository lint | The full `## lint` list, notably `check-references.sh` and `check-installed-citations.sh` for the new citations, and `check-contract-budget.sh` for the edited contract files. |
