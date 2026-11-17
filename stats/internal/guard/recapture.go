package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// flow-guard recapture <tasks.md>:<task-id> <capture-root> <source-worktree>=<base>[,<source-worktree>=<base>…] <command>
//
// Runs an in-run fix task's capture run and leaves the receipt close-task
// asks for (skills/flow/verify-fix-loop.md step 2). The task's tests are
// flow-guard affected-specs' task set for each pair -- `affected-specs -task
// <tasks.md>:<task-id> <source-worktree> <base> <capture-root>` -- united and
// sorted, each a spec or one test of it, `<spec>:<line>`; <command> is the
// project's capture command, or its re-baseline command, with `<spec>` where
// they go. It runs once, through `sh -c` in <capture-root>, with `<spec>`
// replaced by them space-joined, its output passed through.
//
// ONLY WHAT HAS NOT PASSED RUNS. A test an earlier receipt of every pair's
// worktree records as passed on the sources as they stand now -- the same
// source files of the diff, each at the same blob -- is left out, whether
// that run took it alone or inside its spec, so a red run is followed by a
// run of its failing tests alone. When one of the tests left to run failed
// in an earlier such receipt, ` --repeat-each 5 --workers 1` is appended to
// the command, as flow-guard run-verify-list -only appends it: a test that
// failed counts as passed only once it passes five times over at one worker.
// With nothing left to run nothing runs.
// ponytail: the flags go on the end of <command>, so a compound command must
// end in its Playwright call.
//
// THE RECEIPT is one JSON line appended to `flow-recapture.jsonl` in each
// source worktree's git directory: what the run took, its exit, the
// tests a red run failed, and the blob id of every source file of that
// worktree's diff as it stood when the run began -- a deleted file's is
// empty. A source file is a path of the diff that is no capture output and
// no test (affected-specs' header). The receipt is what proves which sources
// a capture saw: a recapture that leaves every pixel where it was changes no
// tracked file.
//
// THE FAILED TESTS of a red run are the `<spec>:<line>` entries its
// Playwright summary names (flow-guard run-verify-list's failed, interrupted
// and flaky blocks); every other test of the run passed. A red run whose
// summary names no test, names a file the run did not take, or counts a test
// that did not run, failed everything it took.
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
	Specs  []string          `json:"specs"`
	Exit   int               `json:"exit"`
	Failed []string          `json:"failed,omitempty"`
	Blobs  map[string]string `json:"blobs"`
}

// rcSpecFile is the spec file of a token: the token less its `:<line>`.
func rcSpecFile(token string) string {
	return asLinesRE.ReplaceAllString(token, "")
}

// rcLeft is what of token no receipt has passed yet, the receipts oldest
// first and all on the same sources: token itself until a run takes it --
// alone or inside its spec -- then the tests that run failed, each until a
// later run passes it. failed is set once a run failed any of it.
func rcLeft(token string, receipts []rcReceipt) (left []string, failed bool) {
	left = []string{token}
	for _, r := range receipts {
		var next []string
		for _, x := range left {
			switch {
			case !slices.Contains(r.Specs, x) && !slices.Contains(r.Specs, rcSpecFile(x)):
				next = append(next, x)
			case r.Exit == 0:
			case len(r.Failed) == 0:
				next, failed = append(next, x), true
			default:
				for _, f := range r.Failed {
					if f == x || f == rcSpecFile(x) {
						f = x
					} else if rcSpecFile(f) != x {
						continue
					}
					if failed = true; !slices.Contains(next, f) {
						next = append(next, f)
					}
				}
			}
		}
		left = next
	}
	return left, failed
}

// rcFile is a worktree's receipt file, "" when git cannot name its directory.
func rcFile(env Env, worktree string) string {
	dir, ok := capture(envGit(env)("-C", worktree, "rev-parse", "--absolute-git-dir"))
	if !ok || dir == "" {
		return ""
	}
	return filepath.Join(dir, "flow-recapture.jsonl")
}

