# pipeline-fix-kan-1026-stored-data-first-on-every-open-28

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `remove-change-worktrees.sh` preserves a worktree-root symlink that resolves to the
regular file at the same path in the main checkout, so a worktree's dev-tooling-linked `.env`
never makes run 2's disclosure stop (kan-1026).

**Architecture:** one predicate, `rcwLinkToCounterpart`, and one switch case in check 4 of
`stats/internal/guard/removechangeworktrees.go`; one subtest in `TestRemoveChangeWorktreesDisclose`;
the preservation sentence in `skills/flow-contracts/finish-contract-run2.md` and the header of
`scripts/remove-change-worktrees.sh`.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or other
  trailer; never write `spectre/changes/` (this change's `verbatim-moves.txt` is already committed
  with the plan).
- The task's change is the unified diff Step 2 gives, applied from the repository root with
  `git apply`; it was applied to this worktree at `8e97fb1e` and checked there. A diff that does
  not apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.

## Review Focus

- Only a symlink whose target is the very regular file at the same root path in the main checkout
  is preserved (`link-to-main-regular-file`); a dangling link, a link into the worktree or one
  elsewhere still stops.
- The existing reverse case, `an entry the main checkout reaches through a symlink stops`, still
  passes.

- [ ] 1. Preserve a root symlink to the main checkout's file

**Files:** `stats/internal/guard/removechangeworktrees.go`, `stats/internal/guard/remove_change_worktrees_test.go`, `skills/flow-contracts/finish-contract-run2.md`, `scripts/remove-change-worktrees.sh`
**Tests:** `an entry linked to the main checkout's proceeds`
**Regression:** `an entry linked to the main checkout's proceeds` — reverting the commit drops the `rcwLinkToCounterpart` case, so the linked `.env` is unpreserved and the call exits 3 with the disclosure stop instead of 0.
**Baseline:** before=24 after=25
<!-- measured: grep -cE '\bt\.Run\(' stats/internal/guard/remove_change_worktrees_test.go @ branch pipeline-fix-kan-1026-stored-data-first-on-every-open-28, before and after Step 2's diff -->
**After:** none
**Commit:** `fix(guard): preserve a worktree-root symlink to the main checkout's file`

**Decision:** link-to-main-regular-file
**Decision:** link-is-preserved
**Decision:** link-sentence

**Build:** green

  - [ ] **Step 1: Reproduce before touching anything; do not guess.** The kan-1026 stop is the
    reproducer: Step 2's failing test is it, run against the untouched `removechangeworktrees.go`.
  - [ ] **Step 2: Failing test.** Save the diff below to `.superpowers/sdd/pf28-task1.diff` (the path
    is gitignored), apply its test hunk alone
    (`git apply --include='stats/internal/guard/remove_change_worktrees_test.go' .superpowers/sdd/pf28-task1.diff`) and run
    `cd stats && go test ./internal/guard/ -run '^TestRemoveChangeWorktreesDisclose$' -count=1`:
    FAIL on `an entry linked to the main checkout's proceeds` with exit 3 and an unpreserved
    `UNCLASSIFIED: <wt> — .env` line; every other subtest passes.

````diff verified:applied with git apply to this worktree at 8e97fb1e; TestRemoveChangeWorktrees* passes with it and the new subtest fails (exit 3, .env unpreserved) with removechangeworktrees.go untouched; gofmt -l and go vet ./internal/guard/ print nothing; check-verbatim-moves exits clean with this change's verbatim-moves.txt
diff --git a/scripts/remove-change-worktrees.sh b/scripts/remove-change-worktrees.sh
index 905089e1..c4f1425a 100755
--- a/scripts/remove-change-worktrees.sh
+++ b/scripts/remove-change-worktrees.sh
@@ -25,8 +25,10 @@
 #     with a distinct regular file at the same path in the repository's
 #     main checkout (the first entry of its worktree list) is preserved —
 #     a per-worktree copy of a main-checkout local config, identical or
