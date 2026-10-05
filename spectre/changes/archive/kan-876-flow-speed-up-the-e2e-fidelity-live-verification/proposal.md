# kan-876-flow-speed-up-the-e2e-fidelity-live-verification

**Jira:** KAN-876

## Why

On KAN-754 the end-to-end specs, the fidelity captures and the live-verification task ran strictly
in sequence and were the run's slowest group:

- The fidelity task declared `**After:**` on the spec task and the live task on the fidelity task,
  with no file dependency between them, so `plan-dispatch-groups.sh`'s chain merge put all three in
  one implementer group, run serially.
- Each started or built its own stack; visual verify then probed, could start and stopped a stack of
  its own; `flow.run-instructions` rebuilt and restarted again.
- The live walk repeated against the running app what `flow.verify` and visual verify already do.
- A spec waited out the production client timeout, and the suite ran on one worker.

## What changes

- The planner writes end-to-end spec and fidelity-capture tasks to run side by side: `**After:**`
  names only the feature tasks they exercise, each seeds its own data, and Decide splits such
  bundles into their own groups so they form one parallel wave.
- The stack those tasks run against is started once per run, by the parent, and reused by
  `flow.verify`, visual verify and `flow.run-instructions`; run-instructions leaves a stack already
  serving this worktree's build running instead of restarting it.
- No plan carries a live-verification task: `design.md`'s `## Live check` names what to exercise and
  what failure looks like, and `flow.verify` runs it, writing `live-verification.md`.
- A spec never waits out a production timing: the plan adds a test-only override, and an
  independent suite runs on the runner's parallel workers.
