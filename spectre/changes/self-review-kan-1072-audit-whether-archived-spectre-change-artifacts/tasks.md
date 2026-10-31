# self-review-kan-1072-audit-whether-archived-spectre-change-artifacts

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** close kan-1072's four self-review findings — F1 (Task 3, its history repair Task 1), F2
(Task 4), F3 (Task 5), F4 (Task 6), as `proposal.md` names them — and the verify blocker planning
found (Task 2).

**Architecture:** a history re-split in the kan-1072 worktrees; a git-history read of removed
acknowledgement files in `check-verbatim-moves`; narrower pathspecs in `commit-split`; one Markdown
key fix; one flag default; one new plan sub-guard.

**Tech Stack:** Go (`stats/`), Bash, git, Markdown.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-31-self-review-kan-1072-audit-whether-archived-spectre-change-artifacts-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject, no `Task-Id:` trailer and no `Co-Authored-By:` trailer. Never write `spectre/changes/` save Task 1's re-split. Never merge or push.
- Every task but Task 1 runs in `/Users/tweety53/Projects/agents-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts` (`<wt>`), branch `spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`; `**Files:**` paths are relative to `<wt>`. Never touch a main checkout, and never create a worktree.
- Task 1 runs first, before any other commit of this change exists.
- Go: `gofmt -w` every Go file you change; `cd stats && gofmt -l .` prints nothing and `cd stats && make -s web-dist && go vet ./...` exits 0.
- Every prose edit to `skills/` lands through the edit script this plan names, byte for byte: this change's `verbatim-moves.txt` already lists each sentence as `scripts/check-verbatim-moves.sh` prints it, so a reworded sentence fails that guard.
- Mutation proofs restore with `scripts/break-and-prove.sh` — never `git checkout --` on a file carrying uncommitted edits.
- Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database, and never run tests against the live database.
- Every Go test run sets a Bash `timeout` of 600000.

## Review Focus

- Task 1: every rewritten tip is tree-identical to its old tip; the plan commit carries only the change folder; nothing was pushed; the `NEW-FIX-BASE:` and `PUSH:` lines are right.
- Task 2: only a `<leaf>/changes/<dir>/verbatim-moves.txt` the base lacks and the branch removed is read; an archive entry's list stays unread.
- Task 3: a `-fix-<N>` folder stays planning; another change's folder, any archive path and every foreign `link.md` land in the implementation commit; the tracked-symlink exit 128 holds.
- Task 6: a read-only citation in an edit step is never reported; the rule matches the measured history.

- [ ] 1. Re-split the kan-1072 branches so the archive removal is its own commit
  - [ ] **Step 1: Check the starting point; do not guess.** In each worktree below, `git -C <W> status --porcelain -- spectre/changes` prints nothing and `git -C <W> diff --quiet HEAD` exits 0; `git -C <W> rev-parse HEAD` equals `git -C <W> rev-parse origin/spectre/kan-1072-audit-whether-archived-spectre-change-artifacts` for spectre, gymie-playwright, flow-eval and gymie-frontend, and for agents `HEAD` is this plan's commit with `45aac32047631607d79bcd1e5c18b8bd5045cd99` as its parent. Any mismatch: stop and report it; never rewrite a branch that moved.
  - [ ] **Step 2: Re-split agents.** Run the block below from any directory. It reuses the implementation and plan commits' messages, makes the removal its own commit, replays every commit above the old plan commit (the archive commit and this plan's commit), and proves the tip tree unchanged.

