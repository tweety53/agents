## Context

- Ticket KAN-1060 (labels `flow`, `flow-cost`, `flow-visual-verify`; KAN-1063 merged into it). The cost
  evidence and the measured breakdown are `proposal.md`'s `## Why`; this section carries what the
  design rests on beyond them. The design was settled over three brainstorming rounds and approved by
  the operator on 2026-10-30.
- **Approved prototype — the spec for the Stage leaderboard's two new columns:**
  `.superpowers/sdd/2026-10-30-stage-leaderboard-main-share-prototype.html` in this worktree
  (`/Users/tweety53/Projects/agents-worktrees/kan-1060-flow-cost-visual-verify-now-costs-more-than/.superpowers/sdd/2026-10-30-stage-leaderboard-main-share-prototype.html`;
  no Artifact URL was recorded for it). It is rendered from the live board, its figures the real 30
  days to 2026-10-30, and the implementation and `flow.visual-verify` compare against it. It draws the
  board at rest; its inline script sorts a column on a header click, the first click ascending.
- **Why later passes find what earlier passes passed.** 50 fix rounds across 15 changes held 67
  round-opening defects (30 rounds opened by a `-full`-keyed run, 20 by a scoped run; sources: the
  verifiers' own reports in the subagent transcripts, `In-run fix` task titles, and `diff=` ratios
  unchanged to 4 decimals on an unchanged frame; 70 of 296 verifier reports could not be recovered):

  | class | defects | what it means |
  |---|---|---|
  | (b) earlier miss | 30 (45%) | the defect was already in a capture an earlier pass judged |
  | (d) flake or non-fidelity failure | 12 | e.g. pad-Enter and poster-wait flakes under suite load |
  | (a) regression | 11 | introduced or left behind by the previous round's fix |
  | (e) new evidence | 7 | a newly drawn state, a `moved:` frame, or the first pass that could capture |
  | (f) other | 6 | e.g. a fix that did not close its defect, a composite misread later withdrawn |
  | (c) scope gap | 1 | only the unscoped run looked at the frame |

  <!-- measured: verifier reports in subagent transcripts, In-run fix task titles and diff= ratios across 15 changes @ machine-local transcripts read 2026-10-30, not re-runnable at a ref -->

  Of the 30 earlier misses, 12 sat on a capture byte-identical to one an earlier pass had judged (10
  called clean or "match", 2 "pre-existing" and left); about 8 came from a skipped sweep or a skipped
  step 10; about 5 from a delta put down to a recorded decision or called "pre-existing" and not fixed
  (kan-1016: a stepper colour delta of 270 attributed to decision `warmup-sublines-w4b` from the first
  pass on); 5 were in motion strips. About 20 of the 30 are offsets, sizes or colours — numbers
  `measure-visual-properties.sh --props matrix` has written into
  `<changeRoot>/visual-verification/matrix/<frame id>.json` on every pass since 2026-10-26, while the
  verdict was prose nothing checked against them (kan-1016's W2 rows sat 3–6px high on captures
  byte-identical through fix-1 to fix-3, two passes calling the frame "match"; that its JSON held the
  delta is inferred from the script being deterministic, not re-run). A second, smaller cause: one
  report split into many rounds — kan-1003 rounds 6–13 came from two verifier reports, one task per
  defect, which `skills/flow/verify-fix-loop.md` step 1 forbids and nothing enforces. Regressions (11)
  and flakes (12) are not addressed here: flakes are fixed at their source by the existing determinism
  rule, and no pipeline cut applies to either.
- **The parent's by-hand frame/matrix check today** (`skills/flow/visual-verify.md`, the paragraph
  opening "The report's `frames:` line counts against the change's frame list") has the main session
  open each mockup beside the verifier's matrix and look for a declared frame with no line, a composed
  frame missing its `bands:`, `seams:` or `matrix:` line, a visible element with no row, an empty or
  wrongly n/a cell, and a sweep neither `done` nor n/a. Its first item became `check-frame-list.sh`
  after KAN-437 (22 frames declared, 15 checked); the rest is by hand. It has never caught anything
  traceable: in 45 days, 13 remainder runs have a readable trigger and `check-verify-report.sh` sent all
  13; across 383 passes in 30 days the main session never reported a gap found by eye and opened at
  most about one image per three passes; and the gaps it exists to catch got through — about 223 matrix
  rows with empty cells across 6 archived changes (kan-753 alone 211), and n/a cells written because a
  measuring command failed (kan-938's `visual-verification.md:100,169,228`, kan-1008's `:404`).
  <!-- measured: archived visual-verification.md records and remainder dispatch rows of the last 45 days @ gymie and agents archives read 2026-10-30 -->
