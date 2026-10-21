# kan-918-flow-improvement-port-the-git-snapshot-and — session narrative

## 2026-10-21 — creating run

- Resumed at implementation; the plan was ready. Four implementer groups ran as recorded: tasks
  1+3 and task 5 as one wave in throwaway copies, tasks 2+4 queued into the second slot, then 6+7
  in the canonical worktree beside the 2+4 gated reviewer.
- Task 3's record declared `TestFoldFixup` under **Tests:**, but the commit only ran that existing
  test unchanged, so `check-task-commit-fields.sh` refused it. The record was corrected in place
  (moved to **Regression:**). A panel fixer later widened task 3's **Files:**/**Tests:** with tests
  that landed in Flow-Key fix commits, which `check-task-records.sh` refused at verify; the widening
  was reverted and recorded as a Correction, because fix-commit tests are never a task's declared
  set.
- Three header/body disagreements surfaced and were each settled on the header: guard-autosquash
  `after` on an unreadable regular tasks.md (the bash swept no tokens and passed; now exit 2), and
  check-panel-diff-size on a failing `git diff` (the bash exited with git's own code, so a git exit
  1 read as over cap; now exit 2). Both are recorded as `header-wins-on-exit-code` in design.md. A
  third gated-review Minor reordered check-tree-markers' snapshot-path checks ahead of the grep
  loop, as the bash ran them.
- The pins audit's Go scanner read any quoted `"diff"` argument as a git call, so the first port
  dodged it through a `const diffTool`. The scanner was fixed instead; the first scanner fix let an
  identifier argument precede the literal, which let `exec.Command(gitBin, "diff", …)` slip through
  — the mutation slot caught it and the regex was narrowed to `lookPath`.
- The mutation slot's pass 1 found eleven surviving mutants across the five ports (untested
  refusal and boundary paths); two fix chunks pinned them all with test-only changes, plus the
  Minors. A round-1 mutation re-run then caught that the batch `cat-file --batch-check` rewrite
  (a call-efficiency Minor) dropped no test for `^{commit}` peeling of an annotated tag; round 2
  pinned it.
- Two LC_ALL=C mutants were withdrawn: `git status --porcelain` v1 and diff(1)'s normal format
  carry no locale-dependent text for these inputs, so no fixture can observe the variable.
- The primary+principles re-run slot ended without writing its report file; the parent wrote it
  from the slot's final message. The mutation re-run's verdict lines cited no fix-diff hunks and
  took one `-cite` send-back.
- Task 7's after-suite median (391s) is above the before median (314s). The critical path is
  `test-integrate-change.sh`, untouched here, in all six runs; read as machine load, not measured.
- At verify, `run-guard-tests.sh` failed once on its watched-tree check — the live check ran in the
  worktree concurrently — and passed on its one re-run (33/33, 272s).
- A shell gotcha cost a retry: zsh passes `${7:+-task $7}` as one word, so a findings-record loop
  failed until it moved to a bash script with an array.

## 2026-10-21 — integrate run

Preflight clean: nothing foreign staged, no drift on main, the branch not yet merged. The
unfinished-work gate cleared — every plan item checked, no open finding, 22 commits traced to
tasks, no UI paths touched. origin/main had not moved since the recorded merge base, so no rebase
ran. Landing by the project's configured default route, merge and push.
