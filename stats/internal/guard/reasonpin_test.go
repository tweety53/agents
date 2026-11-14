package guard

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// exitSiteViolations lists every exit site in files (file name to Go source)
// that records no reason id first (decisions refusal-site-pin,
// pin-reach-extended). exempt maps a function key -- its name, or
// Type.Method for a method -- to why its integers are not its guard's exit
// code. The rules:
//
//	(a) a struct type with a field of type Env or *Env is an env struct; a
//	    function declaration carries an Env when a parameter has type Env or
//	    *Env or its receiver's base type is an env struct; a function literal
//	    carries one when a parameter does or it sits inside a function that
//	    carries one;
//	(b) a site is a return belonging to that function (not to a nested
//	    function literal) one of whose results is a non-zero integer literal
//	    or a call to CannotAnswer, or an expression statement
//	    panic(<identifier>(<non-zero integer literal>));
//	(c) a site in a function whose declaration's key is in exempt -- a
//	    function literal takes its enclosing declaration's key -- is never a
//	    violation;
//	(d) a site in a function that carries no Env is a violation;
//	(e) a site in a function that carries an Env is tagged when, in some
//	    statement list on the path from the function body down to the site, a
//	    statement before the one holding the site is an expression statement
//	    calling Reason, or is an if, for, range, switch, type-switch, select or
//	    block statement that contains a Reason call and no return -- nested
//	    function literals excluded on both counts; an untagged site is a
//	    violation;
//	(f) each violation reads `<file>:<line>: <function key>: <why>`.
func exitSiteViolations(files map[string]string, exempt map[string]string) []string {
	fset := token.NewFileSet()
	parsed := map[string]*ast.File{}
	names := make([]string, 0, len(files))
	for name, src := range files {
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			return []string{fmt.Sprintf("%s: parse: %v", name, err)}
		}
		parsed[name] = f
		names = append(names, name)
	}
	slices.Sort(names)
	envStructs := map[string]bool{}
	for _, f := range parsed {
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, sp := range g.Specs {
				ts, ok := sp.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, fld := range st.Fields.List {
					if pinIsEnv(fld.Type) {
						envStructs[ts.Name.Name] = true
					}
				}
			}
		}
	}
	var out []string
	for _, name := range names {
		for _, d := range parsed[name].Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			key, carries := fd.Name.Name, pinParamsEnv(fd.Type)
			if fd.Recv != nil && len(fd.Recv.List) > 0 {
				base := pinBaseType(fd.Recv.List[0].Type)
				key = base + "." + key
				carries = carries || envStructs[base]
			}
			if _, ok := exempt[key]; ok {
				continue
			}
			p := pinner{fset: fset, key: key}
			p.function(fd.Body, carries)
			out = append(out, p.out...)
		}
	}
	return out
}

type pinner struct {
	fset *token.FileSet
	key  string
	out  []string
}

// function checks one function body, then each function literal in it.
func (p *pinner) function(body *ast.BlockStmt, carries bool) {
	p.list(body.List, carries, false)
	ast.Inspect(body, func(n ast.Node) bool {
		lit, ok := n.(*ast.FuncLit)
		if !ok {
			return true
		}
		p.function(lit.Body, carries || pinParamsEnv(lit.Type))
		return false
	})
}

func (p *pinner) list(stmts []ast.Stmt, carries, tagged bool) {
	for _, s := range stmts {
		p.stmt(s, carries, tagged)
		if pinTags(s) {
			tagged = true
		}
	}
}

func (p *pinner) stmt(s ast.Stmt, carries, tagged bool) {
	switch s := s.(type) {
	case *ast.ReturnStmt:
		for _, r := range s.Results {
			if pinNonZero(r) || pinCalls(r, "CannotAnswer") {
				p.site(s, carries, tagged)
				return
			}
		}
	case *ast.ExprStmt:
		if c, ok := s.X.(*ast.CallExpr); ok && pinIdent(c.Fun, "panic") && len(c.Args) == 1 {
			if in, ok := c.Args[0].(*ast.CallExpr); ok && len(in.Args) == 1 && pinNonZero(in.Args[0]) {
				if _, ok := in.Fun.(*ast.Ident); ok {
					p.site(s, carries, tagged)
				}
			}
		}
	case *ast.BlockStmt:
		p.list(s.List, carries, tagged)
	case *ast.LabeledStmt:
		p.stmt(s.Stmt, carries, tagged)
	case *ast.IfStmt:
		p.list(s.Body.List, carries, tagged)
		if s.Else != nil {
			p.stmt(s.Else, carries, tagged)
		}
	case *ast.ForStmt:
		p.list(s.Body.List, carries, tagged)
	case *ast.RangeStmt:
		p.list(s.Body.List, carries, tagged)
	case *ast.SwitchStmt:
		p.clauses(s.Body, carries, tagged)
	case *ast.TypeSwitchStmt:
		p.clauses(s.Body, carries, tagged)
	case *ast.SelectStmt:
		p.clauses(s.Body, carries, tagged)
	}
}

