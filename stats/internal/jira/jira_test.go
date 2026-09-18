package jira

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNormalizeStatusName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"In Progress", "in progress"},
		{"in-progress", "in progress"},
		{"IN  PROGRESS", "in progress"},
		{"  In Review ", "in review"},
		{"TO DO URGENT", "to do urgent"},
		{"Code\tReview", "code review"},
	}
	for _, tc := range cases {
		if got := normalizeStatusName(tc.in); got != tc.want {
			t.Errorf("normalizeStatusName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPositionForName(t *testing.T) {
	cases := []struct {
		in   string
		want Position
	}{
		{"To Do", PositionToDo},
		{"TO DO URGENT", PositionToDo},
		{"In Progress", PositionInProgress},
		{"In-Progress", PositionInProgress},
		{"In Review", PositionInReview},
		{"Code Review", PositionInReview},
		{"Done", PositionDone},
		// A statusCategory that groups a custom status with In Progress
		// must never matter: the name alone decides.
		{"TO DO URGENT INDETERMINATE", PositionUnknown},
		{"Blocked", PositionUnknown},
		{"", PositionUnknown},
	}
	for _, tc := range cases {
		if got := PositionForName(tc.in); got != tc.want {
			t.Errorf("PositionForName(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseTarget(t *testing.T) {
	if p, err := ParseTarget("in-review"); err != nil || p != PositionInReview {
		t.Errorf("ParseTarget(\"in-review\") = %v, %v, want InReview, nil", p, err)
	}
	if _, err := ParseTarget("Blocked"); !strings.Contains(err.Error(), "accepted") {
		t.Errorf("ParseTarget(\"Blocked\") error = %v, want one naming the accepted spellings", err)
	}
}

// recordingServer answers a happy To Do → In Progress path, counting the
// transitions it fires and checking each fired payload carries its
// transition id.
type recordingServer struct {
	t           *testing.T
	transitions int
}

func (r *recordingServer) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(req.URL.Path, "/transitions") && req.Method == http.MethodPost:
			r.transitions++
			var payload transitionRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				r.t.Errorf("bad transition payload: %v", err)
			}
			if payload.Transition.ID == "" {
				r.t.Error("transition payload carries no transition id")
			}
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(req.URL.Path, "/transitions"):
			fmt.Fprintf(w, `{"transitions":[{"id":"21","to":{"name":"In Progress"}},{"id":"31","to":{"name":"In Review"}}]}`)
		default:
			fmt.Fprintf(w, `{"fields":{"status":{"name":"To Do"}}}`)
		}
	}
}

func newTestClient(t *testing.T, h http.HandlerFunc) (*Client, *[]time.Duration) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := New(Config{Site: srv.URL, Email: "ops@example.com", APIToken: "token"}, srv.Client())
	var pauses []time.Duration
	c.pause = func(ctx context.Context, d time.Duration) error {
		pauses = append(pauses, d)
		return nil
	}
	return c, &pauses
}

func TestTransitionMovesForward(t *testing.T) {
	rec := &recordingServer{t: t}
	c, pauses := newTestClient(t, rec.handler())

	got, err := c.Transition(context.Background(), "KAN-1", PositionInProgress)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if !got.Moved || got.Status != "In Progress" {
		t.Errorf("Result = %+v, want moved to In Progress", got)
	}
	if len(*pauses) != 0 {
		t.Errorf("a clean transition paused %d times, want 0", len(*pauses))
	}
	if rec.transitions != 1 {
		t.Errorf("fired %d transitions, want 1", rec.transitions)
	}
}

