package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

// Every case of scripts/test-land-self-review-report.sh at 9cd35da8, one
// subtest per ok: label, plus pins the harness never asserted: every LAND-*
// line and the usage line byte for byte with its stream, the third branch
// re-assertion (before the pull/push pair), git's own exit code passed
// through, and the caller's collation in the refusal's path lists — each
// expected value what the bash printed at 9cd35da8.
//
// The guard runs in-process over a copy of one template repository per case.
// Every remote it pulls from or pushes to is a bare repository in the case's
// own t.TempDir(). The harness's rejected-push case pushed into a non-bare
// clone with main checked out; here the bare remote refuses through a
// pre-receive hook instead, named by the remote's own core.hooksPath so the
// operator's global config cannot replace it — the refusal the case pins is
// the same: a push git rejects. The harness's PATH git shims (the mid-run
// branch switch, the race between the staged-set check and the commit) are
// the same bash shims, put first on the guard's PATH.

// lsrrLocked serialises writes: git's stdout and stderr are copied by two
// goroutines, and both land in one merged buffer.
type lsrrLocked struct {
	mu *sync.Mutex
	w  io.Writer
}

func (l lsrrLocked) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

func TestLandSelfReviewReport(t *testing.T) {
	t.Parallel()
	// new_repo, built once and copied per case: main carrying one empty base
	// commit, with a repo-local identity so the guard's own commit needs
	// nothing from the operator's config.
	tmpl := t.TempDir() + "/repo"
	var tg fxGit
	tg.git("", "init", "-q", "-b", "main", tmpl)
	for _, kv := range [][2]string{{"user.name", "land-test"}, {"user.email", "land-test@example.com"}, {"commit.gpgsign", "false"}} {
		tg.git(tmpl, "config", kv[0], kv[1])
	}
	tg.git(tmpl, "commit", "-q", "--allow-empty", "-m", "base")
	if tg.err != nil {
		t.Fatal(tg.err)
	}

	type lsrrFx struct {
		dir, repo string
		g         fxGit
		env       Env
	}
	newFx := func(t *testing.T) *lsrrFx {
		fx := &lsrrFx{dir: t.TempDir()}
		fx.repo = fx.dir + "/repo"
		gdcCopyTree(t, tmpl, fx.repo)
		fx.env = Env{Dir: fx.dir, Getenv: os.Getenv, LookupEnv: hermeticGitLookup}
		return fx
	}
	// writeReport is write_report: an untracked report beside a tracked
	// context bundle, the mid-flight state both call sites start from.
	writeReport := func(t *testing.T, fx *lsrrFx, name string) {
		mkdir(t, fx.repo+"/docs/self-review")
		fx.g.write(fx.repo+"/docs/self-review/"+name+"-self-review.md", "report")
		fx.g.write(fx.repo+"/docs/self-review/"+name+"-context.md", "bundle")
		fx.g.git(fx.repo, "add", "docs/self-review/"+name+"-context.md")
		fx.g.git(fx.repo, "commit", "-q", "-m", "docs(self-review): "+name+" context bundle")
	}
	stageStray := func(fx *lsrrFx, f string) {
		fx.g.write(fx.repo+"/"+f, "stray")
		fx.g.git(fx.repo, "add", f)
	}
	// bareOrigin adds a bare origin in the case's directory and pushes main.
	bareOrigin := func(fx *lsrrFx) string {
		origin := fx.dir + "/origin.git"
		fx.g.git("", "init", "-q", "--bare", "-b", "main", origin)
		fx.g.git(fx.repo, "remote", "add", "origin", origin)
		fx.g.git(fx.repo, "push", "-q", "-u", "origin", "main")
		return origin
	}
	run := func(t *testing.T, fx *lsrrFx, args ...string) guardResult {
		t.Helper()
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		fn := Registry["land-self-review-report"]
		if fn == nil {
			return guardResult{rc: -1, out: "land-self-review-report is not registered"}
		}
		var mu sync.Mutex
		var all, o, e bytes.Buffer
		rc := fn(args, fx.env, lsrrLocked{&mu, io.MultiWriter(&o, &all)}, lsrrLocked{&mu, io.MultiWriter(&e, &all)})
		return guardResult{rc, strings.TrimRight(all.String(), "\n"), o.String(), e.String()}
	}
	git := func(t *testing.T, fx *lsrrFx, args ...string) string {
		t.Helper()
		var g fxGit
		out := g.git(fx.repo, args...)
		if g.err != nil {
			t.Fatal(g.err)
		}
		return out
	}
	head := func(t *testing.T, fx *lsrrFx) string { return git(t, fx, "rev-parse", "HEAD") }
	subject := func(t *testing.T, fx *lsrrFx) string { return git(t, fx, "log", "-1", "--format=%s") }
	check := func(t *testing.T, ok bool, r guardResult, what string) {
		t.Helper()
		if !ok {
			t.Fatalf("%s: rc=%d out=%s", what, r.rc, r.out)
		}
	}
	// shim writes <dir>/git: when the chain's verb (the word after -C <repo>)
	// is verb, run action once, then exec the real git — the harness's PATH
	// shims, first on the guard's PATH.
	shim := func(t *testing.T, fx *lsrrFx, verb, action string) {
		writeExec(t, fx.dir+"/shim/git", fmt.Sprintf(`#!/bin/bash
FIRST="$1"
[ "$FIRST" = -C ] && FIRST="$3"
if [ "$FIRST" = %s ] && [ ! -f %q ]; then
  touch %q
  %s
fi
exec %q "$@"
`, verb, fx.dir+"/shim/fired", fx.dir+"/shim/fired", action, fixtureGit))
		fx.env.Getenv = pathEnv(fx.dir + "/shim")
	}
	const rep = "docs/self-review/%s-self-review.md"
	const ctx = "docs/self-review/%s-context.md"
	subj := func(n string) string { return "docs(self-review): " + n + " self-review report" }

	// Case 3: a branch mismatch runs nothing at all.
	mismatch := func(t *testing.T) (*lsrrFx, string, guardResult) {
		fx := newFx(t)
		writeReport(t, fx, "kan-x")
		base := head(t, fx)
		return fx, base, run(t, fx, fx.repo, "spectre/kan-x", subj("kan-x"), fmt.Sprintf(rep, "kan-x"))
	}
	// Case 4: the site-1 landing, pushed to a bare origin.
	site1 := func(t *testing.T) (*lsrrFx, guardResult) {
		fx := newFx(t)
		writeReport(t, fx, "kan-y")
		bareOrigin(fx)
		return fx, run(t, fx, fx.repo, "main", subj("kan-y"), fmt.Sprintf(rep, "kan-y"), fmt.Sprintf(ctx, "kan-y"), "--push", "main")
	}
	// Case 5: the archive shape — no rm path, no --push, no remote.
	archive := func(t *testing.T) (*lsrrFx, guardResult) {
		fx := newFx(t)
		writeReport(t, fx, "kan-z")
		return fx, run(t, fx, fx.repo, "main", subj("kan-z"), fmt.Sprintf(rep, "kan-z"))
	}
	// Case 6: nothing staged is a clean no-op.
	empty := func(t *testing.T) (*lsrrFx, string, guardResult) {
		fx := newFx(t)
		writeReport(t, fx, "kan-w")
		fx.g.git(fx.repo, "add", fmt.Sprintf(rep, "kan-w"))
		fx.g.git(fx.repo, "commit", "-q", "-m", subj("kan-w"))
		base := head(t, fx)
		return fx, base, run(t, fx, fx.repo, "main", subj("kan-w"), fmt.Sprintf(rep, "kan-w"))
	}
	// Case 7: a rejected push leaves the commit local.
	rejected := func(t *testing.T) (*lsrrFx, guardResult) {
		fx := newFx(t)
		writeReport(t, fx, "kan-v")
		origin := bareOrigin(fx)
		fx.g.git(origin, "config", "core.hooksPath", "hooks")
		writeExec(t, origin+"/hooks/pre-receive", "#!/bin/sh\necho 'rejected by the test remote' >&2\nexit 1\n")
		return fx, run(t, fx, fx.repo, "main", subj("kan-v"), fmt.Sprintf(rep, "kan-v"), fmt.Sprintf(ctx, "kan-v"), "--push", "main")
	}
	// Case 8: foreign staged work refuses the commit.
	foreign := func(t *testing.T) (*lsrrFx, string, guardResult) {
		fx := newFx(t)
		writeReport(t, fx, "kan-f")
		stageStray(fx, "stray.txt")
		base := head(t, fx)
		return fx, base, run(t, fx, fx.repo, "main", subj("kan-f"), fmt.Sprintf(rep, "kan-f"))
	}
	// Case 9: the same refusal under an rm path.
	foreignRm := func(t *testing.T) (*lsrrFx, string, guardResult) {
		fx := newFx(t)
		writeReport(t, fx, "kan-r")
		stageStray(fx, "stray.txt")
		base := head(t, fx)
		return fx, base, run(t, fx, fx.repo, "main", subj("kan-r"), fmt.Sprintf(rep, "kan-r"), fmt.Sprintf(ctx, "kan-r"))
	}
	// Case 10: the branch switched on the chain's first diff, between the
	// add and the commit, is re-caught before the commit.
	switched := func(t *testing.T) (*lsrrFx, string, guardResult) {
		fx := newFx(t)
		writeReport(t, fx, "kan-s")
		shim(t, fx, "diff", fmt.Sprintf("%q -C %q checkout -b other >/dev/null 2>&1", fixtureGit, fx.repo))
		base := head(t, fx)
		return fx, base, run(t, fx, fx.repo, "main", subj("kan-s"), fmt.Sprintf(rep, "kan-s"))
	}
	// Case 11: foreign content staged between the staged-set check and the
	// commit stays out of the commit.
	race := func(t *testing.T) (*lsrrFx, guardResult) {
		fx := newFx(t)
		writeReport(t, fx, "kan-q")
		shim(t, fx, "commit", fmt.Sprintf("printf 'race\\n' > %q; %q -C %q add race.txt", fx.repo+"/race.txt", fixtureGit, fx.repo))
		return fx, run(t, fx, fx.repo, "main", subj("kan-q"), fmt.Sprintf(rep, "kan-q"))
	}
	// Case 12: a vanished own path is named under "missing from the index".
	missing := func(t *testing.T) guardResult {
		fx := newFx(t)
		writeReport(t, fx, "kan-m")
		fx.g.git(fx.repo, "add", fmt.Sprintf(rep, "kan-m"))
		fx.g.git(fx.repo, "commit", "-q", "-m", subj("kan-m"))
		stageStray(fx, "stray.txt")
		return run(t, fx, fx.repo, "main", subj("kan-m"), fmt.Sprintf(rep, "kan-m"))
	}
	// Case 13: a staged foreign deletion that pairs as a rename with the
	// chain's own new file under git's default detection.
	renamePair := func(t *testing.T) (*lsrrFx, string, guardResult) {
		fx := newFx(t)
		mkdir(t, fx.repo+"/docs/self-review")
		fx.g.write(fx.repo+"/docs/self-review/old.md", "report content line\ntwo")
		fx.g.git(fx.repo, "add", "docs/self-review/old.md")
		fx.g.git(fx.repo, "commit", "-q", "-m", "old")
		fx.g.git(fx.repo, "rm", "-q", "docs/self-review/old.md")
		mkdir(t, fx.repo+"/docs/self-review")
		fx.g.write(fx.repo+"/docs/self-review/kan-p-self-review.md", "report content line\ntwo")
		base := head(t, fx)
		return fx, base, run(t, fx, fx.repo, "main", subj("kan-p"), fmt.Sprintf(rep, "kan-p"))
	}
	const usage = "usage: land-self-review-report.sh <repo> <branch> <subject> <add-path> [<rm-path>] [--push <base>]\n"
	usageIs := func(t *testing.T, r guardResult) {
		t.Helper()
		if r.rc != 2 || r.stdout != "" || r.err != usage {
			t.Fatalf("rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err)
		}
	}

	for _, tc := range []struct {
		name string
		fn   func(t *testing.T)
	}{
		// 1-2. Usage defects.
		{"test_land_self_review_report_usage: too many positionals exits 2", func(t *testing.T) {
			fx := newFx(t)
			r := run(t, fx, fx.repo, "main", "subject", fmt.Sprintf(rep, "n"), fmt.Sprintf(ctx, "n"), "extra")
			check(t, r.rc == 2, r, "rc")
		}},
		{"test_land_self_review_report_usage: the usage line is printed", func(t *testing.T) {
			fx := newFx(t)
			usageIs(t, run(t, fx, fx.repo, "main", "subject", fmt.Sprintf(rep, "n"), fmt.Sprintf(ctx, "n"), "extra"))
		}},
		{"test_land_self_review_report_usage: --push without a base exits 2", func(t *testing.T) {
			fx := newFx(t)
			usageIs(t, run(t, fx, fx.repo, "main", "subject", fmt.Sprintf(rep, "n"), "--push"))
		}},
		{"test_land_self_review_report_usage: too few arguments exits 2", func(t *testing.T) {
			fx := newFx(t)
			usageIs(t, run(t, fx, fx.repo, "main", "subject"))
		}},

		// 3. A branch mismatch runs nothing at all.
		{"test_land_branch_mismatch_stops_everything: exits 1", func(t *testing.T) {
			_, _, r := mismatch(t)
			check(t, r.rc == 1, r, "rc")
		}},
		{"test_land_branch_mismatch_stops_everything: the mismatch names both branches", func(t *testing.T) {
			_, _, r := mismatch(t)
			check(t, r.stdout == "" && r.err == "LAND-BRANCH-MISMATCH: expected spectre/kan-x, found main — nothing added, committed, pulled or pushed\n", r, "line")
		}},
		{"test_land_branch_mismatch_stops_everything: nothing was staged", func(t *testing.T) {
			fx, _, r := mismatch(t)
			check(t, git(t, fx, "diff", "--cached", "--name-only") == "", r, "index dirty")
		}},
		{"test_land_branch_mismatch_stops_everything: no commit was made", func(t *testing.T) {
			fx, base, r := mismatch(t)
			check(t, head(t, fx) == base, r, "HEAD moved")
		}},
		{"test_land_branch_mismatch_stops_everything: the report is still untracked", func(t *testing.T) {
			fx, _, r := mismatch(t)
			check(t, git(t, fx, "status", "--porcelain", fmt.Sprintf(rep, "kan-x")) == "?? "+fmt.Sprintf(rep, "kan-x"), r, "report staged")
		}},

		// 4. The site-1 landing.
		{"test_land_report_lands_and_pushes: exits 0", func(t *testing.T) {
			_, r := site1(t)
			check(t, r.rc == 0, r, "rc")
		}},
		{"test_land_report_lands_and_pushes: the commit subject is exact", func(t *testing.T) {
			fx, r := site1(t)
			check(t, subject(t, fx) == subj("kan-y"), r, "subject "+subject(t, fx))
		}},
		{"test_land_report_lands_and_pushes: origin/main carries the commit", func(t *testing.T) {
			fx, r := site1(t)
			check(t, git(t, fx, "rev-parse", "origin/main") == head(t, fx), r, "origin/main behind")
			var g fxGit
			check(t, g.git(fx.dir+"/origin.git", "rev-parse", "main") == head(t, fx), r, "the remote itself is behind")
		}},
		{"test_land_report_lands_and_pushes: the context bundle was removed", func(t *testing.T) {
			fx, r := site1(t)
			_, err := os.Stat(fx.repo + "/" + fmt.Sprintf(ctx, "kan-y"))
			check(t, os.IsNotExist(err), r, "bundle still on disk")
		}},
		{"test_land_report_lands_and_pushes: the report is on disk", func(t *testing.T) {
			fx, r := site1(t)
			_, err := os.Stat(fx.repo + "/" + fmt.Sprintf(rep, "kan-y"))
			check(t, err == nil, r, "report missing")
		}},

		// 5. The archive shape.
		{"test_land_report_lands_and_pushes: the archive shape exits 0 with no remote", func(t *testing.T) {
			_, r := archive(t)
			check(t, r.rc == 0, r, "rc")
		}},
		{"test_land_report_lands_and_pushes: the archive shape commits", func(t *testing.T) {
			fx, r := archive(t)
			check(t, subject(t, fx) == subj("kan-z"), r, "subject "+subject(t, fx))
		}},

		// 6. Nothing staged.
		{"test_land_bundle_skips_empty_commit: an empty staged diff exits 0", func(t *testing.T) {
			_, _, r := empty(t)
			check(t, r.rc == 0, r, "rc")
		}},
		{"test_land_bundle_skips_empty_commit: the no-op is named", func(t *testing.T) {
			fx, _, r := empty(t)
			check(t, r.err == "" && r.stdout == "LAND-NOTHING-TO-COMMIT: nothing staged under "+fx.repo+" — no commit, no pull, no push\n", r, "line")
		}},
		{"test_land_bundle_skips_empty_commit: no empty commit was made", func(t *testing.T) {
			fx, base, r := empty(t)
			check(t, head(t, fx) == base, r, "HEAD moved")
		}},

		// 7. A rejected push.
		{"test_land_report_lands_and_pushes: a rejected push exits non-zero", func(t *testing.T) {
			_, r := rejected(t)
			check(t, r.rc == 1 && strings.Contains(r.err, "rejected by the test remote"), r, "rc")
		}},
		{"test_land_report_lands_and_pushes: the commit stays local on a rejected push", func(t *testing.T) {
			fx, r := rejected(t)
			check(t, subject(t, fx) == subj("kan-v"), r, "subject "+subject(t, fx))
			check(t, git(t, fx, "rev-parse", "origin/main") != head(t, fx), r, "origin/main moved")
		}},

		// 8. Foreign staged work refuses the commit.
		{"test_land_foreign_staged_refuses: exits 3", func(t *testing.T) {
			_, _, r := foreign(t)
			check(t, r.rc == 3, r, "rc")
		}},
		{"test_land_foreign_staged_refuses: the refusal is named", func(t *testing.T) {
			_, _, r := foreign(t)
			check(t, r.stdout == "" && r.err == "LAND-FOREIGN-STAGED: expected only docs/self-review/kan-f-self-review.md— foreign staged: stray.txt — missing from the index: — nothing committed, pulled or pushed; clear the staging or land from a clean checkout\n", r, "line")
		}},
		{"test_land_foreign_staged_refuses: the refusal names the foreign path", func(t *testing.T) {
			_, _, r := foreign(t)
			check(t, strings.Contains(r.err, "foreign staged: stray.txt "), r, "foreign path")
		}},
		{"test_land_foreign_staged_refuses: no commit was made", func(t *testing.T) {
			fx, base, r := foreign(t)
			check(t, head(t, fx) == base, r, "HEAD moved")
		}},
		{"test_land_foreign_staged_refuses: the foreign path is still staged", func(t *testing.T) {
			fx, _, r := foreign(t)
			check(t, strings.Contains(git(t, fx, "diff", "--cached", "--name-only"), "stray.txt"), r, "foreign path left the index")
		}},
		{"test_land_foreign_staged_refuses: the report is still staged", func(t *testing.T) {
			fx, _, r := foreign(t)
			check(t, strings.Contains(git(t, fx, "diff", "--cached", "--name-only"), "kan-f-self-review.md"), r, "report left the index")
		}},

		// 9. The same refusal under an rm path.
		{"test_land_foreign_staged_with_rm_refuses: exits 3", func(t *testing.T) {
			_, _, r := foreignRm(t)
			check(t, r.rc == 3, r, "rc")
		}},
		{"test_land_foreign_staged_with_rm_refuses: the refusal is named", func(t *testing.T) {
			_, _, r := foreignRm(t)
			check(t, r.stdout == "rm 'docs/self-review/kan-r-context.md'\n" &&
				r.err == "LAND-FOREIGN-STAGED: expected only docs/self-review/kan-r-context.md docs/self-review/kan-r-self-review.md— foreign staged: stray.txt — missing from the index: — nothing committed, pulled or pushed; clear the staging or land from a clean checkout\n", r, "line")
		}},
		{"test_land_foreign_staged_with_rm_refuses: no commit was made", func(t *testing.T) {
			fx, base, r := foreignRm(t)
			check(t, head(t, fx) == base, r, "HEAD moved")
		}},
		{"test_land_foreign_staged_with_rm_refuses: HEAD still carries the context bundle", func(t *testing.T) {
			fx, base, r := foreignRm(t)
			var g fxGit
			g.git(fx.repo, "cat-file", "-e", base+":"+fmt.Sprintf(ctx, "kan-r"))
			check(t, g.err == nil, r, "bundle gone from HEAD")
		}},

		// 10. A branch switched mid-run.
		{"test_land_branch_switched_midrun_refuses: exits 1", func(t *testing.T) {
			_, _, r := switched(t)
			check(t, r.rc == 1, r, "rc")
		}},
		{"test_land_branch_switched_midrun_refuses: the re-assert names both branches", func(t *testing.T) {
			_, _, r := switched(t)
			check(t, r.stdout == "" && r.err == "LAND-BRANCH-MISMATCH: expected main, found other before the commit — nothing committed, pulled or pushed\n", r, "line")
		}},
		{"test_land_branch_switched_midrun_refuses: no commit was made", func(t *testing.T) {
			fx, base, r := switched(t)
			check(t, head(t, fx) == base, r, "HEAD moved")
		}},

		// 11. The commit is pathspec-limited.
		{"test_land_commit_pathspec_limited: exits 0", func(t *testing.T) {
			_, r := race(t)
			check(t, r.rc == 0, r, "rc")
		}},
		{"test_land_commit_pathspec_limited: the commit carries only the chain's own path", func(t *testing.T) {
			fx, r := race(t)
			check(t, git(t, fx, "show", "--name-only", "--format=", "--no-renames", "HEAD") == fmt.Sprintf(rep, "kan-q"), r, "commit content")
		}},
		{"test_land_commit_pathspec_limited: the race file is left staged, uncommitted", func(t *testing.T) {
			fx, r := race(t)
			check(t, git(t, fx, "diff", "--cached", "--name-only", "--no-renames") == "race.txt", r, "race file left the index")
		}},

		// 12. A vanished own path is named.
		{"test_land_missing_own_path_named: exits 3", func(t *testing.T) {
			r := missing(t)
			check(t, r.rc == 3, r, "rc")
		}},
		{"test_land_missing_own_path_named: the vanished own path is named", func(t *testing.T) {
			r := missing(t)
			check(t, r.err == "LAND-FOREIGN-STAGED: expected only docs/self-review/kan-m-self-review.md— foreign staged: stray.txt — missing from the index: docs/self-review/kan-m-self-review.md — nothing committed, pulled or pushed; clear the staging or land from a clean checkout\n", r, "line")
		}},
		{"test_land_missing_own_path_named: the foreign path is still named", func(t *testing.T) {
			r := missing(t)
			check(t, strings.Contains(r.err, "foreign staged: stray.txt "), r, "foreign path")
		}},

		// 13. The --no-renames pin is load-bearing.
		{"test_land_rename_pair_refuses: exits 3", func(t *testing.T) {
			_, _, r := renamePair(t)
			check(t, r.rc == 3, r, "rc")
		}},
		{"test_land_rename_pair_refuses: the refusal names the deleted foreign path", func(t *testing.T) {
			_, _, r := renamePair(t)
			check(t, strings.Contains(r.err, "foreign staged: docs/self-review/old.md "), r, "deleted path")
		}},
		{"test_land_rename_pair_refuses: HEAD still carries the foreign file", func(t *testing.T) {
			fx, base, r := renamePair(t)
			var g fxGit
			g.git(fx.repo, "cat-file", "-e", base+":docs/self-review/old.md")
			check(t, g.err == nil, r, "deletion swept into a commit")
		}},

		// Beyond the harness.
		// A first --push base that won would fail the pull: origin has no
		// branch other.
		{"pin: --push with an empty base, and a second --push's base wins", func(t *testing.T) {
			fx := newFx(t)
			usageIs(t, run(t, fx, fx.repo, "main", "s", fmt.Sprintf(rep, "n"), "--push", ""))
			writeReport(t, fx, "kan-u")
			bareOrigin(fx)
			r := run(t, fx, fx.repo, "main", subj("kan-u"), fmt.Sprintf(rep, "kan-u"), "--push", "other", "--push", "main")
			check(t, r.rc == 0 && git(t, fx, "rev-parse", "origin/main") == head(t, fx), r, "main not pushed")
		}},
		// The third re-assertion: a branch switched by the commit itself
		// stops the chain before the pull/push pair — the commit stays,
		// nothing reaches the remote.
		{"pin: a branch switched before the pull/push pushes nothing", func(t *testing.T) {
			fx := newFx(t)
			writeReport(t, fx, "kan-t")
			bareOrigin(fx)
			base := head(t, fx)
			shim(t, fx, "commit", fmt.Sprintf("%q \"$@\" || exit; exec %q -C %q checkout -q -b other", fixtureGit, fixtureGit, fx.repo))
			r := run(t, fx, fx.repo, "main", subj("kan-t"), fmt.Sprintf(rep, "kan-t"), "--push", "main")
			check(t, r.rc == 1 && strings.HasSuffix(r.err, "LAND-BRANCH-MISMATCH: expected main, found other before the pull/push — nothing pulled or pushed\n"), r, "line")
			check(t, subject(t, fx) == subj("kan-t"), r, "commit missing")
			var g fxGit
			check(t, g.git(fx.dir+"/origin.git", "rev-parse", "main") == base, r, "the remote moved")
		}},
		// git's own exit code, unmasked: a failing add (a path git cannot
		// match) and a failing branch read (not a repository) each exit with
		// git's status and git's message.
		{"pin: git's own exit code is passed through", func(t *testing.T) {
			fx := newFx(t)
			r := run(t, fx, fx.repo, "main", "s", "no-such-file.md")
			check(t, r.rc == 128 && r.stdout == "" && strings.HasPrefix(r.err, "fatal: pathspec 'no-such-file.md' did not match any files"), r, "add")
			mkdir(t, fx.dir+"/plain")
			r = run(t, fx, "plain", "main", "s", "x.md")
			check(t, r.rc == 128 && strings.HasPrefix(r.err, "fatal: not a git repository"), r, "branch read")
		}},
		// A detached HEAD reads as an empty branch name.
		{"pin: a detached HEAD is a mismatch naming no branch", func(t *testing.T) {
			fx := newFx(t)
			fx.g.git(fx.repo, "checkout", "-q", "--detach")
			r := run(t, fx, "repo", "main", "s", "x.md")
			check(t, r.rc == 1 && r.err == "LAND-BRANCH-MISMATCH: expected main, found  — nothing added, committed, pulled or pushed\n", r, "line")
		}},
		// Locale-sensitive ordering: the refusal's path lists follow the
		// caller's collation, as the bash's sort and comm did.
		{"pin: the refusal's path lists follow the caller's collation", func(t *testing.T) {
			for _, c := range []struct{ lcAll, want string }{
				{"C", "LAND-FOREIGN-STAGED: expected only docs/C.md docs/b.md— foreign staged: B.txt a.txt — missing from the index: — nothing committed, pulled or pushed; clear the staging or land from a clean checkout\n"},
				{"en_US.UTF-8", "LAND-FOREIGN-STAGED: expected only docs/b.md docs/C.md— foreign staged: a.txt B.txt — missing from the index: — nothing committed, pulled or pushed; clear the staging or land from a clean checkout\n"},
			} {
				fx := newFx(t)
				mkdir(t, fx.repo+"/docs")
				fx.g.write(fx.repo+"/docs/C.md", "c")
				fx.g.git(fx.repo, "add", "docs/C.md")
				fx.g.git(fx.repo, "commit", "-qm", "C")
				fx.g.write(fx.repo+"/docs/b.md", "b")
				stageStray(fx, "a.txt")
				stageStray(fx, "B.txt")
				fx.env = envWith(t, []string{"LC_ALL=" + c.lcAll, "GIT_CONFIG_GLOBAL=/dev/null"})
				fx.env.Dir = fx.dir
				r := run(t, fx, "repo", "main", "s", "docs/b.md", "docs/C.md")
				if r.rc != 3 || r.err != c.want {
					t.Errorf("LC_ALL=%s: rc=%d stderr=%q", c.lcAll, r.rc, r.err)
				}
			}
		}},
		// No sort on PATH: the bash's `$(g diff … | sort)` exited with the
		// shell's 127 under pipefail; the port does too.
		{"pin: no sort on PATH exits 127", func(t *testing.T) {
			fx := newFx(t)
			writeReport(t, fx, "kan-s")
			bin := fx.dir + "/gitonly"
			mkdir(t, bin)
			if err := os.Symlink(fixtureGit, bin+"/git"); err != nil {
				t.Fatal(err)
			}
			fx.env.Getenv = func(k string) string {
				if k == "PATH" {
					return bin
				}
				return os.Getenv(k)
			}
			r := run(t, fx, fx.repo, "main", subj("kan-s"), fmt.Sprintf(rep, "kan-s"))
			if r.rc != 127 || !strings.Contains(r.err, "land-self-review-report: sort: command not found") {
				t.Errorf("rc=%d stderr=%q", r.rc, r.err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t)
		})
	}
}

// A finding line whose cause was "not traced" is an open item: the report is
// refused with exit 4 before anything is staged, and lands once traced.
func TestLandSelfReviewReportRefusesUntracedFinding(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	repo := dir + "/repo"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", repo)
	for _, kv := range [][2]string{{"user.name", "land-test"}, {"user.email", "land-test@example.com"}, {"commit.gpgsign", "false"}} {
		g.git(repo, "config", kv[0], kv[1])
	}
	g.git(repo, "commit", "-q", "--allow-empty", "-m", "base")
	mkdir(t, repo+"/docs/self-review")
	const rep = "docs/self-review/kan-u-self-review.md"
	g.write(repo+"/"+rep, "## 5. Go app — `flow-stats-app`\n\n- **[flow-stats-app]** a row shows no tokens; cause in flowd recording, not traced — noted\n")
	if g.err != nil {
		t.Fatal(g.err)
	}
	base := g.git(repo, "rev-parse", "HEAD")
	env := Env{Dir: dir, Getenv: os.Getenv, LookupEnv: hermeticGitLookup}
	var o, e bytes.Buffer
	rc := Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-u self-review report", rep}, env, &o, &e)
	want := "LAND-UNTRACED-FINDING: " + rep + ":3: - **[flow-stats-app]** a row shows no tokens; cause in flowd recording, not traced — noted — trace it to a verdict before landing; nothing added, committed, pulled or pushed\n"
	if rc != 4 || e.String() != want || o.String() != "" {
		t.Fatalf("rc=%d stdout=%q stderr=%q", rc, o.String(), e.String())
	}
	if head := g.git(repo, "rev-parse", "HEAD"); head != base || g.git(repo, "diff", "--cached", "--name-only") != "" {
		t.Fatalf("the refusal moved HEAD or staged a path: %s", g.git(repo, "status", "--short"))
	}

	g.write(repo+"/"+rep, "## 5. Go app — `flow-stats-app`\n\n- **[flow-stats-app]** a row shows no tokens; disproved — the row was the aborted attempt's — noted\n- **[guard]** `LAND-UNTRACED-FINDING` fired on a blank line; disproved — noted\n")
	o.Reset()
	e.Reset()
	if rc := Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-u self-review report", rep}, env, &o, &e); rc != 0 {
		t.Fatalf("a traced report did not land: rc=%d stderr=%q", rc, e.String())
	}
}

func TestLandSelfReviewReportRefusesDeferredFinding(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	repo := dir + "/repo"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", repo)
	for _, kv := range [][2]string{{"user.name", "land-test"}, {"user.email", "land-test@example.com"}, {"commit.gpgsign", "false"}} {
		g.git(repo, "config", kv[0], kv[1])
	}
	g.git(repo, "commit", "-q", "--allow-empty", "-m", "base")
	mkdir(t, repo+"/docs/self-review")
	const rep = "docs/self-review/kan-d-self-review.md"
	const line = "- **[flow-fix]** `feature/cardio` lacks develop's devRestart isolation fix 98da4aad; a branch decision for the operator — noted"
	const tail = "- **[flow-fix]** merging it is your call — noted (left open)"
	const spot = "- **[flow-fix]** 14 goldens regenerated; please spot-check them — noted"
	g.write(repo+"/"+rep, "## 1. Problems — `flow-fix`\n\n"+line+"\n"+tail+"\n"+spot+"\n")
	if g.err != nil {
		t.Fatal(g.err)
	}
	base := g.git(repo, "rev-parse", "HEAD")
	env := Env{Dir: dir, Getenv: os.Getenv, LookupEnv: hermeticGitLookup}
	var o, e bytes.Buffer
	rc := Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-d self-review report", rep}, env, &o, &e)
	want := "LAND-DEFERRED-FINDING: " + rep + ":3: " + line + " — fix it in the run, never leave it to the operator; nothing added, committed, pulled or pushed\n" +
		"LAND-DEFERRED-FINDING: " + rep + ":4: " + tail + " — fix it in the run, never leave it to the operator; nothing added, committed, pulled or pushed\n" +
		"LAND-DEFERRED-FINDING: " + rep + ":5: " + spot + " — fix it in the run, never leave it to the operator; nothing added, committed, pulled or pushed\n"
	if rc != 4 || e.String() != want || o.String() != "" {
		t.Fatalf("rc=%d stdout=%q stderr=%q", rc, o.String(), e.String())
	}
	if head := g.git(repo, "rev-parse", "HEAD"); head != base || g.git(repo, "diff", "--cached", "--name-only") != "" {
		t.Fatalf("the refusal moved HEAD or staged a path: %s", g.git(repo, "status", "--short"))
	}

	g.write(repo+"/"+rep, "## 1. Problems — `flow-fix`\n\n"+
		"- **[flow-fix]** `feature/cardio` lacks develop's devRestart isolation fix 98da4aad; a branch decision for the operator — fixed: abc1234\n"+
		"- **[flow-fix]** resolving it required an operator decision about another change's work — filed: KAN-56\n"+
		"- **[flow-fix]** it needed an operator decision to skip — declined\n")
	o.Reset()
	e.Reset()
	if rc := Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-d self-review report", rep}, env, &o, &e); rc != 0 {
		t.Fatalf("a non-noted report did not land: rc=%d stderr=%q", rc, e.String())
	}
}

