## Context

The `/myflow-*` pipeline is a three-state machine (`STARTED` → `IN_PROGRESS` → `FINISHED`) driven
by three commands, each ending in the state named after it, plus a read-only `/myflow-status`. Two
of the three human gates (reading the proposal, reviewing the diff) sit *inside* a stage; the third
kind of friction is purely inter-stage — the operator has to notice a stage ended cleanly and type
the next command, even when nothing actually blocked moving forward. `/myflow-fast` removes that
inter-stage friction without touching what happens inside any stage.

## Goals / Non-Goals

**Goals:**
- One command that reaches `IN_PROGRESS` from nothing in a single invocation, pausing only for
  brainstorming's own interactive questions (unchanged from `/myflow-start`).
- One command that, once the operator has reviewed the diff and chosen a landing route, reaches
  `FINISHED` in the same invocation whenever nothing external blocks it.
- No change to what happens *inside* `/myflow-start`'s or `/myflow-do`'s stage content, the review
  panel's handoff bar, or any contract's git/handoff/Jira rules.

**Non-Goals:**
- Auto-answering brainstorming's design questions. Confirmed with the operator: brainstorming stays
  fully interactive, identical to `/myflow-start` today.
- Polling for an external PR merge. `/myflow-fast` stops and hands off when a route needs a merge
  outside its control, exactly as `/myflow-finish` run 1 does today.
- A new state, a new state-file field, or a new git boundary. `/myflow-fast` is a sequencing layer
  over the existing three-state machine, not a fourth state or a parallel pipeline.

## Decisions

### `/myflow-fast` chains stage 1 and stage 2 in one invocation

**ID:** `chain-start-and-do-in-one-call`
**Status:** active
**Chosen:** The first invocation (no state file yet) runs full interactive brainstorming, then —
without the operator typing a second command — continues straight into `/myflow-do`'s SDD+TDD
implementation and review panel, ending at `IN_PROGRESS`. There is no human gate between
brainstorming converging and implementation starting, so nothing is lost by not stopping there.
**Considered:** Stopping at `STARTED` and requiring a bare re-invocation to reach `IN_PROGRESS` —
mirrors today's two-command shape exactly, but reintroduces the exact friction (notice the state,
type the next command) this change exists to remove, for a boundary with no human gate.

### Fix vs. integrate disambiguation

**ID:** `arg-present-means-fix`
**Status:** active
**Chosen:** At `IN_PROGRESS`, an argument means fix instructions (resumes `/myflow-do`-style);
bare invocation means proceed to the integrate question. No extra confirmation prompt.
**Considered:** Always asking "fix or integrate?" explicitly — unambiguous, but adds a manual step
on every re-invocation, working against the "one command, one answer" goal.

### Chaining merge+push straight to archive

**ID:** `merge-push-auto-archives`
**Status:** active
**Chosen:** Once the operator picks merge+push, nothing external blocks proceeding — so
`/myflow-fast` runs `/myflow-finish` run 1 then run 2 in the same invocation, ending at `FINISHED`.
**Considered:** Stopping after merge+push like a normal `/myflow-finish` run 1 — mirrors the
existing two-run contract exactly, but leaves a manual second invocation with nothing actually
gating it.

### PR route stops and hands off

**ID:** `pr-route-stops`
**Status:** active
**Chosen:** Opening a PR still requires an external human merge (on the forge, possibly by someone
else), so `/myflow-fast` cannot auto-continue past it. It prints the PR URL and stops at
`IN_PROGRESS`, same as `/myflow-finish` run 1 does today.
**Considered:** Polling until merged, then auto-archiving — would tie up a long-running session or
need a background job for an event this pipeline has no mechanism to observe; disproportionate to
the problem.

### Model defaults: Sonnet everywhere

**ID:** `sonnet-everywhere`
**Status:** active
**Chosen:** `/myflow-fast`-created changes record `implementation`, `reviewPanel` and `panelFix`
all as Sonnet by default, deviating from the normal pipeline's Opus-for-implementer/panel-fix
default, still overridable at the same question `/myflow-start` already asks.
**Considered:** Keeping the normal pipeline's Opus/Sonnet/Opus split — rejected in favor of lower
cost/latency per run, since an operator reaching for `/myflow-fast` has already chosen speed over
the normal default.

