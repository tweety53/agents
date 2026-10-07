# Visual verification — kan-927-in-run-self-review-fixes

Stack: the UI-test stack on http://127.0.0.1:4174 (seeded by `make ui-test-up`), viewport 1280x1024, dark colour scheme, period pinned by query params (support.ts's PINNED_QUERY). No `mockups` root is declared: no frame was composed, and the band/seam pairing, ink inventory, bounded-row sweep and element × property matrix are n/a — no frame.

## Views

### self-review (empty state) — `stats/web/tests/visual/self-review.spec.ts-snapshots/self-review-empty-darwin.png`

Captured by the new spec `stats/web/tests/visual/self-review.spec.ts`, which asserts the "Self-review fixes" nav link carries `aria-current="page"`, the heading is visible, both empty messages show and the dark palette applied, before the capture. Seen, top to bottom: header and nav (nine links, "Self-review fixes" bold and underlined as the active one), project/model/go-to-change controls, the period picker, the heading "Self-review fixes", the description "Every self-review finding: its angle, whether it was fixed, filed or declined, and the commit that fixed it.", two stat panels "FIXED 0" and "FILED 0", the "EVERY FINDING" panel with Search, Change/Angle/Outcome filters, "No self-review findings in this period." and the "Reset period to default" hint, then the "PER CHANGE" panel with Search and "No changes with self-review findings in this period.". No NaN, null, undefined or placeholder text; nothing crosses its container. The page fits the viewport, so nothing scrolls.

Populated state: not captured. Recording findings into the 4174 store with the branch's `flow self-review finding` failed with HTTP 500 — the daemon log shows `self_review_findings_project_key_fkey` violated for project `agents-a740d89c`, which the UI-test seed does not create — and the CLI then exited 0 with "finding not recorded". The UI-test fixture (`stats/cmd/uitest-seed`) seeds no self-review findings, so no populated, sortable or linked-ref row, and no commit link, has been seen rendered.

### full app suite — `stats/web/tests/visual/full-app-suite.spec.ts`

Updated: a `self-review` entry was added, and all ten captures were rewritten (`--update-snapshots=all`) because the nav bar gained a tab on every screen; existing baselines had passed only inside the suite's 2% tolerance. The nav-link locator was made `exact: true`: "Runs" also matched the runs view's "kan-103-runs-view" link and failed strict mode once the list rendered. Every screen shows the new "Self-review fixes" tab as the ninth nav link, in the same style as its siblings, with no wrap.

- `full-app-suite.spec.ts-snapshots/full-self-review-darwin.png` — identical in content to the empty-state capture above.
- `full-state-board-darwin.png` — state board, three changes; new tab present.
- `full-stage-leaderboard-darwin.png` — two stages, three runs; new tab present.
- `full-trend-darwin.png` — two days, chart and table; new tab present.
- `full-cache-efficiency-darwin.png` — two stage rows, ratios "unavailable"; new tab present.
- `full-reviewers-darwin.png` — empty; new tab present.
- `full-decisions-darwin.png` — empty; new tab present.
- `full-runs-darwin.png` — KAN-103 runs group, three rows; new tab present.
- `full-flow-health-darwin.png` — guards empty, two re-entry rows; new tab present.
- `full-run-detail-darwin.png` — kan-201-refactor-thing detail; new tab present.

All paths above are relative to `stats/web/tests/visual/`, committed with this change. Zip rebuilt at `stats/web/tests/visual/full-app-suite.zip` (10 PNGs).

Observed, pre-existing and not this change's: the active nav link is bold, so the links after it shift horizontally by a few px from one view to the next.

## Sweeps (no frame)

- text: every run transcribed; all display text, counts render as integers ("0").
- order: heading, description, stat row, findings table, per-change table; containment clean. Frame order n/a — no frame.
- reach: the page fits the 1024px viewport; no scrolling region.
- derived: n/a — the view has no editable input whose derived value could be re-derived; the counts derive from rows and none could be seeded.
- rows, ink: n/a — no frame.

## Motion

None named.
