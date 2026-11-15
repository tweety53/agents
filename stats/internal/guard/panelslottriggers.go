package guard

import (
	"fmt"
	"io"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// flow-guard panel-slot-triggers — decide, from a change's own diff, whether
// review pass 1 needs each of the three gated slots (KAN-1126). This comment
// is the contract skills/flow/review-panel.md's `### The slot triggers` cites
// as canonical.
//
// Usage: flow-guard panel-slot-triggers <worktree> <merge-base> [<worktree> <merge-base>…]
//
// Output, on exit 0, is exactly three lines in this fixed order:
//
//	test-sanity: run|skip — <reason>
//	call-efficiency: run|skip — <reason>
//	failure-modes: run|skip — <reason>
//
// A slot runs when any worktree pair runs it; its <reason> is that pair's
// first match, `<path>:<line> <rule>` for a content rule and `<path> <rule>`
// for a path rule. A skipped slot's reason is `no <slot> signal in the
// changed code`.
//
// WHAT IS READ, per pair: the change's touched paths — committed since the
// merge base, staged and unstaged, with the worktree's `## panel diff
// exclude` applied (paneltouchedpaths.go) — and the changed lines of one
// `git diff --no-renames -U0 <merge-base>` against the working tree under
// the same excludes: added lines at their new line number, removed lines at
// their old one, so a deleted call still counts. Paths are walked in the
// touched-path order; within a path its changed lines are tried first, in
// diff order, then the path itself.
//
// test-sanity runs when a touched path — added, changed or deleted — is a
// test path:
//
//	go-test       _test.go
//	jvm-test      (Test|Tests|Spec|IT).(kt|java)
//	src-test-dir  src/<name containing test or Test>/
//	js-test       .(test|spec).[cm]?[jt]sx?
//	test-dir      tests?/, __tests__/, e2e/, testdata/, fixtures?/
//	snapshot      __snapshots__/, -snapshots/
//	test-config   {jest,vitest,playwright}.config.*
//
// call-efficiency and failure-modes never read a doc path (isDocPath), a
// lockfile (go.sum, package-lock.json, yarn.lock, pnpm-lock.yaml,
// Cargo.lock, *.lockfile) or a test path. call-efficiency runs on a changed
// line matching
//
//	http-handler    ServeHTTP, HandleFunc, *http.Request
//	spring-handler  @…Mapping, @RestController, `routing {`
//	next-handler    export [async] function GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS, NextResponse
//	query           .Query(, .QueryRow(, .Exec( and their …Context( forms
//	sql             SELECT, INSERT INTO, UPDATE … SET, DELETE FROM
//	jpa-query       @Query
//	http-client     http.Get, http.Post, http.NewRequest
//	fetch           fetch(
//	axios           axios
//	http-client-jvm HttpClient
//	react-query     useQuery, useMutation
//
// or on a path matching
//
//	handler-dir     api|handlers?|controllers?|routes?|presentation|store|
//	                persistence|repositor(y|ies)|clients?|integrations/
//	sql-file        .sql
//
// failure-modes runs on everything call-efficiency runs on, plus a changed
// line matching
//
//	file-io           os.Open|Create|ReadFile|WriteFile|Rename|Remove|MkdirAll
//	exec              exec.Command
//	goroutine         go func
//	sync              sync., atomic.
//	channel           chan, <-, `select {`
//	context           context.With
//	kotlin-coroutine  launch, async, withContext, Mutex, suspend fun
//	ts-async          await, Promise., setTimeout, localStorage, fs.
//
// or a path matching
//
//	worker-dir        scheduler|jobs?|workers?|migrations/
//
// File and exec calls are failure-modes only: the call-efficiency
// condition names handlers, queries and client calls.
//
// KNOWN CEILING: a query or call reached through an unnamed project helper
// outside the listed directories — a second capture(panelGit(…)) in a guard,
// a composable calling a repository method — is invisible here. The
// plan-time slot judge (flow-guard slot-judge) closes it where the plan names
// the work; a slot still skips when neither sees it.
//
// Exit codes:
//
//	0  answered: the three lines on stdout
//	2  cannot answer — no arguments or an odd count, a pair whose worktree is
//	   not a directory or not a git worktree or whose merge base does not
//	   resolve, or a git call that failed; nothing on stdout, and every gated
//	   slot runs
//
// There is no exit 1: three verdicts do not fit one bit.
func init() {
	Registry["panel-slot-triggers"] = panelSlotTriggers
}

// pstRule is one named path or content pattern.
type pstRule struct {
	name string
	re   *regexp.Regexp
}

func pstRules(pairs ...string) []pstRule {
	var rs []pstRule
	for i := 0; i < len(pairs); i += 2 {
		rs = append(rs, pstRule{pairs[i], regexp.MustCompile(pairs[i+1])})
	}
	return rs
}

var (
	pstTestPaths = pstRules(
		"go-test", `_test\.go$`,
		"jvm-test", `(Test|Tests|Spec|IT)\.(kt|java)$`,
		"src-test-dir", `(^|/)src/[A-Za-z]*[Tt]est[A-Za-z]*/`,
		"js-test", `\.(test|spec)\.[cm]?[jt]sx?$`,
		"test-dir", `(^|/)(tests?|__tests__|e2e|testdata|fixtures?)/`,
		"snapshot", `__snapshots__/|-snapshots/`,
		"test-config", `(^|/)(jest|vitest|playwright)\.config\.[^/]*$`,
	)
	pstCallLines = pstRules(
		"http-handler", `ServeHTTP|HandleFunc|\*http\.Request`,
		"spring-handler", `@[A-Za-z]*Mapping\b|@RestController|routing \{`,
		"next-handler", `export (async )?function (GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\b|NextResponse`,
		"query", `\.(Query|QueryRow|Exec)(Context)?\(`,
		"sql", `\b(SELECT|INSERT INTO|DELETE FROM)\b|\bUPDATE\b.*\bSET\b`,
		"jpa-query", `@Query\b`,
		"http-client", `http\.(Get|Post|NewRequest)\b`,
		"fetch", `\bfetch\(`,
		"axios", `\baxios\b`,
		"http-client-jvm", `HttpClient`,
		"react-query", `\b(useQuery|useMutation)\b`,
	)
	pstCallPaths = pstRules(
		"handler-dir", `(^|/)(api|handlers?|controllers?|routes?|presentation|store|persistence|repositor(y|ies)|clients?|integrations)/`,
		"sql-file", `\.sql$`,
	)
	pstFailLines = append(append([]pstRule{}, pstCallLines...), pstRules(
		"file-io", `\bos\.(Open|Create|ReadFile|WriteFile|Rename|Remove|MkdirAll)`,
		"exec", `\bexec\.Command`,
		"goroutine", `\bgo func\b`,
		"sync", `\b(sync|atomic)\.`,
		"channel", `\bchan\b|<-|select \{`,
		"context", `\bcontext\.With`,
		"kotlin-coroutine", `\b(launch|async|withContext|Mutex)\b|suspend fun`,
		"ts-async", `\bawait\b|\bPromise\.|\bsetTimeout\b|\blocalStorage\b|\bfs\.`,
	)...)
	pstFailPaths = append(append([]pstRule{}, pstCallPaths...), pstRules(
		"worker-dir", `(^|/)(scheduler|jobs?|workers?|migrations)/`,
	)...)

	pstLockfiles = []string{"go.sum", "package-lock.json", "yarn.lock", "pnpm-lock.yaml", "Cargo.lock"}
	pstHunk      = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)
)