### Panel roster: light

**ID:** `light-roster-default`
**Status:** active
**Chosen:** `reviewPanelRoster: light` by default, matching the project-wide default already set by
KAN-110.
**Considered:** `standard` — broader coverage, slower fix-cycles; rejected to keep `/myflow-fast` at
the fastest already-supported preset rather than introduce a fourth roster just for this command.

### Keep the name, scope the retired-vocabulary guard

**ID:** `keep-name-scope-guard`
**Status:** active
**Chosen:** Discovered during implementation: `scripts/check-vocabulary.sh` blacklists bare
`myflow-fast` as retired vocabulary — a leftover from the five-state pipeline's own, unrelated
`/myflow-fast` command. Confirmed with the operator on 2026-08-09: keep this command's name rather
than rename it. A boundary-class narrowing (the `code-review` exception's own technique) was tried
first and rejected empirically: this skill's own frontmatter and prose self-references are lexically
identical to the retired command's bare-word shape, so no boundary distinguishes them. The entry was
dropped from the mechanical blacklist instead — the same treatment this guard already gives
`checkpoint` and the `effort` values — with the accepted gap (a bare, non-path, non-backtick
reintroduction of the old retired command no longer mechanically caught) documented in the guard's
own comment.
**Considered:** Renaming the new command (e.g. `/myflow-quick`) — avoids touching the guard at all,
but the name was already settled across the Jira issue, every planning artifact and the published
proposal artifact; renaming this late would be pure churn for a collision the guard itself can be
taught to resolve.

### No proposal artifact

**ID:** `no-proposal-artifact`
**Status:** active
**Chosen:** `/myflow-fast` skips publishing the proposal HTML artifact. The operator participated
in the brainstorming dialogue directly, in the same session, so a separate published summary of
that same conversation adds a step without adding information.
**Considered:** Publishing it anyway for parity with `/myflow-start` — rejected as pure overhead for
a command whose whole purpose is cutting steps that don't carry new information.

### `/myflow-fast` is a driver, not a fourth state

**ID:** `driver-not-new-state`
**Status:** active
**Chosen:** `/myflow-fast` reads and writes the exact same state file shape and the exact same three
states as the rest of the pipeline. It is documented as a command that can call into more than one
existing command's stage content within a single invocation, never as introducing new state-file
fields or a fourth pipeline state.
**Considered:** A parallel lightweight state machine with its own file — rejected in the original
brainstorming round; see `docs/superpowers/specs/2026-08-09-kan-111-myflow-fast-design.md` for that
comparison. Reusing the existing machine means every other command's tooling (`/myflow-status`, the
Jira integration, the finish contract) already understands a change `/myflow-fast` created or
touched, with nothing new to teach them.

## Risks / Trade-offs

- **A single invocation now spans stages that used to be separate commands, so a long-running
  session covers more ground before returning control.** → Mitigated by the fact that the only new
  auto-continuation points (start→do with no gate between them, and merge+push→archive with no gate
  between them) were already gateless; nothing that used to pause for a human now doesn't.
- **The fix-vs-integrate disambiguation by argument presence could misfire if an operator wants to
  integrate but happens to pass stray text.** → Mitigated by the integrate route itself still asking
  an explicit question (PR/merge+push/manual) before doing anything irreversible, so a misrouted bare
  vs. fix call is caught at that prompt rather than silently taking an unintended action.
- **Two commands (`/myflow-do`, `/myflow-finish`) now have their stage content invoked from a third
  entry point.** → Mitigated by citing rather than re-deriving: `myflow-fast-command`'s spec and its
  skill point at the existing skills' sections instead of duplicating their logic, so a future change
  to `/myflow-do`'s panel behavior does not need a matching edit in two places.

## Migration Plan

Purely additive: a new skill, a new command file per harness, and doc/spec updates naming the new
command. No existing command's behavior, state-file shape, or contract changes. No rollback beyond
reverting the added files, since nothing else depends on `/myflow-fast` existing.

## Open Questions

*(none — every question raised during planning was answered; see the brainstorming design doc's own
"Open questions" section, also empty)*
