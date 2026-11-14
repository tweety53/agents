package repotest

// TestFixRunReadsOnce pins KAN-1110: a fix run's load-context and
// isolate-workspace skip what its planning pass already read and checked.

import (
	"strings"
	"testing"
)

func TestFixRunReadsOnce(t *testing.T) {
	t.Parallel()
	load := strings.Join(strings.Fields(section(t, "skills/flow/implement.md", "1. Load context and validate the plan")), " ")
	for _, want := range []string{
		"On a fix run `flow-guard check-plan-shape` is not run here",
		"is not read again unless it was edited since",
	} {
		if !strings.Contains(load, want) {
			t.Errorf("skills/flow/implement.md section 1 lacks %q", want)
		}
	}
	if strings.Contains(load, "print its absolute path") {
		t.Error("skills/flow/implement.md section 1 still prints the learning note's path")
	}
	isolate := strings.Join(strings.Fields(section(t, "skills/flow/implement.md", "2. Isolate the workspace")), " ")
	for _, want := range []string{
		"On a fix run whose state record's `worktrees` map is not empty, that map is this run's resolved set",
		"**Then run `flow workspace-id <name>`",
	} {
		if !strings.Contains(isolate, want) {
			t.Errorf("skills/flow/implement.md section 2 lacks %q", want)
		}
	}
}
