# Visual verification — kan-1047 (closing unscoped run after fix 1)

verify: whole suite exit 0 (67 passed). capture of review-dynamics.spec.ts exit 0, baselines byte-identical to Task 4's commit.

## Views
- review-dynamics — `visual-verification/evidence/review-dynamics-darwin.png`: Range (Last 4 weeks preset row, no "Since last prompt change"), Models, Projects (All · uitest-dynamics), Task size (All · Big · Regular · Small · Micro), Reviewers (Current · All), Prompt changes (flow-prompt-tuning · All · auto prompt tuning); six tiles; two-column chart grid; heatmap. Text runs are display values at frame precision; no NaN/undefined.
- review-dynamics-filters — `visual-verification/evidence/review-dynamics-filters-darwin.png`: Reviewers All and Prompt changes "auto prompt tuning" pressed; prompt markers drawn at 05-18 and 06-01; heatmap gains the test-sanity row (0.00 under sonnet low) that Current hides.
- review-dynamics-hover — `visual-verification/evidence/review-dynamics-hover-darwin.png`: Prompt changes All; tooltip "Week of 05-18 / Critical 1.0 / Important 0.5 / Minor 4.0 / based on 2 changes / ◆ Prompt edit · primary: … (2e053aa)".
- review-dynamics-light — `visual-verification/evidence/review-dynamics-light-darwin.png`: light palette, same structure.
- review-dynamics-heat-hover is byte-identical to review-dynamics: the heat cell hover is a native title tooltip with no hover ink.

## Prototype frame `review-dynamics`
Composite `visual-verification/review-dynamics-light.png`, diff=n/a (sizes differ: frame 1280x2802, capture 1280x2647). Band pairing: 0 paired — the frame's whole height reads as one dark wrapper band (#15161b) because the prototype embeds the page in an annotation header and an inset light panel (left 8/right 8); every capture band is `extra` against it. Cause: frame annotation outside the app's own ink, not a layout departure. Seams likewise unpaired for the same cause. Matrix saved to `visual-verification/matrix/review-dynamics.json`; its rows are the capture's elements, all `extra`, for the same cause. Structural reading by eye against the prototype: bar order, tiles, chart grid and heatmap section agree; prototype's "Since last prompt change" preset and absence of the Prompt changes bar are superseded by design.md `drop-tune-preset` and `marker-kind-filter`; the two-column grid by `two-column-grid`.

