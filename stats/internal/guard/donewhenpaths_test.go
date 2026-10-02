package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dwRepo is a fresh repository with each of paths created (with parents) and
// staged, so `git ls-files` already reports them: staged is tracked, which is
// the index the guard judges against at archive time.
func dwRepo(t *testing.T, paths ...string) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var g fxGit
	g.git(dir, "init", "-q")
	g.git(dir, "commit", "-q", "--allow-empty", "-m", "base")
	for _, p := range paths {
		if g.err == nil {
			g.err = os.MkdirAll(filepath.Dir(dir+"/"+p), 0o755)
		}
		g.write(dir+"/"+p, "x")
		g.git(dir, "add", p)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	return dir
}

// dwFile stages one tracked markdown file whose body is body.
func dwFile(t *testing.T, dir, file, body string) {
	t.Helper()
	var g fxGit
	if g.err = os.MkdirAll(filepath.Dir(dir+"/"+file), 0o755); g.err != nil {
		t.Fatal(g.err)
	}
	g.write(dir+"/"+file, body)
	g.git(dir, "add", file)
	if g.err != nil {
		t.Fatal(g.err)
	}
}

func runDoneWhen(t *testing.T, root string, args ...string) (int, string, string) {
	t.Helper()
	env := Env{Getenv: os.Getenv}
	if len(args) == 0 {
		args = []string{root}
	}
	var out, errb strings.Builder
	code := checkDoneWhenPaths(args, env, &out, &errb)
	return code, out.String(), errb.String()
}

