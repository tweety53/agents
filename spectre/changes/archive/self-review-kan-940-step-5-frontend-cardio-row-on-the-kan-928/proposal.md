# self-review-kan-940-step-5-frontend-cardio-row-on-the-kan-928

## Why

The self-review pass over kan-940 found eleven non-`big` defects, each held against the tree
(`design.md`, **Context**):

- A landing that changes `stats/` leaves the live flowd and the installed `flow` CLI on the old
  build: nothing restarts them, so kan-940's inline closes were refused by a floor the tree had
  already lowered, and two rows stayed open for 13 hours.
  <!-- measured: the pass's reading of kan-940's dispatch rows (machine-local store, not re-runnable) -->
- `close-task.sh` reports a failed dispatch end only as its first warning line and exits 0; any
  output filter drops it.
- The inline verify's **Recording** paragraph names no `-change` or `-session-token`, both of
  which `flow record dispatch begin` requires.
- `flow stage begin … <name> -C <wt>` exits 2 with "expected exactly one argument": a flag after
  the change name is read as a second positional.
- `flow record dispatch begin` silently records `-effort default` when the flag is omitted, and
  two `implement.md` examples omit it.
- `flow record finding -location` accepts prose where the panel contract requires `file:line`.
- The self-review bundle is 56% noise: visual-verification matrix rows (108.7 KB of 231 KB) and
  motion PNGs in every `git log --stat`.
  <!-- measured: awk over kan-940's archived visual-verification.md (108679 bytes) and the pass's bundle size @ 04a3a161 -->
- `-tokens` on an inline row's end bypasses the window floor and overwrites harvested cost even
  under Claude Code, whose transcript the harvester already credits.
- The unfinished-work gate's printed close command carries no `-ended-at`, so closing a stale row
  stamps "now" as its end.
- `-commit` accepts an abbreviated sha.
- Integrate's `VISUAL-VERIFY-MISSING` re-run is always unscoped, though the guard already knows
  which of the change's paths moved since the newest verified HEAD (21 minutes, byte-identical
  captures).
  <!-- measured: the pass's reading of kan-940's verifier dispatch rows (machine-local store, not re-runnable) -->

## What changes

- `pipeline-fix.sh land` and `refresh-main-checkout.sh` run `make -C <main>/stats restart` when the
  fast-forward they made moved a `stats/` path, printing `FLOWD-RESTARTED:` or
  `FLOWD-RESTART-FAILED:`.
- `close-task.sh` repeats a failed dispatch end as its last stderr line, `close-task: UNRECORDED —`
  with the command to run, and exits 3 when everything else closed.
- The inline verify's begin/end pair is spelled in full in `verify-and-handoff.md`.
- `flow stage` and `flow tasks` take flags after the change name, as `flow state` already does.
- `flow record dispatch begin` requires `-effort`; the `implement.md` examples carry it.
- `flow record finding` refuses a `-location` with no `:<line>`.
- The bundle drops matrix table rows under `####` headings and every `*.png` stat line.
- `flow record dispatch end`/`backfill` refuse `-tokens` when `CLAUDE_CODE_SESSION_ID` is set.
- The gate's close command carries `-ended-at <when its work ended>`, and `-commit` for an
  implementer row; `verify-and-handoff.md` says what those are.
- `-commit` must be a full 40-hex sha; `close-task.sh` expands an abbreviated `-end-commit` first.
- A `VISUAL-VERIFY-MISSING` verdict after a verified HEAD names that HEAD's changed paths, and the
  re-run is scoped to that HEAD as a round base.
