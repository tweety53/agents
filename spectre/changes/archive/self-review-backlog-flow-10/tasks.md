# self-review-backlog-flow-10

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the next five live Backlog Flow findings — KAN-899's silent settings outage,
KAN-900's unverified clean integrate rebase, KAN-902's false clean from an absolute-path guard call
in a worktree, KAN-905's guard tests failing in a shallow clone, and KAN-907's timed ignored-signal
looks.

**Architecture:** one sentence and one redirect removed in `skills/flow/SKILL.md`; a lint run after
every clean sync rebase in `scripts/integrate-change.sh`, with its harness case and the three skill
files that state integrate's verification; a cwd-worktree re-exec in `scripts/lib/flow-guard.sh`
called by nine guard shims, with its harness cases; and two test-only changes in
`stats/internal/guard/`.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-19-self-review-backlog-flow-10-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Every task runs in this worktree, `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-10`, branch `self-review-backlog-flow-10`; every `**Files:**` path is relative to its root. Never touch a main checkout.
- Every task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root, checked in order in a scratch clone at `46db35e0`. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Tasks 1 and 2's sentences land character for character: `scripts/check-verbatim-moves.sh` matches them against `verbatim-moves.txt` exactly.
- Every test-bearing task proves its tests bite: each new or kept assertion fails with the mutation its RED step names, and passes with it restored.
- `<scratch>` is a scratch directory outside the worktree; every worktree or clone a step makes there is removed by the same step.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 2: a project with no `## lint` list (repository `b` of every other harness case) still passes prepare after a clean rebase; land's re-sync runs the same check.
- Task 3: the re-exec fires only for another worktree of the same repository that carries the shim; a call from the shim's own checkout or from no repository runs the shim's own guard.
- Task 3: `check-plan-provenance.sh` makes the call before it exports `CHECK_PLAN_PROVENANCE_ROOT`, so the worktree's copy derives its own root.
- Task 4: in a full clone no parity test skips; in a `--depth 1` clone the package passes with each history-bound test skipped by name.
- Task 5: with the ignored-signal filter of `trapExitSignals` removed, both ignored-SIGHUP subtests fail on every run.

- [x] 1. Leave the settings read's stderr visible (KAN-899)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -c -F 'flow settings get 2>/dev/null' skills/flow/SKILL.md`: it prints `1`.
    <!-- measured: the grep above in this worktree @ 46db35e0 → 1 -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 46db35e0; check-verbatim-moves.sh 46db35e0 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow/SKILL.md b/skills/flow/SKILL.md
index fc79ec03..7c5c8071 100644
--- a/skills/flow/SKILL.md
+++ b/skills/flow/SKILL.md
@@ -31,13 +31,15 @@ it back for every later phase, so it is never passed again.
 **Resolve this once, near the top of every run, before any dispatch below reads it:**
 
 ```bash
-SETTINGS_JSON="$(flow settings get 2>/dev/null || true)"
+SETTINGS_JSON="$(flow settings get || true)"
 REVIEWERS="$(printf '%s' "$SETTINGS_JSON" | jq -r '.reviewers[]')"
 VERIFY_MODEL=opus
 ```
 
 A non-zero exit from `flow settings get` means the settings store could not be reached — there is
-no per-change fallback file for this record. Settings unreachable is never a reason to block
+no per-change fallback file for this record. The read leaves the CLI's stderr visible: report it
+verbatim as the cause, and resolve `REVIEWERS` by the table's Unreachable row below, never as the
+empty list the failed read leaves (KAN-899). Settings unreachable is never a reason to block
 implementation.
 
 **Execution mode, every dispatch's model and effort, and the review panel are decided per change,
````

  - [x] **Step 3: Verify.** `grep -c -F 'flow settings get 2>/dev/null' skills/flow/SKILL.md` prints `0` and `grep -c -F 'KAN-899' skills/flow/SKILL.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: both greps after git apply in a scratch clone @ 46db35e0 → 0 and 1 -->
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/SKILL.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): leave the settings read's stderr visible
**After:** none

**Decision:** settings-read-stderr-visible

- [x] 2. Run the lint list after every clean integrate rebase (KAN-900)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -c -F -e 'rebase_lint' -e 'rebase-lint' scripts/integrate-change.sh skills/flow/sync-onto-base.md skills/flow/integrate.md`: each file prints `0` — no step runs any check after a clean sync rebase but the `GUARD-TEST:` harnesses.
    <!-- measured: the grep above in this worktree @ 46db35e0 → 0, 0, 0 -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 46db35e0 after Task 1's diff; bash scripts/test-integrate-change.sh → 178 passed, 0 failed; the rebase_lint case fails with the mutation Step 3 names; check-verbatim-moves.sh 46db35e0 fails only on sentences verbatim-moves.txt lists
diff --git a/scripts/integrate-change.sh b/scripts/integrate-change.sh
index c5e9a1a3..c1a441f9 100755
--- a/scripts/integrate-change.sh
+++ b/scripts/integrate-change.sh
@@ -32,7 +32,8 @@
 #          gate (check-unfinished-work, check-visual-verify-dispatched against
 #          `git merge-base HEAD origin/$BASE`, check-untasked-commits from that
 #          base or an archived re-run's), check-base-moved and
-#          sync-onto-base on MOVED (running each GUARD-TEST it names), and the
+#          sync-onto-base on MOVED (running each GUARD-TEST it names, then
+#          the rebased worktree's ## lint list — rebase_lint below), and the
 #          default landing route (project-get.sh). Prints per worktree its
 #          reshape base and the diffstat the commit subjects are written from,
 #          and records the reshape base in the worktree's own git dir
@@ -120,7 +121,7 @@ missing=""
 for g in check-base-moved check-cleanup-complete check-finish-preflight check-foreign-staged \
   check-main-checkout-drift check-unfinished-work check-untasked-commits check-visual-verify-dispatched commit-archive \
   commit-split land-self-review-report migrate-worktrees project-get refresh-main-checkout \
-  remove-change-worktrees reshape-branch resolve-base-branch sync-onto-base; do
+  remove-change-worktrees reshape-branch resolve-base-branch run-verify-list sync-onto-base; do
   [ -x "$SCRIPT_DIR/$g.sh" ] || missing="$missing $g.sh"
 done
 [ -z "$missing" ] || die "missing beside this script:$missing — run the procedure by hand"
@@ -388,6 +389,22 @@ guard_tests() {
   done < <(printf '%s\n' "$OUT" | grep '^GUARD-TEST: ')
 }
 
