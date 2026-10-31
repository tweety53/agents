## Context

Found by the self-review pass of `/flow` change `kan-1062-flow-implement-one-shot-subagents`
(`.superpowers/sdd/self-review-pass.md` in this worktree); the fixes ride that change's own branch
`spectre/kan-1062-flow-implement-one-shot-subagents`, based at its prompt-tuning commit. The
brainstorm design is
`.superpowers/sdd/2026-10-31-self-review-kan-1062-flow-implement-one-shot-subagents-design.md`.
Finding 6-A is disproved and not planned: the mutation+failure-modes slot begins only once the
full suite exits (**On pass 1 the full suite runs between the begins and the wait**,
`skills/flow/review-panel.md`).

**Finding 1-B — reproducer placeholder read literally (flow-fix).** The REPRODUCER paragraph
(`skills/flow/review-panel.md`) names the file `R-ID-N.sh`; `renderslotprompt.go` appends
"— `<round>` here" after "`R` this round's number" and leaves `R-ID-N` standing. The shape the rest
of the contract states and `check-panel-reproducers.sh` enforces is `<round>-<id>-<n>.sh`.

**Finding 1-C — whole-package `go test ./cmd/flow` leaves a survivor (flow-fix).**
<!-- measured: scripts/run-reproducer.sh "$PWD" .superpowers/sdd/reproducers/sr-1c-whole-package.sh (a script running make -s web-dist and go test ./cmd/flow -count=1 from stats/) → "unverifiable — … exited 0 inside the 600s bound, but it forked a detached child that outlived it … surviving process pid(s): 78536" (exit 3) @ 2026-10-31, this worktree at the plan base -->
`runStageWrap` (`stats/cmd/flow/stage.go`) sets `cmd.Cancel` to SIGTERM the child's pid and
`cmd.WaitDelay` to `defaultTimeout`; after the delay Go kills that pid alone, so a TERM-ignoring
grandchild is orphaned.
<!-- measured: a scratch Go program running sh -c 'trap "" TERM; sleep 30 & echo $! > f; wait' under CommandContext with WaitDelay 500ms → "sleep alive after 2s: true" with the pid-only Cancel; with Setpgid, a group SIGTERM Cancel and a group SIGKILL once Run returned on a cancelled context → "sleep alive after 2s: false" @ 2026-10-31, darwin 25.5.0 -->

**Finding 1-D — gated findings missing from the self-review bundle (flow-fix).** `RenderPanel`
(`stats/internal/records/render.go`) filters `GatedReviewerSlot` rows through `PanelFindings`, as
the panel guards require, and `Bundle` (`stats/internal/selfreview/git.go`) adds only that render;
`run.Findings` holds the gated rows.

**Finding 1-E — the live-store copy has no scratch change (flow-fix).** `flow record` refuses a
change no checkout carries under `spectre/changes/`, `spectre/changes/archive/` or
`.superpowers/sdd/` (`changeCheckouts`, `stats/cmd/flow/record.go`), and the store answers a write
for a change with no state record `ErrChangeNotFound` (`stats/internal/store/records.go`).
`pipeline-fix.sh start` already meets both (`pfStart`, `stats/internal/guard/pipelinefix.go`).

## Decisions

### The reproducer name in the REPRODUCER paragraph

**ID:** reproducer-name-placeholders
**Status:** active
**Chosen:** the paragraph names `<round>-<id>-<n>.sh` — `<round>` this round's number, `<id>` the slot's own resolved reviewer id, `<n>` its own 1-based finding index — with the worked example "round 0's first mutation finding is `0-mutation-1.sh`"; `renderslotprompt.go`'s phrase constant becomes "`<round>` this round's number", still followed by "— `<round>` here". The reworded sentence is listed in this change's `verbatim-moves.txt`.
**Considered:** substituting the round into the name itself in the renderer — a second replacement whose only gain the "— `<n>` here" note already gives; keeping `R-ID-N.sh` and adding the example alone — the bare letters are what three runs misread.

### How stage wrap reaches a grandchild

**ID:** stage-wrap-kills-group
**Status:** active
**Chosen:** `runStageWrap` starts the child with `SysProcAttr{Setpgid: true}`, `Cancel` sends SIGTERM to `-pid`, and once `Run` returns with `ctx.Err() != nil` it sends SIGKILL to `-pid` (an `ESRCH` ignored), reaping what `WaitDelay`'s pid kill missed; a child that exits on its own keeps any background process it started. No skill or script wraps an interactive command, so the SIGTTIN a background group gets on a terminal read reaches no caller.
<!-- measured: grep -rn "stage wrap" skills scripts → no output @ 2026-10-31, this worktree -->
**Considered:** `exec sleep 30` in the test — the test passes while every real wrapped `sh -c` still orphans its TERM-ignoring work; a timer re-killing the group after `WaitDelay` inside `Cancel` — the same reach with a goroutine that outlives `Run`.

