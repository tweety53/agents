package guard

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-python-suppressions.sh at 75261910, one
// subtest per ok label. The harness ran the guard as a subprocess with path
// arguments pointing at fixture trees; here it runs in-process with the
// same arguments. Case 15 broke the no-argument enumeration with a bogus
// GIT_DIR in the process environment; in-process tests share that
// environment, so the case points FLOW_GUARD_REPO_ROOT at a directory that
// is no repository instead — the same failing `git ls-files`, the same
// status check.

// psRun runs the registered guard; out is stdout and stderr interleaved, as
// the harness's 2>&1 read them.
func psRun(t *testing.T, env Env, args ...string) (int, string, string, string) {
	t.Helper()
	fn := Registry["check-python-suppressions"]
	if fn == nil {
		t.Fatal("check-python-suppressions is not registered")
	}
	var out, stdout, stderr bytes.Buffer
	rc := fn(args, env, io.MultiWriter(&out, &stdout), io.MultiWriter(&out, &stderr))
	return rc, out.String(), stdout.String(), stderr.String()
}

func TestCheckPythonSuppressions(t *testing.T) {
	t.Parallel()

	// fixture writes files (path relative to a fresh dir: body) and runs the
	// guard on that dir.
	fixture := func(t *testing.T, files map[string]string) (string, int, string) {
		t.Helper()
		dir := t.TempDir()
		for rel, body := range files {
			writeFile(t, filepath.Join(dir, rel), body)
		}
		rc, out, _, _ := psRun(t, crEnv(dir, nil), dir)
		return dir, rc, out
	}
	exits := func(label string, want int, files map[string]string) {
		t.Run(label, func(t *testing.T) {
			t.Parallel()
			if _, rc, out := fixture(t, files); rc != want {
				t.Errorf("rc=%d, want %d\n%s", rc, want, out)
			}
		})
	}

	exits("case 1: clean file passes", 0, map[string]string{"clean.py": "import sys\n\nsys.exit(0)\n"})
	bare := map[string]string{"bare_noqa.py": "try:\n    risky()\nexcept Exception:  # noqa\n    pass\n"}
	exits("case 2: bare noqa caught", 1, bare)
	t.Run("case 2: hit names the file", func(t *testing.T) {
		t.Parallel()
		if _, _, out := fixture(t, bare); !strings.Contains(out, "bare_noqa.py") {
			t.Errorf("out=%s", out)
		}
	})
	exits("case 3: coded noqa caught", 1, map[string]string{"coded_noqa.py": "import os  # noqa: E402\n"})
	exits("case 4: type: ignore caught", 1, map[string]string{"type_ignore.py": "x: int = something()  # type: ignore\n"})
	exits("case 5: pylint disable caught", 1, map[string]string{"pylint_off.py": "def f():  # pylint: disable=missing-docstring\n    return 1\n"})
	exits("case 6: mypy ignore-errors caught", 1, map[string]string{"mypy_off.py": "# mypy: ignore-errors\n"})
	exits("case 7: nolint caught", 1, map[string]string{"nolint.py": "value = load()  # nolint\n"})
	exits("case 8: NOQA uppercase caught", 1, map[string]string{"upper_noqa.py": "import json  # NOQA\n"})
	exits("case 9: non-python file skipped", 0, map[string]string{"notes.md": "# noqa\n# type: ignore\n"})
	exits("case 10: nested directory scanned", 1, map[string]string{"pkg/deep/nested.py": "raise ValueError()  # noqa\n"})

	t.Run("case 11: missing path is exit 2", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if rc, out, _, _ := psRun(t, crEnv(dir, nil), filepath.Join(dir, "does-not-exist")); rc != 2 {
			t.Errorf("rc=%d\n%s", rc, out)
		}
	})

	exits("case 12: file-level flake8 noqa caught", 1, map[string]string{"filelevel.py": "# flake8: noqa\n"})

	symlinked := func(t *testing.T) (int, string) {
		t.Helper()
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "real.py"), "x = 1  # noqa\n")
		symlink(t, "real.py", filepath.Join(dir, "link.py"))
		rc, out, _, _ := psRun(t, crEnv(dir, nil), dir)
		return rc, out
	}
	t.Run("case 13: symlinked copy still finds the hit", func(t *testing.T) {
		t.Parallel()
		if rc, out := symlinked(t); rc != 1 {
			t.Errorf("rc=%d\n%s", rc, out)
		}
	})
	t.Run("case 13: hit reported exactly once", func(t *testing.T) {
		t.Parallel()
		if _, out := symlinked(t); strings.Count(out, "noqa") != 1 {
			t.Errorf("out=%s", out)
		}
	})

	t.Run("case 14: no-args scan independent of caller cwd", func(t *testing.T) {
		t.Parallel()
		repo := t.TempDir()
		gitRun(t, repo, "init", "-q")
		writeFile(t, filepath.Join(repo, "tool.py"), "x = 1\n")
		gitRun(t, repo, "add", "tool.py")
		if rc, out, _, _ := psRun(t, crEnv(t.TempDir(), map[string]string{"FLOW_GUARD_REPO_ROOT": repo})); rc == 2 {
			t.Errorf("rc=%d\n%s", rc, out)
		}
	})

	t.Run("case 15: git enumeration failure is exit 2", func(t *testing.T) {
		t.Parallel()
		notRepo := t.TempDir()
		if rc, out, _, _ := psRun(t, crEnv(notRepo, map[string]string{"FLOW_GUARD_REPO_ROOT": notRepo})); rc != 2 {
			t.Errorf("rc=%d\n%s", rc, out)
		}
	})

	t.Run("case 16: find enumeration failure is exit 2", func(t *testing.T) {
		t.Parallel()
		if os.Geteuid() == 0 {
			t.Skip("no chmod-unreadable directory exists under root")
		}
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "hidden", "secret.py"), "y = 2  # noqa\n")
		writeFile(t, filepath.Join(dir, "visible.py"), "z = 3\n")
		hidden := filepath.Join(dir, "hidden")
		if err := os.Chmod(hidden, 0); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(hidden, 0o755) })
		if rc, out, _, _ := psRun(t, crEnv(dir, nil), dir); rc != 2 {
			t.Errorf("rc=%d\n%s", rc, out)
		}
	})

	exits("case 17: pyright: ignore caught", 1, map[string]string{"pyright_ignore.py": "x: int = something()  # pyright: ignore\n"})
}

