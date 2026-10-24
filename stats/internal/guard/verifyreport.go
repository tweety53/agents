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
	// Step 10's own blocking exit (skills/flow/visual-verify-verifier.md step 10).
	vrComposeExit = regexp.MustCompile(`\bcompose-mockup-frames(\.sh)? exit [1-9][0-9]*\b`)
	// A bare `- fingerprint: not declared`: step 6 reports exit 2 as `not
	// declared — <the guard's stderr reason>`, and the verifier has that
	// reason only from the prompt (kan-1010's self-review: a verifier wrote
	// it bare for a project declaring the row, its parent's FRESH=0 unseen).
	vrBareFingerprint = regexp.MustCompile(`^\s*- fingerprint:\s*not declared\s*$`)
	// A `## Report` line attributing a departure to the merge base must name
	// its evidence (kan-1015: the kan-466 weight-tab button was attributed
	// pre-existing from the code path alone and re-baselined over the defect).
	vrPreExisting    = regexp.MustCompile(`(?i)(?:^|[^-\w])pre-existing|present at (?:the )?merge base`)
	vrNotPreExisting = regexp.MustCompile(`(?i)\bnot pre-existing`)
	// A scoped re-run's `- verify:` line names how many specs it ran
	// (kan-1015: `specs` printed 2 under the round base, the verifier ran 95).
	vrVerifyLine  = regexp.MustCompile(`^\s*- verify:\s*(.*)$`)
	vrMappedSpecs = regexp.MustCompile(`\b(\d+) mapped specs\b`)
	// A `- wire contract:` line naming an endpoint: the dispatch named no
	// view, so steps 8-10 have nothing to capture (kan-1022's self-review: a
	// verifier wrote them `done` over a capture it never ran).
	vrWireContract = regexp.MustCompile(`^\s*- wire contract:\s*(.*)$`)
)

func checkVerifyReport(args []string, env Env, stdout, stderr io.Writer) int {
	const prog = "check-verify-report"
	const usage = "check-verify-report.sh <report file> <motions named> [<scoped specs>]"
	if len(args) != 2 && len(args) != 3 {
		fmt.Fprintf(stderr, "%s: usage: %s\n", prog, usage)
		return 2
	}
	motions, err := strconv.Atoi(args[1])
	if err != nil || motions < 0 {
		fmt.Fprintf(stderr, "%s: usage: %s — %q is not a count\n", prog, usage, args[1])
		return 2
	}
	scoped := -1
	if len(args) == 3 {
		if scoped, err = strconv.Atoi(args[2]); err != nil || scoped < 0 {
			fmt.Fprintf(stderr, "%s: usage: %s — %q is not a count\n", prog, usage, args[2])
			return 2
		}
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%s: cannot read %s: %v\n", prog, args[0], err)
		return 2
	}
	status := map[string]string{}
	motionLine, bareFingerprint := false, false
	inReport, unevidenced, verify := false, "", ""
	noView := false
	for _, line := range strings.Split(string(data), "\n") {
		if m := vrVerifyLine.FindStringSubmatch(line); m != nil && verify == "" {
			verify = strings.TrimSpace(m[1])
		}
		if strings.HasPrefix(line, "## ") {
			inReport = strings.TrimSpace(line) == "## Report"
		}
		if inReport && unevidenced == "" && vrPreExisting.MatchString(line) && !vrNotPreExisting.MatchString(line) &&
			!strings.Contains(line, "merge-base baseline: ") && !strings.Contains(line, "merge-base capture: ") {
			unevidenced = strings.TrimSpace(line)
			if r := []rune(unevidenced); len(r) > 120 {
				unevidenced = string(r[:120])
			}
		}
		if m := vrStepsLine.FindStringSubmatch(line); m != nil {
			for _, item := range strings.Split(m[1], " | ") {
				id, rest, _ := strings.Cut(strings.TrimSpace(item), " ")
				status[id] = strings.TrimSpace(rest)
			}
		}
		if m := vrMotion.FindStringSubmatch(line); m != nil && m[1] == m[2] && m[2] == strconv.Itoa(motions) {
			motionLine = true
		}
		bareFingerprint = bareFingerprint || vrBareFingerprint.MatchString(line)
		if m := vrWireContract.FindStringSubmatch(line); m != nil && !strings.HasPrefix(strings.TrimSpace(m[1]), "n/a") {
			noView = true
		}
	}
	var undone []string
	if bareFingerprint {
		undone = append(undone, "fingerprint: `not declared` with no reason — state step 6's result as the prompt names it")
	}
	if unevidenced != "" {
		status["10"] = "pre-existing attribution with no merge-base evidence: " + unevidenced
	}
	if scoped >= 0 && !(scoped == 0 && strings.HasPrefix(verify, "not run — no spec since the round base")) {
		if m := vrMappedSpecs.FindStringSubmatch(verify); m == nil {
			status["7"] = "names no mapped-spec count"
		} else if m[1] != strconv.Itoa(scoped) {
			status["7"] = fmt.Sprintf("ran %s mapped specs where specs printed %d", m[1], scoped)
		}
	}
	for _, id := range vrSteps {
		s, ok := status[id]
		naAllowed := id == "4" || id == "10" || (id == "motion" && motions == 0) || (noView && (id == "8" || id == "9"))
		reason, isNA := strings.CutPrefix(s, "n/a — ")
		cause, isBlocked := strings.CutPrefix(s, "blocked — ")
		switch {
		case !ok:
			undone = append(undone, id+" absent")
		case s == "done" && id == "motion" && motions > 0 && !motionLine:
			undone = append(undone, fmt.Sprintf("motion: no `- motion: %d/%d` line", motions, motions))
		case s == "done":
		// Step 10 is n/a only when no mockups are declared, so its reason must say so,
		// or when a wire contract names no view to compose.
		case isNA && naAllowed && strings.TrimSpace(reason) != "" && (id != "10" || noView || strings.Contains(reason, "mockups")):
		// Step 10 measures the captures step 8 wrote, so it blocks only when step 8 did or
		// on compose-mockup-frames' own exit (KAN-1000: two verifiers skipped step 10 with
		// every capture on disk, and departures surfaced one round at a time).
		case isBlocked && id == "10" && !strings.HasPrefix(status["8"], "blocked — ") && !vrComposeExit.MatchString(cause):
			undone = append(undone, strings.TrimSpace(id+" "+s)+" (step 8 captured)")
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
