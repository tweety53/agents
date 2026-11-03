package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-mockup-citations.sh shim (KAN-1058).
//
// flow-guard check-mockup-citations — refuse a plan citing a mockup frame no file in
// the project's mockups root holds (KAN-558, KAN-565: kan-469's design.md and
// tasks.md cited P6.png, no such frame existed, and the gap surfaced only when
// visual verification skipped the comparison).
//
// Usage: flow-guard check-mockup-citations <project root> <mockups root> <file>...
//
// <project root> is the directory `.flow/project.md` lives in (a worktree's
// root); <mockups root> the `mockups` value of its `## visual verification`
// table, relative to <project root>, a trailing `/` ignored
// (skills/flow-contracts/project-configuration-visual.md); each <file> a plan
// artifact — <changeRoot>/design.md and <changeRoot>/tasks.md.
//
// A CITATION is a `<mockups root>/<frame id>.png` token, <frame id> made of
// letters, digits, `_` and `-`, not preceded by a letter, digit, `_`, `.`,
// `/` or `-` — so `other/<mockups root>/X.png` is no citation. A bare
// `<frame id>.png` is not one either: plans name capture screenshots that
// way too, and nothing in the token tells the two apart. Writing plans
// (skills/flow/brainstorm-planner.md, section D) has a plan cite each frame
// by its path for that reason.
//
// Prints ONE verdict line to stdout:
//   MOCKUP-CITATIONS-RESOLVED: <n> cited
//   MOCKUP-CITATIONS-MISSING: <file>:<line> <citation>; …
//
// Exit 0 every citation names a regular file, 1 at least one does not,
// 2 cannot answer (a usage error, a <mockups root> that is not a directory
// under <project root>, an unreadable <file>).
//
// Go, in flow-guard (stats/internal/guard/mockupcitations.go).

// checkMockupCitations is `flow-guard check-mockup-citations`: the comment above is the contract. Every `<mockups root>/<frame id>.png` token
// a plan file carries must name a file under <project root>.
func init() { Registry["check-mockup-citations"] = checkMockupCitations }

func checkMockupCitations(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-mockup-citations"
	if len(args) < 3 {
		fmt.Fprintf(stderr, "%s: usage: flow-guard check-mockup-citations <project root> <mockups root> <file>...\n", prog)
		return 2
	}
	root := strings.TrimSuffix(args[1], "/")
	if fi, err := os.Stat(filepath.Join(args[0], root)); err != nil || !fi.IsDir() {
		fmt.Fprintf(stderr, "%s: mockups root %s is not a directory under %s\n", prog, root, args[0])
		return 2
	}
	cite := regexp.MustCompile(`(^|[^A-Za-z0-9_./-])(` + regexp.QuoteMeta(root) + `/[A-Za-z0-9_-]+\.png)`)
	cited := 0
	var missing []string
	for _, file := range args[2:] {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, file, err)
			return 2
		}
		for i, line := range strings.Split(string(data), "\n") {
			for _, m := range cite.FindAllStringSubmatch(line, -1) {
				cited++
				if fi, err := os.Stat(filepath.Join(args[0], m[2])); err != nil || !fi.Mode().IsRegular() {
					missing = append(missing, fmt.Sprintf("%s:%d %s", file, i+1, m[2]))
				}
			}
		}
	}
	if len(missing) > 0 {
		fmt.Fprintf(stdout, "MOCKUP-CITATIONS-MISSING: %s\n", strings.Join(missing, "; "))
		return 1
	}
	fmt.Fprintf(stdout, "MOCKUP-CITATIONS-RESOLVED: %d cited\n", cited)
	return 0
}
