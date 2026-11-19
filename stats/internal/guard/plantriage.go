package guard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// flow-guard plan-triage <worktree> <ticket-file>
// — decide whether a new /flow change gets full or short planning, with one
// headless model call.
//
// The call is pjRun on the plan-triage role's first base pair read from
// `flow policy get`. Its system prompt is skills/flow/plan-triage-prompt.md
// under FLOW_GUARD_REPO_ROOT, canonical for the criteria and the answer's
// shape; its stdin is <ticket-file> in full and `git ls-files` in
// <worktree>, cut at 200 KB as pair-judge cuts its digest.
//
// The answer is the first line of the result text shaped `TRIAGE: full|short
// — <reason>`. Prints that verdict as one line, `TRIAGE: <verdict> —
// <reason> (judge <s>s[, $<cost>])`.
//
// The fallback, on any failure -- an unreadable ticket, prompt or worktree
// listing, no plan-triage pair, the command missing or failing, the
// deadline, or an answer with no TRIAGE line -- is the safe side: one line,
// `TRIAGE: full — fallback: <cause>`.
//
// Exit 0 on a printed line, whatever the model did. Exit 2, nothing on
// stdout: usage (reason usage).
//
// Go, in flow-guard (stats/internal/guard/plantriage.go).
func init() { Registry["plan-triage"] = planTriage }

var ptVerdict = regexp.MustCompile(`(?m)^\s*TRIAGE:\s*(full|short)\s*[—-]+\s*(\S.*)$`)

func planTriage(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: flow-guard plan-triage <worktree> <ticket-file>")
		Reason(env, "usage")
		return 2
	}
	verdict, reason := ptAsk(env, args[0], args[1])
	fmt.Fprintf(stdout, "TRIAGE: %s — %s\n", verdict, strings.TrimSpace(pjSpace.ReplaceAllString(reason, " ")))
	return 0
}

// ptAsk is the verdict and its reason, `full` and `fallback: <cause>` when
// the model's answer cannot be had or used.
func ptAsk(env Env, wt, ticketFile string) (verdict, reason string) {
	fallback := func(cause string) (string, string) { return "full", "fallback: " + cause }
	ticket, err := os.ReadFile(ticketFile)
	if err != nil {
		return fallback("ticket unreadable: " + err.Error())
	}
	prompt, err := os.ReadFile(filepath.Join(env.Getenv("FLOW_GUARD_REPO_ROOT"), "skills", "flow", "plan-triage-prompt.md"))
	if err != nil {
		return fallback("plan-triage prompt unreadable: " + err.Error())
	}
	files, err := exec.Command("git", "-C", wt, "ls-files").Output()
	if err != nil {
		return fallback(fmt.Sprintf("git ls-files in %s: %v", wt, err))
	}
	rows, err := policyTable(env)
	if err != nil {
		return fallback(err.Error())
	}
	digest := []byte("## Ticket\n\n" + string(ticket) + "\n\n## Tracked files\n\n" + string(files))
	if len(digest) > pjDigestCap {
		digest = append(digest[:pjDigestCap:pjDigestCap], pjTruncated...)
	}
	result, spend, cause := pjRun(env, rows, "plan-triage", string(prompt), digest)
	if cause != "" {
		return fallback(cause)
	}
	m := ptVerdict.FindStringSubmatch(result)
	if m == nil {
		return fallback("no TRIAGE line in the answer: " + result + spend)
	}
	return m[1], m[2] + spend
}
