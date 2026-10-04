package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunSelfReviewBundlePrintsBundle drives `flow self-review bundle`
// end to end: the project key resolves from -C, the request lands on the
// bundle route for that project and change, and the body is printed
// verbatim on stdout with exit 0.
func TestRunSelfReviewBundlePrintsBundle(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

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
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if !strings.HasSuffix(gotPath, "/demo/bundle") || !strings.Contains(gotPath, "/api/v1/self-review/") {
		t.Errorf("request path = %s", gotPath)
	}
	// The resolved main checkout must ride to the daemon as the repo
	// parameter, pinned to its VALUE, not just a prefix: the
	// archive-derived sources are read from the repository the caller's
	// own location resolves to, and a regression that sends the raw -C
	// directory — or nothing — would leave every archive source silently
	// absent.
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

// TestRunSelfReviewBundleRequiresChange pins the caller mistake: no
// -change is exit 2 and a stderr line, never a request.
func TestRunSelfReviewBundleRequiresChange(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"self-review", "bundle", "-addr", "http://127.0.0.1:1", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)
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
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var stdout, stderr strings.Builder
	code := run(context.Background(),
		[]string{"self-review", "bundle", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "demo"},
		strings.NewReader(""), &stdout, &stderr)
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
	repo := gitRepo(t)
	isolatedStateRoot(t)

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
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		return code, stdout.String(), stderr.String()
	}
	full := []string{"-change", "demo", "-angle", "myflow-fix", "-disposition", "fixed", "-ref", "abc1234", "-blast-radius", "3", "-note", "guard gap"}

	if code, _, stderr := finding(srv.URL, full...); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr)
	}
	if !strings.HasPrefix(gotPath, "/api/v1/self-review/") || !strings.HasSuffix(gotPath, "/demo/findings") {
		t.Errorf("request path = %s", gotPath)
	}
	for k, want := range map[string]any{"angle": "myflow-fix", "disposition": "fixed", "ref": "abc1234", "blastRadius": float64(3), "note": "guard gap"} {
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

// TestSelfReviewFindingsCLI drives `flow self-review findings`: the rows
// print as a JSON array on stdout, and an unreachable store is non-zero
// with nothing on stdout -- the read's contract, unlike the write's.
func TestSelfReviewFindingsCLI(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

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
			strings.NewReader(""), &stdout, &stderr)
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
