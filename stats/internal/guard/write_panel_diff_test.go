package guard

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// wpdWT is one fixture worktree: mb its merge base, held a later sha a slot
// last read, and a tip, a staged and an unstaged edit on top of that.
type wpdWT struct{ path, mb, held string }

func wpdNewWT(t *testing.T, g *fxGit, path string) wpdWT {
	t.Helper()
	g.copyShape(path, repoShape{Branch: "work"})
	g.write(path+"/a.txt", "base")
	g.write(path+"/gone.txt", "gone")
	g.git(path, "add", ".")
	g.git(path, "commit", "-qm", "base")
	w := wpdWT{path: path, mb: g.git(path, "rev-parse", "HEAD")}
	g.write(path+"/a.txt", "held")
	g.write(path+"/b.txt", "new")
	g.git(path, "add", ".")
	g.git(path, "commit", "-qm", "held")
	w.held = g.git(path, "rev-parse", "HEAD")
	g.git(path, "rm", "-q", "gone.txt")
	g.git(path, "commit", "-qm", "tip")
	g.write(path+"/c.txt", "staged")
	g.git(path, "add", "c.txt")
	g.write(path+"/a.txt", "unstaged")
	return w
}

// wpdRecipe is the prose's hand recipe: `printf` headers and `git -C <wt>
// diff` sections appended to one file, run by bash under the process
// environment, as the parent ran it. $1 the output file, $2 the semantics
// line ("" for none), $3 the diff flags (--irreversible-delete or
// --name-status), $4 the range shape, then <wt> <header-sha> <from> triples.
const wpdRecipe = `set -e
out=$1 sem=$2 flags=$3 shape=$4; shift 4
: > "$out"
[ -z "$sem" ] || printf '%s\n' "$sem" >> "$out"
while [ $# -gt 0 ]; do
  printf '# worktree: %s — merge base %s\n' "$1" "$2" >> "$out"
  case $shape in
    to-tree) git -C "$1" diff $flags "$3" >> "$out" ;;
    to-head) git -C "$1" diff $flags "$3" HEAD >> "$out" ;;
    dotdot) git -C "$1" diff $flags "$3"..HEAD >> "$out" ;;
  esac
  shift 3
done`

func wpdHand(t *testing.T, out, sem, flags, shape string, triples []string) string {
	t.Helper()
	cmd := exec.Command("bash", append([]string{"-c", wpdRecipe, "recipe", out, sem, flags, shape}, triples...)...)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("hand recipe: %v\n%s", err, b)
	}
	return rbRead(out)
}

