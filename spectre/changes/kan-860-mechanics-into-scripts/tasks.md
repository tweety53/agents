# kan-860-mechanics-into-scripts

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — Tasks 18 and 19 move the archive-commit and worktree-cleanup recipes verbatim into `skills/flow-contracts/finish-hand-fallbacks.md`

`design.md` is canonical for the scope decisions; each task's row design is the per-group note it
cites (`design-<group>.md`, this directory). Tasks cite both, never restate them.

**Conventions every Go task follows** (from `.flow/project.md` and the notes):

- New guard logic is Go in `flow-guard` (`stats/internal/guard/<name>.go`, `Registry["<name>"]` in
  its own `init`), with a `scripts/<name>.sh` shim that ends `flow_guard_exec <name> 2
  "<name>:" "$@"` (the `scripts/check-base-moved.sh` shape), and a relative symlink
  `skills/<skill>/scripts/<name>.sh -> ../../../scripts/<name>.sh` in every skill whose Markdown
  names the guard.
- A shim that needs a sibling script names it as `$SCRIPT_DIR/<sibling>` and gains an entry in
  `TestShimSiblingsDeclared` (`stats/internal/guard/check_guard_symlinks_test.go`).
- A new `flow` verb lives in `stats/cmd/flow/` with its usage line in `main.go`.
- Git fixtures in tests set `user.name`/`user.email` in the fixture repo.
- A bash-parity test reads the retired bash from `ae805186` (**Decision:** parity-ref-on-main),
  the `bashAtBase` shape in `stats/internal/guard/mutate_and_verify_test.go`.
- The prompt keeps one call line plus the exit contract. Every run-loaded sentence the task
  removes or rewords is listed, exactly as `scripts/check-verbatim-moves.sh` prints it after `::`,
  in `spectre/changes/kan-860-mechanics-into-scripts/verbatim-moves.txt` — edited in the worktree and
  **left uncommitted**: it is a planning path, carried by the next `chore(spectre): plan` commit
  (**Planning commits**, `skills/flow-contracts/git-boundaries.md`), never by the task commit.

**Every task's verify step** is its own lint lines from `.flow/project.md` `## lint` that its
`**Files:**` need — Go: `cd stats && gofmt -l . && go vet ./...`; Markdown or shims:
`scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
`scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh`,
`scripts/check-verbatim-moves.sh`, `scripts/check-dispatch-paragraphs.sh`,
`scripts/check-stage-mark-calls.sh` — plus `cd stats && go test <pkg> -run '<the task's Tests>'
-count=1`.

**Baseline, measured before any edit** (`67a082a6`):

- Test counts per file: `plan_class_test.go` 1, `check_task_commit_fields_test.go` 3,
  `check_visual_trigger_test.go` 1, `check_stage_mark_calls_test.go` 1, `check_base_moved_test.go`
  1, `check_cleanup_complete_test.go` 2, `check_guard_symlinks_test.go` 4 (guard package);
  `state_test.go` 53, `record_test.go` 91, `stage_test.go` 30 (`stats/cmd/flow`);
  `watcher_test.go` 46 (`stats/internal/harvest`).
  <!-- measured: grep -c '^func Test' <each file> @ 67a082a6 -->
- `guard-autosquash.sh`, `project-get.sh` and `aside-planning-artifacts.sh` are bash;
  `scripts/test-project-get.sh` and `scripts/test-aside-planning-artifacts.sh` exist.
  <!-- measured: ls scripts/ ; grep -l 'Registry\["guard-autosquash"\]' stats/internal/guard/*.go (none) @ 67a082a6 -->
- The task-close recipe is at `skills/flow/implement.md:582`; the add-worktree read-merge-write
  sentence at `skills/flow/implement.md:250`.
  <!-- measured: grep -n "One Bash call: the implementer's\|Read the current record with" skills/flow/implement.md @ 67a082a6 -->

## Review Focus

- `plan-class.sh` keeps its first three lines byte for byte; `-class` two steps up or any step
  down exits 2 — Task 1 `TestPlanClassOverrideFlag`.
- `flow decision render`'s golden cases reproduce every rule of the worked `## Decision` format
  it replaces — micro, `default` panel, string implementer, free grouping, split suffix,
  bundle-cap suffix — Task 2 `TestDecisionRenderGolden`.
- `flow state add-worktree` never changes `state` or drops a peer worktree, store up or down —
  Task 3.
- `close-task.sh` never pushes when either commit guard refuses (**Decision:**
  close-task-push-after-guards) — Task 6 `TestCloseTaskOrder`.
- `fold-fixup.sh` leaves a conflict in progress (exit 3) rather than resolving it — Task 8.
- `remove-change-worktrees.sh` removes nothing on its first call when anything is unclassified or
  a wave-group copy exists, and runs no stop command a project does not declare in a fence —
  Task 19.
- `sync-panel-base.sh` passes no aside (behaviour parity with the panel prose) — Task 24.
- Every retired bash script's Go port is byte-identical on stdout, stderr and exit code —
  Tasks 16 and 21.

**Live verification:** Task 26 exercises the new `flow` verbs against the worktree's own isolated
store and the new guards against scratch fixtures, recording before/after.

---

- [x] 1. plan-class prints the tree defaults and takes a class override

Rows BP-21/BP-22 — `design-planning.md` § "BP-21 + BP-22".

  - [x] **Step 1: Write the failing tests** in `stats/internal/guard/plan_class_test.go`:
    `TestPlanClassTree` (table over every class × compact/bundle/experimental roll side × candidate
    state: absent dir, empty dir, two candidates; asserts the appended `tree:`, `panel:`,
    `grouping:`, `experimental:` lines, and a prefix compare that the first three lines are
    unchanged) and `TestPlanClassOverrideFlag` (`-class` equal to mechanical and one step up
    accepted; two steps up, any step down, an unknown class → exit 2).
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestPlanClass' -count=1`.
  - [x] **Step 3: Implement** in `stats/internal/guard/planclass.go` (pure `pcTree` plus the
    `-class` flag, experimental dir under `FLOW_GUARD_REPO_ROOT`); export `FLOW_GUARD_REPO_ROOT` in
    `scripts/plan-class.sh` and update its usage/exit header; in `skills/flow/brainstorm-planner.md`
    **Decide**, replace the tree-derivation prose the script now prints with the call line, keeping
    the "raise one step, never lower" reason requirement and the micro-row sentence.
  - [x] **Step 4: Verify** — the targeted run passes; lint lines per the header; verbatim-moves
    lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/planclass.go`, `stats/internal/guard/plan_class_test.go`,
`scripts/plan-class.sh`, `skills/flow/brainstorm-planner.md`, `skills/flow-fast/SKILL.md`
**Tests:** `TestPlanClassTree`, `TestPlanClassOverrideFlag`
**Regression:** reverting drops the tree lines and the flag: `TestPlanClassTree` finds no `tree:`
line and `-class` is an unknown argument.
**Baseline:** before=1 after=3
<!-- measured: cat stats/internal/guard/plan_class_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): plan-class prints the decision tree and takes -class`
**After:** none
**Build:** green

Correction (2026-09-30, panel round 1, F10): `plan-class.sh` and `sync-onto-base.sh` now take the agents-repo root from `flow_guard_root` in `scripts/lib/flow-guard.sh`, with `TestFlowGuardRootThroughSkillSymlink` added — plain commits cfeb543f and b1d927b4 on top, not folded, so the fields stay as this task's commit measured them.

