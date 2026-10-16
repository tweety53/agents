## Context

- Eight non-`big` findings from the kan-939 step-4 self-review pass, planned as one change per item 1
  of **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`), plus one
  pre-existing defect planning found. Each finding is carried by exactly one task; none is dropped.
  All were checked against the tree at `a7196b56` and, for F-7, against the gymie-frontend worktree
  `kan-939-step-4-frontend-cardio-plan-view-start-log` at `7eb56b16`:
  - F-1 (Task 1): `scripts/integrate-change.sh` sets `PLAN_MSG="chore(spectre): plan $NAME"` and
    `--plan-message` overwrites it whole; `commit-split` commits it as `-m "$PLAN_MSG"`, and gymie's
    and gymie-frontend's `.githooks/commit-msg` run `validate-commit-msg.sh`, which refuses a subject
    outside Conventional Commits. `skills/flow/integrate.md` tells the parent to pass the planning
    message "whole as `--plan-message`".
  - F-2 (Task 2): the push-time `check-untasked-commits.sh` sentence sits only in the full-suite fix
    paragraph of `skills/flow/implement.md`; **A branch the remote already holds takes the fix as one
    new commit on top** (`skills/flow/review-panel-fix-round.md`) pushes a panel fix with no check.
  - F-3 (Task 3): signal four (`stats/internal/guard/unfinishedwork.go`) judges a ticked appended
    task by dispatch rows alone; a row back-filled after the fact satisfies it with no commit.
  - F-4 (Task 4): `vvdCompletedVerifier` (`stats/internal/guard/visualverifydispatched.go`) answers
    found for any completed or ended verifier row, whatever landed after it.
  - F-5 (Task 5): `dispatch begin`'s `-started-at` and `dispatch end`'s `-ended-at` are optional, and
    only close-task's RECORD message and its `implement.md` paragraph ask for them. The harvester's
    `maxStampTimeTolerance` comment (`stats/internal/harvest/watcher.go`) shows live panel rounds also
    pass `-started-at` on `begin` and end normally, so `-started-at` alone does not mark a back-fill.
  - F-6 (Task 6): signal four's comment states "pass rows have no read verb"; `flow record` has
    `dispatches` and `findings` read verbs but none for passes. `review-panel.md` and
    `review-panel-fix-round.md` require a pass row on every panel run and fix-round pass.
  - F-7 (Task 7): in the gymie-frontend worktree `./gradlew test --dry-run` lists no `desktopTest`
    task, and `./gradlew test :shared:desktopTest --dry-run` lists `:shared:desktopTest`.
  - F-8 (Task 8): `runRecordDispatchBegin` (`stats/cmd/flow/record.go`) sends `-model` as given.
  - Task 9: `scripts/check-installed-citations.sh` exits 1 at `a7196b56` on
    `skills/flow/verify-and-handoff.md:207`, the citation `scripts/check-visual-verify-dispatched.sh`
    naming no root.
- Every agents-repo task's diff was applied in plan order to the worktree at `a7196b56`, checked, and
  reverted: `go vet` and `gofmt -l` are clean on `./cmd/flow/ ./internal/guard/ ./internal/records/`;
  `go test` of those three packages passes; `scripts/test-integrate-change.sh` passes 101 of 101; each
  task's new tests fail with that task's production files reverted; `check-verbatim-moves.sh` (with
  this change's `verbatim-moves.txt`), `check-references.sh`, `check-markdown-integrity.py`,
  `check-installed-citations.sh`, `check-normative-inventory.sh`, `check-dispatch-paragraphs.sh` and
  `check-guard-symlinks.sh` exit 0. Task 7's diff passes `git apply --check` in the gymie-frontend
  worktree, and `project-get.sh` reads its `## test` back.

## Decisions

### The planning commit's outstanding work rides in its body

