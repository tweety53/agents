# self-review-kan-1009-editor-replace-exercise-remove-from-superset

## Why

The self-review of `kan-1009-editor-replace-exercise-remove-from-superset` (gymie) found six
pipeline defects, none `big`:

1. `flow record dispatch begin` accepted `-key verify-` — a repository suffix that expanded to
   nothing — and five verify rows of the run were keyed `verify-`, `verify-fix-1-` … `verify-fix-4-`,
   which `check-verify-green.sh` then grouped as one worktree.
2. `rebuild-full-app-suite-zip.sh` resolved a relative `<screenshots root>` against the caller's
   working directory, so gymie's `screenshots` value `.` wrote `full-app-suite.zip` into the gymie
   main checkout.
3. `check-frame-list.sh` declared `R1`/`R2` from a `**Considered:**` line that named a rejected
   alternative, and never expanded `LM-R1 … LM-R10`, so six of the ten frames went unchecked.
4. Every subagent row of the run from its 17th dispatch on reads "not measured": the parent typed
   invented `-agent-id` labels (`task-11`, `visual-verify`, `panel-2-extra`), the CLI stored them,
   and the harvester never stamps a row whose agent id is already set.
5. The first visual-verify dispatch handed back with step 10 undone after 22m34s and needed a
   9m34s remainder: its prompt's REPORT FORMAT paragraph presents a rejected report as merely
   costing "a remainder dispatch".
6. gymie-playwright's `scripts/specs-for-diff.mjs` mapped the run's frontend diff to all 94 specs,
   because its caller walk follows comments, misses `suspend fun` and generic-receiver
   declarations, follows a caller file's every declaration, and treats dead code as an
   unmappable branch.

## What changes

- `flow record dispatch begin` refuses a `-key` ending in `-`; a new `flowd rekey-dispatch`
  re-keys the five rows the defect wrote.
- `flow record dispatch begin`/`end` refuse an `-agent-id` that is not `inline`, `none` or a
  harness id shape; `flowd restamp-dispatch` stamps the run's 21 mis-recorded subagent rows onto
  their transcripts.
- `rebuild-full-app-suite-zip.sh` refuses a relative screenshots root with exit 2.
- `check-frame-list.sh` expands `<prefix><a> … <prefix><b>` ranges and ignores `**Considered:**`
  lines.
- The visual verifier's REPORT FORMAT paragraph tells it to run the report guard and finish the
  steps it names before handing back.
- `specs-for-diff.mjs` (gymie-playwright, on the change's own branch) strips comments, reads
  modifier and generic-receiver declarations, follows only the caller declarations that use the
  callee, and ends a dead-code branch with no specs.
