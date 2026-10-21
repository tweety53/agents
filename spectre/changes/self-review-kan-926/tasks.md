# self-review-kan-926

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close KAN-926 — a main checkout holding uncommitted copies of drift already landed on
`origin/<base>` is fast-forwarded by run 2's refresh instead of left dirty for the operator.

**Architecture:** one change to the existing `refresh-main-checkout` Go guard: after its existing
refusals, a dirty checkout whose every changed path is byte-identical to `origin/<base>` has those
paths restored to `HEAD` and is fast-forwarded; its script header states the rule.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-926-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/`. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-926`, branch `self-review-kan-926`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- The task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root; it was applied and checked in this worktree at `b76c2cf1`. Save it by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- A path whose worktree copy matches `origin/<base>` but whose staged index copy differs is refused, never restored — the staged bytes would be lost.
- A dirty checkout already at `origin/<base>`, on another branch, detached, or with local commits keeps its existing refusal, and nothing is restored before every refusal has been checked.
- Paths with spaces or pathspec magic characters (`*`, `:`) reach `git diff`/`git restore` literally (`--literal-pathspecs`, `-z`).

- [ ] 1. Fast-forward a main checkout over copies identical to the landed drift
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Save Step 2's diff, apply its
    test hunk alone (`git apply --include='stats/internal/guard/refresh_main_checkout_test.go' <the saved diff>`)
    and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestRefreshMainCheckoutDiscardsLandedCopies' -count=1`:
    the `copies identical to origin are discarded, then fast-forwarded` subtest fails with
    `REFRESH-REFUSED: <repo> has tracked changes — nothing touched`.
    <!-- measured: the test hunk alone, go test -run TestRefreshMainCheckoutDiscardsLandedCopies @ b76c2cf1 → FAIL, that subtest refused with has tracked changes -->
  - [ ] **Step 2: Apply the fix** — the rest (`git apply --exclude='stats/internal/guard/refresh_main_checkout_test.go' <the saved diff>`):

````diff verified:applied with git apply to this worktree at b76c2cf1; go test ./internal/guard/ -count=1 passes, the new test fails with refreshmaincheckout.go reverted, go vet and gofmt -l clean
diff --git a/scripts/refresh-main-checkout.sh b/scripts/refresh-main-checkout.sh
index f1675d4c..a3cebe1e 100755
--- a/scripts/refresh-main-checkout.sh
+++ b/scripts/refresh-main-checkout.sh
@@ -15,17 +15,31 @@
 # In order: the bounded, credential-free fetch resolve-base-branch.sh runs
 # (`-c core.askpass=true fetch --quiet origin`, its failure ignored — a
 # stale origin/<base> only means a smaller step); then every refusal below,
-# checked before anything moves; then `git merge --ff-only origin/<base>`.
+# checked before anything moves; then the landed-copies discard; then
+# `git merge --ff-only origin/<base>`.
+#
+# LANDED COPIES (KAN-926). Drift landed on origin/<base> from another
+# worktree leaves the main checkout holding uncommitted copies of exactly
+# what the fast-forward brings, and git's merge refuses to overwrite them.
+# When every tracked change is such a copy — each changed path's worktree
+# content, and its index content wherever staged, byte-identical to
+# origin/<base>, with no unmerged entry — those paths alone are restored to
+# HEAD (`git restore --source=HEAD --staged --worktree -- <paths>`) and the
+# fast-forward writes the same bytes back: nothing is lost. Any path that
+# differs refuses the whole set, every copy kept.
 #
 # Exit codes:
-#   0  REFRESH-DONE: <path> fast-forwarded <old> -> <new>
+#   0  REFRESH-DISCARDED: <path> <n> tracked paths identical to origin/<base>
+#      — before REFRESH-DONE, when landed copies were restored
+#      REFRESH-DONE: <path> fast-forwarded <old> -> <new>
 #      REFRESH-CURRENT: <path> already at <tip> (nothing to do)
 #      FLOWD-RESTARTED: <path>/stats — after REFRESH-DONE, when <old>..<new>
 #      moved a stats/ path and <path>/stats/Makefile and stats/cmd/flowd exist: `make -C
 #      <path>/stats restart` ran
 #   1  REFRESH-REFUSED: <path> <reason> — nothing touched. Reasons: is
 #      detached; is on <branch>, not <base>; has tracked changes (any
-#      modified, staged or unmerged entry); has no origin/<base>; has <base>
+#      modified, staged or unmerged entry not a landed copy above); has no
+#      origin/<base>; has <base>
 #      commits origin/<base> lacks (local <base> is not an ancestor of
 #      origin/<base>)
 #   2  usage; or cannot answer — a git read or the merge fails outright