**ID:** outstanding-in-body
**Status:** active
**Chosen:** replace `--plan-message <msg>` with `--outstanding <text>`; `commit` writes
`chore(spectre): plan <name>`, a blank line, then `Outstanding: <text>`, so the subject is always the
script's own. An empty `--outstanding` is a usage refusal. `--plan-message` is removed, so a caller
still passing it stops at argument parsing, before the reshape.
**Considered:** keeping `--plan-message` and refusing a value without a Conventional Commits subject
at parse time — keeps a flag whose only use is the outstanding list, and makes the parent compose a
subject the script already knows; editing `spectre/changes/faster-integrate/verbatim-moves.txt` —
that in-flight change's acknowledgement of the old sentence is harmless once the sentence is gone, so
another change's directory is left untouched.

### The untasked-commit check before a push lives in Commits outside every task

**ID:** untasked-check-in-git-boundaries
**Status:** active
**Chosen:** move the full-suite fix's sentence — run `check-untasked-commits.sh <worktree>
<merge-base>` before the push, amend the tip when it is the only commit listed — into **Commits
outside every task** (`skills/flow-contracts/git-boundaries.md`), stated for every commit outside
every task, and cite it from the full-suite fix (`skills/flow/implement.md`) and the panel fix
(`skills/flow/review-panel-fix-round.md`).
**Considered:** copying the sentence into `review-panel-fix-round.md` — two copies of one rule drift;
a check at `write-panel-diff.sh` — by then the commit is pushed, and amending it is a rewrite **Panel
re-runs** forbids on pushed history.

### Signal four also needs each ticked appended task's commit, searched across the change's ranges

**ID:** task-id-commit-across-ranges
**Status:** active
**Chosen:** `check-unfinished-work.sh` takes trailing `<worktree>=<merge-base>` arguments, one per
worktree of the resolved set, after `[canonical-worktree]`. Signal four searches every range's
`--no-merges` commits for a `Task-Id:` trailer naming each ticked appended task, and reports the ones
none carries. A plan with a ticked appended task and no range passed is exit 2, as is a range git
cannot read. `integrate-change.sh` passes every worktree's `merge-base HEAD origin/<base>`; the fix
run's call in `skills/flow/verify-and-handoff.md` and the call shape in
`skills/flow-contracts/finish-contract-run1.md` carry the ranges.
**Considered:** searching the judged worktree alone — a cross-repo change's fix task often commits in
the peer repository (gymie's frontend), so the gate would report every such task; reading the
implementer row's `commitSha` — a back-filled row can carry any sha, and a rebase rewrites the
recorded one; reading the worktree set from the state record inside the guard —
`finish-contract-run1.md` forbids a raw read of the `worktrees` map for set resolution, and the
callers already hold the resolved set; searching all of `HEAD`'s history — other changes'
`Task-Id: <n>` commits on the base would satisfy it.

### A verifier row counts only for the HEAD it verified

**ID:** verifier-counts-for-its-head
**Status:** active
**Chosen:** the verifier's `dispatch begin` carries `-diff-base` with the sha `git rev-parse HEAD`
prints before the dispatch (`skills/flow/visual-verify.md`, **Recording**). A completed or ended
verifier row counts only when that `diffBase` is a commit sha (7–40 hex) this worktree resolves and
the trigger — the canonical worktree's `ui paths` included — answers "no UI path" for the paths of
this change's own `<merge-base>..HEAD` diff that changed between that sha and HEAD. With qualifying
rows and none fresh, the verdict is a distinct `VISUAL-VERIFY-MISSING` line naming the stale HEAD.
A row with no `diffBase` counts for nothing.
**Considered:** comparing the row's `endedAt` with the last UI-path commit's date — the finding notes
back-filled timestamps are wrong (F-5), and a commit's date survives a rebase that moves it; `-commit`
on `dispatch end` — KAN-809's ended-but-unclosed row lost its end call, so the HEAD would be lost with
it, and `-commit` means the commit a dispatch produced; a new `-verified-head` column — a store
migration for what `diffBase` already holds: the sha the dispatch's view was computed from; counting
a row with no `diffBase` as before — the hole stays open for any parent that omits the flag;
diffing `<sha>..HEAD` without the intersection — a base the branch was rebased onto would read as
this change's UI work.

### A back-fill is one call with its real instants required

**ID:** backfill-is-one-call
**Status:** active
**Chosen:** `flow record dispatch backfill` takes `begin`'s flags and `end`'s together, requires
`-started-at`, `-ended-at` and `-outcome`, validates both instants before writing, then runs `begin`
and `end`. `skills/flow/implement.md` states that every back-fill is that one call, and close-task's
RECORD message names it.
**Considered:** making `begin -started-at` require an `-ended-at` — live panel rounds pass
`-started-at` on `begin` and end later, so that refuses the ordinary path; a store flag marking a
client-supplied `startedAt` and an `end` that refuses a stamped `endedAt` on such a row — a migration
and a daemon rule for what one CLI verb states; prose alone — a sentence nothing enforces is the
state the finding reported.

### Pass rows get a read verb, and the gate reports a dispatched round with none

**ID:** passes-read-verb
**Status:** active
**Chosen:** `flow record passes -change <name>` prints the change's pass rows as one JSON array,
`dispatches`' twin through a shared run-record read. Signal five of `check-unfinished-work` collects
every round a reviewer row keyed `panel-<round>-…` ran in — `records.PanelDispatchRound`, exported
from the renderer that already reads rounds that way — and reports each with no pass row of that
round. The pass rows are read only when such a round exists; a failed or unreadable read is exit 2.
**Considered:** the gate reading the rendered panel record — signal two's comment records why the
gate reads store rows, never rendered files; a pass-row read inside `dispatches`' output — changes a
verb other guards parse.

### gymie-frontend's full suite names the desktop UI tests

**ID:** frontend-test-names-desktop
**Status:** active
**Chosen:** a `## test` section in `/Users/tweety53/Projects/gymie-frontend/.flow/project.md`, landed
on the change branch `spectre/kan-939-step-4-frontend-cardio-plan-view-start-log` in its worktree,
running `./gradlew test :shared:desktopTest`, with a note on why the task is named.
**Considered:** making `test` depend on `desktopTest` in the Gradle build — a build-logic change in a
second repository that also slows every developer's `./gradlew test`, where the suite the pipeline
runs is what the finding names.

