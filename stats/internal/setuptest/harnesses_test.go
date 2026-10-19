package setuptest

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
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

// Claude Code, zcode and muse are the only harnesses. A global run writes nothing under
// ~/.cursor or ~/.codex, leaves an earlier install's directories there exactly as it found
// them, and the retired modes are refused with the usage line rather than half-run.
func TestGlobalWritesNothingUnderCursorCodex(t *testing.T) {
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
		g.assertContains("setup.sh "+mode+" prints the usage line", g.log, "Choose: claude-code | zcode | muse | global")
		var names []string
		entries, _ := os.ReadDir(proj)
		for _, e := range entries {
			names = append(names, e.Name())
		}
		g.assertEq("setup.sh "+mode+" writes nothing into the project", "", strings.Join(names, "\n"))
	}
}

// The ZCode layer. Its contract is isolation in both directions: a global install gives ZCode
// a self-contained copy under ~/.zcode (nothing installed there may point at another
// harness's paths, and the claude block must not see the rewrite), and the per-project zcode
// mode writes project paths only — never user-level state, never another harness's
// directories.
func TestZcodeGlobalSelfContained(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("global with the zcode layer succeeds", g.rc, g.log)
	z := filepath.Join(home, ".zcode")
	g.assertSymlink("zcode skill links land beside the other harnesses'", filepath.Join(z, "skills/demo-skill"))
	g.assertSymlink("zcode gets the claude command set, as the format is shared", filepath.Join(z, "commands/demo.md"))
	g.assertSymlink("zcode rule full texts are linked", filepath.Join(z, "rules/good-always.md"))
	g.assertSymlink("the zcode hook is linked", filepath.Join(z, "hooks/enforce-agent-baseline.py"))
	g.assertSymlink("the zcode flow-active-change hook is linked", filepath.Join(z, "hooks/flow-active-change.py"))
	g.assertSymlink("the zcode protect-main-checkout hook is linked", filepath.Join(z, "hooks/protect-main-checkout.py"))
	g.assertFileExists("the zcode agent baseline is generated, not linked", filepath.Join(z, "rules/agent-baseline.md"))
	g.assertContains("the generated baseline says where it came from", filepath.Join(z, "rules/agent-baseline.md"), "Generated by agents/setup.sh")
	g.assertContains("the zcode block points at zcode-local full texts", filepath.Join(z, "AGENTS.md"), "Full rule: `~/.zcode/rules/cored.md`")
	g.assertNotContains("the zcode block carries no ~/.claude pointer", filepath.Join(z, "AGENTS.md"), "~/.claude/")
	g.assertContains("the claude block keeps its own pointers — no rewrite leak", filepath.Join(home, ".claude/CLAUDE.md"), "Full rule: `~/.claude/rules/cored.md`")
	g.assertContains("the zcode hook registration snippet is printed", g.log, ".zcode/hooks/enforce-agent-baseline.py")
	g.assertContains("the zcode flow-active-change registration snippet is printed", g.log, ".zcode/hooks/flow-active-change.py")
	g.assertSymlink("the zcode enforce-loop-dispatch hook is linked", filepath.Join(z, "hooks/enforce-loop-dispatch.py"))
	g.assertContains("the zcode enforce-loop-dispatch registration snippet is printed", g.log, ".zcode/hooks/enforce-loop-dispatch.py")

	// --- idempotence: a second global must refresh the zcode layer, never duplicate it.
	home = g.newHome()
	g.runSetup(fixture, home, "global", "")
	first := filepath.Join(g.dir, "zcode-block-first.md")
	copyFile(t, filepath.Join(home, ".zcode/AGENTS.md"), first)
	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("the second global with the zcode layer succeeds", g.rc, g.log)
	g.assertIdentical("the zcode block is byte-identical on re-run", filepath.Join(home, ".zcode/AGENTS.md"), first)
}