## Departures
- missing capture (this change's — the Projects and Task size bars are new, absent at the merge base): the prototype draws a project pick (gymie) and Task size Big selected; no capture test drives a Projects pick or a Task size pick, so those selected states and the filtered data are unverified.

## Motion
none named.

## In-run fix 2 (Task 5)
- The fix-1-full run's departure — no capture picks a project or a task size — is fixed: `seedReviewDynamics` now records a plan class per `uitest-dynamics` change (big, small), and `review dynamics filters` captures `review-dynamics-picked-darwin.png` with project `uitest-dynamics` and Task size `Big` pressed; the header Project field follows the pick (reads `uitest-dynamics`), and the counts narrow (Changes reviewed 16 → 8). No other baseline moved with the seed (`npm run test:visual`: 67 passed).

## Round: visual-verify-fix-2-full (HEAD db381140)

### review-dynamics-picked
- Screenshot: `visual-verification/evidence/review-dynamics-picked-darwin.png` — dark scheme; Projects `uitest-dynamics` and Task size `Big` pressed, Reviewers All, Prompt changes auto prompt tuning; KPIs 5.50 / 2.50 / 2.00 / 2.20 / 38% / 8 (▲ 300% vs before); all eight charts and the heatmap (primary, principles, test-sanity rows) drawn; no NaN/null/placeholder.
- Composite: `visual-verification/review-dynamics-picked.png`, frame `review-dynamics` (approved prototype), diff=n/a — sizes differ (frame 1280x2802 with annotation header and inset light panel; capture 1280x2647 dark).
- Bands: 1/1 paired as one whole-page band (frame wrapper `#15161b` vs page `#111217`, distance 6.9; height −155px — the frame's annotation header). Seams: outer edges only. Cause: prototype's annotation wrapper, outside the app's ink.
- Matrix: `visual-verification/matrix/review-dynamics-picked.json` — frame elements unpaired (wrapper), every column n/a — frame annotation wrapper.
- Departures (this change's; merge-base capture failed — Projects/Task size bars absent at the merge base):
  - "Changes reviewed" y-axis prints ticks `1, 1, 0`: with every week at 1 change the integer formatter rounds the 0.5 tick to `1` — a duplicated, wrong axis label.
  - The `review dynamics filters` test clicks Projects/Task size buttons and captures without `page.mouse.move(0, 0)`: the pointer is left on `Big`, so hover ink can reach the baseline.
- Prototype differences superseded by design.md: no "Since last prompt change" preset (`drop-tune-preset`), extra Prompt changes bar (`marker-kind-filter`), two-column grid (`two-column-grid`). Data values differ by seed.

## In-run fix 3 (Task 6)
- The fix-2-full departures are fixed: `DynamicsChart` blanks a y label that repeats the tick above it, so the picked state's "Changes reviewed" axis reads `0`, (blank), `1` with `1` on the line at value 1 (vitest `dynamics chart drops a tick whose label repeats the one above`); `review dynamics filters` moves the pointer to (0, 0) before each screenshot. Only `review-dynamics-picked-darwin.png` moved; `npm run test:visual`: 67 passed.

## Round: visual-verify-fix-3-full (HEAD fce43ca0)

### review-dynamics-picked
- Screenshot: `visual-verification/evidence/review-dynamics-picked-darwin.png` — dark scheme; Projects `uitest-dynamics` and Task size `Big` pressed; KPIs 5.50 / 2.50 / 2.00 / 2.20 / 38% / 8 (▲ 300% vs before); all eight charts and the heatmap drawn; no NaN/null/placeholder. "Changes reviewed" y-axis now reads `0`, (blank), `1` — the duplicated `1` is gone. No hover ink on any control (pointer parked at (0, 0), spec lines 72 and 79).
- Composite: `visual-verification/review-dynamics-picked.png`, frame `review-dynamics` (approved prototype), diff=n/a — sizes differ (frame 1280x2802 with annotation wrapper; capture 1280x2647).
- Bands: 1/1 paired, whole page (wrapper `#15161b` vs page `#111217`, distance 6.9; height −155px). Seams: outer edges only. Cause: the prototype's annotation wrapper, outside the app's ink.
- Matrix: `visual-verification/matrix/review-dynamics-picked.json` — frame elements unpaired (wrapper), every column n/a — frame annotation wrapper.
- Departures: none. Prior rounds' two departures verified fixed.

### review-dynamics-light
- Screenshot: `visual-verification/evidence/review-dynamics-light-darwin.png`; composite `visual-verification/review-dynamics-light.png` diff=n/a; same wrapper-only band pairing. Baseline unchanged this round.

### Other views
- `review-dynamics`, `-filters`, `-hover` captures (evidence/) byte-identical to HEAD's baselines; `review-dynamics-darwin.png` and `review-dynamics-heat-hover-darwin.png` are byte-identical — the heat-cell hover is a native `title` tooltip with no page ink.

## Fix run 1 (Task 7)
- Every Review dynamics y axis now draws ticks `0, s, 2s` on a 1·2·5 step, whole when the format cannot show a half, each label at the height of the value it names (vitest `dynamics chart labels each tick with the value it sits at`). Picked state: "Changes reviewed" reads `0`, `1`, `2` with the line on `1`; "Visual verify passed first time" tops at `100%`; findings charts read e.g. `0.0`, `5.0`, `10.0`. Every `review-dynamics-*` baseline, `full-review-dynamics-darwin.png` and the full-app-suite zip were recaptured and read; `npm run test:visual`: 67 passed.
