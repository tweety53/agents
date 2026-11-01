# self-review-kan-1051-rework-the-trend-dashboard

## Why

The self-review pass of `kan-1051-rework-the-trend-dashboard` found six fixable, non-`big` defects
(evidence: `findings.md` under
`/Users/tweety53/Agents/flow/state/agents-a740d89c/self-review/kan-1051-rework-the-trend-dashboard/`):

- F4: after a fix run restores a dropped change directory, `skills/flow/review-panel-late-fix.md` still
  names the drop commit as the since-close sha, so the restored directory counts as a 12,397-line delta
  and the small append never fires.
- F5: a panel reproducer wrote its Go test file into the shared worktree, and a concurrent guard-test
  run failed on it; only the test-sanity brief demands a scratch copy.
- F6: step 2 of `skills/flow-self-review/SKILL.md` has the pass write `findings.md` and
  `report-draft.md`, which the harness refuses a subagent on every pass.
- F7: dispatch row 9234 (verify-5) carries `metrics: {}`. The finding's premise — records of the run
  in its window left uncredited — is disproved: those records are another session's. The row is still
  wrong, because the harvester's loop parser does not end a `for` body at `& done`, so the settle pass
  never zeroed it.
- F17: `scripts/live-store-copy.sh` has no `restart`; a hand relaunch exited on the old pidfile.
- F18: on merge and push, `pipeline-fix.sh land` runs the same verify the loop has just run.

## What changes

- The late-fix since-close sha is the fix-run planning commit that restored the change directory,
  when one follows the newest drop commit.
- A reproducer writes nothing under `$PWD`; a Go reproducer adds its test through `go test -overlay`;
  `check-dispatch-paragraphs` requires that sentence in the REPRODUCER paragraph.
- The self-review pass hands back its files' full contents, and this session writes them to `<sr-dir>`.
- `scanLoopBody` ends a loop body at `;`, `&` or a newline before `done`; row 9234 is restamped to
  zero tokens.
- `live-store-copy.sh restart <name>` restarts only the copy's flowd.
- On merge and push, `pipeline-fix.sh land` is the loop's verify; the separate verify runs only where
  nothing lands through it.
