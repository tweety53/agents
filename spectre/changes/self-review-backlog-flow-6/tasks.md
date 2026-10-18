# self-review-backlog-flow-6

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the live Backlog Flow findings — KAN-702's hand-picked satellite merge base in the
cross-repo visual trigger, KAN-740's waived live step that silently drops its coverage, KAN-802's
server-clock exception that ignores a server-side pin, KAN-803's rewritten baselines staged unread,
and KAN-733's non-atomic friend check in gymie's `setLevel`.

**Architecture:** four prose additions where their readers already look — the verify stage, the
verifier prompt, the sweep and the fix loop — and one backend change in gymie: a locking read of the
accepted friendship inside `setLevel`'s transaction, and `unfriend` taking the same lock first.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-18-self-review-backlog-flow-6-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Each task names the repository it runs in, by its worktree; every `**Files:**` path is relative to that worktree's root. Tasks 1–4 run in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-6`, branch `spectre/self-review-backlog-flow-6`; Task 5 runs in gymie's `/Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-6`, branch `self-review-backlog-flow-6`, which its Step 1 creates. Never touch a main checkout.
- Tasks 1–4's diffs are the unified diffs their fix steps give, applied with `git apply` from the worktree's root, checked in order in a scratch clone at `cfb9b564`. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Task 5 is test-first: its concurrency tests fail before the fix and pass after it.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- The KAN-702 sentence binds `<merge-base>` to the working notes, the source **Inline verify** in the same file already names, and never to the state file's `worktrees` map.
- KAN-802's sentence narrows the server-clock exception without restating the browser-pin rule above it.
- Task 5: the friend check and the grant write share one transaction holding the friendship row's lock, and `unfriend` locks that row before it deletes the grants — either order serialises, so no grant outlives its friendship.

- [x] 1. Bind the cross-repo visual trigger's merge base to the working notes (KAN-702)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "own merge base from this run's working notes" skills/flow/verify-and-handoff.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ cfb9b564 → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at cfb9b564; check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0, and check-verbatim-moves.sh cfb9b564 fails only on the sentence verbatim-moves.txt lists
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index b6721b17..532f4b43 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -218,6 +218,10 @@ this pipeline restates it. Resolve once per worktree in this run's resolved set,
    Its exit 0 makes the canonical worktree survive this step as the canonical worktree's own exit 0
    does.
 
+   `<merge-base>` in both calls is that worktree's own merge base from this run's working notes —
+   never `origin/<base>` or another ref picked by hand, which matches every path the base gained
+   since the worktree was cut as this change's own (KAN-702).
+
    **On a fix run, both calls read the fix run's own diff.** A worktree that has a since-close sha this run — the one **The late-fix reduction** (`skills/flow/review-panel-late-fix.md`) read for it — feeds them `git -C <worktree> diff --name-only <since-close-sha>..HEAD` in place of `<merge-base>..HEAD`, so a fix run whose delta touches no `ui paths` skips this stage at step 2 as exit 1 does.
    A worktree whose since-close sha is `-` keeps `<merge-base>..HEAD`, and so does every worktree when the change's newest verifier dispatch row (the newest row `flow record dispatches -change <name>` returns with `role` `verifier` and a `key` starting with `visual-verify` — the rows `<agents repo>/scripts/check-visual-verify-dispatched.sh` counts) does not have `outcome` `completed` — an earlier run's UI change that was never verified may lie before the since-close sha.
 
````

  - [x] **Step 3: Verify.** `grep -c "own merge base from this run's working notes" skills/flow/verify-and-handoff.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ cfb9b564 → 1 -->
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — a prose sentence; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): bind the cross-repo visual trigger's merge base to the working notes
**After:** none

**Decision:** visual-trigger-working-notes-base

- [x] 2. Run a waived live-check step on its cheapest safe proxy (KAN-740)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -F "cheapest safe proxy" skills/flow/verify-and-handoff.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ cfb9b564 after the earlier tasks' diffs → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at cfb9b564 after the earlier tasks' diffs; check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0, and check-verbatim-moves.sh cfb9b564 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index 532f4b43..55059c8a 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -148,7 +148,10 @@ removed with `git worktree remove` once the figure is written — never by check
 over the change worktree. The `## Report` gains one line, `live check — <matched | failure look>`. A figure that
 shows the failure look is the branch's own defect and takes **The loop**
 (`skills/flow/verify-fix-loop.md`); a stack that will not start ends the turn as a failed
