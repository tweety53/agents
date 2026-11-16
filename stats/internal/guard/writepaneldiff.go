package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// The contract below moved here from the header of the deleted
// scripts/write-panel-diff.sh shim (KAN-1058).
//
// flow-guard write-panel-diff — write one review-panel diff and its touched list
// (skills/flow/review-panel.md).
//
// Usage:
//   flow-guard write-panel-diff final      <canonical-wt> <wt> <merge-base> [<wt> <merge-base>...]
//   flow-guard write-panel-diff late-fix   <canonical-wt> <wt> <since-close-sha> [...]
//   flow-guard write-panel-diff fix-round <N> <canonical-wt> <wt> <fix-base> [...]
//   flow-guard write-panel-diff slot-delta <round> <slot> <canonical-wt> <wt> <merge-base> <held-sha|-> [...]
//   flow-guard write-panel-diff parts      <canonical-wt> <cap>
//
// Writes two files into <canonical-wt>/.superpowers/sdd/, each atomically
// (a temp file renamed over the target), and prints both paths, one per line:
//   final-review.diff | late-fix.diff | fix-round-<N>.diff |
//   slot-delta-<round>-<slot>.diff
//   <that file>.touched — the [TOUCHED_FILES] list
// A `final` write also removes any late-fix.diff and its .touched, and every
// final-review-part-*.diff and its .touched: a full pass 1 makes them stale,
// and a slot must never read them in its place.
//
// `parts` reads <canonical-wt>/.superpowers/sdd/final-review.diff and its
// .touched, which must hold one `# worktree:` section, and cuts the
// section's `diff --git ` blocks, in order, into parts of at most <cap>
// changed lines (`+`/`-` lines, the `+++ `/`--- ` headers excluded — what
// check-panel-diff-size's --shortstat counts); a block joins the current
// part unless that would take it over <cap> and the part is non-empty, so a
// file over the cap is a part of its own. Each part is written atomically
// beside the diff as final-review-part-<k>.diff, <k> from 1 — the semantics
// line, the section header and its blocks — and final-review-part-<k>.diff.touched,
// the touched header and the --name-status lines naming one of the part's
// paths; each part's diff path is printed, one per line. It lets an over-cap
// single-worktree panel read its diff in parallel parts (kan-1090
// self-review, F9).
//
// Both are sectioned per worktree, in argument order, each section opened by
//   # worktree: <wt> — merge base <sha>
// naming the worktree's own second argument (its merge base, since-close sha
// or fix base). final-review.diff alone opens with the semantics line
//   # final-review.diff — working tree vs merge-base, unstaged changes included
// Each diff section is that worktree's `git diff` over the kind's range, and
// each .touched section `git diff --name-status` over the same range:
//   final       git diff <merge-base>          (the working tree: staged and unstaged)
//   late-fix    git diff <since-close-sha>     (the working tree)
//   fix-round   git diff <fix-base>..HEAD
//   slot-delta  git diff <held-sha> HEAD; a `-` held sha — a worktree in
//               which the slot holds no sha — reads git diff <merge-base>
// Every diff section (never the .touched one) runs with --irreversible-delete,
// so a deleted file shows as its header alone; its old body is
// `git show <sha>:<path>`, the sha the section's diff range starts from —
// for slot-delta with a held sha, the held sha; otherwise the sha its
// section's `# worktree:` header names.
//
// A worktree whose own .flow/project.md declares `## panel diff exclude`
// has every path one of that fence's globs matches left out of both its
// sections: each git diff above then ends `-- . ':(glob,exclude)<glob>'...`,
// one per non-blank fence line (skills/flow-contracts/project-configuration.md).
//
// Exit codes:
//   0  both files written (parts: every part written)
//   2  cannot answer — a usage error, a worktree/sha pair that fails
//      panelValidateWorktree (stats/internal/guard/paneltouchedpaths.go), or a git
//      failure; nothing is written. `parts` also exits 2 when <cap> is not
//      a decimal integer, final-review.diff or its .touched is absent, or
//      the diff holds more than one `# worktree:` section

