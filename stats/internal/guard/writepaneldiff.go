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
	"strings"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/gitops"
)

// The contract below moved here from the header of the deleted
// scripts/write-panel-diff.sh shim (KAN-1058).
//
// flow-guard write-panel-diff — write one review-panel diff and its touched list
// (skills/flow/review-panel.md).
//
// Usage:
//   flow-guard write-panel-diff final      <canonical-wt> <wt> <merge-base> [<wt> <merge-base>...]
//   flow-guard write-panel-diff late-fix   <canonical-wt> <wt> <fix-start-sha> [...]
//   flow-guard write-panel-diff fix-round <N> <canonical-wt> <wt> <fix-base> [...]
//
// Writes two files into <canonical-wt>/.superpowers/sdd/, each atomically
// (a temp file renamed over the target), and prints both paths, one per line:
//   final-review.diff | late-fix.diff | fix-round-<N>.diff
//   <that file>.touched — the [TOUCHED_FILES] list
// A `final` write also removes any late-fix.diff and its .touched: a full
// pass 1 makes them stale, and a slot must never read them in its place.
//
// Both are sectioned per worktree, in argument order, each section opened by
//   # worktree: <wt> — merge base <sha>
// naming the worktree's own second argument (its merge base, fix-start sha
// or fix base). final-review.diff alone opens with the semantics line
//   # final-review.diff — working tree vs merge-base, unstaged changes included
// Each diff section is that worktree's `git diff` over the kind's range, and
// each .touched section `git diff --name-status` over the same range:
//   final       git diff <merge-base>          (the working tree: staged and unstaged)
//   late-fix    git diff <fix-start-sha>       (the working tree)
//   fix-round   git diff <fix-base>..HEAD
// Every diff section (never the .touched one) runs with --irreversible-delete,
// so a deleted file shows as its header alone; its old body is
// `git show <sha>:<path>`, the sha its section's `# worktree:` header names.
//
// A worktree whose own .flow/project.md declares `## panel diff exclude`
// has every path one of that fence's globs matches left out of both its
// sections: each git diff above then ends `-- . ':(glob,exclude)<glob>'...`,
// one per non-blank fence line (skills/flow-contracts/project-configuration.md).
//
// Exit codes:
//   0  both files written
//   2  cannot answer — a usage error, a worktree/sha pair that fails
//      panelValidateWorktree (stats/internal/guard/paneltouchedpaths.go), or a git
//      failure; nothing is written

// writePanelDiff is `flow-guard write-panel-diff`: the comment above is the contract -- one panel diff file and its `.touched` list
// (the `[TOUCHED_FILES]` list), both sectioned per worktree and written
// atomically into <canonical-wt>/.superpowers/sdd/, both paths printed;
// exit 0 written, 2 cannot answer with nothing written.
func init() {
	Registry["write-panel-diff"] = writePanelDiff
}

const wpdUsage = "usage: flow-guard write-panel-diff final <canonical-wt> <wt> <merge-base> [<wt> <merge-base>...]\n" +
	"       flow-guard write-panel-diff late-fix <canonical-wt> <wt> <fix-start-sha> [<wt> <fix-start-sha>...]\n" +
	"       flow-guard write-panel-diff fix-round <N> <canonical-wt> <wt> <fix-base> [<wt> <fix-base>...]\n"

var wpdNumber = regexp.MustCompile(`^[0-9]+$`)

func writePanelDiff(args []string, env core.Env, stdout, stderr io.Writer) int {
	usage := func() int {
		fmt.Fprint(stderr, wpdUsage)
		core.Reason(env, "usage")
		return 2
	}
	if len(args) == 0 {
		return usage()
	}
	kind, args := args[0], args[1:]
	var file string
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
	default:
		return usage()
	}
	if len(args) < 3 || (len(args)-1)%2 != 0 {
		return usage()
	}
	canon, rest := args[0], args[1:]
	if !core.IsDir(core.Abs(env, canon)) {
		fmt.Fprintf(stderr, "write-panel-diff: %s is not a directory — cannot determine anything\n", canon)
		core.Reason(env, "not-a-directory")
		return 2
	}
	gitBin, ok := core.ResolveGit(env, "write-panel-diff", stderr)
	if !ok {
		core.Reason(env, "git-unresolved")
		return 2
	}

	var diff, touched bytes.Buffer
	if kind == "final" {
		diff.WriteString("# final-review.diff — working tree vs merge-base, unstaged changes included\n")
	}
	for i := 0; i < len(rest); i += 2 {
		wt, base := rest[i], rest[i+1]
		if !panelValidateWorktree(env, "write-panel-diff", wt, base, gitBin, stderr) {
			core.Reason(env, "invalid-worktree")
			return 2
		}
		// The range each kind reads, as the prose recipes spelled it.
		var rng []string
		switch kind {
		case "final", "late-fix":
			rng = []string{base}
		case "fix-round":
			rng = []string{base + "..HEAD"}
		}
		excludes := wpdExcludes(core.Abs(env, wt))
		header := fmt.Sprintf("# worktree: %s — merge base %s\n", wt, base)
		for _, o := range []struct {
			buf   *bytes.Buffer
			flags []string
		}{{&diff, []string{"--irreversible-delete"}}, {&touched, []string{"--name-status"}}} {
			out, err := panelGit(env, gitBin, wt, append(append(append([]string{"diff"}, o.flags...), rng...), excludes...)...).Output()
			if err != nil {
				fmt.Fprintf(stderr, "write-panel-diff: git diff %v failed in %s\n", append(o.flags, rng...), wt)
				core.Reason(env, "git-diff-failed")
				return 2
			}
			o.buf.WriteString(header)
			o.buf.Write(out)
		}
	}

	dir := core.Abs(env, canon) + "/.superpowers/sdd"
	path := canon + "/.superpowers/sdd/" + file
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(stderr, "write-panel-diff: %v\n", err)
		core.Reason(env, "cannot-write-diff")
		return 2
	}
	for _, f := range []struct {
		name string
		body []byte
	}{{file, diff.Bytes()}, {file + ".touched", touched.Bytes()}} {
		if err := wpdWriteAtomic(filepath.Join(dir, f.name), f.body); err != nil {
			fmt.Fprintf(stderr, "write-panel-diff: %v\n", err)
			core.Reason(env, "cannot-write-diff")
			return 2
		}
	}
	// A final write opens a full pass 1, so a late-fix.diff left by an
	// earlier run is stale by construction: removed, so no slot can read it
	// in place of final-review.diff.
	if kind == "final" {
		for _, f := range []string{"late-fix.diff", "late-fix.diff.touched"} {
			if err := os.Remove(filepath.Join(dir, f)); err != nil && !errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintf(stderr, "write-panel-diff: %v\n", err)
				core.Reason(env, "cannot-write-diff")
				return 2
			}
		}
	}
	fmt.Fprintf(stdout, "%s\n%s.touched\n", path, path)
	return 0
}

// wpdExcludes is the pathspec tail that leaves out of a worktree's panel
// diff every path a glob of its own .flow/project.md's `## panel diff
// exclude` fence matches -- `-- . ':(glob,exclude)<glob>'...` -- or nil when
// the worktree declares none.
func wpdExcludes(wt string) []string {
	globs := gitops.Fenced(core.ProjectSection(filepath.Join(wt, ".flow", "project.md"), "panel diff exclude"))
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
