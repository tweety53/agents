# pipeline-fix-kan-1034-spectre-optional-bundled-planning-prompts-that-31

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

- [ ] 1. Guard: the peer state read in `check-task-records` follows `FLOW_RECORDS_ADDR`
  - [ ] **Step 1: In `stats/internal/guard/check_task_records_test.go`, `TestCheckTaskRecords`, after case 19, add case 20: build the cross-repo fixture exactly as case 17 does (link.md, task 1 `fix(peer): one`, a peer repo committing `fix(peer): one`), with a stub `flow` that prints the `{"worktrees":{...}}` JSON only when `"$FLOW_ADDR" = http://127.0.0.1:4173` and otherwise `exit 1`; run it with an `Env` whose `Getenv` returns `FLOW_ADDR=http://127.0.0.1:6783`, `FLOW_RECORDS_ADDR=http://127.0.0.1:4173` and the stub dir prepended to `PATH`, and whose `Environ` returns `os.Environ()` plus `FLOW_ADDR=http://127.0.0.1:6783` and `FLOW_RECORDS_ADDR=http://127.0.0.1:4173`. Want `rc == 0 && out == ""`, label `case 20 peer state read follows FLOW_RECORDS_ADDR`. Run `cd stats && go test ./internal/guard/ -run 'TestCheckTaskRecords' -count=1` and see case 20 fail.**
  <!-- measured: in /Users/tweety53/Projects/agents-worktrees/kan-1034-spectre-optional-bundled-planning-prompts-that @ 5ab163d5, `scripts/check-task-records.sh` exit 0; `FLOW_ADDR=http://127.0.0.1:6783 scripts/check-task-records.sh` exit 1, "task 2 ticked but no commit with subject 'feat(skills): vendor spectre-planning and document the two modes' reachable from HEAD" -->
  - [ ] **Step 2: In `stats/internal/guard/taskrecords.go`, `trPeerHeads`, before `cmd.Output()`: `cmd.Env = icChildEnv(env)`; if `a := env.Getenv("FLOW_RECORDS_ADDR"); a != ""` then `cmd.Env = append(cmd.Env, "FLOW_ADDR="+a)`. Update the function's doc comment with one sentence: the state record lives in the records store, so the read follows `FLOW_RECORDS_ADDR` when set.**
  - [ ] **Step 3: Verify: `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckTaskRecords' -count=1` — clean and passing.**
**Build:** green
**Files:** `stats/internal/guard/taskrecords.go`, `stats/internal/guard/check_task_records_test.go`
**Tests:** `Case 20`
**Regression:** reverting the commit lets the child inherit `FLOW_ADDR=http://127.0.0.1:6783`, the stub exits 1, no peer is read, and case 20 fails with the `no commit with subject 'fix(peer): one'` violation.
**Commit:** fix(guard): read a peer's state record from the records store
**After:** none

**Decision:** peer-state-read-follows-records-addr
