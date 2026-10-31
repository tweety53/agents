package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/records"
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
		repo := recordRepo(t, "kan-258")
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
			[]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-task", "6", "-role", "implementer",
				"-model", "unknown (agent-defined)", "-key", "task-6-implementer",
				"-session-token", "mf-record-dispatch-ok"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if stderr.Len() != 0 {
			t.Errorf("stderr = %q, want empty on a clean success", stderr.String())
		}
		// A live begin follows its recorded: line with the launch lines
		// (decision sr3-begin-prints-launch-lines); this change has no plan,
		// so only the Role line.
		if got, want := stdout.String(), "recorded: dispatch 3\nRole: implementer\n"; got != want {
			t.Errorf("stdout = %q, want %q", got, want)
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
		if sent["startedAt"] != "0001-01-01T00:00:00Z" {
			t.Errorf("startedAt = %v, want the zero instant -- the daemon stamps the moment it writes the row, never the caller (KAN-324)", sent["startedAt"])
		}
	})

	t.Run("finding created", func(t *testing.T) {
		repo := recordRepo(t, "kan-258")
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
		repo := recordRepo(t, "kan-258")
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
		repo := recordRepo(t, "kan-258")
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
			args: []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-change", "kan-258", "-task", "6",
				"-role", "implementer", "-model", "opus", "-key", "task-6-implementer",
				"-session-token", "mf-record-dispatch-journal"},
			kind: "dispatch",
		},
		{
			name: "dispatch end",
			args: []string{"record", "dispatch", "end", "-change", "kan-258", "-key", "task-6-implementer",
				"-session-token", "mf-record-dispatch-journal", "-commit", "abc1234000000000000000000000000000000000",
				"-outcome", "completed"},
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
			repo := recordRepo(t, "kan-258")
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

// TestRecordDispatchEndJournalMarksTheAttemptInstant pins the CLI side of
// the replay fix: an end the store could not reach, with no -ended-at, marks
// its journalled request endedAtJournalled so the replay closes the row at
// the attempt instant; an explicit -ended-at is a real end and carries no
// mark.
func TestRecordDispatchEndJournalMarksTheAttemptInstant(t *testing.T) {
	for _, tc := range []struct {
		name  string
		extra []string
		want  bool
	}{
		{"no ended-at", nil, true},
		{"ended-at", []string{"-ended-at", "2026-01-02T03:04:05Z"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := recordRepo(t, "kan-258")
			isolatedStateRoot(t)
			args := []string{"record", "dispatch", "end", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-258", "-key", "task-6-implementer", "-session-token", "mf-record-journal-mark",
				"-commit", "abc1234000000000000000000000000000000000", "-outcome", "completed"}
			args = append(args, tc.extra...)
			var stdout, stderr bytes.Buffer
			if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
			}
			entries, exists := recordJournalEntries(t, repo, "kan-258")
			if !exists || len(entries) != 1 {
				t.Fatalf("record journal entries = %v (exists=%v), want 1", entries, exists)
			}
			var body struct {
				Kind    string `json:"kind"`
				Request struct {
					EndedAtJournalled bool `json:"endedAtJournalled"`
				} `json:"request"`
			}
			if err := json.Unmarshal(entries[0].Body, &body); err != nil {
				t.Fatalf("decode journalled body: %v", err)
			}
			if body.Kind != "dispatch-end" {
				t.Fatalf("journalled kind = %q, want dispatch-end", body.Kind)
			}
			if body.Request.EndedAtJournalled != tc.want {
				t.Errorf("journalled endedAtJournalled = %v, want %v", body.Request.EndedAtJournalled, tc.want)
			}
		})
	}
}

// TestRecordDispatchJournalCarriesTheAttemptInstant pins the replay
// override's source: a begin or end the store could not reach stamps the
// instant its attempt was made into the journalled request, so the replay
// restores that instant instead of dating the dispatch to the replay
// (KAN-324). The live request carries nothing -- the daemon stamps its own
// clock -- so the journal is the only place this CLI ever supplies an
// instant.
func TestRecordDispatchJournalCarriesTheAttemptInstant(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		field string
	}{
		{"begin", []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-change", "kan-258", "-task", "6",
			"-role", "implementer", "-model", "opus", "-key", "task-6-implementer",
			"-session-token", "mf-record-journal-instant"}, "startedAt"},
		{"end", []string{"record", "dispatch", "end", "-change", "kan-258", "-key", "task-6-implementer",
			"-session-token", "mf-record-journal-instant", "-commit", "abc1234000000000000000000000000000000000",
			"-outcome", "completed"}, "endedAt"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := recordRepo(t, "kan-258")
			isolatedStateRoot(t)

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

			entries, exists := recordJournalEntries(t, repo, "kan-258")
			if !exists || len(entries) != 1 {
				t.Fatalf("record journal entries = %v (exists=%v), want 1", entries, exists)
			}
			var body struct {
				Kind    string          `json:"kind"`
				Request json.RawMessage `json:"request"`
			}
			if err := json.Unmarshal(entries[0].Body, &body); err != nil {
				t.Fatalf("decode journalled body: %v", err)
			}
			var sent map[string]any
			if err := json.Unmarshal(body.Request, &sent); err != nil {
				t.Fatalf("decode journalled request: %v", err)
			}
			raw, _ := sent[tc.field].(string)
			stamped, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				t.Fatalf("journalled %s = %q, want an RFC 3339 instant the replay can restore", tc.field, raw)
			}
			if time.Since(stamped) > time.Minute {
				t.Errorf("journalled %s = %s, want the attempt instant (within a minute of the call)", tc.field, stamped)
			}
		})
	}
}

// TestRecordMissingRequiredFlagExitsTwoWithoutJournalling pins the other
// half of the never-block split: a caller mistake is not a store failure,
// so it exits 2 and leaves no journal entry to replay. Journalling it
// would queue a write that can never succeed.
func TestRecordMissingRequiredFlagExitsTwoWithoutJournalling(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"dispatch begin without -model", []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-change", "kan-258",
			"-role", "implementer", "-key", "task-1-implementer", "-session-token", "mf-record-missing-model"}},
		{"dispatch begin without -change", []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none",
			"-role", "implementer", "-model", "opus", "-key", "task-1-implementer", "-session-token", "mf-record-missing-change"}},
		{"dispatch begin with retired -started-at", []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-change", "kan-258",
			"-role", "implementer", "-model", "opus", "-key", "task-1-implementer", "-session-token", "mf-record-retired-started",
			"-started-at", "2026-01-02T03:04:05Z"}},
		{"dispatch begin without -key", []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-change", "kan-258",
			"-role", "implementer", "-model", "opus", "-session-token", "mf-record-missing-key"}},
		{"dispatch end without -key", []string{"record", "dispatch", "end", "-change", "kan-258",
			"-session-token", "mf-record-missing-end-key"}},
		{"dispatch end with retired -ended-at", []string{"record", "dispatch", "end", "-change", "kan-258",
			"-key", "task-1-implementer", "-session-token", "mf-record-retired-ended",
			"-ended-at", "2026-01-02T03:44:05Z"}},
		{"finding without -note", []string{"record", "finding", "-change", "kan-258",
			"-ref", "F1", "-slot", "principles", "-severity", "major", "-status", "open"}},
		{"status without -ref", []string{"record", "status", "-change", "kan-258", "-status", "fixed"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := recordRepo(t, "kan-258")
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
	repo := recordRepo(t, "kan-258")
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
		[]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "architect", "-model", "opus", "-key", "task-1-implementer",
			"-session-token", "mf-record-unknown-role"},
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
	repo := recordRepo(t, "kan-258")
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
		[]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "planner", "-model", "opus", "-key", "task-1-implementer",
			"-session-token", "mf-record-planner-role"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !contacted {
		t.Error("the store was not contacted for the planner role")
	}
}

// TestRecordAcceptsExplorerRole pins that "explorer" is in the accepted-role
// allowlist -- the role `/flow` records for an Explore subagent its planning
// session dispatches during flow.brainstorm (kan-1033).
func TestRecordAcceptsExplorerRole(t *testing.T) {
	repo := recordRepo(t, "kan-1033")
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1,"role":"explorer","model":"sonnet","startedAt":"2026-01-02T03:04:05Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-effort", "low", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-1033", "-role", "explorer", "-model", "sonnet", "-key", "explore-1",
			"-session-token", "mf-record-explorer-role"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !contacted {
		t.Error("the store was not contacted for the explorer role")
	}
}

// TestRecordAcceptsMergerRole pins that "merger" is in the accepted-role
// allowlist -- the role `/flow` records for the one-shot subagent that
// folds a panel round's reports into one summary (review-panel.md).
func TestRecordAcceptsMergerRole(t *testing.T) {
	repo := recordRepo(t, "kan-1050")
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1,"role":"merger","model":"sonnet","startedAt":"2026-01-02T03:04:05Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-effort", "low", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-1050", "-role", "merger", "-model", "sonnet", "-key", "panel-0-merger",
			"-session-token", "mf-record-merger-role"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !contacted {
		t.Error("the store was not contacted for the merger role")
	}
}

// TestRecordAcceptsConductorRole pins that "conductor" is in the
// accepted-role allowlist -- the role `/flow` records for the subagent
// that runs implement.md, review-panel.md and verify-and-handoff.md as
// the conductor (design.md's conductor-runs-implementation-half).
func TestRecordAcceptsConductorRole(t *testing.T) {
	repo := recordRepo(t, "kan-258")
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
		[]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "conductor", "-model", "sonnet", "-key", "conductor",
			"-session-token", "mf-record-conductor-role"},
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
	repo := recordRepo(t, "kan-258")
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
		[]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "verifier", "-model", "sonnet", "-key", "task-1-implementer",
			"-session-token", "mf-record-verifier-role"},
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
	repo := recordRepo(t, "kan-258")
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
		[]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-key", "task-1-implementer",
			"-session-token", "mf-$(date +%s)"},
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

// TestRecordDispatchSendsAgentIDOnlyWhenGiven pins the -agent-id flag in
// its three states: an explicit value is carried verbatim as recorded
// intent; a harness that exposes no identifier lets the caller
// pass the literal "none", recorded as no id; and a begin that omits the
// flag is accepted (KAN-322) -- the daemon now stamps the id onto the row
// from the parent transcript's launch result, so the hand-typed call
// KAN-212 predicted would be forgotten is no longer an obligation, and an
// omitted flag means "stamp it automatically", not "no id".
//
// The "none" and "omitted" cases assert the key is missing from the body,
// not that it is empty. "" means "not reported" and must never match
// another absent id during attribution, so a wire form that spelled
// absence as a present, empty value would be a value the daemon could
// store and the attributor could compare.
func TestRecordDispatchSendsAgentIDOnlyWhenGiven(t *testing.T) {
	dispatchArgs := func(repo, addr string, extra ...string) []string {
		args := []string{"record", "dispatch", "begin", "-effort", "default", "-addr", addr, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "reviewer", "-model", "sonnet", "-key", "panel-1-primary", "-slot", "primary",
			"-session-token", "mf-record-agent-id"}
		return append(args, extra...)
	}

	send := func(t *testing.T, extra ...string) map[string]any {
		t.Helper()
		repo := recordRepo(t, "kan-258")
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
		sent := send(t, "-agent-id", "a0123456789abcdef")
		if sent["agentId"] != "a0123456789abcdef" {
			t.Errorf("agentId = %v, want a0123456789abcdef", sent["agentId"])
		}
	})

	t.Run("none", func(t *testing.T) {
		sent := send(t, "-agent-id", "none")
		if v, ok := sent["agentId"]; ok {
			t.Errorf("agentId = %v, want the key absent -- an unreported id is absence, not an empty value", v)
		}
	})

	t.Run("omitted", func(t *testing.T) {
		sent := send(t)
		if v, ok := sent["agentId"]; ok {
			t.Errorf("agentId = %v, want the key absent -- omission means the daemon stamps the id later, not an empty value", v)
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
	repo := recordRepo(t, "kan-327")
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

	args := []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-327", "-role", "reviewer", "-model", "sonnet", "-key", "panel-1-principles", "-slot", "principles",
		"-session-token", "mf-record-diff-base"}

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

// TestRecordDispatchBeginSendsTheSessionID pins KAN-967's CLI half: a
// begin carries the harness session it ran in, so the store can bind a row
// no stage run's token reaches.
func TestRecordDispatchBeginSendsTheSessionID(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "session-kan967")

	sent := dispatchBeginBody(t)

	if sent["sessionId"] != "session-kan967" {
		t.Errorf("sessionId = %v, want session-kan967", sent["sessionId"])
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
			"-change", "kan-258", "-key", "panel-1-primary", "-session-token", "mf-record-end-agent-id"}
		return append(args, extra...)
	}

	send := func(t *testing.T, extra ...string) map[string]any {
		t.Helper()
		repo := recordRepo(t, "kan-258")
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
		sent := send(t, "-agent-id", "a0123456789abcdef")
		if sent["agentId"] != "a0123456789abcdef" {
			t.Errorf("agentId = %v, want a0123456789abcdef", sent["agentId"])
		}
	})

	t.Run("omitted", func(t *testing.T) {
		sent := send(t)
		if v, ok := sent["agentId"]; ok {
			t.Errorf("agentId = %v, want the key absent -- an end that omits it must never clear an identifier begin already recorded", v)
		}
	})

	// The literal "none" is the documented word for "the harness exposes
	// no id", and the begin handler maps it to absence before the call;
	// the end handler mirrors that mapping, so the word reaches the wire
	// as the same absent key here -- never verbatim, where the store's
	// placeholder rule would refuse it (KAN-560, panel F2/F4).
	t.Run("none", func(t *testing.T) {
		sent := send(t, "-agent-id", "none")
		if v, ok := sent["agentId"]; ok {
			t.Errorf("agentId = %v, want the key absent -- \"none\" means absence on both halves of the pair", v)
		}
	})
}

// --- write-time validation: a blocked outcome's cause (KAN-510) ---

// The cause tests share one call shape: every case below passes the same
// minimal end request, varying only `-outcome` and `-cause`.
func causeEndRun(t *testing.T, extra ...string) (int, bytes.Buffer, bytes.Buffer, bool, string) {
	t.Helper()
	repo := recordRepo(t, "kan-510")
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"verifier","model":"sonnet","endedAt":"2026-01-02T03:44:05Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	args := []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-510", "-key", "visual-verify", "-session-token", "ff-kan510-end-cause"}
	args = append(args, extra...)
	code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
	return code, stdout, stderr, contacted, repo
}

// TestDispatchEndCauseRequired pins that `-outcome blocked` without a cause
// is a caller mistake refused before the store is ever contacted: a blocked
// outcome naming no cause is the prose footnote this flag exists to retire.
func TestDispatchEndCauseRequired(t *testing.T) {
	code, _, stderr, contacted, repo := causeEndRun(t, "-outcome", "blocked")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for a blocked outcome with no cause -- it must be refused first")
	}
	if _, exists := recordJournalEntries(t, repo, "kan-510"); exists {
		t.Error("a refused end wrote a record journal")
	}
}

// TestDispatchEndCauseRejectedWithoutBlocked pins the other direction: a
// cause is meaningful only beside a blocked outcome, so carrying one with
// any other outcome is a caller mistake.
func TestDispatchEndCauseRejectedWithoutBlocked(t *testing.T) {
	for _, outcome := range []string{"completed", "stopped", "aborted", ""} {
		t.Run("outcome="+outcome, func(t *testing.T) {
			extra := []string{}
			if outcome != "" {
				extra = append(extra, "-outcome", outcome)
			}
			extra = append(extra, "-cause", "environment")
			code, _, stderr, contacted, _ := causeEndRun(t, extra...)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for a cause with no blocked outcome -- it must be refused first")
			}
		})
	}
}

// TestDispatchEndCauseUnknown pins the closed vocabulary: a cause outside
// `environment`, `test-failure`, `missing-fixture` is refused before the
// store is contacted, the recordRoles/recordEfforts precedent, and the
// usage names the accepted set.
func TestDispatchEndCauseUnknown(t *testing.T) {
	code, _, stderr, contacted, _ := causeEndRun(t, "-outcome", "blocked", "-cause", "weather")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an unrecognised cause -- it must be refused first")
	}
	for _, want := range []string{"environment", "test-failure", "missing-fixture"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr does not name the accepted cause %q:\n%s", want, stderr.String())
		}
	}
}

// TestDispatchEndOutcomeClosedSet pins the closed outcome vocabulary: a
// free-text outcome no contract names (`done`, `approve`) is refused before
// the store is contacted, its usage naming the accepted set, and every
// contract outcome reaches the store.
func TestDispatchEndOutcomeClosedSet(t *testing.T) {
	for _, outcome := range []string{"done", "approve", "Completed"} {
		t.Run("refused="+outcome, func(t *testing.T) {
			code, _, stderr, contacted, _ := causeEndRun(t, "-outcome", outcome)
			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for an outcome outside the set -- it must be refused first")
			}
			if !strings.Contains(stderr.String(), "completed, clean, fix, blocked, aborted, fallback, timed-out") {
				t.Errorf("stderr does not name the accepted outcomes:\n%s", stderr.String())
			}
		})
	}
	for _, outcome := range recordOutcomes {
		t.Run("accepted="+outcome, func(t *testing.T) {
			extra := []string{"-outcome", outcome}
			if outcome == "blocked" {
				extra = append(extra, "-cause", "environment")
			}
			code, _, stderr, contacted, _ := causeEndRun(t, extra...)
			if code != 0 || !contacted {
				t.Fatalf("exit code = %d, contacted = %v, want 0 and the store contacted; stderr:\n%s", code, contacted, stderr.String())
			}
		})
	}
}

// TestDispatchEndCauseAccepted pins the happy path: `-outcome blocked`
// with a cause from the set reaches the wire carrying it.
func TestDispatchEndCauseAccepted(t *testing.T) {
	repo := recordRepo(t, "kan-510")
	isolatedStateRoot(t)

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotBody, err = readAll(r)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"verifier","model":"sonnet","endedAt":"2026-01-02T03:44:05Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-510", "-key", "visual-verify", "-session-token", "ff-kan510-end-cause",
			"-outcome", "blocked", "-cause", "environment"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	var sent map[string]any
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
	}
	if sent["cause"] != "environment" {
		t.Errorf("cause = %v, want environment on the wire", sent["cause"])
	}
}

// --- render ---

// renderRun is the run the render tests' daemon has already rendered: one
// dispatch and one finding, enough for both renderings to have rows.
var renderRun = records.Run{
	Change: "demo",
	Dispatches: []records.Dispatch{{
		ID: 1, Seq: 1, TaskID: "11", Role: "implementer",
		Model: "unknown (agent-defined)", CommitSHA: "abc1234000000000000000000000000000000000", Outcome: "completed",
		StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}},
	Findings: []records.Finding{{
		Ref: "F1", Round: 0, Slot: "Bugbot", Severity: "Minor",
		Location: "a.go:1", Note: "n", Status: "fixed",
		Reproducer: "none — prose only",
	}},
}

// renderDaemon answers the render route with the envelope the map names
// per kind (missing for any kind absent from it) and fails the test if the
// CLI touches any other GET route -- a render fetches, it never asks for
// rows to render itself.
func renderDaemon(t *testing.T, envelopes map[string]records.Rendered) http.HandlerFunc {
	t.Helper()
	return genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("render sent a %s request; a render only reads", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/render/") {
			t.Errorf("render touched %s; the record's rows are the daemon's to render, not the CLI's", r.URL.Path)
		}
		kind := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		env, ok := envelopes[kind]
		if !ok {
			// A kind the test daemon holds no envelope for is a fetch the
			// test never intended: failing loudly here keeps a wrong-kind
			// fetch from reading as a 200 empty record.
			t.Errorf("render fetched kind %q; the test daemon holds no envelope for it", kind)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(env); err != nil {
			t.Errorf("encode envelope: %v", err)
		}
	})
}

// missingRenderLedger is the envelope map for a ledger the store holds no
// dispatch rows for -- the daemon applied the asymmetry, and the CLI's
// part is only to report it.
func missingRenderLedger() map[string]records.Rendered {
	return map[string]records.Rendered{"ledger": {Missing: true}}
}

// renderedPanelEnvelope is the envelope map for a panel the daemon has
// already rendered from run -- for a clean run, the zero form.
func renderedPanelEnvelope(t *testing.T, run records.Run) map[string]records.Rendered {
	t.Helper()
	return map[string]records.Rendered{"panel": mustEnvelope(t, "panel", run)}
}

// mustEnvelope builds the not-missing envelope a daemon answers with for
// a kind over a run.
func mustEnvelope(t *testing.T, kind string, run records.Run) records.Rendered {
	t.Helper()
	body, ok := records.RenderKind(kind, run)
	if !ok {
		t.Fatalf("RenderKind(%s) reported missing for the fixture run; the test needs a body", kind)
	}
	return records.Rendered{Body: body}
}

// runRecordDaemon answers GET /api/v1/records/{project}/{change} with body
// -- the whole-record read the findings and dispatches verbs consume. The
// render tests use renderDaemon instead: their CLI asks for rendered
// documents, never for rows.
func runRecordDaemon(t *testing.T, body string) http.HandlerFunc {
	t.Helper()
	return genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("record read sent a %s request", r.Method)
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

	srv := httptest.NewServer(renderDaemon(t, map[string]records.Rendered{
		"ledger": mustEnvelope(t, "ledger", renderRun),
		"panel":  mustEnvelope(t, "panel", renderRun),
	}))
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

	srv := httptest.NewServer(renderDaemon(t, missingRenderLedger()))
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

	srv := httptest.NewServer(renderDaemon(t, renderedPanelEnvelope(t, records.Run{Change: "demo"})))
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
	// POSTs its verdict to this same FLOW_ADDR, and reads the change's
	// findings through `flow record findings` against it. That write's
	// outcome is irrelevant here -- the guard's own combined output already
	// silences it -- so this handler answers the render route with the
	// zero-form envelope, the self-review bundle with a recorded summary,
	// every other GET with the empty change's rows, and accepts any other
	// method rather than failing the test over it.
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "/render/") {
			_ = json.NewEncoder(w).Encode(records.Rendered{Body: records.RenderPanel(records.Run{Change: "demo"})})
			return
		}
		// The guard reads the bundle header for its recorded-summary signal.
		if strings.HasSuffix(r.URL.Path, "/bundle") {
			_, _ = w.Write([]byte("# Self-review context bundle for demo\n\nfound: 1 of 1 sources; skipped: 0 of 1 sources\n\n## change summary\n\nwhat changed\n"))
			return
		}
		_, _ = w.Write([]byte(`{"change":"demo","dispatches":[],"findings":[]}`))
	}))
	defer srv.Close()
	// The guard shells out to `flow record findings` with no -addr, and that
	// subprocess inherits this process's environment. Its reads resolve
	// FLOW_RECORDS_ADDR before FLOW_ADDR, so pinning both to srv.URL keeps
	// the guard reading from this test's daemon rather than from the
	// caller's FLOW_RECORDS_ADDR, FLOW_ADDR or the dev daemon on 4173.
	t.Setenv("FLOW_ADDR", srv.URL)
	t.Setenv("FLOW_RECORDS_ADDR", srv.URL)

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

