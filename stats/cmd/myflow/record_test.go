package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
)

// recordJournalEntries reads the record journal for repo's project key and
// the named change, and reports whether the file exists at all. A caller
// mistake must leave no file behind, and ReadJournalEntries returning an
// empty slice for an absent file would make "refused before writing" and
// "wrote an empty journal" indistinguishable.
func recordJournalEntries(t *testing.T, repo, change string) (entries []fallback.Entry, exists bool) {
	t.Helper()
	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	path := fallback.JournalFilePath(projectKey, change) + ".record"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		t.Fatalf("stat record journal: %v", err)
	}
	got, err := fallback.ReadJournalEntries(path)
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	return got, true
}

// --- a successful write prints one line and exits 0 ---

// TestRecordWritePrintsOneLineAndExitsZero pins what each write
// subcommand sends and what it says afterwards. The finding subtests are
// the load-bearing pair: the daemon answers 201 when the upsert inserted
// and 200 when it replaced, and the CLI must say "recorded:" for the first
// and "updated:" for the second -- a CLI printing one word either way
// would make the client's own `created` flag dead weight.
func TestRecordWritePrintsOneLineAndExitsZero(t *testing.T) {
	t.Run("dispatch", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var gotPath string
		var gotBody []byte
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			var err error
			gotBody, err = readAll(r)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":7,"seq":3,"role":"implementer","model":"opus","startedAt":"2026-01-02T03:04:05Z"}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-task", "6", "-role", "implementer",
				"-model", "unknown (agent-defined)", "-key", "task-6-implementer",
				"-session-token", "mf-record-dispatch-ok",
				"-started-at", "2026-01-02T03:04:05Z"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if stderr.Len() != 0 {
			t.Errorf("stderr = %q, want empty on a clean success", stderr.String())
		}
		if got := countLines(stdout.String()); got != 1 {
			t.Errorf("stdout line count = %d, want exactly 1:\n%s", got, stdout.String())
		}
		if !strings.Contains(stdout.String(), "3") {
			t.Errorf("stdout = %q, want it to name the allocated seq 3", stdout.String())
		}
		if want := "/api/v1/records/"; !strings.HasPrefix(gotPath, want) {
			t.Errorf("request path = %s, want it under %s", gotPath, want)
		}
		if !strings.HasSuffix(gotPath, "/kan-258/dispatches") {
			t.Errorf("request path = %s, want it to end in /kan-258/dispatches", gotPath)
		}

		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		if sent["role"] != "implementer" {
			t.Errorf("role = %v, want implementer", sent["role"])
		}
		if sent["model"] != "unknown (agent-defined)" {
			t.Errorf("model = %v, want the literal %q -- never a plausible-looking slug", sent["model"], "unknown (agent-defined)")
		}
		if sent["taskId"] != "6" {
			t.Errorf("taskId = %v, want 6", sent["taskId"])
		}
		if _, ok := sent["commitSha"]; ok {
			t.Errorf("body carries commitSha = %v; begin is sent as the dispatch STARTS, when no commit exists yet", sent["commitSha"])
		}
		if _, ok := sent["outcome"]; ok {
			t.Errorf("body carries outcome = %v; begin is sent as the dispatch STARTS, when no outcome exists yet", sent["outcome"])
		}
		if sent["key"] != "task-6-implementer" {
			t.Errorf("key = %v, want task-6-implementer -- the label `end` closes and a replay collides on", sent["key"])
		}
		if sent["sessionToken"] != "mf-record-dispatch-ok" {
			t.Errorf("sessionToken = %v, want mf-record-dispatch-ok", sent["sessionToken"])
		}
		if sent["startedAt"] != "2026-01-02T03:04:05Z" {
			t.Errorf("startedAt = %v, want 2026-01-02T03:04:05Z", sent["startedAt"])
		}
	})

	t.Run("finding created", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var gotBody []byte
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			var err error
			gotBody, err = readAll(r)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"ref":"F1","round":0,"slot":"principles","severity":"major","note":"n","status":"open"}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-ref", "F1", "-round", "0", "-slot", "principles",
				"-severity", "major", "-location", "stats/cmd/myflow/record.go:1",
				"-status", "open", "-note", "the note"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if got := strings.TrimSpace(stdout.String()); got != "recorded: F1" {
			t.Errorf("stdout = %q, want %q", got, "recorded: F1")
		}

		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		if sent["ref"] != "F1" || sent["slot"] != "principles" || sent["severity"] != "major" {
			t.Errorf("request body did not carry the finding's identity: %s", gotBody)
		}
		if sent["note"] != "the note" || sent["status"] != "open" {
			t.Errorf("request body did not carry the note and status: %s", gotBody)
		}
	})

	t.Run("finding updated", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ref":"F1","round":1,"slot":"principles","severity":"major","note":"n","status":"open"}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-ref", "F1", "-round", "1", "-slot", "principles",
				"-severity", "major", "-status", "open", "-note", "restated"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if got := strings.TrimSpace(stdout.String()); got != "updated: F1" {
			t.Errorf("stdout = %q, want %q -- a 200 means the upsert replaced, not inserted", got, "updated: F1")
		}
	})

	t.Run("status", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var gotMethod, gotPath string
		var gotBody []byte
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			gotMethod, gotPath = r.Method, r.URL.Path
			var err error
			gotBody, err = readAll(r)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "status", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-ref", "F1", "-status", "fixed"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if got := countLines(stdout.String()); got != 1 {
			t.Errorf("stdout line count = %d, want exactly 1:\n%s", got, stdout.String())
		}
		if gotMethod != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", gotMethod)
		}
		if !strings.HasSuffix(gotPath, "/kan-258/findings/F1") {
			t.Errorf("request path = %s, want it to end in /kan-258/findings/F1", gotPath)
		}
		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		if sent["status"] != "fixed" {
			t.Errorf("status = %v, want fixed", sent["status"])
		}
	})
}

