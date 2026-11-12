package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/check-plan-unchanged.sh shim (KAN-1058).
//
// flow-guard check-plan-unchanged snapshot|verify <worktree> <change-name> <snapshot-file>
//
// The guard KAN-635 asked for: the change's plan tree survives every
// reviewer dispatch, asserted by this guard — never by the reviewer's own prose.
// On gymie KAN-579 the gated per-task reviewer ran `git checkout <sha> -- .`
// mid-review, silently destroyed the never-committed design.md/tasks.md
// working-tree edits (task commits never carry plan paths, per
// flow-guard check-task-commit-planning-paths, so the working tree held the only
// current copy), and then self-reported the tree restored — a claim nothing
// checked, and a false one. implement.md's **The plan tree survives every
// reviewer dispatch** and review-panel.md's round-dispatch paragraph wire
// the two modes around every review dispatch:
//
//   snapshot  run immediately BEFORE the dispatch's Agent call, after the
//             parent has committed any uncommitted plan edits — writes the
//             plan tree's content (every regular file under it, walked from
//             disk since the tree is gitignored, each path then its bytes) to
//             <snapshot-file>
//   verify    run when the dispatch's report file(s) exist, BEFORE any
//             verdict is read or acted on — recomputes the same content and
//             compares it, naming each changed path; a checkbox tick
//             (`- [ ] ` to `- [x] `, the parent's `flow tasks tick` during
//             the flight), an inserted `Correction (YYYY-MM-DD):` paragraph
//             and a `**Files:**` line widened without dropping an entry (the
//             parent's task-close edits, kan-1107 self-review F2) are not a
//             change, any other difference is
//
// THE GUARD IS SCOPED TO THE PLAN TREE (`spectre/changes/<name>/`, on disk
// and gitignored), deliberately, and never asserts HEAD or the rest of the worktree: a
// gated reviewer bundle flies BESIDE the next implementer group (implement.md's
// bundled dispatch), whose source edits and commits legitimately move
// everything else while the reviewer is in flight — a whole-worktree assert
// would fire on every normal run, and the incident's own loss was the plan
// tree ("source code was never at risk", the KAN-635 issue records). The
// snapshot is the plan tree's content, every file on disk, so a file
// planted under the plan tree is a violation too, and so is a prose edit
// beside a checkbox tick — only a line whose sole change is a tick, an
// inserted `Correction (YYYY-MM-DD):` paragraph and a `**Files:**` line
// widened without dropping an entry pass (kan-1107 self-review F2);
// the plan directory being absent at snapshot and still absent at verify is
// clean (nothing was there to lose).
//
// THE SNAPSHOT IS THE EVIDENCE, and verify is the only restore claim this
// pipeline accepts: the dispatching parent never reads the reviewer's
// "restored"/"clean" prose as an answer to what happened to the tree. Verify
// exit 1 therefore stops the run — the reviewer reported about a tree it
// also changed — and recovery is by hand from the pre-dispatch commit (which
// holds the plan tree precisely because the parent commits before it
// snapshots), never by re-running the reviewer.
//
// THE CHANGE NAME IS A PATH COMPONENT here, so the package's one
// change-name allowlist (stats/internal/guard/changename.go's
// plainChangeName) refuses anything but a plain change name before the walk
// ever sees it.
//
// Exit codes:
//   0  snapshot written, or the plan tree matches its snapshot
//   1  verify found the plan tree changed since the snapshot — each
//      difference named on stderr
//   2  cannot answer at all — missing arguments, a bad mode, a non-directory
//      worktree, a change name outside the allowlist, a plan tree that cannot
//      be walked, or (on verify) a missing or unreadable snapshot file or an
//      unwritable snapshot path (on snapshot)
//
// The guard itself is Go: stats/internal/guard/planunchanged.go.

// checkPlanUnchanged is `flow-guard check-plan-unchanged`: the comment above is the contract -- snapshot writes the plan tree's content
// to a file, verify recomputes it and compares it, a checkbox tick
// (`- [ ] ` to `- [x] `) not counting as a change; exit 0 written or
// unchanged, 1 changed (each difference on stderr), 2 cannot answer. The
// reasoning for each step, moved here from the bash body it replaced
// (c9b8fde8), sits beside the code it explains.

func init() {
	Registry["check-plan-unchanged"] = checkPlanUnchanged
}