```bash verified:rehearsed on a scratch clone of spectre/kan-1072-audit-whether-archived-spectre-change-artifacts (git clone -b) at 45aac320 — exit 0, tip tree identical, the check block below printed NEW-FIX-BASE
set -euo pipefail
N=kan-1072-audit-whether-archived-spectre-change-artifacts
W=/Users/tweety53/Projects/agents-worktrees/$N
BASE=84e5e25bf9995e8bed250a1b9ac86aa31214c522
IMPL=f2e1a72fecfb592ba8888bfd1502eb65844872cd
PLAN=1a72adc9d6b48a7ba1c23daa81e851ac99eccac8
ARCH=45aac32047631607d79bcd1e5c18b8bd5045cd99
git -C "$W" diff --quiet HEAD
test -z "$(git -C "$W" status --porcelain -- spectre/changes)"
OLD=$(git -C "$W" rev-parse HEAD)
git -C "$W" merge-base --is-ancestor "$ARCH" "$OLD"
echo "OLD-TIP: $OLD"
git -C "$W" reset -q --keep "$PLAN"
git -C "$W" reset -q --soft "$BASE"
git -C "$W" reset -q -- spectre/changes/
git -C "$W" add -- "spectre/changes/$N/link.md"
git -C "$W" commit -q -C "$IMPL"
git -C "$W" add -A -- spectre/changes/ ":(exclude)spectre/changes/$N"
git -C "$W" commit -q -m 'chore(spectre): remove the change archive' -m 'Flow-Key: flow.commit-two'
git -C "$W" add -A -- "spectre/changes/$N"
git -C "$W" commit -q -C "$PLAN"
git -C "$W" cherry-pick "$PLAN..$OLD" >/dev/null
git -C "$W" diff --quiet "$OLD" HEAD
```

  - [ ] **Step 3: Check agents.** Run the block below; it exits 0 and prints `NEW-FIX-BASE: <sha>`, the rewritten archive commit, which the report carries.

```bash verified:run on the same scratch clone after Step 2's block — exit 0, printed NEW-FIX-BASE
set -euo pipefail
N=kan-1072-audit-whether-archived-spectre-change-artifacts
W=/Users/tweety53/Projects/agents-worktrees/$N
BASE=84e5e25bf9995e8bed250a1b9ac86aa31214c522
sub() { git -C "$W" log --format='%H %s' "$BASE..HEAD" | awk -v s="$1" 'substr($0, 42) == s { print $1 }'; }
I=$(sub 'feat(flow): drop the change directory at land instead of archiving it')
R=$(sub 'chore(spectre): remove the change archive')
P=$(sub "chore(spectre): plan $N")
A=$(sub "chore(spectre): archive $N")
for c in "$I" "$R" "$P" "$A"; do test "$(printf '%s\n' "$c" | wc -l)" -eq 1 && test -n "$c"; done
test "$(git -C "$W" show --name-status --format= "$R" | grep -vc $'^D\tspectre/changes/archive/')" = 0
test "$(git -C "$W" show --name-only --format= "$P" | grep -vc "^spectre/changes/$N/")" = 0
test "$(git -C "$W" show --name-only --format= "$I" | grep '^spectre/changes/' | grep -vc "^spectre/changes/$N/link.md\$")" = 0
echo "NEW-FIX-BASE: $A"
```

  - [ ] **Step 4: Re-split the satellites.** Run the block below. gymie (`/Users/tweety53/Projects/gymie-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts`) is already one `chore(spectre): remove the change archive` commit and is not touched. Commit hooks run as configured; a hook refusal is reported, never bypassed.

