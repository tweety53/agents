# self-review-kan-983-render-screens-from-cached-data-first-refresh

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-983's five self-review findings — verify-list templates the runner executes
literally, a regression-checkout worktree missing at the first verifier dispatch, a sibling
worktree's verifier `-diff-base` the guard cannot resolve, the intermittent `addToBuilder`
timeout, and a full spec re-run at an unchanged HEAD — and the red Go guard suite planning found
on `origin/main`.

**Architecture:** one shell guard in `run-verify-list.sh` with its harness case, matched by prose
moves in gymie's and gymie-frontend's `.flow/project.md`; a comma split in the Go
`check-visual-verify-dispatched` guard with one test; three prose additions in
`skills/flow/visual-verify.md` and `skills/flow/visual-verify-verifier.md`; a harness fix in
gymie-playwright's `fixtures.ts`, opened on the failing trace; and two here-strings in
`scripts/test-check-task-records.sh`.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-19-self-review-kan-983-render-screens-from-cached-data-first-refresh-design.md` in the agents worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Each task names the repository it runs in by its worktree; every `**Files:**` path is relative to that worktree's root. The worktrees exist already, each on branch `self-review-kan-983-render-screens-from-cached-data-first-refresh` (decision `outside-worktrees-at-plan-time`); never touch a main checkout, never create a worktree.
  - agents: `/Users/tweety53/Projects/agents-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh`, off `origin/main` — Tasks 3–7; lands by agents' `## default landing route`, merge and push (`pipeline-fix.sh land`).
  - gymie: `/Users/tweety53/Projects/gymie-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh`, off `origin/develop` @ `3d51d3aa` — Task 1; lands by gymie's route, merge and push.
  - gymie-frontend: `/Users/tweety53/Projects/gymie-frontend-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh`, off `origin/develop` @ `9806267d` — Task 2; lands by gymie-frontend's route, merge and push.
  - gymie-playwright: `/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh`, off `origin/main` @ `d416244` — Task 8; gymie-playwright declares no `## default landing route`, so it lands by the `pull request` fallback (`skills/flow-contracts/project-configuration.md`).
- Land Tasks 1 and 2 before Task 4: once Task 4's refusal is on agents' `main`, a gymie or gymie-frontend `## test` list still holding its template exits 2 instead of 1.
- Tasks 3–7's diffs are the unified diffs their fix steps give, applied with `git apply` from the agents worktree's root, checked in task order in a scratch clone at `9df14302`; Tasks 1 and 2's apply the same way in their own worktrees. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Tasks 4 and 6 prove their tests bite: each new test fails with the fix removed, as the task's Step 3 states, and passes with it restored.
- Task 8 runs against gymie's stack served from the gymie worktree above, which detects the gymie-frontend worktree beside it: from `/Users/tweety53/Projects/gymie-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh` run `docker compose up -d`, `./gradlew -q devStackFresh || { ./gradlew devStop && ./gradlew devStart -PisolateOAuthPorts; }` and `./gradlew devStatus` (gymie's `## run`). Once Task 8 is done, stop it with `./gradlew devStop` and confirm with `./gradlew devWorkspaceInventory` that nothing of this worktree still runs. Never touch a stack another worktree started.
- Every Gradle or Playwright command runs in the foreground with a Bash `timeout: 600000`; one command at a time per worktree.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 4: the placeholder check runs over the whole list before any command, matches a `<word>` token only, and exits 2 with the offending line on stderr; the header's exit-2 row names the new case.
- Task 6: the split leaves a single-sha `diffBase` judged exactly as before, and the sibling's freshness check still reads the sibling's own sha.
- Task 7: a non-zero exit of the re-run still blocks; the re-run is offered only when every worktree of the resolved set is at the HEAD the previous dispatch verified.
- Task 8: the fix acts on the race the trace shows, never on a timeout, sleep or retry.

