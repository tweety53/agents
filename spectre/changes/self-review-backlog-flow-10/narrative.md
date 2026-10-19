# self-review-backlog-flow-10 — session narrative

## 2026-10-19 — planning

- Planned as item 1 of `/flow-self-review` step 3, every question auto-resolved; no flow stage marks, no decision gate.
- The Jira issues could not be read from the planning session; each ticket's finding was read from the self-review report that filed it and its context bundle's panel table in git history.
- KAN-901 is not live: the prepare-archive-branch guard and its tests were deleted with the archive-on-the-change-branch rework. KAN-906 is `big` at 85 files and stays filed. Both are cited in `design.md`; the spares KAN-902, KAN-905 and KAN-907 filled the five.
- KAN-902 was reproduced for `check-verbatim-moves` and for `check-references`, so the fix covers every guard that judges its own checkout, not only the one the ticket names.
- KAN-905 was reproduced in a `--depth 1` clone: 14 guard-package tests failed on history the clone does not carry.
- KAN-907's deterministic proof was already in both tests; a mutation that handles the ignored SIGHUP fails them on every run, so the timed looks were removed rather than replaced.
- Every task's diff was drafted and checked in a scratch clone at `46db35e0`: each new or kept assertion fails under its task's named mutation and passes without it.
