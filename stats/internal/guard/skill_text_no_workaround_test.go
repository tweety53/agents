package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// A skill never instructs an agent to hand-edit guard-written output or to get past a
// blocking hook or guard: the agent reports BLOCKED and its parent fixes the tooling.
var skillWorkaroundRe = regexp.MustCompile(`(?i)\b(hand|manual)(ly)?[- ](merge|edit|patch)\b|\bedit\w* (the |a )?(visual-)?(manifest|ledger)\b[^.\n]*\bby hand|\b(work(ing)?|rout(e|ing)|get(ting)?) (a)?round (the |a |that )?(blocking )?(hook|guard)\b`)

func skillWorkaroundHits(body string) []string {
	return skillWorkaroundRe.FindAllString(body, -1)
}

func TestSkillTextHasNoWorkaroundInstruction(t *testing.T) {
	t.Parallel()
	if got := skillWorkaroundHits("Then hand-merge the manifest, or work around the hook."); len(got) != 2 {
		t.Fatalf("planted instruction not caught: %v", got)
	}
	root := filepath.Join("..", "..", "..")
	for _, dir := range []string{"skills", "rules"} {
		filepath.WalkDir(filepath.Join(root, dir), func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || strings.HasSuffix(p, "-rationale.md") || !(strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".mdc")) {
				return nil
			}
			b, _ := os.ReadFile(p)
			if hits := skillWorkaroundHits(string(b)); len(hits) > 0 {
				t.Errorf("%s: %v", p, hits)
			}
			return nil
		})
	}
}
