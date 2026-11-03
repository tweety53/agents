package guard

import (
	"bufio"
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// checkGuardCalls is `flow-guard check-guard-calls`: it keeps every guard
// call site spelled the one way a guard runs — `flow-guard <name>`, which resolves on PATH to the `flow-guard` `setup.sh global` installs over this checkout's scripts/flow-guard.
//
// Argument-free and self-scoped: the scan root is this repository's own root
// (FLOW_GUARD_REPO_ROOT), so no call site can narrow it.
// CHECK_GUARD_CALLS_ROOT is an explicit, opt-in override honoured only when
// set, so the guard's Go tests can point it at a fixture tree — never set it
// for a normal invocation.
//
// It scans the text of skills/**/*.md, commands-claude/*.md, rules/*,
// hooks/*.py, templates/*, .flow/project.md, CLAUDE.md and AGENTS.md — never
// docs/ or spectre/, which are history and keep the spelling a guard had
// when they were written — and reports one `path:line: message` line per
// finding:
//
//	A. Every `flow-guard <name>` names a guard.Registry entry. A typo'd name
//	   is a call that exits "unknown guard" at the moment it was meant to
//	   check something, and nothing reads the call before then. A `<…>`
//	   placeholder, a `…*` glob (`flow-guard check-*`) and a `…` elision
//	   (`flow-guard check-…`) name no guard and are skipped. Only a hyphenated
//	   word is read as a name, so prose such as `flow-guard not found on PATH`
//	   is not a call. An unhyphenated word is a call only when it is within
//	   two edits of an unhyphenated guard (`unchanged`, `workspace`) without
//	   being one — a typo of it.
//	B. No `<name>.sh` where `<name>` is a registry entry. The per-guard .sh
//	   shims are gone; a sentence still spelling one names a file that does
//	   not exist, and an agent running it gets 127. Only registered names
//	   count: `setup.sh` and a panel reproducer's `repro.sh` are not guards.
//	   The same rule reads every string literal of stats/**/*.go outside
//	   *_test.go: a guard's printed NEXT line, usage line or remedy that
//	   spells a deleted shim tells its caller to run a file that does not
//	   exist. Comments are not scanned — a literal is what reaches a caller.
//	C. Every `check-*` guard.Registry entry has its Go test,
//	   stats/internal/guard/<name, - replaced by _>_test.go (kan-387, moved
//	   here from the deleted run-guard-tests): a guard with no test is a
//	   check nothing proves bites, a convention review-panel judgment alone
//	   never held. The finding names the guard and the missing file. A root
//	   without stats/internal/guard — a rule A/B test fixture — skips rule C.
//
// Then GUARD-CALLS-OK or GUARD-CALLS-INVALID. Exit 0 clean, 1 any finding,
// 2 cannot answer (a refusal writes its reason to stderr and nothing to
// stdout).
func init() {
	Registry["check-guard-calls"] = checkGuardCalls
}

var (
	gcCall  = regexp.MustCompile(`flow-guard ([A-Za-z0-9_]+(?:-[A-Za-z0-9_]*)+)(\*|…)?`)
	gcShell = regexp.MustCompile(`([A-Za-z0-9_-]+)\.sh\b`)
	gcBare  = regexp.MustCompile(`flow-guard ([A-Za-z0-9_]+)\b(?:[^-]|$)`)
)

func checkGuardCalls(_ []string, env Env, stdout, stderr io.Writer) int {
	root, set := env.LookupEnv("CHECK_GUARD_CALLS_ROOT")
	if !set {
		root = env.Getenv("FLOW_GUARD_REPO_ROOT")
	}
	if root == "" {
		fmt.Fprint(stderr, "check-guard-calls: no repository root — run flow-guard check-guard-calls, which sets FLOW_GUARD_REPO_ROOT\n")
		return 2
	}
	files, err := gcFiles(root)
	if err != nil {
		fmt.Fprintf(stderr, "check-guard-calls: %v\n", err)
		return 2
	}
	var out strings.Builder
	n := 0
	for _, rel := range files {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			fmt.Fprintf(stderr, "check-guard-calls: cannot read %s: %v\n", rel, err)
			return 2
		}
		sc := bufio.NewScanner(bytes.NewReader(b))
		sc.Buffer(nil, 1<<24)
		for line := 1; sc.Scan(); line++ {
			for _, m := range gcCall.FindAllStringSubmatch(sc.Text(), -1) {
				if m[2] == "" && Registry[m[1]] == nil {
					fmt.Fprintf(&out, "%s:%d: flow-guard %s names no registered guard — fix the name or register the guard (rule A)\n", rel, line, m[1])
					n++
				}
			}
			for _, m := range gcBare.FindAllStringSubmatch(sc.Text(), -1) {
				if gcTypo(m[1]) {
					fmt.Fprintf(&out, "%s:%d: flow-guard %s names no registered guard — fix the name or register the guard (rule A)\n", rel, line, m[1])
					n++
				}
			}
			for _, m := range gcShell.FindAllStringSubmatch(sc.Text(), -1) {
				if Registry[m[1]] != nil {
					fmt.Fprintf(&out, "%s:%d: %s.sh is the deleted shim spelling — write flow-guard %s (rule B)\n", rel, line, m[1], m[1])
					n++
				}
			}
		}
	}
	gn, err := gcGoLiterals(&out, root)
	if err != nil {
		fmt.Fprintf(stderr, "check-guard-calls: %v\n", err)
		return 2
	}
	n += gn
	if isDir(filepath.Join(root, "stats", "internal", "guard")) {
		n += gcCompanions(&out, root, slices.Collect(maps.Keys(Registry)))
	}
	fmt.Fprint(stdout, out.String())
	if n > 0 {
		fmt.Fprintf(stdout, "GUARD-CALLS-INVALID: %s — %d violation(s)\n", root, n)
		return 1
	}
	fmt.Fprintf(stdout, "GUARD-CALLS-OK: %s — %d file(s) scanned\n", root, len(files))
	return 0
}

