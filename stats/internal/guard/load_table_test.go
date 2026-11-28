package guard

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestPhaseLoadTableMatchesTheTree: skills/README.md's **Phase load table** names only files that
// exist, and names every file a conditional load set lists.
func TestPhaseLoadTableMatchesTheTree(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "skills/README.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, table, ok := strings.Cut(string(b), "\n## Phase load table\n")
	if !ok {
		t.Fatal("skills/README.md: no ## Phase load table section")
	}
	if i := strings.Index(table, "\n## "); i >= 0 {
		table = table[:i]
	}
	named := map[string]bool{}
	for _, m := range regexp.MustCompile("`(skills/[a-z0-9/-]+\\.md)`").FindAllStringSubmatch(table, -1) {
		named[m[1]] = true
		if _, err := os.Stat(filepath.Join(root, m[1])); err != nil {
			t.Errorf("Phase load table names %s, which does not exist", m[1])
		}
	}
	for _, set := range conditionalLoadSets(root) {
		for _, f := range set.files {
			rel := strings.TrimPrefix(f, root+"/")
			if !named[rel] {
				t.Errorf("Phase load table omits %s (load set %q)", rel, set.condition)
			}
		}
	}
}
