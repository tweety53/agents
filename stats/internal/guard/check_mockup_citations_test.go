package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the check-mockup-citations contract
// (scripts/check-mockup-citations.sh): a `<mockups root>/<frame id>.png`
// token resolves to a file under the project root, or it is missing.
func TestCheckMockupCitations(t *testing.T) {
	cases := []struct {
		name string
		plan string
		rc   int
		want string
	}{
		{"every cited frame exists",
			"Match `docs/mockups/P1.png` and docs/mockups/P1.png again.\n",
			0, "MOCKUP-CITATIONS-RESOLVED: 2 cited"},
		// KAN-558: kan-469's plan cited P6.png, which no file in the mockups root held.
		{"a cited frame with no file",
			"Match `docs/mockups/P1.png`.\nThe row wraps per `docs/mockups/P6.png`.\n",
			1, "plan.md:2 docs/mockups/P6.png"},
		{"bare ids and other roots are not citations",
			"Per `P6.png`, see other/docs/mockups/P6.png and snapshots/P6.png.\n",
			0, "0 cited"},
		{"a directory named like a frame is no file",
			"See docs/mockups/dir.png.\n",
			1, "docs/mockups/dir.png"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			mockups := filepath.Join(dir, "docs", "mockups")
			if err := os.MkdirAll(filepath.Join(mockups, "dir.png"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(mockups, "P1.png"), nil, 0o644); err != nil {
				t.Fatal(err)
			}
			plan := filepath.Join(dir, "plan.md")
			if err := os.WriteFile(plan, []byte(tc.plan), 0o644); err != nil {
				t.Fatal(err)
			}
			r := runGuard("check-mockup-citations", []string{dir, "docs/mockups/", plan}, Env{Getenv: os.Getenv})
			if r.rc != tc.rc || !strings.Contains(r.out, tc.want) {
				t.Fatalf("rc=%d out=%q; want rc=%d containing %q", r.rc, r.out, tc.rc, tc.want)
			}
		})
	}
	dir := t.TempDir()
	plan := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(plan, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{nil, {dir, "."}, {dir, "absent", plan}, {dir, ".", filepath.Join(dir, "absent.md")}} {
		if r := runGuard("check-mockup-citations", args, Env{Getenv: os.Getenv}); r.rc != 2 {
			t.Fatalf("args %q: rc=%d out=%q; want exit 2", args, r.rc, r.out)
		}
	}
}
