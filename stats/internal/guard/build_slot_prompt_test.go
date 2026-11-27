package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestBuildSlotPrompt renders a primary+principles bundle from the repository's own templates:
// every placeholder filled, bodies dedented, one REPORT FILE per role; a fix round carries the
// FIX-ROUND SCOPE blockquote; mutation and an unfilled placeholder are usage errors.
func TestBuildSlotPrompt(t *testing.T) {
	t.Parallel()
	skill, _ := filepath.Abs("../../../skills/flow")
	wt := t.TempDir()
	r := runGuard("build-slot-prompt", []string{skill, wt, "2", "primary+principles",
		"-diff", "/d/fix-round-2.diff", "-artifacts", "/c/proposal.md,/c/design.md", "-standards", "",
		"-global-constraints", "none", "-repo", "gymie", "-fix-report", "/r/panel-fix-report-2.md"}, Env{})
	if r.rc != 0 {
		t.Fatalf("rc %d: %s", r.rc, r.out)
	}
	path := filepath.Join(wt, ".superpowers/sdd/slot-prompt-2-primary+principles-gymie.md")
	if strings.TrimSpace(r.stdout) != path {
		t.Errorf("stdout %q, want %q", r.stdout, path)
	}
	got, _ := os.ReadFile(path)
	s := string(got)
	for _, want := range []string{
		"Read " + skill + "/slot-brief.md in full first",
		"SECTION: gymie",
		"> **FIX-ROUND SCOPE:**",
		"`/r/panel-fix-report-2.md`",
		"## PASS primary\n\nYou are a Senior Code Reviewer",
		"**Diff file:** /d/fix-round-2.diff",
		"**Change artifacts:** /c/proposal.md, /c/design.md",
		"**Principles:** " + skill + "/engineering-principles.md",
		"REPORT FILE: " + wt + "/.superpowers/sdd/panel-report-2-primary-gymie.md",
		"REPORT FILE: " + wt + "/.superpowers/sdd/panel-report-2-principles-gymie.md",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
	if m := regexp.MustCompile(`\[[A-Z_]+\]`).FindString(s); m != "" {
		t.Errorf("placeholder %s left unfilled", m)
	}
	if strings.Contains(s, "<fix report>") || strings.Contains(s, "```") {
		t.Errorf("unfilled fix report or a fence left in the prompt")
	}
	if r := runGuard("build-slot-prompt", []string{skill, wt, "0", "mutation", "-diff", "/d"}, Env{}); r.rc != 2 {
		t.Errorf("mutation: rc %d, want 2", r.rc)
	}
	if r := runGuard("build-slot-prompt", []string{skill, wt, "0", "primary", "-diff", "/d"}, Env{}); r.rc != 2 {
		t.Errorf("primary without -artifacts: rc %d, want 2", r.rc)
	}
}