// --- a record write never blocks ---

// TestRecordWriteFallsBackToJournalAndExitsZero is the never-block
// guarantee, and the single case in this file that must not be allowed to
// regress into a non-zero exit: a store that cannot be reached journals
// the intent, prints exactly one warning line and exits 0, so the work the
// record describes proceeds unaffected.
func TestRecordWriteFallsBackToJournalAndExitsZero(t *testing.T) {
	cases := []struct {
		name string
		args []string
		kind string
	}{
		{
			name: "dispatch",
			args: []string{"record", "dispatch", "begin", "-change", "kan-258", "-task", "6",
				"-role", "implementer", "-model", "opus", "-key", "task-6-implementer",
				"-session-token", "mf-record-dispatch-journal", "-started-at", "2026-01-02T03:04:05Z"},
			kind: "dispatch",
		},
		{
			name: "dispatch end",
			args: []string{"record", "dispatch", "end", "-change", "kan-258", "-key", "task-6-implementer",
				"-session-token", "mf-record-dispatch-journal", "-commit", "abc1234",
				"-outcome", "completed", "-ended-at", "2026-01-02T03:44:05Z"},
			kind: "dispatch-end",
		},
		{
			name: "finding",
			args: []string{"record", "finding", "-change", "kan-258", "-ref", "F1", "-round", "0",
				"-slot", "principles", "-severity", "major", "-status", "open", "-note", "the note"},
			kind: "finding",
		},
		{
			name: "status",
			args: []string{"record", "status", "-change", "kan-258", "-ref", "F1", "-status", "fixed"},
			kind: "status",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			// The connection flags are spliced in after the verb words
			// -- "record dispatch begin" is three, "record finding" is
			// two -- so the split is the leading run of non-flag tokens
			// rather than a fixed count.
			verbLen := 0
			for verbLen < len(tc.args) && !strings.HasPrefix(tc.args[verbLen], "-") {
				verbLen++
			}
			args := append([]string{}, tc.args[:verbLen]...)
			args = append(args, "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo)
			args = append(args, tc.args[verbLen:]...)

			var stdout, stderr bytes.Buffer
			code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)

			if code != 0 {
				t.Fatalf("exit code = %d, want 0 (a dead store must never block); stderr:\n%s", code, stderr.String())
			}
			if got := countLines(stderr.String()); got != 1 {
				t.Errorf("stderr line count = %d, want exactly 1:\n%s", got, stderr.String())
			}
			if !strings.Contains(stderr.String(), "store unreachable") {
				t.Errorf("stderr = %q, want it to name the store as unreachable", stderr.String())
			}

			entries, exists := recordJournalEntries(t, repo, "kan-258")
			if !exists {
				t.Fatalf("no record journal was written")
			}
			if len(entries) != 1 {
				t.Fatalf("len(record journal entries) = %d, want 1", len(entries))
			}
			var body struct {
				Kind    string          `json:"kind"`
				Request json.RawMessage `json:"request"`
			}
			if err := json.Unmarshal(entries[0].Body, &body); err != nil {
				t.Fatalf("decode journalled body: %v", err)
			}
			if body.Kind != tc.kind {
				t.Errorf("journalled kind = %q, want %q", body.Kind, tc.kind)
			}
			if len(body.Request) == 0 {
				t.Errorf("journalled body carried no request: %s", entries[0].Body)
			}

			// The record journal is its own file, never the state
			// journal or the stage journal: internal/reconcile decodes
			// each file's entries as a different shape, so a record body
			// landing in either would break replay for every entry after
			// it.
			projectKey, _, err := fallback.ProjectKey(repo)
			if err != nil {
				t.Fatalf("ProjectKey: %v", err)
			}
			for _, other := range []struct{ what, path string }{
				{"state journal", fallback.JournalFilePath(projectKey, "kan-258")},
				{"stage journal", fallback.JournalFilePath(projectKey, "kan-258") + ".stage"},
			} {
				otherEntries, err := fallback.ReadJournalEntries(other.path)
				if err != nil {
					t.Fatalf("ReadJournalEntries (%s): %v", other.what, err)
				}
				if len(otherEntries) != 0 {
					t.Errorf("%s has %d entries, want 0 -- a record write must land in its own journal file", other.what, len(otherEntries))
				}
			}
		})
	}
}

// --- caller mistakes exit 2 and journal nothing ---

