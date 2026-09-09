# kan-452-inject-project-hazards-into-dispatch-prompts

## Why

KAN-423's self-review (flow-stats-app angle): the commit-fields-guard hazards that predicted that
incident — `commit-fields-guard-single-repo`, `commit-fields-guard-reverts-head-only` — lived only
in the operator's memory notes, so the first conductor dispatch of the run did not carry them and
the incident happened anyway; only the retry dispatch, written after the incident, included the
warning. Skills carry the lessons every repository needs; there is no home for *project-specific*
warnings, and dispatches should not depend on a human remembering to paste one.

## What changes

- The store gains a per-project `hazards` table (migration `0019`): `name`, `body`, a closed
  `applies` shape (`all` / `cross-repo` / `single-repo`), an `active` flag, `UNIQUE (project_key,
  name)`.
- API routes beside the incidents pair: `POST/GET /api/v1/hazards/{project}`,
  `PATCH /api/v1/hazards/{project}/{name}` (retire — never delete).
- CLI, in the `flow record incident` pattern: `flow hazard add -name -text -applies`,
  `flow hazards [-shape <s>] [-all]`, `flow hazard remove -name`.
- `gather-dispatch-context.sh` renders a `## hazards` section after `## incidents`, filtered by the
  change's shape the caller passes; shape omitted → `all`-only (fail-open). The section never gates
  a run.
- `skills/flow/implement.md`: the conductor computes the shape once from its resolved worktree set
  (more than one repository → `cross-repo`, else `single-repo`) and passes it on every gather call;
  one REQUIRED READING paragraph names the section binding.
- Both KAN-423 hazards are recorded through the new CLI during verification — against the
  disposable UI-test stack (the dev daemon on 4173 is protected and is never restarted by an
  agent), and the same three commands ship in the run instructions for the dev store once the
  operator restarts the daemon on the new build.
