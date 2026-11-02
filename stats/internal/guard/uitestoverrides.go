package guard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-uitest-overrides.sh shim (KAN-1058).
//
// flow-guard check-uitest-overrides — fail when a UITEST_* Makefile variable is
// assigned without `override`.
//
// Why this exists: none of stats/Makefile's ui-test-up / ui-test-down targets
// take parameters, and several of the UITEST_* variables they use are read
// straight into a `dropdb --force`, an unqualified `rm -rf`, or a `kill` on a
// pid read from a file. A plain assignment is silently overridden by a
// command-line or environment value (`make ui-test-up UITEST_LOG=/etc/passwd`,
// or an exported UITEST_LOG under `make -e`), which lets a caller aim any of
// those at an arbitrary path or process. Fixing that has recurred three times
// already, one variable at a time (UITEST_DB, then UITEST_ID, then the rest) —
// this guard checks the whole class at once so a variable added later cannot
// reintroduce the same defect unnoticed.
//
// Rule: every line assigning a UITEST_* variable in the scanned file(s) must
// begin with `override`.
//
// WHAT "ASSIGNS" MEANS. Earlier versions of this guard matched only a plain
// `:=` anchored at column 0, which passed every one of these while GNU Make
// still treats each as an ordinary, command-line-overridable assignment:
//   - a leading SPACE before the variable name (`  UITEST_EVIL := ...`) — a
//     leading TAB is different: that is a recipe line, evaluated by the
//     shell, not by Make, so it is deliberately NOT treated as an assignment
//     here;
//   - `?=` (conditional), `+=` (append), `=` (recursive), `::=` (POSIX
//     simple).
// The pattern below matches Make's assignment grammar instead: optional
// leading spaces, `UITEST_<NAME>`, optional whitespace, then any of those
// five operators. A line is flagged unless the UITEST_* name is immediately
// preceded on the line by `override ` — i.e. `override UITEST_X := ...`
// never matches the violation pattern, because the line does not begin
// (after leading spaces) with `UITEST_`.
//
// THREE MORE FORMS, added after a live check found each beat the guard
// above while still winning a command-line override in GNU Make:
//
//   - `export UITEST_X := ...` / `unexport UITEST_X` — the plain-assignment
//     pattern above requires the line to *begin* (after leading spaces)
//     with `UITEST_`, so `export ` in front hides it from that pattern,
//     while GNU Make still lets a command-line assignment win. The accepted
//     protected form is `override export UITEST_X := ...` (override
//     first) — confirmed against GNU Make 4.4.1 that `override` there still
//     beats a command-line reassignment; the macOS-shipped GNU Make 3.81
//     mishandles `override`+`export` combined regardless of order, which is
//     a bug in that decades-old release, not a reason to trust the
//     combined form less — verified separately, not fixed here (out of
//     this guard's scope, and not a UITEST_* codepath).
//   - `$(eval UITEST_X := ...)` — text (`$(eval `) precedes `UITEST_` on
//     the line, same evasion as `export` above. The accepted protected
//     form is `$(eval override UITEST_X := ...)`.
//   - `define UITEST_X` / `endef` — a `define`d variable is an ordinary
//     recursive variable to Make and is beatable from the command line
//     exactly like `:=`, confirmed live against stats/Makefile's
//     UITEST_SAFE_RMRF macro (`make -n ui-test-down
//     'UITEST_SAFE_RMRF=@echo INJECTED $(1)'` replaced the entire safety
//     `case` block with the injected text). The accepted protected form is
//     `override define UITEST_X` / `endef`. Lines inside a `define`...
//     `endef` body are macro text, not Make assignments, and are not
//     scanned as if they were — only the `define` line itself is checked
//     for `override`.
//
// INCLUDED FILES. `include` / `-include` / `sinclude` directives naming a
// literal path (no `$(...)` expansion) are followed recursively, resolved
// relative to the directory of the file that names them, so a UITEST_*
// variable moved into an included file is still caught. A target that
// contains `$` — a Make variable or function reference this guard cannot
// evaluate without running Make itself — is reported as unresolved and
// skipped; an included file that does not exist on disk is likewise reported
// and skipped. LIMITATION, stated here rather than left silent: wildcarded
// includes (`include $(wildcard *.mk)`) and any include path built from a
// Make variable are not followed, so a UITEST_* variable hidden behind one of
// those would not be caught. stats/Makefile has no `include` directive today.
//
// Usage: flow-guard check-uitest-overrides [<makefile>]
// With no argument, checks stats/Makefile in this repository — the one real
// caller (`make test`, `make check-uitest-overrides`). An explicit argument
// exists so the guard's own test harness can point it at a throwaway
// fixture instead of editing the real Makefile into a broken state. The
// mechanism is stats/internal/guard/uitestoverrides.go; its test is
// stats/internal/guard/check_uitest_overrides_test.go, and the shim exports
// this repository's root as FLOW_GUARD_REPO_ROOT for the default path.

// checkUitestOverrides is `flow-guard check-uitest-overrides`. The comment above states the contract — what counts as an assignment, the
// define/export/eval forms, include resolution and its stated limit; this
// file is the mechanism.
func init() {
	Registry["check-uitest-overrides"] = checkUitestOverrides
}

var uoSpace = "[" + cSpace + "]"

