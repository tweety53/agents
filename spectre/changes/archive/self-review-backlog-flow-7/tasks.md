# self-review-backlog-flow-7

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the next five live Backlog Flow findings — KAN-771's bundle that cannot tell a
stale daemon from an absent source, KAN-790's fixed departures that leave no probe, KAN-801's
undecided frame conflicts, KAN-804's commit-less serving worktree outside the record, and
KAN-815's satellite change directory run 1 never archives.

**Architecture:** one Go change threading `flowd`'s VCS stamp from `main` through an `api` option
into `selfreview.Bundle`'s notes; three prose additions where their readers already look — **The
loop** and **2. Isolate the workspace**; and one orchestration change in `integrate-change.sh
commit`, with its harness case and run 1's prose.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-19-self-review-backlog-flow-7-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Every task runs in this worktree, `/Users/tweety53/Projects/agents-worktrees/self-review-backlog-flow-7`, branch `self-review-backlog-flow-7`; every `**Files:**` path is relative to its root. Never touch a main checkout.
- Every task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root, checked in order in a scratch clone at `02a3d012`. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Tasks 1 and 5 prove their tests bite: each new test fails with the fix removed, as the task's Step 3 states, and passes with it restored.
- Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container; Task 1 lands in the daemon only when the operator restarts it.

## Review Focus

- Task 1: the build note leads the notes and is absent for an empty build; `formatBuild` reads only `vcs.revision`, `vcs.time` and `vcs.modified`; every other `Bundle` caller passes `""`.
- Task 3's classification runs before the round's task is written and closes a conflict on the handoff's `**Decisions:**` lines without asking.
- Task 5: the satellite loop skips the canonical worktree, runs on an archived re-run, and acts only where `<spec>/changes/<name>/` still exists; the changed `planning_only` expectation is the new archive commit, nothing else.

- [x] 1. Name the serving flowd build in the self-review bundle (KAN-771)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -rn -e ReadBuildInfo -e 'served by flowd' stats/cmd stats/internal; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 02a3d012 → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 02a3d012; make -s web-dist, go vet ./cmd/flowd ./internal/api ./internal/selfreview and gofmt -l . clean; go test ./internal/selfreview ./internal/api -run 'Bundle|SelfReview' and go test ./cmd/flowd -run TestFormatBuild pass; TestBundleNamesTheServingBuild fails with the note's append commented out
diff --git a/stats/cmd/flowd/main.go b/stats/cmd/flowd/main.go
index 25d7b97a..5f031862 100644
--- a/stats/cmd/flowd/main.go
+++ b/stats/cmd/flowd/main.go
@@ -23,6 +23,7 @@ import (
 	"os"
 	"os/exec"
 	"os/signal"
+	"runtime/debug"
 	"strings"
 	"syscall"
 	"time"
@@ -163,7 +164,7 @@ func run(logger *slog.Logger) error {
 	}
 
 	srv, err := api.New(cfg, st, st, st, st, st, logger, api.WithSPA(spaHandler), api.WithGuardRuns(st),
-		api.WithCommitBase(commitBaseFromRemote(originRemote())))
+		api.WithCommitBase(commitBaseFromRemote(originRemote())), api.WithBuild(buildIdentity()))
 	if err != nil {
 		// api.New calls config.Config.Validate itself, which keeps
 		// internal/api correct for any other caller, but the loopback
@@ -466,3 +467,42 @@ func commitBaseFromRemote(remote string) string {
 	}
 	return "https://github.com/" + strings.TrimSuffix(strings.TrimSuffix(repo, "/"), ".git") + "/commit"
 }
+
+// buildIdentity names this binary's build from the VCS stamp `go build`
+// writes into it, for the self-review bundle's build note (KAN-771); ""
+// when the binary carries none.
+func buildIdentity() string {
+	info, ok := debug.ReadBuildInfo()
+	if !ok {
+		return ""
+	}
+	return formatBuild(info.Settings)
+}
+
+// formatBuild renders vcs.revision, vcs.time and vcs.modified as
+// "build <rev12> (<time>)", ", modified" inside the parentheses for a dirty
+// tree; "" with no vcs.revision.
+func formatBuild(settings []debug.BuildSetting) string {
+	var rev, at, modified string
+	for _, s := range settings {
+		switch s.Key {
+		case "vcs.revision":
+			rev = s.Value
+		case "vcs.time":
+			at = s.Value
+		case "vcs.modified":
+			modified = s.Value
+		}
+	}
+	if rev == "" {
+		return ""
+	}
+	if len(rev) > 12 {
+		rev = rev[:12]
+	}
+	out := "build " + rev + " (" + at
+	if modified == "true" {
+		out += ", modified"
+	}
+	return out + ")"
+}
diff --git a/stats/cmd/flowd/wiring_test.go b/stats/cmd/flowd/wiring_test.go
index 43cb1b53..ef4efb13 100644
--- a/stats/cmd/flowd/wiring_test.go
+++ b/stats/cmd/flowd/wiring_test.go
@@ -9,6 +9,7 @@ import (
 	"net"
 	"os"
 	"reflect"
+	"runtime/debug"
 	"strconv"
 	"strings"
 	"testing"
@@ -249,3 +250,19 @@ func TestReplaysSharedJournalOnlyIntoTheFlowDatabase(t *testing.T) {
 		}
 	}
 }
