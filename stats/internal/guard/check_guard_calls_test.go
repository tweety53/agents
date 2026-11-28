package guard

import (
	"bufio"
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// gcTree writes files (path relative to a fresh root → content) and runs
// check-guard-calls over it, returning stdout and the exit code.
func gcTree(t *testing.T, files map[string]string) (string, int) {
	t.Helper()
	root := t.TempDir()
	for p, body := range files {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	env := core.Env{
		Getenv:    func(string) string { return "" },
		LookupEnv: func(k string) (string, bool) { return root, k == "CHECK_GUARD_CALLS_ROOT" },
		Dir:       root,
	}
	var out, errb bytes.Buffer
	code := checkGuardCalls(nil, env, &out, &errb)
	if code == 2 {
		t.Fatalf("cannot answer: %s", errb.String())
	}
	return out.String(), code
}

func TestCheckGuardCallsUnknownName(t *testing.T) {
	t.Parallel()
	out, code := gcTree(t, map[string]string{
		"skills/x/SKILL.md": "Run `flow-guard check-nope` first.\n",
	})
	if code != 1 || !strings.Contains(out, "skills/x/SKILL.md:1:") || !strings.Contains(out, "GUARD-CALLS-INVALID") {
		t.Fatalf("unknown name: code %d, out %q", code, out)
	}
	out, code = gcTree(t, map[string]string{
		"skills/x/SKILL.md": "Run `flow-guard unchangd` and\n`flow-guard worspace port`.\n",
	})
	if code != 1 || !strings.Contains(out, "skills/x/SKILL.md:1: flow-guard unchangd") || !strings.Contains(out, "skills/x/SKILL.md:2: flow-guard worspace") {
		t.Fatalf("typo of an unhyphenated guard: code %d, out %q", code, out)
	}
	out, code = gcTree(t, map[string]string{
		"skills/x/SKILL.md": "Run `flow-guard unchanged`, `flow-guard workspace port`, `flow-guard <name>` or every `flow-guard check-*`, then `flow-guard check-guard-calls`; `flow-guard check-…`; flow-guard not found on PATH.\n",
	})
	if code != 0 || !strings.Contains(out, "GUARD-CALLS-OK") {
		t.Fatalf("placeholder and glob: code %d, out %q", code, out)
	}
}

func TestCheckGuardCallsOldSpelling(t *testing.T) {
	t.Parallel()
	out, code := gcTree(t, map[string]string{
		"rules/r.mdc": "x\nrun scripts/check-guard-calls.sh here\n",
	})
	if code != 1 || !strings.Contains(out, "rules/r.mdc:2:") {
		t.Fatalf("old spelling: code %d, out %q", code, out)
	}
	out, code = gcTree(t, map[string]string{
		"CLAUDE.md":            "Run ./setup.sh global, then .superpowers/sdd/reproducers/repro.sh.\n",
		"docs/old.md":          "scripts/check-guard-calls.sh\n",
		"spectre/changes/a.md": "check-guard-calls.sh\n",
	})
	if code != 0 {
		t.Fatalf("setup.sh, repro.sh, docs/: code %d, out %q", code, out)
	}
}

// TestCheckGuardCallsGoLiterals: rule B reads the string literals of
// non-test Go sources under stats/, never their comments or *_test.go.
func TestCheckGuardCallsGoLiterals(t *testing.T) {
	t.Parallel()
	out, code := gcTree(t, map[string]string{
		"stats/x/a.go": "package x\n\n// close-task.sh in a comment\nvar s = \"run close-task.sh now\"\n",
	})
	if code != 1 || !strings.Contains(out, "stats/x/a.go:4: close-task.sh") || strings.Contains(out, "a.go:3:") {
		t.Fatalf("go literal: code %d, out %q", code, out)
	}
	out, code = gcTree(t, map[string]string{
		"stats/x/a.go":      "package x\n\nvar s = `./setup.sh global`\n",
		"stats/x/a_test.go": "package x\n\nvar t = \"close-task.sh\"\n",
	})
	if code != 0 {
		t.Fatalf("setup.sh and _test.go: code %d, out %q", code, out)
	}
	out, code = gcTree(t, map[string]string{
		"stats/x/testdata/a.go": "package x\n\nvar s = \"run close-task.sh now\"\n",
	})
	if code != 0 {
		t.Fatalf("testdata is skipped: code %d, out %q", code, out)
	}
}

func TestCheckGuardCallsCompanion(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "stats", "internal", "guard")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if n := gcCompanions(&out, root, []string{"check-foo", "close-task"}); n != 1 || !strings.Contains(out.String(), "check_foo_test.go") {
		t.Fatalf("no companion: %d findings, out %q", n, out.String())
	}
	if err := os.WriteFile(filepath.Join(dir, "check_foo_test.go"), []byte("package guard\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if n := gcCompanions(&out, root, []string{"check-foo", "close-task"}); n != 0 {
		t.Fatalf("with companion: %d findings, out %q", n, out.String())
	}
	// A guard living in one of guard's subpackages has its test beside it there.
	sub := filepath.Join(dir, "gitops")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "check_bar_test.go"), []byte("package gitops\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if n := gcCompanions(&out, root, []string{"check-bar"}); n != 0 {
		t.Fatalf("with a subpackage companion: %d findings, out %q", n, out.String())
	}
}

// TestCheckGuardCallsCompanionWired runs rule C through the registered guard:
// a root carrying stats/internal/guard/ and no companion tests is a finding
// per check-* guard.
func TestCheckGuardCallsCompanionWired(t *testing.T) {
	t.Parallel()
	out, code := gcTree(t, map[string]string{
		"stats/internal/guard/doc.go": "package guard\n",
	})
	if code != 1 || !strings.Contains(out, "stats/internal/guard/check_guard_calls_test.go:0: guard check-guard-calls has no Go test") {
		t.Fatalf("rule C through checkGuardCalls: code %d, out %q", code, out)
	}
}

// TestCheckGuardCallsMatchesSequential: the concurrent scans write the same
// output, in the same order, with the same exit code as the sequential
// guard they replaced (gcSequential and gcGoLiteralsSequential, kept here
// verbatim as the oracle) -- over this repository and over a fixture with
// more violating files of each kind than workers.
func TestCheckGuardCallsMatchesSequential(t *testing.T) {
	t.Parallel()
	fixture := t.TempDir()
	write := func(p, body string) {
		full := filepath.Join(fixture, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for i := range 64 {
		goBody := fmt.Sprintf("package x\n\nvar a = \"ok\"\nvar b = \"run close-task.sh %d\"\nvar c = `then close-task.sh`\n", i)
		mdBody := fmt.Sprintf("Run `flow-guard check-nope-%d`, then\n`flow-guard unchangd` and close-task.sh.\n", i)
		if i%5 == 0 {
			goBody, mdBody = "package x\n\nvar a = \"clean\"\n", "Run `flow-guard check-guard-calls`.\n"
		}
		write(fmt.Sprintf("stats/p%02d/f%02d.go", i%7, i), goBody)
		write(fmt.Sprintf("skills/s%02d/f%02d.md", i%7, i), mdBody)
	}
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	for name, root := range map[string]string{"fixture": fixture, "repository": repo} {
		env := core.Env{
			Getenv:    func(string) string { return "" },
			LookupEnv: func(k string) (string, bool) { return root, k == "CHECK_GUARD_CALLS_ROOT" },
			Dir:       root,
		}
		var got, errb bytes.Buffer
		code := checkGuardCalls(nil, env, &got, &errb)
		want, wantCode := gcSequential(t, root)
		if code != wantCode || got.String() != want {
			t.Errorf("%s: concurrent exit %d\n%s\nsequential exit %d\n%s", name, code, got.String(), wantCode, want)
		}
		if name == "fixture" && (wantCode != 1 || !strings.Contains(want, "rule A") || !strings.Contains(want, ".go:4:")) {
			t.Errorf("fixture: sequential exit %d without rule A and Go-literal findings, so the comparison proves nothing:\n%s", wantCode, want)
		}
	}
}

// gcSequential is checkGuardCalls' stdout and exit code as the sequential
// scan computed them, for a root that answers.
func gcSequential(t *testing.T, root string) (string, int) {
	t.Helper()
	files, err := gcFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	n := 0
	for _, rel := range files {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		sc := bufio.NewScanner(bytes.NewReader(b))
		sc.Buffer(nil, 1<<24)
		for line := 1; sc.Scan(); line++ {
			for _, m := range gcCall.FindAllStringSubmatch(sc.Text(), -1) {
				if m[2] == "" && Registry[m[1]] == nil {
					fmt.Fprintf(&out, "%s:%d: flow-guard %s names no registered guard — fix the name or register the guard (rule A)\n", rel, line, m[1])
					n++
				}
			}
			for _, m := range gcBare.FindAllStringSubmatch(sc.Text(), -1) {
				if gcTypo(m[1]) {
					fmt.Fprintf(&out, "%s:%d: flow-guard %s names no registered guard — fix the name or register the guard (rule A)\n", rel, line, m[1])
					n++
				}
			}
			for _, m := range gcShell.FindAllStringSubmatch(sc.Text(), -1) {
				if Registry[m[1]] != nil {
					fmt.Fprintf(&out, "%s:%d: %s.sh is the deleted shim spelling — write flow-guard %s (rule B)\n", rel, line, m[1], m[1])
					n++
				}
			}
		}
	}
	gn, err := gcGoLiteralsSequential(&out, root)
	if err != nil {
		t.Fatal(err)
	}
	n += gn
	if core.IsDir(filepath.Join(root, "stats", "internal", "guard")) {
		n += gcCompanions(&out, root, slices.Collect(maps.Keys(Registry)))
	}
	if n > 0 {
		fmt.Fprintf(&out, "GUARD-CALLS-INVALID: %s — %d violation(s)\n", root, n)
		return out.String(), 1
	}
	fmt.Fprintf(&out, "GUARD-CALLS-OK: %s — %d file(s) scanned\n", root, len(files))
	return out.String(), 0
}

func gcGoLiteralsSequential(out io.Writer, root string) (int, error) {
	stats := filepath.Join(root, "stats")
	if !core.IsDir(stats) {
		return 0, nil
	}
	n := 0
	err := filepath.WalkDir(stats, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == "testdata") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		fset := token.NewFileSet()
		file := fset.AddFile(rel, -1, len(src))
		var sc scanner.Scanner
		sc.Init(file, src, nil, 0)
		for {
			pos, tok, lit := sc.Scan()
			if tok == token.EOF {
				return nil
			}
			if tok != token.STRING {
				continue
			}
			for _, m := range gcShell.FindAllStringSubmatch(lit, -1) {
				if Registry[m[1]] != nil {
					fmt.Fprintf(out, "%s:%d: %s.sh is the deleted shim spelling — write flow-guard %s (rule B)\n", rel, fset.Position(pos).Line, m[1], m[1])
					n++
				}
			}
		}
	})
	return n, err
}
