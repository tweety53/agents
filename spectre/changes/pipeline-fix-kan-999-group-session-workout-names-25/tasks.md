# pipeline-fix-kan-999-group-session-workout-names-25

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `remove-change-worktrees.sh` check 4 no longer stops run 2 cleanup on a worktree's own
copy of a main-checkout local config such as `.env`.

**Architecture:** in `stats/internal/guard/removechangeworktrees.go`, `rcwIdentical` is replaced by
`rcwCounterpart` (two distinct regular files) plus `rcwSameBytes`, which picks the label
`preserved: identical in` or `preserved: differs from`; the shim header and
`finish-contract-run2.md` state the rule.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or other
  trailer; never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed
  with the plan.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at 94818c76 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.
- Every Go test command runs with a Bash `timeout: 600000`.

## Review Focus

- An entry with no main-checkout counterpart still exits 3 (subtest "an unclassified entry"), and
  one the main checkout reaches through a symlink still exits 3
  (`samefile-not-a-counterpart`).
- The differing case proceeds and stays visible as `— preserved: differs from <main-checkout>`
  (`differs-label-kept-visible`).

- [ ] 1. Preserve a worktree's copy of a main-checkout local config
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf25-task1.diff`, apply its test hunk alone
    (`git apply --include='stats/internal/guard/remove_change_worktrees_test.go' /tmp/pf25-task1.diff`)
    and run `cd stats && make -s web-dist && go test ./internal/guard -run TestRemoveChangeWorktrees -count=1`:
    `--- FAIL: TestRemoveChangeWorktreesDisclose/an_entry_that_differs_from_the_main_checkout's_proceeds`.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/remove_change_worktrees_test.go' /tmp/pf25-task1.diff`):

````diff verified:applied with git apply to this worktree at 94818c76; go test ./internal/guard -count=1 passes with it and the new subtest fails with removechangeworktrees.go untouched; gofmt, check-references, check-guard-symlinks, check-markdown-integrity and check-verbatim-moves (with this change's verbatim-moves.txt) exit 0
diff --git a/scripts/remove-change-worktrees.sh b/scripts/remove-change-worktrees.sh
index 8b661656..3129ad62 100755
--- a/scripts/remove-change-worktrees.sh
+++ b/scripts/remove-change-worktrees.sh
@@ -21,9 +21,11 @@
 #     in a `.DS_Store` anywhere and, in a worktree whose root holds a
 #     `playwright.config.*`, a `.png` at that root — a screenshot spec's
 #     output; a `.png` in any subdirectory stays unclassified.
-#     An unclassified entry that is a regular file byte-identical to the same
-#     path in the repository's main checkout (the first entry of its worktree
-#     list) is preserved — it exists there.
+#     An unclassified entry that is a regular file with a distinct regular
+#     file at the same path in the repository's main checkout (the first
+#     entry of its worktree list) is preserved — a per-worktree copy of a
+#     main-checkout local config, identical or not; an entry with no such
+#     counterpart stays unpreserved.
 #   the disclosure stop — without --proceed, an unpreserved unclassified
 #     entry or any wave-group copy stops here, before check 5 runs anything;
 #     preserved entries alone are shown and the removal proceeds.
@@ -53,6 +55,7 @@
 #   HELD: <worktree> — <pid> <cwd>[; ...]        check 6 found a live process
 #   UNCLASSIFIED: <worktree> — <path>            check 4, one per entry
 #   UNCLASSIFIED: <worktree> — <path> — preserved: identical in <main-checkout>
+#   UNCLASSIFIED: <worktree> — <path> — preserved: differs from <main-checkout>
 #   REGENERATABLE: <worktree> — <count>          check 4
 #   SKIPPED: check 5 — ## stop declares no fenced command   a declared `## stop`
 #     whose body carries no fence; check 5 is skipped, never silently
diff --git a/skills/flow-contracts/finish-contract-run2.md b/skills/flow-contracts/finish-contract-run2.md
index 8d3fae39..5097bf9a 100644
--- a/skills/flow-contracts/finish-contract-run2.md
+++ b/skills/flow-contracts/finish-contract-run2.md
@@ -143,7 +143,7 @@ the handoff's **Remote branch:** — deleted, already gone, or not deleted. Exit
 `worktrees`. Exit 2: stop and report. Exit 3: the disclosure stop — nothing removed and nothing
 run; relay its `UNCLASSIFIED:` and `DISCLOSE:` lines per check 4 and the wave-group copies below,
 ask only what they ask, then call again with `--proceed`.