+# rebase_lint <worktree>: a textually clean rebase can still break the build
+# — one dropped a Go import no exit code caught (KAN-900) — so the rebased
+# tree runs its project's ## lint list (run-verify-list.sh). A failure stops
+# with the list's report above the stop line; a project declaring no list
+# (exit 3) has nothing to run.
+rebase_lint() {
+  local out rc
+  out="$("$SCRIPT_DIR/run-verify-list.sh" "$1" lint 2>&1)"
+  rc=$?
+  case "$rc" in
+    0) say "REBASE-LINT-PASSED: $1" ;;
+    3) say "REBASE-LINT-NONE: $1 — its project declares no ## lint list" ;;
+    *) printf '%s\n' "$out"; stop rebase-lint "the ## lint list failed on $1 after its rebase (run-verify-list exit $rc) — report its output above; the change stays IN_PROGRESS" ;;
+  esac
+}
+
 cmd_prepare() {
   local i m v ask=0 run1=0 rb vline sha t
   load_set
@@ -522,6 +539,7 @@ cmd_prepare() {
     vline="$(verdict REBASED)"; sha="${vline##*onto }"
     RB[i]="$sha"
     guard_tests
+    rebase_lint "${WT[i]}"
     # The rebase made a HEAD no verifier saw: the gate again.
     visual_ok "${WT[i]}" "${BASES[i]}" || stop visual-verify "rebased onto origin/${BASES[i]} after the gate: $(visual_stop_text "${WT[i]}" prepare)"
   done
@@ -776,6 +794,7 @@ cmd_land() {
             *) stop sync "sync-onto-base exit $RC — report its output" ;;
           esac
           guard_tests
+          rebase_lint "$w"
           reread_fix_shas "$o"
           visual_ok "$w" "$BASE" || stop visual-verify "re-synced onto origin/$BASE after its push was rejected, so it was not pushed there; already pushed: ${pushed:-none}. $(visual_stop_text "$w" land)"
           push_branch "$w"
diff --git a/scripts/test-integrate-change.sh b/scripts/test-integrate-change.sh
index 894f92a2..5ab16e12 100755
--- a/scripts/test-integrate-change.sh
+++ b/scripts/test-integrate-change.sh
@@ -42,6 +42,9 @@
 #                   its line naming the --resume call whole.
 #   guard_test      a failing scoped re-verification stops prepare at
 #                   STOP: guard-test with the test's own output relayed.
+#   rebase_lint     a clean rebase onto a base whose ## lint list fails
+#                   stops prepare at STOP: rebase-lint with the list's own
+#                   output relayed (KAN-900).
 #   gh_open_merged  gh reports the recorded PR OPEN but the branch is on
 #                   origin/main: run 2 still runs to FINISHED.
 #   disclose        both apply worktrees hold an unpreserved ignored file:
@@ -388,6 +391,19 @@ check "guard_test: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
 check "guard_test: the stop is STOP: guard-test" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: guard-test"'
 check "guard_test: the failing test's own output is relayed" 'printf "%s" "$OUT" | grep -qx "gt-test-own-output"'
 
+# --- rebase_lint ---
+# b's base gains a .flow/project.md whose ## lint list fails and nothing this
+# change touches: the rebase is clean, and the rebased tree's lint list runs.
+fixture
+mkdir -p "$WORK/b/.flow"
+printf '# project\n\n## lint\n\n```bash\necho lint-own-output; exit 1\n```\n' >"$WORK/b/.flow/project.md"
+g -C "$WORK/b" add .flow/project.md && g -C "$WORK/b" commit -qm "base: lint list" && g -C "$WORK/b" push -q origin main
+sut prepare "$WORK/a" demo
+check "rebase_lint: prepare stops with exit 1" '[ "$RC" -eq 1 ]'
+check "rebase_lint: the rebase was clean" 'printf "%s" "$OUT" | grep -q "^REBASED: $WORK/b-worktrees/demo"'
+check "rebase_lint: the stop is STOP: rebase-lint" 'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: rebase-lint — the ## lint list failed on $WORK/b-worktrees/demo"'
+check "rebase_lint: the list's own output is relayed" 'printf "%s" "$OUT" | grep -q "lint-own-output"'
+
 # --- gh_open_merged ---
 fixture
 printf '#!/usr/bin/env bash\n[ "$1 $2" = "pr view" ] && { echo OPEN; exit 0; }\nexit 1\n' >"$WORK/bin/gh"
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index 599272f7..1ef48cff 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -389,10 +389,10 @@ remote, so there is nothing to push to or merge into"* — rather than reporting
 failure, which sends the operator debugging the wrong thing. Offer to leave the change at
 `IN_PROGRESS` with the work staged; there is nothing to lose, because nothing was pushed.
 
-**No verification gate runs before integration.** No tests, no linters, no spec-coverage check. Two exceptions exist, both under **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`): a rebase
-that needed conflict resolution runs the whole `## lint` and `## test` lists, and a clean rebase
+**No verification gate runs before integration.** No tests, no linters, no spec-coverage check. Three exceptions exist, all under **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`): a rebase
+that needed conflict resolution runs the whole `## lint` and `## test` lists, a clean rebase
 that changed a file this change also touches runs the scoped re-verification that file is
-canonical for.
+canonical for, and every clean rebase runs the whole `## lint` list.
 
 ### Resolving a change's worktrees
 
diff --git a/skills/flow/integrate.md b/skills/flow/integrate.md
index 37acbe2b..366026b2 100644
--- a/skills/flow/integrate.md
+++ b/skills/flow/integrate.md
@@ -91,7 +91,8 @@ re-verification of **Sync the branch onto the base** (`skills/flow/sync-onto-bas
 `REFUSE`, an exit 2 or an empty resolved set is `STOP: base-moved`, handled as `STOP: preflight`
 above. **Load `skills/flow/sync-onto-base.md` only when** the script stops on `sync-conflict` or
 `sync`, the rebase is resolved in place there, and `prepare` is re-run. `STOP: guard-test` is a
-scoped re-verification that failed: report it and stop at `IN_PROGRESS`.
+scoped re-verification that failed: report it and stop at `IN_PROGRESS`. `STOP: rebase-lint` is
+the project's `## lint` list failing on the rebased tree, handled the same way.
 
 The `ROUTE:` line carries the project's `## default landing route`. **A resolved default skips
 the question entirely** — take that route without asking, and say so in the handoff (`Route:
@@ -193,7 +194,7 @@ stays `IN_PROGRESS`, and the `LOOP-LANDING-REFUSED` lines are relayed — finish
 then re-run `land`. A branch with no loop rows answers `LOOP-LANDING-NA` and passes.
 
 **Every git step here can fail, and none of them may fail silently.** `STOP: push`, `STOP: sync`,
-`STOP: sync-conflict`, `STOP: guard-test`, `STOP: self-review-report` and `STOP: state` are
+`STOP: sync-conflict`, `STOP: guard-test`, `STOP: rebase-lint`, `STOP: self-review-report` and `STOP: state` are
 reported with the command's own output, leaving the change at `IN_PROGRESS`; `STOP: base` and
 `STOP: no-remote` are handled as in **Deciding which run this is** above; a conflict on the merge-and-push route's one re-sync is resolved per
 **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) and `land` re-run as the one
diff --git a/skills/flow/sync-onto-base.md b/skills/flow/sync-onto-base.md
index 90272447..ec2bf14b 100644
--- a/skills/flow/sync-onto-base.md
+++ b/skills/flow/sync-onto-base.md
@@ -27,7 +27,10 @@ and `check-base-moved.sh` performs no fetch of its own.
   never written to the state file, which every later reader of this worktree's recorded merge
   base in this run means instead — most concretely the reshape in **Run 1 — the branch is not merged**
   (`skills/flow-contracts/finish-contract-run1.md`). Run **Scoped re-verification** below, then
-  proceed to the landing question.
+  proceed to the landing question. Between the two, the rebased tree runs the project's whole
+  `## lint` list (`run-verify-list.sh <worktree> lint`): a textually clean rebase can still break
+  the build, as one that dropped a Go import did (KAN-900), so a failure is `STOP: rebase-lint`,
+  leaving the worktree rebased.
 - **Conflict** (exit 1, `CONFLICT: <worktree> — onto <sha>; unmerged: <paths>`): **resolve it in place, automatically.** For every path
   `git status` lists as unmerged, read both sides and write the file that keeps the upstream
   change *and* this change's intent, with no conflict marker left; `git add` it; then
@@ -78,5 +81,5 @@ A `NO-GUARD-TEST: <path>` is
 stated in the handoff as having no verification to run, not silently skipped. A non-zero exit from
 any discovered test blocks this stage exactly like any other verify-stage failure: report it, leave
 the change `IN_PROGRESS`, stop before the landing question, and close the mark `stopped`. **Never**
-re-run the project's whole `## lint`/`## test` list here. A clean rebase whose overlap set clears
+re-run the project's whole `## test` list here. A clean rebase whose overlap set clears
 this stage proceeds to the landing question and closes the mark `completed`.
````

  - [x] **Step 3: RED, then GREEN.** In `scripts/integrate-change.sh` delete the line `    rebase_lint "${WT[i]}"`:
    `bash scripts/test-integrate-change.sh` fails on `rebase_lint: prepare stops with exit 1`, `rebase_lint: the stop is STOP: rebase-lint` and `rebase_lint: the list's own output is relayed`. Restore the line: it passes.
    <!-- measured: the deletion above in a scratch clone @ 46db35e0 after the diff → 175 passed, 3 failed, those three; restored → 178 passed, 0 failed -->
  - [x] **Step 4: Verify.** `bash scripts/test-integrate-change.sh` ends `178 passed, 0 failed`;
    `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 5: Commit** the five files with the subject below.

**Build:** green
**Files:** `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow/sync-onto-base.md`, `skills/flow/integrate.md`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** `rebase_lint: prepare stops with exit 1`, `rebase_lint: the rebase was clean`, `rebase_lint: the stop is STOP: rebase-lint`, `rebase_lint: the list's own output is relayed`
**Regression:** reverting the commit removes the lint run after a clean rebase, so a rebased tree whose `## lint` list fails reaches the landing question again; the four checks go with it.
**Baseline:** before=129 after=133 <!-- measured: grep -c -E '^[[:space:]]*check [^(]' scripts/test-integrate-change.sh @ 46db35e0 → 129, and 133 after the diff in a scratch clone -->
**Commit:** fix(scripts): run the lint list after every clean integrate rebase
**After:** none

**Decision:** rebase-lint-after-clean-sync

- [x] 3. Run a worktree's own guard on an absolute-path call from it (KAN-902)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `git worktree add -q --detach <scratch>/wt HEAD && printf '\nA brand new rule sentence that nobody acknowledged.\n' >> <scratch>/wt/skills/flow/integrate.md && (cd <scratch>/wt && bash /Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-10/scripts/check-verbatim-moves.sh HEAD | tail -1); git worktree remove --force <scratch>/wt`:
    it prints `check-verbatim-moves: 0 violation(s), 0 to review, against HEAD` — the call judged this worktree, not `<scratch>/wt`, whose relative call reports 1.
    <!-- measured: the commands above, under the session scratchpad, in a scratch clone @ 46db35e0 → 0 violation(s); after the diff → 1 violation(s) -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 46db35e0 after Tasks 1-2's diffs; bash scripts/test-lib-flow-guard.sh passes; cases 15a and 16 fail with the mutations Step 3 names; Step 1's call reports 1 violation and check-references.sh's absolute call from a worktree holding a stale reference exits 1
diff --git a/scripts/check-ask-silence.sh b/scripts/check-ask-silence.sh
index 49b2e594..0db0343a 100755
--- a/scripts/check-ask-silence.sh
+++ b/scripts/check-ask-silence.sh
@@ -41,6 +41,7 @@ SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
   echo "check-ask-silence: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
 FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
 export FLOW_GUARD_REPO_ROOT
 flow_guard_exec check-ask-silence 2 "check-ask-silence:" "$@"
diff --git a/scripts/check-dispatch-paragraphs.sh b/scripts/check-dispatch-paragraphs.sh
index 8657204a..20518023 100755
--- a/scripts/check-dispatch-paragraphs.sh
+++ b/scripts/check-dispatch-paragraphs.sh
@@ -345,6 +345,7 @@ SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
   echo "check-dispatch-paragraphs: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
 FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
 export FLOW_GUARD_REPO_ROOT
 flow_guard_exec check-dispatch-paragraphs 2 "check-dispatch-paragraphs:" "$@"
diff --git a/scripts/check-guard-symlinks.sh b/scripts/check-guard-symlinks.sh
index d872f8a5..3a9542e6 100755
--- a/scripts/check-guard-symlinks.sh
+++ b/scripts/check-guard-symlinks.sh
@@ -139,6 +139,7 @@ SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
   echo "check-guard-symlinks: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
 FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
 export FLOW_GUARD_REPO_ROOT
 flow_guard_exec check-guard-symlinks 2 "check-guard-symlinks:" "$@"
diff --git a/scripts/check-installed-citations.sh b/scripts/check-installed-citations.sh
index 7db867d3..66f7865d 100755
--- a/scripts/check-installed-citations.sh
+++ b/scripts/check-installed-citations.sh
@@ -42,6 +42,7 @@ set -euo pipefail
   echo "check-installed-citations: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
 FLOW_GUARD_REPO_ROOT="$(cd -P "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
 export FLOW_GUARD_REPO_ROOT
 flow_guard_exec check-installed-citations 2 "check-installed-citations:" "$@"
diff --git a/scripts/check-plan-provenance.sh b/scripts/check-plan-provenance.sh
index 66e7b499..d700dbf9 100755
--- a/scripts/check-plan-provenance.sh
+++ b/scripts/check-plan-provenance.sh
@@ -44,13 +44,14 @@
 # CHECK_PLAN_PROVENANCE_ROOT, only when that is unset — an explicit value,
 # empty included, reaches the guard untouched.
 set -euo pipefail
+[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
+  echo "check-plan-provenance: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
+  exit 2
+}
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
 root="$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)" || {
   echo "check-plan-provenance: cannot resolve the checkout beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
 export CHECK_PLAN_PROVENANCE_ROOT="${CHECK_PLAN_PROVENANCE_ROOT-$root}"
-[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
-  echo "check-plan-provenance: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
-  exit 2
-}
 flow_guard_exec check-plan-provenance 2 "check-plan-provenance:" "$@"
diff --git a/scripts/check-references.sh b/scripts/check-references.sh
index 616b5712..c2031dfb 100755
--- a/scripts/check-references.sh
+++ b/scripts/check-references.sh
@@ -45,6 +45,7 @@ set -euo pipefail
   echo "check-references: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
 FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
 export FLOW_GUARD_REPO_ROOT
 flow_guard_exec check-references 2 "check-references:" "$@"
diff --git a/scripts/check-self-review-report.sh b/scripts/check-self-review-report.sh
index 0883eae8..4a00c094 100755
--- a/scripts/check-self-review-report.sh
+++ b/scripts/check-self-review-report.sh
@@ -144,6 +144,7 @@ set -euo pipefail
   echo "check-self-review-report: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
 FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
 export FLOW_GUARD_REPO_ROOT
 flow_guard_exec check-self-review-report 2 "check-self-review-report:" "$@"
diff --git a/scripts/check-stage-mark-calls.sh b/scripts/check-stage-mark-calls.sh
index 93add206..f19f577d 100755
--- a/scripts/check-stage-mark-calls.sh
+++ b/scripts/check-stage-mark-calls.sh
@@ -149,6 +149,7 @@ SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
   echo "check-stage-mark-calls: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
 FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
 export FLOW_GUARD_REPO_ROOT
 flow_guard_exec check-stage-mark-calls 2 "check-stage-mark-calls:" "$@"
diff --git a/scripts/check-verbatim-moves.sh b/scripts/check-verbatim-moves.sh
index bb0b154c..ee7ff0d8 100755
--- a/scripts/check-verbatim-moves.sh
+++ b/scripts/check-verbatim-moves.sh
@@ -49,6 +49,7 @@ set -euo pipefail
   echo "check-verbatim-moves: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
 FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
 export FLOW_GUARD_REPO_ROOT
 flow_guard_exec check-verbatim-moves 2 "check-verbatim-moves:" "$@"
diff --git a/scripts/lib/flow-guard.sh b/scripts/lib/flow-guard.sh
index 7e8136f5..b5d7b9c1 100644
--- a/scripts/lib/flow-guard.sh
+++ b/scripts/lib/flow-guard.sh
@@ -81,6 +81,24 @@ flow_guard_root() {
   (cd -P "$(dirname -- "${BASH_SOURCE[0]}")" && cd ../.. && pwd)
 }
 
+# flow_guard_cwd_checkout <script> [arguments...] — a guard that judges its
+# own checkout calls this before deriving its root. When the working
+# directory sits in another worktree of the same repository and that
+# worktree carries scripts/<script's name>, exec that copy with the same
+# arguments, so an absolute-path call from a worktree judges the worktree
+# rather than reporting a false clean on this checkout (KAN-902). Otherwise
+# return and let the caller carry on.
+flow_guard_cwd_checkout() {
+  local script="$1" here top common
+  shift
+  here="$(flow_guard_root)" || return 0
+  top="$(git rev-parse --show-toplevel 2>/dev/null)" && top="$(cd -P "$top" && pwd)" || return 0
+  [ "$top" != "$here" ] && [ -f "$top/scripts/${script##*/}" ] || return 0
+  common="$(git -C "$here" rev-parse --path-format=absolute --git-common-dir 2>/dev/null)" || return 0
+  [ -n "$common" ] && [ "$common" = "$(git -C "$top" rev-parse --path-format=absolute --git-common-dir 2>/dev/null)" ] || return 0
+  exec bash "$top/scripts/${script##*/}" "$@"
+}
+
 # flow_guard_exec <name> <cannot-answer-code> <opening> [guard arguments...]
 # — exec the checkout-built flow-guard as guard <name>, or print
 # "<opening> <cause>" to stderr and exit <cannot-answer-code>.
diff --git a/scripts/test-lib-flow-guard.sh b/scripts/test-lib-flow-guard.sh
index 39331f1d..b479dfdd 100755
--- a/scripts/test-lib-flow-guard.sh
+++ b/scripts/test-lib-flow-guard.sh
@@ -224,6 +224,46 @@ for shim in check-cleanup-complete check-panel-reproducer-exit-contract check-ta
   check "case 14: scripts/$shim.sh run by bare filename loads its lib" $?
 done
 
+# 15. An absolute-path call from another worktree of the same repository
+#     runs that worktree's own copy of the shim, never a false answer about
+#     the shim's checkout (KAN-902): the fixture's shim calls
+#     flow_guard_cwd_checkout, and the worktree's flow-guard prints its own
+#     word. From the shim's own checkout, and from a directory in no
+#     repository, the shim's own guard answers.
+new_fixture v15
+cat > "$FIX/scripts/check-x.sh" <<'EOF'
+#!/usr/bin/env bash
+set -euo pipefail
+. "${BASH_SOURCE[0]%/*}/lib/flow-guard.sh"
+flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
+flow_guard_exec check-x 4 "check-x:" "$@"
+EOF
+git -C "$FIX" init -q && git -C "$FIX" add -A \
+  && git -C "$FIX" -c user.name=Test -c user.email=test@example.invalid -c commit.gpgsign=false commit -qm fixture \
+  && git -C "$FIX" worktree add -q --detach "$FIX-wt"
+(FIX="$FIX-wt" && write_main w15)
+run bash -c 'cd "$1-wt" && bash "$1/scripts/check-x.sh" arg' _ "$FIX"
+[ "$RC" -eq 0 ] && [ "$OUT" = "w15 check-x" ]
+check "case 15a: called by absolute path from another worktree, the worktree's own shim answers" $?
+run bash -c 'cd "$1" && bash "$1/scripts/check-x.sh" arg' _ "$FIX"
+[ "$RC" -eq 0 ] && [ "$OUT" = "v15 check-x" ]
+check "case 15b: called from its own checkout, the shim's own guard answers" $?
+mkdir -p "$WORK/norepo"
+run bash -c 'cd "$1" && bash "$2/scripts/check-x.sh" arg' _ "$WORK/norepo" "$FIX"
+[ "$RC" -eq 0 ] && [ "$OUT" = "v15 check-x" ]
+check "case 15c: called from no repository, the shim's own guard answers" $?
+
+# 16. Every guard that judges its own checkout makes the KAN-902 call before
+#     deriving its root.
+for shim in check-ask-silence check-dispatch-paragraphs check-guard-symlinks \
+  check-installed-citations check-plan-provenance check-references \
+  check-self-review-report check-stage-mark-calls check-verbatim-moves; do
+  call="$(grep -n -x 'flow_guard_cwd_checkout "${BASH_SOURCE\[0\]}" "\$@"' "$REPO/scripts/$shim.sh" | cut -d: -f1)"
+  root="$(grep -n -E '^(FLOW_GUARD_REPO_ROOT|root)=' "$REPO/scripts/$shim.sh" | head -1 | cut -d: -f1)"
+  [ -n "$call" ] && [ -n "$root" ] && [ "$call" -lt "$root" ]
+  check "case 16: scripts/$shim.sh calls flow_guard_cwd_checkout before deriving its root" $?
+done
+
 # 10. A shim with no stats/ beside its checkout exits the cannot-answer code.
 rm -rf "$FIX/stats"
 run "$FIX/scripts/check-x.sh"
````

  - [x] **Step 3: RED, then GREEN.** Apply each mutation in turn, restoring it after, and run `bash scripts/test-lib-flow-guard.sh`:
    `local script="$1" here top common` in `scripts/lib/flow-guard.sh` prefixed with `return 0; ` → it fails on `case 15a: called by absolute path from another worktree, the worktree's own shim answers`;
    the `flow_guard_cwd_checkout` line deleted from `scripts/check-verbatim-moves.sh` → it fails on `case 16: scripts/check-verbatim-moves.sh calls flow_guard_cwd_checkout before deriving its root`. Restored, it passes.
    <!-- measured: both mutations together in a scratch clone @ 46db35e0 after the diff → exactly those two FAIL lines; restored → PASS -->
  - [x] **Step 4: Verify.** `bash scripts/test-lib-flow-guard.sh` ends `✓ PASS`; Step 1's commands print `1 violation(s)`;
    `cd stats && go test ./internal/guard -run 'TestCheckReferences|TestCheckInstalledCitations|TestCheckGuardSymlinks|TestCheckDispatchParagraphs|TestCheckStageMarkCalls|TestCheckSelfReviewReport|TestCheckVerbatimMoves|TestCheckPlanProvenance'` passes;
    `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `scripts/check-plan-provenance.sh` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 5: Commit** the eleven files with the subject below.

**Build:** green
**Files:** `scripts/lib/flow-guard.sh`, `scripts/test-lib-flow-guard.sh`, `scripts/check-ask-silence.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh`, `scripts/check-plan-provenance.sh`, `scripts/check-references.sh`, `scripts/check-self-review-report.sh`, `scripts/check-stage-mark-calls.sh`, `scripts/check-verbatim-moves.sh`
**Tests:** `case 15a: called by absolute path from another worktree, the worktree's own shim answers`, `case 15b: called from its own checkout, the shim's own guard answers`, `case 15c: called from no repository, the shim's own guard answers`, `case 16: scripts/$shim.sh calls flow_guard_cwd_checkout before deriving its root`
**Regression:** reverting the commit makes an absolute-path call from a worktree judge the main checkout again — case 15a sees the shim's own guard answer, and case 16 finds no call in any of the nine shims.
**Baseline:** before=16 after=20 <!-- measured: grep -c -E '^[[:space:]]*check [^(]' scripts/test-lib-flow-guard.sh @ 46db35e0 → 16, and 20 after the diff in a scratch clone -->
**Commit:** fix(scripts): run a worktree's own guard on an absolute-path call from it
**After:** none

**Decision:** guard-runs-cwd-worktree-copy

- [x] 4. Skip a parity test whose pinned commit a shallow clone lacks (KAN-905)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `git clone -q --depth 1 file:///Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-10 <scratch>/shallow && (cd <scratch>/shallow/stats && FLOW_GUARD_CACHE_DIR=<scratch>/cache go test ./internal/guard 2>&1 | grep -c '^--- FAIL'); rm -rf <scratch>/shallow <scratch>/cache`:
    it prints `14`.
    <!-- measured: the commands above, under the session scratchpad, @ 46db35e0 → 14; with the diff applied in a worktree of that shallow clone → ok, 0 FAIL -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 46db35e0 after Tasks 1-3's diffs; gofmt -l and go vet ./internal/guard clean; go test ./internal/guard passes in that full clone with no history skip, and in a --depth 1 clone with the history-bound tests skipped
diff --git a/stats/internal/guard/aside_planning_artifacts_test.go b/stats/internal/guard/aside_planning_artifacts_test.go
index fc15790c..03f43b9d 100644
--- a/stats/internal/guard/aside_planning_artifacts_test.go
+++ b/stats/internal/guard/aside_planning_artifacts_test.go
@@ -19,10 +19,7 @@ func apaBashAt(t *testing.T) string {
 	t.Helper()
 	dir := t.TempDir()
 	for _, rel := range []string{"aside-planning-artifacts.sh", "lib/spec-root.sh"} {
-		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "ae805186:scripts/"+rel).Output()
-		if err != nil {
-			t.Fatalf("git show ae805186:scripts/%s: %v", rel, err)
-		}
+		src := historicalFile(t, "../../..", "ae805186", "scripts/"+rel)
 		writeFile(t, filepath.Join(dir, rel), string(src))
 	}
 	return dir + "/aside-planning-artifacts.sh"
diff --git a/stats/internal/guard/check_dispatch_paragraphs_test.go b/stats/internal/guard/check_dispatch_paragraphs_test.go
index c47e5386..fb49c21e 100644
--- a/stats/internal/guard/check_dispatch_paragraphs_test.go
+++ b/stats/internal/guard/check_dispatch_paragraphs_test.go
@@ -230,10 +230,7 @@ func TestCheckDispatchParagraphs(t *testing.T) {
 	// of task 5), each pinned by running the bash at d71a2327 on the same
 	// root and comparing stdout, stderr and exit byte for byte.
 	bash := filepath.Join(t.TempDir(), "check-dispatch-paragraphs.sh")
-	src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "d71a2327:scripts/check-dispatch-paragraphs.sh").Output()
-	if err != nil {
-		t.Fatal(err)
-	}
+	src := historicalFile(t, "../../..", "d71a2327", "scripts/check-dispatch-paragraphs.sh")
 	writeFile(t, bash, string(src))
 	parity := []struct {
 		label string
diff --git a/stats/internal/guard/check_finish_preflight_test.go b/stats/internal/guard/check_finish_preflight_test.go
index 2c17c2b0..25f982d9 100644
--- a/stats/internal/guard/check_finish_preflight_test.go
+++ b/stats/internal/guard/check_finish_preflight_test.go
@@ -389,10 +389,7 @@ func TestCheckFinishPreflight(t *testing.T) {
 	// so each is run against the bash at d71a2327 live on the same fixture.
 	bashDir := filepath.Join(t.TempDir(), "scripts")
 	for _, rel := range []string{"check-finish-preflight.sh", "check-worktree-location.sh", "lib/resolve-remote-base.sh"} {
-		src, err := exec.Command(fixtureGit, "-C", fpRepoRoot, "show", "d71a2327:scripts/"+rel).Output()
-		if err != nil {
-			t.Fatal(err)
-		}
+		src := historicalFile(t, fpRepoRoot, "d71a2327", "scripts/"+rel)
 		writeFile(t, filepath.Join(bashDir, rel), string(src))
 		if err := os.Chmod(filepath.Join(bashDir, rel), 0o755); err != nil {
 			t.Fatal(err)
diff --git a/stats/internal/guard/check_guard_symlinks_test.go b/stats/internal/guard/check_guard_symlinks_test.go
index 8ce9800c..81efbf5e 100644
--- a/stats/internal/guard/check_guard_symlinks_test.go
+++ b/stats/internal/guard/check_guard_symlinks_test.go
@@ -184,10 +184,7 @@ func gsBashParity(t *testing.T, repo string) {
 	t.Helper()
 	scripts := filepath.Join(t.TempDir(), "scripts")
 	for _, rel := range []string{"check-guard-symlinks.sh", "lib/resolve-file.sh", "lib/coverage.sh"} {
-		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "d71a2327:scripts/"+rel).Output()
-		if err != nil {
-			t.Fatal(err)
-		}
+		src := historicalFile(t, "../../..", "d71a2327", "scripts/"+rel)
 		writeFile(t, filepath.Join(scripts, rel), string(src))
 	}
 	cmd := exec.Command("bash", filepath.Join(scripts, "check-guard-symlinks.sh"))
diff --git a/stats/internal/guard/check_panel_fix_single_dispatch_test.go b/stats/internal/guard/check_panel_fix_single_dispatch_test.go
index 5e1cdd72..9e822fbd 100644
--- a/stats/internal/guard/check_panel_fix_single_dispatch_test.go
+++ b/stats/internal/guard/check_panel_fix_single_dispatch_test.go
@@ -259,10 +259,7 @@ func TestCheckPanelFixSingleDispatch(t *testing.T) {
 			t.Skip("jq is not on PATH; the bash at d71a2327 requires it")
 		}
 		script := filepath.Join(t.TempDir(), "check-panel-fix-single-dispatch.sh")
-		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "d71a2327:scripts/check-panel-fix-single-dispatch.sh").Output()
-		if err != nil {
-			t.Fatal(err)
-		}
+		src := historicalFile(t, "../../..", "d71a2327", "scripts/check-panel-fix-single-dispatch.sh")
 		writeFile(t, script, string(src))
 		row := func(key string) string {
 			return `{"key":` + key + `,"role":"panel-fix","sessionToken":"mf-tok"}`
diff --git a/stats/internal/guard/check_stage_mark_calls_test.go b/stats/internal/guard/check_stage_mark_calls_test.go
index f7de11e6..7fa964ce 100644
--- a/stats/internal/guard/check_stage_mark_calls_test.go
+++ b/stats/internal/guard/check_stage_mark_calls_test.go
@@ -328,10 +328,7 @@ func TestCheckStageMarkCalls(t *testing.T) {
 		t.Parallel()
 		tree := t.TempDir()
 		for _, rel := range []string{"scripts/check-stage-mark-calls.sh", "scripts/lib/coverage.sh"} {
-			src, err := exec.Command(fixtureGit, "-C", smcRepoRoot, "show", "d71a2327:"+rel).Output()
-			if err != nil {
-				t.Fatal(err)
-			}
+			src := historicalFile(t, smcRepoRoot, "d71a2327", rel)
 			// The pinned bash pipes the key list into `grep -q` under
 			// pipefail: bash's printf writes "%s" and "\n" separately, so a
 			// grep that matched and exited between them SIGPIPEs printf and
diff --git a/stats/internal/guard/check_workspace_isolation_test.go b/stats/internal/guard/check_workspace_isolation_test.go
index 5ef7e1e8..12d16a3f 100644
--- a/stats/internal/guard/check_workspace_isolation_test.go
+++ b/stats/internal/guard/check_workspace_isolation_test.go
@@ -876,10 +876,7 @@ func wiParity(t *testing.T) {
 	t.Parallel()
 	tree := t.TempDir()
 	for _, rel := range []string{"scripts/check-workspace-isolation.sh", "scripts/lib/resolve-file.sh"} {
-		src, err := exec.Command(fixtureGit, "-C", wiRepoRoot, "show", "c5379c0a:"+rel).Output()
-		if err != nil {
-			t.Fatal(err)
-		}
+		src := historicalFile(t, wiRepoRoot, "c5379c0a", rel)
 		writeFile(t, tree+"/"+rel, string(src))
 	}
 	base := t.TempDir()
diff --git a/stats/internal/guard/helpers_test.go b/stats/internal/guard/helpers_test.go
index ee6a6fa5..013d048f 100644
--- a/stats/internal/guard/helpers_test.go
+++ b/stats/internal/guard/helpers_test.go
@@ -625,3 +625,20 @@ func TestChildEnvTurnsGuardTelemetryOff(t *testing.T) {
 		t.Fatalf("child FLOW_GUARD_TELEMETRY = %q, want off", out)
 	}
 }
+
+// historicalFile is <rev>:<path> read from the repository at root — the bash
+// a parity test pins its port against, as it stood at <rev>. A clone that
+// does not carry <rev>, a shallow one, skips the calling test naming the
+// commit, rather than failing the whole package on history it never had
+// (KAN-905).
+func historicalFile(t *testing.T, root, rev, path string) []byte {
+	t.Helper()
+	if exec.Command(fixtureGit, "-C", root, "cat-file", "-e", rev+"^{commit}").Run() != nil {
+		t.Skipf("%s is not in this clone's history (a shallow clone?) — the parity check against the bash at it cannot run", rev)
+	}
+	src, err := exec.Command(fixtureGit, "-C", root, "show", rev+":"+path).Output()
+	if err != nil {
+		t.Fatalf("git show %s:%s: %v", rev, path, err)
+	}
+	return src
+}
diff --git a/stats/internal/guard/mutate_and_verify_test.go b/stats/internal/guard/mutate_and_verify_test.go
index bb3ba4f5..bd6ea295 100644
--- a/stats/internal/guard/mutate_and_verify_test.go
+++ b/stats/internal/guard/mutate_and_verify_test.go
@@ -845,10 +845,7 @@ func bashAtBase(t *testing.T, rels ...string) string {
 	t.Helper()
 	dir := t.TempDir()
 	for _, rel := range rels {
-		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "d71a2327:scripts/"+rel).Output()
-		if err != nil {
-			t.Fatalf("git show d71a2327:scripts/%s: %v", rel, err)
-		}
+		src := historicalFile(t, "../../..", "d71a2327", "scripts/"+rel)
 		writeFile(t, filepath.Join(dir, rel), string(src))
 	}
 	return dir
diff --git a/stats/internal/guard/project_get_test.go b/stats/internal/guard/project_get_test.go
index 337043ed..26a1cd6d 100644
--- a/stats/internal/guard/project_get_test.go
+++ b/stats/internal/guard/project_get_test.go
@@ -15,10 +15,7 @@ func pgBashDir(t *testing.T) string {
 	t.Helper()
 	dir := t.TempDir()
 	for _, rel := range []string{"project-get.sh", "lib/project-section.sh", "lib/strip-bom.sh"} {
-		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "ae805186:scripts/"+rel).Output()
-		if err != nil {
-			t.Fatalf("git show ae805186:scripts/%s: %v", rel, err)
-		}
+		src := historicalFile(t, "../../..", "ae805186", "scripts/"+rel)
 		writeFile(t, filepath.Join(dir, rel), string(src))
 	}
 	return dir
diff --git a/stats/internal/guard/prove_reproducer_test.go b/stats/internal/guard/prove_reproducer_test.go
index d82ac972..7f3babf6 100644
--- a/stats/internal/guard/prove_reproducer_test.go
+++ b/stats/internal/guard/prove_reproducer_test.go
@@ -312,10 +312,7 @@ func prBash(t *testing.T, env Env, args ...string) (int, string, string) {
 	t.Helper()
 	dir := t.TempDir()
 	for _, rel := range []string{"prove-reproducer.sh", "lib/reproducer-path.sh"} {
-		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "d71a2327:scripts/"+rel).Output()
-		if err != nil {
-			t.Fatalf("git show d71a2327:scripts/%s: %v", rel, err)
-		}
+		src := historicalFile(t, "../../..", "d71a2327", "scripts/"+rel)
 		writeFile(t, filepath.Join(dir, rel), string(src))
 	}
 	writeFile(t, filepath.Join(dir, "run-reproducer.sh"), "")
diff --git a/stats/internal/guard/throwaway_worktree_test.go b/stats/internal/guard/throwaway_worktree_test.go
index d292b79b..0346e122 100644
--- a/stats/internal/guard/throwaway_worktree_test.go
+++ b/stats/internal/guard/throwaway_worktree_test.go
@@ -24,10 +24,7 @@ import (
 // with each placeholder replaced by its value.
 func twFence(t *testing.T, doc, marker string, n int, repl ...string) string {
 	t.Helper()
-	src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "ae805186:"+doc).Output()
-	if err != nil {
-		t.Fatalf("git show ae805186:%s: %v", doc, err)
-	}
+	src := historicalFile(t, "../../..", "ae805186", doc)
 	s := string(src)
 	i := strings.Index(s, marker)
 	if i < 0 {
````

  - [x] **Step 3: RED, then GREEN.** Step 1's commands, run on a commit carrying this diff, print `0`; on its parent they print `14`.
  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/guard -v -run '^(TestAsidePlanningArtifactsParity|TestBreakAndProve|TestThrowawayWorktreeCreate|TestProveReproducerExecBitCheck|TestProjectGetParity|TestCheckStageMarkCalls|TestCheckPanelFixSingleDispatch|TestCheckWorkspaceIsolation|TestCheckGuardSymlinks|TestCheckDispatchParagraphs|TestCheckFinishPreflight|TestThrowawayWorktreeRemoveFoldBack|TestProveReproducerBashParity|TestMutateAndVerify)$' 2>&1 | grep -c 'shallow clone'` prints `0`, and the same selector without `-v` passes.
    <!-- measured: go test ./internal/guard -v in a full scratch clone @ 46db35e0 after the diff → ok, 0 skip lines naming a shallow clone -->
  - [x] **Step 5: Commit** the twelve files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/helpers_test.go`, `stats/internal/guard/aside_planning_artifacts_test.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`, `stats/internal/guard/check_finish_preflight_test.go`, `stats/internal/guard/check_guard_symlinks_test.go`, `stats/internal/guard/check_panel_fix_single_dispatch_test.go`, `stats/internal/guard/check_stage_mark_calls_test.go`, `stats/internal/guard/check_workspace_isolation_test.go`, `stats/internal/guard/mutate_and_verify_test.go`, `stats/internal/guard/project_get_test.go`, `stats/internal/guard/prove_reproducer_test.go`, `stats/internal/guard/throwaway_worktree_test.go`
**Tests:** none — repairs existing tests: `TestAsidePlanningArtifactsParity`, `TestBreakAndProve`, `TestThrowawayWorktreeCreate`, `TestProveReproducerExecBitCheck`, `TestProjectGetParity`, `TestCheckStageMarkCalls`, `TestCheckPanelFixSingleDispatch`, `TestCheckWorkspaceIsolation`, `TestCheckGuardSymlinks`, `TestCheckDispatchParagraphs`, `TestCheckFinishPreflight`, `TestThrowawayWorktreeRemoveFoldBack`, `TestProveReproducerBashParity`, `TestMutateAndVerify`
**Regression:** reverting the commit makes the fourteen tests named above fail again in a `--depth 1` clone, on history the clone does not carry.
**Commit:** test(guard): skip a parity test whose pinned commit a shallow clone lacks
**After:** none

**Decision:** skip-parity-without-history

- [x] 5. Prove an ignored signal stays ignored without a timed look (KAN-907)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -c -F 'time.Sleep(500 * time.Millisecond)' stats/internal/guard/prove_reproducer_test.go stats/internal/guard/mutate_and_verify_test.go`: each file prints `1`.
    <!-- measured: the grep above in this worktree @ 46db35e0 → 1, 1 -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree @ 384f6a7c (Tasks 1-4 committed); gofmt -l and go vet ./internal/guard clean; go test ./internal/guard -run 'TestProveReproducerSignals|TestMutateAndVerify' -count=10 passes; both ignored-SIGHUP subtests fail 10 of 10 with the mutation Step 3 names
diff --git a/stats/internal/guard/mutate_and_verify_test.go b/stats/internal/guard/mutate_and_verify_test.go
index bd6ea295..e06f576a 100644
--- a/stats/internal/guard/mutate_and_verify_test.go
+++ b/stats/internal/guard/mutate_and_verify_test.go
@@ -670,12 +670,11 @@ func mvSignalCase(t *testing.T, fx mvFixtures, sig syscall.Signal, ignore bool)
 		t.Fatal(err)
 	}
 	if ignore {
-		// An ignored signal leaves no trace to wait on; half a second is
-		// long past a handler's restore, which would undo the mutation
-		// under the running harness.
-		time.Sleep(500 * time.Millisecond)
-		g, _ := os.ReadFile(c.guard)
-		mvCheck(t, "ignored signal: the mutation is still applied", strings.Contains(string(g), mvMutationMark), "restored early")
+		// An ignored signal leaves no trace to wait on, so no look is timed
+		// (KAN-907): the signal is pending before the harness is released,
+		// and trapExitSignals handles every signal delivered before the
+		// run's own exit by restoring and dying of it, so the normal-end
+		// check below fails the moment one is handled.
 		mvRelease(t, releaseF)
 		_ = cmd.Wait()
 		o, _ := os.ReadFile(dir + "/stdout")
diff --git a/stats/internal/guard/prove_reproducer_test.go b/stats/internal/guard/prove_reproducer_test.go
index 7f3babf6..2457dd5a 100644
--- a/stats/internal/guard/prove_reproducer_test.go
+++ b/stats/internal/guard/prove_reproducer_test.go
@@ -10,7 +10,6 @@ import (
 	"sync"
 	"syscall"
 	"testing"
-	"time"
 )
 
 // Every case_N of scripts/test-prove-reproducer.sh at d71a2327, one subtest
@@ -576,15 +575,12 @@ func TestProveReproducerSignals(t *testing.T) {
 			if err := syscall.Kill(target, c.sig); err != nil {
 				t.Fatal(err)
 			}
-			if c.ignore {
-				// An ignored signal leaves no trace to wait on; half a second
-				// is long past a handler's cleanup, which would remove the
-				// scratch under the running leg.
-				time.Sleep(500 * time.Millisecond)
-				if left, _ := filepath.Glob(tmp + "/prove-reproducer.*/scratch"); len(left) != 1 {
-					t.Fatalf("the ignored %v removed the scratch under the running leg", c.sig)
-				}
-			}
+			// An ignored signal leaves no trace to wait on, so no look is
+			// timed (KAN-907): the signal is pending before the leg is
+			// released, and a handled one ends the run in exit 129 -- its
+			// re-raise lands on the restored ignore, and the caller's exit
+			// waits on the handler -- so the finished run's exit 1 below
+			// fails the moment one is handled.
 			done := make(chan struct{})
 			go func() { _ = cmd.Wait(); close(done) }()
 			if c.ignore {
@@ -602,7 +598,7 @@ func TestProveReproducerSignals(t *testing.T) {
 			o, _ := os.ReadFile(dir + "/stdout")
 			e, _ := os.ReadFile(dir + "/stderr")
 			if c.ignore {
-				if ws.Signaled() || !strings.Contains(string(o), "== post-fix leg exit: 1\n") {
+				if ws.Signaled() || ws.ExitStatus() != 1 || !strings.Contains(string(o), "== post-fix leg exit: 1\n") {
 					t.Fatalf("status %v, want a finished run\nstdout:\n%s\nstderr:\n%s", cmd.ProcessState, o, e)
 				}
 			} else if !ws.Signaled() || ws.Signal() != c.sig {
````

  - [x] **Step 3: RED, then GREEN.** In `trapExitSignals` (`stats/internal/guard/provereproducer.go`) replace `if !signal.Ignored(s) {` with `if true || !signal.Ignored(s) {`:
    `cd stats && go test ./internal/guard -run 'TestProveReproducerSignals|TestMutateAndVerify' -count=5` fails on `TestProveReproducerSignals/ignored_SIGHUP` (`status exit status 129, want a finished run`) and `TestMutateAndVerify/ignored_SIGHUP_mid-run` (`status exit status 129`) on every run. Restore the line: it passes.
    <!-- measured: the replacement above in this worktree @ 384f6a7c after the diff, -count=10 → each subtest FAIL 10 of 10, both reporting exit status 129; restored → ok, 10 of 10. Without the diff's `ws.ExitStatus() != 1` term the prove-reproducer subtest passes under the mutation, 3 of 3 (the handled SIGHUP's re-raise lands on the restored ignore, and the run exits 129 after printing the post-fix leg line) -->
  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/guard -run 'TestProveReproducerSignals|TestMutateAndVerify'` passes; Step 1's grep prints `0` for each file.
  - [x] **Step 5: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/prove_reproducer_test.go`, `stats/internal/guard/mutate_and_verify_test.go`
**Tests:** none — repairs existing tests: `TestProveReproducerSignals`, `TestMutateAndVerify`
**Regression:** reverting the commit restores the two half-second timed looks and drops prove-reproducer's exit-1 pin; no verdict changes on an unmutated tree, since the finished-run assertions stay.
**Commit:** test(guard): prove an ignored signal stays ignored without a timed look
**After:** Task 4

**Decision:** ignored-signal-no-timed-look
