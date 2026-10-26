package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the check-sidecar-reach contract (scripts/check-sidecar-reach.sh).
func TestCheckSidecarReach(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string // root-relative path → content
		rc    int
		want  string // "" checks rc only; MOCKUPS stands for the mockups root
		// mockups is the mockups root's files (root-relative path → content);
		// nil runs the guard with the capture root alone.
		mockups map[string]string
	}{
		{"literal reaches", map[string]string{
			"fidelity/a.spec.ts":         "await capture(page, 'a-1.png');",
			"fidelity/a.spec.ts.mockups": "# c\n\na-1.png A-1\n",
		}, 0, "SIDECAR-REACH-OK: 1 lines in 1 sidecars\n", nil},
		{"a frame directive line is not a screenshot", map[string]string{
			"fidelity/a.spec.ts":         "await capture(page, 'a-1.png');",
			"fidelity/a.spec.ts.mockups": "# c\nframe none\na-1.png A-1\n",
		}, 0, "SIDECAR-REACH-OK: 1 lines in 1 sidecars\n", nil},
		{"double-quoted literal reaches", map[string]string{
			"fidelity/a.spec.ts":         `shot("a-2.png")`,
			"fidelity/a.spec.ts.mockups": "a-2.png A-2\n",
		}, 0, "", nil},
		{"template literal reaches", map[string]string{
			"fidelity/a.spec.ts":         "toHaveScreenshot(`cardio-guide-${frame}.png`)",
			"fidelity/a.spec.ts.mockups": "cardio-guide-G1.png G1\n",
		}, 0, "", nil},
		{"stale line fails", map[string]string{
			"fidelity/a.spec.ts":         "'a-1.png'",
			"fidelity/a.spec.ts.mockups": "a-1.png A-1\npin-p2-pad-up.png pin-P2\n",
		}, 1, "fidelity/a.spec.ts.mockups:2: pin-p2-pad-up.png — taken by no screenshot in fidelity/a.spec.ts\n", nil},
		{"absent spec fails", map[string]string{
			"fidelity/gone.spec.ts.mockups": "g-1.png G1\n",
		}, 1, "fidelity/gone.spec.ts.mockups:1: g-1.png — fidelity/gone.spec.ts does not exist\n", nil},
		{"pruned dirs are never walked", map[string]string{
			"node_modules/x/s.spec.ts.mockups": "x.png X\n",
			".git/s.spec.ts.mockups":           "x.png X\n",
			".worktrees/c/s.spec.ts.mockups":   "x.png X\n",
		}, 0, "", nil},
		// gymie-playwright's entry.spec.ts: an odd backtick in a comment
		// must not swallow the quoted literal after it.
		{"a stray backtick does not hide a literal", map[string]string{
			"fidelity/a.spec.ts":         "// it's `odd\nshot('a-1.png')\n",
			"fidelity/a.spec.ts.mockups": "a-1.png A-1\n",
		}, 0, "", nil},
		{"a substring is no literal", map[string]string{
			"fidelity/a.spec.ts":         "'a-10.png'",
			"fidelity/a.spec.ts.mockups": "a-1.png A-1\n",
		}, 1, "", nil},
		{"a frame present in the mockups root passes", map[string]string{
			"fidelity/a.spec.ts":         "'a-1.png'",
			"fidelity/a.spec.ts.mockups": "a-1.png A-1\n",
		}, 0, "SIDECAR-REACH-OK: 1 lines in 1 sidecars\n", map[string]string{"A-1.png": ""}},
		{"a frame absent from the mockups root fails", map[string]string{
			"fidelity/a.spec.ts":         "'a-1.png'",
			"fidelity/a.spec.ts.mockups": "a-1.png A-1\n",
		}, 1, "fidelity/a.spec.ts.mockups:1: a-1.png — frame A-1 has no MOCKUPS/A-1.png\n", map[string]string{"B-1.png": ""}},
		{"a one-field line is not judged on its frame", map[string]string{
			"fidelity/a.spec.ts":         "'a-1.png'",
			"fidelity/a.spec.ts.mockups": "a-1.png\n",
		}, 0, "", map[string]string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			for rel, body := range c.files {
				p := filepath.Join(root, rel)
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{root}
			want := c.want
			if c.mockups != nil {
				mockups := t.TempDir()
				for rel, body := range c.mockups {
					if err := os.WriteFile(filepath.Join(mockups, rel), []byte(body), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				args = append(args, mockups)
				want = strings.ReplaceAll(want, "MOCKUPS", mockups)
			}
			r := runGuard("check-sidecar-reach", args, Env{})
			if r.rc != c.rc || (want != "" && r.stdout != want) {
				t.Fatalf("rc=%d stdout=%q, want rc=%d stdout=%q", r.rc, r.stdout, c.rc, want)
			}
		})
	}

	file := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string][]string{
		"no argument":                  nil,
		"three arguments":              {t.TempDir(), t.TempDir(), t.TempDir()},
		"not a directory":              {file},
		"mockups root not a directory": {t.TempDir(), file},
	} {
		t.Run("exit 2: "+name, func(t *testing.T) {
			r := runGuard("check-sidecar-reach", args, Env{})
			if r.rc != 2 || r.stdout != "" {
				t.Fatalf("rc=%d stdout=%q, want 2 and empty stdout", r.rc, r.stdout)
			}
		})
	}
}
