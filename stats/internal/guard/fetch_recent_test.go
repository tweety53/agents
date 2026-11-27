package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setFetchHead writes path non-empty (empty when !ok, as a failed fetch
// leaves it) with a modification time age ago.
func setFetchHead(t *testing.T, path string, ok bool, age time.Duration) {
	t.Helper()
	body := ""
	if ok {
		body = "0000000000000000000000000000000000000000\t\tbranch 'main' of origin\n"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, body)
	at := time.Now().Add(-age)
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatal(err)
	}
}

// fetchedSince reports whether path was rewritten -- a fetch ran -- after
// before.
func fetchedSince(t *testing.T, path string, before time.Time) bool {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return !fi.ModTime().Before(before)
}

// TestResolveBaseBranchSkipsAFetchUnder60s: a non-empty FETCH_HEAD younger
// than originFreshFor, the repository's or a linked worktree's, stands in for
// the guard's fetch; an older or empty one (a failed fetch) fetches, and
// --fetch always does.
func TestResolveBaseBranchSkipsAFetchUnder60s(t *testing.T) {
	t.Parallel()
	fx, err := rbbMasters()
	if err != nil {
		t.Fatal(err)
	}
	wt := rbbCopy(t, fx.wt)
	head := wt + "/.git/FETCH_HEAD"
	for _, c := range []struct {
		name      string
		args      []string
		ok        bool
		age       time.Duration
		linked    bool
		wantFetch bool
	}{
		{"fresh skips", []string{wt}, true, 10 * time.Second, false, false},
		{"fresh linked worktree skips", []string{wt}, true, 10 * time.Second, true, false},
		{"stale fetches", []string{wt}, true, 2 * time.Minute, false, true},
		{"failed fetch fetches", []string{wt}, false, 10 * time.Second, false, true},
		{"forced fetches", []string{"--fetch", wt}, true, 10 * time.Second, false, true},
	} {
		setFetchHead(t, head, c.ok, c.age)
		if c.linked {
			setFetchHead(t, head, true, 2*time.Minute)
			setFetchHead(t, wt+"/.git/worktrees/other/FETCH_HEAD", true, c.age)
		}
		before := time.Now().Add(-time.Second)
		var out, errb bytes.Buffer
		if code := Registry["resolve-base-branch"](c.args, Env{Dir: t.TempDir(), Getenv: os.Getenv}, &out, &errb); code != 0 || out.String() != "main\n" {
			t.Fatalf("%s: exit %d stdout %q stderr %q", c.name, code, out.String(), errb.String())
		}
		if got := fetchedSince(t, head, before); got != c.wantFetch {
			t.Fatalf("%s: fetched %v, want %v", c.name, got, c.wantFetch)
		}
		if err := os.RemoveAll(wt + "/.git/worktrees"); err != nil {
			t.Fatal(err)
		}
	}
}

// TestCheckBaseBehindSkipsAFetchUnder60s: check-base-behind skips its fetch
// after a recent one; sync-base-behind, which pushes the base, never does.
func TestCheckBaseBehindSkipsAFetchUnder60s(t *testing.T) {
	t.Parallel()
	fx, err := rbbMasters()
	if err != nil {
		t.Fatal(err)
	}
	wt := rbbCopy(t, fx.wt)
	rbbGit(t, wt, "config", "branch.spectre/fixture.flowBase", "feature/offline-mode")
	head := wt + "/.git/FETCH_HEAD"
	env := Env{Getenv: os.Getenv}
	for _, force := range []bool{false, true} {
		setFetchHead(t, head, true, 10*time.Second)
		before := time.Now().Add(-time.Second)
		var out, errb bytes.Buffer
		if code := baseBehind([]string{wt}, env, &out, &errb, force); code != 0 {
			t.Fatalf("force %v: exit %d stdout %q stderr %q", force, code, out.String(), errb.String())
		}
		if got := fetchedSince(t, head, before); got != force {
			t.Fatalf("force %v: fetched %v", force, got)
		}
	}
}
