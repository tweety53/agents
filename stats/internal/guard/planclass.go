package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// planClass is scripts/plan-class.sh: that script's header comment is the
// contract -- the three inputs/class/rolls lines on stdout, exit 0 on a
// printed answer, exit 2 on a refusal -- and states every threshold below.
// The reasoning for each step, moved here from the bash body it replaced
// (c5379c0a), sits beside the code it explains.
func init() {
	Registry["plan-class"] = planClass
}

// pcMicroLineCap is MICRO_LINE_CAP: the micro class's changed-line cap on
// the change's own touched surface.
const pcMicroLineCap = 20

var (
	pcTaskLine  = regexp.MustCompile(`^- \[[ xX]\] [0-9]+\.`)
	pcBacktick  = regexp.MustCompile("`[^`]+`")
	pcMigration = regexp.MustCompile(`^stats/internal/store/migrations/|\.sql$`)
)

// pcRed is grep -E '^\*\*Build:\*\* red\b' in the caller's locale: under
// UTF-8 a non-ASCII letter or digit right after "red" continues the word, as
// grep's locale-aware \b has it, and so does a byte that is not UTF-8 (grep
// at c5379c0a, measured); otherwise only an ASCII one does.
func pcRed(env Env, line string) bool {
	rest, ok := strings.CutPrefix(line, "**Build:** red")
	if !ok || rest == "" {
		return ok
	}
	r, size := utf8.DecodeRuneInString(rest)
	if r >= utf8.RuneSelf {
		if !smcUTF8(env) {
			return true
		}
		if r == utf8.RuneError && size == 1 {
			return false
		}
	}
	return !(r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r))
}

// isDocPath is `\.mdc?$`: a documentation path.
func isDocPath(p string) bool {
	return strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".mdc")
}

func planClass(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 && len(args) != 4 {
		fmt.Fprint(stderr, "usage: plan-class.sh <tasks.md> <repos> [<worktree> <merge-base>]\n")
		return 2
	}
	tasksFile, repos := args[0], args[1]
	if !isFile(smcAbs(env, tasksFile)) {
		fmt.Fprintf(stderr, "plan-class.sh: no such file: %s\n", tasksFile)
		return 2
	}
	if repos == "" || strings.Trim(repos, "0123456789") != "" {
		fmt.Fprintf(stderr, "plan-class.sh: <repos> must be a non-negative integer, got: %s\n", repos)
		return 2
	}
	b, err := os.ReadFile(smcAbs(env, tasksFile))
	if err != nil {
		fmt.Fprintf(stderr, "plan-class.sh: cannot read %s: %v\n", tasksFile, err)
		return 2
	}

	tasks, red := 0, false
	seen := map[string]bool{}
	// Every backticked path on a **Files:** line, deduplicated across the
	// plan. The bash's `sort -u` ordered them too; only their count and
	// kinds are printed, so no order is kept.
	var files []string
	for _, l := range lines(b) {
		if pcTaskLine.MatchString(l) {
			tasks++
		}
		if pcRed(env, l) {
			red = true
		}
		if !strings.HasPrefix(l, "**Files:**") {
			continue
		}
		for _, m := range pcBacktick.FindAllString(l, -1) {
			if f := strings.Trim(m, "`"); !seen[f] {
				seen[f] = true
				files = append(files, f)
			}
		}
	}
	migration, spec, allDocs := false, false, len(files) > 0
	for _, f := range files {
		migration = migration || pcMigration.MatchString(f)
		spec = spec || strings.HasPrefix(f, "spectre/specs/")
		allDocs = allDocs && isDocPath(f)
	}
	unverified := strings.Contains(string(b), "unverified:")

	// `[ "$REPOS" -eq 1 ]` / `-gt 1`: a value past int64 is no integer to
	// bash's test, so it is neither — the error test prints is not.
	n, err := strconv.ParseInt(repos, 10, 64)
	if err != nil {
		n = -1
	}
	class := "regular"
	if tasks <= 8 && len(files) <= 19 && n == 1 && !migration && !spec {
		class = "small"
	} else if tasks >= 22 || len(files) >= 60 || (n > 1 && tasks >= 11) || (migration && tasks >= 11) {
		class = "big"
	}

	// The change's own diff side: with the optional arguments it is
	// answered on every plan, micro-eligible or not — an unanswerable
	// <worktree>/<merge-base> is exit 2 unconditionally, exactly as the
	// header states. One numstat of the merge base against the working tree
	// covers committed, staged and unstaged together, so churn on one file
	// counts once. A numstat entry git cannot count (binary) is
	// fail-closed: never micro.
	if len(args) == 4 {
		worktree, mergebase := args[2], args[3]
		gitBin, ok := panelResolveGit(env, "plan-class", stderr)
		if !ok || !panelValidateWorktree(env, "plan-class", worktree, mergebase, gitBin, stderr) {
			return 2
		}
		surface, ok := panelTouchedPaths(env, "plan-class", worktree, mergebase, gitBin, stderr)
		if !ok {
			return 2
		}
		cmd := panelGit(env, gitBin, worktree, "diff", "--no-renames", "--numstat", "--end-of-options", mergebase)
		cmd.Stderr = stderr
		numstat, err := cmd.Output()
		if err != nil {
			return 2
		}
		changed := 0
		for _, l := range lines(numstat) {
			f := strings.Fields(l)
			if len(f) < 2 {
				continue
			}
			if f[0] == "-" || f[1] == "-" {
				changed = pcMicroLineCap + 1
				break
			}
			a, _ := strconv.Atoi(f[0])
			d, _ := strconv.Atoi(f[1])
			changed += a + d
		}
		// The micro class: the plan side is decidable from the tasks.md
		// alone; the diff side must have been answered too. An empty
		// touched surface (a creating run, nothing implemented yet) passes
		// vacuously; a non-empty one must itself be docs-only and within
		// the cap.
		surfaceDocs := true
		for _, p := range surface {
			surfaceDocs = surfaceDocs && isDocPath(p)
		}
		if class == "small" && tasks <= 2 && !red && allDocs && surfaceDocs && changed <= pcMicroLineCap {
			class = "micro"
		}
	}

	// Rolls are reproducible per change name (the directory holding
	// <tasks.md>): each the first 8 hex digits of a SHA-256, mod 100.
	name := pabBasename(gdcDirname(tasksFile))
	roll := func(s string) uint64 {
		v, _ := strconv.ParseUint(sha256Hex(s)[:8], 16, 64)
		return v % 100
	}
	yn := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}
	fmt.Fprintf(stdout, "inputs: tasks=%d files=%d repos=%s migration=%s spec=%s red=%s unverified=%s\n",
		tasks, len(files), repos, yn(migration), yn(spec), yn(red), yn(unverified))
	fmt.Fprintf(stdout, "class: %s\n", class)
	fmt.Fprintf(stdout, "rolls: compact %d · experimental %d · bundle %d\n", roll(name), roll(name+"exp"), roll(name+"bundle"))
	return 0
}
