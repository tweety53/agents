No syntax issues, no bashisms beyond 3.2-safe constructs (`local`, `printf`, string params — all fine in bash 3.2). Verified everything in the delta: sourcing order correct in both harnesses, no name collisions, usage text byte-identical to each guard's real heredoc output (confirmed by running both suites live), no unescaped `%`, and each harness asserts its own guard's name. Real test execution (not just static reading) confirms the trailing-newline stripping question is a non-issue since `$(...)` strips trailing newlines on both sides regardless.

No edits made to the worktree; nothing to restore.

**PASS**
