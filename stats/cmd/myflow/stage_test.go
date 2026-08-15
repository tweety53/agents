package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/fallback"
)

// --- stage begin: records identity and instant ---

// TestStageBeginRecordsIdentityAndInstant pins that `stage begin` sends
// the store exactly the identity a real begin mark needs: project (derived
// from the git repo, not typed by the caller), change name, command,
// stage, harness and session id -- and that a documented stage key is
// accepted, never rejected as a usage error.
func TestStageBeginRecordsIdentityAndInstant(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotBody, err = readAll(r)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if r.URL.Path != "/api/v1/stages/begin" {
			t.Errorf("request path = %s, want /api/v1/stages/begin", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":42,"attempt":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{
			"stage", "begin",
			"-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/myflow-do", "-stage", "do.sdd-tdd",
			"-harness", "claude-code", "-session", "sess-123",
			"-session-token", "mf-session-token-identity-abc",
			"kan-16",
		},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty on a clean success", stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body sent to store: %v\nbody: %s", err, gotBody)
	}
	if got["changeName"] != "kan-16" {
		t.Errorf("changeName = %v, want kan-16", got["changeName"])
	}
	if got["command"] != "/myflow-do" {
		t.Errorf("command = %v, want /myflow-do", got["command"])
	}
	if got["stage"] != "do.sdd-tdd" {
		t.Errorf("stage = %v, want %q", got["stage"], "do.sdd-tdd")
	}
	if got["harness"] != "claude-code" {
		t.Errorf("harness = %v, want claude-code", got["harness"])
	}
	if got["sessionId"] != "sess-123" {
		t.Errorf("sessionId = %v, want sess-123", got["sessionId"])
	}
	if got["sessionToken"] != "mf-session-token-identity-abc" {
		t.Errorf("sessionToken = %v, want mf-session-token-identity-abc", got["sessionToken"])
	}
	if got["projectKey"] == nil || got["projectKey"] == "" {
		t.Errorf("projectKey was not sent: %s", gotBody)
	}
	if got["startedAt"] == nil || got["startedAt"] == "" {
		t.Errorf("startedAt was not sent: %s", gotBody)
	}
}

// TestStageBeginDefaultsHarnessWhenUnset pins that a begin mark with no
// -harness flag and no MYFLOW_HARNESS still carries a non-empty harness --
// stage_runs.harness is NOT NULL, so an empty value would be rejected by
// the store rather than merely "missing".
func TestStageBeginDefaultsHarnessWhenUnset(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	t.Setenv("MYFLOW_HARNESS", "")

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotBody, err = readAll(r)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/myflow-do", "-stage", "do.sdd-tdd", "-session-token", "mf-session-token-default-harness", "kan-16"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if got["harness"] == nil || got["harness"] == "" {
		t.Errorf("harness = %v, want a non-empty default", got["harness"])
	}
}

// --- stage begin/end: an undocumented stage key is rejected before the
// network is ever touched ---

// TestStageBeginRejectsUndocumentedStageWithoutContactingStore is the
// CLI-level half of task 8's rejection requirement (internal/stages/
// names_test.go pins Validate itself): a stage key absent from README's
// Level 1 table must be refused as a usage error, and the store must
// never be contacted for it -- if it were, the fallback would swallow the
// caller's own mistake as if it were a store outage.
func TestStageBeginRejectsUndocumentedStageWithoutContactingStore(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/myflow-do", "-stage", "a stage nobody documented", "-session-token", "mf-session-token-undocumented-stage", "kan-16"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (an undocumented stage is a usage error); stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an undocumented stage key -- it must be rejected before any network call")
	}
	if stderr.Len() == 0 {
		t.Error("stderr is empty, want an error naming the documented alternatives")
	}
}

// --- stage begin: a sessionToken that cannot identify anything is rejected ---

// TestStageBeginRequiresSessionToken pins tasks.md's "A missing -session-token is a
// caller mistake, not a stage outcome": exit non-zero, name the flag, and
// never contact the store -- exactly like an undocumented stage key.
func TestStageBeginRequiresSessionToken(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/myflow-do", "-stage", "do.sdd-tdd", "kan-16"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (a missing -session-token is a usage error); stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted with no -session-token -- it must be rejected before any network call")
	}
	if !strings.Contains(stderr.String(), "-session-token") {
		t.Errorf("stderr = %q, want it to name -session-token", stderr.String())
	}
}

