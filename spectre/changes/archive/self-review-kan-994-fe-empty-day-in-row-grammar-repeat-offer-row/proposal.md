# self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row

## Why

The self-review pass of `kan-994-fe-empty-day-in-row-grammar-repeat-offer-row` found seven fixable
defects in the pipeline and in gymie's dev tooling:

- `flow record dispatch begin` accepted a review-panel reviewer row with no `-slot`, and the missing
  slot blinded `check-panel-findings-closed.sh` (F-A).
- A fix run's review re-used round `1` although rounds 0–3 were already recorded, and no pass row
  was written for it: signal five of `check-unfinished-work` is keyed by round, so it passed (F-B).
- Neither the fix run nor the second integrate run appended its section to `narrative.md`, and
  nothing checks that a run did (F-C).
- `flow record mutation -path` stored a path no worktree holds (F-D).
- A spec failing in an earlier `verify-report-*.md` — `full-app-suite.spec.ts:110`
  register-from-invite, dismissed as a flake — was hidden by a later green run and cost a fix run
  (F-E).
- gymie-playwright: the per-capture re-baseline loop was done by hand twice, with zsh's unsplit
  `$SPECS` pitfall (F-F).
- gymie: the Compose-vs-CSS line-box lesson was paid twice (kan-933, kan-994) and lives only in
  narratives and four private code copies (F-G).

## What changes

- `flow record dispatch begin` refuses a `panel-` reviewer key without `-slot`, and one whose round
  is below the highest round the change's records carry.
- `flow stage sessions` counts the sessions that marked a stage for a change;
  `check-narrative-runs.sh` compares them with `narrative.md`'s fix-run and integrate-run sections,
  and both `flow.write-in-progress` and `integrate-change.sh commit` run it.
- `flow record mutation` refuses a `-path` that exists in no worktree of the change.
- `check-unfinished-work` reports a spec a verify report of the change failed when no later commit
  touched its capture or spec and the narrative cites no fixing commit.
- gymie-playwright gains `scripts/rebaseline.mjs`; gymie's `.flow/project.md` names it, and gymie
  gains the brief `docs/briefs/compose-text-on-a-css-line-box.md`.
