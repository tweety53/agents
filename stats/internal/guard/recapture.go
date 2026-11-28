package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// flow-guard recapture <tasks.md>:<task-id> <capture-root> <source-worktree>=<base>[,<source-worktree>=<base>…] <command>
//
// Runs an in-run fix task's capture run and leaves the receipt close-task
// asks for (skills/flow/verify-fix-loop.md step 2). The task's tests are
// flow-guard affected-specs' task set for each pair -- `affected-specs -task
// <tasks.md>:<task-id> <source-worktree> <base> <capture-root>` -- united and
// sorted, each a spec or one test of it, `<spec>:<line>` -- or, when that
// selects nothing, every `*.spec.ts` under <capture-root> whose source holds
// a backticked Tests/Regression name that is no path (a test title) as a
// whole quoted string; <command> is the
// project's capture command, or its re-baseline command, with `<spec>` where
// they go. It runs once, through `sh -c` in <capture-root>, with `<spec>`
// replaced by them space-joined and followed by ` --retries=1`, its output
// passed through. A test that passes on its retry is flaky, never failed:
// the receipt counts it passed and the run prints
// `recapture: flaky (<n>): <spec>:<line> … — passed on retry: fix each at
// its source` (kan-1091: one flaky spec in ~680 voided a whole recapture
// three times).
//
// ONLY WHAT HAS NOT PASSED RUNS. A test an earlier receipt of every pair's
// worktree records as passed on the sources as they stand now -- the same
// source files of the diff, each at the same blob -- is left out, whether
// that run took it alone or inside its spec, so a red run is followed by a
// run of its failing tests alone. With nothing left to run nothing runs.
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
// THE RUN'S COST is one `flow record run -kind recapture` row, labelled
// with the task id, as flow-guard run-verify-list files its own (THE RUN'S
// COST there, canonical), under the first pair's worktree.
//
// THE FAILED TESTS of a red run are the `<spec>:<line>` entries its
// Playwright summary names (flow-guard run-verify-list's failed and
// interrupted blocks); every other test of the run passed. A red run whose
// summary names no test, names a file the run did not take, or counts a test
// that did not run, failed everything it took.
//
// Exit codes:
//
//	0  the command exited 0, or there was nothing to capture.
//	1  the command exited non-zero; the receipt records the exit. Or the
//	   task's diff reaches no spec, its Tests/Regression name test titles,
//	   and no spec under <capture-root> holds one as a quoted string:
//	   `recapture: task <id> names tests no spec under <capture-root> holds:
//	   <title>, …`, nothing run, no receipt written.
//	2  usage, a <command> with no `<spec>`, affected-specs' own exit 2, or
//	   a receipt that could not be written.
//	3  no heavy slot freed (HeavySlot): nothing ran and no receipt was
//	   written.
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
// later run passes it.
func rcLeft(token string, receipts []rcReceipt) []string {
	left := []string{token}
	for _, r := range receipts {
		var next []string
		for _, x := range left {
			switch {
			case !slices.Contains(r.Specs, x) && !slices.Contains(r.Specs, rcSpecFile(x)):
				next = append(next, x)
			case r.Exit == 0:
			case len(r.Failed) == 0:
				next = append(next, x)
			default:
				for _, f := range r.Failed {
					if f == x || f == rcSpecFile(x) {
						f = x
					} else if rcSpecFile(f) != x {
						continue
					}
					if !slices.Contains(next, f) {
						next = append(next, f)
					}
				}
			}
		}
		left = next
	}
	return left
}

// rcFile is a worktree's receipt file, "" when git cannot name its directory.
func rcFile(env core.Env, worktree string) string {
	dir, ok := core.Capture(core.EnvGit(env)("-C", worktree, "rev-parse", "--absolute-git-dir"))
	if !ok || dir == "" {
		return ""
	}
	return filepath.Join(dir, "flow-recapture.jsonl")
}

