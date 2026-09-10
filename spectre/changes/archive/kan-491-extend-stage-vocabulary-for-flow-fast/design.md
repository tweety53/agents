## Context

A `flow stage begin` mark is validated by `stages.Validate(command, key)` against the vocabulary
`names.go` transcribes from README.md's Level 1 table; `names_test.go` re-derives the table from
README.md at test time and fails on any disagreement, so the two cannot be edited separately. The
Commands column parser splits the cell on commas and trims backticks, preserving cell order, and
the comparison is `reflect.DeepEqual` — content and order both. `Command` is an exported type whose
value set today is the single constant `Flow = "/flow"`.

`skills/flow-fast/SKILL.md`'s **Stage keys** table is the authoritative statement of which keys
flow-fast marks: five phase-file rows naming 26 distinct keys. The vocabulary extension must match
that table exactly — no more, no fewer.

## Decisions

### Both halves move in one commit, verified by the existing agreement test

**ID:** one-commit-vocabulary-move
**Status:** active
**Chosen:** README's Commands cells and `stages.Table`'s Commands fields change in the same task,
ordered red-then-green: the new drift guard fails against the unextended vocabulary, the extension
turns it green, and `TestStagesMatchReadmeLevelOne` holds the two halves identical throughout.
**Considered:** splitting the README and Go edits into two tasks — rejected, each intermediate
state breaks `TestStagesMatchReadmeLevelOne`, so a split manufactures a red build for no review
benefit.

### The drift guard reads the SKILL.md table, in the external test package

**ID:** skill-table-drift-guard
**Status:** active
**Chosen:** a new test in `names_test.go` parses `skills/flow-fast/SKILL.md`'s Stage keys rows,
derives the key set, and asserts it equals exactly the keys whose Commands carry `/flow-fast` in
`stages.Table` — compared as sets, because the SKILL table is grouped by phase file while the
vocabulary is ordered by README row.
**Considered:** asserting against `byCommand` directly — impossible without moving the test into
package `stages`, which the file's existing `stages_test` package declaration forbids; skipping
the guard — rejected, it is the same re-derivation pattern that already pins README to the code,
and its absence is precisely how this defect shipped.

### Deployment is an operator step, not part of this change

**ID:** cli-install-operator-step
**Status:** active
**Chosen:** the change touches only the vocabulary and its tests; the installed
`~/.local/bin/flow` picks the fix up when the operator next runs `make restart`, which by design
also restarts the protected dev daemon.
**Considered:** an agent-run install of the rebuilt binary — rejected, replacing the live CLI
underneath running sessions is exactly what the `## stop` section's operator-only rule exists to
prevent.

## Open questions
