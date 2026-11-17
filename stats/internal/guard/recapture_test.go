package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestRecapture pins flow-guard recapture: one run of the task's specs
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
		"plan/tasks.md":               "- [ ] 4. In-run fix 1 — visual-verify: x\n\n**Files:** `src/Day.tsx`\n**Tests:** none\n",
		".flow/project.md":            "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `src/**` |\n",
	} {
		writeFile(t, filepath.Join(root, p), body)
	}
	gitRun(t, root, "init", "-q")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	ran := filepath.Join(t.TempDir(), "ran")
	run := func(command string) guardResult {
		return runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", command}, env)
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
	if r := run("echo <spec> >> " + ran); r.rc != 0 || !strings.Contains(r.stdout, "recapture: 1 test selection(s), exit 0") {
		t.Errorf("one affected spec: rc=%d out=%q", r.rc, r.out)
	}
	if got, _ := os.ReadFile(ran); string(got) != "shots/day.spec.ts --retries=1\n" {
		t.Errorf("the command ran with %q, want the day spec alone", got)
	}
	if got := receipts(); !strings.Contains(got, `"specs":["shots/day.spec.ts"],"exit":0`) || !strings.Contains(got, `"src/Day.tsx":"`) {
		t.Errorf("receipts = %q", got)
	}
	writeFile(t, filepath.Join(root, "src/Day.tsx"), "export function Day() { return 2 }\n")
	if r := run("echo <spec>; exit 3"); r.rc != 1 || !strings.Contains(receipts(), `"exit":3`) {
		t.Errorf("a failing command: rc=%d out=%q receipts=%q", r.rc, r.out, receipts())
	}
	for _, args := range [][]string{{root, root + "=HEAD", "echo <spec>"}, {root + "/plan/tasks.md:4", root, root + "=HEAD", "no placeholder"}, {root + "/plan/tasks.md:4", root, root, "echo <spec>"}, {root + "/plan/tasks.md:9", root, root + "=HEAD", "echo <spec>"}} {
		if r := runGuard("recapture", args, env); r.rc != 2 {
			t.Errorf("%v: rc=%d, want 2", args[1:], r.rc)
		}
	}
}

// TestRecaptureTakesTheTaskSetAndRerunsOnlyWhatFailed pins the two halves of
// the kan-1091 fix: a fixture every spec renders selects the specs the task's
// Tests name and no other, and a red run is followed by one run of its
// failing test alone, after which nothing is left to run.
func TestRecaptureTakesTheTaskSetAndRerunsOnlyWhatFailed(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	root := t.TempDir()
	for p, body := range map[string]string{
		"package.json":        "{}\n",
		"shots/fixtures.ts":   "export const seed = 1\n",
		"shots/all.sources":   "shots/fixtures.ts\n",
		"shots/day.spec.ts":   "test('day', () => {})\n",
		"shots/rooms.spec.ts": "test('rooms', () => {})\n",
		"shots/week.spec.ts":  "test('week', () => {})\n",
		"plan/tasks.md":       "- [ ] 4. In-run fix 1 — visual-verify: x\n\n**Files:** `shots/fixtures.ts`\n**Tests:** `shots/day.spec.ts` › day\n**Regression:** `e2e/shots/rooms.spec.ts`\n",
	} {
		writeFile(t, filepath.Join(root, p), body)
		if strings.HasSuffix(p, ".spec.ts") {
			writeFile(t, filepath.Join(root, p+".sources"), "src/"+filepath.Base(p)+".tsx\n")
		}
	}
	gitRun(t, root, "init", "-q")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	writeFile(t, filepath.Join(root, "shots/fixtures.ts"), "export const seed = 2\n")
	ran := filepath.Join(t.TempDir(), "ran")
	run := func(command string) guardResult {
		return runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", command}, env)
	}
	red := "printf '  1 failed\\n    [chromium] › shots/rooms.spec.ts:3:1 › rooms\\n'; echo <spec> >> " + ran + "; exit 1"
	if r := run(red); r.rc != 1 {
		t.Fatalf("the red run: rc=%d out=%q", r.rc, r.out)
	}
	if r := run("echo <spec> >> " + ran); r.rc != 0 {
		t.Fatalf("the re-run: rc=%d out=%q", r.rc, r.out)
	}
	if r := run("echo <spec> >> " + ran); r.rc != 0 || !strings.Contains(r.stdout, "already passed") {
		t.Errorf("nothing left: rc=%d out=%q", r.rc, r.out)
	}
	const want = "shots/day.spec.ts shots/rooms.spec.ts --retries=1\nshots/rooms.spec.ts:3 --retries=1\n"
	if got, _ := os.ReadFile(ran); string(got) != want {
		t.Errorf("the command ran with %q, want %q", got, want)
	}
	// A red run that cannot name its failures failed every spec.
	writeFile(t, filepath.Join(root, "shots/fixtures.ts"), "export const seed = 3\n")
	run("echo <spec> >> " + ran + "; exit 1")
	run("echo <spec> >> " + ran)
	if got, _ := os.ReadFile(ran); !strings.HasSuffix(string(got), "shots/day.spec.ts shots/rooms.spec.ts --retries=1\nshots/day.spec.ts shots/rooms.spec.ts --retries=1\n") {
		t.Errorf("after an unnamed failure the command ran with %q, want both specs again", got)
	}
	// A retry-pass is flaky, never failed: nothing is left to run after it.
	writeFile(t, filepath.Join(root, "shots/fixtures.ts"), "export const seed = 4\n")
	r := run("printf '  1 flaky\\n    [chromium] › shots/rooms.spec.ts:3:1 › rooms\\n  1 passed\\n'; : <spec>")
	if r.rc != 0 || !strings.Contains(r.stdout, "recapture: flaky (1): shots/rooms.spec.ts:3 — passed on retry") {
		t.Errorf("a flaky run: rc=%d out=%q", r.rc, r.out)
	}
	if r := run("echo <spec> >> " + ran); !strings.Contains(r.stdout, "already passed") {
		t.Errorf("after a flaky run: rc=%d out=%q, want nothing left", r.rc, r.out)
	}
}

