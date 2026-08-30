# kan-291-myflow-add-myflow-tasks-tick-so-ticking-a

## Why

`tasks.md`'s checkbox is the pipeline's single source of task completion
(`check-unfinished-work.sh` gates the finish run on it), yet ticking one after
its review passes is a hand-edit today — an ad-hoc script that finds the
task's heading, walks to the next heading, and rewrites `- [ ]` to `- [x]` in
that span. Done by hand twelve times in `kan-258-store-native-run-record`
alone. The failure modes are quiet: ticking the wrong span, ticking a task
whose review has not passed, or leaving one unticked so
`check-unfinished-work.sh` reports OUTSTANDING for work that is done.

## What changes

- New `flow tasks tick [-C dir] <change> <task-id>` subcommand
  (`stats/cmd/flow/tasks.go`), wired into `main.go`.
- Given a task id, it flips that task's own checkbox and every step checkbox
  in its body from `[ ]` to `[x]`, leaving every other line — including other
  tasks' bodies — untouched.
- Refuses (exit 1) rather than no-ops when the task is already ticked, so a
  double-tick is visible.
- Refuses (exit 1) when the task id does not exist in the plan.
- Never records completion anywhere but `tasks.md` itself — no second source
  of completion state.
