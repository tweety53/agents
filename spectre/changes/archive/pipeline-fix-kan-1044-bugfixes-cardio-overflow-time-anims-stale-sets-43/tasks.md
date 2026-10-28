# pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-43

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** every process a flow run starts, load reproductions included, runs under a strict memory cap.

**Architecture:** `scripts/run-capped.sh` + central wiring in `run-verify-list`; loop 41 wording rewritten.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject, no trailer; never write `spectre/changes/`.
- No uncapped load reproduction, multi-worker suite repeat or concurrent app/browser boot while doing this work.
- Go test runs use Bash `timeout: 600000`.
- No new inline suppressions; no weakening of lint config.

- [x] 1. Add the memory-capping wrapper
**Build:** green
**Files:** `scripts/run-capped.sh`, `scripts/test-run-capped.sh`, `skills/flow-contracts/project-configuration.md`
**Tests:** `scripts/test-run-capped.sh`
  - [x] **Step 1 — test first.** `scripts/test-run-capped.sh`: (a) `run-capped.sh -m 64 -- python3 -c
    'b=bytearray(300*1024*1024); import time; time.sleep(10)'` exits 137 within 6 s and stderr contains
    `CAPPED:`; (b) a grandchild allocator (`bash -c 'python3 -c ... & wait'`) is killed too (no surviving
    pid of the group); (c) `run-capped.sh -m 64 -- bash -c 'exit 3'` exits 3; (d) with no `-m` in a temp dir
    whose `.flow/project.md` has `## resource cap` body `128`, a 300 MB child is capped. Fails while the
    script is absent.
  - [x] **Step 2 — `scripts/run-capped.sh`** per design.md Decisions (process group, 1 s `ps -o rss= -g`
    poll summed, TERM then KILL, `CAPPED:` line, exit 137, child status otherwise, INT/TERM trap). Add the
    `# ponytail:` poll-granularity comment. Register the test where `run-guard-tests.sh` picks up
    `test-*.sh` if it does not already.
  - [x] **Step 3 — `skills/flow-contracts/project-configuration.md`**: add a `## resource cap` row after
    `## lint`: optional; first integer of the body is the memory cap in MB for every process a flow run
    starts, enforced by `<agents repo>/scripts/run-capped.sh`; absent means 4096.
  - **Commit:** `feat(scripts): add run-capped.sh memory-capping wrapper`
  - **After:** `bash scripts/test-run-capped.sh` passes.

- [x] 2. Route run-verify-list commands through the cap
**Build:** green
**Files:** `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`
**Tests:** `TestRunVerifyListCapsMemory`
  - [x] **Step 1 — test first.** In the runverifylist Go test, a case whose `## test` fence runs a
    command allocating above a `## resource cap` of 64 asserts the list fails and the output carries
    `CAPPED:`. Fails before Step 2.
  - [x] **Step 2 — `stats/internal/guard/runverifylist.go`**: every `{"bash","-c",c}` argv (plain and
    behind `flow suite record ... --`) becomes `{filepath.Join(scriptDir,"run-capped.sh"),"--","bash","-c",c}`.
  - **Commit:** `feat(guard): run every verify-list command under run-capped.sh`
  - **After:** `cd stats && go test ./internal/guard/ -run 'RunVerifyList'` passes.

- [x] 3. Allow capped load reproductions; cap every started process
**Build:** green
**Files:** `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`, `skills/flow/implement.md`, `rules/fix-determinism-at-the-source.mdc`, `skills/flow/verify-fix-loop.md`, `skills/flow/primary-reviewer-prompt.md`, `skills/flow/unfinished-work-gate.md`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestCheckDispatchParagraphs`
  - [x] **Step 1 — guard first.** `dispatchparagraphs.go`: `proofruns` phrases become
    `"run-capped.sh"`, `"bounded worker count"`, `"straight to code"`,
    `"shows the check failing with the mechanism removed"`; `noloadrepro` (rename key to `cappedrepro`,
    sites updated) phrase becomes `"under run-capped.sh"`. In `check_dispatch_paragraphs_test.go` add a
    case (next free number) whose `implement.md` fixture carries loop 41's PROOF RUNS text ("never a load
    reproduction … at most 5 runs") and expects exit 1 naming `run-capped.sh`; update fixtures carrying
    loop 41 text so passing cases pass.
  - [x] **Step 2 — `skills/flow/implement.md` PROOF RUNS block**, replace with:
    > **PROOF RUNS:** Every process you start — build, test suite, dev stack, probe — runs under
    > `<agents repo>/scripts/run-capped.sh --`, which kills it past the project's `## resource cap`
    > (default 4096 MB). A flake fix is proven by forcing its cause, preferably in one isolated run — a
    > delay or fault injected at the seam, the test alone. A load reproduction (backend or frontend) is
    > allowed only under run-capped.sh with an explicit bounded worker count sized under the cap: Gradle
    > `--max-workers`/`maxParallelForks` and `-Xmx`, Playwright `--workers`, vitest `--maxWorkers`. If the
    > flake does not reproduce within the cap, go straight to code: no further reproduction loops; fix it
    > on deeper analysis of the code and the failing run's saved artifacts, and the commit body says so.
    > Your report states the forced condition, the cap, the worker count and the run count. A determinism
    > fix that was reproduced still shows the check failing with the mechanism removed
    > (`~/.claude/rules/fix-determinism-at-the-source.md`); only one fixed straight to code waives it, stated.
    Also at ~line 620 append to the sentence handing over `## run`: "start each under
    `<agents repo>/scripts/run-capped.sh --`".
  - [x] **Step 3 — `rules/fix-determinism-at-the-source.mdc`** § "Read the flake before touching
    anything", the paragraph beginning "Make it fail by forcing that cause": replace "never by a load
    reproduction: no parallel load, … (kan-1044)." with: a load reproduction is allowed only under
    `<agents repo>/scripts/run-capped.sh` with a bounded worker count sized under the project's
    `## resource cap`, so the operator's machine is never overloaded (kan-1044); and replace the last
    sentence with: where the cause does not show within the cap, go straight to code — fix it on deeper
    code analysis, say so in the commit body; that, and only that, waives the mechanism-removed run.
  - [x] **Step 4 — `skills/flow/verify-fix-loop.md` item 2** (~line 106): "never by a load reproduction
    (PROOF RUNS, …)" → "a load reproduction only under run-capped.sh (PROOF RUNS, …)"; the sentence on a
    cause that cannot be forced reads "one that does not reproduce within the cap goes straight to code".
  - [x] **Step 5 — `skills/flow/primary-reviewer-prompt.md`** (~line 57): "one proven by a load
    reproduction" → "one proven by a load reproduction not run under run-capped.sh with a bounded worker
    count". **`skills/flow/unfinished-work-gate.md`** (~line 24): "never a load re-run" → "a load re-run
    only under run-capped.sh".
  - [x] **Step 6 — `skills/flow/verify-and-handoff.md`** bullet "Every start command comes from
    `<project>/.flow/project.md`'s `## run`" (~line 361): add "each started as
    `<agents repo>/scripts/run-capped.sh -- <command>`".
  - **Commit:** `fix(flow): allow load reproductions only under the run-capped.sh memory cap`
  - **After:** `cd stats && go test ./internal/guard/ -run DispatchParagraphs` passes and the real-tree
    dispatch-paragraphs check passes.
