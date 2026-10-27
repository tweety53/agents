## Context

One spectre change for the nine non-`big` findings of the self-review pass of gymie's
`kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer` (evidence:
`/Users/tweety53/Projects/gymie-worktrees/kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer/.superpowers/sdd/self-review-pass.md`).
Eight are pipeline defects fixed in this repository on `self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer`;
finding 1.6 is gymie dev tooling and finding 1.7's capture is gymie project configuration, both
committed in the gymie worktree
`/Users/tweety53/Projects/gymie-worktrees/kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer` on
`spectre/kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer`, whose change is archived — no
`spectre/` path there is written.

Two findings arrived untraced, and the plan opened on their traces:

- **8.1.** Every verifier report of the run
  (`verify-report-visual-verify-*.md` beside the evidence file) was read element by element. 1.1's
  invalid JSON hid nothing: no report fails to parse a matrix, and verifiers read the markdown
  matrix in `visual-verification.md`. The causes found:
  - the warm-up sheet's frame `templates-A4b`, mapped by the `templates.spec.ts.mockups` sidecar of
    a spec the change edits, was not on the frame list for rounds 1–6 — it joined only once Task 17
    moved its baseline (the list rule names cited frames and `moved:` frames only), and round 6's
    report equated it with `W4b`;
  - round 7 had A4b on the list but one `matrix:` line covering `templates-W1..W7b`, no A4b matrix,
    and reported only the sheet offset; round 8 built A4b's matrix and found the rest at once;
  - W4b's `colour 270` band delta was put down to decision `warmup-sublines-w4b` in rounds 1–4; it
    was the checkbox's size (round 5);
  - a stepper radius was read off the unlabelled side-by-side composite in rounds 5 and 6 and
    withdrawn in round 7 — identical pixels on the raw images.