// TestRecordMissingRequiredFlagExitsTwoWithoutJournalling pins the other
// half of the never-block split: a caller mistake is not a store failure,
// so it exits 2 and leaves no journal entry to replay. Journalling it
// would queue a write that can never succeed.
func TestRecordMissingRequiredFlagExitsTwoWithoutJournalling(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"dispatch begin without -model", []string{"record", "dispatch", "begin", "-change", "kan-258",
			"-role", "implementer", "-key", "k1", "-session-token", "mf-record-missing-model",
			"-started-at", "2026-01-02T03:04:05Z"}},
		{"dispatch begin without -change", []string{"record", "dispatch", "begin",
			"-role", "implementer", "-model", "opus", "-key", "k1", "-session-token", "mf-record-missing-change",
			"-started-at", "2026-01-02T03:04:05Z"}},
		{"dispatch begin without -started-at", []string{"record", "dispatch", "begin", "-change", "kan-258",
			"-role", "implementer", "-model", "opus", "-key", "k1", "-session-token", "mf-record-missing-started"}},
		{"dispatch begin without -key", []string{"record", "dispatch", "begin", "-change", "kan-258",
			"-role", "implementer", "-model", "opus", "-session-token", "mf-record-missing-key",
			"-started-at", "2026-01-02T03:04:05Z"}},
		{"dispatch end without -key", []string{"record", "dispatch", "end", "-change", "kan-258",
			"-session-token", "mf-record-missing-end-key", "-ended-at", "2026-01-02T03:44:05Z"}},
		{"dispatch end without -ended-at", []string{"record", "dispatch", "end", "-change", "kan-258",
			"-key", "k1", "-session-token", "mf-record-missing-ended"}},
		{"finding without -note", []string{"record", "finding", "-change", "kan-258",
			"-ref", "F1", "-slot", "principles", "-severity", "major", "-status", "open"}},
		{"status without -ref", []string{"record", "status", "-change", "kan-258", "-status", "fixed"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			args := append([]string{}, tc.args[:2]...)
			args = append(args, "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo)
			args = append(args, tc.args[2:]...)

			var stdout, stderr bytes.Buffer
			code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2 (a caller mistake); stderr:\n%s", code, stderr.String())
			}
			if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
				t.Errorf("a caller mistake wrote a record journal -- a replay of it could never succeed")
			}
		})
	}
}

// TestRecordRejectsUnknownRoleWithoutContactingStore pins that the role
// allowlist is checked before the network is ever touched, exactly as
// `stage begin` checks its stage key: were it checked after, the fallback
// would swallow the caller's own mistake as if it were a store outage.
func TestRecordRejectsUnknownRoleWithoutContactingStore(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "architect", "-model", "opus", "-key", "k1",
			"-session-token", "mf-record-unknown-role", "-started-at", "2026-01-02T03:04:05Z"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an unrecognised role -- it must be refused first")
	}
	for _, role := range []string{"implementer", "reviewer", "panel-fix", "red-partner"} {
		if !strings.Contains(stderr.String(), role) {
			t.Errorf("stderr does not name the accepted role %q:\n%s", role, stderr.String())
		}
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("an unrecognised role wrote a record journal")
	}
}

// TestRecordRejectsSessionTokenSubstitution pins that the session-token
// rule reaches this verb too: a token carrying a shell substitution lands
// in every transcript as the identical unexpanded string and discriminates
// between no two sessions, so the harvest binding it exists for would bind
// nothing.
func TestRecordRejectsSessionTokenSubstitution(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-key", "k1",
			"-session-token", "mf-$(date +%s)", "-started-at", "2026-01-02T03:04:05Z"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for a session token carrying a substitution")
	}
	if !strings.Contains(stderr.String(), "command substitution") {
		t.Errorf("stderr does not say why the token was refused:\n%s", stderr.String())
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("a refused session token wrote a record journal")
	}
}

// --- usage ---

// TestRecordWithNoSubcommandPrintsUsage pins that `myflow record` alone is
// a usage error naming all four subcommands, so an operator who typed the
// verb and stopped is told what it takes rather than nothing.
func TestRecordWithNoSubcommandPrintsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"record"}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	for _, sub := range []string{"dispatch", "finding", "status", "render"} {
		if !strings.Contains(stderr.String(), sub) {
			t.Errorf("usage does not name the %q subcommand:\n%s", sub, stderr.String())
		}
	}
}

