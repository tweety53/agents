package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the plan-script-paths contract: a live plan's unfenced inline
// span citing a `scripts/…` path the repository lacks is refused, naming the
// task, the line and the path; a present path, a path a task's **Files:**
// names, a placeholder, a fenced example and an archived plan pass.
func TestPlanScriptPaths(t *testing.T) {
	cases := []struct {
		name, rel, plan string
		code            int
		want            []string
	}{
		{"retired guard cited (kan-991 step 6)", "spectre/changes/c/tasks.md",
			"- [ ] 1. A\n\n**Files:** `a.md`\n\n  - [ ] **Step 6: Verify: `scripts/present.sh && scripts/check-vocabulary.sh`.**\n",
			1, []string{"tasks.md:5: task 1 cites `scripts/check-vocabulary.sh`"}},
		{"present path with a line suffix", "spectre/changes/c/tasks.md",
			"- [ ] 1. A\n\n- see `scripts/present.sh:12` and `bash scripts/present.sh`\n", 0, nil},
		{"path the plan creates", "spectre/changes/c/tasks.md",
			"- [ ] 1. A\n\n**Files:** `scripts/new.sh`\n\n- [ ] 2. B\n\n- run `scripts/new.sh`\n", 0, nil},
		{"placeholder", "spectre/changes/c/tasks.md",
			"- [ ] 1. A\n\n- every `scripts/check-*.sh` and `scripts/<name>.sh`\n", 0, nil},
		{"fenced example", "spectre/changes/c/tasks.md",
			"- [ ] 1. A\n\n```bash\n`scripts/gone.sh`\n```\n", 0, nil},
		{"archived plan", "spectre/changes/archive/c/tasks.md",
			"- [ ] 1. A\n\n- `scripts/gone.sh`\n", 0, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for p, body := range map[string]string{"scripts/present.sh": "", tc.rel: tc.plan} {
				if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, p)), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, p), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			var stdout, stderr bytes.Buffer
			code := Registry["plan-script-paths"]([]string{tc.rel}, Env{Dir: dir}, &stdout, &stderr)
			if code != tc.code {
				t.Fatalf("exit %d, want %d; stdout=%q stderr=%q", code, tc.code, stdout.String(), stderr.String())
			}
			for _, w := range tc.want {
				if !strings.Contains(stdout.String(), w) {
					t.Errorf("stdout %q lacks %q", stdout.String(), w)
				}
			}
			if tc.code == 0 && stdout.Len() != 0 {
				t.Errorf("clean plan printed %q", stdout.String())
			}
		})
	}
}
