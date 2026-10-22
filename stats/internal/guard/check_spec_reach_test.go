package guard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-spec-reach.sh at b1b0d601, one subtest per
// case label. As in the harness, Playwright is not installed: a fake `npm`
// first on the PATH the guard's Env returns answers `npm run -s <script> --
// --list --reporter=json` with the fixture's `.list-<script>.json` (a
// `.fail-<script>` file makes it exit 7), and `node` is real.

const srFakeNpm = `#!/usr/bin/env bash
# fake npm: ` + "`npm run -s <script> -- --list --reporter=json`" + `, answered from cwd.
script="$3"
if [ -e ".fail-$script" ]; then echo "playwright: boom for $script" >&2; exit 7; fi
cat ".list-$script.json"
`

type srFx struct {
	t              *testing.T
	root, checkout string
	path           string
}

// srNew is new_root: a project root with .flow/ and a regression checkout.
func srNew(t *testing.T, path string) *srFx {
	t.Helper()
	fx := &srFx{t: t, root: t.TempDir(), checkout: t.TempDir(), path: path}
	if err := os.MkdirAll(filepath.Join(fx.root, ".flow"), 0o755); err != nil {
		t.Fatal(err)
	}
	return fx
}

func (fx *srFx) write(p, body string) {
	fx.t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		fx.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		fx.t.Fatal(err)
	}
}

func (fx *srFx) cfg(body string) { fx.write(filepath.Join(fx.root, ".flow/project.md"), body) }

// cfgWithCheckout is cfg_with_checkout: a minimal section naming the checkout.
func (fx *srFx) cfgWithCheckout(prefix string) {
	fx.cfg(fmt.Sprintf("%s## visual verification\n\n| Setting | Value |\n|---|---|\n| `screenshots` | `.` |\n| `regression checkout` | `%s` |\n", prefix, fx.checkout))
}

// packageJSON is package_json <script>=<command> ...
func (fx *srFx) packageJSON(kvs ...string) {
	var entries []string
	for _, kv := range kvs {
		k, v, _ := strings.Cut(kv, "=")
		entries = append(entries, fmt.Sprintf("%q: %q", k, v))
	}
	fx.write(filepath.Join(fx.checkout, "package.json"), "{ \"scripts\": { "+strings.Join(entries, ", ")+" } }\n")
}

// listing is listing <script> <file>...: canned reporter JSON for one script.
func (fx *srFx) listing(script string, files ...string) {
	var suites []string
	for _, f := range files {
		suites = append(suites, fmt.Sprintf(`{"file": "%s", "suites": []}`, f))
	}
	fx.write(filepath.Join(fx.checkout, ".list-"+script+".json"), fmt.Sprintf(`{ "config": { "rootDir": "%s" }, "suites": [ %s ] }`+"\n", fx.checkout, strings.Join(suites, ", ")))
}

func (fx *srFx) spec(rel string) { fx.write(filepath.Join(fx.checkout, rel), "") }

func (fx *srFx) run(args ...string) guardResult {
	if args == nil {
		args = []string{fx.root}
	}
	return runGuard("check-spec-reach", args, Env{Dir: "/", Getenv: func(k string) string {
		if k == "PATH" {
			return fx.path
		}
		return os.Getenv(k)
	}})
}

func srAssert(t *testing.T, r guardResult, rc int, wants ...string) {
	t.Helper()
	if r.rc != rc {
		t.Fatalf("rc = %d, want %d; out:\n%s", r.rc, rc, r.out)
	}
	for _, w := range wants {
		if !strings.Contains(r.out, w) {
			t.Errorf("output lacks %q:\n%s", w, r.out)
		}
	}
}

