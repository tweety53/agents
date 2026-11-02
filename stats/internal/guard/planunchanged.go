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
	"sort"
	"strings"
)

// checkPlanUnchanged is scripts/check-plan-unchanged.sh: that script's
// header comment is the contract -- snapshot writes the plan tree's content
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
	if ticksOnly(string(want), got.String()) {
		fmt.Fprintf(stdout, "PLAN-UNCHANGED-OK: spectre/changes/%s matches its pre-dispatch snapshot (checkbox ticks only)\n", name)
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

// ticksOnly reports whether cur differs from old only by checkbox ticks: the
// same line count, and every differing line is old's line, opening `- [ ] `
// after its indent, with that first `- [ ] ` turned to `- [x] `. The parent
// ticks a task (`flow tasks tick`) while a reviewer flies; no step unticks.
func ticksOnly(old, cur string) bool {
	o, c := strings.Split(old, "\n"), strings.Split(cur, "\n")
	if len(o) != len(c) {
		return false
	}
	for i := range o {
		if o[i] == c[i] {
			continue
		}
		if !strings.HasPrefix(strings.TrimLeft(o[i], " "), "- [ ] ") || c[i] != strings.Replace(o[i], "- [ ] ", "- [x] ", 1) {
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
