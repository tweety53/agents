# Visual verification — kan-1041 (In-run fix 4, closing full run)

HEAD `3e7e8429c0d22c277dde87b7463f21a470ec0720`, merge base `a9f37e898a4e54b279716f66b0b54a0793c3264b`. Frame source: the approved prototype `.superpowers/sdd/weekly-model-effort-preview.html` (`design.md` `## Context`), rendered per capture state by `visual-verification/prototype/render.cjs` at 1280x1024, mapped by `visual-verification/prototype.mockups` (`frame none`, so no difference panel and no `diff=` ratio).

## Views (step 9)

- `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-darwin.png` — dark; both panels; phase chart all seven series on, pills `$cost ▲n%` / `$cost` stacked per week with leader lines; model·effort chart four series (`opus-5 · session`, `sonnet-5 · session`, `opus-5 · low`, `opus-5 · medium`). Every figure a 2-dp dollar value, every delta a whole percent; deltas re-derived by hand: 2.67/0.90 → ▲197%, 1.10/0.90 → ▲22%, 0.41/0.60 → ▼32%, 2.17/0.35 → ▲520%, 0.30/0.25 → ▲20%, 0.20/0.30 → ▼33%. No NaN, null or raw float.
- `…/weekly-costs-light-darwin.png` — same state, light palette; pills and delta tokens use the light tokens.
- `…/weekly-costs-toggled-darwin.png` — Overall pressed off (dimmed chip), its line, points and pills gone; other series and their pills unchanged.
- `…/weekly-costs-model-effort-darwin.png` — default state: phase chart Overall only, model·effort chart all four on.
- `…/weekly-costs-tooltip-darwin.png` — tooltip on Overall 2026-08-12: `Overall / Week of 2026-08-12 / $3.37 / ▲ 124.5% vs prev week`; pill beside it reads `$3.37 ▲124%`.
- `…/weekly-costs-empty-darwin.png` — both panels show `No data was recorded for this period.`; the prototype draws no empty state, so no frame.
- Identical baselines: none (6 PNGs).

## Composed pairs (step 10)

Composites: `visual-verification/weekly-costs.png`, `weekly-costs-light.png`, `weekly-costs-toggled.png`, `weekly-costs-model-effort.png`, `weekly-costs-tooltip.png` (capture left, prototype right, `diff=n/a`).

### Bands and seams (every frame)

The prototype is a centred 1000px `main` on a flat page (`main{max-width:1000px;margin:0 auto}`); the app is the full-width app shell (nav header, Panel cards). The band read pairs 0–3 of 26–30 frame bands per frame and reads the capture as one page-wide band (`#111217`, 1666px): the layout differs wholesale, as the composite shows — present at the merge base; cause: decision recorded by the parent, full-width layout keeping the shared app shell (merge-base baseline: a9f37e898a4e54b279716f66b0b54a0793c3264b:stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-darwin.png, same shell, Panel caption and full-width card). Seams: 1 missing, 1 extra per frame (3 extra on light) — the frame's dashed note box vs the capture's Panel card edge, same cause.

### Departures

1. **Header copy** — resolved: capture now reads `Wednesday-to-Wednesday cost, last 8 weeks.`, the frame's `<p class="sub">` verbatim (`WeeklyCosts.tsx:320`).
2. **Section titles** (cause: decision recorded by the parent, keep the shared design-system convention — `Panel` caption and subtitle) — frame `h2{font-size:16px}` `Cost by phase` / `Cost by model · effort`; capture uppercase Panel captions plus subtitles the frame does not draw (`The phases and Other add up to Overall.`, `Each subagent dispatch by the model and effort that served it, plus the parent session per model.`). Phase panel caption + subtitle style present at the merge base; cause: decision recorded by the parent, full-width layout keeping the shared app shell (merge-base baseline: a9f37e898a4e54b279716f66b0b54a0793c3264b:stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-darwin.png, `LAST 8 WEEKS` + subtitle); the model·effort panel is a new view: this change's.
3. **Series palette** (cause: decision recorded by the parent, keep the shared design-system convention — `--series-*` tokens) — frame `--c0:#e5e7eb` (Overall / `opus · session` near-white), `--c1:#60a5fa`, `--c2:#fbbf24`, `--c3:#a78bfa`, `--c4:#2dd4bf`, `--c5:#f472b6`, `--c6:#a3e635`; capture uses the app's `--series-*` tokens (Overall blue, Planning green, Implementation yellow, Review orange, Self-review purple). Phase chart palette order present at the merge base; cause: decision recorded by the parent, full-width layout keeping the shared app shell (merge-base baseline: a9f37e898a4e54b279716f66b0b54a0793c3264b:stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-darwin.png); model·effort chart's palette (first series blue, not `#e5e7eb`) is a new view: this change's.
4. **Pill width** — resolved: `WeeklyCosts.tsx:307` `Math.max(hasDelta ? 84 : 40, Math.ceil(6 + text.length * 5.5))` matches the frame's `wd=o.p==null?40:84`; every fixture pill fits the fixed width (longest `$3.37 ▲124%`, 67 < 84), so none widens. Cost-only pill box measured 43px on the frame (40 + stroke/antialias); capture cost-only pills (`$0.90`, `$2.75`, `$0.62`) visibly the same short width, delta pills the same long width. Height/rx/font/offset unchanged and matching.
5. **Axis labels** (cause: decision recorded by the parent, keep the shared design-system convention) — frame `Sep 02` dates and whole-dollar ticks (`$100`); capture ISO dates and `$5.00` ticks. Present at the merge base (merge-base baseline: a9f37e898a4e54b279716f66b0b54a0793c3264b:stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-darwin.png).
6. **Phase default toggles** (cause: decision recorded by the parent, Overall-only default) — frame all pressed; capture Overall only. Present at the merge base (merge-base baseline: a9f37e898a4e54b279716f66b0b54a0793c3264b:stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-darwin.png).
7. 0% delta: frame `▼0%`; app no arrow — named cause: decision `permanent-cost-label`. Not drawn by the fixture.
8. Frame's dashed `.note` box — a preview annotation, not app ink.

