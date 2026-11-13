package smoke

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type fakeTB struct {
	testing.TB
	skipped bool
}

func (f *fakeTB) Helper()          {}
func (f *fakeTB) Skip(args ...any) { f.skipped = true }

func TestHeavySkipsOnlyAnUnkeptShortRun(t *testing.T) {
	oldShort, oldRun, oldGetenv := shortRun, runFlag, getenv
	t.Cleanup(func() { shortRun, runFlag, getenv = oldShort, oldRun, oldGetenv })

	cases := []struct {
		short           bool
		run, full, keep string
		skips           bool
	}{
		{true, "", "", "", true},
		{false, "", "", "", false},
		{true, "^TestX$", "", "", false},
		{true, "", "1", "", false},
		{true, "", "", "stats/internal/smoke/smoke_test.go", false},
		{true, "", "", "a/b stats/internal/smoke/smoke_test.go", false},
		{true, "", "", "stats/internal/smoke/", false},
		{true, "", "", "stats/internal/smoke", false},
		{true, "", "", "stats/", false},
		{true, "", "", "stats/internal/smoke/smoke.go", true},
		{true, "", "", "stats/internal/smok", true},
		{true, "", "", "internal/smoke/smoke_test.go", true},
	}
	for _, c := range cases {
		shortRun = func() bool { return c.short }
		runFlag = func() string { return c.run }
		env := map[string]string{"FLOW_VERIFY_FULL": c.full, "FLOW_SMOKE_KEEP": c.keep}
		getenv = func(k string) string { return env[k] }
		f := &fakeTB{}
		Heavy(f)
		if f.skipped != c.skips {
			t.Errorf("short=%v run=%q full=%q keep=%q: skipped=%v, want %v", c.short, c.run, c.full, c.keep, f.skipped, c.skips)
		}
	}
}

// checkMarks walks root, skipping node_modules, and returns one finding per
// heavy mark or core tag that contradicts the rules of this package.
func checkMarks(root string) ([]string, error) {
	var findings []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		findings = append(findings, checkFile(fset, file, rel)...)
		return nil
	})
	return findings, err
}

func checkFile(fset *token.FileSet, file *ast.File, rel string) []string {
	var findings []string
	at := func(pos token.Pos, what string) {
		findings = append(findings, rel+":"+strconv.Itoa(fset.Position(pos).Line)+": "+what)
	}
	inSmoke := strings.HasPrefix(rel, "internal/smoke/")
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if ok && isCall(call, "testing", "Short") && !inSmoke {
			at(call.Pos(), "testing.Short called outside internal/smoke")
		}
		return true
	})
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		top := isTopLevelTest(fn)
		tagged := hasCoreTag(fn.Doc)
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isCall(call, "smoke", "Heavy") {
				return true
			}
			first := top && len(fn.Body.List) > 0 && isStmtCall(fn.Body.List[0], call)
			switch {
			case !first:
				at(call.Pos(), "smoke.Heavy is not the first statement of a top-level test")
			case tagged:
				at(call.Pos(), "a test tagged core calls smoke.Heavy")
			}
			return true
		})
	}
	return findings
}

func isCall(call *ast.CallExpr, pkg, name string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == pkg
}

func isStmtCall(stmt ast.Stmt, call *ast.CallExpr) bool {
	es, ok := stmt.(*ast.ExprStmt)
	return ok && es.X == call
}

func isTopLevelTest(fn *ast.FuncDecl) bool {
	if fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") {
		return false
	}
	params := fn.Type.Params.List
	if len(params) != 1 || len(params[0].Names) > 1 {
		return false
	}
	star, ok := params[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "T" && isIdent(sel.X, "testing")
}

func isIdent(e ast.Expr, name string) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == name
}

func hasCoreTag(doc *ast.CommentGroup) bool {
	if doc == nil || len(doc.List) == 0 {
		return false
	}
	return strings.HasPrefix(doc.List[len(doc.List)-1].Text, "// smoke: core — ")
}

