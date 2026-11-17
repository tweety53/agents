package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRecapture pins flow-guard recapture: one run of the affected specs
// alone, a receipt for it, nothing run when no spec is affected, and the
// command's failure passed on.
func TestRecapture(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	root := t.TempDir()
	for p, body := range map[string]string{
		"package.json":                "{}\n",
		"src/Day.tsx":                 "export function Day() {}\n",
		"src/Rooms.tsx":               "export function Rooms() {}\n",
		"shots/day.spec.ts":           "test('day', () => {})\n",
		"shots/day.spec.ts.sources":   "src/Day.tsx\n",
		"shots/rooms.spec.ts":         "test('rooms', () => {})\n",
		"shots/rooms.spec.ts.sources": "src/Rooms.tsx\n",
		"docs/notes.md":               "notes\n",
		".flow/project.md":            "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `src/**` |\n",
	} {
		writeFile(t, filepath.Join(root, p), body)
	}
	gitRun(t, root, "init", "-q")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	ran := filepath.Join(t.TempDir(), "ran")
	run := func(command string) guardResult {
		return runGuard("recapture", []string{root, root + "=HEAD", command}, env)
	}
	receipts := func() string {
		raw, _ := os.ReadFile(filepath.Join(root, ".git/flow-recapture.jsonl"))
		return string(raw)
	}

	writeFile(t, filepath.Join(root, "docs/notes.md"), "edited\n")
	if r := run("echo <spec> >> " + ran); r.rc != 0 || !strings.Contains(r.stdout, "nothing to capture") || isFile(ran) {
		t.Errorf("no affected spec: rc=%d out=%q", r.rc, r.out)
	}
	writeFile(t, filepath.Join(root, "src/Day.tsx"), "export function Day() { return 1 }\n")
	if r := run("echo <spec> >> " + ran); r.rc != 0 || !strings.Contains(r.stdout, "recapture: 1 spec(s), exit 0") {
		t.Errorf("one affected spec: rc=%d out=%q", r.rc, r.out)
	}
	if got, _ := os.ReadFile(ran); string(got) != "shots/day.spec.ts\n" {
		t.Errorf("the command ran with %q, want the day spec alone", got)
	}
	if got := receipts(); !strings.Contains(got, `"specs":["shots/day.spec.ts"],"exit":0`) || !strings.Contains(got, `"src/Day.tsx":"`) {
		t.Errorf("receipts = %q", got)
	}
	if r := run("echo <spec>; exit 3"); r.rc != 1 || !strings.Contains(receipts(), `"exit":3`) {
		t.Errorf("a failing command: rc=%d out=%q receipts=%q", r.rc, r.out, receipts())
	}
	for _, args := range [][]string{{root, root + "=HEAD"}, {root, root + "=HEAD", "no placeholder"}, {root, root, "echo <spec>"}} {
		if r := runGuard("recapture", args, env); r.rc != 2 {
			t.Errorf("%v: rc=%d, want 2", args[1:], r.rc)
		}
	}
}
