package setuptest

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Containment. Each of these fails silently if it regresses: an install that looks completely
// successful while shipping a rule it must not, or shadowing a symlink with a stale copy of
// the same skill.

func TestOnlyAlwaysApplyRulesInstall(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("the fixture install succeeds", g.rc, g.log)
	g.assertExists("the always-on rule is installed", filepath.Join(home, ".claude/rules/good-always.md"))
	g.assertAbsent("alwaysApply: false is not installed", filepath.Join(home, ".claude/rules/opt-in-false.md"))
	g.assertAbsent("an unterminated frontmatter is not installed", filepath.Join(home, ".claude/rules/unterminated.md"))
	g.assertAbsent("alwaysApply: true_for_kotlin_only is not installed", filepath.Join(home, ".claude/rules/tricky-prefix.md"))
	ruleCount := 0
	for _, name := range visibleEntries(filepath.Join(home, ".claude/rules")) {
		if name != "agent-baseline.md" {
			ruleCount++
		}
	}
	// Two, and exactly two: good-always and cored are the fixture's always-on rules (the agent
	// baseline linked beside them is not a rule).
	// The count is the assertion that catches an opt-in rule sneaking in under a new code path,
	// so it is stated as a number rather than derived from the fixture.
	g.assertEq("exactly the two always-on rules are installed", 2, ruleCount)
	claudeMD := filepath.Join(home, ".claude/CLAUDE.md")
	g.assertContains("the managed block carries the always-on rule body", claudeMD, "BODY-GOOD-ALWAYS")
	g.assertNotContains("the managed block omits the opt-in rule", claudeMD, "BODY-OPT-IN")
	g.assertNotContains("the managed block omits the unterminated rule", claudeMD, "BODY-UNTERMINATED")
	g.assertNotContains("the managed block omits the true-prefixed rule", claudeMD, "BODY-TRICKY-PREFIX")
}

// The global layer is two halves of one source: the managed block carries each rule's core and
// a pointer, ~/.claude/rules/ carries the full text as a symlink into the repo. The failure
// this guards against is silent — a pointer that names a path no install creates, or a block
// that quietly inlines a rule's entire body again.
func TestCoreExcerptsLinksAndBaseline(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("the install succeeds", g.rc, g.log)

	claudeMD := filepath.Join(home, ".claude/CLAUDE.md")
	g.assertContains("the block carries the core excerpt", claudeMD, "BODY-CORE-EXCERPT")
	g.assertNotContains("the block stops at the closing marker", claudeMD, "BODY-AFTER-CORE")
	g.assertNotContains("no core marker survives into the block", claudeMD, "<!-- core -->")
	g.assertContains("the block points at the installed full rule", claudeMD, "Full rule: `~/.claude/rules/cored.md`.")
	// ZCode reads the same rendered text, so its block must carry the same core.
	g.assertContains("AGENTS.md carries the core too", filepath.Join(home, ".zcode/AGENTS.md"), "BODY-CORE-EXCERPT")

	rules := filepath.Join(home, ".claude/rules")
	g.assertExists("the full text is linked into ~/.claude/rules", filepath.Join(rules, "cored.md"))
	g.assertExists("an unmarked rule is linked there as well", filepath.Join(rules, "good-always.md"))
	g.assertAbsent("an opt-in rule is not linked there", filepath.Join(rules, "opt-in-false.md"))
	g.assertExists("the agent baseline is installed", filepath.Join(rules, "agent-baseline.md"))
	// The pointer in the block is only as good as the file it names.
	g.assertContains("the linked full text has what the block dropped", filepath.Join(rules, "cored.md"), "BODY-AFTER-CORE")

	hooks := filepath.Join(home, ".claude/hooks")
	g.assertExists("the enforcement hook is installed", filepath.Join(hooks, "enforce-agent-baseline.py"))
	g.assertContains("an unregistered hook is reported, not assumed", g.log, "NOT registered")

	g.assertExists("the flow-active-change hook is installed", filepath.Join(hooks, "flow-active-change.py"))
	g.assertContains("an unregistered flow-active-change hook is reported", g.log,
		"flow-active-change hook is installed but NOT registered")

	g.assertExists("the protect-main-checkout hook is installed", filepath.Join(hooks, "protect-main-checkout.py"))

	// Every hook's warning and every printed registration snippet the installers emit is
	// pinned output: the shared registration-warning helper must keep each line byte-identical.
	g.assertContains("the unregistered protect-main-checkout hook is reported", g.log,
		"protect-main-checkout hook is installed but NOT registered")
	g.assertContains("the protect-main-checkout registration snippet is printed", g.log, ".claude/hooks/protect-main-checkout.py")
	g.assertContains("the claude agent-baseline registration snippet is printed", g.log, ".claude/hooks/enforce-agent-baseline.py")
	g.assertContains("the claude flow-active-change registration snippet is printed", g.log, ".claude/hooks/flow-active-change.py")

	// An unbalanced marker renders the wrong amount of text into every block, so it must abort
	// the run rather than ship a half-rule.
	unbalanced := filepath.Join(g.dir, "fixture-repo-unbalanced-core")
	if err := makeFixtureRepo(unbalanced, false); err != nil {
		t.Fatal(err)
	}
	g.seedFile(filepath.Join(unbalanced, "rules/unbalanced.mdc"), 0o644,
		"---\ndescription: fixture unbalanced core\nalwaysApply: true\n---\n\n# Unbalanced\n<!-- core -->\nBODY-UNBALANCED\n")
	home = g.newHome()
	g.runSetup(unbalanced, home, "global", "")
	g.assertRCNonzero("an unbalanced core marker aborts the run", g.rc)
	g.assertContains("the abort names the offending rule", g.log, "unbalanced.mdc")
	g.assertAbsent("nothing was installed: CLAUDE.md", filepath.Join(home, ".claude/CLAUDE.md"))
}

