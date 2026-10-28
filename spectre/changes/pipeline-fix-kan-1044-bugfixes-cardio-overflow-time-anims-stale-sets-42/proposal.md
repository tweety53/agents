# pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-42

## Why

`stats/web` runs `vitest run` on vitest's default `forks` pool: one child process per CPU (about 14
on the operator's laptop, about 180 MB each). When the parent vitest is killed — a Bash timeout, a
stopped agent, an ended session — the fork workers are re-parented to PID 1 and keep about 2.5 GB
until killed by hand. Observed 2026-10-28 from a kan-1047 worktree, on a laptop already out of
memory.

## What changes

- `stats/web/vite.config.ts` sets `test.pool: "threads"` and `test.maxWorkers: 2`: worker threads
  live inside the parent process and die with it, and two workers cap the suite's footprint.
- A new vitest test, `stats/web/src/viteConfig.test.ts`, imports the config and asserts both values,
  so reverting either fails the suite.