-An unclassified entry byte-identical to the same path in the repository's main checkout is preserved: its `UNCLASSIFIED:` line ends `— preserved: identical in <main-checkout>`, it never makes the disclosure stop, and it is relayed per check 4 whatever the exit.
+An unclassified entry that is a regular file with a distinct regular file at the same path in the repository's main checkout — a per-worktree copy of a main-checkout local config such as `.env` — is preserved: its `UNCLASSIFIED:` line ends `— preserved: identical in <main-checkout>`, or `— preserved: differs from <main-checkout>` when the bytes differ, it never makes the disclosure stop, and it is relayed per check 4 whatever the exit; an entry with no counterpart there is unpreserved and still stops.
 
 **Check 6 requires the orchestrating shell's own cwd to be outside every worktree in the resolved
 set before it runs.** `cd` out first, for every worktree, before
diff --git a/stats/internal/guard/remove_change_worktrees_test.go b/stats/internal/guard/remove_change_worktrees_test.go
index 96d9d072..c8cc87b7 100644
--- a/stats/internal/guard/remove_change_worktrees_test.go
+++ b/stats/internal/guard/remove_change_worktrees_test.go
@@ -432,19 +432,21 @@ func TestRemoveChangeWorktreesDisclose(t *testing.T) {
 		})
 	}
 
