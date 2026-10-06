package guard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// pgBashDir is scripts/project-get.sh with the two libs it sources, as they
// stood at ae805186 (Decision: parity-ref-on-main), in a directory of their own.
func pgBashDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, rel := range []string{"project-get.sh", "lib/project-section.sh", "lib/strip-bom.sh"} {
		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "ae805186:scripts/"+rel).Output()
		if err != nil {
			t.Fatalf("git show ae805186:scripts/%s: %v", rel, err)
		}
		writeFile(t, filepath.Join(dir, rel), string(src))
	}
	return dir
}

type pgRes struct {
	rc       int
	out, err string
}

func pgGo(args ...string) pgRes {
	var out, errb bytes.Buffer
	run, ok := Registry["project-get"]
	if !ok {
		return pgRes{-1, "", "no project-get guard registered"}
	}
	rc := run(args, Env{Getenv: os.Getenv}, &out, &errb)
	return pgRes{rc, out.String(), errb.String()}
}

func pgBash(t *testing.T, dir string, args ...string) pgRes {
	t.Helper()
	cmd := exec.Command("bash", append([]string{dir + "/project-get.sh"}, args...)...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	_ = cmd.Run()
	return pgRes{cmd.ProcessState.ExitCode(), out.String(), errb.String()}
}

const pgCfg = "# project\n\n## lint\n\n```bash\nmake lint\n```\n\n## default landing route\n\n`open PR`\n\n## empty\n\n## last\nline one\n### sub\nline two\n"

func TestProjectGetParity(t *testing.T) {
	t.Parallel()
	bash := pgBashDir(t)

	plain := func(t *testing.T, body string) string {
		root := t.TempDir()
		writeFile(t, root+"/.flow/project.md", body)
		return root
	}
	// repo commits head as .flow/project.md under sub (the project root),
	// then leaves wt in the working tree when wt is not nil.
	repo := func(t *testing.T, sub, head string, wt *string) string {
		top := t.TempDir()
		root := filepath.Join(top, sub)
		gitRun(t, top, "init", "-q")
		writeFile(t, root+"/.flow/project.md", head)
		gitRun(t, top, "add", "-A")
		gitRun(t, top, "commit", "-q", "-m", "base")
		if wt != nil {
			writeFile(t, root+"/.flow/project.md", *wt)
		}
		return root
	}
	for _, tc := range []struct {
		name string
		args func(t *testing.T) []string
		rc   int
	}{
		{"present", func(t *testing.T) []string { return []string{plain(t, pgCfg), "lint"} }, 0},
		{"present with subheading", func(t *testing.T) []string { return []string{plain(t, pgCfg), "last"} }, 0},
		{"present empty body", func(t *testing.T) []string { return []string{plain(t, pgCfg), "empty"} }, 0},
		{"key with spaces", func(t *testing.T) []string { return []string{plain(t, pgCfg), "default landing route"} }, 0},
		{"absent key", func(t *testing.T) []string { return []string{plain(t, pgCfg), "nope"} }, 1},
		{"duplicate heading", func(t *testing.T) []string { return []string{plain(t, pgCfg+"## lint\nagain\n"), "lint"} }, 2},
		{"BOM on the first heading", func(t *testing.T) []string { return []string{plain(t, "\xef\xbb\xbf## lint\nmake lint\n"), "lint"} }, 0},
		{"CRLF", func(t *testing.T) []string { return []string{plain(t, "## lint\r\nmake lint\r\n"), "lint"} }, 1},
		{"no file", func(t *testing.T) []string { return []string{t.TempDir(), "lint"} }, 1},
		{"root not a directory", func(t *testing.T) []string { return []string{t.TempDir() + "/missing", "lint"} }, 2},
		{"usage: one argument", func(t *testing.T) []string { return []string{t.TempDir()} }, 2},
		{"usage: three arguments", func(t *testing.T) []string { return []string{t.TempDir(), "a", "b"} }, 2},
		{"HEAD wins over a diverging working tree", func(t *testing.T) []string {
			wt := "## lint\nedited\n"
			return []string{repo(t, "", pgCfg, &wt), "lint"}
		}, 0},
		{"working tree diverging on another key is silent", func(t *testing.T) []string {
			wt := pgCfg + "## other\nx\n"
			return []string{repo(t, "", pgCfg, &wt), "lint"}
		}, 0},
		{"HEAD read from a subdirectory root", func(t *testing.T) []string { return []string{repo(t, "proj", pgCfg, nil), "last"} }, 0},
		{"HEAD carries it, working tree deleted it", func(t *testing.T) []string {
			root := repo(t, "", pgCfg, nil)
			if err := os.Remove(root + "/.flow/project.md"); err != nil {
				t.Fatal(err)
			}
			return []string{root, "lint"}
		}, 0},
		{"HEAD lacks it, working tree has it", func(t *testing.T) []string {
			top := t.TempDir()
			gitRun(t, top, "init", "-q")
			writeFile(t, top+"/README", "x\n")
			gitRun(t, top, "add", "-A")
			gitRun(t, top, "commit", "-q", "-m", "base")
			writeFile(t, top+"/.flow/project.md", pgCfg)
			return []string{top, "lint"}
		}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args := tc.args(t)
			want, got := pgBash(t, bash, args...), pgGo(args...)
			if want.rc != tc.rc {
				t.Fatalf("fixture: the bash exited %d, want %d\n%+v", want.rc, tc.rc, want)
			}
			if got != want {
				t.Errorf("differs from the bash at ae805186\ngot:  %+v\nwant: %+v", got, want)
			}
		})
	}
}