- **What the parent's visual-verify turns are made of** is `proposal.md`'s breakdown: mechanics the
  scripts below absorb, and reports read whole into a 130k–460k context.
- **Constraints the design keeps.** The verifier stays on opus at `flow-low` (Done-when; **The
  verifier dispatch**, `skills/flow/visual-verify.md`). Steps 3, 5, 6, 12 and 13 stay in the parent's
  own Bash calls, never a subagent's (KAN-845, `skills/flow/SKILL-rationale.md`). Only the parent waits
  (`skills/flow-contracts/pipeline.md`). New script logic is Go in `flow-guard` behind a
  `flow_guard_exec` shim (`.flow/project.md` `## scripts`). Verifier dispatch rows carry no stage-run
  id, and no `flow` command reads whether a stage run is open.
- KAN-1062 and KAN-1064 measure their acceptance on the same leaderboard columns.

## Decisions

### A thin parent: one call before each verifier dispatch, one after

**ID:** thin-parent-prepare-close
**Status:** active
**Chosen:** one Go entry, `visual-verify-pass` (`stats/internal/guard/visualverifypass*.go`, shim
`scripts/visual-verify-pass.sh`), on `integrate-change`'s pattern — subcommands, one closing `NEXT:` or
`STOP:` line, exit 0 / 1 / 2. `prepare` runs, in the parent's own Bash call, every mechanical step
before a dispatch: the stage `begin`, steps 1–2's trigger and the skipped-at-Decide check, step 3's
`check-verify-green`, `prepare-workspace | check-visual-preflight` and `check-sidecar-reach`, step 5's
probe, step 6's fingerprint, the `## visual verification` row resolution, the mockup sources with the
bundled-page decode, the `-diff-base` HEADs, the verifier's resolved bundle at
`<abs-worktree>/.superpowers/sdd/verify-prompt-<key>.md` and its `flow record dispatch begin -role
verifier -model opus -effort low` (with `-round-base` on a scoped re-run). `close` runs
`check-verify-report.sh --design` and, with `mockups` declared, `check-frame-list.sh`, ends the row,
commits step 12's outputs, ends the stage and renders the ledger where its outcome allows, and prints
the next step. The parent keeps the first key of each stage, the judgement lines of the Agent prompt
(motions, views touched, wire contract, `moved:` frames, the `sweeps-<n>.md` path), the miss
classification and the tooling analyst, **Blocking**, environment cures, the in-run fix task, and the
launch and wait. `check-stage-mark-calls` holds the new call to the token and harness rules, as it
holds `integrate-change.sh`.
**Considered:** one subcommand — the Agent launch sits between the two halves and stays the parent's;
a run-wide orchestrator subagent for the stage — banned by **Only the parent waits**
(`skills/flow-contracts/pipeline.md`); moving the parent's guard runs into the verifier's prompt —
kan-983's verifier burned ~4.45M tokens on an environment that could not pass, and KAN-870's
self-reported completeness was accepted wrongly; a sonnet summariser after the verifier (the ticket's
merger idea) — one verifier writes one report and can write its own verdict, so a summariser adds a
dispatch and main turns and folds nothing.

### The verifier hands back its verdict alone

**ID:** verdict-only-handback
**Status:** active
**Chosen:** the verifier's turn ends with a `## Verdict` block — its `Model:` line (the handshake reads
it), an `exits:` line, the `steps:` line, `cause:` when a step is blocked, and `departures: <n>` with one
line each — and its report file carries that `## Verdict` with the full `## Report` under it. The
guards keep parsing `## Report` alone; `close` reads the `## Verdict`.
**Considered:** the full report handed back — 2k–43k characters each, staying in a main context
re-read on every later message.

### The parent starts, restarts and stops the stack

