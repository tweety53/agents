# kan-1120-visual-verify-script-fills-matrix
> Design record of change kan-1120-visual-verify-script-fills-matrix (KAN-1120), written at landing. Each decision's own **Status:** line is the status.

## Context

`visual-verify-pass measure` (`stats/internal/guard/visualverifypass_measure.go`) runs
`measure-visual-properties --props matrix` per composed frame and writes
`visual-verification/matrix/<frame>.json`. Each paired row there already carries every measured
figure the verifier's 11-column matrix needs: `ink.*` for a text run,
`box`/`border`/`fill`/`radius`/`shadow`/`content`/`gap` for a fill band or a resolved control, and
`runs` for a rule. Nothing measures `content` as a transcription; that stays sweep 5's. A text
run's `stretch` and `alignment` need a container width, which the element obj's `left`/`right`
insets give.

Row shapes the filler must handle (`mvpMatrix`, `stats/internal/guard/measurevisualproperties.go:1558`):

- paired: `a`, `b`, `kind`, `measured`;
- unresolved: the pairing keys plus `unresolved`, no `measured`;
- missing / extra: `a` or `b` is null, no `kind`;
- clipped: the clipped leaves read `"clipped"` and the delta's `abs` is null.

Today `--props matrix` labels a text run and a control alike as `text/control`.

`stats/internal/guard/verifyreport.go` is the only parser of the verify report lines. It
already reads the manifest for `over-tolerance:` reconciliation (`vrOverTolerance`, L454) with the
`-rest` segment stripped. It checks the matrix table cell by cell in `visual-verification.md`
(`vrMatrixCells`, L556, against `vrNAColumns`, L179) and checks the presence of the
`matrix:`/`bands:`/`seams:` lines per composed frame (L389-399).

A remainder dispatch is triggered only by a guard exit 1 from close
(`stats/internal/guard/visualverifypass_close.go:45-74`); no size trigger exists. `flow change
save` already drops matrix-table rows (`skills/flow-contracts/finish-contract-run1.md:266-273`).

`spectre/specs/` does not exist, so no spec task applies. `README.md` mentions none of this, so
no README task applies.

This is one change because the manifest key, the guard that reads it and the verifier prompt
that stops writing the table must land together: any one alone leaves the guard refusing a
report the prompt tells the verifier to write.

## Decisions

### Where the 11 matrix cells are computed

**ID:** cells-in-measure
**Status:** active
**Chosen:** in `visual-verify-pass measure`, from the matrix JSON it already reads, written into each pair's `cells` in the manifest — the one caller that needs them computes them, once.
**Considered:** computing them in `measure-visual-properties --props matrix` and copying them into the manifest — rejected because every other caller of `--props matrix` would get a larger output, and the cells would be stored twice.

### Which hand-judged coverage lines stay

**ID:** keep-status-only-lines
**Status:** active
**Chosen:** the per-frame `- <frame id> sweeps: text done | order done | …` line and the per-view `- <view>: <PNG> — clean | <defect>` line stay as status-only lines, and the guard keeps checking both.
**Considered:** dropping them too — rejected because a skipped sweep or an unread capture would then read the same as a clean one and the guard could no longer refuse it, losing the kan-1016 and group-allowed-pairs-roles protections.

### How a row's class is known

**ID:** matrix-row-class
**Status:** active
**Chosen:** `mvpMatrix` adds a `class` key to every row — `unresolved` for an unresolved row; `missing` or `extra` for a row whose `a` or `b` is null; otherwise `text run` when only `ink` resolved; `control` when the box columns resolved for a `text/control` element (`measurevisualproperties.go:1593-1609`); `fill band`; `rule`. It is the only change to the JSON contract of `measure-visual-properties`.
**Considered:** guessing the class in the filler from which keys are present — rejected because the classification would then live away from the code that decides which props it measured.

### Where a text run's stretch and alignment come from