var (
	uoEndef   = regexp.MustCompile(`^[ ]*endef(` + uoSpace + `|$)`)
	uoDefine  = regexp.MustCompile(`^[ ]*(override` + uoSpace + `+)?define` + uoSpace + `+(UITEST_[A-Za-z0-9_]*)`)
	uoInclude = regexp.MustCompile(`^[ ]*(-include|sinclude|include)` + uoSpace + `+(.+)$`)
	uoExport  = regexp.MustCompile(`^[ ]*(override` + uoSpace + `+)?(export|unexport)` + uoSpace + `+(UITEST_[A-Za-z0-9_]*)` + uoSpace + `*(::=|:=|\+=|\?=|=)`)
	uoAssign  = regexp.MustCompile(`^[ ]*(UITEST_[A-Za-z0-9_]*)` + uoSpace + `*(::=|:=|\+=|\?=|=)`)
	uoEval    = regexp.MustCompile(`\$\(eval` + uoSpace + `+(override` + uoSpace + `+)?(UITEST_[A-Za-z0-9_]*)` + uoSpace + `*(::=|:=|\+=|\?=|=)`)
)

const uoReassign = " — a caller can reassign it on the command line or via the environment"

func checkUitestOverrides(args []string, env Env, stdout, stderr io.Writer) int {
	var makefile string
	if len(args) > 0 && args[0] != "" {
		makefile = args[0]
	} else {
		// The repository root is FLOW_GUARD_REPO_ROOT, which scripts/flow-guard exports.
		root := env.Getenv("FLOW_GUARD_REPO_ROOT")
		if root == "" {
			fmt.Fprintln(stderr, "check-uitest-overrides: FLOW_GUARD_REPO_ROOT is unset — run flow-guard check-uitest-overrides, which sets it")
			return 2
		}
		makefile = root + "/stats/Makefile"
	}
	if !isFile(smcAbs(env, makefile)) {
		fmt.Fprintf(stderr, "check-uitest-overrides: %s not found\n", makefile)
		return 2
	}

	violations, unread := 0, false
	// scanned holds every file already scanned, so an include cycle ends and
	// a file reachable twice is reported once.
	scanned := map[string]bool{makefile: true}
	var scan func(file string)
	scan = func(file string) {
		baseDir := filepath.Clean(pcAbs(env, filepath.Dir(file)))
		// A file that cannot be read fails the run: the bash's `done <
		// "$file"` did, and a guard reporting clean on a file it never read
		// is fail-open.
		f, err := os.Open(smcAbs(env, file))
		if err != nil {
			fmt.Fprintf(stderr, "check-uitest-overrides: %s: %v\n", file, err)
			unread = true
			return
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		sc.Buffer(nil, 1<<24)
		inDefine := false
		for n := 1; sc.Scan(); n++ {
			line := sc.Text()
			if inDefine {
				inDefine = !uoEndef.MatchString(line)
				continue
			}
			if m := uoDefine.FindStringSubmatch(line); m != nil {
				inDefine = true
				if m[1] == "" {
					fmt.Fprintf(stderr, "check-uitest-overrides: %s:%d: %s is defined without `override`%s\n", file, n, m[2], uoReassign)
					violations++
				}
				continue
			}
			if m := uoInclude.FindStringSubmatch(line); m != nil {
				for _, t := range strings.Fields(m[2]) {
					if strings.Contains(t, "$") {
						fmt.Fprintf(stderr, "check-uitest-overrides: %s:%d: include target '%s' uses Make variable/function expansion this guard cannot resolve — not scanned\n", file, n, t)
						continue
					}
					resolved := t
					if !strings.HasPrefix(t, "/") {
						resolved = baseDir + "/" + t
					}
					if !isFile(resolved) {
						fmt.Fprintf(stderr, "check-uitest-overrides: %s:%d: included file not found, not scanned: %s\n", file, n, resolved)
						continue
					}
					resolved = filepath.Clean(filepath.Dir(resolved)) + "/" + filepath.Base(resolved)
					if !scanned[resolved] {
						scanned[resolved] = true
						scan(resolved)
					}
				}
				continue
			}
			if m := uoExport.FindStringSubmatch(line); m != nil {
				if m[1] == "" {
					fmt.Fprintf(stderr, "check-uitest-overrides: %s:%d: %s is assigned via `%s` without `override`%s\n", file, n, m[3], m[2], uoReassign)
					violations++
				}
				continue
			}
			if m := uoAssign.FindStringSubmatch(line); m != nil {
				fmt.Fprintf(stderr, "check-uitest-overrides: %s:%d: %s is assigned without `override`%s\n", file, n, m[1], uoReassign)
				violations++
			}
			if m := uoEval.FindStringSubmatch(line); m != nil && m[1] == "" {
				fmt.Fprintf(stderr, "check-uitest-overrides: %s:%d: %s is assigned inside `$(eval ...)` without `override`%s\n", file, n, m[2], uoReassign)
				violations++
			}
		}
		if err := sc.Err(); err != nil {
			fmt.Fprintf(stderr, "check-uitest-overrides: %s: %v\n", file, err)
			unread = true
		}
	}
	scan(makefile)

	if unread {
		return 2
	}
	if violations != 0 {
		fmt.Fprintf(stdout, "UITEST-OVERRIDES-FAIL: %d UITEST_* variable(s) assigned without override\n", violations)
		return 1
	}
	fmt.Fprintln(stdout, "UITEST-OVERRIDES-OK: every UITEST_* variable is override")
	return 0
}
