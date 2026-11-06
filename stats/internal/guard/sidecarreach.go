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

// The contract below moved here from the header of the deleted
// scripts/check-sidecar-reach.sh shim (KAN-1058).
//
// flow-guard check-sidecar-reach — pre-flight every `.mockups` sidecar line against
// the capture spec beside it, before the capture runs (kan-998: stale sidecar
// lines surfaced only at flow-guard compose-mockup-frames, after a 13-minute capture).
//
// Usage: flow-guard check-sidecar-reach <capture root | sidecar> [<mockups root>]
//
// Walks <capture root>, never descending into `node_modules`, `.git` or
// `.worktrees`. A SIDECAR is a file ending `.mockups`; its spec is that path
// without the suffix. A sidecar line is `<screenshot name> <frame id>`; blank
// and `#` lines are skipped. A sidecar given in place of <capture root> is
// the only one judged, every printed path relative to its directory.
//
// A name is REACHED when it equals a string literal of the spec ('…', "…" or
// `…`), or fully matches a backtick template literal containing `${`, each
// `${…}` hole standing for one or more characters
// (`cardio-guide-${frame}.png` reaches `cardio-guide-G1.png`).
// Stated limit: a name built outside any literal (concatenation, a computed
// variable), or through a `${…}` hole holding a `}` (the hole ends at its
// first `}`), reads as unreached.
//
// With <mockups root>, a reached line carrying a second field <frame id> must
// also name a frame that exists: <mockups root>/<frame id>.png (kan-1007). A
// line with one field is not judged on its frame.
//
// Prints to stdout, paths relative to <capture root> (or the sidecar's directory), sorted by sidecar then
// line:
//   SIDECAR-REACH-OK: <n> lines in <m> sidecars                      exit 0
//   <sidecar>:<line>: <name> — taken by no screenshot in <spec>      exit 1
//   <sidecar>:<line>: <name> — <spec> does not exist                 exit 1
//   <sidecar>:<line>: <name> — frame <id> has no <mockups root>/<id>.png  exit 1
// (<mockups root> printed as given).
//
// Exit 0 every line reached, 1 at least one not, 2 cannot answer (a usage
// error, a capture root that is neither a directory nor a sidecar, a mockups
// root that is not a directory, an unreadable
// sidecar) — stdout
// empty, the cause on stderr.
//
// Go, in flow-guard (stats/internal/guard/sidecarreach.go).

// checkSidecarReach is `flow-guard check-sidecar-reach`: the comment above is the contract. Every `<spec>.mockups` sidecar line under the root
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
		fmt.Fprintf(stderr, "%s: usage: flow-guard check-sidecar-reach <capture root | sidecar> [<mockups root>]\n", prog)
		return 2
	}
	base := args[0]
	for i, dir := range args {
		st, err := os.Stat(dir)
		if i == 0 && err == nil && st.Mode().IsRegular() && strings.HasSuffix(dir, ".mockups") {
			base = filepath.Dir(dir)
			continue
		}
		if err != nil || !st.IsDir() {
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
		rel, _ := filepath.Rel(base, path)
		relSpec, _ := filepath.Rel(base, specPath)
		spec, specErr := os.ReadFile(specPath)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for n := 1; sc.Scan(); n++ {
			fields := strings.Fields(sc.Text())
			// A `frame` directive (composemockupframes.go) names no screenshot.
			if len(fields) == 0 || strings.HasPrefix(fields[0], "#") || fields[0] == "frame" {
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