**ID:** text-stretch-from-element-insets
**Status:** active
**Chosen:** the element obj's `left`/`right` insets.
**Considered:** a second full-width `ink` call per run — rejected as a second measurement per text run for a figure the element obj already gives.

### A text run's content cell

**ID:** content-cell-for-text
**Status:** active
**Chosen:** reads `n/a — transcribed by sweep 5`, and the guard admits that for the `text run` class.
**Considered:** leaving the cell to be filled by hand — rejected because it keeps one hand-written column in a table nobody writes any more.

### What the record carries in place of the table

**ID:** record-cites-matrix-json
**Status:** active
**Chosen:** `visual-verification.md` cites `visual-verification/matrix/<frame id>.json` per frame and drops the hand table.
**Considered:** `measure` rendering a Markdown table into the record — rejected because it restates the JSON and regrows the record this change shrinks.

### The bands: and seams: lines

**ID:** bands-seams-lines-dropped
**Status:** active
**Chosen:** removed. A band or seam the verifier judges a departure goes on a departure line with its cause.
**Considered:** keeping them — rejected because the guard only checks them for presence, so they cost report size and remainder triggers and prove nothing.

### The spot-check line

**ID:** spot-check-shape
**Status:** active
**Chosen:** each composed frame gets `- <frame id> spot-check: #<row> <column> script <x> hand <y>; …`. It covers at least 3 rows, or every row when the frame has fewer; the rows are taken from rows not on the `over` list; at least one is a `control` or `fill band` row where the frame has one. The guard checks the line is present, the row count, that each row exists in `cells` and is not on the frame's `over` list, and that each `script` figure equals the manifest's. `<x>` is the cell's `b` (app-side) value, written as the manifest renders it (length in px, colour as hex), compared as a string; `<y>` is the verifier's hand figure for the same side.
**Considered:** an unchecked free-text spot-check — rejected because an unchecked line is skipped or invented at no cost.

### A spot-check that disagrees

**ID:** spot-check-disagreement
**Status:** active
**Chosen:** a hand figure that differs from the script's by more than `mvpOverLength`/`mvpOverColour` goes on a departure line with cause `mismeasure`, and the parent fixes it as a pipeline defect in the measuring tool.
**Considered:** the guard marking step 10 undone — rejected because the defect is the tool's, and re-running the verifier cannot fix it.

### Where the cells are stored

**ID:** cells-in-manifest-only
**Status:** active
**Chosen:** the manifest (with the per-frame matrix JSON they derive from).
**Considered:** a third file, `cells/<frame>.json` — rejected as one more file per frame with no second reader.

### How status-only is enforced

**ID:** status-only-by-prompt
**Status:** active
**Chosen:** the verifier prompt makes the `sweeps:` and per-view lines status-only; the guard's parse of them, `<sweep> done…`, stays unchanged.
**Considered:** a guard that refuses trailing narration — rejected because the size cost was in the matrix tables, and a stricter parse would invalidate saved reports for no gain.

### Manifest versioning

**ID:** manifest-no-version-bump
**Status:** active
**Chosen:** `cells` is added with no schema version or migration.
**Considered:** versioning the manifest — rejected because `measure` is its only writer and the guard re-reads a manifest rewritten on every run, so no old manifest outlives a pass.

### How rows that do not pair cleanly fill their cells

**ID:** unpaired-row-cells
**Status:** active
**Chosen:** a missing or extra row fills its present side from the element obj, its absent side and `delta` read null, and the columns the element obj cannot answer — content, border, fill, radius, shadow — read n/a, listed for `missing` and `extra` in `vrNAColumns`. An unresolved row fills `kind` from its element obj and reads `unresolved — <reason>` in every other column. A clipped leaf keeps its `clipped` value inside its cell, so no cell is the string `clipped`. `no mockup source` is no admitted n/a reason (plan-review-1, F4).
**Considered:** a `no mockup source` n/a on such rows — rejected because no filler rule produces it, and every missing, extra or unresolved row already sits on the `over` list with its own cause.

### Where the n/a columns live

