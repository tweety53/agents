# pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-19

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-17-pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-19-design.md` in this worktree.

## Global constraints

- In-run pipeline fix: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`. Never merge or push.
- Each task's change is the unified diff its Step 2 gives, applied with `git apply` from this worktree's root — verified there at `bc79a4d6` (`design.md`, **Context**). A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line the diffs add or reword is copied verbatim: this change's `verbatim-moves.txt` acknowledges each such sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Never touch `scripts/kickoff-worktree.sh`, `scripts/integrate-change.sh`, `skills/flow/integrate.md` or `skills/flow-contracts/cross-repo-worktrees.md`: concurrent fix 18 edits them.

## Review focus

- A `.png` in a subdirectory of a Playwright checkout (`screenshots/run-1/01-signin.png` in the test), or at the root of a worktree with no `playwright.config.*`, must still stop: those can be hand-kept references.
- A failure that is no screenshot mismatch — a timeout, a production-only screen — must never be re-baselined.
- The re-baseline rule is stated once, in `skills/flow-contracts/known-bugs.md`; the other two files cite it.

- [x] 1. Trivial residue never stops run 2's cleanup
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Apply the test hunk alone (`git apply --include='stats/internal/guard/remove_change_worktrees_test.go' <the diff in step 2>`) and run `cd stats && go test ./internal/guard -run 'TestRemoveChangeWorktreesDisclose|TestRcwRegeneratable'`: it FAILs `TestRcwRegeneratable` (`.DS_Store` and `sub/.DS_Store` want `true`) and `TestRemoveChangeWorktreesDisclose/a_.DS_Store_and_a_Playwright_checkout's_root_capture_proceed` (exit 3, the cleanup stopped).
  - [x] **Step 2: Implement the fix.** Apply the rest of this diff (`git apply --exclude='stats/internal/guard/remove_change_worktrees_test.go' <the diff>`):

````diff verified:git apply --check in this worktree at bc79a4d6
diff --git a/scripts/remove-change-worktrees.sh b/scripts/remove-change-worktrees.sh
index c73314ea..8b661656 100755
--- a/scripts/remove-change-worktrees.sh
+++ b/scripts/remove-change-worktrees.sh
@@ -17,7 +17,10 @@
 #     untracked, unignored file; no commit that exists only here (the base
 #     resolved fresh for THIS worktree, in-process as resolve-base-branch.sh
 #     does); and the ignored files `--force` will destroy, split by path into
-#     regeneratable (counted) and unclassified (listed).
+#     regeneratable (counted) and unclassified (listed). Regeneratable takes
+#     in a `.DS_Store` anywhere and, in a worktree whose root holds a
+#     `playwright.config.*`, a `.png` at that root — a screenshot spec's
+#     output; a `.png` in any subdirectory stays unclassified.
 #     An unclassified entry that is a regular file byte-identical to the same
 #     path in the repository's main checkout (the first entry of its worktree
 #     list) is preserved — it exists there.
diff --git a/skills/flow-contracts/finish-hand-fallbacks.md b/skills/flow-contracts/finish-hand-fallbacks.md
index a533e579..26ed8aef 100644
--- a/skills/flow-contracts/finish-hand-fallbacks.md
+++ b/skills/flow-contracts/finish-hand-fallbacks.md
@@ -161,13 +161,17 @@ Split what it found into two buckets by path, never by guessing intent:
   or `devStart` recreates with identical content the next time it runs: any path component named
   `build`, `.gradle`, `.kotlin`, `node_modules`, `dist`, `.next`, `target`, `out`, `coverage`,
   `test-results` or `__pycache__`; any `*.log`, `*.pyc` or `*.tsbuildinfo` — Python and TypeScript
-  compiler output a fresh compile or test run recreates; and this pipeline's own `<abs-worktree>/.superpowers/sdd/` and
+  compiler output a fresh compile or test run recreates; any `.DS_Store`, which Finder writes
+  again; and this pipeline's own `<abs-worktree>/.superpowers/sdd/` and
   `<abs-worktree>/.dev-stack/` trees, which a session's own next run writes fresh. A `.png`/`.jpg`
   capture is regeneratable only when it sits under a `test-results` directory (or an equivalent
-  declared screenshot-output directory) a test run owns end to end — never a capture sitting loose at a
-  project root, which could be a hand-saved reference nothing re-creates.
-- **Everything else** — unclassified, and it stays unclassified: no allowlist of *names* is
-  trusted here, because the operator decides what they ignore, and a path this list does not
+  declared screenshot-output directory) a test run owns end to end, or is a `.png` at the root of a
+  Playwright checkout — a worktree whose root holds a `playwright.config.*` — where a screenshot
+  spec writes its output and re-running the spec writes it again — never a capture sitting loose at
+  any other project root, or in a subdirectory of the checkout, which could be a hand-saved
+  reference nothing re-creates.
+- **Everything else** — unclassified, and it stays unclassified: no allowlist of *names* beyond
+  `.DS_Store` is trusted here, because the operator decides what they ignore, and a path this list does not
   recognize (a `.env`, a local override config, a stray file at the worktree root) is exactly what
   the doubt is for.
 
