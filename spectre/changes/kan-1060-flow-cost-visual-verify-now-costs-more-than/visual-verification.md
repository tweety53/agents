# Visual verification — kan-1060

Frames: the approved prototype `.superpowers/sdd/2026-10-30-stage-leaderboard-main-share-prototype.html`, rendered at 1280x1024 dark, at rest and after one Main share click (`visual-verification/prototype/`, map `prototype.mockups`, `frame none`). No region capture, so no injected CSS.

## stage-leaderboard

- screenshot: `visual-verification/evidence/stage-leaderboard-darwin.png`
- composite: `visual-verification/stage-leaderboard.png`, frame `stage-leaderboard` (prototype), diff=n/a (frame none)
- seen: Stage leaderboard at rest: nine columns, Main session and Main share after Total cost; flow.review-panel reads $0.16 / 39%, every other row 100%; no NaN/null/raw float. Frame data is the dev store's, capture the fixture's: every text difference is data.
- bands: 0 flagged (all paired). seams: 0 flagged.
- over rows: #3 mockup source (unresolved; label row widths are data); #4 text width (Go to change select sized by its selected change name, 894 vs 705); #11 mockup source; #12 mockup source (unresolved, colour sample is text density/Command select width — data).

| element | kind | content | colour | size | position | stretch | alignment | border | fill | radius | shadow |
|---|---|---|---|---|---|---|---|---|---|---|---|
| #1 title "flow stats" | text run — label | transcribed per sweep 5 (frame vs capture data) | #d8d9da / #d8d9da / 0.0 | w 92 / 92 / +0; h 16 / 16 / +0 | left 24 / 24 / +0; top 20 / 20 / +0 | insets 24,1164 / 24,1164 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #2 nav tab row | text run — label | transcribed per sweep 5 (frame vs capture data) | #9fa7b3 / #9fa7b3 / 0.0 | w 1080 / 1080 / +0; h 14 / 14 / +0 | left 25 / 25 / +0; top 55 / 55 / +0 | insets 25,175 / 25,175 | insets equal (25 / 25 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #3 filter labels (Project/Model/Go to change) | text run — label | transcribed per sweep 5 (frame vs capture data) | #9fa7b3 / #9fa7b3 / 0.0 | w 426 / 415 / -11; h 15 / 15 / +0 | left 25 / 25 / +0; top 95 / 95 / +0 | insets 25,829 / 25,840 | insets equal (25 / 25 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #4 filter controls row | text run — label | transcribed per sweep 5 (frame vs capture data) | #d8d9da / #d8d9da / 0.0 | w 894 / 705 / -189; h 29 / 29 / +0 | left 24 / 24 / +0; top 112 / 112 / +0 | insets 24,362 / 24,551 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #5 PERIOD label | text run — label | transcribed per sweep 5 (frame vs capture data) | #9fa7b3 / #9fa7b3 / 0.0 | w 43 / 43 / +0; h 10 / 10 / +0 | left 25 / 25 / +0; top 159 / 159 / +0 | insets 25,1212 / 25,1212 | insets equal (25 / 25 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #6 period inputs + preset buttons | text run — label | transcribed per sweep 5 (frame vs capture data) | #2c3235 / #2c3235 / 0.0 | w 834 / 835 / +1; h 47 / 47 / +0 | left 24 / 24 / +0; top 170 / 170 / +0 | insets 24,422 / 24,421 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #7 filter-bar bottom rule | rule — hairline | n/a — rule | #2c3235 / #2c3235 / 0.0 | thickness 1 / 1 / +0 | left 24 / 24 / +0; top 229 / 229 / +0 | insets 24,24 / 24,24 | n/a — rule | n/a — rule | #2c3235 / #2c3235 / 0.0 | n/a — rule | n/a — rule |
| #8 header divider | rule — hairline | n/a — rule | #2c3235 / #2c3235 / 0.0 | thickness 1 / 1 / +0 | left 8 / 8 / +0; top 262 / 262 / +0 | insets 8,8 / 8,8 | n/a — rule | n/a — rule | #2c3235 / #2c3235 / 0.0 | n/a — rule | n/a — rule |
| #9 heading "Stage leaderboard" | text run — label | transcribed per sweep 5 (frame vs capture data) | #d8d9da / #d8d9da / 0.0 | w 206 / 206 / +0; h 23 / 23 / +0 | left 24 / 24 / +0; top 312 / 312 / +0 | insets 24,1050 / 24,1050 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #10 description line | text run — label | transcribed per sweep 5 (frame vs capture data) | #9fa7b3 / #9fa7b3 / 0.0 | w 648 / 648 / +0; h 16 / 16 / +0 | left 24 / 24 / +0; top 357 / 357 / +0 | insets 24,608 / 24,608 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #11 stat cards (Stages/Total runs) | fill band — panel | n/a — fill band | #d8d9da / #d8d9da / 0.0 | w 1232 / 1232 / +0; h 89 / 89 / +0 | left 24 / 24 / +0; top 389 / 389 / +0 | insets 24,24 / 24,24 | n/a — fill band | n/a — fill band; mockup source: border #2c3235 both (same CSS) | #d8d9da / #d8d9da / 0.0; mockup source: #181b1f both | n/a — fill band; mockup source: var(--panel-radius) both | n/a — fill band |
| #12 table panel "Every costed stage" | fill band — panel | n/a — fill band | #222529 / #4d5053 / 73.9 | w 1232 / 1232 / +0; h 423 / 423 / +0 | left 24 / 24 / +0; top 494 / 494 / +0 | insets 24,24 / 24,24 | n/a — fill band | n/a — fill band; mockup source: border #2c3235 both (same CSS) | #222529 / #4d5053 / 73.9; mockup source: #181b1f both | n/a — fill band; mockup source: var(--panel-radius) both | n/a — fill band |

