# self-review-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the eight non-`big` findings of `kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`'s
self-review pass — a mis-specified fix-round review, an unchecked merge-order record, two port and
artifact gaps in gymie's Playwright rows, a sampling fallback and an unbounded strip pile in the
verifier, an untraced flake, and a harvester that never saw a hook denial.

**Architecture:** prose in `skills/flow/verify-fix-loop.md` and `skills/flow/visual-verify-verifier.md`;
signal ten of `check-unfinished-work` (`stats/internal/guard/unfinishedwork.go`); a tolerant decode of
`toolUseResult` in `stats/internal/harvest/transcript.go`; gymie's `## visual verification` rows; a
load reproduction of cardio-R1f in gymie-playwright.

**Spec:** `design.md` beside this file; the brainstorm record is
`.superpowers/sdd/2026-10-20-self-review-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; Task 8 may end with no commit, as its steps say. Never write `spectre/changes/` in this agents worktree: this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push. No `Co-Authored-By` or other attribution trailer.
- Each task names the repository it runs in by its worktree; every `**Files:**` path is relative to that worktree's root. Tasks 1–5 run in this agents worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy` (branch `self-review-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`). Tasks 6 and 7 run in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`, Task 8 in the gymie-playwright worktree `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`; both are on branch `spectre/kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy`. Never touch a main checkout, and never create a worktree.
- Go: run `cd stats && gofmt -w .` before committing; `gofmt -l .` prints nothing and `go vet` stays clean. Open every Go test run with `make -s web-dist` (`stats/Makefile`). Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container, and never write to the live store: every test runs against fakes.
- The kan-995 dev stack is already running from the gymie worktree. Tasks 6–8 never start, stop or restart it; if it is down, Task 8 stops and reports that, never bringing it up itself.
- No new dependency in any repository.

## Review Focus

- Task 2: a single-repository change, and a change whose ranges are two worktrees of one repository, read nothing new; a `## Part of` satellite `link.md` is not mistaken for the merge-order record; a peer the `spectre/peers` file does not declare is reported, never a crash.
- Task 5: a user line whose `toolUseResult` is an object still yields its launch; a non-agent tool's error result still yields no denial.
- Task 6: an exported `GATEWAY_PORT` or `FRONTEND_PORT` still wins over the file read.
- Task 8: the `dadc7396` reversal is undone before the task ends, whatever the outcome.

- [x] 1. Re-review a verify fix round under panel keys, beside the always-on extra dispatch
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `git grep -n -F -e 'run one dispatch: `primary` alone,' -- skills/flow/verify-fix-loop.md` prints the one line of step 3 that drops the extra dispatch and names no key.
  - [x] **Step 2: Replace step 3's first paragraph** — from `3. **Re-review the delta.**` through `apply as written — a clean or Minor-only read leaves every slot it did not dispatch current.` — with exactly the block below, leaving the rest of the step (from `` `check-late-fix-trigger.sh` is not run``) unchanged (decision `fix-round-review-keys`):

  ```markdown verified:applied at plan time; check-verbatim-moves.sh printed its sentences, listed in verbatim-moves.txt
  3. **Re-review the delta.** Open `flow.review-panel` again and run `primary` beside **The always-on
     extra dispatch** (`skills/flow/review-panel.md`), reading `late-fix.diff` over the range since
     the panel's last clean close, written, paired, ceilinged and recorded as **The late-fix
     reduction** (`skills/flow/review-panel-late-fix.md`) states its dispatch, its prompt carrying
     the FIX-ROUND SCOPE paragraph of **Panel re-runs** (`skills/flow/review-panel-fix-round.md`).
     The two dispatches are keyed `panel-<round>-primary` and
     `panel-<round>-test-sanity+call-efficiency`, `<round>` the number `flow record next-round`
     prints, and the round records its `flow record pass -round <round>` rows like any panel round.
     That file's finding, voiding and staleness rules
     apply as written — a clean or Minor-only read leaves every slot it did not dispatch current.
     Once any round's read raises a Critical or Important, the reduction is void for the rest of the
     run: every later round's step 3 opens the decided full roster, never `primary` alone.
  ```

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0; the absence sweep `git grep -n -F -e 'run one dispatch: `primary` alone,' -- skills/flow/verify-fix-loop.md` prints nothing.
  - [x] **Step 4: Commit** `skills/flow/verify-fix-loop.md` with the subject below.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — a contract wording change; the verbatim, reference and markdown guards cover it.