- [x] 1. Move gymie's verify-bundle template out of the `## test` fence
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh`, run
    `/Users/tweety53/Projects/agents-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh/scripts/project-get.sh . test | awk '/^```bash[[:space:]]*$/{f=1;next}/^```/{f=0;next} f && NF && $1 !~ /^#/'`:
    it prints two lines, the second `./gradlew verifyChange :<module>:test --tests '<selectors>'   # one-invocation verify bundle`;
    `bash -c "echo verifyChange :<module>:test"` prints `bash: line 1: module: No such file or directory` and exits 1.
    <!-- measured: the bash -c line above in a scratch directory @ gymie origin/develop 3d51d3aa → exit 1 -->
  - [x] **Step 2: Apply the fix** (`git apply --unidiff-zero <the saved diff>`; the diff carries no context lines, so no fence marker of the file sits inside this block):

````diff verified:git apply --check --unidiff-zero in /Users/tweety53/Projects/gymie-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh @ 3d51d3aa exits 0; after it, the Step 1 awk prints only the full-suite line
diff --git a/.flow/project.md b/.flow/project.md
index e304cf52..392cdb3c 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -458 +457,0 @@ database section; a seeded account's weigh-ins and targets are replaced wholesal
-./gradlew verifyChange :<module>:test --tests '<selectors>'   # one-invocation verify bundle
@@ -460,0 +460,5 @@ database section; a seeded account's weigh-ins and targets are replaced wholesal
+A task's verify step runs the one-invocation verify bundle, `./gradlew verifyChange :<module>:test --tests '<selectors>'`, its
+`<selectors>` filled in and `<module>` named. It is a template, not a command, so it stays out of
+the fence: `run-verify-list.sh` runs every fence line literally, and refuses a `<word>` placeholder
+inside one.
+
````

  - [x] **Step 3: Verify.** The Step 1 awk prints one line, the full-suite one.
  - [x] **Step 4: Commit** `.flow/project.md` with the subject below.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — project configuration prose; Task 4's harness covers the runner.
**Regression:** none — no test is declared.
**Commit:** chore(flow): state the verify-bundle template outside the test fence
**After:** none

**Decision:** verify-template-in-prose
**Decision:** finding-1-one-task-per-repo

- [x] 2. Move gymie-frontend's verify-bundle template out of the `## test` fence
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/gymie-frontend-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh`, run the awk of Task 1's Step 1:
    it prints two lines, the second `./gradlew verifyChange :shared:desktopTest --tests '<selectors>'     # one-invocation verify bundle`.
  - [x] **Step 2: Apply the fix** (`git apply --unidiff-zero <the saved diff>`; the diff carries no context lines, so no fence marker of the file sits inside this block):

````diff verified:git apply --check --unidiff-zero in /Users/tweety53/Projects/gymie-frontend-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh @ 9806267d exits 0; after it, the Task 1 Step 1 awk prints only the full-suite line
diff --git a/.flow/project.md b/.flow/project.md
index 31a8fea2..238824dd 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -35 +34,0 @@ src="$(git worktree list --porcelain | awk '/^worktree /{print substr($0,10); ex
-./gradlew verifyChange :shared:desktopTest --tests '<selectors>'     # one-invocation verify bundle
@@ -37,0 +37,5 @@ src="$(git worktree list --porcelain | awk '/^worktree /{print substr($0,10); ex
+A task's verify step runs the one-invocation verify bundle, `./gradlew verifyChange :shared:desktopTest --tests '<selectors>'`, its
+`<selectors>` filled in. It is a template, not a command, so it stays out of
+the fence: `run-verify-list.sh` runs every fence line literally, and refuses a `<word>` placeholder
+inside one.
+
````

  - [x] **Step 3: Verify.** The awk prints one line, the full-suite one.
  - [x] **Step 4: Commit** `.flow/project.md` with the subject below.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — project configuration prose; Task 4's harness covers the runner.
**Regression:** none — no test is declared.
**Commit:** chore(flow): state the verify-bundle template outside the test fence
**After:** Task 1

**Decision:** verify-template-in-prose
**Decision:** finding-1-one-task-per-repo