func TestKotlinRuleInstalledByNoMode(t *testing.T) {
	g := newGroup(t)
	if fi, err := os.Stat(filepath.Join(repoRoot, "rules", kotlinRule)); err != nil || !fi.Mode().IsRegular() {
		g.fail("the opt-in Kotlin rule is present to be tested", filepath.Join(repoRoot, "rules", kotlinRule)+" is missing")
		return
	}
	for _, mode := range []string{"global", "claude-code", "zcode"} {
		home := g.newHome()
		proj := filepath.Join(g.dir, fmt.Sprintf("project-%d", g.seq))
		g.runSetup(repoRoot, home, mode, proj)
		g.assertRCZero("mode "+mode+" installs cleanly", g.rc, g.log)
		var hits []string
		for _, root := range []string{home, proj} {
			_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
				if err == nil && d.Name() == kotlinRule {
					hits = append(hits, p)
				}
				return nil
			})
		}
		g.assertEq("mode "+mode+" installs no "+kotlinRule, "", strings.Join(hits, "\n"))
		// Only `global` writes a managed block. For the other two modes the files do not exist,
		// and assertNotContains passes on a missing file — so asserting "no Kotlin rule inlined"
		// there proves nothing. Assert the real property per mode instead: global must have the
		// block WITHOUT the opt-in rule; the rest must have no block.
		for _, rel := range []string{".claude/CLAUDE.md", ".zcode/AGENTS.md"} {
			managed := filepath.Join(home, rel)
			if mode == "global" {
				g.assertFileExists("mode "+mode+": "+rel+" exists", managed)
				g.assertNotContains("mode "+mode+": "+rel+" has no inlined Kotlin rule", managed, "<!-- rule: "+kotlinRule+" -->")
			} else {
				g.assertAbsent("mode "+mode+" writes no managed block at "+rel, managed)
			}
		}
	}
}

