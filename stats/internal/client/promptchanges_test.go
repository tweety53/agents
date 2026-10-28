package client_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
)

// TestRecordPromptChange pins RecordPromptChange: the row is POSTed as JSON
// to /api/v1/prompt-changes, a 204 is success and a 400 surfaces the
// daemon's message.
func TestRecordPromptChange(t *testing.T) {
	var gotPath string
	var got map[string]any
	status := http.StatusNoContent
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(status)
		if status == http.StatusBadRequest {
			_, _ = w.Write([]byte(`{"error":"unknown area \"planner\""}`))
		}
	}))
	defer srv.Close()
	c := client.New(srv.URL, srv.Client())

	slot := "primary"
	row := client.PromptChange{SHA: "abc:t", At: time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC),
		Kind: "edit", Area: "review", Slot: &slot, Target: "t", Summary: "s"}
	if err := c.RecordPromptChange(context.Background(), row); err != nil {
		t.Fatalf("RecordPromptChange: %v", err)
	}
	if gotPath != "POST /api/v1/prompt-changes" || got["sha"] != "abc:t" || got["area"] != "review" || got["slot"] != "primary" {
		t.Errorf("sent %s %v", gotPath, got)
	}

	status = http.StatusBadRequest
	err := c.RecordPromptChange(context.Background(), row)
	if err == nil || !strings.Contains(err.Error(), `unknown area "planner"`) {
		t.Errorf("400: err = %v, want the daemon's message", err)
	}
}