- [x] 3. Read check-task-records' output through a here-string
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `/Users/tweety53/Projects/agents-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh`, run
    `cd stats && make -s web-dist && go test ./internal/guard -run TestNoPipeIntoEarlyExitGrep -count=1`:
    it fails naming `test-check-task-records.sh:392` and `test-check-task-records.sh:411`, `pipe into an early-exit grep; use a here-string`.
    <!-- measured: that go test in /Users/tweety53/Projects/agents-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh @ 9df14302 → FAIL, the two lines above -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply, in task order, to a scratch clone at 9df14302; scripts/run-guard-tests.sh (44 of 44 harnesses), go test ./... -race and check-references, check-installed-citations and check-markdown-integrity exit 0 on the result
diff --git a/scripts/test-check-task-records.sh b/scripts/test-check-task-records.sh
index 6e0c8324..631af73d 100755
--- a/scripts/test-check-task-records.sh
+++ b/scripts/test-check-task-records.sh
@@ -389,7 +389,7 @@ DIRS+=("$STUB")
 printf '#!/bin/sh\nexit 1\n' >"$STUB/flow"
 chmod +x "$STUB/flow"
 PATH="$STUB:$PATH" run_guard "$FIX"
-if [ "$RC" -eq 1 ] && printf '%s' "$OUT" | grep -q "no commit with subject 'fix(peer): one'"; then
+if [ "$RC" -eq 1 ] && grep -q "no commit with subject 'fix(peer): one'" <<<"$OUT"; then
   pass "case 18 unreadable state record keeps the violation"
 else
   fail "case 18 failing flow: rc=$RC out=$OUT"
@@ -408,7 +408,7 @@ DIRS+=("$STUB")
 printf '#!/bin/sh\nprintf %s\n' "'{\"worktrees\":{\"$PLAIN\":\"x\"}}\\n'" >"$STUB/flow"
 chmod +x "$STUB/flow"
 PATH="$STUB:$PATH" run_guard "$FIX"
-if [ "$RC" -eq 1 ] && printf '%s' "$OUT" | grep -q "no commit with subject 'fix(peer): one'"; then
+if [ "$RC" -eq 1 ] && grep -q "no commit with subject 'fix(peer): one'" <<<"$OUT"; then
   pass "case 19 non-repository record path is skipped"
 else
   fail "case 19 non-repo path: rc=$RC out=$OUT"
````

  - [x] **Step 3: Verify.** The Step 1 test passes, and `bash scripts/test-check-task-records.sh` exits 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `scripts/test-check-task-records.sh`
**Tests:** none — repairs existing tests: TestNoPipeIntoEarlyExitGrep
**Regression:** none — no test is declared.
**Commit:** test(scripts): read check-task-records output through a here-string
**After:** none

**Decision:** grep-pipe-here-string

- [x] 4. Refuse a verify-list line holding a placeholder
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run Task 1's Step 1 `bash -c` line: exit 1, the redirect error.
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply, in task order, to a scratch clone at 9df14302; scripts/run-guard-tests.sh (44 of 44 harnesses), go test ./... -race and check-references, check-installed-citations and check-markdown-integrity exit 0 on the result; run-verify-list.sh /Users/tweety53/Projects/gymie-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh test, unfixed gymie, exits 2 naming the template line
diff --git a/scripts/run-verify-list.sh b/scripts/run-verify-list.sh
index e2444968..2eda39b6 100755
--- a/scripts/run-verify-list.sh
+++ b/scripts/run-verify-list.sh
@@ -9,7 +9,10 @@
 #
 # THE LIST is every non-blank, non-`#` line inside a ```bash fence of
 # `project-get.sh <worktree> <list>`'s output, in the order printed; `lint`
-# appends `check-spec-reach.sh <worktree>`. Each command runs from
+# appends `check-spec-reach.sh <worktree>`. A line holding a `<word>`
+# placeholder — `./gradlew verifyChange :<module>:test --tests '<selectors>'`
+# — is a template, not a command: the shell reads `<module>` as a redirect,
+# so the list is refused before anything runs (kan-983 self-review). Each command runs from
 # <worktree> in a fresh `bash -c`, stdin detached (</dev/null) so none waits
 # on a terminal, and the whole list always runs: a failure never stops it.
 # The caller exports the KEY=value lines prepare-workspace.sh printed for
@@ -25,7 +28,8 @@
 #   0  every command exited 0
 #   1  at least one command exited non-zero
 #   2  usage (not exactly `<worktree> lint|test`, or <worktree> is not a
