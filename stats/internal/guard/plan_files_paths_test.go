package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the plan-files-paths contract: a **Files:** path whose
// directory exists only with one segment dropped is refused; an existing
// directory, a new package, a new module, the two-segment floor, a
// placeholder and an archived plan pass.
func TestPlanFilesPaths(t *testing.T) {
	plan := func(p string) string { return "- [ ] 1. A\n\n**Files:** `" + p + "`\n" }
	cases := []struct {
		name, rel, files string
		dirs             []string
		code             int
		want             string
	}{
		{"kan-1013 doubled segment", "spectre/changes/c/tasks.md", "src/app/app/src/main/resources/db/migration/V48__template_review.sql",
			[]string{"src/app/src/main/resources/db/migration"}, 1,
			"tasks.md:3: task 1 names `src/app/app/src/main/resources/db/migration/V48__template_review.sql` in a directory the repository does not carry; `src/app/src/main/resources/db/migration` exists"},
		{"new file in existing directory", "spectre/changes/c/tasks.md", "a/b/New.kt", []string{"a/b"}, 0, ""},
		{"new package", "spectre/changes/c/tasks.md", "a/b/core/services/warmup/X.kt", []string{"a/b/core/services"}, 0, ""},
		{"new module", "spectre/changes/c/tasks.md", "src/app/template/src/main/kotlin/x/Y.kt", []string{"src/app"}, 0, ""},
		{"two-segment floor", "spectre/changes/c/tasks.md", "cmd/spectre/main.go", []string{"spectre"}, 0, ""},
		{"placeholder", "spectre/changes/c/tasks.md", "src/<module>/x.kt", []string{"src/x"}, 0, ""},
		{"archived plan", "spectre/changes/archive/c/tasks.md", "src/app/app/src/main/resources/db/migration/V48__template_review.sql",
			[]string{"src/app/src/main/resources/db/migration"}, 0, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, d := range tc.dirs {
				if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, tc.rel)), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, tc.rel), []byte(plan(tc.files)), 0o644); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := Registry["plan-files-paths"]([]string{tc.rel}, Env{Dir: dir}, &stdout, &stderr)
			if code != tc.code {
				t.Fatalf("exit %d, want %d; stdout=%q stderr=%q", code, tc.code, stdout.String(), stderr.String())
			}
			if tc.want != "" && !strings.Contains(stdout.String(), tc.want) {
				t.Errorf("stdout %q lacks %q", stdout.String(), tc.want)
			}
			if tc.code == 0 && stdout.Len() != 0 {
				t.Errorf("clean plan printed %q", stdout.String())
			}
		})
	}
}
