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
			[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
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
				"-severity", "major", "-location", "stats/cmd/flow/record.go:1",
				"-status", "open", "-reproducer", "scripts/x.sh", "-note", "the note"},
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
				"-severity", "major", "-status", "open", "-reproducer", "scripts/x.sh", "-note", "restated"},
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
			args: []string{"record", "dispatch", "begin", "-agent-id", "none", "-change", "kan-258", "-task", "6",
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
				"-slot", "principles", "-severity", "major", "-status", "open",
				"-reproducer", "scripts/x.sh", "-note", "the note"},
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
		{"dispatch begin without -model", []string{"record", "dispatch", "begin", "-agent-id", "none", "-change", "kan-258",
			"-role", "implementer", "-key", "k1", "-session-token", "mf-record-missing-model",
			"-started-at", "2026-01-02T03:04:05Z"}},
		{"dispatch begin without -change", []string{"record", "dispatch", "begin", "-agent-id", "none",
			"-role", "implementer", "-model", "opus", "-key", "k1", "-session-token", "mf-record-missing-change",
			"-started-at", "2026-01-02T03:04:05Z"}},
		{"dispatch begin without -started-at", []string{"record", "dispatch", "begin", "-agent-id", "none", "-change", "kan-258",
			"-role", "implementer", "-model", "opus", "-key", "k1", "-session-token", "mf-record-missing-started"}},
		{"dispatch begin without -key", []string{"record", "dispatch", "begin", "-agent-id", "none", "-change", "kan-258",
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
		[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "architect", "-model", "opus", "-key", "k1",
			"-session-token", "mf-record-unknown-role", "-started-at", "2026-01-02T03:04:05Z"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an unrecognised role -- it must be refused first")
	}
	for _, role := range []string{"implementer", "reviewer", "panel-fix", "red-partner", "planner", "conductor", "verifier"} {
		if !strings.Contains(stderr.String(), role) {
			t.Errorf("stderr does not name the accepted role %q:\n%s", role, stderr.String())
		}
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("an unrecognised role wrote a record journal")
	}
}

// TestRecordAcceptsPlannerRole pins that "planner" is in the accepted-role
// allowlist alongside implementer/reviewer/panel-fix/red-partner -- the
// role `/flow` records for the subagent it dispatches to plan
// brainstorm.md sections B-D and flow.document-fix.
func TestRecordAcceptsPlannerRole(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1,"role":"planner","model":"opus","startedAt":"2026-01-02T03:04:05Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "planner", "-model", "opus", "-key", "k1",
			"-session-token", "mf-record-planner-role", "-started-at", "2026-01-02T03:04:05Z"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !contacted {
		t.Error("the store was not contacted for the planner role")
	}
}

// TestRecordAcceptsConductorRole pins that "conductor" is in the
// accepted-role allowlist -- the role `/flow` records for the subagent
// that runs implement.md, review-panel.md and verify-and-handoff.md as
// the conductor (design.md's conductor-runs-implementation-half).
func TestRecordAcceptsConductorRole(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1,"role":"conductor","model":"sonnet","startedAt":"2026-01-02T03:04:05Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "conductor", "-model", "sonnet", "-key", "conductor",
			"-session-token", "mf-record-conductor-role", "-started-at", "2026-01-02T03:04:05Z"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !contacted {
		t.Error("the store was not contacted for the conductor role")
	}
}

// TestRecordAcceptsVerifierRole pins that "verifier" is in the
// accepted-role allowlist -- the role `/flow` records for the subagent
// skills/flow/verify-and-handoff.md dispatches for flow.verify and
// flow.visual-verify.
func TestRecordAcceptsVerifierRole(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1,"role":"verifier","model":"sonnet","startedAt":"2026-01-02T03:04:05Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "verifier", "-model", "sonnet", "-key", "k1",
			"-session-token", "mf-record-verifier-role", "-started-at", "2026-01-02T03:04:05Z"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !contacted {
		t.Error("the store was not contacted for the verifier role")
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
		[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
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

// TestRecordWithNoSubcommandPrintsUsage pins that `flow record` alone is
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

// TestRecordDispatchSendsAgentIDOnlyWhenGiven pins the required -agent-id
// flag in its three states: Claude Code exposes a subagent identifier and
// the flag carries it; Cursor and Codex expose none at all and the caller
// passes the literal "none", recorded as no id; and a begin that omits the
// flag is a caller mistake, exit 2, because an unattributable dispatch is
// what this flag exists to prevent.
//
// The "none" case asserts the key is missing from the body, not that it is
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

	t.Run("none", func(t *testing.T) {
		sent := send(t, "-agent-id", "none")
		if v, ok := sent["agentId"]; ok {
			t.Errorf("agentId = %v, want the key absent -- an unreported id is absence, not an empty value", v)
		}
	})

	t.Run("omitted", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), dispatchArgs(repo, "http://127.0.0.1:1"),
			strings.NewReader(""), &stdout, &stderr)
		if code != 2 {
			t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "-agent-id is required") {
			t.Errorf("stderr = %q, want it to name -agent-id as required", stderr.String())
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

	args := []string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
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

	ledgers, err := filepath.Glob(filepath.Join(repo, ".superpowers", "sdd", "ledgers", "*.md"))
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

	panels, err := filepath.Glob(filepath.Join(repo, ".superpowers", "sdd", "reviews", "*.md"))
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

	got, err := filepath.Glob(filepath.Join(repo, ".superpowers", "sdd", "ledgers", "*"))
	if err != nil {
		t.Fatalf("glob ledgers: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ledgers holds %v after a MISSING render; nothing must be written", got)
	}
}

// TestRecordRenderPanelWithNoFindingsStillWritesTheRecord pins the rule
// the review-panel economics requirement states under "The
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

	panels, err := filepath.Glob(filepath.Join(repo, ".superpowers", "sdd", "reviews", "*.md"))
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

	// renderDaemon itself is GET-only -- correct for the tests that call
	// only `record render`. This test also shells out to the guard below,
	// which (since check-unfinished-work.sh's own advisory verdict write)
	// POSTs its verdict to this same FLOW_ADDR. That write's outcome is
	// irrelevant here -- the guard's own combined output already silences
	// it -- so this handler answers GET with the render body and accepts
	// any other method rather than failing the test over it.
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"change":"demo","dispatches":[],"findings":[]}`))
	}))
	defer srv.Close()
	// The guard shells out to `flow record findings` with no -addr, and that
	// subprocess inherits this process's environment. Pinning FLOW_ADDR to
	// srv.URL keeps the guard reading from this test's daemon rather than
	// from the caller's FLOW_ADDR or the dev daemon on 4173.
	t.Setenv("FLOW_ADDR", srv.URL)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "panel", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	// The guard reads the live worktree path /flow writes the record
	// to; the render writes the archive copy. Copying the rendered bytes
	// across is what puts what the COMMAND produced under the real guard. A
	// render that wrote nothing leaves that path absent, and the guard
	// reports an absent record as outstanding -- exactly the verdict this
	// test must not see.
	sddDir := filepath.Join(repo, ".superpowers", "sdd")
	if err := os.MkdirAll(sddDir, 0o755); err != nil {
		t.Fatalf("mkdir sdd dir: %v", err)
	}
	panels, err := filepath.Glob(filepath.Join(repo, ".superpowers", "sdd", "reviews", "*.md"))
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
// protection inherited from preserve-session-records.sh: a symlink placed
// under .superpowers/sdd/ must not carry the render out of the repository.
func TestRecordRenderRefusesADestinationOutsideTheRepo(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	outside := t.TempDir()

	if err := os.MkdirAll(filepath.Join(repo, ".superpowers", "sdd"), 0o755); err != nil {
		t.Fatalf("mkdir .superpowers/sdd: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(repo, ".superpowers", "sdd", "ledgers")); err != nil {
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

	dir := filepath.Join(repo, ".superpowers", "sdd", "ledgers")
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
		[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
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

// --- write-time validation: status and reproducer ---

// TestRecordFindingRejectsUnknownStatus pins that an unrecognised status is
// refused before the store is contacted, exactly as an unrecognised role is
// in TestRecordRejectsUnknownRoleWithoutContactingStore: were it checked
// after, the fallback would swallow the caller's own mistake as if it were
// a store outage.
func TestRecordFindingRejectsUnknownStatus(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ref":"F1","round":0,"slot":"principles","severity":"major","note":"n","status":"open"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-ref", "F1", "-round", "0", "-slot", "principles",
			"-severity", "major", "-status", "in-review", "-reproducer", "scripts/x.sh",
			"-note", "the note"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an unrecognised status -- it must be refused first")
	}
	for _, want := range []string{"open", "fixed", "withdrawn"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr does not name the accepted vocabulary %q:\n%s", want, stderr.String())
		}
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("an unrecognised status wrote a record journal")
	}
}

// TestRecordFindingRejectsWithdrawnWithNoReason pins that a bare "withdrawn"
// -- with no reason following it, whitespace-only counting as none -- is
// refused before the store is contacted.
func TestRecordFindingRejectsWithdrawnWithNoReason(t *testing.T) {
	for _, status := range []string{"withdrawn", "withdrawn   "} {
		t.Run(status, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
					"-change", "kan-258", "-ref", "F1", "-round", "0", "-slot", "principles",
					"-severity", "major", "-status", status, "-reproducer", "scripts/x.sh",
					"-note", "the note"},
				strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for a withdrawn status with no reason -- it must be refused first")
			}
			if !strings.Contains(stderr.String(), "withdrawn") {
				t.Errorf("stderr does not name the accepted vocabulary:\n%s", stderr.String())
			}
			if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
				t.Error("a reasonless withdrawn status wrote a record journal")
			}
		})
	}
}

// --- KAN-451: guard verdicts and incidents ---

// TestRecordVerdictWritesAndFallsBackToJournal pins `flow record verdict`'s
// two outcomes: a reachable store gets a POST body carrying every flag and
// a non-zero recordedAt, the identical never-block fallback every other
// record write shares.
func TestRecordVerdictWritesAndFallsBackToJournal(t *testing.T) {
	t.Run("writes the verdict", func(t *testing.T) {
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
			_, _ = w.Write([]byte(`{"id":1,"guard":"check-unfinished-work","worktree":"/wt/kan-451","verdict":"CLEAR: /wt/kan-451 -- every plan item is checked and no finding is open","recordedAt":"2026-01-02T03:04:05Z"}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "verdict", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-451", "-guard", "check-unfinished-work", "-worktree", "/wt/kan-451",
				"-verdict", "CLEAR: /wt/kan-451 -- every plan item is checked and no finding is open"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if !strings.HasSuffix(gotPath, "/kan-451/verdicts") {
			t.Errorf("request path = %s, want it to end in /kan-451/verdicts", gotPath)
		}

		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		if sent["guard"] != "check-unfinished-work" {
			t.Errorf("guard = %v, want check-unfinished-work", sent["guard"])
		}
		if sent["worktree"] != "/wt/kan-451" {
			t.Errorf("worktree = %v, want /wt/kan-451", sent["worktree"])
		}
		if sent["verdict"] != "CLEAR: /wt/kan-451 -- every plan item is checked and no finding is open" {
			t.Errorf("verdict = %v, want the verbatim verdict line", sent["verdict"])
		}
		if v, ok := sent["recordedAt"]; !ok || v == "" || v == "0001-01-01T00:00:00Z" {
			t.Errorf("recordedAt = %v, want a non-zero instant the CLI stamped", v)
		}
	})

	t.Run("falls back to the journal", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "verdict", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-451", "-guard", "check-unfinished-work", "-worktree", "/wt/kan-451",
				"-verdict", "OUTSTANDING: /wt/kan-451 -- one item unchecked"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0 (a dead store must never block); stderr:\n%s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "store unreachable") {
			t.Errorf("stderr = %q, want it to name the store as unreachable", stderr.String())
		}

		entries, exists := recordJournalEntries(t, repo, "kan-451")
		if !exists || len(entries) != 1 {
			t.Fatalf("record journal entries = %d (exists=%v), want exactly 1", len(entries), exists)
		}
		var body struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(entries[0].Body, &body); err != nil {
			t.Fatalf("decode journalled body: %v", err)
		}
		if body.Kind != "verdict" {
			t.Errorf("journalled kind = %q, want verdict", body.Kind)
		}
	})
}

