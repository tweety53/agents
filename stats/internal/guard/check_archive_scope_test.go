package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-archive-scope.sh at c4f26c84, one subtest per case
// number, each pinning the guard's stdout whole where the harness globbed
// it. Case 5 is the mutation-bearing one: a bare string-prefix match would
// let `prefix-sibling/x` through when only `prefix` is allowed, so the
// path-component rule is asserted against exactly that pair.

// asRepo is the harness's new_repo plus its stage: a fresh repository with
// one base commit and each of paths staged -- a path written `-<path>` is
// committed in the base and its deletion staged, any other is created (with
// parents) and its addition staged. Its path is physical, as the guard's
// verdict names it.
func asRepo(t *testing.T, paths ...string) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var g fxGit
	g.git(dir, "init", "-q")
	var added []string
	for _, p := range paths {
		if d, ok := strings.CutPrefix(p, "-"); ok {
			if g.err == nil {
				g.err = os.MkdirAll(filepath.Dir(dir+"/"+d), 0o755)
			}
			g.write(dir+"/"+d, "x")
			g.git(dir, "add", d)
		} else {
			added = append(added, p)
		}
	}
	g.git(dir, "commit", "-q", "--allow-empty", "-m", "base")
	for _, p := range paths {
		if d, ok := strings.CutPrefix(p, "-"); ok {
			g.git(dir, "rm", "-q", d)
		}
	}
	for _, p := range added {
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

func TestCheckArchiveScope(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}

	for _, c := range []struct {
		name     string
		staged   []string
		prefixes []string
		rc       int
		want     string // REPO stands for the repository's physical path
	}{
		{"case 1: a staged deletion inside the prefix is SCOPE-OK naming the worktree",
			[]string{"-spectre/changes/kan-x/narrative.md"}, []string{"spectre/changes/kan-x"}, 0,
			"SCOPE-OK: REPO\n"},
		{"case 2: a staged path outside the prefix is named verbatim and counted 1",
			[]string{"stats/web/src/App.tsx"}, []string{"spectre/changes/kan-x"}, 1,
			"OUT-OF-SCOPE: stats/web/src/App.tsx\nSCOPE-VIOLATION: REPO — 1\n"},
		// Two offenders and one in-scope deletion: both offenders on lines of
		// their own in the diff's order, the in-scope one never, count 2.
		{"case 3: mixed staged paths name exactly the two offenders",
			[]string{"-spectre/changes/kan-x/narrative.md", "stats/web/src/App.tsx", "stats/go.mod"},
			[]string{"spectre/changes/kan-x"}, 1,
			"OUT-OF-SCOPE: stats/go.mod\nOUT-OF-SCOPE: stats/web/src/App.tsx\nSCOPE-VIOLATION: REPO — 2\n"},
		{"case 4: an empty staged diff is SCOPE-OK", nil, []string{"spectre/changes/kan-x"}, 0,
			"SCOPE-OK: REPO\n"},
		// THE PATH-COMPONENT RULE: a prefix given without its trailing slash is
		// normalized to one, so `prefix` allows `prefix/ok` but never lets
		// `prefix-sibling/bad` through.
		{"case 5: a string-prefix sibling is the only offender",
			[]string{"-prefix/ok.txt", "-prefix-sibling/bad.txt"}, []string{"prefix"}, 1,
			"OUT-OF-SCOPE: prefix-sibling/bad.txt\nSCOPE-VIOLATION: REPO — 1\n"},
		{"case 6: a deletion matching a later prefix is in scope",
			[]string{"-spectre/changes/kan-x-fix-1/tasks.md"}, []string{"spectre/changes/kan-x", "spectre/changes/kan-x-fix-1"}, 0,
			"SCOPE-OK: REPO\n"},
		// Beyond the harness: `${prefix%/}/` strips one slash, so a doubled
		// one still demands it — nothing staged sits under `stats//`.
		{"beyond the harness: one trailing slash is stripped, not all",
			[]string{"-stats/go.mod"}, []string{"stats//"}, 1,
			"OUT-OF-SCOPE: stats/go.mod\nSCOPE-VIOLATION: REPO — 1\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			repo := asRepo(t, c.staged...)
			r := runGuard("check-archive-scope", append([]string{repo}, c.prefixes...), env)
			if want := strings.ReplaceAll(c.want, "REPO", repo); r.rc != c.rc || r.stdout != want {
				t.Fatalf("rc=%d stdout=%q stderr=%q; want %d and %q", r.rc, r.stdout, r.err, c.rc, want)
			}
		})
	}

	// Case 7: cannot answer — exit 2, and no verdict line on stdout. 7a passes
	// a real second argument so the arity check cannot answer for it — the
	// non-directory branch is what exits 2.
	for _, c := range []struct {
		name, err string
		args      func(t *testing.T) []string
	}{
		{"case 7a: a non-directory worktree exits 2", "is not a directory", func(t *testing.T) []string {
			return []string{t.TempDir() + "/missing", "spectre/changes/"}
		}},
		{"case 7b: a non-git directory exits 2", "is not a git worktree", func(t *testing.T) []string {
			return []string{t.TempDir(), "spectre/changes/"}
		}},
		{"case 7c: no allowed prefix exits 2", "usage: check-archive-scope.sh <worktree> <allowed-prefix> [<allowed-prefix> ...]",
			func(t *testing.T) []string { return []string{asRepo(t)} }},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := runGuard("check-archive-scope", c.args(t), env)
			if r.rc != 2 || r.stdout != "" || !strings.HasPrefix(r.err, "check-archive-scope: ") || !strings.Contains(r.err, c.err) {
				t.Fatalf("rc=%d stdout=%q stderr=%q; want 2, empty stdout, %q on stderr", r.rc, r.stdout, r.err, c.err)
			}
		})
	}
}

// TestCheckArchiveScopeAdmitsOnlyTheDrop pins the landing commit's scope: a
// staged deletion under the change's directories passes; a staged addition
// -- even inside them -- or a deletion outside them is refused.
func TestCheckArchiveScopeAdmitsOnlyTheDrop(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	prefixes := []string{"spectre/changes/demo", "spectre/changes/demo-fix-1"}
	for _, c := range []struct {
		name   string
		staged []string
		rc     int
		want   string
	}{
		{"deletions under the change's directories", []string{"-spectre/changes/demo/tasks.md", "-spectre/changes/demo-fix-1/tasks.md"}, 0, "SCOPE-OK: REPO\n"},
		{"an addition inside the change's directory", []string{"spectre/changes/demo/ledger.md"}, 1, "OUT-OF-SCOPE: spectre/changes/demo/ledger.md\nSCOPE-VIOLATION: REPO — 1\n"},
		{"an addition under the archive", []string{"spectre/changes/archive/demo/tasks.md"}, 1, "OUT-OF-SCOPE: spectre/changes/archive/demo/tasks.md\nSCOPE-VIOLATION: REPO — 1\n"},
		{"a deletion outside the change's directories", []string{"-spectre/changes/other/tasks.md"}, 1, "OUT-OF-SCOPE: spectre/changes/other/tasks.md\nSCOPE-VIOLATION: REPO — 1\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			repo := asRepo(t, c.staged...)
			r := runGuard("check-archive-scope", append([]string{repo}, prefixes...), env)
			if want := strings.ReplaceAll(c.want, "REPO", repo); r.rc != c.rc || r.stdout != want {
				t.Fatalf("rc=%d stdout=%q stderr=%q; want %d and %q", r.rc, r.stdout, r.err, c.rc, want)
			}
		})
	}
}
