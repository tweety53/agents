# Visual verification — kan-1061

Composites, matrices (matrix/prototype-<frame>.json) and band/seam reads (matrix/bs-<frame>.json) are in visual-verification/. Captures are the committed baselines under stats/web/tests/visual/interactive.spec.ts-snapshots/.

## Report
- setup: exit 0
- stack: already running
- fingerprint: exit 0
- verify: exit 0
  74 passed (6.8s)
- capture: exit 0
  5 passed (1.7s) — interactive running card, waiting card, running card hovered, running card focused from the keyboard, interactive run view; recaptures byte-identical to the committed baselines
- full app suite: updated — stats/web/tests/visual/full-app-suite.spec.ts, 11 screens, capture exit 0 (state-board and run detail screens recaptured, both unchanged)
- full app suite zip: stats/web/tests/visual/full-app-suite.zip, 11 PNGs
- spec reach: exit 0 (Spec reach: not configured)
- identical: interactive-running-card-focus-darwin.png interactive-running-card-hover-darwin.png — the prototype styles both with one rule, `.card:hover, .card:focus-visible { border-color: var(--st-progress); outline: none; }`, so equal pixels are the spec; both specs assert border-top-color rgb(255, 152, 48)
- spec: stats/web/tests/visual/interactive.spec.ts
- spec departure (step 8): the four card captures are element screenshots (`card.toHaveScreenshot`), a clipped region, where step 8 requires every capture to be the full page or viewport; the run view capture is full-page. This change's (the spec is new)
- interactive-running-card: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-running-card-darwin.png — running card: title, Interactive pill, "Implementing tasks · 3m · Edit: stats/web/src/lib/phase.ts", "Request 7: Show the stage's model on the Agents row", "main session running", orange progress bar, uitest-alpha, "requests 6/8 · $4.12"; no defect
- interactive-waiting-card: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-waiting-card-darwin.png — hollow dot "Waiting for the next request · 6m", "Last: Request 3 landed — Hide the empty streak row", full orange bar, "requests 3/3 · $1.87"; no defect
- interactive-running-card-hover: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-running-card-hover-darwin.png — orange #ff9830 border; no defect
- interactive-running-card-focus: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-running-card-focus-darwin.png — orange #ff9830 border, no outline ring; no defect
- interactive-run-view: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-run-view-darwin.png — header with inline Interactive pill, KPI strip, Agents running (Main · running · opus 5.5 · xhigh · 3m), Main chips, Requests 6/8 list, Timeline, Stage runs with Request column and "Show all 31", Cost by stage; no raw value, NaN or placeholder
- mockups: not declared — approved prototype rendered as frames (frames already rendered into .superpowers/sdd/interactive-frames/ from .superpowers/sdd/planner/prototype-board.html, mapped by interactive.spec.ts.mockups), compose exit 0
- prototype-running-card: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-running-card.png diff=n/a — match; width 380 vs 391 inner (card 396 vs 407): the app's 3-column board grid at 1280px vs the prototype's `repeat(auto-fill, minmax(320px, 1fr))` — known leftover, layout cause; project label uitest-alpha vs agents — fixture
- prototype-running-card bands: 1/1 paired, 0 missing, 0 extra — top/gap_above delta 1px, within tolerance
- prototype-running-card seams: 4/4 paired, 0 missing, 0 extra, 8 lines — every line within 1px; right seam at 395 vs 406 is the card width above
- prototype-running-card sweeps: text done | order done | reach done | derived n/a — no operator-editable input on the card | rows done — card top/bottom gaps 1/1 vs 1/1, inner left/right 12/12 both | ink done — border #2c3235, progress track and fill paired
- prototype-running-card matrix: 9 elements × 11 columns, 0 n/a
- prototype-waiting-card: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-waiting-card.png diff=n/a — height 124 vs 148, line tops 11/40/63/97 vs 14/48/74/118: the prototype's `.cards` grid stretches the card to its row-mate (the running card, `.card { display: grid; gap: 6px }`); in the app its row-mates are 124px cards (see full-state-board) — named cause: row neighbours, fixture
- prototype-waiting-card bands: 1/1 paired, 0 missing, 0 extra — height delta 26px, cause the grid row stretch above
- prototype-waiting-card seams: 4/4 paired, 0 missing, 0 extra, 6 lines — line offsets within 1px; vertical tops differ by the row stretch above
- prototype-waiting-card sweeps: text done | order done | reach done | derived n/a — no editable input | rows done — insets 12/12 both | ink done — border, full bar paired
- prototype-waiting-card matrix: 7 elements × 11 columns, 0 n/a
- prototype-running-card-hover: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-running-card-hover.png diff=n/a — match, border #ff9830 both
- prototype-running-card-hover bands: 1/1 paired, 0 missing, 0 extra
- prototype-running-card-hover seams: 4/4 paired, 0 missing, 0 extra, 8 lines — within 1px
- prototype-running-card-hover sweeps: text done | order done | reach done | derived n/a — no editable input | rows done | ink done
- prototype-running-card-hover matrix: 9 elements × 11 columns, 0 n/a
- prototype-running-card-focus: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-running-card-focus.png diff=n/a — match, border #ff9830 both, no outline
- prototype-running-card-focus bands: 1/1 paired, 0 missing, 0 extra
- prototype-running-card-focus seams: 4/4 paired, 0 missing, 0 extra, 8 lines — within 1px
- prototype-running-card-focus sweeps: text done | order done | reach done | derived n/a — no editable input | rows done | ink done
- prototype-running-card-focus matrix: 9 elements × 11 columns, 0 n/a
- prototype-run-view: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-run-view.png diff=n/a — layout differs wholesale (elements 1 paired, 4 missing, 9 extra). Attributed:
  - app nav, "Project uitest-alpha · KAN-1072 · updated just now" subtitle, status pill inline after the title (prototype right-aligned), Timeline panel, outcome as coloured "● completed" text (prototype grey text), tokens `1200k` vs `1.20M`, bottom row Stage runs + Cost by stage with bar between label and value (prototype `.hbar { grid-template-columns: 170px minmax(0, 1fr) 60px }`): present at the merge base, merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png
  - Requests panel column ~265px beside Agents/Main, so requests 2 and 6 wrap to two lines; the prototype declares `.work { grid-template-columns: minmax(0, 3fr) minmax(0, 2fr) }` (~570/380) with Requests and Cost by stage in the right column: absent at the merge base — this change's defect; merge-base capture failed — the merge base seeds no interactive session and has no Requests panel
  - Agents running Status shows "● running" where the prototype draws the dot alone: this change's defect; merge-base capture failed — same reason
  - Stage runs 5 rows + "Show all 31" chip vs the prototype's 6 rows, no chip: this change's; merge-base capture failed — same reason
  - Started 08:57 vs 11:02: fixture clock vs prototype literal — named cause; Main chips on one line vs two: consequence of the panel width above
