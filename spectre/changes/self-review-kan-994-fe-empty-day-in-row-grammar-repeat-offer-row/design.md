## Context

Found by the self-review pass of `/flow` change `kan-994-fe-empty-day-in-row-grammar-repeat-offer-row`
(a gymie change, canonical worktree `/Users/tweety53/Projects/gymie-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row`).
The agents fixes ride `self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row`, based at
`origin/main` (`bc701950`), in this worktree. F-F and F-G are project dev tooling outside the agents
repository: per **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`) they commit on
the change's own branch `spectre/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row` — F-F in the
gymie-playwright worktree `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row`
(its gymie note in the gymie worktree), F-G in the gymie worktree. Each finding is carried by exactly
one task: Task 1 F-A, Task 2 F-B, Task 3 F-C, Task 4 F-D, Task 5 F-E, Task 6 F-F, Task 7 F-G. None was
disproved; none is `big`.

Facts measured at plan time:

- **F-A.** `TestRecordDispatchBeginRefusesANonNumericPanelRound` passes at `bc701950` and asserts that
  `-role reviewer -key panel-2-primary` with no `-slot` exits 0 — the gap, pinned as accepted.
  `runRecordDispatchBegin` (`stats/cmd/flow/record.go`) validates the key's shape only. kan-994's
  store holds two reviewer rows with no slot, `panel-1-primary+principles` and `panel-1-primary`;
  the second is F-B's re-used round, written after rounds 2 and 3.
- **F-B.** `runRecordDispatchBegin` makes one store call, the write; nothing reads the change's
  recorded rounds. `flow record next-round` (`runRecordNextRound`) already computes the highest
  round over pass, finding and mutation rows.
- **F-C.** flowd's `GET /api/v1/stage-runs?project_key=gymie-7c1f238a&name=kan-994-…&command=/flow&stage=<key>`
  answers 1 session (`0cdee296…`) for `flow.document-fix` and 2 (`12a273c9…`, `c3d515fc…`) for
  `flow.preserve-sessions`; kan-994's `narrative.md` carries no `— fix run` heading and one
  `— integrate run` heading. No `flow` verb reads stage runs today, and every guard reads the store
  through the `flow` CLI. `flow.preserve-sessions` is marked by `integrate-change.sh prepare` exactly
  where it prints `NEXT: append this run's narrative`; `flow.document-fix` only on a fix run
  (`skills/flow/document-fix.md`). A session stopped at the unfinished-work gate marks
  `flow.preflight` and `flow.unfinished-work-gate` but not `flow.preserve-sessions`.
  `scripts/check-verbatim-moves.sh`, run with Task 3's two prose paragraphs applied, prints exactly
  the four FAIL sentences this change's `verbatim-moves.txt` acknowledges.
- **F-D.** kan-994's mutation row names `src/commonMain/kotlin/com/gymie/frontend/ui/home/RepeatCandidatesHolder.kt`;
  the file is `shared/src/commonMain/kotlin/com/gymie/shared/ui/RepeatCandidatesHolder.kt` in the
  gymie-frontend worktree. `flow state get` for kan-994 answers a `worktrees` map naming all three
  worktrees.
- **F-E.** kan-994's `verify-report-visual-verify.md` (mtime 13:52 +0300) failed
  `fidelity/full-app-suite.spec.ts:110 register-from-invite.png`. In the playwright worktree's reflog,
  no commit between that report and the first integrate (`d5497df`, `f5c8946`, `4b1a9db`, `a1a7333`,
  `b0613b2`, then the squash `6d61b82`) touched `register-from-invite-darwin.png` or
  `fidelity/full-app-suite.spec.ts`; `cc0193b` (19:46, the fix run) touched both. Every capture
  `verify-report-visual-verify-fix-1-full.md` (17:06) failed — c6-empty-day-offer, sheet-7-no-match,
  nav-drawer, nav-menu-open, cardio-add-workout-activities — was touched by `a1a7333` (17:29). So
  the signal designed below flags register-from-invite at the first integrate and nothing else.
