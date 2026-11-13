package guard

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// rbNewFx is bmNewFx plus the change's plan on disk and never committed, as
// the gitignored change directory is, and the repository-local config the guard's own git needs, since it runs under the
// process environment: an identity, no signing, and no rebase autostash (a
// global autostash would turn the refused case into a rebase).
func rbNewFx(t *testing.T) *bmFx {
	t.Helper()
	fx := bmNewFx(t)
	for _, kv := range [][2]string{{"user.name", "Test"}, {"user.email", "test@example.invalid"},
		{"commit.gpgsign", "false"}, {"rebase.autoStash", "false"}} {
		fx.g.git(fx.repo, "config", kv[0], kv[1])
	}
	mkdir(t, fx.repo+"/spectre/changes/demo")
	fx.g.write(fx.repo+"/spectre/changes/demo/tasks.md", "plan")
	return fx
}

func TestRebaseOntoTip(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	rebasing := func(fx *bmFx) bool {
		_, err := os.Stat(fx.repo + "/.git/rebase-merge")
		return err == nil
	}

	t.Run("rebased", func(t *testing.T) {
		t.Parallel()
		fx := rbNewFx(t)
		fx.advanceBase("unrelated1.txt")
		fx.g.appendLine(fx.repo+"/base.txt", "demo")
		fx.g.git(fx.repo, "commit", "-qam", "demo work")
		tip := fx.g.git(fx.repo, "rev-parse", "main")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		var errb bytes.Buffer
		o, err := rebaseOntoTip(Env{Dir: fx.dir, Getenv: os.Getenv}, fx.repo, "main", &errb)
		if err != nil || o.kind != "rebased" || o.sha != tip || len(o.unmerged) != 0 {
			t.Fatalf("got %+v, %v; want rebased onto %s\nstderr %s", o, err, tip, errb.String())
		}
		if mb := fx.g.git(fx.repo, "merge-base", "HEAD", "main"); mb != tip {
			t.Errorf("merge base %s, want the tip %s", mb, tip)
		}
	})

	t.Run("conflict leaves the rebase in progress", func(t *testing.T) {
		t.Parallel()
		fx := rbNewFx(t)
		fx.advanceBase("shared.txt")
		fx.g.appendLine(fx.repo+"/shared.txt", "demo")
		fx.g.git(fx.repo, "commit", "-qam", "demo touches shared")
		tip := fx.g.git(fx.repo, "rev-parse", "main")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		var errb bytes.Buffer
		o, err := rebaseOntoTip(Env{Dir: fx.dir, Getenv: os.Getenv}, fx.repo, "main", &errb)
		if err != nil || o.kind != "conflict" || o.sha != tip || strings.Join(o.unmerged, ",") != "shared.txt" {
			t.Fatalf("got %+v, %v; want a conflict on shared.txt onto %s\nstderr %s", o, err, tip, errb.String())
		}
		if !rebasing(fx) {
			t.Error("no rebase in progress: the conflict was not left for resolution")
		}
	})

	t.Run("refused with no rebase in progress", func(t *testing.T) {
		t.Parallel()
		fx := rbNewFx(t)
		fx.advanceBase("unrelated1.txt")
		fx.g.appendLine(fx.repo+"/base.txt", "unstaged implementation edit")
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		var errb bytes.Buffer
		o, err := rebaseOntoTip(Env{Dir: fx.dir, Getenv: os.Getenv}, fx.repo, "main", &errb)
		if err != nil || o.kind != "refused" {
			t.Fatalf("got %+v, %v; want refused\nstderr %s", o, err, errb.String())
		}
		if rebasing(fx) {
			t.Error("a rebase is in progress after a refusal")
		}
	})

	t.Run("the tip is pinned to a sha before a concurrent fetch moves the ref", func(t *testing.T) {
		t.Parallel()
		fx := rbNewFx(t)
		fx.advanceBase("unrelated1.txt")
		pinned := fx.g.git(fx.repo, "rev-parse", "main")
		fx.advanceBase("unrelated2.txt")
		fx.g.git(fx.repo, "branch", "later", "main")
		fx.g.git(fx.repo, "update-ref", "refs/heads/main", pinned)
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		// The stub moves main to `later` the moment rebase starts: a rebase
		// onto the name would land on the moved ref, one onto the sha not.
		stub := fx.dir + "/stub"
		writeExec(t, stub+"/git", fmt.Sprintf(`#!/usr/bin/env bash
for a in "$@"; do
  if [ "$a" = rebase ]; then %[1]q "$1" "$2" update-ref refs/heads/main "$(%[1]q "$1" "$2" rev-parse later)"; break; fi
done
exec %[1]q "$@"
`, fixtureGit))
		var errb bytes.Buffer
		o, err := rebaseOntoTip(Env{Dir: fx.dir, Getenv: pathEnv(stub)}, fx.repo, "main", &errb)
		if err != nil || o.kind != "rebased" || o.sha != pinned {
			t.Fatalf("got %+v, %v; want rebased onto %s\nstderr %s", o, err, pinned, errb.String())
		}
		if later := fx.g.git(fx.repo, "rev-parse", "later"); fx.g.git(fx.repo, "rev-parse", "main") != later {
			t.Fatal("the stub never moved main: the pin was not exercised")
		}
		if fx.g.git(fx.repo, "merge-base", "HEAD", "later") != pinned {
			t.Error("HEAD was rebased onto the moved ref, not the pinned sha")
		}
	})

	t.Run("an unresolvable ref cannot be answered", func(t *testing.T) {
		t.Parallel()
		fx := rbNewFx(t)
		var errb bytes.Buffer
		if o, err := rebaseOntoTip(Env{Dir: fx.dir, Getenv: os.Getenv}, fx.repo, "no-such-base", &errb); err == nil {
			t.Fatalf("got %+v, want an error", o)
		}
	})
}

// rbRead is a file's content, "" when it cannot be read.
func rbRead(path string) string {
	b, _ := os.ReadFile(path)
	return string(b)
}
