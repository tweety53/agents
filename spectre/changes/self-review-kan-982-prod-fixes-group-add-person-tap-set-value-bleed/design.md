## Context

- Eight non-`big` findings from the kan-982 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each kept finding
  is carried by exactly one task; two are dropped below. All were checked against the tree at
  `bc44caac`:
  - 1 (kept, Task 3): `checkVisualVerifyDispatched` (`stats/internal/guard/visualverifydispatched.go`)
    calls `visualTrigger(env, worktree, …)` for the worktree it is given and nothing else, and
    `scripts/integrate-change.sh` calls it once per worktree with no canonical worktree. Its store
    reads run `-C <worktree>`, which in a second repository resolves that repository's project key —
    the hazard `unfinishedwork.go`'s anchor comment already documents.
  - 2 (kept, Task 2): `closeTask` (`stats/internal/guard/closetask.go`) prints a `WARNING` per
    undispatched appended task and carries on to tick and push, as decision
    `close-task-warns-undispatched` (archived `self-review-kan-938-…/design.md`) chose.
  - 3 (kept, Task 4) and 5 (kept, Task 5): `skills/flow/verify-fix-loop.md` step 2 says nothing of a
    load flake or of a fix whose flake does not reproduce; `implement.md`'s PROOF RUNS paragraph
    sizes the runs but never says what a non-reproducing flake does to the fix.
  - 4 (kept, Task 6): **The motion step** (`skills/flow/visual-verify-verifier.md`) writes frames as
    `<nnn>.png` and states no same-sampling rule for comparing two triggers.
  - 6 (dropped): the pipeline itself makes trailerless non-chore commits before the panel — the
    parent's own full-suite fixes (**The parent runs the full suite**, `skills/flow/implement.md`) —
    and after it — panel fixes committed plain at the tip (**Panel re-runs**,
    `skills/flow/review-panel-fix-round.md`). A guard listing every non-chore commit with no
    `Task-Id:` flags those sanctioned routes on every run that takes them, and kan-982's two listed
    commits are squashed out of every branch, so whether they were stray edits or full-suite fixes
    cannot be told. The defect as stated does not hold; a guard that tells a stray commit from a
    sanctioned one needs the sanctioned routes to mark their commits first.
  - 7 (dropped): **Commit scopes name the module** (`rules/commit-scope-is-the-module.mdc`) allows
    any area name by design — "There is no vocabulary of legal module names to maintain; the rule
    refuses three known-wrong shapes and accepts everything else." A spike matching each scope,
    case-folded, against the whole segments and `-`/`_`/`.` tokens of the paths its commit touched
    left 35 of 90 recent scoped gymie-playwright commits and 23 of 159 gymie-frontend ones unmatched,
    many of them legitimate area scopes (`test(visual)` on a snapshot capture, `test(e2e)`), and 7 of
    479 in this repository.
    <!-- measured: python3 -I /tmp/spk/scope.py <repo> 600 (machine-local spike script, not re-runnable from the tree) @ each repository's HEAD on 2026-10-15 -->
    The refusal would misfire on a flat repository's every area scope.
  - 8 (kept, Task 1): `ApplyDispatchRecord` and `ApplyDispatchEnd` (`stats/internal/api/records.go`)
    already keep a supplied `StartedAt`/`EndedAt` — the journal-replay override — and stamp only a
    zero one, but `runRecordDispatchBegin`/`runRecordDispatchEnd` (`stats/cmd/flow/record.go`) expose
    no flag to supply one.
