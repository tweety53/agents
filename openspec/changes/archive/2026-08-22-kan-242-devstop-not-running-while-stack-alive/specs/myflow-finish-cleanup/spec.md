## ADDED Requirements

### Requirement: Worktree cleanup verifies the stack actually stopped, before removing anything

`/myflow-finish` run 2 SHALL NOT remove a worktree while a live process is running from it.

After running the project's `## stop` command, and before any removal, run 2 SHALL check each
worktree in the resolved set for processes whose working directory is at or under that worktree's
path. This check SHALL run whether or not the project declares a `## stop` command, and whether or
not that command reported success: a project that declares no stop command can still have a stack
running from its worktree, and a stop command that exits 0 is not evidence that it worked — which
is the whole of what this requirement adds.

The check SHALL be answered from the operating system's own process bookkeeping, never from a file
the project wrote. The records a project keeps about what it started live inside the worktree being
removed, so they are exactly the evidence that a removal destroys.

The check SHALL be project-agnostic. It SHALL NOT read a port list, invoke a build tool, or require
a project to declare anything, so that it protects every project myflow is installed into rather
than only those that opt in.

**A working directory, not a port, SHALL be what identifies the process.** A port proves nothing: a
live workspace legitimately holds one. What a process launched from a worktree has, and what ties it
to the directory about to be removed, is its working directory.

#### Scenario: A live process blocks the removal

- **WHEN** run 2 finds a process whose working directory is at or under a worktree it is about to remove
- **THEN** the check fails, **no** worktree is removed, run 2 stops at `IN_PROGRESS`, and it reports
  each process's pid, its working directory, and the command that reaches it

#### Scenario: A clean worktree proceeds

- **WHEN** no process has a working directory at or under any worktree in the resolved set
- **THEN** the check passes and cleanup proceeds to removal

#### Scenario: The check cannot be answered

- **WHEN** the check cannot run at all — the scanning tool is absent, or a worktree path cannot be read
- **THEN** it is treated as a failed check, not a passed one: no worktree is removed and the reason is
  reported

#### Scenario: The stop command reported success but did not work

- **WHEN** the project's `## stop` command exits 0 and a process from the worktree is still alive
- **THEN** the check fails on the live process, and the stop command's exit code SHALL NOT be read as
  evidence that the stack stopped

### Requirement: A blocked removal is a failure, never a confirmable disclosure

A live process found by the check above SHALL fail the check outright. Run 2 SHALL NOT offer the
operator a prompt to proceed with the removal anyway.

This is deliberately unlike the ignored-files disclosure, which reports what `--force` will destroy
and asks. That disclosure is safe to confirm because the operator can see what is at stake and
decide. A live process is different in kind: confirming it destroys the only records that can reach
the process afterwards, and the resulting orphan holds ports shared across every workspace. Both
occurrences of this incident reached the operator as something to click through.

`/myflow-fast`'s override of the ignored-files ask SHALL NOT extend to this check. That override is
justified by the records worth keeping already being out of the worktree; a live process is not a
preserved record, and an unattended run is the case most likely to orphan one.

#### Scenario: No override is offered

- **WHEN** the check finds a live process, under `/myflow-finish` or `/myflow-fast`
- **THEN** the run stops with no prompt to proceed, and the operator clears the process and re-runs
