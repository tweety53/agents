package guard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// TestPlanShapeRulesFlagPrintsEveryRule pins decision rules-flag: a lone
// -rules prints one `<id>  <sentence>` line per psRules row, in order, and
// exits 0; any other argument list keeps the usage exit.
func TestPlanShapeRulesFlagPrintsEveryRule(t *testing.T) {
	t.Parallel()
	r := runGuard("check-plan-shape", []string{"-rules"}, Env{Getenv: os.Getenv})
	if r.rc != 0 || r.err != "" {
		t.Fatalf("rc=%d err=%q; want 0 and empty stderr", r.rc, r.err)
	}
	lines := strings.Split(strings.TrimSuffix(r.out, "\n"), "\n")
	if len(lines) != len(psRules) {
		t.Fatalf("%d lines, want %d: %q", len(lines), len(psRules), r.out)
	}
	var ids []string
	for i, row := range psRules {
		if want := row.id + "  " + row.sentence; lines[i] != want {
			t.Errorf("line %d = %q, want %q", i+1, lines[i], want)
		}
		ids = append(ids, row.id)
	}
	for _, id := range []string{"F3a", "F3b", "plan-step-files"} {
		if !slices.Contains(ids, id) {
			t.Errorf("no %s row", id)
		}
	}
	if slices.Contains(ids, "F21") {
		t.Error("retired F21 has a row")
	}
	u := runGuard("check-plan-shape", []string{"-rules", "x"}, Env{Getenv: os.Getenv})
	if u.rc != 2 || !strings.Contains(u.err, "usage: flow-guard check-plan-shape") {
		t.Errorf("-rules x: rc=%d err=%q; want 2 and the usage line", u.rc, u.err)
	}
}

// TestPlanShapeRulesMatchTheirChecks pins the table to the code: every id an
// at/atFile/on call names, and every sub-guard, has exactly one psRules row,
// and every row is used.
func TestPlanShapeRulesMatchTheirChecks(t *testing.T) {
	t.Parallel()
	f, err := parser.ParseFile(token.NewFileSet(), "planshape.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	used := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := c.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		arg := map[string]int{"at": 1, "atFile": 2, "on": 0}
		i, ok := arg[sel.Sel.Name]
		if !ok || len(c.Args) <= i {
			return true
		}
		if lit, ok := c.Args[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			s, _ := strconv.Unquote(lit.Value)
			used[s] = true
		}
		return true
	})
	for _, g := range psSubGuards {
		used[g] = true
	}
	rows := map[string]bool{}
	for _, r := range psRules {
		if rows[r.id] {
			t.Errorf("psRules holds %s twice", r.id)
		}
		rows[r.id] = true
		if !used[r.id] {
			t.Errorf("psRules row %s is used by no check", r.id)
		}
	}
	for id := range used {
		if !rows[id] {
			t.Errorf("check %s has no psRules row", id)
		}
	}
}

// TestPlanShapeLinesNameTheirRule pins decision line-id-position: every
// violation line reads `path:line: <id> <message>`, the id a -rules id, and
// the run records each id as a reason.
func TestPlanShapeLinesNameTheirRule(t *testing.T) {
	t.Parallel()
	plan := "- [ ] 1. A\n\n**Tests:** prose only\n\n- [ ] 2. B\n\n" +
		strings.Split(psfRefused, "\n")[2] + "\n\n**Files:** `a/x.go`\n**Tests:** `TestB`\n"
	root := psfWrite(t, map[string]string{"spectre/changes/c/tasks.md": plan, "a/x.go": "", "a/keep.go": ""})
	r := &Reasons{}
	res := runGuard("check-plan-shape", []string{filepath.Join(root, "spectre/changes/c/tasks.md")}, Env{Getenv: os.Getenv, Reasons: r})
	if res.rc != 1 {
		t.Fatalf("rc=%d out=%q err=%q; want 1", res.rc, res.out, res.err)
	}
	lines := strings.Split(strings.TrimSuffix(res.out, "\n"), "\n")
	want := [][2]string{{"1", "F2"}, {"3", "F6"}, {"7", "plan-step-files"}}
	if len(lines) != len(want) {
		t.Fatalf("lines %q, want %d", lines, len(want))
	}
	ids := map[string]bool{}
	for _, row := range psRules {
		ids[row.id] = true
	}
	re := regexp.MustCompile(`^\S+:(\d+): (\S+) `)
	for i, l := range lines {
		m := re.FindStringSubmatch(l)
		if m == nil || m[1] != want[i][0] || m[2] != want[i][1] || !ids[m[2]] {
			t.Errorf("line %q, want line %s id %s", l, want[i][0], want[i][1])
		}
	}
	if got := r.IDs(); !slices.Equal(got, []string{"F2", "F6", "plan-step-files"}) {
		t.Errorf("recorded %q, want [F2 F6 plan-step-files]", got)
	}
	clean := &Reasons{}
	md := psFixture(t, "- [ ] 1. A clean task\n\n**Files:** `scripts/foo.sh`\n**Tests:** `scripts/test-foo.sh`\n**Commit:** `feat(scripts): add foo`\n")
	if c := runGuard("check-plan-shape", []string{md}, Env{Getenv: os.Getenv, Reasons: clean}); c.rc != 0 || len(clean.IDs()) != 0 {
		t.Errorf("clean plan: rc=%d recorded %q; want 0 and none", c.rc, clean.IDs())
	}
}
