package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestPricingRepricePrintsCounts pins `flow pricing reprice` (decision
// reprice-command): counts on success, usage on a bad -since, and the
// client's error on an unreachable store.
func TestPricingRepricePrintsCounts(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/pricing/reprice" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"priced":3,"skipped":2}`))
	}))
	defer srv.Close()
	slow := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2500 * time.Millisecond)
		_, _ = w.Write([]byte(`{"priced":1,"skipped":0}`))
	}))
	defer slow.Close()
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close()

	for _, tc := range []struct {
		name      string
		args      []string
		code      int
		stdout    string
		stderrHas string
	}{
		{"counts", []string{"-addr", srv.URL, "-since", "2026-09-01"}, 0, "priced 3, skipped 2\n", ""},
		{"missing since", []string{"-addr", srv.URL}, 2, "", "usage:"},
		{"malformed since", []string{"-addr", srv.URL, "-since", "09/01/2026"}, 2, "", "usage:"},
		{"stray argument", []string{"-addr", srv.URL, "-since", "2026-09-01", "extra"}, 2, "", "usage:"},
		// Reprice outlives the 2s default every other command uses.
		{"slow store", []string{"-addr", slow.URL, "-since", "2026-09-01"}, 0, "priced 1, skipped 0\n", ""},
		{"unreachable", []string{"-addr", dead.URL, "-since", "2026-09-01", "-timeout", "2s"}, 1, "", "flow: pricing reprice:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), append([]string{"pricing", "reprice"}, tc.args...),
				strings.NewReader(""), &stdout, &stderr)
			if code != tc.code || stdout.String() != tc.stdout || !strings.Contains(stderr.String(), tc.stderrHas) {
				t.Errorf("exit %d stdout %q stderr %q; want exit %d stdout %q stderr containing %q",
					code, stdout.String(), stderr.String(), tc.code, tc.stdout, tc.stderrHas)
			}
		})
	}
}
