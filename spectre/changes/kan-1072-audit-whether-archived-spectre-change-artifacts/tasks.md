# kan-1072-audit-whether-archived-spectre-change-artifacts

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** Audit spectre's archived change artifacts, replace flow's archive step with one landing commit that deletes the change directory, collapse every archive reader, delete the existing archives at the tip of every repository spectre has run in, and file one umbrella follow-up (KAN-1072).

**Architecture:** a `stats/internal/landed` history reader for landed changes; `commit-archive` becomes the drop commit; every live-else-archive fallback and unlanded-archive scan goes; prose follows; the parent makes six tip-cleanup commits after the panel.

**Tech Stack:** Go (`stats/`), Python 3 stdlib (`hooks/`), Bash shims, Markdown.

**Spec:** `design.md` beside this file.

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts` (`<wt>`), branch `spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`; `**Files:**` paths are relative to `<wt>`. Never touch a main checkout, never create a worktree, never write `spectre/changes/`.
- Go: `gofmt -w` every Go file changed; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean. Every Go test run sets a Bash `timeout` of 600000.
- Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database.
- Mutation proofs restore with `scripts/break-and-prove.sh` — never `git checkout --` over uncommitted edits.
- Guard, shim and stage names stay as they are (decision keep-guard-and-stage-names). The drop subject is `chore(spectre): drop <name>` (decision drop-commit-subject).
- A prose task rewords run-loaded sentences on purpose: its implementer lists, in its report, every `check-verbatim-moves.sh` FAIL line it intends, and the parent appends them to this change's `verbatim-moves.txt` before the before-review planning commit (decision verbatim-moves-appended-by-parent).
- Kept on purpose, never swept: `docs/` history (self-review reports, prompt audits, prompt-tuning cases), historical sentences in `-rationale.md` files that describe what an earlier change did, `skills/flow-contracts/lessons.md`'s narrative scan and `skills/spectre-planning/SKILL.md`'s rejection search (both the umbrella follow-up's — decision archive-readers-empty-until-follow-up), and `stats/internal/landed`'s legacy `archive` branch.

## Tip cleanup — the parent's, after the review panel

Not a task (decision tip-cleanup-outside-tasks). After the review panel's last round closes and before `flow.verify`:

1. Create each peer worktree with `<agents repo>/scripts/kickoff-worktree.sh --from <wt> /Users/tweety53/Projects/<repo> kan-1072-audit-whether-archived-spectre-change-artifacts`, for `<repo>` in `spectre`, `gymie`, `gymie-playwright`, `flow-eval`, `gymie-frontend`, in that order (`spectre/peers` order after Task 2), and land the merge-order planning commit per `skills/flow/cross-repo-worktrees.md`.
2. In each of the six worktrees — `<wt>` and `/Users/tweety53/Projects/<repo>-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts` for each `<repo>` above — one call each:
   `git -C <worktree> rm -r -q spectre/changes/archive`, then
   `git -C <worktree> commit -q -m "chore(spectre): remove the change archive" -m "Flow-Key: flow.sdd-tdd"`, then
   `git -C <worktree> push origin spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`.
3. Prove it: `git -C <worktree> ls-tree -d HEAD spectre/changes/archive` prints nothing in each of the six.

## Review Focus

- Task 4: a re-run after the drop with no new work is a no-op; with new work (directory restored by a fix run) it drops again; a satellite's `link.md` directory is dropped in the same pass; the scope check refuses anything but deletions under the change's own directories.
- Task 3: the history reader picks the newest drop commit's parent, falls back to the archive subject only when no drop commit exists, and never matches a name that is a prefix of another (`kan-1` vs `kan-10`).
- Task 6: no guard still scans `spectre/changes/archive/`, and a live in-flight change is still scanned.
- Task 7: a landed `pipeline-fix-<name>-<k>` keeps `<k>` taken.

- [ ] 1. Write the spectre archive audit
  - [ ] **Step 1: Measure.** In each repository the design's Context table names (main checkouts, read-only — `git -C`, `du`, `ls`, nothing that writes), re-run round 1's measurements (`.superpowers/sdd/planner/round-1.md` in `<wt>`): archive working-tree size and change count, PNG share, `spectre/` blobs in history, repository storage, `chore(spectre)` and `docs(self-review)` commit counts against the non-merge denominator, and commits touching only `spectre/` or `docs/self-review/`. Investigate gymie-playwright's loose storage and state its cause. Keep every command beside its number.
  - [ ] **Step 2: Write `docs/audits/2026-10-31-spectre-archive-audit.md`.** Sections: scope (the six repositories, and the two with no archive); the size and commit tables, each cell followed by its command and the date measured; the consumer list with path:line (round 1's, re-checked against `<wt>` before Task 3 changes it — cite the merge-base revision); the verdict table (round 2's as amended by `answers-2.md` and `answers-3.md`); the spectre surface table; what this change did (the landing drop, the collapsed readers, the tip cleanup — decisions landing-drop-replaces-archive, tip-cleanup-in-this-change) and what it left (decision own-landing-archives); history cleanup numbers and procedure, recommended against (decision no-history-rewrite); the umbrella follow-up's full text — its four items, each with what it changes and why, plus the deletion of the residual archives decision own-landing-archives names (decision one-umbrella-follow-up). No Jira key in the document (decision doc-carries-no-jira-key).
  - [ ] **Step 3: Verify.** `scripts/check-markdown-integrity.sh` and `scripts/check-references.sh` exit 0; every number in the document has its command beside it. Commit.

**Build:** green
**Files:** `docs/audits/2026-10-31-spectre-archive-audit.md`
**Tests:** **none** — a document.
**Regression:** none — no code.
**Commit:** docs(audits): audit spectre's archived change artifacts
**After:** none

**Decision:** audit-doc-location
**Decision:** one-umbrella-follow-up
**Decision:** doc-carries-no-jira-key

- [ ] 2. Declare flow-eval and gymie-frontend as spectre peers
  - [ ] **Step 1: Edit.** Append `flow-eval ../flow-eval` and `gymie-frontend ../gymie-frontend` to `spectre/peers`, one line each, after the existing three.
  - [ ] **Step 2: Verify.** `spectre validate kan-1072-audit-whether-archived-spectre-change-artifacts` (from `<wt>`) exits 0; `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `spectre/peers`
