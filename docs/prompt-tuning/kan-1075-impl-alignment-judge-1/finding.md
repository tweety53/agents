**Severity:** Minor
**Slot:** mutation (round 2)
**Location:** cmd/flow-eval/main_test.go:480

dropping cancel() is caught only by a hang: the test has no read deadline and would fail at go test's 10m timeout
