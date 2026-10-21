# self-review-backlog-flow-1021

## Why

Five backlog findings from earlier self-review passes, planned together as one self-review fix
loop:

- **KAN-910** — `check-panel-fix-single-dispatch` read non-panel fix rows recorded under role
  `panel-fix` as invented panel-fix keys. The inline `full-suite-fix-<n>` case was fixed in
  ca90e129 and the gated fix round's role was stated as `implementer` in 6b2020bf, but no test
  pins which fix shapes the guard accepts, and the panel key rule names only the full-suite
  exemption.
- **KAN-911** — the `--base` feature has three unhandled edges: a recorded base deleted on origin
  still resolves (and a landing push would recreate it) with no documented recovery;
  `/flow-self-review` refuses every checkout but the default branch, so a `--base` change's
  context bundle, landed on its base, cannot be consumed; and run 2 refreshes each main checkout
  onto `origin/HEAD`'s branch, so a main checkout on the change's base reports `REFRESH-REFUSED`.
- **KAN-912** — a repeated `flow record dispatch begin` of the same key silently keeps the first
  begin's slot, model and effort.
- **KAN-915** — a panel re-run slot can answer without writing its report file, and
  `check-panel-findings-closed` still counts that re-run as the clean verification of a `fixed`
  finding.
- **KAN-906** — cross-guard helpers live in files named for one guard (`plainChangeName` in
  `cleanupcomplete.go`, `pcAbs` in `panelreproducers.go`, `capture`/`envGit` in
  `resolveremotebase.go`, …), so a reader cannot find them by name.

## What changes

- A table test pins every fix-dispatch shape `check-panel-fix-single-dispatch` accepts or
  refuses, and the panel key rule and the guard's header state the gated and verify-fix roles.
- `resolve-base-branch.sh` refuses a recorded base origin no longer carries, by name; the `--base`
  paragraph of `skills/flow/SKILL.md` states the recovery. `/flow-self-review` accepts a `--base`
  change's base as its landing branch and lists the origin branches holding a bundle it cannot
  find. Run 2 refreshes a main checkout that has the change's `--base` branch checked out onto
  that branch.
- A repeated dispatch begin updates the row's slot, model and effort to the new begin's values.
- `check-panel-findings-closed` counts a re-run as clean only once the slot's
  `panel-report-<round>-<id>.md` is written, and names a re-run that answered without it.
- Cross-guard helpers move, unchanged, into `paths.go`, `exec.go`, `flowcli.go`, `text.go`,
  `changename.go` and `planparse.go` in `stats/internal/guard/`.
