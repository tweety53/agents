# kan-404-flow-keep-the-mutation-testing-brief-as-a

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

**Spec:** `design.md` beside this file. Both tasks are `Build: green` with their own commit and are
independent of each other: task 1 is the store and CLI, task 2 is documentation. `skills/flow/SKILL.md`
needs no edit — its "Slots dispatched by `subagent_type` (Bugbot, Security)" sentence stays true,
since `mutation` is dispatched general-purpose.

- [x] 1. Accept `mutation` as a reviewer slot

**Build:** green
**Files:** `stats/internal/store/settings.go`, `stats/internal/store/settings_test.go`, `stats/cmd/flow/settings.go`
**Tests:** `TestSettingsStore_RoundTrip` — the existing case, its `Reviewers` list extended with `mutation`
**Regression:** reverting this commit makes `PutSettings` return `ErrInvalidReviewer: "mutation"` for the round-trip value, so `TestSettingsStore_RoundTrip` fails at `PutSettings:`.
**Baseline:** before=6 after=6
<!-- measured: cd stats && go test ./internal/store/ -run TestSettingsStore -v 2>&1 | grep -c '^--- PASS' @ branch main -->
<!-- predicted: cd stats && go test ./internal/store/ -run TestSettingsStore -v 2>&1 | grep -c '^--- PASS' after this task, on branch spectre/kan-404-flow-keep-the-mutation-testing-brief-as-a -->
**Commit:** `feat(stats): accept mutation as a reviewer slot`

  - [x] **Step 1: Extend the round-trip test.** In `stats/internal/store/settings_test.go`, change
    `TestSettingsStore_RoundTrip`'s `Reviewers` value to
    `[]string{"primary", "principles", "code-review-low", "bugbot", "mutation"}`. Run
    `cd stats && go test ./internal/store/ -run TestSettingsStore_RoundTrip` — it fails with
    `store: invalid reviewer: "mutation"`.
  - [x] **Step 2: Add the id.** In `stats/internal/store/settings.go`, add `"mutation": true,` to
    `ValidReviewers` after `"bugbot"`, and change the comment's "these five ids" to "these six ids".

```go verified:read at stats/internal/store/settings.go:26-36 on main
// ValidReviewers is the fixed vocabulary a flow_settings.reviewers entry
// may take. The panel dispatches exactly the resolved list; these six ids
// no longer split into a required subset and an on-demand-only subset
// (design.md's roster-from-settings decision superseded that split).
var ValidReviewers = map[string]bool{
	"primary":         true,
	"principles":      true,
	"code-review-low": true,
	"bugbot":          true,
	"security":        true,
	"mutation":        true,
}
```

  - [x] **Step 3: Name the six ids in the CLI help.** In `stats/cmd/flow/settings.go`, the
    `-reviewers` flag's usage string becomes
    `"comma-separated reviewer slots, from primary,principles,code-review-low,bugbot,security,mutation (required)"`.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./internal/store/ ./cmd/flow/`
    — `gofmt -l` prints nothing, both exit 0. Commit.

- [x] 2. Add the `mutation` slot to the review panel

**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow-contracts/artifacts-registry.md`, `scripts/check-cleanup-complete.sh` (its `registry-row-not-checked:` marker for the renamed registry row, found stale by `flow.verify`'s `run-guard-tests.sh` and fixed there)
**Tests:** none — documentation; the shipped guards below are the check
**Regression:** none — no test is added; `scripts/check-references.sh` and `scripts/check-vocabulary.sh` are the mechanical checks and pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose; the count is the number of tests this task touches @ branch main -->
<!-- predicted: unchanged after this task, on branch spectre/kan-404-flow-keep-the-mutation-testing-brief-as-a -->
**Commit:** `docs(review-panel): add the mutation slot beside bugbot`

  - [x] **Step 1: The roster.** In `skills/flow/review-panel.md`'s roster table (after the
    `security` row), add:

```markdown verified:column layout copied from the roster table at skills/flow/review-panel.md:104-110 on main
| `mutation` | **Mutation** — sabotage-proofing | general-purpose + the mutation-testing brief below, own throwaway worktree copy per repository (see **The throwaway worktree** below) | `DEFAULT_MODEL` |
```

    Change "`ValidReviewers` … is the id vocabulary this table exhausts — five entries, never a
    sixth" to "six entries, never a seventh".
  - [x] **Step 2: The shared brief.** Retitle `### Bugbot's mutation-testing brief` to
    `### The mutation-testing brief`; open its paragraph with "Wherever the panel dispatches Bugbot
    or Mutation, the dispatch prompt carries a mutation-testing brief:" and extend the list to
    "flip a condition, drop a guard, move a boundary, remove a branch, move an interaction off its
    target, overlay an earlier commit's tree and run the tests".
  - [x] **Step 3: The shared worktree.** Retitle `### Bugbot's throwaway worktree` to
    `### The throwaway worktree`; state that Bugbot and Mutation both mutate code in place and both
    run there; replace every `<worktree>-bugbot-<round>` in that section with
    `<worktree>-<slot>-<round>` and say `<slot>` is the id (`bugbot` or `mutation`), so a roster
    carrying both produces two copies per repository per round. "Dispatch Bugbot **once**" becomes
    "Dispatch each of the two slots **once**, its prompt listing every copy made for that slot".
    The removal step and the reproducer paragraph stay as written.
  - [x] **Step 4: The fix round.** In the re-run rule, "**Bugbot and Security**, which read no diff
    file" becomes "**Bugbot, Mutation and Security**, which read no diff file"; in the docs-only
    reclassify paragraph, "Bugbot and Security among them take their pass-1 shape, throwaway
    worktree included" becomes "Bugbot, Mutation and Security among them …"; in the stale-result
    paragraph, "A non-Minor fix Bugbot did not raise leaves Bugbot's result current" becomes "A
    non-Minor fix Bugbot or Mutation did not raise leaves that slot's result current"; in the closing
    paragraph of **The fix round mutation-proves what it changed**, "a run where Bugbot is neither
    in the resolved roster nor added this run" becomes "a run where neither Bugbot nor Mutation is in
    the resolved roster or added this run". Under **An unspawnable id is substituted, not
    skipped**, the sentence "A substituted Bugbot's dispatch runs against the same throwaway
    worktree treatment as the real slot (**Bugbot's throwaway worktree**)" points at **The
    throwaway worktree**; every other cross-reference to the two retitled sections follows
    (`grep -n "Bugbot's mutation-testing brief\|Bugbot's throwaway worktree" skills/ rules/` must
    print nothing).
  - [x] **Step 5: The registry.** In `skills/flow-contracts/artifacts-registry.md`, the row
    "Bugbot's throwaway worktree copy" becomes "Bugbot's or Mutation's throwaway worktree copy",
    its path `<worktree>-<slot>-<round>`, its removal "immediately after that slot's dispatch closes".
  - [x] **Step 6: Verify.** From the repository root run `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh`, `scripts/check-normative-inventory.sh` and
    `scripts/check-contract-budget.sh` — each exits 0. Commit.
