# Self-review context bundle for kan-775-flow-require-a-reverted-prototype-s-rejection

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-775-flow-require-a-reverted-prototype-s-rejection, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-775-flow-require-a-reverted-prototype-s-rejection.md (absent)
skipped: .superpowers/sdd/reviews/kan-775-flow-require-a-reverted-prototype-s-rejection-panel.md (absent)
skipped: spectre/changes/archive/kan-775-flow-require-a-reverted-prototype-s-rejection/tasks.md (absent)
skipped: spectre/changes/archive/kan-775-flow-require-a-reverted-prototype-s-rejection/design.md (absent)
skipped: spectre/changes/archive/kan-775-flow-require-a-reverted-prototype-s-rejection/narrative.md (absent)

## git log --stat

commit 4b466b2f6d1cfac3927486939c4f3dbc11cb9445
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Oct 5 19:28:27 2026 +0300

    docs(flow): require a back-out commit body to carry the rejection reason
    
    KAN-703's run re-derived a reverted prototype's intended polarity from
    contract wording because the revert's body recorded only 'This reverts
    commit <sha>'. State the convention once in the git-boundaries contract —
    a landed-then-reverted prototype or a withdrawn task is backed out by a
    commit whose body says what was wrong and what would have to change for
    the design to come back — and cite it at the two commit-making sites
    (/flow's COMMIT-PER-TASK paragraph, /flow-fast's section 4), so a later
    run inherits the verdict with the artifact.

 skills/flow-contracts/git-boundaries.md | 10 ++++++++++
 skills/flow-fast/SKILL.md               |  4 +++-
 skills/flow/implement.md                |  6 +++++-
 3 files changed, 18 insertions(+), 2 deletions(-)

## Session narrative

The run implemented KAN-775 inline on a `micro` decision: one task, one commit, no panel
(default), no dispatches — the bundle's RECORDS LOSS note is therefore a designed absence, not a
records loss: the empty `flow.review-panel` pair marks a panel the class never runs, and the
ledger and panel sources are absent because none exist. The change states the back-out convention
once in `skills/flow-contracts/git-boundaries.md` (new closing section **Revert and back-out
commits**) and cites it at the two commit-making sites — the FLOW — COMMIT-PER-TASK paragraph in
`skills/flow/implement.md` and the commit rule in `skills/flow-fast/SKILL.md` section 4.

Approaches tried and abandoned: a guard that scans back-out commit bodies for a rejection reason
was considered and deliberately not built — the ticket's fix sentence asks for a stated convention
plus citations, and an executable judge of what counts as a reason would heuristically rule on
prose; the plan's `**Regression:**` field records the consequence (the guards stay green even with
the convention reverted, so the corpus guards alone cannot prove this change's presence — only its
well-formedness). The withdrawal route (`skills/flow/withdrawal.md`) was considered as a third
citation site and skipped: it makes no commits, so a commit-body convention binds nothing there.

Where it struggled: `check-verbatim-moves.sh` first came back 1-of-6 acknowledged — the heading
line `## Revert and back-out commits` was captured verbatim from the guard's own FAIL output, yet
still failed, because a `#`-leading line in `verbatim-moves.txt` is a comment; the guard's header
states the escape (`\## …`), found only after reading the script. Also, the opening
`flow stage mark` for `flow.create-artifacts` recorded a completed pair where the skill's section
3 spells `begin`+`end` — the stage set matches, the duration is absent; later stages used
begin/end. Nothing was left unfinished; every lint command in the project's `## lint` list ran
green in the worktree, and `## test` ran nothing because the diff names only skills Markdown,
which no declared test command scopes to.