// writePanelDiff is `flow-guard write-panel-diff`: the comment above is the contract -- one panel diff file and its `.touched` list
// (the `[TOUCHED_FILES]` list), both sectioned per worktree and written
// atomically into <canonical-wt>/.superpowers/sdd/, both paths printed;
// exit 0 written, 2 cannot answer with nothing written.
func init() {
	Registry["write-panel-diff"] = writePanelDiff
}

const wpdUsage = "usage: flow-guard write-panel-diff final <canonical-wt> <wt> <merge-base> [<wt> <merge-base>...]\n" +
	"       flow-guard write-panel-diff late-fix <canonical-wt> <wt> <since-close-sha> [<wt> <since-close-sha>...]\n" +
	"       flow-guard write-panel-diff fix-round <N> <canonical-wt> <wt> <fix-base> [<wt> <fix-base>...]\n" +
	"       flow-guard write-panel-diff slot-delta <round> <slot> <canonical-wt> <wt> <merge-base> <held-sha|-> [...]\n" +
	"       flow-guard write-panel-diff parts <canonical-wt> <cap>\n"

var (
	wpdNumber = regexp.MustCompile(`^[0-9]+$`)
	wpdSlot   = regexp.MustCompile(`^[A-Za-z0-9._+-]+$`)
)