-#     not; an entry with no such counterpart, or below the root, stays
-#     unpreserved.
+#     not — and so is a symlink there resolving to that regular file
+#     (a link the project's dev tooling made; removal drops the link
+#     alone); an entry with no such counterpart, or below the root,
+#     stays unpreserved.
 #   the disclosure stop — without --proceed, an unpreserved unclassified
 #     entry or any wave-group copy stops here, before check 5 runs anything;
 #     preserved entries alone are shown and the removal proceeds.
@@ -57,6 +59,7 @@
 #   UNCLASSIFIED: <worktree> — <path>            check 4, one per entry
 #   UNCLASSIFIED: <worktree> — <path> — preserved: identical in <main-checkout>
 #   UNCLASSIFIED: <worktree> — <path> — preserved: differs from <main-checkout>
+#   UNCLASSIFIED: <worktree> — <path> — preserved: linked to <main-checkout>
 #   REGENERATABLE: <worktree> — <count>          check 4
 #   SKIPPED: check 5 — ## stop declares no fenced command   a declared `## stop`
 #     whose body carries no fence; check 5 is skipped, never silently
diff --git a/skills/flow-contracts/finish-contract-run2.md b/skills/flow-contracts/finish-contract-run2.md
index 575331c7..16a8493b 100644
--- a/skills/flow-contracts/finish-contract-run2.md
+++ b/skills/flow-contracts/finish-contract-run2.md
@@ -148,7 +148,7 @@ the handoff's **Remote branch:** — deleted, already gone, or not deleted. Exit
 `worktrees`. Exit 2: stop and report. Exit 3: the disclosure stop — nothing removed and nothing
 run; relay its `UNCLASSIFIED:` and `DISCLOSE:` lines per check 4 and the wave-group copies below,
 ask only what they ask, then call again with `--proceed`.
-An unclassified entry at the worktree root that is a regular file with a distinct regular file at the same path in the repository's main checkout — a per-worktree copy of a main-checkout local config such as `.env` — is preserved: its `UNCLASSIFIED:` line ends `— preserved: identical in <main-checkout>`, or `— preserved: differs from <main-checkout>` when the bytes differ, it never makes the disclosure stop, and it is relayed per check 4 whatever the exit; an entry with no counterpart there is unpreserved and still stops.
+An unclassified entry at the worktree root that is a regular file with a distinct regular file at the same path in the repository's main checkout — a per-worktree copy of a main-checkout local config such as `.env` — or a symlink resolving to the regular file at that path, as a project's dev tooling links one in, is preserved: its `UNCLASSIFIED:` line ends `— preserved: identical in <main-checkout>`, `— preserved: differs from <main-checkout>` when the bytes differ, or `— preserved: linked to <main-checkout>` for the symlink, it never makes the disclosure stop, and it is relayed per check 4 whatever the exit; an entry with no counterpart there is unpreserved and still stops.
 
 **Check 6 requires the orchestrating shell's own cwd to be outside every worktree in the resolved
 set before it runs.** `cd` out first, for every worktree, before
diff --git a/stats/internal/guard/remove_change_worktrees_test.go b/stats/internal/guard/remove_change_worktrees_test.go
index d6d18a71..b10ea237 100644
--- a/stats/internal/guard/remove_change_worktrees_test.go
+++ b/stats/internal/guard/remove_change_worktrees_test.go
@@ -378,6 +378,28 @@ func TestRemoveChangeWorktreesDisclose(t *testing.T) {
 		}
 	})
 
+	// kan-1026: a .env the project's dev tooling symlinks to the main
+	// checkout's is preserved — shown, the removal proceeds in the same call,
+	// and the main checkout's file survives it.
+	t.Run("an entry linked to the main checkout's proceeds", func(t *testing.T) {
+		t.Parallel()
+		fx := rcwNewFx(t)
+		fx.g.write(fx.repo+"/.env", "SECRET=1")
+		if err := os.Symlink(fx.repo+"/.env", fx.wt+"/.env"); err != nil {
+			t.Fatal(err)
+		}
+		code, out, errb := fx.run(t, nil)
+		if code != 0 || rcwExists(fx.wt) || len(rcwLines(out, "DISCLOSE: ")) != 0 {
+			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
+		}
+		if l := rcwLines(out, "UNCLASSIFIED: "); len(l) != 1 || !strings.Contains(l[0], " — .env — preserved: linked to ") {
+			t.Errorf("UNCLASSIFIED lines %q, want .env shown as preserved, linked", l)
+		}
+		if !rcwExists(fx.repo + "/.env") {
+			t.Errorf("the main checkout's .env is gone")
+		}
+	})
+
 	// pipeline-fix-19: a Finder .DS_Store anywhere and a Playwright
 	// checkout's root screenshot-spec output are regeneratable — counted, never
 	// a disclosure stop; the same root .png outside a Playwright checkout, or
