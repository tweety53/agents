Critical: 0
Important: 0
Minor: 0
Unread: none

# Panel round 4 — principles — gymie-playwright section

Coverage: same as the primary pass — text hunks of the gymie-playwright section read; PNG/zip baselines not inspected.

### Derived-state check
- No stored summary state written by this section. `FRAME_HEIGHT` entries are literal per-frame constants (1688/2-28), not derivable from other repo state at runtime. Nothing to derive.

### Summary
Touches fidelity specs, fixtures (`typeOver`, `waitForPickedDay`), frame heights, a smoke-line test and baselines. Principles applied: DRY (searched repo for an existing type-over helper — `grep -rn "typeLine\|typeOver" fidelity tests fixtures.ts`: `typeLine` in fidelity/cardio.ts is cardio-specific, so `typeOver` is a justified second helper and its doc cites the shared mechanism), KISS, deterministic-clock rule (typeOver pins the clock rather than adding sleeps/retries). Project standards (CLAUDE.md, CONTRIBUTING.md) resolved; their absolute rules (no new suppressions, no lint weakening, Kotlin layering) — grep for `eslint-disable|@ts-ignore|@ts-expect` in added lines: none. Kotlin layering does not apply to this TS repo.

### Issues
#### Critical (Must Fix)
None.
#### Important (Should Fix)
None.
#### Minor (Nice to Have)
None.

### Assessment
**Principles-compliant?** Yes
**Reasoning:** No suppressions, no tolerance widening for flakes; new helpers have no existing equivalent.