- prototype-run-view bands: 1/1 paired, 0 missing, 0 extra — height 797 vs 671 content band, cause the Timeline panel and grid above
- prototype-run-view seams: paired 1 band; panel left edge 24 vs 0 extra/missing — the app's 24px page inset vs the prototype's frame origin, pre-existing (merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png)
- prototype-run-view sweeps: text done | order done — prototype order header, KPIs, [Agents, Main, Stage runs | Requests, Cost by stage]; capture header, KPIs, [Agents, Main | Requests], Timeline, [Stage runs | Cost by stage] | reach done — page ends at Cost by stage/Stage runs within the viewport | derived n/a — no editable input | rows done — panel headers 12px insets both | ink done — panel borders #2c3235 paired, Timeline panel border extra (pre-existing, merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png)
- prototype-run-view matrix: 14 elements × 11 columns, 0 n/a
- frames: 5/5
- motion: n/a — no motions named
- wire contract: n/a — no wire contract named
- visual-verification.md: written
- steps: 4 done | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — no motions named

## In-run fix 2 (Task 15)

Task 15 moved the interactive run view's Requests and Cost by stage into a right column of a `minmax(0, 3fr) minmax(0, 2fr)` grid, drew the Agents status as the dot alone, showed 6 stage-run rows with no "Show all" chip, and switched the four card captures to viewport shots. The re-verify below supersedes the departures listed above.

## In-run fix 2 — re-verify (round base 9825b70e)

Captures cited at the committed baselines stats/web/tests/visual/interactive.spec.ts-snapshots/; composites at visual-verification/interactive-<frame>.png; matrices at visual-verification/matrix/prototype-<frame>.json and band/seam reads at visual-verification/matrix/bs-<frame>.json, overwritten this round.

- setup: not declared
- stack: already running
- fingerprint: exit 0
- verify: exit 0
  74 passed (6.2s) — whole suite (no specs row declared)
- capture: exit 0
  5 passed (1.9s) — interactive.spec.ts recaptured with --update-snapshots=all; byte-identical to the committed baselines (git status clean under the snapshots)