func srGit(t *testing.T, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.invalid"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestCheckSpecReach(t *testing.T) {
	t.Parallel()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "npm"), []byte(srFakeNpm), 0o755); err != nil {
		t.Fatal(err)
	}
	path := bin + ":" + os.Getenv("PATH")

	t.Run("case 1", func(t *testing.T) {
		fx := srNew(t, path)
		srAssert(t, fx.run(), 0, "Spec reach: not configured")
	})
	t.Run("case 2", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfg("# Project\n\n## run\n\necho hi\n")
		srAssert(t, fx.run(), 0, "Spec reach: not configured")
	})
	t.Run("case 3", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfg("## visual verification\n\n| Setting | Value |\n|---|---|\n| `screenshots` | `.` |\n")
		srAssert(t, fx.run(), 0, "Spec reach: not configured")
	})
	t.Run("case 4", func(t *testing.T) {
		fx := srNew(t, path)
		fx.write(filepath.Join(fx.root, ".flow/project.md"), fmt.Sprintf("## visual verification\n\n| Setting | Value |\n|---|---|\n| `screenshots` | `.` |\n| `regression checkout` | `%s` |\n\n## visual verification\n", fx.checkout))
		srAssert(t, fx.run(), 2, "ambiguous")
	})
	t.Run("case 5", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		if err := os.RemoveAll(fx.checkout); err != nil {
			t.Fatal(err)
		}
		srAssert(t, fx.run(), 2, "not an existing directory")
	})
	t.Run("case 6", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		srAssert(t, fx.run(), 2, "no package.json")
	})
	t.Run("case 7", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		fx.packageJSON("test=vitest run")
		srAssert(t, fx.run(), 2, "playwright test")
	})
	t.Run("case 8", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		fx.packageJSON("test:e2e=playwright test tests/", "shots=playwright test shots.spec.ts --headed", "lint=eslint .")
		fx.spec("tests/a.spec.ts")
		fx.spec("shots.spec.ts")
		fx.spec("node_modules/dep/x.spec.ts")
		fx.listing("test:e2e", "tests/a.spec.ts")
		fx.listing("shots", "shots.spec.ts")
		srAssert(t, fx.run(), 0, "Spec reach: 2 spec(s), all reached")
	})
	t.Run("case 9", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		fx.packageJSON("test:e2e=playwright test tests/")
		fx.spec("tests/a.spec.ts")
		fx.spec("kan-347-route-e.spec.ts")
		fx.listing("test:e2e", "tests/a.spec.ts")
		r := fx.run()
		srAssert(t, r, 1, "kan-347-route-e.spec.ts: reached by no package.json script", "1 of 2 spec(s)")
		if strings.Contains(r.out, "tests/a.spec.ts:") {
			t.Errorf("a reached spec was reported:\n%s", r.out)
		}
	})
	t.Run("case 10", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		fx.packageJSON("test:e2e=playwright test tests/")
		fx.spec("tests/a.spec.ts")
		fx.spec(".fail-test:e2e")
		srAssert(t, fx.run(), 2, "boom for test:e2e")
	})
	t.Run("case 11", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		fx.packageJSON("test:e2e=playwright test tests/")
		fx.spec("tests/a.spec.ts")
		fx.write(filepath.Join(fx.checkout, ".list-test:e2e.json"), "not json\n")
		srAssert(t, fx.run(), 2, "no parsable JSON")
	})
	t.Run("case 12", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("\xef\xbb\xbf")
		fx.packageJSON("test:e2e=playwright test tests/")
		fx.spec("tests/a.spec.ts")
		fx.listing("test:e2e", "tests/a.spec.ts")
		srAssert(t, fx.run(), 0, "all reached")
	})
	t.Run("case 13", func(t *testing.T) {
		fx := srNew(t, path)
		srAssert(t, fx.run(fmt.Sprintf("/no/such/root-%d", os.Getpid())), 2, "is not a directory")
	})
	t.Run("case 14", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		fx.write(filepath.Join(fx.checkout, "package.json"), "{ not valid json")
		fx.spec("tests/a.spec.ts")
		srAssert(t, fx.run(), 2, "could not read scripts from")
	})
	t.Run("case 15", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfg(fmt.Sprintf("## visual verification\n\n| Setting | Value |\n|---|---|\n| `screenshots` | `.` |\n\n## unrelated section\n\n| Setting | Value |\n|---|---|\n| `regression checkout` | `/definitely/not/real/path-%d` |\n", os.Getpid()))
		srAssert(t, fx.run(), 0, "Spec reach: not configured")
	})
	t.Run("case 16", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfg(fmt.Sprintf("## visual verification\n\n| Setting | Value |\n|---|---|\n| `regression checkout` | `%s` |\n| `regression checkout` | `%s` |\n", fx.checkout, t.TempDir()))
		fx.packageJSON("test:e2e=playwright test tests/")
		fx.spec("tests/a.spec.ts")
		fx.listing("test:e2e", "tests/a.spec.ts")
		srAssert(t, fx.run(), 0, "all reached")
	})
	t.Run("case 17", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		fx.packageJSON("real=playwright test real/", "decoy=playwright test decoy/")
		fx.spec("real/a.spec.ts")
		fx.listing("real")
		fx.write(filepath.Join(fx.checkout, ".list-decoy.json"), fmt.Sprintf(`{ "config": { "rootDir": "%s" }, "suites": [ {"file": "/zzz%s/real/a.spec.ts", "suites": []} ] }`+"\n", fx.checkout, fx.checkout))
		srAssert(t, fx.run(), 1, "real/a.spec.ts: reached by no package.json script")
	})
	t.Run("case 18", func(t *testing.T) {
		fx := srNew(t, path)
		fx.cfgWithCheckout("")
		srGit(t, "init", "-q", "-b", "main", fx.checkout)
		srGit(t, "-C", fx.checkout, "commit", "-q", "--allow-empty", "-m", "base")
		srGit(t, "init", "-q", "-b", "main", fx.root)
		srGit(t, "-C", fx.root, "commit", "-q", "--allow-empty", "-m", "base")
		srGit(t, "-C", fx.root, "checkout", "-q", "-b", "spectre/demo")
		fx.packageJSON("test:e2e=playwright test tests/")
		fx.spec("tests/a.spec.ts")
		fx.spec("orphan.spec.ts")
		fx.listing("test:e2e", "tests/a.spec.ts")
		wt := filepath.Join(t.TempDir(), "wt")
		srGit(t, "-C", fx.checkout, "worktree", "add", "-q", "-b", "spectre/demo", wt)
		// The path git's porcelain lists, symlinks resolved, as the harness read it.
		real, err := filepath.EvalSymlinks(wt)
		if err != nil {
			t.Fatal(err)
		}
		fx.checkout = real
		fx.packageJSON("test:e2e=playwright test tests/")
		fx.spec("tests/a.spec.ts")
		fx.listing("test:e2e", "tests/a.spec.ts")
		srAssert(t, fx.run(), 0, "Spec reach: 1 spec(s), all reached")
	})
}
