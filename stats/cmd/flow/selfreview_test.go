package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// TestRunSelfReviewBundlePrintsBundle drives `flow self-review bundle`
// end to end: the project key resolves from -C, the request lands on the
// bundle route for that project and change, and the body is printed
// verbatim on stdout with exit 0.
func TestRunSelfReviewBundlePrintsBundle(t *testing.T) {
	t.Parallel()
	repo := linkedWorktree(t, gitRepo(t), "demo")
	env := newTestEnv(t)

	var gotPath, gotQuery string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Self-review context bundle for demo\n\nfound: 1 of 6 sources\n"))
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"self-review", "bundle", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
		env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if !strings.HasSuffix(gotPath, "/demo/bundle") || !strings.Contains(gotPath, "/api/v1/self-review/") {
		t.Errorf("request path = %s", gotPath)
	}
	// The resolved worktree must ride to the daemon as the repo
	// parameter, pinned to its VALUE, not just a prefix: the change
	// directory's sources are read from the worktree the caller's own
	// location resolves to, and a regression that sends the main checkout
	// — or nothing — would leave every one of them silently absent.
	resolved, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatal(err)
	}
	wantRepo := "repo=" + url.QueryEscape(resolved)
	if !strings.Contains(gotQuery, wantRepo) {
		t.Errorf("request query = %q, want %q carried", gotQuery, wantRepo)
	}
	if !strings.Contains(stdout.String(), "# Self-review context bundle for demo") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

// TestRunSelfReviewBundleWaitsPastTheStoreDeadline pins the bundle's own
// default deadline: the daemon assembles it from the worktree and it has no
// fallback, so a bare call waits past the store-write deadline
// (kan-1072 self-review, F3). The store deadline is shortened
// (shortDefaultTimeout), so the test is sequential.
func TestRunSelfReviewBundleWaitsPastTheStoreDeadline(t *testing.T) {
	shortDefaultTimeout(t)
	slowFor := defaultTimeout + 500*time.Millisecond
	repo := linkedWorktree(t, gitRepo(t), "demo")
	env := newTestEnv(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/bundle") {
			time.Sleep(slowFor)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Self-review context bundle for demo\n"))
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"self-review", "bundle", "-addr", srv.URL, "-C", repo, "-change", "demo"},
		env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Self-review context bundle for demo") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

// TestRunSelfReviewBundleRequiresChange pins the caller mistake: no
// -change is exit 2 and a stderr line, never a request.
func TestRunSelfReviewBundleRequiresChange(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"self-review", "bundle", "-addr", "http://127.0.0.1:1", "-C", repo},
		env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "-change is required") {
		t.Errorf("stderr = %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want nothing", stdout.String())
	}
}

// TestRunSelfReviewBundleUnreachableStoreExitsNonZero carries the read
// contract through the CLI: a store that cannot answer is reported to
// stderr and exits non-zero -- never a partial bundle on stdout a caller
// could mistake for the change's own.
func TestRunSelfReviewBundleUnreachableStoreExitsNonZero(t *testing.T) {
	t.Parallel()
	repo := linkedWorktree(t, gitRepo(t), "demo")
	env := newTestEnv(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"self-review", "bundle", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
		env.get, strings.NewReader(""), &stdout, &stderr)
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero (stderr: %s)", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want nothing on a failed read", stdout.String())
	}
	if !strings.Contains(stderr.String(), "self-review bundle") {
		t.Errorf("stderr = %q, want the failure named", stderr.String())
	}
}

