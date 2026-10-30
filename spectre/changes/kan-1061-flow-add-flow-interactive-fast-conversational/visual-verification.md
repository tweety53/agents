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
