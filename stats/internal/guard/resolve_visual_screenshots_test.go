package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

// Every case of scripts/test-resolve-visual-screenshots.sh at 3915fbc0, one
// subtest per ok: label, each running the guard in-process against real
// files on disk. The label's own assertion is the harness's, over stdout and
// stderr together (the harness captured 2>&1); every "exit N" label also pins
// the whole stdout and stderr the bash printed. The regression-checkout cases
// run the real git against real worktrees. The rows after the harness's pin
// the port's bytewise order (bytewise-screenshot-sort), the logical spelling
// of a root reached through a symlinked cwd, and find's refusal to follow a
// symlink.

const rvsName = "resolve-visual-screenshots: "

func rvsSection(screenshots string) string {
	return fmt.Sprintf("## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `stats/web/src/**` |\n| `screenshots` | `%s` |\n\n| Command | Runs |\n|---------|------|\n| `verify` | `npm run test:visual` |\n| `capture` | `npx playwright test <spec>` |\n", screenshots)
}

func rvsCheckoutSection(checkout string) string {
	return "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `gymie-frontend/**` |\n| `screenshots` | `tests/visual` |\n| `regression checkout` | `" + checkout + "` |\n| `regression repo` | `git@github.com:tweety53/gymie-playwright.git` |\n\n| Command | Runs |\n|---------|------|\n| `verify` | `npm run test:visual` |\n| `capture` | `npx playwright test <spec>` |\n"
}

// rvsFixture is one run: the cwd, the arguments, and the bash's whole output.
type rvsFixture struct {
	dir            string
	args           []string
	rc             int
	stdout, stderr string
}

// rvsRoot is the harness's new_root: a temp root with an empty .flow/.
func rvsRoot(t *testing.T) string {
	root := t.TempDir()
	mkdir(t, root+"/.flow")
	return root
}

func rvsTouch(t *testing.T, paths ...string) {
	for _, p := range paths {
		writeFile(t, p, "")
	}
}

// rvsGit is the harness's case 16/17 sandbox, built once and only read: a
// main checkout `pw` with worktrees on spectre/change and spectre/other, a
// project root on spectre/change and one on a branch no worktree holds. wt is
// the change worktree's path as `git worktree list --porcelain` spells it.
var rvsGit = sync.OnceValues(func() (struct{ pw, projChange, projElse, wt string }, error) {
	var f struct{ pw, projChange, projElse, wt string }
	top, err := os.MkdirTemp(execFixtures.dir, "resolve-visual-screenshots")
	if err != nil {
		return f, err
	}
	git := func(args ...string) (string, error) {
		cmd := exec.Command(fixtureGit, args...)
		cmd.Env = append(append(os.Environ(), fixtureGitEnv...),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git %v: %v\n%s", args, err, out)
		}
		return string(out), nil
	}
	f.pw, f.projChange, f.projElse = top+"/pw", top+"/proj", top+"/proj-elsewhere"
	if err := os.MkdirAll(f.pw+"/app.spec.ts-snapshots", 0o755); err != nil {
		return f, err
	}
	if err := os.WriteFile(f.pw+"/app.spec.ts-snapshots/home-darwin.png", nil, 0o644); err != nil {
		return f, err
	}
	cfg := "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `gymie-frontend/**` |\n| `screenshots` | `.` |\n| `regression checkout` | `" + f.pw + "` |\n| `regression repo` | `git@github.com:tweety53/gymie-playwright.git` |\n"
	steps := [][]string{
		{"init", "-q", "-b", "main", f.pw},
		{"-C", f.pw, "add", "-A"},
		{"-C", f.pw, "commit", "-q", "-m", "init"},
		{"-C", f.pw, "worktree", "add", "-q", f.pw + "/.worktrees/change", "-b", "spectre/change"},
		{"-C", f.pw, "worktree", "add", "-q", f.pw + "/.worktrees/other", "-b", "spectre/other"},
		{"init", "-q", "-b", "spectre/change", f.projChange},
		{"init", "-q", "-b", "spectre/change", f.projElse},
		{"-C", f.projElse, "checkout", "-q", "-b", "elsewhere"},
	}
	for _, s := range steps {
		if _, err := git(s...); err != nil {
			return f, err
		}
	}
	for _, p := range []string{f.projChange, f.projElse} {
		if err := os.MkdirAll(p+"/.flow", 0o755); err != nil {
			return f, err
		}
		if err := os.WriteFile(p+"/.flow/project.md", []byte(cfg), 0o644); err != nil {
			return f, err
		}
	}
	list, err := git("-C", f.pw, "worktree", "list", "--porcelain")
	if err != nil {
		return f, err
	}
	var w string
	for _, l := range strings.Split(list, "\n") {
		if strings.HasPrefix(l, "worktree ") {
			w = l[len("worktree "):]
		}
		if l == "branch refs/heads/spectre/change" {
			f.wt = w
			break
		}
	}
	return f, nil
})

