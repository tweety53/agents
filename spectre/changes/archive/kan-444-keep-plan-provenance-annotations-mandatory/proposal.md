# kan-444-keep-plan-provenance-annotations-mandatory

## Why

KAN-423's plan carried provenance tags on every snippet and every numeric claim, and the outcome
was the cleanest implementation run on record: "19/19 tasks landed first-pass", a single Important
panel finding, fixed in one panel-fix pass. KAN-444 — a self-review finding from KAN-423
(flow-improvement angle) — asks that this outcome stay reproducible: the tagging convention stays
mandatory in the planner, and any plan whose code-referencing blocks lack a `verified:` or
`unverified:` tag is rejected.

The rejection exists but is opt-in. `scripts/check-plan-provenance.sh` is one of the six
project-configured guards: it runs only where a project's `.flow/project.md` `## lint` declares
it, and it is exempt from the shipped-guards symlink rules in `check-guard-symlinks.sh`. A
project that never declares it never rejects an untagged plan — the mandate in the planner
template has no teeth outside this repository.

## What changes

- The plan-provenance guard ships with the flow skill: `check-plan-provenance.sh` and
  `check-plan-provenance.py` are symlinked into `skills/flow/scripts/`, matching the
  `check-plan-shape` pattern, and their project-configured exemption in `check-guard-symlinks.sh`
  rule 3 is removed.
- The writing-plans stage (`skills/flow/brainstorm-planner.md`, section D) runs
  `check-plan-provenance.sh` unconditionally, by basename, immediately after `check-plan-shape.sh`
  — in every project, before a plan is accepted.
- `skills/flow/SKILL.md`'s guard-presence union list gains `check-plan-provenance.sh`.

Untouched: the tag vocabulary (`skills/flow-contracts/plan-provenance.md`), the enforcement
contract (`skills/flow-contracts/plan-provenance-guard.md`), the scan scope (the three planning
files per non-archived change), build-green's project-configured status, and this repository's
own `## lint` entry for the guard.
