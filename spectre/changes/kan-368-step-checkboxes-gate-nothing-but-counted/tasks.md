# kan-368-step-checkboxes-gate-nothing-but-counted

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

- [x] 1. Anchor `count_unticked` to column-0 task lines and update the header comment
  - [ ] **Step 1: Change the regex.** In `scripts/check-unfinished-work.sh`, in `count_unticked()`,
    change the pattern passed to `count_matching` from `'^[[:space:]]*- \[ \]'` to `'^- \[ \]'` —
    matches only a checkbox at column 0, never one indented beneath it.
  - [ ] **Step 2: Update the header comment.** Rewrite the "WHY THE CHECKLIST PATTERN IS ANCHORED"
    comment block (currently starting `# WHY THE CHECKLIST PATTERN IS ANCHORED. `- [ ]` is matched
    only at the start of a line (leading whitespace allowed, for a nested item).`) to say the
    pattern is anchored to column 0 only, that a step checkbox indented two columns beneath its
    task is therefore never counted (per `skills/flow-contracts/build-green.md`'s task/step
    grammar and `skills/flow/implement.md`'s "a step's checkbox tracks the step and gates
    nothing"), and keep the existing measured 44/42 unanchored-vs-anchored figure only if it still
    describes the column-0 anchor accurately — otherwise drop or requalify that sentence rather
    than leave a stale number.
  - [ ] **Step 3: Run the existing suite.** `bash scripts/test-check-unfinished-work.sh` — every
    existing case must still pass, since no fixture in it uses an indented `- [ ]`.

**Build:** green
**Files:** `scripts/check-unfinished-work.sh`
**Tests:** none — task 2 adds the new regression test; this task only narrows the guard's own
regex and updates its header comment, and is verified against the existing suite.
**Regression:** reverting this task restores the unanchored `^[[:space:]]*- \[ \]` pattern, which
re-introduces the KAN-368 failure: `scripts/test-check-unfinished-work.sh`'s new case (added in
task 2) would then fail, catching the regression.
**Baseline:** before=77 after=77 `ok:` assertions in `scripts/test-check-unfinished-work.sh`, all
still passing — this task changes no fixture and no assertion, only the guard's own regex.
**Commit:** `fix(scripts): anchor check-unfinished-work's checklist match to column 0`

- [x] 2. Add a regression test for an unticked step under a fully-checked task
  - [ ] **Step 1: Write the fixture.** In `scripts/test-check-unfinished-work.sh`, add a case whose
    `tasks.md` fixture is two lines: a checked task line (`- [x] 1. done`) followed by an unchecked
    step line indented two columns beneath it (`  - [ ] **Step 1: still shows unticked, and must
    not count**`). Run the guard against it and assert the verdict is `CLEAR:` — the unticked step
    must not be read as an unchecked plan item.
  - [ ] **Step 2: Confirm it fails without task 1.** Before task 1 lands, this new case fails
    against the pre-fix regex (verdict `OUTSTANDING:` instead of `CLEAR:`) — run it against the
    working tree at task 1's base commit to confirm, then rely on task 1's own commit (already
    landed by the time this task runs) to make it pass.

**Build:** green
**Files:** `scripts/test-check-unfinished-work.sh`
**Tests:** `scripts/test-check-unfinished-work.sh` — one new case: "a checked task with an unticked
step beneath it is CLEAR, not OUTSTANDING."
**Regression:** reverting this task removes the one guard against the exact KAN-368 failure mode —
a future change to `count_unticked` could re-widen the anchor and nothing would catch it.
**Baseline:** before=77 after=78 `ok:` assertions in `scripts/test-check-unfinished-work.sh`, all
passing.
**Commit:** `test(scripts): assert an unticked step does not gate check-unfinished-work`