Correction (2026-09-30): the plan declared four files; `skills/flow-fast/SKILL.md:82` still cited "its tree table", which this task removes from `brainstorm-planner.md`, so the line now names "the tree `plan-class.sh` prints" and the file joins `**Files:**` — reported by the group-1 implementer, applied by the parent at pick.

**Decision:** all-designed-rows-in-scope

- [x] 2. flow decision render prints the Decision block

Row BP-23 — `design-planning.md` § "BP-23".

  - [x] **Step 1: Write the failing tests** in new `stats/cmd/flow/decision_test.go`:
    `TestDecisionRenderGolden` (golden cases taken from the worked format in
    `brainstorm-planner.md` **Decide**: micro, small inline, regular free grouping, big sdd with a
    split group, experimental skipped for the bundle cap, `default` panel) and
    `TestDecisionRenderRefusals` (usage, unreadable file, missing required field → exit 2).
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./cmd/flow -run 'TestDecisionRender' -count=1`.
  - [x] **Step 3: Implement** `stats/cmd/flow/decision.go` and the `decision` command plus usage
    line in `stats/cmd/flow/main.go`; in `skills/flow/brainstorm-planner.md` **Decide** replace
    the two-table format rules and the preamble lines with the `flow decision render` call line and
    "print its output verbatim"; in `skills/flow/brainstorm.md` keep the record sequence and point
    its "print the `## Decision` block" sentence at the render call.
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/cmd/flow/decision.go`, `stats/cmd/flow/decision_test.go`,
`stats/cmd/flow/main.go`, `skills/flow/brainstorm-planner.md`, `skills/flow/brainstorm.md`
**Tests:** `TestDecisionRenderGolden`, `TestDecisionRenderRefusals`
**Regression:** reverting removes the verb: both tests fail to find the `decision` command.
**Baseline:** before=0 after=2
<!-- measured: cat stats/cmd/flow/decision_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(stats): flow decision render prints the Decision block`
**After:** Task 1
**Build:** green

**Decision:** all-designed-rows-in-scope

- [x] 3. flow state add-worktree merges one worktree into the record

Row MX4 — `design-planning.md` § "MX4".

  - [x] **Step 1: Write the failing tests** in `stats/cmd/flow/state_test.go`, reusing its
    `genuineDaemon`/`deadPortAddr`/`isolatedStateRoot` fakes: `TestStateAddWorktreeMergesEntry`,
    `TestStateAddWorktreePreservesPeersAndState`, `TestStateAddWorktreeFallback` (store down: read
    and write through the on-disk path) and `TestStateAddWorktreeRefusals` (relative path,
    non-40-hex sha → exit 2; no record, synthetic-only record → exit 1; nothing written).
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./cmd/flow -run 'TestStateAddWorktree' -count=1`.
  - [x] **Step 3: Implement** in `stats/cmd/flow/state.go` (factor `state set`'s stamp → validate
    → put → journal-fallback tail into a helper both use) plus the usage line in
    `stats/cmd/flow/main.go`; replace the "Read the current record with `flow state get`, merge in
    this…" sentence at `skills/flow/implement.md:250` and its twin in
    `skills/flow/cross-repo-worktrees.md` with the call line, keeping the timing sentence.
  - [x] **Step 4: Verify** — targeted run plus `go test ./cmd/flow -run 'TestState' -count=1` (the
    refactored tail); lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/cmd/flow/state.go`, `stats/cmd/flow/state_test.go`, `stats/cmd/flow/main.go`,
`skills/flow/implement.md`, `skills/flow/cross-repo-worktrees.md`
**Tests:** `TestStateAddWorktreeMergesEntry`, `TestStateAddWorktreePreservesPeersAndState`,
`TestStateAddWorktreeFallback`, `TestStateAddWorktreeRefusals`
**Regression:** reverting removes the verb: every test fails on an unknown `state` subcommand.
**Baseline:** before=53 after=57
<!-- measured: cat stats/cmd/flow/state_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(stats): flow state add-worktree merges one worktree into the record`
**After:** Task 2
**Build:** green

**Decision:** all-designed-rows-in-scope

- [x] 4. kickoff-worktree.sh runs the five kickoff steps

Row BR-M1 — `design-planning.md` § "BR-M1", including its cross-file impact on `/flow-plan`.

  - [x] **Step 1: Write the failing test** `TestKickoffWorktree` in new
    `stats/internal/guard/kickoff_worktree_test.go` (fixture repo with a local bare `origin`: fresh
    branch, existing remote branch, `.worktrees` not ignored → appended to `info/exclude`, a
    `## worktree setup` command failing → exit 1 naming it with the worktree already persisted,
    `flow` absent from PATH → exit 2 before any add, location guard refusal → exit 1); add
    `kickoff-worktree.sh` → `{"lib", "check-worktree-location.sh", "project-get.sh"}` to
    `TestShimSiblingsDeclared`.
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestKickoffWorktree|TestShimSiblingsDeclared' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/kickoffworktree.go` (step 3 calls `flow state
    add-worktree`), shim `scripts/kickoff-worktree.sh`, symlinks
    `skills/flow/scripts/kickoff-worktree.sh` and `skills/flow-plan/scripts/kickoff-worktree.sh`;
    in `skills/flow/brainstorm.md` **A**, steps 1–5 become the call line plus exit contract,
    keeping the phrase "kickoff steps 1–5", the stop rule and the persist-before-return sentence;
    add the guard to `skills/flow-plan/SKILL.md`'s guard list.
  - [x] **Step 4: Verify** — targeted run; lint lines (incl. `check-guard-symlinks.sh`);
    verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/kickoffworktree.go`, `stats/internal/guard/kickoff_worktree_test.go`,
`stats/internal/guard/check_guard_symlinks_test.go`, `scripts/kickoff-worktree.sh`,
`skills/flow/scripts/kickoff-worktree.sh`, `skills/flow-plan/scripts/kickoff-worktree.sh`,
`skills/flow/brainstorm.md`, `skills/flow-plan/SKILL.md`
**Tests:** `TestKickoffWorktree`
**Regression:** reverting removes the guard: `TestKickoffWorktree` finds no `kickoff-worktree` in
the registry, and the sibling test finds no shim.
**Baseline:** before=4 after=5
<!-- measured: cat stats/internal/guard/kickoff_worktree_test.go stats/internal/guard/check_guard_symlinks_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): kickoff-worktree.sh runs the five kickoff steps`
**After:** Task 2, 3
**Build:** green

**Decision:** all-designed-rows-in-scope

- [x] 5. check-review-gate.sh judges the per-task review gate

Row MX1 — `design-implement.md` item 1.

  - [x] **Step 1: Write the failing test** `TestCheckReviewGate` in new
    `stats/internal/guard/check_review_gate_test.go` (≤40 lines all declared → `QUIET`; >40 lines;
    <!-- measured: grep -n 'more than 40 lines' skills/flow/implement.md → line 643 @ 67a082a6 -->
    an undeclared path; a refused path that a `Files:` widening would otherwise declare; the map
    form summed across worktrees; a binary `-` counted as 0; unreadable plan → exit 2).
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestCheckReviewGate' -count=1`.
  - [x] **Step 3: Implement** — factor the plan resolution out of `checkTaskCommitFields` in
    `stats/internal/guard/taskcommitfields.go` into one helper both call; add
    `stats/internal/guard/reviewgate.go`, shim `scripts/check-review-gate.sh`, symlink
    `skills/flow/scripts/check-review-gate.sh`; replace the "**The review gate.** After the guard
    passes a task's commit…" recipe in `skills/flow/implement.md` with the call line plus exit
    contract; drop the "one site to re-tune" wording from `scripts/check-dispatch-paragraphs.sh`'s
    header.
  - [x] **Step 4: Verify** — targeted run plus `go test ./internal/guard -run 'TestCheckTaskCommitFields' -count=1`
    (the factored helper); lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/reviewgate.go`, `stats/internal/guard/check_review_gate_test.go`,
`stats/internal/guard/taskcommitfields.go`, `scripts/check-review-gate.sh`,
`skills/flow/scripts/check-review-gate.sh`, `skills/flow/implement.md`,
`scripts/check-dispatch-paragraphs.sh`
**Allowed-collateral:** `stats/internal/guard/dispatchparagraphs.go`
**Tests:** `TestCheckReviewGate`
**Regression:** reverting removes the guard: `TestCheckReviewGate` finds no `check-review-gate`.
**Baseline:** before=3 after=4
<!-- measured: cat stats/internal/guard/check_review_gate_test.go stats/internal/guard/check_task_commit_fields_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): check-review-gate.sh judges the per-task review gate`
**After:** Task 3
**Build:** green

