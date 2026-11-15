package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
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
// runs with its output passing through (a plain line lands on the caller's
// stdout untouched), one row is recorded with this host, a positive
// duration and the child's exit code, and the command exits with the
// child's own code.
// smoke: core — flow suite record runs its child and records one row
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
	if len(sent.Tests) != 0 {
		t.Errorf("tests = %v, want none: a plain-text child carries no go test -json event", sent.Tests)
	}
}

// goTestJSONChild writes a child stream mixing plain lines, test2json events
// and a non-event JSON object, ending in a line with no newline, and returns
// the command words that cat it.
func goTestJSONChild(t *testing.T) []string {
	t.Helper()
	stream := strings.Join([]string{
		`plain before`,
		`{"Action":"run","Package":"ex/a","Test":"TestA"}`,
		`{"Action":"output","Package":"ex/a","Test":"TestA","Output":"=== RUN   TestA\n"}`,
		`{"Action":"pass","Package":"ex/a","Test":"TestA","Elapsed":1.5}`,
		`{"Action":"skip","Package":"ex/a","Test":"TestS","Elapsed":0}`,
		`{"x":1}`,
		`tail`,
	}, "\n")
	path := filepath.Join(t.TempDir(), "stream")
	if err := os.WriteFile(path, []byte(stream), 0o600); err != nil {
		t.Fatalf("write child stream: %v", err)
	}
	return []string{"cat", path}
}

// TestSuiteRecordDecodesGoTestJSON pins the line-by-line decode (decisions
// json-stream-line-detection, store-every-test): a go test -json event prints
// its Output alone, every other line passes through unchanged -- a last line
// without a newline included -- and each non-skipped top-level test's cost is
// sent with the run.
func TestSuiteRecordDecodesGoTestJSON(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = readAll(r)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"suite":"probe","host":"h","durationMs":10,"exitCode":0,"ranAt":"2026-09-09T12:00:00Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	args := append([]string{"suite", "record", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-suite", "probe", "--"},
		goTestJSONChild(t)...)
	if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("suite record exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if want := "plain before\n=== RUN   TestA\n{\"x\":1}\ntail"; stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
	var sent records.SuiteRun
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode record body %s: %v", gotBody, err)
	}
	if got := fmt.Sprint(sent.Tests); got != "[{ex/a TestA 1500}]" {
		t.Errorf("tests = %s, want [{ex/a TestA 1500}]", got)
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
	if !strings.Contains(stderr.String(), "wrote local journal") {
		t.Errorf("stderr = %q, want the journal warning", stderr.String())
	}
}

// TestSuiteRecordJournalsWhenStoreDown pins kan-1040's lost suite row: a run
// the store could not take is journalled for replay, stamped with its start.
func TestSuiteRecordJournalsWhenStoreDown(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	dead := deadPortURL(t)

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"suite", "record", "-addr", dead, "-timeout", "500ms", "-C", repo,
			"-suite", "probe", "--", "/bin/sh", "-c", "exit 5"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 5 {
		t.Fatalf("suite record exit = %d, want 5", code)
	}
	key, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(key, "_suite-runs") + ".record")
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("journal entries = %d, want 1", len(entries))
	}
	var body struct {
		Kind    string           `json:"kind"`
		Request records.SuiteRun `json:"request"`
	}
	if err := json.Unmarshal(entries[0].Body, &body); err != nil {
		t.Fatalf("decode entry: %v", err)
	}
	if body.Kind != "suite" || body.Request.Suite != "probe" || body.Request.ExitCode != 5 || body.Request.RanAt.IsZero() {
		t.Errorf("entry = %+v, want kind suite, suite probe, exit 5, a ranAt", body)
	}
}

