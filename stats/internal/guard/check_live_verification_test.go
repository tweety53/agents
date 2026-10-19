package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the check-live-verification contract
// (scripts/check-live-verification.sh).
func TestCheckLiveVerification(t *testing.T) {
	cases := []struct {
		name, body string
		rc         int
		want       string // %R% is the record path
	}{
		{"before table header passes", "| Check | Before | After |\n|---|---|---|\n| a | 1 | 2 |\n", 0, "LIVE-VERIFICATION-OK: %R%\n"},
		{"before line passes", "- Before: 40 requests\n- After: 28\n", 0, "LIVE-VERIFICATION-OK: %R%\n"},
		{"no before figure fails", "| Check | After |\n| a | 2 |\n", 1, "%R%: no before figure — no line names one\n"},
		// kan-998's record.
		{"before not measured fails",
			"| Check | After |\n\nBefore figure: not measured on a merge-base stack; the pre-change shape is what V12 created.\n", 1,
			"%R%:3: before figure not measured — Before figure: not measured on a merge-base stack; the pre-change shape is what V12 created.\n"},
		{"unmeasured and not taken fail", "before: unmeasured\nBEFORE not taken\n", 1,
			"%R%:1: before figure not measured — before: unmeasured\n%R%:2: before figure not measured — BEFORE not taken\n"},
		{"a not measured cell under a Before header fails",
			"| Check | Before | After |\n|---|---|---|\n| requests | not measured | 28 |\n", 1,
			"%R%:3: before figure not measured — | requests | not measured | 28 |\n"},
		{"word inside another word is no before", "beforehand 3\n", 1, "%R%: no before figure — no line names one\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := filepath.Join(t.TempDir(), "live-verification.md")
			if err := os.WriteFile(rec, []byte(c.body), 0o644); err != nil {
				t.Fatal(err)
			}
			r := runGuard("check-live-verification", []string{rec}, Env{})
			want := strings.ReplaceAll(c.want, "%R%", rec)
			if r.rc != c.rc || r.stdout != want {
				t.Fatalf("rc=%d stdout=%q, want rc=%d stdout=%q", r.rc, r.stdout, c.rc, want)
			}
		})
	}

	dir := t.TempDir()
	for name, args := range map[string][]string{
		"no argument":   nil,
		"two arguments": {"a", "b"},
		"absent path":   {filepath.Join(dir, "absent.md")},
		"a directory":   {dir},
	} {
		t.Run("exit 2: "+name, func(t *testing.T) {
			r := runGuard("check-live-verification", args, Env{})
			if r.rc != 2 || r.stdout != "" {
				t.Fatalf("rc=%d stdout=%q, want 2 and empty stdout", r.rc, r.stdout)
			}
		})
	}
}
