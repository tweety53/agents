// paths.go holds the path helpers several guards share.

package guard

import (
	"os"
	"path/filepath"
	"strings"
)

// isDir and isFile are bash's `[ -d ]` and `[ -f ]`: they follow symlinks.
func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func isFile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.Mode().IsRegular()
}

// pcAbs is a worktree argument as the panel guards read it: joined to the
// working directory when relative, then cleaned.
func pcAbs(env Env, p string) string {
	if !filepath.IsAbs(p) {
		p = filepath.Join(env.Dir, p)
	}
	return filepath.Clean(p)
}

// smcAbs is p as the bash resolved it: relative to the working directory.
func smcAbs(env Env, p string) string {
	if strings.HasPrefix(p, "/") || env.Dir == "" {
		return p
	}
	return env.Dir + "/" + p
}

// gdcDirname is dirname(1).
func gdcDirname(p string) string {
	if p != "" && strings.Trim(p, "/") == "" {
		return "/"
	}
	p = strings.TrimRight(p, "/")
	i := strings.LastIndex(p, "/")
	if i < 0 {
		return "."
	}
	d := strings.TrimRight(p[:i+1], "/")
	if d == "" {
		return "/"
	}
	return d
}

// hasDotDotSegment is the bash `..|../*|*/..|*/../*` case pattern.
func hasDotDotSegment(p string) bool {
	return p == ".." || strings.HasPrefix(p, "../") || strings.HasSuffix(p, "/..") || strings.Contains(p, "/../")
}

// siblingRoot is where every worktree of the repository whose main checkout
// is main lives: <dirname main>/<basename main>-worktrees (design.md:
// sibling-worktrees-layout). Kickoff creates there, this guard accepts only
// there, and migrate-worktrees moves there — one rule, one spelling.
func siblingRoot(main string) string {
	return filepath.Join(filepath.Dir(main), filepath.Base(main)+"-worktrees")
}
