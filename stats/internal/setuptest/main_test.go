// Package setuptest is the regression harness for setup.sh.
//
// Every case runs against a throwaway HOME created under /tmp, and the real ~/.claude and
// ~/.zcode are fingerprinted before the first case and again after the last one — a mismatch
// fails the run. Nothing here writes outside the sandbox. Each assertion group is one
// top-level test marked parallel, with its own sub-sandbox; scripts/test-setup.sh runs the
// package.
//
// WHY THIS EXISTS. setup.sh writes into the user's home directory, and two data-loss defects
// reached review:
//
//  1. Delimiters in reversed order (end above begin) counted 1 begin and 1 end, took the
//     rewrite branch, and silently deleted every line after the begin marker.
//  2. An unbalanced pair took an append branch that could never converge — 1/0 became 2/1
//     became 3/2, each run adding another copy of the block to the file injected into every
//     session.
//
// Both are fixed. Nothing in the repo would have caught either coming back, and nothing would
// catch them coming back again. That is what this package is for: the cases are the shapes
// that actually broke, plus the containment guarantees whose failure is silent (an opt-in rule
// installed globally, a rule installed that never declared itself always-on, a stale skill
// copy left shadowing its own symlink).
//
// WHAT THIS CAN AND CANNOT PROVE — read before trusting a green run.
//
// It proves things about the installer's OBSERVABLE FILESYSTEM EFFECTS: what exists after a
// run, what it points at, what it contains, its mode, and the exit status. That is the whole
// of its scope.
//
// It does NOT prove the installed content is correct or current — that a rule says the right
// thing, that a skill's prose is coherent, or that an agent reading the managed block behaves
// as intended. Agent behaviour is a human job.
//
// It does not cover any harness the installer does not write, nor the per-project modes
// beyond what the containment cases need.
//
// A green run means "the shapes listed here still behave as listed". Any shape nobody thought
// to list passes clean.
//
// Two source trees are used, deliberately:
//
//   - The REAL repo, for the containment guarantees, which are claims about the rules and
//     skills actually shipped here (the opt-in Kotlin rule, the real skill set).
//   - A generated FIXTURE repo — a copy of setup.sh beside synthetic skills/rules/commands —
//     for the cases that need a malformed input the real tree must never contain (a rule body
//     carrying a block delimiter, an unterminated frontmatter). setup.sh never writes into the
//     tree it installs from, so the groups that do not mutate a fixture share one base fixture,
//     read-only; a group that mutates one builds its own.
//
// Set KEEP_SANDBOX=1 to leave the sandbox behind for inspection; it is removed otherwise, on
// success, on failure and on SIGINT/SIGTERM alike. Every installer run is bounded by the test's
// deadline, so a -timeout fails the groups before it kills the binary; only a crashed binary (a
// panic) leaves the sandbox behind.
package setuptest

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
)

// kotlinRule is the opt-in rule two groups in different files assert on.
const kotlinRule = "kotlin-backend-development-standard.mdc"

var (
	repoRoot string // this checkout, the parent of stats/
	realHome string // $HOME as the test binary started
	sandbox  string // /tmp/flow-test-setup.*; every byte the harness writes lies under it
	begin    string // CLAUDE_MD_BEGIN, read out of setup.sh
	end      string // CLAUDE_MD_END, read out of setup.sh
	fixture  string // the shared base fixture repo — read-only for every group using it

	runCtx, stopRuns = context.WithCancel(context.Background()) // cancelled on an interrupt
	running          sync.WaitGroup                             // the installer runs in flight
)

func TestMain(m *testing.M) { os.Exit(runMain(m)) }