// TestSelfReviewFindingCLI drives `flow self-review finding`: every flag
// lands in the posted JSON on the findings route; a missing required flag
// or a negative -blast-radius is exit 2 naming it, before any request; a
// store refusal (400) is exit 2; and an unreachable store is exit 0 with
// exactly one warning line -- the write never blocks.
func TestSelfReviewFindingCLI(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)

	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody = nil
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		if gotBody["disposition"] == "bogus" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"store: invalid self-review finding"}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(gotBody)
	}))
	defer srv.Close()

	finding := func(addr string, extra ...string) (int, string, string) {
		var stdout, stderr strings.Builder
		args := append([]string{"self-review", "finding", "-addr", addr, "-timeout", "500ms", "-C", repo}, extra...)
		code := run(context.Background(), args, env.get, strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
	full := []string{"-change", "demo", "-angle", "myflow-fix", "-disposition", "fixed", "-ref", "abc1234", "-blast-radius", "3", "-severity", "important", "-note", "guard gap"}

	if code, _, stderr := finding(srv.URL, full...); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr)
	}
	if !strings.HasPrefix(gotPath, "/api/v1/self-review/") || !strings.HasSuffix(gotPath, "/demo/findings") {
		t.Errorf("request path = %s", gotPath)
	}
	for k, want := range map[string]any{"angle": "myflow-fix", "disposition": "fixed", "ref": "abc1234", "blastRadius": float64(3), "severity": "important", "note": "guard gap"} {
		if gotBody[k] != want {
			t.Errorf("posted %s = %v, want %v", k, gotBody[k], want)
		}
	}

	gotPath = ""
	if code, _, _ := finding(srv.URL, "-change", "demo", "-angle", "myflow-fix", "-disposition", "declined", "-note", "n"); code != 0 {
		t.Fatalf("declined exit = %d", code)
	}
	if _, ok := gotBody["blastRadius"]; ok {
		t.Errorf("unset -blast-radius posted %v, want it absent", gotBody["blastRadius"])
	}

	if code, _, _ := finding(srv.URL, "-change", "demo", "-angle", "flow-improvement", "-disposition", "noted", "-note", "n"); code != 0 || gotBody["disposition"] != "noted" {
		t.Fatalf("noted: exit = %d, posted disposition %v", code, gotBody["disposition"])
	}

	for _, missing := range []string{"-change", "-angle", "-disposition", "-note"} {
		var args []string
		for i := 0; i < len(full); i += 2 {
			if full[i] != missing {
				args = append(args, full[i], full[i+1])
			}
		}
		gotPath = ""
		code, _, stderr := finding(srv.URL, args...)
		if code != 2 || !strings.Contains(stderr, missing+" is required") {
			t.Errorf("without %s: exit %d, stderr %q; want 2 naming it", missing, code, stderr)
		}
		if gotPath != "" {
			t.Errorf("without %s: a request went out", missing)
		}
	}

	if code, _, stderr := finding(srv.URL, "-change", "demo", "-angle", "a", "-disposition", "declined", "-note", "n", "-blast-radius", "-1"); code != 2 || !strings.Contains(stderr, "-blast-radius") {
		t.Errorf("negative -blast-radius: exit %d, stderr %q; want 2 naming it", code, stderr)
	}

	if code, _, stderr := finding(srv.URL, "-change", "demo", "-angle", "a", "-disposition", "bogus", "-note", "n"); code != 2 || !strings.Contains(stderr, "invalid self-review finding") {
		t.Errorf("refused write: exit %d, stderr %q; want 2 with the store's message", code, stderr)
	}

	code, stdout, stderr := finding("http://127.0.0.1:1", full...)
	if code != 0 {
		t.Fatalf("unreachable store: exit %d, want 0 (stderr %q)", code, stderr)
	}
	if lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n"); len(lines) != 1 || !strings.HasPrefix(lines[0], "⚠ flow: self-review finding not recorded") {
		t.Errorf("unreachable store stderr = %q, want exactly one warning line", stderr)
	}
	if stdout != "" {
		t.Errorf("unreachable store stdout = %q, want nothing", stdout)
	}
}

// TestSelfReviewFindingRefusesUngradedFixedOrFiled: a fixed or filed row
// with no -severity is refused, exit 2, before any request goes out; a
// declined row still records ungraded.
func TestSelfReviewFindingRefusesUngradedFixedOrFiled(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)

	reached := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	for _, tc := range []struct {
		disposition, ref string
		wantCode         int
	}{
		{"fixed", "abc1234", 2},
		{"filed", "KAN-1", 2},
		{"suggested", "", 2},
		{"declined", "", 0},
	} {
		reached = false
		args := []string{"self-review", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "demo", "-angle", "a", "-disposition", tc.disposition, "-note", "n"}
		if tc.ref != "" {
			args = append(args, "-ref", tc.ref)
		}
		var stdout, stderr strings.Builder
		code := run(context.Background(), args, env.get, strings.NewReader(""), &stdout, &stderr)
		if code != tc.wantCode {
			t.Errorf("%s with no -severity: exit %d, want %d (stderr %q)", tc.disposition, code, tc.wantCode, stderr.String())
		}
		refused := tc.wantCode == 2
		if refused && !strings.Contains(stderr.String(), "-severity is required") {
			t.Errorf("%s with no -severity: stderr %q, want it to name -severity", tc.disposition, stderr.String())
		}
		if reached == refused {
			t.Errorf("%s with no -severity: server reached = %v, want %v", tc.disposition, reached, !refused)
		}
	}
}

