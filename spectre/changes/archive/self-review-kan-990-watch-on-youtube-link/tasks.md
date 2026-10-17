# self-review-kan-990-watch-on-youtube-link

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-17-self-review-kan-990-watch-on-youtube-link-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`. Never merge or push either repository.
- **Task 1 commits in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-990-watch-on-youtube-link`** (branch `spectre/kan-990-watch-on-youtube-link`), never in this agents worktree; Tasks 2 and 3 commit in this agents worktree.
- Each task's change is the unified diff its step gives, applied with `git apply` from the root of the repository its task names — every diff was applied and checked there (`design.md`, **Context**). A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line a diff adds or rewords is copied verbatim: this change's `verbatim-moves.txt` acknowledges each such sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.

## Review focus

- A satellite worktree whose only delta is a `link.md` must still be asked a subject: the `nested_link` case stays green (Task 2).
- `commit` must default the subject only when none was passed — a subject the parent passes for the canonical worktree still wins (Task 2).
- No `-<worktree basename>` key suffix may survive in `skills/` or `scripts/` (Task 3).

- [x] 1. gymie includes the regression suite for any ui paths change
  - [x] **Step 1: Reproduce before touching anything; do not guess.** From `/Users/tweety53/Projects/gymie-worktrees/kan-990-watch-on-youtube-link`, run `grep -n 'Regression suite' .flow/project.md` and confirm the row's Notes name no `ui paths` — the defect: a UI change that re-baselines captures gets no playwright worktree.
  - [x] **Step 2: Apply the edit.** From `/Users/tweety53/Projects/gymie-worktrees/kan-990-watch-on-youtube-link`, apply this diff:

````diff verified:git apply --check in the gymie worktree at 909ade0d
diff --git a/.flow/project.md b/.flow/project.md
index f2f787cf..bd46089e 100644
--- a/.flow/project.md
+++ b/.flow/project.md
@@ -15,7 +15,7 @@ uncommitted work.
 | Backend + gateway | `/Users/tweety53/Projects/gymie` | Kotlin / Spring Boot / Gradle | http://localhost:8080 | Primary repo. Include for any backend, API, Flyway, or module work. |
 | KMP frontend | `/Users/tweety53/Projects/gymie-frontend` | Kotlin Multiplatform / Compose | http://localhost:3000 | Include when the change mentions KMP, Compose, web, or mobile. |
 | Admin frontend | `/Users/tweety53/Projects/gymie-admin-frontend` | Next.js / TypeScript | http://localhost:3001 | Include when the change mentions the admin panel. |
-| Regression suite | `/Users/tweety53/Projects/gymie-playwright` | Node / Playwright | — | Include when the change adds or edits a capture spec, a `.mockups` sidecar, or otherwise touches this checkout. Also declared as `regression checkout` under `## visual verification` below — that row is `flow.visual-verify`'s own resolution of the same repository, not a second one. |
+| Regression suite | `/Users/tweety53/Projects/gymie-playwright` | Node / Playwright | — | Include when the change adds or edits a capture spec, a `.mockups` sidecar, touches any `## visual verification` `ui paths` path — `capture` re-baselines that screen's PNGs here — or otherwise touches this checkout. Also declared as `regression checkout` under `## visual verification` below — that row is `flow.visual-verify`'s own resolution of the same repository, not a second one. |
 
 **The URLs above are the main checkout's, which keeps every documented default.** An apply worktree
 gets its own database, cache index, bucket and block of application ports automatically — see
````

  - [x] **Step 3: Verify.** From the same worktree, `git diff --cached --stat` shows `.flow/project.md` alone and `grep -c 'ui paths. path — .capture. re-baselines' .flow/project.md` prints `1`. Commit there.

**Build:** green
**Files:** `.flow/project.md` (gymie repository, worktree `/Users/tweety53/Projects/gymie-worktrees/kan-990-watch-on-youtube-link`)
**Tests:** **none** — project configuration prose.
**Regression:** reverting the commit leaves a UI change that re-baselines captures without a playwright worktree, so `capture` runs in the main checkout again.
**Commit:** docs(flow): include the regression suite for any ui paths change
**After:** none

