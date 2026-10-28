# kan-1041-weekly-dashboard-model-effort-correlation — session narrative

## 2026-10-28 — creating run

- Tasks 1–9 ran inline. The first panel raised F1–F16. The deciding Important was F1: reprice could not finish its 56-day window inside the CLI's 2s default timeout or flowd's 30s write timeout. The fix clears the handler's write deadline and gives the CLI a 30m timeout. The mutation slot's surviving mutants (the served-model tie, the effort order) were pinned with fixtures that make the SQL's `ORDER BY v DESC, k` matter.
- The live check ran against a scratch `pg_dump` copy of the dev `flow` database (`flow_kan1041_live`) behind a scratch flowd on port 6803, so the dev stack and its data were never touched. It found two defects that no test had seen:
  1. Reprice aborted with "cannot scan NULL into *int64". `GetStageRun` read an unattached `plan.session` run's NULL `change_id` (In-run fix 1).
  2. Per-week model·effort sums exceeded Overall by up to 65% (2026-09-30: $1551.51 against $939.78). First hypothesis, abandoned: that the series counted runs with an unpriceable bucket. Filtering on a top-level `cost_usd` alone did not close the gap. The real cause: `Price` never deletes, so a total written before a no-rate `claude-sonnet-5-5` bucket joined the run stayed short while reading as complete. The fix drops the total whenever a bucket misses, and the series counts only runs carrying one (In-run fix 2). After it, all 9 weeks matched exactly.
- The dev store has no pricing row for `claude-sonnet-5-5`. Every run touching that model therefore carries no total and drops out of both charts until the operator adds the rate.
- The first visual verifier compared against the approved prototype. It raised four departures this change introduced, plus five present at the merge base:
  - Fixed: header copy and pill width (In-run fix 3). The pill width is the prototype's fixed 40/84, widened only when a real figure would overflow. A strict 84 would clip `$2,253.34 ▲12%`. The first width estimate of 6px a character made every cost-alone pill 42 wide, so the prototype's 40 never applied; 5.5px matches the prototype's own fit.
  - Kept as the app's shared design-system convention (`--series-*` tokens, the `Panel` component other screens draw), not restyled to the standalone preview's shell: palette, panel caption, layout, axis label format, Overall-only default.
- The full guard suite failed once under `-race`. `runGuard`'s merged buffer was written by both of a child process's stream-copy goroutines (`workspace`'s `docker exec`). It did not reproduce under load in two full-package runs. A forced reproduction (a guard whose child floods both streams) failed reliably without the lock (In-run fix 4).
- Tooling friction:
  - `close-task.sh` cannot rebuild the full-app-suite zip unless `FLOW_GUARD_SCRIPT_DIR` is set: the Go `close-task` execs `rebuild-full-app-suite-zip` without exporting it. The workaround was to set it by hand.
  - A sonnet mutation slot wrote both its reports in one Bash call through two heredocs, and `protect-main-checkout.py` denied it: each body's rest still carried the other body, whose lone `.` and unpaired backtick failed closed. The slot handed back claiming both written; the parent's `test -s` caught the gap and re-asked.
  - The verifier hook needs the dispatch `description` to open with the row key.

## 2026-10-28 — fix run

- The in-run fix rounds 1–4 above opened document-fix stages, which the store marks as fix runs. Their account is the creating run's section above.

## 2026-10-28 — integrate run

- Chained straight from the write-in-progress stage (no review-before-landing). `prepare` passed the preflight, the unfinished-work gate and the base check, and the branch landed by the project's default route.