func TestCheckDoneWhenPaths(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		repo []string         // tracked, staged paths
		md   func(dir string) // the tracked markdown the case needs, staged
		rc   int
		want string // REPO stands for the repository's physical path
	}{
		{"case 1: no Done-when section anywhere is DONE-WHEN-OK",
			[]string{"docs/readme.md"},
			func(dir string) { dwFile(t, dir, "docs/readme.md", "plain prose, no heading\n") },
			0, "DONE-WHEN-OK: REPO\n"},
		{"case 2: a tracked named path is DONE-WHEN-OK",
			[]string{"docs/tickets/kan-x.md", "shots/27.png"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md", "## Done when\n\n`shots/27.png` re-baselined.\n")
			},
			0, "DONE-WHEN-OK: REPO\n"},
		{"case 3: an untracked named path is named verbatim with its file, once per pair",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md", "## Done when\n\n`shots/27.png` re-baselined, shots/27.png again.\n")
			},
			1, "DONE-WHEN-PATH: shots/27.png — docs/tickets/kan-x.md\nDONE-WHEN-VIOLATION: REPO — 1\n"},
		{"case 4: snapshot numbers never read as paths — kan-743's wording",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md", "## Done when\nSnapshots 27/28/29 re-baselined.\n")
			},
			0, "DONE-WHEN-OK: REPO\n"},
		{"case 5: globs, URLs and directories are not checked",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md",
					"## Done when\n- shots/*.png captured\n- evidence at https://x.test/shots/27.png\n- browsed shots/\n")
			},
			0, "DONE-WHEN-OK: REPO\n"},
		{"case 6: markdown links, autolinks and backticks yield the path",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md",
					"## Done when\n- [e](shots/27.png)\n- <shots/28.png>\n- `shots/29.png`\n")
			},
			1, "DONE-WHEN-PATH: shots/27.png — docs/tickets/kan-x.md\nDONE-WHEN-PATH: shots/28.png — docs/tickets/kan-x.md\nDONE-WHEN-PATH: shots/29.png — docs/tickets/kan-x.md\nDONE-WHEN-VIOLATION: REPO — 3\n"},
		{"case 7: a ./ prefix folds; membership stays exact",
			[]string{"docs/tickets/kan-x.md", "shots/27.png"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md", "## Done when\n- ./shots/27.png re-baselined\n")
			},
			0, "DONE-WHEN-OK: REPO\n"},
		{"case 8: the section closes at the next heading",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md",
					"## Done when\n\n`shots/27.png` re-baselined.\n\n## Later\n\n`shots/28.png` mentioned after.\n")
			},
			1, "DONE-WHEN-PATH: shots/27.png — docs/tickets/kan-x.md\nDONE-WHEN-VIOLATION: REPO — 1\n"},
		{"case 9: the heading variants open and are scanned",
			[]string{"docs/a.md", "docs/b.md", "shots/27.png", "shots/28.png", "shots/29.png"},
			func(dir string) {
				dwFile(t, dir, "docs/a.md", "### Done-When:\n\n`shots/27.png` and `shots/28.png` re-baselined.\n")
				dwFile(t, dir, "docs/b.md", "## done when\n\n`shots/29.png` re-baselined.\n")
			},
			0, "DONE-WHEN-OK: REPO\n"},
		{"case 10: a peer-qualified name fails as the untracked name it is here",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md", "## Done when: peer:shots/27.png re-baselined\n")
			},
			1, "DONE-WHEN-PATH: peer:shots/27.png — docs/tickets/kan-x.md\nDONE-WHEN-VIOLATION: REPO — 1\n"},
		{"case 14: punctuation wrapping a backticked path yields the bare path (F1)",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md", "## Done when\n(`shots/27.png` re-baselined) and `shots/28.png`.\n")
			},
			1, "DONE-WHEN-PATH: shots/27.png — docs/tickets/kan-x.md\nDONE-WHEN-PATH: shots/28.png — docs/tickets/kan-x.md\nDONE-WHEN-VIOLATION: REPO — 2\n"},
		{"case 15: a wrapped backticked path that IS tracked is DONE-WHEN-OK (F1)",
			[]string{"docs/tickets/kan-x.md", "shots/27.png"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md", "## Done when\n(`shots/27.png` re-baselined)\n")
			},
			0, "DONE-WHEN-OK: REPO\n"},
		{"case 16: a dotless name is judged like any other (F3)",
			[]string{"docs/tickets/kan-x.md", "src/Makefile"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md", "## Done when\n- `src/Makefile` committed\n- docs/LICENSE signed\n")
			},
			1, "DONE-WHEN-PATH: docs/LICENSE — docs/tickets/kan-x.md\nDONE-WHEN-VIOLATION: REPO — 1\n"},
		{"case 17: a fence line neither closes a section nor is judged (F4)",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md",
					"## Done when\n\n```bash\n# rebuild the fixtures\nshots/decoy.png\n```\n\n`shots/27.png` re-baselined.\n")
			},
			1, "DONE-WHEN-PATH: shots/27.png — docs/tickets/kan-x.md\nDONE-WHEN-VIOLATION: REPO — 1\n"},
		{"case 18: a quoted Done-when template inside a fence opens no section (F4)",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md",
					"## Scope\n\n```markdown\n## Done when\n\n`shots/27.png` re-baselined.\n```\n\nDone.\n")
			},
			0, "DONE-WHEN-OK: REPO\n"},
		{"case 19: a slash command is not a path — gymie's ticket wording",
			[]string{"docs/tickets/kan-x.md"},
			func(dir string) {
				dwFile(t, dir, "docs/tickets/kan-x.md",
					"## Done when\n* A change opened via `/myflow-start <ISSUE-KEY> <slug>` and archived.\n")
			},
			0, "DONE-WHEN-OK: REPO\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir := dwRepo(t, c.repo...)
			c.md(dir)
			code, out, errb := runDoneWhen(t, dir)
			want := strings.ReplaceAll(c.want, "REPO", dir)
			if code != c.rc || out != want {
				t.Fatalf("exit %d, stdout %q, stderr %q; want %d and %q", code, out, errb, c.rc, want)
			}
		})
	}

	// The negative of case 9: a heading that merely begins with "Done when"
	// opens no section, so the untracked path under it is never judged.
	t.Run("case 9 negative: Done whenever opens no section", func(t *testing.T) {
		t.Parallel()
		dir := dwRepo(t, "docs/a.md")
		dwFile(t, dir, "docs/a.md", "## Done whenever the pad rises\n\n`shots/27.png` mentioned.\n")
		code, out, _ := runDoneWhen(t, dir)
		if code != 0 || out != "DONE-WHEN-OK: "+dir+"\n" {
			t.Fatalf("exit %d, stdout %q; want 0 and DONE-WHEN-OK", code, out)
		}
	})

	t.Run("case 11: a tracked file missing from the worktree cannot answer", func(t *testing.T) {
		t.Parallel()
		dir := dwRepo(t, "docs/tickets/kan-x.md")
		if err := os.Remove(dir + "/docs/tickets/kan-x.md"); err != nil {
			t.Fatal(err)
		}
		code, out, errb := runDoneWhen(t, dir)
		if code != 2 || out != "" || !strings.Contains(errb, "cannot read tracked file") {
			t.Fatalf("exit %d, stdout %q, stderr %q; want 2, nothing on stdout, the read failure named", code, out, errb)
		}
	})

	t.Run("case 12: not a worktree", func(t *testing.T) {
		t.Parallel()
		code, out, errb := runDoneWhen(t, t.TempDir())
		if code != 2 || out != "" || errb == "" {
			t.Fatalf("exit %d, stdout %q; want 2 and nothing on stdout", code, out)
		}
	})

	t.Run("case 13: usage", func(t *testing.T) {
		t.Parallel()
		dir := dwRepo(t)
		code, out, _ := runDoneWhen(t, dir, dir, dir)
		if code != 2 || out != "" {
			t.Fatalf("exit %d, stdout %q; want 2 and nothing on stdout", code, out)
		}
	})
}
