package guard

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// checkLiveVerification is scripts/check-live-verification.sh: that script's
// header comment is the contract. A record passes when some line names a
// before figure and no line states a figure was not measured.
func init() { Registry["check-live-verification"] = checkLiveVerification }

var (
	lvBefore     = regexp.MustCompile(`(?i)\bbefore\b`)
	lvUnmeasured = regexp.MustCompile(`(?i)\b(not measured|unmeasured|not taken)\b`)
)

func checkLiveVerification(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-live-verification"
	if len(args) != 1 {
		fmt.Fprintf(stderr, "%s: usage: check-live-verification.sh <live-verification.md>\n", prog)
		return 2
	}
	record := args[0]
	data, err := os.ReadFile(record)
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, record, err)
		return 2
	}
	named := false
	var bad []string
	for i, line := range strings.Split(string(data), "\n") {
		if lvBefore.MatchString(line) {
			named = true
		}
		if lvUnmeasured.MatchString(line) {
			bad = append(bad, fmt.Sprintf("%s:%d: before figure not measured — %s", record, i+1, line))
		}
	}
	switch {
	case len(bad) > 0:
		fmt.Fprintln(stdout, strings.Join(bad, "\n"))
		return 1
	case !named:
		fmt.Fprintf(stdout, "%s: no before figure — no line names one\n", record)
		return 1
	}
	fmt.Fprintf(stdout, "LIVE-VERIFICATION-OK: %s\n", record)
	return 0
}