func TestWritePanelDiffParity(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	const finalSem = "# final-review.diff — working tree vs merge-base, unstaged changes included"
	for _, n := range []int{1, 2} {
		dir := t.TempDir()
		var g fxGit
		var wts []wpdWT
		for i := 0; i < n; i++ {
			wts = append(wts, wpdNewWT(t, &g, dir+"/wt"+string(rune('1'+i))))
		}
		if g.err != nil {
			t.Fatal(g.err)
		}
		canon := wts[0].path
		// pairs builds the guard's per-worktree arguments; triples the
		// recipe's <wt> <header-sha> <from>.
		pairs := func(f func(w wpdWT) []string) []string {
			var a []string
			for _, w := range wts {
				a = append(a, f(w)...)
			}
			return a
		}
		for _, tc := range []struct {
			label, file, sem, shape string
			args                    []string
			triples                 []string
		}{
			{"final", "final-review.diff", finalSem, "to-tree",
				append([]string{"final", canon}, pairs(func(w wpdWT) []string { return []string{w.path, w.mb} })...),
				pairs(func(w wpdWT) []string { return []string{w.path, w.mb, w.mb} })},
			{"slot-delta held", "slot-delta-2-primary.diff", "", "to-head",
				append([]string{"slot-delta", "2", "primary", canon}, pairs(func(w wpdWT) []string { return []string{w.path, w.mb, w.held} })...),
				pairs(func(w wpdWT) []string { return []string{w.path, w.mb, w.held} })},
			{"slot-delta unheld", "slot-delta-2-principles.diff", "", "to-tree",
				append([]string{"slot-delta", "2", "principles", canon}, pairs(func(w wpdWT) []string { return []string{w.path, w.mb, "-"} })...),
				pairs(func(w wpdWT) []string { return []string{w.path, w.mb, w.mb} })},
			{"late-fix", "late-fix.diff", "", "to-tree",
				append([]string{"late-fix", canon}, pairs(func(w wpdWT) []string { return []string{w.path, w.held} })...),
				pairs(func(w wpdWT) []string { return []string{w.path, w.held, w.held} })},
			{"fix-round", "fix-round-3.diff", "", "dotdot",
				append([]string{"fix-round", "3", canon}, pairs(func(w wpdWT) []string { return []string{w.path, w.held} })...),
				pairs(func(w wpdWT) []string { return []string{w.path, w.held, w.held} })},
		} {
			t.Run(tc.label+" over "+string(rune('0'+n))+" worktree(s)", func(t *testing.T) {
				// Subtests share the worktrees and write into canon, so
				// they run in order.
				if n == 2 && tc.label == "slot-delta held" {
					// One worktree in which the slot holds no sha falls
					// back to its merge-base section.
					tc.args = append([]string{}, tc.args...)
					tc.args[len(tc.args)-1] = "-"
					tc.triples = append([]string{}, tc.triples...)
					tc.triples[len(tc.triples)-1] = wts[1].mb
				}
				// A delta's unheld section is `git diff <mb>` to the tree;
				// its held one `git diff <held> HEAD`: the recipe runs per
				// section, so a mixed delta is built section by section.
				var want, wantTouched string
				scratch := t.TempDir()
				for i := 0; i < len(tc.triples); i += 3 {
					shape, sem := tc.shape, ""
					if i == 0 {
						sem = tc.sem
					}
					if strings.HasPrefix(tc.label, "slot-delta") && tc.triples[i+2] == tc.triples[i+1] {
						shape = "to-tree"
					}
					want += wpdHand(t, scratch+"/d", sem, "--irreversible-delete", shape, tc.triples[i:i+3])
					wantTouched += wpdHand(t, scratch+"/t", "", "--name-status", shape, tc.triples[i:i+3])
				}
				r := runGuard("write-panel-diff", tc.args, Env{Dir: dir, Getenv: os.Getenv})
				if r.rc != 0 {
					t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
				}
				path := canon + "/.superpowers/sdd/" + tc.file
				if r.stdout != path+"\n"+path+".touched\n" {
					t.Errorf("stdout %q, want both paths", r.stdout)
				}
				if got := rbRead(path); got != want {
					t.Errorf("%s differs from the hand recipe\n got %q\nwant %q", tc.file, got, want)
				}
				if got := rbRead(path + ".touched"); got != wantTouched {
					t.Errorf("%s.touched differs from --name-status sections\n got %q\nwant %q", tc.file, got, wantTouched)
				}
			})
		}
	}
}

func TestWritePanelDiffFinalRemovesStaleLateFix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	w := wpdNewWT(t, &g, dir+"/wt1")
	if g.err != nil {
		t.Fatal(g.err)
	}
	for _, args := range [][]string{{"late-fix", w.path, w.path, w.held}, {"final", w.path, w.path, w.mb}} {
		if r := runGuard("write-panel-diff", args, Env{Dir: dir, Getenv: os.Getenv}); r.rc != 0 {
			t.Fatalf("%s: exit %d\n%s", args[0], r.rc, r.out)
		}
	}
	for _, f := range []string{"late-fix.diff", "late-fix.diff.touched"} {
		if _, err := os.Stat(w.path + "/.superpowers/sdd/" + f); !os.IsNotExist(err) {
			t.Errorf("%s survived a final write (err %v)", f, err)
		}
	}
}

