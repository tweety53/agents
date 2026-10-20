# self-review-kan-999-group-session-workout-names

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the four non-`big` findings of `kan-999-group-session-workout-names`'s self-review
pass — the both-trees note on a frozen openspec tree, a planner rule for types at their size limit,
the missing end time in rendered ledgers, and gymie's live-check fixture and before-stack guidance.

**Architecture:** one narrowed branch in `spec_root_leaf` and its Go port, one sentence in
`skills/flow/brainstorm-planner.md` section D, one optional line in `RenderLedger`, and two prose
paragraphs in gymie's `.flow/project.md`. Each task applies its committed patch from `patches/`
beside this file.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-20-self-review-kan-999-group-session-workout-names-design.md` in this
worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` in this agents worktree: this change's `patches/` and `verbatim-moves.txt` are already committed with the plan. Never merge or push.
- Each task names the repository it runs in, by its worktree; every `**Files:**` path is relative to that worktree's root. Tasks 1–3 run in this agents worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-999-group-session-workout-names`, branch `self-review-kan-999-group-session-workout-names`; Task 4 runs in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-999-group-session-workout-names`, branch `spectre/kan-999-group-session-workout-names`, whose stack the operator may have running — never start, stop or rebuild it. Never touch a main checkout, and never create a worktree.
- `P` below is this change root, `/Users/tweety53/Projects/agents-worktrees/self-review-kan-999-group-session-workout-names/spectre/changes/self-review-kan-999-group-session-workout-names`. Each task's diff is `$P/patches/task-<n>.diff`, applied with `git apply` from its worktree's root; the four are independent. A patch that does not apply is reported, never hand-merged by guesswork.
- Run `cd stats && gofmt -w .` on Go changes before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 1: both trees present still returns `spectre`, and an `openspec/changes/` holding any entry but `archive/` still prints the note (`TestSpecRoot`'s open-change case).
- Task 3: an open dispatch (`EndedAt` nil) renders no Ended line.

- [ ] 1. Print the both-trees note only for an open openspec change
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, apply the test hunks alone — `git apply --include='stats/internal/guard/helpers_test.go' --include='scripts/test-lib-change-plan.sh' $P/patches/task-1.diff` — and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestSpecRoot' -count=1` and `bash scripts/test-lib-change-plan.sh`: both fail, the archive-only cases printing the note.
    <!-- measured: the test hunks alone, go test -run TestSpecRoot and bash scripts/test-lib-change-plan.sh @ branch self-review-kan-999-group-session-workout-names (c53570af) → FAIL, case 17 got the note -->
  - [ ] **Step 2: Apply the fix** — the rest: `git apply --exclude='stats/internal/guard/helpers_test.go' --exclude='scripts/test-lib-change-plan.sh' $P/patches/task-1.diff` (`spec_root_leaf` and `specRootLeaf` print the note only while `openspec/changes/` holds an entry other than `archive/`; the bash header says why).
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestSpecRoot' -count=1` passes; `bash scripts/test-lib-change-plan.sh` ends `✓ PASS`; `bash -c 'source scripts/lib/spec-root.sh; spec_root_leaf /Users/tweety53/Projects/gymie-worktrees/kan-999-group-session-workout-names'` prints `spectre` and nothing on stderr; `scripts/check-references.sh` exits 0.
  - [ ] **Step 4: Commit** the four files below with the subject below.

**Build:** green
**Files:** `scripts/lib/spec-root.sh`, `scripts/test-lib-change-plan.sh`, `stats/internal/guard/specroot.go`, `stats/internal/guard/helpers_test.go`
**Tests:** Case 17
**Regression:** Case 17 — reverting the commit prints the note for an `openspec/changes/` of `archive/` alone, and case 17 and `TestSpecRoot`'s archive-only case fail.
**Baseline:** before=9 after=10
<!-- measured: grep -cE '^CASE[0-9]+[A-Z]*=' scripts/test-lib-change-plan.sh @ branch self-review-kan-999-group-session-workout-names, before and after patches/task-1.diff -->
**Commit:** fix(scripts): note a both-trees project only for an open openspec change
**After:** none

**Decision:** note-only-for-open-openspec-change
**Decision:** spec-root-both-sides

- [ ] 2. Count a grown type against the size rules when planning it
  - [ ] **Step 1: Apply** `git apply $P/patches/task-2.diff` in this agents worktree: one blockquote sentence in section D of `skills/flow/brainstorm-planner.md`, after the verify-step rule.
  - [ ] **Step 2: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh` and `python3 scripts/check-markdown-integrity.py` exit 0 — the new sentence is listed in `$P/verbatim-moves.txt`.
  - [ ] **Step 3: Commit** `skills/flow/brainstorm-planner.md` with the subject below.

**Build:** green
**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — prose.
**Commit:** docs(flow): count a grown type against the size rules when planning it
**After:** none

**Decision:** size-rule-in-section-d

- [ ] 3. Render when a dispatch ended in the ledger
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, apply the test hunk alone — `git apply --include='stats/internal/records/render_test.go' $P/patches/task-3.diff` — and run `cd stats && make -s web-dist && go test ./internal/records/ -run 'TestRenderLedgerNamesWhenADispatchEnded' -count=1`: it fails, the ledger carrying no `- Ended:` line.
    <!-- measured: the test hunk alone, go test -run TestRenderLedgerNamesWhenADispatchEnded @ branch self-review-kan-999-group-session-workout-names (c53570af) → FAIL -->
  - [ ] **Step 2: Apply the fix** — the rest: `git apply --exclude='stats/internal/records/render_test.go' $P/patches/task-3.diff`.
  - [ ] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/records/ && go test ./internal/records/ -run 'TestRenderLedger' -count=1` passes.
  - [ ] **Step 4: Commit** `stats/internal/records/render.go` and `stats/internal/records/render_test.go` with the subject below.

**Build:** green
**Files:** `stats/internal/records/render.go`, `stats/internal/records/render_test.go`
**Tests:** `TestRenderLedgerNamesWhenADispatchEnded`
**Regression:** `TestRenderLedgerNamesWhenADispatchEnded` — reverting the commit drops the Ended line, and the test fails on its absence.
**Baseline:** before=34 after=35
<!-- measured: grep -c '^func Test' stats/internal/records/render_test.go @ branch self-review-kan-999-group-session-workout-names, before and after patches/task-3.diff -->
**Commit:** feat(records): render when a dispatch ended in the ledger
**After:** none

**Decision:** ledger-ended-line

- [ ] 4. Document gymie live-check fixtures and the before-figure backend
  - [ ] **Step 1: Apply** `git apply $P/patches/task-4.diff` in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-999-group-session-workout-names`: two paragraphs after the `count-request-queries.sh` paragraph of `## run`.
  - [ ] **Step 2: Verify.** `git diff --stat` names `.flow/project.md` alone, and `git diff | grep -c '^+```'` prints `0` — the new paragraphs add no fenced block to `## run`; `/Users/tweety53/Projects/agents/scripts/project-get.sh /Users/tweety53/Projects/gymie-worktrees/kan-999-group-session-workout-names run` exits 0 and its output carries both new paragraphs. No lint line applies to this file.
  - [ ] **Step 3: Commit** `.flow/project.md` on branch `spectre/kan-999-group-session-workout-names` with the subject below; it rides that change's merge and push.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — project configuration prose.
**Regression:** none — prose.
**Commit:** docs(flow): start gymie live checks from dbSeed and document the before backend
**After:** none

**Decision:** gymie-live-check-paragraphs
