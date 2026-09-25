package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/stages"
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
			"-command", "/flow", "-stage", "flow.sdd-tdd",
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
	if got["command"] != "/flow" {
		t.Errorf("command = %v, want /flow", got["command"])
	}
	if got["stage"] != "flow.sdd-tdd" {
		t.Errorf("stage = %v, want %q", got["stage"], "flow.sdd-tdd")
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
// -harness flag and no FLOW_HARNESS still carries a non-empty harness --
// stage_runs.harness is NOT NULL, so an empty value would be rejected by
// the store rather than merely "missing".
func TestStageBeginDefaultsHarnessWhenUnset(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	t.Setenv("FLOW_HARNESS", "")

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
			"-command", "/flow", "-stage", "flow.sdd-tdd", "-session-token", "mf-session-token-default-harness", "kan-16"},
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
			"-command", "/flow", "-stage", "a stage nobody documented", "-session-token", "mf-session-token-undocumented-stage", "kan-16"},
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
			"-command", "/flow", "-stage", "flow.sdd-tdd", "kan-16"},
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
					"-command", "/flow", "-stage", "flow.sdd-tdd", "-session-token", tc.sessionToken, "kan-16"},
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
			"-command", "/flow", "-stage", "flow.sdd-tdd", "-session-token", "mf-20260814-abc123", "kan-16"},
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
			"-command", "/flow", "-stage", "flow.sdd-tdd", "-session-token", expandedValue, "kan-16"},
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

// TestStageBeginDefaultsSessionFromClaudeCodeEnv pins design.md's env bind:
// when -session is not given, a begin mark picks up CLAUDE_CODE_SESSION_ID
// so the row is born with session_id set and the harvester's transcript
// search never has to run for it.
func TestStageBeginDefaultsSessionFromClaudeCodeEnv(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	t.Setenv("CLAUDE_CODE_SESSION_ID", "6af6b60e-72da-4615-8e3f-75d37bdf8f9d")

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
			"-command", "/flow", "-stage", "flow.sdd-tdd", "-session-token", "mf-session-token-env-bind", "kan-16"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if got["sessionId"] != "6af6b60e-72da-4615-8e3f-75d37bdf8f9d" {
		t.Errorf("sessionId = %v, want the CLAUDE_CODE_SESSION_ID value", got["sessionId"])
	}
	if got["sessionToken"] != "mf-session-token-env-bind" {
		t.Errorf("sessionToken = %v, want mf-session-token-env-bind", got["sessionToken"])
	}
}

// TestStageBeginSessionFlagWinsOverClaudeCodeEnv pins the precedence design.md
// states: -session, when given, wins over CLAUDE_CODE_SESSION_ID.
func TestStageBeginSessionFlagWinsOverClaudeCodeEnv(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	t.Setenv("CLAUDE_CODE_SESSION_ID", "6af6b60e-72da-4615-8e3f-75d37bdf8f9d")

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
			"-command", "/flow", "-stage", "flow.sdd-tdd", "-session", "sess-flag",
			"-session-token", "mf-session-token-flag-wins", "kan-16"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if got["sessionId"] != "sess-flag" {
		t.Errorf("sessionId = %v, want sess-flag (the flag must win over the env var)", got["sessionId"])
	}
	if got["sessionToken"] != "mf-session-token-flag-wins" {
		t.Errorf("sessionToken = %v, want mf-session-token-flag-wins", got["sessionToken"])
	}
}

// TestStageBeginOmitsSessionWhenEnvUnset pins that when neither -session nor
// CLAUDE_CODE_SESSION_ID is set, the begin mark omits sessionId entirely
// rather than sending it as an empty string (a mutation slot finding: a
// mark born with session_id = "" would skip the harvester's transcript
// search the same way a real session id does, but attribute nothing).
func TestStageBeginOmitsSessionWhenEnvUnset(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")

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
			"-command", "/flow", "-stage", "flow.sdd-tdd", "-session-token", "mf-session-token-no-env", "kan-16"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var got map[string]any
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if _, ok := got["sessionId"]; ok {
		t.Errorf("sessionId = %v, want the field absent (not empty-string)", got["sessionId"])
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
			"-command", "/flow", "-stage", "flow.sdd-tdd",
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
			"-command", "/flow", "-stage", "flow.sdd-tdd", "-outcome", "completed", "kan-16"},
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
				"-command", "/flow", "-stage", "flow.sdd-tdd", "-session-token", "mf-session-token-fallback-begin", "kan-16"},
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
				"-command", "/flow", "-stage", "flow.sdd-tdd", "-outcome", "completed", "kan-16"},
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
			"-command", "/flow", "-stage", "flow.sdd-tdd", "-session-token", "mf-session-token-fallback-journal", "kan-16"},
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

