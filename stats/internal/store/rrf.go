package store

import "sort"

// rrfHit is one fused item: its key and its summed reciprocal-rank score.
type rrfHit struct {
	Key   int
	Score float64
}

// rrfFuse fuses ranked key lists (best first) by reciprocal rank fusion: each list adds
// 1/(60+rank), rank starting at 1, to every key it holds. A key in one list still scores.
// Best first; ties break by the lower key.
func rrfFuse(lists ...[]int) []rrfHit {
	score := map[int]float64{}
	for _, l := range lists {
		for i, k := range l {
			score[k] += 1 / float64(60+i+1)
		}
	}
	out := make([]rrfHit, 0, len(score))
	for k, s := range score {
		out = append(out, rrfHit{k, s})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Key < out[j].Key
	})
	return out
}
