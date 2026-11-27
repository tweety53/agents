package guard

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestEdrBeginKeyMatchesHook holds edrBeginKey equal to hooks/enforce-agent-baseline.py's
// BEGIN_KEY (ROW_KEY plus the planner keys), the two twins of one key list.
func TestEdrBeginKeyMatchesHook(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile("../../../hooks/enforce-agent-baseline.py")
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	lit := regexp.MustCompile(`r"([^"]*)"`)
	block := regexp.MustCompile(`(?s)\nROW_KEY = \((.*?)\n\)`).FindStringSubmatch(s)
	line := regexp.MustCompile(`\nBEGIN_KEY = re\.compile\((.*)\)\n`).FindStringSubmatch(s)
	if block == nil || line == nil {
		t.Fatal("ROW_KEY or BEGIN_KEY not found in the hook")
	}
	var row strings.Builder
	for _, m := range lit.FindAllStringSubmatch(block[1], -1) {
		row.WriteString(m[1])
	}
	parts := lit.FindAllStringSubmatch(line[1], -1)
	if len(parts) != 2 || !strings.Contains(line[1], "+ ROW_KEY +") {
		t.Fatalf("BEGIN_KEY shape changed: %s", line[1])
	}
	hook := "^" + parts[0][1] + row.String() + parts[1][1] + "$"
	if got := edrBeginKey.String(); got != hook {
		t.Errorf("edrBeginKey drifted from the hook's BEGIN_KEY:\n go: %s\nhook: %s", got, hook)
	}
}