**Tests:** **none** — a configuration line.
**Regression:** none — `kickoff-worktree.sh --from` would exit 2 for those two repositories at the tip cleanup.
**Commit:** chore(spectre): declare flow-eval and gymie-frontend as peers
**After:** none

**Decision:** peers-declare-flow-eval-gymie-frontend

- [ ] 3. Read landed changes from history; resolve the live change directory only
  - [ ] **Step 1: Write the tests.** `stats/internal/landed/landed_test.go`, on a temp git repository built in the test: `TestFileReadsTheDropCommitsParent` — a change committed at `spectre/changes/kan-1/tasks.md`, then a `chore(spectre): drop kan-1` commit removing it: `File(repo, "kan-1", "tasks.md")` returns the pre-drop bytes; a second restore-and-drop cycle with edited content returns the newer bytes. `TestFileFallsBackToTheArchiveCommit` — only a `chore(spectre): archive kan-1` commit holding `spectre/changes/archive/kan-1/tasks.md`: returns those bytes. `TestFileMatchesTheWholeName` — a `drop kan-10` commit does not answer `kan-1`. `TestFileAbsent` — neither subject: `ok` false, no error. In `stats/internal/guard/helpers_test.go`: `TestChangeDirIsLiveOnly` — with only `changes/archive/kan-1/tasks.md` on disk, `ChangeDir(changes, "kan-1")` returns `changes/kan-1`. In `stats/internal/api/plantasks_test.go`: replace `TestPlanTasksArchived` with `TestPlanTasksLanded` — a FINISHED row whose root is a temp repository holding only the drop commit history returns the plan's tasks with `finished=true`. In `stats/cmd/flow/tasks_test.go`, delete `TestTasksTickWritesArchivedPlan`. Run `cd stats && go test ./internal/landed ./internal/guard ./internal/api ./cmd/flow -run 'TestFile|TestChangeDir|TestPlanTasks|TestTasksTick' -count=1` — fails.
  - [ ] **Step 2: Implement.** `stats/internal/landed/landed.go`: `File(repo, name, file string) (b []byte, ok bool, err error)` — `git -C repo log --all -1 --format=%H --fixed-strings --grep 'chore(spectre): drop <name>'` filtered to a subject equal to the whole string, then `git show <sha>^:spectre/changes/<name>/<file>`; else the same for `chore(spectre): archive <name>` at `<sha>:spectre/changes/archive/<name>/<file>`; a git failure other than "not found" is `err`. `ChangeDir` in `stats/internal/guard/changeplan.go` returns `changes/<name>` always; rewrite its doc comment. `stats/cmd/flow/tasks.go` drops its archive fallback comment/branch. `readPlan` in `stats/internal/api/stats.go` reads the live `tasks.md` from each root, then, for a FINISHED row, `landed.File(root, name, "tasks.md")`; rewrite `planTasks`' doc comment.
  - [ ] **Step 3: Prove.** `scripts/break-and-prove.sh` with the whole-subject filter removed — `TestFileMatchesTheWholeName` fails; restore.
  - [ ] **Step 4: Verify.** Step 1's command passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/landed/landed.go`, `stats/internal/landed/landed_test.go`, `stats/internal/guard/changeplan.go`, `stats/internal/guard/helpers_test.go`, `stats/cmd/flow/tasks.go`, `stats/cmd/flow/tasks_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/plantasks_test.go`
**Allowed-collateral:** `stats/internal/guard/*_test.go`, `stats/cmd/flow/*_test.go`
**Tests:** `TestFileReadsTheDropCommitsParent`, `TestFileFallsBackToTheArchiveCommit`, `TestFileMatchesTheWholeName`, `TestFileAbsent`, `TestChangeDirIsLiveOnly`, `TestPlanTasksLanded`
**Regression:** `TestPlanTasksLanded` fails — the board shows no plan for any landed change once its directory leaves the tree.
**Commit:** feat(stats): read landed changes from history and resolve only the live change directory
**After:** none

**Decision:** landed-change-history-reader

