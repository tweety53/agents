package guard

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// checkIdenticalBaselines is scripts/check-identical-baselines.sh: that
// script's header comment is the contract. PNGs of one directory with the
// same bytes are named, one group per line.
func init() { Registry["check-identical-baselines"] = checkIdenticalBaselines }

func checkIdenticalBaselines(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-identical-baselines"
	if len(args) != 0 {
		fmt.Fprintf(stderr, "%s: usage: resolve-visual-screenshots.sh … | check-identical-baselines.sh\n", prog)
		return 2
	}
	in := env.Stdin
	if in == nil {
		in = strings.NewReader("")
	}
	// directory → content hash → base names
	groups := map[string]map[[32]byte][]string{}
	n := 0
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		path := strings.TrimSpace(sc.Text())
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, path, err)
			return 2
		}
		dir := filepath.Dir(path)
		if groups[dir] == nil {
			groups[dir] = map[[32]byte][]string{}
		}
		sum := sha256.Sum256(data)
		groups[dir][sum] = append(groups[dir][sum], filepath.Base(path))
		n++
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", prog, err)
		return 2
	}
	if n == 0 {
		fmt.Fprintf(stderr, "%s: no PNG paths on stdin\n", prog)
		return 2
	}
	var lines []string
	for dir, byHash := range groups {
		for _, names := range byHash {
			if len(names) > 1 {
				sort.Strings(names)
				lines = append(lines, fmt.Sprintf("identical: %s: %s", dir, strings.Join(names, " ")))
			}
		}
	}
	if len(lines) == 0 {
		fmt.Fprintf(stdout, "IDENTICAL-NONE: %d PNGs\n", n)
		return 0
	}
	// "identical: <dir>: <first name> …" sorts by directory, then first name.
	sort.Strings(lines)
	fmt.Fprintln(stdout, strings.Join(lines, "\n"))
	return 1
}
