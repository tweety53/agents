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

// The contract below moved here from the header of the deleted
// scripts/check-tree-markers.sh shim (KAN-1058).
//
// flow-guard check-tree-markers snapshot|verify <worktree> <markers-file> <snapshot-file>
//
// The guard KAN-643 asked for: a subagent's self-reported clean state is
// never evidence for anything — every dispatch that can touch the worktree
// is bracketed by CONTENT MARKERS recorded before the dispatch and
// re-grepped after it, and a mismatch is read as the subagent having
// mutated the tree. On gymie KAN-579 a reviewer destroyed never-committed
// planning artifacts and self-reported the tree restored; nothing
// independent checked the claim until the deferred self-review grepped the
// tree by hand against strings the artifacts were known to contain. This
// guard is that grep, recorded.
//
//   snapshot  run immediately BEFORE the dispatch's Agent call — greps each
//             marker and writes one <path><TAB><marker><TAB><status> line
//             per spec line to <snapshot-file>
//   verify    run when the dispatch's report file exists, BEFORE any report
//             or verdict is read or acted on — recomputes the same lines
//             and compares byte-for-byte
//
// snapshot also writes `git -C <worktree> status --porcelain -uall` after a
// `#porcelain` line, paths under `.superpowers/` dropped; verify fails (1)
// naming every porcelain line present now and absent from the snapshot — a
// path a slot left changed or untracked, a reproducer's scratch copy
// included (kan-1017). A snapshot with no `#porcelain` section verifies on
// its markers alone.
//
// <markers-file> carries one <path><TAB><ERE> line per artifact the
// dispatch must not damage, each ERE a string the artifact is known to
// contain — a decision ID, a task title, a section heading. Paths are
// worktree-relative and are refused before grep ever runs when they are
// absolute or carry a `..` component: the markers file names what grep
// reads, so it is a containment boundary exactly like the change-name case
// block in flow-guard check-plan-unchanged, and it is validated the same way —
// refused, not sanitised.
//
// MARKERS SIT BESIDE flow-guard check-plan-unchanged, never in its place: that
// guard asserts git's view of the plan tree only, while markers pin known
// content anywhere in the tree, independent of git entirely — content that
// was never committed (the KAN-579 artifacts) has no git answer at all.
// Both guards' verify exit 1 stops the run the same way, and for the same
// reason: the dispatch's own clean-or-restored prose is not evidence, so
// recovery is by hand from before the dispatch, never by re-running it.
//
// Exit codes:
//   0  snapshot written, or every recorded marker matches its snapshot
//   1  verify found the tree's markers changed since the snapshot, or a
//      path changed or untracked that the snapshot did not hold — each
//      difference named on stderr
//   2  cannot answer at all — missing arguments, a bad mode, a non-directory
//      worktree, an unreadable markers file, a malformed spec line (no tab,
//      an empty path or marker), a path that is absolute or carries a `..`
//      component, a markers file naming no artifact at all, the snapshot
//      path being an existing directory, a grep failure on an existing
//      file, an unwritable snapshot path (on snapshot), or a missing or
//      unreadable snapshot file (on verify)

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
		return fail("usage: flow-guard check-tree-markers snapshot|verify <worktree> <markers-file> <snapshot-file>")
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

	// sr6 (kan-1017): the tree's own porcelain, after a `#porcelain` line,
	// so verify names any path a slot left changed or untracked. Paths
	// under .superpowers/ are the run's own notes and are dropped.
	porcelain := tmPorcelain(physical)
	if mode == "snapshot" {
		current.WriteString(tmPorcelainMark)
		current.WriteString(strings.Join(porcelain, ""))
		// Written to a temp file beside the snapshot and renamed over it, as
		// the bash's mktemp + mv: never half-made, mode 0600.
		if !writeAtomic(snapAbs, current.Bytes()) {
			return fail("could not write the snapshot at %s", snap)
		}
		fmt.Fprintf(stdout, "TREE-MARKERS-SNAPSHOT-OK: markers pinned at %s\n", snap)
		return 0
	}

	// A snapshot written before sr6 has no porcelain section and verifies
	// on its markers alone.
	recorded, recordedPorcelain, hasPorcelain := bytes.Cut(recorded, []byte(tmPorcelainMark))
	var strays []string
	if hasPorcelain {
		before := map[string]bool{}
		for _, l := range strings.SplitAfter(string(recordedPorcelain), "\n") {
			before[l] = true
		}
		for _, l := range porcelain {
			if !before[l] {
				strays = append(strays, strings.TrimRight(l, "\n"))
			}
		}
	}
	if bytes.Equal(recorded, current.Bytes()) {
		if len(strays) == 0 {
			fmt.Fprintln(stdout, "TREE-MARKERS-OK: every recorded marker matches its pre-dispatch snapshot")
			return 0
		}
		fmt.Fprintln(stderr, "check-tree-markers: the dispatch left paths changed or untracked that the snapshot did not hold -- the dispatch's own clean-or-restored claim is not evidence; remove them or recover by hand from before the dispatch:")
		for _, l := range strays {
			fmt.Fprintln(stderr, "check-tree-markers:   "+l)
		}
		return 1
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
			snapArg := snap
			if hasPorcelain {
				if rec, err := os.CreateTemp(env.Getenv("TMPDIR"), "check-tree-markers-recorded."); err == nil {
					_, _ = rec.Write(recorded)
					_ = rec.Close()
					defer os.Remove(rec.Name())
					snapArg = rec.Name()
				}
			}
			cmd := exec.Command(diffBin, snapArg, tmp.Name())
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

const tmPorcelainMark = "#porcelain\n"

// tmPorcelain is `git -C <worktree> status --porcelain --untracked-files=all`, one
// newline-terminated line per path, .superpowers/ dropped. A worktree git
// cannot read yields no lines.
func tmPorcelain(worktree string) []string {
	out, err := exec.Command("git", "-C", worktree, "status", "--porcelain", "--untracked-files=all").Output()
	if err != nil {
		return nil
	}
	var lines []string
	for _, l := range strings.SplitAfter(string(out), "\n") {
		if len(l) > 3 && !strings.HasPrefix(l[3:], ".superpowers/") && !strings.HasPrefix(l[3:], "\".superpowers/") {
			lines = append(lines, l)
		}
	}
	return lines
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
