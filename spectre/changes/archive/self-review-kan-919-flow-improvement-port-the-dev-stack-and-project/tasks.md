# self-review-kan-919-flow-improvement-port-the-dev-stack-and-project

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-919's five self-review findings — F-A (Task 1), F-B (Task 2), F-C (Task 3),
F-D (Task 4), F-E (Task 5), as `proposal.md` names them.

**Architecture:** a test's shared fixture value made atomic; `run-verify-list.sh` keeps a failing
command's full output in the git dir and marks a data race, with the failing-command rule's prose;
a new `check-plan-shape.py` finding fed by two raw values the commit-fields parser exposes; a new
shape-class refusal in `run-reproducer`.

**Tech Stack:** Go (`stats/`), Bash, Python 3 (standard library), Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-22-self-review-kan-919-flow-improvement-port-the-dev-stack-and-project-design.md`
in this worktree.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-919-flow-improvement-port-the-dev-stack-and-project` (`<wt>` below), on branch `spectre/kan-919-flow-improvement-port-the-dev-stack-and-project`; every `**Files:**` path is relative to it. Never touch the main checkout and never create a worktree.
- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `patches/task-3.diff` and `verbatim-moves.txt` are already committed with the plan. Never merge or push.
- Go: `gofmt -w` every Go file you change before committing; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Prose lands character for character as `patches/task-3.diff` carries it: `scripts/check-verbatim-moves.sh` matches the added sentence against `verbatim-moves.txt` exactly.
- `stats/internal/guard/runreproducer.go` keeps exactly one line matching `rrMarker =` and exactly one matching `range headLines(path, `, and none of `runreproducer.go`, `scripts/run-reproducer.sh` or `skills/flow/review-panel.md` gains a backticked span opening `# mutation-` other than the canonical marker or a "first <N> lines" statement other than 10: `scripts/check-mutation-reproducer-pin.sh` must keep exiting 0.
- Never stop or restart `flowd` on `127.0.0.1:4173` nor the `flow-postgres` container.

## Review Focus

- Task 2: a passing command keeps no file; a run outside a git worktree prints no `full output:` line; a later run never leaves an earlier run's log behind under a path it prints.
- Task 3: a race banner that the 20-line tail dropped still marks its block; a plain failure is not marked.
- Task 4: a single unbackticked path, a `none, unless …` collateral sentence and every backticked list still pass; the rule reads the parser's own joined value, continuation lines included.
- Task 5: an exact marker anywhere in the first 10 lines still declares even beside a near-miss line; a mid-line marker stays generic; the refusal never executes the reproducer. <!-- measured: grep -n 'headLines(path, 10)' stats/internal/guard/runreproducer.go @ 9468cc05 → line 502 -->