### Matrix (every frame; 16 elements × 11 columns)

`measure-visual-properties.sh --props matrix` paired 0 elements (output saved under `visual-verification/matrix/<frame>.json`); cells filled from per-element calls and the two sources:

| element | kind | content | colour | size | position | stretch | alignment | border | fill | radius | shadow |
|---|---|---|---|---|---|---|---|---|---|---|---|
| page title | heading | `Weekly costs` both | fg `#eee` / app text-primary | 20px vs app h1 | top of content | left-aligned in container | left | n/a — text | n/a — text | n/a — text | n/a — text |
| subtitle | caption | `Wednesday-to-Wednesday cost, last 8 weeks.` both | mute | 14px | below title | left | left | n/a — text | n/a — text | n/a — text | n/a — text |
| note box | extra in frame | annotation | `--card` | full width | below subtitle | full | n/a | dashed 1px | `--card` | 6px | none |
| section title ×2 | heading | departure 2 | fg / muted caps | 16px vs caption | above toggles | left | left | n/a — text | n/a — text | n/a — text | n/a — text |
| toggle chip | control | label + 10px dot | `--line` border | pill | row, gap 8px | content width | centred | 1px `--line` | `--card` | 999px | none |
| toggle dot | icon | series colour | departure 3 | 10px | before label | — | centred | none | series | 50% | none |
| toggle off | state | opacity .4 both | dimmed | same | same | same | same | same | same | same | none |
| y tick | text | departure 5 | mute | 11px | left of plot | right-aligned | end | n/a — text | n/a — text | n/a — text | n/a — text |
| x tick | text | departure 5 | mute | 11px | below plot | centred on week | middle | n/a — text | n/a — text | n/a — text | n/a — text |
| gridline | rule | — | `--line` | 1px | per tick | plot width | — | n/a — rule | `--line` | n/a — rule | n/a — rule |
| series line | rule | — | departure 3 | 2.5px both | per week | across gaps broken | — | n/a — rule | none | n/a — rule | n/a — rule |
| point | icon | — | series | r=4 | on line | — | centred | none | series | 50% | none |
| pill rect | control | — | stroke series | h15 rx7.5 both; width 40/84 both (source) | x+8 | fixed 40/84 both | — | 1px series | bg vs surface-panel (both the colour behind it) | 7.5 | none |
| pill cost text | figure | `$x.xx` both | fg / text-primary | 10px 600 | x+6, y+3.5 | — | left | n/a — text | n/a — text | n/a — text | n/a — text |
| pill delta tspan | figure | `▲n%`/`▼n%` | `--up`/`--down` both | 10px 600 | after cost, one space | — | left | n/a — text | n/a — text | n/a — text | n/a — text |
| leader line | rule | — | series | 1px | point+4 → pill | when displaced >2px both | — | n/a — rule | n/a — rule | n/a — rule | n/a — rule |

Values cited from the prototype `<style>`/`draw()` and `stats/web/src/styles.css` `.weekly-delta*`, `.weekly-line` where the pixels paired nothing.

### Sweeps (every frame)

text done (transcribed above); order done (frame: title, sub, note, h2, toggles, chart, h2, toggles, chart; capture: nav, title, sub, panel[caption, subtitle, toggles, chart] ×2 — differences are departures 2 and 8; no pill or line leaves its panel: the 2026-08-12 pills extend into the right margin the plot reserves); reach done (page 1666px tall, captured fullPage, last element the model·effort x-axis caption); derived done (deltas recomputed above; toggled capture drops Overall's pills only); rows done (toggle row gap 8px both; panel padding pre-existing); ink done (frame gridlines, series lines, leader lines, pill borders all present in capture; note-box dashed border absent — departure 8).

Frames: 5/5 composed (weekly-costs, weekly-costs-light, weekly-costs-toggled, weekly-costs-model-effort, weekly-costs-tooltip); `weekly-costs-empty` has no frame — the prototype draws no empty state; the tooltip state is not drawn by the prototype, composed against its resting state.

## Motion

None named.
