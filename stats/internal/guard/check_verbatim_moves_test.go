package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// vmRepo is a git repository in t.TempDir() whose first commit holds base
// (path relative to the root: body) and whose `origin/HEAD` names main, so
// the guard's default base resolves offline. Returns the root.
func vmRepo(t *testing.T, base map[string]string) string {
	t.Helper()
	root := t.TempDir()
	gitRun(t, root, "init", "-q", "-b", "main")
	for rel, body := range base {
		writeFile(t, filepath.Join(root, rel), body)
	}
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-q", "--allow-empty", "-m", "base")
	gitRun(t, root, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitRun(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	return root
}

func vmRun(root string, args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	env := Env{
		Dir:       root,
		Getenv:    func(k string) string { v, _ := vmLookup(root)(k); return v },
		LookupEnv: vmLookup(root),
	}
	rc := checkVerbatimMoves(args, env, &out, &errb)
	return rc, out.String(), errb.String()
}

func vmLookup(root string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		switch k {
		case "FLOW_GUARD_REPO_ROOT":
			return root, true
		case "PATH", "HOME":
			return os.LookupEnv(k)
		case "GIT_CONFIG_GLOBAL":
			return "/dev/null", true
		}
		return "", false
	}
}

const (
	vmPipe = "skills/flow-contracts/pipeline.md"
	vmRat  = "skills/flow-contracts/pipeline-rationale.md"
	// The three sentences of the prototype's self-test at 700e184.
	vmWhy   = "This is why no `*-done` command exists — there would be nothing for one to write."
	vmTable = "**This table is authoritative.**"
	vmMark  = "**A mark never blocks, delays, or alters the stage it marks.**"
)

var vmBase = map[string]string{
	vmPipe: "# Pipeline\n\n" + vmTable + " Every row is a transition.\n\n" + vmMark + "\n\n" + vmWhy + "\n",
	vmRat:  "# Pipeline — rationale\n\nKept for whoever edits the contract.\n",
}

func TestCheckVerbatimMoves(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		base   map[string]string // extra base files, committed with vmBase
		edit   map[string]string // working-tree bodies written over the base
		remove []string
		args   []string
		rc     int
		want   []string // substrings of stdout
		reject []string // substrings stdout must not carry
	}{
		{name: "no change is clean", rc: 0,
			want: []string{"check-verbatim-moves: 0 violation(s), 0 to review"}},
		{name: "a verbatim move into a rationale file passes",
			edit: map[string]string{
				vmPipe: "# Pipeline\n\n" + vmTable + " Every row is a transition.\n\n" + vmMark + "\n",
				vmRat:  "# Pipeline — rationale\n\nKept for whoever edits the contract.\n\n" + vmWhy + "\n",
			},
			rc: 0, want: []string{"ok   moved"}, reject: []string{"FAIL"}},
		{name: "a reworded sentence fails",
			edit: map[string]string{
				vmPipe: "# Pipeline\n\n**This table is canonical.** Every row is a transition.\n\n" + vmMark + "\n\n" + vmWhy + "\n",
			},
			rc: 1, want: []string{"FAIL deleted or reworded :: **This table is authoritative.\n", "FAIL new run-loaded text (paraphrase?) :: **This table is canonical.\n"}},
		{name: "a deleted rule fails",
			edit: map[string]string{
				vmPipe: "# Pipeline\n\n" + vmTable + " Every row is a transition.\n\n" + vmWhy + "\n",
			},
			rc: 1, want: []string{"FAIL deleted or reworded :: " + vmMark}},
		{name: "a rule moved into a rationale file is reported for review, not failed",
			edit: map[string]string{
				vmPipe: "# Pipeline\n\n" + vmTable + " Every row is a transition.\n\n" + vmWhy + "\n",
				vmRat:  "# Pipeline — rationale\n\nKept for whoever edits the contract.\n\n" + vmMark + "\n",
			},
			rc: 0, want: []string{"REVIEW moved to a -rationale.md but carries an imperative marker :: " + vmMark, "1 to review"}},
		{name: "a de-duplication still stated elsewhere passes",
			base: map[string]string{"skills/flow/SKILL.md": "# Flow\n\n" + vmMark + "\n"},
			edit: map[string]string{
				vmPipe:                  "# Pipeline\n\n" + vmTable + " Every row is a transition.\n\n" + vmWhy + "\n",
				"commands-claude/f.md":  "# f\n\nLoad `skills/flow/SKILL.md` only when the run needs it.\n",
				"rules/x.mdc":           "Read **State file** (`skills/flow-contracts/state-file.md`).\n",
				"skills/flow/README.md": "```bash\necho new code line here\n```\n",
			},
			rc: 1, want: []string{"ok   dedup", "FAIL new run-loaded text (paraphrase?) :: echo new code line here"},
			reject: []string{"FAIL new run-loaded text (paraphrase?) :: Load `skills", "FAIL new run-loaded text (paraphrase?) :: Read **State file**", "FAIL deleted or reworded"}},
		{name: "an acknowledged rewording in an in-flight change passes",
			edit: map[string]string{
				vmPipe: "# Pipeline\n\n**This table is canonical.** Every row is a transition.\n\n" + vmMark + "\n\n" + vmWhy + "\n",
				"spectre/changes/demo/verbatim-moves.txt": "# deliberate\n**This table is authoritative.\n**This table is canonical.\n",
			},
			rc: 0, want: []string{"ok   acknowledged", "0 violation(s)"}, reject: []string{"FAIL"}},
		{name: "a heading is acknowledged behind a backslash",
			edit: map[string]string{
				vmPipe: "# Pipeline\n\n## Transition table\n\n" + vmTable + " Every row is a transition.\n\n" + vmMark + "\n\n" + vmWhy + "\n",
				"spectre/changes/demo/verbatim-moves.txt": "\\## Transition table\n",
			},
			rc: 0, want: []string{"ok   acknowledged new text :: ## Transition table", "0 violation(s)"}, reject: []string{"FAIL"}},
		{name: "an unescaped heading in the list is a comment, not an acknowledgement",
			edit: map[string]string{
				vmPipe: "# Pipeline\n\n## Transition table\n\n" + vmTable + " Every row is a transition.\n\n" + vmMark + "\n\n" + vmWhy + "\n",
				"spectre/changes/demo/verbatim-moves.txt": "## Transition table\n",
			},
			rc: 1, want: []string{"FAIL new run-loaded text (paraphrase?) :: ## Transition table"}},
		{name: "an acknowledgement in an archived change does not count",
			edit: map[string]string{
				vmPipe: "# Pipeline\n\n" + vmTable + " Every row is a transition.\n\n" + vmWhy + "\n",
				"spectre/changes/archive/old/verbatim-moves.txt": vmMark + "\n",
			},
			rc: 1, want: []string{"FAIL deleted or reworded :: " + vmMark}},
		{name: "a deleted file fails every sentence it held",
			remove: []string{vmPipe},
			rc:     1, want: []string{"FAIL deleted or reworded :: " + vmWhy}},
		{name: "files outside the scope are ignored",
			edit: map[string]string{"docs/notes.md": "A new sentence nobody loads at all.\n", "rules/sub/deep.md": "A nested rule nobody globs for.\n"},
			rc:   0, reject: []string{"FAIL"}},
		{name: "an explicit base ref is honoured", args: []string{"HEAD"}, rc: 0,
			want: []string{"against HEAD"}},
		{name: "an unknown base ref cannot answer", args: []string{"no-such-ref"}, rc: 2},
		{name: "two arguments are a usage error", args: []string{"HEAD", "HEAD"}, rc: 2},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			base := map[string]string{}
			for _, m := range []map[string]string{vmBase, c.base} {
				for k, v := range m {
					base[k] = v
				}
			}
			root := vmRepo(t, base)
			for rel, body := range c.edit {
				writeFile(t, filepath.Join(root, rel), body)
			}
			for _, rel := range c.remove {
				if err := os.Remove(filepath.Join(root, rel)); err != nil {
					t.Fatal(err)
				}
			}
			rc, out, errb := vmRun(root, c.args...)
			if rc != c.rc {
				t.Fatalf("rc = %d, want %d\nstdout:\n%s\nstderr:\n%s", rc, c.rc, out, errb)
			}
			for _, w := range c.want {
				if !strings.Contains(out, w) {
					t.Errorf("stdout lacks %q:\n%s", w, out)
				}
			}
			for _, r := range c.reject {
				if strings.Contains(out, r) {
					t.Errorf("stdout carries %q:\n%s", r, out)
				}
			}
		})
	}
}