**ID:** parent-starts-the-stack
**Status:** active
**Chosen:** `prepare` probes each app URL and runs the fingerprint; when nothing answers it prints
`NEXT: start-stack`, and on the fingerprint's first exit 1 `NEXT: restart-stack`; the parent starts the
stack from `start` when declared, else `## run`, with its own Bash call, and runs
`prepare --continue --started`. A fingerprint exit 1 under `--started` ends the stage `stopped`. Step
13's stop stays the parent's, which knows it started the stack.
**Considered:** `prepare` running `start` itself, as round 1 drew it — gymie's `start` row carries
`<FRONTEND_ROOT>`-style placeholders the contract gives no mechanical substitution for, and with no
`start` row the stack is `## run`'s prose (this repository's `make ui-test-up`), so the script path
would run in neither known project.

### The parent passes the app URLs

**ID:** parent-passes-apps
**Status:** active
**Chosen:** `prepare --app <worktree>:<app-root>=<resolved-url>`, the argument
`check-visual-preflight.sh` takes today prefixed with its worktree, once per app the stage may probe;
`prepare` uses those under a surviving worktree for the pre-flight, the probe and the bundle.
**Considered:** `prepare` deriving app roots and URLs from `## apps` and `## workspace isolation` —
this repository's visual stack is the UI-test stack on 4174, which neither table names.

### A later call in the same stage says so

**ID:** prepare-continue-flag
**Status:** active
**Chosen:** the stage's first `prepare` runs the stage `begin` and steps 1–3 and writes the stage file
(decision `stage-file-names-its-rows`); `prepare --continue <key>` — after `NEXT: start-stack`,
`restart-stack` or `analyst`, and for a remainder or a re-dispatch — skips those and runs steps 5–6,
the bundle and the row begin.
**Considered:** reading whether `flow.visual-verify` is open from the store — no `flow` command reads
an open stage run, and the parent already knows; a second `flow stage begin` would supersede the open
run.

### The stage file names the stage's rows

**ID:** stage-file-names-its-rows
**Status:** active
**Chosen:** the first `prepare` writes `<canonical-worktree>/.superpowers/sdd/visual-verify-stage.txt`
— the session token and the stage's begin instant, the resolved set's size, each surviving worktree
with its repository name, merge base and round base, then one line per verifier row a `prepare` began,
key and worktree. `prepare --continue` and `close` read it; `close` ends the stage once every surviving
worktree's newest row is closed, `stopped` if any is `blocked`, else `completed`.
**Considered:** deriving the stage's rows from the store — verifier rows carry no stage-run id, and
the session token spans every round of the run; leaving every multi-verifier stage end to the parent —
gymie's keys always carry a repository suffix, so the saving would miss the project that spends most.

### The tooling analyst runs between the pre-flight and the verifier's row

**ID:** analyst-after-preflight
**Status:** active
**Chosen:** on a fix run with a miss the parent passes `--miss`; `prepare` stops after step 3 with
`NEXT: analyst`, the parent dispatches the analyst, then runs `prepare --continue`.
**Considered:** dispatching the analyst after a full `prepare` — the verifier row's window would span
the analyst's run.

### A remainder or re-dispatch names earlier reports by path

**ID:** redispatch-bundle-by-path
**Status:** active
**Chosen:** a `-rest` bundle carries `## Incomplete attempt` with the incomplete report's path and the
guard's verdict line; a `visual-verify-<n>` bundle carries `## Previous attempt` with every earlier
report's path. `prepare --continue` derives both from the key and the reports on disk.
**Considered:** pasting the reports verbatim, as today — the parent holds only each `## Verdict`, and
pasting would read the full report into the main context.

### `close` commits only after the stage's final report

**ID:** close-commits-final-reports
**Status:** active
**Chosen:** `close` runs step 12's pathspec commit and prints `ROUND-BASE:` on every outcome except
`rest`, `redispatch` and `cure-environment`.
**Considered:** committing after every report — those three dispatch again in the same stage against
the same captures, and the stage's step-12 commit follows its final report.

### The verifier runs no `verify`

