package repotest

// TestMakeBuild pins stats/Makefile: the `build` target's recipe, and the
// UI-test stack's agreement with the pidfile flowd writes for itself.
//
// THREE KINDS OF ASSERTION, and the split is deliberate:
//
//   - The build-target cases run `make -n build` in stats/ and assert on the
//     PRINTED RECIPE. `-n` is the whole point: the defect they pin is a MISSING
//     `-o` in the recipe text, and `-n` shows the recipe without running it --
//     no real compile, no `npm ci`, no Vite run.
//
//   - The pidfile-shape cases assert on stats/Makefile's SOURCE TEXT, not on a
//     `make -n` expansion. They are about ui-test-up and ui-test-down, and
//     `make -n` for either is not free of consequences to reach: ui-test-up's
//     prerequisites include web-build.
//
//   - The pid-read case EXECUTES a fragment lifted out of that source text: it
//     asks whether the shell those recipes hand to `kill` actually produces a
//     pid, by running each pid-reading command substitution against a real
//     two-line pidfile in the format internal/pidfile writes. A text assertion
//     cannot catch a `cat` spelled exactly as intended that still yields two
//     lines -- which is the defect it exists for.
//
// WHY THIS EXISTS. `make build` ran `go build ./...` alone, which compiles
// every package and discards the binaries; an operator who skipped the
// hand-written `go build -o bin/flowd ./cmd/flowd` after it ran whatever stale
// bin/flowd was on disk. Each case was first run against the old Makefile and
// failed there.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestMakeBuild(t *testing.T) {
	t.Parallel()
	statsDir := filepath.Join(repoRoot(t), "stats")

	// The recipe, captured once: a `make -n` that itself fails is a hard stop.
	cmd := exec.Command("make", "-n", "build")
	cmd.Dir = statsDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make -n build in %s: %v\n%s", statsDir, err, out)
	}
	recipe := string(out)
	src, err := os.ReadFile(filepath.Join(statsDir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	makefile := string(src)

	recipeMatches := func(name, pattern string) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if !regexp.MustCompile("(?m)" + pattern).MatchString(recipe) {
				t.Errorf("no recipe line matched /%s/; recipe was:\n%s", pattern, recipe)
			}
		})
	}

	// Each `-o` path is anchored on a following space or end-of-line so
	// `bin/flow` cannot be satisfied by the `bin/flowd` line.
	recipeMatches("build target emits bin/flowd",
		`(^|[[:space:]])go build[[:space:]]+-o[[:space:]]+bin/flowd[[:space:]]+\./cmd/flowd([[:space:]]|$)`)
	recipeMatches("build target emits bin/flow",
		`(^|[[:space:]])go build[[:space:]]+-o[[:space:]]+bin/flow[[:space:]]+\./cmd/flow([[:space:]]|$)`)
	recipeMatches("build target emits bin/flow-guard",
		`(^|[[:space:]])go build[[:space:]]+-o[[:space:]]+bin/flow-guard[[:space:]]+\./cmd/flow-guard([[:space:]]|$)`)
	// The compile-all pass survives: without it cmd/uitest-seed and every
	// library with no binary of its own would stop being compiled here.
	recipeMatches("build target still compiles every package",
		`(^|[[:space:]])go build[[:space:]]+\./\.\.\.([[:space:]]|$)`)

	// flowd embeds the model policy that policy-copy refreshes; without it as
	// a prerequisite, the UI-test stack can build on a stale or missing copy.
	t.Run("ui-test-up depends on policy-copy", func(t *testing.T) {
		t.Parallel()
		re := regexp.MustCompile(`(?m)^ui-test-up:[^\n]*\bpolicy-copy\b`)
		if !re.MatchString(makefile) {
			t.Errorf("stats/Makefile's ui-test-up rule does not match /%s/", re)
		}
	})

	// flowd writes $TMPDIR/flowd-<port>.pid itself (internal/pidfile); ANY
	// redirection into $(UITEST_PIDFILE) would be a second writer of one file.
	t.Run("ui-test-up does not write its own pidfile", func(t *testing.T) {
		t.Parallel()
		re := regexp.MustCompile(`>[[:space:]]*\$\(UITEST_PIDFILE\)`)
		for i, line := range strings.Split(makefile, "\n") {
			if re.MatchString(line) {
				t.Errorf("stats/Makefile still matches /%s/ at %d:%s", re, i+1, line)
			}
		}
	})

	// `override` and port-derived, so no caller can aim ui-test-down's `kill`
	// at a path of their choosing; TMPDIR=/tmp on the launch because
	// pidfile.Path joins os.TempDir(), which on macOS is the per-user $TMPDIR.
	t.Run("UITEST_PIDFILE is override and derived from UITEST_PORT", func(t *testing.T) {
		t.Parallel()
		for _, p := range []string{
			`^override[[:space:]]+UITEST_PIDFILE[[:space:]]*:=[[:space:]]*/tmp/flowd-\$\(UITEST_PORT\)\.pid[[:space:]]*$`,
			`(^|[[:space:]])TMPDIR=/tmp([[:space:]]|$)`,
		} {
			if !regexp.MustCompile("(?m)" + p).MatchString(makefile) {
				t.Errorf("no line of stats/Makefile matched /%s/", p)
			}
		}
	})

	// flowd's pidfile has TWO lines -- pid, then the executable path. A
	// `$(cat FILE)` captures both and `kill -0` rejects it as an illegal pid,
	// silently inverting both recipes' guards. Every `$$(... $(UITEST_PIDFILE))`
	// read is rewritten to the shell text make hands /bin/sh and evaluated
	// against a fixture pidfile naming a live process.
	t.Run("ui-test pid reads take only the pidfile's pid line", func(t *testing.T) {
		t.Parallel()
		sleeper := exec.Command("sleep", "120")
		if err := sleeper.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = sleeper.Process.Kill(); _ = sleeper.Wait() })
		fixture := filepath.Join(t.TempDir(), "flowd.pid")
		body := strconv.Itoa(sleeper.Process.Pid) + "\n/tmp/flow-uitest-flowd\n"
		if err := os.WriteFile(fixture, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		reads := regexp.MustCompile(`\$\$\([^()]+\$\(UITEST_PIDFILE\)\)`).FindAllString(makefile, -1)
		// Zero reads is a failure: a case over an empty set asserts nothing.
		if len(reads) == 0 {
			t.Fatal("found no $(UITEST_PIDFILE) read in stats/Makefile to check")
		}
		for _, read := range reads {
			shellRead := strings.ReplaceAll(strings.TrimPrefix(read, "$"), "$(UITEST_PIDFILE)", fixture)
			script := `read_pid="` + shellRead + `"; kill -0 "$read_pid" 2>/dev/null || { printf %s "$read_pid"; exit 1; }`
			if got, err := exec.Command("sh", "-c", script).Output(); err != nil {
				t.Errorf("%s read %q, which kill rejects", read, got)
			}
		}
	})

	// The `build` prerequisite is the only writer of $(LIVE_BIN); the recipe
	// must not have quietly regrown its own build to compensate.
	t.Run("restart depends on build to produce $(LIVE_BIN), and does not rebuild it", func(t *testing.T) {
		t.Parallel()
		var rule string
		var body []string
		inside := false
		for _, line := range strings.Split(makefile, "\n") {
			switch {
			case strings.HasPrefix(line, "restart:"):
				rule, inside = line, true
			case inside && line != "" && line[0] != '\t' && line[0] != '#':
				inside = false
			case inside:
				body = append(body, line)
			}
		}
		if rule == "" {
			t.Error("no `restart:` rule found in stats/Makefile")
		} else if !regexp.MustCompile(`^restart:([[:space:]]|.*[[:space:]])build([[:space:]]|$)`).MatchString(rule) {
			t.Error("`restart:` does not declare `build` as a prerequisite, so nothing refreshes $(LIVE_BIN)")
		}
		if strings.Contains(strings.Join(body, "\n"), "go build -o $(LIVE_BIN)") {
			t.Error("`restart:`'s recipe rebuilds $(LIVE_BIN), which its `build` prerequisite already wrote")
		}
	})

	// KAN-987: web-dist builds the SPA only when internal/web/dist/index.html
	// is absent. `make -n` over a copy of the Makefile, so npm never runs.
	t.Run("web-dist builds the SPA only when it is absent", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "Makefile"), src, 0o600); err != nil {
			t.Fatal(err)
		}
		dist := filepath.Join(dir, "internal", "web", "dist")
		if err := os.MkdirAll(dist, 0o755); err != nil {
			t.Fatal(err)
		}
		index := filepath.Join(dist, "index.html")
		if err := os.WriteFile(index, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		dryRun := func() string {
			out, _ := exec.Command("make", "-s", "-n", "-C", dir, "web-dist").CombinedOutput()
			return string(out)
		}
		built := dryRun()
		if err := os.Remove(index); err != nil {
			t.Fatal(err)
		}
		missing := dryRun()
		if strings.Contains(built, "npm ci") {
			t.Errorf("with index.html present it runs:\n%s", built)
		} else if !strings.Contains(missing, "npm ci") {
			t.Errorf("with no index.html it runs no npm ci:\n%s", missing)
		}
	})
}