// TestRecaptureLeftIsPerTest pins the receipt fold: a test a run passed --
// alone or inside its spec -- is never left to run again, and a red spec
// leaves its failing tests alone.
func TestRecaptureLeftIsPerTest(t *testing.T) {
	t.Parallel()
	whole := rcReceipt{Specs: []string{"a.spec.ts"}, Exit: 1, Failed: []string{"a.spec.ts:9"}}
	for _, c := range []struct {
		token    string
		receipts []rcReceipt
		want     string
	}{
		{"a.spec.ts", nil, "a.spec.ts"},
		{"a.spec.ts", []rcReceipt{whole}, "a.spec.ts:9"},
		{"a.spec.ts:3", []rcReceipt{whole}, ""},
		{"a.spec.ts:9", []rcReceipt{whole}, "a.spec.ts:9"},
		{"a.spec.ts", []rcReceipt{whole, {Specs: []string{"a.spec.ts:9"}}}, ""},
		{"a.spec.ts", []rcReceipt{{Specs: []string{"a.spec.ts:9"}}}, "a.spec.ts"},
		{"a.spec.ts:3", []rcReceipt{{Specs: []string{"a.spec.ts"}, Exit: 1}}, "a.spec.ts:3"},
		{"a.spec.ts:3", []rcReceipt{{Specs: []string{"a.spec.ts"}, Exit: 1, Failed: []string{"a.spec.ts"}}}, "a.spec.ts:3"},
	} {
		if left := rcLeft(c.token, c.receipts); strings.Join(left, " ") != c.want {
			t.Errorf("rcLeft(%s, %+v) = %q, want %q", c.token, c.receipts, left, c.want)
		}
	}
}