**ID:** verifier-runs-no-verify
**Status:** active
**Chosen:** step 7 retires: `flow.verify`'s test group already ends with the visual `verify` row on the
specs `specs` maps (**Inline verify**, `skills/flow/verify-and-handoff.md`, since `a4f3e724`), and
`check-verify-green` refuses any verifier dispatch until it is green. The report's `verify:` line names
the green inline row the bundle names. Retired with it: `rerun specs:`, `scoped specs: <n>` and the
parent's own `specs` run under the round base, `check-verify-report.sh`'s third argument, step 7's
polled `verify`, the round base in the verifier's prompt (the parent still scopes `motions:` with it,
and `prepare` still records `-round-base`), and `check-visual-verify-dispatched`'s kan-1022
rerun-of-failed-specs rule. Step 8 still runs `specs` against the merge base to pick its capture spec.
**Considered:** keeping a scoped `verify` in the verifier — it re-runs the specs `flow.verify` just ran
green on the same build, and its 2–18 minute polled suite is the pass's longest wait and the cause of
most non-zero-exit re-dispatches (kan-1044's first pass, its remainder and fix rounds 1 and 3 all
blocked on it).

### The inline verify row records its HEAD

**ID:** verify-row-records-head
**Status:** active
**Chosen:** `run-verify-list.sh` passes `-diff-base <HEAD>` on its verify row's `begin`.
**Considered:** reading the HEAD from the row's timing — `dispatches.diff_base` is empty on every
`verify*` row, and a time window cannot tell which commit a run saw.

### A green verify counts while only planning paths changed since it

**ID:** green-verify-spans-planning-commits
**Status:** active
**Chosen:** `check-verify-green` counts a green row only when its `-diff-base` resolves in a worktree of
the resolved set and no commit of `<diff-base>..HEAD` there touches a path outside
`<spec root>/changes/` — the exclusion `check-unfinished-work`'s signal seven applies. A row with no
`-diff-base` is not green. It takes the set's other worktrees after its three arguments and judges each
row in the worktree its `-diff-base` resolves in. Its verdict line names the green rows and their HEADs.
**Considered:** a green row whose HEAD is the HEAD being dispatched (round 1) — the scoped pass's
step-12 commit moves HEAD on a project committing captures to the change's own branch, this one
included, so the loop's closing `-full` run would always be refused. A step-12 commit that moved a spec
or a PNG still sends the run back to **Verify**, as signal seven demands before landing.

### Main-session columns on the Stage leaderboard

**ID:** leaderboard-main-columns
**Status:** active
**Chosen:** **Main session** (USD to the cent) and **Main share** (integer percent) after **Total
cost**, sortable, exactly as the approved prototype draws them. Main session is a stage run's cost minus
its own dispatch buckets (`metrics.dispatches.*.cost_usd`, tails included) — the rule the run-detail view
prints as "agents $X + main session $Y" (`stats/web/src/components/StageRunTable.tsx`) — summed over the
period, each run's figure floored at zero; a run with no dispatch bucket is all main session; a run with
any bucket lacking `cost_usd` is left out of both main figures, never priced as zero; Main share divides
Main session by the total cost of the same runs. Under a model filter a run's main figure is that
model's `models.<m>.cost_usd` minus the dispatch buckets that model served, a bucket's served model read
as `dispatchBucket.servedModel()` (`stats/internal/store/pricing.go`) reads it. A stage none of whose
runs yields a main figure draws `—` in both cells, the SPA's existing absent-value mark.
**Considered:** pricing a `main_cost_usd` into every stage run and repricing history (a Plan agent's
route) — the runs' own dispatch buckets already give main's figure, 33% by that rule against 34% priced
from main tokens over the same 30 days, with no new field and no backfill.

### The sorted state keeps the board's sort mark

**ID:** sort-indicator-kept
**Status:** active
**Chosen:** a click on **Main share** sorts ascending and the header reads `Main share ▲`, the mark
`DataTable` gives every sorted column; the prototype's own sort script, a stand-in on a static render,
sorts the same way and draws no mark. The mark is that capture's one expected difference from the
prototype.
**Considered:** dropping the mark to match the stand-in script — it changes every sortable table on the
board, outside what the prototype was drawn to specify.

### The Done-when's baseline is the measured one

