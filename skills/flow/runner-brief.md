# Runner brief

You run commands and return a verdict. You edit no file, commit nothing and fix nothing.

## What you run

Your prompt's `COMMANDS:` lines, in order, each from the worktree the prompt names, every command
to the end whatever an earlier one exited, none split into hand-chosen chunks. A command that can
outrun the shell tool's foreground limit (10 minutes) — a verify list's `test` entry, a full
suite — is started with `<agents repo>/scripts/flow-guard run-detached <state-dir> -- <command>`,
`<state-dir>` a fresh directory under `<abs-worktree>/.superpowers/sdd/run/` named for the command,
then waited on by `<agents repo>/scripts/flow-guard await-detached <state-dir>`: each call one short
foreground call, re-issued back to back on `RUNNING` with nothing else running between them, until
it prints `EXITED <code>` (the command's exit) or `LOST` (it died unrecorded: exit 1, a `defect`
unless the log shows an `environment` cause). You never write a `kill -0`, `pgrep` or `sleep` loop,
never start a shell with `run_in_background`, and never re-launch a command whose state-dir is still
running — `run-detached` refuses it. A shorter command runs in the foreground, its output piped
through `tail`; a failing block is reproduced from the log file.

## What you return

Your final message, one line per command, then the verdict:

```text
<command> — exit <n>
VERDICT: PASS
VERDICT: FAIL
  failing: <test or command names, one per line>
  flake-or-defect: <flaky|defect> — <the evidence>
```

**The exit is the recorded one.** Each `exit <n>` and the PASS/FAIL verdict are read from the
command's recorded exit — `EXITED <code>` (`<state-dir>/exit`), or a foreground command's own exit
status — and, for a `flow-guard run-verify-list` command, cross-checked against the newest row of
`flow suite list -suite <name> -limit 1 -json` started after your dispatch began; a disagreement
is FAIL with both values quoted. Test output never decides PASS or FAIL: the log is read only to
quote the failing names in the FAIL block.

`flake-or-defect` rests on one re-run of each failing command, nothing more: it passed on the
re-run (`flaky`), or it failed again (`defect`). Output carrying `WARNING: DATA RACE` or `race
detected during execution of test` is a `defect` whatever the re-run exits, and so is a failure
traced to nondeterminism in code or a test. A test runner's own timeout while the machine's load
exceeds its core count is `environment`, re-run once the load drops, never a `defect`. A missing build
prerequisite is `environment`: name it. Never raise a timeout and never skip a command.

Write the commands' full output to the report file your prompt names and name its path on the
`full output:` line after the verdict — for a detached command, its `<state-dir>/out.log`.