// TestRecordVerdictFalsePositiveRefusesNoVerdictAndJournalsUnreachable pins
// `flow record verdict false-positive`'s three outcomes: a 404 (no verdict
// recorded yet) is a definitive refusal, an unreachable store journals for
// replay, and an empty -reason is a caller mistake refused before the store
// is ever contacted.
func TestRecordVerdictFalsePositiveRefusesNoVerdictAndJournalsUnreachable(t *testing.T) {
	t.Run("404 refuses without journalling", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"no verdict for guard"}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "verdict", "false-positive", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-451", "-guard", "check-unfinished-work", "-reason", "verified structural"},
			strings.NewReader(""), &stdout, &stderr)

		if code == 0 {
			t.Fatalf("exit code = 0, want non-zero for a 404; stdout:\n%s", stdout.String())
		}
		if _, exists := recordJournalEntries(t, repo, "kan-451"); exists {
			t.Error("a 404 (definitive refusal) wrote a record journal")
		}
	})

	t.Run("dead port journals for replay", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "verdict", "false-positive", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-451", "-guard", "check-unfinished-work", "-reason", "verified structural"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0 (a dead store must never block); stderr:\n%s", code, stderr.String())
		}
		entries, exists := recordJournalEntries(t, repo, "kan-451")
		if !exists || len(entries) != 1 {
			t.Fatalf("record journal entries = %d (exists=%v), want exactly 1", len(entries), exists)
		}
		var body struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(entries[0].Body, &body); err != nil {
			t.Fatalf("decode journalled body: %v", err)
		}
		if body.Kind != "verdict-false-positive" {
			t.Errorf("journalled kind = %q, want verdict-false-positive", body.Kind)
		}
	})

	t.Run("empty reason exits 2 without contacting the store", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		contacted := false
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
			contacted = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "verdict", "false-positive", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-451", "-guard", "check-unfinished-work", "-reason", ""},
			strings.NewReader(""), &stdout, &stderr)

		if code != 2 {
			t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
		}
		if contacted {
			t.Error("the store was contacted for an empty -reason -- it must be refused first")
		}
		if !strings.Contains(stderr.String(), "-reason is required") {
			t.Errorf("stderr = %q, want it to name -reason as required", stderr.String())
		}
	})
}

