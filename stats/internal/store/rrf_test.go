package store

import (
	"math"
	"testing"
)

func TestRRFFuse(t *testing.T) {
	t.Parallel()
	got := rrfFuse([]int{1, 2, 3}, []int{3, 4})
	// 3: 1/63 + 1/61; 1: 1/61; 4: 1/62; 2: 1/62 (ties by key: 2 before 4).
	wantKeys := []int{3, 1, 2, 4}
	if len(got) != 4 {
		t.Fatalf("got %v", got)
	}
	for i, k := range wantKeys {
		if got[i].Key != k {
			t.Fatalf("order %v, want keys %v", got, wantKeys)
		}
	}
	if math.Abs(got[0].Score-(1.0/63+1.0/61)) > 1e-12 {
		t.Errorf("score %v", got[0].Score)
	}
	if got := rrfFuse(); len(got) != 0 {
		t.Errorf("empty: %v", got)
	}
}
