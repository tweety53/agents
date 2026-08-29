Working tree clean except the pre-existing untracked spectre directory. No production code was touched by the fix (docs-only diff, per commit titles), and no defects found.

**Mutation table**

| # | Target | Mutation | Caught? |
|---|--------|----------|---------|
| 1a | check-finish-preflight.sh | drop `>&2` from heredoc (F1) | Caught — stdout-not-empty and stderr-mismatch both fail |
| 1b | check-base-moved.sh | drop `>&2` from heredoc (F1) | Caught — same as above |
| 2a | check-finish-preflight.sh | delete `exit 2` after heredoc (F2) | Caught — stderr exact-match fails (fallthrough diagnostic appended); stdout-empty alone would NOT catch it |
| 2b | check-base-moved.sh | delete `exit 2` after heredoc (F2) | Caught — same as above |
| 3 | check-finish-preflight.sh | reword usage body line, keep grepped phrase intact | Caught — stderr exact-match fails; the `grep -qF` structural check alone would NOT have caught it |
| 4 | check-finish-preflight.sh | indentation-only change in usage body | Caught — stderr exact-match fails |
| 5 | check-finish-preflight.sh | write usage message to both stdout and stderr (`tee /dev/stderr`) | Caught — stdout-not-empty fails |
| 6a | test file | delete only the stdout assertion, then apply F1 mutant | Still caught — stderr exact-match alone catches F1 |
| 6b | test file | delete only the stderr assertion, then apply F2 mutant | **Not caught** — stdout-empty alone is insufficient for F2; confirms stderr assertion, not stdout assertion, is what actually closes F2 (both assertions present in the real fix, so no live gap) |

**F1: closed.** Both guards' new tests fail when `>&2` is dropped or when the message is duplicated to stdout.

**F2: closed.** Both guards' new tests fail when the explicit `exit 2` is deleted, via the stderr exact-match (not via the stdout-empty check, which is redundant for this case but not harmful).

No new defects found by reading or by mutation. No production code changed in this round (docs-only). Both harnesses pass cleanly; worktree restored to a clean state (verified via `git status --porcelain` and `git diff --stat` after every mutation and at the end).

**Verdict: PASS**