// rcReceipts is a worktree's receipts, oldest first.
func rcReceipts(env core.Env, worktree string) []rcReceipt {
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

// rcBlobs is each of paths' object at rev in worktree -- what `git rev-parse
// --verify --quiet <rev>:<path>` prints for it, "" where rev holds no such
// path -- read through one `git cat-file --batch-check` rather than one exec
// per path. Paths come from line-split git output, so none holds a newline.
func rcBlobs(git func(...string) *exec.Cmd, worktree, rev string, paths []string) map[string]string {
	out := map[string]string{}
	if len(paths) == 0 {
		return out
	}
	var in strings.Builder
	for _, f := range paths {
		in.WriteString(rev + ":" + f + "\n")
	}
	cmd := git("-C", worktree, "cat-file", "--batch-check=%(objectname)")
	cmd.Stdin = strings.NewReader(in.String())
	list, _ := core.Capture(cmd)
	lines := strings.Split(list, "\n")
	for i, f := range paths {
		out[f] = ""
		if i < len(lines) && !strings.HasSuffix(lines[i], " missing") && !strings.HasSuffix(lines[i], " ambiguous") {
			out[f] = lines[i]
		}
	}
	return out
}

// rcSourcePaths keeps the paths of a diff that are sources: no capture
// output, and no test outside the capture root.
func rcSourcePaths(env core.Env, worktree, captureRoot string, paths []string) []string {
	wt, root := filepath.Clean(core.CleanAbs(env, worktree)), filepath.Clean(core.CleanAbs(env, captureRoot))
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
	out := rvlSummaryBlocks(log, "failed", "interrupted")
	if slices.ContainsFunc(out, func(e string) bool {
		return !slices.ContainsFunc(specs, func(s string) bool { return rcSpecFile(s) == rcSpecFile(e) })
	}) {
		return nil
	}
	return out
}

func recapture(args []string, env core.Env, stdout, stderr io.Writer) int {
	usage := func(msg string) int {
		fmt.Fprintf(stderr, "recapture: %s\nusage: flow-guard recapture <tasks.md>:<task-id> <capture-root> <source-worktree>=<base>[,<source-worktree>=<base>…] <command with <spec>>\n", msg)
		core.Reason(env, "usage")
		return 2
	}
	if len(args) != 4 {
		return usage(fmt.Sprintf("got %d argument(s)", len(args)))
	}
	task, root, command := args[0], args[1], args[3]
	if !strings.Contains(command, "<spec>") {
		return usage("the command names no <spec>")
	}
	git := core.EnvGit(env)
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
		diff, _ := core.Capture(git("-C", wt, "diff", "--name-only", "--no-renames", base))
		diff += "\n" + asUntracked(env, wt)
		r := &rcReceipt{Blobs: map[string]string{}}
		for _, f := range rcSourcePaths(env, wt, root, strings.Split(diff, "\n")) {
			r.Blobs[f], _ = core.Capture(git("-C", wt, "hash-object", "--", f))
		}
		receipts[wt] = r
		order = append(order, wt)
	}
	if len(all) == 0 {
		titles, found := rcTitleSpecs(task, core.CleanAbs(env, root))
		if len(titles) > 0 && len(found) == 0 {
			fmt.Fprintf(stdout, "recapture: task %s names tests no spec under %s holds: %s\n", task[strings.LastIndex(task, ":")+1:], root, strings.Join(titles, ", "))
			core.Reason(env, "no-spec-for-title")
			return 1
		}
		all = found
	}
	sort.Strings(all)

	// ONLY WHAT HAS NOT PASSED RUNS.
	var specs []string
	for _, s := range all {
		for _, wt := range order {
			same := slices.DeleteFunc(rcReceipts(env, wt), func(r rcReceipt) bool { return !maps.Equal(r.Blobs, receipts[wt].Blobs) })
			for _, l := range rcLeft(s, same) {
				if !slices.Contains(specs, l) {
					specs = append(specs, l)
				}
			}
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
		release, ok := core.HeavySlot(env, "recapture", "recapture "+task, stderr)
		if !ok {
			core.Reason(env, "no-heavy-slot")
			return 3
		}
		defer release()
		var log bytes.Buffer
		start := time.Now()
		cmd := exec.Command("sh", "-c", strings.ReplaceAll(command, "<spec>", strings.Join(specs, " ")+" --retries=1"))
		cmd.Dir = core.CleanAbs(env, root)
		cmd.Stdout, cmd.Stderr = io.MultiWriter(stdout, &log), io.MultiWriter(stderr, &log)
		if rc = core.ExitCode(cmd.Run()); rc != 0 {
			failed = rcFailed(log.Bytes(), specs)
		}
		fmt.Fprintf(stdout, "recapture: %d test selection(s), exit %d\n", len(specs), rc)
		branch, _ := core.Capture(git("-C", order[0], "branch", "--show-current"))
		rvlRecordRun(env, order[0], strings.TrimPrefix(branch, "spectre/"), "recapture", task[strings.LastIndex(task, ":")+1:], start, log.Bytes(), rc)
		if flaky := rvlSummaryBlocks(log.Bytes(), "flaky"); len(flaky) > 0 {
			fmt.Fprintf(stdout, "recapture: flaky (%d): %s — passed on retry: fix each at its source\n", len(flaky), strings.Join(flaky, " "))
		}
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
			core.Reason(env, "cannot-write-receipt")
			return 2
		}
	}
	if rc != 0 {
		core.Reason(env, "capture-failed")
		return 1
	}
	return 0
}

