# KAN-326 — Myflow rework/redesign — research notes

Source: [KAN-326](https://tweety53.atlassian.net/browse/KAN-326), explored via `/myflow-research`.
Staging notes only — no spectre change exists yet. `/flow`'s (or `/myflow-start`'s) brainstorming
stage should read this file when it exists for a topic, treat it as seeded input, and still run its
full interactive round rather than skipping it.

## 1. Rename to `/flow`, single command

- `/myflow-fast` becomes `/flow`; `/myflow-start`, `/myflow-do`, `/myflow-finish` are deleted as
  user-facing commands.
- **Content fate:** `myflow-fast/SKILL.md` today works by citing sections of `myflow-start`/`-do`/`-finish`
  rather than containing the logic itself (~2700 lines of cited content across the three). Deleting
  the citation targets means re-deriving a new file structure **organized by topic**, not preserved
  along the old start/do/finish boundary (e.g. brainstorm.md, implement.md, integrate.md — exact
  split TBD at implementation time).
- `/myflow-status` and `/myflow-research` are not named in the ticket; presumably survive as
  `/flow-status` and `/flow-research` (see §6).

## 2. Pipeline stage audit + cuts

Full stage-by-stage inventory (name, what/why, uses — skills/scripts/CLI/stats app) produced in the
research conversation; canonical source for stage definitions remains `README.md`'s Level 1/2 table
until this rework rewrites it.

### Decided changes to the pipeline itself

- **`start.ask-options` removed.** No planning-effort/model/panel-roster question round on a
  creating run.
- **Models: single default (sonnet) for implement, fix, and review**, operator-overridable at any
  time via plain-language session instruction (unchanged mechanism from today).
- **`start.create-artifacts`** — confirmed unchanged: this stage only ever created
  `proposal.md`/`design.md`/`tasks.md`; a spec edit (`spectre/specs/<capability>.md`) is *planned*
  here as a task, never written here. No cut needed — the audit's original wording implying spec
  deltas were written at this stage was imprecise, not a description of dead functionality.
- **`start.publish-proposal` deleted.** Matches `/myflow-fast`'s existing creating-run behavior
  (`artifactUrl: null`) — `/flow` adopts the fast path's behavior everywhere.
- **Review panel redesigned:** always exactly 3 required slots by default (no roster presets).
  Bugbot and Security become fully optional, dispatched **only** when the operator explicitly asks
  for them, at any point during the run. Removes: roster presets (`light`/`standard`/`full`), the
  diff-size/touched-area trigger heuristics ("Optional slot selection"), and automatic full-roster
  escalation on re-run.
- **`do.run-instructions` moves to immediately before `do.write-in-progress`** — was
  `run-instructions → workspace-export → lint-and-test → stage-diff → write-in-progress`; becomes
  `workspace-export → lint-and-test → stage-diff → run-instructions → write-in-progress`. Run
  instructions are for the human handoff, so producing them last (right before the state write that
  ends the run) fits better than computing them mid-stream.
- **`finish.sync-archive`** — confirmed already unconditional today (no ask, no `--force`, refuses
  rather than asks on unchecked tasks). No change needed.
- **`finish.move-in-review`** folds into `finish.landing-routes` — it always fires on every landing
  route with no distinct condition of its own; one less top-level mark, still logged as a sub-step.
- **`do.workspace-export` + `do.lint-and-test`** merge into one "verify" stage — both are
  unconditional automated pass/block checks with nothing interactive between them.
- **Model policy fields collapse**: `models.implementation` / `models.reviewPanel` / `models.panelFix`
  → single `models.default` field, since all three now default to the same value and override the
  same way. Removes two-thirds of that config surface.
- **`reviewPanelRoster` field removed entirely** from the state file — nothing replaces it except
  possibly a running list of operator-requested additions (Bugbot/Security) for the run.
- **`do.document-fix` stays a separate stage from `do.sdd-tdd`** — considered folding, kept separate:
  the fix note needs to exist before the implementer touches code.
- **`finish.write-in-progress` (run 1)** — candidate to fold into `finish.commit-two` or
  `finish.landing-routes` as an outcome field rather than a standalone stage, since the state doesn't
  actually change here (`IN_PROGRESS` → `IN_PROGRESS`); its only purpose is telemetry. Flagged, not
  firmly decided.

