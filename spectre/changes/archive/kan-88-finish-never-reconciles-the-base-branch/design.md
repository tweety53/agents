# KAN-88 — reconcile the base branch in `/flow`'s integrate/archive phase

## Context

**Why this change exists, and what it changes:** `proposal.md`, beside this file — canonical
for both, and not restated here.

## What already landed

Half 1's ticket option 1 is **already in force**: `finish-contract.md`'s preflight section composes
`<base-ref>` as `origin/$BASE`, and `integrate.md` passes it that way. What that left undone is the
ticket's own Verification item — no regression case covers it — and the correctness is the caller's
rather than the guard's.

## Decisions

### Half 1 hardening — the guard resolves the ref it was handed

**ID:** preflight-resolves-remote-tracking
**Status:** active
**Chosen:** ticket option 2, plus the ticket's regression case — `check-finish-preflight.sh`
prefers `refs/remotes/origin/<base-ref>` when that ref resolves, and every verdict line names the
ref actually used. The script becomes correct by construction rather than by the caller getting it
right.
**Considered:** *option 1 alone (already landed)* — leaves a future caller free to reintroduce the
bug silently, and leaves the ticket's Verification item unmet. *Adding option 4's freshness
reporting* — composes with this and would have made all three occurrences self-diagnosing, but
`origin/$BASE` plus self-defence removes the staleness class the reporting would describe; deferred
rather than rejected on principle. *Option 3 (fast-forward the local base in run 2)* — rejected by
the ticket itself: it leaves the script correct by luck.

### The correct caller must be unaffected

**ID:** remote-lookup-is-a-preference-not-a-rewrite
**Status:** active
**Chosen:** look up `refs/remotes/origin/<base-ref>` and use it only if it resolves; otherwise use
the argument exactly as handed. `origin/main` resolves `refs/remotes/origin/origin/main`, finds
nothing, and passes through unchanged, so the contract's own call site keeps its current behaviour
and the existing test cases — which build repositories with no `origin` remote — keep their current
expectations.
**Considered:** *refusing a bare local name outright* — turns a recoverable, correct-by-substitution
case into a stop, and would break any harness whose contract copy predates the `origin/$BASE`
composition. *Stripping and re-prefixing (`sed 's#^origin/##'` then `origin/`)* — indistinguishable
from the above for every input this pipeline produces, but mangles a legitimately-named local branch
called `origin/...`.

### Half 2 is a shipped guard, not inline prose

**ID:** base-moved-is-a-guard
**Status:** active
**Chosen:** `scripts/check-base-moved.sh`, symlinked into `skills/flow/scripts/`, with
`scripts/test-check-base-moved.sh` beside it — matching how every other run-1 signal in this
pipeline is built.
**Considered:** *inline git commands in the contract* — smaller diff, but no mutation test, and the
overlap computation would live in prose an agent re-derives on every run.

### The prompt fires on overlap, not on movement

**ID:** ask-only-on-overlap
**Status:** active
**Chosen:** base moved with no overlapping paths is reported and the run continues to the landing
question; base moved **and** touching paths this change also touched reports and asks.
**Considered:** *prompting on any movement* — never misses a semantic conflict that touches no
shared path (a renamed symbol, a changed contract), at the cost of prompting on nearly every run
against an active base, which trains the operator to dismiss it.

### One aggregated prompt across worktrees

**ID:** aggregate-the-multi-repo-ask
**Status:** active
**Chosen:** report per worktree, ask once for the whole change — the shape the unfinished-work gate
already uses, and the shape KAN-129's actual decision took (repo ordering across the change).
**Considered:** *one prompt per affected worktree* — finer control the downstream landing routes do
not support; run 1 cannot stop for one repo and continue for another.

### The change's own paths include uncommitted work

**ID:** touched-paths-include-index-and-worktree
**Status:** active
**Chosen:** the union of `<recorded-merge-base>..HEAD`, the index and the working tree. Run 1 is
reached with work staged and uncommitted by design, so anything narrower reports an empty overlap
for a change that does conflict.
**Considered:** *committed only* — simplest and deterministic, and wrong on the normal
`IN_PROGRESS` shape.

### It runs under the existing landing-question mark

**ID:** no-new-stage-key
**Status:** active
**Chosen:** the check runs at the top of `integrate.md` step 2, under the existing
`flow.landing-question` stage mark.
**Considered:** *its own step and `flow.base-moved` key* — cleaner stage telemetry, but no `flow.*`
key is in `stats/internal/stages/names.go` yet, so every new one widens that already-disclosed gap.

