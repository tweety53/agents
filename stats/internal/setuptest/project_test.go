package setuptest

import (
	"fmt"
	"path/filepath"
	"testing"
)

// Project-level opt-in rule rendering.
//
// An opt-in rule is installed nowhere by path, so before this existed the only way a project
// could actually load one was to paste a copy into its own CLAUDE.md and AGENTS.md — two copies
// with nothing keeping them in step with the rule they came from. The installer now renders
// whatever the project's `.flow/project.md ## standards` section names into a managed block in
// both files.
//
// Everything below is a containment or convergence claim about a file the user WROTE (a
// project's CLAUDE.md is hand-maintained far more often than ~/.claude/CLAUDE.md is), so the
// same guarantees the global block carries have to hold here too.

// projectDir is the case's project directory, numbered after the case's HOME.
func (g *group) projectDir(kind string) string {
	return filepath.Join(g.dir, fmt.Sprintf("project-%s-%d", kind, g.seq))
}

func TestProjectOptInRuleRendered(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	proj := g.projectDir("optin")
	g.seedProjectMD(proj, "CLAUDE.md", "opt-in-false.mdc", "good-always.mdc")
	// `claude-code` copies CLAUDE.md in but never AGENTS.md, so AGENTS.md exists below only
	// because the standards rendering created it; CLAUDE.md's rule-body assertions are what
	// prove the rendering reached it.
	g.runSetup(fixture, home, "claude-code", proj)
	g.assertRCZero("the project install succeeds", g.rc, g.log)
	for _, label := range []string{"CLAUDE.md", "AGENTS.md"} {
		f := filepath.Join(proj, label)
		g.assertFileExists(label+" exists after the install", f)
		g.assertContains(label+" carries the opted-in rule body", f, "BODY-OPT-IN")
		g.assertContains(label+" labels the rule it inlined", f, "<!-- rule: opt-in-false.mdc -->")
		g.assertEq(label+" has exactly one begin", 1, countLinesMatching(f, begin))
		g.assertEq(label+" has exactly one end", 1, countLinesMatching(f, end))
		// An always-on rule already reaches every session through the global block. Rendering
		// it again here is the duplication this feature exists to remove, not a harmless extra.
		g.assertNotContains(label+" does not re-render the always-on rule", f, "BODY-GOOD-ALWAYS")
		// `CLAUDE.md` is a bare non-.mdc entry — the project's own file, not a shared rule.
		g.assertNotContains(label+" treats a bare non-.mdc entry as no shared rule", f, "<!-- rule: CLAUDE.md -->")
	}
	g.assertContains("the run reports which rule it rendered", g.log, "opt-in-false.mdc")
}

func TestProjectRenderingIdempotent(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	proj := g.projectDir("idem")
	g.seedProjectMD(proj, "opt-in-false.mdc")
	claudeMD, agentsMD := filepath.Join(proj, "CLAUDE.md"), filepath.Join(proj, "AGENTS.md")
	g.seedFile(claudeMD, 0o640, "# Gymie-like project instructions\n\nPROJECT-HANDWRITTEN — this line predates the managed block.\n")
	original := filepath.Join(g.dir, "project-handwritten-original.md")
	copyFile(t, claudeMD, original)
	claude2 := filepath.Join(g.dir, "project-idem-claudefile-run2.md")
	agents2 := filepath.Join(g.dir, "project-idem-agentsfile-run2.md")
	for run := 1; run <= 3; run++ {
		g.runSetup(fixture, home, "claude-code", proj)
		g.assertRCZero(fmt.Sprintf("project run %d succeeds", run), g.rc, g.log)
		g.assertContains(fmt.Sprintf("project run %d keeps the hand-written line", run), claudeMD, "PROJECT-HANDWRITTEN")
		g.assertEq(fmt.Sprintf("project run %d: exactly one begin in CLAUDE.md", run), 1, countLinesMatching(claudeMD, begin))
		g.assertEq(fmt.Sprintf("project run %d: exactly one begin in AGENTS.md", run), 1, countLinesMatching(agentsMD, begin))
		g.assertEq(fmt.Sprintf("project run %d: CLAUDE.md keeps its 640 mode", run), "640", fileMode(claudeMD))
		if run == 2 {
			copyFile(t, claudeMD, claude2)
			copyFile(t, agentsMD, agents2)
		}
	}
	g.assertIdentical("project run 3 leaves CLAUDE.md byte-identical to run 2", claudeMD, claude2)
	g.assertIdentical("project run 3 leaves AGENTS.md byte-identical to run 2", agentsMD, agents2)
	g.assertIdentical("the project's pre-install copy is the file as the user wrote it", claudeMD+".flow.bak", original)
}