func (p *pinner) clauses(body *ast.BlockStmt, carries, tagged bool) {
	for _, c := range body.List {
		switch c := c.(type) {
		case *ast.CaseClause:
			p.list(c.Body, carries, tagged)
		case *ast.CommClause:
			p.list(c.Body, carries, tagged)
		}
	}
}

func (p *pinner) site(s ast.Stmt, carries, tagged bool) {
	pos := p.fset.Position(s.Pos())
	switch {
	case !carries:
		p.out = append(p.out, fmt.Sprintf("%s:%d: %s: exit site in a function that carries no Env", pos.Filename, pos.Line, p.key))
	case !tagged:
		p.out = append(p.out, fmt.Sprintf("%s:%d: %s: exit site records no Reason first", pos.Filename, pos.Line, p.key))
	}
}

// pinTags reports whether s, sitting before a site in a statement list,
// tags it: a Reason call statement, or a compound statement holding a Reason
// call and no return (nested function literals excluded).
func pinTags(s ast.Stmt) bool {
	if e, ok := s.(*ast.ExprStmt); ok {
		return pinCalls(e.X, "Reason")
	}
	switch s.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt, *ast.BlockStmt:
	default:
		return false
	}
	reason, ret := false, false
	ast.Inspect(s, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			ret = true
		case *ast.CallExpr:
			if pinIdent(n.Fun, "Reason") {
				reason = true
			}
		}
		return true
	})
	return reason && !ret
}

func pinCalls(e ast.Expr, name string) bool {
	c, ok := e.(*ast.CallExpr)
	return ok && pinIdent(c.Fun, name)
}

func pinIdent(e ast.Expr, name string) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == name
}

func pinNonZero(e ast.Expr) bool {
	b, ok := e.(*ast.BasicLit)
	return ok && b.Kind == token.INT && strings.Trim(b.Value, "0") != ""
}

func pinIsEnv(e ast.Expr) bool {
	if s, ok := e.(*ast.StarExpr); ok {
		e = s.X
	}
	return pinIdent(e, "Env")
}

func pinParamsEnv(ft *ast.FuncType) bool {
	for _, f := range ft.Params.List {
		if pinIsEnv(f.Type) {
			return true
		}
	}
	return false
}

func pinBaseType(e ast.Expr) string {
	if s, ok := e.(*ast.StarExpr); ok {
		e = s.X
	}
	if ix, ok := e.(*ast.IndexExpr); ok {
		e = ix.X
	}
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// reasonPinRange reads every non-test .go file of this package's directory
// whose name keep accepts, keyed by file name.
func reasonPinRange(t *testing.T, keep func(name string) bool) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") || !keep(n) {
			continue
		}
		b, err := os.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		files[n] = string(b)
	}
	return files
}

