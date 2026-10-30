# Finding

- **Severity:** Important (round 0, slot failure-modes, F1)
- **Location:** stats/cmd/flow/request.go:222
- **Reviewer's sentence:** "The landed-end gate treats every `ErrUnavailable` from `ListStageRuns` as \"store unreachable\" and passes with a ⚠ line. But `ListStageRuns` (stats/internal/client/client.go:623) also wraps a reachable daemon's non-200 answer in `ErrUnavailable`"
