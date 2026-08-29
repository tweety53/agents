Confirmed. Both tests actually pass against the real guards (ran them). Findings below.

### Summary
Delta adds one behaviour test per harness (`test-check-base-moved.sh`, `test-check-finish-preflight.sh`) that invokes each guard with no arguments and asserts stdout is empty and stderr equals a hardcoded `EXPECTED_USAGE` literal. Ran both test files against the real guards — both pass. Applies Structure (n/a — test-only diff, no production code touched), Simplicity & state (DRY/SSOT, WET), and Robustness & ops (Testing principles). No project-specific standards are declared (`CLAUDE.md`/`AGENTS.md` "Project-specific standards" section is the undeclared template), so the Hard Invariants section is empty per instructions.

**The turning question — ordinary fixture or forbidden second copy.** This is an ordinary test fixture, not the violation `usage-message-is-the-only-site` guards against. That decision governs where the *qualification is written in production code* (the guard's own runtime message) — it says nothing about test oracles, and its own header note ("neither guard's header comment block is touched") is about the guard source, which this delta doesn't touch. A test that hardcodes expected output is the standard form under the Testing principle ("assert observable behavior") and DRY's own exception (WET: "the test must state what it expects"): the literal is checked against the real guard every run, so drift is loud (a failing test), not silent — the exact failure mode SSOT exists to prevent. It also tests something the pre-existing source-grep case (`grep -qF 'prefers refs/remotes/origin/<base-ref>'`, present in both files already) cannot: exact stream separation (stdout empty, stderr *exactly* the message, no appended second diagnostic) — a distinct, non-redundant assertion, verified by actually running both guards above.

### Issues

#### Critical (Must Fix)
None.

#### Important (Should Fix)
None.

#### Minor (Nice to Have)

- `scripts/test-check-base-moved.sh:504-511`, `scripts/test-check-finish-preflight.sh:339-346` — WET: the two `EXPECTED_USAGE` literals are identical apart from the guard name, mirroring the guards' own near-duplication (already an accepted design tradeoff, "both-guards-share-the-defect"). Not worth extracting for two call sites under KISS — noted only as a maintenance cost: a future wording edit now touches two test literals plus two source-grep phrases plus two guards (four sites, all mechanically necessary, none silently divergent).
  Reproducer: `none — a style observation, not a runnable check`

### Assessment
**Principles-compliant?** Yes
**Reasoning:** The hardcoded `EXPECTED_USAGE` strings are ordinary, continuously-verified test fixtures that test observable behavior the existing source-grep case cannot (stream separation) — not the second driftable copy `usage-message-is-the-only-site` forbids, which is scoped to the guards' own runtime source.
