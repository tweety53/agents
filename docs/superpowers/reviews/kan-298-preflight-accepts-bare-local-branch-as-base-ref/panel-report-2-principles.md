Only difference is the trailing newline, which `$(...)` command substitution strips in both branches of the actual test comparison — confirmed correct match, and the full harness passes (`check-base-moved: all cases pass`). Both harnesses ran green in repeated trials; a single flake in `test-check-finish-preflight.sh` (1 of 36 runs) did not reproduce when isolating the exact string comparison in 30 direct iterations, so I'm not attributing it to this diff.

### Summary
Diff extracts the two harnesses' identical `EXPECTED_USAGE` literals into `scripts/lib/base-ref-usage.sh`'s `base_ref_usage_message`, closing F3 from round 1. Structure and Simplicity & state principles bear on it (DRY/WET, KISS, SRQ); project standards resolved (CLAUDE.md/AGENTS.md) but name no rule specific to `scripts/lib/` placement, so the hard-invariants section is empty for this delta. The extraction is sound and complies.

### Issues

#### Critical (Must Fix)
None.

#### Important (Should Fix)
None.

#### Minor (Nice to Have)
None — no new finding. The extraction is well-executed: right home (`scripts/lib/`, matching `test-git-shim.sh`'s established precedent for test-only helpers), right interface (one pure function, one parameter, stdout output — no config/flags/mode enum), and the header correctly preserves the independent-golden-value discipline (`source-grep-over-behaviour-test`, `usage-message-is-the-only-site`) by stating the text must never be derived by sourcing either guard. It correctly stops short of the heavier `test-lib-coverage.sh`/`test-lib-test-git-shim.sh` meta-harness pattern — that pattern exists for helpers with branching pass/fail logic (KAN-88 F17); `base_ref_usage_message` has no branches, and both existing call sites already exercise it end-to-end against real guard output for both guard names, so a dedicated self-test would add process overhead with no coverage gain.

### Assessment
**Principles-compliant?** Yes
**Reasoning:** The extraction reduced accidental duplication (DRY) without over-abstracting (KISS/WET stayed respected — two real callers justify one small function, no generalization beyond what's used), and the golden-value independence the design explicitly required is preserved.
