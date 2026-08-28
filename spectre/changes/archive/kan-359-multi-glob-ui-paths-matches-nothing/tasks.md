# Tasks — kan-359

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Three tasks: the parser fix, the validator made to agree with it, and end-to-end proof against both
real declarations. `design.md` is canonical for the decisions.

**Baseline, measured before any edit:**

- `scripts/test-check-visual-trigger.sh` has 8 cases naming `ui paths`, all using the same
  single-glob value, none containing a comma.
  <!-- measured: grep -c 'ui paths' and grep -o "ui paths\` | [^|]*" | grep -c ',' @ 182ad31 -->
- 37 harnesses in `scripts/`, all passing.
  <!-- measured: for t in scripts/test-*.sh; do bash "$t"; done @ 182ad31 -->

- [x] 1. Split on commas first, then strip each element

`scripts/check-visual-trigger.sh` strips the backticks surrounding the whole `ui paths` cell before
splitting on commas, so every element but the first and last keeps a stray backtick and matches
nothing. Reverse it: split, then strip each element's own backticks and surrounding whitespace.

Write the harness cases first — **RED before GREEN**. Cover two globs, three globs, a mix of
backticked and bare elements, and **a glob containing a space**, which an existing case already
pins and must keep working.

**Files:** `scripts/check-visual-trigger.sh`, `scripts/test-check-visual-trigger.sh`
**Tests:** `scripts/test-check-visual-trigger.sh`
**Regression:** reverting this commit returns every multi-glob project to a silent stage skip — the
failure the whole visual-verification change exists to prevent.
**Baseline:** before=8 after=12 cases naming `ui paths` in that harness
<!-- predicted: grep -c 'ui paths' scripts/test-check-visual-trigger.sh after task 1 -->
**Commit:** `fix(scripts): strip ui paths backticks per element, not per cell`
**Build:** green

- [x] 2. The validator rejects what the trigger cannot parse

`scripts/check-visual-verification.sh` reports `VISUAL-OK` on a `ui paths` value the trigger guard
cannot use. Apply the same parse and report a value that yields no usable glob.

**This is the worse half of the defect**, per `design.md`'s `validator-agrees-with-trigger`: the
validator is the guard an author runs against their own declaration, so it is the one whose wrong
answer misleads.

Harness case verified RED against the current validator.

**Files:** `scripts/check-visual-verification.sh`, `scripts/test-check-visual-verification.sh`
**Tests:** `scripts/test-check-visual-verification.sh`
**Regression:** reverting this commit lets the validator bless a declaration the trigger guard
silently ignores.
**Baseline:** before=25 after=26 cases in that harness
<!-- predicted: the harness's own case count, before and after task 2 -->
**Commit:** `fix(scripts): reject a ui paths value the trigger cannot parse`
**Build:** green

- [x] 3. Prove both real declarations trigger

Validating is not triggering, and this defect is precisely the gap between them.

Run the trigger guard against a real changed path from **each** real declaration:

- this repository — `stats/web/src/App.tsx` against its own `.flow/project.md`;
- Gymie — a gateway path and a domain-module path against the declaration on PR #42
  (`/Users/tweety53/Projects/gymie`, branch `flow/visual-verification-declaration`). **Read-only:
  do not edit or commit in that repository.**

Record the outcome in the change's own artifacts. **If Gymie's declaration still fails after tasks 1
and 2, that is a finding, not a step to skip.**

**Files:** `spectre/changes/kan-359-multi-glob-ui-paths-matches-nothing/design.md`
**Tests:** none added — this task runs the shipped guards against real declarations rather than
fixtures
**Regression:** reverting this commit removes the only evidence that the fix works on the
declarations it was written for, as opposed to on test fixtures.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `docs(spectre): record that both declarations now trigger`
**Build:** green

- [x] 4. `glob_to_ere` measures the wrong variable

Task 3 proved the change had not achieved its own purpose: after tasks 1 and 2, Gymie's declaration
still fails to trigger for six of its eight globs.

**A second, independent bug**, pre-existing and unrelated to the backtick defect.
`scripts/check-visual-trigger.sh:248` reads:

```bash verified:reproduced standalone in bash 3.2 and 5.x — the two-line form yields the correct length, the one-line form does not
local g="$1" out="" i=0 n=${#g} c nc
```

`local`'s arguments are expanded **before** `local` executes, so `${#g}` measures the **caller's**
`g`, not `$1`. The per-element split loop uses `g` as its own loop variable, so the caller's `g` is
in scope — and every glob longer than it is silently truncated mid-conversion.

