package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// planTaskIDRuleCmd is `flow-guard plan-task-id-rule <tasks.md>`, run by
// check-plan-shape.sh beside plan-script-paths on every tasks.md it scans. It
// refuses a `/flow` plan stating, on any line outside a fence, that task
// commits carry no `Task-Id:` trailer: close-task.sh refuses a `/flow` task
// commit without one, so an implementer told otherwise meets a contradiction
// (self-review-backlog-flow-5-oct18's Global constraints did). The rule is a
// fix loop's own (skills/flow-self-review/SKILL.md, step 3 item 1), so a loop
// plan carries it legitimately. A `/flow` plan is one whose change <name> has
// a spectre/<name> branch, local or on origin, in its repository — the test
// hooks/enforce-loop-dispatch.py applies, since a `/flow` change may itself be
// named self-review-…, while a loop's branch is never spectre/<its own name>.
// A repository git cannot read holds no such branch. An archived plan, or a
// file outside a `changes/<name>/` directory, is not checked: its text is
// history, judged when it was live. Exit 0 clean, 1 one `file:line:` line per
// statement, 2 a usage error or a file that cannot be read.
func init() { Registry["plan-task-id-rule"] = planTaskIDRuleCmd }

var ptirRE = regexp.MustCompile("(?i)\\bno\\s+`Task-Id:`")

func planTaskIDRuleCmd(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] == "" {
		fmt.Fprintln(stderr, "usage: flow-guard plan-task-id-rule <tasks.md>")
		return 2
	}
	path := args[0]
	if !filepath.IsAbs(path) {
		path = filepath.Join(env.Dir, path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[0], err)
		return 2
	}
	lines, err := tcfPlanLines(b)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read file: %v\n", args[0], err)
		return 2
	}
	changes := filepath.Dir(filepath.Dir(path))
	if filepath.Base(changes) != "changes" {
		return 0
	}
	name := filepath.Base(filepath.Dir(path))
	root := filepath.Dir(filepath.Dir(changes))
	if out, ok := gitOut(env, "-C", root, "for-each-ref", "--count=1", "--format=x",
		"refs/heads/spectre/"+name, "refs/remotes/origin/spectre/"+name); !ok || out != "x" {
		return 0
	}
	found := 0
	tcfUnfenced(lines, func(i int, line string) bool {
		if ptirRE.MatchString(line) {
			fmt.Fprintf(stdout, "%s:%d: a /flow plan states task commits carry no `Task-Id:` trailer — close-task.sh refuses a /flow task commit without one; that rule is a fix loop's alone\n",
				args[0], i+1)
			found++
		}
		return false
	})
	if found > 0 {
		return 1
	}
	return 0
}