func runMain(m *testing.M) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", a...)
		return 2
	}
	var err error
	if repoRoot, err = filepath.Abs("../../.."); err != nil {
		return die("cannot resolve the repo root: %v", err)
	}
	setupSh := filepath.Join(repoRoot, "setup.sh")
	if fi, err := os.Stat(setupSh); err != nil || fi.Mode().Perm()&0o111 == 0 {
		return die("setup.sh not found or not executable at %s", setupSh)
	}
	// The delimiters are read out of setup.sh rather than restated here. A second copy would
	// drift from the first, and every case depends on matching the installer's literals
	// exactly — a stale copy would make the whole delimiter group pass vacuously.
	begin, end = readLiteral(setupSh, "CLAUDE_MD_BEGIN"), readLiteral(setupSh, "CLAUDE_MD_END")
	if begin == "" || end == "" || begin == end {
		return die("could not read the block delimiters out of setup.sh")
	}
	if realHome = os.Getenv("HOME"); realHome == "" {
		return die("HOME must be set")
	}
	// The sandbox is created under /tmp explicitly, not under $TMPDIR: on macOS $TMPDIR is a
	// per-user path under /var/folders, and the safety claim this harness makes is the
	// concrete one — every byte it writes is under /tmp and is removed again.
	if sandbox, err = os.MkdirTemp("/tmp", "flow-test-setup."); err != nil {
		return die("cannot create the sandbox under /tmp: %v", err)
	}
	defer cleanup()
	// An interrupt kills every in-flight installer's process group, waits for them, and removes
	// the sandbox, as the bash harness's EXIT trap did.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		stopRuns()
		running.Wait()
		cleanup()
		os.Exit(130)
	}()

	homeBefore := realHomeFingerprint(realHome)
	srcBefore := sourceTreeFingerprint(repoRoot)
	fixture = filepath.Join(sandbox, "fixture-repo")
	if err := makeFixtureRepo(fixture, false); err != nil {
		return die("%v", err)
	}
	fixtureBefore := treeFingerprint(fixture)

	fmt.Printf("setup.sh regression harness\n  repo    : %s\n  sandbox : %s\n", repoRoot, sandbox)
	code := m.Run()
	return exitCode(code, closeOut(os.Stdout, []fpCheck{
		{"~/.claude and ~/.zcode are unchanged", homeBefore, func() string { return realHomeFingerprint(realHome) }},
		{"the repo's own skills, rules and commands are unchanged", srcBefore, func() string { return sourceTreeFingerprint(repoRoot) }},
		{"leak detection: the shared fixture repo is unchanged after every group", fixtureBefore, func() string { return treeFingerprint(fixture) }},
	}))
}

// fpCheck is one close-out comparison: a fingerprint taken before the tests, and how to retake it.
type fpCheck struct {
	desc, before string
	now          func() string
}

// closeOut is the safety group: every test ran with a sandboxed HOME; this proves it. It
// returns 1 when any fingerprint moved.
func closeOut(w io.Writer, checks []fpCheck) int {
	fmt.Fprint(w, "\n== Nothing outside the sandbox was touched ==\n")
	code := 0
	for _, c := range checks {
		if got := c.now(); got == c.before {
			fmt.Fprintf(w, "  ✓ %s\n", c.desc)
		} else {
			code = 1
			fmt.Fprintf(w, "  ✗ %s\n      expected [%s], got [%s]\n", c.desc, c.before, got)
		}
	}
	return code
}

// exitCode fails a green test run whose close-out failed.
func exitCode(testCode, closeCode int) int {
	if closeCode != 0 && testCode == 0 {
		return 1
	}
	return testCode
}

// sandboxRemovable is the belt and braces: only ever remove a path this harness created under /tmp.
func sandboxRemovable(p string) bool { return strings.HasPrefix(p, "/tmp/flow-test-setup.") }

func cleanup() {
	if os.Getenv("KEEP_SANDBOX") != "" {
		fmt.Fprintf(os.Stderr, "KEEP_SANDBOX set — sandbox left at %s\n", sandbox)
		return
	}
	if !sandboxRemovable(sandbox) {
		return
	}
	if err := os.RemoveAll(sandbox); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot remove the sandbox %s: %v\n", sandbox, err)
	}
}

// readLiteral returns NAME's single-quoted value from the first `NAME='…'` line of setup.sh.
func readLiteral(setupSh, name string) string {
	f, err := os.Open(setupSh)
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, name+"=") {
			continue
		}
		if i := strings.Index(line, "='"); i >= 0 {
			line = line[i+2:]
		}
		return strings.TrimSuffix(line, "'")
	}
	return ""
}