func TestProjectGetEnum(t *testing.T) {
	t.Parallel()
	route := []string{"--enum", "pull request", "merge and push", "manual"}
	for _, tc := range []struct {
		name, cfg string
		args      []string
		rc        int
		out, err  string
	}{
		{"a match prints the literal", "## default landing route\n\nmanual\ndocumentation below\n", route, 0, "manual\n", ""},
		{"a backticked head", "## default landing route\n`merge and push`\n", route, 0, "merge and push\n", ""},
		{"a padded head", "## default landing route\n   ` pull request `  \t\n", route, 0, "pull request\n", ""},
		{"prose after the literal on the head line", "## default landing route\nmerge and push — the standing choice since kan-512\n\nprose below\n", route, 0, "merge and push\n", ""},
		{"a word extending a literal matches nothing", "## default landing route\nmanually\n", route, 3, "",
			"project-get: '## default landing route' head 'manually' matches none of: 'pull request' 'merge and push' 'manual'\n"},
		{"a suffix extending a literal matches nothing", "## default landing route\nmerge and pushed\n", route, 3, "",
			"project-get: '## default landing route' head 'merge and pushed' matches none of: 'pull request' 'merge and push' 'manual'\n"},
		{"absent key", "## lint\nx\n", route, 1, "", "project-get: <root>/.flow/project.md declares no '## default landing route' section\n"},
		{"no match", "## default landing route\n`Pull Request`\n", route, 3, "",
			"project-get: '## default landing route' head 'Pull Request' matches none of: 'pull request' 'merge and push' 'manual'\n"},
		{"empty body matches nothing", "## default landing route\n\n## next\n", route, 3, "",
			"project-get: '## default landing route' head '' matches none of: 'pull request' 'merge and push' 'manual'\n"},
		{"duplicate heading is still exit 2", "## default landing route\nmanual\n## default landing route\nmanual\n", route, 2, "",
			"project-get: <root>/.flow/project.md declares 2 '## default landing route' sections — a second declaration is ambiguous, so neither was read\n"},
		{"usage: --enum with no literal", "## default landing route\nmanual\n", []string{"--enum"}, 2, "",
			"usage: project-get.sh <project-root> <key>\n"},
		{"usage: a third argument that is not --enum", "## default landing route\nmanual\n", []string{"manual"}, 2, "",
			"usage: project-get.sh <project-root> <key>\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFile(t, root+"/.flow/project.md", tc.cfg)
			got := pgGo(append([]string{root, "default landing route"}, tc.args...)...)
			got.err = strings.ReplaceAll(got.err, root, "<root>")
			if want := (pgRes{tc.rc, tc.out, tc.err}); got != want {
				t.Errorf("got:  %+v\nwant: %+v", got, want)
			}
		})
	}
	t.Run("no file", func(t *testing.T) {
		t.Parallel()
		if got := pgGo(append([]string{t.TempDir(), "default landing route"}, route...)...); got.rc != 1 {
			t.Errorf("exit %d, want 1: %+v", got.rc, got)
		}
	})
}
