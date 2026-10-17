# pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-17

## Why

`skills/flow/sync-onto-base.md`'s **Conflict** bullet makes every binary-file conflict a stop and
ask. For a Playwright snapshot PNG under a `*-snapshots/` directory, or the `full-app-suite.zip`
the visual stage builds, the operator ruled (2026-10-17) that the answer is always the same and must
not be asked: a targeted recapture against a build carrying both sides. kan-940's playwright land
conflicted on three exercise-sheet PNGs and `full-app-suite.zip` against upstream `eb0dbb68`
(KAN-990's YouTube-link re-baseline), and the run stopped and asked.

## What changes

- `sync-onto-base.sh`'s guard prints `RECAPTURE: <path>` after its `CONFLICT:` line for every
  unmerged PNG under a `*-snapshots/` directory and for `full-app-suite.zip`; the exit stays 1.
- The **Conflict** bullet resolves each such path by a targeted recapture — a throwaway local app
  build merging its own base's tip, the owning specs alone run through `capture` with
  `--update-snapshots`, the zip rebuilt, `git add`, `rebase --continue` — never a stop.
- Every other binary conflict keeps stop and ask.
