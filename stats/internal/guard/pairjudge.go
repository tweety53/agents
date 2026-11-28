package guard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/modelpolicy"
)

// flow-guard pair-judge <role> <digest-file> — pick one judged
// dispatch's {model, effort} from its role's allowed pairs (KAN-1119).
//
// Prints one line, `<model> <effort> <judge|fallback|single> <reason>`, the
// reason one line (runs of whitespace collapsed to one space). The same line
// is saved, overwriting, beside the digest file as `<digest less its
// extension>.pick` -- `pair-judge/<key>.md` saves `pair-judge/<key>.pick`, which
// flow-guard hook enforce-agent-baseline reads as the row's -pair-pick (decision
// pair-pick-file); a failed save is one stderr line and leaves the exit 0.
//
// Every role with base pairs in flowd is judged, bar pair-judge itself, which is refused (exit 2).
// A role with no base pair is refused as candidates-empty.
//
// The candidates (decision judge-candidates-capped) are the role's base
// pairs, read with `flow policy get` (PolicyTable), in row order, each parsed into model and effort --
// `opus/flow-low`, `opus/flow-planner-medium`, `sonnet/Explore@low` -- an
// unparseable pair dropped. Which pairs a role admits is its row's alone: no
// model or effort is filtered here. Marked and opt-in pairs are never
// candidates.
//
// The three sources:
//   - single: one candidate, printed with the reason `the only candidate`
//     and no model call;
//   - judge: the judge answered `PAIR: <one candidate, verbatim>` and
//     `REASON: <one line>`; the pick and that reason are printed;
//   - split: the implementer role's judge pick of opus/flow-low, half the
//     time, printed as sonnet/flow-low (reason names the swap) when the row
//     admits it; every other pick stays as it was (Env.PairJudgeCoin in tests);
//   - fallback: the first candidate, the reason naming the cause -- no
//     pair-judge pair to run on, the command missing or not runnable, a
//     non-zero exit, the deadline, an answer that is not JSON, no PAIR or REASON line, or a pick outside
//     the candidates (decisions fallback-first-candidate,
//     missing-judge-row-falls-back).
//
// The judge is headless `claude -p` on the pair-judge role's first base
// pair, its system prompt a fixed template (pjPrompt) filled with the role,
// its Does line from skills/flow-contracts/model-policy.md under
// FLOW_GUARD_REPO_ROOT (left out when unreadable), the candidates and the
// answer format; its stdin is the digest followed by the file parts it names
// (pjAppendNamed, pairjudge_files.go), since the judge has no tools to read a
// path itself. A part is a backticked `<abs-path>#<heading text>` (that
// markdown section), an `<abs-path>:<a>-<b>` or `:<a>` (those lines), or, for
// a tasks.md named bare, the sections of the tasks the digest names
// (`Tasks: 7, 8 of 20`, `PASS task-7:`), each capped at 8 KB under a
// `--- <reference> ---` header. A file named without a part, and a part it
// lacks, are not sent. The content leaves
// the machine, so a file is sent only when, symlinks resolved, it lies inside
// a git working tree, is not gitignored unless under `.superpowers/` or `spectre/changes/`, is not a
// `.env*` file, private key or key store, or file named secret, secrets or
// credentials, and is not under ~/.claude or ~/.config; any other path, and an unreadable one, is silently
// left out. The digest itself is never cut for a file. Measured live
// from a Claude Code Bash call on 2026-11-15, claude 2.1.325, exit 0 in
// 2.5 s with a PAIR and a REASON line; --output-format json spiked the same
// day on the same version: one JSON object with `result` (the PAIR and
// REASON text), `total_cost_usd` and `duration_ms`:
//
//	claude -p --model <m> --effort <e> --tools "" --setting-sources ""
//	  --strict-mcp-config --no-session-persistence --output-format json
//	  --system-prompt <prompt>
//
// Whenever the command ran, the printed reason -- the judge's, or a
// fallback's cause -- ends with its spend, ` (judge <seconds, one
// decimal>s, $<total_cost_usd, four decimals>)`, the cost part left out when
// the answer carries none; an answer that is not that JSON falls back with
// the cause `judge answer is not JSON: <stdout>` (kan-1119 self-review, F12).
//
// The judge runs with CLAUDE_CODE_DISABLE_AUTO_MEMORY=1: `--setting-sources ""`
// still loads the cwd project's auto-memory index, and a judge in gymie read
// its expired "at most opus at low effort" note into 13 of 172 group picks.
// Measured 2026-11-17, claude in a project with a memory index: the index is
// in context without the variable and absent with it, $0.0222 against $0.0047.
//
// The file defaults to ~/.config/flow/headless-api-key when that exists, so
// the switch is the file itself: rename it to go back to the subscription;
// FLOW_HEADLESS_API_KEY_FILE=off does the same for one environment.
// FLOW_HEADLESS_API_KEY_FILE, when set, names a file holding an Anthropic API
// key; the judge child alone gets it as ANTHROPIC_API_KEY, so every headless
// call bills API credits while no other `claude` -- the session's own
// included -- sees the key or leaves the subscription login. An unreadable or
// empty file is the fallback's cause, never a silent run on the subscription.
// The key is added only when the child's ANTHROPIC_BASE_URL is unset or
// Anthropic's own API host (the parsed host, compared exactly): a session
// pointed at another vendor runs the judge on its own credentials, and the
// Anthropic key is never sent to that vendor.
//
// FLOW_PAIR_JUDGE_CMD, when set, names the command run in place of `claude`
// (decision judge-command-seam). Bounds (decision
// judge-timeout-60s-digest-200kb): a 60 s deadline (Env.PairJudgeTimeout in
// tests), and a digest over 204800 bytes is cut there and followed by the
// line `[digest truncated at 200 KB]`; under it, the named parts fill the
// rest, each over 8192 bytes sent as its head -- cut after the last whole
// line, or else the last whole rune -- and the line
// `[file truncated at 8 KB]`, later parts still following; the one crossing
// the 200 KB total is cut there and followed by the digest's line, and no
// part after it. A digest file named `decide-group-<g>…`
// -- one implementer group's Decide digest -- loses every line opening
// `inputs: `: plan-class counts the whole plan there, and the judge read a
// one-task group as the plan's 47 tasks, two repositories and migration.
//
// Exit 0 on the printed line, whatever the judge did. Exit 2, nothing on
// stdout: usage (reason usage), a role outside the judged set
// (role-not-judged), an unreadable digest (digest-unreadable), no pairs to
// read (policy-unreadable), or no base pair to pick from (candidates-empty).
//
// Go, in flow-guard (stats/internal/guard/pairjudge.go).
func init() {
	Registry["pair-judge"] = pairJudge
}

