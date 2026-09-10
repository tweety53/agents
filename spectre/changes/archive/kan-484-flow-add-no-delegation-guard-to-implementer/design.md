# Design — kan-484-flow-add-no-delegation-guard-to-implementer

## Context

The conductor's four permitted dispatches (**Dispatch sites — the conductor's closed list**,
`skills/flow/implement.md`) each carry a fixed set of mandatory paragraphs, stated once per site as
a blockquote under a "Every … dispatch prompt also carries the X paragraph" lead-in, and each
paragraph has a row in `scripts/check-dispatch-paragraphs.sh`'s table (label + short load-bearing
phrases + per-site minimum block count), with `scripts/test-check-dispatch-paragraphs.sh` pinning
the guard against sandboxed fixture trees. The KAN-473 TOOLS addition is the most recent
whole-shape precedent: one blockquote per site, one table entry, harness cases for label-absent
and one-phrase-dropped.

The sites today, by file, with the paragraphs each already carries:

| Site | File | Lead-in shape |
|---|---|---|
| implementer dispatch | `skills/flow/implement.md` §4 | "Every implementer dispatch **must** carry:" then bare blockquotes (COMMIT-PER-TASK … REPORT FILE) |
| panel slot dispatch | `skills/flow/review-panel.md`, after **Bundled dispatch** | "**Every slot's dispatch prompt also carries the X paragraph**:" |
| panel-fix subagent dispatch | `skills/flow/review-panel.md`, the fix step | "**Every fix subagent's dispatch prompt also carries the X paragraph**:" |
| verifier dispatch | `skills/flow/verify-and-handoff.md`, **The verifier dispatch** | "**The prompt also carries the X paragraph**:" |

The planner dispatch (`skills/flow/brainstorm.md`) is a leaf with the Agent tool too, but is the
parent's child, not the conductor's; the operator scoped it out.

`spectre/specs/` is empty in this repository — the skill markdown is the contract surface — so
there is no spec task. Contract budgets (`scripts/check-contract-budget.sh`) have headroom of
roughly 7–11 KB per file against an addition of under 1 KB each, so no budget row moves.

## Sections

### 1. The paragraph

Verbatim, at every site, no variants:

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the conductor's closed list (**Dispatch sites — the
> conductor's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

Guard phrases (shared, all required of every block): `Never call the \`Agent\` tool` ·
`never spawn a subagent` · `the leaf of this run`.

### 2. Sites

- `skills/flow/implement.md` §4: the blockquote joins the implementer's list, directly after the
  TOOLS blockquote. One block in the file — the conductor's own §4 restatement of FOREGROUND
  BUILDS / REPRODUCE, DON'T READ is not a site for this paragraph, since the conductor is already
  bound by the closed list. In **Dispatch sites — the conductor's closed list**, one added sentence
  after the self-check paragraph: the four rows are the whole run's dispatch tree because every
  row's prompt carries the NO DELEGATION paragraph (section **4**, `skills/flow/review-panel.md`,
  `skills/flow/verify-and-handoff.md`) — a leaf never dispatches, so nothing below these rows
  exists.
- `skills/flow/review-panel.md`: "**Every slot's dispatch prompt also carries the NO DELEGATION
  paragraph**:" after the slot TOOLS paragraph, and "**Every fix subagent's dispatch prompt also
  carries the NO DELEGATION paragraph**:" after the fix TOOLS paragraph — two blocks. The
  **Experimental slot** sentence enumerating "REPORT FILE / REPRODUCER / CONTEXT BUNDLE /
  WORKTREES / TOOLS / FOREGROUND BUILDS / MODEL HANDSHAKE / REPRODUCE, DON'T READ" gains
  `NO DELEGATION`.
- `skills/flow/verify-and-handoff.md`, **The verifier dispatch**: "**The prompt also carries the
  NO DELEGATION paragraph**:" after the TOOLS paragraph — one block.