// TestMakeWebBuildRefusesSymlinkedNodeModules pins kan-1037's finding: a
// throwaway copy that symlinked the shared web/node_modules had `npm ci` empty
// the shared directory through the link. web-build and web-test must refuse
// before npm runs. A stub npm on PATH records any invocation.
func TestMakeWebBuildRefusesSymlinkedNodeModules(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile(filepath.Join(repoRoot(t), "stats", "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"web-build", "web-test"} {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			dir, shared, bin := t.TempDir(), t.TempDir(), t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "Makefile"), src, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(dir, "web"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(shared, filepath.Join(dir, "web", "node_modules")); err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(shared, "marker")
			if err := os.WriteFile(marker, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			called := filepath.Join(bin, "called")
			stub := "#!/bin/sh\ntouch " + called + "\nexit 0\n"
			if err := os.WriteFile(filepath.Join(bin, "npm"), []byte(stub), 0o755); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("make", "-s", "-C", dir, target)
			cmd.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"))
			var stderr strings.Builder
			cmd.Stderr = &stderr
			if err := cmd.Run(); err == nil {
				t.Error("make exited 0 over a symlinked web/node_modules")
			}
			if !strings.Contains(stderr.String(), "symlinked web/node_modules") {
				t.Errorf("stderr lacks the refusal:\n%s", stderr.String())
			}
			if _, err := os.Stat(called); err == nil {
				t.Error("npm was invoked")
			}
			if _, err := os.Stat(marker); err != nil {
				t.Errorf("marker gone: %v", err)
			}
		})
	}
}
