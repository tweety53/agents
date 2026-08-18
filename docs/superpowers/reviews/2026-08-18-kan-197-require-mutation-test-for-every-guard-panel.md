# Final review panel — kan-197-require-mutation-test-for-every-guard

Roster: `light` (recorded). Required slots: Primary, Principles, Code review (low).
Panel model: sonnet (`models.reviewPanel`). Merge base: a805d29. Diff: 11 files, +1047/-80.
Optional slots: none — no trigger fired that the light preset escalates.

Panel diff size: `scripts/check-panel-diff-size.sh <worktree> a805d29` reported **1127, under cap**,
exit 0. No operator prompt required. Recorded per the contract's every-run rule.

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Major | `scripts/check-stage-mark-calls.sh:206-256` | 28 of 33 corpus members are static expected-zero declarations; the scan target is all of `skills/` when only 5 files can ever carry a mark |
| F2 | Principles | Minor | `scripts/check-references.sh:467-497` | 29 of 59 declared — same shape, milder; the reviewer explicitly declined to block on it |
| F3 | Principles | Minor | `scripts/lib/coverage.sh:154-172` | a `coverage_declare` for a member never `coverage_record`ed is silently inert; the declaration list can only grow, never self-prune |
| F4 | Principles | Minor | `scripts/check-vocabulary.sh:517` | coverage-violation stderr output has no trailing newline |
| F5 | Primary | Major | `.myflow/project.md` | `scripts/test-check-stage-mark-calls.sh` is absent from the `## test` list, so task 5's new coverage assertions never run in the declared test suite |
| F6 | Primary | Minor | `scripts/check-references.sh:504`, `scripts/check-stage-mark-calls.sh:242` | both iterate an array by value unguarded, against the discipline `lib/coverage.sh`'s own header calls load-bearing |
| F7 | Code review (low) | **Critical** | `scripts/check-stage-mark-calls.sh:327` | a target that enumerates to zero files never calls `coverage_record`, so it vanishes from coverage entirely — the vacuous pass this change exists to close, reproduced inside the fix for it |
| F8 | Code review (low) | Major | `scripts/check-vocabulary.sh:403`, `scripts/check-stage-mark-calls.sh:327` | `coverage_record`'s non-zero return is never read, and both run under `set -uo pipefail` without `-e`, so a rejected call is swallowed and the verdict contradicts its own breakdown |

## Slot notes