- **F-F.** `require('playwright-core/lib/utilsBundle').PNG.sync.read` (an exported subpath of the
  installed `playwright-core` 1.62.1) decodes a committed baseline; the register-from-invite pixel
  diff between `fbef27b` and HEAD has the bounding box x 460, y 424, 282x22 — the widened mask.
  `npx playwright test --help` lists `--update-snapshots [mode]` with the `changed` choice.
- **F-G.** `LineHeightStyle(Alignment.Center, Trim.None)` is declared privately in
  `shared/src/commonMain/kotlin/com/gymie/shared/ui/cardio/CardioPlanParts.kt:142` (`CSS_LINE_BOX`),
  `shared/src/commonMain/kotlin/com/gymie/shared/ui/components/FormSheet.kt:90`
  (`FORM_TITLE_LINE_HEIGHT_STYLE`) and `shared/src/commonMain/kotlin/com/gymie/shared/ui/exercise/ExerciseInfoSheet.kt:67`
  (`FRAME_LINE_BOX`, `internal`), and inline in `shared/src/commonMain/kotlin/com/gymie/shared/ui/HomeScreen.kt:1478`;
  `Alignment.Bottom` in `BroadsheetType.kt:147` and `CardioPlanParts.kt:148`, `Alignment.Top` in
  `EditorBar.kt:118`. gymie has no `docs/briefs/` directory yet.

## Decisions

### A panel reviewer row names its slot

**ID:** panel-reviewer-slot-required
**Status:** active
**Chosen:** `flow record dispatch begin` exits 2 before contacting the store when `-role reviewer`, `-key` starts `panel-` and `-slot` is empty, naming `-slot <slot|slot+slot+slot>` from `review-panel.md`. `dispatch backfill` inherits it, since it runs `begin`.
**Considered:** deriving the slot from the key's tail — a second source of the slot that a renamed bundle would silently disagree with; refusing in the store — the CLI is where every other dispatch-key refusal lives, and a store refusal would journal nothing yet surface later.

### A panel round never goes below the recorded ones

**ID:** panel-round-floor
**Status:** active
**Chosen:** for a `-role reviewer` key `panel-<n>-…`, `dispatch begin` reads the change's run record and exits 2, writing nothing, when `<n>` is below the highest round any pass, finding or mutation row or any panel reviewer dispatch row carries — naming that round and `flow record next-round`. `<n>` equal to it is admitted (a round's other slots, a re-dispatched slot). A read that fails — unreachable store, not-found — skips the check: the begin then journals or writes exactly as before.
**Considered:** requiring `<n>` to equal `next-round`'s answer — refuses every slot after a round's first pass row; refusing on a failed read — blocks recording during a store outage, which the journal fallback exists to survive; checking only pass rows — kan-994's fix run wrote none, the very gap.

### Narrative runs are counted from stage marks

**ID:** narrative-runs-from-stage-marks
**Status:** active
**Chosen:** a fix run is a session that marked `flow.document-fix`, an integrate run a session that marked `flow.preserve-sessions`, both under command `/flow`; a session is its `sessionId`, and a row with none counts as one session of its own. `flow stage sessions -stage <key> [-C dir] <change>` prints that count for one stage as one integer, reading `GET /api/v1/stage-runs` filtered by the project key `-C` resolves, the change name, `/flow` and the stage; a failed read exits 1 with nothing on stdout. `check-narrative-runs.sh <worktree> <change> <narrative.md>` counts the file's `## … — fix run` and `## … — integrate run` headings (an absent file holds none) and exits 1 when either count is below its session count, 2 when it cannot answer, 0 otherwise.
**Considered:** `flow.preflight` for integrate runs — a session stopped at the unfinished-work gate, or a run 2 cleanup, marks it and owes no section; `flow.write-in-progress` for fix runs — kan-994's fix run never reached it, the run the check exists to catch; reading flowd over HTTP from the script — duplicates the address, timeout and daemon-header handling the `flow` CLI owns.

