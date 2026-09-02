# flow — rationale

Reasoning behind `skills/flow/`'s phase files: why a rejected alternative was rejected, which
design.md decision a passage implements, and choices a task made on its own. Moved here verbatim
from the run-loaded files; **no `/flow` run loads this file.** Each heading names the source file
and the section the passage came from.

## SKILL.md — Stage keys

Every stage `/flow` marks uses a `flow.*` key, minted fresh for this command rather than reusing
`start.*`/`do.*`/`finish.*` — the old namespacing was tied to the three commands this one replaces,
and reusing it here would misdescribe a stage that no longer runs under the command its key names.
**This is a design decision this task makes, not one resolved upstream of it**: the alternative —
keeping the old keys, unioned, as `/myflow-fast` did — was rejected because several stages here are
not the old stage unchanged (`flow.verify` merges two, `flow.landing-routes` absorbs two more,
`flow.kickoff` is new), so a reused key would sometimes mean something different than its history
records.

## SKILL.md — preamble

> one state file, the same three states, content reorganized by topic rather than by the old
> start/do/finish boundary (design.md's `flow-rename-content-split`). This file is the router: it

## SKILL.md — Model resolution

> **`DEFAULT_MODEL` is the model for all three roles this run dispatches on** — the implementer
> (`skills/flow/implement.md`), every panel slot that takes a model override, and the panel-fix
> subagent (`skills/flow/review-panel.md`) — per design.md's `model-default-sonnet`: one default,
> chosen once per run, not three per-role defaults.

## SKILL.md — Guardrails

> **Generate this run's session token once, right here, before the first mark any phase file below
> makes — a short, unique literal string — and reuse that exact same value at every `stage begin` this
> run makes, including inside every phase file it dispatches into.** One run, one token, never a fresh
> one per mark or per phase file (design.md's "one token per session, not one per mark").

## brainstorm.md — preamble

> Superpowers Basic Workflow steps **#1** (brainstorming) and **#3** (writing-plans), intertwined
> with spectre artifact creation, run here — the same content `/myflow-start` carried, minus the
> options-question round and the proposal publish, both removed per design.md (`ask-options-removed`,
> `publish-proposal-removed`).

## brainstorm.md — A. Resolve the change and write `STARTED`

> This is design.md's `started-redefined`: `STARTED` is a kickoff marker,
> "the operator started this," not (as under the old `/myflow-start`) a record that a design was
> approved and a proposal published.

> `planningEffort` and `models.default` are written `null` and stay `null` for the life of the
> change: `/flow` asks no planning-effort or model question on a creating run
> (`ask-options-removed`), and models are resolved per run from the settings store
> (`model-default-sonnet`, `settings-scope`), not recorded per change. `artifactUrl` stays `null` —
> `/flow` publishes no proposal artifact (`publish-proposal-removed`).

## brainstorm-planner.md — Seed from a staged research note

> This implements design.md's `flow-research-staging`, the *discovery* half of open
> question `research-staging-mechanism` (the *write* half is `skills/flow-research/SKILL.md`'s own
> job):

> This is
> `flow-research-staging`'s explicit choice: seed, never skip.

> This resolves the remaining half of `research-staging-mechanism`
> left open by design.md: a staging note that outlives its adoption is a second, driftable copy of
> what the change's own `design.md` now states canonically, and `<project>/docs/superpowers/research/` is meant
> to hold notes still waiting for a home, not a permanent archive of every note that found one.

## integrate.md — preamble

> Run 1 of what was `/myflow-finish`, unchanged in procedure except two folds design.md decides:
> `move-in-review-fold` (the Jira "move to In Review" step becomes a sub-step of `flow.landing-routes`
> rather than its own mark) and this task's own resolution of open question `write-in-progress-fold`
> (below).

## integrate.md — rebase

> — per design.md's `rebase-is-a-confirmed-choice`: the rebase never runs on its own, only after the
> operator picks this option, and only against the worktree(s) that actually moved

> , per design.md's `never-auto-abort`

## integrate.md — Scoped re-verification

> , per design.md's `scoped-reverify-not-full-suite`

## integrate.md — landing routes

> This stage carries three sub-steps under one mark, per this task's own resolution of open question
> `write-in-progress-fold` and design.md's `move-in-review-fold`: the git route, the state write, and
> the Jira transition — three sub-steps that used to be three separate top-level marks
> (`finish.landing-routes`, `finish.write-in-progress`, `finish.move-in-review`) now recorded as one.
> **This task's own choice**: fold `write-in-progress` in alongside `move-in-review`, rather than
> leave it standalone. The write is a genuine no-op (`IN_PROGRESS` → `IN_PROGRESS`, nothing changes
> but `prUrl` and `updatedAt`/`updatedBy`) exactly as design.md's open question describes, and by the
> time this mark reaches it, the route sub-step immediately above has already decided what `prUrl`
> becomes — folding the write in with the route that produces its one real input, and with the Jira
> transition that only ever follows a successful route, keeps one mark's three sub-steps in the causal
> order they already have to run in, rather than three marks whose middle one records nothing a reader
> could not already infer from the other two.