// TestRecordRenderWritesTheChangeKeyedNameBesideNoLegacy pins the
// filename rule (kan-399): the render writes `<change>.md`, never a
// date-stamped name, and a legacy dated file left by an older binary is
// abandoned rather than adopted -- the fresh render does not overwrite it
// and does not mint a second dated copy.
func TestRecordRenderWritesTheChangeKeyedNameBesideNoLegacy(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	dir := filepath.Join(repo, ".superpowers", "sdd", "ledgers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir ledgers: %v", err)
	}
	legacy := filepath.Join(dir, "2020-01-01-demo.md")
	if err := os.WriteFile(legacy, []byte("dated render\n"), 0o644); err != nil {
		t.Fatalf("write legacy render: %v", err)
	}

	srv := httptest.NewServer(renderDaemon(t, map[string]records.Rendered{
		"ledger": mustEnvelope(t, "ledger", renderRun),
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "ledger", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	// Resolved, like every Destination expectation: on macOS t.TempDir()'s
	// /var differs from its real /private/var, and the destination carries
	// the resolved form.
	resolvedWant, err := filepath.EvalSymlinks(filepath.Join(dir, "demo.md"))
	if err != nil {
		t.Fatalf("resolve want: %v", err)
	}
	if !strings.Contains(stdout.String(), "rendered: "+resolvedWant) {
		t.Errorf("stdout = %q, want it to report rendered: %s", stdout.String(), resolvedWant)
	}
	body, err := os.ReadFile(resolvedWant)
	if err != nil {
		t.Fatalf("read change-keyed ledger: %v", err)
	}
	if !strings.Contains(string(body), "unknown (agent-defined)") {
		t.Errorf("the change-keyed file does not hold the fresh render:\n%s", body)
	}
	legacyBody, err := os.ReadFile(legacy)
	if err != nil {
		t.Fatalf("read legacy ledger: %v", err)
	}
	if string(legacyBody) != "dated render\n" {
		t.Errorf("the legacy dated file was modified:\n%s", legacyBody)
	}
}

// TestRecordRenderStopsLoudlyOnADaemonWithoutTheRoute pins the one
// non-journalled read failure: the render route answering 404 can only
// mean a daemon too old to carry it (a current daemon answers an unknown
// change from the empty run and a missing ledger as the envelope), and a
// local-render fallback there would put the record's content back under
// this binary's vintage -- the silent staleness this verb exists to
// remove. So the skew exits non-zero, names the fix, and writes nothing.
func TestRecordRenderStopsLoudlyOnADaemonWithoutTheRoute(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`404 page not found`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "all", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero against a daemon without the render route; stdout:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "render route") {
		t.Errorf("stderr = %q, want it to name the render route the daemon lacks", stderr.String())
	}
	if strings.Contains(stdout.String(), "rendered:") || strings.Contains(stdout.String(), "MISSING:") || strings.Contains(stdout.String(), "journalled:") {
		t.Errorf("stdout = %q, want no outcome line at all -- a stop writes nothing and reports nothing rendered", stdout.String())
	}
	ledgers, err := filepath.Glob(filepath.Join(repo, ".superpowers", "sdd", "ledgers", "*"))
	if err != nil {
		t.Fatalf("glob ledgers: %v", err)
	}
	if len(ledgers) != 0 {
		t.Errorf("the skewed stop wrote %v; nothing must be written", ledgers)
	}
	reviews, err := filepath.Glob(filepath.Join(repo, ".superpowers", "sdd", "reviews", "*"))
	if err != nil {
		t.Fatalf("glob reviews: %v", err)
	}
	if len(reviews) != 0 {
		t.Errorf("the skewed stop wrote %v; nothing must be written", reviews)
	}
}

// TestRecordRenderJournalsEveryKindWhenOneKindFails pins the mixed
// outcome the panel caught: a daemon that answers the ledger and dies on
// the panel must leave NOTHING on disk and report `journalled:` for every
// kind -- a half-answer that wrote the earlier kind's file under a stderr
// line claiming nothing was rendered is the contradiction this verb's
// outcome table cannot carry. Every kind is fetched before any file is
// written, so the fetch that failed takes the whole run to the
// journalled path with the earlier kind's body still in memory only.
func TestRecordRenderJournalsEveryKindWhenOneKindFails(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/render/panel") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(mustEnvelope(t, "ledger", renderRun))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"store is down"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "render", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-kind", "all", "-repo", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 -- an unreachable store journals, it does not fail; stderr:\n%s", code, stderr.String())
	}
	for _, kind := range []string{"ledger", "panel"} {
		if !strings.Contains(stdout.String(), "journalled: "+kind+"\n") {
			t.Errorf("stdout = %q, want journalled: %s among the outcome lines", stdout.String(), kind)
		}
	}
	if strings.Contains(stdout.String(), "rendered:") || strings.Contains(stdout.String(), "MISSING:") {
		t.Errorf("stdout = %q, want no rendered:/MISSING: line -- nothing was written", stdout.String())
	}
	written, err := filepath.Glob(filepath.Join(repo, ".superpowers", "sdd", "*", "*"))
	if err != nil {
		t.Fatalf("glob .superpowers/sdd: %v", err)
	}
	if len(written) != 0 {
		t.Errorf("the half-answer wrote %v; stderr's nothing-rendered line must stay true", written)
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
	repo := recordRepo(t, "kan-258")
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
	repo := recordRepo(t, "kan-258")
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
		[]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-task", "6", "-role", "implementer", "-model", "opus",
			"-key", "task-6-implementer", "-session-token", "mf-record-pair"},
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
			"-commit", "abc1234000000000000000000000000000000000", "-outcome", "completed"},
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
		"commitSha":    "abc1234000000000000000000000000000000000",
		"outcome":      "completed",
		"endedAt":      "0001-01-01T00:00:00Z",
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
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)

	var beginOut, beginErr bytes.Buffer
	if code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-effort", "default", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
			"-agent-id", "none", "-change", "kan-258", "-task", "6", "-role", "implementer", "-model", "opus",
			"-key", "task-6-implementer", "-session-token", "mf-record-end-404"},
		strings.NewReader(""), &beginOut, &beginErr); code != 0 {
		t.Fatalf("journalled begin exit = %d; stderr:\n%s", code, beginErr.String())
	}

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"store: dispatch not found"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-key", "task-6-implementer", "-session-token", "mf-record-end-404",
			"-outcome", "completed"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 -- a dispatch end whose begin has not landed yet must never block; stderr:\n%s", code, stderr.String())
	}
	entries, exists := recordJournalEntries(t, repo, "kan-258")
	if !exists {
		t.Fatal("no record journal was written -- the end would be lost and its window left open forever")
	}
	if len(entries) != 2 {
		t.Fatalf("len(record journal entries) = %d, want 2 (the begin, then the end)", len(entries))
	}
	var body struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(entries[1].Body, &body); err != nil {
		t.Fatalf("decode journalled body: %v", err)
	}
	if body.Kind != "dispatch-end" {
		t.Errorf("journalled kind = %q, want dispatch-end", body.Kind)
	}
}

// TestRecordDispatchEndRefusesA404WithNoQueuedBegin: a 404 from a reachable
// store with no journalled begin for the key and session token is a caller
// mistake (kan-929: an end sent with another run's token warned "store
// unreachable" and journalled an entry that 404s on every replay). It is
// reported, names the status, journals nothing and exits non-zero.
func TestRecordDispatchEndRefusesA404WithNoQueuedBegin(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"store: dispatch not found"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-key", "task-6-implementer", "-session-token", "mf-other-run",
			"-outcome", "completed"},
		strings.NewReader(""), &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero for a refused end; stderr:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "404") || strings.Contains(stderr.String(), "store unreachable") {
		t.Errorf("stderr = %q, want the 404 named and no unreachable warning", stderr.String())
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("a refused end was journalled -- it would 404 on every replay")
	}
}

// TestRecordFindingStillRefusesAnUnknownRef is the contrast case for the
// exception above: a 404 from any other record write stays definitive, so
// the exception cannot quietly become the rule.
func TestRecordFindingStillRefusesAnUnknownRef(t *testing.T) {
	repo := recordRepo(t, "kan-258")
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
	repo := recordRepo(t, "kan-258")
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
			repo := recordRepo(t, "kan-258")
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
	srv := httptest.NewServer(runRecordDaemon(t, body))
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
	repo := recordRepo(t, "kan-258")
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
			repo := recordRepo(t, "kan-258")
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
			repo := recordRepo(t, "kan-258")
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
	repo := recordRepo(t, "kan-258")
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
			repo := recordRepo(t, "kan-258")
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
	repo := recordRepo(t, "demo")
	isolatedStateRoot(t)

	body := `{"change":"demo","dispatches":[],"findings":[
	  {"ref":"F1","round":0,"slot":"Bugbot","severity":"Minor","location":"a.go:1","note":"n1","status":"fixed","reproducer":"none — prose only"},
	  {"ref":"F2","round":0,"slot":"Security","severity":"Major","location":"b.go:2","note":"n2","status":"open","reproducer":"scripts/x.sh"}
	]}`
	srv := httptest.NewServer(runRecordDaemon(t, body))
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
	srv := httptest.NewServer(runRecordDaemon(t, body))
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
	srv := httptest.NewServer(runRecordDaemon(t, body))
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
	repo := recordRepo(t, "demo")
	isolatedStateRoot(t)

	body := `{"change":"demo","findings":[{"ref":"F1","round":0,"slot":"Bugbot","severity":"Minor","location":"a.go:1","note":"n1","status":"fixed","reproducer":"none — prose only"}],"dispatches":null}`
	srv := httptest.NewServer(runRecordDaemon(t, body))
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

// --- record roles: serves the dispatch-role vocabulary ---

// TestRunRecordRolesPrintsServedVocabulary pins `flow record roles` to the
// served source: stdout is exactly recordRoles, one role per line, with a
// clean exit and no store contact. It is the dispatch-role vocabulary's
// `flow stage keys` (stage_test.go): the answer a caller about to type
// -role reads instead of guessing the set and paying a refused call per
// guess (KAN-595), and the output any consumer of the vocabulary reads
// rather than keeping a transcription of its own.
func TestRunRecordRolesPrintsServedVocabulary(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "roles"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty on a clean success", stderr.String())
	}
	want := ""
	for _, role := range recordRoles {
		want += role + "\n"
	}
	if want == "" {
		t.Fatal("recordRoles is empty -- the served vocabulary vanished")
	}
	if stdout.String() != want {
		t.Errorf("stdout = %q, want the served roles one per line %q", stdout.String(), want)
	}
}

// TestRunRecordRolesRejectsPositionalArgument pins the exit-2 contract the
// sibling `flow stage keys` holds: an unexpected positional argument is a
// usage error reported on stderr, never silently ignored input.
func TestRunRecordRolesRejectsPositionalArgument(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "roles", "unexpected-arg"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (usage error)", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty on a usage error", stdout.String())
	}
	if want := "flow: record roles takes no positional arguments"; !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr = %q, want it to carry %q", stderr.String(), want)
	}
}

// TestSkillDocRolesAreAllServed pins the skills' documented `-role` call
// sites to the served set, one direction deliberately: every role literal a
// run can copy out of skills/ must be one `flow record dispatch` accepts,
// so the documentation a run authors its dispatch calls from can never name
// a role the CLI refuses (KAN-595's trial-and-error cycle). The scan is
// whole-content, not line-based: a call site wrapped across lines -- the
// form skills/flow/review-panel.md's pass-log paragraphs already use --
// must reach the pin exactly as a same-line one does (F1, KAN-595's panel).
// The reverse holds nowhere in the tree: planner, conductor and red-partner
// are accepted but have no `-role <literal>` call site in living docs --
// the recordRoles comment is their documentation. `*-rationale.md` files
// are excluded because no run ever loads them (skills/flow-contracts/SKILL.md),
// so a rationale's prose is not vocabulary a caller can act on.
func TestSkillDocRolesAreAllServed(t *testing.T) {
	root := filepath.Join("..", "..", "..", "skills")
	var hits []string
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") || strings.HasSuffix(path, "-rationale.md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range roleLiteralRe.FindAllStringSubmatchIndex(string(data), -1) {
			role := string(data[m[2]:m[3]])
			if !slices.Contains(recordRoles, role) {
				hits = append(hits, fmt.Sprintf("%s:%d: -role %s is not a served dispatch role",
					path, 1+bytes.Count(data[:m[0]], []byte("\n")), role))
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("scanning %s: %v", root, walkErr)
	}
	for _, hit := range hits {
		t.Error(hit)
	}
}

// roleLiteralRe matches a literal `-role <word>` token: "-role", whitespace
// -- newlines included, so a call site wrapped after the flag still matches
// -- then the lowercase words that would follow it in a dispatch call. The
// prefix class is what keeps prose hyphenations from matching -- "one-role
// alike" and "one-role dispatch" carry a letter glued to the hyphen, and a
// flag token never does.
var roleLiteralRe = regexp.MustCompile(`(?:^|[^A-Za-z])-role\s+([a-z][a-z-]*)`)

// TestRunRecordDispatchBeginRejectsUnknownEffort pins that -effort is
// checked against the four accepted words before the store is ever
// contacted, exactly as -role is in
// TestRecordRejectsUnknownRoleWithoutContactingStore: a caller mistake
// taking the never-block fallback path would journal a write a replay
// could only ever be refused for a second time.
func TestRunRecordDispatchBeginRejectsUnknownEffort(t *testing.T) {
	repo := recordRepo(t, "kan-258")
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
			"-key", "task-1-implementer", "-session-token", "mf-record-unknown-effort"},
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

// TestRunRecordDispatchBeginUnknownEffortNamesDefault pins that the
// refusal of an effort the session cannot state names the `default` rule
// (model-policy.md, **Harness mapping**): kan-988's inline rows recorded
// an invented `high` after `-effort unknown` was refused without it.
func TestRunRecordDispatchBeginUnknownEffortNamesDefault(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", "http://127.0.0.1:1", "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-effort", "unknown",
			"-key", "task-1-implementer", "-session-token", "mf-record-unknown-effort"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if want := "an effort the session does not state is recorded as default, never guessed"; !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr does not name the default rule %q:\n%s", want, stderr.String())
	}
}

// TestRecordDispatchBeginSendsAnExplicitDefaultEffort pins that an explicit
// -effort default is sent as the literal "default" -- the dispatcher set no
// effort, a known fact; an omitted -effort is refused instead
// (TestRecordDispatchBeginRequiresEffort).
func TestRecordDispatchBeginSendsAnExplicitDefaultEffort(t *testing.T) {
	sent := dispatchBeginBody(t, "-effort", "default")

	if sent["effort"] != "default" {
		t.Errorf("effort = %v, want %q", sent["effort"], "default")
	}
}

// TestRecordDispatchBeginAcceptsEffort pins -effort through to the request
// body verbatim for each of the five accepted non-default words.
func TestRecordDispatchBeginAcceptsEffort(t *testing.T) {
	for _, effort := range []string{"low", "medium", "high", "xhigh", "max"} {
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
	repo := recordRepo(t, "kan-258")
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
	if err := os.WriteFile(filepath.Join(repo, "render.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

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
			repo := recordRepo(t, "kan-258")
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

// TestRecordPassAndMutationMissingFlagsRefuseWithoutChange pins KAN-590:
// `flow record pass` and `flow record mutation` refuse a dropped -change
// before the store is ever contacted -- non-zero exit, nothing recorded, no
// journal -- so a dropped flag is a loud error rather than a mis-associated
// pass-log row discovered later. The contacted stub is the load-bearing
// half: were the identity flags ever swapped for the conn-only set, the
// refusal would vanish and the call would reach the store (or, on a quiet
// machine, journal), and either way this test must name the mechanism.
func TestRecordPassAndMutationMissingFlagsRefuseWithoutChange(t *testing.T) {
	for _, tc := range []struct {
		name string
		verb string
		args []string
	}{
		{"pass without change", "pass", []string{"-note", "guard rc=1 (scripts/test-setup.sh)"}},
		{"mutation without change", "mutation", []string{"-path", "render.go", "-mutated", "none", "-test", "guard was equivalent"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)

			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":1}`))
			}))
			defer srv.Close()

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				append([]string{"record", tc.verb, "-addr", srv.URL, "-timeout", "500ms", "-C", repo}, tc.args...),
				strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for a record with no -change -- it must be refused first")
			}
			if !strings.Contains(stderr.String(), "-change is required") {
				t.Errorf("stderr does not name the missing flag:\n%s", stderr.String())
			}
			// The refused call names no change, so a write that somehow
			// fell through to the never-block path would journal under the
			// empty change -- assert that file's absence, not a named
			// change's, or the assertion guards nothing.
			if _, exists := recordJournalEntries(t, repo, ""); exists {
				t.Errorf("a record with no -change left a journal behind")
			}
		})
	}
}

// TestRecordFindingLineageFlags pins that `flow record finding`'s lineage
// flags reach the recorded finding's wire shape: -supersedes and
// -regression-of name earlier findings' refs, and the daemon receives them
// under the JSON keys the store reads. Omitted flags record a finding
// carrying no lineage, which is the ordinary shape of most findings.
func TestRecordFindingLineageFlags(t *testing.T) {
	cases := []struct {
		name         string
		args         []string
		supersedes   string
		regressionOf string
	}{
		{
			name:       "supersedes only",
			args:       []string{"-supersedes", "F1"},
			supersedes: "F1",
		},
		{
			name:         "regression-of only",
			args:         []string{"-regression-of", "F3"},
			regressionOf: "F3",
		},
		{
			name:         "both links",
			args:         []string{"-supersedes", "F1", "-regression-of", "F3"},
			supersedes:   "F1",
			regressionOf: "F3",
		},
		{
			name: "neither flag",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := recordRepo(t, "kan-507")
			isolatedStateRoot(t)

			var body records.Finding
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&body)
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"ref":"F5","round":1,"slot":"principles","severity":"major","note":"n","status":"open"}`))
			}))
			defer srv.Close()

			args := []string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-507", "-ref", "F5", "-round", "1", "-slot", "principles",
				"-severity", "major", "-status", "open", "-reproducer", "scripts/x.sh",
				"-note", "the note"}
			args = append(args, c.args...)

			var stdout, stderr bytes.Buffer
			code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
			}
			if body.Supersedes != c.supersedes {
				t.Errorf("recorded supersedes = %q, want %q", body.Supersedes, c.supersedes)
			}
			if body.RegressionOf != c.regressionOf {
				t.Errorf("recorded regressionOf = %q, want %q", body.RegressionOf, c.regressionOf)
			}
		})
	}
}

// TestRecordFindingTaskRoundTrips pins -task end to end through the CLI:
// a finding recorded with -task 2 is sent under the wire key the store
// reads, and `flow record findings` prints it back. The fake daemon keeps
// the posted finding and serves it on the run-record read.
func TestRecordFindingTaskRoundTrips(t *testing.T) {
	repo := recordRepo(t, "kan-9")
	isolatedStateRoot(t)

	var stored json.RawMessage
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = fmt.Fprintf(w, `{"change":"kan-9","dispatches":[],"findings":[%s]}`, stored)
			return
		}
		stored, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(stored)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-9", "-ref", "F1", "-slot", "principles", "-severity", "major",
			"-status", "open", "-reproducer", "scripts/x.sh", "-note", "the note", "-task", "2"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("record finding exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	stdout.Reset()
	code = run(context.Background(),
		[]string{"record", "findings", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-9"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("record findings exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	var got []records.Finding
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(got) != 1 || got[0].TaskID != "2" {
		t.Errorf("findings = %+v, want one finding with task 2", got)
	}
}

// TestValidateFindingCategory pins -category's two rules, judged before
// the store is ever contacted: the word must come from the closed
// vocabulary the deferred-Minor breakdown counts on, and it is legal only
// beside a `deferred <reason>` status -- a category on any other status is
// the same caller contradiction the store refuses one layer out.
func TestValidateFindingCategory(t *testing.T) {
	for _, category := range []string{"doc-only", "pre-existing", "cosmetic", "coverage-gap", "out-of-scope", "other"} {
		if err := validateFindingCategory(category, "deferred cosmetic, not worth a fix round"); err != nil {
			t.Errorf("validateFindingCategory(%q, deferred ...) = %v, want nil", category, err)
		}
	}
	for _, status := range []string{"open", "fixed", "withdrawn superseded by F2"} {
		if err := validateFindingCategory("cosmetic", status); err == nil {
			t.Errorf("validateFindingCategory(cosmetic, %q) = nil, want an error -- a category is deferred-only", status)
		}
	}
	if err := validateFindingCategory("nobody-heard-of-this", "deferred whatever"); err == nil {
		t.Error(`validateFindingCategory("nobody-heard-of-this", deferred) = nil, want an error -- the set is closed`)
	}
	if err := validateFindingCategory("", "fixed"); err != nil {
		t.Errorf("validateFindingCategory(\"\", fixed) = %v, want nil -- the empty category is the ordinary category-less write", err)
	}
}

// TestRecordStatusJournalsCategory pins the journal half of the category
// (F8, review panel round 0): the journalled status request is the whole
// call, category included, so a replay of a categorised deferral re-lands
// the category instead of clearing it.
func TestRecordStatusJournalsCategory(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "status", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
			"-change", "kan-258", "-ref", "F2",
			"-status", "deferred doc wording only", "-category", "doc-only"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a dead store must never block); stderr:\n%s", code, stderr.String())
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
	if body.Kind != "status" {
		t.Fatalf("journalled kind = %q, want status", body.Kind)
	}
	var req recordStatusRequest
	if err := json.Unmarshal(body.Request, &req); err != nil {
		t.Fatalf("decode journalled status request: %v", err)
	}
	if req.Category != "doc-only" {
		t.Errorf("journalled category = %q, want doc-only -- a replay without it would clear the deferral's word", req.Category)
	}
}

// TestRecordFindingSendsCategory pins the finding verb's wire half (F9,
// review panel round 0): -category reaches the daemon's POST body, so a
// finding recorded already-deferred lands with its category in one write.
func TestRecordFindingSendsCategory(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		var err error
		gotBody, err = readAll(r)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ref":"F3"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-ref", "F3", "-slot", "principles", "-severity", "Minor",
			"-status", "deferred cosmetic, not worth a fix round", "-category", "cosmetic",
			"-reproducer", "none — cosmetic dead code", "-note", "the handler carries dead code"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var sent records.Finding
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode POST body %s: %v", gotBody, err)
	}
	if sent.Category != "cosmetic" {
		t.Errorf("POST body category = %q, want cosmetic", sent.Category)
	}
}

// TestRecordVerbsValidateCategoryBeforeTheStore pins the invocation half
// of the validator (F10, review panel round 0): a category on a
// non-deferred status, and an unrecognised word, are refused with exit 2
// before any network call -- a dead addr proves the refusal, since a fall
// through to the write path would journal and exit 0.
func TestRecordVerbsValidateCategoryBeforeTheStore(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)

	cases := []struct {
		name string
		args []string
	}{
		{
			name: "status with a category on fixed",
			args: []string{"record", "status", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-258", "-ref", "F1", "-status", "fixed", "-category", "cosmetic"},
		},
		{
			name: "status with an unknown word",
			args: []string{"record", "status", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-258", "-ref", "F1", "-status", "deferred whatever", "-category", "nobody-heard-of-this"},
		},
		{
			name: "finding with a category on open",
			args: []string{"record", "finding", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-258", "-ref", "F4", "-slot", "principles", "-severity", "Minor",
				"-status", "open", "-category", "doc-only",
				"-reproducer", "none — wording only", "-note", "a wording finding"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), tc.args, strings.NewReader(""), &stdout, &stderr)
			if code != 2 {
				t.Fatalf("exit code = %d, want 2 (the validator must refuse before the store is contacted); stderr:\n%s", code, stderr.String())
			}
			entries, _ := recordJournalEntries(t, repo, "kan-258")
			if len(entries) != 0 {
				t.Fatalf("a refused write was journalled anyway: %d entries", len(entries))
			}
		})
	}
}

// TestRecordFindingRefusesADeferredNonMinorBeforeTheStore pins the raise
// verb's half of the Minor-only rule (KAN-508's F5, review panel round 0):
// the CLI can see both -severity and -status in the same invocation, so a
// deferral beside any severity but Minor is refused with exit 2 before any
// network call -- a dead addr proves the refusal, since a fall through to
// the write path would journal and exit 0 -- the same pre-store shape
// validateFindingCategory gives the category rule. A Minor (either case)
// passes the validator and falls through to the network path, journalling
// as every unreachable write does.
func TestRecordFindingRefusesADeferredNonMinorBeforeTheStore(t *testing.T) {
	repo := recordRepo(t, "kan-508", "kan-508-ok")
	isolatedStateRoot(t)

	for _, tc := range []struct{ name, severity string }{
		{name: "Critical", severity: "Critical"},
		{name: "Important", severity: "Important"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), []string{"record", "finding", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-508", "-ref", "F1", "-slot", "principles", "-severity", tc.severity,
				"-status", "deferred cosmetic dead code", "-category", "cosmetic",
				"-reproducer", "none — wording only", "-note", "a wording finding"},
				strings.NewReader(""), &stdout, &stderr)
			if code != 2 {
				t.Fatalf("exit code = %d, want 2 (the validator must refuse before the store is contacted); stderr:\n%s", code, stderr.String())
			}
			entries, _ := recordJournalEntries(t, repo, "kan-508")
			if len(entries) != 0 {
				t.Fatalf("a refused write was journalled anyway: %d entries", len(entries))
			}
		})
	}
	for _, tc := range []struct{ name, severity string }{
		{name: "Minor passes through", severity: "Minor"},
		{name: "lowercase minor passes through", severity: "minor"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), []string{"record", "finding", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-508-ok", "-ref", "F1", "-slot", "principles", "-severity", tc.severity,
				"-status", "deferred cosmetic dead code", "-category", "cosmetic",
				"-reproducer", "none — wording only", "-note", "a wording finding"},
				strings.NewReader(""), &stdout, &stderr)
			if code == 2 {
				t.Fatalf("exit code = 2, want the validator to pass the deferral through to the network path; stderr:\n%s", stderr.String())
			}
		})
	}
}

// TestRecordSubstitutionCommand pins `flow record substitution`'s three
// outcomes: a live store records the row with a CLI-stamped recordedAt, a
// dead store falls back to the journal under kind "substitution", and a
// -shape outside the closed vocabulary is a caller mistake refused before
// the store is ever contacted.
func TestRecordSubstitutionCommand(t *testing.T) {
	t.Run("writes the substitution", func(t *testing.T) {
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
			_, _ = w.Write([]byte(`{"id":1,"guard":"gather-dispatch-context","shape":"cross-repo","substitution":"composed the dispatch bundle by hand","recordedAt":"2026-01-02T03:04:05Z"}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "substitution", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-417", "-guard", "gather-dispatch-context", "-shape", "cross-repo",
				"-substitution", "composed the dispatch bundle by hand"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if !strings.HasSuffix(gotPath, "/kan-417/substitutions") {
			t.Errorf("request path = %s, want it to end in /kan-417/substitutions", gotPath)
		}

		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		if sent["guard"] != "gather-dispatch-context" {
			t.Errorf("guard = %v, want gather-dispatch-context", sent["guard"])
		}
		if sent["shape"] != "cross-repo" {
			t.Errorf("shape = %v, want cross-repo", sent["shape"])
		}
		if sent["substitution"] != "composed the dispatch bundle by hand" {
			t.Errorf("substitution = %v, want the verbatim substitution", sent["substitution"])
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
			[]string{"record", "substitution", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-417", "-guard", "check-task-commit-fields", "-shape", "cross-repo",
				"-substitution", "checked every commit message by hand"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0 (a dead store must never block); stderr:\n%s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "store unreachable") {
			t.Errorf("stderr = %q, want it to name the store as unreachable", stderr.String())
		}

		entries, exists := recordJournalEntries(t, repo, "kan-417")
		if !exists || len(entries) != 1 {
			t.Fatalf("record journal entries = %d (exists=%v), want exactly 1", len(entries), exists)
		}
		var body struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(entries[0].Body, &body); err != nil {
			t.Fatalf("decode journalled body: %v", err)
		}
		if body.Kind != "substitution" {
			t.Errorf("journalled kind = %q, want substitution", body.Kind)
		}
	})

	t.Run("refuses a shape outside the closed set", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "substitution", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-change", "kan-417", "-guard", "gather-dispatch-context", "-shape", "all",
				"-substitution", "s"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 2 {
			t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "-shape") {
			t.Errorf("stderr = %q, want it to name -shape as the offender", stderr.String())
		}
		if _, exists := recordJournalEntries(t, repo, "kan-417"); exists {
			t.Errorf("a refused caller mistake was journalled")
		}
	})
}