// TestStageBeginRejectsShellSubstitutionSessionToken pins design.md's "the sessionToken
// is a literal, never a shell substitution": each of the three shapes the
// task names -- "$(", a backtick, and "$" followed by a name -- is
// rejected with its own case, the store is never contacted, and the error
// says why rather than just "invalid" (design.md: "a reader who does not
// know that will reintroduce the defect").
func TestStageBeginRejectsShellSubstitutionSessionToken(t *testing.T) {
	cases := []struct {
		name         string
		sessionToken string
		wantMessage  string
	}{
		{"command substitution", "mf-$(date +%s)-$$", "command substitution"},
		{"backtick", "mf-`date +%s`", "backtick"},
		{"shell variable", "mf-$SESSION_ID", "shell variable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"stage", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
					"-command", "/myflow-do", "-stage", "do.sdd-tdd", "-session-token", tc.sessionToken, "kan-16"},
				strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2 (a shell-substitution sessionToken is a usage error); stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted with a shell-substitution sessionToken -- it must be rejected before any network call")
			}
			if !strings.Contains(stderr.String(), tc.wantMessage) {
				t.Errorf("stderr = %q, want it to explain why (%q)", stderr.String(), tc.wantMessage)
			}
		})
	}
}

// TestStageBeginAcceptsLiteralSessionToken pins the positive case alongside the
// three rejections above: a literal sessionToken with no shell metacharacters at
// all is accepted and sent to the store unchanged -- covered in detail by
// TestStageBeginRecordsIdentityAndInstant's own "sessionToken" assertion; this
// pins specifically that a sessionToken merely containing "$" on its own (no
// "$(" and no following name character) is not mistaken for a
// substitution.
func TestStageBeginAcceptsLiteralSessionToken(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/myflow-do", "-stage", "do.sdd-tdd", "-session-token", "mf-20260814-abc123", "kan-16"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if got["sessionToken"] != "mf-20260814-abc123" {
		t.Errorf("sessionToken = %v, want mf-20260814-abc123", got["sessionToken"])
	}
}

// TestStageBeginCannotDetectShellExpandedSessionToken documents tasks.md
// task 1b's finding rather than a fix: a sessionToken written
// `-session-token $T` at the call site is expanded by the calling shell
// before this program's argv is ever populated, so what this test sends is
// exactly what the real CLI receives from that invocation -- an ordinary
// literal indistinguishable from one the caller typed by hand. The mark is
// accepted and sent to the store; the transcript for the real invocation
// would still record the unexpanded "$T" and never contain this literal,
// so the mark silently binds nothing. No check at this layer can tell
// these two cases apart (validateSessionToken's own doc comment explains
// why); the actual defence is downstream, in
// internal/harvest.Watcher.resolveSessionTokens's bounded give-up and
// warning when a token never matches any transcript. This test exists so a
// future change cannot "fix" this by asserting a rejection here without
// first reading why one was ruled out.
func TestStageBeginCannotDetectShellExpandedSessionToken(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	// The value a shell would leave behind after expanding an unquoted or
	// double-quoted $T -- ordinary characters, no "$", no backtick: this is
	// what -session-token $T actually delivers to argv, not the literal
	// text "$T" a caller reading the call site would assume was recorded.
	const expandedValue = "mf-20260815-142233-9f3c1a"

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotBody, err = readAll(r)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/myflow-do", "-stage", "do.sdd-tdd", "-session-token", expandedValue, "kan-16"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 -- this shape is, and must remain, indistinguishable "+
			"from a hand-typed literal at this layer; stderr:\n%s", code, stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if got["sessionToken"] != expandedValue {
		t.Errorf("sessionToken = %v, want %v", got["sessionToken"], expandedValue)
	}
}

// --- stage end: records outcome and metrics ---

// TestStageEndRecordsOutcomeAndMetrics pins that `stage end` sends the
// store the outcome and the metrics its own flags describe, deep enough to
// round-trip through -findings' own JSON.
func TestStageEndRecordsOutcomeAndMetrics(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotBody, err = readAll(r)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if r.URL.Path != "/api/v1/stages/end" {
			t.Errorf("request path = %s, want /api/v1/stages/end", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":42,"attempt":2}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{
			"stage", "end",
			"-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/myflow-do", "-stage", "do.sdd-tdd",
			"-outcome", "completed",
			"-fix-rounds", "2", "-panel-rounds", "1",
			"-findings", `{"critical":0,"major":1}`,
			"kan-16",
		},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty on a clean success", stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body sent to store: %v\nbody: %s", err, gotBody)
	}
	if got["outcome"] != "completed" {
		t.Errorf("outcome = %v, want completed", got["outcome"])
	}
	metrics, ok := got["metrics"].(map[string]any)
	if !ok {
		t.Fatalf("metrics = %v, want an object", got["metrics"])
	}
	if metrics["fix_rounds"] != float64(2) {
		t.Errorf("metrics.fix_rounds = %v, want 2", metrics["fix_rounds"])
	}
	if metrics["panel_rounds"] != float64(1) {
		t.Errorf("metrics.panel_rounds = %v, want 1", metrics["panel_rounds"])
	}
	findings, ok := metrics["findings_by_severity"].(map[string]any)
	if !ok {
		t.Fatalf("metrics.findings_by_severity = %v, want an object", metrics["findings_by_severity"])
	}
	if findings["major"] != float64(1) {
		t.Errorf("metrics.findings_by_severity.major = %v, want 1", findings["major"])
	}
}