-command's environment cause does. The verifier row's `end` (**Recording.**, above) is recorded
+command's environment cause does. **A step of the check the operator waives runs on its cheapest safe proxy instead** — a scratch
+database, a throwaway container — never simply dropped, and the `## Report` line names the proxy:
+kan-576's waived V34 DOWN-block step, run by a reviewer on a scratch database, caught a rollback
+that would have kept every `WORKOUTS: READ` grant as a full one (KAN-740). The verifier row's `end` (**Recording.**, above) is recorded
 after this check, not before it: `-outcome blocked -cause test-failure` on a failure look,
 `-cause environment` on a stack that will not start. The stack stays up for **Visual
 verification** and **Resolve the run instructions**.
````

  - [x] **Step 3: Verify.** `grep -c -F "cheapest safe proxy" skills/flow/verify-and-handoff.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ cfb9b564 → 1 -->
  - [x] **Step 4: Commit** the files with the subject below.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): run a waived live-check step on its cheapest safe proxy
**After:** Task 1

**Decision:** waived-live-step-proxy

- [x] 3. Pin the server's clock where the checkout can (KAN-802)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -F "A server-side pin narrows that exception" skills/flow/visual-verify-verifier.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ cfb9b564 after the earlier tasks' diffs → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at cfb9b564 after the earlier tasks' diffs; check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0, and check-verbatim-moves.sh cfb9b564 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index b1e6f292..5bac5327 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -74,6 +74,9 @@ never an exit code the command did not return.
    derives from its own clock** — a window, an age, a "today" computed server-side — which no
    browser clock can pin: such a spec leaves the clock real, its dates relative to the real day,
    and its header says so and why (gymie-playwright's body-calories spec).
+   **A server-side pin narrows that exception.** Where the checkout can send the server its own
+   today (gymie-playwright's `pinClock` sends `X-Gymie-Today`), the spec pins both ends to the one
+   instant, and leaves real or masks only what neither pin reaches (KAN-802).
 
    **Every mockup frame this change adds or updates gets a capture and a sidecar line — never
    optional, with `mockups` declared.** The frames are the change's declared list (step 10) plus
````

  - [x] **Step 3: Verify.** `grep -c -F "A server-side pin narrows that exception" skills/flow/visual-verify-verifier.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ cfb9b564 → 1 -->
  - [x] **Step 4: Commit** the files with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): pin the server's clock beside the browser's where the checkout can
**After:** Task 2

**Decision:** server-clock-pin

- [x] 4. Read every rewritten baseline before staging it (KAN-803)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -F "No rewritten baseline is staged unread" skills/flow-contracts/known-bugs.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ cfb9b564 after the earlier tasks' diffs → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at cfb9b564 after the earlier tasks' diffs; check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0, and check-verbatim-moves.sh cfb9b564 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow-contracts/known-bugs.md b/skills/flow-contracts/known-bugs.md
index 88a197c4..d98a5426 100644
--- a/skills/flow-contracts/known-bugs.md
+++ b/skills/flow-contracts/known-bugs.md
@@ -35,7 +35,11 @@ change's own app code can move its pixels. Any other mismatch, a `verify` run's
 branch's own defect and takes The loop. Whoever ran it recaptures exactly the failing tests — `capture`
 with `<spec>` substituted by each failing test's `<file>:<line>` — stages only the PNGs the failing
 run named, restores every other file the capture rewrote, commits them on the change branch as their
