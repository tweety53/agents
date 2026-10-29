package repotest

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	// A role row of the **Allowed pairs** table: `| `<id>` | `<pair>`…`.
	roleRow = regexp.MustCompile("(?m)^\\| `([a-z0-9-]+)` \\| `[^`]+/")
	// A Role line a skill names, as a line of its own or quoted: `Role: <id>`.
	roleLine = regexp.MustCompile("(?m)(?:^|`)Role: ([^`<\\s][^`\\n]*?)(?:`|$)")
)

// TestEveryRoleLineIsAPolicyRow: every `Role: <id>` line a skill names is a
// role of the **Allowed pairs** table (skills/flow-contracts/model-policy.md),
// so no dispatch template carries a role hooks/enforce-agent-baseline.py
// refuses as unknown.
func TestEveryRoleLineIsAPolicyRow(t *testing.T) {
	root := repoRoot(t)
	b, err := os.ReadFile(filepath.Join(root, "skills/flow-contracts/model-policy.md"))
	if err != nil {
		t.Fatal(err)
	}
	roles := map[string]bool{}
	for _, m := range roleRow.FindAllStringSubmatch(string(b), -1) {
		roles[m[1]] = true
	}
	if len(roles) == 0 {
		t.Fatal("model-policy.md carries no Allowed pairs role row")
	}
	seen := 0
	err = filepath.WalkDir(filepath.Join(root, "skills"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		for _, m := range roleLine.FindAllStringSubmatch(string(body), -1) {
			seen++
			if !roles[m[1]] {
				rel, _ := filepath.Rel(root, p)
				t.Errorf("%s names `Role: %s`, no role of the Allowed pairs table", rel, m[1])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if seen == 0 {
		t.Fatal("no `Role:` line found under skills/")
	}
}