func TestWritePanelDiffRefusals(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	w := wpdNewWT(t, &g, dir+"/wt")
	if g.err != nil {
		t.Fatal(g.err)
	}
	stub := t.TempDir()
	stubGit(t, stub, `has --name-status "$@"`, "name-status refused")
	for _, tc := range []struct {
		label string
		args  []string
		env   Env
		err   string
	}{
		{"no arguments", nil, Env{}, "usage: flow-guard write-panel-diff"},
		{"unknown kind", []string{"whole", w.path, w.path, w.mb}, Env{}, "usage: flow-guard write-panel-diff"},
		{"a worktree without its sha", []string{"final", w.path, w.path}, Env{}, "usage: flow-guard write-panel-diff"},
		{"slot-delta without its held sha", []string{"slot-delta", "1", "primary", w.path, w.path, w.mb}, Env{}, "usage: flow-guard write-panel-diff"},
		{"a round that is not a number", []string{"fix-round", "x", w.path, w.path, w.mb}, Env{}, "usage: flow-guard write-panel-diff"},
		{"a slot naming a path", []string{"slot-delta", "1", "../x", w.path, w.path, w.mb, "-"}, Env{}, "usage: flow-guard write-panel-diff"},
		{"a worktree that is no directory", []string{"final", w.path, dir + "/absent", w.mb}, Env{}, "is not a directory"},
		{"a sha that does not resolve", []string{"final", w.path, w.path, "deadbeef"}, Env{}, "does not resolve"},
		{"a held sha that does not resolve", []string{"slot-delta", "1", "primary", w.path, w.path, w.mb, "deadbeef"}, Env{}, "does not resolve"},
		{"a git failure", []string{"final", w.path, w.path, w.mb}, Env{Getenv: pathEnv(stub)}, "write-panel-diff: git diff"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			env := tc.env
			env.Dir = dir
			if env.Getenv == nil {
				env.Getenv = os.Getenv
			}
			r := runGuard("write-panel-diff", tc.args, env)
			if r.rc != 2 || !strings.Contains(r.err, tc.err) || r.stdout != "" {
				t.Errorf("exit %d stdout %q stderr %q; want 2, nothing on stdout, stderr naming %q", r.rc, r.stdout, r.err, tc.err)
			}
			if entries, _ := os.ReadDir(w.path + "/.superpowers/sdd"); len(entries) != 0 {
				t.Errorf("wrote %d file(s) on a refusal", len(entries))
			}
		})
	}
}

// A deleted file shows in a panel diff as its header alone: its body is
// never sent to a reviewer, and the .touched list still names it.
func TestWritePanelDiffOmitsDeletedBodies(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	w := wpdNewWT(t, &g, dir+"/wt")
	if g.err != nil {
		t.Fatal(g.err)
	}
	r := runGuard("write-panel-diff", []string{"final", w.path, w.path, w.mb}, Env{Dir: dir, Getenv: os.Getenv})
	if r.rc != 0 {
		t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
	}
	diff := rbRead(w.path + "/.superpowers/sdd/final-review.diff")
	if !strings.Contains(diff, "diff --git a/gone.txt b/gone.txt") || !strings.Contains(diff, "deleted file mode") {
		t.Errorf("diff lacks gone.txt's deletion header:\n%s", diff)
	}
	for _, l := range strings.Split(diff, "\n") {
		if l == "-gone" {
			t.Errorf("diff carries the deleted body line -gone:\n%s", diff)
		}
	}
	if touched := rbRead(w.path + "/.superpowers/sdd/final-review.diff.touched"); !strings.Contains(touched, "D\tgone.txt") {
		t.Errorf(".touched lacks D\\tgone.txt:\n%s", touched)
	}
}