+
+func TestFormatBuild(t *testing.T) {
+	for _, c := range []struct {
+		name     string
+		settings []debug.BuildSetting
+		want     string
+	}{
+		{"clean", []debug.BuildSetting{{Key: "vcs.revision", Value: "02a3d012e12a9b66dc9296d1fbafb249d98ac98e"}, {Key: "vcs.time", Value: "2026-10-18T20:46:11Z"}, {Key: "vcs.modified", Value: "false"}}, "build 02a3d012e12a (2026-10-18T20:46:11Z)"},
+		{"modified", []debug.BuildSetting{{Key: "vcs.revision", Value: "02a3d012e12a9b66dc9296d1fbafb249d98ac98e"}, {Key: "vcs.time", Value: "2026-10-18T20:46:11Z"}, {Key: "vcs.modified", Value: "true"}}, "build 02a3d012e12a (2026-10-18T20:46:11Z, modified)"},
+		{"no stamp", []debug.BuildSetting{{Key: "-compiler", Value: "gc"}}, ""},
+	} {
+		if got := formatBuild(c.settings); got != c.want {
+			t.Errorf("%s: formatBuild = %q, want %q", c.name, got, c.want)
+		}
+	}
+}
diff --git a/stats/internal/api/selfreview.go b/stats/internal/api/selfreview.go
index cec2cd98..c7a8068d 100644
--- a/stats/internal/api/selfreview.go
+++ b/stats/internal/api/selfreview.go
@@ -54,6 +54,7 @@ type selfreviewStore interface {
 type selfreviewHandler struct {
 	store  selfreviewStore
 	git    selfreview.Runner
+	build  string
 	logger *slog.Logger
 }
 
@@ -114,7 +115,7 @@ func (h *selfreviewHandler) bundle(w http.ResponseWriter, r *http.Request) {
 		return
 	}
 
-	bundle, err := selfreview.Bundle(change, rec, summary, summaryFound, panelRan, []string{repo}, h.git)
+	bundle, err := selfreview.Bundle(change, rec, summary, summaryFound, panelRan, h.build, []string{repo}, h.git)
 	if err != nil {
 		writeError(w, http.StatusBadRequest, err.Error())
 		return
diff --git a/stats/internal/api/server.go b/stats/internal/api/server.go
index dae86144..71fa7c14 100644
--- a/stats/internal/api/server.go
+++ b/stats/internal/api/server.go
@@ -176,6 +176,7 @@ type serverOptions struct {
 	spa        http.Handler
 	guardRuns  GuardRunStore
 	commitBase string
+	build      string
 }
 
 // WithSPA mounts h at "/" as the daemon's user interface, task 12's
@@ -202,6 +203,15 @@ func WithCommitBase(base string) Option {
 	}
 }
 
+// WithBuild names the daemon's own build in every self-review bundle it
+// serves -- buildIdentity's answer in production (KAN-771). Passing none,
+// or "", renders no build note.
+func WithBuild(build string) Option {
+	return func(o *serverOptions) {
+		o.build = build
+	}
+}
+
 // apiPathPrefix is the prefix every route this daemon's API owns falls
 // under. It is registered on the mux in its own right (see New, below) so
 // that a request under it which matches none of the specific routes above
@@ -308,7 +318,7 @@ func New(cfg config.Config, cs ChangeStore, ss StageStore, sts StatsStore, rs Re
 	sh := &stageHandler{store: ss, logger: logger}
 	sth := &statsHandler{store: sts, logger: logger, commitBase: so.commitBase}
 	rh := &recordHandler{store: rs, logger: logger}
-	srb := &selfreviewHandler{store: rs, git: selfreview.ExecRunner{}, logger: logger}
+	srb := &selfreviewHandler{store: rs, git: selfreview.ExecRunner{}, build: so.build, logger: logger}
 	lh := &lessonsHandler{store: rs, logger: logger}
 	hz := &hazardHandler{store: rs, logger: logger}
 	suh := &suiteHandler{store: rs, logger: logger}
diff --git a/stats/internal/selfreview/bundle_test.go b/stats/internal/selfreview/bundle_test.go
index aa80de1e..38eb5954 100644
--- a/stats/internal/selfreview/bundle_test.go
+++ b/stats/internal/selfreview/bundle_test.go
@@ -52,7 +52,7 @@ func TestBundleAssemblyRendersStoreSources(t *testing.T) {
 		},
 	}
 
-	bundle, err := Bundle("demo", run, "", false, false, []string{"/repo"}, g)
+	bundle, err := Bundle("demo", run, "", false, false, "", []string{"/repo"}, g)
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -91,7 +91,7 @@ func TestBundleAssemblyRendersStoreSources(t *testing.T) {
 func TestBundleAssemblySkipsAbsentSources(t *testing.T) {
 	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, nil, g)
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", nil, g)
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -122,7 +122,7 @@ func TestBundleAssemblySkipsAbsentSources(t *testing.T) {
 func TestBundleAssemblyNotesUnreadableRepo(t *testing.T) {
 	bogus := filepath.Join(t.TempDir(), "not-a-repo")
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{bogus}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{bogus}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -150,7 +150,7 @@ func TestUnreadableRepoProbeIsGitDir(t *testing.T) {
 		},
 	}
 
-	if _, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, g); err != nil {
+	if _, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, g); err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
 
@@ -169,7 +169,7 @@ func TestBundleAssemblyReadsArchivedFilesThroughGit(t *testing.T) {
 		"narrative.md": "# demo narrative\n",
 	})
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -199,7 +199,7 @@ func TestBundleServesArchivedVisualVerification(t *testing.T) {
 		"visual-verification.md": "# Visual verification — demo\n",
 	})
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -228,7 +228,7 @@ func TestBundleElidesMatrixRowsUnderAnyMatrixHeading(t *testing.T) {
 			"### Departures\n\n| element | departure |\n|---|---|\n| bar | colour |\n",
 	})
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -260,7 +260,7 @@ func TestBundleElidesMatrixRowsAndPNGStatLines(t *testing.T) {
 			"### Notes\n\n```md\n#### FENCED-HEADING\n| FENCED-ROW | kept |\n```\n\n## Motions\n",
 	})
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -292,7 +292,7 @@ func TestBundleAssemblyDerivesFinishCommits(t *testing.T) {
 		"chore(spectre): plan and session records")
 	archiveSHA := writeArchiveBranch(t, repo, "demo", map[string]string{"tasks.md": "# demo tasks\n"})
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -326,7 +326,7 @@ func TestBundleAssemblyGitLogFallsBackToChangeBranch(t *testing.T) {
 	implSHA := commitAll(t, repo, "app.go", "package main\n", "feat(demo): do the thing")
 	runGit(t, repo, "checkout", "main")
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -360,7 +360,7 @@ func TestBundleAssemblyGitLogHonoursRecordedBase(t *testing.T) {
 	runGit(t, repo, "config", "branch.demo.flowBase", "feature")
 	runGit(t, repo, "checkout", "main")
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -382,7 +382,7 @@ func TestBundleAssemblyGitLogStaysAbsentWhenBranchMissing(t *testing.T) {
 	runGit(t, repo, "update-ref", "refs/remotes/origin/main", "main")
 	runGit(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -405,7 +405,7 @@ func TestBundleAssemblyFallsBackToCommittedRecords(t *testing.T) {
 		"panel.md":  "# Review panel — demo\n",
 	})
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -457,7 +457,7 @@ func TestBundleAssemblyPrefersStoreRenderOverCommittedRecord(t *testing.T) {
 		}},
 	}
 
-	bundle, err := Bundle("demo", run, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", run, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -483,7 +483,7 @@ func TestBundleAssemblySkipsCommittedRecordsWhenAbsent(t *testing.T) {
 	repo := gitRepo(t)
 	writeArchiveBranch(t, repo, "demo", map[string]string{"tasks.md": "# demo tasks\n"})
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -630,7 +630,7 @@ func TestBundleAssemblyListsEveryCommitOfArchivedRerun(t *testing.T) {
 		"chore(spectre): plan demo")
 	runGit(t, repo, "checkout", "main")
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -669,7 +669,7 @@ func TestBundleAssemblyReadsArchiveFromBaseAfterMerge(t *testing.T) {
 			runGit(t, repo, "branch", "-D", "spectre/demo")
 			runGit(t, repo, "update-ref", "refs/remotes/origin/main", tc.originAt)
 
-			bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+			bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 			if err != nil {
 				t.Fatalf("Bundle: %v", err)
 			}
@@ -684,7 +684,7 @@ func TestBundleAssemblyReadsArchiveFromBaseAfterMerge(t *testing.T) {
 
 func TestBundleRefusesInvalidName(t *testing.T) {
 	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
-	if _, err := Bundle("../escape", records.Run{}, "", false, false, nil, g); err == nil {
+	if _, err := Bundle("../escape", records.Run{}, "", false, false, "", nil, g); err == nil {
 		t.Fatal("an invalid change name must be refused, not assembled")
 	}
 }
@@ -809,7 +809,7 @@ func TestBundleServesRecordedSummaryFirst(t *testing.T) {
 	run := records.Run{Change: "demo"}
 	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
 
-	bundle, err := Bundle("demo", run, "# Change summary\n\nwhat changed and why", true, false, nil, g)
+	bundle, err := Bundle("demo", run, "# Change summary\n\nwhat changed and why", true, false, "", nil, g)
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -839,7 +839,7 @@ func TestBundleServesRecordedSummaryFirst(t *testing.T) {
 func TestBundleSkipsAbsentSummary(t *testing.T) {
 	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, nil, g)
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", nil, g)
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -858,7 +858,7 @@ func TestBundleSkipsAbsentSummary(t *testing.T) {
 func TestBundleNamesRecordsLossWhenPanelRanWithoutDispatches(t *testing.T) {
 	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, true, nil, g)
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, true, "", nil, g)
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -882,7 +882,7 @@ func TestBundleNamesRecordsLossWhenPanelRanWithoutDispatches(t *testing.T) {
 func TestBundleNoRecordsLossNoteWhenPanelDidNotRun(t *testing.T) {
 	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, nil, g)
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", nil, g)
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -899,7 +899,7 @@ func TestBundleNoRecordsLossNoteWhenDispatchesExist(t *testing.T) {
 	}
 	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
 
-	bundle, err := Bundle("demo", run, "", false, true, nil, g)
+	bundle, err := Bundle("demo", run, "", false, true, "", nil, g)
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -931,7 +931,7 @@ func TestBundleAssemblyLogsEveryMergeOrderPeer(t *testing.T) {
 		"link.md":  "## Merge order\n\n1. `.`\n2. `front`\n",
 	})
 
-	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, ExecRunner{})
 	if err != nil {
 		t.Fatalf("Bundle: %v", err)
 	}
@@ -940,3 +940,23 @@ func TestBundleAssemblyLogsEveryMergeOrderPeer(t *testing.T) {
 		t.Errorf("git log section missing the peer's change commit %s under ### front:\n%s", peerSHA, gitLog)
 	}
 }
+
+func TestBundleNamesTheServingBuild(t *testing.T) {
+	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
+
+	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "build 02a3d012e12a (2026-10-18T20:46:11Z)", nil, g)
+	if err != nil {
+		t.Fatalf("Bundle: %v", err)
+	}
+	if !strings.Contains(bundle, "\nnote: served by flowd build 02a3d012e12a (2026-10-18T20:46:11Z) — ") {
+		t.Errorf("no build note:\n%s", bundle)
+	}
+
+	bundle, err = Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", nil, g)
+	if err != nil {
+		t.Fatalf("Bundle: %v", err)
+	}
+	if strings.Contains(bundle, "served by flowd") {
+		t.Errorf("an empty build still rendered a note:\n%s", bundle)
+	}
+}
diff --git a/stats/internal/selfreview/git.go b/stats/internal/selfreview/git.go
index dea031ce..ea417a41 100644
--- a/stats/internal/selfreview/git.go
+++ b/stats/internal/selfreview/git.go
@@ -538,8 +538,11 @@ var archivedFiles = []string{"tasks.md", "design.md", "narrative.md", "visual-ve
 // line — environmental failure keeps a different wording from legitimate
 // absence. An invalid change name is the one error: the same allowlist
 // records.Destination enforces, checked before the name builds a label or
-// a ref.
-func Bundle(change string, run records.Run, summary string, summaryFound bool, panelRan bool, repos []string, g Runner) (string, error) {
+// a ref. build names the serving daemon's own build, the note leading the
+// bundle's notes when non-empty: a source reported skipped that the tree
+// holds then reads as a daemon older than the source, never as an absence
+// (KAN-771).
+func Bundle(change string, run records.Run, summary string, summaryFound bool, panelRan bool, build string, repos []string, g Runner) (string, error) {
 	if !records.ValidChangeName(change) {
 		return "", fmt.Errorf("change name %q is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'", change)
 	}
@@ -570,7 +573,10 @@ func Bundle(change string, run records.Run, summary string, summaryFound bool, p
 	// still stand. A repository git cannot read at all is named in a note
 	// rather than folded into "absent".
 	var notes []string
-	// The records-source loss note leads the notes: a panel whose stage
+	if build != "" {
+		notes = append(notes, "note: served by flowd "+build+" — a source reported skipped that the tree holds means this daemon predates it, not that it is absent")
+	}
+	// The records-source loss note leads the notes after the build: a panel whose stage
 	// demonstrably completed while the store holds no dispatch rows means
 	// the rows were lost, not skipped, and every "skipped (absent)" line
 	// below would otherwise read as the panel never having run (KAN-621).
````

  - [x] **Step 3: RED, then GREEN.** Comment out the one `notes = append(notes, "note: served by flowd "+build+…)` line in `stats/internal/selfreview/git.go`:
    `cd stats && go test ./internal/selfreview -run TestBundleNamesTheServingBuild` fails. Restore the line: it passes.
  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && make -s web-dist && go vet ./... && gofmt -l .` prints nothing and exits 0;
    `cd stats && go test ./internal/selfreview -run 'TestBundle' && go test ./internal/api -run 'TestSelfReviewBundle' && go test ./cmd/flowd -run 'TestFormatBuild'` passes.
  - [x] **Step 5: Commit** the six files with the subject below.

**Build:** green
**Files:** `stats/cmd/flowd/main.go`, `stats/cmd/flowd/wiring_test.go`, `stats/internal/api/selfreview.go`, `stats/internal/api/server.go`, `stats/internal/selfreview/bundle_test.go`, `stats/internal/selfreview/git.go`
**Tests:** `TestBundleNamesTheServingBuild`, `TestFormatBuild`
**Regression:** reverting the commit drops the `build` parameter and `formatBuild` — both tests stop compiling, and with only the note's append removed `TestBundleNamesTheServingBuild` finds no `served by flowd` line.
**Baseline:** before=36 after=38 <!-- measured: cat stats/internal/selfreview/bundle_test.go stats/cmd/flowd/wiring_test.go | grep -c -E '^func Test[A-Za-z0-9_]+\(' @ 02a3d012 → 36, and 38 after the diff in a scratch clone -->
**Commit:** feat(stats): name the serving flowd build in the self-review bundle
**After:** none

**Decision:** bundle-build-note

- [x] 2. Land a pixel probe for every departure the in-run fix loop fixes (KAN-790)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -e "PIXEL PROBE" -e "permanent probe" skills/flow/verify-fix-loop.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 02a3d012 → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 02a3d012 after the earlier tasks' diffs; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-dispatch-paragraphs.sh exit 0, and check-verbatim-moves.sh 02a3d012 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index ad054459..430ffc77 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -45,10 +45,14 @@ One round runs these steps in order:
 2. **Implement it.** **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) executes the appended
    task exactly as a plan-time task: the decision's implementer pair, `close-task.sh` with its
    guard and its review gate. The implementer's dispatch also carries the MUTATION PROOF
-   paragraph (`skills/flow/review-panel.md`), and the round closes only when every reported
+   and PIXEL PROBE paragraphs (`skills/flow/review-panel.md`), and the round closes only when every reported
    `fix-mutation:` line is checked against the fix diff the way **The fix round mutation-proves
    what it changed** (`skills/flow/review-panel-fix-round.md`) walks them — a mutant nothing
    kills is a new fixable defect this loop takes another round on.
+   **Every mockup departure the round fixes lands a permanent probe in the task's one commit** —
+   a test that drives the real composable or component, never a hand-copied constant, and asserts
+   the value the verifier measured against the frame — so a later change that moves it again fails
+   a test instead of waiting for the next visual verification (KAN-790).
    **A task fixing a test that passes alone and fails only in the full suite closes after two
    more full runs.** Such load flakes surface one per full-suite round, so after the task's commit
    and before `close-task.sh` the parent runs the worktree's whole `## test` list twice more, and
````

  - [x] **Step 3: Verify.** `grep -c -F "permanent probe" skills/flow/verify-fix-loop.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 02a3d012 → 1 -->
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): land a pixel probe for every departure the fix loop fixes
**After:** none

**Decision:** departure-probe

- [x] 3. Classify mockup departures before the fix loop writes its round (KAN-801)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -rn -i "mockup conflict" skills; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 02a3d012 after the earlier tasks' diffs → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 02a3d012 after the earlier tasks' diffs; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-dispatch-paragraphs.sh exit 0, and check-verbatim-moves.sh 02a3d012 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index 430ffc77..fbf2c9a0 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -28,6 +28,13 @@ run included, names under `moved:` in its prompt every frame a `<spec>.mockups`
 spec whose baseline moved — the verifier compares each against its mockup exactly as it does the
 change's own frames, so a screen the shared fix regressed is a departure the next round fixes.
 
+**Mockup departures are classified before the round's task is written.** Where every frame that
+draws the departing component agrees, the departure is a defect this round fixes. Where frames
+disagree on a shared component, the component's own design source wins — the frame that draws it
+as its subject, or the design-system frame it is taken from: the departures that source backs are
+fixed, and the other frame's departure closes as a mockup conflict, named on the handoff's
+`**Decisions:**` lines beside the frame it lost to — never fixed, and never asked (KAN-801).
+
 One round runs these steps in order:
 
 1. **Transcribe the fix.** Run **Documenting a fix, before implementing it**
````

  - [x] **Step 3: Verify.** `grep -c -F "Mockup departures are classified before the round's task is written" skills/flow/verify-fix-loop.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 02a3d012 → 1 -->
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): classify mockup departures before the fix loop writes its round
**After:** Task 2

**Decision:** mockup-conflicts-own-source-wins

- [x] 4. Record a commit-less serving worktree like any other (KAN-804)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n -e "serve from a checkout of its own" -e "will carry no commit" skills/flow/implement.md; echo $?`: it prints only `1`.
    <!-- measured: the grep above in a scratch clone @ 02a3d012 → exit 1, no line -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 02a3d012 after the earlier tasks' diffs; check-references.sh, check-installed-citations.sh, check-markdown-integrity.py and check-dispatch-paragraphs.sh exit 0, and check-verbatim-moves.sh 02a3d012 fails only on sentences verbatim-moves.txt lists
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 6e3c62fa..08f8c759 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -299,9 +299,13 @@ construction on every ordinary run.
 
 **Every `## apps` entry is resolved in this stage too, before any implementation runs.** A later
 stage that finds it needs a worktree in a repository the set does not hold — a fix round's task,
-`flow.visual-verify`'s commit destination — runs that entry's recipe of
+`flow.visual-verify`'s commit destination, a stack that must serve from a checkout of its own —
+runs that entry's recipe of
 `skills/flow/cross-repo-worktrees.md` (the `kickoff-worktree.sh --from` call, the regression
-checkout's toolchain, the merge-order extension), never a bare `git worktree add`. Read
+checkout's toolchain, the merge-order extension), never a bare `git worktree add`. A worktree
+that will carry no commit is no exception: the recipe records it in the state record's
+`worktrees` map with its HEAD as its merge base, so run 2's per-worktree checks and its removal
+reach it like any other (KAN-804). Read
 `<project>/.flow/project.md`'s `## apps` table and resolve each entry per the roots-are-main-checkouts rule of **Where the agents repository is**
 (`skills/flow-contracts/project-configuration.md`): an entry whose
 repository already holds a worktree for this change — the kickoff worktree's repository, a
````

  - [x] **Step 3: Verify.** `grep -c -F "will carry no commit" skills/flow/implement.md` prints `1`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
    <!-- measured: grep -c after git apply in a scratch clone @ 02a3d012 → 1 -->
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` covers its acknowledged sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow): record a commit-less serving worktree like any other
**After:** none

**Decision:** commit-less-worktree-in-map

- [x] 5. Archive every satellite's change directory in run 1 (KAN-815)
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Read `scripts/integrate-change.sh`'s `commit`
    step: its only `spectre archive` and `commit-archive.sh` calls name `$CANON`
    (`grep -n -e 'spectre archive' -e 'commit-archive.sh"' scripts/integrate-change.sh` prints two lines, both on `$CANON`).
    <!-- measured: the grep above in a scratch clone @ 02a3d012 → 2 lines, `(cd "$CANON" && spectre archive …)` and `run "$SCRIPT_DIR/commit-archive.sh" "$CANON" "$NAME"` -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to a scratch clone at 02a3d012 after the earlier tasks' diffs; bash scripts/test-integrate-change.sh → 174 passed, 0 failed, and with only the test hunks applied → 171 passed, 3 failed (satellite_archive's two checks and planning_only's archive-commit check); check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0, and check-verbatim-moves.sh 02a3d012 fails only on sentences verbatim-moves.txt lists
diff --git a/scripts/integrate-change.sh b/scripts/integrate-change.sh
index 5a269e6d..c5e9a1a3 100755
--- a/scripts/integrate-change.sh
+++ b/scripts/integrate-change.sh
@@ -56,7 +56,9 @@
 #          never deletes it), reshape-branch per
 #          worktree, the ledger render, commit-split per
 #          worktree in `## Merge order`, `spectre archive` (each
-#          <name>-fix-N too) unless already archived, commit-archive. A
+#          <name>-fix-N too) unless already archived, commit-archive,
+#          then in every other worktree whose <spec>/changes/<name>/ still
+#          exists `spectre archive <name>` and commit-archive. A
 #          worktree with nothing on the implementation side takes no
 #          subject; the canonical one whose only implementation path is a
 #          link.md takes `chore(spectre): merge order`; any other without
@@ -611,6 +613,18 @@ cmd_commit() {
   mark_begin flow.commit-archive
   run "$SCRIPT_DIR/commit-archive.sh" "$CANON" "$NAME"
   [ "$RC" -eq 0 ] || stop commit-archive "commit-archive exit $RC — report its lines"
+  # Every other worktree archives its own change directory -- a
+  # satellite's link.md -- while it is still there, on an archived re-run
+  # too, and commits it on its own branch (KAN-815).
+  local ws
+  for o in "${ORDER[@]}"; do
+    [ "${WT[o]}" = "$CANON" ] && continue
+    ws="$(spec_root_leaf "${WT[o]}" 2>/dev/null)" || continue
+    [ -d "${WT[o]}/$ws/changes/$NAME" ] || continue
+    (cd "${WT[o]}" && spectre archive "$NAME") || stop archive "spectre archive $NAME refused in ${WT[o]} — report it; never --force"
+    run "$SCRIPT_DIR/commit-archive.sh" "${WT[o]}" "$NAME"
+    [ "$RC" -eq 0 ] || stop commit-archive "commit-archive exit $RC in ${WT[o]} — report its lines"
+  done
   # The rewritten branch reaches the remote now, not at land: a run that
   # stops before land leaves the next plain branch-backup push a fast-forward.
   for o in "${ORDER[@]}"; do
diff --git a/scripts/test-integrate-change.sh b/scripts/test-integrate-change.sh
index d901189b..894f92a2 100755
--- a/scripts/test-integrate-change.sh
+++ b/scripts/test-integrate-change.sh
@@ -63,6 +63,10 @@
 #   planning_only   b's branch changes nothing outside spectre/changes/:
 #                   prepare's NEXT asks no subject for b, and commit with no
 #                   subject for b commits b's planning delta alone.
+#   satellite_archive b's branch carries a satellite's
+#                   spectre/changes/demo/link.md: commit archives it in b
+#                   and commits that as b's own `chore(spectre): archive
+#                   demo` (KAN-815).
 #   link_only_canon a's branch changes nothing outside spectre/changes/ but
 #                   its change directory's link.md, the merge-order record:
 #                   prepare's NEXT asks no subject for a, and commit with no
@@ -493,8 +497,19 @@ check "planning_only: prepare's NEXT asks a subject for a and none for b" \
 printf '\n## integrate run\n\nnarrative\n' >>"$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
 sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work"
 check "planning_only: commit with no subject for b exits 0" '[ "$RC" -eq 0 ]'
-check "planning_only: b's branch is the base plus its one planning commit" \
-  '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s main..HEAD | tr "\n" "|")" = "chore(spectre): plan demo|" ]'
+check "planning_only: b's branch is the base plus its planning and archive commits" \
+  '[ "$(git -C "$WORK/b-worktrees/demo" log --format=%s main..HEAD | tr "\n" "|")" = "chore(spectre): archive demo|chore(spectre): plan demo|" ]'
+
+# --- satellite_archive ---
+fixture
+mkdir -p "$WORK/b-worktrees/demo/spectre/changes/demo"
+printf '## Part of\n\n`a:demo`\n' >"$WORK/b-worktrees/demo/spectre/changes/demo/link.md"
+g -C "$WORK/b-worktrees/demo" add -A && g -C "$WORK/b-worktrees/demo" commit -qm "chore(spectre): link demo"
+to_land "satellite_archive: " no-report
+check "satellite_archive: b's change directory is archived" \
+  '[ ! -d "$WORK/b-worktrees/demo/spectre/changes/demo" ] && [ -f "$WORK/b-worktrees/demo/spectre/changes/archive/demo/link.md" ]'
+check "satellite_archive: b's branch ends on its own archive commit" \
+  '[ "$(git -C "$WORK/b-worktrees/demo" log -1 --format=%s)" = "chore(spectre): archive demo" ]'
 
 # --- link_only_canon ---
 fixture
diff --git a/skills/flow-contracts/finish-contract-run1.md b/skills/flow-contracts/finish-contract-run1.md
index 6817ecba..599272f7 100644
--- a/skills/flow-contracts/finish-contract-run1.md
+++ b/skills/flow-contracts/finish-contract-run1.md
@@ -117,15 +117,15 @@ directory**, `skills/flow-contracts/pipeline.md`).
 **An archived re-run's base** is the newest commit in `<merge-base>..HEAD` whose subject is
 `docs(self-review): <name> self-review report` — or, from a run before KAN-927,
 `docs(self-review): <name> self-review context bundle` — else the newest whose subject is
-`chore(spectre): archive <name>`, else `<merge-base>` itself — a satellite worktree's range holds
-neither — `<merge-base>` being the merge base the reshape below would otherwise take. Everything at
+`chore(spectre): archive <name>`, else `<merge-base>` itself — a satellite worktree whose run 1
+predates KAN-815 holds neither — `<merge-base>` being the merge base the reshape below would otherwise take. Everything at
 or below it is kept untouched — a reshape from that merge base itself would fold the archive and
 report commits into the implementation commit. The re-run **carries new work** when a commit sits
 above the canonical worktree's base — every fix run commits its plan there, whichever repositories
 its tasks touched — or any worktree in the set holds an uncommitted tracked change, judged at the
 unfinished-work gate, before this run's own narrative append.
 
-A re-run of an archived change skips `spectre archive`. **Without new work** it also skips, before
+A re-run of an archived change skips the canonical worktree's `spectre archive`. **Without new work** it also skips, before
 it, the unfinished-work gate, the reshape and the two-commit chain — `spectre archive` already
 refused any unchecked task, and nothing above the base needs committing. **With new work** it
 skips none of them: the unfinished-work gate reads the archived `tasks.md`, and the reshape takes
@@ -298,6 +298,14 @@ the same pull request, as the code:
    files, whose authoring side is the design rule of **The checklist**
    (`skills/flow/brainstorm-planner.md`), and an exit 2 refuses the commit too, never read as a
    pass.
+
+   **Every other worktree of the set archives its own change directory the same way**, after the
+   canonical one and in `## Merge order`: where `<project>/spectre/changes/<name>/` — a satellite's
+   `link.md` — still exists there, `spectre archive <name>` runs in that worktree, on an archived
+   re-run too, since the canonical decision says nothing of a satellite a stopped run left behind;
+   then `commit-archive.sh <that worktree> <name>` commits it as that repository's own
+   `chore(spectre): archive <name>` on `spectre/<name>`. Left unarchived, run 2's
+   `check-cleanup-complete.sh` finds it and reports `LEFTOVER` (KAN-815).
 3. **Push every worktree's `spectre/<name>` `--force-with-lease`**, in `## Merge order`, once the
    archive commit lands — the reshape's push (**Branch backup**, `git-boundaries.md`).
 
````

  - [x] **Step 3: RED, then GREEN.** Revert the script's hunks (`git apply -R --include=scripts/integrate-change.sh <the saved diff>`):
    `bash scripts/test-integrate-change.sh` fails `satellite_archive: b's change directory is archived`,
    `satellite_archive: b's branch ends on its own archive commit` and
    `planning_only: b's branch is the base plus its planning and archive commits`. Re-apply them: it passes.
    <!-- measured: the harness in a scratch clone @ 02a3d012 with the test hunks alone → 171 passed, 3 failed; with the whole diff → 174 passed, 0 failed (about five minutes each) -->
  - [x] **Step 4: Verify.** `bash scripts/test-integrate-change.sh` ends `0 failed`;
    `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `python3 scripts/check-markdown-integrity.py` and `scripts/check-verbatim-moves.sh` exit 0.
  - [x] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** `satellite_archive`
**Regression:** reverting the commit leaves b's `spectre/changes/demo/` in place after `commit` — both `satellite_archive` checks and `planning_only`'s archive-commit check fail.
**Baseline:** before=33 after=34 <!-- measured: grep -c '^# --- ' scripts/test-integrate-change.sh @ 02a3d012 → 33, and 34 after the diff in a scratch clone -->
**Commit:** fix(scripts): archive every satellite's change directory in run 1
**After:** none

**Decision:** satellite-archive-run1
