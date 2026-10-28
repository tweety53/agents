package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
)

type fakeRepricer struct {
	since           time.Time
	priced, skipped int
	err             error
}

func (f *fakeRepricer) RepriceSince(_ context.Context, since time.Time) (int, int, error) {
	f.since = since
	return f.priced, f.skipped, f.err
}

// TestPricingRepriceAnswersCounts pins POST /api/v1/pricing/reprice
// (decision reprice-command).
func TestPricingRepriceAnswersCounts(t *testing.T) {
	post := func(t *testing.T, f *fakeRepricer, body string) (int, []byte) {
		t.Helper()
		cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
		fs := newFakeStore()
		srv, err := api.New(cfg, fs, fs, fs, fs, fs, nil, api.WithRepricer(f))
		if err != nil {
			t.Fatalf("api.New: %v", err)
		}
		ts := httptest.NewServer(srv.Handler())
		t.Cleanup(ts.Close)
		resp, err := http.Post(ts.URL+"/api/v1/pricing/reprice", "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatalf("POST: %v", err)
		}
		defer resp.Body.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)
		return resp.StatusCode, buf.Bytes()
	}

	t.Run("counts", func(t *testing.T) {
		f := &fakeRepricer{priced: 3, skipped: 2}
		status, body := post(t, f, `{"since": "2026-09-01"}`)
		if status != http.StatusOK {
			t.Fatalf("status %d, body %s", status, body)
		}
		if want := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC); !f.since.Equal(want) {
			t.Errorf("since = %v, want %v", f.since, want)
		}
		var got struct{ Priced, Skipped int }
		if err := json.Unmarshal(body, &got); err != nil || got.Priced != 3 || got.Skipped != 2 {
			t.Errorf("body = %s, want {\"priced\":3,\"skipped\":2}", body)
		}
	})
	t.Run("malformed since", func(t *testing.T) {
		if status, body := post(t, &fakeRepricer{}, `{"since": "09/01/2026"}`); status != http.StatusBadRequest {
			t.Errorf("status %d, body %s, want 400", status, body)
		}
	})
	t.Run("store error", func(t *testing.T) {
		if status, body := post(t, &fakeRepricer{err: errors.New("boom")}, `{"since": "2026-09-01"}`); status != http.StatusInternalServerError {
			t.Errorf("status %d, body %s, want 500", status, body)
		}
	})
}