**Principles (slot 2)** — no Critical. Two Important (recorded here as F1 Major / F2 Minor to match
this panel's severity vocabulary) and two Minor. Confirmed clean with evidence: all four adoptions
genuinely share `lib/coverage.sh` rather than re-implementing round it; no new suppressions; no lint
weakening; the KAN-73 defect class verifiably closed (reproduced the undeclared-zero catch live); and
design.md's partial-collapse limitation judged honestly stated with sound reasoning.

The slot sharpened the parent's own concern into a distinction the parent had not drawn: a zero
because **coverage silently vanished** (signal) versus a zero because **the member was never a
candidate** (noise). Under `skills/`, 28 of 33 files — contract docs, rationale files, reviewer
prompts — can never carry a stage mark, so declaring them records no judgement and costs a line every
time a new such file is added. Its proposed remedy is to narrow the scan target to the members
capable of holding the property, which preserves detection (a SKILL.md that lost all its marks is
still a member and still fires) while emptying the declaration list.

PARENT VERIFICATION, both claims checked directly rather than taken from the report:
  - F1 ratio: `check-stage-mark-calls.sh | tail -1 | tr '·' '\n'` gives 33 members, 28 declared.
  - F3: recording one real member, declaring a ghost that is never recorded, then running
    coverage_verdict returns rc=0 with no complaint. Confirmed silently inert.

F2 is deliberately recorded at the severity its own reviewer assigned it — it declined to block,
reasoning that `check-references.sh`'s corpus has a legitimate reason to include files carrying no
cross-reference, since that property is not structural the way "is a SKILL.md" is.

**Primary (slot 0)** — no Critical. One Major (F5), one Minor (F6), and a judgement that partly
contradicts the Principles slot's, usefully.

PARENT VERIFICATION:
  - F5 CONFIRMED: `grep -n stage-mark-calls .myflow/project.md` returns only line 115, under `## lint`.
    The harness has never been in `## test` — pre-existing since KAN-172, not introduced here. But
    this change added ~75 lines of coverage assertions to a file the declared test run never invokes,
    which is precisely the thesis the change argues. The parent's own verification ran it only because
    it globbed `scripts/test-*.sh` directly rather than following the project's declared list.
  - F6 CONFIRMED at both cited lines: `for f in "${EXPECTED_ZERO_FILES[@]}"` iterates by value with no
    length guard, while `check-vocabulary.sh:510` guards its equivalent behind `[[ ${#...[@]} -gt 0 ]]`.
    The failure is real on this machine: `/bin/bash 3.2.57` with `set -u` and an empty array gives
    `ARR[@]: unbound variable`. Latent only because both arrays are non-empty literals today.
    (The parent's first grep missed this; the pattern was escaped wrong, not absent. Primary was right.)

**The two slots disagree on the remedy for the declaration-quality problem, and both are worth
keeping.** Principles would narrow the corpus so non-candidates stop being members at all. Primary
would keep the corpus and fix the reasons: it observed that the 29 and 27 entries share ONE boilerplate
reason string attesting to *how* the zero was measured ("ran the association logic against this file")
rather than *why* it is legitimate by design — which is weaker than the requirement's own words,
"legitimately checks nothing". It recommends batching entries by category with a per-category
justification, and explicitly endorses dropping NEITHER guard.

The parent reads these as complementary rather than competing: narrowing is the right fix where
non-candidacy is structural (`check-stage-mark-calls`: only a SKILL.md or pipeline.md can carry a
mark), and better-reasoned declarations are the right fix where it is not (`check-references`: any
Markdown file may legitimately carry no cross-reference).

Primary also confirmed, independently and against a sandboxed fixture rather than the worktree:
  - the partial-collapse limitation generalises to all four guards and matches both design.md and
    Requirement 2's third scenario — spec, design and measured behaviour agree, nothing overpromised;
  - task 5's flipped cases 10 and 14 are corrected expectations, not weakened assertions — case 14
    still asserts the guess-detection it originally tested, via a negative match;
  - both spec requirements are delivered and nothing extra shipped, verified by reading full guard
    sources rather than only the diff.

**Code review (low), slot 3)** — one Critical, one Major, both with runnable reproducers, both
reproduced by the parent verbatim. The cheapest slot again returned the most severe finding, as it
did on KAN-73.

PARENT VERIFICATION:
  - F7 CONFIRMED: `./scripts/check-stage-mark-calls.sh skills/myflow-fast /tmp/emptydir` prints
    `clean (1 call(s) checked)`, exit 0, and the empty target appears nowhere in the breakdown.
    `find` yields no files, the per-file loop's own `[[ -n "${f:-}" ]] || continue` guard skips it, so
    `coverage_record` is never called for that target — and `coverage_verdict` only iterates members
    that WERE recorded. An un-recorded member is invisible rather than caught.
  - F8 CONFIRMED: passing the same file twice makes the library reject the duplicate on stderr, the
    caller ignores the return, and the guard prints `clean (2 call(s) checked)` while its own
    breakdown shows `1`. A visible self-contradiction in successful output, exit 0.

**F3 and F7 share one root cause**, which the fix should address once rather than twice:
`coverage_verdict` reasons only over members that were *recorded*. So a member declared but never
recorded is silent (F3), and a member expected but never recorded is silent (F7). The library knows
what it was told, never what it should have been told. `check-vocabulary.sh` already avoids the F7
half by setting its count "regardless of what follows, including the early return" — that discipline
is the shape of the fix, and it belongs in the library rather than in each caller's memory.

Slot 3 also answered the two questions it was set, and both answers are reassuring rather than
defects: `check-vocabulary.sh`'s empty `EXPECTED_ZERO_TARGETS=()` is safe on bash 3.2 because its one
use site is length-guarded, and when a target first legitimately empties it becomes an undeclared-zero
violation, which is the intended fail-loud behaviour. Case 24's bare real-repo run is a documented
fragility rather than a hidden self-fulfilling assertion, though it is coupled to the live state of
two files.

It deliberately withheld DRY and performance observations (a double scan in `coverage_report`, a
redundant per-skill awk, four separate verdict-to-violation translations) as out of scope for a
low-effort correctness pass. Recorded here so they are not lost.

## Panel totals

8 findings: 1 Critical, 3 Major, 4 Minor. **All 8 fixed and pinned by tests.**

## Fix round

Applied on the panel-fix model (sonnet), each fix folded into the commit that introduced it via
`git commit --fixup` plus one `git rebase --autosquash`. Post-fix shas: task 1 `3999f92`,
task 2 `a785b1b`, task 3 `e8cd7ae`, task 4 `02f081b`, task 5 `3ac2677`.

PARENT VERIFICATION, every one re-tested rather than accepted from the report:
  - F7 (Critical): `check-stage-mark-calls.sh skills/myflow-fast /tmp/emptydir` now prints
    `/tmp/emptydir:0: 0 checked, and not declared expected-zero (coverage)`, exit 1. Before the fix
    the same command reported clean at exit 0 with the target absent entirely.
  - F8 (Major): the duplicate-target invocation now refuses at exit 2 instead of printing a verdict
    that contradicted its own breakdown.
  - F3 (Minor): the ghost-declaration reproducer now returns rc=1 at library level.
  - F1 (Major): the corpus narrowed from 33 members / 28 declared to **8 members / 3 declared**, and
    detection is preserved — renaming a declared member away fires
    `declared expected-zero but never recorded — stale declaration, or this member was never scanned`.
  - F5 (Major): `test-check-stage-mark-calls.sh` is now in `.myflow/project.md`'s `## test` list, and
    the fix round verified no other `scripts/test-*.sh` was missing.

**The fix agent under-claimed its own F3 fix, which is the right direction to err.** It disclosed a
residual — "a declaration silently drops rather than firing loud if its target vanishes from the real
repo" — and asked for the panel's eyes on it. Tested directly: on the default scan that residual does
**not** hold; a renamed-away declared member fires the stale-declaration violation correctly. The
caveat applies only to argument-scoped runs, where declaration is deliberately limited to the current
corpus so that sandboxed fixtures reusing real declared paths do not false-fail. That scoping is a
sound call: the alternative was weakening F3's check to keep ~40 existing test cases green.

**Verification after the fix round:** 25/25 harnesses pass, 10/10 lint guards exit 0, all four guards
exit 0 against the real tree, and all five commit-fields guards exit 0 on the rewritten shas.

**Panel clean: zero open findings at any severity.**

## Marker block

findings-total: 8
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
