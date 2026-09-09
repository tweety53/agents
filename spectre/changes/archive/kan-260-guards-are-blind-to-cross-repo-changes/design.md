# kan-260-guards-are-blind-to-cross-repo-changes — design

## Context

KAN-260 predates KAN-363, whose satellite machinery (`link.md`, `spectre link`,
`scripts/lib/change-plan.sh`, canonical-worktree arguments) already covers in-run cross-repo
resolution; `/flow`'s workspace-isolation stage hard-fails when the link cannot be written. The
change therefore only closes the two residuals: guards run where no change dir exists at all, and
store calls that resolve a project key from the judged worktree. Both fixes stay inside
KAN-363's *guards-take-the-canonical-worktree-path* decision: guards receive caller-supplied
knowledge and never query the store for resolution.

## Decisions

### Reuse the canonical-worktree argument for absent-dir resolution

**ID:** canonical-arg-direct-resolution
**Status:** active
**Chosen:** resolve a same-named plan directly from the supplied canonical worktree when the
local change directory is absent entirely — no new argument, no store access.
**Considered:** the ticket's state-record worktree-list resolution — from the second repo
`flow state get -C <satellite-worktree>` resolves the second repo's project key, the same
wrong-project defect at a lower layer, and it would still need the canonical path passed; a new
explicit `<change-dir>` argument — redundant with the canonical-worktree argument once that
argument works without a `link.md`, and a second way to pass one fact invites caller drift.

### Anchor the store calls at the plan

**ID:** plan-anchored-store-calls
**Status:** active
**Chosen:** `check-unfinished-work.sh` passes `-C` anchored at the directory the plan resolved
from for the findings query and both verdict calls; on the no-plan fall-through it keeps
`-C "$WORKTREE"`. Records live under the project of the repo that ran `/flow` — where the plan
lives — so the anchor is right by construction and the CLI keeps deriving the key.
**Considered:** computing project keys in bash — re-implements `ProjectKey`'s sha1 plus symlink
resolution and drifts from the CLI's own derivation; skipping the satellite-side verdict write —
loses a verdict that is meaningful and joins correctly once anchored.

### The new branch never fires when a local change dir exists

**ID:** no-fallback-for-linked-satellites
**Status:** active
**Chosen:** gate the direct-by-name lookup on the local change **directory** being absent
entirely; a satellite dir carrying `link.md` keeps link resolution and its loud failures
("a supplied canonical worktree is never retried"), unchanged.
**Considered:** trying canonical-by-name before link resolution — could mis-resolve when the
canonical worktree happens to hold an unrelated same-named change, silently, where today it
fails loudly.

## Open questions