**Decision:** regression-row-names-ui-paths

- [x] 2. A link.md-only canonical worktree lands as the merge-order commit
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Apply the test hunks alone (`git apply --include='scripts/test-integrate-change.sh' <the diff in step 2>`) and run `bash scripts/test-integrate-change.sh`: it prints `162 passed, 3 failed`, the FAILs `link_only_canon: prepare's NEXT asks a subject for b and none for a`, `link_only_canon: commit with no subject for a exits 0` and `link_only_canon: a's link.md lands in a commit subjected chore(spectre): merge order`.
  - [x] **Step 2: Implement the fix.** Apply the rest of this diff (`git apply --exclude='scripts/test-integrate-change.sh' <the diff>`):

````diff verified:git apply --check at 04a3a161; with it applied scripts/test-integrate-change.sh prints 165 passed, 0 failed; with integrate-change.sh reverted 162 passed, 3 failed
diff --git a/scripts/integrate-change.sh b/scripts/integrate-change.sh
index cff43f8a..18af1690 100755
--- a/scripts/integrate-change.sh
+++ b/scripts/integrate-change.sh
@@ -36,7 +36,8 @@
 #          and records the reshape base in the worktree's own git dir
 #          (`git rev-parse --git-path flow-reshape-base`) for `commit`. Its
 #          NEXT names `<worktree>=<impl subject>` only for a worktree whose
-#          implementation commit would not be empty (impl_paths below).
+#          implementation commit would not be empty (impl_paths below) and
+#          is not the canonical worktree's link.md alone (link_only below).
 #          --accept-foreign / --accept-outstanding: the operator chose
 #          Continue at that stop; its lines still print. A
 #          VISUAL-VERIFY-MISSING is STOP: visual-verify, which
@@ -55,7 +56,9 @@
 #          worktree in `## Merge order`, `spectre archive` (each
 #          <name>-fix-N too) unless already archived, commit-archive. A
 #          worktree with nothing on the implementation side takes no
-#          subject; any other without one stops `commit-split`. An
+#          subject; the canonical one whose only implementation path is a
+#          link.md takes `chore(spectre): merge order`; any other without
+#          one stops `commit-split`. An
 #          archived change without new work skips straight to commit-archive.
 #          Then, each behind the visual-verify gate, the --force-with-lease
 #          branch push per worktree, in merge
@@ -348,6 +351,17 @@ impl_paths() {
     awk -v p="$p" 'index($0, p) != 1 || substr($0, length(p) + 1) ~ /\/link\.md$/'
 }
 