**ID:** done-when-baseline
**Status:** active
**Chosen:** main's share of `flow.visual-verify` is measured against 34% / $332 over the 30 days to
2026-10-30 at that day's prices, and 38% / $0.88 per run since 2026-10-28 — the ticket's Done-when
amended to it on 2026-10-30.
**Considered:** the ticket's 58% / ~$540 — an estimate at older opus prices with fix-loop work still
billed to the stage (fixed by `e37af3f4`).

### One call for the measuring chain

**ID:** measure-one-call
**Status:** active
**Chosen:** `visual-verify-pass.sh measure <worktree> <name> <key> <spec> [--prototype]`, run by the
verifier right after step 8's capture (after the prototype render and any sidecar it writes), runs
`resolve-visual-screenshots`, `check-identical-baselines`, `compose-mockup-frames` and
`measure-visual-properties` (`--props bands,seams` and `--props matrix`, `--scale 1`) in-process on every
pair, carrying on past a non-zero exit. It writes each frame's matrix JSON to
`<changeRoot>/visual-verification/matrix/<frame id>.json`, the full bands/seams JSON under
`<abs-worktree>/.superpowers/sdd/visual-measure-<key>/`, and the manifest
`<abs-worktree>/.superpowers/sdd/visual-manifest-<key>.json` — per pair its paths, exits, `diff=` ratio
and flagged items only (missing, extra and over-tolerance bands and seams, and the matrix's `over`
rows). It exits 0 when every inner step exited 0, 1 when the manifest is written and an inner step
exited non-zero, 2 on a usage error or an unwritable output. The verifier reads the manifest and every
image it names in one turn of parallel reads; the separate `--props elements` call goes, since `matrix`
pairs the elements itself.
**Considered:** round 1's condition that `measure` run only when the round's diff touches no spec,
sidecar or fixture — the call runs after the verifier's own capture, so it never reads stale input, and
running it every pass is what keeps the matrix JSON fresh; exit 0 whatever the inner exits — a
non-zero exit would then be visible only inside the manifest.

### The matrix measurement lists every over-tolerance element

**ID:** over-list
**Status:** active
**Chosen:** `measure-visual-properties.sh --props matrix`'s `summary` gains `over`, one entry per matrix
row (`row`, its 1-based index in `matrix`) that is `missing` or `extra`, carries `unresolved`, or is
paired with a length delta — any scaled numeric leaf of its pairing or its `measured.delta` — over 2px
or a colour distance over 12. The two thresholds are step 10's band and seam tolerances, now named Go
constants that step 10 cites.
**Considered:** naming rows by their element's coordinates — the index is the one handle both the
manifest and the report can name without rounding.

### Every flagged element is explained in the report

**ID:** over-tolerance-line
**Status:** active
**Chosen:** `check-verify-report.sh` requires, per composed frame, `- <frame id> over-tolerance: <n> —
#<row>: <cause>; …`, `<n>` the frame's `over` count in this dispatch's manifest
(`visual-manifest-<key>.json` beside the report, `<key>` the report's key with any `-rest[-<n>]` segment
removed) and every row named. The cause list is closed: `departure`; `annotation`; `transcription`;
`text width`; `decision <ID>`, the ID an `**ID:**` of `design.md`; `present at the merge base` with the
merge-base evidence the guard already requires; `mockup source`, for an `unresolved` row only. A missing
line, a wrong count, an unnamed row or a cause off the list leaves step 10 undone, and step 10 `done`
with a composed frame and no manifest is undone.
**Considered:** reading `over` from the committed matrix JSON (round 2) — a verifier that skipped
`measure` would leave last round's file behind, the stale-file hole the manifest closes.

### One report, one round

**ID:** one-report-one-round-opener
**Status:** active
**Chosen:** every `In-run fix <k>` task carries `**Opener:** <the dispatch key whose report it
transcribes>`; `check-plan-shape` F24 refuses an `In-run fix` task with no `**Opener:**` and two naming
the same opener.
**Considered:** refusing duplicates alone — omitting the field would pass the check.

### The by-hand frame/matrix check becomes automatic

