# kan-441-frontend-implementer-loops-dominate-cache-read

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** cut the number and weight of test runs inside an implementer's loop — targeted runs per
task, the full suite once per worktree at the last bundle — without moving `FOREGROUND BUILDS`.

**Architecture:** two new dispatch paragraphs in `skills/flow/implement.md`. `TARGETED TESTS` is
guarded by `scripts/check-dispatch-paragraphs.sh` at the implementer dispatch and the panel-fix
dispatch (`skills/flow/review-panel.md`), following KAN-263's table-row pattern. `FULL SUITE` is
appended by the conductor to the last bundle's implementer prompt only and is ordinary prose; the
conductor's last-boundary paragraph gains the red-branch stop.

**Tech Stack:** Markdown skill files; bash guard and its fixture-based harness.

**Spec:** `spectre/changes/kan-441-frontend-implementer-loops-dominate-cache-read/design.md`

## Global Constraints

- `FOREGROUND BUILDS` blocks and the guard's `foreground` row stay byte-identical (design decision
  `foreground-kept-cut-runs`).
- No new `## test targeted` key or any other `.flow/project.md` change (`selector-from-tests-field`).
- New prose carries no `SHALL`/`MUST` sentence, so `scripts/check-normative-inventory.sh`'s
  inventory is unchanged; `skills/flow/implement.md` stays under its 32500-byte budget in
  `scripts/check-contract-budget.sh` (28389 bytes today).
  <!-- measured: wc -c skills/flow/implement.md; grep -n 'skills/flow/implement.md' scripts/check-contract-budget.sh @ main 981247b -->
- Every guard listed under `## lint` in `.flow/project.md` exits clean before a task is reported done.

---

- [x] 1. Add the guarded `TARGETED TESTS` paragraph to the implementer and panel-fix dispatches

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`,
`scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
**Tests:** Case 22: the `TARGETED TESTS` label absent from `implement.md` exits 1 naming the file
and the label; Case 23: a block missing "the build tool's own selector" exits 1; Case 24: a block
missing "once for RED, once for GREEN" exits 1; Case 25: a block missing "module or repository
suite mid-task" exits 1; Case 26: the label absent from `review-panel.md` exits 1 naming that file
**Regression:** revert this commit and cases 22–26 do not exist — the harness knows nothing of the
label — and `scripts/check-dispatch-paragraphs.sh` passes a tree from which `TARGETED TESTS` has
been trimmed away, the silent drift KAN-289 burned on.
**Baseline:** before=46 after=56
<!-- measured: bash scripts/test-check-dispatch-paragraphs.sh | grep -c '^ok:' @ main 981247b (46 ok lines, "all cases passed") -->
<!-- predicted: the same command after task 1 — cases 22–26 each print two ok lines (exit code, then the named file/label/phrase), 46 + 10 -->
**Commit:** feat(flow): run only the task's own tests in the implementer loop

  - [x] **Step 1: Write the failing harness cases first**

    In `scripts/test-check-dispatch-paragraphs.sh`, after case 21 (its last `esac`, line 673) and
    before the `if [ "$FAILURES" -ne 0 ]` tail, add cases 22–26, each in the shape of cases 16–19
    (lines 513–600): `new_root`, two `write_site` calls, `run_guard`, an `[ "$RC" -eq 1 ]` check,
    then a `case "$OUT" in` naming what the output must carry. Define the fixture blocks beside
    `FOREGROUND_BLOCK` (line 158): `TARGETED_BLOCK` verbatim as Step 3 states it, plus
    `TARGETED_BLOCK_NO_SELECTOR`, `TARGETED_BLOCK_NO_RED_GREEN` and `TARGETED_BLOCK_NO_SUITE`, each
    the same block with exactly one required phrase reworded away, mirroring
    `FOREGROUND_BLOCK_NO_*` (lines 162–175).
    <!-- measured: grep -n 'FOREGROUND_BLOCK\|^# Case\|^if \[ "\$FAILURES"' scripts/test-check-dispatch-paragraphs.sh @ main 981247b -->

    - Case 22: `review-panel.md` correct (its case-1 content plus one `TARGETED_BLOCK`),
      `implement.md` its case-1 content with no `TARGETED_BLOCK` — exit 1, output carries
      `implement.md` and `TARGETED TESTS`.
    - Cases 23–25: both files correct except `implement.md`'s block is the `_NO_SELECTOR`,
      `_NO_RED_GREEN`, `_NO_SUITE` variant respectively — exit 1, output carries `implement.md`
      and the dropped phrase.
    - Case 26: `implement.md` correct, `review-panel.md` its case-1 content with no
      `TARGETED_BLOCK` — exit 1, output carries `review-panel.md` and `TARGETED TESTS`.

    Update the header comment (lines 11–45) with a `# Cases 22-26` paragraph in the style of the
    `# Cases 16-19` one.

  - [x] **Step 2: Run the harness and see cases 22–26 fail**

    `bash scripts/test-check-dispatch-paragraphs.sh` — expect `FAIL: case 22 …` through `case 26`
    (the guard has no `targeted` row yet, so every fixture exits 0) and the final `5 case(s)
    failed`, while cases 1–21 still pass.

  - [x] **Step 3: Add the paragraph at both dispatch sites**

    In `skills/flow/implement.md`, insert this block immediately after the `FOREGROUND BUILDS`
    block that follows "Every implementer dispatch **must** also carry:" (lines 383–387) and before
    the `REPRODUCE, DON'T READ` block (line 389), blank line on each side:
    <!-- measured: grep -n 'also carry\|FOREGROUND BUILDS\|REPRODUCE' skills/flow/implement.md @ main 981247b -->

