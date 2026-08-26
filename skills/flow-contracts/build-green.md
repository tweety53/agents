# Build green

**This file is canonical for the build-green tag.** Skills and guards reference it by name; the
guard script's own module docstring points here rather than restating the rule — a second copy is
the same Single Source of Truth violation `check-plan-provenance.py`'s docstring warns against. If
a rule below and a skill or script comment ever disagree, this file wins.

## The build-green tag

Every task in a plan's `tasks.md` carries exactly one tag, stated as the **first matching line in
that task's body**:

- `**Build:** green` — the project builds, and this task's own verification command can run,
  given only the state left by the tasks before it.
- `**Build:** red` — this task alone does not leave a green build. A `red`-tagged task also
  carries a separate `**Squash-with:** Task <N>` field, naming one or more other tasks in the same
  plan (comma- or whitespace-separated ids, e.g. `2, 4`) that this task is dispatched
  together with as a single unit and whose commit this task's commit folds into. The field's own
  grammar still admits a dotted id, so a partner written `2.1` parses and then names no task —
  reported as the absent partner it is, rather than silently ignored.

**Placement.** A task begins at a column-0 `- [ ] <id>. <title>` checkbox line — spectre's own task
grammar, where `<id>` is a **flat integer** (`1`, `2`, `17`, …) and the mark between the brackets
carries only whether the task is done, so `- [x]` opens a task exactly as `- [ ]` does. That id is
the task's identity for every violation message and every `Squash-with: Task <N>` reference; a
dotted id on a task line is a malformed task line to spectre and no task at all here, so a sub-task
is renumbered flat rather than written `1.1`. A task's body runs from that line to the next task
line, to the next level-2 or level-3 heading — whichever comes first — or to the end of the file.
The body is everything below the task line, its `  - [ ] **Step N: …**` step checkboxes included: a
step is indented two columns beneath its task, belongs to that task's body, and is never a task of
its own, which is also what keeps spectre's malformed-task check off it.

**The two columns of indent belong to the steps alone: a task's FIELDS sit at column 0.** The
`Build:` tag is read as `^\*\*Build:\*\*\s+(green|red)\s*$` — anchored at column 0, exactly as the
task line above it is — and `**Squash-with:**` is anchored the same way, as is every field the
`flow-task-commit-fields` family adds to a task. Indenting the fields along with the steps is the
natural reading of "the body sits beneath its task", and it is wrong in a way nothing catches
kindly: `spectre validate` reports no findings, because an indented `**Build:**` line is no more a
task line to spectre than a step is, while this file's own guard reports
`task <id> has no **Build:** tag` — naming the consequence and hiding the cause, since the tag is
there, one column short of where its regex looks. Measured on a one-task plan written both ways.

Within that body, the `Build:` tag is the **first** line matching the vocabulary above; a body with
no such line has no tag at all, and a line that merely resembles one (`**Build:** yellow`) is
treated the same as no tag rather than as a separate malformed-tag class. This is the same
placement rule the guard script's own docstring states as a regex — this file states it in prose,
and the two are required to describe the same rule, the column-0 anchor included.

## The guard's scope

A guard script (`<agents repo>/scripts/check-task-build-green.py`, wrapped by
`<agents repo>/scripts/check-task-build-green.sh` to resolve which `tasks.md` files to scan) parses a single
`tasks.md` and fails the run when:

- a task has no `**Build:**` tag;
- a task is tagged `red` with no `**Squash-with:**` field naming a merge partner;
- a partner named by `**Squash-with:**` does not exist among the tasks in that same plan; or
- a partner named by `**Squash-with:**` exists but is itself tagged `red` — which is how the guard
  enforces that a `red` chain resolves to `green` rather than to another unresolved `red`: a `red`
  task's named partner must itself carry a `green` tag, not a further `red` one.

A `red` task with no partner pointing back at it, and a `red` task whose partner is `green`, are
both accepted — "unreferenced" is not a violation shape this guard checks for.

`/myflow-start` runs this guard, when the project declares one, before publishing the proposal
artifact — the same point at which the plan-provenance guard runs — and fixes any hit before
publishing.

**What the guard does not do.** The guard verifies only that every task declares a build state and
that no declared-`red` task is left unresolved. It does **not** attempt to verify that a `green`
tag is true — it never runs the build, never runs the task's own verification command, and never
checks that the state left by preceding tasks is actually what a `green` tag claims. This is the
same accepted limit as **What the guard does not do**
(`skills/flow-contracts/plan-provenance.md`): a script can confirm a claim was written down, not
that the claim is correct. Whether a `green` tag is honest is a human judgment made when writing
or reviewing the plan, not something this guard can hold anyone to.
