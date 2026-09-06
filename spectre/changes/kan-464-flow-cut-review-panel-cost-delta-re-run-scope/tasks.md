# kan-464-flow-cut-review-panel-cost-delta-re-run-scope

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — task 3 moves the operational sentences of **The fix round mutation-proves
> what it changed** out of that section and into the MUTATION PROOF blockquote the fix subagent's
> dispatch carries.

Every task edits prose whose exact replacement wording is `design.md`'s; this file names the site,
the order and the checks, and never restates the wording.

- [x] 1. Narrow which slots re-run after a fix round

**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — the change is contract prose; no guard in this repository parses **Panel
re-runs**' bullets, and inventing one to assert three sentences exist would test the fixture, not
the pipeline.
**Regression:** reverting this commit restores "every diff-reading slot … re-runs on its delta",
which is the cost KAN-464 exists to remove; `scripts/check-normative-inventory.sh` shows the
sentence returning.
**Baseline:** before=56 after=56
<!-- measured: bash scripts/test-check-dispatch-paragraphs.sh | grep -c '^ok:' @ branch spectre/kan-464-flow-cut-review-panel-cost-delta-re-run-scope -->
**Commit:** `docs(flow): re-run only the panel slots that raised a finding`
**Build:** green

  - [x] **Step 1: replace the three bullets under "When the round raised anything above Minor,
    re-run on deltas".** Write the two-bullet rule and the unchanged third bullet exactly as
    `design.md`'s Part 1 gives them. The paragraph's opener, its per-slot-per-worktree held-sha
    rule and its delta-path definition are not touched.
  - [x] **Step 2: widen the stale-result carve-out.** Replace the sentence naming Bugbot and
    Mutation with the any-slot form `design.md`'s Part 1 gives. Do this in the same commit as
    step 1: on its own, step 1 leaves the handoff blocking on every slot the new rule declines to
    re-run.
  - [x] **Step 3: confirm nothing else moved.** `git diff` shows edits only inside **Panel
    re-runs**; the Minor-only rule, the docs-only reclassify clause, `-diff-base` and the re-run
    cap check are untouched.
  - [x] **Step 4: verify.** `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
    `scripts/check-markdown-integrity.py` and `scripts/check-contract-budget.sh` exit clean.

- [x] 2. Let a reviewer slot write its reproducer as a script

**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — `check-panel-reproducers.sh` and `run-reproducer.sh` already accept the
shape and are not edited; their own harnesses cover it, and this task adds no code to test.
**Regression:** reverting this commit removes the script form from every slot's dispatch prompt,
returning the panel to the inline-command-only pattern under which 22 of 33 findings recorded
`none — …`.
**Baseline:** before=56 after=56
<!-- measured: bash scripts/test-check-dispatch-paragraphs.sh | grep -c '^ok:' @ branch spectre/kan-464-flow-cut-review-panel-cost-delta-re-run-scope -->
**Commit:** `docs(flow): let a panel slot record a reproducer script it wrote`
**Build:** green

  - [x] **Step 1: extend the "Every slot must supply, per finding, a reproducer" paragraph** with
    the five points `design.md`'s Part 3 lists, keeping the paragraph's two existing forms and its
    closing carry-it-on-every-dispatch instruction. Add no blockquote: this is the one required
    instruction that rides on an existing paragraph rather than a labelled one, so it needs no row
    in `scripts/check-dispatch-paragraphs.sh`.
  - [x] **Step 2: check the path against the runner's real rules.** Confirm by reading
    `scripts/run-reproducer.sh` that the recorded form must be worktree-relative, must resolve to a
    regular file inside the worktree, and must carry execute permission — the three constraints the
    new prose states — and that `scripts/check-panel-reproducers.sh` needs no edit to accept it.
  - [x] **Step 3: verify.** `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
    `scripts/check-markdown-integrity.py` and `scripts/check-contract-budget.sh` exit clean.

- [x] 3. Move the fix round's mutation-proof onto the fix subagent

