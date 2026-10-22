# kan-919 — live verification

Before figures measured on the main checkout at the merge base `b1b0d601` (the bash guards); after figures on this worktree's HEAD (the Go shims).

| Check | Before (b1b0d601) | After (HEAD) |
|---|---|---|
| `scripts/check-spec-reach.sh .` | `Spec reach: not configured`, exit 0 | `Spec reach: not configured`, exit 0 |
| `make -s check-uitest-overrides` / `scripts/check-uitest-overrides.sh` | `UITEST-OVERRIDES-OK: every UITEST_* variable is override`, exit 0 | same line, exit 0 |
| `scripts/check-dev-stack-fresh.sh .` (no stack on 4174) | fingerprint line, build output, `stack-stale: …`, exit 1 (design.md `## Live check`) | fingerprint line, build output, `stack-stale: …`, exit 1 |

Byte-for-byte stdout/stderr/exit comparisons of old bash against the new shims on fixtures: identical for all three guards, except check-spec-reach's orphan ordering, identical under `LC_ALL=C` only (tasks.md task 2 Correction).

Result: matched — no failure look (no `flow-guard: unknown guard`, no exit 2 from a shim).
