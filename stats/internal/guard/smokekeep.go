package guard

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// flow-guard smoke-keep — the heavy tests the files a branch touched keep in a default test run.
//
// Usage: flow-guard smoke-keep [<path prefix>=<kept path>]...
//
// Reads the paths the branch changed since HEAD's merge base on origin/main — committed,
// staged, unstaged or untracked, the diff `flow-guard unchanged` reads — and prints, sorted,
// space-separated and each once on one line, the paths a default run keeps for them:
//   - the kept path of every argument whose path prefix the changed path starts with;
//   - a changed `_test.go` file itself;
//   - for any other changed `.go` file, every `_test.go` file of its directory, in the working
//     tree, whose name pairs with it — and, when none does, one stderr line naming it, with
//     nothing kept for it; a directory that exists but cannot be read gets one stderr line naming
//     the read error instead.
//
// Two names pair when one normal form is a prefix of the other and neither is empty: lowercased,
// `_` removed, the extension dropped; on the test side also the `test` suffix and a leading
// `check`. Nothing changed prints one empty line.
//
// Exit codes:
//
//	0  the keep list was printed
//	2  usage: an argument that is not <prefix>=<path> with both sides non-empty, or a git call
//	   that failed (no repository, no origin/main)
func init() { Registry["smoke-keep"] = smokeKeep }

// smokeKeep is `flow-guard smoke-keep`: the comment above is the contract.
// ponytail: the list is space-separated, so a path holding a space splits in
// FLOW_SMOKE_KEEP; no tracked path has one -- quote the list if one appears.
func smokeKeep(args []string, env Env, stdout, stderr io.Writer) int {
	type row struct{ prefix, kept string }
	var rows []row
	for _, a := range args {
		prefix, kept, found := strings.Cut(a, "=")
		if !found || prefix == "" || kept == "" {
			fmt.Fprintln(stderr, "usage: flow-guard smoke-keep [<path prefix>=<kept path>]...")
			return 2
		}
		rows = append(rows, row{prefix, kept})
	}
	_, paths, ok := changedSinceBase(env, "smoke-keep", stderr)
	if !ok {
		return 2
	}
	top := ""
	if len(paths) > 0 {
		gitBin, _ := panelResolveGit(env, "smoke-keep", stderr)
		cmd := exec.Command(gitBin, "rev-parse", "--show-toplevel")
		cmd.Dir = env.Dir
		out, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(stderr, "smoke-keep: git rev-parse --show-toplevel failed: %v\n", err)
			return 2
		}
		top = strings.TrimSpace(string(out))
	}
	kept := map[string]bool{}
	for _, p := range paths {
		for _, r := range rows {
			if strings.HasPrefix(p, r.prefix) {
				kept[r.kept] = true
			}
		}
		switch {
		case strings.HasSuffix(p, "_test.go"):
			kept[p] = true
		case strings.HasSuffix(p, ".go"):
			paired := false
			entries, err := os.ReadDir(filepath.Join(top, path.Dir(p)))
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintf(stderr, "smoke-keep: %s: %v — nothing kept for it\n", p, err)
				continue
			}
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), "_test.go") && smokeKeepPairs(path.Base(p), e.Name()) {
					kept[path.Join(path.Dir(p), e.Name())] = true
					paired = true
				}
			}
			if !paired {
				fmt.Fprintf(stderr, "smoke-keep: %s pairs with no test file — nothing kept for it\n", p)
			}
		}
	}
	list := make([]string, 0, len(kept))
	for k := range kept {
		list = append(list, k)
	}
	sort.Strings(list)
	fmt.Fprintln(stdout, strings.Join(list, " "))
	return 0
}

// smokeKeepPairs reports whether source file src and test file test pair.
func smokeKeepPairs(src, test string) bool {
	norm := func(n string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimSuffix(n, path.Ext(n))), "_", "")
	}
	s := norm(src)
	t := strings.TrimPrefix(strings.TrimSuffix(norm(test), "test"), "check")
	return s != "" && t != "" && (strings.HasPrefix(s, t) || strings.HasPrefix(t, s))
}
