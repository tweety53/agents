# pipeline-fix-kan-1060-flow-cost-visual-verify-now-costs-more-than-49

## Why

`origin/main` (`81ca7e1b`) is red on its own Playwright visual suite (`stats/web`, `npx playwright test`
against `make -C stats ui-test-up`): 13 of 75 fail. Two causes:

- `70594ba2` (load finished changes on demand) derives the live state board's Project chips from rows,
  so a project whose changes are all finished (`uitest-beta`) has no chip until the Finished group is
  opened. The chip set no longer names every project — an unintended loss; `controls.spec.ts:264` times
  out on it.
- Several stats commits since the last re-record (`9ace8801`) changed the run view and the board on
  purpose without touching `stats/web/tests/visual/`, leaving baselines and two assertions stale.

Separately, `make -C stats ui-test-up` on a fresh worktree fails to build flowd (`pattern
model-policy.md: no matching files found`): the target lacks `policy-copy` as a prerequisite.

## What changes

- `stats/Makefile`: `ui-test-up` depends on `policy-copy`.
- `GET /api/v1/projects` lists every registered project key; the state board names those as Project
  chips until the finished changes load, and pressing a chip loads them.
- `controls.spec.ts` and `interactive.spec.ts` follow the intended UI; the stale screenshots are
  re-recorded, each diff named in `design.md`.

## Impact

`stats` only: one Makefile line, one read-only API route, `StateBoard.tsx`, visual specs and baselines.
The on-demand finished read of `70594ba2` stays.