**Decision:** all-designed-rows-in-scope

- [x] 6. close-task.sh runs the task-close sequence

Row MX3 — `design-implement.md` item 2.

  - [x] **Step 1: Write the failing tests** in new `stats/internal/guard/close_task_test.go`:
    `TestCloseTaskOrder` (a fields refusal → exit 1, no tick, no push; a planning-paths refusal →
    exit 1, no push; clean → gate, tick of QUIET tasks only, one push per distinct worktree, in that
    order) and `TestCloseTaskExitContract` (any guard exit 2 → exit 2; a failed tick or push → exit
    2; single and map task forms). Record the commands run through a fake `flow`/`git` on PATH.
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestCloseTask' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/closetask.go` (composes guards in-process via
    `Registry`), shim `scripts/close-task.sh`, symlink `skills/flow/scripts/close-task.sh`; replace
    step 2 of "**The next implementer overlaps the guard.**" in `skills/flow/implement.md` with the
    call line, keeping the exit-2-stops / exit-1-recommit contract.
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`,
`scripts/close-task.sh`, `skills/flow/scripts/close-task.sh`, `skills/flow/implement.md`
**Tests:** `TestCloseTaskOrder`, `TestCloseTaskExitContract`
**Regression:** reverting removes the guard: both tests find no `close-task` in the registry.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/close_task_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): close-task.sh runs the task-close sequence`
**After:** Task 3, 5
**Build:** green

Correction (2026-09-30, panel round 1, F1/F6/F2): the `skills/flow/scripts/check-task-commit-fields.sh` symlink is restored (the fix round still cites it), `check-guard-symlinks` rule 2 classifies a backtick span that wraps a line (`guardsymlinks.go`, `TestCheckGuardSymlinks` case 3m), and `skills/flow-fast/SKILL.md` states its sdd task close without `close-task.sh` — plain commits a64338db, 9b627111 and 1f924109 on top, not folded.

**Decision:** close-task-push-after-guards

- [x] 7. throwaway-worktree.sh creates and removes a throwaway copy

Row MX2/OS05 — `design-implement.md` item 3.

  - [x] **Step 1: Write the failing tests** in new `stats/internal/guard/throwaway_worktree_test.go`:
    `TestThrowawayWorktreeCreate` (dirty tree incl. a rename and an untracked file copied; `--sdd`
    copies `.superpowers/sdd`) and `TestThrowawayWorktreeRemoveFoldBack` (`--fold-back` copies only
    newer files; remove leaves no worktree), each asserting parity with the fence it replaces.
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestThrowawayWorktree' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/throwawayworktree.go`, shim
    `scripts/throwaway-worktree.sh`, symlink `skills/flow/scripts/throwaway-worktree.sh`; replace
    the create fence and the `worktree remove --force` sentence in `skills/flow/sdd-dispatch.md`
    and the "## The throwaway worktree" fences in `skills/flow/review-panel-optional-slots.md`
    with call lines.
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/throwawayworktree.go`,
`stats/internal/guard/throwaway_worktree_test.go`, `scripts/throwaway-worktree.sh`,
`skills/flow/scripts/throwaway-worktree.sh`, `skills/flow/sdd-dispatch.md`,
`skills/flow/review-panel-optional-slots.md`
**Tests:** `TestThrowawayWorktreeCreate`, `TestThrowawayWorktreeRemoveFoldBack`
**Regression:** reverting removes the guard: both tests find no `throwaway-worktree`.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/throwaway_worktree_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): throwaway-worktree.sh creates and removes a throwaway copy`
**After:** none
**Build:** green

**Decision:** all-designed-rows-in-scope

- [x] 8. fold-fixup.sh folds a fixup into its task commit

Row MX7 — `design-implement.md` item 4. `guard-autosquash.sh` stays bash and is a sibling.

  - [x] **Step 1: Write the failing tests** in new `stats/internal/guard/fold_fixup_test.go`:
    `TestFoldFixupFolds`, `TestFoldFixupEmptyDrop` (tip and non-tip emptied commit),
    `TestFoldFixupConflictLeftInProgress` (exit 3, rebase still in progress, `--finish` completes
    after a hand resolution) and `TestFoldFixupGuardRefusal` (a `guard-autosquash` refusal → exit
    1); add `fold-fixup.sh` → `{"lib", "guard-autosquash.sh"}` to `TestShimSiblingsDeclared`.
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestFoldFixup|TestShimSiblingsDeclared' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/foldfixup.go` (`GIT_SEQUENCE_EDITOR=: git
    rebase -i --autosquash`), shim `scripts/fold-fixup.sh`, symlink
    `skills/flow/scripts/fold-fixup.sh`; in `skills/flow/review-panel-fix-round.md` replace the
    fold recipe from "**Rewrite-based folding is for unpushed history only**" through "**A fixup
    whose fold empties its target commit is dropped…**" with the call line plus exit contract,
    keeping the unpushed-history rule and "clean autosquash is not evidence".
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/foldfixup.go`, `stats/internal/guard/fold_fixup_test.go`,
`stats/internal/guard/check_guard_symlinks_test.go`, `scripts/fold-fixup.sh`,
`skills/flow/scripts/fold-fixup.sh`, `skills/flow/review-panel-fix-round.md`,
`skills/flow/scripts/aside-planning-artifacts.sh`
**Tests:** `TestFoldFixupFolds`, `TestFoldFixupEmptyDrop`, `TestFoldFixupConflictLeftInProgress`,
`TestFoldFixupGuardRefusal`
**Regression:** reverting removes the guard: every test finds no `fold-fixup`.
**Baseline:** before=4 after=8
<!-- measured: cat stats/internal/guard/fold_fixup_test.go stats/internal/guard/check_guard_symlinks_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): fold-fixup.sh folds a fixup into its task commit`
**After:** Task 4
**Build:** green

Correction (2026-09-30): the plan declared the empty-fold drop as `reset --hard` at the tip, `rebase --onto` below it; on git 2.50.1 a fixup that empties its target stops the rebase at the fixup instead, so the drop is `git reset --soft HEAD^` then `git rebase --continue` at that stop, and it runs before the planning paths are restored (a reset after the restore would wipe them). The fold recipe was `skills/flow`'s last citation of `aside-planning-artifacts.sh`, so its symlink there was deleted and joins `**Files:**`. Reported by the group-3 implementer, transcribed by the parent.

**Decision:** all-designed-rows-in-scope

- [x] 9. write-panel-diff.sh writes every panel diff and its touched list

Row RP35 — `design-panel.md` § "RP35".

  - [x] **Step 1: Write the failing tests** in new `stats/internal/guard/write_panel_diff_test.go`:
    `TestWritePanelDiffParity` (each of `final`, `slot-delta` held and `-`, `late-fix`,
    `fix-round`, over one and two worktrees: byte-equal to the hand recipe's output, `.touched`
    equal to `git diff --name-status` sections) and `TestWritePanelDiffRefusals` (usage, a pair
    failing `panelValidateWorktree`, a git failure → exit 2, nothing written).
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestWritePanelDiff' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/writepaneldiff.go` (reusing
    `paneltouchedpaths.go`), shim `scripts/write-panel-diff.sh`, symlink
    `skills/flow/scripts/write-panel-diff.sh`; replace the diff-writing recipes in
    `skills/flow/review-panel.md` and the `fix-round-N.diff` recipe in
    `skills/flow/review-panel-fix-round.md` with call lines.
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/writepaneldiff.go`, `stats/internal/guard/write_panel_diff_test.go`,
`scripts/write-panel-diff.sh`, `skills/flow/scripts/write-panel-diff.sh`,
`skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`,
`skills/flow/review-panel-late-fix.md`
**Tests:** `TestWritePanelDiffParity`, `TestWritePanelDiffRefusals`
**Regression:** reverting removes the guard: both tests find no `write-panel-diff`.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/write_panel_diff_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): write-panel-diff.sh writes every panel diff and its touched list`
**After:** Task 8
**Build:** green

