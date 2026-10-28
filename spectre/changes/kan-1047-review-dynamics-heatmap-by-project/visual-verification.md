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
