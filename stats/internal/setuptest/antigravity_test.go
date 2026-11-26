package setuptest

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAntigravityInstall(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	g.runSetup(repoRoot, home, "antigravity", "")
	g.assertRCZero("antigravity succeeds", g.rc, g.log)

	pluginDir := filepath.Join(home, ".gemini/config/plugins/flow")
	g.assertFileExists("plugin.json is written", filepath.Join(pluginDir, "plugin.json"))
	g.assertSymlink("hooks.json is symlinked", filepath.Join(pluginDir, "hooks.json"))

	// Validate hooks.json is well-formed JSON.
	rawHooks, err := os.ReadFile(filepath.Join(pluginDir, "hooks.json"))
	if err != nil {
		t.Fatalf("read hooks.json: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(rawHooks, &parsed); err != nil {
		t.Fatalf("hooks.json invalid JSON: %v", err)
	}
	for _, name := range []string{"block-unsafe-commands", "enforce-agent-baseline", "enforce-loop-dispatch", "flow-active-change", "flow-session-activity"} {
		if _, ok := parsed[name]; !ok {
			t.Errorf("hooks.json missing hook %q", name)
		}
	}

	// Skills: every repo skill is symlinked, zero dangling symlinks under $HOME.
	expectedSkills := len(repoDirs(t, "skills"))
	skillsDir := filepath.Join(pluginDir, "skills")
	installed := 0
	for _, name := range visibleEntries(skillsDir) {
		installed++
		entry := filepath.Join(skillsDir, name)
		if fi, err := os.Stat(entry); !isSymlink(entry) || err != nil || !fi.IsDir() {
			t.Errorf("skill %q is not a resolvable symlinked dir", name)
		}
	}
	g.assertEq("every repo skill is installed in plugin", expectedSkills, installed)

	var dangling []string
	_ = filepath.WalkDir(home, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.Type()&fs.ModeSymlink != 0 {
			if _, statErr := os.Stat(p); statErr != nil {
				dangling = append(dangling, p)
			}
		}
		return nil
	})
	g.assertEq("antigravity leaves zero dangling symlinks", "", strings.Join(dangling, "\n"))

	// Rules: managed block in rules/AGENTS.md with Antigravity pointers and harness mapping;
	// full rule texts + rendered agent-baseline.md in rules-full/; opt-in Kotlin rule excluded.
	agentsMD := filepath.Join(pluginDir, "rules/AGENTS.md")
	g.assertFileExists("rules/AGENTS.md exists", agentsMD)
	g.assertContains("rules/AGENTS.md points into rules-full", agentsMD, "Full rule: `~/.gemini/config/plugins/flow/rules-full/")
	g.assertContains("rules/AGENTS.md includes Antigravity harness mapping", agentsMD, "## Antigravity Harness Mapping")
	g.assertNotContains("rules/AGENTS.md excludes opt-in Kotlin rule", agentsMD, "<!-- rule: "+kotlinRule+" -->")

	baseline := filepath.Join(pluginDir, "rules-full/agent-baseline.md")
	g.assertFileExists("rendered agent-baseline.md exists", baseline)
	g.assertContains("agent-baseline.md rewrites rule paths to Antigravity plugin", baseline, "~/.gemini/config/plugins/flow/rules-full/")
	g.assertNotContains("agent-baseline.md drops ~/.claude/rules/ references", baseline, "~/.claude/rules/")

	// Hooks and flow-guard shim.
	g.assertSymlink("harness_io.py is linked beside hooks", filepath.Join(pluginDir, "hooks/harness_io.py"))
	g.assertSymlink("block-unsafe-commands.py is linked", filepath.Join(pluginDir, "hooks/block-unsafe-commands.py"))
	g.assertExecutable("flow-guard shim is installed", filepath.Join(home, ".local/bin/flow-guard"))

	// Idempotent re-run.
	g.runSetup(repoRoot, home, "antigravity", "")
	g.assertRCZero("second antigravity run succeeds", g.rc, g.log)
}
