## Context

The store already records per-project **incidents** (KAN-451) — a guard actually costing time —
and `gather-dispatch-context.sh` injects them into every dispatch bundle. Hazards are the proactive
sibling: warnings recorded *before* they cost time, shape-scoped, injected by the same mechanical
path. The two KAN-423 lessons that fit every repository already live in implement.md's guard
paragraphs (KAN-442/KAN-423); what has no home is the project-specific warning no global skill file
should carry. The spectre specs tree is empty, so the change edits no capability spec.

Injection point is the dispatch context bundle, not the prompt: the bundle is gathered by a fixed
step at every dispatch boundary (implementer bundles per task, whole-plan for panel slots and fix
agents), so injection there cannot be forgotten — the exact failure KAN-423 recorded was a human
remembering to paste a warning.

## Shape

`0019_hazards.sql`, following `0018_guard_log.sql`'s conventions (new file; header explains the
schema):

```sql verified:conventions read from 0018_guard_log.sql at plan time
CREATE TABLE hazards (
  id          BIGSERIAL PRIMARY KEY,
  project_key TEXT NOT NULL REFERENCES projects(project_key),
  name        TEXT NOT NULL,
  body        TEXT NOT NULL,
  applies     TEXT NOT NULL CHECK (applies IN ('all','cross-repo','single-repo')),
  active      BOOLEAN NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_key, name)
);
CREATE INDEX hazards_project_key ON hazards (project_key);
```

`name` is a short stable identifier (`commit-fields-guard-single-repo`); `body` is the warning text
a dispatch carries verbatim; `applies` is the closed shape vocabulary. UNIQUE makes re-adding an
existing hazard a clean store refusal, not a duplicate row.

- `internal/store/hazards.go`: `AddHazard(projectKey, Hazard)`, `ListHazards(projectKey, shape,
  includeInactive)` — newest first; `shape` is a validated closed-set value filtering to
  `applies IN ('all', <shape>)`; `includeInactive` lifts the `active` filter for the operator's
  `-all` view — and `RetireHazard(projectKey, name)` setting `active=false`.
- `internal/api/hazards.go` + routes in `server.go`: `POST/GET /api/v1/hazards/{project}`,
  `PATCH /api/v1/hazards/{project}/{name}` (retire).
- `cmd/flow/hazard.go`: `flow hazard add -name -text -applies` → `recorded: hazard`;
  `flow hazards [-shape <s>] [-all]` — without `-shape` the operator sees every active hazard
  regardless of `applies` (an unfiltered listing is the operator's view); with `-shape` it filters
  as the store does; `-all` includes retired rows. `flow hazard remove -name` (retire). Store
  unreachable behaves like every record write: journal/fallback, one warning line, exit 0 — never
  a gate. The bundle script always passes `-shape` — the caller's shape, or the literal `all` when
  its caller gave none — so the fail-open rule lives at the one call site that needs it.

## Injection

`gather-dispatch-context.sh` grows an optional 8th positional argument `<shape>`. It calls
`flow hazards -C <worktree> -shape <shape>` and renders a `## hazards` section after
`## incidents` — one bullet per hazard, `- **name (applies):** body` — inside the hashed body, so a
newly recorded hazard forces a bundle rebuild. Failure modes mirror incidents exactly: `flow`
absent or failing → skipped as `hazards (flow unavailable)`; empty result → `hazards (none)`; the
section never gates a run. Shape omitted or empty → only `all` hazards inject (fail-open).

The caller — implement.md's conductor — computes the shape once from its resolved worktree set
(worktree-resolution.md): more than one repository → `cross-repo`, otherwise `single-repo`, and
passes it on every gather call it makes. implement.md changes: the added argument in the gather
invocation plus one REQUIRED READING paragraph naming `## hazards` binding project-specific warning
text. That, plus the bundle, is the ticket's "conductor/implementer dispatch template auto-inject":
the bundle is part of every dispatch the template assembles.

## Verification

- Store tests against a real migrated schema; API handler tests; CLI flag/refusal tests following
  `record_test.go`'s shape.
- Guard-harness test for the script's new section — real `flow` invocation, empty result,
  unavailable-skip — however the existing harness exercises the incidents section.
- REPRODUCE, DON'T READ: the injection test exercises the real CLI→daemon→store path (UI-test stack
  on port 4174), never a fake.
- Seeding: both KAN-423 hazards recorded through the new CLI against the disposable UI-test stack
  (port 4174 — the dev daemon on 4173 is a protected service an agent never restarts), then one
  real gather run proving they inject — `commit-fields-guard-reverts-head-only` (`all`) and
  `commit-fields-guard-single-repo` (`single-repo`). The same three CLI commands, against the dev
  store, are the operator's post-restart step, stated in the run instructions.

## Decisions

### Closed shape vocabulary

**ID:** closed-shape-vocabulary
**Status:** active
**Chosen:** `applies` is one of `all` / `cross-repo` / `single-repo` — matches how the pipeline
already classifies a change (worktree-resolution's one-repo vs satellites), guardable, no fuzzy
matching.
**Considered:** free-form tags — more expressive, but the tag vocabulary drifts per project and the
matching rules get fuzzy (operator chose the closed set).

### CLI-only management

**ID:** cli-only-management
**Status:** active
**Chosen:** `flow hazard` subcommands only; no SPA surface — the requirement names the store and
the dispatch template, no UI is asked for (YAGNI).
**Considered:** CLI + SPA UI beside incidents — wider scope (components, PUT/DELETE, visual
baseline impact) for a need nothing stated.

### Inject via the context bundle

**ID:** inject-via-context-bundle
**Status:** active
**Chosen:** hazards render in `gather-dispatch-context.sh`'s bundle — the one site every
conductor-stage dispatch already passes through mechanically.
**Considered:** inlining into each dispatch prompt by the dispatcher — moves the incident's failure
("remember to paste") sideways ("remember to fetch") instead of removing it; a `.flow/project.md`
`## hazards` prose section — contradicts the requirement's "into the store as structured,
per-project records" and has no lifecycle.

### Retire, never delete

**ID:** retire-not-delete
**Status:** active
**Chosen:** `flow hazard remove` sets `active=false`; rows survive like incident history does.
**Considered:** hard delete — loses the record of what warnings a project carried when an incident
happened anyway.

### Seed the KAN-423 hazards

**ID:** seed-kan-423-hazards
**Status:** active
**Chosen:** record both known hazards for this project through the new CLI during verification —
the mechanism ships with its motivating data and the injection path is verified against real rows.
**Considered:** shipping the mechanism empty — leaves the motivating gap open and verifies
injection against nothing real (operator chose seeding).

## Open questions

<!-- empty by convergence: nothing left open at design approval -->
