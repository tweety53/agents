package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func jiraTransitionTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return ts
}

func TestClientJiraTransition(t *testing.T) {
	t.Run("200 is the result", func(t *testing.T) {
		ts := jiraTransitionTestServer(t, func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set(daemonHeaderName, daemonHeaderValue)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"key":"KAN-571","status":"In Review","moved":true}`))
		})
		got, err := New(ts.URL, ts.Client()).JiraTransition(context.Background(), "KAN-571", "In Review")
		if err != nil {
			t.Fatalf("JiraTransition: %v", err)
		}
		if got.Key != "KAN-571" || got.Status != "In Review" || !got.Moved {
			t.Errorf("got %+v, want the moved In Review result", got)
		}
	})

	t.Run("a 400 is the caller-mistake class", func(t *testing.T) {
		ts := jiraTransitionTestServer(t, func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set(daemonHeaderName, daemonHeaderValue)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"jira: \"Blocked\" matches no pipeline position (accepted: To Do, In Progress, In Review, Done)"}`))
		})
		_, err := New(ts.URL, ts.Client()).JiraTransition(context.Background(), "KAN-571", "Blocked")
		if !errors.Is(err, ErrJiraCallerMistake) {
			t.Fatalf("err = %v, want ErrJiraCallerMistake", err)
		}
		if errors.Is(err, ErrJiraRefused) {
			t.Errorf("a 400 must not read as ErrJiraRefused: the two exit classes differ")
		}
	})

	t.Run("a mapped refusal carries the daemon's message", func(t *testing.T) {
		ts := jiraTransitionTestServer(t, func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set(daemonHeaderName, daemonHeaderValue)
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error":"jira: workflow offers no transition to the requested position"}`))
		})
		_, err := New(ts.URL, ts.Client()).JiraTransition(context.Background(), "KAN-1", "Done")
		if !errors.Is(err, ErrJiraRefused) {
			t.Fatalf("err = %v, want ErrJiraRefused", err)
		}
		if !strings.Contains(err.Error(), "no transition") {
			t.Errorf("err %v does not carry the daemon's message", err)
		}
	})

	t.Run("a look-alike server is not trusted", func(t *testing.T) {
		ts := jiraTransitionTestServer(t, func(w http.ResponseWriter, req *http.Request) {
			// No Flow-Daemon header: not a store answer.
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"key":"KAN-571","status":"Done","moved":true}`))
		})
		_, err := New(ts.URL, ts.Client()).JiraTransition(context.Background(), "KAN-571", "Done")
		if !errors.Is(err, ErrUnavailable) {
			t.Fatalf("err = %v, want ErrUnavailable for a headerless answer", err)
		}
	})

	t.Run("an unmapped status is unavailable", func(t *testing.T) {
		ts := jiraTransitionTestServer(t, func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set(daemonHeaderName, daemonHeaderValue)
			w.WriteHeader(http.StatusTeapot)
		})
		_, err := New(ts.URL, ts.Client()).JiraTransition(context.Background(), "KAN-571", "Done")
		if !errors.Is(err, ErrUnavailable) {
			t.Fatalf("err = %v, want ErrUnavailable for an unmapped status", err)
		}
	})
}
