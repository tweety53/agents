# pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-17

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Spec:** `design.md` beside this file; the brainstorm design is `.superpowers/sdd/2026-10-17-pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-17-design.md` in this worktree.

## Global constraints

- In-run pipeline fix: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`. Never merge or push.
- The change is the unified diff Step 2 gives, applied with `git apply` from this worktree's root — verified there at `2d9dc724` (`design.md`, **Context**). A diff that does not apply is reported, never hand-merged by guesswork.
- Every prose line the diff adds or rewords is copied verbatim: this change's `verbatim-moves.txt` acknowledges each such sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.

## Review focus

- A PNG outside every `*-snapshots/` directory (`logo.png` in the test) must get no `RECAPTURE:` line — it keeps stop and ask.
- The exit stays 1 for a conflict naming recapture paths: `integrate-change.sh` routes only on that exit.
- The recapture serves the change's own app worktrees, already synced onto its base — no throwaway merge, never the conflicting checkout's `$BASE`.

- [ ] 1. A conflicted screenshot baseline is recaptured, not asked
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Apply the test hunk alone (`git apply --include='stats/internal/guard/sync_onto_base_test.go' <the diff in step 2>`) and run `cd stats && go test ./internal/guard -run TestSyncOntoBase`: it FAILs `TestSyncOntoBase/a_conflicted_snapshot_PNG_or_full_app_suite_zip_is_named_for_recapture`, the got stdout ending at the `CONFLICT:` line where the want carries `RECAPTURE: e2e/a.spec.ts-snapshots/x.png` and `RECAPTURE: full-app-suite.zip`.
  - [ ] **Step 2: Implement the fix.** Apply the rest of this diff (`git apply --exclude='stats/internal/guard/sync_onto_base_test.go' <the diff>`):

````diff verified:git apply --check in this worktree at 2d9dc724
diff --git a/scripts/sync-onto-base.sh b/scripts/sync-onto-base.sh
index 73ca3d01..f8c36114 100755
--- a/scripts/sync-onto-base.sh
+++ b/scripts/sync-onto-base.sh
@@ -34,7 +34,11 @@
 #             1, the worktree left mid-rebase exactly as git left it and the
 #             aside still set aside — resolve in place, then --resume. The
 #             aside's stash sha (or none) is recorded in the worktree's own
-#             git dir (`git rev-parse --git-path flow-sync-aside`).
+#             git dir (`git rev-parse --git-path flow-sync-aside`). Each
+#             unmerged path that is a PNG under a `*-snapshots/` directory, or
+#             `full-app-suite.zip`, follows as `RECAPTURE: <path>`, in the
+#             CONFLICT line's order — resolved by a targeted recapture
+#             (skills/flow/sync-onto-base.md), never a stop.
 #
 # --resume runs once an in-place resolution has finished the rebase: it
 # refuses while a rebase is still in progress and unless <onto> — the sha the
diff --git a/skills/flow/sync-onto-base.md b/skills/flow/sync-onto-base.md
index cb075aa8..dd71eb7e 100644
--- a/skills/flow/sync-onto-base.md
+++ b/skills/flow/sync-onto-base.md
@@ -37,8 +37,23 @@ and `check-base-moved.sh` performs no fetch of its own.
   it restores the aside and re-checks, and its exits read as the first call's.
   Never `rebase --abort`, never `rebase --skip`, never `-X ours`/`-X theirs`, and never a
   resolution that drops one side wholesale — a conflict is two changes to the same lines, and both
-  ship. **Stop and ask** only where no honest resolution exists: a modify/delete conflict, a binary
-  file, or an upstream commit that removed something this change depends on. In that case leave
+  ship. **A path the guard names on a `RECAPTURE: <path>` line is resolved by a targeted recapture,
+  never by merging its bytes and never by a stop** — a PNG under a `*-snapshots/` directory, or
+  `full-app-suite.zip`, is a capture of a build, and only a build carrying both sides captures it
+  honestly:
+  1. Serve the change's own app worktrees (already synced onto the change's base) per visual-verify
+     steps 5–6 (`skills/flow/visual-verify.md`, **Steps 3–13**).
+  2. Run only the specs owning the named PNGs — the spec each `*-snapshots/` directory is
+     named for, and `full-app-suite.spec.ts` for the zip — through `capture` with `<spec>`
+     substituted and `--update-snapshots` appended, in the checkout the conflict is in.
+  3. Rebuild `full-app-suite.zip` when it was named, exactly as step 8 of
+     `skills/flow/visual-verify-verifier.md` does.
+  4. `git add` every recaptured PNG and the zip, stop the stack if step 1 started it, then continue the
+     rebase as above.
+
+  **Stop and ask** only where no honest resolution exists: a modify/delete conflict, a binary
+  file no `RECAPTURE:` line names, or an upstream commit that removed something this change
+  depends on. In that case leave
   the worktree mid-rebase exactly as `git rebase` left it, report the file(s) and why, and hand off
   `git -C <worktree> rebase --continue` (after the operator resolves it) or
   `git -C <worktree> rebase --abort` as the next manual step; the run stops before the landing
diff --git a/stats/internal/guard/sync_onto_base_test.go b/stats/internal/guard/sync_onto_base_test.go
index 834a0778..684ea7b9 100644
--- a/stats/internal/guard/sync_onto_base_test.go
+++ b/stats/internal/guard/sync_onto_base_test.go
@@ -128,6 +128,28 @@ func TestSyncOntoBase(t *testing.T) {
 					t.Errorf("stash list %q, want the aside kept", top)
 				}
 			}},
+		{label: "a conflicted snapshot PNG or full app suite zip is named for recapture", fx: func(t *testing.T) *bmFx {
+			fx := rbNewFx(t)
+			paths := []string{"e2e/a.spec.ts-snapshots/x.png", "full-app-suite.zip", "logo.png", "shared.txt"}
+			fx.g.git(fx.repo, "checkout", "-q", "main")
+			mkdir(t, fx.repo+"/e2e/a.spec.ts-snapshots")
+			for _, f := range paths {
+				fx.g.write(fx.repo+"/"+f, "base side")
+			}
+			fx.g.git(fx.repo, "add", "-A")
+			fx.g.git(fx.repo, "commit", "-qm", "base re-baselines")
+			fx.g.git(fx.repo, "checkout", "-q", "demo")
+			mkdir(t, fx.repo+"/e2e/a.spec.ts-snapshots")
+			for _, f := range paths {
+				fx.g.write(fx.repo+"/"+f, "change side")
+			}
+			fx.g.git(fx.repo, "add", "-A")
+			fx.g.git(fx.repo, "commit", "-qm", "demo re-baselines")
+			return fx
+		}, code: 1,
+			out: "MOVED: $WT — 1 commits on main since the recorded merge base; overlaps: e2e/a.spec.ts-snapshots/x.png, full-app-suite.zip, logo.png, shared.txt\n" +
+				"CONFLICT: $WT — onto $TIP; unmerged: e2e/a.spec.ts-snapshots/x.png, full-app-suite.zip, logo.png, shared.txt\n" +
+				"RECAPTURE: e2e/a.spec.ts-snapshots/x.png\nRECAPTURE: full-app-suite.zip\n"},
 		{label: "REFUSE exits 2", fx: rbNewFx, args: func(fx *bmFx) []string { return []string{fx.repo, "main", "-"} }, code: 2,
 			out: "REFUSE: no merge base recorded for $WT — cannot tell whether the base has moved\n"},
 		{label: "a worktree that is not a directory exits 2", fx: rbNewFx, args: func(fx *bmFx) []string { return []string{fx.repo + "/missing", "main", fx.recorded} }, code: 2},
diff --git a/stats/internal/guard/syncontobase.go b/stats/internal/guard/syncontobase.go
index d99b308d..7f3a74a9 100644
--- a/stats/internal/guard/syncontobase.go
+++ b/stats/internal/guard/syncontobase.go
@@ -124,6 +124,11 @@ func syncOntoBase(args []string, env Env, stdout, stderr io.Writer) int {
 				return refuse("cannot record the aside in %s: %v — the rebase is left in progress", rec, err)
 			}
 			fmt.Fprintf(stdout, "CONFLICT: %s — onto %s; unmerged: %s\n", wt, o.sha, strings.Join(o.unmerged, ", "))
+			for _, u := range o.unmerged {
+				if sobRecapture(u) {
+					fmt.Fprintf(stdout, "RECAPTURE: %s\n", u)
+				}
+			}
 			return 1
 		case "refused":
 			return refuse("git refused to rebase %s onto %s — nothing changed", wt, o.sha)
@@ -155,3 +160,22 @@ func syncOntoBase(args []string, env Env, stdout, stderr io.Writer) int {
 	fmt.Fprintf(stdout, "REBASED: %s — onto %s\n", wt, onto)
 	return 0
 }
+
+// sobRecapture reports whether an unmerged path is a captured screenshot
+// baseline -- a PNG under a Playwright `*-snapshots/` directory -- or the full
+// app suite's archive: a conflict the caller resolves by a targeted recapture,
+// never by merging bytes or stopping to ask.
+func sobRecapture(p string) bool {
+	if filepath.Base(p) == "full-app-suite.zip" {
+		return true
+	}
+	if !strings.EqualFold(filepath.Ext(p), ".png") {
+		return false
+	}
+	for _, dir := range strings.Split(filepath.Dir(p), "/") {
+		if strings.HasSuffix(dir, "-snapshots") {
+			return true
+		}
+	}
+	return false
+}
````

  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard -run TestSyncOntoBase` prints `ok`; `gofmt -l .` prints nothing; `go vet ./internal/guard` exits 0; `scripts/check-verbatim-moves.sh` reports `0 violation(s)`; `scripts/check-references.sh` and `scripts/check-installed-citations.sh` exit 0. Commit.

**Build:** green
**After:** none
**Files:** `stats/internal/guard/syncontobase.go`, `stats/internal/guard/sync_onto_base_test.go`, `scripts/sync-onto-base.sh`, `skills/flow/sync-onto-base.md`
**Allowed-collateral:** `spectre/changes/pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-17/**`
**Tests:** `stats/internal/guard/sync_onto_base_test.go` — case `a conflicted snapshot PNG or full app suite zip is named for recapture`.
**Regression:** reverting `syncontobase.go` drops the two `RECAPTURE:` lines and the new case fails; the Conflict bullet then has no line to act on and a snapshot conflict stops and asks again.
**Commit:** fix(guard): recapture a conflicted screenshot baseline instead of asking

**Decision:** guard-names-recapture-paths
