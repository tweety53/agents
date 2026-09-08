## Context

The tag vocabulary is canonical in `skills/flow-contracts/plan-provenance.md`; what
`scripts/check-plan-provenance.py` enforces is canonical in
`skills/flow-contracts/plan-provenance-guard.md`. The planner template already mandates tagging
(`skills/flow/brainstorm-planner.md`, section D: load the vocabulary, tag every fenced block and
every numeric claim, keep untaggable code as `unverified:`), and the guard runs at that stage only
where a project declares it. This repository's guards split into two families — shipped guards,
symlinked into `skills/<skill>/scripts/` and invoked by the skills themselves, and
project-configured guards, run only through a project's `## lint`. The provenance guard sits in
the second family; KAN-444 needs it in the first.

## Approaches considered

- **Ship the guard (chosen)** — the enforcement already exists and already rejects exactly what
  KAN-444 names; only the wiring is missing.
- **Template wording only (rejected)** — the mandate already has mandatory wording; without the
  guard wired in, projects that never declare it never reject an untagged plan.
- **Bolt the scan into `check-plan-shape.sh` (rejected)** — couples two guards with independent
  exit-code contracts, and a provenance hit would surface under the wrong guard's name.

## Design

**Behavior.** The writing-plans stage runs `check-plan-provenance.sh` by basename — resolving to
`<skill-dir>/scripts/` per Guard resolution (`skills/flow-contracts/pipeline.md`) — immediately
after `check-plan-shape.sh`, unconditionally, in every project, and fixes any hit before the plan
is accepted. A hit names its `file:line`; the remedy is the guard's own exit-1 banner: tag the
block, quote the number, or reword the line.

**Wiring.** Both files are symlinked into the skill's scripts directory with the same
relative-target shape the `check-plan-shape` links already use:

```bash verified:the relative-target shape read from skills/flow/scripts/check-plan-shape.sh's existing symlinks
ln -s ../../../scripts/check-plan-provenance.sh skills/flow/scripts/check-plan-provenance.sh
ln -s ../../../scripts/check-plan-provenance.py skills/flow/scripts/check-plan-provenance.py
```

The wrapper execs `$SCRIPT_DIR/check-plan-provenance.py`, so both links are required — the `.sh`
link alone would look for the interpreter script inside the skills directory. Rule 3 of
`check-guard-symlinks.sh` drops `check-plan-provenance.sh` from its project-configured exemption
map, and the comment above the map names the shrunken family. `skills/flow/SKILL.md`'s
guard-presence union list gains `check-plan-provenance.sh`; the `.py` needs no entry of its own,
because the presence check derives a guard's sibling dependencies by grepping its source for
`$SCRIPT_DIR/<name>`.

**What does not change.** The vocabulary and the enforcement contract are untouched. The scan
scope stays the three planning files per non-archived change. Build-green stays
project-configured. This repository's own `## lint` entry keeps invoking
`scripts/check-plan-provenance.sh` from the repository root — the shipped and lint invocations
run the same script through two paths.

**Dogfooding.** The plan's second task runs the guard over this change's own planning files, so
the shipped invocation is exercised before the review panel sees it.

## Decisions

### Ship the guard rather than restate the mandate

**ID:** ship-provenance-guard
**Status:** active
**Chosen:** symlink the guard into the skill's scripts directory and invoke it unconditionally at
the writing-plans stage — the enforcement exists; only its reach was missing.
**Considered:** template wording only — rejected because the mandate already had mandatory
wording and no effect outside projects that declare the guard; bolting the scan into
`check-plan-shape.sh` — rejected because it couples two guards' exit-code contracts and
misattributes provenance hits.

### Enforce at the writing-plans stage only

**ID:** enforce-at-writing-plans
**Status:** active
**Chosen:** reject at the stage that produces the plan, before its `## Plan` return.
**Considered:** a second unconditional run when implementation loads context — rejected as a
duplicate on every run; a plan that reached implementation was already gated at the writing-plans
stage.

## Open questions

None.