func TestGlobalInstallPopulatesSkillDirs(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(repoRoot, home, "global", "")
	g.assertRCZero("the real-repo global install succeeds", g.rc, g.log)

	expectedSkills := len(skillDirs(t))
	if expectedSkills == 0 {
		g.fail("the repo has skills to install", filepath.Join(repoRoot, "skills")+" contains no directories")
	}
	for _, rel := range []string{".claude/skills", ".zcode/skills"} {
		skillsDir := filepath.Join(home, rel)
		installed, linked := 0, 0
		for _, name := range visibleEntries(skillsDir) {
			installed++
			entry := filepath.Join(skillsDir, name)
			if fi, err := os.Stat(entry); isSymlink(entry) && err == nil && fi.IsDir() {
				linked++
			}
		}
		g.assertEq(rel+" holds every skill", expectedSkills, installed)
		g.assertEq(rel+" holds only resolvable symlinks", expectedSkills, linked)
	}

	var dangling []string
	_ = filepath.WalkDir(home, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.Type()&fs.ModeSymlink != 0 {
			if _, statErr := os.Stat(p); statErr != nil {
				dangling = append(dangling, p)
			}
		}
		return nil
	})
	g.assertEq("a global install leaves zero dangling symlinks", "", strings.Join(dangling, "\n"))

	// flow-contracts must install like any other skill, through install_skills, and must NOT be
	// inlined into the managed block — that is the whole point of extracting it from the
	// always-on rule.
	g.assertExists("flow-contracts installs into .claude/skills", filepath.Join(home, ".claude/skills/flow-contracts/SKILL.md"))
	g.assertExists("flow-contracts installs into .zcode/skills", filepath.Join(home, ".zcode/skills/flow-contracts/SKILL.md"))
	g.assertExists("flow-contracts installs state-file.md", filepath.Join(home, ".claude/skills/flow-contracts/state-file.md"))
	g.assertNotContains("the managed CLAUDE.md does not inline the state-file write template",
		filepath.Join(home, ".claude/CLAUDE.md"), `PROJECT_KEY="$(basename`)
	g.assertNotContains("the managed AGENTS.md does not inline the state-file write template",
		filepath.Join(home, ".zcode/AGENTS.md"), `PROJECT_KEY="$(basename`)

	// The commit-scope-is-the-module rule is an always-on rule like any other: its core excerpt
	// renders into the managed block with a pointer to the installed full text, and the full
	// text itself is a symlink into the repository at ~/.claude/rules/, same as
	// build-the-simplest-thing.mdc and every other always-on rule checked above.
	g.assertContains("the block carries the commit-scope rule's core excerpt", filepath.Join(home, ".claude/CLAUDE.md"),
		"A commit's scope names the module or area inside the repository that the commit moved")
	g.assertContains("the block points at the installed commit-scope rule", filepath.Join(home, ".claude/CLAUDE.md"),
		"Full rule: `~/.claude/rules/commit-scope-is-the-module.md`.")
	commitScope := filepath.Join(home, ".claude/rules/commit-scope-is-the-module.md")
	g.assertExists("the commit-scope rule's full text is installed", commitScope)
	g.assertSymlink("the installed commit-scope rule is a symlink into the repo", commitScope)
}

// TestGuardsReachInstallAndRunFromIt is two bash groups in one test: the second runs the guard
// out of the first's HOME.
//
// KAN-73's task 5 guard (check-guard-symlinks.sh) checks the REPOSITORY: that every
// skills/*/scripts/ entry is a symlink that resolves. This checks the INSTALL — the only
// assertion that would have caught the original bug, because the symlinks could be perfect in
// the repo and still not reach ~/.claude/skills/ or ~/.zcode/skills/ if install_skills() ever
// stopped carrying them.
//
// The expected guard list is read from the repository tree itself (skills/<skill>/scripts/*)
// rather than hardcoded here — a hardcoded copy goes stale the moment a guard is added or
// removed.
func TestGuardsReachInstallAndRunFromIt(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(repoRoot, home, "global", "")
	g.assertRCZero("the global install for the guard-reachability check succeeds", g.rc, g.log)

	// Derived, not hardcoded: every skill that actually carries a scripts/ directory. A
	// hardcoded list goes stale the moment a skill gains or loses one — as it did when
	// flow-status's directory was dropped for invoking no guard.
	var guardSkills []string
	for _, skill := range visibleEntries(filepath.Join(repoRoot, "skills")) {
		if fi, err := os.Stat(filepath.Join(repoRoot, "skills", skill, "scripts")); err == nil && fi.IsDir() {
			guardSkills = append(guardSkills, skill)
		}
	}
	if len(guardSkills) == 0 {
		g.fail("at least one skill carries a scripts/ directory", "none found under "+repoRoot+"/skills/*/scripts")
	}
	for _, skill := range guardSkills {
		skillScripts := filepath.Join(repoRoot, "skills", skill, "scripts")
		for _, name := range visibleEntries(skillScripts) {
			entry := filepath.Join(skillScripts, name)
			// A guard's own executable bit is read from its real source, not assumed: a sourced
			// companion is never executed directly and need not be marked executable in the
			// repository either.
			fi, err := os.Stat(entry)
			wantExec := err == nil && fi.Mode().IsRegular() && executable(entry)
			for _, harness := range []string{".claude", ".zcode"} {
				rel := harness + "/skills/" + skill + "/scripts/" + name
				installed := filepath.Join(home, rel)
				g.assertExists(rel+" is present at the installed path", installed)
				if wantExec {
					g.assertExecutable(rel+" is executable at the installed path", installed)
				}
			}
		}
	}

	// check-panel-reproducers.sh runs through the installed path and finds its dependencies.
	//
	// A flow-guard shim: it sources lib/flow-guard.sh from its own directory and execs the
	// flow-guard binary built from the stats/ tree beside it. Presence is not reachability —
	// this actually RUNS it through the installed symlink chain and inspects what it printed,
	// not just that the file exists.
	guardPath := filepath.Join(home, ".claude/skills/flow/scripts/check-panel-reproducers.sh")
	g.assertExists("the installed check-panel-reproducers.sh exists to invoke", guardPath)
	guardLog := filepath.Join(g.dir, "check-panel-reproducers-installed.log")
	// One argument (a real directory, standing in for the worktree) and no change name: the
	// guard checks the worktree argument before the change name, so a true zero-argument call
	// would exit on the worktree check first. Supplying just the worktree still exercises the
	// dependencies this is about — the shim loads lib/flow-guard.sh and runs the built binary
	// before either argument is validated — while reaching the guard's own usage line.
	var out bytes.Buffer
	cmd := exec.Command(guardPath, g.dir)
	// The shim builds flow-guard into this cache; unset, it would build into the operator's
	// ~/.cache, outside the sandbox.
	guardCache := filepath.Join(g.dir, "flow-guard-cache")
	cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache)
	cmd.Stdout, cmd.Stderr = &out, &out
	guardRC := 0
	if err := cmd.Run(); err != nil {
		guardRC = -1
		if ee, ok := err.(*exec.ExitError); ok {
			guardRC = ee.ExitCode()
		}
	}
	if err := os.WriteFile(guardLog, out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	// No change name: the guard's own documented behaviour is to print a usage line and exit 2.
	// That is expected here and is NOT a failure — the failure this proves the absence of is
	// the shim reporting that it could not load lib/flow-guard.sh or build flow-guard.
	g.assertRCNonzero("check-panel-reproducers.sh with no change name exits non-zero (expected — not a failure)", guardRC)
	g.assertContains("check-panel-reproducers.sh reaches its own usage message (proves lib/flow-guard.sh loaded and the flow-guard binary ran, not just that the file exists)",
		guardLog, "usage: check-panel-reproducers.sh")
	g.assertNotContains("check-panel-reproducers.sh through the installed path does not report a missing lib/flow-guard.sh",
		guardLog, "cannot load lib/flow-guard.sh")
	built, _ := filepath.Glob(filepath.Join(guardCache, "*", "flow-guard"))
	g.assertEq("the installed guard built flow-guard inside the sandbox, not in the operator's cache", 1, len(built))
}