// TestSelfReviewFindingsCLI drives `flow self-review findings`: the rows
// print as a JSON array on stdout, and an unreachable store is non-zero
// with nothing on stdout -- the read's contract, unlike the write's.
func TestSelfReviewFindingsCLI(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/demo/findings") {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"change":"demo","angle":"myflow-fix","note":"n","disposition":"declined","recordedAt":"2026-10-04T00:00:00Z"}]`))
	}))
	defer srv.Close()

	read := func(addr string) (int, string) {
		var stdout, stderr strings.Builder
		code := run(context.Background(),
			[]string{"self-review", "findings", "-addr", addr, "-timeout", "500ms", "-C", repo, "-change", "demo"},
			env.get, strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String()
	}

	code, stdout := read(srv.URL)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(stdout), &rows); err != nil || len(rows) != 1 || rows[0]["disposition"] != "declined" {
		t.Errorf("stdout = %q (err %v), want the one-row JSON array", stdout, err)
	}

	if code, stdout := read("http://127.0.0.1:1"); code == 0 || stdout != "" {
		t.Errorf("unreachable: exit %d, stdout %q; want non-zero and nothing", code, stdout)
	}
}

// TestSelfReviewFindingSendsItsDetail: a suggested finding's -detail reaches the store whole.
func TestSelfReviewFindingSendsItsDetail(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)
	var got records.SelfReviewFinding
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	detail := "## Why\nslow\n## How\ncache\n## DoD\nfaster"
	var stdout, stderr strings.Builder
	code := run(context.Background(), []string{"self-review", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
		"-change", "demo", "-angle", "flow-improvement", "-disposition", "suggested", "-severity", "important", "-note", "cache it", "-detail", detail},
		env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 0 || got.Detail != detail || got.Disposition != "suggested" {
		t.Errorf("exit %d (stderr %q), sent %+v; want 0 and the detail %q", code, stderr.String(), got, detail)
	}
}