// TestRecordDispatchSendsAgentIDOnlyWhenGiven pins the optional -agent-id
// flag in both of its states, which are two ordinary states rather than a
// good one and a degraded one: Claude Code exposes a subagent identifier
// and the flag carries it, while Cursor and Codex expose none at all and
// the dispatch is recorded without one.
//
// The absent case asserts the key is missing from the body, not that it is
// empty. "" means "not reported" and must never match another absent id
// during attribution, so a wire form that spelled absence as a present,
// empty value would be a value the daemon could store and the attributor
// could compare.
func TestRecordDispatchSendsAgentIDOnlyWhenGiven(t *testing.T) {
	dispatchArgs := func(repo, addr string, extra ...string) []string {
		args := []string{"record", "dispatch", "begin", "-addr", addr, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "reviewer", "-model", "sonnet", "-key", "panel-primary",
			"-session-token", "mf-record-agent-id", "-started-at", "2026-01-02T03:04:05Z"}
		return append(args, extra...)
	}

	send := func(t *testing.T, extra ...string) map[string]any {
		t.Helper()
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var gotBody []byte
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			var err error
			gotBody, err = readAll(r)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"reviewer","model":"sonnet","startedAt":"2026-01-02T03:04:05Z"}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(), dispatchArgs(repo, srv.URL, extra...),
			strings.NewReader(""), &stdout, &stderr)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}

		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		return sent
	}

	t.Run("given", func(t *testing.T) {
		sent := send(t, "-agent-id", "agent-example0001")
		if sent["agentId"] != "agent-example0001" {
			t.Errorf("agentId = %v, want agent-example0001", sent["agentId"])
		}
	})

	t.Run("omitted", func(t *testing.T) {
		sent := send(t)
		if v, ok := sent["agentId"]; ok {
			t.Errorf("agentId = %v, want the key absent -- an unreported id is absence, not an empty value", v)
		}
	})
}