// TestRecaptureRecordsItsCost: a recapture that ran files one `flow record
// run` row, labelled with its task id and carrying the run's test counts.
func TestRecaptureRecordsItsCost(t *testing.T) {
	t.Parallel()
	stub, log := rvlFlowStub(t, "{}", 0)
	env := Env{Getenv: func(k string) string {
		if k == "PATH" {
			return stub + ":" + os.Getenv("PATH")
		}
		return os.Getenv(k)
	}}
	root := t.TempDir()
	for p, body := range map[string]string{
		"package.json":              "{}\n",
		"src/Day.tsx":               "export function Day() {}\n",
		"shots/day.spec.ts":         "test('day', () => {})\n",
		"shots/day.spec.ts.sources": "src/Day.tsx\n",
		"plan/tasks.md":             "- [ ] 4. In-run fix 1 — visual-verify: x\n\n**Files:** `src/Day.tsx`\n**Tests:** none\n",
	} {
		writeFile(t, filepath.Join(root, p), body)
	}
	gitRun(t, root, "init", "-q", "-b", "spectre/demo")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	writeFile(t, filepath.Join(root, "src/Day.tsx"), "export function Day() { return 1 }\n")
	r := runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", "printf '  3 passed (2s)\\n'; : <spec>"}, env)
	calls, _ := os.ReadFile(log)
	if want := `record run -C ` + regexp.QuoteMeta(root) + ` -change demo -kind recapture -label 4 -seconds \d+ -tests 3 -failed 0 -exit 0\n`; r.rc != 0 || !regexp.MustCompile(want).Match(calls) {
		t.Errorf("rc=%d out=%q flow calls %q, want %q", r.rc, r.out, calls, want)
	}
}

// TestRecaptureTaskCapturedStandsUntilASourceMoves pins rcTaskCaptured: a
// spec a task captured on its own commit stands for the whole-change pass,
// and stops standing once a later commit edits a file that reaches it -- a
// shared file its screen imports included -- or when no task captured it.
func TestRecaptureTaskCapturedStandsUntilASourceMoves(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	root := t.TempDir()
	for p, body := range map[string]string{
		"package.json":                "{}\n",
		".flow/project.md":            "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `src/**` |\n",
		"src/Day.tsx":                 "import { Chip } from \"./Chip\";\nexport function Day() {}\n",
		"src/Rooms.tsx":               "export function Rooms() {}\n",
		"src/Week.tsx":                "export function Week() {}\n",
		"src/Chip.tsx":                "export function Chip() {}\n",
		"shots/day.spec.ts":           "test('day', () => {})\n",
		"shots/day.spec.ts.sources":   "src/Day.tsx\n",
		"shots/rooms.spec.ts":         "test('rooms', () => {})\n",
		"shots/rooms.spec.ts.sources": "src/Rooms.tsx\n",
		"shots/week.spec.ts":          "test('week', () => {})\n",
		"shots/week.spec.ts.sources":  "src/Week.tsx\n",
		"plan/tasks.md":               "- [ ] 1. Day\n\n**Tests:** none\n\n- [ ] 2. Rooms\n\n**Tests:** none\n",
	} {
		writeFile(t, filepath.Join(root, p), body)
	}
	commit := func(msg string) {
		t.Helper()
		gitRun(t, root, "add", "-A")
		gitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", msg)
	}
	gitRun(t, root, "init", "-q")
	commit("base")
	base := vgHead(t, root)
	// task edits path, captures its task set, then commits.
	task := func(id, path string) {
		t.Helper()
		old, _ := os.ReadFile(filepath.Join(root, path))
		writeFile(t, filepath.Join(root, path), string(old)+"export const edited = '"+id+"'\n")
		if r := runGuard("recapture", []string{root + "/plan/tasks.md:" + id, root, root + "=HEAD", "true <spec>"}, env); r.rc != 0 {
			t.Fatalf("recapture of task %s: rc=%d out=%q", id, r.rc, r.out)
		}
		commit("task " + id)
	}
	stands := func() string {
		return strings.Join(rcTaskCaptured(env, root, map[string]string{root: base}), " ")
	}
	task("1", "src/Day.tsx")
	task("2", "src/Rooms.tsx")
	if got, want := stands(), "shots/day.spec.ts shots/rooms.spec.ts"; got != want {
		t.Fatalf("two tasks, each captured: %q, want %q", got, want)
	}
	// A shared file the day screen imports moves it; rooms still stands.
	writeFile(t, filepath.Join(root, "src/Chip.tsx"), "export function Chip() { return 1 }\n")
	commit("a later edit of a shared file")
	if got, want := stands(), "shots/rooms.spec.ts"; got != want {
		t.Errorf("after a shared file moved: %q, want %q", got, want)
	}
	// A screen no task captured never stands.
	writeFile(t, filepath.Join(root, "src/Week.tsx"), "export function Week() { return 1 }\n")
	commit("an uncaptured edit")
	if got, want := stands(), "shots/rooms.spec.ts"; got != want {
		t.Errorf("after an uncaptured edit: %q, want %q", got, want)
	}
}
