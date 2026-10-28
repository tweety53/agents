# kan-1048-rework-live-board-and-run-view — session narrative

## 2026-10-28 — creating run

- **Resumed at STARTED, with the plan ready.** The base was rebased before task 1. All five tasks ran inline. Each task's gate fired on size, so the panel did the per-task review.
- **The full suite failed in the panel's first pass because the UI-test stack's sweeper closed the seed's open runs.** flowd's 6h silence sweeper closes every open stage run started before now minus 6h. The fixture's runs are dated 2026-08-15, so they were "abandoned" about a minute after `ui-test-up`, and the board drew no running cards. It first looked like a stale stack, because a standalone run straight after `ui-test-up` passed. The cause showed only in the `stage_runs` rows: `outcome = abandoned`, ended at the restart time. Fixed with `FLOW_SWEEP_OFF=1` on the UI-test daemon.
- **Tooling defects found and fixed on `main` mid-run.**
  - `run-capped.sh` was never installed beside `run-verify-list.sh` (6557ccc1).
  - `check-verify-green` did not treat a `verify-<k>` re-run as superseding the blocked `verify` (499b813a).
  - The run-capped probes flaked because idle zero-filled pages are compressed out of RSS on macOS within about 5s (7b10e569).
  - Abandoned approach: the first theory was memory compression under the suite's load. I pushed an urandom allocator for it (90a1f2ff) before finding that cc15d1dc had already fixed a different cause, the nested `RUN_CAPPED_PGID` exec, and reverted it (6a53dd04). The compression mechanism turned out to be real after all. Sampled RSS went 145 MB → 36 MB → 8 MB after about 5s. It needed a resident, re-touched allocation, not just random bytes.
- **The base moved twice during the run (16 commits, then 4).** Both rebases were clean. The live check's before-figure stands across the second rebase, because the only `stats/` changes in between were in `internal/guard`.
- **Visual verification found two prototype departures,** fixed as in-run fix 1:
  - Cost in period summed only the filtered rows.
  - The timeline tooltip lingered one paint.
  - The tooltip lag never reproduced under a real pointer. React 19 flushes the discrete mouseleave before paint, so its new spec passes with or without `flushSync`. The fix went straight to code, with that recorded.
- **close-task refused the first task-6 commit** because it carried the `visual-verification/` evidence directory, which is a planning path. `visual-verification.md` itself is exempt. The evidence rides a planning commit instead.
- **A mutation reproducer quoting bug.** A mutation reproducer escaped `\$2` inside a quoted heredoc. Its edit never applied, so it read as a surviving mutant. I re-authored and proved it both ways.

## 2026-10-29 — fix run

- **In-run fix 1, after `flow.visual-verify`.** It appended task 6, which makes Cost in period sum every row and hides the timeline tooltip under `flushSync`.
- **Re-review and re-verify were clean.** The late-fix re-review (primary plus both extra bundles) found nothing. Verify re-ran green once `main`'s run-capped probe fix was rebased in. The unscoped `visual-verify-fix-1-full` run confirmed both departures fixed, with no new ones.
