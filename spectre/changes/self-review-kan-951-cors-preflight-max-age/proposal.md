# self-review-kan-951-cors-preflight-max-age

## Why

The self-review pass over kan-951 found four non-`big` defects, each held against the tree
(`design.md`, **Context**):

- gymie's `## visual verification` `verify` and `capture` rows export `BACKEND_URL` but not
  `FRONTEND_URL`, so on a worktree stack Playwright fell back to the main checkout's frontend port
  and its first full run was refused.
- The parent's full-suite fix paragraph in `skills/flow/implement.md` names none of what a commit
  outside every task owes: kan-951's full-suite fix carried no `Flow-Key:` trailer, its ledger row
  recorded Task 1's commit, and its determinism claim had no run with the fix reverted.
- A fix run re-ran the whole visual verify although its own delta touched only two gateway test
  files, because the trigger is fed the change's whole merge-base diff on every run.
- The panel record lost the creating run's finding: the store upserts findings on (change, ref), the
  third run numbered its finding `F1` again and rewrote the earlier `F1`, and a panel round whose
  parent recorded no pass row (round 1) vanished from the rendered pass log.

## What changes

- gymie's `verify` and `capture` rows set `FRONTEND_URL` from `FRONTEND_PORT`, falling back to the
  main checkout's port, beside `BACKEND_URL`.
- The full-suite fix commit carries `Flow-Key: full-suite-fix-<n>`, is checked by
  `check-untasked-commits.sh` before it is pushed, its row's `end` records its own sha, and a
  determinism fix shows the failure returning with the fix reverted before it is committed.
- On a fix run, Visual verification's trigger reads each worktree's since-close diff — the range the
  late-fix reduction already resolved — so a fix run touching no `ui paths` skips the stage.
- `flow record next-ref -change <name>` prints the first free finding ref; `review-panel.md` opens a
  run's findings there instead of at `F1`. The rendered pass log shows a round a panel bundle ran in
  even when no pass row was recorded for it, naming the dispatch keys.
