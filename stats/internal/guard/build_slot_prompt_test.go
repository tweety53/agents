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
		"-global-constraints", "none", "-repo", "gymie", "-fix-report", "/r/panel-fix-report-2.md", "-panel", "primary,principles,security"}, Env{})
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
		"PANEL SLOTS: primary, principles, security",
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
		t.Errorf("mutation without -copies: rc %d, want 2", r.rc)
	}
	if r := runGuard("build-slot-prompt", []string{skill, wt, "0", "primary", "-diff", "/d"}, Env{}); r.rc != 2 {
		t.Errorf("primary without -artifacts: rc %d, want 2", r.rc)
	}
}

// TestBuildSlotPromptUnread renders a role's same-round re-dispatch over the sections its report
// left unread: one SECTION line per section, and its brief and report named `…-unread`.
func TestBuildSlotPromptUnread(t *testing.T) {
	t.Parallel()
	skill, _ := filepath.Abs("../../../skills/flow")
	wt := t.TempDir()
	r := runGuard("build-slot-prompt", []string{skill, wt, "1", "principles", "-diff", "/d/final-review.diff",
		"-standards", "", "-global-constraints", "none", "-unread", "gymie,gymie-frontend", "-panel", "primary,principles"}, Env{})
	if r.rc != 0 {
		t.Fatalf("rc %d: %s", r.rc, r.out)
	}
	path := filepath.Join(wt, ".superpowers/sdd/slot-prompt-1-principles-unread.md")
	if strings.TrimSpace(r.stdout) != path {
		t.Errorf("stdout %q, want %q", r.stdout, path)
	}
	got, _ := os.ReadFile(path)
	s := string(got)
	for _, want := range []string{"SECTION: gymie\n", "SECTION: gymie-frontend\n",
		"REPORT FILE: " + wt + "/.superpowers/sdd/panel-report-1-principles-unread.md"} {
		if !strings.Contains(s, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
	if r := runGuard("build-slot-prompt", []string{skill, wt, "1", "principles", "-diff", "/d", "-standards", "",
		"-global-constraints", "none", "-unread", "a", "-repo", "b"}, Env{}); r.rc != 2 {
		t.Errorf("-unread with -repo: rc %d, want 2", r.rc)
	}
}

// TestBuildSlotPromptMutation renders the mutation slot's brief: the mutation-testing brief and the
// MUTATION ENTRY CONTEXT paragraph from the skill's own files, the throwaway copies it works in,
// and each named test listed by path beneath the paragraph.
func TestBuildSlotPromptMutation(t *testing.T) {
	t.Parallel()
	skill, _ := filepath.Abs("../../../skills/flow")
	wt := t.TempDir()
	test := filepath.Join(t.TempDir(), "foo_test.go")
	if err := os.WriteFile(test, []byte("func TestFoo(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runGuard("build-slot-prompt", []string{skill, wt, "1", "mutation", "-diff", "/d/final-review.diff",
		"-copies", "/w-mutation-1,/f-mutation-1", "-tests", test, "-panel", "primary,mutation"}, Env{})
	if r.rc != 0 {
		t.Fatalf("rc %d: %s", r.rc, r.out)
	}
	got, _ := os.ReadFile(filepath.Join(wt, ".superpowers/sdd/slot-prompt-1-mutation.md"))
	s := string(got)
	for _, want := range []string{
		"## PASS mutation\n",
		"**Diff file:** /d/final-review.diff",
		"MUTATION COPIES: /w-mutation-1, /f-mutation-1",
		"Wherever the panel dispatches Mutation, the dispatch prompt carries a mutation-testing",
		"Never run `git worktree add`",
		"> **MUTATION ENTRY CONTEXT:**",
		"### " + test + "\n\nRead by path.",
		"REPORT FILE: " + wt + "/.superpowers/sdd/panel-report-1-mutation.md",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
	if strings.Contains(s, "## The throwaway worktree") || strings.Contains(s, "### Principles") {
		t.Errorf("the mutation brief ran past its section")
	}
	if r := runGuard("build-slot-prompt", []string{skill, wt, "1", "mutation+primary", "-diff", "/d", "-copies", "/w"}, Env{}); r.rc != 2 {
		t.Errorf("mutation bundled: rc %d, want 2", r.rc)
	}
}

// TestBuildSlotPromptMutationTestsNotInlined: a test file is named by path, never inlined, so the
// prompt stays small however large the file is.
func TestBuildSlotPromptMutationTestsNotInlined(t *testing.T) {
	t.Parallel()
	skill, _ := filepath.Abs("../../../skills/flow")
	wt := t.TempDir()
	big := filepath.Join(t.TempDir(), "big_test.go")
	if err := os.WriteFile(big, []byte(strings.Repeat("x", 300_000)), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runGuard("build-slot-prompt", []string{skill, wt, "1", "mutation", "-diff", "/d/final-review.diff",
		"-copies", "/w", "-tests", big, "-panel", "primary,mutation"}, Env{})
	if r.rc != 0 {
		t.Fatalf("rc %d: %s", r.rc, r.out)
	}
	got, _ := os.ReadFile(filepath.Join(wt, ".superpowers/sdd/slot-prompt-1-mutation.md"))
	if len(got) >= 100000 || !strings.Contains(string(got), "### "+big+"\n\nRead by path.") {
		t.Errorf("prompt is %d bytes or lacks the by-path line", len(got))
	}
}