Correction (2026-09-30): the `late-fix.diff` recipe lives in `skills/flow/review-panel-late-fix.md` after KAN-857, so that file joins `**Files:**`. Reported by the group-3 implementer, transcribed by the parent.

**Decision:** all-designed-rows-in-scope

- [x] 10. check-late-fix-trigger.sh decides the late-fix reduction

`design-panel.md` § "Late-fix trigger".

  - [x] **Step 1: Write the failing test** `TestCheckLateFixTrigger` in new
    `stats/internal/guard/check_late_fix_trigger_test.go` — one case per condition 1–5 failing
    alone (exit 1, its line printed), all holding (exit 0, the `late-fix reduction:` line), a
    binary numstat entry, an untracked `tasks.md`, `-` as the since-close sha, and a cannot-answer
    (exit 2).
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestCheckLateFixTrigger' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/latefixtrigger.go` (calling
    `checkPanelFindingsClosed` in-process), shim `scripts/check-late-fix-trigger.sh`, symlink
    `skills/flow/scripts/check-late-fix-trigger.sh`; replace the trigger conditions under **The
    late-fix reduction** in `skills/flow/review-panel.md` with the call line plus exit contract
    (exit 2 read as full path).
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/latefixtrigger.go`,
`stats/internal/guard/check_late_fix_trigger_test.go`, `scripts/check-late-fix-trigger.sh`,
`skills/flow/scripts/check-late-fix-trigger.sh`, `skills/flow/review-panel-late-fix.md`
**Tests:** `TestCheckLateFixTrigger`
**Regression:** reverting removes the guard: the test finds no `check-late-fix-trigger`.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_late_fix_trigger_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): check-late-fix-trigger.sh decides the late-fix reduction`
**After:** Task 9
**Build:** green

Correction (2026-09-30): the five late-fix conditions live in `skills/flow/review-panel-late-fix.md`, not `skills/flow/review-panel.md`, which this task leaves unchanged; `**Files:**` names the file the commit edits. Reported by the group-3 implementer, transcribed by the parent.

**Decision:** all-designed-rows-in-scope

- [x] 11. render-slot-prompt.sh renders a panel dispatch's brief