// TestRecordVerdictsPrintsArrayAndFailsLoudly pins the read verb's contract
// -- the same one `findings` already carries -- plus its two query
// parameters landing on the request.
func TestRecordVerdictsPrintsArrayAndFailsLoudly(t *testing.T) {
	t.Run("prints the array verbatim", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		body := `[{"id":1,"guard":"check-unfinished-work","worktree":"/wt/kan-451","verdict":"CLEAR: /wt/kan-451 -- ok","recordedAt":"2026-01-02T03:04:05Z","falsePositive":false}]`
		var gotQuery string
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(body))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "verdicts", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-guard", "check-unfinished-work", "-false-positive"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if got := strings.TrimRight(stdout.String(), "\n"); got != body {
			t.Errorf("stdout = %q, want the array verbatim %q", got, body)
		}
		if !strings.Contains(gotQuery, "guard=check-unfinished-work") {
			t.Errorf("query = %q, want guard=check-unfinished-work", gotQuery)
		}
		if !strings.Contains(gotQuery, "falsePositive=true") {
			t.Errorf("query = %q, want falsePositive=true", gotQuery)
		}
	})

	t.Run("empty for no rows", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "verdicts", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if got := strings.TrimRight(stdout.String(), "\n"); got != "[]" {
			t.Fatalf("stdout = %q, want exactly []", got)
		}
	})

	t.Run("dead port fails loudly", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "verdicts", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo},
			strings.NewReader(""), &stdout, &stderr)

		if code == 0 {
			t.Fatalf("exit code = 0, want non-zero; stdout:\n%s stderr:\n%s", stdout.String(), stderr.String())
		}
		if strings.Contains(stdout.String(), "[") {
			t.Errorf("stdout = %q, want no JSON array when the store is unreachable", stdout.String())
		}
		if stderr.Len() == 0 {
			t.Error("stderr is empty, want it to report the unreachable store")
		}
	})
}

