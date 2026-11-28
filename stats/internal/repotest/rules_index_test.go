package repotest

// The always-on rule layer has two hand-kept indexes and one size budget:
// rules/agent-baseline.md's table must name every always-on rule, the
// flow-contracts SKILL.md index must name every contract file, and the core
// text every session's prompt carries must stay under a ceiling.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// coreBytesCeiling sits a little above the rendered core of the always-on rules;
// raise it only together with a reason the global prompt must grow.
const coreBytesCeiling = 15200

var alwaysApplyRe = regexp.MustCompile(`(?m)^alwaysApply:[ \t]*true[ \t]*$`)

// alwaysOnRules maps each always-on rules/*.mdc name (without extension) to its
// body after the frontmatter.
func alwaysOnRules(t *testing.T) map[string]string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(repoRoot(t), "rules", "*.mdc"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if !strings.HasPrefix(s, "---\n") {
			continue
		}
		end := strings.Index(s[4:], "\n---\n")
		if end < 0 || !alwaysApplyRe.MatchString(s[4:4+end]) {
			continue
		}
		out[strings.TrimSuffix(filepath.Base(p), ".mdc")] = s[4+end+5:]
	}
	if len(out) == 0 {
		t.Fatal("no always-on rule found under rules/")
	}
	return out
}

func TestAgentBaselineNamesEveryAlwaysOnRule(t *testing.T) {
	t.Parallel()
	baseline := repoText(t, "rules/agent-baseline.md")
	for name := range alwaysOnRules(t) {
		if want := "| `~/.claude/rules/" + name + ".md` |"; !strings.Contains(baseline, want) {
			t.Errorf("rules/agent-baseline.md's table has no row pointing at %s", want)
		}
	}
}

// TestRenderedCoreUnderCeiling mirrors setup.sh's render_managed_block: a rule's
// core is its body up to <!-- /core --> with the markers dropped; an unmarked
// rule renders whole.
func TestRenderedCoreUnderCeiling(t *testing.T) {
	t.Parallel()
	total := 0
	for _, body := range alwaysOnRules(t) {
		if i := strings.Index(body, "\n<!-- /core -->\n"); i >= 0 {
			body = body[:i+1]
		}
		total += len(strings.ReplaceAll(body, "<!-- core -->\n", ""))
	}
	if total > coreBytesCeiling {
		t.Errorf("always-on rule core is %d bytes, over the %d-byte ceiling", total, coreBytesCeiling)
	}
}

func TestFlowContractsIndexNamesEveryContract(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(repoRoot(t), "skills", "flow-contracts")
	index := repoText(t, "skills/flow-contracts/SKILL.md")
	paths, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		name := filepath.Base(p)
		if name == "SKILL.md" {
			continue
		}
		if want := "[" + name + "](" + name + ")"; !strings.Contains(index, want) {
			t.Errorf("skills/flow-contracts/SKILL.md's index has no row for %s", name)
		}
	}
}
