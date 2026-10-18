## Context

Found running `/flow`'s self-review on gymie `kan-941-step-6-playwright-fidelity-specs-for-the-cardio`:
`flow self-review bundle` exits on `response body overflows the 1048576-byte cap`. The saved bundle
(2,828,523 bytes) carries the archived `visual-verification.md` with its matrices under
`## Element × property matrices`, each `### cardio-add-workout:cardio-add-workout-A1 matrix (A1)` and
the like, between its lines 1106 and 12058.
<!-- measured: wc -c k941-bundle.md (scratchpad copy of the served bundle) @ 02d83a03 → 2828523 -->

Cause: `elideMatrixRows` sets `inMatrix` only for a line starting `#### ` (decision
`bundle-text-filters`, archived change `self-review-kan-940-step-5-frontend-cardio-row-on-the-kan-928`),
the level that change's fixture used. `skills/flow/visual-verify-verifier.md` step 11 prescribes no
heading level for the matrix, and gymie's archived records use every level: of 26 archived
`visual-verification.md` files with a matrix heading, the shapes include `#### pin-P1 element ×
property matrix`, `### pin-P2 matrix`, `### X3 matrix (…)`, `## Matrix`, `## Element × property
matrix (all three frames)` — every one names `matrix` — and no `####` heading lacking `matrix` heads a
table in any archived record.
<!-- measured: grep -l -i matrix gymie/spectre/changes/archive/*/visual-verification.md | wc -l @ gymie develop 3c57055d → 26; awk scan for a table under a `####` heading without "matrix" → no hit -->

Applying the chosen rule to the saved bundle drops it to 120,924 bytes, under the cap.
<!-- measured: awk '/^#/{m=(tolower($0) ~ /matrix/); print; next} m && /^[[:space:]]*\|/ {next} {print}' k941-bundle.md | wc -c @ 02d83a03 → 120924 -->

Blast radius: `elideMatrixRows` has one caller, `Bundle`, applied to `visual-verification.md` alone;
the elision is described in one run-loaded sentence, in `skills/flow-contracts/finish-contract-run1.md`.
The change touches `git.go`, `bundle_test.go` and that contract. The rewording is acknowledged in
this change's `verbatim-moves.txt`. No saved memory encodes this defect.

## Decisions

### What makes a heading a matrix heading

**ID:** matrix-heading-by-name
**Status:** active
**Chosen:** any Markdown heading (outside a fence) whose text holds `matrix`, any case, at any level; its table rows up to the next heading are elided as before.
**Considered:** keep `####` and add `###` — the archive already holds `##` matrix headings, so the next verifier variant overflows again; every heading nested under `## Element × property matrices` — only kan-941 uses that parent heading, and the other 25 records would lose the elision a `####`-free shape gets; pin a heading shape in the verifier template and leave the parser — the archived records keep their shapes and the bundle still overflows for them.
**Residual risk:** a non-matrix table under a heading that happens to say `matrix` is elided; the elision line still counts its rows, and none exists in the archive.

### The verifier template is unchanged

**ID:** verifier-template-unchanged
**Status:** active
**Chosen:** leave `skills/flow/visual-verify-verifier.md` as is — it prescribes no heading level, so the parser fitting every level fixes the defect at its one consumer.
**Considered:** prescribing a `####` matrix heading in step 11 — a second rule for the verifier to follow, and archived records would still not conform.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — a one-line predicate change in a text filter with one caller, pinned by a new test that fails without it, and one contract sentence.
**Considered:** `flow-medium` — the bundle feeds every self-review, but the change only widens which rows are elided, and the existing `TestBundleElidesMatrixRowsAndPNGStatLines` still pins the `####` case, fences and the kept departures table.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

## Live check

Rebuild the `flow` binary from this tree and, from the gymie main checkout, run
`flow self-review bundle -change kan-941-step-6-playwright-fidelity-specs-for-the-cardio`: exit 0,
output about 121 KB, carrying `_(<n> matrix rows elided)_` lines under its `### … matrix (…)`
headings and no `| ` rows beneath them. The installed binary at `02d83a03` fails the same call with
`response body overflows the 1048576-byte cap`; that error after the rebuild means the fix did not
reach the served bundle (or flowd was not restarted onto the new build).