// TestCheckPythonSuppressionsOutput pins the exact stdout/stderr split and
// lines the bash printed at 75261910, captured from that script: hits as
// path:line:text on stdout in argument order, the count line on stderr, a
// relative argument printed as given, a binary file skipped as grep -I
// skips it, and the two clean verdicts.
func TestCheckPythonSuppressionsOutput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.py"), "ok = 1\nx = 1  # NoQA: E1\n")
	writeFile(t, filepath.Join(dir, "sub", "b.py"), "# type:ignore\ny = 2")
	writeFile(t, filepath.Join(dir, "bin.py"), "# noqa\x00\n")
	writeFile(t, filepath.Join(dir, "clean.py"), "z = 3\n")
	env := crEnv(dir, nil)
	for _, tc := range []struct {
		name           string
		args           []string
		rc             int
		stdout, stderr string
	}{
		{"hits", []string{"a.py", "sub/", "bin.py"}, 1,
			"a.py:2:x = 1  # NoQA: E1\nsub/b.py:1:# type:ignore\n",
			"suppressions-found: 2 marker hit(s) — fix the code, never widen this guard\n"},
		{"clean", []string{"clean.py"}, 0, "suppressions-ok: no linter suppression marker in the python in scope\n", ""},
		{"no python", []string{"sub/b.py.txt"}, 2, "", "ERROR: no such file or directory: sub/b.py.txt\n"},
		{"empty scope", []string{t.TempDir()}, 0, "suppressions-ok: no python file in scope\n", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rc, _, stdout, stderr := psRun(t, env, tc.args...)
			if rc != tc.rc || stdout != tc.stdout || stderr != tc.stderr {
				t.Errorf("rc=%d\nstdout=%q\nstderr=%q", rc, stdout, stderr)
			}
		})
	}
}