- [ ] 4. Land a drop commit instead of archiving the change
  - [ ] **Step 1: Write the tests.** `stats/internal/guard/commit_archive_test.go`: `TestCommitArchiveDropsTheChange` — on a temp branch `spectre/demo` holding `spectre/changes/demo/` and `spectre/changes/demo-fix-1/`, `commit-archive <wt> demo` leaves neither directory nor any `spectre/changes/archive/` at HEAD, the subject is `chore(spectre): drop demo`, no `ledger.md`/`panel.md` was written, stdout `ARCHIVE-COMMITTED`. `stats/internal/guard/check_archive_scope_test.go`: `TestCheckArchiveScopeAdmitsOnlyTheDrop` — staged deletions under `spectre/changes/demo/` pass; a staged addition anywhere, or a deletion outside the change's directories, fails. `stats/internal/guard/integrate_change_prepare_test.go`: `TestIntegrateChangePrepareRerunAfterDrop` — a branch whose log holds `chore(spectre): drop demo` and whose live directory is absent reports the already-dropped re-run (the token the archived re-run printed); with the directory restored, prepare proceeds as a first pass. `stats/internal/guard/integrate_change_test.go`: `TestIntegrateChangeDropsSatelliteLink` — a satellite worktree holding `spectre/changes/demo/link.md` ends with that directory deleted by its own drop commit. Update `TestCommitArchive`, `TestCheckArchiveScope`, `TestIntegrateChangeHappyPath`, `TestIntegrateChangePrepareCommit`, `TestIntegrateChangeLandCleanup`, `TestReshapeBranch` to the drop. Run `cd stats && go test ./internal/guard -run 'TestCommitArchive|TestCheckArchiveScope|TestIntegrateChange|TestReshapeBranch' -count=1` — fails.
  - [ ] **Step 2: Implement.** `commitarchive.go`: after the branch and Done-when checks, `git rm -r -q` the live `spectre/changes/<name>` and every `spectre/changes/<name>-fix-<digits>` directory, no record copy, `checkArchiveScope`, commit `chore(spectre): drop <name>`; tokens unchanged. `archivescope.go`: admit only staged deletions under those directories. `integratechange_commit.go`: remove `spectreArchive` and its calls; `flow.sync-archive` marks around nothing but the existence check; `commit-archive` runs in the canonical worktree and in each satellite worktree holding `spectre/changes/<name>`. `integratechange_prepare.go`: archive-undo becomes drop-undo (a staged, uncommitted drop is restored with `git restore --staged --worktree`); "archived" becomes "dropped" — the drop subject in `git log` and the live directory absent; the with/without-new-work branches keep their tokens. `integratechange.go`: `findCanonical` probes the live `tasks.md` only; `mergeOrder` reads the live `link.md`, else `landed.File(wt, name, "link.md")`; `archBase`/`ArchivedRerunBase`/`selfReviewBase` match `chore(spectre): drop <name>` and `chore(spectre): drop self-review-<name>`. Reword the comments in `donewhenpaths.go` and `untaskedcommits.go`.
  - [ ] **Step 3: Prove.** `scripts/break-and-prove.sh` with the `-fix-<digits>` removal deleted — `TestCommitArchiveDropsTheChange` fails; restore.
  - [ ] **Step 4: Verify.** Step 1's command passes, plus `cd stats && go test ./cmd/flow -run 'TestStage' -count=1`; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/commitarchive.go`, `stats/internal/guard/archivescope.go`, `stats/internal/guard/integratechange.go`, `stats/internal/guard/integratechange_commit.go`, `stats/internal/guard/integratechange_prepare.go`, `stats/internal/guard/donewhenpaths.go`, `stats/internal/guard/untaskedcommits.go`, `stats/internal/guard/commit_archive_test.go`, `stats/internal/guard/check_archive_scope_test.go`, `stats/internal/guard/integrate_change_prepare_test.go`, `stats/internal/guard/integrate_change_test.go`
**Allowed-collateral:** `stats/internal/guard/*_test.go`
**Tests:** `TestCommitArchiveDropsTheChange`, `TestCheckArchiveScopeAdmitsOnlyTheDrop`, `TestIntegrateChangePrepareRerunAfterDrop`, `TestIntegrateChangeDropsSatelliteLink`
**Regression:** `TestCommitArchiveDropsTheChange` fails — run 1 moves the change into `archive/` again and the tip keeps every landed change.
**Commit:** feat(guard): drop the change directory at land instead of archiving it
**After:** Task 3

**Decision:** landing-drop-replaces-archive
**Decision:** drop-commit-subject
**Decision:** fix-run-restores-change-dir
**Decision:** keep-guard-and-stage-names

