package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-markdown-integrity.sh at bbfe685e, one
// subtest per ok label, plus the refusals and the exact output line the
// harness never pinned. The harness ran the .py as a subprocess against a
// fixture root; here the guard runs in-process with the same argument.

func mdiRun(t *testing.T, env Env, args ...string) (int, string, string) {
	t.Helper()
	fn := Registry["check-markdown-integrity"]
	if fn == nil {
		t.Fatal("check-markdown-integrity is not registered")
	}
	var stdout, stderr bytes.Buffer
	rc := fn(args, env, &stdout, &stderr)
	return rc, stdout.String(), stderr.String()
}

// mdiProject writes body (plus the trailing newline the harness's printf
// added) at rel under a fresh root and returns the root.
func mdiProject(t *testing.T, rel, body string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, rel), body+"\n")
	return root
}

func TestCheckMarkdownIntegrity(t *testing.T) {
	t.Parallel()
	const skill = "skills/foo/SKILL.md"

	// expect runs the guard on a one-file fixture; needle "" checks the exit only.
	expect := func(label string, want int, needle, body string) {
		t.Run(label, func(t *testing.T) {
			t.Parallel()
			root := mdiProject(t, skill, body)
			rc, out, errOut := mdiRun(t, crEnv(root, nil), root)
			if rc != want || !strings.Contains(out+errOut, needle) {
				t.Errorf("rc=%d want %d, needle %q\n%s%s", rc, want, needle, out, errOut)
			}
		})
	}

	expect("case 1: escaped backtick orphans the real closer — exit 1", 1, "broken code span",
		"This span is `foo\\`bar` and now the closing backtick is orphaned.")
	expect("case 2: double-backtick delimiter around a literal backtick — exit 0", 0, "",
		"Use ``get `foo` done`` to run it.")
	expect("case 3: odd backtick count outside a fence — exit 1", 1, "broken code span",
		"This line has one stray backtick ` here and nothing closes it.")
	expect("case 4: unmarked prose torn from its blockquote — exit 1", 1, "orphaned blockquote body",
		"The following important point stands out\n> and continues here without its own marker.")
	expect("case 5: blockquote after a completed sentence — exit 0", 0, "",
		"The following point stands out.\n> and this blockquote begins fresh from a real point.")
	expect("case 6: paragraph torn mid-clause — exit 1", 1, "unterminated paragraph",
		"This paragraph starts fine but then just stops without")
	expect("case 7: torn-looking text inside fence/table/list-item — exit 0", 0, "",
		"```text\nThis looks unterminated and has a lone backtick ` here too\n```\n\n"+
			"| Col | text without period |\n|-----|----------------------|\n| a   | b                    |\n\n"+
			"- This list item also stops without terminal punctuation and has ` a backtick")
	expect("case 8: colon followed by a list — exit 0", 0, "",
		"The steps are as follows:\n\n- Step one\n- Step two")
	expect("case 9: colon promise with nothing following before a heading — exit 1", 1, "dangling promise",
		"The following applies here as well:\n\n## Next heading")
	expect("case 10: a citation ending in a full stop before a heading — exit 0", 0, "",
		"This section explains the rationale in detail. See **Widget** (`widget.md`) for why this design was chosen.\n\n## Next heading")

	t.Run("case 11: an unreadable file is cannot-answer — exit 2", func(t *testing.T) {
		t.Parallel()
		root := mdiProject(t, skill, "Some ordinary prose that ends with a period.")
		path := filepath.Join(root, skill)
		if err := os.Chmod(path, 0); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(path, 0o644)
		rc, out, errOut := mdiRun(t, crEnv(root, nil), root)
		want := "check-markdown-integrity: cannot read " + path + ": [Errno 13] Permission denied: '" + path + "'\n"
		if rc != 2 || out != "" || errOut != want {
			t.Errorf("rc=%d out=%q err=%q", rc, out, errOut)
		}
	})
	t.Run("case 12: a nonexistent scope root — exit 2", func(t *testing.T) {
		t.Parallel()
		root := filepath.Join(t.TempDir(), "nonexistent-check-markdown-integrity-root")
		rc, _, errOut := mdiRun(t, crEnv(t.TempDir(), nil), root)
		if rc != 2 || errOut != "check-markdown-integrity: cannot answer — scope root does not exist: "+root+"\n" {
			t.Errorf("rc=%d err=%q", rc, errOut)
		}
	})

	expect("case 13: YAML frontmatter is not scanned as prose — exit 0", 0, "",
		"---\nname: foo\ndescription: bar\n---\n\n# Title\n\nSome closing prose that ends properly.")
	expect("case 14: a paragraph introducing a fenced command — exit 0", 0, "",
		"Run the following command\n\n```bash\necho hi\n```")
	expect("case 15: a bold label directly above the list it introduces — exit 0", 0, "",
		"**Explore the problem space**\n- Ask clarifying questions\n- Challenge assumptions")
	expect("case 16: an indented code block is not scanned as prose — exit 0", 0, "",
		"This text introduces an indented example below.\n\n    literal example line with no terminal punctuation")
	expect("case 17: torn paragraph before an unrelated fence — exit 1", 1, "unterminated paragraph",
		"This paragraph starts fine but then just stops without\n\n```bash\necho x\n```")
	expect("case 19: dangling word before a fence, resumed afterward — exit 0", 0, "",
		"This paragraph spans an embedded example and\n\n```bash\necho x\n```\n\nresumes right where it left off.")
	expect("case 20: dangling word before a fence, followed by a new sentence — exit 1", 1, "unterminated paragraph",
		"This paragraph looks like it tears off after the fence and\n\n```bash\necho x\n```\n\nNew sentence begins here instead of resuming.")
	expect("case 21: a blockquote immediately after a heading — exit 0", 0, "",
		"## A heading with no terminal punctuation\n> a blockquote that begins right after a heading")
	expect("case 22: a blockquote immediately after a fence — exit 0", 0, "",
		"~~~text\necho hi\n~~~\n> a blockquote that begins right after a fence")
	expect("case 23: a blockquote immediately after a list item — exit 0", 0, "",
		"- a list item with no terminal punctuation\n> a blockquote that begins right after a list item")
	expect("case 24: a blockquote immediately after a table — exit 0", 0, "",
		"| Col | Text |\n|-----|------|\n> a blockquote that begins right after a table")
	expect("case 25: unmarked prose (not a completing kind) torn from its blockquote — exit 1", 1, "orphaned blockquote body",
		"This important point stands out but was never marked\n> and continues here without its own marker.")

	// Beyond the harness: the exact report lines, sorted by path then line,
	// with rules/*.mdc scanned and files outside the scope ignored.
	t.Run("report lines are exact and sorted", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "skills/b/x.md"), "Torn and\n\nOne `stray.\n")
		writeFile(t, filepath.Join(root, "rules/r.mdc"), "Promise:\n")
		writeFile(t, filepath.Join(root, "rules/skip.md"), "Torn and\n")
		writeFile(t, filepath.Join(root, "docs/d.md"), "Torn and\n")
		rc, out, errOut := mdiRun(t, crEnv(root, nil), root+"/")
		want := root + "/rules/r.mdc:1: dangling promise: a colon promises content, and the block ends with nothing following it\n" +
			root + "/skills/b/x.md:1: unterminated paragraph: this paragraph's last line ends with no terminal punctuation\n" +
			root + "/skills/b/x.md:3: broken code span: a backtick run in this paragraph has no same-length partner, so a code span closes early or never closes at all\n"
		if rc != 1 || out != want || errOut != "" {
			t.Errorf("rc=%d err=%q\n got %q\nwant %q", rc, errOut, out, want)
		}
	})
	t.Run("no argument scans FLOW_GUARD_REPO_ROOT", func(t *testing.T) {
		t.Parallel()
		root := mdiProject(t, skill, "Torn and")
		rc, out, _ := mdiRun(t, crEnv(t.TempDir(), map[string]string{"FLOW_GUARD_REPO_ROOT": root}))
		if rc != 1 || !strings.HasPrefix(out, root+"/skills/foo/SKILL.md:1: unterminated paragraph") {
			t.Errorf("rc=%d out=%q", rc, out)
		}
	})
	t.Run("a relative root resolves against the working directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "p", skill), "Torn and\n")
		rc, out, _ := mdiRun(t, crEnv(dir, nil), "p")
		if rc != 1 || !strings.HasPrefix(out, "p/skills/foo/SKILL.md:1: ") {
			t.Errorf("rc=%d out=%q", rc, out)
		}
	})
	t.Run("no argument and FLOW_GUARD_REPO_ROOT unset — exit 2", func(t *testing.T) {
		t.Parallel()
		rc, _, errOut := mdiRun(t, crEnv(t.TempDir(), nil))
		if rc != 2 || errOut != "check-markdown-integrity: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-markdown-integrity.sh, which sets it\n" {
			t.Errorf("rc=%d err=%q", rc, errOut)
		}
	})
	t.Run("two arguments — exit 2", func(t *testing.T) {
		t.Parallel()
		rc, _, errOut := mdiRun(t, crEnv(t.TempDir(), nil), "a", "b")
		if rc != 2 || errOut != "check-markdown-integrity: unknown argument — this guard takes at most one, an optional project root\n" {
			t.Errorf("rc=%d err=%q", rc, errOut)
		}
	})
	t.Run("a dash argument — exit 2", func(t *testing.T) {
		t.Parallel()
		rc, _, errOut := mdiRun(t, crEnv(t.TempDir(), nil), "--help")
		if rc != 2 || errOut != "check-markdown-integrity: unknown argument '--help' — this guard takes at most one, an optional project root\n" {
			t.Errorf("rc=%d err=%q", rc, errOut)
		}
	})
	t.Run("invalid UTF-8 is cannot-answer — exit 2", func(t *testing.T) {
		t.Parallel()
		root := mdiProject(t, skill, "ok.\xff")
		path := filepath.Join(root, skill)
		rc, _, errOut := mdiRun(t, crEnv(root, nil), root)
		want := "check-markdown-integrity: cannot read " + path + ": 'utf-8' codec can't decode byte 0xff in position 3: invalid start byte\n"
		if rc != 2 || errOut != want {
			t.Errorf("rc=%d err=%q", rc, errOut)
		}
	})
	t.Run("an unreadable rules directory is cannot-answer — exit 2", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		rules := filepath.Join(root, "rules")
		mkdir(t, rules)
		if err := os.Chmod(rules, 0); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(rules, 0o755)
		rc, _, errOut := mdiRun(t, crEnv(root, nil), root)
		if rc != 2 || !strings.HasPrefix(errOut, "check-markdown-integrity: cannot read "+rules+": ") {
			t.Errorf("rc=%d err=%q", rc, errOut)
		}
	})
	t.Run("CR and form-feed line breaks split as Python's splitlines does", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		writeFile(t, filepath.Join(root, skill), "Ok.\f# H\rTorn and\r\n")
		rc, out, _ := mdiRun(t, crEnv(root, nil), root)
		if rc != 1 || out != root+"/skills/foo/SKILL.md:3: unterminated paragraph: this paragraph's last line ends with no terminal punctuation\n" {
			t.Errorf("rc=%d out=%q", rc, out)
		}
	})
	t.Run("a symlinked file is scanned, a symlinked directory is not", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		outside := t.TempDir()
		writeFile(t, filepath.Join(outside, "t.md"), "Torn and\n")
		mkdir(t, filepath.Join(root, "skills"))
		if err := os.Symlink(filepath.Join(outside, "t.md"), filepath.Join(root, "skills/l.md")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(root, "skills/d")); err != nil {
			t.Fatal(err)
		}
		rc, out, _ := mdiRun(t, crEnv(root, nil), root)
		if rc != 1 || out != root+"/skills/l.md:1: unterminated paragraph: this paragraph's last line ends with no terminal punctuation\n" {
			t.Errorf("rc=%d out=%q", rc, out)
		}
	})
}

// TestCheckMarkdownIntegrityLayoutSignals: signal 5 (missing final newline)
// and signal 6 (consecutive blank lines outside a fence), kan-1034.
func TestCheckMarkdownIntegrityLayoutSignals(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		label, body string
		want        int
		needle      string
	}{
		{"terminated", "Text.\n", 0, ""},
		{"no final newline", "Text.", 1, "skills/foo/SKILL.md:1: missing final newline:"},
		{"doubled blank", "One.\n\n\nTwo.\n", 1, ":3: consecutive blank lines:"},
		{"doubled blank inside a fence", "One.\n\n```\na\n\n\nb\n```\n", 0, ""},
		{"empty file", "", 0, ""},
	} {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFile(t, filepath.Join(root, "skills/foo/SKILL.md"), c.body)
			rc, out, errOut := mdiRun(t, crEnv(root, nil), root)
			if rc != c.want || !strings.Contains(out+errOut, c.needle) {
				t.Errorf("rc=%d want %d, needle %q\n%s%s", rc, c.want, c.needle, out, errOut)
			}
		})
	}
}
