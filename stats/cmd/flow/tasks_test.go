package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
)

func writePlan(t *testing.T, dir, changesLeaf, change, body string) {
	t.Helper()
	planDir := filepath.Join(dir, changesLeaf, "changes", change)
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestTasksTickWritesFile(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n\n  - [ ] **Step 1: a**\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "ticked task 1 (1 steps)") {
		t.Errorf("stdout = %q", stdout.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "spectre", "changes", "c", "tasks.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(got), "- [x] 1. First") || !strings.Contains(string(got), "- [x] **Step 1: a**") {
		t.Errorf("file not ticked:\n%s", got)
	}
}

func TestTasksTickAlreadyTickedExitsNonZero(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [x] 1. First\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1"}, nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("code = 0, want non-zero; stderr = %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "already ticked") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestTasksTickTaskNotFoundExitsNonZero(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "9"}, nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("code = 0, want non-zero; stderr = %s", stderr.String())
	}
}

func TestTasksTickInvalidIDExitsTwo(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1.1"}, nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
	}
}

func TestTasksTickMissingPlanExitsNonZero(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1"}, nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("code = 0, want non-zero; stderr = %s", stderr.String())
	}
}

func TestTasksTickBadArgCountExitsTwo(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c"}, nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
	}
}

func TestPlanPathRejectsContainmentEscapes(t *testing.T) {
	dir := t.TempDir()
	for _, change := range []string{"../escape", ".hidden", "-flag", "_underscore", ""} {
		if _, err := planPath(dir, change); err == nil {
			t.Errorf("planPath(%q) = nil error, want rejection", change)
		}
	}
}

func TestPlanPathAcceptsValidName(t *testing.T) {
	dir := t.TempDir()
	got, err := planPath(dir, "valid-name-123")
	if err != nil {
		t.Fatalf("planPath: %v", err)
	}
	want := filepath.Join(dir, "spectre", "changes", "valid-name-123", "tasks.md")
	if got != want {
		t.Errorf("planPath = %q, want %q", got, want)
	}
}

func TestTasksTickRejectsContainmentEscapesEndToEnd(t *testing.T) {
	dir := t.TempDir()
	for _, change := range []string{"../escape", ".hidden", "-flag", "_underscore"} {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), []string{"tasks", "tick", "-C", dir, change, "1"}, nil, &stdout, &stderr)
		if code != 2 {
			t.Errorf("change %q: code = %d, want 2; stderr = %s", change, code, stderr.String())
		}
	}
}

func TestSpecRootLeafPrefersSpectre(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "spectre", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := specRootLeaf(dir); got != "spectre" {
		t.Errorf("specRootLeaf = %q, want spectre", got)
	}
}

func TestSpecRootLeafFallsBackToOpenspec(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := specRootLeaf(dir); got != "openspec" {
		t.Errorf("specRootLeaf = %q, want openspec", got)
	}
}

func TestSpecRootLeafDefaultsToSpectreWhenNeitherExists(t *testing.T) {
	dir := t.TempDir()
	if got := specRootLeaf(dir); got != "spectre" {
		t.Errorf("specRootLeaf = %q, want spectre", got)
	}
}

func TestSpecRootLeafIgnoresFileAtSpectrePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "spectre"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A plain file at <dir>/spectre/changes must not count as the
	// spectre tree existing -- only a directory does.
	if err := os.WriteFile(filepath.Join(dir, "spectre", "changes"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "changes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := specRootLeaf(dir); got != "openspec" {
		t.Errorf("specRootLeaf = %q, want openspec", got)
	}
}

func TestTasksTickHelpExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-h"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr = %s", code, stderr.String())
	}
}

func TestTasksTickExtraPositionalArgExitsTwo(t *testing.T) {
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", "- [ ] 1. First\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"tasks", "tick", "-C", dir, "c", "1", "bogus-extra-arg"}, nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2; stderr = %s", code, stderr.String())
	}
}

func TestPlanPathRejectsTrailingCharsAfterValidPrefix(t *testing.T) {
	dir := t.TempDir()
	for _, change := range []string{"valid/x", "valid/../../etc/passwd"} {
		if _, err := planPath(dir, change); err == nil {
			t.Errorf("planPath(%q) = nil error, want rejection", change)
		}
	}
}

