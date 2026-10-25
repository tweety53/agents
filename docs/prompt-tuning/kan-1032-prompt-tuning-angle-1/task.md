- [x] 3. Guard: admit a cases-only self-review branch in `check-loop-landing`
  - [x] **Step 1: In `stats/internal/guard/looplanding.go`, before the flowd row checks: for a `self-review-*` branch, list `git diff --name-only origin/<default-branch>...<branch>`; when non-empty and every path begins `docs/prompt-tuning/`, print `LOOP-LANDING-OK (prompt-tuning cases only)` and return 0. Extend the header comment of the function and of `scripts/check-loop-landing.sh` with this admission and its reason (design.md §4).**
  - [x] **Step 2: Tests in `stats/internal/guard/check_loop_landing_test.go`, beside the existing harness: `TestLoopLandingAdmitsPromptTuningOnly` — a `self-review-x` branch whose only commit adds `docs/prompt-tuning/x-1/case.json` exits 0 with no flowd rows; the same branch with one more commit touching `skills/flow/implement.md` is still refused for its missing rows. Verify: `cd stats && gofmt -l . && go vet ./internal/guard/... && go test ./internal/guard/... -run 'TestLoopLandingAdmitsPromptTuningOnly' -race -count=1`.**
**Build:** green
**Files:** `stats/internal/guard/looplanding.go`, `stats/internal/guard/check_loop_landing_test.go`, `scripts/check-loop-landing.sh`
**Tests:** `TestLoopLandingAdmitsPromptTuningOnly`
**Regression:** reverting the commit drops the admission: `TestLoopLandingAdmitsPromptTuningOnly` fails with `LOOP-LANDING-REFUSED` on the cases-only branch.
**Baseline:** before=2 after=3 <!-- measured: grep -c '^func Test' stats/internal/guard/check_loop_landing_test.go @ 43b6caf7 → 2 -->
**Commit:** feat(guard): land a self-review branch carrying only prompt-tuning cases
**After:** none

**Decision:** cases-ride-fix-branch

