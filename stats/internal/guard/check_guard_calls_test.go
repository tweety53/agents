package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	env := Env{
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
		"skills/x/SKILL.md": "Run `flow-guard unchanged`, `flow-guard workspace port`, `flow-guard <name>` or every `flow-guard check-*`, then `flow-guard check-plan-shape`; `flow-guard check-…`; flow-guard not found on PATH.\n",
	})
	if code != 0 || !strings.Contains(out, "GUARD-CALLS-OK") {
		t.Fatalf("placeholder and glob: code %d, out %q", code, out)
	}
}

func TestCheckGuardCallsOldSpelling(t *testing.T) {
	t.Parallel()
	out, code := gcTree(t, map[string]string{
		"rules/r.mdc": "x\nrun scripts/check-plan-shape.sh here\n",
	})
	if code != 1 || !strings.Contains(out, "rules/r.mdc:2:") {
		t.Fatalf("old spelling: code %d, out %q", code, out)
	}
	out, code = gcTree(t, map[string]string{
		"CLAUDE.md":            "Run ./setup.sh global, then .superpowers/sdd/reproducers/repro.sh.\n",
		"docs/old.md":          "scripts/check-plan-shape.sh\n",
		"spectre/changes/a.md": "check-plan-shape.sh\n",
	})
	if code != 0 {
		t.Fatalf("setup.sh, repro.sh, docs/: code %d, out %q", code, out)
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
	if n := gcCompanions(&out, root, []string{"check-foo", "plan-class"}); n != 1 || !strings.Contains(out.String(), "check_foo_test.go") {
		t.Fatalf("no companion: %d findings, out %q", n, out.String())
	}
	if err := os.WriteFile(filepath.Join(dir, "check_foo_test.go"), []byte("package guard\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if n := gcCompanions(&out, root, []string{"check-foo", "plan-class"}); n != 0 {
		t.Fatalf("with companion: %d findings, out %q", n, out.String())
	}
}