// TestRecordSubstitutionsReadback pins `flow record substitutions`' read
// contract: the array verbatim, newest first as the daemon answers it, []
// for no rows, and a loud non-zero failure on an unreachable store.
func TestRecordSubstitutionsReadback(t *testing.T) {
	t.Run("prints the array verbatim", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		body := `[{"id":1,"change":"kan-417","guard":"gather-dispatch-context","shape":"cross-repo","substitution":"composed by hand","recordedAt":"2026-01-02T03:04:05Z"}]`
		var gotQuery string
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(body))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "substitutions", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-guard", "gather-dispatch-context", "-shape", "cross-repo"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if got := strings.TrimRight(stdout.String(), "\n"); got != body {
			t.Errorf("stdout = %q, want the array verbatim %q", got, body)
		}
		if !strings.Contains(gotQuery, "guard=gather-dispatch-context") {
			t.Errorf("query = %q, want guard=gather-dispatch-context", gotQuery)
		}
		if !strings.Contains(gotQuery, "shape=cross-repo") {
			t.Errorf("query = %q, want shape=cross-repo", gotQuery)
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
			[]string{"record", "substitutions", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
			strings.NewReader(""), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		if got := strings.TrimRight(stdout.String(), "\n"); got != "[]" {
			t.Fatalf("stdout = %q, want exactly []", got)
		}
	})

	t.Run("refuses a shape outside the closed set", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "substitutions", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
				"-shape", "all"},
			strings.NewReader(""), &stdout, &stderr)

		if code != 2 {
			t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "-shape") {
			t.Errorf("stderr = %q, want it to name -shape as the offender", stderr.String())
		}
	})

	t.Run("dead port fails loudly", func(t *testing.T) {
		repo := gitRepo(t)
		isolatedStateRoot(t)

		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "substitutions", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo},
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

// TestRecordFindingPatternFlag pins the write path's new flag: `-pattern`
// rides on the finding POST's JSON body and a well-formed value reaches the
// store, while the pattern's normalization stays the store's business --
// the CLI forwards verbatim what the caller typed.
func TestRecordFindingPatternFlag(t *testing.T) {
	repo := recordRepo(t, "kan-416")
	isolatedStateRoot(t)

	var gotPattern string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("decode finding body: %v", err)
		}
		gotPattern, _ = body["pattern"].(string)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ref":"F1","round":0,"slot":"principles","severity":"major","note":"n","status":"open","pattern":"restyled-row-loses-its-handler"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-416", "-ref", "F1", "-round", "0", "-slot", "principles",
			"-severity", "major", "-status", "open", "-reproducer", "scripts/x.sh",
			"-note", "the restyled row lost its handler",
			"-pattern", "Restyled Row Loses its Handler!"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if gotPattern != "Restyled Row Loses its Handler!" {
		t.Errorf("body pattern = %q, want the caller's spelling forwarded verbatim -- the store normalizes, not the CLI", gotPattern)
	}
}

// TestRecordFindingPatternsVerb pins the registry summary read: a project's
// patterns as a JSON array on stdout, exit 0, with no -change flag at all --
// the verdicts-and-incidents project scope, since a pattern's recurrence
// spans changes.
func TestRecordFindingPatternsVerb(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/api/v1/finding-patterns/") {
			t.Errorf("request = %s %s, want GET /api/v1/finding-patterns/<project>", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"pattern":"restyled-row-loses-its-handler","occurrences":3,"changes":2,
			"firstSeen":"2026-09-01T09:00:00Z","lastSeen":"2026-09-12T09:00:00Z"}]`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding-patterns", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	var got []struct {
		Pattern     string `json:"pattern"`
		Occurrences int    `json:"occurrences"`
		Changes     int    `json:"changes"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout as JSON array: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(got) != 1 || got[0].Pattern != "restyled-row-loses-its-handler" || got[0].Occurrences != 3 || got[0].Changes != 2 {
		t.Errorf("patterns = %+v, want one restyled-row row at 3 occurrences in 2 changes", got)
	}
}

// TestRecordFindingPatternVerb pins the detail read: `-name` required (exit
// 2 without it, the caller-mistake contract), the pattern escaped into the
// path, and a pattern nothing carries printed as exactly "[]" with exit 0.
func TestRecordFindingPatternVerb(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var requested string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		requested = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"change":"kan-1","ref":"F1","severity":"major","status":"open",
			"note":"the restyled row lost its handler","recordedAt":"2026-09-01T09:00:00Z"}]`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding-pattern", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-name", "Restyled Row"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.HasPrefix(requested, "/api/v1/finding-patterns/") || !strings.HasSuffix(requested, "/Restyled%20Row") {
		t.Errorf("requested %q, want the project-scoped registry path ending in the escaped pattern", requested)
	}
	var got []struct {
		Change string `json:"change"`
		Ref    string `json:"ref"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout as JSON array: %v\nstdout:\n%s", err, stdout.String())
	}
	if len(got) != 1 || got[0].Change != "kan-1" || got[0].Ref != "F1" {
		t.Errorf("occurrences = %+v, want one kan-1/F1 row", got)
	}

	// The caller mistake: no -name, usage on stderr, exit 2.
	stdout.Reset()
	stderr.Reset()
	code = run(context.Background(),
		[]string{"record", "finding-pattern", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Errorf("missing -name exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "-name is required") {
		t.Errorf("stderr = %q, want it to name -name as required", stderr.String())
	}
}

// TestRecordFindingPatternsEmptyRegistryPrintsEmptyArray pins the guard the
// daemon's null body rides on (KAN-416 panel F5): the store answers an
// unlabeled project with a JSON null, and the CLI must print exactly []
// regardless -- a caller scripting against this verb must never see null
// where the documented contract says an empty array.
func TestRecordFindingPatternsEmptyRegistryPrintsEmptyArray(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`null`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding-patterns", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := strings.TrimRight(stdout.String(), "\n"); got != "[]" {
		t.Fatalf("stdout = %q, want exactly []", got)
	}
}

// TestRecordFindingPatternEmptyDetailPrintsEmptyArray pins the detail
// verb's nil-guard (KAN-416 panel round-1 F10): the daemon answers a
// pattern nothing carries with a JSON null, and this verb — like its
// plural sibling — must print exactly [].
func TestRecordFindingPatternEmptyDetailPrintsEmptyArray(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`null`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding-pattern", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-name", "pattern-nobody-labeled"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := strings.TrimRight(stdout.String(), "\n"); got != "[]" {
		t.Fatalf("stdout = %q, want exactly []", got)
	}
}

// --- write-time validation: an inline dispatch's reported tokens (KAN-525) ---

// TestDispatchEndTokensParserRefusals pins the closed grammar -tokens
// accepts: comma-separated key=value pairs over {input, output,
// cache-read, cache-creation}, at least one pair, no duplicate key, every
// value a non-negative integer. Each refusal below is a caller mistake,
// refused with exit 2 before the store is ever contacted -- the
// recordEfforts/recordCauses precedent -- and before anything is
// journalled, since a replay of it could only ever be refused a second
// time.
func TestDispatchEndTokensParserRefusals(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	cases := []struct{ name, value string }{
		{"no equals sign", "input"},
		{"unknown key", "bandwidth=5"},
		{"duplicate key", "input=1,input=2"},
		{"negative", "input=-3"},
		{"not a number", "input=many"},
		{"missing value", "input="},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, _, stderr, contacted, repo := causeEndRun(t, "-tokens", tc.value)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for an unparseable -tokens value -- it must be refused first")
			}
			if _, exists := recordJournalEntries(t, repo, "kan-510"); exists {
				t.Error("a refused -tokens value wrote a record journal")
			}
			if !strings.Contains(stderr.String(), "-tokens") {
				t.Errorf("stderr does not name -tokens:\n%s", stderr.String())
			}
		})
	}
}

// TestDispatchEndTokensCarriedOnWire pins the happy path: a well-formed
// -tokens value rides the end call's body as the `tokens` object, its
// hyphenated CLI keys stored under the bag's own snake_case names -- and
// an end that omits -tokens carries no `tokens` key at all, an absent key
// rather than an empty value, the same contract agentId's omitted case
// pins.
func TestDispatchEndTokensCarriedOnWire(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	send := func(t *testing.T, extra ...string) map[string]any {
		t.Helper()
		repo := recordRepo(t, "kan-525")
		isolatedStateRoot(t)

		var gotBody []byte
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			var err error
			gotBody, err = readAll(r)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"sonnet","endedAt":"2026-01-02T03:44:05Z"}`))
		}))
		defer srv.Close()

		var stdout, stderr bytes.Buffer
		args := []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-525", "-key", "inline-implementer", "-session-token", "ff-kan525-tokens"}
		args = append(args, extra...)
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}

		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		return sent
	}

	t.Run("reported", func(t *testing.T) {
		sent := send(t, "-tokens", "input=1234,output=567,cache-read=89,cache-creation=12")
		tokens, ok := sent["tokens"].(map[string]any)
		if !ok {
			t.Fatalf("tokens = %v, want an object", sent["tokens"])
		}
		want := map[string]float64{"input": 1234, "output": 567, "cache_read": 89, "cache_creation": 12}
		if len(tokens) != len(want) {
			t.Fatalf("tokens = %v, want exactly the four reported keys", tokens)
		}
		for k, v := range want {
			if tokens[k] != v {
				t.Errorf("tokens[%q] = %v, want %v", k, tokens[k], v)
			}
		}
	})

	t.Run("omitted", func(t *testing.T) {
		sent := send(t)
		if v, ok := sent["tokens"]; ok {
			t.Errorf("tokens = %v, want the key absent -- an end that reports no usage must not write an empty report", v)
		}
	})
}

func TestRecordFlagsDefaultToTheRecordsAddress(t *testing.T) {
	t.Setenv("FLOW_ADDR", "http://127.0.0.1:4174")
	t.Setenv("FLOW_RECORDS_ADDR", "http://127.0.0.1:9999")

	var f recordIdentityFlags
	fset := flag.NewFlagSet("record", flag.ContinueOnError)
	registerRecordConnFlags(fset, &f)

	if f.addr != "http://127.0.0.1:9999" {
		t.Fatalf("record -addr default = %q, want the records address %q", f.addr, "http://127.0.0.1:9999")
	}
}

// TestRunRecordSummary pins `flow record summary`'s write shape: the
// change's summary text goes to the store verbatim from -file (or stdin),
// the daemon's row id comes back as "recorded: summary <id>", a missing
// -change and an empty body are caller mistakes refused with exit 2
// before the store is contacted, and the request path ends in
// /<change>/summary.
func TestRunRecordSummary(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"id":7,"summary":"what changed and why","recordedAt":"2026-09-19T12:00:00Z"}`))
	}))
	defer srv.Close()

	file := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(file, []byte("what changed and why"), 0o644); err != nil {
		t.Fatalf("write summary file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "summary", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-598", "-file", file},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty on a clean success", stderr.String())
	}
	if !strings.Contains(stdout.String(), "recorded: summary 7") {
		t.Errorf("stdout = %q, want it to name the recorded row id 7", stdout.String())
	}
	if !strings.HasSuffix(gotPath, "/kan-598/summary") {
		t.Errorf("request path = %s, want it to end in /kan-598/summary", gotPath)
	}
	var sent map[string]any
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode request body %s: %v", gotBody, err)
	}
	if sent["summary"] != "what changed and why" {
		t.Errorf("request summary = %v, want the file's text verbatim", sent["summary"])
	}

	// Caller mistakes, both refused before the store is ever contacted:
	// a missing -change, and an empty body.
	emptyFile := filepath.Join(t.TempDir(), "empty.md")
	if err := os.WriteFile(emptyFile, []byte("   \n"), 0o644); err != nil {
		t.Fatalf("write empty summary file: %v", err)
	}
	contacted := false
	srv2 := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv2.Close()

	for name, args := range map[string][]string{
		"missing change": {"record", "summary", "-addr", srv2.URL, "-timeout", "500ms", "-C", repo, "-file", file},
		"empty body":     {"record", "summary", "-addr", srv2.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-598", "-file", emptyFile},
	} {
		var stdout2, stderr2 bytes.Buffer
		if code := run(context.Background(), args, strings.NewReader(""), &stdout2, &stderr2); code != 2 {
			t.Errorf("%s: exit code = %d, want 2; stderr:\n%s", name, code, stderr2.String())
		}
	}
	if contacted {
		t.Error("the store was contacted for a caller mistake -- it must be refused first")
	}
}

// TestRunRecordSummaryJournalsWhenStoreDown pins the never-block fallback
// every other write verb carries: an unreachable store journals the
// summary under kind "summary" and still exits 0.
func TestRunRecordSummaryJournalsWhenStoreDown(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	file := filepath.Join(t.TempDir(), "summary.md")
	if err := os.WriteFile(file, []byte("what changed and why"), 0o644); err != nil {
		t.Fatalf("write summary file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "summary", "-addr", deadPortAddr(t), "-timeout", "300ms", "-C", repo,
			"-change", "kan-598", "-file", file},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (a dead store must never block); stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "store unreachable") {
		t.Errorf("stderr = %q, want it to name the store as unreachable", stderr.String())
	}

	entries, exists := recordJournalEntries(t, repo, "kan-598")
	if !exists || len(entries) != 1 {
		t.Fatalf("record journal entries = %d (exists=%v), want exactly 1", len(entries), exists)
	}
	var got struct {
		Kind    string `json:"kind"`
		Request struct {
			Summary string `json:"summary"`
		} `json:"request"`
	}
	if err := json.Unmarshal(entries[0].Body, &got); err != nil {
		t.Fatalf("decode journalled body: %v", err)
	}
	if got.Kind != "summary" {
		t.Errorf("journalled kind = %q, want summary", got.Kind)
	}
	if got.Request.Summary != "what changed and why" {
		t.Errorf("journalled summary = %q, want the file's text verbatim", got.Request.Summary)
	}
}

// --- handoff-lines ------------------------------------------------------

// handoffDaemon serves the one read `record handoff-lines` makes: the cost
// status. Any other request fails the test.
func handoffDaemon(t *testing.T) http.HandlerFunc {
	t.Helper()
	return genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("handoff-lines sent a %s request; it only reads", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/cost-status") {
			t.Errorf("handoff-lines requested %s; it reads cost-status only", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"unattributed":0,"reasons":{}}`))
	})
}

// writeRecordJournal writes n complete entries into the record journal of
// dir's project for change, the file journal-count counts.
func writeRecordJournal(t *testing.T, dir, change string, n int) string {
	t.Helper()
	key, _, err := fallback.ProjectKey(dir)
	if err != nil {
		t.Fatalf("project key: %v", err)
	}
	path := recordJournalPath(key, change)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("{\"kind\":\"finding\"}\n", n)), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func runHandoffLines(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), append([]string{"record", "handoff-lines", "-timeout", "500ms"}, args...),
		strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// TestRecordHandoffLinesFormat pins the block the verb prints: the three
// Records spellings and the Costs line, nothing else.
func TestRecordHandoffLinesFormat(t *testing.T) {
	cases := []struct {
		name    string
		journal int // -1: make the journal unreadable
		want    string
	}{
		{"clean", 0, "**Records:** all writes reached the store\n**Costs:** 0 unattributed\n"},
		{"journalled", 2, "**Records:** 2 write(s) journalled — the store was unreachable\n**Costs:** 0 unattributed\n"},
		{"uncountable", -1, "**Records:** unknown — the journal could not be counted\n**Costs:** 0 unattributed\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)
			if tc.journal > 0 {
				writeRecordJournal(t, repo, "demo", tc.journal)
			}
			if tc.journal < 0 {
				path := writeRecordJournal(t, repo, "demo", 0)
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			srv := httptest.NewServer(handoffDaemon(t))
			defer srv.Close()

			code, out, errOut := runHandoffLines(t, "-addr", srv.URL, "-C", repo, "-change", "demo")
			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, errOut)
			}
			if out != tc.want {
				t.Errorf("stdout =\n%s\nwant\n%s", out, tc.want)
			}
		})
	}
}

// TestRecordHandoffLinesDedupByProject pins the aggregation: the journal is
// keyed per project, so two worktrees of one project are counted once, never
// double-counted.
func TestRecordHandoffLinesDedupByProject(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	wt := filepath.Join(t.TempDir(), "wt")
	if out, err := exec.Command("git", "-C", repo, "worktree", "add", "-q", "--detach", wt).CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}
	writeRecordJournal(t, repo, "demo", 3)

	srv := httptest.NewServer(handoffDaemon(t))
	defer srv.Close()

	code, out, errOut := runHandoffLines(t, "-addr", srv.URL, "-C", repo, "-change", "demo", "-worktree", wt)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, errOut)
	}
	want := "**Records:** 3 write(s) journalled — the store was unreachable\n**Costs:** 0 unattributed\n"
	if out != want {
		t.Errorf("stdout =\n%s\nwant\n%s", out, want)
	}
}

// TestRecordHandoffLinesReadsNoFindings pins that the verb makes no findings
// read: a records endpoint that fails is never hit, and both lines print.
func TestRecordHandoffLinesReadsNoFindings(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	recordReads := 0
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/cost-status") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"unattributed":0,"reasons":{}}`))
			return
		}
		recordReads++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	code, out, errOut := runHandoffLines(t, "-addr", srv.URL, "-C", repo, "-change", "demo")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, errOut)
	}
	want := "**Records:** all writes reached the store\n**Costs:** 0 unattributed\n"
	if out != want {
		t.Errorf("stdout =\n%s\nwant\n%s", out, want)
	}
	if recordReads != 0 {
		t.Errorf("records endpoint hit %d time(s), want 0 -- handoff-lines reads no findings", recordReads)
	}
}

// TestRunRecordNextRound pins the round a panel run opens with: one past
// the highest round any pass, finding or mutation row carries, 0 for a
// change the store has never heard of, and no number on a failed read.
func TestRunRecordNextRound(t *testing.T) {
	repo := recordRepo(t, "demo")
	isolatedStateRoot(t)

	for _, tc := range []struct {
		name, body string
		status     int
		want       string
	}{
		{"highest row's round plus one", `{"change":"demo","dispatches":[],
		  "findings":[{"ref":"F1","round":2,"slot":"primary","severity":"Minor","note":"n","status":"fixed"}],
		  "passes":[{"id":1,"round":4,"note":"a"},{"id":2,"round":1,"note":"b"}],
		  "mutations":[{"id":1,"round":3,"path":"p","mutated":"m","test":"t"}]}`, http.StatusOK, "5"},
		{"no rows", ``, http.StatusNotFound, "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "next-round", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
				strings.NewReader(""), &stdout, &stderr)
			if code != 0 || strings.TrimSpace(stdout.String()) != tc.want {
				t.Fatalf("exit %d stdout %q, want 0 and %s; stderr:\n%s", code, stdout.String(), tc.want, stderr.String())
			}
		})
	}

	t.Run("unreachable store prints no number", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "next-round", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo, "-change", "demo"},
			strings.NewReader(""), &stdout, &stderr)
		if code == 0 || stdout.Len() != 0 {
			t.Fatalf("exit %d stdout %q, want non-zero and empty", code, stdout.String())
		}
	})
}

