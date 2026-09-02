# kan-379-resolve-project-md-keys-through-flow-scripts

## Why

KAN-379. kan-378 deleted restated prose and left every `/flow` run still loading three contracts
at run time to read one value each: `project-configuration.md` (45252 bytes) at four points of a
creating run, `workspace-isolation.md` (25115 bytes) at two — one of them only to compute a
workspace id `flow workspace-id <name>` already derives — and `plan-provenance.md` (25221 bytes)
by the planner, of which 5822 bytes are what a tagger applies. The `.flow/project.md` section
extractor those loads replace already exists three times in `scripts/`, unshared. `pipeline.md`
carries a change-name resolution procedure that is a mechanical union `flow` can compute, and a
stage-marks section that mostly describes what `stage.go` and `check-stage-mark-calls.sh` already
enforce. `SKILL.md`'s model-resolution block says the project's `## planning model` and
`## self review model` keys win, and never reads them.
<!-- measured: wc -c skills/flow-contracts/{project-configuration,workspace-isolation,plan-provenance}.md @ ed2fbf8 -->

## What changes

- `scripts/project-get.sh <project-root> <key>` prints one `## <key>` body of `.flow/project.md`,
  over a shared `scripts/lib/project-section.sh` that `gather-dispatch-context.sh` and
  `check-model-keys.sh` source instead of their own copies. Shipped through `skills/flow/scripts/`,
  tested by `scripts/test-project-get.sh`.
- `review-panel.md`, `verify-and-handoff.md`, `integrate.md` and `archive.md` read their key
  through it and no longer load `project-configuration.md`; `implement.md` runs
  `flow workspace-id <name>` and no longer loads `workspace-isolation.md`;
  `verify-and-handoff.md` loads `workspace-isolation.md` only when `prepare-workspace.sh` exits
  non-zero or names a `cache index` row.
- `plan-provenance.md` keeps the run-applied half; the guard-facing half moves verbatim to
  `plan-provenance-guard.md`.
- `flow state resolve` computes the candidate set `pipeline.md`'s **Change name resolution**
  describes; that section and **Stage marks** keep their normative sentences and move their
  explanation to `pipeline-rationale.md`. `/flow-status` enumerates through the new command.
- `flow settings models` prints `ValidModels`; the `## Model resolution` block reads both project
  keys through `project-get.sh`, validates through it, and `check-model-resolution-shell.sh`
  proves each case per key.
- Before/after load-set figures recorded in `design.md`, measured to the leaf with kan-378's
  script.