**ID:** by-hand-check-automated
**Status:** active
**Chosen:** `check-verify-report.sh`, which `close` runs, gains the mechanical parts: every composed
frame has its `bands:`, `seams:` and `matrix:` lines; no matrix table in `visual-verification.md` has an
empty cell; an n/a cell stands only in a column that does not apply to its row's kind (step 10's column
table) or reading `no mockup source`; every `sweeps:` item reads `done` or `n/a — <why>`. The parent's
by-hand reconciliation paragraph goes, and the main session reads no images.
**Considered:** keeping the by-eye half beside the automatic one — it costs main turns every pass and
has never caught anything traceable; what retiring it loses is a look, by someone besides the verifier,
for an element neither the verifier nor the script found.

### A matrix row's kind cell opens with its class

**ID:** matrix-kind-opens-with-class
**Status:** active
**Chosen:** a matrix table is a table whose header's first cell opens with `element`; its `kind` cell
opens with the row's class — `text run`; `control`, `row` or `card`; `rule`, `divider` or `fill band` —
which is what tells the guard which columns may be n/a. A row whose class is unknown admits only
`no mockup source` n/a cells.
**Considered:** classifying from free-text roles (heading, button, hairline…) — a heuristic the guard
would get wrong on the next new role word.

### Frames byte-identical to an earlier clean pass are still checked

**ID:** no-skip-identical-frames
**Status:** active
**Chosen:** not built; every pass re-checks every frame on the change's list.
**Considered:** skipping a frame whose capture is byte-identical to one an earlier pass judged clean
(round 1's follow-up 1, buildable in 3 tasks, ~$20–35/30d) — 12 of the 67 round-opening defects sat on
exactly such a capture and skipping would have shipped all 12; verifiers already carry verdicts forward
informally, and kan-930's round 6 surfaced right after one. Explaining every flagged element (decision
`over-tolerance-line`) makes the re-check cheaper to get right instead.

### The closing `-full` run stays

