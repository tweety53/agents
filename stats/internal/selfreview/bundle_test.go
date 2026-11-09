package selfreview

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// fakeGit answers every Output call with canned stdout, recording the
// invocations; calls whose repo is in failOn report an error instead, the
// way a missing ref, a missing path or an absent repository fails git.
type fakeGit struct {
	responses map[string]string
	failOn    map[string]bool
	calls     []string
}

func (f *fakeGit) Output(repo string, args ...string) ([]byte, error) {
	invoked := repo + " " + strings.Join(args, " ")
	f.calls = append(f.calls, invoked)
	if f.failOn[invoked] {
		return nil, errors.New("git failed: " + invoked)
	}
	if out, ok := f.responses[invoked]; ok {
		return []byte(out), nil
	}
	return nil, errors.New("git has no canned answer for: " + invoked)
}

func TestBundleAssemblyRendersStoreSources(t *testing.T) {
	run := records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{{
			Seq: 1, Role: "implementer", Model: "sonnet", SessionToken: "mf-demo",
		}},
	}
	repo := gitRepo(t)
	writeChangeDir(t, repo, "demo", map[string]string{"tasks.md": "# demo tasks\n"})

	bundle, err := Bundle("demo", run, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	for _, want := range []string{
		"# Self-review context bundle for demo",
		"## .superpowers/sdd/ledgers/demo.md",
		"## .superpowers/sdd/reviews/demo-panel.md",
		"## spectre/changes/demo/tasks.md",
		"# demo tasks",
	} {
		if !strings.Contains(bundle, want) {
			t.Errorf("bundle missing %q:\n%s", want, bundle)
		}
	}
	if strings.Contains(bundle, "skipped: .superpowers/sdd/ledgers/demo.md") {
		t.Errorf("ledger rendered but reported skipped:\n%s", bundle)
	}
	for _, label := range []string{
		"spectre/changes/demo/design.md",
		"spectre/changes/demo/narrative.md",
		"spectre/changes/demo/visual-verification.md",
	} {
		if !strings.Contains(bundle, "skipped: "+label+" (absent)") {
			t.Errorf("absent source %q not reported skipped:\n%s", label, bundle)
		}
	}
	// The labels name the change directory repo-relatively: no absolute
	// repository path may appear anywhere in the bundle.
	if strings.Contains(bundle, repo) {
		t.Errorf("bundle leaks a repository path:\n%s", bundle)
	}
}

func TestBundleAssemblySkipsAbsentSources(t *testing.T) {
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", nil, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	// A change the store has never heard of renders nothing: the ledger
	// needs dispatch rows and the panel needs any row at all.
	if !strings.Contains(bundle, "found: 0 of 10 sources; skipped: 10 of 10 sources") {
		t.Errorf("summary line wrong:\n%s", bundle)
	}
	if strings.Contains(bundle, "## .superpowers/sdd/ledgers/demo.md") {
		t.Errorf("no dispatch rows, yet a ledger section rendered:\n%s", bundle)
	}
	if strings.Contains(bundle, "## .superpowers/sdd/reviews/demo-panel.md") {
		t.Errorf("no rows at all, yet a panel section rendered:\n%s", bundle)
	}
	if !strings.Contains(bundle, "skipped: gated per-task reviewer findings (absent)") {
		t.Errorf("no gated rows, yet the gated source is not reported absent:\n%s", bundle)
	}
	if !strings.Contains(bundle, "skipped: git log --stat (absent)") {
		t.Errorf("no repository recorded, yet git log resolved:\n%s", bundle)
	}
	if strings.Contains(bundle, "note: repository") {
		t.Errorf("no repository was supplied, yet an unreadable note appeared:\n%s", bundle)
	}
}

