package gitops

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// kwFlowBuilt is the real flow CLI, built once from this checkout: step 3's
// `flow state add-worktree` runs against it, never a stand-in, pointed at a
// dead port and a temp FLOW_STATE_DIR so it writes the on-disk fallback and
// never reaches the dev store.
var kwFlowBuilt = sync.OnceValues(func() (string, error) {
	dir := filepath.Join(guardtest.ExecFixtures.Dir, "flow-cli")
	cmd := exec.Command("go", "build", "-o", dir+"/flow", "./cmd/flow")
	cmd.Dir = "../../.."
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("go build ./cmd/flow: %v\n%s", err, out)
	}
	return dir, nil
})

// kwFx is one kickoff fixture: a bare origin whose main carries
// .flow/project.md, the project cloned from it, and the environment the shim
// runs under.
type kwFx struct {
	origin, project string
	env             []string
}

const kwName = "kan-1-demo"

// kwNew builds the fixture; projectMD is .flow/project.md on origin's main
// ("" for none). withFlow puts the real flow CLI on PATH.
func kwNew(t *testing.T, projectMD string, withFlow bool) kwFx {
	t.Helper()
	fx := kwNoRepo(t, withFlow)
	guardtest.FixtureRepoWithOrigin(t, fx.project, fx.origin, guardtest.RepoShape{Branch: "main", Commit: "init"})
	var g guardtest.FxGit
	g.Git(fx.project, "config", "user.name", "test")
	g.Git(fx.project, "config", "user.email", "test@example.com")
	// Every project ignores .superpowers/, so kickoff adds no commit; one
	// that does not is TestKickoffWorktreeIgnoresScratch's.
	guardtest.WriteFile(t, fx.project+"/.gitignore", ".superpowers/\n")
	if projectMD != "" {
		guardtest.WriteFile(t, fx.project+"/.flow/project.md", projectMD)
	} else {
		guardtest.WriteFile(t, fx.project+"/README.md", "readme\n")
	}
	g.Git(fx.project, "add", "-A")
	g.Git(fx.project, "commit", "-q", "-m", "project")
	g.Git(fx.project, "push", "-q", "origin", "HEAD:main")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	return fx
}

// kwNoRepo is kwNew's paths and environment with no repository built: the
// fixture of a refusal kickoff makes before it reads any repository.
func kwNoRepo(t *testing.T, withFlow bool) kwFx {
	t.Helper()
	// Physical, as git records a worktree's path.
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fx := kwFx{origin: tmp + "/origin.git", project: tmp + "/project"}
	// PATH without any directory holding a flow binary, so "flow absent"
	// means absent; the real CLI is prepended when asked for.
	var path []string
	if withFlow {
		dir, err := kwFlowBuilt()
		if err != nil {
			t.Fatalf("build flow: %v", err)
		}
		path = append(path, dir)
	}
	// The shim and its guards run git over this PATH (FixtureGitPath).
	gitDir, err := guardtest.FixtureGitDir()
	if err != nil {
		t.Fatal(err)
	}
	path = append(path, gitDir)
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		if _, err := os.Stat(filepath.Join(d, "flow")); err != nil {
			path = append(path, d)
		}
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	dead := "http://" + ln.Addr().String()
	_ = ln.Close()
	fx.env = append(os.Environ(),
		"PATH="+strings.Join(path, string(filepath.ListSeparator)),
		"FLOW_GUARD_CACHE_DIR="+guardtest.GuardCache(t),
		"FLOW_ADDR="+dead, "FLOW_RECORDS_ADDR="+dead,
		"FLOW_STATE_DIR="+tmp+"/state",
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	return fx
}

// flow runs the real CLI under the fixture's environment.
func (fx kwFx) flow(t *testing.T, stdin string, args ...string) string {
	t.Helper()
	dir, err := kwFlowBuilt()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(dir+"/flow", args...)
	cmd.Env, cmd.Stdin = fx.env, strings.NewReader(stdin)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("flow %v: %v", args, err)
	}
	return string(out)
}

// started writes the STARTED record kickoff runs after.
func (fx kwFx) started(t *testing.T) {
	t.Helper()
	fx.flow(t, `{"state":"STARTED","updatedBy":"/flow","worktrees":{}}`, "state", "set", "-C", fx.project, kwName)
}

