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
	g := &fakeGit{
		responses: map[string]string{
			"/repo show chore/archive-demo:spectre/changes/archive/demo/tasks.md": "# demo tasks\n",
		},
		failOn: map[string]bool{
			"/repo show chore/archive-demo:spectre/changes/archive/demo/design.md":    true,
			"/repo show chore/archive-demo:spectre/changes/archive/demo/narrative.md": true,
		},
	}

	bundle, err := Bundle("demo", run, []string{"/repo"}, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	for _, want := range []string{
		"# Self-review context bundle for demo",
		"## .superpowers/sdd/ledgers/demo.md",
		"## .superpowers/sdd/reviews/demo-panel.md",
		"## spectre/changes/archive/demo/tasks.md",
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
		"spectre/changes/archive/demo/design.md",
		"spectre/changes/archive/demo/narrative.md",
		"git log --stat",
	} {
		if !strings.Contains(bundle, "skipped: "+label+" (absent)") {
			t.Errorf("absent source %q not reported skipped:\n%s", label, bundle)
		}
	}
	// The labels name the archived change repo-relatively: no absolute
	// repository path may appear anywhere in the bundle.
	if strings.Contains(bundle, "/repo") {
		t.Errorf("bundle leaks a repository path:\n%s", bundle)
	}
}

