# Proposal — kan-857-slim-implement-and-review-panel

**Jira:** KAN-857 (epic KAN-851). **Audits:** `docs/prompt-audit-2026-09-29/audit-implement.md` and
`audit-review-panel.md`, `## Candidates`. **Blocked by:** KAN-852 (guard), KAN-853 (reconcile) —
both done.

## Why

- `review-panel.md` (102.9 KB) and `implement.md` (84.6 KB) are the two biggest files any session
  loads. Both load in every implementation session.
- Reachability, checked on `b8cc6e75`: section 3 of `implement.md` opens with "**Fix runs only**",
  and the Waves paragraph with "on every `sdd` decision". A first `inline` run still loaded both.
  It also loaded the whole fix-round machinery of `review-panel.md` before any round had recorded
  a finding.

## What changes

This change makes verbatim moves and declared-duplicate cuts only. The only rewording is in the
lines listed in `verbatim-moves.txt`.

1. **`implement.md` lazy splits:**
   - **L1:** the body of `## 3. Documenting a fix` moved to `skills/flow/document-fix.md`, loaded
     on a fix run.
   - **L2:** the per-group implementer, the Waves, the per-group gather and its failure handling,
     and the sixth-argument scoping moved to `skills/flow/sdd-dispatch.md`, loaded when
     `execution` is `sdd`.
   - **L4:** the created worktree's `## worktree setup`, `spectre link`, the rest of `## apps`
     resolution, the regression-checkout toolchain and the merge-order record moved to
     `skills/flow/cross-repo-worktrees.md`.
   - **L3:** the gated reviewer's `fix` path moved to `skills/flow/gated-review-fix.md`.
   - **L5:** the single-model-harness paragraph moved into **Harness mapping** in
     `model-policy.md`.
2. **`review-panel.md` lazy splits:**
   - **RP01:** the fix-round block moved to `skills/flow/review-panel-fix-round.md`. It covers the
     round-boundary base check, the commit routes, the re-run rules, the reproducer guards and
     re-runs, the mutation proof, the round close, the fix chunks and the non-convergence loop.
     It loads once a round records a Critical or Important finding. The pinned fix-subagent
     paragraphs and the panel-fix `flow record dispatch` pair stay in `review-panel.md`.
   - **RP03:** the late-fix reduction and its staleness carve-out moved to
     `skills/flow/review-panel-late-fix.md`, loaded on a fix run.
   - **RP04 and RP05:** MUTATION ENTRY CONTEXT and the mutating-role sentence moved into
     `review-panel-optional-slots.md`.
   - **OS04:** **Experimental slot** moved to `skills/flow/review-panel-experimental-slot.md`.
3. **Rationale moved verbatim to `skills/flow/SKILL-rationale.md`:**
   - from `implement.md`: R1–R16, R18, R20–R23, R29–R32 and D17 (see design.md D5 for the rows
     not done);
   - from `review-panel.md`: RP07, RP20–RP30 and RP33;
   - from the optional-slots file: OS01–OS03.
4. **Duplicates cut:** R34, D39, RP09, RP10 and RP12.
5. **Guard and support updates:**
   - `document-fix.md` joins `smcCandidates`.
   - Citers are repointed to the new files.
   - `stage-keys.md`, Read discipline's phase-file list and `load-sets.sh` follow the moves.

## Out of scope

- The pinned fix-subagent paragraphs (RP02) and the mutation-testing brief (RP06). `dpSites` and
  the pin guard hold both in place.
- The MECHANICS rows (MX1–MX7, RP34–RP37, OS05). They need code and parity tests.
- D40, RP18 and RP19. KAN-853 already resolved them.
