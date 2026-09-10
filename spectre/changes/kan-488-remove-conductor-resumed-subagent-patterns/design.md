## Context

Seeded from `docs/superpowers/research/flow-remove-resumed-subagents.md` (35K, adopted manually —
predates the strict keyed-filename seed convention this change itself introduces, so the automatic
`kan-488*.md` glob does not match it; approved as-is by the operator). Full measurement detail
(kan-472/kan-473 transcript analysis, cache-TTL facts) lives in that note's history — deleted from
the tree as part of this change's artifact commit, its content now canonical here.

## Decisions

### Remove the conductor — the parent orchestrates directly

**ID:** remove-conductor
**Status:** active
**Chosen:** Delete `implement.md`'s **Dispatch the conductor** and **The parent backstop**; the
parent runs `implement.md` §1/2/4, `review-panel.md`, `verify-and-handoff.md` itself — every
guard, gather, `flow record`, `flow stage`, worktree add/remove, report read and diff walk. The
four one-shot dispatch rows (implementer, panel-bundle, panel-fix, verifier) are unchanged: still
dispatched fresh per task/round, never resumed.
**Considered:** tuning `subagentPromptCacheTtl`/`experimental.cacheTtl` instead of removing the
role — rejected (see `reject-global-cache-ttl` below).

### Planning inline — no planner, no researcher subagent

**ID:** planning-inline
**Status:** active
**Chosen:** `/flow`'s brainstorming (`brainstorm-planner.md` B/C/D) and `/flow-research` (any
argument shape) both run entirely in the current session, on the session's own model — no
dispatched planner/researcher subagent, no relay, no `Model:` handshake, no `opus` fallback.
`AskUserQuestion` calls happen directly, batched up to four independent questions per call exactly
as the relay already batches.
**Considered:** keeping the planner/researcher on a fixed `PLANNING_MODEL` for cost/quality
control — rejected: the session is already on the 1-hour TTL and already holds the operator's
conversation, so a relay through a second context bought only a model choice the operator no
longer wants.
**One shared mechanism, not two copies:** `brainstorm-planner.md`'s **Seed from a staged research
note** section stays canonical for the seed-lookup/checklist/convergence mechanics; `skills/flow-research/SKILL.md`
cites it rather than re-describing the same logic in its own words, so `/flow-research`'s inline
checklist and `/flow`'s inline brainstorm checklist cannot drift apart. `/flow-research` still
creates no worktree and no spectre artifacts — it only points at the same shared mechanics for the
seed check, the checklist invocation, and the convergence loop.

### Strict research-artifact path, keyed by Jira key

**ID:** strict-artifact-path
**Status:** active
**Chosen:** `/flow-research` writes to exactly one of three deterministic destinations —
`<project>/docs/superpowers/research/<jira-key-lowercased>.md` (key known, no change yet), the
existing change's `design.md` (change exists), or `<topic-slug>.md` (no key, bare session).
`/flow KAN-XXX` checks the keyed path with one `test -f` — never a glob, never inference. The
`<key>-*.md` wildcard fallback and its "more than one match" disambiguation prompt are removed
from `brainstorm-planner.md`'s **Seed from a staged research note**.
**Considered:** keeping the wildcard fallback for hand-written notes — rejected: the one file it
existed for, `docs/superpowers/research/kan-326-myflow-rework.md`, belongs to an archived change
and is deleted in this change instead (see `delete-kan-326-fixture` below).

### Startup visibility for resolved dynamic choices

**ID:** startup-visibility
**Status:** active
**Chosen:** The seed result (found vs. none) prints as the second line after "Using flow for
change `<name>`.", on every run. Seeded path: the full block — planning mode/model, the three
toggles, `DEFAULT_MODEL`/`REVIEWERS`, the recorded decision or "not yet decided" — follows the seed
line, before any stage past kickoff. No-seed path: those same lines join the existing
post-`writing-plans` `## Decision` print, neither moved nor duplicated.
**Considered:** printing the full block always at kickoff, even with no seed — rejected: before
writing-plans runs on a no-seed path, most of the block (the recorded decision) does not exist yet.

### Remove `planningModel` / `## planning model` entirely

**ID:** remove-planning-model-setting
**Status:** active
**Chosen:** Delete `skills/flow/SKILL.md`'s `PLANNING_MODEL` resolution block; `flow settings set
-planning-model`, the `PlanningModel` store field and its `ValidModels` check, and the `planning
model:` line in `/flow-settings`'s output; the `## planning model` row/paragraph in
`project-configuration.md`; the key's case in `check-model-keys.sh`; `/flow-research`'s "resolves
`PLANNING_MODEL`" sentence. The store's `planning_model` column is dropped by a migration or left
nullable and unread — implementer's call at the schema.
**Considered:** leaving the dead setting in place for later cleanup — rejected per this
repository's own no-dead-configurability stance; nothing reads `PLANNING_MODEL` once planning is
inline, so leaving it is deferred debt with no future caller.

### Remove the inline-parent context ceiling

**ID:** remove-context-ceiling
**Status:** active
**Chosen:** Delete the 250k/400k/six-bundle stop from `implement.md`'s **Inline — the parent
implements**, and the `## Context ceiling — clear and resume` handoff block from
`handoff-blocks.md`. A parent on the 1-hour TTL pays 0.1x per-call reads, not a rewrite, at any
context size, so the run's only size control becomes the read-discipline rules below and the
model's own window. Applies identically to the orchestrating parent (post-conductor-removal) and
the inline-implementing parent (small/regular class) — the operator confirmed no distinction:
"truly unlimited, same answer as the orchestrating parent."
**Considered:** raising the ceiling instead of removing it — rejected: any fixed number is still
an arbitrary stop for an agent that no longer pays the cost the ceiling was invented to guard
against.

