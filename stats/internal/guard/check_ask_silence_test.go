package guard

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of the check-ask-silence contract, one subtest per behaviour:
// the guard fails a corpus section that mentions AskUserQuestion and states
// neither its own silence outcome, nor a citation of Unanswered mid-run
// asks, nor a delegation of the ask — and passes every corpus that carries
// no such section. Like the rest of the suite, no case touches this
// checkout: FLOW_GUARD_REPO_ROOT names a t.TempDir() corpus the test lays
// out itself.

// asFixtureRoot lays out a minimal corpus — every scope root the guard
// refuses when missing — and returns its path.
func asFixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range askSilenceScopes {
		mkdir(t, filepath.Join(root, dir))
	}
	return root
}

// asRunNoRoot invokes the guard with no FLOW_GUARD_REPO_ROOT set.
func asRunNoRoot(t *testing.T) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	env := Env{Getenv: func(string) string { return "" }}
	code := Registry["check-ask-silence"](nil, env, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// asRunWithRoot invokes the guard against root.
func asRunWithRoot(t *testing.T, root string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	env := Env{Getenv: func(k string) string {
		if k == "FLOW_GUARD_REPO_ROOT" {
			return root
		}
		return ""
	}}
	code := Registry["check-ask-silence"](nil, env, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestAskSilence(t *testing.T) {
	t.Parallel()

	t.Run("unset root refuses with exit 2", func(t *testing.T) {
		t.Parallel()
		code, _, stderr := asRunNoRoot(t)
		if code != 2 {
			t.Errorf("rc=%d, want 2\nstdout+stderr: %s%s", code, "", stderr)
		}
		if !strings.Contains(stderr, "FLOW_GUARD_REPO_ROOT is unset") {
			t.Errorf("refusal does not name the variable:\n%s", stderr)
		}
	})

	t.Run("missing scope root refuses with exit 2", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		mkdir(t, filepath.Join(root, "skills")) // only one of the five roots
		code, _, stderr := asRunWithRoot(t, root)
		if code != 2 {
			t.Errorf("rc=%d, want 2\nstderr:\n%s", code, stderr)
		}
		if !strings.Contains(stderr, "scope root missing or unreadable") {
			t.Errorf("refusal does not name the missing root:\n%s", stderr)
		}
	})

	t.Run("symlinked scope root refuses with exit 2", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		if err := os.RemoveAll(filepath.Join(root, "rules")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(root, "skills"), filepath.Join(root, "rules")); err != nil {
			t.Fatal(err)
		}
		code, _, stderr := asRunWithRoot(t, root)
		if code != 2 {
			t.Errorf("rc=%d, want 2\nstderr:\n%s", code, stderr)
		}
		if !strings.Contains(stderr, "scope root is a symlink") {
			t.Errorf("refusal does not name the symlinked root:\n%s", stderr)
		}
	})

	t.Run("empty corpus is OK", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		writeFile(t, filepath.Join(root, "skills", "s.md"), "# Heading\n\nNo asks here.\n")
		writeFile(t, filepath.Join(root, "README.md"), "# Root file\n\nPlain prose, owned at the root.\n")
		code, stdout, stderr := asRunWithRoot(t, root)
		if code != 0 {
			t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if !strings.Contains(stdout, "no AskUserQuestion site in the corpus") {
			t.Errorf("OK verdict does not report the empty corpus:\n%s", stdout)
		}
	})

	t.Run("silent ask section fails, named by file and line", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		writeFile(t, filepath.Join(root, "skills", "s.md"),
			"# Intro\n\nPreamble prose.\n\n## The ask\n\nAsk the operator through **AskUserQuestion** with the options.\n\n## Next\n\nNothing here.\n")
		code, stdout, stderr := asRunWithRoot(t, root)
		if code != 1 {
			t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if !strings.Contains(stderr, "skills/s.md:6") {
			t.Errorf("violation does not name file:line:\n%s", stderr)
		}
		if !strings.Contains(stderr, `"## The ask"`) {
			t.Errorf("violation does not name the section heading:\n%s", stderr)
		}
		if !strings.Contains(stdout, "ASK-SILENCE-FAIL") {
			t.Errorf("no FAIL verdict on stdout:\n%s", stdout)
		}
	})

	t.Run("each pass reason passes", func(t *testing.T) {
		t.Parallel()
		cases := map[string]string{
			"citation":    "Ask the operator through **AskUserQuestion**; silence resolves under **Unanswered mid-run asks**.",
			"vocabulary":  "Ask the operator through **AskUserQuestion**; the first option is the default and is marked recommended.",
			"explicit":    "Ask through **AskUserQuestion**; proceed only on an explicit yes.",
			"without-ask": "Ask through **AskUserQuestion**, pull request recommended; with handoff none, take it without asking.",
			"delegation":  "Ask through **AskUserQuestion** exactly as **Change name resolution** (`skills/flow-contracts/pipeline.md`) defines it.",
			"applies":     "Ask through **AskUserQuestion** — the section applies **The handshake** unchanged.",
			"under-bold":  "The ask's approval is the first option under **Convergence** below, asked through **AskUserQuestion**.",
			"case-fold":   "Ask through **AskUserQuestion** — silence takes Continue; this file spells it SILENCE.",
			"case-fold-2": "Ask through **AskUserQuestion** — take the DEFAULT option.",
		}
		for name, body := range cases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				root := asFixtureRoot(t)
				writeFile(t, filepath.Join(root, "skills", fmt.Sprintf("%s.md", name)), "# H\n\n"+body+"\n")
				code, stdout, stderr := asRunWithRoot(t, root)
				if code != 0 {
					t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
				}
				if !strings.Contains(stdout, "ASK-SILENCE-OK") {
					t.Errorf("no OK verdict:\n%s", stdout)
				}
			})
		}
	})

	t.Run("mention and outcome inside a fence are not read", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		writeFile(t, filepath.Join(root, "skills", "fenced.md"),
			"# H\n\n```\nAskUserQuestion in a worked example, silence default recommended\n```\n\nNo live ask.\n")
		code, stdout, stderr := asRunWithRoot(t, root)
		if code != 0 {
			t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
	})

	t.Run("outcome inside a fence does not pass a live ask", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		writeFile(t, filepath.Join(root, "skills", "fenced-outcome.md"),
			"# H\n\n```\nsilence takes the default\n```\n\nAsk through **AskUserQuestion**.\n")
		code, _, stderr := asRunWithRoot(t, root)
		if code != 1 {
			t.Errorf("rc=%d, want 1 — a fenced outcome must not pass a live ask\nstderr:\n%s", code, stderr)
		}
	})

	t.Run("non-ask sections are never scored", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		writeFile(t, filepath.Join(root, "rules", "r.mdc"),
			"# Rules\n\nNames the ask tool nowhere. The word silence, the word default, the word recommended — unscored without it.\n")
		code, stdout, stderr := asRunWithRoot(t, root)
		if code != 0 {
			t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if strings.Contains(stdout, "1 ask section") {
			t.Errorf("a section without the tool mention was scored:\n%s", stdout)
		}
	})

	t.Run("preamble ask section is scored and fails", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		writeFile(t, filepath.Join(root, ".flow", "project.md"),
			"Ask the operator through **AskUserQuestion** before anything runs.\n\n# Heading\n\nAfter.\n")
		code, _, stderr := asRunWithRoot(t, root)
		if code != 1 {
			t.Errorf("rc=%d, want 1\nstderr:\n%s", code, stderr)
		}
		if !strings.Contains(stderr, `"(preamble)"`) {
			t.Errorf("violation does not name the preamble section:\n%s", stderr)
		}
	})

	t.Run("nested symlink hiding markdown refuses with exit 2", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		outside := t.TempDir()
		writeFile(t, filepath.Join(outside, "hidden.md"), "# Hidden\n\nAskUserQuestion somewhere.\n")
		mkdir(t, filepath.Join(root, "skills", "sub"))
		if err := os.Symlink(outside, filepath.Join(root, "skills", "sub", "link")); err != nil {
			t.Fatal(err)
		}
		code, _, stderr := asRunWithRoot(t, root)
		if code != 2 {
			t.Errorf("rc=%d, want 2\nstderr:\n%s", code, stderr)
		}
		if !strings.Contains(stderr, "hides Markdown from the corpus") {
			t.Errorf("refusal does not name the hiding link:\n%s", stderr)
		}
	})

	t.Run("symlink to a bare file is skipped, not refused", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		outside := t.TempDir()
		writeFile(t, filepath.Join(outside, "plain.md"), "# Plain\n\nAskUserQuestion here lives outside the corpus.\n")
		mkdir(t, filepath.Join(root, "skills", "sub"))
		if err := os.Symlink(filepath.Join(outside, "plain.md"), filepath.Join(root, "skills", "sub", "file-link.md")); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := asRunWithRoot(t, root)
		if code != 0 {
			t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
	})

	t.Run("dangling symlink is skipped, not refused", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		mkdir(t, filepath.Join(root, "skills", "sub"))
		if err := os.Symlink(filepath.Join(root, "skills", "sub", "nowhere.md"), filepath.Join(root, "skills", "sub", "dangling")); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := asRunWithRoot(t, root)
		if code != 0 {
			t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
	})

	t.Run("excluded trees are not walked", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		writeFile(t, filepath.Join(root, "skills", "x", "node_modules", "vendored.md"), "AskUserQuestion silent default\n")
		writeFile(t, filepath.Join(root, "skills", "x", ".superpowers", "scratch.md"), "AskUserQuestion silent default\n")
		writeFile(t, filepath.Join(root, "spectre", "changes", "archive", "old", "a.md"), "AskUserQuestion silent default\n")
		writeFile(t, filepath.Join(root, "docs", "superpowers", "v.md"), "AskUserQuestion silent default\n")
		code, stdout, stderr := asRunWithRoot(t, root)
		if code != 0 {
			t.Errorf("rc=%d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if !strings.Contains(stdout, "no AskUserQuestion site") {
			t.Errorf("an excluded tree leaked into the corpus: code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	})

	t.Run("delegation without a bold target is not credit", func(t *testing.T) {
		t.Parallel()
		root := asFixtureRoot(t)
		writeFile(t, filepath.Join(root, "skills", "d.md"), "# H\n\nAsk through **AskUserQuestion** per the usual convention.\n")
		code, _, stderr := asRunWithRoot(t, root)
		if code != 1 {
			t.Errorf("rc=%d, want 1 — `per the` carries no bolded target\nstderr:\n%s", code, stderr)
		}
	})
}
