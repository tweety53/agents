# self-review-kan-1008-list-mode-step-2-frontend-list-runner-folding

## Why

The self-review pass of kan-1008 (List mode step 2, frontend List runner folding) found ten
non-`big` defects in the pipeline and in two of the change's own repositories:

1. **F1** — `flow.verify`'s stage end checks nothing: the only live-check gate is integrate's
   `liveCheckGate` (`stats/internal/guard/integratechange_prepare.go`), so kan-1008's live check
   first ran at integrate, its product defect landed after the panel, and a fresh 24.9-minute
   visual verify followed.
2. **F2** — `check-verify-green.sh` given a sibling worktree reads no row of the change and reports
   "this run recorded no inline verify", exit 1, as if verify had never run.
3. **F3** — gymie-frontend's declared `## test` line, `./gradlew test allTests`, pulls in
   `iosSimulatorArm64Test`, whose link always fails; iOS is not a target.
4. **F4** — a visual verifier left steps 10/11/motion "by choice" once it had found departures,
   costing two remainder dispatches (19 and 10 minutes).
   <!-- measured: the self-review pass's dispatch durations for kan-1008, as its findings file records them (not re-runnable) -->
5. **F5** — `check-frame-list` reads `- LM-L9 / LM-L11 …:` as LM-L9 only, so LM-L11 counted as
   unreported and the dispatch was blocked.
6. **F6** — a height motion recorded in an 816px viewport twice read "no height animation", the
   item's bottom edge having left the viewport.
7. **F7** — a verifier's wrong `-diff-base` could only be corrected by back-filling a second,
   agentless row: `flow record dispatch end` takes no `-diff-base`. kan-1008's rows 48/49 carry it.
8. **F8** — `close-task` judged only the worktrees of the parent-supplied commit map; the map
   omitted the backend repository, so backend commits of Tasks 13 and 14 whose paths were absent
   from `Files:` passed.
9. **F9** — gymie-playwright's `fidelity/floor-states.spec.ts` runs under `retries: 6`, a widened
   retry count hiding J5a's flake (`:553`, flaky in kan-1008's verify).
10. **F10** — `fidelity/kan-749-inset-scroll-runner.spec.ts:364` and `:446` timed out at 15 s in a
    5-worker verify with the page still booting; no cause found.

## What changes

- `check-live-verification.sh` also takes the change directory, and **Close the stage**
  (`skills/flow/verify-and-handoff.md`) chains it before `flow.verify`'s stage end (F1).
- `check-verify-green.sh` exits 2 naming the canonical worktree when the store read returns no row
  of the change; **Visual verification** passes `<canonical-worktree>` (F2).
- gymie-frontend's `## test` line excludes `iosSimulatorArm64Test` (F3).
- The verifier contract gains two sentences: a departure never shortens the pass (F4), and a height
  motion is recorded in a viewport tall enough for its settled height (F6).
- `check-frame-list` accepts ` / `-joined frame ids on one per-frame line (F5).
- `flow record dispatch end` takes `-diff-base`; `flowd fold-dispatch <keep-id> <drop-id>` repairs
  a row corrected the old way; kan-1008's rows 48/49 are folded (F7).
- `close-task` joins every recorded worktree of the change the commit map left out (F8).
- J5a's flake and the kan-749 inset timeouts are traced and fixed at their source, J5a's widened
  `retries: 6` removed (F9, F10).
