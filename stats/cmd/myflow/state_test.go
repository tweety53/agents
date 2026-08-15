package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/stages"
)

// --- test git repo fixture ---

// gitRepo creates a real git repository in a fresh temp directory and
// returns its path -- required by TestProjectKeyIsIdenticalFromAWorktree's
// own instruction to use a real worktree, not a simulated path, and used
// as the -C target by every other test here so ProjectKey has something
// real to resolve.
func gitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
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
	return dir
}

// isolatedStateRoot points MYFLOW_STATE_DIR at a fresh temp directory for
// the duration of the test, so a test never touches the real
// /Users/tweety53/Agents/myflow/state tree.
func isolatedStateRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("MYFLOW_STATE_DIR", root)
	return root
}

func deadPortAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := "http://" + ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	return addr
}

// daemonHeaderName and daemonHeaderValue must match internal/client's own
// (unexported) constants and internal/api.DaemonHeader / DaemonHeaderValue
// exactly. See internal/api.DaemonHeader's doc comment for why this header
// exists: F1's fix for a look-alike server on the configured port
// answering with a status the CLI would otherwise trust as a genuine
// store answer.
const (
	daemonHeaderName  = "Myflow-Daemon"
	daemonHeaderValue = "myflowd/1"
)

// genuineDaemon wraps handler so every response it writes carries the
// header a real myflowd sets on every response. Tests exercising the CLI
// against a store that genuinely answered use this; tests exercising F1's
// look-alike-server fallback deliberately do not.
func genuineDaemon(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(daemonHeaderName, daemonHeaderValue)
		handler(w, r)
	}
}

// countWarningLines counts non-empty lines in s. The exit-0 fallback
// guarantee requires exactly one -- not zero (silently swallowed), not a
// stack trace's worth.
func countLines(s string) int {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

// --- state set: the fallback / never-block guarantee, one failure mode at a time ---

func TestStateSetFallsBackOnDeadPort(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(`{"state":"IN_PROGRESS","updatedAt":"2026-08-13T10:00:00Z"}`),
		&stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (dead port must never block); stderr:\n%s", code, stderr.String())
	}
	if got := countLines(stderr.String()); got != 1 {
		t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
	}
}

func TestStateSetFallsBackOnTimeout(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", srv.URL, "-timeout", "20ms", "-C", repo, "kan-16"},
		strings.NewReader(`{"state":"IN_PROGRESS","updatedAt":"2026-08-13T10:00:00Z"}`),
		&stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a slow daemon must never block); stderr:\n%s", code, stderr.String())
	}
	if got := countLines(stderr.String()); got != 1 {
		t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
	}
}

func TestStateSetFallsBackOnNon2xxThatIsNotRefusal(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(`{"state":"IN_PROGRESS","updatedAt":"2026-08-13T10:00:00Z"}`),
		&stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a 500 is not a refusal); stderr:\n%s", code, stderr.String())
	}
	if got := countLines(stderr.String()); got != 1 {
		t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
	}
}

func TestStateSetFallsBackOnMalformedResponse(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`this is not json`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(`{"state":"IN_PROGRESS","updatedAt":"2026-08-13T10:00:00Z"}`),
		&stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a malformed response must fall back); stderr:\n%s", code, stderr.String())
	}
	if got := countLines(stderr.String()); got != 1 {
		t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
	}
}

func TestStateSetFallbackExitsZero(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "kan-16"},
		strings.NewReader(`{"state":"IN_PROGRESS","updatedAt":"2026-08-13T10:00:00Z"}`),
		&stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want exactly 0", code)
	}
}

func TestStateSetFallbackWritesStateFileAndJournalEntry(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	body := `{"state":"IN_PROGRESS","updatedAt":"2026-08-13T10:00:00Z"}`
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "kan-16"},
		strings.NewReader(body),
		&stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}

	stateBody, err := fallback.ReadStateFile(fallback.StateFilePath(projectKey, "kan-16"))
	if err != nil {
		t.Fatalf("ReadStateFile: %v", err)
	}
	if !jsonEqual(t, stateBody, []byte(body)) {
		t.Errorf("on-disk state file = %s, want %s", stateBody, body)
	}

	entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-16"))
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(journal entries) = %d, want 1", len(entries))
	}
	if !jsonEqual(t, entries[0].Body, []byte(body)) {
		t.Errorf("journal entry body = %s, want %s", entries[0].Body, body)
	}
	if entries[0].Project != projectKey || entries[0].Name != "kan-16" {
		t.Errorf("journal entry identity = %s/%s, want %s/kan-16", entries[0].Project, entries[0].Name, projectKey)
	}
}

