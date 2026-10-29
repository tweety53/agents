package embedded

import (
	"testing"

	"github.com/tweety53/agents/stats/internal/modelpolicy"
)

// TestEmbeddedTableParses pins that the copy make placed beside embed.go is
// the real Allowed pairs table: it parses and holds plan-designer.
func TestEmbeddedTableParses(t *testing.T) {
	rows, err := modelpolicy.Parse(Table)
	if err != nil {
		t.Fatalf("Parse(Table): %v", err)
	}
	if len(rows["plan-designer"]) == 0 {
		t.Fatalf("embedded table has no plan-designer row: %v", rows)
	}
}