func checkPlanUnchanged(args []string, env Env, stdout, stderr io.Writer) int {
	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}
	mode, worktree, name, snap := arg(0), arg(1), arg(2), arg(3)
	usage := func() int {
		fmt.Fprintln(stderr, "check-plan-unchanged: usage: flow-guard check-plan-unchanged snapshot|verify <worktree> <change-name> <snapshot-file>")
		return 2
	}
	if mode != "snapshot" && mode != "verify" {
		return usage()
	}
	if worktree == "" || !isDir(smcAbs(env, worktree)) {
		shown := worktree
		if shown == "" {
			shown = "<missing>"
		}
		fmt.Fprintf(stderr, "check-plan-unchanged: not a directory: %s\n", shown)
		return 2
	}
	if name == "" || snap == "" {
		return usage()
	}

	// CONTAINMENT: the name lands in a filesystem path, so `../escape` is a
	// hazard here exactly as it is in a
	// state-file-derived change name elsewhere. plainChangeName is the
	// package's one copy of the allowlist (changename.go is canonical for
	// why it is enumerated).
	if !plainChangeName(name) {
		fmt.Fprintf(stderr, "check-plan-unchanged: change name '%s' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n", name)
		return 2
	}

	// Canonicalise before the walk (`cd && pwd -P`), and refuse
	// rather than proceed if it vanished between the directory check above
	// and here.
	wt, err := filepath.EvalSymlinks(smcAbs(env, worktree))
	if err == nil {
		wt, err = filepath.Abs(wt)
	}
	if err != nil {
		fmt.Fprintf(stderr, "check-plan-unchanged: worktree vanished: %s\n", worktree)
		return 2
	}

	// The snapshot is the plan tree's content: every regular file and
	// symlink under it (a symlink read through), walked from disk -- the tree is gitignored (KAN-1073), so git
	// lists none of it -- sorted by repository-relative path, each as
	// `== <path>` then its bytes, or `== <path> (missing)` when walked but
	// unreadable. An absent plan tree is an empty dump.
	planStatus := func(out io.Writer) bool {
		root := filepath.Join(wt, "spectre", "changes", name)
		var paths []string
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if p == root && errors.Is(err, fs.ErrNotExist) {
					return fs.SkipAll
				}
				return err
			}
			if d.Type().IsRegular() || d.Type()&fs.ModeSymlink != 0 {
				rel, err := filepath.Rel(wt, p)
				if err != nil {
					return err
				}
				paths = append(paths, filepath.ToSlash(rel))
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(stderr, "check-plan-unchanged: %v\n", err)
			return false
		}
		sort.Strings(paths)
		var dump bytes.Buffer
		for _, p := range paths {
			b, err := os.ReadFile(filepath.Join(wt, p))
			if err != nil {
				fmt.Fprintf(&dump, "== %s (missing)\n", p)
				continue
			}
			fmt.Fprintf(&dump, "== %s\n", p)
			dump.Write(b)
			if len(b) > 0 && b[len(b)-1] != '\n' {
				dump.WriteByte('\n')
			}
		}
		_, err = out.Write(dump.Bytes())
		return err == nil
	}
	cannotRead := func() int {
		fmt.Fprintf(stderr, "check-plan-unchanged: the plan tree under %s could not be read, or its content "+
			"could not be written -- cannot answer\n", wt)
		return 2
	}
	snapPath := smcAbs(env, snap)

	if mode == "snapshot" {
		// `plan_status > "$SNAP"`: the file is opened (truncated) before the
		// walk, so an unwritable path never reaches it, and any failure
		// removes the file rather than leave a partial snapshot behind.
		f, err := os.Create(snapPath)
		if err != nil {
			fmt.Fprintf(stderr, "check-plan-unchanged: %s: %v\n", snap, err)
			_ = os.Remove(snapPath)
			return cannotRead()
		}
		ok := planStatus(f)
		if cerr := f.Close(); cerr != nil {
			ok = false
		}
		if !ok {
			_ = os.Remove(snapPath)
			return cannotRead()
		}
		fmt.Fprintf(stdout, "PLAN-SNAPSHOT-OK: spectre/changes/%s content pinned at %s\n", name, snap)
		return 0
	}

	// THE SNAPSHOT IS THE EVIDENCE: a missing or unreadable one is cannot
	// answer -- snapshot runs before the dispatch, never after it.
	want, err := os.ReadFile(snapPath)
	if fi, serr := os.Stat(snapPath); serr != nil || !fi.Mode().IsRegular() || err != nil {
		fmt.Fprintf(stderr, "check-plan-unchanged: no readable snapshot at %s -- cannot answer; snapshot runs before the dispatch, never after it\n", snap)
		return 2
	}
	var got bytes.Buffer
	if !planStatus(&got) {
		return cannotRead()
	}
	if bytes.Equal(want, got.Bytes()) {
		fmt.Fprintf(stdout, "PLAN-UNCHANGED-OK: spectre/changes/%s matches its pre-dispatch snapshot\n", name)
		return 0
	}
	if closeEditsOnly(string(want), got.String()) {
		fmt.Fprintf(stdout, "PLAN-UNCHANGED-OK: spectre/changes/%s matches its pre-dispatch snapshot (checkbox ticks and close-time corrections only)\n", name)
		return 0
	}
	fmt.Fprintf(stderr, "check-plan-unchanged: spectre/changes/%s changed across the review dispatch -- the reviewer's own clean-or-restored claim is not evidence; stop and recover by hand from the pre-dispatch commit\n", name)
	// The content dump's diff lines carry no path, so each changed path is
	// named first.
	for _, p := range changedPaths(string(want), got.String()) {
		fmt.Fprintf(stderr, "check-plan-unchanged: changed: %s\n", p)
	}
	// `diff "$SNAP" "$current" | sed 's/^/check-plan-unchanged: /' >&2`:
	// diff(1)'s normal format names each difference; the current plan
	// content goes through a temp file, as the bash's mktemp did.
	tmpdir := env.Getenv("TMPDIR")
	if tmpdir == "" {
		tmpdir = "/tmp"
	}
	cur, err := os.CreateTemp(tmpdir, "check-plan-unchanged-verify.")
	if err != nil {
		fmt.Fprintln(stderr, "check-plan-unchanged: mktemp failed -- cannot answer")
		return 2
	}
	defer os.Remove(cur.Name())
	_, werr := cur.Write(got.Bytes())
	if cerr := cur.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		fmt.Fprintln(stderr, "check-plan-unchanged: mktemp failed -- cannot answer")
		return 2
	}
	diffBin := "diff"
	if p, ok := lookPath(env, "diff"); ok {
		diffBin = p
	}
	d := exec.Command(diffBin, snapPath, cur.Name())
	d.Env = append(os.Environ(), "LC_ALL=C")
	out, _ := d.Output()
	for _, l := range strings.SplitAfter(string(out), "\n") {
		if l != "" {
			fmt.Fprint(stderr, "check-plan-unchanged: "+l)
		}
	}
	return 1
}

