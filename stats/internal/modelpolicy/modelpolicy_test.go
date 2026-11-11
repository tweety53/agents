package modelpolicy

import (
	"os"
	"reflect"
	"testing"
)

const header = "| Role | Does |\n|---|---|\n"

// TestRowsTwoColumns pins the roles parse: one entry per row in table
// order, each under the nearest `###` heading, the scan ending at the next
// `## ` heading; every malformed table refused.
func TestRowsTwoColumns(t *testing.T) {
	got, err := Rows([]byte("## Allowed pairs\n\nintro\n\n### G\n\n" + header +
		"| `a` | Does a. |\n" +
		"| `b` | Does b. |\n\n### H\n\n" + header +
		"| `c` | Does c. |\n\n## Next\n\n### I\n\n" + header +
		"| `d` | Does d. |\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []Entry{{Role: "a", Group: "G", Does: "Does a."}, {Role: "b", Group: "G", Does: "Does b."}, {Role: "c", Group: "H", Does: "Does c."}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Rows = %#v, want %#v", got, want)
	}
	for name, table := range map[string]string{
		"missing table": "no table here\n",
		"two roles":     header + "| `a`, `b` | x |\n",
		"no does":       header + "| `a` | — |\n",
		"three cells":   header + "| `a` | x | y |\n",
		"role twice":    header + "| `a` | x |\n| `a` | y |\n",
	} {
		if _, err := Rows([]byte(table)); err == nil {
			t.Errorf("%s: parsed", name)
		}
	}

	b, err := os.ReadFile("../../../" + File)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := Rows(b)
	if err != nil {
		t.Fatalf("the real table: %v", err)
	}
	roles := map[string]bool{}
	var groups []string
	for _, e := range entries {
		roles[e.Role] = true
		if e.Group == "" {
			t.Errorf("role %q has no group", e.Role)
		}
		if len(groups) == 0 || groups[len(groups)-1] != e.Group {
			groups = append(groups, e.Group)
		}
	}
	for _, role := range []string{
		"explorer", "plan-designer", "planner", "implementer", "panel-fix", "panel-reviewer", "panel-sanity-efficiency", "panel-mutation-failure", "panel-rerun", "gated-reviewer",
		"gated-re-reviewer", "visual-verifier", "late-fix", "panel-merger", "tooling-analyst",
		"self-review-pass", "self-review-plan", "pipeline-fix-plan", "self-review-fix", "pipeline-fix-fix",
		"self-review-review-1", "pipeline-fix-review-1", "self-review-review-r", "pipeline-fix-review-r",
	} {
		if !roles[role] {
			t.Errorf("role %q has no row", role)
		}
	}
	if want := []string{"Planning", "Implementation", "Review", "Visual verification", "Self-review", "Pipeline fix"}; !reflect.DeepEqual(groups, want) {
		t.Errorf("groups = %q, want %q", groups, want)
	}
}

// TestParseRefusesPairsColumn: the four-column table the pairs used to live
// in is no roles table -- the pairs' one source is flowd's store.
func TestParseRefusesPairsColumn(t *testing.T) {
	old := "| Role | Allowed pairs | Requires marker | Does |\n|---|---|---|---|\n| `a` | `opus/flow-low` | — | Does a. |\n"
	if _, err := Rows([]byte(old)); err == nil {
		t.Fatal("Rows read a table carrying an Allowed pairs column")
	}
	if _, err := Rows([]byte(header + "| `a` | `opus/flow-low` | Does a. |\n")); err == nil {
		t.Fatal("Rows read a row carrying a pairs cell")
	}
}