**ID:** na-columns-shared
**Status:** active
**Chosen:** `vrNAColumns` (`stats/internal/guard/verifyreport.go`) is the one table: the filler writes n/a exactly where it lists a column for the row's class, and the guard admits n/a nowhere else. Its keys become the matrix classes — `text run` (gaining `content`), `control`, `fill band`, `rule`, `missing`, `extra` — and the hand table's `row`, `card` and `divider` go (plan-review-1, F5).
**Considered:** a second copy in the filler — rejected because two tables drift, and the guard would then refuse what the filler writes.

### Which leaves each column reads

**ID:** cell-leaves-per-class
**Status:** active
**Chosen:** a `control` or `fill band` row reads step 10's box columns from `measured`; a `text run` reads `ink.colour`, `ink.height` and `ink.width` from `measured` and its position, stretch and alignment from the element obj; a `rule`, `missing` or `extra` row reads the element obj alone, a rule's `runs` staying in the matrix JSON. Each cell's `delta` is the one the matrix row already carries — `measured.delta`'s, or the pairing's `gap`, `height` and `colour` — never recomputed.
**Considered:** a rule's `runs` lists as cell leaves — rejected because a list carries no delta and no single figure a spot-check can compare; recomputing a delta for every element leaf — rejected because the `over` list is already judged on the row's own deltas.

### What a spot-check's script figure is compared with

**ID:** spot-check-any-b-leaf
**Status:** active
**Chosen:** `<x>` must equal one of the `b` figures the named column's cell carries, each written as the `over` list's figures are — a number shortest-exact, a colour as its hex.
**Considered:** naming the leaf in each item, `#<row> <column>.<leaf>` — rejected as a longer line for the same check, where spot-check-shape names a column.

## Open questions

none

## Recommended approach

`--props matrix` labels each paired row's `class` (matrix-row-class). `visual-verify-pass measure`
then turns the verifier's column table (`skills/flow/visual-verify-verifier.md:435-447`) into
code: for every matrix row of every composed frame it writes one entry into the pair's `cells` —
`row`, `class` and the 11 columns (kind, content, colour, size, position, stretch, alignment,
border, fill, radius, shadow) — each cell either `{a, b, delta}` or `n/a — <why>` where the column
does not apply to the class (cells-in-measure, cells-in-manifest-only). The mapping per class:

- text run: `ink.*`, with stretch and alignment from the element obj's insets
  (text-stretch-from-element-insets) and content `n/a — transcribed by sweep 5`
  (content-cell-for-text);
- control and fill band: `box`/`border`/`fill`/`radius`/`shadow`/`content`/`gap`/`content.padding`;
- rule: the element obj, its `runs` staying in the matrix JSON (cell-leaves-per-class).

Which columns are n/a for each class is read from `vrNAColumns` (na-columns-shared); every other column is filled.

Rows that do not pair cleanly fill their cells per unpaired-row-cells; the `over` list already
requires the `mockup source` cause for an unresolved row.

The verifier stops transcribing (record-cites-matrix-json, bands-seams-lines-dropped): it reads
every composite, names a cause for every `over` row on the unchanged `over-tolerance:` line, still
crops interactive controls at 2–3x, and writes one checked `spot-check:` line per frame
(spot-check-shape, spot-check-disagreement). The `sweeps:` and per-view lines stay, status-only
(keep-status-only-lines, status-only-by-prompt).

`check-verify-report` drops the `matrix:`/`bands:`/`seams:` presence checks and `vrMatrixCells`
over the record, and adds a check over the manifest's `cells` for every composed frame — no empty
cell, n/a only where `vrNAColumns` lists the column for the class (na-columns-shared), and
`unresolved — <reason>` only on an unresolved row — plus the `spot-check:` line check
(spot-check-any-b-leaf). `over-tolerance:` reconciliation, the `sweeps:` checks and
`vrUnaccounted`'s per-view checks are unchanged.

