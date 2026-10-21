package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Every case of scripts/test-check-normative-inventory.sh at 75261910, one
// subtest per ok label. The harness pointed the guard at a sandboxed fixture
// tree through CHECK_NORMATIVE_INVENTORY_ROOT so it never depended on the
// real corpus's current prose; here the guard runs in-process with the same
// override on Env, over fixtures in t.TempDir(). Every case was proved
// against a deliberately broken guard before it was trusted (the harness's
// own note); the mutations are recorded in KAN-920's task record.

// niRun runs the registered guard with CHECK_NORMATIVE_INVENTORY_ROOT=root.
func niRun(t *testing.T, vars map[string]string) (int, string, string) {
	t.Helper()
	fn := Registry["check-normative-inventory"]
	if fn == nil {
		t.Fatal("check-normative-inventory is not registered")
	}
	var stdout, stderr bytes.Buffer
	rc := fn(nil, crEnv(t.TempDir(), vars), &stdout, &stderr)
	return rc, stdout.String(), stderr.String()
}

// niInv is the harness's inv: the inventory for root, its trailing newline
// dropped as $(...) dropped it. A non-zero exit is a broken premise, fatal.
func niInv(t *testing.T, root string) string {
	t.Helper()
	rc, out, errOut := niRun(t, map[string]string{"CHECK_NORMATIVE_INVENTORY_ROOT": root})
	if rc != 0 {
		t.Fatalf("guard exited %d on fixture %s, wanted 0\nstderr:\n%s", rc, root, errOut)
	}
	return strings.TrimRight(out, "\n")
}

// niRC is the harness's expect_exit: the exit code for root.
func niRC(t *testing.T, root string) int {
	t.Helper()
	rc, _, _ := niRun(t, map[string]string{"CHECK_NORMATIVE_INVENTORY_ROOT": root})
	return rc
}

// niRoot is mkroot: a fixture root carrying every scope root, plus files.
func niRoot(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range ownedCorpusScopes {
		mkdir(t, filepath.Join(root, dir))
	}
	for rel, body := range files {
		writeFile(t, filepath.Join(root, rel), body)
	}
	return root
}

// niSorted is printf '%s\n' … | LC_ALL=C sort, newline-joined.
func niSorted(lines ...string) string {
	s := append([]string(nil), lines...)
	sort.Strings(s)
	return strings.Join(s, "\n")
}

