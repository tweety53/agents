package guard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// setFetchHead writes path as a full fetch of origin of the repository it
// sits in leaves it (empty when !ok, as a failed fetch leaves it) with a
// modification time age ago.
func setFetchHead(t *testing.T, path string, ok bool, age time.Duration) {
	t.Helper()
	body := ""
	if ok {
		repo, _, _ := strings.Cut(path, "/.git/")
		url, _ := core.Capture(exec.Command("git", "-C", repo, "remote", "get-url", "origin"))
		refs, _ := core.Capture(exec.Command("git", "-C", repo, "for-each-ref", "--format=%(refname:lstrip=3)", "refs/remotes/origin"))
		for _, b := range strings.Fields(refs) {
			body += "0000000000000000000000000000000000000000\tnot-for-merge\tbranch '" + b + "' of " + fetchURL(strings.TrimSpace(url)) + "\n"
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	guardtest.WriteFile(t, path, body)
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
		if code := Registry["resolve-base-branch"](c.args, core.Env{Dir: t.TempDir(), Getenv: os.Getenv}, &out, &errb); code != 0 || out.String() != "main\n" {
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
	env := core.Env{Getenv: os.Getenv}
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

// TestFetchedRecentlyNeedsAFullOriginFetch: a fresh FETCH_HEAD counts only
// when a fetch of origin's every branch wrote it -- a single-ref fetch or
// another remote's fetch leaves origin's other refs as old as they were.
func TestFetchedRecentlyNeedsAFullOriginFetch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	up, down := dir+"/up.git", dir+"/down"
	rbbGit(t, dir, "init", "-q", "-b", "main", dir+"/seed")
	rbbGit(t, dir+"/seed", "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "a")
	rbbGit(t, dir+"/seed", "branch", "b2")
	rbbGit(t, dir, "clone", "-q", "--bare", dir+"/seed", up)
	rbbGit(t, dir, "clone", "-q", up, down)
	rbbGit(t, down, "remote", "add", "other", dir+"/seed")
	git := core.EnvGit(core.Env{Getenv: os.Getenv})
	for _, c := range []struct {
		name  string
		fetch []string
		want  bool
	}{
		{"full fetch of origin counts", []string{"origin"}, true},
		{"single-ref fetch does not", []string{"origin", "main"}, false},
		{"another remote's fetch does not", []string{"other"}, false},
	} {
		rbbGit(t, down, append([]string{"fetch", "-q"}, c.fetch...)...)
		if got := fetchedRecently(git, down); got != c.want {
			t.Fatalf("%s: fetchedRecently %v, want %v", c.name, got, c.want)
		}
	}
}