```bash verified:rehearsed on scratch clones of the four satellite branches (git clone -b) — exit 0, every tip tree identical
set -euo pipefail
N=kan-1072-audit-whether-archived-spectre-change-artifacts
P=/Users/tweety53/Projects
# spectre: the skill re-copy under its pre-reshape subject, then the archive removal; no plan commit.
W=$P/spectre-worktrees/$N
test "$(git -C "$W" rev-parse HEAD)" = 10eb27f3b8cedbded8ae900719f3b03de8d778a8
git -C "$W" diff --quiet HEAD
git -C "$W" reset -q --soft 8749ce5e8c60b81bfc49d8b4fa88286921148019
git -C "$W" reset -q -- spectre/changes/
git -C "$W" commit -q -m 'docs(spectre-planning): re-copy the planning skill — the rejection search reads history' -m 'Flow-Key: flow.commit-two'
git -C "$W" add -A -- spectre/changes/
git -C "$W" commit -q -m 'chore(spectre): remove the change archive' -m 'Flow-Key: flow.commit-two'
git -C "$W" diff --quiet 10eb27f3b8cedbded8ae900719f3b03de8d778a8 HEAD
# gymie-playwright: the two commits become the one removal commit.
W=$P/gymie-playwright-worktrees/$N
test "$(git -C "$W" rev-parse HEAD)" = 0a3bacdec482be82c3341e937af27adab739b765
git -C "$W" diff --quiet HEAD
git -C "$W" reset -q --soft dda8d21502ceca7cb6441602f5050132ba42de7d
git -C "$W" commit -q -C 8b1812d148a4bf9f0194b8c69a33f3553e28bd98
git -C "$W" diff --quiet 0a3bacdec482be82c3341e937af27adab739b765 HEAD
# flow-eval: the plan commit becomes the removal commit.
W=$P/flow-eval-worktrees/$N
test "$(git -C "$W" rev-parse HEAD)" = 324746b0dc1bc27260782a34c40c951426b3ff04
git -C "$W" diff --quiet HEAD
git -C "$W" reset -q --soft 4a2a1b986a5b501365205edcbbf667fa00e3c3d6
git -C "$W" commit -q -m 'chore(spectre): remove the change archive' -m 'Flow-Key: flow.commit-two'
git -C "$W" diff --quiet 324746b0dc1bc27260782a34c40c951426b3ff04 HEAD
# gymie-frontend: the two commits become the one removal commit.
W=$P/gymie-frontend-worktrees/$N
test "$(git -C "$W" rev-parse HEAD)" = 367a2cd44d3b980d793a9d0cb5fd1a08af67bd9a
git -C "$W" diff --quiet HEAD
git -C "$W" reset -q --soft c2d98593f3096fffead89012d155c1d79c1fcdd4
git -C "$W" commit -q -C 90f425a11b539f3ed7388d05528c8c4874e8ce45
git -C "$W" diff --quiet 367a2cd44d3b980d793a9d0cb5fd1a08af67bd9a HEAD
```

  - [ ] **Step 5: Check the satellites.** For each of spectre, gymie-playwright, flow-eval and gymie-frontend, `git -C <W> log --format='%s' <base>..HEAD` lists exactly the subjects the block's comment names (spectre: the re-copy subject, then `chore(spectre): remove the change archive`; the other three: `chore(spectre): remove the change archive` alone), and `git -C <W> show --name-status --format= <removal sha> | grep -vc $'^D\tspectre/changes/archive/'` prints `0`, `<base>` the sha each block soft-resets to.
  - [ ] **Step 6: Report, never push.** The report carries `NEW-FIX-BASE: <sha>` from Step 3 and one line per rewritten repository, each its old tip as the lease:
    - `PUSH: git -C /Users/tweety53/Projects/agents-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts push --force-with-lease=spectre/kan-1072-audit-whether-archived-spectre-change-artifacts:<agents OLD-TIP from Step 2> origin spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`
    - `PUSH: git -C /Users/tweety53/Projects/spectre-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts push --force-with-lease=spectre/kan-1072-audit-whether-archived-spectre-change-artifacts:10eb27f3b8cedbded8ae900719f3b03de8d778a8 origin spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`
    - `PUSH: git -C /Users/tweety53/Projects/gymie-playwright-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts push --force-with-lease=spectre/kan-1072-audit-whether-archived-spectre-change-artifacts:0a3bacdec482be82c3341e937af27adab739b765 origin spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`
    - `PUSH: git -C /Users/tweety53/Projects/flow-eval-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts push --force-with-lease=spectre/kan-1072-audit-whether-archived-spectre-change-artifacts:324746b0dc1bc27260782a34c40c951426b3ff04 origin spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`
    - `PUSH: git -C /Users/tweety53/Projects/gymie-frontend-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts push --force-with-lease=spectre/kan-1072-audit-whether-archived-spectre-change-artifacts:367a2cd44d3b980d793a9d0cb5fd1a08af67bd9a origin spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`

    The session runs them, in that order, before the landing push (decision `repair-first-new-fix-base`) and uses `NEW-FIX-BASE` as `<fix-base>` from here on.

**Build:** green
**Files:** none
**Tests:** none
**Regression:** none — a history repair; Step 3's and Step 5's checks are its proof, and Task 3's test pins the split that wrote the defect.
**Commit:** chore(spectre): remove the change archive
**After:** none

**Decision:** repair-resplits-six-branches, repair-first-new-fix-base, archive-removal-commit-trailer

