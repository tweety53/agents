## Context

A standalone self-review fix loop over five backlog tickets, in worktree
`/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-1021` on branch
`self-review-backlog-flow-1021`, based at `1c4c3e38` (origin/main). Each finding was checked
against that commit, its history and the dev store (read-only `SELECT`s) before its fix was
designed.

**KAN-910 — mostly fixed upstream.** Inline `full-suite-fix-<n>` rows are skipped since
`ca90e129`; a subagent's row under that key is still refused on purpose — `ca90e129` records the
KAN-482 reason (the parent runs the full suite, never a subagent), and this design does not
revisit it. Gated fix rounds (`task-<n+n>-implementer-fix-<k>`) were recorded under `panel-fix`
while `skills/flow/gated-review-fix.md` named no role; `6b2020bf` made it `-role implementer`,
and the store's last such row under `panel-fix` began before that commit landed. What is left: no
test pins the accepted shapes, and the key rule in `skills/flow/review-panel-fix-round.md` and
the guard's bash header name only the full-suite exemption.

**KAN-911 — three edges, all live.** (1) `resolvebasebranch.go` returns a recorded `flowBase`
without asking origin whether it still exists; its fetch does not prune, so a deleted base keeps
resolving off a stale `origin/<b>`, and a landing push would recreate it. `git ls-remote
--exit-code --heads origin refs/heads/<b>` tells the cases apart — spiked in a scratch repository:
a present branch exits 0, a deleted one 2, an unreachable remote 128.
<!-- measured: git ls-remote --exit-code --heads origin refs/heads/feature/x in a scratch bare-remote clone, before and after deleting the branch, and against /nonexistent @ git 2.x on this machine, 2026-10-21 -->
(2) `/flow-self-review` step 1 refuses every branch but `<default-branch>`; context bundles were
committed in the days after `flowBase` landed (`f11cbfd0`), so a `--base` change's bundle can sit
on its base. (3) `integrate-change.sh cleanup` passes each main checkout `origin/HEAD`'s branch,
so a main checkout on the change's base is refused `is on <b>, not <default>`.

**KAN-912.** `insertDispatch`'s `ON CONFLICT … DO UPDATE` writes only `dispatch_key` and, on an
aborted re-begin, `started_at`/`ended_at`/`cause`/`outcome`/`agent_id`; `slot`, `model` and
`effort` keep the first begin's values.