// TestStageEndNoOpenRunNamesMissingBegin pins the KAN-700 diagnosis: a
// stage end the store answers with "no open stage run" (HTTP 404 -- the
// store was reached and answered definitively) names the missing begin
// instead of printing the store-unreachable fallback a real outage gets.
// The answer is definitive, so nothing is journalled either: replaying an
// end for a run that never opened can only fail the same way again, which
// is exactly the wasted journal round trip KAN-573 paid for.
func TestStageEndNoOpenRunNamesMissingBegin(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/stages/end" {
			t.Errorf("unexpected request path %q; an end mark alone was sent", r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("api: no open stage run for that change, command and stage"))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/flow", "-stage", "flow.review-panel", "-outcome", "completed", "kan-700-demo"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a definitive store answer never blocks); stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "no open stage run for kan-700-demo/flow.review-panel — was `stage begin` recorded?") {
		t.Errorf("stderr = %q, want the missing-begin diagnosis", stderr.String())
	}
	if strings.Contains(stderr.String(), "store unreachable") {
		t.Errorf("stderr = %q, want no store-unreachable line -- the store was reached and answered", stderr.String())
	}

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-700-demo") + ".stage")
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("stage journal has %d entries, want 0 -- a definitive no-open-run answer is never journalled", len(entries))
	}
}

// TestRunStageWrapEndNoOpenRunNamesMissingBegin pins the same KAN-700
// diagnosis on `stage wrap`'s end half: begin opened a run, the store then
// answers the end mark with no-open-stage-run, and the wrapper must print
// that diagnosis -- never the outage fallback -- while still exiting with
// the child's own code and journaling nothing.
func TestRunStageWrapEndNoOpenRunNamesMissingBegin(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/stages/end" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("api: no open stage run for that change, command and stage"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "wrap", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/flow-fast", "-stage", "flow.brainstorm", "-harness", "zcode",
			"-session-token", "ff-session-token-wrap-noopen", "kan-700-demo",
			"--", "sh", "-c", "true"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (the child's own exit code); stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "no open stage run for kan-700-demo/flow.brainstorm — was `stage begin` recorded?") {
		t.Errorf("stderr = %q, want the missing-begin diagnosis", stderr.String())
	}
	if strings.Contains(stderr.String(), "store unreachable") {
		t.Errorf("stderr = %q, want no store-unreachable line -- the store was reached and answered", stderr.String())
	}

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-700-demo") + ".stage")
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("stage journal has %d entries, want 0 -- a definitive no-open-run answer is never journalled", len(entries))
	}
}

func TestStageBeginJiraKeySendsAPlanSessionMark(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var got map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/flow-plan", "-stage", "plan.session", "-session-token", "fp-cli-plan", "-jira-key", "KAN-900"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got["jiraKey"] != "KAN-900" || got["changeName"] != "" {
		t.Errorf("request (jiraKey, changeName) = (%v, %v), want (KAN-900, \"\")", got["jiraKey"], got["changeName"])
	}
}

func TestStageBeginJiraKeyAndChangeNameTogetherIsAUsageError(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "begin", "-C", repo, "-command", "/flow-plan", "-stage", "plan.session",
			"-session-token", "fp-cli-plan", "-jira-key", "KAN-900", "kan-900-slug"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
}

// --- stage wrap: one call marks the pair around the work ---

