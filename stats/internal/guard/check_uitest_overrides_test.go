package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-uitest-overrides.sh at b1b0d601, one
// subtest per case label, plus the default-Makefile resolution the harness
// never exercised. The harness pointed the guard at a fixture Makefile and
// matched its combined output; here the guard runs in-process.

func uoRun(env map[string]string, args ...string) guardResult {
	return runGuard("check-uitest-overrides", args, Env{Dir: "/", Getenv: func(k string) string { return env[k] }})
}

// uoFixture writes files into a fresh directory and returns its Makefile.
func uoFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for p, body := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, p)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, p), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "Makefile")
}

func uoAssert(t *testing.T, r guardResult, rc int, wants ...string) {
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

func TestCheckUitestOverrides(t *testing.T) {
	t.Parallel()
	// assert_rejects: the body as printf '%s\n' wrote it, exit 1, FAIL verdict.
	for name, body := range map[string]string{
		"case 1 (leading space, plain :=)":   "  UITEST_EVIL := /etc/passwd",
		"case 2 (?=)":                        "UITEST_EVIL ?= /etc/passwd",
		"case 3 (+=)":                        "UITEST_EVIL += /etc/passwd",
		"case 4 (recursive =)":               "UITEST_EVIL = /etc/passwd",
		"case 5 (::=)":                       "UITEST_EVIL ::= /etc/passwd",
		"case 11 (define, no override)":      "define UITEST_EVIL\nfoo\nendef",
		"case 13 (export, no override)":      "export UITEST_EVIL := /etc/passwd",
		"case 15 ($(eval ...), no override)": "$(eval UITEST_EVIL := /etc/passwd)",
	} {
		t.Run(name, func(t *testing.T) {
			mk := uoFixture(t, map[string]string{"Makefile": body + "\n"})
			uoAssert(t, uoRun(nil, mk), 1, "UITEST-OVERRIDES-FAIL:")
		})
	}
	for name, body := range map[string]string{
		"case 6":  "override UITEST_GOOD := ok\n",
		"case 7":  "OTHER_VAR := x\noverride OTHER_VAR2 := y\n",
		"case 8":  "dummy:\n\tUITEST_EVIL := not-a-make-assignment\n",
		"case 12": "override define UITEST_GOOD\nfoo\nendef\n",
		"case 14": "override export UITEST_GOOD := ok\n",
		"case 16": "$(eval override UITEST_GOOD := ok)\n",
	} {
		t.Run(name, func(t *testing.T) {
			mk := uoFixture(t, map[string]string{"Makefile": body})
			uoAssert(t, uoRun(nil, mk), 0, "UITEST-OVERRIDES-OK:")
		})
	}
	t.Run("case 9", func(t *testing.T) {
		mk := uoFixture(t, map[string]string{"Makefile": "include extra.mk\noverride UITEST_GOOD := ok\n", "extra.mk": "UITEST_HIDDEN := /etc/passwd\n"})
		r := uoRun(nil, mk)
		uoAssert(t, r, 1, "extra.mk")
		if i := strings.Index(r.out, "extra.mk"); i < 0 || !strings.Contains(r.out[i:], "UITEST_HIDDEN") {
			t.Errorf("output does not name extra.mk then UITEST_HIDDEN:\n%s", r.out)
		}
	})
	t.Run("case 10", func(t *testing.T) {
		mk := uoFixture(t, map[string]string{"Makefile": "include $(EXTRA_FILES)\noverride UITEST_GOOD := ok\n"})
		uoAssert(t, uoRun(nil, mk), 0, "cannot resolve")
	})
	t.Run("default makefile", func(t *testing.T) {
		mk := uoFixture(t, map[string]string{"stats/Makefile": "UITEST_EVIL := x\n"})
		root := filepath.Dir(mk)
		uoAssert(t, uoRun(map[string]string{"FLOW_GUARD_REPO_ROOT": root}), 1, root+"/stats/Makefile:1: UITEST_EVIL")
	})
	t.Run("missing makefile", func(t *testing.T) {
		uoAssert(t, uoRun(nil, "/no/such/Makefile"), 2, "check-uitest-overrides: /no/such/Makefile not found")
	})
	t.Run("unreadable makefile", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("root reads a mode-000 file")
		}
		mk := uoFixture(t, map[string]string{"Makefile": "include extra.mk\n", "extra.mk": "UITEST_EVIL := x\n"})
		if err := os.Chmod(filepath.Join(filepath.Dir(mk), "extra.mk"), 0); err != nil {
			t.Fatal(err)
		}
		uoAssert(t, uoRun(nil, mk), 2, "extra.mk: open")
	})
	t.Run("line too long", func(t *testing.T) {
		mk := uoFixture(t, map[string]string{"Makefile": strings.Repeat("x", 1<<24+1) + "\nUITEST_EVIL := x\n"})
		uoAssert(t, uoRun(nil, mk), 2, "token too long")
	})
	t.Run("repo root unset", func(t *testing.T) {
		uoAssert(t, uoRun(nil), 2, "FLOW_GUARD_REPO_ROOT is unset")
	})
}
