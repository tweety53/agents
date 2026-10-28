## 1. Guard accepts a cache-skipped verify row as green

- [ ] 1.1 In `stats/internal/guard/check_verify_green_test.go`, add case `{"a cache-skipped verify is green", "[" + row("verify-2-gymie", "mf-tok", "skipped") + "," + row("verify-2-gymie-frontend", "mf-tok", "skipped") + "]", nil, 0, "VERIFY-GREEN"}`; run `cd stats && go test ./internal/guard -run TestCheckVerifyGreen` and see it fail (exit 1, `VERIFY-NOT-GREEN`).
- [ ] 1.2 In `stats/internal/guard/verifygreen.go`, change `case "completed":` to `case "completed", "skipped":`, and the success line to `VERIFY-GREEN: %s — %d inline verify row(s) green`.
- [ ] 1.3 Update existing test needles that match `inline verify row(s) completed` to `inline verify row(s) green`.
- [ ] 1.4 In `scripts/check-verify-green.sh` header: verdict line `VERIFY-GREEN:     <change> — <n> inline verify row(s) green`; exit-0 sentence `Exit 0 every judged row carries outcome `completed` or `skipped` (a whole-list cache skip run-verify-list.sh records only when the list's last green run on a clean tree covered HEAD) and at least one exists;`.
- [ ] 1.5 `cd stats && go test ./internal/guard -run TestCheckVerifyGreen` green; `gofmt -l internal/guard` empty.

**Commit:** `fix(guard): check-verify-green accepts a cache-skipped verify row as green`
