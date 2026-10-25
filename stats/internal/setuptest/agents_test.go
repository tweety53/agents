package setuptest

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestSetupAgents asserts setup.sh installs the three flow-<effort> agent definitions on
// `./setup.sh global`, run against a sandboxed HOME so it never touches the real ~/.claude.
// Three symlinks must exist under $HOME/.claude/agents/ and each resolve; each flow-<effort>
// frontmatter carries its own `effort:` line and no `model:` line — the Agent tool's
// dispatch-time `model` parameter overrides a definition's `model`, so the model axis needs no
// definition of its own while the effort axis does. Each `tools:` allowlist must omit `Agent`
// and `Task` — every panel slot dispatches on a flow-<effort> definition, flow-low on a
// `default` panel, so a reviewer structurally cannot fork (KAN-495).
func TestSetupAgents(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(repoRoot, home, "global", "")
	g.assertRCZero("setup.sh global succeeds", g.rc, g.log)

	agentsDir := filepath.Join(home, ".claude", "agents")
	hasModel := regexp.MustCompile(`(?m)^model: `)
	hasTools := regexp.MustCompile(`(?m)^tools: `)
	toolsFork := regexp.MustCompile(`(?m)^tools: .*(Agent|Task)`)
	for _, effort := range []string{"low", "medium", "high"} {
		t.Run("flow-"+effort, func(t *testing.T) {
			link := filepath.Join(agentsDir, "flow-"+effort+".md")
			if !isSymlink(link) {
				t.Fatalf("%s is not a symlink", link)
			}
			b, err := os.ReadFile(link)
			if err != nil {
				t.Fatalf("%s does not resolve: %v", link, err)
			}
			if !regexp.MustCompile(`(?m)^effort: ` + effort + `$`).Match(b) {
				t.Errorf("%s frontmatter missing 'effort: %s'", link, effort)
			}
			if hasModel.Match(b) {
				t.Errorf("%s frontmatter still carries a 'model:' line", link)
			}
			if !hasTools.Match(b) {
				t.Errorf("%s frontmatter missing a 'tools:' allowlist", link)
			} else if toolsFork.Match(b) {
				t.Errorf("%s 'tools:' allowlist grants Agent or Task — a dispatched role must not fork", link)
			}
		})
	}

	found, _ := filepath.Glob(filepath.Join(agentsDir, "flow-*.md"))
	g.assertEq("six flow-* symlinks present under "+agentsDir, 6, len(found))
	// The flow-planner-<effort> definitions alone carry Agent: the self-review
	// and in-run fix planners dispatch Explore and Plan agents (kan-1033).
	for _, e := range []string{"low", "medium", "high"} {
		b, err := os.ReadFile(filepath.Join(agentsDir, "flow-planner-"+e+".md"))
		if err != nil || !toolsFork.Match(b) {
			t.Errorf("flow-planner-%s.md missing or its 'tools:' allowlist lacks Agent (err %v)", e, err)
		}
	}
}