func TestMarksCheckerRefusesEachContradiction(t *testing.T) {
	const head = "package p\n\nimport (\n\t\"testing\"\n\n\t\"github.com/tweety53/agents/stats/internal/smoke\"\n)\n\n"
	refused := []struct {
		name, path, body, want string
	}{
		{"tag and call", "a/a_test.go",
			"// smoke: core — x\nfunc TestA(t *testing.T) {\n\tsmoke.Heavy(t)\n}\n",
			"a/a_test.go:11: a test tagged core calls smoke.Heavy"},
		{"second statement", "a/a_test.go",
			"func TestA(t *testing.T) {\n\tt.Parallel()\n\tsmoke.Heavy(t)\n}\n",
			"a/a_test.go:11: smoke.Heavy is not the first statement of a top-level test"},
		{"in t.Run", "a/a_test.go",
			"func TestA(t *testing.T) {\n\tt.Run(\"x\", func(t *testing.T) {\n\t\tsmoke.Heavy(t)\n\t})\n}\n",
			"a/a_test.go:11: smoke.Heavy is not the first statement of a top-level test"},
		{"helper", "a/a_test.go",
			"func helper(t *testing.T) {\n\tsmoke.Heavy(t)\n}\n",
			"a/a_test.go:10: smoke.Heavy is not the first statement of a top-level test"},
		{"benchmark", "a/a_test.go",
			"func BenchmarkA(b *testing.B) {\n\tsmoke.Heavy(b)\n}\n",
			"a/a_test.go:10: smoke.Heavy is not the first statement of a top-level test"},
		{"testing.Short outside", "a/a_test.go",
			"func TestA(t *testing.T) {\n\t_ = testing.Short()\n}\n",
			"a/a_test.go:10: testing.Short called outside internal/smoke"},
	}
	for _, c := range refused {
		t.Run(c.name, func(t *testing.T) {
			got := checkFixture(t, c.path, head+c.body)
			if len(got) != 1 || got[0] != c.want {
				t.Errorf("findings = %q, want [%q]", got, c.want)
			}
		})
	}

	accepted := []struct {
		name, path, body string
	}{
		{"tag without call", "a/a_test.go",
			"// smoke: core — x\nfunc TestA(t *testing.T) {\n\t_ = smoke.Heavy\n}\n"},
		{"first statement before Parallel", "a/a_test.go",
			"func TestA(t *testing.T) {\n\tsmoke.Heavy(t)\n\tt.Parallel()\n}\n"},
		{"tag ends a longer doc", "a/a_test.go",
			"// TestA proves a thing.\n//\n// smoke: core — x\nfunc TestA(t *testing.T) {\n\tt.Parallel()\n}\n"},
		{"tag text on a first line is no tag", "a/a_test.go",
			"// smoke: core — x\n// TestA is heavy.\nfunc TestA(t *testing.T) {\n\tsmoke.Heavy(t)\n}\n"},
		{"testing.Short inside smoke", "internal/smoke/x_test.go",
			"func TestA(t *testing.T) {\n\t_ = testing.Short()\n\t_ = smoke.Heavy\n}\n"},
		{"node_modules", "web/node_modules/m/a_test.go",
			"func helper(t *testing.T) {\n\tsmoke.Heavy(t)\n\t_ = testing.Short()\n}\n"},
	}
	for _, c := range accepted {
		t.Run(c.name, func(t *testing.T) {
			if got := checkFixture(t, c.path, head+c.body); len(got) != 0 {
				t.Errorf("findings = %q, want none", got)
			}
		})
	}
}

func checkFixture(t *testing.T, rel, src string) []string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := checkMarks(root)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// TestTreeMarksAreConsistent fails on every contradiction among the module's
// heavy marks and core tags.
//
// smoke: core — the heavy marks and the core tags agree
func TestTreeMarksAreConsistent(t *testing.T) {
	got, err := checkMarks("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		t.Error(f)
	}
}