**KAN-915.** The rendered REPORT FILE line already carries the re-run's own round and id —
`render-slot-prompt.sh` substitutes both, pinned by `render_slot_prompt_test.go`. Nothing checks
the file afterwards: `check-panel-findings-closed` counts any later `completed` reviewer row of
the slot as the clean re-run, and the parent's own fallback writes `no verbatim report captured —
<reason>` into a missing report (**Recording findings**, `skills/flow/review-panel.md`).

**KAN-906.** An inventory of `stats/internal/guard/` — every top-level function referenced in
code by another non-test file — found the helpers Task 5 moves: each is defined in one guard's
own file (one that registers a guard in `Registry`) and called from three or more other files, or
is a generic helper in a shared file named for something else.

## Decisions

### Gated and verify fix rounds keep their own role

**ID:** non-panel-fix-roles-stated
**Status:** active
**Chosen:** pin the accepted shapes in a table test and state, at the key rule in `review-panel-fix-round.md` and in the guard's bash header, that a gated fix round records under `-role implementer` and an in-run verify fix under its appended task's implementer pair — the guard's code stays as it is.
**Considered:** widening the guard to skip `task-<ids>-implementer-fix-<k>` rows under `panel-fix` — a row recorded that way now contradicts `gated-review-fix.md`, and skipping it would hide the misrecord; skipping subagent `full-suite-fix-<n>` rows too — rejected in `ca90e129` for KAN-482, and nothing since changes that reason.

### A recorded base gone from origin is refused

**ID:** recorded-base-gone-refused
**Status:** active
**Chosen:** `resolve-base-branch` asks `git ls-remote --exit-code --heads origin refs/heads/<base>` when the base came from `flowBase`; its exit 2 is a named refusal, the guard's exit 1; any other result — an unreachable remote included — resolves as today. The recovery — re-point every worktree with the three commands **`STOP: base-mismatch`** names — is one sentence in `skills/flow/SKILL.md`.
**Considered:** documenting the recovery alone — nothing tells the operator the base is gone, and the landing push recreates it; `fetch --prune` then reading `refs/remotes/origin/<b>` — prunes every stale ref of the repository as a side effect of a read-only resolution; checking an `origin/HEAD`-resolved base too — `origin/HEAD` always names a live branch.

### The landing branch is derived from where the bundle landed

**ID:** self-review-landing-branch
**Status:** active
**Chosen:** `/flow-self-review` step 1 accepts `<landing-branch>`: `<default-branch>`, or a branch other than the change's own whose `origin/<branch>` tree holds the bundle — where the bundle landed is where the change landed. The absent case lists every `origin/<branch>` holding the bundle. Step 7 lands and pushes onto `<landing-branch>`, and the guardrail allows it.
**Considered:** reading `branch.spectre/<name>.flowBase` — the branch and its config are gone once run 2 removed the worktree, which is when a standalone pass runs; reading the bundle from `origin/<base>` without a checkout — step 7's landing script commits in the checkout it is given.

### Run 2 refreshes onto the base the main checkout has out

**ID:** refresh-onto-checked-out-base
**Status:** active
**Chosen:** `cmd_cleanup` keeps each worktree's base from its merge check, and a main checkout whose current branch is that base is refreshed onto it; every other main checkout stays on `origin/HEAD`'s branch.
**Considered:** always refreshing onto the change's base — a main checkout on the default branch, the common case, would then be refused `is on main, not <b>`; teaching `refresh-main-checkout` to follow the current branch — it would fast-forward whatever branch is out, the change's or not.

### A repeated begin updates slot, model and effort

**ID:** rebegin-last-wins
**Status:** active
**Chosen:** the conflict clause sets `slot`, `model` and `effort` from the new begin on every repeat, aborted re-begin or not; a journal replay carries the same values, so it stays a no-op.
**Considered:** refusing a repeat begin whose values differ — an aborted loop key is re-dispatched on the same key, and its new pair is the truth; updating only on an aborted re-begin — an open row re-begun on a new pair keeps the stale one, the reported defect.

### No repair of past rows

**ID:** rebegin-no-repair
**Status:** active
**Chosen:** no rows are rewritten — a re-begin leaves no trace in the row, so the rows it left wrong cannot be told apart from rows misrecorded for any other reason.
**Considered:** rewriting every subagent row whose recorded model disagrees with `signals.sidechain.served_models` — the dev store holds some, but that rewrites misrecords of other causes under this defect's name.

### A re-run counts once its report is written

**ID:** rerun-needs-report
**Status:** active
**Chosen:** `check-panel-findings-closed` counts a reviewer row as the clean re-run of a slot component only when `.superpowers/sdd/panel-report-<round>-<id>.md` for that row's round and that component is non-empty and does not open with the parent's `no verbatim report captured` placeholder; a fixed Critical or Important left unverified that way gets a second stderr line naming each missing report file. The REPORT FILE paragraph stays as written.
**Considered:** restating the path in the REPORT FILE paragraph — the rendered line already names the exact path; checking every dispatch's report — a first-pass slot's report is already compared by the count check, and only the re-run is the verification `fixed` rests on.

### Cohesion moves only

**ID:** guard-helpers-move-only
**Status:** active
**Chosen:** move each listed helper, its doc comment and any declaration only it uses, unchanged, into a shared file named for what it does; no renames, no merges of near-duplicates (`pcAbs` cleans its path where `smcAbs` and `wiAbs` do not), and every `(<old file>.go)` citation of a moved name is updated.
**Considered:** renaming the guard-prefixed helpers (`pcAbs` → `absPath`) — touches every call site for no behaviour; merging the near-duplicates — changes behaviour where they differ.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — a store upsert every dispatch goes through, a guard semantic that can fail every panel close, and a refusal in the base resolution every phase calls.
**Considered:** `flow-low` — each task is small and test-pinned, but a wrong condition in the base refusal or the re-run check stops every run, not one.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-backlog-flow-1021-fix`, Tasks 1 to 5 in order, the move last so it runs after Tasks 1 and 4 edit files it moves helpers out of.
**Considered:** one fixer per task — five dispatches for small tasks, and Task 5 would still wait on Tasks 1 and 4.

## Open questions

## Live check

Task 3 changes what the running `flowd` writes. After this lands and the operator rebuilds
`flowd` by the usual route, against a throwaway change in a throwaway project: `flow record
dispatch begin -role reviewer -model opus -effort high -slot primary -key live-rebegin …`, then
the same key with `-model sonnet -effort medium -slot principles`; `flow record dispatches -change
<it>` shows one row carrying `sonnet`, `medium`, `principles`. Not working looks like the row
still on `opus`, `high`, `primary`. Tasks 1, 2, 4 and 5 change guards and scripts whose suites
are their check; nothing else holds runtime state.
