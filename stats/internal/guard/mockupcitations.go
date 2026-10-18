package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// checkMockupCitations is scripts/check-mockup-citations.sh: that script's
// header comment is the contract. Every `<mockups root>/<frame id>.png` token
// a plan file carries must name a file under <project root>.
func init() { Registry["check-mockup-citations"] = checkMockupCitations }

func checkMockupCitations(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-mockup-citations"
	if len(args) < 3 {
		fmt.Fprintf(stderr, "%s: usage: check-mockup-citations.sh <project root> <mockups root> <file>...\n", prog)
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
