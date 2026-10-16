# pipeline-fix-kan-954-user-row-loaded-twice-per-request-11

## Why

The gymie `/flow` run of `kan-954-user-row-loaded-twice-per-request` reached FINISHED with a final
report still listing two open items: a self-review finding recorded `noted` with "cause in flowd
recording, not traced" (angle `flow-stats-app`: a dispatch row showing "Tokens: not measured" and
an out-of-order start), and an unverified claim that store rows were written after a store
timeout. The operator's instruction, verbatim: "update flow in a way that if the task is finished
then nothing 'still open' should left".

The agent baseline already says every noticed defect is traced to a verdict, never handed back as
"possible, not checked". The pipeline let an untraced item through anyway: step 2 of
`skills/flow-self-review/SKILL.md` let the pass return a finding with no verdict, `noted` took it,
step 6 let a store failure pass without re-reading the store, and nothing between the report and
the Finished handoff (`skills/flow/cleanup.md`) refused it.

## What changes

- `land-self-review-report.sh` — the one landing chain for every project's self-review report,
  in-run and standalone — refuses a report whose finding line says `not traced` or `untraced`:
  one `LAND-UNTRACED-FINDING` line per such line, exit 4, nothing added, committed, pulled or
  pushed.
- `skills/flow-self-review/SKILL.md`: step 2 requires every suspected defect traced to a verdict
  before the pass returns it — reproduced → fixable, disproved → `noted` with its evidence — and
  routes an untraced one to step 3's planner, whose plan opens on the trace; step 6 states a
  store row as recorded only once `flow self-review findings` lists it; step 7 names the refusal
  and its remedy.
- `skills/flow/cleanup.md`: the Finished handoff and the run's final report carry no open,
  unverified or untraced item.
- `skills/flow/integrate.md` and `skills/flow-fast/SKILL.md`: a `LAND-UNTRACED-FINDING` refusal at
  their landing call is resolved and landed again, not a stop.