### The narrative check runs at both appends

**ID:** narrative-check-wiring
**Status:** active
**Chosen:** `flow.write-in-progress` (`skills/flow/verify-and-handoff.md`) runs the check right after its append; `integrate-change.sh commit` runs it after the in-run fix line check, over the change's live or archived `narrative.md`, and stops `STOP: narrative` on exit 1 or 2, relaying the check's line. `skills/flow/integrate.md` names the new stop cause; `skills/flow/scripts/` gains the guard's symlink.
**Considered:** integrate only — catches the gap one run late, after the fix run's session is gone; the unfinished-work gate — it runs before integrate's own append, so the integrate run's own section could never be there yet.

### A mutation path names a file a worktree of the change holds

**ID:** mutation-path-exists
**Status:** active
**Chosen:** `flow record mutation` exits 2, writing nothing, unless `-path` names an existing file: an absolute path as given; a relative one under the git top level of `-C`'s checkout or under any worktree the change's state record (`flow state get`'s `worktrees` map) names. A state read that fails leaves the `-C` checkout alone as the candidate. The refusal names every root searched.
**Considered:** requiring a repo-qualified path — every existing row and the contract line `fix-mutation: <path> — …` use worktree-relative paths, so a new spelling would split the record; checking only `-C`'s checkout — refuses a cross-repo fix's mutation, kan-994's own case.

### A failing verify-report spec must be addressed before landing

**ID:** verify-failure-signal
**Status:** active
**Chosen:** `check-unfinished-work` signal nine, read where signal three is (the worktree holding the plan) whenever a range exists, archived plan or not: every `<worktree>/.superpowers/sdd/verify-report-*.md` is parsed for its `- verify: exit <n>` item with `<n>` not 0, and each `<spec>.spec.ts:<line>` in that item's indented continuation lines is a failure, its capture the first `<name>.png` after it and before the next `.spec.ts:`. A failure is addressed when a commit in any range, committed (`%ct`) at or after the report's mtime, touches a path whose base name is `<name>-<anything>.png` or `<name>.png` — or, for a failure naming no capture, the spec file by base name — or when the change's `narrative.md` holds a line `Verify failure addressed: <spec>:<line> — <sha>` whose sha is a commit in one of the ranges. Each unaddressed failure is one reason, `verify failure not addressed: <spec>:<line> [<name>.png] (<report>)`, with the course: fix it in this run, or cite the fixing commit with that narrative line.
**Considered:** listing failures as an advisory only — kan-994's was dismissed once already; matching a touch of the spec file for every failure — a spec's other tests change it constantly; ignoring commit time — the creating run's re-baseline of a capture, made before the report that failed it, would clear the failure; skipping archived plans as signal seven does — kan-994's second integrate ran on an archived plan and still lacked the fix. Accepted cost: a run 1 reshape gives its squash a fresh date, so a capture touched before the report and squashed after it reads as addressed; the first integrate's gate, before any reshape, still sees the original commits.

### Re-baselining is one script over a spec list

**ID:** rebaseline-script
**Status:** active
**Chosen:** `gymie-playwright/scripts/rebaseline.mjs <spec>...` runs `npx playwright test <spec>... --update-snapshots=changed` with the caller's environment, then prints, for each PNG `git status` reports modified or new, `<path> — <x>,<y> <w>x<h>, <n> px` (the bounding box of differing pixels against `HEAD`'s blob), `<path> — size <w>x<h> → <w>x<h>`, or `<path> — new`; it exits with Playwright's status. Specs arrive as arguments, so no shell variable is split. PNGs decode through `playwright-core/lib/utilsBundle`'s `PNG`, already installed; `diffBox(a, b)` is exported and tested with `node --test`.
**Considered:** a zsh or bash loop — the `${=SPECS}` pitfall is the shell's; adding `pngjs` — a dependency the installed Playwright already bundles; restoring unwanted PNGs automatically — the box is the operator's evidence, and a capture outside the moved region is a judgment, not a rule.

