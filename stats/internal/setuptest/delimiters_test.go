package setuptest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Malformed delimiters must abort with the target untouched.
//
// These are the two data-loss defects. Each case seeds a managed file with a shape the
// installer must refuse, runs `global`, and requires both a non-zero exit AND a target that is
// byte-for-byte what it was. "Aborted" is not enough on its own: the reversed-delimiter defect
// aborted nothing and deleted content, and a future one could abort *after* writing.
func TestMalformedDelimitersAbortByteIdentical(t *testing.T) {
	g := newGroup(t)

	// --- reversed: end above begin. Counts 1 and 1, which is what made this dangerous.
	home := g.newHome()
	target := filepath.Join(home, ".claude/CLAUDE.md")
	g.seedFile(target, 0o644, "# My own notes\nkeep-me-above\n\n"+end+"\ngenerated-looking content\n"+begin+
		"\n\n# More of my own notes\nkeep-me-below\n")
	expected := filepath.Join(g.dir, "reversed-expected.md")
	copyFile(t, target, expected)
	g.runSetup(fixture, home, "global", "")
	g.assertRCNonzero("reversed delimiters abort the run", g.rc)
	g.assertContains("the abort names the delimiter shape", g.log, "not one begin followed by one end")
	g.assertIdentical("reversed: CLAUDE.md is byte-identical afterwards", target, expected)
	g.assertAbsent("reversed: no .flow.bak was written", target+".flow.bak")

	// --- duplicated: two begins and two ends, seeded in the ZCODE target so the second
	// managed file is exercised too (the first one is written before this is reached).
	home = g.newHome()
	target = filepath.Join(home, ".zcode/AGENTS.md")
	g.seedFile(target, 0o644, "# My own notes\n\n"+begin+"\nfirst block\n"+end+"\n\n"+begin+"\nsecond block\n"+end+"\n")
	expected = filepath.Join(g.dir, "duplicated-expected.md")
	copyFile(t, target, expected)
	g.runSetup(fixture, home, "global", "")
	g.assertRCNonzero("duplicated delimiters abort the run", g.rc)
	g.assertIdentical("duplicated: AGENTS.md is byte-identical afterwards", target, expected)
	g.assertAbsent("duplicated: no .flow.bak was written", target+".flow.bak")

	// --- CRLF markers: the installer matches LF-terminated lines, so these count zero and
	// would take the append branch. It must stop and say how to fix the file.
	home = g.newHome()
	target = filepath.Join(home, ".claude/CLAUDE.md")
	expected = filepath.Join(g.dir, "crlf-expected.md")
	if err := os.WriteFile(expected, []byte("# My own notes\r\n\r\n"+begin+"\r\nold block\r\n"+end+"\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mkdirAll(t, filepath.Dir(target))
	copyFile(t, expected, target)
	g.runSetup(fixture, home, "global", "")
	g.assertRCNonzero("CRLF delimiters abort the run", g.rc)
	g.assertContains("the CRLF abort carries the remediation hint", g.log, "Convert the file to LF endings")
	g.assertIdentical("CRLF: CLAUDE.md is byte-identical afterwards", target, expected)
}

// An unbalanced pair — a lone begin or a lone end — is the second data-loss shape: it once took
// the append branch, which never converges (1/0 became 2/1 became 3/2). It must abort with the
// target untouched, in either managed file. The bash harness never seeded this shape.
func TestUnbalancedDelimitersAbortByteIdentical(t *testing.T) {
	g := newGroup(t)
	for _, c := range []struct{ name, rel, body string }{
		{"lone begin", ".claude/CLAUDE.md", "# My own notes\n\n" + begin + "\nhalf a block\n"},
		{"lone end", ".zcode/AGENTS.md", "# My own notes\n\nhalf a block\n" + end + "\n"},
	} {
		home := g.newHome()
		target := filepath.Join(home, c.rel)
		g.seedFile(target, 0o644, c.body)
		expected := filepath.Join(g.dir, strings.ReplaceAll(c.name, " ", "-")+"-expected.md")
		copyFile(t, target, expected)
		g.runSetup(fixture, home, "global", "")
		g.assertRCNonzero(c.name+": the run aborts", g.rc)
		g.assertIdentical(c.name+": "+c.rel+" is byte-identical afterwards", target, expected)
		g.assertAbsent(c.name+": no .flow.bak was written", target+".flow.bak")
	}
}

// Repeated runs converge, and the user's own content survives them.
//
// The append-forever defect showed up only on the SECOND and THIRD run, so one run proves
// nothing here. Three runs, counting the delimiters in both managed files each time, and
// requiring run 3 to be byte-identical to run 2 — the definition of converged.
func TestThreeGlobalRunsConverge(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	claudeMD := filepath.Join(home, ".claude/CLAUDE.md")
	zcodeMD := filepath.Join(home, ".zcode/AGENTS.md")
	claude2 := filepath.Join(g.dir, "converge-claude-run2.md")
	zcode2 := filepath.Join(g.dir, "converge-zcode-run2.md")
	for run := 1; run <= 3; run++ {
		g.runSetup(fixture, home, "global", "")
		g.assertRCZero(fmt.Sprintf("run %d succeeds", run), g.rc, g.log)
		g.assertEq(fmt.Sprintf("run %d: exactly one begin in CLAUDE.md", run), 1, countLinesMatching(claudeMD, begin))
		g.assertEq(fmt.Sprintf("run %d: exactly one end in CLAUDE.md", run), 1, countLinesMatching(claudeMD, end))
		g.assertEq(fmt.Sprintf("run %d: exactly one begin in AGENTS.md", run), 1, countLinesMatching(zcodeMD, begin))
		g.assertEq(fmt.Sprintf("run %d: exactly one end in AGENTS.md", run), 1, countLinesMatching(zcodeMD, end))
		if run == 2 {
			copyFile(t, claudeMD, claude2)
			copyFile(t, zcodeMD, zcode2)
		}
	}
	g.assertIdentical("run 3 leaves CLAUDE.md byte-identical to run 2", claudeMD, claude2)
	g.assertIdentical("run 3 leaves AGENTS.md byte-identical to run 2", zcodeMD, zcode2)
}

func TestHandWrittenContentAndModeSurvive(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	claudeMD := filepath.Join(home, ".claude/CLAUDE.md")
	// Mode 640 rather than 600 or 644 on purpose: those are what a temp file or a shell
	// redirect produce, so a rewrite that replaced the file wholesale, or a backup taken with a
	// redirect, could land on them by accident and these assertions could never fail. 640
	// matches nothing the installer creates on its own. A plain copy keeps the source's mode
	// under a 022 umask, so the mode assertions do not tell `cp` from `cp -p`.
	g.seedFile(claudeMD, 0o640, "# My own global instructions\n\nHANDWRITTEN-ABOVE — this line predates flow.\n")
	original := filepath.Join(g.dir, "handwritten-original.md")
	copyFile(t, claudeMD, original)

	g.runSetup(fixture, home, "global", "")
	g.assertRCZero("first run over a hand-written file succeeds", g.rc, g.log)
	g.assertContains("run 1 keeps the hand-written line", claudeMD, "HANDWRITTEN-ABOVE")
	g.assertExists("run 1 takes a pre-install backup", claudeMD+".flow.bak")
	g.assertIdentical("the backup is the file as it was before flow touched it", claudeMD+".flow.bak", original)

	// Content added AFTER the end marker exercises the rewrite branch's blast radius — this is
	// what the reversed-delimiter defect destroyed.
	appendTo(t, claudeMD, "\nHANDWRITTEN-BELOW — added after the managed block.\n")

	run2 := filepath.Join(g.dir, "handwritten-run2.md")
	for run := 2; run <= 3; run++ {
		g.runSetup(fixture, home, "global", "")
		g.assertRCZero(fmt.Sprintf("run %d over a hand-written file succeeds", run), g.rc, g.log)
		g.assertContains(fmt.Sprintf("run %d keeps content above the block", run), claudeMD, "HANDWRITTEN-ABOVE")
		g.assertContains(fmt.Sprintf("run %d keeps content below the block", run), claudeMD, "HANDWRITTEN-BELOW")
		g.assertEq(fmt.Sprintf("run %d: still exactly one begin", run), 1, countLinesMatching(claudeMD, begin))
		g.assertEq(fmt.Sprintf("run %d: still exactly one end", run), 1, countLinesMatching(claudeMD, end))
		g.assertEq(fmt.Sprintf("run %d: the target keeps its 640 mode", run), "640", fileMode(claudeMD))
		g.assertEq(fmt.Sprintf("run %d: the .flow.bak keeps the 640 mode", run), "640", fileMode(claudeMD+".flow.bak"))
		g.assertIdentical(fmt.Sprintf("run %d: the pre-install backup is never overwritten", run), claudeMD+".flow.bak", original)
		if run == 2 {
			copyFile(t, claudeMD, run2)
		}
	}
	g.assertIdentical("run 3 leaves the hand-written file byte-identical to run 2", claudeMD, run2)
}

// All-or-nothing preflight.
//
// A rule the renderer must refuse has to be refused BEFORE anything is installed. A partial
// install is the worst outcome available: one harness picks up rules the other never
// receives, and every later run reproduces the split identically.
func TestDelimiterRuleInstallsNothing(t *testing.T) {
	g := newGroup(t)
	badFixture := filepath.Join(g.dir, "fixture-repo-bad-rule")
	if err := makeFixtureRepo(badFixture, true); err != nil {
		t.Fatal(err)
	}
	home := g.newHome()
	g.runSetup(badFixture, home, "global", "")
	g.assertRCNonzero("a rule body carrying a delimiter aborts the run", g.rc)
	g.assertContains("the abort names the offending rule", g.log, "bad-delimiter.mdc")
	for _, leftover := range []string{
		".claude/skills", ".zcode/skills",
		".claude/commands", ".zcode/commands", ".claude/rules", ".zcode/rules",
		".claude/CLAUDE.md", ".zcode/AGENTS.md",
	} {
		p := filepath.Join(home, leftover)
		g.assertAbsent("nothing installed: "+strings.TrimPrefix(p, home+"/"), p)
	}
}