// TestSuiteRecordJournalsTestsWhenStoreDown pins that a store outage keeps
// the per-test costs (decision per-test-costs-in-flowd): the journalled entry
// carries them for replay.
func TestSuiteRecordJournalsTestsWhenStoreDown(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)
	dead := deadPortURL(t)

	var stdout, stderr bytes.Buffer
	args := append([]string{"suite", "record", "-addr", dead, "-timeout", "500ms", "-C", repo, "-suite", "probe", "--"},
		goTestJSONChild(t)...)
	if code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("suite record exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	key, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	entries, err := fallback.ReadJournalEntries(fallback.JournalFilePath(key, "_suite-runs") + ".record")
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("journal entries = %d, want 1", len(entries))
	}
	var body struct {
		Kind    string           `json:"kind"`
		Request records.SuiteRun `json:"request"`
	}
	if err := json.Unmarshal(entries[0].Body, &body); err != nil {
		t.Fatalf("decode entry: %v", err)
	}
	if got := fmt.Sprint(body.Request.Tests); body.Kind != "suite" || got != "[{ex/a TestA 1500}]" {
		t.Errorf("entry kind %q tests %s, want kind suite, tests [{ex/a TestA 1500}]", body.Kind, got)
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
// smoke: core — the rows and medians the default list is watched by
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

// TestSuiteListTestsReadsTheNewestRunWithTests pins the per-test read
// (decision suite-tests-read): one header line in the list's row format, then
// the run's tests heaviest first as cost seconds, package and test; -limit
// rides the request, -json prints the response alone, and -tests with -suite
// is a usage error.
func TestSuiteListTestsReadsTheNewestRunWithTests(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	const body = `{"id":7,"suite":"guard-tests","host":"h","durationMs":61000,"exitCode":0,"ranAt":"2026-09-09T12:00:00Z",` +
		`"tests":[{"package":"ex/a","test":"TestSlow","costMs":2500},{"package":"ex/b","test":"TestFast","costMs":40}]}`
	var gotPath, gotQuery string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	key, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	list := func(extra ...string) (int, string, string) {
		var stdout, stderr bytes.Buffer
		args := append([]string{"suite", "list", "-addr", srv.URL, "-timeout", "500ms", "-C", repo}, extra...)
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}

	code, out, errOut := list("-tests", "guard-tests")
	if code != 0 {
		t.Fatalf("suite list -tests exit = %d, want 0; stderr:\n%s", code, errOut)
	}
	if want := "guard-tests\t2026-09-09T12:00:00Z\t1m1s\th\texit 0\n2.50\tex/a\tTestSlow\n0.04\tex/b\tTestFast\n"; out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
	if gotPath != "/api/v1/suites/"+key+"/tests" || gotQuery != "limit=20&suite=guard-tests" {
		t.Errorf("request = %s?%s, want /api/v1/suites/%s/tests with suite=guard-tests&limit=20", gotPath, gotQuery, key)
	}

	if code, _, errOut := list("-tests", "guard-tests", "-limit", "1"); code != 0 || gotQuery != "limit=1&suite=guard-tests" {
		t.Errorf("-limit 1: exit %d query %q, want 0 and limit=1; stderr:\n%s", code, gotQuery, errOut)
	}
	if code, out, _ := list("-tests", "guard-tests", "-json"); code != 0 || out != body+"\n" {
		t.Errorf("-json: exit %d stdout %q, want 0 and the response alone", code, out)
	}
	if code, _, errOut := list("-tests", "guard-tests", "-suite", "x"); code != 2 || !strings.Contains(errOut, "usage") {
		t.Errorf("-tests with -suite: exit %d stderr %q, want 2 and the usage block", code, errOut)
	}
}

// TestSuiteListTestsReportsNoRunAndAnUnreachableStore pins the read's two
// edges: no run with recorded costs is a plain line and exit 0, and an
// unreachable store is a read failure, exit 1.
func TestSuiteListTestsReportsNoRunAndAnUnreachableStore(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`null`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	if code := run(context.Background(),
		[]string{"suite", "list", "-tests", "guard-tests", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr); code != 0 || stdout.String() != "no test costs recorded for guard-tests\n" {
		t.Errorf("no run: exit %d stdout %q, want 0 and the no-costs line; stderr:\n%s", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(),
		[]string{"suite", "list", "-tests", "guard-tests", "-addr", deadPortURL(t), "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "flow: suite list:") {
		t.Errorf("dead store: exit %d stderr %q, want 1 and flow: suite list:", code, stderr.String())
	}
}

// TestSuiteResolvesTheRecordsAddress pins that both suite verbs default
// -addr to FLOW_RECORDS_ADDR over FLOW_ADDR: a worktree's FLOW_ADDR names
// its own short-lived daemon, and runtime rows recorded there would never
// reach the cross-run median the persistent store serves.
func TestSuiteResolvesTheRecordsAddress(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var recorded, listed bool
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			recorded = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":1,"suite":"probe","host":"h","durationMs":10,"exitCode":0,"ranAt":"2026-09-09T12:00:00Z"}`))
			return
		}
		listed = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	t.Setenv("FLOW_RECORDS_ADDR", srv.URL)
	t.Setenv("FLOW_ADDR", deadPortURL(t))

	var stdout, stderr bytes.Buffer
	if code := run(context.Background(),
		[]string{"suite", "record", "-timeout", "500ms", "-C", repo, "-suite", "probe", "--", "/bin/sh", "-c", "true"},
		strings.NewReader(""), &stdout, &stderr); code != 0 || !recorded {
		t.Fatalf("suite record exit = %d, recorded at FLOW_RECORDS_ADDR = %v; stderr:\n%s", code, recorded, stderr.String())
	}
	if code := run(context.Background(),
		[]string{"suite", "list", "-json", "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr); code != 0 || !listed {
		t.Fatalf("suite list exit = %d, read from FLOW_RECORDS_ADDR = %v; stderr:\n%s", code, listed, stderr.String())
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