```text verified:run against the guard at a7e6904 with Gymie's real declaration from PR #42
outer `g` length                              19
gymie-frontend/webApp/src/Main.kt             exit 0  (its glob is <= 19 chars)
src/app/src/main/resources/application.yml    exit 0
src/gateway/src/main/resources/application.yml            exit 1  — should match
src/app/auth/.../infrastructure/presentation/web/A.kt     exit 1  — should match
```

Fix it — the smallest correct form is a separate statement for the length, so it is evaluated after
`g` is assigned. **Check the file for any other `local` that expands a variable it assigns on the
same line**, since the same quirk applies wherever that pattern appears.

**Pin it with a harness case using a glob longer than any plausible caller variable**, verified RED.
A case using only short globs cannot catch this — which is exactly why nothing did.

**Files:** `scripts/check-visual-trigger.sh`, `scripts/test-check-visual-trigger.sh`
**Tests:** `scripts/test-check-visual-trigger.sh`
**Regression:** reverting this commit silently truncates every long glob, so a declaration triggers
on its short entries and ignores its long ones — the shape hardest to notice, since it half-works.
**Baseline:** before=12 after=13 cases naming `ui paths` in that harness
<!-- predicted: grep -c 'ui paths' scripts/test-check-visual-trigger.sh after task 4 -->
**Commit:** `fix(scripts): evaluate the glob length after assigning it`
**Build:** green

## Part 2 — the review panel's findings

Two slots blocked on the same function, and the findings compose: extracting it means fixing its
complexity once rather than twice.

- [x] 5. Extract `trim_glob_element`, and strip in one pass

`trim_glob_element` is **byte-identical** in `check-visual-trigger.sh` and
`check-visual-verification.sh` — same body, same signature, no parameterisation needed for either
caller.

**Extraction is smaller than the duplication, not larger**, and this was mis-framed when the
operator first declined it. Both guards **already** source from `scripts/lib/`
(`sanitize-display.sh`, `strip-bom.sh`) and **already** share `visual-table-cells.awk` for this same
table's cells. The plumbing exists in both files; this is one new lib plus two `source` lines.

```text verified:diff of the awk-extracted function bodies, and grep for existing source lines, run in this worktree
trim_glob_element        identical in both guards, 18 lines
existing lib sources     check-visual-trigger.sh 2, check-visual-verification.sh 3
shared awk lib           visual-table-cells.awk, both files
```

**While extracting, replace the char-at-a-time strip with a single pass.** Both copies remove
leading and trailing whitespace and backticks one character at a time (`s="${s:1}"`, `s="${s%?}"`),
each a full string copy — quadratic in the padding length. Measured on a hostile
`.flow/project.md`:

```text verified:/usr/bin/time -p against scripts/check-visual-trigger.sh, run in this worktree
 8,000 backticks per side ( 16 KB)  0.05s
20,000 backticks per side ( 39 KB)  0.19s
60,000 backticks per side (117 KB)  0.84s
```

**Report the severity honestly where it is recorded**: the growth is quadratic, but the constant is
small and the worst case measured here is sub-second — this is a shape worth fixing while the code
is open, not a denial of service. It is also **pre-existing**: the same shape was present before
this change, via pattern trimming rather than a character loop.

`design.md`'s `no-shared-parser-yet` decision must be **superseded**, not deleted, recording that it
was reversed on the panel's evidence and on a corrected estimate of the cost.

The existing harnesses must pass **unchanged** — if a case needs editing, behaviour moved, so stop
and report.

**Files:** `scripts/lib/**`, `scripts/check-visual-trigger.sh`,
`scripts/check-visual-verification.sh`,
`spectre/changes/kan-359-multi-glob-ui-paths-matches-nothing/design.md`
**Tests:** the existing harnesses, unchanged, plus a case pinning the single-pass strip on a padded
element
**Regression:** reverting this commit restores two hand-aligned copies of one function — the
alignment this repository's own history says decays, having already had to extract the BOM handling
after it reached one guard of three.
**Baseline:** before=13 after=14 cases naming `ui paths` in the trigger harness
<!-- predicted: grep -c 'ui paths' scripts/test-check-visual-trigger.sh after task 5 -->
**Commit:** `refactor(scripts): share trim_glob_element and strip it in one pass`
**Build:** green

- [x] 6. Find out whether the harness can fail spuriously

A panel slot observed a single cold run of `scripts/test-check-visual-trigger.sh` failing its
multi-glob cases with the **same symptom shape as the original defect** — a second glob in a list
not matching — followed by many clean runs, sequential and parallel under load. It correctly
excluded the observation from its findings, being unable to reproduce it. A later attempt could not
reproduce it either.

**This matters more than it looks.** If the harness can fail spuriously in the exact shape of a real
bug, then every RED→GREEN result in this change is weakened: a RED might be a flake, and a GREEN
might be luck. Establish which it is before shipping.