func TestModesRenderingProjectStandards(t *testing.T) {
	g := newGroup(t)
	for _, mode := range []string{"claude-code", "zcode"} {
		home := g.newHome()
		proj := g.projectDir("mode")
		g.seedProjectMD(proj, "opt-in-false.mdc")
		g.runSetup(fixture, home, mode, proj)
		g.assertRCZero("mode "+mode+" installs cleanly over a project with standards", g.rc, g.log)
		g.assertContains("mode "+mode+" renders the rule into CLAUDE.md", filepath.Join(proj, "CLAUDE.md"), "BODY-OPT-IN")
		g.assertContains("mode "+mode+" renders the rule into AGENTS.md", filepath.Join(proj, "AGENTS.md"), "BODY-OPT-IN")
	}

	// `global` installs no project files at all, so it must not create these either — otherwise
	// a user-level install would start writing into whatever directory it happened to be run
	// from.
	home := g.newHome()
	proj := g.projectDir("global")
	g.seedProjectMD(proj, "opt-in-false.mdc")
	g.runSetup(fixture, home, "global", proj)
	g.assertRCZero("global installs cleanly beside a project with standards", g.rc, g.log)
	g.assertAbsent("global writes no project CLAUDE.md", filepath.Join(proj, "CLAUDE.md"))
	g.assertAbsent("global writes no project AGENTS.md", filepath.Join(proj, "AGENTS.md"))
}

func TestProjectWithNothingToRenderLeftAlone(t *testing.T) {
	g := newGroup(t)
	expected := filepath.Join(g.dir, "noblock-claude-expected.md")
	// CLAUDE.md is seeded as the user wrote it, so claude-code's own copy step skips it and any
	// change afterwards is the standards rendering's doing. AGENTS.md is left unseeded:
	// claude-code never copies it, so its absence afterwards proves the rendering created none.
	seedInstructionFiles := func(proj string) {
		g.seedFile(filepath.Join(proj, "CLAUDE.md"), 0o644, "# hand-written CLAUDE.md\n")
		copyFile(t, filepath.Join(proj, "CLAUDE.md"), expected)
	}

	// No .flow/project.md at all.
	home := g.newHome()
	proj := g.projectDir("noconfig")
	seedInstructionFiles(proj)
	g.runSetup(fixture, home, "claude-code", proj)
	g.assertRCZero("a project with no .flow/project.md installs cleanly", g.rc, g.log)
	g.assertIdentical("CLAUDE.md is untouched for a project with no config", filepath.Join(proj, "CLAUDE.md"), expected)
	g.assertAbsent("no AGENTS.md is created for a project with no config", filepath.Join(proj, "AGENTS.md"))

	// A project.md that names only its own files — no shared rule to render, so no block.
	home = g.newHome()
	proj = g.projectDir("noshared")
	g.seedProjectMD(proj, "CLAUDE.md", "CONTRIBUTING.md", "docs/standards/api.mdc")
	seedInstructionFiles(proj)
	g.runSetup(fixture, home, "claude-code", proj)
	g.assertRCZero("a project naming no shared rule installs cleanly", g.rc, g.log)
	g.assertIdentical("no block is written when no entry resolves to the shared library", filepath.Join(proj, "CLAUDE.md"), expected)
	g.assertAbsent("no AGENTS.md is created for a project naming no shared rule", filepath.Join(proj, "AGENTS.md"))
	// `docs/standards/api.mdc` contains a `/`, so it is a project path — form 3, never the
	// shared library. Resolving it there would be the containment bypass.
	g.assertNotContains("a slashed .mdc entry is not looked up in the shared library", g.log, "docs/standards/api.mdc")
}

func TestMissingNamedRuleReportedAndSkipped(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	proj := g.projectDir("missing")
	g.seedProjectMD(proj, "no-such-rule.mdc", "opt-in-false.mdc")
	g.runSetup(fixture, home, "claude-code", proj)
	g.assertContains("the missing rule is reported by name", g.log, "no-such-rule.mdc")
	g.assertContains("the rules that do exist are still rendered", filepath.Join(proj, "CLAUDE.md"), "BODY-OPT-IN")
	g.assertNotContains("a bullet inside a fenced block is not read as an entry", g.log, "fenced-not-an-entry.mdc")
	// Reported, not fatal: the rest of the install completed. But a run that could not deliver
	// something the project asked for is not a clean install either, and the exit status is
	// where that has to show — same contract as every other skip in this installer.
	g.assertRCNonzero("a missing rule makes the run report a skip in its exit status", g.rc)
	g.assertContains("both instruction files still got the rules that do exist", filepath.Join(proj, "AGENTS.md"), "BODY-OPT-IN")
}