// TestRunRecordPasses pins `flow record passes`: the change's pass rows as
// one JSON array, `[]` for a change the store has never heard of, and
// nothing on stdout with a non-zero exit when the store cannot be read --
// check-unfinished-work's signal five reads it.
func TestRunRecordPasses(t *testing.T) {
	repo := recordRepo(t, "demo")
	isolatedStateRoot(t)
	for _, tc := range []struct {
		name, body string
		status     int
		want       string
	}{
		{"pass rows", `{"change":"demo","dispatches":[],"passes":[{"id":1,"round":0,"note":"roster: full"}]}`, http.StatusOK, `[{"id":1,"round":0,"note":"roster: full"}]`},
		{"no pass rows", `{"change":"demo","dispatches":[]}`, http.StatusOK, `[]`},
		{"unknown change", ``, http.StatusNotFound, `[]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "passes", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
				strings.NewReader(""), &stdout, &stderr)
			if code != 0 || strings.TrimSpace(stdout.String()) != tc.want {
				t.Fatalf("exit %d stdout %q, want 0 and %s; stderr:\n%s", code, stdout.String(), tc.want, stderr.String())
			}
		})
	}
	t.Run("unreachable store prints nothing", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "passes", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo, "-change", "demo"},
			strings.NewReader(""), &stdout, &stderr)
		if code == 0 || stdout.Len() != 0 {
			t.Fatalf("exit %d stdout %q, want non-zero and empty", code, stdout.String())
		}
	})
}

// TestRunRecordNextRef pins the ref a run records its first finding under:
// one past the highest F<n> any finding row carries, compared as numbers,
// F1 for a change the store has never heard of, and nothing on a failed
// read -- kan-951's third run numbered from F1 and overwrote the creating
// run's F1.
func TestRunRecordNextRef(t *testing.T) {
	repo := recordRepo(t, "demo")
	isolatedStateRoot(t)

	for _, tc := range []struct {
		name, body string
		status     int
		want       string
	}{
		{"highest ref plus one", `{"change":"demo","dispatches":[],
		  "findings":[{"ref":"F9","round":0,"slot":"primary","severity":"Minor","note":"n","status":"fixed"},
		              {"ref":"F10","round":3,"slot":"primary","severity":"Minor","note":"n","status":"open"}]}`, http.StatusOK, "F11"},
		{"no rows", ``, http.StatusNotFound, "F1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "next-ref", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
				strings.NewReader(""), &stdout, &stderr)
			if code != 0 || strings.TrimSpace(stdout.String()) != tc.want {
				t.Fatalf("exit %d stdout %q, want 0 and %s; stderr:\n%s", code, stdout.String(), tc.want, stderr.String())
			}
		})
	}

	t.Run("unreachable store prints nothing", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "next-ref", "-addr", deadPortAddr(t), "-timeout", "500ms", "-C", repo, "-change", "demo"},
			strings.NewReader(""), &stdout, &stderr)
		if code == 0 || stdout.Len() != 0 {
			t.Fatalf("exit %d stdout %q, want non-zero and empty", code, stdout.String())
		}
	})
}

// TestRecordFindingRejectsMalformedRef pins that a -ref, -supersedes or
// -regression-of not shaped F<positive int> is refused with exit 2 before
// the store is contacted -- kan-964 stored two rows keyed by commit shas,
// a shape only `flow self-review finding` takes.
func TestRecordFindingRejectsMalformedRef(t *testing.T) {
	for _, tc := range []struct{ flag, value string }{
		{"-ref", "d1650b7"},
		{"-ref", "F0"},
		{"-ref", "F01"},
		{"-ref", "f3"},
		{"-ref", "F-1"},
		{"-supersedes", "1e728c4"},
		{"-regression-of", "KAN-9"},
	} {
		t.Run(tc.flag+"="+tc.value, func(t *testing.T) {
			repo := recordRepo(t, "kan-964")
			isolatedStateRoot(t)

			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			args := []string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-964", "-ref", "F2", "-slot", "principles",
				"-severity", "major", "-status", "open", "-reproducer", "go test ./...",
				"-note", "the note"}
			if tc.flag == "-ref" {
				args[11] = tc.value
			} else {
				args = append(args, tc.flag, tc.value)
			}
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if contacted {
				t.Error("the store was contacted for a malformed ref -- it must be refused first")
			}
			if !strings.Contains(stderr.String(), tc.flag+" ") || !strings.Contains(stderr.String(), "F<n>") {
				t.Errorf("stderr does not name %s and the F<n> shape:\n%s", tc.flag, stderr.String())
			}
		})
	}
}

// TestRecordPassRefusesMarkerNote pins the kan-980 self-review fix: a pass
// note opening with a marker label is a marker line recorded in the wrong
// place -- the renderer neutralises its colon, so the round loses the
// structural row it meant to carry -- and `flow record pass` refuses it
// before the store is contacted, naming the command that records it. The
// -addr is a closed port, so a regressed refusal falls through to the
// journal instead of reaching a live daemon.
func TestRecordPassRefusesMarkerNote(t *testing.T) {
	for _, note := range []string{
		"fix-mutation: stats/x.go — dropped the guard — TestX",
		"  finding-status: F9 fixed",
	} {
		t.Run(note, func(t *testing.T) {
			repo := recordRepo(t, "kan-258")
			isolatedStateRoot(t)

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "pass", "-addr", "http://127.0.0.1:1", "-timeout", "500ms", "-C", repo,
					"-change", "kan-258", "-round", "4", "-note", note},
				strings.NewReader(""), &stdout, &stderr)
			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "flow record mutation") || !strings.Contains(stderr.String(), "flow record finding") {
				t.Errorf("stderr = %q, want it to name flow record mutation and flow record finding", stderr.String())
			}
			if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
				t.Errorf("a refused marker note left a record journal behind")
			}
		})
	}
}

// TestRecordPassRefusesOneWordNote pins the kan-991 self-review fix: a
// one-word pass note is a placeholder -- kan-991's parent looped `-note "x"`
// and landed two stray rows -- and `flow record pass` refuses it before the
// store is contacted. The -addr is a closed port, so a regressed refusal
// falls through to the journal instead of reaching a live daemon.
func TestRecordPassRefusesOneWordNote(t *testing.T) {
	for _, note := range []string{"x", "  probe  "} {
		t.Run(note, func(t *testing.T) {
			repo := recordRepo(t, "kan-258")
			isolatedStateRoot(t)

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "pass", "-addr", "http://127.0.0.1:1", "-timeout", "500ms", "-C", repo,
					"-change", "kan-258", "-round", "0", "-note", note},
				strings.NewReader(""), &stdout, &stderr)
			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "is one word") {
				t.Errorf("stderr = %q, want it to say the note is one word", stderr.String())
			}
			if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
				t.Errorf("a refused one-word note left a record journal behind")
			}
		})
	}
}

// TestRecordWriteRefusesAChangeNoCheckoutCarries pins the project-key
// sanity check: a run-record write for a change no checkout of the -C
// repository carries is refused before the store is contacted, so it never
// files the row under the wrong project; a change carried by a linked
// worktree, an archive or a /flow-fast .superpowers/sdd root is accepted
// from the main checkout.
func TestRecordWriteRefusesAChangeNoCheckoutCarries(t *testing.T) {
	for _, c := range []struct {
		name, root string
		linked     bool
		want       int
	}{
		{"no checkout carries the change", "", false, 2},
		{"a linked worktree carries it", "spectre/changes/kan-7", true, 0},
		{"the archive carries it", "spectre/changes/archive/kan-7", false, 0},
		{"a flow-fast root carries it", ".superpowers/sdd/kan-7", true, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)
			holder := repo
			if c.linked {
				holder = filepath.Join(t.TempDir(), "wt")
				if out, err := exec.Command("git", "-C", repo, "worktree", "add", "-q", "-b", "kan-7", holder).CombinedOutput(); err != nil {
					t.Fatalf("git worktree add: %v\n%s", err, out)
				}
			}
			if c.root != "" {
				if err := os.MkdirAll(filepath.Join(holder, c.root), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"verifier","model":"sonnet","endedAt":"2026-01-02T03:44:05Z"}`))
			}))
			defer srv.Close()
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-7", "-key", "k", "-session-token", "mf-kan7-end", "-outcome", "completed"}, strings.NewReader(""), &stdout, &stderr)
			if code != c.want || contacted != (c.want == 0) {
				t.Fatalf("exit %d, contacted %v; want exit %d, contacted %v\nstderr:\n%s", code, contacted, c.want, c.want == 0, stderr.String())
			}
			if c.want != 0 && !strings.Contains(stderr.String(), `carries change "kan-7"`) {
				t.Errorf("stderr does not name the change:\n%s", stderr.String())
			}
		})
	}
}

// recordRepo is gitRepo carrying spectre/changes/<change> for every change
// named: a run-record write refuses a change no checkout of its repository
// carries (requireChangeCheckout).
func recordRepo(t *testing.T, changes ...string) string {
	t.Helper()
	repo := gitRepo(t)
	for _, c := range changes {
		if err := os.MkdirAll(filepath.Join(repo, "spectre", "changes", c), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return repo
}

// A name that resolves to a directory without being a change -- the archive
// root, a climb out of the tree -- is refused, not stat-matched.
func TestRequireChangeCheckoutRefusesNonChangeNames(t *testing.T) {
	repo := recordRepo(t, "real")
	if err := os.MkdirAll(filepath.Join(repo, "spectre", "changes", "archive"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{"archive", "..", "../changes", "real/..", "."} {
		if requireChangeCheckout(repo, c, io.Discard) {
			t.Errorf("change %q accepted", c)
		}
	}
	if !requireChangeCheckout(repo, "real", io.Discard) {
		t.Error("real change refused")
	}
}

// TestRecordDispatchBeginModelAlias pins that a full Claude model id is
// recorded under its alias -- kan-939's parent rows read claude-opus-5-5
// beside dispatches' opus -- and any other model verbatim.
func TestRecordDispatchBeginModelAlias(t *testing.T) {
	for given, want := range map[string]string{
		"claude-opus-5-5":   "opus",
		"claude-sonnet-4-6": "sonnet",
		"opus":              "opus",
		"glm-5.3-flash":     "glm-5.3-flash",
	} {
		if sent := dispatchBeginBody(t, "-model", given); sent["model"] != want {
			t.Errorf("-model %s recorded as %v, want %s", given, sent["model"], want)
		}
	}
}

// TestRecordDispatchBackfill pins the kan-939 self-review fix: a back-fill
// is one `dispatch backfill` call whose real instants are required. It
// sends begin carrying -started-at, then end carrying -ended-at and the
// outcome; a call missing either instant, or carrying one that is not RFC
// 3339, is exit 2 before the store is called -- never half a row.
func TestRecordDispatchBackfill(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	const start, end = "2026-10-15T12:08:50Z", "2026-10-15T12:31:07Z"
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	var bodies []map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		raw, _ := readAll(r)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		bodies = append(bodies, b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus"}`))
	}))
	defer srv.Close()
	base := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-task", "7", "-role", "implementer", "-model", "opus", "-agent-id", "none",
		"-key", "task-7-implementer", "-session-token", "mf-record-backfill", "-commit", "abc1234000000000000000000000000000000000"}
	var stdout, stderr bytes.Buffer
	args := append(append([]string{}, base...), "-started-at", start, "-ended-at", end, "-outcome", "completed")
	if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if len(bodies) != 2 {
		t.Fatalf("store calls = %d, want begin then end: %v", len(bodies), bodies)
	}
	if bodies[0]["startedAt"] != start || bodies[0]["taskId"] != "7" || bodies[0]["endedAt"] != nil {
		t.Errorf("begin body = %v, want startedAt %s, taskId 7, no endedAt", bodies[0], start)
	}
	if bodies[1]["endedAt"] != end || bodies[1]["outcome"] != "completed" || bodies[1]["commitSha"] != "abc1234000000000000000000000000000000000" {
		t.Errorf("end body = %v, want endedAt %s, outcome completed, commitSha abc1234000000000000000000000000000000000", bodies[1], end)
	}

	for name, extra := range map[string][]string{
		"no -ended-at":                {"-started-at", start, "-outcome", "completed"},
		"no -started-at":              {"-ended-at", end, "-outcome", "completed"},
		"no -outcome":                 {"-started-at", start, "-ended-at", end},
		"a malformed -ended-at":       {"-started-at", start, "-ended-at", "yesterday", "-outcome", "completed"},
		"-started-at after -ended-at": {"-started-at", end, "-ended-at", start, "-outcome", "completed"},
		"an unknown -outcome":         {"-started-at", start, "-ended-at", end, "-outcome", "done"},
		"-cause with completed":       {"-started-at", start, "-ended-at", end, "-outcome", "completed", "-cause", "other"},
		"malformed -tokens":           {"-started-at", start, "-ended-at", end, "-outcome", "completed", "-tokens", "bogus"},
	} {
		bodies = nil
		stderr.Reset()
		if code := run(context.Background(), append(append([]string{}, base...), extra...), strings.NewReader(""), &stdout, &stderr); code != 2 || bodies != nil {
			t.Errorf("%s: exit %d, store calls %d; want 2 and none (stderr %q)", name, code, len(bodies), stderr.String())
		}
	}
}

// TestRecordDispatchBackfillInstants pins the kan-982 self-review fix: a
// back-filled row carries the dispatch's real instants. -started-at reaches
// begin's body as startedAt and -ended-at end's as endedAt; an instant that
// is not RFC 3339 is a usage refusal, exit 2, before the store is called.
func TestRecordDispatchBackfillInstants(t *testing.T) {
	const start, end = "2026-10-15T12:08:50Z", "2026-10-15T12:31:07Z"
	if sent := dispatchBeginBody(t, "-started-at", start); sent["startedAt"] != start {
		t.Errorf("begin startedAt = %v, want %s", sent["startedAt"], start)
	}

	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = readAll(r)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"reviewer","model":"sonnet","endedAt":"` + end + `"}`))
	}))
	defer srv.Close()
	endArgs := []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-key", "panel-1-primary", "-session-token", "mf-record-backfill"}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), append(endArgs, "-ended-at", end), strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("end exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	var sent map[string]any
	if err := json.Unmarshal(gotBody, &sent); err != nil || sent["endedAt"] != end {
		t.Errorf("end endedAt = %v (%v), want %s", sent["endedAt"], err, end)
	}

	for _, args := range [][]string{
		{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "reviewer", "-model", "sonnet", "-key", "k", "-session-token", "mf-record-backfill",
			"-started-at", "15 Oct 12:08"},
		append(append([]string{}, endArgs...), "-ended-at", "yesterday"),
	} {
		stdout.Reset()
		stderr.Reset()
		gotBody = nil
		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "is not an RFC 3339 instant") || gotBody != nil {
			t.Errorf("%s: exit %d, store called %v, stderr %q; want 2, no call, an RFC 3339 refusal", args[2], code, gotBody != nil, stderr.String())
		}
	}
}

// fixtureCommit makes an empty commit in repo with its committer (and
// author) time pinned to at, and returns its sha.
func fixtureCommit(t *testing.T, repo, at string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", repo, "commit", "--allow-empty", "-q", "-m", "fixture "+at)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_AUTHOR_DATE="+at, "GIT_COMMITTER_DATE="+at)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	out, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

// postBodies serves a genuine daemon that records every POST body and
// answers anything else 404.
func postBodies(t *testing.T) (*httptest.Server, *[]map[string]any) {
	t.Helper()
	var bodies []map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		raw, _ := readAll(r)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		bodies = append(bodies, b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &bodies
}

// TestRecordDispatchEndSendsCommitAt pins kan-1003's CLI half: an end's
// -commit carries its committer time as commitAt, so the store can refuse
// an inline row begun after it; a sha no worktree knows sends none.
func TestRecordDispatchEndSendsCommitAt(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	sha := fixtureCommit(t, repo, "2026-10-21T11:10:31Z")
	srv, bodies := postBodies(t)
	base := []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-key", "task-1-implementer", "-session-token", "mf-record-commitat", "-outcome", "completed"}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), append(append([]string{}, base...), "-commit", sha), strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr.String())
	}
	if len(*bodies) != 1 || (*bodies)[0]["commitAt"] != "2026-10-21T11:10:31Z" {
		t.Fatalf("bodies = %v, want one with commitAt 2026-10-21T11:10:31Z", *bodies)
	}
	*bodies = nil
	if code := run(context.Background(), append(append([]string{}, base...), "-commit", "abc1234000000000000000000000000000000000"), strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("unknown sha: exit %d; stderr:\n%s", code, stderr.String())
	}
	if len(*bodies) != 1 {
		t.Fatalf("unknown sha: bodies = %v, want one", *bodies)
	}
	if _, has := (*bodies)[0]["commitAt"]; has {
		t.Errorf("unknown sha sent commitAt: %v", (*bodies)[0])
	}
}

// TestRecordDispatchEndAcceptsSkipped pins `skipped` in recordOutcomes:
// run-verify-list ends an all-skipped run with it, and a refusal would leave
// that row open.
func TestRecordDispatchEndAcceptsSkipped(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	srv, bodies := postBodies(t)
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-key", "verify-list-lint", "-session-token", "mf-record-skipped", "-outcome", "skipped"},
		strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr.String())
	}
	if len(*bodies) != 1 || (*bodies)[0]["outcome"] != "skipped" {
		t.Fatalf("bodies = %v, want one with outcome skipped", *bodies)
	}
}

// TestRecordDispatchBackfillDerivesInlineInstants pins kan-1003's derived
// back-fill: an inline back-fill with -commit takes its end from that
// commit and its start from the commit before it, and refuses either
// instant given by hand.
func TestRecordDispatchBackfillDerivesInlineInstants(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	fixtureCommit(t, repo, "2026-10-21T11:02:00Z")
	sha := fixtureCommit(t, repo, "2026-10-21T11:10:31Z")
	srv, bodies := postBodies(t)
	base := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-task", "1", "-role", "implementer", "-model", "opus",
		"-key", "task-1-implementer", "-session-token", "mf-record-derive", "-outcome", "completed"}
	inline := append(append([]string{}, base...), "-agent-id", "inline", "-commit", sha)
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), inline, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d; stderr:\n%s", code, stderr.String())
	}
	if len(*bodies) != 2 || (*bodies)[0]["startedAt"] != "2026-10-21T11:02:00Z" || (*bodies)[1]["endedAt"] != "2026-10-21T11:10:31Z" {
		t.Fatalf("bodies = %v, want begin startedAt 11:02:00Z and end endedAt 11:10:31Z", *bodies)
	}
	for name, args := range map[string][]string{
		"-started-at beside an inline -commit": append(append([]string{}, inline...), "-started-at", "2026-10-21T11:00:00Z"),
		"-ended-at beside an inline -commit":   append(append([]string{}, inline...), "-ended-at", "2026-10-21T11:20:00Z"),
		"no instants with -agent-id none":      append(append([]string{}, base...), "-agent-id", "none", "-commit", sha),
	} {
		*bodies = nil
		stderr.Reset()
		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 2 || len(*bodies) != 0 {
			t.Errorf("%s: exit %d, store calls %d; want 2 and none (stderr %q)", name, code, len(*bodies), stderr.String())
		}
	}
}

// TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers pins the
// kan-1027 self-review fix: an inline panel-fix-<r> back-fill never starts
// before the latest end among round r-1's reviewer rows, however early the
// commit before its fix commit is; a round with no ended reviewer row, and
// any other role, keep the commit-derived start.
func TestRecordDispatchBackfillFloorsAPanelFixAtItsRoundsReviewers(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	fixtureCommit(t, repo, "2026-10-23T09:19:34Z")
	sha := fixtureCommit(t, repo, "2026-10-23T09:30:03Z")
	record := `{"change":"kan-258","dispatches":[` +
		`{"key":"panel-0-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-23T09:20:01Z","endedAt":"2026-10-23T09:28:18Z"},` +
		`{"key":"panel-0-tests","role":"reviewer","slot":"tests","startedAt":"2026-10-23T09:20:01Z","endedAt":"2026-10-23T09:20:39Z"},` +
		`{"key":"panel-1-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-23T09:30:32Z","endedAt":"2026-10-23T09:30:56Z"},` +
		`{"key":"panel-3-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-23T09:39:00Z","endedAt":"2026-10-23T09:40:00Z"},` +
		`{"key":"panel-0-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"},{"key":"panel-2-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"},{"key":"panel-3-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
	var bodies []map[string]any
	recordFails := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			if recordFails {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		raw, _ := readAll(r)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		bodies = append(bodies, b)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"panel-fix","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	for _, tc := range []struct {
		role, key, want string
		recordFails     bool
	}{
		{"panel-fix", "panel-fix-1", "2026-10-23T09:28:18Z", false},
		{"panel-fix", "panel-fix-1-2", "2026-10-23T09:28:18Z", false},
		{"panel-fix", "panel-fix-3", "2026-10-23T09:19:34Z", false},
		{"implementer", "task-1-implementer", "2026-10-23T09:19:34Z", false},
		// a reviewer end later than the fix commit never becomes the floor
		{"panel-fix", "panel-fix-4", "2026-10-23T09:19:34Z", false},
		// a failed record read keeps the commit-derived start
		{"panel-fix", "panel-fix-1", "2026-10-23T09:19:34Z", true},
	} {
		bodies = nil
		recordFails = tc.recordFails
		args := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", tc.role, "-model", "opus", "-key", tc.key,
			"-session-token", "mf-record-panel-fix-floor", "-outcome", "completed", "-agent-id", "inline", "-commit", sha}
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
			t.Fatalf("%s: exit %d; stderr:\n%s", tc.key, code, stderr.String())
		}
		if len(bodies) != 2 || bodies[0]["startedAt"] != tc.want {
			t.Errorf("%s: bodies = %v, want begin startedAt %s", tc.key, bodies, tc.want)
		}
	}
}

// TestRecordDispatchBackfillFloorsAFullSuiteFixAtThePanelsReviewers pins
// the kan-1039 self-review fix: an inline full-suite-fix-<n> back-fill never
// starts before the newest reviewer end not after its fix commit, so it no
// longer claims the whole panel wait; a record with no reviewer row, and a
// key whose <n> is not a positive integer, keep the commit-derived start.
func TestRecordDispatchBackfillFloorsAFullSuiteFixAtThePanelsReviewers(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	fixtureCommit(t, repo, "2026-10-23T09:19:34Z")
	sha := fixtureCommit(t, repo, "2026-10-23T09:30:03Z")
	reviewers := `{"change":"kan-258","dispatches":[` +
		`{"key":"panel-0-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-23T09:20:01Z","endedAt":"2026-10-23T09:28:18Z"},` +
		`{"key":"panel-0-tests","role":"reviewer","slot":"tests","startedAt":"2026-10-23T09:20:01Z","endedAt":"2026-10-23T09:20:39Z"},` +
		`{"key":"panel-1-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-23T09:30:32Z","endedAt":"2026-10-23T09:30:56Z"},` +
		`{"key":"panel-3-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-23T09:39:00Z","endedAt":"2026-10-23T09:40:00Z"}]}`
	record := reviewers
	var bodies []map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		raw, _ := readAll(r)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		bodies = append(bodies, b)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"panel-fix","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	for _, tc := range []struct{ key, record, want string }{
		{"full-suite-fix-1", reviewers, "2026-10-23T09:28:18Z"},
		{"full-suite-fix-1", `{"change":"kan-258","dispatches":[]}`, "2026-10-23T09:19:34Z"},
		{"full-suite-fix-x", reviewers, "2026-10-23T09:19:34Z"},
	} {
		bodies = nil
		record = tc.record
		args := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "panel-fix", "-model", "opus", "-key", tc.key,
			"-session-token", "mf-record-full-suite-fix-floor", "-outcome", "completed", "-agent-id", "inline", "-commit", sha}
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
			t.Fatalf("%s: exit %d; stderr:\n%s", tc.key, code, stderr.String())
		}
		if len(bodies) != 2 || bodies[0]["startedAt"] != tc.want {
			t.Errorf("%s: bodies = %v, want begin startedAt %s", tc.key, bodies, tc.want)
		}
	}
}

// TestRecordDispatchBeginRequiresEffort pins that an omitted -effort is
// refused before the store is contacted, never recorded as "default" --
// an explicit -effort default is still accepted.
func TestRecordDispatchBeginRequiresEffort(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
	}))
	defer srv.Close()

	args := []string{"record", "dispatch", "begin", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-role", "implementer", "-model", "opus",
		"-key", "task-1-implementer", "-session-token", "mf-record-requires-effort"}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "-effort") {
		t.Errorf("stderr does not name -effort:\n%s", stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for a begin with no -effort -- it must be refused first")
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Error("a begin with no -effort wrote a record journal")
	}

	stdout.Reset()
	stderr.Reset()
	code = run(context.Background(), append(args, "-effort", "default"), strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("explicit -effort default exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
}

// TestRecordFindingRefusesALocationWithNoLine pins that a non-empty
// -location naming no line is refused before the store is contacted, while
// a path:line range and an omitted -location are both accepted.
func TestRecordFindingRefusesALocationWithNoLine(t *testing.T) {
	cases := []struct {
		name     string
		location []string
		want     int
	}{
		{"prose", []string{"-location", "EditorScreen delete confirm"}, 2},
		{"range", []string{"-location", "stats/cmd/flow/record.go:12-20"}, 0},
		{"omitted", nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := recordRepo(t, "kan-258")
			isolatedStateRoot(t)

			contacted := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
				contacted = true
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"ref":"F1","round":0,"slot":"principles","severity":"major","note":"n","status":"open"}`))
			}))
			defer srv.Close()

			args := append([]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-ref", "F1", "-round", "0", "-slot", "principles",
				"-severity", "major", "-status", "open", "-reproducer", "scripts/x.sh",
				"-note", "the note"}, tc.location...)
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
			if code != tc.want {
				t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, tc.want, stderr.String())
			}
			if tc.want == 2 {
				if !strings.Contains(stderr.String(), "file:line") {
					t.Errorf("stderr does not name file:line:\n%s", stderr.String())
				}
				if contacted {
					t.Error("the store was contacted for a -location naming no line")
				}
			}
		})
	}
}

// TestRecordDispatchEndRefusesTokensUnderClaudeCode pins that -tokens is
// refused before the store is contacted where CLAUDE_CODE_SESSION_ID is set
// -- the harvester credits that row from the transcript, and a report would
// overwrite it -- on both `dispatch end` and `dispatch backfill`.
func TestRecordDispatchEndRefusesTokensUnderClaudeCode(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "s")
	code, _, stderr, contacted, _ := causeEndRun(t, "-tokens", "input=1,output=2")
	if code != 2 {
		t.Fatalf("end exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	for _, want := range []string{"-tokens", "CLAUDE_CODE_SESSION_ID"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("end stderr does not name %s:\n%s", want, stderr.String())
		}
	}
	if contacted {
		t.Error("the store was contacted for -tokens under Claude Code")
	}

	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	contacted = false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus"}`))
	}))
	defer srv.Close()
	var stdout, berr bytes.Buffer
	bcode := run(context.Background(), []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-role", "implementer", "-model", "opus", "-agent-id", "none",
		"-key", "k", "-session-token", "mf-record-backfill-tokens",
		"-started-at", "2026-10-15T12:08:50Z", "-ended-at", "2026-10-15T12:31:07Z", "-outcome", "completed",
		"-tokens", "input=1,output=2"}, strings.NewReader(""), &stdout, &berr)
	if bcode != 2 || contacted {
		t.Errorf("backfill exit = %d, contacted = %v; want 2 and not contacted (stderr %q)", bcode, contacted, berr.String())
	}
	if !strings.Contains(berr.String(), "CLAUDE_CODE_SESSION_ID") {
		t.Errorf("backfill stderr does not name CLAUDE_CODE_SESSION_ID:\n%s", berr.String())
	}

	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	code, _, stderr, _, _ = causeEndRun(t, "-tokens", "input=1,output=2")
	if code != 0 {
		t.Errorf("end with CLAUDE_CODE_SESSION_ID empty exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
}

// TestRecordDispatchRefusesAnAbbreviatedCommit pins that `dispatch end` and
// `dispatch backfill` refuse an abbreviated -commit before the store is
// contacted, and that a full 40-hex sha is accepted.
func TestRecordDispatchRefusesAnAbbreviatedCommit(t *testing.T) {
	code, _, stderr, contacted, _ := causeEndRun(t, "-commit", "99153ab4")
	if code != 2 || contacted {
		t.Errorf("end exit = %d, contacted = %v; want 2 and not contacted", code, contacted)
	}
	if !strings.Contains(stderr.String(), "40") {
		t.Errorf("end stderr does not name 40:\n%s", stderr.String())
	}

	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	contacted = false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus"}`))
	}))
	defer srv.Close()
	var stdout, berr bytes.Buffer
	bcode := run(context.Background(), []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-role", "implementer", "-model", "opus", "-agent-id", "none",
		"-key", "k", "-session-token", "mf-record-backfill-commit",
		"-started-at", "2026-10-15T12:08:50Z", "-ended-at", "2026-10-15T12:31:07Z", "-outcome", "completed",
		"-commit", "99153ab4"}, strings.NewReader(""), &stdout, &berr)
	if bcode != 2 || contacted {
		t.Errorf("backfill exit = %d, contacted = %v; want 2 and not contacted (stderr %q)", bcode, contacted, berr.String())
	}
	if !strings.Contains(berr.String(), "40") {
		t.Errorf("backfill stderr does not name 40:\n%s", berr.String())
	}

	code, _, stderr, _, _ = causeEndRun(t, "-commit", "99153ab4"+strings.Repeat("0", 32))
	if code != 0 {
		t.Errorf("full sha exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
}

// dispatchBeginKey runs `flow record dispatch begin` with the given role and
// key against a fake daemon and reports the exit code, stderr, and whether
// the daemon was contacted.
func dispatchBeginKey(t *testing.T, role, key string) (int, string, bool) {
	t.Helper()
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
	}))
	defer srv.Close()
	argv := []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-role", role, "-model", "opus", "-key", key,
		"-session-token", "mf-record-key-shape"}
	if role == "reviewer" && strings.HasPrefix(key, "panel-") {
		argv = append(argv, "-slot", "primary")
	}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), argv, strings.NewReader(""), &stdout, &stderr)
	return code, stderr.String(), contacted
}

