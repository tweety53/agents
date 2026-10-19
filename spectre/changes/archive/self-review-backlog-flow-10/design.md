## Context

- Five Backlog Flow findings — the next five live ones not already fixed, oldest first: KAN-899,
  KAN-900, KAN-902, KAN-905, KAN-907 — planned as one change per item 1 of **3. Fix every finding
  that is not `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-19-self-review-backlog-flow-10-design.md` in this worktree. Every
  question was auto-resolved to its recommended option
  (`skills/flow-contracts/operator-prompts-auto-resolution.md`). Each ticket's finding was read
  from the self-review report that filed it and the panel table of that report's context bundle in
  git history; the Jira issues themselves were not readable from this session.
- **Not in this change.** KAN-764 (a design ask only the operator can draw) and KAN-944 (another
  repository's tooling) were excluded by the dispatch. Two candidates were set aside while
  verifying:
  - **KAN-901** — the untested recompute-failure exit and the no-mutation pinning of cases 23–25 in
    `prepare-archive-branch`: not live, for the parent to move to Done. `9721e67e` ("archive and save
    the self-review bundle at integrate on the change branch") deleted
    `stats/internal/guard/preparearchivebranch.go`, `prepare_archive_branch_test.go` and
    `scripts/prepare-archive-branch.sh`; no code path or test the finding names exists.
    <!-- measured: git show --stat 9721e67e @ 46db35e0 → the three files deleted (354, 886 and 170 lines); ls stats/internal/guard | grep -i archive @ 46db35e0 → archivescope.go, commitarchive.go, unlandedarchives.go and their tests only -->
  - **KAN-906** — shared helpers living in unrelated guards' files under their prefixes: `big`,
    left filed, the blast-radius line being the one **Pipeline defects found mid-run**
    (`skills/flow-contracts/pipeline.md`) draws. Moving every guard-prefixed helper used outside its
    own file — `smcAbs`, `pcAbs`, the `tcf*`, `pc*`, `lp*`, `vt*`/`vv*` families and the rest — to a
    shared file under neutral names touches 85 files of `stats/internal/guard/`, counting definition
    files, callers and the tests that name them.
    <!-- measured: for every `^func [a-z][a-z0-9]{1,4}[A-Z]` helper in a non-test file of stats/internal/guard, less the neutral is*/git*/env*/check*/base*/run*/spec*/... names, the union of its file, the other non-test files naming it and the _test.go files naming it, sort -u | wc -l @ 46db35e0 → 85 -->
- **KAN-899 — Task 1.** Source: kan-820's panel F5, deferred with the fix named. **Model
  resolution** (`skills/flow/SKILL.md`) reads `SETTINGS_JSON="$(flow settings get 2>/dev/null ||
  true)"`: an unreachable store leaves an empty string, `jq -r '.reviewers[]'` over it prints
  nothing, and the empty `REVIEWERS` matches the roster table's "Reachable, list empty" row
  (`primary` alone) where the "Unreachable" row (`primary`, `principles`) applies — the discarded
  stderr is the only signal that tells them apart. `/flow-settings` already reports the stderr
  verbatim. Blast radius: one file.
- **KAN-900 — Task 2.** Source: kan-820's narrative — a rebase "textually clean but semantically
  not" dropped a `sync` import, "a broken build no exit code caught". Today a mid-implementation
  rebase does not happen (**2. Isolate the workspace**, `skills/flow/implement.md`), the review
  panel's clean rebase is followed by `flow.verify`, and `/flow-fast` re-runs its lint after any
  moving rebase. Integrate is the gap: `integrate-change.sh` runs only each `GUARD-TEST:` line
  `sync-onto-base.sh` prints after a clean rebase, and those name guard harnesses only — a clean
  rebase that breaks a Go or TypeScript build lands unverified. Blast radius: five files.
- **KAN-902 — Task 3.** Source: kan-826's run. Every shim that derives the root it judges from its
  own location and judges that checkout's text — `FLOW_GUARD_REPO_ROOT` from the script's parent
  in `check-ask-silence`, `check-dispatch-paragraphs`, `check-guard-symlinks`,
  `check-installed-citations`, `check-references`, `check-self-review-report`,
  `check-stage-mark-calls` and `check-verbatim-moves`, and `CHECK_PLAN_PROVENANCE_ROOT` in
  `check-plan-provenance` — judges the main checkout when the main checkout's copy is called by
  absolute path from a worktree. Blast radius: eleven files.
  <!-- measured: a scratch clone @ 46db35e0 with a linked worktree holding one unacknowledged new sentence, then one stale **No Such Section Anywhere** reference, in skills/flow/integrate.md → the main copy of check-verbatim-moves.sh called by absolute path from the worktree: 0 violations, its relative call: 1; check-references.sh absolute: exit 0, relative: exit 1 -->
- **KAN-905 — Task 4.** Source: kan-841's panel F2 and F11. The parity tests read the bash they
  were ported from at `d71a2327`, `ae805186` or `c5379c0a` through `git show`, and `t.Fatal` when
  the commit is absent. Blast radius: twelve test files.
  <!-- measured: git clone --depth 1 of this branch, then cd stats && go test ./internal/guard @ 46db35e0 → FAIL, 14 top-level tests (TestAsidePlanningArtifactsParity, TestBreakAndProve, TestThrowawayWorktreeCreate, TestProveReproducerExecBitCheck, TestProjectGetParity, TestCheckStageMarkCalls, TestCheckPanelFixSingleDispatch, TestCheckWorkspaceIsolation, TestCheckGuardSymlinks, TestCheckDispatchParagraphs, TestCheckFinishPreflight, TestThrowawayWorktreeRemoveFoldBack, TestProveReproducerBashParity, TestMutateAndVerify) -->
- **KAN-907 — Task 5.** Source: kan-841's panel F10. Both ignored-signal cases sleep half a
  second after sending SIGHUP and then look for its effect. `trapExitSignals`
  (`stats/internal/guard/provereproducer.go`) handles every signal it registered by cleaning up and
  dying of it, and its stop function drains any signal delivered before the run's own exit; the
  signal is pending before the test releases the blocked leg, so a handled SIGHUP always ends the
  run abnormally. An inherited ignore makes that exit 129, not death: `die` re-raises after
  `signal.Reset` has restored the ignore, so the re-raise does nothing, and the caller's exit waits
  on the handler, which sleeps and exits 129 after the run has printed both legs.
  `mutate-and-verify`'s finished-run assertion already fails on that; `prove-reproducer`'s checks
  only "not signaled" and the post-fix leg line, so it passes, and its 500 ms look was the only
  thing catching a handled ignored signal. Blast radius: two test files.
  <!-- measured: `if !signal.Ignored(s) {` replaced by `if true || !signal.Ignored(s) {` in trapExitSignals with the sleeps removed and no exit pin, go test ./internal/guard -run 'TestProveReproducerSignals|TestMutateAndVerify' in this worktree → TestMutateAndVerify/ignored_SIGHUP_mid-run FAIL 5 of 5, TestProveReproducerSignals/ignored_SIGHUP PASS 3 of 3; with Task 5's diff (exit pinned to 1) @ 384f6a7c at -count=10 → both FAIL 10 of 10, exit status 129; restored → ok, 10 of 10 -->
- **No red check on `origin/main`.** `go test ./internal/guard`, `bash scripts/test-integrate-change.sh`
  and `bash scripts/test-lib-flow-guard.sh` pass at `46db35e0`, so no task repairs one.

## Decisions

### Report the settings read's stderr and take the Unreachable row

**ID:** settings-read-stderr-visible
**Status:** active
**Chosen:** drop `2>/dev/null` from the read and add one sentence telling the run to report the stderr verbatim and resolve `REVIEWERS` by the Unreachable row, never as the empty list the failed read leaves.
**Considered:** capturing the exit code into a variable the table keys on — a second variable and a branch in a block the run executes by reading it, where the visible stderr already names the outage; keeping the redirect and adding a reachability probe — a second store call for what the first call's own stderr says.

### Run the lint list after every clean integrate rebase

**ID:** rebase-lint-after-clean-sync
**Status:** active
**Chosen:** `integrate-change.sh` runs `run-verify-list.sh <worktree> lint` after the scoped guard tests of every clean sync rebase — prepare's base-moved sync and land's re-sync — stopping `STOP: rebase-lint` with the list's report on a failure, and passing a project that declares no `## lint` list (`REBASE-LINT-NONE`). The lint list is where this repository compiles its Go (`go vet ./...`) and type-checks its SPA (`tsc -b`), and `/flow-fast` already re-runs lint after a moving rebase.
**Considered:** the whole `## test` list as well — reverses `scoped-reverify-not-full-suite`, the recorded choice to keep integrate off the full suite, for a defect class the build check catches; running lint only when the rebase overlapped this change's files — kan-820's break, and any upstream rename a new file of this change calls, come with no overlap; telling the run to lint in prose alone — integrate's sync is the script's, and a prose step beside a script that already reported success is the step that gets skipped.

### An own-checkout guard runs the calling worktree's copy

**ID:** guard-runs-cwd-worktree-copy
**Status:** active
**Chosen:** `flow_guard_cwd_checkout` in `scripts/lib/flow-guard.sh` execs `<top>/scripts/<the shim's name>` with the same arguments when the working directory's worktree `<top>` is not the shim's own checkout, shares its git common directory, and carries that shim; otherwise it returns. The eight guards that export `FLOW_GUARD_REPO_ROOT` to judge their checkout's text, and `check-plan-provenance`, call it before deriving their root. The worktree's own copy runs its own guard code over its own tree, exactly as the relative call does.
**Considered:** `check-installed-rules` and `check-hand-notes-in-step` too — their subject is the `$HOME` install, which tracks the main checkout, so the main copy judging it from a worktree is the right answer, and the worktree copy only reports a linked worktree has nothing to check; only `check-verbatim-moves`, the guard the ticket names — `check-references` reproduces the same false clean, and the other own-checkout guards share the root derivation; pointing the main copy's root at the worktree — the main checkout's guard binary would then judge the worktree's tree, a version skew the relative call never has; refusing the call with exit 2 — still makes the run remember the relative form; a redirect inside `flow_guard_exec` for every shim — guards that take explicit paths (`check-base-moved`, `sync-onto-base`) and the loop tools would switch code under integrate and the self-review loop for no defect; `check-normative-inventory.sh` too — its header states it is self-scoped from any cwd, and it prints an inventory, never a verdict a wrong tree could falsely clear.

### Skip a parity test whose pinned commit the clone lacks

**ID:** skip-parity-without-history
**Status:** active
**Chosen:** one `historicalFile(t, root, rev, path)` helper in `helpers_test.go` that `t.Skipf`s naming the commit when `git cat-file -e <rev>^{commit}` fails, and otherwise returns `git show <rev>:<path>`; every one of the eleven sites routes through it.
**Considered:** vendoring the pinned bash files into `testdata/` — about twenty copies of retired scripts to keep a shallow clone, where no clone of this repository runs shallow by default; fetching the commit inside the test — network access from a unit test; failing with a clearer message — the package still fails where only the history is missing.

### The finished-run assertion is the ignored-signal proof

**ID:** ignored-signal-no-timed-look
**Status:** active
**Chosen:** delete the 500 ms sleep and look from both tests, and pin `prove-reproducer`'s finished run to its own exit status, 1, with a comment naming the mechanism: the signal is pending before the leg is released, `trapExitSignals` drains any delivered signal before the run's own exit, and a handled one ends the run in exit 129, never in the run's own status.
**Considered:** deleting the look alone, trusting "not signaled" — a handled ignored SIGHUP exits 129 without dying of it, so `prove-reproducer`'s check passes under the mutation; a longer sleep — a wider tolerance on the same race; a barrier written by the released leg — a handler acts asynchronously, so the leg's look is timed too; reading the process's signal disposition — darwin's `ps` offers no ignored or caught column.
<!-- measured: ps -o sigignore=,sigcatch= -p <pid> on darwin @ 46db35e0 → "ps: sigignore: keyword not found" -->

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — Tasks 2 and 3 change what every integrate and every own-checkout guard call does.
**Considered:** flow-low — Tasks 1, 4 and 5 are a sentence or test plumbing, but the two script changes alter control flow a low-effort pass could under-read.

### One fixer group

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, Tasks 1–5 in order, one fixer dispatch (`self-review-backlog-flow-10-fix`): every task is a `git apply` of a diff verified in a scratch clone, and the files are disjoint.
**Considered:** one group per task — five dispatches for work one fixer finishes in series; a scripts group and a Go group — Tasks 4 and 5 are test-only and short.

## Open questions

## Live check

No running service or persistent state is touched: Task 2 changes `integrate-change.sh`, a script
each integrate runs from the checkout; Task 3 changes guard shims run per call; Tasks 4 and 5 are
tests and Task 1 is prose.
