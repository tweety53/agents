# kan-1047-review-dynamics-heatmap-by-project — session narrative

## 2026-10-28 — creating run

- The resumed implementation run found task 1 already written as uncommitted work (the prototype built during planning). It was stashed so the base rebase could run and the inline row could begin, then restored, verified (targeted Go and vitest tests, regression proven by blanking `classFilterSQL`) and committed.
- The operator added three requests mid-run, each appended as its own task and re-decided (class stayed `small`): two charts per row plus a Prompt changes bar (`flow-prompt-tuning` default · `All` · `auto prompt tuning`, mapped to marker kinds `tuning` / both / `edit`); then dropping the "Since last prompt change" range preset. The kind mapping for "auto prompt tuning" was an auto-resolved reading — `edit` markers are recorded automatically from commits.
- Panel round 0 raised three Importants, all on task 1's test: the Projects list's filter-independence, reviewer-only role and `< $2` bound were unpinned. Fixed test-only by seeding three more projects. Two mutation reproducers were broken as authored (an EXIT trap restoring a relative path after `cd stats`, which left a mutant in the tree; single quotes inside a single-quoted sed) and one used the inverted exit convention; the parent repaired them before the dispatch-time runs.
- The parent forgot to record the fix-mutation rows before launching the round-1 re-review; `flow record mutation` then refused them forever and `check-panel-findings-closed` refused the panel close — a deadlock. Fixed in-run as pipeline-fix 40 (`flow record mutation -late '<reason>'`), landed on main, rows recorded late with the reason.
- A re-review slot cited absolute paths in its verdict lines; one `-cite` send-back fixed them.
- Visual verification took three in-run fix rounds: stale spec assertions (old 8-week default, an edit marker the new default hides) and no capture of the new bar states; then no Projects/Task-size pick (the UI-test seed had no decisions — it now records big/small classes for `uitest-dynamics`); then a duplicated `1` y-axis label on a small integer chart (blanked at the shared `DynamicsChart` source — a first attempt blanked the wrong tick of the pair) and a pointer left on a button before a screenshot.
- No `specs` command is declared, so each post-fix verifier ran the closing unscoped `-full` run directly.
- Live check: `scripts/live-store-copy.sh` refused while an earlier copy's state dir existed (`down` first); the API needs both `from` and `to`. Before figure taken on the dev flowd at main (merge base plus a cmd/flow-only fix).

## 2026-10-28 — integrate run

- Chained from the implementation run (no review-before-landing). `prepare` passed the unfinished-work and visual gates first time; the base had not moved since the round-3 rebase onto the landed pipeline fix.

In-run pipeline fix: 69938fe7946f3b09fb96e63d15f2fb4fe293eec1 — a missed fix-mutation row could never be recorded once its round re-review began, deadlocking the panel close (blast radius 3 files)

## 2026-10-28 — fix run 1

- The self-review pass of the second integrate run found a product defect (the sweep): Review dynamics y-axis labels named rounded values other than the ones their ticks sat at — peak 3 drew `2` at 1.5, an empty chart left its axis line unlabelled, and Task 6's blank-label fix had locked the mislabel into a test. Appended as Task 7 (archived change, append taken); re-decided `small`→`regular`, same roster.
- Implemented inline: ticks `0, s, 2s` on a 1·2·5 step. A first rule (whole steps for whole-number data) put a 200% tick on a 0/1 share chart and was replaced, before review, by a floor read off the format.
- Panel round 6 (small append: primary + the two extras): primary F20 (Important) — the half-only floor still let `toFixed(1)` at peak 0.06 read `0.0, 0.1, 0.1`; mutation F21–F24 — four surviving mutants. Fixed by picking the smallest step whose three labels the format tells apart; four flips mutation-proved, rows recorded `-late` (proved before the re-review, recorded after it). Round-6's mutation slot left a live mutant (`2 * s * 1.12` → `2 * s`) in the canonical worktree instead of a throwaway copy; the parent restored it before fixing.
- Round 7 re-review (sonnet): clean of Critical/Important; Minors F25 (an all-zero share chart topped at 200%) and F26 (dead `+ 1e-9`) fixed inline.
- Visual verify `visual-verify-fix-4-full` at 83059f85: no departures; every y label sits on its gridline.

## 2026-10-28 — integrate run

- Second integrate run (bare `/flow kan-1047`): `prepare` found the change archived without new work, gates green, base unmoved, route merge and push from the project default. The self-review pass stopped on the product defect above and the run went into fix run 1.
