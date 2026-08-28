**Mutations performed (all restored, verified via `git status --porcelain`):**
1. `scripts/lib/test-git-shim.sh` — removed `.fired` sentinel write -> caught (both harnesses exit 1, 3 cases each)
2. `scripts/lib/test-git-shim.sh` — `assert_shim_fired` made to unconditionally `pass` -> **not caught** (both harnesses exit 0)
3. `scripts/check-base-moved.sh` — cap boundary `-gt 10` -> `-gt 11` -> caught (exit 1, 8 failures)
4. `scripts/check-base-moved.sh` — `COUNT = "0"` -> `COUNT = "1"` (CLEAR/MOVED branch) -> caught (exit 1, 23 failures)
5. `scripts/lib/resolve-remote-base.sh` — inverted the `rev-parse` condition -> caught in both harnesses
6. `scripts/check-base-moved.sh` — `comm -12` -> `comm -3` (intersection) -> caught (exit 1, 17 failures)
7. `scripts/check-finish-preflight.sh` — `ANCESTOR_RC -eq 1` -> `-eq 99` -> caught (exit 1)
8. `scripts/check-finish-preflight.sh` — `DIRTY != "0"` -> `= "0"` -> caught (exit 1)
9. `scripts/check-base-moved.sh` — stripped `--end-of-options` from the COUNT `rev-list` call -> **not caught** by either harness or by a live real-repo reproduction with a dash-prefixed ref

**Findings**

- Minor — `scripts/lib/test-git-shim.sh:95` (`assert_shim_fired`): the sentinel *assertion* itself is untested — vacuously passing it produces no failure anywhere. Inherent to test-helper code (nothing tests the tester, same as `pass`/`fail` themselves); flagged per the brief's mutation protocol, not a shippable defect.
  Reproducer: `none — meta-level (the assertion helper has no test of its own by design; verified by inline python mutation, not reproducible as a plain command)`

- Minor — `scripts/lib/test-git-shim.sh:45` (`TEST_GIT_SHIM_REAL_GIT`): sourcing the library twice while a shim sits on `PATH` recaptures the shim as "real git" instead of the genuine binary. Confirmed by direct repro (double-`source` with a shim prefixed onto `PATH` between the two). Not currently reachable: both `test-check-base-moved.sh` and `test-check-finish-preflight.sh` source it exactly once, at the top, before touching `PATH` — so this is latent fragility, not a live bug. The header's "always runs while PATH still names only the real git" claim holds only under single-sourcing, which it doesn't state.
  Reproducer: `none — requires an inline bash -c double-source repro (shown to reviewer), not expressible as a plain worktree-relative command`

- Info — `scripts/check-base-moved.sh:87` (`rev-list --count --end-of-options "${RECORDED_SHA}..${EFFECTIVE_REF}"`): the `--end-of-options` guard is unreachable/decorative on this call and its `diff --name-only` sibling — the single-argument range always starts with `RECORDED_SHA` (a resolved commit hash, never `-`-prefixed), so a dash-prefixed `EFFECTIVE_REF` can never reach the leading position. Confirmed with a real repo containing both an `origin/-weird`-style ref and a local-only `-onlylocal` branch: removing the flag from this call changes nothing. Not a vulnerability — no test exercises it, but nothing exploits its absence either. No action needed.
  Reproducer: `none — the unreachability was proven against a throwaway repo outside the worktree, not the worktree itself`

No contract/code drift found in `skills/flow-contracts/finish-contract.md` or `skills/flow/integrate.md` — verdict wording, exit contract, per-worktree/aggregated-ask rules, and the guard-symlink/`SKILL.md` listing all match the shipped script. `check-references.sh` and `check-guard-symlinks.sh` both pass clean.

VERDICT: findings
