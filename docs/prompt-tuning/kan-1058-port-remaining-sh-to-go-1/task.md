- [x] 4. Port `live-store-copy` to Go
  - [x] **Step 1: Add `stats/internal/guard/livestorecopy.go` registering `live-store-copy`, carrying `scripts/live-store-copy.sh`'s behaviour and header contract unchanged (`up <name> [<port>]`, `restart <name>`, `down <name>`; the `flow_<name>_live` copy, the build of `flowd` from `FLOW_GUARD_REPO_ROOT`, the refused ports 4173 and 4174, the scratch change registration, `FLOW_LIVE_COPY_TRANSCRIPTS`, the `_`/`-` name alias, its exit codes), shelling out to `docker`, `make`, `go` and `curl` exactly where the bash does.**
  - [x] **Step 2: Replace `scripts/live-store-copy.sh` with a `flow_guard_exec live-store-copy 2 "live-store-copy:"` shim that exports `FLOW_GUARD_REPO_ROOT` (decision `ports-as-shims-first`). Add `stats/internal/guard/live_store_copy_test.go` with exactly one test function, `TestLiveStoreCopy`, porting every case of `scripts/test-live-store-copy.sh` as a subtest with the same `docker`/`make`/`go`/`curl` stubs on `PATH`. Delete `scripts/test-live-store-copy.sh`.**
  - [x] **Step 3: Verify: `cd stats && go test -race -count=1 -run 'TestLiveStoreCopy' ./internal/guard/ && gofmt -l . && go vet ./internal/guard/`, then `scripts/run-guard-tests.sh`.**
**Build:** green
**Files:** `stats/internal/guard/livestorecopy.go`, `stats/internal/guard/live_store_copy_test.go`, `scripts/live-store-copy.sh`, `scripts/test-live-store-copy.sh`
**Tests:** `TestLiveStoreCopy`
**Regression:** reverting the commit removes the Go guard: `TestLiveStoreCopy` vanishes and the bash script returns untested by the Go runner.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/live_store_copy_test.go 2>/dev/null | grep -cE '^func Test[A-Z]' || true @ branch spectre/kan-1058-port-remaining-sh-to-go -->
**Commit:** refactor(guard): port live-store-copy to Go
**After:** Task 1

**Decision:** ports-as-shims-first