// TestRunStageWrapRunsChildAndMarksPair pins the wrapper's happy path: the
// begin mark is sent before the child runs, the child's stdout reaches the
// wrapper's, and the end mark carries outcome "completed". The wrapper
// exits with the child's exit code -- it records the work, it never alters
// it.
func TestRunStageWrapRunsChildAndMarksPair(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var mu sync.Mutex
	var paths []string
	var arrivals []time.Time
	bodies := map[string][]byte{}
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		paths = append(paths, r.URL.Path)
		arrivals = append(arrivals, time.Now())
		b, err := readAll(r)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		bodies[r.URL.Path] = b
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "wrap", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/flow-fast", "-stage", "flow.brainstorm", "-harness", "zcode",
			"-session-token", "ff-session-token-wrap-happy", "kan-323",
			"--", "sh", "-c", "sleep 0.2; echo wrapped-output"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (the child's own exit code); stderr:\n%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty on a clean success", stderr.String())
	}
	if !strings.Contains(stdout.String(), "wrapped-output") {
		t.Errorf("stdout = %q, want it to carry the child's output", stdout.String())
	}

	mu.Lock()
	defer mu.Unlock()
	if len(paths) != 2 || paths[0] != "/api/v1/stages/begin" || paths[1] != "/api/v1/stages/end" {
		t.Fatalf("request paths = %v, want begin then end, exactly once each", paths)
	}
	// The plan pins "StartedAt before the child and EndedAt after it": the
	// child sleeps 200ms between the two marks, so the gap between the
	// marks' arrivals at the store proves the pair bracketed the work
	// rather than both firing after it.
	if gap := arrivals[1].Sub(arrivals[0]); gap < 200*time.Millisecond {
		t.Errorf("gap between begin and end arrivals = %v, want >= 200ms -- the marks must bracket the child, not bookend it", gap)
	}

	var beginReq map[string]any
	if err := json.Unmarshal(bodies["/api/v1/stages/begin"], &beginReq); err != nil {
		t.Fatalf("decode begin body: %v\nbody: %s", err, bodies["/api/v1/stages/begin"])
	}
	if beginReq["changeName"] != "kan-323" {
		t.Errorf("begin changeName = %v, want kan-323", beginReq["changeName"])
	}
	if beginReq["stage"] != "flow.brainstorm" {
		t.Errorf("begin stage = %v, want flow.brainstorm", beginReq["stage"])
	}
	if beginReq["sessionToken"] != "ff-session-token-wrap-happy" {
		t.Errorf("begin sessionToken = %v, want ff-session-token-wrap-happy", beginReq["sessionToken"])
	}
	if beginReq["startedAt"] == nil || beginReq["startedAt"] == "" {
		t.Errorf("begin startedAt was not sent: %s", bodies["/api/v1/stages/begin"])
	}

	var endReq map[string]any
	if err := json.Unmarshal(bodies["/api/v1/stages/end"], &endReq); err != nil {
		t.Fatalf("decode end body: %v\nbody: %s", err, bodies["/api/v1/stages/end"])
	}
	if endReq["outcome"] != "completed" {
		t.Errorf("end outcome = %v, want completed for a child exiting 0", endReq["outcome"])
	}
	if endReq["endedAt"] == nil || endReq["endedAt"] == "" {
		t.Errorf("end endedAt was not sent: %s", bodies["/api/v1/stages/end"])
	}
}

// TestRunStageWrapChildFailureMarksFailedOutcome pins that a child exiting
// non-zero is recorded with outcome "failed" and that the wrapper exits
// with the child's own code, not 0 and not the store's.
func TestRunStageWrapChildFailureMarksFailedOutcome(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotOutcome string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/stages/end" {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode end body: %v", err)
			}
			gotOutcome, _ = body["outcome"].(string)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "wrap", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/flow-fast", "-stage", "flow.verify",
			"-session-token", "ff-session-token-wrap-fail", "kan-323",
			"--", "sh", "-c", "exit 3"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 3 {
		t.Fatalf("exit code = %d, want 3 (the child's own exit code); stderr:\n%s", code, stderr.String())
	}
	if gotOutcome != "failed" {
		t.Errorf("end outcome = %q, want failed for a child exiting non-zero", gotOutcome)
	}
}