### Read-discipline rules for the parent

**ID:** read-discipline
**Status:** active
**Chosen:** Five rules, stated once in `implement.md` and cited from `review-panel.md` and
`verify-and-handoff.md`: never `cat` a report (verdict section via `sed -n`); never read
`final-review.diff`, a dispatch-context bundle, or a whole fix diff in the parent; test/lint
output through `tail`; each phase file read in full once per run, then by section via
`grep -n`/`sed -n`; change artifacts read once at `flow.load-context`.
**Considered:** no explicit rules, relying on judgment — rejected: kan-472's conductor reached
920k tokens doing exactly the things these rules forbid (whole-file `cat`s, repeated phase-file
reads), and with the ceiling removed these rules become the run's only size control.

### Delete the `kan-326-myflow-rework.md` fixture

**ID:** delete-kan-326-fixture
**Status:** active
**Chosen:** Delete `docs/superpowers/research/kan-326-myflow-rework.md` in this change, in the
same task that removes the `<key>-*.md` glob from **Seed from a staged research note**. It is the
one file that glob existed to match, belongs to an already-archived change, and the operator chose
deletion over a rename.
**Considered:** renaming it to fit some other convention — rejected per the operator's explicit
choice ("Delete it").

### Housekeeping bundled into this change

**ID:** bundled-housekeeping
**Status:** active
**Chosen:** Reword `pipeline.md`'s **Progress visibility** paragraph away from the conductor's
`## Stage` relay (the mechanism — one task-list entry per stage, marked at each `flow stage end` —
is unchanged); lower `check-contract-budget.sh`'s byte-size rows for `implement.md`,
`review-panel.md`, `verify-and-handoff.md`, `SKILL.md`, `brainstorm.md` and
`skills/flow-research/SKILL.md` to the measured post-cut sizes (`wc -c`, same commit that makes the
cut, so the budget cannot silently re-grow).
**Considered:** leaving these as follow-up tickets — rejected per the operator's explicit choice
to bundle them into this change.

### Rejected: a global cache-TTL setting

**ID:** reject-global-cache-ttl
**Status:** active
**Chosen:** No TTL tuning. Structural removal of the resumed roles instead.
**Considered:** `subagentPromptCacheTtl: "1h"` / `CLAUDE_CODE_SUBAGENT_PROMPT_CACHE_TTL=1h` (whole
"everything else" bucket, every project, 2.0x write multiplier instead of 1.25x) and the
per-definition `experimental.cacheTtl` frontmatter field — both rejected: the global setting is a
pure cost increase for every one-shot subagent that never idles past 5 minutes (Explore searches,
panel bundles, verifiers);
<!-- measured: code.claude.com/docs/en/prompt-caching, read and confirmed by the operator during the research session -->
the per-definition field would narrow it, but the operator declined TTL tuning altogether in favor
of removing the resumed roles.

### Worktree creation moves to the end of planning

**ID:** worktree-created-at-end-of-planning
**Status:** active
**Chosen:** `/flow`'s worktree (`git worktree add`, `## worktree setup`) is created **after**
section **D** (writing-plans + the Decide step) completes, not between design approval and section
**C** as today. `spectre new` and the three artifacts (**C**) and the writing-plans enrichment
(**D**) run directly in the main checkout's `<project>/spectre/changes/<name>/` — uncommitted,
never staged or committed there, per the existing git-boundaries rule. `.superpowers/sdd/decision.json`
is written to `<project>/spectre/changes/<name>/.superpowers-sdd-decision.json` in the main
checkout during **D**, since no worktree exists yet to hold it at its usual path. Once **D**
finishes, `/flow` creates the worktree (**The three returns**' steps 1-4, unchanged), then moves
`<project>/spectre/changes/<name>/` into `<worktree>/spectre/changes/<name>/` and the interim
decision file into `<worktree>/.superpowers/sdd/decision.json`, leaving nothing behind in the main
checkout. Applies to `/flow`'s planning uniformly — seeded from a `/flow-plan` staging note or run
bare with no seed alike; `/flow-plan` itself creates no worktree either way and is unaffected.
**Considered:** worktree created between design approval and **C** (today's order, and this
session's own already-committed task 2 edit) — superseded by explicit operator instruction
mid-session: "worktree/worktrees should be created at the end of planning (for both flow-plan and
bare flow without research)." Also considered writing **C**/**D**'s output to a scratch/temp
location instead of the main checkout directly, moving it into the worktree afterward — rejected by
the operator in favor of writing straight into the main checkout's `spectre/changes/<name>/`, since
no copy step is needed for that content beyond the one move into the worktree at the end.

### Rename `flow-research` to `flow-plan`

**ID:** rename-flow-research-to-flow-plan
**Status:** active
**Chosen:** `skills/flow-research/` → `skills/flow-plan/`, `commands/flow-research.md` →
`commands/flow-plan.md`, `commands-claude/flow-research.md` → `commands-claude/flow-plan.md`, and
every live-tree reference to the skill/command updated to match. Archived changes and
`docs/superpowers/{research,specs,plans}/` notes keep their historical name — they document what
was true when written.
**Considered:** keeping the `flow-research` name — the operator's own instruction, mid-session:
"flow-plan fits better" now that the skill's only job (per `planning-inline`) is running research
and brainstorm-adjacent planning inline in the current session, with no dispatched
researcher/planner subagent left to distinguish "research" from "plan".

## Open questions

<!-- none — every item the research session surfaced was closed by asking; see design.md's
     Context note and the adopted staging note's own §9 for the record of those answers -->
