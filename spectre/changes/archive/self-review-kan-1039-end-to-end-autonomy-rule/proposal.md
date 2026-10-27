# self-review-kan-1039-end-to-end-autonomy-rule

## Why

The self-review pass of `kan-1039-end-to-end-autonomy-rule`
(`.superpowers/sdd/self-review-pass.md` in this worktree, which carries each finding's evidence)
found nine non-`big` pipeline defects:

- 1.1 Four places still tell the parent to record a round's `begin` rows in the Bash call after
  the launch message, while `hooks/enforce-agent-baseline.py` denies any launch whose row is not
  open yet.
- 1.2 `hooks/protect-main-checkout.py` reads a backtick inside single quotes as a command
  substitution, so a data heredoc followed by a single-quoted backtick has its body scanned as
  shell and is denied.
- 1.3 `pipeline-fix.sh verify` never runs the tests that scan every shell script of the tree for
  a touched script, so a loop landed a script that turned `TestNoPipeIntoEarlyExitGrep` red on
  main.
- 1.4 The pass-1 full suite runs while the mutation slot runs the tests per mutant on the same
  CPUs, and the guard package hit `go test`'s timeout.
- 1.5 A fix the planning session lands has no narrative to record its line in, and integrate's
  `commit` stops on the missing line although the store row holds every field of it.
- 1.6 The fix-mutation row of a fixed Critical/Important finding is checked only at the landing
  gate, not when the panel closes.
- 2.1 An inline `full-suite-fix-<n>` back-fill starts at the planning commit before the panel,
  so its row claims the whole panel wait.
- 5.1 No pipeline step records suite runtimes, while `.flow/project.md` `## test` says they are
  recorded per run.
- 6.1 The mutation brief does not say that a cached test result misses a mutated file a test
  reads through a subprocess.

## What changes

- Every prescription of the begin order states begins before the launch message (1.1).
- The main-checkout hook tests for substitutions outside single-quoted spans only (1.2).
- `pipeline-fix.sh verify` adds the script-scanning tests for a touched `setup.sh` or
  `scripts/**/*.sh` (1.3).
- Pass 1 launches the `mutation+failure-modes` dispatch only once the suite has exited (1.4).
- `integrate-change.sh commit` appends a missing in-run fix line from its store row, and stops
  only when the row lacks the note or the blast radius (1.5).
- `check-panel-findings-closed.sh` refuses the close while a fixed Critical/Important round has
  no fix-mutation row (1.6).
- A `full-suite-fix-<n>` back-fill is floored at the newest panel reviewer end before its
  commit, and this change's mis-stamped row is repaired (2.1).
- `run-verify-list.sh` records each `# suite:`-declared `## test` command's runtime through
  `flow suite record` (5.1).
- The mutation brief requires the no-cache flag on every mutant's test run (6.1).
