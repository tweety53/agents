package guard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// fragmentIDRE matches a reason id that ends on a function word — a message
// prefix cut mid-phrase (`failed-in`, `change-name-is-not`) rather than a name
// for the reason (tagging rule 2, decision reason-id-shape).
var fragmentIDRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*-(for|with|is|the|a|to|of|in|on|at|and|or|from|by|its|no|not|than|after|under|must|but|into|an|whether|while|cannot|onto|outside|via|as|against|through|over)$`)

// reasonIDExempt holds multi-word ids the rule would flag that still name a
// reason, each with why (decision sub-guard-names-exempt).
var reasonIDExempt = map[string]string{}

// reasonVerdictWords are the state, verdict and predicate words: a multi-word
// id names a reason only when one of its words is among them; without one it
// names only a subject (`change-name`, `setup-sh`).
var reasonVerdictWords = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`aborted absent absolute ambiguous appended bad behind cannot capped carries changed changes clean conflict crashed current dangles demonstrated denied detached differs dirty disagree done dotted drift duplicate embedded empty escapes executable exist exists exited fail failed failing failure findings flaky forced foreign found gap given gone green hit incomplete inside interrupted invalid is joined late leftover lost malformed mismatch mismatched miss misses missing moved must names needs no not nothing off ok only open outside outstanding over overlap pending positive readable red refused refusing rejected relative required reused runnable running sandboxed stale still stopped stray swept timed too touched touches twice unaddressed unchecked uncited unclassifiable unclosed uncommitted uncounted undercounted underivable undispatched unfetchable unfinished unknown unnarrowable unparseable unproved unreachable unreadable unrecognized unrecorded unresolved unreviewed unset unstaged untasked unticked untracked unusable unverifiable unverified vanished violation violations without writable wrong zero`) {
		reasonVerdictWords[w] = true
	}
}

// reasonIDAllowed is the whole set of single-word ids: one word names a
// reason only when it is a state or verdict on its own, never a subject.
var reasonIDAllowed = map[string]string{
	"usage":        "the shared id for a malformed invocation, the reason every guard's usage exit names",
	"conflict":     "a merge or rebase stopped on conflicting hunks",
	"overlap":      "two sets that must be disjoint share a member",
	"unreadable":   "an input the guard must read could not be read",
	"capped":       "the command was cut off at its cap",
	"interrupted":  "the run was interrupted before it finished",
	"sandboxed":    "the run was refused by the sandbox",
	"uncited":      "a citation the guard requires is absent",
	"untasked":     "a commit carries no Task-Id",
	"unverifiable": "the verdict cannot be established",
	"changed":      "a tree that must stay unchanged changed",
}

// idShapeRE matches a kebab-case reason id.
var idShapeRE = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// badReasonID reports whether id is a fragment or a bare subject rather than
// a reason name: any single-word id outside reasonIDAllowed, and any
// multi-word id outside reasonIDExempt that ends on a function word or
// carries no word of reasonVerdictWords.
func badReasonID(id string) bool {
	if !idShapeRE.MatchString(id) {
		return false // a message argument of a same-named helper, not an id
	}
	if !strings.Contains(id, "-") {
		return reasonIDAllowed[id] == ""
	}
	if reasonIDExempt[id] != "" {
		return false
	}
	if fragmentIDRE.MatchString(id) {
		return true
	}
	for _, w := range strings.Split(id, "-") {
		if reasonVerdictWords[w] {
			return false
		}
	}
	return true
}

// reasonIDCallees are the functions and methods whose first argument is a
// reason id: Reason's second, and the closures and methods that route an exit
// through Reason (tagging rule 3) their first.
var reasonIDCallees = map[string]bool{
	"die": true, "refuse": true, "fail": true, "stop": true, "stopAs": true, "cannotAnswer": true, "usage": true,
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

// TestReasonIDsNameTheReason fails on any reason id in the package that ends on
// a function word or, multi-word, carries no verdict word.
func TestReasonIDsNameTheReason(t *testing.T) {
	t.Parallel()
	fset := token.NewFileSet()
	for name, src := range reasonPinRange(t, func(string) bool { return true }) {
		for pos, id := range reasonIDLiterals(t, fset, name, src) {
			if badReasonID(id) {
				t.Errorf("%s: reason id %q names no reason: it ends on a function word or carries no verdict word (reasonVerdictWords)", pos, id)
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
	c.stopAs("push", "push", "")
	Reason(env, "pair")
	return refuse("usage", "")
}`
	got := map[string]bool{}
	for _, id := range reasonIDLiterals(t, token.NewFileSet(), "f.go", src) {
		got[id] = badReasonID(id)
	}
	want := map[string]bool{"failed-in": true, "head-does-not-resolve-in": true, "no": true, "worktree": true, "change-name": true, "push": true, "pair": true, "usage": false}
	for id, frag := range want {
		if f, ok := got[id]; !ok || f != frag {
			t.Errorf("id %q: found=%v fragment=%v, want found fragment=%v", id, ok, f, frag)
		}
	}
}

// TestMultiWordReasonIDsCarryAVerdict pins both halves of the multi-word rule:
// a verdict word somewhere in the id, and no function word at its end.
func TestMultiWordReasonIDsCarryAVerdict(t *testing.T) {
	t.Parallel()
	for id, want := range map[string]bool{
		"visual-verify": true, "path-token": true, "setup-sh": true, "review-panel-md": true,
		"reading-the-prompt": true, "fix-the-reference": true, "identical-none": true,
		"git-refused-to-rebase-onto": true, "target-file-is-outside": true, "change-name": true,
		"not-a-worktree": false, "cannot-read-file": false, "line-is-a-template": false,
		"worktree-vanished": false, "bundle-violations": false, "base-behind": false,
		"usage": false,
	} {
		if got := badReasonID(id); got != want {
			t.Errorf("badReasonID(%q) = %v, want %v", id, got, want)
		}
	}
}