**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — task 4 adds the harness cases that hold this task's new paragraph in place;
splitting them here would test a guard row that does not exist yet.
**Regression:** reverting this commit puts the mutation-proof back on the conductor ("*you* then
mutate each one") and removes the plan-field check from the parent's hunk walk, which — with task
1 already landed — would leave a stale `tasks.md` field caught by nothing.
**Baseline:** before=56 after=56
<!-- measured: bash scripts/test-check-dispatch-paragraphs.sh | grep -c '^ok:' @ branch spectre/kan-464-flow-cut-review-panel-cost-delta-re-run-scope -->
**Commit:** `docs(flow): hand the fix round's mutation-proof to the fix subagent`
**Build:** green

  - [x] **Step 1: rewrite the section body of "The fix round mutation-proves what it changed"** per
    `design.md`'s Part 2, keeping its two opening lines, its fenced `fix-mutation:` block, the
    "never inside the marker block" line, the marker-label ban paragraph and its closing
    "This binds the fix round every run" sentence verbatim.
  - [x] **Step 2: add the MUTATION PROOF blockquote** to the fix subagent's dispatch paragraphs,
    beside PLAN FIELDS, FOREGROUND BUILDS, TARGETED TESTS and REPORT FILE, with the wording
    `design.md`'s Part 2 gives. The relocated sentences from step 1 live here and nowhere else.
  - [x] **Step 3: extend the fix subagent's REPORT FILE paragraph** to name the `fix-mutation:` and
    `fix-mutations-total:` lines alongside what it already names.
  - [x] **Step 4: add the plan-field sentence to the parent's hunk walk**, per `design.md`'s Part
    2. The walk's existing clauses, including the removed-or-weakened-test clause, are untouched.
  - [x] **Step 5: confirm the relocation is a move, not a cut.** Every operational sentence removed
    from the section in step 1 appears in the blockquote from step 2 —
    `scripts/check-normative-inventory.sh` before and after the commit differs only where a
    sentence's subject changed from the parent to the fix subagent, never by a sentence
    disappearing.
  - [x] **Step 6: verify.** `scripts/check-dispatch-paragraphs.sh` still exits clean (it has no
    MUTATION PROOF row yet, so the new block is simply unchecked), and
    `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
    `scripts/check-markdown-integrity.py` and `scripts/check-contract-budget.sh` exit clean.

- [x] 4. Guard the MUTATION PROOF paragraph

**Files:** `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
**Tests:** Case 27: the MUTATION PROOF label absent from `review-panel.md` exits 1 and names the
file and the missing block; Case 28, Case 29 and Case 30: one required phrase dropped in turn, each
exiting 1 and naming the missing phrase.
**Regression:** reverting this commit removes the guard row, and task 3's paragraph can then be
trimmed one line at a time with every check green — the KAN-289 failure mode this guard exists to
close.
**Baseline:** before=56 after=64
<!-- measured: bash scripts/test-check-dispatch-paragraphs.sh | grep -c '^ok:' @ branch spectre/kan-464-flow-cut-review-panel-cost-delta-re-run-scope -->
<!-- predicted: 4 new cases at 2 assertions each, the rate cases 22-26 already run at -->
**Commit:** `feat(scripts): guard the MUTATION PROOF dispatch paragraph`
**Build:** green

  - [x] **Step 1 (RED): add the four cases to `scripts/test-check-dispatch-paragraphs.sh`.** Define
    `MUTATION_PROOF_BLOCK` from the blockquote task 3 wrote, plus three variants each dropping one
    required phrase; add case 27 (label absent from `review-panel.md`) and cases 28-30 (one phrase
    dropped each), following the shape cases 22-26 use for TARGETED TESTS. Run the harness: the
    four new cases fail, because the guard has no row for the label yet.
  - [x] **Step 2 (GREEN): add the table row to `scripts/check-dispatch-paragraphs.sh`** — one entry
    each in `ENTRY_LABEL`, `ENTRY_SHARED_PHRASES` and `ENTRY_VARIANTS`, and one element appended to
    each of `SITE_ENTRY`, `SITE_PATHS`, `SITE_MIN_BLOCKS` and `SITE_VARIANTS`, with the values
    `design.md`'s Part 2 table gives. Touch none of the guard's machinery.
  - [x] **Step 3: add the block to every `review-panel.md` fixture the harness asserts clean**, and
    only those — a fixture written deliberately deficient for another label keeps its current
    shape, exactly as the TARGETED TESTS cases left theirs.
  - [x] **Step 4: update the guard's header.** Add the row to its paragraph table and the entry's
    own shared-phrase note, and record in `scripts/test-check-dispatch-paragraphs.sh`'s header what
    cases 27-30 cover, in the shape the KAN-441 note above them already uses.
  - [x] **Step 5: mutation-test the new cases by hand**, per the practice that harness's header
    records: remove the new row from a throwaway copy of the guard, re-run the four cases, and
    confirm each one's failure signal disappears — proving the case depends on the row rather than
    on another check catching the same fixture.
  - [x] **Step 6: verify.** `bash scripts/test-check-dispatch-paragraphs.sh` prints
    `all cases passed`, `scripts/check-dispatch-paragraphs.sh` exits clean against the real tree,
    and `scripts/run-guard-tests.sh` is green.