### `STARTED` state redefined

`STARTED` no longer means "design approved and proposal published" (today's meaning). It becomes a
kickoff marker: written the moment the operator invokes the command, before brainstorming even
begins — "the task is started by operator," the very first stage. The two/three-state shape of the
pipeline itself is otherwise unchanged (`STARTED` → `IN_PROGRESS` → `FINISHED`).

## 3. `/flow-settings`

- Writes **global defaults** for new `/flow` runs to the stats app store: implementer model, fixer
  model, reviewer model(s), and the reviewers list (which agents make up the required panel).
- The stats app is both **validator** (a strict, defined list — rejects values not on it) and
  **source of truth** for these settings — not a per-change state file copy.
- Any given `/flow` run can override the global defaults via plain-language session instruction
  (same mechanism as today's model-policy override) — the override is recorded back to the store
  against that specific run, not persisted as a new global default.

## 4. `/flow-fast` (was `/flow-light`)

A real second command, not a mode flag or a merge into `/flow`. Two things it does differently from
plain `/flow` once the §2 cuts land:

- Shorter brainstorming — fewer clarifying rounds before the convergence gate accepts.
- Panel reduced to Primary only — skips the other two required slots.

## 5. Stats UI cut

- **Keep:** `StateBoard.tsx`, `Trend.tsx`, `CacheEfficiency.tsx`, `RunDetail.tsx` (drill-down page,
  not a fourth dashboard — kept).
- **Cut:** `CostPerChange.tsx`, `ModelComparison.tsx`, `PanelEconomics.tsx`, `ReworkRate.tsx`.
- **Backend removal is not a clean delete** — grep found the four cut views' backend support spread
  across `stats/internal/api/stats.go`, `stats/internal/store/aggregate.go`,
  `stats/internal/store/pricing.go`, `stats/internal/store/stageruns.go`, and
  `stats/internal/sweep/sweep.go`, intermixed with code the 4 surviving views also depend on.
  Untangling this is implementation-time work, not something resolved here.

## 6. `/flow-research`

Confirmed: this is already `myflow-research` (this skill), which replaced `openspec-explore` in the
2026-08-25 spectre cutover (`docs/superpowers/plans/2026-08-25-myflow-on-spectre.md:101-167`). The
ticket line describing "openspec explore" predates that cutover and is stale — no new functionality
needed, just the rename alongside everything else in §1.

**New requirement surfaced in this research conversation, not in the original ticket text:**
`/flow-research` needs a way to leave spectre-adjacent output behind for `/flow` to pick up later,
without creating the change itself (that stays `/flow`'s job — a research session shouldn't force a
change into existence). Landed design:

- `/flow-research` writes to a **staging location outside `spectre/changes/`** — this file is the
  first instance of that pattern (`docs/superpowers/research/<topic>.md`), not yet a mechanized
  feature.
- When `/flow`'s brainstorming stage starts and finds staged research notes for the topic, it
  **still runs its full interactive brainstorming round** — the notes seed it (faster convergence),
  they don't let it skip straight to writing artifacts.
- The actual mechanism (how `/flow` discovers a staging file exists for a given topic/change name,
  what happens to the staging file once adopted) is undesigned — this note only fixes the shape
  (staging + seed, not staging + skip).

## 7. MD file cleanup

Two-pass, sequenced:

1. **Rework-driven cleanup**, done inline while implementing §1-6 — whatever this rework itself
   obsoletes (roster logic, `ask-options` text, `publish-proposal` steps, the old start/do/finish
   split).
2. **General verbosity/redundancy pass**, done afterward once the new shape has settled — independent
   of this rework, targeting prose in the contract files (`pipeline.md`, `model-policy.md`, etc.)
   that over-explains, restates, or could be cut regardless of what else changes.

## Open / undesigned

- Exact new file structure for `/flow`'s SKILL.md(s) once split by topic (§1).
- Whether `finish.write-in-progress` (run 1) actually folds away, or stays a standalone mark (§2).
- The staging-file discovery/adoption mechanism for `/flow-research` → `/flow` (§6).
- Full backend untangling plan for the stats UI cut (§5).
