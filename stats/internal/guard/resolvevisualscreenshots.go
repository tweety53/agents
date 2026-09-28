package guard

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

// resolveVisualScreenshots is scripts/resolve-visual-screenshots.sh: that
// script's header comment is the contract -- one absolute PNG path per line
// on stdout and exit 0, exit 1 on zero matches, exit 2 when it cannot answer.
// The reasoning for each step, moved here from the bash body it replaced
// (3915fbc0), sits beside the code it explains.
func init() {
	Registry["resolve-visual-screenshots"] = resolveVisualScreenshots
}

const rvsPrefix = "resolve-visual-screenshots: "

func resolveVisualScreenshots(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprint(stderr, rvsPrefix+"usage: resolve-visual-screenshots.sh <project root> <capture spec basename>\n")
		return 2
	}
	root, spec := args[0], args[1]
	if spec == "" {
		fmt.Fprint(stderr, rvsPrefix+"the capture spec basename argument is empty\n")
		return 2
	}
	if !isDir(rvsAt(env, root)) {
		fmt.Fprintf(stderr, rvsPrefix+"%s is not a directory — cannot resolve what it declares\n", root)
		return 2
	}
	rootAbs := rvsLogical(env, root)
	cfg := rootAbs + "/.flow/project.md"

	fi, err := os.Stat(cfg)
	if _, lerr := os.Lstat(cfg); err != nil && lerr != nil {
		fmt.Fprintf(stderr, rvsPrefix+"%s has no .flow/project.md — cannot resolve where screenshots live\n", rootAbs)
		return 2
	}
	if err != nil || !fi.Mode().IsRegular() {
		fmt.Fprintf(stderr, rvsPrefix+"%s is not a regular file — cannot resolve what it declares\n", cfg)
		return 2
	}
	if syscall.Access(cfg, 4) != nil {
		fmt.Fprintf(stderr, rvsPrefix+"%s exists but is not readable — cannot resolve what it declares\n", cfg)
		return 2
	}
	data, err := os.ReadFile(cfg)
	if err != nil {
		// The bash's grep exiting 2 here: a failure to look, not an absence.
		fmt.Fprintf(stderr, rvsPrefix+"grep exited 2 while looking for the '## visual verification' heading in %s — that is a failure to look, not an absence\n", cfg)
		return 2
	}
	// A leading UTF-8 BOM is stripped before either heading read or the table
	// parse (stripBOM, inside vvHeadingCount and vvSectionLines): a BOM would
	// otherwise make a present `## visual verification` heading read as
	// absent.
	text := string(data)
	switch n := vvHeadingCount(text); {
	case n == 0:
		fmt.Fprintf(stderr, rvsPrefix+"%s declares no '## visual verification' section — cannot resolve where screenshots live\n", cfg)
		return 2
	case n > 1:
		fmt.Fprintf(stderr, rvsPrefix+"%s declares %d '## visual verification' sections — a second declaration is ambiguous, so neither was read\n", cfg, n)
		return 2
	}

	// `screenshots` and `regression checkout`, whichever are present, the
	// first row of each.
	var screenshots, checkout string
	var haveShots, haveCheckout bool
	for _, l := range vvSectionLines(text) {
		cells := vtSplitCells(l.Text)
		if len(cells) != 2 {
			continue
		}
		switch vtFoldCell(cells[0]) {
		case "screenshots":
			if !haveShots {
				screenshots, haveShots = vtTrimCell(cells[1]), true
			}
		case "regression checkout":
			if !haveCheckout {
				checkout, haveCheckout = vtTrimCell(cells[1]), true
			}
		}
	}

	// Every interpolated cell that reaches an exit-2 message passes through
	// sanitizeDisplay: `.flow/project.md` is attacker-influenced, and a cell
	// holding an escape sequence would otherwise write it to the terminal.
	if screenshots == "" {
		fmt.Fprint(stderr, sanitizeDisplay(fmt.Sprintf(rvsPrefix+"`screenshots` is absent or empty in %s — cannot resolve where captures land\n", cfg)))
		return 2
	}

	base := rootAbs
	if haveCheckout {
		// Relative to the caller's cwd, as the bash's `[ -d ]` and `cd` read it.
		if !isDir(rvsAt(env, checkout)) {
			fmt.Fprint(stderr, sanitizeDisplay(fmt.Sprintf(rvsPrefix+"`regression checkout` names `%s`, which is not an existing directory — cannot resolve `screenshots` relative to it\n", checkout)))
			return 2
		}
		base = rvsLogical(env, checkout)
		// The declared checkout is a main checkout (project-configuration.md,
		// "Roots in `## apps` are main checkouts"): while a worktree of it sits
		// on the project root's own branch, that worktree holds the change's
		// captures, and the main checkout holds only what already landed.
		git := envGit(env)
		if branch, _ := capture(git("-C", rootAbs, "branch", "--show-current")); branch != "" {
			if wt := rvsWorktreeOn(git("-C", base, "worktree", "list", "--porcelain"), branch); wt != "" {
				base = wt
			}
		}
	}

	searchRoot := base + "/" + screenshots
	var matches []string
	// A `screenshots` directory that does not exist yet is zero matches:
	// `capture` creates it on its first run.
	if isDir(searchRoot) {
		matches = rvsWalk(rvsLogical(env, searchRoot), spec, stderr)
	}
	// Bytewise (bytewise-screenshot-sort): the bash `sort` used the caller's
	// collation, and neither consumer reads the order.
	sort.Strings(matches)
	if len(matches) == 0 {
		fmt.Fprintf(stderr, rvsPrefix+"zero PNGs under %s matched `%s`\n", searchRoot, spec)
		return 1
	}
	for _, m := range matches {
		fmt.Fprintln(stdout, m)
	}
	return 0
}