// run execs the real shim, so its siblings and the flow-guard build are the
// ones production runs.
func (fx kwFx) run(t *testing.T, args ...string) guardtest.GuardResult {
	t.Helper()
	cmd := exec.Command("bash", append([]string{guardtest.GuardShim(t, "kickoff-worktree")}, args...)...)
	cmd.Env = fx.env
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	rc := 0
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		rc = ee.ExitCode()
	}
	return guardtest.GuardResult{Rc: rc, Stdout: o.String(), Err: e.String(), Out: o.String() + e.String()}
}

func (fx kwFx) git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	var g guardtest.FxGit
	out := g.Git(dir, args...)
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	return out
}

const kwSetupMD = "# project\n\n## worktree setup\n\n```bash\ntouch setup-ran\necho second >> setup-ran\n```\n\nTrailing prose, never run.\n"

// Cost: 4.89 s — twelve runs of the real shim, each adding a worktree from a
// real origin, running setup, pushing and recording through the real flow
// CLI; the two refusals made before any repository is read build none.
// smoke: core — kickoff: worktree, branch, setup, push, the STARTED record
func TestKickoffWorktree(t *testing.T) {
	t.Parallel()
	if core.Registry["kickoff-worktree"] == nil {
		t.Fatal("kickoff-worktree is not registered")
	}
	// Built once before the fan-out, so no subtest waits on the build.
	if _, err := kwFlowBuilt(); err != nil {
		t.Fatalf("build flow: %v", err)
	}

	t.Run("fresh branch: add from origin/main, setup, push, persisted", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, kwSetupMD, true)
		fx.started(t)
		r := fx.run(t, fx.project, kwName)
		wt := fx.project + "-worktrees/" + kwName
		mb := fx.git(t, fx.origin, "rev-parse", "main")
		if r.Rc != 0 || r.Stdout != "worktree: "+wt+"\nmerge-base: "+mb+"\n" {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if b := fx.git(t, wt, "branch", "--show-current"); b != "spectre/"+kwName {
			t.Errorf("branch = %q", b)
		}
		if got, err := os.ReadFile(wt + "/setup-ran"); err != nil || string(got) != "second\n" {
			t.Errorf("setup-ran = %q, %v", got, err)
		}
		if tip := fx.git(t, fx.origin, "rev-parse", "spectre/"+kwName); tip != mb {
			t.Errorf("remote spectre/%s = %q, want %s", kwName, tip, mb)
		}
		if up := fx.git(t, wt, "rev-parse", "--abbrev-ref", "@{upstream}"); up != "origin/spectre/"+kwName {
			t.Errorf("upstream = %q", up)
		}
		if rec := fx.flow(t, "", "state", "get", "-C", fx.project, kwName); !strings.Contains(rec, `"`+wt+`":"`+mb+`"`) ||
			!strings.Contains(rec, `"state":"STARTED"`) {
			t.Errorf("record = %s", rec)
		}
	})

	t.Run("existing remote branch: tracks it, merge base is its tip", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		other := t.TempDir() + "/other"
		fx.git(t, "", "clone", "-q", fx.origin, other)
		fx.git(t, other, "config", "user.name", "test")
		fx.git(t, other, "config", "user.email", "test@example.com")
		fx.git(t, other, "commit", "-q", "--allow-empty", "-m", "captured plan")
		fx.git(t, other, "push", "-q", "origin", "HEAD:refs/heads/spectre/"+kwName)
		tip := fx.git(t, other, "rev-parse", "HEAD")
		fx.started(t)
		r := fx.run(t, fx.project, kwName)
		wt := fx.project + "-worktrees/" + kwName
		if r.Rc != 0 || r.Stdout != "worktree: "+wt+"\nmerge-base: "+tip+"\n" {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if !strings.Contains(r.Err, "no ## worktree setup") {
			t.Errorf("stderr does not say setup was skipped: %s", r.Err)
		}
	})

	t.Run("a failing setup command stops with exit 1, worktree already persisted", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "## worktree setup\n\n```bash\ntouch first-ran\nsh -c 'echo boom; exit 3'\ntouch never-ran\n```\n", true)
		fx.started(t)
		r := fx.run(t, fx.project, kwName)
		wt := fx.project + "-worktrees/" + kwName
		if r.Rc != 1 || !strings.Contains(r.Err, "sh -c 'echo boom; exit 3'") || !strings.Contains(r.Err, "boom") {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if _, err := os.Stat(wt + "/first-ran"); err != nil {
			t.Errorf("first command did not run: %v", err)
		}
		if _, err := os.Stat(wt + "/never-ran"); err == nil {
			t.Errorf("a command after the failing one ran")
		}
		if rec := fx.flow(t, "", "state", "get", "-C", fx.project, kwName); !strings.Contains(rec, `"`+wt+`":`) {
			t.Errorf("worktree not persisted: %s", rec)
		}
		var g guardtest.FxGit
		if g.Git(fx.origin, "rev-parse", "-q", "--verify", "spectre/"+kwName); g.Err == nil {
			t.Errorf("branch was pushed after a failed setup")
		}
	})

	t.Run("flow absent from PATH: exit 2 before any add", func(t *testing.T) {
		t.Parallel()
		fx := kwNoRepo(t, false)
		r := fx.run(t, fx.project, kwName)
		if r.Rc != 2 || r.Stdout != "" || !strings.Contains(r.Err, "kickoff-worktree: no flow binary on PATH") {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if _, err := os.Stat(fx.project + "-worktrees"); err == nil {
			t.Errorf("<project>-worktrees exists: a worktree was added")
		}
	})

	t.Run("no STARTED record: exit 1 after the add, nothing pushed", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		r := fx.run(t, fx.project, kwName)
		if r.Rc != 1 || !strings.Contains(r.Err, "write STARTED before adding a worktree") {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		var g guardtest.FxGit
		if g.Git(fx.origin, "rev-parse", "-q", "--verify", "spectre/"+kwName); g.Err == nil {
			t.Errorf("branch was pushed with nothing persisted")
		}
	})

	t.Run("named base: add from origin/<base>, base recorded on the branch", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		other := t.TempDir() + "/other"
		fx.git(t, "", "clone", "-q", fx.origin, other)
		fx.git(t, other, "commit", "-q", "--allow-empty", "-m", "feature work")
		fx.git(t, other, "push", "-q", "origin", "HEAD:refs/heads/feature/offline-mode")
		tip := fx.git(t, other, "rev-parse", "HEAD")
		fx.started(t)
		r := fx.run(t, fx.project, kwName, "feature/offline-mode")
		wt := fx.project + "-worktrees/" + kwName
		if r.Rc != 0 || r.Stdout != "worktree: "+wt+"\nmerge-base: "+tip+"\n" {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if got := fx.git(t, wt, "config", "--get", "branch.spectre/"+kwName+".flowBase"); got != "feature/offline-mode" {
			t.Errorf("flowBase = %q", got)
		}
	})

	t.Run("named base missing on origin: created from the default branch and recorded", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		fx.started(t)
		def := fx.git(t, fx.project, "rev-parse", "origin/HEAD")
		r := fx.run(t, fx.project, kwName, "feature/nope")
		if r.Rc != 0 || !strings.Contains(r.Err, "kickoff-worktree: origin/feature/nope did not exist") {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if got := fx.git(t, "", "--git-dir", fx.origin, "rev-parse", "refs/heads/feature/nope"); got != def {
			t.Errorf("origin feature/nope = %q, want %q", got, def)
		}
		wt := fx.project + "-worktrees/" + kwName
		if got := fx.git(t, wt, "config", "--get", "branch.spectre/"+kwName+".flowBase"); got != "feature/nope" {
			t.Errorf("flowBase = %q", got)
		}
	})

	// kwPushFeature pushes an empty commit to feature/x on fx's origin and
	// returns its tip.
	kwPushFeature := func(t *testing.T, fx kwFx) string {
		t.Helper()
		other := t.TempDir() + "/other"
		fx.git(t, "", "clone", "-q", fx.origin, other)
		fx.git(t, other, "commit", "-q", "--allow-empty", "-m", "feature work")
		fx.git(t, other, "push", "-q", "origin", "HEAD:refs/heads/feature/x")
		return fx.git(t, other, "rev-parse", "HEAD")
	}

	t.Run("--from: cut from the recorded base, recorded under the change's project", func(t *testing.T) {
		t.Parallel()
		canon, peer := kwNew(t, "", true), kwNew(t, "", true)
		kwPushFeature(t, canon)
		tip := kwPushFeature(t, peer)
		canon.started(t)
		if r := canon.run(t, canon.project, kwName, "feature/x"); r.Rc != 0 {
			t.Fatalf("kickoff rc=%d out=%s", r.Rc, r.Out)
		}
		r := canon.run(t, "--from", canon.project+"-worktrees/"+kwName, peer.project, kwName)
		wt := peer.project + "-worktrees/" + kwName
		if r.Rc != 0 || r.Stdout != "worktree: "+wt+"\nmerge-base: "+tip+"\n" {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if got := canon.git(t, wt, "config", "--get", "branch.spectre/"+kwName+".flowBase"); got != "feature/x" {
			t.Errorf("flowBase = %q", got)
		}
		if got := canon.flow(t, "", "state", "get", "-C", canon.project, kwName); !strings.Contains(got, wt) {
			t.Errorf("state record lacks %s: %s", wt, got)
		}
	})

	t.Run("--from: base missing on that origin: created there and recorded", func(t *testing.T) {
		t.Parallel()
		canon, peer := kwNew(t, "", true), kwNew(t, "", true)
		kwPushFeature(t, canon)
		canon.started(t)
		if r := canon.run(t, canon.project, kwName, "feature/x"); r.Rc != 0 {
			t.Fatalf("kickoff rc=%d out=%s", r.Rc, r.Out)
		}
		r := canon.run(t, "--from", canon.project+"-worktrees/"+kwName, peer.project, kwName)
		if r.Rc != 0 || !strings.Contains(r.Err, "origin/feature/x did not exist") {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if _, err := os.Stat(peer.project + "-worktrees/" + kwName); err != nil {
			t.Errorf("no worktree added: %v", err)
		}
		if got := canon.git(t, peer.project+"-worktrees/"+kwName, "config", "--get", "branch.spectre/"+kwName+".flowBase"); got != "feature/x" {
			t.Errorf("flowBase = %q", got)
		}
	})

	t.Run("--from a non-worktree path: exit 2, nothing added", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		fx.started(t)
		r := fx.run(t, "--from", t.TempDir(), fx.project, kwName)
		if r.Rc != 2 || !strings.Contains(r.Err, "is not a git worktree") {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if _, err := os.Stat(fx.project + "-worktrees/" + kwName); err == nil {
			t.Errorf("a worktree was added off an unreadable base")
		}
	})

	t.Run("pushed branch not on the named base: exit 1, rebase named, nothing added", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		kwPushFeature(t, fx)
		fx.git(t, fx.project, "push", "-q", "origin", "main:refs/heads/spectre/"+kwName)
		fx.started(t)
		r := fx.run(t, fx.project, kwName, "feature/x")
		if r.Rc != 1 || !strings.Contains(r.Err, "does not descend from origin/feature/x") || !strings.Contains(r.Err, "git rebase --onto origin/feature/x") {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
		if _, err := os.Stat(fx.project + "-worktrees/" + kwName); err == nil {
			t.Errorf("a worktree was added off the wrong base")
		}
	})

	t.Run("usage: exit 2", func(t *testing.T) {
		t.Parallel()
		fx := kwNoRepo(t, true)
		r := fx.run(t, fx.project)
		if r.Rc != 2 || r.Err != "usage: flow-guard kickoff-worktree [--from <worktree>] <project> <name> [<base>]\n" {
			t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
		}
	})
}

// TestKickoffWorktreeSiblingLayout pins kickoff on the sibling layout
// (design.md: sibling-worktrees-layout): the worktree is created at
// <parent>/<repo>-worktrees/<name>, outside the repository, nothing is
// written to info/exclude, and the `worktree:` line names that path.
func TestKickoffWorktreeSiblingLayout(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := kwNew(t, "", true)
	fx.started(t)
	exclude := fx.git(t, fx.project, "rev-parse", "--path-format=absolute", "--git-path", "info/exclude")
	before, _ := os.ReadFile(exclude)
	r := fx.run(t, fx.project, kwName)
	wt := filepath.Dir(fx.project) + "/project-worktrees/" + kwName
	if r.Rc != 0 || !strings.HasPrefix(r.Stdout, "worktree: "+wt+"\n") {
		t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
	}
	if b := fx.git(t, wt, "branch", "--show-current"); b != "spectre/"+kwName {
		t.Errorf("branch = %q", b)
	}
	if after, _ := os.ReadFile(exclude); string(after) != string(before) {
		t.Errorf("info/exclude changed: %q -> %q", before, after)
	}
	if st := fx.git(t, fx.project, "status", "--porcelain", "--ignored"); st != "" {
		t.Errorf("project status = %q, want clean (the worktree is outside the repository)", st)
	}
}

// A project reached through a symlink gets its worktree beside the physical
// checkout.
func TestKickoffWorktreeSymlinkedProject(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := kwNew(t, "", true)
	fx.started(t)
	link := filepath.Dir(fx.project) + "/link/via"
	guardtest.Mkdir(t, filepath.Dir(link))
	if err := os.Symlink(fx.project, link); err != nil {
		t.Fatal(err)
	}
	r := fx.run(t, link, kwName)
	wt := fx.project + "-worktrees/" + kwName
	if r.Rc != 0 || !strings.HasPrefix(r.Stdout, "worktree: "+wt+"\n") {
		t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
	}
}

// kan-1017 (sr15): the hand-written merge-order record failed `spectre
// validate` twice, so kickoff --from writes it into the canonical worktree.
func TestKickoffWorktreeFromWritesMergeOrder(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, tc := range []struct {
		name        string
		peers, link string // "" link: none written beforehand
		changeDir   bool
		rc          int
		want        string // link.md afterwards; "" none
	}{
		{"no link.md: written", "peer", "", true, 0, "## Merge order\n\n1. `.`\n2. `peer`\n"},
		{"appended after the entries found", "peer", "# link\n\n## Merge order\n\n1. `.`\n2. `other`\n\n## Notes\n\nx\n", true, 0,
			"# link\n\n## Merge order\n\n1. `.`\n2. `other`\n3. `peer`\n\n## Notes\n\nx\n"},
		{"already named: unchanged", "peer", "## Merge order\n\n1. `.`\n2. `peer`\n", true, 0, "## Merge order\n\n1. `.`\n2. `peer`\n"},
		{"missing from peers: exit 2", "", "", true, 2, ""},
		{"no change directory: nothing written", "peer", "", false, 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			canon, peer := kwNew(t, "", true), kwNew(t, "", true)
			canon.started(t)
			if r := canon.run(t, canon.project, kwName); r.Rc != 0 {
				t.Fatalf("kickoff rc=%d out=%s", r.Rc, r.Out)
			}
			from := canon.project + "-worktrees/" + kwName
			rel, err := filepath.Rel(canon.project, peer.project)
			if err != nil {
				t.Fatal(err)
			}
			peers := "frontend ../frontend\n"
			if tc.peers != "" {
				peers += tc.peers + " " + rel + "\n"
			}
			guardtest.WriteFile(t, from+"/spectre/peers", peers)
			link := from + "/spectre/changes/" + kwName + "/link.md"
			if tc.changeDir {
				guardtest.Mkdir(t, filepath.Dir(link))
			}
			if tc.link != "" {
				guardtest.WriteFile(t, link, tc.link)
			}
			r := canon.run(t, "--from", from, peer.project, kwName)
			if r.Rc != tc.rc {
				t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
			}
			got, _ := os.ReadFile(link)
			if string(got) != tc.want {
				t.Errorf("link.md:\n%s\nwant:\n%s", got, tc.want)
			}
			if tc.rc == 2 {
				if _, err := os.Stat(peer.project + "-worktrees/" + kwName); err == nil {
					t.Errorf("a worktree was added")
				}
				if st := canon.flow(t, "", "state", "get", "-C", canon.project, kwName); strings.Contains(st, peer.project) {
					t.Errorf("worktree recorded: %s", st)
				}
			}
		})
	}
}

// TestKickoffWorktreeFromListsEveryRecordedWorktree: a peer added with
// --from before the change directory existed is still listed when a later
// peer's kickoff writes link.md, in spectre/peers order.
func TestKickoffWorktreeFromListsEveryRecordedWorktree(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	canon, a, b := kwNew(t, "", true), kwNew(t, "", true), kwNew(t, "", true)
	canon.started(t)
	if r := canon.run(t, canon.project, kwName); r.Rc != 0 {
		t.Fatalf("kickoff rc=%d out=%s", r.Rc, r.Out)
	}
	from := canon.project + "-worktrees/" + kwName
	peers := ""
	for _, p := range []struct{ name, dir string }{{"alpha", a.project}, {"beta", b.project}} {
		rel, err := filepath.Rel(canon.project, p.dir)
		if err != nil {
			t.Fatal(err)
		}
		peers += p.name + " " + rel + "\n"
	}
	guardtest.WriteFile(t, from+"/spectre/peers", peers)
	if r := canon.run(t, "--from", from, a.project, kwName); r.Rc != 0 {
		t.Fatalf("alpha rc=%d out=%s", r.Rc, r.Out)
	}
	link := from + "/spectre/changes/" + kwName + "/link.md"
	guardtest.Mkdir(t, filepath.Dir(link))
	if r := canon.run(t, "--from", from, b.project, kwName); r.Rc != 0 {
		t.Fatalf("beta rc=%d out=%s", r.Rc, r.Out)
	}
	got, _ := os.ReadFile(link)
	if want := "## Merge order\n\n1. `.`\n2. `alpha`\n3. `beta`\n"; string(got) != want {
		t.Errorf("link.md:\n%s\nwant:\n%s", got, want)
	}
}

// TestKwPeerNameNamesWhichReadFailed pins the kan-1024 self-review fix: a
// missing spectre/peers and an unresolvable main checkout are reported as
// the two distinct causes they are.
func TestKwPeerNameNamesWhichReadFailed(t *testing.T) {
	t.Parallel()
	env := core.Env{Getenv: os.Getenv}
	from := t.TempDir()
	if _, msg := kwPeerName(env, "git", from, "/p"); !strings.HasPrefix(msg, "cannot read ") {
		t.Errorf("no peers file: %q", msg)
	}
	if err := os.MkdirAll(filepath.Join(from, "spectre"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(from, "spectre/peers"), []byte("p ../p\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, msg := kwPeerName(env, "git", from, "/p"); !strings.HasPrefix(msg, "cannot resolve ") {
		t.Errorf("no main checkout: %q", msg)
	}
}

// TestKickoffWorktreeIgnoresScratch pins kickoff's ignore-scratch call
// (kan-1081 self-review, F2): in a project that does not ignore
// .superpowers/, the new branch's first commit ignores it, pushed with the
// branch, while the merge base stays the commit the add checked out.
func TestKickoffWorktreeIgnoresScratch(t *testing.T) {
	t.Parallel()
	fx := kwNew(t, "", true)
	other := t.TempDir() + "/other"
	fx.git(t, "", "clone", "-q", fx.origin, other)
	fx.git(t, other, "rm", "-q", ".gitignore")
	fx.git(t, other, "-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "-m", "drop ignore")
	fx.git(t, other, "push", "-q", "origin", "HEAD:main")
	fx.started(t)
	r := fx.run(t, fx.project, kwName)
	wt := fx.project + "-worktrees/" + kwName
	if r.Rc != 0 {
		t.Fatalf("rc=%d out=%s", r.Rc, r.Out)
	}
	if s := fx.git(t, wt, "log", "-1", "--format=%s"); s != "chore: ignore .superpowers/ scratch" {
		t.Errorf("HEAD subject = %q", s)
	}
	head := fx.git(t, wt, "rev-parse", "HEAD")
	if tip := fx.git(t, wt, "rev-parse", "origin/spectre/"+kwName); tip != head {
		t.Errorf("origin/spectre/%s = %s, want HEAD %s", kwName, tip, head)
	}
	if want := "worktree: " + wt + "\nmerge-base: " + fx.git(t, wt, "rev-parse", "HEAD^") + "\n"; r.Stdout != want {
		t.Errorf("stdout = %q, want %q", r.Stdout, want)
	}
}
