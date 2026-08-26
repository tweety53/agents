package fallback_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func initRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("commit", "--allow-empty", "-q", "-m", "initial")
}

// --- ProjectKey ---

// contractPath is skills/myflow-contracts/state-file.md's location relative
// to this package -- stats/internal/fallback is three directories below the
// repository root (fallback -> internal -> stats -> root), and the myflow
// skills live at the root alongside the stats Go module, not inside it.
const contractPath = "../../../skills/myflow-contracts/state-file.md"

// contractProjectKeyRecipe reads the MAIN_CHECKOUT= and PROJECT_KEY=
// assignment lines directly out of skills/myflow-contracts/state-file.md,
// rather than a hand-copied string literal in this test file. F3's review
// finding is exactly what a hand-copied literal invites: a prior version
// of this test had its own copy of the recipe, which got silently edited
// to agree with this package's code instead of being checked against the
// contract -- the copy and the contract could disagree indefinitely and
// nothing would notice. Reading the two assignment lines out of the
// contract file itself closes that gap: if the contract's recipe changes,
// this test changes with it or fails loudly, and there is exactly one
// place (the contract) that states the algorithm.
//
// Only the two derivation lines are extracted, not the whole fenced block:
// the block's remaining lines build STATE_FILE under the real, hardcoded
// `/Users/tweety53/Agents/flow/state` path and `mkdir -p` it, which this
// test must never execute -- it would create real directories on the
// operator's machine outside any test sandbox.
func contractProjectKeyRecipe(t *testing.T) (mainCheckoutLine, projectKeyLine string) {
	t.Helper()
	data, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("read contract %s: %v", contractPath, err)
	}
	text := string(data)

	idx := strings.Index(text, "MAIN_CHECKOUT=")
	if idx == -1 {
		t.Fatalf("contract %s no longer contains a MAIN_CHECKOUT= assignment -- the recipe this test pins against has moved or been renamed", contractPath)
	}
	lines := strings.SplitN(text[idx:], "\n", 3)
	if len(lines) < 2 {
		t.Fatalf("contract %s: MAIN_CHECKOUT= line has no following line", contractPath)
	}
	mainCheckoutLine = strings.TrimSpace(lines[0])
	projectKeyLine = strings.TrimSpace(lines[1])
	if !strings.HasPrefix(projectKeyLine, "PROJECT_KEY=") {
		t.Fatalf("contract %s: expected a PROJECT_KEY= line immediately after MAIN_CHECKOUT=, got %q -- the recipe's shape changed and this test needs updating alongside it", contractPath, projectKeyLine)
	}
	return mainCheckoutLine, projectKeyLine
}

func TestProjectKeyMatchesStateFileContract(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	initRepo(t, dir)

	key, mainCheckout, err := fallback.ProjectKey(dir)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}

	mainCheckoutLine, projectKeyLine := contractProjectKeyRecipe(t)
	script := mainCheckoutLine + "\n" + projectKeyLine + "\n" + `printf '%s\n%s' "$PROJECT_KEY" "$MAIN_CHECKOUT"` + "\n"

	cmd := exec.Command("bash", "-c", "set -e\n"+script)
	cmd.Dir = dir // the extracted lines assume cwd is already the target directory, exactly as the contract itself assumes.
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("contract recipe (extracted from %s):\n%s\nerror: %v", contractPath, script, err)
	}
	parts := strings.SplitN(string(out), "\n", 2)
	if len(parts) != 2 {
		t.Fatalf("contract recipe produced unexpected output: %q", out)
	}
	wantKey, wantMainCheckout := parts[0], parts[1]

	if key != wantKey {
		t.Errorf("ProjectKey key = %q, want %q (contract recipe, extracted from %s)", key, wantKey, contractPath)
	}
	if mainCheckout != wantMainCheckout {
		t.Errorf("ProjectKey mainCheckout = %q, want %q (contract recipe, extracted from %s)", mainCheckout, wantMainCheckout, contractPath)
	}
}