// TestBundleAssemblyNotesUnreadableRepo pins the refused-versus-absent
// split: a supplied repository git cannot read at all is named in a note
// line, never folded into the same "skipped (absent)" a change that was
// simply has no directory on disk wears.
func TestBundleAssemblyNotesUnreadableRepo(t *testing.T) {
	bogus := filepath.Join(t.TempDir(), "not-a-repo")

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{bogus}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if !strings.Contains(bundle, "note: repository "+bogus+" could not be read") {
		t.Errorf("unreadable repository not noted:\n%s", bundle)
	}
	if !strings.Contains(bundle, "skipped: spectre/changes/demo/tasks.md (absent)") {
		t.Errorf("unreadable repository's sources not skipped:\n%s", bundle)
	}
}

// TestUnreadableRepoProbeIsGitDir pins the health probe's command shape:
// `rev-parse --git-dir` answers for any repository, commit-less ones
// included — a mutant probing a ref instead would stamp a false note on a
// valid repository that simply has no commits yet.
func TestUnreadableRepoProbeIsGitDir(t *testing.T) {
	repo := gitRepo(t)
	g := &fakeGit{
		responses: map[string]string{
			repo + " rev-parse --git-dir": repo + "/.git",
		},
		failOn: map[string]bool{},
	}

	if _, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, g); err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	for _, call := range g.calls {
		if strings.HasPrefix(call, repo+" rev-parse") && call != repo+" rev-parse --git-dir" {
			t.Errorf("health probe issued %q, want rev-parse --git-dir", call)
		}
	}
}

func TestBundleElidesMatrixRowsUnderAnyMatrixHeading(t *testing.T) {
	repo := gitRepo(t)
	commitAll(t, repo, "a.go", "package a\n", "feat(demo): a")
	writeChangeDir(t, repo, "demo", map[string]string{
		"tasks.md": "# demo tasks\n",
		"visual-verification.md": "# Visual verification — demo\n\n## Element × property matrices\n\n" +
			"### cardio-add-workout:cardio-add-workout-A1 matrix (A1)\n\n| element | x |\n|---|---|\n| ROW-H3 | 1 |\n\n" +
			"## Matrix\n\n| element | x |\n|---|---|\n| ROW-H2 | 1 |\n\n" +
			"### Departures\n\n| element | departure |\n|---|---|\n| bar | colour |\n",
	})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	for _, want := range []string{"### cardio-add-workout:cardio-add-workout-A1 matrix (A1)", "## Matrix", "_(3 matrix rows elided)_", "| bar | colour |"} {
		if !strings.Contains(bundle, want) {
			t.Errorf("bundle missing %q:\n%s", want, bundle)
		}
	}
	for _, gone := range []string{"ROW-H3", "ROW-H2"} {
		if strings.Contains(bundle, gone) {
			t.Errorf("bundle still carries matrix row %q:\n%s", gone, bundle)
		}
	}
}

func TestBundleElidesBandsSeamsAndInkSections(t *testing.T) {
	repo := gitRepo(t)
	commitAll(t, repo, "a.go", "package a\n", "feat(demo): a")
	writeChangeDir(t, repo, "demo", map[string]string{
		"tasks.md": "# demo tasks\n",
		"visual-verification.md": "# Visual verification — demo\n\nHEADER-LINE\n\n## Verify\n\nVERIFY-LINE\n\n" +
			"## Views\n\n- q2: departure — VERDICT-LINE\n\n" +
			"## Bands and seams (raw summaries)\n\n### q2-answer\n\nBAND-LINE\n\n" +
			"## Ink (runs through each band)\n\nINK-LINE\n\n## Motions\n\nMOTION-LINE\n\n### Q1 sweeps (rows, ink)\n\nSWEEP-LINE\n",
	})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	label := "spectre/changes/demo/visual-verification.md"
	for _, want := range []string{"HEADER-LINE", "VERIFY-LINE", "VERDICT-LINE", "MOTION-LINE", "SWEEP-LINE",
		"_(6 lines elided: Bands and seams (raw summaries) — in " + label + ")_",
		"_(4 lines elided: Ink (runs through each band) — in " + label + ")_"} {
		if !strings.Contains(bundle, want) {
			t.Errorf("bundle missing %q:\n%s", want, bundle)
		}
	}
	for _, gone := range []string{"BAND-LINE", "INK-LINE", "### q2-answer"} {
		if strings.Contains(bundle, gone) {
			t.Errorf("bundle still carries measurement line %q:\n%s", gone, bundle)
		}
	}
}