- [ ] 2. Read the acknowledgements of a change directory the branch removed
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** `cd <wt> && scripts/check-verbatim-moves.sh` exits 1 with FAIL lines for kan-1072's own prose edits, every one listed in the kan-1072 change's archived `verbatim-moves.txt`, which `vmAcks` no longer reads.
    <!-- measured: cd <wt> && scripts/check-verbatim-moves.sh — 178 FAIL lines, all present in 45aac320^:spectre/changes/kan-1072-audit-whether-archived-spectre-change-artifacts/verbatim-moves.txt @ 45aac320 -->
  - [ ] **Step 2: Write the test.** `TestCheckVerbatimMovesReadsAcksOfARemovedChange` in `stats/internal/guard/check_verbatim_moves_test.go`, on `vmRepo` and the fixture `TestCheckVerbatimMovesReadsLiveAcksOnly` uses (`vmPipe`, `vmMark`). Cases, each committing on a branch over the fixture's base: (a) a commit removing the `vmMark` sentence from `vmPipe` and adding `spectre/changes/demo/verbatim-moves.txt` holding `vmMark`, then a commit moving `spectre/changes/demo` to `spectre/changes/archive/demo` (`git mv`) — exit 0 and `ok   acknowledged removal`; (b) the same with the second commit a `git rm -r spectre/changes/demo` — exit 0; (c) `spectre/changes/old/verbatim-moves.txt` holding `vmMark` committed in the base, the branch removing the sentence and deleting that file — exit 1 and the FAIL line (a list the base already held acknowledges nothing). Run `cd stats && go test ./internal/guard -run 'TestCheckVerbatimMovesReadsAcksOfARemovedChange' -count=1` — cases (a) and (b) fail. unverified:confirm how `vmRepo` commits its base and what `vmRun`'s default base resolves to; pass the base as `vmRun`'s argument if the default cannot see the branch commits.
  - [ ] **Step 3: Implement.** In `stats/internal/guard/verbatimmoves.go`, per decision `acks-of-removed-change-dirs`, give `vmAcks` the guard's `git` and `base` and add the history read: `capture(git("-C", root, "log", "--no-renames", "--diff-filter=D", "--format=%H", "--name-only", base+"..HEAD", "--", leaf+"/changes/"))`; for each deleted path matching `^<leaf>/changes/[^/]+/verbatim-moves\.txt$` whose `git cat-file -e <base>:<path>` fails, read `git show <sha>^:<path>` and add its lines exactly as the file loop does (comments skipped, a leading `\` dropped, whitespace folded). A failed `git log` adds nothing. Replace the doc comment's sentence "Nothing in the changes tree's archive directory is read: run 1 drops a landed change's directory, acknowledgements included, and an archive left from before holds acknowledgements for a diff that has already landed." with "Nothing in the changes tree's archive directory is read from the tree: an archive left from before holds acknowledgements for a diff that has already landed. A change directory this branch itself removed since the base — archived or dropped on the branch — is read from the commit before its removal, so the change's own self-review loop still sees its list (kan-1072 self-review)." In `scripts/check-verbatim-moves.sh`'s header, after "— run 1 drops a change's directory, its list with it, once it has landed.", add the line "# A change directory the branch itself removed since the base is read from the commit before its removal (kan-1072 self-review)."
  - [ ] **Step 4: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, make the history read add nothing — cases (a) and (b) fail; restore.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCheckVerbatimMoves' -count=1` passes; `cd <wt> && scripts/check-verbatim-moves.sh` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/verbatimmoves.go`, `stats/internal/guard/check_verbatim_moves_test.go`, `scripts/check-verbatim-moves.sh`
**Tests:** `TestCheckVerbatimMovesReadsAcksOfARemovedChange`
**Regression:** `TestCheckVerbatimMovesReadsAcksOfARemovedChange` fails — a change's own acknowledgements vanish once its directory is archived or dropped on the branch, and the self-review verify goes red on sentences the change acknowledged.
**Commit:** fix(stats): read the acknowledgements of a change directory the branch removed
**After:** Task 1

**Decision:** acks-of-removed-change-dirs

