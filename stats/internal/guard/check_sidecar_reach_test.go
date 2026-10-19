package guard

import (
	"os"
	"path/filepath"
	"testing"
)

// Every case of the check-sidecar-reach contract (scripts/check-sidecar-reach.sh).
func TestCheckSidecarReach(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string // root-relative path → content
		rc    int
		want  string // "" checks rc only
	}{
		{"literal reaches", map[string]string{
			"fidelity/a.spec.ts":         "await capture(page, 'a-1.png');",
			"fidelity/a.spec.ts.mockups": "# c\n\na-1.png A-1\n",
		}, 0, "SIDECAR-REACH-OK: 1 lines in 1 sidecars\n"},
		{"double-quoted literal reaches", map[string]string{
			"fidelity/a.spec.ts":         `shot("a-2.png")`,
			"fidelity/a.spec.ts.mockups": "a-2.png A-2\n",
		}, 0, ""},
		{"template literal reaches", map[string]string{
			"fidelity/a.spec.ts":         "toHaveScreenshot(`cardio-guide-${frame}.png`)",
			"fidelity/a.spec.ts.mockups": "cardio-guide-G1.png G1\n",
		}, 0, ""},
		{"stale line fails", map[string]string{
			"fidelity/a.spec.ts":         "'a-1.png'",
			"fidelity/a.spec.ts.mockups": "a-1.png A-1\npin-p2-pad-up.png pin-P2\n",
		}, 1, "fidelity/a.spec.ts.mockups:2: pin-p2-pad-up.png — taken by no screenshot in fidelity/a.spec.ts\n"},
		{"absent spec fails", map[string]string{
			"fidelity/gone.spec.ts.mockups": "g-1.png G1\n",
		}, 1, "fidelity/gone.spec.ts.mockups:1: g-1.png — fidelity/gone.spec.ts does not exist\n"},
		{"pruned dirs are never walked", map[string]string{
			"node_modules/x/s.spec.ts.mockups": "x.png X\n",
			".git/s.spec.ts.mockups":           "x.png X\n",
			".worktrees/c/s.spec.ts.mockups":   "x.png X\n",
		}, 0, ""},
		// gymie-playwright's entry.spec.ts: an odd backtick in a comment
		// must not swallow the quoted literal after it.
		{"a stray backtick does not hide a literal", map[string]string{
			"fidelity/a.spec.ts":         "// it's `odd\nshot('a-1.png')\n",
			"fidelity/a.spec.ts.mockups": "a-1.png A-1\n",
		}, 0, ""},
		{"a substring is no literal", map[string]string{
			"fidelity/a.spec.ts":         "'a-10.png'",
			"fidelity/a.spec.ts.mockups": "a-1.png A-1\n",
		}, 1, ""},
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
			r := runGuard("check-sidecar-reach", []string{root}, Env{})
			if r.rc != c.rc || (c.want != "" && r.stdout != c.want) {
				t.Fatalf("rc=%d stdout=%q, want rc=%d stdout=%q", r.rc, r.stdout, c.rc, c.want)
			}
		})
	}

	file := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string][]string{
		"no argument":     nil,
		"two arguments":   {t.TempDir(), t.TempDir()},
		"not a directory": {file},
	} {
		t.Run("exit 2: "+name, func(t *testing.T) {
			r := runGuard("check-sidecar-reach", args, Env{})
			if r.rc != 2 || r.stdout != "" {
				t.Fatalf("rc=%d stdout=%q, want 2 and empty stdout", r.rc, r.stdout)
			}
		})
	}
}
