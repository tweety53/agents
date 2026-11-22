package store

import "testing"

func TestBM25Rank(t *testing.T) {
	t.Parallel()
	docs := []string{
		"the guard passes falsely when the cache is stale",
		"button colour in the dashboard",
		"stale cache makes the guard pass falsely",
	}
	hits := bm25Rank("guard false pass stale cache falsely", docs, 5)
	if len(hits) != 2 || hits[0].Index == 1 || hits[1].Index == 1 {
		t.Fatalf("hits = %+v, want the two guard documents and not the unrelated one", hits)
	}
	if hits[0].Score < hits[1].Score {
		t.Fatalf("hits = %+v, want best first", hits)
	}
	if got := bm25Rank("guard", docs, 1); len(got) != 1 {
		t.Fatalf("n=1 gave %d hits", len(got))
	}
	if got := bm25Rank("zzz", docs, 5); len(got) != 0 {
		t.Fatalf("no overlap gave %+v", got)
	}
}