-own `test(fidelity): re-baseline <spec basenames>` commit, and re-runs the failing tests. A failure
+own `test(fidelity): re-baseline <spec basenames>` commit, and re-runs the failing tests.
+**No rewritten baseline is staged unread.** Each PNG a capture rewrote is read against its old
+baseline, and against its frame where a sidecar names one, before it is staged; one showing
+another screen than the test meant is restored, never committed — `--update-snapshots` writes
+whatever the run captured, a missed click included (KAN-803). A failure
 that is no screenshot mismatch — a timeout, a screen only production serves, a missing service — is
 never re-baselined: it is reported, and blocks as above.
 
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index 02ae0678..ad054459 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -21,6 +21,7 @@ changed screen alone to spare the others, and its steps carry two more. Every ba
 moves is recaptured: every spec `specs` prints for the fix's own diff — the whole suite when
 `specs` is not declared or prints nothing — runs through `capture` with `--update-snapshots=all`
 appended (step 8, `skills/flow/visual-verify-verifier.md`), and every PNG whose bytes changed is
+read first, per **The sweep** (`skills/flow-contracts/known-bugs.md`), and
 committed with the task, in the checkout step 12 of **Steps 3–13** (`skills/flow/visual-verify.md`)
 commits the stage's PNGs to. And every step-4 `visual-verify` re-run of the round, the closing full
 run included, names under `moved:` in its prompt every frame a `<spec>.mockups` sidecar maps to a
````

  - [x] **Step 3: Verify.** `grep -c -F "No rewritten baseline is staged unread" skills/flow-contracts/known-bugs.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ cfb9b564 → 1 -->
  - [x] **Step 4: Commit** the files with the subject below.

**Build:** green
**Files:** `skills/flow-contracts/known-bugs.md`, `skills/flow/verify-fix-loop.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow-contracts): read every rewritten baseline before staging it
**After:** Task 3

**Decision:** rewritten-baseline-read

- [x] 5. Make setLevel's friend check and write atomic against a concurrent unfriend (KAN-733)
  - [x] **Step 1: Create the gymie worktree.** `git -C /Users/tweety53/Projects/gymie fetch origin`, then
    `git -C /Users/tweety53/Projects/gymie worktree add /Users/tweety53/Projects/gymie-worktrees/self-review-backlog-flow-6 -b self-review-backlog-flow-6 origin/develop`.
    Every later step runs in that worktree.
  - [x] **Step 2: Reproduce before touching anything; do not guess.** Read
    `src/app/profile/src/main/kotlin/com/gymie/profile/core/services/access/impl/AccessServiceImpl.kt`:
    `setLevel` is not `@Transactional` and checks `friendships.findPair(...)`, a plain `SELECT` that
    takes no lock; `FriendshipRepositoryAdapter.unfriend` reads the pair with a plain `SELECT` too.
    <!-- measured: git show origin/develop:<those two files> @ gymie 8d2bb89d → setLevel has no @Transactional and calls findPair; unfriend's SELECT has no FOR UPDATE -->
  - [x] **Step 3: Write the two failing tests** in `src/app/profile/src/test/kotlin/com/gymie/profile/infrastructure/persistence/SocialGraphAdapterTest.kt`,
    on the lock-wait pattern `AccessGrantRepositoryAdapterTest`'s `two concurrent drops of the two scopes leave the pair deleted` uses
    (a raw JDBC connection with `autoCommit = false` holding a row lock, the other side on a `CompletableFuture`, `pg_locks WHERE NOT granted` polled until it queues, then commit):
    - `lockAcceptedPair waits for a concurrent unfriend and then answers false` — the held connection runs `unfriend`'s statements
      (`SELECT … FROM friendship WHERE id = ? FOR UPDATE`, both grant deletes, the friendship delete); `lockAcceptedPair(a, b)` in a
      transaction queues behind it and, after the commit, answers `false`.
    - `unfriend waits for a held pair lock and then removes the grant written under it` — the held connection runs
      `SELECT 1 FROM friendship WHERE user_id = ? AND friend_id = ? AND state = 'accepted' FOR UPDATE` and inserts an `access_grant`
      row for the pair; `unfriend` queues behind it and, after the commit, leaves no `access_grant` row for the pair in either direction.
    Run `./gradlew :profile:test --tests 'com.gymie.profile.infrastructure.persistence.SocialGraphAdapterTest'`: the first test does not compile
    (no `lockAcceptedPair`) — comment it out to watch the second fail on the unchanged `unfriend` (a grant row survives), then restore it.
    <!-- unverified: the Gradle project path `:profile` and the test-application wiring SocialGraphAdapterTest needs for a transactional call — confirm with ./gradlew projects and the file's own TestApplication -->
  - [x] **Step 4: Fix.** Add to `FriendshipReadRepository`
    `fun lockAcceptedPair(a: UUID, b: UUID): Boolean` — KDoc: "Whether [a] and [b] are accepted friends, holding the
    pair's row lock to the end of the caller's transaction; a concurrent [FriendshipWriteRepository.unfriend] waits for it, or it for that." —
    implemented in `FriendshipRepositoryAdapter` as
    `SELECT 1 FROM friendship WHERE user_id = ? AND friend_id = ? AND state = 'accepted' FOR UPDATE` on the ordered pair.
    Add `FOR UPDATE` to `unfriend`'s first `SELECT`. In `AccessServiceImpl`, annotate `setLevel` `@Transactional`
    (`org.springframework.transaction.annotation.Transactional`) and replace the `findPair(...)?.state != ACCEPTED` check with
    `!friendships.lockAcceptedPair(callerId, targetUserId)`; update the class KDoc's first line to say the check holds the
    friendship row's lock through the write (KAN-733). In `AccessServiceImplTest`, mock `lockAcceptedPair` where the tests mock `findPair`
    (`false` for the null and `PENDING` cases, `true` for `ACCEPTED`).
  - [x] **Step 5: Verify.** `./gradlew ktlintFormat`, then
    `./gradlew verifyChange :profile:test --tests 'com.gymie.profile.infrastructure.persistence.SocialGraphAdapterTest' --tests 'com.gymie.profile.core.services.access.AccessServiceImplTest'`
    is `BUILD SUCCESSFUL`.
  - [x] **Step 6: Commit** the five files with the subject below.

