## Context

In-run pipeline fix raised by `/flow` change `kan-1047-review-dynamics-heatmap-by-project`; the
fix branch is based on `origin/main` at `edf8e4e8`.

**The defect (blast radius 3 files).** <!-- measured: git diff --stat of the Task 1 diff @ edf8e4e8 → 3 files -->
The kan-1046 self-review fix (1d) made `flow record mutation` refuse a proof for round r once a
reviewer row of panel round r has begun, so a proof can never contradict what that re-review saw.
The close guards accept a round-N finding's proof only at round N or N+1. Together they leave no
legal row once the re-review began without one: a deadlock with no recovery but editing the store.

## Decisions

### Late form on the CLI

**ID:** late-mutation-flag
**Status:** active
**Chosen:** a `-late '<reason>'` flag on `flow record mutation` that skips the ordering refusal; the refusal message names it.
**Considered:** dropping the refusal — loses the kan-1046 guarantee for every run that records in order; letting the guards accept round N+2 rows — any later row would then prove any earlier round, and the record would not say the proof came late.

### The reason rides a pass note

**ID:** late-reason-pass-note
**Status:** active
**Chosen:** after the mutation row is written, `-late` writes one pass row at the same round, note `late mutation proof for <path>: <reason>` — an existing row kind the panel render already prints in that round's pass log, so no schema change.
**Considered:** a new `late` column on the mutation row — a store migration, API field and render change for a one-line annotation; carrying the reason in `-test` — that field is the test that failed, and the guards and render read it as such.

### Guards unchanged

**ID:** guards-unchanged
**Status:** active
**Chosen:** `uwUnprovedFixRounds` and its two callers keep counting round N/N+1 rows; a late row at N+1 counts as any row does.
**Considered:** making the guards require a late note — nothing to enforce: the CLI writes the note itself.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one flag, one skipped branch and one extra write, pinned by an extended test.
**Considered:** `flow-medium` — the CLI is used by every run, but the default path is byte-for-byte unchanged.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `pipeline-fix-kan-1047-review-dynamics-heatmap-by-project-40-fix`, Task 1.
**Considered:** none — the plan has one task.

## Open questions