// rvsCases builds each harness case's fixture; a label's subtest calls it.
var rvsCases = map[string]func(t *testing.T) rvsFixture{
	"case 1": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		rvsTouch(t, root+"/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png")
		writeFile(t, root+"/.flow/project.md", rvsSection("tests/visual"))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 0,
			root + "/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png\n", ""}
	},
	"case 2": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		rvsTouch(t, root+"/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png")
		writeFile(t, root+"/.flow/project.md", rvsSection("."))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 0,
			root + "/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png\n", ""}
	},
	"case 3": func(t *testing.T) rvsFixture {
		root := rvsCase3(t)
		return rvsFixture{root, []string{root, "other.spec.ts"}, 1, "",
			rvsName + "zero PNGs under " + root + "/. matched `other.spec.ts`\n"}
	},
	"case 3b": func(t *testing.T) rvsFixture {
		root := rvsCase3(t)
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 0,
			root + "/node_modules/some-pkg/assets/baseline.spec.ts.png\n" +
				root + "/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png\n", ""}
	},
	"case 4": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		mkdir(t, root+"/tests/visual")
		writeFile(t, root+"/.flow/project.md", rvsSection("tests/visual"))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 1, "",
			rvsName + "zero PNGs under " + root + "/tests/visual matched `baseline.spec.ts`\n"}
	},
	"case 5": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		writeFile(t, root+"/.flow/project.md", rvsSection("tests/visual"))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 1, "",
			rvsName + "zero PNGs under " + root + "/tests/visual matched `baseline.spec.ts`\n"}
	},
	"case 6": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		rvsTouch(t, root+"/checkout/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png",
			root+"/tests/visual/baseline.spec.ts-snapshots/decoy.png")
		writeFile(t, root+"/.flow/project.md", rvsCheckoutSection(root+"/checkout"))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 0,
			root + "/checkout/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png\n", ""}
	},
	// An empty path is not a directory: bash `[ -d "" ]` is false, where
	// joining "" onto the cwd would name the cwd itself.
	"empty root": func(t *testing.T) rvsFixture {
		return rvsFixture{rvsRoot(t), []string{"", "baseline.spec.ts"}, 2, "",
			rvsName + " is not a directory — cannot resolve what it declares\n"}
	},
	"empty checkout": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		rvsTouch(t, root+"/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png")
		writeFile(t, root+"/.flow/project.md", strings.Replace(rvsCheckoutSection("x"), "| `x` |", "|  |", 1))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 2, "",
			rvsName + "`regression checkout` names ``, which is not an existing directory — cannot resolve `screenshots` relative to it\n"}
	},
	"case 7": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 2, "",
			rvsName + root + " has no .flow/project.md — cannot resolve where screenshots live\n"}
	},
	"case 8": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		writeFile(t, root+"/.flow/project.md", "# Project\n\n## run\n\necho hi\n\n")
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 2, "",
			rvsName + root + "/.flow/project.md declares no '## visual verification' section — cannot resolve where screenshots live\n"}
	},
	"case 9": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		writeFile(t, root+"/.flow/project.md", "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `stats/web/src/**` |\n\n| Command | Runs |\n|---------|------|\n| `verify` | `npm run test:visual` |\n| `capture` | `npx playwright test <spec>` |\n")
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 2, "",
			rvsName + "`screenshots` is absent or empty in " + root + "/.flow/project.md — cannot resolve where captures land\n"}
	},
	"case 10": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		writeFile(t, root+"/.flow/project.md", "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `stats/web/src/**` |\n| `screenshots` | |\n\n| Command | Runs |\n|---------|------|\n| `verify` | `npm run test:visual` |\n| `capture` | `npx playwright test <spec>` |\n")
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 2, "",
			rvsName + "`screenshots` is absent or empty in " + root + "/.flow/project.md — cannot resolve where captures land\n"}
	},
	"case 11": func(t *testing.T) rvsFixture {
		root := t.TempDir() + "/nonexistent"
		return rvsFixture{t.TempDir(), []string{root, "baseline.spec.ts"}, 2, "",
			rvsName + root + " is not a directory — cannot resolve what it declares\n"}
	},
	"case 12": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		writeFile(t, root+"/.flow/project.md", rvsSection("tests/visual"))
		return rvsFixture{root, []string{root}, 2, "",
			rvsName + "usage: resolve-visual-screenshots.sh <project root> <capture spec basename>\n"}
	},
	"case 13": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		writeFile(t, root+"/.flow/project.md", rvsCheckoutSection(root+"/does-not-exist"))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 2, "",
			rvsName + "`regression checkout` names `" + root + "/does-not-exist`, which is not an existing directory — cannot resolve `screenshots` relative to it\n"}
	},
	"case 14": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		rvsTouch(t, root+"/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png",
			root+"/tests/visual/visual-baseline.spec.ts-snapshots/decoy.png")
		writeFile(t, root+"/.flow/project.md", rvsSection("tests/visual"))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 0,
			root + "/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png\n", ""}
	},
	"case 15": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		rvsTouch(t, root+"/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png")
		writeFile(t, root+"/.flow/project.md", "\xef\xbb\xbf"+rvsSection("tests/visual"))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 0,
			root + "/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png\n", ""}
	},
	"case 16": func(t *testing.T) rvsFixture {
		g := rvsGitFixture(t)
		return rvsFixture{g.projChange, []string{g.projChange, "app.spec.ts"}, 0,
			g.wt + "/app.spec.ts-snapshots/home-darwin.png\n", ""}
	},
	"case 17": func(t *testing.T) rvsFixture {
		g := rvsGitFixture(t)
		return rvsFixture{g.projElse, []string{g.projElse, "app.spec.ts"}, 0,
			g.pw + "/app.spec.ts-snapshots/home-darwin.png\n", ""}
	},
	// Matches print in byte order, never the caller's collation.
	"bytewise order": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		d := root + "/tests/visual/s.spec.ts-snapshots/"
		rvsTouch(t, d+"B.png", d+"a.png", d+"_x.png", d+"A-1.png")
		writeFile(t, root+"/.flow/project.md", rvsSection("tests/visual"))
		return rvsFixture{root, []string{root, "s.spec.ts"}, 0,
			d + "A-1.png\n" + d + "B.png\n" + d + "_x.png\n" + d + "a.png\n", ""}
	},
	// A relative root from a cwd reached through a symlink prints the
	// logical, $PWD-joined prefix `cd … && pwd` printed.
	"symlinked cwd": func(t *testing.T) rvsFixture {
		td := t.TempDir()
		proj := td + "/real/proj"
		mkdir(t, proj+"/.flow")
		rvsTouch(t, proj+"/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png")
		writeFile(t, proj+"/.flow/project.md", rvsSection("tests/visual"))
		if err := os.Symlink(td+"/real", td+"/link"); err != nil {
			t.Fatal(err)
		}
		return rvsFixture{td + "/link", []string{"proj", "baseline.spec.ts"}, 0,
			td + "/link/proj/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png\n", ""}
	},
	// find -type f without -L: a symlinked PNG is not matched and a
	// symlinked directory is not walked.
	"symlinks not followed": func(t *testing.T) rvsFixture {
		root := rvsRoot(t)
		d := root + "/tests/visual/baseline.spec.ts-snapshots/"
		rvsTouch(t, d+"real.png", root+"/elsewhere/baseline.spec.ts.png")
		if err := os.Symlink(d+"real.png", d+"link.png"); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(root+"/elsewhere", root+"/tests/visual/baseline.spec.ts-linked"); err != nil {
			t.Fatal(err)
		}
		writeFile(t, root+"/.flow/project.md", rvsSection("tests/visual"))
		return rvsFixture{root, []string{root, "baseline.spec.ts"}, 0, d + "real.png\n", ""}
	},
}

