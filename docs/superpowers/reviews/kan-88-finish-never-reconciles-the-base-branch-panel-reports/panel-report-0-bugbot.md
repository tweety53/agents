Worktree confirmed clean — only pre-existing untracked spectre/design files remain (not touched by me), all three mutated scripts match originals exactly.

## Findings

- **Minor**, `scripts/check-base-moved.sh:98` (COUNT capture) — no test exercises a failing `git rev-list`/`git diff` invocation in this guard (unlike `check-finish-preflight.sh` case 8b's shim test for `git status`). Removing the `|| { …; exit 2; }` guard on the `COUNT=` assignment survives the full suite untouched — the guard is correctly written in the code, but nothing proves it stays that way.
  Reproducer: `none — coverage gap only provable by a multi-step mutate/run/revert sequence (performed and reverted during this review; see mutation log)`

- **Minor**, `scripts/check-finish-preflight.sh:147` (`ANCESTOR_RC -ne 0` guard) — same class: collapsing the `-eq 1` / `-ne 0` two-step into "any nonzero → RUN1" (i.e. deleting the explicit exit-2 branch for a `merge-base` failure that isn't "not an ancestor") survives the whole suite. Code is currently correct — this is exactly the "git failure read as RUN1" shape the brief calls out as dangerous — but no regression test protects it.
  Reproducer: `none — coverage gap only provable by a multi-step mutate/run/revert sequence (performed and reverted during this review; see mutation log)`

- **Minor**, `scripts/check-base-moved.sh:137` (`TOTAL -gt 10` cap) — the exactly-10-overlap boundary is untested; changing `-gt 10` to `-ge 10` survives the suite because case 6 only exercises 11 overlaps. Verified by hand that the real code is correct at exactly 10 (all 10 shown, no `(+N more)` tail), so this is a test-gap, not a live bug.
  Reproducer: `none — boundary untested; verified correct by hand outside the test suite, not by a suite-catchable single command`

## Mutations performed (all reverted, worktree left clean)

1. `check-finish-preflight.sh`: reordered signal (b) after base-ref resolution -> **caught** (case "zero-commit branch, unresolvable base ref" failed as expected).
2. `check-finish-preflight.sh`: collapsed ancestor-RC handling to treat any nonzero as RUN1 -> **not caught** (finding above).
3. `check-base-moved.sh`: dropped `COUNT` failure guard -> **not caught** (finding above).
4. `check-base-moved.sh`: `TOTAL -gt 10` -> `-ge 10` (off-by-one at cap boundary) -> **not caught** (finding above); manually verified real code correct at N=10.
5. `resolve-remote-base.sh`: `refs/remotes/origin/${base_ref}` -> `origin/${base_ref}` (shortened ref form) -> **not caught**, but confirmed behaviorally equivalent under git's own ref disambiguation (not a real defect, discounted).
6. `resolve-remote-base.sh`: swapped the if/else branches (rewrite when it should pass through, pass through when it should rewrite) -> **caught hard** (13 failures across both suites).
7. Manually verified overlap detection against filenames containing spaces and a leading `-` (`/tmp/spacetest`) — correct, no defect.

No functional defect found in the shipped behavior of `resolve-remote-base.sh`, `check-finish-preflight.sh`, or `check-base-moved.sh`. All three findings are test-coverage gaps around already-correct defensive code, not live bugs.

VERDICT: findings