func TestZcodeProjectModeProjectPathsOnly(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	zproj := filepath.Join(g.dir, "zcode-project")
	g.runSetup(fixture, home, "zcode", zproj)
	g.assertRCZero("the zcode project install succeeds", g.rc, g.log)
	g.assertSymlink("project skills land in .zcode/skills", filepath.Join(zproj, ".zcode/skills/demo-skill"))
	g.assertSymlink("project commands land in .zcode/commands", filepath.Join(zproj, ".zcode/commands/demo.md"))
	g.assertFileExists("AGENTS.md is copied to the project root", filepath.Join(zproj, "AGENTS.md"))
	g.assertAbsent("no user-level zcode state is created", filepath.Join(home, ".zcode"))
	g.assertAbsent("no claude state is created", filepath.Join(home, ".claude"))
}

// The Muse layer is project-only: the harness reads project skills from
// .agents/skills/ and project rules from AGENTS.md, and has no commands
// directory, no project agent definitions, no hooks and no global rules file
// of its own — so the mode installs skills and that file and nothing else,
// and global installs nothing for it.
func TestMuseProjectModeProjectPathsOnly(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	mproj := filepath.Join(g.dir, "muse-project")
	g.runSetup(fixture, home, "muse", mproj)
	g.assertRCZero("the muse project install succeeds", g.rc, g.log)
	g.assertSymlink("project skills land in .agents/skills", filepath.Join(mproj, ".agents/skills/demo-skill"))
	g.assertFileExists("AGENTS.md is copied to the project root", filepath.Join(mproj, "AGENTS.md"))
	g.assertAbsent("no user-level state is created", filepath.Join(home, ".claude"))
	g.assertAbsent("no zcode state is created", filepath.Join(home, ".zcode"))

	home = g.newHome()
	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("global still succeeds", g.rc, g.log)
	g.assertAbsent("global installs no muse-native skills root", filepath.Join(home, ".agents"))
}

func TestZcodeCompactWindowEnvBlock(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	zshrc, bashrc := filepath.Join(home, ".zshrc"), filepath.Join(home, ".bashrc")
	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("global with the env step succeeds", g.rc, g.log)
	g.assertContains(".zshrc carries the managed export", zshrc, "export Z_COMPACT_WINDOW=500000")
	g.assertContains(".zshrc block is delimited", zshrc, "# flow:zcode-env:begin")
	g.assertEq("exactly one begin marker", 1, countLinesMatching(zshrc, "# flow:zcode-env:begin"))
	g.assertAbsent("an absent .bashrc is not created", bashrc)

	// --- a pre-existing .bashrc gets the block too; re-run refreshes, never duplicates.
	g.seedFile(bashrc, 0o644, "# my bash notes\n")
	g.runSetup(fixture, home, "global", "")
	g.assertContains(".bashrc gets the block when it exists", bashrc, "export Z_COMPACT_WINDOW=500000")
	g.assertContains("the user's own .bashrc content survives", bashrc, "# my bash notes")
	g.runSetup(fixture, home, "global", "")
	g.assertEq("re-run leaves exactly one block in .zshrc", 1, countLinesMatching(zshrc, "# flow:zcode-env:begin"))

	// --- an unmanaged export wins: the installer warns and does not touch the file.
	home = g.newHome()
	zshrc = filepath.Join(home, ".zshrc")
	g.seedFile(zshrc, 0o644, "# my own zsh setup\nexport Z_COMPACT_WINDOW=200000\n")
	expected := filepath.Join(g.dir, "zshrc-unmanaged-expected.md")
	copyFile(t, zshrc, expected)
	g.runSetup(fixture, home, "global", "")
	g.assertIdentical("an unmanaged Z_COMPACT_WINDOW line is left byte-identical", zshrc, expected)
	g.assertContains("the skip is explained in the log", g.log, "outside the managed block")
}

