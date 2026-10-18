## Context

Found by the self-review pass of `/flow` change `self-review-backlog-flow-6`; the fixes ride that
change's own branch `spectre/self-review-backlog-flow-6`, based at `95deb422`.

**Finding 1 — concurrent inline rows (flow-fix, blast radius 2).**
<!-- measured: Task 1's **Files:** names 2 @ 95deb422 -->
`verify-agents` and `verify-gymie` were begun in one Bash call, 30 ms apart, and both ended
together; the store gave `verify-gymie` 7,178,574 cache-read tokens and `verify-agents` 145,107
<!-- measured: the pass report's begin times 22:47:53.783 and 22:47:53.813 and flow record dispatches -change self-review-backlog-flow-6 @ 95deb422 -->
(`flow record dispatches -change self-review-backlog-flow-6`).
`DispatchWindowsForSession`'s window for an open stage-less
row is `MIN(n.started_at)` over every later inline begin in the session, so while both rows were
open the first covered 30 ms and every record harvested in real time went to the second.
<!-- measured: the planned test at 95deb422 prints the first window ending 30 ms after its begin -->
`minInlineWindow` (3 s) is already the shortest window the store lets an inline row close with, so
a begin closer than that cannot be a successor that took the main thread over. Reproduced at
`95deb422`: the planned test fails without the fix, the first window capped at the sibling's begin
30 ms in; with the fix it passes, and `go test ./internal/store/ ./internal/harvest/` stays green.
<!-- measured: go test ./internal/store/ -run TestDispatchWindowsForSession -count=1 with and without the records.go hunk @ 95deb422 -->
The already-recorded skewed figures are not rewritten.

**Finding 2 — a task created a worktree (flow-fix, blast radius 1).** The archived
`self-review-backlog-flow-6` `tasks.md` Task 5 Step 1 ran `git worktree add … -b
self-review-backlog-flow-6 origin/develop`; the run used `kickoff-worktree.sh --from` instead,
branch `spectre/self-review-backlog-flow-6`, and the plan needed a dated Correction.
`grep -n "git worktree add" skills/flow/brainstorm-planner.md` finds nothing at `95deb422`.

**Finding 3 — hand-built verify runner (flow-automation, blast radius 4).** **Inline verify**
(`skills/flow/verify-and-handoff.md`) prescribes the list, its order, the stdin detach, the
no-stop rule and the `## Report` line shape in prose, and no script runs it; the last run's parent
wrote one in its scratchpad and patched the stdin detach in on a second call. Spiked at
`95deb422`: the planned runner's harness passes, fails on the stdin case with the `</dev/null`
removed, and the runner run over this worktree's own `## lint` list reports every command, exit 1
only on `check-plan-shape.sh` against this change's then-unwritten stub `tasks.md`.

## Decisions

### Sibling discriminator for the inline cap

**ID:** inline-sibling-min-window
**Status:** active
**Chosen:** an open stage-less row is capped only at a later inline begin at least `minInlineWindow` after its own, passed to the query as `$2` from the Go constant — the store's own floor for a real inline window, so a closer begin is a sibling begun beside it.
**Considered:** dropping the cap — an abandoned inline row would again take every later main-thread record of the session; matching the `verify-` key prefix — a naming convention the store does not own, and any other parallel inline pair stays broken; a different session token on the later row — the two verifier rows already carry different tokens, so it tells nothing apart.

### The worktree rule is prose

**ID:** worktree-rule-prose
**Status:** active
**Chosen:** one blockquote rule in section D of `skills/flow/brainstorm-planner.md`, beside the other "Write …" task rules, citing **2. Isolate the workspace** (`skills/flow/implement.md`); its new sentence is listed in this change's `verbatim-moves.txt`.
**Considered:** a `check-plan-shape.sh` refusal of `git worktree add` in a live plan — plans that build or test worktree tooling (`kickoff-worktree.sh`, `throwaway-worktree.sh`, harness fixtures) carry the command legitimately, and only one archived plan ever needed the correction, so a guard buys false refusals for a one-off slip.

### The runner is bash

**ID:** runner-is-bash
**Status:** active
**Chosen:** `scripts/run-verify-list.sh`, bash, symlinked into `skills/flow/scripts/` — it judges nothing about the tree, it runs the project's configured commands, the same kind of orchestration `project-get.sh` and `prepare-workspace.sh` are; the new-guards-in-Go rule binds guard logic.
**Considered:** a `flow-guard` subcommand behind a `flow_guard_exec` shim — a guard's verdict machinery for a script that has none, and a build of `flow-guard` before every verify list.

### Runner contract

**ID:** runner-contract
**Status:** active
**Chosen:** exit 0 all green, 1 any command non-zero, 2 usage or `project-get.sh` exit 2, 3 the list is not declared (the caller auto-detects); each command in a fresh `bash -c` from the worktree with stdin detached to `/dev/null`; output past 20 lines printed as a `(truncated — last 20 of N lines)` line and the last 20; no retry — the one inline re-run of a failing command stays the caller's, as **Inline verify — a failing command** states.
<!-- measured: bash scripts/test-run-verify-list.sh with the Task 3 diff applied @ 95deb422 → all cases passed -->
**Considered:** folding "not declared" into exit 2 — the caller could not tell "auto-detect" from "broken config"; retrying a failed command inside the runner — the rule says one inline re-run judged by the caller, after the whole list.

### Inline verify keeps its prose

**ID:** inline-verify-keeps-prose
**Status:** active
**Chosen:** one added sentence naming the runner as the two groups; the paragraphs stating the list, its order and the report shape stay, as the contract the runner's header points back at.
**Considered:** replacing those paragraphs with the citation — deletes normative sentences (`check-verbatim-moves.sh` FAILs) and makes the script's header the only statement of what the verify stage runs.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — a new script every verify stage will run, with shell quoting and exit-code edges, plus a change to the token-attribution query the stats app reports from.
**Considered:** `flow-low` — every task is small and test-pinned, but a quoting slip in the runner or an off-by-one in the cap reaches every later run.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-self-review-backlog-flow-6-fix`, Tasks 1, 2 and 3 in order.
**Considered:** one fixer per task — three dispatches for three file-disjoint, pre-verified diffs.

## Open questions

## Live check

Finding 1 changes what the running `flowd` attributes. Against the dev store, after this lands and
`flowd` is rebuilt by the operator's usual route: in a run whose verify stage has two worktrees,
`flow record dispatches -change <that change>` shows both `verify-<repo>` rows with cache-read
figures of the same order, each carrying `apportioned` records. Not working looks like the
first-begun row holding only a small fraction of its sibling's tokens, as `verify-agents` did here.
Findings 2 and 3 have no runtime of their own: prose, and a script whose harness is its check.
