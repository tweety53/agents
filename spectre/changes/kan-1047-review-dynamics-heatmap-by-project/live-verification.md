# Live verification — kan-1047

Range: the last 28 days to 2026-10-28, review-dynamics view.

- Before (dev flowd on main 435bae51 — merge base edf8e4e8 plus a cmd/flow-only pipeline fix; no stats/internal change): `class=big` → 400 (unknown parameter); the answer carries no `projects` field; 140 changes reviewed.
- After (this branch's flowd on a store copy, `scripts/live-store-copy.sh up k1047 4175`): `projects` = agents, flow-eval, gymie, spectre (gymie-frontend had no reviewer dispatch in these 28 days, so it is absent — the design's list was read over a wider range); 140 changes overall; gymie 67; gymie + Big 30; heat slots for gymie + Big are the six current slots only; `class=huge` → 400.
- Failure look (header field not following a click, counts not changing with Task size, retired rows under Current): not seen — Task size moves gymie 67 → 30.
