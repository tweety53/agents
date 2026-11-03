package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-python-suppressions.sh shim (KAN-1058).
//
// flow-guard check-python-suppressions — reject linter suppression markers in Python.
//
// Usage: flow-guard check-python-suppressions [path ...]     # dirs or files
//
// With no arguments it scans every tracked `*.py` file in the repository —
// `git ls-files '*.py'` from the repo root resolved from this script's own
// location, paths made absolute before the scan so the answer never depends
// on the caller's working directory. With arguments it scans the given files
// or directories instead: the seam its Go test
// (stats/internal/guard/check_python_suppressions_test.go) uses to point the
// guard at fixture trees without touching the real one. The mechanism is
// stats/internal/guard/pythonsuppressions.go. Untracked or ignored Python is never
// scanned — the guard answers for what the repository owns.
//
// WHAT IT CATCHES: the inline markers whose job is to silence a linter or
// type checker on the line (or file) they sit on — `# noqa` (flake8/ruff,
// bare, coded, or file-level `# flake8: noqa`), `# type: ignore` (mypy and
// pyright), `# pyright: ignore` (pyright's own spelling), `# pylint:
// disable`, `# mypy: ignore-errors`, and `# nolint` (the golangci/eslint
// spelling). Matched case-insensitively, anywhere on the line, docstrings
// included — a marker in prose is still an instruction a future reader can
// copy. The Lint Fix Priority rule forbids every one of these outright; the
// fix for the finding a marker hides is the code, never the marker.
//
// WHAT A GREEN RUN DOES NOT PROVE — read before trusting it. This is a
// regression guard for a fixed list of markers. It does not
// prove the Python is lint-clean: no linter runs over this repository's
// Python, and a suppression form outside the list (`# ruff:
// per-file-ignores` in a config file, a decorator no entry names) passes
// clean, as does every non-Python file. It proves one thing: the markers
// listed above have not come back.
//
// Exit codes: 0 no marker found (or no Python file to scan); 1 at least one
// marker found; 2 the guard cannot answer at all — a path argument that is
// neither file nor directory, a git or find failure while enumerating the
// scan set, or a grep that failed for any reason other than "no match". An
// enumeration failure is never read as an empty scan set: a guard that
// cannot answer must say so loudly, because a vacuous pass is the failure
// mode this exit code exists for.

// checkPythonSuppressions is `flow-guard check-python-suppressions`. The comment above states the contract — the markers, what a green run does
// not prove, and the exit codes; this file is the mechanism.
func init() {
	Registry["check-python-suppressions"] = checkPythonSuppressions
}

// psMarker is one alternation covering every marker the header names, case-
// insensitive; `noqa` alone also covers `# flake8: noqa`, `# ruff: noqa`
// and every coded `# noqa: <codes>` spelling.
var psMarker = regexp.MustCompile(`(?i)noqa|nolint|type:[[:space:]]*ignore|pyright:[[:space:]]*ignore|pylint:[[:space:]]*disable|mypy:[[:space:]]*ignore-errors`)

func checkPythonSuppressions(args []string, env Env, stdout, stderr io.Writer) int {
	files, ok := psScanSet(args, env, stderr)
	if !ok {
		return 2
	}
	// No Python in scope is a clean answer, not an error.
	if len(files) == 0 {
		fmt.Fprintln(stdout, "suppressions-ok: no python file in scope")
		return 0
	}

	// Every hit line carries its filename even when the scan set holds a
	// single file — a hit a reader cannot locate is half a report (the
	// bash's grep -H).
	var hits []string
	for _, f := range files {
		b, err := os.ReadFile(smcAbs(env, f))
		if err != nil {
			// grep's trouble exit, reported as the bash reported it.
			fmt.Fprintf(stderr, "grep: %s: %s\nERROR: grep failed with exit 2 — cannot answer\n", f, psStrerror(err))
			return 2
		}
		// A file holding a NUL byte is binary and yields no hit — how grep
		// -I classifies one (python-suppressions-binary-skip).
		if bytes.IndexByte(b, 0) >= 0 {
			continue
		}
		for i, line := range fileLines(string(b)) {
			if psMarker.MatchString(line) {
				hits = append(hits, fmt.Sprintf("%s:%d:%s", f, i+1, line))
			}
		}
	}
	if len(hits) == 0 {
		fmt.Fprintln(stdout, "suppressions-ok: no linter suppression marker in the python in scope")
		return 0
	}
	fmt.Fprintln(stdout, strings.Join(hits, "\n"))
	fmt.Fprintf(stderr, "suppressions-found: %d marker hit(s) — fix the code, never widen this guard\n", len(hits))
	return 1
}