var (
	pjPairShape = regexp.MustCompile(`^([^/]+)/(?:[A-Za-z]+@|[a-z-]*-)(low|medium|high|xhigh|max)$`)
	pjSpace     = regexp.MustCompile(`\s+`)
	pjPick      = regexp.MustCompile(`(?m)^PAIR:\s*(\S+)\s*$`)
	pjWhy       = regexp.MustCompile(`(?m)^REASON:[ \t]*(\S.*)$`)
)

const (
	pjDigestCap     = 204800
	pjTruncated     = "\n[digest truncated at 200 KB]\n"
	pjFileCap       = 8192
	pjFileTruncated = "[file truncated at 8 KB]\n"
	pjDefaultLimit  = 60 * time.Second
)

// pjPrompt is the judge's system prompt: the role, its Does line, the
// candidates one per line, the signals and the answer format.
const pjPrompt = `You pick the model and effort pair a coding-agent dispatch runs on.

The dispatch's role is %s.%s

The candidates, one per line:
%s
The user message is a digest of the dispatch's work. Weigh what the work is, not how big it looks. Verification-only, documentation, configuration and test-only work sits at the cheap end; concurrency, data-model, platform-interop and performance-sensitive work sits at the dear end. Counts of tasks, files and lines are hints, never the verdict: a mass rename touches many files and is trivial.

Pick the cheapest candidate that is near-certain to succeed on the first try.

Answer with exactly two lines and nothing else:
PAIR: <one candidate, verbatim>
REASON: <one line>
`

// pjPlanInputs is plan-class's `inputs:` line, which counts the whole plan.
var pjPlanInputs = regexp.MustCompile(`(?m)^inputs: .*\n?`)

// pjCandidate is one parsed pair.
type pjCandidate struct{ pair, model, effort string }

func (c pjCandidate) line(source, reason string) string {
	return c.model + " " + c.effort + " " + source + " " + strings.TrimSpace(pjSpace.ReplaceAllString(reason, " "))
}

func pjParse(pair string) (pjCandidate, bool) {
	m := pjPairShape.FindStringSubmatch(pair)
	if m == nil {
		return pjCandidate{}, false
	}
	return pjCandidate{pair, m[1], m[2]}, true
}

