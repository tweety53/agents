package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// kan-1024: a STARTED change goes IN_PROGRESS only once the store shows the
// handoff's own stages -- run-instructions completed, write-in-progress open.
func TestStateSetInProgressNeedsTheHandoffStages(t *testing.T) {
	cases := []struct {
		name         string
		stored       string
		instructions string
		writing      string
		want         int
	}{
		{"hand flip refused", "STARTED", `[]`, `[]`, 1},
		{"instructions missing refused", "STARTED", `[]`, `[{}]`, 1},
		{"instructions aborted refused", "STARTED", `[{"outcome":"aborted"}]`, `[{}]`, 1},
		{"write-in-progress not begun refused", "STARTED", `[{"outcome":"completed"}]`, `[{"outcome":"completed"}]`, 1},
		{"handoff stages pass", "STARTED", `[{"outcome":"completed"}]`, `[{"outcome":"completed"},{}]`, 0},
		{"fix run at IN_PROGRESS passes", "IN_PROGRESS", `[]`, `[]`, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := gitRepo(t)
			isolatedStateRoot(t)
			put := false
			srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/api/v1/stage-runs" && r.URL.Query().Get("stage") == "flow.run-instructions":
					_, _ = w.Write([]byte(`{"stageRuns":` + tc.instructions + `}`))
				case r.URL.Path == "/api/v1/stage-runs":
					_, _ = w.Write([]byte(`{"stageRuns":` + tc.writing + `}`))
				case r.Method == http.MethodPut:
					put = true
					_, _ = w.Write([]byte(`{"state":"IN_PROGRESS"}`))
				default:
					_, _ = w.Write([]byte(`{"state":"` + tc.stored + `"}`))
				}
			}))
			defer srv.Close()

			var stdout, stderr bytes.Buffer
			code := run(context.Background(),
				[]string{"state", "set", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "kan-16"},
				strings.NewReader(`{"state":"IN_PROGRESS"}`), &stdout, &stderr)

			if code != tc.want {
				t.Fatalf("exit = %d, want %d; stderr:\n%s", code, tc.want, stderr.String())
			}
			if put != (tc.want == 0) {
				t.Errorf("store written = %v, want %v", put, tc.want == 0)
			}
		})
	}
}
