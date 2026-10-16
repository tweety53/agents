# pipeline-fix-kan-954-user-row-loaded-twice-per-request-10

## Why

In the gymie `/flow` run of `kan-954-user-row-loaded-twice-per-request`,
`flow self-review findings -change kan-954-user-row-loaded-twice-per-request` lists the `fixed`
row for `a4a32afc` twice and each of `14a900ef`, `697ad6ae` and `b0d0cdd5` three times:
`integrate-change.sh land`'s step-5 recording of the report's `fixed:` lines wrote them, then the
parent recorded them again by hand with `flow self-review finding` — one of those writes after a
"store unavailable … context deadline exceeded" warning whose row had nevertheless been
committed. `RecordSelfReviewFinding` (`stats/internal/store/selfreviewfindings.go`) is a plain
`INSERT`, and migration 0035 deliberately left the table without a uniqueness constraint. A
repeat of the same finding is never meaningful: `skills/flow-self-review/SKILL.md` step 6 records
one row per sha not yet listed, and a rewritten sha is a new ref.

## What changes

- A new migration, `0038_self_review_findings_once.sql`, collapses existing repeats onto their
  first row and adds a unique index on (project, change, angle, disposition, ref — or the note's
  md5 when the row carries no ref).
- `RecordSelfReviewFinding` writes through `ON CONFLICT` on that key: a repeat stores no second
  row and answers with the recorded one, its missing blast radius filled from the repeat.
- Two store tests: one for the write, one for the migration's cleanup.
