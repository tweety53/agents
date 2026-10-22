package repotest

// TestFlowTaskListLanded runs the flow-task-list mod's landedIn
// (mods/flow-task-list/hooks/landed.ts) under node against a real git
// repository. The mod's own suite (`claude plugin test`) cannot spawn a
// process, so it only replays git's answers; this test proves the argv it
// builds resolves the base and parses the Task-Id trailers the way real git
// answers them.
//
// The fixture: origin's main gains two other changes' Task-Id commits (5, 6)
// after the clone, so the local main is behind origin; origin's feat carries
// Task-Id 8. A change branch cut from origin/main with no recorded base must
// land only its own tasks (base refs/remotes/origin/HEAD); a change branch cut
// from origin/feat with branch.<it>.flowBase=feat must land only its own task
// (base refs/remotes/origin/feat); a directory that is no repository lands
// nothing.

import (
	"path/filepath"
	"strings"
	"testing"
)

const landedScript = `
import { spawnSync } from "node:child_process"
const { landedIn } = await import(process.argv[1])
const run = async argv => {
  const r = spawnSync(argv[0], argv.slice(1), { encoding: "utf8" })
  return { exitCode: r.status ?? 1, stdout: r.stdout }
}
console.log((await landedIn(run, process.argv[2])).sort((a, b) => a - b).join(" "))
`

func TestFlowTaskListLanded(t *testing.T) {
	t.Parallel()
	landed := filepath.Join(repoRoot(t), "mods", "flow-task-list", "hooks", "landed.ts")
	tmp := t.TempDir()
	env := []string{"HOME=" + t.TempDir()}
	g := func(args ...string) {
		t.Helper()
		sh(t, tmp, env, "git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t",
			"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main"}, args...)...)
	}
	commit := func(dir, msg string) { g("-C", dir, "commit", "-q", "--allow-empty", "-m", msg) }
	p := func(name string) string { return filepath.Join(tmp, name) }

	g("init", "-q", p("origin"))
	commit(p("origin"), "init")
	g("clone", "-q", p("origin"), p("work"))
	commit(p("origin"), "other change\n\nTask-Id: 5")
	g("-C", p("origin"), "checkout", "-q", "-b", "feat")
	commit(p("origin"), "feat work\n\nTask-Id: 8")
	g("-C", p("origin"), "checkout", "-q", "main")
	commit(p("origin"), "more other\n\nTask-Id: 6")
	g("-C", p("work"), "fetch", "-q", "origin")

	// The main checkout stays on its local main, behind origin; each change is a worktree beside it.
	g("-C", p("work"), "worktree", "add", "-q", "-b", "change", p("change"), "origin/main")
	commit(p("change"), "task one\n\nTask-Id: 1")
	commit(p("change"), "tasks two and three\n\nTask-Id: 2+3")
	g("-C", p("work"), "worktree", "add", "-q", "-b", "feat-change", p("feat"), "origin/feat")
	g("-C", p("feat"), "config", "branch.feat-change.flowBase", "feat")
	commit(p("feat"), "task four\n\nTask-Id: 4")

	for _, c := range []struct{ name, dir, want string }{
		{"no recorded base: since refs/remotes/origin/HEAD", p("change"), "1 2 3"},
		{"recorded flowBase: since refs/remotes/origin/<it>", p("feat"), "4"},
		{"no repository: nothing landed", tmp, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := sh(t, tmp, env, "node", "--no-warnings", "--input-type=module", "-e", landedScript, landed, c.dir)
			if got != c.want {
				t.Errorf("want %q, got %q", c.want, strings.TrimSpace(got))
			}
		})
	}
}
