package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// checkTreeMarkers is scripts/check-tree-markers.sh: that script's header
// comment is the contract -- snapshot writes one <path><TAB><marker><TAB>
// <status> line per spec line, verify recomputes them and compares
// byte-for-byte; exit 0 snapshot written or every marker matches, 1 verify
// found the markers changed (each difference named on stderr), 2 cannot
// answer. The reasoning for each step, moved here from the bash body it
// replaced (c9b8fde8), sits beside the code it explains.
func init() {
	Registry["check-tree-markers"] = checkTreeMarkers
}

func checkTreeMarkers(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	mode, worktree, markers, snap := arg(0), arg(1), arg(2), arg(3)
	fail := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "check-tree-markers: "+format+"\n", a...)
		return 2
	}
	usage := func() int {
		return fail("usage: check-tree-markers.sh snapshot|verify <worktree> <markers-file> <snapshot-file>")
	}
	if mode != "snapshot" && mode != "verify" {
		return usage()
	}
	if worktree == "" || !isDir(smcAbs(env, worktree)) {
		return fail("not a directory: %s", orMissing(worktree))
	}
	var specBytes []byte
	readErr := os.ErrNotExist
	if markers != "" && isFile(smcAbs(env, markers)) {
		specBytes, readErr = os.ReadFile(smcAbs(env, markers))
	}
	if readErr != nil {
		return fail("no readable markers file: %s", orMissing(markers))
	}
	if snap == "" {
		return usage()
	}
	// `cd "$WORKTREE" && pwd -P`: the physical path every grep reads under,
	// and the one a grep-failure message names.
	physical, err := filepath.EvalSymlinks(smcAbs(env, worktree))
	if err != nil {
		return fail("worktree vanished: %s", worktree)
	}

	// Validate every spec line BEFORE anything is written: the snapshot
	// file must never come into existence half-made. `read -r` reads a last
	// line with no newline too, and an empty line is skipped.
	type spec struct{ path, marker string }
	var specs []spec
	lines := strings.Split(string(specBytes), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		if line == "" {
			continue
		}
		path, marker, ok := strings.Cut(line, "\t")
		if !ok {
			return fail("line without a tab: %s", line)
		}
		// CONTAINMENT: the markers file names what grep reads. An absolute
		// path or any `..` component would read outside the worktree, so the
		// spec is refused rather than sanitised -- the parent fixes its
		// marker list, the guard never widens it.
		if path == "" {
			return fail("empty path in markers file")
		}
		if strings.HasPrefix(path, "/") {
			return fail("absolute path '%s' in markers file -- refused", path)
		}
		for _, c := range strings.Split(path, "/") {
			if c == ".." {
				return fail("path '%s' carries a '..' component -- refused", path)
			}
		}
		if marker == "" {
			return fail("empty marker for '%s' in markers file", path)
		}
		specs = append(specs, spec{path, marker})
	}
	// F2 (panel round 1): a markers file with zero valid lines would
	// snapshot and verify clean over a destroyed tree -- two empty line-sets
	// compare equal. An empty pin set pins nothing, so it cannot answer.
	if len(specs) == 0 {
		return fail("the markers file names no artifact -- cannot answer")
	}

	// Both snapshot-path checks run before any marker is recomputed, as the
	// bash ran them ahead of record_all.
	snapAbs := smcAbs(env, snap)
	var recorded []byte
	if mode == "snapshot" {
		// F1 (panel round 1): mv files the snapshot INSIDE an existing
		// directory, so the promised path would hold nothing while exit 0
		// claimed a snapshot -- refused before anything is written.
		if isDir(snapAbs) {
			return fail("snapshot path is an existing directory: %s", snap)
		}
	} else {
		var err error
		recorded, err = os.ReadFile(snapAbs)
		if err != nil || !isFile(snapAbs) {
			return fail("no readable snapshot at %s -- cannot answer; snapshot runs before the dispatch, never after it", snap)
		}
	}

	// One snapshot line per spec line: <path><TAB><marker><TAB><status>,
	// where status is the literal `missing` or the grep -cE line count.
	// grep runs as the bash ran it, under LC_ALL=C over env's PATH, so the
	// ERE dialect is grep's own and a snapshot the bash wrote verifies here.
	// grep -c exits 1 on a zero count -- a valid answer; any other failure on
	// an existing file cannot be answered at all.
	grep := "grep"
	if p, ok := lookPath(env, "grep"); ok {
		grep = p
	}
	var current bytes.Buffer
	for _, s := range specs {
		file := physical + "/" + s.path
		if !isFile(file) {
			fmt.Fprintf(&current, "%s\t%s\tmissing\n", s.path, s.marker)
			continue
		}
		cmd := exec.Command(grep, "-cE", "--", s.marker, file)
		cmd.Env = append(os.Environ(), "LC_ALL=C")
		out, err := cmd.Output()
		var exit *exec.ExitError
		if err != nil && (!errors.As(err, &exit) || exit.ExitCode() > 1) {
			return fail("grep failed on %s -- cannot answer", file)
		}
		count := strings.TrimRight(string(out), "\n")
		if count == "" {
			count = "0"
		}
		fmt.Fprintf(&current, "%s\t%s\t%s\n", s.path, s.marker, count)
	}

	if mode == "snapshot" {
		// Written to a temp file beside the snapshot and renamed over it, as
		// the bash's mktemp + mv: never half-made, mode 0600.
		if !writeAtomic(snapAbs, current.Bytes()) {
			return fail("could not write the snapshot at %s", snap)
		}
		fmt.Fprintf(stdout, "TREE-MARKERS-SNAPSHOT-OK: markers pinned at %s\n", snap)
		return 0
	}

	if bytes.Equal(recorded, current.Bytes()) {
		fmt.Fprintln(stdout, "TREE-MARKERS-OK: every recorded marker matches its pre-dispatch snapshot")
		return 0
	}
	fmt.Fprintln(stderr, "check-tree-markers: the tree's recorded markers changed across the dispatch -- the dispatch's own clean-or-restored claim is not evidence; stop and recover by hand from before the dispatch")
	// `diff "$SNAP" "$CURRENT" | sed 's/^/check-tree-markers: /'`: the
	// difference in diff's own format, every line prefixed.
	if tmp, err := os.CreateTemp(env.Getenv("TMPDIR"), "check-tree-markers-verify."); err == nil {
		_, werr := tmp.Write(current.Bytes())
		if cerr := tmp.Close(); werr == nil && cerr == nil {
			diffBin := "diff"
			if p, ok := lookPath(env, "diff"); ok {
				diffBin = p
			}
			cmd := exec.Command(diffBin, snap, tmp.Name())
			cmd.Dir = env.Dir
			cmd.Env = append(os.Environ(), "LC_ALL=C") // the bash exported LC_ALL=C for its whole run
			out, _ := cmd.Output()
			for _, l := range strings.SplitAfter(string(out), "\n") {
				if l != "" {
					fmt.Fprint(stderr, "check-tree-markers: "+l)
				}
			}
		}
		_ = os.Remove(tmp.Name())
	}
	return 1
}

func orMissing(s string) string {
	if s == "" {
		return "<missing>"
	}
	return s
}

// writeAtomic writes body to a fresh 0600 temp file in path's directory and
// renames it over path.
func writeAtomic(path string, body []byte) bool {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".check-tree-markers-snap.")
	if err != nil {
		return false
	}
	_, werr := tmp.Write(body)
	cerr := tmp.Close()
	if werr != nil || cerr != nil || os.Rename(tmp.Name(), path) != nil {
		_ = os.Remove(tmp.Name())
		return false
	}
	return true
}
