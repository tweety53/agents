# Design: findings out of Markdown, into store queries

## Problem

`check-unfinished-work.sh` signal two and all of `check-panel-reproducers.sh` re-derive facts
(finding count, status, reproducer shape) by grep/awk over the rendered Markdown panel record. The
rows already exist in the `findings` table (KAN-258); the guards were left reading a rendering of
those rows rather than the rows themselves.

## Decisions

### Write-time validation moves into Go

**ID:** write-time-validation
**Status:** active
**Chosen:** `runRecordFinding`/`runRecordStatus` (`stats/cmd/myflow/record.go`) validate `status`
(`open`, `fixed`, or `withdrawn <reason>` with a non-empty reason) and `reproducer` (non-empty on
`record finding`) before the store call, rejecting a bad write with exit 2 — the same "caller
mistake" contract `requireRecordFlags` already uses.
**Considered:** leaving validation read-time-only in the shell guards — rejected because it lets
invalid data sit in the store between writes, and the guard would have to keep re-deriving legality
on every read instead of trusting a store-level guarantee.

### A new read verb, not a JSON mode on `record render`

**ID:** new-findings-verb
**Status:** active
**Chosen:** `myflow record findings -change <name> [-C dir] [-addr] [-timeout]` — a new verb
alongside `dispatch|finding|status|render`, calling the same `client.GetRunRecord` `record render`
already uses, printing `run.Findings` as a JSON array to stdout. Exit 0 on success, including `[]`
for "no rows yet" (the same case `render` treats as MISSING). No journal: a failed read cannot be
answered later the way a failed write can be replayed, so store-unreachable is a non-zero exit with
a stderr message, never a fabricated empty result.
**Considered:** adding `-json` to `record render` — rejected to keep that verb's contract (write
Markdown files) unmixed with a caller reading data back out.

### Reproducer command-safety check stays in the shell guard

**ID:** reproducer-safety-in-shell
**Status:** active
**Chosen:** Go validates the reproducer field is *present* and *shaped* (`none — reason` or
non-empty command text); `check-panel-reproducers.sh` keeps checking the command text itself for
shell metacharacters, a leading `-` on the path token, an absolute path, or a `..` segment. This is a
security control on a string a subagent authored, checked immediately before it becomes a candidate
for `/myflow-do` to exec directly — it belongs at the point of trust for execution, not folded into
generic field-shape validation.
**Considered:** moving the whole check into Go at write time — rejected: it duplicates the same
security-relevant logic on the write path for no gain, since the guard must run it again anyway
right before treating a reproducer as executable, and one enforcement point close to the exec is
easier to audit than two.

## What the guard rewrites drop entirely

With one JSON array as the only surface, these no longer have anything to reconcile against:

- the `findings-total:`/`reproducers-total:` checksum lines (there is nothing a JSON array's own
  length can drift from)
- the "table rows vs marker block name the same findings" comparison (there is no second
  representation)
- the "marker lines must be one unbroken block" rule (nothing to quote-and-impersonate in a fenced
  example)
- NUL-byte/CRLF/locale handling in `scripts/lib/panel-record.sh` (JSON decoding replaces byte-level
  grep entirely)

`scripts/lib/panel-record.sh` is deleted once both guards stop sourcing it.

## Unchanged

`myflow record render -kind panel` still writes the human-readable Markdown to
`docs/superpowers/reviews/`. `spectre/changes/<name>/tasks.md` and its checkbox parsing (signal one
of `check-unfinished-work.sh`) are untouched — per KAN-258's explicit non-goal, plan completion stays
Markdown.

## Open questions

None.