**ID:** full-run-stays
**Status:** active
**Chosen:** **One full run closes the loop** (`skills/flow/verify-fix-loop.md`) is unchanged.
**Considered:** dropping it (a Plan agent's largest lever, ~$137/30d) — most `-full` keys are not the
closing run; of the 30 sampled `-full`-opened rounds about 3 followed a clean scoped re-run, and 2 of
those found real departures (kan-1003 round 2, kan-1023 round 5).

### The verifier stays on opus

**ID:** verifier-stays-opus
**Status:** active
**Chosen:** opus at `flow-low`, its row begun by `prepare` with `-model opus -effort low`, so every
verifier row carries its model and effort by construction; no new dispatch kind.
**Considered:** capture on sonnet, judging on opus — the Done-when keeps the verifier on opus, and the
capture work is already scripts.

### Scope

**ID:** scope-six-parts
**Status:** active
**Chosen:** all six parts in this change, the operator's answers to rounds 1 and 2: the thin parent,
the retired `verify`, the leaderboard columns, the one measuring call, every flagged element explained
with one report one round, and the automatic frame/matrix check.
**Considered:** filing the measuring call and the explain-every-element rule as follow-ups — the
operator chose this change; rewriting cache-expired prompts after long verifier waits — 2 of 9 sampled
passes, 12% of creation, and retiring step 7 removes the longest waits anyway.

## Open questions

## Recommended approach

Go first, then the contracts that call it, then the board. `prepare` and `close` land as one entry with
three subcommands; the verify record and `check-verify-green` change with them so the verifier's step 7
can retire; the `over` list precedes `measure`, which writes it into the manifest
`check-verify-report.sh` then reads; F24 is independent. The contract text lands in two passes — the
thin parent with the retired `verify`, then the measuring call with the over-tolerance line and the
automatic check — each applying a patch planned with the change, since every sentence it adds or
removes is listed in `verbatim-moves.txt`. The leaderboard's store, view and capture run beside the
rest. Rejected: one subcommand, an orchestrator, guards in the verifier and a summariser
(`thin-parent-prepare-close`); `prepare` starting the stack (`parent-starts-the-stack`); a store-derived
stage (`stage-file-names-its-rows`); a scoped verifier `verify` (`verifier-runs-no-verify`); a same-HEAD
green rule (`green-verify-spans-planning-commits`); a priced `main_cost_usd`
(`leaderboard-main-columns`); a conditional `measure` (`measure-one-call`); the committed JSON as the
over source (`over-tolerance-line`); skipping identical frames (`no-skip-identical-frames`); dropping
the `-full` run (`full-run-stays`); a sonnet capture split (`verifier-stays-opus`).

## Critical files

- `stats/internal/guard/visualverifypass.go`, `visualverifypass_prepare.go`, `visualverifypass_close.go`,
  `visualverifypass_measure.go` and their `visual_verify_pass_*_test.go` — the new entry.
- `scripts/visual-verify-pass.sh`, `skills/flow/scripts/visual-verify-pass.sh` — its shim and the
  skill's symlink.
- `stats/internal/guard/stagemarkcalls.go` — holds the new call to the token and harness rules.
- `stats/internal/guard/runverifylist.go` — the verify row's `-diff-base`.
- `stats/internal/guard/verifygreen.go`, `scripts/check-verify-green.sh` — the HEAD rule.
- `stats/internal/guard/verifyreport.go`, `scripts/check-verify-report.sh` — the retired third argument,
  the over-tolerance line and the automatic checks.
- `stats/internal/guard/visualverifydispatched.go` — the retired rerun rule.
- `stats/internal/guard/measurevisualproperties.go`, `scripts/measure-visual-properties.sh` — `over`.
- `stats/internal/guard/planshape.go` — F24.
- `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`, `skills/flow/verify-fix-loop.md`,
  `skills/flow/verify-and-handoff.md`, `skills/flow/SKILL-rationale.md` — the contracts.
- `stats/internal/store/aggregate.go`, `stats/internal/api/stats.go`, `stats/web/src/api.ts`,
  `stats/web/src/views/StageLeaderboard.tsx`, `stats/web/tests/visual/views.spec.ts` — the board.

## Existing code to reuse

- `integrateChange` and its helpers (`stats/internal/guard/integratechange.go`: `ic.say`, `ic.stop`,
  `ic.markBegin`/`markEnd`, `icExit`) — the subcommand, closing-line and stage-mark pattern.
- The registry functions run in-process: `checkVisualTrigger`, `checkVerifyGreen`, `prepareWorkspace`,
  `checkVisualPreflight`, `checkSidecarReach`, `checkDevStackFresh`, `checkVerifyReport`,
  `checkFrameList`, `resolveVisualScreenshots`, `checkIdenticalBaselines`, `composeMockupFrames`,
  `measureVisualProperties` — `verifyreport.go` already calls `resolveVisualScreenshots` this way.
- `pfdRead` and `pcFlow` (`stats/internal/guard/flowcli.go`) — dispatch rows and `flow` calls.
- `specRootLeaf` and signal seven's `git log … ':(exclude)<leaf>/changes/'`
  (`stats/internal/guard/unfinishedwork.go`) — the planning-paths exclusion.
- `vrMockupSource` (`stats/internal/guard/verifyreport.go`) — the mockup sources `design.md` cites.
- The `## visual verification` readers in `stats/internal/guard/visualsection.go` and
  `runverifylist.go` (the `specs`/`verify` row and `<frontend-root>`/`<merge-base>` substitution).
- `mvpPairInOrder`, `mvpDelta`, `mvpUnscaled` (`measurevisualproperties.go`) — the matrix rows `over`
  reads.
- `psVisualFixRE`, `tcfParseTask`, `psFieldLine` (`stats/internal/guard/planshape.go`, `planparse.go`) —
  the `In-run fix` title and field readers F24 builds on.
- `readDispatchRows`/`readCostUsd` (`stats/web/src/metrics.ts`, `StageRunTable.tsx`) — the run-detail
  split the leaderboard's rule mirrors; `formatUsd` (`stats/web/src/format.ts`); `DataTable`'s sortable
  column (`stats/web/src/components/DataTable.tsx`).

## Verification

- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` clean; `cd stats && go test -race
  -count=1 ./internal/guard/...` and the `stats-go` suite green; `cd stats/web && npx tsc -b && npm test`
  green; `run-verify-list.sh <worktree> lint` clean, `check-verbatim-moves.sh` included.
- The Stage leaderboard capture in `views.spec.ts` matches the approved prototype at rest and sorted by
  **Main share**, save decision `sort-indicator-kept`'s mark.
- Live check below.

## Live check

Every step runs at `flow.verify`, against a scratch clone of this worktree, the disposable UI-test stack
on 4174 and a copy of the dev store; nothing touches `flowd` on 4173 or the `flow` database, no record
lands in the real store, and no commit lands on the change branch.

1. **The stand-ins.** `git clone --quiet --branch spectre/<name> <worktree> <scratch>/kan-1060-live` —
   the scratch clone every step below passes as `<clone>`, so the commit `close` makes lands there. Then
   `scripts/live-store-copy.sh up kan-1060` and export the `FLOW_ADDR` / `FLOW_RECORDS_ADDR` lines it
   prints, so every `flow` call below — the ones `prepare` and `close` make included — writes the copy.
   In it, record a green inline verify row at the clone's HEAD under the scratch token
   `mf-k1060-livecheck`: `flow record dispatch begin -change <name> -role verifier -model
   opus -effort medium -agent-id inline -key verify-livecheck -session-token mf-k1060-livecheck
   -diff-base <HEAD>`, then `flow record dispatch end … -outcome completed`.
2. **`prepare` on this change's own board.** `scripts/visual-verify-pass.sh prepare <clone> <name>
   visual-verify-livecheck --harness claude-code --session-token mf-k1060-livecheck --merge-base
   <clone>=<merge base> --app <clone>:stats/web=http://127.0.0.1:4174`; on `NEXT: start-stack`,
   `make -C stats ui-test-up` and `prepare … --continue --started`. Record the closing line, the bundle
   path and its resolved `capture` row, and the `visual-verify-livecheck` row `flow record dispatches`
   shows — `model` opus, `effort` low, a `diffBase` equal to HEAD. Failure looks like: `NEXT: dispatch`
   with no row begun, a bundle without the resolved `capture` row, or `VERIFY-NOT-GREEN` over step 1's
   row.
3. **`measure` on this change's board captures.** Render the approved prototype per step 10's prototype
   path — with this worktree's installed Playwright — into the clone's
   `<changeRoot>/visual-verification/prototype/`, then `scripts/visual-verify-pass.sh measure <clone>
   <name> visual-verify-livecheck <clone>/stats/web/tests/visual/views.spec.ts --prototype`. Record the manifest's pairs and each pair's `over` count. Failure: no pair for the
   leaderboard captures, a manifest with zero pairs, or exit 0 while an inner exit is non-zero.
4. **`close` and the report guard.** Write `verify-report-visual-verify-livecheck.md` by hand from step
   3's manifest — a `## Verdict`, and a `## Report` whose `over-tolerance:` lines name every row — and
   run `scripts/visual-verify-pass.sh close <clone> <name> visual-verify-livecheck 0 --session-token
   mf-k1060-livecheck`; record its closing line, the row's outcome and the commit it made in the clone. Then drop one row from an
   `over-tolerance:` line and run `scripts/check-verify-report.sh` on it: exit 1 naming that row.
   Then run `scripts/check-verify-report.sh` on a scratch report whose `--design` is
   `/Users/tweety53/Projects/gymie/spectre/changes/archive/kan-938-step-3-frontend-cardio-type-picker-on-add-a/design.md`
   (its `visual-verification.md` beside it), step 10 `done`: exit 1 naming the n/a and empty cells of
   lines 100, 169 and 228. An exit 0 on either reads as failure.
5. **The leaderboard on real data.** On the copy's port, read `GET
   /api/v1/stats/stage-leaderboard?from=2026-09-30T00:00:00.000Z&to=2026-10-30T00:00:00.000Z` and open
   the board: record `flow.visual-verify`'s Main session and Main share beside decision
   `done-when-baseline`'s 34% / $332. A main figure equal to its total on a stage whose runs carry
   dispatches, a missing column, or a visual-verify share outside 25–45% reads as failure.
   <!-- predicted: the copy's leaderboard read at flow.verify; 34% / $332 measured on 2026-10-30 in the brainstorm rounds -->
6. **Clean up.** `make -C stats ui-test-down` if step 2 started the stack, `scripts/live-store-copy.sh
   down kan-1060`, and `rm -rf <scratch>/kan-1060-live`.
