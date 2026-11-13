package setuptest

import (
	"bytes"
	"fmt"
	"github.com/tweety53/agents/stats/internal/smoke"
	"io/fs"
	"os"
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
// smoke: core — the lightest run of the real setup.sh, over a fixture tree
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
	g.assertExists("the enforce-loop-dispatch hook is installed", filepath.Join(hooks, "enforce-loop-dispatch.py"))
	g.assertContains("an unregistered enforce-loop-dispatch hook is reported", g.log,
		"enforce-loop-dispatch hook is installed but NOT registered")
	g.assertContains("the claude enforce-loop-dispatch registration snippet is printed", g.log, ".claude/hooks/enforce-loop-dispatch.py")

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
	smoke.Heavy(t)
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

// smoke: core — the real setup.sh global over the real tree
func TestGlobalInstallPopulatesSkillDirs(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(repoRoot, home, "global", "")
	g.assertRCZero("the real-repo global install succeeds", g.rc, g.log)

	expectedSkills := len(repoDirs(t, "skills"))
	// install_mods links each mods/<name>/ beside the skills in Claude Code's folder alone.
	expected := map[string]int{".claude/skills": expectedSkills + len(repoDirs(t, "mods")), ".zcode/skills": expectedSkills}
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
		g.assertEq(rel+" holds every skill", expected[rel], installed)
		g.assertEq(rel+" holds only resolvable symlinks", expected[rel], linked)
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

func TestPreexistingSkillDirMovedOut(t *testing.T) {
	smoke.Heavy(t)
	g := newGroup(t)
	home := g.newHome()
	skills := repoDirs(t, "skills")
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
// repoDirs lists the directories directly under <repo>/<dir>: "skills", or "mods" (absent → none).
func repoDirs(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	for _, name := range visibleEntries(filepath.Join(repoRoot, dir)) {
		if fi, err := os.Stat(filepath.Join(repoRoot, dir, name)); err == nil && fi.IsDir() {
			out = append(out, name)
		}
	}
	return out
}

func TestRegisteredHookTimeoutIsChecked(t *testing.T) {
	smoke.Heavy(t)
	g := newGroup(t)
	home := g.newHome()
	g.seedFile(filepath.Join(home, ".claude/settings.json"), 0o644, `{"hooks":{"PreToolUse":[
  {"matcher":"Bash","hooks":[{"type":"command","command":"python3 \"$HOME/.claude/hooks/protect-main-checkout.py\""}]},
  {"matcher":"Agent|Task","hooks":[{"type":"command","command":"python3 \"$HOME/.claude/hooks/enforce-loop-dispatch.py\"","timeout":300}]}
]}}`)
	g.runSetup(repoRoot, home, "global", "")
	g.assertRCZero("the install succeeds", g.rc, g.log)
	g.assertContains("a registered hook missing its timeout is reported", g.log,
		"protect-main-checkout hook is registered in "+filepath.Join(home, ".claude/settings.json")+` without "timeout": 300`)
	g.assertNotContains("a registered hook with its timeout is not reported", g.log,
		"enforce-loop-dispatch hook is registered in")
}