// dispatchBeginBody runs one `record dispatch begin` against a daemon that
// captures the request body, and returns that body decoded. The two
// -diff-base tests below differ only in the flags they pass and in what
// they assert about the result, so the wiring they share sits here rather
// than being written out twice.
func dispatchBeginBody(t *testing.T, extra ...string) map[string]any {
	t.Helper()
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotBody, err = readAll(r)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"reviewer","model":"sonnet","startedAt":"2026-01-02T03:04:05Z"}`))
	}))
	defer srv.Close()

	args := []string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-327", "-role", "reviewer", "-model", "sonnet", "-key", "panel-principles",
		"-session-token", "mf-record-diff-base", "-started-at", "2026-01-02T03:04:05Z"}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(), append(args, extra...), strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var sent map[string]any
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
	}
	return sent
}

// TestRecordDispatchBeginAcceptsDiffBase pins -diff-base through to the
// request body. A panel slot re-run against its own delta is dispatched
// with the sha that delta starts from, and recording it is the whole
// reason the ledger can afterwards say what that slot actually read: a
// flag the command accepted and then dropped would leave the store holding
// nothing while every caller believed the base had been recorded.
func TestRecordDispatchBeginAcceptsDiffBase(t *testing.T) {
	const base = "0f1e2d3c4b5a69788796a5b4c3d2e1f009182736"

	sent := dispatchBeginBody(t, "-diff-base", base)

	if sent["diffBase"] != base {
		t.Errorf("diffBase = %v, want %s", sent["diffBase"], base)
	}
}

// TestRecordDispatchBeginDiffBaseIsOptional pins the flag's absence, which
// is an ordinary state rather than a degraded one: every implementer
// dispatch, and every panel slot reading the whole diff, legitimately
// records no base, so a required flag here would break every existing call
// site at once.
//
// The assertion is that the key is missing from the body, not that it is
// empty -- the same distinction -agent-id draws for the same reason. An
// absent base means "not recorded", and a present empty value is one the
// store would keep and a reader could mistake for a base that was.
func TestRecordDispatchBeginDiffBaseIsOptional(t *testing.T) {
	sent := dispatchBeginBody(t)

	if v, ok := sent["diffBase"]; ok {
		t.Errorf("diffBase = %v, want the key absent -- an unrecorded base is absence, not an empty value", v)
	}
}

// TestDispatchEndAcceptsAgentID pins the delta spec's "A dispatch's
// identifier may be recorded when it becomes known": on Claude Code the
// harness reports a subagent's identifier only once the dispatch has
// actually been launched, so `begin` cannot always carry it and `end` must
// be able to.
//
// The "omitted" subtest is the load-bearing half. `begin` may already have
// recorded a real identifier for this dispatch; an `end` that sent an
// empty one would clear it, destroying the very thing this requirement
// exists to capture. So the wire body must carry no agentId key at all
// when the caller gives none -- an absent key, not an empty value -- the
// same contract TestRecordDispatchSendsAgentIDOnlyWhenGiven pins for
// begin.
func TestDispatchEndAcceptsAgentID(t *testing.T) {
	endArgs := func(repo, addr string, extra ...string) []string {
		args := []string{"record", "dispatch", "end", "-addr", addr, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-key", "panel-primary", "-session-token", "mf-record-end-agent-id",
			"-ended-at", "2026-01-02T03:44:05Z"}
		return append(args, extra...)
	}

	send := func(t *testing.T, extra ...string) map[string]any {
		t.Helper()
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var gotBody []byte
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			var err error
			gotBody, err = readAll(r)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"reviewer","model":"sonnet","endedAt":"2026-01-02T03:44:05Z"}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(), endArgs(repo, srv.URL, extra...),
			strings.NewReader(""), &stdout, &stderr)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}

		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		return sent
	}

	t.Run("given", func(t *testing.T) {
		sent := send(t, "-agent-id", "agent-example0001")
		if sent["agentId"] != "agent-example0001" {
			t.Errorf("agentId = %v, want agent-example0001", sent["agentId"])
		}
	})

	t.Run("omitted", func(t *testing.T) {
		sent := send(t)
		if v, ok := sent["agentId"]; ok {
			t.Errorf("agentId = %v, want the key absent -- an end that omits it must never clear an identifier begin already recorded", v)
		}
	})
}

// --- render ---

// renderRunRecordJSON is the body a genuine daemon answers
// GET /api/v1/records/{project}/{change} with: one dispatch and one
// finding, enough for both renderings to have rows.
const renderRunRecordJSON = `{"change":"demo",
  "dispatches":[{"id":1,"seq":1,"taskId":"11","role":"implementer","model":"unknown (agent-defined)","commitSha":"abc1234","outcome":"completed","startedAt":"2026-01-02T03:04:05Z"}],
  "findings":[{"ref":"F1","round":0,"slot":"Bugbot","severity":"Minor","location":"a.go:1","note":"n","status":"fixed","reproducer":"none — prose only"}]}`

// renderDaemon answers the run-record GET with body, and fails the test if
// the CLI sends any other request -- a render reads and writes nothing.
func renderDaemon(t *testing.T, body string) http.HandlerFunc {
	t.Helper()
	return genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("render sent a %s request; a render only reads", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	})
}

// TestRecordRenderWritesBothFilesAndPrintsRendered pins the render's
// success outcome: one `rendered: <dest>` line per kind, the file actually
// on disk, exit 0.
func TestRecordRenderWritesBothFilesAndPrintsRendered(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(renderDaemon(t, renderRunRecordJSON))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "all", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := countLines(stdout.String()); got != 2 {
		t.Fatalf("stdout line count = %d, want one outcome line per kind:\n%s", got, stdout.String())
	}
	if !strings.Contains(stdout.String(), "rendered: ") {
		t.Errorf("stdout = %q, want it to report rendered: <dest>", stdout.String())
	}

	ledgers, err := filepath.Glob(filepath.Join(repo, "docs", "superpowers", "ledgers", "*.md"))
	if err != nil {
		t.Fatalf("glob ledgers: %v", err)
	}
	if len(ledgers) != 1 {
		t.Fatalf("ledger files = %v, want exactly one", ledgers)
	}
	body, err := os.ReadFile(ledgers[0])
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	if !strings.Contains(string(body), "unknown (agent-defined)") {
		t.Errorf("rendered ledger does not name the dispatch's model:\n%s", body)
	}

	panels, err := filepath.Glob(filepath.Join(repo, "docs", "superpowers", "reviews", "*.md"))
	if err != nil {
		t.Fatalf("glob reviews: %v", err)
	}
	if len(panels) != 1 {
		t.Fatalf("panel files = %v, want exactly one", panels)
	}
}

// TestRecordRenderWithNoLedgerRowsPrintsMissingAndWritesNothing pins the
// value the run-record requirement names as distinct from a failure: the
// store holds no dispatches, which is reported, exits 0, and creates no
// file -- an empty ledger written to disk would be indistinguishable from
// a real one that happened to be empty.
//
// MISSING IS THE LEDGER'S RULE ALONE. A change with no dispatch rows
// genuinely has no ledger, and that is a fact worth reporting rather than
// a file worth inventing. The panel is the opposite case, pinned by
// TestRecordRenderPanelWithNoFindingsStillWritesTheRecord below: a panel
// that raised nothing has to SAY so.
func TestRecordRenderWithNoLedgerRowsPrintsMissingAndWritesNothing(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(renderDaemon(t, `{"change":"demo","dispatches":[],"findings":[]}`))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "ledger", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "MISSING: ledger") {
		t.Errorf("stdout = %q, want it to report MISSING: ledger", stdout.String())
	}
	if !strings.Contains(stdout.String(), "demo") {
		t.Errorf("stdout = %q, want it to name the change", stdout.String())
	}

	got, err := filepath.Glob(filepath.Join(repo, "docs", "superpowers", "ledgers", "*"))
	if err != nil {
		t.Fatalf("glob ledgers: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ledgers holds %v after a MISSING render; nothing must be written", got)
	}
}

// TestRecordRenderPanelWithNoFindingsStillWritesTheRecord pins the rule
// openspec/specs/myflow-review-panel-economics/spec.md states under "The
// panel record declares how many findings it carries": a panel that raised
// no finding says so with `findings-total: 0`, which is a DECLARATION and
// clears, where silence is not.
//
// So `-kind panel` ALWAYS writes. A clean panel produces no finding rows,
// and reporting MISSING for it would leave no record at all -- which
// check-unfinished-work.sh reads as outstanding, for a change that is
// genuinely clean. The command is invoked at panel close, so the
// invocation is itself the evidence a panel ran; no sentinel row exists in
// the store and none is needed.
func TestRecordRenderPanelWithNoFindingsStillWritesTheRecord(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(renderDaemon(t, `{"change":"demo","dispatches":[],"findings":[]}`))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "panel", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "MISSING:") {
		t.Errorf("stdout = %q, want no MISSING: for a panel — zero findings is a declaration, not an absence", stdout.String())
	}
	if !strings.Contains(stdout.String(), "rendered: ") {
		t.Errorf("stdout = %q, want it to report rendered: <dest>", stdout.String())
	}

	panels, err := filepath.Glob(filepath.Join(repo, "docs", "superpowers", "reviews", "*.md"))
	if err != nil {
		t.Fatalf("glob reviews: %v", err)
	}
	if len(panels) != 1 {
		t.Fatalf("panel files = %v, want exactly one written for a panel that raised nothing", panels)
	}
	body, err := os.ReadFile(panels[0])
	if err != nil {
		t.Fatalf("read panel: %v", err)
	}
	if !strings.Contains(string(body), "findings-total: 0\n") {
		t.Errorf("rendered panel does not declare findings-total: 0:\n%s", body)
	}
}

// TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard is the
// load-bearing case of this rule, and the reason the test above is not
// enough on its own. It runs the REAL scripts/check-unfinished-work.sh
// against the file the command actually rendered, beside a plan with every
// box ticked, so the only thing the verdict can turn on is the panel
// record.
//
// A guard re-implemented in Go would agree with the renderer by
// construction and prove nothing. A renderer that reported MISSING for a
// clean panel leaves the guard with no record to read, which it reports as
// outstanding -- for a change that is genuinely clean. That is the whole
// failure this rule removes, and this is the assertion that sees it.
func TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard(t *testing.T) {
	guard := unfinishedWorkGuard(t)
	repo := gitRepo(t)
	isolatedStateRoot(t)

	planDir := filepath.Join(repo, "spectre", "changes", "demo")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("mkdir plan dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte("# Tasks\n\n- [x] 1 done\n- [x] 2 done\n"), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}

	srv := httptest.NewServer(renderDaemon(t, `{"change":"demo","dispatches":[],"findings":[]}`))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "panel", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	// The guard reads the live worktree path /myflow-do writes the record
	// to; the render writes the archive copy. Copying the rendered bytes
	// across is what puts what the COMMAND produced under the real guard. A
	// render that wrote nothing leaves that path absent, and the guard
	// reports an absent record as outstanding -- exactly the verdict this
	// test must not see.
	sddDir := filepath.Join(repo, ".superpowers", "sdd")
	if err := os.MkdirAll(sddDir, 0o755); err != nil {
		t.Fatalf("mkdir sdd dir: %v", err)
	}
	panels, err := filepath.Glob(filepath.Join(repo, "docs", "superpowers", "reviews", "*.md"))
	if err != nil {
		t.Fatalf("glob reviews: %v", err)
	}
	if len(panels) == 1 {
		body, readErr := os.ReadFile(panels[0])
		if readErr != nil {
			t.Fatalf("read panel: %v", readErr)
		}
		if err := os.WriteFile(filepath.Join(sddDir, "final-review-panel.md"), body, 0o644); err != nil {
			t.Fatalf("write panel record: %v", err)
		}
	}

	out, err := exec.Command("bash", guard, repo, "demo").CombinedOutput()
	if err != nil {
		t.Fatalf("check-unfinished-work.sh exited non-zero (%v):\n%s", err, out)
	}
	verdict := strings.TrimSpace(string(out))
	if !strings.HasPrefix(verdict, "CLEAR:") {
		t.Fatalf("guard verdict = %q, want CLEAR — a panel that raised nothing declares findings-total: 0, and a declaration clears", verdict)
	}
}

// unfinishedWorkGuard resolves scripts/check-unfinished-work.sh from this
// package's own directory, so the test above runs the real guard rather
// than a re-implementation of its parsing.
func unfinishedWorkGuard(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "..", "scripts", "check-unfinished-work.sh"))
	if err != nil {
		t.Fatalf("resolve guard path: %v", err)
	}
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("the real guard must be runnable from this test, not stubbed: %v", err)
	}
	return abs
}

// TestRecordRenderRefusesAChangeNameOutsideTheAllowlist pins that the name
// is judged before the store is ever contacted: a refused name is a caller
// mistake, and reaching the network first would let an unreachable store
// turn it into the exit-0 fallback path.
func TestRecordRenderRefusesAChangeNameOutsideTheAllowlist(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo,
			"-change", "../escape", "-kind", "ledger", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero for a refused change name; stdout:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "../escape") {
		t.Errorf("stderr = %q, want it to name the refused change", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repo, "docs")); !os.IsNotExist(err) {
		t.Errorf("a refused render created %s", filepath.Join(repo, "docs"))
	}
	if _, exists := recordJournalEntries(t, repo, "../escape"); exists {
		t.Errorf("a caller mistake journalled an entry; a replay of it could never succeed")
	}
}

// TestRecordRenderRefusesADestinationOutsideTheRepo pins the second path
// protection inherited from preserve-session-records.sh: docs/superpowers/
// is an ordinary tracked path, so a symlink placed there in a pull request
// must not carry the render out of the repository.
func TestRecordRenderRefusesADestinationOutsideTheRepo(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	outside := t.TempDir()

	if err := os.MkdirAll(filepath.Join(repo, "docs", "superpowers"), 0o755); err != nil {
		t.Fatalf("mkdir docs/superpowers: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(repo, "docs", "superpowers", "ledgers")); err != nil {
		t.Fatalf("symlink ledgers: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "ledger", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero for a destination outside the repository; stdout:\n%s", stdout.String())
	}
	got, err := os.ReadDir(outside)
	if err != nil {
		t.Fatalf("read outside dir: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("the render followed the symlink and wrote %d entries outside the repository", len(got))
	}
}

// TestRecordRenderReusesTheFirstRendersDate pins the date rule carried
// over from the retired script: a fix round overwrites the change's
// existing dated file rather than leaving one dated duplicate per round.
func TestRecordRenderReusesTheFirstRendersDate(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	dir := filepath.Join(repo, "docs", "superpowers", "ledgers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir ledgers: %v", err)
	}
	first := filepath.Join(dir, "2020-01-01-demo.md")
	if err := os.WriteFile(first, []byte("first render\n"), 0o644); err != nil {
		t.Fatalf("write first render: %v", err)
	}

	srv := httptest.NewServer(renderDaemon(t, renderRunRecordJSON))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "ledger", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	got, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		t.Fatalf("glob ledgers: %v", err)
	}
	if len(got) != 1 || got[0] != first {
		t.Fatalf("ledger files = %v, want only the existing %s overwritten in place", got, first)
	}
	body, err := os.ReadFile(first)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	if strings.Contains(string(body), "first render") {
		t.Errorf("the existing dated file was not overwritten:\n%s", body)
	}
}

// --- journal-count ---

// recordJournalFilePath is where the record journal for repo's project key
// and the named change lives. The test derives it the same way
// recordJournalEntries does, rather than through the CLI, so a subcommand
// whose whole job is to stop callers hand-deriving this path is not itself
// verified by asking it where it looked.
func recordJournalFilePath(t *testing.T, repo, change string) string {
	t.Helper()
	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	return fallback.JournalFilePath(projectKey, change) + ".record"
}

// TestRecordJournalCountAbsentJournalCountsZero pins the ordinary case,
// which is the one that has to be right most often: every write reached
// the store and no journal was ever created. An absent file is 0, not an
// error, because the handoff line this count feeds prints on a clean run
// exactly as it prints on a degraded one -- a line printed only when
// something went wrong is indistinguishable from a line nobody printed.
func TestRecordJournalCountAbsentJournalCountsZero(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "journal-count", "-change", "kan-258", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := stdout.String(); got != "0\n" {
		t.Errorf("stdout = %q, want %q", got, "0\n")
	}
}

// TestRecordJournalCountIgnoresPartialTrailingLine is why the count is not
// `wc -l`.
//
// internal/reconcile's splitCompleteLines already draws this line for
// replay: fallback.AppendJournalEntry writes an entry and its newline in
// one Write, so the only way bytes can trail the last newline is a process
// that died mid-syscall, and that span is never parsed and never retired.
// A count that disagreed would report one more pending write than a replay
// will ever apply -- at a gate whose whole job is to be believable.
func TestRecordJournalCountIgnoresPartialTrailingLine(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	projectKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	path := recordJournalFilePath(t, repo, "kan-258")
	for _, body := range []string{`{"kind":"dispatch"}`, `{"kind":"finding"}`} {
		if err := fallback.AppendJournalEntry(path, projectKey, "kan-258", []byte(body), time.Now()); err != nil {
			t.Fatalf("AppendJournalEntry: %v", err)
		}
	}
	// The partial third: a line the kernel accepted only part of before
	// the writing process died, so it carries no terminating newline and
	// can never complete itself on a later run.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open journal: %v", err)
	}
	if _, err := f.WriteString(`{"projectKey":"` + projectKey + `","name":"kan-2`); err != nil {
		t.Fatalf("write partial line: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close journal: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "journal-count", "-change", "kan-258", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := stdout.String(); got != "2\n" {
		t.Errorf("stdout = %q, want %q -- a partial trailing line is not an entry", got, "2\n")
	}
}

