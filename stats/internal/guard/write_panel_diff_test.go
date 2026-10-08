package guard

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// wpdWT is one fixture worktree: mb its merge base, held a later sha a slot
// last read, and a tip, a staged and an unstaged edit on top of that.
type wpdWT struct{ path, mb, held string }

func wpdNewWT(t *testing.T, g *fxGit, path string) wpdWT {
	t.Helper()
	g.git("", "init", "-q", "-b", "work", path)
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
// line ("" for none), $3 the diff flags ("" or --name-status), $4 the range
// shape, then <wt> <header-sha> <from> triples.
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
					want += wpdHand(t, scratch+"/d", sem, "", shape, tc.triples[i:i+3])
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
		{"no arguments", nil, Env{}, "usage: write-panel-diff.sh"},
		{"unknown kind", []string{"whole", w.path, w.path, w.mb}, Env{}, "usage: write-panel-diff.sh"},
		{"a worktree without its sha", []string{"final", w.path, w.path}, Env{}, "usage: write-panel-diff.sh"},
		{"slot-delta without its held sha", []string{"slot-delta", "1", "primary", w.path, w.path, w.mb}, Env{}, "usage: write-panel-diff.sh"},
		{"a round that is not a number", []string{"fix-round", "x", w.path, w.path, w.mb}, Env{}, "usage: write-panel-diff.sh"},
		{"a slot naming a path", []string{"slot-delta", "1", "../x", w.path, w.path, w.mb, "-"}, Env{}, "usage: write-panel-diff.sh"},
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
