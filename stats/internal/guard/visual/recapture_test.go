package visual

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// TestRecapture pins flow-guard recapture: one run of the task's specs
// alone, a receipt for it, nothing run when no spec is affected, and the
// command's failure passed on.
func TestRecapture(t *testing.T) {
	t.Parallel()
	env := core.Env{Getenv: os.Getenv}
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
		guardtest.WriteFile(t, filepath.Join(root, p), body)
	}
	guardtest.GitRun(t, root, "init", "-q")
	guardtest.GitRun(t, root, "add", "-A")
	guardtest.GitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	ran := filepath.Join(t.TempDir(), "ran")
	run := func(command string) guardtest.GuardResult {
		return runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", command}, env)
	}
	receipts := func() string {
		raw, _ := os.ReadFile(filepath.Join(root, ".git/flow-recapture.jsonl"))
		return string(raw)
	}

	guardtest.WriteFile(t, filepath.Join(root, "docs/notes.md"), "edited\n")
	if r := run("echo <spec> >> " + ran); r.Rc != 0 || !strings.Contains(r.Stdout, "nothing to capture") || core.IsFile(ran) {
		t.Errorf("no affected spec: rc=%d out=%q", r.Rc, r.Out)
	}
	guardtest.WriteFile(t, filepath.Join(root, "src/Day.tsx"), "export function Day() { return 1 }\n")
	if r := run("echo <spec> >> " + ran); r.Rc != 0 || !strings.Contains(r.Stdout, "recapture: 1 test selection(s), exit 0") {
		t.Errorf("one affected spec: rc=%d out=%q", r.Rc, r.Out)
	}
	if got, _ := os.ReadFile(ran); string(got) != "shots/day.spec.ts --retries=1\n" {
		t.Errorf("the command ran with %q, want the day spec alone", got)
	}
	if got := receipts(); !strings.Contains(got, `"specs":["shots/day.spec.ts"],"exit":0`) || !strings.Contains(got, `"src/Day.tsx":"`) {
		t.Errorf("receipts = %q", got)
	}
	guardtest.WriteFile(t, filepath.Join(root, "src/Day.tsx"), "export function Day() { return 2 }\n")
	if r := run("echo <spec>; exit 3"); r.Rc != 1 || !strings.Contains(receipts(), `"exit":3`) {
		t.Errorf("a failing command: rc=%d out=%q receipts=%q", r.Rc, r.Out, receipts())
	}
	for _, args := range [][]string{{root, root + "=HEAD", "echo <spec>"}, {root + "/plan/tasks.md:4", root, root + "=HEAD", "no placeholder"}, {root + "/plan/tasks.md:4", root, root, "echo <spec>"}, {root + "/plan/tasks.md:9", root, root + "=HEAD", "echo <spec>"}} {
		if r := runGuard("recapture", args, env); r.Rc != 2 {
			t.Errorf("%v: rc=%d, want 2", args[1:], r.Rc)
		}
	}
}