// TestRecordJournalCountPrintsUnknownWhenItCannotCount is the never-block
// guarantee in the one shape this subcommand can break it.
//
// This command exists to make a handoff line honest. A count it cannot
// produce must therefore say so and get out of the way -- `unknown`, exit
// 0 -- rather than becoming the reason the handoff does not print at all.
// A non-zero exit here would put the gate's own output behind a filesystem
// the run has no other reason to care about.
func TestRecordJournalCountPrintsUnknownWhenItCannotCount(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	// A directory where the journal file belongs: every read of it fails,
	// on every platform and whatever the test runs as -- unlike a mode
	// stripped to 0000, which root reads regardless.
	path := recordJournalFilePath(t, repo, "kan-258")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir over journal path: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "journal-count", "-change", "kan-258", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 -- an uncountable journal must never block the handoff; stderr:\n%s", code, stderr.String())
	}
	if got := stdout.String(); got != "unknown\n" {
		t.Errorf("stdout = %q, want %q", got, "unknown\n")
	}
}

// --- cost-status ------------------------------------------------------

// TestCostStatusPrintsOneLine pins cost-status's success shape: one line
// on stdout naming how many of the change's dispatches are unattributed
// and, where the count is non-zero, why -- the reason wording the ledger
// (internal/records) already renders, one `<reason>: <count>` clause per
// reason present. Unlike journal-count this verb DOES contact the store,
// because only the store can answer the question: it exists to state a
// figure journal-count has no way to derive.
func TestCostStatusPrintsOneLine(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotPath, gotMethod string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"unattributed":2,"reasons":{"session never bound":2}}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "cost-status", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if gotMethod != http.MethodGet {
		t.Errorf("cost-status sent a %s request; it only reads", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/kan-258/cost-status") {
		t.Errorf("request path = %s, want it to end in /kan-258/cost-status", gotPath)
	}
	want := "2 unattributed — session never bound: 2\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

// TestCostStatusNeverBlocks pins the one guarantee cost-status exists
// for: it must never itself become the reason a handoff does not print
// the figure it names. An unreachable store -- a daemon not yet started,
// a stale addr -- prints the literal `unknown` and exits 0, exactly the
// contract journal-count carries for the filesystem it reads instead.
func TestCostStatusNeverBlocks(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "cost-status", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
			"-change", "kan-258"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 -- cost-status must never block a handoff; stderr:\n%s", code, stderr.String())
	}
	if got := stdout.String(); got != "unknown\n" {
		t.Errorf("stdout = %q, want %q", got, "unknown\n")
	}
}

