package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"hash/fnv"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// bowEmbedder is a deterministic bag-of-words embedder: each word hashes into one of 384 dims.
type bowEmbedder struct{ err error }

func (b bowEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	if b.err != nil {
		return nil, b.err
	}
	out := make([][]float32, len(texts))
	for i, t := range texts {
		v := make([]float32, 384)
		for _, w := range splitWords(t) {
			h := fnv.New32a()
			_, _ = h.Write([]byte(w))
			v[h.Sum32()%384]++
		}
		var n float64
		for _, x := range v {
			n += float64(x * x)
		}
		n = math.Sqrt(n)
		for j := range v {
			if n > 0 {
				v[j] /= float32(n)
			}
		}
		out[i] = v
	}
	return out, nil
}

func splitWords(s string) []string {
	var out []string
	cur := ""
	for _, r := range s + " " {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			cur += string(r)
		} else if cur != "" {
			out = append(out, cur)
			cur = ""
		}
	}
	return out
}

func TestHTTPEmbedder(t *testing.T) {
	t.Parallel()
	var batches []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Texts []string }
		_ = json.NewDecoder(r.Body).Decode(&in)
		batches = append(batches, len(in.Texts))
		vs := make([][]float32, len(in.Texts))
		for i := range vs {
			vs[i] = []float32{float32(i)}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"vectors": vs})
	}))
	defer srv.Close()
	texts := make([]string, 130)
	got, err := store.NewHTTPEmbedder(srv.URL).Embed(context.Background(), texts)
	if err != nil || len(got) != 130 || len(batches) != 3 || batches[0] != 64 || batches[1] != 64 || batches[2] != 2 {
		t.Fatalf("got %d vectors, batches %v, err %v; want 130 in 64+64+2", len(got), batches, err)
	}
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "boom", 500) }))
	defer bad.Close()
	if _, err := store.NewHTTPEmbedder(bad.URL).Embed(context.Background(), []string{"x"}); err == nil {
		t.Error("500: want an error")
	}
}

func recordTwo(t *testing.T, st *store.Store, projectKey string) {
	t.Helper()
	for _, f := range []records.SelfReviewFinding{
		{Change: "kan-1", Angle: "a", Note: "guard passes on a stale cache", Disposition: "noted"},
		{Change: "kan-2", Angle: "a", Note: "button colour", Disposition: "noted", Detail: "dashboard styling"},
	} {
		if _, err := st.RecordSelfReviewFinding(context.Background(), projectKey, f); err != nil {
			t.Fatal(err)
		}
	}
}

// Unconfigured: BM25 only, mode bm25.
func TestSimilarModeUnconfigured(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	recordTwo(t, st, projectKey)
	got, mode, herr, err := st.SimilarSelfReviewFindingsMode(context.Background(), projectKey, "stale cache", 5)
	if err != nil || herr != nil || mode != store.SimilarModeBM25 || len(got) != 1 {
		t.Fatalf("got %v mode %q herr %v err %v", got, mode, herr, err)
	}
}

// A failing embedder or an unreachable pgvector falls back to BM25 and says why.
func TestSimilarHybridFallback(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	recordTwo(t, st, projectKey)
	for name, e := range map[string]store.Embedder{"embedder down": bowEmbedder{err: errors.New("down")}, "pgvector down": bowEmbedder{}} {
		st.ConfigureHybrid(e, "postgres://flow:flow@127.0.0.1:1/none?sslmode=disable&connect_timeout=1")
		got, mode, herr, err := st.SimilarSelfReviewFindingsMode(context.Background(), projectKey, "stale cache", 5)
		if err != nil || herr == nil || mode != store.SimilarModeFallback || len(got) != 1 {
			t.Errorf("%s: got %v mode %q herr %v err %v", name, got, mode, herr, err)
		}
	}
}

// Needs a real pgvector database: FLOW_TEST_PGVECTOR_DSN.
func TestSimilarHybridPgvector(t *testing.T) {
	dsn := os.Getenv("FLOW_TEST_PGVECTOR_DSN")
	if dsn == "" {
		t.Skip("FLOW_TEST_PGVECTOR_DSN not set: the real pgvector side is not exercised")
	}
	st, _, projectKey := newHazardStore(t)
	recordTwo(t, st, projectKey)
	st.ConfigureHybrid(bowEmbedder{}, dsn)
	// Run twice: the first call backfills, the second finds every hash current.
	for i := 0; i < 2; i++ {
		got, mode, herr, err := st.SimilarSelfReviewFindingsMode(context.Background(), projectKey, "stale cache guard", 2)
		if err != nil || herr != nil || mode != store.SimilarModeHybrid || len(got) != 2 || got[0].Change != "kan-1" || got[0].Score <= got[1].Score {
			t.Fatalf("run %d: got %v mode %q herr %v err %v", i, got, mode, herr, err)
		}
	}
}
