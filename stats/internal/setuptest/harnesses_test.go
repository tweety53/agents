package setuptest

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// A re-install after a skill or command is DELETED from the source tree must remove the stale
// symlink it left behind. Installs iterate the current source tree only, so without a prune
// every deletion leaves a dangling link at every destination forever — and a stale command
// link still matches the harness's command glob, so a deleted command goes on being offered
// and then fails on a broken link. The 12→3 state rename retired fifteen skills and thirteen
// commands at once, which is what made this reachable.
//
// It deletes from its fixture, so it builds its own; the bash harness left that fixture in
// place for every later group, and the groups after it use the shared base fixture instead —
// the same content.
func TestReinstallPrunesDeletedSourceLinks(t *testing.T) {
	smoke.Heavy(t)
	g := newGroup(t)
	fx := filepath.Join(g.dir, "fixture-repo-prune")
	if err := makeFixtureRepo(fx, false); err != nil {
		t.Fatal(err)
	}
	home := g.newHome()
	g.seedFile(filepath.Join(fx, "skills/doomed-skill/SKILL.md"), 0o644, "---\nname: doomed-skill\n---\nbody\n")
	g.seedFile(filepath.Join(fx, "commands-claude/doomed.md"), 0o644, "---\nname: /doomed\n---\nbody\n")

	g.runSetup(fx, home, "global", "")
	g.assertRCZero("first install succeeds", g.rc, g.log)
	g.assertSymlink("the doomed skill is installed", filepath.Join(home, ".claude/skills/doomed-skill"))
	g.assertSymlink("the doomed command is installed", filepath.Join(home, ".claude/commands/doomed.md"))

	// Two things of the user's own, sitting alongside — both must survive the prune.
	commands := filepath.Join(home, ".claude/commands")
	if err := os.WriteFile(filepath.Join(commands, "my-own-note.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A BROKEN symlink the user made, pointing outside this repo — e.g. an unmounted volume or
	// a checkout they moved. It is broken right now, which is exactly what makes it
	// indistinguishable from a stale installer link unless the prune checks where the link
	// points.
	if err := os.Symlink(filepath.Join(g.dir, "not-mounted-right-now/thing.md"), filepath.Join(commands, "user-own-broken.md")); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(filepath.Join(fx, "skills/doomed-skill")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(fx, "commands-claude/doomed.md")); err != nil {
		t.Fatal(err)
	}
	g.runSetup(fx, home, "global", "")
	g.assertRCZero("re-install after deletion succeeds", g.rc, g.log)
	g.assertAbsent("the stale skill link is pruned", filepath.Join(home, ".claude/skills/doomed-skill"))
	g.assertAbsent("the stale command link is pruned", filepath.Join(commands, "doomed.md"))
	g.assertFileExists("a real file the user put there is NOT pruned", filepath.Join(commands, "my-own-note.md"))
	g.assertSymlink("a user's own BROKEN symlink outside this repo is NOT pruned", filepath.Join(commands, "user-own-broken.md"))
	g.assertSymlink("a still-live skill link survives the prune", filepath.Join(home, ".claude/skills/demo-skill"))
	g.assertSymlink("a still-live command link survives the prune", filepath.Join(commands, "demo.md"))
}

// Claude Code is the only harness. A global run writes nothing under
// ~/.cursor or ~/.codex, leaves an earlier install's directories there exactly as it found
// them, and the retired modes are refused with the usage line rather than half-run.
func TestGlobalWritesNothingUnderCursorCodex(t *testing.T) {
	smoke.Heavy(t)
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("global succeeds", g.rc, g.log)
	g.assertAbsent("no ~/.cursor is created", filepath.Join(home, ".cursor"))
	g.assertAbsent("no ~/.codex is created", filepath.Join(home, ".codex"))

	// --- an earlier install's ~/.cursor and ~/.codex are left alone: not deleted, not written.
	home = g.newHome()
	g.seedFile(filepath.Join(home, ".cursor/rules/old.mdc"), 0o644, "old cursor rule\n")
	g.seedFile(filepath.Join(home, ".codex/AGENTS.md"), 0o644, "old codex block\n")
	dirs := []string{filepath.Join(home, ".cursor"), filepath.Join(home, ".codex")}
	oldHarnessFP := func() string {
		var paths []string
		for _, d := range dirs {
			_ = filepath.WalkDir(d, func(p string, _ fs.DirEntry, err error) error {
				if err == nil {
					paths = append(paths, p)
				}
				return nil
			})
		}
		sort.Strings(paths)
		return sha256hex(strings.Join(paths, "\n") + "\n" + treeFingerprint(dirs...))
	}
	before := oldHarnessFP()
	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("global over an earlier cursor/codex install succeeds", g.rc, g.log)
	g.assertEq("~/.cursor and ~/.codex are byte-identical after the run", before, oldHarnessFP())
}

func TestRetiredModesRefused(t *testing.T) {
	g := newGroup(t)
	for _, mode := range []string{"cursor", "codex", "all"} {
		home := g.newHome()
		proj := filepath.Join(g.dir, "retired-"+mode)
		g.runSetup(fixture, home, mode, proj)
		g.assertRCNonzero("setup.sh "+mode+" exits non-zero", g.rc)
		g.assertContains("setup.sh "+mode+" prints the usage line", g.log, "Choose: claude-code | global")
		var names []string
		entries, _ := os.ReadDir(proj)
		for _, e := range entries {
			names = append(names, e.Name())
		}
		g.assertEq("setup.sh "+mode+" writes nothing into the project", "", strings.Join(names, "\n"))
	}
}

// TestGlobalInstallRefusedFromLinkedWorktree pins KAN-773: `global` into the operator's own
// home (SETUP_REAL_HOME here) from a linked worktree exits non-zero naming the main checkout's
// command and links nothing; the main checkout still installs there, and the worktree still
// installs into any other home — the sandboxed one every other test here uses.
func TestGlobalInstallRefusedFromLinkedWorktree(t *testing.T) {
	smoke.Heavy(t)
	g := newGroup(t)
	main := filepath.Join(g.dir, "fixture-repo-main")
	if err := makeFixtureRepo(main, false); err != nil {
		t.Fatal(err)
	}
	wt := filepath.Join(g.dir, "fixture-repo-wt")
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"add", "-A"},
		{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "fixture"},
		{"worktree", "add", "-q", wt},
	} {
		if out, err := exec.Command("git", append([]string{"-C", main}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	home := g.newHome()
	g.env = []string{"SETUP_REAL_HOME=" + home}
	g.runSetup(wt, home, "global", "")
	g.assertRCNonzero("global from a linked worktree exits non-zero", g.rc)
	g.assertContains("the refusal names the main checkout's command", g.log, filepath.Join(main, "setup.sh")+" global")
	g.assertAbsent("nothing is linked into $HOME", filepath.Join(home, ".claude/rules"))

	g.runSetup(main, home, "global", "")
	g.assertRCZero("global from the main checkout succeeds", g.rc, g.log)

	g.runSetup(wt, g.newHome(), "global", "")
	g.assertRCZero("global from the worktree into a sandboxed home succeeds", g.rc, g.log)
}

// setup.sh global writes exactly one exec shim, flow-guard, into ~/.local/bin, so the
// `flow-guard <name>` the skills cite resolves on PATH (kan-998, KAN-1058). A marked per-name .sh
// shim an earlier install wrote is removed, and a user's own file there — unmarked, or in the
// flow-guard shim's place — is never overwritten or removed.
func TestGlobalInstallsScriptShims(t *testing.T) {
	smoke.Heavy(t)
	const marker = "# flow script shim — written by setup.sh global; do not edit"
	g := newGroup(t)
	fx := filepath.Join(g.dir, "fixture-repo-shims")
	if err := makeFixtureRepo(fx, false); err != nil {
		t.Fatal(err)
	}
	home := g.newHome()
	bin := filepath.Join(home, ".local/bin")
	g.seedFile(filepath.Join(bin, "check-plan-shape.sh"), 0o755, "#!/bin/sh\n"+marker+"\nexec '/old/scripts/check-plan-shape.sh' \"$@\"\n")
	own := "#!/bin/sh\necho mine\n"
	g.seedFile(filepath.Join(bin, "mine.sh"), 0o755, own)

	g.runSetup(fx, home, "global", "")
	g.assertRCZero("global succeeds", g.rc, g.log)
	var marked []string
	ents, _ := os.ReadDir(bin)
	for _, e := range ents {
		if b, err := os.ReadFile(filepath.Join(bin, e.Name())); err == nil && strings.Contains(string(b), marker+"\n") {
			marked = append(marked, e.Name())
		}
	}
	if len(marked) != 1 || marked[0] != "flow-guard" {
		g.fail("exactly one marked shim, flow-guard", strings.Join(marked, " "))
	}
	p := filepath.Join(bin, "flow-guard")
	g.assertExecutable("flow-guard shim is executable", p)
	if isSymlink(p) {
		g.fail("flow-guard shim is not a symlink", p+" is a symlink")
	}
	g.assertAbsent("a marked per-name .sh shim is removed", filepath.Join(bin, "check-plan-shape.sh"))
	if b, _ := os.ReadFile(filepath.Join(bin, "mine.sh")); string(b) != own {
		g.fail("an unmarked file is kept byte-identical", string(b))
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "exec '"+filepath.Join(fx, "scripts/flow-guard")+"' \"$@\"\n") {
		g.fail("the flow-guard shim execs the checkout's scripts/flow-guard", string(b))
	}
	proj := filepath.Join(g.dir, "shim-proj")
	g.seedFile(filepath.Join(proj, ".flow/project.md"), 0o644, "# p\n\n## default landing route\n\nmerge and push\n")
	out, err := exec.Command(p, "project-get", proj, "default landing route").Output()
	if err != nil || strings.TrimSpace(string(out)) != "merge and push" {
		g.fail("the flow-guard shim runs a guard through the checkout's bootstrap", string(out)+" "+errString(err))
	}

	home = g.newHome()
	bin = filepath.Join(home, ".local/bin")
	userOwn := "#!/bin/sh\necho users own flow-guard\n"
	g.seedFile(filepath.Join(bin, "flow-guard"), 0o755, userOwn)
	g.runSetup(fx, home, "global", "")
	if g.rc != 1 {
		g.fail("a user's file in the shim's place is SKIPPED (rc 1)", g.log)
	}
	if b, _ := os.ReadFile(filepath.Join(bin, "flow-guard")); string(b) != userOwn {
		g.fail("the user's file is byte-identical", string(b))
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