// TestRecordDispatchBeginRefusesANonNumericPanelRound pins that a reviewer's
// panel key carries its numeric round: a key PanelDispatchRound cannot read
// would drop the row out of every panel-round signal.
func TestRecordDispatchBeginRefusesANonNumericPanelRound(t *testing.T) {
	code, stderr, contacted := dispatchBeginKey(t, "reviewer", "panel-fix1-primary")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "panel-<round>-<slot>") {
		t.Errorf("stderr does not name panel-<round>-<slot>:\n%s", stderr)
	}
	if contacted {
		t.Error("a refused key was sent to the daemon")
	}
	for _, ok := range [][2]string{{"reviewer", "panel-2-primary"}, {"panel-fix", "panel-fix-2"}} {
		if code, stderr, _ := dispatchBeginKey(t, ok[0], ok[1]); code != 0 {
			t.Errorf("-role %s -key %s: exit %d, want 0; stderr:\n%s", ok[0], ok[1], code, stderr)
		}
	}
}

// TestRecordDispatchBeginRefusesAnOffConventionImplementerKey pins the
// implementer key shape task-<ids>-implementer[-fix-<k>], ids +-joined: a
// hyphen-joined task-11-12 reads as task 11 alone to every task-key reader.
func TestRecordDispatchBeginRefusesAnOffConventionImplementerKey(t *testing.T) {
	for _, bad := range []string{"task-11-12", "task-12"} {
		code, stderr, contacted := dispatchBeginKey(t, "implementer", bad)
		if code != 2 {
			t.Errorf("-key %s: exit code = %d, want 2; stderr:\n%s", bad, code, stderr)
		}
		if !strings.Contains(stderr, "task-<ids>-implementer") {
			t.Errorf("-key %s: stderr does not name task-<ids>-implementer:\n%s", bad, stderr)
		}
		if contacted {
			t.Errorf("-key %s: a refused key was sent to the daemon", bad)
		}
	}
	for _, ok := range []string{"task-12-implementer", "task-11+12-implementer", "task-3+4-implementer-fix-1",
		"task-3-implementer-retry", "task-3+4-implementer-fix-1-retry",
		"self-review-demo-fix-g1"} {
		if code, stderr, _ := dispatchBeginKey(t, "implementer", ok); code != 0 {
			t.Errorf("-key %s: exit %d, want 0; stderr:\n%s", ok, code, stderr)
		}
	}
}

// TestRecordDispatchBeginRefusesAPanelReviewerWithoutSlot pins that a
// review-panel reviewer row carries its slot (decision
// panel-reviewer-slot-required), while a non-panel reviewer key and every
// other role still record with none.
func TestRecordDispatchBeginRefusesAPanelReviewerWithoutSlot(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
	}))
	defer srv.Close()
	begin := func(role, key string) (int, string) {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-role", role, "-model", "opus", "-key", key,
				"-session-token", "mf-record-slot-required"},
			strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String()
	}
	code, stderr := begin("reviewer", "panel-1-primary")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "-slot") {
		t.Errorf("stderr does not name -slot:\n%s", stderr)
	}
	if contacted {
		t.Error("a refused panel reviewer row was sent to the daemon")
	}
	for _, ok := range [][2]string{{"reviewer", "self-review-demo-review-1"}, {"implementer", "task-1-implementer"}} {
		if code, stderr := begin(ok[0], ok[1]); code != 0 {
			t.Errorf("-role %s -key %s: exit %d, want 0; stderr:\n%s", ok[0], ok[1], code, stderr)
		}
	}
}

// TestRecordDispatchBeginRefusesAVisualVerifyVerifierWithoutDiffBase pins
// that a visual-verify verifier row carries the HEAD it verifies, while a
// plain verify verifier still records with none.
func TestRecordDispatchBeginRefusesAVisualVerifyVerifierWithoutDiffBase(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
	}))
	defer srv.Close()
	begin := func(key string, extra ...string) (int, string) {
		var stdout, stderr bytes.Buffer
		args := append([]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "verifier", "-model", "opus", "-key", key,
			"-session-token", "mf-record-vv-diff-base"}, extra...)
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String()
	}
	code, stderr := begin("visual-verify-gymie")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr)
	}
	if !strings.Contains(stderr, "-diff-base") {
		t.Errorf("stderr does not name -diff-base:\n%s", stderr)
	}
	if contacted {
		t.Error("a refused visual-verify verifier row was sent to the daemon")
	}
	if code, stderr := begin("visual-verify-gymie", "-diff-base", "0123456789abcdef0123456789abcdef01234567"); code != 0 {
		t.Errorf("with -diff-base: exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if code, stderr := begin("verify-gymie"); code != 0 {
		t.Errorf("-key verify-gymie: exit %d, want 0; stderr:\n%s", code, stderr)
	}
}

// TestRecordDispatchBeginRefusesAPanelRoundBelowTheRecorded pins decision
// panel-round-floor: a panel reviewer key opening a round below the highest
// the change's records carry is refused, while a round at or above it, and
// any begin whose record read fails, still records.
func TestRecordDispatchBeginRefusesAPanelRoundBelowTheRecorded(t *testing.T) {
	recorded := `{"change":"kan-258","passes":[{"id":1,"round":3,"note":"n"}],"dispatches":[{"key":"panel-2-primary","role":"reviewer","slot":"primary"},{"key":"panel-0-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"},{"key":"panel-2-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"},{"key":"panel-3-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
	begin := func(t *testing.T, record string, readStatus int, key string, extra ...string) (int, string, int) {
		t.Helper()
		repo := recordRepo(t, "kan-258")
		isolatedStateRoot(t)
		posts := 0
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/decisions") {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("[]"))
				return
			}
			if r.Method == http.MethodGet {
				if !strings.Contains(r.URL.Path, "/records/") || !strings.HasSuffix(r.URL.Path, "/kan-258") {
					t.Errorf("unexpected read %s", r.URL.Path)
				}
				w.WriteHeader(readStatus)
				_, _ = w.Write([]byte(record))
				return
			}
			posts++
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
		}))
		defer srv.Close()
		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			append([]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-role", "reviewer", "-model", "opus", "-key", key, "-slot", "primary",
				"-session-token", "mf-record-round-floor"}, extra...),
			strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String(), posts
	}
	t.Run("below the recorded round is refused", func(t *testing.T) {
		code, stderr, posts := begin(t, recorded, http.StatusOK, "panel-1-primary")
		if code != 2 || posts != 0 {
			t.Fatalf("exit %d, posts %d; want 2, 0; stderr:\n%s", code, posts, stderr)
		}
		if !strings.Contains(stderr, "3") || !strings.Contains(stderr, "flow record next-round") {
			t.Errorf("stderr names neither round 3 nor flow record next-round:\n%s", stderr)
		}
	})
	for _, key := range []string{"panel-3-primary", "panel-4-primary"} {
		t.Run(key+" at or above is admitted", func(t *testing.T) {
			if code, stderr, posts := begin(t, recorded, http.StatusOK, key); code != 0 || posts != 1 {
				t.Fatalf("exit %d, posts %d; want 0, 1; stderr:\n%s", code, posts, stderr)
			}
		})
	}
	t.Run("a back-filled earlier round is admitted", func(t *testing.T) {
		if code, stderr, posts := begin(t, recorded, http.StatusOK, "panel-1-primary", "-started-at", "2026-10-19T10:00:00Z"); code != 0 || posts != 1 {
			t.Fatalf("exit %d, posts %d; want 0, 1; stderr:\n%s", code, posts, stderr)
		}
	})
	t.Run("a not-found read skips the check", func(t *testing.T) {
		if code, stderr, posts := begin(t, `{"error":"not found"}`, http.StatusNotFound, "panel-0-primary"); code != 0 || posts != 1 {
			t.Fatalf("exit %d, posts %d; want 0, 1; stderr:\n%s", code, posts, stderr)
		}
	})
	t.Run("a mutation row's round counts", func(t *testing.T) {
		rec := `{"change":"kan-258","mutations":[{"id":1,"round":2,"path":"a.go"}]}`
		if code, stderr, posts := begin(t, rec, http.StatusOK, "panel-1-primary"); code != 2 || posts != 0 {
			t.Fatalf("exit %d, posts %d; want 2, 0; stderr:\n%s", code, posts, stderr)
		}
	})
}

// TestRecordMutationRefusesAPathNoWorktreeHolds pins decision
// mutation-path-exists: a mutation row names a regular file in the -C
// repository or in a worktree the change's state record names.
func TestRecordMutationRefusesAPathNoWorktreeHolds(t *testing.T) {
	write := func(root, rel string) {
		t.Helper()
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	repo := gitRepo(t)
	isolatedStateRoot(t)
	peer := t.TempDir()
	write(repo, "local/Fixed.kt")
	write(peer, "peer/Fixed.kt")
	top, err := exec.Command("git", "-C", repo, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatal(err)
	}
	root := strings.TrimSpace(string(top))

	mutate := func(t *testing.T, changeFound bool, path string) (int, string, int) {
		t.Helper()
		posts := 0
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				if !changeFound {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"not found"}`))
					return
				}
				body, _ := json.Marshal(map[string]any{"worktrees": map[string]string{peer: "abc123"}})
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(body)
				return
			}
			posts++
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1,"round":1,"path":"x","mutated":"m","test":"t"}`))
		}))
		defer srv.Close()
		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "mutation", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-round", "1", "-path", path, "-mutated", "m", "-test", "t"},
			strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String(), posts
	}

	t.Run("a path no worktree holds is refused", func(t *testing.T) {
		code, stderr, posts := mutate(t, true, "src/Missing.kt")
		if code != 2 || posts != 0 {
			t.Fatalf("exit %d, posts %d; want 2, 0; stderr:\n%s", code, posts, stderr)
		}
		for _, want := range []string{"-path", root, peer} {
			if !strings.Contains(stderr, want) {
				t.Errorf("stderr does not name %q:\n%s", want, stderr)
			}
		}
	})
	for _, p := range []string{"local/Fixed.kt", "peer/Fixed.kt", filepath.Join(root, "local/Fixed.kt")} {
		t.Run("admits "+p, func(t *testing.T) {
			if code, stderr, posts := mutate(t, true, p); code != 0 || posts != 1 {
				t.Fatalf("exit %d, posts %d; want 0, 1; stderr:\n%s", code, posts, stderr)
			}
		})
	}
	t.Run("admits a file a commit deleted", func(t *testing.T) {
		write(repo, "gone/Deleted.kt")
		for _, args := range [][]string{{"add", "gone/Deleted.kt"}, {"commit", "-q", "-m", "add"}, {"rm", "-q", "gone/Deleted.kt"}, {"commit", "-q", "-m", "delete"}} {
			cmd := exec.Command("git", append([]string{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@e"}, args...)...)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
		}
		if code, stderr, posts := mutate(t, true, "gone/Deleted.kt"); code != 0 || posts != 1 {
			t.Fatalf("exit %d, posts %d; want 0, 1; stderr:\n%s", code, posts, stderr)
		}
	})
	t.Run("a change read answering 404 names no peer", func(t *testing.T) {
		if code, stderr, _ := mutate(t, false, "peer/Fixed.kt"); code != 2 {
			t.Fatalf("peer path: exit %d, want 2; stderr:\n%s", code, stderr)
		}
		if code, stderr, _ := mutate(t, false, "local/Fixed.kt"); code != 0 {
			t.Fatalf("local path: exit %d, want 0; stderr:\n%s", code, stderr)
		}
	})
}

// TestRecordFindingFromAssignsRefs pins -from's numbering of a ref-less
// line (decision from-assigns-refs): each takes the next free ref above both
// the store's highest and every ref the file names, the refs are written
// into the file before anything is recorded so a re-run rewrites the same
// rows, and a failed store read records nothing and leaves the file as it
// was.
func TestRecordFindingFromAssignsRefs(t *testing.T) {
	repo := recordRepo(t, "kan-918")
	isolatedStateRoot(t)

	readStatus := http.StatusOK
	var gotPath string
	var posted []records.Finding
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			gotPath = r.URL.String()
			w.WriteHeader(readStatus)
			_, _ = w.Write([]byte(`{"change":"kan-918","dispatches":[],"findings":[` +
				`{"ref":"F1","round":0,"slot":"primary","severity":"Minor","note":"n","status":"fixed"},` +
				`{"ref":"F3","round":0,"slot":"primary","severity":"Minor","note":"n","status":"open"},` +
				`{"ref":"F2","round":0,"slot":"primary","severity":"Minor","note":"n","status":"open"}]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		var f records.Finding
		if err := json.Unmarshal(body, &f); err != nil {
			t.Errorf("decode posted finding: %v\n%s", err, body)
		}
		posted = append(posted, f)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	file := filepath.Join(t.TempDir(), "findings-gated-1.jsonl")
	load := func() (int, string) {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918",
				"-slot", "gated-reviewer", "-round", "0", "-status", "open", "-reproducer", "none — a gated pass", "-dispatch-seq", "4", "-from", file},
			strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String()
	}
	write := func(s string) {
		t.Helper()
		if err := os.WriteFile(file, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	refs := func() []string {
		var out []string
		for _, f := range posted {
			out = append(out, f.Ref)
		}
		return out
	}

	// (a) Two ref-less lines take F4 then F5, with the command line's own
	// slot, round and dispatch-seq and each line's own fields, and the file
	// now names them.
	write(`{"task":1,"severity":"Minor","location":"a.go:1","note":"x"}` + "\n" +
		`{"task":2,"severity":"Important","location":"b.go:2","note":"y"}` + "\n")
	if code, stderr := load(); code != 0 {
		t.Fatalf("(a) exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if got := refs(); !slices.Equal(got, []string{"F4", "F5"}) {
		t.Fatalf("(a) posted %v, want [F4 F5]", got)
	}
	for i, want := range []struct{ task, severity, location, note string }{{"1", "Minor", "a.go:1", "x"}, {"2", "Important", "b.go:2", "y"}} {
		f := posted[i]
		if f.Slot != "gated-reviewer" || f.Round != 0 || f.DispatchSeq == nil || *f.DispatchSeq != 4 ||
			f.TaskID != want.task || f.Severity != want.severity || f.Location != want.location || f.Note != want.note {
			t.Errorf("(a) posted[%d] = %+v, want slot gated-reviewer, round 0, dispatch-seq 4 and %+v", i, f, want)
		}
	}
	data, _ := os.ReadFile(file)
	if !strings.Contains(string(data), `"ref":"F4"`) || !strings.Contains(string(data), `"ref":"F5"`) {
		t.Fatalf("(a) file does not name F4 and F5:\n%s", data)
	}

	// (b) Loading the rewritten file again rewrites F4 and F5, adding no F6.
	posted = nil
	if code, stderr := load(); code != 0 {
		t.Fatalf("(b) exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if got := refs(); !slices.Equal(got, []string{"F4", "F5"}) {
		t.Fatalf("(b) posted %v, want [F4 F5] again", got)
	}

	// (c) A ref the file names above the store's highest raises the floor;
	// the line naming it keeps its bytes.
	posted = nil
	first := `{"ref":"F9","task":1,"severity":"Minor","location":"a.go:1","note":"x"}`
	write(first + "\n" + `{"task":2,"severity":"Minor","location":"b.go:2","note":"y"}` + "\n")
	if code, stderr := load(); code != 0 {
		t.Fatalf("(c) exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if got := refs(); !slices.Equal(got, []string{"F9", "F10"}) {
		t.Fatalf("(c) posted %v, want [F9 F10]", got)
	}
	data, _ = os.ReadFile(file)
	if line := strings.SplitN(string(data), "\n", 2)[0]; line != first {
		t.Errorf("(c) first line rewritten to %q, want its bytes unchanged %q", line, first)
	}

	// (d) A failed run-record read exits 1 naming the file, posts nothing,
	// and leaves the file's bytes unchanged.
	posted = nil
	readStatus = http.StatusInternalServerError
	refless := `{"task":1,"severity":"Minor","location":"a.go:1","note":"x"}` + "\n"
	write(refless)
	code, stderr := load()
	if code != 1 || !strings.Contains(stderr, file) {
		t.Errorf("(d) exit %d stderr %q, want 1 naming %s", code, stderr, file)
	}
	if len(posted) != 0 {
		t.Errorf("(d) posted %v, want nothing", refs())
	}
	if data, _ := os.ReadFile(file); string(data) != refless {
		t.Errorf("(d) file rewritten to %q, want its bytes unchanged", data)
	}

	// (e) A null line is a malformed line reported by exit, never a panic.
	readStatus = http.StatusOK
	write(`{"ref":"F1","task":1,"severity":"Minor","location":"a.go:1","note":"x"}` + "\nnull\n")
	if code, stderr := load(); code != 2 || !strings.Contains(stderr, file+":2: a line must be a JSON object") {
		t.Errorf("(e) null line exit %d stderr %q, want 2 naming line 2 as not a JSON object", code, stderr)
	}

	// (e2) A present ref that is not a string is refused naming its line,
	// never renumbered: the file keeps its bytes and nothing is posted.
	posted = nil
	numRef := `{"ref":4,"task":1,"severity":"Minor","location":"a.go:1","note":"x"}` + "\n"
	write(numRef)
	if code, stderr := load(); code != 2 || !strings.Contains(stderr, file+":1: \"ref\" must be a string") {
		t.Errorf("(e2) numeric ref exit %d stderr %q, want 2 naming line 1", code, stderr)
	}
	if data, _ := os.ReadFile(file); string(data) != numRef || len(posted) != 0 {
		t.Errorf("(e2) file %q posted %v, want bytes unchanged and nothing posted", data, refs())
	}

	// (f) -ref beside -from is refused: each line names its own ref.
	var stdout, stderrF bytes.Buffer
	code = run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918",
			"-ref", "F7", "-slot", "gated-reviewer", "-round", "0", "-status", "open", "-reproducer", "none — a gated pass", "-from", file},
		strings.NewReader(""), &stdout, &stderrF)
	if code != 2 || !strings.Contains(stderrF.String(), "-ref") {
		t.Errorf("(f) -ref with -from exit %d stderr %q, want 2 naming -ref", code, stderrF.String())
	}
	for _, form := range []string{"-ref=F7", "--ref=F7"} {
		stdout.Reset()
		stderrF.Reset()
		code = run(context.Background(),
			[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918",
				form, "-slot", "gated-reviewer", "-round", "0", "-status", "open", "-reproducer", "none — a gated pass", "-from", file},
			strings.NewReader(""), &stdout, &stderrF)
		if code != 2 || !strings.Contains(stderrF.String(), "-ref and -from cannot be combined") {
			t.Errorf("(f) %s with -from exit %d stderr %q, want 2 refusing the combination", form, code, stderrF.String())
		}
	}

	// (g) The run-record read resolves the project from the command line's
	// -C, not the working directory.
	wantKey, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatal(err)
	}
	gotPath = ""
	write(refless)
	if code, stderr := load(); code != 0 {
		t.Fatalf("(g) exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if !strings.Contains(gotPath, wantKey) {
		t.Errorf("(g) run-record read %q, want it to name -C's project %q", gotPath, wantKey)
	}

	// (h) The rewrite keeps <, > and & literal rather than \u-escaping them.
	write(`{"task":1,"severity":"Minor","location":"a.go:1","note":"a<b && c>d"}` + "\n")
	if code, stderr := load(); code != 0 {
		t.Fatalf("(h) exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if data, _ := os.ReadFile(file); !strings.Contains(string(data), `"note":"a<b && c>d"`) {
		t.Errorf("(h) rewritten file escaped the note:\n%s", data)
	}

	// (i) A rewrite that cannot write leaves the original bytes and mode,
	// and no temp file behind.
	if runtime.GOOS != "windows" && os.Getuid() != 0 {
		orig := `{"task":1,"severity":"Minor","location":"a.go:1","note":"x"}` + "\n"
		write(orig)
		if err := os.Chmod(file, 0o600); err != nil {
			t.Fatal(err)
		}
		dir := filepath.Dir(file)
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		code, _ := load()
		_ = os.Chmod(dir, 0o700)
		if code == 0 {
			t.Errorf("(i) exit 0 with an unwritable directory, want non-zero")
		}
		if data, _ := os.ReadFile(file); string(data) != orig {
			t.Errorf("(i) original changed:\n%s", data)
		}
		if entries, _ := os.ReadDir(dir); len(entries) != 1 {
			t.Errorf("(i) %d entries in the directory, want only the file", len(entries))
		}
		// A writable rewrite keeps the file's mode.
		if code, stderr := load(); code != 0 {
			t.Fatalf("(i) exit %d, want 0; stderr:\n%s", code, stderr)
		}
		if fi, err := os.Stat(file); err != nil || fi.Mode().Perm() != 0o600 {
			t.Errorf("(i) mode after rewrite = %v (%v), want 0600", fi.Mode().Perm(), err)
		}
	}
}

// TestRecordFindingFromFile pins `flow record finding -from <file>`: every
// non-blank JSON line is one finding, recorded in file order with the
// command line's own flags, a note carrying a tab, quotes and a `$` arrives
// intact, and a malformed line stops the run at its line number.
func TestRecordFindingFromFile(t *testing.T) {
	repo := recordRepo(t, "kan-918")
	isolatedStateRoot(t)

	var posted []records.Finding
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var f records.Finding
		if err := json.Unmarshal(body, &f); err != nil {
			t.Errorf("decode posted finding: %v\n%s", err, body)
		}
		posted = append(posted, f)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	file := filepath.Join(t.TempDir(), "findings.jsonl")
	lines := `{"ref":"F1","round":1,"slot":"primary","severity":"Important","location":"a.go:3","status":"open","reproducer":"scripts/x.sh","note":"tab\there \"q\" ${7:+-task $7}","dispatch-seq":4,"task":"3"}

{"ref":"F2","slot":"principles","severity":"Minor","status":"open","reproducer":"none — prose","note":"second"}
`
	if err := os.WriteFile(file, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918", "-from", file},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("record finding -from exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if len(posted) != 2 || posted[0].Ref != "F1" || posted[1].Ref != "F2" {
		t.Fatalf("posted = %+v, want F1 then F2", posted)
	}
	if got := posted[0]; got.Round != 1 || got.TaskID != "3" || got.DispatchSeq == nil || *got.DispatchSeq != 4 ||
		got.Note != "tab\there \"q\" ${7:+-task $7}" {
		t.Errorf("F1 = %+v, want round 1, task 3, dispatch-seq 4 and the note byte for byte", got)
	}
	if got := posted[1]; got.Round != 0 || got.TaskID != "" {
		t.Errorf("F2 = %+v, want round 0 and no task", got)
	}

	if err := os.WriteFile(file, []byte(`{"ref":"F3","slot":"primary","severity":"Minor","status":"open","reproducer":"none — x","note":"ok"}`+"\n{not json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stderr.Reset()
	code = run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918", "-from=" + file},
		strings.NewReader(""), &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), file+":2:") {
		t.Errorf("malformed line 2: exit %d, stderr %q; want 2 naming %s:2", code, stderr.String(), file)
	}
	if len(posted) != 3 || posted[2].Ref != "F3" {
		t.Errorf("posted = %+v, want line 1 (F3) recorded before line 2 stopped the run", posted)
	}

	// A null reads as a missing key; a non-scalar value is refused.
	if err := os.WriteFile(file, []byte(`{"ref":"F4","slot":"primary","severity":"Minor","status":"open","reproducer":"none — x","note":"n","task":null}`+"\n"+`{"ref":"F5","slot":"primary","severity":"Minor","status":"open","reproducer":"none — x","note":["a"]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stderr.Reset()
	code = run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918", "-from", file},
		strings.NewReader(""), &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), file+":2:") {
		t.Errorf("non-scalar line 2: exit %d, stderr %q; want 2 naming %s:2", code, stderr.String(), file)
	}
	if len(posted) != 4 || posted[3].Ref != "F4" || posted[3].TaskID != "" {
		t.Errorf("posted = %+v, want F4 recorded with no task and F5 refused", posted)
	}

	// A flag value that is the bare word `from` is not the -from flag.
	stderr.Reset()
	code = run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918",
			"-note", "from", "-ref", "F6", "-slot", "primary", "-severity", "Minor", "-status", "open", "-reproducer", "none — x"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 || len(posted) != 5 || posted[4].Note != "from" {
		t.Errorf("-note from: exit %d, stderr %q, posted %+v; want F6 recorded with note \"from\"", code, stderr.String(), posted)
	}
}

// TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork pins kan-920's
// self-review: task 5's inline row began after its design.md edit and
// recorded 7 seconds. An inline begin refuses a worktree carrying tracked
// modifications, naming them -- save the change's own tasks.md, whose
// close-task ticks stay uncommitted -- while a dispatched row and a
// back-filled one are not judged.
func TestRecordDispatchBeginRefusesAnInlineRowOverEarlierWork(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	change := filepath.Join(repo, "spectre", "changes", "kan-258")
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(change, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("tasks.md", "- [ ] 1. x\n")
	write("design.md", "## Context\n")
	for _, args := range [][]string{{"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "plan"}} {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	begin := func(extra ...string) (int, string, bool) {
		t.Helper()
		contacted := false
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			contacted = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
		}))
		defer srv.Close()
		argv := append([]string{"record", "dispatch", "begin", "-effort", "default", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-key", "task-5-implementer",
			"-session-token", "mf-record-inline-dirty"}, extra...)
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), argv, strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String(), contacted
	}

	write("tasks.md", "- [x] 1. x\n")
	if code, stderr, _ := begin("-agent-id", "inline"); code != 0 {
		t.Fatalf("a ticked tasks.md alone: exit %d, want 0; stderr:\n%s", code, stderr)
	}
	write("design.md", "## Context\nedited\n")
	code, stderr, contacted := begin("-agent-id", "inline")
	if code != 2 || !strings.Contains(stderr, "spectre/changes/kan-258/design.md") || strings.Contains(stderr, "tasks.md") {
		t.Fatalf("an inline begin over an edited design.md: exit %d, want 2 naming design.md alone; stderr:\n%s", code, stderr)
	}
	if contacted {
		t.Error("a refused inline begin was sent to the daemon")
	}
	if code, stderr, _ := begin("-agent-id", "none"); code != 0 {
		t.Errorf("a dispatched row over the same tree: exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if code, stderr, _ := begin("-agent-id", "inline", "-started-at", "2026-10-21T19:48:16Z"); code != 0 {
		t.Errorf("a back-filled inline row: exit %d, want 0; stderr:\n%s", code, stderr)
	}
}

// A review-panel round reads the whole diff, so every task must be closed
// first (kan-1007: panel round 4 began before task 8's close).
func TestRecordDispatchBeginRefusesAPanelRoundOverAnUntickedTask(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	plan := filepath.Join(repo, "spectre", "changes", "kan-258", "tasks.md")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(plan, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	begin := func(key string, extra ...string) (int, string, bool) {
		t.Helper()
		contacted := false
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			contacted = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
		}))
		defer srv.Close()
		argv := append([]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "reviewer", "-model", "opus", "-key", key,
			"-session-token", "mf-record-panel-unticked"}, extra...)
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), argv, strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String(), contacted
	}

	write("- [x] 1. a\n- [ ] 8. b\n")
	code, stderr, contacted := begin("panel-4-primary", "-slot", "primary")
	if code != 2 || !strings.Contains(stderr, "8") || !strings.Contains(stderr, "close-task.sh") {
		t.Fatalf("a panel begin over unticked task 8: exit %d, want 2 naming 8 and close-task.sh; stderr:\n%s", code, stderr)
	}
	if contacted {
		t.Error("a refused panel begin was sent to the daemon")
	}
	if code, stderr, _ := begin("panel-4-primary", "-slot", "primary", "-started-at", "2026-10-22T12:18:45Z"); code != 0 {
		t.Errorf("a back-filled panel row: exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if code, stderr, _ := begin("self-review-demo-review-1"); code != 0 {
		t.Errorf("a non-panel reviewer key: exit %d, want 0; stderr:\n%s", code, stderr)
	}
	write("- [x] 1. a\n- [x] 8. b\n")
	if code, stderr, _ := begin("panel-4-primary", "-slot", "primary"); code != 0 {
		t.Errorf("a panel begin over a ticked plan: exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if err := os.Remove(plan); err != nil {
		t.Fatal(err)
	}
	if code, stderr, _ := begin("panel-4-primary", "-slot", "primary"); code != 0 {
		t.Errorf("a panel begin with no tasks.md: exit %d, want 0; stderr:\n%s", code, stderr)
	}
}

func TestRecordDispatchBeginInlineReadsTheChangeWorktree(t *testing.T) {
	main := gitRepo(t)
	isolatedStateRoot(t)
	wt := filepath.Join(t.TempDir(), "wt")
	gitIn := func(dir string, args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(main, "other.txt"), "base\n")
	gitIn(main, "add", "-A")
	gitIn(main, "commit", "-q", "-m", "other")
	gitIn(main, "worktree", "add", "-q", "-b", "spectre/kan-259", wt)
	change := filepath.Join(wt, "spectre", "changes", "kan-259")
	write(filepath.Join(change, "tasks.md"), "- [ ] 1. x\n")
	write(filepath.Join(change, "design.md"), "## Context\n")
	gitIn(wt, "add", "-A")
	gitIn(wt, "commit", "-q", "-m", "plan")
	write(filepath.Join(main, "other.txt"), "another session's edit\n")
	begin := func(extra ...string) (int, string) {
		t.Helper()
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
		}))
		defer srv.Close()
		argv := append([]string{"record", "dispatch", "begin", "-effort", "default", "-addr", srv.URL, "-timeout", "500ms", "-C", main,
			"-change", "kan-259", "-role", "implementer", "-model", "opus", "-key", "task-6-implementer",
			"-session-token", "mf-record-inline-dirty"}, extra...)
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), argv, strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String()
	}

	if code, stderr := begin("-agent-id", "inline"); code != 0 {
		t.Fatalf("a dirty main checkout beside a clean change worktree: exit %d, want 0; stderr:\n%s", code, stderr)
	}
	write(filepath.Join(change, "design.md"), "## Context\nedited\n")
	code, stderr := begin("-agent-id", "inline")
	if code != 2 || !strings.Contains(stderr, "spectre/changes/kan-259/design.md") || strings.Contains(stderr, "other.txt") {
		t.Fatalf("an inline begin over an edited change worktree: exit %d, want 2 naming design.md alone; stderr:\n%s", code, stderr)
	}
}

func TestRecordDispatchBeginInlineKeepsANamedCarryingWorktree(t *testing.T) {
	main := gitRepo(t)
	isolatedStateRoot(t)
	root := t.TempDir()
	gitIn := func(dir string, args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(main, "other.txt"), "base\n")
	gitIn(main, "add", "-A")
	gitIn(main, "commit", "-q", "-m", "other")
	var wts []string
	for _, n := range []string{"a", "b"} {
		wt := filepath.Join(root, n)
		gitIn(main, "worktree", "add", "-q", "-b", "spectre/kan-259-"+n, wt)
		write(filepath.Join(wt, "spectre", "changes", "kan-259", "tasks.md"), "- [ ] 1. x\n")
		gitIn(wt, "add", "-A")
		gitIn(wt, "commit", "-q", "-m", "plan")
		wts = append(wts, wt)
	}
	// The first carrying worktree is dirty; -C names the clean second one.
	write(filepath.Join(wts[0], "other.txt"), "edit in a\n")
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
	}))
	defer srv.Close()
	argv := []string{"record", "dispatch", "begin", "-effort", "default", "-addr", srv.URL, "-timeout", "500ms", "-C", wts[1],
		"-change", "kan-259", "-role", "implementer", "-model", "opus", "-key", "task-6-implementer",
		"-session-token", "mf-record-inline-named", "-agent-id", "inline"}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), argv, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("-C a clean carrying worktree beside a dirty one: exit %d, want 0; stderr:\n%s", code, stderr.String())
	}
}

// TestDispatchEndDiffBaseCarriedOnWire pins `dispatch end -diff-base`: the
// correction reaches the daemon as diffBase, and an end without it sends
// no diffBase key, so begin's value stands (kan-1008's self-review).
func TestDispatchEndDiffBaseCarriedOnWire(t *testing.T) {
	send := func(t *testing.T, extra ...string) map[string]any {
		t.Helper()
		repo := recordRepo(t, "kan-1008")
		isolatedStateRoot(t)
		var gotBody []byte
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			var err error
			gotBody, err = readAll(r)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":7,"seq":48,"role":"verifier","model":"opus","endedAt":"2026-10-23T00:55:20Z"}`))
		}))
		defer srv.Close()
		var stdout, stderr bytes.Buffer
		args := append([]string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-1008", "-key", "visual-verify-fix-3-full-gymie", "-session-token", "mf-k1008r2", "-outcome", "completed"}, extra...)
		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
		var sent map[string]any
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode request body: %v\nbody: %s", err, gotBody)
		}
		return sent
	}
	if sent := send(t, "-diff-base", "aa758b6a,79c0541d"); sent["diffBase"] != "aa758b6a,79c0541d" {
		t.Errorf("diffBase = %v, want aa758b6a,79c0541d", sent["diffBase"])
	}
	if sent := send(t); sent["diffBase"] != nil {
		t.Errorf("diffBase = %v, want no key when -diff-base is omitted", sent["diffBase"])
	}
}

