package repotest

import (
	"strings"
	"testing"
)

// TestSmallAppendNamesNoModel pins that the small append's sentence in
// review-panel-late-fix.md names no model: the late-fix row of **Allowed
// pairs** decides its pair (self-review-kan-1110 F3).
func TestSmallAppendNamesNoModel(t *testing.T) {
	t.Parallel()
	text := repoText(t, "skills/flow/review-panel-late-fix.md")
	i := strings.Index(text, "never the rerun pair")
	if i < 0 {
		t.Fatal("review-panel-late-fix.md has no sentence holding `never the rerun pair`")
	}
	start := strings.LastIndex(text[:i], ". ") + 2
	end := len(text)
	if j := strings.Index(text[i:], ". "); j >= 0 {
		end = i + j
	}
	sentence := text[start:end]
	if !strings.Contains(sentence, "the delta is new source") {
		t.Errorf("the small append's sentence lost its reason clause: %q", sentence)
	}
	for _, m := range []string{"`opus`", "`sonnet`"} {
		if strings.Contains(sentence, m) {
			t.Errorf("the small append's sentence names %s; the late-fix row decides: %q", m, sentence)
		}
	}
}
