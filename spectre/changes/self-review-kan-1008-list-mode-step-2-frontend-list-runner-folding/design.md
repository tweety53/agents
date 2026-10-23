## Context

The ten findings of `proposal.md` are planned as one self-review change on branch
`self-review-kan-1008-list-mode-step-2-frontend-list-runner-folding`, based at origin/main
`50045985`. F3 lives in gymie-frontend and F9, F10 in gymie-playwright; per **Pipeline defects found
mid-run** (`skills/flow-contracts/pipeline.md`) they are fixed on kan-1008's own change branch
`spectre/kan-1008-list-mode-step-2-frontend-list-runner-folding`, in
`/Users/tweety53/Projects/gymie-frontend-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding`
(HEAD `0276408a`) and
`/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1008-list-mode-step-2-frontend-list-runner-folding`
(HEAD `4258134`).

Every agents-repo fix and F3's line were drafted and checked in this worktree at `50045985` (F3
against the gymie-frontend worktree's HEAD); each task's diff is a file under `patches/` beside this
file, and `tasks.md` carries the measured results. With Tasks 1, 2 and 4–8 applied together,
`go test ./internal/guard` (159 s), `go test ./internal/store`, `./cmd/flow`, `./cmd/flowd`,
`./internal/records`, `./internal/api`, `./internal/reconcile` and `./internal/client` pass, `gofmt -l`
and `go vet ./...` are clean, and `check-references`, `check-markdown-integrity`,
`check-dispatch-paragraphs`, `check-installed-citations`, `check-guard-symlinks`,
`check-stage-mark-calls` and `check-normative-inventory` exit 0.
<!-- measured: those commands in this worktree with the seven patches applied @ 50045985 -->

**History.** `git log --since=2026-10-21` over every path the findings name finds no fix of any of
them; `7287cc2c` (kan-1019) added integrate's `liveCheckGate`, which is F1's premise. No archived
design in agents or gymie rejects a remedy proposed here.

**F1's trace.** `skills/flow/verify-and-handoff.md` **Live check** already says the check is never
deferred past `flow.verify`, but **Close the stage** ends the stage with no command that could
refuse; integrate's `liveCheckGate` is the first mechanical check. `icLiveCheckNamed` (same
package) is the predicate both gates share.

**F2's trace.** `flow record dispatches -change kan-1008-… -C <gymie-frontend worktree>` returns
`[]` and the backend worktree returns the rows; `verifygreen.go` treats the empty set as `seen == 0`
and exits 1.

**F3's spike.** In the gymie-frontend worktree, `./gradlew test allTests --dry-run` schedules
`:shared:…IosSimulatorArm64…` tasks; with `-x iosSimulatorArm64Test` it schedules none (`grep -ci
ios` → 0, 226 tasks).
<!-- measured: ./gradlew test allTests [-x iosSimulatorArm64Test] --dry-run | grep -i ios @ gymie-frontend 0276408a -->

**F5's trace.** `framelist.go`'s per-frame pattern is `^\s*- <id>( [^:]*)?:`; `- LM-L9 / LM-L11 …:`
matches LM-L9 through the optional tail and never LM-L11.

**F7's measurement.** `flow record dispatches -change kan-1008-list-mode-step-2-frontend-list-runner-folding`
(from `/Users/tweety53/Projects/gymie`): id 7195 seq 48 `visual-verify-fix-3-full-gymie`, agent
`a9a8709ccc39e1f2f`, diffBase `aa758b6a`; id 7196 seq 49 `visual-verify-fix-3-full-gymie-heads`,
agent null, diffBase
`aa758b6a,79c0541d10701fad1b7542f313daaf3d38cdbbbf,980d3332936494f5eb80a8340747c60a91deaa55,aa758b6a41d419a353612f2e74f1d46fe2762714`,
metrics `{}`, both session token `mf-k1008r2`. No store command deletes a dispatch row, and only
`findings.dispatch_id` references `dispatches(id)`.
<!-- measured: flow record dispatches -change kan-1008-list-mode-step-2-frontend-list-runner-folding @ dev store 2026-10-23 -->

**F8's trace.** gymie's `5c19610e` and `4e6e3e77` carry `Task-Id: 14`, `6868c140` `Task-Id: 13`;
close-task judged only the map the parent passed, which named no backend pair.

**F9, F10.** Not traced at plan time; both tasks open on the trace against the running kan-1008
stack (frontend `http://localhost:6740` and gateway `11820` answered 200 at plan time).

## Decisions

### Live-check gate at the verify stage end

**ID:** live-check-gate-at-verify-end
**Status:** active
**Chosen:** `check-live-verification.sh <changeRoot>` — a directory argument reads `design.md` through `icLiveCheckNamed`, passes `LIVE-CHECK-NONE` when no check is named, and otherwise judges `live-verification.md`, a missing one exit 1 — chained with `&&` before `flow stage end -stage flow.verify` in **Close the stage**.
**Considered:** refusing in `flow stage end` itself — the CLI would have to resolve the change root for one stage of one command, a wider change for the same verdict; a new `check-live-check-gate.sh` — a second script and companion for what one guard's directory form answers.

### Empty store read cannot answer

**ID:** verify-green-empty-read-exit-2
**Status:** active
**Chosen:** zero rows of the change at all → exit 2 on stderr naming the canonical worktree; the skill line passes `<canonical-worktree>`, the member **Close the stage** renders the ledger into.
**Considered:** resolving the canonical worktree inside the guard — the caller already knows it, and a guard that guesses another worktree hides the wrong argument.