// --- the dispatch pair ----------------------------------------------------

// TestRecordDispatchBeginPrintsTheAllocatedSeqAndEndClosesIt pins the pair
// end to end against a daemon that answers both halves: begin prints the
// seq the store allocated -- the identifier `record finding -dispatch-seq`
// names a dispatch by, and the only way a caller learns where in the
// change's record its dispatch landed -- and end names its row by the key
// and token begin carried, never by that seq.
func TestRecordDispatchBeginPrintsTheAllocatedSeqAndEndClosesIt(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var endPath string
	var endBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/dispatches/end") {
			endPath = r.URL.Path
			var err error
			if endBody, err = readAll(r); err != nil {
				t.Errorf("read end body: %v", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":7,"seq":4,"role":"implementer","model":"opus","endedAt":"2026-01-02T03:44:05Z"}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"seq":4,"role":"implementer","model":"opus","startedAt":"2026-01-02T03:04:05Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-task", "6", "-role", "implementer", "-model", "opus",
			"-key", "task-6-implementer", "-session-token", "mf-record-pair",
			"-started-at", "2026-01-02T03:04:05Z"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("begin exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "4") {
		t.Errorf("begin stdout = %q, want it to name the allocated seq 4", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = run(context.Background(),
		[]string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-key", "task-6-implementer", "-session-token", "mf-record-pair",
			"-commit", "abc1234", "-outcome", "completed", "-ended-at", "2026-01-02T03:44:05Z"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("end exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.HasSuffix(endPath, "/kan-258/dispatches/end") {
		t.Errorf("end request path = %s, want it to end in /kan-258/dispatches/end", endPath)
	}

	var sent map[string]any
	if err := json.Unmarshal(endBody, &sent); err != nil {
		t.Fatalf("decode end body: %v\nbody: %s", err, endBody)
	}
	for field, want := range map[string]string{
		"key":          "task-6-implementer",
		"sessionToken": "mf-record-pair",
		"commitSha":    "abc1234",
		"outcome":      "completed",
		"endedAt":      "2026-01-02T03:44:05Z",
	} {
		if sent[field] != want {
			t.Errorf("end body %s = %v, want %q", field, sent[field], want)
		}
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("a clean pair wrote a record journal")
	}
}

// TestRecordDispatchEndJournalsAKeyTheStoreDoesNotKnowYet pins the one
// place this verb treats a 404 as retryable rather than definitive.
//
// Everywhere else a 404 means a caller's typo: the store answered, a replay
// would be refused identically forever, and journalling it would queue a
// write that can never succeed. "No dispatch under this key" has a second,
// entirely ordinary cause -- the begin that would have created the row was
// itself journalled and is still queued ahead of this entry -- and refusing
// it would lose the end and leave the window that begin opened open
// forever, which is the defect the end call exists to prevent.
func TestRecordDispatchEndJournalsAKeyTheStoreDoesNotKnowYet(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"store: dispatch not found"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-key", "task-6-implementer", "-session-token", "mf-record-end-404",
			"-outcome", "completed", "-ended-at", "2026-01-02T03:44:05Z"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 -- a dispatch end whose begin has not landed yet must never block; stderr:\n%s", code, stderr.String())
	}
	entries, exists := recordJournalEntries(t, repo, "kan-258")
	if !exists {
		t.Fatal("no record journal was written -- the end would be lost and its window left open forever")
	}
	if len(entries) != 1 {
		t.Fatalf("len(record journal entries) = %d, want 1", len(entries))
	}
	var body struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(entries[0].Body, &body); err != nil {
		t.Fatalf("decode journalled body: %v", err)
	}
	if body.Kind != "dispatch-end" {
		t.Errorf("journalled kind = %q, want dispatch-end", body.Kind)
	}
}

// TestRecordFindingStillRefusesAnUnknownRef is the contrast case for the
// exception above: a 404 from any other record write stays definitive, so
// the exception cannot quietly become the rule.
func TestRecordFindingStillRefusesAnUnknownRef(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"store: finding not found"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "status", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-ref", "F99", "-status", "fixed"},
		strings.NewReader(""), &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero -- a ref naming nothing is refused identically forever")
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("a refused ref wrote a record journal -- a replay of it could never succeed")
	}
}
