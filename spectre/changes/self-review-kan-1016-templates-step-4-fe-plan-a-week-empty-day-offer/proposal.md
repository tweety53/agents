# self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer

## Why

The self-review pass of gymie's `kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer` found
nine non-`big` defects, eight in the pipeline and one in gymie's dev tooling:

- 1.1 `measure-visual-properties.sh --props matrix` writes invalid JSON: a pair resolved in neither
  image gets `errors.Join(errA, errB).Error()` (`stats/internal/guard/measurevisualproperties.go`),
  whose raw newline `pyWrite`'s string case (`stats/internal/guard/pyjson.go`) writes unescaped.
  All 27 matrix files the run committed fail `json.load`.
- 1.2 Task 20 committed a hand-built `full-app-suite.zip` (Python `zipfile.writestr` entries:
  creator 2.0, mode 0600) instead of `rebuild-full-app-suite-zip.sh`'s output, and nothing refused
  it, so the next scripted rebuild over byte-identical PNGs commits different bytes.
- 1.3 `run-verify-list.sh` begins its row `-role verifier -agent-id inline`, and
  `flow record dispatch begin` refuses every inline begin over tracked modifications
  (`inlineDirtyPaths`, `stats/cmd/flow/record.go`) — a guard for editing work that a verify row,
  which edits nothing, trips whenever `visual-verification.md` is modified.
- 1.4 The verify-fix loop closes the failing verifier dispatch row but never the
  `flow.visual-verify` stage mark, so every round's next `flow stage begin` closes it
  `superseded`: attempts 1–9 and 11 of this run, and `flow.write-in-progress` 1 when the fix run
  re-entered the loop.
- 1.5 Playwright `aa1ca18`, a commit made inside `flow.visual-verify`, lacked its
  `Flow-Key: flow.visual-verify` trailer and was caught only by integrate's unfinished-work gate,
  after the branch was pushed.
- 1.6 gymie's dev stack opens `.dev-stack/logs/<service>.log` with `Redirect.appendTo` on every
  start (`gradle/dev-lifecycle.gradle.kts`) and never rotates it: `backend.log` reached
  1,106,597,487 bytes and crashed `admin-templates.spec.ts` past V8's string limit.
- 1.7 Load-only Playwright boot timeouts recurred in the full 99-spec run, and the regression run
  kept no trace, no video and no per-request timing, so no fixer could trace them.
- 2.1 No contract names the file a verifier saves a frame's matrix to, and the run committed each
  frame's matrix three times under three names.
- 8.1 The shared warm-up sheet's departures surfaced one round at a time across rounds 6–9, and a
  composite misread raised a false departure in two rounds.

## What changes

- `pyWrite` JSON-escapes every string it writes; a two-sided `unresolved` matrix row parses.
- `close-task.sh` refuses a task commit whose `full-app-suite.zip` differs from what
  `rebuild-full-app-suite-zip.sh` builds from the same PNGs.
- `flow record dispatch begin` admits an inline `-role verifier` row over a dirty tree; every other
  inline role is still refused.
- **The loop** (`skills/flow/verify-fix-loop.md`) closes a failing `flow.visual-verify` stage mark,
  and an open `flow.write-in-progress` mark on re-entry, `-outcome stopped`.
- `flow stage end -stage flow.visual-verify -outcome completed` refuses while a recorded worktree
  of the change carries a commit outside every task without its `Flow-Key:` trailer.
- gymie's dev stack moves an existing `<service>.log` to `<service>.log.1` on each start and opens a
  fresh log.
- gymie's `verify` row runs Playwright with `--trace retain-on-failure`, so every failing test keeps
  a trace with its requests' timing.
- The verifier saves one matrix file per frame, `<changeRoot>/visual-verification/matrix/<frame id>.json`.
- The verifier's frame list carries every frame an edited capture spec's sidecar maps; a recorded
  decision explains only the deltas it names; radius, stroke, glyph size and colour are never read
  off the composite; and every composed frame carries its own `matrix:` line, which
  `check-verify-report.sh` enforces.