// --- state set: F4 -- a non-object JSON body is a usage error, never storable ---

// TestStateSetRejectsNullBody reproduces F4: a bare JSON `null` satisfies
// json.Valid, previously reached withMainCheckoutPath, and panicked
// assigning into the nil map that null decodes into. The panic was
// recovered (never-block held, exit 0), but the literal `null` was then
// written to both the state file and the journal as if it were a real
// record. It must instead be rejected before either write, as a usage
// error -- neither the state file nor the journal may exist afterward.
func TestStateSetRejectsNullBody(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "kan-16"},
		strings.NewReader("null"),
		&stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (a null body is a usage error); stderr:\n%s", code, stderr.String())
	}

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	if _, err := fallback.ReadStateFile(fallback.StateFilePath(projectKey, "kan-16")); err == nil {
		t.Error("a null body was written to the state file -- it must be rejected before either write")
	}
	entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-16"))
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("journal has %d entries, want 0 -- a null body must never be journaled", len(entries))
	}
}

// TestStateSetRejectsNonObjectJSONBody covers the other non-object JSON
// shapes json.Valid alone would have let through: an array, a bare string,
// and a bare number.
func TestStateSetRejectsNonObjectJSONBody(t *testing.T) {
	repo := gitRepo(t)

	for _, body := range []string{`[1,2,3]`, `"just a string"`, `42`, `true`} {
		t.Run(body, func(t *testing.T) {
			isolatedStateRoot(t)
			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"state", "set", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "kan-16"},
				strings.NewReader(body),
				&stdout, &stderr)
			if code != 2 {
				t.Fatalf("body %s: exit code = %d, want 2; stderr:\n%s", body, code, stderr.String())
			}
		})
	}
}

// --- state set: a monotonic refusal is not a fallback trigger ---

func TestMonotonicRefusalIsNotAFallback(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"store: refused: write would move state backwards"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(`{"state":"STARTED","updatedAt":"2026-08-13T10:00:00Z"}`),
		&stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero: a refusal is not a fallback trigger, it is the store correctly answering \"no\"")
	}

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	if _, err := fallback.ReadStateFile(fallback.StateFilePath(projectKey, "kan-16")); err == nil {
		t.Error("state file was written on a refusal -- a refusal must not take the fallback path")
	}
	entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-16"))
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("journal has %d entries, want 0 -- a refusal must not be journaled", len(entries))
	}
	if strings.Contains(stderr.String(), "unreachable") {
		t.Errorf("stderr reports the refusal as unreachable, which conflates the two: %s", stderr.String())
	}
}

// --- state set: F1 -- a look-alike 409 from a non-daemon server must fall back ---

// TestStateSetTreatsForeignServer409AsFallback is the CLI-level
// reproduction of F1: pointing `-addr` at a bare server that answers every
// PUT with 409 -- no myflowd behind it, no DaemonHeader on the response --
// previously made the CLI trust it as a genuine monotonic refusal and
// exit 1, stopping the pipeline with no store involved at all. It must
// instead take the fallback path: state file and journal written, exactly
// one warning line, exit 0.
func TestStateSetTreatsForeignServer409AsFallback(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A bare foreign server: no Myflow-Daemon header, ever -- the
		// generic shape of "something is listening on this port, and it
		// happens to answer 409" that F1 was found against.
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`Conflict`))
	}))
	defer srv.Close()

	body := `{"state":"IN_PROGRESS","updatedAt":"2026-08-13T10:00:00Z"}`
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(body),
		&stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a look-alike 409 must fall back, not block); stderr:\n%s", code, stderr.String())
	}
	if got := countLines(stderr.String()); got != 1 {
		t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
	}

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	if _, err := fallback.ReadStateFile(fallback.StateFilePath(projectKey, "kan-16")); err != nil {
		t.Errorf("state file was not written on the fallback path: %v", err)
	}
	entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-16"))
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("journal has %d entries, want 1 -- a look-alike 409 must be journaled like any other unreachable-store write", len(entries))
	}
}