// makeFixtureRepo builds a minimal agents repo the harness fully controls. The real tree
// cannot host a rule whose body carries a block delimiter or an unterminated frontmatter, and
// those are exactly the inputs two of the guarantees are about. It returns an error rather
// than failing a test: TestMain calls it before any *testing.T exists.
func makeFixtureRepo(d string, badRule bool) error {
	for _, sub := range []string{"rules", "commands-claude", "skills/demo-skill", "hooks"} {
		if err := os.MkdirAll(filepath.Join(d, sub), 0o755); err != nil {
			return err
		}
	}
	src, err := os.ReadFile(filepath.Join(repoRoot, "setup.sh"))
	if err != nil {
		return fmt.Errorf("cannot copy setup.sh into the fixture repo: %w", err)
	}
	// setup.sh resolves {{lint-commands}} through scripts/project-get.sh, which sources
	// scripts/lib/project-section.sh, which sources scripts/lib/strip-bom.sh — all
	// resolved relative to the fixture's own layout, so the fixture carries the same
	// files verbatim that the real tree does.
	for _, c := range []struct {
		rel  string
		mode os.FileMode
	}{
		{"scripts/project-get.sh", 0o755},
		{"scripts/lib/project-section.sh", 0o644},
		{"scripts/lib/strip-bom.sh", 0o644},
	} {
		b, err := os.ReadFile(filepath.Join(repoRoot, c.rel))
		if err != nil {
			return fmt.Errorf("cannot copy %s into the fixture repo: %w", c.rel, err)
		}
		p := filepath.Join(d, c.rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, b, c.mode); err != nil {
			return err
		}
		if err := os.Chmod(p, c.mode); err != nil {
			return err
		}
	}
	files := []struct {
		path, body string
		mode       os.FileMode
	}{
		{"setup.sh", string(src), 0o755},
		{"skills/demo-skill/SKILL.md", "# demo skill\n", 0o644},
		{"CLAUDE.md", "# fixture CLAUDE.md\n", 0o644},
		{"AGENTS.md", "# fixture AGENTS.md\n", 0o644},
		{"commands-claude/demo.md", "# demo command\n", 0o644},
		// Always-on: the one rule that must install.
		{"rules/good-always.mdc", "---\ndescription: fixture always-on rule\nalwaysApply: true\n---\n\n# Good\nBODY-GOOD-ALWAYS\n", 0o644},
		// Opt-in: declares false, must never install.
		{"rules/opt-in-false.mdc", "---\ndescription: fixture opt-in rule\nalwaysApply: false\n---\n\n# OptIn\nBODY-OPT-IN\n", 0o644},
		// Opt-in rule carrying the lint-commands placeholder: the project render must replace
		// it with the commands declared in the project's own ## lint section.
		{"rules/lint-placeholder.mdc", "---\ndescription: fixture rule carrying the lint placeholder\nalwaysApply: false\n---\n\n# LintPlaceholder\nBODY-LINT-PLACEHOLDER\n\n```bash\n{{lint-commands}}\n```\n", 0o644},
		// Frontmatter that never closes: the `alwaysApply: true` below is prose, not a declaration.
		{"rules/unterminated.mdc", "---\ndescription: fixture unterminated frontmatter\nalwaysApply: true\n\n# Unterminated\nBODY-UNTERMINATED\n", 0o644},
		// A value that merely starts with `true` must not be read as `true`.
		{"rules/tricky-prefix.mdc", "---\ndescription: fixture tricky value\nalwaysApply: true_for_kotlin_only\n---\n\n# Tricky\nBODY-TRICKY-PREFIX\n", 0o644},
		// An always-on rule carrying a CORE excerpt. Globally the block gets the core plus a
		// pointer to the installed full text; a project's own block gets the whole body. Two
		// sentinels, so those renders can be told apart.
		{"rules/cored.mdc", "---\ndescription: fixture rule with a core excerpt\nalwaysApply: true\n---\n\n# Cored\n<!-- core -->\nBODY-CORE-EXCERPT\n<!-- /core -->\n\n## Detail\nBODY-AFTER-CORE\n", 0o644},
		// Not a rule, and carries no frontmatter: the baseline a dispatched subagent is told to
		// read. install_rules_claude links it beside the rules and warns when it is missing,
		// so the fixture needs one for an otherwise-clean install to stay clean.
		{"rules/agent-baseline.md", "# fixture agent baseline\n", 0o644},
		// The hook that enforces the baseline in dispatches. Named as the real one is, because
		// install_hooks reports on registration by looking for that name in settings.json.
		{"hooks/enforce-agent-baseline.py", "#!/usr/bin/env python3\n\"\"\"fixture hook\"\"\"\n", 0o644},
		// The UserPromptSubmit hook that names the session's active flow change. Named as the
		// real one is, because install_hooks/install_hooks_zcode report on its registration by
		// looking for that name in settings.json / config.json.
		{"hooks/flow-active-change.py", "#!/usr/bin/env python3\n\"\"\"fixture hook\"\"\"\n", 0o644},
		// The main-checkout protection hook. Named as the real one is, because install_hooks
		// reports on its registration by looking for that name in settings.json — the pins in
		// the "Core excerpts" group assert that warning, so the fixture has to install the hook
		// the warning names for their precondition to be real.
		{"hooks/protect-main-checkout.py", "#!/usr/bin/env python3\n\"\"\"fixture hook\"\"\"\n", 0o644},
	}
	if badRule {
		// An always-on rule whose BODY carries a bare begin delimiter. Inlining it would put a
		// second delimiter inside the managed block and corrupt every later run.
		files = append(files, struct {
			path, body string
			mode       os.FileMode
		}{"rules/bad-delimiter.mdc", "---\ndescription: fixture rule carrying a delimiter\nalwaysApply: true\n---\n\n# Bad\n" + begin + "\nBODY-BAD-DELIMITER\n", 0o644})
	}
	for _, f := range files {
		p := filepath.Join(d, f.path)
		if err := os.WriteFile(p, []byte(f.body), f.mode); err != nil {
			return err
		}
		if err := os.Chmod(p, f.mode); err != nil {
			return err
		}
	}
	return nil
}
