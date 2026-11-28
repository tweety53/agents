package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReviewPanelFindingsExampleRecords: the worked findings-<round>.jsonl example in
// skills/flow/review-panel.md is accepted line for line by `flow record finding -from`.
func TestReviewPanelFindingsExampleRecords(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "skills/flow/review-panel.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, block, ok := strings.Cut(string(b), "\n```jsonl\n")
	if !ok {
		t.Fatal("skills/flow/review-panel.md: no ```jsonl example block")
	}
	block, _, _ = strings.Cut(block, "\n```")
	want := strings.Count(strings.TrimSpace(block), "\n") + 1

	repo := recordRepo(t, "kan-918")
	env := newTestEnv(t)
	posted := 0
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if strings.Contains(r.URL.Path, "worktree") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(`{"change":"kan-918","dispatches":[],"findings":[]}`))
			return
		}
		_, _ = io.ReadAll(r.Body)
		posted++
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	file := filepath.Join(t.TempDir(), "findings-0.jsonl")
	if err := os.WriteFile(file, []byte(block+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "finding", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-change", "kan-918", "-from", file},
		env.get, strings.NewReader(""), &stdout, &stderr)
	if code != 0 || posted != want {
		t.Fatalf("record finding -from the example: exit %d, %d of %d lines posted; stderr:\n%s", code, posted, want, stderr.String())
	}
}