// TestGlobalInstallRefusedFromLinkedWorktree pins KAN-773: `global` into the operator's own
// home (SETUP_REAL_HOME here) from a linked worktree exits non-zero naming the main checkout's
// command and links nothing; the main checkout still installs there, and the worktree still
// installs into any other home — the sandboxed one every other test here uses.
func TestGlobalInstallRefusedFromLinkedWorktree(t *testing.T) {
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

// setup.sh global writes an exec shim per flow script into ~/.local/bin, so the bare script
// names the skills cite resolve on PATH (kan-998). A shim reaches the script's own lib/, a shim
// of a deleted script is removed, and a user's own file there is never overwritten or removed.
func TestGlobalInstallsScriptShims(t *testing.T) {
	const marker = "# flow script shim — written by setup.sh global; do not edit"
	g := newGroup(t)
	fx := filepath.Join(g.dir, "fixture-repo-shims")
	if err := makeFixtureRepo(fx, false); err != nil {
		t.Fatal(err)
	}
	g.seedFile(filepath.Join(fx, "scripts/doomed.sh"), 0o755, "#!/bin/sh\necho doomed\n")
	g.seedFile(filepath.Join(fx, "scripts/test-demo.sh"), 0o755, "#!/bin/sh\necho test\n")
	g.seedFile(filepath.Join(fx, "scripts/sourced.sh"), 0o644, "sourced=1\n")
	home := g.newHome()
	bin := filepath.Join(home, ".local/bin")

	g.runSetup(fx, home, "global", "")
	g.assertRCZero("global succeeds", g.rc, g.log)
	for _, name := range []string{"project-get.sh", "doomed.sh"} {
		p := filepath.Join(bin, name)
		g.assertExecutable(name+" shim is executable", p)
		if isSymlink(p) {
			g.fail(name+" shim is not a symlink", p+" is a symlink")
		}
		if b, err := os.ReadFile(p); err != nil || !strings.Contains(string(b), marker+"\n") {
			g.fail(name+" shim carries the marker", p+": "+string(b))
		}
	}
	g.assertAbsent("a test- script gets no shim", filepath.Join(bin, "test-demo.sh"))
	g.assertAbsent("a non-executable script gets no shim", filepath.Join(bin, "sourced.sh"))

	proj := filepath.Join(g.dir, "shim-proj")
	g.seedFile(filepath.Join(proj, ".flow/project.md"), 0o644, "# p\n\n## default landing route\n\nmerge and push\n")
	out, err := exec.Command(filepath.Join(bin, "project-get.sh"), proj, "default landing route").Output()
	if err != nil || strings.TrimSpace(string(out)) != "merge and push" {
		g.fail("the shim reaches the script's own lib/", string(out)+" "+errString(err))
	}

	if err := os.Remove(filepath.Join(fx, "scripts/doomed.sh")); err != nil {
		t.Fatal(err)
	}
	g.seedFile(filepath.Join(bin, "mine.sh"), 0o755, "#!/bin/sh\necho mine\n")
	g.runSetup(fx, home, "global", "")
	g.assertRCZero("re-install succeeds", g.rc, g.log)
	g.assertAbsent("a deleted script's shim is removed", filepath.Join(bin, "doomed.sh"))
	if b, _ := os.ReadFile(filepath.Join(bin, "mine.sh")); string(b) != "#!/bin/sh\necho mine\n" {
		g.fail("a user's own file is unchanged", string(b))
	}

	home = g.newHome()
	bin = filepath.Join(home, ".local/bin")
	own := "#!/bin/sh\necho users own doomed\n"
	g.seedFile(filepath.Join(bin, "doomed.sh"), 0o755, own)
	g.seedFile(filepath.Join(fx, "scripts/doomed.sh"), 0o755, "#!/bin/sh\necho doomed\n")
	g.runSetup(fx, home, "global", "")
	if g.rc != 1 {
		g.fail("a user's file in a shim's place is SKIPPED (rc 1)", g.log)
	}
	if b, _ := os.ReadFile(filepath.Join(bin, "doomed.sh")); string(b) != own {
		g.fail("the user's file is byte-identical", string(b))
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
