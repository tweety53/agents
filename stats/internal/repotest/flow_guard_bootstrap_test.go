package repotest

import (
	"errors"
	"fmt"
	"github.com/tweety53/agents/stats/internal/smoke"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// The bootstrap scripts/flow-guard is run for real — through bash, against a
// real source tree and a real `go build` — since the defect it closes is a
// binary that is not the checkout's source, which only a real build shows.
// Two kinds of tree are used: a copy of this repository's flow-guard build
// sources (the real guards), and a fixture whose tiny flow-guard prints a
// version word and its arguments, so a rebuild after a source edit is visible
// in the guard's own output. Every case has its own FLOW_GUARD_CACHE_DIR: none
// writes into the operator's real cache.

// noGoPath is a PATH holding bash and the coreutils but no go.
const noGoPath = "/usr/bin:/bin"

// bootstrapRun is one bootstrap call's outcome.
type bootstrapRun struct {
	stdout, stderr string
	rc             int
}

func (r bootstrapRun) out() string { return strings.TrimSpace(r.stdout + r.stderr) }

// runBootstrap runs path (a scripts/flow-guard, or a link to one) from dir
// with env as its whole environment.
func runBootstrap(t *testing.T, dir string, env []string, path string, args ...string) bootstrapRun {
	t.Helper()
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	cmd.Env = env
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	r := bootstrapRun{stdout: out.String(), stderr: errb.String()}
	var ee *exec.ExitError
	switch {
	case errors.As(err, &ee):
		r.rc = ee.ExitCode()
	case err != nil:
		t.Fatalf("run %s: %v", path, err)
	}
	return r
}

// bootstrapEnv is the test process's environment without anything that would
// steer the bootstrap, plus the Go caches, a dead flow store and cache as the
// FLOW_GUARD_CACHE_DIR ("" sets none); extra entries are appended last.
func bootstrapEnv(cache string, extra ...string) []string {
	var env []string
	for _, kv := range os.Environ() {
		switch strings.SplitN(kv, "=", 2)[0] {
		case "FLOW_GUARD_CACHE_DIR", "FLOW_GUARD_REPO_ROOT", "GOOS", "GOARCH", "GOFLAGS", "PATH":
			continue
		}
		env = append(env, kv)
	}
	env = append(append(env, goCacheEnv()...), "FLOW_GUARD_TELEMETRY=off", "FLOW_ADDR=http://127.0.0.1:1", "FLOW_RECORDS_ADDR=http://127.0.0.1:1")
	if cache != "" {
		env = append(env, "FLOW_GUARD_CACHE_DIR="+cache)
	}
	return append(env, extra...)
}

// withPath is env carrying path as its PATH.
func withPath(env []string, path string) []string { return append(env, "PATH="+path) }

// writeFile writes body to path, creating its directory.
func writeFile(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

// appendFile appends line to path.
func appendFile(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		t.Fatal(err)
	}
}

// copyBootstrap puts this repository's scripts/flow-guard at <root>/scripts/flow-guard.
func copyBootstrap(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "scripts", "flow-guard"))
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(root, "scripts", "flow-guard")
	writeFile(t, p, string(b), 0o755)
	return p
}

// newBootstrapFixture is a fixture checkout whose flow-guard prints "<word> <args…>".
func newBootstrapFixture(t *testing.T, word string) string {
	t.Helper()
	fix := t.TempDir()
	copyBootstrap(t, fix)
	writeFile(t, filepath.Join(fix, "stats/go.mod"), "module fixture\n\ngo 1.21\n", 0o644)
	writeFile(t, filepath.Join(fix, "stats/go.sum"), "", 0o644)
	writeFile(t, filepath.Join(fix, "stats/internal/guard/guard.go"), "package guard\n", 0o644)
	writeFile(t, filepath.Join(fix, "stats/internal/guard/guard_test.go"), "package guard\n", 0o644)
	writeFixtureMain(t, fix, word)
	return fix
}