`design-panel.md` § "Renderer".

  - [x] **Step 1: Write the failing tests** in new `stats/internal/guard/render_slot_prompt_test.go`:
    `TestRenderSlotPrompt` (against the real `skills/flow/` tree: one slot and a `+`-bundle;
    each `-diff` kind; `-fix-report` adds FIX-ROUND SCOPE; `-no-bundle`; CITATION CHECK once per
    existing `-standard` file; every placeholder of `review-panel.md`'s table filled) and
    `TestRenderSlotPromptUnresolvedPlaceholder` (exit 1 naming it; `mutation` refused).
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestRenderSlotPrompt' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/renderslotprompt.go`, shim
    `scripts/render-slot-prompt.sh`, symlink `skills/flow/scripts/render-slot-prompt.sh`; in
    `skills/flow/review-panel.md` the dispatch recipe becomes the render call plus the items still
    typed in the Agent call (the note's list). The blocks stay in `review-panel.md`, so the
    `check-dispatch-paragraphs.sh` pins still hold.
  - [x] **Step 4: Verify** — targeted run; lint lines incl. `check-dispatch-paragraphs.sh`;
    verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/renderslotprompt.go`,
`stats/internal/guard/render_slot_prompt_test.go`, `scripts/render-slot-prompt.sh`,
`skills/flow/scripts/render-slot-prompt.sh`, `skills/flow/review-panel.md`
**Tests:** `TestRenderSlotPrompt`, `TestRenderSlotPromptUnresolvedPlaceholder`
**Regression:** reverting removes the guard: both tests find no `render-slot-prompt`.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/render_slot_prompt_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): render-slot-prompt.sh renders a panel dispatch's brief`
**After:** Task 9, 10
**Build:** green

Correction (2026-09-30): Step 1 says CITATION CHECK once per `-standard` file; the design's rule — once per worktree whose `citation-check.md` exists — is what shipped, `-standard` filling `[STANDARDS_PATHS]`. The first cut refused any slot list naming `mutation`, which dropped INDEPENDENT PASSES from the real `failure-modes+mutation` bundle; the review fix renders every other role and the shared paragraph and refuses only `mutation` alone. Reported by the group-3 implementer and the gated reviewer.

**Decision:** all-designed-rows-in-scope

- [x] 12. Cut the reviewer templates' Placeholders lists

`design-panel.md` § "Follow-up cut".

  - [x] **Step 1: Cut** the **Placeholders:** list from `skills/flow/primary-reviewer-prompt.md`,
    `skills/flow/principles-reviewer-prompt.md` and `skills/flow/failure-modes-reviewer-prompt.md`
    — the renderer fills from `review-panel.md`'s table.
  - [x] **Step 2: Verify** — `go test ./internal/guard -run 'TestRenderSlotPrompt' -count=1` still
    passes; `scripts/check-installed-citations.sh` coverage stays non-zero (primary keeps 3
    citations in its body); lint lines; verbatim-moves lines recorded.
  - [x] **Step 3: Commit.**

**Files:** `skills/flow/primary-reviewer-prompt.md`, `skills/flow/principles-reviewer-prompt.md`,
`skills/flow/failure-modes-reviewer-prompt.md`, `stats/internal/guard/references.go`,
`stats/internal/guard/installedcitations.go`
**Tests:** none — prose cut; Task 11's renderer test is the check
**Regression:** none — a revert restores lists the renderer ignores.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 67a082a6 -->
**Commit:** `refactor(flow): cut the reviewer templates' placeholder lists`
**After:** Task 11
**Build:** green

Correction (2026-09-30): Step 2 predicted primary keeps 3 citations in its body; measured, every citation `check-installed-citations.sh` and `check-references.sh` counted in all three templates sat inside the **Placeholders:** list, so each drops to 0 and both guards refuse an undeclared zero. The three templates are declared expected-zero in `crExpectedZero` (`references.go`) and `cicExpectedZero` (`installedcitations.go`), which join `**Files:**` — option (a) of the group-3 implementer's BLOCKED report, taken as the recommended answer under the operator's standing instruction; the parent implemented the task inline (a finished child is never resumed).

**Decision:** all-designed-rows-in-scope

- [x] 13. check-visual-trigger.sh names its exit-2 cause

Row VH-25 — `design-verify.md` § "VH-25".

  - [x] **Step 1: Write the failing test** `TestVisualTriggerExit2Tokens` in
    `stats/internal/guard/check_visual_trigger_test.go` (NOT-CONFIGURED for the two unconfigured
    cases, CANNOT-ANSWER for every other exit 2); update the existing table's exit-2 `err:` pins
    with the appended token line.
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestCheckVisualTrigger|TestVisualTriggerExit2Tokens' -count=1`.
  - [x] **Step 3: Implement** in `stats/internal/guard/visualtrigger.go`; add both tokens to
    `scripts/check-visual-trigger.sh`'s exit-contract header; in `skills/flow/verify-and-handoff.md`
    steps 1–2 become the one call, keeping the step numbering.
  - [x] **Step 4: Verify** — targeted run plus `go test ./internal/guard -run 'TestCheckVisualVerifyDispatched' -count=1`;
    lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/visualtrigger.go`, `stats/internal/guard/check_visual_trigger_test.go`,
`scripts/check-visual-trigger.sh`, `skills/flow/verify-and-handoff.md`
**Tests:** `TestVisualTriggerExit2Tokens`
**Regression:** reverting drops the token lines: the new test finds neither token.
**Baseline:** before=1 after=2
<!-- measured: cat stats/internal/guard/check_visual_trigger_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): check-visual-trigger.sh names its exit-2 cause`
**After:** none
**Build:** green

**Decision:** all-designed-rows-in-scope

- [x] 14. flow record handoff-lines prints the handoff's record lines

Row VH-24 — `design-verify.md` § "VH-24".

  - [x] **Step 1: Write the failing tests** in `stats/cmd/flow/record_test.go`:
    `TestRecordHandoffLinesFormat` (the three Records spellings, the Deferred filter — `deferred`
    prefix only — and `none`), `TestRecordHandoffLinesDedupByProject` (two worktrees of one project
    count once) and `TestRecordHandoffLinesUnknown` (a failed findings read → `**Deferred:** unknown
    — the findings could not be read`, exit 0).
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./cmd/flow -run 'TestRecordHandoffLines' -count=1`.
  - [x] **Step 3: Implement** in `stats/cmd/flow/record.go` (extract the helpers from
    `runRecordJournalCount`, `runRecordCostStatus`, `runRecordFindings` so both paths share them)
    plus the usage line in `stats/cmd/flow/main.go`; in `skills/flow/verify-and-handoff.md`,
    "**Produce the handoff's `Records:` count** …" through "… reads `none` when the count is `0`."
    becomes the call line plus "exits 0 always; render each line exactly as printed"; update the
    `journal-count` citation in `skills/flow-contracts/handoff-blocks.md`.
  - [x] **Step 4: Verify** — targeted run plus `go test ./cmd/flow -run 'TestRecord' -count=1`;
    lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/cmd/flow/main.go`,
`skills/flow/verify-and-handoff.md`, `skills/flow-contracts/handoff-blocks.md`
**Tests:** `TestRecordHandoffLinesFormat`, `TestRecordHandoffLinesDedupByProject`,
`TestRecordHandoffLinesUnknown`
**Regression:** reverting removes the verb: every test fails on an unknown `record` subcommand.
**Baseline:** before=91 after=94
<!-- measured: cat stats/cmd/flow/record_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(stats): flow record handoff-lines prints the handoff's record lines`
**After:** Task 2, 3, 13
**Build:** green

**Decision:** handoff-deferred-unknown

- [x] 15. check-visual-preflight.sh runs the four visual preflight checks

Row VV-18 — `design-verify.md` § "VV-18", all four checks.

  - [x] **Step 1: Write the failing tests** in new
    `stats/internal/guard/check_visual_preflight_test.go`, fake `lsof`/`node` injected through
    `Env.Getenv("PATH")`: `TestCheckVisualPreflightPorts` (free; held and answering; held and
    silent → FAIL; `lsof` missing → exit 2), `TestCheckVisualPreflightBaseURL` (`:<default>`
    matched, a bare number not matched, a line naming the Variable overridden, a `start` command
    naming it overridden, skipped dirs and binaries), `TestCheckVisualPreflightOrigins` (an
    `allowed_origins` list missing an app origin → FAIL; source files excluded) and
    `TestCheckVisualPreflightPlaywright` (resolved outside the worktree → FAIL; unresolvable →
    info line, pass). Add one case pinning this repository clean.
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestCheckVisualPreflight' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/visualpreflight.go`, shim
    `scripts/check-visual-preflight.sh`, symlink `skills/flow/scripts/check-visual-preflight.sh`;
    in `skills/flow/visual-verify.md` step 3's four bullets become the call line plus exit
    contract, keeping "Any failing check ends the stage here".
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/visualpreflight.go`,
`stats/internal/guard/check_visual_preflight_test.go`, `scripts/check-visual-preflight.sh`,
`skills/flow/scripts/check-visual-preflight.sh`, `skills/flow/visual-verify.md`
**Tests:** `TestCheckVisualPreflightPorts`, `TestCheckVisualPreflightBaseURL`,
`TestCheckVisualPreflightOrigins`, `TestCheckVisualPreflightPlaywright`
**Regression:** reverting removes the guard: every test finds no `check-visual-preflight`.
**Baseline:** before=0 after=4
<!-- measured: cat stats/internal/guard/check_visual_preflight_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): check-visual-preflight.sh runs the visual preflight checks`
**After:** none
**Build:** green

**Decision:** visual-preflight-all-checks

- [x] 16. Port project-get.sh to Go

Row MI1/MA2, first half — `design-finish.md` § "MI1/MA2" (the byte-for-byte port).

  - [x] **Step 1: Write the failing test** `TestProjectGetParity` in new
    `stats/internal/guard/project_get_test.go` — runs the bash at `ae805186` (`scripts/project-get.sh`
    with `lib/project-section.sh` and `lib/strip-bom.sh`) and the Go guard over the same fixtures
    (present, absent, duplicate heading, BOM, CRLF, no file) and compares stdout, stderr and exit.
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestProjectGetParity' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/projectget.go`; `scripts/project-get.sh`
    becomes a `flow_guard_exec project-get 2 …` shim. `scripts/lib/project-section.sh` stays (other
    bash sources it).
  - [x] **Step 4: Verify** — targeted run plus `scripts/test-project-get.sh`; lint lines.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/projectget.go`, `stats/internal/guard/project_get_test.go`,
`scripts/project-get.sh`
**Tests:** `TestProjectGetParity`
**Regression:** reverting restores the bash and removes the guard: the test finds no `project-get`.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/project_get_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `perf(guard): port project-get.sh to Go`
**After:** none
**Build:** green

**Decision:** parity-ref-on-main

- [x] 17. project-get.sh --enum resolves a single-line literal key

Row MI1/MA2, second half — `design-finish.md` § "MI1/MA2" (the enum shape).

  - [x] **Step 1: Write the failing test** `TestProjectGetEnum` in
    `stats/internal/guard/project_get_test.go` (a match → the literal, exit 0; backticked and padded
    heads; absent → exit 1; no match → exit 3 with one stderr line quoting the head; usage → 2).
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestProjectGetEnum' -count=1`.
  - [x] **Step 3: Implement** in `stats/internal/guard/projectget.go`; document `--enum` in
    `scripts/project-get.sh`'s header and in `skills/flow-contracts/project-configuration.md`'s
    "A single-line-literal key's value…"; replace the hand parse in `skills/flow/integrate.md`
    ("Run `project-get.sh <main-checkout> "default landing route"` … backticks removed …") and in
    `skills/flow-fast/SKILL.md` §6 and §4 with `--enum` calls.
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/projectget.go`, `stats/internal/guard/project_get_test.go`,
`scripts/project-get.sh`, `skills/flow-contracts/project-configuration.md`,
`skills/flow/integrate.md`, `skills/flow-fast/SKILL.md`
**Tests:** `TestProjectGetEnum`
**Regression:** reverting removes `--enum`: the test's calls exit 2 as an unknown argument.
**Baseline:** before=1 after=2
<!-- measured: cat stats/internal/guard/project_get_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): project-get.sh --enum resolves a single-line literal key`
**After:** Task 1, 16
**Build:** green

**Decision:** all-designed-rows-in-scope

- [x] 18. commit-archive.sh makes the archive commit

Row MA1 — `design-finish.md` § "MA1".

  - [x] **Step 1: Write the failing test** `TestCommitArchive` in new
    `stats/internal/guard/commit_archive_test.go` (`ARCHIVE-COMMITTED: <sha>`,
    `ARCHIVE-NOTHING-STAGED`, `ARCHIVE-WRONG-BRANCH: <found>` exit 1, a scope violation exit 1,
    cannot answer exit 2); add `commit-archive.sh` → `{"lib", "check-archive-scope.sh"}` to
    `TestShimSiblingsDeclared`.
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestCommitArchive|TestShimSiblingsDeclared' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/commitarchive.go`, shim
    `scripts/commit-archive.sh` (exports `FLOW_GUARD_SELF`, asserts
    `$SCRIPT_DIR/check-archive-scope.sh`), symlink `skills/flow/scripts/commit-archive.sh`; move
    `skills/flow/archive.md` step 4's recipe verbatim into
    `skills/flow-contracts/finish-hand-fallbacks.md` under `commit-archive.sh — Run 2, step 4`, and
    leave the call line plus exit contract in its place.
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/commitarchive.go`, `stats/internal/guard/commit_archive_test.go`,
`stats/internal/guard/check_guard_symlinks_test.go`, `scripts/commit-archive.sh`,
`skills/flow/scripts/commit-archive.sh`, `skills/flow/archive.md`,
`skills/flow-contracts/finish-hand-fallbacks.md`
**Tests:** `TestCommitArchive`
**Regression:** reverting removes the guard: the test finds no `commit-archive`.
**Baseline:** before=4 after=5
<!-- measured: cat stats/internal/guard/commit_archive_test.go stats/internal/guard/check_guard_symlinks_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): commit-archive.sh makes the archive commit`
**After:** Task 4, 8
**Build:** green

Correction (2026-09-30, panel round 1, F5): `scripts/generate-relocation-comparison.py` reads **Files:** in the inline form too (`scripts/test-generate-relocation-comparison.sh` case ii); for this plan it now exits 2 on the symlinked skill paths rather than writing nothing silently — plain commit 29961384 on top, not folded.

Correction (2026-09-30): the guard checks the branch before copying the ledger and panel record (the recipe copied first, so a wrong branch now leaves the landing worktree untouched), stops with exit 2 when an existing ledger or record cannot be copied (the recipe ignored the failed copy), and prints only its own verdict, passing on the scope guard's lines only when that guard refuses. Reported by the group-5 implementer.

**Decision:** all-designed-rows-in-scope

- [x] 19. remove-change-worktrees.sh runs the worktree cleanup

Row MR2 — `design-finish.md` § "MR2".

  - [x] **Step 1: Write the failing tests** in new
    `stats/internal/guard/remove_change_worktrees_test.go`: `TestRemoveChangeWorktrees` (clean
    removal with REMOVED and REMOTE-* lines; a remote ref already gone), 
    `TestRemoveChangeWorktreesDisclose` (an unclassified bucket or a wave-group copy → exit 3,
    DISCLOSE lines, nothing removed; `--proceed` then removes) and
    `TestRemoveChangeWorktreesStopCommand` (a `## stop` body with no fence → check skipped; a fenced
    command → run under the 60-second bound, nothing else run); add `remove-change-worktrees.sh` →
    `{"lib", "check-worktree-processes.sh"}` to `TestShimSiblingsDeclared`.
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestRemoveChangeWorktrees|TestShimSiblingsDeclared|TestCheckCleanupComplete' -count=1`.
  - [x] **Step 3: Implement** — factor `cleanupcomplete.go`'s porcelain loop into a shared helper;
    add `stats/internal/guard/removechangeworktrees.go`, shim `scripts/remove-change-worktrees.sh`,
    symlink `skills/flow/scripts/remove-change-worktrees.sh`; move the six-check recipe, the bucket
    list and the remove/remote blocks verbatim from `skills/flow-contracts/finish-contract-run2.md`
    (and `skills/flow/archive.md` where it carries them) into
    `skills/flow-contracts/finish-hand-fallbacks.md`, leaving the call lines, the
    gates-vs-disclosure rules, the ask and the relay.
  - [x] **Step 4: Verify** — targeted run (incl. `TestCheckCleanupComplete`, the refactored
    helper); lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/removechangeworktrees.go`,
`stats/internal/guard/remove_change_worktrees_test.go`, `stats/internal/guard/cleanupcomplete.go`,
`stats/internal/guard/check_guard_symlinks_test.go`, `scripts/remove-change-worktrees.sh`,
`skills/flow/scripts/remove-change-worktrees.sh`, `skills/flow-contracts/finish-contract-run2.md`,
`skills/flow/archive.md`, `skills/flow-contracts/finish-hand-fallbacks.md`
**Tests:** `TestRemoveChangeWorktrees`, `TestRemoveChangeWorktreesDisclose`,
`TestRemoveChangeWorktreesStopCommand`
**Regression:** reverting removes the guard: every test finds no `remove-change-worktrees`.
**Baseline:** before=6 after=9
<!-- measured: cat stats/internal/guard/remove_change_worktrees_test.go stats/internal/guard/check_guard_symlinks_test.go stats/internal/guard/check_cleanup_complete_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): remove-change-worktrees.sh runs the worktree cleanup`
**After:** Task 4, 8, 18
**Build:** green

Correction (2026-09-30, panel round 1, F3/F4/F7/F9): check 4 classifies an image by its location, check 5 reads `## stop` with the `## worktree setup` fence reader and prints `SKIPPED:` when no fence is declared, and `skills/flow-contracts/project-configuration.md`'s `## stop` row carries the fence shape run 2 now cites — plain commits 29632383 and c9cdaabf on top, not folded.

Correction (2026-09-30): the guard runs `worktree prune` before `branch -d`, so a leftover registration cannot block the delete; runs checks 1–4 before the disclosure stop, so the stop command never runs on an exit-3 call; treats a `## stop` section with no code fence as no command (check 5 skipped, per `design-finish.md`); and lists a screenshot in a declared screenshot-output directory as unclassified rather than regeneratable. Exit codes: 0 all passed (the remote-branch outcome reported, step 7 verifies it), 1 a check failed and nothing was removed or a removal failed, 2 cannot answer, 3 the disclosure stop. Reported by the group-5 implementer.

**Decision:** remove-change-worktrees-included

- [x] 20. flow stage mark marks several stages at once

Row FF1 — `design-finish.md` § "FF1".

  - [x] **Step 1: Write the failing tests**: in `stats/cmd/flow/stage_test.go`
    `TestStageMarkBeginsAndEndsEachKey`, `TestStageMarkValidatesAllKeysFirst` (one bad key → exit
    2, no store call) and `TestStageMarkJournalFallback`; in
    `stats/internal/harvest/watcher_test.go` `TestWatcherMatchesStageMark`; in
    `stats/internal/guard/check_stage_mark_calls_test.go` `TestStageMarkCallsChecksStageMark`
    (missing token, placeholder harness, an unserved key in `-stages`).
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./cmd/flow ./internal/harvest ./internal/guard -run 'TestStageMark|TestWatcherMatchesStageMark' -count=1`.
  - [x] **Step 3: Implement** `stage mark` in `stats/cmd/flow/stage.go` (usage line in
    `stats/cmd/flow/main.go`), extend `stageMarkInvocationPattern` in
    `stats/internal/harvest/watcher.go` and the detection in
    `stats/internal/guard/stagemarkcalls.go`; replace the empty begin/end pairs in
    `skills/flow-fast/SKILL.md` with `flow stage mark` lines, leaving its **Stage keys** table
    untouched.
  - [x] **Step 4: Verify** — targeted run plus `go test ./internal/guard -run 'TestStageKeysMatchFlowFastSkillTable' -count=1`;
    lint lines incl. `check-stage-mark-calls.sh`; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`, `stats/cmd/flow/main.go`,
`stats/internal/harvest/watcher.go`, `stats/internal/harvest/watcher_test.go`,
`stats/internal/guard/stagemarkcalls.go`, `stats/internal/guard/check_stage_mark_calls_test.go`,
`skills/flow-fast/SKILL.md`
**Tests:** `TestStageMarkBeginsAndEndsEachKey`, `TestStageMarkValidatesAllKeysFirst`,
`TestStageMarkJournalFallback`, `TestWatcherMatchesStageMark`, `TestStageMarkCallsChecksStageMark`
**Regression:** reverting removes the verb: the stage tests hit an unknown subcommand, the watcher
test's `stage mark` line binds no session, and the guard test's bad call passes.
**Baseline:** before=77 after=82
<!-- measured: cat stats/cmd/flow/stage_test.go stats/internal/harvest/watcher_test.go stats/internal/guard/check_stage_mark_calls_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(stats): flow stage mark marks several stages at once`
**After:** Task 1, 2, 3, 14, 17
**Build:** green

Correction (2026-09-30): Step 4's command names `./internal/guard`, where no such test exists; `TestStageKeysMatchFlowFastSkillTable` lives in `stats/internal/stages/names_test.go` — `go test ./internal/stages -run 'TestStageKeysMatchFlowFastSkillTable' -count=1`, measured passing. Reported by the group-9 implementer.

**Decision:** all-designed-rows-in-scope

- [x] 21. Port aside-planning-artifacts.sh to Go

`design-basesync.md` § "Planned shape" item 1.

  - [x] **Step 1: Write the failing test** `TestAsidePlanningArtifactsParity` in new
    `stats/internal/guard/aside_planning_artifacts_test.go` — the bash at `ae805186` (with
    `lib/spec-root.sh`) against the Go guard over the same fixtures (aside and restore; nothing to
    aside; both spec roots; not a dir; not git; unknown action), comparing stdout, stderr and exit.
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestAsidePlanningArtifactsParity' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/asideplanningartifacts.go`;
    `scripts/aside-planning-artifacts.sh` becomes a `flow_guard_exec` shim.
  - [x] **Step 4: Verify** — targeted run plus `scripts/test-aside-planning-artifacts.sh`; lint lines.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/asideplanningartifacts.go`,
`stats/internal/guard/aside_planning_artifacts_test.go`, `scripts/aside-planning-artifacts.sh`
**Tests:** `TestAsidePlanningArtifactsParity`
**Regression:** reverting restores the bash: the test finds no `aside-planning-artifacts` guard.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/aside_planning_artifacts_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `perf(guard): port aside-planning-artifacts.sh to Go`
**After:** none
**Build:** green

**Decision:** parity-ref-on-main

- [x] 22. A shared base-moved verdict and rebase-onto-tip core

`design-basesync.md` items 2–3.

  - [x] **Step 1: Write the failing tests** `TestBaseMovedFullOverlap` in
    `stats/internal/guard/check_base_moved_test.go` (the verdict carries the full sorted overlap
    while the printed line keeps its 10-path cut) and `TestRebaseOntoTip` in new
    `stats/internal/guard/base_rebase_test.go` (rebased; conflict with unmerged paths, aside kept;
    refused with no rebase in progress, aside restored; tip pinned to a sha before a concurrent
    fetch).
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestBaseMovedFullOverlap|TestRebaseOntoTip|TestCheckBaseMoved' -count=1`.
  - [x] **Step 3: Implement** — refactor `stats/internal/guard/basemoved.go` into `baseMoved(…)
    bmVerdict` with `checkBaseMoved` printing `v.line` unchanged; add
    `stats/internal/guard/baserebase.go` (`rebaseOntoTip`, aside in-process via Task 21's port).
  - [x] **Step 4: Verify** — targeted run (`TestCheckBaseMoved` unchanged); lint lines.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/basemoved.go`, `stats/internal/guard/check_base_moved_test.go`,
`stats/internal/guard/baserebase.go`, `stats/internal/guard/base_rebase_test.go`
**Tests:** `TestBaseMovedFullOverlap`, `TestRebaseOntoTip`
**Regression:** reverting removes `baseMoved`/`rebaseOntoTip`: neither test compiles against them.
**Baseline:** before=1 after=3
<!-- measured: cat stats/internal/guard/check_base_moved_test.go stats/internal/guard/base_rebase_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `refactor(guard): share the base-moved verdict and a rebase-onto-tip core`
**After:** Task 21
**Build:** green

Correction (2026-09-30, panel round 1, F8): one `inProgress` helper in `baserebase.go` replaces the rebase-in-progress checks in `foldfixup.go` and `syncontobase.go` — plain commit 16fe6ff9 on top, not folded.

**Decision:** all-designed-rows-in-scope

- [x] 23. sync-onto-base.sh syncs a finishing worktree onto its base

`design-basesync.md` item 4.

  - [x] **Step 1: Write the failing tests** in new `stats/internal/guard/sync_onto_base_test.go`:
    `TestSyncOntoBase` (CLEAR → `CLEAN:`; MOVED → `REBASED:` with `GUARD-TEST:`/`NO-GUARD-TEST:`
    lines, re-check loop capped at 3; conflict → exit 1 left mid-rebase; REFUSE → exit 2) and
    `TestSyncOntoBaseResume` (refuses mid-rebase; requires `<onto>` an ancestor of HEAD; restores
    the aside; no guard-test lines).
  - [x] **Step 2: Run them; they fail** — `cd stats && go test ./internal/guard -run 'TestSyncOntoBase' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/syncontobase.go`, shim
    `scripts/sync-onto-base.sh` (agents root exported as the note says), symlink
    `skills/flow/scripts/sync-onto-base.sh`; replace the recipe in `skills/flow/sync-onto-base.md`
    with the call lines plus exit contract, keeping the conflict-resolution rule.
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/syncontobase.go`, `stats/internal/guard/sync_onto_base_test.go`,
`scripts/sync-onto-base.sh`, `skills/flow/scripts/sync-onto-base.sh`, `skills/flow/sync-onto-base.md`
**Tests:** `TestSyncOntoBase`, `TestSyncOntoBaseResume`
**Regression:** reverting removes the guard: both tests find no `sync-onto-base`.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/sync_onto_base_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): sync-onto-base.sh syncs a finishing worktree onto its base`
**After:** Task 22
**Build:** green

**Decision:** all-designed-rows-in-scope

- [x] 24. sync-panel-base.sh brings the panel's worktree onto the base

Row RP34 — `design-basesync.md` item 5 (no aside).

  - [x] **Step 1: Write the failing test** `TestSyncPanelBase` in new
    `stats/internal/guard/sync_panel_base_test.go` (`BASE:` line; MOVED without overlap → rebased,
    `REBASED: … merge base <sha>`; MOVED with overlap → no rebase unless `--rebase`; a fresh overlap
    on re-check stops; conflict → exit 1; no aside taken).
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestSyncPanelBase' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/syncpanelbase.go`, shim
    `scripts/sync-panel-base.sh`, symlink `skills/flow/scripts/sync-panel-base.sh`; replace the
    base-movement recipe in `skills/flow/review-panel.md` with the call line, keeping the operator's
    **Rebase** prompt.
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/syncpanelbase.go`, `stats/internal/guard/sync_panel_base_test.go`,
`scripts/sync-panel-base.sh`, `skills/flow/scripts/sync-panel-base.sh`, `skills/flow/review-panel.md`, `skills/flow/archive.md`, `skills/flow/review-panel-fix-round.md`
**Tests:** `TestSyncPanelBase`
**Regression:** reverting removes the guard: the test finds no `sync-panel-base`.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/sync_panel_base_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): sync-panel-base.sh brings the panel worktree onto the base`
**After:** Task 8, 9, 10, 11, 18, 19, 22
**Build:** green

Correction (2026-09-30): removing the fenced call from `review-panel.md` left `skills/flow/scripts/resolve-base-branch.sh` with no citation `check-guard-symlinks.sh` counts (rule 6), though `archive.md` and three other files still use it in prose; `archive.md:30` now names it as an imperative call. `review-panel-fix-round.md` still described the retired `resolve-base-branch.sh` then `check-base-moved.sh` pair and now names the `sync-panel-base.sh` call. Both files join `**Files:**`. The guard also adds exit 3 (overlap, ask the operator) beside the design's 0/1/2, and prints a `REBASED:` line after every clean rebase so a later exit 2 or 3 never loses the new merge base. Reported by the group-10 implementer; the two prose edits were applied by the parent at pick, the recommended answers under the operator's standing instruction.

**Decision:** all-designed-rows-in-scope

- [x] 25. refresh-plan-base.sh reports what moved under the plan

Row MX6 — `design-basesync.md` item 6.

  - [x] **Step 1: Write the failing test** `TestRefreshPlanBase` in new
    `stats/internal/guard/refresh_plan_base_test.go` (UNMOVED one line; MOVED intersected with the
    plan's `**Files:**` by exact and leading-directory match; a spec printed at `origin/<base>`
    between delimiters, or reported absent; never rebases).
  - [x] **Step 2: Run it; it fails** — `cd stats && go test ./internal/guard -run 'TestRefreshPlanBase' -count=1`.
  - [x] **Step 3: Implement** `stats/internal/guard/refreshplanbase.go`, shim
    `scripts/refresh-plan-base.sh`, symlink `skills/flow/scripts/refresh-plan-base.sh`; replace the
    base-refresh recipe in `skills/flow/implement.md` with the call line, keeping the collision
    judgment and "never rebases" as prose.
  - [x] **Step 4: Verify** — targeted run; lint lines; verbatim-moves lines recorded.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/refreshplanbase.go`, `stats/internal/guard/refresh_plan_base_test.go`,
`scripts/refresh-plan-base.sh`, `skills/flow/scripts/refresh-plan-base.sh`, `skills/flow/implement.md`
**Tests:** `TestRefreshPlanBase`
**Regression:** reverting removes the guard: the test finds no `refresh-plan-base`.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/refresh_plan_base_test.go 2>/dev/null | awk '/^func Test/{n++} END{print n+0}' @ 67a082a6 -->
**Commit:** `feat(guard): refresh-plan-base.sh reports what moved under the plan`
**After:** Task 3, 5, 6, 22
**Build:** green

**Decision:** all-designed-rows-in-scope

- [x] 26. Live verification against the worktree's own store and scratch fixtures

  - [x] **Step 1: Before** — against the worktree's isolated stack (`.flow/project.md` `## run` /
    `## workspace isolation`, never the dev workspace's `flowd` or `flow` database), record `flow
    state get` for a scratch change `kan-860-live-check` created with `flow state set`, and
    `flow record handoff-lines -change kan-860-live-check -C <worktree>`'s absence (unknown verb on
    the old binary).
  - [x] **Step 2: Exercise** with the worktree-built `flow`: `flow state add-worktree` twice (a
    second path) and read back both entries with `state` unchanged; `flow stage mark` over two
    keys, then `flow stage` records show both begun and ended; `flow record handoff-lines`;
    `flow decision render` on this change's `.superpowers/sdd/decision.json`; `plan-class.sh` on
    this change's `tasks.md`. In a scratch repo with a local bare `origin`: `kickoff-worktree.sh`,
    then `remove-change-worktrees.sh` (first call discloses, `--proceed` removes).
  - [x] **Step 3: Record** before/after figures in the commit message body. **Not working looks
    like:** the second `add-worktree` dropping the first entry or changing `state`; a stage key
    begun and never ended; `decision render` differing from the block printed at the plan gate;
    `remove-change-worktrees.sh` removing anything on its first call.
  - [x] **Step 4: Clean up** the scratch change record and scratch repo.
  - [x] **Step 5: Commit** (empty commit carrying the record).

**Files:** none
**Tests:** none — live run; the figures are the record
**Regression:** none — verification only.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 67a082a6 -->
**Commit:** `test(stats): live-verify the KAN-860 verbs and guards`
**After:** Task 7, 12, 15, 19, 20, 23, 24, 25
**Build:** green

Correction (2026-09-30): `check-task-commit-fields.sh` read `**Files:** none` as a declared path the empty commit never touched, so this task's close failed; `Files:` now follows `Tests:`, where `none` declares nothing (commit 57542e60, `stats/internal/guard/taskcommitfields.go` + case 156). Found at pick, applied by the parent.
