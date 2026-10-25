## Context

Operator's request, verbatim: "update flow to always sync before implementation". Evidence from
kan-1032: the base moved 13 commits during planning, overlapping a planned file; the review panel's
`sync-panel-base.sh` hit exit 3 (overlap) and stopped for an operator answer, and the rebase pulled
in a new self-review report that broke a guard.

What runs before task 1 today, in `skills/flow/implement.md` section 1: `check-base-behind.sh`
(syncs a recorded `--base` with the branch it tracks, never the change branch) and
`refresh-plan-base.sh` (names what moved; its paragraph ends "The step names, it never rebases").
`/flow-fast` section 4 runs neither and cites only `implement.md`'s section 4.

What already does the rebase: `sync-panel-base.sh [--rebase] <worktree> <merge-base>`
(`stats/internal/guard/syncpanelbase.go`). Under `--rebase` it fetches, rebases onto
`origin/<base>`'s tip on any `MOVED` verdict, overlap or not, prints `REBASED: <wt> — merge base
<sha>` per clean rebase, and exits 0 settled, 1 `CONFLICT` (worktree left mid-rebase), 3 only when
the base moved again with an overlap after a rebase, 2 otherwise — including a rebase git refuses
over a dirty tree (kan-1030).
<!-- measured: grep -n 'n > 0 || !rebase' stats/internal/guard/syncpanelbase.go @ dd2268bd → line 56 -->
The review panel's **Check base movement first** (`skills/flow/review-panel.md`) owns the
conflict handling this step reuses.

Pushing the rewritten branch: `hooks/protect-main-checkout.py` judges a push to `spectre/<name>`
with `check-untasked-commits.sh --not-remotes`, which drops every commit reachable from an origin
ref, so the base's own commits a rebase brings in are never listed. Before task 1 the branch carries
only `chore(spectre)` planning commits, which that guard passes.

Blast radius: 9 files cite `refresh-plan-base` or `check-base-behind`.
<!-- measured: grep -rl 'refresh-plan-base\|check-base-behind' skills scripts stats rules hooks | wc -l @ dd2268bd → 9 -->
The fix touches five files: `skills/flow/implement.md`, `skills/flow-fast/SKILL.md`,
`skills/flow-contracts/git-boundaries.md`, `scripts/refresh-plan-base.sh` and a new
`stats/internal/repotest/base_sync_before_task_one_test.go`. No saved memory encodes the
defect. Not `big`.

Task 1's diff was applied to this worktree at `dd2268bd` and checked there: the new test passes
with it and fails with the skill edits reverted; `check-references`, `check-verbatim-moves` (with
this change's `verbatim-moves.txt`), `check-markdown-integrity`, `check-guard-symlinks`,
`check-dispatch-paragraphs`, `check-installed-citations`, `check-normative-inventory`,
`check-ask-silence`, `check-hand-notes-in-step`, `gofmt -l` and `go vet ./internal/repotest` exit 0.

## Decisions

### Which script rebases

**ID:** reuse-sync-panel-base-rebase
**Status:** active
**Chosen:** `sync-panel-base.sh --rebase <worktree> <merge-base>` — it already fetches, resolves the base (recorded `--base` or `origin/HEAD`), rebases on any movement, prints the new merge base and stops on a conflict without resolving it.
**Considered:** `sync-onto-base.sh` — integrate's sync, which resolves conflicts in place, where the requirement is a stop for the operator; a new script — a second copy of what `sync-panel-base.sh --rebase` does; making `refresh-plan-base.sh` rebase — it would then name movement against the rebased tip and lose its collision report.

### Where the step sits

**ID:** after-plan-refresh
**Status:** active
**Chosen:** in `implement.md` section 1, right after the plan-refresh paragraph and before the task-checkbox read — the refresh still names what moved against the old merge base, then the branch is rebased.
**Considered:** section 2 (isolate workspace) — the resolved set is final there, but section 1's base checks already run over it, and the step belongs with them; before the refresh — the refresh would then always print `UNMOVED:`.

### Exit codes

**ID:** stop-on-conflict-and-unsettled
**Status:** active
**Chosen:** exit 1 handled by the panel's **Conflict** bullet and the run stops for the operator; exit 2 or 3 stops the run naming its output.
**Considered:** re-running on exit 3 with the newest merge base — exit 3 under `--rebase` means the base moved again, overlapping, mid-sync, which is rare enough that a stop costs less than a retry loop.

### Recording and pushing

**ID:** record-and-force-push
**Status:** active
**Chosen:** after a `REBASED:` line, `flow state add-worktree <name> <worktree> <sha>` and `git -C <worktree> push --force-with-lease origin spectre/<name>`; **Branch backup** gains the base-sync rebase as a `--force-with-lease` push.
**Considered:** leaving the push to the next task commit — a plain push is then rejected non-fast-forward; a plain push — rejected for the same reason.

### `/flow-fast`

**ID:** flow-fast-same-step
**Status:** active
**Chosen:** `/flow-fast` section 4 cites the step, with no `flow state add-worktree` call (it keeps no state file) and `<name>` as the pushed branch, and its `<merge-base>` substitution follows the last `REBASED:` sha so the panel's later sync compares against the rebased base.
**Considered:** leaving `/flow-fast` out — its branch is cut minutes before implementation, but the operator's "always" covers it and the cost is three lines.

### Test

**ID:** doc-pin-test
**Status:** active
**Chosen:** `TestBaseSyncBeforeTaskOne` in `stats/internal/repotest/` reads `implement.md` section 1 and `/flow-fast` section 4 and fails unless the step, its command lines and the `/flow-fast` citation are there and the "it never rebases" deferral is gone.
**Considered:** a new `scripts/check-*` guard — a permanent lint line for a two-file pin; a behaviour test of `sync-panel-base.sh --rebase` — it already exists (`sync_panel_base_test.go`) and does not fail without this fix.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — a pipeline behaviour change across two skills, with exit-code handling and a history-rewriting push the reviewer must check against the hook and the guard.
**Considered:** `flow-low` — the diff is small and the plan gives it literally, but the push and conflict semantics are judgements, not mechanics.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

None.