// TestExitSitePinFlagsUntaggedReturns pins exitSiteViolations' rules on
// fixtures: each flagged function yields exactly one violation, each clean
// one none.
func TestExitSitePinFlagsUntaggedReturns(t *testing.T) {
	t.Parallel()
	const prelude = `package guard
type Env struct{}
func Reason(env Env, id string) {}
func CannotAnswer(name string) int { return 3 }
type xExit int
type holder struct{ env Env }
`
	flagged := map[string]string{
		"bare": `func bare(env Env) int { return 1 }`,
		"laterReturn": `func laterReturn(env Env, a bool) int {
	if a {
		Reason(env, "x")
		return 1
	}
	return 2
}`,
		"dieLit": `func dieLit(env Env) int {
	die := func(m string) int { return 2 }
	return die("m")
}`,
		"h":       `func h() int { return 3 }`,
		"cannot":  `func cannot(env Env) int { return CannotAnswer("g") }`,
		"panicky": `func (v *holder) panicky() { panic(xExit(2)) }`,
		"litTag": `func litTag(env Env) int {
	func() { Reason(env, "x") }()
	return 1
}`,
	}
	clean := map[string]string{
		"direct": `func direct(env Env) int {
	Reason(env, "x")
	return 1
}`,
		"loop": `func loop(env Env, xs []int) int {
	n := 0
	for _, x := range xs {
		if x > 0 {
			Reason(env, "x")
			n++
		}
	}
	if n > 0 {
		return 1
	}
	return 0
}`,
		"twoIfs": `func twoIfs(env Env, a, b bool) int {
	if a {
		Reason(env, "a")
	}
	if b {
		Reason(env, "b")
	}
	return 1
}`,
		"h":     `func h() int { return 3 }`,
		"zero":  `func zero(env Env) (string, int) { return "", 0 }`,
		"zero1": `func zero1(env Env) int { return 0 }`,
		"panicTagged": `func (v *holder) panicTagged() {
	Reason(v.env, "x")
	panic(xExit(2))
}`,
		"pair": `func pair(env Env) (string, int) {
	Reason(env, "x")
	return "", 2
}`,
	}
	for name, src := range flagged {
		got := exitSiteViolations(map[string]string{"f.go": prelude + src}, map[string]string{"CannotAnswer": "the code table"})
		if len(got) != 1 {
			t.Errorf("flagged %s: violations %q, want exactly one", name, got)
		}
	}
	for name, src := range clean {
		exempt := map[string]string{"CannotAnswer": "the code table"}
		if name == "h" {
			exempt["h"] = "returns a count"
		}
		if got := exitSiteViolations(map[string]string{"f.go": prelude + src}, exempt); len(got) != 0 {
			t.Errorf("clean %s: violations %q, want none", name, got)
		}
	}
}

// reasonPinExempt names every function whose integers are not its guard's
// exit code, with why; its callers record at their own exits.
var reasonPinExempt = map[string]string{
	"psWorse":               "combines exit codes its callers recorded",
	"psPlan.afterFieldLine": "returns a line number",
	"bpRunArgv":             "returns a child process's exit status",
	"exitCode":              "returns a child process's exit status",
	"pcFlow":                "returns the flow CLI's exit status",
	"changePlanStoreDir":    "returns a lookup status (0 found, 1 none, 3 ambiguous), not an exit code",
	"changePlanDir":         "returns a lookup status (0 found, 1 none, 3 ambiguous), not an exit code",
	"ccRunSurvivors":        "returns the survivors command's exit status",
	"rcOf":                  "returns a child process's exit status",
	"mvRunCommand":          "returns a child process's exit status",
	"mvpStep":               "returns a step, not an exit code",
	"ownedCorpus":           "returns a cannot-answer its callers record",
	"checkTreeRestored":     "returns a status post-mutation-check records at its own exit",
	"gitExec":               "returns a child process's exit status",
	"ppListMarkerEnd":       "returns a position",
	"ppCheckboxEnd":         "returns a position",
	"ppResolveExitCode":     "combines codes recorded where each finding is found",
	"pwExitCode":            "returns a child process's exit status",
	"pfState":               "returns a status its callers record at their own exits",
	"CannotAnswer":          "is the cannot-answer code table",
	"rgiRun":                "returns a child process's exit status",
	"rcInHunk":              "counts lines",
	"wsExitCode":            "returns a child process's exit status",
}

// reasonIDRE is a reason id outside check-plan-shape: local kebab-case
// (decision reason-id-shape).
var reasonIDRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// TestEveryExitSiteRecordsAReason is the package-wide reason pin (decisions
// refusal-site-pin, pin-per-range): every non-zero exit site of every
// non-test file records a reason first, and every literal id is kebab-case
// or a check-plan-shape rule id.
func TestEveryExitSiteRecordsAReason(t *testing.T) {
	t.Parallel()
	files := reasonPinRange(t, func(string) bool { return true })
	for _, v := range exitSiteViolations(files, reasonPinExempt) {
		t.Error(v)
	}
	rules := map[string]bool{}
	for _, r := range psRules {
		rules[r.id] = true
	}
	fset := token.NewFileSet()
	for name, src := range files {
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok || !pinIdent(c.Fun, "Reason") || len(c.Args) != 2 {
				return true
			}
			lit, ok := c.Args[1].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			id, _ := strconv.Unquote(lit.Value)
			if !reasonIDRE.MatchString(id) && !rules[id] {
				t.Errorf("%s: reason id %q is neither kebab-case nor a check-plan-shape rule id", fset.Position(lit.Pos()), id)
			}
			return true
		})
	}
}
