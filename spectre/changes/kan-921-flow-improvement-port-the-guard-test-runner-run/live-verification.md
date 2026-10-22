# Live check — scripts/run-guard-tests.sh on the real tree

| | Runner | Summary line | Exit |
|---|---|---|---|
| Before (detached worktree at `2a3a0297`) | bash | `19 harnesses, 19 passed, 0 failed, 202s wall` | 0 |
| After (change branch) | Go, via the shim | `16 harnesses, 16 passed, 0 failed, 155s wall` | 0 |

Matched: after = before − the three deleted harnesses (`test-lib-parallel.sh`, `test-run-guard-tests.sh`, `test-lib-coverage.sh`), `0 failed`, exit 0, same summary shape. No failure look (no shim build error, no `no test-*.sh harnesses found`, no unchanged count).