Additions only: no existing normative sentence is cut or reworded.

### 3. The guard row

`scripts/check-dispatch-paragraphs.sh`: entry key `delegation`, `ENTRY_LABEL` `**NO DELEGATION:**`,
`ENTRY_SHARED_PHRASES` the three phrases of section 1, `ENTRY_VARIANTS` empty; four rows appended
to the parallel site arrays — `skills/flow/implement.md` min 1, `skills/flow/review-panel.md`
min 2, `skills/flow/verify-and-handoff.md` min 1 — and the header comment's history paragraph and
table gain the KAN-484 entry in the shape of the KAN-473 one.

### 4. The harness

`scripts/test-check-dispatch-paragraphs.sh`: a `DELEGATION_BLOCK` literal (the section 1
blockquote); `new_root` seeds `verify-and-handoff.md` with it alongside TOOLS and MODEL HANDSHAKE;
`CLEAN_IMPLEMENT` gains one copy, `CLEAN_REVIEW_PANEL` two, and every earlier inline fixture that a
case relies on being otherwise clean gains the same — the precedent every prior paragraph followed,
so no existing case's assertion changes meaning. New cases, numbered after 45:

- 46 — label absent from `implement.md`: exit 1, output names `implement.md` and `NO DELEGATION`.
- 47, 48, 49 — one per shared phrase, dropped from one of `review-panel.md`'s two blocks while the
  other stays correct: exit 1, output names the missing phrase.
- 50 — label absent from `verify-and-handoff.md` (its `new_root` default overridden): exit 1.
- 51 — `review-panel.md` with exactly one correct block and the second omitted: exit 1, output
  carries the min-blocks message (`requires at least 2 block(s)`), pinning the threshold from the
  start rather than after a review finds the gap as cases 20–21 did for FOREGROUND BUILDS.

The harness must fail before the guard row lands (task 2 red) and pass after (task 3 green);
`scripts/check-dispatch-paragraphs.sh` against the real tree passes throughout because the prose
lands first (task 1).

## Toggles

Resolved `dynamic` for all three (`## execution mode`, `## implementer model`, `## review panel`);
the decision is `tasks.md`'s `## Decision` block and `.superpowers/sdd/decision.json`.

## Decisions

### The paragraph reaches the verifier as well as the ticket's three sites

**ID:** four-leaf-sites
**Status:** active
**Chosen:** implementer, panel slot, panel-fix and verifier — every leaf row of the conductor's
closed list is spawned the same way with the same tool set, and the ticket's failure shape (a
conductor's child forking an unrecorded grandchild) applies to each of them equally.
**Considered:** the ticket's three sites only — leaves the verifier as the one conductor child
with no such instruction; all five leaf sites including the planner — the planner is the parent's
dispatch, not the conductor's, and outside the failure shape the issue names, so the operator
excluded it.

### The paragraph gets a guard row, not prose alone

**ID:** guard-row
**Status:** active
**Chosen:** a `NO DELEGATION` entry in `scripts/check-dispatch-paragraphs.sh` plus harness cases —
the guard is what has kept every prior mandatory paragraph from being trimmed away by a later edit
(KAN-289's root cause), and a paragraph outside its table is the one a future edit could drop
silently.
**Considered:** prose only, exactly the ticket's acceptance list — cheaper, but leaves this
paragraph the sole unguarded one among its siblings.

### The wording is the operator-approved draft, verbatim at every site

**ID:** draft-verbatim
**Status:** active
**Chosen:** one blockquote, identical at all four sites, no variants — the paragraph's job is to
be propagated unchanged by a conductor, and identical text with three short guard phrases is what
makes that mechanically checkable.
**Considered:** per-site variants (an implementer-flavoured and a reviewer-flavoured block, as
REPRODUCE, DON'T READ has) — nothing in the obligation differs by role, so a variant would be a
second wording of the same rule.

## Open questions

None.
