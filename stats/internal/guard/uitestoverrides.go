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

// checkUitestOverrides is scripts/check-uitest-overrides.sh's guard. The
// shim's header states the contract — what counts as an assignment, the
// define/export/eval forms, include resolution and its stated limit; this
// file is the mechanism.
func init() {
	Registry["check-uitest-overrides"] = checkUitestOverrides
}

const uoSpace = `[ \t\n\v\f\r]`

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
		// The repository root is the shim's parent; the binary lives in a
		// cache, so the shim exports it.
		root := env.Getenv("FLOW_GUARD_REPO_ROOT")
		if root == "" {
			fmt.Fprintln(stderr, "check-uitest-overrides: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-uitest-overrides.sh, which sets it")
			return 2
		}
		makefile = root + "/stats/Makefile"
	}
	if !isFile(smcAbs(env, makefile)) {
		fmt.Fprintf(stderr, "check-uitest-overrides: %s not found\n", makefile)
		return 2
	}

	violations := 0
	// scanned holds every file already scanned, so an include cycle ends and
	// a file reachable twice is reported once.
	scanned := map[string]bool{makefile: true}
	var scan func(file string)
	scan = func(file string) {
		baseDir := filepath.Clean(pcAbs(env, filepath.Dir(file)))
		f, err := os.Open(smcAbs(env, file))
		if err != nil {
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
	}
	scan(makefile)

	if violations != 0 {
		fmt.Fprintf(stdout, "UITEST-OVERRIDES-FAIL: %d UITEST_* variable(s) assigned without override\n", violations)
		return 1
	}
	fmt.Fprintln(stdout, "UITEST-OVERRIDES-OK: every UITEST_* variable is override")
	return 0
}
