package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// planScriptPathsCmd is `flow-guard plan-script-paths <tasks.md>`, run by
// check-plan-shape.sh beside plan-suite-steps on every tasks.md it scans. It
// refuses a task body citing, in an inline code span outside a fence, a
// `scripts/…` path the repository does not carry — a step that names a
// retired guard runs nothing (kan-991's step 6 cited a check-vocabulary.sh
// 52f9660b had deleted). The repository is the one whose
// `<spec-root>/changes/<name>/tasks.md` this is; a path any task's
// `**Files:**` names is the plan's own to create or delete and passes; a
// token carrying `<`, `*`, `{` or `$` is a placeholder, not a path. An
// archived plan, or a file outside a `changes/<name>/` directory, is not
// checked: its citations are history, judged when it was live. Exit 0 clean,
// 1 one `file:line: task <id> cites … ` line per missing path, 2 a usage
// error or a file that cannot be read.
func init() { Registry["plan-script-paths"] = planScriptPathsCmd }

var pspTokenRE = regexp.MustCompile(`(^|[\s(=])(scripts/[^\s` + "`" + `]+)`)

func planScriptPathsCmd(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] == "" {
		fmt.Fprintln(stderr, "usage: flow-guard plan-script-paths <tasks.md>")
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
	root := filepath.Dir(filepath.Dir(changes))
	tasks := tcfIterTasks(lines)
	declared := map[string]bool{}
	for _, task := range tasks {
		if t, ok := tcfParseTask(lines, task.id); ok {
			for _, f := range t.files {
				declared[f] = true
			}
		}
	}
	found := 0
	for _, task := range tasks {
		var text strings.Builder
		var starts, offsets []int
		tcfUnfenced(task.lines, func(offset int, line string) bool {
			starts = append(starts, text.Len())
			offsets = append(offsets, offset)
			text.WriteString(line + "\n")
			return false
		})
		joined := text.String()
		for _, m := range pssSpanRE.FindAllStringSubmatchIndex(joined, -1) {
			for _, tm := range pspTokenRE.FindAllStringSubmatch(joined[m[2]:m[3]], -1) {
				p := strings.TrimRight(tm[2], ".,;:)'\"")
				if i := strings.IndexByte(p, ':'); i >= 0 {
					p = p[:i]
				}
				if strings.ContainsAny(p, "<*{$") || declared[p] {
					continue
				}
				if _, err := os.Lstat(filepath.Join(root, p)); err == nil {
					continue
				}
				i := 0
				for i+1 < len(starts) && starts[i+1] <= m[0] {
					i++
				}
				fmt.Fprintf(stdout, "%s:%d: task %s cites `%s`, which the repository does not carry — name a path that exists, or list it in a task's **Files:** when the plan creates it\n",
					args[0], task.taskLine+1+offsets[i], task.id, p)
				found++
			}
		}
	}
	if found > 0 {
		return 1
	}
	return 0
}