// rcIsTitle reports whether n can be a Playwright test title.
func rcIsTitle(n string) bool {
	return strings.Contains(n, " ") && !strings.ContainsAny(n, "/{};=") && !strings.HasSuffix(n, ".ts")
}

// rcTitleSpecs is the task's backticked Tests/Regression names that can be
// Playwright test titles -- holding a space, and no `/`, `.ts` suffix or code
// punctuation, so a Go test name, a `<spec>:<line>` ref or a code fragment is
// none -- and every `*.spec.ts` under the capture root, relative
// to it, whose source holds one of them as a whole quoted string.
func rcTitleSpecs(task, root string) (titles, specs []string) {
	named, _ := asTaskNames(task)
	for _, m := range regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(named, -1) {
		if n := m[1]; rcIsTitle(n) && !slices.Contains(titles, n) {
			titles = append(titles, n)
		}
	}
	if len(titles) == 0 {
		return nil, nil
	}
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if n := d.Name(); n == "node_modules" || n == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".spec.ts") {
			return nil
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		for _, t := range titles {
			for _, q := range []string{`"`, "'", "`"} {
				if bytes.Contains(src, []byte(q+t+q)) {
					rel, _ := filepath.Rel(root, p)
					specs = append(specs, rel)
					return nil
				}
			}
		}
		return nil
	})
	return titles, specs
}

// errorsJoinClose closes f and returns err, else the close's error.
func errorsJoinClose(err error, f *os.File) error {
	if cerr := f.Close(); err == nil {
		return cerr
	}
	return err
}

// rcCommitReceipts is worktree's receipts that saw the blobs commit sha
// holds for its own source paths, and whether the commit holds any.
func rcCommitReceipts(env core.Env, worktree, captureRoot, sha string) ([]rcReceipt, bool) {
	git := core.EnvGit(env)
	names, _ := core.Capture(git("-C", worktree, "diff-tree", "--no-renames", "--no-commit-id", "--name-only", "-r", sha))
	want := rcBlobs(git, worktree, sha, rcSourcePaths(env, worktree, captureRoot, strings.Split(names, "\n")))
	return slices.DeleteFunc(rcReceipts(env, worktree), func(r rcReceipt) bool {
		for f, blob := range want {
			if got, seen := r.Blobs[f]; !seen || got != blob {
				return true
			}
		}
		return false
	}), len(want) > 0
}

// rcCovered reports whether worktree's receipts that saw the blobs commit
// sha holds for its own source paths leave nothing of specs to run (rcLeft).
func rcCovered(env core.Env, worktree, captureRoot, sha string, specs []string) bool {
	receipts, _ := rcCommitReceipts(env, worktree, captureRoot, sha)
	return !slices.ContainsFunc(specs, func(s string) bool {
		return len(rcLeft(s, receipts)) > 0
	})
}

