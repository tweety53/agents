## Context

- The defect, as dispatched: the parent of kan-939 skipped steps of the pipeline-fix and
  self-review fix loops because both are prose only — fixers dispatched on `flow-high` with no
  planner, and three docs fixes (b8e0e063, bb273948, 3fd761da on agents `main`) landed with no
  review dispatch. The operator ordered skipping made impossible, mechanically: a dispatch-role
  hook, a landing guard and a loop script, each with a test that fails without it.
- The loops record nothing today: `flow record dispatches` returns `[]` for
  `pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-5` and for
  `self-review-kan-966-flow-mutation-proofs-bound-a-mutants-run-and`, and kan-939's own rows carry
  no `pipeline-fix-*` key. So the hook itself must record them.
- The running change of a loop may live in another project with no checkout left: kan-939 is
  `FINISHED` in gymie's project, its archive on `feature/cardio`, no gymie worktree carries it, so
  `flow record dispatch begin -C <gymie> -change kan-939-…` is refused by `requireChangeCheckout`
  (`stats/cmd/flow/record.go`). A dispatch row also needs a `changes` row in the store
  (`insertDispatch`, `stats/internal/store/records.go`, selects `FROM changes`).
- Install path: `setup.sh`'s `install_hooks` / `install_hooks_zcode` link every `hooks/*` file
  and print a registration snippet per hook through `warn_unregistered_hook`; they never edit
  `settings.json`. `enforce-agent-baseline.py` and `protect-main-checkout.py` are registered on
  this machine; a new hook is not until the operator adds its entry.
- The existing landing guard is `hooks/protect-main-checkout.py`'s `unarchived_landing`, a
  `PreToolUse` check on a typed `git push <remote> <src>:<default>`. A push made inside a script
  is invisible to it.
- `.flow/project.md` makes every new guard Go in `flow-guard` with a thin shim; scripts reached
  through a skill's `scripts/` symlink resolve the checkout with `flow_guard_root`, never
  `$SCRIPT_DIR/..` (`check-guard-symlinks` rule 4).
- Every task's diff below was applied in plan order to a scratch clone at `a1bca81b` and checked
  there: `go vet`, `gofmt -l`, `go test ./internal/guard/ ./internal/setuptest/ -race -count=1`
  and `scripts/run-guard-tests.sh` pass at the tip, each task's new tests fail with its production
  change absent, and every `## lint` line passes but two that read the machine rather than the
  tree (`check-installed-rules.sh` against `~/.claude`, `npx tsc -b` without `node_modules`).
- Blast radius: the files the three tasks touch plus the files citing the changed sections —
  well under the `big` threshold.

## Decisions

### A loop is its own change in the agents project

**ID:** loop-is-its-own-change
**Status:** active
**Chosen:** a loop's dispatch rows are filed under a change named after its branch —
`pipeline-fix-<name>-<k>` or `self-review-<name>` — in the agents repo's flowd project, the
session token being that branch name; `pipeline-fix.sh start` creates the change (`flow state set`
`STARTED`) and the `.superpowers/sdd/<branch>/` directory `requireChangeCheckout` accepts before
the planner commits, and `land`/`drop` close it `FINISHED`. A self-review on the change's own
branch has no branch of its own and files under `<name>`, the agents change it is part of.
**Considered:** filing under the running change — refused for a change with no checkout left
(kan-939) and needs a cross-project lookup; a local ledger in git config — the operator ordered
flowd.

### The pairs live in one table in the skill

**ID:** pairs-table-in-skill
**Status:** active
**Chosen:** the pairs paragraph of `skills/flow-self-review/SKILL.md` step 3 becomes a table, one
row per key suffix with its `subagent_type`s, model and escalation type; `check-loop-dispatch.sh`
parses that table at run time and a Go test pins it, so the pairs are stated once.
**Considered:** a data file the skill cites — a second file to keep in step for one rule;
parsing the old prose paragraph — breaks on any rewording.

### A re-plan takes its own key

