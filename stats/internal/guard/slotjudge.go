package guard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// flow-guard slot-judge <digest-file> — ask the pair-judge model, at Decide
// over the plan, which of the three gated review slots the plan needs
// (KAN-1126; skills/flow/brainstorm-planner.md Decide step 7).
//
// Prints exactly three lines, in this fixed order, panel-slot-triggers'
// format (decision slot-judge-same-output-shape):
//
//	test-sanity: run|skip — <reason>
//	call-efficiency: run|skip — <reason>
//	failure-modes: run|skip — <reason>
//
// On an answer each reason is the judge's own one line followed by
// `; judge` and the spend suffix, ` (judge <s>s[, $<cost>])`. On any
// fallback all three print as `<slot>: run — fallback: <cause>`, the cause
// carrying the spend whenever the command ran (decision slot-judge-fail-open).
//
// The judge is pjRun: headless `claude -p` -- or
// FLOW_PAIR_JUDGE_CMD -- on the pair-judge row's first base pair, read with
// `flow policy get`; the 60 s deadline (Env.PairJudgeTimeout); the digest
// on stdin, cut at 204800 bytes and followed by `[digest truncated at
// 200 KB]`. Its system prompt is sjPrompt, a fixed template naming each slot
// with the ticket's condition. The digest is Decide's: the plan's tasks and
// its `inputs:` line.
//
// Fallback causes: no pair-judge pair, claude not on PATH, the command not runnable, a non-zero exit, the
// deadline, an answer that is not JSON; `policy unreadable: <err>`; and
// `judge answer is malformed: <answer>` when a slot's line is missing, a
// slot is answered twice, or a line opening with a slot name is not
// `<slot>: run|skip — <one line>`.
//
// Exit 0 on the three lines, whatever the judge did. Exit 2, nothing on
// stdout: usage (reason usage) or an unreadable digest (digest-unreadable).
//
// Go, in flow-guard (stats/internal/guard/slotjudge.go).
func init() {
	Registry["slot-judge"] = slotJudge
}

var sjLine = regexp.MustCompile(`^(test-sanity|call-efficiency|failure-modes):[ \t]*(run|skip)[ \t]+—[ \t]+(\S.*?)[ \t]*$`)

// sjPrompt is the judge's system prompt; the conditions are KAN-1126's,
// verbatim.
const sjPrompt = `You decide which optional code-review slots a change's review panel needs.

The user message is a plan written before any code exists: its tasks, the files each names, and an inputs: line counting them.

The three slots, each with the condition under which it runs:
- test-sanity: runs only when tests are added or changed.
- call-efficiency: runs only when the diff touches request handlers, queries or client calls.
- failure-modes: runs only for I/O, persistence or concurrency.

A slot runs unless the plan clearly does no such work. When in doubt, answer run.

Answer with exactly three lines and nothing else:
test-sanity: run|skip — <one line>
call-efficiency: run|skip — <one line>
failure-modes: run|skip — <one line>
`

func slotJudge(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: flow-guard slot-judge <digest-file>")
		Reason(env, "usage")
		return 2
	}
	digest, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "slot-judge: %v\n", err)
		Reason(env, "digest-unreadable")
		return 2
	}
	if len(digest) > pjDigestCap {
		digest = append(digest[:pjDigestCap:pjDigestCap], pjTruncated...)
	}
	verdicts, cause := sjAsk(env, digest)
	for _, s := range pstSlots {
		if cause != "" {
			fmt.Fprintf(stdout, "%s: run — fallback: %s\n", s, strings.TrimSpace(pjSpace.ReplaceAllString(cause, " ")))
		} else {
			fmt.Fprintf(stdout, "%s: %s\n", s, verdicts[s])
		}
	}
	return 0
}

// sjAsk runs the judge and parses its answer into each slot's
// `<verdict> — <reason>; judge<spend>`; a non-empty cause is why the
// answer cannot be used.
func sjAsk(env Env, digest []byte) (map[string]string, string) {
	rows, err := policyTable(env)
	if err != nil {
		return nil, "policy unreadable: " + err.Error()
	}
	result, spend, cause := pjRun(env, rows, sjPrompt, digest)
	if cause != "" {
		return nil, cause
	}
	malformed := "judge answer is malformed: " + strings.TrimSpace(result) + spend
	verdicts := map[string]string{}
	for _, line := range strings.Split(result, "\n") {
		line = strings.TrimSpace(line)
		slot, _, ok := strings.Cut(line, ":")
		if !ok || !slices.Contains(pstSlots, slot) {
			continue
		}
		m := sjLine.FindStringSubmatch(line)
		if m == nil || verdicts[slot] != "" {
			return nil, malformed
		}
		verdicts[slot] = m[2] + " — " + m[3] + "; judge" + spend
	}
	if len(verdicts) != len(pstSlots) {
		return nil, malformed
	}
	return verdicts, ""
}

