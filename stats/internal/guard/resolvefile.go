package guard

import (
	"os"
	"path/filepath"
	"strings"
)

// resolveFile is the Go twin of scripts/lib/resolve-file.sh, which stays the
// source of truth while check-dev-stack-fresh.sh, check-plan-shape.sh
// and plan-dispatch-bundles.sh still source it;
// its header carries the reasoning (KAN-73's F1, F9, KAN-201's F16), and
// TestResolveFileParity fails when the two print or return differently for
// the same path.

// resolveFile is lib/resolve-file.sh's resolve_file:
// follow the leaf's symlinks (at most 40 hops), then resolve its directory
// physically, as `cd -P`. false where the bash function returns non-zero.
func resolveFile(p string) (string, bool) {
	for p != "/" && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	for hops := 1; ; hops++ {
		fi, err := os.Lstat(p)
		if err != nil || fi.Mode()&os.ModeSymlink == 0 {
			break
		}
		if hops > 40 {
			return "", false
		}
		target, err := os.Readlink(p)
		if err != nil {
			return "", false
		}
		switch dir := gdcDirname(p); {
		case strings.HasPrefix(target, "/"):
			p = target
		case dir == "/":
			p = "/" + target
		default:
			p = dir + "/" + target
		}
	}
	base := p[strings.LastIndex(p, "/")+1:]
	if base == "." || base == ".." {
		return physicalDir(p)
	}
	dir, ok := physicalDir(gdcDirname(p))
	if !ok {
		return "", false
	}
	if dir == "/" {
		return "/" + base, true
	}
	return dir + "/" + base, true
}

// physicalDir is `cd -P -- dir && pwd -P`: always absolute, a relative
// dir taken from the process's physical cwd, as the shell's.
func physicalDir(dir string) (string, bool) {
	if !strings.HasPrefix(dir, "/") {
		cwd, err := os.Getwd()
		if err != nil {
			return "", false
		}
		dir = cwd + "/" + dir
	}
	real, err := filepath.EvalSymlinks(dir)
	if err != nil || !isDir(real) {
		return "", false
	}
	return real, true
}
