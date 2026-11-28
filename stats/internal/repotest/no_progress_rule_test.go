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
	fix := repoText(t, "skills/flow-contracts/pipeline-defects.md")
	if want := "three red lands take **Three failed rounds** (`skills/flow-contracts/pipeline.md`)."; !strings.Contains(fix, want) {
		t.Errorf("pipeline-defects.md lacks %q", want)
	}
	// Every loop's own file cites the rule.
	loops := []struct{ file, phrase string }{
		{"skills/flow/fix.md", "a round being the **One round** column of the site's row: the fourth round after three failed ones is the fresh dispatch of **Three failed rounds**"},
		{"skills/flow/implement.md", "its rounds escalate as **The fix loop** (`skills/flow/fix.md`) states"},
		{"skills/flow/review-panel-fix-round.md", "Its rounds escalate as **The fix loop** (`skills/flow/fix.md`) states"},
		{"skills/flow/visual-verify.md", "as do three remainders of one dispatch that each close `NEXT: rest` (**Three failed rounds**, there)"},
		{"skills/flow/visual-verify.md", "Three re-dispatches that each close `NEXT: redispatch` take **Three failed rounds**"},
		{"skills/flow/unfinished-work-gate.md", "take **Three failed rounds** (`skills/flow-contracts/pipeline.md`), applied inline by the parent from every earlier verdict"},
		{"skills/flow/integrate.md", "Three stops in a row of one call take **Three failed rounds**"},
		{"skills/flow-self-review/SKILL.md", "three red lands in a row take **Three failed rounds**"},
		{"skills/flow/cleanup.md", "Three stops in a row of `cleanup` take **Three failed rounds**"},
		{"skills/flow/integrate.md", "only stops the parent can cure itself. A stop that exists to ask the operator — `STOP: main-checkout`, `STOP: base-mismatch` — is never counted toward it"},
		{"skills/flow/cleanup.md", "only stops the parent can cure itself. A stop that exists to ask or to hand to the operator — `STOP: worktree-cleanup` — is never counted toward it"},
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

// TestRoundCeilingRule pins the round ceiling: defined once, in pipeline.md's
// **Fewest operator actions**, run by `flow-guard round-ceiling`, configured
// by `## fix round ceiling`, cited by fix.md's fix loop, and carried as the
// one deferred item a handoff may name.
func TestRoundCeilingRule(t *testing.T) {
	t.Parallel()
	rule := strings.Join(strings.Fields(section(t, "skills/flow-contracts/pipeline.md", "Fewest operator actions")), " ")
	for _, want := range []string{
		"### The round ceiling",
		"**The round ceiling bounds a loop's rounds on one item, never the run.**",
		"`flow-guard round-ceiling <worktree> <failed rounds>`",
		"`NEXT: defer`",
		"A deferral waives no gate",
	} {
		if !strings.Contains(rule, want) {
			t.Errorf("pipeline.md's Fewest operator actions lacks %q", want)
		}
	}
	if strings.Contains(rule, "No round count ends") {
		t.Error("pipeline.md still says no round count ends a loop")
	}
	for _, l := range []struct{ file, phrase string }{
		{"skills/flow/fix.md", "for **Three failed rounds** and for **The round ceiling**"},
		{"skills/flow-contracts/project-configuration.md", "| `## fix round ceiling` |"},
	} {
		if !strings.Contains(repoText(t, l.file), l.phrase) {
			t.Errorf("%s lacks %q", l.file, l.phrase)
		}
	}
	if strings.Contains(repoText(t, "skills/flow/fix.md"), "No round count ends the loop.") {
		t.Error("skills/flow/fix.md still says no round count ends the loop")
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

// TestNoProgressRuleCountsFixRounds pins that pipeline.md's no-progress rule
// counts three failed rounds of one site of fix.md's fix loop, a round as
// that site's row defines it.
func TestNoProgressRuleCountsFixRounds(t *testing.T) {
	t.Parallel()
	rule := strings.Join(strings.Fields(section(t, "skills/flow-contracts/pipeline.md", "Fewest operator actions")), " ")
	if want := "three failed rounds of one site of **The fix loop** (`skills/flow/fix.md`), a round as that site's row defines it"; !strings.Contains(rule, want) {
		t.Errorf("pipeline.md's Fewest operator actions lacks %q", want)
	}
}
