# kan-310-myflow-a-fix-round-s-rebase-autosquash-can

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

- [x] 1. Add the re-verify-by-content rule to Panel re-runs
  - [x] **Step 1: Insert the rule paragraph.** In `skills/flow/review-panel.md`, in the **Panel
    re-runs** section, insert one new paragraph immediately after the existing sentence ending
    "Immediately `git rebase --autosquash` to fold it in." (verified: currently line 354) and
    before the `| Mode | Who re-runs | Diff they get |` table (verified: currently line 356). The
    paragraph's text, verbatim: A clean `git rebase --autosquash` is not evidence the fix survived
    it. Where the fixup and the commit it folds into touch nearby lines, git's 3-way auto-merge can
    resolve in favour of the pre-fix side — it exits 0, prints no conflict marker, and leaves no
    `fixup!` commit behind. The reproducer rerun and diff check below (Once the fix subagent
    reports…) are what catch this; they must run against the post-rebase file content, never be
    satisfied by the fixup commit's presence or the rebase's own exit code. Bold the opening clause
    ("A clean … survived it.") and the "Once the fix subagent reports…" cross-reference to match the
    section's existing bolding convention.

  - [x] **Step 2: Verify placement and rendering.** Run `sed -n '340,360p'
    skills/flow/review-panel.md` and confirm the new paragraph sits between the fixup/rebase
    sentence and the mode table, with correct Markdown (no broken bold markers, no stray
    indentation pulling it into the preceding line's paragraph).

**Files:** `skills/flow/review-panel.md`
**Tests:** none — a documentation-only change; no test suite covers skill prose.
**Regression:** none — no runnable behavior changes.
**Baseline:** before=0 after=0
**Commit:** `docs(review-panel): re-verify fix-round repairs by content after autosquash`
**Build:** green
