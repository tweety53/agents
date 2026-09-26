package guard

import (
	"fmt"
	"io"
	"os"
)

// specRootLeaf is scripts/lib/spec-root.sh's spec_root_leaf, whose header is
// canonical for why: `spectre` when dir/spectre/changes/ exists, `openspec`
// when only dir/openspec/changes/ does, `spectre` when neither does; both
// present picks `spectre` and says so on stderr.
func specRootLeaf(dir string, stderr io.Writer) string {
	hasSpectre, hasOpenspec := isDir(dir+"/spectre/changes"), isDir(dir+"/openspec/changes")
	if hasSpectre && hasOpenspec {
		fmt.Fprintf(stderr, "spec-root: %s carries both spectre/changes/ and openspec/changes/ — using spectre/; every change under openspec/changes/ is invisible to this guard until that tree is moved\n", dir)
		return "spectre"
	}
	if hasOpenspec {
		return "openspec"
	}
	return "spectre"
}

// isDir and isFile are bash's `[ -d ]` and `[ -f ]`: they follow symlinks.
func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular()
}
