# kan-289 — brainstorming record

**Jira:** KAN-289 — *myflow: put "reproduce, don't read" into the review-panel slot template*

The approved design is canonical in
`spectre/changes/kan-289-reproduce-rather-than-read-in-slot-template/design.md`. This file records
what the brainstorming round itself settled, and is not a second copy of that design.

## What the round decided

Four questions were put to the operator and answered:

1. **How far the "reproduce, don't read" paragraph reaches** — all three dispatch-prompt families
   (panel slots, per-task reviewer, implementer), not the panel alone the ticket names.
2. **Whether it gets a guard** — yes, a presence check plus its harness, scoped to the new block.
3. **One text or two** — two variants under one shared label, reviewer-facing and
   implementer-facing.
4. **What the guard asserts** — the label plus a set of load-bearing phrases per variant, not label
   presence alone and not the full literal text.

The convergence confirm was answered **move on**, with no questions held.

## Scope added after the round closed

The operator then added a second, larger concern to the same change: rename `myflow` to `flow`
throughout the live corpus, identifiers included. Its boundary was settled by two further questions —
how far the rename reaches (everything, identifiers included) and what happens to the Jira label
taxonomy (renamed in the documentation, existing issues untouched). It is recorded on KAN-289 under
`## Added during implementation`.

## Blocker resolved during the round

`spectre new` was unavailable — the `spectre` CLI was not installed on this machine. The operator
asked for it to be installed; it was built from `~/Projects/spectre` at `3b16340` into
`~/.local/bin/spectre`, which was already on PATH, so no shell profile was changed.
