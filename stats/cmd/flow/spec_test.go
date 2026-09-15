package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSpecRecordPostsAndPrintsConfirmation pins the record verb: the spec
// name travels as the POST body, and the confirmation carries the store's
// stamped row -- the recorded timestamp and the changes-since figure.
func TestSpecRecordPostsAndPrintsConfirmation(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotBody, _ = readAll(r)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"spec":"baseline.spec.ts","lastRanAt":"2026-09-09T12:00:00Z","changesSince":2}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"spec", "record", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-spec", "baseline.spec.ts"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("spec record exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if gotMethod != http.MethodPost || !strings.HasPrefix(gotPath, "/api/v1/specs/") {
		t.Errorf("record request = %s %s, want POST under /api/v1/specs/", gotMethod, gotPath)
	}
	if !strings.Contains(string(gotBody), `"baseline.spec.ts"`) {
		t.Errorf("record body = %s, want the spec name", gotBody)
	}
	if !strings.Contains(stderr.String(), "recorded: spec baseline.spec.ts") ||
		!strings.Contains(stderr.String(), "2026-09-09T12:00:00Z") {
		t.Errorf("stderr = %q, want the recorded confirmation with the store's stamp", stderr.String())
	}
}

// TestSpecListPrintsRowsAndNeverRunSpecs pins the read verb: rows carry the
// spec, its last-run timestamp and the unrun-for-N stat, a spec with no
// recorded run rendering as `never` with its whole-history count. -dir
// inventories the directory's own spec files, carrying their basenames to
// the store as the named universe.
func TestSpecListPrintsRowsAndNeverRunSpecs(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotSpecs []string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotSpecs = r.URL.Query()["spec"]
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"spec":"baseline.spec.ts","lastRanAt":"2026-09-09T12:00:00Z","changesSince":3},
			{"spec":"controls.spec.ts","lastRanAt":null,"changesSince":7}
		]`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"spec", "list", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("spec list exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"baseline.spec.ts", "2026-09-09T12:00:00Z", "3 changes since",
		"controls.spec.ts", "never", "7 changes since",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}

	// -dir inventories the directory: the on-disk spec files become the
	// named universe travelling to the store.
	dir := t.TempDir()
	for _, name := range []string{"a.spec.ts", "b.spec.ts", "not-a-spec.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte{}, 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	gotSpecs = nil
	stdout.Reset()
	stderr.Reset()
	code = run(context.Background(),
		[]string{"spec", "list", "-dir", dir, "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("spec list -dir exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if len(gotSpecs) != 2 || gotSpecs[0] != "a.spec.ts" || gotSpecs[1] != "b.spec.ts" {
		t.Errorf("named universe = %v, want the directory's two spec basenames", gotSpecs)
	}
	if strings.Contains(stdout.String(), "not-a-spec") {
		t.Errorf("-dir swept in a non-spec file:\n%s", stdout.String())
	}
}

// TestSpecListEmptyDirPrintsNoSpecFiles pins the zero-match -dir case: a
// directory that exists but holds no *.spec.ts files answers its own empty
// inventory without consulting the store -- never the recorded-only
// listing an absent -dir produces, which is the silent degradation the
// -dir guard exists to prevent.
func TestSpecListEmptyDirPrintsNoSpecFiles(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("an empty -dir must not reach the store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"spec":"ghost.spec.ts","lastRanAt":"2026-09-01T00:00:00Z","changesSince":4}]`))
	}))
	defer srv.Close()

	empty := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"spec", "list", "-dir", empty, "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("spec list -dir empty exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "no *.spec.ts files in") {
		t.Errorf("stdout = %q, want the empty-inventory line", stdout.String())
	}
	if strings.Contains(stdout.String(), "ghost.spec.ts") {
		t.Errorf("stdout = %q, want no recorded rows for a named empty universe", stdout.String())
	}
}