func pairJudge(args []string, env core.Env, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: flow-guard pair-judge <role> <digest-file>")
		core.Reason(env, "usage")
		return 2
	}
	role := args[0]
	// A stale pick from an earlier run under the key must not outlive this one's refusal.
	if err := os.Remove(strings.TrimSuffix(args[1], filepath.Ext(args[1])) + ".pick"); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(stderr, "pair-judge: %v\n", err)
	}
	if role == "pair-judge" { // the judge's own pair is never judged: it would pick itself
		fmt.Fprintln(stderr, "pair-judge: the pair-judge role is not judged")
		core.Reason(env, "role-not-judged")
		return 2
	}
	digest, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Fprintf(stderr, "pair-judge: %v\n", err)
		core.Reason(env, "digest-unreadable")
		return 2
	}
	if strings.HasPrefix(filepath.Base(args[1]), "decide-group-") {
		digest = pjPlanInputs.ReplaceAll(digest, nil)
	}
	if len(digest) > pjDigestCap {
		digest = append(digest[:pjDigestCap:pjDigestCap], pjTruncated...)
	} else {
		digest = pjAppendNamed(env, args[1], digest)
	}
	rows, err := core.PolicyTable(env)
	if err != nil {
		fmt.Fprintf(stderr, "pair-judge: %v\n", err)
		core.Reason(env, "policy-unreadable")
		return 2
	}
	cands := pjCandidates(rows, role)
	if len(cands) == 0 {
		fmt.Fprintf(stderr, "pair-judge: %s has no base pair to pick from\n", role)
		core.Reason(env, "candidates-empty")
		return 2
	}
	var line string
	if len(cands) == 1 {
		line = cands[0].line("single", "the only candidate")
	} else if pick, reason, cause := pjAsk(env, rows, role, cands, digest); cause != "" {
		line = cands[0].line("fallback", cause)
	} else {
		line = pick.line("judge", reason)
	}
	line = pjSplitImplementer(env, role, cands, line)
	line += "\n"
	pickFile := strings.TrimSuffix(args[1], filepath.Ext(args[1])) + ".pick"
	if err := os.WriteFile(pickFile, []byte(line), 0o644); err != nil {
		fmt.Fprintf(stderr, "pair-judge: could not save %s: %v\n", pickFile, err)
	}
	fmt.Fprint(stdout, line)
	return 0
}

// pjSplitImplementer turns half of the implementer's opus/flow-low picks, the
// judge's or the only candidate's, into sonnet/flow-low when the row admits
// it; any other pick is returned as it was. The reason names the swap.
func pjSplitImplementer(env core.Env, role string, cands []pjCandidate, line string) string {
	const from = "opus low judge "
	if role != "implementer" || !strings.HasPrefix(line, from) {
		return line
	}
	i := slices.IndexFunc(cands, func(c pjCandidate) bool { return c.pair == "sonnet/flow-low" })
	if i < 0 {
		return line
	}
	coin := env.PairJudgeCoin
	if coin == nil {
		coin = func() bool { return rand.IntN(2) == 0 }
	}
	if !coin() {
		return line
	}
	return cands[i].line("split", "opus low pick moved to sonnet low, a 50/50 split")
}

// pjCandidates is role's base row, parsed.
func pjCandidates(rows map[string][]core.PolicyRow, role string) []pjCandidate {
	var cands []pjCandidate
	if rs := rows[role]; len(rs) > 0 {
		for _, p := range rs[0].Pairs {
			if c, ok := pjParse(p); ok {
				cands = append(cands, c)
			}
		}
	}
	return cands
}

// pjAsk runs the judge; a non-empty cause is why its answer cannot be used.
func pjAsk(env core.Env, rows map[string][]core.PolicyRow, role string, cands []pjCandidate, digest []byte) (pjCandidate, string, string) {
	result, spend, cause := pjRun(env, rows, "pair-judge", pjSystemPrompt(env, role, cands), digest, "", pjDefaultLimit)
	if cause != "" {
		return pjCandidate{}, "", cause
	}
	pm, wm := pjPick.FindStringSubmatch(result), pjWhy.FindStringSubmatch(result)
	if pm == nil || wm == nil {
		return pjCandidate{}, "", "judge answer has no PAIR or REASON line: " + strings.TrimSpace(result) + spend
	}
	i := slices.IndexFunc(cands, func(c pjCandidate) bool { return c.pair == pm[1] })
	if i < 0 {
		return pjCandidate{}, "", fmt.Sprintf("judge pick %s is outside the candidates", pm[1]) + spend
	}
	return cands[i], wm[1] + spend, ""
}