**Regression:** none — no test is added; reverting restores the `primary`-alone wording the sweep finds.
**Commit:** fix(flow): re-review a verify fix round under panel keys beside the always-on extra
**After:** none

**Decision:** fix-round-review-keys

- [x] 2. Report a multi-repository change whose merge order misses a repository
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Write the test below and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckUnfinishedWorkMergeOrderRecord' -count=1`: the missing-record subtest fails, the verdict carrying no merge-order reason.
  - [x] **Step 2: Write `TestCheckUnfinishedWorkMergeOrderRecord`** in `stats/internal/guard/check_unfinished_work_test.go`, on the file's existing fixture of a worktree holding the plan, plus a second git repository `peer` beside the plan repository's main checkout and a `spectre/peers` file in the plan worktree naming `peer ../peer`. The guard is called with one range per repository. Subtests:
    - no `link.md`: the verdict names `merge order does not name <peer main checkout>`;
    - `link.md` holding `## Merge order` and only ``1. `.` ``: the same reason;
    - `link.md` holding ``1. `.` `` and ``2. `peer` ``: no such reason;
    - `link.md` holding ``1. `.` `` and ``2. `peer` `` under `## Part of` instead of `## Merge order`: the reason stands;
    - one range only (the plan repository's): no such reason, with no `link.md`;
    - the peer repository absent from `spectre/peers`, `link.md` naming ``2. `peer` ``: the reason stands.
  - [x] **Step 3: Implement signal ten** in `stats/internal/guard/unfinishedwork.go`, inside the `anchor == worktree` block beside signal nine and only when `planRC == 0`, per decision `merge-order-signal`: the plan worktree's main checkout from `changePlanMainCheckout`; for each range whose worktree's main checkout differs (both `filepath.Clean`ed), it is covered when `planDir + "/link.md"`'s `## Merge order` section — the lines after that heading up to the next `## ` — holds a numbered line whose backticked name `n` gives `changePlanPeerRoot(env, worktree, specRootLeaf(worktree, io.Discard), n)` equal to that main checkout. Each uncovered main checkout, once, appends the reason `merge order does not name <main checkout>: <planDir>/link.md — write the change's ## Merge order per skills/flow/cross-repo-worktrees.md ("Then make the merge-order record cover the whole set") before landing`. An absent `link.md` reads as no entries; one that exists and cannot be read is exit 2 like any unreadable file here. Add a `SIGNAL TEN` paragraph to `scripts/check-unfinished-work.sh`'s header after signal nine's, and a matching comment in the Go.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1` passes; `bash scripts/test-integrate-change.sh` ends with no failure.
  - [x] **Step 5: Commit** the three files below with the subject below.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`
**Tests:** `TestCheckUnfinishedWorkMergeOrderRecord`
**Regression:** `TestCheckUnfinishedWorkMergeOrderRecord` — reverting the commit clears a two-repository change with no merge-order record, and the missing-record subtest fails.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/check_unfinished_work_test.go @ branch self-review-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy (5 at 9adea2ca) -->
**Commit:** feat(guard): report a multi-repository change whose merge order misses a repository
**After:** none

**Decision:** merge-order-signal

- [x] 3. Record motion by CDP screencast only
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `git grep -n -F -e 'sampled back-to-back where no screencast is' -- skills/flow/visual-verify-verifier.md` prints the fallback line.
  - [x] **Step 2: Edit the motion step** of `skills/flow/visual-verify-verifier.md` so its lines from `trigger until the view settles` through `frame's own timestamp … read every one, and` read exactly as below — the fallback clause and the sampled-clock `<ms>` clause gone, one sentence added (decision `motion-screencast-only`):

  ```markdown verified:applied at plan time; check-verbatim-moves.sh printed its sentences, listed in verbatim-moves.txt
  trigger until the view settles — a CDP screencast (`Page.startScreencast`, each frame acknowledged
  with `Page.screencastFrameAck`) — while driving the trigger (a tap, a navigation, a drag through `page.mouse` in small
  steps). Never sample with `page.screenshot`: the Chromium page this step mandates always
  carries the screencast, and sampling skips the frames it catches (kan-995). Write the frames, in order, to `<changeRoot>/visual-verification/motion/<motion
  id>/<nnn>-<ms>.png`, `<ms>` the frame's offset in milliseconds from the trigger — the screencast
  frame's own timestamp — read every one, and
  ```

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0; the absence sweep `git grep -n -F -e 'sampled back-to-back where no screencast is' -e 'the clock read as a sampled screenshot' -- skills/flow/visual-verify-verifier.md` prints nothing.
  - [x] **Step 4: Commit** `skills/flow/visual-verify-verifier.md` with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — a contract wording change; the verbatim, reference and markdown guards cover it.
**Regression:** none — no test is added; reverting restores the fallback clause the sweep finds.
**Commit:** fix(flow): record motion by CDP screencast only
**After:** none

**Decision:** motion-screencast-only

- [x] 4. Replace a motion's strip on a re-record (shipped rooted at `<changeRoot>/visual-verification/motion/<motion id>/` for the citations guard; rule unchanged)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `git grep -n -F -e 'A re-record replaces the strip' -- skills/flow/visual-verify-verifier.md` prints nothing: no rule says what a second recording of a motion does with the first.
  - [x] **Step 2: Insert the paragraph below** in `skills/flow/visual-verify-verifier.md`, as its own lines right after the motion step's line ending `named by its index and what it shows.` and before `**A difference between two triggers is judged at one sampling.**` (decision `motion-screencast-only`):

  ```markdown verified:applied at plan time; check-verbatim-moves.sh printed its sentences, listed in verbatim-moves.txt
  **A re-record replaces the strip.** A later round's recording of a motion first deletes
  `motion/<motion id>/` and writes its frames there, never into a suffixed sibling such as
  `<motion id>-r2/`, and `visual-verification.md` cites an earlier round's strip by its round number,
  never by a path (kan-995: superseded strips were 28.6 MB of a 43 MB archive).
  ```
  <!-- measured: du -sh over the kan-995 archive's visual-verification/motion directories @ commit 3cf7aaaa, as recorded by the self-review pass -->

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0; `git grep -n -F -e 'A re-record replaces the strip' -- skills/flow/visual-verify-verifier.md` prints one line.
  - [x] **Step 4: Commit** `skills/flow/visual-verify-verifier.md` with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — a contract wording change; the verbatim, reference and markdown guards cover it.
**Regression:** none — no test is added; reverting removes the rule the grep finds.
**Commit:** fix(flow): replace a motion's strip when the verifier records it again
**After:** Task 3

**Decision:** motion-screencast-only

- [x] 5. Decode a transcript line whose toolUseResult is a string
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Write the test below and run `cd stats && make -s web-dist && go test ./internal/harvest/ -run 'TestWatcherRetiresBeginsOfHookDeniedLaunches' -count=1`: it fails, the panel begins stamped with the flake agent's and the primary reviewer's ids.
  - [x] **Step 2: Write `TestWatcherRetiresBeginsOfHookDeniedLaunches`** in `stats/internal/harvest/watcher_test.go`, replaying kan-995's panel round 1 (parent transcript lines 3918–3985) with `stampRecordingDeps`, its lines in this order and timestamps:
    - an assistant `Bash` line at `00:35:51.286Z` whose command runs two begins, `-role reviewer -key "panel-1-primary+principles" -slot "primary+principles" -session-token mf-k995` and `-key "panel-1-test-sanity+call-efficiency" -slot "test-sanity+call-efficiency"` with the same token;
    - two denied Agent calls: an assistant `tool_use` named `Agent` (ids `toolu_a`, `toolu_b`) at `00:36:03.257Z` and `00:36:11.607Z`, each followed by a user line at `00:36:03.314Z` / `00:36:11.666Z` whose content is a `tool_result` with `"is_error":true` for that id and whose top-level `"toolUseResult"` is the string `"Error: PreToolUse:Agent hook error: Dispatch blocked"`;
    - the flake agent's launch, a user line at `00:36:21.557Z` with `"toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"a0c94877856e38f6d"}`;
    - the two panel launches at `00:36:47.533Z` (`ad7ad88ef851a442c`) and `00:36:47.537Z` (`a783c4dc147d1cd18`);
    - an assistant `Bash` line at `00:36:51.097Z` running `flow record dispatch begin … -role panel-fix -key verify-flake-1 -session-token mf-k995`.
    Two subtests, each on a fresh watcher: every line in one batch, and one `RunOnce` after each appended line (`appendLine`). Both assert the stamps are exactly `[{mf-k995 verify-flake-1 a0c94877856e38f6d}]`.
  - [x] **Step 3: Implement** in `stats/internal/harvest/transcript.go`: `func (r *rawToolUseResult) UnmarshalJSON(b []byte) error` that returns nil and leaves `r` zero when `b` does not start with `{`, and otherwise decodes into the struct through a local alias type; a doc comment naming the hook-denied line's string `toolUseResult` and kan-995 (decision `tool-use-result-string`).
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./internal/harvest/ && go test ./internal/harvest/ -run 'TestWatcherRetiresBeginsOfHookDeniedLaunches' -count=1` passes; the whole harvest package, whose every `rawLine` parser now reads these lines, runs in the full suite after the last task.
  - [x] **Step 5: Commit** the two files below with the subject below.

**Build:** green
**Files:** `stats/internal/harvest/transcript.go`, `stats/internal/harvest/watcher_test.go`
**Tests:** `TestWatcherRetiresBeginsOfHookDeniedLaunches`
**Regression:** `TestWatcherRetiresBeginsOfHookDeniedLaunches` — reverting the commit drops both denials, the panel begins claim the flake and primary launches, and both subtests fail.
**Baseline:** before=61 after=62
<!-- measured: grep -c '^func Test' stats/internal/harvest/watcher_test.go @ branch self-review-kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy (61 at 9adea2ca) -->
**Commit:** fix(harvest): decode a transcript line whose toolUseResult is a string
**After:** none

**Decision:** tool-use-result-string
**Decision:** k995-rows-left

- [ ] 6. Derive the worktree's ports in gymie's verify and capture rows (landed in gymie as a8cc4e8d on spectre/kan-995-…)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In the gymie-playwright worktree, `echo ${GATEWAY_PORT:-8080}` prints the main checkout's port while the backend worktree's `.dev-stack/ports/gateway.port` holds the kan-995 stack's gateway port.
  - [x] **Step 2: Edit the gymie worktree's `.flow/project.md`** (decision `rows-derive-ports`): the `verify` and `capture` rows of the `## visual verification` command table become exactly the two rows below; in the paragraph opening **A change that moves captures re-baselines them**, the words naming the capture row's URL prefix become "under the `capture` row's prefix — its `GW`/`FE` reads and its `FRONTEND_URL`/`BACKEND_URL` —"; and in the paragraph opening **`verify` and `capture` set `BACKEND_URL` to the gateway**, its last sentence — the one opening "In the main checkout" — becomes the sentence below the rows.

  ```markdown verified:spiked at plan time from the kan-995 gymie-playwright worktree (9480, 4400) and the gymie-playwright main checkout (8080, 3000), under zsh and bash
  | `verify` | `GW=$(cat ../../gymie-worktrees/"${PWD##*/}"/.dev-stack/ports/gateway.port 2>/dev/null); FE=$(cat ../../gymie-worktrees/"${PWD##*/}"/.dev-stack/ports/frontend.port 2>/dev/null); FRONTEND_URL=http://localhost:${FRONTEND_PORT:-${FE:-3000}} BACKEND_URL=http://localhost:${GATEWAY_PORT:-${GW:-8080}} npx playwright test <specs>` |
  | `capture` | `GW=$(cat ../../gymie-worktrees/"${PWD##*/}"/.dev-stack/ports/gateway.port 2>/dev/null); FE=$(cat ../../gymie-worktrees/"${PWD##*/}"/.dev-stack/ports/frontend.port 2>/dev/null); FRONTEND_URL=http://localhost:${FRONTEND_PORT:-${FE:-3000}} BACKEND_URL=http://localhost:${GATEWAY_PORT:-${GW:-8080}} npx playwright test <spec> --update-snapshots` |
  ```

  ```markdown verified:authored in-tree for this change; its claims are the spike above
  `devStart` exports `GATEWAY_PORT` and `FRONTEND_PORT` only to the processes it spawns, never to the shell that runs a row, so each row reads them from the backend worktree's `.dev-stack/ports/` — `../../gymie-worktrees/<this checkout's directory name>/`, the change's backend worktree beside its regression worktree; an exported `GATEWAY_PORT` or `FRONTEND_PORT` still wins, and the main checkout, which has no such sibling, falls back to its own gateway and frontend, 8080 and 3000 (kan-995).
  ```

  - [x] **Step 3: Verify.** From the gymie-playwright worktree, the `verify` row's prefix with `env` in place of `npx playwright test <specs>`, piped to `grep _URL`, prints the backend worktree's `gateway.port` and `frontend.port` values; from `/Users/tweety53/Projects/gymie-playwright` (read only, nothing written there), the same prints `8080` and `3000`; with `GATEWAY_PORT=1` exported, `BACKEND_URL` reads port `1`. `git -C <gymie worktree> diff --check` is clean, and each of the two edited table rows still holds exactly three `|` characters, so the table keeps two cells per row.
  - [x] **Step 4: Commit** `.flow/project.md` in the gymie worktree with the subject below.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — project configuration; Step 3's live read of the prefix is the check.
**Regression:** none — no test is added; reverting brings back the 8080/3000 read the Step 1 reproduction shows.
**Commit:** fix(flow): derive the worktree's ports in the verify and capture rows
**After:** none

**Decision:** rows-derive-ports

- [ ] 7. Keep every Playwright run's artifacts in a directory of its own (landed in gymie as 480e9852 on spectre/kan-995-…)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `git -C <gymie worktree> grep -n -F -e '--output' -- .flow/project.md` prints nothing: both rows write to Playwright's default `test-results/`, which each run clears.
  - [x] **Step 2: Append ` --output test-results/$(date -u +%Y%m%dT%H%M%SZ)` to the command** in each of the `verify` and `capture` rows Task 6 wrote — after `<specs>` and after `--update-snapshots` — and add, after the paragraph Task 6 ended with the ports sentence, the paragraph below (decision `per-run-output-dir`).

  ```markdown verified:npx playwright test fidelity/kan-940-cardio-rows.spec.ts --list --output test-results/plan-spike exited 0 in the kan-995 gymie-playwright worktree
  **`verify` and `capture` give each run its own `--output` directory under `test-results/`** because Playwright clears its output directory when a run starts: a failure seen once under load lost its trace and error context to the next run (kan-995). `test-results/` is gitignored and goes with the worktree.
  ```

  - [x] **Step 3: Verify.** In the gymie-playwright worktree, the `verify` row's command with `--list` added and `<specs>` replaced by `fidelity/kan-940-cardio-rows.spec.ts` exits 0, and after it `test-results/` holds a directory named by the UTC stamp; delete that directory. The two rows still hold exactly three `|` characters each, and `git diff --check` is clean.
  - [x] **Step 4: Commit** `.flow/project.md` in the gymie worktree with the subject below.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — project configuration; Step 3's run is the check.
**Regression:** none — no test is added; reverting returns both rows to the shared output directory Step 1 shows.
**Commit:** fix(flow): give each Playwright run its own output directory
**After:** Task 6

**Decision:** per-run-output-dir

- [ ] 8. Trace cardio-R1f under load and fix it where the trace points (not reproduced: 20/20 at HEAD, 20/20 with dadc7396 reverted, full fidelity 279 passed — nothing to commit)
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In the gymie-playwright worktree, under the `verify` row's prefix as Tasks 6 and 7 left it, run `npx playwright test fidelity/kan-940-cardio-rows.spec.ts:134 --repeat-each 20` (the config's own workers) at HEAD and record runs and failures. Then reverse `dadc7396`'s `fixtures.ts` hunk in the working tree (`git show dadc7396 -- fixtures.ts | git apply -R`), run the same command again, record its counts, and restore (`git checkout -- fixtures.ts`); `git status --short` shows no change before the next step. Neither run failing: run the whole `fidelity` project once under the same prefix — the bounded full-suite exception of gymie's `docs/flake-bisection.md`, stated before the run — and record its cardio-R1f result.
  - [ ] **Step 2: Decide by the counts** (decision `r1f-trace-first`):
    - failures only with `dadc7396` reversed: `dadc7396` already covers the flake — commit nothing, and report both runs' counts and the failing assertion as the finding's closing evidence;
    - failures at HEAD: read the failing run's trace and error context from its `--output` directory, fix the nondeterminism at its source in `fixtures.ts` or `fidelity/kan-940-cardio-rows.spec.ts` — never a longer timeout, a retry or a sleep — then prove it: the repeat run passes with the fix, and fails again with the fix's mechanism removed; restore the fix;
    - no failure anywhere: commit nothing, and report every run's counts as evidence the flake did not reproduce.
  - [ ] **Step 3: Verify** (fix branch only). `npx playwright test fidelity/kan-940-cardio-rows.spec.ts --list` exits 0 (the spec and `fixtures.ts` compile), and the Step 1 repeat command passes at the fix's commit.
  - [ ] **Step 4: Commit** (fix branch only) the files the fix touched, a subset of the two below, with the subject below.

**Build:** green
**Files:** `fixtures.ts`, `fidelity/kan-940-cardio-rows.spec.ts`
**Tests:** none — repairs existing tests: cardio-R1f, whose repeat run is the reproduction.
**Regression:** none — no test is added; on the fix branch, reverting the commit brings back the failures the repeat run counted.
**Commit:** fix(fidelity): settle cardio-R1f's fold taps under load
**After:** Task 6, 7

**Decision:** r1f-trace-first