// A worktree's `## panel diff exclude` globs leave matching paths out of
// both the diff and its .touched list; every other path stays (kan-1012: the
// vendored handoff_* directories were a third of final-review.diff).
func TestWritePanelDiffHonoursProjectExcludes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	w := wpdNewWT(t, &g, dir+"/wt")
	for _, d := range []string{"/.flow", "/handoff_x"} {
		if err := os.MkdirAll(w.path+d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	g.write(w.path+"/.flow/project.md", "# p\n\n## panel diff exclude\n\n```text\nhandoff_*/**\n```\n")
	g.write(w.path+"/handoff_x/screen.html", "vendored")
	g.git(w.path, "add", ".flow/project.md", "handoff_x/screen.html")
	if g.err != nil {
		t.Fatal(g.err)
	}
	r := runGuard("write-panel-diff", []string{"final", w.path, w.path, w.mb}, Env{Dir: dir, Getenv: os.Getenv})
	if r.rc != 0 {
		t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
	}
	for _, f := range []string{"final-review.diff", "final-review.diff.touched"} {
		body := rbRead(w.path + "/.superpowers/sdd/" + f)
		if strings.Contains(body, "handoff_x/") {
			t.Errorf("%s carries the excluded handoff_x/:\n%s", f, body)
		}
		if !strings.Contains(body, "b.txt") || !strings.Contains(body, ".flow/project.md") {
			t.Errorf("%s lost a path no glob matches:\n%s", f, body)
		}
	}
}

// The size cap and the docs-only check leave a worktree's `## panel diff
// exclude` paths out as the panel diff does (kan-1012 self-review): a
// 40-line vendored handoff_x/screen.html beside a Markdown-only change
// counts toward neither.
func TestPanelGuardsHonourProjectExcludes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	wt := dir + "/wt"
	g.copyShape(wt, repoShape{Branch: "work"})
	g.write(wt+"/README.md", "base")
	g.git(wt, "add", ".")
	g.git(wt, "commit", "-qm", "base")
	mb := g.git(wt, "rev-parse", "HEAD")
	for _, d := range []string{"/.flow", "/handoff_x"} {
		mkdir(t, wt+d)
	}
	g.write(wt+"/.flow/project.md", "# p\n\n## panel diff exclude\n\n```text\nhandoff_*/**\n```\n")
	g.write(wt+"/handoff_x/screen.html", strings.Repeat("vendored\n", 40))
	g.git(wt, "add", ".")
	g.git(wt, "commit", "-qm", "change")
	if g.err != nil {
		t.Fatal(g.err)
	}
	env := Env{Dir: dir, Getenv: os.Getenv, LookupEnv: hermeticGitLookup}
	if r := runGuard("check-panel-diff-size", []string{wt, mb, "10"}, env); r.rc != 0 || r.stdout != "8\nunder cap\n" {
		t.Errorf("check-panel-diff-size: exit %d, stdout %q; want 0, 8 under cap\n%s", r.rc, r.stdout, r.out)
	}
	if r := runGuard("check-panel-docs-only", []string{wt, mb}, env); r.rc != 0 {
		t.Errorf("check-panel-docs-only: exit %d, want 0\n%s", r.rc, r.out)
	}
}