-#      directory), or project-get.sh exit 2 — its stderr line says why
+#      directory), a list line holding a `<word>` placeholder — its stderr
+#      line names it — or project-get.sh exit 2 — its stderr line says why
 #   3  the list is not declared (project-get.sh exit 1): the caller
 #      auto-detects, as **Inline verify** says
 set -uo pipefail
@@ -55,6 +59,13 @@ done < <(printf '%s\n' "$BODY" | awk '
   f && NF && $1 !~ /^#/
 ')
 
+for i in "${!CMDS[@]}"; do
+  if grep -Eq '<[A-Za-z][A-Za-z0-9_-]*>' <<<"${CMDS[$i]}"; then
+    echo "run-verify-list.sh: $LIST line is a template, not a command — move it out of the fence: ${CMDS[$i]}" >&2
+    exit 2
+  fi
+done
+
 OUT="$(mktemp "${TMPDIR:-/tmp}/run-verify-list.XXXXXX")"
 trap 'rm -f "$OUT"' EXIT
 
diff --git a/scripts/test-run-verify-list.sh b/scripts/test-run-verify-list.sh
index 356d2455..b19feb3d 100755
--- a/scripts/test-run-verify-list.sh
+++ b/scripts/test-run-verify-list.sh
@@ -75,6 +75,12 @@ printf '## lint\n\n```bash\ntrue\n```\n' > "$ROOT/.flow/project.md"
 run_bin "$ROOT" test
 expect_rc "an undeclared list" 3
 
+printf '## test\n\n```bash\nprintf ran-first\n./gradlew verifyChange :<module>:test --tests %s\n```\n' "'<selectors>'" > "$ROOT/.flow/project.md"
+run_bin "$ROOT" test
+expect_rc "a placeholder line refuses the list" 2
+expect_has "the refusal names the line" "test line is a template, not a command — move it out of the fence: ./gradlew verifyChange :<module>:test --tests '<selectors>'"
+expect_lacks "nothing runs before the refusal" 'ran-first'
+
 run_bin "$ROOT" build
 expect_rc "an unknown list" 2
 run_bin "$ROOT/missing" lint
diff --git a/skills/flow-contracts/project-configuration.md b/skills/flow-contracts/project-configuration.md
index 06e150e0..1dd3ab54 100644
--- a/skills/flow-contracts/project-configuration.md
+++ b/skills/flow-contracts/project-configuration.md
@@ -24,7 +24,7 @@ exactly what that row specifies and nothing else. The row is authoritative.
 | `## run` | How to start the stack and individual services locally — the actual commands, including any parameter that must point at another app's root. |
 | `## stop` | Optional. The command that stops the project's local stack, run by bare `/flow` before its stack-stopped check so nothing holds a port or file handle open in a worktree about to be removed. Same shape as `## worktree setup`: one command per line inside a fence, read literally and run in order from the worktree; prose outside a fence is never run. **Absent means that check is skipped, not failed** — cleanup proceeds on the strength of the other two checks. A body with no fenced command declares no command and is skipped the same way. A project with no stack should say so here rather than omit the key, so its absence is a recorded fact rather than an oversight. |
 | `## credentials` | Local-development-only sign-in credentials (which app, user, password) and what seeds them. Never deployed-environment secrets. |
-| `## test` | The command(s) that run the project's tests. |
+| `## test` | The command(s) that run the project's tests. Every non-`#` line inside its bash fence runs literally, as `## lint`'s do: a template with a `<word>` placeholder, such as a verify bundle's `:<module>:test` selector, belongs in prose outside the fence, and `<agents repo>/scripts/run-verify-list.sh` refuses the list while one stands inside it. |
 | `## worktree setup` | Optional. A fenced command block run once per worktree, from the worktree root, immediately after the run creates the worktree — **2. Isolate the workspace** (`skills/flow/implement.md`) on `/flow`, the kickoff section (`skills/flow-fast/SKILL.md`) on a `/flow-fast` creating run — and before anything else touches the tree — the place for a build that a gitignored, embedded artifact needs before the project's first `go test`/`go build` can succeed. Absent means nothing runs. Same shape as `## lint` and `## test`: one command per line inside the fence, read literally and run in order; resolved through `project-get.sh <worktree> "worktree setup"`, whose exit 1 is the absent case. |
 | `## lint` | The command(s) that verify lint, and the auto-fix command to run first. |
 | `## standards` | The project's own written standards: the files the principles reviewer receives, plus any opt-in shared rule the project has adopted. This same list is both the opt-in list and the reviewer's standards list. Absent, no standards resolve: the reviewer gets an empty list, never auto-detected files. |
````

  - [x] **Step 3: Prove the cases bite.** `git stash push scripts/run-verify-list.sh`, run `bash scripts/test-run-verify-list.sh`:
    `a placeholder line refuses the list`, `the refusal names the line` and `nothing runs before the refusal` FAIL; `git stash pop`, run it again: `all cases passed`.
  - [x] **Step 4: Verify.** `bash scripts/test-run-verify-list.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py`, `scripts/check-verbatim-moves.sh` and
    `cd stats && make -s web-dist && go test ./internal/guard -run TestNoPipeIntoEarlyExitGrep -count=1` exit 0.
  - [x] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `scripts/run-verify-list.sh`, `scripts/test-run-verify-list.sh`, `skills/flow-contracts/project-configuration.md`
**Tests:** `a placeholder line refuses the list`, `the refusal names the line`, `nothing runs before the refusal`
**Regression:** reverting the commit runs the template line: the list exits 1, the error names no line, and `printf ran-first` runs first.
**Baseline:** before=20 after=23 <!-- measured: grep -c '^expect_' scripts/test-run-verify-list.sh @ scratch clone of 9df14302, before and after the diff → 20, 23 -->
**Commit:** fix(scripts): refuse a verify-list line holding a placeholder
**After:** none

**Decision:** verify-list-placeholder-refusal

- [x] 5. Create a missing regression-checkout worktree before the visual pre-flight
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -F "gets one first" skills/flow/visual-verify.md; echo $?`: it prints only `1`.
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply, in task order, to a scratch clone at 9df14302; scripts/run-guard-tests.sh (44 of 44 harnesses), go test ./... -race and check-references, check-installed-citations and check-markdown-integrity exit 0 on the result
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index f6549ba0..dd425515 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -180,7 +180,13 @@ verifier reports on a `wire contract` line blocks as a defect in a captured scre
 (**Blocking** below).
 
 3. **Pre-flight the workspace, before anything is dispatched.** The verifier's one re-dispatch
-   cannot repair an environment that cannot pass. With the parent's own Bash call, before the
+   cannot repair an environment that cannot pass. **A `regression checkout` this run's resolved
+   worktree set holds no worktree for gets one first.** When `## apps` declares the repository
+   `## visual verification`'s `regression checkout` names and the set holds no worktree in it, the
+   parent runs that entry's recipe — the `kickoff-worktree.sh --from` call, then the regression
+   checkout's `setup` (`skills/flow/cross-repo-worktrees.md`) — before the checks below, and adds
+   the worktree to the set: kan-983's first verifier was dispatched without it and blocked on the
+   missing checkout, ~4.45M tokens spent for nothing. With the parent's own Bash call, before the
    dispatch below, run the four checks — ports held only by their own app, the base-URL
    configuration resolving to the workspace's values, origin allowed, one Playwright checkout per
    workspace — against this worktree:
````

  - [x] **Step 3: Verify.** `grep -c -F "gets one first" skills/flow/visual-verify.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): create a missing regression-checkout worktree before the visual pre-flight
**After:** none

**Decision:** regression-worktree-before-preflight

- [x] 6. Judge each sha of a comma-joined verifier diff base
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -F 'strings.Split(head, ",")' stats/internal/guard/visualverifydispatched.go; echo $?`: it prints only `1` — a `diffBase` holding a comma fails `vvdSHA` and verifies nothing.
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply, in task order, to a scratch clone at 9df14302; scripts/run-guard-tests.sh (44 of 44 harnesses), go test ./... -race and check-references, check-installed-citations and check-markdown-integrity exit 0 on the result
diff --git a/scripts/check-visual-verify-dispatched.sh b/scripts/check-visual-verify-dispatched.sh
index 40a0b4d9..0bb43e42 100755
--- a/scripts/check-visual-verify-dispatched.sh
+++ b/scripts/check-visual-verify-dispatched.sh
@@ -98,7 +98,10 @@
 # (skills/flow/visual-verify.md, **Recording**) — is a commit sha this
 # worktree resolves, and no path of this change's own `<merge-base>..HEAD`
 # diff that changed between that sha and HEAD matches the `ui paths`, asked
-# as the trigger above is, the canonical worktree's included. A row with no
+# as the trigger above is, the canonical worktree's included. A `diffBase`
+# naming several shas, comma-joined — a canonical verifier that also covered
+# a sibling worktree's HEAD (kan-983 self-review) — is read one sha at a time,
+# each judged as a row of its own. A row with no
 # `diffBase`, a ref name such as HEAD in place of a sha, or a sha this
 # worktree cannot resolve verified nothing here. With qualifying rows and
 # none fresh, the verdict is
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index dd425515..479da503 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -73,6 +73,7 @@ more than one worktree — the worktree's main checkout's basename, the name
 `check-visual-trigger.sh --repo` takes, since every worktree's own basename is the change name — the pair's semantics are section 4 of `skills/flow/implement.md`, cited
 here, not restated.
 Its `begin` carries `-diff-base` with the sha `git -C <worktree> rev-parse HEAD` prints before the dispatch, the HEAD it verifies: `check-visual-verify-dispatched.sh` counts a verifier row only for that HEAD, and only while no UI path the change touches has changed since.
+A verifier whose worktree's `ui paths` also match a sibling worktree's diff — a `<repo>/**` glob naming the sibling's repository — verifies that sibling's HEAD as well, so its `-diff-base` names each such HEAD too, comma-joined after its own: the guard checks each worktree against the one sha it resolves, and no row is recorded for a sibling no verifier was dispatched to (kan-983's `visual-verify-2-gymie-frontend`).
 
 **Every report is checked for completeness before anything else reads it:**
 
diff --git a/stats/internal/guard/check_visual_verify_dispatched_test.go b/stats/internal/guard/check_visual_verify_dispatched_test.go
index 3867f8ee..5e62fddb 100644
--- a/stats/internal/guard/check_visual_verify_dispatched_test.go
+++ b/stats/internal/guard/check_visual_verify_dispatched_test.go
@@ -642,3 +642,14 @@ func TestCheckVisualVerifyDispatchedMissingNamesPathsSinceVerifiedHead(t *testin
 		t.Fatalf("got exit %d stdout %q, want exit 1 ending %q", r.code, r.out, want)
 	}
 }
+
+// TestCheckVisualVerifyDispatchedCommaJoinedDiffBase pins the kan-983
+// self-review fix: a canonical verifier that also covered this worktree
+// records both HEADs comma-joined, and the sha this worktree resolves counts.
+func TestCheckVisualVerifyDispatchedCommaJoinedDiffBase(t *testing.T) {
+	t.Parallel()
+	r := vvdRun(t, vvdCase{dispatches: []byte(`[{"key":"visual-verify-gymie","role":"verifier","outcome":"completed","diffBase":"0123456789abcdef0123456789abcdef01234567,@HEAD@"}]`)})
+	if r.code != 0 || !strings.HasPrefix(r.out, "VISUAL-VERIFY-OK: UI paths touched") {
+		t.Fatalf("got exit %d stdout %q stderr %q, want exit 0 and the OK verdict", r.code, r.out, r.err)
+	}
+}
diff --git a/stats/internal/guard/visualverifydispatched.go b/stats/internal/guard/visualverifydispatched.go
index e3cd41d6..60021513 100644
--- a/stats/internal/guard/visualverifydispatched.go
+++ b/stats/internal/guard/visualverifydispatched.go
@@ -188,7 +188,14 @@ func checkVisualVerifyDispatched(args []string, env Env, stdout, stderr io.Write
 	// is scoped to.
 	var lastHead string
 	var lastChanged []string
+	// A canonical verifier that also covered a sibling worktree names each
+	// HEAD it verified, comma-joined (kan-983 self-review): each sha is judged
+	// on its own, and only the one this worktree resolves can count here.
+	var shas []string
 	for _, head := range heads {
+		shas = append(shas, strings.Split(head, ",")...)
+	}
+	for _, head := range shas {
 		if !vvdSHA.MatchString(head) {
 			continue
 		}
````

  - [x] **Step 3: Prove the test bites.** `git stash push stats/internal/guard/visualverifydispatched.go`, run
    `cd stats && go test ./internal/guard -run TestCheckVisualVerifyDispatchedCommaJoinedDiffBase -count=1`: it FAILs; `git stash pop`, run it again: `ok`.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./... && go test ./internal/guard -run TestCheckVisualVerifyDispatched -count=1` prints no file and passes;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 5: Commit** the four files with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify.md`, `scripts/check-visual-verify-dispatched.sh`, `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `stats/cmd/flow/record.go`, `stats/internal/records/types.go`
**Tests:** `TestCheckVisualVerifyDispatchedCommaJoinedDiffBase`
**Regression:** reverting the commit reads `<sha>,<sha>` as one non-sha: the test's row verifies nothing and the guard answers MISSING, exit 1.
**Baseline:** before=4 after=5 <!-- measured: grep -c '^func Test' stats/internal/guard/check_visual_verify_dispatched_test.go @ scratch clone of 9df14302, before and after the diff → 4, 5 -->
**Commit:** fix(guard): judge each sha of a comma-joined verifier diff base
**After:** Task 5

**Decision:** comma-joined-diff-base

- [x] 7. Re-run only the failed specs at an unchanged HEAD
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -F "rerun specs:" skills/flow/visual-verify.md skills/flow/visual-verify-verifier.md; echo $?`: it prints only `1`.
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply, in task order, to a scratch clone at 9df14302; scripts/run-guard-tests.sh (44 of 44 harnesses), go test ./... -race and check-references, check-installed-citations and check-markdown-integrity exit 0 on the result
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 5bac5327..82f3feb4 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -21,6 +21,9 @@ never an exit code the command did not return.
    spec from `specs` run against the merge base.
    A `specs` with no `<merge-base>` cannot take a round base: it runs as declared, its full list,
    and the report names `specs: not scoped — specs has no <merge-base>`.
+   **A prompt naming `rerun specs:` runs those specs alone** — substituted for `<specs>` in place
+   of `specs`' output, which is not run — and the report's `verify:` line reads `exit <n> — rerun
+   of <k> failed specs at unchanged HEAD` (**Steps 3–13**, `skills/flow/visual-verify.md`).
    A non-zero `verify` exit blocks, a failing test this change did not introduce included
    (**the sweep**, `skills/flow-contracts/known-bugs.md`), save a stale snapshot baseline in a spec
    this change never touched, which the sweep re-baselines and re-runs before the exit is read.
@@ -554,7 +557,7 @@ or a measurement the pixels could not give is step 10 undone.
 - setup: <not declared | exit <n>>
 - stack: <already running | started and stopped | could not be started — <output>>
 - fingerprint: <not declared | exit 0 | mismatch → restarted → exit <n>>
-- verify: exit <n> | not run — no spec since the round base
+- verify: exit <n> | exit <n> — rerun of <k> failed specs at unchanged HEAD | not run — no spec since the round base
   <output, verbatim or last 40 lines>
 - capture: exit <n>
   <output, verbatim or last 40 lines>
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 479da503..7b0ba5f6 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -110,7 +110,13 @@ before the next re-dispatch, and ends your turn with `## Question` naming the fa
 its output, verbatim, only when the cure needs a decision only the operator can make. Never run
 the failing command yourself to check it. A re-dispatch whose report repeats the previous one's
 failure on the same evidence takes **Fewest operator actions**
-(`skills/flow-contracts/pipeline.md`). The ledger render and this stage's `end` mark follow
+(`skills/flow-contracts/pipeline.md`). **A re-dispatch at an unchanged HEAD re-runs only the
+specs that failed.** When every worktree of the resolved set is at the HEAD the previous dispatch
+verified and that report's `verify:` exit was non-zero, the prompt adds `rerun specs:` and the
+spec paths that report's `verify` output names as failing; the verifier's step 7 runs those alone
+(`skills/flow/visual-verify-verifier.md`), and their exit is the `verify` exit — a non-zero one
+still blocks, exactly as above: kan-983's `visual-verify-2` re-ran all 87 specs, ~14 minutes, to
+clear one the previous run had already passed 30 times out of 30. The ledger render and this stage's `end` mark follow
 whichever report was last.
 
 **A final report that is a block closes its dispatch `-outcome blocked -cause <cause>`, never
````

  - [x] **Step 3: Verify.** `grep -c -F "rerun specs:" skills/flow/visual-verify.md` and the same over `skills/flow/visual-verify-verifier.md` each print `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify.md`, `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): re-run only the failed specs at an unchanged HEAD
**After:** Task 5, 6

**Decision:** rerun-failed-specs-at-unchanged-head

- [ ] 8. Fix `addToBuilder`'s intermittent workout-field timeout at its source
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** With the stack of Global constraints up, export the
    `KEY=value` lines `/Users/tweety53/Projects/agents-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh/scripts/prepare-workspace.sh /Users/tweety53/Projects/gymie-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh` prints, then from `/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh` run
    `FRONTEND_URL=http://localhost:${FRONTEND_PORT:-3000} BACKEND_URL=http://localhost:${GATEWAY_PORT:-8080} npx playwright test tests-full/group-session-journey.spec.ts --repeat-each=20 --workers=4 --trace=retain-on-failure`.
    <!-- predicted: at least one repeat fails with `waitForTag: no live tag 'add_participant_workout_field'`, as kan-983's verifier saw -->
    No failure in that run: run it once more at `--workers=8`; still none — report the two runs, commit nothing, and stop here.
    unverified:the env names FRONTEND_PORT and GATEWAY_PORT come from gymie's `## visual verification` `verify` row — confirm prepare-workspace.sh prints them.
  - [ ] **Step 2: Read the first failing trace.** Unzip its `trace.zip` under `test-results/` into a scratch directory outside the worktree and read
    the actions and screenshots around the `add_participant_row_<id>` tap: where the tap landed, the row's box in the frames before and after it,
    and whether the sheet left the person step. Record what it shows in the report.
  - [ ] **Step 3: Fix the cause the trace shows, in `fixtures.ts`.**
    - The tap landed while the sheet was still moving, or on another row — the leading hypothesis, the race `filterTo` already guards (KAN-761):
      in `addToBuilder`, before tapping the row, wait for the candidate list to finish replacing its skeleton —
      `await waitForTagGone(page, 'add_participant_skeleton')` and then `await waitForTag(page, 'add_participant_count')` — and extend the doc comment with the reason, as `filterTo`'s states it.
      unverified:that the skeleton and count tags bracket the reflow — confirm from the trace frames.
    - The tap reached the row and the sheet left the person step, yet the workout field never drew: the cause is outside the harness — report it with the trace's evidence and commit nothing.
    - Never a longer timeout, a sleep or a retry (`rules/fix-determinism-at-the-source.mdc`).
  - [ ] **Step 4: Prove the fix.** Re-run Step 1's command: every repeat passes. Revert the Step 3 edit, run it again: the failure returns; restore the edit.
  - [ ] **Step 5: Verify.** `npx tsc --noEmit -p .` exits 0 in `/Users/tweety53/Projects/gymie-playwright-worktrees/self-review-kan-983-render-screens-from-cached-data-first-refresh`.
    unverified:that gymie-playwright's tsconfig type-checks fixtures.ts — confirm the command; with none, `npx playwright test --list tests-full/group-session-journey.spec.ts` exits 0.
  - [ ] **Step 6: Commit** `fixtures.ts` with the subject below; a different cause restates the subject in a Correction line under this task.

**Build:** green
**Files:** `fixtures.ts`
**Tests:** **none** — a harness fix, proved by Step 4's repeated runs.
**Regression:** none — no test is declared.
**Commit:** fix(fixtures): tap an add-participant row only once its list has settled
**After:** none

**Decision:** addtobuilder-trace-first