// dispatchCall runs one `flow record dispatch <verb>` against a fake daemon
// and returns its exit code, stderr and how many requests the daemon saw.
func dispatchCall(t *testing.T, verb string, extra ...string) (int, string, int) {
	t.Helper()
	repo := recordRepo(t, "kan-327")
	isolatedStateRoot(t)

	requests := 0
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"verifier","model":"opus","startedAt":"2026-01-02T03:04:05Z","endedAt":"2026-01-02T03:44:05Z"}`))
	}))
	defer srv.Close()

	args := []string{"record", "dispatch", verb, "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-327", "-session-token", "mf-record-key-shape"}
	if verb == "begin" {
		args = append(args, "-role", "verifier", "-model", "opus", "-effort", "default")
	}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), append(args, extra...), strings.NewReader(""), &stdout, &stderr)
	return code, stderr.String(), requests
}

// TestRecordDispatchBeginRefusesAKeyEndingInADash pins kan-1009's verify
// rows: a key whose suffix expanded to nothing (`verify-`) is refused before
// any request, while a key that merely ends in a letter still records.
func TestRecordDispatchBeginRefusesAKeyEndingInADash(t *testing.T) {
	for _, key := range []string{"verify-", "verify-fix-1-"} {
		code, stderr, requests := dispatchCall(t, "begin", "-agent-id", "inline", "-key", key)
		if code != 2 || !strings.Contains(stderr, fmt.Sprintf("%q", key)) || requests != 0 {
			t.Errorf("-key %s: exit %d, %d requests, stderr %q; want exit 2, 0 requests, the key named", key, code, requests, stderr)
		}
	}
	if code, stderr, _ := dispatchCall(t, "begin", "-agent-id", "inline", "-key", "verify-gymie"); code != 0 {
		t.Errorf("-key verify-gymie: exit %d, stderr %q; want 0", code, stderr)
	}
}

// TestRecordDispatchRefusesAnInventedAgentID pins kan-1009's unmeasured
// rows: a dispatch label passed as -agent-id blocks the daemon's stamp, so
// begin and end refuse anything but a harness id, inline or none.
func TestRecordDispatchRefusesAnInventedAgentID(t *testing.T) {
	for _, c := range []struct{ verb, id string }{{"begin", "task-11"}, {"end", "visual-verify"}} {
		code, stderr, requests := dispatchCall(t, c.verb, "-key", "verify-gymie", "-agent-id", c.id)
		if code != 2 || !strings.Contains(stderr, fmt.Sprintf("%q", c.id)) || !strings.Contains(stderr, "omit -agent-id") || requests != 0 {
			t.Errorf("%s -agent-id %s: exit %d, %d requests, stderr %q; want exit 2, 0 requests, the value named and omit -agent-id", c.verb, c.id, code, requests, stderr)
		}
	}
	for _, id := range []string{"a30f48a5e14b11392", "agent_541a9582-890b-417e-a3c3-419047bb4ee6", "inline", "none"} {
		if code, stderr, _ := dispatchCall(t, "begin", "-key", "verify-gymie", "-agent-id", id); code != 0 {
			t.Errorf("begin -agent-id %s: exit %d, stderr %q; want 0", id, code, stderr)
		}
	}
}

// TestRecordDispatchBeginRefusesALoopKeyOffItsLoop pins kan-1032's self-review: a pipeline-fix
// loop key begun under a change that is not its loop's is refused before the store is
// contacted -- the loop hook records it under the loop's change, so a hand-recorded row is a twin.
func TestRecordDispatchBeginRefusesALoopKeyOffItsLoop(t *testing.T) {
	cases := []struct {
		change, key string
		want        int
	}{
		{"kan-1", "pipeline-fix-7-plan", 2},
		{"pipeline-fix-kan-1-7", "pipeline-fix-7-plan", 0},
		{"kan-1", "self-review-kan-1-plan", 0},
	}
	for _, c := range cases {
		t.Run(c.change+"/"+c.key, func(t *testing.T) {
			repo := recordRepo(t, c.change)
			isolatedStateRoot(t)
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
			}))
			defer srv.Close()
			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
					"-change", c.change, "-role", "planner", "-model", "opus", "-key", c.key, "-session-token", "mf-loop-key"},
				strings.NewReader(""), &stdout, &stderr)
			if code != c.want {
				t.Fatalf("exit code = %d, want %d; stderr:\n%s", code, c.want, stderr.String())
			}
			if c.want == 2 && !strings.Contains(stderr.String(), "enforce-loop-dispatch.py") {
				t.Errorf("stderr does not name the loop hook:\n%s", stderr.String())
			}
		})
	}
}

// beginUnderDecision runs `flow record dispatch begin` with args against a
// daemon whose decisions read answers decisions (an HTTP 500 when empty)
// and whose dispatch write succeeds; it returns the exit code and stderr.
func beginUnderDecision(t *testing.T, decisions string, args ...string) (int, string) {
	t.Helper()
	repo := recordRepo(t, "kan-327")
	isolatedStateRoot(t)
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/decisions") {
			if decisions == "" {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(decisions))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus","startedAt":"2026-01-02T03:04:05Z"}`))
	}))
	defer srv.Close()
	base := []string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-327", "-session-token", "mf-record-inline"}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), append(base, args...), strings.NewReader(""), &stdout, &stderr)
	return code, stderr.String()
}

func decisionsBody(decision string) string {
	return `[{"id":1,"sessionToken":"mf-a","recordedAt":"2026-01-01T00:00:00Z","decision":{"class":"small"}},` +
		`{"id":2,"sessionToken":"mf-b","recordedAt":"2026-01-02T00:00:00Z","decision":` + decision + `}]`
}

// TestRecordDispatchBeginRefusesSubagentUnderInline pins the inline
// decision's enforcement: an implementer or panel-fix row with a subagent
// behind it is refused, while an inline row, a reviewer, a self-review
// fixer, a subagent decision and a failed decisions read all begin.
func TestRecordDispatchBeginRefusesSubagentUnderInline(t *testing.T) {
	inline := decisionsBody(`{"execution":"inline"}`)
	impl := []string{"-role", "implementer", "-key", "task-1-implementer", "-model", "opus", "-effort", "low"}
	fix := []string{"-role", "panel-fix", "-key", "panel-fix-1", "-model", "opus", "-effort", "low"}
	cases := []struct {
		name, decisions string
		args            []string
		code            int
	}{
		{"subagent implementer", inline, append([]string{"-agent-id", "none"}, impl...), 2},
		{"subagent panel-fix", inline, append([]string{"-agent-id", "none"}, fix...), 2},
		{"inline implementer", inline, append([]string{"-agent-id", "inline"}, impl...), 0},
		{"inline panel-fix", inline, append([]string{"-agent-id", "inline"}, fix...), 0},
		{"reviewer", inline, []string{"-agent-id", "none", "-role", "reviewer", "-key", "panel-1-primary", "-slot", "primary", "-model", "opus", "-effort", "low"}, 0},
		{"self-review fixer", inline, []string{"-agent-id", "none", "-role", "implementer", "-key", "self-review-x-fix", "-model", "opus", "-effort", "low"}, 0},
		{"subagent decision", decisionsBody(`{"execution":"subagent"}`), append([]string{"-agent-id", "none"}, impl...), 0},
		{"decisions read fails", "", append([]string{"-agent-id", "none"}, impl...), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, stderr := beginUnderDecision(t, tc.decisions, tc.args...)
			if code != tc.code {
				t.Fatalf("exit %d, want %d; stderr:\n%s", code, tc.code, stderr)
			}
			if tc.code == 2 && !strings.Contains(stderr, "execution: inline") {
				t.Errorf("stderr %q does not name execution: inline", stderr)
			}
		})
	}
}

// TestRecordDispatchBeginRefusesInlineUnderSdd pins the sdd decision's
// enforcement: an implementer or panel-fix row begun -agent-id inline on a
// task- or full-suite-fix- key is the parent's own work and is refused,
// while a self-review or panel fixer, a subagent row, a back-fill and a
// failed decisions read all begin.
func TestRecordDispatchBeginRefusesInlineUnderSdd(t *testing.T) {
	sdd := decisionsBody(`{"execution":"sdd"}`)
	row := func(agentID, role, key string, extra ...string) []string {
		return append([]string{"-agent-id", agentID, "-role", role, "-key", key, "-model", "opus", "-effort", "low"}, extra...)
	}
	cases := []struct {
		name, decisions string
		args            []string
		code            int
		want            string
	}{
		{"inline task implementer", sdd, row("inline", "implementer", "task-1-implementer"), 2, "execution: sdd"},
		{"inline full-suite implementer", sdd, row("inline", "implementer", "full-suite-fix-1"), 2, "execution: sdd"},
		{"inline full-suite panel-fix", sdd, row("inline", "panel-fix", "full-suite-fix-1"), 2, "execution: sdd"},
		{"inline self-review fixer", sdd, row("inline", "implementer", "self-review-x-fix"), 0, ""},
		{"inline panel-fix round", sdd, row("inline", "panel-fix", "panel-fix-1"), 0, ""},
		{"subagent implementer", sdd, row("none", "implementer", "task-1-implementer"), 0, ""},
		{"back-fill", sdd, row("inline", "implementer", "task-1-implementer", "-started-at", "2026-01-02T03:04:05Z"), 0, ""},
		{"decisions read fails", "", row("inline", "implementer", "task-1-implementer"), 0, ""},
		// Every fixer key an sdd parent dispatches is an implementer key
		// (skills/flow/implement.md, Every fix is a fresh one-shot).
		{"subagent full-suite fixer", sdd, row("none", "implementer", "full-suite-fix-1"), 0, ""},
		{"subagent recommit fixer", sdd, row("none", "implementer", "task-2-implementer-recommit-1"), 0, ""},
		{"subagent pick fixer", sdd, row("none", "implementer", "task-5+6-implementer-pick-1"), 0, ""},
		{"malformed full-suite key", sdd, row("none", "implementer", "full-suite-fixes"), 2, "is not task-<ids>-implementer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, stderr := beginUnderDecision(t, tc.decisions, tc.args...)
			if code != tc.code {
				t.Fatalf("exit %d, want %d; stderr:\n%s", code, tc.code, stderr)
			}
			if !strings.Contains(stderr, tc.want) {
				t.Errorf("stderr %q does not name %q", stderr, tc.want)
			}
		})
	}
}

// TestRecordDispatchBeginRefusesAboveSessionCap pins the decision's
// session_cap: a row above the cap's model or effort, an unstated default
// effort included, is refused; a row at or under it, and a decision with no
// cap, begins.
func TestRecordDispatchBeginRefusesAboveSessionCap(t *testing.T) {
	capped := func(model string) string {
		return decisionsBody(`{"execution":"subagent","session_cap":{"model":"` + model + `","effort":"low","instruction":"opus/low until 2026-10-28"}}`)
	}
	begin := func(model, effort string) []string {
		return []string{"-agent-id", "none", "-role", "reviewer", "-key", "panel-1-primary", "-slot", "primary", "-model", model, "-effort", effort}
	}
	cases := []struct {
		name, decisions string
		args            []string
		code            int
	}{
		{"effort above", capped("opus"), begin("opus", "medium"), 2},
		{"default effort", capped("opus"), begin("opus", "default"), 2},
		{"at the cap", capped("opus"), begin("opus", "low"), 0},
		{"lower model", capped("opus"), begin("sonnet", "low"), 0},
		{"model above", capped("sonnet"), begin("opus", "low"), 2},
		{"unranked model", capped("opus"), begin("inherit", "low"), 2},
		{"inline default effort", capped("opus"), []string{"-agent-id", "inline", "-role", "implementer", "-key", "task-1-implementer", "-model", "opus", "-effort", "default"}, 0},
		{"no cap", decisionsBody(`{"execution":"subagent","session_cap":null}`), begin("opus", "high"), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, stderr := beginUnderDecision(t, tc.decisions, tc.args...)
			if code != tc.code {
				t.Fatalf("exit %d, want %d; stderr:\n%s", code, tc.code, stderr)
			}
			if tc.code == 2 && !strings.Contains(stderr, "session cap") {
				t.Errorf("stderr %q does not name the session cap", stderr)
			}
		})
	}
}

// TestRecordFindingRejectsAbsoluteReproducer pins kan-1014's self-review: a
// reproducer whose path token is absolute can never be resolved by
// check-panel-findings-closed, so it is refused before the store is
// contacted; a relative command with a later absolute argument, and
// "none — <reason>", are still accepted.
func TestRecordFindingRejectsAbsoluteReproducer(t *testing.T) {
	repo := recordRepo(t, "kan-258")
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
			"-severity", "major", "-status", "open", "-reproducer", "/tmp/x.sh arg",
			"-note", "the note"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr:\n%s", code, stderr.String())
	}
	if contacted {
		t.Error("the store was contacted for an absolute reproducer -- it must be refused first")
	}
	for _, want := range []string{"absolute path", "relative to the worktree"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr does not name %q:\n%s", want, stderr.String())
		}
	}
	for _, ok := range []string{"scripts/run-verify-list.sh /abs", "none — prose only"} {
		if err := validateFindingReproducer(ok); err != nil {
			t.Errorf("%q refused: %v", ok, err)
		}
	}
}

// TestRecordDispatchBeginRefusesAVerifierBesideAnOpenMutator pins decision
// sr5-change-wide: a visual verifier builds the stack from the tree and
// writes into the change root, so it never runs beside an open implementer,
// panel-fix or red-partner row of the change, nor one of those beside an
// open visual verifier. A non-visual verifier and a reviewer are never
// refused, and a failed run-record read skips the check.
func TestRecordDispatchBeginRefusesAVerifierBesideAnOpenMutator(t *testing.T) {
	const ended = `"endedAt":"2026-10-26T10:00:00Z"`
	begin := func(t *testing.T, record string, readStatus int, role, key string) (int, string, int) {
		t.Helper()
		repo := recordRepo(t, "kan-258")
		isolatedStateRoot(t)
		posts := 0
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/decisions") {
				_, _ = w.Write([]byte("[]"))
				return
			}
			if r.Method == http.MethodGet {
				w.WriteHeader(readStatus)
				_, _ = w.Write([]byte(record))
				return
			}
			posts++
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
		}))
		defer srv.Close()
		args := []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", role, "-model", "opus", "-key", key, "-session-token", "mf-record-vv-mutator"}
		if role == "verifier" {
			args = append(args, "-diff-base", "0123456789abcdef0123456789abcdef01234567")
		}
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String(), posts
	}
	cases := []struct {
		name, record, role, key, names string
		status, want                   int
	}{
		{"a visual verifier beside an open panel-fix", `{"dispatches":[{"key":"panel-fix-7","role":"panel-fix"}]}`, "verifier", "visual-verify-2", "panel-fix-7", 200, 2},
		{"a panel-fix beside an open visual verifier", `{"findings":[{"ref":"F1","round":7,"slot":"primary","severity":"Important","note":"n","status":"open"}],"dispatches":[{"key":"visual-verify-2","role":"verifier"},{"key":"panel-7-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`, "panel-fix", "panel-fix-8", "visual-verify-2", 200, 2},
		{"a visual verifier beside an open verify-* verifier", `{"dispatches":[{"key":"verify-test-fix-1","role":"verifier"}]}`, "verifier", "visual-verify-2", "", 200, 0},
		{"a visual verifier beside an open reviewer", `{"dispatches":[{"key":"panel-3-primary","role":"reviewer"}]}`, "verifier", "visual-verify-2", "", 200, 0},
		{"a visual verifier after the panel-fix ended", `{"dispatches":[{"key":"panel-fix-7","role":"panel-fix",` + ended + `}]}`, "verifier", "visual-verify-2", "", 200, 0},
		{"a panel-fix after the visual verifier ended", `{"findings":[{"ref":"F1","round":7,"slot":"primary","severity":"Important","note":"n","status":"open"}],"dispatches":[{"key":"visual-verify-2","role":"verifier",` + ended + `},{"key":"panel-7-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`, "panel-fix", "panel-fix-8", "", 200, 0},
		{"a failed read skips the check", `{"error":"boom"}`, "verifier", "visual-verify-2", "", 500, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, stderr, posts := begin(t, c.record, c.status, c.role, c.key)
			if code != c.want || (posts == 1) != (c.want == 0) {
				t.Fatalf("exit %d, posts %d; want exit %d; stderr:\n%s", code, posts, c.want, stderr)
			}
			if c.names != "" && !strings.Contains(stderr, c.names) {
				t.Errorf("stderr does not name %s:\n%s", c.names, stderr)
			}
		})
	}
}

