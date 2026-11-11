package embedded

import (
	"slices"
	"testing"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
)

// TestEmbeddedTableParses pins that the copy make placed beside embed.go is
// the real roles table: it parses and holds plan-designer.
func TestEmbeddedTableParses(t *testing.T) {
	rows, err := modelpolicy.Rows(Table)
	if err != nil {
		t.Fatalf("Rows(Table): %v", err)
	}
	if !slices.ContainsFunc(rows, func(e modelpolicy.Entry) bool { return e.Role == "plan-designer" }) {
		t.Fatalf("embedded table has no plan-designer row: %v", rows)
	}
}
