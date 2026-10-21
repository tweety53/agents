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

// checkPythonSuppressions is scripts/check-python-suppressions.sh's guard.
// The shim's header states the contract — the markers, what a green run does
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
		// working directory. The root is this shim's parent, exported as
		// FLOW_GUARD_REPO_ROOT.
		root := env.Getenv("FLOW_GUARD_REPO_ROOT")
		if root == "" {
			fmt.Fprintln(stderr, "check-python-suppressions: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-python-suppressions.sh, which sets it")
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
