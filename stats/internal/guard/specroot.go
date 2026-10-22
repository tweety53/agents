package guard

import (
	"fmt"
	"io"
	"os"
)

// specRootLeaf is the Go twin of spec_root_leaf from scripts/lib/spec-root.sh,
// retired by KAN-1020 once no script sourced it; its header as it stood at
// 8c54d6d3 carries the why: `spectre` when dir/spectre/changes/ exists, `openspec`
// when only dir/openspec/changes/ does, `spectre` when neither does; both
// present picks `spectre`, and says so on stderr only while openspec/changes/
// holds an entry other than archive/ -- a tree of frozen history alone hides
// no open change.
func specRootLeaf(dir string, stderr io.Writer) string {
	hasSpectre, hasOpenspec := isDir(dir+"/spectre/changes"), isDir(dir+"/openspec/changes")
	if hasSpectre && hasOpenspec {
		if !holdsOpenChange(dir + "/openspec/changes") {
			return "spectre"
		}
		fmt.Fprintf(stderr, "spec-root: %s carries both spectre/changes/ and openspec/changes/ — using spectre/; every change under openspec/changes/ is invisible to this guard until that tree is moved\n", dir)
		return "spectre"
	}
	if hasOpenspec {
		return "openspec"
	}
	return "spectre"
}

// holdsOpenChange reports whether changes/ holds any entry but archive/.
func holdsOpenChange(changes string) bool {
	entries, _ := os.ReadDir(changes)
	for _, e := range entries {
		if e.Name() != "archive" {
			return true
		}
	}
	return false
}