- [ ] 5. Build the self-review bundle from the dropped change's history
  - [ ] **Step 1: Write the tests.** In `stats/internal/selfreview/bundle_test.go`: `TestBundleAssemblyReadsDroppedFilesThroughGit` — after a drop commit on `spectre/<n>`, the bundle serves `tasks.md`, `design.md`, `narrative.md`, `visual-verification.md`, `live-verification.md` and `link.md` from the drop commit's parent; `TestDeriveFinishCommitsFollowsTheDrop` — the finish commits are derived from the drop subject and a re-run's second drop commit is followed; `TestBundleAssemblyRendersRecordsFromTheStoreOnly` — with no store render, the ledger and panel are absent (no committed copy is read). Rewrite the archive-based tests (`TestBundleAssemblyReadsArchivedFilesThroughGit`, `TestBundleServesArchivedVisualVerification`, `TestBundleServesArchivedLiveVerification`, `TestBundleAssemblyDerivesFinishCommits`, `TestBundleAssemblyListsEveryCommitOfArchivedRerun`, `TestBundleAssemblyReadsArchiveFromBaseAfterMerge`, `TestBundleAssemblyLogsEveryMergeOrderPeer`, `TestDeriveFinishCommitsFollowsArchivedRerun`, `TestDeriveFinishCommitsSiblingArchiveSubjectLoses`) onto the drop, and delete `TestBundleAssemblyFallsBackToCommittedRecords` and `TestBundleAssemblySkipsCommittedRecordsWhenAbsent`. Run `cd stats && go test ./internal/selfreview ./internal/api -run 'TestBundle|TestDeriveFinishCommits|TestSelfReview' -count=1` — fails.
  - [ ] **Step 2: Implement.** `stats/internal/selfreview/git.go`: remove `archiveDir`; `archivedFiles`, `archiveRepo`/`archiveRev`, `peerLogs` and `Bundle` read through `landed.File`; `reservedShapes` and `deriveFinishCommits` match `chore(spectre): drop <name>` and log `-- spectre/changes/<name>`; `committedRecord` and its fallback go. Reword the comments in `stats/internal/api/selfreview.go` and `stats/internal/client/selfreview.go`.
  - [ ] **Step 3: Prove.** `scripts/break-and-prove.sh` with `deriveFinishCommits` back on the archive subject — `TestDeriveFinishCommitsFollowsTheDrop` fails; restore.
  - [ ] **Step 4: Verify.** Step 1's command passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/selfreview/git.go`, `stats/internal/selfreview/bundle_test.go`, `stats/internal/api/selfreview.go`, `stats/internal/client/selfreview.go`
**Allowed-collateral:** `stats/internal/api/selfreview_test.go`
**Tests:** `TestBundleAssemblyReadsDroppedFilesThroughGit`, `TestDeriveFinishCommitsFollowsTheDrop`, `TestBundleAssemblyRendersRecordsFromTheStoreOnly`
**Regression:** `TestBundleAssemblyReadsDroppedFilesThroughGit` fails — the self-review pass reads an empty bundle for every change once run 1 drops its directory.
**Commit:** feat(selfreview): read the bundle from the dropped change's history
**After:** Task 3, 4

**Decision:** landed-change-history-reader
**Decision:** landing-drop-replaces-archive