func TestLandSelfReviewReportRefusesAReportFailingItsCheck(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	repo := dir + "/repo"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", repo)
	for _, kv := range [][2]string{{"user.name", "land-test"}, {"user.email", "land-test@example.com"}, {"commit.gpgsign", "false"}} {
		g.git(repo, "config", kv[0], kv[1])
	}
	g.git(repo, "commit", "-q", "--allow-empty", "-m", "base")
	mkdir(t, repo+"/docs/self-review")
	mkdir(t, repo+"/scripts")
	g.write(repo+"/scripts/check-self-review-report.sh", `if [ -e "$1/UNANSWERABLE" ]; then exit 2; fi; if cat "$1"/*-self-review.md 2>/dev/null | grep -q BAD; then echo "$1/kan-c-self-review.md:3: BAD line"; exit 1; fi; exit 0`+"\n")
	const ctx = "docs/self-review/kan-c-context.md"
	const rep = "docs/self-review/kan-c-self-review.md"
	g.write(repo+"/"+ctx, "BAD\n")
	if g.err != nil {
		t.Fatal(g.err)
	}
	env := Env{Dir: dir, Getenv: os.Getenv, LookupEnv: hermeticGitLookup}
	var o, e bytes.Buffer
	land := func(path string) int {
		o.Reset()
		e.Reset()
		return Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-c self-review report", path}, env, &o, &e)
	}

	// (a) only a *-self-review.md add-path is checked.
	if rc := land(ctx); rc != 0 {
		t.Fatalf("(a) a context file did not land: rc=%d stderr=%q", rc, e.String())
	}

	// (b) a violation refuses before anything is staged.
	g.write(repo+"/"+rep, "# kan-c\n\nBAD\n")
	base := g.git(repo, "rev-parse", "HEAD")
	rc := land(rep)
	want := "LAND-REPORT-INVALID: " + rep + " fails check-self-review-report — fix the lines it names; nothing added, committed, pulled or pushed\n"
	if rc != 5 || e.String() != want || !strings.Contains(o.String(), "kan-c-self-review.md:3: BAD line") {
		t.Fatalf("(b) rc=%d stdout=%q stderr=%q", rc, o.String(), e.String())
	}
	if head := g.git(repo, "rev-parse", "HEAD"); head != base || g.git(repo, "diff", "--cached", "--name-only") != "" {
		t.Fatalf("(b) the refusal moved HEAD or staged a path: %s", g.git(repo, "status", "--short"))
	}

	// (c) a check that cannot answer refuses too.
	g.write(repo+"/"+rep, "# kan-c\n\nfine\n")
	g.write(repo+"/docs/self-review/UNANSWERABLE", "")
	rc = land(rep)
	want = "LAND-REPORT-UNCHECKED: " + rep + " — check-self-review-report could not answer (exit 2); nothing added, committed, pulled or pushed\n"
	if rc != 5 || e.String() != want {
		t.Fatalf("(c) rc=%d stdout=%q stderr=%q", rc, o.String(), e.String())
	}
	if head := g.git(repo, "rev-parse", "HEAD"); head != base {
		t.Fatalf("(c) the refusal moved HEAD")
	}

	// (d) a passing check lands the report.
	if err := os.Remove(repo + "/docs/self-review/UNANSWERABLE"); err != nil {
		t.Fatal(err)
	}
	if rc := land(rep); rc != 0 {
		t.Fatalf("(d) a passing report did not land: rc=%d stderr=%q", rc, e.String())
	}
	if g.git(repo, "log", "-1", "--format=%s", "--", rep) == "" {
		t.Fatalf("(d) the report was not committed")
	}
}

