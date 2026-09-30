package guard

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
)

// kwFlowBuilt is the real flow CLI, built once from this checkout: step 3's
// `flow state add-worktree` runs against it, never a stand-in, pointed at a
// dead port and a temp FLOW_STATE_DIR so it writes the on-disk fallback and
// never reaches the dev store.
var kwFlowBuilt = sync.OnceValues(func() (string, error) {
	dir := filepath.Join(execFixtures.dir, "flow-cli")
	cmd := exec.Command("go", "build", "-o", dir+"/flow", "./cmd/flow")
	cmd.Dir = "../.."
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
	tmp := t.TempDir()
	fx := kwFx{origin: tmp + "/origin.git", project: tmp + "/project"}
	var g fxGit
	g.git("", "init", "-q", "--bare", "-b", "main", fx.origin)
	seed := tmp + "/seed"
	g.git("", "clone", "-q", fx.origin, seed)
	g.git(seed, "config", "user.name", "test")
	g.git(seed, "config", "user.email", "test@example.com")
	if projectMD != "" {
		writeFile(t, seed+"/.flow/project.md", projectMD)
	} else {
		writeFile(t, seed+"/README.md", "readme\n")
	}
	g.git(seed, "add", "-A")
	g.git(seed, "commit", "-q", "-m", "init")
	g.git(seed, "push", "-q", "origin", "HEAD:main")
	g.git("", "clone", "-q", fx.origin, fx.project)
	g.git(fx.project, "config", "user.name", "test")
	g.git(fx.project, "config", "user.email", "test@example.com")
	if g.err != nil {
		t.Fatal(g.err)
	}

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
		"FLOW_GUARD_CACHE_DIR="+guardCache(t),
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
func (fx kwFx) run(t *testing.T, args ...string) guardResult {
	t.Helper()
	cmd := exec.Command("bash", append([]string{tcfScriptsDir(t) + "/kickoff-worktree.sh"}, args...)...)
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
	return guardResult{rc: rc, stdout: o.String(), err: e.String(), out: o.String() + e.String()}
}

func (fx kwFx) git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	var g fxGit
	out := g.git(dir, args...)
	if g.err != nil {
		t.Fatal(g.err)
	}
	return out
}

const kwSetupMD = "# project\n\n## worktree setup\n\n```bash\ntouch setup-ran\necho second >> setup-ran\n```\n\nTrailing prose, never run.\n"

func TestKickoffWorktree(t *testing.T) {
	t.Parallel()
	if Registry["kickoff-worktree"] == nil {
		t.Fatal("kickoff-worktree is not registered")
	}

	t.Run("fresh branch: add from origin/main, setup, push, persisted", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, kwSetupMD, true)
		fx.started(t)
		r := fx.run(t, fx.project, kwName)
		wt := fx.project + "/.worktrees/" + kwName
		mb := fx.git(t, fx.origin, "rev-parse", "main")
		if r.rc != 0 || r.stdout != "worktree: "+wt+"\nmerge-base: "+mb+"\n" {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
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
		wt := fx.project + "/.worktrees/" + kwName
		if r.rc != 0 || r.stdout != "worktree: "+wt+"\nmerge-base: "+tip+"\n" {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
		}
		if !strings.Contains(r.err, "no ## worktree setup") {
			t.Errorf("stderr does not say setup was skipped: %s", r.err)
		}
	})

	t.Run(".worktrees not ignored: appended to info/exclude once", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		fx.started(t)
		exclude := fx.git(t, fx.project, "rev-parse", "--path-format=absolute", "--git-path", "info/exclude")
		if r := fx.run(t, fx.project, kwName); r.rc != 0 {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
		}
		b, _ := os.ReadFile(exclude)
		if n := strings.Count("\n"+string(b), "\n.worktrees/\n"); n != 1 || !strings.HasSuffix(string(b), "\n.worktrees/\n") {
			t.Errorf("info/exclude = %q", b)
		}
		if st := fx.git(t, fx.project, "status", "--porcelain"); st != "" {
			t.Errorf("project status = %q, want clean (nothing committed or untracked)", st)
		}
	})

	t.Run(".worktrees already ignored: info/exclude untouched", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		fx.started(t)
		exclude := fx.git(t, fx.project, "rev-parse", "--path-format=absolute", "--git-path", "info/exclude")
		writeFile(t, exclude, ".worktrees\n")
		if r := fx.run(t, fx.project, kwName); r.rc != 0 {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
		}
		if b, _ := os.ReadFile(exclude); string(b) != ".worktrees\n" {
			t.Errorf("info/exclude = %q", b)
		}
	})

	t.Run("a failing setup command stops with exit 1, worktree already persisted", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "## worktree setup\n\n```bash\ntouch first-ran\nsh -c 'echo boom; exit 3'\ntouch never-ran\n```\n", true)
		fx.started(t)
		r := fx.run(t, fx.project, kwName)
		wt := fx.project + "/.worktrees/" + kwName
		if r.rc != 1 || !strings.Contains(r.err, "sh -c 'echo boom; exit 3'") || !strings.Contains(r.err, "boom") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
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
		var g fxGit
		if g.git(fx.origin, "rev-parse", "-q", "--verify", "spectre/"+kwName); g.err == nil {
			t.Errorf("branch was pushed after a failed setup")
		}
	})

	t.Run("flow absent from PATH: exit 2 before any add", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", false)
		r := fx.run(t, fx.project, kwName)
		if r.rc != 2 || r.stdout != "" || !strings.Contains(r.err, "kickoff-worktree: no flow binary on PATH") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
		}
		if _, err := os.Stat(fx.project + "/.worktrees"); err == nil {
			t.Errorf(".worktrees exists: a worktree was added")
		}
	})

	t.Run("location guard refusal: exit 1 with its lines, nothing added", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		fx.started(t)
		stray := t.TempDir() + "/stray"
		fx.git(t, fx.project, "worktree", "add", "-q", "-b", "stray", stray)
		r := fx.run(t, fx.project, kwName)
		if r.rc != 1 || !strings.Contains(r.out, "STRAY: ") || !strings.Contains(r.out, "LOCATION-STRAY: ") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
		}
		if _, err := os.Stat(fx.project + "/.worktrees/" + kwName); err == nil {
			t.Errorf("a worktree was added after the location guard refused")
		}
	})

	t.Run("no STARTED record: exit 1 after the add, nothing pushed", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		r := fx.run(t, fx.project, kwName)
		if r.rc != 1 || !strings.Contains(r.err, "write STARTED before adding a worktree") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
		}
		var g fxGit
		if g.git(fx.origin, "rev-parse", "-q", "--verify", "spectre/"+kwName); g.err == nil {
			t.Errorf("branch was pushed with nothing persisted")
		}
	})

	t.Run("usage: exit 2", func(t *testing.T) {
		t.Parallel()
		fx := kwNew(t, "", true)
		r := fx.run(t, fx.project)
		if r.rc != 2 || r.err != "usage: kickoff-worktree.sh <project> <name>\n" {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
		}
	})
}