### The gymie note is its own commit

**ID:** rebaseline-note-own-commit
**Status:** active
**Chosen:** Task 6 commits the script in gymie-playwright under its `**Commit:**` subject and the `.flow/project.md` note in the gymie worktree as `docs(flow): name the re-baseline script under visual verification` — a commit's scope names its module, and the two repositories' modules differ.
**Considered:** a separate task for the note — two tasks for one finding.

### The line-box lesson is a brief, not a refactor

**ID:** line-box-brief
**Status:** active
**Chosen:** `docs/briefs/compose-text-on-a-css-line-box.md` in gymie, the lessons home (**Process lessons**, `skills/flow-contracts/lessons.md`), stating the rule, the measured exceptions, where the copies live and how to verify — and pointing a new caller at the `internal` `FRAME_LINE_BOX` rather than a fifth copy.
**Considered:** consolidating the copies into one theme token now — touches four files of another repository's shipped UI with re-baselines, outside a self-review fix's finding.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — `dispatch begin` sits on every run's recording path and gains a store read, and the unfinished-work gate gains a signal that can block landing; a slip in either blocks or silently passes every later run.
**Considered:** `flow-low` — each task is small and test-pinned, but two of them change what the pipeline's own gates admit.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** three fixer dispatches in series: `fix-g1` Tasks 1, 2, 4 — all in `stats/cmd/flow/record.go` and `record_test.go`, in that order; `fix-g2` Tasks 3, 5 — the narrative guard and the unfinished-work signal, shell harnesses and `stats/internal/guard`; `fix-g3` Tasks 6, 7 — gymie-playwright and gymie, a Node script and prose.
**Considered:** one fixer for all seven — one context carrying the Go suite, the shell harnesses and two other repositories; one per task — seven dispatches, three of them editing the same two files in series.

## Open questions

## Live check

Every live step reads; none writes the store. From this worktree, with `stats/` built (`cd stats && make -s web-dist`):

- `go run ./cmd/flow stage sessions -C /Users/tweety53/Projects/gymie-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row -stage flow.document-fix kan-994-fe-empty-day-in-row-grammar-repeat-offer-row` prints `1`, and `-stage flow.preserve-sessions` prints `2` (the figures measured above).
- `scripts/check-narrative-runs.sh /Users/tweety53/Projects/gymie-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row kan-994-fe-empty-day-in-row-grammar-repeat-offer-row /Users/tweety53/Projects/gymie-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row/spectre/changes/archive/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row/narrative.md` exits 1 naming 1 fix run against 0 sections and 2 integrate runs against 1 — unless the self-review run has since appended them, in which case it exits 0 and the counts it prints match the file.
- `go run ./cmd/flow record dispatch begin -C /Users/tweety53/Projects/gymie-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row -change kan-994-fe-empty-day-in-row-grammar-repeat-offer-row -role reviewer -slot primary -key panel-1-primary -model opus -effort default -agent-id none -session-token mf-live-check-panel-round-floor` exits 2 naming round 3 — the highest kan-994's records carry (`flow record next-round` answers `4` at plan time) — and `flow record next-round`. Run it only after Task 2's test passes; if it exits 0, it wrote a row — close it at once with `flow record dispatch end … -key panel-1-primary -session-token mf-live-check-panel-round-floor -outcome aborted` and treat the check as failed.
- `scripts/check-unfinished-work.sh` over the kan-994 gymie worktree, with the gymie-frontend and gymie-playwright worktrees' ranges from its state record, prints no `verify failure not addressed` reason: every failure its three reports list was touched after its report (`cc0193b`, `fb32fac`, `a1a7333`). Any such reason there is a false positive — the signal not working.

Not working looks like: `stage sessions` printing `0` for kan-994 (a filter that matches nothing), the narrative check exiting 0 over the unchanged narrative, the dispatch begin writing a row, or the gate naming a kan-994 capture.