### Where the gated findings enter the bundle

**ID:** gated-findings-own-source
**Status:** active
**Chosen:** `records.RenderGatedFindings(r Run) (string, bool)` in `render.go` renders the `GatedReviewerSlot` rows as one table, `| ID | Task | Severity | Location | Note | Status |`, through `tableCell`, `ok` false when the run holds none; `Bundle` adds it after the panel source under the label `gated per-task reviewer findings`, so the bundle counts 10 sources and reports the source `skipped … (absent)` when no gated row exists. No contract prose lists the bundle's sources, so no skill edit is needed.
**Considered:** appending a gated section to `RenderPanel` — the panel record is what `check-unfinished-work.sh` and the panel guards parse, and a second findings table there would be read as panel rows; adding the source only when gated rows exist — the found/skipped count would then vary by change instead of naming an absence.

### What live-store-copy registers

**ID:** live-copy-registers-scratch-change
**Status:** active
**Chosen:** `up <name>` refuses (exit 2, before any side effect) when `<checkout>/.superpowers/sdd/<name>/` exists, creates it, records its path in the state directory, and once the copy's `flowd` answers writes a STARTED state record for change `<name>` with `FLOW_ADDR` pointed at the copy, through `flow state set -C <checkout> <name>` reading the record `pfRecord` writes (`state`, `branch`, `worktrees` of the checkout at `HEAD`, the three null URLs and issue, `updatedBy` `/flow`); a failed write is exit 1. It prints `scratch change: <name>`. `down` removes only the directory `up` recorded. `.flow/project.md`'s tool sentence says so.
**Considered:** printing the two commands only — the setup stays a hand step, which is the defect; a separate scratch-change argument — the copy's name already names one scratch change per copy.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 2 changes process-group signalling in `flow stage wrap`, platform interop where a wrong `-pid` reaches the wrong processes.
**Considered:** `flow-low` — Tasks 1, 3 and 4 are a sentence, a render and a script step, but Task 2's kill path needs the closer read.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch for Tasks 1–4 — four file-disjoint tasks of a few dozen lines each.
**Considered:** one fixer per task — three extra dispatches for small, independent edits.

## Open questions

## Recommended approach

Fix each defect where every caller routes through it: the REPRODUCER paragraph every slot prompt
renders (`reproducer-name-placeholders`), `runStageWrap`'s kill path (`stage-wrap-kills-group`),
`Bundle`'s source list (`gated-findings-own-source`) and `live-store-copy.sh up`
(`live-copy-registers-scratch-change`).

## Critical files

- `skills/flow/review-panel.md`, `stats/internal/guard/renderslotprompt.go`, `stats/internal/guard/render_slot_prompt_test.go`.
- `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`.
- `stats/internal/records/render.go`, `stats/internal/selfreview/git.go`, `stats/internal/selfreview/bundle_test.go`.
- `scripts/live-store-copy.sh`, `scripts/test-live-store-copy.sh`, `.flow/project.md`.

## Existing code to reuse

- `tableCell`, `neutraliseMarkers` and `GatedReviewerSlot` (`stats/internal/records`).
- The group-kill pattern of `stats/internal/guard/cleanupcomplete.go` (`Setpgid`, `syscall.Kill(-pid, …)`).
- `pfStart`'s directory and `pfRecord`'s state record shape (`stats/internal/guard/pipelinefix.go`).

## Verification

- `cd stats && make -s web-dist && go test -count=1 ./internal/guard/ -run 'TestRenderSlotPrompt'`, `./cmd/flow/ -run 'TestRunStageWrap'`, `./internal/selfreview/ ./internal/records/` pass; `go vet ./...` and `gofmt -l .` are clean.
- `scripts/test-live-store-copy.sh` prints `all cases passed`.
- `scripts/run-reproducer.sh <worktree> .superpowers/sdd/reproducers/sr-1c-whole-package.sh` exits 1, naming no surviving pid.
- `scripts/run-verify-list.sh <worktree> lint` is green, `check-verbatim-moves.sh` acknowledging this change's `verbatim-moves.txt`.

## Live check

None — the changes are a contract sentence, a wrapper's kill path, a bundle render and a dev-only script, each exercised by its own test or harness against fixtures; no running service or stored state changes.
