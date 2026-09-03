## Context

`check-base-moved.sh` exists and runs at integrate run 1 step 2 (`skills/flow/integrate.md`),
after the unfinished-work gate and before the landing question, with a Stop / Rebase now / Continue
prompt on overlap (KAN-371). It performs no fetch of its own; `resolve-base-branch.sh` fetches when
the caller resolves `$BASE`. The review panel (`skills/flow/review-panel.md`) runs before that, on
every implementation run, and opens with `check-panel-citation-trigger.sh <worktree> <merge-base>`
and a dispatch-context rebuild — both reading the diff against a merge base nothing has checked. On
a creating run the state file's `worktrees` map does not exist yet; the merge base lives in the
run's working notes (`skills/flow/implement.md`, isolate-workspace) until `verify-and-handoff`
writes it. The pipeline has no detectable notion of a "recorded baseline artifact" — KAN-265's was
its own inventory byte budgets. `spectre/specs` is empty, so no capability spec is altered.

## How

### A base-moved check at panel entry

A new first step in `skills/flow/review-panel.md`, under the `flow.review-panel` begin mark and
before the citation pre-check, on every panel run (creating, resumed, fix). Once per worktree in the
run's resolved set:

```bash unverified:the same pair integrate step 2 already runs, not yet exercised from the panel by this change
BASE="$(resolve-base-branch.sh <worktree>)"
check-base-moved.sh <worktree> "origin/$BASE" <working-notes-merge-base>
```

Verdict handling is integrate step 2's, cited rather than restated where the file already states
it: every verdict is reported; `MOVED` with no overlap continues with no prompt; `REFUSE`, exit 2
with no verdict line, or an empty resolved set stops and asks; an overlap from any worktree asks
once for the whole change:

```markdown unverified:the literal prompt is integrate step 2's; the implementer copies it when task 1 writes the step
**The base branch has moved and touches paths this change also touched — how should the panel
proceed?**
- **Stop — I'll rebase or reorder first** *(recommended)*
- **Rebase onto `<base>` now, then continue**
- **Continue — review as is**
```

**Stop** closes `flow.review-panel` with `-outcome stopped`, leaving the change at its current
state with nothing committed by this stage. **Continue** carries the reported movement into the
handoff and proceeds to the citation pre-check.

### On the rebase choice, at panel entry

`git -C <worktree> rebase origin/$BASE` only in a worktree whose own verdict was `MOVED`, never one
whose verdict was `CLEAR`.

- **Clean** — that worktree's working-notes merge base becomes `origin/$BASE`'s resolved tip at
  rebase time. Every later `<merge-base>` in this file, and the `worktrees` map
  `skills/flow/verify-and-handoff.md` writes, read the working notes, so the new value propagates
  without further plumbing. Re-run `check-base-moved.sh` once against the new value; a fresh
  overlap re-offers the prompt rather than looping silently. A rebase rewrites every sha on the
  branch, so it clears every slot's held last-reviewed sha: a re-run after a rebase reads the whole
  `final-review.diff` under the existing no-held-sha rule in **Panel re-runs**. No scoped
  re-verification runs here — the panel reads the rebased tree and `flow.verify` follows.
- **Conflict** — never auto-abort and never resolve. Leave the worktree mid-rebase, report the
  conflicting files from `git status`, hand off `git -C <worktree> rebase --continue` (after the
  operator resolves) or `git -C <worktree> rebase --abort`, and close the mark `stopped`.

### The baseline sentence

After a clean rebase at either site — the panel-entry step above and integrate step 2's clean
branch — the report ends with one fixed literal:

> If this change's verification compares against a recorded baseline, recapture it now — a proof
> taken against the pre-rebase base is void.

Advisory only: nothing detects a baseline artifact.

### Unchanged

Integrate step 2 keeps its check as the later net. `skills/flow-contracts/finish-contract-run1.md`
stays canonical for it and is not edited. No script changes; `scripts/test-check-base-moved.sh`
already covers the guard.

## Decisions

### Where the new check lives

**ID:** panel-entry-only-integrate-kept
**Status:** active
**Chosen:** Add the check at panel entry and keep integrate step 2's — the panel reviews a fresh
base, and the base can still move between panel close and merge (KAN-304, the case KAN-371 fixed).
**Considered:** Move the check from integrate to panel entry — rejected: reopens the KAN-304 gap.
Integrate-only wording change, treating the issue body as the spec — rejected: KAN-371 already
shipped it, leaving the title's trigger unaddressed.

### The recorded-baseline reminder

**ID:** one-fixed-baseline-sentence
**Status:** active
**Chosen:** One fixed advisory sentence after a clean rebase at both sites — cheap, honest about
being undetectable.
**Considered:** Drop it as a KAN-265 one-off — rejected by the operator: the cost it warns about is
a full re-derivation. Detect a baseline artifact mechanically — rejected: the pipeline has no such
concept, and inventing one is speculative.

### Held slot shas after a rebase

**ID:** rebase-clears-held-slot-shas
**Status:** active
**Chosen:** A panel-entry rebase clears every slot's held last-reviewed sha; slots read the whole
diff under the existing no-held-sha rule.
**Considered:** Keep the held shas and diff `<held-sha>..HEAD` — rejected: the rewritten history
makes that diff include upstream commits, so a scoped re-read would review the wrong delta.

### Re-verification after a panel-entry rebase

**ID:** no-scoped-reverification-at-panel-entry
**Status:** active
**Chosen:** None — the panel reads the rebased tree and `flow.verify` runs after it.
**Considered:** Copy integrate step 2's scoped guard-test re-run — rejected: it exists there because
integrate has no verification gate; the panel stage is followed by one.

## Open questions

None — the design converged in one round with no deferred question.