// closeEditsOnly reports whether cur differs from old only by the parent's
// task-close edits (kan-1107 self-review F2), walking old line i and cur
// line j: equal lines, a checkbox tick (an old line opening `- [ ] ` after
// its indent with that first `- [ ] ` turned to `- [x] ` -- the parent's
// `flow tasks tick`; no step unticks) and a `**Files:**` line whose backticked
// entries are a superset of the old line's advance both; an inserted run of
// optional blank lines, a `Correction (YYYY-MM-DD):` line, the lines up to the
// next blank line and any blank lines bordering it advances j alone. Anything
// else -- a deleted, reworded or narrowed line, any other insertion -- fails.
func closeEditsOnly(old, cur string) bool {
	o, c := strings.Split(old, "\n"), strings.Split(cur, "\n")
	var walk func(i, j int) bool
	walk = func(i, j int) bool {
		if i == len(o) && j == len(c) {
			return true
		}
		if i < len(o) && j < len(c) && (o[i] == c[j] || isTick(o[i], c[j]) || filesWidened(o[i], c[j])) && walk(i+1, j+1) {
			return true
		}
		// A bullet of backticked entries inserted under a multi-line
		// `**Files:**` block widens it.
		if j > 0 && j < len(c) && filesBulletRE.MatchString(c[j]) &&
			(strings.HasPrefix(c[j-1], "**Files:**") || filesBulletRE.MatchString(c[j-1])) && walk(i, j+1) {
			return true
		}
		k := j
		for k < len(c) && c[k] == "" {
			k++
		}
		if k == len(c) || !correctionRE.MatchString(c[k]) {
			return false
		}
		for k < len(c) && c[k] != "" {
			k++
		}
		// The blank lines after the paragraph may border it or belong to
		// the old text: try each split.
		for ; ; k++ {
			if walk(i, k) {
				return true
			}
			if k == len(c) || c[k] != "" {
				return false
			}
		}
	}
	return walk(0, 0)
}

var (
	correctionRE = regexp.MustCompile(`^Correction \(\d{4}-\d{2}-\d{2}\):`)
	backtickRE   = regexp.MustCompile("`([^`]*)`")
	// filesBulletRE: a multi-line `**Files:**` entry, `- Create: `x``,
	// holding backticked entries only.
	filesBulletRE = regexp.MustCompile("^- [A-Za-z]+: `[^`]+`(,? `[^`]+`)*$")
)

func isTick(o, c string) bool {
	return strings.HasPrefix(strings.TrimLeft(o, " "), "- [ ] ") && c == strings.Replace(o, "- [ ] ", "- [x] ", 1)
}

// filesWidened: both lines open `**Files:**` and every backticked entry of
// the old one is in the new one.
func filesWidened(o, c string) bool {
	if !strings.HasPrefix(o, "**Files:**") || !strings.HasPrefix(c, "**Files:**") {
		return false
	}
	have := map[string]bool{}
	for _, m := range backtickRE.FindAllStringSubmatch(c, -1) {
		have[m[1]] = true
	}
	for _, m := range backtickRE.FindAllStringSubmatch(o, -1) {
		if !have[m[1]] {
			return false
		}
	}
	return true
}

// changedPaths names, sorted, each path whose `== <path>` section differs
// between two content dumps, or is in one only.
func changedPaths(old, cur string) []string {
	sections := func(dump string) map[string]string {
		m := map[string]string{}
		key := ""
		for _, l := range strings.SplitAfter(dump, "\n") {
			if strings.HasPrefix(l, "== ") {
				key = strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(l, "== "), "\n"), " (missing)")
			}
			m[key] += l
		}
		return m
	}
	o, c := sections(old), sections(cur)
	var out []string
	for k, v := range o {
		if c[k] != v {
			out = append(out, k)
		}
	}
	for k := range c {
		if _, ok := o[k]; !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