Investigate systematically rather than by re-running and hoping:

- **Fixture isolation** — do cases share a `TMPDIR` path, a fixture name, or a file that a parallel
  or repeated run could collide on? Run the harness many times in parallel and under load.
- **State leaking between cases** — an exported variable, a `cd`, a stale file from an earlier case.
- **Environment leakage into the guard** — this change's own second bug was an outer variable
  reaching inside a function. Check whether the harness can leak a variable the guard reads,
  `GIT_*` and `IFS` included.
- **A stale binary or a cached copy** — the guard being invoked from somewhere other than the tree
  under test.

**Either reproduce it and fix the cause, or establish positively that it cannot happen** and say
what rules it out. **"Could not reproduce" alone is not an answer** — that is where this started.

Record the conclusion in `design.md`.

**Files:** `scripts/test-check-visual-trigger.sh`,
`spectre/changes/kan-359-multi-glob-ui-paths-matches-nothing/design.md`
**Tests:** `scripts/test-check-visual-trigger.sh`
**Regression:** reverting this commit returns the harness to a state where a spurious failure
indistinguishable from a real regression has been observed and not explained.
**Baseline:** before=14 after=14 cases; this task explains or fixes, it does not add coverage
<!-- predicted: no new case unless the investigation finds a cause needing one -->
**Commit:** `fix(scripts): make the visual-trigger harness deterministic`
**Build:** green

## Part 3 — round 2's findings

- [x] 7. Correct the timing claim, measured the way it is described

`scripts/lib/trim-glob-element.sh`'s header says its numbers were "**Measured on the ORIGINAL
char-loop**". They were not. They were end-to-end guard runs with the padding at the **edges** of
the `ui paths` cell — where `visual-table-cells.awk`'s whole-cell trim strips it **before**
`trim_glob_element` ever runs. The claim describes one measurement and reports another.

**The coordinator produced those numbers and passed them on; the mistake is not the implementer's.**

Measured properly, with the padding on an **interior** element, which whole-cell trimming cannot
reach:

```text verified:/usr/bin/time -p against the pre-fix guard extracted from 8650918 and the current guard, this worktree
interior padding   pre-fix (char loop)   current (single pass)
 4,000 per side          0.97s                  —
 8,000 per side          1.96s                0.35s
20,000 per side            —                  1.80s
```

Two consequences, and both must reach the file:

- **The fix is worth far more than recorded** — roughly a fivefold improvement at 8,000, not the
  marginal gain the header implies.
- **The "not a denial of service" conclusion must be re-derived, not restated.** It was drawn from
  numbers that measured the wrong path. State what the current guard actually costs on hostile
  interior padding and let the conclusion follow from that. If it no longer holds, say so.

Correct or remove every number in that header that was obtained the edge-padded way, and say which
method produced each number that remains.

**Files:** `scripts/lib/trim-glob-element.sh`
**Tests:** none added — this corrects a recorded measurement; the behaviour is unchanged
**Regression:** reverting this commit restores a "verified" claim that describes a measurement
nobody took, used to justify a risk conclusion.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `docs(scripts): measure the trim the way the header says it was measured`
**Build:** green

- [x] 8. Make the validator harness actually exercise the shared function

Replacing `trim_glob_element` with an identity function leaves
`scripts/test-check-visual-verification.sh` **fully green**, while
`scripts/test-check-visual-trigger.sh` fails four cases. The validator harness does not exercise the
shared function at all.

Case 26's two elements sit at **both list boundaries**, so `visual-table-cells.awk`'s whole-cell trim
already reduces the cell to a bare `,` before the shared function runs — the same edge-versus-interior
distinction task 7 corrects.

A constructed fixture with an **interior** whitespace-and-backtick element shows a real false pass:
the real lib reports `VISUAL-INVALID` (exit 1); an identity `trim_glob_element` reports `VISUAL-OK`
(exit 0). The shipped suite does not catch it.

**This is a regression-net gap, not a production bug** — both guards call the same shared function on
the same data, so they cannot disagree today. But a future break in that function would be caught by
only one of its two harnesses.

Add a validator case whose element is interior, and **verify it by mutation**: replace the lib's
function with an identity and confirm **both** harnesses now fail.

**Files:** `scripts/test-check-visual-verification.sh`
**Tests:** `scripts/test-check-visual-verification.sh`
**Regression:** reverting this commit returns the validator harness to passing against a broken
shared function.
**Baseline:** before=26 after=27 cases in that harness
<!-- predicted: the harness's own case count, before and after task 8 -->
**Commit:** `test(scripts): exercise the shared trim from the validator harness`
**Build:** green

