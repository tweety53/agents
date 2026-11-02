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

// The contract below moved here from the header of the deleted
// scripts/resolve-visual-screenshots.sh shim (KAN-1058).
//
// flow-guard resolve-visual-screenshots — resolve a captured spec's PNGs, in a guard
// instead of prose.
//
// Usage: flow-guard resolve-visual-screenshots <project root> <capture spec basename>
//
// Replaces flow.visual-verify's step 9 resolution prose
// (skills/flow/visual-verify-verifier.md): "search recursively beneath
// `screenshots` for every PNG whose path contains the capture spec's own
// basename." Reading, still what a script cannot do, stays the agent's job —
// this guard only resolves WHICH paths to read.
//
// Prints one ABSOLUTE PNG path per line, nothing else, on a match — the
// stage feeds this straight into "read every captured PNG", so stdout here
// is data, not a report.
//
// THREE EXIT CODES:
//   0  at least one PNG matched; its path(s) are on stdout
//   1  zero matches — the case the stage must block on, per design.md. A
//      `screenshots` directory that does not exist YET is the identical
//      answer: `capture` creates it on its first run, so its absence is
//      `capture`'s own success path, not a guard failure.
//   2  cannot answer: usage error, the project root is not a directory,
//      `.flow/project.md` is missing/not a regular file/unreadable, the
//      heading scan failed to look, the section is declared twice, the
//      `screenshots` setting is absent or empty, or a declared `regression
//      checkout` does not resolve to an existing directory (so there is no
//      base to search relative to at all).
//
// `screenshots` RESOLVES RELATIVE TO THE REGRESSION CHECKOUT WHEN ONE IS
// DECLARED, OTHERWISE THE PROJECT ROOT — design.md's "The contract" table,
// cited rather than restated: this guard enforces it, it does not explain
// it. This guard does not otherwise validate the section's shape (that a
// declared `regression checkout`/`regression repo` pair is well formed, or
// that `regression repo` matches the checkout's real `origin`) —
// flow-guard check-visual-verification already does, and this guard's job is
// narrower: resolve one search root or say it cannot.
//
// THE FILTER READS THE WHOLE PATH, WITH NO SPECIAL-CASED EXCLUSION OF
// ANYTHING — a `screenshots` root wide enough to sweep in a `node_modules`
// tree still filters correctly, because "some path segment starts with the
// spec basename" is exactly as narrow as the stage's own step 9 says it is.
// Pinned rather than assumed: TestResolveVisualScreenshots
// (stats/internal/guard/resolve_visual_screenshots_test.go) plants a decoy
// PNG inside `node_modules` whose name deliberately DOES start with the spec basename
// and confirms it is found — the filter is not a `node_modules` carve-out
// that could hide a real match sitting there too (a project genuinely
// emitting PNGs into a vendored path is not this guard's problem to solve).
//
// THE MATCH IS ANCHORED AT A PATH-SEGMENT BOUNDARY, not a bare substring
// test — task 16's own panel-demonstrated defect. `visual-baseline.spec.ts`
// CONTAINS `baseline.spec.ts` as a substring, so a naive `case "$png" in
// *"$SPEC_BASENAME"*)` filter resolving `baseline.spec.ts` also returned
// `visual-baseline.spec.ts-snapshots/`'s PNGs — a different spec's own
// baseline, sitting under a directory that merely happens to end with the
// queried name. Splitting the path on `/` and testing whether any SEGMENT
// starts with `$SPEC_BASENAME` closes that: `visual-baseline.spec.ts-…`
// starts with `visual-`, not `baseline.spec.ts`, so it is correctly
// rejected, while `baseline.spec.ts-snapshots` and `baseline.spec.ts.png`
// (the node_modules decoy above) both still match, because each IS the
// start of its own segment. Only the START is anchored — the string after
// the spec basename within a segment is unconstrained on purpose, which is
// what lets both the real `-snapshots` suffix and a literal `.png` extension
// still match.
//
// THE INPUT IS ATTACKER-INFLUENCED, exactly as flow-guard check-visual-verification's
// own header states: `.flow/project.md` is tracked and editable in any pull
// request. Every interpolated cell that reaches an exit-2 message passes
// through sanitizeDisplay first (stats/internal/guard/visualsection.go), for
// the identical forged-verdict reason that function's comment documents. The PNG paths this guard
// prints on success are filesystem paths this guard discovered by walking a
// real directory, never text lifted out of the config file, so they carry no
// such risk and are printed unescaped.
//
// THE HEADING SCAN AND TABLE PARSER ARE flow-guard check-visual-verification's OWN
// CONVENTIONS, reused rather than reinvented — the Go twins in
// stats/internal/guard/visualsection.go that every visual guard shares.
//
// A LEADING UTF-8 BOM IS STRIPPED BEFORE EITHER HEADING READ OR THE TABLE
// PARSE, via stripBOM (stats/internal/guard/visualsection.go) — see its
// comment for why a BOM would otherwise make this guard read a present
// `## visual verification` heading as absent and exit 2.
//
// The body is Go: stats/internal/guard/resolvevisualscreenshots.go. Matches
// print in byte order, whatever the caller's locale.

// resolveVisualScreenshots is `flow-guard resolve-visual-screenshots`: the comment above is the contract -- one absolute PNG path per line
// on stdout and exit 0, exit 1 on zero matches, exit 2 when it cannot answer.
// The reasoning for each step, moved here from the bash body it replaced
// (3915fbc0), sits beside the code it explains.
func init() {
	Registry["resolve-visual-screenshots"] = resolveVisualScreenshots
}

const rvsPrefix = "resolve-visual-screenshots: "

func resolveVisualScreenshots(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprint(stderr, rvsPrefix+"usage: flow-guard resolve-visual-screenshots <project root> <capture spec basename>\n")
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
// `.worktrees/` held other changes' checkouts of the same files before the
// sibling `<repo>-worktrees/` layout, and still does for one not yet migrated:
// sweeping them in returns one PNG name per checkout, and a zip built from the
// list refuses the repeats.
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
