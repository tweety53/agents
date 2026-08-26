# kan-271-move-panel-findings-out-of-markdown-and-into

## Why

`check-unfinished-work.sh`'s findings signal and all of `check-panel-reproducers.sh` re-derive facts
— finding count, status, reproducer shape — by grep/awk over `docs/superpowers/reviews/*-panel.md`,
a Markdown *rendering* of rows KAN-258 already put in the `findings` table. The guards' own comments
catalogue what that buys: NUL-byte binary-mode misreads, a CRLF checksum bug that silently skipped
the check, a fenced-marker impersonation attack, and a digit-run bound — six review passes' worth of
scar tissue, all consequences of parsing prose instead of querying rows. Found during KAN-266's
self-review.

## What changes

- `myflow record finding` / `myflow record status` reject a malformed status (not `open`, `fixed`, or
  `withdrawn <reason>`) or an empty reproducer at write time — the store can no longer hold what the
  guards used to have to detect and reject on read.
- A new read-only CLI verb, `myflow record findings -change <name>`, prints a change's findings as a
  JSON array.
- `check-unfinished-work.sh` and `check-panel-reproducers.sh` query that verb instead of parsing
  `docs/superpowers/reviews/*-panel.md`. The checksum/marker-span/table-vs-marker-agreement/
  duplicate-identifier machinery — all of it there to reconcile two surfaces that could disagree —
  is deleted, because there is only one surface now.
- `scripts/lib/panel-record.sh` is deleted; nothing sources it once both guards read JSON.
- `docs/superpowers/reviews/*.md` keeps being rendered from the store for human readers —
  unaffected; only the guards' dependence on parsing it goes away.