func TestProjectKeyIsIdenticalFromAWorktree(t *testing.T) {
	requireGit(t)
	mainDir := t.TempDir()
	initRepo(t, mainDir)

	worktreeParent := t.TempDir()
	worktreeDir := filepath.Join(worktreeParent, "wt")
	cmd := exec.Command("git", "worktree", "add", "-q", worktreeDir)
	cmd.Dir = mainDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}

	mainKey, mainCheckout, err := fallback.ProjectKey(mainDir)
	if err != nil {
		t.Fatalf("ProjectKey(main): %v", err)
	}
	wtKey, wtMainCheckout, err := fallback.ProjectKey(worktreeDir)
	if err != nil {
		t.Fatalf("ProjectKey(worktree): %v", err)
	}

	if mainKey != wtKey {
		t.Errorf("project key differs between main checkout (%q) and worktree (%q)", mainKey, wtKey)
	}
	if mainCheckout != wtMainCheckout {
		t.Errorf("main checkout differs between main checkout call (%q) and worktree call (%q)", mainCheckout, wtMainCheckout)
	}
}

// --- state file read/write ---

func TestWriteStateFileThenReadRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "proj", "change.json")
	body := []byte(`{"state":"IN_PROGRESS"}`)

	if err := fallback.WriteStateFile(path, body); err != nil {
		t.Fatalf("WriteStateFile: %v", err)
	}
	got, err := fallback.ReadStateFile(path)
	if err != nil {
		t.Fatalf("ReadStateFile: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("round trip: got %s, want %s", got, body)
	}
}

func TestReadStateFileMissingIsAnError(t *testing.T) {
	_, err := fallback.ReadStateFile(filepath.Join(t.TempDir(), "absent.json"))
	if err == nil {
		t.Fatal("expected an error for a missing state file, got nil")
	}
}

// --- ListStateFileNames ---
//
// cmd/flow's `state list` (F1's fix) reads the fallback directory
// through this function when the store cannot be reached -- these tests
// are its own coverage, independent of the CLI command built on top of it.

func TestListStateFileNamesFindsEveryJSONFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("FLOW_STATE_DIR", root)

	for _, name := range []string{"kan-1", "kan-2", "kan-3"} {
		if err := fallback.WriteStateFile(fallback.StateFilePath("proj", name), []byte(`{}`)); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	// A journal file beside them must not be reported as a change name.
	if err := fallback.AppendJournalEntry(fallback.JournalFilePath("proj", "kan-1"), "proj", "kan-1", []byte(`{}`), time.Now()); err != nil {
		t.Fatalf("seed journal: %v", err)
	}

	names, err := fallback.ListStateFileNames("proj")
	if err != nil {
		t.Fatalf("ListStateFileNames: %v", err)
	}
	got := map[string]bool{}
	for _, n := range names {
		got[n] = true
	}
	for _, want := range []string{"kan-1", "kan-2", "kan-3"} {
		if !got[want] {
			t.Errorf("names = %v, missing %q", names, want)
		}
	}
	if len(names) != 3 {
		t.Errorf("names = %v, want exactly 3 (the journal file must not count)", names)
	}
}

func TestListStateFileNamesMissingDirectoryIsNotAnError(t *testing.T) {
	root := t.TempDir()
	t.Setenv("FLOW_STATE_DIR", root)

	names, err := fallback.ListStateFileNames("never-written-to")
	if err != nil {
		t.Fatalf("ListStateFileNames: %v, want nil error for a project with no fallback directory yet", err)
	}
	if len(names) != 0 {
		t.Errorf("names = %v, want empty", names)
	}
}

// --- journal ---

func TestAppendJournalEntryWritesJournalEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proj", "change.journal")
	body := []byte(`{"state":"IN_PROGRESS","updatedAt":"2026-08-13T10:00:00Z"}`)
	now := time.Date(2026, 8, 13, 10, 0, 1, 0, time.UTC)

	if err := fallback.AppendJournalEntry(path, "myrepo-abcd1234", "kan-16", body, now); err != nil {
		t.Fatalf("AppendJournalEntry: %v", err)
	}

	entries, err := fallback.ReadJournalEntries(path)
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.Project != "myrepo-abcd1234" || e.Name != "kan-16" {
		t.Errorf("Project/Name = %q/%q, want myrepo-abcd1234/kan-16", e.Project, e.Name)
	}
	if !e.RecordedAt.Equal(now) {
		t.Errorf("RecordedAt = %v, want %v", e.RecordedAt, now)
	}
	var gotBody, wantBody map[string]any
	if err := json.Unmarshal(e.Body, &gotBody); err != nil {
		t.Fatalf("decode entry body: %v", err)
	}
	if err := json.Unmarshal(body, &wantBody); err != nil {
		t.Fatalf("decode want body: %v", err)
	}
	if gotBody["state"] != wantBody["state"] {
		t.Errorf("entry Body state = %v, want %v", gotBody["state"], wantBody["state"])
	}
}

func TestAppendJournalEntryAppendsInCallOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proj", "change.journal")
	now := time.Now()

	for i := 0; i < 3; i++ {
		body := []byte(fmt.Sprintf(`{"n":%d}`, i))
		if err := fallback.AppendJournalEntry(path, "p", "n", body, now); err != nil {
			t.Fatalf("AppendJournalEntry #%d: %v", i, err)
		}
	}

	entries, err := fallback.ReadJournalEntries(path)
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("len(entries) = %d, want 3", len(entries))
	}
	for i, e := range entries {
		want := fmt.Sprintf(`{"n":%d}`, i)
		if !jsonEqualBytes(t, e.Body, []byte(want)) {
			t.Errorf("entries[%d].Body = %s, want %s (replay order is call order)", i, e.Body, want)
		}
	}
}

// TestAppendJournalEntryIsSafeUnderConcurrentWriters reproduces the
// scenario F2 found: parallel worktrees is a real shape of this pipeline,
// so concurrent `myflow state set` invocations appending to the same
// project's journal at the same time is expected, not hypothetical. Every
// one of goroutineCount concurrent appends must land as its own complete
// entry -- O_APPEND's atomicity is what this function relies on for safety
// under concurrent writers instead of a lock (see AppendJournalEntry's doc
// comment), and this test is what would catch that guarantee breaking
// (a torn write producing an unparseable line, or a lost append producing
// fewer than goroutineCount entries) if a future change replaced the
// O_APPEND flag with something not atomic across processes.
func TestAppendJournalEntryIsSafeUnderConcurrentWriters(t *testing.T) {
	const goroutineCount = 20
	path := filepath.Join(t.TempDir(), "proj", "change.journal")
	now := time.Now()

	var wg sync.WaitGroup
	errCh := make(chan error, goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := []byte(fmt.Sprintf(`{"n":%d}`, i))
			errCh <- fallback.AppendJournalEntry(path, "p", "n", body, now)
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("AppendJournalEntry: %v", err)
		}
	}

	entries, err := fallback.ReadJournalEntries(path)
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != goroutineCount {
		t.Fatalf("len(entries) = %d, want %d -- no entry may be lost or torn under concurrent writers", len(entries), goroutineCount)
	}

	seen := map[int]bool{}
	for _, e := range entries {
		var v struct {
			N int `json:"n"`
		}
		if err := json.Unmarshal(e.Body, &v); err != nil {
			t.Fatalf("entry body did not parse (a torn write?): %s: %v", e.Body, err)
		}
		if seen[v.N] {
			t.Fatalf("n=%d appeared twice", v.N)
		}
		seen[v.N] = true
	}
	if len(seen) != goroutineCount {
		t.Fatalf("saw %d distinct entries, want %d", len(seen), goroutineCount)
	}
}

func jsonEqualBytes(t *testing.T, a, b []byte) bool {
	t.Helper()
	var av, bv any
	if err := json.Unmarshal(a, &av); err != nil {
		t.Fatalf("decode a: %v (%s)", err, a)
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		t.Fatalf("decode b: %v (%s)", err, b)
	}
	aj, _ := json.Marshal(av)
	bj, _ := json.Marshal(bv)
	return string(aj) == string(bj)
}

func TestReadJournalEntriesMissingFileIsNotAnError(t *testing.T) {
	entries, err := fallback.ReadJournalEntries(filepath.Join(t.TempDir(), "absent.journal"))
	if err != nil {
		t.Fatalf("ReadJournalEntries on a missing journal: %v", err)
	}
	if entries != nil {
		t.Errorf("entries = %v, want nil", entries)
	}
}

func TestJournalIsOneEntryPerLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proj", "change.journal")
	now := time.Now()
	if err := fallback.AppendJournalEntry(path, "p", "n", []byte(`{"a":1}`), now); err != nil {
		t.Fatalf("AppendJournalEntry: %v", err)
	}
	if err := fallback.AppendJournalEntry(path, "p", "n", []byte(`{"a":2}`), now); err != nil {
		t.Fatalf("AppendJournalEntry: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read journal file: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("len(lines) = %d, want 2 (one JSON object per line)", len(lines))
	}
	for i, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Errorf("line %d is not valid JSON on its own: %s", i, line)
		}
	}
}