## stage-leaderboard-main-share-sorted

- screenshot: `visual-verification/evidence/stage-leaderboard-main-share-sorted-darwin.png`
- composite: `visual-verification/stage-leaderboard-main-share-sorted.png`, frame `stage-leaderboard-main-share-sorted` (prototype), diff=n/a (frame none)
- seen: After one Main share click: header reads "Main share ▲", flow.review-panel (39%) first, the 100% rows follow in prior order; pointer parked, no hover ink. Frame data is the dev store's, capture the fixture's: every text difference is data.
- bands: 0 flagged (all paired). seams: 0 flagged.
- over rows: #3 mockup source (unresolved; label row widths are data); #4 text width (Go to change select sized by its selected change name, 894 vs 705); #11 mockup source; #12 mockup source (unresolved, colour sample is text density/Command select width — data).

| element | kind | content | colour | size | position | stretch | alignment | border | fill | radius | shadow |
|---|---|---|---|---|---|---|---|---|---|---|---|
| #1 title "flow stats" | text run — label | transcribed per sweep 5 (frame vs capture data) | #d8d9da / #d8d9da / 0.0 | w 92 / 92 / +0; h 16 / 16 / +0 | left 24 / 24 / +0; top 20 / 20 / +0 | insets 24,1164 / 24,1164 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #2 nav tab row | text run — label | transcribed per sweep 5 (frame vs capture data) | #9fa7b3 / #9fa7b3 / 0.0 | w 1080 / 1080 / +0; h 14 / 14 / +0 | left 25 / 25 / +0; top 55 / 55 / +0 | insets 25,175 / 25,175 | insets equal (25 / 25 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #3 filter labels (Project/Model/Go to change) | text run — label | transcribed per sweep 5 (frame vs capture data) | #9fa7b3 / #9fa7b3 / 0.0 | w 426 / 415 / -11; h 15 / 15 / +0 | left 25 / 25 / +0; top 95 / 95 / +0 | insets 25,829 / 25,840 | insets equal (25 / 25 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #4 filter controls row | text run — label | transcribed per sweep 5 (frame vs capture data) | #d8d9da / #d8d9da / 0.0 | w 894 / 705 / -189; h 29 / 29 / +0 | left 24 / 24 / +0; top 112 / 112 / +0 | insets 24,362 / 24,551 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #5 PERIOD label | text run — label | transcribed per sweep 5 (frame vs capture data) | #9fa7b3 / #9fa7b3 / 0.0 | w 43 / 43 / +0; h 10 / 10 / +0 | left 25 / 25 / +0; top 159 / 159 / +0 | insets 25,1212 / 25,1212 | insets equal (25 / 25 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #6 period inputs + preset buttons | text run — label | transcribed per sweep 5 (frame vs capture data) | #2c3235 / #2c3235 / 0.0 | w 834 / 835 / +1; h 47 / 47 / +0 | left 24 / 24 / +0; top 170 / 170 / +0 | insets 24,422 / 24,421 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #7 filter-bar bottom rule | rule — hairline | n/a — rule | #2c3235 / #2c3235 / 0.0 | thickness 1 / 1 / +0 | left 24 / 24 / +0; top 229 / 229 / +0 | insets 24,24 / 24,24 | n/a — rule | n/a — rule | #2c3235 / #2c3235 / 0.0 | n/a — rule | n/a — rule |
| #8 header divider | rule — hairline | n/a — rule | #2c3235 / #2c3235 / 0.0 | thickness 1 / 1 / +0 | left 8 / 8 / +0; top 262 / 262 / +0 | insets 8,8 / 8,8 | n/a — rule | n/a — rule | #2c3235 / #2c3235 / 0.0 | n/a — rule | n/a — rule |
| #9 heading "Stage leaderboard" | text run — label | transcribed per sweep 5 (frame vs capture data) | #d8d9da / #d8d9da / 0.0 | w 206 / 206 / +0; h 23 / 23 / +0 | left 24 / 24 / +0; top 312 / 312 / +0 | insets 24,1050 / 24,1050 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #10 description line | text run — label | transcribed per sweep 5 (frame vs capture data) | #9fa7b3 / #9fa7b3 / 0.0 | w 648 / 648 / +0; h 16 / 16 / +0 | left 24 / 24 / +0; top 357 / 357 / +0 | insets 24,608 / 24,608 | insets equal (24 / 24 left) | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control | n/a — not a boxed control |
| #11 stat cards (Stages/Total runs) | fill band — panel | n/a — fill band | #d8d9da / #d8d9da / 0.0 | w 1232 / 1232 / +0; h 89 / 89 / +0 | left 24 / 24 / +0; top 389 / 389 / +0 | insets 24,24 / 24,24 | n/a — fill band | n/a — fill band; mockup source: border #2c3235 both (same CSS) | #d8d9da / #d8d9da / 0.0; mockup source: #181b1f both | n/a — fill band; mockup source: var(--panel-radius) both | n/a — fill band |
| #12 table panel "Every costed stage" | fill band — panel | n/a — fill band | #4e5054 / #34373b / 43.9 | w 1232 / 1232 / +0; h 423 / 423 / +0 | left 24 / 24 / +0; top 494 / 494 / +0 | insets 24,24 / 24,24 | n/a — fill band | n/a — fill band; mockup source: border #2c3235 both (same CSS) | #4e5054 / #34373b / 43.9; mockup source: #181b1f both | n/a — fill band; mockup source: var(--panel-radius) both | n/a — fill band |

## trend (no frame)

- screenshot: `visual-verification/evidence/trend-darwin.png` — unchanged view, Trend over time renders: 2 days, $9.36, line chart, two day rows. No frame composed.

## Motions

none named.