diff --git a/stats/internal/guard/removechangeworktrees.go b/stats/internal/guard/removechangeworktrees.go
index 842f1b94..52b3c26d 100644
--- a/stats/internal/guard/removechangeworktrees.go
+++ b/stats/internal/guard/removechangeworktrees.go
@@ -59,7 +59,8 @@ func removeChangeWorktrees(args []string, env Env, stdout, stderr io.Writer) int
 	listed, copies := changeWorktrees(porcelain, name)
 	// The main checkout is the porcelain's first entry, whatever <repo> named:
 	// an unclassified entry with a counterpart at the same path there — a
-	// per-worktree copy of a main-checkout local config — is preserved.
+	// per-worktree copy of a main-checkout local config, or a symlink to
+	// it — is preserved.
 	mainCheckout, _ := strings.CutPrefix(strings.SplitN(string(porcelain), "\n", 2)[0], "worktree ")
 	// An already-removed worktree is success: only its registration is left,
 	// which the prune below clears.
@@ -119,6 +120,8 @@ func removeChangeWorktrees(args []string, env Env, stdout, stderr io.Writer) int
 					how = "identical in"
 				}
 				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s — preserved: %s %s\n", wt, rcwFlat(p), how, rcwFlat(mainCheckout))
+			case wt != mainCheckout && !strings.Contains(p, "/") && rcwLinkToCounterpart(wt+"/"+p, mainCheckout+"/"+p):
+				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s — preserved: linked to %s\n", wt, rcwFlat(p), rcwFlat(mainCheckout))
 			default:
 				unpreserved = true
 				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s\n", wt, rcwFlat(p))
@@ -344,6 +347,17 @@ func rcwCounterpart(a, b string) bool {
 	return errA == nil && errB == nil && fa.Mode().IsRegular() && fb.Mode().IsRegular() && !os.SameFile(fa, fb)
 }
 
+// rcwLinkToCounterpart reports whether a is a symlink resolving to b, a
+// regular file — check 4's "preserved" for a main-checkout local config a
+// project's dev tooling links into the worktree (kan-1026: gymie's
+// linkEnvFile .env). Removing the worktree removes the link alone.
+func rcwLinkToCounterpart(a, b string) bool {
+	la, errL := os.Lstat(a)
+	fa, errA := os.Stat(a)
+	fb, errB := os.Lstat(b)
+	return errL == nil && errA == nil && errB == nil && la.Mode()&os.ModeSymlink != 0 && fb.Mode().IsRegular() && os.SameFile(fa, fb)
+}
+
 // rcwSameBytes reports whether a and b hold the same bytes — which of the
 // two preserved labels check 4 prints.
 func rcwSameBytes(a, b string) bool {
````

  - [ ] **Step 3: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/remove_change_worktrees_test.go' .superpowers/sdd/pf28-task1.diff`) and rerun Step 2's test: PASS;
    then `cd stats && go test ./internal/guard/ -run '^TestRemoveChangeWorktrees' -count=1`: ok.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/` prints nothing;
    `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and
    `scripts/check-markdown-integrity.sh` exit 0.
  - [ ] **Step 5: Commit** `stats/internal/guard/removechangeworktrees.go`, `stats/internal/guard/remove_change_worktrees_test.go`,
    `skills/flow-contracts/finish-contract-run2.md` and `scripts/remove-change-worktrees.sh` with the
    subject above.