func rvsCase3(t *testing.T) string {
	root := rvsRoot(t)
	rvsTouch(t, root+"/node_modules/some-pkg/assets/icon.png",
		root+"/node_modules/some-pkg/assets/baseline.spec.ts.png",
		root+"/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png")
	writeFile(t, root+"/.flow/project.md", rvsSection("."))
	return root
}

func rvsGitFixture(t *testing.T) struct{ pw, projChange, projElse, wt string } {
	g, err := rvsGit()
	if err != nil {
		t.Fatal(err)
	}
	if g.wt == "" {
		t.Fatal("git worktree list names no worktree on spectre/change")
	}
	return g
}

// rvsCheck is one ok: label: its case, the label's text after "<case>: ",
// and its assertion over stderr and stdout together. An "exit N" label (ok
// nil) pins the exit code, stdout and stderr exactly.
type rvsCheck struct {
	kase, label string
	ok          func(out string) bool
}

func rvsExit(kase string, rc int) rvsCheck {
	return rvsCheck{kase, fmt.Sprintf("exit %d", rc), nil}
}

func rvsNames(kase, needle string) rvsCheck {
	return rvsCheck{kase, "output names '" + needle + "'",
		func(out string) bool { return strings.Contains(out, needle) }}
}

func rvsOmits(kase, needle string) rvsCheck {
	return rvsCheck{kase, "output correctly omits '" + needle + "'",
		func(out string) bool { return !strings.Contains(out, needle) }}
}

