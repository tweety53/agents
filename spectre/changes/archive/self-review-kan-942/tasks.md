# self-review-kan-942

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** close KAN-942 — the visual verifier captures a view at the merge base before attributing
a departure on it, when the merge base draws the view but carries no baseline PNG for it.

**Architecture:** one paragraph of `skills/flow/visual-verify-verifier.md`'s attribution rule
rewritten; no code, no guard.

**Tech Stack:** Markdown skills.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-942-design.md` in this worktree.

## Global Constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-942`, branch `self-review-kan-942`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- The prose diff in Step 2 is applied with `git apply` from the worktree's root after copying the lines between its fence markers into a scratch file outside the worktree; it was applied and checked at `b76c2cf1`. A diff that does not apply is reported, never hand-merged by guesswork. Each sentence lands character for character: `scripts/check-verbatim-moves.sh` matches the new ones against `verbatim-moves.txt` exactly.

## Review Focus

- Task 1: a view this change creates is still wholly the change's, and a merge-base capture that cannot run still leaves the departure the change's — the fix never turns a missing capture into a pre-existing attribution.
- Task 1: the teardown removes the derived workspace before the worktrees, and nothing the merge-base capture writes reaches a commit or the change's baseline.

- [x] 1. KAN-942 — capture a baseline-less existing view at the merge base before attributing its departures
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Confirm the defect on the
    untouched tree: `git grep -n -F -e "and neither has one the merge base" -- skills/flow/visual-verify-verifier.md`
    prints line 232.
  - [x] **Step 2: Apply the prose diff.**

````diff verified:applied with git apply at b76c2cf1; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py, check-normative-inventory.sh and check-verbatim-moves.sh (with this change's verbatim-moves.txt) exit 0
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index be4e4d2b..325970f4 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -229,9 +229,25 @@ never an exit code the command did not return.
    **The loop** (`skills/flow/verify-fix-loop.md`) per **Steps 3–13**
    (`skills/flow/visual-verify.md`); **corrected or gone at 2x** — a mis-measurement, the
    corrected number standing in its place. A
-   view this change creates has no rendering at the merge base, and neither has one the merge base
-   carries no baseline PNG for: every departure on either is this change's, and the report names
-   the missing baseline. Every line that reports the departure names its cause or its attribution
+   view this change creates has no rendering at the merge base: every departure on it is this
+   change's. **A view the merge base already draws but carries no baseline PNG for is captured at
+   the merge base before any departure on it is attributed — never counted as this change's for
+   want of a baseline** (KAN-942: KAN-758's verifier counted some 39 over-tolerance band pairs as
+   its change's, every one geometry the merge base already shipped or a frame artefact). One
+   capture per spec serves every departure on its views. In each repository the stack and the
+   capture run from, add a detached worktree `<name>-merge-base` beside the change's own, at that
+   repository's merge base (`git worktree add --detach ../<name>-merge-base <merge-base>`); copy
+   this change's capture spec, its `<spec>.mockups` sidecar and its fixtures into the one that
+   holds specs; start the stack there from `start` when declared, else `## run` — on its own
+   derived workspace, never the change's running stack — and run `capture` from there with
+   `<spec>` substituted. Measure the same element on the PNG it writes and judge it as above. Then
+   stop that stack, remove the workspace it derived with the project's declared removal command
+   (**Creation and cleanup**, `skills/flow-contracts/workspace-isolation.md`),
+   and remove those worktrees before the report is written: nothing the merge-base capture writes
+   is committed or stands as this change's baseline. A merge-base capture that cannot run — the
+   spec reaches a route or control the merge base lacks, the stack does not start — leaves the
+   departure this change's, its line naming `merge-base capture failed — <reason>` beside it.
+   Every line that reports the departure names its cause or its attribution
    beside it — never a bare `departure`.
    A departure within 1dp is reported like any other, its measured value beside it: whether it
    stands is the parent's call under the Blocking paragraph of **Steps 3–13**
````

  - [x] **Step 3: Verify.** `git grep -n -F -e "and neither has one the merge base" -- skills/flow/visual-verify-verifier.md`
    prints nothing; `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py`, `scripts/check-normative-inventory.sh` and
    `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — a prose contract change; Step 3's absence sweep and `check-verbatim-moves.sh` cover it.
**Regression:** none — no test is added; reverting the commit restores the sentence Step 1 finds.
**Commit:** docs(flow): capture a baseline-less view at the merge base before attributing its departures
**After:** none

**Decision:** merge-base-capture
**Decision:** merge-base-worktree-name
**Decision:** prose-only