Rejected: computing cells in `--props matrix` (cells-in-measure); dropping the status lines
(keep-status-only-lines); a Markdown table rendered into the record (record-cites-matrix-json);
a separate cells file (cells-in-manifest-only); an unchecked spot-check (spot-check-shape).

## Critical files

- `stats/internal/guard/measurevisualproperties.go` — `mvpMatrix` gains the `class` key; header contract updated; `mvpOverLength`/`mvpOverColour` are the spot-check disagreement thresholds.
- `stats/internal/guard/measure_visual_properties_test.go` — `TestMeasureVisualPropertiesMatrix` extended for `class`.
- `stats/internal/guard/visualverifypass_measure.go` — `vvmPair` gains `cells`; the filler that maps matrix rows to the 11 columns.
- `stats/internal/guard/visual_verify_pass_measure_test.go` — `TestVisualVerifyPassMeasureOverFigures` and `vvmFixture` are the template for the cells tests.
- `stats/internal/guard/visualverifypass.go` — header contract naming the manifest's keys.
- `stats/internal/guard/visualverifypass_close.go` — read only: the guard exit that triggers a remainder dispatch.
- `stats/internal/guard/verifyreport.go` — drops the matrix/bands/seams line checks and `vrMatrixCells` over the record; adds the manifest-cells and `spot-check:` checks; header contract.
- `stats/internal/guard/check_verify_report_test.go` — `TestCheckVerifyReportMatrixLinePerFrame`, `TestCheckVerifyReportComposedFrameLines`, `TestCheckVerifyReportMatrixCells`, `TestVrMatrixCellsMissingSide` rewritten to the new contract.
- `skills/flow/visual-verify-verifier.md` — step 9, step 10's matrix section (L433-464), step 11, the step-10-done paragraph (L658-662) and the report template (L700-709).
- `skills/flow/visual-verify.md` — L337-342 and L359-367.
- `skills/flow/SKILL-rationale.md` — L1106 and the kan-1015/1016/1017 entries (L1237-1277).

## Existing code to reuse

- `mvpMatrix` (`stats/internal/guard/measurevisualproperties.go:1558`) — the rows the filler reads.
- `vvmPair`, `vvmFixture` (`stats/internal/guard/visualverifypass_measure.go`, `visual_verify_pass_measure_test.go`) — the manifest shape and test fixture.
- `vrOverTolerance` (`stats/internal/guard/verifyreport.go:454`) — its manifest read, `-rest` stripping included, is reused by the cells and spot-check checks.
- `vrNAColumns` (`stats/internal/guard/verifyreport.go:179`) — the applicability table the cells check enforces.
- `mvpOverLength`/`mvpOverColour` (`stats/internal/guard/measurevisualproperties.go:1620-1621`) — the disagreement thresholds.

## Verification

- `cd stats && go test ./internal/guard/ -run 'MeasureVisualProperties|VisualVerifyPass|CheckVerifyReport'` — the class key, the cells per row shape (paired, unresolved, missing, extra, clipped), and the guard accepting a report with no matrix/bands/seams lines while refusing an empty cell, a wrong-class n/a, a missing or short `spot-check:` line, and a `script` figure that disagrees with the manifest.
- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` clean, and every `scripts/flow-guard check-*` guard in `.flow/project.md`'s `## lint` passes.

## Live check

None — the change edits `flow-guard` subcommands, each run per invocation from the worktree's own build, and skill prose; no daemon, store or scheduler holds state it changes. Task 3's `TestVisualVerifyPassMeasureCells`, which runs `measure` over `vvmFixture`, is the end-to-end check.

## Non-goals

- No change to the band and seam measurements, the tolerances, compose, or the over-list causes.
- No transcription by script: the text in each run stays sweep 5's work.

## Risks

- A text run's stretch and alignment are read against the app surface's insets, not its own container; inside a card that is coarser than a hand-cropped container. The spot-check and the `over` causes are the backstop.
- The class heuristic could misread a control as a text run, and that row's box columns would then read n/a. Only rows with a resolved box are classed `control`, so the error goes the safe way: fewer cells, never invented ones.