func TestProjectDelimiterGuardsFire(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	proj := g.projectDir("reversed")
	g.seedProjectMD(proj, "opt-in-false.mdc")
	claudeMD := filepath.Join(proj, "CLAUDE.md")
	g.seedFile(claudeMD, 0o644, "# Project notes\n\n"+end+"\ngenerated-looking content\n"+begin+"\n\nkeep-me-below\n")
	expected := filepath.Join(g.dir, "project-reversed-expected.md")
	copyFile(t, claudeMD, expected)
	g.runSetup(fixture, home, "claude-code", proj)
	g.assertRCNonzero("reversed delimiters in a project file abort the run", g.rc)
	g.assertContains("the project abort names the delimiter shape", g.log, "not one begin followed by one end")
	g.assertIdentical("reversed: the project CLAUDE.md is byte-identical afterwards", claudeMD, expected)
	g.assertAbsent("reversed: no project .flow.bak was written", claudeMD+".flow.bak")
}

func TestRealKotlinStandardReachesProject(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	proj := g.projectDir("kotlin")
	g.seedProjectMD(proj, "CLAUDE.md", kotlinRule)
	g.runSetup(repoRoot, home, "claude-code", proj)
	g.assertRCZero("the real-repo project install succeeds", g.rc, g.log)
	for _, label := range []string{"CLAUDE.md", "AGENTS.md"} {
		f := filepath.Join(proj, label)
		g.assertContains(label+" inlines the Kotlin standard", f, "<!-- rule: "+kotlinRule+" -->")
		g.assertContains(label+" carries the Kotlin standard's text", f, "Kotlin Backend Development Standard")
		// The frontmatter is stripped exactly as it is for the global block; leaving `globs:`
		// or `alwaysApply:` in the rendered body would feed rule-file frontmatter to every
		// harness.
		g.assertNotContains(label+" strips the rule's frontmatter", f, "alwaysApply: false")
		// The standard's command examples are placeholders resolved from the project's own
		// ## lint section: the rendered block carries the declared commands, never the
		// hardcoded ./gradlew ktlintCheck detekt copies that drifted from every project
		// that verifies differently, and never a raw placeholder.
		g.assertContains(label+" renders the declared lint commands as a bash fence", f, "```bash\n./gradlew ktlintFormat")
		g.assertContains(label+" renders the second declared command", f, "./gradlew verifyChange")
		g.assertNotContains(label+" carries no stale hardcoded command", f, "./gradlew ktlintCheck detekt")
		g.assertNotContains(label+" carries no raw placeholder", f, "{{lint-commands}}")
	}
}

// Lint-command substitution: an opt-in rule body may carry {{lint-commands}}, which the
// project render replaces with the commands declared in the project's own
// `.flow/project.md ## lint` section — so a managed block can never name a command the
// project does not declare.
func TestLintCommandsSubstitutedFromProjectLint(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	proj := g.projectDir("lintsub")
	g.seedProjectMD(proj, "lint-placeholder.mdc")
	g.runSetup(fixture, home, "claude-code", proj)
	g.assertRCZero("the project install succeeds", g.rc, g.log)
	for _, label := range []string{"CLAUDE.md", "AGENTS.md"} {
		f := filepath.Join(proj, label)
		g.assertContains(label+" renders the rule body", f, "BODY-LINT-PLACEHOLDER")
		g.assertContains(label+" renders the declared lint commands as a bash fence", f, "```bash\n./gradlew ktlintFormat")
		g.assertContains(label+" renders the second declared command", f, "./gradlew verifyChange")
		g.assertNotContains(label+" renders no raw placeholder", f, "{{lint-commands}}")
	}
}

// A project opting into a placeholder rule without declaring ## lint must abort the run
// before any target is touched, naming the missing section — never render an empty or
// invented command block.
func TestLintCommandsRefusedWithoutProjectLint(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	proj := g.projectDir("lintnolint")
	g.seedFile(filepath.Join(proj, ".flow/project.md"), 0o644,
		"# flow project configuration — fixture project\n\n## standards\n\n- `lint-placeholder.mdc` — seeded by the harness\n")
	g.runSetup(fixture, home, "claude-code", proj)
	g.assertRCNonzero("a placeholder rule with no ## lint to render from aborts the run", g.rc)
	g.assertContains("the abort names the missing section", g.log, "## lint")
	g.assertNotContains("nothing was rendered into CLAUDE.md", filepath.Join(proj, "CLAUDE.md"), "BODY-LINT-PLACEHOLDER")
}