- [ ] 3. Scope the plan commit to the change's own folders
  - [ ] **Step 1: Write the test.** `TestCommitSplitScopesThePlanCommitToTheChangeFolder` in `stats/internal/guard/commit_split_test.go`, on `splitRepo` with a seed holding `spectre/changes/archive/old/tasks.md`, `spectre/changes/archive/old/link.md`, `spectre/changes/other/proposal.md` and `README.md`. On the `spectre/demo` worktree: delete both `archive/old` files, edit `other/proposal.md`, add `src/a.go`, `spectre/changes/demo/tasks.md`, `spectre/changes/demo/link.md`, `spectre/changes/demo-fix-1/tasks.md` and `spectre/changes/demo-fixer/tasks.md`; run `fx.split`. The implementation commit's files are exactly `src/a.go`, both `archive/old` paths, `other/proposal.md`, `demo/link.md` and `demo-fixer/tasks.md`; the plan commit's are exactly `demo/tasks.md` and `demo-fix-1/tasks.md`. Run `cd stats && go test ./internal/guard -run 'TestCommitSplit' -count=1` — the new test fails. Any existing `TestCommitSplit` case asserting that a path outside the change's own folders (for instance an edit of the seed's `spectre/changes/seed.md`) lands in the plan commit is updated to the implementation commit, each named in the report.
  - [ ] **Step 2: Implement.** In `stats/internal/guard/commitsplit.go`, per decision `split-scopes-change-folder`: keep the `reset -q -- <leaf>/changes/` step; the exclude-add becomes `add -A -- . ':(exclude)<leaf>/changes/<name>'` plus one `':(exclude)<leaf>/changes/<dir>'` per entry of `os.ReadDir(<worktree>/<leaf>/changes)` that `isFixOf(<dir>, name)` accepts; the `link.md` add becomes `add -A -- <leaf>/changes/<name>/link.md`, run when `os.Lstat` finds that file; delete `hasChangeLinkMD`. Reword the comments the change leaves stale. Then run `python3 -I .superpowers/sdd/prose-edits-f1.py <wt>` (it asserts each replacement matched once), and in `scripts/commit-split.sh` replace the two header paragraphs opening "THE PLANNING PATH INSIDE THE SPECTRE TREE" and "A CHANGE DIRECTORY'S `link.md`" with the block below, and in the tracked-symlink paragraph change `':(exclude)spectre/changes/'` to `':(exclude)spectre/changes/<name>'`.

```bash verified:authored in-tree for this change; a tracked spectre/changes symlink makes `git add -A -- . ':(exclude)spectre/changes/demo'` exit 128 (git 2.50.1 spike)
# THE PLANNING PATH IS THE CHANGE'S OWN FOLDER, `spectre/changes/<name>/`,
# WITH EACH `<name>-fix-<N>/` SUB-CHANGE FOLDER THE WORKTREE HOLDS — NOT
# `spectre/changes/` AND NOT `spectre/`. Every other path under
# `spectre/changes/` — another change's directory, an archive deletion — is
# implementation and lands in the implementation commit (kan-1072
# self-review, F1). A capability spec under `spectre/specs/` is
# implementation too — the implementer writes and commits it on the change
# branch, in the task commit that implements the requirement — so widening
# the pathspec back to `spectre/` would classify it as planning again. See
# skills/flow-contracts/git-boundaries.md, canonical for that boundary.
#
# THE CHANGE'S OWN `link.md` IS ALSO IMPLEMENTATION, CARVED OUT BY FILE
# RATHER THAN BY DIRECTORY. It lives at `spectre/changes/<name>/link.md`,
# inside the planning exclusion, so it gets its own `git add` after the
# exclude-add when it exists, and a change without one stays a no-op rather
# than a failure — stats/internal/guard/commitsplit.go carries the pathspec
# reasoning beside the code.
```

  - [ ] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, put the exclude back to the whole `<leaf>/changes/` tree — the new test fails; restore.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestCommitSplit|TestReshapeBranch|TestIntegrateChange' -count=1` passes; `bash -n scripts/commit-split.sh` exits 0; `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `stats/internal/guard/commitsplit.go`, `stats/internal/guard/commit_split_test.go`, `scripts/commit-split.sh`, `skills/flow-contracts/git-boundaries-commit-chain.md`
**Tests:** `TestCommitSplitScopesThePlanCommitToTheChangeFolder`
**Regression:** `TestCommitSplitScopesThePlanCommitToTheChangeFolder` fails — an archive deletion or another change's folder folds into the plan commit again.
**Commit:** fix(stats): scope the plan commit to the change's own folders
**After:** Task 2

**Decision:** split-scopes-change-folder

