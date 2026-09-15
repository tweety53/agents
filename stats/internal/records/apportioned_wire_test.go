package records

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// TestApportionedWireKeySurvivesTheWriteReadRoundTrip is the drift guard
// review finding F5 asked for: the "apportioned" key has two authors on
// two sides of a wire that is deliberately not one struct --
// internal/harvest's MetricsPatch writes it, this package's dispatchMetrics
// reads it -- and the reason-strings pin test's own tradeoff (render.go's
// comment) covers prose, not json tags. A tag renamed on either side
// would otherwise break the wire contract with both suites green: the
// figure would simply stop qualifying. Marshalling the writer's patch and
// reading it back through the reader's shape, with both key names spelled
// in ONE literal, is the cheapest guard that cannot drift from either
// side.
//
// The import direction is the mirror of internal/harvest's own pin test
// (reason_strings_pin_test.go), which spells out why a test-only edge in
// the forbidden direction is the cheap alternative: only
// internal/records -> internal/harvest IN PRODUCTION CODE is forbidden;
// this file's test binary import is the same accepted exception, one
// level over.
func TestApportionedWireKeySurvivesTheWriteReadRoundTrip(t *testing.T) {
	patch, err := json.Marshal(harvest.MetricsPatch{
		Tokens: harvest.TokenDelta{
			Sidechain: harvest.Bucket{Input: 3},
		},
		Apportioned: &harvest.ApportionedDelta{Records: 3},
	})
	if err != nil {
		t.Fatalf("marshal MetricsPatch: %v", err)
	}
	if want := `"apportioned":{"records":3}`; !strings.Contains(string(patch), want) {
		t.Fatalf("writer's patch = %s, want it to carry %s verbatim -- the write-side key moved", patch, want)
	}

	var bag dispatchMetrics
	if err := json.Unmarshal(patch, &bag); err != nil {
		t.Fatalf("unmarshal into dispatchMetrics: %v", err)
	}
	if bag.Tokens == nil || bag.Tokens.Sidechain.Input != 3 {
		t.Fatalf("tokens sidechain input = %v, want 3 -- the read side lost the figures", bag.Tokens)
	}
	if bag.Apportioned == nil || bag.Apportioned.Records != 3 {
		t.Fatalf("apportioned = %+v, want records 3 -- the read side lost the concurrency-attribution key", bag.Apportioned)
	}
}
