# self-review-kan-941-step-6-playwright-fidelity-specs-for-the-cardio

## Why

Six pipeline defects the kan-941 self-review pass found, none `big`, each reproduced or traced at
`f4844c40` (`design.md`, **Context**):

- F1: `reshape-branch.sh` resets to the recorded merge base even when the change branch merged in
  another branch, folding that branch's commits into the change commit — kan-941's frontend
  `develop` merge and playwright `main` merge landed under kan-941 subjects.
- F2: the unfinished-work gate's signal four judges only the tasks `**Tasks appended:**` counts,
  and the verify-fix loop never raises it, so kan-941's `In-run fix` tasks 12–16 passed with no
  dispatch row and no review.
- F3: `unfinished-work-gate.md` runs the structural-verdict false-positive call once per worktree
  with `-C <worktree>`, but every verdict is recorded under the canonical project, so a satellite's
  call is refused.
- F4: `check-untasked-commits` lists a merged-in branch's commits as UNTASKED.
- F5: `check-plan-shape` accepts a task whose `**Files:**` all lie under the planning paths, whose
  commit then cannot carry a `Task-Id:`.
- F6: the out-of-task commit's `Flow-Key:` trailer and its pre-push check are prose only; four
  kan-941 frontend commits were pushed without it.

## What changes

- `reshape-branch.sh` stops with `RESHAPE-FOREIGN-MERGE` (exit 1) before HEAD moves when a merge
  brought in commits the merge base does not reach.
- The unfinished-work gate, and close-task's warning that shares its predicate, judge every
  `In-run fix <k> — …` task as appended.
- The structural-verdict false-positive call runs once, at the canonical worktree.
- `check-untasked-commits` reads the first-parent chain only.
- `check-plan-shape` refuses a task whose `**Files:**` all lie under the planning paths (F14).
- A push of a change branch, typed or made by `close-task.sh`, is refused while an unpushed commit
  carries neither a `Task-Id:` nor a `Flow-Key:` trailer.
