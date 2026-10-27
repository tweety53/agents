# self-review-kan-1040-copyable-next-command

## Why

The self-review pass of `kan-1040-copyable-next-command` found eight non-`big` pipeline defects:
the flow store's postgres container does not come back after Docker Desktop restarts; a reviewer
reproducer can destroy an uncommitted edit before `run-reproducer.sh` notices; the
`panel-fix-<round>` key is ambiguous and two runs recorded `panel-fix-0`; a journalled dispatch end
dates a reviewer's finish to the outage instead of its transcript; a stuck journal entry floods
the live daemon log while postgres logs nothing that would explain a stall; a re-run of
`panel-round-begins.sh` duplicates a pass-log note; the harvester settles an inline row against a
stale end and zeroes its tokens; and `flow suite record` loses its record when the store is down.
Several of them already wrote wrong rows into the dev store.

## What changes

- `flow-postgres` restarts with Docker (`restart: unless-stopped`), in compose and on the live container.
- `run-reproducer` snapshots the bytes of every tracked file that differs from `HEAD` before a run and restores any the run changed, still exiting 5.
- `panel-fix-<round>` counts fix rounds from 1; `flow record dispatch` refuses `panel-fix-0` and every other out-of-shape `panel-fix-` key, the panel-fix guard's key regexp agrees, and the contract says so.
- A journalled `dispatch end` is marked as such, and both the daemon handler and the journal replay close the row at its agent transcript's last write.
- The reconciler warns once per stuck journal entry instead of on every pass; `stats/README.md` names the live log; `flow-postgres` logs lock waits and slow statements.
- The store's pass-log write drops a note identical to one already recorded for the change and round; `flowd drop-pass` removes a duplicate row.
- The harvester's settle pass waits out a row end older than the candidate's own begin.
- `flow suite record` journals its row on store failure, and the reconciler replays it.
- Repairs: kan-1040 dispatch rows 8223–8225 (ends), 8228 and 8232 and kan-1036 row 7834 (keys, and 8228's start), 8235 (tokens), and pass row 2504 (duplicate).
