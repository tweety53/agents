# SDD ledger — kan-182-duplicate-change-row-from-best-guess-mark

Command: `/myflow-fast` (creating run). Harness: claude-code. Session token: `mf-kan182-a1`.
Recorded models (myflow-fast defaults, unmodified by the operator): implementation `sonnet`,
review panel `sonnet`, panel fixes `sonnet`. Review panel roster: `light`.

Worktree: `/Users/tweety53/Projects/agents-kan-182` · branch
`openspec/kan-182-duplicate-change-row-from-best-guess-mark` · merge base `c07fed7` ·
workspace id `kan-182-ac74`.

## Dispatches

Task 1: complete (commit `8f2c186`, review clean after one fix round, model: sonnet, review:
combined). Implementer dispatched on sonnet; combined per-task reviewer dispatched on sonnet.
Original commit `9ac8723`; the reviewer's finding F1 (Major — the change-argument extraction did not
trim trailing whitespace, so a guessed placeholder followed by a space evaded the guard) was fixed by
the same implementer and folded in via `git commit --fixup=9ac8723` + `git rebase --autosquash`.
`scripts/check-task-commit-fields.sh` clean against the folded commit.

Task 2: complete (commit `dfd9827`, review clean, model: sonnet, review: combined). Implementer on
sonnet; combined per-task reviewer on sonnet. Commit `7da9551` before task 1's autosquash rewrote it;
content unchanged by the rebase. `check-task-commit-fields.sh` initially failed on this task's
declared `Tests:` field — the parent repaired the plan (the field named existing guard scripts, which
the guard reads as tests that must appear in the diff) rather than sending the task back, since
implementers never edit `openspec/`. Clean after the repair.

Task 3: complete (commit `1852ddf`, model: sonnet, review: combined — reviewer dispatched on sonnet).
`skills/myflow-contracts/pipeline.md` 41042 → 41838 bytes against a 44574 budget.
`check-task-commit-fields.sh` clean.

Task 4: complete (no commit — this task changes no repository file, per its own plan entry). Run by
the parent rather than dispatched: it is one irreversible delete against the shared development
database, and the plan's own steps are the whole of it.

- `psql` is not on this machine's PATH; the statements were run through
  `docker exec myflow-postgres psql -U myflow -d myflow`, against the shared `myflow` database on
  `localhost:5433` — not any worktree's isolated database.
- Step 1 fixed the target by id: `changes.id = 6` is `gymie-7c1f238a/kan-175` (1 stage run, 0
  change_repos rows); `id = 7` is `kan-175-more-ui-ux-fixes` (8 stage runs), which stays.
- Step 2, one transaction, children first: `DELETE 1` from `stage_runs`, `DELETE 0` from
  `change_repos`, `DELETE 1` from `changes`, `COMMIT`.
- Step 3 verified from both surfaces: the SELECT returns one row (`kan-175-more-ui-ux-fixes`), and
  the state-board API returns one `kan-175-*` row for that project.

## Notes carried to the handoff

- The plan's task 1 `**Baseline:**` predicted 19 → 23 harness assertions; the real figure is 19 → 27
  (25 after the first commit, 27 after F1's fix). The prediction was left in place rather than
  rewritten after the fact — it is tagged `predicted:`, and correcting it retroactively would erase
  the miss.
