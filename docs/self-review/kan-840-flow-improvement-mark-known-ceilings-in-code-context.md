# Self-review context bundle for kan-840-flow-improvement-mark-known-ceilings-in-code

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-840-flow-improvement-mark-known-ceilings-in-code, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-840-flow-improvement-mark-known-ceilings-in-code.md (absent)
skipped: .superpowers/sdd/reviews/kan-840-flow-improvement-mark-known-ceilings-in-code-panel.md (absent)
skipped: spectre/changes/archive/kan-840-flow-improvement-mark-known-ceilings-in-code/tasks.md (absent)
skipped: spectre/changes/archive/kan-840-flow-improvement-mark-known-ceilings-in-code/design.md (absent)
skipped: spectre/changes/archive/kan-840-flow-improvement-mark-known-ceilings-in-code/narrative.md (absent)

## git log --stat

commit e6452b8500d38a65d8bef3d1a7ae0617b06ef334
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Oct 5 00:48:59 2026 +0300

    docs(briefs): mark known ceilings with a searchable ponytail tag at the site
    
    Promote the ponytail practice to the lessons home: a known, accepted
    limitation is marked in code at the exact site that has it, with a
    ponytail: comment naming what the code does not cover, so the next
    reader and the next grep find the ceiling before re-deriving it as a
    bug. The ticket's named instance (the gymie build-input allowlist's
    deletion ceiling) is already marked at its site; the practice itself
    had no durable statement anywhere.
    
    Filed as KAN-840, from kan-692's deferred self-review pass.

 .../mark-known-ceilings-with-a-ponytail-tag.md     | 32 ++++++++++++++++++++++
 1 file changed, 32 insertions(+)

## Session narrative

This run promoted KAN-840 — mark known ceilings with a searchable `ponytail:`
tag at the site — to the lessons home as
`docs/briefs/mark-known-ceilings-with-a-ponytail-tag.md`, the practice's first
durable statement; the `flow lesson resolve` the brainstorm ran found zero
briefs. The run's one struggle was locating the ticket's named instance: this
repository carries no wasm code, and kan-692 (the change whose deferred
self-review filed this ticket) is not in this project's store — the instance
turned out to live in the gymie project, whose buildSrc `DevStackFreshness.kt`
has carried the exact `ponytail:` mark since 2026-09-27, the day the ticket was
filed. That settled the deliverable as the practice brief rather than a code
mark, a judgment this run made and names in its summary. Implementation was one
commit on an inline micro decision; verification ran the full `## lint` list —
the first `go vet` and `tsc` passes failed on the fresh worktree's missing SPA
dist and node_modules, supplied by the project's own `make build` — plus the
fast-route record check; no `## test` command had scope in a docs-only diff.