diff --git a/stats/internal/guard/remove_change_worktrees_test.go b/stats/internal/guard/remove_change_worktrees_test.go
index cdfd9fec..96d9d072 100644
--- a/stats/internal/guard/remove_change_worktrees_test.go
+++ b/stats/internal/guard/remove_change_worktrees_test.go
@@ -4,6 +4,8 @@ import (
 	"bytes"
 	"os"
 	"os/exec"
+	"path/filepath"
+	"slices"
 	"strings"
 	"testing"
 	"time"
@@ -376,6 +378,60 @@ func TestRemoveChangeWorktreesDisclose(t *testing.T) {
 		}
 	})
 
+	// pipeline-fix-19: a Finder .DS_Store anywhere and a Playwright
+	// checkout's root screenshot-spec output are regeneratable — counted, never
+	// a disclosure stop; the same root .png outside a Playwright checkout, or
+	// in a kept subdirectory, still stops.
+	t.Run("a .DS_Store and a Playwright checkout's root capture proceed", func(t *testing.T) {
+		t.Parallel()
+		fx := rcwNewFx(t)
+		fx.g.write(fx.repo+"/.git/info/exclude", ".DS_Store\n*.png\nplaywright.config.ts\n")
+		for _, d := range []string{fx.wt, fx.repo} {
+			fx.g.write(d+"/playwright.config.ts", "export default {}")
+		}
+		fx.g.write(fx.wt+"/.DS_Store", "finder")
+		mkdir(t, fx.wt+"/sub")
+		fx.g.write(fx.wt+"/sub/.DS_Store", "finder")
+		fx.g.write(fx.wt+"/01-signin.png", "capture")
+		code, out, errb := fx.run(t, nil)
+		if code != 0 || rcwExists(fx.wt) || len(rcwLines(out, "DISCLOSE: ")) != 0 {
+			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
+		}
+		if l := rcwLines(out, "REGENERATABLE: "); len(l) != 1 || !strings.HasSuffix(l[0], " — 3") {
+			t.Errorf("REGENERATABLE lines %q, want 3", l)
+		}
+	})
+
+	for _, tc := range []struct {
+		label, png string
+		config     bool
+	}{
+		{"a root capture outside a Playwright checkout stops", "01-signin.png", false},
+		{"a Playwright checkout's kept capture in a subdirectory stops", "screenshots/run-1/01-signin.png", true},
+	} {
+		t.Run(tc.label, func(t *testing.T) {
+			t.Parallel()
+			fx := rcwNewFx(t)
+			fx.g.write(fx.repo+"/.git/info/exclude", "*.png\nplaywright.config.ts\n")
+			if tc.config {
+				for _, d := range []string{fx.wt, fx.repo} {
+					fx.g.write(d+"/playwright.config.ts", "export default {}")
+				}
+			}
+			mkdir(t, filepath.Dir(fx.wt+"/"+tc.png))
+			fx.g.write(fx.wt+"/"+tc.png, "capture")
+			code, out, _ := fx.run(t, nil)
+			if code != 3 {
+				t.Fatalf("exit %d, stdout %q; want 3", code, out)
+			}
+			l := slices.DeleteFunc(rcwLines(out, "UNCLASSIFIED: "), func(s string) bool { return strings.Contains(s, " — preserved: ") })
+			if len(l) != 1 || !strings.HasSuffix(l[0], " — "+tc.png) {
+				t.Errorf("UNCLASSIFIED lines %q, want %s unpreserved", l, tc.png)
+			}
+			fx.assertUntouched(t)
+		})
+	}
+
 	t.Run("an entry that differs from the main checkout's stops", func(t *testing.T) {
 		t.Parallel()
 		fx := rcwNewFx(t)
@@ -557,6 +613,8 @@ func TestRcwRegeneratable(t *testing.T) {
 		{"stats/web/tsconfig.tsbuildinfo", true},
 		{"build/out.o", true},
 		{"run.log", true},
+		{".DS_Store", true},
+		{"sub/.DS_Store", true},
 		{".env", false},
 		{".env/secret", false},
 		{"icon.png", false},
diff --git a/stats/internal/guard/removechangeworktrees.go b/stats/internal/guard/removechangeworktrees.go
index a1a50f84..77d042cf 100644
--- a/stats/internal/guard/removechangeworktrees.go
+++ b/stats/internal/guard/removechangeworktrees.go
@@ -6,6 +6,7 @@ import (
 	"io"
 	"os"
 	"os/exec"
+	"path/filepath"
 	"slices"
 	"strings"
 	"time"
@@ -108,7 +109,7 @@ func removeChangeWorktrees(args []string, env Env, stdout, stderr io.Writer) int
 		for _, p := range strings.Split(string(out), "\x00") {
 			switch {
 			case p == "":
-			case rcwRegeneratable(p):
+			case rcwRegeneratable(p), rcwCaptureOutput(wt, p):
 				regen++
 			case wt != mainCheckout && rcwIdentical(wt+"/"+p, mainCheckout+"/"+p):
 				fmt.Fprintf(stdout, "UNCLASSIFIED: %s — %s — preserved: identical in %s\n", wt, rcwFlat(p), rcwFlat(mainCheckout))
@@ -294,10 +295,11 @@ func rcwRestoreUpstream(git func(...string) *exec.Cmd, repo, branch, remote, mer
 // only under a regeneratable location, test-results among them — a capture
 // loose anywhere else matches no rule below. Python (__pycache__, *.pyc) and
 // TypeScript (*.tsbuildinfo) compiler output is regeneratable wherever it
-// sits; everything the list does not name stays unclassified.
+// sits, and so is a Finder .DS_Store; everything the list does not name stays
+// unclassified.
 func rcwRegeneratable(p string) bool {
 	parts := strings.Split(p, "/")
-	if strings.HasPrefix(p, ".superpowers/sdd/") || strings.HasPrefix(p, ".dev-stack/") || strings.HasSuffix(p, ".log") || strings.HasSuffix(p, ".pyc") || strings.HasSuffix(p, ".tsbuildinfo") {
+	if parts[len(parts)-1] == ".DS_Store" || strings.HasPrefix(p, ".superpowers/sdd/") || strings.HasPrefix(p, ".dev-stack/") || strings.HasSuffix(p, ".log") || strings.HasSuffix(p, ".pyc") || strings.HasSuffix(p, ".tsbuildinfo") {
 		return true
 	}
 	for _, c := range parts {
@@ -308,6 +310,19 @@ func rcwRegeneratable(p string) bool {
 	return false
 }
 
+// rcwCaptureOutput is check 4's one worktree-aware regeneratable rule: an
+// ignored .png at the root of a Playwright checkout (a playwright.config.* at
+// the worktree root) is a screenshot spec's output, which a re-run of the
+// spec writes again; a tracked baseline never reaches check 4, and a .png in
+// any subdirectory (a kept screenshots/run-* folder) stays unclassified.
+func rcwCaptureOutput(wt, p string) bool {
+	if strings.Contains(p, "/") || !strings.EqualFold(filepath.Ext(p), ".png") {
+		return false
+	}
+	configs, _ := filepath.Glob(wt + "/playwright.config.*")
+	return len(configs) > 0
+}
+
 // rcwIdentical reports whether a and b are two distinct regular files with
 // the same bytes — check 4's "preserved": the entry still exists in the main
 // checkout. One file reached by both paths (a symlinked directory in the main
````

  - [x] **Step 3: Verify.** `cd stats && go test ./internal/guard -run 'TestRemoveChangeWorktrees|TestRcwRegeneratable'` prints `ok`; `gofmt -l .` prints nothing; `go vet ./internal/guard` exits 0; `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0. Commit.

**Build:** green
**After:** none
**Files:** `stats/internal/guard/removechangeworktrees.go`, `stats/internal/guard/remove_change_worktrees_test.go`, `scripts/remove-change-worktrees.sh`, `skills/flow-contracts/finish-hand-fallbacks.md`
**Allowed-collateral:** `spectre/changes/pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-19/**`
**Tests:** `TestRcwRegeneratable`, `a .DS_Store and a Playwright checkout's root capture proceed`, `a root capture outside a Playwright checkout stops`, `a Playwright checkout's kept capture in a subdirectory stops`
**Regression:** reverting `removechangeworktrees.go` fails `TestRcwRegeneratable` and `a .DS_Store and a Playwright checkout's root capture proceed` — the cleanup stops on `.DS_Store` and root screenshot PNGs again; broadening `rcwCaptureOutput` to every image fails the two `… stops` cases.
**Commit:** fix(guard): never stop run 2 cleanup on .DS_Store or screenshot spec output

**Decision:** playwright-root-png-regeneratable
**Decision:** ds-store-regeneratable

- [x] 2. A stale snapshot baseline on the base is re-baselined, never asked
  - [x] **Step 1: Reproduce before touching anything; do not guess.** `grep -n 're-baseline' skills/flow-contracts/known-bugs.md skills/flow/sync-onto-base.md skills/flow/visual-verify-verifier.md` prints nothing: no contract answers a stale baseline on the base, so the run asks.
  - [x] **Step 2: Implement the fix.** Apply this diff (`git apply <the diff>`):

````diff verified:git apply --check in this worktree at bc79a4d6
diff --git a/skills/flow-contracts/known-bugs.md b/skills/flow-contracts/known-bugs.md
index b12f742a..0a2861ac 100644
--- a/skills/flow-contracts/known-bugs.md
+++ b/skills/flow-contracts/known-bugs.md
@@ -17,6 +17,17 @@ file, never filed for later, and never skipped, worked around or
 tolerance-widened (**Fix determinism at the source, never by widening tolerance**,
 `rules/fix-determinism-at-the-source.mdc`).
 
+**A stale snapshot baseline on the base is re-baselined, never asked.** A screenshot mismatch in a
+spec this change never touched — the spec file and its `*-snapshots/` directory unchanged in
+`git diff <merge-base>..HEAD` — is the base's baselines lagging the running app, wherever it
+surfaces: a `verify` run, or a sync's targeted recapture
+(`skills/flow/sync-onto-base.md`). Whoever ran it recaptures exactly the failing tests — `capture`
+with `<spec>` substituted by each failing test's `<file>:<line>` — stages only the PNGs the failing
+run named, restores every other file the capture rewrote, commits them on the change branch as their
+own `test(fidelity): re-baseline <spec basenames>` commit, and re-runs the failing tests. A failure
+that is no screenshot mismatch — a timeout, a screen only production serves, a missing service — is
+never re-baselined: it is reported, and blocks as above.
+
 **A product defect the self-review pass meets is never dropped.** It stays out of the pass's
 report and filing prompt, and stops the landing. Met inside a run, the pass stops at the end of its
 step 2 — no fix, no prompt, no store row, no report: nothing is committed for the pass, so the
diff --git a/skills/flow/sync-onto-base.md b/skills/flow/sync-onto-base.md
index dd71eb7e..70f47e1b 100644
--- a/skills/flow/sync-onto-base.md
+++ b/skills/flow/sync-onto-base.md
@@ -51,6 +51,10 @@ and `check-base-moved.sh` performs no fetch of its own.
   4. `git add` every recaptured PNG and the zip, stop the stack if step 1 started it, then continue the
      rebase as above.
 
+  A spec outside the named ones failing on a screenshot mismatch in any later run of this sync is a
+  stale baseline on the base: re-baselined per **the sweep**
+  (`skills/flow-contracts/known-bugs.md`), never asked.
+
   **Stop and ask** only where no honest resolution exists: a modify/delete conflict, a binary
   file no `RECAPTURE:` line names, or an upstream commit that removed something this change
   depends on. In that case leave
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 4291b768..f5779fe4 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -22,7 +22,8 @@ never an exit code the command did not return.
    A `specs` with no `<merge-base>` cannot take a round base: it runs as declared, its full list,
    and the report names `specs: not scoped — specs has no <merge-base>`.
    A non-zero `verify` exit blocks, a failing test this change did not introduce included
-   (**the sweep**, `skills/flow-contracts/known-bugs.md`).
+   (**the sweep**, `skills/flow-contracts/known-bugs.md`), save a stale snapshot baseline in a spec
+   this change never touched, which the sweep re-baselines and re-runs before the exit is read.
    **A `verify` that can outrun the shell tool's time limit runs whole, in the background, polled
    to completion** — one Bash call starts it detached,
    `rm -f <exit-file>; nohup sh -c '<verify> > <log> 2>&1; echo $? > <exit-file>' > /dev/null 2>&1 < /dev/null &`
````

  - [x] **Step 3: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh` and `scripts/check-ask-silence.sh` exit 0; `scripts/check-verbatim-moves.sh` reports `0 violation(s)`. Commit.

**Build:** green
**After:** none
**Files:** `skills/flow-contracts/known-bugs.md`, `skills/flow/sync-onto-base.md`, `skills/flow/visual-verify-verifier.md`
**Allowed-collateral:** `spectre/changes/pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-19/**`
**Tests:** **none** — prose; `check-verbatim-moves.sh` pins every added sentence's exact text.
**Regression:** reverting the commit removes the only rule answering a stale baseline on the base: it blocks under **the sweep** and the run asks again.
**Commit:** docs(flow): re-baseline a stale snapshot on the base instead of asking

**Decision:** stale-baseline-rebaselined
**Decision:** rebaseline-by-test-location