- [x] 1. Order the dev-stack freshness test's served bundle for the race detector
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `cd stats && make -s web-dist && go test ./internal/guard -race -run 'TestCheckDevStackFresh$' -count=1` fails with `WARNING: DATA RACE` between `check_dev_stack_fresh_test.go:71` and `:75`.
    <!-- measured: cd stats && go test ./internal/guard -race -run 'TestCheckDevStackFresh$' -count=1 @ 9468cc05 → FAIL -->
  - [x] **Step 2: Fix it** per decision `race-atomic-served`: in case 3 of `TestCheckDevStackFresh`, `served` becomes a `sync/atomic` `atomic.Value` holding `"stale-bundle\n"`, the handler writes `served.Load()`, and the switch to `"fresh-bundle\n"` is `served.Store(…)`. Nothing else in the file changes.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -race -run 'TestCheckDevStackFresh$' -count=5` passes with no `DATA RACE` line. Commit.

**Build:** green
**Files:** `stats/internal/guard/check_dev_stack_fresh_test.go`
**Tests:** none — repairs existing tests: `TestCheckDevStackFresh`
**Regression:** none — Step 1's `-race` run fails again on revert.
**Commit:** fix(guard): order the dev-stack freshness test's served bundle for the race detector
**After:** none

**Decision:** race-atomic-served

- [x] 2. Keep a failing verify command's full output in the git dir
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `sed -n '121,122p' scripts/run-verify-list.sh` prints the `mktemp -d` line and `trap 'rm -rf "$OUTDIR"' EXIT`; `grep -c 'full output' scripts/run-verify-list.sh` prints `0`.
    <!-- measured: both commands @ 9468cc05 → the two lines; 0 -->
  - [x] **Step 2: Write the cases first** in `scripts/test-run-verify-list.sh`. After the `lint appends check-spec-reach.sh` assertion add `expect_lacks "outside a git worktree no output is kept" 'full output:'`. Before `run_bin "$ROOT" build`, add a git fixture `F="$ROOT/fail"` (its own `git init`, committed with `-c user.name=t -c user.email=t@t`) whose `.flow/project.md` `## test` fence holds `seq 1 25; exit 4` then `true`; run `run_bin "$F" test`, set `KEPT="$(git -C "$F" rev-parse --path-format=absolute --git-path flow-verify-failed)/test-0.log"`, then exactly these six lines: `expect_rc "a failing list in a git worktree" 1`; `expect_has "a failing command names its kept full output" "  full output: $KEPT"`; `[ "$(head -n 1 "$KEPT" 2>/dev/null)" = 1 ] && pass "the kept output holds what the report truncated" || fail "kept output: $(head -n 1 "$KEPT" 2>/dev/null)"`; `expect_lacks "a passing command keeps no output" 'test-1.log'`; `[ -z "$(git -C "$F" status --porcelain)" ] && pass "the kept output leaves the tree clean" || fail "the kept output dirtied the tree"`; then rewrite the fence to `true` alone, commit, `run_bin "$F" test`, and `[ ! -e "$KEPT" ] && pass "a later run clears an earlier run's kept output" || fail "a stale kept output survived"`. Run `bash scripts/test-run-verify-list.sh` — the new cases fail.
  - [x] **Step 3: Implement** per decision `verify-output-kept-in-git-dir` in `scripts/run-verify-list.sh`: resolve `git -C "$WT" rev-parse --path-format=absolute --git-path flow-verify-failed` once (empty outside git); when set, `mkdir -p` it and remove `"$LIST"-*.log` in it before any command runs; in `report`, on a non-zero exit with the directory set, copy the output file to `<dir>/<list>-<n>.log` (`<n>` the command's 0-based list position, `spec-reach` for the appended `check-spec-reach.sh` run) and print `  full output: <path>` as the block's last line. Update the header's OUTPUT section to name that line and the file, and drop nothing else from the header.
  - [x] **Step 4: Verify.** `bash scripts/test-run-verify-list.sh` prints `all cases passed`; `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `scripts/run-verify-list.sh`, `scripts/test-run-verify-list.sh`
**Tests:** `a failing command names its kept full output`, `the kept output holds what the report truncated`, `a later run clears an earlier run's kept output`
**Regression:** `a failing command names its kept full output` fails — a failing command's output is deleted with the temporary directory again.
**Baseline:** before=34 after=41 <!-- measured: grep -c -E '(^expect_(rc|has|lacks) |&& pass ")' scripts/test-run-verify-list.sh @ 9468cc05 → 34 -->
**Commit:** fix(scripts): keep a failing verify command's full output in the git dir
**After:** none

**Decision:** verify-output-kept-in-git-dir

- [x] 3. Never treat a data race as an environmental flake
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -c 'DATA RACE' skills/flow/verify-and-handoff.md scripts/run-verify-list.sh` prints `0` for both files.
    <!-- measured: that grep @ 9468cc05 → 0 and 0 -->
  - [x] **Step 2: Write the cases first** in `scripts/test-run-verify-list.sh`, after the `test never appends check-spec-reach.sh` assertion: write `$ROOT/.flow/project.md` (a `printf` as the file's other rewrites do) with a `## test` fence of three commands — `printf 'WARNING: DATA RACE\n'; seq 1 25; exit 1`, `printf 'testing.go:1: race detected during execution of test\n'; exit 1` and `printf 'plain\n'; exit 1` — run `run_bin "$ROOT" test`, then exactly these three lines: `expect_rc "a list with a race fails" 1`; `expect_has "a race banner the tail dropped still marks its block" $'  25\n  data race — never environmental'`; `[ "$(grep -c 'data race — never environmental' <<<"$OUT")" -eq 2 ] && pass "each race block is marked, a plain failure is not" || fail "race marks: $(grep -c 'data race — never environmental' <<<"$OUT")"`. Run `bash scripts/test-run-verify-list.sh` — the new cases fail.
  - [x] **Step 3: Implement** per decision `race-never-environmental`: in `report` of `scripts/run-verify-list.sh`, after the output lines (and before Task 2's `full output:` line), print `  data race — never environmental` when `grep -Eq 'WARNING: DATA RACE|race detected during execution of test'` matches the whole output file; name the line in the header's OUTPUT section and append to the header's "retries nothing" sentence that a race-marked command takes **The loop** whatever its re-run exits. Then `git apply <wt>/spectre/changes/self-review-kan-919-flow-improvement-port-the-dev-stack-and-project/patches/task-3.diff` from `<wt>` for the prose in `skills/flow/verify-and-handoff.md`.
    <!-- measured: git apply --check patches/task-3.diff @ 9468cc05 → exit 0; scripts/check-verbatim-moves.sh with the patch applied reports exactly the one sentence verbatim-moves.txt lists -->
  - [x] **Step 4: Verify.** `bash scripts/test-run-verify-list.sh` prints `all cases passed`; `scripts/check-verbatim-moves.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-references.sh` and `scripts/check-dispatch-paragraphs.sh` exit 0. Commit.

**Build:** green
**Files:** `scripts/run-verify-list.sh`, `scripts/test-run-verify-list.sh`, `skills/flow/verify-and-handoff.md`
**Tests:** `a race banner the tail dropped still marks its block`, `each race block is marked, a plain failure is not`
**Regression:** `each race block is marked, a plain failure is not` fails — a race's block carries no mark and a re-run clears it again.
**Baseline:** before=41 after=44 <!-- measured: grep -c -E '(^expect_(rc|has|lacks) |&& pass ")' scripts/test-run-verify-list.sh @ 9468cc05 → 34, plus Task 2's seven -->
**Commit:** fix(flow): never treat a data race as an environmental flake
**After:** Task 2

**Decision:** race-never-environmental

- [x] 4. Refuse an unbackticked multi-path Files or Allowed-collateral value
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `git show 75ded465:spectre/changes/kan-919-flow-improvement-port-the-dev-stack-and-project/tasks.md | sed '/^\*\*Files:\*\*/s/`//g' > "$TMPDIR/stripped-tasks.md" && scripts/check-plan-shape.sh "$TMPDIR/stripped-tasks.md"` exits 0.
    <!-- measured: that pipeline with a scratch path @ 9468cc05 → exit 0 -->
  - [x] **Step 2: Write the case first** in `scripts/test-check-plan-shape.sh`: `# Case 40 (F16)` after case 39, a fixture of four tasks — `**Files:** a.py, b.py`; `**Files:** \`c.py\`` with `**Allowed-collateral:** docs/a.md docs/b.md`; `**Files:** d.py` (one unbackticked path); `**Files:** \`e.py\`` with `**Allowed-collateral:** none, unless step 1 says otherwise` — each with a backticked `**Tests:**`, `**After:** none` and a `**Commit:**`. Assert exit 1; the output names task 1's `**Files:**` and task 2's `**Allowed-collateral:**` with the substring `names no backticked path`; it names neither task 3 nor task 4; then `assert_mutation_removes_finding "F16" "$TASKS_MD" "names no backticked path" "case 40"`. Update the header's "One case per finding (F1-F15)" to F1-F16. Run `bash scripts/test-check-plan-shape.sh` — case 40 fails.
  - [x] **Step 3: Implement** per decision `plan-shape-f16`. In `scripts/check-task-commit-fields.py`, add `files_value: str = ""` and `collateral_value: str = ""` to `TaskFields` beside `tests_value`, set from the `files_value` and `collateral_value` strings `parse_task_fields` already joins; nothing else there changes. In `scripts/check-plan-shape.py`, in `_check_task`, for each of the two fields whose joined value is non-empty, carries no `BACKTICK_RE` match, does not match `NONE_OPEN_RE`, and contains a comma or whitespace, report `<path>:<field line>: task <id>'s **<Field>:** value names no backticked path but carries a comma or whitespace — the commit guard reads it as one path; backtick each path`, the decisive `if` line ending `# F16` (the mutation helper's `sed` anchors on it). Add F16 to the module docstring's Findings list, make "Fifteen findings, F1-F15" and the exit-1 line read F1-F16, and amend F2's note on the prose fallback to say F16 catches its multi-path shape.
  - [x] **Step 4: Verify.** `bash scripts/test-check-plan-shape.sh` prints `all cases passed`; `scripts/check-plan-shape.sh` (every in-flight plan) and `scripts/check-python-suppressions.sh` exit 0; Step 1's pipeline now exits 1 naming F16's message. Commit.

**Build:** green
**Files:** `scripts/check-plan-shape.py`, `scripts/test-check-plan-shape.sh`, `scripts/check-task-commit-fields.py`
**Tests:** Case 40
**Regression:** Case 40 fails — an unbackticked comma-separated `**Files:**` list passes the plan-shape guard again.
**Baseline:** before=39 after=40 <!-- measured: grep -c -E '^# Case [0-9]+' scripts/test-check-plan-shape.sh @ 9468cc05 → 39 -->
**Commit:** fix(scripts): refuse an unbackticked multi-path Files value in plan shape
**After:** none

**Decision:** plan-shape-f16

- [x] 5. Refuse a near-miss mutation-reproducer marker
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 'case 28.b' stats/internal/guard/runreproducer_test.go` shows a marker line carrying extra text expected to run generic (`want: 1`).
    <!-- measured: that grep @ 9468cc05 → line 183, want: 1 -->
  - [x] **Step 2: Write the cases first** in `TestRunReproducer`'s rows: case 28.b becomes `case 28.b: a marker line carrying extra text is refused unrun`, `want: 2`, `needle: "near-miss"`, `notRan: true`; add `case 28.e: a marker mid-line is prose and declares nothing` (body `: see # mutation-reproducer\nexit 0`, `want: 1`) and `case 28.f: an indented marker is a near-miss and is refused` (body `  # mutation-reproducer\nexit 0`, `want: 2`, `needle: "near-miss"`, `notRan: true`). Run `cd stats && go test ./internal/guard -run 'TestRunReproducer$' -count=1` — 28.b and 28.f fail.
  - [x] **Step 3: Implement** per decision `mutation-marker-near-miss-refused` in `stats/internal/guard/runreproducer.go`: a function returning the first of `headLines(file, 10)` whose `strings.TrimSpace` opens with `rrMarker` while the line is not exactly `rrMarker`; right after `mutation := declaresMutation(resolved)`, when `mutation` is false and that line exists, `return refuse(…)` with a message naming the line and the word `near-miss`, stating that the declaration is that exact line alone. `declaresMutation` itself and its `range headLines(path, 10)` line stay as they are. In `scripts/run-reproducer.sh`'s header, add one sentence to THE MUTATION-REPRODUCER CONVENTION paragraph (a line within the first 10 lines that opens with the marker but is not exactly it is refused, never read under either convention) and name that refusal in exit 2's shape class. <!-- measured: grep -n 'headLines(path, 10)' stats/internal/guard/runreproducer.go @ 9468cc05 → line 502 -->
  - [x] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestRunReproducer$|TestCheckMutationReproducerPin|TestCheckPanelReproducerExitContract' -count=1` passes; `scripts/check-mutation-reproducer-pin.sh` and `scripts/check-references.sh` exit 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`, `scripts/run-reproducer.sh`
**Tests:** `case 28.e: a marker mid-line is prose and declares nothing`, `case 28.f: an indented marker is a near-miss and is refused`
**Regression:** `case 28.f: an indented marker is a near-miss and is refused` fails — a near-miss marker runs under the generic convention, its verdict inverted, again.
**Baseline:** before=26 after=28 <!-- measured: grep -c -E '\{label: "case ' stats/internal/guard/runreproducer_test.go @ 9468cc05 → 26 -->
**Commit:** fix(guard): refuse a near-miss mutation-reproducer marker
**After:** none

**Decision:** mutation-marker-near-miss-refused
