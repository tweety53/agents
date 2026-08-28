**Bugbot — fix round 1, Full mode — verdict: findings**

**Part 1 — regression fixes confirmed (agree with dispatcher):**
1. `scripts/check-finish-preflight.sh:117-124` — collapsed `-eq 1`/`-ne 0` mutation -> 3 assertions fail in case 8c (`test-check-finish-preflight.sh` case "merge-base failure"). Caught.
2. `scripts/check-base-moved.sh:85-88` — dropped `COUNT=` error guard -> case 9b fails (rc=128 not 2, no message, `test-check-base-moved.sh`). Caught.
3. `scripts/check-base-moved.sh:137` — `-gt 10` -> `-ge 10` -> case 6b fails ("exactly 10" gets a spurious `(+0 more)` tail). Caught.

All three restored; worktree confirmed clean via `git status --porcelain`.

**Part 2 — surviving mutants found (unprotected):**

- MEDIUM `scripts/check-finish-preflight.sh:77-80` — dropping the `|| { …; exit 2; }` guard on `HEAD_SHA=` (unresolvable/unborn `HEAD`) is not caught by any test in `scripts/test-check-finish-preflight.sh`; the suite never exercises a worktree where `HEAD` fails to resolve.
  Reproducer: `none — reproduced via direct file mutation and full suite run this round; edit line 77-80 to drop the error-guard braces, run bash scripts/test-check-finish-preflight.sh, observe 0 failures, then restore.`

- MEDIUM `scripts/check-base-moved.sh:95-98` — dropping the `|| { …; exit 2; }` guard on `MOVED_RAW=` (the base-side `diff --name-only` capture) is not caught by `scripts/test-check-base-moved.sh`. Only the `COUNT=`/`rev-list` guard (case 9b) is shim-tested; the sibling guards on `MOVED_RAW`, `COMMITTED_RAW`, `STAGED_RAW`, `UNSTAGED_RAW` (lines 95, 105, 110, 115) share the exact same shape and are equally unprotected — I confirmed `MOVED_RAW` concretely; the other three are the same pattern and very likely share the gap (not individually reproduced, to stay in budget).
  Reproducer: `none — reproduced via direct file mutation and full suite run this round; edit line 95-98 to drop the error-guard braces, run bash scripts/test-check-base-moved.sh, observe 0 failures, then restore.`

- LOW `scripts/lib/resolve-remote-base.sh:35` — dropping `--end-of-options` from the `rev-parse --verify` call is not caught by either suite (both suites' `base_ref` values never begin with `-`). The file's own header states this flag exists specifically so a ref beginning with `-` is read as a ref, but nothing exercises that case.
  Reproducer: `none — reproduced via direct file mutation and full suite run this round; remove --end-of-options from line 35, run both test-check-finish-preflight.sh and test-check-base-moved.sh, observe 0 failures, then restore.`

**No issue found** with the three new tests' own validity — their shims scope narrowly to the intended subcommand (`status`/`merge-base`/`rev-list`) and pass every other call through to real `git`; sandboxes are per-case `mktemp -d`, tracked and cleaned via the `REPOS` array/EXIT trap; assertions on verdict prefix + reason text are specific enough that none pass for a substring-coincidence reason.

Contract prose (`skills/flow-contracts/finish-contract.md`) matches the scripts' actual exit codes and verdict vocabulary — no drift found there.

VERDICT: findings