// TestRecordDispatchBeginRefusesAnUnscopedVisualVerifyAfterAnInRunFix pins
// decision sr17-tasks-md-read: once an In-run fix <k> task is ticked, and
// until a -fix-<k>-full verifier row completed, a visual-verify key without
// -fix-<k> is refused, and so is a non-full one with no -round-base sharing
// its repository suffix with a completed verifier row.
func TestRecordDispatchBeginRefusesAnUnscopedVisualVerifyAfterAnInRunFix(t *testing.T) {
	const ticked = "# plan\n\n- [x] 1. a task\n- [x] 3. In-run fix 2 — flow.visual-verify: the chip colour\n"
	const done = `{"key":"visual-verify-gymie","role":"verifier","outcome":"completed","endedAt":"2026-10-26T10:00:00Z"}`
	const full = `{"key":"visual-verify-fix-2-full-gymie","role":"verifier","outcome":"completed","endedAt":"2026-10-26T11:00:00Z"}`
	begin := func(t *testing.T, plan, dispatches, key string, extra ...string) (int, string, int) {
		t.Helper()
		repo := recordRepo(t, "kan-258")
		isolatedStateRoot(t)
		if err := os.WriteFile(filepath.Join(repo, "spectre", "changes", "kan-258", "tasks.md"), []byte(plan), 0o644); err != nil {
			t.Fatal(err)
		}
		posts := 0
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/decisions") {
				_, _ = w.Write([]byte("[]"))
				return
			}
			if r.Method == http.MethodGet {
				_, _ = w.Write([]byte(`{"change":"kan-258","dispatches":[` + dispatches + `]}`))
				return
			}
			posts++
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
		}))
		defer srv.Close()
		var stdout, stderr bytes.Buffer
		code := run(context.Background(),
			append([]string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
				"-change", "kan-258", "-role", "verifier", "-model", "opus", "-key", key, "-session-token", "mf-record-vv-in-run-fix",
				"-diff-base", "0123456789abcdef0123456789abcdef01234567"}, extra...),
			strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String(), posts
	}
	cases := []struct {
		name, plan, dispatches, key string
		extra                       []string
		want                        int
	}{
		{"an unscoped re-run after the fix", ticked, done, "visual-verify-3-gymie", nil, 2},
		{"a scoped re-run with no round base", ticked, done, "visual-verify-fix-2-gymie", nil, 2},
		{"a scoped re-run with its round base", ticked, done, "visual-verify-fix-2-gymie", []string{"-round-base", "abc1234"}, 0},
		{"the full run", ticked, done, "visual-verify-fix-2-full-gymie", nil, 0},
		{"a scoped run in a repository with no completed row", ticked, done, "visual-verify-fix-2-frontend", nil, 0},
		{"an unscoped run once the full run completed", ticked, done + "," + full, "visual-verify-4-gymie", nil, 0},
		{"a scoped run of the highest of two open fixes", ticked + "- [x] 4. In-run fix 3 — flow.visual-verify: the label\n", done, "visual-verify-fix-3-gymie", []string{"-round-base", "abc1234"}, 0},
		{"a scoped run of the lower of two open fixes", ticked + "- [x] 4. In-run fix 3 — flow.visual-verify: the label\n", done, "visual-verify-fix-2-gymie", []string{"-round-base", "abc1234"}, 2},
		{"no ticked in-run fix", "# plan\n\n- [x] 1. a task\n- [ ] 3. In-run fix 2 — flow.visual-verify: x\n", done, "visual-verify-2-gymie", nil, 0},
		{"a ticked in-run fix inside a fence", "# plan\n\n```\n- [x] 3. In-run fix 2 — flow.visual-verify: x\n```\n", done, "visual-verify-2-gymie", nil, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, stderr, posts := begin(t, c.plan, c.dispatches, c.key, c.extra...)
			if code != c.want || (posts == 1) != (c.want == 0) {
				t.Fatalf("exit %d, posts %d; want exit %d; stderr:\n%s", code, posts, c.want, stderr)
			}
		})
	}
}

// A change archived only on another branch -- a feature base the default
// branch's main checkout has not merged -- is still this repository's change.
func TestRequireChangeCheckoutAcceptsAnArchiveOnAnotherBranch(t *testing.T) {
	repo := gitRepo(t)
	for _, args := range [][]string{
		{"checkout", "-q", "-b", "feature/base"},
		{"commit", "-q", "--allow-empty", "-m", "base"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	dir := filepath.Join(repo, "spectre", "changes", "archive", "kan-7")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "-A"},
		{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "commit", "-q", "-m", "archive"},
		{"checkout", "-q", "-"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if !requireChangeCheckout(repo, "kan-7", io.Discard) {
		t.Error("a change archived on feature/base refused")
	}
	if requireChangeCheckout(repo, "kan-8", io.Discard) {
		t.Error("a change no branch carries accepted")
	}
}

// A creating /flow run records its brainstorming dispatches before spectre new
// has made spectre/changes/<change>: its worktree on branch spectre/<change>
// already carries the change.
func TestRequireChangeCheckoutAcceptsTheChangeBranch(t *testing.T) {
	repo := gitRepo(t)
	for _, args := range [][]string{
		{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "commit", "-q", "--allow-empty", "-m", "base"},
		{"worktree", "add", "-q", "-b", "spectre/kan-9", filepath.Join(t.TempDir(), "wt")},
	} {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if !requireChangeCheckout(repo, "kan-9", io.Discard) {
		t.Error("a worktree on spectre/kan-9 refused")
	}
	if requireChangeCheckout(repo, "kan-10", io.Discard) {
		t.Error("a change no worktree carries accepted")
	}
}

// A verify row records a run, not an edit: an inline verifier begin over a
// modified record is admitted, an inline implementer begin still refused
// (design.md verifier-skips-inline-dirty).
func TestRecordDispatchBeginInlineDirtyTree(t *testing.T) {
	repo := recordRepo(t, "x")
	isolatedStateRoot(t)
	record := filepath.Join(repo, "spectre", "changes", "x", "visual-verification.md")
	if err := os.WriteFile(record, []byte("# v\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "record"}} {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if err := os.WriteFile(record, []byte("# v\nround 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	begin := func(role, key string) (int, string) {
		t.Helper()
		srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
		}))
		defer srv.Close()
		argv := []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "inline", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "x", "-role", role, "-model", "opus", "-key", key, "-session-token", "mf-record-inline-verifier"}
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), argv, strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String()
	}
	if code, stderr := begin("verifier", "verify-x"); code != 0 {
		t.Errorf("an inline verifier begin over a modified record: exit %d, want 0; stderr:\n%s", code, stderr)
	}
	if code, stderr := begin("implementer", "task-1-implementer"); code != 2 || !strings.Contains(stderr, "spectre/changes/x/visual-verification.md") {
		t.Errorf("an inline implementer begin over the same tree: exit %d, want 2 naming the record; stderr:\n%s", code, stderr)
	}
}

// TestRecordDispatchRefusesPanelFixZero is kan-1040's self-review: the fix
// after panel-0 is panel-fix-1, so begin and backfill refuse a panel-fix
// key whose round is 0 (or carries a leading zero) before any store call,
// and journal nothing; panel-fix-1 and full-suite-fix-1 pass this check.
func TestRecordDispatchRefusesPanelFixZero(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-1040")
	addr := deadPortAddr(t)
	const refusal = "counts fix rounds from 1"
	args := func(verb, key string) []string {
		a := []string{"record", "dispatch", verb, "-addr", addr, "-timeout", "200ms", "-C", repo,
			"-change", "kan-1040", "-role", "panel-fix", "-model", "opus", "-effort", "high", "-agent-id", "none",
			"-key", key, "-session-token", "mf-panel-fix-zero"}
		if verb == "backfill" {
			a = append(a, "-started-at", "2026-10-27T12:00:00Z", "-ended-at", "2026-10-27T12:10:00Z", "-outcome", "completed")
		}
		return a
	}
	for _, verb := range []string{"begin", "backfill"} {
		for _, key := range []string{"panel-fix-0", "panel-fix-01"} {
			root := isolatedStateRoot(t)
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), args(verb, key), strings.NewReader(""), &stdout, &stderr)
			if code != 2 || !strings.Contains(stderr.String(), key) || !strings.Contains(stderr.String(), "panel-fix-1") {
				t.Errorf("%s %s: exit %d, stderr %q; want 2 naming the key and panel-fix-1", verb, key, code, stderr.String())
			}
			if entries, _ := os.ReadDir(root); len(entries) != 0 {
				t.Errorf("%s %s: state root holds %d entries, want nothing journalled", verb, key, len(entries))
			}
		}
		for _, key := range []string{"panel-fix-1", "full-suite-fix-1"} {
			isolatedStateRoot(t)
			var stdout, stderr bytes.Buffer
			run(context.Background(), args(verb, key), strings.NewReader(""), &stdout, &stderr)
			if strings.Contains(stderr.String(), refusal) {
				t.Errorf("%s %s: refused as a round-0 key: %q", verb, key, stderr.String())
			}
		}
	}
}

// TestRecordDispatchRefusesAPanelFixKeyedARoundLow pins the kan-1041
// self-review fix: a panel-fix-<r> row is refused once round r's reviewers
// began before the fix ended -- it fixes round r and is panel-fix-<r+1> --
// on an inline back-fill, an explicit back-fill and a live begin alike; the
// right key, a back-filled begin and a failed record read pass.
func TestRecordDispatchRefusesAPanelFixKeyedARoundLow(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	fixtureCommit(t, repo, "2026-10-28T02:40:00Z")
	sha := fixtureCommit(t, repo, "2026-10-28T02:48:46Z")
	record := `{"change":"kan-258","findings":[{"ref":"F1","round":5,"slot":"primary","severity":"Important","note":"n","status":"open"}],"dispatches":[` +
		`{"key":"panel-5-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-28T02:45:33Z","endedAt":"2026-10-28T02:47:03Z"},` +
		`{"key":"panel-4-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"},{"key":"panel-5-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
	var posts int
	recordFails := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			if recordFails {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		posts++
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"panel-fix","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	conn := []string{"-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-258", "-role", "panel-fix", "-model", "opus"}
	backfill := func(key string, extra ...string) []string {
		return append(append(append([]string{"record", "dispatch", "backfill"}, conn...), "-key", key,
			"-session-token", "mf-record-panel-fix-round", "-outcome", "completed"), extra...)
	}
	begin := func(key string, extra ...string) []string {
		return append(append(append([]string{"record", "dispatch", "begin"}, conn...), "-key", key,
			"-session-token", "mf-record-panel-fix-round", "-effort", "default"), extra...)
	}
	for _, tc := range []struct {
		name        string
		args        []string
		recordFails bool
		want        int
	}{
		{"inline back-fill a round low", backfill("panel-fix-5", "-agent-id", "inline", "-commit", sha), false, 2},
		{"inline back-fill chunk a round low", backfill("panel-fix-5-2", "-agent-id", "inline", "-commit", sha), false, 2},
		{"explicit back-fill a round low", backfill("panel-fix-5", "-started-at", "2026-10-28T02:47:03Z", "-ended-at", "2026-10-28T02:48:46Z"), false, 2},
		{"live begin a round low", begin("panel-fix-5"), false, 2},
		{"inline back-fill on the fix round", backfill("panel-fix-6", "-agent-id", "inline", "-commit", sha), false, 0},
		{"live begin on the fix round", begin("panel-fix-6"), false, 0},
		{"back-fill ending before the round began", backfill("panel-fix-5", "-started-at", "2026-10-28T02:40:00Z", "-ended-at", "2026-10-28T02:45:00Z"), false, 0},
		{"a failed record read", backfill("panel-fix-5", "-agent-id", "inline", "-commit", sha), true, 0},
	} {
		posts = 0
		recordFails = tc.recordFails
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), tc.args, strings.NewReader(""), &stdout, &stderr)
		if code != tc.want {
			t.Errorf("%s: exit %d, want %d; stderr:\n%s", tc.name, code, tc.want, stderr.String())
		}
		if tc.want == 2 && (posts != 0 || !strings.Contains(stderr.String(), "panel-fix-6")) {
			t.Errorf("%s: posts %d, stderr %q; want none and a message naming panel-fix-6", tc.name, posts, stderr.String())
		}
	}
}

// TestRecordDispatchBackfillFloorsAnInlineRowAtTheOtherInlineRows pins the
// kan-1041 self-review fix: an inline back-fill's derived start never comes
// before the latest end, not after its commit, among the change's other
// inline rows, so its window holds no other inline row; a non-inline row,
// one ending after the commit and a failed record read set no floor.
func TestRecordDispatchBackfillFloorsAnInlineRowAtTheOtherInlineRows(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	fixtureCommit(t, repo, "2026-10-28T02:48:46Z")
	sha := fixtureCommit(t, repo, "2026-10-28T03:02:22Z")
	withInline := `{"change":"kan-258","dispatches":[` +
		`{"key":"verify-fix-3","agentId":"inline","role":"verifier","startedAt":"2026-10-28T02:51:19Z","endedAt":"2026-10-28T03:02:17Z"},` +
		`{"key":"panel-6-mutation","agentId":"a1","role":"reviewer","slot":"mutation","startedAt":"2026-10-28T02:52:00Z","endedAt":"2026-10-28T03:02:20Z"},` +
		`{"key":"verify-fix-4","agentId":"inline","role":"verifier","startedAt":"2026-10-28T03:02:30Z","endedAt":"2026-10-28T03:08:00Z"}]}`
	record := withInline
	recordFails := false
	var bodies []map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			if recordFails {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		raw, _ := readAll(r)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		bodies = append(bodies, b)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	for _, tc := range []struct {
		name, record, want string
		recordFails        bool
	}{
		{"floored at the other inline row's end", withInline, "2026-10-28T03:02:17Z", false},
		{"no other inline row", `{"change":"kan-258","dispatches":[]}`, "2026-10-28T02:48:46Z", false},
		{"a failed record read", withInline, "2026-10-28T02:48:46Z", true},
	} {
		bodies = nil
		record, recordFails = tc.record, tc.recordFails
		args := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-key", "task-13-implementer", "-task", "13",
			"-session-token", "mf-record-inline-floor", "-outcome", "completed", "-agent-id", "inline", "-commit", sha}
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
			t.Fatalf("%s: exit %d; stderr:\n%s", tc.name, code, stderr.String())
		}
		if len(bodies) != 2 || bodies[0]["startedAt"] != tc.want {
			t.Errorf("%s: bodies = %v, want begin startedAt %s", tc.name, bodies, tc.want)
		}
	}
}

// panelFixBegin begins key as a panel-fix of kan-258 under token against a
// daemon serving record, returning the exit code, stderr and the posts made.
func panelFixBegin(t *testing.T, record, key, token string, extra ...string) (int, string, int) {
	t.Helper()
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	posts := 0
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/decisions") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("[]"))
			return
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(record))
			return
		}
		posts++
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
	}))
	defer srv.Close()
	verb := "begin"
	if len(extra) > 0 {
		verb = "backfill"
	}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		append([]string{"record", "dispatch", verb, "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "panel-fix", "-model", "opus", "-key", key, "-session-token", token}, extra...),
		strings.NewReader(""), &stdout, &stderr)
	return code, stderr.String(), posts
}

func TestRecordDispatchRefusesPanelFixWithoutFindings(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	minor := `{"change":"kan-258","findings":[{"ref":"F1","round":1,"slot":"primary","severity":"Minor","note":"n","status":"open"}],"dispatches":[{"key":"panel-1-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
	important := `{"change":"kan-258","findings":[{"ref":"F1","round":1,"slot":"primary","severity":"important","note":"n","status":"open"}],"dispatches":[{"key":"panel-1-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
	t.Run("a Minor-only round is refused", func(t *testing.T) {
		code, stderr, posts := panelFixBegin(t, minor, "panel-fix-2", "mf-pf-a")
		if code != 2 || posts != 0 || !strings.Contains(stderr, "round 1") || !strings.Contains(stderr, "flow record finding -from") {
			t.Fatalf("exit %d, posts %d; want 2, 0 naming round 1; stderr:\n%s", code, posts, stderr)
		}
	})
	t.Run("an important row admits it", func(t *testing.T) {
		if code, stderr, posts := panelFixBegin(t, important, "panel-fix-2", "mf-pf-a"); code != 0 || posts != 1 {
			t.Fatalf("exit %d, posts %d; want 0, 1; stderr:\n%s", code, posts, stderr)
		}
	})
	t.Run("a backfill is admitted", func(t *testing.T) {
		if code, stderr, posts := panelFixBegin(t, minor, "panel-fix-2", "mf-pf-a",
			"-started-at", "2026-10-27T12:00:00Z", "-ended-at", "2026-10-27T12:10:00Z", "-outcome", "completed"); code != 0 || posts == 0 {
			t.Fatalf("exit %d, posts %d; want 0 and a post; stderr:\n%s", code, posts, stderr)
		}
	})
}

func TestRecordDispatchRefusesReusedPanelFixKey(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	rec := func(outcome string) string {
		return `{"change":"kan-258","findings":[{"ref":"F1","round":0,"slot":"primary","severity":"Important","note":"n","status":"open"}],` +
			`"dispatches":[{"key":"panel-fix-1","role":"panel-fix","model":"opus","sessionToken":"mf-pf-other","outcome":"` + outcome + `","startedAt":"2026-10-27T12:00:00Z"},{"key":"panel-0-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
	}
	t.Run("another session's key is refused", func(t *testing.T) {
		code, stderr, posts := panelFixBegin(t, rec("completed"), "panel-fix-1", "mf-pf-mine")
		if code != 2 || posts != 0 || !strings.Contains(stderr, "mf-pf-other") {
			t.Fatalf("exit %d, posts %d; want 2, 0 naming mf-pf-other; stderr:\n%s", code, posts, stderr)
		}
	})
	t.Run("an aborted reuse is admitted", func(t *testing.T) {
		if code, stderr, posts := panelFixBegin(t, rec("aborted"), "panel-fix-1", "mf-pf-mine"); code != 0 || posts != 1 {
			t.Fatalf("exit %d, posts %d; want 0, 1; stderr:\n%s", code, posts, stderr)
		}
	})
	t.Run("the same token is admitted", func(t *testing.T) {
		if code, stderr, posts := panelFixBegin(t, rec("completed"), "panel-fix-1", "mf-pf-other"); code != 0 || posts != 1 {
			t.Fatalf("exit %d, posts %d; want 0, 1; stderr:\n%s", code, posts, stderr)
		}
	})
}

// TestPanelFixBeginRefusalReusedKeyNamesTheAbortedRemedy pins the kan-1024
// self-review fix: a reused panel-fix-1 whose round 1 holds no blocking
// finding is pointed at ending the other row aborted, never at a
// panel-fix-2 the first check would refuse; a later round is suggested only
// when round <next> holds a Critical or Important finding.
func TestPanelFixBeginRefusalReusedKeyNamesTheAbortedRemedy(t *testing.T) {
	t.Parallel()
	run := records.Run{
		Findings:   []records.Finding{{Round: 0, Severity: "Critical"}},
		Dispatches: []records.Dispatch{{Key: "panel-fix-1", SessionToken: "other", Outcome: "blocked"}},
	}
	msg := panelFixBeginRefusal(run, "panel-fix-1", "1", "mine")
	if !strings.Contains(msg, "-session-token other -outcome aborted") || strings.Contains(msg, "panel-fix-2") {
		t.Errorf("reused key with an empty round 1: %q", msg)
	}
	run.Findings = append(run.Findings, records.Finding{Round: 1, Severity: "Important"})
	if msg := panelFixBeginRefusal(run, "panel-fix-1", "1", "mine"); !strings.Contains(msg, "a later fix round is panel-fix-2") {
		t.Errorf("reused key with a blocking round 1: %q", msg)
	}
}

// TestRecordDispatchBackfillInlineFloorsAtVisualVerify pins the kan-1046
// self-review fix (1b): an inline back-fill never starts before a
// visual-verify verifier row's end not after its commit, since an in-run
// visual fix answers that verdict; a verdict ending after the commit is no
// floor.
func TestRecordDispatchBackfillInlineFloorsAtVisualVerify(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	fixtureCommit(t, repo, "2026-10-28T14:26:41Z")
	sha := fixtureCommit(t, repo, "2026-10-28T14:37:26Z")
	verdict := func(end string) string {
		return `{"change":"kan-258","dispatches":[` +
			`{"key":"visual-verify","agentId":"a1","role":"verifier","startedAt":"2026-10-28T14:31:14Z","endedAt":"` + end + `"}]}`
	}
	var record string
	var bodies []map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		raw, _ := readAll(r)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		bodies = append(bodies, b)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	for _, tc := range []struct{ name, record, want string }{
		{"floored at the verdict", verdict("2026-10-28T14:34:31Z"), "2026-10-28T14:34:31Z"},
		{"a verdict after the commit", verdict("2026-10-28T14:40:00Z"), "2026-10-28T14:26:41Z"},
	} {
		bodies = nil
		record = tc.record
		args := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-key", "task-3-implementer", "-task", "3",
			"-session-token", "mf-record-visual-floor", "-outcome", "completed", "-agent-id", "inline", "-commit", sha}
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
			t.Fatalf("%s: exit %d; stderr:\n%s", tc.name, code, stderr.String())
		}
		if len(bodies) != 2 || bodies[0]["startedAt"] != tc.want {
			t.Errorf("%s: bodies = %v, want begin startedAt %s", tc.name, bodies, tc.want)
		}
	}
}

// TestRecordDispatchBackfillRefusesPanelFixOverOpenRound pins the kan-1046
// self-review fix (1c): an inline panel-fix-<r> back-fill is refused, exit 2
// with no store call, while a reviewer row of round r-1 is open; with that
// row ended it passes and floors at its end.
func TestRecordDispatchBackfillRefusesPanelFixOverOpenRound(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	fixtureCommit(t, repo, "2026-10-28T14:37:39Z")
	sha := fixtureCommit(t, repo, "2026-10-28T14:41:35Z")
	open := `{"change":"kan-258","dispatches":[` +
		`{"key":"panel-1-primary","agentId":"a1","role":"reviewer","slot":"primary","startedAt":"2026-10-28T14:38:10Z"},` +
		`{"key":"panel-1-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
	closed := `{"change":"kan-258","dispatches":[` +
		`{"key":"panel-1-primary","agentId":"a1","role":"reviewer","slot":"primary","startedAt":"2026-10-28T14:38:10Z","endedAt":"2026-10-28T14:40:55Z"},` +
		`{"key":"panel-1-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
	var record string
	var bodies []map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		raw, _ := readAll(r)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		bodies = append(bodies, b)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"panel-fix","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	args := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "kan-258", "-role", "panel-fix", "-model", "opus", "-key", "panel-fix-2",
		"-session-token", "mf-record-open-round", "-outcome", "completed", "-agent-id", "inline", "-commit", sha}
	record = open
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 2 || len(bodies) != 0 {
		t.Fatalf("open round: exit %d, store calls %d; want 2 and none (stderr %q)", code, len(bodies), stderr.String())
	}
	if !strings.Contains(stderr.String(), "close panel round 1's reviewer rows first") {
		t.Errorf("open round: stderr %q does not name the open round", stderr.String())
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Errorf("open round: wrote a record journal")
	}
	record = closed
	stderr.Reset()
	if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("closed round: exit %d; stderr:\n%s", code, stderr.String())
	}
	if len(bodies) != 2 || bodies[0]["startedAt"] != "2026-10-28T14:40:55Z" {
		t.Errorf("closed round: bodies = %v, want begin startedAt 14:40:55Z", bodies)
	}
}

// TestRecordMutationRefusesAfterReReviewBegan pins the kan-1046 self-review
// fix (1d): a mutation proof for round r is refused, exit 2 with no store
// write, once a reviewer row of panel round r has begun; a round with no
// reviewer row yet passes.
func TestRecordMutationRefusesAfterReReviewBegan(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	if err := os.WriteFile(filepath.Join(repo, "render.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	record := `{"change":"kan-258","dispatches":[` +
		`{"key":"panel-1-primary","agentId":"a1","role":"reviewer","slot":"primary","startedAt":"2026-10-28T14:38:10Z"}]}`
	posts := 0
	lastBody := ""
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		posts++
		body, _ := io.ReadAll(r.Body)
		lastBody = string(body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":12,"round":2,"path":"render.go","mutated":"x","test":"y"}`))
	}))
	defer srv.Close()
	args := func(round string) []string {
		return []string{"record", "mutation", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-round", round, "-path", "render.go", "-mutated", "x", "-test", "y"}
	}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), args("1"), strings.NewReader(""), &stdout, &stderr); code != 2 || posts != 0 {
		t.Fatalf("round 1: exit %d, posts %d; want 2 and none (stderr %q)", code, posts, stderr.String())
	}
	if !strings.Contains(stderr.String(), "panel round 1's re-review began") {
		t.Errorf("round 1: stderr %q does not name panel round 1", stderr.String())
	}
	if _, exists := recordJournalEntries(t, repo, "kan-258"); exists {
		t.Errorf("round 1: wrote a record journal")
	}
	stderr.Reset()
	if code := run(context.Background(), args("2"), strings.NewReader(""), &stdout, &stderr); code != 0 || posts != 1 {
		t.Fatalf("round 2: exit %d, posts %d; want 0 and one (stderr %q)", code, posts, stderr.String())
	}
	// The late form admits round 1 and records its reason as a pass note
	// (kan-1047 pipeline fix).
	stderr.Reset()
	if code := run(context.Background(), append(args("1"), "-late", "missed before re-review"), strings.NewReader(""), &stdout, &stderr); code != 0 || posts != 3 {
		t.Fatalf("late round 1: exit %d, posts %d; want 0 and three (stderr %q)", code, posts, stderr.String())
	}
	if !strings.Contains(lastBody, "late mutation proof for render.go: missed before re-review") {
		t.Errorf("late round 1: pass body %q carries no late note", lastBody)
	}
}

// TestRecordMutationLateEdges pins the kan-1047 pipeline-fix review's
// Minors: a failed mutation write writes no pass row; -late on a round whose
// re-review has not begun writes no late note; a -late retry whose mutation
// row already landed writes only the pass note.
func TestRecordMutationLateEdges(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	if err := os.WriteFile(filepath.Join(repo, "render.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dispatch := `{"key":"panel-1-primary","agentId":"a1","role":"reviewer","slot":"primary","startedAt":"2026-10-28T14:38:10Z"}`
	record, failMutation := "", false
	var posts []string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		posts = append(posts, r.URL.Path)
		if failMutation && strings.Contains(r.URL.Path, "mutation") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"bad"}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":12}`))
	}))
	defer srv.Close()
	args := func(round string) []string {
		return []string{"record", "mutation", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-round", round, "-path", "render.go", "-mutated", "x", "-test", "y", "-late", "missed"}
	}
	var stdout, stderr bytes.Buffer
	cases := []struct {
		name, record string
		fail         bool
		wantZero     bool
		want         []string
	}{
		{"failed mutation write", `{"change":"kan-258","dispatches":[` + dispatch + `]}`, true, false, []string{"mutation"}},
		{"re-review not begun", `{"change":"kan-258"}`, false, true, []string{"mutation"}},
		{"retry after landed row", `{"change":"kan-258","dispatches":[` + dispatch + `],"mutations":[{"round":1,"path":"render.go","mutated":"x","test":"y"}]}`, false, true, []string{"pass"}},
	}
	for _, c := range cases {
		record, failMutation, posts = c.record, c.fail, nil
		stderr.Reset()
		code := run(context.Background(), args("1"), strings.NewReader(""), &stdout, &stderr)
		if (code == 0) != c.wantZero || len(posts) != len(c.want) {
			t.Fatalf("%s: exit %d, posts %v; want zero=%v and %v (stderr %q)", c.name, code, posts, c.wantZero, c.want, stderr.String())
		}
		for i, w := range c.want {
			if !strings.Contains(posts[i], w) {
				t.Errorf("%s: post %d = %q, want %s", c.name, i, posts[i], w)
			}
		}
	}
}

// recordServer serves *record on GET, records each POST body and answers it
// as a written dispatch (kan-1044 self-review, F1-b).
func recordServer(t *testing.T, record *string) (*httptest.Server, *[]map[string]any) {
	t.Helper()
	var bodies []map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/decisions") {
			_, _ = w.Write([]byte("[]"))
			return
		}
		if r.Method == http.MethodGet {
			if *record == "" {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(*record))
			return
		}
		raw, _ := readAll(r)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		bodies = append(bodies, b)
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"implementer","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &bodies
}

// TestRecordDispatchEndInlineDerivesItsStart pins kan-1044 self-review F1-b:
// a completed end of an inline task implementer row with -commit sends the
// start the back-fill would derive; a subagent row sends none.
func TestRecordDispatchEndInlineDerivesItsStart(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	fixtureCommit(t, repo, "2026-10-21T11:02:00Z")
	sha := fixtureCommit(t, repo, "2026-10-21T11:10:31Z")
	reviewer := `,{"key":"task-8-reviewer","agentId":"a2","role":"reviewer","outcome":"completed","startedAt":"2026-10-21T11:02:00Z","endedAt":"2026-10-21T11:06:00Z"}`
	for _, tc := range []struct{ name, agent, other, want string }{
		{"inline", "inline", "", "2026-10-21T11:02:00Z"},
		{"subagent", "a1", "", ""},
		{"inline after a reviewer", "inline", reviewer, "2026-10-21T11:06:00Z"},
	} {
		record := `{"change":"kan-258","dispatches":[{"key":"task-9-implementer","agentId":"` + tc.agent +
			`","role":"implementer","sessionToken":"mf-end-derive","startedAt":"2026-10-21T11:10:25Z"}` + tc.other + `]}`
		srv, bodies := recordServer(t, &record)
		args := []string{"record", "dispatch", "end", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-key", "task-9-implementer", "-session-token", "mf-end-derive",
			"-outcome", "completed", "-commit", sha}
		var stdout, stderr bytes.Buffer
		if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
			t.Fatalf("%s: exit %d; stderr:\n%s", tc.name, code, stderr.String())
		}
		got, _ := (*bodies)[0]["startedAt"].(string)
		if len(*bodies) != 1 || got != tc.want {
			t.Errorf("%s: bodies = %v, want startedAt %q", tc.name, *bodies, tc.want)
		}
	}
}

