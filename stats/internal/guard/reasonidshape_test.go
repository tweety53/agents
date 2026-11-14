package guard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"testing"
)

// fragmentIDRE matches a reason id that ends on a function word — a message
// prefix cut mid-phrase (`failed-in`, `change-name-is-not`) rather than a name
// for the reason (tagging rule 2, decision reason-id-shape).
var fragmentIDRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*-(for|with|is|the|a|to|of|in|on|at|and|or|from|by|its|no|not|than|after|under|must|but|into|an|whether|while|cannot)$`)

// fragmentIDExempt holds ids that end on such a word and still name a reason.
var fragmentIDExempt = map[string]string{}

// bareSubjectIDs are ids that name only the subject of a message — its first
// word or words cut off before the verb (`worktree`, `change-name`) — and so
// no reason at all.
var bareSubjectIDs = map[string]bool{
	"no": true, "only": true, "worktree": true, "change-root": true, "file": true,
	"declares": true, "slot-carries": true, "finding": true, "flowd": true,
	"stop": true, "record": true, "change-name": true, "peer-name": true,
	"tasks-md": true, "merge-base": true, "base-branch": true, "recorded-base": true,
	"loop": true, "crash": true,
}

// singleWordFunctionRE matches an id that is one function word alone.
var singleWordFunctionRE = regexp.MustCompile(`^(for|with|is|the|a|to|of|in|on|at|and|or|from|by|its|no|not|than|after|under|must|but|into|an|whether|while|cannot|only)$`)

// reasonIDAllowed holds single-word ids that genuinely name a reason.
var reasonIDAllowed = map[string]string{
	"usage": "the shared id for a malformed invocation, the reason every guard's usage exit names",
}

// badReasonID reports whether id is a fragment rather than a reason name.
func badReasonID(id string) bool {
	if reasonIDAllowed[id] != "" {
		return false
	}
	return (fragmentIDRE.MatchString(id) && fragmentIDExempt[id] == "") || bareSubjectIDs[id] || singleWordFunctionRE.MatchString(id)
}

// reasonIDCallees are the functions and methods whose first argument is a
// reason id: Reason's second, and the closures and methods that route an exit
// through Reason (tagging rule 3) their first.
var reasonIDCallees = map[string]bool{
	"die": true, "refuse": true, "fail": true, "stop": true, "cannotAnswer": true, "usage": true,
}

// reasonIDLiterals returns every literal reason id in src keyed by position.
func reasonIDLiterals(t *testing.T, fset *token.FileSet, name, src string) map[string]string {
	t.Helper()
	f, err := parser.ParseFile(fset, name, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		var arg ast.Expr
		switch fn := c.Fun.(type) {
		case *ast.Ident:
			if fn.Name == "Reason" && len(c.Args) == 2 {
				arg = c.Args[1]
			} else if reasonIDCallees[fn.Name] && len(c.Args) > 0 {
				arg = c.Args[0]
			}
		case *ast.SelectorExpr:
			if reasonIDCallees[fn.Sel.Name] && len(c.Args) > 0 {
				arg = c.Args[0]
			}
		}
		if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			id, _ := strconv.Unquote(lit.Value)
			ids[fset.Position(lit.Pos()).String()] = id
		}
		return true
	})
	return ids
}

// TestReasonIDsNameTheReason fails on any reason id in the package that is a
// truncated message fragment.
func TestReasonIDsNameTheReason(t *testing.T) {
	t.Parallel()
	fset := token.NewFileSet()
	for name, src := range reasonPinRange(t, func(string) bool { return true }) {
		for pos, id := range reasonIDLiterals(t, fset, name, src) {
			if badReasonID(id) {
				t.Errorf("%s: reason id %q is a message fragment, not a name for the reason", pos, id)
			}
		}
	}
}

// TestReasonIDsNameTheReasonFlagsAFragment proves the scan sees ids at each
// call shape it reads.
func TestReasonIDsNameTheReasonFlagsAFragment(t *testing.T) {
	t.Parallel()
	src := `package guard
func f(env Env, v *x) int {
	Reason(env, "failed-in")
	v.die("head-does-not-resolve-in", "")
	Reason(env, "no")
	Reason(env, "worktree")
	v.stop("change-name", "")
	return refuse("usage", "")
}`
	got := map[string]bool{}
	for _, id := range reasonIDLiterals(t, token.NewFileSet(), "f.go", src) {
		got[id] = badReasonID(id)
	}
	want := map[string]bool{"failed-in": true, "head-does-not-resolve-in": true, "no": true, "worktree": true, "change-name": true, "usage": false}
	for id, frag := range want {
		if f, ok := got[id]; !ok || f != frag {
			t.Errorf("id %q: found=%v fragment=%v, want found fragment=%v", id, ok, f, frag)
		}
	}
}