func writePanelDiff(args []string, env Env, stdout, stderr io.Writer) int {
	usage := func() int {
		fmt.Fprint(stderr, wpdUsage)
		Reason(env, "usage")
		return 2
	}
	if len(args) == 0 {
		return usage()
	}
	kind, args := args[0], args[1:]
	var file string
	width := 2 // arguments per worktree
	switch kind {
	case "final":
		file = "final-review.diff"
	case "late-fix":
		file = "late-fix.diff"
	case "fix-round":
		if len(args) == 0 || !wpdNumber.MatchString(args[0]) {
			return usage()
		}
		file, args = "fix-round-"+args[0]+".diff", args[1:]
	case "parts":
		if len(args) != 2 || !wpdNumber.MatchString(args[1]) {
			return usage()
		}
		return wpdParts(env, args[0], args[1], stdout, stderr)
	case "slot-delta":
		if len(args) < 2 || !wpdNumber.MatchString(args[0]) || !wpdSlot.MatchString(args[1]) || args[1] == "." || args[1] == ".." {
			return usage()
		}
		file, args, width = "slot-delta-"+args[0]+"-"+args[1]+".diff", args[2:], 3
	default:
		return usage()
	}
	if len(args) < 1+width || (len(args)-1)%width != 0 {
		return usage()
	}
	canon, rest := args[0], args[1:]
	if !isDir(smcAbs(env, canon)) {
		fmt.Fprintf(stderr, "write-panel-diff: %s is not a directory — cannot determine anything\n", canon)
		Reason(env, "not-a-directory")
		return 2
	}
	gitBin, ok := panelResolveGit(env, "write-panel-diff", stderr)
	if !ok {
		Reason(env, "git-unresolved")
		return 2
	}

	var diff, touched bytes.Buffer
	if kind == "final" {
		diff.WriteString("# final-review.diff — working tree vs merge-base, unstaged changes included\n")
	}
	for i := 0; i < len(rest); i += width {
		wt, base := rest[i], rest[i+1]
		if !panelValidateWorktree(env, "write-panel-diff", wt, base, gitBin, stderr) {
			Reason(env, "invalid-worktree")
			return 2
		}
		// The range each kind reads, as the prose recipes spelled it.
		var rng []string
		switch kind {
		case "final", "late-fix":
			rng = []string{base}
		case "fix-round":
			rng = []string{base + "..HEAD"}
		case "slot-delta":
			rng = []string{base}
			if held := rest[i+2]; held != "-" {
				if !panelValidateWorktree(env, "write-panel-diff", wt, held, gitBin, stderr) {
					Reason(env, "invalid-worktree")
					return 2
				}
				rng = []string{held, "HEAD"}
			}
		}
		excludes := wpdExcludes(smcAbs(env, wt))
		header := fmt.Sprintf("# worktree: %s — merge base %s\n", wt, base)
		for _, o := range []struct {
			buf   *bytes.Buffer
			flags []string
		}{{&diff, []string{"--irreversible-delete"}}, {&touched, []string{"--name-status"}}} {
			out, err := panelGit(env, gitBin, wt, append(append(append([]string{"diff"}, o.flags...), rng...), excludes...)...).Output()
			if err != nil {
				fmt.Fprintf(stderr, "write-panel-diff: git diff %v failed in %s\n", append(o.flags, rng...), wt)
				Reason(env, "git-diff-failed")
				return 2
			}
			o.buf.WriteString(header)
			o.buf.Write(out)
		}
	}

	dir := smcAbs(env, canon) + "/.superpowers/sdd"
	path := canon + "/.superpowers/sdd/" + file
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(stderr, "write-panel-diff: %v\n", err)
		Reason(env, "cannot-write-diff")
		return 2
	}
	for _, f := range []struct {
		name string
		body []byte
	}{{file, diff.Bytes()}, {file + ".touched", touched.Bytes()}} {
		if err := wpdWriteAtomic(filepath.Join(dir, f.name), f.body); err != nil {
			fmt.Fprintf(stderr, "write-panel-diff: %v\n", err)
			Reason(env, "cannot-write-diff")
			return 2
		}
	}
	// A final write opens a full pass 1, so a late-fix.diff left by an
	// earlier run is stale by construction: removed, so no slot can read it
	// in place of final-review.diff; so are the parts an earlier final was cut into.
	if kind == "final" {
		stale, _ := filepath.Glob(filepath.Join(dir, "final-review-part-*.diff*"))
		for _, f := range append([]string{"late-fix.diff", "late-fix.diff.touched"}, stale...) {
			if err := os.Remove(filepath.Join(dir, filepath.Base(f))); err != nil && !errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintf(stderr, "write-panel-diff: %v\n", err)
				Reason(env, "cannot-write-diff")
				return 2
			}
		}
	}
	fmt.Fprintf(stdout, "%s\n%s.touched\n", path, path)
	return 0
}

