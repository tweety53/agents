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

// TestSettingsCmd_Get asserts `myflow settings get` prints the store's
// settings record as one line of JSON and exits 0 -- the same shape
// `state get` already prints for a change record (state_test.go), applied
// to task 2's GET /api/v1/settings instead.
func TestSettingsCmd_Get(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/settings" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"defaultModel":"opus","reviewers":["primary","principles","code-review-low"]}`))
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
		DefaultModel string   `json:"defaultModel"`
		Reviewers    []string `json:"reviewers"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode stdout %q: %v", stdout.String(), err)
	}
	if got.DefaultModel != "opus" {
		t.Errorf("defaultModel = %q, want %q", got.DefaultModel, "opus")
	}
	want := []string{"primary", "principles", "code-review-low"}
	if len(got.Reviewers) != len(want) {
		t.Fatalf("reviewers = %v, want %v", got.Reviewers, want)
	}
	for i, r := range want {
		if got.Reviewers[i] != r {
			t.Errorf("reviewers[%d] = %q, want %q", i, got.Reviewers[i], r)
		}
	}
}

// TestSettingsCmd_Set_PrintsRejectionReason asserts `myflow settings set`
// against an invalid -model/-reviewers value surfaces the API's 400
// rejection reason on stderr and exits non-zero -- unlike `state`/`stage`'s
// never-block-on-store-failure pattern, this is a caller mistake with no
// fallback value to record, per the task's own instruction.
func TestSettingsCmd_Set_PrintsRejectionReason(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/settings" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"store: invalid model: \"gpt-5\""}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"settings", "set", "-addr", srv.URL, "-timeout", "2s",
			"-model", "gpt-5", "-reviewers", "primary,principles,code-review-low"},
		strings.NewReader(""), &stdout, &stderr)

	if code == 0 {
		t.Fatalf("exit code = 0, want non-zero")
	}
	if !strings.Contains(stderr.String(), "gpt-5") {
		t.Errorf("stderr = %q, want it to name the rejected value %q", stderr.String(), "gpt-5")
	}
}

// TestSettingsCmd_Set_Valid asserts a well-formed `settings set` writes
// through to the store and prints the settings the store echoes back,
// exiting 0 -- the round-trip TestSettingsCmd_Get proves for a read,
// proved here for a write.
func TestSettingsCmd_Set_Valid(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/settings" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			DefaultModel string   `json:"defaultModel"`
			Reviewers    []string `json:"reviewers"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.DefaultModel != "sonnet" {
			t.Errorf("request defaultModel = %q, want %q", body.DefaultModel, "sonnet")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"settings", "set", "-addr", srv.URL, "-timeout", "2s",
			"-model", "sonnet", "-reviewers", "primary,principles,code-review-low"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "sonnet") {
		t.Errorf("stdout = %q, want it to echo back the written settings", stdout.String())
	}
}
