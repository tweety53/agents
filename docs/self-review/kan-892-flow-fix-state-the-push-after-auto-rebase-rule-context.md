# Self-review context bundle for kan-892-flow-fix-state-the-push-after-auto-rebase-rule

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-892-flow-fix-state-the-push-after-auto-rebase-rule, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-892-flow-fix-state-the-push-after-auto-rebase-rule.md (absent)
skipped: .superpowers/sdd/reviews/kan-892-flow-fix-state-the-push-after-auto-rebase-rule-panel.md (absent)
skipped: spectre/changes/archive/kan-892-flow-fix-state-the-push-after-auto-rebase-rule/tasks.md (absent)
skipped: spectre/changes/archive/kan-892-flow-fix-state-the-push-after-auto-rebase-rule/design.md (absent)
skipped: spectre/changes/archive/kan-892-flow-fix-state-the-push-after-auto-rebase-rule/narrative.md (absent)

## git log --stat

commit cc473119b8a0dca7ecf72a8a7c22d5b0353aab45
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 01:45:07 2026 +0300

    docs(flow): state the push-after-auto-rebase force-with-lease rule at the push sites

 skills/flow-fast/SKILL.md             | 3 ++-
 skills/flow/review-panel-fix-round.md | 5 ++++-
 2 files changed, 6 insertions(+), 2 deletions(-)

## Session narrative

A `/flow-fast` creating run for KAN-892: state the push-after-auto-rebase rule (`--force-with-lease`,
never a bare `git push`, never `--force`) at the two push sites the ticket names. The roll came out
`micro`, so the decision collapsed to inline execution with the `default` panel and no dispatches;
one task, one commit touching `skills/flow/review-panel-fix-round.md` (the fix-round contract's
fix-commit push paragraph) and `skills/flow-fast/SKILL.md` (the per-commit Branch-backup push step).
Both new sentences are run-loaded prose, so each was listed verbatim in the change's
`verbatim-moves.txt` exactly as `check-verbatim-moves.sh` printed it after `::`; the guard, the
reference guard and markdown integrity then all passed, and the normative inventory was proven
unchanged by grep (no added or removed line carries MUST/SHALL). The full `## lint` list ran green in
the worktree after a first `make web-build`; the `## test` list took an empty scope, the diff naming
two prose files no test command reads. Two approaches were tried and abandoned: editing
`skills/flow-fast/SKILL.md` from this session's own loaded copy of the skill failed twice (the file
was not read from disk, and the loaded text's line wrapping differs from the worktree file's, so the
exact-match edit missed) — re-reading the file and matching its actual wrapping worked, and is the
lesson. Editing `skills/flow-contracts/git-boundaries.md`'s "every other push is plain" clause was
considered and rejected: it is outside the two sites the ticket names, and rewording a normative
sentence the ticket did not ask to change buys drift risk, not the fix. The bundle's RECORDS LOSS
note above is a heuristic misfire, not a loss: `/flow-fast` marks `flow.review-panel` as an empty
pair on every run, and this micro run dispatched no panel, so the store correctly holds no dispatch
rows.
