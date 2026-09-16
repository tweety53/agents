package client_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
)

// TestGetSelfReviewBundleReturnsTheBodyVerbatim pins the transport rule
// the bundle read shares with record render: the body arrives already
// assembled and is returned exactly as the daemon wrote it -- markdown,
// not an envelope the CLI would have to decode into shape. The caller-
// resolved repository root rides as the repo query parameter.
func TestGetSelfReviewBundleReturnsTheBodyVerbatim(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Self-review context bundle for demo\n\nfound: 1 of 6 sources\n"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	body, err := c.GetSelfReviewBundle(context.Background(), "proj", "demo", "/tmp/repo")
	if err != nil {
		t.Fatalf("GetSelfReviewBundle: %v", err)
	}
	if gotPath != "/api/v1/self-review/proj/demo/bundle" {
		t.Errorf("request path = %s", gotPath)
	}
	if gotQuery != "repo=%2Ftmp%2Frepo" {
		t.Errorf("request query = %s, want the escaped repo parameter", gotQuery)
	}
	if want := "# Self-review context bundle for demo\n\nfound: 1 of 6 sources\n"; string(body) != want {
		t.Errorf("body = %q, want %q", body, want)
	}
}

// TestGetSelfReviewBundleReportsNotFoundDistinctly carries the read
// contract's one distinct error: a 404 means this daemon predates the
// endpoint, which a caller may want to tell apart from a store that could
// not be reached at all.
func TestGetSelfReviewBundleReportsNotFoundDistinctly(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.GetSelfReviewBundle(context.Background(), "proj", "demo", "")
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if errors.Is(err, client.ErrUnavailable) {
		t.Errorf("a legitimate 404 must not also read as ErrUnavailable")
	}
}

// TestGetSelfReviewBundleLookalikeServerIsUnavailable pins the daemon
// header: a look-alike server's answer is never trusted as a store read.
func TestGetSelfReviewBundleLookalikeServerIsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Self-review context bundle for demo\n"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.GetSelfReviewBundle(context.Background(), "proj", "demo", "")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

// TestGetSelfReviewBundleRefusesCapTruncatedBody pins the truncation
// guard: a body that fills the client's whole response cap was cut off
// mid-read, and a partial bundle is refused outright -- never handed to a
// reasoning pass as if it were the change's own.
func TestGetSelfReviewBundleRefusesCapTruncatedBody(t *testing.T) {
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bytes.Repeat([]byte("#"), 2<<20))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	_, err := c.GetSelfReviewBundle(context.Background(), "proj", "demo", "")
	if err == nil {
		t.Fatal("a cap-truncated bundle was returned, want a refusal")
	}
	if !errors.Is(err, client.ErrUnavailable) {
		t.Errorf("err = %v, want ErrUnavailable", err)
	}
	if !strings.Contains(err.Error(), "truncated") {
		t.Errorf("err = %v, want the truncation named", err)
	}
}

// TestGetSelfReviewBundleUnreachableStoreIsUnavailable carries the other
// half: a dead port is ErrUnavailable, the never-a-partial-bundle case.
func TestGetSelfReviewBundleUnreachableStoreIsUnavailable(t *testing.T) {
	c := client.New(deadPortURL(t), &http.Client{Timeout: time.Second})
	_, err := c.GetSelfReviewBundle(context.Background(), "proj", "demo", "")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
