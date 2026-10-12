package guard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// planSuiteStepsCmd is `flow-guard plan-suite-steps <tasks.md>`, run by
// check-plan-shape.sh beside check-plan-shape.py on every tasks.md it scans.
// It refuses a task body that runs a module or repository test suite: an
// inline code span outside a fence whose command — at its start or after
// `&&`, `;` or `|` — invokes Gradle on a test task with no
// `--tests`, `go test` with no `-run`, or `npm test`/vitest with no `-t`. A
// task verifies with its lint and its own `**Tests:**` selector; the full
// `## test` list runs once, after the last task (**The parent runs the full
// suite — never an implementer.**, skills/flow/implement.md). Exit 0 clean,
// 1 one `file:line: task <id> runs the suite …` line per span, 2 a usage
// error or a file that cannot be read.
func init() { Registry["plan-suite-steps"] = planSuiteStepsCmd }

var (
	pssSpanRE       = regexp.MustCompile("`([^`]+)`")
	pssHead         = `(^|&&|;|\|)\s*`
	pssGradleRE     = regexp.MustCompile(pssHead + `(\S*/)?gradlew?(\s|$)`)
	pssGradleTestRE = regexp.MustCompile(`(?i)^(\S*:)?\w*test$`)
	pssGoTestRE     = regexp.MustCompile(pssHead + `go\s+test(\s|$)`)
	pssJSTestRE     = regexp.MustCompile(pssHead + `(npm\s+(run\s+)?test|(npx\s+)?vitest)(\s|$)`)
	pssGradleSel    = regexp.MustCompile(`(^|\s)--tests(\s|=|$)`)
	pssGoSel        = regexp.MustCompile(`(^|\s)-run(\s|=|$)`)
	pssJSSel        = regexp.MustCompile(`(^|\s)-t(\s|=|$)`)
)

func planSuiteStepsCmd(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 || args[0] == "" {
		fmt.Fprintln(stderr, "usage: flow-guard plan-suite-steps <tasks.md>")
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
	found := 0
	for _, task := range tcfIterTasks(lines) {
		// Join the unfenced body so a span wrapped across two lines is read
		// whole; starts maps each joined line back to its body offset.
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
			cmd := strings.Join(strings.Fields(joined[m[2]:m[3]]), " ")
			if !pssRunsSuite(cmd) {
				continue
			}
			i := 0
			for i+1 < len(starts) && starts[i+1] <= m[0] {
				i++
			}
			fmt.Fprintf(stdout, "%s:%d: task %s runs the suite `%s` — a task verifies with its own **Tests:** selector; the full `## test` list runs once, after the last task\n",
				args[0], task.taskLine+1+offsets[i], task.id, cmd)
			found++
		}
	}
	if found > 0 {
		return 1
	}
	return 0
}

// pssRunsSuite reports whether one inline command runs a test suite with no
// selector narrowing it to named tests.
func pssRunsSuite(cmd string) bool {
	if pssGradleRE.MatchString(cmd) && !pssGradleSel.MatchString(cmd) {
		for _, tok := range strings.Fields(cmd) {
			if !strings.HasPrefix(tok, "-") && pssGradleTestRE.MatchString(tok) {
				return true
			}
		}
	}
	if pssGoTestRE.MatchString(cmd) && !pssGoSel.MatchString(cmd) {
		return true
	}
	return pssJSTestRE.MatchString(cmd) && !pssJSSel.MatchString(cmd)
}