// --- state set: succeeds against a real running fake store ---

func TestStateSetSucceedsAgainstReachableStore(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotBody, err = readAll(r)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state":"IN_PROGRESS"}`))
	}))
	defer srv.Close()

	body := `{"state":"IN_PROGRESS","updatedAt":"2026-08-13T10:00:00Z"}`
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "set", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(body),
		&stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty on a clean success", stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body sent to store: %v", err)
	}
	if got["mainCheckoutPath"] == nil || got["mainCheckoutPath"] == "" {
		t.Errorf("request body did not carry mainCheckoutPath: %s", gotBody)
	}
	if got["state"] != "IN_PROGRESS" {
		t.Errorf("request body state = %v, want IN_PROGRESS", got["state"])
	}

	// The successful path must not also write the fallback -- the store
	// accepted the write, so there is nothing to fall back to and no
	// journal entry to retire.
	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	if _, err := fallback.ReadStateFile(fallback.StateFilePath(projectKey, "kan-16")); err == nil {
		t.Error("state file was written on a successful store write -- the fallback path must be untouched")
	}
}

// --- state get: falls back and says so ---

func TestStateGetFallsBackAndSaysSo(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	diskBody := []byte(`{"state":"IN_PROGRESS","updatedAt":"2026-08-13T09:00:00Z"}`)
	if err := fallback.WriteStateFile(fallback.StateFilePath(projectKey, "kan-16"), diskBody); err != nil {
		t.Fatalf("seed state file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "get", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo, "kan-16"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a get must never block either); stderr:\n%s", code, stderr.String())
	}
	if !jsonEqual(t, stdout.Bytes(), diskBody) {
		t.Errorf("stdout = %s, want the on-disk record %s", stdout.Bytes(), diskBody)
	}
	if got := countLines(stderr.String()); got != 1 {
		t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
	}
	if !strings.Contains(strings.ToLower(stderr.String()), "fallback") && !strings.Contains(strings.ToLower(stderr.String()), "unreachable") {
		t.Errorf("stderr does not say the value came from the fallback: %q", stderr.String())
	}
}

// TestStateGetTreatsForeignServer404AsFallback is the CLI-level
// reproduction of F1 on the read path: pointing `-addr` at a bare server
// that answers every GET with 404 -- no myflowd behind it -- previously
// made the CLI report "no such change" and exit 1, skipping the on-disk
// fallback read entirely even with a valid local record present. It must
// instead take the fallback path and return the on-disk record.
func TestStateGetTreatsForeignServer404AsFallback(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	diskBody := []byte(`{"state":"IN_PROGRESS","updatedAt":"2026-08-13T09:00:00Z"}`)
	if err := fallback.WriteStateFile(fallback.StateFilePath(projectKey, "kan-16"), diskBody); err != nil {
		t.Fatalf("seed state file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A bare foreign server: no Myflow-Daemon header, ever.
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`Not Found`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "get", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a look-alike 404 must fall back, not block); stderr:\n%s", code, stderr.String())
	}
	if !jsonEqual(t, stdout.Bytes(), diskBody) {
		t.Errorf("stdout = %s, want the on-disk record %s (the fallback read, skipped by the pre-F1 bug)", stdout.Bytes(), diskBody)
	}
	if got := countLines(stderr.String()); got != 1 {
		t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
	}
}

func TestStateGetSucceedsAgainstReachableStore(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state":"IN_PROGRESS"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "get", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !jsonEqual(t, stdout.Bytes(), []byte(`{"state":"IN_PROGRESS"}`)) {
		t.Errorf("stdout = %s", stdout.Bytes())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty when the store answers cleanly", stderr.String())
	}
}

