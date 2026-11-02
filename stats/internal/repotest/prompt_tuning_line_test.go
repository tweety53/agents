package repotest

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestPromptTuningLineShape pins the Finished handoff's `**Prompt tuning:**`
// line to its operator-requested shape (pipeline-fix 35): the value is an
// `index.html` path or the literal NONE with no reason, and it and the
// Learning line are shown only in the Finished handoff, never an earlier one
// (kan-1038).
func TestPromptTuningLineShape(t *testing.T) {
	root := repoRoot(t)
	read := func(rel string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	cleanup := read("skills/flow/cleanup.md")
	if !strings.Contains(cleanup, "**Prompt tuning:** <each case's `index.html` absolute path in `<agents repo>`'s main checkout> | NONE") {
		t.Error("skills/flow/cleanup.md: Finished block lacks the '**Prompt tuning:** <path> | NONE' line")
	}
	for _, l := range strings.Split(cleanup, "\n") {
		if strings.Contains(l, "Prompt tuning:") && strings.Contains(l, "| NONE —") {
			t.Error("skills/flow/cleanup.md: Prompt tuning NONE carries a reason")
		}
	}
	earlier := regexp.MustCompile(`(?m)^\*\*(Learning|Prompt tuning):\*\* <`)
	for _, f := range []string{"skills/flow/integrate.md", "skills/flow/verify-and-handoff.md", "skills/flow-contracts/handoff-blocks.md"} {
		if earlier.MatchString(read(f)) {
			t.Errorf("%s: an earlier handoff block prints a Learning or Prompt tuning line", f)
		}
	}
	if !strings.Contains(cleanup, "**The `Learning:` and `Prompt tuning:` lines are shown only once the change is fully done**") {
		t.Error("skills/flow/cleanup.md: no only-at-the-end sentence for the Learning and Prompt tuning lines")
	}
	if !strings.Contains(read("skills/flow-fast/SKILL.md"), "`**Prompt tuning:**`") {
		t.Error("skills/flow-fast/SKILL.md: does not cite the '**Prompt tuning:**' line")
	}

	// The hyphenated spelling, built so this file never matches itself.
	hyphenated := "Prompt" + "-tuning:"
	self := filepath.Join(root, "stats/internal/repotest/prompt_tuning_line_test.go")
	for _, dir := range []string{"skills", "rules", "scripts", "stats"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			switch filepath.Ext(p) {
			case ".md", ".mdc", ".go":
			default:
				return nil
			}
			if p == self {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if strings.Contains(string(b), hyphenated) {
				t.Errorf("%s: a '%s' spelling remains", p, hyphenated)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