- full app suite: updated — stats/web/tests/visual/full-app-suite.spec.ts, 11 screens, capture exit 0 (run detail screen recaptured, unchanged)
- full app suite zip: stats/web/tests/visual/full-app-suite.zip, 11 PNGs (rebuild exit 0, unchanged)
- spec reach: exit 0 (not configured)
- identical: interactive-running-card-darwin.png interactive-waiting-card-darwin.png — both are now viewport shots of the same board in the same resting state (both cards sit in the one viewport), so equal pixels are expected; consequence: the waiting-card test no longer captures anything the running-card test does not — fixed by Task 16 (below)
- identical: interactive-running-card-focus-darwin.png interactive-running-card-hover-darwin.png — prototype styles both with one rule, `.card:hover, .card:focus-visible { border-color: var(--st-progress); outline: none; }`; the specs assert border-top-color rgb(255, 152, 48)
- spec: stats/web/tests/visual/interactive.spec.ts
- fix confirmed — card captures are viewport shots (1280×1024), no clip; prototype frames in .superpowers/sdd/interactive-frames/ are 1280×1024
- interactive-running-card: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-running-card-darwin.png — board viewport; kan-1072 card third in row 1: "Implementing tasks · 3m · Edit: stats/web/src/lib/phase.ts", "Request 7: Show the stage's model on the Agents row", "main session running", orange bar, uitest-alpha, "requests 6/8 · $4.12"; no defect
- interactive-waiting-card: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-waiting-card-darwin.png — same pixels as above; kan-1073 card: hollow dot "Waiting for the next request · 6m", "Last: Request 3 landed — Hide the empty streak row", full orange bar, "requests 3/3 · $1.87"; no defect
- interactive-running-card-hover: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-running-card-hover-darwin.png — kan-1072 card border #ff9830; no defect
- interactive-running-card-focus: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-running-card-focus-darwin.png — same as hover, no outline ring; no defect
- interactive-run-view: stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-run-view-darwin.png — fixes confirmed: work area is two columns, left 25–756 / right 767–1255 (731/488 ≈ 3fr/2fr, prototype `.work { grid-template-columns: minmax(0, 3fr) minmax(0, 2fr) }` 21–757 / 768–1259); right column holds Requests then Cost by stage; Agents Status is the green dot alone; Stage runs has 6 rows (R7 ×2, R6 ×4) and no "Show all" chip. Requests rows no longer wrap. No raw value, NaN or placeholder
- mockups: not declared — approved prototype rendered as frames (.superpowers/sdd/interactive-frames/, from .superpowers/sdd/planner/prototype-board.html, mapped by interactive.spec.ts.mockups), compose exit 0
- prototype-running-card: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-running-card.png diff=n/a — card: box 395×147 vs 407×146, radius 2/2, fill #181b1f both; width −12: the app's 3-column board grid vs the prototype's `.cards { repeat(auto-fill, minmax(320px, 1fr)) }` — known leftover, layout cause. Page-level: the prototype frame draws a bare "Live state board / Active runs" header and the cards at y=75; the app draws nav, title, filter bar and KPI tiles above the grid (present at the merge base, merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-state-board-darwin.png); card order kan-102, kan-101, kan-1072, kan-1073 vs prototype kan-1072, kan-1073, kan-1068 — fixture (the seeded board's other changes); project label uitest-alpha vs agents — fixture
- prototype-running-card bands: 0/3 paired, 3 missing, 1 extra — the app's full-height page band vs the prototype's header/subtitle/card-row bands: the page chrome above, present at the merge base (merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-state-board-darwin.png); card-region read paired box and fill
- prototype-running-card seams: n/a pairing at page scope (no paired boxed band); card-region box/radius/fill paired, right edge −12 is the width cause above
- prototype-running-card sweeps: text done | order done — card internals same order | reach done | derived n/a — no operator-editable input | rows done — card inner insets 12/12 both | ink done — border #2c3235, progress track and fill paired
- prototype-running-card matrix: 14 elements × 11 columns, 0 n/a
- prototype-waiting-card: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-waiting-card.png diff=n/a — card 395×122 vs 406×146: height −24 is the prototype grid row stretching the card to its running row-mate; width the board-grid cause above; page chrome as above
- prototype-waiting-card bands: 0/3 paired, 3 missing, 1 extra — page chrome cause as prototype-running-card
- prototype-waiting-card seams: n/a pairing at page scope; card-region box/radius/fill paired
- prototype-waiting-card sweeps: text done | order done | reach done | derived n/a — no editable input | rows done — insets 12/12 both | ink done — border, full bar paired
- prototype-waiting-card matrix: 14 elements × 11 columns, 0 n/a
- prototype-running-card-hover: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-running-card-hover.png diff=n/a — border #ff9830 both; card box as running card
- prototype-running-card-hover bands: 0/3 paired, 3 missing, 1 extra — page chrome cause as above
- prototype-running-card-hover seams: n/a pairing at page scope; card-region paired
- prototype-running-card-hover sweeps: text done | order done | reach done | derived n/a — no editable input | rows done | ink done
- prototype-running-card-hover matrix: 14 elements × 11 columns, 0 n/a
- prototype-running-card-focus: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-running-card-focus.png diff=n/a — border #ff9830 both, no outline
- prototype-running-card-focus bands: 0/3 paired, 3 missing, 1 extra — page chrome cause as above
- prototype-running-card-focus seams: n/a pairing at page scope; card-region paired
- prototype-running-card-focus sweeps: text done | order done | reach done | derived n/a — no editable input | rows done | ink done
- prototype-running-card-focus matrix: 14 elements × 11 columns, 0 n/a
- prototype-run-view: spectre/changes/kan-1061-flow-add-flow-interactive-fast-conversational/visual-verification/interactive-run-view.png diff=n/a — the three fixes match the prototype. Remaining:
  - NEW — Stage runs Stage cells truncate with an ellipsis: "Implementing ..." (R7) and "Updating the main ..." (R6); the prototype declares `td { white-space: nowrap }` with no clip on the Stage cell and draws "Updating the main checkout" in full. Stage column ~147px (x 119–266) vs ~187px (103–290): the app keeps its session-row indent on Stage cells and a wider "● completed" Outcome, now inside the 3fr column. Absent at the merge base (merge base Stage column ~315px, no truncation, merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png) — this change's defect; fixed by Task 16 (below)
  - Cost by stage labels truncate ("Running lint and t…", "Landing the chan…", "Designing the ch…", "Updating the mai…"), label column ~115px vs the prototype's `.hbar { grid-template-columns: 170px minmax(0, 1fr) 60px }`, and the bar has no track (prototype `.hbar .b { background: var(--border) }`): present at the merge base ("SDD + TDD per t…", no track), merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png
  - Timeline panel, app nav, subtitle, inline status pill, coloured outcome text, tokens 1200k vs 1.20M, KPI figure size: present at the merge base, merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png
  - Stage runs header shows "click a row for its dispatches" where the prototype shows count "31": present at the merge base, merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png
  - Cost by stage 8 rows vs 5, values (Loading context $0.16 vs $0.41), Started 08:57 vs 11:02: fixture — named cause
