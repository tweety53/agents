package guard

import (
	"os"
	"os/exec"
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
				var want, wantTouched string
				scratch := t.TempDir()
				for i := 0; i < len(tc.triples); i += 3 {
					sem := ""
					if i == 0 {
						sem = tc.sem
					}
					want += wpdHand(t, scratch+"/d", sem, "--irreversible-delete", tc.shape, tc.triples[i:i+3])
					wantTouched += wpdHand(t, scratch+"/t", "", "--name-status", tc.shape, tc.triples[i:i+3])
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
		{"a round that is not a number", []string{"fix-round", "x", w.path, w.path, w.mb}, Env{}, "usage: flow-guard write-panel-diff"},
		{"a worktree that is no directory", []string{"final", w.path, dir + "/absent", w.mb}, Env{}, "is not a directory"},
		{"a sha that does not resolve", []string{"final", w.path, w.path, "deadbeef"}, Env{}, "does not resolve"},
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
	if r := runGuard("check-panel-docs-only", []string{wt, mb}, env); r.rc != 0 {
		t.Errorf("check-panel-docs-only: exit %d, want 0\n%s", r.rc, r.out)
	}
}