// TestTasksCountRecordsThePlanTaskCount pins the command's core: the
// column-0 task checkbox lines are counted (fenced examples never are),
// the figure is POSTed as one observation under the change's own
// task-counts collection, and the recorded row's stamps come back on
// stdout.
func TestTasksCountRecordsThePlanTaskCount(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	writePlan(t, repo, "spectre", "c",
		"- [ ] 1. First\n\n  - [ ] **Step 1: a**\n\n- [ ] 2. Second\n"+
			"```text\n- [ ] 9. Fenced example is documentation\n```\n")

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}

	var gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotBody, _ = readAll(r)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"totalTasks":2,"observedAt":"2026-09-15T12:00:00Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"tasks", "count", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "c"},
		nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("tasks count exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "recorded: task-count 7 (total 2)") {
		t.Errorf("stdout = %q", stdout.String())
	}
	wantPath := "/api/v1/records/" + projectKey + "/c/task-counts"
	if gotMethod != http.MethodPost || gotPath != wantPath {
		t.Errorf("record request = %s %s, want POST %s", gotMethod, gotPath, wantPath)
	}

	var sent records.TaskCount
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode record body %s: %v", gotBody, err)
	}
	if sent.TotalTasks != 2 {
		t.Errorf("totalTasks = %d, want 2 -- the real task lines, never the fenced one", sent.TotalTasks)
	}
}

// TestTasksCountPostsCorrections: the plan's `Correction (YYYY-MM-DD):`
// paragraphs are counted beside its tasks and posted in the same
// observation -- the plan-corrections metric experiments compare
// (KAN-1067). A Correction line shown inside a fenced example is
// documentation, never counted, the same fence rule the task count keeps.
func TestTasksCountPostsCorrections(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	writePlan(t, repo, "spectre", "c",
		"- [ ] 1. First\n\n  Correction (2026-10-29): the path was wrong.\n\n"+
			"- [ ] 2. Second\n\nCorrection (2026-10-30): the test name was wrong.\n\n"+
			"```text\nCorrection (YYYY-MM-DD): template\nCorrection (2026-01-01): fenced example\n```\n")

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = readAll(r)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"totalTasks":2,"corrections":2,"observedAt":"2026-09-15T12:00:00Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"tasks", "count", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "c"},
		nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("tasks count exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var sent map[string]any
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode record body %s: %v", gotBody, err)
	}
	if sent["totalTasks"] != float64(2) || sent["corrections"] != float64(2) {
		t.Errorf("record body = %s, want totalTasks 2 and corrections 2 -- never the fenced lines", gotBody)
	}
}

// TestTasksCountUnreachableStoreExitsZero: an observation the store cannot
// take costs one warning line and changes no exit code, the `flow suite
// record` rule -- a plan-counting stage must not fail because flowd is
// down.
func TestTasksCountUnreachableStoreExitsZero(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	writePlan(t, repo, "spectre", "c", "- [ ] 1. First\n- [ ] 2. Second\n")

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"tasks", "count", "-addr", deadPortURL(t), "-timeout", "500ms", "-C", repo, "c"},
		nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("tasks count exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "not recorded") || !strings.Contains(stderr.String(), "2") {
		t.Errorf("stderr = %q, want one warning naming the unrecorded count", stderr.String())
	}
}

// TestTasksCountBadChangeNameExitsTwo: a change name that fails the
// containment allowlist is a caller mistake refused before the file
// system or the store is touched, exactly as tasks tick refuses it.
func TestTasksCountBadChangeNameExitsTwo(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"tasks", "count", "-C", repo, "../escape"},
		nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("tasks count exit = %d, want 2; stderr:\n%s", code, stderr.String())
	}
}