// pstSlots is the fixed output order.
var pstSlots = []string{"test-sanity", "call-efficiency", "failure-modes"}

// pstLine is one changed line of the diff.
type pstLine struct {
	n    int
	text string
}

func panelSlotTriggers(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "panel-slot-triggers"
	if len(args) == 0 || len(args)%2 != 0 {
		fmt.Fprintf(stderr, "%s: usage: flow-guard %s <worktree> <merge-base> [<worktree> <merge-base>…]\n", prog, prog)
		Reason(env, "usage")
		return 2
	}
	gitBin, ok := panelResolveGit(env, prog, stderr)
	if !ok {
		Reason(env, "git-not-found")
		return 2
	}
	reasons := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		worktree, mergebase := args[i], args[i+1]
		if !panelValidateWorktree(env, prog, worktree, mergebase, gitBin, stderr) {
			Reason(env, "invalid-worktree")
			return 2
		}
		paths, ok := panelTouchedPaths(env, prog, worktree, mergebase, gitBin, stderr)
		if !ok {
			Reason(env, "cannot-list-touched-paths")
			return 2
		}
		diff, ok := capture(panelGit(env, gitBin, worktree, append([]string{"diff", "--no-renames", "--no-color", "--no-ext-diff",
			"--src-prefix=a/", "--dst-prefix=b/", "-U0", "--end-of-options", mergebase}, wpdExcludes(smcAbs(env, worktree))...)...))
		if !ok {
			fmt.Fprintf(stderr, "%s: cannot read this change's diff in %s\n", prog, worktree)
			Reason(env, "cannot-read-diff")
			return 2
		}
		for i, q := range paths {
			paths[i] = pstUnquote(q)
		}
		lines := pstChangedLines(diff)
		for slot, r := range pstDecide(paths, lines) {
			if _, seen := reasons[slot]; !seen {
				reasons[slot] = r
			}
		}
	}
	for _, slot := range pstSlots {
		if r, ok := reasons[slot]; ok {
			fmt.Fprintf(stdout, "%s: run — %s\n", slot, r)
		} else {
			fmt.Fprintf(stdout, "%s: skip — no %s signal in the changed code\n", slot, slot)
		}
	}
	return 0
}