+# link_only <i> <rev>: worktree <i> is the canonical one and every path
+# impl_paths gives it is a link.md — the merge-order record alone, whose
+# implementation commit takes the literal subject `chore(spectre): merge
+# order` (git-boundaries.md, Planning commits) rather than one asked for.
+link_only() {
+  local p
+  [ "${WT[$1]}" = "$CANON" ] || return 1
+  p="$(impl_paths "${WT[$1]}" "$2")"
+  [ -n "$p" ] && ! printf '%s\n' "$p" | grep -qv '/link\.md$'
+}
+
 write_state() { # <state> [<keep-json-array>] [<pr-url>]
   local j
   j="$(flow state get -C "$MAIN" "$NAME")" || return 1
@@ -517,7 +531,7 @@ cmd_prepare() {
     sp="$(spec_root_leaf "${WT[i]}")"
     git -C "${WT[i]}" diff --no-renames --stat "$rb" -- . ":(exclude)$sp/changes/"
     git -C "${WT[i]}" ls-files --others --exclude-standard -- . ":(exclude)$sp/changes/" | sed 's/^/ new: /'
-    [ -n "$(impl_paths "${WT[i]}" "$rb")" ] && subjects="$subjects ${WT[i]}=<impl subject>"
+    [ -n "$(impl_paths "${WT[i]}" "$rb")" ] && ! link_only "$i" "$rb" && subjects="$subjects ${WT[i]}=<impl subject>"
   done
   local narrative="$live/narrative.md"
   [ "$ARCHIVED" -eq 1 ] && narrative="$arch/narrative.md"
@@ -561,6 +575,7 @@ cmd_commit() {
       for s in ${SUBJECTS[@]+"${SUBJECTS[@]}"}; do
         case "${s%%=*}" in "${WT[o]}"|"${REPO[o]}") subj="${s#*=}" ;; esac
       done
+      [ -n "$subj" ] || ! link_only "$o" "$rb" || subj="chore(spectre): merge order"
       [ -n "$subj" ] || [ -z "$(impl_paths "${WT[o]}" "$rb")" ] ||
         stop commit-split "no implementation subject given for ${WT[o]} — pass ${WT[o]}=<subject>"
       run "$SCRIPT_DIR/commit-split.sh" "${WT[o]}" "$NAME" "$subj" "$PLAN_MSG"
diff --git a/scripts/test-integrate-change.sh b/scripts/test-integrate-change.sh
index 8e5bd310..bca0d36c 100755
--- a/scripts/test-integrate-change.sh
+++ b/scripts/test-integrate-change.sh
@@ -63,6 +63,11 @@
 #   planning_only   b's branch changes nothing outside spectre/changes/:
 #                   prepare's NEXT asks no subject for b, and commit with no
 #                   subject for b commits b's planning delta alone.
+#   link_only_canon a's branch changes nothing outside spectre/changes/ but
+#                   its change directory's link.md, the merge-order record:
+#                   prepare's NEXT asks no subject for a, and commit with no
+#                   subject for a lands that link.md in an implementation
+#                   commit subjected `chore(spectre): merge order`.
 #   commit_pushes   both branches were pushed before the reshape: commit
 #                   leaves each remote spectre/demo at its reshaped HEAD, so
 #                   a later plain branch-backup push is a fast-forward.
@@ -487,6 +492,24 @@ check "planning_only: commit with no subject for b exits 0" '[ "$RC" -eq 0 ]'
 check "planning_only: b's branch is the base plus its one planning commit" \
   '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s main..HEAD | tr "\n" "|")" = "chore(spectre): plan demo|" ]'
 
+# --- link_only_canon ---
+fixture
+g -C "$WORK/a-worktrees/demo" reset -q --hard main
+c="$WORK/a-worktrees/demo/spectre/changes/demo"
+mkdir -p "$c"
+printf '# Tasks\n\n- [x] 1. Do the work\n' >"$c/tasks.md"
+printf '# demo\n\n## Merge order\n\n1. `b`\n2. `.`\n' >"$c/link.md"
+g -C "$WORK/a-worktrees/demo" add -A && g -C "$WORK/a-worktrees/demo" commit -qm "chore(spectre): merge order"
+sut prepare "$WORK/a" demo
+check "link_only_canon: prepare exits 0" '[ "$RC" -eq 0 ]'
+check "link_only_canon: prepare's NEXT asks a subject for b and none for a" \
+  'printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/b-worktrees/demo=<impl subject>" && ! printf "%s" "$OUT" | tail -1 | grep -qF "$WORK/a-worktrees/demo="'
+printf '\n## integrate run\n\nnarrative\n' >>"$c/narrative.md"
+sut commit "$WORK/a" demo "$WORK/b-worktrees/demo=feat(b): do the work"
+check "link_only_canon: commit with no subject for a exits 0" '[ "$RC" -eq 0 ]'
+check "link_only_canon: a's link.md lands in a commit subjected chore(spectre): merge order" \
+  '[ "$(git -C "$WORK/a-worktrees/demo" log --format=%s --diff-filter=A -1 -- spectre/changes/archive/demo/link.md spectre/changes/demo/link.md)" = "chore(spectre): merge order" ]'
+
 # --- commit_pushes ---
 fixture
 for r in a b; do g -C "$WORK/$r-worktrees/demo" push -q -u origin spectre/demo; done
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index 7dacb2a1..6d4d05ca 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -226,7 +226,10 @@ All three routes then commit the work, in **two** commits and never one: the imp
 subject `<type>(<module>): <what the implementation does>` with `<module>` naming the area the
 reshaped diff carries, then the whole `<project>/spectre/changes/` planning delta — every planning
 commit's content, this run's narrative and anything the operator edited. The landed branch is therefore: the implementation commit, then that one
-planning commit.
+planning commit. The canonical worktree whose implementation side is its change directory's
+`link.md` alone — the merge-order record — takes no subject: its implementation commit is
+`chore(spectre): merge order`, the subject **Planning commits** gives that record's write
+(`git-boundaries.md`), never the planning commit's own.
 
 **Nothing under `<abs-worktree>/.superpowers/sdd/` is committed** — not the rendered ledger and
 panel record, not the brainstorm design document. They are worktree-lifetime files, removed with
diff --git a/skills/flow/integrate.md b/skills/flow/integrate.md
index 82f92fef..9958547d 100644
--- a/skills/flow/integrate.md
+++ b/skills/flow/integrate.md
@@ -119,7 +119,9 @@ integrate-change.sh commit <main-checkout> <name> <worktree>="<type>(<module>):
 `<type>`, `<module>` and `<what the implementation does>` are derived from the reshaped diff — the
 diffstat `prepare` printed under each `WORKTREE:` line, one subject per worktree the `NEXT:` line
 names. A worktree whose reshaped diff holds nothing outside `<project>/spectre/changes/` and no change directory's
-`link.md` has an empty implementation commit, so `NEXT:` names no subject for it and `commit` asks none. On
+`link.md` has an empty implementation commit, so `NEXT:` names no subject for it and `commit` asks none — nor
+for the canonical worktree whose only such path is its `link.md`, whose subject **Finish
+contract** fixes (`skills/flow-contracts/finish-contract-run1.md`). On
 **Continue** at **1** the planning commit also lists the outstanding work: pass it as
 `--outstanding <text>`, which `commit` writes into that commit's body as `Outstanding: <text>` under
 its own `chore(spectre): plan <name>` subject. An archived change without new work takes no subject: its `NEXT:` names
````

  - [x] **Step 3: Verify.** `bash scripts/test-integrate-change.sh` prints `0 failed`; from the repository root `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and `scripts/check-installed-citations.sh` exit 0. Commit.

**Build:** green
**Files:** `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow-contracts/finish-contract-run1.md`, `skills/flow/integrate.md`
**Tests:** `link_only_canon`
**Regression:** reverting `integrate-change.sh` makes `prepare` ask a subject for the canonical worktree again and `commit` stop at `STOP: commit-split` without one; three `link_only_canon` checks fail.
**Baseline:** before=31 after=32
<!-- measured: grep -c '^# --- ' scripts/test-integrate-change.sh @ 04a3a161 (before; after with this task's diff applied) -->
**Commit:** fix(scripts): land a link.md-only canonical worktree as the merge-order commit
**After:** none

**Decision:** link-only-canon-merge-order-subject

- [x] 3. Suffix verifier keys with the repository name
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -rn 'worktree basename\|worktree-basename' skills/flow scripts` and confirm the hits in `visual-verify.md`, `verify-and-handoff.md`, `verify-fix-loop.md`, `visual-verify-tooling-analysis.md`, `check-verify-green.sh` and `check-visual-verify-dispatched.sh` — the defect: every worktree's basename is the change name, so the suffix tells none apart.
  - [x] **Step 2: Apply the edit.** Apply this diff:

````diff verified:git apply --check at 04a3a161; with it and Task 2's diff applied check-references, check-markdown-integrity, check-dispatch-paragraphs, check-installed-citations, check-ask-silence and check-normative-inventory exit 0 and check-verbatim-moves reports 0 violations
diff --git a/scripts/check-verify-green.sh b/scripts/check-verify-green.sh
index 2dd05e69..d6108915 100755
--- a/scripts/check-verify-green.sh
+++ b/scripts/check-verify-green.sh
@@ -9,7 +9,7 @@
 #
 # Reads `flow record dispatches -change <change-name> -C <worktree>` and
 # judges this session token's inline verify rows — role `verifier`, key
-# `verify` or `verify-<worktree basename>`; the visual verifier's
+# `verify` or `verify-<repository name>`; the visual verifier's
 # `visual-verify…` rows never count. Only the latest such row per worktree is
 # judged: rows group by key with a `-fix-<k>` suffix removed
 # (skills/flow/verify-fix-loop.md step 4), and the last row of each group in
diff --git a/scripts/check-visual-verify-dispatched.sh b/scripts/check-visual-verify-dispatched.sh
index 63ac3d53..97b886b6 100755
--- a/scripts/check-visual-verify-dispatched.sh
+++ b/scripts/check-visual-verify-dispatched.sh
@@ -70,7 +70,7 @@
 # <name> -C <canonical worktree when given>` answers with the decoded JSON array of every
 # dispatch row this change has ever recorded, across every run. A row
 # counts when its `role` is `verifier`, its `key` starts with
-# `visual-verify` (`visual-verify`, `visual-verify-<worktree-basename>`,
+# `visual-verify` (`visual-verify`, `visual-verify-<repository-name>`,
 # `visual-verify-2`, `visual-verify-retry`, and combinations — this guard
 # does not validate the key's exact shape, only that the stage's own
 # canonical prefix was used; a malformed key is a different guard's job,
diff --git a/skills/flow/verify-and-handoff.md b/skills/flow/verify-and-handoff.md
index d58b3309..b3f03676 100644
--- a/skills/flow/verify-and-handoff.md
+++ b/skills/flow/verify-and-handoff.md
@@ -112,7 +112,7 @@ passing re-run as license to skip the rest of the list — every remaining comma
 above still runs.
 
 **Recording.** One `dispatches` row per worktree the list runs for, `-role verifier -key verify -model <parent
-model> -effort <parent effort> -agent-id inline`, suffixed `-<worktree basename>` when this
+model> -effort <parent effort> -agent-id inline`, suffixed `-<repository name>` when this
 run's resolved set holds more than one worktree — the same convention
 **The verifier dispatch** (`skills/flow/visual-verify.md`) uses for the verifier's rows. `begin` is
 recorded before the first command in the list; `end` after the `## Report` is written, carrying
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index 78ff6d3d..0ed54c72 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -58,7 +58,7 @@ One round runs these steps in order:
 4. **Re-capture and re-run.** Run **Verify** (`skills/flow/verify-and-handoff.md`) again from its
    `begin` mark, then **Visual verification** from its `begin` mark, so the fix's own lint, tests and
    captures are what the handoff reports. This round's `verify` and `visual-verify` dispatch keys
-   carry the suffix `-fix-<k>`, before any `-<worktree basename>` suffix. **The `verify` re-run
+   carry the suffix `-fix-<k>`, before any `-<repository name>` suffix. **The `verify` re-run
    runs only in a worktree whose HEAD moved since its previous verify of this run, or whose test
    commands drive the stack of one that did** — an unmoved worktree's lint and tests already
    passed on that same HEAD. **The `visual-verify`
@@ -84,8 +84,8 @@ One round runs these steps in order:
 
 **One full run closes the loop.** A scoped re-run whose final report carries no fixable defect is
 followed by **Visual verification** once more from its `begin` mark, unscoped — no round base,
-every motion of the change — its dispatch keys suffixed `-fix-<k>-full` before any `-<worktree
-basename>` suffix. A fixable defect it reports opens the next round; its clean report is the one the
+every motion of the change — its dispatch keys suffixed `-fix-<k>-full` before any `-<repository
+name>` suffix. A fixable defect it reports opens the next round; its clean report is the one the
 handoff reports.
 
 ## No cap
diff --git a/skills/flow/visual-verify-tooling-analysis.md b/skills/flow/visual-verify-tooling-analysis.md
index 2ff62ad0..b0306877 100644
--- a/skills/flow/visual-verify-tooling-analysis.md
+++ b/skills/flow/visual-verify-tooling-analysis.md
@@ -52,7 +52,7 @@ this task, as written:
 `<n>` is this fix run's ordinal, the one `flow.document-fix` recorded.
 
 **Recording.** The parent records the dispatch `-role planner`, `-task` omitted, `-model
-opus -effort high`, `-key tooling-analysis-<n>`, suffixed `-<worktree basename>` under
+opus -effort high`, `-key tooling-analysis-<n>`, suffixed `-<repository name>` under
 the verifier's rule — the pair's semantics are section 4 of `skills/flow/implement.md`, cited here,
 not restated. **Handshake** as **The verifier dispatch** (`skills/flow/visual-verify.md`) states it, compared against `opus`.
 
diff --git a/skills/flow/visual-verify.md b/skills/flow/visual-verify.md
index 05e4bb15..b22db43a 100644
--- a/skills/flow/visual-verify.md
+++ b/skills/flow/visual-verify.md
@@ -68,8 +68,9 @@ own system prompt>`.
 > one pass.
 
 **Recording.** The parent records each dispatch, `-role verifier`, `-task` omitted, `-model
-opus -effort low`, `-key visual-verify`, suffixed `-<worktree basename>` when this run's resolved set holds
-more than one worktree — the pair's semantics are section 4 of `skills/flow/implement.md`, cited
+opus -effort low`, `-key visual-verify`, suffixed `-<repository name>` when this run's resolved set holds
+more than one worktree — the worktree's main checkout's basename, the name
+`check-visual-trigger.sh --repo` takes, since every worktree's own basename is the change name — the pair's semantics are section 4 of `skills/flow/implement.md`, cited
 here, not restated.
 Its `begin` carries `-diff-base` with the sha `git -C <worktree> rev-parse HEAD` prints before the dispatch, the HEAD it verifies: `check-visual-verify-dispatched.sh` counts a verifier row only for that HEAD, and only while no UI path the change touches has changed since.
 
@@ -83,8 +84,8 @@ check-verify-report.sh <abs-worktree>/.superpowers/sdd/verify-report-<key>.md <m
 the guard's header is canonical for the report's `steps:` line and its admissible statuses. Exit 0
 → the report is complete; read it. Exit 1 → **the report is incomplete and is never accepted,
 whatever its prose says**: close the dispatch `-outcome aborted` and dispatch its remainder at
-once, recorded under the incomplete dispatch's own key suffixed `-rest` before any `-<worktree
-basename>` suffix, its prompt identical to the incomplete dispatch's plus that report verbatim
+once, recorded under the incomplete dispatch's own key suffixed `-rest` before any `-<repository
+name>` suffix, its prompt identical to the incomplete dispatch's plus that report verbatim
 under a `## Incomplete attempt` heading and the guard's verdict line, which names the steps it
 runs; it carries the incomplete report's lines for the steps already done into its own. A
 remainder completes its dispatch rather than adding one, so it counts toward no re-dispatch and
@@ -98,7 +99,7 @@ does, below.
 
 **A complete `## Report` carrying any non-zero exit is re-dispatched, and no count ends the
 re-dispatches.** Each is recorded under `-key visual-verify-<n>`, `<n>` counting from 2 — the same
-`-<worktree basename>` suffix rule — with a prompt identical to the first plus every earlier
+`-<repository name>` suffix rule — with a prompt identical to the first plus every earlier
 `## Report` verbatim under a `## Previous attempt` heading, so the verifier can tell an
 environmental failure (a build the worktree lacked, a flaky harness) from a defect in the branch.
 From the second report on, a non-zero exit whose block cause is `test-failure` or
````

  - [x] **Step 3: Verify.** `grep -rn 'worktree basename\|worktree-basename' skills/flow scripts` prints nothing; from the repository root `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify.md`, `skills/flow/verify-and-handoff.md`, `skills/flow/verify-fix-loop.md`, `skills/flow/visual-verify-tooling-analysis.md`, `scripts/check-verify-green.sh`, `scripts/check-visual-verify-dispatched.sh`
**Tests:** **none** — no code builds the key; `check-verbatim-moves.sh` pins the reworded sentences.
**Regression:** reverting the commit keys two worktrees' verifiers alike again, so the second `begin` overwrites the first's `-diff-base`.
**Commit:** docs(flow): suffix verifier keys with the repository name
**After:** none

**Decision:** key-suffix-repository-name
