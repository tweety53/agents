# pipeline-fix-kan-954-user-row-loaded-twice-per-request-10

## Context

`RecordSelfReviewFinding` (`stats/internal/store/selfreviewfindings.go`) inserts every call as a
new row, and `self_review_findings` carries no uniqueness constraint — 0035's header says so on
purpose: "a pass interrupted after its store writes and re-run records its findings again". The
gymie run of `kan-954-user-row-loaded-twice-per-request` shows the cost: rows 365/373 repeat
`fixed a4a32afc` with the same note; rows 381/384/387, 382/385/388 and 383/386/389 repeat
`14a900ef`, `697ad6ae` and `b0d0cdd5`, the first of each written by `integrate-change.sh land`
(the `fixed:` line's text as note, no blast radius) and the next two by hand (a shorter note,
blast radius 1 or 2). So a repeat's note and blast radius differ from the first write's — only
the ref identifies a `fixed` row. A refless (`declined`, `noted`) row has nothing but its note.
<!-- measured: flow self-review findings -change kan-954-user-row-loaded-twice-per-request -C /Users/tweety53/Projects/gymie (read-only) @ 2026-10-16 -->

Callers are unaffected: the handler (`stats/internal/api/selfreview.go`) still answers `201` with
the store's row, and the CLI and client only relay it. `skills/flow-self-review/SKILL.md` step 6's
caller-side rule (record a sha only when it is not listed yet) stays; the store now holds even
when a caller misses it, or a timed-out write is retried after it committed.

Reproduced at plan time at 752f3a28: with the task's test hunks alone,
`TestSelfReviewFindingRecordedOnce` fails (`a repeat answered with a new row: ids [1 2 3 4 5 6 7]`)
and `TestSelfReviewFindingsMigrationCollapsesRepeats` fails (`index "self_review_findings_once"
does not exist`); with the whole diff, `go test ./internal/store/ ./internal/api/
./internal/client/ -count=1` is green.

Blast radius — the defective file, its test, every file naming `RecordSelfReviewFinding` or
`self_review_findings` outside `.git`, `node_modules` and the spectre archive, plus the new
migration — not `big`: 15 files.
<!-- measured: grep -rl -e RecordSelfReviewFinding -e self_review_findings --exclude-dir=.git --exclude-dir=archive --exclude-dir=node_modules . | wc -l → 14 @ 752f3a28, +1 new migration -->

## Decisions

### A finding's key is its ref, or its note when it has none

**ID:** finding-key
**Status:** active
**Chosen:** unique on (project_key, change, angle, disposition, `COALESCE(ref, md5(note))`) — a
`fixed` row is its sha, a `filed` row its issue key, a refless row its note; one expression index
covers all four dispositions. `md5(note)` keeps a long note inside the btree index's row-size
limit.
**Considered:** the note in the key for every row — the gymie repeats carry a different note from
the first write, so they would not collapse; `filed` keyed by note as well — two findings filed
under one issue key and one angle then stay two rows, but a repeat with a reworded note would
not collapse, and one issue key is one piece of tracked work; a raw `note` column in the index —
an insert of a note past ~2.7 kB would fail.

### A repeat is absorbed at the store

**ID:** upsert-at-store
**Status:** active
**Chosen:** `INSERT … ON CONFLICT (<key>) DO UPDATE SET blast_radius =
COALESCE(self_review_findings.blast_radius, EXCLUDED.blast_radius) RETURNING id, note,
blast_radius, recorded_at` — one statement, the caller gets the recorded row back, and the
landing's blast-radius-less row picks up the count a later hand write carries.
**Considered:** `DO NOTHING` — returns no row, so a second `SELECT` would be needed to answer;
refusing the repeat with a conflict error — every caller (the landing, the parent, a retry)
would need to treat it as success; fixing only the callers — the store write after a client
timeout is not a caller decision.

### Existing repeats collapse onto their first row

**ID:** migration-cleanup
**Status:** active
**Chosen:** the same migration first fills a null blast radius from the first non-null among the
row's repeats, then deletes every repeat but the lowest id, then creates the index — the index
cannot be created while repeats remain.
**Considered:** deleting without the fill — the gymie `14a900ef`/`697ad6ae`/`b0d0cdd5` rows would
keep the landing's null blast radius and lose the hand-recorded counts; leaving old rows and
indexing only new ones — a partial index cannot express "rows after id N" and the doubled counts
would stay in the stats view.

### First review on low

**ID:** first-review-effort
**Status:** active
**Chosen:** the first review, `pipeline-fix-10-review-1`, runs on `flow-low` (effort `low`) on
`opus`: one migration, one statement change and two tests, verified at plan time both ways.
**Considered:** `medium` — the data-model seam is one expression index whose behaviour both tests
pin.

### One fixer group

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Task 1 in one `pipeline-fix-10-fix` dispatch.
**Considered:** none other — the plan has one task.

## Open questions

## Live check

After landing and a `flowd` restart (it runs migrations on start),
`flow self-review findings -change kan-954-user-row-loaded-twice-per-request -C
/Users/tweety53/Projects/gymie` lists each of `a4a32afc`, `14a900ef`, `697ad6ae` and `b0d0cdd5`
once — before: 2, 3, 3, 3 rows — and the `14a900ef`/`697ad6ae`/`b0d0cdd5` rows carry blast radius
1, 2, 2. Failure looks like the same repeats still listed, or `flowd` refusing to start on
`apply migration 0038_self_review_findings_once.sql`. The check only reads; it records nothing.