// wpdParts is the `parts` kind: see the contract header.
func wpdParts(env Env, canon, capArg string, stdout, stderr io.Writer) int {
	fail := func(reason, format string, a ...any) int {
		fmt.Fprintf(stderr, "write-panel-diff: "+format+"\n", a...)
		Reason(env, reason)
		return 2
	}
	limit, err := strconv.Atoi(capArg)
	if err != nil {
		return fail("usage", "cap %q is not a decimal integer", capArg)
	}
	dir := smcAbs(env, canon) + "/.superpowers/sdd"
	diff, err := os.ReadFile(filepath.Join(dir, "final-review.diff"))
	if err != nil {
		return fail("no-final-diff", "%v", err)
	}
	touched, err := os.ReadFile(filepath.Join(dir, "final-review.diff.touched"))
	if err != nil {
		return fail("no-final-diff", "%v", err)
	}
	lines := strings.SplitAfter(string(diff), "\n")
	var semantics, header string
	var blocks []string
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "# final-review.diff"):
			semantics = l
		case strings.HasPrefix(l, "# worktree: "):
			if header != "" {
				return fail("multi-worktree-refused", "final-review.diff holds more than one worktree section — parts cuts a single-worktree diff only")
			}
			header = l
		case strings.HasPrefix(l, "diff --git "):
			blocks = append(blocks, l)
		case len(blocks) > 0:
			blocks[len(blocks)-1] += l
		}
	}
	// Cut the blocks into parts by changed-line count.
	var parts [][]string
	size := 0
	for _, b := range blocks {
		n := wpdChangedLines(b)
		if len(parts) == 0 || (len(parts[len(parts)-1]) > 0 && size+n > limit) {
			parts, size = append(parts, nil), 0
		}
		parts[len(parts)-1] = append(parts[len(parts)-1], b)
		size += n
	}
	touchedLines := strings.SplitAfter(string(touched), "\n")
	for k, part := range parts {
		paths := map[string]bool{}
		for _, b := range part {
			first, _, _ := strings.Cut(b, "\n")
			for _, p := range wpdHeaderPaths(first) {
				paths[p] = true
			}
		}
		var tb strings.Builder
		for _, l := range touchedLines {
			if strings.HasPrefix(l, "# worktree: ") {
				tb.WriteString(l)
				continue
			}
			fields := strings.Split(strings.TrimRight(l, "\n"), "\t")
			if slices.ContainsFunc(fields[1:], func(f string) bool { return paths[f] }) {
				tb.WriteString(l)
			}
		}
		name := fmt.Sprintf("final-review-part-%d.diff", k+1)
		if err := wpdWriteAtomic(filepath.Join(dir, name), []byte(semantics+header+strings.Join(part, ""))); err != nil {
			return fail("cannot-write-diff", "%v", err)
		}
		if err := wpdWriteAtomic(filepath.Join(dir, name+".touched"), []byte(tb.String())); err != nil {
			return fail("cannot-write-diff", "%v", err)
		}
		fmt.Fprintf(stdout, "%s/.superpowers/sdd/%s\n", canon, name)
	}
	return 0
}

// wpdChangedLines counts a diff block's `+`/`-` lines, its `+++ `/`--- `
// headers excluded.
func wpdChangedLines(block string) int {
	n := 0
	for _, l := range strings.Split(block, "\n") {
		if (strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++ ")) || (strings.HasPrefix(l, "-") && !strings.HasPrefix(l, "--- ")) {
			n++
		}
	}
	return n
}

// wpdExcludes is the pathspec tail that leaves out of a worktree's panel
// diff every path a glob of its own .flow/project.md's `## panel diff
// exclude` fence matches -- `-- . ':(glob,exclude)<glob>'...` -- or nil when
// the worktree declares none.
func wpdExcludes(wt string) []string {
	globs := kwFenced(projectSection(filepath.Join(wt, ".flow", "project.md"), "panel diff exclude"))
	if len(globs) == 0 {
		return nil
	}
	tail := []string{"--", "."}
	for _, g := range globs {
		tail = append(tail, ":(glob,exclude)"+strings.TrimSpace(g))
	}
	return tail
}

// wpdWriteAtomic writes body to a temp file beside path and renames it over
// path, so a reader never sees a half-written diff.
func wpdWriteAtomic(path string, body []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	// CreateTemp's 0600 would differ from the recipe's `>` redirection.
	err = f.Chmod(0o644)
	if err == nil {
		_, err = f.Write(body)
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		_ = os.Remove(f.Name())
	}
	return err
}

// wpdHeaderPaths returns the paths a `diff --git a/X b/Y` header names. A path
// may itself contain " b/", so the split that gives X == Y wins; a rename
// (X != Y) falls back to the first " b/".
func wpdHeaderPaths(first string) []string {
	rest := strings.TrimPrefix(strings.TrimSuffix(first, "\n"), "diff --git a/")
	for i := 0; ; {
		j := strings.Index(rest[i:], " b/")
		if j < 0 {
			break
		}
		i += j
		if rest[:i] == rest[i+3:] {
			return []string{rest[:i]}
		}
		i++
	}
	if a, b, ok := strings.Cut(rest, " b/"); ok {
		return []string{a, b}
	}
	return nil
}
