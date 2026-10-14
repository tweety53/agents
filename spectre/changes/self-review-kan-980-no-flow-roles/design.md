## Context

- Three non-`big` findings from the kan-980 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task. All three were verified against the tree at `ff835bde`:
  - `runRecordPass` (`stats/cmd/flow/record.go`) checks only that `-note` is non-empty;
    `api.ApplyPassRecord` likewise. The renderer's `markerLabels` (`stats/internal/records/render.go`)
    is the one list of marker labels, and `neutraliseMarkers` rewrites each label's colon in free text.
  - `skills/flow/review-panel-late-fix.md` defines `<since-close-sha>` as the sha the last clean close
    reviewed, read from the pass log or panel record, and has no rule for a branch whose shas
    integrate's `commit` (**3. Commit the staged work**, `skills/flow/integrate.md`) has rewritten.
    `check-late-fix-trigger.sh` takes each since-close sha as an argument and derives none, so it needs
    no change.
  - `skills/flow/review-panel-fix-round.md` describes the `-cite` send-back and its outcomes but not
    that the round waits for it; `check-rerun-citations.sh` is run only by the parent.
- The new run-loaded sentences are acknowledged in this change's `verbatim-moves.txt`, spiked at plan time
  with both prose tasks applied: `check-verbatim-moves.sh`, `check-references.sh`,
  `check-installed-citations.sh`, `check-markdown-integrity.py` and `check-normative-inventory.sh` all
  exit 0. Each task copies its prose verbatim from `tasks.md`, or the acknowledgement no longer matches.
- Task 1's Go was spiked at `ff835bde`: the new test fails without the refusal (exit 0, a journal
  written) and passes with it; `go vet` and `gofmt -l` clean; every `TestRecordPass*` and the
  `internal/records` tests pass.
- No finding was dropped.

## Decisions

### The refusal lives in the CLI, before the store is contacted

**ID:** pass-note-refusal-in-cli
**Status:** active
**Chosen:** `runRecordPass` refuses, exit 2, a `-note` whose whitespace-trimmed text opens with
`<label>:` for any label in the renderer's list, exported as `records.MarkerLabels`; the message names
`flow record mutation` for a fix-mutation line and `flow record finding` for a finding line. Exit 2 is the
command's existing caller-mistake code (`TestRecordPassAndMutationMissingFlagsExitTwo`), and refusing
before the store is contacted keeps the note out of the offline journal too.
**Considered:** refusing in `api.ApplyPassRecord` — a store-side refusal of a journaled write surfaces
only at replay, after the parent has moved on, and the API error would have to name CLI commands;
a second copy of the label list in the CLI — two lists drift; matching the label anywhere in the note —
a pass note may legitimately mention a marker in prose, which the renderer already neutralises.

### After a reshape the since-close sha is the archive commit

**ID:** since-close-after-reshape
**Status:** active
**Chosen:** one sentence in `skills/flow/review-panel-late-fix.md`: after integrate's reshape the
since-close sha is the canonical worktree's archive commit and a peer worktree's reshaped branch head —
the reshaped commits carry exactly what the clean close reviewed. `check-late-fix-trigger.sh` is
unchanged: it takes the sha as given.
**Considered:** teaching the guard to derive the sha — it derives no sha today and the caller already
supplies one; a peer's archive commit — a peer carries none, only the two-commit chain.

### A pending citation send-back holds the stage close

**ID:** cite-sendback-holds-close
**Status:** active
**Chosen:** one sentence in `skills/flow/review-panel-fix-round.md`, after the send-back's exit codes:
the round neither closes nor hands to `flow.verify` until the `-cite` dispatch has returned and
`check-rerun-citations.sh` has been run on its report.
**Considered:** a guard at the stage close reading the dispatch rows for an open `-cite` row — not
trivially cheap: no close-time guard reads dispatch rows by key suffix today.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `low` — one small refusal in a record write with its test, and two single-sentence prose
additions; nothing on a gate's verdict path.
**Considered:** `medium` — the diff carries no concurrency, data-model or verdict change for a
low-effort read to miss.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, tasks 1–3, one fixer in plan order — the tasks are file-disjoint and small, and two
are single-sentence docs edits not worth a dispatch of their own.
**Considered:** one fixer for the Go task and one for the docs — two dispatches for work one fixer
finishes inside its context.

## Open questions

## Live check

None — the change adds a CLI-side refusal before any store contact and two prose sentences; the Go test
exercises the refusal end to end through `run`, and no daemon or store behaviour changes.