```markdown verified:authored in-tree for this change
> **TARGETED TESTS:** Run only the tests this task's `**Tests:**` field names, through the build
> tool's own selector — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t
> '<name>'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never
> run the module or repository suite mid-task: the full `## test` list runs once per worktree at
> the last bundle, and again in `flow.verify`. Pipe a test run's output through `tail` so a green
> run costs lines of context, not a build log.
```

    In `skills/flow/review-panel.md`, insert the same block with its own lead-in, immediately after
    the panel-fix `FOREGROUND BUILDS` block (lines 728–732) and before the `REPORT FILE` lead-in
    (line 734):
    <!-- measured: grep -n 'also carries the FOREGROUND\|also carries the REPORT FILE' skills/flow/review-panel.md @ main 981247b -->

```markdown verified:authored in-tree for this change
**Every fix subagent's dispatch prompt also carries the TARGETED TESTS paragraph**:
```

    followed by the blockquote above, verbatim. Do not add it to the panel slot dispatch or the
    conductor's own §4 restatement — reviewers do not run the task's tests, and the guard's minimum
    is one block per file.

  - [x] **Step 4: Add the `targeted` row to the guard**

    In `scripts/check-dispatch-paragraphs.sh` (tables at lines 109–137):
    <!-- measured: grep -n 'ENTRY_LABEL=\|ENTRY_SHARED_PHRASES=\|ENTRY_VARIANTS=\|SITE_ENTRY=\|SITE_VARIANTS=' scripts/check-dispatch-paragraphs.sh @ main 981247b -->

    - `ENTRY_LABEL`: `[targeted]="**TARGETED TESTS:**"`.
    - `ENTRY_SHARED_PHRASES`: `[targeted]="the build tool's own selector${US}once for RED, once
      for GREEN${US}module or repository suite mid-task"`.
    - `ENTRY_VARIANTS`: `[targeted]=""`.
    - Append one element to each of the four `SITE_*` arrays, twice: `targeted` /
      `skills/flow/implement.md` / `1` / `""`, and `targeted` / `skills/flow/review-panel.md` /
      `1` / `""`.
    - Header comment: add the two table lines and a "TARGETED TESTS shared phrases" paragraph in
      the shape of the `FOREGROUND BUILDS` ones (lines 40–58), and extend the "KAN-263 added a
      third required paragraph" sentence with one naming KAN-441's fourth.

  - [x] **Step 5: Add the block to every existing fixture, run the harness green**

    Every `write_site` in cases 1–21 that builds `implement.md` or `review-panel.md` as a *correct*
    site now also carries one `$TARGETED_BLOCK` (blank line separated), exactly as KAN-263 added
    `$FOREGROUND_BLOCK` to them — otherwise the new row fails every case for the wrong reason.
    Case 15's glued-block fixture is left alone; it tests continuation, not this row.
    `bash scripts/test-check-dispatch-paragraphs.sh` — expect `all cases passed` and 56 `ok:`
    lines. Then `scripts/check-dispatch-paragraphs.sh` against the real tree — exit 0.

  - [x] **Step 6: Run the project's lint list and commit**

    Every command under `## lint` in `.flow/project.md`, in order; fix any hit. Commit with the
    declared subject and a `Task-Id: 1` trailer.

