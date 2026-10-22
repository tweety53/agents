# kan-921-flow-improvement-port-the-guard-test-runner-run — session narrative

## 2026-10-22 — creating run

- Inline execution. Harness capture goes to a file per job, not a pipe: a pipe makes `Wait` block on any background grandchild holding the harness's stdout, which `bash -c … > file` never did.
- First signal design used `signal.NotifyContext` plus a second `Notify` channel — redundant; replaced with one channel and a watcher goroutine. Review then moved signal registration above `MkdirTemp` to close the window where a signal leaked the temp dir.
- One flake on the first green run: the durations case asserted `(0s)`, but elapsed is whole Unix seconds end minus start (decision elapsed-unix-seconds), so a harness straddling a second boundary reads 1s. The assertion now accepts `\d+s`; that granularity is the requirement, not a widened tolerance.
- Mutation proofs: reversed duration sort, removed signal cleanup, removed the delta's before-set filter. The last one survived the original 7a case (clean watched repo), so 7a gained a pre-existing untracked file.
- Plan step 2.3 was wrong: the two fixture-copy lists copy `coverage.sh` from git history at `d71a2327` for bash-parity cases, not from the tree; they stay. Recorded as Corrections on tasks 2 and 3, and the absence sweep's allowed output became those two lines.
- `check-verbatim-moves` flagged the deliberate reword in `skills/flow-self-review/SKILL.md` (coverage.sh → coverage call); acknowledged in `verbatim-moves.txt`.
- Panel: two Minors (glob byte order vs LC_COLLATE, documented in code; a duplicate coverage table row, deleted), fixed inline.
