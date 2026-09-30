# kan-860-mechanics-into-scripts — session narrative

## 2026-09-30 — creating run

Resumed at `STARTED` with decision #101 (class big, sdd, 12 groups on opus). The operator raised
the wave cap to 5 parallel implementers for this run, and mid-run had a subagent remove `xhigh`
from the pickable efforts (merged to main as 43125acf); group 5, decided at `xhigh`, ran at `high`
instead. The operator also said to take the recommended option on every question and continue,
which is how every implementer question and gated-review decision below was settled.

The gated per-task reviewers earned their keep on the destructive seams: fold-fixup and
sync-onto-base popped whichever marker stash sat on top of the shared stash list (a Critical on
task 8 — another worktree's planning paths could land in this one), throwaway-worktree left its
copy behind when the fold-back failed, and the late-fix trigger and refresh-plan-base inherited
the machine's rename detection. Each was fixed inline by the parent with a regression test proven
to fail on the old code. Task 12's citation counts dropped to zero once the reviewer templates
lost their Placeholders lists; declaring them expected-zero was the recommended route.

The full `## test` list surfaced three things no per-task review saw: `check-task-commit-fields`
read `**Files:** none` as a path (task 26's empty live-verification commit was the first to
declare a commit alongside it), six new git reads lacked the `check-git-config-pins` pins, and the
setup fixtures no longer carried what `project-get.sh` needs once it became a flow-guard shim
(task 16). Running the three suites concurrently produced contention failures in `setuptest` that
vanished run alone; `reconcile`'s append-vs-retire race flaked once and is already in
KNOWN-BUGS.md.

Main moved six commits during the run — the xhigh removal, the test-sanity experimental reviewer
and the medium-effort roll, the last two dispatched and merged at the operator's request from this
same session. The medium-effort roll conflicts with task 1's plan-class work in `plan-class.sh`,
`planclass.go`, `plan_class_test.go` and `brainstorm-planner.md`, so the panel stopped once at the
base check (recommended **Stop**), then, on the operator's "continue", reviewed the branch as is;
that conflict is left for integrate's sync onto main.

The panel (primary+principles, opus/high, compact) raised six Important and four Minor findings;
one fix round on opus/high closed nine and withdrew F5 (the relocation comparison cannot be
generated for symlinked Files under an existing rule — the generator now fails loudly instead of
silently writing nothing). The fixes were plain commits on the pushed branch, so their plan-field
changes are Correction paragraphs rather than field edits, which would have failed
`check-task-records` and `check-plan-shape`. Both sonnet/low re-runs came back clean.

## 2026-09-30 — integrate run

Preflight, foreign-staged and drift checks were clean; the unfinished-work gate returned `CLEAR`
and no UI path was touched. `check-base-moved.sh` reported `MOVED` — six commits on `main`, nine
overlapping paths. The sync used the installed (main) `sync-onto-base.md`, since this change's own
`sync-onto-base.sh` is not installed yet; reading the worktree's copy first cost one failed call.

The rebase of 62 commits onto `5d022b50` stopped twice. Task 1's plan-class commit conflicted in
`plan-class.sh`, `planclass.go` and `plan_class_test.go`: resolved keeping both the effort roll
and the four tree lines, with the pins' prefix match kept. The decision-render commit conflicted in
`brainstorm-planner.md`, where main had added a `roll: effort` row to the inline table spec this
change moves into `flow decision render`: resolved to the render command, and the effort row
carried into the renderer (`decision.go`, its tests) in a new fix commit, with the two
`TestPlanClassTree` pins and the `brainstorm-planner.md` "three rolls" wording brought to four.
`check-verbatim-moves.sh` then flagged main's three reworded sentences; `verbatim-moves.txt` was
updated to list them. The full `## lint` and `## test` lists passed afterwards.