// TestCheckVerbatimMovesSymlinkSkipped pins the prototype's `not
// p.is_symlink()`: skills/flow/scripts/ holds symlinks into scripts/, and a
// symlinked Markdown file is the same text counted twice.
func TestCheckVerbatimMovesSymlinkSkipped(t *testing.T) {
	t.Parallel()
	root := vmRepo(t, vmBase)
	writeFile(t, filepath.Join(root, "docs/linked.md"), "A sentence reachable only through a link.\n")
	mkdir(t, filepath.Join(root, "skills/flow"))
	if err := os.Symlink("../../docs/linked.md", filepath.Join(root, "skills/flow/linked.md")); err != nil {
		t.Fatal(err)
	}
	if rc, out, errb := vmRun(root); rc != 0 {
		t.Fatalf("rc = %d\n%s%s", rc, out, errb)
	}
}

// TestCheckVerbatimMovesNoRoot pins the cannot-answer exit when the shim's
// root is missing or is not a git worktree.
func TestCheckVerbatimMovesNoRoot(t *testing.T) {
	t.Parallel()
	var out, errb bytes.Buffer
	env := Env{Getenv: func(string) string { return "" }, LookupEnv: func(string) (string, bool) { return "", false }}
	if rc := checkVerbatimMoves(nil, env, &out, &errb); rc != 2 {
		t.Fatalf("unset root: rc = %d, want 2", rc)
	}
	dir := t.TempDir()
	if rc, _, _ := vmRun(dir); rc != 2 {
		t.Fatalf("non-repo root: rc = %d, want 2", rc)
	}
}

// TestVMSentences pins the prototype's block and sentence splitting, the
// part a byte-for-byte port gets wrong most easily.
func TestVMSentences(t *testing.T) {
	t.Parallel()
	in := "# A heading that is long\n" +
		"First sentence runs on\nacross a wrap. **Second one.** Third (one).\n" +
		"\n- a list item that is long\n1. a numbered item that is long\n" +
		"> quoted text that is long enough\n" +
		"| a | table row long |\n" +
		"```\n  code stays one line  \n```\n" +
		"short.\n" +
		"\nFollow the rule of **4. Execute (SDD + TDD)** in the file. It ends at step 2. Then more follows here.\n"
	got := vmSentences(in)
	want := []string{
		"# A heading that is long",
		"First sentence runs on across a wrap.",
		"**Second one.",
		"Third (one).",
		"a list item that is long",
		"a numbered item that is long",
		"quoted text that is long enough",
		"| a | table row long |",
		"code stays one line",
		// A bold numbered heading does not end a sentence; a plain one does.
		"Follow the rule of **4. Execute (SDD + TDD)** in the file.",
		"It ends at step 2.",
		"Then more follows here.",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