// TestRecordIncidentRejectsNegativeMinutesWithoutContactingStore pins
// `flow record incident`'s caller-mistake checks -- a -minutes-lost that
// does not parse as a non-negative integer is refused before the store is
// ever contacted -- and that an omitted -change (incidents take no
// required change identity) leaves the field out of the request body.
func TestRecordIncidentRejectsNegativeMinutesWithoutContactingStore(t *testing.T) {
	for _, minutes := range []string{"-1", "abc"} {
		t.Run(minutes, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "incident", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
					"-guard", "check-unfinished-work", "-symptom", "s", "-recovery", "r",
					"-minutes-lost", minutes},
				strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for an invalid -minutes-lost -- it must be refused first")
			}
		})
	}

	t.Run("valid write without -change omits change from the body", func(t *testing.T) {
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
			_, _ = w.Write([]byte(`{"id":1,"guard":"check-unfinished-work","symptom":"s","recovery":"r","minutesLost":5}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "incident", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-guard", "check-unfinished-work", "-symptom", "s", "-recovery", "r",
				"-minutes-lost", "5"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		if _, ok := sent["change"]; ok {
			t.Errorf("body carries change = %v, want it omitted when -change is not given", sent["change"])
		}
	})
}

// TestRecordIncidentsPrintsArray pins the second read verb's contract --
// identical to `verdicts`' own.
func TestRecordIncidentsPrintsArray(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	body := `[{"id":1,"guard":"check-unfinished-work","symptom":"s","recovery":"r","minutesLost":5,"occurredAt":"2026-01-02T03:04:05Z"}]`
	srv := httptest.NewServer(renderDaemon(t, body))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "incidents", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := strings.TrimRight(stdout.String(), "\n"); got != body {
		t.Errorf("stdout = %q, want the array verbatim %q", got, body)
	}
}

// TestRecordFindingRejectsEmptyReproducer pins that an empty -reproducer is
// refused before the store is contacted.
func TestRecordFindingRejectsEmptyReproducer(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-ref", "F1", "-round", "0", "-slot", "principles",
			"-severity", "major", "-status", "open", "-reproducer", "",
			"-note", "the note"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an empty reproducer -- it must be refused first")
	}
	if !strings.Contains(stderr.String(), "reproducer") {
		t.Errorf("stderr does not name the reproducer field:\n%s", stderr.String())
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("an empty reproducer wrote a record journal")
	}
}

// TestRecordFindingRejectsBareNoneReproducer pins that a "none" reproducer
// with no reason following it -- whitespace-only counting as none -- is
// refused before the store is contacted, since "none — <reason>" is the
// only legal way to say "not reproducible."
func TestRecordFindingRejectsBareNoneReproducer(t *testing.T) {
	for _, reproducer := range []string{"none", "none  "} {
		t.Run(reproducer, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
					"-change", "kan-258", "-ref", "F1", "-round", "0", "-slot", "principles",
					"-severity", "major", "-status", "open", "-reproducer", reproducer,
					"-note", "the note"},
				strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for a bare none reproducer -- it must be refused first")
			}
			if !strings.Contains(stderr.String(), "none — ") {
				t.Errorf("stderr does not name the required \"none — <reason>\" form:\n%s", stderr.String())
			}
			if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
				t.Error("a bare none reproducer wrote a record journal")
			}
		})
	}
}

// TestRecordFindingAcceptsWellFormedStatusAndReproducer is the control
// case: a well-formed status and reproducer still reach the fake server,
// so the validation above rejects only what it names and nothing else.
func TestRecordFindingAcceptsWellFormedStatusAndReproducer(t *testing.T) {
	cases := []struct {
		name       string
		status     string
		reproducer string
	}{
		{"open with a command reproducer", "open", "scripts/x.sh"},
		{"withdrawn with a reason and a none reproducer", "withdrawn duplicate of F1", "none — duplicate"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"ref":"F1","round":0,"slot":"principles","severity":"major","note":"n","status":"open"}`))
			}))
			defer srv.Close()

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
					"-change", "kan-258", "-ref", "F1", "-round", "0", "-slot", "principles",
					"-severity", "major", "-status", c.status, "-reproducer", c.reproducer,
					"-note", "the note"},
				strings.NewReader(""), &stdout, &stderr)

			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
			}
			if !contacted {
				t.Error("a well-formed status and reproducer never reached the fake server")
			}
		})
	}
}

// TestRecordStatusRejectsUnknownStatus pins that `record status` applies
// the same status vocabulary check as `record finding`, before the store
// is contacted.
func TestRecordStatusRejectsUnknownStatus(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "status", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-ref", "F1", "-status", "in-review"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an unrecognised status -- it must be refused first")
	}
	for _, want := range []string{"open", "fixed", "withdrawn"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr does not name the accepted vocabulary %q:\n%s", want, stderr.String())
		}
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("an unrecognised status wrote a record journal")
	}
}

