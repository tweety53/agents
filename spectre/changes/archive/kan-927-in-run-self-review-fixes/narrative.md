# kan-927-in-run-self-review-fixes — session narrative

## 2026-10-08 — creating run

Resumed at `STARTED` with a ready four-task plan; `small` class, inline execution, one bundled `primary+principles` panel.

- **Plan corrections at task close.** Three tasks' `**Files:**` fields were short: task 1 also had to teach two more `api.StatsStore` fakes (`internal/web/embed_test.go`, `internal/client/client_test.go`) the new method; task 2 had to move `RunDetail.test.tsx`'s pinned `VIEW_NAMES` count from 8 to 9; task 4 also touched `git-boundaries.md`, `flow-contracts/SKILL.md` and `stages/names.go`. Tasks 3 and 4 also declared `verbatim-moves.txt`, a `spectre/changes/` path no task commit may carry — dropped from `**Files:**`; it rides the planning commits.
- **Stale fix shas — three gated-review rounds.** The first design recorded each change-branch fix's sha at the pass. The reviewer showed any later rebase (`/flow-fast`'s §7 rebase, run 1's rejected-push re-sync) leaves those shas outside history. First attempt — re-read shas and amend the report before every push, record rows after "the push" — was rejected: "the push" was ambiguous on merge-and-push, the raw amend bypassed the landing script's guards, and on `/flow-fast` the in-memory findings were gone by the time a later invocation pushed. Final shape: the committed report is the record; the landing push (named per route), in whichever invocation makes it, re-reads shas by subject, commits a rewritten report through `land-self-review-report.sh`, then records `fixed` rows from the report's lines. A squash/rebase PR merge is a stated ceiling.
- **Re-run duplicates.** `/flow-fast` fix re-runs and `/flow`'s archived re-runs both re-ran the pass, duplicating store rows and re-asking the filing question; both now skip the pass once a report is committed. New work on such a re-run gets the review panel, not a second pass.
- **Panel round 0** raised 1 Important (the archived re-run above) and 3 Minors; all fixed inline, re-run clean. Reproducers 1–3 had to be re-authored: their premises cited the very lines the fix changed.
- **Full Go suite, second run:** `reconcile.TestConcurrentAppendVersusRetirePreservesEveryEntry` lost 1 of 600 entries under full-suite load (15/15 green in isolation). The in-run pipeline fix found the cause — `AppendJournalEntry` writes unlocked after a 50ms lock wait, and a macOS `F_FULLFSYNC` retire outlasts it — and fixed it on `fix-journal-retire-race` (c8979d5d) by waiting for the lock unbounded. That reverses a recorded design choice (bounded wait, loss acceptable), so it is deferred to the operator rather than landed.
- **Live check** ran against a worktree flowd on 4693 (its database had to be created with `scripts/workspace.sh create` and the project row seeded; the first start failed on the missing database). The worktree flowd replays the shared state-dir journals by design; both journals on disk were empty, so nothing was consumed. The 4173 figures in design.md wait for the operator's reload.
- **Visual verify** captured only the empty view: the UI-test seed has no `agents-a740d89c` project row, so seeding findings there failed the FK. The verifier also made a test-only edit (`exact: true` on the full-suite nav locator) outside its relay contract.

In-run pipeline fix: deferred — fallback journal append writes unlocked after a 50ms wait and loses entries to a concurrent retire (blast radius 5 files)

## 2026-10-08 — integrate run

- **Preflight:** `RUN1`; main checkout staged-clean and drift-clean; no retired-layout worktree to migrate.
- **Unfinished-work gate:** `CLEAR` — every plan item checked, no open finding; visual verify dispatched (`VISUAL-VERIFY-OK`).
- **Base:** `origin/main` had not moved since the recorded merge base `3b4aa4dc`; no rebase.
- **Route:** merge and push, from the project's configured default.
- **Self-review:** run on the installed (pre-change) integrate, so the pass is deferred to a saved bundle; this change's in-run pass applies from the next integrate on.