- [x] 2. Add the `FULL SUITE` paragraph to the last bundle's dispatch and the conductor's red-branch stop

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** **none** — a dispatch-prose change at one site, unguarded by design
(`targeted-guarded-full-suite-not`); the change's own verification is the guard list in Step 3.
**Regression:** none declared — reverting restores today's behaviour, where the full list first
runs in `flow.verify`, after the panel.
**Baseline:** before=56 after=56
<!-- measured: bash scripts/test-check-dispatch-paragraphs.sh | grep -c '^ok:' @ branch spectre/kan-441-frontend-implementer-loops-dominate-cache-read after task 1 (predicted there as 46 + 10) -->
<!-- predicted: the same command after task 2 — no harness change in this task -->
**Commit:** feat(flow): run the full suite once, at the last bundle

  - [x] **Step 1: State the last bundle's extra paragraph beside the preamble**

    In `skills/flow/implement.md`, after the `REPORT FILE` block (lines 395–399, the last block
    under "Every implementer dispatch **must** also carry:") and before "**The next implementer
    overlaps the guard.**" (line 401), add:
    <!-- measured: grep -n 'REPORT FILE\|The next implementer overlaps' skills/flow/implement.md @ main 981247b -->

```markdown verified:authored in-tree for this change
**The last bundle's implementer dispatch — the last `bundle <k>` line `plan-dispatch-bundles.sh`
printed — alone also carries:**

> **FULL SUITE:** Yours is the last bundle. After GREEN and before your commit, run the resolved
> `## test` list once, in the foreground, in the order the context bundle carries it. A failure in
> a file this task's `**Files:**` field names is yours: fix it and re-run. Any other failure is
> not: record the command and its output verbatim in your REPORT FILE under a `## Full suite`
> heading, unfixed, and still commit your own task.
```

  - [x] **Step 2: Add the red-branch stop to the last boundary**

    Replace the paragraph "**The last bundle's guard pass is the stage's last boundary.**" (lines
    433–435) with:
    <!-- measured: grep -n "The last bundle's guard pass" skills/flow/implement.md @ main 981247b -->

```markdown verified:authored in-tree for this change
**The last bundle's guard pass is the stage's last boundary.** `final-review.diff` is written and
the slots dispatched once it has passed and the last implementer's report carries no `## Full
suite` failure; the review panel's pre-work may share the last implementer's wait, in its one
call. A report that records a full-suite failure ends your turn with `## Question` — the failing
command and its output, verbatim — before `final-review.diff` is written: the panel never runs on
a red branch, and the operator resolves it through a fix run.
```

  - [x] **Step 3: Run the project's lint list and commit**

    Every command under `## lint` in `.flow/project.md`, in order — `check-dispatch-paragraphs.sh`
    still sees two `FOREGROUND BUILDS` blocks and one `TARGETED TESTS` block in `implement.md`;
    `check-contract-budget.sh` still reports `BUDGET-OK`. Commit with the declared subject and a
    `Task-Id: 2` trailer.
