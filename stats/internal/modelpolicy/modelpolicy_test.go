package modelpolicy

import (
	"os"
	"reflect"
	"testing"
)

const header = "| Role | Allowed pairs | Requires marker |\n|---|---|---|\n"

// TestParse pins the table parse: rows keyed by role in table order, a marked
// row's marker, the table ending at its first non-row line; and every
// malformed table refused -- never a row loaded fail-open.
func TestParse(t *testing.T) {
	got, err := Parse([]byte("intro\n\n" + header +
		"| `a` | `opus/flow-low`, `sonnet/flow-low` | — |\n" +
		"| `a` | `opus/flow-high` | `No-progress escalation` |\n" +
		"| `b` | `opus/Plan@low` | — |\n" +
		"\n| `c` | `opus/flow-low` | — |\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]Row{
		"a": {{Pairs: []string{"opus/flow-low", "sonnet/flow-low"}}, {Pairs: []string{"opus/flow-high"}, Marker: "No-progress escalation"}},
		"b": {{Pairs: []string{"opus/Plan@low"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Parse = %#v, want %#v", got, want)
	}
	entries, err := Rows([]byte(header +
		"| `a` | `opus/flow-low` | — |\n" +
		"| `b` | `opus/Plan@low` | — |\n" +
		"| `a` | `opus/flow-high` | `No-progress escalation` |\n"))
	if err != nil {
		t.Fatal(err)
	}
	wantEntries := []Entry{
		{Role: "a", Row: Row{Pairs: []string{"opus/flow-low"}}},
		{Role: "b", Row: Row{Pairs: []string{"opus/Plan@low"}}},
		{Role: "a", Row: Row{Pairs: []string{"opus/flow-high"}, Marker: "No-progress escalation"}},
	}
	if !reflect.DeepEqual(entries, wantEntries) {
		t.Fatalf("Rows = %#v, want %#v -- table order, interleaved roles kept apart", entries, wantEntries)
	}

	for name, table := range map[string]string{
		"missing table": "no table here\n",
		"two roles":     header + "| `a`, `b` | `opus/flow-low` | — |\n",
		"no pair":       header + "| `a` | — | — |\n",
		"two markers":   header + "| `a` | `opus/flow-high` | `X`, `Y` |\n",
		"two cells":     header + "| `a` | `opus/flow-low` |\n",
	} {
		if _, err := Parse([]byte(table)); err == nil {
			t.Errorf("%s: parsed", name)
		}
	}

	b, err := os.ReadFile("../../../" + File)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := Parse(b)
	if err != nil {
		t.Fatalf("the real table: %v", err)
	}
	for _, role := range []string{
		"explorer", "plan-designer", "implementer", "panel-fix", "panel-reviewer", "panel-extra", "panel-rerun", "gated-reviewer",
		"gated-re-reviewer", "visual-verifier", "late-fix", "panel-merger", "tooling-analyst",
		"self-review-pass", "self-review-plan", "pipeline-fix-plan", "self-review-fix", "pipeline-fix-fix",
		"self-review-review-1", "pipeline-fix-review-1", "self-review-review-r", "pipeline-fix-review-r",
	} {
		if len(rows[role]) == 0 {
			t.Errorf("role %q has no row", role)
		}
	}
}
