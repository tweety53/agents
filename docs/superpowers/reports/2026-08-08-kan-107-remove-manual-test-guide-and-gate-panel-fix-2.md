# Panel fix round 2 — report

All five pass-2 findings (F7, F10, F11, F12, F13) fixed in one pass. Worktree:
`/Users/tweety53/Projects/agents/.worktrees/kan-107-remove-manual-test-guide-and-gate`.

## Files changed

- `scripts/test-check-unfinished-work.sh` — F7: rewrote case 8f (lines ~809-833). The planted
  panel record under `$PLANTED/.superpowers/sdd/final-review-panel.md` was dropped — `PANEL` is
  built from `$WORKTREE` alone and never from `$NAME`, so it was unreachable by construction. The
  name passed to the guard changed from the absolute `"$PLANTED/openspec/changes/clear"` to the
  relative `"../../../$(basename "$PLANTED")/openspec/changes/clear"`, which genuinely resolves
  onto the planted tree once concatenated by the guard (`$WT` and `$PLANTED` are both `mktemp -d`
  siblings directly under the same `TMPDIR`). The comment was rewritten to state what the case
  actually proves and to record the mutation proof. Also F13: case 2 (lines ~192-199) previously
  ran `rm -rf "$WT/docs/manual-test"` on a directory `new_fixture` never creates, so it was
  byte-identical in effect to case 1. Chose to make it genuinely distinct rather than delete it: it
  now creates `docs/manual-test/guide.md`, removes the directory, and asserts the worktree still
  reads `CLEAR` — proving a worktree that once carried the guide is not treated differently.
- `skills/myflow-contracts/plan-provenance.md` — F11: the table row at line 285 named only "this
  file and the delta spec" as homes for the counter-example lines; now reads "this file, the delta
  spec, or an archived copy of either", matching `check-plan-provenance.py:591-593`'s own phrasing.
  F10: measured, no text changed — see below.
- `skills/myflow-status/SKILL.md` — F12: the citation at lines 219-224 restated the two deltas
  resolution follows but dropped `/myflow-do`'s own "never the project's declared base" invariant.
  Added that clause in place, without expanding the citation into a second copy of the resolution
  rules.
- `.superpowers/sdd/final-review-panel.md` — all five pass-2 `finding-status:` markers flipped
  `open` → `fixed`; `findings-total: 13` unchanged.

## F7 — mutation proof (both runs)

1. Backed up `scripts/check-unfinished-work.sh`, then deleted the allowlist `case "$NAME" in ...
   esac` block at lines 112-118 (confirmed by `grep -n` before mutating). Ran
   `./scripts/test-check-unfinished-work.sh`: case 8f's two assertions **failed**, exactly as
   expected —
   ```
   FAIL: traversal: expected exit 2 from the allowlist, got rc=0 out=CLEAR: /var/.../unfinished-work-test.zVsrAK — every plan item is checked and no finding is open
   FAIL: traversal: the guard produced a verdict from outside the worktree: CLEAR: /var/.../unfinished-work-test.zVsrAK — every plan item is checked and no finding is open
   ```
   The verdict is `CLEAR`, not `OUTSTANDING` — proving the repaired relative name genuinely
   resolves onto the planted tree once the allowlist stops rejecting it, and that the allowlist,
   not path-concatenation failure, is what stops it in the real guard.
2. Restored the file from the pre-mutation backup. `diff` against the backup reported no
   differences (byte-identical, F1's earlier "signal two" fix and every other prior edit intact).
   Ran the suite again: `check-unfinished-work: all cases pass`.

## F10 — measurement

Re-derived the "twelve lines / eleven counter-examples" count using the guard's own functions
rather than reimplementing the matching, per the finding. Loaded `CLAIM_RE`, `quotation_regions`,
`_is_quoted`, `_encloses` and `_QUOTE_PAIRS` directly from `scripts/check-plan-provenance.py` via
`importlib`, and implemented only the lenient scanner the doc's own snippet describes (escape and
angle-bracket vetoes ignored; a second opener re-arms `pending` instead of vetoing the class; a
leftover `pending` is dropped instead of vetoing the class) — everything else (the strict pass, the
enclosure test) comes from the real module. Ran it over every file from `git ls-files '*.md'`.

Command: `python3 /private/tmp/.../scratchpad/measure_veto_cost.py` (script committed to the
scratchpad, not the repo).

Raw output:
```
docs/superpowers/ledgers/2026-07-30-kan-20-widen-plan-provenance-guard-scan-scope.md:22: 99 tests
openspec/changes/archive/2026-07-29-kan-14-plan-provenance/tasks.md:139: 197 tests
openspec/changes/archive/2026-07-30-kan-20-widen-plan-provenance-guard-scan-scope/design.md:112: 77 tests
openspec/changes/archive/2026-07-30-kan-20-widen-plan-provenance-guard-scan-scope/specs/myflow-plan-provenance/spec.md:96: 77 tests
openspec/changes/archive/2026-07-30-kan-20-widen-plan-provenance-guard-scan-scope/specs/myflow-plan-provenance/spec.md:110: 77 tests
openspec/specs/myflow-plan-provenance/spec.md:142: 77 tests
openspec/specs/myflow-plan-provenance/spec.md:156: 77 tests
skills/myflow-contracts/plan-provenance.md:198: 99 tests
skills/myflow-contracts/plan-provenance.md:213: 99 tests
skills/myflow-contracts/plan-provenance.md:241: 77 tests
skills/myflow-contracts/plan-provenance.md:248: 77 tests
skills/myflow-contracts/plan-provenance.md:287: 197 tests
TOTAL: 12
```

Twelve lines, confirmed against the guard's own functions. Of those twelve, the one at
`openspec/changes/archive/2026-07-29-kan-14-plan-provenance/tasks.md:139` is the genuine
quotation the table's second row already documents by name (an archived table row whose
`<!-- measured: … -->` comment quotes `"197 tests"`); the other eleven — including the reference
to that same quote at `plan-provenance.md:287` — are deliberate counter-examples. This settles the
two conflicting pass-2 estimates (12/11 from the fix-round-1 report vs. 10/9 from Bugbot's lenient
reimplementation) in favor of **12 and 11**: the existing prose at lines 279/285 was already
correct, so no numeric text changed. Only the table row's list of homes changed (F11, above).

## Verification — full `## test` and `## lint` lists from `.myflow/project.md`

All twelve `## test` scripts: exit 0 — `test-setup.sh` (192 assertions, 0 failures),
`test-check-references.sh`, `test-check-plan-provenance.sh`, `test-check-finish-preflight.sh`,
`test-preserve-session-records.sh`, `test-check-unfinished-work.sh`, `test-check-cleanup-complete.sh`,
`test-gather-self-review-context.sh`, `test-uncommitted-review-package.sh`,
`test-check-task-build-green.sh`, `test-check-workspace-isolation.sh`, `test-check-contract-budget.sh`.

All six `## lint` guards: exit 0 — `check-vocabulary.sh` (both guards clean),
`check-references.sh` (all referenced sections resolve), `check-plan-provenance.sh` (3 files
scanned, all provenance stated), `check-task-build-green.sh` (silent 0), `check-workspace-isolation.sh`
(ISOLATION-OK), `check-contract-budget.sh` (BUDGET-OK: 24 contract files within budget).

## Findings left open

None. All five (F7, F10, F11, F12, F13) fixed.

## No commits

Nothing committed or pushed. All edits left uncommitted in the worktree, as instructed.