diff --git a/stats/internal/guard/refresh_main_checkout_test.go b/stats/internal/guard/refresh_main_checkout_test.go
index 68157c44..2a08e00d 100644
--- a/stats/internal/guard/refresh_main_checkout_test.go
+++ b/stats/internal/guard/refresh_main_checkout_test.go
@@ -221,3 +221,52 @@ func TestRefreshMainCheckoutRestartsFlowdWhenStatsMoved(t *testing.T) {
 		})
 	}
 }
+
+// KAN-926: drift landed on origin from another worktree leaves the main
+// checkout holding identical uncommitted copies -- an unstaged edit, a
+// staged addition -- which git's own fast-forward refuses to overwrite.
+func TestRefreshMainCheckoutDiscardsLandedCopies(t *testing.T) {
+	t.Parallel()
+	env := Env{Getenv: os.Getenv}
+
+	t.Run("copies identical to origin are discarded, then fast-forwarded", func(t *testing.T) {
+		t.Parallel()
+		repo, _ := behindRepo(t)
+		old := rmcRev(t, repo, "HEAD")
+		writeFile(t, repo+"/f.txt", "b\n")
+		writeFile(t, repo+"/g.txt", "new\n")
+		gitRun(t, repo, "add", "g.txt")
+		r := runGuard("refresh-main-checkout", []string{repo, "main"}, env)
+		tip := rmcRev(t, repo, "origin/main")
+		want := "REFRESH-DISCARDED: " + repo + " 2 tracked paths identical to origin/main\n" +
+			"REFRESH-DONE: " + repo + " fast-forwarded " + old + " -> " + tip + "\n"
+		if r.rc != 0 || r.stdout != want {
+			t.Fatalf("want %q; got %+v", want, r)
+		}
+		var g fxGit
+		if s := g.git(repo, "status", "--porcelain"); g.err != nil || s != "" {
+			t.Fatalf("checkout not clean: %q %v", s, g.err)
+		}
+	})
+
+	// One path matching origin does not carry another that does not: the
+	// whole set is refused and every copy kept.
+	t.Run("a copy differing from origin refuses the whole set", func(t *testing.T) {
+		t.Parallel()
+		repo, _ := behindRepo(t)
+		before := rmcRev(t, repo, "main")
+		writeFile(t, repo+"/f.txt", "b\n")
+		writeFile(t, repo+"/g.txt", "mine\n")
+		gitRun(t, repo, "add", "g.txt")
+		r := runGuard("refresh-main-checkout", []string{repo, "main"}, env)
+		if want := "REFRESH-REFUSED: " + repo + " has tracked changes — nothing touched\n"; r.rc != 1 || r.stdout != want {
+			t.Fatalf("want exit 1, stdout %q; got %+v", want, r)
+		}
+		if got := rmcRev(t, repo, "main"); got != before {
+			t.Fatalf("local main moved: %s -> %s", before, got)
+		}
+		if body, _ := os.ReadFile(repo + "/f.txt"); string(body) != "b\n" {
+			t.Fatal("identical copy discarded beside a differing one")
+		}
+	})
+}
diff --git a/stats/internal/guard/refreshmaincheckout.go b/stats/internal/guard/refreshmaincheckout.go
index 173042ea..2f35f7e5 100644
--- a/stats/internal/guard/refreshmaincheckout.go
+++ b/stats/internal/guard/refreshmaincheckout.go
@@ -3,6 +3,9 @@ package guard
 import (
 	"fmt"
 	"io"
+	"os/exec"
+	"slices"
+	"strings"
 )
 
 // refreshMainCheckout is scripts/refresh-main-checkout.sh: that script's
@@ -45,14 +48,12 @@ func refreshMainCheckout(args []string, env Env, stdout, stderr io.Writer) int {
 		return refuse("is on " + cur + ", not " + base)
 	}
 	// Any tracked entry -- modified, staged or unmerged -- is work the
-	// fast-forward could collide with; untracked files are git's own concern.
+	// fast-forward could collide with, unless it is a landed copy (below);
+	// untracked files are git's own concern.
 	status, ok := g("status", "--porcelain", "--untracked-files=no")
 	if !ok {
 		return 2
 	}