func rvsPNGLines(kase string, n int) rvsCheck {
	return rvsCheck{kase, fmt.Sprintf("%d PNG line(s)", n), func(out string) bool {
		got := 0
		for _, l := range strings.Split(out, "\n") {
			if strings.HasSuffix(l, ".png") {
				got++
			}
		}
		return got == n
	}}
}

func TestResolveVisualScreenshots(t *testing.T) {
	t.Parallel()
	checks := []rvsCheck{
		rvsExit("case 1", 0),
		rvsNames("case 1", "tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"),
		rvsPNGLines("case 1", 1),
		{"case 1", "printed path is absolute", func(out string) bool { return strings.HasPrefix(out, "/") }},
		rvsExit("case 2", 0), rvsPNGLines("case 2", 1),
		rvsExit("case 3", 1), rvsOmits("case 3", "node_modules"),
		rvsExit("case 3b", 0), rvsPNGLines("case 3b", 2),
		rvsNames("case 3b", "dashboard-darwin.png"),
		rvsNames("case 3b", "node_modules/some-pkg/assets/baseline.spec.ts.png"),
		rvsExit("case 4", 1),
		rvsExit("case 5", 1),
		rvsExit("case 6", 0),
		rvsNames("case 6", "checkout/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"),
		rvsOmits("case 6", "decoy.png"),
		rvsExit("case 7", 2), rvsExit("case 8", 2), rvsExit("case 9", 2), rvsExit("case 10", 2),
		rvsExit("case 11", 2), rvsExit("case 12", 2), rvsExit("case 13", 2),
		rvsExit("case 14", 0), rvsPNGLines("case 14", 1),
		rvsNames("case 14", "baseline.spec.ts-snapshots/dashboard-darwin.png"),
		rvsOmits("case 14", "visual-baseline.spec.ts-snapshots"),
		rvsOmits("case 14", "decoy.png"),
		rvsExit("case 15", 0),
		rvsNames("case 15", "tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png"),
		rvsOmits("case 15", "declares no"),
		rvsExit("case 16", 0), rvsPNGLines("case 16", 1),
		rvsNames("case 16", "/.worktrees/change/app.spec.ts-snapshots/home-darwin.png"),
		rvsExit("case 17", 0), rvsPNGLines("case 17", 1), rvsOmits("case 17", ".worktrees"),
		{"empty root", "exit 2 with the bash's line", nil}, {"empty checkout", "exit 2 with the bash's line", nil},
		rvsExit("bytewise order", 0), rvsExit("symlinked cwd", 0), rvsExit("symlinks not followed", 0),
	}
	for _, c := range checks {
		build := rvsCases[c.kase]
		name := c.kase + ": " + c.label
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := build(t)
			var stdout, stderr bytes.Buffer
			rc := 2
			if fn := Registry["resolve-visual-screenshots"]; fn != nil {
				rc = fn(f.args, Env{Getenv: os.Getenv, Dir: f.dir}, &stdout, &stderr)
			} else {
				stderr.WriteString("resolve-visual-screenshots is not registered\n")
			}
			if c.ok == nil {
				if rc != f.rc || stdout.String() != f.stdout || stderr.String() != f.stderr {
					t.Errorf("rc=%d stdout=%q stderr=%q\nwant rc=%d stdout=%q stderr=%q",
						rc, stdout.String(), stderr.String(), f.rc, f.stdout, f.stderr)
				}
				return
			}
			if out := stderr.String() + stdout.String(); !c.ok(out) {
				t.Errorf("%s failed: rc=%d out=%q", name, rc, out)
			}
		})
	}
}
