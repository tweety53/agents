- [x] 11. In-run fix 1 — flow.visual-verify: intro type, heatmap packing and order, effort order, tooltip surface, missing state captures
  - [x] **Step 1: Write the failing probes**:
    - `stats/web/src/views/views.test.tsx` gets `review dynamics orders the heatmap busiest slot first and effort low to high`. It checks that heatmap rows run by dispatch count descending, that columns run by model and then low → medium → high, and that the reviewer-effort legend reads low, medium, high.
    - `stats/web/tests/visual/review-dynamics.spec.ts` asserts the measures the verifier took against the prototype: the intro is 12px, a heatmap cell is 92px wide with no bottom rule, and the tooltip background is `rgb(34, 37, 43)`. Run them; they go RED.
  - [x] **Step 2: Implement**:
    - The intro uses `.dynamics-sub`.
    - `.dynamics-heat` is `width: auto`. Its cells drop the global table rule and padding.
    - `byEffort` orders heatmap columns and series names.
    - Heatmap slots are ordered busiest first.
    - The tooltip uses a new `--surface-tip` token: `#22252b` dark, `#ffffff` light.
    - The hover capture is full page.
    - New captures: `review dynamics by version` (Version pressed), `review dynamics default preset` (clock pinned to 2026-06-29, Last 8 weeks pressed), and `review dynamics heatmap cell hover`. The cell hover is the native title, so its text is asserted and the hovered page captured.
  - [x] **Step 3: Verify:** `cd stats/web && npx tsc -b && npx vitest run src/views/views.test.tsx -t "review dynamics"`, then `npx playwright test tests/visual/review-dynamics.spec.ts`. Recapture `full-review-dynamics` and rebuild `full-app-suite.zip`.

**Files:** `spectre/changes/kan-1046-review-dynamics-dashboard/visual-verification.md`, `stats/web/src/views/ReviewDynamics.tsx`, `stats/web/src/styles.css`, `stats/web/src/theme.css`, `stats/web/src/views/views.test.tsx`, `stats/web/tests/visual/review-dynamics.spec.ts`, `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-review-dynamics-darwin.png`, `stats/web/tests/visual/full-app-suite.zip`
**Allowed-collateral:** `stats/web/tests/visual/review-dynamics.spec.ts-snapshots/*.png`
**Tests:** `review dynamics orders the heatmap busiest slot first and effort low to high`, `review dynamics by version`, `review dynamics default preset`, `review dynamics heatmap cell hover`
**Regression:** `review dynamics orders the heatmap busiest slot first and effort low to high` fails under alphabetical sorting. The spec's CSS assertions fail on the old `.view-description` intro, the stretched heatmap and the `--surface-2` tooltip.
**Baseline:** before=3 after=6 <!-- predicted: test( count in stats/web/tests/visual/review-dynamics.spec.ts -->
**Build:** green
**After:** Task 6, 7, 10
**Commit:** `fix(stats): review dynamics matches its prototype`
