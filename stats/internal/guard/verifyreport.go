package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// checkVerifyReport is scripts/check-verify-report.sh: that script's header
// comment is the contract. It reads a visual-verify verifier's report file
// and answers whether every required step carries an admissible status on the
// report's `- steps:` line, so the parent never accepts a partial report by
// reading its prose.
func init() { Registry["check-verify-report"] = checkVerifyReport }

var (
	vrSteps     = []string{"4", "7", "8", "9", "10", "11", "motion"}
	vrStepsLine = regexp.MustCompile(`^\s*- steps:\s*(.*)$`)
	vrMotion    = regexp.MustCompile(`^\s*- motion:\s*(\d+)/(\d+)\b`)
	vrCitedExit = regexp.MustCompile(`\bexit [1-9][0-9]*\b`)
)

func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-verify-report"
	if len(args) != 2 {
		fmt.Fprintf(stderr, "%s: usage: check-verify-report.sh <report file> <motions named>\n", prog)
		return 2
	}
	motions, err := strconv.Atoi(args[1])
	if err != nil || motions < 0 {
		fmt.Fprintf(stderr, "%s: usage: check-verify-report.sh <report file> <motions named> — %q is not a count\n", prog, args[1])
		return 2
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[0], err)
		return 2
	}
	status := map[string]string{}
	motionLine := false
	for _, line := range strings.Split(string(data), "\n") {
		if m := vrStepsLine.FindStringSubmatch(line); m != nil {
			for _, item := range strings.Split(m[1], " | ") {
				id, rest, _ := strings.Cut(strings.TrimSpace(item), " ")
				status[id] = strings.TrimSpace(rest)
			}
		}
		if m := vrMotion.FindStringSubmatch(line); m != nil && m[1] == m[2] && m[2] == strconv.Itoa(motions) {
			motionLine = true
		}
	}
	var undone []string
	for _, id := range vrSteps {
		s, ok := status[id]
		naAllowed := id == "4" || id == "10" || (id == "motion" && motions == 0)
		reason, isNA := strings.CutPrefix(s, "n/a — ")
		cause, isBlocked := strings.CutPrefix(s, "blocked — ")
		switch {
		case !ok:
			undone = append(undone, id+" absent")
		case s == "done" && id == "motion" && motions > 0 && !motionLine:
			undone = append(undone, fmt.Sprintf("motion: no `- motion: %d/%d` line", motions, motions))
		case s == "done":
		case isNA && naAllowed && strings.TrimSpace(reason) != "":
		case isBlocked && vrCitedExit.MatchString(cause):
		default:
			undone = append(undone, strings.TrimSpace(id+" "+s))
		}
	}
	if len(undone) > 0 {
		fmt.Fprintf(stdout, "VERIFY-REPORT-INCOMPLETE: %s — %s\n", args[0], strings.Join(undone, "; "))
		return 1
	}
	fmt.Fprintf(stdout, "VERIFY-REPORT-COMPLETE: %s\n", args[0])
	return 0
}
