# kan-326-myflow-rework-redesign — design

## How

Full stage-by-stage audit and rationale for each cut/reorder is recorded in
`docs/superpowers/research/kan-326-myflow-rework.md` (pre-change research staging notes) — this
file adapts that into decisions/open-questions rather than duplicating the audit table.

Command-surface rename: `skills/myflow-start`, `skills/myflow-do`, `skills/myflow-finish` are
retired as user-facing commands; their content is redistributed into `/flow`'s skill file(s),
organized by topic (brainstorm / implement / integrate, or similar — exact split is
implementation-time). `skills/myflow-fast` becomes `skills/flow`. `skills/myflow-status` →
`skills/flow-status`, `skills/myflow-research` → `skills/flow-research` (rename only). New:
`skills/flow-settings`, `skills/flow-fast`.

Settings store: extends the existing stats Go daemon (`stats/cmd/myflow`, `stats/internal/api`,
`stats/internal/store`) with a defaults table (implementer/fixer/reviewer model, reviewers list)
and validation against an enumerated list, plus a small API surface `/flow-settings` calls to
read/write it. Per-change state file (`state-file.md`) drops the three `models.*` role fields and
`reviewPanelRoster`.

Stats UI/backend cut: `stats/web/src/views/{CostPerChange,ModelComparison,PanelEconomics,
ReworkRate}.tsx` removed along with their route wiring; backend support in `stats/internal/api/stats.go`,
`stats/internal/store/{aggregate,pricing,stageruns}.go`, `stats/internal/sweep/sweep.go` needs
per-symbol untangling since it's intermixed with code the surviving 4 views depend on — not a
directory-level delete.

### Decisions

### Content reorganized by topic, not by old command boundary

**ID:** flow-rename-content-split
**Status:** active
**Chosen:** when `/myflow-start`/`-do`/`-finish` are deleted, their cited content is re-derived
into a new file structure organized by topic — never preserved along the start/do/finish split.
**Considered:** inlining everything into one `flow.md` (rejected — unwieldy at ~2700+ lines);
keeping the three files as internal-only, non-command modules `/flow` still cites (rejected —
keeps the boundary the ticket asked to remove).

### `/myflow-status` and `/myflow-research` are pure renames

**ID:** flow-status-research-rename
**Status:** active
**Chosen:** `/flow-status` and `/flow-research` are the existing commands renamed, no functional
change. `/flow-research` in particular already replaced `openspec-explore` in the 2026-08-25
spectre cutover — the ticket's "uses openspec explore" line is stale, predating that cutover.
**Considered:** building `/flow-research` as something new (rejected — would duplicate existing,
working functionality).

### `start.ask-options` removed

**ID:** ask-options-removed
**Status:** active
**Chosen:** no planning-effort/model/review-panel-roster question round on a creating run.
**Considered:** keeping planning-effort alone while dropping only the model/roster questions
(rejected — not raised as a concern; operator asked for removal of the whole stage in the context
of the model-default and fixed-panel decisions below, which remove its other two purposes anyway).

### Single default model (sonnet) for implement, fix, and review

**ID:** model-default-sonnet
**Status:** active
**Chosen:** one default, sonnet, for all three roles (implementer, fixer, reviewer), operator can
override at any time via plain-language session instruction.
**Considered:** keeping the three separate roles with separate defaults, as today (Opus for
implementer/panel-fix, Sonnet for review) — rejected, since the operator wants uniform defaults and
override reach, and separate roles add config surface for no longer-justified benefit.

### `start.publish-proposal` removed