func TestLandSelfReviewReportRefusesAnUnlandedCase(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	repo, agents := dir+"/repo", dir+"/agents"
	var g fxGit
	for _, r := range []string{repo, agents} {
		g.git("", "init", "-q", "-b", "main", r)
		for _, kv := range [][2]string{{"user.name", "land-test"}, {"user.email", "land-test@example.com"}, {"commit.gpgsign", "false"}} {
			g.git(r, "config", kv[0], kv[1])
		}
	}
	g.git(repo, "commit", "-q", "--allow-empty", "-m", "base")
	mkdir(t, agents+"/docs/prompt-tuning/kan-c-1")
	g.write(agents+"/docs/prompt-tuning/kan-c-1/case.json", "{}\n")
	g.git(agents, "add", "-A")
	g.git(agents, "commit", "-q", "-m", "case")
	g.git(agents, "update-ref", "refs/remotes/origin/main", "HEAD")
	mkdir(t, repo+"/docs/self-review")
	const rep = "docs/self-review/kan-c-self-review.md"
	const bare = "- **[flow-prompt-tuning]** Prompt-tuning case: an implementer reads per item — noted"
	const absent = "- **[flow-prompt-tuning]** reads per item — implementer — docs/prompt-tuning/kan-c-2/ — noted"
	g.write(repo+"/"+rep, "## 8. Prompt tuning — `flow-prompt-tuning`\n\n"+bare+"\n"+absent+"\n")
	if g.err != nil {
		t.Fatal(g.err)
	}
	getenv := func(k string) string {
		if k == "FLOW_GUARD_REPO_ROOT" {
			return agents
		}
		return os.Getenv(k)
	}
	env := Env{Dir: dir, Getenv: getenv, LookupEnv: hermeticGitLookup}
	var o, e bytes.Buffer
	rc := Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-c self-review report", rep}, env, &o, &e)
	if rc != 4 || strings.Count(e.String(), "LAND-CASE-MISSING") != 2 || !strings.Contains(e.String(), "kan-c-2/ is not on") {
		t.Fatalf("rc=%d stderr=%q", rc, e.String())
	}
	g.write(repo+"/"+rep, "## 8. Prompt tuning — `flow-prompt-tuning`\n\n- **[flow-prompt-tuning]** reads per item — implementer — docs/prompt-tuning/kan-c-1/ — noted\n")
	o.Reset()
	e.Reset()
	if rc := Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-c self-review report", rep}, env, &o, &e); rc != 0 {
		t.Fatalf("a landed case did not land: rc=%d stderr=%q", rc, e.String())
	}

	// A case committed on the report's own branch, not yet on origin/main,
	// lands with it.
	mkdir(t, repo+"/docs/prompt-tuning/kan-c-3")
	g.write(repo+"/docs/prompt-tuning/kan-c-3/case.json", "{}\n")
	g.git(repo, "add", "docs/prompt-tuning")
	g.git(repo, "commit", "-q", "-m", "case on the branch")
	g.write(repo+"/"+rep, "## 8. Prompt tuning — `flow-prompt-tuning`\n\n- **[flow-prompt-tuning]** reads per item — implementer — docs/prompt-tuning/kan-c-3/ — noted\n")
	if g.err != nil {
		t.Fatal(g.err)
	}
	o.Reset()
	e.Reset()
	if rc := Registry["land-self-review-report"]([]string{repo, "main", "docs(self-review): kan-c self-review report", rep}, env, &o, &e); rc != 0 {
		t.Fatalf("a case on the report's branch did not land: rc=%d stderr=%q", rc, e.String())
	}
}
