# kan-444 — keep plan provenance annotations mandatory — design

Approved design from the brainstorming dialogue of 2026-09-08. `design.md` in the change root is
derived from this document; this file is the brainstorming record.

## Problem

KAN-423's plan tagged every snippet and every numeric claim, and its implementation landed
first-pass with a single Important panel finding. KAN-444 (self-review finding, KAN-423,
flow-improvement angle) asks that the outcome stay reproducible: the provenance convention stays
mandatory in the planner, and any plan whose code-referencing blocks lack a `verified:` or
`unverified:` tag is rejected.

The rejection exists — `scripts/check-plan-provenance.sh` — but is opt-in: one of the six
project-configured guards, run only where a project's `.flow/project.md` `## lint` declares it.

## Approaches considered

- **Ship the guard (chosen)** — the enforcement already exists and rejects exactly what KAN-444
  names; only the wiring is missing.
- **Template wording only (rejected)** — the mandate already has mandatory wording; without the
  guard wired in, declaring projects are the only ones that reject.
- **Bolt the scan into `check-plan-shape.sh` (rejected)** — couples two guards with independent
  exit-code contracts; provenance hits would surface under the wrong guard's name.

## Design

1. **Behavior.** The writing-plans stage (`skills/flow/brainstorm-planner.md` section D) runs
   `check-plan-provenance.sh` by basename — resolving to `<skill-dir>/scripts/` — immediately
   after `check-plan-shape.sh`, unconditionally, in every project, and fixes any hit before the
   plan is accepted.
2. **Wiring.** Symlink `scripts/check-plan-provenance.sh` and `check-plan-provenance.py` into
   `skills/flow/scripts/` (relative `../../../scripts/…` targets, the `check-plan-shape`
   pattern; both files, because the wrapper execs `$SCRIPT_DIR/check-plan-provenance.py`).
   Remove `EXEMPT["check-plan-provenance.sh"] = 1` from `check-guard-symlinks.sh` rule 3 and
   update its comment — the project-configured family shrinks. Add `check-plan-provenance.sh` to
   `skills/flow/SKILL.md`'s guard-presence union list; the `.py` is covered by the
   sibling-dependency grep rule.
3. **Untouched.** Tag vocabulary, enforcement contract, scan scope, build-green's
   project-configured status, this repository's own `## lint` entry.

## Decisions

- `ship-provenance-guard` — ship and invoke, rather than restate the mandate; see the change's
  `design.md`.
- `enforce-at-writing-plans` — reject where plans are produced; no duplicate run at
  implementation entry.

## Testing

`scripts/run-guard-tests.sh` covers the guard's own harness and `check-guard-symlinks`'s; the
shipped invocation is exercised by this change's own plan (the writing-plans stage runs the guard
on its own `tasks.md`), and citation resolution via `check-references.sh` and
`check-installed-citations.sh`.
