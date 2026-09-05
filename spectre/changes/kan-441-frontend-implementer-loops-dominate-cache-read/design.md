## Context

The implementer's test commands come from two places it already has: the context bundle carries the
project's resolved `## test` list, and its own task block carries `**Tests:**`, the names of the
tests the task adds. Nothing today tells it which of the two to run when, so on a Gradle project it
alternates between `--tests '<class>'` and the whole `:shared:desktopTest` module at its own
discretion — and a whole-module run on gymie-frontend is 5–10 minutes and a log the context keeps.
<!-- measured: KAN-441 issue text, "waiting on a 5-10 minute Gradle run" — the self-review reading of KAN-423's transcripts, not re-run here -->

The dispatch preamble is a set of `>`-quoted paragraphs repeated verbatim at every dispatch site
and mechanically guarded by `scripts/check-dispatch-paragraphs.sh`, whose table
(`ENTRY_LABEL`/`ENTRY_SHARED_PHRASES`/`ENTRY_VARIANTS`, `SITE_*`) names each paragraph, its
load-bearing phrases and the files that must carry it. KAN-263 added `FOREGROUND BUILDS` to that
table after four implementers stalled on backgrounded builds; that paragraph and its row stay
exactly as they are.

The full `## test` list runs today in `flow.verify`, by the verifier, after the panel closes. A
regression from an early task therefore surfaces only after the panel — a fix run and a panel
re-run. Running it once at the end of `flow.sdd-tdd` moves that discovery before the panel without
adding a run to every task.

## Decisions

### Drop the background-and-monitor lever; cut test runs per task instead

**ID:** foreground-kept-cut-runs
**Status:** active
**Chosen:** keep `FOREGROUND BUILDS` and its guard row verbatim and reduce the number and weight
of test runs per implementer turn loop — one targeted run at RED, one at GREEN, a re-run only after
a source edit, output through `tail`.
**Considered:** the issue's `run_in_background` + Monitor lever — rejected: KAN-263's finding is
that a dispatched subagent ending its turn on a background command produces nothing, and a
background launch plus its completion is two turns where a foreground call under the Bash cap is
one; it would re-create the stall and add a turn per run.

### Targeted form derives from `**Tests:**` and the build tool, not a project key

**ID:** selector-from-tests-field
**Status:** active
**Chosen:** the paragraph tells the implementer to run the tests its task's `**Tests:**` field
names through the build tool's own selector — `--tests '<class>'` for Gradle, `-run` for `go
test`, `-t` for vitest — the plan already carries the names and the tool is visible in the tree.
**Considered:** a new optional `## test targeted` key in `.flow/project.md` carrying a template
with a `<class>` placeholder — rejected: configuration nobody asked for, one more key every project
must learn, and the implementer already derives the selector unaided today (half its runs are
targeted); what it lacks is the rule, not the syntax.

### The last bundle's implementer runs the full suite once per worktree

**ID:** full-suite-last-bundle
**Status:** active
**Chosen:** the conductor appends a `**FULL SUITE:**` paragraph to the last bundle's implementer
prompt only — it knows the last bundle from `plan-dispatch-bundles.sh` — telling it to run the
resolved `## test` list once, in the foreground, after GREEN and before its commit; fix a failure
inside its own task's `**Files:**`, report every other failure verbatim and unfixed. A report naming
a failure ends the conductor's turn with `## Question` before `final-review.diff` is written.
**Considered:** the conductor running the list itself after the last guard pass — rejected: the
same 5–10 minutes spent as a turn on the conductor's larger context, for the same failure path.
<!-- measured: KAN-441 issue text, "waiting on a 5-10 minute Gradle run" — the self-review reading of KAN-423's transcripts, not re-run here -->
Leaving it to `flow.verify` alone — rejected: a regression from an early task then surfaces after
the panel and costs a fix run plus a panel re-run, the quality loss the issue rules out.

### `TARGETED TESTS` is guarded at two files; `FULL SUITE` is not guarded

**ID:** targeted-guarded-full-suite-not
**Status:** active
**Chosen:** add a `targeted` row to `scripts/check-dispatch-paragraphs.sh` — three shared phrases,
no variants, `skills/flow/implement.md` min 1 block, `skills/flow/review-panel.md` min 1 block (the
panel-fix dispatch, which is implementer-shaped and re-runs tests per finding) — with fixture cases
in `scripts/test-check-dispatch-paragraphs.sh` mirroring KAN-263's. `FULL SUITE` lives at one site
and is conditional on the bundle, so it is ordinary prose, not a guarded paragraph.
**Considered:** guarding neither — rejected: an unguarded dispatch instruction is what KAN-289
burned on and KAN-263 fixed; the guard's table exists so a new paragraph costs one row. Guarding
`FULL SUITE` too — rejected: the guard counts unconditional blocks per file, and this one is
appended by the conductor to one dispatch, which the table cannot express without a variant it
does not need.

## Open questions

None.