// pjRun runs the judge -- headless `claude -p` (or FLOW_PAIR_JUDGE_CMD) on
// judgeRole's first base pair, under the deadline, with system as its system
// prompt and digest on stdin -- for pair-judge (role pair-judge) and
// the other judges alike. tools is the --tools list ("" none) and limit the deadline, which
// Env.PairJudgeTimeout overrides when set. result is
// the answer JSON's `result` text; spend the ` (judge <s>s[, $<cost>])`
// suffix; cause, when non-empty, why there is no usable answer, already
// carrying spend whenever the command ran.
func pjRun(env core.Env, rows map[string][]core.PolicyRow, judgeRole, system string, digest []byte, tools string, limit time.Duration) (result, spend, cause string) {
	var judge pjCandidate
	if rs := rows[judgeRole]; len(rs) > 0 && len(rs[0].Pairs) > 0 {
		judge, _ = pjParse(rs[0].Pairs[0])
	}
	if judge.model == "" {
		return "", "", "no " + judgeRole + " pair"
	}
	bin := env.Getenv("FLOW_PAIR_JUDGE_CMD")
	if bin == "" {
		var ok bool
		if bin, ok = core.LookPath(env, "claude"); !ok {
			return "", "", "claude not on PATH"
		}
	}
	if env.PairJudgeTimeout != 0 {
		limit = env.PairJudgeTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-p", "--model", judge.model, "--effort", judge.effort,
		"--tools", tools, "--setting-sources", "", "--strict-mcp-config", "--no-session-persistence",
		"--output-format", "json", "--system-prompt", system)
	cmd.Dir = env.Dir
	cmd.Env = append(icChildEnv(env), "CLAUDE_CODE_DISABLE_AUTO_MEMORY=1")
	f := env.Getenv("FLOW_HEADLESS_API_KEY_FILE")
	if !pjAnthropicBase(cmd.Env) {
		f = "off"
	}
	if f == "" && env.Getenv("HOME") != "" {
		if d := filepath.Join(env.Getenv("HOME"), ".config", "flow", "headless-api-key"); isRegular(d) {
			f = d
		}
	}
	if f != "" && f != "off" {
		key, err := os.ReadFile(f)
		if k := strings.TrimSpace(string(key)); err != nil || k == "" {
			return "", "", "FLOW_HEADLESS_API_KEY_FILE " + f + " is unreadable or empty"
		} else {
			cmd.Env = append(cmd.Env, "ANTHROPIC_API_KEY="+k)
		}
	}
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
	elapsed := time.Since(start)
	h := core.Headless{Role: judgeRole, Model: judge.model, Effort: judge.effort, DurationMS: int(elapsed.Milliseconds())}
	if jerr == nil {
		h.CostUSD = ans.Cost
	}
	core.RecordHeadless(env, h)
	spend = fmt.Sprintf(" (judge %.1fs", elapsed.Seconds())
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

// pjSystemPrompt fills pjPrompt for role.
func pjSystemPrompt(env core.Env, role string, cands []pjCandidate) string {
	does := ""
	if d := pjDoes(env, role); d != "" {
		does = " It: " + d
	}
	var list strings.Builder
	for _, c := range cands {
		list.WriteString("- " + c.pair + "\n")
	}
	return fmt.Sprintf(pjPrompt, role, does, list.String())
}

// pjDoes is role's Does line from the roles table under FLOW_GUARD_REPO_ROOT;
// "" when the root is unset or the table or the role's row cannot be read.
func pjDoes(env core.Env, role string) string {
	root := env.Getenv("FLOW_GUARD_REPO_ROOT")
	if root == "" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(root, modelpolicy.File))
	if err != nil {
		return ""
	}
	entries, err := modelpolicy.Rows(b)
	if err != nil {
		return ""
	}
	if i := slices.IndexFunc(entries, func(e modelpolicy.Entry) bool { return e.Role == role }); i >= 0 {
		return entries[i].Does
	}
	return ""
}

// pjAnthropicBase reports whether the child environment's last
// ANTHROPIC_BASE_URL is unset, empty or Anthropic's own API host.
func pjAnthropicBase(environ []string) bool {
	base := ""
	for _, kv := range environ {
		if v, ok := strings.CutPrefix(kv, "ANTHROPIC_BASE_URL="); ok {
			base = v
		}
	}
	if base == "" {
		return true
	}
	u, err := url.Parse(base)
	return err == nil && u.Scheme == "https" && strings.EqualFold(u.Hostname(), "api.anthropic.com")
}
