package store

import (
	"math"
	"sort"
	"strings"
)

// bm25Hit is one document that shares a term with the query: its index in the input and its score.
type bm25Hit struct {
	Index int
	Score float64
}

// bm25Tokens lowercases s and splits it into runs of [a-z0-9].
func bm25Tokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
}

// bm25Rank ranks docs against query by Okapi BM25 (k1 1.5, b 0.75) and returns the top n with a
// score above zero, best first; ties keep the input order.
func bm25Rank(query string, docs []string, n int) []bm25Hit {
	const k1, b = 1.5, 0.75
	tfs := make([]map[string]int, len(docs))
	lens := make([]int, len(docs))
	df := map[string]int{}
	total := 0
	for i, d := range docs {
		toks := bm25Tokens(d)
		lens[i] = len(toks)
		total += len(toks)
		tfs[i] = map[string]int{}
		for _, t := range toks {
			if tfs[i][t] == 0 {
				df[t]++
			}
			tfs[i][t]++
		}
	}
	if len(docs) == 0 || total == 0 {
		return nil
	}
	avg := float64(total) / float64(len(docs))
	var hits []bm25Hit
	for i := range docs {
		score := 0.0
		for _, q := range bm25Tokens(query) {
			tf := float64(tfs[i][q])
			if tf == 0 {
				continue
			}
			idf := math.Log(1 + (float64(len(docs))-float64(df[q])+0.5)/(float64(df[q])+0.5))
			score += idf * tf * (k1 + 1) / (tf + k1*(1-b+b*float64(lens[i])/avg))
		}
		if score > 0 {
			hits = append(hits, bm25Hit{i, score})
		}
	}
	sort.SliceStable(hits, func(a, c int) bool { return hits[a].Score > hits[c].Score })
	if len(hits) > n {
		hits = hits[:n]
	}
	return hits
}
