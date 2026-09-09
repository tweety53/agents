# kan-252-derive-suite-runtime-figures-from-recorded

## Why

`.flow/project.md`'s `## test` section pastes a measured suite runtime into prose
("about 52 to 56s wall", with `<!-- measured: ... -->` provenance comments). The
measurement has no owner and nothing checks it: it went stale once already (the
118.63s figure KAN-252 was filed against survived two reworkings before anyone
noticed), and every guard harness added since silently invalidates it again. A
written duration is the same failure mode the file's own notes warn about for
written counts, applied to a duration.

## What changes

Suite runtimes become recorded data in the myflow store — per run, per suite, per
machine — instead of pasted prose:

- New `suite_runs` table (project-scoped; suite, host, duration_ms, exit_code,
  ran_at) via migration `0019_suite_runs.sql`.
- `flow suite record -suite <name> -- <command…>` — CLI wrapper that times any
  command, records the duration + hostname + exit code, and passes stdout/stderr
  and the exit code through untouched.
- `flow suite list [-suite <name>] [-limit N] [-json]` — recent rows plus a
  per-(suite, host) summary: median duration of the last 10 passing runs.
- `.flow/project.md`'s `## test` prose cites `flow suite list` instead of
  carrying a number; canonical suite names (`guard-tests`, `stats-go`,
  `stats-spa`) are named there.

The harness's default tool-timeout advice stays, anchored to recorded figures. The
interim step KAN-252 proposed (hand-correct the current number) is moot: the
mechanism and the prose replacement land in the same change.
