# pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-44

## Why

`run-verify-list.sh` closes a verify row `-outcome skipped` when every test command is cache-skipped because its `# inputs:` (and `.flow/project.md`) are unchanged since the list's own last green run on a clean tree (`skills/flow/verify-and-handoff.md`, the cache-skip paragraphs). `check-verify-green.sh` (`stats/internal/guard/verifygreen.go`) exits 0 only when every judged row is `completed`, so it prints `VERIFY-NOT-GREEN: … verify-2-gymie (skipped), …` and `flow.visual-verify` step 3 refuses to dispatch a verifier after a verify that is green by cache. Two guards contradict (observed 2026-10-28, kan-1044).

## What changes

- `check-verify-green` judges a latest row with outcome `skipped` as green, like `completed`. A `skipped` row is only ever recorded when the list's last green run covers HEAD, so it carries the same evidence.
- The guard's header in `scripts/check-verify-green.sh` names both outcomes as green.
- One test case in `stats/internal/guard/check_verify_green_test.go` that fails without the fix.

Blast radius: 4 live files cite the guard (`scripts/check-verify-green.sh`, `stats/internal/guard/verifygreen.go`, `stats/internal/guard/check_verify_green_test.go`, `skills/flow/visual-verify.md`); archived change folders are history and are not touched. Not `big`.