// TestRecordStatusRejectsWithdrawnWithNoReason pins that `record status`
// applies the same reasonless-withdrawn check as `record finding`.
func TestRecordStatusRejectsWithdrawnWithNoReason(t *testing.T) {
	for _, status := range []string{"withdrawn", "withdrawn   "} {
		t.Run(status, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusNoContent)
			}))
			defer srv.Close()

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "status", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
					"-change", "kan-258", "-ref", "F1", "-status", status},
				strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for a withdrawn status with no reason -- it must be refused first")
			}
			if !strings.Contains(stderr.String(), "withdrawn") {
				t.Errorf("stderr does not name the accepted vocabulary:\n%s", stderr.String())
			}
			if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
				t.Error("a reasonless withdrawn status wrote a record journal")
			}
		})
	}
}

// --- findings (read verb) ---

// TestRecordFindingsPrintsJSONArray pins the success shape: the store's
// two findings come back as a JSON array on stdout, decodable into objects
// carrying ref/status/reproducer -- the only fields a guard consuming this
// verb needs.
func TestRecordFindingsPrintsJSONArray(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	body := `{"change":"demo","dispatches":[],"findings":[
	  {"ref":"F1","round":0,"slot":"Bugbot","severity":"Minor","location":"a.go:1","note":"n1","status":"fixed","reproducer":"none — prose only"},
	  {"ref":"F2","round":0,"slot":"Security","severity":"Major","location":"b.go:2","note":"n2","status":"open","reproducer":"scripts/x.sh"}
	]}`
	srv := httptest.NewServer(renderDaemon(t, body))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "findings", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var got []struct {
		Ref        string `json:"ref"`
		Status     string `json:"status"`
		Reproducer string `json:"reproducer"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout as JSON array: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(got) != 2 {
		t.Fatalf("findings = %d, want 2:\n%s", len(got), stdout.String())
	}
	if got[0].Ref != "F1" || got[0].Status != "fixed" || got[0].Reproducer != "none — prose only" {
		t.Errorf("got[0] = %+v, want F1/fixed/\"none — prose only\"", got[0])
	}
	if got[1].Ref != "F2" || got[1].Status != "open" || got[1].Reproducer != "scripts/x.sh" {
		t.Errorf("got[1] = %+v, want F2/open/scripts/x.sh", got[1])
	}
}

// TestRecordFindingsWithNoRowsPrintsEmptyArray pins the "no rows for this
// change" outcome -- the store answering 404/ErrNotFound, exactly as
// TestRecordRenderWithNoLedgerRowsPrintsMissingAndWritesNothing's own
// no-rows case -- as an empty JSON array and exit 0, not a failure: a
// change the store has never heard of has no findings, which is a fact,
// not an error.
func TestRecordFindingsWithNoRowsPrintsEmptyArray(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("findings sent a %s request; a read only reads", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "findings", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := strings.TrimRight(stdout.String(), "\n"); got != "[]" {
		t.Fatalf("stdout = %q, want exactly []", got)
	}
}

// TestRecordFindingsStoreUnreachableExitsNonZero pins the
// new-findings-verb decision: unlike `record render`'s `journalled:`
// fallback, a failed read has nothing to replay, so it must never report
// success for a question it could not answer -- non-zero exit, and no
// JSON array on stdout for a caller to misread as "no findings."
func TestRecordFindingsStoreUnreachableExitsNonZero(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "findings", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo,
			"-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero; stdout:\n%s stderr:\n%s", stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), "[") {
		t.Errorf("stdout = %q, want no JSON array when the store is unreachable", stdout.String())
	}
}

// TestRecordFindingsRequiresChange pins that `findings` shares the same
// caller-mistake contract every other record subcommand uses:
// requireRecordFlags's message, exit 2.
func TestRecordFindingsRequiresChange(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "findings", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stdout:\n%s stderr:\n%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "-change is required") {
		t.Errorf("stderr = %q, want it to name -change as required", stderr.String())
	}
}

// TestValidateFindingStatusRejectsWithdrawnConcatenatedNoSpace pins panel
// finding F1 (kan-271): strings.CutPrefix(status, "withdrawn") strips only
// the literal prefix, with no separator check, so a concatenated value like
// "withdrawnfoo" is wrongly accepted as a legal status.
func TestValidateFindingStatusRejectsWithdrawnConcatenatedNoSpace(t *testing.T) {
	if err := validateFindingStatus("withdrawnfoo"); err == nil {
		t.Fatal(`validateFindingStatus("withdrawnfoo") = nil, want an error -- no space before the reason`)
	}
}

// TestValidateFindingStatusDeferred pins design.md's `deferred <reason>`
// section: the CLI accepts "deferred <reason>" -- the exact `withdrawn`
// shape, checked before the store is ever contacted -- and refuses a bare
// "deferred" the same way a bare "withdrawn" is refused.
func TestValidateFindingStatusDeferred(t *testing.T) {
	for _, status := range []string{"deferred", "deferred   ", "deferredfoo"} {
		if err := validateFindingStatus(status); err == nil {
			t.Errorf("validateFindingStatus(%q) = nil, want an error -- deferred requires a reason", status)
		}
	}
	if err := validateFindingStatus("deferred cosmetic, not worth a fix round"); err != nil {
		t.Errorf("validateFindingStatus with a reason = %v, want nil", err)
	}
}

// TestValidateFindingReproducerRejectsWhitespaceOnly pins panel finding F2
// (kan-271): strings.Fields(" ") returns an empty slice, so a whitespace-only
// reproducer is treated as neither empty nor a bare "none" and wrongly
// accepted as a legal reproducer.
func TestValidateFindingReproducerRejectsWhitespaceOnly(t *testing.T) {
	if err := validateFindingReproducer(" "); err == nil {
		t.Fatal(`validateFindingReproducer(" ") = nil, want an error -- whitespace-only is not a legal reproducer`)
	}
}

// TestRecordFindingsWithZeroFindingsOnExistingChangePrintsEmptyArray pins
// panel finding F5 (kan-271): a change that exists (carries dispatch rows)
// but has raised zero findings must still print "[]", not the literal
// "null" a nil Findings slice marshals to -- both rewritten guards' jq
// pipelines start with ".[]" and crash on a top-level null under
// set -euo pipefail.
func TestRecordFindingsWithZeroFindingsOnExistingChangePrintsEmptyArray(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	body := `{"change":"demo","dispatches":[{"id":1,"seq":1,"role":"implementer","model":"sonnet","startedAt":"2026-01-01T00:00:00Z"}],"findings":null}`
	srv := httptest.NewServer(renderDaemon(t, body))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "findings", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	got := strings.TrimRight(stdout.String(), "\n")
	if got != "[]" {
		t.Errorf("stdout = %q, want exactly \"[]\" for a change with dispatches but zero findings", got)
	}
}

// --- dispatches (read verb) ---

// TestRunRecordDispatchesPrintsDispatchesAsJSONArray pins the success
// shape: the change's dispatch rows come back as a JSON array on stdout in
// the order the store sent them (seq order, pinned store-side), decodable
// into objects carrying key/role/sessionToken -- what
// check-panel-fix-single-dispatch.sh filters and counts on.
func TestRunRecordDispatchesPrintsDispatchesAsJSONArray(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	body := `{"change":"demo","findings":[],"dispatches":[
	  {"id":1,"seq":1,"key":"panel-fix-1","role":"panel-fix","model":"sonnet","sessionToken":"mf-tok","startedAt":"2026-09-10T09:00:00Z"},
	  {"id":2,"seq":2,"key":"panel-fix-1-retry","role":"panel-fix","model":"sonnet","sessionToken":"mf-tok","startedAt":"2026-09-10T09:05:00Z"}
	]}`
	srv := httptest.NewServer(renderDaemon(t, body))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatches", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var got []struct {
		Seq          int    `json:"seq"`
		Key          string `json:"key"`
		Role         string `json:"role"`
		SessionToken string `json:"sessionToken"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout as JSON array: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(got) != 2 {
		t.Fatalf("dispatches = %d, want 2:\n%s", len(got), stdout.String())
	}
	if got[0].Seq != 1 || got[0].Key != "panel-fix-1" || got[0].Role != "panel-fix" || got[0].SessionToken != "mf-tok" {
		t.Errorf("got[0] = %+v, want seq 1/panel-fix-1/panel-fix/mf-tok", got[0])
	}
	if got[1].Seq != 2 || got[1].Key != "panel-fix-1-retry" {
		t.Errorf("got[1] = %+v, want seq 2/panel-fix-1-retry", got[1])
	}
}

// TestRunRecordDispatchesEmptyChangePrintsEmptyArray pins the no-rows
// outcome -- the store answering 404/ErrNotFound, exactly as the findings
// verb treats it -- as an empty JSON array and exit 0: a change the store
// has never heard of has no dispatches, which is a fact, not an error.
func TestRunRecordDispatchesEmptyChangePrintsEmptyArray(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("dispatches sent a %s request; a read only reads", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatches", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := strings.TrimRight(stdout.String(), "\n"); got != "[]" {
		t.Fatalf("stdout = %q, want exactly []", got)
	}
}

// TestRunRecordDispatchesUnreachableStoreFails pins the findings verb's
// read contract for its sibling: a failed read never journals and never
// reports success for a question it could not answer -- non-zero exit, and
// no JSON array on stdout for a caller to misread as "no dispatches."
func TestRunRecordDispatchesUnreachableStoreFails(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatches", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo,
			"-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero; stdout:\n%s stderr:\n%s", stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), "[") {
		t.Errorf("stdout = %q, want no JSON array when the store is unreachable", stdout.String())
	}
	if !strings.Contains(stderr.String(), "flow: dispatches:") {
		t.Errorf("stderr = %q, want it to name the failing verb", stderr.String())
	}
}

