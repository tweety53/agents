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
	fix := strings.Join(strings.Fields(section(t, "skills/flow-contracts/pipeline.md", "Pipeline defects found mid-run")), " ")
	if want := "As `skills/flow-self-review/SKILL.md` states for its own loop, three failed review or verify rounds take **Three failed rounds** above."; !strings.Contains(fix, want) {
		t.Errorf("pipeline.md's Pipeline defects found mid-run lacks %q", want)
	}
	// Every loop's own file cites the rule.
	loops := []struct{ file, phrase string }{
		{"skills/flow/implement.md", "A fourth `full-suite-fix-<n>` after three failed ones is the fresh dispatch of **Three failed rounds**"},
		{"skills/flow/gated-review-fix.md", "after three failed re-reviews of one task is the fresh dispatch of **Three failed rounds**"},
		{"skills/flow/review-panel-fix-round.md", "as does a fourth round after three failed ones (**Three failed rounds**, there)"},
		{"skills/flow/review-panel.md", "the slot is re-dispatched fresh as **Three failed rounds**"},
		{"skills/flow/visual-verify.md", "as do three remainders of one dispatch that each close `NEXT: rest` (**Three failed rounds**, there)"},
		{"skills/flow/visual-verify.md", "Three re-dispatches that each close `NEXT: redispatch` take **Three failed rounds**"},
		{"skills/flow/unfinished-work-gate.md", "take **Three failed rounds** (`skills/flow-contracts/pipeline.md`), applied inline by the parent from every earlier verdict"},
		{"skills/flow/integrate.md", "Three stops in a row of one call take **Three failed rounds**"},
		{"skills/flow-self-review/SKILL.md", "Three failed rounds of the review loop or of the verify loop are answered the same way, by the next re-plan (**Three failed rounds**"},
		{"skills/flow/cleanup.md", "Three stops in a row of `cleanup` take **Three failed rounds**"},
		{"skills/flow/integrate.md", "only stops the parent can cure itself. A stop that exists to ask the operator — `STOP: main-checkout`, `STOP: base-mismatch` — is never counted toward it"},
		{"skills/flow/cleanup.md", "only stops the parent can cure itself. A stop that exists to ask or to hand to the operator — `STOP: worktree-cleanup`, `STOP: not-merged` — is never counted toward it"},
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

// TestCapsNeverStop pins that a cap bounds an approach, never the run: it is
// defined once, in pipeline.md's **Caps never stop the run**.
func TestCapsNeverStop(t *testing.T) {
	t.Parallel()
	pipeline := repoText(t, "skills/flow-contracts/pipeline.md")
	for _, want := range []string{
		"### Caps never stop the run",
		"**Every cap, ceiling, budget or count limit this pipeline names bounds an approach, never the run**",
		"the next attempt differs",
		"where no different approach exists the work goes past the bound with the fact recorded.",
		"It is never a stop and never an ask, and it counts toward **Three failed rounds** above.",
	} {
		if !strings.Contains(pipeline, want) {
			t.Errorf("pipeline.md lacks %q", want)
		}
	}
}

// TestCapsNeverStopListsNoAsk pins that **Caps never stop the run** lists no
// bound another contract asks about: document-fix.md asks Re-plan / Append
// anyway at the re-plan budget (self-review-kan-1110 F2).
func TestCapsNeverStopListsNoAsk(t *testing.T) {
	t.Parallel()
	text := repoText(t, "skills/flow-contracts/pipeline.md")
	_, caps, ok := strings.Cut(text, "### Caps never stop the run")
	if !ok {
		t.Fatal("pipeline.md has no ### Caps never stop the run")
	}
	if i := strings.Index(caps, " ## "); i >= 0 {
		caps = caps[:i]
	}
	if strings.Contains(caps, "re-plan budget") {
		t.Error("pipeline.md's Caps never stop the run lists the re-plan budget, which document-fix.md asks about")
	}
}

// section is the text of the "## <heading>" section of a Markdown file, up to
// the next level-2 heading.
func section(t *testing.T, rel, heading string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	i := strings.Index(s, "\n## "+heading)
	if i < 0 {
		t.Fatalf("%s: no %q section", rel, "## "+heading)
	}
	s = s[i+1:]
	if j := strings.Index(s[3:], "\n## "); j >= 0 {
		s = s[:j+3]
	}
	return s
}
