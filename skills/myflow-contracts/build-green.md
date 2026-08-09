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
  plan (comma- or whitespace-separated dotted ids, e.g. `2.1, 3.4`) that this task is dispatched
  together with as a single unit and whose commit this task's commit folds into.

**Placement.** A task begins at a `### <dotted-id> …` heading — a level-3 heading whose text
starts with a dotted id (`1`, `2.1`, `9.9.2`, …), which is that task's identity for every
violation message and every `Squash-with: Task <N>` reference. A task's body runs from that
heading to the next level-2 or level-3 heading — whichever comes first, including a `###` aside
that is not itself a task heading — or to the end of the file. Within that body, the `Build:` tag
is the **first** line matching the vocabulary above; a body with no such line has no tag at all,
and a line that merely resembles one (`**Build:** yellow`) is treated the same as no tag rather
than as a separate malformed-tag class. This is the same placement rule the guard script's own
docstring states as a regex — this file states it in prose, and the two are required to describe
the same rule.

## The guard's scope

A guard script (`scripts/check-task-build-green.py`, wrapped by
`scripts/check-task-build-green.sh` to resolve which `tasks.md` files to scan) parses a single
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
(`skills/myflow-contracts/plan-provenance.md`): a script can confirm a claim was written down, not
that the claim is correct. Whether a `green` tag is honest is a human judgment made when writing
or reviewing the plan, not something this guard can hold anyone to.
