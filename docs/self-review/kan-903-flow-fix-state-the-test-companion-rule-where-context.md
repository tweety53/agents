# Self-review context bundle for kan-903-flow-fix-state-the-test-companion-rule-where

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-903-flow-fix-state-the-test-companion-rule-where, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-903-flow-fix-state-the-test-companion-rule-where.md (absent)
skipped: .superpowers/sdd/reviews/kan-903-flow-fix-state-the-test-companion-rule-where-panel.md (absent)
skipped: spectre/changes/archive/kan-903-flow-fix-state-the-test-companion-rule-where/tasks.md (absent)
skipped: spectre/changes/archive/kan-903-flow-fix-state-the-test-companion-rule-where/design.md (absent)
skipped: spectre/changes/archive/kan-903-flow-fix-state-the-test-companion-rule-where/narrative.md (absent)

## git log --stat

commit 37658b3c1943df56e10493a058f2975a8902d075
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Oct 6 01:45:30 2026 +0300

    docs(flow): state the test-companion rule where shims are planned
    
    The companion requirement — every new check-*.sh lands with its
    test-check-*.sh companion in the same change — lived only inside
    run-guard-tests.sh's own header, so kan-838's plan author and review panel
    never saw it and the verify stage caught the missing harness as an
    unplanned commit. State the rule where shims are planned and reviewed: one
    sentence in the shim/port-pattern paragraph and one in the ## test
    section, each naming the runner's refusal and, in the first, its Go-test
    exception (kan-903).

 .flow/project.md | 7 +++++--
 1 file changed, 5 insertions(+), 2 deletions(-)

## Session narrative

KAN-903 arrived as the change's key in the `/flow-fast` command text, resolved against Jira
(To Do → In Progress), and named `kan-903-flow-fix-state-the-test-companion-rule-where` from the
summary's first 48 slug characters. The plan is one task in `.flow/project.md`, classed `micro`
(rolls compact 47 · experimental 84 · bundle 56 · effort 76), so the run went inline with the
default panel and visual verification skipped (config prose; nothing a user sees). The ticket's
"the shim template paragraph (the port pattern the skills state)" needed one judgment call: the
skills never restate the port pattern — `skills/flow/SKILL.md`'s only shim paragraph is the
sibling-dependency rule of the guard-presence check — so both sentences landed in
`.flow/project.md`, the only living prose stating the pattern, in the shim/port-pattern paragraph
and the `## test` section, one sentence each per the ticket's fix. A first idea of adding the
runner's Go-test exception to both sentences was cut back to the shim paragraph alone, keeping
each place at exactly one sentence and leaving `run-guard-tests.sh`'s header canonical for the
mechanics; the ## test sentence instead carries the ticket's purpose clause (plans declare the
companion task, panels look for it). Before editing, `check-verbatim-moves.sh`'s corpus was
checked and read as `skills/`, `commands-claude/` and `rules/` only, so `.flow/project.md` prose
needs no `verbatim-moves.txt` listing — confirmed by the guard's green run. Verification ran the
full `## lint` list in the worktree (all 27 entries exit 0; the SPA build was made first for
`go vet` and `tsc -b` in the fresh worktree), and the `## test` commands took an empty scope: the
diff names only `.flow/project.md`, which none of `run-guard-tests.sh`, the stats Go suite or the
SPA suite scopes to. The bundle's "RECORDS LOSS" note is expected, not a defect: the micro
decision's `flow.review-panel` stage is marked through with no panel and so writes no dispatch
rows.
