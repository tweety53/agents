# kan-1072-audit-whether-archived-spectre-change-artifacts

## Why

Every landed change leaves its whole spectre directory behind in `spectre/changes/archive/<name>/`:
plan text, narrative, verification reports, rendered ledger and panel records, self-review scratch,
and — the bulk of the bytes — visual-verification screenshots. Nothing reads most of it after the
change lands, yet it is committed in every repository flow has run in, and its bookkeeping commits
make up a large share of each repository's history. KAN-1072 asked whether these artifacts earn
their place; the operator's direction is to cut everything unused or rarely used from spectre.
`tasks.md`, enriched by writing-plans, is the artifact the run needs most, and it too is not
needed once integration is done.

## What changes

- **The audit**: `docs/audits/2026-10-31-spectre-archive-audit.md` records, per repository with a
  `spectre/` tree, the size and commit tables (each cell with its command), the consumer list, a
  verdict per artifact and per spectre surface, the replacement for every post-land reader, and the
  umbrella follow-up's content.
- **No archive any more**: integrate's run 1 no longer moves a change into
  `spectre/changes/archive/<name>/`. It lands one commit, `chore(spectre): drop <name>`, that
  deletes `spectre/changes/<name>/` (and each `<name>-fix-<N>` sub-change, and each satellite's
  `link.md` directory) from the tip. The self-review sub-change and the pipeline-fix loop drop
  their change directories the same way. Rendered ledger and panel records, self-review patches and
  `verbatim-moves.txt` are never committed past the run: they leave with the directory.
- **Readers follow**: every "live path, else `archive/<name>`" fallback reads the live path;
  readers that run after the drop (the self-review bundle, the merge-order read, the board's plan
  of a finished change, the record checkout check) read the change's files from history through one
  helper; the unlanded-archive scans are cut; cleanup's check reports a change directory still at
  the tip; the pipeline-fix key scan counts landed changes from history.
- **The existing archives go**: one commit per repository deletes `spectre/changes/archive/` from
  the tip in agents, gymie, gymie-frontend, gymie-playwright, flow-eval and spectre. History keeps
  every file; nothing is rewritten.
- **One umbrella follow-up Jira task** carries the audit's remaining recommendations (visual
  evidence out of git, the self-review report out of git, lessons and rejection search off the
  archive with decisions recorded in flowd, the spectre CLI cuts), filed only after the operator has
  seen the findings.