**ID:** replan-key
**Status:** active
**Chosen:** a re-plan is keyed `<prefix>plan-<n>`, and a key that already has a row is refused,
so every admitted dispatch opens a fresh row — save a row closed `aborted`, which may be
dispatched again under its key: the checker records a dispatch before Claude Code runs it, so a
later denial (another hook, a permission prompt) leaves a row that never ran. The upsert returns
that row and the next `end` overwrites its outcome. A loop dispatch whose prompt lacks the
agent-baseline pointer is denied before it is recorded, since `enforce-agent-baseline.py` would
deny it after.
**Considered:** reusing `-plan` — the store's `(change, session token, key)` upsert would return
the earlier, closed row and the re-plan would go unrecorded.

### The escalation is a git-config marker

**ID:** escalation-marker
**Status:** active
**Chosen:** `pipeline-fix.sh escalate <key>` adds `<loop branch>:<key>` to
`git config flow.escalation` in the agents repo, refused unless the pairs row admits an
escalation and an earlier round of the same kind completed; `check-loop-dispatch.sh` admits the
row's escalation type only for a marker of this loop and only when the rows still show that
earlier completed round, so a hand-set or another loop's marker admits nothing; `land` and
`drop` clear the loop's markers.
**Considered:** a flowd decision row — one row per session token, and an own-branch loop's
decisions are its run's own Decide record; a dispatch row — an open row absorbs harvested tokens.

### The parent closes each dispatch with its outcome

**ID:** close-with-outcome
**Status:** active
**Chosen:** `pipeline-fix.sh end <key> <outcome>` — every dispatch but a review `completed` or
`aborted`; a review `end <key> <report-file>`, its outcome derived from the report's
`Critical: <n>` and `Important: <n>` lines (`clean` when both are 0, `fix` otherwise), never
the parent's word, or `aborted`; the next loop dispatch is refused while
any earlier one is open, so the close cannot be skipped.
**Considered:** a `PostToolUse` hook reading the report — a background Agent call's tool result
is its launch, not its report.

### Every new check is Go in flow-guard

**ID:** loop-guards-in-go
**Status:** active
**Chosen:** `check-loop-dispatch`, `check-loop-landing` and `pipeline-fix` are `flow-guard`
subcommands (`stats/internal/guard/loop.go`, `looplanding.go`, `pipelinefix.go`) behind thin
shims, per `.flow/project.md`; the hooks stay Python, like their siblings, and call the shims.
**Considered:** bash — refused by the project rule.

### A new hook file, registered by the operator

**ID:** separate-loop-hook
**Status:** active
**Chosen:** `hooks/enforce-loop-dispatch.py` beside `enforce-agent-baseline.py`, installed by the
same `install_hooks` loop, with `warn_unregistered_hook` snippets for Claude Code (`timeout` 300,
for a first `flow-guard` build) and ZCode. Until the operator registers it no loop dispatch is
recorded, so the landing guard refuses every loop landing: unregistered fails safe.
**Considered:** folding the check into the registered `enforce-agent-baseline.py` — immediate,
but the operator ordered a new hook beside it.

### Both loop checks fail closed

**ID:** loop-checks-fail-closed
**Status:** active
**Chosen:** a loop dispatch or loop landing whose checker cannot answer — flowd down, the table
unreadable, the shim missing — is denied; anything that is not a loop dispatch or landing keeps
the hooks' fail-open behaviour.
**Considered:** fail open everywhere, the hooks' convention — an unchecked loop step is the skip
this change exists to stop.

### The landing check rides the existing push guard

**ID:** landing-in-protect-main-checkout
**Status:** active
**Chosen:** `protect-main-checkout.py` resolves every push refspec landing on the default branch
— `<src>:<default>`, a colon-less `<default>`, no refspec — to a commit, and runs
`check-loop-landing.sh` for every `pipeline-fix-*`, `self-review-*` and `spectre/*` branch whose
tip that commit contains and `origin/<default>` does not, so `@:main`, `<sha>:main`, a temporary
branch and a loop pulled into the default branch are all checked; a `spectre/<name>` branch is a
loop only when self-review rows are recorded under `<name>`, and lands unchecked when flowd is
unreadable, so no ordinary landing depends on flowd. `git -c <k>=<v>` consumes its value before
the verb. `pipeline-fix.sh land` always verifies, refuses an unarchived
`spectre/changes/<branch>/`, then runs the same Go check in-process before its own push. Every
loop hook's subprocess wait (240s) stays under its registered timeout (300s), since a harness
that times a hook out lets the call through.
**Considered:** a git `pre-push` hook — not installed in this repository and per-clone.

