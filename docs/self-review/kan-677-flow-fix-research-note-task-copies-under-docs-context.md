# Self-review context bundle for kan-677-flow-fix-research-note-task-copies-under-docs

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-677-flow-fix-research-note-task-copies-under-docs, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-677-flow-fix-research-note-task-copies-under-docs.md (absent)
skipped: .superpowers/sdd/reviews/kan-677-flow-fix-research-note-task-copies-under-docs-panel.md (absent)
skipped: spectre/changes/archive/kan-677-flow-fix-research-note-task-copies-under-docs/tasks.md (absent)
skipped: spectre/changes/archive/kan-677-flow-fix-research-note-task-copies-under-docs/design.md (absent)
skipped: spectre/changes/archive/kan-677-flow-fix-research-note-task-copies-under-docs/narrative.md (absent)

## git log --stat

commit 9d6b017eea9ae66a5e97d3fabca989efe982f775
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 01:33:12 2026 +0300

    docs(flow): the research note is an immutable input, the plan canonical

 skills/flow/brainstorm-planner.md | 7 +++++++
 1 file changed, 7 insertions(+)

## Session narrative

KAN-677 carries kan-517's deferred panel finding F20: the research-note plan copy under
`docs/research/` goes stale when the plan is corrected mid-run. The reachability check settled the
scope first: the pipeline stopped producing such copies when flow-plan-inits-spectre retired the
writer (05154744) — capture commits only `<project>/spectre/changes/<name>/` — so the defect as a
pipeline behavior does not reproduce, and the one live stale copy is gymie's
`docs/research/kan-517/` (another repository, pre-retirement data, out of this run's reach and
deliberately not touched). The gap this tree still had is that the seeded-note rules in
`skills/flow/brainstorm-planner.md` never stated the note's status — neither canonical nor marked
stale, exactly the issue's words — so this change adds one paragraph there: the note is an
immutable input, the plan canonical, corrections land in `tasks.md` alone. Struggles: the first
draft failed `check-installed-citations.sh` because a backticked `docs/research/<stem>/` names no
root — fixed by prefixing `<project>/`; the fresh worktree needed `## worktree setup`'s
`make web-build` before `go vet` and `tsc` could run; and `check-worktree-location.sh` fails on a
pre-existing stray worktree — another session's scratchpad at
`/private/tmp/claude-501/.../scratchpad/ctl`, detached, modified today — which this run
deliberately left alone, making it the one lint command still failing (environmental, not the
diff). One caution for the deferred pass: the bundle's RECORDS LOSS note above is a flow-fast
false alarm — a `micro` decision runs no review panel, and `/flow-fast` marks `flow.review-panel`
as an empty pair so its stage set matches `/flow`'s, so no dispatch rows exist to find.
