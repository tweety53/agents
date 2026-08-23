package records_test

import (
	"encoding/json"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
)

// TestCostStatusOf pins CostStatusOf's load-bearing precedence:
// tokenLine's own rule -- a real "tokens" figure outranks a stale
// "unattributed" stamp -- applied to a whole run rather than one rendered
// line. See CostStatusOf's own doc comment for why the two must agree.
func TestCostStatusOf(t *testing.T) {
	cases := []struct {
		name    string
		bag     string
		wantN   int
		reasons map[string]int
	}{
		{
			name:    "tokens only -- measured, not counted",
			bag:     `{"tokens":{"main":{"input":100,"output":20,"cache_read":0,"cache_creation":0},"sidechain":{"input":0,"output":0,"cache_read":0,"cache_creation":0}}}`,
			wantN:   0,
			reasons: map[string]int{},
		},
		{
			name:    "unattributed only -- counted, reason tallied",
			bag:     `{"unattributed":{"reason":"session never bound"}}`,
			wantN:   1,
			reasons: map[string]int{"session never bound": 1},
		},
		{
			name:    "both tokens and a stale unattributed stamp -- tokens wins, not counted",
			bag:     `{"tokens":{"main":{"input":100,"output":20,"cache_read":0,"cache_creation":0},"sidechain":{"input":0,"output":0,"cache_read":0,"cache_creation":0}},"unattributed":{"reason":"session never bound"}}`,
			wantN:   0,
			reasons: map[string]int{},
		},
		{
			name:    "an unreadable metrics bag -- not counted",
			bag:     `{oops`,
			wantN:   0,
			reasons: map[string]int{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := records.CostStatusOf(records.Run{
				Change: "demo",
				Dispatches: []records.Dispatch{{
					Seq: 1, Role: "implementer", Model: "opus",
					Metrics: json.RawMessage(tc.bag),
				}},
			})
			if got.Unattributed != tc.wantN {
				t.Errorf("Unattributed = %d, want %d", got.Unattributed, tc.wantN)
			}
			if len(got.Reasons) != len(tc.reasons) {
				t.Fatalf("Reasons = %v, want %v", got.Reasons, tc.reasons)
			}
			for k, v := range tc.reasons {
				if got.Reasons[k] != v {
					t.Errorf("Reasons[%q] = %d, want %d", k, got.Reasons[k], v)
				}
			}
		})
	}
}

// TestCostStatusOfTalliesTwoSimultaneousReasons pins that CostStatusOf
// tallies each dispatch under its own reason and both survive in the
// map -- two dispatches unattributed for two different reasons must not
// collapse into one bucket or overwrite each other.
func TestCostStatusOfTalliesTwoSimultaneousReasons(t *testing.T) {
	got := records.CostStatusOf(records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{
			{Seq: 1, Role: "implementer", Model: "opus", Metrics: json.RawMessage(`{"unattributed":{"reason":"session never bound"}}`)},
			{Seq: 2, Role: "reviewer", Model: "sonnet", Metrics: json.RawMessage(`{"unattributed":{"reason":"matched more than one dispatch","candidates":2}}`)},
		},
	})

	if got.Unattributed != 2 {
		t.Fatalf("Unattributed = %d, want 2", got.Unattributed)
	}
	want := map[string]int{"session never bound": 1, "matched more than one dispatch": 1}
	if len(got.Reasons) != len(want) {
		t.Fatalf("Reasons = %v, want %v", got.Reasons, want)
	}
	for k, v := range want {
		if got.Reasons[k] != v {
			t.Errorf("Reasons[%q] = %d, want %d", k, got.Reasons[k], v)
		}
	}
}
