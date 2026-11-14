package repotest

// TestNoProgressRule pins the three-failed-rounds rule (KAN-1110): it is
// defined once, in skills/flow-contracts/pipeline.md's **Fewest operator
// actions**, and every retry loop's own file cites it.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoText is the text of a repository file, whitespace runs collapsed to one
// space so a phrase matches across line wraps.
func repoText(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(strings.Fields(string(b)), " ")
}

func TestNoProgressRule(t *testing.T) {
	t.Parallel()
	rule := strings.Join(strings.Fields(section(t, "skills/flow-contracts/pipeline.md", "Fewest operator actions")), " ")
	for _, want := range []string{
		"### Three failed rounds",
		"**Three failed rounds of one loop are a round without progress too, whatever their defects.**",
		"It is never a stop and never an ask.",
		"the implement phase's full-suite fix",
		"a guard or integrate call re-run after a stop",
	} {
		if !strings.Contains(rule, want) {
			t.Errorf("pipeline.md's Fewest operator actions lacks %q", want)
		}
	}
	// Every loop's own file cites the rule.
	loops := []struct{ file, phrase string }{
		{"skills/flow/implement.md", "A fourth `full-suite-fix-<n>` after three failed ones is the fresh dispatch of **Three failed rounds**"},
		{"skills/flow/gated-review-fix.md", "after three failed re-reviews of one task is the fresh dispatch of **Three failed rounds**"},
		{"skills/flow/review-panel-fix-round.md", "as does a fourth round after three failed ones (**Three failed rounds**, there)"},
		{"skills/flow/review-panel.md", "the slot is re-dispatched fresh as **Three failed rounds**"},
		{"skills/flow/verify-fix-loop.md", "and for **Three failed rounds** of this loop — a round being one `In-run fix <k>` task"},
	}
	if strings.Contains(repoText(t, "skills/flow/review-panel.md"), "a third breach of the same slot is asked") {
		t.Error("skills/flow/review-panel.md still asks on a slot's third breach")
	}
	for _, l := range loops {
		if !strings.Contains(repoText(t, l.file), l.phrase) {
			t.Errorf("%s lacks %q", l.file, l.phrase)
		}
	}
}
