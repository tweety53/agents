## Findings

**MEDIUM — `scripts/lib/test-git-shim.sh:33`** — the shim's argument-matching (`[ "$a" = "$match_arg" ]`) is exact-equality, but no test exercises the boundary between exact match and substring match: broadening it to `[[ "$a" == *"$match_arg"* ]]` (a legitimate "matching too broad" defect per the brief) produces zero test failures across both suites, because no other argument in any guard invocation happens to contain a match_arg as a substring. The suite's ability to catch an over-broad shim is accidental (an artifact of current call shapes), not verified. If a future case's match_arg becomes a substring of another argument in the same invocation (e.g. a ref name containing "status"), the shim would silently fail the wrong call and the doc comment's warning at line 18-20 ("MUST be a more specific argument ... so the shim fails only the one call under test") would be unenforced.
Reproducer: broaden the matcher to a substring test, run both suites, observe all-pass with 0 failures, then restore.

## Mutations performed (all restored, confirmed via `git status --porcelain`)

1. `test-git-shim.sh`: exact-match -> substring match — **not caught** (finding above)
2. `test-git-shim.sh`: passthrough `exec` -> swallow non-matching calls — caught (both suites fail)
3. `test-git-shim.sh`: shim exit code 128 -> 0 — caught (7 + 9 failures)
4. `check-base-moved.sh`: `comm -12` intersection -> no intersection (treat all moved paths as overlap) — caught (case 2, documented in the file's own header)
5. `check-finish-preflight.sh`: ancestor-test `-eq 1` -> `-ge 1` — caught (case 8c)
6. `resolve-remote-base.sh`: swapped the two branches' outputs — caught massively (nearly every case in both suites)

No order-dependence or leakage found: repeated back-to-back runs of `test-check-finish-preflight.sh` produce identical output and exit 0, and leave no stray temp dirs under `TMPDIR`. Contract prose in `finish-contract.md` and `integrate.md` matches the scripts' verdicts and exit-code semantics exactly — no drift found.

VERDICT: findings

---

**Dispatcher note, added after the fact:** this file was written from the slot's report after fix
round 3 had already been dispatched, because the dispatcher failed to capture it verbatim at the
time the slot reported, as **Recording findings** requires. Fix round 3 therefore worked from
`final-review-panel.md`'s summary and the dispatcher's own independent confirmation of F14's mutant
rather than from this file, and said so in its report. The omission is the dispatcher's.