- [ ] 6. Cut the unlanded-archive scans and the guards' archive fallbacks
  - [ ] **Step 1: Write the tests.** `stats/internal/guard/check_plan_shape_test.go`: `TestPlanParseScansLiveChangesOnly` — a branch holding `spectre/changes/archive/old/tasks.md` (absent on the base) with a shape violation and a live change without one: `check-plan-shape` passes and names no `archive/` path. `stats/internal/guard/check_verbatim_moves_test.go`: `TestCheckVerbatimMovesReadsLiveAcksOnly` — an ack only in `spectre/changes/archive/old/verbatim-moves.txt` acknowledges nothing; one in the live change's file does. `stats/internal/guard/check_task_records_test.go`: `TestCheckTaskRecordsIgnoresArchiveDir` — an `archive/<n>/tasks.md` on the branch is not checked. Update `TestCheckTaskBuildGreen`, `TestPlanDispatchBundles`, `TestCheckPlanProvenance`, `TestCheckPlanUnchanged`, `TestCheckUnfinishedWork`, `TestCheckVisualVerifyDispatchedIgnoresAChangeRecordMove`, `TestCheckVisualVerifyDispatchedRequiresTheVerifierRecord`, `TestCheckVisualVerifyDispatchedCommaJoinedDiffBase` to the live path. Delete `stats/internal/guard/unlanded_archives_test.go`. Run `cd stats && go test ./internal/guard -run 'TestPlanParse|TestCheckPlanShape|TestCheckVerbatimMoves|TestCheckTaskRecords|TestCheckTaskBuildGreen|TestPlanDispatchBundles|TestCheckPlanProvenance|TestCheckPlanUnchanged|TestCheckUnfinishedWork|TestCheckVisualVerifyDispatched' -count=1` — fails.
  - [ ] **Step 2: Implement.** Delete `stats/internal/guard/unlandedarchives.go` (and its `Registry` entry). Remove its calls and the archived-name lists in `planparse.go`, `taskrecords.go` (and the `p.archived` relaxation), `planprovenance.go` (the appended `archive/<n>` dirs and the prefix strip) and `verbatimmoves.go` (`vmAcks`' archive read). `planunchanged.go` lists `spectre/changes/<n>` only. `unfinishedwork.go` drops its `/changes/archive/` skip. `visualverifydispatched.go` reads the live `visual-verification.md` only. `specroot.go`'s `holdsOpenChange`, `planprovenance.go`'s directory walk and `ownedcorpus.go`'s owned prefix drop their `archive` entries. Reword every comment these edits leave stale.
  - [ ] **Step 3: Prove.** `scripts/break-and-prove.sh` with `vmAcks` reading the archive again — `TestCheckVerbatimMovesReadsLiveAcksOnly` fails; restore.
  - [ ] **Step 4: Verify.** Step 1's command passes; `git grep -n -F -e 'unlandedArchives' -e 'changes/archive' -- stats/internal/guard/planparse.go stats/internal/guard/taskrecords.go stats/internal/guard/planprovenance.go stats/internal/guard/verbatimmoves.go stats/internal/guard/planunchanged.go stats/internal/guard/unfinishedwork.go stats/internal/guard/visualverifydispatched.go stats/internal/guard/specroot.go stats/internal/guard/ownedcorpus.go` prints nothing; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/unlandedarchives.go`, `stats/internal/guard/unlanded_archives_test.go`, `stats/internal/guard/planparse.go`, `stats/internal/guard/taskrecords.go`, `stats/internal/guard/planprovenance.go`, `stats/internal/guard/verbatimmoves.go`, `stats/internal/guard/planunchanged.go`, `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/specroot.go`, `stats/internal/guard/ownedcorpus.go`, `stats/internal/guard/check_plan_shape_test.go`, `stats/internal/guard/check_verbatim_moves_test.go`, `stats/internal/guard/check_task_records_test.go`
**Allowed-collateral:** `stats/internal/guard/*_test.go`
**Tests:** `TestPlanParseScansLiveChangesOnly`, `TestCheckVerbatimMovesReadsLiveAcksOnly`, `TestCheckTaskRecordsIgnoresArchiveDir`
**Regression:** `TestPlanParseScansLiveChangesOnly` fails — the plan guards read archive directories no change writes any more.
**Commit:** refactor(guard): cut the unlanded-archive scans and archive fallbacks
**After:** Task 3

**Decision:** landing-drop-replaces-archive

- [ ] 7. Keep pipeline-fix keys unique from history and land on a dropped change
  - [ ] **Step 1: Write the tests.** `stats/internal/guard/pipeline_fix_test.go`: `TestPipelineFixHighestKCountsLandedChanges` — a base whose tree holds `spectre/changes/pipeline-fix-x-1/`, and whose log holds `chore(spectre): drop pipeline-fix-x-3` and `chore(spectre): archive pipeline-fix-x-2`: the next key is 4. `TestPipelineFixLandRefusesUndroppedChange` — `land` with `spectre/changes/<branch>` still at HEAD refuses, and its message names `chore(spectre): drop <branch>`. Run `cd stats && go test ./internal/guard -run 'TestPipelineFix' -count=1` — fails.
  - [ ] **Step 2: Implement.** `pfHighestK` in `stats/internal/guard/pipelinefix.go` per decision pipeline-fix-key-from-history (live directories on the base plus both subjects in `git log --format=%s <base>`); drop the `spectre/changes/archive/` listing. `pfLand`'s refusal message names `git rm -r spectre/changes/<b>` and the drop subject. Reword the header of `scripts/pipeline-fix.sh`.
  - [ ] **Step 3: Prove.** `scripts/break-and-prove.sh` with the log scan removed — `TestPipelineFixHighestKCountsLandedChanges` fails; restore.
  - [ ] **Step 4: Verify.** Step 1's command passes; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/pipeline_fix_test.go`, `scripts/pipeline-fix.sh`
**Tests:** `TestPipelineFixHighestKCountsLandedChanges`, `TestPipelineFixLandRefusesUndroppedChange`
**Regression:** `TestPipelineFixHighestKCountsLandedChanges` fails — a new pipeline-fix reuses a landed change's key.
**Commit:** fix(guard): count landed pipeline-fix keys from history
**After:** none

**Decision:** pipeline-fix-key-from-history
**Decision:** landing-drop-replaces-archive

- [ ] 8. Report a change directory left at the tip; refuse an undropped landing in the hook
  - [ ] **Step 1: Write the tests.** `stats/internal/guard/check_cleanup_complete_test.go`: `TestCheckCleanupCompleteNamesAnUndroppedChange` — `spectre/changes/demo` and `spectre/changes/demo-fix-1` at the tip are LEFTOVER rows whose text says the directory was not dropped and names `chore(spectre): drop demo`, and never says "archive". `stats/internal/repotest/protect_main_checkout_test.go`: `TestProtectMainCheckoutUndroppedLandingMessage` — the hook's refusal names `git rm -r spectre/changes/<change>` and `chore(spectre): drop <change>`, never `spectre archive`; a `chore(spectre): drop <change>` subject counts as a chore subject. Run `cd stats && go test ./internal/guard ./internal/repotest -run 'TestCheckCleanupComplete|TestProtectMainCheckout' -count=1` — fails.
  - [ ] **Step 2: Implement.** `cleanupcomplete.go` row four and its COMPLETE text reworded. `hooks/protect-main-checkout.py`: `unarchived_landing` keeps its predicate (the change directory at the tip), its docstring, the module docstring and the refusal message at the landing site name the drop; the chore-subject list admits `drop`. Header of `scripts/check-cleanup-complete.sh` reworded.
  - [ ] **Step 3: Verify.** Step 1's command passes; `/usr/bin/python3 -m py_compile hooks/protect-main-checkout.py` exits 0; `scripts/check-python-suppressions.sh` exits 0; lint. Commit.

**Build:** green
**Files:** `stats/internal/guard/cleanupcomplete.go`, `stats/internal/guard/check_cleanup_complete_test.go`, `hooks/protect-main-checkout.py`, `stats/internal/repotest/protect_main_checkout_test.go`, `scripts/check-cleanup-complete.sh`
**Tests:** `TestCheckCleanupCompleteNamesAnUndroppedChange`, `TestProtectMainCheckoutUndroppedLandingMessage`
**Regression:** `TestProtectMainCheckoutUndroppedLandingMessage` fails — the hook tells the operator to run `spectre archive --force`, a step the flow no longer has.
**Commit:** fix(guard): report an undropped change directory at cleanup and landing
**After:** none

**Decision:** landing-drop-replaces-archive

- [ ] 9. Resolve records and state without the archive directory
  - [ ] **Step 1: Write the tests.** `stats/cmd/flow/record_test.go`: replace `TestRequireChangeCheckoutAcceptsAnArchiveOnAnotherBranch` with `TestRequireChangeCheckoutAcceptsALandedChange` — a change whose only trace is a drop commit on another branch (`rev-list --all -- spectre/changes/<change>` non-empty) is accepted. `stats/cmd/flow/state_test.go`: replace `TestStateResolveFallbackDropsArchived` and `TestStateResolveFallbackKeepsArchivedInProgress` with `TestStateResolveFallbackDropsLanded` — a FINISHED record with no live directory is dropped from the fallback list — and `TestStateResolveFallbackKeepsInProgressWithoutDir` — an IN_PROGRESS record with no live directory is kept. Run `cd stats && go test ./cmd/flow -run 'TestRequireChangeCheckout|TestRecordWrite|TestStateResolve|TestStateList|TestStageEnd' -count=1` — fails.
  - [ ] **Step 2: Implement.** `stats/cmd/flow/record.go`: `requireChangeCheckout` runs `rev-list --all -- spectre/changes/<change>`; `changeCheckouts` drops `spectre/changes/archive` from its roots; the `archive` name refusal stays (still not a change name). `stats/cmd/flow/state.go`: `resolveCandidates` drops the `archive/<name>` stat; a record is dropped when it is FINISHED and no live directory holds it. `stats/cmd/flow/stage.go`: comments at the `ArchivedRerunBase` callers reworded.
  - [ ] **Step 3: Verify.** Step 1's command passes; lint. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/cmd/flow/state.go`, `stats/cmd/flow/state_test.go`, `stats/cmd/flow/stage.go`
**Allowed-collateral:** `stats/cmd/flow/*_test.go`
**Tests:** `TestRequireChangeCheckoutAcceptsALandedChange`, `TestStateResolveFallbackDropsLanded`, `TestStateResolveFallbackKeepsInProgressWithoutDir`
**Regression:** `TestRequireChangeCheckoutAcceptsALandedChange` fails — `flow record` refuses every landed change once no archive directory exists.
**Commit:** fix(stats): resolve records and state of a landed change without the archive
**After:** Task 4

**Decision:** landing-drop-replaces-archive

- [ ] 10. Reword the shim headers that describe the archive
  - [ ] **Step 1: Edit.** Rewrite each header sentence describing `spectre archive`, the archive commit, the archived path or an unlanded-archive scan to the drop, in: `scripts/commit-archive.sh`, `scripts/check-archive-scope.sh`, `scripts/integrate-change.sh`, `scripts/check-plan-shape.sh`, `scripts/plan-dispatch-bundles.sh`, `scripts/check-task-build-green.sh`, `scripts/check-task-commit-fields.sh`, `scripts/check-plan-unchanged.sh`, `scripts/check-task-reviewer-single-dispatch.sh`, `scripts/check-visual-verify-dispatched.sh`, `scripts/check-unfinished-work.sh`, `scripts/check-untasked-commits.sh`, `scripts/check-done-when-paths.sh`, `scripts/check-verbatim-moves.sh`. Comment lines only — no command line changes.
  - [ ] **Step 2: Verify.** `git diff -U0 -- scripts/ | grep -E '^[-+][^-+]' | grep -vE '^[-+]#'` prints nothing (comment-only hunks); `bash -n` each file exits 0; `git grep -n -F -e 'spectre archive' -e 'changes/archive' -- scripts/commit-archive.sh scripts/check-archive-scope.sh scripts/integrate-change.sh scripts/check-plan-shape.sh scripts/plan-dispatch-bundles.sh scripts/check-task-build-green.sh scripts/check-task-commit-fields.sh scripts/check-plan-unchanged.sh scripts/check-task-reviewer-single-dispatch.sh scripts/check-visual-verify-dispatched.sh scripts/check-unfinished-work.sh scripts/check-untasked-commits.sh scripts/check-done-when-paths.sh scripts/check-verbatim-moves.sh` prints nothing; `scripts/check-references.sh` exits 0. Commit.

**Build:** green
**Files:** `scripts/commit-archive.sh`, `scripts/check-archive-scope.sh`, `scripts/integrate-change.sh`, `scripts/check-plan-shape.sh`, `scripts/plan-dispatch-bundles.sh`, `scripts/check-task-build-green.sh`, `scripts/check-task-commit-fields.sh`, `scripts/check-plan-unchanged.sh`, `scripts/check-task-reviewer-single-dispatch.sh`, `scripts/check-visual-verify-dispatched.sh`, `scripts/check-unfinished-work.sh`, `scripts/check-untasked-commits.sh`, `scripts/check-done-when-paths.sh`, `scripts/check-verbatim-moves.sh`
**Tests:** **none** — comments.
**Regression:** none — comments only.
**Commit:** docs(scripts): describe the landing drop in the shim headers
**After:** Task 4, 6

**Decision:** landing-drop-replaces-archive

- [ ] 11. Rewrite the finish contracts for the landing drop
  - [ ] **Step 1: Edit.** `skills/flow-contracts/finish-contract-run1.md`: **Archive on the change branch** becomes the drop (decision landing-drop-replaces-archive) — `commit-archive` stages `git rm -r` of the change and its `-fix-<N>` sub-changes, no record copy, the Done-when check before it, satellites drop their `link.md` directory in their own worktree; the archive-undo, "already archived" and archived re-run paragraphs become drop-undo, "already dropped" and the dropped re-run (decision fix-run-restores-change-dir); the bundle paragraph reads the dropped change from history (decision landed-change-history-reader). `finish-contract-run2.md`: run 2 drops nothing; ledger and panel ride nothing. `finish-hand-fallbacks.md`: the `commit-archive.sh` and `check-archive-scope.sh` hand fallbacks become the drop. Move each replaced sentence's reasoning, where it is reasoning, verbatim into `finish-contract-rationale.md`.
  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh <wt> lint` exits 0 save the `check-verbatim-moves.sh` FAIL lines this task intends, which the report lists; `git grep -n -F -e 'spectre archive' -e 'changes/archive/' -- skills/flow-contracts/finish-contract-run1.md skills/flow-contracts/finish-contract-run2.md skills/flow-contracts/finish-hand-fallbacks.md` prints nothing. Commit.

**Build:** green
**Files:** `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-contracts/finish-contract-run2.md`, `skills/flow-contracts/finish-hand-fallbacks.md`, `skills/flow-contracts/finish-contract-rationale.md`
**Tests:** **none** — prose.
**Regression:** none — prose; integrate would follow archive steps the guards no longer take.
**Commit:** docs(finish-contract): land a drop commit instead of archiving
**After:** Task 4

**Decision:** landing-drop-replaces-archive
**Decision:** fix-run-restores-change-dir

- [ ] 12. Rewrite the pipeline, registry and boundary contracts for the landing drop
  - [ ] **Step 1: Edit.** `skills/flow-contracts/pipeline.md`: `FINISHED` means dropped, not archived; the loop's Land step drops `pipeline-fix-<name>-<k>` with `chore(spectre): drop …`; `<k>` is unique across live directories and landed subjects (decision pipeline-fix-key-from-history); the in-flight exclusion and **A change's directory** describe the live directory deleted at land. `artifacts-registry.md`: the change-directory row is "deleted by the landing drop"; ledger and panel are never copied into it. `git-boundaries.md`: the change directory has no archived path; the reshape follows the drop commit. `plan-provenance-guard.md`: no archived-on-branch exclusion. `state-file.md`, `known-bugs.md`, `operator-prompts-auto-resolution.md`, `skills/flow-contracts/SKILL.md`: each archive sentence follows. Move replaced reasoning verbatim into the matching `-rationale.md`.
  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh <wt> lint` exits 0 save the intended `check-verbatim-moves.sh` FAIL lines, listed in the report; `git grep -n -F -e 'spectre archive' -e 'changes/archive/' -- skills/flow-contracts/pipeline.md skills/flow-contracts/artifacts-registry.md skills/flow-contracts/git-boundaries.md skills/flow-contracts/plan-provenance-guard.md skills/flow-contracts/state-file.md skills/flow-contracts/known-bugs.md` prints nothing. Commit.

**Build:** green
**Files:** `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/artifacts-registry.md`, `skills/flow-contracts/git-boundaries.md`, `skills/flow-contracts/plan-provenance-guard.md`, `skills/flow-contracts/state-file.md`, `skills/flow-contracts/known-bugs.md`, `skills/flow-contracts/operator-prompts-auto-resolution.md`, `skills/flow-contracts/SKILL.md`
**Allowed-collateral:** `skills/flow-contracts/*-rationale.md`
**Tests:** **none** — prose.
**Regression:** none — prose.
**Commit:** docs(flow-contracts): describe the change directory deleted at land
**After:** Task 4, 7

**Decision:** landing-drop-replaces-archive
**Decision:** pipeline-fix-key-from-history

- [ ] 13. Rewrite the /flow skill prose for the landing drop
  - [ ] **Step 1: Edit.** `skills/flow/integrate.md`: prepare's archive-undo and the archived re-run gate become the drop's; section 4 drops and runs the self-review pass; the `archive`/`commit-archive` steps become the drop. `skills/flow/document-fix.md`: a fix run after the drop restores the change directory from the drop commit's parent as its fix-run planning commit (decision fix-run-restores-change-dir). `skills/flow/cleanup.md`: the report line names the dropped directory, not an archive path. `skills/flow/implement.md`: no archived-change branch for `spectre validate`, `list --json` or the task counts. `skills/flow/brainstorm-planner.md`: the Done-when check runs at the drop commit. `skills/flow/SKILL.md`, `stage-keys.md` (keys unchanged, their description the drop — decision keep-guard-and-stage-names), `verify-fix-loop.md`, `review-panel-late-fix.md`, `review-panel-experimental-slot.md`: each archive sentence follows. Move replaced reasoning verbatim into `skills/flow/SKILL-rationale.md` where it is reasoning.
  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh <wt> lint` exits 0 save the intended `check-verbatim-moves.sh` FAIL lines, listed in the report; `git grep -n -F -e 'spectre archive' -e 'changes/archive/' -- skills/flow/integrate.md skills/flow/document-fix.md skills/flow/cleanup.md skills/flow/implement.md skills/flow/brainstorm-planner.md skills/flow/SKILL.md skills/flow/stage-keys.md skills/flow/verify-fix-loop.md skills/flow/review-panel-late-fix.md skills/flow/review-panel-experimental-slot.md` prints nothing. Commit.

**Build:** green
**Files:** `skills/flow/integrate.md`, `skills/flow/document-fix.md`, `skills/flow/cleanup.md`, `skills/flow/implement.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/SKILL.md`, `skills/flow/stage-keys.md`, `skills/flow/verify-fix-loop.md`, `skills/flow/review-panel-late-fix.md`, `skills/flow/review-panel-experimental-slot.md`
**Allowed-collateral:** `skills/flow/*-rationale.md`
**Tests:** **none** — prose.
**Regression:** none — prose.
**Commit:** docs(flow): drop the change directory at land in the /flow prose
**After:** Task 4

**Decision:** landing-drop-replaces-archive
**Decision:** fix-run-restores-change-dir
**Decision:** keep-guard-and-stage-names

- [ ] 14. Rewrite the self-review, status and plan skills for the landing drop
  - [ ] **Step 1: Edit.** `skills/flow-self-review/SKILL.md`: the sub-change `self-review-<name>` is dropped with `chore(spectre): drop self-review-<name>`, not archived; the push hook refuses an undropped change; autosquash keeps the drop commit last. `skills/flow-status/SKILL.md`: task progress comes from `spectre list --json` alone — an open change always has its live directory — and the archived-`tasks.md` branch goes. `skills/flow-plan/SKILL.md`: the "never run `spectre archive`" line goes. Move replaced reasoning verbatim into `skills/flow-self-review/SKILL-rationale.md` and `skills/flow-status/SKILL-rationale.md` where it is reasoning.
  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh <wt> lint` exits 0 save the intended `check-verbatim-moves.sh` FAIL lines, listed in the report; `git grep -n -F -e 'spectre archive' -e 'changes/archive/' -- skills/flow-self-review/SKILL.md skills/flow-status/SKILL.md skills/flow-plan/SKILL.md` prints nothing. Commit.

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`, `skills/flow-status/SKILL.md`, `skills/flow-plan/SKILL.md`
**Allowed-collateral:** `skills/flow-self-review/SKILL-rationale.md`, `skills/flow-status/SKILL-rationale.md`
**Tests:** **none** — prose.
**Regression:** none — prose.
**Commit:** docs(flow-self-review): drop the self-review sub-change instead of archiving it
**After:** Task 5

**Decision:** landing-drop-replaces-archive

- [ ] 15. Rewrite the repository docs and the manual-review rule for the landing drop
  - [ ] **Step 1: Edit.** `README.md`: the run 1 diagram and sentences name the drop; `check-archive-scope.sh`'s description is the drop's scope. `rules/flow-manual-review.mdc`: the archiving clause names the drop. `.flow/project.md`: `verbatim-moves.txt` is committed with the plan and deleted by the drop; the archived-patches line names history instead. `AGENTS.md`: the "non-archived" sole active change wording follows (`CLAUDE.md` carries no counterpart). Move replaced reasoning verbatim into `.flow/project-rationale.md` where it is reasoning.
  - [ ] **Step 2: Verify.** `scripts/run-verify-list.sh <wt> lint` exits 0 save the intended `check-verbatim-moves.sh` FAIL lines, listed in the report; `git grep -n -F -e 'spectre archive' -e 'non-archived' -- README.md rules/flow-manual-review.mdc AGENTS.md` prints nothing. Commit.

**Build:** green
**Files:** `README.md`, `rules/flow-manual-review.mdc`, `.flow/project.md`, `AGENTS.md`
**Allowed-collateral:** `.flow/project-rationale.md`
**Tests:** **none** — prose.
**Regression:** none — prose.
**Commit:** docs: describe the landing drop in the repository docs
**After:** Task 4

**Decision:** landing-drop-replaces-archive
