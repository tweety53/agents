package guard

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// checkSidecarReach is scripts/check-sidecar-reach.sh: that script's header
// comment is the contract. Every `<spec>.mockups` sidecar line under the root
// must name a screenshot some string literal in <spec> takes.
func init() { Registry["check-sidecar-reach"] = checkSidecarReach }

var (
	// Each quote kind is scanned on its own: in one alternation a stray
	// apostrophe or backtick in a comment pairs with the next quote and
	// swallows the real literals after it (gymie-playwright's entry.spec.ts).
	srQuoted   = []*regexp.Regexp{regexp.MustCompile(`'([^'\n]*)'`), regexp.MustCompile(`"([^"\n]*)"`)}
	srBacktick = regexp.MustCompile("`([^`]*)`")
	srHole     = regexp.MustCompile(`\$\{[^}]*\}`)
)

// srReached reports whether name equals a literal of spec, or fully matches a
// backtick template literal whose `${…}` holes each stand for one or more
// characters.
func srReached(name string, spec []byte) bool {
	for _, re := range srQuoted {
		for _, m := range re.FindAllSubmatch(spec, -1) {
			if string(m[1]) == name {
				return true
			}
		}
	}
	for _, m := range srBacktick.FindAllSubmatch(spec, -1) {
		tpl := string(m[1])
		if tpl == name {
			return true
		}
		if strings.Contains(tpl, "${") {
			pieces := srHole.Split(tpl, -1)
			for i, p := range pieces {
				pieces[i] = regexp.QuoteMeta(p)
			}
			if regexp.MustCompile("^" + strings.Join(pieces, ".+") + "$").MatchString(name) {
				return true
			}
		}
	}
	return false
}

func checkSidecarReach(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-sidecar-reach"
	if len(args) != 1 && len(args) != 2 {
		fmt.Fprintf(stderr, "%s: usage: check-sidecar-reach.sh <capture root> [<mockups root>]\n", prog)
		return 2
	}
	for _, dir := range args {
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			fmt.Fprintf(stderr, "%s: not a directory: %s\n", prog, dir)
			return 2
		}
	}
	root, mockups := args[0], ""
	if len(args) == 2 {
		mockups = args[1]
	}
	var misses []string
	lines, sidecars := 0, 0
	// WalkDir visits in lexical order, so sidecars come sorted by path.
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "node_modules", ".git", ".worktrees":
				return filepath.SkipDir
			}
			return nil
		}
		specPath, ok := strings.CutSuffix(path, ".mockups")
		if !ok {
			return nil
		}
		sidecars++
		rel, _ := filepath.Rel(root, path)
		relSpec, _ := filepath.Rel(root, specPath)
		spec, specErr := os.ReadFile(specPath)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for n := 1; sc.Scan(); n++ {
			fields := strings.Fields(sc.Text())
			if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
				continue
			}
			lines++
			name := fields[0]
			switch {
			case specErr != nil:
				misses = append(misses, fmt.Sprintf("%s:%d: %s — %s does not exist", rel, n, name, relSpec))
			case !srReached(name, spec):
				misses = append(misses, fmt.Sprintf("%s:%d: %s — taken by no screenshot in %s", rel, n, name, relSpec))
			case mockups != "" && len(fields) > 1:
				if _, err := os.Stat(filepath.Join(mockups, fields[1]+".png")); err != nil {
					misses = append(misses, fmt.Sprintf("%s:%d: %s — frame %s has no %s/%s.png", rel, n, name, fields[1], mockups, fields[1]))
				}
			}
		}
		return sc.Err()
	})
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", prog, err)
		return 2
	}
	if len(misses) > 0 {
		fmt.Fprintln(stdout, strings.Join(misses, "\n"))
		return 1
	}
	fmt.Fprintf(stdout, "SIDECAR-REACH-OK: %d lines in %d sidecars\n", lines, sidecars)
	return 0
}