### A full Claude model id is recorded under its alias

**ID:** model-alias-on-record
**Status:** active
**Chosen:** `dispatch begin` (and so `backfill`) maps a `-model` starting `claude-opus-` or
`claude-sonnet-` to `opus` or `sonnet`; every other value is kept verbatim.
**Considered:** normalising when rendering or querying — every reader would need the rule; telling the
parent to pass the alias — the parent's own model id is what its session reports.

### The root-less citation is fixed in this change

**ID:** citation-root-fix
**Status:** active
**Chosen:** prefix the citation in `skills/flow/verify-and-handoff.md` with `<agents repo>/`, the
form `check-installed-citations.sh` names, as its own task.
**Considered:** leaving it for another change — `origin/main` fails the guard today and the fix is one
token.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — Tasks 3, 4 and 6 change what two integrate gates accept, including cross-repo
and rebased-branch behaviour, and Task 5 adds a CLI verb composing two writes; a low-effort read could
pass over a gate that refuses legitimate history.
**Considered:** `low` — four of the nine tasks are a config row or a few prose lines, but the guard
tasks carry the review risk.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, g1: Tasks 1–9 in plan order, one fixer — every diff is spiked and applies with
`git apply`; Tasks 3 and 6 share `unfinishedwork.go`, Tasks 5, 6 and 8 share `record.go`, and Tasks
2, 3, 5 and 9 share skill files, so the agents-repo tasks must run in series anyway; Task 7 is a
one-section config edit in the gymie-frontend worktree.
**Considered:** a second group for Task 7 alone — one more dispatch for a `git apply` and a commit;
splitting the Go tasks from the prose tasks — the shared files make them serial, so a split adds a
dispatch and saves nothing.

## Open questions

## Live check

After Task 6, from this worktree, `cd stats && go run ./cmd/flow record passes -change
kan-939-step-4-frontend-cardio-plan-view-start-log -C /Users/tweety53/Projects/gymie` against the
running flowd prints one JSON array — the change's pass rows, `[]` if it has none — at exit 0; an
`unknown record command` refusal, `null`, or a non-zero exit with flowd up is the change not working.
Before the change the same command is refused as an unknown subcommand. The other tasks change guards
whose tests drive them against git fixtures and HTTP doubles, a config row run only by a pipeline's
full suite, and prose.