// TestTasksCountTakesFlagsAfterTheChangeName: `tasks count <name> -C <dir>`
// parses the trailing flag rather than refusing it as a second argument.
func TestTasksCountTakesFlagsAfterTheChangeName(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	writePlan(t, repo, "spectre", "c", "- [ ] 1. First\n\n- [ ] 2. Second\n")

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"totalTasks":2,"observedAt":"2026-09-15T12:00:00Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"tasks", "count", "-addr", srv.URL, "-timeout", "500ms", "c", "-C", repo},
		nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("tasks count exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "recorded: task-count 7 (total 2)") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

// TestTasksFilesAddWidensOneTask pins decision sr16-files-verb: `flow tasks
// files <change> <id> add <path>...` widens that one task's **Files:** field,
// deduped, never another task's identical line nor a fenced one, and leaves
// every other byte as it was.
func TestTasksFilesAddWidensOneTask(t *testing.T) {
	const plan = "# p\n\n" +
		"- [ ] 1. One\n  - [ ] **Step 1: a**\n\n```markdown\n**Files:** `fenced.go`\n```\n\n**Files:** `x.go`\n**Commit:** one\n\n" +
		"- [ ] 2. Two\n\n**Files:** none — no code\n\n" +
		"- [ ] 3. Three\n\n**Files:** `x.go`\n"
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", plan)
	path := filepath.Join(dir, "spectre", "changes", "c", "tasks.md")
	files := func(args ...string) int {
		var stdout, stderr bytes.Buffer
		return run(context.Background(), append([]string{"tasks", "files", "-C", dir, "c"}, args...), nil, &stdout, &stderr)
	}
	read := func() string {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if code := files("1", "add", "a.go", "b.go"); code != 0 {
		t.Fatalf("add: exit %d", code)
	}
	want := strings.Replace(plan, "**Files:** `x.go`\n**Commit:** one", "**Files:** `x.go`, `a.go`, `b.go`\n**Commit:** one", 1)
	if got := read(); got != want {
		t.Fatalf("after add:\n%s\nwant:\n%s", got, want)
	}
	if code := files("1", "add", "a.go", "x.go"); code != 0 || read() != want {
		t.Fatalf("a second add changed the plan (exit %d):\n%s", code, read())
	}
	if code := files("2", "add", "c.go"); code != 0 {
		t.Fatalf("add to none: exit %d", code)
	}
	want = strings.Replace(want, "**Files:** none — no code", "**Files:** `c.go`", 1)
	if got := read(); got != want {
		t.Fatalf("after add to none:\n%s\nwant:\n%s", got, want)
	}
	if code := files("9", "add", "d.go"); code != 1 {
		t.Fatalf("an unknown task: exit %d, want 1", code)
	}
}

func TestTasksFilesRemoveNarrowsOneTask(t *testing.T) {
	const plan = "# p\n\n" +
		"- [ ] 1. One\n  - [ ] **Step 1: a**\n\n```markdown\n**Files:** `x.go`, `a.go`, `b.go`\n```\n\n**Files:** `x.go`, `a.go`, `b.go`\n**Commit:** one\n\n" +
		"- [ ] 3. Three\n\n**Files:** `x.go`, `a.go`, `b.go`\n"
	dir := t.TempDir()
	writePlan(t, dir, "spectre", "c", plan)
	path := filepath.Join(dir, "spectre", "changes", "c", "tasks.md")
	files := func(args ...string) int {
		var stdout, stderr bytes.Buffer
		return run(context.Background(), append([]string{"tasks", "files", "-C", dir, "c"}, args...), nil, &stdout, &stderr)
	}
	read := func() string {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if code := files("1", "remove", "a.go"); code != 0 {
		t.Fatalf("remove: exit %d", code)
	}
	want := strings.Replace(plan, "**Files:** `x.go`, `a.go`, `b.go`\n**Commit:** one", "**Files:** `x.go`, `b.go`\n**Commit:** one", 1)
	if got := read(); got != want {
		t.Fatalf("after remove:\n%s\nwant:\n%s", got, want)
	}
	if code := files("1", "remove", "x.go", "b.go"); code != 0 {
		t.Fatalf("remove the rest: exit %d", code)
	}
	want = strings.Replace(want, "**Files:** `x.go`, `b.go`\n**Commit:** one", "**Files:** none\n**Commit:** one", 1)
	if got := read(); got != want {
		t.Fatalf("after removing the rest:\n%s\nwant:\n%s", got, want)
	}
	if code := files("1", "remove", "q.go"); code != 1 || read() != want {
		t.Fatalf("an unnamed path: exit %d, want 1, plan:\n%s", code, read())
	}
	if code := files("9", "remove", "a.go"); code != 1 {
		t.Fatalf("an unknown task: exit %d, want 1", code)
	}
	if code := files("1", "remove"); code != 2 {
		t.Fatalf("no path: exit %d, want 2", code)
	}
}