// TestRunStageWrapStoreUnreachableStillRunsChild pins the never-block
// guarantee for the wrapper: a dead store journals both marks, prints the
// one warning line, and the child still runs -- the wrapper's exit code is
// the child's, never the store failure's.
func TestRunStageWrapStoreUnreachableStillRunsChild(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "wrap", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
			"-command", "/flow-fast", "-stage", "flow.brainstorm",
			"-session-token", "ff-session-token-wrap-dead", "kan-323",
			"--", "sh", "-c", "echo wrapped-anyway"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (the child's own exit code, dead store must never block); stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "wrapped-anyway") {
		t.Errorf("stdout = %q, want the child to have run", stdout.String())
	}
	if got := countLines(stderr.String()); got != 2 {
		t.Errorf("stderr line count = %d, want exactly 2 (one warning per half):\n%s", got, stderr.String())
	}

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(projectKey, "kan-323") + ".stage")
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(stage journal entries) = %d, want 2 (begin and end)", len(entries))
	}
	var beginBody, endBody map[string]any
	if err := json.Unmarshal(entries[0].Body, &beginBody); err != nil {
		t.Fatalf("decode journalled begin: %v", err)
	}
	if err := json.Unmarshal(entries[1].Body, &endBody); err != nil {
		t.Fatalf("decode journalled end: %v", err)
	}
	if beginBody["kind"] != "begin" || endBody["kind"] != "end" {
		t.Errorf(`journalled kinds = (%v, %v), want (begin, end)`, beginBody["kind"], endBody["kind"])
	}
}

// TestRunStageWrapUsageErrorsRunNoChild pins that every caller mistake --
// no "--" separator, an empty child after it, a missing -session-token --
// exits 2 with the usage text and never runs the child.
func TestRunStageWrapUsageErrorsRunNoChild(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"no separator", []string{
			"-command", "/flow-fast", "-stage", "flow.brainstorm",
			"-session-token", "ff-session-token-wrap-nosep", "kan-323",
			"sh", "-c", "echo should-not-run"}},
		{"empty child", []string{
			"-command", "/flow-fast", "-stage", "flow.brainstorm",
			"-session-token", "ff-session-token-wrap-empty", "kan-323", "--"}},
		{"missing session token", []string{
			"-command", "/flow-fast", "-stage", "flow.brainstorm", "kan-323",
			"--", "sh", "-c", "echo should-not-run"}},
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
				append([]string{"stage", "wrap", "-addr", srv.URL, "-timeout", "500ms", "-C", repo}, tc.args...),
				strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2 (a usage error); stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for a usage error -- it must be rejected before any network call")
			}
			if strings.Contains(stdout.String(), "should-not-run") {
				t.Error("the child ran on a usage error -- it must never run")
			}
			if stderr.Len() == 0 {
				t.Error("stderr is empty, want the usage text")
			}
		})
	}
}

// TestRunStageWrapUndocumentedStageRunsNoChild pins that an undocumented
// stage key is refused exactly as `stage begin` refuses it -- a caller
// defect, not a store outage -- and the child never runs behind it.
func TestRunStageWrapUndocumentedStageRunsNoChild(t *testing.T) {
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
		[]string{"stage", "wrap", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/flow-fast", "-stage", "a stage nobody documented",
			"-session-token", "ff-session-token-wrap-undoc", "kan-323",
			"--", "sh", "-c", "echo should-not-run"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (an undocumented stage is a usage error); stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an undocumented stage key -- it must be rejected before any network call")
	}
	if strings.Contains(stdout.String(), "should-not-run") {
		t.Error("the child ran behind an undocumented stage key -- it must never run")
	}
}

// --- stage wrap: the signal contract (fix-round 1's F8/F9/F10) ---

