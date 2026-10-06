# Gate the draw at the root, never enumerate exits

When a property must hold at every exit from a state — a reset, a revocation, a
teardown — enforce it where the thing is drawn or used, re-checked on every draw, instead
of resetting it at each place the state can end. Enumerating exits is open-ended: the state
machine grows, a new exit appears, and each one the enumeration missed is a live leak nobody
listed. A draw-site gate has no exits to miss — whatever path reached here, the check runs
now.

## Why

kan-517's round 2 closed "the overlay survives an account switch" by resetting
`FrameRateSetting` at every place a session or account could end — threading a
`BooleanPreferences` parameter through `SwitchState`, `rememberSwitchState` and two test
fixtures — and still missed a third exit: `SessionRefresher`'s `session.clear()` on a refused
refresh. Round 3 replaced the shape instead of extending it: gate the overlay's draw on
`isFrameRateAllowed(ownEmail)` at the composition root, re-checked every composition. That
closes every exit path there is, including ones nothing explicitly resets, and let the
parameter threading and its test fixtures come back out entirely.

## How to apply

- The smell is reaching to thread a reset parameter through every holder of a state, or to
  list the places a state can end. The listing is the defect being written, not the fix.
- The gate is the check placed at the draw or use site — the composition root, the render
  function, the handler that serves the thing — evaluated every time the thing is drawn,
  not cached from an earlier pass.
- Prove the gate load-bearing by mutation: hardcode the gate to the failing value and watch
  the suite fail. A mutant nothing kills means the gate guards nothing (kan-517's suite hung
  with the gate pinned to `true`, which is the proof the gate is what holds it shut).

Recorded from kan-679; the practice was named in kan-517's deferred self-review
(`kan-517-body-group-workouts-release-fixes-1`).