// TestRecordDispatchBackfillRefusesADuplicateTaskCommit pins kan-1044
// self-review F1-b: a task implementer back-fill on the commit a completed
// row of the same task carries exits 2 before any store call; another
// commit, and a failed record read, are admitted.
func TestRecordDispatchBackfillRefusesADuplicateTaskCommit(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	fixtureCommit(t, repo, "2026-10-28T15:10:00Z")
	sha := fixtureCommit(t, repo, "2026-10-28T15:25:48Z")
	same := `{"change":"kan-258","dispatches":[{"key":"task-8-implementer","agentId":"inline","role":"implementer","outcome":"completed","commitSha":"` + sha + `","startedAt":"2026-10-28T15:20:08Z"}]}`
	for _, tc := range []struct {
		name, record, key string
		want              int
	}{
		{"same task, same commit", same, "task-8-implementer-retry", 2},
		{"same task, other commit", `{"change":"kan-258","dispatches":[{"key":"task-8-implementer","agentId":"inline","role":"implementer","outcome":"completed","commitSha":"0123456789abcdef0123456789abcdef01234567","startedAt":"2026-10-28T15:20:08Z"}]}`, "task-8-implementer-retry", 0},
		{"a failed record read", "", "task-8-implementer-retry", 0},
		{"a folded fix round", same, "task-8-implementer-fix-1", 0},
	} {
		record := tc.record
		srv, bodies := recordServer(t, &record)
		args := []string{"record", "dispatch", "backfill", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-key", tc.key, "-task", "8",
			"-session-token", "mf-record-dup", "-outcome", "completed", "-agent-id", "inline", "-commit", sha}
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		if code != tc.want || (tc.want == 2 && len(*bodies) != 0) {
			t.Errorf("%s: exit %d, posts %d; want %d (stderr %q)", tc.name, code, len(*bodies), tc.want, stderr.String())
		}
	}
}

// TestRecordDispatchBeginRefusesACompletedTask pins kan-1044 self-review
// F1-b: an inline begin of a task implementer key is refused while the task
// already holds a completed implementer row.
func TestRecordDispatchBeginRefusesACompletedTask(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	for _, tc := range []struct {
		name, outcome, key string
		want               int
	}{
		{"completed", "completed", "task-7-implementer", 2},
		{"aborted", "aborted", "task-7-implementer", 0},
		{"a gated-review fix round", "completed", "task-7-implementer-fix-1", 0},
	} {
		record := `{"change":"kan-258","dispatches":[{"key":"task-7-implementer-retry","agentId":"inline","role":"implementer","outcome":"` +
			tc.outcome + `","commitSha":"37c868ea2100fe869ff241bada9e7d122a93d352","startedAt":"2026-10-28T14:22:58Z"}]}`
		srv, bodies := recordServer(t, &record)
		args := []string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "implementer", "-model", "opus", "-effort", "default", "-key", tc.key, "-task", "7",
			"-session-token", "mf-record-done", "-agent-id", "inline"}
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		if code != tc.want || (tc.want == 2 && len(*bodies) != 0) {
			t.Errorf("%s: exit %d, posts %d; want %d (stderr %q)", tc.name, code, len(*bodies), tc.want, stderr.String())
		}
	}
}

// trailerCommit commits on repo, authored and committed at, carrying the
// Flow-Key: trailer.
func trailerCommit(t *testing.T, repo, at, trailer string) {
	t.Helper()
	cmd := exec.Command("git", "-C", repo, "commit", "--allow-empty", "-q", "-m", "fix: review findings", "-m", "Flow-Key: "+trailer)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_AUTHOR_DATE="+at, "GIT_COMMITTER_DATE="+at)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
}

// TestRecordDispatchRefusesAReReviewOverAnOffRoundPanelFixTrailer pins the
// kan-1059 self-review fix, 1a: a live re-review of round 5 is refused while
// a commit authored since round 4's reviewers began, in a worktree the
// change's state names, carries Flow-Key: panel-fix-4; the fix round's own
// trailer, a back-fill and a commit authored before the round pass.
func TestRecordDispatchRefusesAReReviewOverAnOffRoundPanelFixTrailer(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	record := `{"change":"kan-258","dispatches":[` +
		`{"key":"panel-4-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-28T02:45:33Z","endedAt":"2026-10-28T02:47:03Z"},` +
		`{"key":"panel-4-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
	var peer string
	posts := 0
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method == http.MethodGet {
			body, _ := json.Marshal(map[string]any{"worktrees": map[string]string{peer: "abc123"}})
			_, _ = w.Write(body)
			return
		}
		posts++
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"reviewer","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	for _, tc := range []struct {
		name, at, trailer string
		extra             []string
		want              int
	}{
		{"a trailer a round low", "2026-10-28T02:50:00Z", "panel-fix-4", nil, 2},
		{"the fix round's trailer", "2026-10-28T02:50:00Z", "panel-fix-5", nil, 0},
		{"a back-fill", "2026-10-28T02:50:00Z", "panel-fix-4", []string{"-started-at", "2026-10-28T02:55:00Z"}, 0},
		{"a commit authored before the round", "2026-10-28T02:40:00Z", "panel-fix-4", nil, 0},
	} {
		peer = gitRepo(t)
		trailerCommit(t, peer, tc.at, tc.trailer)
		posts = 0
		args := append([]string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "reviewer", "-model", "opus", "-key", "panel-5-primary", "-slot", "primary",
			"-session-token", "mf-record-rereview", "-effort", "default"}, tc.extra...)
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		if code != tc.want {
			t.Errorf("%s: exit %d, want %d; stderr:\n%s", tc.name, code, tc.want, stderr.String())
		}
		if tc.want == 2 && (posts != 0 || !strings.Contains(stderr.String(), "panel-fix-5") || !strings.Contains(stderr.String(), "commit ")) {
			t.Errorf("%s: posts %d, stderr %q; want none and a message naming the commit and panel-fix-5", tc.name, posts, stderr.String())
		}
	}
}

// TestRecordDispatchRefusesAReReviewBeforeItsMutationRows pins the kan-1059
// self-review fix, 1b: a live re-review of round 5 is refused while round 4
// holds a fixed Important finding and no fix-mutation row at round 5;
// a row at round 5, a -mutated none row, a Minor-only round and a
// back-fill pass.
func TestRecordDispatchRefusesAReReviewBeforeItsMutationRows(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	const reviewer = `{"key":"panel-4-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-28T02:45:33Z","endedAt":"2026-10-28T02:47:03Z"}`
	var record string
	posts := 0
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			_, _ = w.Write([]byte(record))
			return
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		posts++
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"reviewer","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	finding := func(sev string) string {
		return `{"ref":"F1","round":4,"slot":"primary","severity":"` + sev + `","note":"n","status":"fixed"}`
	}
	for _, tc := range []struct {
		name, finding, mutations string
		extra                    []string
		want                     int
	}{
		{"no mutation row", finding("Important"), ``, nil, 2},
		{"a row at round 5", finding("Important"), `{"id":1,"round":5,"path":"a.go","mutated":"m","test":"t"}`, nil, 0},
		{"only a row at round 4", finding("Important"), `{"id":1,"round":4,"path":"a.go","mutated":"m","test":"t"}`, nil, 2},
		{"a -mutated none row at round 5", finding("Important"), `{"id":1,"round":5,"path":"","mutated":"none","test":"prose only"}`, nil, 0},
		{"only a Minor fixed finding", finding("Minor"), ``, nil, 0},
		{"a back-fill", finding("Important"), ``, []string{"-started-at", "2026-10-28T02:55:00Z"}, 0},
	} {
		record = `{"change":"kan-258","findings":[` + tc.finding + `],"mutations":[` + tc.mutations + `],"dispatches":[` + reviewer + `,{"key":"panel-4-merger","role":"merger","startedAt":"2026-10-01T00:00:00Z","endedAt":"2026-10-01T00:01:00Z","outcome":"completed"}]}`
		posts = 0
		args := append([]string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "reviewer", "-model", "opus", "-key", "panel-5-primary", "-slot", "primary",
			"-session-token", "mf-record-rereview", "-effort", "default"}, tc.extra...)
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		if code != tc.want {
			t.Errorf("%s: exit %d, want %d; stderr:\n%s", tc.name, code, tc.want, stderr.String())
		}
		if tc.want == 2 && (posts != 0 || !strings.Contains(stderr.String(), "round 4") || !strings.Contains(stderr.String(), "flow record mutation -round 5")) {
			t.Errorf("%s: posts %d, stderr %q; want none and a message naming round 4 and the remedy", tc.name, posts, stderr.String())
		}
	}
}

// TestRecordDispatchRefusesAFixOrReReviewBeforeItsRoundsMerger pins the
// kan-1045 self-review fix, F-C (decision merger-enforced-at-begin): a
// panel-fix-1, live or back-filled, and a live reviewer of round 1 are
// refused while round 0 holds no completed panel-0-merger row.
func TestRecordDispatchRefusesAFixOrReReviewBeforeItsRoundsMerger(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	const finding = `{"ref":"F1","round":0,"slot":"primary","severity":"Important","note":"n","status":"open"}`
	const reviewer = `{"key":"panel-0-primary","role":"reviewer","slot":"primary","startedAt":"2026-10-28T02:45:33Z","endedAt":"2026-10-28T02:47:03Z","outcome":"completed"}`
	merger := func(tail string) string {
		return `{"key":"panel-0-merger","role":"merger","startedAt":"2026-10-28T02:48:00Z"` + tail + `}`
	}
	completed := merger(`,"endedAt":"2026-10-28T02:49:00Z","outcome":"completed"`)
	record := func(dispatches ...string) string {
		return `{"change":"kan-258","findings":[` + finding + `],"dispatches":[` + strings.Join(dispatches, ",") + `]}`
	}
	backfill := []string{"-started-at", "2026-10-28T03:00:00Z", "-ended-at", "2026-10-28T03:10:00Z", "-outcome", "completed"}
	for _, tc := range []struct {
		name, record string
		extra        []string
		want         int
	}{
		{"no merger row", record(reviewer), nil, 2},
		{"a completed merger row", record(reviewer, completed), nil, 0},
		{"an open merger row", record(reviewer, merger(``)), nil, 2},
		{"a failed merger row", record(reviewer, merger(`,"endedAt":"2026-10-28T02:49:00Z","outcome":"failed"`)), nil, 2},
		{"a back-fill with no merger row", record(reviewer), backfill, 2},
		{"a failed record read", `not json`, nil, 0},
	} {
		code, stderr, posts := panelFixBegin(t, tc.record, "panel-fix-1", "mf-pf-merger", tc.extra...)
		if code != tc.want {
			t.Errorf("panel-fix, %s: exit %d, want %d; stderr:\n%s", tc.name, code, tc.want, stderr)
		}
		if tc.want == 2 && (posts != 0 || !strings.Contains(stderr, "panel-0-merger")) {
			t.Errorf("panel-fix, %s: posts %d, stderr %q; want none and a message naming panel-0-merger", tc.name, posts, stderr)
		}
	}

	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	var rec string
	posts := 0
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/records/") {
			_, _ = w.Write([]byte(rec))
			return
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		posts++
		_, _ = w.Write([]byte(`{"id":7,"seq":1,"role":"reviewer","model":"opus"}`))
	}))
	t.Cleanup(srv.Close)
	for _, tc := range []struct {
		name, key, record string
		want              int
	}{
		{"a re-review with no merger row", "panel-1-primary", record(reviewer), 2},
		{"a re-review after a completed merger", "panel-1-primary", record(reviewer, completed), 0},
		{"round 0 is never checked", "panel-0-primary", record(), 0},
	} {
		rec = tc.record
		posts = 0
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), []string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "reviewer", "-model", "opus", "-key", tc.key, "-slot", "primary",
			"-session-token", "mf-record-merger", "-effort", "default"}, strings.NewReader(""), &stdout, &stderr)
		if code != tc.want {
			t.Errorf("%s: exit %d, want %d; stderr:\n%s", tc.name, code, tc.want, stderr.String())
		}
		if tc.want == 2 && (posts != 0 || !strings.Contains(stderr.String(), "panel-0-merger")) {
			t.Errorf("%s: posts %d, stderr %q; want none and a message naming panel-0-merger", tc.name, posts, stderr.String())
		}
	}
}

// TestRecordDispatchBeginRefusesASlotOffItsPanelKey pins that a panel
// reviewer row's -slot is its key's slot part (decision
// sr1-slot-matches-key): a joined key under one of its roles, or a
// single-slot key under a joined slot, is refused; a suffixed key under its
// own slot and a non-reviewer panel row are admitted.
func TestRecordDispatchBeginRefusesASlotOffItsPanelKey(t *testing.T) {
	repo := recordRepo(t, "kan-258")
	isolatedStateRoot(t)
	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1}`))
	}))
	defer srv.Close()
	begin := func(role, key, slot string) (int, string) {
		var stdout, stderr bytes.Buffer
		argv := []string{"record", "dispatch", "begin", "-effort", "default", "-agent-id", "none", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", role, "-model", "opus", "-key", key,
			"-session-token", "mf-record-slot-key"}
		if slot != "" {
			argv = append(argv, "-slot", slot)
		}
		code := run(context.Background(), argv, strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String()
	}
	for _, bad := range [][2]string{
		{"panel-3-test-sanity+call-efficiency", "test-sanity"},
		{"panel-4-mutation+failure-modes", "mutation"},
		{"panel-3-extra", "test-sanity+call-efficiency"},
	} {
		contacted = false
		code, stderr := begin("reviewer", bad[0], bad[1])
		if code != 2 {
			t.Errorf("-key %s -slot %s: exit %d, want 2; stderr:\n%s", bad[0], bad[1], code, stderr)
		}
		if !strings.Contains(stderr, bad[0]) || !strings.Contains(stderr, "skills/flow/review-panel.md") {
			t.Errorf("-key %s: stderr does not name the key and review-panel.md:\n%s", bad[0], stderr)
		}
		if contacted {
			t.Errorf("-key %s: a refused row was sent to the daemon", bad[0])
		}
	}
	for _, ok := range [][3]string{
		{"reviewer", "panel-3-test-sanity+call-efficiency", "test-sanity+call-efficiency"},
		{"reviewer", "panel-2-primary", "primary"},
		{"reviewer", "panel-1-primary+principles-cite", "primary+principles"},
		{"reviewer", "panel-0-primary-retry", "primary"},
		{"reviewer", "panel-9-primary-agents", "primary"},
		{"merger", "panel-3-merger", ""},
	} {
		if code, stderr := begin(ok[0], ok[1], ok[2]); code != 0 {
			t.Errorf("-role %s -key %s -slot %s: exit %d, want 0; stderr:\n%s", ok[0], ok[1], ok[2], code, stderr)
		}
	}
}

// TestRecordDispatchBeginPrintsTheLaunchLines pins decision
// sr3-begin-prints-launch-lines: a live begin prints, after its recorded:
// line, the row's description opening and, where the row decides it, its
// Role line; a back-fill and an inline row print only the recorded: line.
func TestRecordDispatchBeginPrintsTheLaunchLines(t *testing.T) {
	repo := recordRepo(t, "kan-258", "kan-259")
	isolatedStateRoot(t)
	plan := "# plan\n\n- [ ] 1. one\n- [ ] 2. two\n  - [ ] 3. nested step, not a task\n- [ ] 3. three\n"
	if err := os.WriteFile(filepath.Join(repo, "spectre", "changes", "kan-258", "tasks.md"), []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":7}`))
	}))
	defer srv.Close()
	begin := func(change, role, key string, extra ...string) (int, string, string) {
		var stdout, stderr bytes.Buffer
		argv := append([]string{"record", "dispatch", "begin", "-effort", "default", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", change, "-role", role, "-model", "opus", "-key", key,
			"-session-token", "mf-record-launch-lines"}, extra...)
		code := run(context.Background(), argv, strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
	for _, c := range []struct {
		change, role, key string
		extra             []string
		want              string
	}{
		{"kan-258", "implementer", "task-2-implementer", nil, "recorded: dispatch 7\ndescription: Task 2/3\nRole: implementer\n"},
		{"kan-258", "reviewer", "task-1+3-reviewer", nil, "recorded: dispatch 7\ndescription: Tasks 1+3/3\nRole: gated-reviewer\n"},
		{"kan-258", "reviewer", "task-1-reviewer-fix-1", nil, "recorded: dispatch 7\ndescription: Task 1/3\nRole: gated-re-reviewer\n"},
		{"kan-259", "reviewer", "panel-2-primary", []string{"-slot", "primary"}, "recorded: dispatch 7\ndescription: panel-2-primary\n"},
		{"kan-258", "merger", "panel-2-merger", nil, "recorded: dispatch 7\ndescription: panel-2-merger\nRole: panel-merger\n"},
		{"kan-258", "panel-fix", "panel-fix-3", nil, "recorded: dispatch 7\ndescription: panel-fix-3\nRole: panel-fix\n"},
		{"kan-258", "verifier", "visual-verify-2", []string{"-diff-base", "abc123"}, "recorded: dispatch 7\ndescription: visual-verify-2\nRole: visual-verifier\n"},
		{"kan-258", "planner", "tooling-analysis-1", nil, "recorded: dispatch 7\ndescription: tooling-analysis-1\nRole: tooling-analyst\n"},
		{"kan-258", "implementer", "task-1-implementer", []string{"-started-at", "2026-10-30T09:00:00Z"}, "recorded: dispatch 7\n"},
		{"kan-258", "verifier", "verify", []string{"-agent-id", "inline"}, "recorded: dispatch 7\n"},
		{"kan-259", "implementer", "task-1-implementer", nil, "recorded: dispatch 7\nRole: implementer\n"},
	} {
		code, stdout, stderr := begin(c.change, c.role, c.key, c.extra...)
		if code != 0 {
			t.Errorf("-key %s: exit %d, want 0; stderr:\n%s", c.key, code, stderr)
			continue
		}
		if stdout != c.want {
			t.Errorf("-key %s: stdout = %q, want %q", c.key, stdout, c.want)
		}
	}
}

func TestRecordNextRoundSkipsGatedRows(t *testing.T) {
	repo := recordRepo(t, "demo")
	isolatedStateRoot(t)
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"change":"demo","dispatches":[],"passes":[],"mutations":[],
		  "findings":[{"ref":"F1","round":0,"slot":"gated-reviewer","severity":"Important","note":"n","status":"open"}]}`))
	}))
	defer srv.Close()
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "next-round", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 || strings.TrimSpace(stdout.String()) != "0" {
		t.Fatalf("exit %d stdout %q, want 0 and 0; stderr:\n%s", code, stdout.String(), stderr.String())
	}
}