func TestCheckNormativeInventory(t *testing.T) {
	t.Parallel()

	t.Run("a deleted SHALL sentence changes the inventory", func(t *testing.T) {
		t.Parallel()
		before := niInv(t, niRoot(t, map[string]string{"skills/a.md": "# A skill\n\nThe guard SHALL print every normative sentence. It runs over the owned corpus\nonly. Reading it is cheap.\n"}))
		after := niInv(t, niRoot(t, map[string]string{"skills/a.md": "# A skill\n\nIt runs over the owned corpus only. Reading it is cheap.\n"}))
		if before != "The guard SHALL print every normative sentence." || strings.TrimSpace(after) != "" {
			t.Errorf("before=[%s] after=[%s]", before, after)
		}
	})

	oneLine := "It SHALL normalise each sentence's internal whitespace to single spaces and strip surrounding whitespace, so a reflowed paragraph does not read as a changed requirement.\n"
	t.Run("a reflowed sentence does not change the inventory", func(t *testing.T) {
		t.Parallel()
		a := niInv(t, niRoot(t, map[string]string{"rules/a.mdc": oneLine}))
		b := niInv(t, niRoot(t, map[string]string{"rules/a.mdc": "It SHALL normalise each sentence's internal   whitespace to single\n\tspaces and strip surrounding whitespace, so a reflowed\n  paragraph does not read as a changed requirement.\n"}))
		if a != b {
			t.Errorf("[%s] vs [%s]", a, b)
		}
	})

	t.Run("a reworded sentence changes the inventory", func(t *testing.T) {
		t.Parallel()
		a := niInv(t, niRoot(t, map[string]string{"rules/a.mdc": oneLine}))
		b := niInv(t, niRoot(t, map[string]string{"rules/a.mdc": "It SHALL emit each sentence's internal whitespace as single spaces.\n"}))
		if a == b {
			t.Errorf("unchanged: [%s]", a)
		}
	})

	orderA := map[string]string{"skills/zzz.md": "The inventory SHALL be sorted.\n", "rules/aaa.mdc": "The corpus MUST be resolved once.\n"}
	orderB := map[string]string{"skills/aaa.md": "The corpus MUST be resolved once.\n", "rules/zzz.mdc": "The inventory SHALL be sorted.\n"}
	t.Run("the inventory is independent of file order", func(t *testing.T) {
		t.Parallel()
		if a, b := niInv(t, niRoot(t, orderA)), niInv(t, niRoot(t, orderB)); a != b {
			t.Errorf("[%s] vs [%s]", a, b)
		}
	})
	t.Run("the inventory is sorted, not merely stable", func(t *testing.T) {
		t.Parallel()
		if got := niInv(t, niRoot(t, orderA)); got != "The corpus MUST be resolved once.\nThe inventory SHALL be sorted." {
			t.Errorf("[%s]", got)
		}
	})

	t.Run("an excluded tree contributes nothing", func(t *testing.T) {
		t.Parallel()
		base := niInv(t, niRoot(t, map[string]string{"skills/a.md": "The owned corpus SHALL be scanned.\n"}))
		plus := niInv(t, niRoot(t, map[string]string{
			"skills/a.md":                              "The owned corpus SHALL be scanned.\n",
			"skills/vendor/node_modules/pkg/a.md":      "A vendored package SHALL not be inventoried.\n",
			".superpowers/sdd/a.md":                    "A session record SHALL not be inventoried.\n",
			"docs/superpowers/skills/a.md":             "A vendored skill SHALL not be inventoried.\n",
			"spectre/changes/archive/kan-1/specs/a.md": "An archived spec SHALL not be inventoried.\n",
		}))
		if base != plus {
			t.Errorf("[%s]", plus)
		}
	})

	// A readable file in a SECOND scope root, so a guard that skipped the
	// unreadable one would print it and exit 0 rather than reach the
	// empty-corpus refusal for the wrong reason.
	t.Run("an unreadable scope root cannot be answered", func(t *testing.T) {
		t.Parallel()
		if os.Geteuid() == 0 {
			t.Skip("running as root reads it anyway")
		}
		root := niRoot(t, map[string]string{"skills/a.md": "The corpus SHALL be scanned.\n", "rules/b.mdc": "The scope roots MUST all be readable.\n"})
		skills := filepath.Join(root, "skills")
		if err := os.Chmod(skills, 0); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(skills, 0o755) })
		if rc := niRC(t, root); rc != 2 {
			t.Errorf("rc=%d, want 2", rc)
		}
	})

	t.Run("a missing scope root cannot be answered", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		mkdir(t, filepath.Join(root, "rules"))
		writeFile(t, filepath.Join(root, "skills", "a.md"), "The corpus SHALL be scanned.\n")
		if rc := niRC(t, root); rc != 2 {
			t.Errorf("rc=%d, want 2", rc)
		}
	})

	t.Run("a symlinked scope root cannot be answered", func(t *testing.T) {
		t.Parallel()
		root := niRoot(t, map[string]string{"rules/b.mdc": "The corpus SHALL be scanned.\n", "real-skills/a.md": "A symlinked scope root SHALL be refused.\n"})
		if err := os.RemoveAll(filepath.Join(root, "skills")); err != nil {
			t.Fatal(err)
		}
		symlink(t, filepath.Join(root, "real-skills"), filepath.Join(root, "skills"))
		if rc := niRC(t, root); rc != 2 {
			t.Errorf("rc=%d, want 2", rc)
		}
	})

	t.Run("a nested symlinked directory holding Markdown cannot be answered", func(t *testing.T) {
		t.Parallel()
		root := niRoot(t, map[string]string{"skills/top.md": "The visible corpus SHALL be scanned.\n", "outside/hidden/hidden.md": "The hidden corpus SHALL be scanned too.\n"})
		symlink(t, filepath.Join(root, "outside", "hidden"), filepath.Join(root, "skills", "nested-link"))
		if rc := niRC(t, root); rc != 2 {
			t.Errorf("rc=%d, want 2", rc)
		}
	})

	t.Run("a nested symlinked directory holding no Markdown is skipped, not refused", func(t *testing.T) {
		t.Parallel()
		root := niRoot(t, map[string]string{"skills/top.md": "The visible corpus SHALL be scanned.\n", "outside/lib/helper.sh": "echo hi\n"})
		symlink(t, filepath.Join(root, "outside", "lib"), filepath.Join(root, "skills", "nested-link"))
		if got := niInv(t, root); got != "The visible corpus SHALL be scanned." {
			t.Errorf("[%s]", got)
		}
	})

	t.Run("a nonexistent repository root cannot be answered", func(t *testing.T) {
		t.Parallel()
		if rc := niRC(t, filepath.Join(t.TempDir(), "absent")); rc != 2 {
			t.Errorf("rc=%d, want 2", rc)
		}
	})

	t.Run("a corpus with no Markdown at all cannot be answered", func(t *testing.T) {
		t.Parallel()
		if rc := niRC(t, niRoot(t, nil)); rc != 2 {
			t.Errorf("rc=%d, want 2", rc)
		}
	})

	t.Run("an empty override is refused", func(t *testing.T) {
		t.Parallel()
		rc, _, errOut := niRun(t, map[string]string{"CHECK_NORMATIVE_INVENTORY_ROOT": "", "FLOW_GUARD_REPO_ROOT": niRoot(t, map[string]string{"skills/a.md": "It MUST be scanned.\n"})})
		if rc != 2 || errOut != "CHECK_NORMATIVE_INVENTORY_ROOT is set but empty\n" {
			t.Errorf("rc=%d stderr=%q, want 2 and the set-but-empty refusal", rc, errOut)
		}
	})

	t.Run("each block shape yields its own sentence", func(t *testing.T) {
		t.Parallel()
		got := niInv(t, niRoot(t, map[string]string{"commands-claude/a.md": "## A heading that SHALL be its own block\n\n- A bullet that MUST stand alone.\n- Another bullet, with no obligation.\n\n| Rule | Detail |\n|------|--------|\n| A cell that MUST stand alone. | An adjacent cell with no obligation. |\n\n> A quoted line that SHALL be unwrapped\n> across two source lines.\n\n```bash\n# A fenced line that MUST be scanned.\n```\n"}))
		want := niSorted(
			"## A heading that SHALL be its own block",
			"# A fenced line that MUST be scanned.",
			"A bullet that MUST stand alone.",
			"A cell that MUST stand alone.",
			"A quoted line that SHALL be unwrapped across two source lines.",
		)
		if got != want {
			t.Errorf("got [%s] want [%s]", got, want)
		}
	})

	t.Run("a sentence restated in two files is printed twice", func(t *testing.T) {
		t.Parallel()
		s := "A restated sentence SHALL survive in both files.\n"
		if got := niInv(t, niRoot(t, map[string]string{"skills/a.md": s, "rules/a.mdc": s})); got != strings.TrimSuffix(s+s, "\n") {
			t.Errorf("[%s]", got)
		}
	})

	t.Run("only SHALL and MUST as whole words are keywords", func(t *testing.T) {
		t.Parallel()
		got := niInv(t, niRoot(t, map[string]string{"spectre/specs/a.md": "The guard SHOULD be fast. It MAY print a count.\n\nA caller MUST_NOT_TOUCH the table.\n\nThe exit code MUST be 2.\n"}))
		if got != "The exit code MUST be 2." {
			t.Errorf("[%s]", got)
		}
	})

	abbrevBefore := "The tool SHALL validate e.g. input for XSS and SQL injection before use."
	t.Run("a sentence is not truncated at an abbreviation", func(t *testing.T) {
		t.Parallel()
		if got := niInv(t, niRoot(t, map[string]string{"skills/a.md": abbrevBefore + "\n"})); got != abbrevBefore {
			t.Errorf("[%s]", got)
		}
	})
	t.Run("a substantive change after an abbreviation changes the inventory", func(t *testing.T) {
		t.Parallel()
		a := niInv(t, niRoot(t, map[string]string{"skills/a.md": abbrevBefore + "\n"}))
		b := niInv(t, niRoot(t, map[string]string{"skills/a.md": "The tool SHALL validate e.g. nothing at all here whatsoever.\n"}))
		if a == b {
			t.Errorf("unchanged: [%s]", a)
		}
	})

	t.Run("every exempt abbreviation keeps its sentence whole", func(t *testing.T) {
		t.Parallel()
		lines := []string{
			"A caller MUST pass a root, e.g. the repository it was asked to scan.",
			"A caller MUST pass one root, i.e. exactly one directory.",
			"The guard SHALL refuse a symlink, cf. the budget guard's own refusal.",
			"Bytes SHALL be counted, vs. lines, which a long-line file cheats.",
			"The corpus SHALL cover skills, rules, specs, etc. as one owned set.",
			"The finding SHALL cite Knuth et al. as its source.",
			"The report SHALL name Dr. Reviewer as its author.",
			"The row SHALL carry No. 7 as its identifier.",
		}
		got := niInv(t, niRoot(t, map[string]string{"rules/a.mdc": "- " + strings.Join(lines, "\n- ") + "\n"}))
		if want := niSorted(lines...); got != want {
			t.Errorf("got [%s] want [%s]", got, want)
		}
	})
}
