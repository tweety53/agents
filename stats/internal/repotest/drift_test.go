package repotest

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
	"github.com/tweety53/agents/stats/internal/stages"
)

// The drift checks: each pins one index or table against the files or code it names, so a file,
// role, guard or stage key added or removed on one side cannot leave the other stale.

// readTree returns every file under root's dirs whose name ends in one of exts, keyed by its
// repository-relative path; a missing dir holds none.
func readTree(t *testing.T, root string, dirs []string, exts ...string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, d := range dirs {
		err := filepath.WalkDir(filepath.Join(root, d), func(p string, e fs.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}
			if e.IsDir() {
				if n := e.Name(); n == "node_modules" || n == "dist" || n == "testdata" {
					return filepath.SkipDir
				}
				return nil
			}
			if !slices.ContainsFunc(exts, func(x string) bool { return strings.HasSuffix(p, x) }) {
				return nil
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, p)
			out[rel] = string(b)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return out
}

// TestEveryContractIsIndexed: every skills/flow-contracts/*.md is named in that directory's
// SKILL.md index, so no contract is reachable only by someone who already knows its name.
func TestEveryContractIsIndexed(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root := repoRoot(t)
	dir := filepath.Join(root, "skills/flow-contracts")
	index, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	for _, f := range files {
		if n := filepath.Base(f); n != "SKILL.md" && !strings.Contains(string(index), n) {
			t.Errorf("skills/flow-contracts/%s is not named in skills/flow-contracts/SKILL.md", n)
		}
	}
}

// slashCommand is a `/flow…` slash command in prose: preceded by nothing a path segment carries.
var slashCommand = regexp.MustCompile(`(?:^|[^A-Za-z0-9_./~-])/(flow(?:-[a-z]+)*)`)

// TestNoStaleSlashCommand: every `/flow…` command the instruction files name is a skill of
// skills/ with a SKILL.md, so a retired command (`/flow-status`, `/flow-fast`) is never offered.
func TestNoStaleSlashCommand(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root := repoRoot(t)
	seen := 0
	for rel, body := range readTree(t, root, []string{"skills", "rules", ".flow", "mods", "CLAUDE.md", "README.md"}, ".md", ".mdc", ".tsx", ".ts") {
		for _, m := range slashCommand.FindAllStringSubmatch(body, -1) {
			seen++
			if m[1] == "flow-contracts" {
				t.Errorf("%s names /flow-contracts as a command; it is a contracts directory", rel)
				continue
			}
			if _, err := os.Stat(filepath.Join(root, "skills", m[1], "SKILL.md")); err != nil {
				t.Errorf("%s names /%s, no skill of skills/", rel, m[1])
			}
		}
	}
	if seen == 0 {
		t.Fatal("no /flow command found: the scan is broken")
	}
}

// derivedRoles are the roles no skill or guard names as a literal: lpRole (stats/internal/guard/loop.go)
// derives each as "loop-" + its loop key's suffix family.
var derivedRoles = []string{"loop-fix", "loop-plan", "loop-review-1", "loop-review-r"}

// TestEveryPolicyRoleHasACaller: every role of the **Allowed pairs** table
// (skills/flow-contracts/model-policy.md) is dispatched somewhere — a skill's `Role: <id>` line or
// `pair-judge <id>` call, or the id as a string literal in a guard — so a role nothing dispatches
// is retired rather than left in the policy.
func TestEveryPolicyRoleHasACaller(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root := repoRoot(t)
	policy, err := os.ReadFile(filepath.Join(root, "skills/flow-contracts/model-policy.md"))
	if err != nil {
		t.Fatal(err)
	}
	var corpus strings.Builder
	for rel, body := range readTree(t, root, []string{"skills", "rules"}, ".md", ".mdc") {
		if rel != "skills/flow-contracts/model-policy.md" {
			corpus.WriteString(body)
		}
	}
	text := corpus.String()
	var code strings.Builder
	for rel, body := range readTree(t, root, []string{"stats/internal/guard"}, ".go") {
		if !strings.HasSuffix(rel, "_test.go") {
			code.WriteString(body)
		}
	}
	rows := roleRow.FindAllStringSubmatch(string(policy), -1)
	if len(rows) == 0 {
		t.Fatal("model-policy.md carries no Allowed pairs role row")
	}
	for _, m := range rows {
		id := m[1]
		called := regexp.MustCompile(`(?:Role: |pair-judge )`+regexp.QuoteMeta(id)+`(?:[^a-z0-9-]|$)`).MatchString(text) ||
			strings.Contains(code.String(), `"`+id+`"`) || slices.Contains(derivedRoles, id)
		if !called {
			t.Errorf("role %s has no caller: no skill names `Role: %s` or `pair-judge %s`, and no guard names it", id, id, id)
		}
	}
}

// registryKey is a guard's registration: `Registry["<name>"] = `.
var registryKey = regexp.MustCompile(`Registry\["([a-z0-9-]+)"\] = `)

// operatorGuards are registered guards only the operator runs, by hand.
var operatorGuards = map[string]string{
	"recover-guard-incident": "the operator's recovery from a stuck guard revert; no flow step runs it",
}

// TestEveryRegisteredGuardIsInvoked: every guard.Registry key is run from somewhere — a
// `flow-guard <name>` call in a skill, rule, hook, mod, script or project file, or its name as a
// string literal in another Go file — so a guard nothing calls is retired rather than kept.
func TestEveryRegisteredGuardIsInvoked(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root := repoRoot(t)
	var text strings.Builder
	for _, body := range readTree(t, root, []string{"skills", "rules", "hooks", "mods", "scripts", ".flow", "CLAUDE.md", "setup.sh"}, ".md", ".mdc", ".tsx", ".ts", ".sh", ".json", "flow-guard") {
		text.WriteString(body)
	}
	goFiles := readTree(t, root, []string{"stats/internal", "stats/cmd"}, ".go")
	registered := map[string]string{}
	for rel, body := range goFiles {
		if strings.HasPrefix(rel, "stats/internal/guard/") && !strings.HasSuffix(rel, "_test.go") {
			for _, m := range registryKey.FindAllStringSubmatch(body, -1) {
				registered[m[1]] = rel
			}
		}
	}
	if len(registered) == 0 {
		t.Fatal("no guard.Registry key found: the scan is broken")
	}
	for name, home := range registered {
		if _, ok := operatorGuards[name]; ok {
			continue
		}
		if regexp.MustCompile(`flow-guard ` + regexp.QuoteMeta(name) + `(?:[^a-z0-9-]|$)`).MatchString(text.String()) {
			continue
		}
		lit, reg := `"`+name+`"`, `Registry["`+name+`"] = `
		called := false
		for rel, body := range goFiles {
			if strings.HasSuffix(rel, "_test.go") || rel == "stats/internal/records/types.go" {
				continue
			}
			if n := strings.Count(body, lit); n > strings.Count(body, reg) {
				called = true
				break
			}
		}
		if !called {
			t.Errorf("guard %s (%s) is registered and never invoked", name, home)
		}
	}
}

// stageKeysTable is the **Stage keys** table of skills/flow-contracts/pipeline.md: a row's mark,
// then its last cell's retired keys.
var (
	stageKeysRow  = regexp.MustCompile("(?m)^\\| `(flow\\.[a-z-]+)` \\|.*\\| ([^|]+) \\|$")
	backtickedKey = regexp.MustCompile("`(flow\\.[a-z-]+)`")
	phaseKeys     = regexp.MustCompile(`(?s)const PHASE_KEYS[^{]*\{(.*?)\n\}`)
	phaseEntry    = regexp.MustCompile(`(?s)'(flow-[a-z]+)': \[(.*?)\]`)
	quoted        = regexp.MustCompile(`'([a-z-]+)'`)
)

// markPhase is the flow-task-list phase each mark's keys show under.
var markPhase = map[string]string{
	"flow.plan": "flow-plan", "flow.implement": "flow-implement", "flow.review": "flow-implement",
	"flow.verify": "flow-implement", "flow.land": "flow-integrate", "flow.self-review": "flow-integrate",
	"flow.cleanup": "flow-integrate",
}

// TestStageKeysTableMatchesTheCode: the **Stage keys** table names exactly stages.Retired, each
// retired key under the mark it folds into; every mark is a stages.Table key; and the flow-task-list
// mod's PHASE_KEYS shows each mark and retired key under its mark's phase — save flow.decide, which
// marks both the brainstorm and a fix run and so keeps whichever phase is current.
func TestStageKeysTableMatchesTheCode(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	root := repoRoot(t)
	pipeline, err := os.ReadFile(filepath.Join(root, "skills/flow-contracts/pipeline.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, section, _ := strings.Cut(string(pipeline), "### Stage keys\n")
	section, _, _ = strings.Cut(section, "\n### ")
	rows := stageKeysRow.FindAllStringSubmatch(section, -1)
	if len(rows) == 0 {
		t.Fatal("pipeline.md's Stage keys table has no row")
	}
	table := map[string]string{}
	keyOf := map[string]string{}
	for _, m := range rows {
		mark := m[1]
		if _, ok := stages.Name(mark); !ok {
			t.Errorf("Stage keys mark %s is no stages.Table key", mark)
		}
		keyOf[mark] = mark
		for _, k := range backtickedKey.FindAllStringSubmatch(m[2], -1) {
			table[k[1]] = mark
			keyOf[k[1]] = mark
		}
	}
	for k, mark := range stages.Retired {
		if table[k] != mark {
			t.Errorf("stages.Retired folds %s into %s; the Stage keys table has %q", k, mark, table[k])
		}
	}
	for k, mark := range table {
		if stages.Retired[k] != mark {
			t.Errorf("the Stage keys table folds %s into %s; stages.Retired has %q", k, mark, stages.Retired[k])
		}
	}
	mod, err := os.ReadFile(filepath.Join(root, "mods/flow-task-list/hooks/register.tsx"))
	if err != nil {
		t.Fatal(err)
	}
	block := phaseKeys.FindStringSubmatch(string(mod))
	if block == nil {
		t.Fatal("register.tsx carries no PHASE_KEYS")
	}
	phaseOf := map[string]string{}
	for _, e := range phaseEntry.FindAllStringSubmatch(block[1], -1) {
		for _, q := range quoted.FindAllStringSubmatch(e[2], -1) {
			phaseOf["flow."+q[1]] = e[1]
		}
	}
	for k, mark := range keyOf {
		want := markPhase[mark]
		if k == "flow.decide" {
			want = ""
		}
		if phaseOf[k] != want {
			t.Errorf("PHASE_KEYS shows %s under %q, want %q (its mark %s)", k, phaseOf[k], want, mark)
		}
	}
	for k := range phaseOf {
		if _, ok := keyOf[k]; !ok {
			t.Errorf("PHASE_KEYS names %s, no mark or retired key of the Stage keys table", k)
		}
	}
}
