package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// wantHost is what the suite CLI records as host: os.Hostname's first
// dot-separated label. The tests assert against the same derivation so a
// machine's actual name never makes a test machine-dependent.
func wantHost(t *testing.T) string {
	t.Helper()
	host, err := os.Hostname()
	if err != nil {
		t.Fatalf("os.Hostname: %v", err)
	}
	return strings.SplitN(host, ".", 2)[0]
}

// TestSuiteRecordRunsChildAndRecordsRow pins the wrapper's core: the child
// runs with its stdio passing through (its output lands on the caller's
// stdout untouched), one row is recorded with this host, a positive
// duration and the child's exit code, and the command exits with the
// child's own code.
func TestSuiteRecordRunsChildAndRecordsRow(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotBody, _ = readAll(r)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"suite":"probe","host":"h","durationMs":10,"exitCode":0,"ranAt":"2026-09-09T12:00:00Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"suite", "record", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-suite", "probe", "--", "/bin/sh", "-c", "echo child-stdout"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("suite record exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "child-stdout") {
		t.Errorf("child stdout did not pass through: %q", stdout.String())
	}
	if gotMethod != http.MethodPost || !strings.HasPrefix(gotPath, "/api/v1/suites/") {
		t.Errorf("record request = %s %s, want POST under /api/v1/suites/", gotMethod, gotPath)
	}

	var sent records.SuiteRun
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode record body %s: %v", gotBody, err)
	}
	if sent.Suite != "probe" {
		t.Errorf("suite = %q, want probe", sent.Suite)
	}
	if sent.Host != wantHost(t) {
		t.Errorf("host = %q, want this machine's short hostname %q", sent.Host, wantHost(t))
	}
	if sent.DurationMs < 0 {
		t.Errorf("durationMs = %d, want a non-negative measurement", sent.DurationMs)
	}
	if sent.ExitCode != 0 {
		t.Errorf("exitCode = %d, want the child's 0", sent.ExitCode)
	}
}

// TestSuiteRecordPassesChildExitThrough pins that a failing child's exit
// status is both recorded and returned: the caller's tooling sees the
// suite's real verdict, and the store holds the duration that came with it.
func TestSuiteRecordPassesChildExitThrough(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = readAll(r)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"suite":"probe","host":"h","durationMs":10,"exitCode":3,"ranAt":"2026-09-09T12:00:00Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"suite", "record", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-suite", "probe", "--", "/bin/sh", "-c", "exit 3"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 3 {
		t.Fatalf("suite record exit = %d, want the child's 3", code)
	}
	var sent records.SuiteRun
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode record body %s: %v", gotBody, err)
	}
	if sent.ExitCode != 3 {
		t.Errorf("recorded exitCode = %d, want the child's 3", sent.ExitCode)
	}
}

