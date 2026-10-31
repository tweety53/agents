package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// psfRefused is the first refused case's plan: an Implement step deleting
// its declared file and saying a sibling's function drops its entry -- an
// edit of `a/keep.go` its **Files:** never names.
const psfRefused = "- [ ] 1. A\n\n  - [ ] **Step 1: Implement.** Delete `a/x.go`. `keep.go`'s `f` drops its entry.\n\n**Files:** `a/x.go`\n"

// psfWrite writes every rel: body of files under a fresh temp root and
// returns it.
func psfWrite(t *testing.T, files map[string]string) string {
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
	return dir
}

func TestPlanStepFiles(t *testing.T) {
	t.Parallel()
	const rel = "spectre/changes/c/tasks.md"
	cases := []struct {
		name, rel, plan string
		disk            []string
		code            int
		want            string
	}{
		{"a sibling's function drops its entry", rel, psfRefused,
			[]string{"a/x.go", "a/keep.go"}, 1, "tasks.md:3: task 1 step 1 edits `a/keep.go`"},
		{"a list after in:", rel,
			"- [ ] 1. A\n\n  - [ ] **Step 1: Edit.** Rewrite each header sentence, in: `s/one.sh`, `s/two.sh`.\n\n**Files:** `s/one.sh`\n",
			[]string{"s/one.sh", "s/two.sh"}, 1, "tasks.md:3: task 1 step 1 edits `s/two.sh`"},
		{"a clause-opening list followed by a colon", rel,
			"- [ ] 1. A\n\n  - [ ] **Step 1: Edit.** `d/a.md`: x. `d/b.md` (y), `c.md`: each follows.\n\n**Files:** `d/a.md`, `d/b.md`\n",
			[]string{"d/a.md", "d/b.md", "d/c.md"}, 1, "tasks.md:3: task 1 step 1 edits `d/c.md`"},
		{"every edited path declared", rel,
			"- [ ] 1. A\n\n  - [ ] **Step 1: Implement.** Delete `a/x.go`. `keep.go`'s `f` drops its entry.\n\n**Files:** `a/x.go`, `a/keep.go`\n",
			[]string{"a/x.go", "a/keep.go"}, 0, ""},
		{"Allowed-collateral covers the path", rel,
			psfRefused + "**Allowed-collateral:** `a/*.go`\n",
			[]string{"a/x.go", "a/keep.go"}, 0, ""},
		{"a Verify step runs a script", rel,
			"- [ ] 1. A\n\n  - [ ] **Step 1: Verify.** Run `tools/x.sh`; `tools/x.sh`'s exit is 0.\n\n**Files:** `a/x.go`\n",
			[]string{"a/x.go", "tools/x.sh"}, 0, ""},
		{"a read citation in an edit step", rel,
			"- [ ] 1. A\n\n  - [ ] **Step 1: Implement.** copy the rows from the table in `d/c.md` (`d/c.md`).\n\n**Files:** `d/a.md`\n",
			[]string{"d/a.md", "d/c.md"}, 0, ""},
		{"a path that carries something is read, not edited", rel,
			"- [ ] 1. A\n\n  - [ ] **Step 1: Edit.** Rewrite the paragraph. `x/y.go` carries every key today.\n\n**Files:** `a/x.go`\n",
			[]string{"a/x.go", "x/y.go"}, 0, ""},
		{"an edit position naming a path that does not exist", rel,
			"- [ ] 1. A\n\n  - [ ] **Step 1: Edit.** Delete `a/gone.go`.\n\n**Files:** `a/x.go`\n",
			[]string{"a/x.go"}, 0, ""},
		{"a fenced example", rel,
			"- [ ] 1. A\n\n  - [ ] **Step 1: Edit.** Run the block.\n\n```bash\nDelete `a/keep.go`.\n```\n\n**Files:** `a/x.go`\n",
			[]string{"a/x.go", "a/keep.go"}, 0, ""},
		{"an archived plan", "spectre/changes/archive/c/tasks.md", psfRefused,
			[]string{"a/x.go", "a/keep.go"}, 0, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{tc.rel: tc.plan}
			for _, p := range tc.disk {
				files[p] = ""
			}
			dir := psfWrite(t, files)
			var stdout, stderr bytes.Buffer
			code := Registry["plan-step-files"]([]string{tc.rel}, Env{Dir: dir}, &stdout, &stderr)
			if code != tc.code {
				t.Fatalf("exit %d, want %d; stdout=%q stderr=%q", code, tc.code, stdout.String(), stderr.String())
			}
			if tc.want != "" && !strings.Contains(stdout.String(), tc.want) {
				t.Errorf("stdout %q lacks %q", stdout.String(), tc.want)
			}
			if n := strings.Count(stdout.String(), "\n"); tc.code == 1 && n != 1 {
				t.Errorf("stdout %q: want one line, got %d", stdout.String(), n)
			}
			if tc.code == 0 && stdout.Len() != 0 {
				t.Errorf("clean plan printed %q", stdout.String())
			}
		})
	}

	// The wiring: check-plan-shape runs the sub-guard on the plan it scans.
	t.Run("check-plan-shape runs it", func(t *testing.T) {
		t.Parallel()
		dir := psfWrite(t, map[string]string{rel: psfRefused, "a/x.go": "", "a/keep.go": ""})
		var stdout, stderr bytes.Buffer
		code := checkPlanShape([]string{rel}, Env{Dir: dir, Getenv: func(string) string { return "" }}, &stdout, &stderr)
		if code != 1 || !strings.Contains(stdout.String(), "tasks.md:3: task 1 step 1 edits `a/keep.go`") {
			t.Fatalf("exit %d, want 1 and the plan-step-files line; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
}