// rvsAt is p as a bash `[ -d "$p" ]` run in env.Dir reads it. An empty p
// stays empty — `[ -d "" ]` is false — rather than naming env.Dir itself.
func rvsAt(env Env, p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return env.Dir + "/" + p
}

// rvsLogical is `cd "$p" && pwd`'s spelling: the logical path, joined onto
// the caller's $PWD (env.Dir) when relative, `..` taken lexically.
func rvsLogical(env Env, p string) string {
	return filepath.Clean(rvsAt(env, p))
}

// rvsWorktreeOn is the path of the worktree `git worktree list --porcelain`
// lists on branch, or "". A git that fails prints nothing, and nothing
// matches.
func rvsWorktreeOn(cmd *exec.Cmd, branch string) string {
	out, _ := cmd.Output()
	var w string
	for _, l := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(l, "worktree ") {
			w = l[len("worktree "):]
		}
		if l == "branch refs/heads/"+branch {
			return w
		}
	}
	return ""
}

// rvsWalk is `find "$root" -type d -name .worktrees -prune -o -type f -name
// '*.png' -print`, filtered: regular files only, symlinks neither followed
// nor matched (find without -L; WalkDir follows none, the root included).
//
// `.worktrees/` holds other changes' checkouts of the same files: sweeping
// them in returns one PNG name per checkout, and a zip built from the list
// refuses the repeats.
//
// THE MATCH IS ANCHORED AT A PATH-SEGMENT BOUNDARY: a path is kept when any
// `/`-separated segment of it starts with spec, so `visual-baseline.spec.ts-…`
// does not match `baseline.spec.ts` while `baseline.spec.ts-snapshots` and
// `baseline.spec.ts.png` both do. The whole path is read with no
// special-cased exclusion, a `node_modules` tree included.
//
// A directory it cannot read is named on stderr and skipped, as find did; the
// walk goes on.
func rvsWalk(root, spec string, stderr io.Writer) []string {
	var matches []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(stderr, rvsPrefix+"%v\n", err)
			return nil
		}
		if d.IsDir() && d.Name() == ".worktrees" {
			return filepath.SkipDir
		}
		if !d.Type().IsRegular() || !strings.HasSuffix(d.Name(), ".png") {
			return nil
		}
		for _, seg := range strings.Split(p, "/") {
			if strings.HasPrefix(seg, spec) {
				matches = append(matches, p)
				break
			}
		}
		return nil
	})
	return matches
}