func TestBundleElidesMatrixRowsAndPNGStatLines(t *testing.T) {
	repo := gitRepo(t)
	onChangeBranch(t, repo, "demo")
	if err := os.WriteFile(filepath.Join(repo, "shot.png"), []byte("\x89PNG"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, repo, "a.go", "package a\n", "feat(demo): a with a screenshot")
	writeChangeDir(t, repo, "demo", map[string]string{
		"tasks.md": "# demo tasks\n",
		"visual-verification.md": "# Visual verification — demo\n\n### Departures\n\n" +
			"| element | departure |\n|---|---|\n| bar | colour |\n\n" +
			"#### cardio-R1 matrix\n\n| prop | ok |\n|---|---|\n| ROW-ONE | yes |\n```sh\n# FENCED-COMMENT\n```\n| ROW-TWO | yes |\n\n" +
			"### Notes\n\n```md\n#### FENCED-HEADING\n| FENCED-ROW | kept |\n```\n\n## Motions\n",
	})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	for _, want := range []string{"| bar | colour |", "#### cardio-R1 matrix", "_(3 matrix rows elided)_", "_(1 matrix rows elided)_", "# FENCED-COMMENT", "| FENCED-ROW | kept |", "## Motions"} {
		if !strings.Contains(bundle, want) {
			t.Errorf("bundle missing %q:\n%s", want, bundle)
		}
	}
	for _, gone := range []string{"ROW-ONE", "ROW-TWO", "| prop | ok |"} {
		if strings.Contains(bundle, gone) {
			t.Errorf("bundle still carries matrix row %q:\n%s", gone, bundle)
		}
	}
	gitLog := bundle[strings.Index(bundle, "## git log --stat"):]
	if !strings.Contains(gitLog, "a.go") || strings.Contains(gitLog, "shot.png") {
		t.Errorf("git log --stat must name a.go and not shot.png:\n%s", gitLog)
	}
}

// TestBundleAssemblyGitLogFallsBackToChangeBranch pins the KAN-583
// fallback: a change with no change directory — a /flow-fast change — still
// gets its git-log source, derived from the main checkout the caller
// passes as the change branch's own commits against the remote default
// branch. The branch's commits exist and there is no change directory;
// absent must not be the answer.
func TestBundleAssemblyGitLogFallsBackToChangeBranch(t *testing.T) {
	repo := gitRepo(t)
	runGit(t, repo, "update-ref", "refs/remotes/origin/main", "main")
	runGit(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	runGit(t, repo, "checkout", "-b", "demo")
	implSHA := commitAll(t, repo, "app.go", "package main\n", "feat(demo): do the thing")
	runGit(t, repo, "checkout", "main")

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if strings.Contains(bundle, "skipped: git log --stat") {
		t.Fatalf("git log source resolved from the change branch but reported skipped:\n%s", bundle)
	}
	gitLog := bundle[strings.Index(bundle, "## git log --stat"):]
	if !strings.Contains(gitLog, "commit "+implSHA) {
		t.Errorf("git log section missing the change branch's commit %s:\n%s", implSHA, gitLog)
	}
	// One found source — the git log; every other source a flow-fast run
	// has no change directory to read and no store rows to render.
	if !strings.Contains(bundle, "found: 1 of 10 sources; skipped: 9 of 10 sources") {
		t.Errorf("summary line wrong:\n%s", bundle)
	}
}

// TestBundleAssemblyGitLogHonoursRecordedBase pins that a change cut from
// a recorded base (branch.<change>.flowBase) logs against that base, so the
// base branch's own commits never read as the change's.
func TestBundleAssemblyGitLogHonoursRecordedBase(t *testing.T) {
	repo := gitRepo(t)
	runGit(t, repo, "update-ref", "refs/remotes/origin/main", "main")
	runGit(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	runGit(t, repo, "checkout", "-b", "feature")
	featureSHA := commitAll(t, repo, "feature.go", "package main\n", "feat(feature): base work")
	runGit(t, repo, "update-ref", "refs/remotes/origin/feature", "feature")
	runGit(t, repo, "checkout", "-b", "demo")
	implSHA := commitAll(t, repo, "app.go", "package main\n", "feat(demo): do the thing")
	runGit(t, repo, "config", "branch.demo.flowBase", "feature")
	runGit(t, repo, "checkout", "main")

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	gitLog := bundle[strings.Index(bundle, "## git log --stat"):]
	if !strings.Contains(gitLog, "commit "+implSHA) {
		t.Errorf("git log section missing the change branch's commit %s:\n%s", implSHA, gitLog)
	}
	if strings.Contains(gitLog, "commit "+featureSHA) {
		t.Errorf("git log section carries the recorded base's own commit %s:\n%s", featureSHA, gitLog)
	}
}

// TestBundleAssemblyGitLogStaysAbsentWhenBranchMissing pins the fallback's
// degradation: no change branch to walk — it landed and was deleted, or
// never existed — leaves the source skipped, never a confident wrong
// answer.
func TestBundleAssemblyGitLogStaysAbsentWhenBranchMissing(t *testing.T) {
	repo := gitRepo(t)
	runGit(t, repo, "update-ref", "refs/remotes/origin/main", "main")
	runGit(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if !strings.Contains(bundle, "skipped: git log --stat (absent)") {
		t.Errorf("no change branch to walk, yet git log resolved:\n%s", bundle)
	}
}

// TestBundleAssemblyPrefersStoreRenderOverCommittedRecord pins the
// precedence: the store is the terminal record, so rows that DID reach it
// render the sections, and a ledger.md or panel.md a change
// directory carried is never served beside them.
func TestBundleAssemblyPrefersStoreRenderOverCommittedRecord(t *testing.T) {
	repo := gitRepo(t)
	writeChangeDir(t, repo, "demo", map[string]string{
		"tasks.md":  "# demo tasks\n",
		"ledger.md": "# SDD ledger — demo (stale copy)\n",
		"panel.md":  "# Review panel — demo (stale copy)\n",
	})
	run := records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{{
			Seq: 1, Role: "implementer", Model: "sonnet", SessionToken: "mf-demo",
		}},
	}

	bundle, err := Bundle("demo", run, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if !strings.Contains(bundle, "## .superpowers/sdd/ledgers/demo.md") {
		t.Errorf("store ledger not rendered:\n%s", bundle)
	}
	for _, stale := range []string{
		"## spectre/changes/demo/ledger.md",
		"## spectre/changes/demo/panel.md",
		"stale copy",
	} {
		if strings.Contains(bundle, stale) {
			t.Errorf("bundle serves the committed copy beside the store render:\n%s", bundle)
		}
	}
}

// TestBundleAssemblyRendersRecordsFromTheStoreOnly pins that the ledger and
// panel come from the store alone: with no store rows they are skipped
// under the store labels, and copies the change directory carries are
// never read.
func TestBundleAssemblyRendersRecordsFromTheStoreOnly(t *testing.T) {
	repo := gitRepo(t)
	writeChangeDir(t, repo, "demo", map[string]string{
		"tasks.md":  "# demo tasks\n",
		"ledger.md": "# SDD ledger — demo (committed copy)\n",
		"panel.md":  "# Review panel — demo (committed copy)\n",
	})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	for _, want := range []string{
		"skipped: .superpowers/sdd/ledgers/demo.md (absent)",
		"skipped: .superpowers/sdd/reviews/demo-panel.md (absent)",
	} {
		if !strings.Contains(bundle, want) {
			t.Errorf("bundle missing %q:\n%s", want, bundle)
		}
	}
	for _, gone := range []string{"committed copy", "## spectre/changes/demo/ledger.md", "## spectre/changes/demo/panel.md"} {
		if strings.Contains(bundle, gone) {
			t.Errorf("bundle reads a committed record copy (%q):\n%s", gone, bundle)
		}
	}
}

func TestBundleRefusesInvalidName(t *testing.T) {
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
	if _, err := Bundle("../escape", records.Run{}, "", false, false, "", nil, nil, g); err == nil {
		t.Fatal("an invalid change name must be refused, not assembled")
	}
}

// gitRepo initialises an empty repository with an initial commit on main,
// the shape every derived query below expects to walk.
func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	commitAll(t, dir, "README.md", "repo\n", "initial")
	return dir
}

// writeChangeDir writes the change's files under spectre/changes/<name>/ in
// repo's working tree, ignored and never committed, as a /flow run keeps
// them.
func writeChangeDir(t *testing.T, repo, name string, files map[string]string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repo, ".git", "info", "exclude"), []byte("spectre/changes/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for file, body := range files {
		path := filepath.Join(repo, "spectre/changes", name, file)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// onChangeBranch gives repo an origin/main base and checks out
// spectre/<name> from it, as a /flow worktree sits.
func onChangeBranch(t *testing.T, repo, name string) {
	t.Helper()
	runGit(t, repo, "update-ref", "refs/remotes/origin/main", "main")
	runGit(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	runGit(t, repo, "checkout", "-q", "-b", "spectre/"+name)
}

func commitAll(t *testing.T, repo, file, body, subject string) string {
	t.Helper()
	path := filepath.Join(repo, file)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "-A")
	runGit(t, repo, "commit", "-m", subject)
	out, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func runGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestExecRunnerBoundKillsHangingGit pins the bound itself, through the
// seam a test can drive: a PATH-shim git that never returns, an
// ExecRunner with a tiny Bound. The call must come back with an error well
// inside the bound — a mutant ignoring the injected Bound would
// blow this guard on the 10s default. The shim execs sleep so the
// context kill lands on the git process itself: the real path returns
// in ~500ms, not after WaitDelay.
func TestExecRunnerBoundKillsHangingGit(t *testing.T) {
	shimDir := t.TempDir()
	shim := filepath.Join(shimDir, "git")
	if err := os.WriteFile(shim, []byte("#!/bin/bash\nexec sleep 300\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	repo := t.TempDir()
	runner := ExecRunner{Bound: 500 * time.Millisecond}

	done := make(chan error, 1)
	go func() {
		_, err := runner.Output(repo, "rev-parse", "HEAD")
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("a hanging git answered successfully under the shim")
		}
	case <-time.After(8 * time.Second):
		// A mutant that ignores the injected Bound runs the 10s default —
		// past this guard. The real path returns in ~500ms; WaitDelay only
		// extends calls whose killed git left pipes held, and sleep does
		// not.
		t.Fatal("Output ignored the injected bound — still blocked 5s in")
	}
}

// TestBundleGitBudgetCapsSlowRepository pins the bundle's overall git
// budget: with every git call slow (a shim sleeping past each call's
// Bound), the serial calls of one bundle would add up to Bound per call —
// thirty-odd calls on a loaded machine outran the server's writeTimeout
// and the client got an empty reply. A Deadline caps them all together,
// so Bundle still returns, degraded, once the deadline passes. Without the
// Deadline check the shimmed bundle below runs 20 repositories' calls at
// 300ms each, several seconds past this guard.
func TestBundleGitBudgetCapsSlowRepository(t *testing.T) {
	shimDir := t.TempDir()
	shim := filepath.Join(shimDir, "git")
	if err := os.WriteFile(shim, []byte("#!/bin/bash\nexec sleep 300\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	var repos []string
	for range 20 {
		repos = append(repos, t.TempDir())
	}
	runner := ExecRunner{Bound: 300 * time.Millisecond, Deadline: time.Now().Add(500 * time.Millisecond)}

	start := time.Now()
	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", repos, nil, runner)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	if took := time.Since(start); took > 3*time.Second {
		t.Fatalf("Bundle took %s past a 500ms git deadline — the deadline does not cap the serial git calls", took)
	}
	if !strings.Contains(bundle, "could not be read") {
		t.Errorf("degraded bundle does not name the unreadable repositories:\n%s", bundle)
	}
}

// TestBundleServesRecordedSummaryFirst asserts the recorded change summary
// is the bundle's first source: the change's own statement of what it did
// and why is the central content a deferred reasoning pass reads, served
// verbatim from the store, ahead of the ledger and every history-derived
// source, and counted in the found line.
func TestBundleServesRecordedSummaryFirst(t *testing.T) {
	run := records.Run{Change: "demo"}
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}

	bundle, err := Bundle("demo", run, "# Change summary\n\nwhat changed and why", true, false, "", nil, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	summaryAt := strings.Index(bundle, "## change summary")
	ledgerAt := strings.Index(bundle, "## .superpowers/sdd/ledgers/demo.md")
	if summaryAt == -1 {
		t.Fatalf("bundle carries no change summary section:\n%s", bundle)
	}
	if ledgerAt != -1 && ledgerAt < summaryAt {
		t.Errorf("ledger section precedes the change summary:\n%s", bundle)
	}
	for _, want := range []string{
		"## change summary",
		"what changed and why",
		"found: 1 of 10 sources",
	} {
		if !strings.Contains(bundle, want) {
			t.Errorf("bundle missing %q:\n%s", want, bundle)
		}
	}
}

// TestBundleSkipsAbsentSummary asserts the missing-source shape: a change
// with no recorded summary reports it skipped like any other absent
// source, in the same words, and the found line counts the nine sources.
func TestBundleSkipsAbsentSummary(t *testing.T) {
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", nil, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if !strings.Contains(bundle, "found: 0 of 10 sources; skipped: 10 of 10 sources") {
		t.Errorf("summary line wrong:\n%s", bundle)
	}
	if !strings.Contains(bundle, "skipped: change summary (absent)") {
		t.Errorf("absent summary not reported skipped:\n%s", bundle)
	}
	if strings.Contains(bundle, "## change summary") {
		t.Errorf("no summary recorded, yet a section rendered:\n%s", bundle)
	}
}

func TestBundleNamesRecordsLossWhenPanelRanWithoutDispatches(t *testing.T) {
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, true, "", nil, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if !strings.Contains(bundle, "note: RECORDS LOSS") {
		t.Errorf("bundle carries no records-loss note:\n%s", bundle)
	}
	if !strings.Contains(bundle, "no dispatch rows") {
		t.Errorf("records-loss note never names the missing dispatch rows:\n%s", bundle)
	}
	// The note is a note, not a source: the found/skipped line still
	// counts exactly the nine real sources.
	if !strings.Contains(bundle, "found: 0 of 10 sources; skipped: 10 of 10 sources") {
		t.Errorf("found/skipped line disturbed by the note:\n%s", bundle)
	}
	if strings.Contains(bundle, "skipped: note:") {
		t.Errorf("the note leaked into the skipped list:\n%s", bundle)
	}
}

func TestBundleNoRecordsLossNoteWhenPanelDidNotRun(t *testing.T) {
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", nil, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if strings.Contains(bundle, "RECORDS LOSS") {
		t.Errorf("bundle names a records loss with no panel stage run behind it:\n%s", bundle)
	}
}

func TestBundleNoRecordsLossNoteWhenDispatchesExist(t *testing.T) {
	run := records.Run{
		Change:     "demo",
		Dispatches: []records.Dispatch{{Seq: 1, Role: "implementer", Model: "sonnet"}},
	}
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}

	bundle, err := Bundle("demo", run, "", false, true, "", nil, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if strings.Contains(bundle, "RECORDS LOSS") {
		t.Errorf("bundle names a records loss while the store holds dispatch rows:\n%s", bundle)
	}
}

// TestBundleAssemblyLogsEveryMergeOrderPeer pins a cross-repo change's
// satellite commits in the git-log source: each peer the change's link.md's
// ## Merge order names, resolved through spectre/peers relative to the
// canonical repository's main checkout, contributes its spectre/<name> commits.
func TestBundleAssemblyLogsEveryMergeOrderPeer(t *testing.T) {
	repo := gitRepo(t)
	peer := gitRepo(t)
	runGit(t, peer, "update-ref", "refs/remotes/origin/main", "main")
	runGit(t, peer, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	runGit(t, peer, "checkout", "-b", "spectre/demo")
	peerSHA := commitAll(t, peer, "app.go", "package main\n", "feat(app): the satellite's half")
	runGit(t, peer, "checkout", "main")
	rel, err := filepath.Rel(repo, peer)
	if err != nil {
		t.Fatal(err)
	}
	commitAll(t, repo, "spectre/peers", "front "+rel+"\n", "chore: declare the peer")
	writeChangeDir(t, repo, "demo", map[string]string{
		"tasks.md": "# demo tasks\n",
		"link.md":  "## Merge order\n\n1. `.`\n2. `front`\n",
	})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	gitLog := bundle[strings.Index(bundle, "## git log --stat"):]
	if !strings.Contains(gitLog, "### front") || !strings.Contains(gitLog, "commit "+peerSHA) {
		t.Errorf("git log section missing the peer's change commit %s under ### front:\n%s", peerSHA, gitLog)
	}
}

func TestBundleNamesTheServingBuild(t *testing.T) {
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "build 02a3d012e12a (2026-10-18T20:46:11Z)", nil, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	if !strings.Contains(bundle, "\nnote: served by flowd build 02a3d012e12a (2026-10-18T20:46:11Z) — ") {
		t.Errorf("no build note:\n%s", bundle)
	}

	bundle, err = Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", nil, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	if strings.Contains(bundle, "served by flowd") {
		t.Errorf("an empty build still rendered a note:\n%s", bundle)
	}
}

// recordingRunner is ExecRunner recording every call's args, so a test can
// assert a range was never read.
type recordingRunner struct {
	calls *[]string
}

func (r recordingRunner) Output(repo string, args ...string) ([]byte, error) {
	*r.calls = append(*r.calls, strings.Join(args, " "))
	return ExecRunner{}.Output(repo, args...)
}

// TestBundleRendersRequests pins a /flow-interactive session's git-log
// source (KAN-1061, design.md's Explicit stop step 2): each request in n
// order under `### Request <n>: <title> — <outcome>`, a landed one followed
// by its base..commit range's `git log --stat`, in place of the session
// branch's own log, which is empty by stop time; with no requests the
// section is the change branch's log as before.
func TestBundleRendersRequests(t *testing.T) {
	repo := gitRepo(t)
	base := commitAll(t, repo, "base.go", "package main\n", "chore: base")
	runGit(t, repo, "update-ref", "refs/remotes/origin/main", "main")
	runGit(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	runGit(t, repo, "checkout", "-b", "demo")
	first := commitAll(t, repo, "a.go", "package main\n", "feat(a): first landed")
	landed := commitAll(t, repo, "b.go", "package main\n", "feat(b): second landed")
	runGit(t, repo, "checkout", "main")
	runGit(t, repo, "merge", "--ff-only", "demo")
	runGit(t, repo, "update-ref", "refs/remotes/origin/main", "main")
	unlanded := commitAll(t, repo, "c.go", "package main\n", "feat(c): never landed")

	requests := []records.Request{
		{N: 2, Title: "Answer why", Outcome: "answered", State: "answered"},
		{N: 1, Title: "Fix the pill", Outcome: "landed", State: "landed", BaseSHA: base, CommitSHA: landed},
	}
	var calls []string
	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, requests, recordingRunner{&calls})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	gitLog := bundle[strings.Index(bundle, "## git log --stat"):]
	one := strings.Index(gitLog, "### Request 1: Fix the pill — landed")
	two := strings.Index(gitLog, "### Request 2: Answer why — answered")
	if one < 0 || two < 0 || one > two {
		t.Fatalf("git log section lacks the two request headings in n order:\n%s", gitLog)
	}
	landedLog := gitLog[one:two]
	for _, sha := range []string{first, landed} {
		if !strings.Contains(landedLog, "commit "+sha) {
			t.Errorf("request 1's log lacks commit %s:\n%s", sha, landedLog)
		}
	}
	if strings.Contains(gitLog[two:], "commit ") {
		t.Errorf("an answered request carries a log:\n%s", gitLog[two:])
	}
	if strings.Contains(gitLog, unlanded) {
		t.Errorf("git log section carries a commit outside every request's range:\n%s", gitLog)
	}
	for _, c := range calls {
		if strings.HasSuffix(c, "..demo") {
			t.Errorf("Bundle read the change branch's range %q with requests present", c)
		}
	}

	plain, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, "", []string{repo}, nil, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle without requests: %v", err)
	}
	if strings.Contains(plain, "### Request") {
		t.Errorf("a change without requests renders request headings:\n%s", plain)
	}
}

// TestRequestsLogPrintsEachCommitOnce pins requestsLog's de-duplication
// (kan-1081 self-review, F6): two requests whose landed ranges share commits
// print each commit's --stat block once, under the first request, and the
// second names it as landed with that request; an empty title renders
// (untitled).
func TestRequestsLogPrintsEachCommitOnce(t *testing.T) {
	repo := gitRepo(t)
	base := commitAll(t, repo, "base.go", "package main\n", "chore: base")
	c1 := commitAll(t, repo, "a.go", "package main\n", "feat(a): one")
	c2 := commitAll(t, repo, "b.go", "package main\n", "feat(b): two")
	c3 := commitAll(t, repo, "c.go", "package main\n", "feat(c): three")

	log := requestsLog(ExecRunner{}, repo, []records.Request{
		{N: 1, Title: "First", Outcome: "landed", State: "landed", BaseSHA: base, CommitSHA: c2},
		{N: 2, Title: "Second", Outcome: "landed", State: "landed", BaseSHA: base, CommitSHA: c2},
		{N: 3, Title: "", Outcome: "landed", State: "landed", BaseSHA: c2, CommitSHA: c3},
	})
	two := strings.Index(log, "### Request 2: Second — landed")
	three := strings.Index(log, "### Request 3: (untitled) — landed")
	if !strings.HasPrefix(log, "### Request 1: First — landed") || two < 0 || three < two {
		t.Fatalf("headings missing or out of order:\n%s", log)
	}
	for _, sha := range []string{c1, c2} {
		if n := strings.Count(log, "commit "+sha); n != 1 {
			t.Errorf("commit %s printed %d times, want once:\n%s", sha, n, log)
		}
		if !strings.Contains(log[:two], "commit "+sha) {
			t.Errorf("commit %s not under request 1:\n%s", sha, log)
		}
		if want := "commit " + sha[:12] + " — landed with Request 1"; !strings.Contains(log[two:three], want) {
			t.Errorf("request 2 lacks %q:\n%s", want, log)
		}
	}
	if !strings.Contains(log[three:], "commit "+c3) {
		t.Errorf("request 3 lacks its own commit %s:\n%s", c3, log)
	}
}

func TestBundleServesGatedReviewerFindings(t *testing.T) {
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
	run := records.Run{Change: "demo", Findings: []records.Finding{
		{Ref: "F1", Slot: records.GatedReviewerSlot, TaskID: "6", Severity: "Important", Location: "skills/flow/implement.md:12", Note: "no Model: line", Status: "fixed"},
		{Ref: "F2", Slot: "primary", Severity: "Minor", Note: "panel row", Status: "fixed"},
	}}
	bundle, err := Bundle("demo", run, "", false, false, "", nil, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	const heading = "## gated per-task reviewer findings"
	i := strings.Index(bundle, heading)
	if i < 0 {
		t.Fatalf("no gated section:\n%s", bundle)
	}
	if !strings.Contains(bundle, "| F1 | 6 | Important |") {
		t.Errorf("gated row missing:\n%s", bundle)
	}
	if strings.Contains(bundle[i:], "F2") {
		t.Errorf("panel row F2 leaked into the gated section:\n%s", bundle[i:])
	}
	panel := bundle[:i]
	if j := strings.Index(panel, "## .superpowers/sdd/reviews/demo-panel.md"); j < 0 || strings.Contains(panel[j:], "F1") {
		t.Errorf("panel section missing or carries the gated F1:\n%s", panel)
	}
}
