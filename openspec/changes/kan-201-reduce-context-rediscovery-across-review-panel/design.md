# Design — reduce context rediscovery across dispatches

**Change:** `kan-201-reduce-context-rediscovery-across-review-panel`
**Jira:** KAN-201
**Date:** 2026-08-18

## Problem

KAN-73's self-review measured ~1.98M subagent tokens across 12 dispatches and named two
instances of one shape: a dispatch boundary discards work already done on the other side of it.

**C1 — every panel slot gathers the same context independently.** The `light` roster's three
required slots each separately locate and read `proposal.md`, `design.md`, `tasks.md`, the delta
spec, `engineering-principles.md`, and `final-review.diff`. Nothing is shared: each dispatch starts
cold and re-derives the identical reading set. The waste scales with roster size.

**C2 — the fix round is the largest single dispatch and starts cold.** At 321,403 subagent tokens
it re-read the panel record, the design, the plan and the guard source that the three review slots
had just analysed. Every finding already carried a file, a line and a runnable reproducer, so the
located evidence existed and was simply not carried across the boundary.

The same shape holds for the implementer dispatches this change also covers: task 5's implementer
cost 275,064 tokens and bundle 3's cost 257,357, each re-deriving the same plan and principles.

## What the saving actually is

The bundle does not reduce the bytes of planning context a dispatch reads. It removes the
**rediscovery**: globbing for delta specs, resolving the principles path, opening `tasks.md` to
learn which tasks exist, reading a standards file to find it irrelevant, and backtracking after a
wrong guess. That is what "starts cold and re-derives the identical reading set" names, and it is
what this change removes. Stating this plainly matters, because a design that promised a
content-size saving would be measured against a number it cannot move.

## Design

Two halves, deliberately independent — either lands without the other.

### Half A — the gathered bundle

`scripts/gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path>` prints one
bundle to stdout; the caller redirects it to `<worktree>/.superpowers/sdd/dispatch-context.md`.

It follows `scripts/gather-self-review-context.sh`'s contract exactly, including its security
posture: each passed path is lexically normalized, semantically resolved, and the two compared for
exact equality, refusing on any divergence — the single general mechanism that script's header
documents in place of four bounded, individually-bypassable patches. The change-name allowlist is
the same one. **Exit 2 on a malformed invocation; otherwise always exit 0**, a missing source
reported as `skipped: <src> (absent)` rather than treated as a failure, since a change may
legitimately have no `design.md`.

Sources, in order: `proposal.md`, `design.md`, `tasks.md`, every delta spec under
`<change-root>/specs/`, and the content of `engineering-principles.md`.

**The principles path is passed in, not derived.** The rule that it must be the absolute path of
the file beside the running `SKILL.md` lives in `skills/myflow-do/SKILL.md`; a script that
re-derived it would be a second copy of that rule, free to drift. The script validates the path it
is given and no more.

**`[STANDARDS_PATHS]` stays out of the bundle.** Those entries resolve through the entry-form table
and containment rule in `skills/myflow-contracts/project-configuration.md`, they belong to the
principles slot alone rather than to every dispatch, and re-implementing that containment logic in
Bash is exactly the duplication the contract exists to prevent.

The bundle header names the generating timestamp and `HEAD`'s sha, so a reader can always tell
which state of the tree it describes.

#### Freshness

Rebuilt at the start of section 4 (before implementer dispatches), at the start of section 5
(before panel slots), and before each fix round. Same path, overwritten.

Gathering once per run was rejected: section 3 documents a fix by editing `proposal.md` and
`tasks.md`, so a run-scoped bundle would leave every later dispatch reading a plan that no longer
exists — a correctness hazard, not merely a stale-cost one. Rebuilding on every dispatch was also
rejected: it costs nothing but makes the bundle non-deterministic across slots that are supposed to
be seeing identical inputs, which is the property the bundle exists to establish.

#### Authority — advisory, never authoritative

Every dispatch prompt gains one paragraph:

> **CONTEXT BUNDLE:** `<abs>/.superpowers/sdd/dispatch-context.md` carries this change's proposal,
> design, plan, delta specs and engineering principles, gathered for you — you need not go looking
> for them. You may open any file it names. You **must** still read the actual diff and the actual
> code you are reviewing or changing: the bundle is shared *input*, never a substitute for the
> source, and never a shared conclusion.

This is KAN-201's own constraint — sharing the inputs is fine, sharing conclusions is not — stated
where a dispatch will actually read it. An authoritative bundle was rejected because a slot that
spots a gap in it would have no recourse, and a truncated bundle would become silently
authoritative.

### Half B — per-dispatch cost attribution

The data already exists on disk and is already harvested.
`~/.claude/projects/<project>/<session>/subagents/agent-<id>.jsonl` carries `agentId` and
`isSidechain: true` on every line, under the **parent's** `sessionId`; the sibling
`agent-<id>.meta.json` carries `agentType`, `description`, `toolUseId`, `spawnDepth` and `model`.
`discoverTranscripts` already walks these files. Their tokens land in `tokens.sidechain` today,
merged across every dispatch the stage made — which is why `do.review-panel` cannot distinguish a
slot from a fix round, and `do.sdd-tdd` cannot distinguish an implementer from a per-task reviewer.

Four steps, each following a pattern that already exists twice in this codebase:

1. `harvest.Record` gains `AgentID`, decoded from the line's existing `agentId` field.
2. For a transcript matching `subagents/agent-<id>.jsonl`, the sibling `.meta.json` is read once per
   file for its descriptors. An absent or unreadable sidecar attributes the **tokens** normally and
   records no descriptors.
3. `harvest.Delta` gains `Dispatches`, built exactly as `Models` is, feeding the metrics bag's
   `dispatches.<agentId>` key.
4. Per-dispatch cost reuses `internal/store/pricing.go` against that dispatch's own recorded model.
   No new pricing path.

**Nothing existing is rekeyed.** `tokens.main`, `tokens.sidechain` and `models.*` keep their current
meaning and every current aggregation reads them unchanged. `dispatches.*` is purely additive, which
is what keeps historical stage runs readable rather than retroactively wrong.

**Absence is not a value, at the new granularity too.** A record with no `agentId` — the main
session — contributes to `Total` and to `Models`, and to no dispatch entry, mirroring the existing
rule that a model-less record creates no fabricated `unknown` key.

### The surface

`RunDetail`'s stage-run table gains an expandable row: a stage run opens onto its own dispatches —
description, agent type, model, tokens, cost — sorted by cost descending. The existing `Unavailable`
component renders a dispatch the store never measured, which keeps `myflow-stats-views`' "a recorded
but unmeasured run is distinguishable from an absent one" requirement satisfied at this granularity.

No ninth view, so that spec's "The eight views are served" requirement is untouched. `RunDetail`
builds its own panel chrome for reasons its own header comment documents; the sub-table follows that
local pattern rather than forcing a shared generic across two differently-shaped unions.

## Failure modes

| Failure | Behaviour |
|---|---|
| Gather script missing (copied rather than linked skill dir) | Guard presence check reports it; dispatches proceed with today's prompt shape. The bundle is an optimization and never stops a run. |
| A bundle source absent | Reported `skipped: <src> (absent)`; dispatches read what is there. |
| A passed path resolving through a symlink | Refused — exit 2, the same disposition as any malformed invocation. |
| Stale bundle | Prevented structurally by the rebuild points; detectable after the fact from the header's sha. |
| Meta sidecar absent or unreadable | Tokens attributed, descriptors omitted. |
| Dispatch with no `agentId` | No dispatch entry; totals unaffected. |

## Decisions

### Scope reaches implementer dispatches, not only panel and fix

**ID:** scope-includes-implementers
**Status:** active
**Chosen:** panel slots, the fix subagent and implementer dispatches — the planning context all
three re-derive is the same set, so one bundle serves all three and excluding implementers would
leave the two largest non-fix dispatches (275K and 257K tokens) untouched.
**Considered:** panel + fix only, as KAN-201 scopes it — smaller blast radius, but the mechanism
would have to be re-opened immediately for the dispatches that cost nearly as much.

### One bundle, not role-scoped bundles

**ID:** one-bundle-not-three
**Status:** active
**Chosen:** a single `dispatch-context.md` every role reads — the planning context is identical
across roles, and each role's own material (the findings dossier, the task bundle, the diff) is
already carried on its dispatch prompt and stays there.
**Considered:** three role-scoped bundles — more precise, but three gather code paths and three
freshness rules for a difference that is carried on the prompt anyway.

### Per-dispatch attribution as a third metrics breakout

**ID:** dispatches-in-metrics-bag
**Status:** active
**Chosen:** `metrics.dispatches.<agentId>`, built the way `Models` already is — no schema
migration, no new query plumbing, and the third instance of a pattern that exists twice.
**Considered:** a `stage_run_dispatches` table — cleaner relational querying and sorting, at the
cost of a migration plus store, API and allowlist plumbing the metrics bag already provides.

### The bundle is advisory

**ID:** bundle-is-advisory
**Status:** active
**Chosen:** dispatches may open any file and must still read the real diff and code; the bundle
only means they need not go looking.
**Considered:** an authoritative bundle forbidding re-reads — a larger saving, but a slot that
spots a gap has no recourse and a truncated bundle becomes silently authoritative.

### Rebuild per dispatching stage

**ID:** rebuild-per-stage
**Status:** active
**Chosen:** rebuilt at section 4's start, section 5's start, and before each fix round.
**Considered:** once per run — cheapest, but a fix documented in section 3 edits `proposal.md` and
`tasks.md`, leaving later dispatches reading a plan that no longer exists. Per dispatch — maximum
freshness, but destroys the "every slot saw the same inputs" property.

### The fix dossier carries locations, not source excerpts

**ID:** dossier-without-excerpts
**Status:** active
**Chosen:** each finding's `file:line`, theme, severity, reproducer and bounce history, plus the
bundle; the fixer still opens the real code.
**Considered:** inlining source excerpts around each `file:line` — the fixer would open nothing,
but would then be editing against a snapshot that the fix round itself invalidates.

### Per-dispatch cost surfaces inside RunDetail

**ID:** surface-in-rundetail
**Status:** active
**Chosen:** an expandable per-dispatch row under an existing stage run — reuses an existing route
and its controls, and the data is exactly one level below what that table already shows.
**Considered:** API-only — smallest, but a harness nobody reads does not get read. A ninth view —
most visible, but `myflow-stats-views` states that eight views are served, and dispatch cost belongs
under a change anyway.

## Open questions

None. Every question raised during brainstorming was answered.

## Out of scope

Roster defaults, panel size, the finding-severity bar, and yield per slot (KAN-198). Nothing here
changes which slots run or what blocks a handoff.