// gcGoLiterals is rule B over the string literals of every non-test .go file
// under root/stats. It returns the finding count.
func gcGoLiterals(out io.Writer, root string) (int, error) {
	stats := filepath.Join(root, "stats")
	if !isDir(stats) {
		return 0, nil
	}
	n := 0
	err := filepath.WalkDir(stats, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == "testdata") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		fset := token.NewFileSet()
		file := fset.AddFile(rel, -1, len(src))
		var sc scanner.Scanner
		sc.Init(file, src, nil, 0)
		for {
			pos, tok, lit := sc.Scan()
			if tok == token.EOF {
				return nil
			}
			if tok != token.STRING {
				continue
			}
			for _, m := range gcShell.FindAllStringSubmatch(lit, -1) {
				if Registry[m[1]] != nil {
					fmt.Fprintf(out, "%s:%d: %s.sh is the deleted shim spelling — write flow-guard %s (rule B)\n", rel, fset.Position(pos).Line, m[1], m[1])
					n++
				}
			}
		}
	})
	return n, err
}

// gcCompanions is rule C over names: one finding per check-* name with no
// stats/internal/guard/<name, - as _>_test.go under root. It returns the
// finding count.
func gcCompanions(out io.Writer, root string, names []string) int {
	slices.Sort(names)
	n := 0
	for _, name := range names {
		if !strings.HasPrefix(name, "check-") {
			continue
		}
		test := "stats/internal/guard/" + strings.ReplaceAll(name, "-", "_") + "_test.go"
		if !isFile(filepath.Join(root, test)) {
			fmt.Fprintf(out, "%s:0: guard %s has no Go test — add it beside the guard (rule C)\n", test, name)
			n++
		}
	}
	return n
}

// gcFiles lists the scanned files, relative to root, in a stable order. A
// missing directory or file is skipped — a fixture carries only what it tests.
func gcFiles(root string) ([]string, error) {
	var files []string
	skills := filepath.Join(root, "skills")
	if isDir(skills) {
		err := filepath.WalkDir(skills, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(p, ".md") {
				rel, _ := filepath.Rel(root, p)
				files = append(files, rel)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for _, pat := range []string{"commands-claude/*.md", "rules/*", "hooks/*.py", "templates/*", ".flow/project.md", "CLAUDE.md", "AGENTS.md"} {
		m, err := filepath.Glob(filepath.Join(root, pat))
		if err != nil {
			return nil, err
		}
		for _, p := range m {
			if isFile(p) {
				rel, _ := filepath.Rel(root, p)
				files = append(files, rel)
			}
		}
	}
	slices.Sort(files)
	return files, nil
}

// gcTypo reports whether an unhyphenated word is not a registered guard but
// lies within two edits of an unhyphenated one -- a typo of that guard,
// where any farther word is prose such as `flow-guard not found`.
func gcTypo(w string) bool {
	if Registry[w] != nil {
		return false
	}
	for name := range Registry {
		if !strings.Contains(name, "-") && gcDistance(w, name) <= 2 {
			return true
		}
	}
	return false
}

// gcDistance is the Levenshtein distance between a and b.
func gcDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}
