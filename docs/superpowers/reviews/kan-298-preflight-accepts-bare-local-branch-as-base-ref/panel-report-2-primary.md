Working tree is now clean and matches HEAD exactly (verified via `git diff --stat` empty), except the pre-existing untracked `spectre/changes/...` directory. Both stray mutations were found already present in the worktree before I touched any file — not introduced by me — and have been restored to match `HEAD` byte-for-byte.

## Findings

**Important** — `spectre/changes/kan-298-preflight-accepts-bare-local-branch-as-base-ref/tasks.md:9,63` — Task 1's and task 2's `**Files:**` fields (`scripts/check-finish-preflight.sh`, `scripts/test-check-finish-preflight.sh` / `scripts/check-base-moved.sh`, `scripts/test-check-base-moved.sh`) omit `scripts/lib/base-ref-usage.sh`, which task 1's commit (`01a6d144`) actually created and task 2's harness sources. The plan is out of step with the branch and needs correcting before handoff — add the new lib file to task 1's `Files:` list (it's the file's only creator) and note it as a sourced dependency in task 2's entry.
Reproducer: `none — static comparison of tasks.md's Files fields against `git show --stat 01a6d144dded443dc88fc01bfb683a6d0bb3b010`.

**Important (workspace hygiene, pre-existing, not from this reviewer)** — the worktree, before this review touched anything, carried two uncommitted stray mutations left over from fix-round mutation testing: `scripts/check-base-moved.sh` had an extra space inserted before `(origin/main)` on the usage-message continuation line, and `scripts/lib/base-ref-usage.sh` had its last line's text replaced with `STALE MUTATED LIB`. Both are consistent with task step 5's "mutate, verify failure, restore" instruction never having been completed — the restore step was skipped. Both were restored to match `HEAD` (`git checkout -- scripts/check-base-moved.sh scripts/lib/base-ref-usage.sh`); `git diff --stat` against HEAD is now empty. Flag this to the operator: whoever ran round 2's mutation tests should re-verify their local checkout is clean before any future push.

## Verified clean

- **Non-tautology (item 2):** confirmed by direct mutation in an isolated worktree at `cc186b6b` — appending text to `check-finish-preflight.sh`'s usage message causes `test-check-finish-preflight.sh` to fail with `stderr does not match the usage message exactly`, proving `base_ref_usage_message` is never derived from the guard. Reproducer: `scripts/test-check-finish-preflight.sh` (run against a build where the guard's usage text has been hand-edited).
- **`scripts/lib/` precedent (item 3):** real — `scripts/lib/test-git-shim.sh` already documents itself as test-only, consumed only by the same two harnesses, with the identical "nothing scopes `scripts/lib/` to production-only members" reasoning the new file's header echoes almost verbatim.
- **Bash 3.2 / conventions (item 4):** `local`, `printf`, no arrays/`[[`/associative maps — compatible. Header comment format matches sibling `scripts/lib/*.sh` files.
- **Commits (item 5):** both isolated at their own SHA are independently green — `test-check-finish-preflight.sh` (45 ok lines) and `test-check-base-moved.sh` (all pass) both ran clean in detached worktrees at `01a6d144` and `cc186b6b` respectively. `Task-Id: 1`/`Task-Id: 2` trailers and both commit subjects match `tasks.md`'s `**Commit:**` fields exactly — no drift across the two rebases.

## Verdict

FINDINGS