- [ ] 4. Name a group implementer's key by its joined task ids
  - [ ] **Step 1: Edit.** Run `python3 -I .superpowers/sdd/prose-edits-f2.py <wt>` from `<wt>`: in `skills/flow/implement.md` it rewrites the roles-table implementer key, both `flow record dispatch` calls, the `-key` sentence after them and the prompt stub's `Key:` line to `task-<ids>-implementer` (and `-task <ids>`), per decision `implementer-key-ids`. It asserts each replacement matched once.
  - [ ] **Step 2: Verify.** `git -C <wt> grep -n -F 'task-<n>-implementer' -- skills/flow/implement.md | grep -v -e 'implementer-recommit-<k>' -e 'implementer…'` prints nothing; `scripts/run-verify-list.sh <wt> lint` exits 0. Commit.

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** none
**Regression:** none — prose; the hook's `ROW_KEY` already refuses the old shape at dispatch.
**Commit:** fix(flow): name a group implementer's key by its joined task ids
**After:** Task 3

**Decision:** implementer-key-ids

- [ ] 5. Give the self-review bundle its own default deadline
  - [ ] **Step 1: Write the test.** `TestRunSelfReviewBundleWaitsPastTheStoreDeadline` in `stats/cmd/flow/selfreview_test.go`, on `TestRunSelfReviewBundlePrintsBundle`'s fixture (`gitRepo`, `isolatedStateRoot`, `genuineDaemon`): the handler sleeps `defaultTimeout + 500*time.Millisecond` on a path ending `/bundle` before answering 200 with a bundle body; run `self-review bundle -addr <srv> -C <repo> -change demo` with no `-timeout` — exit 0 and the body on stdout. Run `cd stats && go test ./cmd/flow -run 'TestRunSelfReviewBundleWaitsPastTheStoreDeadline' -count=1` — it fails with a deadline error.
  - [ ] **Step 2: Implement.** In `stats/cmd/flow/selfreview.go`, per decision `bundle-default-deadline`: `const selfReviewBundleTimeout = time.Minute`, with a comment saying the bundle is assembled from git history on the daemon's side and has no fallback, so the store-write fallback deadline does not apply (kan-1072 self-review, F3); in `runSelfReviewBundle`, right after `registerRecordConnFlags(fset, &f)`, set `f.timeout = selfReviewBundleTimeout` and `fset.Lookup("timeout").DefValue = selfReviewBundleTimeout.String()`.
  - [ ] **Step 3: Prove the test pins the fix.** With `scripts/break-and-prove.sh`, drop the `f.timeout` assignment — the test fails; restore.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l cmd/flow` prints nothing; `cd stats && go test ./cmd/flow -run 'TestRunSelfReviewBundle' -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/selfreview.go`, `stats/cmd/flow/selfreview_test.go`
**Tests:** `TestRunSelfReviewBundleWaitsPastTheStoreDeadline`
**Regression:** `TestRunSelfReviewBundleWaitsPastTheStoreDeadline` fails — the bare bundle call times out at the store-write deadline again.
**Commit:** fix(stats): give flow self-review bundle its own default deadline
**After:** Task 1

**Decision:** bundle-default-deadline

- [ ] 6. Refuse a path an edit step names outside its task's Files
  - [ ] **Step 1: Write the tests.** `TestPlanStepFiles` in `stats/internal/guard/plan_step_files_test.go`, on `TestPlanScriptPaths`'s table shape (a temp root holding the files a case names, the plan at `spectre/changes/c/tasks.md`). Refused (exit 1, one `file:line:` line naming the task and the path): an `Implement` step reading "Delete `a/x.go`. `keep.go`'s `f` drops its entry." with `**Files:**` `a/x.go` and `a/keep.go` on disk (names `a/keep.go`); an `Edit` step "Rewrite each header sentence, in: `s/one.sh`, `s/two.sh`." with `**Files:**` `s/one.sh` (names `s/two.sh`); an `Edit` step "`d/a.md`: x. `d/b.md` (y), `c.md`: each follows." with `**Files:**` `d/a.md`, `d/b.md` and `d/c.md` on disk (names `d/c.md`). Passing (exit 0): every edited path in `**Files:**`; a path an `**Allowed-collateral:**` glob covers; a `Verify` step running `tools/x.sh` with `tools/x.sh` on disk; an `Implement` step "copy the rows from the table in `d/c.md` (`d/c.md`)"; an edit position naming a path that does not exist; a fenced example; a plan under `spectre/changes/archive/c/`. One more case runs `checkPlanShape` on the first refused plan and expects exit 1 with the same line, proving the wiring. Run `cd stats && go test ./internal/guard -run 'TestPlanStepFiles' -count=1` — it fails.
  - [ ] **Step 2: Implement.** Add `stats/internal/guard/planstepfiles.go`, registered as `plan-step-files`, on `planScriptPathsCmd`'s shape (argument handling, a plan outside a `changes` directory passes, the root two levels above `changes`), implementing the rule in the brainstorm design file's F4 section exactly (decision `plan-step-files-subguard`; `.superpowers/sdd/plan-step-files-prototype.py` is its reference reading): tasks through `tcfIterTasks` and `tcfParseTask`, fences through `tcfUnfenced`, the cover test `tcfCheckFiles(task, []string{path})` returning nothing. Each finding prints `<file>:<line>: task <id> step <n> edits `<path>`, which no **Files:** or **Allowed-collateral:** entry covers — add it to **Files:** when this task's commit changes it, else reword the step` and the guard exits 1. In `stats/internal/guard/planshape.go`, append `"plan-step-files"` to `psCheckOne`'s list and add it, with a one-line description, to the header's list of sub-guards. Mark the one measured false-positive shape with a `ponytail:` comment: a fixture path written into a test's scratch repository in an edit step reads as an edit — upgrade path, a step-local exclusion when a second plan hits it.
  - [ ] **Step 3: Prove the tests pin the fix.** With `scripts/break-and-prove.sh`, (a) remove `"plan-step-files"` from `psCheckOne`'s list — the wiring case fails; (b) make the edit-position test always true — the read-context case fails; restore each.
  - [ ] **Step 4: Measure against history.** Run the block below from `<wt>`. It prints exactly seven lines: kan-921 task 2 naming `stats/internal/guard/check_guard_symlinks_test.go` and `stats/internal/guard/check_stage_mark_calls_test.go`; self-review-kan-922 task 2 naming `scripts/check-self-review-report.sh`; and k1072 (the kan-1072 plan) task 6 naming `stats/internal/guard/specroot.go`, task 10 naming `scripts/integrate-change.sh` and `scripts/check-task-commit-fields.sh`, task 13 naming `skills/flow/verify-fix-loop.md`. A difference is traced: a line the rule as designed yields — the Go cover test is `tcfCheckFiles`'s exact match, where the prototype also accepted a path under a Files directory — is named in the report with its verdict; any other difference is fixed in the rule until the output matches.
    <!-- measured: python3 -I .superpowers/sdd/plan-step-files-prototype.py over the archived plans of 84e5e25b and the kan-1072 tasks.md of 45aac320 — 3 hits in 2 plans, and 4 hits in kan-1072 @ 84e5e25b -->

