# Visual verification — kan-1047 (review-1)

## review-dynamics (full page, dark)
- Screenshot: `visual-verification/evidence/review-dynamics-darwin.png`
- Frame: `visual-verification/prototype/review-dynamics.png`, rendered from `.superpowers/sdd/heatmap-by-project-prototype.html` at 1280x1024, dsf 1.
- Seen: bars in order Range, Models, Projects (All · uitest-dynamics), Task size (All · Big · Regular · Small · Micro), Reviewers (Current · All), Prompt changes (flow-prompt-tuning · All · auto prompt tuning, default flow-prompt-tuning); no "Since last prompt change"; six tiles; every chart row holds two charts; heatmap rows primary, principles, test-sanity. Matches the prototype's structure and the three superseding decisions.
- Not composed: the prototype page is wrapped in an annotation banner, is light-themed on a white card and plots live store data (gymie + Big), so no pixel pairing against the dark seeded capture is meaningful; bands/seams/matrix not measured.
- Departures found in review-1, fixed by Task 4 (In-run fix 1):
  1. "default preset" asserted `Last 8 weeks`; it now asserts `Last 4 weeks`.
  2. "chart hover" expected an edit marker the default `flow-prompt-tuning` filter hides; it now picks Prompt changes `All` first, so the capture draws the purple marker.
  3. The new bar states were undriven; `review dynamics filters` now captures Reviewers `All` + Prompt changes `auto prompt tuning` (`review-dynamics-filters-darwin.png`) and asserts a Task size and a Projects pick press their buttons.
- Note:
  4. `review-dynamics-heat-hover` is byte-identical to `review-dynamics` (native title tooltip, no hover ink) — by design per the spec's comment.
- Intro copy still says "hover a purple marker", while the default filter draws none over the seed — as the prototype's copy, not a departure.
