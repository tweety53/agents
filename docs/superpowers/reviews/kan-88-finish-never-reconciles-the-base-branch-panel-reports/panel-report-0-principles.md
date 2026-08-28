VERDICT: clean

Summary: diff adds `scripts/lib/resolve-remote-base.sh` (shared resolution, avoiding duplication between two guards — good DRY/SRP), a new `check-base-moved.sh` guard following the established verdict-line/exit-code protocol of its siblings, hardens `check-finish-preflight.sh`'s ref resolution, and wires both through `finish-contract.md` (canonical) and `integrate.md` (execution-specific prose, citing rather than restating verdict semantics — matching the pre-existing unfinished-work-gate pattern exactly). All three principle groups (Structure, Simplicity & state, Robustness & ops) apply and were checked; no violations found. No project standards beyond lint-fix-priority resolved (`CLAUDE.md`/`AGENTS.md` "Project-specific standards" section is the unfilled template) — no new suppressions, no lint-config weakening, no layer violations.

Verification performed (reproduced, not just read):
- `scripts/test-check-base-moved.sh` — all 26 assertions pass.
- `scripts/test-check-finish-preflight.sh` — passed 11/11 consecutive runs. One earlier run in this session showed a transient failure on a pre-existing (non-diff-touched) case; irreproducible across 11 subsequent attempts and the same case passes when the guard is called directly, so no reproducible defect to report, and it's outside my Principles scope (defect findings belong to another panel slot) regardless.
- `check-references.sh`, `check-vocabulary.sh`, `check-guard-symlinks.sh`, `check-contract-budget.sh`, `check-stage-mark-calls.sh`, `check-installed-citations.sh`, `check-markdown-integrity.py`, `test-check-guard-symlinks.sh` — all clean, matching task 3's required six checks plus task 2's symlink guard.

No Critical, Important, or Minor findings.