// pstDecide returns, for each slot that runs, its first match over paths in
// order: a path's changed lines first, then the path itself.
func pstDecide(paths []string, lines map[string][]pstLine) map[string]string {
	out := map[string]string{}
	for _, p := range paths {
		if r, ok := pstMatch(pstTestPaths, p); ok && out["test-sanity"] == "" {
			out["test-sanity"] = p + " " + r
		}
		if isDocPath(p) || pstIsLockfile(p) || pstIsTest(p) {
			continue
		}
		for _, s := range []struct {
			slot         string
			lines, paths []pstRule
		}{
			{"call-efficiency", pstCallLines, pstCallPaths},
			{"failure-modes", pstFailLines, pstFailPaths},
		} {
			if out[s.slot] != "" {
				continue
			}
			for _, l := range lines[p] {
				if r, ok := pstMatch(s.lines, l.text); ok {
					out[s.slot] = p + ":" + strconv.Itoa(l.n) + " " + r
					break
				}
			}
			if r, ok := pstMatch(s.paths, p); ok && out[s.slot] == "" {
				out[s.slot] = p + " " + r
			}
		}
	}
	return out
}

func pstMatch(rules []pstRule, s string) (string, bool) {
	for _, r := range rules {
		if r.re.MatchString(s) {
			return r.name, true
		}
	}
	return "", false
}

func pstIsTest(p string) bool {
	_, ok := pstMatch(pstTestPaths, p)
	return ok
}

func pstIsLockfile(p string) bool {
	base := path.Base(p)
	for _, l := range pstLockfiles {
		if base == l {
			return true
		}
	}
	return strings.HasSuffix(base, ".lockfile")
}

// pstUnquote decodes a name git C-quoted; a name not opening with `"`, or
// one strconv.Unquote refuses, is returned as written.
func pstUnquote(s string) string {
	if !strings.HasPrefix(s, `"`) {
		return s
	}
	if u, err := strconv.Unquote(s); err == nil {
		return u
	}
	return s
}

// pstChangedLines parses a -U0 diff into each path's changed lines. A
// `diff --git` line opens a file header, which ends at its first `@@`; in
// it `--- a/<p>` and `+++ b/<p>` name the path (`---` for a deletion, whose
// `+++` is /dev/null), less the tab git ends a header with when the name
// holds a space; panelGit's core.quotePath=false leaves a non-ASCII name
// unquoted, and a name holding `"`, `\` or a control character, which git
// C-quotes whatever core.quotePath says (`--- "a/<p>"`), is decoded by
// pstUnquote before the prefix is dropped. Each hunk header sets the old and new counters; a
// `+` line is added at the new one, a `-` line removed at the old one.
func pstChangedLines(diff string) map[string][]pstLine {
	out := map[string][]pstLine{}
	var p string
	var header bool
	var oldN, newN int
	for _, l := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(l, "diff --git "):
			header, p = true, ""
		case header && strings.HasPrefix(l, "--- a/"):
			p = strings.TrimSuffix(strings.TrimPrefix(l, "--- a/"), "\t")
		case header && strings.HasPrefix(l, "+++ b/"):
			p = strings.TrimSuffix(strings.TrimPrefix(l, "+++ b/"), "\t")
		case header && strings.HasPrefix(l, `--- "`):
			p = strings.TrimPrefix(pstUnquote(l[4:]), "a/")
		case header && strings.HasPrefix(l, `+++ "`):
			p = strings.TrimPrefix(pstUnquote(l[4:]), "b/")
		case strings.HasPrefix(l, "@@"):
			header = false
			if m := pstHunk.FindStringSubmatch(l); m != nil {
				oldN, _ = strconv.Atoi(m[1])
				newN, _ = strconv.Atoi(m[2])
			}
		case header:
		case strings.HasPrefix(l, "+"):
			out[p] = append(out[p], pstLine{newN, l[1:]})
			newN++
		case strings.HasPrefix(l, "-"):
			out[p] = append(out[p], pstLine{oldN, l[1:]})
			oldN++
		}
	}
	return out
}
