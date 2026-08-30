## Context

`skills/flow/review-panel.md` dispatches "**separate** review subagents — one per included slot"
concurrently, in every affected worktree. Bugbot is the one slot whose brief requires editing the
tree in place (mutation testing); every other slot only reads it. Concurrent dispatch into one
shared worktree makes a reading slot's evidence and a live mutation indistinguishable, in either
direction — the bias observed on KAN-173 was benign (a reverted mutation reads as a false
"surviving mutant" finding, not a false clean), but nothing in the contract guarantees that
direction, and the reverse (a reading slot's verdict taken against mutated code) is exactly as
possible.

This is one change because the fix touches one seam: how Bugbot's dispatch is pointed at a
filesystem path, everywhere `review-panel.md` dispatches it.

## Decisions

### bugbot-gets-a-throwaway-worktree

**ID:** bugbot-gets-a-throwaway-worktree
**Status:** active
**Chosen:** give Bugbot's dispatch its own throwaway git worktree, built fresh at every dispatch
site (pass 1, targeted re-run, full re-run), so its edits never reach the tree reading slots see —
**Considered:**
- Serialize mutating slots after every reading slot returns — simplest, but drops the concurrency
  the panel currently relies on for wall-clock cost; rejected in favor of preserving full
  concurrency.
- Warn reading slots that concurrent mutation may be present — cheapest, but only reduces
  misdiagnosis risk, it does not eliminate a reading slot's evidence being corrupted by a live
  mutation; rejected as treating the symptom, not the collision itself.

### isolation-scope-is-bugbot-only

**ID:** isolation-scope-is-bugbot-only
**Status:** active
**Chosen:** only Bugbot's dispatch is isolated. Security is dispatched the same way
(`subagent_type`, "unknown (agent-defined)" behaviour) but `review-panel.md` documents no
mutation-testing brief for it.
**Considered:** isolating both Bugbot and Security, on the grounds that neither's internal
behaviour is fully controlled by this contract — rejected: isolating a slot the contract does not
document as mutating would be solving a problem not established to exist, and adds a throwaway
worktree with no stated reason.

### copy-mechanism-is-worktree-add-plus-diff-transplant

**ID:** copy-mechanism-is-worktree-add-plus-diff-transplant
**Status:** active
**Chosen:** `git worktree add --detach <worktree>-bugbot-<round> HEAD`, then transplant the apply
worktree's actual uncommitted state onto it — `git diff --binary` piped through `git apply`, plus a
plain copy of any untracked files (`git status --porcelain` `??` entries) — so Bugbot sees exactly
the working tree the other slots see, but as its own independent git worktree with its own
index/HEAD. Bugbot's own `git checkout --`/`git add` calls during mutation testing therefore behave
normally.
**Considered:** a plain `cp -a` of the whole worktree directory, `.git` worktree-link file
included — rejected: the copy would share git administrative state
(`.git/worktrees/<name>`) with the original, so git commands run from the copy risk corrupting or
racing the original's index/HEAD.

### copy-is-created-and-removed-around-one-dispatch

**ID:** copy-is-created-and-removed-around-one-dispatch
**Status:** active
**Chosen:** the throwaway copy is created immediately before Bugbot's dispatch and removed
(`git worktree remove --force`) immediately after that dispatch closes — success, a timed-out
breach, or a stopped run — never left for `/flow`'s worktree-cleanup phase to find. Findings and
their `file:line` are repo-relative and unaffected by which worktree Bugbot wrote them from;
reproducer verification already runs against the real `<worktree>`, never against Bugbot's copy.
**Considered:** none — this follows directly from the copy being single-purpose scratch space for
one dispatch, not a resource any later phase needs to know about.

## Open questions

None.
