package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSettingsCmd_Get asserts `flow settings get` prints the store's
// settings record as one line of JSON and exits 0 -- the same shape
// `state get` already prints for a change record (state_test.go), applied
// to task 2's GET /api/v1/settings instead.
func TestSettingsCmd_Get(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/settings" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"reviewers":["primary","principles","mutation"]}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"settings", "get", "-addr", srv.URL, "-timeout", "2s"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr.String())
	}
	var got struct {
		Reviewers []string `json:"reviewers"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout %q: %v", stdout.String(), err)
	}
	want := []string{"primary", "principles", "mutation"}
	if len(got.Reviewers) != len(want) {
		t.Fatalf("reviewers = %v, want %v", got.Reviewers, want)
	}
	for i, r := range want {
		if got.Reviewers[i] != r {
			t.Errorf("reviewers[%d] = %q, want %q", i, got.Reviewers[i], r)
		}
	}
}

// TestSettingsCmd_Set_PrintsRejectionReason asserts `flow settings set`
// against an invalid -reviewers value surfaces the API's 400
// rejection reason on stderr and exits non-zero -- unlike `state`/`stage`'s
// never-block-on-store-failure pattern, this is a caller mistake with no
// fallback value to record, per the task's own instruction.
func TestSettingsCmd_Set_PrintsRejectionReason(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/settings" {
			_, _ = w.Write([]byte(`{"reviewers":["primary"],"learning":true}`))
			return
		}
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/settings" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"store: invalid reviewer: \"not-a-slot\""}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"settings", "set", "-addr", srv.URL, "-timeout", "2s",
			"-reviewers", "primary,not-a-slot"},
		strings.NewReader(""), &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "not-a-slot") {
		t.Errorf("stderr = %q, want it to name the rejected value %q", stderr.String(), "not-a-slot")
	}
}

// TestSettingsCmd_Set_Valid asserts a well-formed `settings set` writes
// through to the store and prints the settings the store echoes back,
// exiting 0 -- the round-trip TestSettingsCmd_Get proves for a read,
// proved here for a write.
func TestSettingsCmd_Set_Valid(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/settings" {
			_, _ = w.Write([]byte(`{"reviewers":["primary"],"learning":true}`))
			return
		}
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/settings" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			Reviewers []string `json:"reviewers"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if strings.Join(body.Reviewers, ",") != "primary,principles,mutation" {
			t.Errorf("request reviewers = %v, want primary,principles,mutation", body.Reviewers)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"settings", "set", "-addr", srv.URL, "-timeout", "2s",
			"-reviewers", "primary,principles,mutation"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "mutation") {
		t.Errorf("stdout = %q, want it to echo back the written settings", stdout.String())
	}
}

// TestSettingsCmd_Set_RejectsModelFlag asserts the retired -model flag is
// a usage error (exit 2, stderr naming it) and never reaches the store.
func TestSettingsCmd_Set_RejectsModelFlag(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"settings", "set", "-addr", srv.URL, "-timeout", "2s",
			"-model", "opus", "-reviewers", "primary"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "-model") {
		t.Errorf("stderr = %q, want it to name -model", stderr.String())
	}
}

// TestSettingsCmd_ModelsIsUnknown asserts the retired `settings models`
// subcommand is refused as an unknown settings command.
func TestSettingsCmd_ModelsIsUnknown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"settings", "models"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), `unknown settings command "models"`) {
		t.Errorf("stderr = %q, want it to name the unknown command", stderr.String())
	}
}

// settingsServer fakes GET/PUT /api/v1/settings over one in-memory record,
// so `settings set`'s read-merge-write is exercised end to end.
func settingsServer(t *testing.T, stored *map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Method == http.MethodPut {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			*stored = body
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(*stored)
	}))
}

func runSettingsArgs(addr string, extra ...string) (int, string) {
	var stdout, stderr bytes.Buffer
	args := append([]string{"settings", "set", "-addr", addr, "-timeout", "2s"}, extra...)
	code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
	return code, stderr.String()
}

// TestSettingsSet_LearningKeepsReviewers asserts `set -learning off` keeps
// the stored reviewer list (settings-set-merges).
func TestSettingsSet_LearningKeepsReviewers(t *testing.T) {
	stored := map[string]any{"reviewers": []any{"primary", "mutation"}, "learning": true}
	srv := settingsServer(t, &stored)
	defer srv.Close()

	if code, stderr := runSettingsArgs(srv.URL, "-learning", "off"); code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr)
	}
	if got, _ := json.Marshal(stored); string(got) != `{"learning":false,"reviewers":["primary","mutation"]}` {
		t.Errorf("stored = %s, want reviewers kept and learning false", got)
	}
}

// TestSettingsSet_ReviewersKeepLearning asserts `set -reviewers` keeps a
// stored learning off (settings-set-merges).
func TestSettingsSet_ReviewersKeepLearning(t *testing.T) {
	stored := map[string]any{"reviewers": []any{"primary", "mutation"}, "learning": false}
	srv := settingsServer(t, &stored)
	defer srv.Close()

	if code, stderr := runSettingsArgs(srv.URL, "-reviewers", "primary"); code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr)
	}
	if got, _ := json.Marshal(stored); string(got) != `{"learning":false,"reviewers":["primary"]}` {
		t.Errorf("stored = %s, want learning kept off", got)
	}
}

// TestSettingsSet_NeedsAFlag asserts `set` with neither flag is a usage
// error.
func TestSettingsSet_NeedsAFlag(t *testing.T) {
	if code, _ := runSettingsArgs("http://127.0.0.1:1"); code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
}

// TestSettingsSet_LearningRejectsBadValue asserts -learning takes on or
// off only.
func TestSettingsSet_LearningRejectsBadValue(t *testing.T) {
	if code, _ := runSettingsArgs("http://127.0.0.1:1", "-learning", "maybe"); code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
}