-	t.Run("an entry that differs from the main checkout's stops", func(t *testing.T) {
+	// kan-999: a worktree's own copy of a main-checkout local config — a .env
+	// carrying the worktree's isolated ports — is preserved too: shown, and
+	// the removal proceeds in the same call.
+	t.Run("an entry that differs from the main checkout's proceeds", func(t *testing.T) {
 		t.Parallel()
 		fx := rcwNewFx(t)
-		fx.g.write(fx.wt+"/.env", "SECRET=1")
-		fx.g.write(fx.repo+"/.env", "SECRET=2")
-		code, out, _ := fx.run(t, nil)
-		if code != 3 {
-			t.Fatalf("exit %d, stdout %q; want 3", code, out)
+		fx.g.write(fx.wt+"/.env", "PORT=3100")
+		fx.g.write(fx.repo+"/.env", "PORT=3000")
+		code, out, errb := fx.run(t, nil)
+		if code != 0 || rcwExists(fx.wt) || len(rcwLines(out, "DISCLOSE: ")) != 0 {
+			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
 		}
-		if l := rcwLines(out, "UNCLASSIFIED: "); len(l) != 1 || !strings.HasSuffix(l[0], " — .env") {
-			t.Errorf("UNCLASSIFIED lines %q, want .env unpreserved", l)
+		if l := rcwLines(out, "UNCLASSIFIED: "); len(l) != 1 || !strings.Contains(l[0], " — .env — preserved: differs from ") {
+			t.Errorf("UNCLASSIFIED lines %q, want .env shown as preserved, differing", l)
 		}
-		fx.assertUntouched(t)
 	})
 
 	// A main-checkout directory symlinked into the worktree reaches the
diff --git a/stats/internal/guard/removechangeworktrees.go b/stats/internal/guard/removechangeworktrees.go
index 7d3cbda7..739ec8f2 100644
--- a/stats/internal/guard/removechangeworktrees.go
+++ b/stats/internal/guard/removechangeworktrees.go
@@ -58,7 +58,8 @@ func removeChangeWorktrees(args []string, env Env, stdout, stderr io.Writer) int
 	}
 	listed, copies := changeWorktrees(porcelain, name)
 	// The main checkout is the porcelain's first entry, whatever <repo> named:
-	// an unclassified entry byte-identical to the same path there is preserved.
+	// an unclassified entry with a counterpart at the same path there — a
+	// per-worktree copy of a main-checkout local config — is preserved.
 	mainCheckout, _ := strings.CutPrefix(strings.SplitN(string(porcelain), "\n", 2)[0], "worktree ")
 	// An already-removed worktree is success: only its registration is left,
 	// which the prune below clears.
@@ -112,8 +113,12 @@ func removeChangeWorktrees(args []string, env Env, stdout, stderr io.Writer) int
 			case p == "":
 			case rcwRegeneratable(p), rcwCaptureOutput(playwright, p):
 				regen++
-			case wt != mainCheckout && rcwIdentical(wt+"/"+p, mainCheckout+"/"+p):
-				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s — preserved: identical in %s\n", wt, rcwFlat(p), rcwFlat(mainCheckout))
+			case wt != mainCheckout && rcwCounterpart(wt+"/"+p, mainCheckout+"/"+p):
+				how := "differs from"
+				if rcwSameBytes(wt+"/"+p, mainCheckout+"/"+p) {
+					how = "identical in"
+				}
+				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s — preserved: %s %s\n", wt, rcwFlat(p), how, rcwFlat(mainCheckout))
 			default:
 				unpreserved = true
 				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s\n", wt, rcwFlat(p))
@@ -327,17 +332,21 @@ func rcwPlaywrightCheckout(wt string) bool {
 	return len(configs) > 0
 }
 
-// rcwIdentical reports whether a and b are two distinct regular files with
-// the same bytes — check 4's "preserved": the entry still exists in the main
-// checkout. One file reached by both paths (a symlinked directory in the main
-// checkout pointing into the worktree) is not a copy: removing the worktree
-// destroys it.
-func rcwIdentical(a, b string) bool {
+// rcwCounterpart reports whether a and b are two distinct regular files —
+// check 4's "preserved": the entry has a counterpart in the main checkout, a
+// per-worktree copy of a main-checkout local config (kan-999: a worktree's
+// .env with its own ports). One file reached by both paths (a symlinked
+// directory in the main checkout pointing into the worktree) is not a
+// counterpart: removing the worktree destroys it.
+func rcwCounterpart(a, b string) bool {
 	fa, errA := os.Lstat(a)
 	fb, errB := os.Lstat(b)
-	if errA != nil || errB != nil || !fa.Mode().IsRegular() || !fb.Mode().IsRegular() || fa.Size() != fb.Size() || os.SameFile(fa, fb) {
-		return false
-	}
+	return errA == nil && errB == nil && fa.Mode().IsRegular() && fb.Mode().IsRegular() && !os.SameFile(fa, fb)
+}
+
+// rcwSameBytes reports whether a and b hold the same bytes — which of the
+// two preserved labels check 4 prints.
+func rcwSameBytes(a, b string) bool {
 	da, errA := os.ReadFile(a)
 	db, errB := os.ReadFile(b)
 	return errA == nil && errB == nil && bytes.Equal(da, db)
````

  - [ ] **Step 3: Verify.** `cd stats && gofmt -l .` prints nothing;
    `cd stats && go vet ./internal/guard/ && go test ./internal/guard -run TestRemoveChangeWorktrees -count=1`
    passes; `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`,
    `scripts/check-guard-symlinks.sh` and `scripts/check-markdown-integrity.py` exit 0.
  - [ ] **Step 4: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/removechangeworktrees.go`, `stats/internal/guard/remove_change_worktrees_test.go`, `scripts/remove-change-worktrees.sh`, `skills/flow-contracts/finish-contract-run2.md`
**Tests:** `an entry that differs from the main checkout's proceeds`
**Regression:** reverting the commit makes the worktree `.env` that differs from the main checkout's stop cleanup again: `--- FAIL: TestRemoveChangeWorktreesDisclose/an_entry_that_differs_from_the_main_checkout's_proceeds`, exit 3.
**Baseline:** before=23 after=23
<!-- measured: grep -c 't.Run(' stats/internal/guard/remove_change_worktrees_test.go (before at 94818c76 → 23; after with this task's diff applied → 23, the subtest replaced in place) -->
**Commit:** fix(remove-change-worktrees): preserve a worktree's copy of a main-checkout local config
**After:** none

**Decision:** counterpart-in-main-checkout
**Decision:** differs-label-kept-visible
**Decision:** samefile-not-a-counterpart