// TestRunRecordDispatchesNullDispatchesPrintsEmptyArray pins the nil-guard
// on an EXISTING change whose dispatches field arrives null -- the shape
// TestRecordFindingsWithZeroFindingsOnExistingChangePrintsEmptyArray pins
// for the findings verb, and the mutant kan-482 pass 1 flagged (M3): the
// same "[]"-not-"null" rule, because the fix-single-dispatch guard's jq
// pipeline starts with .[] and crashes on a top-level null under set -euo
// pipefail.
func TestRunRecordDispatchesNullDispatchesPrintsEmptyArray(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	body := `{"change":"demo","findings":[{"ref":"F1","round":0,"slot":"Bugbot","severity":"Minor","location":"a.go:1","note":"n1","status":"fixed","reproducer":"none — prose only"}],"dispatches":null}`
	srv := httptest.NewServer(renderDaemon(t, body))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatches", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	got := strings.TrimRight(stdout.String(), "\n")
	if got != "[]" {
		t.Errorf("stdout = %q, want exactly \"[]\" for an existing change with null dispatches", got)
	}
}

// TestRunRecordDecisionRefusesBadBody pins the caller-mistake path `flow
// record decision` shares with every other write verb: a body that is not
// valid JSON is refused before the store is ever contacted, exit 2, no
// journal entry -- the validateFindingStatus posture the dispatch prompt
// names.
func TestRunRecordDecisionRefusesBadBody(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	file := filepath.Join(t.TempDir(), "decision.json")
	if err := os.WriteFile(file, []byte("not json"), 0o644); err != nil {
		t.Fatalf("write decision file: %v", err)
	}

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "decision", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-472", "-session-token", "mf-abc123", "-file", file},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for a non-JSON body -- it must be refused first")
	}
	if _, exists := recordJournalEntries(t, repo, "kan-472"); exists {
		t.Error("a caller mistake wrote a record journal")
	}
}