// TestStageEndOmitsMetricsWhenNoFlagsGiven pins that `stage end` sends no
// metrics field at all when none of -fix-rounds, -panel-rounds or
// -findings were given -- MergeMetrics requires a non-nil patch
// (store.ErrNilMetricsPatch), so the daemon must not be asked to merge an
// empty one for the common case of a stage with nothing but an outcome.
func TestStageEndOmitsMetricsWhenNoFlagsGiven(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/myflow-do", "-stage", "do.sdd-tdd", "-outcome", "completed", "kan-16"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if _, present := got["metrics"]; present {
		t.Errorf("request body carried a metrics field with no flags given: %s", gotBody)
	}
}

// --- a mark never blocks ---

// TestStageMarkFallsBackAndExitsZero pins the never-block guarantee for
// both `stage begin` and `stage end`: a dead store must never stop the
// pipeline, and the intent must be journalled somewhere durable rather
// than silently dropped.
func TestStageMarkFallsBackAndExitsZero(t *testing.T) {
	t.Run("begin", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"stage", "begin", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-command", "/myflow-do", "-stage", "do.sdd-tdd", "-session-token", "mf-session-token-fallback-begin", "kan-16"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0 (dead port must never block); stderr:\n%s", code, stderr.String())
		}
		if got := countLines(stderr.String()); got != 1 {
			t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
		}

		projectKey, _, err := fallback.ProjectKey(repo)
		if err != nil {
			t.Fatalf("ProjectKey: %v", err)
		}
		entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-16") + ".stage")
		if err != nil {
			t.Fatalf("ReadJournalEntries: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("len(stage journal entries) = %d, want 1", len(entries))
		}
		var body map[string]any
		if err := json.Unmarshal(entries[0].Body, &body); err != nil {
			t.Fatalf("decode journalled body: %v", err)
		}
		if body["kind"] != "begin" {
			t.Errorf(`journalled kind = %v, want "begin"`, body["kind"])
		}
	})

	t.Run("end", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"stage", "end", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-command", "/myflow-do", "-stage", "do.sdd-tdd", "-outcome", "completed", "kan-16"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0 (dead port must never block); stderr:\n%s", code, stderr.String())
		}
		if got := countLines(stderr.String()); got != 1 {
			t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
		}

		projectKey, _, err := fallback.ProjectKey(repo)
		if err != nil {
			t.Fatalf("ProjectKey: %v", err)
		}
		entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-16") + ".stage")
		if err != nil {
			t.Fatalf("ReadJournalEntries: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("len(stage journal entries) = %d, want 1", len(entries))
		}
		var body map[string]any
		if err := json.Unmarshal(entries[0].Body, &body); err != nil {
			t.Fatalf("decode journalled body: %v", err)
		}
		if body["kind"] != "end" {
			t.Errorf(`journalled kind = %v, want "end"`, body["kind"])
		}
	})
}

// TestStageMarkFallbackDoesNotTouchStateJournal pins that a stage mark's
// fallback lands in its own journal file, never in the state journal
// `state set` uses: internal/reconcile's replay (task 6) decodes every
// entry in that file as a whole change PUT body with
// DisallowUnknownFields, so a stage-mark body landing there would corrupt
// replay for every entry after it.
func TestStageMarkFallbackDoesNotTouchStateJournal(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "begin", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
			"-command", "/myflow-do", "-stage", "do.sdd-tdd", "-session-token", "mf-session-token-fallback-journal", "kan-16"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	stateEntries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-16"))
	if err != nil {
		t.Fatalf("ReadJournalEntries (state journal): %v", err)
	}
	if len(stateEntries) != 0 {
		t.Errorf("state journal has %d entries, want 0 -- a stage mark must land in its own journal file", len(stateEntries))
	}
}
