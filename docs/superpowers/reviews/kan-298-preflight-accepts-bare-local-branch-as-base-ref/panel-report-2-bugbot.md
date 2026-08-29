Clean — only the pre-existing untracked directory remains, no commits made.

**Independence check (the critical one):** `base_ref_usage_message` is a hand-authored golden value, not derived from either guard. Confirmed by mutation (L1): mutating the lib's text made both harnesses fail. The extraction is **independent, not tautological**.

**Mutation table**

| # | Target | Mutation | Caught? |
|---|--------|----------|---------|
| M1 | check-finish-preflight.sh | reword usage text visibly | Yes (fail) |
| M2 | check-finish-preflight.sh | drop `>&2` from heredoc | Yes (fail — stdout non-empty + stderr mismatch) |
| M3 | check-finish-preflight.sh | delete `exit 2` after `EOF` | Yes (fail) |
| M4 | check-finish-preflight.sh | reword line, keep "prefers refs/remotes/origin/<base-ref>" phrase | Yes (fail) |
| M5 | check-finish-preflight.sh | indentation-only change in usage body | Yes (fail) |
| M6 | check-finish-preflight.sh | change guard name on synopsis line | Yes (fail) |
| B1 | check-base-moved.sh | drop `>&2` | Yes (fail) |
| B2 | check-base-moved.sh | delete `exit 2` | Yes (fail) |
| B3 | check-base-moved.sh | reword, keep key phrase | Yes (fail) |
| B4 | check-base-moved.sh | indentation-only | Yes (fail) |
| B5 | check-base-moved.sh | guard name on synopsis | Yes (fail) |
| L1 | lib/base-ref-usage.sh | mutate text inside `base_ref_usage_message` | Yes — **both** harnesses fail (proves independence) |
| L2 | test-check-finish-preflight.sh | swap guard-name arg to `"check-base-moved.sh"` | Yes (fail — harness correctly distinguishes own guard's message) |
| L3 | test-check-finish-preflight.sh | delete the `. "$SCRIPT_DIR/lib/base-ref-usage.sh"` source line | Loud failure: rc=127, `base_ref_usage_message: command not found` (no silent skip; `set -euo pipefail` does its job) |
| L4 | lib/base-ref-usage.sh | add trailing `\n` to final `printf` | No behavior change — harmless. Both `$(guard-heredoc)` and `$(base_ref_usage_message)` go through command substitution, which strips trailing newlines on both sides regardless. The omitted trailing newline is not load-bearing; nothing to fix. |

No surviving mutants. No new ordinary defects found by reading (file placement, permissions, and lint guards `check-vocabulary.sh` / `check-references.sh` all pass).

Worktree left clean — only pre-existing untracked `spectre/changes/...` directory, no commits, no index/HEAD/branch changes.

**Verdict: PASS**