func TestPreexistingSkillDirMovedOut(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	skills := skillDirs(t)
	if len(skills) == 0 {
		t.Fatal("cannot pick a skill name to displace")
	}
	victim := skills[0]
	g.seedStaleSkillDir(filepath.Join(home, ".claude/skills", victim))

	g.runSetup(repoRoot, home, "global", "")
	g.assertRCZero("the install over a pre-existing skill directory succeeds", g.rc, g.log)
	g.assertSymlink("the displaced skill is now a symlink", filepath.Join(home, ".claude/skills", victim))
	// `<name>.bak` beside it would still be discovered by a SKILL.md walk — two loadable skills
	// declaring the same name. The backup has to leave the scanned tree entirely.
	g.assertAbsent("no <name>.bak is left inside the skills tree", filepath.Join(home, ".claude/skills", victim+".bak"))
	g.assertEq("the stale copy is preserved outside the scanned tree", 1, sentinelHits(filepath.Join(home, ".claude/skills-backup")))
	g.assertEq("the stale copy is no longer reachable from the skills tree", 0, sentinelHits(filepath.Join(home, ".claude/skills")))
}

// sentinelHits counts the SKILL.md files under root, links followed, carrying the stale-copy
// sentinel.
func sentinelHits(root string) int {
	n := 0
	for _, p := range findFollow(root, "SKILL.md") {
		if b, err := os.ReadFile(p); err == nil && bytes.Contains(b, []byte("SENTINEL-PREEXISTING-SKILL")) {
			n++
		}
	}
	return n
}

// visibleEntries is a shell `dir/*`: the non-dot entry names, sorted.
func visibleEntries(dir string) []string {
	entries, _ := os.ReadDir(dir)
	var names []string
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	return names
}

// skillDirs is `skills/*/`: the non-dot entries of the repo's skills/ that stat as directories.
func skillDirs(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, name := range visibleEntries(filepath.Join(repoRoot, "skills")) {
		if fi, err := os.Stat(filepath.Join(repoRoot, "skills", name)); err == nil && fi.IsDir() {
			out = append(out, name)
		}
	}
	return out
}