// writeFixtureMain makes the fixture's flow-guard print "<word> <args…>".
func writeFixtureMain(t *testing.T, fix, word string) {
	t.Helper()
	writeFile(t, filepath.Join(fix, "stats/cmd/flow-guard/main.go"), fmt.Sprintf(
		"package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\t\"strings\"\n)\n\nfunc main() { fmt.Println(%q, strings.Join(os.Args[1:], \" \")) }\n", word), 0o644)
}

// copyRealSources copies this repository's flow-guard build sources — the
// files the bootstrap keys — and its bootstrap into a temp checkout.
func copyRealSources(t *testing.T) string {
	t.Helper()
	dst := t.TempDir()
	src := filepath.Join(repoRoot(t), "stats")
	files := []string{"go.mod", "go.sum"}
	for _, dir := range []string{"cmd/flow-guard", "internal/guard", "internal/records", "internal/modelpolicy", "internal/prompttuning"} {
		matches, err := filepath.Glob(filepath.Join(src, dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range matches {
			if !strings.HasSuffix(m, "_test.go") {
				files = append(files, filepath.Join(dir, filepath.Base(m)))
			}
		}
	}
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(src, f))
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dst, "stats", f), string(b), 0o644)
	}
	copyBootstrap(t, dst)
	return dst
}

// cacheKeys lists the key directories under cache.
func cacheKeys(t *testing.T, cache string) []string {
	t.Helper()
	entries, err := os.ReadDir(cache)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	var keys []string
	for _, e := range entries {
		if e.IsDir() {
			keys = append(keys, e.Name())
		}
	}
	sort.Strings(keys)
	return keys
}

// cachedBinaries lists every flow-guard binary under cache.
func cachedBinaries(t *testing.T, cache string) []string {
	t.Helper()
	bins, err := filepath.Glob(filepath.Join(cache, "*", "flow-guard"))
	if err != nil {
		t.Fatal(err)
	}
	return bins
}

func goBinary(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go is not on PATH — every case needs a real build")
	}
	return p
}