func TestTransitionAlreadyThere(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			t.Error("an already-at-target no-op must not fire a transition")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"fields":{"status":{"name":"In Progress"}}}`)
	}
	c, _ := newTestClient(t, h)

	got, err := c.Transition(context.Background(), "KAN-1", PositionInProgress)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if got.Moved || got.Status != "In Progress" {
		t.Errorf("Result = %+v, want the unmoved In Progress no-op", got)
	}
}

// An issue past the target is the same successful no-op, never a backward
// move: a Done issue asked for In Progress stays Done.
func TestTransitionNeverMovesBackward(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			t.Error("a past-target no-op must not fire a transition")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"fields":{"status":{"name":"Done"}}}`)
	}
	c, _ := newTestClient(t, h)

	got, err := c.Transition(context.Background(), "KAN-1", PositionInProgress)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if got.Moved || got.Status != "Done" {
		t.Errorf("Result = %+v, want the unmoved Done no-op", got)
	}
}

func TestTransitionUnrecognizedStatus(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"fields":{"status":{"name":"Blocked"}}}`)
	}
	c, _ := newTestClient(t, h)

	_, err := c.Transition(context.Background(), "KAN-1", PositionDone)
	if !errors.Is(err, ErrUnrecognizedStatus) {
		t.Fatalf("Transition error = %v, want ErrUnrecognizedStatus", err)
	}
	if !strings.Contains(err.Error(), "Blocked") {
		t.Errorf("error %v does not name the unrecognized status", err)
	}
}

func TestTransitionNoTransitionOffered(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(req.URL.Path, "/transitions") {
			fmt.Fprintf(w, `{"transitions":[{"id":"11","to":{"name":"To Do"}}]}`)
			return
		}
		fmt.Fprintf(w, `{"fields":{"status":{"name":"To Do"}}}`)
	}
	c, _ := newTestClient(t, h)

	_, err := c.Transition(context.Background(), "KAN-1", PositionDone)
	if !errors.Is(err, ErrNoTransition) {
		t.Fatalf("Transition error = %v, want ErrNoTransition", err)
	}
}

func TestTransitionUnknownTargetRefused(t *testing.T) {
	c, _ := newTestClient(t, http.NotFoundHandler().ServeHTTP)
	if _, err := c.Transition(context.Background(), "KAN-1", PositionUnknown); err == nil {
		t.Fatal("Transition to PositionUnknown must be refused")
	}
}

// Two 500s on the status read, then a clean run: the whole sequence
// retries and the transition lands.
func TestTransitionRetriesTransient(t *testing.T) {
	calls := 0
	h := func(w http.ResponseWriter, req *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if strings.HasSuffix(req.URL.Path, "/transitions") {
			fmt.Fprintf(w, `{"transitions":[{"id":"21","to":{"name":"In Progress"}}]}`)
			return
		}
		fmt.Fprintf(w, `{"fields":{"status":{"name":"To Do"}}}`)
	}
	c, pauses := newTestClient(t, h)

	got, err := c.Transition(context.Background(), "KAN-1", PositionInProgress)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if !got.Moved {
		t.Errorf("Result = %+v, want moved", got)
	}
	if len(*pauses) != 2 {
		t.Errorf("paused %d times, want 2 (%v)", len(*pauses), *pauses)
	}
}

// A definitive answer never retries: one call, one typed error.
func TestTransitionDoesNotRetryDefinitive(t *testing.T) {
	calls := 0
	h := func(w http.ResponseWriter, req *http.Request) {
		calls++
		w.WriteHeader(http.StatusNotFound)
	}
	c, pauses := newTestClient(t, h)

	_, err := c.Transition(context.Background(), "KAN-1", PositionInProgress)
	if !errors.Is(err, ErrIssueNotFound) {
		t.Fatalf("Transition error = %v, want ErrIssueNotFound", err)
	}
	if calls != 1 {
		t.Errorf("served %d calls, want 1 with no retry", calls)
	}
	if len(*pauses) != 0 {
		t.Errorf("paused %d times on a definitive answer, want 0", len(*pauses))
	}
}

// The replay case: the POST lands server-side but its 5xx response is
// lost. The next attempt re-reads the status, sees the target reached and
// reports the no-op — not an error, not a second transition.
func TestTransitionReplaysAfterLostResponse(t *testing.T) {
	fired := 0
	h := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case req.Method == http.MethodPost:
			fired++
			w.WriteHeader(http.StatusBadGateway)
		case strings.HasSuffix(req.URL.Path, "/transitions"):
			fmt.Fprintf(w, `{"transitions":[{"id":"31","to":{"name":"In Review"}}]}`)
		default:
			if fired > 0 {
				fmt.Fprintf(w, `{"fields":{"status":{"name":"In Review"}}}`)
			} else {
				fmt.Fprintf(w, `{"fields":{"status":{"name":"To Do"}}}`)
			}
		}
	}
	c, _ := newTestClient(t, h)

	got, err := c.Transition(context.Background(), "KAN-1", PositionInReview)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if fired != 1 {
		t.Errorf("fired %d transitions, want exactly 1", fired)
	}
	if got.Moved || got.Status != "In Review" {
		t.Errorf("Result = %+v, want the unmoved In Review no-op on replay", got)
	}
}

func TestTransitionHonorsRetryAfter(t *testing.T) {
	calls := 0
	h := func(w http.ResponseWriter, req *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			w.Header().Set("Retry-After", strconv.Itoa(90))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		if strings.HasSuffix(req.URL.Path, "/transitions") {
			fmt.Fprintf(w, `{"transitions":[{"id":"21","to":{"name":"In Progress"}}]}`)
			return
		}
		fmt.Fprintf(w, `{"fields":{"status":{"name":"To Do"}}}`)
	}
	c, pauses := newTestClient(t, h)

	if _, err := c.Transition(context.Background(), "KAN-1", PositionInProgress); err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if len(*pauses) != 1 || (*pauses)[0] != maxRetryAfter {
		t.Errorf("pauses = %v, want one capped at %v", *pauses, maxRetryAfter)
	}
}

// The retry budget is bounded: every attempt failing transiently ends in
// ErrTransientExhausted, not an unbounded loop.
func TestTransitionTransientExhausted(t *testing.T) {
	calls := 0
	h := func(w http.ResponseWriter, req *http.Request) {
		calls++
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	c, _ := newTestClient(t, h)

	_, err := c.Transition(context.Background(), "KAN-1", PositionInProgress)
	if !errors.Is(err, ErrTransientExhausted) {
		t.Fatalf("Transition error = %v, want ErrTransientExhausted", err)
	}
	if calls != maxAttempts {
		t.Errorf("served %d calls, want exactly maxAttempts (%d)", calls, maxAttempts)
	}
}

func TestBackoffDelay(t *testing.T) {
	if d := backoffDelay(1, 2*time.Second); d != 2*time.Second {
		t.Errorf("backoffDelay with a Retry-After = %v, want the hint verbatim", d)
	}
	if d := backoffDelay(1, 90*time.Second); d != maxRetryAfter {
		t.Errorf("backoffDelay with an oversized hint = %v, want the %v cap", d, maxRetryAfter)
	}
	for attempt := 1; attempt < maxAttempts; attempt++ {
		d := backoffDelay(attempt, 0)
		base := baseDelay << (attempt - 1)
		if d < time.Duration(float64(base)*0.8) || d > time.Duration(float64(base)*1.2) {
			t.Errorf("backoffDelay(attempt %d) = %v, outside the ±20%% jitter window around %v", attempt, d, base)
		}
	}
}

// The budget is the ladder's hard stop: against an always-transient
// upstream, with real (uninjected) waits, the operation returns
// ErrTransientExhausted inside its own budget even though attempts remain
// -- the whole-operation bound the panel finding demanded, proven without
// waiting out the full ladder.
func TestTransitionStopsAtBudget(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()
	c := New(Config{Site: srv.URL, Email: "e@example.com", APIToken: "t"}, srv.Client())
	c.budget = 300 * time.Millisecond

	start := time.Now()
	_, err := c.Transition(context.Background(), "KAN-1", PositionInProgress)
	elapsed := time.Since(start)
	if !errors.Is(err, ErrTransientExhausted) {
		t.Fatalf("Transition error = %v, want ErrTransientExhausted", err)
	}
	if elapsed >= transitionBudget {
		t.Errorf("ladder ran %s, at or past its own %v budget", elapsed, transitionBudget)
	}
}