// TestRunStageWrapContextCancelTerminatesChildAndExits143 pins F1's fix in
// its unit-testable half: a cancelled ctx tears the child down (the child
// here sleeps far past the test) and the wrapper exits 143, the shell's
// 128+SIGTERM convention for its own termination -- with the end mark still
// recorded, on the fresh context, as failed.
func TestRunStageWrapContextCancelTerminatesChildAndExits143(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotOutcome string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/stages/end" {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotOutcome, _ = body["outcome"].(string)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()
	var stdout, stderr bytes.Buffer
	code := run(ctx,
		[]string{"stage", "wrap", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/flow-fast", "-stage", "flow.brainstorm",
			"-session-token", "ff-session-token-wrap-cancel", "kan-323",
			"--", "sh", "-c", "sleep 30"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 143 {
		t.Fatalf("exit code = %d, want 143 (128+SIGTERM, the wrapper's own termination convention); stderr:\n%s", code, stderr.String())
	}
	if gotOutcome != "failed" {
		t.Errorf("end outcome = %q, want failed -- the end mark must be recorded even when the wrapper itself was cancelled", gotOutcome)
	}
}

// TestRunStageWrapContextCancelForceKillsChildIgnoringTerm pins F9's fix:
// a child that ignores SIGTERM does not hang the wrapper -- WaitDelay
// force-kills it after the same bound every store call carries, and the
// invocation ends with the end mark recorded.
func TestRunStageWrapContextCancelForceKillsChildIgnoringTerm(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	endSeen := make(chan struct{})
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/stages/end" {
			close(endSeen)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()
	var stdout, stderr bytes.Buffer
	code := run(ctx,
		[]string{"stage", "wrap", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/flow-fast", "-stage", "flow.verify",
			"-session-token", "ff-session-token-wrap-stubborn", "kan-323",
			"--", "sh", "-c", "trap \"\" TERM; sleep 30"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 143 {
		t.Fatalf("exit code = %d, want 143; stderr:\n%s", code, stderr.String())
	}
	select {
	case <-endSeen:
	case <-time.After(2 * time.Second):
		t.Error("the end mark was never recorded -- the wrapper must still close the stage run it opened")
	}
}

// TestRunStageWrapChildKilledBySignalExitsConventional128PlusSig pins F2's
// 128+signal mapping for a child that dies to its own signal, both
// spellings of the convention.
func TestRunStageWrapChildKilledBySignalExitsConventional128PlusSig(t *testing.T) {
	cases := []struct {
		kill string
		want int
	}{
		{"kill -TERM $$", 143},
		{"kill -INT $$", 130},
	}
	for _, tc := range cases {
		t.Run(tc.kill, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"stageRunId":1,"attempt":1}`))
			}))
			defer srv.Close()

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"stage", "wrap", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
					"-command", "/flow-fast", "-stage", "flow.brainstorm",
					"-session-token", "ff-session-token-wrap-sigchild", "kan-323",
					"--", "sh", "-c", tc.kill},
				strings.NewReader(""), &stdout, &stderr)

			if code != tc.want {
				t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, tc.want, stderr.String())
			}
		})
	}
}

// --- stage keys: serves the documented vocabulary ---

// TestStageKeysSubcommandPrintsServedKeys pins `flow stage keys` to the
// served source: stdout is exactly stages.Keys(), one key per line, with a
// clean exit and no store contact. The check-stage-mark-calls guard
// consumes this output as the stage-key vocabulary's one served source, so
// an empty, padded or reordered printing would move the guard off the
// documented table this package validates marks against.
func TestStageKeysSubcommandPrintsServedKeys(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "keys"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty on a clean success", stderr.String())
	}
	want := ""
	for _, key := range stages.Keys() {
		want += key + "\n"
	}
	if want == "" {
		t.Fatal("stages.Keys() is empty -- the served vocabulary vanished")
	}
	if stdout.String() != want {
		t.Errorf("stdout = %q, want the served keys one per line %q", stdout.String(), want)
	}
}

// TestStageKeysTakesNoPositionalArguments pins the exit-2 contract the
// sibling `flow state list` holds: an unexpected positional argument is a
// usage error reported on stderr, never silently ignored input.
func TestStageKeysTakesNoPositionalArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"stage", "keys", "unexpected-arg"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (usage error)", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty on a usage error", stdout.String())
	}
	if want := "flow: stage keys takes no positional arguments"; !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr = %q, want it to carry %q", stderr.String(), want)
	}
}

// TestRunStageBeginWarnsOnSupersededRuns pins KAN-618's CLI surface: a
// begin whose daemon answer names superseded runs still exits 0, but
// warns on stderr naming what was closed -- the marking slip is caught
// at write time, in the session that made it.
func TestRunStageBeginWarnsOnSupersededRuns(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stageRunId":42,"attempt":2,"superseded":[{"id":41,"command":"/flow","stage":"flow.review-panel","attempt":1}]}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{
			"stage", "begin",
			"-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-command", "/flow", "-stage", "flow.verify",
			"-harness", "claude-code", "-session-token", "ff-session-token-warn-abc",
			"kan-618",
		},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	w := stderr.String()
	if !strings.Contains(w, "superseded") || !strings.Contains(w, "flow.review-panel") {
		t.Errorf("stderr = %q, want a supersede warning naming flow.review-panel", w)
	}
}