### Verdict protocol copied from its siblings

**ID:** verdict-protocol-matches-siblings
**Status:** active
**Chosen:** one verdict line to stdout, `CLEAR:` / `MOVED:` / `REFUSE:`, exit 0 whenever a verdict
was reached and exit 2 when the tree cannot be read — identical to `check-finish-preflight.sh` and
`check-unfinished-work.sh`, with the breakdown carried inline on the verdict line as
`check-unfinished-work.sh` already does. The guard performs no fetch of its own:
`resolve-base-branch.sh` fetched that worktree when the caller resolved `$BASE`.
**Considered:** *fetching inside the guard* — a second fetch per worktree per run, for a ref the
caller just refreshed. *An exit-code-only protocol* — makes "cannot determine" indistinguishable
from "the base moved", the confusion `check-finish-preflight.sh`'s header already records.

## Open questions

None.

## How

### `scripts/lib/resolve-remote-base.sh`

Both guards need the same resolution, so it is a shared library function rather than two copies:
given a worktree and a base ref, print `origin/<ref>` when `refs/remotes/origin/<ref>` resolves to a
commit, otherwise print the ref unchanged. Resolution uses `rev-parse --verify --end-of-options`, so
a ref beginning with `-` is read as a ref rather than parsed as a git option.

### `scripts/check-finish-preflight.sh`

The effective ref is computed once, before signal 2's existing "does the base ref resolve" check, so
an unresolvable effective ref is still its own named `REFUSE` rather than an accidental `RUN1`.
Signal 1 is untouched — it reads only `HEAD` and the recorded merge base, and is answered before any
base ref is resolved. Every verdict line naming a base ref names the effective one.

### `scripts/check-base-moved.sh`

```text unverified:the guard does not exist yet; these are the literals tasks.md's task 2 asserts against
Usage: check-base-moved.sh <worktree> <base-ref> <recorded-merge-base|->

CLEAR:  <worktree> — <ref> has not moved since the recorded merge base
MOVED:  <worktree> — <n> commits on <ref> since the recorded merge base; no overlap
MOVED:  <worktree> — <n> commits on <ref> since the recorded merge base; overlaps: <paths>
REFUSE: <reason>
```

- Count: `git rev-list --count <recorded>..<effective-ref>`.
- Moved paths: `git diff --name-only <recorded>..<effective-ref>`.
- The change's paths: the union of `git diff --name-only <recorded>..HEAD`, `git diff --name-only
  --cached` and `git diff --name-only`.
- Overlap: the intersection, sorted, listed inline and capped with a `(+N more)` tail so one verdict
  line stays one line.
- `REFUSE` covers a `-` recorded merge base, a recorded merge base that does not resolve, and a base
  ref that does not resolve. Exit 2 covers a missing argument, a directory that is absent or not a
  git worktree, and a git invocation that fails for an environmental reason.

### `skills/flow/integrate.md`

Step 2 gains the check ahead of its prompt, per worktree in the resolved set, with the aggregated
ask on overlap and `-outcome stopped` on **Stop**. A `REFUSE` or an exit 2 from any worktree stops
and asks, exactly as the preflight's own does.

### `skills/flow-contracts/finish-contract.md`

Canonical subsection under Run 1 for the guard: the three verdicts, the exit contract, the
per-worktree rule, the aggregated ask, and the hand-run fallback for a harness whose repository does
not carry the script.

## Verification

- `scripts/test-check-finish-preflight.sh` — the ticket's own case: a branch merged into
  `origin/main` in a real clone, local `main` deliberately left behind, handed the bare name `main`,
  must return `RUN2`. Plus: `origin/main` handed explicitly still returns `RUN2`; a repository with
  no `origin` remote behaves exactly as before.
- `scripts/test-check-base-moved.sh` — `CLEAR` on an unmoved base; `MOVED` with no overlap;
  `MOVED` with overlap, including overlap contributed only by staged and only by unstaged work;
  `REFUSE` on `-`, on an unresolvable recorded merge base and on an unresolvable base ref; exit 2 on
  a missing argument and on a non-worktree directory.
- The repository's `## lint` list, in full, and `scripts/test-check-guard-symlinks.sh` for the new
  symlink.