// TestFlowGuardBootstrap ports scripts/test-lib-flow-guard.sh's cases 1–14
// and 5b to scripts/flow-guard.
func TestFlowGuardBootstrap(t *testing.T) {
	smoke.Heavy(t)
	goBin := goBinary(t)
	path := os.Getenv("PATH")

	// Cases 1–3 and 14 share one copy of the real sources and one cache:
	// case 1 builds it, the rest answer from it with no go on PATH.
	realTree := copyRealSources(t)
	realBoot := filepath.Join(realTree, "scripts", "flow-guard")
	realEnv := bootstrapEnv(t.TempDir())

	t.Run("case 1: a flow-guard on PATH is ignored; the checkout-built guard answers", func(t *testing.T) {
		stale := t.TempDir()
		writeFile(t, filepath.Join(stale, "flow-guard"), "#!/bin/sh\necho STALE-BINARY-ANSWERED\nexit 0\n", 0o755)
		r := runBootstrap(t, realTree, withPath(realEnv, stale+":"+path), realBoot, "check-task-commit-fields")
		if r.rc != 2 || strings.Contains(r.out(), "STALE-BINARY-ANSWERED") || !strings.Contains(r.out(), "usage") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out())
		}
	})
	t.Run("case 2: a cached flow-guard answers with no go on PATH", func(t *testing.T) {
		r := runBootstrap(t, realTree, withPath(realEnv, noGoPath), realBoot, "check-task-commit-fields")
		if r.rc != 2 || !strings.Contains(r.out(), "usage") || strings.Contains(r.out(), "cannot build flow-guard") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out())
		}
	})
	t.Run("case 3: reached through a symlink, the bootstrap runs the checkout's cached flow-guard", func(t *testing.T) {
		link := filepath.Join(t.TempDir(), "flow-guard")
		if err := os.Symlink(realBoot, link); err != nil {
			t.Fatal(err)
		}
		r := runBootstrap(t, realTree, withPath(realEnv, noGoPath), link, "check-cleanup-complete")
		if r.rc != 2 || !strings.Contains(r.out(), "usage") || strings.Contains(r.out(), "cannot build flow-guard") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out())
		}
	})
	t.Run("case 14: ./flow-guard run from scripts/ answers every real guard from the cache", func(t *testing.T) {
		for _, name := range []string{"check-cleanup-complete", "check-panel-reproducer-exit-contract",
			"check-task-commit-fields", "gather-dispatch-context", "run-reproducer"} {
			r := runBootstrap(t, filepath.Join(realTree, "scripts"), withPath(realEnv, noGoPath), "./flow-guard", name)
			out := r.out()
			if !(strings.HasPrefix(out, name+": ") || strings.HasPrefix(out, "usage:")) || strings.Contains(out, "cannot build flow-guard") {
				t.Errorf("%s: rc=%d out=%s", name, r.rc, out)
			}
		}
	})

	t.Run("case 4: a source edit rebuilds; the edited source answers, not the cached binary", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v1")
		env := withPath(bootstrapEnv(t.TempDir()), path)
		boot := filepath.Join(fix, "scripts", "flow-guard")
		if r := runBootstrap(t, fix, env, boot, "check-x"); r.rc != 0 || r.out() != "v1 check-x" {
			t.Fatalf("4a: rc=%d out=%s", r.rc, r.out())
		}
		writeFixtureMain(t, fix, "v2")
		if r := runBootstrap(t, fix, env, boot, "check-x"); r.rc != 0 || r.out() != "v2 check-x" {
			t.Fatalf("4b: rc=%d out=%s", r.rc, r.out())
		}
	})

	// keyAfter runs the fixture's bootstrap and returns the cache's key set.
	keyAfter := func(t *testing.T, fix string, env []string, cache string) []string {
		t.Helper()
		if r := runBootstrap(t, fix, env, filepath.Join(fix, "scripts", "flow-guard"), "check-x"); r.rc != 0 {
			t.Fatalf("rc=%d out=%s", r.rc, r.out())
		}
		return cacheKeys(t, cache)
	}

	t.Run("case 5: the cache key ignores _test.go and tracks internal/guard and go.mod", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v5")
		cache := t.TempDir()
		env := withPath(bootstrapEnv(cache), path)
		k0 := keyAfter(t, fix, env, cache)
		appendFile(t, filepath.Join(fix, "stats/internal/guard/guard_test.go"), "// edit\n")
		k1 := keyAfter(t, fix, env, cache)
		appendFile(t, filepath.Join(fix, "stats/internal/guard/guard.go"), "// edit\n")
		k2 := keyAfter(t, fix, env, cache)
		appendFile(t, filepath.Join(fix, "stats/go.mod"), "// edit\n")
		k3 := keyAfter(t, fix, env, cache)
		if len(k0) != 1 || len(k1) != 1 || len(k2) != 2 || len(k3) != 3 {
			t.Fatalf("key sets after each edit: %v %v %v %v", k0, k1, k2, k3)
		}
	})
	t.Run("case 5b: the cache key tracks the non-test files of every imported package", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v5b")
		cache := t.TempDir()
		env := withPath(bootstrapEnv(cache), path)
		want := len(keyAfter(t, fix, env, cache))
		for _, pkg := range []string{"records", "modelpolicy", "prompttuning"} {
			dir := filepath.Join(fix, "stats/internal", pkg)
			writeFile(t, filepath.Join(dir, "x_test.go"), "package "+pkg+"\n", 0o644)
			if got := len(keyAfter(t, fix, env, cache)); got != want {
				t.Fatalf("%s: a _test.go changed the key (%d keys, want %d)", pkg, got, want)
			}
			writeFile(t, filepath.Join(dir, "x.go"), "package "+pkg+"\n", 0o644)
			want++
			if got := len(keyAfter(t, fix, env, cache)); got != want {
				t.Fatalf("%s: adding a non-test file kept the key (%d keys, want %d)", pkg, got, want)
			}
			appendFile(t, filepath.Join(dir, "x.go"), "// edit\n")
			want++
			if got := len(keyAfter(t, fix, env, cache)); got != want {
				t.Fatalf("%s: editing a non-test file kept the key (%d keys, want %d)", pkg, got, want)
			}
		}
	})
	t.Run("case 6: no go and no cached binary exits the cannot-answer code naming the missing go", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v6")
		r := runBootstrap(t, fix, withPath(bootstrapEnv(t.TempDir()), noGoPath), filepath.Join(fix, "scripts", "flow-guard"), "run-reproducer")
		if r.rc != 4 || !strings.HasPrefix(r.out(), "run-reproducer: cannot build flow-guard from ") || !strings.HasSuffix(r.out(), "no go on PATH") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out())
		}
	})
	t.Run("case 7: a failed build exits the cannot-answer code, shows go's error on stderr, caches nothing", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v7")
		writeFile(t, filepath.Join(fix, "stats/cmd/flow-guard/main.go"), "package main\n\nfunc main() { undefined() }\n", 0o644)
		cache := t.TempDir()
		r := runBootstrap(t, fix, withPath(bootstrapEnv(cache), path), filepath.Join(fix, "scripts", "flow-guard"), "run-reproducer")
		if r.rc != 4 || r.stdout != "" || !strings.Contains(r.stderr, "undefined") ||
			!strings.Contains(r.stderr, "run-reproducer: could not build flow-guard from") || len(cachedBinaries(t, cache)) != 0 {
			t.Fatalf("rc=%d stdout=%q stderr=%s cached=%v", r.rc, r.stdout, r.stderr, cachedBinaries(t, cache))
		}
	})
	t.Run("case 8: with no FLOW_GUARD_CACHE_DIR, XDG_CACHE_HOME or HOME the bootstrap exits the cannot-answer code", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v8")
		r := runBootstrap(t, fix, []string{"PATH=" + path}, filepath.Join(fix, "scripts", "flow-guard"), "run-reproducer")
		if r.rc != 4 || !strings.Contains(r.out(), "no cache location") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out())
		}
	})
	t.Run("case 9: the default cache is ${XDG_CACHE_HOME}/flow-guard/<key>/flow-guard", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v9")
		xdg := t.TempDir()
		r := runBootstrap(t, fix, withPath(bootstrapEnv("", "XDG_CACHE_HOME="+xdg), path), filepath.Join(fix, "scripts", "flow-guard"), "check-x")
		if bins := cachedBinaries(t, filepath.Join(xdg, "flow-guard")); r.rc != 0 || len(bins) != 1 {
			t.Fatalf("rc=%d out=%s bins=%v", r.rc, r.out(), bins)
		}
	})
	t.Run("case 10: a checkout without stats/ exits the cannot-answer code", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v10")
		if err := os.RemoveAll(filepath.Join(fix, "stats")); err != nil {
			t.Fatal(err)
		}
		r := runBootstrap(t, fix, withPath(bootstrapEnv(t.TempDir()), path), filepath.Join(fix, "scripts", "flow-guard"), "run-reproducer")
		if r.rc != 4 || !strings.Contains(r.out(), "no stats/ source tree") {
			t.Fatalf("rc=%d out=%s", r.rc, r.out())
		}
	})
	t.Run("case 11: sources edited mid-build exit the cannot-answer code and cache nothing", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v11")
		wrap := t.TempDir()
		writeFile(t, filepath.Join(wrap, "go"), fmt.Sprintf("#!/bin/sh\nprintf '// edited mid-build\\n' >> '%s'\nexec '%s' \"$@\"\n",
			filepath.Join(fix, "stats/cmd/flow-guard/main.go"), goBin), 0o755)
		cache := t.TempDir()
		r := runBootstrap(t, fix, withPath(bootstrapEnv(cache), wrap+":"+path), filepath.Join(fix, "scripts", "flow-guard"), "run-reproducer")
		if r.rc != 4 || !strings.Contains(r.out(), "changed during its build") || len(cachedBinaries(t, cache)) != 0 {
			t.Fatalf("rc=%d out=%s", r.rc, r.out())
		}
	})
	t.Run("case 12: a cross-compile environment neither breaks its own call nor poisons the cache", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v12")
		other, otherArch := "darwin", "amd64"
		if runtime.GOOS == "darwin" {
			other = "linux"
		}
		if runtime.GOARCH == "amd64" {
			otherArch = "arm64"
		}
		base := withPath(bootstrapEnv(t.TempDir()), path)
		boot := filepath.Join(fix, "scripts", "flow-guard")
		first := runBootstrap(t, fix, append(base, "GOOS="+other, "GOARCH="+otherArch, "GOFLAGS=-trimpath"), boot, "check-x")
		second := runBootstrap(t, fix, base, boot, "check-x")
		if first.rc != 0 || first.out() != "v12 check-x" || second.rc != 0 || second.out() != "v12 check-x" {
			t.Fatalf("first rc=%d out=%s; second rc=%d out=%s", first.rc, first.out(), second.rc, second.out())
		}
	})
	t.Run("case 13: a failed rename exits the cannot-answer code naming the install", func(t *testing.T) {
		t.Parallel()
		fix := newBootstrapFixture(t, "v13")
		wrap := t.TempDir()
		writeFile(t, filepath.Join(wrap, "go"), fmt.Sprintf("#!/bin/sh\n'%s' \"$@\" || exit $?\nwhile [ $# -gt 0 ]; do [ \"$1\" = -o ] && chmod 555 \"$(dirname \"$2\")\"; shift; done\n", goBin), 0o755)
		cache := t.TempDir()
		r := runBootstrap(t, fix, withPath(bootstrapEnv(cache), wrap+":"+path), filepath.Join(fix, "scripts", "flow-guard"), "run-reproducer")
		for _, k := range cacheKeys(t, cache) {
			_ = os.Chmod(filepath.Join(cache, k), 0o755)
		}
		if r.rc != 4 || !strings.Contains(r.out(), "run-reproducer: could not install flow-guard into") || len(cachedBinaries(t, cache)) != 0 {
			t.Fatalf("rc=%d out=%s", r.rc, r.out())
		}
	})
}