**Build:** green
**Files:** `src/app/profile/src/main/kotlin/com/gymie/profile/core/ports/persistence/profile/FriendshipReadRepository.kt`, `src/app/profile/src/main/kotlin/com/gymie/profile/infrastructure/persistence/FriendshipRepositoryAdapter.kt`, `src/app/profile/src/main/kotlin/com/gymie/profile/core/services/access/impl/AccessServiceImpl.kt`, `src/app/profile/src/test/kotlin/com/gymie/profile/core/services/access/AccessServiceImplTest.kt`, `src/app/profile/src/test/kotlin/com/gymie/profile/infrastructure/persistence/SocialGraphAdapterTest.kt`, `src/app/profile/src/test/kotlin/com/gymie/profile/core/services/profile/ProfileServiceImplTest.kt`, `src/app/profile/src/test/kotlin/com/gymie/profile/core/services/profile/VisibilityTest.kt`
**Tests:** `lockAcceptedPair waits for a concurrent unfriend and then answers false`, `unfriend waits for a held pair lock and then removes the grant written under it`
**Regression:** reverting the commit drops `lockAcceptedPair` (the first test stops compiling) and `unfriend`'s `FOR UPDATE` (the second leaves a grant row between non-friends).
**Baseline:** before=30 after=32 <!-- measured: grep -c -w '@Test' src/app/profile/src/test/kotlin/com/gymie/profile/infrastructure/persistence/SocialGraphAdapterTest.kt @ gymie 8d2bb89d → 30 -->
**Commit:** fix(profile): hold the friendship row lock through setLevel's check and write
**After:** none

**Decision:** setlevel-friendship-row-lock

Correction (2026-10-18): Step 1 declared a plain `git worktree add … -b self-review-backlog-flow-6 origin/develop`; the worktree was created by `kickoff-worktree.sh --from` (the recipe **2. Isolate the workspace** requires for a repository holding no worktree), on branch `spectre/self-review-backlog-flow-6` off `origin/develop` @ `8d2bb89d`. `**Files:**` widened by `ProfileServiceImplTest.kt` and `VisibilityTest.kt`: each holds a test fake implementing `FriendshipReadRepository`, which must override the new `lockAcceptedPair` to compile — the design's "only implementation is the adapter" held for main code only. RED was taken on an unlocked first cut of `lockAcceptedPair` (no `FOR UPDATE` on either read) rather than by commenting a test out: both tests failed, neither call queued on the held lock.