// TestRunRecordDecisionJournalsWhenUnreachable pins the never-block
// fallback every other write verb already carries: an unreachable store
// journals the decision under kind "decision" and still exits 0.
func TestRunRecordDecisionJournalsWhenUnreachable(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	file := filepath.Join(t.TempDir(), "decision.json")
	body := `{"toggles":{"executionMode":"dynamic"},"class":"regular"}`
	if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
		t.Fatalf("write decision file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "decision", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
			"-change", "kan-472", "-session-token", "mf-abc123", "-file", file},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a dead store must never block); stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "store unreachable") {
		t.Errorf("stderr = %q, want it to name the store as unreachable", stderr.String())
	}

	entries, exists := recordJournalEntries(t, repo, "kan-472")
	if !exists || len(entries) != 1 {
		t.Fatalf("record journal entries = %d (exists=%v), want exactly 1", len(entries), exists)
	}
	var got struct {
		Kind    string `json:"kind"`
		Request struct {
			SessionToken string          `json:"sessionToken"`
			Decision     json.RawMessage `json:"decision"`
		} `json:"request"`
	}
	if err := json.Unmarshal(entries[0].Body, &got); err != nil {
		t.Fatalf("decode journalled body: %v", err)
	}
	if got.Kind != "decision" {
		t.Errorf("journalled kind = %q, want decision", got.Kind)
	}
	if got.Request.SessionToken != "mf-abc123" {
		t.Errorf("journalled sessionToken = %q, want mf-abc123", got.Request.SessionToken)
	}
	if string(got.Request.Decision) != body {
		t.Errorf("journalled decision = %s, want %s", got.Request.Decision, body)
	}
}

// TestRunRecordDecisionsPrintsJSON pins `flow record decisions`' read
// contract -- `incidents`'/`verdicts`' own, verbatim: the store's array,
// printed as-is, and an empty array (never null) when the store holds none.
func TestRunRecordDecisionsPrintsJSON(t *testing.T) {
	t.Run("prints the array verbatim", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		body := `[{"id":2,"sessionToken":"mf-def","recordedAt":"2026-01-02T03:04:05Z","decision":{"class":"big"}},` +
			`{"id":1,"sessionToken":"mf-abc","recordedAt":"2026-01-01T00:00:00Z","decision":{"class":"small"}}]`
		var gotPath string
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(body))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "decisions", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-472"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if !strings.HasSuffix(gotPath, "/kan-472/decisions") {
			t.Errorf("request path = %s, want it to end in /kan-472/decisions", gotPath)
		}
		got := strings.TrimRight(stdout.String(), "\n")
		if got != body {
			t.Errorf("stdout = %s, want the store's array verbatim: %s", got, body)
		}
	})

	t.Run("no rows prints an empty array", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "decisions", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-472"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		got := strings.TrimRight(stdout.String(), "\n")
		if got != "[]" {
			t.Errorf("stdout = %q, want exactly \"[]\"", got)
		}
	})

	t.Run("store unreachable exits non-zero", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "decisions", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-472"},
			strings.NewReader(""), &stdout, &stderr)

		if code == 0 {
			t.Fatalf("exit code = 0, want non-zero -- a read has nothing to journal; stdout:\n%s", stdout.String())
		}
	})
}

// TestRecordUsageNamesEveryRole pins the usage text's "-role is one of:"
// line to recordRoles in both directions, so neither can drift out of sync
// with the other: a role added to the slice cannot stay missing from the
// help a caller reads -- which is how "verifier" was accepted by the CLI
// while the usage still listed six roles -- and a role dropped from the
// slice cannot stay advertised as accepted by a stale usage line.
func TestRecordUsageNamesEveryRole(t *testing.T) {
	for _, role := range recordRoles {
		if !strings.Contains(recordUsage, " "+role+",") && !strings.Contains(recordUsage, " "+role+".") {
			t.Errorf("recordUsage does not list role %q on its -role line", role)
		}
	}

	const marker = "-role is one of: "
	var usageLine string
	for _, line := range strings.Split(recordUsage, "\n") {
		if strings.Contains(line, marker) {
			usageLine = line
			break
		}
	}
	if usageLine == "" {
		t.Fatal(`recordUsage has no "-role is one of:" line`)
	}
	list := strings.TrimSuffix(strings.TrimSpace(strings.SplitN(usageLine, marker, 2)[1]), ".")

	known := make(map[string]bool, len(recordRoles))
	for _, role := range recordRoles {
		known[role] = true
	}
	for _, role := range strings.Split(list, ", ") {
		if !known[role] {
			t.Errorf("recordUsage's -role line names %q, which is not in recordRoles", role)
		}
	}
}

