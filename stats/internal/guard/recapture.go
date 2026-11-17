package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// flow-guard recapture <capture-root> <source-worktree>=<base>[,<source-worktree>=<base>…] <command>
//
// Runs a fix round's recapture and leaves the receipt close-task asks for
// (skills/flow/verify-fix-loop.md step 2). The specs are flow-guard
// affected-specs' for each pair -- `affected-specs <source-worktree> <base>
// <capture-root>` -- united and sorted; <command> is the project's capture
// command, or its re-baseline command, with `<spec>` where the spec paths
// go. It runs once, through `sh -c` in <capture-root>, with `<spec>`
// replaced by the paths space-joined, its output passed through. With no
// affected spec nothing runs.
//
// THE RECEIPT is one JSON line appended to `flow-recapture.jsonl` in each
// source worktree's git directory: the specs the run took, its exit, and the
// blob id of every source file of that worktree's diff as it stood when the
// run began -- a deleted file's is empty. A source file is a path of the
// diff that is no capture output and no test (affected-specs' header). The
// receipt is what proves which sources a capture saw: a recapture that
// leaves every pixel where it was changes no tracked file.
//
// Exit codes:
//
//	0  the command exited 0, or there was nothing to capture.
//	1  the command exited non-zero; the receipt records the exit.
//	2  usage, a <command> with no `<spec>`, affected-specs' own exit 2, or
//	   a receipt that could not be written.
func init() { Registry["recapture"] = recapture }

// rcReceipt is one line of flow-recapture.jsonl.
type rcReceipt struct {
	Specs []string          `json:"specs"`
	Exit  int               `json:"exit"`
	Blobs map[string]string `json:"blobs"`
}

// rcFile is a worktree's receipt file, "" when git cannot name its directory.
func rcFile(env Env, worktree string) string {
	dir, ok := capture(envGit(env)("-C", worktree, "rev-parse", "--absolute-git-dir"))
	if !ok || dir == "" {
		return ""
	}
	return filepath.Join(dir, "flow-recapture.jsonl")
}

// rcSourcePaths keeps the paths of a diff that are sources: no capture
// output, and no test outside the capture root.
func rcSourcePaths(env Env, worktree, captureRoot string, paths []string) []string {
	wt, root := filepath.Clean(pcAbs(env, worktree)), filepath.Clean(pcAbs(env, captureRoot))
	var out []string
	for _, f := range paths {
		if f == "" || filepath.Base(f) == "full-app-suite.zip" || strings.Contains(filepath.Dir(f)+"/", "-snapshots/") {
			continue
		}
		if abs := filepath.Join(wt, f); asTestRE.MatchString(f) && abs != root && !strings.HasPrefix(abs, root+string(filepath.Separator)) {
			continue
		}
		out = append(out, f)
	}
	return out
}

func recapture(args []string, env Env, stdout, stderr io.Writer) int {
	usage := func(msg string) int {
		fmt.Fprintf(stderr, "recapture: %s\nusage: flow-guard recapture <capture-root> <source-worktree>=<base>[,<source-worktree>=<base>…] <command with <spec>>\n", msg)
		Reason(env, "usage")
		return 2
	}
	if len(args) != 3 {
		return usage(fmt.Sprintf("got %d argument(s)", len(args)))
	}
	root, command := args[0], args[2]
	if !strings.Contains(command, "<spec>") {
		return usage("the command names no <spec>")
	}
	git := envGit(env)
	var specs []string
	receipts := map[string]*rcReceipt{} // by source worktree
	var order []string
	for _, pair := range strings.Split(args[1], ",") {
		wt, base, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if !ok || wt == "" || base == "" {
			return usage("not <source-worktree>=<base>: " + pair)
		}
		var out bytes.Buffer
		if rc := Registry["affected-specs"]([]string{wt, base, root}, env, &out, stderr); rc != 0 {
			return rc
		}
		for _, s := range strings.Fields(out.String()) {
			if !slices.Contains(specs, s) {
				specs = append(specs, s)
			}
		}
		diff, _ := capture(git("-C", wt, "diff", "--name-only", "--no-renames", base))
		diff += "\n" + asUntracked(env, wt)
		r := &rcReceipt{Blobs: map[string]string{}}
		for _, f := range rcSourcePaths(env, wt, root, strings.Split(diff, "\n")) {
			r.Blobs[f], _ = capture(git("-C", wt, "hash-object", "--", f))
		}
		receipts[wt] = r
		order = append(order, wt)
	}
	sort.Strings(specs)

	rc := 0
	if len(specs) == 0 {
		fmt.Fprintln(stdout, "recapture: no affected spec — nothing to capture")
	} else {
		cmd := exec.Command("sh", "-c", strings.ReplaceAll(command, "<spec>", strings.Join(specs, " ")))
		cmd.Dir = pcAbs(env, root)
		cmd.Stdout, cmd.Stderr = stdout, stderr
		rc = exitCode(cmd.Run())
		fmt.Fprintf(stdout, "recapture: %d spec(s), exit %d\n", len(specs), rc)
	}
	for _, wt := range order {
		r := receipts[wt]
		r.Specs, r.Exit = specs, rc
		line, _ := json.Marshal(r)
		path := rcFile(env, wt)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			_, err = f.Write(append(line, '\n'))
			err = errorsJoinClose(err, f)
		}
		if path == "" || err != nil {
			fmt.Fprintf(stderr, "recapture: cannot write the receipt for %s\n", wt)
			Reason(env, "cannot-write-receipt")
			return 2
		}
	}
	if rc != 0 {
		Reason(env, "capture-failed")
		return 1
	}
	return 0
}

// errorsJoinClose closes f and returns err, else the close's error.
func errorsJoinClose(err error, f *os.File) error {
	if cerr := f.Close(); err == nil {
		return cerr
	}
	return err
}

// rcCovered reports whether a receipt of worktree's file records a passing
// run of every spec in specs on the blobs commit sha holds for its own
// source paths.
func rcCovered(env Env, worktree, captureRoot, sha string, specs []string) bool {
	raw, err := os.ReadFile(rcFile(env, worktree))
	if err != nil {
		return false
	}
	git := envGit(env)
	names, _ := capture(git("-C", worktree, "diff-tree", "--no-renames", "--no-commit-id", "--name-only", "-r", sha))
	want := map[string]string{}
	for _, f := range rcSourcePaths(env, worktree, captureRoot, strings.Split(names, "\n")) {
		want[f], _ = capture(git("-C", worktree, "rev-parse", "--verify", "--quiet", sha+":"+f))
	}
	for _, line := range bytes.Split(raw, []byte("\n")) {
		var r rcReceipt
		if json.Unmarshal(line, &r) != nil || r.Exit != 0 {
			continue
		}
		ok := !slices.ContainsFunc(specs, func(s string) bool { return !slices.Contains(r.Specs, s) })
		for f, blob := range want {
			got, seen := r.Blobs[f]
			ok = ok && seen && got == blob
		}
		if ok {
			return true
		}
	}
	return false
}
