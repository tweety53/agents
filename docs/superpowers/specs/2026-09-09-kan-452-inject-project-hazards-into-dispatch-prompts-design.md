# kan-452 — Inject project hazards into dispatch prompts automatically from the store

Design doc for `kan-452-inject-project-hazards-into-dispatch-prompts`, approved in the
brainstorming session of 2026-09-09. `spectre/changes/<name>/design.md` is canonical for the same
content; this file records the approved design and the approaches considered.

## Problem

KAN-423's self-review (flow-stats-app angle): the commit-fields-guard hazards that predicted that
incident (`commit-fields-guard-single-repo`, `commit-fields-guard-reverts-head-only`) lived only in
the operator's memory notes, so the first conductor dispatch of the run did not carry them and the
incident happened anyway; only the retry dispatch, written after the incident, included the
warning. Skills carry the lessons every repository needs; project-specific warnings have no home,
and dispatches must not depend on a human remembering to paste one.

## Approaches considered

- **A — Store entity, injected through the dispatch context bundle (chosen).** Hazards become a
  structured per-project store table managed by a `flow hazard` CLI; `gather-dispatch-context.sh`
  renders them as a `## hazards` section beside `## incidents`. Injection happens at the one site
  every conductor-stage dispatch passes through mechanically, so it cannot be forgotten.
- **B — A `.flow/project.md` prose section, no store.** Simplest possible, but contradicts the
  requirement's own words ("into the store as structured, per-project records") and has no
  lifecycle — stale warnings accumulate silently.
- **C — Store entity, inlined into each dispatch prompt by the dispatcher.** Moves the incident's
  failure ("remember to paste") one step sideways ("remember to fetch") instead of removing it.

## Design

- **Store** — migration `0019_hazards.sql`: `hazards(id, project_key, name, body, applies, active,
  created_at)` with `applies CHECK IN ('all','cross-repo','single-repo')`, `active` defaulting
  true, `UNIQUE (project_key, name)`; indexed on `project_key`. Store methods `AddHazard`,
  `ListHazards(projectKey, shape)` (active rows where `applies IN ('all', <shape>)`, newest first;
  shape optional), `RetireHazard` (sets `active=false` — never delete).
- **API** — `POST/GET /api/v1/hazards/{project}`, `PATCH /api/v1/hazards/{project}/{name}`
  (retire), beside the incidents routes.
- **CLI** — `flow hazard add -name -text -applies`, `flow hazards [-shape <s>] [-all]` (no
  `-shape` = unfiltered operator listing; `-shape` filters to `applies IN ('all', <s>)`; `-all`
  includes retired rows), `flow hazard remove -name`, in the `flow record incident` pattern; store
  unreachable = one warning line, exit 0, never a gate. The bundle script always passes `-shape` —
  the caller's shape, or the literal `all` when its caller gave none (fail-open).
- **Injection** — `gather-dispatch-context.sh` takes an optional 8th positional `<shape>`, calls
  `flow hazards -C <worktree> -shape <shape>`, renders `## hazards` after `## incidents` (one
  bullet per hazard, `- **name (applies):** body`) inside the hashed body. Skips as
  `hazards (flow unavailable)` / `hazards (none)`; never gates a run; shape omitted → `all`-only
  (fail-open).
- **Shape determination** — implement.md's conductor computes the change's shape once from its
  resolved worktree set (more than one repository → `cross-repo`, else `single-repo`) and passes it
  on every gather call; implement.md gains the added gather argument and one REQUIRED READING
  paragraph naming `## hazards` binding.
- **Verification** — store/API/CLI tests per existing patterns; a guard-harness test for the new
  section exercising the real CLI→daemon→store path (UI-test stack, port 4174); both KAN-423
  hazards recorded for this project through the new CLI, then one real gather run proving they
  inject.

## Operator decisions taken during brainstorming

- Closed shape vocabulary (over free-form tags) — operator choice.
- CLI-only management (no SPA surface) — operator choice, YAGNI.
- Seed both KAN-423 hazards as real rows during verification — operator choice.