// TestWritePanelDiffWritesParts pins `parts` (kan-1090 self-review, F9): an
// over-cap single-worktree final-review.diff is cut at file boundaries into
// parts of at most <cap> changed lines, a file over the cap a part of its own.
func TestWritePanelDiffWritesParts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	wt := dir + "/wt"
	g.copyShape(wt, repoShape{Branch: "work"})
	g.write(wt+"/keep.txt", "keep")
	g.git(wt, "add", ".")
	g.git(wt, "commit", "-qm", "base")
	base := g.git(wt, "rev-parse", "HEAD")
	g.write(wt+"/a.txt", "1\n2\n3")
	g.write(wt+"/b.txt", "1\n2\n3\n4")
	g.write(wt+"/c.txt", "1\n2")
	g.git(wt, "add", ".")
	if g.err != nil {
		t.Fatal(g.err)
	}
	env := Env{Dir: dir, Getenv: os.Getenv}
	sdd := wt + "/.superpowers/sdd/"
	run := func(args ...string) guardResult {
		t.Helper()
		return runGuard("write-panel-diff", args, env)
	}
	// Refusals before any final-review.diff exists.
	for _, args := range [][]string{{"parts", wt, "x"}, {"parts", wt, "6"}} {
		if r := run(args...); r.rc != 2 || r.stdout != "" {
			t.Errorf("%v: exit %d stdout %q; want 2 and nothing", args, r.rc, r.stdout)
		}
	}
	if r := run("final", wt, wt, base); r.rc != 0 {
		t.Fatalf("final: exit %d\n%s", r.rc, r.out)
	}
	header := "# worktree: " + wt + " — merge base " + base + "\n"
	const sem = "# final-review.diff — working tree vs merge-base, unstaged changes included\n"
	r := run("parts", wt, "6")
	if want := sdd + "final-review-part-1.diff\n" + sdd + "final-review-part-2.diff\n"; r.rc != 0 || r.stdout != want {
		t.Fatalf("parts 6: exit %d stdout %q; want 0 and %q\n%s", r.rc, r.stdout, want, r.err)
	}
	for k, want := range [][]string{{"a.txt"}, {"b.txt", "c.txt"}} {
		name := sdd + "final-review-part-" + string(rune('1'+k)) + ".diff"
		body, touched := rbRead(name), rbRead(name+".touched")
		if !strings.HasPrefix(body, sem+header) || !strings.HasPrefix(touched, header) {
			t.Errorf("part %d lacks its headers:\n%s\n%s", k+1, body, touched)
		}
		for _, f := range []string{"a.txt", "b.txt", "c.txt"} {
			in := slices.Contains(want, f)
			if strings.Contains(body, "diff --git a/"+f) != in || strings.Contains(touched, "A\t"+f) != in {
				t.Errorf("part %d: %s present=%v, want %v\n%s\n%s", k+1, f, !in, in, body, touched)
			}
		}
	}
	if r := run("parts", wt, "2"); r.rc != 0 || strings.Count(r.stdout, "\n") != 3 {
		t.Fatalf("parts 2: exit %d stdout %q; want three parts", r.rc, r.stdout)
	}
	if body := rbRead(sdd + "final-review-part-2.diff"); !strings.Contains(body, "b/b.txt") || strings.Contains(body, "c.txt") {
		t.Errorf("parts 2: b.txt is not alone in part 2:\n%s", body)
	}
	if r := run("final", wt, wt, base); r.rc != 0 {
		t.Fatalf("final: exit %d", r.rc)
	}
	if stale, _ := filepath.Glob(sdd + "final-review-part-*"); len(stale) != 0 {
		t.Errorf("a final write left %v", stale)
	}
	if r := run("final", wt, wt, base, wt, base); r.rc != 0 {
		t.Fatalf("final over two sections: exit %d", r.rc)
	}
	if r := run("parts", wt, "6"); r.rc != 2 || !strings.Contains(r.err, "more than one worktree section") {
		t.Errorf("two sections: exit %d stderr %q; want 2", r.rc, r.err)
	}
}

// TestWpdHeaderPaths pins that a path containing " b/" is not cut at it.
func TestWpdHeaderPaths(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		in   string
		want []string
	}{
		{"diff --git a/x.txt b/x.txt", []string{"x.txt"}},
		{"diff --git a/d b/e.txt b/d b/e.txt", []string{"d b/e.txt"}},
		{"diff --git a/old.txt b/new.txt", []string{"old.txt", "new.txt"}},
	} {
		if got := wpdHeaderPaths(c.in); !slices.Equal(got, c.want) {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
}
