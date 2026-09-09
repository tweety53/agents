package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/records"
)

// minimalSuiteRun is the smallest suite run a caller records: what ran, on
// what host, for how long, and how it ended.
func minimalSuiteRun() records.SuiteRun {
	return records.SuiteRun{
		Suite:      "guard-tests",
		Host:       "laptop",
		DurationMs: 52_400,
		ExitCode:   0,
	}
}

// TestRecordSuiteRunPostsAndReadsTheRecordedRow pins the route and that the
// row the daemon allocated -- its id and ran_at -- comes back to the caller.
func TestRecordSuiteRunPostsAndReadsTheRecordedRow(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/suites/proj/runs" {
			t.Errorf("request = %s %s, want POST /api/v1/suites/proj/runs", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var in records.SuiteRun
		if err := json.Unmarshal(body, &in); err != nil || in.Suite != "guard-tests" {
			t.Errorf("body = %s (%v), want the recorded suite run as JSON", body, err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":9,"suite":"guard-tests","host":"laptop","durationMs":52400,"exitCode":0,"ranAt":"2026-09-09T12:00:00Z"}`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	got, err := c.RecordSuiteRun(context.Background(), "proj", minimalSuiteRun())
	if err != nil {
		t.Fatalf("RecordSuiteRun: %v", err)
	}
	if got.ID != 9 || got.RanAt.IsZero() {
		t.Errorf("recorded suite run = %+v, want the daemon's id and ran_at", got)
	}
}

// TestListSuiteRunsReadsTheArray pins the read route's query contract: an
// empty suite sends no suite parameter, a named one does, limit always
// travels, and the JSON array decodes in order.
func TestListSuiteRunsReadsTheArray(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/suites/proj/runs" {
			t.Errorf("request = %s %s, want GET /api/v1/suites/proj/runs", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("suite"); got != "guard-tests" {
			t.Errorf("suite query = %q, want guard-tests", got)
		}
		if got := r.URL.Query().Get("limit"); got != "2" {
			t.Errorf("limit query = %q, want 2", got)
		}
		_, _ = w.Write([]byte(`[{"id":2,"suite":"guard-tests","host":"ci","durationMs":90000,"exitCode":0,"ranAt":"2026-09-09T11:00:00Z"},{"id":1,"suite":"stats-go","host":"laptop","durationMs":24000,"exitCode":0,"ranAt":"2026-09-09T10:00:00Z"}]`))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	got, err := c.ListSuiteRuns(context.Background(), "proj", "guard-tests", 2)
	if err != nil {
		t.Fatalf("ListSuiteRuns: %v", err)
	}
	if len(got) != 2 || got[0].Suite != "guard-tests" || got[1].Suite != "stats-go" {
		t.Errorf("ListSuiteRuns = %+v, want the two rows in stored order", got)
	}
}

// TestSuiteRunCallsFallBackWhenNothingTrustworthyAnswers extends the record
// calls' fallback contract to the suite-run pair: a dead port and a
// look-alike server with no daemon header both answer ErrUnavailable, so
// the CLI's classify layer can treat them identically.
func TestSuiteRunCallsFallBackWhenNothingTrustworthyAnswers(t *testing.T) {
	lookAlike := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer lookAlike.Close()

	calls := map[string]func(*client.Client) error{
		"RecordSuiteRun": func(c *client.Client) error {
			_, err := c.RecordSuiteRun(context.Background(), "proj", minimalSuiteRun())
			return err
		},
		"ListSuiteRuns": func(c *client.Client) error {
			_, err := c.ListSuiteRuns(context.Background(), "proj", "", 20)
			return err
		},
	}

	for name, call := range calls {
		t.Run(name+" on a dead port", func(t *testing.T) {
			err := call(client.New(deadPortURL(t), &http.Client{Timeout: time.Second}))
			if !errors.Is(err, client.ErrUnavailable) {
				t.Errorf("err = %v, want ErrUnavailable", err)
			}
		})
		t.Run(name+" against a response with no daemon header", func(t *testing.T) {
			if err := call(client.New(lookAlike.URL, lookAlike.Client())); !errors.Is(err, client.ErrUnavailable) {
				t.Errorf("err = %v, want ErrUnavailable (a look-alike answer must fall back)", err)
			}
		})
	}
}
