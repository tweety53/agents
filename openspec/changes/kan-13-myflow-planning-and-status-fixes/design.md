## Context

Two independent defects surfaced during the KAN-9 frontend run (see KAN-13), both in myflow
itself, both scoped to this `agents` repo (the other two KAN-13 items belong to `gymie`):

1. `/myflow-start` produced a plan whose task order left `:shared` uncompilable from Task 2 until
   Task 9 — tests couldn't run for seven consecutive tasks, two tasks had to be merged mid-flight
   to reach a green state, and the plan needed a human-adjudicated repair mid-run.
2. `/myflow-status` (and, via the same shared code path, every other `/myflow-*` command's
   change-name resolution) enumerates open changes from `openspec list --json` alone. That command
   only sees change directories present in the checkout it runs in, so a change whose
   `openspec/changes/` directory lives only in a worktree — staged but never committed to the main
   checkout — is silently invisible, even while it sits at the human gate with a fully staged diff.

The full brainstorm and the two alternatives considered for each fix are recorded in
`docs/superpowers/specs/2026-08-05-kan-13-myflow-planning-and-status-fixes-design.md`; this file
adapts that approved design into the OpenSpec shape.

## Goals / Non-Goals

**Goals:**
- A plan `/myflow-start` produces must declare, per task, whether the build is left green, and a
  guard must catch a plan that leaves a declared-red task unresolved before it publishes.
- Every `/myflow-*` command's change-name resolution must see a change whose planning artifacts
  exist only in a worktree, not only ones committed to the main checkout.

**Non-Goals:**
- Verifying that a plan's `**Build:** green` claim is actually true — no script can confirm an
  arbitrary project's code compiles for an arbitrary language, and this design does not attempt it
  (mirrors the existing `plan-provenance` guard, which verifies a claim is *stated*, never that it
  is true).
- The two `gymie`/`gymie-frontend` items on KAN-13 (untracking `.superpowers/`, the deferred
  review-panel minors) — separate changes, separate repos.
- Any change to `/myflow-do` or `/myflow-finish`'s own logic beyond the shared resolution section
  they both call.

## Decisions

### Bundle both agents-repo items into one change

**Chosen:** one combined change — both are myflow pipeline fixes discovered together during the
same KAN-9 run, and bundling avoids two proposals sharing one Jira issue with near-identical
context.
**Considered:** two separate changes, mirroring how KAN-9 itself split by repo — rejected because
these two items share a repo and neither depends on the other, so splitting by fix rather than by
repo would add review overhead for no isolation benefit. Also considered scoping this run to the
planning rule alone and deferring the status fix — rejected because the status fix is small and
already fully specified by the Jira issue.

### Build-green check is a structural tag, not a semantic build check

**Chosen:** a required per-task `**Build:**` tag with a merge-partner rule
(`green` / `red — merges with Task <N>`), checked structurally by a new script that parses
`tasks.md`. Mirrors the existing `verified:`/`unverified:` provenance-tag pattern in
`plan-provenance.md`, generalizes to any reason a task is red (not only file deletions), and costs
one line per task.
**Considered:** a file-touch-graph heuristic (`Creates:`/`Modifies:`/`Deletes:` per task, flagging
a verification command whose target is deleted by a later task) — catches the literal KAN-9 shape
automatically, but only that shape, and requires a new required field on every task just for this
one check. Also considered a fully automated semantic/AST-based compileability check — rejected as
infeasible in general (myflow plans span arbitrary languages and projects) and as over-engineering
for what the postmortem actually needed.

### Fix the shared resolution section, not just `/myflow-status`

**Chosen:** fix `pipeline.md`'s one shared `## Change name resolution` section. KAN-13 itself
flags that `/myflow-finish`'s resolution has the identical blind spot, and it uses the same code
path, so fixing the shared section fixes every command at once and prevents the two from drifting.
**Considered:** fixing only `/myflow-status/SKILL.md`'s step 1 — narrower and faster, but leaves
`/myflow-do` and `/myflow-finish` silently exposed to the exact bug this change exists to close.

## Risks / Trade-offs

- **[Risk]** A plan author tags a task `**Build:** green` dishonestly (the tag is unverifiable by
  the guard) → **Mitigation:** same accepted limit as `plan-provenance`'s guard; the obligation to
  tag honestly falls on whoever writes the plan, and the guard's job is only to ensure every task
  is tagged and no red task is left unresolved.
- **[Risk]** The state-directory union surfaces a change whose worktree/branch no longer exists
  (stale state file) → **Mitigation:** unchanged from today — self-heal already handles a state
  file contradicted by artifacts; this change only affects which names reach that resolution step,
  not what happens once a name is resolved.
- **[Risk]** A corrupt/unreadable file in the state directory silently shrinks the enumeration →
  **Mitigation:** explicitly out of scope for silence — an unreadable file is reported and skipped
  from the union, never silently dropped.

## Migration Plan

No data migration. Existing `tasks.md` files for already-`IN_PROGRESS` changes are not
retroactively tagged — the guard only runs during `/myflow-start`'s writing-plans stage, i.e. on
plans not yet published. No rollback beyond reverting the change; nothing is destructive.

## Open Questions

None.
