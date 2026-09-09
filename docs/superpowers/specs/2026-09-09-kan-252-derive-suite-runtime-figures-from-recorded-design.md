# kan-252 — suite runtimes recorded in the store, cited by project.md

Approved design record — 2026-09-09. The change's canonical design is
`spectre/changes/kan-252-derive-suite-runtime-figures-from-recorded/design.md`; this file is
what the operator approved at the design gate.

## Problem

A measured suite runtime pasted into `.flow/project.md`'s `## test` prose goes stale
invisibly. The 118.63s figure KAN-252 was filed against survived two reworkings; kan-362's
"52 to 56s" replacement is pasted prose with the same defect — no owner, nothing checks it,
and every guard harness added since silently invalidates it.

## Design

- **Data.** Migration `0020_suite_runs.sql`: project-scoped `suite_runs` table — `suite`,
  `host`, `duration_ms` (CHECK >= 0), `exit_code`, `ran_at`; index on
  `(project_key, suite, ran_at)`. Not change-scoped: suites run outside changes too. The
  exit code is recorded so a failed run's duration never presents itself as a runtime
  figure.
- **Record.** `flow suite record -suite <name> -- <command…>`: child stdio passes through
  untouched, wall-timed, row recorded with the hostname, child's exit code returned as the
  command's own exit. Store unreachable → one warning line, exit unchanged — never a gate.
- **Read.** `flow suite list [-suite <name>] [-limit N] [-json]`: rows newest first plus
  one summary line per (suite, host) — the median of the last 10 passing runs, the figure
  project.md cites. Median so one cold-cache run cannot move the number.
- **Consume.** `.flow/project.md`'s `## test` section cites `flow suite list` and carries
  no number; canonical suites named there: `guard-tests`, `stats-go`, `stats-spa`. The
  tool-timeout advice stays, anchored to recorded figures.
- **Layers.** Store accessors, `POST/GET /api/v1/suites/{project}/runs`, client methods —
  each following the hazards precedent (migration → store → routes → CLI).

## Decisions made with the operator

1. **Cite the app only** — no drift guard. A guard would make every lint run depend on the
   daemon being up; cite-only removes the stale figure outright, and a guard can be added
   later if recorded figures go ignored.
2. **Record via a CLI wrapper** — harness self-reporting would touch three surfaces;
   recording inside `flow verify` would leave figures only for change-time runs.
3. The issue's interim step (hand-correct the current number) is moot — the mechanism and
   the prose replacement land in the same change.