// TestSelfReviewFindingGetCLI drives `flow self-review finding-get <id>`: the one finding prints as a
// JSON object with its detail, the id may come before or after the flags, an id the project does not
// have or an unreachable store is non-zero with nothing on stdout, and a missing or non-numeric id is 2.
func TestSelfReviewFindingGetCLI(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/7"):
			_, _ = w.Write([]byte(`{"id":7,"change":"demo","angle":"flow-fix","note":"n","disposition":"suggested","detail":"## Proposal\nx","recordedAt":"2026-10-04T00:00:00Z"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"no such self-review finding"}`))
		}
	}))
	defer srv.Close()

	get := func(args ...string) (int, string) {
		var stdout, stderr strings.Builder
		code := run(context.Background(), append([]string{"self-review", "finding-get"}, args...), env.get, strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String()
	}

	for _, args := range [][]string{
		{"-addr", srv.URL, "-timeout", "500ms", "-C", repo, "7"},
		{"7", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
	} {
		code, stdout := get(args...)
		var row map[string]any
		if code != 0 || json.Unmarshal([]byte(stdout), &row) != nil || row["id"] != float64(7) || row["detail"] != "## Proposal\nx" {
			t.Errorf("%v: exit %d, stdout %q; want the finding's JSON with its detail", args, code, stdout)
		}
	}
	if code, stdout := get("-addr", srv.URL, "-timeout", "500ms", "-C", repo, "8"); code != 1 || stdout != "" {
		t.Errorf("unknown id: exit %d, stdout %q; want 1 and nothing", code, stdout)
	}
	if code, stdout := get("-addr", "http://127.0.0.1:1", "-timeout", "500ms", "-C", repo, "7"); code == 0 || stdout != "" {
		t.Errorf("unreachable: exit %d, stdout %q; want non-zero and nothing", code, stdout)
	}
	for _, args := range [][]string{{"-C", repo}, {"-C", repo, "abc"}, {"-C", repo, "0"}, {"-C", repo, "1", "2"}} {
		if code, _ := get(args...); code != 2 {
			t.Errorf("%v: exit %d, want 2", args, code)
		}
	}
}

// TestSelfReviewSimilarCLI drives `flow self-review similar <text...>`: the words join into one query,
// the output is the short JSON array, an unreachable store is 1, and missing text or a bad -n is 2.
func TestSelfReviewSimilarCLI(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)

	var gotQuery string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q") + "|" + r.URL.Query().Get("n")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":7,"change":"demo","angle":"a","note":"n","disposition":"noted","detail":"long","theme":"t","severity":"minor","reviewStatus":"declined","score":1.5,"recordedAt":"2026-10-04T00:00:00Z"}]`))
	}))
	defer srv.Close()

	sim := func(args ...string) (int, string) {
		var stdout, stderr strings.Builder
		code := run(context.Background(), append([]string{"self-review", "similar"}, args...), env.get, strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String()
	}
	code, stdout := sim("stale", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "cache", "-n", "3")
	var rows []map[string]any
	if code != 0 || json.Unmarshal([]byte(stdout), &rows) != nil || len(rows) != 1 || rows[0]["id"] != float64(7) || rows[0]["score"] != 1.5 || rows[0]["theme"] != "t" || rows[0]["reviewStatus"] != "declined" {
		t.Errorf("exit %d, stdout %q; want the one hit", code, stdout)
	}
	if _, has := rows[0]["detail"]; has || gotQuery != "stale cache|3" {
		t.Errorf("detail present %v, query %q; want no detail and %q", has, gotQuery, "stale cache|3")
	}
	if code, stdout := sim("-addr", "http://127.0.0.1:1", "-timeout", "500ms", "-C", repo, "x"); code == 0 || stdout != "" {
		t.Errorf("unreachable: exit %d, stdout %q; want non-zero and nothing", code, stdout)
	}
	for _, args := range [][]string{{"-C", repo}, {"-C", repo, "x", "-n", "0"}, {"-C", repo, "x", "-n", "51"}} {
		if code, _ := sim(args...); code != 2 {
			t.Errorf("%v: exit %d, want 2", args, code)
		}
	}
}

// TestSelfReviewSimilarFallbackWarning: an X-Similar-Mode of bm25-fallback prints the one warning line on
// stderr and still the JSON array on stdout; any other mode prints no warning.
func TestSelfReviewSimilarFallbackWarning(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)
	mode := "bm25-fallback"
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Similar-Mode", mode)
		w.Header().Set("X-Similar-Error", "embed: connection refused")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	sim := func() (int, string, string) {
		var stdout, stderr strings.Builder
		code := run(context.Background(), []string{"self-review", "similar", "x", "-addr", srv.URL, "-timeout", "500ms", "-C", repo}, env.get, strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
	code, stdout, stderr := sim()
	if code != 0 || strings.TrimSpace(stdout) != "[]" || !strings.Contains(stderr, "⚠ flow: hybrid similarity unavailable, BM25 only: embed: connection refused\n") {
		t.Errorf("fallback: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	mode = "hybrid"
	if _, _, stderr := sim(); strings.Contains(stderr, "hybrid similarity unavailable") {
		t.Errorf("hybrid: stderr %q, want no warning", stderr)
	}
}

// TestSelfReviewFindingImplementCLI drives `flow self-review finding-implement <id> -commit <sha>`:
// it PATCHes the finding's review status to approved with the landing sha as the reason (idempotent
// on repeat), an unknown id is exit 1 with the store's message, a bad sha or id is 2.
func TestSelfReviewFindingImplementCLI(t *testing.T) {
	t.Parallel()
	repo := gitRepo(t)
	env := newTestEnv(t)
	const sha = "1444d5f772a17d12c479ff47f2d6523514e345d8"

	var bodies []string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch && strings.HasSuffix(r.URL.Path, "/findings/7/review") {
			b, _ := io.ReadAll(r.Body)
			bodies = append(bodies, string(b))
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"no such suggested self-review finding"}`))
	}))
	defer srv.Close()

	impl := func(args ...string) (int, string) {
		var stdout, stderr strings.Builder
		code := run(context.Background(), append([]string{"self-review", "finding-implement", "-addr", srv.URL, "-timeout", "500ms", "-C", repo}, args...), env.get, strings.NewReader(""), &stdout, &stderr)
		return code, stderr.String()
	}
	for i := 0; i < 2; i++ {
		if code, stderr := impl("7", "-commit", sha); code != 0 {
			t.Fatalf("run %d: exit %d, stderr %q; want 0", i, code, stderr)
		}
	}
	var got map[string]string
	if len(bodies) != 2 || json.Unmarshal([]byte(bodies[0]), &got) != nil || got["status"] != "approved" || got["reason"] != "implemented in "+sha {
		t.Errorf("bodies = %q, want two approving with reason %q", bodies, "implemented in "+sha)
	}
	if code, stderr := impl("8", "-commit", sha); code != 1 || !strings.Contains(stderr, "no such suggested") {
		t.Errorf("unknown id: exit %d, stderr %q; want 1 naming the store's message", code, stderr)
	}
	for _, args := range [][]string{{"7"}, {"7", "-commit", "xyz"}, {"abc", "-commit", sha}, {"-commit", sha}} {
		if code, _ := impl(args...); code != 2 {
			t.Errorf("%v: exit %d, want 2", args, code)
		}
	}
}
