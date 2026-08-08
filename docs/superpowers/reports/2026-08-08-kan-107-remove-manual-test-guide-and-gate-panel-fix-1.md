# Panel fix round 1 — report

All nine findings (F1–F9) fixed in one pass. Worktree:
`/Users/tweety53/Projects/agents/.worktrees/kan-107-remove-manual-test-guide-and-gate`.

## Files changed

- `scripts/check-unfinished-work.sh` — F1: `ids_of`'s comment now says "see signal two" (line 176). Grepped the whole file for other stale signal-number references; none found (the remaining "signal two" at line 397 was already correct).
- `scripts/gather-self-review-context.sh` — F2: header's source-list comment (line ~101) now says `"plan and session records"` and states both wordings are matched, agreeing with the body comment near line 448 and both greps. F9: `IMPL_SHA`'s exclusion grep now anchors the plan-commit branch the same way the archive branch already is (`^[0-9a-f]+ chore\(${NAME_RE}\): plan(, test guide and| and) session records`), so it no longer matches the phrase as an unanchored substring against an implementation commit's subject.
- `.gitignore` — untouched (F3's `.worktrees/` entry is correct and stays).
- `skills/myflow-contracts/finish-contract.md` — F3: corrected the now-false example sentence about this repo's worktree layout; kept the "never guess a path" rule intact, only the example changed.
- `openspec/changes/kan-107-remove-manual-test-guide-and-gate/design.md` — F3: added new Decision entry `worktrees-dir-gitignored` recording the `.gitignore` deviation and the considered-and-rejected sibling-directory alternative. (Untracked/unstaged, as required — this path is never `git add`ed.)
- `skills/myflow-contracts/handoff-blocks.md` — F4: added a pointer sentence citing `skills/myflow-do/SKILL.md`'s "6. Resolve the run instructions" as canonical for both commands.
- `skills/myflow-status/SKILL.md` — F4: added a bullet in the detail-view step (§4) citing the same section, stating `/myflow-status` resolves `Run it:` from the worktree and project configuration exactly as `/myflow-do` does, without restating the rules. `scripts/check-references.sh` passes.
- `skills/myflow-do/SKILL-rationale.md` — F8: heading corrected to "Why the two planning paths are left unstaged...".
- `scripts/test-gather-self-review-context.sh` — F9: added a case proving an implementation commit whose subject contains the plan-commit phrase still resolves as `IMPL_SHA`.
- `scripts/test-check-unfinished-work.sh` — F5: restored an equivalent read-failure case, targeted at `tasks.md` (not the panel record, to isolate `count_matching`'s own discipline from `ids_of`'s separate guard). F6: added a combined-signal case (unticked plan item + open finding) asserting the one `OUTSTANDING` line names both. F7: case 8f now plants a real `tasks.md` and panel record under `$PLANTED/openspec/changes/clear` and passes that exact path as the name.
- `.superpowers/sdd/final-review-panel.md` — all nine `finding-status:` markers flipped `open` → `fixed`; `findings-total: 9` unchanged.

## F5 mutation proof (both runs)

1. Mutated `count_matching`'s guard at line 168 (`if [ "$rc" -gt 1 ]` → `if false && [ "$rc" -gt 1 ]`), which swallows a read failure. Ran `./scripts/test-check-unfinished-work.sh`: **3 case(s) failed** — the new "unreadable tasks.md" case failed exactly as expected (`rc=0`, verdict `CLEAR:...`, no stderr message) — proving the guard is load-bearing.
2. Restored the file from a pre-mutation backup (`diff` confirmed byte-identical to the pre-mutation version, including the F1 fix at line 176). Ran the suite again: **`check-unfinished-work: all cases pass`**.

## Verification — full `## test` and `## lint` lists from `.myflow/project.md`

All twelve `## test` scripts: exit 0 (`test-setup.sh`, `test-check-references.sh`, `test-check-plan-provenance.sh`, `test-check-finish-preflight.sh`, `test-preserve-session-records.sh`, `test-check-unfinished-work.sh`, `test-check-cleanup-complete.sh`, `test-gather-self-review-context.sh`, `test-uncommitted-review-package.sh`, `test-check-task-build-green.sh`, `test-check-workspace-isolation.sh`, `test-check-contract-budget.sh`).

All six `## lint` guards: exit 0 —
`check-vocabulary.sh` (✓ both guards clean), `check-references.sh` (all referenced sections resolve), `check-plan-provenance.sh` (3 files scanned, all provenance stated), `check-task-build-green.sh` (silent 0), `check-workspace-isolation.sh` (ISOLATION-OK), `check-contract-budget.sh` (BUDGET-OK: 24 contract files within budget — no row needed raising).

## Findings left open

None. All nine fixed.

## No commits

Nothing committed or pushed. All edits left uncommitted in the worktree, as instructed.
