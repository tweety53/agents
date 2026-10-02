package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRunLessonResolvePrintsAnswer drives `flow lesson resolve` end to
// end: the topic rides as the query parameter and the rendered answer is
// printed verbatim on stdout with exit 0.
func TestRunLessonResolvePrintsAnswer(t *testing.T) {
	isolatedStateRoot(t)

	var gotPath, gotQuery string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Lessons: flake-bisection\n\nfound: 1 briefs, 0 narrative mentions\n"))
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"lesson", "resolve", "-addr", srv.URL, "-timeout", "500ms", "-topic", "flake-bisection"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if gotPath != "/api/v1/lessons/resolve" {
		t.Errorf("request path = %s", gotPath)
	}
	if gotQuery != "topic=flake-bisection" {
		t.Errorf("request query = %q, want the topic carried", gotQuery)
	}
	if !strings.Contains(stdout.String(), "# Lessons: flake-bisection") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

// TestRunLessonResolveRequiresTopic pins the caller mistake: no -topic is
// exit 2 and a stderr line, never a request.
func TestRunLessonResolveRequiresTopic(t *testing.T) {
	isolatedStateRoot(t)

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"lesson", "resolve", "-addr", "http://127.0.0.1:1"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "-topic is required") {
		t.Errorf("stderr = %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want nothing", stdout.String())
	}
}

// TestRunLessonResolveRefusesPositional pins the same rule for a stray
// positional: a mistyped flag is reported as one, never dropped.
func TestRunLessonResolveRefusesPositional(t *testing.T) {
	isolatedStateRoot(t)

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"lesson", "resolve", "-addr", "http://127.0.0.1:1", "-topic", "t", "extra"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "positional") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

// TestRunLessonResolveUnreachableStoreExitsNonZero carries the read
// contract through the CLI: a store that cannot answer is reported to
// stderr and exits non-zero -- never a partial answer on stdout a caller
// could mistake for the workspace's own.
func TestRunLessonResolveUnreachableStoreExitsNonZero(t *testing.T) {
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"lesson", "resolve", "-addr", srv.URL, "-timeout", "500ms", "-topic", "flake"},
		strings.NewReader(""), &stdout, &stderr)
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero (stderr: %s)", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want nothing on a failed read", stdout.String())
	}
	if !strings.Contains(stderr.String(), "lesson resolve") {
		t.Errorf("stderr = %q, want the failure named", stderr.String())
	}
}