**ID:** publish-proposal-removed
**Status:** active
**Chosen:** `/flow` never publishes a proposal artifact, matching what `/myflow-fast`'s
creating-run branch already does today (`artifactUrl: null`).
**Considered:** keeping it, publishing to a URL every creating run (rejected — the operator answers
brainstorming's questions in the same session that produces the design, so a published summary of
that conversation adds nothing, exactly as myflow-fast's rationale already states).

### `STARTED` redefined as a kickoff marker

**ID:** started-redefined
**Status:** active
**Chosen:** `STARTED` is written the moment the operator invokes the command, before brainstorming
begins — "the task is started by the operator," the first stage. This replaces today's meaning
("design approved and proposal published").
**Considered:** dropping `STARTED` entirely, since removing `ask-options` and `publish-proposal`
left nothing distinguishing it from "no state yet" (rejected explicitly by the operator — `STARTED`
stays, redefined rather than removed).

### Review panel fixed at 3 required slots; Bugbot/Security fully on-demand

**ID:** review-panel-fixed-3
**Status:** active
**Chosen:** every `/flow` run dispatches exactly 3 required reviewer slots by default (no roster
presets). Bugbot and Security are dispatched only when the operator explicitly asks, at any point
during the run — never by an automatic trigger (diff size, touched area) or roster preset.
**Considered:** keeping the `light`/`standard`/`full` presets with the current trigger heuristics
(rejected — the operator wants the on/off decision to be entirely explicit, not heuristic-driven).

### `do.run-instructions` moves to immediately before `do.write-in-progress`

**ID:** run-instructions-reorder
**Status:** active
**Chosen:** stage order becomes `... → do.stage-diff → do.run-instructions → do.write-in-progress`
(was `do.run-instructions → do.workspace-export → do.lint-and-test → do.stage-diff → do.write-in-progress`).
**Considered:** leaving the order as-is (rejected — run instructions serve the human handoff, so
producing them last, right before the state write that ends the run, fits better than computing
them mid-stream).

### `finish.move-in-review` folds into `finish.landing-routes`

**ID:** move-in-review-fold
**Status:** active
**Chosen:** the Jira "move to In Review" step becomes a sub-step of `finish.landing-routes` rather
than its own top-level mark — it fires unconditionally on every landing route with no distinct
condition of its own.
**Considered:** keeping it separate as today (rejected — no information is lost by folding it in,
since it's still logged as a sub-step; one less top-level stage).

### `do.workspace-export` and `do.lint-and-test` merge into one verify stage

**ID:** workspace-export-lint-merge
**Status:** active
**Chosen:** merge into a single "verify" stage — both are unconditional automated pass/block checks
with nothing interactive between them.
**Considered:** keeping them separate (rejected on the same reasoning as the move-in-review fold).

### Model policy fields collapse to `models.default`

**ID:** models-fields-collapse
**Status:** active
**Chosen:** `models.implementation` / `models.reviewPanel` / `models.panelFix` become one
`models.default` field, since all three default to the same value now and override the same way.
**Considered:** keeping the three fields with the same recorded value (rejected — no longer earns
its complexity once the three roles share a default and an override path).

### `reviewPanelRoster` removed from the state file

**ID:** review-panel-roster-removed
**Status:** active
**Chosen:** the field is removed entirely; nothing replaces it in the persisted state except
possibly a running list of operator-requested Bugbot/Security additions for that run.
**Considered:** keeping the field but repurposing it (rejected — roster presets are gone per
`review-panel-fixed-3`, so there's nothing left for the field to record as a default).

### Model policy and reviewer list live in a stats-app settings store, not per-change state

**ID:** settings-store-stats-app
**Status:** active
**Chosen:** the stats app becomes both validator (rejects values off a strict enumerated list) and
source of truth for global model/reviewer defaults — not a value copied into each change's state
file at creation.
**Considered:** keeping the per-change recording model as today, just with fewer fields (rejected —
the operator wants a single global place to change these, with per-run override layered on top,
rather than re-answering per change).

### `/flow-settings` writes global defaults; a run can override, recorded per-run

**ID:** settings-scope
**Status:** active
**Chosen:** `/flow-settings` sets defaults for new `/flow` runs (implementer/fixer/reviewer models +
reviewers list) in the store. A given run can override via plain-language session instruction
(same mechanism as today's model-policy override); the override is recorded back to the store
against that specific run, never promoted to a new global default.
**Considered:** an explicit `--session`-scoped flag on `/flow-settings` for overrides (rejected —
operator chose the existing plain-language mechanism, no new flag).

### `/flow-fast` (renamed from `/flow-light` mid-conversation): shorter brainstorming + Primary-only reviewer

**ID:** flow-fast-scope
**Status:** superseded by flow-fast-cancelled
**Chosen:** a second command, not a mode flag. Shortens brainstorming (fewer clarifying rounds
before the convergence gate accepts) and reduces the review panel to Primary only, skipping the
other two required slots.
**Considered:** no separate command at all, since `review-panel-fixed-3` and `ask-options-removed`
already lighten plain `/flow` considerably (rejected — operator wants an even lighter path for
small/obvious changes). Also considered: skip brainstorming entirely (rejected in favor of
shorter-not-skipped) and skip the review panel entirely (rejected in favor of Primary-only).

### `/flow-fast` is cancelled — not built at all

**ID:** flow-fast-cancelled
**Status:** active
**Chosen:** drop `/flow-fast` entirely. `review-panel-fixed-3` and `ask-options-removed` already
lighten plain `/flow` enough that a third command carrying its own shortened-brainstorming and
Primary-only-panel behavior is no longer worth the surface it would add — the same tradeoff
`flow-fast-scope`'s own "Considered" section weighed and initially rejected, revisited and decided
the other way once the rest of the design had settled.
**Considered:** keeping it as designed under `flow-fast-scope` (superseded by this entry, during
implementation, once its own commit was already dispatching — no `/flow-fast` code had landed at
the point of cancellation, so nothing needed to be reverted, only task 10 dropped from the plan).

### Stats UI: keep 4 views, cut 4, `RunDetail` kept as a drill-down page

**ID:** stats-ui-cut
**Status:** active
**Chosen:** keep `StateBoard`, `Trend`, `CacheEfficiency` (the three the ticket names) plus
`RunDetail` (a drill-down page reached from a row click, not a standalone dashboard). Cut
`CostPerChange`, `ModelComparison`, `PanelEconomics`, `ReworkRate` and untangle their backend
support from the surviving views' shared code.
**Considered:** cutting `RunDetail` too, leaving only the 3 named views (rejected by the operator —
it's a detail page, not a fourth dashboard).
**Note:** the repository also has a 5th surviving top-level dashboard, `StageLeaderboard`
(registered in `App.tsx`), which this decision never names and this change never touches — it was
out of scope from the start. The 4-kept-plus-4-cut accounting above covers only the views this
decision actually addresses, not the full set of top-level dashboards in the app.

### `/flow-research` staging: outside `spectre/changes/`, seeds but never skips brainstorming

**ID:** flow-research-staging
**Status:** active
**Chosen:** `/flow-research` writes notes to a location outside `spectre/changes/` (this change's
own `docs/superpowers/research/kan-326-myflow-rework.md` is the first instance of the pattern,
written by hand pending the mechanized version). `/flow`'s brainstorming stage, on finding staged
notes for a topic, seeds from them but still runs its full interactive round.
**Considered:** having `/flow-research` create the change itself, i.e. run what's today's
`spectre new` (rejected — that stays `/flow`'s job; a research session shouldn't force a change
into existence). Also considered: brainstorming skipping straight to artifact-writing when staged
notes exist (rejected — still brainstorm, but seeded, per the operator's explicit choice).

### Staging notes follow a strict, fixed structure

**ID:** flow-research-artifact-structure
**Status:** active
**Chosen:** a staging note is not free-form prose — it follows a fixed section structure every
`/flow-research` session fills in the same way, so `/flow`'s brainstorming stage (task 11) can
parse it mechanically rather than re-reading loose prose. This change's own
`docs/superpowers/research/kan-326-myflow-rework.md` is the closest existing example but is **not**
itself the fixed structure — it predates this decision and was written by hand before the format
was defined; the actual section list (its headings, and which are required vs. optional) is
task-time work, not fixed here. At minimum it needs: a source line (Jira key/URL or "none"), one
section per topic/thread discussed, and — per `flow-research-depth` above — a step-by-step
breakdown section that is always present, since that mode is now the default rather than opt-in.
**Considered:** leaving staging notes free-form, as this change's own note is (rejected — a
brainstorming stage that has to parse arbitrary prose to seed from is exactly the fragility a fixed
structure avoids).

### `/flow-research` goes deeper: more investigation, more questions

**ID:** flow-research-depth
**Status:** active
**Chosen:** `/flow-research` does noticeably more legwork than today's `myflow-research` before
treating a topic as understood — more codebase investigation (broader searches, more files read,
not stopping at the first plausible answer) and more clarifying questions put to the operator
before writing a staging note, rather than accepting a thin answer and moving on. The exact
stopping rule (a round count, a convergence test like `/flow`'s own brainstorming loop, or an
operator-facing "anything else to dig into?" offer) is left to task-time design — this decision
fixes the direction (more, not less) and that it applies to both halves of the mode: investigation
depth and question count. `/flow-research` also produces a structured step-by-step analysis of the
topic under discussion **by default** — name/what/why/uses-style breakdown, the same shape this
change's own research conversation produced for the pipeline's stages
(`docs/superpowers/research/kan-326-myflow-rework.md`'s stage-audit section) — rather than only
prose discussion.
**Considered:** leaving `myflow-research`'s current depth unchanged and only adding the
staging-write mechanism (`flow-research-staging`) on top (rejected — raised explicitly as its own
requirement, separate from staging). Also considered, in order: the step-by-step breakdown as
opt-in-only, triggered by explicit operator request (this was the operator's own first answer,
superseded moments later in the same conversation by the choice recorded above — default output,
not opt-in).

### `do.document-fix` stays a separate stage from `do.sdd-tdd`

**ID:** document-fix-stays-separate
**Status:** active
**Chosen:** keep them as two marks — the fix note needs to exist before the implementer touches
code.
**Considered:** folding them into one stage to reduce stage count (raised as a candidate cut,
explicitly not adopted).

### MD cleanup: rework-driven first, general pass second

**ID:** md-cleanup-sequenced
**Status:** active
**Chosen:** two passes. First, whatever this rework itself obsoletes (roster logic,
`ask-options` text, `publish-proposal` steps, the old start/do/finish split), done inline while
implementing the decisions above. Second, a general verbosity/redundancy pass over the contract
files, done afterward once the new shape has settled — independent of this rework, out of this
change's scope.
**Considered:** doing only the rework-driven pass and treating general cleanup as never-scheduled
(rejected — operator wants both, sequenced, not one substituting for the other).

### Open questions

### Exact file split for `/flow`'s skill content

**ID:** flow-file-split
**Status:** open
**Why it is open:** "organized by topic, not by old command boundary" was decided
(`flow-rename-content-split`), but the actual file boundaries (how many files, what topics) were
not — left for whoever implements this to propose against the actual content.
**What it affects:** the shape of `skills/flow/` (or equivalent) on disk; doesn't affect any
runtime behavior decided above.

### Whether `finish.write-in-progress` (run 1) folds away

**ID:** write-in-progress-fold
**Status:** open
**Why it is open:** flagged as a candidate (it writes `IN_PROGRESS` → `IN_PROGRESS`, no actual
state change, telemetry-only purpose) but not firmly decided — unlike `move-in-review-fold` and
`workspace-export-lint-merge`, which were adopted.
**What it affects:** whether run 1 has one fewer top-level stage mark; no behavior change either
way.

### Staging-file discovery/adoption mechanism for `/flow-research` → `/flow`

**ID:** research-staging-mechanism
**Status:** open
**Why it is open:** `flow-research-staging` fixes the shape (staging location, seed-not-skip) but
not the mechanism — how `/flow`'s brainstorming stage discovers a staging file exists for a given
topic/change name, and what happens to the staging file once its content is adopted into the
change (deleted? left in place? cross-referenced?).
**What it affects:** the concrete implementation of `flow-research-staging`; needs resolving before
that stage can be built.

### Backend untangling plan for the stats UI cut

**ID:** stats-backend-untangle
**Status:** open
**Why it is open:** grep found the four cut views' backend support spread across 5 files
(`api/stats.go`, `store/aggregate.go`, `store/pricing.go`, `store/stageruns.go`, `sweep/sweep.go`)
<!-- measured: grep -rli "cost-per-change|costperchange|model-comparison|modelcomparison|panel-economics|panelecon|rework-rate|reworkrate" stats/internal @ 2da91eb -->
intermixed with code the 4 surviving views also use. No per-symbol plan exists yet for
what's safe to delete versus what must stay.
**What it affects:** scope and risk of the `stats-ui-cut` decision's backend half; needs its own
investigation pass during implementation.
