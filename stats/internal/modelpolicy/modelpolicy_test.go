package modelpolicy

import (
	"os"
	"reflect"
	"testing"
)

const header = "| Role | Allowed pairs | Requires marker | Does |\n|---|---|---|---|\n"

// TestParse pins the table parse: rows keyed by role in table order, a marked
// row's marker, the table ending at its first non-row line; and every
// malformed table refused -- never a row loaded fail-open.
func TestParse(t *testing.T) {
	got, err := Parse([]byte("intro\n\n" + header +
		"| `a` | `opus/flow-low`, `sonnet/flow-low` | — | Does a. |\n" +
		"| `a` | `opus/flow-high` | `No-progress escalation` | — |\n" +
		"| `b` | `opus/Plan@low` | — | Does b. |\n" +
		"\n| `c` | `opus/flow-low` | — | Does c. |\n"))
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
		"| `a` | `opus/flow-low` | — | Does a. |\n" +
		"| `b` | `opus/Plan@low` | — | Does b. |\n" +
		"| `a` | `opus/flow-high` | `No-progress escalation` | — |\n"))
	if err != nil {
		t.Fatal(err)
	}
	wantEntries := []Entry{
		{Role: "a", Does: "Does a.", Row: Row{Pairs: []string{"opus/flow-low"}}},
		{Role: "b", Does: "Does b.", Row: Row{Pairs: []string{"opus/Plan@low"}}},
		{Role: "a", Does: "", Row: Row{Pairs: []string{"opus/flow-high"}, Marker: "No-progress escalation"}},
	}
	if !reflect.DeepEqual(entries, wantEntries) {
		t.Fatalf("Rows = %#v, want %#v -- table order, interleaved roles kept apart", entries, wantEntries)
	}

	for name, table := range map[string]string{
		"missing table": "no table here\n",
		"two roles":     header + "| `a`, `b` | `opus/flow-low` | — | x |\n",
		"no pair":       header + "| `a` | — | — | x |\n",
		"two markers":   header + "| `a` | `opus/flow-high` | `X`, `Y` | — |\n",
		"two cells":     header + "| `a` | `opus/flow-low` |\n",
		"three cells":   header + "| `a` | `opus/flow-low` | — |\n",
		"no does":       header + "| `a` | `opus/flow-low` | — | — |\n",
		"two groups":    "### G\n\n" + header + "| `a` | `opus/flow-low` | — | x |\n\n### H\n\n" + header + "| `a` | `opus/flow-medium` | — | x |\n",
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
		"explorer", "plan-designer", "planner", "implementer", "panel-fix", "panel-reviewer", "panel-sanity-efficiency", "panel-mutation-failure", "panel-rerun", "gated-reviewer",
		"gated-re-reviewer", "visual-verifier", "late-fix", "panel-merger", "tooling-analyst",
		"self-review-pass", "self-review-plan", "pipeline-fix-plan", "self-review-fix", "pipeline-fix-fix",
		"self-review-review-1", "pipeline-fix-review-1", "self-review-review-r", "pipeline-fix-review-r",
	} {
		if len(rows[role]) == 0 {
			t.Errorf("role %q has no row", role)
		}
	}
	if len(rows["panel-extra"]) != 0 {
		t.Errorf("panel-extra still has a row")
	}
	entries, err = Rows(b)
	if err != nil {
		t.Fatal(err)
	}
	var groups []string
	for _, e := range entries {
		if e.Group == "" {
			t.Errorf("role %q has no group", e.Role)
		}
		if e.Marker == "" && e.Does == "" {
			t.Errorf("role %q's unmarked row has no Does", e.Role)
		}
		if len(groups) == 0 || groups[len(groups)-1] != e.Group {
			groups = append(groups, e.Group)
		}
	}
	if want := []string{"Planning", "Implementation", "Review", "Visual verification", "Self-review", "Pipeline fix"}; !reflect.DeepEqual(groups, want) {
		t.Errorf("groups = %q, want %q", groups, want)
	}
}

// TestParseGroups: every table under the section is read, each row's group
// the nearest `###` heading above it, and the scan ends at the next `## `
// heading -- a table past it is not read.
func TestParseGroups(t *testing.T) {
	got, err := Rows([]byte("## Allowed pairs\n\nintro\n\n### G\n\n" + header +
		"| `a` | `opus/flow-low` | — | Does a. |\n" +
		"| `a` | `opus/flow-high` | `M` | — |\n\n### H\n\n" + header +
		"| `b` | `opus/Plan@low` | — | Does b. |\n\n## Next\n\n### I\n\n" + header +
		"| `c` | `opus/flow-low` | — | Does c. |\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []Entry{
		{Role: "a", Group: "G", Does: "Does a.", Row: Row{Pairs: []string{"opus/flow-low"}}},
		{Role: "a", Group: "G", Row: Row{Pairs: []string{"opus/flow-high"}, Marker: "M"}},
		{Role: "b", Group: "H", Does: "Does b.", Row: Row{Pairs: []string{"opus/Plan@low"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Rows = %#v, want %#v", got, want)
	}
	got, err = Rows([]byte(header + "| `a` | `opus/flow-low` | — | Does a. |\n"))
	if err != nil || len(got) != 1 || got[0].Group != "" {
		t.Fatalf("Rows with no heading = %#v, %v, want one entry in group \"\"", got, err)
	}
}
