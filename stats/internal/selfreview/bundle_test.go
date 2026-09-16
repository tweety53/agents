package selfreview

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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
			"/repo rev-parse --verify --quiet chore/archive-demo": "",
		},
		failOn: map[string]bool{
			"/repo show chore/archive-demo:spectre/changes/archive/demo/tasks.md":     true,
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
	} {
		if !strings.Contains(bundle, want) {
			t.Errorf("bundle missing %q:\n%s", want, bundle)
		}
	}
	if strings.Contains(bundle, "skipped: .superpowers/sdd/ledgers/demo.md") {
		t.Errorf("ledger rendered but reported skipped:\n%s", bundle)
	}
	for _, label := range []string{
		"spectre/changes/archive/demo/tasks.md",
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

	// The panel record renders unconditionally (records.RenderKind's own
	// presence rule), so one of the six is always found.
	if !strings.Contains(bundle, "found: 1 of 6 sources; skipped: 5 of 6 sources") {
		t.Errorf("summary line wrong:\n%s", bundle)
	}
	if strings.Contains(bundle, "## .superpowers/sdd/ledgers/demo.md") {
		t.Errorf("no dispatch rows, yet a ledger section rendered:\n%s", bundle)
	}
	if !strings.Contains(bundle, "skipped: git log --stat (absent)") {
		t.Errorf("no repository recorded, yet git log resolved:\n%s", bundle)
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
	writeArchiveBranch(t, repo, "demo", map[string]string{"tasks.md": "# demo tasks\n"})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, []string{repo}, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}

	gitLog := bundle[strings.Index(bundle, "## git log --stat"):]
	for _, sha := range []string{implSHA, planSHA} {
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
// miniature. The commit subject is the exact archive subject the
// derivation matches.
func writeArchiveBranch(t *testing.T, repo, name string, files map[string]string) {
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
	runGit(t, repo, "checkout", "main")
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