// rcTaskCaptured is the specs of captureRoot whose capture a task of the
// change took and nothing has moved since, so a pass of the whole change
// need not take it again (skills/flow/visual-verify-verifier.md step 8;
// kan-1091: the stage recaptured every spec its tasks had just captured).
// bases is each worktree of the change and its merge base. In a worktree a
// spec's anchor is the newest commit since the base that holds a source
// file and whose receipts passed the whole spec on that commit's sources
// (rcCovered). A spec stands when at least one worktree anchors it and, in
// every worktree, flow-guard affected-specs selects nothing of it over
// `<anchor>..HEAD` -- over `<base>..HEAD` where the worktree has no anchor
// for it: a shared file a later task edits moves every spec it reaches. Nil
// when a worktree's commits or affected specs cannot be read.
// ponytail: a capture is taken as the task commit's state -- a sibling
// worktree's uncommitted edit at capture time is not seen; record every
// worktree's HEAD in the receipt if that ever bites.
func rcTaskCaptured(env core.Env, captureRoot string, bases map[string]string) []string {
	git := core.EnvGit(env)
	stands, moved := map[string]bool{}, map[string]bool{}
	for wt, base := range bases {
		log, ok := core.Capture(git("-C", wt, "rev-list", base+"..HEAD"))
		if !ok {
			return nil
		}
		affected := map[string]map[string]bool{} // by range base
		reach := func(from string) (map[string]bool, bool) {
			if affected[from] == nil {
				var out bytes.Buffer
				if Registry["affected-specs"]([]string{wt, from + "..HEAD", captureRoot}, core.QuietReasons(env), &out, io.Discard) != 0 {
					return nil, false
				}
				affected[from] = map[string]bool{}
				for _, t := range strings.Fields(out.String()) {
					affected[from][rcSpecFile(t)] = true
				}
			}
			return affected[from], true
		}
		anchor := map[string]string{}
		for _, sha := range strings.Fields(log) {
			receipts, sourced := rcCommitReceipts(env, wt, captureRoot, sha)
			if !sourced {
				continue
			}
			for _, r := range receipts {
				for _, s := range r.Specs {
					if anchor[s] == "" && s == rcSpecFile(s) && len(rcLeft(s, receipts)) == 0 {
						anchor[s] = sha
					}
				}
			}
		}
		whole, ok := reach(base)
		if !ok {
			return nil
		}
		for s := range whole {
			moved[s] = moved[s] || anchor[s] == ""
		}
		for s, sha := range anchor {
			since, ok := reach(sha)
			if !ok {
				return nil
			}
			stands[s], moved[s] = true, moved[s] || since[s]
		}
	}
	var out []string
	for s := range stands {
		if !moved[s] {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// rcCaptured is the specs a passing receipt already covers on what every
// worktree of bases -- worktree to round base -- holds at HEAD: each worktree
// whose `<base>..HEAD` holds a source file has a receipt that passed the spec
// on exactly those files, each at HEAD's blob. Nil when no range holds one.
// A spec only some tests of which a receipt took is never among them.
// ponytail: equal blob maps, not a run instant -- a file reverted and
// re-applied across rounds matches the older receipt; record HEAD's tree in
// the receipt if that ever bites.
func rcCaptured(env core.Env, captureRoot string, bases map[string]string) []string {
	git := core.EnvGit(env)
	var out []string
	ranged := false
	for wt, base := range bases {
		names, _ := core.Capture(git("-C", wt, "diff", "--name-only", "--no-renames", base, "HEAD"))
		want := rcBlobs(git, wt, "HEAD", rcSourcePaths(env, wt, captureRoot, strings.Split(names, "\n")))
		if len(want) == 0 {
			continue
		}
		var here []string
		same := slices.DeleteFunc(rcReceipts(env, wt), func(r rcReceipt) bool { return !maps.Equal(r.Blobs, want) })
		for _, r := range same {
			for _, s := range r.Specs {
				if len(rcLeft(s, same)) == 0 && s == rcSpecFile(s) && !slices.Contains(here, s) {
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