// psScanSet is the files to scan, as printed. An enumeration failure is
// never read as an empty scan set: a guard that cannot answer must say so
// loudly (false, after its message), because a vacuous pass is the failure
// mode exit 2 exists for. Symlinked entries are skipped — skills/*/scripts/
// carry symlinks to scripts/'s own copies, and following them would report
// the same target once per link; the target is always in the set from its
// real path.
func psScanSet(args []string, env Env, stderr io.Writer) ([]string, bool) {
	isLink := func(p string) bool {
		fi, err := os.Lstat(smcAbs(env, p))
		return err == nil && fi.Mode()&fs.ModeSymlink != 0
	}
	var files []string
	if len(args) == 0 {
		// Every tracked *.py from the repository root, paths made absolute
		// before the scan so the answer never depends on the caller's
		// working directory. The root is FLOW_GUARD_REPO_ROOT, which scripts/flow-guard exports.
		root := env.Getenv("FLOW_GUARD_REPO_ROOT")
		if root == "" {
			fmt.Fprintln(stderr, "check-python-suppressions: FLOW_GUARD_REPO_ROOT is unset — run flow-guard check-python-suppressions, which sets it")
			return nil, false
		}
		out, rc := gitExec(root, nil, stderr, "ls-files", "-z", "--", "*.py")
		if rc != 0 {
			fmt.Fprintln(stderr, "ERROR: git ls-files failed — cannot enumerate the tracked Python")
			return nil, false
		}
		for _, f := range strings.Split(out, "\x00") {
			if f != "" && !isLink(root+"/"+f) {
				files = append(files, root+"/"+f)
			}
		}
		return files, true
	}
	for _, arg := range args {
		info, err := os.Stat(smcAbs(env, arg))
		switch {
		case err == nil && info.Mode().IsRegular():
			if !isLink(arg) && strings.HasSuffix(arg, ".py") {
				files = append(files, arg)
			}
		case err == nil && info.IsDir():
			found, ok := psFind(env, arg)
			if !ok {
				fmt.Fprintf(stderr, "ERROR: find failed on %s — cannot enumerate the scan set\n", arg)
				return nil, false
			}
			files = append(files, found...)
		default:
			fmt.Fprintf(stderr, "ERROR: no such file or directory: %s\n", arg)
			return nil, false
		}
	}
	return files, true
}

// psFind is `find <dir> -type f -name '*.py'`: a physical walk in directory
// order (unsorted, as find's fts reads it), paths spelled as find spells
// them — the argument plus "/" unless it already ends in one. Like find, it
// carries on past a directory it cannot read and fails at the end, and
// never follows a symlink — not even a symlinked starting point.
func psFind(env Env, dir string) ([]string, bool) {
	if fi, err := os.Lstat(smcAbs(env, dir)); err == nil && fi.Mode()&fs.ModeSymlink != 0 {
		return nil, true
	}
	var files []string
	ok := true
	var walk func(p string)
	walk = func(p string) {
		f, err := os.Open(smcAbs(env, p))
		if err != nil {
			ok = false
			return
		}
		ents, err := f.ReadDir(-1)
		_ = f.Close()
		if err != nil {
			ok = false
		}
		sep := "/"
		if strings.HasSuffix(p, "/") {
			sep = ""
		}
		for _, ent := range ents {
			child := p + sep + ent.Name()
			switch {
			case ent.Type().IsRegular() && strings.HasSuffix(ent.Name(), ".py"):
				files = append(files, child)
			case ent.IsDir():
				walk(child)
			}
		}
	}
	walk(dir)
	return files, ok
}

// psStrerror is err's errno text as grep prints it, strerror-capitalised:
// "Permission denied", not Go's "open f: permission denied".
func psStrerror(err error) string {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		err = pe.Err
	}
	msg := err.Error()
	if msg == "" {
		return msg
	}
	return strings.ToUpper(msg[:1]) + msg[1:]
}