// TestStateGetMarksSyntheticRecord is kan-174 task 2's "A synthetic record
// is not a state" half: a change row whose only author is a stage mark's
// own bootstrap side effect (stages.SyntheticChangeUpdatedBy) is
// surfaced as `"synthetic": true` in `state get`'s output, so a caller
// (skills/myflow-fast/SKILL.md's state gate) can test a field instead of
// comparing "updatedBy" strings itself.
func TestStateGetMarksSyntheticRecord(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	syntheticBody, err := json.Marshal(map[string]string{
		"state":     "STARTED",
		"updatedBy": stages.SyntheticChangeUpdatedBy,
	})
	if err != nil {
		t.Fatalf("marshal synthetic fixture: %v", err)
	}
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(syntheticBody)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "get", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	var out struct {
		State     string `json:"state"`
		UpdatedBy string `json:"updatedBy"`
		Synthetic bool   `json:"synthetic"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("decode stdout: %v (%s)", err, stdout.Bytes())
	}
	if !out.Synthetic {
		t.Errorf("synthetic = false, want true for updatedBy %q", out.UpdatedBy)
	}
	if out.State != "STARTED" {
		t.Errorf("state = %q, want the record's own state preserved", out.State)
	}
}

// TestStateGetDoesNotMarkGenuineRecordSynthetic is the negative case
// alongside TestStateGetMarksSyntheticRecord: a record written by an
// actual pipeline command is passed through with no "synthetic" field at
// all, not merely a false one -- proving markSyntheticIfNeeded leaves an
// ordinary record untouched rather than annotating every record it sees.
func TestStateGetDoesNotMarkGenuineRecordSynthetic(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state":"IN_PROGRESS","updatedBy":"/myflow-do"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "get", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("decode stdout: %v (%s)", err, stdout.Bytes())
	}
	if _, present := out["synthetic"]; present {
		t.Errorf(`stdout carries a "synthetic" field for a genuine record: %s`, stdout.Bytes())
	}
}

// --- state list ---
//
// `state list` is F1's fix: skills/myflow-status/SKILL.md and
// skills/myflow-contracts/pipeline.md's Change name resolution enumerate
// through this command rather than a hand-written curl call against GET
// /api/v1/stats/state-board, so they inherit the daemon-header check, the
// ErrUnavailable classification and the request timeout the CLI's other
// state commands already have (see internal/client's own ListStateBoard
// tests for the header-mutation proof at the client layer). These tests
// cover the CLI layer on top of that: the never-block guarantee, the
// distinct fallback source and completeness reported to the caller, and
// that an unreadable fallback file is named rather than dropped.

type stateListOutputForTest struct {
	Source   string `json:"source"`
	Complete bool   `json:"complete"`
	Records  []struct {
		Name       string `json:"name"`
		State      string `json:"state"`
		UpdatedAt  string `json:"updatedAt"`
		UpdatedBy  string `json:"updatedBy"`
		Unreadable bool   `json:"unreadable"`
	} `json:"records"`
}

func decodeStateListOutput(t *testing.T, raw []byte) stateListOutputForTest {
	t.Helper()
	var out stateListOutputForTest
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode state list output: %v (%s)", err, raw)
	}
	return out
}

func TestStateListSucceedsAgainstReachableStore(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"view":"state-board","rows":[
			{"projectKey":"proj","name":"kan-1","state":"STARTED","updatedAt":"2026-08-13T10:00:00Z","updatedBy":"/myflow-start","nextCommand":"/myflow-do"}
		]}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "list", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty when the store answers cleanly", stderr.String())
	}
	out := decodeStateListOutput(t, stdout.Bytes())
	if out.Source != "store" {
		t.Errorf("source = %q, want %q", out.Source, "store")
	}
	if !out.Complete {
		t.Errorf("complete = false, want true when the store answered")
	}
	if len(out.Records) != 1 || out.Records[0].Name != "kan-1" || out.Records[0].State != "STARTED" {
		t.Errorf("records = %+v", out.Records)
	}
}

func TestStateListFallbackExitsZeroAndWarnsOnce(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "list", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a dead port must never block); stderr:\n%s", code, stderr.String())
	}
	if got := countLines(stderr.String()); got != 1 {
		t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
	}
	out := decodeStateListOutput(t, stdout.Bytes())
	if out.Source != "fallback" {
		t.Errorf("source = %q, want %q", out.Source, "fallback")
	}
	if out.Complete {
		t.Errorf("complete = true, want false when the store could not be reached -- the fallback directory is never a full list")
	}
}

func TestStateListFallbackReportsLocalRecords(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	seeded := []byte(`{"state":"IN_PROGRESS","updatedAt":"2026-08-13T09:00:00Z","updatedBy":"/myflow-do"}`)
	if err := fallback.WriteStateFile(fallback.StateFilePath(projectKey, "kan-9"), seeded); err != nil {
		t.Fatalf("seed state file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "list", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	out := decodeStateListOutput(t, stdout.Bytes())
	if out.Source != "fallback" || out.Complete {
		t.Fatalf("source/complete = %q/%v, want fallback/false", out.Source, out.Complete)
	}
	if len(out.Records) != 1 {
		t.Fatalf("records = %+v, want exactly one", out.Records)
	}
	r := out.Records[0]
	if r.Name != "kan-9" || r.State != "IN_PROGRESS" || r.UpdatedBy != "/myflow-do" || r.Unreadable {
		t.Errorf("records[0] = %+v", r)
	}
}

// TestStateListFallbackReportsUnreadableFileByName is this command's own
// version of the CLI-wide "an unreadable record is named, never rebuilt by
// inference" rule (skills/myflow-status/SKILL.md, skills/myflow-contracts/
// state-file.md): a fallback file that does not even parse as JSON must
// still appear in the list, marked unreadable, rather than silently
// vanish from an enumeration that is already degraded.
func TestStateListFallbackReportsUnreadableFileByName(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	if err := fallback.WriteStateFile(fallback.StateFilePath(projectKey, "kan-corrupt"), []byte(`not json at all`)); err != nil {
		t.Fatalf("seed corrupt state file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "list", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	out := decodeStateListOutput(t, stdout.Bytes())
	if len(out.Records) != 1 {
		t.Fatalf("records = %+v, want exactly one (named, not dropped)", out.Records)
	}
	if out.Records[0].Name != "kan-corrupt" || !out.Records[0].Unreadable {
		t.Errorf("records[0] = %+v, want name=kan-corrupt unreadable=true", out.Records[0])
	}
}

// TestStateListTreatsForeignServerAsFallback is F1's CLI-level
// reproduction for `state list`, parallel to
// TestStateGetTreatsForeignServer404AsFallback: a bare look-alike server
// with no myflowd behind it, answering 200 with a plausible board body,
// must still be treated as unreachable rather than trusted as a complete
// list.
func TestStateListTreatsForeignServerAsFallback(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No Myflow-Daemon header -- not a real myflowd.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"view":"state-board","rows":[{"name":"forged","state":"FINISHED"}]}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "list", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	out := decodeStateListOutput(t, stdout.Bytes())
	if out.Source != "fallback" {
		t.Errorf("source = %q, want %q (a look-alike 200 must not be trusted)", out.Source, "fallback")
	}
	for _, r := range out.Records {
		if r.Name == "forged" {
			t.Errorf("the look-alike server's row leaked into the fallback list: %+v", r)
		}
	}
}

func TestStateListTakesNoPositionalArguments(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "list", "-C", repo, "unexpected-arg"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (usage error)", code)
	}
}

func TestResolveDefaultAddr(t *testing.T) {
	tests := []struct {
		name string
		env  string
		set  bool
		want string
	}{
		{name: "unset leaves the built-in default", set: false, want: defaultAddr},
		{name: "set replaces the built-in default", env: "http://127.0.0.1:4174", set: true, want: "http://127.0.0.1:4174"},
		{name: "empty is treated as unset", env: "", set: true, want: defaultAddr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv("MYFLOW_ADDR", tt.env)
			} else {
				t.Setenv("MYFLOW_ADDR", "")
				if err := os.Unsetenv("MYFLOW_ADDR"); err != nil {
					t.Fatalf("unset MYFLOW_ADDR: %v", err)
				}
			}
			if got := resolveDefaultAddr(); got != tt.want {
				t.Fatalf("resolveDefaultAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveDefaultAddrFlagBeatsEnvironment(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	t.Setenv("MYFLOW_ADDR", "http://127.0.0.1:4174")

	var handled string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		handled = r.Host
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"state", "get", "-addr", srv.URL, "-C", repo, "some-change"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stderr=%s", code, stderr.String())
	}
	if handled == "" {
		t.Fatalf("explicit -addr was not used: request never reached the test server")
	}
}

// --- helpers ---

func jsonEqual(t *testing.T, a, b []byte) bool {
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

func readAll(r *http.Request) ([]byte, error) {
	defer func() { _ = r.Body.Close() }()
	return io.ReadAll(r.Body)
}
