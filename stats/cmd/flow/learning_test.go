package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func runLearningArgs(args ...string) (int, string, string) {
	var stdout, stderr strings.Builder
	code := run(context.Background(), append([]string{"learning"}, args...), strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// TestLearningAdd_Records asserts `flow learning add` posts the topic, note,
// change and the project key resolved from -C, and exits 0.
func TestLearningAdd_Records(t *testing.T) {
	repo := gitRepo(t)
	var got map[string]string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/learning" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(got)
	}))
	defer srv.Close()

	code, _, stderr := runLearningArgs("add", "cqrs", "-addr", srv.URL, "-timeout", "2s", "-C", repo, "-note", "split reads from writes", "-change", "demo")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr)
	}
	if got["topic"] != "cqrs" || got["note"] != "split reads from writes" || got["change"] != "demo" || got["projectKey"] == "" {
		t.Errorf("request body = %v, want topic, note, change and a project key", got)
	}
}

// TestLearningAdd_ExhaustedExits1 asserts a 409 exits 1 naming the cap.
func TestLearningAdd_ExhaustedExits1(t *testing.T) {
	repo := gitRepo(t)
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"store: learning topic exhausted: \"cqrs\" shown 3 times"}`))
	}))
	defer srv.Close()

	code, _, stderr := runLearningArgs("add", "cqrs", "-addr", srv.URL, "-timeout", "2s", "-C", repo, "-note", "n", "-change", "demo")
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "three times") {
		t.Errorf("stderr = %q, want it to name the three-times cap", stderr)
	}
}

// TestLearningList_PrintsCounts asserts `flow learning list` prints the
// store's topic array as one line of JSON.
func TestLearningList_PrintsCounts(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/learning" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"topic":"cqrs","count":2,"lastNote":"n"}]`))
	}))
	defer srv.Close()

	code, stdout, stderr := runLearningArgs("list", "-addr", srv.URL, "-timeout", "2s")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr)
	}
	if want := `[{"topic":"cqrs","count":2,"lastNote":"n"}]` + "\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}