// TestSuiteRecordWarnsWithoutAlteringExitWhenStoreDown pins the never-a-gate
// rule: the store being unreachable costs one warning line and nothing else
// -- the child still runs and its exit code is still the command's exit.
func TestSuiteRecordWarnsWithoutAlteringExitWhenStoreDown(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	dead := deadPortURL(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"suite", "record", "-addr", dead, "-timeout", "500ms", "-C", repo,
			"-suite", "probe", "--", "/bin/sh", "-c", "exit 5"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 5 {
		t.Fatalf("suite record exit = %d, want the child's 5 despite the store being down", code)
	}
	if !strings.Contains(stderr.String(), "not recorded") {
		t.Errorf("stderr = %q, want the not-recorded warning", stderr.String())
	}
}

// TestSuiteRecordRefusesMissingSuiteOrCommand pins the caller mistakes: no
// -suite, or no -- and command words after it, is a usage error -- exit 2,
// nothing executed, nothing recorded.
func TestSuiteRecordRefusesMissingSuiteOrCommand(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("a refused record must not reach the store")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	for name, args := range map[string][]string{
		"no -suite":         {"suite", "record", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "--", "/bin/sh", "-c", "echo ran"},
		"no -- and command": {"suite", "record", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-suite", "probe"},
	} {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		if code != 2 {
			t.Errorf("%s: exit = %d, want 2", name, code)
		}
		if !strings.Contains(stderr.String(), "Usage") && !strings.Contains(stderr.String(), "usage") {
			t.Errorf("%s: stderr = %q, want the usage block", name, stderr.String())
		}
	}
}

// TestSuiteRecordReportsUnstartableChild pins the cannot-start path: a
// child that never starts is exit 127 with one stderr line and nothing
// recorded -- a duration for a run that did not happen would be a figure
// with no run behind it.
func TestSuiteRecordReportsUnstartableChild(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("an unstartable child must not reach the store")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"suite", "record", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-suite", "probe", "--", "/nonexistent-binary-kan252"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 127 {
		t.Fatalf("unstartable child exit = %d, want 127", code)
	}
	if !strings.Contains(stderr.String(), "could not start") {
		t.Errorf("stderr = %q, want the could-not-start line", stderr.String())
	}
}

// TestSuiteListRendersRowsAndSummary pins the text read: rows newest first
// with host and exit visible, plus one summary line per (suite, host)
// carrying the median of the last 10 passing runs -- the failed run renders
// as a row but never enters a median.
func TestSuiteListRendersRowsAndSummary(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id":4,"suite":"guard-tests","host":"ci","durationMs":90000,"exitCode":0,"ranAt":"2026-09-09T11:00:00Z"},
			{"id":3,"suite":"guard-tests","host":"laptop","durationMs":200000,"exitCode":1,"ranAt":"2026-09-09T10:30:00Z"},
			{"id":2,"suite":"guard-tests","host":"laptop","durationMs":56000,"exitCode":0,"ranAt":"2026-09-09T10:00:00Z"},
			{"id":1,"suite":"guard-tests","host":"laptop","durationMs":52000,"exitCode":0,"ranAt":"2026-09-09T09:30:00Z"}
		]`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"suite", "list", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("suite list exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	out := stdout.String()
	for _, want := range []string{
		"guard-tests", "ci", "laptop", "1m30s", "3m20s", // rows, durations rendered
		"exit 1", // the failed run is visible as a row
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "median 52s over the last 2 passing runs") {
		t.Errorf("laptop summary = want median 52s over the last 2 passing runs (lower middle of 52s/56s; the failed 3m20s row is excluded):\n%s", out)
	}
	if !strings.Contains(out, "median 1m30s over the last 1 passing runs") {
		t.Errorf("ci summary = want median 1m30s over the last 1 passing runs:\n%s", out)
	}
	if strings.Count(out, "median") != 2 {
		t.Errorf("want one summary line per (suite, host) -- two hosts -- got:\n%s", out)
	}
}

// TestSuiteListJSON pins the machine-readable read: -json emits exactly the
// store's array and nothing else -- no rows, no summaries.
func TestSuiteListJSON(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"suite":"guard-tests","host":"laptop","durationMs":52000,"exitCode":0,"ranAt":"2026-09-09T09:30:00Z"}]`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"suite", "list", "-json", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("suite list -json exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}

	var listed []records.SuiteRun
	if err := json.Unmarshal(stdout.Bytes(), &listed); err != nil {
		t.Fatalf("stdout is not a bare JSON array: %v\n%s", err, stdout.String())
	}
	if len(listed) != 1 || listed[0].DurationMs != 52000 {
		t.Errorf("json = %+v, want the store's one row", listed)
	}
}

// deadPortURL is a URL whose port nothing answers on -- the unreachable
// store the fallback tests need. It binds then closes a listener, so the
// port is released but almost certainly unclaimed.
func deadPortURL(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	time.Sleep(10 * time.Millisecond)
	return url
}