### iOS excluded from the frontend test line

**ID:** frontend-test-excludes-ios
**Status:** active
**Chosen:** `./gradlew test allTests -x iosSimulatorArm64Test`, plus one paragraph saying why — the dry run schedules no iOS task.
**Considered:** listing every wanted test task instead of `allTests` — a line that silently drops a new target's tests; disabling the iOS target in the build — a build change outside this finding.

### A departure never shortens the pass

**ID:** departure-never-shortens-pass
**Status:** active
**Chosen:** one bold sentence in **No required step is left undone** (`skills/flow/visual-verify-verifier.md`) naming the excuse kan-1008's verifier gave.
**Considered:** a `check-verify-report` change — the guard already refused the report; the cost was the verifier choosing to stop, which only the contract text reaches.

### Slash-joined frame ids

**ID:** frame-list-slash-ids
**Status:** active
**Chosen:** the per-frame pattern admits `([^\s:]+ / )*` before the id, so every id of `- A / B …:` is reported.
**Considered:** requiring one line per frame in the report template — kan-1008's verifier wrote the joined form for frames sharing one verdict, and the guard should read the report as written.

### Height motions in a tall viewport

**ID:** tall-viewport-for-height-motion
**Status:** active
**Chosen:** one bold sentence in the motion step: step 8's viewport made taller for that recording alone.
**Considered:** a fixed taller viewport for every motion — changes every recorded strip's frame, where only height motions need it.

### Diff base on dispatch end, and a fold command for the repair

**ID:** end-diff-base-and-fold
**Status:** active
**Chosen:** `flow record dispatch end -diff-base` sets the row's `diff_base` (`COALESCE`, so an omitted flag keeps begin's); one skill line in **Recording.** (`skills/flow/visual-verify.md`); `flowd fold-dispatch <keep-id> <drop-id>` moves an agentless row's diff base onto the kept row and deletes it in one statement, refusing rows of different changes or session tokens or a drop row with an agent id; the repair runs `fold-dispatch 7195 7196`.
**Considered:** a generic `delete-dispatch` plus the new `end` flag — a sharper tool than the repair needs, and the live daemon would have to be rebuilt first for `end` to carry the field; editing the rows with `psql` — unrepeatable and untested.

### close-task joins the recorded worktrees

**ID:** close-task-joins-recorded-worktrees
**Status:** active
**Chosen:** for each task, every worktree of `flow state get`'s worktrees map the argument left out adds its newest commit since `vmDefaultBase` whose `Task-Id:` names the task, as one more `<worktree>=<sha>` pair, named on stderr; the push covers it too.
**Considered:** refusing a map that omits a recorded worktree — a worktree with no commit of the task is the ordinary case and would refuse every cross-repo close; scanning every worktree on the machine — the recorded set is the change's own.

### Trace J5a before fixing

**ID:** floor-flake-trace-first
**Status:** active
**Chosen:** the task drops the describe's `retries: 6`, reproduces J5a under the verify's 5-worker load, traces it, and fixes it at the source per `rules/fix-determinism-at-the-source.mdc`; a floor test the drop unmasks is traced and fixed the same way, or reported `big`.
**Considered:** lowering the retry count — still a widened tolerance; dropping retries without a trace — swaps a hidden flake for a red suite.

### Trace the inset timeouts before fixing

**ID:** inset-timeout-trace-first
**Status:** active
**Chosen:** the task reproduces the `:364`/`:446` timeouts under 5-worker load first, names the boot step that spends the budget, and fixes it at the source; never a raised timeout. If the stated runs do not reproduce them, the task makes no commit and records the runs here, as Task 10 Step 3 allows. They did not reproduce: `npx playwright test fidelity/kan-749-inset-scroll-runner.spec.ts --repeat-each=10 --workers=5` → 70/70 passed; kan-749 + list-mode + quick-mode + full-app-suite at `--repeat-each=3 --workers=5` → 150 passed, 3 skipped; the verify artifacts kept no error context for the original failures.
**Considered:** raising the timeout — it would pass with the defect present.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — a store statement that deletes a row, a CLI wire field, and close-task logic every task close runs, beside two flake traces whose fix is not known yet.
**Considered:** `flow-low` — every agents-repo diff is small and pre-verified, but a wrong fold predicate or join reaches every later run.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** two groups, in series — g1, key `self-review-kan-1008-list-mode-step-2-frontend-list-runner-folding-fix-g1`: Tasks 1–8 in order, every one a pre-verified patch (Task 3 in gymie-frontend); g2, key `…-fix-g2`: Tasks 9 and 10 in gymie-playwright, trace work on the running stack.
**Considered:** one group — the mechanical patches would wait behind open-ended trace work in one context; one fixer per task — ten dispatches for eight one-patch tasks.

## Open questions

## Live check

F7's repair writes the dev store. Before (measured at plan time, above): id 7195 diffBase
`aa758b6a`, id 7196 present, agentless. After Task 7's repair step, from
`/Users/tweety53/Projects/gymie`, `flow record dispatches -change kan-1008-list-mode-step-2-frontend-list-runner-folding`
shows id 7195 carrying id 7196's four-sha diffBase and no id 7196. Not working looks like
7196 still listed, or 7195 still at `aa758b6a`. The other findings have no runtime of their own:
guards and a CLI flag pinned by their tests, prose, a project-file line, and two test-harness fixes
whose proof is their own repeated runs.
