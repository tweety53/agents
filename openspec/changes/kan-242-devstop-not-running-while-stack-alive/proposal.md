# Verify the stop took effect before removing a worktree, and stop inferring `not running`

**Jira:** KAN-242

## Why

`/myflow-finish` run 2 removed two worktrees while four processes launched from them were still
alive, holding 9931, 8080, 3000 and 4851. Two of those are the OAuth-pinned ports every workspace
shares, so the leak blocks the next `devStart` in any worktree. Once the worktree is gone the
project's own `devStop` can never reach those processes — it reads pid files from the worktree's own
`.dev-state`, removed with the directory. The only remedy left is killing pids by hand.

This is the **second** occurrence. `.myflow/project.md`'s `## stop` key was added after the first
(KAN-129) so run 2 could stop the stack before removing a worktree. It did run the stop command.
Nothing checked whether the stop worked.

Two independent defects made that possible:

- **`./gradlew devStop` reported `not running` for all four live services.** A stop that cannot see
  the process it started is worse than one that fails loudly.
- **Run 2 has no gate on live processes.** Its cleanup verdict checks worktrees, branches and
  workspace resources — nothing on this machine's process table. `devWorkspaceInventory` caught the
  leak, by chance, after the fact; it was the only control that did.

## What changes

**In `agents` — run 2 gains a live-process gate.**

- A new guard `scripts/check-worktree-processes.sh <worktree>` reports every process whose working
  directory is at or under a worktree path, from one `lsof -d cwd` pass. Project-agnostic: no ports,
  no build tool, nothing for a project to declare.
- Worktree cleanup gains **check 6**, running after check 5's `## stop` command and before any
  removal. A live process fails the check; every worktree is left alone and run 2 stops at
  `IN_PROGRESS`, naming the pids and the command that reaches them.
- An incidental correction in the same block: the "any failed check leaves every worktree alone"
  bullet cites the `## stop` command as *check 4*; it is check 5.

**In `gymie` — `devStop` stops inferring `not running` from missing records.**

- `stopService` may print `not running` only on a definite negative from both halves. Anything
  short of that becomes a third outcome naming which half was unknown and the command that answers
  it.

## Impact

- **Affected specs:** `myflow-finish-cleanup`
- **Affected code:** `scripts/check-worktree-processes.sh` and its test harness (new);
  `skills/myflow-contracts/finish-contract.md`; `gymie`'s `gradle/dev-lifecycle.gradle.kts` and
  `buildSrc/src/main/kotlin/com/gymie/gradle/`
- **The gymie half ships with no delta spec.** The two repositories hold disjoint capability sets and
  one change has one change directory. Recorded as decision `change-directory-in-agents` in
  `design.md`, with the cost stated.
- A project whose stack legitimately runs from a worktree at finish time will now be stopped rather
  than silently orphaned. That is the intended behaviour change, and it is a gate the operator
  clears by stopping the stack and re-running.
