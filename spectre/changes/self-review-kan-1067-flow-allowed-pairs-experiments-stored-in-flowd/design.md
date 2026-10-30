## Context

Self-review fix change for `kan-1067-flow-allowed-pairs-experiments-stored-in-flowd`, planned on
that change's own branch `spectre/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd` at
`ca792d7b`, in `/Users/tweety53/Projects/agents-worktrees/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd`.
Six non-`big` findings, as `proposal.md` names them, each carried by one task in `tasks.md`: 1-A
(Task 1), 1-B (Task 2), 1-C (Task 3), 1-D (Task 4), 1-E (Task 5), 1-F (Task 6). Each was traced on
that HEAD before its fix was designed; every question was auto-resolved to its recommended option
(**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). The evidence is
the pass's `.superpowers/sdd/self-review-pass.md` in this worktree; the traces are in
`.superpowers/sdd/2026-10-30-self-review-kan-1067-flow-allowed-pairs-experiments-stored-in-flowd-design.md`.

`flowd` on 4173, the `flow-postgres` container and the `flow` database are never stopped, restarted
or dropped; the live check below runs on a scratch copy.

## Decisions

### check-task-records joins a task's Task-Id commits

**ID:** records-join-task-id-commits
**Status:** active
**Chosen:** check 2 compares the declared `**Files:**`/`**Allowed-collateral:**` union, in both
directions, against the union of the subject commit's files and the files of every non-merge
`<base>..HEAD` commit in the group's repository whose `Task-Id:` trailer names one of the group's
task ids — the trailer read the way `tcfTaskIDCommits` reads it (`ctTaskNumberRE`, so `Task-Id: 2+3`
counts for both). A `git log` that fails in that repository adds nothing. The untouched-file line
keeps naming the subject commit.
**Considered:** dropping the untouched direction for tasks with a fix commit — loses the half of the
check `check-task-commit-fields` cannot make; narrowing `check-task-commit-fields` instead — KAN-976
made it join fix commits so a gated fix's paths are declared, and that stays the rule.

### close-task runs check-plan-shape on every close

**ID:** plan-shape-on-every-close
**Status:** active
**Chosen:** the `uwAppended` condition goes; every close runs `check-plan-shape` on the canonical
`tasks.md` before any commit guard, as it already did for an appended task. A plan that passed at
plan time costs one parse. `TestCloseTaskRunsPlanShapeOnInRunFix`'s "a planned task never runs the
guard" case pins the old behaviour and flips to refused.
**Considered:** running it only when `tasks.md` changed since the plan commit — a second mechanism
for the same one-parse cost; running it inside `flow tasks files add` — the widening is not the only
hand edit of `tasks.md`.

### run-verify-list refuses an off-contract key

**ID:** verify-key-refused-at-run
**Status:** active
**Chosen:** with the recording arguments given, `run-verify-list` folds `<key>` with
`check-verify-green`'s own `vgFixSuffix` and refuses, exit 2 before any command and before the
`begin`, a result that is neither `verify` nor `verify-<s>` where `<s>` is the list worktree's
basename or its repository's name (the basename of the directory holding
`git rev-parse --path-format=absolute --git-common-dir`; a git failure admits only the basename).
Both are admitted because `skills/flow/verify-and-handoff.md` **Recording.** names the suffix
`-<repository name>` in prose and `<worktree basename>` in its command block.
**Considered:** loosening `vgFixSuffix` — reads `verify-3d-viewer` as a re-run (kan-1048
self-review, 1-A); refusing only `verify-fix-*` keys that do not fold — misses `verify-2b`.

### The experiments tests read the policy table from its source

**ID:** api-test-reads-policy-source
**Status:** active
**Chosen:** `newExperimentsTestServer` reads `../../../skills/flow-contracts/model-policy.md` with
`os.ReadFile` (fatal on error) in place of `embedded.Table`; the `modelpolicy/embedded` import
leaves `internal/api` entirely. Same bytes; no make dependency.
**Considered:** copying ignored files into the throwaway copy — `throwaway-worktree.sh` would then
copy every build artefact; running `make policy-copy` in the mutation slot — one more step every
slot must remember.

### live-store-copy admits a hyphen

**ID:** live-copy-hyphen-maps-to-underscore
**Status:** active
**Chosen:** `<name>` matches `^[a-z0-9][a-z0-9_-]*$`; the database is `flow_${name//-/_}_live`,
while `/tmp/flow-live-<name>/` keeps the name as given. Two names differing only in `-`/`_` share a
database, and the existing "database exists" refusal stops the second `up`.
**Considered:** quoting the identifier — every `createdb`/`dropdb`/`psql` call would need it.

### Markdown-editing tasks verify with the whole lint list

