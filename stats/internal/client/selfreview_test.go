package client_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
)

// TestGetSelfReviewBundleReturnsTheBodyVerbatim pins the transport rule
// the bundle read shares with record render: the body arrives already
// assembled and is returned exactly as the daemon wrote it -- markdown,
// not an envelope the CLI would have to decode into shape.
func TestGetSelfReviewBundleReturnsTheBodyVerbatim(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Self-review context bundle for demo\n\nfound: 1 of 6 sources\n"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	body, err := c.GetSelfReviewBundle(context.Background(), "proj", "demo")
	if err != nil {
		t.Fatalf("GetSelfReviewBundle: %v", err)
	}
	if gotPath != "/api/v1/self-review/proj/demo/bundle" {
		t.Errorf("request path = %s", gotPath)
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
	_, err := c.GetSelfReviewBundle(context.Background(), "proj", "demo")
	if !errors.Is(err, client.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if errors.Is(err, client.ErrUnavailable) {
		t.Errorf("a legitimate 404 must not also read as ErrUnavailable")
	}
}

// TestGetSelfReviewBundleUnreachableStoreIsUnavailable carries the other
// half: a dead port is ErrUnavailable, the never-a-partial-bundle case.
func TestGetSelfReviewBundleUnreachableStoreIsUnavailable(t *testing.T) {
	c := client.New(deadPortURL(t), &http.Client{Timeout: time.Second})
	_, err := c.GetSelfReviewBundle(context.Background(), "proj", "demo")
	if !errors.Is(err, client.ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
