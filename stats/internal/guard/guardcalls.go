package guard

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// checkGuardCalls is `flow-guard check-guard-calls`: it keeps every guard
// call site spelled the one way a guard runs — `flow-guard <name>`, which
// resolves on PATH to the shim `setup.sh global` installs over this
// checkout's scripts/flow-guard.
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
//	   is not a call; the two unhyphenated guards, `unchanged` and
//	   `workspace`, are therefore not checked for typos.
//	B. No `<name>.sh` where `<name>` is a registry entry. The per-guard .sh
//	   shims are gone; a sentence still spelling one names a file that does
//	   not exist, and an agent running it gets 127. Only registered names
//	   count: `setup.sh` and a panel reproducer's `repro.sh` are not guards.
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
			for _, m := range gcShell.FindAllStringSubmatch(sc.Text(), -1) {
				if Registry[m[1]] != nil {
					fmt.Fprintf(&out, "%s:%d: %s.sh is the deleted shim spelling — write flow-guard %s (rule B)\n", rel, line, m[1], m[1])
					n++
				}
			}
		}
	}
	fmt.Fprint(stdout, out.String())
	if n > 0 {
		fmt.Fprintf(stdout, "GUARD-CALLS-INVALID: %s — %d violation(s)\n", root, n)
		return 1
	}
	fmt.Fprintf(stdout, "GUARD-CALLS-OK: %s — %d file(s) scanned\n", root, len(files))
	return 0
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
