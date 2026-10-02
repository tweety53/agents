package guard

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// check-fast-route-record reads a /flow-fast branch's commit series as its
// record (KAN-838): the fast route writes no plan, so the series is the
// whole record and this guard is what reads it back before the route lands
// anything. Every commit since the merge base of HEAD and the base must
// carry a Conventional Commits subject (lowercase type, optional non-empty
// scope, optional breaking `!`, a non-empty subject after ": "), a scope
// naming a module — never the change name, a Jira key or a task id — and a
// body with no attribution trailer (`Co-Authored-By:`, a `Generated
// with/by` footer) and no `Task-Id:` trailer. A merge commit is judged like
// any other: the fast route's series is linear by construction, so one on
// the branch is exactly the mislabelled commit this guard exists to catch.
//
//	check-fast-route-record.sh <worktree> <base>
//
//	Exit 0   every commit reads as the record — or the series is empty.
//	Exit 1   at least one finding, one `<sha> <finding>` line per hit on
//	         stdout, walk order oldest first.
//	Exit 2   cannot answer — wrong argument count, <worktree> absent or
//	         not a git worktree, <base> resolving neither as origin/<base>
//	         nor as a revision of its own, or a failed merge-base or
//	         rev-list.
func init() {
	Registry["check-fast-route-record"] = checkFastRouteRecord
}

var (
	frrSubject = regexp.MustCompile(`^[a-z]+\([^)]+\)!?: \S|^[a-z]+!?: \S`)
	frrScope   = regexp.MustCompile(`^[a-z]+\(([^)]+)\)`)
	frrJiraKey = regexp.MustCompile(`(?i)[a-z]{2,10}-[0-9]+`)
	frrTaskID  = regexp.MustCompile(`^([0-9]+|task-[0-9]+|[0-9]+/[0-9]+)$`)
	frrAttrCo  = regexp.MustCompile(`(?i)^co-authored-by:`)
	frrAttrGen = regexp.MustCompile(`(?i)^generated (with|by)\b`)
	frrTaskRef = regexp.MustCompile(`(?i)^task-id:`)
)

func checkFastRouteRecord(args []string, env Env, stdout, stderr io.Writer) int {
	const self = "check-fast-route-record"
	die := func(format string, a ...any) int {
		fmt.Fprintf(stderr, self+": "+format+"\n", a...)
		return 2
	}
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: check-fast-route-record.sh <worktree> <base>")
		return 2
	}
	wt := pcAbs(env, args[0])
	if !isDir(wt) {
		return die("%s is not a directory", args[0])
	}
	git := envGit(env)
	if git("-C", wt, "rev-parse", "--git-dir").Run() != nil {
		return die("%s is not a git worktree", args[0])
	}
	// <base> is the bare name resolve-base-branch.sh prints: origin/<base>
	// first — the fetched base a fast route lands against — the name itself
	// as a revision second, so a recorded local base or a raw sha still
	// resolves.
	resolved, ok := capture(git("-C", wt, "rev-parse", "--verify", "--quiet", "origin/"+args[1]+"^{commit}"))
	if !ok {
		resolved, ok = capture(git("-C", wt, "rev-parse", "--verify", "--quiet", args[1]+"^{commit}"))
	}
	if !ok {
		return die("%s resolves neither as origin/%s nor as a revision", args[1], args[1])
	}
	mergeBase, ok := capture(git("-C", wt, "merge-base", "HEAD", resolved))
	if !ok {
		return die("cannot merge-base HEAD with %s", resolved)
	}
	branch, _ := capture(git("-C", wt, "branch", "--show-current"))
	series, ok := capture(git("-C", wt, "rev-list", "--reverse", mergeBase+"..HEAD"))
	if !ok {
		return die("cannot list %s..HEAD", mergeBase)
	}
	findings := 0
	say := func(sha, format string, a ...any) {
		findings++
		fmt.Fprintf(stdout, sha+" "+format+"\n", a...)
	}
	for _, sha := range strings.Fields(series) {
		msg, ok := capture(git("-C", wt, "show", "-s", "--format=%s%x00%b", sha))
		if !ok {
			return die("cannot read %s", sha)
		}
		subject, body, _ := strings.Cut(msg, "\x00")
		if !frrSubject.MatchString(subject) {
			say(sha, "subject not in Conventional Commits form: %s", subject)
		}
		if m := frrScope.FindStringSubmatch(subject); m != nil {
			scope := m[1]
			switch {
			case branch != "" && strings.EqualFold(scope, branch):
				say(sha, "scope %q names the change", scope)
			case frrTaskID.MatchString(strings.ToLower(scope)):
				say(sha, "scope %q names a task id", scope)
			case frrJiraKey.MatchString(scope):
				say(sha, "scope %q names a Jira key", scope)
			}
		}
		for _, line := range strings.Split(body, "\n") {
			trimmed := strings.TrimSpace(line)
			switch {
			case frrAttrCo.MatchString(line):
				say(sha, "attribution trailer: %s", trimmed)
			case frrAttrGen.MatchString(line):
				say(sha, "attribution footer: %s", trimmed)
			case frrTaskRef.MatchString(line):
				say(sha, "Task-Id trailer: %s", trimmed)
			}
		}
	}
	if findings > 0 {
		return 1
	}
	return 0
}