-	if status != "" {
-		return refuse("has tracked changes")
-	}
 	tip, ok := g("rev-parse", "--short", "HEAD")
 	if !ok {
 		return 2
@@ -62,12 +63,25 @@ func refreshMainCheckout(args []string, env Env, stdout, stderr io.Writer) int {
 		return refuse("has no origin/" + base)
 	}
 	if tip == remote {
+		if status != "" {
+			return refuse("has tracked changes")
+		}
 		fmt.Fprintf(stdout, "REFRESH-CURRENT: %s already at %s\n", repo, tip)
 		return 0
 	}
 	if err := git("-C", repo, "merge-base", "--is-ancestor", "HEAD", "refs/remotes/origin/"+base).Run(); err != nil {
 		return refuse("has " + base + " commits origin/" + base + " lacks")
 	}
+	if status != "" {
+		paths := landedCopies(git, repo, "refs/remotes/origin/"+base)
+		if paths == nil {
+			return refuse("has tracked changes")
+		}
+		if _, ok := g(append([]string{"--literal-pathspecs", "restore", "--source=HEAD", "--staged", "--worktree", "--"}, paths...)...); !ok {
+			return 2
+		}
+		fmt.Fprintf(stdout, "REFRESH-DISCARDED: %s %d tracked paths identical to origin/%s\n", repo, len(paths), base)
+	}
 	if _, ok := g("merge", "--ff-only", "-q", "refs/remotes/origin/"+base); !ok {
 		return 2
 	}
@@ -77,3 +91,48 @@ func refreshMainCheckout(args []string, env Env, stdout, stderr io.Writer) int {
 	}
 	return 0
 }
+
+// landedCopies returns the tracked paths a dirty main checkout holds when
+// every one is byte-identical to ref -- its worktree copy, and its index
+// copy wherever one is staged -- the copies a change landed from another
+// worktree leaves behind (KAN-926); nil when any path differs, any entry is
+// unmerged, or a read fails. An unstaged path's index copy is HEAD's, already
+// committed, so restoring exactly these paths to HEAD loses nothing the
+// fast-forward does not write back.
+func landedCopies(git func(args ...string) *exec.Cmd, repo, ref string) []string {
+	if u, ok := capture(git("-C", repo, "ls-files", "--unmerged")); !ok || u != "" {
+		return nil
+	}
+	names := func(cached ...string) ([]string, bool) {
+		args := append([]string{"-C", repo, "diff", "--name-only", "-z", "--no-renames"}, cached...)
+		out, ok := capture(git(append(args, "HEAD")...))
+		var ps []string
+		for _, p := range strings.Split(out, "\x00") {
+			if p != "" {
+				ps = append(ps, p)
+			}
+		}
+		return ps, ok
+	}
+	// same is `git diff --quiet ref -- paths`: worktree, or with --cached
+	// the index, byte-identical to ref at every path.
+	same := func(paths []string, cached ...string) bool {
+		args := append([]string{"-C", repo, "--literal-pathspecs", "diff", "--quiet"}, cached...)
+		return len(paths) == 0 || git(append(append(args, ref, "--"), paths...)...).Run() == nil
+	}
+	worktree, ok := names()
+	staged, ok2 := names("--cached")
+	if !ok || !ok2 || !same(staged, "--cached") {
+		return nil
+	}
+	paths := append([]string{}, worktree...)
+	for _, p := range staged {
+		if !slices.Contains(paths, p) {
+			paths = append(paths, p)
+		}
+	}
+	if len(paths) == 0 || !same(paths) {
+		return nil
+	}
+	return paths
+}
````

  - [ ] **Step 3: Verify.** Run `cd stats && gofmt -l ./internal/guard && make -s web-dist && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestRefreshMainCheckout' -count=1`
    — `gofmt` prints nothing and the tests pass — then `scripts/check-references.sh` and
    `scripts/check-guard-symlinks.sh`, both exit 0.
  - [ ] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/refreshmaincheckout.go`, `stats/internal/guard/refresh_main_checkout_test.go`, `scripts/refresh-main-checkout.sh`
**Tests:** `TestRefreshMainCheckoutDiscardsLandedCopies`
**Regression:** `TestRefreshMainCheckoutDiscardsLandedCopies` — reverting the commit refuses a checkout whose only changes are copies identical to `origin/main` with `has tracked changes`, and the test fails on its expected `REFRESH-DISCARDED`/`REFRESH-DONE` output.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/refresh_main_checkout_test.go @ b76c2cf1 -->
**Commit:** fix(guard): fast-forward a main checkout over copies identical to the landed drift
**After:** none

**Decision:** discard-in-refresh
**Decision:** identity-check
**Decision:** restore-to-head
**Decision:** discarded-line