- [x] 9. The shared cell splitter is the real slow path

Task 7's re-measurement found the guard still costs seconds on hostile input **after**
`trim_glob_element` was made linear. The remaining cost is in the shared table parser.

```text verified:/usr/bin/time -p against scripts/check-visual-trigger.sh with 60,000 backticks on an interior list element, this worktree
120 KB config, 60,000 interior padding per side -> 15.44s
```

**`split_cells` accumulates one character at a time** (`cur = cur ch`), and each concatenation
copies the whole accumulated string — quadratic in cell length. **This is the same defect class this
change already fixed once**, in `trim_glob_element`'s char-at-a-time strip; record that, because the
pattern evidently recurs and the next reader should recognise it.

`scripts/lib/visual-table-cells.awk` is shared by **three** production guards —
`check-visual-trigger.sh`, `check-visual-verification.sh`, `resolve-visual-screenshots.sh` — so this
is one fix for all three.

**What forces the character loop is escape handling** — `\|` and `\\` must not split a cell. Any
replacement keeps that semantics exactly; the existing harness cases pin it and must pass
**unchanged**. If a case needs editing, behaviour moved: stop and report.

**Measure before and after with the same method task 7 established** — end-to-end, interior padding,
`/usr/bin/time -p` — and record both. Do not reuse task 7's numbers; they measured a different
function.

**This is a pre-existing cost, not one this change introduced**, and the operator chose to fix it
here rather than defer it. Say so where it is recorded.

**Files:** `scripts/lib/visual-table-cells.awk`, `scripts/test-check-visual-trigger.sh`
**Tests:** `scripts/test-check-visual-trigger.sh` — a case pinning the parser's cost shape, plus the
existing escaped-pipe cases unchanged
**Regression:** reverting this commit restores a quadratic cell splitter shared by three guards, on
input any pull request can edit.
**Baseline:** before=15.44s after=0.05s at 60,000 interior padding per side
<!-- measured: /usr/bin/time -p against check-visual-trigger.sh, same fixture, after task 9 -->
**Commit:** `perf(scripts): split table cells without quadratic concatenation`
**Build:** green

- [x] 10. Pin the escape semantics no test protects

`split_cells` must not split a cell on `\|`, and must treat `\\` as a literal backslash. **No
committed test in any of the three harnesses pins either** — confirmed twice, by the implementer of
task 9 and independently by a panel slot.

Task 9 rewrote that function and verified escape behaviour out of band: a hand-built matrix plus
randomized fuzz lines, diffing old against new byte-for-byte. That was diligent and it is **gone** —
it lives in a transcript, not in the repository. A future change to this file would be caught by
nothing.

**`scripts/lib/visual-table-cells.awk` is shared by three production guards**, so a silent escape
regression reaches all of them.

Add committed cases pinning, at minimum: a cell containing `\|` that must **not** split; `\\`
treated as one literal backslash; a trailing lone backslash at end of line; and an escaped backslash
immediately followed by a real delimiter (`\\|`), which must split.

**Verify each case RED against a deliberately broken splitter** — mutate the escape handling and
confirm the new cases fail. A case that passes against broken escape handling pins nothing.

Put them in whichever harness gives the shared lib real coverage, and say why there.

**Files:** `scripts/test-check-visual-trigger.sh`
**Tests:** `scripts/test-check-visual-trigger.sh`
**Regression:** reverting this commit returns a shared lib behind three guards to having its escape
semantics protected only by a fuzz run that no longer exists.
**Baseline:** before=26 after=31 cases in that harness
<!-- measured: highest case number in scripts/test-check-visual-trigger.sh before task 10 and at HEAD -->

**Each case is RED under at least one broken splitter, and case 30 needed a mutation of its own.**
Forcing backslash detection off reds 27 and 28; removing the `nc == "|" || nc == "\\"`
discrimination reds 29 alone; making a `\\` pair also swallow a following `|` reds 30 alone. The
implementing agent's first report and commit message credited the first mutation with catching case
30 as well, which is false — it stays green there, since the real `|` splits either way. Case 30 is
not vacuous, but establishing that took a fourth mutation, written into the amended commit message.
This is the same failure the rest of this change kept hitting: a verification claim believed on the
shape of the code rather than run.

**Round 4 added case 31**, closing the one mutation the panel found that all three harnesses stayed
green under — the lone-backslash branch advancing one byte too far. Case 29 covered that branch but
sat at a cell boundary `trimcell` normalises away, so it could not observe the arithmetic. See
**Round 4 — the panel found a mutation nothing caught** in `design.md`.
**Commit:** `test(scripts): pin the table splitter's escape semantics`
**Build:** green