- **1.7.** The full regression run (`playwright.config.ts`: 5 workers, `fullyParallel`, no
  `retries`, no `trace`/`video`/`screenshot` in `use`; gymie's `verify` row passes only `--output`)
  keeps no trace and no per-request timing, so the stall could not be traced from any run. The
  suspected mechanism was measured at plan time against the kan-1016 worktree's running frontend
  (port 3350, read-only `GET`s): the 49,708,247-byte wasm is served brotli-compressed
  (`Content-Encoding: br`, 9,560,018 bytes) in 0.26 s, and five parallel fetches take 0.36–0.40 s
  each — the dev server's wasm serving does not explain a `#app-loading` stall measured in tens of
  seconds. The defect this plan can fix is the capture gap.

## Decisions

### JSON string escaping

**ID:** pyjson-escape
**Status:** active
**Chosen:** `pyWrite`'s string case escapes `"`, `\` and every control character below U+0020 the way Python's `json.dumps` does (`\n`, `\r`, `\t`, `\b`, `\f`, else `\u00XX`), and writes every other rune raw, so every output that is valid today stays byte-identical; keys go through the same writer.
**Considered:** `ensure_ascii` `\uXXXX` escaping of non-ASCII — `json.dumps`' default, but it would change the bytes of every single-sided `unresolved` row (its `—`) that parses today, for no reader's benefit; joining the two errors with `; ` at `measurevisualproperties.go:1278` instead — closes this one caller and leaves the writer able to emit invalid JSON for the next string.

### Hand-built suite zip refused at close-task

**ID:** suite-zip-at-close-task
**Status:** active
**Chosen:** step 2 of `close-task` (`stats/internal/guard/closetask.go`), for each task commit (every map pair) that touches a path whose base name is `full-app-suite.zip`, runs `rebuild-full-app-suite-zip` in-process on that worktree with the zip's directory as `<screenshots root>`, compares the rebuilt bytes with `git show <sha>:<path>`, writes the committed bytes back to the file, and refuses (exit 1) on a difference with `close-task: task <id>: <path> is not rebuild-full-app-suite-zip.sh's output — rebuild it with that script and re-commit`. When the worktree's PNGs under that root differ from the commit (`git status --porcelain` non-empty there) it answers exit 2, cannot answer. No skill sentence is added: `scripts/close-task.sh`'s header is the contract and gains the step.
**Considered:** a structural check of the zip's headers (creator version 3.0, file-type bits) — a heuristic that admits any other non-script zip; checking at the verifier's own commit — the hand-built zip came through a task commit, which only `close-task` sees; a contract sentence in `skills/flow/verify-fix-loop.md` — it already says to use the script, and the gap was enforcement.

### Verifier rows skip the inline dirty-tree refusal

**ID:** verifier-skips-inline-dirty
**Status:** active
**Chosen:** the `inlineDirtyPaths` refusal in `flow record dispatch begin` runs only when `-role` is not `verifier`: a verify row records a run, not an edit, and `run-verify-list` already refuses to write a green sha over a dirty tree.
**Considered:** having `run-verify-list` pass `-started-at` — back-fills a row that is in fact beginning, falsifying its window; a new `-agent-id` value for verify rows — a second spelling of "inline" every reader would need to learn.

### Stage marks closed in the loop

**ID:** loop-closes-stage-mark
**Status:** active
**Chosen:** one sentence in step 1 of **The loop** (`skills/flow/verify-fix-loop.md`): a failing `flow.visual-verify` has its stage mark ended `-outcome stopped` beside its dispatch row, and a run re-entering the loop with `flow.write-in-progress` open ends that mark the same way first. `stopped` is the outcome the contracts already use for a stage that ends without completing (`skills/flow/integrate.md`, `skills/flow/review-panel.md`).
**Considered:** `-outcome blocked` — a dispatch-row outcome, not one any stage end uses; having `flow stage begin` refuse while another mark is open — turns every slip into a stop, where the store's supersede already keeps the record consistent.

### Trailer gate at the visual-verify stage end

**ID:** visual-verify-end-trailer-gate
**Status:** active
**Chosen:** `flow stage end -command '/flow' -stage flow.visual-verify -outcome completed` runs `check-untasked-commits` in-process over every worktree of the change's state `worktrees` map, each from `git merge-base HEAD origin/<base>` with the base integrate's unfinished-work gate uses for that worktree, and refuses (exit 1) on any listed commit, naming it and the amend (`git commit --amend` adding `-m "Flow-Key: flow.visual-verify"`). The checkout is resolved as `gatePanelFindingsClosed` resolves it with no `-C`. A guard that cannot answer refuses, as `gateLiveVerification` does. The gate exempts what integrate's gate exempts, so it refuses nothing integrate would admit — it moves the same refusal before the push.
**Considered:** `--not-remotes` against `origin/spectre/<name>` — `aa1ca18` was pushed before integrate saw it, and a commit already pushed would escape; checking only commits since the stage began — needs the stage's start time from the store, a second source for what git already answers; a push hook — the regression checkout is another repository this one does not install hooks into.

### Log rotation on start

**ID:** dev-log-rotate-on-start
**Status:** active
**Chosen:** a `rotateDevLog(file: File)` beside `devLogFile` in gymie `buildSrc/src/main/kotlin/com/gymie/gradle/DevLogs.kt` moves an existing non-empty `<service>.log` to `<service>.log.1` (replacing an older `.1`) before `startService` (`gradle/dev-lifecycle.gradle.kts`) opens the log; the redirect stays `appendTo`, now onto a fresh file. Every reader keeps working: `tailDevLogs` and the health-check messages name `<service>.log`, `admin-templates.spec.ts` reads its first MiB for the new start's `DB_URL`, and the docs' "wait for webpack's next compiled line in `frontend.log`" reads the running server's log.
**Considered:** `Redirect.to` (truncate) — loses the previous run's log a failed start is diagnosed from; size-capped rotation inside a running service — needs a logging change in every service for a growth bounded per run already; lowering the `local` profile's SQL DEBUG logging — the operator reads it.

### Trace capture for load-only failures

**ID:** verify-row-trace-on-failure
**Status:** active
**Chosen:** gymie's `.flow/project.md` `verify` row adds `--trace retain-on-failure` to its `npx playwright test` (a flag Playwright 1.62.1 lists among `--trace`'s choices), so every failing test of a verify run keeps `trace.zip` — its requests' timing and the boot's console — under the run's `test-results/<UTC timestamp>/`, where **The loop**'s rule to keep a failing run's result files already sends its fixer. The `capture` row is left unchanged.
**Considered:** `trace: 'retain-on-failure'` in `playwright.config.ts` — gymie-playwright is another repository, and the config would also trace every developer's ad-hoc run; `--trace on` — passing tests' traces cost disk on every 99-spec run, and a passing boot is not the one to trace; copying traces out in `run-verify-list`'s `keepResults` — the fixer reads them mid-run, before cleanup removes the worktree; compressing or caching the wasm — measured above, its serving is not the stall.

### One matrix file per frame

**ID:** matrix-one-file
**Status:** active
**Chosen:** step 10 of `skills/flow/visual-verify-verifier.md` names the one path, `<changeRoot>/visual-verification/matrix/<frame id>.json`, overwritten every round.
**Considered:** naming it by composite stem — frame ids are what the report's `matrix:` lines and the sidecar use.

### Warm-up departures: four verifier rules and one guard

**ID:** verifier-frame-list-and-matrix-lines
**Status:** active
**Chosen:** four sentences in `skills/flow/visual-verify-verifier.md` — the frame list carries every frame an edited capture spec's sidecar maps and never equates two ids; a recorded decision is a named cause only for the deltas it names; the composite's panel order is stated and radius, stroke, glyph size and colour are read off the raw images only; every composed frame carries its own `matrix:` line — and `check-verify-report.sh` refuses a `done` step 10 while a frame with a `- <frame id>: <path> diff=` line has no `- <frame id> matrix:` line.
**Considered:** captions drawn into the composite's gutters by `compose-mockup-frames` — changes every composite's bytes and size, where the rule only needs the raw images read; a frame list built from every sidecar in the regression checkout — verifies frames no change touched; prose without the guard — a single range line read exactly like per-frame coverage in round 7.

### Archived evidence left as committed

**ID:** archived-evidence-unrepaired
**Status:** active
**Chosen:** the 27 invalid matrix JSON files and their duplicate copies under gymie's `spectre/changes/archive/kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer/visual-verification/matrix/` are not rewritten: the dispatcher's constraint forbids `spectre/` edits in that archived change, and no reader parses them. The fixed writer serves every later run.
**Considered:** rewriting them in place — the constraint above.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — four Go gates on paths every run takes (close-task, dispatch begin, a stage end, the verify-report guard) plus a gradle start-path change, where a wrong refusal stops runs.
**Considered:** `flow-low` — the doc tasks alone would qualify, the gates do not.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** three groups, in series — `self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer-fix-g1`: Tasks 1, 2, 3, 4 (agents Go gates); `-fix-g2`: Tasks 5, 6, 7 (agents verifier and loop contract, and the verify-report guard); `-fix-g3`: Tasks 8, 9 (the gymie worktree). Each on `flow-low`/`opus`.
**Considered:** one fixer for all nine — two repositories and a Gradle build in one context; one per task — nine dispatches where three file-coherent groups serve.

## Open questions

none

## Recommended approach

Fix each defect where it lives: the JSON writer (`pyjson-escape`), close-task's commit step
(`suite-zip-at-close-task`), the dispatch-begin admission (`verifier-skips-inline-dirty`), the
stage-end gates (`visual-verify-end-trailer-gate`), the loop and verifier contracts
(`loop-closes-stage-mark`, `matrix-one-file`, `verifier-frame-list-and-matrix-lines`) and the
verify-report guard, and in gymie the dev-stack start (`dev-log-rotate-on-start`) and the `verify`
row (`verify-row-trace-on-failure`). Every rejected alternative is under its decision.

## Critical files

- `stats/internal/guard/pyjson.go` — the JSON writer.
- `stats/internal/guard/measurevisualproperties.go` — the two-sided `unresolved` row (read, unchanged).
- `stats/internal/guard/closetask.go`, `scripts/close-task.sh` — the close sequence and its header contract.
- `stats/internal/guard/rebuildfullappsuitezip.go` — the rebuild the check reruns (read, unchanged).
- `stats/cmd/flow/record.go` — `flow record dispatch begin`'s inline refusal.
- `stats/cmd/flow/stage.go` — `runStageEnd` and its gates.
- `stats/internal/guard/verifyreport.go`, `scripts/check-verify-report.sh` — the verify-report guard and its header contract.
- `skills/flow/verify-fix-loop.md` — **The loop**.
- `skills/flow/visual-verify-verifier.md` — the verifier contract.
- `spectre/changes/self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer/verbatim-moves.txt` — the six acknowledged new sentences.
- gymie `buildSrc/src/main/kotlin/com/gymie/gradle/DevLogs.kt`, `gradle/dev-lifecycle.gradle.kts` — the log file and the service start.
- gymie `.flow/project.md` — the `verify` row.

## Existing code to reuse

- `Registry["rebuild-full-app-suite-zip"]` and the fixtures of `stats/internal/guard/rebuild_full_app_suite_zip_test.go`.
- `ctRecordedWorktrees` / `flow state get`'s `worktrees` map (`stats/internal/guard/closetask.go`, `stats/cmd/flow/state.go`).
- `Registry["check-untasked-commits"]`, and the base resolution integrate's unfinished-work gate uses (`stats/internal/guard/integratechange.go`).
- `gatePanelFindingsClosed`'s checkout resolution and `gateLiveVerification`'s shape (`stats/cmd/flow/stage.go`).
- `devLogFile` (`buildSrc/src/main/kotlin/com/gymie/gradle/DevLogs.kt`).

## Verification

- `cd stats && make -s web-dist && go test ./internal/guard/ ./cmd/flow/ -count=1` and `scripts/run-guard-tests.sh` green; each new test fails with its fix reverted.
- `scripts/check-verbatim-moves.sh` passes with the six sentences acknowledged.
- `measure-visual-properties.sh` on a pair whose element resolves in neither image prints a document `python3 -m json.tool` accepts.
- In the gymie worktree: `./gradlew -p buildSrc test --tests 'com.gymie.gradle.DevLogsTest'`, then `./gradlew ktlintFormat` and `./gradlew verifyChange`.

## Live check

None — no daemon, store schema or scheduler changes: the Go changes are client-side admissions and guards exercised by their tests, the doc changes have no runtime, and the gymie log rotation runs only on a dev-stack start, which this loop must not perform on the running stack; its unit test drives the rotation on real files.
