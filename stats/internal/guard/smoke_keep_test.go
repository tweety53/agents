package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// smokeKeepRepo builds a repository whose base commit holds files and is origin/main.
func smokeKeepRepo(t *testing.T, g *guardtest.FxGit, files ...string) string {
	t.Helper()
	repo := t.TempDir() + "/r"
	g.CopyShape(repo, guardtest.RepoShape{Branch: "main"})
	for _, f := range files {
		if g.Err == nil {
			g.Err = os.MkdirAll(filepath.Dir(repo+"/"+f), 0o755)
		}
		g.Write(repo+"/"+f, f)
	}
	g.Git(repo, "add", ".")
	g.Git(repo, "commit", "-q", "-m", "base")
	g.Git(repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	return repo
}

func TestSmokeKeep(t *testing.T) {
	t.Parallel()
	var g guardtest.FxGit
	repo := smokeKeepRepo(t, &g,
		"pkg/integratechange_land.go", "pkg/integrate_change_test.go",
		"pkg/planshape.go", "pkg/check_plan_shape_test.go", "pkg/run.go",
		"pkg/run_verify_list_test.go", "pkg/exec.go", "pkg/quiet.go", "pkg/quiet_test.go",
		"pkg/_.go", "pkg/old.go", "pkg/old_test.go", "other/planshape_test.go", "hooks/a.py", "setup.sh", "docs/a.md")
	g.Write(repo+"/pkg/integratechange_land.go", "changed")
	g.Write(repo+"/hooks/a.py", "changed")
	g.Git(repo, "mv", "pkg/old.go", "pkg/moved.go")
	g.Git(repo, "commit", "-q", "-am", "change")
	for _, f := range []string{"pkg/a b_test.go", "pkg/_.go", "pkg/planshape.go", "pkg/run.go", "pkg/exec.go", "setup.sh", "docs/a.md", "pkg/new_test.go"} {
		g.Write(repo+"/"+f, "edited")
	}
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	env := core.Env{Getenv: os.Getenv, Dir: repo}
	// A rename keeps its old name's tests; a path holding a space stays one path.
	const tests = "pkg/ pkg/a b_test.go pkg/check_plan_shape_test.go pkg/integrate_change_test.go pkg/new_test.go pkg/old_test.go pkg/run_verify_list_test.go"
	// pkg/_.go has an empty normal form, so it pairs with nothing and keeps pkg/.
	const unpaired = "smoke-keep: pkg/_.go pairs with no test file — keeping its directory pkg/\n" +
		"smoke-keep: pkg/exec.go pairs with no test file — keeping its directory pkg/\n" +
		"smoke-keep: pkg/moved.go pairs with no test file — keeping its directory pkg/\n"

	r := runGuard("smoke-keep", []string{"hooks/=t/repotest/", "hooks/=t/setuptest/", "setup.sh=t/setuptest/", "a.py=t/mid/"}, env)
	if want := tests + " t/repotest/ t/setuptest/\n"; r.Rc != 0 || r.Stdout != want || r.Err != unpaired {
		t.Errorf("with keep rows: rc=%d stdout=%q stderr=%q, want 0, %q and %q", r.Rc, r.Stdout, r.Err, want, unpaired)
	}
	r = runGuard("smoke-keep", nil, env)
	if want := tests + "\n"; r.Rc != 0 || r.Stdout != want || r.Err != unpaired {
		t.Errorf("no argument: rc=%d stdout=%q stderr=%q, want 0, %q and %q", r.Rc, r.Stdout, r.Err, want, unpaired)
	}
	r = runGuard("smoke-keep", nil, core.Env{Getenv: os.Getenv, Dir: repo + "/pkg"})
	if want := tests + "\n"; r.Rc != 0 || r.Stdout != want || r.Err != unpaired {
		t.Errorf("from a subdirectory: rc=%d stdout=%q stderr=%q, want 0, %q and %q", r.Rc, r.Stdout, r.Err, want, unpaired)
	}

	var g2 guardtest.FxGit
	clean := smokeKeepRepo(t, &g2, "pkg/run.go", "pkg/run_test.go")
	if g2.Err != nil {
		t.Fatal(g2.Err)
	}
	r = runGuard("smoke-keep", []string{"pkg/=t/x/"}, core.Env{Getenv: os.Getenv, Dir: clean})
	if r.Rc != 0 || r.Stdout != "\n" || r.Err != "" {
		t.Errorf("nothing changed: rc=%d stdout=%q stderr=%q, want 0 and one empty line", r.Rc, r.Stdout, r.Err)
	}
}

func TestSmokeKeepUnpairedKeepsItsDirectory(t *testing.T) {
	t.Parallel()
	var g guardtest.FxGit
	repo := smokeKeepRepo(t, &g, "a/b/x.go", "a/b/other_test.go")
	g.Write(repo+"/a/b/x.go", "edited")
	g.Git(repo, "commit", "-q", "-am", "change")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	r := runGuard("smoke-keep", nil, core.Env{Getenv: os.Getenv, Dir: repo})
	const wantErr = "smoke-keep: a/b/x.go pairs with no test file — keeping its directory a/b/\n"
	if r.Rc != 0 || r.Stdout != "a/b/\n" || r.Err != wantErr {
		t.Errorf("rc=%d stdout=%q stderr=%q, want 0, %q and %q", r.Rc, r.Stdout, r.Err, "a/b/\n", wantErr)
	}
}

func TestSmokeKeepUnreadableDirectory(t *testing.T) {
	t.Parallel()
	var g guardtest.FxGit
	repo := smokeKeepRepo(t, &g, "pkg/run.go", "pkg/run_test.go")
	g.Write(repo+"/pkg/run.go", "edited")
	if g.Err == nil {
		g.Err = os.Chmod(repo+"/pkg", 0o300)
	}
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	t.Cleanup(func() { _ = os.Chmod(repo+"/pkg", 0o755) })
	r := runGuard("smoke-keep", nil, core.Env{Getenv: os.Getenv, Dir: repo})
	if r.Rc != 0 || r.Stdout != "\n" || !strings.HasPrefix(r.Err, "smoke-keep: pkg/run.go: open ") || !strings.Contains(r.Err, "permission denied") {
		t.Errorf("unreadable directory: rc=%d stdout=%q stderr=%q, want 0, an empty line and the read error", r.Rc, r.Stdout, r.Err)
	}
}

func TestSmokeKeepRefusals(t *testing.T) {
	t.Parallel()
	var g guardtest.FxGit
	repo := smokeKeepRepo(t, &g, "a.go")
	noOrigin := t.TempDir() + "/n"
	g.CopyShape(noOrigin, guardtest.RepoShape{Branch: "main", Commit: "base"})
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	for _, arg := range []string{"hooks/", "=t/x/", "hooks/="} {
		r := runGuard("smoke-keep", []string{arg}, core.Env{Getenv: os.Getenv, Dir: repo})
		if r.Rc != 2 || r.Err != "usage: flow-guard smoke-keep [<path prefix>=<kept path>]...\n" || r.Stdout != "" {
			t.Errorf("%q: rc=%d stderr=%q stdout=%q, want 2 and the usage line", arg, r.Rc, r.Err, r.Stdout)
		}
	}
	for name, dir := range map[string]string{"no repository": t.TempDir(), "no origin/main": noOrigin} {
		r := runGuard("smoke-keep", nil, core.Env{Getenv: os.Getenv, Dir: dir})
		if r.Rc != 2 || !strings.HasPrefix(r.Err, "smoke-keep: ") || r.Stdout != "" {
			t.Errorf("%s: rc=%d stderr=%q stdout=%q, want 2 and a smoke-keep: line", name, r.Rc, r.Err, r.Stdout)
		}
	}
}
