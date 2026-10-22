package guard

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestRebuildFullAppSuiteZip carries the cases of the retired
// scripts/test-rebuild-full-app-suite-zip.sh. Every step runs twice, on twin
// projects: through the Go guard and through the bash at 8c54d6d3, whose
// resolve-visual-screenshots.sh sibling is the live shim. Each step asserts
// the harness's expectation and byte-for-byte stdout/stderr/exit parity, with
// each twin's root written as <P>.
func TestRebuildFullAppSuiteZip(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("zip"); err != nil {
		t.Skip("zip is not on PATH")
	}
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	bashDir := t.TempDir()
	writeExec(t, filepath.Join(bashDir, "rebuild-full-app-suite-zip.sh"),
		string(historicalFile(t, repo, "8c54d6d3", "scripts/rebuild-full-app-suite-zip.sh")))
	for _, l := range []string{"lib", "resolve-visual-screenshots.sh"} {
		if err := os.Symlink(filepath.Join(repo, "scripts", l), filepath.Join(bashDir, l)); err != nil {
			t.Fatal(err)
		}
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}

	type twin struct{ root string }
	newTwin := func(t *testing.T) twin {
		p := t.TempDir()
		writeFile(t, filepath.Join(p, ".flow/project.md"), "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `screenshots` | `shots` |\n")
		writeFile(t, filepath.Join(p, "shots/full-app-suite.spec.ts-snapshots/a.png"), "a")
		writeFile(t, filepath.Join(p, "shots/full-app-suite.spec.ts-snapshots/b.png"), "b")
		writeFile(t, filepath.Join(p, "shots/other.spec.ts-snapshots/c.png"), "c")
		return twin{p}
	}
	type res struct {
		rc          int
		out, errOut string
	}
	runGo := func(p string, args ...string) res {
		var o, e bytes.Buffer
		rc := Registry["rebuild-full-app-suite-zip"](args, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: p}, &o, &e)
		return res{rc, strings.ReplaceAll(o.String(), p, "<P>"), strings.ReplaceAll(e.String(), p, "<P>")}
	}
	runBash := func(t *testing.T, p string, args ...string) res {
		cmd := exec.Command("bash", append([]string{filepath.Join(bashDir, "rebuild-full-app-suite-zip.sh")}, args...)...)
		cmd.Dir, cmd.Env = p, environ
		var o, e bytes.Buffer
		cmd.Stdout, cmd.Stderr = &o, &e
		rc := 0
		if err := cmd.Run(); err != nil {
			var ee *exec.ExitError
			if !errors.As(err, &ee) {
				t.Fatal(err)
			}
			rc = ee.ExitCode()
		}
		return res{rc, strings.ReplaceAll(o.String(), p, "<P>"), strings.ReplaceAll(e.String(), p, "<P>")}
	}
	// rebuild runs one step on both twins: the Go result, after asserting parity.
	rebuild := func(t *testing.T, g, b twin, args func(p string) []string) res {
		t.Helper()
		gr, br := runGo(g.root, args(g.root)...), runBash(t, b.root, args(b.root)...)
		if gr != br {
			t.Fatalf("port %+v\nbash %+v", gr, br)
		}
		return gr
	}
	full := func(p string) []string { return []string{p, p + "/shots"} }
	entries := func(t *testing.T, p string) string {
		t.Helper()
		out, err := exec.Command("unzip", "-Z1", filepath.Join(p, "shots/full-app-suite.zip")).Output()
		if err != nil {
			t.Fatalf("unzip -Z1: %v", err)
		}
		names := strings.Fields(string(out))
		sort.Strings(names)
		return strings.Join(names, " ")
	}
	zipBytes := func(t *testing.T, p string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(p, "shots/full-app-suite.zip"))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	both := func(g, b twin, f func(p string)) { f(g.root); f(b.root) }

	t.Run("case 1 the suite's PNGs only", func(t *testing.T) {
		t.Parallel()
		g, b := newTwin(t), newTwin(t)
		if r := rebuild(t, g, b, full); r.rc != 0 || entries(t, g.root) != "a.png b.png" {
			t.Errorf("exit %d, entries %q", r.rc, entries(t, g.root))
		}
	})

	t.Run("case 2 byte-identical rebuild", func(t *testing.T) {
		t.Parallel()
		g, b := newTwin(t), newTwin(t)
		rebuild(t, g, b, full)
		first := zipBytes(t, g.root)
		both(g, b, func(p string) {
			now := time.Now().Add(-time.Hour) // far enough to leave zip's 2-second mtime granularity
			if err := os.Chtimes(filepath.Join(p, "shots/full-app-suite.spec.ts-snapshots/a.png"), now, now); err != nil {
				t.Fatal(err)
			}
		})
		rebuild(t, g, b, full)
		if !bytes.Equal(zipBytes(t, g.root), first) {
			t.Error("the rebuild differs")
		}
		if !bytes.Equal(zipBytes(t, g.root), zipBytes(t, b.root)) {
			t.Error("the Go zip differs from the bash zip")
		}
	})

	t.Run("case 3 a gone PNG leaves the zip", func(t *testing.T) {
		t.Parallel()
		g, b := newTwin(t), newTwin(t)
		rebuild(t, g, b, full)
		both(g, b, func(p string) { os.Remove(filepath.Join(p, "shots/full-app-suite.spec.ts-snapshots/b.png")) })
		rebuild(t, g, b, full)
		if got := entries(t, g.root); got != "a.png" {
			t.Errorf("entries %q", got)
		}
	})

	t.Run("case 4 nothing to zip", func(t *testing.T) {
		t.Parallel()
		g, b := newTwin(t), newTwin(t)
		rebuild(t, g, b, full)
		both(g, b, func(p string) {
			os.Remove(filepath.Join(p, "shots/full-app-suite.spec.ts-snapshots/a.png"))
			os.Remove(filepath.Join(p, "shots/full-app-suite.spec.ts-snapshots/b.png"))
		})
		r := rebuild(t, g, b, full)
		if _, err := os.Stat(filepath.Join(g.root, "shots/full-app-suite.zip")); r.rc != 1 || err == nil {
			t.Errorf("exit %d, zip stat err %v", r.rc, err)
		}
	})

	t.Run("case 5 usage", func(t *testing.T) {
		t.Parallel()
		g, b := newTwin(t), newTwin(t)
		if r := rebuild(t, g, b, func(p string) []string { return []string{p} }); r.rc != 2 {
			t.Errorf("exit %d", r.rc)
		}
	})

	t.Run("case 6 cannot answer keeps the zip", func(t *testing.T) {
		t.Parallel()
		g, b := newTwin(t), newTwin(t)
		rebuild(t, g, b, full)
		both(g, b, func(p string) { writeFile(t, filepath.Join(p, ".flow/project.md"), "# p\n") })
		r := rebuild(t, g, b, full)
		if _, err := os.Stat(filepath.Join(g.root, "shots/full-app-suite.zip")); r.rc != 2 || err != nil {
			t.Errorf("exit %d, zip stat err %v", r.rc, err)
		}
	})
}
