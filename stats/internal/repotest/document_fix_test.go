package repotest

// The document-fix tests pin KAN-1110's fix-run planning: a re-decide whose
// class did not move carries the row, and the plan's closing commands run as
// one call.

import (
	"strings"
	"testing"
)

func TestReDecideCarriesAnUnmovedClass(t *testing.T) {
	t.Parallel()
	text := repoText(t, "skills/flow/document-fix.md")
	for _, want := range []string{
		"When the `class:` line it prints equals the newest recorded row's `classMechanical`, the row is carried",
		"`flow-guard plan-dispatch-groups`",
		"`inputs` from the `inputs:` line",
		"`groups_mechanical` and `groups`",
		"`visual` raised to `required`",
		"the run follows the new row, never the first row's panel or pairs",
		"flow record decision -change <name> -session-token mf-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("skills/flow/document-fix.md lacks %q", want)
		}
	}
}

func TestDocumentFixClosesInOneCall(t *testing.T) {
	t.Parallel()
	text := repoText(t, "skills/flow/document-fix.md")
	for _, want := range []string{
		"The count, `flow-guard check-plan-shape`, the project's plan-provenance guard and the snapshot command run as one Bash call, joined with `&&` in that order",
		"flow tasks count -C <worktree> <name>",
		"flow-guard check-plan-shape <changeRoot>/tasks.md",
		"before any implementer is dispatched",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("skills/flow/document-fix.md lacks %q", want)
		}
	}
}
