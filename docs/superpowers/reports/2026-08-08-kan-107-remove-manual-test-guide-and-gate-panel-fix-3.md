# Panel fix round 3 — report

All five pass-3 findings (F14, F15, F16/F13, F17) fixed. Worktree:
`/Users/tweety53/Projects/agents/.worktrees/kan-107-remove-manual-test-guide-and-gate`.

## Files changed

- `openspec/changes/kan-107-remove-manual-test-guide-and-gate/design.md` and
  `docs/superpowers/specs/2026-08-08-kan-107-remove-manual-test-guide-and-gate-design.md` — F14: the
  "Guards, commit subject and tests" paragraph claimed `check-unfinished-work.sh` "keeps its exit
  codes unchanged: `0` clear, `1` outstanding, `2` cannot determine." The guard's own header states
  it never exits 1 — the verdict line carries `CLEAR:`/`OUTSTANDING:`, the exit status does not — and
  exits 2 only when it cannot determine anything (unreadable worktree, name outside the allowlist, or
  a file that exists but cannot be read). Reworded both copies, taken from the guard's own header, to
  state that contract while keeping the sentence's purpose: this change did not alter it. Both files
  carry identical replacement text, kept in step as required.
- `skills/myflow-contracts/plan-provenance.md` — F15: the table row at line 286 named the eleven
  counter-examples' homes as "this file, the delta spec, or an archived copy of either", which omits
  two of the eleven. Re-measured with the guard's own imported functions (see below) and rewrote the
  row to state the actual five homes: this file (five lines), the live delta spec
  (`openspec/specs/myflow-plan-provenance/spec.md`, two lines), an archived copy of that spec (two
  lines), an archived design document (one line), and an SDD ledger (one line).
- `scripts/check-plan-provenance.py` — F15, same defect in the neighbouring "9 lines containing `<`"
  measurement at lines 590-594: the comment claimed those 9 come from "`plan-provenance.md`'s own
  prose, the live delta spec, or an archived copy of either", which likewise misses one line living in
  an archived **design document** rather than an archived copy of either named file. Reworded to add
  that home. Re-ran the script's own described measurement (37 exempted, 9 on `<`) against the current
  tree and it reproduced exactly: 37 and 9, confirming the comment's counts were already correct and
  only the home list needed the fix.
- `scripts/test-check-unfinished-work.sh` — F16/F13 (one finding, both flip together): case 2 (lines
  ~192-202) created `docs/manual-test/guide.md`, then `rm -rf`'d the directory *before* calling the
  guard, so the guard saw a tree byte-identical to case 1's and the case proved nothing about a
  leftover guide. Chose the "assert with the directory still present at invocation time" option: the
  case now creates the guide and leaves it in place, then asserts `CLEAR:`. This is directly
  observable — the guard is handed a tree that genuinely differs from case 1's — and it would fail if
  anyone reinstated a presence check for `docs/manual-test/`. Verified by mutation (see below), then
  restored the guard byte-for-byte.
- `skills/myflow-status/SKILL.md` — F17: lines 219-224 cited `/myflow-do` section 6 as canonical for
  the run-instructions resolution, stated the invariant "never the project's declared base", and then
  said "Do not restate the resolution rules here" — a blanket clause that reads as forbidding the
  invariant it had just stated two lines above (`skills/myflow-do/SKILL.md:338` states the same
  invariant verbatim). Reworded only the guardrail sentence: it now forbids restating the resolution
  **procedure** ("the steps that compute each app root, start command and URL") and explicitly says
  naming the invariant above is not one of those steps. One sentence changed; the invariant clause is
  untouched.
- `.superpowers/sdd/final-review-panel.md` — all five pass-3 `finding-status:` markers (F13, F14,
  F15, F16, F17) flipped `open` → `fixed`; `findings-total: 17` unchanged. (This file is gitignored —
  `.gitignore:2` — so it was edited but could not be `git add`ed; that is expected and not an error.)

## F15 — measurement

Re-derived the eleven counter-example homes with the guard's own imported functions, per the
finding's instruction, rather than reimplementing the matching logic:

```
python3 <scratchpad>/measure_veto_cost.py   # from panel-fix-2 (F10), reused
```

reproduced the same 12-line / 11-counter-example set already recorded in the pass-2 fix report:
5 lines in `skills/myflow-contracts/plan-provenance.md` (198, 213, 241, 248, 287), 2 in
`openspec/specs/myflow-plan-provenance/spec.md` (142, 156), 2 in the archived copy of that spec
under `openspec/changes/archive/2026-07-30-kan-20-.../specs/myflow-plan-provenance/spec.md` (96,
110), 1 in the archived design document
`openspec/changes/archive/2026-07-30-kan-20-.../design.md:112`, and 1 in the SDD ledger
`docs/superpowers/ledgers/2026-07-30-kan-20-....md:22`. That is 11; the twelfth
(`openspec/changes/archive/2026-07-29-kan-14-plan-provenance/tasks.md:139`) is the genuine
quotation, not a counter-example, matching the existing table row below it.

Separately, imported `CLAIM_RE`, `_is_quoted`, `_has_escaped_delimiter`, `_quote_regions` and
`_QUOTE_PAIRS` from `scripts/check-plan-provenance.py` to reproduce the neighbouring "37 exempted, 9
on `<`" measurement the script's own comment describes (escape + class-wide vetoes only, angle-bracket
veto left out), run via a new one-off script at
`/private/tmp/claude-501/.../scratchpad/measure_angle_cost.py` (not committed to the repo). It
reproduced 37 and 9 exactly, with the 9 hits' homes being: 3 in `plan-provenance.md` (241, 248, 287),
2 in the live delta spec (142, 156), 2 in the archived copy of the spec (96, 110), 1 in the archived
design document (112), and 1 already carrying a provenance tag (the genuine quote at
`archive/2026-07-29-kan-14-plan-provenance/tasks.md:139`). No counts changed in either measurement;
only the prose naming the homes was corrected.

## F16/F13 — mutation proof

1. Backed up `scripts/check-unfinished-work.sh` to `/tmp/check-unfinished-work.sh.bak`.
2. Inserted a mutation immediately before the `CLEAR:`/`OUTSTANDING:` branch: `if [ -e
   "$WORKTREE/docs/manual-test" ]; then add "MUTATION: docs/manual-test/ is present"; fi` —
   reinstating a presence check the removed guide-guard used to have.
3. Ran `./scripts/test-check-unfinished-work.sh`: the rewritten case 2 **failed** —
   ```
   FAIL: a leftover docs/manual-test/ present at invocation time is not a signal: expected a line
   beginning CLEAR:, got: OUTSTANDING: ... — MUTATION: docs/manual-test/ is present
   ```
   proving the case is load-bearing against exactly the regression it exists to catch.
4. Restored `scripts/check-unfinished-work.sh` from the backup. `diff -q` against the backup reported
   no differences (byte-identical). Re-ran the suite: `check-unfinished-work: all cases pass`.

## Panel record guard

```
./scripts/check-unfinished-work.sh . kan-107-remove-manual-test-guide-and-gate
```

printed `CLEAR: . — every plan item is checked and no finding is open`.

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

None. All five (F13, F14, F15, F16, F17) fixed.

## No commits

Nothing committed or pushed. `git add` was run only on the five files listed above (never
`openspec/` or `docs/superpowers/`); everything else stays uncommitted/staged exactly as the prior
fix rounds left it.