// TestRunRecordDispatchBeginRejectsUnknownEffort pins that -effort is
// checked against the four accepted words before the store is ever
// contacted, exactly as -role is in
// TestRecordRejectsUnknownRoleWithoutContactingStore: a caller mistake
// taking the never-block fallback path would journal a write a replay
// could only ever be refused for a second time.
func TestRunRecordDispatchBeginRejectsUnknownEffort(t *testing.T) {
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
		[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-effort", "extreme",
			"-key", "k1", "-session-token", "mf-record-unknown-effort", "-started-at", "2026-01-02T03:04:05Z"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an unrecognised effort -- it must be refused first")
	}
	for _, effort := range []string{"low", "medium", "high", "default"} {
		if !strings.Contains(stderr.String(), effort) {
			t.Errorf("stderr does not name the accepted effort %q:\n%s", effort, stderr.String())
		}
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("an unrecognised effort wrote a record journal")
	}
}

// TestRecordDispatchBeginDefaultsEffortToDefault pins that an omitted
// -effort is sent as the literal "default" rather than left absent from
// the wire body -- the store column is NOT NULL DEFAULT 'default' because
// an absent effort is a known fact (the dispatcher set none), never an
// unknown one, and the CLI is where that fact is established.
func TestRecordDispatchBeginDefaultsEffortToDefault(t *testing.T) {
	sent := dispatchBeginBody(t)

	if sent["effort"] != "default" {
		t.Errorf("effort = %v, want %q when -effort is omitted", sent["effort"], "default")
	}
}

// TestRecordDispatchBeginAcceptsEffort pins -effort through to the request
// body verbatim for each of the three accepted non-default words.
func TestRecordDispatchBeginAcceptsEffort(t *testing.T) {
	for _, effort := range []string{"low", "medium", "high"} {
		t.Run(effort, func(t *testing.T) {
			sent := dispatchBeginBody(t, "-effort", effort)

			if sent["effort"] != effort {
				t.Errorf("effort = %v, want %q", sent["effort"], effort)
			}
		})
	}
}

// --- the pass log (KAN-331) ---

// TestRecordPassCommand pins what the pass-log write sends and says: a POST
// to .../passes carrying round and note verbatim, one stdout line naming the
// allocated row id, exit 0.
func TestRecordPassCommand(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"id":11,"round":1,"note":"not re-run"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "pass", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-round", "1", "-note", "not re-run — nothing new since its last read"},
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
	if !strings.Contains(stdout.String(), "recorded: pass 11") {
		t.Errorf("stdout = %q, want it to name the recorded row id 11", stdout.String())
	}
	if !strings.HasSuffix(gotPath, "/kan-258/passes") {
		t.Errorf("request path = %s, want it to end in /kan-258/passes", gotPath)
	}

	var sent map[string]any
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
	}
	if sent["round"] != float64(1) {
		t.Errorf("round = %v, want 1", sent["round"])
	}
	if sent["note"] != "not re-run — nothing new since its last read" {
		t.Errorf("note = %v, want the line verbatim", sent["note"])
	}
}

// TestRecordMutationCommand pins the mutation-proof write the same way: a
// POST to .../mutations carrying all three of the contract line's fields,
// one stdout line, exit 0.
func TestRecordMutationCommand(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"id":12,"round":1,"path":"render.go","mutated":"none","test":"guard was equivalent"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "mutation", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-round", "1", "-path", "render.go",
			"-mutated", "none", "-test", "guard was equivalent"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "recorded: mutation 12") {
		t.Errorf("stdout = %q, want it to name the recorded row id 12", stdout.String())
	}
	if !strings.HasSuffix(gotPath, "/kan-258/mutations") {
		t.Errorf("request path = %s, want it to end in /kan-258/mutations", gotPath)
	}

	var sent map[string]any
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
	}
	if sent["path"] != "render.go" || sent["mutated"] != "none" || sent["test"] != "guard was equivalent" {
		t.Errorf("body = %v, want the three contract fields verbatim", sent)
	}
}

// TestRecordPassAndMutationMissingFlagsExitTwo pins that a missing required
// field is refused before the store is contacted and leaves no journal
// behind -- the same caller-mistake contract every other record write
// carries.
func TestRecordPassAndMutationMissingFlagsExitTwo(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"pass without note", []string{"record", "pass", "-change", "kan-258"}},
		{"mutation without path", []string{"record", "mutation", "-change", "kan-258", "-mutated", "x", "-test", "y"}},
		{"mutation without test", []string{"record", "mutation", "-change", "kan-258", "-path", "p", "-mutated", "none"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			var stdout, stderr bytes.Buffer
			code := run(context.Background(), tc.args, strings.NewReader(""), &stdout, &stderr)
			if code != 2 {
				t.Fatalf("exit code = %d, want 2", code)
			}
			if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
				t.Errorf("a caller mistake left a record journal behind")
			}
		})
	}
}