```bash unverified:runs only once plan-step-files exists; confirm the seven lines
S=$(mktemp -d)
git -C . archive 84e5e25bf9995e8bed250a1b9ac86aa31214c522 | tar -x -C "$S"
for d in "$S"/spectre/changes/archive/*/; do mv "$d" "$S/spectre/changes/"; done
mkdir -p "$S/spectre/changes/k1072"
git -C . show HEAD:spectre/changes/archive/kan-1072-audit-whether-archived-spectre-change-artifacts/tasks.md > "$S/spectre/changes/k1072/tasks.md"
for f in "$S"/spectre/changes/*/tasks.md; do scripts/check-plan-shape.sh "$f" 2>/dev/null; done | grep -F 'which no **Files:** or **Allowed-collateral:** entry covers'
rm -rf "$S"
```

  - [ ] **Step 5: Verify.** `cd stats && gofmt -l internal/guard` prints nothing; `cd stats && go test ./internal/guard -run 'TestPlanStepFiles|TestCheckPlanShape|TestPlanScriptPaths' -count=1` passes; `scripts/check-plan-shape.sh` exits 0 (every live plan of the branch, this one included). Commit.

**Build:** green
**Files:** `stats/internal/guard/planstepfiles.go`, `stats/internal/guard/plan_step_files_test.go`, `stats/internal/guard/planshape.go`
**Tests:** `TestPlanStepFiles`
**Regression:** `TestPlanStepFiles` fails — a step-edited path outside `**Files:**` passes plan shape and surfaces only when the task's commit is refused.
**Commit:** feat(stats): refuse a path an edit step names outside its task's Files
**After:** Task 1

**Decision:** plan-step-files-subguard