// TestRecaptureTakesTheTaskSetAndRerunsOnlyWhatFailed pins the two halves of
// the kan-1091 fix: a fixture every spec renders selects the specs the task's
// Tests name and no other, and a red run is followed by one run of its
// failing test alone, after which nothing is left to run.
func TestRecaptureTakesTheTaskSetAndRerunsOnlyWhatFailed(t *testing.T) {
	t.Parallel()
	env := core.Env{Getenv: os.Getenv}
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
		guardtest.WriteFile(t, filepath.Join(root, p), body)
		if strings.HasSuffix(p, ".spec.ts") {
			guardtest.WriteFile(t, filepath.Join(root, p+".sources"), "src/"+filepath.Base(p)+".tsx\n")
		}
	}
	guardtest.GitRun(t, root, "init", "-q")
	guardtest.GitRun(t, root, "add", "-A")
	guardtest.GitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	guardtest.WriteFile(t, filepath.Join(root, "shots/fixtures.ts"), "export const seed = 2\n")
	ran := filepath.Join(t.TempDir(), "ran")
	run := func(command string) guardtest.GuardResult {
		return runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", command}, env)
	}
	red := "printf '  1 failed\\n    [chromium] › shots/rooms.spec.ts:3:1 › rooms\\n'; echo <spec> >> " + ran + "; exit 1"
	if r := run(red); r.Rc != 1 {
		t.Fatalf("the red run: rc=%d out=%q", r.Rc, r.Out)
	}
	if r := run("echo <spec> >> " + ran); r.Rc != 0 {
		t.Fatalf("the re-run: rc=%d out=%q", r.Rc, r.Out)
	}
	if r := run("echo <spec> >> " + ran); r.Rc != 0 || !strings.Contains(r.Stdout, "already passed") {
		t.Errorf("nothing left: rc=%d out=%q", r.Rc, r.Out)
	}
	const want = "shots/day.spec.ts shots/rooms.spec.ts --retries=1\nshots/rooms.spec.ts:3 --retries=1\n"
	if got, _ := os.ReadFile(ran); string(got) != want {
		t.Errorf("the command ran with %q, want %q", got, want)
	}
	// A red run that cannot name its failures failed every spec.
	guardtest.WriteFile(t, filepath.Join(root, "shots/fixtures.ts"), "export const seed = 3\n")
	run("echo <spec> >> " + ran + "; exit 1")
	run("echo <spec> >> " + ran)
	if got, _ := os.ReadFile(ran); !strings.HasSuffix(string(got), "shots/day.spec.ts shots/rooms.spec.ts --retries=1\nshots/day.spec.ts shots/rooms.spec.ts --retries=1\n") {
		t.Errorf("after an unnamed failure the command ran with %q, want both specs again", got)
	}
	// A retry-pass is flaky, never failed: nothing is left to run after it.
	guardtest.WriteFile(t, filepath.Join(root, "shots/fixtures.ts"), "export const seed = 4\n")
	r := run("printf '  1 flaky\\n    [chromium] › shots/rooms.spec.ts:3:1 › rooms\\n  1 passed\\n'; : <spec>")
	if r.Rc != 0 || !strings.Contains(r.Stdout, "recapture: flaky (1): shots/rooms.spec.ts:3 — passed on retry") {
		t.Errorf("a flaky run: rc=%d out=%q", r.Rc, r.Out)
	}
	if r := run("echo <spec> >> " + ran); !strings.Contains(r.Stdout, "already passed") {
		t.Errorf("after a flaky run: rc=%d out=%q, want nothing left", r.Rc, r.Out)
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
	env := core.Env{Getenv: func(k string) string {
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
		guardtest.WriteFile(t, filepath.Join(root, p), body)
	}
	guardtest.GitRun(t, root, "init", "-q", "-b", "spectre/demo")
	guardtest.GitRun(t, root, "add", "-A")
	guardtest.GitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	guardtest.WriteFile(t, filepath.Join(root, "src/Day.tsx"), "export function Day() { return 1 }\n")
	r := runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", "printf '  3 passed (2s)\\n'; : <spec>"}, env)
	calls, _ := os.ReadFile(log)
	if want := `record run -C ` + regexp.QuoteMeta(root) + ` -change demo -kind recapture -label 4 -seconds \d+ -tests 3 -failed 0 -exit 0 -flaky 0\n`; r.Rc != 0 || !regexp.MustCompile(want).Match(calls) {
		t.Errorf("rc=%d out=%q flow calls %q, want %q", r.Rc, r.Out, calls, want)
	}
}

// TestRecaptureTaskCapturedStandsUntilASourceMoves pins rcTaskCaptured: a
// spec a task captured on its own commit stands for the whole-change pass,
// and stops standing once a later commit edits a file that reaches it -- a
// shared file its screen imports included -- or when no task captured it.
func TestRecaptureTaskCapturedStandsUntilASourceMoves(t *testing.T) {
	t.Parallel()
	env := core.Env{Getenv: os.Getenv}
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
		guardtest.WriteFile(t, filepath.Join(root, p), body)
	}
	commit := func(msg string) {
		t.Helper()
		guardtest.GitRun(t, root, "add", "-A")
		guardtest.GitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", msg)
	}
	guardtest.GitRun(t, root, "init", "-q")
	commit("base")
	base := vgHead(t, root)
	// task edits path, captures its task set, then commits.
	task := func(id, path string) {
		t.Helper()
		old, _ := os.ReadFile(filepath.Join(root, path))
		guardtest.WriteFile(t, filepath.Join(root, path), string(old)+"export const edited = '"+id+"'\n")
		if r := runGuard("recapture", []string{root + "/plan/tasks.md:" + id, root, root + "=HEAD", "true <spec>"}, env); r.Rc != 0 {
			t.Fatalf("recapture of task %s: rc=%d out=%q", id, r.Rc, r.Out)
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
	guardtest.WriteFile(t, filepath.Join(root, "src/Chip.tsx"), "export function Chip() { return 1 }\n")
	commit("a later edit of a shared file")
	if got, want := stands(), "shots/rooms.spec.ts"; got != want {
		t.Errorf("after a shared file moved: %q, want %q", got, want)
	}
	// A screen no task captured never stands.
	guardtest.WriteFile(t, filepath.Join(root, "src/Week.tsx"), "export function Week() { return 1 }\n")
	commit("an uncaptured edit")
	if got, want := stands(), "shots/rooms.spec.ts"; got != want {
		t.Errorf("after an uncaptured edit: %q, want %q", got, want)
	}
}

// recaptureTitleFixture is a git repo whose task 4 names tests by title only:
// a capture root with tests/visual/a.spec.ts holding "goal sheet at rest" and
// tests/visual/b.spec.ts holding another, and no spec the diff reaches.
func recaptureTitleFixture(t *testing.T, tests string) string {
	t.Helper()
	root := t.TempDir()
	for p, body := range map[string]string{
		"package.json":           "{}\n",
		"src/Goal.tsx":           "export function Goal() {}\n",
		"tests/visual/a.spec.ts": "test(\"goal sheet at rest\", () => {})\ntest(\"goal sheet at rest open\", () => {})\n",
		"tests/visual/b.spec.ts": "test(\"goal sheet\", () => {})\n",
		"plan/tasks.md":          "- [ ] 4. In-run fix 1 — visual-verify: x\n\n**Files:** `src/Goal.tsx`\n**Tests:** " + tests + "\n",
	} {
		guardtest.WriteFile(t, filepath.Join(root, p), body)
	}
	guardtest.GitRun(t, root, "init", "-q")
	guardtest.GitRun(t, root, "add", "-A")
	guardtest.GitRun(t, root, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base")
	guardtest.WriteFile(t, filepath.Join(root, "src/Goal.tsx"), "export function Goal() { return 1 }\n")
	return root
}

// TestRecaptureSelectsSpecsByTitle pins that a task naming only a test title
// recaptures the spec whose source holds that title as a whole quoted string.
func TestRecaptureSelectsSpecsByTitle(t *testing.T) {
	t.Parallel()
	root := recaptureTitleFixture(t, "`goal sheet at rest`")
	ran := filepath.Join(t.TempDir(), "ran")
	r := runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", "echo <spec> >> " + ran}, core.Env{Getenv: os.Getenv})
	if r.Rc != 0 {
		t.Fatalf("rc=%d out=%q", r.Rc, r.Out)
	}
	if got, _ := os.ReadFile(ran); string(got) != "tests/visual/a.spec.ts --retries=1\n" {
		t.Errorf("the command ran with %q, want tests/visual/a.spec.ts alone", got)
	}
}

// TestRecaptureFailsWhenNoSpecHoldsATitle pins the loud failure: a title no
// spec holds exits 1 with one line, runs nothing and writes no receipt; a task
// naming no test still has nothing to capture.
func TestRecaptureFailsWhenNoSpecHoldsATitle(t *testing.T) {
	t.Parallel()
	root := recaptureTitleFixture(t, "`no such title`")
	ran := filepath.Join(t.TempDir(), "ran")
	r := runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", "echo <spec> >> " + ran}, core.Env{Getenv: os.Getenv})
	want := "recapture: task 4 names tests no spec under " + root + " holds: no such title\n"
	if r.Rc != 1 || r.Stdout != want || core.IsFile(ran) || core.IsFile(filepath.Join(root, ".git/flow-recapture.jsonl")) {
		t.Errorf("rc=%d stdout=%q ran=%v, want rc 1, %q, nothing run, no receipt", r.Rc, r.Stdout, core.IsFile(ran), want)
	}

	for _, named := range []string{"none", "`TestGoalSheet`, `a.spec.ts:12`, `x := f()`"} {
		root = recaptureTitleFixture(t, named)
		r = runGuard("recapture", []string{root + "/plan/tasks.md:4", root, root + "=HEAD", "echo <spec> >> " + ran}, core.Env{Getenv: os.Getenv})
		if r.Rc != 0 || r.Stdout != "recapture: no affected spec — nothing to capture\n" || core.IsFile(ran) {
			t.Errorf("no title named in %q: rc=%d stdout=%q", named, r.Rc, r.Stdout)
		}
	}
}

// TestRcBlobsMatchesRevParse: rcBlobs' one ls-tree reads, for every path, what
// `git rev-parse --verify --quiet <rev>:<path>` printed per path -- a nested
// file, a name with a space, tab and non-ASCII byte, a path the rev lacks, a
// path that is a directory there.
func TestRcBlobsMatchesRevParse(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	guardtest.FixtureRepo(t, wt, guardtest.RepoShape{Branch: "main"})
	paths := []string{"top.txt", "a/b/nested.txt", "sp ace\tтаб.txt", "gone.txt", "a/b"}
	for _, p := range paths[:3] {
		guardtest.WriteFile(t, filepath.Join(wt, p), p+"\n")
	}
	guardtest.GitRun(t, wt, "add", "-A")
	guardtest.GitRun(t, wt, "commit", "-qm", "x")
	git := core.EnvGit(core.Env{Getenv: os.Getenv, Dir: wt})
	got := rcBlobs(git, wt, "HEAD", paths)
	for _, p := range paths {
		want, _ := core.Capture(git("-C", wt, "rev-parse", "--verify", "--quiet", "HEAD:"+p))
		if got[p] != want {
			t.Errorf("%q: rcBlobs %q, rev-parse %q", p, got[p], want)
		}
	}
	if len(got) != len(paths) {
		t.Errorf("rcBlobs keys %v, want one per path", got)
	}
}