- prototype-run-view bands: 1/1 paired, 0 missing, 0 extra — content band height 757 vs 671 (+86): app nav and subtitle plus the Timeline panel, present at the merge base, merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png
- prototype-run-view seams: 1 band paired; panel left edge 25 vs 21 — the app's 24px page inset, present at the merge base, merge-base baseline: 35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png
- prototype-run-view sweeps: text done — Stage "Implementing ..."/"Updating the main ..." ellipses the departure above | order done — prototype header, KPIs, [Agents, Main, Stage runs | Requests, Cost by stage]; capture header, KPIs, [Agents, Main, Timeline, Stage runs | Requests, Cost by stage]; Timeline extra (merge base) | reach done — page ends at Stage runs/Cost by stage inside the viewport | derived n/a — no editable input | rows done — panel header insets 12px both images | ink done — panel borders #2c3235 paired; Timeline border extra (merge base); Cost by stage bar track absent (merge base)
- prototype-run-view matrix: 12 elements × 11 columns, 0 n/a
- frames: 5/5
- motion: n/a — no motions named
- wire contract: n/a — no wire contract named
- visual-verification.md: written
- steps: 4 n/a — setup not declared | 7 done | 8 done | 9 done | 10 done | 11 done | motion n/a — no motions named

## In-run fix 3 (Task 16)

Task 16 fixed the two departures the re-verify above raised. The interactive run view's Stage runs table now sizes its columns to their content (`table-layout: auto`, as the prototype's table does) and its Stage cells carry no session-row indent, so every Stage cell draws its text in full: in the recaptured stats/web/tests/visual/interactive.spec.ts-snapshots/interactive-run-view-darwin.png the Stage column is 190px (prototype ~187px), "Updating the main checkout" and "Implementing tasks" show whole, and no cell's scrollWidth exceeds its clientWidth, which interactive.spec.ts's run-view test now asserts. The waiting-card test hovers the kan-1073 card before capturing (border #ff9830, asserted), so interactive-waiting-card-darwin.png differs from interactive-running-card-darwin.png; interactive.spec.ts.mockups names that state, so the prototype-waiting-card frame is rendered with the waiting card hovered on the next compose. The `/flow` and `/flow-fast` run views and every other baseline are unchanged; full-app-suite.zip rebuilt unchanged. The Cost by stage truncation and the other merge-base leftovers listed above stand.
