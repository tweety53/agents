# Design — pipeline-fix-kan-1060-flow-cost-visual-verify-now-costs-more-than-49

## Context

Authoritative run at `81ca7e1b`, from `stats/`: `make -s policy-copy && make ui-test-up`, then
`cd web && npx playwright test` — 13 failed, 62 passed:

| Spec | Failure | Cause |
|------|---------|-------|
| `controls.spec.ts:264` Project chip narrows | no `uitest-beta` chip | chips derive from `inPeriod` rows (`StateBoard.tsx` `projects`); finished rows load only when the Finished group opens or its tile is pressed (`70594ba2`, `wantFinished`), and every `uitest-beta` change is finished |
| `controls.spec.ts:286` Updated All widens | Finished tile stays `—` | intended by `70594ba2`: the tile reads `—` until finished changes load, and the Updated chip is not a load trigger; the spec never loads them |
| `interactive.spec.ts:63` run view | `getByRole('region',{name:'Main'})` strict-mode violation | the grouped totals added a `Main totals` region (`ef13b41f`/`a26b30f4`); the substring name now matches two regions |
| `baseline.spec.ts` dashboard; `full-app-suite.spec.ts` state-board | screenshot | intended `70594ba2`: Finished tile, Cost in period and the Finished group count read `—`; plus the lost `uitest-beta` chip (fixed below, re-gains it with a `—` count) |
| `combinations.spec.ts` chip plus tile | screenshot | follows the chip fix: re-record once the chip exists and its press loads finished rows |
| `baseline.spec.ts` run-detail, run-detail-hover; `full-app-suite.spec.ts` run detail | screenshot | intended run-view features since `9ace8801`: grouped Main · Self-review totals panels with fix rounds, full panel runs, reviewer runs and C · I · M findings (`ef13b41f`, `425a4779`, `a26b30f4`, `c891fe0d`, `998863e6`, `b8ee4473`); the Failed runs KPI gone from the header row; Decisions panel under Cost by stage (`2c58b1bf`); phase header dot dropped (`01a5d393`) |
| `interactive.spec.ts` running/waiting/hovered/focused cards, run view | screenshot | intended: interactive-card and run-view changes `188083a4`, `d23cfdca`, `a249ffe4`, `b383cf14`, plus the `—` tiles above |

Diff PNGs inspected for dashboard and run-detail: only the moves named above.

`make -C stats ui-test-up` (`stats/Makefile:397`) depends on `web-build check-uitest-overrides` but not
`policy-copy`, which `test`, `web-dist` and `build` all carry; its `go build ./cmd/flowd` then fails on
`internal/modelpolicy/embedded`'s `//go:embed model-policy.md` in any worktree that never ran another
make target.

`store.ProjectRoots` (`stats/internal/store/projects.go`) already reads every `projects` row in key order.

## Decisions

**ID:** ui-test-up-policy-copy
**Status:** active
**Chosen:** add `policy-copy` to `ui-test-up`'s prerequisites — the same fix `test`/`web-dist`/`build` carry.
**Considered:** documenting a manual `make policy-copy` first — rejected: a target that cannot build on
its own is the defect.

**ID:** project-chips-name-every-project
**Status:** active
**Chosen:** a read-only `GET /api/v1/projects` returns `{"projects":[<key>...]}` from `ProjectRoots`.
`StateBoard` fetches it once on mount. Until the finished changes have loaded, the Project chips are the
union of those keys and the in-period keys, each count reading `—`; once loaded, chips and counts derive
from `inPeriod` exactly as before. Choosing a project chip sets `wantFinished`, as the Finished group and
tile already do. A failed projects fetch falls back to the in-period keys.
**Considered:** reverting the board's on-demand finished read — rejected: `70594ba2` made it on purpose
and its unit tests pin it. Updating the spec to open the Finished group first — rejected: leaves a
user-visible chip missing. Adding `projects` to the state-board view response — rejected: that view's
generic rows-only shape is shared by every stats view.

**ID:** stale-specs-follow-intent
**Status:** active
**Chosen:** `controls.spec.ts:286` opens the Finished group (click the `▾Finished` row) before reading
the tile, so the period widening is measured on loaded rows; `interactive.spec.ts:63` passes
`exact: true` to the `Main` region. Every screenshot in the Context table is re-recorded only after the
fixer opens its diff PNG and sees no move beyond the ones named there; any other move is a regression to
fix, not re-record.
**Considered:** making the Updated chip a load trigger — rejected: `Clear filters` would then not restore
the `—` default the spec compares against, and `70594ba2` names only the group and the tile.

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — a new API route plus a state-dependent UI branch and re-recorded baselines
that a reviewer must check against named intent.
**Considered:** `flow-low` — production UI code moves, so not enough.

**ID:** fixer-grouping
**Status:** active
**Chosen:** one fixer for tasks 1–3 in order, `opus` on `flow-medium`, one commit per task: task 3's
re-records depend on task 2's chips, and both need task 1's stack.

## Open questions

None.
