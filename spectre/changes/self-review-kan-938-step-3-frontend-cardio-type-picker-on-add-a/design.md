## Context

- Six non-`big` findings from the kan-938 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`), each carried by
  exactly one task. F4 and F5 (gymie dev tooling) are not in this change. Verified against the tree
  at `25c1f07c`:
  - `stats/internal/guard/closetask.go` builds `record dispatch end` without `-C`, while its
    `record decisions` and `tasks tick` calls pass `-C <canonical>`.
  - `stats/internal/guard/unfinishedwork.go` signal four counts a `reviewer` row only when its key
    starts `panel-`; its unreviewed-task message and `scripts/check-unfinished-work.sh`'s SIGNAL FOUR
    header say any reviewer row.
  - `skills/flow/visual-verify-verifier.md` step 7 and `skills/flow/verify-fix-loop.md` step 4 say
    "the round base" without saying which worktree's, where `specs` runs with `<frontend-root>`
    substituted.
  - `close-task` reads no dispatch rows at all.
  - `fallback.ProjectKey` (`stats/internal/fallback/statefile.go`) derives the key from `-C` (default
    cwd); no record write checks the change belongs to that repository.
  - `runRecordDispatchEnd` (`stats/cmd/flow/record.go`) passes `-outcome` through unchecked. The
    contracts close dispatch rows with `completed`, `clean`, `fix` (a gated reviewer,
    `skills/flow/implement.md`), `blocked`, `aborted`, `fallback` (`skills/flow/implement.md`) and
    `timed-out` (`skills/flow/review-panel.md`); `stopped`, `skipped`, `captured`, `abandoned` and
    `started` are stage-mark outcomes only.
- Every task's patch was spiked at plan time: applied in chain order from `25c1f07c`, the task's
  targeted tests green after each, each new test failing with its fix removed, `gofmt -l` and
  `go vet` clean. Task 4's two new sentences are acknowledged in this change's `verbatim-moves.txt`;
  `check-verbatim-moves.sh`, `check-references.sh`, `check-installed-citations.sh` and
  `check-markdown-integrity.py` exit 0 with task 4 applied.
- No finding was dropped.

## Decisions

### close-task ends the dispatch row under the canonical worktree

**ID:** close-task-end-under-canonical
**Status:** active
**Chosen:** `-C <canonical>` on close-task's `record dispatch end`, as its other two store calls
already pass.
**Considered:** documenting that close-task must run from the canonical worktree — a cwd convention
is exactly what failed.

### Signal four's message names the key it counts

**ID:** signal-four-names-panel-key
**Status:** active
**Chosen:** correct the breakdown message and the script header to name a panel reviewer row, key
`panel-<round>-<slot>`; the predicate is unchanged.
**Considered:** widening the predicate to any reviewer row — a non-panel reviewer (a pipeline-fix
review) does not review a fix run's appended task, and the existing test pins that case.

### specs takes the frontend worktree's own round base

**ID:** frontend-root-round-base
**Status:** active
**Chosen:** the verifier's step 7 and the fix loop's step 4 say the base substituted into `specs`'
`<merge-base>` is the `<frontend-root>` worktree's own round base, never another worktree's, and
that a `<frontend-root>` with none runs `specs` unscoped; `motions:` keeps the dispatched worktree's
own round base.
**Considered:** one round base per change — the worktrees' HEADs move independently, so a
backend sha means nothing to `git diff` in the frontend repository.

### close-task warns on an undispatched appended task

**ID:** close-task-warns-undispatched
**Status:** active
**Chosen:** after its dispatch end, close-task reads `flow record dispatches` only when a task it
closes is appended (the plan's `**Tasks appended:**` line) and prints one stderr `WARNING` per such
task no row names; its exit code and every later step are unchanged. The predicate is signal
four's, shared as `uwRowTasks` and `uwAppended` in `unfinishedwork.go`, so the two cannot drift.
**Considered:** refusing the close — close-task's own rule is that a record never blocks, and
integrate's gate already refuses; judging every task — plan-time tasks run grouped or inline and
most carry no row of their own, so every grouped run would warn.

### dispatch end takes a closed outcome set

**ID:** dispatch-outcome-closed-set
**Status:** active
**Chosen:** `recordOutcomes` = `completed`, `clean`, `fix`, `blocked`, `aborted`, `fallback`,
`timed-out` — every outcome a contract closes a dispatch row with; any other non-empty value exits 2
before the store is contacted, naming the set. An empty `-outcome` stays accepted.
**Considered:** mapping `done`/`approve` to `completed` — a silent rewrite hides which caller is
wrong; requiring `-outcome` — a separate change, and signal three already reports the open row.

### Run-record writes refuse a change no checkout carries

**ID:** change-checkout-check
**Status:** active
**Chosen:** `flow record dispatch begin`, `dispatch end`, `finding`, `status` and `pass` exit 2,
before the store is contacted, unless some worktree of the `-C` repository (`git worktree list`,
the main checkout included) holds `spectre/changes/<change>`, `spectre/changes/archive/<change>` or
`.superpowers/sdd/<change>` (a `/flow-fast` change root).
**Considered:** the `-C` checkout alone — a `/flow` parent runs most record calls from the main
checkout, where the change lives only in its worktree until it lands, so every run would refuse;
also checking `flow stage`, `flow tasks` and `flow self-review` — `flow stage begin` runs before
`spectre new` creates the change, `flow tasks` already reads the plan from `-C`, and a standalone
self-review of a `/flow-fast` change runs after its worktree is gone, when no root exists anywhere;
asking the store whether it knows the change — a round trip per write and no answer while the store
is down, when writes journal.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — tasks 5 and 6 add refusals to record writes every run makes, and a false
refusal there loses rows across every project.
**Considered:** `low` — under-reads a write-path refusal whose false positive is silent until a
ledger is read.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, tasks 1–6, one fixer in plan order — every task is a spiked patch applied
with `git apply`, and tasks 1/3 and 5/6 share files.
**Considered:** two groups (guard tasks 1–4, CLI tasks 5–6) — two dispatches for work one fixer
finishes inside its context.

## Open questions

## Live check

Build the CLI from the branch and run it against a dead store address with an isolated state root,
so nothing reaches the live `flowd` or the real journal:

```bash verified:spiked as go test cases at 25c1f07c; this exact invocation is the verify stage's own run
cd stats && go build -o /tmp/flow-sr ./cmd/flow
FLOW_STATE_DIR="$(mktemp -d)" /tmp/flow-sr record dispatch end -addr http://127.0.0.1:1 -timeout 300ms \
  -C /Users/tweety53/Projects/agents -change no-such-change-kan938 -key x -session-token mf-livecheck -outcome completed; echo "exit=$?"
FLOW_STATE_DIR="$(mktemp -d)" /tmp/flow-sr record dispatch end -addr http://127.0.0.1:1 -timeout 300ms \
  -C /Users/tweety53/Projects/agents -change self-review-noted -key x -session-token mf-livecheck -outcome done; echo "exit=$?"
```

Expected after: both print `exit=2` — the first naming `carries change "no-such-change-kan938"`,
the second naming `completed, clean, fix, blocked, aborted, fallback, timed-out`. Before — a build
of `25c1f07c`, run exactly so at plan time: both print `⚠ flow: store unreachable — wrote local
journal` and exit 0, each leaving journal files under its temporary state root. A `exit=0` after the change, or a
journal file appearing under the temporary state root, is the change not working.
