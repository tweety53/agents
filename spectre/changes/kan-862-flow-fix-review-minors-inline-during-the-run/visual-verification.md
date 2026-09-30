# Visual verification — kan-862-flow-fix-review-minors-inline-during-the-run

Stack: disposable UI-test flowd on 127.0.0.1:4174 (`make ui-test-up`), fingerprint-verified as
this worktree's build. `verify`: 47/48 — the one failure is the known stale
`runs-darwin.png` baseline (KNOWN-BUGS.md, introduced by dd995076). Capture ran on the operator's
explicit approval after the verifier's capture call was denied by the permission classifier.

## Touched views

- **Reviewers** — `../../../stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-reviewers-darwin.png`: nav marks Reviewers current; description reads
  "…and how much of that was withdrawn."; no Deferred column (the pinned period has no reviewer
  rows). The committed `reviewers-darwin.png` still showed the old "deferred or withdrawn" copy,
  passing only within tolerance — refreshed in the same commit.
- **Run detail** — `../../../stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png`: five stat panels (Runs, Measured, Total cost,
  Total input tokens, Total duration), no `Deferred minor` panel; timeline and stage-run table
  intact. The committed `run-detail-darwin.png` already lacked the panel (older nav), unchanged.

## Full app suite (created — 9 screens, capture exit 0)

- state-board, stage-leaderboard, trend, cache-efficiency, decisions, runs, flow-health: each
  renders with data or its empty state, dark palette, nav marking the screen current; no defect.
