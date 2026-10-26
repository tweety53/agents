# self-review-kan-1017-templates-step-3-admin-templates-queue-create

## Why

kan-1017's self-review pass raised fifteen findings to fix, none `big` (SR-9 and SR-10 fold into
SR-11 and SR-12; the flow-improvement and flow-eval entries are noted, not fixed):

- **Panel close unguarded (SR-1).** `check-panel-findings-closed.sh` ran only for round 1; rounds
  2–7 closed without it and nine Critical/Important findings reached integrate unrecorded.
- **Late-fix reason hides the refs (SR-2).** `check-late-fix-trigger` drops the findings-closed
  guard's stderr from its condition-1 reason, so the parent misread a record gap as a scope cause.
- **flowBase sibling worktrees unchecked (SR-3).** The rule that a flowBase change needs a worktree
  in every `## apps` repository is prose only; the first visual verify ran against the wrong build.
- **Verify-list rows left open, then hand-closed wrong (SR-4).** `run-verify-list` never ends a
  failed list's row and cannot end a sub-3s skipped list's row; three rows were closed by hand
  with invented end times.
- **Verifier beside a mutator (SR-5).** A visual verifier ran while a panel-fix edited the same
  worktree; nothing refuses it.
- **Reviewer left a stray file (SR-6).** A panel slot's reproducer left an untracked file in a
  change worktree; the tree-markers guard reads markers only.
- **Steps shipped unticked (SR-7).** Tasks 11–16 were ticked by `sed` on the task line; 27 step
  lines stayed `[ ]` and nothing refuses that.
- **Phone frame geometry crops desktop frames (SR-8).** `mockup frame` is project-wide; every
  admin frame failed composition.
- **Matrix assembled by hand (SR-11, SR-9).** Step 10's matrix is one script call per element, and
  a band at the image edge exits 2.
- **No mockup source read (SR-12, SR-10).** `design.md` cites a tracked bundled HTML source, yet
  every verifier reported `no mockup source`.
- **Band colour flips (SR-13).** The band `colour` is a modal pick that flips between border and
  text on equal bands.
- **Redis handshake flake (SR-14).** 36 `AuthFlowIntegrationTest` tests failed once on a Redis
  handshake; unexplained.
- **Merge order hand-written (SR-15).** `kickoff-worktree.sh --from` leaves the `link.md`
  `## Merge order` record to hand-writing.
- **Files widened by free-hand sed (SR-16).** One widening `sed` rewrote five other tasks'
  `**Files:**` lines.
- **Unscoped visual re-runs (SR-17).** Six visual re-runs after in-run fixes ran the whole suite
  unscoped.

## What changes

- `flow stage end -stage flow.review-panel -outcome completed` runs the findings-closed guard in
  process; the late-fix trigger names the open refs.
- `check-visual-preflight.sh` fails a flowBase change missing an `## apps` sibling worktree.
- `run-verify-list` ends every row it begins — `blocked` on a failed list, a new `skipped` outcome
  on an all-skipped one the store admits under its inline floor; `flowd reoutcome-dispatch` exists,
  and kan-1017's dispatch rows 47, 65 and 76 are retimed, re-outcomed and restamped.
- `flow record dispatch begin` refuses a visual verifier beside an open mutator and the reverse,
  and an unscoped or misnamed visual re-run after an in-run fix.
- `check-tree-markers.sh verify` fails on a path a slot left changed or untracked.
- `check-task-records.sh` refuses a ticked task with an unticked step; kan-1017's archived
  `tasks.md` gets its 27 steps ticked on the gymie change branch.
- A `<spec>.mockups` sidecar may carry `frame <geometry>|none`; kan-1017's admin sidecar says
  `frame none`.
- `measure-visual-properties.sh --props matrix` fills the matrix in one call with regions clipped to
  the image; band `colour` reports the darkest ink plus the top-two share.
- The visual-verify prompt names `mockup sources:` with a decode one-liner for bundled pages, and
  `check-verify-report.sh --design` refuses `no mockup source` while a cited source exists.
- `kickoff-worktree.sh --from` writes or extends the canonical `link.md` `## Merge order`.
- `flow tasks files <change> <id> add <path>…` widens exactly one task's `**Files:**`.
- gymie's Redis handshake flake is bisected and fixed at its source on the kan-1017 change branch.