// TestFlowGuardBootstrapRepoRootAndCodes pins what the bootstrap adds over
// scripts/lib/flow-guard.sh: FLOW_GUARD_REPO_ROOT reaches the guard as the
// physical checkout root, over any inherited value, and a call that cannot
// derive the binary exits its guard's own cannot-answer code after a
// "<name>:" opening.
func TestFlowGuardBootstrapRepoRootAndCodes(t *testing.T) {
	smoke.Heavy(t)
	path := os.Getenv("PATH")
	goBinary(t)
	fix := newBootstrapFixture(t, "root")
	writeFile(t, filepath.Join(fix, "stats/cmd/flow-guard/main.go"),
		"package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() { fmt.Println(os.Getenv(\"FLOW_GUARD_REPO_ROOT\")) }\n", 0o644)
	physical, err := filepath.EvalSymlinks(fix)
	if err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(fix, alias); err != nil {
		t.Fatal(err)
	}
	env := withPath(bootstrapEnv(t.TempDir()), path)

	t.Run("the physical checkout root reaches the guard", func(t *testing.T) {
		r := runBootstrap(t, alias, env, filepath.Join(alias, "scripts", "flow-guard"), "check-x")
		if r.rc != 0 || r.out() != physical {
			t.Fatalf("rc=%d out=%q, want %q", r.rc, r.out(), physical)
		}
	})
	t.Run("an inherited FLOW_GUARD_REPO_ROOT is replaced by the bootstrap's own root", func(t *testing.T) {
		r := runBootstrap(t, alias, append(env, "FLOW_GUARD_REPO_ROOT=/inherited/root"), filepath.Join(alias, "scripts", "flow-guard"), "check-x")
		if r.rc != 0 || r.out() != physical {
			t.Fatalf("rc=%d out=%q, want %q", r.rc, r.out(), physical)
		}
	})
	for _, c := range []struct {
		name string
		code int
	}{{"run-reproducer", 4}, {"check-installed-rules", 1}, {"check-plan-shape", 2}} {
		t.Run(c.name+" cannot answer with no go and nothing cached", func(t *testing.T) {
			fresh := newBootstrapFixture(t, "codes")
			r := runBootstrap(t, fresh, withPath(bootstrapEnv(t.TempDir()), noGoPath), filepath.Join(fresh, "scripts", "flow-guard"), c.name)
			if r.rc != c.code || r.stdout != "" || !strings.HasPrefix(r.stderr, c.name+": ") {
				t.Fatalf("rc=%d (want %d) stdout=%q stderr=%q", r.rc, c.code, r.stdout, r.stderr)
			}
		})
	}
}
