package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// smokeKeepRepo builds a repository whose base commit holds files and is origin/main.
func smokeKeepRepo(t *testing.T, g *fxGit, files ...string) string {
	t.Helper()
	repo := t.TempDir() + "/r"
	g.git("", "init", "-q", "-b", "main", repo)
	for _, f := range files {
		if g.err == nil {
			g.err = os.MkdirAll(filepath.Dir(repo+"/"+f), 0o755)
		}
		g.write(repo+"/"+f, f)
	}
	g.git(repo, "add", ".")
	g.git(repo, "commit", "-q", "-m", "base")
	g.git(repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	return repo
}

func TestSmokeKeep(t *testing.T) {
	t.Parallel()
	var g fxGit
	repo := smokeKeepRepo(t, &g,
		"pkg/integratechange_land.go", "pkg/integrate_change_test.go",
		"pkg/planshape.go", "pkg/check_plan_shape_test.go", "pkg/run.go",
		"pkg/run_verify_list_test.go", "pkg/exec.go", "pkg/quiet.go", "pkg/quiet_test.go",
		"other/planshape_test.go", "hooks/a.py", "setup.sh", "docs/a.md")
	g.write(repo+"/pkg/integratechange_land.go", "changed")
	g.write(repo+"/hooks/a.py", "changed")
	g.git(repo, "commit", "-q", "-am", "change")
	for _, f := range []string{"pkg/planshape.go", "pkg/run.go", "pkg/exec.go", "setup.sh", "docs/a.md", "pkg/new_test.go"} {
		g.write(repo+"/"+f, "edited")
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	env := Env{Getenv: os.Getenv, Dir: repo}
	const tests = "pkg/check_plan_shape_test.go pkg/integrate_change_test.go pkg/new_test.go pkg/run_verify_list_test.go"
	const unpaired = "smoke-keep: pkg/exec.go pairs with no test file — nothing kept for it\n"

	r := runGuard("smoke-keep", []string{"hooks/=t/repotest/", "hooks/=t/setuptest/", "setup.sh=t/setuptest/"}, env)
	if want := tests + " t/repotest/ t/setuptest/\n"; r.rc != 0 || r.stdout != want || r.err != unpaired {
		t.Errorf("with keep rows: rc=%d stdout=%q stderr=%q, want 0, %q and %q", r.rc, r.stdout, r.err, want, unpaired)
	}
	r = runGuard("smoke-keep", nil, env)
	if want := tests + "\n"; r.rc != 0 || r.stdout != want || r.err != unpaired {
		t.Errorf("no argument: rc=%d stdout=%q stderr=%q, want 0, %q and %q", r.rc, r.stdout, r.err, want, unpaired)
	}

	var g2 fxGit
	clean := smokeKeepRepo(t, &g2, "pkg/run.go", "pkg/run_test.go")
	if g2.err != nil {
		t.Fatal(g2.err)
	}
	r = runGuard("smoke-keep", []string{"pkg/=t/x/"}, Env{Getenv: os.Getenv, Dir: clean})
	if r.rc != 0 || r.stdout != "\n" || r.err != "" {
		t.Errorf("nothing changed: rc=%d stdout=%q stderr=%q, want 0 and one empty line", r.rc, r.stdout, r.err)
	}
}

func TestSmokeKeepRefusals(t *testing.T) {
	t.Parallel()
	var g fxGit
	repo := smokeKeepRepo(t, &g, "a.go")
	noOrigin := t.TempDir() + "/n"
	g.git("", "init", "-q", "-b", "main", noOrigin)
	g.git(noOrigin, "commit", "-q", "--allow-empty", "-m", "base")
	if g.err != nil {
		t.Fatal(g.err)
	}
	for _, arg := range []string{"hooks/", "=t/x/", "hooks/="} {
		r := runGuard("smoke-keep", []string{arg}, Env{Getenv: os.Getenv, Dir: repo})
		if r.rc != 2 || r.err != "usage: flow-guard smoke-keep [<path prefix>=<kept path>]...\n" || r.stdout != "" {
			t.Errorf("%q: rc=%d stderr=%q stdout=%q, want 2 and the usage line", arg, r.rc, r.err, r.stdout)
		}
	}
	for name, dir := range map[string]string{"no repository": t.TempDir(), "no origin/main": noOrigin} {
		r := runGuard("smoke-keep", nil, Env{Getenv: os.Getenv, Dir: dir})
		if r.rc != 2 || !strings.HasPrefix(r.err, "smoke-keep: ") || r.stdout != "" {
			t.Errorf("%s: rc=%d stderr=%q stdout=%q, want 2 and a smoke-keep: line", name, r.rc, r.err, r.stdout)
		}
	}
}