### The loop guards stop a careless skip, not an adversarial parent

**ID:** loop-guard-ceiling
**Status:** active
**Chosen:** the guards catch a parent that skips a step by habit; a parent set on bypassing them
is out of scope. Known ceilings: a raw `flow record dispatch begin`/`end` falsifies the rows; a
push run inside `bash -c '…'` or another interpreter is not parsed; a command that itself creates
the landing ref (`git branch tmp && git push origin tmp:main`) is checked against HEAD, and one
that pulls a loop into the default branch and pushes it in the same command is not caught until
the push runs alone; a prompt that names a loop with no `Key:` line is denied by a name match,
not a parse.
**Considered:** sealing each hole — a signed record per dispatch, a server-side `pre-receive`
hook — far past the operator's goal of stopping the skips actually seen.

### k is unique across the repository

**ID:** repo-unique-k
**Status:** active
**Chosen:** `<k>` is one more than the highest `<n>` of any `pipeline-fix-*-<n>` branch (local or
remote) or change folder, live or archived, and `pipeline-fix.sh start` refuses any other — so
the `pipeline-fix-<k>-` key prefix names one loop, and two concurrent loops never collide.
**Considered:** unique per change name — two changes' loops with the same `<k>` share a prefix.

### start takes an optional k

**ID:** start-optional-k
**Status:** active
**Chosen:** `start <name> <k>` starts `pipeline-fix-<name>-<k>`; `start <name>` starts
`self-review-<name>`.
**Considered:** `start <name> <k>` alone — the self-review loop needs the same mechanics.

### verify picks lines by path

**ID:** verify-lines-by-path
**Status:** active
**Chosen:** every guard line of `## lint`, its `cd stats` lines only for touched Go, its
`cd stats/web` lines only for a touched SPA file; `go test` of each touched stats package, the
`scripts/test-<n>.sh` / `scripts/test-<n>-hook.sh` harness of a touched script or hook, the
`setuptest` package for `setup.sh`, vitest of a touched SPA test or the test beside a source.
**Considered:** the parent picking lines by hand — the step being replaced; the whole `## lint`
and `## test` lists — the skill forbids them for cost.

### A deferred loop is dropped by the script

**ID:** drop-subcommand
**Status:** active
**Chosen:** `pipeline-fix.sh drop <worktree>` writes the loop's `FINISHED` record with
`withdrawn: true` and removes the worktree and branch.
**Considered:** the parent removing them by hand — leaves the loop's change `STARTED` forever.

### Forward-only

**ID:** forward-only
**Status:** active
**Chosen:** b8e0e063, bb273948 and 3fd761da stay as landed; nothing back-fills rows for them.
**Considered:** back-filling — the dispatches never happened, so there is nothing true to record.

### First review on medium

**ID:** first-review-effort
**Status:** active
**Chosen:** the first review, `pipeline-fix-6-review-1`, runs on `flow-medium` (effort `medium`)
on `opus`: the change gates every loop dispatch and every loop landing.
**Considered:** `low` — too little for two hooks on the dispatch and push paths.

### One fixer group

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Tasks 1, 2 and 3 in series in one `pipeline-fix-6-fix` dispatch: Task 2's
tests reuse Task 1's fixture and Task 3's code calls both guards' helpers.
**Considered:** a fixer per task (`-fix-g1`…`-fix-g3`) — three dispatches re-reading the same
helpers for no parallelism, the tasks being serial.

## Open questions

None.

## Live check

After this lands and the operator registers `enforce-loop-dispatch.py`, the next loop is the
check, against the real flowd: `scripts/pipeline-fix.sh start <name> <k>` prints its worktree and
`flow state get -C <agents repo> pipeline-fix-<name>-<k>` reads `STARTED`; the plan dispatch is
admitted and `flow record dispatches -C <agents repo> -change pipeline-fix-<name>-<k>` lists its
row with the loop's branch as `sessionToken`; a fix dispatched before `pipeline-fix.sh end` closes
the plan is denied `OPEN`; a typed `git push origin pipeline-fix-<name>-<k>:main` before a clean
review is denied. Not working looks like an empty `dispatches` list after an admitted dispatch,
or a fix admitted on `flow-high` with no `escalate`.
