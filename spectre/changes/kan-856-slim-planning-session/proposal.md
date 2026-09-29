# Proposal — kan-856-slim-planning-session

**Jira:** KAN-856 (epic KAN-851). **Audit:** `docs/prompt-audit-2026-09-29/audit-planning.md`,
`## Candidates`. **Blocked by:** KAN-852 (guard), KAN-853 (reconcile) — both done.

## Why

- Reachability, checked on `966d2573`: `skills/flow/SKILL.md` routed `STARTED` to **Resuming at
  `STARTED`** in `brainstorm.md`. So every post-`/clear` implementation session loaded all 17 KB of
  `brainstorm.md` to read 1.6 KB of it.
- `brainstorm.md` also cited `handoff-blocks.md` (16 KB) for the 0.6 KB `STARTED` block. That
  contradicts `handoff-blocks.md`'s own "Loaded by `/flow-status` and no other command".
- The rest of the planning load set holds rationale, duplicates and one-condition blocks.

## What changes

This change makes verbatim moves and declared-duplicate cuts only. The only rewording is in the
lines listed in `verbatim-moves.txt`.

1. **BR-08:** **Resuming at `STARTED`** and the resume half of "Resume and fix runs" moved to
   `skills/flow/resume.md`. The router and every citer now point there. It loads `brainstorm.md`
   only when planning is unfinished.
2. **HB-01:** `brainstorm.md` carries the `STARTED` block itself and no longer cites
   `handoff-blocks.md`.
3. **BR-09 + BP-02:** the withdrawal route and the `flow-fix`/`flow-cost` reachability check moved
   to `skills/flow/withdrawal.md`. It loads on those labels or on a planless resume.
4. **Lazy blocks:**
   - BP-04: the seeded note moved to `skills/flow/seeded-note.md`.
   - JI-03, JI-06, JI-11 and JI-12, plus the follow-up pointer, moved to
     `jira-integration-finish.md`, loaded by run 1 and by the commands that create issues.
   - PV-04 moved to `plan-amendment.md`, which the implementer dispatch cites.
5. **Duplicates cut:** BR-01, BR-02, BR-04, BR-10, BR-11, BR-13, BR-14, BR-16, BP-01, BP-10,
   BP-15, BP-16, BP-17, BP-25, JI-01, JI-02 and BG-04. Also cut: the fix-run half of "Resume and
   fix runs" (BR-17), which is canonical in `implement.md` §3.
6. **Rationale** moved to `-rationale.md` files: BR-03, BR-05–07, BR-12, BP-11–14, BP-18, JI-04,
   JI-05, JI-07–10, PV-01–03, BG-02, BG-03 and BG-05.
7. **Guard and support updates:**
   - `brainstorm-planner.md` joins `smcCandidates`.
   - `build-green.md` is declared expected-zero in `check-references`.
   - `stage-keys.md` and `load-sets.sh` follow the moves.

## Out of scope

- The Decide mechanics (BP-21–23, BR-M1) belong to the mechanics issue.
- BP-20 and BP-08 are deferred (design.md D2, D3).
- BR-15, BP-06 and BP-19 were already fixed by KAN-853/854.
