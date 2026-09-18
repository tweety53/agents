package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// jiraFlowdFake is a stand-in for flowd's POST /api/v1/jira/transition,
// answering from wantStatus/wantBody and recording the last request body,
// so the tests assert what the CLI actually sent.
type jiraFlowdFake struct {
	t          *testing.T
	wantStatus int
	wantBody   string
	lastBody   string
}

func (f *jiraFlowdFake) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Flow-Daemon", "flowd/1")
		w.Header().Set("Content-Type", "application/json")
		body, err := io.ReadAll(req.Body)
		if err != nil {
			f.t.Fatalf("read request body: %v", err)
		}
		f.lastBody = string(body)
		w.WriteHeader(f.wantStatus)
		_, _ = w.Write([]byte(f.wantBody))
	}
}

func runJiraTransitionForTest(t *testing.T, ts *httptest.Server, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	code := run(context.Background(), append([]string{"jira", "transition", "-addr", ts.URL}, args...), strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestRunJiraTransition(t *testing.T) {
	fake := &jiraFlowdFake{t: t, wantStatus: http.StatusOK,
		wantBody: `{"key":"KAN-571","status":"In Review","moved":true}`}
	ts := httptest.NewServer(fake.handler())
	defer ts.Close()

	code, stdout, stderr := runJiraTransitionForTest(t, ts, "KAN-571", "In Review")
	if code != 0 {
		t.Fatalf("exit = %d, stderr %s, want 0", code, stderr)
	}
	if stdout != "Jira: KAN-571 → In Review\n" {
		t.Errorf("stdout = %q, want the moved report", stdout)
	}
	if !strings.Contains(fake.lastBody, `"target":"In Review"`) {
		t.Errorf("request body %s does not carry the target", fake.lastBody)
	}
}

func TestRunJiraTransitionAlreadyThere(t *testing.T) {
	fake := &jiraFlowdFake{t: t, wantStatus: http.StatusOK,
		wantBody: `{"key":"KAN-571","status":"Done","moved":false}`}
	ts := httptest.NewServer(fake.handler())
	defer ts.Close()

	code, stdout, stderr := runJiraTransitionForTest(t, ts, "KAN-571", "In Review")
	if code != 0 {
		t.Fatalf("exit = %d, stderr %s, want 0", code, stderr)
	}
	if stdout != "Jira: KAN-571 already Done (no transition)\n" {
		t.Errorf("stdout = %q, want the no-op report", stdout)
	}
}

func TestRunJiraTransitionRefused(t *testing.T) {
	fake := &jiraFlowdFake{t: t, wantStatus: http.StatusUnprocessableEntity,
		wantBody: `{"error":"jira: current status matches no name mapped onto the four ordered positions: \"Blocked\""}`}
	ts := httptest.NewServer(fake.handler())
	defer ts.Close()

	code, stdout, stderr := runJiraTransitionForTest(t, ts, "KAN-571", "Done")
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "Blocked") {
		t.Errorf("stderr %q does not carry the daemon's reason", stderr)
	}
}

// A 400 from flowd -- the bad-target case among them -- is the caller-
// mistake exit class: 2, exactly as jiraUsage documents.
func TestRunJiraTransitionCallerMistake(t *testing.T) {
	fake := &jiraFlowdFake{t: t, wantStatus: http.StatusBadRequest,
		wantBody: `{"error":"jira: \"Blocked\" matches no pipeline position (accepted: To Do, In Progress, In Review, Done)"}`}
	ts := httptest.NewServer(fake.handler())
	defer ts.Close()

	code, stdout, stderr := runJiraTransitionForTest(t, ts, "KAN-571", "Blocked")
	if code != 2 {
		t.Fatalf("exit = %d, stderr %s, want 2", code, stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "matches no pipeline position") {
		t.Errorf("stderr %q does not carry the daemon's reason", stderr)
	}
}

func TestRunJiraTransitionUsage(t *testing.T) {
	var stdout, stderr strings.Builder
	if code := run(context.Background(), []string{"jira"}, strings.NewReader(""), &stdout, &stderr); code != 2 {
		t.Errorf("bare `flow jira` exit = %d, want 2", code)
	}
	if code := run(context.Background(), []string{"jira", "transition", "KAN-571"}, strings.NewReader(""), &stdout, &stderr); code != 2 {
		t.Errorf("missing target exit = %d, want 2", code)
	}
	if code := run(context.Background(), []string{"jira", "wat"}, strings.NewReader(""), &stdout, &stderr); code != 2 {
		t.Errorf("unknown subcommand exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "four-position") {
		t.Errorf("usage output %q does not describe the transition", stderr.String())
	}
}
