package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/client"
	"github.com/tweety53/agents/stats/internal/records"
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

// TestGetSelfReviewBundleAcceptsExactCapBody pins the cap boundary's other
// half: a body of exactly maxResponseBytes was served whole, and refusing
// it would call a complete answer truncated. The reader takes one byte
// past the cap precisely so this case stays a success.
func TestGetSelfReviewBundleAcceptsExactCapBody(t *testing.T) {
	head := []byte("# Self-review context bundle for demo\n")
	body := bytes.Repeat([]byte("x"), 1<<20-len(head))
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(append(head, body...))
	}))
	defer srv.Close()

	c := client.New(srv.URL, srv.Client())
	got, err := c.GetSelfReviewBundle(context.Background(), "proj", "demo", "")
	if err != nil {
		t.Fatalf("GetSelfReviewBundle: %v", err)
	}
	if len(got) != 1<<20 {
		t.Errorf("body = %d bytes, want the whole %d", len(got), 1<<20)
	}
}

// TestClientSelfReviewFindings pins the findings pair's wire shape against
// the handler's: the write POSTs the row as JSON to the findings route and
// decodes the 201 echo, a 400 is ErrRecordRejected (a caller mistake, not
// a store miss), and the read GETs the same route and decodes the array.
func TestClientSelfReviewFindings(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody records.SelfReviewFinding
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Errorf("decode request body: %v", err)
			}
			if gotBody.Disposition == "bogus" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"invalid self-review finding"}`))
				return
			}
			gotBody.ID = 7
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(gotBody)
		default:
			_, _ = w.Write([]byte(`[{"id":7,"change":"demo","angle":"myflow-fix","note":"n","disposition":"declined","recordedAt":"2026-10-04T00:00:00Z"}]`))
		}
	}))
	defer srv.Close()
	c := client.New(srv.URL, srv.Client())
	ctx := context.Background()

	two := 2
	out, err := c.RecordSelfReviewFinding(ctx, "proj", "demo", records.SelfReviewFinding{
		Angle: "myflow-fix", Note: "n", Disposition: "fixed", Ref: "abc1234", BlastRadius: &two,
	})
	if err != nil {
		t.Fatalf("RecordSelfReviewFinding: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/v1/self-review/proj/demo/findings" {
		t.Errorf("request = %s %s", gotMethod, gotPath)
	}
	if gotBody.Change != "demo" || gotBody.Ref != "abc1234" || gotBody.BlastRadius == nil || *gotBody.BlastRadius != 2 {
		t.Errorf("posted body = %+v, want change demo, ref abc1234, blast radius 2", gotBody)
	}
	if out.ID != 7 {
		t.Errorf("returned ID = %d, want the daemon's 7", out.ID)
	}

	if _, err := c.RecordSelfReviewFinding(ctx, "proj", "demo", records.SelfReviewFinding{Disposition: "bogus"}); !errors.Is(err, client.ErrRecordRejected) {
		t.Errorf("bogus write err = %v, want ErrRecordRejected", err)
	}

	list, err := c.ListSelfReviewFindings(ctx, "proj", "demo")
	if err != nil {
		t.Fatalf("ListSelfReviewFindings: %v", err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v1/self-review/proj/demo/findings" {
		t.Errorf("list request = %s %s", gotMethod, gotPath)
	}
	if len(list) != 1 || list[0].ID != 7 || list[0].Disposition != "declined" {
		t.Errorf("list = %+v", list)
	}
}