- Every task's code and prose was spiked together in a scratch clone at `bc44caac`: each task's diff
  applies in plan order, `go vet` and `gofmt -l` are clean, `go test ./internal/guard/ ./cmd/flow/`
  passes, each new test fails with its production change reverted, and `check-verbatim-moves.sh`
  (with this change's `verbatim-moves.txt`), `check-references.sh`, `check-markdown-integrity.py`,
  `check-dispatch-paragraphs.sh`, `check-installed-citations.sh`, `check-normative-inventory.sh`,
  `check-guard-symlinks.sh` and `check-stage-mark-calls.sh` exit 0.

## Decisions

### Back-fill instants are CLI flags over the existing replay override

**ID:** backfill-instant-flags
**Status:** active
**Chosen:** `-started-at` on `flow record dispatch begin` and `-ended-at` on `end`, each RFC 3339,
sent as the row's `startedAt`/`endedAt`; omitted, the daemon stamps as today. An unparseable value
is exit 2 before the store is called. The daemon needs no change: it already keeps a supplied
instant exactly as it keeps a journal replay's.
**Considered:** the store deriving the start from the row's commit date, the finding's alternative —
a begin carries no commit, and a commit date is when a dispatch ended, not when it began; refusing a
future instant — no defect seen needs it.

### close-task refuses an undispatched appended task

**ID:** close-task-refuses-undispatched
**Status:** active
**Chosen:** an appended task no dispatch row names is exit 1 with `close-task: RECORD —`, after the
commit guards and before the review gate, nothing ticked or pushed; the per-task line names the
record call and the back-fill flags, and the parent records the row and re-runs without the
`-end-*` flags. A store that cannot be read still refuses nothing.
**Considered:** keeping the warning — kan-982 ticked three tasks past it; folding the refusal into
the commit guards' exit 1 — its remedy is a re-commit, which cannot repair a missing record.

### close-task warns on an undispatched appended task

**ID:** close-task-warns-undispatched
**Status:** superseded by close-task-refuses-undispatched
**Chosen:** recorded in the archived `self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a`
change's `design.md`, unchanged there.
**Considered:** recorded there.
**Superseded because:** kan-982 ticked Tasks 19–21 past the warning, and the gap surfaced only at
integrate's unfinished-work gate.

### The canonical worktree joins integrate's visual-verify check

**ID:** visual-dispatched-canonical
**Status:** active
**Chosen:** `check-visual-verify-dispatched.sh` takes an optional fourth argument, the canonical
worktree. When it differs from `<worktree>` and the worktree's own trigger is not 0, the diff's paths
— prefixed with the worktree's repository name, its main checkout's basename read from
`git rev-parse --git-common-dir` — are matched against the canonical worktree's `ui paths`, and the
better answer stands, 0 over 1 over 2; every store call anchors at the canonical worktree.
`integrate-change.sh` passes `$CANON`, and **Run 1** of the finish contract names the argument.
**Considered:** a `--repo <name>` argument from integrate — integrate would derive the name the guard
can read itself, and the contract's call would grow two arguments where one does; matching against
the canonical project only when the worktree's own is not configured — misses a frontend whose own
project declares a section that does not name the files.

### Load flakes run the full list twice more before their task closes

**ID:** load-flake-two-more-runs
**Status:** active
**Chosen:** in **The loop** step 2, a task fixing a test that passes alone and fails only in the
full suite closes after the parent runs the worktree's whole `## test` list twice more; every further
failure is fixed in the same task, a further commit with the same `Task-Id:`. The parent runs it,
since **The parent runs the full suite — never an implementer** (`skills/flow/implement.md`).
**Considered:** the implementer running the suite — contradicts that rule; a run count sized to the
flake rate — the finding names a fixed "≥2 more", and a load flake's rate is what the extra runs
measure.

### A non-reproducing determinism fix is dropped

**ID:** drop-unreproduced-determinism-fix
**Status:** active
**Chosen:** a fix whose flake shows no failure without it across the PROOF RUNS sizing is dropped;
the task closes with no code change and reports the measured rate.
**Considered:** committing it with the rate recorded — the kan-982 Task 17 outcome the finding
names, a change with no observed defect behind it.

### Motion frames carry their offset; triggers compare at one sampling

**ID:** motion-frame-offsets
**Status:** active
**Chosen:** frames are written `<nnn>-<ms>.png`, `<ms>` the millisecond offset from the trigger, and
a difference between two triggers is reported only after the sibling is recorded in the same
session by the same capture method and the strips are compared by offset.
**Considered:** a timestamp sidecar per strip — a second file for what one name carries; a
minimum sampling rate — fixes nothing when the two strips are captured by different methods.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — Task 2 changes an exit code the parent acts on, and Task 3 changes which
project a guard's store calls read and what a satellite diff matches; both alter run behaviour a
low-effort read could pass over.
**Considered:** `low` — three of the six tasks are one prose paragraph each, but the two guard tasks
carry the review risk.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** three groups, each one fixer in plan order — g1: Tasks 1–2 (the record flags, then the
close-task refusal whose message and `implement.md` clause name them); g2: Task 3 (the visual-verify
guard, its shim, integrate and the finish contract); g3: Tasks 4–6 (the prose of the fix loop and
the verifier; Tasks 4 and 5 share `skills/flow/verify-fix-loop.md`).
**Considered:** one group — four Go-test tasks and three prose tasks in one fixer context; two groups
merging g1 and g2 — no file shared, and g2 alone carries the largest diff.

## Open questions

## Live check

None — Task 1's flags reach a store path the daemon already serves (the journal replay's supplied
instants), and its CLI test pins the wire body; exercising it against the running flowd would write a
fabricated dispatch row into the operator's live stats store. Tasks 2 and 3 change guards whose Go
tests drive them end to end, and Tasks 4–6 change prose.