// rcReceipts is a worktree's receipts, oldest first.
func rcReceipts(env Env, worktree string) []rcReceipt {
	raw, _ := os.ReadFile(rcFile(env, worktree))
	var out []rcReceipt
	for _, line := range bytes.Split(raw, []byte("\n")) {
		var r rcReceipt
		if json.Unmarshal(line, &r) == nil {
			out = append(out, r)
		}
	}
	return out
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

var rcDidNotRun = regexp.MustCompile(`(?m)^\s+\d+ did not run$`)

// rcFailed is the tests of a red run its log names as failing, nil when the
// log cannot tell them from the rest (THE FAILED TESTS).
func rcFailed(log []byte, specs []string) []string {
	if rcDidNotRun.Match(rvlANSI.ReplaceAll(log, nil)) {
		return nil
	}
	out := rvlVisualFailures(log)
	if slices.ContainsFunc(out, func(e string) bool {
		return !slices.ContainsFunc(specs, func(s string) bool { return rcSpecFile(s) == rcSpecFile(e) })
	}) {
		return nil
	}
	return out
}

func recapture(args []string, env Env, stdout, stderr io.Writer) int {
	usage := func(msg string) int {
		fmt.Fprintf(stderr, "recapture: %s\nusage: flow-guard recapture <tasks.md>:<task-id> <capture-root> <source-worktree>=<base>[,<source-worktree>=<base>…] <command with <spec>>\n", msg)
		Reason(env, "usage")
		return 2
	}
	if len(args) != 4 {
		return usage(fmt.Sprintf("got %d argument(s)", len(args)))
	}
	task, root, command := args[0], args[1], args[3]
	if !strings.Contains(command, "<spec>") {
		return usage("the command names no <spec>")
	}
	git := envGit(env)
	var all []string
	receipts := map[string]*rcReceipt{} // by source worktree
	var order []string
	for _, pair := range strings.Split(args[2], ",") {
		wt, base, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if !ok || wt == "" || base == "" {
			return usage("not <source-worktree>=<base>: " + pair)
		}
		var out bytes.Buffer
		if rc := Registry["affected-specs"]([]string{"-task", task, wt, base, root}, env, &out, stderr); rc != 0 {
			return rc
		}
		for _, s := range strings.Fields(out.String()) {
			if !slices.Contains(all, s) {
				all = append(all, s)
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
	sort.Strings(all)

	// ONLY WHAT HAS NOT PASSED RUNS.
	var specs []string
	repeat := false
	for _, s := range all {
		for _, wt := range order {
			same := slices.DeleteFunc(rcReceipts(env, wt), func(r rcReceipt) bool { return !maps.Equal(r.Blobs, receipts[wt].Blobs) })
			left, failed := rcLeft(s, same)
			for _, l := range left {
				if !slices.Contains(specs, l) {
					specs = append(specs, l)
				}
			}
			repeat = repeat || (failed && len(left) > 0)
		}
	}

	rc := 0
	var failed []string
	switch {
	case len(all) == 0:
		fmt.Fprintln(stdout, "recapture: no affected spec — nothing to capture")
	case len(specs) == 0:
		fmt.Fprintf(stdout, "recapture: all %d test selection(s) already passed on these sources — nothing to capture\n", len(all))
	default:
		if repeat {
			command += " --repeat-each 5 --workers 1"
		}
		var log bytes.Buffer
		cmd := exec.Command("sh", "-c", strings.ReplaceAll(command, "<spec>", strings.Join(specs, " ")))
		cmd.Dir = pcAbs(env, root)
		cmd.Stdout, cmd.Stderr = io.MultiWriter(stdout, &log), io.MultiWriter(stderr, &log)
		if rc = exitCode(cmd.Run()); rc != 0 {
			failed = rcFailed(log.Bytes(), specs)
		}
		fmt.Fprintf(stdout, "recapture: %d test selection(s), exit %d\n", len(specs), rc)
	}
	for _, wt := range order {
		r := receipts[wt]
		r.Specs, r.Exit, r.Failed = specs, rc, failed
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

// rcCovered reports whether worktree's receipts that saw the blobs commit
// sha holds for its own source paths leave nothing of specs to run (rcLeft).
func rcCovered(env Env, worktree, captureRoot, sha string, specs []string) bool {
	git := envGit(env)
	names, _ := capture(git("-C", worktree, "diff-tree", "--no-renames", "--no-commit-id", "--name-only", "-r", sha))
	want := map[string]string{}
	for _, f := range rcSourcePaths(env, worktree, captureRoot, strings.Split(names, "\n")) {
		want[f], _ = capture(git("-C", worktree, "rev-parse", "--verify", "--quiet", sha+":"+f))
	}
	receipts := slices.DeleteFunc(rcReceipts(env, worktree), func(r rcReceipt) bool {
		for f, blob := range want {
			if got, seen := r.Blobs[f]; !seen || got != blob {
				return true
			}
		}
		return false
	})
	return !slices.ContainsFunc(specs, func(s string) bool {
		left, _ := rcLeft(s, receipts)
		return len(left) > 0
	})
}

// rcCaptured is the specs a passing receipt already covers on what every
// worktree of bases -- worktree to round base -- holds at HEAD: each worktree
// whose `<base>..HEAD` holds a source file has a receipt that passed the spec
// on exactly those files, each at HEAD's blob. Nil when no range holds one.
// A spec only some tests of which a receipt took is never among them.
// ponytail: equal blob maps, not a run instant -- a file reverted and
// re-applied across rounds matches the older receipt; record HEAD's tree in
// the receipt if that ever bites.
func rcCaptured(env Env, captureRoot string, bases map[string]string) []string {
	git := envGit(env)
	var out []string
	ranged := false
	for wt, base := range bases {
		names, _ := capture(git("-C", wt, "diff", "--name-only", "--no-renames", base, "HEAD"))
		want := map[string]string{}
		for _, f := range rcSourcePaths(env, wt, captureRoot, strings.Split(names, "\n")) {
			want[f], _ = capture(git("-C", wt, "rev-parse", "--verify", "--quiet", "HEAD:"+f))
		}
		if len(want) == 0 {
			continue
		}
		var here []string
		same := slices.DeleteFunc(rcReceipts(env, wt), func(r rcReceipt) bool { return !maps.Equal(r.Blobs, want) })
		for _, r := range same {
			for _, s := range r.Specs {
				if left, _ := rcLeft(s, same); len(left) == 0 && s == rcSpecFile(s) && !slices.Contains(here, s) {
					here = append(here, s)
				}
			}
		}
		if ranged {
			here = slices.DeleteFunc(here, func(s string) bool { return !slices.Contains(out, s) })
		}
		out, ranged = here, true
	}
	sort.Strings(out)
	return out
}
