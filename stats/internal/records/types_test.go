package records_test

import (
	"bytes"
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
		{
			name:    "an empty metrics bag -- never measured, never stamped -- counted as not measured",
			bag:     `{}`,
			wantN:   1,
			reasons: map[string]int{"not measured": 1},
		},
		{
			name:    "no metrics bag at all -- a hand-built Dispatch, not a store row -- not counted",
			bag:     ``,
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

// TestDispatchEndTokenReportMarshalsAllFourKeys pins the wire shape a
// caller-reported token report takes: all four keys ride the body even
// when a reported figure is zero -- an explicit zero is a real figure, the
// same rule render.go's tokenLine applies on the way out -- and an end
// carrying no report writes no `tokens` key at all.
func TestDispatchEndTokenReportMarshalsAllFourKeys(t *testing.T) {
	body, err := json.Marshal(records.DispatchEnd{
		SessionToken: "ff-kan525",
		Key:          "inline-implementer",
		Tokens:       &records.TokenReport{},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatalf("decode: %v", err)
	}
	tokens, ok := sent["tokens"].(map[string]any)
	if !ok {
		t.Fatalf("tokens = %v, want an object", sent["tokens"])
	}
	for _, k := range []string{"input", "output", "cache_read", "cache_creation"} {
		if v, ok := tokens[k]; !ok {
			t.Errorf("tokens key %q absent -- a reported zero is a reported figure, never an omitted one", k)
		} else if v != float64(0) {
			t.Errorf("tokens[%q] = %v, want 0", k, v)
		}
	}

	plain, err := json.Marshal(records.DispatchEnd{SessionToken: "ff-kan525", Key: "inline-implementer"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if bytes.Contains(plain, []byte(`"tokens"`)) {
		t.Errorf("a report-less end marshalled a tokens key: %s", plain)
	}
}

// TestRunHasRowsSeesEveryRowKind pins the row set: a run holding only
// passes, or only mutations, is a run with rows — a panel-bearing record
// the self-review bundle must render rather than report skipped.
func TestRunHasRowsSeesEveryRowKind(t *testing.T) {
	if (records.Run{}).HasRows() {
		t.Error("an empty run has no rows")
	}
	for name, run := range map[string]records.Run{
		"dispatches": {Dispatches: []records.Dispatch{{Seq: 1}}},
		"findings":   {Findings: []records.Finding{{Ref: "F1"}}},
		"passes":     {Passes: []records.Pass{{Note: "roster: full"}}},
		"mutations":  {Mutations: []records.Mutation{{Path: "x.go"}}},
	} {
		if !run.HasRows() {
			t.Errorf("a %s-only run has rows; HasRows said otherwise", name)
		}
	}
}