var (
	pjPairShape = regexp.MustCompile(`^([^/]+)/(?:[A-Za-z]+@|[a-z-]*-)(low|medium|high|xhigh|max)$`)
	pjSpace     = regexp.MustCompile(`\s+`)
)

const (
	pjDigestCap    = 204800
	pjTruncated    = "\n[digest truncated at 200 KB]\n"
	pjDefaultLimit = 60 * time.Second
)

// pjCandidate is one parsed pair.
type pjCandidate struct{ pair, model, effort string }

func pjParse(pair string) (pjCandidate, bool) {
	m := pjPairShape.FindStringSubmatch(pair)
	if m == nil {
		return pjCandidate{}, false
	}
	return pjCandidate{pair, m[1], m[2]}, true
}

// pjRun runs the judge -- headless `claude -p` (or FLOW_PAIR_JUDGE_CMD) on
// the pair-judge row's first base pair, under the deadline, with system as
// its system prompt and digest on stdin. It runs with
// CLAUDE_CODE_DISABLE_AUTO_MEMORY=1: `--setting-sources ""` still loads the
// cwd project's auto-memory index. FLOW_PAIR_JUDGE_CMD, when set, names the
// command run in place of `claude`. result is
// the answer JSON's `result` text; spend the ` (judge <s>s[, $<cost>])`
// suffix; cause, when non-empty, why there is no usable answer, already
// carrying spend whenever the command ran.
func pjRun(env Env, rows map[string][]policyRow, system string, digest []byte) (result, spend, cause string) {
	var judge pjCandidate
	if rs := rows["pair-judge"]; len(rs) > 0 && len(rs[0].Pairs) > 0 {
		judge, _ = pjParse(rs[0].Pairs[0])
	}
	if judge.model == "" {
		return "", "", "no pair-judge pair"
	}
	bin := env.Getenv("FLOW_PAIR_JUDGE_CMD")
	if bin == "" {
		var ok bool
		if bin, ok = lookPath(env, "claude"); !ok {
			return "", "", "claude not on PATH"
		}
	}
	limit := env.PairJudgeTimeout
	if limit == 0 {
		limit = pjDefaultLimit
	}
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-p", "--model", judge.model, "--effort", judge.effort,
		"--tools", "", "--setting-sources", "", "--strict-mcp-config", "--no-session-persistence",
		"--output-format", "json", "--system-prompt", system)
	cmd.Dir = env.Dir
	cmd.Env = append(os.Environ(), "CLAUDE_CODE_DISABLE_AUTO_MEMORY=1")
	cmd.Stdin = bytes.NewReader(digest)
	cmd.WaitDelay = time.Second // an orphaned child holding the pipes cannot outlast the deadline by more
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	start := time.Now()
	err := cmd.Run()
	var ans struct {
		Result string   `json:"result"`
		Cost   *float64 `json:"total_cost_usd"`
	}
	jerr := json.Unmarshal(out.Bytes(), &ans)
	spend = fmt.Sprintf(" (judge %.1fs", time.Since(start).Seconds())
	if jerr == nil && ans.Cost != nil {
		spend += fmt.Sprintf(", $%.4f", *ans.Cost)
	}
	spend += ")"
	var ee *exec.ExitError
	switch {
	case ctx.Err() != nil:
		return "", spend, fmt.Sprintf("judge timed out after %v", limit) + spend
	case errors.As(err, &ee):
		return "", spend, fmt.Sprintf("judge %s exit %d: %s", filepath.Base(bin), ee.ExitCode(), strings.TrimSpace(errb.String())) + spend
	case err != nil:
		return "", spend, fmt.Sprintf("judge command %s not runnable: %v", bin, err) + spend
	case jerr != nil:
		return "", spend, "judge answer is not JSON: " + strings.TrimSpace(out.String()) + spend
	}
	return ans.Result, spend, ""
}