**ID:** markdown-tasks-run-whole-lint
**Status:** active
**Chosen:** one sentence after the verify-step rule in section D of `skills/flow/brainstorm-planner.md`,
listed in this change's `verbatim-moves.txt`. `skills/spectre-planning/SKILL.md` carries no
verify-step sentence, so nothing is mirrored into spectre's vendored copy.
**Considered:** naming `check-ask-silence.sh` in the hand-picked list — the next unlisted guard fails
the same way; whole lint for every task — the rule exists because Go-only tasks need no prose guards.

### What the defects already wrote is left as written

**ID:** no-record-repair
**Status:** active
**Chosen:** no repair. 1-A narrowed tasks 4 and 13 of the archived kan-1067 plan, each with a
`Correction` paragraph that states the history truthfully; the task commits are collapsed into one
implementation commit, so no guard re-reads them, and a self-review fixer never writes
`spectre/changes/`. 1-C's `verify-fix-1b`/`verify-fix-2b` rows each carry their real outcome and
were superseded by `verify-fix-2`; nothing in the store is wrong.
**Considered:** re-widening the archived records — would contradict their own Correction paragraphs.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 2 puts `check-plan-shape` on every close of every run and Task 3 a
refusal on every recorded verify; a wrong predicate in either stops runs.
**Considered:** `flow-low` — the diffs are small, but both sit on every run's path.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, `-fix`, Tasks 1–6 in series in this worktree.
**Considered:** one fixer per task — six small, file-disjoint diffs; a second context costs more than
it saves. Two groups (Go vs scripts/prose) — Tasks 5 and 6 are a few lines each.

## Open questions

## Live check

Each step runs before landing, with every task committed, from this worktree; it never touches
4173 or the `flow` database beyond `live-store-copy.sh`'s read-only dump.

1. **Hyphenated copy.** `scripts/live-store-copy.sh up sr-1067 4176`. Before: exit 2, "name
   'sr-1067' is not [a-z0-9][a-z0-9_]*". After: prints `http://127.0.0.1:4176` and the exports, and
   `docker exec flow-postgres psql -U flow -d flow -tAc "select 1 from pg_database where datname='flow_sr_1067_live'"`
   prints `1`. Not working looks like the exit 2, or no `flow_sr_1067_live` database.
2. **Off-contract key on real records.** With the printed exports set,
   `scripts/run-verify-list.sh . lint kan-1067-flow-allowed-pairs-experiments-stored-in-flowd verify-fix-1b mf-live-1067 opus low`
   exits 2 with `run-verify-list`'s own key line, before any command, and
   `flow record dispatches -C . -change kan-1067-flow-allowed-pairs-experiments-stored-in-flowd`
   lists no row keyed `verify-fix-1b` with session token `mf-live-1067`. Before the fix the same
   call hands the key to `flow record dispatch begin` in the copy. Not working looks like any list
   command's output, a `flow record dispatch` message, or that row.
3. `scripts/live-store-copy.sh down sr-1067`; the `pg_database` query of step 1 prints nothing.

## Recommended approach

Fix each defect where it starts: the records guard's commit set (1-A,
`records-join-task-id-commits`); the close-time guard call (1-B, `plan-shape-on-every-close`); the
one tool that records a verify key (1-C, `verify-key-refused-at-run`); the test's table source (1-D,
`api-test-reads-policy-source`); the script's name check (1-E,
`live-copy-hyphen-maps-to-underscore`); the writing-plans rule (1-F,
`markdown-tasks-run-whole-lint`). Each finding has one obvious fix, so no Explore or Plan agents
were dispatched.

## Critical files

- `stats/internal/guard/taskrecords.go`, `stats/internal/guard/check_task_records_test.go` — 1-A.
- `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go` — 1-B.
- `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go` — 1-C;
  `stats/internal/guard/verifygreen.go` — `vgFixSuffix`, read only.
- `stats/internal/api/experiments_test.go` — 1-D.
- `scripts/live-store-copy.sh`, `scripts/test-live-store-copy.sh` — 1-E.
- `skills/flow/brainstorm-planner.md` — 1-F.

## Existing code to reuse

- `tcfTaskIDCommits` and `ctTaskNumberRE` (trailer reading), `trCommitFiles`, `tcfCheckFiles`,
  `trRepo`/`trFx` fixtures in `stats/internal/guard`.
- `Registry["check-plan-shape"]`, `ctFixture`, `ctPlan`, `bt` in `stats/internal/guard/close_task_test.go`.
- `vgFixSuffix` in `stats/internal/guard/verifygreen.go`; the recording fixture of
  `run_verify_list_test.go` (fake `flow` logging its calls).
- `scripts/live-store-copy.sh`, `scripts/break-and-prove.sh`, `scripts/run-verify-list.sh`.

## Verification

Beyond each task's verify step: `scripts/run-guard-tests.sh` and `cd stats && make -s web-dist &&
go test -race -count=1 ./...` pass; `scripts/run-verify-list.sh <wt> lint` passes; the **Live
check** above.