func TestBundleAssemblySkipsAbsentSources(t *testing.T) {
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, nil, g)
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	// A change the store has never heard of renders nothing: the ledger
	// needs dispatch rows and the panel needs any row at all.
	if !strings.Contains(bundle, "found: 0 of 6 sources; skipped: 6 of 6 sources") {
		t.Errorf("summary line wrong:\n%s", bundle)
	}
	if strings.Contains(bundle, "## .superpowers/sdd/ledgers/demo.md") {
		t.Errorf("no dispatch rows, yet a ledger section rendered:\n%s", bundle)
	}
	if strings.Contains(bundle, "## .superpowers/sdd/reviews/demo-panel.md") {
		t.Errorf("no rows at all, yet a panel section rendered:\n%s", bundle)
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
// simply never archived wears.
func TestBundleAssemblyNotesUnreadableRepo(t *testing.T) {
	bogus := filepath.Join(t.TempDir(), "not-a-repo")

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, []string{bogus}, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if !strings.Contains(bundle, "note: repository "+bogus+" could not be read") {
		t.Errorf("unreadable repository not noted:\n%s", bundle)
	}
	if !strings.Contains(bundle, "skipped: spectre/changes/archive/demo/tasks.md (absent)") {
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
		failOn: map[string]bool{
			repo + " show chore/archive-demo:spectre/changes/archive/demo/tasks.md": true,
		},
	}

	if _, err := Bundle("demo", records.Run{Change: "demo"}, []string{repo}, g); err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	for _, call := range g.calls {
		if strings.HasPrefix(call, repo+" rev-parse") && call != repo+" rev-parse --git-dir" {
			t.Errorf("health probe issued %q, want rev-parse --git-dir", call)
		}
	}
}

func TestBundleAssemblyReadsArchivedFilesThroughGit(t *testing.T) {
	repo := gitRepo(t)
	writeArchiveBranch(t, repo, "demo", map[string]string{
		"tasks.md":     "# demo tasks\n",
		"design.md":    "# demo design\n",
		"narrative.md": "# demo narrative\n",
	})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, []string{repo}, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	for _, want := range []string{
		"## spectre/changes/archive/demo/tasks.md",
		"# demo tasks",
		"## spectre/changes/archive/demo/design.md",
		"# demo design",
		"## spectre/changes/archive/demo/narrative.md",
		"# demo narrative",
	} {
		if !strings.Contains(bundle, want) {
			t.Errorf("bundle missing %q:\n%s", want, bundle)
		}
	}
}

func TestBundleAssemblyDerivesFinishCommits(t *testing.T) {
	repo := gitRepo(t)
	// The archive branch is cut from main AFTER the implementation and
	// planning commits, as run 2 cuts it — the branch carries all three
	// commits, which is what lets the derivation resolve them from the
	// branch alone.
	implSHA := commitAll(t, repo, "app.go", "package main\n", "feat(demo): do the thing")
	planSHA := commitAll(t, repo, "spectre/changes/demo/tasks.md", "- [ ] 1. do it\n",
		"chore(spectre): plan and session records")
	archiveSHA := writeArchiveBranch(t, repo, "demo", map[string]string{"tasks.md": "# demo tasks\n"})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, []string{repo}, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	gitLog := bundle[strings.Index(bundle, "## git log --stat"):]
	for _, sha := range []string{implSHA, planSHA, archiveSHA} {
		if !strings.Contains(gitLog, "commit "+sha) {
			t.Errorf("git log section missing commit %s:\n%s", sha, gitLog)
		}
	}
	// Order: implementation before planning before archive.
	if strings.Index(gitLog, "commit "+implSHA) > strings.Index(gitLog, "commit "+planSHA) {
		t.Errorf("implementation commit must precede the planning commit:\n%s", gitLog)
	}
	if strings.Contains(bundle, "skipped: git log --stat") {
		t.Errorf("git log source resolved but reported skipped:\n%s", bundle)
	}
}

// TestBundleAssemblyFallsBackToCommittedRecords pins the KAN-552 fallback:
// a change whose rows never reached the store — the kan-468 failure —
// still gets its ledger and panel sections, read out of the copies run 2
// step 4 commits onto the archive branch. The sections are labelled by the
// committed path, the one provenance that is true of their content.
func TestBundleAssemblyFallsBackToCommittedRecords(t *testing.T) {
	repo := gitRepo(t)
	writeArchiveBranch(t, repo, "demo", map[string]string{
		"tasks.md":  "# demo tasks\n",
		"ledger.md": "# SDD ledger — demo\n",
		"panel.md":  "# Review panel — demo\n",
	})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, []string{repo}, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	for _, want := range []string{
		"## spectre/changes/archive/demo/ledger.md",
		"# SDD ledger — demo",
		"## spectre/changes/archive/demo/panel.md",
		"# Review panel — demo",
		// tasks.md and the archive commit's git log resolve besides the two
		// fallback sections; design.md and narrative.md stay skipped.
		"found: 4 of 6 sources",
		"skipped: spectre/changes/archive/demo/design.md (absent)",
		"skipped: spectre/changes/archive/demo/narrative.md (absent)",
	} {
		if !strings.Contains(bundle, want) {
			t.Errorf("bundle missing %q:\n%s", want, bundle)
		}
	}
	for _, absent := range []string{
		"## .superpowers/sdd/ledgers/demo.md",
		"## .superpowers/sdd/reviews/demo-panel.md",
		"skipped: spectre/changes/archive/demo/ledger.md",
		"skipped: spectre/changes/archive/demo/panel.md",
	} {
		if strings.Contains(bundle, absent) {
			t.Errorf("bundle carries %q:\n%s", absent, bundle)
		}
	}
}

// TestBundleAssemblyPrefersStoreRenderOverCommittedRecord pins the
// precedence: the store is the terminal record, so rows that DID reach it
// render the sections and the step-4 committed copies are never served
// beside them.
func TestBundleAssemblyPrefersStoreRenderOverCommittedRecord(t *testing.T) {
	repo := gitRepo(t)
	writeArchiveBranch(t, repo, "demo", map[string]string{
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

	bundle, err := Bundle("demo", run, []string{repo}, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	if !strings.Contains(bundle, "## .superpowers/sdd/ledgers/demo.md") {
		t.Errorf("store ledger not rendered:\n%s", bundle)
	}
	for _, stale := range []string{
		"## spectre/changes/archive/demo/ledger.md",
		"## spectre/changes/archive/demo/panel.md",
		"stale copy",
	} {
		if strings.Contains(bundle, stale) {
			t.Errorf("bundle serves the committed copy beside the store render:\n%s", bundle)
		}
	}
}

// TestBundleAssemblySkipsCommittedRecordsWhenAbsent keeps the fallback
// honest in the other direction: a readable archive branch that carries no
// copies leaves the store labels' skip lines exactly as they were.
func TestBundleAssemblySkipsCommittedRecordsWhenAbsent(t *testing.T) {
	repo := gitRepo(t)
	writeArchiveBranch(t, repo, "demo", map[string]string{"tasks.md": "# demo tasks\n"})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, []string{repo}, ExecRunner{})
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
}

// TestDeriveFinishCommitsRefusesMergeParent pins the merge gate's visible
// half: a plan commit whose parent is a merge resolves no implementation
// commit. The gate and the subject rejections are additionally
// defence-in-depth against future edits to commit-split.sh — their other
// conditions are unreachable-dead through this pipeline's own commits, a
// documented state, not a coverage gap.
func TestDeriveFinishCommitsRefusesMergeParent(t *testing.T) {
	repo := gitRepo(t)

	// Two diverging branches merged into one commit: whatever the merge's
	// subject and whatever it touches, a merge is never the implementation
	// commit. It sits directly below the plan commit here.
	base := commitAll(t, repo, "base.txt", "base\n", "base")
	runGit(t, repo, "checkout", "-b", "side", base)
	_ = commitAll(t, repo, "side.txt", "side\n", "side")
	runGit(t, repo, "checkout", "main")
	merge := mergeCommit(t, repo, "side", "Merge branch 'side'")
	commitAll(t, repo, "spectre/changes/demo/tasks.md", "- [ ] 1. do it\n",
		"chore(spectre): plan and session records")
	// The archive branch is cut last, so the derivation's branch-scoped
	// queries see the merge and the plan commit.
	writeArchiveBranch(t, repo, "demo", map[string]string{"tasks.md": "# demo tasks\n"})

	fc := deriveFinishCommits(ExecRunner{}, repo, "demo")
	if fc.impl != "" {
		t.Errorf("merge commit %s accepted as the implementation commit", merge)
	}
	if fc.plan == "" || fc.archive == "" {
		t.Errorf("plan and archive commits must still resolve, got %+v", fc)
	}
}

// TestDeriveFinishCommitsRefusesPlanningOnlyParent pins the exclusion's
// visible half: a plan commit whose parent touched only the planning trees
// is not an implementation commit, even with an unreserved subject and a
// single parent.
func TestDeriveFinishCommitsRefusesPlanningOnlyParent(t *testing.T) {
	repo := gitRepo(t)
	commitAll(t, repo, "spectre/changes/demo/tasks.md", "- [ ] 1. do it\n", "work")
	planSHA := commitAll(t, repo, "spectre/changes/demo/tasks.md", "- [ ] 1. do it\n\t- [ ] step\n",
		"chore(spectre): plan and session records")
	writeArchiveBranch(t, repo, "demo", map[string]string{"tasks.md": "# demo tasks\n"})

	fc := deriveFinishCommits(ExecRunner{}, repo, "demo")
	if fc.impl != "" {
		t.Errorf("planning-only commit accepted as the implementation commit below %s", planSHA)
	}
	if fc.plan == "" || fc.archive == "" {
		t.Errorf("plan and archive commits must still resolve, got %+v", fc)
	}
}

// TestDeriveFinishCommitsSiblingArchiveSubjectLoses pins the archive
// subject's end anchor: a later sibling change's "archive demo-fix-1"
// commit must never outrank this change's own archive commit.
func TestDeriveFinishCommitsSiblingArchiveSubjectLoses(t *testing.T) {
	repo := gitRepo(t)
	implSHA := commitAll(t, repo, "app.go", "package main\n", "feat(demo): do the thing")
	planSHA := commitAll(t, repo, "spectre/changes/demo/tasks.md", "- [ ] 1. do it\n",
		"chore(spectre): plan and session records")
	archiveSHA := writeArchiveBranch(t, repo, "demo", map[string]string{"tasks.md": "# demo tasks\n"})
	// A later commit ON THE ARCHIVE BRANCH, whose subject carries the
	// change's name as a proper prefix of a DIFFERENT change's archive
	// subject.
	runGit(t, repo, "checkout", "chore/archive-demo")
	commitAll(t, repo, "spectre/changes/archive/demo/notes.md", "note\n",
		"chore(spectre): archive demo-fix-1")
	runGit(t, repo, "checkout", "main")

	fc := deriveFinishCommits(ExecRunner{}, repo, "demo")
	if fc.archive != archiveSHA {
		t.Errorf("archive sha = %s, want the change's own %s", fc.archive, archiveSHA)
	}
	if fc.plan != planSHA || fc.impl != implSHA {
		t.Errorf("plan/impl = %s/%s, want %s/%s", fc.plan, fc.impl, planSHA, implSHA)
	}
}

func TestBundleRefusesInvalidName(t *testing.T) {
	g := &fakeGit{responses: map[string]string{}, failOn: map[string]bool{}}
	if _, err := Bundle("../escape", records.Run{}, nil, g); err == nil {
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

// writeArchiveBranch creates the archived change's files and commits them
// on chore/archive-<name>, then returns to main — run 2's archive step in
// miniature, returning the archive commit's sha.
func writeArchiveBranch(t *testing.T, repo, name string, files map[string]string) string {
	t.Helper()
	for file, body := range files {
		path := filepath.Join(repo, "spectre/changes/archive", name, file)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, repo, "checkout", "-b", "chore/archive-"+name)
	runGit(t, repo, "add", "-A")
	runGit(t, repo, "commit", "-m", "chore(spectre): archive "+name)
	out, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.TrimSpace(string(out))
	runGit(t, repo, "checkout", "main")
	return sha
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

func mergeCommit(t *testing.T, repo, branch, subject string) string {
	t.Helper()
	runGit(t, repo, "merge", "--no-ff", branch, "-m", subject)
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
