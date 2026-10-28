# self-review-kan-1041-weekly-dashboard-model-effort-correlation

## Why

The self-review pass of `kan-1041-weekly-dashboard-model-effort-correlation` found seven fixable
pipeline defects:

- `hooks/protect-main-checkout.py` denies a report write of two `cat > <file> <<'E'` heredocs in
  one Bash call: `strip_data_heredocs` judges each body's rest with the other body still in it, so
  a lone `.` reads as `source` and an unpaired backtick leaves the rest unparseable, and both
  bodies stay scanned. A review slot's reports were lost to it, and the archived narrative blamed
  the throwaway copy.
- Panel-fix rows were keyed one round low (`panel-fix-5`, `panel-fix-7` for fixes of rounds 5 and
  7), which also defeated `panelFixFloor`: `flow record dispatch` refuses only `panel-fix-0`.
- An inline back-fill's derived window can contain another inline row of the same run
  (`task-13-implementer` held all of `verify-fix-3`), so the main thread's records belong to two
  rows and the late one reads "not measured".
- `skills/flow/review-panel.md` offers `PIPESTATUS`, which is empty in zsh, the shell this
  repository's Bash tool runs.
- `skills/flow/document-fix.md` names the plan guards with no invocation, and implies the
  provenance guard ships among the skill's scripts.
- `measure-visual-properties --props matrix` pairs 0 elements on any dark-theme image: the
  surface walk stops on the bottom row and the surface top lands at the image height.
- The live check against a copy of the dev store is hand-built in about 15 calls every run.

## What changes

- `strip_data_heredocs` judges a body's rest with every other heredoc body that runs nothing
  removed, opener and delimiter lines kept; the archived narrative's stated cause is corrected.
- `flow record dispatch begin` and `backfill` refuse `panel-fix-<r>` once a round-`r` reviewer
  began before the fix's end, naming `panel-fix-<r+1>`; the two mis-keyed store rows are re-keyed
  and re-timed with the existing `flowd rekey-dispatch` / `retime-dispatch`, and their pass-log
  note corrected.
- An inline back-fill's derived start is floored at the latest end among the change's other
  inline rows not after its commit; `task-13-implementer` is re-timed, and the archived ledger
  and panel record are re-rendered from the repaired store.
- `review-panel.md` drops `PIPESTATUS`; `document-fix.md` states `check-plan-shape.sh
  <changeRoot>/tasks.md` and names the project's configured provenance guard.
- `measureElements` takes the whole image as the surface when the walk moved no row.
- A new `scripts/live-store-copy.sh up|down`, with a stubbed harness, named in `.flow/project.md`'s
  `## run`.
